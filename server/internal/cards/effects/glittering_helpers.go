package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// glittering_helpers.go — the Glittering cats' two abilities (#1859, ADR
// 0106 amendment of 2026-10-07). Glittering Lion and Glittering Lynx
// print the same pair, differing only in the price of the second:
//
//	"Prevent all damage that would be dealt to this creature.
//	 {N}: Until end of turn, this creature loses "Prevent all damage
//	 that would be dealt to this creature." Any player may activate
//	 this ability."
//
// The first is a standing self replacement (ToThisCreature,
// PreventAllDamageDealtTo), row 0 of Spec.Replacements. The second is an
// any-player row (CR 602.2, 602.1b) whose effect is game.LoseOwnAbilityMod
// naming that row: an ordinary layer-6 effect with its own timestamp
// (CR 613.1f) and an until-end-of-turn duration (CR 611.2a), pinned to the
// object (CR 400.7). It takes the OBJECT'S OWN ability and nothing else,
// so a copy of the cat still has it, and a later grant of the same text is
// a different row that survives.
//
// Neither row declares a Purpose: the bot never pays to make a creature it
// does not control vulnerable (ADR 0106 owner decision 2).
//
// Append-only.

// glitteringPreventionRow is the index of the prevention in the cats'
// Spec.Replacements; the removal names it, and the card tests pin that it
// is the right one.
const glitteringPreventionRow = 0

// glitteringAbilities returns the cat's replacement and its any-player
// row, the latter costing `cost` ("{3}", "{2}").
func glitteringAbilities(name, cost string) ([]game.ReplacementEffect, []ActivatedAbility) {
	reps := []game.ReplacementEffect{
		PreventAllDamageDealtTo(PreventionStatic{
			To:    ToThisCreature,
			Label: name + " — prevent all damage that would be dealt to it",
		}),
	}
	acts := []ActivatedAbility{{
		Label:     cost + `: Until end of turn, this creature loses "Prevent all damage that would be dealt to this creature." Any player may activate this ability.`,
		Cost:      ManaCost(cost),
		AnyPlayer: true,
		// ADR 0142 sweep rulings: turning off a protection is restrict
		// (by analogy, here it is the controller's own).
		Purpose: game.Purpose{Answers: game.AnswerRestrict},
		Effect:  thisLosesOwnReplacementUntilEndOfTurn(glitteringPreventionRow, name+" — loses its damage prevention until end of turn"),
	}}
	return reps, acts
}
