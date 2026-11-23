package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ovika, Enigma Goliath — Legendary Creature — Phyrexian Nightmare
// {5}{U}{R}, 6/6 (Foundations, #4465):
//
//	"Flying
//	 Ward—{3}, Pay 3 life.
//	 Whenever you cast a noncreature spell, create X 1/1 red Phyrexian
//	 Goblin creature tokens, where X is the mana value of that spell. They
//	 gain haste until end of turn."
//
// The ward is the mana-and-life cost (WardManaAndLife, Captain Howler's).
// X is the spell's mana value read as the trigger resolves, with its
// announced X while it is still on the stack (the ruling,
// triggeringSpellManaValue). The Goblins are made together, and the haste
// is a CR 611.2c "until end of turn" grant to just those tokens, given
// once they have landed (CardTail, Artistic Process's).
//
// No simplification.
func init() {
	const label = "Ovika, Enigma Goliath — create X 1/1 red Phyrexian Goblin tokens, where X is that spell's mana value; they gain haste until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "b05a2b5f-1b5c-4f1b-8a6b-525c5c8b8ba1",
		Name:            "Ovika, Enigma Goliath",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardManaAndLife("{3}", 3), "Ovika, Enigma Goliath — ward {3}, pay 3 life"),
			effects.WheneverYouCast(effects.Noncreature(), label, func(g *game.Game, item *game.StackItem) error {
				n := effects.TriggeringSpellManaValue(g, item)
				if n <= 0 {
					return nil
				}
				ctx := effects.NewContext(g, item)
				return g.CreateTokensKeyedForEffect(game.TokenCreation{
					Controller: item.Controller,
					Source:     item.SourceCardID,
					Groups:     []game.TokenGroup{{Template: ctx.TextToken(effects.TokenCard("1/1 red Phyrexian Goblin")), Count: n}},
				}, ctx.ItemTail(ovikaGoblinsHaste))
			}),
		},
	})
}

// ovikaGoblinsHaste is "they gain haste until end of turn" over the
// Goblins that landed. An on-disk key: never renamed, never reused.
var ovikaGoblinsHaste = effects.CardTail("ovika-enigma-goliath/goblins-haste", func(ctx *effects.Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if !effects.OnBattlefield(ctx.Game, id) {
			continue
		}
		if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Ovika, Enigma Goliath — haste"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
