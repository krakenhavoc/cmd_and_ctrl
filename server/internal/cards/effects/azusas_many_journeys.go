package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
const azusasManyJourneysOracleID = "02c60b8b-f665-4279-a00f-ae03991df4a6"

func init() {
	const name = "Azusa's Many Journeys"
	Register(Spec{
		OracleID:     azusasManyJourneysOracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "you may play an additional land this turn"),
				func(g *game.Game, item *game.StackItem) error {
					g.GrantAdditionalLandPlayForEffect(item.Controller, 1)
					return nil
				}),
			TriggerWithPurpose(ChapterTrigger(2, SagaChapterLabel(name, 2, "you gain 3 life"),
				Do(GainLife{Amount: 3})), game.Purpose{LifeGain: 3}),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "exile it, then return it transformed"),
				ChapterExileAndReturnTransformed),
		},
	})
	Register(Spec{
		OracleID:     azusasManyJourneysOracleID + "#1",
		Name:         "Likeness of the Seeker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBecomesBlocked, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			}, "Likeness of the Seeker — untap up to three lands you control",
				Do(UntapUpToLands{N: 3, Question: "Likeness of the Seeker — untap up to three lands you control"})),
		},
	})
}
