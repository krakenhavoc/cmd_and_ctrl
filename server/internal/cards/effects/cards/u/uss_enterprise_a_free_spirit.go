package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// U.S.S. Enterprise-A, Free Spirit — Legendary Artifact — Spacecraft
// {3}{W}{W}, 5/5 (Star Trek):
//
//	Ward {2}
//	When U.S.S. Enterprise-A enters, create three 1/1 red Officer
//	creature tokens.
//	Station (Tap another creature you control: Put charge counters
//	equal to its power on this Spacecraft. Station only as a sorcery.
//	It's an artifact creature at 6+.)
//	6+ | Flying, lifelink
//
// The Seriema's station shape (Station, SpacecraftAt, ThresholdKeywords)
// with ward above the bar: ward protects the permanent whether or not it
// is a creature yet. The three Officers can station it the turn after.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5e2958fd-ede6-4563-babc-4924d3e95754",
		Name:         "U.S.S. Enterprise-A, Free Spirit",
		Purpose:      game.Purpose{Tokens: 3, CreatureTokens: 3},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "U.S.S. Enterprise-A — ward {2}"),
			effects.WhenThisEnters("U.S.S. Enterprise-A — create three 1/1 red Officer creature tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 red Officer"), N: 3})),
		},
		Activated: []effects.ActivatedAbility{effects.Station()},
		Static: append(
			effects.SpacecraftAt(6, 5, 5),
			effects.ThresholdKeywords(6, "flying", "lifelink"),
		),
	})
}
