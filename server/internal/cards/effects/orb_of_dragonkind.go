package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Orb of Dragonkind — Artifact {1}{R}:
//
//	"{1}, {T}: Add two mana in any combination of colors. Spend this
//	 mana only to cast Dragon spells or activate abilities of Dragons.
//	 {R}, {T}, Sacrifice this artifact: Look at the top seven cards of
//	 your library. You may reveal a Dragon card from among them and put
//	 it into your hand. Put the rest on the bottom of your library in a
//	 random order."
//
// The first ability is a mana ability with a mana cost (a Signet's
// shape); each of the two mana is any colour, chosen as it is added, and
// can pay only for a spell with the Dragon creature type or an ability of
// a Dragon source. The second is the look-and-take of Horn of the Mark: a
// card is a Dragon card when its creature type is Dragon, not when its
// name says so.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d9c36ed5-629c-4709-bff5-2f10fcde608a",
		Name:         "Orb of Dragonkind",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Mana: "{1}", Tap: true},
			Produced: AnyCombinationOfColors(2),
			Label:    "{1}, {T}: Add two mana in any combination of colors. Spend this mana only to cast Dragon spells or activate abilities of Dragons.",
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictCast, game.ManaRestrictSubtype("Dragon")},
				[]string{game.ManaRestrictActivate, game.ManaRestrictSubtype("Dragon")},
			)},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{R}, {T}, Sacrifice this artifact: Look at the top seven cards of your library. You may reveal a Dragon card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.",
			Cost:    Plus(ManaCost("{R}"), TapCost(), SacrificeThis()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: LookAtTopThenMayTakeToHand(7, AnySubtype("Dragon"), 1,
				"Orb of Dragonkind — reveal a Dragon card and put it into your hand"),
		}},
	})
}
