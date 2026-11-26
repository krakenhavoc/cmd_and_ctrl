package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reincarnation — Instant {1}{G}{G} (Commander 2013, S82 OLD b386):
//
//	"Choose target creature. When that creature dies this turn, return a
//	 creature card from its owner's graveyard to the battlefield under the
//	 control of that creature's owner."
//
// Rushed Rebirth's shape: the spell creates a delayed triggered ability about
// the chosen creature as the object it is now (CR 603.7a, 603.7c), controlled
// by the caster. As it resolves the CASTER chooses the creature card (it is
// not targeted), from the dead creature's owner's graveyard, the creature
// itself a legal choice; the card enters under that owner's control. With no
// creature card there, nothing happens. A token's owner is read as its last
// controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d6bf5e22-8d33-43a9-8824-435068e0a87a",
		Name:         "Reincarnation",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Answers: game.AnswerSurvive},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			creature, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.WhenThatCreatureDiesThisTurn(ctx, creature,
				"Reincarnation — return a creature card from its owner's graveyard to the battlefield under that owner's control",
				reincarnationDiesBody)
		},
	})
}

// reincarnationDiesBody is the fired trigger's body: the caster picks a creature
// card from the dead creature's owner's graveyard. An on-disk key.
var reincarnationDiesBody = game.DelayedBody("old-b386/reincarnation-return",
	func(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
		ctx := effects.NewContext(g, item)
		obj, ok := effects.PayloadObject(ctx)
		if !ok {
			return nil
		}
		owner := ctx.Controller()
		if info, ok := g.PermanentForEffect(obj); ok {
			owner = info.Controller
		}
		if c, ok := g.LookupCardForEffect(obj.ID); ok {
			owner = c.Owner
		}
		ids := effects.GraveyardCardIDs(ctx, owner, func(c game.Card) bool { return c.IsCreature() })
		if len(ids) == 0 {
			return nil
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    ctx.Controller(),
			FromPlayer: owner,
			Source:     item.SourceCardID,
			Question:   "Reincarnation — choose a creature card to return to the battlefield under its owner's control",
			Cards:      ids,
			Min:        1,
			Max:        1,
			LiveZone:   game.ZoneGraveyard,
			ThenKey:    reincarnationReturn,
			Asked:      game.NewAskedItem(item),
		})
		return nil
	})

// reincarnationReturn puts the chosen card onto the battlefield under its
// owner's control. An on-disk key.
var reincarnationReturn = effects.AskedPick("choose-cards/old-b386-reincarnation-return", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) != 1 {
		return nil
	}
	return effects.ReturnFromGraveyard{Target: r.Picked[0], Dest: game.ZoneBattlefield}.Apply(ctx)
})
