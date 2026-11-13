package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Disarm — Instant {U} (Mirrodin, #5606):
//
//	"Unattach all Equipment from target creature."
//
// Every Equipment attached to the creature as the spell resolves comes
// off and stays on the battlefield, under whoever controls it (CR
// 701.3d). Auras and other attachments that are not Equipment stay. The
// Equipment are collected before the first one comes off, so nothing is
// skipped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "44d54814-2eeb-4df8-b8d3-789f0bcb307a",
		Name:         "Disarm",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				var equipment []uuid.UUID
				for _, c := range ctx.Game.BattlefieldCardsForEffect() {
					if c.HasSubtype("Equipment") && c.IsAttachedTo(t.ID) {
						equipment = append(equipment, c.InstanceID)
					}
				}
				for _, id := range equipment {
					if err := ctx.Game.UnattachForEffect(id); err != nil {
						return err
					}
				}
			}
			return nil
		},
	})
}
