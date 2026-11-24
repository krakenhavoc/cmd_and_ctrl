package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lorehold Apprentice — Creature — Human Cleric {R}{W}, 2/2 (Strixhaven:
// School of Mages, #5502):
//
//	"Magecraft — Whenever you cast or copy an instant or sorcery spell, until
//	 end of turn, Spirit creatures you control gain "{T}: This creature deals
//	 1 damage to each opponent.""
//
// A duration grant (ADR 0093, Banishing Knack's shape) over the Spirit
// creatures you control as the trigger resolves (CR 611.2c): one that enters
// later does not gain it, and a creature that leaves and returns is a new
// object without it. The granted ability is the creature's own, so the
// usual summoning-sickness rule for {T} applies, and the damage is dealt by
// that creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3858a2ff-cdc9-4e40-bc2c-42ed691aae08",
		Name:         "Lorehold Apprentice",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: loreholdApprenticePing,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: This creature deals 1 damage to each opponent.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.TapCost(),
				Effect:  effects.Do(effects.DamageEachOpponent{N: 1}),
			}},
			Text: "{T}: This creature deals 1 damage to each opponent.",
		}},
		Triggered: []game.TriggeredAbility{
			effects.Magecraft("Lorehold Apprentice — until end of turn, Spirit creatures you control gain \"{T}: This creature deals 1 damage to each opponent.\" (magecraft)",
				func(g *game.Game, item *game.StackItem) error {
					return effects.GrantAbilitiesFor{
						Match: effects.And(effects.Creature(), effects.YouControl(), effects.OfSubtype("Spirit")),
						Keys:  []string{loreholdApprenticePing},
						Label: "Lorehold Apprentice — Spirits gain \"{T}: 1 damage to each opponent\"",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// loreholdApprenticePing is the granted bundle's key. An on-disk identity:
// never renamed, never reused.
const loreholdApprenticePing = "lorehold-apprentice/ping"
