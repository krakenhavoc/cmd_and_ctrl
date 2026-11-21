package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Clash of the Eikons — Sorcery {G} (Final Fantasy, #4462):
//
//	"Choose one or more —
//	 • Target creature you control fights target creature an opponent
//	   controls.
//	 • Remove a lore counter from target Saga you control. (Removing lore
//	   counters doesn't cause chapter abilities to trigger.)
//	 • Put a lore counter on target Saga you control."
//
// Inscription of Abundance's fight bullet (two target clauses on one
// bullet, CR 701.14, fighting only while both are still legal), and two
// lore-counter bullets that each take their own Saga (the two may name the
// same one, since each bullet has its own targets). The bullets resolve in
// printed order (CR 700.2). Removing a lore counter puts nothing on the
// stack (CR 714.2b: a chapter ability triggers when lore counters are put
// onto the Saga, not when one is removed), and putting one on triggers
// its next chapter in the ordinary way.
//
// No simplification.
func init() {
	sagaYouControl := func() *game.TargetSpec {
		return effects.TargetPermanent("target Saga you control", effects.OfSubtype("Saga"), effects.YouControl())
	}
	effects.Register(effects.Spec{
		OracleID:     "b1c38b70-575d-4d87-bc50-0b77b48ee823",
		Name:         "Clash of the Eikons",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOneOrMore(
			effects.ModeWithPurpose(effects.ModeDoing("Target creature you control fights target creature an opponent controls.",
				effects.Clauses(
					effects.TargetCreature("target creature you control", effects.YouControl()),
					effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
				),
				effects.FightTheModesTargets), game.Purpose{Answers: game.AnswerRemove}),
			effects.ModeDoing("Remove a lore counter from target Saga you control.", sagaYouControl(),
				effects.FinB3LoreCounterOnTheModesTarget(-1)),
			effects.ModeDoing("Put a lore counter on target Saga you control.", sagaYouControl(),
				effects.FinB3LoreCounterOnTheModesTarget(1)),
		),
	})
}
