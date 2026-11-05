package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b38_cards_test.go — Star Trek Commander, batch 38 (#4974). Each
// test plays the real catalog entry through the engine and asserts what
// its text says.

const (
	trekB38RomOracle       = "b98ab107-3908-4425-89d3-634043bb4d4b"
	trekB38AmandaOracle    = "13d4a70f-c48c-4739-8f3e-5f0302c5484c"
	trekB38ChapelOracle    = "330e8a18-d224-4416-b59d-02b749c3a674"
	trekB38JaniceOracle    = "2130193c-3518-4a7d-9cb3-ea65aec845b4"
	trekB38SuluOracle      = "565cfa73-8c6b-4296-bcb8-388de8c960ff"
	trekB38AlexanderOracle = "d629311f-a7fc-4af5-91ce-5bf506a3cdb3"
	trekB38BElannaOracle   = "c905227d-b5fe-42e6-a11d-5008b7a7922f"
	trekB38BestOracle      = "a53d9c36-a033-450b-a095-7645876c9ef3"
	trekB38JanewayOracle   = "488ff0d9-a74c-4ba9-9c10-3aa88230c2bd"
	trekB38KirkOracle      = "5228fffe-6347-4177-bccb-b15621c7159c"
)

func TestTrekB38EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB38RomOracle:       "Rom and Leeta, Unlikely Pair",
		trekB38AmandaOracle:    "Amanda Grayson, Logic's Link",
		trekB38ChapelOracle:    "Christine Chapel, Combat Medic",
		trekB38JaniceOracle:    "Janice Rand, Essential Aide",
		trekB38SuluOracle:      "Lieutenant Hikaru Sulu",
		trekB38AlexanderOracle: "Alexander, Ship's Fool",
		trekB38BElannaOracle:   "B'Elanna Torres, Scrapper",
		trekB38BestOracle:      "The Best of Both Worlds",
		trekB38JanewayOracle:   "Captain Janeway, Endgame",
		trekB38KirkOracle:      "Captain Kirk, Boldly Going",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- shared test helpers ------------------------------------------------------

// trekB38Named counts the battlefield permanents named `name` that
// `controller` controls, and how many of them are tapped.
func trekB38Named(g *game.Game, controller uuid.UUID, name string) (n, tapped int) {
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			n++
			if c.Tapped {
				tapped++
			}
		}
	}
	return n, tapped
}

// trekB38Counters is the +1/+1 counters on a battlefield permanent.
func trekB38Counters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Counters[game.CounterPlusOne]
}

// trekB38Activate floats `mana` and activates ability `idx` strictly.
func trekB38Activate(t *testing.T, g *game.Game, who, id uuid.UUID, idx int, mana string, params game.ActivateAbilityParams) error {
	t.Helper()
	if mana != "" {
		g.WithWriteLock(func() {
			if err := g.AddManaForEffect(who, uuid.Nil, mana); err != nil {
				t.Fatalf("AddManaForEffect(%s): %v", mana, err)
			}
		})
	}
	params.Strict = true
	return g.ActivateCatalogAbility(who, id, idx, params)
}

// trekB38Attack declares the attackers at `opp`, locks the attack in and
// lets the attack triggers resolve.
func trekB38Attack(t *testing.T, g *game.Game, opp uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, a := range attackers {
		if err := g.DeclareAttacker(a, opp); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
}

// --- Rom and Leeta, Unlikely Pair -----------------------------------------------

// The second card a player draws in a turn makes two tapped Treasures, on
// an opponent's turn too; the first and the third make nothing.
func TestTrekB38RomAndLeetaMakeTwoTappedTreasuresOnTheSecondDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1] // not the active player: "each turn"
	b27Push(g, me.ID, "Rom and Leeta, Unlikely Pair", "Legendary Creature — Ferengi Bajoran", trekB38RomOracle, "{3}{R}", 4, 3, "R")
	advanceToMain(t, g)

	drawFor(t, g, me, 1)
	passPriorityAroundTable(t, g)
	if n, _ := trekB38Named(g, me.ID, "Treasure"); n != 0 {
		t.Fatalf("the first draw made %d Treasures", n)
	}
	drawFor(t, g, me, 1)
	passPriorityAroundTable(t, g)
	if n, tapped := trekB38Named(g, me.ID, "Treasure"); n != 2 || tapped != 2 {
		t.Fatalf("after the second draw: %d Treasures, %d tapped; want 2 tapped", n, tapped)
	}
	drawFor(t, g, me, 1)
	passPriorityAroundTable(t, g)
	if n, _ := trekB38Named(g, me.ID, "Treasure"); n != 2 {
		t.Errorf("the third draw made more Treasures: %d", n)
	}
}

