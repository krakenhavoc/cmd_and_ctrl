package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Druidic Ritual — Sorcery {2}{G} (Commander Legends: Battle for Baldur's
// Gate, S82 GAP #5948):
//
//	"You may mill three cards. Then return up to one creature card and up to
//	 one land card from your graveyard to your hand."
//
// Another Chance's mill question (asked only when the library holds three
// cards), then two independent "up to one" picks from the graveyard as it is
// now, milled cards included: a creature card, then a land card (a card that
// is both may be taken as either, not both). Each is chosen, not targeted.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "79edf527-2de1-433a-9cac-c91f461b1ebf",
		Name:         "Druidic Ritual",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if p := ctx.Game.PlayerByIDForEffect(ctx.Controller()); p != nil && p.Library != nil && p.Library.Size() >= 3 {
				return effects.MayChoice{
					Question: "Druidic Ritual — mill three cards?",
					Then:     druidicRitualMill,
				}.Apply(ctx)
			}
			return druidicRitualReturn(ctx)
		},
	})
}

// druidicRitualMill is the answer to the mill question: mill on a yes, then the
// returns either way. An on-disk key.
var druidicRitualMill = effects.MayAnswer("confirm/druidic-ritual-mill", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	if r.Accepted {
		if err := (effects.MillCards{N: 3}).Apply(ctx); err != nil {
			return err
		}
	}
	return druidicRitualReturn(ctx)
})

func druidicRitualReturn(ctx *effects.Context) error {
	if err := (effects.ReturnChosenFromGraveyard{
		Player:   ctx.Controller(),
		Question: "Druidic Ritual — return up to one creature card from your graveyard to your hand",
		Match:    func(c game.Card) bool { return c.IsCreature() },
		Max:      1,
		ToHand:   true,
	}).Apply(ctx); err != nil {
		return err
	}
	return effects.ReturnChosenFromGraveyard{
		Player:   ctx.Controller(),
		Question: "Druidic Ritual — return up to one land card from your graveyard to your hand",
		Match:    func(c game.Card) bool { return c.IsLand() },
		Max:      1,
		ToHand:   true,
	}.Apply(ctx)
}
