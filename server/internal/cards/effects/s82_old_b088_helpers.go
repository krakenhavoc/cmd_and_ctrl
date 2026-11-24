package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// PowerAndToughnessDiffer is "whose power and toughness aren't equal"
// (Gilt-Leaf Winnower): the finished power, unclamped, against the finished
// toughness. A creature at -1/0 differs and a 0/0 does not.
func PowerAndToughnessDiffer() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.PowerForComparison() != c.CurrentToughness()
	}
}

// DefendingPlayerOfAttacker is the player `c` is attacking: the player itself,
// or the controller of the planeswalker or battle it attacks. uuid.Nil when it
// attacks nothing.
func DefendingPlayerOfAttacker(g *game.Game, c game.Card) uuid.UUID {
	target := c.AttackingTarget
	if target == uuid.Nil {
		return uuid.Nil
	}
	if g.PlayerByIDForEffect(target) != nil {
		return target
	}
	if t, ok := g.LookupCardForEffect(target); ok {
		return t.Controller
	}
	return uuid.Nil
}

// DefendingPlayerControlsAnArtifact is "as long as defending player controls
// an artifact" (Scrapdiver Serpent), read of the attacker as blockers are
// declared (a CantBeBlockedWhile condition): the player it attacks controls
// an artifact, by effective type.
func DefendingPlayerControlsAnArtifact(g *game.Game, _ uuid.UUID, attacker game.Card) bool {
	defender := DefendingPlayerOfAttacker(g, attacker)
	if defender == uuid.Nil {
		return false
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == defender && c.IsArtifact() {
			return true
		}
	}
	return false
}

// PlagueReaverReturns is the body of Plague Reaver's delayed trigger: "Return
// this creature to the battlefield under that player's control". The card
// returns only while it is still the object that was sacrificed (CR 400.7),
// still in a graveyard.
var PlagueReaverReturns = game.DelayedBody("s82-old-b088/plague-reaver-return",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		if p.Player == uuid.Nil || !ObjectStillInAGraveyard(g, p.Object) {
			return nil
		}
		return g.ReturnFromGraveyardUnderControlForEffect(p.Object.ID, game.ZoneBattlefield, p.Player)
	})
