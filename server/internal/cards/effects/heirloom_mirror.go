package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Heirloom Mirror // Inherited Fiend — Artifact {1}{B} // Creature — Demon
// 4/4 (Midnight Hunt, Innistrad: Double Feature):
//
//	Heirloom Mirror
//	  "{1}, {T}, Pay 1 life, Discard a card: Draw a card, mill a card, then
//	   put a ritual counter on this artifact. Then if it has three or more
//	   ritual counters on it, remove them and transform it. Activate only
//	   as a sorcery."
//	Inherited Fiend
//	  "Flying
//	   {2}{B}: Exile target creature card from a graveyard. Put a +1/+1
//	   counter on this creature."
//
// Brass's Tunnel-Grinder's shape: the ritual counter goes through the
// counter window (a Doubling Season doubles it), so the three-counter check
// waits for it to land, and "remove them and transform it" takes off what
// is there and turns the same permanent over in place. A Mirror that has
// left the battlefield by resolution gets no counter. The Fiend's counter
// does not depend on the exile (the ability fizzles whole if its target is
// gone), so it is put first, which no one can tell from the other order.
//
// No simplification.
func init() {
	const oracle = "b33390dd-f2f8-4059-83a9-45c0edccd5e4"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Heirloom Mirror",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{1}, {T}, Pay 1 life, Discard a card: Draw a card, mill a card, then put a ritual counter on this artifact. Then if it has three or more ritual counters on it, remove them and transform it. Activate only as a sorcery.",
			Purpose:      game.Purpose{Draws: 1, Discards: 1},
			Cost:         Plus(ManaCost("{1}"), TapCost(), PayLife(1), DiscardACard()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (DrawCards{N: 1}).Apply(ctx); err != nil {
					return err
				}
				if err := (MillCards{N: 1}).Apply(ctx); err != nil {
					return err
				}
				if !b09SourceStillOnBattlefield(g, item) {
					return nil
				}
				return g.AddCounterByKeyedForEffect(uuid.Nil, item.SourceCardID, heirloomMirrorRitual, 1,
					ctx.ItemTail(heirloomMirrorCheck))
			},
		}},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Inherited Fiend",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}: Exile target creature card from a graveyard. Put a +1/+1 counter on this creature.",
			Cost:    ManaCost("{2}{B}"),
			Targets: TargetCardInGraveyard("target creature card from a graveyard", Creature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (CounterOnThis{Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
					return err
				}
				for _, id := range legalTargetIDs(ctx) {
					return ExileTarget{Target: id}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}

// heirloomMirrorRitual is the printed counter kind.
const heirloomMirrorRitual = "ritual"

// heirloomMirrorCheck is "then if it has three or more ritual counters on
// it, remove them and transform it", read once the new counter has landed.
// An on-disk key: never renamed, never reused.
var heirloomMirrorCheck = CardTail("heirloom-mirror/check-ritual", func(ctx *Context, _ game.TailOutcome) error {
	source := ctx.Item.SourceCardID
	n := b39CountersOn(ctx.Game, source, heirloomMirrorRitual)
	if n < 3 {
		return nil
	}
	return ctx.Game.AddCounterByKeyedForEffect(uuid.Nil, source, heirloomMirrorRitual, -n,
		ctx.ItemTail(heirloomMirrorTransform))
})

// heirloomMirrorTransform is the transform behind the ritual counters'
// removal. An on-disk key: never renamed, never reused.
var heirloomMirrorTransform = CardTail("heirloom-mirror/transform", func(ctx *Context, _ game.TailOutcome) error {
	return TransformThis{}.Apply(ctx)
})
