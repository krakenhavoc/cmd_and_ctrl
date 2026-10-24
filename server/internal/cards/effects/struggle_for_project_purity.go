package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Struggle for Project Purity — Enchantment {3}{U}:
//
//	"As this enchantment enters, choose Brotherhood or Enclave.
//	 • Brotherhood — At the beginning of your upkeep, each opponent draws
//	   a card. You draw a card for each card drawn this way.
//	 • Enclave — Whenever a player attacks you with one or more
//	   creatures, that player gets twice that many rad counters."
//
// #2809. An anchor-word card (ADR 0071), Outpost Siege's shape: each
// bullet is a trigger gated on its word.
//
//   - Brotherhood: each opponent draws a card in turn order, then you
//     draw one for each card they really drew — a draw from an empty
//     library or one a replacement turned into something else put no
//     card in a hand and counts nothing. The count is the opponents'
//     draws in this resolution's event batch, read once every draw has
//     settled (the second step waits for a prompt the first asks, CR
//     608.2c).
//   - Enclave: one trigger per declaration of attackers (CR 603.2c,
//     508.3e), counting only the creatures attacking YOU, not your
//     planeswalkers or battles, and none put onto the battlefield
//     attacking (they are not declared). "That many" is that count
//     (ThatManyEvents), and the attacking player gets twice as many rad
//     counters from you.
//
// No simplification.
func init() {
	const (
		brotherhood = "Struggle for Project Purity — Brotherhood: each opponent draws a card; you draw a card for each card drawn this way"
		enclave     = "Struggle for Project Purity — Enclave: that player gets twice that many rad counters"
	)
	Register(Spec{
		OracleID:     "5958e9e3-9457-48e1-afc1-a5c89e3b0ed0",
		Name:         "Struggle for Project Purity",
		Completeness: CompletenessFull,
		AsEnters:     ChooseOptionAsEnters("Struggle for Project Purity", "Brotherhood", "Enclave"),
		Triggered: []game.TriggeredAbility{
			WhenChosen("Brotherhood", AtYourUpkeep(brotherhood, Do(
				Instruction(func(ctx *Context) error {
					for _, opp := range opponentsInTurnOrder(ctx.Game, ctx.Controller()) {
						if err := (DrawCards{Player: opp, N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
				Instruction(func(ctx *Context) error {
					return DrawCards{N: opponentsDrawsThisResolution(ctx)}.Apply(ctx)
				}),
			))),
			WhenChosen("Enclave", withAtBatchEnd(On(game.EventAttack, aCreatureAttackedYouDirectly, enclave,
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return playerGetsRadCounters(g, item.Controller, item.Trigger.Event.Actor, 2*settledEventAmount(item))
				}), ThatManyEvents())),
		},
	})
}

// aCreatureAttackedYouDirectly is "a player attacks you with a
// creature" (CR 508.3e): a creature declared attacking the source's
// controller themself, by another player. An attack at their
// planeswalker or battle is not one.
func aCreatureAttackedYouDirectly(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventAttack && ev.Actor != uuid.Nil && ev.Actor != source.Controller && ev.Target == source.Controller
}

// opponentsDrawsThisResolution counts the cards the controller's
// opponents drew in the resolving item's event batch: "for each card
// drawn this way" (Struggle for Project Purity's Brotherhood).
func opponentsDrawsThisResolution(ctx *Context) int {
	g, you := ctx.Game, ctx.Controller()
	batch := g.CurrentEventBatchForEffect()
	n := 0
	events := g.EventsThisTurn()
	for i := len(events) - 1; i >= 0; i-- {
		ev := events[i]
		if ev.Batch < batch {
			break
		}
		if ev.Batch == batch && ev.Kind == game.EventDrawCard && ev.Actor != uuid.Nil && ev.Actor != you {
			n++
		}
	}
	return n
}
