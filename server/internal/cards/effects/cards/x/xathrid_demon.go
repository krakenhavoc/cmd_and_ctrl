package cardsx

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Xathrid Demon — Creature — Demon {3}{B}{B}{B} 7/7 (Commander 2014):
//
//	"Flying, trample
//	 At the beginning of your upkeep, sacrifice a creature other than this
//	 creature, then each opponent loses life equal to the sacrificed
//	 creature's power. If you can't sacrifice a creature, tap this creature
//	 and you lose 7 life."
//
// Lord of the Pit's shape: the controller chooses the sacrifice; the run's
// continuation reads whether one was made. The drain is the sacrificed
// creature's last-known power (CR 608.2h); when none could be sacrificed the
// Demon taps (if it is still there) and its controller loses 7 life.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dd2bbc90-0474-42b7-afe4-3655a120ab02",
		Name:            "Xathrid Demon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Xathrid Demon — sacrifice a creature other than this creature, then each opponent loses life equal to its power; if you can't, tap this creature and you lose 7 life",
				func(g *game.Game, item *game.StackItem) error {
					source := item.SourceCardID
					return g.PlayerSacrificesThenKeyedForEffect(source, item.Controller,
						effects.SacrificeSpec("a creature other than this creature", effects.Creature(), effects.NotSelf(source)),
						"Xathrid Demon — sacrifice a creature other than this creature", 1,
						xathridDemonAfterSacrifice, []uuid.UUID{item.Controller})
				}),
		},
	})
}

var xathridDemonAfterSacrifice = game.RegisterPromptRunThen("prompt-run/xathrid-demon-sacrifice", func(g *game.Game, r game.RunDone) error {
	if len(r.Carry) == 0 {
		return nil
	}
	controller := r.Carry[0]
	ctx := effects.KeyedContext(g, controller, r.Source)
	if cards := r.Sacrifices().Cards(); len(cards) > 0 {
		power := effects.DepartedCreaturePower(g, cards[0])
		if power <= 0 {
			return nil
		}
		return effects.EachOpponentLosesLife{N: power}.Apply(ctx)
	}
	if err := (effects.TapTarget{Target: r.Source}).Apply(ctx); err != nil {
		return err
	}
	return g.ChangePlayerLifeForEffect(r.Source, controller, -7)
})
