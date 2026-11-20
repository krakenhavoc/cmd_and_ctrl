package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thalakos Dreamsower — Creature — Thalakos Wizard (1/1) for {2}{U}:
//
//	"Shadow (This creature can block or be blocked by only creatures with shadow.)
//	 You may choose not to untap this creature during your untap step.
//	 Whenever this creature deals damage to an opponent, tap target creature. That creature doesn't untap during its controller's untap step for as long as this creature remains tapped."
//
// ADR 0109 owner decision 4: the trigger's rider is Rust Tick's untap
// hold (TapAndHoldWhileThisRemainsTapped). The hold only starts if the
// Dreamsower is still tapped as the trigger resolves (CR 611.2b); the
// tap happens either way.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5f71c09a-47fb-4bb2-9329-78f5bb195e5d",
		Name:            "Thalakos Dreamsower",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"shadow"},
		UntapOptOuts:    []game.UntapOptOut{effects.MayChooseNotToUntapSelf("Thalakos Dreamsower — you may choose not to untap this creature")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventDealDamage, effects.ThisDealtDamageToAnOpponent,
				"Thalakos Dreamsower — tap target creature; it doesn't untap while this remains tapped",
				effects.TapTargetsAndHoldWhileThisRemainsTapped), effects.TargetCreature("target creature")),
		},
	})
}
