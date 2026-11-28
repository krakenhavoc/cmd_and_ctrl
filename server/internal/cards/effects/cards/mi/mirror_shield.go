package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mirror Shield — Artifact — Equipment {2} (Theros Beyond Death, #5514):
//
//	"Equipped creature gets +0/+2 and has hexproof and "Whenever a creature
//	 with deathtouch blocks or becomes blocked by this creature, destroy
//	 that creature."
//	 Equip {2}"
//
// Dead-Iron Sledge's per-pair block trigger (CR 509.3b, 509.3d), read off the
// Equipment: one trigger for each blocker-attacker pair that names the
// equipped creature, fired only when the other creature has deathtouch as it
// stands (after every effect). The destruction reads the other creature as the
// trigger resolves; one that has left is not destroyed. The pump and hexproof
// are the Equipment's statics (CR 301.5).
//
// No simplification.
func init() {
	const label = "Mirror Shield — destroy that creature"
	effects.Register(effects.Spec{
		OracleID:     "57f2858e-5793-41fc-9bbb-1fdd3b99fe79",
		Name:         "Mirror Shield",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(0, 2),
			effects.GrantToAttached("hexproof"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBlock},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				other := mirrorShieldOther(ev, source)
				if other == uuid.Nil {
					return false
				}
				c, ok := g.LookupCardForEffect(other)
				return ok && game.HasKeyword(&c, "deathtouch")
			},
			Key: label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				other := mirrorShieldOther(ev, source)
				item.Params.Object = game.ObjectRef{ID: other, Epoch: g.ObjectEpochForEffect(other)}
				return item
			},
			Effect: effects.DestroyTheObjectNamedByTheTrigger,
		}},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}

// mirrorShieldOther is the other creature of the blocker-attacker pair the
// equipped creature is in, or the zero ID when it is in neither.
func mirrorShieldOther(ev game.Event, source *game.Card) uuid.UUID {
	switch {
	case ev.Target == uuid.Nil:
		return uuid.Nil
	case source.IsAttachedTo(ev.CardID):
		return ev.Target
	case source.IsAttachedTo(ev.Target):
		return ev.CardID
	}
	return uuid.Nil
}
