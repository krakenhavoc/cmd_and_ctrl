package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cosmos Charger — Creature — Horse Spirit {3}{U}, 3/3 (Kaldheim):
//
//	"Flash
//	 Flying
//	 Foretelling cards from your hand costs {1} less and can be done on any
//	 player's turn.
//	 Foretell {2}{U} (During your turn, you may pay {2} and exile this card
//	 from your hand face down. Cast it on a later turn for its foretell cost.)"
//
// The discount is Ranar the Ever-Watchful's partition: a reduction of generic
// mana on the foretell special action (CostModifier.SpecialActions), so the
// {2} becomes {1}. "Any player's turn" is Spec.ForetellOnAnyTurn, read by the
// special-action timing table (the engine, the enumerator and the view agree);
// the card still can't be cast the turn it was foretold (CR 702.143a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:          "603e518c-8f12-432c-ae03-790db5be77bf",
		Name:              "Cosmos Charger",
		Completeness:      effects.CompletenessFull,
		PrintedKeywords:   []string{"flash", "flying"},
		SpecialActions:    []game.SpecialAction{effects.Foretell("{2}{U}")},
		ForetellOnAnyTurn: true,
		CostModifiers: []game.CostModifier{
			effects.SpecialActionCostsLess(1, "Foretelling cards from your hand costs {1} less.",
				effects.YourSpell(), effects.ASpecialActionOfKind(game.SpecialActionForetell)),
		},
	})
}
