package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adr0108_pr8_counters_test.go — ADR 0108 PR 8 (#1906): the counter
// prevention statics. Each card is driven with damage from several
// sources at once (one damage instance, so the unit its printed subject
// names is what decides how often the additional effect runs) and, where
// the rulings speak, with damage that can't be prevented (CR 615.12).

const (
	pr8cPhantomFlock        = "bcded242-6e54-416f-bdc8-093211c50e3f"
	pr8cPhantomNantuko      = "0951b529-646c-4dfd-88ad-84ee117ce722"
	pr8cPhantomNishoba      = "e43e06fb-52b7-4f38-8fac-f31973b043f7"
	pr8cPhantomNomad        = "bf65a7a7-a590-41eb-9844-184e5d63e32a"
	pr8cPhantomTiger        = "1755b3d4-0e40-40c4-b913-0960d55d411b"
	pr8cPhantomWurm         = "a866fb67-6614-45e6-928e-b4b59d95335f"
	pr8cOathswornKnight     = "16c293c8-bd1d-4db9-b197-da14f2f37cb1"
	pr8cUnbreathingHorde    = "e6cd9203-e4d3-4d9f-b59f-4e454fc5a477"
	pr8cUndergrowthChampion = "cb0eb84d-b41b-4147-aa7c-089b7fabd835"
	pr8cPolukranos          = "649e7237-b38b-43e9-83f4-763751fb1bea"
	pr8cUginsConjurant      = "787055e6-5d65-458e-a56e-66ef0574d2ca"
	pr8cMagmaPummeler       = "1a7c5807-afdc-4855-afd5-39839d96fc77"
	pr8cAntiVenom           = "3c7bafe9-80cd-48d0-bcae-e7910c9fb83b"
	pr8cPantherHabit        = "75ec25b3-ae91-46be-af8b-2649727638f4"
	pr8cStormwildCapridor   = "2e83e8c8-adc4-4e33-9f8b-966b7231a6c9"
	pr8cPhyrexianHydra      = "b16085d5-6d00-4d47-ab8b-d18d55c72141"
	pr8cIronscaleHydra      = "cef72b9b-91b1-47ff-ab6e-91d3c548e98b"
)

// pr8cPermanent puts a catalogued creature with printed `power` /
// `toughness` and `counters` +1/+1 counters on the battlefield for
// `owner`, as if it had entered (no entry replacement runs).
func pr8cPermanent(t *testing.T, g *game.Game, owner uuid.UUID, name, oracle string, power, toughness, counters int) uuid.UUID {
	t.Helper()
	id := apaPush(g, owner, owner, game.Card{Name: name, OracleID: oracle, TypeLine: "Creature — Test", Power: power, Toughness: toughness})
	if counters > 0 {
		g.WithWriteLock(func() {
			if err := g.AddCounterForEffect(id, game.CounterPlusOne, counters); err != nil {
				t.Fatal(err)
			}
		})
	}
	return id
}

// pr8cThree is three opposing creatures, three sources of damage.
func pr8cThree(g *game.Game, opp uuid.UUID) (uuid.UUID, uuid.UUID, uuid.UUID) {
	return pr7Creature(g, opp, "A", 2, "R"), pr7Creature(g, opp, "B", 2, "G"), pr7Creature(g, opp, "C", 2, "W")
}

// The Phantoms: each enters with its printed counters; three sources at
// once are all prevented and cost one counter.
func TestADR0108PR8CountersPhantoms(t *testing.T) {
	for _, c := range []struct {
		name, oracle, typeLine string
		counters               int
	}{
		{"Phantom Flock", pr8cPhantomFlock, "Creature — Bird Soldier Spirit", 3},
		{"Phantom Nantuko", pr8cPhantomNantuko, "Creature — Insect Spirit", 2},
		{"Phantom Nishoba", pr8cPhantomNishoba, "Creature — Cat Beast Spirit", 7},
		{"Phantom Nomad", pr8cPhantomNomad, "Creature — Spirit Nomad", 2},
		{"Phantom Tiger", pr8cPhantomTiger, "Creature — Cat Spirit", 2},
		{"Phantom Wurm", pr8cPhantomWurm, "Creature — Wurm Spirit", 4},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			id := pr8Enter(t, g, me, c.name, c.typeLine, c.oracle)
			if n := pr8Counters(g, id, game.CounterPlusOne); n != c.counters {
				t.Fatalf("entered with %d counters, want %d", n, c.counters)
			}
			a, b, x := pr8cThree(g, opp.ID)
			pr8Hit(t, g, id, map[uuid.UUID]int{a: 1, b: 1, x: 1})
			if pr6Marked(g, id) != 0 || pr8Counters(g, id, game.CounterPlusOne) != c.counters-1 {
				t.Fatalf("three at once: damage %d (want 0), counters %d (want %d)",
					pr6Marked(g, id), pr8Counters(g, id, game.CounterPlusOne), c.counters-1)
			}
		})
	}
}

