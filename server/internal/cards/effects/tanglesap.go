package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tanglesap — Instant {1}{G}:
//
//	"Prevent all combat damage that would be dealt this turn by
//	 creatures without trample."
//
// #2026's negation over a keyword, read as each creature would deal
// combat damage (CR 609.7b; the ruling: "Creatures are checked to see
// whether they have trample at the time they'd deal combat damage").
// Attackers and blockers alike.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "a533df83-782f-4f77-a0be-312ae56f6447",
		Name:         "Tanglesap",
		Completeness: CompletenessFull,
		OnResolve: sourceShieldSpell(combatShieldAgainstCreatures(game.DamageSourceFilter{
			Except: []game.PermanentQuery{{Keyword: "trample"}},
		})),
	})
}
