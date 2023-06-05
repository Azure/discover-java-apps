package main

import (
	"github.com/Azure/discover-java-apps/weblogic"
)

func WebLogicNewUsernamePasswordCredentialProvider(username, password, weblogicusername, weblogicpassword string, weblogicport int) weblogic.CredentialProvider {
	return &weblogicUsernamePasswordCredentialProvider{Username: username, Password: password, WeblogicUsername: weblogicusername, WeblogicPassword: weblogicpassword, Weblogicport: weblogicport}
}

type weblogicUsernamePasswordCredentialProvider struct {
	Username         string
	Password         string
	WeblogicUsername string
	WeblogicPassword string
	Weblogicport     int
}

func (p weblogicUsernamePasswordCredentialProvider) GetCredentials() ([]*weblogic.Credential, error) {
	return []*weblogic.Credential{
		{
			Id:               p.Username,
			Username:         p.Username,
			Password:         p.Password,
			WeblogicUsername: p.WeblogicUsername,
			WeblogicPassword: p.WeblogicPassword,
			Weblogicport:     p.Weblogicport,
		},
	}, nil
}
