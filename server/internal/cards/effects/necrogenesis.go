package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Necrogenesis — Enchantment {B}{G} (Shards of Alara, #5588):
//
//	"{2}: Exile target creature card from a graveyard. Create a 1/1 green
//	 Saproling creature token."
//
// Cemetery Reaper's activation with no tap and a Saproling: the target
// is a creature card in any graveyard. A target gone in response makes
// the ability fizzle, so no Saproling is made (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ad08952a-bcae-41d6-8093-c18535dad7d5",
		Name:         "Necrogenesis",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}: Exile target creature card from a graveyard. Create a 1/1 green Saproling creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1, CreatureTokens: 1},
			Cost:    ManaCost("{2}"),
			Targets: targetCreatureInAnyGraveyard(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
				return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 green Saproling"), N: 1}.Apply(ctx)
			},
		}},
	})
}
