package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cephalid Coliseum — Land (EDHREC rank 853):
//
//	"{T}: Add {U}. This land deals 1 damage to you.
//	 Threshold — {U}, {T}, Sacrifice this land: Target player draws
//	 three cards, then discards three cards. Activate only if there are
//	 seven or more cards in your graveyard."
//
// Odyssey's blue threshold land: a pain {U}, and a sacrifice that
// wheels three for any player. Threshold is the activation condition
// (CR 602.1b, #743), GraveyardAtLeast(7) over any cards. The target
// draws three, then chooses three to discard from the hand those draws
// made — the draw comes first, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c733873e-77db-471f-8061-139db24f7e7c",
		Name:         "Cephalid Coliseum",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:      effects.ManaAbilityCost{Tap: true},
			Produced:  "{U}",
			Label:     "Add {U}. This land deals 1 damage to you.",
			PainToYou: 1,
		}},
		Activated: []effects.ActivatedAbility{{
			Label:     "Threshold — {U}, {T}, Sacrifice this land: Target player draws three cards, then discards three cards. Activate only if there are seven or more cards in your graveyard.",
			Cost:      effects.Plus(effects.ManaCost("{U}"), effects.TapCost(), effects.SacrificeThis()),
			Targets:   effects.TargetPlayer("target player"),
			Purpose:   effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 3, Discards: 3}),
			Condition: effects.GraveyardAtLeast(7, nil),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetPlayer {
						continue
					}
					if err := effects.DrawThenDiscard(g, t.ID, item.SourceCardID, 3, 3, ""); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
