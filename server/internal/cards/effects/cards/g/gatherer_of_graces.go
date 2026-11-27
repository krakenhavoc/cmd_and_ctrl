package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gatherer of Graces — Creature — Human Druid {1}{G}, 1/2 (Guildpact):
//
//	"This creature gets +1/+1 for each Aura attached to it.
//	 Sacrifice an Aura: Regenerate this creature."
//
// Kor Spiritdancer's self-only layer 7c count of the Auras whose attachment
// points at the creature, re-read on every layer pass. The cost is any Aura
// its controller controls, attached to the Gatherer or not (CR 701.21); the
// shield is the shared "regenerate this creature" body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b19054c-ba33-44b4-b328-bb8ea3f52dab",
		Name:         "Gatherer of Graces",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Gets +1/+1 for each Aura attached to it.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, target *game.Card, g *game.Game, _ *game.Card) {
				n := effects.OldB572AurasAttachedTo(g, target.InstanceID)
				c.Power += n
				c.Toughness += n
			},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice an Aura: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerSacOutlet},
			Cost:    game.AbilityCost{SacrificeOther: effects.SacrificeSpec("an Aura", effects.Subtype("Aura"))},
			Effect:  effects.RegenerateThisSource,
		}},
	})
}
