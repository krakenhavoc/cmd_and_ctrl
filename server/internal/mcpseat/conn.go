package mcpseat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// The reconnect ladder, the browser's (docs/protocol.md "The reconnect
// ladder"): 500 ms doubling to 30 s, half-jittered, and a GET /me after
// three failed dials to tell a dead session from a restarting server.
const (
	reconnectBase         = 500 * time.Millisecond
	reconnectCap          = 30 * time.Second
	deadSessionCheckAfter = 3
	maxInboundFrame       = 64 << 20
	closeGrace            = time.Second
	writeTimeout          = 10 * time.Second
)

// frameSink receives what the connection reads. Every method is called on
// the connection's read goroutine, one at a time.
type frameSink interface {
	connected()
	frame(protocol.Frame)
	// dropped is a non-terminal loss: anything in flight is unknown.
	dropped()
	// ended is terminal: the game is gone, or the session is dead.
	ended(reason string)
}

// wsConn is the seat's WebSocket to the server, redialled on the
// browser's ladder. It authenticates with `Authorization: Bearer` on the
// upgrade, never `?token=` (§8), and sends no Origin header, which the
// server admits for a non-browser client.
type wsConn struct {
	wsURL  string
	origin string
	token  string
	dialer *websocket.Dialer
	api    *api
	sink   frameSink
	log    *slog.Logger

	wmu  sync.Mutex
	conn *websocket.Conn

	cancel context.CancelFunc
	done   chan struct{}
}

// wsURLFor is <origin>/ws?game=<id>&player=<id>, with ws or wss.
func wsURLFor(origin, gameID, playerID string) (string, error) {
	u, err := url.Parse(origin)
	if err != nil {
		return "", err
	}
	switch u.Scheme {
	case "https":
		u.Scheme = "wss"
	case "http":
		u.Scheme = "ws"
	default:
		return "", fmt.Errorf("origin %q: not http(s)", origin)
	}
	u.Path = "/ws"
	q := url.Values{}
	q.Set("game", gameID)
	q.Set("player", playerID)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func newWSConn(wsURL, origin, token string, dialer *websocket.Dialer, a *api, sink frameSink, log *slog.Logger) *wsConn {
	if dialer == nil {
		dialer = &websocket.Dialer{HandshakeTimeout: 15 * time.Second, Proxy: http.ProxyFromEnvironment}
	}
	return &wsConn{wsURL: wsURL, origin: origin, token: token, dialer: dialer, api: a, sink: sink, log: log}
}

// start runs the dial-read-redial loop until close or a terminal end.
func (c *wsConn) start() {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	c.done = make(chan struct{})
	go c.loop(ctx)
}

// close ends the connection for good with a normal closure.
func (c *wsConn) close() {
	if c.cancel == nil {
		return
	}
	c.cancel()
	c.wmu.Lock()
	if c.conn != nil {
		_ = c.conn.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "leaving"), time.Now().Add(closeGrace))
		_ = c.conn.Close()
	}
	c.wmu.Unlock()
	<-c.done
}

// reconnectDelay is attempt n's delay: uniform in [nominal/2, nominal].
func reconnectDelay(n int) time.Duration {
	nominal := reconnectBase
	for i := 0; i < n && nominal < reconnectCap; i++ {
		nominal *= 2
	}
	if nominal > reconnectCap {
		nominal = reconnectCap
	}
	half := nominal / 2
	return half + time.Duration(rand.Int64N(int64(half)+1))
}

func (c *wsConn) loop(ctx context.Context) {
	defer close(c.done)
	failures := 0
	checked := false
	attempt := 0
	for {
		if ctx.Err() != nil {
			return
		}
		h := http.Header{}
		h.Set("Authorization", "Bearer "+c.token)
		conn, resp, err := c.dialer.DialContext(ctx, c.wsURL, h)
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			if resp != nil && resp.StatusCode == http.StatusNotFound {
				c.sink.ended("the server no longer has this game")
				return
			}
			failures++
			c.log.Warn("websocket dial failed", "attempt", failures, "err", err)
			if failures >= deadSessionCheckAfter && !checked {
				checked = true
				if dead := c.sessionDead(ctx); dead {
					c.sink.ended("the session ended (the server no longer accepts it); the seat is lost to this binary, as for any guest")
					return
				}
			}
			if sleepCtx(ctx, reconnectDelay(attempt)) != nil {
				return
			}
			attempt++
			continue
		}
		failures, checked, attempt = 0, false, 0
		conn.SetReadLimit(maxInboundFrame)
		c.wmu.Lock()
		c.conn = conn
		c.wmu.Unlock()
		c.sink.connected()

		code := c.read(conn)

		c.wmu.Lock()
		c.conn = nil
		c.wmu.Unlock()
		_ = conn.Close()
		if ctx.Err() != nil {
			c.sink.dropped()
			return
		}
		if code == websocket.CloseNormalClosure {
			c.sink.ended("the server closed the table connection (the game was deleted, or the session revoked)")
			return
		}
		c.sink.dropped()
		if sleepCtx(ctx, reconnectDelay(attempt)) != nil {
			return
		}
		attempt++
	}
}

// read reads frames until the socket fails, and returns the close code
// (1006 when there was no close frame).
func (c *wsConn) read(conn *websocket.Conn) int {
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			var ce *websocket.CloseError
			if errors.As(err, &ce) {
				return ce.Code
			}
			return websocket.CloseAbnormalClosure
		}
		var f protocol.Frame
		if err := json.Unmarshal(raw, &f); err != nil {
			c.log.Warn("unreadable frame from the server", "err", err)
			continue
		}
		c.sink.frame(f)
	}
}

// sessionDead asks GET /me. Only a definite 401 is dead: a 5xx, a timeout
// or a network failure is a server that may only be restarting.
func (c *wsConn) sessionDead(ctx context.Context) bool {
	cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	_, err := c.api.me(cctx, c.origin, c.token)
	var ae *apiError
	return errors.As(err, &ae) && ae.Status == http.StatusUnauthorized
}

// errNotConnected is a send while the socket is down.
var errNotConnected = errors.New("not connected to the table right now (reconnecting)")

// send writes one frame.
func (c *wsConn) send(kind, id string, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	f := protocol.Frame{V: protocol.Version, Kind: protocol.Kind(kind), ID: id, Payload: raw}
	c.wmu.Lock()
	defer c.wmu.Unlock()
	if c.conn == nil {
		return errNotConnected
	}
	_ = c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	if err := c.conn.WriteJSON(f); err != nil {
		return fmt.Errorf("send %s: %w", strings.TrimSpace(kind), err)
	}
	return nil
}
