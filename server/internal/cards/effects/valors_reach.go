package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Valor's Reach — Plane — Kylem (#4870):
//
//	"Whenever your team attacks with exactly two creatures, those creatures
//	 gain double strike until end of turn.
//	 Whenever chaos ensues, untap up to two target creatures your team
//	 controls. If it's a main phase, there is an additional combat phase after
//	 this phase, followed by an additional main phase."
//
// "Your team" is "you" outside a team game (CR 102.4), and the engine has no
// team games, so it is the planar controller. The attack ability triggers once
// per declaration, however many creatures attack: the trigger belongs to the
// first declared attacker (game.DeclaredAttackersOfPlayerForEffect), and only
// when the declaration is exactly two creatures. The chaos ability's phases
// are added only when it resolves in a main phase (ExtraCombatAndMainAfterThisMain).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "61b5984a-de6a-4d5a-8b20-e2fb3b553a34",
		Name:         "Valor's Reach",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, b6YouAttackedWithExactlyTwo,
				"Valor's Reach — those creatures gain double strike until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, id := range g.DeclaredAttackersOfPlayerForEffect(item.Controller) {
						if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"double strike"}, Label: "Valor's Reach — double strike"}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
			Targeting(WheneverChaosEnsues("Valor's Reach — untap up to two target creatures your team controls; if it's a main phase, an additional combat phase and main phase follow",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := untapEach(ctx, legalTargetCards(item, g)); err != nil {
						return err
					}
					return ExtraCombatAndMainAfterThisMain().Apply(ctx)
				}),
				TargetCreature("up to two target creatures your team controls", YouControl()).WithCount(0, 2)),
		},
	})
}

// b6YouAttackedWithExactlyTwo is "your team attacks with exactly two
// creatures", true once per declaration: on the first of its attackers.
func b6YouAttackedWithExactlyTwo(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !attackDeclaredByYou(ev, source.Controller) {
		return false
	}
	attackers := g.DeclaredAttackersOfPlayerForEffect(source.Controller)
	return len(attackers) == 2 && attackers[0] == ev.CardID
}
