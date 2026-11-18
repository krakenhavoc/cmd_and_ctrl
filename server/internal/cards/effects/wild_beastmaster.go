package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const wildBeastmasterLabel = "Wild Beastmaster — each other creature you control gets +X/+X"

// Wild Beastmaster — Creature — Human Shaman {2}{G}, 1/1:
//
//	"Whenever this creature attacks, each other creature you control
//	 gets +X/+X until end of turn, where X is this creature's power."
//
// X is the Beastmaster's power as the trigger resolves, or its
// last-known power if it has left (CR 608.2h); the creatures are read
// once, at resolution. A power of 0 or less gives nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "52058c39-4edf-40f5-af18-e6de93f302a3",
		Name:         "Wild Beastmaster",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks(wildBeastmasterLabel, wildBeastmasterPump),
		},
	})
}

func wildBeastmasterPump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	info, ok := ctx.SourcePermanent()
	if !ok || info.Power <= 0 {
		return nil
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != item.Controller || !c.IsCreature() || c.InstanceID == item.SourceCardID {
			continue
		}
		if err := (BoostUntilEOT{Target: c.InstanceID, Power: info.Power, Toughness: info.Power, Label: wildBeastmasterLabel}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
