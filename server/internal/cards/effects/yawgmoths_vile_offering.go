package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Yawgmoth's Vile Offering — Legendary Sorcery {4}{B}:
//
//	"(You may cast a legendary sorcery only if you control a legendary
//	 creature or planeswalker.)
//	 Put up to one target creature or planeswalker card from a graveyard
//	 onto the battlefield under your control. Destroy up to one target
//	 creature or planeswalker. Exile Yawgmoth's Vile Offering."
//
// The cast gate is CR 205.4e's LegendarySorcery, checked at announce and
// never at resolution (Urza's Ruinous Blast).
//
// Two SEPARATE target clauses, each "up to one" (Min 0, Max 1): slot 0 is
// the card in a graveyard, slot 1 is the permanent to destroy. Each is
// judged on its own at resolution (CR 608.2b), so one clause going
// illegal only turns its half off, while the other half and the exile
// still happen. The distinction the rules draw is between CHOOSING no
// targets and having every chosen target become illegal:
//
//   - zero targets chosen (both clauses declined) is a spell with no
//     targets, which resolves and exiles itself;
//   - at least one target chosen and every one of them illegal on
//     resolution is "countered by game rules" (CR 608.2b): none of it
//     happens and the card goes to its owner's graveyard, not exile.
//
// The engine's resolution frame makes that second call before the spell's
// effect runs, so this body never sees it.
//
// The reanimation is the batch entry door (ReturnFromGraveyardTogether),
// which can stop to ask a CR 614 question (a replacement choice on entry).
// The destroy and the exile are its continuation, so they wait for the
// answer instead of running on the line after it. The destroyed
// permanent was chosen at announce, so the card just reanimated cannot be
// it. "Exile Yawgmoth's Vile Offering" is part of the resolution, last:
// the spell moves itself and the frame leaves it there (a copy of the
// spell is not a card and has nothing to exile).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "740f9ab2-fc94-4abc-82b4-6b53116bf1a4",
		Name:         "Yawgmoth's Vile Offering",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCardInGraveyard("up to one target creature or planeswalker card from a graveyard",
				Or(Creature(), Planeswalker())).WithCount(0, 1),
			TargetPermanent("up to one target creature or planeswalker",
				Or(Creature(), Planeswalker())).WithCount(0, 1),
		),
		CastCondition:      LegendarySorcery(),
		CastConditionLabel: LegendarySorceryLabel,
		OnResolve:          yawgmothsVileOfferingResolve,
	})
}

func yawgmothsVileOfferingResolve(item *game.StackItem, ctx *Context) error {
	var back []uuid.UUID
	if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
		back = append(back, t.ID)
	}
	return ReturnFromGraveyardTogether{
		Targets:    back,
		Controller: ctx.Controller(),
		Then: func(ctx *Context, _ []uuid.UUID) error {
			if t, ok := ctx.ClauseTarget(1); ok && t.Kind == game.TargetCard {
				if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			if ctx.Item != nil && ctx.Item.IsCopy {
				return nil
			}
			return ctx.Game.ExileCardForEffect(item.SourceCardID)
		},
	}.Apply(ctx)
}
