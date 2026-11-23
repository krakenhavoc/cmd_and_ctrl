package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "93563b7b-5685-4832-b164-0701dbf3cf18",
		Name:         "Abzan Monument",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Abzan Monument — search for a basic Plains, Swamp, or Forest card", func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.IsBasicLandOfAnySubtype("Plains", "Swamp", "Forest"),
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Abzan Monument — a basic Plains, Swamp, or Forest card",
				}.Apply(effects.NewContext(g, item))
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{1}{W}{B}{G}, {T}, Sacrifice this artifact: Create an X/X white Spirit creature token, " +
				"where X is the greatest toughness among creatures you control. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{1}{W}{B}{G}"), effects.TapCost(), effects.SacrificeThis()),
			SorcerySpeed: true,
			Effect:       abzanMonumentSpirit,
		}},
	})
}

// abzanMonumentSpirit creates the X/X white Spirit, X read now.
func abzanMonumentSpirit(g *game.Game, item *game.StackItem) error {
	x := effects.GlintWeaverGreatestToughnessControlledBy(g, item.Controller)
	return effects.CreateToken{Template: game.Card{
		Name:      "Spirit",
		TypeLine:  "Token Creature — Spirit",
		Colors:    []string{"W"},
		Power:     x,
		Toughness: x,
	}, N: 1}.Apply(effects.NewContext(g, item))
}
