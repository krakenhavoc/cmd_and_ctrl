package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brought Back — Instant {W}{W} (Avatar: The Last Airbender Eternal, #5075):
//
//	"Choose up to two target permanent cards in your graveyard that were
//	 put there from the battlefield this turn. Return them to the
//	 battlefield tapped."
//
// The target clause is Samwise the Stouthearted's: a permanent card
// (artifact, creature, enchantment, land, planeswalker or battle) in your
// graveyard that this turn's record says went there from the battlefield,
// so a countered permanent spell never qualifies (the ruling). Both
// return as ONE simultaneous entry (ReturnFromGraveyardTogether, CR
// 603.6a) under their owner's control (you), tapped. A target that left
// the graveyard in response is dropped (CR 608.2b); an Aura chooses what
// it enchants as it enters and stays behind if it has nothing to enchant
// (CR 303.4f, the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f6ed46e2-08f9-4e89-aae3-e5dd7f30f007",
		Name:         "Brought Back",
		Completeness: CompletenessFull,
		Targets: TargetCardInGraveyard("up to two target permanent cards in your graveyard that were put there from the battlefield this turn",
			YouOwn(), Permanent(), putThereFromTheBattlefieldThisTurn).WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ReturnFromGraveyardTogether{Targets: legalTargetCardIDs(ctx), Tapped: true}.Apply(ctx)
		},
	})
}
