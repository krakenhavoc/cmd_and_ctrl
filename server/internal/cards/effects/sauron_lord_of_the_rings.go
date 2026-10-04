package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sauron, Lord of the Rings — Legendary Creature — Avatar Horror
// {5}{U}{B}{R}, 9/9:
//
//	"When you cast this spell, amass Orcs 5, mill five cards, then
//	 return a creature card from your graveyard to the battlefield.
//	 Trample
//	 Whenever a commander an opponent controls dies, the Ring tempts
//	 you."
//
// The cast trigger is a FromStack ability (cascade's mechanism,
// Kozilek's shape): it fires as the spell is cast, is controlled by the
// caster, and resolves above Sauron, so a countered Sauron still amasses
// and still returns a creature. The creature card is chosen from the
// whole graveyard after the mill — "The creature card you choose
// doesn't need to be a card you milled" (2023-06-16 ruling) — and the
// return is not optional when there is one.
//
// "Whenever a commander an opponent controls dies" reads the commander's
// controller as it left the battlefield (CR 603.10a), so a commander of
// yours that an opponent stole counts, and a noncreature commander
// counts too.
//
// A commander dies like any creature (CR 903.9a, ADR 0115): the trigger
// sees it die "even if the owner of the commander that died chooses to
// return it to the command zone after it dies" (the card's ruling),
// because the owner is only asked once it is in the graveyard.
func init() {
	Register(Spec{
		OracleID:        "69c674c7-48a5-49c8-b0be-3f2b5c6a548c",
		Name:            "Sauron, Lord of the Rings",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			{
				FromStack: true,
				Watches:   []game.EventKind{game.EventCast},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.CardID == source.InstanceID
				},
				Key: sauronLordOfTheRingsCastLabel,
				// The cast trigger's controller is the CASTER (ev.Actor),
				// as Kozilek's is: a fill-in Build beside the declared
				// Effect (ADR 0041 P9).
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, sauronLordOfTheRingsCastLabel)
					item.Controller, item.Owner = ev.Actor, ev.Actor
					return item
				},
				Effect: sauronLordOfTheRingsCast,
			},
			On(game.EventLTB, aCommanderAnOpponentControlsDied, "Sauron, Lord of the Rings — the Ring tempts you", Do(TheRingTemptsYou{})),
		},
	})
}

const sauronLordOfTheRingsCastLabel = "Sauron, Lord of the Rings — amass Orcs 5, mill five cards, then return a creature card from your graveyard to the battlefield"

// sauronLordOfTheRingsCast is the cast trigger's body, in printed order:
// the amass, then the mill, then the return.
func sauronLordOfTheRingsCast(g *game.Game, item *game.StackItem) error {
	return Amass{Subtype: "Orc", N: 5, Then: func(ctx *Context, _ uuid.UUID) error {
		return roomsBMillThenReturnAChosenCreatureCard(5)(ctx.Game, ctx.Item)
	}}.Apply(NewContext(g, item))
}

// aCommanderAnOpponentControlsDied is "whenever a commander an opponent
// controls dies": a commander went from the battlefield to a graveyard
// (CR 700.4) while a player other than this ability's controller
// controlled it.
func aCommanderAnOpponentControlsDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !c.IsCommander {
		return false
	}
	controller := leftUnderControlOf(ev, c)
	return controller != uuid.Nil && controller != source.Controller
}
