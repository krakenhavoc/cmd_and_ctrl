package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tolabow, Loch Rascal — Legendary Creature — Otter {2}{U}{U}, 3/2
// (Mystery Booster Commander, #5948):
//
//	"Rulebreaker — If Tolabow, Loch Rascal is your commander, the color
//	 identity of instant and sorcery cards in your deck can include one color
//	 of your choice not in your commander's color identity, and your deck can
//	 have any basic land cards.
//	 Whenever you cast an instant or sorcery spell, create a 1/1 blue and red
//	 Otter creature token with prowess."
//
// The trigger is Aziza's cast watch over instants and sorceries, making the
// Otter of Stormchaser's Talent (1/1 blue and red, prowess). The Rulebreaker
// sentence is deck construction, which is no part of play and is read by the
// deck validator (internal/deck), not here.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "56d71eda-4037-45fb-9e98-1725147af6fc",
		Name:         "Tolabow, Loch Rascal",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Or(Instant(), Sorcery()),
				"Tolabow, Loch Rascal — create a 1/1 blue and red Otter creature token with prowess",
				Do(CreateToken{Template: TokenCard("1/1 blue and red Otter with prowess"), N: 1})),
		},
	})
}
