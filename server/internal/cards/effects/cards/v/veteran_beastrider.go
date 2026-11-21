package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Veteran Beastrider — Creature — Human Knight {1}{G}{W}, 3/4
// (Aetherdrift, #4464):
//
//	"At the beginning of your end step, untap each creature you control.
//	 {2}{G}{W}: Creatures you control get +1/+1 until end of turn."
//
// The end-step trigger is UntapAllCreaturesYouControl, over the
// creatures its controller controls as it resolves. The pump is Bearer
// of Glory's body: the set is fixed when it resolves (CR 611.2c), so a
// creature that enters later this turn is not pumped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "29f1484c-91b6-436a-a6be-3ab8b58c0fc1",
		Name:         "Veteran Beastrider",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Veteran Beastrider — untap each creature you control",
				effects.Do(effects.UntapAllCreaturesYouControl{})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}{W}: Creatures you control get +1/+1 until end of turn.",
			Cost:    effects.ManaCost("{2}{G}{W}"),
			Purpose: game.Purpose{TeamPump: &game.Pump{Power: 1, Toughness: 1}, Answers: game.AnswerPump},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Toughness: 1,
					Label: "Veteran Beastrider — creatures you control get +1/+1",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
