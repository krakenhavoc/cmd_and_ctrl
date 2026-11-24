package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Koll, the Forgemaster — Legendary Creature — Dwarf Warrior {R}{W}, 2/2:
//
//	"Whenever another nontoken creature you control dies, if it was enchanted
//	 or equipped, return it to its owner's hand.
//	 Creature tokens you control that are enchanted or equipped get +1/+1."
//
// The death trigger is Cass, Hand of Vengeance's: an intervening "if" (CR
// 603.4) read from the departure record of what was attached as the creature
// died (CR 603.10a), checked when it dies and again as the trigger resolves.
// The card returns only if it is still the object that died, in its graveyard.
// The static is a layer 7c anthem over the creature tokens its controller
// controls with an Aura or Equipment attached right now.
//
// No simplification.
func init() {
	wasAttached := func(g *game.Game, id uuid.UUID) bool {
		return len(effects.GapB64AttachmentsOnTheDead(g, id, true))+len(effects.GapB64AttachmentsOnTheDead(g, id, false)) > 0
	}
	cond := effects.CondBoard{
		Name: "it was enchanted or equipped",
		Test: func(ctx *effects.Context) bool {
			return ctx.Item != nil && ctx.Item.Trigger != nil && wasAttached(ctx.Game, ctx.Item.Trigger.Event.CardID)
		},
	}
	trigger := effects.On(game.EventLTB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return effects.HofriGhostforgeAnotherNontokenCreatureDied(ev, source, lki, g) && wasAttached(g, ev.CardID)
	}, "Koll, the Forgemaster — return it to its owner's hand", func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		dead := item.Trigger.Event.CardID
		if !effects.DiedCardStillInGraveyard(ctx, dead) {
			return nil
		}
		return effects.ReturnFromGraveyard{Target: dead, Dest: game.ZoneHand}.Apply(ctx)
	})
	effects.Register(effects.Spec{
		OracleID:     "344dc8b0-302a-4721-99b6-5a6626a349c1",
		Name:         "Koll, the Forgemaster",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16Anthem(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.IsToken() && target.Controller == source.Controller &&
					effects.RprB03IsEnchantedOrEquipped(g, target.InstanceID)
			}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{effects.InterveningIf(cond, trigger)},
	})
}
