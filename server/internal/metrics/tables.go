package metrics

import "github.com/prometheus/client_golang/prometheus"

// The tables collector: the lobby's tables and the hub's sockets, read
// at scrape time (ADR 0123 §3). It reports
//
//	cmdctrl_games{state,archived}       tables in the lobby, practice excluded
//	cmdctrl_practice_games              open practice (tutorial) tables
//	cmdctrl_seats{kind}                 seats at active tables
//	cmdctrl_seats_connected{kind,account}  of those, seats with a live socket
//	cmdctrl_spectators_connected        live read-only sockets
//	cmdctrl_ws_connections{role}        live sockets by role
//
// "Active tables" are the non-practice tables in state active that are
// not archived. Archiving a running table is how an operator closes it
// (it counts in cmdctrl_games_ended_total as closed), so its seats are
// no longer at an active table.
//
// # Locking
//
// A scrape reads the hub, then the lobby, through one accessor each
// (SocketSource, TableSource). Each accessor copies what it needs under
// its own lock and releases it before returning, so the collector never
// holds two of the server's locks at once and takes no room or game
// lock at all. The two reads are not one instant: a socket that
// connects between them can be counted against a table the lobby read
// after it. That is a scrape's worth of skew on a gauge, and the next
// scrape is right.

// Seat kinds: the kind label of cmdctrl_seats and
// cmdctrl_seats_connected (which has no bots: a bot seat plays in
// process, with no socket).
const (
	SeatHuman = "human"
	SeatBot   = "bot"
	SeatAgent = "agent"
)

// Accounts: the account label of cmdctrl_seats_connected.
const (
	AccountSignedIn = "signed_in"
	AccountGuest    = "guest"
)

// Table states: the state label of cmdctrl_games, game.State's values.
var tableStateLabels = []string{"lobby", "active", "ended"}

var archivedLabels = []string{"false", "true"}

var seatKindLabels = []string{SeatHuman, SeatBot, SeatAgent}

var connectedKindLabels = []string{SeatHuman, SeatAgent}

var accountLabels = []string{AccountSignedIn, AccountGuest}

// Key is a game or player UUID as the collector compares them. The
// sources convert uuid.UUID (a [16]byte) directly; this package
// imports no UUID library.
type Key [16]byte

// Table is one lobby table as the collector reads it.
type Table struct {
	Game     Key
	State    string // "lobby" | "active" | "ended"
	Archived bool
	Practice bool
	Seats    []Seat
}

// Seat is one seat at a Table.
type Seat struct {
	Player Key
	// Kind is SeatHuman, SeatBot or SeatAgent.
	Kind string
	// SignedIn is true for a seat held by a signed-in person (a users
	// row), false for a guest.
	SignedIn bool
}

// Socket is one live WebSocket connection.
type Socket struct {
	Game   Key
	Player Key // zero for a connection with no seat
	// Role is RoleSeat, RoleSpectator or RoleAdmin (WSRole).
	Role string
}

// TableSource is the lobby: every table it holds, practice tables
// included, copied under its lock.
type TableSource interface {
	MetricsTables() []Table
}

// SocketSource is the hub: every live connection, copied under its
// lock.
type SocketSource interface {
	MetricsSockets() []Socket
}

var (
	gamesDesc = prometheus.NewDesc("cmdctrl_games",
		"Tables in the lobby registry, by state and whether archived. Practice tables are not counted.",
		[]string{"state", "archived"}, nil)
	practiceGamesDesc = prometheus.NewDesc("cmdctrl_practice_games",
		"Open practice (tutorial) tables.", nil, nil)
	seatsDesc = prometheus.NewDesc("cmdctrl_seats",
		"Seats at active tables (not practice, not archived), by kind.",
		[]string{"kind"}, nil)
	seatsConnectedDesc = prometheus.NewDesc("cmdctrl_seats_connected",
		"Seats at active tables with at least one live socket now, by kind and whether a signed-in person holds the seat: players active currently.",
		[]string{"kind", "account"}, nil)
	spectatorsConnectedDesc = prometheus.NewDesc("cmdctrl_spectators_connected",
		"Live read-only (spectator) sockets.", nil, nil)
	wsConnectionsDesc = prometheus.NewDesc("cmdctrl_ws_connections",
		"Live WebSocket connections, by role.",
		[]string{"role"}, nil)
)

type tablesCollector struct {
	tables  TableSource
	sockets SocketSource
}

// NewTablesCollector returns the collector over the lobby and the hub.
// main.go registers it on Registry with the live objects.
func NewTablesCollector(tables TableSource, sockets SocketSource) prometheus.Collector {
	return &tablesCollector{tables: tables, sockets: sockets}
}

