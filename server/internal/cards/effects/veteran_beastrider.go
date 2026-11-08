package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "29f1484c-91b6-436a-a6be-3ab8b58c0fc1",
		Name:         "Veteran Beastrider",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Veteran Beastrider — untap each creature you control",
				Do(UntapAllCreaturesYouControl{})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{G}{W}: Creatures you control get +1/+1 until end of turn.",
			Cost:    ManaCost("{2}{G}{W}"),
			Purpose: game.Purpose{TeamPump: &game.Pump{Power: 1, Toughness: 1}, Answers: game.AnswerPump},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return BoostUntilEOT{
					Match: And(Creature(), YouControl()), Power: 1, Toughness: 1,
					Label: "Veteran Beastrider — creatures you control get +1/+1",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
