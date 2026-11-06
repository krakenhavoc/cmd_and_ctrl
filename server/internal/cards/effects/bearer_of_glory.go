package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bearer of Glory — Creature — Human Soldier {1}{W}, 2/1 (Tarkir:
// Dragonstorm, #4463):
//
//	"During your turn, this creature has first strike.
//	 {4}{W}: Creatures you control get +1/+1 until end of turn."
//
// The first strike is Leonin Surveyor's layer 6 grant on its controller's
// own turn, read at every layer pass. The pump is Battershield Warrior's
// body without boast: the creatures you control as it resolves (CR
// 611.2c), so one that enters later this turn is not pumped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "963947d5-5fe4-49cd-be0c-0bdb5f7bcbb6",
		Name:         "Bearer of Glory",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return thisPermanent(target, g, source) && isActivePlayer(g, source.Controller)
			}, "first strike"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}{W}: Creatures you control get +1/+1 until end of turn.",
			Cost:    ManaCost("{4}{W}"),
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return BoostUntilEOT{
					Match: And(Creature(), YouControl()), Power: 1, Toughness: 1,
					Label: "Bearer of Glory — creatures you control get +1/+1",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