// --- Amanda Grayson, Logic's Link -----------------------------------------------

// Affinity for Vulcans: two Vulcans make her cost {1}{G}; one does not.
func TestTrekB38AmandaGraysonCostsOneLessPerVulcan(t *testing.T) {
	for _, c := range []struct {
		vulcans int
		ok      bool
	}{{2, true}, {1, false}} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		for i := 0; i < c.vulcans; i++ {
			b12Creature(g, me.ID, "Vulcan", "Creature — Vulcan Scientist", 1, 1)
		}
		b12Creature(g, me.ID, "Human", "Creature — Human", 1, 1)
		advanceToMain(t, g)
		floatMana(t, g, me, "{C}{G}")
		id := pushHandCardWithManaCost(me, "Amanda Grayson, Logic's Link", "Legendary Creature — Human Advisor", "{3}{G}")
		g.WithWriteLock(func() { me.Hand.Cards[len(me.Hand.Cards)-1].OracleID = trekB38AmandaOracle })
		err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
		if (err == nil) != c.ok {
			t.Errorf("%d Vulcans: cast for {1}{G} = %v, want ok %v", c.vulcans, err, c.ok)
		}
	}
}

// Landfall puts two +1/+1 counters on the creature you control it targets.
func TestTrekB38AmandaGraysonLandfallPutsTwoCountersOnTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	amanda := b27Push(g, me.ID, "Amanda Grayson, Logic's Link", "Legendary Creature — Human Advisor", trekB38AmandaOracle, "{3}{G}", 2, 3, "G")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	playLandFromHand(t, g, "Forest", "")
	for _, ch := range g.PendingChoices {
		if ch != nil && ch.Kind == game.PendingChoicePickTarget {
			if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
				t.Fatal("landfall targeted a creature an opponent controls")
			}
		}
	}
	answerPickTargetIfOpen(t, g, bear)
	passPriorityAroundTable(t, g)
	if got := trekB38Counters(t, g, bear); got != 2 {
		t.Errorf("the Bear has %d counters, want 2", got)
	}
	if got := trekB38Counters(t, g, amanda); got != 0 {
		t.Errorf("Amanda has %d counters, want 0", got)
	}
}

// --- Christine Chapel, Combat Medic ---------------------------------------------

// She gains life for each differently named land, and every life gain puts
// two counters on her, the enter trigger's own included.
func TestTrekB38ChristineChapelGainsPerLandNameAndGrowsOnLifeGain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushLands(g, me.ID, 2) // two Forests: one name
	b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	life := me.Life

	chapel := castCatalogSpell(t, g, "Christine Chapel, Combat Medic", "Legendary Creature — Human Doctor", trekB38ChapelOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)

	if me.Life != life+2 {
		t.Errorf("life = %d, want %d (two land names)", me.Life, life+2)
	}
	if got := trekB38Counters(t, g, chapel); got != 2 {
		t.Fatalf("after the enter trigger's gain: %d counters, want 2", got)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	passPriorityAroundTable(t, g)
	if got := trekB38Counters(t, g, chapel); got != 4 {
		t.Errorf("after another gain: %d counters, want 4 (one trigger per gain, not per point)", got)
	}
}

// --- Janice Rand, Essential Aide ------------------------------------------------

