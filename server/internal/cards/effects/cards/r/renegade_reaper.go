package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Renegade Reaper — Creature — Angel Berserker {2}{B}, 2/3:
//
//	"Flying
//	 When this creature enters, mill four cards. If at least one Angel card is
//	 milled this way, you gain 4 life."
//
// Trystan's keyed mill: the life follows the mill once it has settled and is told
// the cards that actually reached the graveyard (CR 701.17a), so a card a
// replacement kept out of it is not counted. An Angel card is read off the card
// as it sits in the graveyard (its printed types, CR 205.3d), a changeling too.
// A library with fewer than four cards mills what it has.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d72489b5-a213-4bd8-a519-8caeb1acea5f",
		Name:            "Renegade Reaper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{LifeGain: 4},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Renegade Reaper — mill four cards; if at least one Angel card is milled this way, you gain 4 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return g.MillToZoneKeyedForEffect(item.Controller, 4, game.ZoneGraveyard, game.MillUntil{}, ctx.ItemTail(renegadeReaperMilled))
				}),
		},
	})
}

// renegadeReaperMilled is "if at least one Angel card is milled this way, you
// gain 4 life", told the cards that were milled. An on-disk key: never renamed,
// never reused.
var renegadeReaperMilled = effects.CardTail("mill-then/renegade-reaper-angel", func(ctx *effects.Context, o game.TailOutcome) error {
	angel := ctx.TextSubtype("Angel")
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.HasSubtype(angel) {
			return effects.GainLife{Amount: 4}.Apply(ctx)
		}
	}
	return nil
})
