package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Yuriko, the Tiger's Shadow — Legendary Creature — Human Ninja
// {1}{U}{B}, 1/3:
//
//	"Commander ninjutsu {U}{B} ({U}{B}, Return an unblocked attacker
//	 you control to hand: Put this card onto the battlefield from your
//	 hand or the command zone tapped and attacking.)
//	 Whenever a Ninja you control deals combat damage to a player,
//	 reveal the top card of your library and put that card into your
//	 hand. Each opponent loses life equal to that card's mana value."
//
// #1278's proof, and the only printing of commander ninjutsu
// (CR 702.49c) outside the joke sets. The keyword is
// CommanderNinjutsu("{U}{B}") — ninjutsu with the command zone as a
// second zone the ability functions from (ninjutsu.go). From the
// command zone she comes in without casting, so without the CR 903.8
// tax and without adding to it: the whole reason the deck plays her
// as a commander.
//
// The trigger is Dark Confidant's flip pointed the other way — the
// same reveal (not a draw: no EventDrawCard, so a draw payoff does not
// see it), the same mana value read before the move — with the life
// lost by EACH OPPONENT rather than by her controller, and a Ninja
// connecting rather than an upkeep as the condition. "Whenever a Ninja
// you control" is per creature (CR 603.2c) and reads the subtype
// post-layer off the damage source, the condition Ingenious
// Infiltrator already ships on, so Yuriko connecting after her own
// ninjutsu flips a card for herself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a7043fbd-1dfd-42cf-be4b-cc343d0949e5",
		Name:         "Yuriko, the Tiger's Shadow",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{CommanderNinjutsu("{U}{B}")},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b43CreatureOfSubtypeYouControlDealtCombatDamageToPlayer(ev, source, g, "Ninja")
			}, "Yuriko, the Tiger's Shadow — reveal the top card of your library", yurikoFlip),
		},
	})
}

// yurikoFlip reveals the top card of the controller's library, puts it
// into their hand and drains each opponent for its mana value.
func yurikoFlip(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var revealed []uuid.UUID
	if err := (RevealTopOfLibrary{
		Player:   item.Controller,
		N:        1,
		Reason:   "Yuriko, the Tiger's Shadow — reveal the top card of your library",
		Revealed: &revealed,
	}).Apply(ctx); err != nil {
		return err
	}
	if len(revealed) == 0 {
		// An empty library reveals nothing: nothing to put into hand
		// and no mana value to lose.
		return nil
	}
	flipped := revealed[0]
	// Read before the move, for Dark Confidant's reason: the last
	// moment the card is certainly findable where the effect found it.
	// CR 202.3e makes the number the same in either zone.
	loss := 0
	if c, ok := g.LookupCardForEffect(flipped); ok {
		loss = c.ManaValue()
	}
	if err := (BounceToHand{Target: flipped}).Apply(ctx); err != nil {
		return err
	}
	if loss == 0 {
		return nil
	}
	for _, opp := range ctx.Opponents() {
		if err := g.ChangePlayerLifeForEffect(ctx.Source(), opp, -loss); err != nil {
			return err
		}
	}
	return nil
}
