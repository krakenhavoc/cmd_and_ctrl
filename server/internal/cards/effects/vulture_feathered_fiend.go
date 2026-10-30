package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vulture, Feathered Fiend — Legendary Creature — Human Artificer Villain
// {2}{U}{U}, 2/4 (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 Whenever one or more creatures you control with flying deal combat
//	 damage to a player, put a +1/+1 counter on each of those creatures
//	 and draw a card."
//
// Mu Yanling's batch trigger (one per player hit in a combat damage step,
// CR 603.2c, flying read off each creature as it deals its damage). "Each
// of those creatures" is the roster of that batch: the engine emits one
// damage event per creature, so the effect reads the event log for the
// flying creatures of the controller that dealt combat damage to the same
// player in the same batch (CountersYouPutInTheSameBatch's scan). A
// creature that has left the battlefield by then gets no counter; the draw
// is not conditional on it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c598d97f-fec6-4238-8eed-8051e9541d31",
		Name:            "Vulture, Feathered Fiend",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(HasKeyword("flying"),
				"Vulture, Feathered Fiend — put a +1/+1 counter on each of those creatures and draw a card",
				vultureCountersAndDraw),
		},
	})
}

func vultureCountersAndDraw(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := NewContext(g, item)
	ev := item.Trigger.Event
	seen := map[uuid.UUID]bool{}
	for i := len(g.Events) - 1; i >= 0; i-- {
		e := g.Events[i]
		if e.Batch < ev.Batch {
			break
		}
		if e.Batch != ev.Batch || e.Target != ev.Target || seen[e.Source] || !combatDamageToPlayerBy(e, item.Controller, g) {
			continue
		}
		c, ok := g.LookupCardForEffect(e.Source)
		if !ok || !HasKeyword("flying")(g, item.Controller, c) {
			continue
		}
		seen[e.Source] = true
		if !onBattlefield(g, e.Source) {
			continue
		}
		if err := (AddCounter{Target: e.Source, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
}
