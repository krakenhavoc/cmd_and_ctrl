package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ellie and Alan, Paleontologists — Legendary Creature — Human Scientist
// {2}{G}{W}{U}, 2/5:
//
//	"{T}, Exile a creature card from your graveyard: Discover X, where X
//	 is the mana value of the exiled card. Activate only as a sorcery."
//
// The exile is a cost (#1297): the card is picked at announce and is in
// exile before the ability is on the stack, so the effect reads it back
// through ctx.Exiled(). X is its mana value in exile. Discover is ADR
// 0099's (game/discover.go).
func init() {
	Register(Spec{
		OracleID:     "0f4d513e-100d-4cbb-b919-09249cef5fad",
		Name:         "Ellie and Alan, Paleontologists",
		Completeness: CompletenessFull,
		Discovers:    true,
		Activated: []ActivatedAbility{{
			Label:        "{T}, Exile a creature card from your graveyard: Discover X, where X is the mana value of the exiled card",
			Cost:         Plus(TapCost(), ExileFromGraveyard(1, "a creature card", MatchCreature)),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				x := 0
				for _, id := range ctx.Exiled() {
					if c, ok := g.LookupCardForEffect(id); ok {
						x, _ = g.ManaValueForEffect(c)
					}
				}
				return Discover{N: x}.Apply(ctx)
			},
		}},
	})
}
