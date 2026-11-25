package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Force of Virtue — Enchantment {2}{W}{W} (Modern Horizons):
//
//	"If it's not your turn, you may exile a white card from your hand rather
//	 than pay this spell's mana cost.
//	 Flash
//	 Creatures you control get +1/+1."
//
// Force of Vigor's pitch (an alternative cost gated on "not your turn", the
// exiled white card named as the cost is paid), on an enchantment spell with
// flash (a printed keyword), so it can be cast on another player's turn. The
// anthem is Gaea's: a layer 7c static over the creatures its controller
// controls, the Force not being a creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4262893c-8cbe-4854-9762-b3b6b75c0e35",
		Name:            "Force of Virtue",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		AlternativeCosts: []game.AlternativeCost{
			{
				Key:           "pitch",
				Label:         "If it's not your turn, exile a white card from your hand",
				ExileFromHand: effects.CardInYourHand("a white card from your hand", effects.OfColor("W")),
				PayLabel:      "a white card from your hand",
				Condition: func(g *game.Game, controller uuid.UUID) bool {
					return !effects.IsActivePlayer(g, controller)
				},
			},
		},
		Static: []game.StaticAbility{effects.Labelled(effects.B16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsCreature() && target.Controller == source.Controller
		}, 1, 1), "Creatures you control get +1/+1.")},
	})
}
