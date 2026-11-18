package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dying to Serve — Enchantment {2}{B} (Innistrad: Crimson Vow):
//
//	"Whenever you discard one or more cards, create a tapped 2/2 black Zombie
//	 creature token. This ability triggers only once each turn."
//
// "One or more" is OncePerBatch (the engine emits one discard event per
// card, so a Windfall is one trigger); "only once each turn" is Morbid
// Opportunist's tally of the ability already put on the stack this turn,
// which a countered trigger still spends.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2b719057-1bc0-4c3a-97c3-8377844570dd",
		Name:         "Dying to Serve",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(OncePerBatch(On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return discardedByYou(ev, source) && !b11TriggeredThisTurn(g, source.InstanceID, dyingToServeLabel)
			}, dyingToServeLabel, func(g *game.Game, item *game.StackItem) error {
				tmpl := BlackZombieToken()
				tmpl.Tapped = true
				return CreateToken{Controller: item.Controller, Template: tmpl, N: 1}.Apply(NewContext(g, item))
			})), game.Purpose{Tokens: 1, CreatureTokens: 1}),
		},
	})
}

const dyingToServeLabel = "Dying to Serve — create a tapped 2/2 black Zombie creature token"
