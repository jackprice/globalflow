package globalflow

import (
	"github.com/tidwall/redcon"
	bolt "go.etcd.io/bbolt"
	"strings"
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

		err := server.DB.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("DATA"))
			v := b.Get([]byte(key))

			if v == nil {
				conn.WriteNull()
				return nil
			}

			conn.WriteString(string(v))

			return nil
		})
		if err != nil {
			conn.WriteError("ERR " + err.Error())
			return
		}
	}
}
