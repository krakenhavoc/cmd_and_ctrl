package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hardened Academic — Creature — Bird Cleric {R}{W}, 2/1 (Secrets of
// Strixhaven, #4456):
//
//	"Flying, haste
//	 Discard a card: This creature gains lifelink until end of turn.
//	 Whenever one or more cards leave your graveyard, put a +1/+1 counter on
//	 target creature you control."
//
// The discard is the activation's cost, paid up front, and the lifelink goes
// to this object only while it is still on the battlefield
// (thisCreatureUntilEOT). The trigger is a batch trigger over cards of yours
// that leave a graveyard, by a zone move or by being cast from it
// (socB26CardsLeaveYourGraveyard); its target is chosen as it goes on the
// stack and re-checked as it resolves (CR 608.2b). It targets any creature
// you control, not only this one.
//
// No simplification.
func init() {
	const lifelinkLabel = "Hardened Academic — lifelink until end of turn"
	Register(Spec{
		OracleID:        "921252df-6679-4f70-ae26-d4956c90b7df",
		Name:            "Hardened Academic",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Activated: []ActivatedAbility{{
			Label:   "Discard a card: This creature gains lifelink until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Keywords: []string{"lifelink"}}},
			Cost:    DiscardACard(),
			Effect:  thisCreatureUntilEOT(lifelinkLabel, 0, 0, "lifelink"),
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(socB26CardsLeaveYourGraveyard("Hardened Academic — a +1/+1 counter on target creature you control",
				PlusOneCounterOnTheTarget),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
