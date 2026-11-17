package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Moira, Urborg Haunt — Legendary Creature — Spirit Wizard {2}{B}, 3/2:
//
//	"Menace
//	 Whenever Moira deals combat damage to a player, return to the
//	 battlefield target creature card in your graveyard that was put there
//	 from the battlefield this turn."
//
// "Put there from the battlefield this turn" is Samwise's reading of the
// turn's events (putThereFromTheBattlefieldThisTurn). The body is Sepulchral
// Primordial's: every still-legal pick enters under its controller's control,
// which is its owner, since the clause names your graveyard (CR 400.7: it is
// a new object). A target that left the
// graveyard in response is dropped by the CR 608.2b re-check.
//
// No simplification.
func init() {
	trig := WheneverThisDealsCombatDamageToAPlayer("Moira, Urborg Haunt — return a creature card that died this turn to the battlefield",
		sepulchralPrimordialEffect)
	trig.Targets = TargetCardInGraveyard("target creature card in your graveyard that was put there from the battlefield this turn",
		YouOwn(), Creature(), putThereFromTheBattlefieldThisTurn)
	Register(Spec{
		OracleID:        "48060406-dd5a-4df5-84e9-3e797f0248cd",
		Name:            "Moira, Urborg Haunt",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered:       []game.TriggeredAbility{trig},
	})
}
