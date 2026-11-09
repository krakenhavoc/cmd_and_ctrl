package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Corroding Dragonstorm — Enchantment {1}{B} (Tarkir: Dragonstorm, #4463):
//
//	"When this enchantment enters, each opponent loses 2 life and you gain
//	 2 life. Surveil 2. (Look at the top two cards of your library, then
//	 put any number of them into your graveyard and the rest on top of your
//	 library in any order.)
//	 When a Dragon you control enters, return this enchantment to its owner's
//	 hand."
//
// Two abilities. The enters trigger is Ayara's drain followed by Surveil,
// the prompt queued last so the surveil is the final step. The second is
// Fleeting Effigy's bounce on a Dragon entering under your control (any
// Dragon, read through the layers as it enters): it returns the
// enchantment only while it is still the same permanent on the battlefield
// (CR 400.7), so a Dragon that enters in response to the first trigger
// brings it back after the drain.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4c55ca3b-1254-4d44-96ac-1e1cadaf49f5",
		Name:         "Corroding Dragonstorm",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Corroding Dragonstorm — each opponent loses 2 life and you gain 2 life; surveil 2",
				Do(EachOpponentLosesLife{N: 2}, GainLife{Amount: 2}, Surveil{N: 2})),
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.HasSubtype(source.TextSubtype("Dragon"))
			}, "Corroding Dragonstorm — return this enchantment to its owner's hand",
				Do(Instruction(func(ctx *Context) error {
					if !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
						return nil
					}
					return BounceToHand{Target: ctx.Item.SourceCardID}.Apply(ctx)
				}))),
		},
	})
}
