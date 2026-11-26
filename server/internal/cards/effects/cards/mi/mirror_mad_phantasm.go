package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mirror-Mad Phantasm — {3}{U}{U} Creature — Spirit 5/1 (Innistrad, S82 OLD
// b446, #5572):
//
//	"Flying
//	 {1}{U}: This creature's owner shuffles it into their library. If that
//	 player does, they reveal cards from the top of that library until a card
//	 named Mirror-Mad Phantasm is revealed. The player puts that card onto the
//	 battlefield and all other cards revealed this way into their graveyard."
//
// The shuffle-in is the ordinary tuck (a commander's owner is offered the
// command zone, CR 903.9), with the reveal as its continuation: it runs only if
// the card really reached the library ("if that player does"), after the
// shuffle. The card named Mirror-Mad Phantasm may be the one just shuffled in
// or another copy; the owner puts it onto the battlefield, and a library with
// no such card reveals itself entirely and all of it goes to the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fd26127d-6807-41e3-9dcf-20ef257d71d3",
		Name:            "Mirror-Mad Phantasm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}: This creature's owner shuffles it into their library. If that player does, they reveal cards from the top of that library until a card named Mirror-Mad Phantasm is revealed. The player puts that card onto the battlefield and all other cards revealed this way into their graveyard.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{1}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id := item.SourceCardID
				if ctx.IsNewSourceObject(id) || !effects.B15OnBattlefield(g, id) {
					return nil
				}
				c, ok := g.LookupCardForEffect(id)
				if !ok {
					return nil
				}
				return g.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{id}, game.TuckOptions{},
					game.KeyedTail{Ref: mirrorMadPhantasmShuffled, Player: c.Owner, Source: id, Carry: []uuid.UUID{id}})
			},
		}},
	})
}

// mirrorMadPhantasmShuffled is "if that player does, they reveal …": it runs
// only when the Phantasm is in its owner's library. An on-disk key: never
// renamed, never reused.
var mirrorMadPhantasmShuffled = effects.CardTail("mirror-mad-phantasm/shuffled-in-then-reveal", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	owner := o.Tail.Player
	if z := ctx.Game.FindCardZoneForEffect(o.Tail.Carry[0]); z == nil || z.Kind != game.ZoneLibrary {
		return nil
	}
	if err := ctx.Game.ShuffleLibraryForEffect(owner); err != nil {
		return err
	}
	run, hit := effects.RevealUntil(ctx, owner, func(c game.Card) bool { return c.Name == "Mirror-Mad Phantasm" },
		"Mirror-Mad Phantasm — reveal until a card named Mirror-Mad Phantasm")
	var rest []uuid.UUID
	for _, id := range run {
		if id != hit {
			rest = append(rest, id)
		}
	}
	if hit != uuid.Nil {
		if _, err := ctx.Game.PutFromLibraryOntoBattlefieldForEffect(hit, game.LibraryEntryOptions{Controller: owner}); err != nil {
			return err
		}
	}
	return effects.RestIntoGraveyard(ctx.Game, rest)
})
