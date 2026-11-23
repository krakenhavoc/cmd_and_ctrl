package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:        "2bcc9f11-5b12-433e-9680-0f4b18aa521c",
		Name:            "Ambrosia Whiteheart",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Ambrosia Whiteheart — you may return another permanent you control", ambrosiaReturnAnother),
			effects.Landfall("Ambrosia Whiteheart — +1/+0 until end of turn (landfall)", func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{
					Target: item.SourceCardID,
					Power:  1,
					Label:  "Ambrosia Whiteheart — +1/+0 until end of turn",
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}

// ambrosiaReturnAnother asks the controller which, if any, of their
// OTHER permanents goes back to its owner's hand, and bounces it.
func ambrosiaReturnAnother(g *game.Game, item *game.StackItem) error {
	return effects.S82OneB04ReturnAnotherPermanent(g, item,
		"Ambrosia Whiteheart — you may return another permanent you control to its owner's hand",
		ambrosiaReturned)
}

// ambrosiaReturned is the chosen permanent returning to its owner's hand,
// keyed (#4290).
var ambrosiaReturned = effects.RegisterPermanentsPick("prompt-run/ambrosia-whiteheart-return", effects.PermanentsDef{
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			if !effects.OnBattlefield(ctx.Game, id) {
				continue
			}
			if err := (effects.BounceToHand{Target: id}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	},
})