// Phantom Nantuko's tap ability puts a counter back.
func TestADR0108PR8CountersPhantomNantukoTaps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, "Phantom Nantuko", "Creature — Insect Spirit", pr8cPhantomNantuko, false)
	pr7Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	if n := pr8Counters(g, id, game.CounterPlusOne); n != 1 {
		t.Fatalf("counters %d after {T}, want 1", n)
	}
}

// Phantom Nishoba: whenever it deals damage, you gain that much life.
func TestADR0108PR8CountersPhantomNishobaGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := pr8cPermanent(t, g, me.ID, "Phantom Nishoba", pr8cPhantomNishoba, 0, 0, 7)
	life := me.Life
	pr7Hit(t, g, id, opp.ID, 7)
	passPriorityAroundTable(t, g)
	if me.Life != life+7 {
		t.Fatalf("life %d, want %d", me.Life, life+7)
	}
}

// Oathsworn Knight: several sources at once cost one counter; damage
// that can't be prevented is dealt and still costs one (the rulings).
func TestADR0108PR8CountersOathswornKnight(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := pr8Enter(t, g, me, "Oathsworn Knight", "Creature — Human Knight", pr8cOathswornKnight)
	if n := pr8Counters(g, id, game.CounterPlusOne); n != 4 {
		t.Fatalf("entered with %d counters, want 4", n)
	}
	a, b, _ := pr8cThree(g, opp.ID)
	pr8Hit(t, g, id, map[uuid.UUID]int{a: 3, b: 3})
	if pr6Marked(g, id) != 0 || pr8Counters(g, id, game.CounterPlusOne) != 3 {
		t.Fatalf("two at once: damage %d (want 0), counters %d (want 3)", pr6Marked(g, id), pr8Counters(g, id, game.CounterPlusOne))
	}
	pr8Unpreventable(t, g, a, id, 1)
	if pr6Marked(g, id) != 1 || pr8Counters(g, id, game.CounterPlusOne) != 2 {
		t.Fatalf("unpreventable: damage %d (want 1), counters %d (want 2)", pr6Marked(g, id), pr8Counters(g, id, game.CounterPlusOne))
	}
}

