package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Congregation Gryff — Creature — Hippogriff Mount {1}{G}{W}:
//
//	"Flying, lifelink
//	 Whenever this creature attacks while saddled, it gets +X/+X until
//	 end of turn, where X is the number of Mounts you control.
//	 Saddle 3"
//
// X is counted when the trigger resolves, off the live battlefield
// (CR 608.2h), and the Gryff counts itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "376cbc85-2941-4d97-b031-bffff86550e7",
		Name:            "Congregation Gryff",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Activated:       []ActivatedAbility{Saddle(3)},
		Triggered: []game.TriggeredAbility{
			AttacksWhileSaddled("Congregation Gryff — +X/+X, where X is the number of Mounts you control", func(g *game.Game, item *game.StackItem) error {
				x := countControlled(g, item.Controller, func(c game.Card) bool { return c.HasSubtype("Mount") })
				return BoostUntilEOT{Target: item.SourceCardID, Power: x, Toughness: x, Label: "Congregation Gryff — +X/+X"}.Apply(NewContext(g, item))
			}),
		},
	})
}
