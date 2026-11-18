package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	valiantEndeavorOracle  = "4ba1f2c5-6456-4809-a563-fd5bd311869a"
	graveEndeavorOracle    = "917dc5a0-071d-4658-ab58-29ac841260f7"
	wildEndeavorOracle     = "2c4ee0df-e720-475b-aa04-e6394f3d708f"
	arcaneEndeavorOracle   = "901ee0e3-1011-40e1-815a-09246112e8dc"
	brazenDwarfOracle      = "256ab1e2-0c6a-41c1-ba01-021f3c90cfa2"
	feywildTricksterOracle = "b385edc3-c897-48e3-b251-242c77159fa0"
	vrondissOracle         = "2160c837-342b-409f-a1db-0ec5ac023ccf"
	everfullPurseOracle    = "f70b9eee-3874-4391-8df8-e10bbd19307e"
)

// afrChosen answers an Endeavor's "choose one result" (a yes is the first
// roll as the chosen result) and returns the chosen and the other result.
// Equal rolls are no choice and nothing is asked.
func afrChosen(t *testing.T, g *game.Game, me uuid.UUID, sides int, accept bool) (chosen, other int) {
	t.Helper()
	rolls := randomEvents(g, game.EventRollDie)
	if len(rolls) != 2 || rolls[0].Sides != sides || rolls[0].BatchSeq != rolls[1].BatchSeq {
		t.Fatalf("rolls: %+v", rolls)
	}
	chosen, other = rolls[0].Amount, rolls[1].Amount
	if chosen == other {
		for _, c := range g.PendingChoices {
			if c.Kind == game.PendingChoiceConfirm {
				t.Fatal("equal rolls asked which result to choose")
			}
		}
		return chosen, other
	}
	if len(g.PendingChoices) != 1 {
		t.Fatalf("unequal rolls %d/%d asked %d questions, want 1", chosen, other, len(g.PendingChoices))
	}
	if err := g.ResolveConfirm(g.PendingChoices[0].ID, me, accept); err != nil {
		t.Fatal(err)
	}
	if !accept {
		chosen, other = other, chosen
	}
	return chosen, other
}

func TestValiantEndeavorDestroysByPowerThenKnights(t *testing.T) {
	for _, accept := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		var ids []uuid.UUID
		for p := 1; p <= 6; p++ {
			ids = append(ids, b12Push(g, g.Seats[1].ID, "Body", "Creature", "", p, 9))
		}
		castCatalogSpell(t, g, "Valiant Endeavor", "Sorcery", valiantEndeavorOracle, nil)
		passPriorityAroundTable(t, g)
		chosen, other := afrChosen(t, g, me.ID, 6, accept)
		g.RunStateChecksForTest()
		for p := 1; p <= 6; p++ {
			if alive := g.Battlefield.Contains(ids[p-1]); alive != (p < chosen) {
				t.Errorf("accept=%v power %d alive=%v with chosen result %d", accept, p, alive, chosen)
			}
		}
		if got := randomTokenCount(g, "Knight"); got != other {
			t.Errorf("accept=%v Knights=%d, want %d", accept, got, other)
		}
		for _, c := range g.Battlefield.Cards {
			if c.IsToken() && c.Name == "Knight" && !hasEffectiveKeyword(t, g, c.InstanceID, "vigilance") {
				t.Error("a Knight lacks vigilance")
			}
		}
	}
}

func TestWildEndeavorBeastsThenBasics(t *testing.T) {
	for _, accept := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		var forests []uuid.UUID
		for i := 0; i < 5; i++ {
			c := game.NewCard("Forest", me.ID)
			c.TypeLine = "Basic Land — Forest"
			me.Library.PushBottom(c)
			forests = append(forests, c.InstanceID)
		}
		castCatalogSpell(t, g, "Wild Endeavor", "Sorcery", wildEndeavorOracle, nil)
		passPriorityAroundTable(t, g)
		chosen, other := afrChosen(t, g, me.ID, 4, accept)
		if got := randomTokenCount(g, "Beast"); got != chosen {
			t.Errorf("accept=%v Beasts=%d, want %d", accept, got, chosen)
		}
		search := latestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, me.ID)
		if search == nil {
			t.Fatal("no library search")
		}
		picks := forests
		if len(picks) > other {
			picks = picks[:other]
		}
		if err := g.ResolveSearchLibrary(search.ID, me.ID, picks); err != nil {
			t.Fatal(err)
		}
		lands := 0
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Forest" {
				lands++
				if !c.Tapped {
					t.Error("a fetched land entered untapped")
				}
			}
		}
		if lands != other {
			t.Errorf("accept=%v lands=%d, want %d", accept, lands, other)
		}
	}
}

