package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dizzying Gaze — Enchantment — Aura {R} (Exodus, S82 OLD b739, #5626):
//
//	"Enchant creature you control
//	 {R}: Enchanted creature deals 1 damage to target creature with flying."
//
// The enchant clause is "creature you control" (Mortal Obstinacy's), which the
// Aura is held to as a state-based action (CR 704.5m): it falls off if its
// creature changes hands. The ability is the Aura's, so its controller
// activates it, but the damage source is the creature it is attached to as the
// ability resolves, named by object (CR 400.7); with the creature gone the
// ability does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e04f46d9-43e7-45ef-b59c-1810f30bec72",
		Name:         "Dizzying Gaze",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.YouControl()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: Enchanted creature deals 1 damage to target creature with flying.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Cost:    effects.ManaCost("{R}"),
			Targets: effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				hostID, ok := effects.EnchantedCreatureAsItResolves(ctx)
				if !ok {
					return nil
				}
				host, ok := g.LookupCardForEffect(hostID)
				if !ok || !effects.OnBattlefield(g, hostID) {
					return nil
				}
				ref := game.ObjectRef{ID: host.InstanceID, Epoch: host.ObjectEpoch}
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return effects.DealDamage{SourceObject: &ref, Target: t.ID, Amount: 1}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
