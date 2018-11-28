package main

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/ldap.v2"
	"strings"
)

func pLDAPConnect() *ldap.Conn {
	l, err := ldap.Dial("tcp", configuration.LDAPServer+":"+configuration.LDAPPort)
	Fail(err)
	return l
}

func pLDAPConnectAnon() *ldap.Conn {
	l := pLDAPConnect()
	// Bind with anonymous user
	err := l.Bind("", "")
	Fail(err)
	return l
}

func pLDAPConnectAdmin() *ldap.Conn {
	l := pLDAPConnect()
	// Bind with Admin credentials
	err := l.Bind(configuration.LDAPAdmin, configuration.LDAPPass)
	Fail(err)
	return l
}

func LDAPAuthenticateAdmin(admin User) bool {
	// Connect to LDAP
	l := pLDAPConnectAnon()
	defer l.Close()

	sr, err := pLDAPSearch([]string{"dn"}, fmt.Sprintf(configuration.LDAPAdminfilter, admin.Username))
	Fail(err)

	if len(sr) != 1 {
		return false
		// User does not exist or too many entries returned
	}

	// Bind as the user to verify their password
	err = l.Bind(sr[0].DN, admin.Password)
	if err != nil {
		return false
		// Wrong password
	}
	return true
}

func LDAPAddUser(dn string, user User) error {
	l := pLDAPConnectAdmin()
	if user.Password == "" {
		return errors.New("Empty password supplied.")
	}

	// Add User Entry
	ar := ldap.NewAddRequest(dn)
	ar.Attribute("objectclass", []string{"inetOrgPerson", "person", "top", "organizationalPerson"})
	ar.Attribute("cn", []string{user.Username})
	ar.Attribute("sn", []string{user.Username})
	ar.Attribute("displayName", []string{user.Username})
	ar.Attribute("userPassword", []string{user.Password})
	err := l.Add(ar)
	l.Close()
	if err != nil {
		return err
	}
	// Add User to appropiate Group
	err = LDAPAddUserToGroup(user.Username, user.Fs)
	return err
}

func LDAPAddUserToGroup(username, groupname string) error {
	l := pLDAPConnectAdmin()

	// Validate User
	sr, err := pLDAPSearch([]string{"dn"}, fmt.Sprintf(configuration.LDAPUserfilter, username))
	Fail(err)
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
	l := pLDAPConnectAdmin()

	// Validate User
	sr, err := pLDAPSearch([]string{"dn"}, fmt.Sprintf(configuration.LDAPUserfilter, username))
	Fail(err)
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
	l := pLDAPConnectAdmin()

	ar := ldap.NewAddRequest(dn)
	ar.Attribute("objectclass", []string{"groupOfNames", "top"})
	ar.Attribute("member", []string{""})
	err := l.Add(ar)
	l.Close()
	return err
}

func LDAPRemoveUserFromGroup(dn, group string, l *ldap.Conn) error {
	if l == nil {
		l = pLDAPConnectAdmin()
	}
	// Delete User from Group
	mr := ldap.NewModifyRequest(group)
	mr.Delete("member", []string{dn})
	err := l.Modify(mr)
	return err
}

func LDAPDeleteDN(dn string) error {
	l := pLDAPConnectAdmin()
	// Delete Entry
	dr := ldap.NewDelRequest(dn, []ldap.Control{})
	err := l.Del(dr)
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

func LDAPViewUsers() (users []string, err error) {
	result, err := pLDAPSearch(
		[]string{"cn"},
		"(objectClass=organizationalPerson)",
	)
	Fail(err)
	users = make([]string, len(result))
	for i := range result {
		groups, err := pLDAPSearch([]string{"cn"}, fmt.Sprintf(configuration.LDAPUserGroups, result[i].DN))
		Fail(err)
		var groupList = ""
		for j := range groups {
			groupList += groups[j].DN + ";"
		}
		groupList = groupList[:len(groupList)-1]

		users[i] = "{" + "\"name\": \"" + result[i].DN + "\"," +
			"\"groups\": \"" + groupList + "\"}"
		users[i] = strings.Replace(users[i], ","+configuration.LDAPBaseDN, "", -1)
	}

	return users, nil
}

func pLDAPSearch(attributes []string, filter string) (result []*ldap.Entry, err error) {
	l := pLDAPConnectAnon()
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
