package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rigo, Streetwise Mentor — Legendary Creature — Cat Citizen
// {G/W}{W}{W/U}, 2/2 (Streets of New Capenna):
//
//	"Rigo enters with a shield counter on it. (If it would be dealt
//	 damage or destroyed, remove a shield counter from it instead.)
//	 Whenever you attack a player or planeswalker with one or more
//	 creatures with power 1 or less, draw a card."
//
// One declaration is one trigger (OncePerBatch over the engine's
// one-event-per-attacker); an attacker counts when its power through
// the layers is 1 or less and the defender it was declared against is
// a player or a planeswalker, not a battle (CR 506.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1f42a809-bbfc-4225-9c31-f7d0fa8fa17a",
		Name:         "Rigo, Streetwise Mentor",
		Completeness: CompletenessFull,
		Replacements: snc6Entry("Rigo"),
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !attackDeclaredByYou(ev, source.Controller) || !snc6AttacksPlayerOrPlaneswalker(g, ev) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && c.CurrentPower() <= 1
			}, "Rigo, Streetwise Mentor — draw a card", Do(DrawCards{N: 1}))),
		},
	})
}
