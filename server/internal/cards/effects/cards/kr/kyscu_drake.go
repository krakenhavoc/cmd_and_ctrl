package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kyscu Drake — Creature — Drake {3}{G}, 2/2 (Visions, S82 OLD batch 787,
// #5631):
//
//	"Flying
//	 {G}: This creature gets +0/+1 until end of turn. Activate only once each
//	 turn.
//	 Sacrifice this creature and a creature named Spitting Drake: Search your
//	 library for a card named Viashivan Dragon, put that card onto the
//	 battlefield, then shuffle."
//
// Flying is an engine keyword (CR 702.9b). The pump is a layer 7c change
// (CR 613.4c) limited to one activation each turn. The sacrifice ability's
// cost is the Drake itself and a second creature whose name is exactly
// Spitting Drake (CR 201.2; a different creature does not pay it); the search
// finds up to one card of that name, puts it onto the battlefield and
// shuffles.
//
// No simplification.
func init() {
	const pump = "{G}: This creature gets +0/+1 until end of turn. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:        "94806688-329d-429d-8126-2c9187e71f54",
		Name:            "Kyscu Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{
			{
				Label:     pump,
				Purpose:   game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Toughness: 1}},
				Cost:      effects.ManaCost("{G}"),
				Condition: effects.OncePerTurnActivation(pump),
				Effect:    effects.ThisGetsUntilEndOfTurn(0, 1, "Kyscu Drake — +0/+1 until end of turn"),
			},
			{
				Label:   "Sacrifice this creature and a creature named Spitting Drake: Search your library for a card named Viashivan Dragon, put that card onto the battlefield, then shuffle.",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet, Tutors: 1},
				Cost:    effects.Plus(effects.SacrificeThis(), effects.SacrificeN(1, "a creature named Spitting Drake", effects.Creature(), effects.GapB219NamedCard("Spitting Drake"))),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.SearchLibrary{
						Player:    ctx.Controller(),
						Predicate: func(c game.Card) bool { return game.CardNameMatches(c, "Viashivan Dragon") },
						Dest:      game.ZoneBattlefield,
						Limit:     1,
						Shuffle:   true,
						Reason:    "Kyscu Drake — a card named Viashivan Dragon",
					}.Apply(ctx)
				},
			},
		},
	})
}
