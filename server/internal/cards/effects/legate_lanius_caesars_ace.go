package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Legate Lanius, Caesar's Ace — Legendary Creature — Human Soldier
// {2}{B}{R}, 2/2 (Fallout, #5948):
//
//	"Decimate — When Legate Lanius enters, each opponent sacrifices a
//	 tenth of the creatures they control of their choice, rounded up.
//	 Whenever an opponent sacrifices a creature, put a +1/+1 counter on
//	 Legate Lanius."
//
// The count is worked out once, per opponent, as the trigger resolves:
// the creatures they control then, divided by ten and rounded up (so any
// opponent with a creature loses at least one), each a creature of their
// choice (Rush of Dread's sacrifice seam, PlayersSacrificeForEffect). The
// counter trigger reads the sacrifice event (CR 701.21): an opponent's
// sacrifice of a creature, however it came about, token included, one
// counter per creature, Lanius's own edict among them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c6b5ca62-e5ee-4636-b12e-7c1d455e8384",
		Name:         "Legate Lanius, Caesar's Ace",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Legate Lanius — each opponent sacrifices a tenth of the creatures they control, rounded up",
				legateLaniusDecimate),
			On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor == source.Controller || ev.Actor == uuid.Nil {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature()
			}, "Legate Lanius — put a +1/+1 counter on this creature", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}

// legateLaniusDecimate asks each opponent, in turn order, for a tenth of
// their creatures (rounded up), one prompt per creature owed.
func legateLaniusDecimate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var asks []uuid.UUID
	for _, opp := range ctx.Opponents() {
		owed := (len(permanentsControlledByMatching(g, opp, Creature())) + 9) / 10
		for i := 0; i < owed; i++ {
			asks = append(asks, opp)
		}
	}
	if len(asks) == 0 {
		return nil
	}
	g.PlayersSacrificeForEffect(ctx.Source(), asks, sacrificeSpec("a creature", Creature()),
		"Legate Lanius — sacrifice a tenth of the creatures you control, rounded up")
	return nil
}
