package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soul of Shandalar — Creature — Avatar {4}{R}{R}, 6/6 (Magic 2015, #5556):
//
//	"First strike
//	 {3}{R}{R}: This creature deals 3 damage to target player or planeswalker
//	 and 3 damage to up to one target creature that player or that
//	 planeswalker's controller controls.
//	 {3}{R}{R}, Exile this card from your graveyard: It deals 3 damage to
//	 target player or planeswalker and 3 damage to up to one target creature
//	 that player or that planeswalker's controller controls."
//
// Soul of Zendikar's two rows (the battlefield one and a graveyard one whose
// cost exiles the card, so it resolves with the source gone), over Ravager of
// the Fells' two clauses: the creature must be controlled by the first
// target's controller (ControlledByTargetsController), and each target still
// legal takes its 3 damage as one instruction (CR 120).
//
// No simplification.
func init() {
	const label1 = "{3}{R}{R}: This creature deals 3 damage to target player or planeswalker and 3 damage to up to one target creature that player or that planeswalker's controller controls."
	const label2 = "{3}{R}{R}, Exile this card from your graveyard: It deals 3 damage to target player or planeswalker and 3 damage to up to one target creature that player or that planeswalker's controller controls."
	effects.Register(effects.Spec{
		OracleID:        "ac96d27c-dbae-4094-9209-3ef143f64713",
		Name:            "Soul of Shandalar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   label1,
				Cost:    effects.ManaCost("{3}{R}{R}"),
				Targets: effects.S82OldB358PlayerOrPlaneswalkerAndTheirCreature(),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 3)),
				Effect:  effects.S82OldB358DamageBothTargets(3),
			},
			{
				Label:   label2,
				Cost:    effects.Plus(effects.ManaCost("{3}{R}{R}"), effects.ExileThis()),
				Zones:   []game.ZoneKind{game.ZoneGraveyard},
				Targets: effects.S82OldB358PlayerOrPlaneswalkerAndTheirCreature(),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 3)),
				Effect:  effects.S82OldB358DamageBothTargets(3),
			},
		},
	})
}
