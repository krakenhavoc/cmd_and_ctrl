package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// the_ring_cards_test.go — ADR 0114 PR 3: the five cards of #2062's
// deck that tempt.

const (
	callOfTheRingOracle         = "9fcb920c-b8d7-4a79-a335-f63050182cca"
	nazgulOracle                = "48a62778-7c11-486f-a0e1-020c283a7ef9"
	ringsightOracle             = "66e42be3-5125-41b4-9efe-8a6a66899bb6"
	sauronTheDarkLordOracle     = "9f8f29ba-7e64-455d-be05-ec7973f2bd0a"
	sauronLordOfTheRingsOracle  = "69c674c7-48a5-49c8-b0be-3f2b5c6a548c"
	sauronLordOfTheRingsTypeStr = "Legendary Creature — Avatar Horror"
)

// castRingCreature puts a creature card with a size into the active
// seat's hand and casts it from a main phase.
func castRingCreature(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// armiesOf is every Army `player` controls, with its +1/+1 counters.
func armiesOf(g *game.Game, player uuid.UUID) map[uuid.UUID]int {
	out := map[uuid.UUID]int{}
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	for _, c := range g.Battlefield.Cards {
		if c.Controller == player && c.HasSubtype("Army") && c.HasSubtype("Orc") {
			out[c.InstanceID] = c.Counters["+1/+1"]
		}
	}
	return out
}

// --- Call of the Ring --------------------------------------------

// "At the beginning of your upkeep, the Ring tempts you."
func TestCallOfTheRingTemptsAtYourUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	pushCatalogPermanent(g, opp.ID, "Call of the Ring", "Enchantment", callOfTheRingOracle, false)
	advanceToUpkeepOf(t, g, 1)
	ringSettle(t, g)
	if got := ringCount(g, opp.ID); got != 1 {
		t.Fatalf("after its controller's upkeep the Ring has tempted them %d times, want 1", got)
	}
	if ringCount(g, g.Seats[0].ID) != 0 {
		t.Fatal("Call of the Ring tempted a player who does not control it")
	}
}

// "Whenever you choose a creature as your Ring-bearer, you may pay 2
// life. If you do, draw a card." Not with no creature to choose; again
// on re-choosing the same creature (2023-06-16 rulings).
func TestCallOfTheRingDrawsWhenYouChooseARingBearer(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Call of the Ring", "Enchantment", callOfTheRingOracle, false)

	ringTempt(t, g, me.ID)
	ringSettle(t, g)
	if latestConfirmFor(g, me.ID) != nil {
		t.Fatal("Call of the Ring triggered with no creature to choose")
	}

	b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	life, hand := me.Life, me.Hand.Size()
	ringTempt(t, g, me.ID)
	ringSettle(t, g)
	if c := latestConfirmFor(g, me.ID); c == nil || c.LifeCost != 2 {
		t.Fatalf("no pay-2-life prompt after choosing a Ring-bearer: %+v", c)
	}
	answerMayChoice(t, g, me.ID, true)
	ringSettle(t, g)
	if me.Life != life-2 || me.Hand.Size() != hand+1 {
		t.Fatalf("paid: life %d → %d, hand %d → %d; want -2 and +1", life, me.Life, hand, me.Hand.Size())
	}

	// Re-choosing the same creature triggers it again. Declining costs
	// nothing.
	life, hand = me.Life, me.Hand.Size()
	ringTempt(t, g, me.ID)
	ringSettle(t, g)
	if latestConfirmFor(g, me.ID) == nil {
		t.Fatal("re-choosing the Ring-bearer did not trigger Call of the Ring")
	}
	answerMayChoice(t, g, me.ID, false)
	ringSettle(t, g)
	if me.Life != life || me.Hand.Size() != hand {
		t.Fatal("declining paid life or drew a card")
	}
}

// --- Nazgûl ------------------------------------------------------

// "When this creature enters, the Ring tempts you. Whenever the Ring
// tempts you, put a +1/+1 counter on each Wraith you control." Only
// Wraiths, only yours.
func TestNazgulTemptsAndGrowsEachWraith(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wraith := b12Creature(g, me.ID, "Barrow-Wight", "Creature — Wraith", 2, 2)
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Wraith", "Creature — Wraith", 2, 2)

	nazgul := castRingCreature(t, g, "Nazgûl", "Creature — Wraith Knight", nazgulOracle, 1, 2)
	ringSettle(t, g)
	if ringPrompt(g, me.ID) == nil {
		t.Fatal("Nazgûl entered and the Ring did not ask for a Ring-bearer among three creatures")
	}
	answerRing(t, g, me.ID, nazgul)
	ringSettle(t, g)

	if ringBearerOf(g, me.ID) != nazgul || ringCount(g, me.ID) != 1 {
		t.Fatal("Nazgûl is not the Ring-bearer after one temptation")
	}
	for id, want := range map[uuid.UUID]int{nazgul: 1, wraith: 1, bear: 0, theirs: 0} {
		if got := ringBFCard(t, g, id).Counters["+1/+1"]; got != want {
			t.Errorf("%s has %d +1/+1 counters, want %d", ringBFCard(t, g, id).Name, got, want)
		}
	}
	if !game.HasKeyword(ringBFCard(t, g, nazgul), "deathtouch") {
		t.Error("Nazgûl has no deathtouch")
	}
}