// A commander of yours attacking gets a counter, and the reveal puts the
// first land onto the battlefield tapped and the rest on the bottom.
func TestTrekB38JaniceRandCommanderAttacksCounterAndLand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Janice Rand, Essential Aide", "Legendary Creature — Human Advisor", trekB38JaniceOracle, "{3}{G}", 3, 3, "G")
	cmd := b12Creature(g, me.ID, "My Commander", "Legendary Creature — Human", 2, 2)
	plain := b12Creature(g, me.ID, "Not a Commander", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == cmd {
				g.Battlefield.Cards[i].IsCommander = true
			}
		}
	})
	land := seedLibraryTop(me, "Found Forest", "Basic Land — Forest", "")
	spells := seedLibrary(me, "Spell A", "Spell B")
	libBefore := me.Library.Size()

	trekB38Attack(t, g, opp.ID, cmd, plain)

	if got := trekB38Counters(t, g, cmd); got != 1 {
		t.Errorf("the commander has %d counters, want 1", got)
	}
	if got := trekB38Counters(t, g, plain); got != 0 {
		t.Errorf("a non-commander attacker got %d counters", got)
	}
	c, ok := battlefieldCard(g, land)
	if !ok || !c.Tapped || c.Controller != me.ID {
		t.Fatalf("the land is not on the battlefield tapped under my control (ok %v)", ok)
	}
	if me.Library.Size() != libBefore-1 {
		t.Errorf("library %d, want %d", me.Library.Size(), libBefore-1)
	}
	n := me.Library.Size()
	bottom := map[uuid.UUID]bool{me.Library.Cards[0].InstanceID: true, me.Library.Cards[1].InstanceID: true}
	top := map[uuid.UUID]bool{me.Library.Cards[n-1].InstanceID: true, me.Library.Cards[n-2].InstanceID: true}
	for _, id := range spells {
		if !bottom[id] && !top[id] {
			t.Errorf("a revealed card is in the middle of the library")
		}
		if top[id] {
			t.Errorf("a revealed card stayed on top of the library")
		}
	}
}

// An opponent's commander attacking does not trigger it.
func TestTrekB38JaniceRandIgnoresAnOpponentsCommander(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, opp.ID, "Janice Rand, Essential Aide", "Legendary Creature — Human Advisor", trekB38JaniceOracle, "{3}{G}", 3, 3, "G")
	cmd := b12Creature(g, me.ID, "My Commander", "Legendary Creature — Human", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == cmd {
				g.Battlefield.Cards[i].IsCommander = true
			}
		}
	})
	trekB38Attack(t, g, opp.ID, cmd)
	if got := trekB38Counters(t, g, cmd); got != 0 {
		t.Errorf("an opponent's Janice put %d counters on my commander", got)
	}
}

// --- Lieutenant Hikaru Sulu -----------------------------------------------------

// Attacking taps him, so he explores: a land on top goes to hand.
func TestTrekB38SuluExploresWhenHeBecomesTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sulu := b27Push(g, me.ID, "Lieutenant Hikaru Sulu", "Legendary Creature — Human Pilot", trekB38SuluOracle, "{2}{G}", 3, 3, "G")
	land := seedLibraryTop(me, "Top Forest", "Basic Land — Forest", "")

	trekB38Attack(t, g, opp.ID, sulu)

	if !me.Hand.Contains(land) {
		t.Fatal("the explored land is not in hand")
	}
	if got := trekB38Counters(t, g, sulu); got != 0 {
		t.Errorf("a land explore put %d counters on him", got)
	}
}

// A nonland on top: a +1/+1 counter, and the card is kept or binned.
func TestTrekB38SuluExploresANonlandForACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sulu := b27Push(g, me.ID, "Lieutenant Hikaru Sulu", "Legendary Creature — Human Pilot", trekB38SuluOracle, "{2}{G}", 3, 3, "G")
	seedLibraryTop(me, "Top Spell", "Sorcery", "{1}")
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(sulu) })
	passPriorityAroundTable(t, g)
	if c := latestConfirmFor(g, me.ID); c != nil {
		answerMayChoice(t, g, me.ID, true)
		passPriorityAroundTable(t, g)
	}
	if got := trekB38Counters(t, g, sulu); got != 1 {
		t.Errorf("he has %d counters, want 1", got)
	}
}

