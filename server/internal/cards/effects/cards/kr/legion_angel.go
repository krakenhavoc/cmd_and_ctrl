package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Legion Angel — Creature — Angel Warrior {2}{W}{W}, 4/3 (Zendikar Rising,
// #5508):
//
//	"Flying
//	 When this creature enters, you may reveal a card you own named Legion
//	 Angel from outside the game and put it into your hand."
//
// Death Wish's TakeFromOutside (ADR 0191, ADR 0198) narrowed to the name
// "Legion Angel", revealed as it is taken: from the sideboard when the deck
// seated one, otherwise from the card database. The player may take nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "843c9962-dee1-449d-8106-6383df32bcf6",
		Name:            "Legion Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Legion Angel — you may reveal a card named Legion Angel you own from outside the game and put it into your hand",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					return effects.TakeFromOutside{
						Pick:     legionAngelTake,
						Question: "Legion Angel — reveal a card named Legion Angel you own from outside the game and put it into your hand, or none",
					}.Apply(ctx)
				}))),
		},
	})
}

// legionAngelTake is the keyed pick: only a card named Legion Angel, revealed.
var legionAngelTake = effects.RegisterOutsideTake("choose-cards/legion-angel-take",
	func(c game.Card) bool { return c.Name == "Legion Angel" }, true)
