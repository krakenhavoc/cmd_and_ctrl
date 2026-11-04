package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Force of Vigor — Instant {2}{G}{G} (Marvel Universe, #4460):
//
//	"If it's not your turn, you may exile a green card from your hand
//	 rather than pay this spell's mana cost.
//	 Destroy up to two target artifacts and/or enchantments."
//
// Force of Negation's pitch (an alternative cost gated on "not your
// turn", the exiled card named as the cost is paid) over Wear Down's
// destroy-each-legal-target loop with an "up to two" clause. With no
// target chosen the spell does nothing; a target that stops being legal
// is skipped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5f627e45-deee-4a4f-8128-57b0d8ebc553",
		Name:         "Force of Vigor",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("up to two target artifacts and/or enchantments", Or(Artifact(), Enchantment())).WithCount(0, 2),
		AlternativeCosts: []game.AlternativeCost{
			{
				Key:           "pitch",
				Label:         "If it's not your turn, exile a green card from your hand",
				ExileFromHand: CardInYourHand("a green card from your hand", OfColor("G")),
				PayLabel:      "a green card from your hand",
				Condition: func(g *game.Game, controller uuid.UUID) bool {
					return !isActivePlayer(g, controller)
				},
			},
		},
		OnResolve: destroyEachLegalTarget,
	})
}
