package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ambrosia Whiteheart — Legendary Creature — Bird {1}{W}, 2/2:
//
//	"Flash
//	 When Ambrosia Whiteheart enters, you may return another permanent
//	 you control to its owner's hand.
//	 Landfall — Whenever a land you control enters, Ambrosia Whiteheart
//	 gets +1/+0 until end of turn."
//
// The ETB does NOT target (no "target"), so it is a resolution-time
// choice rather than a pick_target prompt: ChoosePermanents over the
// controller's own permanents, asked as the trigger resolves (CR
// 608.2). Nothing is chosen while the trigger waits on the stack, a
// permanent with hexproof or shroud is as returnable as any other, and
// the trigger resolves whether or not there is anything to return.
// "You may" is the floor of zero: choosing nothing is an answer.
// "Another" excludes Ambrosia by instance (the item's SourceCardID), so
// a land, a token or a second same-named permanent is on offer and
// Ambrosia is not.
//
// Landfall is BoostUntilEOT on Ambrosia herself, layer 7c, swept at
// cleanup.
func init() {
	Register(Spec{
		OracleID:        "2bcc9f11-5b12-433e-9680-0f4b18aa521c",
		Name:            "Ambrosia Whiteheart",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ambrosia Whiteheart — you may return another permanent you control", ambrosiaReturnAnother),
			Landfall("Ambrosia Whiteheart — +1/+0 until end of turn (landfall)", func(g *game.Game, item *game.StackItem) error {
				return BoostUntilEOT{
					Target: item.SourceCardID,
					Power:  1,
					Label:  "Ambrosia Whiteheart — +1/+0 until end of turn",
				}.Apply(NewContext(g, item))
			}),
		},
	})
}

// ambrosiaReturnAnother asks the controller which, if any, of their
// OTHER permanents goes back to its owner's hand, and bounces it.
func ambrosiaReturnAnother(g *game.Game, item *game.StackItem) error {
	self := item.SourceCardID
	return ChoosePermanents{
		Question: "Ambrosia Whiteheart — you may return another permanent you control to its owner's hand",
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var out []uuid.UUID
			for _, id := range PermanentsControlledBy(g, of) {
				if id != self {
					out = append(out, id)
				}
			}
			return out, 0, 1
		},
		Then: func(ctx *Context, picked game.PromptedPicks) error {
			for _, id := range picked.Cards() {
				if !onBattlefield(ctx.Game, id) {
					continue
				}
				if err := (BounceToHand{Target: id}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	}.Apply(NewContext(g, item))
}
