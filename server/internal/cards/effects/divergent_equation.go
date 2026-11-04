package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Divergent Equation — Instant {X}{X}{U} (Secrets of Strixhaven, #4456):
//
//	"Return up to X target instant and/or sorcery cards from your
//	 graveyard to your hand.
//	 Exile Divergent Equation."
//
// "Up to X" is the announced X as the ceiling (TargetSpec.CountFromX with
// UpToX, March of Swirling Mist's shape), chosen as the spell is cast and
// paid for with {X}{X}, so X targets cost 2X generic. Every pick still
// legal as the spell resolves goes to hand (CR 608.2b). The spell then
// exiles itself, so it is not put into the graveyard at the end of its
// resolution (CR 608.2n applies to a spell still on the stack): it does so
// at X = 0 too, so the card does not declare XMatters. A copy is no card
// and has nothing to exile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "84681f12-2978-4f62-ab89-b579033c5192",
		Name:         "Divergent Equation",
		Completeness: CompletenessFull,
		Targets:      divergentEquationTargets(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, id := range legalTargetCardIDs(ctx) {
				if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
					return err
				}
			}
			if item.IsCopy || sourceIsNewObject(ctx.Game, item) {
				return nil
			}
			return ctx.Game.ExileCardForEffect(item.SourceCardID)
		},
	})
}

// divergentEquationTargets is "up to X target instant and/or sorcery cards
// from your graveyard".
func divergentEquationTargets() *game.TargetSpec {
	spec := TargetCardInGraveyard("up to X target instant and/or sorcery cards from your graveyard",
		YouOwn(), Or(Instant(), Sorcery()))
	spec.CountFromX = true
	spec.UpToX = true
	return spec
}
