package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Graf Harvest — Enchantment {B} (Jumpstart 2022):
//
//	"Zombies you control have menace.
//	 {3}{B}, Exile a creature card from your graveyard: Create a 2/2 black
//	 Zombie creature token."
//
// A layer 6 menace grant to the Zombies its controller controls (Piper of the
// Swarm's, read through the source's text) and Ashnod's graveyard-exile cost
// feeding a 2/2 black Zombie token. The exiled card is a cost (CR 118), so it
// is gone before the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d3ba6922-c2f7-45ab-87a3-d4bbd770d1ba",
		Name:         "Graf Harvest",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && target.HasSubtype(source.TextSubtype("Zombie"))
			}, "menace"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{B}, Exile a creature card from your graveyard: Create a 2/2 black Zombie creature token.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}{B}"), effects.ExileFromGraveyard(1, "a creature card", effects.IsCreatureCard)),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie"), N: 1}),
		}},
	})
}
