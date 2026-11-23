package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghost Lantern // Bind Spirit — an adventure card (Commander Legends: Battle
// for Baldur's Gate, #5948):
//
//	Ghost Lantern {B}, Artifact — Equipment:
//	"Whenever a creature you control dies, put a +1/+1 counter on equipped
//	 creature.
//	 Equip {1}"
//
//	Bind Spirit {1}{B}, Instant — Adventure:
//	"Return target creature card from your graveyard to your hand."
//
// The adventure layout is Brazen Borrower's: the Equipment under the bare
// oracle ID, the Adventure half under "<oracle_id>#1", and the engine exiles
// the resolved Adventure with permission to cast the Equipment (CR 715.3d).
// The trigger is Blade of the Bloodchief's, narrowed to a creature of the
// Lantern's controller (read as it last existed, CR 603.10a); the host is read
// as the trigger resolves, so an unattached Lantern, or a host that has left,
// gets nothing. The equipped creature's own death counts but there is nothing
// left to put the counter on.
//
// No simplification.
func init() {
	const oracleID = "4af03564-1d1c-4582-9248-cfe9551bad30"

	effects.Register(effects.Spec{
		OracleID:     oracleID,
		Name:         "Ghost Lantern",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverACreatureYouControlDies("Ghost Lantern — put a +1/+1 counter on equipped creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					host, ok := effects.Cmdr2B21EquippedCreature(g, ctx.Source())
					if !ok {
						return nil
					}
					return effects.AddCounter{Target: host.InstanceID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})

	effects.Register(effects.Spec{
		OracleID:     oracleID + "#1",
		Name:         "Bind Spirit",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			return effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}.Apply(ctx)
		},
	})
}
