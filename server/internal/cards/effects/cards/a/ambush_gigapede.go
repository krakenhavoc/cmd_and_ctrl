package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ambush Gigapede — Creature — Insect {4}{B}{B}, 6/2 (Outlaws of Thunder
// Junction, #4468):
//
//	"Flash
//	 When this creature enters, target creature an opponent controls
//	 gets -2/-2 until end of turn."
//
// Poisoner's Apprentice's enters trigger without its condition: a
// layer 7c -2/-2 on the still-legal target (CR 608.2b), gone at cleanup.
// Flash is the printed keyword the cast-timing gate reads.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4faf62c3-3ba7-4540-be26-777305ef184e",
		Name:            "Ambush Gigapede",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Ambush Gigapede — target creature an opponent controls gets -2/-2 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.BoostUntilEOT{
						Target:    id,
						Power:     -2,
						Toughness: -2,
						Label:     "Ambush Gigapede — -2/-2 until end of turn",
					}.Apply(ctx)
				}),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
				effects.ForTargets(effects.HarmUntilEOT(0, -2, -2))),
		},
	})
}
