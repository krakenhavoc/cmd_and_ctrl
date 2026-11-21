package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Barrow-Blade — Artifact — Equipment {1} (The Lord of the Rings: Tales of
// Middle-earth):
//
//	"Equipped creature gets +1/+1.
//	 Whenever equipped creature blocks or becomes blocked by a creature,
//	 that creature loses all abilities until end of turn.
//	 Equip {1}"
//
// Dead-Iron Sledge's per-pair block trigger (CR 509.3b, 509.3d): one
// trigger for each blocker-attacker pair the equipped creature is in. "That
// creature" is the other one of the pair, recorded as the trigger goes on
// the stack (a fill-in Build, Params.Object with its CR 400.7 epoch), so an
// Equipment that is moved in response still strips the right creature, and
// one that left the battlefield is not stripped. The loss of abilities is a
// layer-6 scoped effect on that object until the end of the turn
// (CR 514.2).
//
// No simplification.
func init() {
	const label = "Barrow-Blade — that creature loses all abilities until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "867c5592-50d5-420c-ac71-50014bc390eb",
		Name:         "Barrow-Blade",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBlock},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Target != uuid.Nil && (source.IsAttachedTo(ev.CardID) || source.IsAttachedTo(ev.Target))
			},
			Key: label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				other := ev.Target
				if source.IsAttachedTo(ev.Target) {
					other = ev.CardID
				}
				item.Params.Object = game.ObjectRef{ID: other, Epoch: g.ObjectEpochForEffect(other)}
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				obj := item.Params.Object
				if obj.ID == uuid.Nil || g.ObjectEpochForEffect(obj.ID) != obj.Epoch {
					return nil
				}
				return effects.UntilEndOfTurn(effects.NewContext(g, item), obj.ID, nil, label, game.LoseAllAbilitiesMod())
			},
		}},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
