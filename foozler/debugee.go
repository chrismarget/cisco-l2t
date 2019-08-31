package foozler

import (
	"bufio"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"strings"
	"time"
)

// Debugee represents a Cisco switch that runs l2t. In this case - as a
// a debug target. Upon connection, the target switch will be configured
// to produce l2t debug output.
type Debugee struct {
	session *ssh.Session
	stdin   io.WriteCloser
	enable  chan bool
	out     chan string
	onDeath chan error
	stop    chan chan struct{}
}

// Enable enables output from the switch.
func (o *Debugee) Enable() {
	o.enable <- true
}

// Disable disables output from the switch.
func (o *Debugee) Disable() {
	o.enable <- false
}

// Wait returns a channel that receives nil, or an error when the SSH
// session ends.
func (o *Debugee) Wait() <-chan error {
	return o.onDeath
}

// Output returns a channel that receives debug output from a switch
// when output is enabled.
func (o *Debugee) Output() <-chan string {
	return o.out
}

// Close closes the SSH session.
func (o *Debugee) Close() {
	rejoin := make(chan struct{})
	o.stop <- rejoin
	<-rejoin
}

// Execute executes a command on the switch.
func (o *Debugee) Execute(command string) error {
	_, err := io.WriteString(o.stdin, fmt.Sprintf("%s\r\n", command))
	if err != nil {
		return err
	}

	return nil
}

type DebugeeConfig struct {
	Address        string
	Port           int
	ClientConfig   *ssh.ClientConfig
	TrimTimestamps bool
}

// ConnectTo connects to a Cisco switch via SSH to facilitate debugging.
func ConnectTo(config DebugeeConfig) (*Debugee, error) {
	client, err := ssh.Dial("tcp",
		fmt.Sprintf("%s:%d", config.Address, config.Port),
		config.ClientConfig)
	if err != nil {
		return nil, err
	}

	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}

	sshIn, err := session.StdinPipe()
	if err != nil {
		return nil, err
	}

	sshOut, err := session.StdoutPipe()
	if err != nil {
		return nil, err
	}

	sshErr, err := session.StderrPipe()
	if err != nil {
		return nil, err
	}

	onSessionDeath := make(chan error, 2)
	rawOutput := make(chan []byte)
	go func() {
		scanner := bufio.NewScanner(io.MultiReader(sshOut, sshErr))

		for scanner.Scan() {
			rawOutput <- scanner.Bytes()
		}

		err := scanner.Err()
		if err != nil {
			onSessionDeath <- fmt.Errorf("stderr/stdout scanner exited - %s", err.Error())
		}
	}()

	d := &Debugee{
		session: session,
		stdin:   sshIn,
		enable:  make(chan bool),
		out:     make(chan string, 1),
		onDeath: onSessionDeath,
		stop:    make(chan chan struct{}),
	}

	go func() {
		discardIntialLines := 9
		discardTimeout := time.NewTimer(5 * time.Second)
		keepAliveDuration := 5 * time.Minute
		keepAliveTicker := time.NewTicker(keepAliveDuration)
		lastKeepAlive := time.Now()
		isEnabled := false

		for {
			select {
			case <-keepAliveTicker.C:
				if !isEnabled || time.Since(lastKeepAlive) > 6 * keepAliveDuration {
					lastKeepAlive = time.Now()
					d.Execute("show clock")
				}
			case isEnabled = <-d.enable:
			case <-discardTimeout.C:
				discardIntialLines = 0
			case raw := <-rawOutput:
				if discardIntialLines > 0 {
					discardIntialLines--
					continue
				}
				if isEnabled {
					var output string
					if config.TrimTimestamps {
						output = removeTimestamp(string(raw))
					} else {
						output = string(raw)
					}
					d.out <- output
				}
			case rejoin := <-d.stop:
				client.Close()
				keepAliveTicker.Stop()
				rejoin <- struct{}{}
				return
			}
		}
	}()

	err = session.Shell()
	if err != nil {
		return nil, err
	}

	err = d.Execute("no debug all")
	if err != nil {
		return nil, err
	}

	err = d.Execute("debug l2trace")
	if err != nil {
		return nil, err
	}

	err = d.Execute("terminal monitor")
	if err != nil {
		return nil, err
	}

	go func() {
		onSessionDeath <- session.Wait()
	}()

	return d, nil
}

func removeTimestamp(s string) string {
	if strings.Count(s, ":") < 3 {
		return s
	}

	firstSpace := strings.Index(s, " ")
	if firstSpace < 0 || firstSpace == len(s)-1 {
		return s
	}

	secondSpace := strings.Index(s, " ")
	if secondSpace < 0 || secondSpace == len(s)-1 || secondSpace - firstSpace > 2 {
		return s
	}

	end := strings.Index(s, ": ")
	if end - secondSpace < 10 || end - secondSpace > 20 {
		return s
	}

	if end + 2 > len(s) {
		return s[end+1:]
	}

	return s[end+2:]
}
