package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// monumentToEnduranceLabel is the trigger's stack label, and with it
// the key its "hasn't been chosen this turn" memory is kept under.
const monumentToEnduranceLabel = "Monument to Endurance — you discarded a card"

// Monument to Endurance — Artifact {3}:
//
//	"Whenever you discard a card, choose one that hasn't been chosen
//	 this turn —
//	 • Draw a card.
//	 • Create a Treasure token.
//	 • Each opponent loses 3 life."
//
// A discard payoff (#369 reported it missing). The trigger is Bag of
// Holding's condition — EventDiscardCard by the controller, which the
// one discard path emits for every cause (an effect, a cost, the
// cleanup step) — so it fires once per card discarded, and each
// instance chooses its own bullet.
//
// "That hasn't been chosen this turn" is ChooseOneNotChosenThisTurn
// (ADR 0097, #1749): the engine remembers the bullets this Monument's
// ability has chosen this turn, recorded as each is chosen. Discarding
// three cards at once queues three prompts that take three different
// bullets; a fourth discard that turn triggers and is removed with no
// effect (CR 700.2b). The memory resets as the next turn begins.
//
// "Each opponent loses 3 life" is life loss, not damage
// (eachOpponentLosesLife), so nothing that prevents damage sees it.
//
// The trigger declares a discard payoff (ADR 0126's amendment of
// 2026-10-06) for the curated Izzet deck (#2480): any card, one token.
// The Treasure is the bullet the purpose can say, and the least the
// bot is paid: the draw and the 3 life are worth more, and a fourth
// discard in a turn pays nothing, which the declaration does not see.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e69e8de4-b521-4888-8074-17f1efe2f345",
		Name:         "Monument to Endurance",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(monumentToEnduranceTrigger(),
				game.Purpose{DiscardPayoff: &game.DiscardPayoff{Any: true, Tokens: 1}}),
		},
	})
}

func monumentToEnduranceTrigger() game.TriggeredAbility {
	t := effects.On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return effects.DiscardedByYou(ev, source)
	}, monumentToEnduranceLabel, func(_ *game.Game, _ *game.StackItem) error { return nil })
	t.Modes = effects.ChooseOneNotChosenThisTurn(
		effects.ModeDoing("Draw a card.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
			}),
		effects.ModeDoing("Create a Treasure token.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.TreasureToken(), N: 1}.Apply(ctx)
			}),
		effects.ModeDoing("Each opponent loses 3 life.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.EachOpponentLosesLifeNow(ctx.Game, item, 3)
			}),
	)
	return t
}
