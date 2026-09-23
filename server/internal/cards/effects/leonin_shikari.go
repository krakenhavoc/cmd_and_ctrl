package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leonin Shikari — Creature — Cat Soldier {1}{W}, 2/2:
//
//	"You may activate equip abilities any time you could cast an
//	 instant."
//
// One `Spec.ActivationTimings` entry and nothing else. The whole card
// is the sentence #1208 built: a statement about a PLAYER's
// activations, derived from the battlefield on every query, so two
// Shikari compose (harmlessly — the verdict is a bit, not a count), a
// Shikari that has lost its abilities under CR 613.1f stops saying
// it, and one dying in response to an equip announcement shuts the
// window before the announcement is validated.
//
// IT IS THE ABILITY THAT IS NARROWED, not the source. Every other
// statement on this seam names an object ("loyalty abilities of
// Teferi"); this one names a KEYWORD, and the Equipment carrying it
// could be any artifact on the board. That is why #1208 put the
// narrowing on `ActivationAbility` rather than on a
// `PermissionFilter` over card types, and why `EquipAbility` marks
// the ability it builds — see activation_timing.go.
//
// "You may activate" is the SHIKARI's controller, so an opponent's
// Equipment keeps its sorcery-speed window. The clause opens the
// window; it does not touch the equip COST, CR 702.6d's "re-activate
// to move it", or the target clause.
//
// No simplification, and no caveat: there is nothing else on the
// card.
func init() {
	Register(Spec{
		OracleID:     "857d94f7-113c-45a4-a88a-5d087347d57f",
		Name:         "Leonin Shikari",
		Completeness: CompletenessFull,
		ActivationTimings: []game.ActivationTiming{
			EquipAbilitiesAtInstantSpeed(
				"You may activate equip abilities any time you could cast an instant.", false),
		},
	})
}
