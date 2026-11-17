package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Orthion, Hero of Lavabrink — Legendary Creature — Human Soldier {3}{R},
// 3/3 (March of the Machine Commander):
//
//	"{1}{R}, {T}: Create a token that's a copy of another target creature
//	 you control. It gains haste. Sacrifice it at the beginning of the next
//	 end step. Activate only as a sorcery.
//	 {6}{R}{R}{R}, {T}: Create five tokens that are copies of another target
//	 creature you control. They gain haste. Sacrifice them at the beginning
//	 of the next end step. Activate only as a sorcery."
//
// Jaxis's copy with the sacrifice as Mishra's delayed trigger
// (s82CopiesHasteSacrifice): the target is read as the ability resolves and
// a creature that left in response makes nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "31895957-5fa8-46dc-a76e-37bbdc6072ad",
		Name:         "Orthion, Hero of Lavabrink",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:        "{1}{R}, {T}: Create a token that's a copy of another target creature you control. It gains haste. Sacrifice it at the beginning of the next end step. Activate only as a sorcery.",
				Cost:         Plus(ManaCost("{1}{R}"), TapCost()),
				Targets:      Another(TargetCreature("another target creature you control", YouControl())),
				SorcerySpeed: true,
				Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return s82CopiesHasteSacrifice(NewContext(g, item), 1, "Orthion")
				},
			},
			{
				Label:        "{6}{R}{R}{R}, {T}: Create five tokens that are copies of another target creature you control. They gain haste. Sacrifice them at the beginning of the next end step. Activate only as a sorcery.",
				Cost:         Plus(ManaCost("{6}{R}{R}{R}"), TapCost()),
				Targets:      Another(TargetCreature("another target creature you control", YouControl())),
				SorcerySpeed: true,
				Purpose:      game.Purpose{Tokens: 5, CreatureTokens: 5},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return s82CopiesHasteSacrifice(NewContext(g, item), 5, "Orthion")
				},
			},
		},
	})
}
