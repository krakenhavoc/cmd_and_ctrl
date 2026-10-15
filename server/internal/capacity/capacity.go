// Package capacity is how much the server agrees to carry (ADR 0148,
// #2977): the weighted table-load cap, the WebSocket connection cap
// and the table-creation rates, all read from the environment with
// defaults taken from the prod load test (#2978).
//
// Prod is one 2 vCPU / 4 GB VM. The load test found it collapses at
// about 31-36 four-bot tables, and on boot it restored the tables
// that had killed it. A cap is the only thing between one popular
// post and that loop, so it is on by default; "0" switches a cap off.
//
// The package holds policy and arithmetic only. Where a cap is
// enforced is the lobby's (table load, creation) and the hub's
// (connections); this package imports nothing from either, so both
// can use it.
package capacity

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CodeServerFull is the stable `code` of the JSON error body (and of
// the WebSocket error frame) a refused request carries. The client
// matches on it, never on the message.
const CodeServerFull = "server_full"

// Message is the sentence a player sees. One string, used by the HTTP
// error, the WebSocket frame and (as a fallback) the client, so the
// three never drift.
const Message = "The server is full right now. Please try again in a few minutes."

// RetryAfterSeconds is the Retry-After a refusal carries: long enough
// that a retry loop does not make a full server fuller, short enough
// that a table freed in the meantime is not missed for long.
const RetryAfterSeconds = 30

// ErrServerFull is what a refused request returns. Its text is the
// player-facing Message; the HTTP layer maps it to 503 with
// CodeServerFull.
//
//nolint:staticcheck // ST1005: Message is a player-facing sentence, so it ends in a full stop on purpose.
var ErrServerFull = errors.New(Message)

// Environment variable names. See docs/lobby.md and AGENTS.md.
const (
	EnvMaxLoad       = "CMDCTRL_MAX_TABLE_LOAD"
	EnvWeightTable   = "CMDCTRL_LOAD_WEIGHT_TABLE"
	EnvWeightHuman   = "CMDCTRL_LOAD_WEIGHT_HUMAN_SEAT"
	EnvWeightBot     = "CMDCTRL_LOAD_WEIGHT_BOT_SEAT"
	EnvMaxConns      = "CMDCTRL_MAX_WS_CONNECTIONS"
	EnvAccountBurst  = "CMDCTRL_CREATE_ACCOUNT_BURST"
	EnvAccountEvery  = "CMDCTRL_CREATE_ACCOUNT_INTERVAL"
	EnvIPBurst       = "CMDCTRL_CREATE_IP_BURST"
	EnvIPEvery       = "CMDCTRL_CREATE_IP_INTERVAL"
	defaultMaxLoad   = 180
	defaultTable     = 2
	defaultHuman     = 1
	defaultBot       = 2
	defaultMaxConns  = 400
	defaultAccBurst  = 3
	defaultAccEvery  = 30 * time.Second
	defaultIPBurst   = 6
	defaultIPEvery   = 20 * time.Second
	maxConfiguredInt = 1 << 20
)

// Limits are the server's caps. The zero value has every cap off and
// every creation limit at its default; use FromEnv or Defaults for a
// deployment's.
type Limits struct {
	// MaxLoad is the weighted table-load ceiling: the sum of
	// TableCost over every unended, unarchived table. 0 is no cap.
	// It is a SOFT cap: it gates the creation of load (a new table, a
	// practice table, a bot seat), never a seat or an action at a
	// table that already exists.
	MaxLoad int
	// TableWeight is what an open table costs before anyone sits at
	// it: the room, its restore point writes and its idle memory.
	TableWeight int
	// HumanSeatWeight is what a human (or agent) seat adds.
	HumanSeatWeight int
	// BotSeatWeight is what a bot seat adds. Larger than a human
	// seat's: a bot runs a decision loop and a legal-move enumerator
	// for the whole game, which is what saturated the CPU in the
	// load test.
	BotSeatWeight int
	// MaxConns is the ceiling on live WebSocket connections. 0 is no
	// cap. Connections of a player already seated at a table are
	// exempt (see ws.Hub.SetMaxConnections).
	MaxConns int

	// AccountBurst and AccountEvery are the per-account table-creation
	// bucket: AccountBurst creations at once, then one more every
	// AccountEvery. IPBurst and IPEvery are the same per client IP,
	// which is what bounds guests (who cannot create a table, but do
	// open practice tables) and anyone minting accounts.
	AccountBurst int
	AccountEvery time.Duration
	IPBurst      int
	IPEvery      time.Duration
}

// Defaults are the shipped limits. See ADR 0148 for the derivation:
// a four-human table costs 6, a four-bot table 10, and 180 is both
// the owner's 30-game target of human tables and 18 four-bot tables,
// well under the ~31 the load test collapsed at.
func Defaults() Limits {
	return Limits{
		MaxLoad:         defaultMaxLoad,
		TableWeight:     defaultTable,
		HumanSeatWeight: defaultHuman,
		BotSeatWeight:   defaultBot,
		MaxConns:        defaultMaxConns,
		AccountBurst:    defaultAccBurst,
		AccountEvery:    defaultAccEvery,
		IPBurst:         defaultIPBurst,
		IPEvery:         defaultIPEvery,
	}
}

// WithCreateDefaults returns l with any zero table-creation field set
// to its default, so a caller that built only part of a Limits (a test,
// or a Config that never named the rates) creates at the shipped rate
// rather than with a zero-length interval.
func (l Limits) WithCreateDefaults() Limits {
	if l.AccountBurst <= 0 {
		l.AccountBurst = defaultAccBurst
	}
	if l.AccountEvery <= 0 {
		l.AccountEvery = defaultAccEvery
	}
	if l.IPBurst <= 0 {
		l.IPBurst = defaultIPBurst
	}
	if l.IPEvery <= 0 {
		l.IPEvery = defaultIPEvery
	}
	return l
}

// TableCost is what a table with these seats costs against MaxLoad.
func (l Limits) TableCost(humans, bots int) int {
	return l.TableWeight + humans*l.HumanSeatWeight + bots*l.BotSeatWeight
}

// LoadCapped reports whether the table-load cap is on.
func (l Limits) LoadCapped() bool { return l.MaxLoad > 0 }

// Fits reports whether adding cost to a load of used stays within the
// cap. Always true with the cap off.
func (l Limits) Fits(used, cost int) bool {
	return !l.LoadCapped() || used+cost <= l.MaxLoad
}

// FromEnv reads the limits from getenv, defaulting every unset one. A
// malformed value is an error naming the variable: a typo in a cap
// must stop the boot rather than quietly leave the server uncapped.
func FromEnv(getenv func(string) string) (Limits, error) {
	l := Defaults()
	ints := []struct {
		name string
		dst  *int
		min  int
	}{
		{EnvMaxLoad, &l.MaxLoad, 0},
		{EnvWeightTable, &l.TableWeight, 0},
		{EnvWeightHuman, &l.HumanSeatWeight, 0},
		{EnvWeightBot, &l.BotSeatWeight, 0},
		{EnvMaxConns, &l.MaxConns, 0},
		{EnvAccountBurst, &l.AccountBurst, 1},
		{EnvIPBurst, &l.IPBurst, 1},
	}
	for _, f := range ints {
		raw := strings.TrimSpace(getenv(f.name))
		if raw == "" {
			continue
		}
		n, err := strconv.Atoi(raw)
		if err != nil || n < f.min || n > maxConfiguredInt {
			return Limits{}, fmt.Errorf("%s=%q: want an integer from %d to %d", f.name, raw, f.min, maxConfiguredInt)
		}
		*f.dst = n
	}
	durs := []struct {
		name string
		dst  *time.Duration
	}{
		{EnvAccountEvery, &l.AccountEvery},
		{EnvIPEvery, &l.IPEvery},
	}
	for _, f := range durs {
		raw := strings.TrimSpace(getenv(f.name))
		if raw == "" {
			continue
		}
		d, err := time.ParseDuration(raw)
		if err != nil || d <= 0 {
			return Limits{}, fmt.Errorf("%s=%q: want a positive Go duration such as 30s", f.name, raw)
		}
		*f.dst = d
	}
	return l, nil
}

// Usage is a reading of what the server carries now, against the
// limits. Zero limits mean the cap is off.
type Usage struct {
	Load      int
	LoadLimit int
	Conns     int
	ConnLimit int
}
