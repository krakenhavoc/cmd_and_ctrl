package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Return of the Mole Man — {1}{G} Enchantment (Marvel Super Heroes
// Commander, #4455):
//
//	"Landfall — Whenever a land you control enters, you may mill two cards.
//	 {5}{G}, Sacrifice this enchantment: Create X 1/1 green Minion creature
//	 tokens named Moloid, where X is the number of permanent cards in your
//	 graveyard. The tokens have "Whenever this token attacks, you may mill a
//	 card." Activate only as a sorcery."
//
// The Moloid is Mole Man, Moloid Master's catalog token (mshB6MoloidToken).
// X is counted as the ability resolves, so the sacrificed enchantment, a
// permanent card now in the graveyard, counts itself; tokens never reach a
// graveyard, so only cards are counted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a38f7ce1-6ccc-4fd3-9e02-631e8933c1f8",
		Name:         "Return of the Mole Man",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Landfall("Return of the Mole Man — mill two cards", effects.Do(effects.MillCards{N: 2})),
				"Return of the Mole Man — mill two cards?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{5}{G}, Sacrifice this enchantment: Create X 1/1 green Minion creature tokens named Moloid, where X is the number of permanent cards in your graveyard. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{5}{G}"), effects.SacrificeThis()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				n := 0
				if p := g.PlayerByIDForEffect(item.Controller); p != nil && p.Graveyard != nil {
					for _, c := range p.Graveyard.Cards {
						if c.IsPermanent() {
							n++
						}
					}
				}
				return effects.CreateToken{Template: effects.MshB6MoloidToken(), N: n}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
