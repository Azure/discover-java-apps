package core

type Credential struct {
	Id             string `json:"Id,omitempty"`
	FriendlyName   string `json:"FriendlyName,omitempty"`
	Username       string `json:"UserName,omitempty"`
	Password       string `json:"Password,omitempty"`
	CredentialType string `json:"CredentialType,omitempty"`
}

type CredentialProvider interface {
	GetCredentials() ([]*Credential, error)
}

type usernamePasswordCredentialProvider struct {
	username string
	password string
}

func NewUsernamePasswordCredentialProvider(username string, password string) CredentialProvider {
	return &usernamePasswordCredentialProvider{
		username: username,
		password: password,
	}
}

func (u usernamePasswordCredentialProvider) GetCredentials() ([]*Credential, error) {
	var cred = Credential{
		Username: u.username,
		Password: u.password,
	}
	return []*Credential{&cred}, nil
}
