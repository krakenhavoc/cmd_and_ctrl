package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scholar of the Lost Trove — Creature — Sphinx {5}{U}{U}, 5/5
// (Jumpstart 2020):
//
//	"Flying
//	 When this creature enters, you may cast target instant, sorcery, or
//	 artifact card from your graveyard without paying its mana cost. If an
//	 instant or sorcery spell cast this way would be put into your
//	 graveyard, exile it instead."
//
// Torrential Gearhulk's enters trigger over a wider target (an instant,
// sorcery or artifact card you own in your graveyard), with Bilbo, Thief
// in the Night's narrower replacement (jmpB9FreeCastTargetExilingInstants-
// AndSorceries). The cast is a free ({0}, X locked at 0), any-timing
// (CR 608.2g), cast-only permission over the target, offered during the
// resolution; declining leaves the card where it is. An instant or sorcery
// spell cast this way is exiled instead of going to your graveyard whether
// it resolves, is countered or fizzles; an artifact spell resolves onto the
// battlefield as usual and, if countered, goes to the graveyard. The target
// is chosen as the trigger goes on the stack and re-checked as it resolves
// (CR 608.2b).
//
// No simplification. Deck request #4894 (Jumpstart Azorius).
func init() {
	enters := WhenThisEnters("Scholar of the Lost Trove — you may cast target instant, sorcery, or artifact card from your graveyard without paying its mana cost",
		jmpB9FreeCastTargetExilingInstantsAndSorceries("Scholar of the Lost Trove"))
	enters.Targets = TargetCardInGraveyard("target instant, sorcery, or artifact card from your graveyard",
		Or(Instant(), Sorcery(), Artifact()), YouOwn())
	Register(Spec{
		OracleID:        "5f1e9098-f554-4505-974b-cef4b4b7b23d",
		Name:            "Scholar of the Lost Trove",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}
