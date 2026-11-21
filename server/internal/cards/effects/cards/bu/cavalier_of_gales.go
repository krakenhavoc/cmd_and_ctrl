package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cavalier of Gales — Creature — Elemental Knight {2}{U}{U}{U}, 5/5
// (Foundations Jumpstart):
//
//	"Flying
//	 When this creature enters, draw three cards, then put two cards from
//	 your hand on top of your library in any order.
//	 When this creature dies, shuffle it into its owner's library, then
//	 scry 2."
//
// The enters trigger is Hidetsugu and Kairi's Brainstorm. The dies trigger
// tucks the card from the graveyard into its OWNER's library (Chaos, the
// Endless' route: a replacement window and a commander's command-zone
// offer are honoured; a card that left the graveyard in response is a new
// object and is left alone, CR 400.7), the owner shuffles, and the
// Cavalier's controller then scries 2.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "08e2fead-53e9-445b-8fd1-c108bb09d81c",
		Name:            "Cavalier of Gales",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Scry: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Cavalier of Gales — draw three cards, then put two cards from your hand on top of your library in any order",
				func(g *game.Game, item *game.StackItem) error {
					return g.DrawNThenForEffect(item.Controller, 3, game.DrawThen{
						Ref: cavalierOfGalesPutBack, Player: item.Controller, Source: item.SourceCardID, N: 2,
					})
				}),
			effects.WhenThisDies("Cavalier of Gales — shuffle it into its owner's library, then scry 2", cavalierOfGalesTuck),
		},
	})
}

var cavalierOfGalesPutBack = game.RegisterDrawThen("s82-bns-b27/cavalier-of-gales-put-back", func(g *game.Game, d game.DrawThen) error {
	ctx := effects.NewContext(g, &game.StackItem{Controller: d.Player, SourceCardID: d.Source})
	return effects.PutFromHandOnTopInAnyOrder{
		Player: d.Player,
		N:      d.N,
		Label:  "Cavalier of Gales — put two cards from your hand on top of your library",
	}.Apply(ctx)
})

// cavalierOfGalesTuck shuffles the Cavalier into its owner's library and,
// once that is done, has its controller scry 2. A Cavalier no longer in
// the graveyard is not moved, and so nothing follows ("then" is a
// sequence, not a condition, but the scry is the same instruction after a
// shuffle that did not happen: CR 608.2c, the controller still scries).
func cavalierOfGalesTuck(g *game.Game, item *game.StackItem) error {
	z := g.FindCardZoneForEffect(item.SourceCardID)
	owner := item.Owner
	if c, ok := g.LookupCardForEffect(item.SourceCardID); ok {
		owner = c.Owner
	}
	if z == nil || z.Kind != game.ZoneGraveyard {
		return effects.Scry{Player: item.Controller, N: 2}.Apply(effects.NewContext(g, item))
	}
	return g.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{item.SourceCardID}, game.TuckOptions{},
		game.KeyedTail{Ref: cavalierOfGalesShuffleScry, Player: owner, Source: item.SourceCardID, Carry: []uuid.UUID{item.Controller}})
}

// cavalierOfGalesShuffleScry is "shuffle it into its owner's library, then
// scry 2": Tail.Player shuffles, then Tail.Carry[0] scries. An on-disk key.
var cavalierOfGalesShuffleScry = effects.CardTail("s82-bns-b27/cavalier-of-gales-shuffle-scry", func(ctx *effects.Context, o game.TailOutcome) error {
	if err := ctx.Game.ShuffleLibraryForEffect(o.Tail.Player); err != nil {
		return err
	}
	scryer := o.Tail.Player
	if len(o.Tail.Carry) > 0 {
		scryer = o.Tail.Carry[0]
	}
	return effects.Scry{Player: scryer, N: 2}.Apply(ctx)
})
