package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Binding of the Titans — Enchantment — Saga for {1}{G}:
//
//	"I — Each player mills three cards.
//	 II — Exile up to two target cards from graveyards. For each
//	      creature card exiled this way, you gain 1 life.
//	 III — Return target creature or land card from your graveyard
//	       to your hand."
//
// Chapter II is the "up to two" shape: one target clause with
// WithCount(0, 2), and the effect iterates ctx.LegalTargets() so a
// card that left the graveyard between announce and resolution is
// skipped rather than erroring (CR 608.2b). The life gain is counted
// off what was ACTUALLY exiled, which is why the card type is read
// before the move and the total is applied after — reading it after
// would find the card in exile with its characteristics intact, but
// counting a card that failed to move would gain life for nothing.
//
// #870 made "actually exiled" mean it. The count used to come from
// the exile calls that returned no error, and a graveyard is a
// CR 903.9 zone: exiling an opponent's commander card out of it asks
// their owner about the command zone, the call returns nil with
// nothing moved, and the chapter gained a life for a card still lying
// in the graveyard. The clause runs from the batch's continuation
// instead, over the cards that ARRIVED in exile (CR 400.7) — so a
// commander that takes the offer is not one of them.
func init() {
	Register(Spec{
		OracleID:     "f0435065-a8ca-4b4d-a7da-0ef41749118f",
		Name:         "The Binding of the Titans",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, "The Binding of the Titans — I: each player mills three", titansMillAll),
			ChapterTriggerTargeting(2, "The Binding of the Titans — II: exile up to two cards from graveyards",
				TargetCardInGraveyard("up to two target cards in graveyards").WithCount(0, 2),
				titansExileFromGraveyards),
			ChapterTriggerTargeting(3, "The Binding of the Titans — III: return a creature or land to hand",
				TargetCardInGraveyard("target creature or land card in your graveyard",
					YouOwn(), Or(Creature(), Land())),
				returnTargetCardToHand),
		},
	})
}

func titansMillAll(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, p := range g.Seats {
		if p.Eliminated {
			continue
		}
		if err := (MillCards{Player: p.ID, N: 3}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

func titansExileFromGraveyards(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	wasCreature := map[uuid.UUID]bool{}
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		card, ok := g.LookupCardForEffect(t.ID)
		if !ok {
			continue
		}
		ids = append(ids, t.ID)
		wasCreature[t.ID] = card.IsCreature()
	}
	return g.ExileCardsThenForEffect(ids, func(g *game.Game, exiled []uuid.UUID) error {
		creatures := 0
		for _, id := range exiled {
			if wasCreature[id] {
				creatures++
			}
		}
		if creatures == 0 {
			return nil
		}
		return GainLife{Player: item.Controller, Amount: creatures}.Apply(NewContext(g, item))
	})
}
