package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sporogenic Infection — Enchantment — Aura {1}{B} (Duskmourn, #4466):
//
//	"Enchant creature
//	 When this Aura enters, target player sacrifices a creature of their
//	 choice other than enchanted creature.
//	 When enchanted creature is dealt damage, destroy it."
//
// The enters trigger targets a PLAYER (any player, the ruling), who chooses
// the creature, so hexproof on their creatures is irrelevant. The creature
// the Aura enchants is read as the trigger resolves, as it is or as it last
// was if the Aura has left (CR 608.2h), and is not among the choices; a
// player with no other creature sacrifices nothing (CR 701.21a), and a
// target that left leaves nothing to do (CR 608.2b).
//
// The damage trigger is Sleep Magic's condition (any damage dealt to the
// creature this Aura enchants, combat or not) and destroys that creature if
// it is still on the battlefield; a second damage event of the same moment
// finds it gone and does nothing more.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "659a259d-f75e-4af3-b568-c63225e647de",
		Name:         "Sporogenic Infection",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Sporogenic Infection — target player sacrifices a creature other than enchanted creature",
				sporogenicInfectionEdict),
				TargetPlayer("target player")),
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Amount > 0 && source.AttachedTo.Kind == game.TargetCard && source.AttachedTo.ID == ev.Target
			}, "Sporogenic Infection — destroy enchanted creature", sporogenicInfectionDestroy),
		},
	})
}

// sporogenicInfectionEdict is the enters trigger's body: the first still
// legal player target sacrifices a creature of their choice other than the
// creature the Aura enchants.
func sporogenicInfectionEdict(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	host := uuid.Nil
	if aura, ok := ctx.SourcePermanent(); ok && aura.AttachedTo.Kind == game.TargetCard {
		host = aura.AttachedTo.ID
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			g.PlayerSacrificesForEffect(item.SourceCardID, t.ID,
				sacrificeSpec("a creature other than enchanted creature", Creature(), OtherThan(host)),
				"Sporogenic Infection — sacrifice a creature other than enchanted creature")
			return nil
		}
	}
	return nil
}

// sporogenicInfectionDestroy destroys the creature that was dealt damage,
// named by the triggering event.
func sporogenicInfectionDestroy(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	return DestroyTarget{Target: item.Trigger.Event.Target}.Apply(NewContext(g, item))
}
