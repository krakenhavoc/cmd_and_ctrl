package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reality Scramble — Sorcery {2}{R}{R}:
//
//	"Put target permanent you own on the bottom of your library. Reveal
//	 cards from the top of your library until you reveal a card that
//	 shares a card type with that permanent. Put that card onto the
//	 battlefield and the rest on the bottom of your library in a random
//	 order.
//	 Retrace"
//
// The reveal is the shared reveal-until sentence
// (RevealUntilThenPutOntoBattlefield, #745) with a Match built at
// resolution: the target's card types are read BEFORE it is tucked (it
// has left the battlefield by the time the reveal runs, and a token
// tucked this way ceases to exist but still had types), and the closure
// carries only those strings, never a *Card.
//
// The tuck goes through TuckToLibraryThenForEffect, so a commander's
// owner is offered the command zone (CR 903.9) and the reveal waits for
// that answer. The tucked card sits at the bottom, so a library with no
// other matching card eventually reveals it again; that is the printed
// outcome. A revealed token never stops the run (CR 108.2).
//
// Retrace (CR 702.81) is the graveyard cast for the printed cost plus a
// discarded land card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "afad2e76-4b53-4884-8f96-ba68b0808990",
		Name:             "Reality Scramble",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Retrace("{2}{R}{R}")},
		Targets:          TargetPermanent("target permanent you own", YouOwn()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			// Keyed (#4253 slice 10a): the card's types ride the tail's
			// Word, read before it moved.
			types := strings.Join(c.Effective().Types, ",")
			return ctx.Game.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{id}, game.TuckOptions{ToBottom: true},
				withWord(ctx.ItemTail(realityScrambleReveal), types))
		},
	})
}

// realityScrambleReveal is "reveal cards until one that shares a card type
// with it", after the tuck: the types are the tail's Word.
var realityScrambleReveal = CardTail("then/reality-scramble-reveal", func(ctx *Context, o game.TailOutcome) error {
	types := strings.Split(o.Tail.Word, ",")
	shares := func(r game.Card) bool {
		for _, t := range types {
			if t != "" && r.HasCardType(strings.ToLower(t)) {
				return true
			}
		}
		return false
	}
	return RevealUntilThenPutOntoBattlefield{
		Match:  shares,
		Reason: "Reality Scramble — revealed until a card sharing a type",
	}.Apply(ctx)
})
