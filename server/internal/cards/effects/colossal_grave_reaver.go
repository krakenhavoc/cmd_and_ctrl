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
// The second watches EventMill — the engine emits one per card, so
// "one or more" is the per-label dedup: the first creature card of
// a mill fires the trigger and the rest of that mill are declined as
// later events of the same batch (OncePerBatch; see docs/adding-cards.md).
// At resolution the batch is read back off the event log
// (b17MilledCreatureCards) — every creature card that mill put into
// the controller's graveyard and is still there — and one of them
// comes back under the controller's control.
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
				Watches:      []game.EventKind{game.EventMill},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.Actor != source.Controller || ev.NewZone != game.ZoneGraveyard {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					if !ok || !c.IsCreature() || c.Owner != source.Controller {
						return false
					}
					return true
				},
				Key: b17GraveReaverReturnLabel,
				Effect: func(g *game.Game, item *game.StackItem) error {
					return b17ReanimateOneMilled(g, item, item.Controller,
						"Colossal Grave-Reaver — put one of the milled creature cards onto the battlefield",
						b17MilledCreatureCards(g, item.Controller, item.Trigger.Event.Seq))
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
		Zone:       game.ZoneGraveyard,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) == 0 {
				return nil
			}
			return ReturnFromGraveyard{
				Target:     picked[0],
				Dest:       game.ZoneBattlefield,
				Controller: item.Controller,
			}.Apply(NewContext(g, item))
		},
	})
	return nil
}
