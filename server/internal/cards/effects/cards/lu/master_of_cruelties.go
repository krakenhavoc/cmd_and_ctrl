package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Master of Cruelties — Creature — Demon {3}{B}{R}, 1/4:
//
//	"First strike, deathtouch
//	 This creature can only attack alone.
//	 Whenever this creature attacks a player and isn't blocked, that
//	 player's life total becomes 1. This creature assigns no combat
//	 damage this combat."
//
// #3531, ADR 0187 amendment 2026-10-16. "Can only attack alone" is a
// restriction on the creature (CR 508.1c, 506.5): an attack limit of its
// own (CanOnlyAttackAlone), so both declaration verbs refuse it beside
// another attacker, the legal-move enumerator never offers it, and the
// CR 508.1d requirement count weighs it attacking alone against the
// others attacking without it.
//
// The trigger is CR 509.3g's "attacks and isn't blocked", narrowed to an
// attack on the player (not a planeswalker or battle they defend). The
// life total becoming 1 is a life gain or loss of the difference
// (CR 119.5). "This combat" is the combat the trigger resolves in, so in
// an added combat phase Master deals its damage again
// (AssignsNoCombatDamageThisCombatForEffect).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d135c457-9be5-4a6e-a6fb-5b35cfb70cb0",
		Name:            "Master of Cruelties",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "deathtouch"},
		AttackLimits:    []game.AttackLimit{effects.CanOnlyAttackAlone()},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBlockersDeclared},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttacksAndIsNotBlocked(ev, source, g) && source.AttackingTarget == ev.Actor
			},
			Key: "Master of Cruelties — that player's life total becomes 1",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return masterOfCruelties(ctx, effects.TriggeringActor(item))
			},
		}},
	})
}

func masterOfCruelties(ctx *effects.Context, defender uuid.UUID) error {
	if effects.DefendingPlayerStillIn(ctx.Game, defender) {
		if err := effects.B31LifeBecomes(ctx, defender, 1); err != nil {
			return err
		}
	}
	if !ctx.IsNewSourceObjectAsThis(ctx.Source()) {
		ctx.Game.AssignsNoCombatDamageThisCombatForEffect(ctx.Source(), ctx.Source(),
			"Master of Cruelties — assigns no combat damage this combat")
	}
	return nil
}
