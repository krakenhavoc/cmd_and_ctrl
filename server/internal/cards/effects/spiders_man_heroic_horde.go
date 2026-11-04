package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spiders-Man, Heroic Horde — Legendary Creature — Spider Hero {1}{G}, 2/3
// (Marvel's Spider-Man, #4460):
//
//	"Web-slinging {4}{G}{G} (You may cast this spell for {4}{G}{G} if you
//	 also return a tapped creature you control to its owner's hand.)
//	 When Spiders-Man enters, if they were cast using web-slinging, you
//	 gain 3 life and create two 2/1 green Spider creature tokens with reach."
//
// Web-slinging is Spider-Sense's alternative cost. "If they were cast using
// web-slinging" is an intervening if (CR 603.4) over the permanent's cast
// provenance (CR 400.7d), checked as the trigger fires. It is not checked a
// second time on resolution: that the permanent was cast with web-slinging
// is a fact that cannot change, and last-known information (CR 608.2h)
// would keep it true even if Spiders-Man had left the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d21d7557-9702-46c1-8277-42a3442084c2",
		Name:         "Spiders-Man, Heroic Horde",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{{
			Key:          "web_slinging",
			Label:        "Web-slinging {4}{G}{G}",
			ManaCost:     "{4}{G}{G}",
			ReturnToHand: PermanentYouControl("a tapped creature you control", Creature(), b751Tapped()),
			PayLabel:     "a tapped creature you control",
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AllOf(Self, spidersManCastWithWebSlinging),
				"Spiders-Man, Heroic Horde — gain 3 life and create two 2/1 green Spider tokens with reach",
				Do(GainLife{Amount: 3}, CreateToken{Template: TokenCard("2/1 green Spider with reach"), N: 2})),
		},
	})
}

// spidersManCastWithWebSlinging is "if they were cast using web-slinging":
// the entering permanent's cast provenance names the web-slinging cost.
func spidersManCastWithWebSlinging(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return source != nil && source.Provenance.AltCost == "web_slinging"
}
