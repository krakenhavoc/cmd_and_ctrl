package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s44_test.go — batch 44, theft from libraries and hands.

const (
	s44GontiOracle      = "d5be0ff6-39d9-4f9f-a028-e62dc38463f2"
	s44BlightwingOracle = "770894c0-01e7-4c1e-bcfb-989eee221dd2"
	s44RhetoricOracle   = "e7eb7ac8-08ab-419b-8508-987afd0bb93b"
	s44HeartlessOracle  = "cb9b317b-bb31-4c19-8b78-0fb9a61324b3"
	s44NarfiOracle      = "35c416c6-f522-4b4c-93db-75c4e01843d1"
)

func TestS44GontiLordOfLuxuryExilesOneOfTheTopFourAndBottomsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	opp.Library.Cards = nil
	var lib []uuid.UUID
	for i := 1; i <= 7; i++ {
		c := game.NewCard("Library Card "+string(rune('A'+i-1)), opp.ID)
		c.TypeLine = "Creature — Test"
		c.ManaCost = "{1}"
		opp.Library.PushTop(c)
		lib = append([]uuid.UUID{c.InstanceID}, lib...)
	}
	castHoldCreature(t, g, "Gonti, Lord of Luxury", "Legendary Creature — Aetherborn Rogue", s44GontiOracle, 2, 3)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("want a choose-exactly-one prompt, got %+v", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{lib[4]}); err == nil {
		t.Fatal("a card below the top four was accepted")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{lib[2]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	perm := e2Permission(g, lib[2])
	if perm == nil || perm.Player != me.ID || !perm.CastOnly || !perm.AnyType {
		t.Errorf("permission %+v, want an any-type cast grant to Gonti's controller", perm)
	}
	e2AssertFaceDownFor(t, g, lib[2], me.ID)
	if got := b12ZoneOf(g, lib[2]); got != game.ZoneExile {
		t.Errorf("picked card is in %q, want exile", got)
	}
	if len(opp.Library.Cards) != 6 {
		t.Fatalf("library has %d cards, want 6", len(opp.Library.Cards))
	}
	if top := opp.Library.Cards[len(opp.Library.Cards)-1].InstanceID; top != lib[4] {
		t.Errorf("top of library is %s, want the fifth card", top)
	}
	bottom := map[uuid.UUID]bool{}
	for _, c := range opp.Library.Cards[:3] {
		bottom[c.InstanceID] = true
	}
	for _, i := range []int{0, 1, 3} {
		if !bottom[lib[i]] {
			t.Errorf("looked-at card %d is not among the bottom three", i)
		}
	}
}

// s44Stack pushes n known cards on top of a player's library, returning their
// IDs top first.
func s44Stack(p *game.Player, n int, typeLine string) []uuid.UUID {
	p.Library.Cards = nil
	var ids []uuid.UUID
	for i := 1; i <= n; i++ {
		c := game.NewCard("Library Card "+string(rune('A'+i-1)), p.ID)
		c.TypeLine = typeLine
		c.ManaCost = "{1}"
		p.Library.PushTop(c)
		ids = append([]uuid.UUID{c.InstanceID}, ids...)
	}
	return ids
}

func TestS44BlightwingBanditStealsOnlyOnTheFirstSpellOfAnOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Blightwing Bandit", "Creature — Faerie Rogue", s44BlightwingOracle, 2, 2)
	// My own turn: silent.
	opp := g.Seats[1]
	lib := s44Stack(opp, 4, "Land")
	castCatalogSpell(t, g, "Shock", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if got := b12ZoneOf(g, lib[0]); got != game.ZoneLibrary {
		t.Fatalf("a spell on my own turn exiled the top card (%s)", got)
	}
	// The opponent's turn: the first spell steals, the second does not.
	advanceToNextSeatsTurn(t, g)
	advanceTo(t, g, game.StepPrecombatMain)
	active := g.Seats[g.Turn.ActiveSeat]
	if active == me {
		t.Fatal("still my turn")
	}
	lib = s44Stack(active, 4, "Land")
	b22CastInstantAs(t, g, me, "Opt")
	passPriorityAroundTable(t, g)
	if got := b12ZoneOf(g, lib[0]); got != game.ZoneExile {
		t.Fatalf("top card is in %q, want exile", got)
	}
	perm := e2Permission(g, lib[0])
	if perm == nil || perm.Player != me.ID || perm.CastOnly || !perm.AnyType {
		t.Errorf("permission %+v, want an any-type play grant to Blightwing's controller (a land is playable)", perm)
	}
	e2AssertFaceDownFor(t, g, lib[0], me.ID)
	b22CastInstantAs(t, g, me, "Brainstorm")
	passPriorityAroundTable(t, g)
	if got := b12ZoneOf(g, lib[1]); got != game.ZoneLibrary {
		t.Errorf("the second spell that turn exiled the next card (%s)", got)
	}
}

func TestS44CunningRhetoricTriggersOncePerAttackingOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	b21Push(g, me.ID, "Cunning Rhetoric", "Enchantment", s44RhetoricOracle, 0, 0)
	a := pushVanillaCreature(g, opp.ID, "Raider A", 1, 1)
	b := pushVanillaCreature(g, opp.ID, "Raider B", 1, 1)
	c := pushVanillaCreature(g, opp.ID, "Raider C", 1, 1)
	advanceToMainOf(t, g, 1)
	libOpp := s44Stack(opp, 5, "Creature — Test")
	libOther := s44Stack(other, 5, "Creature — Test")
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, d := range []struct{ id, at uuid.UUID }{{a, me.ID}, {b, me.ID}, {c, other.ID}} {
		if err := g.DeclareAttacker(d.id, d.at); err != nil {
			t.Fatal(err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := b12ZoneOf(g, libOpp[0]); got != game.ZoneExile {
		t.Fatalf("the attacker's top card is in %q, want exile", got)
	}
	if got := b12ZoneOf(g, libOpp[1]); got != game.ZoneLibrary {
		t.Errorf("two creatures at me exiled two cards (%s): one trigger per attacking opponent", got)
	}
	if got := b12ZoneOf(g, libOther[0]); got != game.ZoneLibrary {
		t.Errorf("the bystander's library was touched (%s)", got)
	}
	perm := e2Permission(g, libOpp[0])
	if perm == nil || perm.Player != me.ID || perm.CastOnly || !perm.AnyColor || perm.AnyType {
		t.Errorf("permission %+v, want a play grant spending mana as any color (not any type)", perm)
	}
}

func TestS44CunningRhetoricIgnoresAnAttackAtSomeoneElse(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	b21Push(g, me.ID, "Cunning Rhetoric", "Enchantment", s44RhetoricOracle, 0, 0)
	a := pushVanillaCreature(g, opp.ID, "Raider A", 1, 1)
	advanceToMainOf(t, g, 1)
	lib := s44Stack(opp, 3, "Creature — Test")
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, other.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := b12ZoneOf(g, lib[0]); got != game.ZoneLibrary {
		t.Errorf("an attack at another player exiled a card (%s)", got)
	}
}

func TestS44HeartlessConscriptionExilesEveryCreatureAndLetsYouPlayThem(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	theirs := pushCatalogPermanent(g, opp.ID, "Ogre", "Creature — Ogre", "", false)
	rock := pushCatalogPermanent(g, opp.ID, "Rock", "Artifact", "", false)
	spell := castCatalogSpell(t, g, "Heartless Conscription", "Sorcery", s44HeartlessOracle, nil)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirs} {
		if battlefieldHas(g, id) {
			t.Errorf("creature %s survived", id)
		}
		if got := b12ZoneOf(g, id); got != game.ZoneExile {
			t.Errorf("creature %s is in %q, want exile", id, got)
		}
		perm := e2Permission(g, id)
		if perm == nil || perm.Player != me.ID || perm.CastOnly || !perm.AnyType {
			t.Errorf("creature %s permission %+v, want an any-type play grant to the caster", id, perm)
		}
	}
	if !battlefieldHas(g, rock) {
		t.Error("a noncreature artifact was exiled")
	}
	if got := b12ZoneOf(g, spell); got != game.ZoneExile {
		t.Errorf("Heartless Conscription is in %q, want exile", got)
	}
}

func TestS44KingNarfisBetrayalMillsExilesFromEachGraveyardAndCastsLater(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	// Four cards each are milled: the creature sits in the top four.
	mineLib := s44Stack(me, 6, "Land")
	oppLib := s44Stack(opp, 6, "Land")
	for _, p := range []struct {
		pl   *game.Player
		ids  []uuid.UUID
		kind string
	}{{me, mineLib, "Creature — Test"}, {opp, oppLib, "Creature — Test"}} {
		for i := range p.pl.Library.Cards {
			if p.pl.Library.Cards[i].InstanceID == p.ids[1] {
				p.pl.Library.Cards[i].TypeLine = p.kind
			}
		}
	}
	saga := castCatalogSpell(t, g, "King Narfi's Betrayal", "Enchantment — Saga", s44NarfiOracle, nil)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mineLib[0], mineLib[3], oppLib[0], oppLib[3]} {
		if got := b12ZoneOf(g, id); got != game.ZoneGraveyard {
			t.Fatalf("card %s is in %q after the mill, want the graveyard", id, got)
		}
	}
	if got := b12ZoneOf(g, mineLib[4]); got != game.ZoneLibrary {
		t.Fatalf("the fifth card is in %q, want the library (four are milled)", got)
	}
	// One optional pick per graveyard with a creature or planeswalker card.
	picks := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards && c.Chooser == me.ID {
			picks++
			if c.ChooseMin != 0 || c.ChooseMax != 1 {
				t.Errorf("pick bounds %d..%d, want 0..1", c.ChooseMin, c.ChooseMax)
			}
		}
	}
	if picks != 2 {
		t.Fatalf("%d graveyard picks, want 2", picks)
	}
	for _, c := range append([]*game.PendingChoice(nil), g.PendingChoices...) {
		if c == nil || c.Kind != game.PendingChoiceChooseCards {
			continue
		}
		if len(c.ChooseCards) != 1 {
			t.Fatalf("a pick offers %d cards, want only the creature card", len(c.ChooseCards))
		}
		// Take the opponent's creature, leave mine.
		if c.FromPlayer == opp.ID {
			if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{oppLib[1]}); err != nil {
				t.Fatalf("pick: %v", err)
			}
		} else if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
			t.Fatalf("decline: %v", err)
		}
	}
	if got := b12ZoneOf(g, oppLib[1]); got != game.ZoneExile {
		t.Fatalf("the picked creature card is in %q, want exile", got)
	}
	if got := b12ZoneOf(g, mineLib[1]); got != game.ZoneGraveyard {
		t.Errorf("the declined card is in %q, want the graveyard", got)
	}
	if perm := e2Permission(g, oppLib[1]); perm != nil {
		t.Errorf("chapter I granted a permission %+v", perm)
	}
	// Chapter II: until end of turn, cast it with mana of any color.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(saga, game.CounterLore, 1) })
	passPriorityAroundTable(t, g)
	perm := e2Permission(g, oppLib[1])
	if perm == nil || perm.Player != me.ID || !perm.CastOnly || !perm.AnyColor || perm.AnyType {
		t.Fatalf("chapter II permission %+v, want a cast-only any-color grant to the Saga's controller", perm)
	}
	if perm := e2Permission(g, mineLib[1]); perm != nil {
		t.Errorf("a card not exiled with the Saga got %+v", perm)
	}
}
