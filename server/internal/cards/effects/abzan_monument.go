package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Abzan Monument — Artifact {2} (Tarkir: Dragonstorm, #4463):
//
//	"When this artifact enters, search your library for a basic Plains,
//	 Swamp, or Forest card, reveal it, put it into your hand, then
//	 shuffle.
//	 {1}{W}{B}{G}, {T}, Sacrifice this artifact: Create an X/X white
//	 Spirit creature token, where X is the greatest toughness among
//	 creatures you control. Activate only as a sorcery."
//
// The entry search is the Overlook lands' basic-of-three predicate
// (IsBasicLandOfAnySubtype) sent to the hand and revealed. X is counted
// as the ability resolves (CR 608.2h), after the Monument has been
// sacrificed as part of the cost. With no creature, X is 0 and the 0/0
// Spirit is created and dies to state-based actions, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "93563b7b-5685-4832-b164-0701dbf3cf18",
		Name:         "Abzan Monument",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Abzan Monument — search for a basic Plains, Swamp, or Forest card", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: IsBasicLandOfAnySubtype("Plains", "Swamp", "Forest"),
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Abzan Monument — a basic Plains, Swamp, or Forest card",
				}.Apply(NewContext(g, item))
			}),
		},
		Activated: []ActivatedAbility{{
			Label: "{1}{W}{B}{G}, {T}, Sacrifice this artifact: Create an X/X white Spirit creature token, " +
				"where X is the greatest toughness among creatures you control. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{1}{W}{B}{G}"), TapCost(), SacrificeThis()),
			SorcerySpeed: true,
			Effect:       abzanMonumentSpirit,
		}},
	})
}

// abzanMonumentSpirit creates the X/X white Spirit, X read now.
func abzanMonumentSpirit(g *game.Game, item *game.StackItem) error {
	x := glintWeaverGreatestToughnessControlledBy(g, item.Controller)
	return CreateToken{Template: game.Card{
		Name:      "Spirit",
		TypeLine:  "Token Creature — Spirit",
		Colors:    []string{"W"},
		Power:     x,
		Toughness: x,
	}, N: 1}.Apply(NewContext(g, item))
}
