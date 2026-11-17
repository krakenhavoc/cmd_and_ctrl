package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hangar Scrounger — Creature — Dwarf Pilot {2}{R}, 2/1 (March of the
// Machine, #5477):
//
//	"Backup 1 (When this creature enters, put a +1/+1 counter on target
//	 creature. If that's another creature, it gains the following ability
//	 until end of turn.)
//	 Whenever this creature becomes tapped, you may discard a card. If you
//	 do, draw a card."
//
// Backup (CR 702.165) grants the tap trigger, a bundle in Spec.Grants that
// the Scrounger also has itself. "Becomes tapped" is the tap event and an
// attack that taps it (ThisBecameTapped), as Attentive Sunscribe's; the
// rummage is chosen as the ability resolves and draws only if a card was
// discarded.
//
// No simplification.
const hangarScroungerGrant = "hangar-scrounger/tapped-rummage"

func init() {
	Register(Spec{
		OracleID:     "d1b74fb9-8ef3-40be-bc79-45ffb4c597eb",
		Name:         "Hangar Scrounger",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       hangarScroungerGrant,
			Triggered: []game.TriggeredAbility{hangarScroungerTapped()},
			Text:      "Whenever this creature becomes tapped, you may discard a card. If you do, draw a card.",
		}},
		Triggered: []game.TriggeredAbility{
			Backup("Hangar Scrounger", 1, hangarScroungerGrant),
			hangarScroungerTapped(),
		},
	})
}

func hangarScroungerTapped() game.TriggeredAbility {
	return OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
		"Hangar Scrounger — you may discard a card; if you do, draw a card",
		tdmMayDiscardThenDrawOne("Hangar Scrounger — you may discard a card; if you do, draw a card"))
}