func TestGraveEndeavorCountersThenDrain(t *testing.T) {
	for _, accept := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		body := pushGraveyardPermanent(me, "Corpse", "Creature — Zombie", "")
		pushGraveyardPermanent(me, "Relic", "Artifact", "")
		life, oppLife := me.Life, opp.Life
		castCatalogSpell(t, g, "Grave Endeavor", "Instant", graveEndeavorOracle, nil)
		passPriorityAroundTable(t, g)
		chosen, other := afrChosen(t, g, me.ID, 10, accept)
		pick := chooseCardsChoiceFor(g, me.ID)
		if pick == nil {
			t.Fatal("no graveyard pick")
		}
		if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != body {
			t.Fatalf("candidates %v, want only the creature card", pick.ChooseCards)
		}
		answerChooseCards(t, g, me.ID, body)
		if got := counterCount(g, body, game.CounterPlusOne); got != chosen {
			t.Errorf("accept=%v counters=%d, want %d", accept, got, chosen)
		}
		if me.Life != life+other || opp.Life != oppLife-other {
			t.Errorf("accept=%v life %d/%d, want %d/%d", accept, me.Life, opp.Life, life+other, oppLife-other)
		}
	}
}

func TestGraveEndeavorDrainsWithNoCreatureInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life, oppLife := me.Life, opp.Life
	castCatalogSpell(t, g, "Grave Endeavor", "Instant", graveEndeavorOracle, nil)
	passPriorityAroundTable(t, g)
	_, other := afrChosen(t, g, me.ID, 10, true)
	if me.Life != life+other || opp.Life != oppLife-other {
		t.Errorf("life %d/%d, want %d/%d", me.Life, opp.Life, life+other, oppLife-other)
	}
}

func TestArcaneEndeavorDrawsThenOffersAFreeSpell(t *testing.T) {
	for _, accept := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		for i := 0; i < 10; i++ {
			me.Library.PushBottom(game.NewCard("filler", me.ID))
		}
		cheap := handCardForTest(me, "Cheap Instant", "Instant", "")
		sorcery := handCardForTest(me, "Cheap Sorcery", "Sorcery", "")
		creature := handCardForTest(me, "Cheap Creature", "Creature — Bear", "")
		big := handCardForTest(me, "Big Instant", "Instant", "")
		for id, mc := range map[uuid.UUID]string{cheap: "{R}", sorcery: "{1}{R}", creature: "{R}", big: "{9}"} {
			for i := range me.Hand.Cards {
				if me.Hand.Cards[i].InstanceID == id {
					me.Hand.Cards[i].ManaCost = mc
				}
			}
		}
		hand := me.Hand.Size()
		castCatalogSpell(t, g, "Arcane Endeavor", "Sorcery", arcaneEndeavorOracle, nil)
		passPriorityAroundTable(t, g)
		chosen, other := afrChosen(t, g, me.ID, 8, accept)
		if got := me.Hand.Size(); got != hand+chosen {
			t.Errorf("accept=%v hand=%d, want %d (drew %d)", accept, got, hand+chosen, chosen)
		}
		var offer *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceResolutionCast && c.Chooser == me.ID {
				offer = c
			}
		}
		if offer == nil {
			t.Fatalf("accept=%v: no free-cast offer (other result %d)", accept, other)
		}
		has := map[uuid.UUID]bool{}
		for _, id := range offer.DiscardOptions {
			has[id] = true
		}
		if !has[cheap] || has[sorcery] != (other >= 2) || has[creature] || has[big] {
			t.Errorf("accept=%v other=%d offered %v: want the mana value 1 instant, the mana value 2 sorcery only if the other result is 2+, never the creature or the mana value 9 instant", accept, other, has)
		}
	}
}

func TestBrazenDwarfPingsOncePerRollInstruction(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dwarf := pushCatalogPermanent(g, me.ID, "Brazen Dwarf", "Creature — Dwarf Shaman", brazenDwarfOracle, false)
	life := opp.Life
	d5681Roll(t, g, me.ID, dwarf, 6, 3, 0)
	if opp.Life != life-1 {
		t.Errorf("three dice in one instruction: life %d, want %d", opp.Life, life-1)
	}
	d5681Roll(t, g, opp.ID, uuid.New(), 6, 1, 0)
	if opp.Life != life-1 {
		t.Errorf("an opponent's roll pinged: life %d, want %d", opp.Life, life-1)
	}
	d5681Roll(t, g, me.ID, dwarf, 20, 1, 0)
	if opp.Life != life-2 {
		t.Errorf("a second instruction: life %d, want %d", opp.Life, life-2)
	}
}

