package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hateful Eidolon — Enchantment Creature — Spirit {B}, 1/2 (Secrets of
// Strixhaven Commander, #4834):
//
//	"Lifelink
//	 Whenever an enchanted creature dies, draw a card for each Aura you
//	 controlled that was attached to it."
//
// Any player's creature counts, and it is "enchanted" as it last existed
// on the battlefield (CR 603.10a): an Aura was attached to it as it died.
// The draw counts the Auras attached to it then whose controller, then,
// was this ability's controller (socB6AurasOnTheDead), so another
// player's Aura on it triggers the ability but draws nothing for it, and
// an Aura destroyed at the same time as the creature still counts, as the
// rulings say. The Eidolon's own death beside an enchanted creature still
// triggers it (a leaves-the-battlefield look-back), itself included when
// it was enchanted.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a3b39590-34b7-4fd0-877f-748e6df30e88",
		Name:            "Hateful Eidolon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				if _, died := diedCreature(ev, g); !died {
					return false
				}
				return len(socB6AurasOnTheDead(g, ev)) > 0
			}, "Hateful Eidolon — draw a card for each Aura you controlled that was attached to it",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					n := 0
					for _, controller := range socB6AurasOnTheDead(g, item.Trigger.Event) {
						if controller == item.Controller {
							n++
						}
					}
					if n == 0 {
						return nil
					}
					return DrawCards{N: n}.Apply(NewContext(g, item))
				}),
		},
	})
}