// {2}{U}: target creature can't be blocked this turn, and only for {U}.
func TestTrekB38SuluMakesACreatureUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sulu := b27Push(g, me.ID, "Lieutenant Hikaru Sulu", "Legendary Creature — Human Pilot", trekB38SuluOracle, "{2}{G}", 3, 3, "G")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	blocker := b12Creature(g, opp.ID, "Blocker", "Creature — Wall", 0, 4)
	advanceToMain(t, g)
	if err := trekB38Activate(t, g, me.ID, sulu, 0, "{G}{G}{G}", game.ActivateAbilityParams{Targets: cardRefs(bear)}); err == nil {
		t.Fatal("activated without {U}")
	}
	if err := trekB38Activate(t, g, me.ID, sulu, 0, "{U}", game.ActivateAbilityParams{Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("activation: %v", err)
	}
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, bear); err == nil {
		t.Error("the creature was blocked")
	}
}

// --- Alexander, Ship's Fool -----------------------------------------------------

// Every player reveals and draws; one opponent takes damage equal to the
// greatest revealed mana value.
func TestTrekB38AlexanderRevealsDrawsAndHitsARandomOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	alex := b27Push(g, me.ID, "Alexander, Ship's Fool", "Legendary Creature — Klingon Human", trekB38AlexanderOracle, "{3}{R}{W}", 3, 5, "R", "W")
	costs := []string{"{1}", "{4}{R}", "", "{2}"}
	for i, p := range g.Seats {
		seedLibraryTop(p, "Top", "Sorcery", costs[i])
	}
	advanceToMain(t, g)
	hands := make([]int, len(g.Seats))
	lives := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		hands[i], lives[i] = p.Hand.Size(), p.Life
	}
	if err := trekB38Activate(t, g, me.ID, alex, 0, "", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activation: %v", err)
	}
	passPriorityAroundTable(t, g)

	hit := 0
	for i, p := range g.Seats {
		if p.Hand.Size() != hands[i]+1 {
			t.Errorf("seat %d drew %d cards, want 1", i, p.Hand.Size()-hands[i])
		}
		switch lost := lives[i] - p.Life; {
		case lost == 0:
		case lost == 5 && i != 0:
			hit++
		default:
			t.Errorf("seat %d lost %d life", i, lost)
		}
	}
	if hit != 1 {
		t.Errorf("%d opponents took 5 damage, want exactly 1", hit)
	}
	if g.ActivateCatalogAbility(me.ID, alex, 0, game.ActivateAbilityParams{Strict: true}) == nil {
		t.Error("activated again while tapped")
	}
}

// --- B'Elanna Torres, Scrapper --------------------------------------------------

func trekB38BElannaTable(t *testing.T, top game.Card) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	b := b27Push(g, me.ID, "B'Elanna Torres, Scrapper", "Legendary Creature — Klingon Human Artificer", trekB38BElannaOracle, "{1}{U}{R}", 2, 2, "U", "R")
	seedLibrary(me, "Surveil A", "Surveil B")
	// Below the two surveilled cards, the card the reveal finds.
	top.InstanceID, top.Owner, top.Controller = uuid.New(), me.ID, me.ID
	g.WithWriteLock(func() {
		cards := me.Library.Cards
		n := len(cards)
		me.Library.Cards = append(append(append([]game.Card(nil), cards[:n-2]...), top), cards[n-2:]...)
	})
	advanceToMain(t, g)
	if err := trekB38Activate(t, g, me.ID, b, 0, "{C}", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated with {1}")
	}
	if err := trekB38Activate(t, g, me.ID, b, 0, "{C}", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activation: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	return g, me, b
}

