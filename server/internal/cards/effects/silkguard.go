package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Silkguard — Instant {X}{G}:
//
//	"Put a +1/+1 counter on each of up to X target creatures you control.
//	 Auras, Equipment, and modified creatures you control gain hexproof
//	 until end of turn. (Equipment, Auras you control, and counters are
//	 modifications.)"
//
// The targets are "up to X" (TargetSpec.CountFromX + UpToX), chosen as the
// spell is cast, and every still-legal one gets its counter first (CR
// 608.2b). Then the hexproof grant: the set is read once as the spell
// resolves (CR 611.2c), AFTER the counters, so a creature that Silkguard
// itself just put a counter on is modified and gets it. A creature is
// modified (CR 700.9) when it has a counter, is equipped, or is enchanted
// by an Aura its controller controls (b07IsModified); Auras and Equipment
// you control are covered whatever they are attached to.
//
// No simplification.
func init() {
	spec := TargetCreature("up to X target creatures you control", YouControl())
	spec.CountFromX = true
	spec.UpToX = true
	Register(Spec{
		OracleID:     "6dbeeb49-55c6-4ceb-94dd-3e4110dff8c0",
		Name:         "Silkguard",
		Completeness: CompletenessFull,
		Targets:      spec,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := putPlusOneCounterOnEachLegalTarget(ctx.Game, item); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{
				Match:    And(YouControl(), silkguardProtected),
				Keywords: []string{"hexproof"},
				Label:    "Silkguard — hexproof",
			}.Apply(ctx)
		},
	})
}

// silkguardProtected is "Auras, Equipment, and modified creatures": an
// Aura or an Equipment, or a creature with a modification.
func silkguardProtected(g *game.Game, _ uuid.UUID, c game.Card) bool {
	return c.IsAura() || c.HasSubtype("Equipment") || (c.IsCreature() && b07IsModified(g, c))
}
