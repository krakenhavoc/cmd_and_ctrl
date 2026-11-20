package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Azusa's Many Journeys // Likeness of the Seeker — a transforming Saga
// (Kamigawa: Neon Dynasty, #5493, S82).
//
// Front face, Enchantment — Saga {1}{G}:
//
//	"I — You may play an additional land this turn.
//	 II — You gain 3 life.
//	 III — Exile this Saga, then return it to the battlefield transformed
//	       under your control."
//
// Back face, Enchantment Creature — Human Monk, 3/3:
//
//	"Whenever this creature becomes blocked, untap up to three lands you
//	 control."
//
// Chapter I is Explore's one-shot extra land drop. The untap is a "choose up
// to three" prompt over your tapped lands (UntapUpToLands), asked as the
// trigger resolves; "becomes blocked" is once per blocked attacker.
//
// No simplification.
const AzusasManyJourneysOracleID = "02c60b8b-f665-4279-a00f-ae03991df4a6"

func init() {
	const name = "Azusa's Many Journeys"
	effects.Register(effects.Spec{
		OracleID:     AzusasManyJourneysOracleID,
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "you may play an additional land this turn"),
				func(g *game.Game, item *game.StackItem) error {
					g.GrantAdditionalLandPlayForEffect(item.Controller, 1)
					return nil
				}),
			effects.TriggerWithPurpose(effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "you gain 3 life"),
				effects.Do(effects.GainLife{Amount: 3})), game.Purpose{LifeGain: 3}),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "exile it, then return it transformed"),
				effects.ChapterExileAndReturnTransformed),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     AzusasManyJourneysOracleID + "#1",
		Name:         "Likeness of the Seeker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			}, "Likeness of the Seeker — untap up to three lands you control",
				effects.Do(effects.UntapUpToLands{N: 3, Question: "Likeness of the Seeker — untap up to three lands you control"})),
		},
	})
}
