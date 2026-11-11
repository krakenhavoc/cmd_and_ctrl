package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Boros Reckoner — Creature — Minotaur Wizard {R/W}{R/W}{R/W}, 3/3 (Lost
// Caverns of Ixalan Commander):
//
//	"Whenever this creature is dealt damage, it deals that much damage to any
//	 target.
//	 {R/W}: This creature gains first strike until end of turn."
//
// Wrathful Raptors' reflection: the trigger waits for the damage to finish
// and reflects the TOTAL the Reckoner was dealt (CR 120.3, CR 510.2), from
// the Reckoner, even if it died to it.
func init() {
	const label = "Boros Reckoner — first strike until end of turn"
	Register(Spec{
		OracleID:     "5f079a75-899e-467d-b443-3b87ce2fb548",
		Name:         "Boros Reckoner",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(cmdr2B08EnrageTrigger("Boros Reckoner — it deals that much damage to any target", damagedCreatureReflectsChosenAmount),
				TargetAny()),
		},
		Activated: []ActivatedAbility{{
			Label:   "{R/W}: This creature gains first strike until end of turn.",
			Cost:    ManaCost("{R/W}"),
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GrantKeywordUntilEOT{Target: NewContext(g, item).Source(), Keywords: []string{"first strike"}, Label: label}.Apply(NewContext(g, item))
			},
		}},
	})
}
