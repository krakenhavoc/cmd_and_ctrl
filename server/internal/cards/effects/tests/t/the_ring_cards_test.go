package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.PushCatalogPermanent(g, opp.ID, "Call of the Ring", "Enchantment", effectstest.CallOfTheRingOracle, false)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.RingSettle(t, g)
	if got := effectstest.RingCount(g, opp.ID); got != 1 {
		t.Fatalf("after its controller's upkeep the Ring has tempted them %d times, want 1", got)
	}
	if effectstest.RingCount(g, g.Seats[0].ID) != 0 {
		t.Fatal("Call of the Ring tempted a player who does not control it")
	}
}

// "Whenever you choose a creature as your Ring-bearer, you may pay 2
// life. If you do, draw a card." Not with no creature to choose; again
// on re-choosing the same creature (2023-06-16 rulings).
func TestCallOfTheRingDrawsWhenYouChooseARingBearer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Call of the Ring", "Enchantment", effectstest.CallOfTheRingOracle, false)

	effectstest.RingTempt(t, g, me.ID)
	effectstest.RingSettle(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Fatal("Call of the Ring triggered with no creature to choose")
	}

	effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	life, hand := me.Life, me.Hand.Size()
	effectstest.RingTempt(t, g, me.ID)
	effectstest.RingSettle(t, g)
	if c := effectstest.LatestConfirmFor(g, me.ID); c == nil || c.LifeCost != 2 {
		t.Fatalf("no pay-2-life prompt after choosing a Ring-bearer: %+v", c)
	}
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.RingSettle(t, g)
	if me.Life != life-2 || me.Hand.Size() != hand+1 {
		t.Fatalf("paid: life %d → %d, hand %d → %d; want -2 and +1", life, me.Life, hand, me.Hand.Size())
	}

	// Re-choosing the same creature triggers it again. Declining costs
	// nothing.
	life, hand = me.Life, me.Hand.Size()
	effectstest.RingTempt(t, g, me.ID)
	effectstest.RingSettle(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) == nil {
		t.Fatal("re-choosing the Ring-bearer did not trigger Call of the Ring")
	}
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.RingSettle(t, g)
	if me.Life != life || me.Hand.Size() != hand {
		t.Fatal("declining paid life or drew a card")
	}
}

// --- Nazgûl ------------------------------------------------------

