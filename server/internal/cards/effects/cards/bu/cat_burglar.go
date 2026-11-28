package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cat Burglar — Creature — Kor Rogue Minion {3}{B}, 2/2 (Exodus, S82 OLD b737, #5626):
//
//	"{2}{B}, {T}: Target player discards a card. Activate only as a sorcery."
//
// Wall of Distortion's discard on a target player (who chooses the card,
// CR 701.9a), gated to the sorcery-speed window at activation (CR 602.5d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "16aea5b3-9f43-4852-b8b8-80e87909b794",
		Name:         "Cat Burglar",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{2}{B}, {T}: Target player discards a card. Activate only as a sorcery.",
			Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, Discards: 1}),
			Cost:         effects.Plus(effects.ManaCost("{2}{B}"), effects.TapCost()),
			SorcerySpeed: true,
			Targets:      effects.TargetPlayer("target player"),
			Effect:       effects.TargetPlayerDiscards(1),
		}},
	})
}
