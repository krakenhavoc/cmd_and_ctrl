package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hiredClawGrowLabel is the counter ability's label, which its
// once-each-turn condition reads.
const hiredClawGrowLabel = "{1}{R}: Put a +1/+1 counter on this creature. Activate only if an opponent lost life this turn and only once each turn."

// Hired Claw — Creature — Lizard Mercenary {R}, 1/2 (Bloomburrow, #4467):
//
//	"Whenever you attack with one or more Lizards, this creature deals 1
//	 damage to target opponent.
//	 {1}{R}: Put a +1/+1 counter on this creature. Activate only if an
//	 opponent lost life this turn and only once each turn."
//
// The attack trigger is Sidar Jabari's "attack with a Knight" over the
// Lizards (one trigger per attack declaration, OncePerBatch, whatever the
// number of Lizards; Hired Claw itself is one) with Deadeye Duelist's
// damage to target opponent. "Lizards" is a word of the Claw's text (CR
// 612.2). The counter ability is gated on Rakdos, Lord of Riots' read of
// the opponents' life lost this turn and on the activation tally
// (OncePerTurnActivation: an announcement uses up the turn, CR 602.1b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "031cfc7d-8bda-454b-b88a-15d3f5fe47e2",
		Name:         "Hired Claw",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(Targeting(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller) && attackerHasSubtype(g, ev, source.TextSubtype("Lizard"))
			}, "Hired Claw — 1 damage to target opponent", sourceDealsDamageToEachLegalTarget(1)),
				TargetPlayer("target opponent", Opponent()))),
		},
		Activated: []ActivatedAbility{{
			Label:   hiredClawGrowLabel,
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{1}{R}"),
			Condition: AllConditions(func(g *game.Game, controller, _ uuid.UUID) bool {
				return opponentLifeLostThisTurn(g, controller) > 0
			}, OncePerTurnActivation(hiredClawGrowLabel)),
			Effect: Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
		}},
	})
}
