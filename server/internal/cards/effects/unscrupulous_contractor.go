package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:       "00ff5598-2f7f-4e28-8b63-5ae1ce18f0c5",
		Name:           "Unscrupulous Contractor",
		Completeness:   CompletenessFull,
		SpecialActions: []game.SpecialAction{Plot("{2}{B}")},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Unscrupulous Contractor — you may sacrifice a creature", unscrupulousContractorSacrifice),
		},
	})
}

func unscrupulousContractorSacrifice(g *game.Game, item *game.StackItem) error {
	return maySacrificeOne(g, item, "Unscrupulous Contractor — sacrifice a creature so target player draws two cards and loses 2 life?",
		Creature(), false, unscrupulousContractorSacrificed)
}

// unscrupulousContractorSacrificed is the "when you do". An on-disk key:
// never renamed, never reused.
var unscrupulousContractorSacrificed = RegisterSacrificedOne("prompt-run/unscrupulous-contractor-sacrifice", func(ctx *Context, _ uuid.UUID) error {
	return WhenYouDo("Unscrupulous Contractor — target player draws two cards and loses 2 life", unscrupulousContractorDrawBody).Apply(ctx)
})

// unscrupulousContractorDrawBody is the reflexive trigger. An on-disk
// key: never renamed, never reused.
var unscrupulousContractorDrawBody = game.ReflexiveBody("unscrupulous-contractor/draw-and-lose",
	simpleBody(func(g *game.Game, item *game.StackItem) error {
		return b32TargetPlayerDrawsAndLosesLife(item, NewContext(g, item), 2, 2)
	}),
	constTargets(func() *game.TargetSpec { return TargetPlayer("target player") }))
