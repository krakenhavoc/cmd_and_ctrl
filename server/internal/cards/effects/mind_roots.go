package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mind Roots — Sorcery {1}{B}{G} (Secrets of Strixhaven, #4456):
//
//	"Target player discards two cards. Put up to one land card discarded
//	 this way onto the battlefield tapped under your control."
//
// The target chooses the two cards from their own hand (CR 701.9b; a hand
// of one or none discards what it has). The land clause reads what really
// reached a graveyard that way, not what the player held, so a land that a
// replacement sent elsewhere (or one already gone again) is not offered; the
// caster, not the target, picks "up to one" of the lands (a real zero is a
// legal answer), and it enters tapped under the caster's control from the
// graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5d32dbf3-c157-4543-8813-da88f463de0e",
		Name:         "Mind Roots",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			victim := TargetedPlayer(ctx)
			if victim == uuid.Nil {
				return nil
			}
			return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
				Player:   victim,
				Source:   item.SourceCardID,
				N:        2,
				Question: "Mind Roots — discard two cards",
				Asked:    game.NewAskedItem(item),
			}, mindRootsLands, nil)
		},
	})
}

// mindRootsLands is the clause after the discard: the caster may put one of
// the land cards discarded this way onto the battlefield tapped.
var mindRootsLands = game.RegisterDiscardThen("choose-cards/mind-roots-lands", func(g *game.Game, r game.DiscardsDone) error {
	if r.Item == nil {
		return nil
	}
	var lands []uuid.UUID
	for _, id := range r.Discarded.Cards() {
		z := g.FindCardZoneForEffect(id)
		if z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if c, ok := g.LookupCardForEffect(id); ok && c.IsLand() {
			lands = append(lands, id)
		}
	}
	if len(lands) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    r.Item.Controller,
		FromPlayer: r.Player,
		Source:     r.Source,
		Question:   "Mind Roots — put up to one land card discarded this way onto the battlefield tapped under your control",
		Cards:      lands,
		Min:        0,
		Max:        1,
		LiveZone:   game.ZoneGraveyard,
		ThenKey:    mindRootsPut,
		Asked:      game.NewAskedItem(r.Item),
	})
	return nil
})

// mindRootsPut puts the chosen land onto the battlefield tapped under the
// caster's control.
var mindRootsPut = askedPick("choose-cards/mind-roots-put-land", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return ReturnFromGraveyard{Target: r.Picked[0], Dest: game.ZoneBattlefield, Controller: ctx.Controller(), Tapped: true}.Apply(ctx)
})
