package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gwen Stacy // Ghost-Spider — a modal double-faced card whose front
// face transforms (#3987, found finishing #3192).
//
//	Gwen Stacy                          {1}{R}   2/1
//	Legendary Creature — Human Performer Hero
//	  When Gwen Stacy enters, exile the top card of your library. You
//	  may play that card for as long as you control this creature.
//	  {2}{U}{R}{W}: Transform Gwen Stacy. Activate only as a sorcery.
//
// The back face is ghost_spider.go, registered under this oracle ID
// plus "#1" (game.CatalogKey).
//
// "For as long as you control this creature" is the source-bound
// window Lightning, Security Sergeant uses (CR 611.2b): it names the
// Gwen Stacy OBJECT, so it survives the transform (CR 712.18, the same
// object) and ends for good when she leaves or changes hands. A Gwen
// already gone as the trigger resolves still exiles the card, which is
// then never playable.
//
// The transform is TransformThis: a modal double-faced permanent can be
// transformed by an instruction that says so (CR 712.9, the layout
// game.CanTransform admits). Cast as Ghost-Spider, the card never had
// the front face's trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     gwenStacyOracle,
		Name:         "Gwen Stacy",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Gwen Stacy — exile the top card of your library; you may play it for as long as you control her",
				exileTopPlayableWhileYouControlThis),
		},
		Activated: []ActivatedAbility{{
			Label:        "{2}{U}{R}{W}: Transform Gwen Stacy. Activate only as a sorcery.",
			Cost:         ManaCost("{2}{U}{R}{W}"),
			SorcerySpeed: true,
			Effect:       Do(TransformThis{}),
		}},
	})
}

// gwenStacyOracle is shared with the back face (ghost_spider.go).
const gwenStacyOracle = "5f143d87-35fb-40e3-94ae-f85cfff9adb5"
