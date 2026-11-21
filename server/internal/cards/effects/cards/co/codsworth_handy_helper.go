package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Codsworth, Handy Helper — Legendary Artifact Creature — Robot {2}{W}, 2/3
// (Fallout):
//
//	"Commanders you control have ward {2}.
//	 {T}: Add {W}{W}. Spend this mana only to cast Aura and/or Equipment
//	 spells.
//	 {T}: Attach target Aura or Equipment you control to target creature you
//	 control. Activate only as a sorcery."
//
// The ward grant is Guardian Augmenter's commander test with Bronze
// Guardian's WardGranted. The mana ability carries a cast-only restriction
// to the Aura and Equipment subtypes (a subtype tag admits any of its "|"
// alternatives). The attach is Brass Squire's two-slot clause pair, with an
// Aura or Equipment as the first clause: AttachForEffect refuses an Aura
// that can't enchant the creature and leaves it where it is (CR 701.3b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f496ff42-3644-494a-9f21-27f781ee880b",
		Name:         "Codsworth, Handy Helper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WardGranted(effects.WardMana("{2}"), "Codsworth, Handy Helper — commanders you control have ward {2}",
				func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.IsCommander && target.Controller == source.Controller
				}),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:         effects.ManaAbilityCost{Tap: true},
			Produced:     "{W}{W}",
			Label:        "Add {W}{W}. Spend this mana only to cast Aura and/or Equipment spells",
			Restrictions: []string{effects.ManaRestrictCast, effects.ManaRestrictSubtype("Aura|Equipment")},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:        "{T}: Attach target Aura or Equipment you control to target creature you control. Activate only as a sorcery.",
			Cost:         effects.TapCost(),
			SorcerySpeed: true,
			Targets: effects.Clauses(
				effects.TargetPermanent("target Aura or Equipment you control", effects.Or(effects.HasSubtype("Aura"), effects.HasSubtype("Equipment")), effects.YouControl()),
				effects.TargetCreature("target creature you control", effects.YouControl()),
			),
			Effect: effects.AttachClauseTargets,
		}},
	})
}
