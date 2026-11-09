package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Teferi, Mage of Zhalfir — Legendary Creature — Human Wizard
// {2}{U}{U}{U}, 3/4 (Final Fantasy, #4462):
//
//	"Flash
//	 Creature cards you own that aren't on the battlefield have flash.
//	 Each opponent can cast spells only any time they could cast a
//	 sorcery."
//
// The grant is Spec.CastTimings (Vedalken Orrery's seam) narrowed to creature
// cards the CASTER owns (CastTimingRule.OwnedByCaster, CR 108.3): from a hand,
// a graveyard, exile or the command zone (a cast never begins on the
// battlefield), and not a creature card you cast from an opponent's
// graveyard. Because the board says it, the card may be cast at instant
// speed under the whole cast-timing read, and suspended at instant speed too
// (CR 702.62a, the ruling): the suspend special action asks the same read.
// Split second still stops both (the ruling).
//
// The opponent restriction is Teferi, Time Raveler's (CR 101.2): every
// opponent's cast must be one that could be a sorcery's, so on their own
// turn in a main phase with an empty stack, even an instant, and even a spell
// another effect lets them cast at instant speed. Teferi's own flash is the
// printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2b6cbd55-5adc-4a34-ac45-8531705a7ee1",
		Name:            "Teferi, Mage of Zhalfir",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		CastTimings: []game.CastTimingRule{
			finB12OwnedCreatureCardsHaveFlash("Creature cards you own that aren't on the battlefield have flash."),
			OpponentsCastAtSorcerySpeed(),
		},
	})
}
