package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skull Raid — Sorcery {3}{B} (Kaldheim):
//
//	"Target opponent discards two cards. If fewer than two cards were
//	 discarded this way, you draw cards equal to the difference.
//	 Foretell {1}{B} (During your turn, you may pay {2} and exile this card
//	 from your hand face down. Cast it on a later turn for its foretell cost.)"
//
// The opponent chooses the cards (a real discard prompt, CR 701.9b); what was
// actually discarded is counted when the discard is done, so a one-card hand
// gives you one draw and an empty hand two. A target that is gone does
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "7d22c7ed-2f45-42af-b999-b921f4f4db25",
		Name:           "Skull Raid",
		Completeness:   effects.CompletenessFull,
		SpecialActions: []game.SpecialAction{effects.Foretell("{1}{B}")},
		Targets:        effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			var victim uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetPlayer {
					victim = t.ID
				}
			}
			if victim == uuid.Nil {
				return nil
			}
			return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
				Player: victim,
				Source: item.SourceCardID,
				N:      2,
				Asked:  game.NewAskedItem(item),
			}, skullRaidDiscarded, []uuid.UUID{item.Controller})
		},
	})
}

// skullRaidDiscarded draws the difference for the caster. An on-disk key:
// never renamed, never reused.
var skullRaidDiscarded = game.RegisterDiscardThen("choose-cards/s82-old-b063-skull-raid-discarded", func(g *game.Game, r game.DiscardsDone) error {
	if len(r.Carry) == 0 {
		return nil
	}
	short := 2 - len(r.Discarded.By(r.Player))
	if short <= 0 {
		return nil
	}
	return g.DrawNForEffect(r.Carry[0], short)
})
