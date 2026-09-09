package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Corsair Captain — 2/2 Creature — Human Pirate for {2}{U}:
//
//	"When this creature enters, create a Treasure token.
//	 Other Pirates you control get +1/+1."
//
// The deck's lord. Two ability types on one card: an ETB trigger and
// a Layer 7c static, which is the first time the catalog pairs them.
//
// "Other Pirates" excludes the Captain itself — the static's
// AppliesTo checks the subtype and then rules out the source, which
// is why it can't reuse the Glorious Anthem predicate (the Anthem
// isn't a creature, so it never had to exclude itself).
func init() {
	Register(Spec{
		OracleID: "a7ec13c6-7ade-433a-b5a2-047854eef486",
		Name:     "Corsair Captain",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Corsair Captain — create a Treasure",
					func(g *game.Game, item *game.StackItem) error {
						return CreateToken{
							Controller: item.Controller,
							Template:   TreasureToken(),
							N:          1,
						}.Apply(NewContext(g, item))
					})
			},
		}},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				if target.InstanceID == source.InstanceID {
					return false // "OTHER Pirates"
				}
				return target.Controller == source.Controller && isPirate(*target)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
				c.Toughness++
			},
		}},
	})
}

// isPirate reports whether a card is a Pirate creature.
//
// It reads the PRINTED type line rather than Effective().Subtypes,
// because this runs inside a Layer 6/7 AppliesTo predicate — the
// target's effective characteristics are mid-rebuild at that point,
// so asking for them recurses into the computation being performed.
// The cost is that a creature turned into a Pirate by another effect
// isn't counted by the lord; that wants the layer engine to expose a
// partially-applied view, which it doesn't today.
func isPirate(c game.Card) bool {
	return c.IsCreature() && containsFoldASCII(c.TypeLine, "pirate")
}
