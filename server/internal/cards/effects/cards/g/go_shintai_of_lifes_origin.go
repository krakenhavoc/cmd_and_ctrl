package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Go-Shintai of Life's Origin — Legendary Enchantment Creature — Shrine
// {3}{G}, 3/4 (Kamigawa: Neon Dynasty, #5493):
//
//	"{W}{U}{B}{R}{G}, {T}: Return target enchantment card from your
//	 graveyard to the battlefield.
//	 Whenever Go-Shintai of Life's Origin or another nontoken Shrine you
//	 control enters, create a 1/1 colorless Shrine enchantment creature
//	 token."
//
// The activation is a tap cost on a creature, so summoning sickness applies
// (CR 302.6). The trigger fires for the Go-Shintai itself (token or not)
// and for any other nontoken Shrine entering under your control; the Shrine
// token it makes is a Shrine, but being a token it does not trigger it
// again. "Shrine" reads through the text lens (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6dbcd542-ddb0-434e-b8ee-3e31e1a8b051",
		Name:         "Go-Shintai of Life's Origin",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}{U}{B}{R}{G}, {T}: Return target enchantment card from your graveyard to the battlefield.",
			Cost:    effects.Plus(effects.ManaCost("{W}{U}{B}{R}{G}"), effects.TapCost()),
			Targets: effects.TargetCardInGraveyard("target enchantment card from your graveyard", effects.YouOwn(), effects.Enchantment()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				t, ok := effects.FirstLegalTarget(ctx)
				if !ok {
					return nil
				}
				return effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}.Apply(ctx)
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
				if !ok || !c.HasSubtypeWord(source.TextSubtype("Shrine")) {
					return false
				}
				return ev.CardID == source.InstanceID || !effects.IsToken(c)
			}, "Go-Shintai of Life's Origin — create a 1/1 colorless Shrine enchantment creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Shrine enchantment"), N: 1})),
		},
	})
}
