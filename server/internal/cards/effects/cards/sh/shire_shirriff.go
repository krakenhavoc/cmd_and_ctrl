package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shire Shirriff — Creature — Halfling Soldier {1}{W}, 2/2 (The Lord of the
// Rings: Tales of Middle-earth Commander):
//
//	"Vigilance
//	 When this creature enters, you may sacrifice a token. When you do,
//	 exile target creature an opponent controls until this creature leaves
//	 the battlefield."
//
// Vigilance is printed. The "you may sacrifice a token" is Buzzard-Wasp
// Colony's may-sacrifice (any token its controller controls, chosen as the
// ability resolves, and nobody is asked with no token to give); the "when
// you do" is a CR 603.12 reflexive trigger whose target is chosen as it
// goes on the stack, and the exile lasts until the Shirriff leaves the
// battlefield (Angel of Sanctions's duration), returning the creature then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "14f36157-f55f-495d-9ffe-f38135a59750",
		Name:            "Shire Shirriff",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Shire Shirriff — you may sacrifice a token; when you do, exile target creature an opponent controls until this creature leaves the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, "Shire Shirriff — sacrifice a token to exile target creature an opponent controls?",
						effects.IsTokenPredicate(), false, shireShirriffSacrificed)
				}),
		},
	})
}

var shireShirriffSacrificed = effects.RegisterSacrificedOne("shire-shirriff/exile", func(ctx *effects.Context, _ uuid.UUID) error {
	return effects.WhenYouDo("Shire Shirriff — exile target creature an opponent controls until this creature leaves the battlefield", shireShirriffExileBody).Apply(ctx)
})

var shireShirriffExileBody = game.ReflexiveBody("shire-shirriff/exile-body",
	effects.SimpleBody(effects.ExileChosenTargetUntilThisLeaves("Shire Shirriff — the exiled creature returns when Shire Shirriff leaves the battlefield")),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())
	}))
