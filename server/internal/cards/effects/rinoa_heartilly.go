package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rinoa Heartilly — Legendary Creature — Human Rebel Warlock
// {3}{G}{W}, 4/4 (Final Fantasy, #4462):
//
//	"When Rinoa Heartilly enters, create Angelo, a legendary 1/1 green
//	 and white Dog creature token.
//	 Angelo Cannon — Whenever Rinoa Heartilly attacks, another target
//	 creature you control gets +1/+1 until end of turn for each creature
//	 you control."
//
// The creatures are counted as the attack trigger resolves (CR 608.2h),
// Angelo and the target included. If the target is gone then, nothing
// happens (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8896c74b-d466-4c25-94cf-82b5e2b1fdef",
		Name:         "Rinoa Heartilly",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Rinoa Heartilly — create Angelo, a legendary 1/1 green and white Dog",
				Do(CreateToken{Template: TokenCard("1/1 green and white Legendary Dog named Angelo"), N: 1})),
			Targeting(
				On(game.EventAttack, ThisAttacked,
					"Rinoa Heartilly — Angelo Cannon: another target creature you control gets +1/+1 until end of turn for each creature you control",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						target, ok := b16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						n := countControlled(g, item.Controller, func(c game.Card) bool { return c.IsCreature() })
						return BoostUntilEOT{Target: target, Power: n, Toughness: n, Label: "Angelo Cannon — +1/+1 for each creature you control"}.Apply(ctx)
					}),
				Another(TargetCreature("another target creature you control", YouControl()))),
		},
	})
}
