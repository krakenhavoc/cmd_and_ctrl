package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Imaginary Threats — Instant {2}{U}{U} (Hour of Devastation, #5534):
//
//	"Creatures target opponent controls attack this turn if able. During that
//	 player's next untap step, creatures they control don't untap.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Incite War's attack requirement (a rule-scoped record over that player's
// creatures, read live, so a creature that arrives later is bound too, CR
// 508.1d) and Misstep's untap lock over their creatures (the next untap step
// of that player: the creatures it holds as the spell resolves, CR 611.2c).
// A player who has left the game makes the target illegal. Cycling is the
// shared activated ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b1bc4498-43c0-461c-afba-ef81d743e5aa",
		Name:         "Imaginary Threats",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetPlayer {
						continue
					}
					ctx.Game.RegisterScopedRuleEffectForEffect(ctx.Source(), game.ScopeYourCreatures, t.ID,
						[]game.Mod{game.AddAttackRequirementMod(uuid.Nil)}, effects.DurationUntilEndOfTurn(ctx),
						"Imaginary Threats — creatures that player controls attack this turn if able")
				}
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.TargetPlayersPermanentsDontUntap(ctx, "Imaginary Threats — creatures don't untap during that player's next untap step", "creature")
			}),
		),
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
