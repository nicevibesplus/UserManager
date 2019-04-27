package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/ldap.v2"
	"strings"
)

func pLDAPConnect() (*ldap.Conn, error) {
	l, err := ldap.Dial("tcp", configuration.LDAPServer+":"+configuration.LDAPPort)
	return l, err
}

func pLDAPConnectAnon() (*ldap.Conn, error) {
	l, err := pLDAPConnect()
	if err != nil {
		return nil, err
	}
	// Bind with anonymous user
	err = l.Bind("", "")
	return l, err
}

func pLDAPConnectAdmin() (*ldap.Conn, error) {
	l, err := pLDAPConnect()
	if err != nil {
		return nil, err
	}
	// Bind with Admin credentials
	err = l.Bind(configuration.LDAPAdmin, configuration.LDAPPass)
	return l, err
}

func LDAPAuthenticateAdmin(admin User) (bool, error) {
	// Connect to LDAP
	l, err := pLDAPConnectAnon()
	if err != nil {
		return false, err
	}
	defer l.Close()

	sr, err := pLDAPSearch([]string{"dn"}, fmt.Sprintf(configuration.LDAPAdminfilter, admin.Username))
	if err != nil {
		return false, nil
	}

	// User does not exist or too many entries returned
	if len(sr) != 1 {
		return false, nil
	}

	// Bind as the user to verify their password
	err = l.Bind(sr[0].DN, admin.Password)
	if err != nil {
		return false, nil
		// Wrong password
	}
	return true, nil
}

func LDAPAddUser(dn string, user User) error {
	l, err := pLDAPConnectAdmin()
	if err != nil {
		return err
	}

	if user.Password == "" {
		return errors.New("Empty password supplied.")
	}

	// Decode hex-encoded SHA512 Hash to Base64 encoding
	src := make([]byte, hex.DecodedLen(len(user.Password)))
	_, err = hex.Decode(src, []byte(user.Password))
	if err != nil {
		return err
	}
	encoded := base64.StdEncoding.EncodeToString(src)

	// Add User Entry
	ar := ldap.NewAddRequest(dn)
	ar.Attribute("objectclass", []string{"inetOrgPerson", "person", "top", "organizationalPerson"})
	ar.Attribute("cn", []string{user.Username})
	ar.Attribute("sn", []string{user.Username})
	ar.Attribute("displayName", []string{user.Username})
	ar.Attribute("userPassword", []string{"{SHA512}" + encoded})
	err = l.Add(ar)
	l.Close()
	if err != nil {
		return err
	}
	// Add User to appropiate Group
	err = LDAPAddUserToGroup(user.Username, user.Fs)
	return err
}

func LDAPAddUserToGroup(username, groupname string) error {
	l, err := pLDAPConnectAdmin()
	if err != nil {
		return err
	}
	// Validate User
	sr, err := pLDAPSearch([]string{"dn"}, fmt.Sprintf(configuration.LDAPUserfilter, username))
	if err != nil {
		return err
	}
	if len(sr) != 1 {
		// User does not exist or too many entries returned
		return errors.New("Invalid Username supplied!")
	}

	mr := ldap.NewModifyRequest("cn=" + groupname + "," + configuration.LDAPBaseDN)
	mr.Add("member", []string{sr[0].DN})
	err = l.Modify(mr)
	l.Close()
	return err
}

func LDAPChangeUserPassword(username, password string) error {
	l, err := pLDAPConnectAdmin()
	if err != nil {
		return err
	}

	// Validate User
	sr, err := pLDAPSearch([]string{"dn"}, fmt.Sprintf(configuration.LDAPUserfilter, username))
	if err != nil {
		return err
	}
	if len(sr) != 1 {
		// User does not exist or too many entries returned
		return errors.New("Invalid Username supplied!")
	}

	mr := ldap.NewModifyRequest(sr[0].DN)
	mr.Replace("userPassword", []string{password})
	err = l.Modify(mr)
	l.Close()
	return err
}

func LDAPAddGroup(dn string) error {
	l, err := pLDAPConnectAdmin()
	if err != nil {
		return err
	}

	ar := ldap.NewAddRequest(dn)
	ar.Attribute("objectclass", []string{"groupOfNames", "top"})
	ar.Attribute("member", []string{""})
	err = l.Add(ar)
	l.Close()
	return err
}

func LDAPRemoveUserFromGroup(dn, group string) error {
	conn, err := pLDAPConnectAdmin()
	if err != nil {
		return err
	}
	// Delete User from Group

	mr := ldap.NewModifyRequest(group)
	mr.Delete("member", []string{dn})
	err = conn.Modify(mr)
	return err
}

func LDAPDeleteDN(dn string) error {
	l, err := pLDAPConnectAdmin()
	if err != nil {
		return err
	}
	// Delete Entry
	dr := ldap.NewDelRequest(dn, []ldap.Control{})
	err = l.Del(dr)
	l.Close()
	return err
}

func LDAPViewGroups() (groups []string, err error) {
	result, err := pLDAPSearch(
		[]string{"cn", "member"},
		"(objectClass=groupOfNames)",
	)
	if err != nil {
		return nil, err
	}

	groups = make([]string, len(result))
	for i := range result {
		groups[i] = result[i].DN
		memberList := strings.Join(result[i].GetAttributeValues("member"), ";")
		strings.Replace(memberList, ","+configuration.LDAPBaseDN, "", -1)
		groups[i] = "{" + "\"name\": \"" + result[i].DN + "\"," +
			"\"members\": \"" + memberList + "\"}"
		groups[i] = strings.Replace(groups[i], ","+configuration.LDAPBaseDN, "", -1)
	}

	return groups, nil
}

func LDAPViewUsers() ([]string, error) {
	result, err := pLDAPSearch(
		[]string{"cn", "memberOf"},
		"(objectClass=organizationalPerson)",
	)
	if err != nil {
		return nil, err
	}
	users := make([]string, len(result))

	for i := range result {
		var groupList = ""
		if result[i].Attributes[1] == nil {
			// Invalid
			continue
		}
		for j := range result[i].Attributes[1].Values {
			groupList += result[i].Attributes[1].Values[j] + ";"
		}
		groupList = groupList[:len(groupList) - 1] // remove trailing `;`

		users[i] = "{" + "\"name\": \"" + result[i].DN + "\"," +
			"\"groups\": \"" + groupList + "\"}"
		users[i] = strings.Replace(users[i], ","+configuration.LDAPBaseDN, "", -1)
	}

	return users, nil
}

func pLDAPSearch(attributes []string, filter string) (result []*ldap.Entry, err error) {
	l, err := pLDAPConnectAnon()
	if err != nil {
		return nil, err
	}
	defer l.Close()

	// Search for the given username
	searchRequest := ldap.NewSearchRequest(
		configuration.LDAPBaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		attributes,
		nil,
	)
	sr, err := l.Search(searchRequest)
	if err != nil {
		return nil, err
	}
	return sr.Entries, nil
}
