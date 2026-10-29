package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Demonic Consultation — Instant {B} (EDHREC rank 920):
//
//	"Choose a card name. Exile the top six cards of your library, then
//	 reveal cards from the top of your library until you reveal a card
//	 with the chosen name. Put that card into your hand and exile all
//	 other cards revealed this way."
//
// Three steps in order, each waiting on the one before:
//
//  1. The name is chosen as the spell resolves (the 2008-10-01 ruling),
//     with Tamiyo, Collector of Tales' card-name prompt.
//  2. The top six are exiled through MillToZone, so a commander among
//     them offers its owner the command zone (CR 903.9).
//  3. The reveal runs from the new top until a card with that name
//     (revealUntil, CardNameMatches: any face's name). The hit goes to
//     your hand and every other revealed card to exile.
//
// A name that is not in the rest of the library reveals and exiles the
// whole library. That is not a loss on its own (the ruling): you lose
// the next time you would draw from the empty library, which is the
// Thassa's Oracle line the card is played for.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9a1412db-45ad-46ea-8f12-a85d203113d8",
		Name:         "Demonic Consultation",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			player, source := ctx.Controller(), ctx.Source()
			ctx.Game.QueueCardNameChoiceKeyedForEffect(player, source,
				"Demonic Consultation — choose a card name",
				ctx.ItemTail(demonicConsultationNamed))
			return nil
		},
	})
}

// demonicConsultationNamed is the rest of the spell once the name (the
// tail's Word) is known, run against the asking spell. An on-disk key.
var demonicConsultationNamed = CardTail("then/demonic-consultation-named", func(ctx *Context, o game.TailOutcome) error {
	return demonicConsultationExileSix(ctx, o.Tail.Player, o.Tail.Word)
})

// demonicConsultationExileSix exiles the top six, then reveals for the
// named card.
func demonicConsultationExileSix(ctx *Context, player uuid.UUID, name string) error {
	return MillToZone{
		Player:   player,
		N:        6,
		To:       game.ZoneExile,
		ThenTail: withWord(withCarry(ctx.ItemTail(demonicConsultationMilled), player), name),
	}.Apply(ctx)
}

// demonicConsultationMilled is the clause after the mill (#4253 slice 9).
var demonicConsultationMilled = CardTail("demonic-consultation/after-mill", func(ctx *Context, o game.TailOutcome) error {
	player, name := tailCarry0(o), o.Tail.Word

	run, hit := revealUntil(ctx, player, func(c game.Card) bool {
		return game.CardNameMatches(c, name)
	}, "Demonic Consultation — reveal until a card named "+name)
	var rest []uuid.UUID
	for _, id := range run {
		if id != hit {
			rest = append(rest, id)
		}
	}
	if hit == uuid.Nil {
		return demonicConsultationExileRest(ctx.Game, rest)
	}
	// #4253: "exile all other cards revealed this way" is the take's keyed
	// tail, the revealed cards riding its Carry, so a prompt the move to
	// hand pauses on is a restore point.
	return ctx.Game.TakeFromLibraryToHandKeyedForEffect(player, []uuid.UUID{hit},
		game.KeyedTail{Ref: demonicConsultationTaken, Player: player, Source: ctx.Source(), Carry: rest})
})

// demonicConsultationTaken exiles the other revealed cards (the Carry) once
// the named card's move to hand has settled. An on-disk key.
var demonicConsultationTaken = game.RegisterTail("demonic-consultation/taken", func(g *game.Game, o game.TailOutcome) error {
	return demonicConsultationExileRest(g, o.Tail.Carry)
})

// demonicConsultationExileRest exiles the revealed cards still in a library.
func demonicConsultationExileRest(g *game.Game, rest []uuid.UUID) error {
	if len(rest) == 0 {
		return nil
	}
	return g.ExileCardsKeyedForEffect(cardsStillInALibrary(g, rest), game.KeyedTail{})
}
