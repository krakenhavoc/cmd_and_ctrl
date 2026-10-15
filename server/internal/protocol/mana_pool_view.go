package protocol

import (
	"encoding/json"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ManaPoolRun is one run of PlayerView.ManaPool (#3041): Count mana
// of one colour, adjacent in the pool.
type ManaPoolRun struct {
	Color string `json:"color"`
	Count int    `json:"count"`
}

// UnmarshalJSON also reads the pre-#3041 wire form, one bare symbol
// string per mana, as a run of one: a view recorded before runs (a bot
// suite position, a decision log, an arena replay) still decodes.
func (r *ManaPoolRun) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '"' {
		var color string
		if err := json.Unmarshal(data, &color); err != nil {
			return err
		}
		*r = ManaPoolRun{Color: color, Count: 1}
		return nil
	}
	type plain ManaPoolRun
	return json.Unmarshal(data, (*plain)(r))
}

// ManaPoolTotal is how many mana a projected pool holds.
func ManaPoolTotal(runs []ManaPoolRun) int {
	n := 0
	for _, r := range runs {
		n += r.Count
	}
	return n
}

// ManaPoolOf builds a projected pool from one symbol per mana, in
// order, folding adjacent symbols of one colour into a run. For a
// caller that thinks in mana, mostly tests and fixtures.
func ManaPoolOf(colors ...string) []ManaPoolRun {
	var out []ManaPoolRun
	for _, c := range colors {
		if n := len(out); n > 0 && out[n-1].Color == c {
			out[n-1].Count++
			continue
		}
		out = append(out, ManaPoolRun{Color: c, Count: 1})
	}
	return out
}

// manaPoolRunsOf projects an engine pool onto the wire: its runs in
// order, with adjacent runs of the same colour folded together, since
// the wire tells mana apart by colour alone. Nil for an empty pool.
// It costs O(runs), not O(mana).
func manaPoolRunsOf(pool game.ManaPool) []ManaPoolRun {
	var out []ManaPoolRun
	for _, t := range pool {
		if n := len(out); n > 0 && out[n-1].Color == t.Color {
			out[n-1].Count += t.N()
			continue
		}
		out = append(out, ManaPoolRun{Color: t.Color, Count: t.N()})
	}
	return out
}
