package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mjölnir, Hammer of Thor — {3}{R} Legendary Artifact — Equipment (Marvel
// Super Heroes, #4455):
//
//	"When Mjölnir enters, it deals 4 damage to up to one target creature.
//	 Double all damage equipped creature would deal.
//	 Equip worthy {1} (A creature is worthy if it's a legendary non-Villain
//	 that's red and/or white.)
//	 {2}{R}, Discard this card: It deals 2 damage to each creature."
//
// The enter trigger targets up to one creature (none is a legal choice) and
// Mjölnir is the damage source. The doubling is a replacement effect
// (CR 614.1b, the Inquisitor's Flail shape) that applies to every damage
// event, combat or not, whose source is the creature the Equipment is
// attached to. Equip worthy is a normal equip ability (CR 702.6a, sorcery
// speed) whose target clause is the worthiness test, judged only while
// activating. The last ability is a hand-zone channel-style ability: the card
// is discarded as a cost and, from the graveyard, still deals the damage to
// each creature as one instance.
//
// No simplification.
func init() {
	worthy := effects.EquipOnlyAbility("Equip worthy {1}", "{1}",
		effects.TargetCreature("target worthy creature you control", effects.YouControl(), effects.Legendary(),
			effects.Not(effects.Subtype("Villain")), effects.Or(effects.OfColor("R"), effects.OfColor("W"))))
	effects.Register(effects.Spec{
		OracleID:     "7f9a8845-d760-44a7-a4c9-8a20dba4e14a",
		Name:         "Mjölnir, Hammer of Thor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Mjölnir — 4 damage to up to one target creature", effects.SourceDealsDamageToEachLegalTarget(4)),
				effects.TargetCreature("up to one target creature").WithCount(0, 1)),
		},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventDamage && ev.DamageAmount > 0 &&
					ev.DamageSource != uuid.Nil && src.IsAttachedTo(ev.DamageSource)
			},
			Replace:    effects.B39DoubleTheDamage,
			Controller: effects.B39SourceController,
			Label:      "Mjölnir: double all damage the equipped creature would deal",
		}},
		Activated: []effects.ActivatedAbility{
			worthy,
			{
				Label:   "{2}{R}, Discard this card: It deals 2 damage to each creature.",
				Cost:    game.AbilityCost{Mana: "{2}{R}", DiscardSelf: true},
				Zones:   []game.ZoneKind{game.ZoneHand},
				Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2, Partial: true}},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DamageEachMatching(effects.NewContext(g, item), effects.Creature(), 2)
				},
			},
		},
	})
}
