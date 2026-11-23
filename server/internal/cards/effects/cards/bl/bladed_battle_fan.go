package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bladed Battle-Fan — Artifact — Equipment {1}{B} (March of the Machine):
//
//	"Flash
//	 When this Equipment enters, attach it to target creature you control.
//	 That creature gains indestructible until end of turn.
//	 Equipped creature gets +1/+0.
//	 Equip {1}"
//
// One enters trigger with one target: it attaches the Equipment and gives that
// creature indestructible until end of turn. A target that has gone by
// resolution gets neither (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c4614b6c-5e77-465f-996f-504121ea8eff",
		Name:            "Bladed Battle-Fan",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static:          []game.StaticAbility{effects.PumpAttached(1, 0)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Bladed Battle-Fan — attach it to target creature you control; that creature gains indestructible until end of turn", bladedBattleFanEnters),
				effects.PermanentYouControl("target creature you control", effects.Creature()),
			),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}

var bladedBattleFanIndestructible = effects.Do(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
	return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"indestructible"},
		Label: "Bladed Battle-Fan — indestructible until end of turn"}.Apply(ctx)
}))

func bladedBattleFanEnters(g *game.Game, item *game.StackItem) error {
	if err := effects.AttachSourceToTarget(g, item); err != nil {
		return err
	}
	return bladedBattleFanIndestructible(g, item)
}
