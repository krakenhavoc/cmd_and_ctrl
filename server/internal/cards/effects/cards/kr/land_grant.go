package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Land Grant — Sorcery {1}{G} (Mercadian Masques, S82 OLD, #5619):
//
//	"If you have no land cards in hand, you may reveal your hand rather
//	 than pay this spell's mana cost.
//	 Search your library for a Forest card, reveal that card, put it into
//	 your hand, then shuffle."
//
// The alternative cost is #6214's RevealHand (CR 118.9): showing the hand is the
// whole payment, offered only while the caster holds no land card (checked with the
// spell on the stack, so Land Grant itself is not in the hand). The search is
// Hunting Cheetah's, for any card with the Forest subtype.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e36f4649-934d-4e8a-bd5f-39f164ea6937",
		Name:         "Land Grant",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		AlternativeCosts: []game.AlternativeCost{{
			Key:        "land-grant",
			Label:      "If you have no land cards in hand, reveal your hand rather than pay this spell's mana cost",
			RevealHand: true,
			Condition:  landGrantNoLandsInHand,
		}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: effects.IsLandWithSubtype("forest"),
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Land Grant — a Forest card",
			}.Apply(ctx)
		},
	})
}

// landGrantNoLandsInHand is "if you have no land cards in hand".
func landGrantNoLandsInHand(g *game.Game, controller uuid.UUID) bool {
	p := g.PlayerByIDForEffect(controller)
	if p == nil {
		return false
	}
	for _, c := range p.Hand.Cards {
		if c.IsLand() {
			return false
		}
	}
	return true
}
