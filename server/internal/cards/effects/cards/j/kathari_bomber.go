package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kathari Bomber — Creature — Bird Shaman {1}{B}{R}, 2/2 (Duel Decks: Speed
// vs. Cunning, #5555):
//
//	"Flying
//	 When this creature deals combat damage to a player, create two 1/1 red
//	 Goblin creature tokens and sacrifice this creature.
//	 Unearth {3}{B}{R}"
//
// One trigger: the two Goblins first, then the sacrifice (a Bomber that has
// already left, or come back as a new object, is not sacrificed, CR 400.7).
// An unearthed Bomber that connects is sacrificed, which is exiled instead.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1ad39f85-799a-42ae-852c-9117d689fe65",
		Name:            "Kathari Bomber",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer(
				"Kathari Bomber — create two 1/1 red Goblin creature tokens and sacrifice this creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.CreateToken{Template: effects.TokenCard("1/1 red Goblin"), N: 2}).Apply(ctx); err != nil {
						return err
					}
					return effects.SacrificeThisIfStillOnBattlefield(g, item)
				}),
		},
		Activated: []effects.ActivatedAbility{effects.Unearth("{3}{B}{R}")},
	})
}
