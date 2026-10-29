package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Insidious Roots — Enchantment {B}{G}:
//
//	"Creature tokens you control have "{T}: Add one mana of any
//	 color."
//	 Whenever one or more creature cards leave your graveyard, create
//	 a 0/1 green Plant creature token, then put a +1/+1 counter on
//	 each Plant you control."
//
// The grant is ADR 0093's layer-6 grant; the trigger is Desecrated
// Tomb's batch trigger ("one or more … leave", OncePerBatch). The
// counters go on every Plant you control, the new one included, and
// only after the token has really entered — a creation can pause on a
// CR 616 prompt, so the counters ride the creation's continuation.
//
// No simplification.
const insidiousRootsGrant = "insidious-roots/any-color"

func init() {
	Register(Spec{
		OracleID:     "d75b2b8e-05c9-47da-b359-a867256d78ea",
		Name:         "Insidious Roots",
		Completeness: CompletenessFull,
		Grants:       []AbilityGrant{AnyColorManaGrant(insidiousRootsGrant)},
		Static:       []game.StaticAbility{Labelled(GrantAbilities(creatureTokensYouControl, insidiousRootsGrant), "Creature tokens you control have \"{T}: Add one mana of any color.\"")},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventZoneMove,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b02CreatureCardLeftYourGraveyard(ev, source, g)
				},
				"Insidious Roots — create a 0/1 Plant, then a +1/+1 counter on each Plant",
				insidiousRootsGrow)),
		},
	})
}

// insidiousRootsGrow is the trigger's resolution.
func insidiousRootsGrow(g *game.Game, item *game.StackItem) error {
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: item.Controller,
		Source:     item.SourceCardID,
		Groups:     []game.TokenGroup{{Template: g.TextReaderForEffect(item).TokenThroughText(TokenCard("0/1 green Plant")), Count: 1}},
	}, NewContext(g, item).ItemTail(insidiousRootsCounters))
}

// insidiousRootsCounters is "then put a +1/+1 counter on each Plant you
// control" as a keyed tail (#4253 slice 6): "Plant" is a word of the
// source's text, which the asking item that rides the tail gives back. An
// on-disk key.
var insidiousRootsCounters = CardTail("insidious-roots/counter-on-each-plant", func(ctx *Context, _ game.TailOutcome) error {
	g, controller := ctx.Game, ctx.Controller()
	plant := g.TextReaderForEffect(ctx.Item).TextSubtype("Plant")
	member := ctx.asGroupMember()
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller || !c.HasSubtype(plant) {
			continue
		}
		if err := (AddCounter{Target: c.InstanceID, Kind: game.CounterPlusOne, N: 1}).Apply(member); err != nil {
			return err
		}
	}
	return nil
})
