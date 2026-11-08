package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Downwind Ambusher — Creature — Skunk Assassin {3}{B}, 4/2
// (Bloomburrow, #4467):
//
//	"Flash
//	 When this creature enters, choose one —
//	 • Target creature an opponent controls gets -1/-1 until end of turn.
//	 • Destroy target creature an opponent controls that was dealt damage
//	   this turn."
//
// Flash is the printed keyword. The entry trigger is Qutrub Forayer's
// modal shape (CR 700.2): the mode and its target are chosen as the
// trigger goes on the stack. The first bullet is Azula Always Lies'
// -1/-1, declared as a harm until end of turn; the second is Stingblade
// Assassin's destroy over the shared DealtDamageThisTurn predicate. With
// no damaged creature an opponent controls, that bullet cannot be chosen
// (CR 603.3c). A target gone by resolution is skipped (CR 608.2b).
//
// No simplification.
func init() {
	enters := WhenThisEnters("Downwind Ambusher — choose one", func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = ChooseOne(
		ModeWithPurpose(ModeDoing("Target creature an opponent controls gets -1/-1 until end of turn.",
			TargetCreature("target creature an opponent controls", OpponentControls()),
			BoostTheModesTarget(-1, -1, "Downwind Ambusher — -1/-1 until end of turn")), ForTargets(HarmUntilEOT(0, -1, -1))),
		ModeDoing("Destroy target creature an opponent controls that was dealt damage this turn.",
			TargetCreature("target creature an opponent controls that was dealt damage this turn", OpponentControls(), DealtDamageThisTurn()),
			DestroyTheModesTarget),
	)
	Register(Spec{
		OracleID:        "1e0a2709-2cd2-4954-81ca-60654ab03d7c",
		Name:            "Downwind Ambusher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}
