package effects

import (
	"github.com/google/uuid"

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
// "Another" is the b03NotNamed clause rather than NotSelf: a target
// spec is built once at Register and has no instance to compare
// against. It is also why a token copy of Reflection cannot target the
// original, which is the same (slightly over-strict) reading every
// other "another target creature you control" in the catalog takes.
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
	Register(Spec{
		OracleID:     fableOfTheMirrorBreakerOracleID + "#1",
		Name:         "Reflection of Kiki-Jiki",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{1}, {T}: Create a token that's a copy of another target nonlegendary creature you control, except it has haste. Sacrifice it at the beginning of the next end step.",
			Cost:  Plus(ManaCost("{1}"), TapCost()),
			Targets: TargetCreature("another target nonlegendary creature you control",
				YouControl(), Not(Legendary()), b03NotNamed("Reflection of Kiki-Jiki")),
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
	ctx := NewContext(g, item)
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
	cursor := b25LastEventSeq(g)
	if err := (CreateTokenCopy{
		Controller: item.Controller,
		Copy:       copyOf,
		N:          1,
		Except:     reflectionOfKikiJikiHasteException,
	}).Apply(ctx); err != nil {
		return err
	}
	tokens := b27TokensCreatedByAfter(g, item.Controller, cursor)
	if len(tokens) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label:  "Reflection of Kiki-Jiki — sacrifice the token",
		Cards:  tokens,
		Effect: b33SacrificeListedCards,
	}.Apply(ctx)
}

// reflectionOfKikiJikiHasteException is the "except it has haste"
// clause. Haste rides on the TEMPLATE, as a printed keyword, because
// CreateTokenForEffect copies the template wholesale and printed
// keywords are read straight off the instance — a token has no catalog
// key to hang a layer-6 grant off (see fable_of_the_mirror_breaker.go
// and #521).
func reflectionOfKikiJikiHasteException(t *game.Card) {
	for _, kw := range t.Keywords {
		if kw == "haste" {
			return
		}
	}
	t.Keywords = append(t.Keywords, "haste")
}
