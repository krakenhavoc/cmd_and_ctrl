package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shipwreck Singer — Creature — Siren {U}{B}, 1/2 (Conspiracy: Take the
// Crown, S82 older sets batch 273, #5540):
//
//	"Flying
//	 {1}{U}: Target creature an opponent controls attacks this turn if able.
//	 {1}{B}, {T}: Attacking creatures get -1/-1 until end of turn."
//
// Incite's attack requirement (a scoped CR 508.1d requirement until end of
// turn on the target) and Warmonger Hellkite's pump over every attacking
// creature, whoever controls it: the set is fixed as the ability resolves
// (CR 611.2c), so a creature that attacks later is untouched.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8aae6f84-02ac-4f8f-9721-b12138588b6b",
		Name:            "Shipwreck Singer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{U}: Target creature an opponent controls attacks this turn if able.",
				Cost:    effects.ManaCost("{1}{U}"),
				Targets: effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.ScopedEffectFor{
						Target:   id,
						Mods:     []game.Mod{game.AddAttackRequirementMod(uuid.Nil)},
						Duration: effects.DurationUntilEndOfTurn(ctx),
						Label:    "Shipwreck Singer — attacks this turn if able",
					}.Apply(ctx)
				},
			},
			{
				Label:   "{1}{B}, {T}: Attacking creatures get -1/-1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerRemove},
				Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{Match: effects.AttackingCreature(), Power: -1, Toughness: -1,
						Label: "Shipwreck Singer — attacking creatures get -1/-1"}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
