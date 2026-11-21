package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Angelic Intervention — Instant {1}{W} (March of the Machine):
//
//	"Target creature or planeswalker you control gains protection from
//	 colorless or from the color of your choice until end of turn. If
//	 it's a creature, put a +1/+1 counter on it."
//
// One prompt as the spell resolves, with six answers: the five colours
// and colorless (game.ColorPrompt.OrColorless, #5650). The permanent
// gains protection from the answer until end of turn, and then gets the
// counter if it is a creature as the spell resolves, so a planeswalker
// that is also a creature gets one and a plain planeswalker does not. A
// target that left in response is asked nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3d036f68-5f3a-4182-88f0-cdd685817dd0",
		Name:         "Angelic Intervention",
		Purpose:      effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "protection")), game.AnswerProtect|game.AnswerPump),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target creature or planeswalker you control", effects.Or(effects.Creature(), effects.Planeswalker()), effects.YouControl()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				then := ctx.ItemTail(angelicInterventionThen)
				then.Carry = []uuid.UUID{t.ID}
				effects.ChooseColorOrColorlessTail(game.ColorForProtection, ctx.Game, item.Controller, ctx.Source(),
					"Angelic Intervention — choose colorless or a color", then)
				return nil
			}
			return nil
		},
	})
}

// angelicInterventionThen is the rest of Angelic Intervention once the
// quality is known: the carried permanent gains protection from it until
// end of turn, then gets a +1/+1 counter if it is a creature. An on-disk
// key.
var angelicInterventionThen = effects.CardTail("then/angelic-intervention-protection-then-counter", func(ctx *effects.Context, o game.TailOutcome) error {
	token := game.ProtectionFromColorOrColorless(o.Tail.Word)
	if token == "" || len(o.Tail.Carry) == 0 {
		return nil
	}
	target := o.Tail.Carry[0]
	if err := (effects.GrantKeywordUntilEOT{
		Target:   target,
		Keywords: []string{token},
		Label:    "Angelic Intervention — " + token,
	}).Apply(ctx); err != nil {
		return err
	}
	if c, ok := ctx.Game.LookupCardForEffect(target); !ok || !c.IsCreature() {
		return nil
	}
	return effects.AddCounter{Target: target, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
})
