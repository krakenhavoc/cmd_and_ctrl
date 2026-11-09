package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Twilight Diviner — Creature — Elf Cleric {2}{B}, 3/3 (Lorwyn Eclipsed,
// #4458):
//
//	"When this creature enters, surveil 2.
//	 Whenever one or more other creatures you control enter, if they
//	 entered or were cast from a graveyard, create a token that's a copy
//	 of one of them. This ability triggers only once each turn."
//
// The second ability is Celes, Rune Knight's batch trigger with its
// intervening if (celesOtherCreatureFromGraveyard: entered from a
// graveyard, EventETB.EnteredFrom, or was cast from one,
// Card.Provenance.FromZone), made once a turn by TriggersOnlyOnceEachTurn.
// As it resolves the creatures of that batch that qualify are offered
// and the controller picks one (no pick with a single candidate); the
// token is a copy of its copiable values (TokenCopyTemplate, CR 707.2),
// its last-known ones if it has left (CR 608.2h), so a creature copying
// something else hands on what it copies, and its enters abilities fire
// on the token (the rulings).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "05d7222b-cc08-4f1f-8e35-6058599572ba",
		Name:         "Twilight Diviner",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Twilight Diviner — surveil 2", Do(Surveil{N: 2})),
			TriggersOnlyOnceEachTurn(OncePerBatch(On(game.EventETB, celesOtherCreatureFromGraveyard,
				"Twilight Diviner — create a token that's a copy of one of the creatures that came from a graveyard",
				twilightDivinerCopyOne))),
		},
	})
}

// twilightDivinerCopyOne offers the trigger's batch of creatures that
// came from a graveyard and copies the one picked.
func twilightDivinerCopyOne(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	first := item.Trigger.Event
	source := &game.Card{InstanceID: item.SourceCardID, Controller: item.Controller}
	var candidates []uuid.UUID
	seen := map[uuid.UUID]bool{}
	for _, ev := range g.EventsThisTurn() {
		// An unstamped event (Batch zero) is a batch of one: itself.
		sameBatch := ev.Batch == first.Batch && (first.Batch != 0 || ev.Seq == first.Seq)
		if !sameBatch || seen[ev.CardID] || !celesOtherCreatureFromGraveyard(ev, source, game.Characteristic{}, g) {
			continue
		}
		seen[ev.CardID] = true
		candidates = append(candidates, ev.CardID)
	}
	if len(candidates) == 0 && first.CardID != uuid.Nil {
		// The triggering creature itself, whatever has become of it.
		candidates = []uuid.UUID{first.CardID}
	}
	switch len(candidates) {
	case 0:
		return nil
	case 1:
		return CreateTokenCopy{Controller: item.Controller, Copy: candidates[0], N: 1}.Apply(NewContext(g, item))
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Twilight Diviner — choose a creature to create a token copy of",
		Cards:    candidates,
		Min:      1,
		Max:      1,
		ThenKey:  twilightDivinerPick,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// twilightDivinerPick is the pick's continuation: a token copy of the
// creature chosen, for the trigger's controller. An on-disk key.
var twilightDivinerPick = game.RegisterChooseCardsThen("choose-cards/twilight-diviner-copy", func(g *game.Game, r game.CardsChosen) error {
	item := r.Asked
	if item == nil || len(r.Picked) == 0 {
		return nil
	}
	return CreateTokenCopy{Controller: item.Controller, Copy: r.Picked[0], N: 1}.Apply(NewContext(g, item))
})
