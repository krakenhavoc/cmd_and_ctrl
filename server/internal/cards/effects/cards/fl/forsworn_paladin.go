package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Forsworn Paladin — Creature — Human Knight {B}, 1/1 (Adventures in the
// Forgotten Realms):
//
//	"Menace
//	 {1}{B}, {T}, Pay 1 life: Create a Treasure token.
//	 {2}{B}: Target creature gets +2/+0 until end of turn. If mana from a
//	 Treasure was spent to activate this ability, that creature also gains
//	 deathtouch until end of turn."
//
// The Treasure rider reads the activation's own payment as it resolves
// (Jetmir's Fixer's reading: ReadsOwnActivations with ReadsSourceKinds). Mana
// the engine did not see counts as no Treasure mana: weaker, never stronger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d3564614-4463-4683-83f0-07561ab3431c",
		Name:            "Forsworn Paladin",
		ReadsManaSpent:  game.ReadsOwnActivations | game.ReadsSourceKinds,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{B}, {T}, Pay 1 life: Create a Treasure token.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.TapCost(), effects.PayLife(1)),
				Effect:  effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1}),
			},
			{
				Label:   "{2}{B}: Target creature gets +2/+0 until end of turn. If mana from a Treasure was spent to activate this ability, that creature also gains deathtouch until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 2, 0))},
				Cost:    effects.ManaCost("{2}{B}"),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
						return nil
					}
					ctx := effects.NewContext(g, item)
					target := item.Targets[0].ID
					if err := (effects.BoostUntilEOT{Target: target, Power: 2, Label: "Forsworn Paladin — +2/+0 until end of turn"}).Apply(ctx); err != nil {
						return err
					}
					if !ctx.ManaSpent().FromTreasure() {
						return nil
					}
					return effects.GrantKeywordUntilEOT{Target: target, Keywords: []string{"deathtouch"},
						Label: "Forsworn Paladin — deathtouch until end of turn"}.Apply(ctx)
				},
			},
		},
	})
}
