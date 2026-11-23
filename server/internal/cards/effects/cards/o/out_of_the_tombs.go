package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Out of the Tombs — Enchantment {2}{B} (Warhammer 40,000 Commander, S82 GAP
// batch 196, #5948):
//
//	"At the beginning of your upkeep, put two eon counters on this enchantment,
//	 then mill cards equal to the number of eon counters on it.
//	 If you would draw a card while your library has no cards in it, instead
//	 return a creature card from your graveyard to the battlefield. If you can't,
//	 you lose the game."
//
// The upkeep trigger adds two counters (only while it is still the same
// permanent) and then mills as many cards as it now has. The draw clause is
// Laboratory Maniac's window (CR 614.11), Ormos's draw-instead: for each draw
// of the controller's that finds the library empty, the creature card is chosen
// from their graveyard and put onto the battlefield, and the rest of the draw
// goes on; with no creature card to return they lose the game (CR 104.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "473a7998-669d-4181-83fe-f178a2e781f0",
		Name:         "Out of the Tombs",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.GapB196OutOfTheTombsReplacement()},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Out of the Tombs — put two eon counters on this enchantment, then mill cards equal to the number of eon counters on it",
				effects.GapB196OutOfTheTombsUpkeep),
		},
	})
}
