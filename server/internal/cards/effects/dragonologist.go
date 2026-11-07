package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dragonologist — Creature — Human Wizard {2}{U}, 1/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"When this creature enters, look at the top six cards of your
//	 library. You may reveal an instant, sorcery, or Dragon card from
//	 among them and put it into your hand. Put the rest on the bottom of
//	 your library in a random order.
//	 Untapped Dragons you control have hexproof."
//
// The entry trigger is Master Piandao's look-and-take with six cards and
// an instant, sorcery, or Dragon filter. The grant is a layer-6 static
// (CR 613.1f) on each untapped Dragon you control, the Dragonologist's
// own type word read through its text (CR 612.2); a Dragon that becomes
// tapped loses hexproof at once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f6791df0-a37b-44dc-8bf3-40ac2304d188",
		Name:         "Dragonologist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dragonologist — look at the top six cards; you may reveal an instant, sorcery, or Dragon card and put it into your hand",
				LookAtTopThenMayTakeToHand(6, Or(Instant(), Sorcery(), AnySubtype("Dragon")), 1,
					"Dragonologist — reveal an instant, sorcery, or Dragon card and put it into your hand")),
		},
		Static: []game.StaticAbility{
			Labelled(KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && !target.Tapped && target.HasSubtype(source.TextSubtype("Dragon"))
			}, "hexproof"), "Untapped Dragons you control have hexproof."),
		},
	})
}
