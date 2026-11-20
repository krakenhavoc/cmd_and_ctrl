package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bearer of Memory — Enchantment Creature — Human Monk {2}{G}, 3/2
// (Kamigawa: Neon Dynasty, S82):
//
//	"{5}{G}: Put a +1/+1 counter on target enchantment creature. It gains
//	 trample until end of turn."
//
// The target is any enchantment creature, this one included. The counter and
// the trample both go to the same target; a target that is no longer legal
// gets neither (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "51e76d8f-95c6-428f-9030-3f8718139f93",
		Name:         "Bearer of Memory",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{G}: Put a +1/+1 counter on target enchantment creature. It gains trample until end of turn.",
			Cost:    effects.ManaCost("{5}{G}"),
			Targets: effects.TargetCreature("target enchantment creature", effects.Enchantment()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id := effects.FirstLegalCardTarget(ctx)
				if id == uuid.Nil {
					return nil
				}
				if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Bearer of Memory — trample"}.Apply(ctx)
			},
		}},
	})
}
