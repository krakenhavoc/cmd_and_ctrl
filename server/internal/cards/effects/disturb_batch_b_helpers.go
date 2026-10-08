package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// disturb_batch_b_helpers.go — shared pieces of the second batch of
// disturb cards (#1855, ADR 0107 §4). Each one is used by both faces of
// one card, or by two cards, which is why it is not in a card file.

// AttackingAlone is CR 506.5's "is attacking alone": the creature is
// attacking and no other creature is. Read live, so a creature whose
// fellow attackers were removed from combat is attacking alone from
// then on. Counted by game.AttackedAlone, exalted's own count.
func AttackingAlone() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.AttackingTarget != uuid.Nil && game.AttackedAlone(g)
	}
}

// youCastASpellFromYourGraveyard is "whenever you cast a spell from
// your graveyard" (Devoted Grafkeeper): the source's controller cast
// it, the spell's announce record says it came out of a graveyard, and
// the card is the caster's own — a card in a graveyard is in its
// owner's graveyard (CR 404.1), so a spell cast out of an opponent's
// graveyard is not one cast from YOUR graveyard.
func youCastASpellFromYourGraveyard(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	item := g.StackItemForEffect(ev.CardID)
	if item == nil || item.CastFromZone != game.ZoneGraveyard {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	return ok && spell.Owner == source.Controller
}

// shuffleTargetsIntoYourLibrary is Ghostly Castigator's body: the
// target cards still in the graveyard go into their owner's — the
// controller's, since the clause names "your graveyard" — library as
// one move, and that library is shuffled.
func shuffleTargetsIntoYourLibrary(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	controller := item.Controller
	return g.TuckCardsToLibraryThenForEffect(ids, game.TuckOptions{}, func(g *game.Game, _ []uuid.UUID) error {
		return g.ShuffleLibraryForEffect(controller)
	})
}
