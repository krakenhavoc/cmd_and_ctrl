package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cyclopean Giant — Creature — Zombie Giant {2}{B}{B}, 4/2:
//
//	"When this creature dies, target land becomes a Swamp. Exile this
//	 card."
//
// ADR 0109 §1 decision 5 (#1881): CR 305.7's type set with no duration,
// lasting until the game ends (CR 611.2a), pinned to the land (CR 400.7).
// The land's land types are replaced by Swamp (other subtypes stay, CR
// 205.1a), it loses its rules-text abilities and taps for {B} (CR 305.6).
// Then the Giant's card is exiled from the graveyard it died into; one
// that has already left it is not followed. A trigger whose target land is
// gone by resolution does not resolve at all (CR 608.2b), so the card then
// stays in the graveyard.
//
// No simplification.
func init() {
	swamp := TargetLandBecomesIndefinitely("Cyclopean Giant", "Swamp")
	t := WhenThisDies("Cyclopean Giant — target land becomes a Swamp, then exile this card",
		func(g *game.Game, item *game.StackItem) error {
			if err := swamp(g, item); err != nil {
				return err
			}
			return ExileTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
		})
	t.Targets = TargetPermanent("target land", Land())
	Register(Spec{
		OracleID:     "d02427fa-3ef0-484f-acad-63a1d5218727",
		Name:         "Cyclopean Giant",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
