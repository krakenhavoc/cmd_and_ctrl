package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "a38f7ce1-6ccc-4fd3-9e02-631e8933c1f8",
		Name:         "Return of the Mole Man",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(Landfall("Return of the Mole Man — mill two cards", Do(MillCards{N: 2})),
				"Return of the Mole Man — mill two cards?"),
		},
		Activated: []ActivatedAbility{{
			Label:        "{5}{G}, Sacrifice this enchantment: Create X 1/1 green Minion creature tokens named Moloid, where X is the number of permanent cards in your graveyard. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{5}{G}"), SacrificeThis()),
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
				return CreateToken{Template: mshB6MoloidToken(), N: n}.Apply(NewContext(g, item))
			},
		}},
	})
}