// An artifact card revealed after the surveil: draw it, no counters.
func TestTrekB38BElannaDrawsARevealedArtifact(t *testing.T) {
	g, me, b := trekB38BElannaTable(t, game.Card{Name: "Widget", TypeLine: "Artifact"})
	hand := me.Hand.Size()
	if me.Graveyard.Size() != 2 {
		t.Fatalf("graveyard %d, want the two surveilled cards", me.Graveyard.Size())
	}
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
	if got := trekB38Counters(t, g, b); got != 0 {
		t.Errorf("%d counters after drawing, want 0", got)
	}
}

// A nonartifact revealed, or no reveal at all: two counters, no draw.
func TestTrekB38BElannaGrowsWithoutADraw(t *testing.T) {
	for _, reveal := range []bool{true, false} {
		top := game.Card{Name: "Widget", TypeLine: "Artifact"}
		if reveal {
			top = game.Card{Name: "Plain Spell", TypeLine: "Sorcery"}
		}
		g, me, b := trekB38BElannaTable(t, top)
		hand := me.Hand.Size()
		answerMayChoice(t, g, me.ID, reveal)
		passPriorityAroundTable(t, g)
		if me.Hand.Size() != hand {
			t.Errorf("reveal %v: drew %d cards", reveal, me.Hand.Size()-hand)
		}
		if got := trekB38Counters(t, g, b); got != 2 {
			t.Errorf("reveal %v: %d counters, want 2", reveal, got)
		}
	}
}

// --- The Best of Both Worlds ----------------------------------------------------

func TestTrekB38BestOfBothWorldsAssimilatesPumpsAndReturns(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf Warrior", 2, 2)
	b12Creature(g, opp.ID, "Their Golem", "Artifact Creature — Golem", 3, 3)
	saga := castCatalogSpell(t, g, "The Best of Both Worlds", "Enchantment — Saga", trekB38BestOracle, nil)
	passPriorityAroundTable(t, g)
	answerPickTargetIfOpen(t, g, victim)
	passPriorityAroundTable(t, g)

	c := layeredCard(t, g, victim)
	if c.Controller != me.ID {
		t.Fatal("chapter I did not gain control of the creature")
	}
	if !c.IsArtifact() || !c.IsCreature() || !c.HasSubtype("Borg") || c.HasSubtype("Elf") || c.HasSubtype("Warrior") {
		t.Errorf("assimilated creature is %v, want a Borg artifact creature with no other creature type", c.Effective().Subtypes)
	}
	if got := trekB38Counters(t, g, victim); got != 1 {
		t.Errorf("assimilated creature has %d counters, want 1", got)
	}
	if bc, _ := battlefieldCard(g, saga); bc.LinkedPermanent.ID != victim {
		t.Error("the Saga does not remember the creature it assimilated")
	}

	// Chapter II: artifact creatures you control get +2/+2.
	advanceToPrecombatMainOf(t, g, g.Turn.ActiveSeat)
	passPriorityAroundTable(t, g)
	if c := layeredCard(t, g, victim); c.PowerForComparison() != 5 || c.CurrentToughness() != 5 {
		t.Errorf("chapter II: the Borg is %d/%d, want 5/5 (2/2, its counter, +2/+2)", c.PowerForComparison(), c.CurrentToughness())
	}

	// Chapter III: the Borg is exiled and returns to its owner, a new object.
	advanceToPrecombatMainOf(t, g, g.Turn.ActiveSeat)
	passPriorityAroundTable(t, g)
	var back *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Their Elf" {
			back = &g.Battlefield.Cards[i]
		}
	}
	if back == nil {
		t.Fatal("the Borg did not return to the battlefield")
	}
	l := layeredCard(t, g, back.InstanceID)
	if l.Controller != opp.ID || l.IsArtifact() || !l.HasSubtype("Elf") || l.Counters[game.CounterPlusOne] != 0 {
		t.Errorf("returned creature: controller me=%v artifact=%v elf=%v counters=%d", l.Controller == me.ID, l.IsArtifact(), l.HasSubtype("Elf"), l.Counters[game.CounterPlusOne])
	}
	if loreCountersOn(g, saga) >= 0 {
		t.Error("the Saga survived chapter III")
	}
}

