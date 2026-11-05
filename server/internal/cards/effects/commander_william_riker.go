package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Commander William Riker — Legendary Creature — Human Officer {1}{R}{W},
// 3/3 (Star Trek Commander, #4974):
//
//	"Vigilance, haste
//	 {1}, {T}: You and target opponent each draw two cards."
//
// Vigilance and haste are printed and enforced; haste lets the {T} be paid
// the turn Riker arrives (CR 302.6). The ability is Loran of the Third
// Path's with two cards each: you draw two, then the opponent it targets
// draws two. If that opponent is no longer a legal target the whole ability
// does nothing (CR 608.2b), so you draw nothing either.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cef73cb5-88c2-43cd-8648-1baca085b8a4",
		Name:            "Commander William Riker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "haste"},
		Activated: []ActivatedAbility{{
			Label: "{1}, {T}: You and target opponent each draw two cards.",
			// ADR 0126 §6: the opponent draws too, which no field says.
			Purpose: game.Purpose{Draws: 2, Targets: game.ForTargets(game.TargetPurpose{Slot: 0, Draws: 2})},
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Targets: TargetPlayer("target opponent", Opponent()),
			Effect:  trekB39YouAndTargetOpponentEachDraw(2),
		}},
	})
}