// Undergrowth Champion: prevents only while it has a counter; landfall
// adds one.
func TestADR0108PR8CountersUndergrowthChampion(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := pr8cPermanent(t, g, me.ID, "Undergrowth Champion", pr8cUndergrowthChampion, 2, 2, 0)
	a, b, _ := pr8cThree(g, opp.ID)
	pr6Damage(t, g, a, id, 1)
	if pr6Marked(g, id) != 1 {
		t.Fatalf("no counter: damage %d, want 1 (nothing prevented)", pr6Marked(g, id))
	}
	land := game.NewCard("Forest", me.ID)
	land.TypeLine = "Basic Land — Forest"
	me.Hand.PushTop(land)
	g.WithWriteLock(func() {
		if _, err := g.PutFromHandOntoBattlefieldForEffect(land.InstanceID, game.HandEntryOptions{Controller: me.ID}); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if n := pr8Counters(g, id, game.CounterPlusOne); n != 1 {
		t.Fatalf("landfall: counters %d, want 1", n)
	}
	pr8Hit(t, g, id, map[uuid.UUID]int{a: 2, b: 2})
	if pr6Marked(g, id) != 1 || pr8Counters(g, id, game.CounterPlusOne) != 0 {
		t.Fatalf("two at once: damage %d (want 1), counters %d (want 0)", pr6Marked(g, id), pr8Counters(g, id, game.CounterPlusOne))
	}
}

// Unbreathing Horde counts other Zombies you control and Zombie cards in
// your graveyard, and loses one counter however much is prevented.
func TestADR0108PR8CountersUnbreathingHorde(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	apaPush(g, me.ID, me.ID, game.Card{Name: "Zombie One", TypeLine: "Creature — Zombie", Power: 2, Toughness: 2})
	apaPush(g, opp.ID, opp.ID, game.Card{Name: "Their Zombie", TypeLine: "Creature — Zombie", Power: 2, Toughness: 2})
	z := game.NewCard("Dead Zombie", me.ID)
	z.TypeLine = "Creature — Zombie"
	me.Graveyard.PushTop(z)
	id := pr8Enter(t, g, me, "Unbreathing Horde", "Creature — Zombie", pr8cUnbreathingHorde)
	if n := pr8Counters(g, id, game.CounterPlusOne); n != 2 {
		t.Fatalf("entered with %d counters, want 2 (one Zombie you control, one in your graveyard)", n)
	}
	a, b, _ := pr8cThree(g, opp.ID)
	pr8Hit(t, g, id, map[uuid.UUID]int{a: 4, b: 4})
	if pr6Marked(g, id) != 0 || pr8Counters(g, id, game.CounterPlusOne) != 1 {
		t.Fatalf("8 at once: damage %d (want 0), counters %d (want 1)", pr6Marked(g, id), pr8Counters(g, id, game.CounterPlusOne))
	}
}

// Ugin's Conjurant removes that many — every counter when that is
// fewer — and still removes them under unpreventable damage.
func TestADR0108PR8CountersUginsConjurant(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := pr8cPermanent(t, g, me.ID, "Ugin's Conjurant", pr8cUginsConjurant, 0, 0, 5)
	a, b, _ := pr8cThree(g, opp.ID)
	pr8Hit(t, g, id, map[uuid.UUID]int{a: 1, b: 1})
	if pr6Marked(g, id) != 0 || pr8Counters(g, id, game.CounterPlusOne) != 3 {
		t.Fatalf("2 at once: damage %d (want 0), counters %d (want 3)", pr6Marked(g, id), pr8Counters(g, id, game.CounterPlusOne))
	}
	pr8Unpreventable(t, g, a, id, 1)
	if pr6Marked(g, id) != 1 || pr8Counters(g, id, game.CounterPlusOne) != 2 {
		t.Fatalf("unpreventable: damage %d (want 1), counters %d (want 2)", pr6Marked(g, id), pr8Counters(g, id, game.CounterPlusOne))
	}
	pr6Damage(t, g, a, id, 9)
	if findBattlefieldCardForTest(g, id) != nil {
		t.Fatal("9 damage to a 2/2 with 1 damage: every counter should come off and it should die")
	}
}

// Polukranos: six counters on an ordinary entry, twelve when it escapes;
// that many come off; it fights another target creature.
func TestADR0108PR8CountersPolukranos(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := pr8Enter(t, g, me, "Polukranos, Unchained", "Legendary Creature — Zombie Hydra", pr8cPolukranos)
	if n := pr8Counters(g, id, game.CounterPlusOne); n != 6 {
		t.Fatalf("entered with %d counters, want 6", n)
	}
	a, b, _ := pr8cThree(g, opp.ID)
	pr8Hit(t, g, id, map[uuid.UUID]int{a: 2, b: 1})
	if pr6Marked(g, id) != 0 || pr8Counters(g, id, game.CounterPlusOne) != 3 {
		t.Fatalf("3 at once: damage %d (want 0), counters %d (want 3)", pr6Marked(g, id), pr8Counters(g, id, game.CounterPlusOne))
	}
	victim := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Victim", TypeLine: "Creature — Test", Power: 1, Toughness: 9})
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("activate fight: %v", err)
	}
	passPriorityAroundTable(t, g)
	if pr6Marked(g, victim) != 3 || pr6Marked(g, id) != 0 || pr8Counters(g, id, game.CounterPlusOne) != 2 {
		t.Fatalf("fight: victim damage %d (want 3), Polukranos damage %d (want 0), counters %d (want 2)",
			pr6Marked(g, victim), pr6Marked(g, id), pr8Counters(g, id, game.CounterPlusOne))
	}
}

// Damage that can't be prevented is dealt to Polukranos and still
// removes that many counters, before lethal damage is checked (the
// ruling).
func TestADR0108PR8CountersPolukranosUnpreventable(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := pr8cPermanent(t, g, me.ID, "Polukranos, Unchained", pr8cPolukranos, 0, 0, 6)
	a, _, _ := pr8cThree(g, opp.ID)
	pr8Unpreventable(t, g, a, id, 2)
	if pr6Marked(g, id) != 2 || pr8Counters(g, id, game.CounterPlusOne) != 4 {
		t.Fatalf("unpreventable: damage %d (want 2), counters %d (want 4)", pr6Marked(g, id), pr8Counters(g, id, game.CounterPlusOne))
	}
	pr8Unpreventable(t, g, a, id, 2)
	if findBattlefieldCardForTest(g, id) != nil {
		t.Fatal("a 2/2 with 4 damage is still on the battlefield")
	}
}

// Polukranos escapes with twelve counters, not eighteen.
func TestADR0108PR8CountersPolukranosEscapesWithTwelve(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	id := seedGraveyardCard(t, g, "Polukranos, Unchained", "Legendary Creature — Zombie Hydra", pr8cPolukranos)
	pay := seedGraveyardFodder(t, g, 6)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "escape", AltCostIDs: pay,
	}); err != nil {
		t.Fatalf("escape cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := pr8Counters(g, id, game.CounterPlusOne); n != 12 {
		t.Fatalf("escaped with %d counters, want 12", n)
	}
}

