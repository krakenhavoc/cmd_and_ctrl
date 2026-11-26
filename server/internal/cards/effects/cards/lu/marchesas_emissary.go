package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Marchesa's Emissary — Creature — Human Rogue {3}{U}, 2/2 (Conspiracy, S82
// OLD batch 363, #5557):
//
//	"Hexproof
//	 Dethrone (Whenever this creature attacks the player with the most life
//	 or tied for most life, put a +1/+1 counter on it.)"
//
// Hexproof is the printed keyword; dethrone is the shared GapB201Dethrone
// trigger (CR 702.105a): the defender is judged as the attack is declared, so
// an attack at a planeswalker never triggers it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7a67cb39-6b53-4f2b-a52c-3861119eba12",
		Name:            "Marchesa's Emissary",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
		Triggered: []game.TriggeredAbility{
			effects.GapB201Dethrone("Marchesa's Emissary — dethrone: put a +1/+1 counter on it"),
		},
	})
}
