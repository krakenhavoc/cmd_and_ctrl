package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vivien, Champion of the Wilds — Legendary Planeswalker — Vivien
// {2}{G}, loyalty 4:
//
//	"You may cast creature spells as though they had flash.
//	 +1: Until your next turn, up to one target creature gains reach and
//	 vigilance.
//	 −2: Look at the top three cards of your library. Exile one face down
//	 and put the rest on the bottom of your library in any order. For as
//	 long as it remains exiled, you may cast it if it's a creature
//	 spell."
//
// The static is Yeva, Nature's Herald's creature-only flash timing
// (CastKindAsThoughFlash). The +1 is Dovin Baan's shape: one data record
// pinned to the creature until Vivien's controller's next turn, adding
// reach and vigilance; "up to one" may choose nothing.
//
// The −2 is a look, not a reveal: only Vivien's controller sees the
// three. Exiling one is not optional, so the prompt's floor is one and a
// library showing a single card needs no prompt. The card lands face
// down with the grant on it (ExileCardsFaceDownWithPermissionForEffect,
// FaceDownPermitted), which is what lets its holder, and no one else,
// look at it for as long as it stays exiled (a card you can't see is a
// card you can't cast; the rulings let you look at it). The grant is a cast
// permission for as long as the card remains exiled, gated on the card
// being a creature (the permission's Filter, #3192 / ADR 0066's second
// 2026-10-13 amendment), so a land or a noncreature card is just exiled.
// It is a cast at the card's normal timing, which Vivien's own static
// turns into flash speed while she is on the battlefield. The rest go
// on the bottom in the order the player chooses (ADR 0088), after the
// exile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "146eb8c8-93d5-4dd5-af7a-46d5acd4a89d",
		Name:         "Vivien, Champion of the Wilds",
		Completeness: CompletenessFull,
		// The fallback for tokens, fixtures and the dev spawner; an
		// imported deck reads printed loyalty (ADR 0032 §1).
		StartingLoyalty: 4,
		CastTimings: []game.CastTimingRule{
			CastKindAsThoughFlash(game.PermissionFilter{CreatureOnly: true},
				"You may cast creature spells as though they had flash."),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Until your next turn, up to one target creature gains reach and vigilance.",
				Cost:    LoyaltyCost(1),
				Targets: TargetCreature("up to one target creature").WithCount(0, 1),
				Effect:  vivienPlusOne,
			},
			{
				Label:  "−2: Look at the top three cards of your library. Exile one face down and put the rest on the bottom of your library in any order. For as long as it remains exiled, you may cast it if it's a creature spell.",
				Cost:   LoyaltyCost(-2),
				Effect: vivienMinusTwo,
			},
		},
	})
}

// vivienPlusOne pins reach and vigilance to the chosen creature until
// Vivien's controller's next turn. Nothing chosen, or a target that left
// in response, does nothing.
func vivienPlusOne(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 || targets[0].Kind != game.TargetCard {
		return nil
	}
	return ScopedEffectFor{
		Target:   targets[0].ID,
		Mods:     []game.Mod{game.AddKeywordsMod("reach", "vigilance")},
		Duration: DurationUntilYourNextTurn(ctx, ctx.Controller()),
		Label:    "Vivien, Champion of the Wilds — reach and vigilance until your next turn",
	}.Apply(ctx)
}

// vivienMinusTwo is the look, the face-down exile with its creature-only
// grant, and the rest on the bottom in any order. Captures only scalars,
// so an undo replays it against the restored game.
func vivienMinusTwo(g *game.Game, item *game.StackItem) error {
	controller, source := item.Controller, item.SourceCardID
	looked := g.LookAtTopOfLibraryForEffect(controller, 3)
	if len(looked) == 0 {
		return nil
	}
	finish := func(g *game.Game, picked []uuid.UUID) error {
		var exileErr error
		if len(picked) > 0 {
			exileErr = g.ExileCardsFaceDownWithPermissionForEffect(picked[:1], controller, game.CastPermission{
				CastOnly:   true,
				Duration:   game.WhileInZoneDuration(),
				Filter:     game.PermissionFilter{CreatureOnly: true},
				Source:     source,
				SourceName: "Vivien, Champion of the Wilds",
				Label:      "For as long as it remains exiled, you may cast it if it's a creature spell",
			})
		}
		rest := make([]uuid.UUID, 0, len(looked))
		for _, id := range cardsStillInALibrary(g, looked) {
			if len(picked) == 0 || id != picked[0] {
				rest = append(rest, id)
			}
		}
		if err := g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
			Chooser:   controller,
			Source:    source,
			Cards:     rest,
			From:      game.ZoneLibrary,
			Placement: game.LibraryPlaceBottom,
			Reason:    "Vivien, Champion of the Wilds — put the rest on the bottom of your library in any order",
		}); err != nil && exileErr == nil {
			exileErr = err
		}
		return exileErr
	}
	if len(looked) == 1 {
		return finish(g, looked)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  controller,
		Source:   source,
		Question: "Vivien, Champion of the Wilds — exile one of them face down",
		Cards:    looked,
		Min:      1,
		Max:      1,
		Zone:     game.ZoneLibrary,
		Then:     finish,
	})
	return nil
}
