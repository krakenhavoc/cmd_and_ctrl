package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mastery of the Unseen — Enchantment {1}{W} (Murders at Karlov Manor
// Commander):
//
//	"Whenever a permanent you control is turned face up, you gain 1 life for
//	 each creature you control.
//	 {3}{W}: Manifest the top card of your library."
//
// Any permanent turning face up counts, a disguised land included. The life
// is counted as the trigger resolves, so the creature that was just turned
// up is counted if it is still a creature you control (CR 708.8).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ae0701c1-7a99-46fe-bcb1-ea33ad2a0fc9",
		Name:         "Mastery of the Unseen",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventTurnedFaceUp, APermanentYouControlWasTurnedFaceUp,
				"Mastery of the Unseen — you gain 1 life for each creature you control",
				cmdr2S37GainLifePerCreature),
		},
		Activated: []ActivatedAbility{{
			Label:  "{3}{W}: Manifest the top card of your library.",
			Cost:   ManaCost("{3}{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error { return Manifest{N: 1}.Apply(NewContext(g, item)) },
		}},
	})
}
