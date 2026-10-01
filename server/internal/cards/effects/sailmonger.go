package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sailmonger — Creature — Human Monger {3}{U}, 3/3:
//
//	"{2}: Target creature gains flying until end of turn. Any player may
//	 activate this ability."
//
// An any-player row (CR 602.2, 602.1b): whoever activates it pays the
// {2} out of their own pool (CR 602.1a) and chooses the target (CR
// 602.2b, 601.2c), any creature on the battlefield. The flying is a
// layer-6 grant until end of turn (CR 613.1f, 514.2); a target that has
// become illegal by resolution gets nothing (CR 608.2b).
//
// No Purpose: whether flying helps depends on whose creature it is and
// on the board, so the bot does not reach across the table for it (ADR
// 0106 owner decision 2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b15af529-a61a-4228-bf2a-7304c9050720",
		Name:         "Sailmonger",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{2}: Target creature gains flying until end of turn. Any player may activate this ability.",
			Cost:      ManaCost("{2}"),
			Targets:   TargetCreature("target creature"),
			AnyPlayer: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"flying"},
						Label: "Sailmonger — flying until end of turn"}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
