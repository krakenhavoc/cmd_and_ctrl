package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gurmag Drowner — Creature — Snake Wizard {3}{U}, 2/4 (Dragons of Tarkir, #5551):
//
//	"Exploit (When this creature enters, you may sacrifice a creature.)
//	 When this creature exploits a creature, look at the top four cards of your library. Put one of them into your hand and the rest into your graveyard."
//
// Exploit is the shared constructor (exploit.go, CR 702.110); the payoff is a reflexive trigger raised only when a creature was really sacrificed.
// The look and the pick are Forbidden Alchemy's (a take of one, the rest to the graveyard).
//
// No simplification.

func init() {
	effects.Register(effects.Spec{
		OracleID:     "9839f80b-d28b-49c7-a31a-40c245af2edc",
		Name:         "Gurmag Drowner",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Exploit("Gurmag Drowner", gurmagDrownerLookBody),
		},
	})
}

// gurmagDrownerLookBody is "When this creature exploits a creature, look at the top four cards of your library. Put one of them into your hand and the rest into your graveyard." An on-disk key:
// never renamed, never reused.
var gurmagDrownerLookBody = game.ReflexiveBody("gurmag-drowner/look", effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	player := ctx.Controller()
	return effects.TakeFromLibraryToHand{
		Player: player,
		Cards:  g.LookAtTopOfLibraryForEffect(player, 4),
		Max:    1,
		Label:  "Gurmag Drowner — put one of them into your hand and the rest into your graveyard",
		Key:    effects.TakeRestGraveyard,
	}.Apply(ctx)
}), nil)
