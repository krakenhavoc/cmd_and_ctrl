package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:        "5b8084b2-0c3c-4a54-86dc-41f4fe513e99",
		Name:            "Apocalypse Demon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
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
			effects.AtYourUpkeep("Apocalypse Demon — tap this unless you sacrifice another creature", ApocalypseDemonUpkeep),
		},
	})
}

func ApocalypseDemonUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	source := item.SourceCardID
	if len(effects.PermanentsControlledByMatching(g, item.Controller, effects.And(effects.Creature(), effects.NotSelf(source)))) == 0 {
		return effects.TapTarget{Target: source}.Apply(ctx)
	}
	return effects.PickOption{
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
var apocalypseDemonAnswer = effects.OptionPickThen("option-pick/apocalypse-demon-sacrifice-or-tap", func(ctx *effects.Context, r game.OptionPicked) error {
	source := ctx.Source()
	if r.Index == 1 {
		ctx.Game.PlayerSacrificesForEffect(source, ctx.Controller(),
			effects.SacrificeSpec("another creature", effects.Creature(), effects.NotSelf(source)),
			"Apocalypse Demon — sacrifice another creature")
		return nil
	}
	return effects.TapTarget{Target: source}.Apply(ctx)
})
