package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Smellerbee, Rebel Fighter — Legendary Creature — Human Rebel Ally
// {3}{R}, 3/3 (Avatar: The Last Airbender Eternal, #5075):
//
//	"First strike
//	 Other creatures you control have haste.
//	 Whenever Smellerbee attacks, you may discard your hand. If you do,
//	 draw cards equal to the number of attacking creatures."
//
// The haste is a layer 6 grant over the other creatures its controller
// controls. The attack trigger is Astonishing Spider-Man's keyed yes/no:
// on a yes the whole hand is discarded, and only a hand of one or more
// cards counts as "if you do"; the draw counts the creatures attacking as
// it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "127bc05b-2945-434c-bc6a-283a267cc72a",
		Name:            "Smellerbee, Rebel Fighter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Static:          []game.StaticAbility{b16GrantKeywords(OtherCreaturesYouControl, "haste")},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Smellerbee, Rebel Fighter — you may discard your hand; if you do, draw cards equal to the number of attacking creatures",
				func(g *game.Game, item *game.StackItem) error {
					return MayChoice{
						Question: "Smellerbee, Rebel Fighter — discard your hand? If you do, draw cards equal to the number of attacking creatures.",
						Then:     smellerbeeDiscardHand,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// smellerbeeDiscardHand is the yes. An on-disk key: never renamed, never
// reused.
var smellerbeeDiscardHand = MayYes("confirm/smellerbee-rebel-fighter-discard-hand", func(ctx *Context, r game.ConfirmAnswer) error {
	n, err := discardWholeHand(ctx.Game, r.Chooser)
	if err != nil || n == 0 {
		return err
	}
	return DrawCards{Player: r.Chooser, N: len(AttackingCreatures(ctx.Game))}.Apply(ctx)
})
