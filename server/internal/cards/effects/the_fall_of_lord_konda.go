package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     theFallOfLordKondaOracleID,
		Name:         "The Fall of Lord Konda",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(ChapterTriggerTargeting(1, SagaChapterLabel("The Fall of Lord Konda", 1, "exile target creature an opponent controls with mana value 4 or greater"),
				TargetCreature("target creature an opponent controls with mana value 4 or greater", OpponentControls(), ManaValueGE(4)),
				ExileFirstTarget), game.Purpose{Answers: game.AnswerRemove}),
			ChapterTrigger(2, SagaChapterLabel("The Fall of Lord Konda", 2, "each player gains control of all permanents they own"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item).asGroupMember()
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == c.Owner {
							continue
						}
						if err := (GainControl{
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
			ChapterTrigger(3, SagaChapterLabel("The Fall of Lord Konda", 3, "exile it, then return it transformed"),
				ChapterExileAndReturnTransformed),
		},
	})

	Register(Spec{
		OracleID:        theFallOfLordKondaOracleID + "#1",
		Name:            "Fragment of Konda",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Fragment of Konda — draw a card", Do(DrawCards{N: 1})),
		},
	})
}

// theFallOfLordKondaOracleID is shared with the back face ("#1").
const theFallOfLordKondaOracleID = "6ba8e099-5871-4ee3-8a65-06fd8ccec340"
