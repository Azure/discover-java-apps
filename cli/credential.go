package main

import "github.com/Azure/discover-java-apps/springboot"

func NewUsernamePasswordCredentialProvider(username, password string) springboot.CredentialProvider {
	return &usernamePasswordCredentialProvider{Username: username, Password: password}
}

type usernamePasswordCredentialProvider struct {
	Username string
	Password string
}

func (p usernamePasswordCredentialProvider) GetCredentials() ([]*springboot.Credential, error) {
	return []*springboot.Credential{
		{
			Id:       p.Username,
			Username: p.Username,
			Password: p.Password,
		},
	}, nil
}
