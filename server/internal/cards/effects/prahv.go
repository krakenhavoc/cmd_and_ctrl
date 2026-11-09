package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prahv — Plane — Ravnica (Planechase Anthology, #4870):
//
//	"If you cast a spell this turn, you can't attack with creatures.
//	 If you attacked with creatures this turn, you can't cast spells.
//	 Whenever chaos ensues, you gain life equal to the number of cards in
//	 your hand."
//
// A plane (ADR 0216 §6), where "you" is the planar controller (CR 901.6).
// The first static is Angelic Arbiter's CantAttack over the planar
// controller's creatures once their cast tally this turn is not zero
// (restrictions are checked at declaration, CR 508.1c); the second is its
// cast restriction, read off this turn's declared attackers only, so a
// creature put onto the battlefield attacking is no attack (the ruling,
// CR 508.4). The chaos ability counts the hand as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "34a31865-b701-4ce1-8912-29fc41ecdda3",
		Name:         "Prahv",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:               "If you cast a spell this turn, you can't attack with creatures.",
			Layer:               game.Layer6Ability,
			DependsOnSpellsCast: true,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller &&
					g.CastTallyFor(target.Controller).Total > 0
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantAttack
			},
		}},
		CastRestrictions: []game.CastRestriction{{
			Label: "If you attacked with creatures this turn, you can't cast spells.",
			Forbids: func(q game.CastQuery) bool {
				return q.Source.Controller == q.Controller && q.Game.DistinctAttackersThisTurn(q.Controller) > 0
			},
		}},
		Triggered: []game.TriggeredAbility{
			WheneverChaosEnsues("Prahv — you gain life equal to the number of cards in your hand",
				func(g *game.Game, item *game.StackItem) error {
					p := g.PlayerByIDForEffect(item.Controller)
					if p == nil || p.Hand == nil {
						return nil
					}
					return GainLife{Player: item.Controller, Amount: p.Hand.Size()}.Apply(NewContext(g, item))
				}),
		},
	})
}
