package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Fall of Lord Konda // Fragment of Konda — a transforming Saga
// (Kamigawa: Neon Dynasty, #5493).
//
//	I   — Exile target creature an opponent controls with mana value 4
//	      or greater.
//	II  — Each player gains control of all permanents they own.
//	III — Exile this Saga, then return it to the battlefield transformed
//	      under your control.
//
// Fragment of Konda, Enchantment Creature — Human Noble 1/3: "Defender.
// When this creature dies, draw a card."
//
// Chapter II hands back every permanent whose controller is not its
// owner, whoever took it and however (Alicia Masters' "regain" shape,
// run for every seat). Control effects that already gave a permanent to
// its owner leave nothing to do.
func init() {
	effects.Register(effects.Spec{
		OracleID:     TheFallOfLordKondaOracleID,
		Name:         "The Fall of Lord Konda",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.ChapterTriggerTargeting(1, effects.SagaChapterLabel("The Fall of Lord Konda", 1, "exile target creature an opponent controls with mana value 4 or greater"),
				effects.TargetCreature("target creature an opponent controls with mana value 4 or greater", effects.OpponentControls(), effects.ManaValueGE(4)),
				effects.ExileFirstTarget), game.Purpose{Answers: game.AnswerRemove}),
			effects.ChapterTrigger(2, effects.SagaChapterLabel("The Fall of Lord Konda", 2, "each player gains control of all permanents they own"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item).AsGroupMember()
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == c.Owner {
							continue
						}
						if err := (effects.GainControl{
							Target:     c.InstanceID,
							Controller: c.Owner,
							Duration:   game.IndefiniteDuration(),
							Label:      "The Fall of Lord Konda — its owner gains control of it",
						}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
			effects.ChapterTrigger(3, effects.SagaChapterLabel("The Fall of Lord Konda", 3, "exile it, then return it transformed"),
				effects.ChapterExileAndReturnTransformed),
		},
	})

	effects.Register(effects.Spec{
		OracleID:        TheFallOfLordKondaOracleID + "#1",
		Name:            "Fragment of Konda",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Fragment of Konda — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}

// theFallOfLordKondaOracleID is shared with the back face ("#1").
const TheFallOfLordKondaOracleID = "6ba8e099-5871-4ee3-8a65-06fd8ccec340"
