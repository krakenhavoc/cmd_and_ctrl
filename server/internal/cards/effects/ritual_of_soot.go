package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ritual of Soot — Sorcery {2}{B}{B}:
//
//	"Destroy all creatures with mana value 3 or less."
//
// The asymmetric wrath. Four mana that clears a token board, a
// Birds-and-Elves mana base and every hatebear at the table while
// leaving your own six-drops standing — which is the entire reason
// to play it over Damnation at the same cost.
//
// Mana value is read off the PRINTED mana cost, so a token (no mana
// cost, mana value 0) is destroyed and an X spell's permanent counts
// X as 0 on the battlefield (CR 202.3e). Both are correct and both
// follow from ManaValueLE, which is the same predicate a targeting
// clause would use.
func init() {
	Register(Spec{
		OracleID: "29e9cf1c-a6bd-4bee-9000-ac1b4e19d6b0",
		Name:     "Ritual of Soot",
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: And(Creature(), ManaValueLE(3))}.Apply(ctx)
		},
	})
}
