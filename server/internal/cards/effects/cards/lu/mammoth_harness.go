package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mammoth Harness — Enchantment — Aura {3}{G} (Homelands, S82 OLD b825, #5635):
//
//	"Enchant creature
//	 Enchanted creature loses flying.
//	 Whenever enchanted creature blocks or becomes blocked by a creature, the
//	 other creature gains first strike until end of turn."
//
// Losing flying is a layer 6 removal read live off the attachment. The trigger
// is Wooden Stake's per-pair block trigger (CR 509.3b, 509.3d); "the other
// creature" is recorded with its CR 400.7 epoch as the trigger goes on the
// stack, so one that left is not granted anything.
//
// No simplification.
func init() {
	const label = "Mammoth Harness — the other creature gains first strike until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "f2e106c6-7666-42cd-9a02-e51327fb80e8",
		Name:         "Mammoth Harness",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.LoseKeywords(effects.AttachedToSource, "flying")},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBlock},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				_, ok := effects.EnchantedCreatureBlockPairOther(ev, source)
				return ok
			},
			Key: label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				other, _ := effects.EnchantedCreatureBlockPairOther(ev, source)
				item.Params.Object = game.ObjectRef{ID: other, Epoch: g.ObjectEpochForEffect(other)}
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				obj := item.Params.Object
				if obj.ID == uuid.Nil || g.ObjectEpochForEffect(obj.ID) != obj.Epoch {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: obj.ID, Keywords: []string{"first strike"}, Label: label}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
