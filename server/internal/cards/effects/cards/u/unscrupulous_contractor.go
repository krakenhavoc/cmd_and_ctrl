package cardsu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Unscrupulous Contractor — Creature — Human Assassin {2}{B}, 3/2
// (Outlaws of Thunder Junction, #4468):
//
//	"When this creature enters, you may sacrifice a creature. When you
//	 do, target player draws two cards and loses 2 life.
//	 Plot {2}{B} (You may pay {2}{B} and exile this card from your hand.
//	 Cast it as a sorcery on a later turn without paying its mana cost.
//	 Plot only as a sorcery.)"
//
// Gastal Blockbuster's shape: the sacrifice is asked as the trigger
// resolves, over the creatures you control then (the Contractor itself
// among them, since the text doesn't say "another"), with a floor of
// zero for the "may". Only once something has really been sacrificed is
// the reflexive trigger created, and it chooses its player as it goes on
// the stack (CR 603.12). That player draws two and loses 2 life, Blood
// Pact's body. Plot is the ordinary special action (CR 702.170).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "00ff5598-2f7f-4e28-8b63-5ae1ce18f0c5",
		Name:           "Unscrupulous Contractor",
		Completeness:   effects.CompletenessFull,
		SpecialActions: []game.SpecialAction{effects.Plot("{2}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Unscrupulous Contractor — you may sacrifice a creature", unscrupulousContractorSacrifice),
		},
	})
}

func unscrupulousContractorSacrifice(g *game.Game, item *game.StackItem) error {
	return effects.MaySacrificeOne(g, item, "Unscrupulous Contractor — sacrifice a creature so target player draws two cards and loses 2 life?",
		effects.Creature(), false, unscrupulousContractorSacrificed)
}

// unscrupulousContractorSacrificed is the "when you do". An on-disk key:
// never renamed, never reused.
var unscrupulousContractorSacrificed = effects.RegisterSacrificedOne("prompt-run/unscrupulous-contractor-sacrifice", func(ctx *effects.Context, _ uuid.UUID) error {
	return effects.WhenYouDo("Unscrupulous Contractor — target player draws two cards and loses 2 life", unscrupulousContractorDrawBody).Apply(ctx)
})

// unscrupulousContractorDrawBody is the reflexive trigger. An on-disk
// key: never renamed, never reused.
var unscrupulousContractorDrawBody = game.ReflexiveBody("unscrupulous-contractor/draw-and-lose",
	effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
		return effects.B32TargetPlayerDrawsAndLosesLife(item, effects.NewContext(g, item), 2, 2)
	}),
	effects.ConstTargets(func() *game.TargetSpec { return effects.TargetPlayer("target player") }))
