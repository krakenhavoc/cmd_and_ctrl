package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wake the Dragon — Sorcery {4}{B}{R} (Tales of Middle-earth Commander):
//
//	"Create a 6/6 black and red Dragon creature token with flying, menace, and
//	 "Whenever this token deals combat damage to a player, gain control of
//	 target artifact that player controls."
//	 Flashback {6}{B}{R} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// The Dragon is a catalog token template (ADR 0083), so its trigger travels
// with the token. The trigger's target is an artifact controlled by the player
// the Dragon just damaged ("that player controls", read off
// the damage event), chosen as the ability goes on the stack; the control
// lasts indefinitely (CR 611.2a), as Hellkite Tyrant's does. Flashback is the
// shared alternative cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "cfa5b168-d112-4b92-ac42-8bba30f15d77",
		Name:             "Wake the Dragon",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Tokens: 1, CreatureTokens: 1},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{6}{B}{R}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Template: effects.S82GapB120DragonToken(), N: 1}.Apply(ctx)
		},
	})
}
