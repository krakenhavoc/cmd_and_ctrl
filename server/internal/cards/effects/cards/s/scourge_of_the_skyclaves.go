package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scourge of the Skyclaves — Creature — Demon {1}{B}, */* (Zendikar Rising,
// S82 OLD #5508):
//
//	"Kicker {4}{B}
//	 When you cast this spell, if it was kicked, each player loses half
//	 their life, rounded up.
//	 Scourge of the Skyclaves's power and toughness are each equal to 20
//	 minus the highest life total among players."
//
// The cast trigger is an intervening "if" (CR 603.4) on the kicker of the
// spell on the stack; it resolves above the spell, so it resolves even if
// the spell is countered. The size is a layer 7a characteristic-defining
// ability (CR 604.3, 613.4a) and can be negative, which the engine reads as
// 0 for combat and as lethal for state-based actions (CR 704.5f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "c6dfc91f-bfac-40c2-99e5-0cec981f7f1f",
		Name:          "Scourge of the Skyclaves",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{4}{B}")},
		Triggered:     []game.TriggeredAbility{scourgeOfTheSkyclavesTrigger()},
		Static: []game.StaticAbility{{
			Label:    "Power and toughness are each equal to 20 minus the highest life total among players.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, _ *game.Card) {
				n := 20 - effects.S82OldB090HighestLifeTotal(g)
				c.Power, c.Toughness = n, n
			},
		}},
	})
}

func scourgeOfTheSkyclavesTrigger() game.TriggeredAbility {
	t := effects.WhenYouCastThisSpell("Scourge of the Skyclaves — kicked: each player loses half their life, rounded up",
		effects.S82OldB090EachPlayerLosesHalfTheirLife)
	when := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		if !when(ev, source, lki, g) {
			return false
		}
		spell, ok := g.LookupCardForEffect(ev.CardID)
		return ok && game.KickedTimesPaid(spell, g.StackItemPaidForEffect(ev.CardID).OptionalCosts) > 0
	}
	return t
}
