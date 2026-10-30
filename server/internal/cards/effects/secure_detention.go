package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Secure Detention — Enchantment — Aura {3}{W} (Marvel Super Heroes
// Commander, #4455):
//
//	"Enchant artifact or creature
//	 When this Aura enters, create a 1/1 white Soldier creature token.
//	 Enchanted permanent can't attack or block, and its activated
//	 abilities can't be activated."
//
// Arrest's restriction bundle (attack, block, and every activated
// ability including mana abilities) on an artifact-or-creature host,
// plus an enters trigger for the Soldier. The enchant clause is the
// spec the CR 704.5m check re-runs, so it names the permanent types.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ce95e51d-4ef3-44a8-ae47-cf1bf71ebd3f",
		Name:         "Secure Detention",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("enchant artifact or creature", Or(Artifact(), Creature())),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			RestrictAttached(
				game.CantAttackOrBlock | game.CantActivate | game.CantActivateMana,
			),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Secure Detention — create a 1/1 white Soldier creature token",
				Do(CreateToken{Template: TokenCard("1/1 white Soldier"), N: 1})),
		},
	})
}
