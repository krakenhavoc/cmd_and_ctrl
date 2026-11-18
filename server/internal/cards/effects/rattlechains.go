package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "ee2ca311-7639-436d-8ecb-d2aad075f25f",
		Name:            "Rattlechains",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		CastTimings: []game.CastTimingRule{
			CastKindAsThoughFlash(game.PermissionFilter{CreatureType: "Spirit"},
				"You may cast Spirit spells as though they had flash."),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Rattlechains — target Spirit gains hexproof until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						id, ok := b16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return GrantKeywordUntilEOT{
							Target:   id,
							Keywords: []string{"hexproof"},
							Label:    "Rattlechains — hexproof until end of turn",
						}.Apply(ctx)
					}),
				TargetCreature("target Spirit", OfCreatureType("Spirit"))),
		},
	})
}
