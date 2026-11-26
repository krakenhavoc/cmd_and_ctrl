package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b343_test.go — Commander 2014 cards, S82 OLD batch 343 (#5553),
// played through the real catalog entries.

const (
	oldB343Packmaster = "d4d27fef-e86b-4dd3-a15e-a0db8d0725ee"
	oldB343Crown      = "fed77e0a-a4a6-4409-9cab-bf0fc2cd9f9b"
)

func TestOldB343EveryCardIsRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{oldB343Packmaster: "Wren's Run Packmaster", oldB343Crown: "Crown of Doom"} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

func oldB343CastPackmaster(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Wren's Run Packmaster", TypeLine: "Creature — Elf Warrior", OracleID: oldB343Packmaster,
		ManaCost: "{3}{G}", Power: 5, Toughness: 5, Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestOldB343PackmasterChampionsAnElfAndReturnsItWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	elf := effectstest.B12Creature(g, me.ID, "Llanowar Elf", "Creature — Elf Druid", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	pack := oldB343CastPackmaster(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, elf)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(pack) {
		t.Fatal("the Packmaster was sacrificed despite championing an Elf")
	}
	if g.Battlefield.Contains(elf) {
		t.Error("the championed Elf is still on the battlefield")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("an unrelated creature left")
	}
	effectstest.KillCreature(t, g, me.ID, pack)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(oldB343Named(g, "Llanowar Elf", me.ID)); got != 1 {
		t.Errorf("Llanowar Elf on the battlefield %d times after the Packmaster left, want 1", got)
	}
}

func TestOldB343PackmasterIsSacrificedWithNoOtherElf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	pack := oldB343CastPackmaster(t, g)
	if g.Battlefield.Contains(pack) {
		t.Error("the Packmaster stayed with no Elf to champion")
	}
}

func TestOldB343PackmasterMakesWolvesAndGivesYourWolvesDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	pack := effectstest.MirB1Push(g, me.ID, "Wren's Run Packmaster", "Creature — Elf Warrior", oldB343Packmaster, 5, 5)
	mine := effectstest.B12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Wolf", "Creature — Wolf", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.MirB1Activate(g, me.ID, pack, 0); err == nil {
		t.Fatal("activated without paying {2}{G}")
	}
	effectstest.DcMana(t, g, me.ID, "{C}{C}{G}")
	if err := effectstest.MirB1Activate(g, me.ID, pack, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	wolves := oldB343Named(g, "Wolf", me.ID)
	var token game.Card
	for _, w := range wolves {
		if w.IsToken() {
			token = w
		}
	}
	if token.InstanceID == uuid.Nil || token.Power != 2 || token.Toughness != 2 {
		t.Fatalf("no 2/2 Wolf token: %+v", wolves)
	}
	for _, id := range []uuid.UUID{mine, token.InstanceID} {
		if !effectstest.HasKeywordOnBattlefield(t, g, id, "deathtouch") {
			t.Errorf("%s has no deathtouch", id)
		}
	}
	if effectstest.HasKeywordOnBattlefield(t, g, theirs, "deathtouch") || effectstest.HasKeywordOnBattlefield(t, g, bear, "deathtouch") {
		t.Error("an opponent's Wolf or a non-Wolf gained deathtouch")
	}
}

func TestOldB343CrownOfDoomPumpsCreaturesAttackingYouOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.PushCatalogPermanent(g, me.ID, "Crown of Doom", "Artifact", oldB343Crown, false)
	a := effectstest.B12Creature(g, opp.ID, "Attacker A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Attacker B", "Creature — Bear", 2, 2)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 1)
	if err := g.DeclareAttacker(a, me.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.DeclareAttacker(b, third.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, a); got != 4 {
		t.Errorf("the creature attacking me has power %d, want 4", got)
	}
	if got := effectstest.EffectivePower(t, g, b); got != 2 {
		t.Errorf("the creature attacking a third player has power %d, want 2", got)
	}
}

func TestOldB343CrownOfDoomHandsOverToAnyPlayerButItsOwner(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	crown := effectstest.PushCatalogPermanent(g, me.ID, "Crown of Doom", "Artifact", oldB343Crown, false)
	effectstest.DcMana(t, g, me.ID, "{C}{C}")
	if err := effectstest.MirB1Activate(g, me.ID, crown, 0, effectstest.PlayerTarget(me.ID)); err == nil {
		t.Fatal("the owner was accepted as the new controller")
	}
	if err := effectstest.MirB1Activate(g, me.ID, crown, 0, effectstest.PlayerTarget(opp.ID)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.B12Card(t, g, crown); c.Controller != opp.ID || c.Owner != me.ID {
		t.Errorf("controller %v owner %v, want the opponent to control it and me to own it", c.Controller, c.Owner)
	}
}

func oldB343Named(g *game.Game, name string, controller uuid.UUID) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == controller {
			out = append(out, c)
		}
	}
	return out
}
