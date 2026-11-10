package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Felothar, Dawn of the Abzan — Legendary Creature — Human Warrior {W}{B}{G},
// 3/3 (Tarkir: Dragonstorm, #4463):
//
//	"Trample
//	 Whenever Felothar enters or attacks, you may sacrifice a nonland
//	 permanent. When you do, put a +1/+1 counter on each creature you
//	 control."
//
// One ability with two conditions (WhenThisEntersOrAttacks). The sacrifice
// is Gastal Blockbuster's: asked as the trigger resolves over the
// nonland permanents you control (Felothar himself among them, since the
// text doesn't say "another"), with a floor of zero for the "may". Only
// once something has really been sacrificed is the counters created, as a
// reflexive trigger (CR 603.12); it counts the creatures you control as
// it resolves, so Felothar's own sacrifice leaves him out.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fd666669-ad23-4e41-b78c-1cd7b82eb8b4",
		Name:            "Felothar, Dawn of the Abzan",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Felothar, Dawn of the Abzan — you may sacrifice a nonland permanent",
				func(g *game.Game, item *game.StackItem) error {
					return maySacrificeOne(g, item, "Felothar, Dawn of the Abzan — sacrifice a nonland permanent to put a +1/+1 counter on each creature you control?",
						Nonland(), false, felotharSacrificed)
				}),
		},
	})
}

// felotharSacrificed is the "when you do". An on-disk key: never renamed,
// never reused.
var felotharSacrificed = RegisterSacrificedOne("prompt-run/felothar-sacrifice", func(ctx *Context, _ uuid.UUID) error {
	return WhenYouDo("Felothar, Dawn of the Abzan — put a +1/+1 counter on each creature you control", felotharCountersBody).Apply(ctx)
})

// felotharCountersBody is the reflexive trigger. An on-disk key: never
// renamed, never reused.
var felotharCountersBody = game.ReflexiveBody("felothar-dawn-of-the-abzan/counters",
	simpleBody(b11PutCounterOnEachCreatureYouControl), nil)
