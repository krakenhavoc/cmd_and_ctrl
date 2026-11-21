package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Desertion — Instant {3}{U}{U} (Visions; Special Guests, #5931):
//
//	"Counter target spell. If an artifact or creature spell is countered
//	 this way, put that card onto the battlefield under your control
//	 instead of into its owner's graveyard."
//
// A destination counter (Remand's shape, CR 614.1a "instead"), with the
// battlefield as the destination when the spell is an artifact or a
// creature spell as Desertion resolves: the engine's countered-onto-the-
// battlefield entry (game/counter_onto_battlefield.go), which runs the
// CR 614 entry window under Desertion's controller and fires the card's
// ETB. Any other spell is countered into its owner's graveyard. A spell
// that can't be countered is neither countered nor moved, and a copy of a
// spell is not a card and ceases to exist (CR 707.10a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ee3d1f44-e0ca-4ce9-be76-4b675a115156",
		Name:         "Desertion",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			counter := effects.CounterTarget{StackID: t.ID}
			if spell, found := ctx.Game.LookupCardForEffect(t.ID); found && (spell.IsArtifact() || spell.IsCreature()) {
				counter.Dest = game.ZoneRef{Kind: game.ZoneBattlefield, Owner: ctx.Controller()}
			}
			return counter.Apply(ctx)
		},
	})
}
