package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mindmelter — {1}{U}{B} Creature — Eldrazi Drone 2/2 (Oath of the Gatewatch,
// S82 OLD b296):
//
//	"Devoid (This card has no color.)
//	 This creature can't be blocked.
//	 {3}{C}: Target opponent exiles a card from their hand. Activate only as a
//	 sorcery."
//
// Dimir Infiltrator's evasion bit and a sorcery-speed activation: the target
// opponent chooses which card to exile (CR 406.3). The {C} in the cost is paid
// with colorless mana only.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dbcceed6-51a9-4ab4-b9a8-d65140661195",
		Name:            "Mindmelter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Static:          []game.StaticAbility{effects.RestrictSelf(game.CantBeBlocked)},
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}{C}: Target opponent exiles a card from their hand. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{3}{C}"),
			Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
			SorcerySpeed: true,
			Effect:       effects.OldB296TargetOpponentExilesFromHand,
		}},
	})
}
