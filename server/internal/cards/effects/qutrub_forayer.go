package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Qutrub Forayer — Creature — Zombie Horror {2}{B}, 3/2:
//
//	"When this creature enters, choose one —
//	 • Destroy target creature that was dealt damage this turn.
//	 • Exile up to two target cards from a single graveyard."
//
// #1807. The single-graveyard clause shipped with ADR 0106 §5 and the
// "was dealt damage this turn" predicate with Covert Cutpurse (#1855),
// so both halves this card was held for now exist. A modal enters
// trigger (CR 700.2, Entomber Exarch's shape): the mode and its target
// group are chosen as the trigger goes on the stack, and each bullet
// reads its own group (ModeTargets). The destroy bullet has no
// restriction on whose creature, so it can hit your own damaged
// creature, as printed. With no damaged creature on the table that
// bullet cannot be chosen (CR 603.3c), while the "up to" bullet always
// can.
//
// No simplification.
func init() {
	forayer := WhenThisEnters("Qutrub Forayer — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	forayer.Modes = ChooseOne(
		ModeDoing("Destroy target creature that was dealt damage this turn.",
			TargetCreature("target creature that was dealt damage this turn", DealtDamageThisTurn()),
			DestroyTheModesTarget),
		ModeDoing("Exile up to two target cards from a single graveyard.",
			upToNCardsFromASingleGraveyard(2),
			exileTheModesTargetCards),
	)
	Register(Spec{
		OracleID:     "500ccb61-83b0-4b1f-ada6-65850ed5a825",
		Name:         "Qutrub Forayer",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{forayer},
	})
}

// exileTheModesTargetCards is "exile up to N target cards from a single
// graveyard" as a modal bullet's body: every card of THIS bullet's
// group that is still a legal target (CR 608.2b), in one simultaneous
// move. Ebony Charm's second bullet reads item.Targets through
// LegalTargets because it is a spell with one targeted bullet; a
// trigger's bullet reads its own occurrence.
func exileTheModesTargetCards(_ *game.StackItem, ctx *Context, occ int) error {
	var ids []uuid.UUID
	for _, t := range ctx.ModeTargets(occ) {
		if t.Kind != game.TargetCard || t.ID == uuid.Nil || !ctx.IsTargetLegal(t) {
			continue
		}
		ids = append(ids, t.ID)
	}
	return exileCardsThen(ctx, ids, nil)
}
