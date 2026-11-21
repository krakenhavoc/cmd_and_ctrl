package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skittering Precursor — Creature — Eldrazi Drone {2}{R}, 3/3 (Modern
// Horizons 3):
//
//	"Devoid (This card has no color.)
//	 Menace
//	 Whenever you sacrifice a nontoken permanent, create a 0/1 colorless
//	 Eldrazi Spawn creature token with "Sacrifice this token: Add {C}.""
//
// Devoid and menace are printed keywords. The trigger watches
// EventSacrifice (announced while the permanent is still there, so a
// Precursor that sacrifices itself sees its own sacrifice), for the
// controller's nontoken permanents only; one sacrifice, one token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7b1519b1-665c-47b8-8f96-0027410ed81c",
		Name:            "Skittering Precursor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid, "menace"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventSacrifice, effects.GapB53YouSacrificedANontoken,
				"Skittering Precursor — create a 0/1 colorless Eldrazi Spawn",
				effects.Do(effects.CreateToken{Template: effects.EldraziSpawnToken(), N: 1})),
		},
	})
}
