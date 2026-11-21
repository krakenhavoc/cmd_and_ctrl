package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sea Hag // Aquatic Ingress — an Adventure creature (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	Sea Hag {4}{U}, Creature — Hag, 3/5
//	  "When this creature enters, creatures your opponents control get -4/-0
//	   until end of turn."
//	Aquatic Ingress {2}{U}, Instant — Adventure
//	  "Up to two target creatures each get +1/+0 and can't be blocked this
//	   turn."
//
// Bewildering Blizzard's one-sided mass -N/-0, taken from the creatures the
// opponents control as the trigger resolves (CR 611.2c). Aquatic Ingress
// applies +1/+0 and Minamo Sightbender's cannot-be-blocked restriction to each
// target still legal.
//
// No simplification.
func init() {
	const oracleID = "788a1d3f-8d73-4ecd-97ac-77906a873cfe"
	effects.Register(effects.Spec{
		OracleID:     oracleID,
		Name:         "Sea Hag",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sea Hag — creatures your opponents control get -4/-0 until end of turn",
				effects.Do(effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.OpponentControls()), Power: -4,
					Label: "Sea Hag — creatures your opponents control get -4/-0",
				})),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracleID + "#1",
		Name:         "Aquatic Ingress",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: 1, Label: "Aquatic Ingress — +1/+0"}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.RestrictUntilEOT{Target: t.ID, Restrictions: game.CantBeBlocked, Label: "Aquatic Ingress — can't be blocked"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
