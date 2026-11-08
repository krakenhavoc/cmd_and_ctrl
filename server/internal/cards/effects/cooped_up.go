package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cooped Up — Enchantment — Aura {1}{W} (Wilds of Eldraine, #4471):
//
//	"Enchant creature
//	 Enchanted creature can't attack or block.
//	 {2}{W}: Exile enchanted creature."
//
// Pacifism, with Redemption Arc's activated exile at {2}{W}: the ability
// names the creature through the Aura's attachment as it resolves (CR
// 608.2h) and exiles it. It does not target, and the Aura is not
// sacrificed, so it then goes to the graveyard as a state-based action
// with nothing to enchant (CR 704.5m).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3e8bcb44-b875-4fe4-b96f-97a20cbfec93",
		Name:         "Cooped Up",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			RestrictAttached(game.CantAttackOrBlock),
		},
		Activated: []ActivatedAbility{{
			Label: "{2}{W}: Exile enchanted creature.",
			Cost:  ManaCost("{2}{W}"),
			// Removes a permanent with no target (ADR 0142's AnswerRemove).
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Effect:  exileEnchantedCreature,
		}},
	})
}
