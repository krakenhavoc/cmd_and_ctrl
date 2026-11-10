package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sunset Saboteur — Creature — Human Rogue {1}{B}, 4/1 (Edge of
// Eternities):
//
//	"Menace
//	 Ward—Discard a card.
//	 Whenever this creature attacks, put a +1/+1 counter on target
//	 creature an opponent controls."
//
// Ward's cost is a discard (CR 702.21a, WardDiscard): the payer picks the
// card, and a payer with an empty hand cannot pay, so the spell or
// ability is countered. The attack trigger targets a creature an
// opponent controls and puts the counter on it if it is still a legal
// target as the trigger resolves (CR 608.2b).
//
// No simplification.
func init() {
	attack := WheneverThisAttacks("Sunset Saboteur — put a +1/+1 counter on target creature an opponent controls", rfCreatureFCountersOnTarget(1))
	attack.Targets = TargetCreature("target creature an opponent controls", OpponentControls())
	Register(Spec{
		OracleID:        "c3731735-2b78-4a15-bff4-4d8559401f68",
		Name:            "Sunset Saboteur",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			Ward(WardDiscard(1), "Sunset Saboteur — ward, discard a card"),
			attack,
		},
	})
}
