package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Boon of Boseiju — Instant {1}{G} (Kamigawa: Neon Dynasty, #5493):
//
//	"Target creature gets +X/+X until end of turn, where X is the greatest
//	 mana value among permanents you control. Untap it."
//
// X is counted as the spell resolves (CR 608.2), over the permanents its
// controller controls then, on the battlefield, where {X} in a cost is 0
// (CR 202.3e). The pump is a layer 7c effect until end of turn; the untap
// is the second sentence and happens even when X is 0.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "44e1946d-01b5-45db-abdc-2f6d6480cf60",
		Name:         "Boon of Boseiju",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0)),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			id := item.Targets[0].ID
			x := cmdrB13GreatestManaValueYouControl(ctx.Game, ctx.Controller())
			if err := (BoostUntilEOT{Target: id, Power: x, Toughness: x,
				Label: "Boon of Boseiju — +X/+X"}).Apply(ctx); err != nil {
				return err
			}
			return UntapTarget{Target: id}.Apply(ctx)
		},
	})
}
