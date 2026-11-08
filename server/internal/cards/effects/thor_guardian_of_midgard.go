package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thor, Guardian of Midgard — Legendary Creature — God Warrior Hero
// {3}{R}{R}, 5/5 (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 Whenever a source you control deals noncombat damage to an opponent,
//	 you may exile that many cards from the top of your library. You may
//	 play those cards this turn."
//
// One trigger per damage event, so two sources are two triggers. "A
// source you control" is read off the damage event (Daredevil's
// b34DamageSourceController: a permanent or the spell still on the stack),
// the amount is the event's, and the exile is the impulse window of
// Breeches ("play", so a land counts; the grant's zero duration is this
// turn). The exile is optional, asked as the trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "74acdcbc-5185-422a-a236-0a126edd2eee",
		Name:            "Thor, Guardian of Midgard",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventDealDamage, ASourceYouControlDealtNoncombatDamageToAnOpponent,
				"Thor, Guardian of Midgard — exile that many cards from the top of your library; you may play them this turn",
				exileThatManyYouMayPlayThisTurn), "Exile that many cards from the top of your library? You may play them this turn."),
		},
	})
}
