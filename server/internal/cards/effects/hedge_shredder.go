package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hedge Shredder — Artifact — Vehicle {2}{G}{G}, 5/5:
//
//	"Whenever this Vehicle attacks, you may mill two cards.
//	 Whenever one or more land cards are put into your graveyard from
//	 your library, put them onto the battlefield tapped.
//	 Crew 1"
//
// The attack trigger is World Shaper's optional mill.
//
// The land trigger is a "one or more" batch (OncePerBatch, CR 603.2c):
// every card a mill, surveil or other library-to-graveyard move puts in
// your graveyard emits its own EventMill / EventZoneMove, and the first
// land card of the batch fires the ability once. "Them" is every land
// card of that batch, read as the ability resolves from the batch's
// events — a land moved out of the graveyard in response is a new
// object (CR 400.7) and stays where it is, and one put back is not the
// card that was milled. They enter together, tapped, under their
// owner's control (which "your graveyard" makes you).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f22665cd-4710-48df-83c1-3c9081097cb5",
		Name:         "Hedge Shredder",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(WheneverThisAttacks("Hedge Shredder — mill two cards", Do(MillCards{N: 2})), "Hedge Shredder — mill two cards?"),
			OncePerBatch(OnAny([]game.EventKind{game.EventMill, game.EventZoneMove}, hedgeShredderLandMilled,
				"Hedge Shredder — put the land cards onto the battlefield tapped", hedgeShredderReturnLands)),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Hedge Shredder"),
		}},
	})
}

// hedgeShredderLandMilled: a land card you own went from your library
// to your graveyard.
func hedgeShredderLandMilled(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.OldZone != game.ZoneLibrary || ev.NewZone != game.ZoneGraveyard {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsLand() && c.Owner == source.Controller
}

func hedgeShredderReturnLands(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	lands := landsPutIntoGraveyardFromLibraryInBatch(g, item.Trigger.Event.Batch, item.Controller)
	if len(lands) == 0 {
		return nil
	}
	return ReturnFromGraveyardTogether{Targets: lands, Tapped: true}.Apply(NewContext(g, item))
}

// landsPutIntoGraveyardFromLibraryInBatch is the land cards `owner`
// put into their graveyard from their library in event batch `batch`,
// in the order they landed, and only those whose most recent zone
// change this turn is still that move — still the same object in the
// graveyard (CR 400.7).
func landsPutIntoGraveyardFromLibraryInBatch(g *game.Game, batch uint64, owner uuid.UUID) []uuid.UUID {
	events := g.EventsThisTurn()
	lastMove := map[uuid.UUID]int{}
	for i, ev := range events {
		if ev.CardID != uuid.Nil && ev.OldZone != "" && ev.NewZone != "" {
			lastMove[ev.CardID] = i
		}
	}
	var out []uuid.UUID
	for i, ev := range events {
		if ev.Batch != batch || (ev.Kind != game.EventMill && ev.Kind != game.EventZoneMove) {
			continue
		}
		if ev.OldZone != game.ZoneLibrary || ev.NewZone != game.ZoneGraveyard || lastMove[ev.CardID] != i {
			continue
		}
		if z := g.FindCardZoneForEffect(ev.CardID); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if c, ok := g.LookupCardForEffect(ev.CardID); ok && c.IsLand() && c.Owner == owner {
			out = append(out, ev.CardID)
		}
	}
	return out
}
