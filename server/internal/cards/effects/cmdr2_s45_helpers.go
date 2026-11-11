package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s45_helpers.go — batch 45 (Ninjutsu and combat-damage theft).

// cmdr2S45LookExileOneRestToGraveyard is Thief of Sanity's "look at the top
// three cards of that player's library, exile one of them face down, then put
// the rest into their graveyard. You may cast that card for as long as it
// remains exiled, and mana of any type can be spent to cast that spell."
// `from` is the player whose library is looked at, `controller` the looker.
func cmdr2S45LookExileOneRestToGraveyard(g *game.Game, item *game.StackItem, from uuid.UUID, n int, name string) error {
	controller, source := item.Controller, item.SourceCardID
	looked := g.LookAtTopOfPlayersLibraryForEffect(controller, from, n)
	if len(looked) == 0 {
		return nil
	}
	if len(looked) == 1 {
		return cmdr2S45ThiefFinish(g, controller, source, name, looked, looked)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    controller,
		FromPlayer: from,
		Source:     source,
		Question:   name + " — exile one of them face down",
		Cards:      looked,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneLibrary,
		ThenKey:    cmdr2S45ThiefPicked,
		Carry:      looked,
	})
	return nil
}

// cmdr2S45ThiefPicked is the card chosen: the carry is the cards looked at.
// An on-disk key: never renamed, never reused.
var cmdr2S45ThiefPicked = game.RegisterChooseCardsThen("choose-cards/cmdr2-s45-thief-of-sanity", func(g *game.Game, r game.CardsChosen) error {
	return cmdr2S45ThiefFinish(g, r.Chooser, r.Source, "Thief of Sanity", r.Carry, r.Picked)
})

func cmdr2S45ThiefFinish(g *game.Game, controller, source uuid.UUID, name string, looked, picked []uuid.UUID) error {
	var firstErr error
	if len(picked) > 0 {
		firstErr = g.ExileCardsFaceDownWithPermissionForEffect(picked, controller, game.CastPermission{
			AnyColor:   true,
			AnyType:    true,
			CastOnly:   true,
			Duration:   game.WhileInZoneDuration(),
			Source:     source,
			SourceName: name,
			Label:      "You may cast that card for as long as it remains exiled; mana of any type can be spent to cast that spell",
		})
	}
	took := make(map[uuid.UUID]bool, len(picked))
	for _, id := range picked {
		took[id] = true
	}
	var rest []uuid.UUID
	for _, id := range cardsStillInALibrary(g, looked) {
		if !took[id] {
			rest = append(rest, id)
		}
	}
	if len(rest) > 0 {
		if err := g.PutCardsIntoGraveyardKeyedForEffect(rest, game.KeyedTail{}); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// cmdr2S45SpellYouDontOwn is "spells you cast but don't own": the spell is
// cast by the modifier's controller and its owner is someone else (Gonti,
// Canny Acquisitor). Ownership is the card's printed owner, so a spell cast
// from an opponent's library, graveyard or exile counts.
func cmdr2S45SpellYouDontOwn() CostPredicate {
	return func(q game.CostQuery) bool { return q.Card.Owner != q.Controller }
}

// cmdr2S45EquippedCreatureDealtCombatDamageToAPlayer is "whenever equipped
// creature deals combat damage to a player": the Equipment's own trigger,
// the damaged player the event's Target.
func cmdr2S45EquippedCreatureDealtCombatDamageToAPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 {
		return false
	}
	return source.IsAttachedTo(ev.Source) && g.PlayerByIDForEffect(ev.Target) != nil
}

// cmdr2S45ExileTopFaceDownToPlay is "look at the top card of that player's
// library, then exile it face down. You may play that card for as long as it
// remains exiled, and mana of any type can be spent to cast that spell" for
// the trigger's damaged player, the card going to the ability's controller.
func cmdr2S45ExileTopFaceDownToPlay(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	return gontiExileTopForPlay(NewContext(g, item), item.Trigger.Event.Target, item.Controller, 1)
}
