package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Treasure Map // Treasure Cove — a transforming artifact (The Lost Caverns
// of Ixalan, #4470):
//
//	Treasure Map — Artifact {2}
//	  "{1}, {T}: Scry 1. Put a landmark counter on this artifact. Then if
//	   there are three or more landmark counters on it, remove those
//	   counters, transform this artifact, and create three Treasure
//	   tokens."
//	Treasure Cove — Land
//	  "{T}: Add {C}.
//	   {T}, Sacrifice a Treasure: Draw a card."
//
// The ability is Brass's Tunnel-Grinder's bore counter after a scry: the
// landmark counter waits for the scry's answer (Scry.ThenTail), and goes
// on only while the Map is still the object that activated it (CR 400.7);
// it goes through the counter window, so a Doubling Season doubles it,
// and the three-counter check waits for it to land. "Remove those counters
// and transform this artifact" removes every landmark counter and turns
// the permanent over in place (ADR 0079's Transform, CR 701.27a), then the
// three Treasures are made. A Map that left before the ability resolved
// takes no counter, so nothing after it happens.
//
// The back face's {T}: Add {C} is a plain mana ability. "{T}, Sacrifice a
// Treasure: Draw a card" is an untargeted instant-speed draw, so it
// declares AnswerValue (ADR 0142), as the Map's own ability does; the Map's
// row also declares its scry 1 rider (#4978).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     treasureMapOracleID,
		Name:         "Treasure Map",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Scry 1. Put a landmark counter on this artifact. Then if there are three or more landmark counters on it, remove those counters, transform this artifact, and create three Treasure tokens.",
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 1},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return Scry{N: 1, ThenTail: ctx.ItemTail(treasureMapLandmark)}.Apply(ctx)
			},
		}},
	})

	Register(Spec{
		OracleID:     treasureMapOracleID + "#1",
		Name:         "Treasure Cove",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "{T}: Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice a Treasure: Draw a card.",
			Cost:    Plus(TapCost(), game.AbilityCost{SacrificeOther: sacrificeSpec("a Treasure", isTreasure)}),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}

// treasureMapOracleID is shared with the back face, which registers under
// it plus "#1" (game.CatalogKey).
const treasureMapOracleID = "0b55eac6-a745-4bf4-8926-5ce83bc38d7d"

// treasureMapLandmarkCounter is the printed counter kind.
const treasureMapLandmarkCounter = "landmark"

// treasureMapLandmark is "put a landmark counter on this artifact", once
// the scry is answered. An on-disk key: never renamed, never reused.
var treasureMapLandmark = CardTail("treasure-map/landmark", func(ctx *Context, _ game.TailOutcome) error {
	if !b09SourceStillOnBattlefield(ctx.Game, ctx.Item) {
		return nil
	}
	return ctx.Game.AddCounterByKeyedForEffect(uuid.Nil, ctx.Item.SourceCardID, treasureMapLandmarkCounter, 1,
		ctx.ItemTail(treasureMapCheckLandmarks))
})

// treasureMapCheckLandmarks is "then if there are three or more landmark
// counters on it, remove those counters", once the new counter has
// landed. An on-disk key: never renamed, never reused.
var treasureMapCheckLandmarks = CardTail("treasure-map/check-landmarks", func(ctx *Context, _ game.TailOutcome) error {
	source := ctx.Item.SourceCardID
	n := b39CountersOn(ctx.Game, source, treasureMapLandmarkCounter)
	if n < 3 || !b09SourceStillOnBattlefield(ctx.Game, ctx.Item) {
		return nil
	}
	return ctx.Game.AddCounterByKeyedForEffect(uuid.Nil, source, treasureMapLandmarkCounter, -n,
		ctx.ItemTail(treasureMapTransform))
})

// treasureMapTransform is "transform this artifact, and create three
// Treasure tokens", behind the counters' removal. An on-disk key: never
// renamed, never reused.
var treasureMapTransform = CardTail("treasure-map/transform", func(ctx *Context, _ game.TailOutcome) error {
	if err := (TransformThis{}).Apply(ctx); err != nil {
		return err
	}
	return CreateToken{Controller: ctx.Controller(), Template: TreasureToken(), N: 3}.Apply(ctx)
})
