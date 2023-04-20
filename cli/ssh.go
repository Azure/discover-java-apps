package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"net"
)

var publicKeyStore = make(map[string]string)

func MemoryHostKeyCallbackFunction() ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		fingerprint := ssh.FingerprintSHA256(key)
		if stored, ok := publicKeyStore[hostname]; !ok {
			publicKeyStore[hostname] = fingerprint
		} else if stored != fingerprint {
			return fmt.Errorf("ssh: hostkey mismatch, previous: %s, current: %s", stored, fingerprint)
		}
		return nil
	}
}
