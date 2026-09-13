package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kambal, Consul of Allocation — Legendary Creature — Human Advisor
// {1}{W}{B}, 2/3:
//
//	"Whenever an opponent casts a noncreature spell, that player
//	 loses 2 life and you gain 2 life."
//
// The issue files this under cost modification, and it is worth being
// precise about why it is NOT one: Kambal does not change what
// anything costs. The spell is announced and paid for exactly as
// printed; Kambal's ability then TRIGGERS on the cast (CR 603) and
// drains afterwards. Thalia is the cost modifier; Kambal is the tax
// you pay in life after the fact, which is a different engine.
//
// Two consequences of that, both visible at the table: the drain
// happens even if the spell is countered (the trigger is on the CAST,
// not the resolution), and it goes on the stack ABOVE the spell, so
// the life totals move first.
//
// "That player loses 2 life and YOU gain 2 life" is two separate life
// changes rather than a drain, which matters for anything watching
// life gain and for a Platinum Angel-shaped effect on either side.
func init() {
	Register(Spec{
		OracleID: "4987c458-604a-4727-b360-170616e91e67",
		Name:     "Kambal, Consul of Allocation",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor == source.Controller {
					return false // "an OPPONENT casts"
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && !spell.IsCreature()
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				caster := ev.Actor
				return game.NewTriggeredItem(source, "Kambal — that player loses 2 life, you gain 2 life",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						// Losing life is a negative GainLife, not
						// damage: Kambal's text says "loses 2 life",
						// so nothing that prevents or redirects
						// damage touches it.
						if err := (GainLife{Player: caster, Amount: -2}).Apply(ctx); err != nil {
							return err
						}
						return GainLife{Player: item.Controller, Amount: 2}.Apply(ctx)
					})
			},
		}},
	})
}
