package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Apocalypse Demon — Creature — Demon {4}{B}{B}, */*:
//
//	"Flying
//	 Apocalypse Demon's power and toughness are each equal to the
//	 number of cards in your graveyard.
//	 At the beginning of your upkeep, tap this creature unless you
//	 sacrifice another creature."
//
// A layer 7a CDA over the controller's graveyard, and The Gitrog
// Monster's "unless" shape with a creature to sacrifice and a tap in
// place of the sacrifice. The sacrifice is the player's choice: with
// no other creature the Demon simply taps.
func init() {
	Register(Spec{
		OracleID:        "5b8084b2-0c3c-4a54-86dc-41f4fe513e99",
		Name:            "Apocalypse Demon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				if p := g.PlayerByIDForEffect(source.Controller); p != nil && p.Graveyard != nil {
					n = p.Graveyard.Size()
				}
				c.Power = n
				c.Toughness = n
			},
		}},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Apocalypse Demon — tap this unless you sacrifice another creature", apocalypseDemonUpkeep),
		},
	})
}

func apocalypseDemonUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	source := item.SourceCardID
	if len(permanentsControlledByMatching(g, item.Controller, And(Creature(), NotSelf(source)))) == 0 {
		return TapTarget{Target: source}.Apply(ctx)
	}
	return PickOption{
		Question: "Apocalypse Demon — sacrifice another creature, or tap Apocalypse Demon",
		Options: []game.ChoiceOption{
			{Label: "Tap Apocalypse Demon"},
			{Label: "Sacrifice another creature"},
		},
		ThenKey: apocalypseDemonAnswer,
	}.Apply(ctx)
}

// apocalypseDemonAnswer is the upkeep's choice answered, keyed (#4253): it
// reads only the controller and the Demon, both on the prompt. The creature
// to sacrifice is asked as the engine's own sacrifice prompt, plain data.
var apocalypseDemonAnswer = OptionPickThen("option-pick/apocalypse-demon-sacrifice-or-tap", func(ctx *Context, r game.OptionPicked) error {
	source := ctx.Source()
	if r.Index == 1 {
		ctx.Game.PlayerSacrificesForEffect(source, ctx.Controller(),
			sacrificeSpec("another creature", Creature(), NotSelf(source)),
			"Apocalypse Demon — sacrifice another creature")
		return nil
	}
	return TapTarget{Target: source}.Apply(ctx)
})
