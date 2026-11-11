package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Epic Experiment — Sorcery {X}{U}{R} (Outlaws of Thunder Junction
// Commander):
//
//	"Exile the top X cards of your library. You may cast instant and sorcery
//	 spells with mana value X or less from among them without paying their
//	 mana costs. Then put all cards exiled this way that weren't cast into
//	 your graveyard."
//
// Hazoret's Undying Fury's window (a free, flash-speed, cast-only permission
// with a cap on each spell and any number of casts), narrowed to instant and
// sorcery spells; the casts are the resolution's (CR 608.2g). Once the prompt
// is answered every card of the exile still there, lands and the spells that
// were not cast included, goes to its owner's graveyard. An X in a cast
// spell's cost is 0 (CR 107.3b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "08bc10fd-8901-4836-ad4d-a867e1bcba77",
		Name:         "Epic Experiment",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			perm := game.CastPermission{
				Cost:        "{0}",
				Timing:      game.TimingFlash,
				CastOnly:    true,
				LapseOnPass: game.LapseStaysInExile,
				Filter:      game.PermissionFilter{InstantOrSorceryOnly: true},
				Source:      ctx.Source(),
				SourceName:  "Epic Experiment",
				Label:       "Cast any number of instant and sorcery spells with mana value " + strconv.Itoa(x) + " or less without paying their mana costs",
			}
			perm.MaxSpellManaValue = &x
			perm.AnyNumber = true
			return ctx.Game.ExileTopWithPermissionKeyedForEffect(ctx.Controller(), ctx.Controller(), x, perm,
				game.KeyedTail{Ref: cmdr2S43EpicExperimentOffer, Player: ctx.Controller(), Source: ctx.Source(), Word: "Epic Experiment"})
		},
	})
}
