package globalflow

import (
	"fmt"
	"github.com/tidwall/redcon"
	"globalflow/globalflow/db"
	"strings"
	"time"
)

func (server *Server) Redis(conn redcon.Conn, cmd redcon.Command) {
	switch strings.ToLower(string(cmd.Args[0])) {
	default:
		conn.WriteError("ERR unknown command '" + string(cmd.Args[0]) + "'")

	case "get":
		if len(cmd.Args) != 2 {
			conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
			return
		}

		key := string(cmd.Args[1])

		v, err := server.db.Get(db.Time(time.Now().UnixMilli()), key)
		if db.IsErrorNotFound(err) {
			conn.WriteNull()
			return
		}
		if err != nil {
			conn.WriteError("ERR " + err.Error())
			return
		}

		conn.WriteString(v)

		return

	case "set":
		if len(cmd.Args) != 3 {
			conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
			return
		}

		args := make([]string, len(cmd.Args)-1)

		for i := 1; i < len(cmd.Args); i++ {
			args[i-1] = string(cmd.Args[i])
		}

		message := server.NewCommandMessage(
			string(cmd.Args[0]),
			args,
		)

		server.processCommand(message)
		err := server.broadcast(message)
		if err != nil {
			conn.WriteError("ERR " + err.Error())
			return
		}

		conn.WriteString("OK")

		return

	case "del":
		if len(cmd.Args) != 2 {
			conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
			return
		}

		args := make([]string, len(cmd.Args)-1)

		for i := 1; i < len(cmd.Args); i++ {
			args[i-1] = string(cmd.Args[i])
		}

		message := server.NewCommandMessage(
			string(cmd.Args[0]),
			args,
		)

		server.processCommand(message)
		err := server.broadcast(message)
		if err != nil {
			conn.WriteError("ERR " + err.Error())
			return
		}

		conn.WriteString("OK")

		return

	case "info":
		conn.WriteBulkString(fmt.Sprintf("peers:%d\r\n", len(server.gossip.Members())))
	}
}
