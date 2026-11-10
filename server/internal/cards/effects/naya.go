package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Naya — Plane — Alara (Planechase Anthology, #4870):
//
//	"You may play any number of lands on each of your turns.
//	 Whenever chaos ensues, target red, green, or white creature you
//	 control gets +1/+1 until end of turn for each land you control."
//
// A plane (ADR 0216 §6). The static is an unbounded extra land allowance
// for the planar controller (game.UnlimitedLandPlays, read for the face-up
// planes by EffectiveLandDropsLocked); the lands played this way are
// additional land plays that outlive the plane (the ruling), because they
// are counted per turn, not per plane. The chaos ability targets a red,
// green or white creature the planar controller controls, and the bonus is
// the number of lands they control as it resolves, fixed then (the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:            "5ed6be01-af06-4f3d-8b43-ef8993c52c97",
		Name:                "Naya",
		Completeness:        CompletenessFull,
		AdditionalLandPlays: game.UnlimitedLandPlays,
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverChaosEnsues("Naya — target red, green, or white creature you control gets +1/+1 until end of turn for each land you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					n := countControlled(g, ctx.Controller(), func(c game.Card) bool { return c.IsLand() })
					if n <= 0 {
						return nil
					}
					return ScopedEffectFor{
						Target:   id,
						Mods:     []game.Mod{game.ModifyPTMod(n, n)},
						Duration: DurationUntilEndOfTurn(ctx),
						Label:    "Naya — +1/+1 for each land you control until end of turn",
					}.Apply(ctx)
				}),
				TargetCreature("target red, green, or white creature you control",
					YouControl(), Or(OfColor("R"), OfColor("G"), OfColor("W")))),
		},
	})
}
