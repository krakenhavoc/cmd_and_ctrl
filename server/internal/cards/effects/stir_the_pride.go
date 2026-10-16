package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const stirThePrideLifeGrant = "stir-the-pride/deals-damage-gain-that-much"

// Stir the Pride — Instant {4}{W}:
//
//	"Choose one —
//	 • Creatures you control get +2/+2 until end of turn.
//	 • Until end of turn, creatures you control gain "Whenever this
//	   creature deals damage, you gain that much life."
//	 Entwine {1}{W} (Choose both if you pay the entwine cost.)"
//
// Entwine is CR 702.42 (#3866). The pump is Overrun's BoostUntilEOT
// over the creatures you control as it resolves (CR 611.2c). The
// quoted trigger is a catalog bundle granted until end of turn (ADR
// 0093 PR 4, Requiem Monolith's shape), so it is each creature's own
// ability and "you" is that creature's controller; it reads each damage
// event the creature deals, with the amount dealt, as Spirit Link and
// Phantom Nishoba do.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7e46d2c7-59ae-4b0e-a2ec-d0ac1b68d594",
		Name:         "Stir the Pride",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: stirThePrideLifeGrant,
			Triggered: []game.TriggeredAbility{{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Amount > 0 && ev.Source == source.InstanceID
				},
				Key: "Stir the Pride — you gain that much life",
				Effect: func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return GainLife{Player: item.Controller, Amount: item.Trigger.Event.Amount}.Apply(NewContext(g, item))
				},
			}},
			Text: "Whenever this creature deals damage, you gain that much life.",
		}},
		OptionalCosts: []game.AdditionalCost{Entwine("{1}{W}")},
		Modes: Entwining(ChooseOne(
			ModeDoing("Creatures you control get +2/+2 until end of turn.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return BoostUntilEOT{
						Match: And(Creature(), YouControl()), Power: 2, Toughness: 2,
						Label: "Stir the Pride — +2/+2",
					}.Apply(ctx)
				}),
			ModeDoing("Until end of turn, creatures you control gain \"Whenever this creature deals damage, you gain that much life.\"", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return GrantAbilitiesFor{
						Match: And(Creature(), YouControl()),
						Keys:  []string{stirThePrideLifeGrant},
						Label: "Stir the Pride — whenever this creature deals damage, you gain that much life",
					}.Apply(ctx)
				}),
		)),
	})
}
