package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reflection of Kiki-Jiki — "{1}, {T}: Create a token that's a copy of
// another target nonlegendary creature you control, except it has
// haste. Sacrifice it at the beginning of the next end step."
//
// The BACK face of Fable of the Mirror-Breaker, registered under
// "<oracle_id>#1". It reaches the battlefield only through chapter
// III's exile-and-return (ADR 0079's second verb), and because the
// returning permanent is a NEW object it is summoning sick that turn —
// so the {T} in this cost cannot be paid until the following one,
// which is what the printed card does.
//
// "Another" is object identity (effects.Another, CR 109.1): a token copy of
// Reflection can target the original, as printed.
//
// The delayed sacrifice is a CR 603.7 trigger scheduled onto the Game
// rather than a closure held by this ability: the ability is long gone
// by the end step, and a delayed trigger has to survive Clone and undo
// by sharing a package-level Effect with the snapshot.
//
// "Sacrifice IT" is the TOKEN, not this creature — a misreading that
// would make the card sacrifice itself every activation. The token's
// ID is read back off EventTokenCreated after the copy is made, which
// is the only place it exists: CreateTokenCopy mints the instance
// inside the engine and returns nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     effects.FableOfTheMirrorBreakerOracleID + "#1",
		Name:         "Reflection of Kiki-Jiki",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{1}, {T}: Create a token that's a copy of another target nonlegendary creature you control, except it has haste. Sacrifice it at the beginning of the next end step.",
			Cost:  effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.Another(effects.TargetCreature("another target nonlegendary creature you control",
				effects.YouControl(), effects.Not(effects.Legendary()))),
			Effect: reflectionOfKikiJikiCopyWithHaste,
		}},
	})
}

// reflectionOfKikiJikiCopyWithHaste makes the hasty copy and schedules
// its sacrifice.
//
// The target is re-read out of ctx.LegalTargets() rather than off
// item.Targets, so a creature that left in response is skipped (CR
// 608.2b) instead of producing a copy of something that is not there.
func reflectionOfKikiJikiCopyWithHaste(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var copyOf uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			copyOf = t.ID
			break
		}
	}
	if copyOf == uuid.Nil {
		return nil
	}
	cursor := effects.B25LastEventSeq(g)
	if err := (effects.CreateTokenCopy{
		Controller: item.Controller,
		Copy:       copyOf,
		N:          1,
		Except:     effects.TokenCopyGainsHaste,
	}).Apply(ctx); err != nil {
		return err
	}
	tokens := effects.B27TokensCreatedByAfter(g, item.Controller, cursor)
	if len(tokens) == 0 {
		return nil
	}
	return effects.ScheduleDelayedTrigger{
		Label: "Reflection of Kiki-Jiki — sacrifice the token",
		Cards: tokens,
		Body:  effects.SacrificeListedCardsBody,
	}.Apply(ctx)
}
