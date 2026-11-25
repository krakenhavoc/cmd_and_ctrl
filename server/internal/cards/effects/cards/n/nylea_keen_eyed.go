package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nylea, Keen-Eyed — Legendary Enchantment Creature — God {3}{G}, 5/6
// (Theros Beyond Death, #5514):
//
//	"Indestructible
//	 As long as your devotion to green is less than five, Nylea isn't a
//	 creature.
//	 Creature spells you cast cost {1} less to cast.
//	 {2}{G}: Reveal the top card of your library. If it's a creature card,
//	 put it into your hand. Otherwise, you may put it into your graveyard."
//
// The God clause is the layer-4 devotion static every God shares. The cost
// reduction is Hamza's static slot with a flat {1}. The ability is Kellan's
// reveal: the card is shown to every player, a creature card goes to the
// hand, anything else may be put into the graveyard (a plain move, not a
// mill) and declining leaves it on top.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "642bbdf6-9a8c-4158-83c8-d59feb49101b",
		Name:            "Nylea, Keen-Eyed",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Static:          []game.StaticAbility{effects.GodUnlessDevotion("G")},
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Creature spells you cast cost {1} less to cast.",
				effects.YourSpell(), effects.CreatureSpell()),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}: Reveal the top card of your library. If it's a creature card, put it into your hand. Otherwise, you may put it into your graveyard.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{2}{G}"),
			Effect:  effects.OldB137NyleaRevealTop,
		}},
	})
}
