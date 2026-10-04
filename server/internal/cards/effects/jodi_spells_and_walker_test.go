package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// jodi_spells_and_walker_test.go — S58 PR 5: Ulamog, the mana doublers,
// Ugin's Binding, Sunbird's Invocation, Emergent Ultimatum and Nicol
// Bolas. See jodi_big_spells_test.go for the oracle constants.

// --- Ulamog, the Ceaseless Hunger -------------------------------------

func TestUlamogCastTriggerExilesTwoTargetPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushCatalogPermanent(g, opp.ID, "Bear A", "Creature — Bear", "", false)
	b := pushCatalogPermanent(g, opp.ID, "Bear B", "Creature — Bear", "", false)
	keep := pushCatalogPermanent(g, opp.ID, "Bear C", "Creature — Bear", "", false)
	id := putInHand(me, game.Card{Name: "Ulamog, the Ceaseless Hunger", TypeLine: "Legendary Creature — Eldrazi",
		OracleID: ulamogCeaselessOracle, ManaCost: "{10}", Power: 10, Toughness: 10})
	toMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell Ulamog: %v", err)
	}
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the cast trigger should ask for its two targets")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b},
	}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	if !g.Battlefield.Contains(a) {
		t.Error("nothing is exiled until the trigger resolves")
	}
	passPriorityAroundTable(t, g)

	if !inExile(g, a) || !inExile(g, b) {
		t.Errorf("both targets should be exiled (a=%v b=%v)", inExile(g, a), inExile(g, b))
	}
	if !g.Battlefield.Contains(keep) {
		t.Error("an untargeted permanent is untouched")
	}
	card, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("Ulamog resolves after its cast trigger")
	}
	if !game.HasKeyword(&card, "indestructible") {
		t.Error("Ulamog is indestructible")
	}
}

func TestUlamogAttackExilesTwentyFromTheDefendersLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	third := g.Seats[2]
	for i := 0; i < 30; i++ {
		opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Land", Owner: opp.ID, Controller: opp.ID})
	}
	ulamog := pushCatalogPermanent(g, me.ID, "Ulamog, the Ceaseless Hunger", "Legendary Creature — Eldrazi", ulamogCeaselessOracle, false)
	libBefore, thirdBefore := opp.Library.Size(), third.Library.Size()
	declareAttack(t, g, opp.ID, ulamog)
	passPriorityAroundTable(t, g)

	if got := opp.Library.Size(); got != libBefore-20 {
		t.Errorf("defender's library %d -> %d, want 20 fewer", libBefore, got)
	}
	if got := third.Library.Size(); got != thirdBefore {
		t.Errorf("a non-defender's library changed: %d -> %d", thirdBefore, got)
	}
	exiled := 0
	for _, c := range g.Exile.Cards {
		if c.Owner == opp.ID {
			exiled++
		}
	}
	if exiled != 20 {
		t.Errorf("%d of the defender's cards in exile, want 20", exiled)
	}
	if opp.Graveyard.Size() != 0 {
		t.Error("the cards are exiled, not milled")
	}
}

// --- mana doublers -----------------------------------------------------

func TestZendikarResurgentDoublesALandsMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Zendikar Resurgent", "Enchantment", zendikarResurgentOracle, false)
	forest := b31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	tapForMana(t, g, me.ID, forest)
	if got := sortedPool(me); len(got) != 2 || got[0] != "G" || got[1] != "G" {
		t.Errorf("pool = %v, want {G}{G}", got)
	}
	g.ReadSnapshot(func() {
		if len(g.PendingTriggers) != 0 {
			t.Errorf("a triggered mana ability does not use the stack (CR 605.4a): %d pending", len(g.PendingTriggers))
		}
	})
}

func TestZendikarResurgentDrawsOnACreatureSpell(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		draws          int
	}{
		{"creature", "Creature — Bear", 1},
		{"noncreature", "Sorcery", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			pushCatalogPermanent(g, me.ID, "Zendikar Resurgent", "Enchantment", zendikarResurgentOracle, false)
			before := me.Hand.Size()
			castCatalogSpell(t, g, "A Spell", tc.typeLine, "", nil)
			passPriorityAroundTable(t, g)
			if got := me.Hand.Size(); got != before+tc.draws {
				t.Errorf("hand %d -> %d, want +%d", before, got, tc.draws)
			}
		})
	}
}

func TestVorinclexDoublesYourManaButNotAnOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Vorinclex, Voice of Hunger", "Legendary Creature — Phyrexian Praetor", vorinclexVoiceOracle, false)
	mine := b31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	theirs := b31Push(g, opp.ID, "Island", "Basic Land — Island", "", "", 0, 0)
	tapForMana(t, g, me.ID, mine)
	if got := sortedPool(me); len(got) != 2 {
		t.Errorf("my pool = %v, want two mana", got)
	}
	tapForMana(t, g, opp.ID, theirs)
	if got := sortedPool(opp); len(got) != 1 {
		t.Errorf("an opponent's land is not doubled: pool = %v", got)
	}
}

// An opponent's tapped land misses its controller's next untap step —
// and only that one.
func TestVorinclexFreezesAnOpponentsLandThatTapsForMana(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Vorinclex, Voice of Hunger", "Legendary Creature — Phyrexian Praetor", vorinclexVoiceOracle, false)
	island := b31Push(g, opp.ID, "Island", "Basic Land — Island", "", "", 0, 0)
	tapForMana(t, g, opp.ID, island)
	if tapped, _ := battlefieldCardTapped(g, island); !tapped {
		t.Fatal("the Island should be tapped")
	}
	passPriorityAroundTable(t, g)

	// The opponent's next untap step: the Island stays tapped.
	advanceToUpkeepOfSeat(t, g, 1)
	if tapped, _ := battlefieldCardTapped(g, island); !tapped {
		t.Error("the Island untapped during its controller's next untap step")
	}
	// The one after: it untaps again.
	advanceToUpkeepOfSeat(t, g, 0)
	advanceToUpkeepOfSeat(t, g, 1)
	if tapped, _ := battlefieldCardTapped(g, island); tapped {
		t.Error("the freeze lasts for one untap step only")
	}
}

// Vorinclex's own controller's lands are never frozen.
func TestVorinclexDoesNotFreezeYourOwnLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Vorinclex, Voice of Hunger", "Legendary Creature — Phyrexian Praetor", vorinclexVoiceOracle, false)
	forest := b31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	tapForMana(t, g, me.ID, forest)
	passPriorityAroundTable(t, g)
	advanceToUpkeepOfSeat(t, g, 0)
	if tapped, _ := battlefieldCardTapped(g, forest); tapped {
		t.Error("your own land untaps normally")
	}
}

// battlefieldCardTapped reads the live Tapped flag.
func battlefieldCardTapped(g *game.Game, id uuid.UUID) (bool, bool) {
	c, ok := battlefieldCard(g, id)
	return c.Tapped, ok
}
