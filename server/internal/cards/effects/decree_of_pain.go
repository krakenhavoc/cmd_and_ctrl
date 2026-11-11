package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Decree of Pain — Sorcery {6}{B}{B}:
//
//	"Destroy all creatures. They can't be regenerated. Draw a card for each
//	 creature destroyed this way.
//	 Cycling {3}{B}{B}
//	 When you cycle this card, all creatures get -2/-2 until end of turn."
//
// The wipe is Wrath of God's (regeneration is ignored, CR 701.19c) with the
// draw hung on the destruction as Death Begets Life's tail, which counts the
// creatures that really died. Cycling is the shared CR 702.29a ability; the
// cycle trigger is Fractured Sanity's graveyard watcher (the card is in the
// graveyard by the time it fires, CR 702.29c), and the -2/-2 is Golgari
// Charm's one-turn boost over every creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "932668fa-d6e3-41c0-ad0c-8e0a00e68d11",
		Name:         "Decree of Pain",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: Creature(), CantBeRegenerated: true, ThenTail: ctx.Tail(deathBegetsLifeDraw)}.Apply(ctx)
		},
		Activated: []ActivatedAbility{Cycling("{3}{B}{B}")},
		Triggered: []game.TriggeredAbility{
			InGraveyard(On(game.EventCycle, Self,
				"Decree of Pain — cycled: all creatures get -2/-2 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{Match: Creature(), Power: -2, Toughness: -2, Label: "Decree of Pain — -2/-2"}.Apply(NewContext(g, item))
				})),
		},
	})
}
