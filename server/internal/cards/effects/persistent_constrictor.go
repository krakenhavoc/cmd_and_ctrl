package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Persistent Constrictor — Creature — Zombie Snake {4}{B}, 5/3:
//
//	"At the beginning of each opponent's upkeep, they lose 1 life and
//	 you put a -1/-1 counter on up to one target creature they control.
//	 Persist (When this creature dies, if it had no -1/-1 counters on
//	 it, return it to the battlefield under its owner's control with a
//	 -1/-1 counter on it.)"
//
// The upkeep's player is the trigger event's actor, and the target
// clause is built from it (TargetsFrom), so only a creature THAT player
// controls can be chosen. "Up to one": with none chosen the ability
// still resolves and the life is lost. With one chosen that has become
// an illegal target, the ability has no legal targets and does not
// resolve at all, so no life is lost either (CR 608.2b, the 2024-09-20
// ruling).
//
// Persist is PrintedKeywords; the engine derives the trigger
// (game/undying_persist.go, #2075).
//
// No simplification.
func init() {
	t := AtEachOpponentsUpkeep("Persistent Constrictor — they lose 1 life, -1/-1 counter on up to one creature they control",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			player := ctx.Trigger().Event.Actor
			if player == uuid.Nil {
				return nil
			}
			if err := (GainLife{Player: player, Amount: -1}).Apply(ctx); err != nil {
				return err
			}
			for _, t := range ctx.LegalTargets() {
				return AddCounter{Target: t.ID, Kind: game.CounterMinusOne, N: 1}.Apply(ctx)
			}
			return nil
		})
	t.TargetsFrom = func(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
		spec := TargetCreature("up to one target creature they control", ControlledBy(tc.Event.Actor))
		spec.Min = 0
		return spec
	}
	Register(Spec{
		OracleID:        "bf0ee9da-6071-440c-9149-d441f730fcfb",
		Name:            "Persistent Constrictor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordPersist},
		Triggered:       []game.TriggeredAbility{t},
	})
}
