package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gandalf, White Rider — Legendary Creature — Avatar Wizard {3}{W}, 3/3 (The
// Lord of the Rings: Tales of Middle-earth Commander):
//
//	"Vigilance
//	 Whenever you cast a spell, creatures you control get +1/+0 until end of
//	 turn. Scry 1.
//	 When Gandalf dies, you may put it into its owner's library fifth from
//	 the top."
//
// Vigilance is printed. The cast watch pumps the creatures its controller
// controls as the ability resolves (CR 611.2c), then scries 1 (Purpose
// declares the one card). The death trigger is Braided Quipu's positioned
// landing (TuckToLibraryAtDepthForEffect) behind a "you may", for a Gandalf
// still in its owner's graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7a80212a-aae3-483b-abb6-7271a62dff8a",
		Name:            "Gandalf, White Rider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Purpose:         game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.ByYou, "Gandalf, White Rider — creatures you control get +1/+0 until end of turn; scry 1",
				effects.Do(effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Label: "Gandalf, White Rider — +1/+0"}, effects.Scry{N: 1})),
			effects.Optional(effects.WhenThisDies("Gandalf, White Rider — put it into its owner's library fifth from the top",
				func(g *game.Game, item *game.StackItem) error {
					return g.TuckToLibraryAtDepthForEffect(item.SourceCardID, 5)
				}), "Gandalf, White Rider — put Gandalf into its owner's library fifth from the top?"),
		},
	})
}
