package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fractured Loyalty — Enchantment — Aura {1}{R} (Mirrodin, #5606):
//
//	"Enchant creature
//	 Whenever enchanted creature becomes the target of a spell or
//	 ability, that spell or ability's controller gains control of that
//	 creature. (This effect lasts indefinitely.)"
//
// The trigger reads EventBecomesTarget, which the engine emits for a
// spell or an ability alike, once per target slot, as the targets are
// chosen (CR 601.2c, 602.2b, 603.3d); its Actor is that spell or
// ability's controller, carried on the item. The trigger goes on the
// stack above the spell or ability, so the new controller has the
// creature before it resolves. "That creature" is the enchanted creature
// as the object it was when it was targeted: one that left the
// battlefield in response, and came back as a new object (CR 400.7), is
// not taken. The control change is a layer 2 effect with no duration
// (CR 613.1b, 611.2a), so it outlasts the Aura. The Aura's own enchant
// target is chosen while it is a spell, before this ability exists.
//
// No simplification.
func init() {
	const label = "Fractured Loyalty — that spell or ability's controller gains control of the creature"
	Register(Spec{
		OracleID:     "ccc168ef-b93b-4bff-89a3-80984d3625de",
		Name:         "Fractured Loyalty",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBecomesTarget},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventBecomesTarget && ev.CardID != uuid.Nil && source.IsAttachedTo(ev.CardID)
			},
			Key: label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				item.Params.Object, _ = g.PermanentRefForEffect(ev.CardID)
				item.Params.Player = ev.Actor
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ref := item.Params.Object
				if info, ok := g.PermanentForEffect(ref); !ok || info.Left || g.PlayerByIDForEffect(item.Params.Player) == nil {
					return nil
				}
				return GainControl{Target: ref.ID, Controller: item.Params.Player, Duration: game.IndefiniteDuration(),
					Label: "Fractured Loyalty — gains control"}.Apply(NewContext(g, item))
			},
		}},
	})
}
