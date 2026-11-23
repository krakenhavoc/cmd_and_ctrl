package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bender's Waterskin — Artifact {3} (EDHREC rank 579):
//
//	"Untap this artifact during each other player's untap step.
//	 {T}: Add one mana of any color."
//
// A three-mana any-colour rock that taps once per PLAYER's turn
// rather than once per rotation — four activations a lap at a full
// Commander table, which is the whole card. Its untap clause is the
// same permission the Muse has, aimed at itself alone.
//
// The mana ability is the Birds shape at the printed width: any
// colour, not narrowed to the commander's identity, because the card
// does not say "of any color in your commander's identity".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3bc46467-323f-4cbf-9d0a-f3d2ae3c3a34",
		Name:         "Bender's Waterskin",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		UntapStep: []game.UntapStepPermission{
			effects.UntapSelfDuringEachOtherPlayersUntapStep(
				"Bender's Waterskin — untap this artifact"),
		},
	})
}
