package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	s33BackInTownOracle    = "ce723db6-2a80-42a5-a176-cd322b4e1e27"
	s33WeRideAtDawnOracle  = "5675ac63-b69b-41b9-a7ed-4085a4e431b2"
	s33LifeInsuranceOracle = "50779f73-c538-4364-aca4-4ab069a00814"
)

func TestS33CardsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s33BackInTownOracle:    "Back in Town",
		s33LifeInsuranceOracle: "Life Insurance",
		s33WeRideAtDawnOracle:  "We Ride at Dawn",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestS33LifeInsuranceTreasureForEachNontokenCreatureDeath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Life Insurance", "Enchantment", s33LifeInsuranceOracle, 0, 0)
	tok := socB2Token(g, me.ID, "Goblin", "Goblin", 1, 1)
	b27Kill(g, tok)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Treasure"); n != 0 {
		t.Fatalf("%d Treasures from a token dying, want 0", n)
	}
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := me.Life
	b27Kill(g, theirs)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("%d Treasures after an opponent's creature died, want 1", n)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b27Kill(g, mine)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Treasure"); n != 2 {
		t.Errorf("%d Treasures after two deaths, want 2", n)
	}
	_ = game.ZoneBattlefield
}

func TestS33BackInTownReturnsXOutlawCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pirate := pushGraveyardPermanent(me, "Pirate", "Creature — Human Pirate", "{1}{U}")
	rogue := pushGraveyardPermanent(me, "Rogue", "Creature — Human Rogue", "{1}{B}")
	bear := pushGraveyardPermanent(me, "Bear", "Creature — Bear", "{1}{G}")
	rock := pushGraveyardPermanent(me, "Rock", "Artifact — Mercenary", "{1}")

	castXSpell(t, g, "Back in Town", "Sorcery", s33BackInTownOracle, "{X}{2}{B}", 2, cardRefs(pirate, rogue))
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(pirate) || !g.Battlefield.Contains(rogue) {
		t.Error("a targeted outlaw creature card did not return")
	}
	if !me.Graveyard.Contains(bear) || !me.Graveyard.Contains(rock) {
		t.Error("an untargeted card left the graveyard")
	}
}

func TestS33BackInTownTargetsExactlyXOutlawCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pirate := pushGraveyardPermanent(me, "Pirate", "Creature — Human Pirate", "{1}{U}")
	bear := pushGraveyardPermanent(me, "Bear", "Creature — Bear", "{1}{G}")
	rock := pushGraveyardPermanent(me, "Rock", "Artifact — Mercenary", "{1}")
	advanceToMain(t, g)
	cast := func(x int, targets ...uuid.UUID) error {
		floatMana(t, g, me, "{B}{C}{C}"+string(slices.Repeat([]byte("{C}"), x)))
		_, err := castFromHandErr(g, me, "Back in Town", "Sorcery", "{X}{2}{B}", s33BackInTownOracle, game.CastSpellParams{Strict: true, XValue: x, Targets: cardRefs(targets...)})
		g.WithWriteLock(func() { me.ManaPool = game.ManaPool{} })
		return err
	}
	if err := cast(1, bear); err == nil {
		t.Error("a non-outlaw creature card was a legal target")
	}
	if err := cast(1, rock); err == nil {
		t.Error("a noncreature outlaw card was a legal target")
	}
	if err := cast(2, pirate); err == nil {
		t.Error("one target was accepted for X = 2")
	}
	if err := cast(1, pirate); err != nil {
		t.Fatalf("X = 1 with an outlaw creature card was refused: %v", err)
	}
}

func TestS33WeRideAtDawnGivesLegendaryCreatureSpellsConvokeOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	b12Push(g, me.ID, "We Ride at Dawn", "Enchantment", s33WeRideAtDawnOracle, 0, 0)
	elves := cgTwoGreenCreatures(g, me.ID)

	plain := cgBear(me.ID)
	g.WithWriteLock(func() { me.Hand.PushTop(plain) })
	if err := g.CastSpell(me.ID, plain.InstanceID, game.CastSpellParams{Strict: true, TapIDs: elves}); err == nil {
		t.Fatal("a nonlegendary creature spell was convoked")
	}
	legend := cgBear(me.ID)
	legend.Name, legend.TypeLine = "Legend", "Legendary Creature — Elf"
	g.WithWriteLock(func() { me.Hand.PushTop(legend) })
	if err := g.CastSpell(me.ID, legend.InstanceID, game.CastSpellParams{Strict: true, TapIDs: elves}); err != nil {
		t.Fatalf("convoking a legendary creature spell: %v", err)
	}
	if !cgTapped(g, elves) {
		t.Error("the convoking creatures were not tapped")
	}
}

func TestS33WeRideAtDawnMakesAMercenaryWhenYourCommanderAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Push(g, me.ID, "We Ride at Dawn", "Enchantment", s33WeRideAtDawnOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	cmdr := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Boss", TypeLine: "Legendary Creature — Human",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, IsCommander: true,
	})
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Mercenary"); n != 0 {
		t.Fatalf("%d Mercenaries after a non-commander attacked, want 0", n)
	}
	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	advanceToMain(t, g2)
	b12Push(g2, me2.ID, "We Ride at Dawn", "Enchantment", s33WeRideAtDawnOracle, 0, 0)
	cmdr2 := pushBattlefieldCardWithTimestamp(g2, game.Card{
		InstanceID: uuid.New(), Name: "Boss", TypeLine: "Legendary Creature — Human",
		Power: 2, Toughness: 2, Owner: me2.ID, Controller: me2.ID, IsCommander: true,
	})
	declareAttack(t, g2, opp2.ID, cmdr2)
	passPriorityAroundTable(t, g2)
	if n, _ := countTokensNamed(g2, me2.ID, "Mercenary"); n != 1 {
		t.Errorf("%d Mercenaries after the commander attacked, want 1", n)
	}
	_ = cmdr
}
