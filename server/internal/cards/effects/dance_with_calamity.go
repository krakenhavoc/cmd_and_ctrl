package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dance with Calamity — Sorcery {7}{R}:
//
//	"Shuffle your library. As many times as you choose, you may exile
//	 the top card of your library. If the total mana value of the cards
//	 exiled this way is 13 or less, you may cast any number of spells
//	 from among those cards without paying their mana costs."
//
// The shuffle, then a keyed yes/no loop: each "yes" exiles the top card
// face up, and the question comes back naming the running total, so the
// player sees every card before deciding whether to take another. "No",
// or an empty library, ends the loop. Only a card that really reached
// exile counts as "exiled this way" (a replacement that sent it elsewhere
// adds nothing), and the total is their mana values as cards in exile.
//
// With a total of 13 or less, each exiled card is granted a free cast
// (cost {0}, so X is 0) and offered during the resolution, one answer per
// card until the player declines or none is left (Emergent Ultimatum's
// grant). A land among them is never offered, since it is played, not
// cast. With 14 or more nothing is cast. Every card that is not cast
// stays in exile.
//
// No simplification.
func init() {
	// Registered here, not in a var block: the answer asks again, which
	// names this key (an initialization cycle otherwise, as in Hidden
	// Strings).
	danceWithCalamityAnswer = game.RegisterConfirmThen("confirm/dance-with-calamity-exile-top", danceWithCalamityAnswered)
	Register(Spec{
		OracleID:     "615b4189-e800-408b-99e0-71ae7089b937",
		Name:         "Dance with Calamity",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			you := ctx.Controller()
			if err := (ShuffleLibrary{Player: you}).Apply(ctx); err != nil {
				return err
			}
			return danceWithCalamityAsk(ctx.Game, you, item.SourceCardID, nil)
		},
	})
}

const danceWithCalamityName = "Dance with Calamity"

// danceWithCalamityAsk asks whether to exile another card, or settles the
// casts when the library is empty.
func danceWithCalamityAsk(g *game.Game, you, source uuid.UUID, exiled []uuid.UUID) error {
	p := g.PlayerByIDForEffect(you)
	if p == nil || p.Library == nil || len(p.Library.Cards) == 0 {
		return danceWithCalamitySettle(g, you, source, exiled)
	}
	_, err := g.QueueConfirmKeyedForEffect(game.ConfirmPrompt{
		Chooser: you,
		Source:  source,
		Question: danceWithCalamityName + " — exile the top card of your library? Total mana value exiled so far: " +
			strconv.Itoa(danceWithCalamityTotal(g, exiled)) + " (13 or less lets you cast them free)",
		AcceptLabel:  "Exile the top card",
		DeclineLabel: "Stop",
	}, danceWithCalamityAnswer, exiled)
	return err
}

// danceWithCalamityAnswer is the loop's answer: exile the top card and
// ask again, or stop and settle. An on-disk key, assigned in init.
var danceWithCalamityAnswer game.ConfirmThen

func danceWithCalamityAnswered(g *game.Game, r game.ConfirmAnswer) error {
	if !r.Accepted {
		return danceWithCalamitySettle(g, r.Chooser, r.Source, r.Carry)
	}
	p := g.PlayerByIDForEffect(r.Chooser)
	if p == nil || p.Library == nil || len(p.Library.Cards) == 0 {
		return danceWithCalamitySettle(g, r.Chooser, r.Source, r.Carry)
	}
	top := p.Library.Cards[len(p.Library.Cards)-1].InstanceID
	return g.ExileCardKeyedForEffect(top, game.KeyedTail{
		Ref: danceWithCalamityExiled, Player: r.Chooser, Source: r.Source, Carry: append([]uuid.UUID(nil), r.Carry...),
	})
}

// danceWithCalamityExiled adds the card that reached exile to the cards
// exiled this way and asks again.
var danceWithCalamityExiled = game.RegisterTail("then/dance-with-calamity-exiled", func(g *game.Game, o game.TailOutcome) error {
	exiled := append(append([]uuid.UUID(nil), o.Tail.Carry...), o.IDs...)
	return danceWithCalamityAsk(g, o.Tail.Player, o.Tail.Source, exiled)
})

// danceWithCalamityTotal is the total mana value of the cards exiled this
// way.
func danceWithCalamityTotal(g *game.Game, exiled []uuid.UUID) int {
	total := 0
	for _, id := range exiled {
		if c, ok := g.LookupCardForEffect(id); ok {
			total += c.ManaValue()
		}
	}
	return total
}

// danceWithCalamitySettle is "if the total mana value of the cards exiled
// this way is 13 or less, you may cast any number of spells from among
// those cards without paying their mana costs".
func danceWithCalamitySettle(g *game.Game, you, source uuid.UUID, exiled []uuid.UUID) error {
	if len(exiled) == 0 || danceWithCalamityTotal(g, exiled) > 13 {
		return nil
	}
	grantFreeCasts(g, you, source, danceWithCalamityName, game.TimingFlash, game.LapseStaysInExile, exiled)
	return nil
}
