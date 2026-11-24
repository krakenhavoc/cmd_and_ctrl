package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kogla, the Titan Ape — Legendary Creature — Ape {3}{G}{G}{G}, 7/6 (Secret
// Lair Drop):
//
//	"When Kogla enters, it fights up to one target creature you don't control.
//	 Whenever Kogla attacks, destroy target artifact or enchantment defending
//	 player controls.
//	 {1}{G}: Return target Human you control to its owner's hand. Kogla gains
//	 indestructible until end of turn."
//
// The fight is Mind Meanderer's ItFightsTheFirstTarget over "up to one" (none
// is a legal choice). The attack trigger reads the player Kogla attacked, as
// Grimgrin does. The third ability bounces the Human and then gives Kogla the
// indestructible, the second half only while the same permanent is there;
// Kogla can be its own Human target only if it is one (it is an Ape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8b21062e-97b0-4967-96df-30e8309f4fba",
		Name:         "Kogla, the Titan Ape",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: effects.Self,
				Key:       "Kogla, the Titan Ape — it fights up to one target creature you don't control",
				Targets:   effects.TargetCreature("up to one target creature you don't control", effects.OpponentControls()).WithCount(0, 1),
				Effect:    effects.ItFightsTheFirstTarget,
			},
			{
				Watches:     []game.EventKind{game.EventAttack},
				AppliesTo:   effects.ThisAttacked,
				TargetsFrom: effects.TargetArtifactOrEnchantmentDefendingPlayerControls,
				Key:         "Kogla, the Titan Ape — destroy target artifact or enchantment defending player controls",
				Effect:      effects.DestroyFirstLegalTarget,
			},
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}: Return target Human you control to its owner's hand. Kogla gains indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{1}{G}"),
			Targets: effects.TargetCreature("target Human you control", func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
				return c.HasSubtypeWord("Human")
			}, effects.YouControl()),
			Effect: effects.S82OldB023ReturnHumanThenIndestructible,
		}},
	})
}
