package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Queen of Ice // Rage of Winter — Creature — Human Noble Wizard {2}{U},
// 2/3 // Sorcery — Adventure {1}{U} (Throne of Eldraine, #5516):
//
//	Queen of Ice — "Whenever this creature deals combat damage to a
//	 creature, tap that creature. It doesn't untap during its controller's
//	 next untap step."
//	Rage of Winter — "Tap target creature. It doesn't untap during its
//	 controller's next untap step."
//
// Both faces register, the Adventure under "<oracle_id>#1". The trigger is
// Voracious Cobra's shape: it names the damaged creature as the object it was
// when the damage was dealt (CR 510.2), so one that has left is not tapped.
// The Adventure lifecycle is the engine's.
func init() {
	const oracle = "f5511802-56fe-4922-83d2-ae723e1beab7"
	const label = "Queen of Ice — tap that creature; it doesn't untap during its controller's next untap step"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Queen of Ice",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventDealDamage},
			AppliesTo: effects.ThisDealtCombatDamageToACreature,
			Key:       label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				item.Params.Object = effects.DamagedCreatureRef(g, ev.Target)
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ref := item.Params.Object
				c, ok := g.LookupCardForEffect(ref.ID)
				if !ok || c.ObjectEpoch != ref.Epoch {
					return nil
				}
				if z := g.FindCardZoneForEffect(ref.ID); z == nil || z.Kind != game.ZoneBattlefield {
					return nil
				}
				return effects.TapAndFreeze{Targets: []uuid.UUID{ref.ID}, Label: label}.Apply(effects.NewContext(g, item))
			},
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Rage of Winter",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.TapAndFreeze{Targets: []uuid.UUID{id}, Label: "Rage of Winter — it doesn't untap during its controller's next untap step"}.Apply(ctx)
		},
	})
}
