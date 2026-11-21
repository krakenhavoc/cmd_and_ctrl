package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "7336a541-3ae9-4ab2-acb8-f574c0a5b3a9",
		Name:         "Vito, Thorn of the Dusk Rose",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouGainLifeTargetOpponentLosesThatMuch("Vito, Thorn of the Dusk Rose — target opponent loses that much life"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{B}{B}: Creatures you control gain lifelink until end of turn.",
			Purpose: game.Purpose{TeamPump: &game.Pump{Keywords: []string{"lifelink"}}, Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{3}{B}{B}"),
			Effect: effects.Do(effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Creature(), effects.YouControl()),
				Keywords: []string{"lifelink"},
				Label:    "Vito, Thorn of the Dusk Rose — lifelink",
			}),
		}},
	})
}
