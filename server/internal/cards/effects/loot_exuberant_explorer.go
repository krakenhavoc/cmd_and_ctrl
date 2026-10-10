package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Loot, Exuberant Explorer — Legendary Creature — Beast Noble {2}{G},
// 1/4 (slice 296-m):
//
//	"You may play an additional land on each of your turns.
//	 {4}{G}{G}, {T}: Look at the top six cards of your library. You may
//	 reveal a creature card with mana value less than or equal to the
//	 number of lands you control from among them and put it onto the
//	 battlefield. Put the rest on the bottom in a random order."
//
// The land drop is Exploration's field, `Spec.AdditionalLandPlays`.
// The tap ability is Ureni of the Unwritten's shared sentence
// (LookAtTopThenMayPutOntoBattlefield): the six cards are LOOKED at,
// so only Loot's controller sees them, and the match predicate reads
// the number of lands the activator controls at resolution time —
// exactly the printed "mana value less than or equal to the number of
// lands you control", not a value frozen at announce.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:            "f6476884-f73f-465f-8e7f-ea312a5b306d",
		Name:                "Loot, Exuberant Explorer",
		Completeness:        CompletenessFull,
		AdditionalLandPlays: 1,
		Purpose:             game.Purpose{ExtraLandDrops: 1}, // #2678: the bot reads the extra drop
		Activated: []ActivatedAbility{{
			Label:   "{4}{G}{G}, {T}: Look at the top six cards of your library. You may reveal a creature card with mana value less than or equal to the number of lands you control from among them and put it onto the battlefield. Put the rest on the bottom in a random order.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{4}{G}{G}"), TapCost()),
			Effect: LookAtTopThenMayPutOntoBattlefield(6, lootCreatureWithinLandCount, 1,
				"Loot, Exuberant Explorer — put a creature card onto the battlefield"),
		}},
	})
}

// lootCreatureWithinLandCount is "a creature card with mana value less
// than or equal to the number of lands you control", read fresh
// against the activator's board at the moment the choice is offered.
func lootCreatureWithinLandCount(g *game.Game, caster uuid.UUID, c game.Card) bool {
	if !c.IsCreature() {
		return false
	}
	return c.ManaValue() <= b10LandsControlled(g, caster)
}
