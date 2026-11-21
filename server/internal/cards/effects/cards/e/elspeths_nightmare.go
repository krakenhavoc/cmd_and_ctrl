package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elspeth's Nightmare — Enchantment — Saga {2}{B}:
//
//	"I — Destroy target creature an opponent controls with power 2 or
//	     less.
//	 II — Target opponent reveals their hand. You choose a noncreature,
//	      nonland card from it. That player discards that card.
//	 III — Exile target opponent's graveyard."
//
// Three chapter abilities (CR 714.2b). Chapter I reads the creature's
// power as the chapter goes on the stack and again as it resolves
// (CR 608.2b), so a creature pumped above 2 in response is no longer a
// legal target. Chapter II is Duress's pick (ADR 0116), chosen by the
// Saga's controller (CR 113.8): a hand with no noncreature, nonland card
// is revealed and nothing is discarded (CR 609.3). Chapter III exiles
// every card in the target's graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ff768016-67f8-409e-8359-9ed05bcb46d2",
		Name:         "Elspeth's Nightmare",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTriggerTargeting(1, effects.SagaChapterLabel("Elspeth's Nightmare", 1, "destroy a creature with power 2 or less"),
				effects.TargetCreature("target creature an opponent controls with power 2 or less", effects.OpponentControls(), effects.PowerLE(2)),
				effects.DestroyChosenPermanent),
			effects.ChapterTriggerTargeting(2, effects.SagaChapterLabel("Elspeth's Nightmare", 2, "target opponent reveals their hand"),
				effects.TargetPlayer("target opponent", effects.Opponent()),
				effects.TargetRevealsYouChooseDiscardAbility(effects.And(effects.Noncreature(), effects.Nonland()), "noncreature, nonland card")),
			effects.ChapterTriggerTargeting(3, effects.SagaChapterLabel("Elspeth's Nightmare", 3, "exile target opponent's graveyard"),
				effects.TargetPlayer("target opponent", effects.Opponent()),
				elspethsNightmareExileGraveyard),
		},
	})
}

// elspethsNightmareExileGraveyard is chapter III: the target opponent's
// graveyard, if the target is still legal.
func elspethsNightmareExileGraveyard(g *game.Game, item *game.StackItem) error {
	return effects.B30ExileTargetGraveyards(effects.NewContext(g, item))
}
