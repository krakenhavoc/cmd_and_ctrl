package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dismiss into Dream — Enchantment {6}{U} (Magic 2014, S82 OLD batch 401, #5563):
//
//	"Each creature your opponents control is an Illusion in addition to its other
//	 types and has "When this creature becomes the target of a spell or ability,
//	 sacrifice it.""
//
// Two statics over the same set, read on every layer pass (CR 611.3a), so a
// creature an opponent gains later is covered and one that changes hands is not:
// a layer-4 type addition (CR 205.1b: "in addition", so its other types stay) and
// a layer-6 grant of Gossamer Phantasm's own trigger (Phantasmal Image's
// sacrifice, put on the stack when the creature becomes a target, CR 603.2).
//
// No simplification.
func init() {
	opponentsCreatures := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.IsCreature() && target.Controller != source.Controller
	}
	effects.Register(effects.Spec{
		OracleID:     "b76566a9-f998-4a20-b073-4b60862843c9",
		Name:         "Dismiss into Dream",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: dismissIntoDreamSacrifice,
			Triggered: []game.TriggeredAbility{
				effects.On(game.EventBecomesTarget, effects.Self, "Dismiss into Dream — sacrifice it", effects.SacrificeThisIfStillOnBattlefield),
			},
			Text: "When this creature becomes the target of a spell or ability, sacrifice it.",
		}},
		Static: []game.StaticAbility{
			{
				Label:     "Each creature your opponents control is an Illusion in addition to its other types.",
				Layer:     game.Layer4Type,
				AppliesTo: opponentsCreatures,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
					illusion := source.TextSubtype("Illusion")
					for _, s := range c.Subtypes {
						if s == illusion {
							return
						}
					}
					c.Subtypes = append(c.Subtypes, illusion)
				},
			},
			effects.GrantAbilities(opponentsCreatures, dismissIntoDreamSacrifice),
		},
	})
}

const dismissIntoDreamSacrifice = "dismiss-into-dream/sacrifice-when-targeted"
