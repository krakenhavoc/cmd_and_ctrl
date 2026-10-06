package protocol

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// ability_rows_view.go — #2219: the art tile's ability chips on the
// wire. A tile shows no rules text, so the server lists the object's
// non-keyword abilities by kind and the client draws one chip per kind
// with a count (⚡ triggered, ◆ static, ↻ activated) whose hover or
// focus lists the labels. The client never reads oracle text; every
// rule about what counts lives in game.AbilityRowsOf.

// AbilityRowView is one non-keyword ability of a card, as its tile
// chip lists it.
type AbilityRowView struct {
	// Kind is "triggered", "static" or "activated".
	Kind string `json:"kind"`
	// Label is the row's player-facing text: what the stack labels the
	// ability with for a triggered or activated row, a static's own
	// label, or what the effect is when the catalog gives it no text.
	Label string `json:"label"`
}

// viewOfAbilityRows projects game.AbilityRowsOf. Stamped by viewOfZone
// on the two zones a tile is drawn in, the battlefield and the hand,
// and nowhere else: a graveyard card is never a tile, and the list
// would only make every frame longer.
func viewOfAbilityRows(c game.Card) []AbilityRowView {
	rows := game.AbilityRowsOf(c)
	if len(rows) == 0 {
		return nil
	}
	out := make([]AbilityRowView, len(rows))
	for i, r := range rows {
		out[i] = AbilityRowView{Kind: string(r.Kind), Label: r.Label}
	}
	return out
}
