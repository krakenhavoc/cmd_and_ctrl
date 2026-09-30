package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Season of Growth — Enchantment {1}{G} (EDHREC rank 2871):
//
//	"Whenever a creature you control enters, scry 1.
//	 Whenever you cast a spell that targets a creature you control,
//	 draw a card."
//
// The first ability is Impact Tremors' landfall-style shape:
// enteredUnderYourControl reads the entering permanent off the ETB
// event and checks it is a creature the controller controls.
//
// The second reads the spell's OWN announced target list off the
// stack (g.StackItemForEffect(ev.CardID).Targets), the same route
// Hydroid Krasis reads its X off — the item is already on the stack
// by the time EventCast fires, targets included, so this is one
// trigger per SPELL CAST (as printed), not one per target slot the
// way "becomes the target" phrasing (Monk Gyatso) would be. A spell
// that targets two creatures the controller controls still draws
// only one card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3f4e300a-ec5c-42f3-a97b-d58e62abe22b",
		Name:         "Season of Growth",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature()
			}, "Season of Growth — scry 1", Do(Scry{N: 1})),
			On(game.EventCast, seasonOfGrowthTargetsYourCreature,
				"Season of Growth — draw a card", Do(DrawCards{N: 1})),
		},
	})
}

// seasonOfGrowthTargetsYourCreature is "you cast a spell that targets
// a creature you control" — read off the announced spell's own
// target list, once per cast.
func seasonOfGrowthTargetsYourCreature(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Actor != source.Controller {
		return false
	}
	item := g.StackItemForEffect(ev.CardID)
	if item == nil {
		return false
	}
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := g.LookupCardForEffect(t.ID)
		if ok && c.IsCreature() && c.Controller == source.Controller {
			return true
		}
	}
	return false
}
