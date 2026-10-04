// Package mcpseat is ADR 0122's local MCP seat: a stdio MCP server that
// sits an MCP client (Claude Code, Codex, any other) in an ordinary guest
// seat at a real table, over the same WebSocket a browser uses.
//
// The seat gets no special access. It sees the filtered view and the
// closed legal-move list its seat is sent, nothing else, and it can do
// only what that list offers (owner decision 1). Trivial windows are
// answered by bot Layer A, imported from aiseat/rules rather than written
// again (§4), so only real choices reach the model. Every seat it takes
// carries the "AI agent" badge, declared at join (§7).
//
// # Layout
//
//   - transport.go is the ONLY file that imports the MCP SDK (§1).
//     It registers the tools and converts their results. Everything else
//     works in this package's own types, so the handlers are tested
//     without the SDK in the way.
//   - tools.go holds the ten tools' inputs and handlers.
//   - seat.go holds the seat's state: the session, the latest view, the
//     current decision window, and the autopilot that answers trivial
//     windows.
//   - conn.go is the WebSocket client and its reconnect ladder; httpapi.go
//     is the lobby's HTTP routes.
//   - render.go is the compact view: aiseat/boardtext's board, with the
//     table's own text wrapped and cut (§8), held to the §5 budgets.
//   - wire_join.go is the join body's agent declaration (ADR 0122 PR 2).
//     Every other frame and field is the protocol package's own type:
//     the ack, legal_moves_request and its reply, the cut report and the
//     open-set value marker (PR 5).
//
// # What it may import
//
// The binary runs on the owner's workstation and holds a socket to a
// server elsewhere, so there is no *game.Game in its process to read.
// The type rule says the same thing anyway: imports_test.go fails on a
// direct import of game, ws, lobby, actions, db, auth or aiseat/tiers.
package mcpseat
