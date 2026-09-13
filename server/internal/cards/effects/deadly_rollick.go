package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deadly Rollick — Instant {3}{B}:
//
//	"If you control a commander, you may cast this spell without
//	paying its mana cost.
//	 Exile target creature."
//
// One of the five "free spells" from Commander Legends. The exile
// half is ordinary removal; the free-cast half is what makes the card
// a staple.
//
// SANDBOX SIMPLIFICATION — the alternative cost is NOT implemented.
// Deadly Rollick costs {3}{B} here, always. CR 118.9 alternative
// costs need a cast-time branch (an extra cost mode on the cast
// action, a commander-on-battlefield gate, and a client affordance to
// choose between them) that the engine has no seam for; nothing in
// AdditionalCost or ManaAbilityCost expresses "instead of".
//
// The direction matters: paying full price is strictly WEAKER than
// printed. A player who wants the free cast can still take it in the
// sandbox by casting with the permissive cost gate, which is the same
// posture every other cost approximation in the catalog takes — the
// automation just never grants it silently.
//
// The same gap covers Fierce Guardianship, Deflecting Swat, Flawless
// Maneuver and Obscuring Haze; see
// docs/decklists/top-100-commander-staples.md.
func init() {
	Register(Spec{
		OracleID:     "0456ec64-2c81-4763-a352-8ff64a4c3d6b",
		Name:         "Deadly Rollick",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"You always pay {3}{B} — the free cast while you control a commander is not offered."},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return ExileTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
