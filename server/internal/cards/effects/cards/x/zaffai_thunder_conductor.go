package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zaffai, Thunder Conductor — Legendary Creature — Human Shaman {2}{U}{R},
// 1/4 (Commander 2021, #5504):
//
//	"Magecraft — Whenever you cast or copy an instant or sorcery spell,
//	 scry 1. If that spell's mana value is 5 or greater, create a 4/4 blue
//	 and red Elemental creature token. If that spell's mana value is 10 or
//	 greater, Zaffai deals 10 damage to an opponent chosen at random."
//
// One magecraft trigger (CR 207.2c): it fires once per cast and once per
// copy. The spell's mana value is read as the trigger resolves, the
// announced X included while the spell is on the stack (CR 202.3e), and
// rides the scry's tail so the token and the damage come after the scry is
// answered ("scry 1. If ..."). The two thresholds are independent, so a
// mana value of 10 or more makes the Elemental and then the damage. The
// opponent is drawn through the keyed random stream among the opponents
// still in the game, and the damage is Zaffai's (last-known if it has left).
//
// No simplification.
func init() {
	trigger := effects.Magecraft("Zaffai, Thunder Conductor — scry 1; at mana value 5 or more a 4/4 Elemental; at 10 or more 10 damage to a random opponent",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			mv := effects.TriggeringSpellManaValue(g, item)
			return effects.Scry{N: 1, ThenTail: effects.WithN(ctx.ItemTail(zaffaiAfterScry), mv)}.Apply(ctx)
		})
	effects.Register(effects.Spec{
		OracleID:     "02d529df-6299-4e8f-9794-70c8061efdc4",
		Name:         "Zaffai, Thunder Conductor",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Triggered:    []game.TriggeredAbility{trigger},
	})
}

// zaffaiAfterScry is the rest of the ability once the scry is answered:
// the tail's N is the spell's mana value.
var zaffaiAfterScry = effects.CardTail("then/zaffai-after-scry", func(ctx *effects.Context, o game.TailOutcome) error {
	mv := o.Tail.N
	if mv >= 5 {
		if err := (effects.CreateToken{Template: effects.TokenCard("4/4 blue and red Elemental"), N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	if mv < 10 {
		return nil
	}
	opponents := ctx.Opponents()
	if len(opponents) == 0 {
		return nil
	}
	pick := ctx.Game.ChooseAtRandomForEffect(effects.RandomDraw(ctx), opponents, 1)
	if len(pick) == 0 {
		return nil
	}
	return effects.DealDamage{Source: ctx.Source(), Target: pick[0], Amount: 10}.Apply(ctx)
})
