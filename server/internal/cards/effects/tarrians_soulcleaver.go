package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tarrian's Soulcleaver — Legendary Artifact — Equipment {1}:
//
//	"Equipped creature has vigilance.
//	 Whenever another artifact or creature is put into a graveyard
//	 from the battlefield, put a +1/+1 counter on equipped creature.
//	 Equip {2}"
//
// "Another" excludes only the Soulcleaver's own instance — the
// equipped creature dying still counts, and so does any OTHER
// Equipment or artifact on the board, which is
// anotherArtifactOrCreaturePutIntoGraveyardFromBattlefield's whole
// shape. EventLTB already means "left the battlefield" (see the doc
// comment on the kind itself), so no separate OldZone check is
// needed; NewZone == ZoneGraveyard is the "into a graveyard" half.
//
// With nothing equipped the counter has nowhere to land and the
// trigger's effect is a quiet no-op, the same posture every other
// "put a counter on equipped creature" card takes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2ffb38ec-5852-4e91-85a5-cfccd1f23556",
		Name:         "Tarrian's Soulcleaver",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			GrantToAttached("vigilance"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return anotherArtifactOrCreaturePutIntoGraveyardFromBattlefield(ev, source, g)
			},
			Key:    "Tarrian's Soulcleaver — put a +1/+1 counter on equipped creature",
			Effect: tarriansSoulcleaverPutCounter,
		}},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}

func tarriansSoulcleaverPutCounter(g *game.Game, item *game.StackItem) error {
	host := attachedHostFor(g, item.SourceCardID)
	if host == nil {
		return nil
	}
	return AddCounter{Target: host.InstanceID, Kind: "+1/+1", N: 1}.Apply(NewContext(g, item))
}
