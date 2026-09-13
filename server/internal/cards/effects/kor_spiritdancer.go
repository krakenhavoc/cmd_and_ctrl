package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kor Spiritdancer — Creature — Kor Wizard {1}{W}, 0/2 (EDHREC rank
// 1955):
//
//	"This creature gets +2/+2 for each Aura attached to it.
//	 Whenever you cast an Aura spell, you may draw a card."
//
// The Aura deck's two-drop. The pump is a self-only layer 7c static
// that counts the Auras whose AttachedTo points at the Spiritdancer
// (#379's relation) at every recompute — an attach or unattach bumps
// the layer version, so the body follows the Auras. The draw is a
// cast trigger on any Aura spell the controller casts, read off the
// stack where the spell sits, with the "may" as a real prompt.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2bfc7467-1316-43d3-9ca6-ff36cc2de607",
		Name:         "Kor Spiritdancer",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, target *game.Card, g *game.Game, _ *game.Card) {
				n := b18AurasAttachedTo(g, target.InstanceID)
				c.Power += 2 * n
				c.Toughness += 2 * n
			},
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b18SpellCastByYouIsAura(ev, source, g)
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Kor Spiritdancer — draw a card?"},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Kor Spiritdancer — draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
