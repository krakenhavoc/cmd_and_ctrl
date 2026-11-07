package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pugnacious Hammerskull — Creature — Dinosaur {2}{G}, 6/6 (The Lost
// Caverns of Ixalan, #4470):
//
//	"Whenever this creature attacks while you don't control another
//	 Dinosaur, put a stun counter on it. (If a permanent with a stun
//	 counter would become untapped, remove one from it instead.)"
//
// "While" is a condition of the attack, not an intervening "if"
// (CR 603.4), Combadge's shape: it is read once, as the Hammerskull is
// declared an attacker, against the other permanents with the Dinosaur
// subtype its controller controls after the layers (a changeling counts).
// A Dinosaur that arrives before the trigger resolves does not stop it.
// The stun counter is the engine's (CR 122.1d), and goes on the
// Hammerskull only while it is still the same object. "Dinosaur" is a
// word of the card's text (CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "da989677-3647-46b8-aa56-b2cfb60f7a2a",
		Name:         "Pugnacious Hammerskull",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclared(ev, source) && hobA9OtherSubtypeYouControl(g, source, source.TextSubtype("Dinosaur")) == 0
			}, "Pugnacious Hammerskull — put a stun counter on it",
				func(g *game.Game, item *game.StackItem) error {
					if !sourceIsStillThisPermanent(g, item) {
						return nil
					}
					return AddCounter{Target: item.SourceCardID, Kind: game.CounterStun, N: 1}.Apply(NewContext(g, item))
				}),
		},
	})
}
