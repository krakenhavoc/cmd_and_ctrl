package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Coral Colony — Creature — Wall {1}{U}, 1/4 (Dominaria United):
//
//	"Defender
//	 {1}{U}, {T}: Target player mills X cards, where X is the number of
//	 creatures you control with defender."
//
// X is counted when the ability resolves, through the layers, and includes
// the Colony itself while it is on the battlefield (it has defender).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5e186931-992b-491c-bffc-6877a613d5c3",
		Name:            "Coral Colony",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []ActivatedAbility{{
			Label:   "{1}{U}, {T}: Target player mills X cards, where X is the number of creatures you control with defender.",
			Cost:    Plus(ManaCost("{1}{U}"), TapCost()),
			Targets: TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				victim, ok := firstLegalPlayerTarget(ctx)
				if !ok {
					return nil
				}
				return MillCards{Player: victim, N: cmdrB17DefendersControlled(g, ctx.Controller())}.Apply(ctx)
			},
		}},
	})
}
