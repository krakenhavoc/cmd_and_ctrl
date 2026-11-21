package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Collision of Realms — Sorcery {6}{R} (Kamigawa: Neon Dynasty Commander,
// S82 #5493):
//
//	"Each player shuffles all creatures they own into their library. Each
//	 player who shuffled a nontoken creature into their library this way
//	 reveals cards from the top of their library until they reveal a
//	 creature card, then puts that card onto the battlefield and the rest
//	 on the bottom in a random order."
//
// Every creature on the battlefield goes to its OWNER's library as one
// simultaneous exit (Chaos Warp's tuck; a token ceases to exist, CR 111.7,
// and a commander's owner is offered the command zone, CR 903.9), and the
// rest is the tuck's keyed tail: each owner of a creature shuffles, then
// each player whose tuck really put a nontoken creature into their library
// (the cards the tail is told reached a library) reveals until a creature
// card, in turn order from the active player (CR 101.4), puts it onto the
// battlefield under their control and the rest on the bottom at random
// (The Regalia's sentence, RevealUntilThenPutOntoBattlefield). A library with
// no creature card is all revealed and bottomed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "40cbf2c0-041a-44df-a775-580012caab67",
		Name:         "Collision of Realms",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var creatures, owners []uuid.UUID
			seen := map[uuid.UUID]bool{}
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if !c.IsCreature() {
					continue
				}
				creatures = append(creatures, c.InstanceID)
				if !seen[c.Owner] {
					seen[c.Owner] = true
					owners = append(owners, c.Owner)
				}
			}
			if len(creatures) == 0 {
				return nil
			}
			return ctx.Game.TuckCardsToLibraryKeyedForEffect(creatures, game.TuckOptions{},
				effects.WithCarry(ctx.Tail(collisionOfRealmsShuffleReveal), owners...))
		},
	})
}

// collisionOfRealmsShuffleReveal is the rest of the card: Carry is every
// owner of a creature that was on the battlefield, o.IDs the cards that
// reached a library. An on-disk key: never renamed, never reused.
var collisionOfRealmsShuffleReveal = game.RegisterTail("then/neo-collision-of-realms-shuffle-reveal", func(g *game.Game, o game.TailOutcome) error {
	shuffled := map[uuid.UUID]bool{}
	for _, id := range o.Tail.Carry {
		shuffled[id] = true
	}
	nontoken := map[uuid.UUID]bool{}
	for _, id := range o.IDs {
		if c, ok := g.LookupCardForEffect(id); ok && !c.IsToken() {
			nontoken[c.Owner] = true
		}
	}
	for _, seat := range effects.ApnapLivingSeats(g) {
		if shuffled[seat] {
			if err := g.ShuffleLibraryForEffect(seat); err != nil {
				return err
			}
		}
	}
	for _, seat := range effects.ApnapLivingSeats(g) {
		if !nontoken[seat] {
			continue
		}
		ctx := effects.KeyedContext(g, seat, o.Tail.Source)
		if err := g.PutByEnteringPlayerForEffect(func() error {
			return effects.RevealUntilThenPutOntoBattlefield{
				Match:  func(c game.Card) bool { return c.IsCreature() },
				Reason: "Collision of Realms — reveal cards until a creature card",
			}.Apply(ctx)
		}); err != nil {
			return err
		}
	}
	return nil
})
