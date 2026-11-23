package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rattlechains — Creature — Spirit {1}{U}, 2/1:
//
//	"Flash
//	 Flying
//	 When this creature enters, target Spirit gains hexproof until end of
//	 turn.
//	 You may cast Spirit spells as though they had flash."
//
// The timing grant is Breath of the Sleepless'. The ETB targets any
// Spirit (Rattlechains itself included) and gives it hexproof until end
// of turn as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ee2ca311-7639-436d-8ecb-d2aad075f25f",
		Name:            "Rattlechains",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		CastTimings: []game.CastTimingRule{
			effects.CastKindAsThoughFlash(game.PermissionFilter{CreatureType: "Spirit"},
				"You may cast Spirit spells as though they had flash."),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Rattlechains — target Spirit gains hexproof until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return effects.GrantKeywordUntilEOT{
							Target:   id,
							Keywords: []string{"hexproof"},
							Label:    "Rattlechains — hexproof until end of turn",
						}.Apply(ctx)
					}),
				effects.TargetCreature("target Spirit", effects.OfCreatureType("Spirit"))),
		},
	})
}
