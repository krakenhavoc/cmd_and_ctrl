package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fear of Immobility — Enchantment Creature — Nightmare {4}{W}, 4/4
// (Duskmourn: House of Horror, #4466):
//
//	"When this creature enters, tap up to one target creature. If an
//	 opponent controls that creature, put a stun counter on it. (If a
//	 permanent with a stun counter would become untapped, remove one from
//	 it instead.)"
//
// Meat Locker's "tap up to one target creature" enters trigger, with the
// stun counter conditional: who controls the creature is read as the
// trigger resolves, after the tap, so your own creature is only tapped. A
// target gone by resolution gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fb2600a8-0ba6-4888-ac8a-f206469d4c71",
		Name:         "Fear of Immobility",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Fear of Immobility — tap up to one target creature; if an opponent controls it, stun it",
				fearOfImmobilityTap), TargetCreature("up to one target creature").WithCount(0, 1)),
		},
	})
}

// fearOfImmobilityTap taps the still-legal target and stuns it when its
// controller is not the trigger's controller.
func fearOfImmobilityTap(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ts := ctx.LegalTargets()
	if len(ts) == 0 || ts[0].Kind != game.TargetCard {
		return nil
	}
	if err := (TapTarget{Target: ts[0].ID}).Apply(ctx); err != nil {
		return err
	}
	c, ok := g.LookupCardForEffect(ts[0].ID)
	if !ok || c.Controller == item.Controller {
		return nil
	}
	return AddCounter{Target: ts[0].ID, Kind: game.CounterStun, N: 1}.Apply(ctx)
}
