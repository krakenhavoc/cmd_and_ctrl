package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leering Emblem — Artifact — Equipment {2}:
//
//	"Whenever you cast a spell, equipped creature gets +2/+2 until end of turn.
//	 Equip {2}"
//
// Sigil of Valor's shape: the trigger is the Equipment's, watching your casts
// (any spell); as it resolves the creature then equipped gets the pump, and
// with nothing equipped it does nothing (CR 608.2b). Equip {2} is the shared
// equip ability.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a5cfd1e4-cd14-454a-9734-b5f6952d6a50",
		Name:         "Leering Emblem",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(nil, "Leering Emblem — equipped creature gets +2/+2 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					id, ok := effects.B43AttachedCreature(g, item.SourceCardID)
					if !ok {
						return nil
					}
					return effects.BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Leering Emblem — +2/+2 until end of turn"}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
