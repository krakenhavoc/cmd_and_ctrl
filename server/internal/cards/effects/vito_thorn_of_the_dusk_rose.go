package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vito, Thorn of the Dusk Rose — Legendary Creature — Vampire Cleric
// {2}{B}, 1/3:
//
//	"Whenever you gain life, target opponent loses that much life.
//	 {3}{B}{B}: Creatures you control gain lifelink until end of turn."
//
// The first line is Sanguine Bond's row (life_gain_drain.go). The
// second is a layer-6 grant snapshotted at resolution (CR 611.2c), so a
// creature that enters afterwards does not gain lifelink.
//
// No simplification. Mass Casualty deck request (#3245).
func init() {
	Register(Spec{
		OracleID:     "7336a541-3ae9-4ab2-acb8-f574c0a5b3a9",
		Name:         "Vito, Thorn of the Dusk Rose",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouGainLifeTargetOpponentLosesThatMuch("Vito, Thorn of the Dusk Rose — target opponent loses that much life"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{B}{B}: Creatures you control gain lifelink until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{3}{B}{B}"),
			Effect: Do(GrantKeywordUntilEOT{
				Match:    And(Creature(), YouControl()),
				Keywords: []string{"lifelink"},
				Label:    "Vito, Thorn of the Dusk Rose — lifelink",
			}),
		}},
	})
}
