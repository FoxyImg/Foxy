package cli

import "foxy/internal/server"

type ServerCmd struct {
}

func (serverCmd *ServerCmd) Run(ctx *Context) error {
	server.StartServer()
	return nil
}
