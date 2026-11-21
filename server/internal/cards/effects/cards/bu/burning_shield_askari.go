package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Burning Shield Askari — Creature — Human Knight {2}{R}, 2/2 (Mirage,
// #5633):
//
//	"Flanking (Whenever a creature without flanking blocks this creature,
//	 the blocking creature gets -1/-1 until end of turn.)
//	 {R}{R}: This creature gains first strike until end of turn."
//
// Flanking is flanking.go's trigger and the declared keyword (#5757), so
// a blocker with flanking is spared (CR 702.25a). The ability is a
// layer-6 grant pinned to the source through the shared
// thisGainsKeywordUntilEndOfTurn body; gained after blockers are declared
// and before combat damage, it still makes the Askari deal its damage in
// the first-strike step (CR 510.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "042fd95a-5e85-4ce5-bb68-066a295c54fb",
		Name:            "Burning Shield Askari",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordFlanking},
		Triggered:       []game.TriggeredAbility{effects.RetiredFlankingRow("Burning Shield Askari")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}{R}: This creature gains first strike until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant, Pump: &game.Pump{Keywords: []string{"first strike"}}},
			Cost:    effects.ManaCost("{R}{R}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("first strike", "Burning Shield Askari — first strike until end of turn"),
		}},
	})
}
