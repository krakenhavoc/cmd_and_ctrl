package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Quest for the Nihil Stone — Enchantment {B} (Worldwake, S82 OLD b487):
//
//	"Whenever an opponent discards a card, you may put a quest counter on this
//	 enchantment.
//	 At the beginning of each opponent's upkeep, if that player has no cards in
//	 hand and this enchantment has two or more quest counters on it, you may have
//	 that player lose 5 life."
//
// The first is Geth's Grimoire's watch on an opponent's discard (one trigger per
// card discarded), the counter an optional CR 603.5 "you may". The second is an
// intervening "if" (CR 603.4): it triggers only when the upkeep's player, an
// opponent, has an empty hand and the enchantment has two or more quest counters,
// and is checked again before the "you may" is asked, the counters read as they
// were if the enchantment has left (last-known information). The life loss is a
// loss, not damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7f65178e-dd2e-44aa-a68f-f8694e0b250a",
		Name:         "Quest for the Nihil Stone",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.B18OpponentDiscarded(ev, source)
			}, "Quest for the Nihil Stone — put a quest counter", effects.Do(effects.CounterOnThis{Kind: "quest", N: 1})),
				"Quest for the Nihil Stone — put a quest counter on it?"),
			effects.Optional(questNihilStoneUpkeep(), "Quest for the Nihil Stone — have that player lose 5 life?"),
		},
	})
}

// questNihilStoneUpkeep is the second ability: an opponent's upkeep, that player's
// hand empty, two or more quest counters, checked as it triggers and as it resolves.
func questNihilStoneUpkeep() game.TriggeredAbility {
	holds := func(g *game.Game, source uuid.UUID, player uuid.UUID) bool {
		return player != uuid.Nil && effects.CardsInHand(g, player) == 0 && effects.S82OldB487QuestCounters(g, source) >= 2
	}
	t := effects.On(game.EventBeginUpkeep, effects.ByAnOpponent, "Quest for the Nihil Stone — that player loses 5 life",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			player := ctx.Trigger().Event.Actor
			if player == uuid.Nil {
				return nil
			}
			return g.ChangePlayerLifeForEffect(item.SourceCardID, player, -5)
		})
	when := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return source != nil && holds(g, source.InstanceID, ev.Actor) && (when == nil || when(ev, source, lki, g))
	}
	return effects.InterveningIf(effects.CondBoard{Name: "that player has no cards in hand and this has two or more quest counters", Test: func(ctx *effects.Context) bool {
		return holds(ctx.Game, ctx.Source(), ctx.Trigger().Event.Actor)
	}}, t)
}
