package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// any_player_activation.go — "Any player may activate this ability."
// ADR 0106 §1, #1793. The engine half is game/any_player_activation.go.
//
// A card file writes one bit on the row:
//
//	Activated: []ActivatedAbility{{
//	    Label:     "{3}: Xantcha's controller loses 2 life and you draw a card. Any player may activate this ability.",
//	    Cost:      game.AbilityCost{Mana: "{3}"},
//	    AnyPlayer: true,
//	    Purpose:   game.Purpose{Draws: 1, ControllerLosesLife: 2},
//	    Effect:    ...,
//	}},
//
// and writes its effect with the activator as "you" (CR 109.5):
// ctx.Controller is the player who activated it, and
// ctx.SourcePermanent().Controller is "this permanent's controller".
//
// Purpose is optional and is for the bot alone (owner decision 2). On
// an any-player row, set it only when the printed effect plainly helps
// a player who does not control the permanent: a row without one is
// never chosen by a bot reaching across the table, and the bot prices
// another player's row by its Draws and ControllerLosesLife alone.

// checkAnyPlayerAbility is ADR 0106 §1 decision 1's registration guard.
// Each failure is a card file that is wrong in a way no game would show
// until somebody tried it:
//
//   - A {T}, loyalty, crew, sacrifice-this or exert component on an any-player
//     row. No printed card has one, and who may tap or sacrifice
//     another player's permanent is a rule nobody has tested.
//   - A zone other than the battlefield. MayActivate keeps the CR 108.4a
//     owner rule off the battlefield, so the permission would be
//     silently ignored there.
//
// What the row's Purpose may say is checkActivatedPurpose's (purpose.go):
// ControllerLosesLife only here, on an any-player row.
func checkAnyPlayerAbility(name, where string, ab ActivatedAbility) {
	named := 0
	for _, b := range []bool{ab.AnyPlayer, ab.OpponentsOnly, ab.OwnerOnly} {
		if b {
			named++
		}
	}
	if named == 0 {
		return
	}
	switch {
	case named > 1:
		panic(fmt.Sprintf("effects.Register: %q %s names more than one of AnyPlayer, OpponentsOnly and OwnerOnly (ADR 0106 §1)", name, where))
	case ab.OpponentsOnly && ab.Purpose != (game.Purpose{}):
		panic(fmt.Sprintf("effects.Register: %q %s declares a Purpose on an opponents-only row; the bot never reads one there (ADR 0106 §1 amendment 2026-10-07)", name, where))
	case ab.Cost.Tap:
		panic(fmt.Sprintf("effects.Register: %q %s is an any-player ability with a {T} cost — not modelled (ADR 0106 §1)", name, where))
	case ab.Cost.Loyalty != nil:
		panic(fmt.Sprintf("effects.Register: %q %s is an any-player loyalty ability — not modelled (ADR 0106 §1)", name, where))
	case ab.Cost.Crew > 0:
		panic(fmt.Sprintf("effects.Register: %q %s is an any-player crew ability — not modelled (ADR 0106 §1)", name, where))
	case ab.Cost.SacrificeSelf:
		panic(fmt.Sprintf("effects.Register: %q %s is an any-player ability that sacrifices its source — not modelled (ADR 0106 §1)", name, where))
	case ab.Cost.Exert:
		panic(fmt.Sprintf("effects.Register: %q %s is an any-player ability that exerts its source — not modelled (ADR 0106 §1, ADR 0130 §4)", name, where))
	}
	for _, z := range ab.Zones {
		if z != game.ZoneBattlefield {
			panic(fmt.Sprintf("effects.Register: %q %s is an any-player ability that functions from the %s — only the battlefield is modelled (ADR 0106 §1)",
				name, where, z))
		}
	}
}
