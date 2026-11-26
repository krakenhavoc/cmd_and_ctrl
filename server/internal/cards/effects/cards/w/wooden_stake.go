package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wooden Stake — {2} Artifact — Equipment (Innistrad, S82 OLD b448, #5572):
//
//	"Equipped creature gets +1/+0.
//	 Whenever equipped creature blocks or becomes blocked by a Vampire, destroy
//	 that creature. It can't be regenerated.
//	 Equip {1}"
//
// Barrow-Blade's per-pair block trigger (CR 509.3b, 509.3d), restricted to a
// Vampire on the other side of the pair (read as the block is declared, through
// the layers). "That creature" is the Vampire, recorded as the trigger goes on
// the stack with its CR 400.7 epoch, so one that left is not destroyed; the
// destruction can't be regenerated (CR 701.19c).
//
// No simplification.
func init() {
	const label = "Wooden Stake — destroy that Vampire; it can't be regenerated"
	effects.Register(effects.Spec{
		OracleID:     "1b26dbb7-0ae5-4279-87b6-1b722c30a386",
		Name:         "Wooden Stake",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 0)},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBlock},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				other, ok := woodenStakeOther(ev, source)
				if !ok {
					return false
				}
				c, found := g.LookupCardForEffect(other)
				return found && c.HasSubtype(source.TextSubtype("Vampire"))
			},
			Key: label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				other, _ := woodenStakeOther(ev, source)
				item.Params.Object = game.ObjectRef{ID: other, Epoch: g.ObjectEpochForEffect(other)}
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				obj := item.Params.Object
				if obj.ID == uuid.Nil || g.ObjectEpochForEffect(obj.ID) != obj.Epoch {
					return nil
				}
				return effects.DestroyTarget{Target: obj.ID, CantBeRegenerated: true}.Apply(effects.NewContext(g, item))
			},
		}},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}

// woodenStakeOther is the other creature of a blocker-attacker pair when the
// equipped creature is one of them.
func woodenStakeOther(ev game.Event, source *game.Card) (uuid.UUID, bool) {
	if ev.Target == uuid.Nil {
		return uuid.Nil, false
	}
	switch {
	case source.IsAttachedTo(ev.CardID):
		return ev.Target, true
	case source.IsAttachedTo(ev.Target):
		return ev.CardID, true
	}
	return uuid.Nil, false
}
