package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Perplex — Instant {1}{U}{B} (Ravnica: City of Guilds):
//
//	"Counter target spell unless its controller discards their hand.
//	 Transmute {1}{U}{B} ({1}{U}{B}, Discard this card: Search your library for a
//	 card with the same mana value as this card, reveal it, put it into your hand,
//	 then shuffle. Transmute only as a sorcery.)"
//
// A pay-unless (CR 118.12a) asked of the TARGET SPELL'S controller as Perplex
// resolves: they either discard every card in their hand (named, as the engine
// names a discard) or the spell is countered. With no cards in hand there is
// nothing to discard and the spell stands. A target that has left the stack makes
// Perplex do nothing (CR 608.2b). Transmute is the keyword's constructor
// (CR 702.53a), finding a mana value 3 card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "904f9e0c-cb7f-4611-8bf7-2e408e8ab237",
		Name:         "Perplex",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.OldB580CounterUnlessItsControllerDiscardsTheirHand(ctx.Game, item)
		},
		Activated: []effects.ActivatedAbility{effects.Transmute("{1}{U}{B}")},
	})
}
