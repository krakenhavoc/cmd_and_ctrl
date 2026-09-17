package effects

import (
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
	Register(Spec{
		OracleID:     "c733873e-77db-471f-8061-139db24f7e7c",
		Name:         "Cephalid Coliseum",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}. This land deals 1 damage to you.",
			Rider:    PainRider(1),
		}},
		Activated: []ActivatedAbility{{
			Label:     "Threshold — {U}, {T}, Sacrifice this land: Target player draws three cards, then discards three cards. Activate only if there are seven or more cards in your graveyard.",
			Cost:      Plus(ManaCost("{U}"), TapCost(), SacrificeThis()),
			Targets:   TargetPlayer("target player"),
			Condition: GraveyardAtLeast(7, nil),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetPlayer {
						continue
					}
					if err := (DrawCards{Player: t.ID, N: 3}).Apply(ctx); err != nil {
						return err
					}
					g.DiscardChoiceForEffect(t.ID, 3)
				}
				return nil
			},
		}},
	})
}
