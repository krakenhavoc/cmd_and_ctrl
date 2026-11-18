package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Moonveil Regent — Creature — Dragon {3}{R}, 4/4:
//
//	"Flying
//	 Whenever you cast a spell, you may discard your hand. If you do, draw
//	 a card for each of that spell's colors.
//	 When this creature dies, it deals X damage to any target, where X is
//	 the number of colors among permanents you control."
//
// The cast trigger counts the spell's colors when it is built (Ramos,
// Dragon Engine's shape), so a spell countered in response is read as it
// last was (CR 608.2h); the count rides the prompt as its word. A "yes"
// always discards the hand, even an empty one, and always draws (Vaultguard
// Trooper's reading of "if you do"). The dies trigger targets as it is put
// on the stack and counts the colors among the permanents its controller
// controls as it resolves, the Regent already gone, so it does not count
// itself (CR 603.10a for the leaves-the-battlefield look-back applies to
// the trigger, not to this count).
func init() {
	const castLabel = "Moonveil Regent — you may discard your hand; if you do, draw a card for each of that spell's colors"
	Register(Spec{
		OracleID:        "61c2a2cb-6424-4a20-81fb-f8e15a8f95ce",
		Name:            "Moonveil Regent",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventCast},
				Key:     castLabel,
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return source != nil && ev.Actor == source.Controller
				},
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, castLabel)
					item.Params.Amount = spellColorCount(g, ev.CardID)
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return MayChoice{
						Question: "Moonveil Regent — discard your hand? If you do, draw a card for each of that spell's colors (" + strconv.Itoa(item.Params.Amount) + ").",
						YesLabel: "Discard and draw",
						NoLabel:  "Keep your hand",
						Word:     strconv.Itoa(item.Params.Amount),
						Then:     moonveilRegentDiscardHand,
					}.Apply(NewContext(g, item))
				},
			},
			Targeting(
				WhenThisDies("Moonveil Regent — it deals X damage to any target, X the number of colors among permanents you control",
					sourceDealsComputedDamageToEachLegalTarget(func(ctx *Context) int {
						return len(colorsAmongPermanentsYouControl(ctx.Game, ctx.Controller()))
					})),
				TargetAny()),
		},
	})
}

// moonveilRegentDiscardHand is the yes: discard the hand, then draw one
// card for each of the spell's colors (the answer's word). An on-disk key.
var moonveilRegentDiscardHand = MayYes("confirm/moonveil-regent-discard-hand", func(ctx *Context, r game.ConfirmAnswer) error {
	if _, err := discardWholeHand(ctx.Game, r.Chooser); err != nil {
		return err
	}
	n, err := strconv.Atoi(r.Word)
	if err != nil || n <= 0 {
		return nil
	}
	return DrawCards{Player: r.Chooser, N: n}.Apply(ctx)
})
