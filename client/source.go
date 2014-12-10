package client

import (
	"bufio"
	"fmt"
	"net/rpc"
	"os"

	"github.com/dullgiulio/ringio/agents"
	"github.com/dullgiulio/ringio/pipe"
	"github.com/dullgiulio/ringio/server"
	"github.com/dullgiulio/ringio/utils"
)

func addSourceAgentPipe(client *rpc.Client, response *server.RpcResp, pipeName string) {
	var id int

	p := pipe.New(pipeName)

	if err := p.OpenWriteErr(); err != nil {
		utils.Fatal(fmt.Errorf("Couldn't open pipe for writing: %s", err))
	}

	if err := client.Call("RpcServer.Add", &server.RpcReq{
		Agent: &agents.AgentDescr{
			Args: []string{pipeName},
			Meta: agents.AgentMetadata{Role: agents.AgentRoleSource},
			Type: agents.AgentTypePipe,
		},
	}, &id); err != nil {
		utils.Fatal(err)
	}

	// Write to pipe from stdin.
	r := bufio.NewReader(os.Stdin)

	if _, err := r.WriteTo(p); err != nil {
		utils.Fatal(err)
	}
}

func addSourceAgentCmd(client *rpc.Client, response *server.RpcResp, args []string) {
	var id int

	if err := client.Call("RpcServer.Add", &server.RpcReq{
		Agent: &agents.AgentDescr{
			Args: args,
			Meta: agents.AgentMetadata{Role: agents.AgentRoleSource},
			Type: agents.AgentTypeCmd,
		},
	}, &id); err != nil {
		utils.Fatal(err)
	}
}