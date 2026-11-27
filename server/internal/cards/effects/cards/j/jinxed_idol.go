package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jinxed Idol — Artifact {2} (Magic 2011, #5579):
//
//	"At the beginning of your upkeep, this artifact deals 2 damage to you.
//	 Sacrifice a creature: Target opponent gains control of this artifact."
//
// The upkeep trigger is mandatory and the Idol is the damage source, so its
// controller's damage prevention and lifelink-style readers see an artifact.
// The activated ability's cost is sacrificing any creature you control, and
// the control change is indefinite (CR 611.2a), applied to the Idol only
// while it is still the permanent the ability came from (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "89b1eab9-4d24-4367-99f7-263349f69c3a",
		Name:         "Jinxed Idol",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Jinxed Idol — 2 damage to you", S82OldB475IdolDamagesItsController),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a creature: Target opponent gains control of this artifact.",
			Cost:    effects.SacrificeACreature(),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect:  S82OldB475IdolChangesHands,
		}},
	})
}

// S82OldB475IdolDamagesItsController is "this artifact deals 2 damage to
// you": the Idol's controller as the trigger resolves.
func S82OldB475IdolDamagesItsController(g *game.Game, item *game.StackItem) error {
	return effects.DealDamage{Source: item.SourceCardID, Target: item.Controller, Amount: 2}.Apply(effects.NewContext(g, item))
}

// S82OldB475IdolChangesHands is "target opponent gains control of this
// artifact", indefinitely, if the Idol is still the same permanent.
func S82OldB475IdolChangesHands(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 || !effects.SourceIsStillThisPermanent(g, item) {
		return nil
	}
	return effects.GainControl{
		Target:     item.SourceCardID,
		Controller: targets[0].ID,
		Duration:   game.IndefiniteDuration(),
		Label:      "Jinxed Idol — target opponent gains control of it",
	}.Apply(ctx)
}
