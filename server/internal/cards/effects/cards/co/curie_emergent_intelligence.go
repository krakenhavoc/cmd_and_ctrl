package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Curie, Emergent Intelligence — Legendary Artifact Creature — Robot {1}{U},
// 1/3 (Fallout):
//
//	"Whenever Curie deals combat damage to a player, draw cards equal to its
//	 base power.
//	 {1}{U}, Exile another nontoken artifact creature you control: Curie
//	 becomes a copy of the exiled creature, except it has "Whenever this
//	 creature deals combat damage to a player, draw cards equal to its base
//	 power.""
//
// The cost is Mechtitan Core's exile clause narrowed to nontoken artifact
// creatures (ExilePermanentsCost.ArtifactCreature and Nontoken), chosen from
// the board, so the exiled creature is on the cost's record. The copy is
// Dimir Doppelganger's duration-free copy with an "except" grant: the exiled
// creature's copiable values are read from its last-known record (CR 608.2h,
// CR 707.2: a creature that was itself a copy is copied as that copy), and
// the granted trigger is the printed one, so a Curie that became a copy
// still draws. "Its base power" is the layered base power of the creature
// that dealt the damage (CR 208.4b): Curie's printed 1, or the copied
// creature's after a copy, not a counter or a pump.
//
// No simplification.
const curieGrantKey = "curie-emergent-intelligence/damage-draw"

func curieDamageTrigger() game.TriggeredAbility {
	return effects.WheneverThisDealsCombatDamageToAPlayer("Curie, Emergent Intelligence — draw cards equal to its base power",
		func(g *game.Game, item *game.StackItem) error {
			c, ok := g.LookupCardForEffect(item.SourceCardID)
			if !ok {
				return nil
			}
			n := c.BasePower()
			if info, ok := g.LastKnownPermanentForEffect(item.SourceCardID); ok && info.Left {
				n = info.BasePower
			}
			if n <= 0 {
				return nil
			}
			return effects.DrawCards{N: n}.Apply(effects.NewContext(g, item))
		})
}

func init() {
	effects.Register(effects.Spec{
		OracleID:     "48cf6b88-6bca-49d4-9951-36d98befa588",
		Name:         "Curie, Emergent Intelligence",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       curieGrantKey,
			Text:      "Whenever this creature deals combat damage to a player, draw cards equal to its base power.",
			Triggered: []game.TriggeredAbility{curieDamageTrigger()},
		}},
		Triggered: []game.TriggeredAbility{curieDamageTrigger()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}, Exile another nontoken artifact creature you control: Curie becomes a copy of the exiled creature, except it has \"Whenever this creature deals combat damage to a player, draw cards equal to its base power.\"",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerCombatGrant | game.AnswerSacOutlet | game.AnswerProtect | game.AnswerSurvive},
			Cost: effects.Plus(effects.ManaCost("{1}{U}"), game.AbilityCost{ExilePermanents: &game.ExilePermanentsCost{
				Count:            1,
				ExcludeSource:    true,
				ArtifactCreature: true,
				Nontoken:         true,
				Label:            "another nontoken artifact creature you control",
			}}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				exiled := ctx.Paid().Exiled
				if len(exiled) == 0 {
					return nil
				}
				var values game.PrintedValues
				if info, ok := g.LastKnownPermanentForEffect(exiled[0]); ok && info.Copiable != nil {
					values = info.Copiable.Clone()
				} else if c, ok := g.LookupCardForEffect(exiled[0]); ok {
					values = game.OwnPrintedValues(c)
				} else {
					return nil
				}
				return effects.BecomeCopy{
					Targets:  []uuid.UUID{item.SourceCardID},
					Of:       exiled[0],
					Values:   &values,
					Duration: effects.CopyIndefinite,
					Except:   func(v *game.PrintedValues) { v.GrantAbility(curieGrantKey) },
					Label:    "Curie, Emergent Intelligence — becomes a copy of the exiled creature",
				}.Apply(ctx)
			},
		}},
	})
}
