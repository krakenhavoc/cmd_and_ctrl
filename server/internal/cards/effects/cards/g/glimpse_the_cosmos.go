package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Glimpse the Cosmos — Sorcery {1}{U} (Kaldheim):
//
//	"Look at the top three cards of your library. Put one of them into your
//	 hand and the rest on the bottom of your library in any order.
//	 As long as you control a Giant, you may cast this card from your
//	 graveyard by paying {U} rather than paying its mana cost. If you cast
//	 this card this way and it would be put into your graveyard, exile it
//	 instead."
//
// Impulse's look with a mandatory take. The graveyard cast is open only while
// its owner controls a Giant (CastableZonesIf, read by the engine, the view
// and the enumerator alike), for {U} as an alternative cost (CR 118.9); a
// spell cast that way is exiled instead of going to the graveyard
// (ExileOnLeavingStack), as flashback's is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "27065d34-b22a-47af-aea2-980f47bafcea",
		Name:         "Glimpse the Cosmos",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1, Answers: game.AnswerValue},
		CastableZonesIf: []game.ConditionalCastZone{{
			Zone: game.ZoneGraveyard,
			Cond: func(g *game.Game, owner uuid.UUID, _ game.Card) bool {
				return len(effects.PermanentsControlledByMatching(g, owner, effects.HasSubtype("Giant"))) > 0
			},
		}},
		AlternativeCosts: []game.AlternativeCost{{
			Key:                 "glimpse-the-cosmos",
			Label:               "Pay {U} (you control a Giant)",
			ManaCost:            "{U}",
			FromZone:            game.ZoneGraveyard,
			ExileOnLeavingStack: true,
		}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.LookAtTopThenTakeOneRestOnBottom(3, "Glimpse the Cosmos — put one of those cards into your hand")(ctx.Game, item)
		},
	})
}
