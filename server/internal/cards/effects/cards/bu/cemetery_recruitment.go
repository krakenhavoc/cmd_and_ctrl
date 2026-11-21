package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cemetery Recruitment — Sorcery {1}{B} (Foundations, #4465):
//
//	"Return target creature card from your graveyard to your hand. If
//	 it's a Zombie card, draw a card."
//
// Raise Dead's target and return, then the draw when the card is a
// Zombie. "Zombie" is read through the spell's text (CR 612.2). A return
// from a graveyard to a hand never waits on a prompt, so the draw
// follows it. A target that is gone by resolution makes the spell do
// nothing, the draw included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "07792d72-d5db-41b9-871c-e720e12ce659",
		Name:         "Cemetery Recruitment",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()),
		OnResolve:    cemeteryRecruitmentResolve,
	})
}

func cemeteryRecruitmentResolve(item *game.StackItem, ctx *effects.Context) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	zombie := false
	if c, ok := ctx.Game.LookupCardForEffect(item.Targets[0].ID); ok {
		for _, sub := range ctx.TextSubtypes("Zombie") {
			zombie = zombie || c.HasSubtype(sub)
		}
	}
	if err := effects.ReturnTargetFromGraveyardToHand(item, ctx); err != nil {
		return err
	}
	if !zombie {
		return nil
	}
	return effects.DrawCards{N: 1}.Apply(ctx)
}
