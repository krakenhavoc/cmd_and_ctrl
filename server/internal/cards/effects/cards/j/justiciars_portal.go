package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Justiciar's Portal — Instant {1}{W} (Jumpstart 2022, #5948):
//
//	"Exile target creature you control, then return that card to the
//	 battlefield under its owner's control. It gains first strike until
//	 end of turn."
//
// Essence Flux's exile-then-return with a tail on the entry: the creature
// comes back as a new object (CR 400.7), and the tail grants first strike
// to that new object, not to the one that left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0214805d-d207-42da-a25f-2a8e1990904e",
		Name:         "Justiciar's Portal",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			target := item.Targets[0].ID
			return effects.ExileTarget{
				Target:   target,
				ThenTail: effects.WithCarry(ctx.ItemTail(justiciarsPortalExiled), target),
			}.Apply(ctx)
		},
	})
}

var justiciarsPortalFirstStrike = effects.EnteredTail("justiciars-portal/first-strike", func(ctx *effects.Context, entered uuid.UUID, _ game.TailOutcome) error {
	if entered == uuid.Nil {
		return nil
	}
	return effects.GrantKeywordUntilEOT{Target: entered, Keywords: []string{"first strike"}, Label: "Justiciar's Portal — first strike"}.Apply(ctx)
})

var justiciarsPortalExiled = effects.CardTail("justiciars-portal/after-exile", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) != 1 {
		return nil
	}
	return effects.ReturnFromExile{
		Target:   effects.TailCarry0(o),
		ThenTail: ctx.ItemTail(justiciarsPortalFirstStrike),
	}.Apply(ctx)
})
