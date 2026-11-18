package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kyler, Sigardian Emissary — Legendary Creature — Human Cleric {3}{G}{W},
// 2/2 (Midnight Hunt Commander, #5497):
//
//	"Whenever another Human you control enters, put a +1/+1 counter on
//	 Kyler.
//	 Other Humans you control get +1/+1 for each counter on Kyler."
//
// The trigger is Avatar Enthusiasts' for Humans. The static is a layer
// 7c bonus per counter of any kind on Kyler (not only +1/+1), counted
// from Kyler as the layers are applied, so it follows the counters as
// they change; with none it gives nothing. It reaches other Humans only,
// through effective subtypes, so a changeling counts and Kyler does not
// pump itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "726cd041-5d0b-436c-bced-9335f56c0b0d",
		Name:         "Kyler, Sigardian Emissary",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype("Human"))
			}, "Kyler, Sigardian Emissary — put a +1/+1 counter on it", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		Static: []game.StaticAbility{{
			Label:     "Other Humans you control get +1/+1 for each counter on Kyler.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: b16OtherCreaturesYouControlOfSubtype("Human"),
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				n := 0
				for _, k := range source.Counters {
					if k > 0 {
						n += k
					}
				}
				c.Power += n
				c.Toughness += n
			},
		}},
	})
}
