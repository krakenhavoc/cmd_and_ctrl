package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Measure of Wickedness — Enchantment {3}{B} (Saviors of Kamigawa, #5601):
//
//	"At the beginning of your end step, sacrifice this enchantment and you
//	 lose 8 life.
//	 Whenever another card is put into your graveyard from anywhere, target
//	 opponent gains control of this enchantment."
//
// "You" is whoever controls it. The end-step ability loses the life whether or
// not the sacrifice happens. The graveyard trigger is Jinxed Idol's hand-over,
// indefinite (CR 611.2a) and applied only while this is still the permanent it
// came from; it fires once for each card that reaches the controller's
// graveyard (a token is not a card, CR 108.2), from any zone, and not for
// this enchantment itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8f1d65ae-7ce5-4bb6-8c00-eace25ff837c",
		Name:         "Measure of Wickedness",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Measure of Wickedness — sacrifice this enchantment and you lose 8 life", effects.OldB586SacrificeThisAndLose(8)),
			effects.Targeting(effects.OnAny([]game.EventKind{game.EventZoneMove}, effects.OldB586AnotherCardPutIntoYourGraveyard,
				"Measure of Wickedness — target opponent gains control of this enchantment",
				effects.OldB586TargetOpponentGainsControlOfThis("Measure of Wickedness — target opponent gains control of it")),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
