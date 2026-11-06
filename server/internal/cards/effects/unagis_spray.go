package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Unagi's Spray — Instant {U} (Avatar: The Last Airbender Eternal,
// #5075):
//
//	"Target creature gets -4/-0 until end of turn. If you control a
//	 Fish, Octopus, Otter, Seal, Serpent, or Whale, draw a card."
//
// The draw checks your creature types (read through the layered view,
// so a changeling counts) as the spell resolves, not as it is cast (the
// ruling). A target that has become illegal makes the spell do nothing
// at all, the draw included (the ruling, CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "74b8976a-15cb-4604-a65d-ca9735cd6000",
		Name:         "Unagi's Spray",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(HarmUntilEOT(0, -4, 0)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			if err := (BoostUntilEOT{Target: t.ID, Power: -4, Label: "Unagi's Spray — -4/-0"}).Apply(ctx); err != nil {
				return err
			}
			if !avB29ControlsASeaCreature(ctx.Game, item.Controller) {
				return nil
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}

// avB29ControlsASeaCreature is "you control a Fish, Octopus, Otter,
// Seal, Serpent, or Whale" — any permanent with one of those subtypes.
func avB29ControlsASeaCreature(g *game.Game, you uuid.UUID) bool {
	g.RecomputeLayersIfStaleLocked()
	for _, c := range g.Battlefield.Cards {
		if c.Controller != you {
			continue
		}
		for _, s := range []string{"Fish", "Octopus", "Otter", "Seal", "Serpent", "Whale"} {
			if c.HasSubtype(s) {
				return true
			}
		}
	}
	return false
}
