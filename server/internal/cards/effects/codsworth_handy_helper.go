package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "f496ff42-3644-494a-9f21-27f781ee880b",
		Name:         "Codsworth, Handy Helper",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WardGranted(WardMana("{2}"), "Codsworth, Handy Helper — commanders you control have ward {2}",
				func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.IsCommander && target.Controller == source.Controller
				}),
		},
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			Produced:     "{W}{W}",
			Label:        "Add {W}{W}. Spend this mana only to cast Aura and/or Equipment spells",
			Restrictions: []string{ManaRestrictCast, ManaRestrictSubtype("Aura|Equipment")},
		}},
		Activated: []ActivatedAbility{{
			Label:        "{T}: Attach target Aura or Equipment you control to target creature you control. Activate only as a sorcery.",
			Cost:         TapCost(),
			SorcerySpeed: true,
			Targets: Clauses(
				TargetPermanent("target Aura or Equipment you control", Or(HasSubtype("Aura"), HasSubtype("Equipment")), YouControl()),
				TargetCreature("target creature you control", YouControl()),
			),
			Effect: AttachClauseTargets,
		}},
	})
}
