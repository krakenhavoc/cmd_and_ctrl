package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wake to Slaughter — Sorcery {3}{B}{R} (Innistrad: Midnight Hunt, #5497):
//
//	"Choose up to two target creature cards in your graveyard. An
//	 opponent chooses one of them. Return that card to your hand. Return
//	 the other to the battlefield under your control. It gains haste.
//	 Exile it at the beginning of the next end step.
//	 Flashback {4}{B}{R}"
//
// "An opponent" is untargeted: with two legal targets you name which
// opponent chooses (Elemental Teachings' ChoosePlayer), and that opponent
// picks the card that goes to your hand; the other returns to the
// battlefield under your control (the ruling: you do not choose). With one
// legal target left, that card is the one chosen and goes to your hand (CR
// 608.2b: an illegal target is not affected); with none the spell does
// nothing. The returned creature gains haste for as long as it stays and
// is exiled at the beginning of the next end step (CR 603.7), by a delayed
// trigger that exiles it only if it is still that object.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "78eb5399-8c8b-4310-b585-e4ccd9d7c9cc",
		Name:         "Wake to Slaughter",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCardInGraveyard("up to two target creature cards in your graveyard",
			effects.YouOwn(), effects.Creature()).WithCount(0, 2),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{4}{B}{R}")},
		OnResolve:        wakeToSlaughterResolve,
	})
}

func wakeToSlaughterResolve(_ *game.StackItem, ctx *effects.Context) error {
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	switch len(ids) {
	case 0:
		return nil
	case 1:
		return wakeToSlaughterSettle(ctx, ids, nil)
	}
	return effects.ChoosePlayer{
		Among:    effects.Opponents,
		Question: "Wake to Slaughter — choose an opponent to pick which creature card returns to your hand",
		ThenKey:  wakeToSlaughterOpponent,
		Carry:    ids,
	}.Apply(ctx)
}

// wakeToSlaughterOpponent is the answer: that opponent picks one card for
// the hand. Nobody to name leaves both in the graveyard. An on-disk key.
var wakeToSlaughterOpponent = effects.ChoosePlayerThen("option-pick/wake-to-slaughter-opponent", func(ctx *effects.Context, chosen uuid.UUID, found []uuid.UUID) error {
	if chosen == uuid.Nil {
		return nil
	}
	return effects.RevealPick{
		Player:   chosen,
		Owner:    ctx.Controller(),
		Question: "Wake to Slaughter — choose one of these creature cards; it returns to their hand and the other to the battlefield",
		Cards:    found,
		Min:      1,
		Max:      1,
		Keyed:    wakeToSlaughterPicked,
	}.Apply(ctx)
})

var wakeToSlaughterPicked = effects.RegisterRevealPick("choose-cards/wake-to-slaughter-settle", func(ctx *effects.Context, picked, left []uuid.UUID, _ []int) error {
	return wakeToSlaughterSettle(ctx, picked, left)
})

// wakeToSlaughterSettle returns the picked card to hand and the rest to the
// battlefield under the caster's control with haste and the end-step exile.
func wakeToSlaughterSettle(ctx *effects.Context, picked, left []uuid.UUID) error {
	g := ctx.Game
	for _, id := range picked {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	for _, id := range left {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if err := g.ReturnFromGraveyardWithCountersKeyedForEffect(id, ctx.Controller(), false, nil,
			ctx.Tail(effects.S82Mid10ReturnedHasteAndExile)); err != nil {
			return err
		}
	}
	return nil
}
