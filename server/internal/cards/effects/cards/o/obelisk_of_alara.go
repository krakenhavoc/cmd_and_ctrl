package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "75fcb9ce-ac01-4048-82b7-9100143e3b07",
		Name:         "Obelisk of Alara",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{W}, {T}: You gain 5 life.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.TapCost()),
				Effect:  effects.Do(effects.GainLife{Amount: 5}),
			},
			{
				Label:   "{1}{U}, {T}: Draw a card, then discard a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{1}{U}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.B16DrawThenDiscard(g, item, 1)
				},
			},
			{
				Label:   "{1}{B}, {T}: Target creature gets -2/-2 until end of turn.",
				Purpose: effects.ForTargets(effects.HarmUntilEOT(0, -2, -2)),
				Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.BoostTargetUntilEOT("Obelisk of Alara — -2/-2 until end of turn", -2, -2),
			},
			{
				Label:   "{1}{R}, {T}: This artifact deals 3 damage to target player or planeswalker.",
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 3)),
				Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.TapCost()),
				Targets: effects.TargetPlayerOrPlaneswalker(),
				Effect:  effects.SourceDealsDamageToEachLegalTarget(3),
			},
			{
				Label:   "{1}{G}, {T}: Target creature gets +4/+4 until end of turn.",
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 4, 4)),
				Cost:    effects.Plus(effects.ManaCost("{1}{G}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.BoostTargetUntilEOT("Obelisk of Alara — +4/+4 until end of turn", 4, 4),
			},
		},
	})
}
