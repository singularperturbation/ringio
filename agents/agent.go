package agents

import (
	"fmt"
	"strings"
	"time"

	"github.com/dullgiulio/ringbuf"
	"github.com/dullgiulio/ringio/config"
	"github.com/dullgiulio/ringio/msg"
)

type AgentType int

const (
	AgentTypeNull = iota
	AgentTypeCmd
	AgentTypePipe
)

const (
	AgentStatusNone = iota
	AgentStatusRunning
	AgentStatusKill
	AgentStatusFinished
)

const (
	AgentRoleSink = iota
	AgentRoleSource
	AgentRoleErrors
	AgentRoleLog
)

type AgentRole int
type AgentStatus int

func (s AgentStatus) IsFinished() bool {
	return s == AgentStatusFinished
}

func (s AgentStatus) String() string {
	switch s {
	case AgentStatusRunning:
		return "R"
	case AgentStatusKill:
		return "F"
	case AgentStatusFinished:
		return "F"
	}

	return "?"
}

type AgentMetadata struct {
	Id       int
	Role     AgentRole
	Status   AgentStatus
	Started  time.Time
	Finished time.Time
	Filter   *msg.Filter
}

type AgentDescr struct {
	Args   []string
	Meta   AgentMetadata
	Type   AgentType
	Filter *msg.Filter
}

func (a *AgentDescr) String() string {
	var args string

	flow := "->"

	if a.Type == AgentTypeCmd {
		args = strings.Join(a.Args, " ")
	} else if a.Type == AgentTypePipe {
		args = "[pipe]"
	}

	if a.Meta.Role == AgentRoleSource {
		flow = "<-"
	}

	if a.Meta.Role == AgentRoleSink &&
		a.Meta.Filter != nil {
		flow = fmt.Sprintf("-> [%s]", a.Meta.Filter.String())
	}

	return fmt.Sprintf("%d %s %s %s",
		a.Meta.Id, a.Meta.Status.String(), flow, args)
}

func (a *AgentDescr) Text() string {
	var started, finished string

	str := a.String()

	if a.Meta.Status != AgentStatusNone {
		started = fmt.Sprintf("  Started: %s\n", a.Meta.Started.Format("2006-01-02 15:04:05 -0700 MST"))
	}

	if a.Meta.Status == AgentStatusFinished {
		finished = fmt.Sprintf("  Finished: %s\n", a.Meta.Finished.Format("2006-01-02 15:04:05 -0700 MST"))
	}

	return fmt.Sprintf("%s\n%s%s", str, started, finished)
}

type Agent interface {
	Init()
	Stop()
	Cancel() error
	String() string
	Meta() *AgentMetadata
	Descr() AgentDescr
	InputToRingbuf(errors, output *ringbuf.Ringbuf)
	OutputFromRingbuf(stdout, errors, output *ringbuf.Ringbuf)
}

func (ac *Collection) _responseOk(a Agent) {
	resp := NewAgentMessageResponseBool()
	ac.SetAgentStatusFinished(a, &resp)
	resp.Get()
}

func (ac *Collection) inputToRingbuf(a Agent) {
	a.InputToRingbuf(ac.errors, ac.output)
	ac._responseOk(a)
}

func (ac *Collection) outputFromRingbuf(a Agent) {
	a.OutputFromRingbuf(ac.stdout, ac.errors, ac.output)
	ac._responseOk(a)
}

func (ac *Collection) errorsFromRingbuf(a Agent) {
	// We both read and write on errors.
	a.OutputFromRingbuf(ac.errors, ac.errors, ac.errors)
	ac._responseOk(a)
}

func (ac *Collection) logFromRingbuf(a Agent) {
	logring := config.GetLogRingbuf()

	a.OutputFromRingbuf(logring, logring, logring)
	ac._responseOk(a)
}

func (ac *Collection) runAgent(a Agent) {
	meta := a.Meta()
	meta.Started = time.Now()
	meta.Status = AgentStatusRunning

	go func(ac *Collection, a Agent) {
		switch meta.Role {
		case AgentRoleSource:
			ac.inputToRingbuf(a)
		case AgentRoleErrors:
			ac.errorsFromRingbuf(a)
		case AgentRoleSink:
			ac.outputFromRingbuf(a)
		case AgentRoleLog:
			ac.logFromRingbuf(a)
		}
	}(ac, a)
}
