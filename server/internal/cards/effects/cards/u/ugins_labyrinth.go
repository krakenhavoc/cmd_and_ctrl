package cardsu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ugin's Labyrinth — Land (Modern Horizons 3, #5948):
//
//	"Imprint — When this land enters, you may exile a colorless card with
//	 mana value 7 or greater from your hand.
//	 {T}: Add {C}. If a card is exiled with this land, add {C}{C} instead.
//	 {T}: Return the exiled card to its owner's hand."
//
// Chrome Mox's imprint: a "when this land enters" trigger (so it uses the
// stack and can be answered), a pick from the hand with a floor of zero for
// the "may", and the link read back off the event log (b27ExiledWith). The
// mana ability reads that link as it is activated, so a land that imprinted
// nothing, or whose card has since left exile, taps for {C} alone. The third
// ability returns whatever is still exiled with it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d565cd3d-68d4-4039-9e45-7e69e31d0ffb",
		Name:         "Ugin's Labyrinth",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(uginsLabyrinthLabel, uginsLabyrinthImprint),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost: effects.ManaAbilityCost{Tap: true},
			ProducedFunc: func(g *game.Game, _, source uuid.UUID) string {
				if len(effects.B27ExiledWith(g, source, uginsLabyrinthLabel)) > 0 {
					return "{C}{C}"
				}
				return "{C}"
			},
			Label: "Add {C}. If a card is exiled with this land, add {C}{C} instead",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Return the exiled card to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				for _, id := range effects.B27ExiledWith(g, item.SourceCardID, uginsLabyrinthLabel) {
					if err := g.BounceToHandForEffect(id); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}

// uginsLabyrinthLabel is the imprint trigger's stack label; the "exiled with
// this land" record keys on it.
const uginsLabyrinthLabel = "Ugin's Labyrinth — exile a colorless card with mana value 7 or greater from your hand"

func uginsLabyrinthImprint(g *game.Game, item *game.StackItem) error {
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil || p.Hand == nil {
		return nil
	}
	var candidates []uuid.UUID
	for _, c := range p.Hand.Cards {
		if c.IsColorless() && c.ManaValue() >= 7 {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: uginsLabyrinthLabel,
		Cards:    candidates,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneHand,
		ThenKey:  effects.ChromeMoxExileThePick,
	})
	return nil
}
