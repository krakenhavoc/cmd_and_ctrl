package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pinchy McStingbutt — Creature — Scorpion {2}{B}, 1/3 (Mystery Booster
// 2, #5948):
//
//	"Deathtouch
//	 Whenever Pinchy McStingbutt deals combat damage to a player, conjure
//	 a card named Death of a Thousand Stings, Lethal Sting, or Stinging
//	 Shot at random into your hand and reveal it. (Treat conjured cards
//	 like cards.)"
//
// The name is drawn from the keyed pick stream, so a restore replays it;
// the card is a real card (ADR 0099 amendment 2026-11-24) and is
// revealed to the table.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4ad0664d-fede-4b75-bb3f-53d5060ca01d",
		Name:            "Pinchy McStingbutt",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Pinchy McStingbutt — conjure a sting at random into your hand and reveal it",
				effects.Do(effects.CheckConjure("Pinchy McStingbutt", effects.Conjure{
					Names:    []string{"Death of a Thousand Stings", "Lethal Sting", "Stinging Shot"},
					Zone:     game.ZoneHand,
					AtRandom: true,
					Reveal:   true,
				}))),
		},
	})
}