// Magma Pummeler: dealt more than it has counters, all of it is
// prevented, every counter comes off, and it deals that much damage to
// any target (the ruling).
func TestADR0108PR8CountersMagmaPummeler(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := pr8cPermanent(t, g, me.ID, "Magma Pummeler", pr8cMagmaPummeler, 0, 0, 2)
	a, b, _ := pr8cThree(g, opp.ID)
	pr8Hit(t, g, id, map[uuid.UUID]int{a: 2, b: 2})
	life := opp.Life
	for i := 0; i < 4; i++ {
		c := pr8cPickTarget(g)
		if c == nil {
			passPriorityAroundTable(t, g)
			continue
		}
		if err := g.ResolvePickTarget(c.ID, c.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatalf("pick target: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	if findBattlefieldCardForTest(g, id) != nil {
		t.Fatal("the Pummeler should have lost both counters and died")
	}
	if opp.Life != life-4 {
		t.Fatalf("opponent life %d, want %d: it deals that much (4) damage", opp.Life, life-4)
	}
}

func pr8cPickTarget(g *game.Game) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			return c
		}
	}
	return nil
}

// Anti-Venom: that many counters, still added under unpreventable damage;
// put onto the battlefield without being cast, it returns nothing.
func TestADR0108PR8CountersAntiVenom(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := game.NewCard("Dead Bear", me.ID)
	dead.TypeLine = "Creature — Bear"
	me.Graveyard.PushTop(dead)
	id := pr8Enter(t, g, me, "Anti-Venom, Horrifying Healer", "Legendary Creature — Symbiote Hero", pr8cAntiVenom)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dead.InstanceID) || len(g.PendingChoices) != 0 {
		t.Fatal("Anti-Venom was not cast, so nothing returns")
	}
	a, b, _ := pr8cThree(g, opp.ID)
	pr8Hit(t, g, id, map[uuid.UUID]int{a: 2, b: 1})
	if pr6Marked(g, id) != 0 || pr8Counters(g, id, game.CounterPlusOne) != 3 {
		t.Fatalf("3 at once: damage %d (want 0), counters %d (want 3)", pr6Marked(g, id), pr8Counters(g, id, game.CounterPlusOne))
	}
	pr8Unpreventable(t, g, a, id, 2)
	if pr6Marked(g, id) != 2 || pr8Counters(g, id, game.CounterPlusOne) != 5 {
		t.Fatalf("unpreventable: damage %d (want 2), counters %d (want 5)", pr6Marked(g, id), pr8Counters(g, id, game.CounterPlusOne))
	}
}

