package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Darksteel Splicer — Creature — Phyrexian Artificer {6}{W}, 1/1 (March
// of the Machine Commander):
//
//	"Whenever this creature or another nontoken Phyrexian you control
//	 enters, create X 3/3 colorless Phyrexian Golem artifact creature
//	 tokens, where X is the number of opponents you have.
//	 Golems you control have indestructible."
//
// X is the number of opponents the controller has as the ability
// resolves. The trigger watches this creature or any other nontoken
// Phyrexian under the controller's control; "Phyrexian" is a word of the
// source's text. The static grants indestructible to every Golem the
// controller controls, the Splicer's own tokens included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "63a0f0a2-bafc-4cd0-a4ed-42ce58cd3e9b",
		Name:         "Darksteel Splicer",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "Golems you control have indestructible.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && target.IsCreature() && target.HasSubtype(source.TextSubtype("Golem"))
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Abilities = append(c.Abilities, "indestructible")
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				if !ok {
					return false
				}
				if ev.CardID == source.InstanceID {
					return true
				}
				return !c.IsToken() && c.HasSubtype(source.TextSubtype("Phyrexian"))
			}, "Darksteel Splicer — create X 3/3 colorless Phyrexian Golem artifact creature tokens, X = number of opponents",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return CreateToken{Template: TokenCard("3/3 colorless Phyrexian Golem artifact"), N: len(ctx.Opponents())}.Apply(ctx)
				}),
		},
	})
}
