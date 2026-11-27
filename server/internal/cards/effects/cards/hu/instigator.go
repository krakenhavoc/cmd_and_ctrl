package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Instigator — Creature — Human Spellshaper {1}{B}, 1/1 (Mercadian Masques,
// S82 OLD b681, #5619):
//
//	"{1}{B}{B}, {T}, Discard a card: Creatures target player controls attack
//	 this turn if able."
//
// Incite War's first bullet: a CR 508.1d attack requirement scoped to the
// creatures the target player controls (ScopeYourCreatures with that player
// as its "you"), until end of turn. It is a requirement, not a locked set
// (CR 611.2c), so a creature they gain this turn must attack too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0fe1a1f7-41c9-415e-a5e3-b5ba0c0271e5",
		Name:         "Instigator",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}{B}, {T}, Discard a card: Creatures target player controls attack this turn if able.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.Plus(effects.ManaCost("{1}{B}{B}"), effects.TapCost(), effects.DiscardACard()),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 || ts[0].Kind != game.TargetPlayer {
					return nil
				}
				g.RegisterScopedRuleEffectForEffect(ctx.Source(), game.ScopeYourCreatures, ts[0].ID,
					[]game.Mod{game.AddAttackRequirementMod(uuid.Nil)}, effects.DurationUntilEndOfTurn(ctx),
					"Instigator — creatures that player controls attack this turn if able")
				return nil
			},
		}},
	})
}
