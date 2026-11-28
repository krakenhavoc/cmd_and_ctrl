package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soldevi Golem — Artifact Creature — Golem {4}, 5/3 (Ice Age, S82 OLD batch 847, #5636):
//
//	"This creature doesn't untap during your untap step.
//	 At the beginning of your upkeep, you may untap target tapped creature an
//	 opponent controls. If you do, untap this creature."
//
// Brass Man's untap-step restriction. The upkeep trigger targets a tapped
// creature an opponent controls as it goes on the stack (no such creature, no
// trigger, CR 603.3d); as it resolves its controller may untap it, and if they
// do the Golem untaps too. A target that is no longer tapped and an opponent's
// creature is still a legal one (CR 608.2b); "if you do" is the untap having
// happened, so a target that has gone, or that was not tapped any more, does
// not untap the Golem.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "cd444267-712c-450f-a349-e27faf4f4c7c",
		Name:                  "Soldevi Golem",
		Completeness:          effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringYourUntapStep()},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.AtYourUpkeep("Soldevi Golem — you may untap target tapped creature an opponent controls; if you do, untap this creature",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: "Soldevi Golem — untap the target creature and untap this creature?",
						Then:     oldB847GolemUntap,
					}.Apply(effects.NewContext(g, item))
				}), effects.TargetCreature("target tapped creature an opponent controls", effects.TappedPermanent(), effects.OpponentControls())),
		},
	})
}

// oldB847GolemUntap is the yes branch: untap the target if it is still a legal
// one and tapped, and only then untap the Golem. An on-disk key: never renamed,
// never reused.
var oldB847GolemUntap = effects.MayYes("confirm/old-b847-soldevi-golem-untap", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := ctx.Game.LookupCardForEffect(t.ID)
		if !ok || !c.Tapped {
			return nil
		}
		if err := (effects.UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
		return effects.UntapTarget{Target: ctx.Source()}.Apply(ctx)
	}
	return nil
})
