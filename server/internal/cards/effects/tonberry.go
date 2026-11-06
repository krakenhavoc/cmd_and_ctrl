package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tonberry — Creature — Salamander Horror {B}, 2/1 (Final Fantasy,
// #4462):
//
//	"This creature enters tapped with a stun counter on it. (If it would
//	 become untapped, remove a stun counter from it instead.)
//	 Chef's Knife — During your turn, this creature has first strike and
//	 deathtouch."
//
// The entry is one self replacement that taps it and puts the stun
// counter on (CR 614.1d), so its first untap step only removes the
// counter. "During your turn" is Triumphant Adventurer's layer 6 grant,
// read on whose turn it is. "Chef's Knife" is a flavor word with no rules
// meaning.
//
// No simplification.
func init() {
	yourTurn := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return thisPermanent(target, g, source) && isActivePlayer(g, source.Controller)
	}
	Register(Spec{
		OracleID:     "a26f3cbb-39b5-4688-92f7-26534613a9f3",
		Name:         "Tonberry",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTappedWithCounters(game.CounterStun, 1)},
		Static: []game.StaticAbility{
			KeywordGrant(yourTurn, "first strike"),
			KeywordGrant(yourTurn, "deathtouch"),
		},
	})
}
