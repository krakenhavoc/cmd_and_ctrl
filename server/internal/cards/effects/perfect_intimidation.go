package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Perfect Intimidation — Sorcery {3}{B} (Lorwyn Eclipsed, #4458):
//
//	"Choose one or both —
//	 • Target opponent exiles two cards from their hand.
//	 • Remove all counters from target creature."
//
// The bullets resolve in printed order (CR 608.2c), each with its own
// target. The first is an edict the opponent answers (Ruthless
// Negotiation's prompt): they choose which two cards, a hand of two or
// fewer goes whole, and an empty hand is not asked. The second removes
// every counter of every kind from the creature, as separate removals
// (Nexus Mentality's removal).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "da099c34-74b2-4db8-8583-09bf2c56606f",
		Name:         "Perfect Intimidation",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose one or both", 1, 2,
			ModeDoing("Target opponent exiles two cards from their hand.",
				TargetPlayer("target opponent", Opponent()), perfectIntimidationExile),
			ModeDoing("Remove all counters from target creature.",
				TargetCreature("target creature"), perfectIntimidationRemoveCounters),
		),
	})
}

// perfectIntimidationExile queues the opponent's choice of two cards.
func perfectIntimidationExile(_ *game.StackItem, ctx *Context, occ int) error {
	for _, t := range ctx.ModeTargets(occ) {
		if t.Kind != game.TargetPlayer || !ctx.IsTargetLegal(t) {
			continue
		}
		p := ctx.Game.PlayerByIDForEffect(t.ID)
		if p == nil || p.Hand == nil || p.Hand.Size() == 0 {
			return nil
		}
		cards := make([]uuid.UUID, 0, p.Hand.Size())
		for _, c := range p.Hand.Cards {
			cards = append(cards, c.InstanceID)
		}
		n := 2
		if len(cards) < n {
			n = len(cards)
		}
		ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    t.ID,
			FromPlayer: t.ID,
			Source:     ctx.Item.SourceCardID,
			Question:   "Perfect Intimidation — exile two cards from your hand",
			Cards:      cards,
			Min:        n,
			Max:        n,
			ThenKey:    ruthlessNegotiationExiled,
			LiveZone:   game.ZoneHand,
		})
		return nil
	}
	return nil
}

// perfectIntimidationRemoveCounters removes every counter from the target.
func perfectIntimidationRemoveCounters(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(t.ID)
	if !ok {
		return nil
	}
	have := socB9CounterCounts(c)
	for _, kind := range socB9CounterKinds(c) {
		if _, err := socB9RemoveCounters(ctx, t.ID, kind, have[kind]); err != nil {
			return err
		}
	}
	return nil
}