// --- Ringsight ---------------------------------------------------

// The search reads the board the tempt left: the Ring-bearer just
// chosen is legendary, so its color counts, and a colorless legendary
// creature adds nothing.
func TestRingsightSearchesForTheNewBearersColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	elf := b16Creature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1, "G")
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Colorless Legend", TypeLine: "Legendary Artifact Creature — Golem",
		ManaCost: "{4}", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	pushLibraryCardForTest(me, game.Card{Name: "Green Card", TypeLine: "Sorcery", ManaCost: "{G}", Colors: []string{"G"}})
	pushLibraryCardForTest(me, game.Card{Name: "Gold Card", TypeLine: "Sorcery", ManaCost: "{R}{G}", Colors: []string{"R", "G"}})
	pushLibraryCardForTest(me, game.Card{Name: "Red Card", TypeLine: "Sorcery", ManaCost: "{R}", Colors: []string{"R"}})
	pushLibraryCardForTest(me, game.Card{Name: "Colorless Card", TypeLine: "Artifact", ManaCost: "{2}"})

	castCatalogSpell(t, g, "Ringsight", "Sorcery", ringsightOracle, nil)
	passPriorityAroundTable(t, g)
	answerRing(t, g, me.ID, elf)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search after the tempt")
	}
	for _, name := range []string{"Green Card", "Gold Card"} {
		if searchOptionNamed(g, c, name) == uuid.Nil {
			t.Errorf("%s shares the new Ring-bearer's color and was not offered", name)
		}
	}
	for _, name := range []string{"Red Card", "Colorless Card"} {
		if searchOptionNamed(g, c, name) != uuid.Nil {
			t.Errorf("%s was offered; it shares no color with a legendary creature", name)
		}
	}
	answerSearchNamed(t, g, me.ID, "Green Card")
	ringSettle(t, g)
	found := false
	for _, h := range me.Hand.Cards {
		if h.Name == "Green Card" {
			found = true
		}
	}
	if !found {
		t.Fatal("the card found is not in hand")
	}
}

// --- Sauron, the Dark Lord ---------------------------------------

// "Whenever an opponent casts a spell, amass Orcs 1."
func TestSauronTheDarkLordAmassesWhenAnOpponentCasts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Sauron, the Dark Lord", "Legendary Creature — Avatar Horror", sauronTheDarkLordOracle, 7, 6)
	batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	ringSettle(t, g)
	armies := armiesOf(g, me.ID)
	if len(armies) != 1 {
		t.Fatalf("%d Orc Armies after an opponent's spell, want 1", len(armies))
	}
	for _, n := range armies {
		if n != 1 {
			t.Fatalf("the Army has %d counters, want 1", n)
		}
	}
}

// "Whenever an Army you control deals combat damage to a player, the
// Ring tempts you. Whenever the Ring tempts you, you may discard your
// hand. If you do, draw four cards." A non-Army does not tempt.
func TestSauronTheDarkLordArmiesTemptAndTheTemptWheels(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceTo(t, g, game.StepPrecombatMain)
	sauron := b12Push(g, me.ID, "Sauron, the Dark Lord", "Legendary Creature — Avatar Horror", sauronTheDarkLordOracle, 7, 6)
	army := b12Creature(g, me.ID, "Orc Army", "Creature — Orc Army", 2, 2)
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)

	dealCombatDamageToPlayer(g, bear, opp.ID, 2)
	ringSettle(t, g)
	if ringCount(g, me.ID) != 0 {
		t.Fatal("a non-Army's combat damage tempted")
	}

	dealCombatDamageToPlayer(g, army, opp.ID, 2)
	ringSettle(t, g)
	answerRing(t, g, me.ID, sauron)
	ringSettle(t, g)
	if ringCount(g, me.ID) != 1 {
		t.Fatalf("the Army's combat damage tempted %d times, want 1", ringCount(g, me.ID))
	}
	old := append([]game.Card(nil), me.Hand.Cards...)
	if len(old) == 0 {
		t.Fatal("setup: empty hand")
	}
	answerMayChoice(t, g, me.ID, true)
	ringSettle(t, g)
	if me.Hand.Size() != 4 {
		t.Fatalf("hand %d after discarding it and drawing four, want 4", me.Hand.Size())
	}
	for _, c := range old {
		if !me.Graveyard.Contains(c.InstanceID) {
			t.Fatalf("%s was not discarded", c.Name)
		}
	}
}

