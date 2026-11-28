package cardslu

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Megatherium — Creature — Beast {2}{G}, 4/4 (Mercadian Masques, #5619):
//
//	"Trample
//	 When this creature enters, sacrifice it unless you pay {1} for each card
//	 in your hand."
//
// Rupture Spire's enters-and-pay shape, with the cost counted as the trigger
// resolves: {1} for each card in its controller's hand then (with an empty
// hand there is nothing to pay and the Beast stays). A Megatherium that has
// left, or come back as a new object, is not sacrificed (CR 400.7).
//
// No simplification.
func init() {
	const name = "Megatherium"
	effects.Register(effects.Spec{
		OracleID:        "b98123b2-c913-41b9-bb24-10a287d24bca",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(name+" — sacrifice it unless you pay {1} for each card in your hand",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) {
						return nil
					}
					p := g.PlayerByIDForEffect(item.Controller)
					if p == nil || p.Hand.Size() == 0 {
						return nil
					}
					cost := "{" + strconv.Itoa(p.Hand.Size()) + "}"
					return effects.UpkeepPayUnless{
						Chooser:  item.Controller,
						Cost:     cost,
						Question: name + " — pay " + cost + " or sacrifice " + name + "?",
						Else:     effects.UpkeepSacrificeSource,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
