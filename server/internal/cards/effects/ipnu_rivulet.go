package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ipnu Rivulet — Land — Desert:
//
//	"{T}: Add {C}.
//	 {T}, Pay 1 life: Add {U}.
//	 {1}{U}, {T}, Sacrifice a Desert: Target player mills four cards."
//
// The Desert shape (cmdr2B13DesertLand): the sacrificed Desert may be
// this land or another.
//
// No simplification.
func init() {
	Register(cmdr2B13DesertLand("c17d799f-adc9-4c41-87cf-b243b5ea3be1", "Ipnu Rivulet", "U", ActivatedAbility{
		Label:   "{1}{U}, {T}, Sacrifice a Desert: Target player mills four cards.",
		Cost:    Plus(ManaCost("{1}{U}"), TapCost(), cmdr2B13SacrificeADesert()),
		Targets: TargetPlayer("target player"),
		Effect: func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			victim, ok := firstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			return MillCards{Player: victim, N: 4}.Apply(ctx)
		},
	}))
}
