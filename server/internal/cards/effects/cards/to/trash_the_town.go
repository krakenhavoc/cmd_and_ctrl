package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trash the Town — Instant {G} (Outlaws of Thunder Junction, #4468):
//
//	"Spree (Choose one or more additional costs.)
//	 + {2} — Put two +1/+1 counters on target creature.
//	 + {1} — Target creature gains trample until end of turn.
//	 + {1} — Until end of turn, target creature gains "Whenever this
//	   creature deals combat damage to a player, draw two cards.""
//
// Three Spree bullets (CR 702.172a), each with its own target, re-checked
// against its own clause as it resolves (ADR 0065 §2), so the bullets may
// name the same creature or three different ones. The third is a duration
// grant (ADR 0093, Blitzball Stadium's Go for the Goal! shape): the trigger
// is the creature's own ability until end of turn, controlled by that
// creature's controller, who draws the two cards.
//
// No simplification.
const trashTheTownGrant = "trash-the-town/combat-damage-draw-two"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "b47f7f1e-2d81-4b1b-870d-4c1a4f020505",
		Name:         "Trash the Town",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: trashTheTownGrant,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisDealsCombatDamageToAPlayer("Trash the Town — draw two cards",
					func(g *game.Game, item *game.StackItem) error {
						return effects.DrawCards{N: 2}.Apply(effects.NewContext(g, item))
					}),
			},
			Text: "Whenever this creature deals combat damage to a player, draw two cards.",
		}},
		Modes: effects.Spree(
			effects.SpreeModeDoing("Put two +1/+1 counters on target creature.", "{2}",
				effects.TargetCreature("target creature"),
				effects.TwoPlusOneCountersOnModeTarget),
			effects.ModeWithPurpose(effects.SpreeModeDoing("Target creature gains trample until end of turn.", "{1}",
				effects.TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"trample"}, Label: "Trash the Town — trample"}.Apply(ctx)
				}), effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "trample"))),
			effects.ModeWithPurpose(effects.SpreeModeDoing("Until end of turn, target creature gains \"Whenever this creature deals combat damage to a player, draw two cards.\"", "{1}",
				effects.TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return effects.GrantAbilitiesFor{
						Target: t.ID,
						Keys:   []string{trashTheTownGrant},
						Label:  "Trash the Town — draws two cards when it deals combat damage to a player",
					}.Apply(ctx)
				}), effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0))),
		),
	})
}
