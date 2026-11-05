package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jadzia Dax, Living Legacy — {1}{G} Legendary Creature — Trill
// Scientist, 0/1 (Star Trek, #4973):
//
//	"At the beginning of your first main phase, put a +1/+1 counter on
//	 Jadzia Dax, then add an amount of {G} equal to the number of
//	 counters on her.
//	 Transfer Symbiont — {2}{G}, Discard this card: Put two +1/+1
//	 counters on target creature."
//
// The trigger is Hulking Raptor's (EventBeginPrecombatMain, by you). It
// puts the counter on her while she is still the object that triggered
// (CounterOnThis), then counts every counter of every kind on her as the
// ability resolves; if she has left, the count is her last-known
// information (CR 608.2h). The mana lands in her controller's pool.
//
// Transfer Symbiont works from the hand (Armus, Skin of Evil's shape): the
// discard is part of the cost, and a target gone by resolution gets
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "67601646-f906-4824-9269-3cf89766d25f",
		Name:         "Jadzia Dax, Living Legacy",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourPrecombatMain("Jadzia Dax, Living Legacy — put a +1/+1 counter on her, then add {G} for each counter on her",
				jadziaDaxCounterThenMana),
		},
		Activated: []ActivatedAbility{{
			Label:   "Transfer Symbiont — {2}{G}, Discard this card: Put two +1/+1 counters on target creature.",
			Cost:    game.AbilityCost{Mana: "{2}{G}", DiscardSelf: true},
			Zones:   []game.ZoneKind{game.ZoneHand},
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return AddCounter{Target: id, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
			},
		}},
	})
}

func jadziaDaxCounterThenMana(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (CounterOnThis{Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	info, ok := ctx.SourcePermanent()
	if !ok {
		return nil
	}
	n := 0
	for _, k := range info.Counters {
		if k > 0 {
			n += k
		}
	}
	if n == 0 {
		return nil
	}
	return g.AddManaForEffect(item.Controller, item.SourceCardID, strings.Repeat("{G}", n))
}
