package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Yuna's Decision — Sorcery {3}{G}:
//
//	"Choose one —
//	 • Continue the Pilgrimage — Sacrifice a creature. If you do, draw a
//	   card, then you may put a creature card and/or a land card from
//	   your hand onto the battlefield.
//	 • Find Another Way — Return one or two target permanent cards from
//	   your graveyard to your hand."
//
// The first mode's sacrifice is chosen as it resolves, through the one
// sacrifice prompt. With nothing sacrificed (no creature) nothing else
// happens. Otherwise the card is drawn first, so it can be put, and then
// one prompt offers up to two cards from hand that are a creature card
// and a land card, entering together (PutFromHandOntoBattlefield with a
// set rule); a card that is both counts as either.
//
// The second mode returns the picks still in the graveyard (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "77dd6ce0-91fc-4e71-99ff-47cdf3359067",
		Name:         "Yuna's Decision",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Continue the Pilgrimage — Sacrifice a creature. If you do, draw a card, then you may put a creature card and/or a land card from your hand onto the battlefield.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					return ctx.PlayersSacrificeThen([]uuid.UUID{item.Controller},
						sacrificeSpec("a creature", Creature()), "Yuna's Decision — sacrifice a creature",
						yunasDecisionSacrificed, ctx.RunData(nil))
				}),
			ModeDoing("Find Another Way — Return one or two target permanent cards from your graveyard to your hand.",
				TargetCardInGraveyard("one or two target permanent cards from your graveyard", Permanent(), YouOwn()).WithCount(1, 2),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					for _, t := range ctx.ModeTargets(occ) {
						if !ctx.IsTargetLegal(t) {
							continue
						}
						if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		),
	})
}

// yunasDecisionPutRef is the rest of the first mode, run once the card
// is drawn: the put from hand, asked of the player who drew.
var yunasDecisionPutRef = game.RegisterDrawThen("yunas-decision/put", func(g *game.Game, d game.DrawThen) error {
	return PutFromHandOntoBattlefield{
		Player:   d.Player,
		Match:    Or(Creature(), Land()),
		Optional: true,
		Max:      2,
		Validate: aCreatureAndALand,
		Label:    "Yuna's Decision — you may put a creature card and/or a land card from your hand onto the battlefield",
	}.Apply(NewContext(g, &game.StackItem{Controller: d.Player, SourceCardID: d.Source}))
})

// aCreatureAndALand is the set rule for "a creature card and/or a land
// card": at most one of each, a card that is both filling either slot.
func aCreatureAndALand(picked []game.Card) bool {
	switch len(picked) {
	case 0:
		return true
	case 1:
		return picked[0].IsCreature() || picked[0].IsLand()
	case 2:
		a, b := picked[0], picked[1]
		return (a.IsCreature() && b.IsLand()) || (a.IsLand() && b.IsCreature())
	}
	return false
}

// yunasDecisionSacrificed is the card drawn once the creature has really
// been sacrificed, keyed (#4290).
var yunasDecisionSacrificed = RunThen("prompt-run/yunas-decision-sacrificed", func(ctx *Context, r game.RunDone) error {
	if len(r.Sacrifices().Cards()) == 0 {
		return nil
	}
	return ctx.Game.DrawNThenForEffect(ctx.Controller(), 1, game.DrawThen{
		Ref: yunasDecisionPutRef, Player: ctx.Controller(), Source: ctx.Source(),
	})
})
