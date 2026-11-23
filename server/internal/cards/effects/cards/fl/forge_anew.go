package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Forge Anew — Enchantment {2}{W} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"When this enchantment enters, return target Equipment card from your
//	 graveyard to the battlefield.
//	 During your turn, you may activate equip abilities any time you could
//	 cast an instant.
//	 You may pay {0} rather than pay the equip cost of the first equip
//	 ability you activate during each of your turns."
//
// Three abilities. The return is the ordinary targeted graveyard-to-
// battlefield, removed with no Equipment card to name (CR 603.3d). The timing
// is Leonin Shikari's equip-at-instant-speed grant with its "during your
// turn" flag (the activation timing seam). The alternative cost is Bruenor
// Battlehammer's {0} for the first equip activation each turn, gated on its
// controller's own turn. Only its controller gets either clause.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4df579df-fdbb-418a-af2d-709e140ee569",
		Name:         "Forge Anew",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Forge Anew — return target Equipment card from your graveyard to the battlefield",
					effects.SheoldredReanimate),
				effects.TargetCardInGraveyard("target Equipment card from your graveyard", effects.YouOwn(), effects.HasSubtype("Equipment"))),
		},
		ActivationTimings: []game.ActivationTiming{
			effects.EquipAbilitiesAtInstantSpeed("During your turn, you may activate equip abilities any time you could cast an instant.", true),
		},
		EquipCostAlternatives: []game.EquipCostAlternative{{
			Label:         "You may pay {0} rather than pay the equip cost of the first equip ability you activate during each of your turns.",
			Cost:          "{0}",
			FirstEachTurn: true,
			Condition: func(g *game.Game, controller uuid.UUID) bool {
				return g.Turn.ActiveSeat >= 0 && g.Turn.ActiveSeat < len(g.Seats) && g.Seats[g.Turn.ActiveSeat].ID == controller
			},
		}},
	})
}
