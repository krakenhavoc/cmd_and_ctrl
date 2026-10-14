package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Krav, the Unredeemed — Legendary Creature — Demon {4}{B}, 3/3:
//
//	"Partner with Regna, the Redeemer (When this creature enters,
//	 target player may put Regna into their hand from their library,
//	 then shuffle.)
//	 {B}, Sacrifice X creatures: Target player draws X cards and gains
//	 X life. Put X +1/+1 counters on Krav."
//
// Partner with is the shared CR 702.124j pair: the commander pairing is
// read off the oracle text by internal/deck, and the entry search is
// PartnerWith. The activation is Grim Hireling's variable sacrifice
// (SacrificeX): X is announced with the activation (CR 602.2b), exactly
// X creatures are named and leave together, and the effect reads X back
// with ctx.X(). Krav may be one of the creatures sacrificed; the target
// player still draws and gains, and the counters then have nowhere to
// go. A target player who has left the game (CR 608.2b) gets nothing,
// and Krav still gets his counters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "316d49a0-620b-4942-a5e8-04cff15e6c01",
		Name:         "Krav, the Unredeemed",
		Completeness: CompletenessFull,
		XMatters:     true,
		Triggered: []game.TriggeredAbility{
			PartnerWith("Krav, the Unredeemed", "Regna, the Redeemer"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{B}, Sacrifice X creatures: Target player draws X cards and gains X life. Put X +1/+1 counters on Krav.",
			Cost:    Plus(ManaCost("{B}"), SacrificeX("X creatures", Creature())),
			Targets: TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				x := ctx.X()
				if x <= 0 {
					return nil
				}
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetPlayer {
						continue
					}
					if err := (DrawCards{Player: t.ID, N: x}).Apply(ctx); err != nil {
						return err
					}
					if err := (GainLife{Player: t.ID, Amount: x}).Apply(ctx); err != nil {
						return err
					}
				}
				return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: x}.Apply(ctx)
			},
		}},
	})
}
