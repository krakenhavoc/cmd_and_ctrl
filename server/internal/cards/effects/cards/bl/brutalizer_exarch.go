package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brutalizer Exarch — Creature — Phyrexian Cleric {5}{G}, 3/3 (Planechase 2012,
// S82 OLD, #5569):
//
//	"When this creature enters, choose one —
//	 • Search your library for a creature card, reveal it, then shuffle and put
//	   that card on top.
//	 • Put target noncreature permanent on the bottom of its owner's library."
//
// The mode is chosen as the trigger is put on the stack (CR 700.2), and only the
// second mode targets (CR 601.2c). The first is Moon-Blessed Cleric's search to
// the top (find, reveal, shuffle, then the card on top; it may find nothing, CR
// 701.23b); the second is Run Aground's tuck to the bottom over a noncreature
// permanent, which a target gone in response does nothing to.
//
// No simplification.
func init() {
	exarch := effects.WhenThisEnters("Brutalizer Exarch — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	exarch.Modes = effects.ChooseOne(
		effects.ModeDoing("Search your library for a creature card, reveal it, then shuffle and put that card on top.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.IsCreature() },
					Dest:      game.ZoneLibrary,
					ToTop:     true,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Brutalizer Exarch — a creature card",
				}.Apply(ctx)
			}),
		effects.ModeDoing("Put target noncreature permanent on the bottom of its owner's library.",
			effects.TargetPermanent("target noncreature permanent", effects.Noncreature()),
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				if t, ok := effects.FirstTargetCard(item); ok && ctx.IsTargetLegal(t) {
					return ctx.Game.TuckToLibraryForEffect(t.ID, true)
				}
				return nil
			}),
	)
	effects.Register(effects.Spec{
		OracleID:     "bdec14d1-1b38-4223-869e-a0ffc76b2b0c",
		Name:         "Brutalizer Exarch",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered:    []game.TriggeredAbility{exarch},
	})
}
