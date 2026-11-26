package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fanatic of Xenagos — {1}{R}{G} Creature — Centaur Warrior 3/3 (Born of the Gods):
//
//	"Trample
//	 Tribute 1 (As this creature enters, an opponent of your choice may put a +1/+1
//	 counter on it.)
//	 When this creature enters, if tribute wasn't paid, it gets +1/+1 and gains haste
//	 until end of turn."
//
// One effect, one timestamp: the pump and the haste are a single until-end-of-turn
// record on the creature as it entered.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f1072af9-9e6f-45ce-8c4b-8e678e7da8ce",
		Name:            "Fanatic of Xenagos",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements:    effects.Tribute("Fanatic of Xenagos", 1),
		Triggered: []game.TriggeredAbility{
			effects.WhenEntersIfTributeWasntPaid("Fanatic of Xenagos — it gets +1/+1 and gains haste until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.B09SourceStillOnBattlefield(g, item) {
						return nil
					}
					return effects.UntilEndOfTurn(effects.NewContext(g, item), item.SourceCardID, nil,
						"Fanatic of Xenagos — +1/+1 and haste until end of turn",
						game.ModifyPTMod(1, 1), game.AddKeywordsMod("haste"))
				}),
		},
	})
}
