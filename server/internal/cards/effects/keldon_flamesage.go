package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Keldon Flamesage — Creature — Human Shaman {2}{R}, 2/3 (Dominaria United,
// #5487, seam #5658):
//
//	"Enlist
//	 Whenever this creature attacks, look at the top X cards of your
//	 library, where X is this creature's power. You may exile an instant
//	 or sorcery card with mana value X or less from among them. Put the
//	 rest on the bottom of your library in a random order. You may cast
//	 the exiled card without paying its mana cost."
//
// Enlist is the engine's keyword (CR 702.154, game/enlist.go). Its linked
// +X/+0 trigger and the attack trigger go on the stack together, in the
// order the player chooses, so the enlisted creature's power is in X when
// the enlist trigger resolves first.
//
// The attack trigger is The Key to the Vault's look, exile, bottom and
// cast. X is the Flamesage's power as the trigger resolves, its last-known
// power if it has left (CR 608.2h), and the same X bounds the mana value
// of the card that may be exiled. The look is private. The chosen card is
// exiled first, the rest go to the bottom in a random order, and then the
// free cast is offered as part of the resolution (CR 608.2g); a card that
// is not cast stays in exile, as printed. A card with {X} in its cost is
// cast with X as 0 (CR 107.3b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0efedec3-b57c-4395-ad72-19d79560c99f",
		Name:            "Keldon Flamesage",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordEnlist},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Keldon Flamesage — look at the top X cards, exile an instant or sorcery with mana value X or less, cast it free",
				keldonFlamesageLook),
		},
	})
}

// keldonFlamesageLook looks at the top X cards and asks which instant or
// sorcery card with mana value X or less, if any, to exile.
func keldonFlamesageLook(g *game.Game, item *game.StackItem) error {
	x := 0
	if info, ok := NewContext(g, item).SourcePermanent(); ok {
		x = info.Power
	}
	if x <= 0 {
		return nil
	}
	seen := g.LookAtTopOfLibraryForEffect(item.Controller, x)
	if len(seen) == 0 {
		return nil
	}
	var spells []uuid.UUID
	for _, id := range seen {
		c, ok := g.LookupCardForEffect(id)
		if ok && (c.IsInstant() || c.IsSorcery()) && c.ManaValue() <= x {
			spells = append(spells, id)
		}
	}
	if len(spells) == 0 {
		return keldonFlamesageFinish(g, item.Controller, item.SourceCardID, seen, uuid.Nil)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Keldon Flamesage — exile an instant or sorcery card with mana value " + strconv.Itoa(x) + " or less? You may cast it without paying its mana cost.",
		Cards:    spells,
		Min:      0,
		Max:      1,
		Zone:     game.ZoneLibrary,
		ThenKey:  keldonFlamesagePicked,
		Carry:    seen,
	})
	return nil
}

// keldonFlamesagePicked is the answer to the pick; the cards looked at are
// its Carry. An on-disk key.
var keldonFlamesagePicked = game.RegisterChooseCardsThen("choose-cards/keldon-flamesage", func(g *game.Game, r game.CardsChosen) error {
	exiled := uuid.Nil
	if len(r.Picked) == 1 {
		exiled = r.Picked[0]
	}
	return keldonFlamesageFinish(g, r.Chooser, r.Source, r.Carry, exiled)
})

// keldonFlamesageFinish exiles the chosen card, if any; the exile's tail
// does the rest.
func keldonFlamesageFinish(g *game.Game, you, source uuid.UUID, seen []uuid.UUID, exiled uuid.UUID) error {
	var others []uuid.UUID
	for _, id := range cardsStillInALibrary(g, seen) {
		if id != exiled {
			others = append(others, id)
		}
	}
	then := game.KeyedTail{Ref: keldonFlamesageExiled, Player: you, Source: source, Carry: others}
	if exiled == uuid.Nil {
		return then.Run(g, nil)
	}
	return g.ExileCardKeyedForEffect(exiled, then)
}

// keldonFlamesageExiled is the exile's tail: the other cards (its Carry)
// go to the bottom in a random order, then the card that reached exile may
// be cast free during the resolution. An on-disk key.
var keldonFlamesageExiled = game.RegisterTail("then/keldon-flamesage-bottom-then-cast", func(g *game.Game, o game.TailOutcome) error {
	if err := g.PutOnBottomInRandomOrderForEffect(o.Tail.Player, game.ZoneLibrary, o.Tail.Carry); err != nil {
		return err
	}
	if len(o.IDs) != 1 {
		return nil
	}
	grantFreeCasts(g, o.Tail.Player, o.Tail.Source, "Keldon Flamesage", game.TimingFlash, game.LapseStaysInExile, o.IDs)
	return nil
})
