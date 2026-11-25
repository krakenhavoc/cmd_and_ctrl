package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Niko Defies Destiny — Enchantment — Saga {1}{W}{U} (Kaldheim, #5505):
//
//	"I — You gain 2 life for each foretold card you own in exile.
//	 II — Add {W}{U}. Spend this mana only to foretell cards or cast spells
//	 that have foretell.
//	 III — Return target card with foretell from your graveyard to your hand."
//
// Chapter I counts, as it resolves, the cards you own in exile that are face
// down by foretell (CR 702.143b). Chapter II's mana is spendable only on the
// foretell special action's {2} (ManaRestrictForetell, new with this card) or
// on a spell that has foretell, cast for any cost (ManaRestrictSpellForetell);
// it pays nothing else, and no other mana restriction of the pool admits it.
// Chapter III's target is a card in your graveyard that prints foretell
// (game.CardHasForetell), rechecked as the chapter resolves.
//
// No simplification.
func init() {
	const name = "Niko Defies Destiny"
	effects.Register(effects.Spec{
		OracleID:     "e79bf038-071e-448e-ad95-234c207a9575",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "you gain 2 life for each foretold card you own in exile"),
				func(g *game.Game, item *game.StackItem) error {
					n := 0
					for _, c := range g.Exile.Cards {
						if c.Owner == item.Controller && game.CardIsForetold(c) {
							n++
						}
					}
					return effects.GainLife{Amount: 2 * n}.Apply(effects.NewContext(g, item))
				}),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "add {W}{U}; spend this mana only to foretell cards or cast spells that have foretell"),
				effects.Do(effects.AddMana{
					Produced: "{W}{U}",
					Restrictions: []string{game.ManaRestrictAnyOf(
						[]string{game.ManaRestrictForetell},
						[]string{game.ManaRestrictSpellForetell},
					)},
				})),
			effects.ChapterTriggerTargeting(3, effects.SagaChapterLabel(name, 3, "return target card with foretell from your graveyard to your hand"),
				effects.TargetCardInGraveyard("target card with foretell from your graveyard", effects.YouOwn(), nikoDefiesDestinyHasForetell),
				func(g *game.Game, item *game.StackItem) error {
					return effects.ReturnLegalGraveyardTargetsToHand(effects.NewContext(g, item))
				}),
		},
	})
}

// nikoDefiesDestinyHasForetell is "card with foretell".
func nikoDefiesDestinyHasForetell(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	return game.CardHasForetell(c)
}
