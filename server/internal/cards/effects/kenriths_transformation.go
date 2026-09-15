package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kenrith's Transformation — Enchantment — Aura for {1}{G}:
//
//	"Enchant creature
//	 When this Aura enters, draw a card.
//	 Enchanted creature loses all abilities and is a green Elk
//	 creature with base power and toughness 3/3. (It loses all other
//	 card types and creature types.)"
//
// Green's answer to a commander, and the reason it is played over
// Beast Within is the replacement card: a two-mana removal spell
// that cantrips leaves you up on cards even when the Elk survives.
// The 3/3 body it leaves behind is the price — this is the one of
// the three that hands the opponent something that can still attack.
//
// Layers, in printed order:
//
//   - layer 4 — "is a green Elk creature", replacing card types and
//     creature types. Still a creature, so unlike Song of the Dryads
//     this does NOT make the Equipment and Auras already on the host
//     fall off.
//   - layer 5 — "green". The catalog's first colour-changing effect;
//     layer 5 has had a bucket in the recompute since S16 with
//     nothing to put in it.
//   - layer 6 — "loses all abilities", with nothing kept.
//   - layer 7b — base 3/3, so counters and anthems still apply.
//
// The cantrip is a triggered ability on the stack: "When this Aura
// enters, draw a card" IS a trigger, and the draw can be responded to
// between the Aura entering and the card being drawn. It ran from the
// direct entry hook until #578 moved every "When ~ enters" onto
// Triggered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a492a323-df8a-40fd-bff0-50091baa7700",
		Name:         "Kenrith's Transformation",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Kenrith's Transformation — draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
		Static: []game.StaticAbility{
			SetAttachedTypes([]string{"Creature"}, []string{"Elk"}),
			SetAttachedColors("G"),
			LoseAllAbilities(),
			SetAttachedBasePT(3, 3),
		},
	})
}
