package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shaman of Forgotten Ways — Creature — Human Shaman {2}{G}, 2/3 (Dragons of
// Tarkir, S82 older sets):
//
//	"{T}: Add two mana in any combination of colors. Spend this mana only to cast
//	 creature spells.
//	 Formidable — {9}{G}{G}, {T}: Each player's life total becomes the number of
//	 creatures they control. Activate only if creatures you control have total
//	 power 8 or greater."
//
// The first ability is Roving Keep's any-combination mana ability for two with
// Bucolic Ranch's spend restriction over creature spells (CR 106.6); the engine
// refuses the tap while the Shaman is summoning sick (CR 302.6). The second needs
// the formidable total (CR 602.5b), read as it is announced: every player still in
// the game is counted as it resolves, then each life total is set (CR 118.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ca405696-8c30-4c01-bf90-11cdf8837942",
		Name:         "Shaman of Forgotten Ways",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:         effects.ManaAbilityCost{Tap: true},
			Produced:     effects.AnyCombinationOfColors(2),
			Restrictions: []string{effects.ManaRestrictCast, effects.ManaRestrictType("Creature")},
			Label:        "{T}: Add two mana in any combination of colors. Spend this mana only to cast creature spells.",
		}},
		Activated: []effects.ActivatedAbility{effects.OldB324FormidableRow(effects.ActivatedAbility{
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{9}{G}{G}"), effects.TapCost()),
			Effect:  effects.OldB324EachPlayersLifeBecomesTheirCreatures,
		}, "{9}{G}{G}, {T}: Each player's life total becomes the number of creatures they control.")},
	})
}
