package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flayer of the Hatebound — Creature — Devil {5}{R}, 4/2:
//
//	"Undying
//	 Whenever this creature or another creature enters from your
//	 graveyard, that creature deals damage equal to its power to any
//	 target."
//
// "Your graveyard" is EventETB.EnteredFromOwner (ADR 0113 amendment
// 2026-10-08), so a creature an opponent returned from THEIR graveyard
// does not trigger it. The entering creature is the damage source. The
// amount is its power when the trigger resolves (a pump in response
// counts), falling back to the power at entry if it has left the
// battlefield (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ba9f3c5d-556f-48d1-8d87-723f7893cfcd",
		Name:            "Flayer of the Hatebound",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordUndying},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				if !EnteredFromYourGraveyard(ev, source, lki, g) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature()
			},
			Targets: TargetAny(),
			Key:     "Flayer of the Hatebound — that creature deals damage equal to its power to any target",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Flayer of the Hatebound — that creature deals damage equal to its power to any target")
				if c, ok := g.LookupCardForEffect(ev.CardID); ok {
					item.Params.Amount = c.CurrentPower()
				}
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 {
					return nil
				}
				entered := item.Trigger.Event.CardID
				amount := item.Params.Amount
				if z := g.FindCardZoneForEffect(entered); z != nil && z.Kind == game.ZoneBattlefield {
					if c, ok := g.LookupCardForEffect(entered); ok {
						amount = c.CurrentPower()
					}
				}
				return DealDamage{
					Source: entered,
					Target: item.Targets[0].ID,
					Amount: amount,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
