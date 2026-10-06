package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lord of the Nazgûl — Legendary Creature — Wraith Noble {3}{U}{B},
// 4/3:
//
//	"Flying
//	 Wraiths you control have protection from Ring-bearers.
//	 Whenever you cast an instant or sorcery spell, create a 3/3 black
//	 Wraith creature token with menace. Then if you control nine or
//	 more Wraiths, Wraiths you control have base power and toughness
//	 9/9 until end of turn."
//
// The protection is a layer-6 grant to every Wraith its controller
// has, the Lord included, of the Ring-bearer quality (#2145, CR
// 701.54e): a source matches while it is its controller's Ring-bearer
// when the check is made, so a Ring-bearer can't block, target, damage
// or enchant a Wraith.
//
// The count is read AFTER the token is made ("Then if you control nine
// or more"), and the 9/9 set is the Wraiths that exist as the ability
// resolves, locked then (CR 611.2c) — a Wraith that arrives later this
// turn is not 9/9.
//
// No simplification.
func init() {
	wraiths := TribeFilter{Tribes: []string{"Wraith"}, YoursOnly: true}
	Register(Spec{
		OracleID:        "2e94bfb2-9f7d-43af-8995-6cd5ffb15b21",
		Name:            "Lord of the Nazgûl",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			TribalKeywordGrant(wraiths, "protection from Ring-bearers"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return instantOrSorceryCastByYou(ev, source, g)
			}, "Lord of the Nazgûl — create a 3/3 black Wraith with menace", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (CreateToken{Controller: item.Controller, Template: TokenCard("3/3 black Wraith with menace"), N: 1}).Apply(ctx); err != nil {
					return err
				}
				isWraith := func(c game.Card) bool { return c.HasSubtype("Wraith") }
				if countControlled(g, item.Controller, isWraith) < 9 {
					return nil
				}
				return untilEndOfTurn(ctx, uuid.Nil, And(Subtype("Wraith"), ControlledBy(item.Controller)),
					"Lord of the Nazgûl — Wraiths have base power and toughness 9/9",
					game.SetBasePowerMod(9), game.SetBaseToughnessMod(9))
			}),
		},
	})
}
