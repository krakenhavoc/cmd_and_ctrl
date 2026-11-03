package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nev, the Practical Dean — Legendary Creature — Merfolk Wizard {2}{G},
// 2/2 (Secrets of Strixhaven Commander, #4834):
//
//	"Creatures you control with counters on them have trample.
//	 Whenever you cast your first spell with {X} in its mana cost each
//	 turn, put X +1/+1 counters on Nev."
//
// The static is a layer-6 grant read off the permanent's counters, any
// kind of counter, so a creature gains trample the moment it has one and
// loses it when the last is removed. "Your first spell with {X} in its
// mana cost each turn" scans the whole turn's casts, so an X spell cast
// before Nev arrived uses the turn up (the ruling). X is the value
// announced for that spell (CR 107.3a), read as the trigger resolves off
// the spell, or off its last-known record if it was countered in
// response (CR 608.2h). X = 0 puts nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "42750e30-d022-4623-a290-d91fa80721c9",
		Name:         "Nev, the Practical Dean",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			Labelled(b16GrantKeywords(socB9CreatureYouControlWithACounter, "trample"), "Creatures you control with counters on them have trample."),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, socB9FirstXSpellThisTurn, "Nev, the Practical Dean — put X +1/+1 counters on Nev", nevPutXCounters),
		},
	})
}

// nevPutXCounters puts the triggering spell's X in +1/+1 counters on Nev.
func nevPutXCounters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	x, _ := g.SpellXValueForEffect(ctx.Trigger().Event.CardID)
	if x <= 0 {
		return nil
	}
	return CounterOnThis{Kind: game.CounterPlusOne, N: x}.Apply(ctx)
}
