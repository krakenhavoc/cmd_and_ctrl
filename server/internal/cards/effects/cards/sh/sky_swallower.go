package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sky Swallower — Creature — Leviathan {3}{U}{U}, 8/8 (Guildpact):
//
//	"Flying
//	 When this creature enters, target opponent gains control of all other
//	 permanents you control."
//
// One control-changing effect per permanent, each with no duration (CR 611.2a,
// Admiral Beckett Brass's GainControl): the set is "all other permanents you
// control" as the ability resolves (CR 611.2c), so permanents that arrive
// later stay yours, and the Swallower itself is left out. The new controller
// is the targeted opponent; a target that is illegal on resolution means the
// ability does nothing (CR 608.2b). Creatures changing hands are summoning
// sick under their new controller (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b345c413-df5a-4dae-8b92-5fab54a097fc",
		Name:            "Sky Swallower",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Sky Swallower — target opponent gains control of all other permanents you control", skySwallowerGive),
				effects.TargetPlayer("target opponent", effects.Opponent()),
			),
		},
	})
}

func skySwallowerGive(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	opponent, ok := effects.FirstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range effects.MatchingBattlefield(ctx, effects.YouControl()) {
		if c.InstanceID != item.SourceCardID {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := (effects.GainControl{
			Target:     id,
			Controller: opponent,
			Duration:   game.IndefiniteDuration(),
			Label:      "Sky Swallower — control",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
