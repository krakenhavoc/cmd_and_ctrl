package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Visions of Dominance — Sorcery {2}{G} (Midnight Hunt Commander):
//
//	"Put a +1/+1 counter on target creature, then double the number of +1/+1
//	 counters on it.
//	 Flashback {8}{G}{G}. This spell costs {X} less to cast this way, where X
//	 is the greatest mana value of a commander you own on the battlefield or
//	 in the command zone."
//
// Invigorating Surge's body on any creature: two placements in printed
// order, so a bare creature ends on two. The flashback discount is the
// Visions cycle's (visionsFlashbackDiscount).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:          "14290772-ca17-4a55-86b2-7a638929b12d",
		Name:              "Visions of Dominance",
		Completeness:      CompletenessFull,
		CastableZones:     []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts:  []game.AlternativeCost{visionsFlashbackAlternative("{8}{G}{G}")},
		SelfCostModifiers: []game.CostModifier{visionsFlashbackDiscount()},
		Targets:           TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b34PutCounterThenDoubleCounters(ctx)
		},
	})
}
