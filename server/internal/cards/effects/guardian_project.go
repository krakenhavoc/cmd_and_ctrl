package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Guardian Project — Enchantment {3}{G} (EDHREC rank 321):
//
//	"Whenever a nontoken creature you control enters, if it doesn't
//	 have the same name as another creature you control or a creature
//	 card in your graveyard, draw a card."
//
// The singleton format's Glimpse of Nature: in Commander every
// nontoken creature is a fresh name, so it reads "whenever a
// nontoken creature you control enters, draw a card" — the clause
// exists to stop token and clone decks abusing it.
//
// The name check is an intervening-if (CR 603.4), evaluated when
// the trigger would go on the stack: the entering creature is
// compared against every OTHER creature its controller controls
// (itself excluded by instance ID) and every creature card in that
// player's graveyard. Same posture as every intervening-if card in
// the catalog — checked at trigger time, not re-checked on
// resolution — which cannot make this card stronger: a duplicate
// that appears in response would only have made the printed trigger
// do nothing.
//
// No simplification beyond that shared intervening-if posture.
func init() {
	Register(Spec{
		OracleID:     "4f9e07ae-6341-4b46-9f77-f17ab659d266",
		Name:         "Guardian Project",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				if !ok || !c.IsCreature() || IsToken(c) {
					return false
				}
				for _, other := range g.BattlefieldCardsForEffect() {
					if other.InstanceID != c.InstanceID && other.Controller == source.Controller &&
						other.IsCreature() && other.Name == c.Name {
						return false
					}
				}
				p := g.PlayerByIDForEffect(source.Controller)
				if p == nil || p.Graveyard == nil {
					return true
				}
				for _, gc := range p.Graveyard.Cards {
					if gc.IsCreature() && gc.Name == c.Name {
						return false
					}
				}
				return true
			}, "Guardian Project — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
