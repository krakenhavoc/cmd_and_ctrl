package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seluma, Light of Aysen — Legendary Creature — Angel Warrior {4}{W}, 4/4
// (Mystery Booster Commander, #5948):
//
//	"Rulebreaker — A deck with this commander can have Angel cards of any
//	 color identity and any basic land cards.
//	 Flying
//	 Whenever Seluma deals combat damage to a player, return target Angel
//	 creature card from your graveyard to the battlefield."
//
// The trigger is Moira's: combat damage to a player, a target card in your
// graveyard (an Angel creature card), put onto the battlefield under your
// control (sepulchralPrimordialEffect). The Rulebreaker sentence is deck
// construction and lives in the deck validator (internal/deck).
//
// No simplification.
func init() {
	trig := WheneverThisDealsCombatDamageToAPlayer("Seluma, Light of Aysen — return target Angel creature card from your graveyard to the battlefield",
		sepulchralPrimordialEffect)
	trig.Targets = TargetCardInGraveyard("target Angel creature card in your graveyard",
		YouOwn(), Creature(), HasSubtype("Angel"))
	Register(Spec{
		OracleID:        "dbbf6379-01b4-400d-a277-5887e6be1d0f",
		Name:            "Seluma, Light of Aysen",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{trig},
	})
}
