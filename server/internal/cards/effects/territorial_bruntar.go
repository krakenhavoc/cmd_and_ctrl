package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Territorial Bruntar — Creature — Beast {4}{R}{R}, 6/6 (Edge of
// Eternities, #4461):
//
//	"Reach
//	 Landfall — Whenever a land you control enters, exile cards from the
//	 top of your library until you exile a nonland card. You may cast that
//	 card this turn."
//
// Exploratory Mission's landfall walk (MillToZone until a nonland card)
// without its counter and mana: the keyed tail grants a cast permission
// for this turn over "that card", the nonland card, alone, and the lands
// exiled on the way stay in exile. The permission is cast-only and keeps
// every cost and timing rule, as the ruling says: the card's mana cost is
// paid, and a sorcery is cast only in a main phase with the stack empty.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "03f04c2c-1036-497e-8b13-06c61b0b0ba6",
		Name:            "Territorial Bruntar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			Landfall("Territorial Bruntar — landfall: exile until a nonland card, which you may cast this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return MillToZone{
						Player:   ctx.Controller(),
						To:       game.ZoneExile,
						Until:    UntilNonland,
						ThenTail: ctx.ItemTail(territorialBruntarExiled),
					}.Apply(ctx)
				}),
		},
	})
}

// territorialBruntarExiled grants the cast permission over the nonland
// card the walk ended on, if it is still in exile. An on-disk key: never
// renamed, never reused.
var territorialBruntarExiled = CardTail("territorial-bruntar/cast-nonland-this-turn", func(ctx *Context, o game.TailOutcome) error {
	var nonland uuid.UUID
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		z := ctx.Game.FindCardZoneForEffect(id)
		if ok && !c.IsLand() && z != nil && z.Kind == game.ZoneExile {
			nonland = id
		}
	}
	if nonland == uuid.Nil {
		return nil
	}
	c, _ := ctx.Game.LookupCardForEffect(nonland)
	perm := ExileTopWithPermission{CastOnly: true}.permission()
	perm.Player = ctx.Controller()
	perm.Zone = game.ZoneExile
	ctx.Game.GrantCastPermissionToCardsForEffect(perm, []game.Card{c})
	return nil
})
