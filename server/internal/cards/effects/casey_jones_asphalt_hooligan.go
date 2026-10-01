package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Casey Jones, Asphalt Hooligan — Legendary Creature — Human Berserker
// {2}{R}, 2/2:
//
//	"Double strike (This creature deals both first-strike and regular
//	 combat damage.)
//	 {4}: Double Casey Jones's power until end of turn. Any player may
//	 activate this ability."
//
// Double strike is a printed keyword (CR 702.4). The activation is an
// any-player row (CR 602.2, 602.1b): whoever activates it pays the {4}
// out of their own pool (CR 602.1a).
//
// "Double its power" is CR 701.10b: Casey gets +X/+0 until end of turn,
// where X is its power as the ability resolves, counters and other
// effects included. It is a one-time reading, not a multiplier: a pump
// later in the turn is not doubled again. A negative power doubles
// downwards (CR 701.10c), so X is the unclamped power. A Casey that has
// left the battlefield, or left and came back as a new object (CR
// 400.7), is not doubled.
//
// No Purpose: the bot never pays to pump a creature it does not
// control (ADR 0106 owner decision 2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1eaa37a3-8769-4213-bcd5-39bcf4771273",
		Name:            "Casey Jones, Asphalt Hooligan",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Activated: []ActivatedAbility{{
			Label:     "{4}: Double Casey Jones's power until end of turn. Any player may activate this ability.",
			Cost:      ManaCost("{4}"),
			AnyPlayer: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				c, ok := g.LookupCardForEffect(item.SourceCardID)
				if !ok {
					return nil
				}
				return BoostUntilEOT{
					Target: item.SourceCardID,
					Power:  c.PowerForComparison(),
					Label:  "Casey Jones, Asphalt Hooligan — double power",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
