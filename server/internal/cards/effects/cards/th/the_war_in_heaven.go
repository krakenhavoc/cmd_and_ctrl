package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The War in Heaven — Enchantment — Saga {3}{B}{B}{B} (Warhammer 40,000
// Commander):
//
//	I   — You draw three cards and you lose 3 life.
//	II  — Mill three cards.
//	III — Choose up to three target creature cards with total mana value 8
//	      or less in your graveyard. Return each of them to the battlefield
//	      with a necrodermis counter on it. They're artifacts in addition
//	      to their other types.
//
// Chapter I is Ancient Craving's draw then loss. Chapter III's targets
// carry a ceiling on their summed mana value (the Super Hero Civil War
// shape), judged at announce and again as the chapter resolves: when the
// survivors no longer fit, none of them is legal (CR 608.2b). They return
// together as one entry, each with a necrodermis counter as part of the
// entry (CR 122.6) and as an artifact in addition to its other types from
// the moment it lands (ADR 0210's EntersAs), so it is a creature and an
// artifact when its enters triggers look.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5924cbb0-19b5-449c-a7c5-722e744a54f8",
		Name:         "The War in Heaven",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 3},
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel("The War in Heaven", 1, "you draw three cards and you lose 3 life"),
				effects.Do(effects.DrawCards{N: 3}, effects.GainLife{Amount: -3})),
			effects.ChapterTrigger(2, effects.SagaChapterLabel("The War in Heaven", 2, "mill three cards"),
				effects.Do(effects.MillCards{N: 3})),
			effects.ChapterTriggerTargeting(3, effects.SagaChapterLabel("The War in Heaven", 3, "return up to three target creature cards with total mana value 8 or less from your graveyard with a necrodermis counter on each; they're artifacts"),
				effects.TargetCardInGraveyard("up to three target creature cards with total mana value 8 or less in your graveyard",
					effects.Creature(), effects.YouOwn()).WithCount(0, 3).WithTotal(effects.TotalManaValueAtMost(8)),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					var ids []uuid.UUID
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							ids = append(ids, t.ID)
						}
					}
					if len(ids) == 0 {
						return nil
					}
					return effects.PutCardsFromZoneTogetherWith(ctx, game.ZoneGraveyard, ids, game.ZoneEntryOptions{
						Controller:         item.Controller,
						EntersWithCounters: map[string]int{"necrodermis": 1},
						EntersAs: &game.EntersAs{
							AddTypes: []string{"Artifact"},
							Label:    "The War in Heaven — an artifact in addition to its other types",
						},
					}, game.TailRef{}, 0)
				}),
		},
	})
}
