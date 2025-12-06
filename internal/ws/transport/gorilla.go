package transport

import (
    "net/http"
    "github.com/gorilla/websocket"
)

type gorillaConn struct {
    *websocket.Conn
}

func (c *gorillaConn) WriteJSON(v interface{
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return ErrClosed
	}
	return c.conn.WriteJSON(v)
}
}) error { return c.Conn.WriteJSON(v) }
func (c *gorillaConn) Close() error { return c.Conn.Close() }

type GorillaUpgrader struct {
    Upgrader websocket.Upgrader
}

func (g *GorillaUpgrader) Upgrade(w http.ResponseWriter, r *http.Request) (Conn, error) {
    conn, err := g.Upgrader.Upgrade(w, r, nil)
    if err != nil {
        return nil, err
    }
    return &gorillaConn{Conn: conn}, nil
}
