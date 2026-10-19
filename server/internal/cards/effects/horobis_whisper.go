package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "2447f1f4-3569-4afb-a119-2ad9778eaf39",
		Name:         "Horobi's Whisper",
		Completeness: CompletenessFull,
		Splice:       SpliceOntoArcaneCost(SpliceExileFromGraveyard(4, "four cards")),
		Targets:      TargetCreature("target nonblack creature", NonBlack()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if !ControlsA(ctx.TextSubtype("Swamp"))(ctx.Game, ctx.Controller()) {
				return nil
			}
			t, ok := firstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			return DestroyTarget{Target: t.ID}.Apply(ctx)
		},
	})
}
