package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Curious Forager — Creature — Squirrel Druid {2}{G}, 3/2 (Bloomburrow,
// #4467):
//
//	"When this creature enters, you may forage. When you do, return
//	 target permanent card from your graveyard to your hand. (To forage,
//	 exile three cards from your graveyard or sacrifice a Food.)"
//
// MayForage (#5184), then "when you do" is a CR 603.12 reflexive
// trigger: the target is chosen only once the forage is paid, so a card
// the forage exiled is not a legal pick, and a Food card sacrificed to it
// can be (a Food token is gone).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "30af656a-99ab-4e0a-81fe-3cdb0ea6a0cc",
		Name:         "Curious Forager",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Curious Forager — you may forage; when you do, return a permanent card from your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					return MayForage(NewContext(g, item), "Curious Forager — forage to return a permanent card from your graveyard?", curiousForagerForaged)
				}),
		},
	})
}

// curiousForagerForaged is the paid half, keyed (#4264): the "when you
// do" reflexive trigger is queued against the asking item.
var curiousForagerForaged = PayUnlessBranches("pay-unless/curious-forager-return",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return WhenYouDo("Curious Forager — return target permanent card from your graveyard to your hand",
			curiousForagerReturnBody).Apply(ctx)
	}, nil)

// curiousForagerReturnBody is the reflexive trigger: the permanent card
// chosen as it went on the stack returns to its owner's hand, re-checked
// at resolution (CR 608.2b).
var curiousForagerReturnBody = game.ReflexiveBody("curious-forager/return-from-graveyard",
	simpleBody(b35ReturnChosenGraveyardCardToHand),
	constTargets(func() *game.TargetSpec {
		return TargetCardInGraveyard("target permanent card from your graveyard", Permanent(), YouOwn())
	}))
