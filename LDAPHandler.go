package main

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/ldap.v2"
	"strings"
)

// pLDAPConnect connects to LDAP
func pLDAPConnect() (*ldap.Conn, error) {
	l, err := ldap.Dial("tcp", configuration.LDAPServer+":"+configuration.LDAPPort)
	return l, err
}

// pLDAPConnectAnon binds to LDAP anonymously (only read access)
func pLDAPConnectAnon() (*ldap.Conn, error) {
	l, err := pLDAPConnect()
	if err != nil {
		return nil, err
	}
	// Bind with anonymous user
	err = l.Bind("", "")
	return l, err
}

// pLDAPConnectAdmin binds to LDAP with editing permissions
func pLDAPConnectAdmin() (*ldap.Conn, error) {
	l, err := pLDAPConnect()
	if err != nil {
		return nil, err
	}
	// Bind with Admin credentials
	err = l.Bind(configuration.LDAPAdmin, configuration.LDAPPass)
	return l, err
}

// LDAPAuthenticateAdmin checks whether given user has admin permissions
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

// LDAPAddUser adds user with given dn to LDAP
func LDAPAddUser(dn string, user User) error {
	l, err := pLDAPConnectAdmin()
	if err != nil {
		return err
	}

	password, err := ldapEncodePassword(user.Password)
	if err != nil {
		return err
	}

	// Add User Entry
	ar := ldap.NewAddRequest(dn)
	ar.Attribute("objectclass", []string{"inetOrgPerson", "person", "top", "organizationalPerson"})
	ar.Attribute("cn", []string{user.Username})
	ar.Attribute("sn", []string{user.Username})
	ar.Attribute("displayName", []string{user.Username})
	ar.Attribute("userPassword", password)
	err = l.Add(ar)
	l.Close()
	if err != nil {
		return err
	}
	// Add User to appropriate Group
	err = LDAPAddUserToGroup(user.Username, user.Fs)
	return err
}

// LDAPAddUserToGroup adds user to Group
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

// LDAPRemoveUserFromGroup removes user from group
func LDAPRemoveUserFromGroup(username, groupname string) error {
	conn, err := pLDAPConnectAdmin()
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
	// Remove from group
	mr := ldap.NewModifyRequest("cn=" + groupname + "," + configuration.LDAPBaseDN)
	mr.Delete("member", []string{sr[0].DN})
	err = conn.Modify(mr)
	return err
}

// LDAPChangeUserPassword changes password of user given username and new password
func LDAPChangeUserPassword(username, password string) error {
	l, err := pLDAPConnectAdmin()
	if err != nil {
		return err
	}
	defer l.Close()

	// Validate User
	sr, err := pLDAPSearch([]string{"dn"}, fmt.Sprintf(configuration.LDAPUserfilter, username))
	if err != nil {
		return err
	}
	if len(sr) != 1 {
		// User does not exist or too many entries returned
		return errors.New("Invalid Username supplied!")
	}

	pass, err := ldapEncodePassword(password)
	if err != nil {
		return err
	}

	mr := ldap.NewModifyRequest(sr[0].DN)
	mr.Replace("userPassword", pass)
	err = l.Modify(mr)
	return err
}

// LDAPAddGroup adds Group with given dn to LDAP
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

// LDAPDeleteDN removes given dn from LDAP
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

// LDAPViewGroups gets dn of all groups from LDAP
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

// LDAPViewUsers gets dn of all users from LDAP
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
		groupList = groupList[:len(groupList)-1] // remove trailing `;`

		users[i] = "{" + "\"name\": \"" + result[i].DN + "\"," +
			"\"groups\": \"" + groupList + "\"}"
		users[i] = strings.Replace(users[i], ","+configuration.LDAPBaseDN, "", -1)
	}

	return users, nil
}

// pLDAPSearch searches LDAP for dn with given attributes matching given filter
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
