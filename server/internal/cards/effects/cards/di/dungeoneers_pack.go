package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dungeoneer's Pack — Artifact {3}:
//
//	"This artifact enters tapped.
//	 {2}, {T}, Sacrifice this artifact: You take the initiative, gain 3
//	 life, draw a card, and create a Treasure token. Activate only as a
//	 sorcery."
//
// Enters tapped is the self-replacement every tapped permanent uses. The
// ability is a sorcery-speed CR 602 ability whose cost includes
// sacrificing the Pack; its effect runs in the printed order, and taking
// the initiative (ADR 0203, #4096) triggers the venture into Undercity,
// which goes on the stack once the ability has resolved.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c66976d3-57f1-4c48-ada0-d74081f1699a",
		Name:         "Dungeoneer's Pack",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Activated: []effects.ActivatedAbility{{
			Label:        "{2}, {T}, Sacrifice this artifact: You take the initiative, gain 3 life, draw a card, and create a Treasure token. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeThis()),
			SorcerySpeed: true,
			Effect: effects.Do(
				effects.TakeTheInitiative{},
				effects.GainLife{Amount: 3},
				effects.DrawCards{N: 1},
				effects.CreateToken{Template: effects.TreasureToken(), N: 1},
			),
		}},
	})
}
