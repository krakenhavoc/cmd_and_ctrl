package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Felonious Rage — Instant {R} (Murders at Karlov Manor, #4469):
//
//	"Target creature you control gets +2/+0 and gains haste until end of
//	 turn. When that creature dies this turn, create a 2/2 white and blue
//	 Detective creature token."
//
// A Good Day to Die's shape with haste added: the pump and the haste
// until end of turn, then a delayed triggered ability the spell creates
// as it resolves (CR 603.7a), bound to that one object, so a creature
// that blinks and comes back is a new object whose death makes nothing
// (CR 400.7). The Detective is the spell's controller's, who controls
// the delayed trigger (CR 603.7d). A target gone by resolution does
// nothing at all (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e1905321-180b-41d2-b7ef-0974ab90f188",
		Name:         "Felonious Rage",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 0, "haste")),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || !onBattlefield(ctx.Game, t.ID) {
				return nil
			}
			if err := (BoostUntilEOT{Target: t.ID, Power: 2, Label: "Felonious Rage — +2/+0"}).Apply(ctx); err != nil {
				return err
			}
			if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"haste"}, Label: "Felonious Rage — haste"}).Apply(ctx); err != nil {
				return err
			}
			return whenThatCreatureDiesThisTurn(ctx, t.ID, "Felonious Rage — create a 2/2 Detective", feloniousRageDetectiveBody)
		},
	})
}

// feloniousRageDetectiveBody is the delayed trigger's body: a 2/2 white
// and blue Detective for its controller. An on-disk key.
var feloniousRageDetectiveBody = game.SimpleDelayedBody("felonious-rage/detective", func(g *game.Game, item *game.StackItem) error {
	return CreateToken{Template: TokenCard("2/2 white and blue Detective"), N: 1}.Apply(NewContext(g, item))
})
