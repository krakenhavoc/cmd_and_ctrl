package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thalakos Deceiver — Creature — Thalakos Wizard {3}{U}, 1/1 (Stronghold, S82
// OLD b749, #5627):
//
//	"Shadow
//	 Whenever this creature attacks and isn't blocked, you may sacrifice it. If
//	 you do, gain control of target creature. (This effect lasts
//	 indefinitely.)"
//
// Shadow is an engine keyword carried as printed data (CR 702.28). The trigger
// is Bone Dancer's "attacks and isn't blocked" (CR 509.1h) with a target,
// chosen as it goes on the stack (CR 603.3d), and Ogre-Head Helm's "you may
// sacrifice it. If you do": the question is asked as the trigger resolves and
// the control change hangs off the sacrifice actually happening, so a Deceiver
// that has already left does nothing. The target is checked again as the
// control change is made (CR 608.2b). The change has no stated duration, so it
// lasts indefinitely (CR 611.2a), as Dominate's does.
//
// No simplification.
func init() {
	const label = "Thalakos Deceiver — you may sacrifice it. If you do, gain control of target creature"
	effects.Register(effects.Spec{
		OracleID:        "c304777a-39a9-4a2a-8cec-8116afb34c5b",
		Name:            "Thalakos Deceiver",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"shadow"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBlockersDeclared},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttacksAndIsNotBlocked(ev, source, g)
			},
			Targets: effects.TargetCreature("target creature"),
			Key:     label,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if !effects.OnBattlefield(g, ctx.Source()) {
					return nil
				}
				if _, ok := effects.B16FirstLegalTargetCard(ctx); !ok {
					return nil
				}
				return effects.MayChoice{
					Question: "Thalakos Deceiver — sacrifice it to gain control of the target creature?",
					YesLabel: "Sacrifice",
					NoLabel:  "Decline",
					Then:     thalakosDeceiverAnswer,
				}.Apply(ctx)
			},
		}},
	})
}

// thalakosDeceiverSteal is "gain control of target creature" after the
// sacrifice. Keyed (CR 608.2b: a target that left is skipped).
var thalakosDeceiverSteal = effects.RegisterThenSteps("then-steps/thalakos-deceiver-steal",
	effects.Instruction(func(ctx *effects.Context) error {
		id, ok := effects.B16FirstLegalTargetCard(ctx)
		if !ok || id == uuid.Nil {
			return nil
		}
		return effects.GainControl{
			Target:     id,
			Controller: ctx.Controller(),
			Duration:   game.IndefiniteDuration(),
			Label:      "Thalakos Deceiver — gain control (no stated duration)",
		}.Apply(ctx)
	}),
)

// thalakosDeceiverAnswer is the "yes": sacrifice the Deceiver, and only if it
// went, take the creature. Keyed (#4291).
var thalakosDeceiverAnswer = effects.MayYes("confirm/thalakos-deceiver-sacrifice", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	return effects.SacrificePermanent{Target: ctx.Source(), ThenTail: ctx.IfItWentThen(thalakosDeceiverSteal)}.Apply(ctx)
})