// "When this creature enters, the Ring tempts you. Whenever the Ring
// tempts you, put a +1/+1 counter on each Wraith you control." Only
// Wraiths, only yours.
func TestNazgulTemptsAndGrowsEachWraith(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wraith := effectstest.B12Creature(g, me.ID, "Barrow-Wight", "Creature — Wraith", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Wraith", "Creature — Wraith", 2, 2)

	nazgul := effectstest.CastRingCreature(t, g, "Nazgûl", "Creature — Wraith Knight", effectstest.NazgulOracle, 1, 2)
	effectstest.RingSettle(t, g)
	if effectstest.RingPrompt(g, me.ID) == nil {
		t.Fatal("Nazgûl entered and the Ring did not ask for a Ring-bearer among three creatures")
	}
	effectstest.AnswerRing(t, g, me.ID, nazgul)
	effectstest.RingSettle(t, g)

	if effectstest.RingBearerOf(g, me.ID) != nazgul || effectstest.RingCount(g, me.ID) != 1 {
		t.Fatal("Nazgûl is not the Ring-bearer after one temptation")
	}
	for id, want := range map[uuid.UUID]int{nazgul: 1, wraith: 1, bear: 0, theirs: 0} {
		if got := effectstest.RingBFCard(t, g, id).Counters["+1/+1"]; got != want {
			t.Errorf("%s has %d +1/+1 counters, want %d", effectstest.RingBFCard(t, g, id).Name, got, want)
		}
	}
	if !game.HasKeyword(effectstest.RingBFCard(t, g, nazgul), "deathtouch") {
		t.Error("Nazgûl has no deathtouch")
	}
}

// --- Ringsight ---------------------------------------------------

// The search reads the board the tempt left: the Ring-bearer just
// chosen is legendary, so its color counts, and a colorless legendary
// creature adds nothing.
func TestRingsightSearchesForTheNewBearersColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	elf := effectstest.B16Creature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1, "G")
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Colorless Legend", TypeLine: "Legendary Artifact Creature — Golem",
		ManaCost: "{4}", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Green Card", TypeLine: "Sorcery", ManaCost: "{G}", Colors: []string{"G"}})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Gold Card", TypeLine: "Sorcery", ManaCost: "{R}{G}", Colors: []string{"R", "G"}})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Red Card", TypeLine: "Sorcery", ManaCost: "{R}", Colors: []string{"R"}})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Colorless Card", TypeLine: "Artifact", ManaCost: "{2}"})

	effectstest.CastCatalogSpell(t, g, "Ringsight", "Sorcery", effectstest.RingsightOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerRing(t, g, me.ID, elf)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search after the tempt")
	}
	for _, name := range []string{"Green Card", "Gold Card"} {
		if effectstest.SearchOptionNamed(g, c, name) == uuid.Nil {
			t.Errorf("%s shares the new Ring-bearer's color and was not offered", name)
		}
	}
	for _, name := range []string{"Red Card", "Colorless Card"} {
		if effectstest.SearchOptionNamed(g, c, name) != uuid.Nil {
			t.Errorf("%s was offered; it shares no color with a legendary creature", name)
		}
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Green Card")
	effectstest.RingSettle(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Sauron, the Dark Lord", "Legendary Creature — Avatar Horror", effectstest.SauronTheDarkLordOracle, 7, 6)
	effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.RingSettle(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	sauron := effectstest.B12Push(g, me.ID, "Sauron, the Dark Lord", "Legendary Creature — Avatar Horror", effectstest.SauronTheDarkLordOracle, 7, 6)
	army := effectstest.B12Creature(g, me.ID, "Orc Army", "Creature — Orc Army", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)

	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 2)
	effectstest.RingSettle(t, g)
	if effectstest.RingCount(g, me.ID) != 0 {
		t.Fatal("a non-Army's combat damage tempted")
	}

	effectstest.DealCombatDamageToPlayer(g, army, opp.ID, 2)
	effectstest.RingSettle(t, g)
	effectstest.AnswerRing(t, g, me.ID, sauron)
	effectstest.RingSettle(t, g)
	if effectstest.RingCount(g, me.ID) != 1 {
		t.Fatalf("the Army's combat damage tempted %d times, want 1", effectstest.RingCount(g, me.ID))
	}
	old := append([]game.Card(nil), me.Hand.Cards...)
	if len(old) == 0 {
		t.Fatal("setup: empty hand")
	}
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.RingSettle(t, g)
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
	g, _, opp, sauron := effectstest.WardTable(t, "Sauron, the Dark Lord", "Legendary Creature — Avatar Horror", effectstest.SauronTheDarkLordOracle)
	bearer := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	effectstest.RingTempt(t, g, opp.ID)
	if effectstest.RingBearerOf(g, opp.ID) != bearer {
		t.Fatal("setup: the Bear is not their Ring-bearer")
	}
	effectstest.B16Creature(g, opp.ID, "Plain Bear", "Creature — Bear", 2, 2, "G")
	relic := effectstest.B12Permanent(g, opp.ID, "Legendary Relic", "Legendary Artifact")
	effectstest.B12Permanent(g, opp.ID, "Plain Rock", "Artifact")

	effectstest.CastAtWardedCreature(t, g, opp, sauron)
	// The spell also triggers Sauron's amass: settle the ordering, and
	// stop at the ward's prompt.
	effectstest.RingSettle(t, g)
	prompt := effectstest.ConfirmChoiceFor(g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt")
	}
	if err := g.ResolveConfirm(prompt.ID, opp.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	pick := effectstest.ChooseCardsChoiceFor(g, opp.ID)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	old := effectstest.Batch01GraveyardCard(me, "Old Bear", "Creature — Bear")
	milled := effectstest.SeedLibraryTop(me, "Milled Bear", "Creature — Bear", "{1}{G}")
	library := me.Library.Size()

	sauron := effectstest.CastRingCreature(t, g, "Sauron, Lord of the Rings", effectstest.SauronLordOfTheRingsTypeStr, effectstest.SauronLordOfTheRingsOracle, 9, 9)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
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
	effectstest.AnswerChooseCards(t, g, me.ID, old)
	effectstest.RingSettle(t, g)

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
// you." An opponent's commander that dies triggers it; your own does
// not. Since ADR 0115 (#2085) the commander dies before its owner is
// asked about the command zone (CR 903.9a), so it triggers "even if the
// owner of the commander that died chooses to return it to the command
// zone after it dies" (the card's ruling).
func TestSauronLordOfTheRingsTemptsWhenAnOpponentsCommanderDies(t *testing.T) {
	for _, tc := range []struct {
		name        string
		theirs      bool
		commandZone bool
		want        int
	}{
		{"opponent's commander left in the graveyard", true, false, 1},
		{"opponent's commander then sent to the command zone", true, true, 1},
		{"my own commander left in the graveyard", false, false, 0},
		{"my own commander then sent to the command zone", false, true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.B12Push(g, me.ID, "Sauron, Lord of the Rings", effectstest.SauronLordOfTheRingsTypeStr, effectstest.SauronLordOfTheRingsOracle, 9, 9)
			owner := me
			if tc.theirs {
				owner = opp
			}
			commander := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: "A Commander", TypeLine: "Legendary Creature — Human",
				Power: 2, Toughness: 2, Owner: owner.ID, Controller: owner.ID, IsCommander: true,
			})
			g.WithWriteLock(func() {
				if err := (effects.DestroyTarget{Target: commander}).Apply(effectstest.CtxFor(g, &game.StackItem{Controller: me.ID, SourceCardID: uuid.New()})); err != nil {
					t.Fatalf("destroy: %v", err)
				}
			})
			if !owner.Graveyard.Contains(commander) {
				t.Fatal("setup: the destroyed commander is not in its owner's graveyard")
			}
			effectstest.AnswerCommanderReturn(t, g, owner.ID, tc.commandZone)
			effectstest.RingSettle(t, g)
			if got := effectstest.RingCount(g, me.ID); got != tc.want {
				t.Fatalf("the Ring tempted Sauron's controller %d times, want %d", got, tc.want)
			}
		})
	}
}
