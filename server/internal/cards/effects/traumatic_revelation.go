package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Traumatic Revelation — Sorcery {1}{B} (Reality Fracture, tracker
// #2795):
//
//	"Target opponent reveals their hand. You may choose a creature or
//	 battle card from it. If you do, that player discards that card. If
//	 you don't, incubate 3."
//
// The optional revealed-hand pick (#2115, ADR 0116's 2026-10-05
// amendment) with a keyed continuation, Nightsnare's shape: the whole
// table sees the hand (CR 701.20a), the chooser may choose a card or
// nothing, a chosen card is discarded, and "if you don't" is the
// continuation run after the pick with nothing chosen, including when
// the hand held no creature or battle card to choose. The incubate
// (CR 701.53a, ADR 0185) is the chooser's, not the discarding player's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e47253d1-9b12-4989-80c0-604906c88b5e",
		Name:         "Traumatic Revelation",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ChooseFromRevealedHand{
				Player:   TargetedPlayer(ctx),
				Filter:   Or(Creature(), Battle()),
				Label:    "creature or battle card",
				Optional: true,
				Then:     traumaticRevelationIfYouDont,
			}.Apply(ctx)
		},
	})
}

var traumaticRevelationIfYouDont = RevealedPickThen("revealed-pick/traumatic-revelation-incubate",
	func(ctx *Context, pick game.RevealedPick) error {
		if len(pick.Chosen) > 0 {
			return nil
		}
		return Incubate{Controller: pick.Chooser, N: 3}.Apply(ctx)
	})
