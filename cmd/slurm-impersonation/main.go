package main

import (
	"fmt"
	"golang.org/x/sys/unix"
	"log"
	"net"
)

const (
	unixSocketPath = "/tmp/slurm-impersonation.sock"
)

func getPeerCredentials(conn *net.UnixConn) (*unix.Ucred, error) {
	file, err := conn.File()
	if err != nil {
		return nil, fmt.Errorf("unable to get file from connection: %w", err)
	}
	defer file.Close()

	fd := file.Fd()

	cred, err := unix.GetsockoptUcred(
		int(fd),
		unix.SOL_SOCKET,
		unix.SO_PEERCRED,
	)

	if err != nil {
		return nil, fmt.Errorf("unable to get connected user: %w", err)
	}

	return cred, nil
}

func handleConnection(listener *net.UnixListener) error {
	connection, err := listener.AcceptUnix()
	if err != nil {
		log.Fatalf("unable to accept connection: %v", err)
	}
	defer connection.Close()

	peerCreds, err := getPeerCredentials(connection)
	if err != nil {
		log.Fatalf("unable to get peer credentials: %v", err)
	}

	msg := fmt.Sprintf("Your uid is %v", peerCreds.Uid)
	_, err = connection.Write([]byte(msg))
	if err != nil {
		log.Fatalf("unable to write message: %v", err)
	}

	return nil
}

func main() {
	unixAddr, err := net.ResolveUnixAddr("unix", unixSocketPath)
	if err != nil {
		log.Fatalf("unable to resolve unix address %q: %v", unixAddr, err)
	}

	listener, err := net.ListenUnix("unix", unixAddr)
	if err != nil {
		log.Fatalf("Unable to listen on socket: %v", err)
	}
	defer listener.Close()

	for {
		err = handleConnection(listener)
		if err != nil {
			log.Fatal(err)
		}
	}
}
