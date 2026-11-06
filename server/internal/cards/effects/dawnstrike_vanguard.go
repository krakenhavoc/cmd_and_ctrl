package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dawnstrike Vanguard — Creature — Human Knight {5}{W}, 4/5 (Edge of
// Eternities, #4461):
//
//	"Lifelink
//	 At the beginning of your end step, if you control two or more tapped
//	 creatures, put a +1/+1 counter on each creature you control other than
//	 this creature."
//
// Lifelink is a printed keyword. The "if" is an intervening if (CR 603.4):
// checked as the end step begins and again as the ability resolves, so
// untapping a creature in response leaves it doing nothing. The Vanguard
// itself counts toward the two tapped creatures. The set of creatures is
// read as it resolves; the Vanguard is never in it, even after it has left
// the battlefield, and each counter is its own placement
// (putPlusOneCountersOnEachYouControlThat).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "69892dc7-4087-447c-a41c-99cf3d5c0c12",
		Name:            "Dawnstrike Vanguard",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "you control two or more tapped creatures",
				Test: func(ctx *Context) bool { return b24TappedCreaturesControlled(ctx.Game, ctx.Controller()) >= 2 },
			}, On(game.EventBeginEndStep, AllOf(ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b24TappedCreaturesControlled(g, source.Controller) >= 2
			}), "Dawnstrike Vanguard — put a +1/+1 counter on each other creature you control",
				dawnstrikeCounters)),
		},
	})
}

// dawnstrikeCounters puts a +1/+1 counter on each creature the ability's
// controller controls other than its source.
func dawnstrikeCounters(g *game.Game, item *game.StackItem) error {
	source := item.SourceCardID
	return putPlusOneCountersOnEachYouControlThat(1,
		func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.InstanceID != source })(g, item)
}
