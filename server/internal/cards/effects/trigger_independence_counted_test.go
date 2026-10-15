package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trigger_independence_counted_test.go — #3033: the last three cards on
// the #3013 ratchet. Arboreal Alliance's X comes from the engine's cast
// record, Cadira counts the board, and Ikra Shidiqi's subject is the
// creature that dealt the damage.

const cadiraCountedOracle = "fcf22321-2f55-41ca-b8e1-4792540ba3ee"

func TestTheCountedAndXRowsDeclareFootprints(t *testing.T) {
	kinds := func(oracle string, row int) []game.FootprintStep {
		rows := game.CatalogTriggers(oracle)
		if len(rows) <= row {
			t.Fatalf("%s: %d rows", oracle, len(rows))
		}
		return rows[row].Footprint
	}
	if fp := kinds(arborealAllianceOracle, 0); len(fp) != 1 || fp[0].Kind != game.FootprintCreateToken {
		t.Errorf("Arboreal Alliance: %+v, want one token step", fp)
	}
	if rows := game.CatalogTriggers(arborealAllianceOracle); !rows[0].XFromCast || rows[0].Build != nil {
		t.Error("Arboreal Alliance: the engine stamps X, with no Build")
	}
	if fp := kinds(cadiraCountedOracle, 0); len(fp) != 2 || fp[0].Cond != game.FootprintCondBoard || fp[1].Kind != game.FootprintCreateToken {
		t.Errorf("Cadira: %+v, want a board read and a token step", fp)
	}
	if fp := kinds(b35IkraShidiqiOracle, 0); len(fp) != 1 || fp[0].Kind != game.FootprintGainLife || !fp[0].AmountFromSubject {
		t.Errorf("Ikra Shidiqi: %+v, want a gain from the subject", fp)
	}
	if rows := game.CatalogTriggers(b35IkraShidiqiOracle); !rows[0].SubjectIsEventSource || rows[0].Build != nil {
		t.Error("Ikra Shidiqi: the subject is the event's source, with no Build")
	}
}

// Ikra's subject is the creature that connected, and the gain is its
// toughness as it last existed once it has left (CR 608.2h), where the
// old body used the toughness it had when the damage was dealt.
func TestIkraReadsTheConnectingCreatureAsItsSubject(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ikra := b12Push(g, me.ID, "Ikra Shidiqi, the Usurper", "Legendary Creature — Snake Wizard", b35IkraShidiqiOracle, 3, 7)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 3)
	declareAttack(t, g, opp.ID, bear)
	advanceTo(t, g, game.StepCombatDamage)
	item := triggerOnStack(g, ikra)
	if item == nil {
		t.Fatal("no Ikra trigger on the stack after the Bear connected")
	}
	if item.Trigger == nil || item.Trigger.Object == nil || item.Trigger.Object.ID != bear {
		t.Fatalf("the trigger's subject is %+v, want the Bear", item.Trigger)
	}
	life := me.Life
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("the Bear died in response: life %d → %d, want its last-known toughness 3", life, me.Life)
	}
}

// Cadira counts the tokens as her trigger resolves.
func TestCreateTokensForEachCountsAsItResolves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	src := pushBattlefieldCreature(g, me.ID)
	var source game.Card
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == src {
			source = c
		}
	}
	item := game.NewTriggeredItem(&source, "Count probe")
	step := CreateTokensForEach{Template: TokenCard("1/1 white Rabbit"), Name: "token you control",
		Count: func(ctx *Context) int { return b36TokensControlled(ctx.Game, ctx.Controller()) }}
	g.WithWriteLock(func() { _ = step.Apply(NewContext(g, item)) })
	if n := b36TokensControlled(g, me.ID); n != 0 {
		t.Fatalf("no tokens to count made %d", n)
	}
	g.WithWriteLock(func() {
		_ = CreateToken{Template: TokenCard("1/1 white Rabbit"), N: 2}.Apply(NewContext(g, item))
		_ = step.Apply(NewContext(g, item))
	})
	if n := b36TokensControlled(g, me.ID); n != 4 {
		t.Errorf("two tokens made two more: %d, want 4", n)
	}
}
