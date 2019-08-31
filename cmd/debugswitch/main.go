package main

import (
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh"
	"log"
	"os/user"
	"time"

	"github.com/chrismarget/cisco-l2t/foozler"
	"github.com/stephen-fox/sshutil"
	"github.com/stephen-fox/userutil"
)

func main() {
	address := flag.String("a", "", "The address")
	flag.Parse()

	username, err := userutil.GetUserInput("Username", userutil.PromptOptions{
		ShouldHideInput: true,
	})
	if err != nil {
		log.Fatalln(err.Error())
	}
	if len(username) == 0 {
		u, err := user.Current()
		if err != nil {
			log.Fatalln(err.Error())
		}

		username = u.Username
	}

	password, err := userutil.GetUserInput("Password", userutil.PromptOptions{
		ShouldHideInput: true,
	})

	onHostKey := func(i sshutil.SSHHostKeyPromptInfo) bool {
		b, _ := userutil.GetYesOrNoUserInput(i.UserFacingPrompt, userutil.PromptOptions{})
		return b
	}

	clientConfig := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: sshutil.ImitateSSHClientHostKeyCallBack(onHostKey),
		Timeout:         10 * time.Second,
	}

	clientConfig.Ciphers = append(clientConfig.Ciphers, "aes128-cbc")

	d, err := foozler.ConnectTo(foozler.DebugeeConfig{
		Address:        *address,
		Port:           22,
		ClientConfig:   clientConfig,
		TrimTimestamps: true,
	})
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer d.Close()

	log.Println("ready")
	d.Enable()

	for {
		select {
		case err := <-d.Wait():
			if err != nil {
				log.Fatalf("session ended - %s", err.Error())
			}
			return
		case s := <-d.Output():
			fmt.Println(s)
		}
	}
}
