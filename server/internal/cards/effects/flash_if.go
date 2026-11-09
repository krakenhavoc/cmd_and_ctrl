package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// flash_if.go — "You may cast this spell as though it had flash if
// <condition>" (#5207): the card's own conditional timing permission,
// Spec.FlashIf (game/flash_if.go). It is not an alternative cost, so a
// flash cast still claims whichever cost the caster picks (CR 118.9a
// limits alternative costs, not timing).
//
// Three constructors, one per shape of condition:
//
//   - FlashIfYou: a condition on the game, read as the cast begins (CR
//     601.3): Bard's Company's "if you control a Human".
//   - FlashIfXAtMost: "if X is N or less" (Silver Scrutiny). CR 601.3b
//     lets the caster begin at instant speed because the X they will
//     announce could make the permission apply; a cast that needs it is
//     then refused above N.
//   - FlashIfTargets: "if it targets <what>" (Timely Ward), the same rule
//     for the target the caster will choose: a cast that needs it is
//     announced under the narrower clause.
//
// A condition on paying a cost ("if you pay {2} more", "if you behold a
// Dragon") is not here: it is an optional cost that opens the window
// (FlashForMore, OptionalBeholdOpeningFlash; CR 601.3c).

// FlashIfYou is "as though it had flash if <cond>" for a condition on the
// game, judged for the caster as the cast begins (CR 601.3).
//
//	FlashIf: FlashIfYou("if you control a Human", youControlAHuman),
func FlashIfYou(label string, cond func(g *game.Game, you uuid.UUID) bool) *game.FlashIf {
	return &game.FlashIf{Label: "You may cast this spell as though it had flash " + label + ".", Condition: cond}
}

// FlashIfXAtMost is "as though it had flash if X is `max` or less" (CR
// 601.3b). The plain sorcery-speed cast announces any X.
//
//	FlashIf: FlashIfXAtMost(3),
func FlashIfXAtMost(max int) *game.FlashIf {
	return &game.FlashIf{
		Label:    fmt.Sprintf("You may cast this spell as though it had flash if X is %d or less.", max),
		XCeiling: max,
	}
}

// FlashIfTargets is "as though it had flash if it targets <what>" (CR
// 601.3b): `clause` is the card's own target clause narrowed to what the
// condition names, which a cast that needs the permission is announced
// under. The plain sorcery-speed cast keeps the printed clause, and so
// does resolution (CR 608.2b).
//
//	FlashIf: FlashIfTargets("a commander", EnchantCreature(ACommander())),
func FlashIfTargets(what string, clause *game.TargetSpec) *game.FlashIf {
	return &game.FlashIf{
		Label:   "You may cast this spell as though it had flash if it targets " + what + ".",
		Targets: clause,
	}
}

// checkFlashIf is Register's guard for Spec.FlashIf: a bound it cannot
// apply, and a narrowed clause on a card with no clause to narrow.
func checkFlashIf(spec Spec) {
	fi := spec.FlashIf
	if fi == nil {
		return
	}
	if fi.XCeiling < 0 {
		panic(fmt.Sprintf("effects.Register: %q declares a FlashIf with a negative XCeiling", spec.Name))
	}
	if fi.Targets != nil && spec.Targets == nil {
		panic(fmt.Sprintf("effects.Register: %q declares a FlashIf target clause but the card has no target clause to narrow", spec.Name))
	}
	if fi.Condition == nil && fi.XCeiling == 0 && fi.Targets == nil {
		panic(fmt.Sprintf("effects.Register: %q declares a FlashIf with no condition; an unconditional \"as though it had flash\" is another shape", spec.Name))
	}
	if spec.FlashWithCleanupSacrifice {
		panic(fmt.Sprintf("effects.Register: %q declares both FlashIf and FlashWithCleanupSacrifice", spec.Name))
	}
}
