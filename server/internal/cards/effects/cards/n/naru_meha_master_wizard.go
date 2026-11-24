package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Naru Meha, Master Wizard — Legendary Creature — Human Wizard {2}{U}{U},
// 3/3 (Commander 2021):
//
//	"Flash
//	 When Naru Meha enters, copy target instant or sorcery spell you control.
//	 You may choose new targets for the copy.
//	 Other Wizards you control get +1/+1."
//
// Dualcaster Mage's enters trigger over the spells its controller controls
// (the Wizard's own flash lets it be cast with the spell still on the
// stack), and Leaf-Crowned Visionary's anthem for the other Wizards its
// controller has. A spell that leaves the stack before the trigger
// resolves is nothing to copy (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "562c9c68-e5a7-45ab-b098-5b649d165ff5",
		Name:            "Naru Meha, Master Wizard",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Wizard"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: effects.TargetInstantOrSorcerySpell("target instant or sorcery spell you control", effects.YouControl()),
			Key:     "Naru Meha, Master Wizard — copy target instant or sorcery spell you control",
			Effect:  effects.CopyTargetedSpellWithNewTargets,
		}},
	})
}