// Chapter I cannot target an artifact creature or one you control.
func TestTrekB38BestOfBothWorldsTargetsOnlyAnOpponentsNonartifactCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Elf", "Creature — Elf", 2, 2)
	golem := b12Creature(g, opp.ID, "Their Golem", "Artifact Creature — Golem", 3, 3)
	victim := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 2, 2)
	castCatalogSpell(t, g, "The Best of Both Worlds", "Enchantment — Saga", trekB38BestOracle, nil)
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("chapter I asked for no target")
	}
	for _, bad := range []uuid.UUID{mine, golem} {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Errorf("chapter I accepted an illegal target %s", bad)
		}
	}
	answerPickTarget(t, g, victim)
	passPriorityAroundTable(t, g)
	if c := layeredCard(t, g, victim); c.Controller != me.ID {
		t.Error("the legal target was not assimilated")
	}
}

// A Borg that left and came back is a new object: chapter III does nothing.
func TestTrekB38BestOfBothWorldsForgetsABorgThatLeft(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 2, 2)
	saga := castCatalogSpell(t, g, "The Best of Both Worlds", "Enchantment — Saga", trekB38BestOracle, nil)
	passPriorityAroundTable(t, g)
	answerPickTargetIfOpen(t, g, victim)
	passPriorityAroundTable(t, g)
	flickerForTest(t, g, me.ID, victim)
	var epoch int
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Their Elf" {
			epoch = c.ObjectEpoch
		}
	}
	advanceToPrecombatMainOf(t, g, g.Turn.ActiveSeat)
	passPriorityAroundTable(t, g)
	advanceToPrecombatMainOf(t, g, g.Turn.ActiveSeat)
	passPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Their Elf" && c.ObjectEpoch != epoch {
			t.Error("chapter III flickered a creature that was not the one assimilated")
		}
	}
	if loreCountersOn(g, saga) >= 0 {
		t.Error("the Saga survived chapter III")
	}
}

// --- Captain Janeway, Endgame ---------------------------------------------------

// Two Officers connecting with one player: one trigger, one card and one
// Officer token.
func TestTrekB38JanewayDrawsAndMakesAnOfficerOncePerPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jw := b27Push(g, me.ID, "Captain Janeway, Endgame", "Legendary Creature — Human Officer", trekB38JanewayOracle, "{1}{U}{R}{W}", 3, 3, "U", "R", "W")
	o2 := b12Creature(g, me.ID, "Ensign", "Creature — Human Officer", 1, 1)
	trekB38Attack(t, g, opp.ID, jw, o2)
	hand := me.Hand.Size()
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("drew %d cards, want 1", me.Hand.Size()-hand)
	}
	if n, _ := trekB38Named(g, me.ID, "Officer"); n != 1 {
		t.Errorf("%d Officer tokens, want 1", n)
	}
}

// A non-Officer connecting does nothing.
func TestTrekB38JanewayIgnoresANonOfficer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Captain Janeway, Endgame", "Legendary Creature — Human Officer", trekB38JanewayOracle, "{1}{U}{R}{W}", 3, 3, "U", "R", "W")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	trekB38Attack(t, g, opp.ID, bear)
	hand := me.Hand.Size()
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("a Bear's damage drew %d cards", me.Hand.Size()-hand)
	}
}

