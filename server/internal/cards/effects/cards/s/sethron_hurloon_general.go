package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sethron, Hurloon General — Legendary Creature — Minotaur Warrior
// {3}{R}{R}, 4/4 (Jumpstart, #5510):
//
//	"Whenever Sethron or another nontoken Minotaur you control enters, create
//	 a 2/3 red Minotaur creature token.
//	 {2}{B/R}: Minotaurs you control get +1/+0 and gain menace and haste
//	 until end of turn."
//
// Arahbo's "or another nontoken" entry condition over Minotaurs (the
// effective subtypes, so a changeling counts; the tokens it makes never
// trigger it). The activation locks the set of Minotaurs as it resolves
// (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "561c1ce9-41e7-474a-9476-bb9a7d24f0d5",
		Name:         "Sethron, Hurloon General",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind == game.EventETB && ev.CardID == source.InstanceID {
					return true
				}
				c, ok := effects.EnteredUnderYourControl(ev, source, g, true)
				return ok && c.IsCreature() && !effects.IsToken(c) && c.HasSubtype(source.TextSubtype("Minotaur"))
			}, "Sethron, Hurloon General — create a 2/3 red Minotaur creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/3 red Minotaur"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B/R}: Minotaurs you control get +1/+0 and gain menace and haste until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{2}{B/R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				minotaurs := effects.And(effects.Creature(), effects.YouControl(), effects.Subtype("Minotaur"))
				if err := (effects.BoostUntilEOT{Match: minotaurs, Power: 1, Label: "Sethron — Minotaurs get +1/+0"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Match: minotaurs, Keywords: []string{"menace", "haste"}, Label: "Sethron — Minotaurs gain menace and haste"}.Apply(ctx)
			},
		}},
	})
}
