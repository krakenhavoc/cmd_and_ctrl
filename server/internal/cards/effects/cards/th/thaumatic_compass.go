package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thaumatic Compass // Spires of Orazca — Artifact {2} (Ixalan, S82 older
// sets):
//
//	"{3}, {T}: Search your library for a basic land card, reveal it, put it
//	 into your hand, then shuffle.
//	 At the beginning of your end step, if you control seven or more lands,
//	 transform this artifact."
//
//	Spires of Orazca — Land:
//	"{T}: Add {C}.
//	 {T}: Untap target attacking creature an opponent controls and remove it
//	 from combat."
//
// The end-step "if" is intervening (CR 603.4): checked as the end step begins
// and again on resolution. The untap-and-remove is the land's own tap
// ability (not a mana ability), aimed at an attacking creature an opponent
// controls.
//
// No simplification.
func init() {
	const oracle = "f9085e55-2833-41b7-9100-a35dc04dee93"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Thaumatic Compass",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Search your library for a basic land card, reveal it, put it into your hand, then shuffle.",
			Purpose: game.Purpose{Tutors: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.IsBasicLand,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Thaumatic Compass — a basic land card",
				}.Apply(effects.NewContext(g, item))
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("you control seven or more lands",
				func(g *game.Game, you uuid.UUID) bool { return effects.B02CountLandsControlledBy(g, you) >= 7 },
				effects.AtYourEndStep("Thaumatic Compass — transform this artifact", effects.Do(effects.TransformThis{}))),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Spires of Orazca",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "{T}: Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Untap target attacking creature an opponent controls and remove it from combat.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target attacking creature an opponent controls", effects.AttackingCreature(), effects.OpponentControls()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (effects.UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
					return g.RemoveFromCombatForEffect(t.ID)
				}
				return nil
			},
		}},
	})
}
