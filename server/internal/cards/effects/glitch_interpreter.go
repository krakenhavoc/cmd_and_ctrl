package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Glitch Interpreter — Creature — Human Wizard {2}{U}:
//
//	"When this creature enters, if you control no face-down permanents,
//	 return this creature to its owner's hand and manifest dread.
//	 Whenever one or more colorless creatures you control deal combat
//	 damage to a player, draw a card."
//
// The enter trigger's "if" is an intervening if (CR 603.4): checked as
// the creature enters and again as the trigger resolves, so a face-down
// permanent that appears in response keeps the Interpreter on the
// battlefield. Both halves of the instruction happen even when the
// Interpreter has left by then (the bounce simply finds nothing); the
// manifest waits behind the bounce, which can pause on a commander's
// command-zone question.
//
// The draw trigger is one per player damaged per damage step, however
// many colorless creatures connect (CR 603.2c). A face-down creature is
// colorless (CR 708.2), as is any artifact creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cf2c507a-41b4-457b-bbe8-edcdfcc681ac",
		Name:         "Glitch Interpreter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID && !controlsAFaceDownPermanent(g, source.Controller)
			}, "Glitch Interpreter — return this creature to its owner's hand and manifest dread",
				func(g *game.Game, item *game.StackItem) error {
					// CR 603.4: the condition is checked again on resolution.
					if controlsAFaceDownPermanent(g, item.Controller) {
						return nil
					}
					ctx := NewContext(g, item)
					return BounceToHand{Target: item.SourceCardID, Then: func(ctx *Context, _ bool) error {
						return ManifestDread{}.Apply(ctx)
					}}.Apply(ctx)
				}),
			OncePerBatchPerPlayer(On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !combatDamageToPlayerBy(ev, source.Controller, g) {
					return false
				}
				dealer, ok := g.LookupCardForEffect(ev.Source)
				return ok && dealer.IsColorless()
			}, "Glitch Interpreter — draw a card", Do(DrawCards{N: 1}))),
		},
	})
}

// controlsAFaceDownPermanent reports whether `player` controls any
// face-down permanent.
func controlsAFaceDownPermanent(g *game.Game, player uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.FaceDown {
			return true
		}
	}
	return false
}
