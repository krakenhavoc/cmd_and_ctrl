package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	enters := effects.WhenThisEnters("Downwind Ambusher — choose one", func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = effects.ChooseOne(
		effects.ModeWithPurpose(effects.ModeDoing("Target creature an opponent controls gets -1/-1 until end of turn.",
			effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
			effects.BoostTheModesTarget(-1, -1, "Downwind Ambusher — -1/-1 until end of turn")), effects.ForTargets(effects.HarmUntilEOT(0, -1, -1))),
		effects.ModeDoing("Destroy target creature an opponent controls that was dealt damage this turn.",
			effects.TargetCreature("target creature an opponent controls that was dealt damage this turn", effects.OpponentControls(), effects.DealtDamageThisTurn()),
			effects.DestroyTheModesTarget),
	)
	effects.Register(effects.Spec{
		OracleID:        "1e0a2709-2cd2-4954-81ca-60654ab03d7c",
		Name:            "Downwind Ambusher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}
