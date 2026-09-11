package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
func init() {
	Register(Spec{
		OracleID: "f0435065-a8ca-4b4d-a7da-0ef41749118f",
		Name:     "The Binding of the Titans",
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, "The Binding of the Titans — I: each player mills three", titansMillAll),
			ChapterTriggerTargeting(2, "The Binding of the Titans — II: exile up to two cards from graveyards",
				TargetCardInGraveyard("up to two target cards in graveyards").WithCount(0, 2),
				titansExileFromGraveyards),
			ChapterTriggerTargeting(3, "The Binding of the Titans — III: return a creature or land to hand",
				TargetCardInGraveyard("target creature or land card in your graveyard",
					YouOwn(), Or(Creature(), Land())),
				titansReturnToHand),
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
	creatures := 0
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		card, ok := g.LookupCardForEffect(t.ID)
		if !ok {
			continue
		}
		wasCreature := card.IsCreature()
		if err := g.ExileCardForEffect(t.ID); err != nil {
			return err
		}
		if wasCreature {
			creatures++
		}
	}
	if creatures == 0 {
		return nil
	}
	return GainLife{Player: item.Controller, Amount: creatures}.Apply(ctx)
}

func titansReturnToHand(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	return ReturnFromGraveyard{Target: item.Targets[0].ID, Dest: game.ZoneHand}.Apply(NewContext(g, item))
}
