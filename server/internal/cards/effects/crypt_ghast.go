package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crypt Ghast — Creature — Spirit {3}{B}, 2/2:
//
//	"Extort (Whenever you cast a spell, you may pay {W/B}. If you do,
//	 each opponent loses 1 life and you gain that much life.)
//	 Whenever you tap a Swamp for mana, add an additional {B}."
//
// Extort is the shared keyword trigger (extort.go). The second line is
// a CR 605.1b triggered mana ability: it fires as a Swamp of yours is
// tapped for mana, adds its {B} at once with no stack (CR 605.4a), and
// so is spendable on the spell the Swamp was tapped for. The land must
// be a Swamp by its effective type line, so Urborg, Tomb of Yawgmoth
// makes every land count. The extra mana is always {B}, whatever the
// Swamp itself produced, and it is one {B} per Swamp tapped.
//
// Same declared auto-tap simplification as every triggered mana
// ability (ADR 0074 §7): the auto-tapper does not count the extra
// mana, so it floats the difference.
//
// No simplification of the card itself.
func init() {
	Register(Spec{
		OracleID:     "a3c8d817-7949-4dae-b9f5-f9d952479270",
		Name:         "Crypt Ghast",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Extort("Crypt Ghast")},
		ManaTriggers: []game.ManaTrigger{
			WheneverYouTapALandOfSubtypeForMana(
				"Crypt Ghast — add an additional {B}", "Swamp", AddsFixedMana("{B}")),
		},
	})
}
