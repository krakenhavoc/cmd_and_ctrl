package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Burning-Rune Demon — Creature — Demon Berserker {4}{B}{B}, 6/6 (Kaldheim):
//
//	"Flying
//	 When this creature enters, you may search your library for exactly two
//	 cards not named Burning-Rune Demon that have different names. If you do,
//	 reveal those cards. An opponent chooses one of them. Put the chosen card
//	 into your hand and the other into your graveyard, then shuffle."
//
// The search is "you may", for two cards with different names (the registered
// different-names set rule); a short find (fewer than two) is not "exactly
// two" and puts nothing anywhere. The cards stay in the library (put on top
// after the shuffle, so the search reports them found) while the
// opponent, whom the controller picks ("an opponent"), chooses one of them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d551fae8-6b40-4118-b9db-29d2f5dff405",
		Name:            "Burning-Rune Demon",
		Completeness:    effects.CompletenessFull,
		Purpose:         game.Purpose{Tutors: 1},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Burning-Rune Demon — you may search your library for exactly two cards not named Burning-Rune Demon that have different names; an opponent chooses one: it goes to your hand and the other to your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.Name != "Burning-Rune Demon" },
						Dest:      game.ZoneLibrary,
						ToTop:     true,
						Limit:     2,
						Optional:  true,
						Reveal:    true,
						Shuffle:   true,
						Rule:      effects.SearchRuleDifferentNames,
						ThenTail:  ctx.ItemTail(burningRuneDemonFound),
						Reason:    "Burning-Rune Demon — two cards not named Burning-Rune Demon with different names",
					}.Apply(ctx)
				}),
		},
	})
}

// burningRuneDemonFound asks for the opponent once the two cards are found.
// An on-disk key.
var burningRuneDemonFound = effects.CardTail("then/s82-old-b061-burning-rune-demon-found", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) != 2 || len(ctx.Opponents()) == 0 {
		return nil
	}
	return effects.ChoosePlayer{
		Among:    effects.Opponents,
		Question: "Burning-Rune Demon — choose an opponent; they choose one of the two revealed cards",
		ThenKey:  burningRuneDemonOpponent,
		Carry:    append([]uuid.UUID(nil), o.IDs...),
	}.Apply(ctx)
})

// burningRuneDemonOpponent asks the chosen opponent to pick one card. An
// on-disk key.
var burningRuneDemonOpponent = effects.ChoosePlayerThen("option-pick/s82-old-b061-burning-rune-demon-opponent", func(ctx *effects.Context, chosen uuid.UUID, carry []uuid.UUID) error {
	if chosen == uuid.Nil || len(carry) != 2 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    chosen,
		FromPlayer: ctx.Controller(),
		Source:     ctx.Source(),
		Question:   "Burning-Rune Demon — choose one of the revealed cards: it goes to their hand, the other to their graveyard",
		Cards:      carry,
		Min:        1,
		Max:        1,
		ThenKey:    burningRuneDemonPicked,
		Carry:      carry,
		LiveZone:   game.ZoneLibrary,
		Asked:      game.NewAskedItem(ctx.Item),
	})
	return nil
})

// burningRuneDemonPicked puts the chosen card into the controller's hand and
// the other into their graveyard. An on-disk key.
var burningRuneDemonPicked = effects.AskedPick("choose-cards/s82-old-b061-burning-rune-demon-picked", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) != 1 {
		return nil
	}
	var rest []uuid.UUID
	for _, id := range r.Carry {
		if id != r.Picked[0] {
			rest = append(rest, id)
		}
	}
	return effects.FactOrFictionSettle(ctx.Game, r.Picked, rest)
})
