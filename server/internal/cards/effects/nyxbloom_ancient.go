package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nyxbloom Ancient — Enchantment Creature — Elemental {4}{G}{G}{G},
// 5/5 (EDHREC rank ~900):
//
//	"Trample
//	 If you tap a permanent for mana, it produces three times as much
//	 of that mana instead."
//
// Mana Reflection with a bigger number and a body. The same declared
// effect in a different card file would be wrong — a THREE is not a
// TWO, so these are two distinct declared effects and CR 616.1 orders
// them when both are out (#792's identical-window skip covers two
// Nyxblooms, not a Nyxbloom and a Reflection). With both on the
// battlefield a Forest makes six, in either order, which is why the
// ordering is never actually put to anybody: a production cannot pause
// (CR 605.3b, ADR 0013 §5ab) and multiplication commutes.
//
// Trample rides PrintedKeywords. Everything else about the mana half
// is written on Mana Reflection's card file — where the window opens
// for each shape of source, and what a tap for mana is not (a spell's
// "Add {B}{B}{B}", an untapped mana ability, a triggered mana
// ability's own output).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "8b610f8f-c8dd-4eeb-bc6e-3bc706d5f63e",
		Name:            "Nyxbloom Ancient",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements: []game.ReplacementEffect{
			YouTapForThriceAsMuchMana("Nyxbloom Ancient: three times as much of that mana"),
		},
	})
}
