package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Enigma Thief — Creature — Sphinx Rogue {5}{U}{U}, 5/5 (Zendikar Rising
// Commander, #5509):
//
//	"Prowl {3}{U} (You may cast this spell for its prowl cost if you dealt
//	 combat damage to a player this turn with a Sphinx or Rogue.)
//	 Flying
//	 When this creature enters, for each opponent, return up to one target
//	 nonland permanent that player controls to its owner's hand."
//
// Prowl is the gated alternative cost of prowl_keyword.go (CR 702.76a): the
// spell's own creature types, as it is in hand, against the creatures you
// controlled that dealt combat damage to a player this turn. Its mana value
// stays seven whatever is paid (the ruling).
//
// The enters trigger is Bronzebeak Foragers' per-opponent "up to one"
// clauses, one per opponent in turn order (CR 603.3d), each re-checked as it
// resolves (CR 608.2b): a permanent that changed control in response is no
// longer controlled by the player its clause names and stays where it is,
// while the other picks still return (the ruling). They go to their owners'
// hands together.
//
// No simplification.
func init() {
	enters := effects.Targeting(
		effects.WhenThisEnters("Enigma Thief — for each opponent, return up to one target nonland permanent that player controls to its owner's hand",
			enigmaThiefBounce),
		effects.PerOpponentClauses(game.MaxPlayers-1, "nonland permanent", effects.Nonland()))
	effects.Register(effects.Spec{
		OracleID:         "d39a1e12-bb6b-4e7a-96a4-5dab98084716",
		Name:             "Enigma Thief",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{effects.Prowl("{3}{U}")},
		Triggered:        []game.TriggeredAbility{enters},
	})
}

// enigmaThiefBounce returns every still-legal pick to its owner's hand as
// one simultaneous event (one instruction, CR 608.2c).
func enigmaThiefBounce(g *game.Game, item *game.StackItem) error {
	ids := effects.LegalTargetIDs(effects.NewContext(g, item))
	if len(ids) == 0 {
		return nil
	}
	g.BounceCardsToHandForEffect(ids)
	return nil
}
