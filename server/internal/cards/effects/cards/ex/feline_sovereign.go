package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Feline Sovereign — Creature — Cat {2}{G}, 2/3 (Core Set 2021, #5511):
//
//	"Other Cats you control get +1/+1 and have protection from Dogs.
//	 Whenever one or more Cats you control deal combat damage to a player,
//	 destroy up to one target artifact or enchantment that player controls."
//
// The first sentence is the tribal lord over the post-layer view (the
// anthem in layer 7c, the protection token in layer 6; a Dog is a source of
// the quality by its subtype, a changeling Dog included). The trigger is
// once per combat damage step per player hit (CR 603.2c), however many Cats
// connect; its target is chosen as it goes on the stack from the artifacts
// and enchantments of the player who was dealt the damage, "up to one" so it
// triggers with no target too, and is re-checked on resolution.
//
// No simplification.
func init() {
	cats := effects.TribeFilter{Tribes: []string{"Cat"}, Others: true, YoursOnly: true}
	isCat := func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.HasSubtypeWord("Cat") }
	effects.Register(effects.Spec{
		OracleID:     "05f0feb6-6783-49fd-acf6-68b44aaff846",
		Name:         "Feline Sovereign",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(cats, 1, 1),
			effects.TribalKeywordGrant(cats, "protection from Dogs"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(isCat,
					"Feline Sovereign — destroy up to one target artifact or enchantment that player controls",
					effects.Do(effects.Instruction(func(ctx *effects.Context) error {
						return effects.S82MidDestroyTheTarget(nil, ctx)
					}))),
				effects.ControlledByThatPlayer(effects.TargetPermanent("up to one target artifact or enchantment that player controls",
					effects.Or(effects.Artifact(), effects.Enchantment())).WithCount(0, 1))),
		},
	})
}
