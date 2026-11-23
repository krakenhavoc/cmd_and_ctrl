package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Canoptek Tomb Sentinel — Artifact Creature — Insect {4}, 4/3 (Warhammer
// 40,000 Commander, S82 GAP batch 187, #5948):
//
//	"Vigilance
//	 Exile Cannon — When this creature enters from a graveyard, exile up
//	 to one target nonland permanent.
//	 Unearth {7}"
//
// The trigger is Treacherous Pit-Dweller's: it fires only for an entry
// from a graveyard (unearth, reanimation), not from a hand cast.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "678b3fb5-aa9f-4268-9433-2a89dd37b927",
		Name:            "Canoptek Tomb Sentinel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated:       []effects.ActivatedAbility{effects.Unearth("{7}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventETB, effects.AllOf(effects.Self, effects.EnteredFromAGraveyard),
				"Canoptek Tomb Sentinel — exile up to one target nonland permanent",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					t, ok := effects.FirstLegalTarget(ctx)
					if !ok {
						return nil
					}
					return effects.ExileTarget{Target: t.ID}.Apply(ctx)
				}),
				effects.TargetPermanent("up to one target nonland permanent", effects.Nonland()).WithCount(0, 1)),
		},
	})
}
