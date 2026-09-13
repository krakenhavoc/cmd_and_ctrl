package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hellrider — 3/3 Creature — Devil for {2}{R}{R}:
//
//	"Haste
//	Whenever a creature you control attacks, this creature deals 1
//	damage to the player or planeswalker it's attacking."
//
// S22 attack triggers: the card that reads the event's payload.
// EventAttack fires once per attacking creature (like EventDrawCard
// per card), so a five-creature alpha strike stacks five separate
// Hellrider triggers — paper behaviour, and the reason the pings
// resolve one at a time with a response window between them.
//
// The damaged player is the DEFENDER that particular attacker was
// declared against, read off ev.Target and captured in Build — not
// "an opponent of Hellrider's controller". In a four-player game two
// creatures can attack two different seats in the same combat, and
// each ping has to follow its own attacker.
//
// Hellrider counts itself: the text is "a creature you control", not
// "another".
//
// Sandbox simplification: "the player or planeswalker" is always the
// player. DeclareAttacker takes a player ID and validates it against
// the seats — the engine has no attack-a-planeswalker path — so
// there is no planeswalker case being dropped, only one that cannot
// arise yet.
func init() {
	Register(Spec{
		OracleID:        "f02557b0-f422-48cb-875e-2c814c39967d",
		Name:            "Hellrider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				// Captured now, not read off the board at resolution:
				// the attacker can be removed from combat in response
				// and the trigger still deals its damage (CR 603.7 —
				// the ability is independent of its source).
				defender := ev.Target
				if defender == uuid.Nil {
					return nil
				}
				return game.NewTriggeredItem(source, "Hellrider — 1 damage to the defending player",
					func(g *game.Game, item *game.StackItem) error {
						return DealDamage{
							Source: item.SourceCardID,
							Target: defender,
							Amount: 1,
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
