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
//	    Purpose:   game.ActivationPurpose{Draws: 1, ControllerLosesLife: 2},
//	    Effect:    ...,
//	}},
//
// and writes its effect with the activator as "you" (CR 109.5):
// ctx.Controller is the player who activated it, and
// ctx.SourcePermanent().Controller is "this permanent's controller".
//
// Purpose is optional and is for the bot alone (owner decision 2). Set
// it only when the printed effect plainly helps a player who does not
// control the permanent; a row without one is never chosen by a bot
// reaching across the table.

// checkAnyPlayerAbility is ADR 0106 §1 decision 1's registration guard.
// Each failure is a card file that is wrong in a way no game would show
// until somebody tried it:
//
//   - A {T}, loyalty, crew or sacrifice-this component on an any-player
//     row. No printed card has one, and who may tap or sacrifice
//     another player's permanent is a rule nobody has tested.
//   - A zone other than the battlefield. MayActivate keeps the CR 108.4a
//     owner rule off the battlefield, so the permission would be
//     silently ignored there.
//   - A Purpose on a row that is not AnyPlayer, which nothing reads.
func checkAnyPlayerAbility(name, where string, ab ActivatedAbility) {
	if !ab.AnyPlayer {
		if !ab.Purpose.IsZero() {
			panic(fmt.Sprintf("effects.Register: %q %s declares a Purpose but not AnyPlayer — a purpose is for a player who does not control the permanent",
				name, where))
		}
		return
	}
	switch {
	case ab.Cost.Tap:
		panic(fmt.Sprintf("effects.Register: %q %s is an any-player ability with a {T} cost — not modelled (ADR 0106 §1)", name, where))
	case ab.Cost.Loyalty != nil:
		panic(fmt.Sprintf("effects.Register: %q %s is an any-player loyalty ability — not modelled (ADR 0106 §1)", name, where))
	case ab.Cost.Crew > 0:
		panic(fmt.Sprintf("effects.Register: %q %s is an any-player crew ability — not modelled (ADR 0106 §1)", name, where))
	case ab.Cost.SacrificeSelf:
		panic(fmt.Sprintf("effects.Register: %q %s is an any-player ability that sacrifices its source — not modelled (ADR 0106 §1)", name, where))
	}
	for _, z := range ab.Zones {
		if z != game.ZoneBattlefield {
			panic(fmt.Sprintf("effects.Register: %q %s is an any-player ability that functions from the %s — only the battlefield is modelled (ADR 0106 §1)",
				name, where, z))
		}
	}
	if ab.Purpose.Draws < 0 || ab.Purpose.ControllerLosesLife < 0 {
		panic(fmt.Sprintf("effects.Register: %q %s declares a negative Purpose amount", name, where))
	}
}
