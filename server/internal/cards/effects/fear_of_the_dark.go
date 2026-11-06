package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fear of the Dark — Enchantment Creature — Nightmare {4}{B}, 5/5
// (Duskmourn: House of Horror, #4466):
//
//	"Whenever this creature attacks, if defending player controls no
//	 Glimmer creatures, it gains menace and deathtouch until end of turn.
//	 (A creature with menace can't be blocked except by two or more
//	 creatures.)"
//
// An intervening "if" (CR 603.4), Robber of the Rich's shape: the
// defending player (the player attacked, or the controller of the
// planeswalker or battle attacked) must control no Glimmer creature as the
// attack is declared and again as the trigger resolves, so a Glimmer made
// in response stops it. Glimmer is read through the layers on creatures
// only (b43CreaturesOfSubtypeControlled), in the card's own text (CR
// 612.2). The keywords are the eerie creatures' self grant
// (thisCreatureUntilEOT), nothing if it has left.
//
// No simplification.
func init() {
	const label = "Fear of the Dark — gains menace and deathtouch until end of turn"
	gains := thisCreatureUntilEOT(label, 0, 0, "menace", "deathtouch")
	Register(Spec{
		OracleID:     "9285cdf1-ad25-4aee-a865-1bf0cc905183",
		Name:         "Fear of the Dark",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclared(ev, source) &&
					!controlsAGlimmerCreature(g, defendingPlayerOfAttack(g, ev), source.TextSubtype("Glimmer"))
			}, label, func(g *game.Game, item *game.StackItem) error {
				if controlsAGlimmerCreature(g, defendingPlayerOfAttack(g, item.Trigger.Event), NewContext(g, item).TextSubtype("Glimmer")) {
					return nil
				}
				return gains(g, item)
			}),
		},
	})
}

// controlsAGlimmerCreature reports whether `player` controls a creature
// of the given type (Glimmer, as the card's text reads it).
func controlsAGlimmerCreature(g *game.Game, player uuid.UUID, glimmer string) bool {
	return player != uuid.Nil && b43CreaturesOfSubtypeControlled(g, player, glimmer) > 0
}
