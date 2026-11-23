package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wispdrinker Vampire — Creature — Vampire Rogue {2}{W}{B} (Murders at
// Karlov Manor, #4469):
//
//	"Flying
//	 Whenever another creature you control with power 2 or less enters,
//	 each opponent loses 1 life and you gain 1 life.
//	 {5}{W}{B}: Creatures you control with power 2 or less gain deathtouch
//	 and lifelink until end of turn."
//
// The trigger judges the power as the creature enters, counters included
// (AnotherCreatureWithPower2OrLessEnteredUnderYourControl), so a face-down
// 2/2 counts. The activated ability affects the creatures you control with
// power 2 or less as it resolves (CR 611.2c): ones that arrive later get
// nothing, and the set is fixed then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6317d45a-fe02-4879-b515-bc5327ecb5ae",
		Name:            "Wispdrinker Vampire",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AnotherCreatureWithPower2OrLessEnteredUnderYourControl,
				"Wispdrinker Vampire — each opponent loses 1 life and you gain 1 life",
				effects.Do(effects.EachOpponentLosesLife{N: 1}, effects.GainLife{Amount: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{W}{B}: Creatures you control with power 2 or less gain deathtouch and lifelink until end of turn.",
			Cost:    effects.ManaCost("{5}{W}{B}"),
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller != item.Controller || !c.IsCreature() || c.CurrentPower() > 2 {
						continue
					}
					if err := (effects.GrantKeywordUntilEOT{
						Target:   c.InstanceID,
						Keywords: []string{"deathtouch", "lifelink"},
						Label:    "Wispdrinker Vampire — deathtouch and lifelink until end of turn",
					}).Apply(ctx.AsGroupMember()); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
