package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Careful Cultivation — Enchantment — Aura {2}{G} (Wilds of Eldraine
// Commander, #5368):
//
//	"Enchant artifact or creature
//	 As long as enchanted permanent is a creature, it gets +1/+3 and has
//	 reach and "{T}: Add {G}{G}."
//	 Channel — {1}{G}, Discard this card: Create a 1/1 green Human Monk
//	 creature token with "{T}: Add {G}.""
//
// The enchant clause is the spec the CR 704.5m check reruns, so the Aura
// stays on an artifact that stops being a creature. The three clauses
// share one condition, read on the host as the earlier layers leave it
// (cmdr2B19AttachedIsCreature): an animated artifact gets the pump, reach
// and the mana ability, and loses them when it stops being a creature.
// Channel works from the hand only (Boseiju's shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "83959c3d-053c-4d95-a1b0-fe92f2b906db",
		Name:         "Careful Cultivation",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("enchant artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{
			effects.TapForManaGrant(carefulCultivationMana, "{G}{G}", "Add {G}{G}", "{T}: Add {G}{G}."),
		},
		Static: []game.StaticAbility{
			effects.B43PumpAttachedWhen(effects.Cmdr2B19AttachedIsCreature, 1, 3),
			effects.B43GrantToAttachedWhen(effects.Cmdr2B19AttachedIsCreature, "reach"),
			effects.GrantAbilities(effects.Cmdr2B19AttachedIsCreature, carefulCultivationMana),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Channel — {1}{G}, Discard this card: Create a 1/1 green Human Monk creature token with \"{T}: Add {G}.\"",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1, CreatureTokens: 1},
			Cost:    game.AbilityCost{Mana: "{1}{G}", DiscardSelf: true},
			Zones:   []game.ZoneKind{game.ZoneHand},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Template: effects.Cmdr2B19HumanMonkToken(), N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

const carefulCultivationMana = "careful-cultivation/mana"
