package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Triad of Fates — Legendary Creature — Human Wizard {2}{W}{B}, 3/3 (Theros):
//
//	"{1}, {T}: Put a fate counter on another target creature.
//	 {W}, {T}: Exile target creature that has a fate counter on it, then
//	 return it to the battlefield under its owner's control.
//	 {B}, {T}: Exile target creature that has a fate counter on it. Its
//	 controller draws two cards."
//
// Fate counters are Oblivion Stone's kind. The blink is Icewind Stalwart's
// (a new object, so it comes back without its counter, CR 400.7); the exile's
// "its controller" is read before the exile, as the creature last existed
// (CR 608.2h), and draws even if the exile was replaced.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fd09875f-71a1-41c2-b404-ff58f4d4cb6c",
		Name:         "Triad of Fates",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}, {T}: Put a fate counter on another target creature.",
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
				Targets: effects.Another(effects.TargetCreature("another target creature")),
				Effect:  effects.OldB396PutFateCounterOnTarget,
			},
			{
				Label:   "{W}, {T}: Exile target creature that has a fate counter on it, then return it to the battlefield under its owner's control.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature that has a fate counter on it", effects.OldB396HasFateCounter()),
				Effect:  effects.FlickerFirstLegalTargetUnderOwner,
			},
			{
				Label:   "{B}, {T}: Exile target creature that has a fate counter on it. Its controller draws two cards.",
				Purpose: game.Purpose{Answers: game.AnswerRemove},
				Cost:    effects.Plus(effects.ManaCost("{B}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature that has a fate counter on it", effects.OldB396HasFateCounter()),
				Effect:  effects.OldB396ExileTargetItsControllerDraws(2),
			},
		},
	})
}
