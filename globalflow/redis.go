package globalflow

import (
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

		message := CommandMessage{
			Command:    "set",
			Arguments:  []string{string(cmd.Args[1]), string(cmd.Args[2])},
			Time:       server.clock.Get(),
			Originator: server.container.Configuration.NodeID,
		}

		server.processCommand(message)
		server.broadcast(message)

		conn.WriteString("OK")
	}
}
