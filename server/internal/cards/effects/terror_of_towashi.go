package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Terror of Towashi — Creature — Phyrexian Ogre {2}{B}{B}, 4/3 (March of
// the Machine):
//
//	"Deathtouch
//	 Whenever this creature attacks, you may pay {3}{B}. When you do,
//	 return target creature card from your graveyard to the battlefield.
//	 It's a Phyrexian in addition to its other types."
//
// The payment is made as the attack trigger resolves; paying it creates a
// reflexive trigger (CR 603.12) whose target is chosen as that trigger goes
// on the stack, so a card put into the graveyard by the payment's timing is
// a legal pick. The creature enters as a Phyrexian in addition to its other
// types (game.EntersAs, a layer-4 addition that lasts while it is that
// permanent, as Path of the Schemer does for Artifact).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c4b6dd8f-cdef-413a-a820-ec47e168b3bd",
		Name:            "Terror of Towashi",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks(terrorOfTowashiLabel, terrorOfTowashiOffer),
		},
	})
}

const terrorOfTowashiLabel = "Terror of Towashi — you may pay {3}{B}: return target creature card from your graveyard to the battlefield as a Phyrexian too"

func terrorOfTowashiOffer(g *game.Game, item *game.StackItem) error {
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{3}{B}",
		Question: "Terror of Towashi — pay {3}{B} to return a creature card from your graveyard to the battlefield?",
		Then:     terrorOfTowashiPaid,
	}.Apply(NewContext(g, item))
}

var terrorOfTowashiPaid = PayUnlessBranches("pay-unless/terror-of-towashi-return",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return ReflexiveTrigger{
			Label: "Terror of Towashi — return target creature card from your graveyard to the battlefield; it's a Phyrexian in addition to its other types",
			Body:  terrorOfTowashiReturnBody,
		}.Apply(ctx)
	}, nil)

var terrorOfTowashiReturnBody = game.ReflexiveBody("terror-of-towashi/return",
	simpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		ids := legalTargetCardIDs(ctx)
		if len(ids) == 0 {
			return nil
		}
		return ReturnFromGraveyardTogether{
			Targets: ids,
			EntersAs: &game.EntersAs{
				AddSubtypes: []string{"Phyrexian"},
				Label:       "Terror of Towashi — a Phyrexian in addition to its other types",
			},
		}.Apply(ctx)
	}),
	constTargets(func() *game.TargetSpec {
		return TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn())
	}))
