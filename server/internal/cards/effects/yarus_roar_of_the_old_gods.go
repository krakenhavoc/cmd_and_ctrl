package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Yarus, Roar of the Old Gods — Legendary Creature — Centaur Druid
// {2}{R}{G} (Murders at Karlov Manor, #4469):
//
//	"Other creatures you control have haste.
//	 Whenever one or more face-down creatures you control deal combat
//	 damage to a player, draw a card.
//	 Whenever a face-down creature you control dies, return it to the
//	 battlefield face down under its owner's control if it's a permanent
//	 card, then turn it face up."
//
// The draw is the "one or more" batch, once for each player they hit in a
// combat damage step (CR 603.2c, the ruling). The dies trigger reads
// whether the creature was face down as it left (Event.WasFaceDown, the
// last-known stamp: the card in the graveyard is face up, CR 708.9), so a
// creature that was face up never triggers it, and it triggers for each
// face-down creature that dies together with Yarus. It returns "it": the
// card must still be in the graveyard as the ability resolves (the
// ruling), be a permanent card (an instant or sorcery stays), and comes
// back under its OWNER's control, face down (CR 708.2: no enters trigger),
// then is turned face up (CR 708.8: it did not enter again, so no enters
// trigger fires either). One that cannot be turned face up stays face down.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "30fd6171-3e57-4c05-a86b-e1c0479ba236",
		Name:         "Yarus, Roar of the Old Gods",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{YoursOnly: true, Others: true}, "haste"),
		},
		Triggered: []game.TriggeredAbility{
			WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(
				func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.FaceDownIsPermanent() },
				"Yarus — draw a card", Do(DrawCards{N: 1})),
			On(game.EventLTB, yarusFaceDownCreatureDied,
				"Yarus — return it to the battlefield face down under its owner's control, then turn it face up",
				yarusReturn),
		},
	})
}

// yarusFaceDownCreatureDied is "a face-down creature you control dies": a
// creature put into a graveyard from the battlefield that was face down, and
// under the source's controller, as it last existed.
func yarusFaceDownCreatureDied(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
		return false
	}
	if was, known := ev.WasFaceDown(); !known || !was {
		return false
	}
	if was, known := ev.WasType("creature"); !known || !was {
		return false
	}
	controller, known := ev.LeftUnderControlOf()
	return known && controller == source.Controller
}

func yarusReturn(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	id := item.Trigger.Event.CardID
	z := g.FindCardZoneForEffect(id)
	c, ok := g.LookupCardForEffect(id)
	if z == nil || z.Kind != game.ZoneGraveyard || !ok || !c.IsPermanent() {
		return nil
	}
	ctx := NewContext(g, item)
	return g.PutOntoBattlefieldTogetherKeyedForEffect([]game.BatchEntry{{CardID: id, From: game.ZoneGraveyard}},
		game.ZoneEntryOptions{Controller: c.Owner, FaceDown: game.FaceDownTurned}, ctx.ItemTail(yarusTurnFaceUp))
}

// yarusTurnFaceUp is "then turn it face up", run once the card has entered.
var yarusTurnFaceUp = CardTail("tail/yarus-turn-face-up", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		ctx.Game.TurnFaceUpForEffect(ctx.Source(), ctx.Controller(), id)
	}
	return nil
})
