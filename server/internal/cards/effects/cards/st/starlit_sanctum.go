package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Starlit Sanctum — Land (Commander Legends: Battle for Baldur's Gate, S82
// GAP #5948):
//
//	"{T}: Add {C}.
//	 {W}, {T}, Sacrifice a Cleric creature: You gain life equal to the
//	 sacrificed creature's toughness.
//	 {B}, {T}, Sacrifice a Cleric creature: Target player loses life equal to
//	 the sacrificed creature's power."
//
// Miren's sacrifice outlet with a Cleric filter, read from last-known
// information (CR 608.2h): the toughness (or power) the Cleric had on the
// battlefield, counters and anthems included; zero or less is nothing. The
// second ability targets a player and checks it as it resolves.
func init() {
	cleric := effects.Plus(effects.TapCost(), game.AbilityCost{SacrificeOther: effects.SacrificeSpec("a Cleric creature", effects.Creature(), effects.Subtype("Cleric"))})
	effects.Register(effects.Spec{
		OracleID:     "d16298ac-67bd-4f9d-9979-23c1b7e4b359",
		Name:         "Starlit Sanctum",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{W}, {T}, Sacrifice a Cleric creature: You gain life equal to the sacrificed creature's toughness.",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
				Cost:    effects.Plus(effects.ManaCost("{W}"), cleric),
				Effect:  effects.GainLifeEqualToSacrificedToughness,
			},
			{
				Label:   "{B}, {T}, Sacrifice a Cleric creature: Target player loses life equal to the sacrificed creature's power.",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
				Cost:    effects.Plus(effects.ManaCost("{B}"), cleric),
				Targets: effects.TargetPlayer("target player"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					fed, ok := effects.B17PermanentSacrificedToPay(g, item)
					if !ok {
						return nil
					}
					n := effects.DepartedCreaturePower(g, fed)
					if n <= 0 {
						return nil
					}
					for _, t := range effects.NewContext(g, item).LegalTargets() {
						if t.Kind == game.TargetPlayer {
							return g.ChangePlayerLifeForEffect(item.SourceCardID, t.ID, -n)
						}
					}
					return nil
				},
			},
		},
	})
}
