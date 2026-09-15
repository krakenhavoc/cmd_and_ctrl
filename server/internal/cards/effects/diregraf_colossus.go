package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Diregraf Colossus — Creature — Zombie Giant {2}{B}, 2/2 (EDHREC
// rank 2130):
//
//	"This creature enters with a +1/+1 counter on it for each Zombie
//	 card in your graveyard.
//	 Whenever you cast a Zombie spell, create a tapped 2/2 black
//	 Zombie creature token."
//
// The Zombie deck's three-drop: as big as the graveyard, and every
// Zombie cast afterwards brings a friend. The counters are a CR 614
// self-entry replacement counted as the Colossus enters (so a
// Zombie milled in response is counted, and a counter doubler sees
// the whole batch); the trigger reads the spell off the stack by its
// printed subtype and makes the token tapped through the S21 token
// entry options.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fd62ad01-601f-4250-bc2a-8ef3982e45c4",
		Name:         "Diregraf Colossus",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b19EntersWithCountersCounted("+1/+1", func(g *game.Game, src *game.Card) int {
				return b19ZombieCardsInGraveyard(g, src.Controller)
			}, "Diregraf Colossus: enters with a +1/+1 counter per Zombie card in your graveyard"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b19ZombieSpellCastByYou(ev, source, g)
			}, "Diregraf Colossus — create a tapped 2/2 Zombie", func(g *game.Game, item *game.StackItem) error {
				return CreateTokenAdvanced{
					Controller: item.Controller,
					Spec:       Token(BlackZombieToken()).EntersTapped(),
					N:          1,
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
