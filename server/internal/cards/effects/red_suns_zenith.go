package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Red Sun's Zenith — Sorcery {X}{R}:
//
//	"Red Sun's Zenith deals X damage to any target. If a creature dealt
//	 damage this way would die this turn, exile it instead. Shuffle Red
//	 Sun's Zenith into its owner's library."
//
// "A creature dealt damage this way" is registered from the damage's
// continuation (ADR 0108 §1 decision 3): only on a creature that took
// more than 0 damage. The shuffle is Blue Sun's Zenith's tuck-and-
// shuffle, run from the damage's final continuation so it comes after
// the damage even when a CR 616 prompt pauses it; the resolution sees
// the card has left the stack and does not route it (CR 608.2n). A
// target that has become illegal counters the spell (CR 608.2b): it goes
// to the graveyard, as a countered spell does, and is not shuffled.
//
// One engine-wide simplification: if the damage pauses for a CR 616
// choice, the resolution finishes first and puts the card in the
// graveyard, and it is shuffled into the library from there once the
// choice is made, rather than from the stack.
func init() {
	Register(Spec{
		OracleID:     "82e61db8-4625-488f-8a5f-66ace9bbf34a",
		Name:         "Red Sun's Zenith",
		XMatters:     true,
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			// The shuffle follows the damage as data (#4253): the tail's
			// Player is the owner, its Source the Zenith.
			shuffle := game.KeyedTail{Ref: redSunsZenithShuffle, Player: item.Owner, Source: ctx.Source()}
			t, ok := firstLegalTarget(ctx)
			if !ok {
				return nil
			}
			if x := ctx.X(); x > 0 {
				return ctx.Game.DealDamageEachEachKeyedForEffect(ctx.Source(), []uuid.UUID{t.ID}, x,
					ExileIfDealtDamageWouldDie(item, false), shuffle)
			}
			return shuffle.Run(ctx.Game, nil)
		},
	})
}

// redSunsZenithShuffle is "Shuffle Red Sun's Zenith into its owner's
// library", run with the damage's total: the Source is tucked and its
// owner (the Player) shuffles. An on-disk key.
var redSunsZenithShuffle = game.RegisterTail("red-suns-zenith/shuffle-into-library", func(g *game.Game, o game.TailOutcome) error {
	return g.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{o.Tail.Source}, game.TuckOptions{}, game.KeyedTail{Ref: thenShuffle, Player: o.Tail.Player})
})
