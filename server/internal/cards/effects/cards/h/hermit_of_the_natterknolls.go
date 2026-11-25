package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hermit of the Natterknolls // Lone Wolf of the Natterknolls — Creature —
// Human Werewolf {2}{G}, 2/3 // Creature — Werewolf 3/5 (Shadows over
// Innistrad, #5543):
//
//	Front: "Whenever an opponent casts a spell during your turn, draw a
//	        card.
//	        At the beginning of each upkeep, if no spells were cast last
//	        turn, transform this creature."
//	Back:  "Whenever an opponent casts a spell during your turn, draw two
//	        cards.
//	        At the beginning of each upkeep, if a player cast two or more
//	        spells last turn, transform this creature."
//
// The draw trigger is Voice of Resurgence's cast condition: an opponent
// casts a spell while it is the controller's turn. The transforms are the
// shared Innistrad werewolf triggers (innistrad_werewolves.go), intervening
// ifs on the previous turn's casts.
//
// No simplification.
func init() {
	const oracle = "6d32f688-a854-4d82-b654-d720b93fb766"
	opponentCastsOnYourTurn := func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return ev.Kind == game.EventCast && effects.ByAnOpponent(ev, source, lki, g) && effects.IsYourTurn(g, source.Controller)
	}
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Hermit of the Natterknolls",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, opponentCastsOnYourTurn,
				"Hermit of the Natterknolls — draw a card", effects.Do(effects.DrawCards{N: 1})),
			effects.TransformAtUpkeepIfNoSpellsLastTurn("Hermit of the Natterknolls"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Lone Wolf of the Natterknolls",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, opponentCastsOnYourTurn,
				"Lone Wolf of the Natterknolls — draw two cards", effects.Do(effects.DrawCards{N: 2})),
			effects.TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("Lone Wolf of the Natterknolls"),
		},
	})
}
