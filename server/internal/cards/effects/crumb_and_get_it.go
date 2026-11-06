package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crumb and Get It — Instant {W} (Bloomburrow, #4467):
//
//	"Gift a Food (You may promise an opponent a gift as you cast this
//	 spell. If you do, they create a Food token before its other
//	 effects. It's an artifact with "{2}, {T}, Sacrifice this token: You
//	 gain 3 life.")
//	 Target creature you control gets +2/+2 until end of turn. If the
//	 gift was promised, that creature also gains indestructible until
//	 end of turn."
//
// The gift keyword (gift.go, CR 702.174): the chosen opponent's Food is
// made before the pump. The pump and the grant are two until-end-of-turn
// records on the one target; a target that has left counters the spell
// (CR 608.2b), and the opponent gets no Food then either (CR 702.174j).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "93596258-bb44-4aa3-833d-9a0ab16b5ec1",
		Name:         "Crumb and Get It",
		Completeness: CompletenessFull,
		Gift:         GiftAFood(),
		Targets:      TargetCreature("target creature you control", YouControl()),
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Crumb and Get It — +2/+2"}).Apply(ctx); err != nil {
					return err
				}
				if !ctx.GiftPromised() {
					return nil
				}
				return GrantKeywordUntilEOT{
					Target:   t.ID,
					Keywords: []string{"indestructible"},
					Label:    "Crumb and Get It — indestructible",
				}.Apply(ctx)
			}
			return nil
		},
	})
}