func TestFeywildTricksterMakesAFaerieDragonPerInstruction(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	src := pushCatalogPermanent(g, me.ID, "Feywild Trickster", "Creature — Gnome Warlock", feywildTricksterOracle, false)
	d5681Roll(t, g, me.ID, src, 6, 2, 0)
	if got := randomTokenCount(g, "Faerie Dragon"); got != 1 {
		t.Fatalf("Faerie Dragons=%d after two dice in one instruction, want 1", got)
	}
	d5681Roll(t, g, g.Seats[1].ID, uuid.New(), 6, 1, 0)
	if got := randomTokenCount(g, "Faerie Dragon"); got != 1 {
		t.Errorf("an opponent's roll made one: %d", got)
	}
	d5681Roll(t, g, me.ID, src, 6, 1, 0)
	if got := randomTokenCount(g, "Faerie Dragon"); got != 2 {
		t.Errorf("Faerie Dragons=%d after a second instruction, want 2", got)
	}
}

func TestVrondissEnrageAndSelfDamage(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	v := b27Push(g, me.ID, "Vrondiss, Rage of Ancients", "Legendary Creature — Dragon Barbarian", vrondissOracle, "{3}{R}{G}", 5, 4, "R", "G")
	theirs := b16Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2, "G")
	cmdr2B08Damage(g, theirs, v, 1)
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 1 {
		t.Fatalf("enrage asked %d questions", len(g.PendingChoices))
	}
	if err := g.ResolveConfirm(g.PendingChoices[0].ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	var spirit uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Dragon Spirit" {
			spirit = c.InstanceID
			if c.Power != 5 || c.Toughness != 4 || len(c.Colors) != 2 {
				t.Errorf("token is %d/%d %v, want 5/4 red and green", c.Power, c.Toughness, c.Colors)
			}
		}
	}
	if spirit == uuid.Nil {
		t.Fatal("no Dragon Spirit")
	}
	// The token sacrifices itself when it deals damage.
	b27Damage(g, spirit, theirs, 1)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(spirit) {
		t.Error("the Dragon Spirit survived dealing damage")
	}

	// A roll offers the self-damage; declining does nothing, accepting
	// deals 1 to Vrondiss (which enrages, a second question).
	d5681Roll(t, g, me.ID, v, 6, 2, 0)
	if len(g.PendingChoices) != 1 {
		t.Fatalf("a roll asked %d questions, want 1", len(g.PendingChoices))
	}
	if err := g.ResolveConfirm(g.PendingChoices[0].ID, me.ID, false); err != nil {
		t.Fatal(err)
	}
	if c, _ := g.LookupCardForEffect(v); c.DamageMarked != 1 {
		t.Errorf("declined, yet damage %d beyond the enrage hit", c.DamageMarked)
	}
	d5681Roll(t, g, me.ID, v, 6, 1, 0)
	if len(g.PendingChoices) != 1 {
		t.Fatalf("a second roll asked %d questions, want 1", len(g.PendingChoices))
	}
	if err := g.ResolveConfirm(g.PendingChoices[0].ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	if c, _ := g.LookupCardForEffect(v); c.DamageMarked != 2 {
		t.Errorf("accepted: damage %d, want 2 (the enrage hit and 1 more)", c.DamageMarked)
	}
}

func TestBucknardsEverfullPurseTreasuresThenPassesRight(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	purse := pushCatalogPermanent(g, me.ID, "Bucknard's Everfull Purse", "Artifact", everfullPurseOracle, false)
	right := g.NearestOpponentForEffect(me.ID, game.AttackDirectionRight)
	if err := g.ActivateCatalogAbility(me.ID, purse, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	rolls := randomEvents(g, game.EventRollDie)
	if len(rolls) != 1 || rolls[0].Sides != 4 {
		t.Fatalf("rolls: %+v", rolls)
	}
	if got := randomTokenCount(g, "Treasure"); got != rolls[0].Amount {
		t.Errorf("Treasures=%d, want %d", got, rolls[0].Amount)
	}
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Treasure" && c.Controller != me.ID {
			t.Error("the Treasures went to the wrong player")
		}
		if c.InstanceID == purse && c.Controller != right {
			t.Errorf("Purse controlled by %v, want the player to the right %v", c.Controller, right)
		}
	}
}
