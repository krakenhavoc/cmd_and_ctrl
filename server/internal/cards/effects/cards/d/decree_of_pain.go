package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "932668fa-d6e3-41c0-ad0c-8e0a00e68d11",
		Name:         "Decree of Pain",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.Creature(), CantBeRegenerated: true, ThenTail: ctx.Tail(effects.DeathBegetsLifeDraw)}.Apply(ctx)
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{3}{B}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.InGraveyard(effects.On(game.EventCycle, effects.Self,
				"Decree of Pain — cycled: all creatures get -2/-2 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{Match: effects.Creature(), Power: -2, Toughness: -2, Label: "Decree of Pain — -2/-2"}.Apply(effects.NewContext(g, item))
				})),
		},
	})
}
