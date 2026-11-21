package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Marit Lage's Slumber — Legendary Snow Enchantment {1}{U} (Jumpstart 2022,
// #5948):
//
//	"Whenever Marit Lage's Slumber or another snow permanent you control
//	 enters, scry 1.
//	 At the beginning of your upkeep, if you control ten or more snow
//	 permanents, sacrifice Marit Lage's Slumber. If you do, create Marit Lage,
//	 a legendary 20/20 black Avatar creature token with flying and
//	 indestructible."
//
// The scry is one entry trigger over the source or any snow permanent that
// entered under its controller. The upkeep ability is an intervening "if"
// (CR 603.4), checked when it would trigger and again as it resolves; the
// sacrifice and the Dark Depths token follow Dark Depths (SacrificeThisThen),
// so Marit Lage is made only if the enchantment really left.
//
// No simplification.
func init() {
	const label = "Marit Lage's Slumber — sacrifice it; if you do, create Marit Lage"
	effects.Register(effects.Spec{
		OracleID:     "a11d890e-38d4-467c-bc04-948979c142e0",
		Name:         "Marit Lage's Slumber",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.GapB169SnowYouControlEntered,
				"Marit Lage's Slumber — scry 1",
				effects.Do(effects.Scry{N: 1})),
			effects.InterveningIf(effects.GapB169CondSnowPermanentsAtLeast(10),
				effects.On(game.EventBeginUpkeep, effects.AllOf(effects.ByYou, effects.GapB169SnowPermanentsAtLeast(10)), label,
					effects.SacrificeThisThen("then/marit-lages-slumber-marit-lage", func(ctx *effects.Context) error {
						return effects.CreateToken{Template: effects.TokenCard("20/20 black Marit Lage"), N: 1}.Apply(ctx)
					}))),
		},
	})
}
