package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Beorn, Reluctant Host // Till and Tend — an Adventure card (CR 715,
// The Hobbit, #4454), oracle 288da486…:
//
//	Beorn, Reluctant Host — Legendary Creature — Human Bear Shapeshifter
//	  {4}{G}, 5/5
//	  "Trample"
//	Till and Tend — Sorcery — Adventure {1}{G}
//	  "You may play an additional land this turn. (Then exile this card.
//	   You may cast the creature later from exile.)"
//
// The extra land drop is Explore's grant (CR 305.2): one more land play
// this turn, declared on the Purpose so the bot and the guard see it.
//
// No simplification.

// beornReluctantHostOracleID is shared by both faces of the card.
const beornReluctantHostOracleID = "288da486-393c-4995-ad02-c5093a71970d"

func init() {
	Register(Spec{
		OracleID:        beornReluctantHostOracleID,
		Name:            "Beorn, Reluctant Host",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
	Register(Spec{
		OracleID:     beornReluctantHostOracleID + "#1",
		Name:         "Till and Tend",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{ExtraLandDrops: 1},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ctx.Game.GrantAdditionalLandPlayForEffect(item.Controller, 1)
			return nil
		},
	})
}
