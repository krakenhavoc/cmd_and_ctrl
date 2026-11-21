package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spider-Man, Miles Morales — Legendary Creature — Spider Human Hero
// {4}{G}{G}, 5/5 (Marvel's Spider-Man, #4460):
//
//	"Vigilance, trample
//	 Whenever Spider-Man enters or attacks, put a +1/+1 counter on each
//	 other creature you control. Those creatures gain trample until end
//	 of turn."
//
// One printed trigger, one registered ability on enters-or-attacks. The
// set is the other creatures the controller controls as the trigger
// resolves, taken once (CR 611.2c): "those creatures" are exactly the ones
// that get a counter. Trample is granted to each before the counters land
// (neither depends on the other, and a counter placement can pause on a
// CR 616 ordering prompt).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "229ecaa1-b3c9-44a1-8e7a-857ea5b8b0df",
		Name:            "Spider-Man, Miles Morales",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrAttacks("Spider-Man, Miles Morales — a +1/+1 counter on each other creature you control; they gain trample until end of turn",
				milesMoralesPumpTeam),
		},
	})
}

func milesMoralesPumpTeam(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() && c.InstanceID != item.SourceCardID {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Spider-Man, Miles Morales — trample"}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return effects.B13PutCounterOnEach(ctx.AsGroupMember(), ids)
}
