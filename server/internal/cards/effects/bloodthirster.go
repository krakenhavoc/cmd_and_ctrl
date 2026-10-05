package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bloodthirster — Creature — Demon {5}{R}, 6/6:
//
//	"Flying, trample
//	 Whenever this creature deals combat damage to a player, untap it.
//	 After this phase, there is an additional combat phase.
//	 This creature can't attack a player it has already attacked this
//	 turn."
//
// The restriction is a CR 508.1c clause on game.AttackTargetRestriction
// (NotAlreadyAttackedThisTurn, #2171) that reads TurnTally.Attacks for
// this object: a flickered Bloodthirster is a new object (CR 400.7) and
// may attack again. Planeswalkers and battles are not players and are
// never refused by it. The attack validator, the enumerator's
// per-attacker target list and the CR 508.1d search all go through the
// same predicate.
//
// No simplification.
func init() {
	const label = "Bloodthirster — untap it, additional combat phase"
	Register(Spec{
		OracleID:        "e971249a-64a3-4a9b-9a0c-e9d858ca8a55",
		Name:            "Bloodthirster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Static:          []game.StaticAbility{CantAttackAPlayerItAlreadyAttackedThisTurn()},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Source == source.InstanceID && combatDamageToPlayerBy(ev, source.Controller, g)
			}, label, bloodthirsterUntapAndCombat),
		},
	})
}

// bloodthirsterUntapAndCombat untaps the source, then adds the combat.
func bloodthirsterUntapAndCombat(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := untapEach(ctx, []uuid.UUID{item.SourceCardID}); err != nil {
		return err
	}
	return ExtraCombatAfterThisPhase().Apply(ctx)
}
