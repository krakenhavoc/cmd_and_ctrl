package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chill of the Grave — Instant {2}{U} (Innistrad: Crimson Vow, #5495):
//
//	"This spell costs {1} less to cast if you control a Zombie.
//	 Tap target creature. It doesn't untap during its controller's next
//	 untap step.
//	 Draw a card."
//
// The discount is a self cost modifier read as the spell is cast, through
// the layers (a changeling is a Zombie). The tap and the freeze are
// Celestial Regulator's TapAndFreeze (CR 611.2c, the ADR 0058 one-shot
// mark on the creature); then the controller draws. A target that has
// become illegal counters the whole spell, draw included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2ee4f7ba-7238-44b5-8315-9f50502ed796",
		Name:         "Chill of the Grave",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(1, "This spell costs {1} less to cast if you control a Zombie.",
				effects.YouControlA(effects.Subtype("Zombie"))),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.TapAndFreeze{Targets: []uuid.UUID{id}, Label: "Chill of the Grave — it doesn't untap during its controller's next untap step"}).Apply(ctx); err != nil {
				return err
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
