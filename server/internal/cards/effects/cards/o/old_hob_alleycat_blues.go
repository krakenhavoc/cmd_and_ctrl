package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Old Hob, Alleycat Blues — {4}{R} Legendary Creature — Cat Mutant Rebel 4/4
// (Teenage Mutant Ninja Turtles, #4457):
//
//	"At the beginning of combat on your turn, create a 2/2 red Mutant
//	 creature token. It gains haste until end of turn. Destroy it at the
//	 beginning of the next end step.
//	 {1}{W}: Target attacking creature token gains indestructible until
//	 end of turn."
//
// The token is made with a keyed tail told which tokens really landed
// (Artistic Process's haste): each gains haste until end of turn and gets a
// delayed trigger (CR 603.7a) that destroys it at the beginning of the next
// end step, Glyph of Destruction's body, only while it is still that object
// (CR 400.7). Destroy, not sacrifice: an indestructible token survives,
// which is what the activated ability is for.
//
// No simplification.
func init() {
	const label = "Old Hob — create a 2/2 red Mutant; it gains haste; destroy it at the next end step"
	effects.Register(effects.Spec{
		OracleID:     "e88b1520-aa9a-41ee-9036-4d4bebd1a0c0",
		Name:         "Old Hob, Alleycat Blues",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfYourCombat(label, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return g.CreateTokensKeyedForEffect(game.TokenCreation{
					Controller: item.Controller,
					Source:     item.SourceCardID,
					Groups:     []game.TokenGroup{{Template: ctx.TextToken(effects.TokenCard("2/2 red Mutant")), Count: 1}},
				}, ctx.ItemTail(oldHobHasteAndDestroy))
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}: Target attacking creature token gains indestructible until end of turn.",
			Cost:    effects.ManaCost("{1}{W}"),
			Targets: effects.TargetCreature("target attacking creature token", effects.AttackingCreature(), effects.IsTokenPredicate()),
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "indestructible")),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("indestructible", "Old Hob — indestructible"),
		}},
	})
}

// oldHobHasteAndDestroy is "It gains haste until end of turn. Destroy it at
// the beginning of the next end step." over the tokens the creation made.
// An on-disk key: never renamed, never reused.
var oldHobHasteAndDestroy = effects.CardTail("old-hob/token-haste-and-destroy", func(ctx *effects.Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || !effects.OnBattlefield(ctx.Game, id) {
			continue
		}
		if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Old Hob — the Mutant gains haste"}).Apply(ctx); err != nil {
			return err
		}
		if err := (effects.ScheduleDelayedTrigger{
			At:         game.StepEnd,
			Controller: ctx.Item.Controller,
			Label:      "Old Hob — destroy the Mutant",
			Cards:      []uuid.UUID{id},
			Body:       effects.DestroyTheObjectBody,
			Params:     game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
