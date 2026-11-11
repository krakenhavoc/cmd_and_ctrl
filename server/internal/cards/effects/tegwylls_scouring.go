package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tegwyll's Scouring — Sorcery {4}{B}{B} (Wilds of Eldraine Commander,
// #5368):
//
//	"You may cast this spell as though it had flash by tapping three
//	 untapped creatures you control with flying in addition to paying its
//	 other costs.
//	 Destroy all creatures. Create three 1/1 black Faerie Rogue creature
//	 tokens with flying."
//
// The flash is CR 601.3c's additional-cost form: an optional cost that
// taps three untapped flyers and opens the instant-speed window for that
// cast alone (cmdr2B11FlashByTapping). The sweep is Day of Judgment's,
// simultaneous and with no regeneration rider; the tokens are made after
// it, so they survive it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "58ebaaba-7e35-4588-9679-23b22e39789b",
		Name:          "Tegwyll's Scouring",
		Purpose:       game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{cmdr2B11FlashByTapping(3, "three untapped creatures you control with flying", HasKeyword("flying"))},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := wrathDestroyAllCreatures(item, ctx); err != nil {
				return err
			}
			return CreateToken{Template: FaerieRogueToken(), N: 3}.Apply(ctx)
		},
	})
}
