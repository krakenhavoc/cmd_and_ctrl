package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Flower // Flourish — split card, oracle 2f8243a4-f347-4e18-9094-570b34df903b
// (Guilds of Ravnica):
//
//	Flower {G/W} Sorcery: "Search your library for a basic Forest or Plains
//	 card, reveal it, put it into your hand, then shuffle."
//	Flourish {4}{G}{W} Sorcery: "Creatures you control get +2/+2 until end of
//	 turn."
//
// Two specs, Flourish under "<oracle_id>#1" as Discovery // Dispersal is.
// Flower is Farfinder's optional fetch restricted to a basic Forest or Plains
// (a basic land with one of those types); Flourish is Ampryn Tactician's team
// pump, the creatures being those its controller has as it resolves.
//
// No simplification.
func init() {
	const oracle = "2f8243a4-f347-4e18-9094-570b34df903b"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Flower",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:    item.Controller,
				Predicate: effects.IsBasicLandOfAnySubtype("Forest", "Plains"),
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Optional:  true,
				Reason:    "Flower — a basic Forest or Plains card, to your hand",
			}.Apply(ctx)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Flourish",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BoostUntilEOT{
				Match: effects.And(effects.Creature(), effects.YouControl()),
				Power: 2, Toughness: 2,
				Label: "Flourish — creatures you control get +2/+2",
			}.Apply(ctx)
		},
	})
}
