package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dutiful Replicator — Artifact Creature — Assembly-Worker {3}, 3/2
// (Dominaria Remastered):
//
//	"When this creature enters, you may pay {1}. When you do, create a
//	 token that's a copy of target token you control not named Dutiful
//	 Replicator."
//
// Ancestral Katana's shape: the may-pay is the entry trigger's; the paid
// branch creates a CR 603.12 reflexive trigger whose target (a token you
// control not named Dutiful Replicator) is chosen as it goes on the stack,
// so with no such token it is removed (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "46670c03-d7eb-489f-ba1a-c7f805c486d2",
		Name:         "Dutiful Replicator",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Dutiful Replicator — you may pay {1}; when you do, create a token that's a copy of target token you control not named Dutiful Replicator",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Dutiful Replicator — pay {1}? When you do, copy target token you control not named Dutiful Replicator.",
						Then:     dutifulReplicatorPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var dutifulReplicatorPaid = effects.PayUnlessBranches("pay-unless/dutiful-replicator-copy",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.ReflexiveTrigger{
			Label: "Dutiful Replicator — create a token that's a copy of target token you control not named Dutiful Replicator",
			Body:  dutifulReplicatorCopyBody,
		}.Apply(ctx)
	}, nil)

// dutifulReplicatorCopyBody copies the reflexive trigger's target. An
// on-disk key.
var dutifulReplicatorCopyBody = game.ReflexiveBody("s82-gap-b145/dutiful-replicator-copy",
	effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		id, ok := effects.B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return effects.CreateTokenCopy{Controller: item.Controller, Copy: id, N: 1}.Apply(ctx)
	}),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetPermanent("target token you control not named Dutiful Replicator", effects.IsTokenPredicate(), effects.YouControl(),
			func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.Name != "Dutiful Replicator" })
	}))
