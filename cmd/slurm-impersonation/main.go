package main

import (
	"encoding/binary"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/s-bauer/slurm-k8s/internal/proto/pb"
	"github.com/s-bauer/slurm-k8s/internal/util"
	"log"
	"net"
)

const (
	unixSocketPath = "/tmp/slurm-impersonation.sock"
)

func passOrDie(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func joinNewNode(joinNode *pb.JoinNode) error {
	labels := fmt.Sprintf("userNodeFor=%s", joinNode.Uid)
	taints := fmt.Sprintf("userNodeFor=%s:NoSchedule", joinNode.Uid)
	cmdResult, err := util.RunCommand(
		"sudo",
		"-u",
		fmt.Sprintf("\\#%s", joinNode.Uid),
		"srun",
		"--nodes",
		"1",
		"--exclusive",
		"--comment",
		fmt.Sprintf("user-node-%s", joinNode.ServerApiEndpoint),
		"--k8s-join-cluster",
		"--k8s-join-token",
		joinNode.Token,
		"--k8s-join-cert-hash",
		joinNode.CertHash,
		"--k8s-join-api-server",
		joinNode.ServerApiEndpoint,
		"--k8s-kublet-node-labels",
		labels,
		"--k8s-kublet-node-taints",
		taints,
	)
	if cmdResult.ExitCode != 0 {
		log.Print(cmdResult.Stdout)
		log.Print(cmdResult.Stderr)
		log.Fatal("Command Failed")
	}
	return err
}

func stopSlurmJob(joinNode *pb.JoinNode) {

}

func handleConnection(listener *net.UnixListener) error {
	connection, err := listener.AcceptUnix()
	if err != nil {
		log.Fatalf("unable to accept connection: %v", err)
	}
	defer connection.Close()

	var messageLength int32
	// encoding/binary is quite clever! Notice that we're reading from the socket and encoding directly into the int32 here. Exactly 4 bytes (the size of messageLength) will be read.
	err = binary.Read(connection, binary.BigEndian, &messageLength)
	passOrDie(err)
	log.Printf("Message Length: %v", messageLength)

	joinNode := new(pb.JoinNode)

	buf := make([]byte, messageLength)
	err = binary.Read(connection, binary.BigEndian, buf)
	passOrDie(err)
	err = proto.Unmarshal(buf, joinNode)
	passOrDie(err)

	if !joinNode.IsDelete {
		joinNewNode(joinNode)
	} else {
		stopSlurmJob(joinNode)
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
