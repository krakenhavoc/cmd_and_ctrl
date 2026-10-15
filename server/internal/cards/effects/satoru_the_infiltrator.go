package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Satoru, the Infiltrator — Legendary Creature — Human Ninja Rogue
// {U}{B}, 2/3 (EDHREC rank 3232):
//
//	"Menace
//	 Whenever Satoru and/or one or more other nontoken creatures you
//	 control enter, if none of them were cast or no mana was spent
//	 to cast them, draw a card."
//
// The cheat-into-play commander. Menace rides PrintedKeywords. "One or
// more" is one draw per event batch (OncePerBatch, CR 603.2c), so a mass
// reanimation draws once, as printed.
//
// The condition is a CR 603.4 intervening "if" over the WHOLE set (ADR
// 0109 §11 decision 4, #1552): no nontoken creature that entered under
// the controller's control in that batch was cast with mana spent on
// it. A reanimated, blinked or "put onto the battlefield" creature
// qualifies, and so does one cast without spending mana — a plotted
// creature, a cascade hit, a {0} creature. It is read when the ability
// triggers and again as it resolves (noneCastWithManaInBatch), so a
// cast-with-mana creature entering alongside one that was not stops the
// draw.
//
// A waived payment (strict mana off with too little in the pool, or Cast
// anyway) no longer leaves Satoru guessing: it declares ReadsManaSpent,
// so the caster is asked what they paid outside the pool, or whether they
// cast it without paying its mana cost, and the record it reads is exact
// (#3855, ADR 0068 amendment 2026-10-15).
func init() {
	Register(Spec{
		OracleID:        "7555c429-5f2d-4171-b6b0-8e3c8da7f314",
		Name:            "Satoru, the Infiltrator",
		ReadsManaSpent:  game.ReadsYourCasts,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventETB, satoruEntered, "Satoru, the Infiltrator — draw a card", satoruDraw)),
		},
	})
}

// satoruEntered is the trigger event and the check as it triggers: a
// nontoken creature entered under the controller's control, and none of
// the batch so far was cast with mana spent.
func satoruEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	if !ok || !c.IsCreature() || IsToken(c) {
		return false
	}
	return noneCastWithManaInBatch(g, ev.Batch, source.Controller)
}

// satoruDraw re-checks the condition over the whole batch as it
// resolves (CR 603.4), then draws.
var satoruDraw = Do(If{
	Cond: CondBoard{Name: "none of the creatures that entered was cast with mana", Test: func(ctx *Context) bool {
		item := ctx.Item
		return item.Trigger != nil && noneCastWithManaInBatch(ctx.Game, item.Trigger.Event.Batch, item.Controller)
	}},
	Then: []Applier{DrawCards{N: 1}},
})
