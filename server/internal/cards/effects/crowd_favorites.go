package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crowd Favorites — Creature — Human Soldier {6}{W}, 4/4 (Onslaught):
//
//	"{3}{W}: Tap target creature.
//	 {3}{W}: This creature gets +0/+5 until end of turn."
//
// Two activated abilities paid in mana alone, with no {T} symbol, so
// both may be activated the turn it arrives and as often as its
// controller can pay. The tap is
// Arashin Sunshield's body on any creature; the toughness pump has no
// target and declares that it pumps (ADR 0142).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1ead750f-14a6-4f25-9eb8-9472c2fdac35",
		Name:         "Crowd Favorites",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{3}{W}: Tap target creature.",
				Cost:    ManaCost("{3}{W}"),
				Targets: TargetCreature("target creature"),
				Effect:  tapChosenPermanent,
			},
			{
				Label:   "{3}{W}: This creature gets +0/+5 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Toughness: 5}},
				Cost:    ManaCost("{3}{W}"),
				Effect:  thisGetsUntilEndOfTurn(0, 5, "Crowd Favorites — +0/+5 until end of turn"),
			},
		},
	})
}
