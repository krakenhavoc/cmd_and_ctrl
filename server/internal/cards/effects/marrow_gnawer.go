package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Marrow-Gnawer — Legendary Creature — Rat Rogue {3}{B}{B}, 2/3:
//
//	"All Rats have fear.
//	 {T}, Sacrifice a Rat: Create X 1/1 black Rat creature tokens,
//	 where X is the number of Rats you control."
//
// "All Rats" is every Rat on the battlefield, whoever controls it
// (no YoursOnly). The Rat sacrificed is a cost, so it is gone before
// the count is taken at resolution (CR 608.2h): X is the Rats left
// under your control, and Marrow-Gnawer may sacrifice itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2b934761-5720-477b-94b7-d1d91dc581a6",
		Name:         "Marrow-Gnawer",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Rat"}}, "fear"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice a Rat: Create X 1/1 black Rat creature tokens, where X is the number of Rats you control.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerMakesBlocker, TokensPerControlled: "rat"},
			Cost: Plus(TapCost(), game.AbilityCost{
				SacrificeOther: sacrificeSpec("a Rat", Subtype("Rat")),
			}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				rat := ctx.TextSubtype("Rat")
				n := 0
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == item.Controller && c.IsCreature() && c.HasSubtype(rat) {
						n++
					}
				}
				if n == 0 {
					return nil
				}
				return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 black Rat"), N: n}.Apply(ctx)
			},
		}},
	})
}
