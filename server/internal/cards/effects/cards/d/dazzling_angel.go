package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dazzling Angel — Creature — Angel {2}{W}, 2/3 (EDHREC rank 3071):
//
//	"Flying
//	 Whenever another creature you control enters, you gain 1 life."
//
// Soul Warden on a flier, for your own creatures only. Flying rides
// PrintedKeywords; the trigger is Impact Tremors' condition
// (b13AnotherCreatureYouControlEntered — tokens included, the Angel
// itself excluded) and the life is the source's, so a lifegain
// payoff sees it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7de464e3-fae3-44cc-8233-776fc727c00a",
		Name:            "Dazzling Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.CreatureEnterPaying(effects.WheneverAnotherCreatureEntersUnderYourControl("Dazzling Angel — you gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
				game.CreatureEnterPayoff{LifeGain: 1}),
		},
	})
}
