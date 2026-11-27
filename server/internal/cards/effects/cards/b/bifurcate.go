package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bifurcate — Sorcery {3}{G} (Mercadian Masques, S82 OLD, #5619):
//
//	"Search your library for a permanent card with the same name as target
//	 nontoken creature, put that card onto the battlefield, then shuffle."
//
// The name is read from the target as the spell resolves; a target that has
// gone (CR 608.2b) fizzles the spell before this runs. Only a permanent card can
// go onto the battlefield, so the search offers just those, and it may find
// nothing (CR 701.23b). A face-down creature has no name (CR 708.2), so
// nothing matches it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f21ec455-aaf6-4842-aea2-69a7fd99b80f",
		Name:         "Bifurcate",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Targets:      effects.TargetCreature("target nontoken creature", effects.Not(effects.IsTokenPredicate())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok || c.FaceDown || c.Name == "" {
					return nil
				}
				name := c.Name
				return effects.SearchLibrary{
					Player:    ctx.Controller(),
					Predicate: func(l game.Card) bool { return l.Name == name && l.IsPermanent() },
					Dest:      game.ZoneBattlefield,
					Limit:     1,
					Shuffle:   true,
					Reason:    "Bifurcate — a permanent card named " + name,
				}.Apply(ctx)
			}
			return nil
		},
	})
}
