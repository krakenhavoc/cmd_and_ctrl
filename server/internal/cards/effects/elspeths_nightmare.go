package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ff768016-67f8-409e-8359-9ed05bcb46d2",
		Name:         "Elspeth's Nightmare",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, SagaChapterLabel("Elspeth's Nightmare", 1, "destroy a creature with power 2 or less"),
				TargetCreature("target creature an opponent controls with power 2 or less", OpponentControls(), PowerLE(2)),
				destroyChosenPermanent),
			ChapterTriggerTargeting(2, SagaChapterLabel("Elspeth's Nightmare", 2, "target opponent reveals their hand"),
				TargetPlayer("target opponent", Opponent()),
				TargetRevealsYouChooseDiscardAbility(And(Noncreature(), Nonland()), "noncreature, nonland card")),
			ChapterTriggerTargeting(3, SagaChapterLabel("Elspeth's Nightmare", 3, "exile target opponent's graveyard"),
				TargetPlayer("target opponent", Opponent()),
				elspethsNightmareExileGraveyard),
		},
	})
}

// elspethsNightmareExileGraveyard is chapter III: the target opponent's
// graveyard, if the target is still legal.
func elspethsNightmareExileGraveyard(g *game.Game, item *game.StackItem) error {
	return b30ExileTargetGraveyards(NewContext(g, item))
}
