package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rally the Horde — Sorcery {5}{R} (Saviors of Kamigawa, #5601):
//
//	"Exile the top card of your library. Exile the top card of your
//	 library. Exile the top card of your library. If the last card exiled
//	 isn't a land card, repeat this process. Create a 1/1 red Warrior
//	 creature token for each nonland card exiled this way."
//
// The process is a run of one-card exiles in groups of three (a mill run
// with its own stop clause, OldB590UntilAGroupOfThreeEndsOnALand): it stops
// once a group's last card is a land, or when the library is empty
// (CR 701.17b). The Warriors are made once the run has ended, one for each
// nonland card among the cards that landed in exile; lands exiled on the way
// make none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fd9aeda1-9e44-4bc6-aa0d-19d58918d1f4",
		Name:         "Rally the Horde",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return ctx.Game.MillToZoneKeyedForEffect(ctx.Controller(), 0, game.ZoneExile,
				effects.OldB590UntilAGroupOfThreeEndsOnALand, ctx.ItemTail(rallyTheHordeWarriors))
		},
	})
}

// rallyTheHordeWarriors makes a Warrior for each nonland card the run
// exiled. An on-disk key.
var rallyTheHordeWarriors = effects.CardTail("rally-the-horde/warriors", func(ctx *effects.Context, o game.TailOutcome) error {
	n := effects.OldB590CardsExiledThatAreNonland(ctx.Game, o.IDs)
	if n <= 0 {
		return nil
	}
	return effects.CreateToken{Template: effects.TokenCard("1/1 red Warrior"), N: n}.Apply(ctx)
})
