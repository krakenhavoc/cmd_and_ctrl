package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Guardian Project — Enchantment for {3}{G}:
//
//	"Whenever a nontoken creature you control enters, if it doesn't
//	 have the same name as another creature you control or a creature
//	 card in your graveyard, draw a card."
//
// Green's Beast Whisperer with a singleton clause bolted on — which
// costs nothing in Commander, where the deck is singleton by rule and
// the only repeats are tokens (already excluded) and a recurring
// creature you have already drawn off once.
//
// Three filters, all of them ordinary reads:
//
//   - NONTOKEN: IsToken reads the "Token" supertype the token
//     templates stamp. Excluding tokens is what stops an Avenger of
//     Zendikar from drawing a dozen cards.
//   - YOU CONTROL: the entering creature's controller, not its owner.
//   - THE NAME CHECK: no other creature YOU CONTROL and no creature
//     card in YOUR graveyard shares the name. "Another" excludes the
//     entering creature itself, which is already on the battlefield
//     when EventETB fires — without that exclusion the trigger would
//     never fire at all, which is the failure mode this comment
//     exists to stop the next reader from reintroducing.
//
// "Your graveyard" is the CONTROLLER's graveyard, and a card in a
// graveyard is always in its owner's, so the walk is over the
// controller's graveyard zone directly.
//
// DECLARED SIMPLIFICATION — THE INTERVENING-IF IS CHECKED ONCE. CR
// 603.4: an intervening-if clause is checked when the ability would
// trigger AND again as it resolves, and the ability does nothing if
// the condition is false either time. This checks it at trigger time
// only (in AppliesTo). The difference shows up when a second copy of
// the same name arrives, or a same-named creature card is milled
// into your graveyard, in RESPONSE to the trigger: printed, the
// trigger is removed from the stack and draws nothing; here it still
// draws. That is a card marginally STRONGER than printed in a narrow
// window, and it is the one thing on this card worth watching. It is
// declared rather than engineered around because Spec has no
// resolution-time condition hook — a re-check would have to live
// inside the Build closure's effect, and the effect has no access to
// the entering card's identity once the event is gone. The clean fix
// is a `Condition` field on game.TriggeredAbility consulted at both
// points.
func init() {
	Register(Spec{
		OracleID: "4f9e07ae-6341-4b46-9f77-f17ab659d266",
		Name:     "Guardian Project",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				entering, ok := g.LookupCardForEffect(ev.CardID)
				if !ok || !entering.IsCreature() || IsToken(entering) {
					return false
				}
				if entering.Controller != source.Controller {
					return false
				}
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.InstanceID == entering.InstanceID {
						continue
					}
					if c.Controller == source.Controller && c.IsCreature() && c.Name == entering.Name {
						return false
					}
				}
				if p := g.PlayerByIDForEffect(source.Controller); p != nil && p.Graveyard != nil {
					for _, c := range p.Graveyard.Cards {
						if c.IsCreature() && c.Name == entering.Name {
							return false
						}
					}
				}
				return true
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Guardian Project — draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
