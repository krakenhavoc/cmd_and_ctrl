package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// jodi_spells_and_walker_test.go — S58 PR 5: Ulamog, the mana doublers,
// Ugin's Binding, Sunbird's Invocation, Emergent Ultimatum and Nicol
// Bolas. See jodi_big_spells_test.go for the oracle constants.

// --- Ulamog, the Ceaseless Hunger -------------------------------------

func TestUlamogCastTriggerExilesTwoTargetPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.PushCatalogPermanent(g, opp.ID, "Bear A", "Creature — Bear", "", false)
	b := effectstest.PushCatalogPermanent(g, opp.ID, "Bear B", "Creature — Bear", "", false)
	keep := effectstest.PushCatalogPermanent(g, opp.ID, "Bear C", "Creature — Bear", "", false)
	id := effectstest.PutInHand(me, game.Card{Name: "Ulamog, the Ceaseless Hunger", TypeLine: "Legendary Creature — Eldrazi",
		OracleID: effectstest.UlamogCeaselessOracle, ManaCost: "{10}", Power: 10, Toughness: 10})
	effectstest.ToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell Ulamog: %v", err)
	}
	p := effectstest.LatestPickTarget(g, me.ID)
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
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.InExile(g, a) || !effectstest.InExile(g, b) {
		t.Errorf("both targets should be exiled (a=%v b=%v)", effectstest.InExile(g, a), effectstest.InExile(g, b))
	}
	if !g.Battlefield.Contains(keep) {
		t.Error("an untargeted permanent is untouched")
	}
	card, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		t.Fatal("Ulamog resolves after its cast trigger")
	}
	if !game.HasKeyword(&card, "indestructible") {
		t.Error("Ulamog is indestructible")
	}
}

func TestUlamogAttackExilesTwentyFromTheDefendersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	third := g.Seats[2]
	for i := 0; i < 30; i++ {
		opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Land", Owner: opp.ID, Controller: opp.ID})
	}
	ulamog := effectstest.PushCatalogPermanent(g, me.ID, "Ulamog, the Ceaseless Hunger", "Legendary Creature — Eldrazi", effectstest.UlamogCeaselessOracle, false)
	libBefore, thirdBefore := opp.Library.Size(), third.Library.Size()
	effectstest.DeclareAttack(t, g, opp.ID, ulamog)
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Zendikar Resurgent", "Enchantment", effectstest.ZendikarResurgentOracle, false)
	forest := effectstest.B31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	effectstest.TapForMana(t, g, me.ID, forest)
	if got := effectstest.SortedPool(me); len(got) != 2 || got[0] != "G" || got[1] != "G" {
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
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.PushCatalogPermanent(g, me.ID, "Zendikar Resurgent", "Enchantment", effectstest.ZendikarResurgentOracle, false)
			before := me.Hand.Size()
			effectstest.CastCatalogSpell(t, g, "A Spell", tc.typeLine, "", nil)
			effectstest.PassPriorityAroundTable(t, g)
			if got := me.Hand.Size(); got != before+tc.draws {
				t.Errorf("hand %d -> %d, want +%d", before, got, tc.draws)
			}
		})
	}
}

func TestVorinclexDoublesYourManaButNotAnOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Vorinclex, Voice of Hunger", "Legendary Creature — Phyrexian Praetor", effectstest.VorinclexVoiceOracle, false)
	mine := effectstest.B31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	theirs := effectstest.B31Push(g, opp.ID, "Island", "Basic Land — Island", "", "", 0, 0)
	effectstest.TapForMana(t, g, me.ID, mine)
	if got := effectstest.SortedPool(me); len(got) != 2 {
		t.Errorf("my pool = %v, want two mana", got)
	}
	effectstest.TapForMana(t, g, opp.ID, theirs)
	if got := effectstest.SortedPool(opp); len(got) != 1 {
		t.Errorf("an opponent's land is not doubled: pool = %v", got)
	}
}

// An opponent's tapped land misses its controller's next untap step —
// and only that one.
func TestVorinclexFreezesAnOpponentsLandThatTapsForMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Vorinclex, Voice of Hunger", "Legendary Creature — Phyrexian Praetor", effectstest.VorinclexVoiceOracle, false)
	island := effectstest.B31Push(g, opp.ID, "Island", "Basic Land — Island", "", "", 0, 0)
	effectstest.TapForMana(t, g, opp.ID, island)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, island); !tapped {
		t.Fatal("the Island should be tapped")
	}
	effectstest.PassPriorityAroundTable(t, g)

	// The opponent's next untap step: the Island stays tapped.
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, island); !tapped {
		t.Error("the Island untapped during its controller's next untap step")
	}
	// The one after: it untaps again.
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, island); tapped {
		t.Error("the freeze lasts for one untap step only")
	}
}

// Vorinclex's own controller's lands are never frozen.
func TestVorinclexDoesNotFreezeYourOwnLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Vorinclex, Voice of Hunger", "Legendary Creature — Phyrexian Praetor", effectstest.VorinclexVoiceOracle, false)
	forest := effectstest.B31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	effectstest.TapForMana(t, g, me.ID, forest)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, forest); tapped {
		t.Error("your own land untaps normally")
	}
}
