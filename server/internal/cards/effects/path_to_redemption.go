package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Path to Redemption — Enchantment — Aura {1}{W} (Avatar: The Last
// Airbender Eternal, #5075):
//
//	"Enchant creature
//	 Enchanted creature can't attack or block.
//	 {5}, Sacrifice this Aura: Exile enchanted creature. Create a 1/1
//	 white Ally creature token. Activate only during your turn."
//
// Uneasy Alliance's shape with the set's Ally and an "only during your
// turn" window (any step of your turn, not sorcery speed). "Enchanted
// creature" is the creature the Aura was attached to as it was
// sacrificed (CR 608.2h), so the exile still finds it; it does not
// target. The activator makes the Ally even if the creature is gone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ac94860b-70b8-4624-9758-4bb024a56f60",
		Name:         "Path to Redemption",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			RestrictAttached(game.CantAttackOrBlock),
		},
		Activated: []ActivatedAbility{{
			Label:     "{5}, Sacrifice this Aura: Exile enchanted creature. Create a 1/1 white Ally creature token. Activate only during your turn.",
			Cost:      Plus(ManaCost("{5}"), SacrificeThis()),
			Condition: DuringYourTurn(),
			Purpose:   game.Purpose{Answers: game.AnswerRemove},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := exileEnchantedCreature(g, item); err != nil {
					return err
				}
				return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 white Ally"), N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
