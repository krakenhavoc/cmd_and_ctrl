package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gyruda, Doom of Depths — Legendary Creature — Demon Kraken {4}{U/B}{U/B},
// 6/6:
//
//	"Companion — Your starting deck contains only cards with even mana
//	 values. (If this card is your chosen companion, you may put it into
//	 your hand from outside the game for {3} as a sorcery.)
//	 When Gyruda enters, each player mills four cards. Put a creature
//	 card with an even mana value from among the milled cards onto the
//	 battlefield under your control."
//
// Every live seat mills four in APNAP order (CR 101.4), each mill settling
// before the next, and the keyed tail is told every card that reached a
// graveyard. Of those still in a graveyard the creature cards with an even
// mana value (a card with no mana cost is 0, which is even, CR 202.3a; {X}
// is 0 outside the stack, CR 202.3e) are offered in one prompt and the
// controller must pick one, from any owner's graveyard; it enters under
// their control. It is not a target. With none eligible nothing enters.
//
// Companion (ADR 0172) is not this file's business: the condition is the
// deck check's (internal/deck, the companion registry), and the {3} special
// action is the engine's. In the 99, or as a commander, the companion
// ability does nothing, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "25eaa977-ca85-4f8d-8be5-9297a3edb14f",
		Name:         "Gyruda, Doom of Depths",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Gyruda, Doom of Depths — each player mills four cards; put a creature card with an even mana value from among the milled cards onto the battlefield under your control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.MillEachPlayerInTurn(ctx, effects.ApnapPlayers(g), effects.ConstantMill(4), ctx.ItemTail(gyrudaMilled))
				}),
		},
	})
}

// gyrudaMilled is Gyruda's tail, told every card the mills put into a
// graveyard: it asks the controller for one creature card with an even
// mana value among them. An on-disk key.
var gyrudaMilled = effects.CardTail("gyruda/even-creature-among-milled", func(ctx *effects.Context, o game.TailOutcome) error {
	var candidates []uuid.UUID
	for _, id := range o.IDs {
		z := ctx.Game.FindCardZoneForEffect(id)
		if z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsCreature() && c.ManaValue()%2 == 0 {
			candidates = append(candidates, id)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Gyruda, Doom of Depths — put a creature card with an even mana value from among the milled cards onto the battlefield under your control",
		Cards:    candidates,
		Min:      1,
		Max:      1,
		ThenKey:  gyrudaPicked,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
})

// gyrudaPicked is the chosen creature card entering under Gyruda's
// controller's control. An on-disk key.
var gyrudaPicked = effects.ItemChoice("choose-cards/gyruda-reanimate", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return effects.ReturnFromGraveyard{Target: r.Picked[0], Dest: game.ZoneBattlefield, Controller: ctx.Controller()}.Apply(ctx)
})
