package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Passionate Archaeologist — Legendary Enchantment — Background {1}{R}
// (Commander Legends: Battle for Baldur's Gate, #5948):
//
//	"Commander creatures you own have "Whenever you cast a spell from exile,
//	 this creature deals damage equal to that spell's mana value to target
//	 opponent.""
//
// A granted trigger (ADR 0093), the commander creature's own, reading a cast
// from exile off the cast event (Kami of Celebration's read of EventCast,
// which stamps the origin zone). The target opponent is chosen as the trigger
// goes on the stack; the damage is the spell's mana value read as the spell
// was cast (a spell with {X} counts X as 0 on the stack, CR 202.3e), dealt by
// the commander creature. Choose a Background is deck construction (ADR 0144).
//
// No simplification.
const passionateArchaeologistGrant = "passionate-archaeologist/damage"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "6299d18e-a6a7-414c-85c0-d57a8e7d2627",
		Name:         "Passionate Archaeologist",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: passionateArchaeologistGrant,
			Triggered: []game.TriggeredAbility{
				effects.Targeting(effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Actor == source.Controller && ev.OldZone == game.ZoneExile
				}, "Passionate Archaeologist — this creature deals damage equal to that spell's mana value to target opponent",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						spell, ok := g.LookupCardForEffect(item.Trigger.Event.CardID)
						if !ok {
							return nil
						}
						mv := spell.ManaValue()
						if mv <= 0 {
							return nil
						}
						for _, t := range ctx.LegalTargets() {
							if t.Kind == game.TargetPlayer {
								return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: mv}.Apply(ctx)
							}
						}
						return nil
					}),
					effects.TargetPlayer("target opponent", effects.Opponent())),
			},
			Text: "Whenever you cast a spell from exile, this creature deals damage equal to that spell's mana value to target opponent.",
		}},
		Static: []game.StaticAbility{effects.GrantToCommanderCreaturesYouOwn(passionateArchaeologistGrant)},
	})
}