// "Ward—Sacrifice a legendary artifact or legendary creature." The
// caster's own Ring-bearer is legendary through the Ring, so it pays; a
// plain creature and a plain artifact do not.
func TestSauronTheDarkLordWardTakesALegendIncludingARingBearer(t *testing.T) {
	g, _, opp, sauron := wardTable(t, "Sauron, the Dark Lord", "Legendary Creature — Avatar Horror", sauronTheDarkLordOracle)
	bearer := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	ringTempt(t, g, opp.ID)
	if ringBearerOf(g, opp.ID) != bearer {
		t.Fatal("setup: the Bear is not their Ring-bearer")
	}
	b16Creature(g, opp.ID, "Plain Bear", "Creature — Bear", 2, 2, "G")
	relic := b12Permanent(g, opp.ID, "Legendary Relic", "Legendary Artifact")
	b12Permanent(g, opp.ID, "Plain Rock", "Artifact")

	castAtWardedCreature(t, g, opp, sauron)
	// The spell also triggers Sauron's amass: settle the ordering, and
	// stop at the ward's prompt.
	ringSettle(t, g)
	prompt := confirmChoiceFor(g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt")
	}
	if err := g.ResolveConfirm(prompt.ID, opp.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	pick := chooseCardsChoiceFor(g, opp.ID)
	if pick == nil {
		t.Fatal("accepting the ward queued no sacrifice pick")
	}
	got := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		got[id] = true
	}
	if len(got) != 2 || !got[bearer] || !got[relic] {
		t.Fatalf("candidates %v, want exactly the Ring-bearer and the legendary artifact", pick.ChooseCards)
	}
}

// --- Sauron, Lord of the Rings -----------------------------------

// "When you cast this spell, amass Orcs 5, mill five cards, then return
// a creature card from your graveyard to the battlefield." The card
// chosen need not be one just milled (2023-06-16 ruling).
func TestSauronLordOfTheRingsCastTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	old := batch01GraveyardCard(me, "Old Bear", "Creature — Bear")
	milled := seedLibraryTop(me, "Milled Bear", "Creature — Bear", "{1}{G}")
	library := me.Library.Size()

	sauron := castRingCreature(t, g, "Sauron, Lord of the Rings", sauronLordOfTheRingsTypeStr, sauronLordOfTheRingsOracle, 9, 9)
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no creature-card choice after the mill")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		offered[id] = true
	}
	if !offered[old] || !offered[milled] {
		t.Fatalf("offered %v, want the old creature card and the milled one", pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, old)
	ringSettle(t, g)

	if me.Library.Size() != library-5 {
		t.Errorf("library %d → %d, want five milled", library, me.Library.Size())
	}
	armies := armiesOf(g, me.ID)
	if len(armies) != 1 {
		t.Fatalf("%d Orc Armies, want 1", len(armies))
	}
	for _, n := range armies {
		if n != 5 {
			t.Errorf("the Army has %d counters, want 5", n)
		}
	}
	if !ringOnBattlefield(g, old) || ringOnBattlefield(g, milled) {
		t.Error("the chosen creature card did not return, or the other one did")
	}
	if !ringOnBattlefield(g, sauron) {
		t.Error("Sauron did not resolve")
	}
}

// "Whenever a commander an opponent controls dies, the Ring tempts
// you." An opponent's commander whose owner lets it go to the graveyard
// triggers it; your own does not. One whose owner takes the command
// zone does not either — the card's caveat, until #2085.
func TestSauronLordOfTheRingsTemptsWhenAnOpponentsCommanderDies(t *testing.T) {
	for _, tc := range []struct {
		name        string
		theirs      bool
		commandZone bool
		want        int
	}{
		{"opponent's commander to the graveyard", true, false, 1},
		{"opponent's commander to the command zone (caveat)", true, true, 0},
		{"my own commander to the graveyard", false, false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			b12Push(g, me.ID, "Sauron, Lord of the Rings", sauronLordOfTheRingsTypeStr, sauronLordOfTheRingsOracle, 9, 9)
			owner := me
			if tc.theirs {
				owner = opp
			}
			commander := pushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: "A Commander", TypeLine: "Legendary Creature — Human",
				Power: 2, Toughness: 2, Owner: owner.ID, Controller: owner.ID, IsCommander: true,
			})
			g.WithWriteLock(func() {
				if err := (DestroyTarget{Target: commander}).Apply(ctxFor(g, &game.StackItem{Controller: me.ID, SourceCardID: uuid.New()})); err != nil {
					t.Fatalf("destroy: %v", err)
				}
			})
			offer := latestChoiceOfKind(g, game.PendingChoiceOptionalReplacement)
			if offer == nil {
				t.Fatal("setup: the commander's owner was not offered the command zone")
			}
			if err := g.ResolveOptionalReplacement(offer.ID, owner.ID, tc.commandZone); err != nil {
				t.Fatalf("ResolveOptionalReplacement: %v", err)
			}
			ringSettle(t, g)
			if got := ringCount(g, me.ID); got != tc.want {
				t.Fatalf("the Ring tempted Sauron's controller %d times, want %d", got, tc.want)
			}
		})
	}
}
