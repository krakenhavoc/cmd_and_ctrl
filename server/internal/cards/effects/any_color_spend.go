package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// any_color_spend.go — the card-side sentences over
// game/spend_any_color.go (#1600, ADR 0066's 2026-10-02 amendment):
// a PERMANENT's "you may spend mana as though it were mana of any
// color" (CR 609.4b), declared on Spec.AnyColorSpend.
//
// The engine owns the rest: it reads the grant at every payment the
// covered player makes, widens each coloured symbol to any mana while
// still paying it with its printed colour when the pool has one, and
// leaves {C} alone (CR 106.1b). A card says only whose payments and,
// for the narrowed form, which ones.
//
// A cast PERMISSION's "you may spend mana as though it were mana of any
// color to cast that spell" (Breeches, impulse exile) is
// CastPermission.AnyColor, which belongs to one spell, not this.

// YouMaySpendManaAsAnyColor is "You may spend mana as though it were
// mana of any color." — every cost the permanent's controller pays
// (Chromatic Orrery).
func YouMaySpendManaAsAnyColor() []game.AnyColorSpendStatic {
	return []game.AnyColorSpendStatic{{
		Label: "You may spend mana as though it were mana of any color.",
		Whose: game.AnyColorSpendYou,
	}}
}

// PlayersMaySpendManaAsAnyColor is "Players may spend mana as though it
// were mana of any color." — every cost every player pays (Mycosynth
// Lattice).
func PlayersMaySpendManaAsAnyColor() []game.AnyColorSpendStatic {
	return []game.AnyColorSpendStatic{{
		Label: "Players may spend mana as though it were mana of any color.",
		Whose: game.AnyColorSpendEveryPlayer,
	}}
}

// YouMaySpendManaAsAnyColorToCast is "You may spend mana as though it
// were mana of any color to cast <what>." — narrowed to casting a spell
// with one of `types` (Oath of Nissa's "planeswalker spells"). It
// reaches neither an activation nor any other cost the controller pays.
func YouMaySpendManaAsAnyColorToCast(what string, types ...string) []game.AnyColorSpendStatic {
	return []game.AnyColorSpendStatic{{
		Label: "You may spend mana as though it were mana of any color to cast " + what + ".",
		Whose: game.AnyColorSpendYou,
		Covers: func(ctx game.ManaSpendContext) bool {
			if ctx.Purpose != game.SpendPurposeCast {
				return false
			}
			for _, have := range ctx.Types {
				for _, want := range types {
					if strings.EqualFold(have, want) {
						return true
					}
				}
			}
			return false
		},
	}}
}