// Anti-Venom cast: his enter trigger returns the target creature card.
func TestADR0108PR8CountersAntiVenomCastReturnsACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dead := game.NewCard("Dead Bear", me.ID)
	dead.TypeLine = "Creature — Bear"
	me.Graveyard.PushTop(dead)
	castCatalogSpell(t, g, "Anti-Venom, Horrifying Healer", "Legendary Creature — Symbiote Hero", pr8cAntiVenom, nil)
	passPriorityAroundTable(t, g)
	for i := 0; i < 3 && !g.Battlefield.Contains(dead.InstanceID); i++ {
		if c := pr8cPickTarget(g); c != nil {
			if err := g.ResolvePickTarget(c.ID, c.Chooser, game.TargetRef{Kind: game.TargetCard, ID: dead.InstanceID}); err != nil {
				t.Fatalf("pick target: %v", err)
			}
		}
		passPriorityAroundTable(t, g)
	}
	if !g.Battlefield.Contains(dead.InstanceID) {
		t.Fatal("cast Anti-Venom: the creature card did not return")
	}
}

// Panther Habit: the equipped creature gets that many counters, under
// unpreventable damage too.
func TestADR0108PR8CountersPantherHabit(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	habit := pushCatalogPermanent(g, me.ID, "Panther Habit", "Artifact — Equipment", pr8cPantherHabit, false)
	bear := apaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	other := apaPush(g, me.ID, me.ID, game.Card{Name: "Other", TypeLine: "Creature — Bear", Power: 2, Toughness: 5})
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(habit, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
			t.Fatal(err)
		}
	})
	a, b, _ := pr8cThree(g, opp.ID)
	pr8Hit(t, g, bear, map[uuid.UUID]int{a: 2, b: 1})
	if pr6Marked(g, bear) != 0 || pr8Counters(g, bear, game.CounterPlusOne) != 3 {
		t.Fatalf("3 at once: damage %d (want 0), counters %d (want 3)", pr6Marked(g, bear), pr8Counters(g, bear, game.CounterPlusOne))
	}
	pr8Unpreventable(t, g, a, bear, 1)
	if pr6Marked(g, bear) != 1 || pr8Counters(g, bear, game.CounterPlusOne) != 4 {
		t.Fatalf("unpreventable: damage %d (want 1), counters %d (want 4)", pr6Marked(g, bear), pr8Counters(g, bear, game.CounterPlusOne))
	}
	pr6Damage(t, g, a, other, 2)
	if pr6Marked(g, other) != 2 {
		t.Fatalf("an unequipped creature's damage was prevented: %d", pr6Marked(g, other))
	}
}

