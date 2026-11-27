package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spoils of the Vault — Instant {B} (Mirrodin, S82 OLD, #5606):
//
//	"Choose a card name. Reveal cards from the top of your library until
//	 you reveal a card with that name, then put that card into your hand.
//	 Exile all other cards revealed this way, and you lose 1 life for each
//	 of the exiled cards."
//
// Demonic Consultation's name-and-reveal without the first exile: the name
// is chosen as the spell resolves (not validated against any list, so a
// name that is not in the library is a legal, ruinous answer), the reveal
// runs from the top until a card with that name (any face, CardNameMatches),
// which goes to hand, and every other revealed card is exiled. The life
// loss is one for each card that really reached exile, counted when the
// exile has settled; a name nowhere in the library reveals and exiles the
// whole library at a life apiece. The loss is not damage (CR 119.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ef86f861-1fdf-4be0-9497-8adc064af3de",
		Name:         "Spoils of the Vault",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			ctx.Game.QueueCardNameChoiceKeyedForEffect(ctx.Controller(), ctx.Source(),
				"Spoils of the Vault — choose a card name",
				ctx.ItemTail(spoilsOfTheVaultNamed))
			return nil
		},
	})
}

// spoilsOfTheVaultNamed is the rest of the spell once the name (the tail's
// Word) is known. An on-disk key.
var spoilsOfTheVaultNamed = effects.CardTail("then/s82-old-b602-spoils-of-the-vault-named", func(ctx *effects.Context, o game.TailOutcome) error {
	player, name := o.Tail.Player, o.Tail.Word
	run, hit := effects.RevealUntil(ctx, player, func(c game.Card) bool {
		return game.CardNameMatches(c, name)
	}, "Spoils of the Vault — reveal until a card named "+name)
	var rest []uuid.UUID
	for _, id := range run {
		if id != hit {
			rest = append(rest, id)
		}
	}
	if hit == uuid.Nil {
		return spoilsOfTheVaultExileRest(ctx.Game, player, ctx.Source(), rest)
	}
	return ctx.Game.TakeFromLibraryToHandKeyedForEffect(player, []uuid.UUID{hit},
		game.KeyedTail{Ref: spoilsOfTheVaultTaken, Player: player, Source: ctx.Source(), Carry: rest})
})

// spoilsOfTheVaultTaken exiles the other revealed cards (the Carry) once
// the named card's move to hand has settled. An on-disk key.
var spoilsOfTheVaultTaken = game.RegisterTail("s82-old-b602-spoils-of-the-vault/taken", func(g *game.Game, o game.TailOutcome) error {
	return spoilsOfTheVaultExileRest(g, o.Tail.Player, o.Tail.Source, o.Tail.Carry)
})

// spoilsOfTheVaultExileRest exiles the revealed cards still in a library;
// the life loss follows the exile.
func spoilsOfTheVaultExileRest(g *game.Game, player, source uuid.UUID, rest []uuid.UUID) error {
	left := effects.CardsStillInALibrary(g, rest)
	if len(left) == 0 {
		return nil
	}
	return g.ExileCardsKeyedForEffect(left, game.KeyedTail{Ref: spoilsOfTheVaultExiled, Player: player, Source: source})
}

// spoilsOfTheVaultExiled is "you lose 1 life for each of the exiled
// cards": the cards that really landed in exile. An on-disk key.
var spoilsOfTheVaultExiled = game.RegisterTail("s82-old-b602-spoils-of-the-vault/exiled", func(g *game.Game, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return g.ChangePlayerLifeForEffect(o.Tail.Source, o.Tail.Player, -len(o.IDs))
})
