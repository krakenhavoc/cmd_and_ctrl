package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dreadwing Scavenger — Creature — Nightmare Bird {1}{U}{B}, 2/2
// (Foundations, #4465):
//
//	"Flying
//	 Whenever this creature enters or attacks, draw a card, then discard
//	 a card.
//	 Threshold — This creature gets +1/+1 and has deathtouch as long as
//	 there are seven or more cards in your graveyard."
//
// Bellowing Crier's loot on WhenThisEntersOrAttacks, and Reality
// Fracture's threshold statics: the +1/+1 (layer 7c) and the deathtouch
// grant are read live, so they switch on and off as the graveyard
// crosses seven cards.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bbfd0de0-3558-4833-bae4-8efae44b7fff",
		Name:            "Dreadwing Scavenger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{effects.FraThresholdSelfPT(1, 1), effects.FraThresholdSelfKeywords("deathtouch")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrAttacks("Dreadwing Scavenger — draw a card, then discard a card", effects.B35LootOne),
		},
	})
}
