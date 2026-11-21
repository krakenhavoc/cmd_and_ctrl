package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soothsaying — Enchantment {U}:
//
//	"{3}{U}{U}: Shuffle your library.
//	 {X}: Look at the top X cards of your library, then put them back
//	 in any order."
//
// The other half of the {X}-on-an-ability seam, and the opposite
// shape to Treasure Vault's: one X slot, no target, no sacrifice,
// and an effect whose SIZE is the announcement. X=0 is a legal and
// completely pointless activation — the card prints no floor, so the
// engine offers none, and the prompt simply looks at nothing.
//
// The two abilities are printed in this order and kept in it: the
// shuffle is ability 0 and the look is ability 1, because
// `ability_index` is the wire's whole name for an ability and
// reordering them would silently repoint every client that had one
// open.
//
// Both halves lean on machinery that was already here. The look is
// LookAtTopKeyedForEffect — the scry family's third member, whose doc
// comment names this card — so the cards are marked known to the
// controller ONLY and the wire redacts them for everyone else; a
// table that could read your Soothsaying is a real information leak,
// not a cosmetic one. The shuffle is the instruction-in-its-own-
// right twin of the "then shuffle" a tutor runs, and it clears every
// KnownBy in the zone, which is what makes shuffling a live answer
// to an opponent's Sensei's Divining Top rather than a no-op.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "517b702a-2c5c-40d7-825e-6c674019b298",
		Name:         "Soothsaying",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{3}{U}{U}: Shuffle your library.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.ManaCost("{3}{U}{U}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.ShuffleLibrary{Player: item.Controller}.Apply(ctx)
				},
			},
			{
				Label:   "{X}: Look at the top X cards of your library, then put them back in any order.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.ManaCost("{X}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.LookAtTop{Player: item.Controller, N: ctx.X()}.Apply(ctx)
				},
			},
		},
	})
}
