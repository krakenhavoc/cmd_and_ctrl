package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Daring Fiendbonder — Creature — Human Warlock {3}{B}, 5/1 (Commander 2020,
// #5513):
//
//	"Haste
//	 This creature attacks each combat if able.
//	 {1}{B}, Exile this card from your graveyard: Put an indestructible
//	 counter on target creature. Activate only as a sorcery."
//
// The attack requirement is the AttacksEachCombat static (CR 508.1d). The
// graveyard ability is renew's shape without the keyword: it functions only
// from your graveyard, its cost exiles the card, and it is activated only
// as a sorcery. The counter lands on the target if it is still legal as the
// ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0a746c49-cc44-467a-85b7-3c498559aea7",
		Name:            "Daring Fiendbonder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Static:          []game.StaticAbility{AttacksEachCombat()},
		Activated: []ActivatedAbility{{
			Label:        "{1}{B}, Exile this card from your graveyard: Put an indestructible counter on target creature. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{1}{B}"), ExileThis()),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Targets:      TargetCreature("target creature"),
			Effect:       putACounterOnTheTarget(game.CounterIndestructible),
		}},
	})
}
