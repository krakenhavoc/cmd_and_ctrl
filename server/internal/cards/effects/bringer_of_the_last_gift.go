package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bringer of the Last Gift — Creature — Vampire Demon {6}{B}{B}, 6/6 (The
// Lost Caverns of Ixalan, #4470):
//
//	"Flying
//	 When this creature enters, if you cast it, each player sacrifices all
//	 other creatures they control. Then each player returns all creature
//	 cards from their graveyard that weren't put there this way to the
//	 battlefield."
//
// "If you cast it" is an intervening if (CR 603.4) read through the
// permanent's cast record (a reanimated Bringer does nothing). As it
// resolves, each graveyard's creature cards are read first: those are the
// cards "not put there this way". Then every other creature on the
// battlefield is sacrificed as one simultaneous exit (CR 701.21a), and
// only the creature cards read before it return, together, under their
// owners' control (CR 400.3, 614.12), so a creature just sacrificed stays
// in its graveyard, and an entering creature sees the others enter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0fb9a8a2-4331-4259-9460-a37397554962",
		Name:            "Bringer of the Last Gift",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepSacrifice}},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return b06SelfETB(ev, source, lki, g) && b16EnteredFromStack(g, source.InstanceID)
			}, "Bringer of the Last Gift — each player sacrifices all other creatures; each returns the creature cards that weren't put there this way",
				bringerOfTheLastGiftResolve),
		},
	})
}

func bringerOfTheLastGiftResolve(g *game.Game, item *game.StackItem) error {
	if !b16EnteredFromStack(g, item.SourceCardID) {
		return nil
	}
	ctx := NewContext(g, item)
	var returning []uuid.UUID
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		returning = append(returning, graveyardCardIDs(ctx, p.ID, func(c game.Card) bool {
			return c.IsCreature() && !c.IsToken()
		})...)
	}
	var doomed []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.InstanceID != item.SourceCardID {
			doomed = append(doomed, c.InstanceID)
		}
	}
	tail := ctx.Tail(bringerReturnTail)
	tail.Carry = returning
	return g.SacrificeAllKeyedForEffect(ctx.Source(), doomed, tail)
}

// bringerReturnTail is "then each player returns all creature cards from
// their graveyard that weren't put there this way": the cards read before
// the sacrifice ride the tail as its carry and return together once the
// sacrifice is complete, under their owners' control. An on-disk key:
// never renamed, never reused.
var bringerReturnTail = CardTail("then/bringer-of-the-last-gift-return", func(ctx *Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	return ReturnFromGraveyardTogether{Targets: o.Tail.Carry, Controller: uuid.Nil}.Apply(ctx)
})
