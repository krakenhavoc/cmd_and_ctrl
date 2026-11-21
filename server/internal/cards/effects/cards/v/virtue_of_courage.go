package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Virtue of Courage // Embereth Blaze — Enchantment {3}{R}{R} // Instant —
// Adventure {1}{R} (Wilds of Eldraine, #4471):
//
//	Virtue of Courage — "Whenever a source you control deals noncombat
//	 damage to an opponent, you may exile that many cards from the top of
//	 your library. You may play those cards this turn."
//	Embereth Blaze — "Embereth Blaze deals 2 damage to any target. (Then
//	 exile this card. You may cast the enchantment later from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"), as Virtue of Persistence does. The enchantment is Thor,
// Guardian of Midgard's trigger without the body (the shared
// exileThatManyYouMayPlayThisTurn): one trigger per damage
// event, "a source you control" read off the event (a permanent, or the
// spell still on the stack), that many cards exiled with Breeches'
// play-this-turn window, the exile asked as the trigger resolves.
// Embereth Blaze is a plain 2 damage to any target. The Adventure
// lifecycle is the engine's (CR 715.3d).
//
// No simplification.
const VirtueOfCourageOracleID = "af6438b9-3fde-45c9-af69-dff686c1e0de"

func init() {
	effects.Register(effects.Spec{
		OracleID:     VirtueOfCourageOracleID,
		Name:         "Virtue of Courage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventDealDamage, effects.ASourceYouControlDealtNoncombatDamageToAnOpponent,
				"Virtue of Courage — exile that many cards from the top of your library; you may play them this turn",
				effects.ExileThatManyYouMayPlayThisTurn), "Exile that many cards from the top of your library? You may play them this turn."),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     VirtueOfCourageOracleID + "#1",
		Name:         "Embereth Blaze",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 2)),
		OnResolve:    effects.DamageToFirstTargetStep(2),
	})
}
