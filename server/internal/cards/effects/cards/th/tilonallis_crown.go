package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tilonalli's Crown — Enchantment — Aura {1}{R} (Rivals of Ixalan, S82 OLD
// batch 217, #5531):
//
//	"Enchant creature
//	 When this Aura enters, it deals 1 damage to enchanted creature.
//	 Enchanted creature gets +3/+0 and has trample."
//
// The pump and the trample are the Aura's statics. The enters trigger deals 1
// damage from the Aura to the creature it is attached to as the trigger
// resolves (CR 608.2b: nothing if it is no longer attached to anything).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fbc2f469-09a5-4641-b990-b483d5d15c0d",
		Name:         "Tilonalli's Crown",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 0),
			effects.GrantToAttached("trample"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Tilonalli's Crown — deal 1 damage to enchanted creature", tilonallisCrownPing),
		},
	})
}

func tilonallisCrownPing(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	src, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || src.AttachedTo.Kind != game.TargetCard {
		return nil
	}
	return effects.DealDamage{Source: item.SourceCardID, Target: src.AttachedTo.ID, Amount: 1}.Apply(ctx)
}
