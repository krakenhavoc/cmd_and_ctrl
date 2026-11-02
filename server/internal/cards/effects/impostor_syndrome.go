package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Impostor Syndrome — Enchantment {4}{U}{U} (Marvel's Spider-Man, #4460):
//
//	"Whenever a nontoken creature you control deals combat damage to a
//	 player, create a token that's a copy of it, except it isn't
//	 legendary."
//
// One trigger per creature that connects (CR 603.2c). The creature is
// read as the damage is dealt: it must be a nontoken creature its
// controller controlled, and the copy is taken from wherever the card sits
// when the trigger resolves (a creature that died in response is copied
// from its graveyard card). The exception edits the copy's type line, as
// Helm of the Host's does; the token then runs the ordinary entry pipeline,
// so its enters triggers fire.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7be47440-b627-411e-a28f-d5a92e5e2fb7",
		Name:         "Impostor Syndrome",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, impostorSyndromeNontokenCreatureConnected,
				"Impostor Syndrome — create a token that's a copy of it, except it isn't legendary",
				impostorSyndromeCopy),
		},
	})
}

// impostorSyndromeNontokenCreatureConnected is the trigger condition:
// combat damage to a player from a nontoken creature the enchantment's
// controller controls.
func impostorSyndromeNontokenCreatureConnected(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !combatDamageToPlayerBy(ev, source.Controller, g) {
		return false
	}
	src, ok := g.LookupCardForEffect(ev.Source)
	return ok && !src.IsToken()
}

// impostorSyndromeCopy makes the nonlegendary token copy of the creature
// that dealt the damage.
func impostorSyndromeCopy(g *game.Game, item *game.StackItem) error {
	return CreateTokenCopy{
		Controller: item.Controller,
		Copy:       item.Trigger.Event.Source,
		N:          1,
		Except: func(t *game.Card) {
			t.TypeLine = removeLegendaryFromTypeLine(t.TypeLine)
		},
	}.Apply(NewContext(g, item))
}
