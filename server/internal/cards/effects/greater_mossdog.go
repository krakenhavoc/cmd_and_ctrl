package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Greater Mossdog — Creature — Plant Dog {3}{G}, 3/3 (Ravnica: City of
// Guilds, #5600):
//
//	"Dredge 3 (If you would draw a card, you may mill three cards instead.
//	 If you do, return this card from your graveyard to your hand.)"
//
// Dredge is the engine's draw replacement (CR 702.52, dredge.go): while the
// card is in its owner's graveyard and they have at least three cards in
// their library, each draw offers to mill three and return it instead.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aedc3e69-3163-4b8d-97b2-0ad92f28a6dc",
		Name:         "Greater Mossdog",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{Dredge(3)},
	})
}
