package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Mirari Conjecture — Enchantment — Saga {4}{U} (Dominaria):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Return target instant card from your graveyard to your hand.
//	 II — Return target sorcery card from your graveyard to your hand.
//	 III — Until end of turn, whenever you cast an instant or sorcery spell,
//	 copy it. You may choose new targets for the copy."
//
// I and II are targeted returns from your own graveyard, chosen as the
// chapter goes on the stack (CR 603.3d) and refused as it resolves if the
// card has gone (CR 608.2b). III is Doublecast's copy as a repeating delayed
// trigger (CR 603.7b) over every instant or sorcery spell you cast until
// cleanup; the copy is created, not cast, so it does not trigger it again.
//
// No simplification.
func init() {
	const name = "The Mirari Conjecture"
	effects.Register(effects.Spec{
		OracleID:     "56528234-46c0-4fba-972a-0c8011996f8d",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTriggerTargeting(1, effects.SagaChapterLabel(name, 1, "return target instant card from your graveyard to your hand"),
				effects.TargetCardInGraveyard("target instant card from your graveyard", effects.YouOwn(), effects.Instant()), effects.S82Mom16ReturnTargetToHand),
			effects.ChapterTriggerTargeting(2, effects.SagaChapterLabel(name, 2, "return target sorcery card from your graveyard to your hand"),
				effects.TargetCardInGraveyard("target sorcery card from your graveyard", effects.YouOwn(), effects.Sorcery()), effects.S82Mom16ReturnTargetToHand),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "until end of turn, whenever you cast an instant or sorcery spell, copy it"),
				func(g *game.Game, item *game.StackItem) error {
					return effects.DelayedOnEvent{
						Label:      name + " — copy that spell; you may choose new targets",
						On:         []game.EventKind{game.EventCast},
						Condition:  effects.YouNextCastCondition,
						CondParams: game.EffectParams{Filter: game.CastFilter{Types: []string{"Instant", "Sorcery"}}},
						Body:       effects.CopyTheSpellBody,
						Repeats:    true,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
