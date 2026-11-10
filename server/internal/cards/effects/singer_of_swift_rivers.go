package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Singer of Swift Rivers — Creature — Merfolk Shaman {1}{G}{U}, 3/2:
//
//	"Flash
//	 When this creature enters, put a shield counter on another target
//	 creature you control.
//	 You may cast Merfolk spells as though they had flash."
//
// The shield counter is CR 122.1c's engine-owned replacement and
// prevention (game/shield_counters.go). "Another" keeps the Singer out of
// its own target list. The flash grant is a cast-timing rule narrowed to
// Merfolk spells (Yeva's shape, with a creature-type filter).
func init() {
	Register(Spec{
		OracleID:        "7df43670-c948-464d-9bb9-db672e755f0c",
		Name:            "Singer of Swift Rivers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		CastTimings: []game.CastTimingRule{
			CastKindAsThoughFlash(game.PermissionFilter{CreatureType: "Merfolk"},
				"You may cast Merfolk spells as though they had flash."),
		},
		Triggered: []game.TriggeredAbility{
			withTargetsFrom(
				WhenThisEnters("Singer of Swift Rivers — put a shield counter on another target creature you control",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						id, ok := b16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return AddCounter{Target: id, Kind: game.CounterShield, N: 1}.Apply(ctx)
					}),
				AnotherTarget(func(other CardPredicate) *game.TargetSpec {
					return TargetCreature("another target creature you control", YouControl(), other)
				})),
		},
	})
}
