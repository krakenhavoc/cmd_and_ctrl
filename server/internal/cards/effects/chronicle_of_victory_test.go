package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const chronicleOfVictoryOracle = "e949b362-4988-4482-ac51-c1049af9040c"

func TestChronicleOfVictoryIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(chronicleOfVictoryOracle)
	if !ok || spec.Name != "Chronicle of Victory" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// Cast for {6}, Goblin chosen as it enters: my Goblins are 4/4 first
// strike and trample, nothing else is; my Goblin spells (a kindred one
// too) draw a card, other spells and an opponent's Goblin spell do not.
func TestChronicleOfVictoryAnthemsAndDrawsForTheChosenType(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	gob := b12Creature(g, me.ID, "My Goblin", "Creature — Goblin Warrior", 2, 2)
	elf := b12Creature(g, me.ID, "My Elf", "Creature — Elf", 2, 2)
	theirGob := b12Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 2, 2)

	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Chronicle of Victory", TypeLine: "Legendary Artifact",
		OracleID: chronicleOfVictoryOracle, ManaCost: "{6}", Owner: me.ID, Controller: me.ID})
	b06AddMana(me, "C", "C", "C", "C", "C", "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerCreatureType(t, g, me.ID, "Goblin")

	c, _ := battlefieldCard(g, gob)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 || !game.HasKeyword(&c, "first strike") || !game.HasKeyword(&c, "trample") {
		t.Errorf("my Goblin is %d/%d first strike %v trample %v, want a 4/4 with both",
			c.CurrentPower(), c.CurrentToughness(), game.HasKeyword(&c, "first strike"), game.HasKeyword(&c, "trample"))
	}
	for name, other := range map[string]uuid.UUID{"my Elf": elf, "their Goblin": theirGob} {
		o, _ := battlefieldCard(g, other)
		if o.CurrentPower() != 2 || game.HasKeyword(&o, "first strike") || game.HasKeyword(&o, "trample") {
			t.Errorf("%s was boosted", name)
		}
	}

	seedLibrary(me, "A", "B", "C", "D")
	seedLibrary(opp, "E", "F")
	before := me.Hand.Size()
	castFree(t, g, me, "Goblin Test", "Creature — Goblin")
	if got := me.Hand.Size() - before; got != 1 {
		t.Fatalf("a Goblin creature spell drew %d, want 1", got)
	}
	before = me.Hand.Size()
	castFree(t, g, me, "Goblin Surge", "Kindred Instant — Goblin")
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("a Kindred Goblin instant drew %d, want 1", got)
	}
	before = me.Hand.Size()
	castFree(t, g, me, "Elf Test", "Creature — Elf")
	if got := me.Hand.Size() - before; got != 0 {
		t.Errorf("an Elf spell drew %d, want 0", got)
	}
	oppBefore, myBefore := opp.Hand.Size(), me.Hand.Size()
	castFree(t, g, opp, "Their Goblin Surge", "Kindred Instant — Goblin")
	if opp.Hand.Size()-oppBefore != 0 || me.Hand.Size()-myBefore != 0 {
		t.Error("an opponent's Goblin spell drew a card")
	}
}

// {6} is the cost: five mana does not cast it.
func TestChronicleOfVictoryRefusesAnUnpaidCast(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Chronicle of Victory", TypeLine: "Legendary Artifact",
		OracleID: chronicleOfVictoryOracle, ManaCost: "{6}", Owner: me.ID, Controller: me.ID})
	b06AddMana(me, "C", "C", "C", "C", "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("cast for five mana")
	}
}
