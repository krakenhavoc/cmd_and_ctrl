package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Abstract Paintmage — Creature — Djinn Sorcerer {U}{U/R}{R}, 2/2
// (Secrets of Strixhaven, #4456):
//
//	"At the beginning of your first main phase, add {U}{R}. Spend this
//	 mana only to cast instant and sorcery spells."
//
// A plain triggered ability (not a mana ability, CR 605.1b), so it uses
// the stack. The mana carries the cast-purpose restriction and an
// any-of Instant/Sorcery type restriction (Ore-Rich Stalactite's tags).
// Your first main phase is the precombat main phase, which an extra
// phase added later does not repeat.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bc899d5c-0a91-4f19-a3e4-3078ae966110",
		Name:         "Abstract Paintmage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourPrecombatMain("Abstract Paintmage — add {U}{R} (spend only on instant and sorcery spells)",
				Do(AddMana{
					Produced:     "{U}{R}",
					Restrictions: []string{ManaRestrictCast, game.ManaRestrictAnyType("Instant", "Sorcery")},
				})),
		},
	})
}