// {2}, tap an untapped creature you control: she can't be blocked.
func TestTrekB38JanewayTapsACreatureToBeUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jw := b27Push(g, me.ID, "Captain Janeway, Endgame", "Legendary Creature — Human Officer", trekB38JanewayOracle, "{1}{U}{R}{W}", 3, 3, "U", "R", "W")
	helper := b12Creature(g, me.ID, "Helper", "Creature — Bear", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	blocker := b12Creature(g, opp.ID, "Blocker", "Creature — Wall", 0, 4)
	advanceToMain(t, g)
	if err := trekB38Activate(t, g, me.ID, jw, 0, "{C}", game.ActivateAbilityParams{TapIDs: []uuid.UUID{helper}}); err == nil {
		t.Fatal("activated with {1}")
	}
	if err := trekB38Activate(t, g, me.ID, jw, 0, "{C}", game.ActivateAbilityParams{TapIDs: []uuid.UUID{theirs}}); err == nil {
		t.Fatal("tapped an opponent's creature to pay")
	}
	if err := trekB38Activate(t, g, me.ID, jw, 0, "", game.ActivateAbilityParams{TapIDs: []uuid.UUID{helper}}); err != nil {
		t.Fatalf("activation: %v", err)
	}
	if c, _ := battlefieldCard(g, helper); !c.Tapped {
		t.Error("the helper was not tapped to pay")
	}
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(jw, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, jw); err == nil {
		t.Error("Janeway was blocked")
	}
}

// --- Captain Kirk, Boldly Going -------------------------------------------------

// Spacecraft creatures you control get +3/+3; other creatures and an
// opponent's Spacecraft do not.
func TestTrekB38KirkPumpsYourSpacecraftCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Captain Kirk, Boldly Going", "Legendary Creature — Human Officer", trekB38KirkOracle, "{1}{G}{W}{U}", 3, 4, "G", "W", "U")
	ship := b12Creature(g, me.ID, "My Ship", "Artifact Creature — Spacecraft", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Ship", "Artifact Creature — Spacecraft", 2, 2)
	for id, want := range map[uuid.UUID]int{ship: 5, bear: 2, theirs: 2} {
		if c := layeredCard(t, g, id); c.Effective().Power != want || c.Effective().Toughness != want {
			t.Errorf("%s is %d/%d, want %d/%d", c.Name, c.Effective().Power, c.Effective().Toughness, want, want)
		}
	}
}

// Entering, he fetches a card named Starship Enterprise from the graveyard
// or a basic land from the library.
func TestTrekB38KirkFetchesOnEnterAndAttack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	lgKnowLibrary(me)
	enterprise := pushGraveyardCardWithTypeLineID(me, "Starship Enterprise", "Legendary Artifact — Spacecraft")
	castCatalogSpell(t, g, "Captain Kirk, Boldly Going", "Legendary Creature — Human Officer", trekB38KirkOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("Kirk's enter trigger did not search")
	}
	for _, id := range c.SearchCards {
		if lc, ok := g.LookupCardForEffect(id); ok && lc.Name != "Forest" && lc.Name != "Starship Enterprise" {
			t.Errorf("the search offered %q", lc.Name)
		}
	}
	answerSearchZones(t, g, me.ID, game.SearchZoneGraveyard)
	answerSearchByID(t, g, me.ID, enterprise)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(enterprise) {
		t.Fatal("Starship Enterprise was not put into hand")
	}
	if libraryKnownTo(me, me.ID) == 0 {
		t.Error("the library was shuffled though only the graveyard was searched")
	}

}

// Attacking searches too; a basic land from the library this time, and
// taking it shuffles.
func TestTrekB38KirkFetchesABasicLandWhenHeAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kirk := b27Push(g, me.ID, "Captain Kirk, Boldly Going", "Legendary Creature — Human Officer", trekB38KirkOracle, "{1}{G}{W}{U}", 3, 4, "G", "W", "U")
	seedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	pushLibraryCardForTest(me, game.Card{Name: "Plain Spell", TypeLine: "Sorcery"})
	lgKnowLibrary(me)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(kirk, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil || len(c.SearchCards) == 0 {
		t.Fatal("Kirk's attack trigger did not offer the library's Forest")
	}
	forest := c.SearchCards[0]
	answerSearchByID(t, g, me.ID, forest)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(forest) {
		t.Error("the basic land was not put into hand")
	}
	if libraryKnownTo(me, me.ID) != 0 {
		t.Error("the library was searched but not shuffled")
	}
}
