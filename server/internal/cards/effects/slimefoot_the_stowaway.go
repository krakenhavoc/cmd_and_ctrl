package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Slimefoot, the Stowaway — Legendary Creature — Fungus {1}{B}{G}, 2/3
// (Commander Masters):
//
//	"Whenever a Saproling you control dies, Slimefoot deals 1 damage to
//	 each opponent and you gain 1 life.
//	 {4}: Create a 1/1 green Saproling creature token."
//
// The drain reads the Saproling as it died (CR 603.10a), so a token
// counts and an opponent's Saproling does not. Slimefoot may itself be the
// one that has gone: the damage is dealt by it as it last existed (CR 608.2h).
// "Saproling" is a word of this card's text (CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fd541139-de13-4b32-93ba-070ecf711d33",
		Name:         "Slimefoot, the Stowaway",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(On(game.EventLTB, gapB101SaprolingYouControlDied,
				"Slimefoot, the Stowaway — 1 damage to each opponent; gain 1 life",
				Do(DamageEachOpponent{N: 1}, GainLife{Amount: 1})), game.Purpose{DeathPayoff: true}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}: Create a 1/1 green Saproling creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    ManaCost("{4}"),
			Effect:  Do(CreateToken{Template: TokenCard("1/1 green Saproling"), N: 1}),
		}},
	})
}

// gapB101SaprolingYouControlDied is "whenever a Saproling you control dies".
func gapB101SaprolingYouControlDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := diedCreature(ev, g)
	return ok && leftUnderControlOf(ev, dead) == source.Controller && leftAsSubtype(ev, dead, source.TextSubtype("Saproling"))
}
