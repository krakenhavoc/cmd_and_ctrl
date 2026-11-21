package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Death of Gwen Stacy — Enchantment — Saga {2}{B} (Marvel's Spider-Man,
// #4460):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Destroy target creature.
//	 II — Each player may discard a card. Each player who doesn't loses 3 life.
//	 III — Exile any number of target players' graveyards."
//
// Chapter II is one discard run over every seat in turn order starting with
// the active player (each seat is asked "up to one card"); once every answer
// is in, each player who did not discard (a declined prompt, or an empty
// hand) loses 3 life, in that same order. The losses wait for the last answer
// so no player decides with another's loss already paid.
//
// Chapter III targets any number of players (zero included), and each chosen
// player's graveyard is exiled as the chapter resolves; a player who is no
// longer a legal target is left alone (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2624b5c5-ba59-4e20-8e66-e463268366e4",
		Name:         "The Death of Gwen Stacy",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTriggerTargeting(1, effects.SagaChapterLabel("The Death of Gwen Stacy", 1, "destroy target creature"),
				effects.TargetCreature("target creature"),
				effects.DestroyChosenPermanent),
			effects.ChapterTrigger(2, effects.SagaChapterLabel("The Death of Gwen Stacy", 2, "each player may discard a card; each who doesn't loses 3 life"),
				func(g *game.Game, item *game.StackItem) error {
					return g.EachPlayerDiscardsThenKeyedForEffect(uuid.Nil,
						game.DiscardPrompt{
							Source:   item.SourceCardID,
							N:        1,
							UpTo:     true,
							Question: "The Death of Gwen Stacy — discard a card, or lose 3 life?",
						}, gwenStacyNonDiscardersLose, nil)
				}),
			effects.ChapterTriggerTargeting(3, effects.SagaChapterLabel("The Death of Gwen Stacy", 3, "exile any number of target players' graveyards"),
				effects.TargetPlayer("any number of target players").WithCount(0, 0),
				func(g *game.Game, item *game.StackItem) error {
					return effects.B30ExileTargetGraveyards(effects.NewContext(g, item))
				}),
		},
	})
}

// gwenStacyNonDiscardersLose is chapter II's second sentence: every seat in
// turn order that did not discard loses 3 life. An on-disk key: never
// renamed, never reused.
var gwenStacyNonDiscardersLose = game.RegisterPromptRunThen("prompt-run/the-death-of-gwen-stacy-lose-life", func(g *game.Game, r game.RunDone) error {
	discards := r.Discards()
	n := len(g.Seats)
	if n == 0 {
		return nil
	}
	for i := 0; i < n; i++ {
		p := g.Seats[(g.Turn.ActiveSeat+i)%n]
		if p == nil || p.Eliminated || len(discards.By(p.ID)) > 0 {
			continue
		}
		if err := g.ChangePlayerLifeForEffect(r.Source, p.ID, -3); err != nil {
			return err
		}
	}
	return nil
})
