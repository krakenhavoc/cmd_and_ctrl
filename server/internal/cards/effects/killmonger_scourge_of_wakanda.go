package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Killmonger, Scourge of Wakanda — Legendary Creature — Human Mercenary
// Villain {2}{B}{G}, 3/3 (Marvel Super Heroes, #4455):
//
//	"When Killmonger enters, you may sacrifice another creature. When you
//	 do, destroy target nonland permanent an opponent controls.
//	 As long as there are two or more creature cards in your graveyard,
//	 Killmonger gets +2/+1."
//
// Dire Blunderbuss's shape: the "you may" is the parent's confirm, the
// sacrifice is a prompt over the controller's other creatures, and the
// "when you do" reflexive trigger (CR 603.12) goes on the stack only once
// a creature has really been sacrificed, choosing its target as it does.
// The +2/+1 is a layer 7c modification of Killmonger himself, read live on
// every layer pass, so it comes and goes as the graveyard fills and empties
// (Blackbloom Rogue's shape).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1c3ba54a-a8ef-4838-b3aa-cb6df827a619",
		Name:         "Killmonger, Scourge of Wakanda",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Killmonger — you may sacrifice another creature", killmongerOffer),
		},
		Static: []game.StaticAbility{{
			Label:    "As long as there are two or more creature cards in your graveyard, Killmonger gets +2/+1.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && GraveyardAtLeast(2, func(c game.Card) bool { return c.IsCreature() })(g, source.Controller, source.InstanceID)
			},
			Apply: pumpBy(2, 1),
		}},
	})
}

// killmongerOffer is the ETB: the sacrifice is offered only when there is
// another creature to give.
func killmongerOffer(g *game.Game, item *game.StackItem) error {
	if countControlled(g, item.Controller, func(c game.Card) bool { return c.IsCreature() && c.InstanceID != item.SourceCardID }) == 0 {
		return nil
	}
	return MayChoice{
		Question: "Killmonger — sacrifice another creature to destroy target nonland permanent an opponent controls?",
		YesLabel: "Sacrifice a creature",
		NoLabel:  "Don't",
		Then:     killmongerAnswer,
	}.Apply(NewContext(g, item))
}

// killmongerAnswer is the "yes": sacrifice another creature, keyed (#4291).
var killmongerAnswer = MayYes("confirm/killmonger-sacrifice", func(ctx *Context, _ game.ConfirmAnswer) error {
	return ctx.SacrificeThen(ctx.Item.Controller,
		sacrificeSpec("another creature", Creature(), NotSelf(ctx.Item.SourceCardID)),
		"Killmonger — sacrifice another creature",
		1, killmongerSacrificed)
})

// killmongerSacrificed raises the reflexive destroy once a creature has
// really gone, keyed (#4290).
var killmongerSacrificed = RunThen("prompt-run/killmonger-sacrificed", func(ctx *Context, r game.RunDone) error {
	if r.Sacrifices().Count() == 0 {
		return nil
	}
	return WhenYouDo("Killmonger — destroy target nonland permanent an opponent controls", killmongerDestroyBody).Apply(ctx)
})

// killmongerDestroyBody is the reflexive "destroy target nonland permanent
// an opponent controls".
var killmongerDestroyBody = game.ReflexiveBody("killmonger-scourge-of-wakanda/destroy", simpleBody(destroyChosenPermanent),
	constTargets(func() *game.TargetSpec {
		return TargetPermanent("target nonland permanent an opponent controls", Nonland(), OpponentControls())
	}))
