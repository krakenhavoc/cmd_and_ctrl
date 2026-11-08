package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vaultguard Trooper — Creature — Kavu Soldier {4}{R}, 5/5 (Edge of
// Eternities, #4461):
//
//	"At the beginning of your end step, if you control two or more tapped
//	 creatures, you may discard your hand. If you do, draw two cards."
//
// Dawnstrike Vanguard's end-step intervening "if" (CR 603.4): checked as
// the end step begins and again as the ability resolves, so untapping a
// creature in response makes it do nothing, as the ruling says. The
// Trooper counts toward the two when tapped. The "you may" is Narset,
// Jeskai Waymaster's keyed yes/no: a yes discards the whole hand and draws
// two. A hand with no cards in it may still be discarded, so a yes always
// counts as "if you do".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e15c78d4-6fb1-4bd3-8219-c66b83fcaba1",
		Name:         "Vaultguard Trooper",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "you control two or more tapped creatures",
				Test: func(ctx *Context) bool { return b24TappedCreaturesControlled(ctx.Game, ctx.Controller()) >= 2 },
			}, On(game.EventBeginEndStep, AllOf(ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b24TappedCreaturesControlled(g, source.Controller) >= 2
			}), "Vaultguard Trooper — you may discard your hand; if you do, draw two cards",
				func(g *game.Game, item *game.StackItem) error {
					return MayChoice{
						Question: "Vaultguard Trooper — discard your hand? If you do, draw two cards.",
						YesLabel: "Discard and draw two",
						NoLabel:  "Keep your hand",
						Then:     vaultguardTrooperDiscardHand,
					}.Apply(NewContext(g, item))
				})),
		},
	})
}

// vaultguardTrooperDiscardHand is the yes. An on-disk key: never renamed,
// never reused.
var vaultguardTrooperDiscardHand = MayYes("confirm/vaultguard-trooper-discard-hand", func(ctx *Context, r game.ConfirmAnswer) error {
	if _, err := discardWholeHand(ctx.Game, r.Chooser); err != nil {
		return err
	}
	return DrawCards{Player: r.Chooser, N: 2}.Apply(ctx)
})
