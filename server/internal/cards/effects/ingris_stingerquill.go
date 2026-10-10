package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ingris Stingerquill — Legendary Creature — Elder Sphinx {B}{R}{R},
// 1/4:
//
//	"Flying
//	 Whenever a creature you control attacks, that creature deals 1
//	 damage to each opponent.
//	 {4}: Create a 2/2 colorless Wizard Soldier creature token named
//	 Cadet. Then creatures you control gain haste until end of turn."
//
// The attack trigger fires once per attacking creature (EventAttack is
// per attacker) and the attacker, read off the triggering event, is the
// damage source. The activated ability makes the Cadet first, so it is
// among the creatures that gain haste; the set that gains haste is read
// as the ability resolves (CR 611.2c).
//
// No simplification.
func init() {
	const attackLabel = "Ingris Stingerquill — the attacking creature deals 1 damage to each opponent"
	Register(Spec{
		OracleID:        "bda098cb-31ec-41a1-a9d7-122877cc69d2",
		Name:            "Ingris Stingerquill",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, attackLabel, func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil {
					return nil
				}
				attacker := item.Trigger.Event.CardID
				ctx := NewContext(g, item)
				return g.DamageInstanceForEffect(func() error {
					for _, opp := range ctx.Opponents() {
						if err := g.DealDamageToPlayerForEffect(attacker, opp, 1); err != nil {
							return err
						}
					}
					return nil
				})
			}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}: Create a 2/2 colorless Wizard Soldier creature token named Cadet. Then creatures you control gain haste until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker | game.AnswerCombatGrant},
			Cost:    ManaCost("{4}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (CreateToken{Template: TokenCard("2/2 colorless Wizard Soldier named Cadet"), N: 1}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{
					Match:    And(Creature(), YouControl()),
					Keywords: []string{"haste"},
					Label:    "Ingris Stingerquill — haste until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
