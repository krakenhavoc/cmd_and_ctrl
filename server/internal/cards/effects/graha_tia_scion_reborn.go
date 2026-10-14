package effects

import (
	"errors"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// G'raha Tia, Scion Reborn — Legendary Creature — Cat Wizard {W}{U}{B},
// 2/3:
//
//	"Lifelink
//	 Throw Wide the Gates — Whenever you cast a noncreature spell, you
//	 may pay X life, where X is that spell's mana value. If you do,
//	 create a 1/1 colorless Hero creature token and put X +1/+1
//	 counters on it. Do this only once each turn."
//
// Lifelink is printed data. Throw Wide the Gates is Pantlaza,
// Sun-Favored's "you may … do this only once each turn" shape: the
// "may" is the trigger's optional prompt, and the ability is offered
// only while it has not resolved this turn, so declining leaves it
// free for the next noncreature spell. Two triggers put on the stack
// before either resolves are both offered; the second does nothing
// once the first has resolved.
//
// X is Monstrous Vortex's "that spell's mana value", read as the
// ability resolves, so an X the spell was cast for counts (CR 202.3e)
// and a spell countered in response is read where it went. Paying the
// life is a payment (CR 119.4): it needs X life or more, and "if you
// do" creates the token only once it is paid. X of 0 is a payment of
// nothing and still makes the 1/1. The counters arrive as the token
// enters.
//
// The simplification: "you may" is answered when the ability
// triggers rather than as it resolves, as every optional trigger in
// the catalog is. Answered yes, it uses up the once even when the
// life is no longer there to pay by the time it resolves.
func init() {
	Register(Spec{
		OracleID:        "f558c758-8f12-45f9-8683-eee33ba1fb0d",
		Name:            "G'raha Tia, Scion Reborn",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"If you can no longer pay the life when Throw Wide the Gates resolves, it still counts as your once this turn."},
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventCast, AllOf(YouCast(Noncreature()), pantlazaNotYetThisTurn(grahaTiaLabel)), grahaTiaLabel, grahaTiaThrowWideTheGates),
				"G'raha Tia — pay X life, where X is that spell's mana value, to create a 1/1 Hero with X +1/+1 counters?"),
		},
	})
}

const grahaTiaLabel = "G'raha Tia — Throw Wide the Gates: pay X life to create a Hero with X +1/+1 counters"

// grahaTiaThrowWideTheGates pays X life and, only if that was paid,
// makes the Hero with X counters.
func grahaTiaThrowWideTheGates(g *game.Game, item *game.StackItem) error {
	if g.ResolvedThisTurn(item.SourceCardID, grahaTiaLabel) > 1 {
		return nil
	}
	x := triggeringSpellManaValue(g, item)
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil || p.Eliminated || p.Life < x {
		return nil
	}
	if err := g.PayLifeForEffect(item.SourceCardID, item.Controller, x); err != nil {
		if errors.Is(err, game.ErrInvalidParam) {
			return nil // a payment that can't be made is not made (CR 119.8)
		}
		return err
	}
	opts := game.TokenEntryOptions{}
	if x > 0 {
		opts.Counters = map[string]int{game.CounterPlusOne: x}
	}
	_, err := g.CreateTokensForEffect(item.Controller, TokenCard("1/1 colorless Hero"), 1, opts)
	return err
}
