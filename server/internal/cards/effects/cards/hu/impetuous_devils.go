package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Impetuous Devils — Creature — Devil {2}{R}{R}, 6/1 (Eldritch Moon, S82
// OLD, #5541):
//
//	"Trample, haste
//	 When this creature attacks, up to one target creature defending player
//	 controls blocks it this combat if able.
//	 At the beginning of the end step, sacrifice this creature."
//
// Fighter Class's forced block with the target read off the Devils' own
// attack (Warkite Marauder's defending-player clause, with its floor at
// zero). The sacrifice is Blistering Firecat's, at every end step.
//
// No simplification.
func init() {
	const label = "Impetuous Devils — up to one target creature defending player controls blocks it this combat if able"
	attacks := effects.WheneverThisAttacks(label, func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		blocker, ok := effects.B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		ref, ok := g.PermanentRefForEffect(item.SourceCardID)
		if !ok || !effects.OnBattlefield(g, item.SourceCardID) {
			return nil
		}
		return effects.BlocksAttackerThisCombat{Blocker: blocker, Attacker: ref,
			Label: "Impetuous Devils — blocks it this combat if able"}.Apply(ctx)
	})
	attacks.TargetsFrom = effects.S82OldB277UpToOneCreatureDefendingPlayerControls
	effects.Register(effects.Spec{
		OracleID:        "01add64a-205d-4c0e-a393-89e7a1a0ea44",
		Name:            "Impetuous Devils",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Triggered: []game.TriggeredAbility{
			attacks,
			effects.AtEachStep(game.StepEnd, "Impetuous Devils — sacrifice it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SacrificePermanent{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
