package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Drana, the Last Bloodchief — Legendary Creature — Vampire Cleric {3}{B}{B},
// 4/4 (Secret Lair Drop):
//
//	"Flying
//	 Whenever Drana attacks, defending player chooses a nonlegendary creature
//	 card in your graveyard. You return that card to the battlefield with an
//	 additional +1/+1 counter on it. The creature is a Vampire in addition to
//	 its other types."
//
// "Defending player" is the player attacked, or the controller of the
// planeswalker or battle attacked (CR 506.3d). That player is the chooser of
// a keyed choose-cards prompt over Drana's controller's graveyard (Bane of
// Bala Ged's shape, the picked cards being the controller's), so a table
// waiting on it is a restore point. With no nonlegendary creature card there
// is nothing to ask. The picked card enters under Drana's controller's
// control with one +1/+1 counter and the Vampire subtype (Shilgengar's
// EntersAs); a card that left the graveyard meanwhile stays put.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3ae3b901-4d36-40e2-b861-26209fa1e823",
		Name:            "Drana, the Last Bloodchief",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.ThisAttacked,
				"Drana — defending player chooses a nonlegendary creature card in your graveyard; you return it with an additional +1/+1 counter, as a Vampire",
				dranaAsk),
		},
	})
}

func dranaAsk(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	defender := effects.DefendingPlayerOfAttack(g, item.Trigger.Event)
	p := g.PlayerByIDForEffect(item.Controller)
	if defender == uuid.Nil || p == nil || p.Graveyard == nil {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		if c.IsCreature() && !c.IsLegendary() {
			ids = append(ids, c.InstanceID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    defender,
		FromPlayer: item.Controller,
		Source:     item.SourceCardID,
		Question:   "Drana, the Last Bloodchief — choose a nonlegendary creature card in the attacker's graveyard; they return it with an additional +1/+1 counter",
		Cards:      ids,
		Min:        1,
		Max:        1,
		ThenKey:    dranaReturn,
		LiveZone:   game.ZoneGraveyard,
	})
	return nil
}

// dranaReturn is the keyed answer. An on-disk key: never renamed, never reused.
var dranaReturn = game.RegisterChooseCardsThen("choose-cards/drana-the-last-bloodchief-return", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	ctx := effects.KeyedContext(g, r.From, r.Source)
	return effects.PutCardsFromZoneTogetherWith(ctx, game.ZoneGraveyard, r.Picked[:1], game.ZoneEntryOptions{
		Controller:         r.From,
		EntersWithCounters: map[string]int{game.CounterPlusOne: 1},
		EntersAs: &game.EntersAs{
			AddSubtypes: ctx.TextSubtypes("Vampire"),
			Label:       "Drana — a Vampire in addition to its other types",
		},
	}, game.TailRef{}, 0)
})
