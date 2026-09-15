package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Keeper of Fables — Creature — Cat {3}{G}{G}, 4/5 (EDHREC rank
// 3210):
//
//	"Whenever one or more non-Human creatures you control deal
//	 combat damage to a player, draw a card."
//
// The non-Human beatdown deck's Bident. "ONE OR MORE" is one trigger
// per combat damage step, not one per creature: the engine emits one
// damage event per creature, so the condition declines any event
// that arrives while a Keeper trigger is already queued or on the
// stack (OncePerBatch — both queues, since a damage
// batch can straddle a state check). Without it three attackers
// would draw three, which is stronger than printed. First-strike and
// regular damage are two batches and two draws, as in paper. Human
// is read off the creature's effective subtypes, so a changeling is
// a Human and does not count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c8ca3116-e0f0-4e27-aa0f-99ed85927040",
		Name:         "Keeper of Fables",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b30NonHumanCreatureYouControlDealtCombatDamageToPlayer(ev, source, g)
			}, "Keeper of Fables — draw a card", Do(DrawCards{N: 1}))),
		},
	})
}
