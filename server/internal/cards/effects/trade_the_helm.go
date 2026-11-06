package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trade the Helm — Sorcery {4}{U} (Aetherdrift, #4464):
//
//	"Exchange control of target artifact or creature you control and
//	 target artifact or creature an opponent controls.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Switcheroo's exchange over two clauses (ADR 0065): the first must be
// yours and the second an opponent's, each an artifact or a creature, so
// a Vehicle that is not crewed is a legal pick on either side. The
// exchange is all or nothing (CR 701.12a): if either target is illegal at
// resolution, nothing changes hands. It has no duration. Cycling is the
// hand ability (CR 702.29a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c478fb5c-9fde-400d-a2cd-f8c4fa6b957a",
		Name:         "Trade the Helm",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetPermanent("target artifact or creature you control", Or(Artifact(), Creature()), YouControl()),
			TargetPermanent("target artifact or creature an opponent controls", Or(Artifact(), Creature()), OpponentControls()),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			mine, okMine := ctx.ClauseTarget(0)
			theirs, okTheirs := ctx.ClauseTarget(1)
			if !okMine || !okTheirs {
				return nil
			}
			return ExchangeControl{A: mine.ID, B: theirs.ID, Label: "Trade the Helm — exchange control"}.Apply(ctx)
		},
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
