package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grave Titan — Creature — Giant {4}{B}{B}, 6/6:
//
//	"Deathtouch
//	 Whenever this creature enters or attacks, create two 2/2 black
//	 Zombie creature tokens."
//
// Ten power across three bodies for six mana, and four more power
// every time it attacks. The best black fatty ever printed and an
// aristocrats deck's single biggest source of fuel.
//
// "Enters or attacks" is ONE ability with two trigger conditions, so
// it is one TriggeredAbility watching EventETB and EventAttack —
// Sun Titan's shape, and the reason both event kinds carry the
// permanent in CardID. A Titan that enters with haste and attacks
// the same turn makes four Zombies, in two separate batches, which
// is paper behaviour.
//
// Issue #73 recorded "there is no attack event in events.go at all"
// as a blocker for exactly this card. That note is stale — EventAttack
// landed in S22 (game/events.go) and Hellrider, Sun Titan and Krenko,
// Tin Street Kingpin already consume it.
func init() {
	Register(Spec{
		OracleID:        "f3abd4d1-a975-4e85-8684-aa0fce029670",
		Name:            "Grave Titan",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB, game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Grave Titan — create two 2/2 black Zombies",
					func(g *game.Game, item *game.StackItem) error {
						return CreateToken{
							Controller: item.Controller,
							Template:   BlackZombieToken(),
							N:          2,
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
