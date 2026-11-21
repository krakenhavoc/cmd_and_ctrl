package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Butcher's Cleaver — Artifact — Equipment {3} (Innistrad Remastered, #5948):
//
//	"Equipped creature gets +3/+0.
//	 As long as equipped creature is a Human, it has lifelink.
//	 Equip {3}"
//
// The pump is the ordinary layer-7c attachment bonus. The lifelink is a layer-6
// grant gated on the host's creature types after layer 4 (a creature that
// becomes a Human gets it, one that stops being one loses it), the host's
// subtype read as a word of its own text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5cc5be7a-cbe6-4853-bb95-0390e25b4f2b",
		Name:         "Butcher's Cleaver",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 0),
			effects.Labelled(effects.GrantToAttachedWhile(func(host *game.Card, _ *game.Game, _ *game.Card) bool {
				return host.HasSubtypeWord("Human")
			}, "lifelink"), "As long as equipped creature is a Human, it has lifelink."),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
