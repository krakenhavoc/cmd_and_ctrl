package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grolnok, the Omnivore — Legendary Creature — Frog {2}{G}{U}, 3/3
// (Innistrad: Crimson Vow):
//
//	"Whenever a Frog you control attacks, mill three cards.
//	 Whenever a permanent card is put into your graveyard from your
//	 library, exile it with a croak counter on it.
//	 You may play lands and cast spells from among cards you own in exile
//	 with croak counters on them."
//
// Three abilities. The attack trigger reads "Frog" as a word of the
// source's text (CR 612.2) and is Grolnok's own attack too. The second is
// one trigger for a permanent card (not an instant or sorcery) of yours
// reaching your graveyard from your library by a mill or any other move;
// the card is exiled only if it is still in the graveyard when the trigger
// resolves, and the croak counter lands once it has reached exile. The
// third is a standing permission over cards its controller owns in exile
// with a croak counter, a land played with the land drop and a spell cast
// at its ordinary timing for its ordinary cost; the counter is per object,
// so a card that leaves exile and comes back carries none.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "53029460-50a0-4934-8fa9-0271517cd5d0",
		Name:         "Grolnok, the Omnivore",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, grolnokAFrogAttacked,
				"Grolnok, the Omnivore — mill three cards", Do(MillCards{N: 3})),
			OnAny([]game.EventKind{game.EventMill, game.EventZoneMove}, grolnokPermanentCardMilled,
				"Grolnok, the Omnivore — exile it with a croak counter on it", grolnokExileWithCroak),
		},
		CastPermissions: []game.CastPermission{{
			Zone:   game.ZoneExile,
			Filter: game.PermissionFilter{OwnedByHolder: true, WithCounter: grolnokCroakCounter},
			Label:  "Play a land or cast a spell from among cards you own in exile with croak counters on them (Grolnok, the Omnivore)",
		}},
	})
}

const grolnokCroakCounter = "croak"

func grolnokAFrogAttacked(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !attackDeclaredByYou(ev, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.HasSubtype(source.TextSubtype("Frog"))
}

func grolnokPermanentCardMilled(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.OldZone != game.ZoneLibrary || ev.NewZone != game.ZoneGraveyard {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Owner == source.Controller && c.IsPermanent()
}

func grolnokExileWithCroak(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	id := item.Trigger.Event.CardID
	if !discardedCardStillInGraveyard(g, item, id) {
		return nil
	}
	return ExileTarget{Target: id, ThenTail: NewContext(g, item).ItemTail(grolnokCroaked)}.Apply(NewContext(g, item))
}

// grolnokCroaked puts the croak counter on the card once it is in exile. An
// on-disk key: never renamed, never reused.
var grolnokCroaked = CardTail("grolnok/croak-counter", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) != 1 || o.IDs[0] == uuid.Nil {
		return nil
	}
	return ctx.Game.AddCounterForEffect(o.IDs[0], grolnokCroakCounter, 1)
})
