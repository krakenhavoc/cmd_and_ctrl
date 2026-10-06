package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hunter's Ambush — Instant {2}{G}:
//
//	"Prevent all combat damage that would be dealt by nongreen creatures
//	 this turn."
//
// #2026's negation over a colour, read as each creature would deal
// combat damage (CR 609.7b; the ruling: it applies "even if that
// creature was green or wasn't on the battlefield when Hunter's Ambush
// resolved"). A colourless creature is nongreen.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "1f085891-0d99-46a6-8f09-b84f44d701f8",
		Name:         "Hunter's Ambush",
		Completeness: CompletenessFull,
		OnResolve: sourceShieldSpell(combatShieldAgainstCreatures(game.DamageSourceFilter{
			Except: []game.PermanentQuery{QueryColors("G")},
		})),
	})
}