// Stormwild Capridor: noncombat damage is prevented with a counter per 1;
// unpreventable damage puts none (the ruling); combat damage is untouched.
func TestADR0108PR8CountersStormwildCapridor(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := pr8cPermanent(t, g, me.ID, "Stormwild Capridor", pr8cStormwildCapridor, 1, 3, 0)
	a, b, _ := pr8cThree(g, opp.ID)
	pr8Hit(t, g, id, map[uuid.UUID]int{a: 2, b: 1})
	if pr6Marked(g, id) != 0 || pr8Counters(g, id, game.CounterPlusOne) != 3 {
		t.Fatalf("3 at once: damage %d (want 0), counters %d (want 3)", pr6Marked(g, id), pr8Counters(g, id, game.CounterPlusOne))
	}
	pr8Unpreventable(t, g, a, id, 1)
	if pr6Marked(g, id) != 1 || pr8Counters(g, id, game.CounterPlusOne) != 3 {
		t.Fatalf("unpreventable: damage %d (want 1), counters %d (want 3)", pr6Marked(g, id), pr8Counters(g, id, game.CounterPlusOne))
	}
	if err := g.MarkCombatDamage(a, id, 2); err != nil {
		t.Fatal(err)
	}
	if pr6Marked(g, id) != 3 {
		t.Fatalf("combat damage: %d marked, want 3", pr6Marked(g, id))
	}
}

// Phyrexian Hydra: a -1/-1 counter per 1 prevented; none under
// unpreventable damage (the ruling).
func TestADR0108PR8CountersPhyrexianHydra(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := pr8cPermanent(t, g, me.ID, "Phyrexian Hydra", pr8cPhyrexianHydra, 7, 7, 0)
	a, b, _ := pr8cThree(g, opp.ID)
	pr8Hit(t, g, id, map[uuid.UUID]int{a: 2, b: 1})
	if pr6Marked(g, id) != 0 || pr8Counters(g, id, game.CounterMinusOne) != 3 {
		t.Fatalf("3 at once: damage %d (want 0), -1/-1 counters %d (want 3)", pr6Marked(g, id), pr8Counters(g, id, game.CounterMinusOne))
	}
	pr8Unpreventable(t, g, a, id, 2)
	if pr6Marked(g, id) != 2 || pr8Counters(g, id, game.CounterMinusOne) != 3 {
		t.Fatalf("unpreventable: damage %d (want 2), -1/-1 counters %d (want 3)", pr6Marked(g, id), pr8Counters(g, id, game.CounterMinusOne))
	}
}

// Ironscale Hydra: blocked by two creatures, both creatures' combat
// damage is prevented and it gets two counters (one per creature, however
// much); noncombat damage is untouched.
func TestADR0108PR8CountersIronscaleHydra(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hydra := pr8cPermanent(t, g, me.ID, "Ironscale Hydra", pr8cIronscaleHydra, 5, 5, 0)
	b1 := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Blocker One", TypeLine: "Creature — Test", Power: 3, Toughness: 2})
	b2 := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Blocker Two", TypeLine: "Creature — Test", Power: 4, Toughness: 2})
	pinger := pr7Creature(g, opp.ID, "Pinger", 1, "R")
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(hydra, opp.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	for _, b := range []uuid.UUID{b1, b2} {
		if err := g.DeclareBlocker(b, hydra); err != nil {
			t.Fatal(err)
		}
	}
	advanceTo(t, g, game.StepCombatDamage)
	for i := 0; i < 4; i++ {
		var c *game.PendingChoice
		for _, p := range g.PendingChoices {
			if p != nil && p.Kind == game.PendingChoiceDamageAssignment {
				c = p
			}
		}
		if c == nil {
			break
		}
		entries := []game.DamageAssignmentEntry{{BlockerID: b1, Amount: 2}, {BlockerID: b2, Amount: 3}}
		if err := g.ResolveDamageAssignment(c.ID, c.Chooser, entries, 0); err != nil {
			t.Fatalf("assign: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if pr6Marked(g, hydra) != 0 || pr8Counters(g, hydra, game.CounterPlusOne) != 2 {
		t.Fatalf("blocked by two: damage %d (want 0), counters %d (want 2)", pr6Marked(g, hydra), pr8Counters(g, hydra, game.CounterPlusOne))
	}
	pr6Damage(t, g, pinger, hydra, 1)
	if pr6Marked(g, hydra) != 1 {
		t.Fatalf("noncombat damage from a creature: %d marked, want 1", pr6Marked(g, hydra))
	}
}
