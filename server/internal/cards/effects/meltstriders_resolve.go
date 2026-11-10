package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Meltstrider's Resolve — Enchantment — Aura {G} (Edge of Eternities,
// #4461):
//
//	"Enchant creature you control
//	 When this Aura enters, enchanted creature fights up to one target
//	 creature an opponent controls.
//	 Enchanted creature gets +0/+2 and can't be blocked by more than one
//	 creature."
//
// Fight the _____ Fight's Aura (EnchantCreature(YouControl())) with its
// enters fight: the enters trigger targets up to one creature an opponent
// controls and the fight is between the creature the Aura is attached to as
// it resolves and that target (b10Fight, each dealing damage equal to its
// power to the other). With no target, or the target gone, nothing fights.
// The static half is Aquitect's Defenses' PumpAttached (0, 2) and Vorrac
// Battlehorns' MaxBlockers on the attached creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6f9bc0c6-1792-4771-a839-95a22d9c699a",
		Name:         "Meltstrider's Resolve",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Targets:      EnchantCreature(YouControl()),
		Static: []game.StaticAbility{
			PumpAttached(0, 2),
		},
		BlockRules: []game.BlockRule{MaxBlockers(OnAttached(), 1)},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Meltstrider's Resolve — enchanted creature fights up to one target creature an opponent controls",
					meltstridersResolveFight),
				TargetCreature("up to one target creature an opponent controls", OpponentControls()).WithCount(0, 1)),
		},
	})
}

// meltstridersResolveFight is the enters trigger: the enchanted creature
// fights the chosen creature, if both are still there.
func meltstridersResolveFight(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	host, ok := enchantedCreatureAsItResolves(ctx)
	if !ok {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return b10Fight(ctx, host, t.ID)
		}
	}
	return nil
}
