package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// avatar_b20_helpers.go — the shared pieces of Avatar: The Last Airbender
// batches 20 and 21 (#4459).

// avatarB20CyclingTapland is the Avatar common dual cycle (Airship Engine
// Room, Boiling Rock Prison, Foggy Bottom Swamp, Kyoshi Village, Meditation
// Pools, Misty Palms Oasis, North Pole Gates, Omashu City, Serpent's Pass,
// Sun-Blessed Peak):
//
//	"This land enters tapped.
//	 {T}: Add {A} or {B}.
//	 {4}, {T}, Sacrifice this land: Draw a card."
//
// Enters-tapped is the self-replacement every tapland uses; the dual is
// the one-row pipe every two-colour land uses; the cash-in is the Horizon
// Canopy row (b15CanopyLand) at {4}: a CR 602 ability with a
// mana-tap-sacrifice cost that uses the stack, so the land is gone before
// the card is drawn and a response can't save it. Each member has its own
// file and registers through here.
func avatarB20CyclingTapland(oracleID, name, a, b string) Spec {
	return Spec{
		OracleID:      oracleID,
		Name:          name,
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility(a, b)},
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}, Sacrifice this land: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{4}"), TapCost(), SacrificeThis()),
			Effect:  drawOneCardEffect,
		}},
	}
}
