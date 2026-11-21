package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Prize Pig — Creature — Boar {1}{G}, 0/3 (Tales of Middle-earth Commander,
// #5948):
//
//	"Whenever you gain life, put that many ribbon counters on this creature.
//	 Then if there are three or more ribbon counters on this creature, remove
//	 those counters and untap it.
//	 {T}: Add one mana of any color."
//
// Beckett Mariner's shape: the counters go on through the keyed placement
// (so a counter doubler's order prompt never splits it from the check), as
// many as the life gained, read from the event as the trigger was queued, and
// the "Then if" runs once they have landed. With three or more the Pig
// removes all its ribbon counters and untaps; the mana ability is the
// Birds' any-colour tap (the width is the printed one, not narrowed to the
// commander's identity), which is why the untap matters. A Pig that has left
// the battlefield, or come back as a new object (CR 400.7), gets no counters
// and untaps nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "14f55fc9-ec19-4b4a-b3de-d75898c19edf",
		Name:         "Prize Pig",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouGainLife("Prize Pig — put that many ribbon counters on this creature, then at three or more remove them and untap it",
				prizePigRibbons),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color",
		}},
	})
}

const prizePigCounter = "ribbon"

func prizePigRibbons(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Event.Amount <= 0 {
		return nil
	}
	if !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) {
		return nil
	}
	return g.AddCounterByKeyedForEffect(uuid.Nil, item.SourceCardID, prizePigCounter, item.Trigger.Event.Amount,
		effects.NewContext(g, item).ItemTail(prizePigCheckTail))
}

// prizePigCheckTail is the "Then if …" after the counters settle. An on-disk
// identity: never renamed, never reused.
var prizePigCheckTail = effects.CardTail("then/prize-pig-ribbon-check", func(ctx *effects.Context, _ game.TailOutcome) error {
	g, id := ctx.Game, ctx.Source()
	if !effects.OnBattlefield(g, id) || effects.SourceIsNewObject(g, ctx.Item) {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	n := c.Counters[prizePigCounter]
	if n < 3 {
		return nil
	}
	if err := g.AddCounterForEffect(id, prizePigCounter, -n); err != nil {
		return err
	}
	return effects.UntapTarget{Target: id}.Apply(ctx)
})
