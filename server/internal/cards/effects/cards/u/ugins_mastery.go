package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ugin's Mastery — Enchantment {4} (Murders at Karlov Manor Commander):
//
//	"Whenever you cast a colorless creature spell, manifest the top card of
//	 your library.
//	 Whenever you attack with creatures with total power 6 or greater, you
//	 may turn a face-down creature you control face up."
//
// The cast trigger reads the spell on the stack: a colorless creature spell
// (an artifact creature, an Eldrazi). The attack trigger is one per
// declaration (OncePerBatch), gated on the total power of the creatures its
// controller declared (CR 208.3). The turn-up targets nothing: the controller
// picks a face-down creature they control as it resolves, and the engine
// refuses a manifested noncreature card (CR 701.40b). Turning it up by an
// effect pays no cost (CR 708.8).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d6158860-ada8-4e39-8156-23dee2f68f79",
		Name:         "Ugin's Mastery",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.Cmdr2S37ColorlessCreatureSpellCastByYou,
				"Ugin's Mastery — manifest the top card of your library", effects.Do(effects.Manifest{N: 1})),
			effects.OncePerBatch(effects.On(game.EventAttack, effects.Cmdr2S37AttackedWithTotalPower(6),
				"Ugin's Mastery — you may turn a face-down creature you control face up",
				effects.Cmdr2S37OfferTurnFaceUp)),
		},
	})
}
