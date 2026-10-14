package protocol

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestCompoundWardViewCarriesMandatoryLifeAndAffordability(t *testing.T) {
	g := buildActiveGame(t)
	payer := g.Seats[0]
	spell := uuid.New()
	g.Stack.PushTop(game.Card{InstanceID: spell, Name: "Spell", TypeLine: "Instant", Owner: payer.ID, Controller: payer.ID})
	if g.StackMeta == nil {
		g.StackMeta = make(map[uuid.UUID]*game.StackItem)
	}
	g.StackMeta[spell] = &game.StackItem{ID: spell, Kind: game.StackItemSpell, Controller: payer.ID, Owner: payer.ID, SourceCardID: spell}
	g.WithWriteLock(func() {
		if err := g.QueueCounterUnlessPaidForEffect(game.CounterUnlessPaidPrompt{StackItem: spell, Cost: "{2}", LifeCost: 2}); err != nil {
			t.Fatal(err)
		}
	})
	for _, life := range []int{40, 1} {
		payer.Life = life
		v := ViewOfGame(g)
		if len(v.PendingChoices) != 1 {
			t.Fatal("missing compound payment view")
		}
		c := v.PendingChoices[0]
		if c.PayCost != "{2}" || c.LifeCost != 2 || c.LifeCostUnpayable != (life < 2) {
			t.Fatalf("wrong compound view: %+v", c)
		}
	}
	payer.Life = 40
	g.WithWriteLock(func() { g.GrantLifeTotalLockForEffect(payer.ID, "Life lock", uuid.Nil, game.IndefiniteDuration()) })
	if !ViewOfGame(g).PendingChoices[0].LifeCostUnpayable {
		t.Fatal("a life lock must make the payment unavailable on the wire")
	}
}
