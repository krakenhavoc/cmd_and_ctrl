package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Horobi's Whisper — Instant — Arcane {1}{B}{B}:
//
//	"If you control a Swamp, destroy target nonblack creature.
//	 Splice onto Arcane—Exile four cards from your graveyard. (As you cast an
//	 Arcane spell, you may reveal this card from your hand and pay its splice
//	 cost. If you do, add this card's effects to that spell.)"
//
// The target is chosen whether or not the caster controls a Swamp; the
// condition is read as the spell resolves ("you" is the controller of the
// spell this text is part of, CR 702.47c). The splice cost is four cards
// from the caster's graveyard, named as the spell is cast and exiled with
// it on the stack (#4055, ADR 0202's 2026-10-19 amendment); a graveyard of
// fewer than four cards cannot pay it (CR 118.3).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2447f1f4-3569-4afb-a119-2ad9778eaf39",
		Name:         "Horobi's Whisper",
		Completeness: effects.CompletenessFull,
		Splice:       effects.SpliceOntoArcaneCost(effects.SpliceExileFromGraveyard(4, "four cards")),
		Targets:      effects.TargetCreature("target nonblack creature", effects.NonBlack()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if !effects.ControlsA(ctx.TextSubtype("Swamp"))(ctx.Game, ctx.Controller()) {
				return nil
			}
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			return effects.DestroyTarget{Target: t.ID}.Apply(ctx)
		},
	})
}
