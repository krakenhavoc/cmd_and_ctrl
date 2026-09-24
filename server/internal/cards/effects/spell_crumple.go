package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spell Crumple — Instant {1}{U}{U}:
//
//	"Counter target spell. If that spell is countered this way, put it
//	 on the bottom of its owner's library instead of into that player's
//	 graveyard. Put Spell Crumple on the bottom of its owner's library."
//
// #1298, on ADR 0088's put_in_library (2026-09-23 amendment). Hinder's
// sibling with no choice in it: a pile of one on a one-lane placement
// raises no prompt, so CounterToLibrary with LibraryPlaceBottom is the
// counter to the bottom and nothing else. The shared stack exit keeps a
// countered commander's CR 903.9 offer and flashback's exile.
//
// "Put Spell Crumple on the bottom of its owner's library" is the
// card's own instruction, so it runs inside the resolution, where the
// resolution frame sees the spell already gone and leaves it there
// (#489, spellMovedItselfLocked). It is written on the next line rather
// than in the counter's Then on purpose: the frame routes a still-
// resolving spell to the graveyard as soon as OnResolve returns, so a
// Then that ran later — after a countered commander's owner answered
// CR 903.9 — would tuck Spell Crumple out of the graveyard, having
// already been put there. The cost of the next line is an order no one
// sees outside one corner: countering your OWN commander, and declining
// the command zone, puts it under Spell Crumple instead of over it. A
// copy of the spell has no card to put anywhere (CR 707.10) and skips
// the second sentence.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d4ab7848-5c37-4c6b-be29-0bb703333e5b",
		Name:         "Spell Crumple",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) > 0 {
				if err := (CounterToLibrary{
					StackID:   item.Targets[0].ID,
					Placement: game.LibraryPlaceBottom,
					Label:     "Spell Crumple — put it on the bottom of its owner's library",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			if item.IsCopy {
				return nil
			}
			return PutIntoLibrary{Card: ctx.Source(), ToBottom: true}.Apply(ctx)
		},
	})
}
