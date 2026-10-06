package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Last Rites — Sorcery {2}{B}:
//
//	"Discard any number of cards. Target player reveals their hand,
//	 then you choose a nonland card from it for each card discarded
//	 this way. That player discards those cards."
//
// A count read from an earlier instruction (#2115, ADR 0116's
// 2026-10-05 amendment). The discard comes first (CR 608.2c): you
// choose any number of cards from your hand, none included, as the
// spell resolves. Once they have been discarded, the number really
// discarded this way is the pick's count — a prompted discard run
// (#1027), so a card that went to exile with madness still counts (CR
// 702.35a: it was discarded). Then the target player reveals their
// hand to the table (CR 701.20a) and you choose that many nonland cards
// from it, as many as there are (CR 609.3), and they are discarded.
// Discarding nothing still reveals the hand.
//
// The pick is queued from the discard's answer while the resolution is
// still open, so state-based actions and triggers wait for it (#1289).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3196e1f4-7f94-4eb6-ae2b-ed11ba392552",
		Name:         "Last Rites",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			victim := TargetedPlayer(ctx)
			if victim == uuid.Nil {
				return nil
			}
			you, source := ctx.Controller(), ctx.Source()
			hand := 0
			if p := ctx.Game.PlayerByIDForEffect(you); p != nil && p.Hand != nil {
				hand = p.Hand.Size()
			}
			return ctx.Game.PlayerDiscardsThenForEffect(game.DiscardPrompt{
				Player:   you,
				Source:   source,
				N:        hand,
				UpTo:     true,
				Question: "Last Rites — discard any number of cards",
			}, func(g *game.Game, discarded game.PromptedDiscards) error {
				nonland := Nonland()
				g.QueueDiscardFromRevealedHand(game.RevealedHandDiscard{
					Chooser:    you,
					FromPlayer: victim,
					Source:     source,
					Count:      discarded.Count(),
					Filter:     func(c game.Card) bool { return nonland(g, you, c) },
					Label:      "nonland card",
				})
				return nil
			})
		},
	})
}
