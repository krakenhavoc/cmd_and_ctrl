package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lena, Selfless Champion — Legendary Creature — Human Knight {4}{W}{W},
// 3/3 (Dominaria; Jumpstart):
//
//	"When Lena enters, create a 1/1 white Soldier creature token for each
//	 nontoken creature you control.
//	 Sacrifice Lena: Creatures you control with power less than Lena's
//	 power gain indestructible until end of turn."
//
// The entry counts the nontoken creatures you control as the trigger
// resolves, Lena herself included while she is still there. The
// sacrifice is paid as a cost, so "Lena's power" is her last-known power
// on the battlefield (CR 608.2h), with any pump and counters she had;
// the set of creatures is fixed as the ability resolves (CR 611.2c), so
// a creature that later shrinks or enters gains nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a8a8407c-d0e3-4311-82ed-1ab1f75159d9",
		Name:         "Lena, Selfless Champion",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Lena, Selfless Champion — a 1/1 white Soldier for each nontoken creature you control",
				func(g *game.Game, item *game.StackItem) error {
					n := jmpB5NontokenCreaturesYouControl(g, item.Controller)
					if n == 0 {
						return nil
					}
					return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 white Soldier"), N: n}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice Lena: Creatures you control with power less than Lena's power gain indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive, Saves: &game.Saves{Matches: game.SweepCreatures, Partial: true}},
			Cost:    SacrificeThis(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				info, ok := ctx.SourcePermanent()
				if !ok {
					return nil
				}
				lena := info.Power
				return GrantKeywordUntilEOT{
					Match: And(Creature(), YouControl(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
						return c.PowerForComparison() < lena
					}),
					Keywords: []string{"indestructible"},
					Label:    "Lena, Selfless Champion — indestructible",
				}.Apply(ctx)
			},
		}},
	})
}
