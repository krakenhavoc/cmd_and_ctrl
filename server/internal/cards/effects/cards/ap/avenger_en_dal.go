package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Avenger en-Dal — Creature — Human Spellshaper {1}{W}, 1/1 (Nemesis, S82 OLD
// batch 679, #5618):
//
//	"{2}{W}, {T}, Discard a card: Exile target attacking creature. Its
//	 controller gains life equal to its toughness."
//
// Swords to Plowshares' shape with toughness. The target is an attacking
// creature, anyone's (it is re-checked as the ability resolves, CR 608.2b).
// The toughness and the controller are read before the exile, as the
// creature last existed on the battlefield (CR 608.2h), and the life is
// gained only if it really was exiled and the toughness is above zero.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "62009b63-ca6c-4fb0-8c9a-eccdec51cb18",
		Name:         "Avenger en-Dal",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}, {T}, Discard a card: Exile target attacking creature. Its controller gains life equal to its toughness.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{2}{W}"), effects.TapCost(), effects.DiscardACard()),
			Targets: effects.TargetCreature("target attacking creature", effects.IsAttackingCreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				victim, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				card, found := g.LookupCardForEffect(victim)
				if !found {
					return nil
				}
				return effects.ExileTarget{Target: victim, ThenTail: game.KeyedTail{
					Ref: avengerEnDalGainsToughness, Player: card.Controller, Source: item.SourceCardID, N: card.CurrentToughness(),
				}}.Apply(ctx)
			},
		}},
	})
}

// avengerEnDalGainsToughness is the rest of the ability once the exile has
// finished: the creature's controller, as it was on the battlefield, gains
// life equal to its toughness as it last existed (CR 608.2h, both carried on
// the tail), only if the creature really reached exile. An on-disk key: never
// renamed, never reused.
var avengerEnDalGainsToughness = effects.CardTail("tail/avenger-en-dal-gains-toughness", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || o.Tail.N <= 0 {
		return nil
	}
	return effects.GainLife{Player: o.Tail.Player, Amount: o.Tail.N}.Apply(ctx)
})
