package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Apocalypse Runner — Artifact — Vehicle {2}{B}{R}, 6/5 (Aetherdrift,
// #4464):
//
//	"{T}: Target creature you control with power 2 or less gains
//	 lifelink until end of turn and can't be blocked this turn.
//	 Crew 3"
//
// The tap ability works whether or not the Vehicle is crewed: it taps
// the Vehicle itself, never the creature it targets. Its power bound is
// read as the ability is activated and again as it resolves (CR 608.2b),
// so a creature pumped above 2 in response gets nothing.
// The grant and the restriction are pinned to that creature until
// cleanup, as Ant-Man, Reformed Rogue's are.
//
// No simplification.
func init() {
	const label = "Apocalypse Runner — lifelink until end of turn and can't be blocked this turn"
	effects.Register(effects.Spec{
		OracleID:     "0ca1f76d-ae77-490a-8b32-d30ad33afc2d",
		Name:         "Apocalypse Runner",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Target creature you control with power 2 or less gains lifelink until end of turn and can't be blocked this turn.",
				Cost:    effects.TapCost(),
				Targets: effects.TargetCreature("target creature you control with power 2 or less", effects.YouControl(), effects.PowerLE(2)),
				Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0, "lifelink")),
				Effect: effects.Do(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
					if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"lifelink"}, Label: label}).Apply(ctx); err != nil {
						return err
					}
					return effects.RestrictUntilEOT{Target: id, Restrictions: game.CantBeBlocked, Label: label}.Apply(ctx)
				})),
			},
			{
				Label:   "Crew 3",
				Cost:    effects.CrewCost(3),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  effects.CrewEffect("Apocalypse Runner"),
			},
		},
	})
}
