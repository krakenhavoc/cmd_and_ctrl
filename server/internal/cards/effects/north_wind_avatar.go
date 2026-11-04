package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// North Wind Avatar — {2}{U}{U}{R} Creature — Dragon Spirit Avatar 5/5
// (Teenage Mutant Ninja Turtles, #4457):
//
//	"Flying
//	 When this creature enters, if you cast it, you may put a card you own
//	 from outside the game into your hand."
//
// "If you cast it" is Zacama's b16EnteredFromStack, an intervening if
// (CR 603.4) checked as it enters and again as the trigger resolves: a
// reanimated or flickered Avatar does nothing, and neither does one a
// player cast but someone else controls (ADR 0104). The take is Death
// Wish's TakeFromOutside with no filter and no reveal (ADR 0191, ADR
// 0198): from the sideboard when the deck seated one, otherwise from the
// card database, and the player may take nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b408c345-b308-4920-8494-4f204d856aa3",
		Name:            "North Wind Avatar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return b06SelfETB(ev, source, lki, g) && b16EnteredFromStack(g, source.InstanceID)
			}, "North Wind Avatar — you may put a card you own from outside the game into your hand",
				func(g *game.Game, item *game.StackItem) error {
					if !b16EnteredFromStack(g, item.SourceCardID) {
						return nil
					}
					return TakeFromOutside{
						Pick:     northWindAvatarTake,
						Question: "North Wind Avatar — put a card you own from outside the game into your hand, or none",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// northWindAvatarTake is the keyed pick (#4256). No filter, no reveal.
var northWindAvatarTake = registerOutsideTake("choose-cards/north-wind-avatar-take", nil, false)
