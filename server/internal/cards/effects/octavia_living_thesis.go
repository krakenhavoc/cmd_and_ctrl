package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Octavia, Living Thesis — Legendary Creature — Elemental Octopus {8}{U}{U},
// 8/8 (Outlaws of Thunder Junction Commander):
//
//	"This spell costs {8} less to cast if you have eight or more instant
//	 and/or sorcery cards in your graveyard.
//	 Ward {8}
//	 Magecraft — Whenever you cast or copy an instant or sorcery spell,
//	 target creature has base power and toughness 8/8 until end of turn."
//
// The discount is a SelfCostModifier of generic mana only, applied at CR
// 601.2f when the graveyard holds eight such cards. Magecraft is the shared
// cast-or-copy trigger with a creature target (any creature, not only your
// own); the effect sets base power and toughness to 8/8 until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "65681a2a-8395-4390-845f-9be974994dcb",
		Name:            "Octavia, Living Thesis",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"ward"},
		SelfCostModifiers: []game.CostModifier{
			CostsLess(8, "This spell costs {8} less to cast if you have eight or more instant and/or sorcery cards in your graveyard.",
				cmdr2S41EightOrMoreInstantsSorceriesInGraveyard),
		},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{8}"), "Octavia, Living Thesis — ward {8}"),
			Targeting(
				Magecraft("Octavia, Living Thesis — target creature has base power and toughness 8/8 until end of turn (magecraft)",
					cmdr2S41BaseEightEight),
				TargetCreature("target creature")),
		},
	})
}
