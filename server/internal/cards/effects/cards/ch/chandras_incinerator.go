package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chandra's Incinerator — Creature — Elemental {5}{R}, 6/6 (Core Set 2021,
// #5511):
//
//	"This spell costs {X} less to cast, where X is the total amount of
//	 noncombat damage dealt to your opponents this turn.
//	 Trample
//	 Whenever a source you control deals noncombat damage to an opponent,
//	 this creature deals that much damage to target creature or planeswalker
//	 that player controls."
//
// The cost reduction is read from the turn's noncombat damage record as the
// spell is priced (CR 601.2f): damage from any source to any of your
// opponents counts, and it lowers only the generic part of the cost. The
// trigger is Fear of Burning Alive's watch (one trigger per damage event):
// "that player" is the opponent the event names, the target is chosen from
// that player's creatures and planeswalkers as it goes on the stack, and the
// damage is the Incinerator's own, equal to the event's amount.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "515cee6a-a2e1-4eeb-8e1d-293724313b14",
		Name:            "Chandra's Incinerator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(
				func(q game.CostQuery) int {
					return effects.S82OldB106NoncombatDamageToOpponentsThisTurn(q.Game, q.Controller)
				},
				"This spell costs {X} less to cast, where X is the total amount of noncombat damage dealt to your opponents this turn."),
		},
		Triggered: []game.TriggeredAbility{{
			Watches:     []game.EventKind{game.EventDealDamage},
			AppliesTo:   effects.ASourceYouControlDealtNoncombatDamageToAnOpponent,
			TargetsFrom: effects.S82OldB106CreatureOrPlaneswalkerThatPlayerControls,
			Key:         "Chandra's Incinerator — it deals that much damage to target creature or planeswalker that player controls",
			Effect:      effects.B24DamageToTargetEqualToEvent,
		}},
	})
}
