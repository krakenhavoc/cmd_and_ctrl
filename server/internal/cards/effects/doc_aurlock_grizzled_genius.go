package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Doc Aurlock, Grizzled Genius — Legendary Creature — Bear Druid {G}{U},
// 2/2 (Outlaws of Thunder Junction, #4468):
//
//	"Spells you cast from your graveyard or from exile cost {2} less to
//	 cast.
//	 Plotting cards from your hand costs {2} less."
//
// Two cost modifiers. The first is a spell reduction on the zone the cast
// comes from (flashback, escape, a plotted card, an impulse-exiled card are
// all matched; hand and command zone are not), the CR 601.2f generic
// reduction. The second is the special-action twin (SpecialActionCostsLess,
// #1319): it prices the plot action (CR 702.170a) and only the controller's
// own, so an opponent's plotting is not reduced. Plotting a card that is
// then cast from exile is reduced a second time, by the first modifier,
// as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3ca32660-bf5f-4398-bef6-a0ec8bc767cf",
		Name:         "Doc Aurlock, Grizzled Genius",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(2, "Spells you cast from your graveyard or from exile cost {2} less to cast.",
				YourSpell(), CastFromGraveyardOrExile()),
			SpecialActionCostsLess(2, "Plotting cards from your hand costs {2} less.",
				ASpecialActionOfKind(game.SpecialActionPlot), b35YourSpecialAction()),
		},
	})
}
