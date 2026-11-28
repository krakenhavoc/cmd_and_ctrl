package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nature's Chosen — Enchantment — Aura {G} (Alliances, #5634):
//
//	"Enchant creature you control
//	 {0}: Untap enchanted creature. Activate only during your turn and only once
//	 each turn.
//	 Tap enchanted creature: Untap target artifact, creature, or land. Activate
//	 only if enchanted creature is white and untapped and only once each turn."
//
// Two Aura abilities over the creature it is attached to. The {0} untap is
// limited to the controller's own turn and once each turn (the announce tally,
// CR 602.5). The second pays "Tap enchanted creature" as a cost (a tap-others
// cost on the host, not {T}), needs the host to be white as the Aura's text
// reads it and untapped, and is also once each turn, counted apart from the
// first; its target may be the very creature just tapped.
//
// No simplification.
func init() {
	const (
		name   = "Nature's Chosen"
		label0 = "{0}: Untap enchanted creature. Activate only during your turn and only once each turn."
		label1 = "Tap enchanted creature: Untap target artifact, creature, or land. Activate only if enchanted creature is white and untapped and only once each turn."
	)
	effects.Register(effects.Spec{
		OracleID:     "7f425406-2746-45a1-86bf-f9645015ffc2",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.YouControl()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Activated: []effects.ActivatedAbility{
			{
				Label:     label0,
				Purpose:   game.Purpose{Answers: game.AnswerSurvive},
				Cost:      effects.ManaCost("{0}"),
				Condition: effects.AllConditions(effects.DuringYourTurn(), effects.OncePerTurnActivation(label0)),
				Effect:    effects.OldB816UntapEnchantedCreature,
			},
			{
				Label:   label1,
				Cost:    effects.OldB816TapEnchantedCreature(),
				Targets: effects.TargetPermanent("target artifact, creature, or land", effects.Or(effects.Artifact(), effects.Creature(), effects.Land())),
				Condition: effects.AllConditions(
					func(g *game.Game, _, source uuid.UUID) bool {
						return effects.OldB816EnchantedCreatureIs(g, source, "W")
					},
					func(g *game.Game, _, source uuid.UUID) bool {
						return effects.OldB816EnchantedCreatureIsUntapped(g, source)
					},
					effects.OncePerTurnActivation(label1)),
				Effect: effects.UntapFirstLegalTarget,
			},
		},
	})
}
