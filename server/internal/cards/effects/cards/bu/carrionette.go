package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Carrionette — Creature — Skeleton {1}{B}, 1/1 (Tempest, S82 OLD b752, #5628):
//
//	"{2}{B}{B}: Exile this card and target creature unless that creature's
//	 controller pays {2}. Activate only if this card is in your graveyard."
//
// A graveyard activation (Firemane Angel's, Zones graveyard) that targets a
// creature. As it resolves the creature's controller, not the Carrionette's, is
// asked to pay {2} (Soul Charmer's PayUnless); if they pay nothing is exiled, and
// if they don't both are, each only if it is still where the ability found it:
// the card in its owner's graveyard as the same object, the creature on the
// battlefield (CR 608.2b, 400.7). A target that left before resolution makes the
// ability fizzle (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "24315f60-731a-40bb-85ba-eedba1b011ff",
		Name:         "Carrionette",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}{B}: Exile this card and target creature unless that creature's controller pays {2}. Activate only if this card is in your graveyard.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.ManaCost("{2}{B}{B}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				c, ok := g.LookupCardForEffect(id)
				if !ok {
					return nil
				}
				return effects.PayUnless{
					Chooser:  c.Controller,
					Cost:     "{2}",
					Question: "Carrionette — pay {2}, or the creature and Carrionette are exiled?",
					Else:     carrionetteExile,
					Carry:    []uuid.UUID{id},
				}.Apply(ctx)
			},
		}},
	})
}

// carrionetteExile is the decline: exile Carrionette from the graveyard and the
// target creature from the battlefield, each only if it is still there. Keyed, an
// on-disk identity.
var carrionetteExile = effects.PayUnlessElse("pay-unless/s82-old-b752-carrionette-exile",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		this := ctx.Source()
		if z := ctx.Game.FindCardZoneForEffect(this); z != nil && z.Kind == game.ZoneGraveyard && !ctx.IsNewSourceObject(this) {
			if err := (effects.ExileTarget{Target: this}).Apply(ctx); err != nil {
				return err
			}
		}
		for _, id := range r.Carry {
			if effects.OnBattlefield(ctx.Game, id) {
				if err := (effects.ExileTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
			}
		}
		return nil
	})
