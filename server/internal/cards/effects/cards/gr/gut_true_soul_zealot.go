package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gut, True-Soul Zealot — Legendary Creature — Goblin Shaman {2}{R}, 2/2 (Commander
// Legends: Battle for Baldur's Gate):
//
//	"Whenever you attack, you may sacrifice another creature or an artifact. If you
//	 do, create a 4/1 black Skeleton creature token with menace that's tapped and
//	 attacking. (It can't be blocked except by two or more creatures.)
//	 Choose a Background (You can have a Background as a second commander.)"
//
// "Whenever you attack" is once per declaration (OncePerBatch). The sacrifice is asked
// as the trigger resolves (MaySacrificeOne over another creature or an artifact, none
// meaning no question) and only a permanent that really went makes the Skeleton, made
// tapped and attacking by the controller's choice of what it attacks (CR 508.4), so it
// never "attacked" (CR 506.3c). Choose a Background is a deck-building rule the deck
// validator applies (deck/partner.go).
//
// No simplification.
func init() {
	const label = "Gut, True-Soul Zealot — you may sacrifice another creature or an artifact; if you do, create a 4/1 black Skeleton token with menace tapped and attacking"
	effects.Register(effects.Spec{
		OracleID:     "60e7130a-d321-4e50-93a4-999013c76054",
		Name:         "Gut, True Soul Zealot",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller)
			}, label, func(g *game.Game, item *game.StackItem) error {
				return effects.MaySacrificeOne(g, item, "Gut, True-Soul Zealot — sacrifice another creature or an artifact? If you do, create a 4/1 Skeleton tapped and attacking",
					effects.Or(effects.Creature(), effects.Artifact()), true, gutSacrificed)
			})),
		},
	})
}

// gutSacrificed makes the Skeleton once a permanent has really gone, keyed.
var gutSacrificed = effects.RegisterSacrificedOne("prompt-run/gap-b219-gut-true-soul-zealot-sacrifice", func(ctx *effects.Context, _ uuid.UUID) error {
	return effects.CreateTokensAttackingYourChoice{
		Controller: ctx.Controller(),
		Template:   effects.TokenCard("4/1 black Skeleton with menace"),
		N:          1,
		Tapped:     true,
		Label:      "Gut, True-Soul Zealot",
	}.Apply(ctx)
})
