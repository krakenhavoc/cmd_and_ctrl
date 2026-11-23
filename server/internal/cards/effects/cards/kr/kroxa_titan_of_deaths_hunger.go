package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kroxa, Titan of Death's Hunger — Legendary Creature — Elder Giant {B}{R},
// 6/6:
//
//	"When Kroxa enters, sacrifice it unless it escaped.
//	 Whenever Kroxa enters or attacks, each opponent discards a card, then
//	 each opponent who didn't discard a nonland card this way loses 3 life.
//	 Escape—{B}{B}{R}{R}, Exile five other cards from your graveyard."
//
// Phlage's and Uro's shape for the escape and the sacrifice. Each
// opponent chooses their own discard; once every opponent has answered,
// the keyed continuation counts who did not discard a nonland card (an
// empty hand, or a land, or no discard at all) and takes 3 life from each
// of them. The card's type is read where it landed (CR 608.2c), so a
// discard that was replaced and left no card behind counts as not
// discarding a nonland card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "2cb1c213-b3f2-4433-8f89-6a3b4c3d2795",
		Name:             "Kroxa, Titan of Death's Hunger",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Escape("{B}{B}{R}{R}", 5)},
		Triggered: []game.TriggeredAbility{
			effects.SacrificeThisUnlessItEscaped("Kroxa"),
			effects.WhenThisEntersOrAttacks("Kroxa — each opponent discards a card, then each who didn't discard a nonland card loses 3 life",
				func(g *game.Game, item *game.StackItem) error {
					return g.EachPlayerDiscardsThenKeyedForEffect(item.Controller,
						game.DiscardPrompt{
							Source:   item.SourceCardID,
							N:        1,
							Question: "Kroxa — discard a card",
						},
						kroxaTitanLosesLife, []uuid.UUID{item.Controller})
				}),
		},
	})
}

// kroxaTitanLosesLife is "each opponent who didn't discard a nonland card
// this way loses 3 life", once every opponent has answered. An on-disk
// key: never renamed, never reused.
var kroxaTitanLosesLife = game.RegisterPromptRunThen("prompt-run/kroxa-titan-of-deaths-hunger-life", func(g *game.Game, r game.RunDone) error {
	if len(r.Carry) == 0 {
		return nil
	}
	controller := r.Carry[0]
	discarded := r.Discards()
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == controller {
			continue
		}
		nonland := false
		for _, id := range discarded.By(p.ID) {
			if c, found := g.LookupCardForEffect(id); found && !c.IsLand() {
				nonland = true
			}
		}
		if nonland {
			continue
		}
		if err := g.ChangePlayerLifeForEffect(r.Source, p.ID, -3); err != nil {
			return err
		}
	}
	return nil
})
