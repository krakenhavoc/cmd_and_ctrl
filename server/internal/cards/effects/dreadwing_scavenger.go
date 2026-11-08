package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "bbfd0de0-3558-4833-bae4-8efae44b7fff",
		Name:            "Dreadwing Scavenger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{fraThresholdSelfPT(1, 1), fraThresholdSelfKeywords("deathtouch")},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Dreadwing Scavenger — draw a card, then discard a card", b35LootOne),
		},
	})
}