func (c *tablesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- gamesDesc
	ch <- practiceGamesDesc
	ch <- seatsDesc
	ch <- seatsConnectedDesc
	ch <- spectatorsConnectedDesc
	ch <- wsConnectionsDesc
}

type seatKey struct{ game, player Key }

// GamesKey is one cmdctrl_games series: a table state and whether the
// table is archived.
type GamesKey struct {
	State    string
	Archived bool
}

// ConnectedKey is one cmdctrl_seats_connected series: a seat kind and
// whether a signed-in person holds the seat (AccountSignedIn or
// AccountGuest).
type ConnectedKey struct {
	Kind    string
	Account string
}

// Counts is what the tables collector reports, before it is turned
// into series. Every map is keyed by a label value; a key that is not
// there counts zero.
type Counts struct {
	// Games counts the non-practice tables by state and archived
	// (cmdctrl_games).
	Games map[GamesKey]int
	// Practice counts the practice tables (cmdctrl_practice_games).
	Practice int
	// Seats counts the seats at active tables by kind (cmdctrl_seats).
	Seats map[string]int
	// Connected counts the seats at active tables with a live seat
	// socket, by kind and account (cmdctrl_seats_connected). Bot seats
	// are never in it.
	Connected map[ConnectedKey]int
	// Roles counts the live sockets by role (cmdctrl_ws_connections);
	// Roles[RoleSpectator] is cmdctrl_spectators_connected.
	Roles map[string]int
}

// PlayersConnected is the Overview's "Players connected" tile: every
// seat at an active table with a live socket, human or agent, signed
// in or guest (the sum of cmdctrl_seats_connected).
func (c Counts) PlayersConnected() int {
	n := 0
	for _, v := range c.Connected {
		n += v
	}
	return n
}

// Tally counts tables and sockets the way the tables collector
// reports them. It is the one definition of those numbers: the
// collector emits its gauges from it, and the admin views' Live now
// (ADR 0124 §5) computes its totals with it, so the page and the
// Grafana tiles cannot count differently. It is pure: it reads only
// its arguments and takes no lock.
func Tally(tables []Table, sockets []Socket) Counts {
	out := Counts{
		Games:     map[GamesKey]int{},
		Seats:     map[string]int{},
		Connected: map[ConnectedKey]int{},
		Roles:     map[string]int{},
	}
	live := map[seatKey]bool{}
	for _, s := range sockets {
		out.Roles[s.Role]++
		if s.Role == RoleSeat {
			live[seatKey{s.Game, s.Player}] = true
		}
	}
	for _, t := range tables {
		if t.Practice {
			out.Practice++
			continue
		}
		out.Games[GamesKey{t.State, t.Archived}]++
		if t.State != "active" || t.Archived {
			continue
		}
		for _, s := range t.Seats {
			out.Seats[s.Kind]++
			if s.Kind == SeatBot || !live[seatKey{t.Game, s.Player}] {
				continue
			}
			account := AccountGuest
			if s.SignedIn {
				account = AccountSignedIn
			}
			out.Connected[ConnectedKey{s.Kind, account}]++
		}
	}
	return out
}

func (c *tablesCollector) Collect(ch chan<- prometheus.Metric) {
	// The hub first, then the lobby, each through its own accessor:
	// never two locks at once (see Locking above).
	sockets := c.sockets.MetricsSockets()
	tables := c.tables.MetricsTables()
	n := Tally(tables, sockets)

	// Every series in each closed set is reported, zeros included, and
	// nothing outside it: a state or kind the sets do not name is not
	// a series.
	for _, st := range tableStateLabels {
		for _, a := range archivedLabels {
			v := n.Games[GamesKey{st, a == "true"}]
			ch <- prometheus.MustNewConstMetric(gamesDesc, prometheus.GaugeValue, float64(v), st, a)
		}
	}
	ch <- prometheus.MustNewConstMetric(practiceGamesDesc, prometheus.GaugeValue, float64(n.Practice))
	for _, k := range seatKindLabels {
		ch <- prometheus.MustNewConstMetric(seatsDesc, prometheus.GaugeValue, float64(n.Seats[k]), k)
	}
	for _, k := range connectedKindLabels {
		for _, a := range accountLabels {
			ch <- prometheus.MustNewConstMetric(seatsConnectedDesc, prometheus.GaugeValue,
				float64(n.Connected[ConnectedKey{k, a}]), k, a)
		}
	}
	ch <- prometheus.MustNewConstMetric(spectatorsConnectedDesc, prometheus.GaugeValue, float64(n.Roles[RoleSpectator]))
	for _, r := range roleLabels {
		ch <- prometheus.MustNewConstMetric(wsConnectionsDesc, prometheus.GaugeValue, float64(n.Roles[r]), r)
	}
}
