package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wizard's Staff — Artifact — Equipment {1}{U}:
//
//	"Equipped creature has prowess.
//	 If a triggered ability of equipped creature triggers, that ability
//	 triggers an additional time.
//	 Equip Wizard {1}
//	 Equip {3}"
//
// The doubling is Cloud, Midgar Mercenary's CR 603.2d shape, hung on the
// Equipment: it asks of every triggered ability whether its source is the
// creature this Equipment is attached to, as the ability triggers, so a
// creature that loses the Staff in response triggers once. It does not say
// "you control", so a stolen wearer still doubles. The prowess the Staff
// grants is itself a triggered ability of the equipped creature, so it
// triggers twice per noncreature spell. The two equip abilities are the
// Dúnedain Blade pair: the cheap one restricted to Wizards.
//
// No simplification. The Hobbit (#4454).
func init() {
	doubler := DoublesAbilitiesOf(nil, DoublesAbilitiesOfOptions{
		SkipControllerCheck: true,
		SourceMatch: func(_ *game.Game, q game.TriggerDoublingQuery) bool {
			return q.Doubler.IsAttachedTo(q.Source.InstanceID)
		},
	})
	doubler.Label = "Wizard's Staff"
	doubler.RowLabel = "If a triggered ability of equipped creature triggers, that ability triggers an additional time."
	Register(Spec{
		OracleID:        "30c3c700-46f4-4a77-8c45-5c7e3a21bd62",
		Name:            "Wizard's Staff",
		Completeness:    CompletenessFull,
		Static:          []game.StaticAbility{GrantToAttached("prowess")},
		TriggerDoublers: []game.TriggerDoubler{doubler},
		Activated: []ActivatedAbility{
			EquipOnlyAbility("Equip Wizard {1}", "{1}",
				TargetCreature("target Wizard you control", YouControl(), Subtype("Wizard"))),
			EquipAbility("{3}"),
		},
	})
}
