package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Enslave — Enchantment — Aura {4}{B}{B} (Planar Chaos, #5595):
//
//	"Enchant creature
//	 You control enchanted creature.
//	 At the beginning of your upkeep, enchanted creature deals 1 damage to its
//	 owner."
//
// Control Magic's layer-2 control change (ControlAttachedBySource) and an
// upkeep trigger of the Aura's controller. The damage's source is the
// enchanted creature as the trigger resolves, and it goes to that
// creature's OWNER, not its controller; with no creature enchanted any
// more, nothing is dealt.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "879ba7c4-4557-4166-adb0-b52454816e5a",
		Name:         "Enslave",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			ControlAttachedBySource(),
		},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Enslave — enchanted creature deals 1 damage to its owner", enslaveDamagesTheOwner),
		},
	})
}

// enslaveDamagesTheOwner is Enslave's upkeep trigger: the creature the
// Aura enchants as it resolves deals 1 damage to that creature's owner.
func enslaveDamagesTheOwner(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	host, ok := enchantedCreatureAsItResolves(ctx)
	if !ok {
		return nil
	}
	c, ok := g.LookupCardForEffect(host)
	if !ok {
		return nil
	}
	return DealDamage{Source: host, Target: c.Owner, Amount: 1}.Apply(ctx)
}
