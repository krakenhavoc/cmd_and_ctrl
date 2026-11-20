package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Obelisk of Alara — Artifact {6} (Conflux, #5587):
//
//	"{1}{W}, {T}: You gain 5 life.
//	 {1}{U}, {T}: Draw a card, then discard a card.
//	 {1}{B}, {T}: Target creature gets -2/-2 until end of turn.
//	 {1}{R}, {T}: This artifact deals 3 damage to target player or
//	 planeswalker.
//	 {1}{G}, {T}: Target creature gets +4/+4 until end of turn."
//
// Five {T} rows, so one a turn: Fountain of Youth's life, Strix
// Lookout's loot, Grandmother Sengir's shrink, Aethertorch Renegade's
// damage to a player or planeswalker and Captive Flame's pump. An
// artifact's {T} is not held back by summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "75fcb9ce-ac01-4048-82b7-9100143e3b07",
		Name:         "Obelisk of Alara",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{1}{W}, {T}: You gain 5 life.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{1}{W}"), TapCost()),
				Effect:  Do(GainLife{Amount: 5}),
			},
			{
				Label:   "{1}{U}, {T}: Draw a card, then discard a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{1}{U}"), TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return b16DrawThenDiscard(g, item, 1)
				},
			},
			{
				Label:   "{1}{B}, {T}: Target creature gets -2/-2 until end of turn.",
				Purpose: ForTargets(HarmUntilEOT(0, -2, -2)),
				Cost:    Plus(ManaCost("{1}{B}"), TapCost()),
				Targets: TargetCreature("target creature"),
				Effect:  boostTargetUntilEOT("Obelisk of Alara — -2/-2 until end of turn", -2, -2),
			},
			{
				Label:   "{1}{R}, {T}: This artifact deals 3 damage to target player or planeswalker.",
				Purpose: ForTargets(DamageToTarget(0, 3)),
				Cost:    Plus(ManaCost("{1}{R}"), TapCost()),
				Targets: targetPlayerOrPlaneswalker(),
				Effect:  sourceDealsDamageToEachLegalTarget(3),
			},
			{
				Label:   "{1}{G}, {T}: Target creature gets +4/+4 until end of turn.",
				Purpose: ForTargets(BenefitUntilEOT(0, 4, 4)),
				Cost:    Plus(ManaCost("{1}{G}"), TapCost()),
				Targets: TargetCreature("target creature"),
				Effect:  boostTargetUntilEOT("Obelisk of Alara — +4/+4 until end of turn", 4, 4),
			},
		},
	})
}
