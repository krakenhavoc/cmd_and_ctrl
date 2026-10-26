package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const b17GraveReaverReturnLabel = "Colossal Grave-Reaver — put a milled creature card onto the battlefield"

// Colossal Grave-Reaver — Creature — Dragon {6}{B}{G}, 7/6 (EDHREC
// rank 1890):
//
//	"Flying
//	 Whenever this creature enters or attacks, mill three cards.
//	 Whenever one or more creature cards are put into your graveyard
//	 from your library, put one of them onto the battlefield."
//
// The self-mill Dragon that reanimates as it goes. The first trigger
// is Sun Titan's "enters or attacks" shape with a three-card mill.
// The second watches every library-to-graveyard arrival — a mill's
// EventMill and the EventZoneMove of a surveil or a search (#2809,
// library_to_graveyard.go) — one per card, so "one or more" is the
// per-label dedup: the first creature card of a batch fires the
// trigger and the rest are declined as later events of the same batch
// (OncePerBatch; see docs/adding-cards.md). At resolution the batch is
// read back off the event log (cardsPutIntoGraveyardFromLibraryInBatch)
// — every creature card the batch put into the controller's graveyard
// from their library that is still there — and one of them comes back
// under the controller's control.
//
// "Put ONE OF THEM" is the controller's choice (#2523): with more than
// one creature card in the batch the trigger's resolution queues a
// choose-cards prompt over the batch (exactly one, from the graveyard,
// re-checked on submit), and a lone creature card just comes back with
// nothing to decide. The batch is read at RESOLUTION, so a milled
// creature card that has left the graveyard since is no candidate.
// Two mills in one resolution that are not one batch fire once, not
// twice: weaker, never stronger.
func init() {
	Register(Spec{
		OracleID:        "df8e0d1b-b47c-4807-9c9b-84dcec835254",
		Name:            "Colossal Grave-Reaver",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Colossal Grave-Reaver — mill three cards", Do(MillCards{N: 3})),
			{
				OncePerBatch: true,
				Watches:      libraryToGraveyardKinds,
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return cardPutIntoYourGraveyardFromYourLibrary(ev, source, g, isCreatureCard)
				},
				Key: b17GraveReaverReturnLabel,
				Effect: func(g *game.Game, item *game.StackItem) error {
					return b17ReanimateOneMilled(g, item, item.Controller,
						"Colossal Grave-Reaver — put one of the milled creature cards onto the battlefield",
						cardsPutIntoGraveyardFromLibraryInBatch(g, item.Trigger.Event.Batch, item.Controller, isCreatureCard))
				},
			},
		},
	})
}

// b17ReanimateOneMilled is "put one of them onto the battlefield": a
// lone candidate returns outright, several are offered to the
// controller as an exactly-one pick. `from` is the graveyard's owner
// (Helm of Obedience reanimates out of somebody else's).
func b17ReanimateOneMilled(g *game.Game, item *game.StackItem, from uuid.UUID, question string, creatures []uuid.UUID) error {
	switch len(creatures) {
	case 0:
		return nil
	case 1:
		return ReturnFromGraveyard{
			Target:     creatures[0],
			Dest:       game.ZoneBattlefield,
			Controller: item.Controller,
		}.Apply(NewContext(g, item))
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: from,
		Source:     item.SourceCardID,
		Question:   question,
		Cards:      creatures,
		Min:        1,
		Max:        1,
		ThenKey:    b17ReanimatePicked,
		LiveZone:   game.ZoneGraveyard,
		Asked:      game.NewAskedItem(item),
	})
	return nil
}

// b17ReanimatePicked is the keyed answer to "put one of them onto the
// battlefield" (#4253): the asking item rides the prompt as data. Shared by
// every card that reanimates one of a milled batch. An on-disk key: never
// renamed, never reused.
var b17ReanimatePicked = game.RegisterChooseCardsThen("choose-cards/reanimate-one-milled", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 || r.Asked == nil {
		return nil
	}
	return ReturnFromGraveyard{
		Target:     r.Picked[0],
		Dest:       game.ZoneBattlefield,
		Controller: r.Asked.Controller,
	}.Apply(NewContext(g, r.Asked))
})
