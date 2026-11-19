package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Consign to Memory — Instant {U} (Special Guests):
//
//	"Replicate {1}
//	 Counter target triggered ability or colorless spell."
//
// Replicate is the engine's keyword (replicate.go, CR 702.56a) with a
// mana cost. The clause is one pick over the stack: any triggered ability
// (an activated ability or a coloured spell is refused at announce, CR
// 601.2c, and the target is judged again as the spell resolves, CR 608.2b).
// A copy may choose a new target (CR 707.10c).
func init() {
	Register(Spec{
		OracleID:      "16c39265-72d9-4c86-a961-8dd38ff9adcd",
		Name:          "Consign to Memory",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Answers: game.AnswerRemove},
		OptionalCosts: []game.AdditionalCost{replicateMana("{1}", 20)},
		Triggered:     []game.TriggeredAbility{Replicate()},
		Targets: TargetSpellOrAbility("target triggered ability or colorless spell",
			AnyStackItem(bnsB06TriggeredAbilityItem(), ASpellItem(Colorless()))),
		OnResolve: counterTheTargetSpell,
	})
}
