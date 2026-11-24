package testsru

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap2_b10_test.go — S82 GAP2 batch 10 (#5948): per-object notes
// (#6097) and token-creation and grant variants (#6099).

const (
	gap2B10LongList   = "01376449-5cd2-4079-8aaa-5128634ef20b"
	gap2B10Volo       = "9b61f5b1-7688-4336-8fe9-66da37a34df9"
	gap2B10DayOfMoon  = "7ec3aaa4-8b6a-4046-b599-654f9844ca6c"
	gap2B10Wisp       = "d39df714-bd0b-4de8-b9df-55713bab7f35"
	gap2B10Rani       = "ac9b0bdf-caf8-4dba-a900-e59de35e77d7"
	gap2B10Arna       = "436e7548-f8d5-43af-947d-b382da0d19b3"
	gap2B10HenryWu    = "d94a792f-a48b-4181-b656-86fe452a4875"
	gap2B10Gandalf    = "5b3b5f6a-375e-4479-9384-a942eda83f9b"
	gap2B10Tremors    = "9242cd3e-1a71-4700-8182-9c1005616033"
	gap2B10Zulaport   = "76b003e0-15af-4f22-bdf2-1ade5430964a"
	gap2B10BearTL     = "Creature — Bear"
	gap2B10Generation = "Creature — Human Soldier"
)

func TestGap2B10EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gap2B10LongList:  "Long List of the Ents",
		gap2B10Volo:      "Volo, Itinerant Scholar",
		gap2B10DayOfMoon: "Day of the Moon",
		gap2B10Wisp:      "Beckoning Will-o'-Wisp",
		gap2B10Rani:      "The Rani",
		gap2B10Arna:      "Arna Kennerüd, Skycaptain",
		gap2B10HenryWu:   "Henry Wu, InGen Geneticist",
		gap2B10Gandalf:   "Gandalf the White",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func gap2B10Pending(g *game.Game, kind game.PendingChoiceKind) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == kind {
			return c
		}
	}
	return nil
}

func gap2B10Card(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	var c game.Card
	var ok bool
	g.ReadSnapshot(func() { c, ok = g.LookupCardForEffect(id) })
	if !ok {
		t.Fatalf("card %s not found", id)
	}
	return c
}

func gap2B10Notes(g *game.Game, id uuid.UUID) []string {
	var notes []string
	g.ReadSnapshot(func() { notes = g.NotesForEffect(id) })
	return notes
}

func gap2B10CastCreature(t *testing.T, g *game.Game, p *game.Player, name, typeLine string) uuid.UUID {
	t.Helper()
	id := effectstest.ImportToHand(cards.Card{ID: uuid.New(), OracleID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: "{0}", Power: "2", Toughness: "2"}, p)
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

// --- Beckoning Will-o'-Wisp -------------------------------------------------

func TestGap2B10WispGivesCreaturesAttackingTheChosenPlayerPlusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	wisp := effectstest.B12Push(g, me.ID, "Beckoning Will-o'-Wisp", "Creature — Spirit", gap2B10Wisp, 1, 3)
	if !effectstest.HasKeywordOnBattlefield(t, g, wisp, "flying") {
		t.Error("Beckoning Will-o'-Wisp has no flying")
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	chosenID := g.ChosenPlayerOf(wisp)
	var chosen, other *game.Player
	for _, s := range g.Seats[1:] {
		if s.ID == chosenID {
			chosen = s
		} else if other == nil {
			other = s
		}
	}
	if chosen == nil {
		t.Fatal("no opponent was chosen")
	}
	toChosen := effectstest.B12Creature(g, me.ID, "Bear A", gap2B10BearTL, 2, 2)
	toOther := effectstest.B12Creature(g, me.ID, "Bear B", gap2B10BearTL, 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Bear C", gap2B10BearTL, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(toChosen, chosen.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(toOther, other.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	if p := effectstest.EffectivePower(t, g, toChosen); p != 3 {
		t.Errorf("a creature attacking the chosen player has power %d, want 3", p)
	}
	if p := effectstest.EffectivePower(t, g, toOther); p != 2 {
		t.Errorf("a creature attacking another player has power %d, want 2", p)
	}
	if p := effectstest.EffectivePower(t, g, idle); p != 2 {
		t.Errorf("a creature not attacking has power %d, want 2", p)
	}
}

// --- Gandalf the White -----------------------------------------------------

func TestGap2B10GandalfDoublesEnteringArtifactsButNotOtherCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Gandalf the White", "Legendary Creature — Avatar Wizard", gap2B10Gandalf, 4, 5)
	effectstest.PushCatalogPermanent(g, me.ID, "Impact Tremors", "Enchantment", gap2B10Tremors, false)
	before := effectstest.LifeOfOpponents(g)
	effectstest.EnterFromHand(t, g, me.ID, "Construct", "Artifact Creature — Construct", "")
	effectstest.PassPriorityAroundTable(t, g)
	for i, life := range before {
		if got := g.Seats[i+1].Life; got != life-2 {
			t.Errorf("an artifact creature entering: %d -> %d, want -2", life, got)
		}
	}
	before = effectstest.LifeOfOpponents(g)
	effectstest.EnterFromHand(t, g, me.ID, "Bear", gap2B10BearTL, "")
	effectstest.PassPriorityAroundTable(t, g)
	for i, life := range before {
		if got := g.Seats[i+1].Life; got != life-1 {
			t.Errorf("a nonlegendary nonartifact creature entering: %d -> %d, want -1", life, got)
		}
	}
}

func TestGap2B10GandalfDoublesALegendaryCreatureLeaving(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Gandalf the White", "Legendary Creature — Avatar Wizard", gap2B10Gandalf, 4, 5)
	effectstest.PushCatalogPermanent(g, me.ID, "Zulaport Cutthroat", "Creature — Human Rogue Ally", gap2B10Zulaport, false)
	legend := effectstest.B12Creature(g, me.ID, "Some Legend", "Legendary Creature — Human", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", gap2B10BearTL, 2, 2)

	life := me.Life
	effectstest.KillCreature(t, g, me.ID, legend)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Life - life; got != 2 {
		t.Errorf("a legendary creature dying: gained %d, want 2 (doubled)", got)
	}
	life = me.Life
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Life - life; got != 1 {
		t.Errorf("a nonlegendary creature dying: gained %d, want 1", got)
	}
}

func TestGap2B10GandalfDoesNotDoubleAnOpponentsAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Gandalf the White", "Legendary Creature — Avatar Wizard", gap2B10Gandalf, 4, 5)
	effectstest.PushCatalogPermanent(g, opp.ID, "Impact Tremors", "Enchantment", gap2B10Tremors, false)
	life := me.Life
	effectstest.EnterFromHand(t, g, opp.ID, "Construct", "Artifact Creature — Construct", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - me.Life; got != 1 {
		t.Errorf("an opponent's Impact Tremors dealt %d, want 1", got)
	}
}

// --- Henry Wu, InGen Geneticist ------------------------------------------

func TestGap2B10HenryWuGivesHumansExploitAndPaysForANonHuman(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.B12Push(g, me.ID, "Henry Wu, InGen Geneticist", "Legendary Creature — Human Scientist", gap2B10HenryWu, 1, 4)
	fodder := effectstest.B12Creature(g, me.ID, "Big Beast", "Creature — Beast", 4, 4)
	hand := me.Hand.Size()
	effectstest.EnterFromHand(t, g, me.ID, "Recruit", gap2B10Generation, "")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Fatal("the Human's granted exploit did not sacrifice the Beast")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("Treasures = %d, want 1 for a power-4 creature", n)
	}
}

func TestGap2B10HenryWuPaysNothingForAHumanAndNoTreasureForASmallOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.B12Push(g, me.ID, "Henry Wu, InGen Geneticist", "Legendary Creature — Human Scientist", gap2B10HenryWu, 1, 4)
	human := effectstest.B12Creature(g, me.ID, "Human Fodder", "Creature — Human", 5, 5)
	small := effectstest.B12Creature(g, me.ID, "Small Beast", "Creature — Beast", 2, 2)
	hand := me.Hand.Size()

	effectstest.EnterFromHand(t, g, me.ID, "Recruit", gap2B10Generation, "")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, human)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("exploiting a Human drew %d cards, want 0", got)
	}

	effectstest.EnterFromHand(t, g, me.ID, "Second Recruit", gap2B10Generation, "")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, small)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("exploiting a non-Human drew %d cards in all, want 1", got)
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Treasure"); n != 0 {
		t.Errorf("Treasures = %d, want 0 for a power-2 creature", n)
	}
}

func TestGap2B10HenryWuDoesNotGiveExploitToANonHuman(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.B12Push(g, me.ID, "Henry Wu, InGen Geneticist", "Legendary Creature — Human Scientist", gap2B10HenryWu, 1, 4)
	effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Beast", 4, 4)
	effectstest.EnterFromHand(t, g, me.ID, "Elf", "Creature — Elf", "")
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 {
		t.Errorf("a non-Human entering asked %d questions, want none", len(g.PendingChoices))
	}
}

// --- The Rani ---------------------------------------------------------------

func TestGap2B10RaniMarksAnotherCreatureAndInvestigatesOnGoadedHits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", gap2B10BearTL, 2, 2)
	effectstest.EnterFromHand(t, g, me.ID, "The Rani", "Legendary Creature — Time Lord Scientist", gap2B10Rani)
	effectstest.PickTriggerTarget(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)

	var mark game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Mark of the Rani" {
			mark = c
		}
	}
	if mark.InstanceID == uuid.Nil || mark.Controller != me.ID || !mark.IsAttachedTo(theirs) {
		t.Fatalf("Mark of the Rani = %+v, want a token of mine attached to their Bear", mark)
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 4 {
		t.Errorf("the marked creature has power %d, want 4", p)
	}
	if !gap2B10Card(t, g, theirs).Goaded() {
		t.Error("the marked creature is not goaded")
	}

	victim := g.Seats[2]
	effectstest.DealCombatDamageToPlayer(g, theirs, victim.ID, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Clue"); n != 1 {
		t.Errorf("Clues = %d after a goaded creature hit an opponent, want 1", n)
	}
	// Damage to me is not damage to one of my opponents, and a creature
	// that isn't goaded investigates nothing.
	effectstest.DealCombatDamageToPlayer(g, theirs, me.ID, 4)
	plain := effectstest.B12Creature(g, opp.ID, "Plain Bear", gap2B10BearTL, 2, 2)
	effectstest.DealCombatDamageToPlayer(g, plain, victim.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Clue"); n != 1 {
		t.Errorf("Clues = %d, want still 1", n)
	}
}

// --- Arna Kennerüd, Skycaptain ----------------------------------------------

func TestGap2B10ArnaDoublesCountersAndCopiesAttachments(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	arna := effectstest.B12Push(g, me.ID, "Arna Kennerüd, Skycaptain", "Legendary Creature — Human Knight", gap2B10Arna, 4, 4)
	for _, kw := range []string{"flying", "lifelink"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, arna, kw) {
			t.Errorf("Arna has no %s", kw)
		}
	}
	knight := effectstest.B12Creature(g, me.ID, "Knight", "Creature — Human Knight", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain Bear", gap2B10BearTL, 2, 2)
	sword := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Plain Sword", TypeLine: "Artifact — Equipment", Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			switch g.Battlefield.Cards[i].InstanceID {
			case knight:
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterPlusOne: 2, "flying": 1}
			case sword:
				g.Battlefield.Cards[i].AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: knight}
			}
		}
	})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	opp := g.Seats[1]
	if err := g.DeclareAttacker(knight, opp.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(plain, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)

	c := gap2B10Card(t, g, knight)
	if c.Counters[game.CounterPlusOne] != 4 || c.Counters["flying"] != 2 {
		t.Errorf("knight counters = %v, want four +1/+1 and two flying", c.Counters)
	}
	swords := 0
	for _, b := range g.Battlefield.Cards {
		if b.Name == "Plain Sword" && b.IsAttachedTo(knight) {
			swords++
			if b.InstanceID != sword && !effects.IsToken(b) {
				t.Errorf("the second Sword is not a token: %+v", b)
			}
		}
	}
	if swords != 2 {
		t.Errorf("Swords attached to the knight = %d, want 2 (the card and a token copy)", swords)
	}
	if p := gap2B10Card(t, g, plain); len(p.Counters) != 0 {
		t.Errorf("an unmodified attacker got counters %v", p.Counters)
	}
}

// --- Day of the Moon --------------------------------------------------------

func gap2B10AnswerName(t *testing.T, g *game.Game, who uuid.UUID, name string) {
	t.Helper()
	c := gap2B10Pending(g, game.PendingChoiceCardName)
	if c == nil {
		t.Fatalf("no card-name prompt is open (answering %s)", name)
	}
	if err := g.ResolveCardNameChoice(c.ID, who, name); err != nil {
		t.Fatalf("ResolveCardNameChoice(%s): %v", name, err)
	}
}

// gap2B10NextLore adds a lore counter to the Saga and puts the chapter it
// reaches on the stack (CR 714.2b), without walking a turn the goaded
// creatures would have to attack in.
func gap2B10NextLore(t *testing.T, g *game.Game, me, saga uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() { _ = g.AddCounterByForEffect(me, saga, game.CounterLore, 1) })
	effectstest.PassPriorityAroundTable(t, g)
}

func TestGap2B10DayOfTheMoonGoadsEveryNameChosenForIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	bears := effectstest.B12Creature(g, opp.ID, "Grizzly Bears", gap2B10BearTL, 2, 2)
	elves := effectstest.B12Creature(g, opp.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	land := effectstest.B12Creature(g, opp.ID, "Mutavault", "Land", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == land {
				// An animated land: a creature now, but its card is a land.
				g.Battlefield.Cards[i].TypeLine = "Land"
			}
		}
	})
	saga := effectstest.EnterFromHand(t, g, me.ID, "Day of the Moon", "Enchantment — Saga", gap2B10DayOfMoon)
	effectstest.PassPriorityAroundTable(t, g)
	gap2B10AnswerName(t, g, me.ID, "Grizzly Bears")
	effectstest.PassPriorityAroundTable(t, g)
	if !gap2B10Card(t, g, bears).Goaded() {
		t.Error("chapter I did not goad the named creature")
	}
	if gap2B10Card(t, g, elves).Goaded() {
		t.Error("chapter I goaded a creature with another name")
	}
	if got := gap2B10Notes(g, saga); !slices.Equal(got, []string{"Grizzly Bears"}) {
		t.Errorf("notes after chapter I = %v", got)
	}

	gap2B10NextLore(t, g, me.ID, saga)
	gap2B10AnswerName(t, g, me.ID, "Llanowar Elves")
	effectstest.PassPriorityAroundTable(t, g)
	if !gap2B10Card(t, g, bears).Goaded() || !gap2B10Card(t, g, elves).Goaded() {
		t.Error("chapter II did not goad both names chosen for the Saga")
	}
	if got := gap2B10Notes(g, saga); !slices.Equal(got, []string{"Grizzly Bears", "Llanowar Elves"}) {
		t.Errorf("notes after chapter II = %v", got)
	}

	gap2B10NextLore(t, g, me.ID, saga)
	gap2B10AnswerName(t, g, me.ID, "Mutavault")
	effectstest.PassPriorityAroundTable(t, g)
	if gap2B10Card(t, g, land).Goaded() {
		t.Error("a land card's name goaded an animated land: not a creature card name")
	}
}

// --- Long List of the Ents --------------------------------------------------

func TestGap2B10LongListNotesATypeAndPromisesACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	saga := effectstest.EnterFromHand(t, g, me.ID, "Long List of the Ents", "Enchantment — Saga", gap2B10LongList)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerCreatureType(t, g, me.ID, "Elf")
	effectstest.PassPriorityAroundTable(t, g)
	if got := gap2B10Notes(g, saga); !slices.Equal(got, []string{"Elf"}) {
		t.Fatalf("notes = %v, want [Elf]", got)
	}

	bear := gap2B10CastCreature(t, g, me, "Bear", gap2B10BearTL)
	effectstest.PassPriorityAroundTable(t, g)
	elf := gap2B10CastCreature(t, g, me, "Elf", "Creature — Elf Warrior")
	effectstest.PassPriorityAroundTable(t, g)
	second := gap2B10CastCreature(t, g, me, "Second Elf", "Creature — Elf")
	effectstest.PassPriorityAroundTable(t, g)
	if c := gap2B10Card(t, g, bear); c.Counters[game.CounterPlusOne] != 0 {
		t.Errorf("a Bear got %d counters from an Elf promise", c.Counters[game.CounterPlusOne])
	}
	if c := gap2B10Card(t, g, elf); c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("the next Elf got %d counters, want 1", c.Counters[game.CounterPlusOne])
	}
	if c := gap2B10Card(t, g, second); c.Counters[game.CounterPlusOne] != 0 {
		t.Errorf("a second Elf got %d counters, want 0", c.Counters[game.CounterPlusOne])
	}

	// Chapter II: Elf has been noted and can't be chosen again.
	effectstest.HobA9NextChapter(t, g)
	c := gap2B10Pending(g, game.PendingChoiceCreatureType)
	if c == nil {
		t.Fatal("chapter II asked no creature type")
	}
	if !slices.Contains(c.ExcludedTypes, "Elf") {
		t.Errorf("chapter II excludes %v, want Elf among them", c.ExcludedTypes)
	}
	if err := g.ResolveCreatureTypeChoice(c.ID, me.ID, "Elf"); err == nil {
		t.Error("a noted type was accepted again")
	}
	effectstest.AnswerCreatureType(t, g, me.ID, "Treefolk")
	effectstest.PassPriorityAroundTable(t, g)
	if got := gap2B10Notes(g, saga); !slices.Equal(got, []string{"Elf", "Treefolk"}) {
		t.Errorf("notes after chapter II = %v", got)
	}
}

// --- Volo, Itinerant Scholar ------------------------------------------------

func TestGap2B10VoloJournalNotesTypesAndVoloDrawsForThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	volo := effectstest.EnterFromHand(t, g, me.ID, "Volo, Itinerant Scholar", "Legendary Creature — Human Wizard", gap2B10Volo)
	effectstest.PassPriorityAroundTable(t, g)
	var journal game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Volo's Journal" {
			journal = c
		}
	}
	if journal.InstanceID == uuid.Nil || journal.Controller != me.ID {
		t.Fatal("Volo made no Journal")
	}
	if !journal.IsLegendary() || !journal.IsArtifact() || !journal.HasSubtype("Book") || len(journal.Colors) != 0 {
		t.Errorf("the Journal is %q %v", journal.TypeLine, journal.Colors)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, journal.InstanceID, "hexproof") {
		t.Error("the Journal has no hexproof")
	}

	gap2B10CastCreature(t, g, me, "Elf Druid", "Creature — Elf Druid")
	effectstest.PassPriorityAroundTable(t, g)
	pick := gap2B10Pending(g, game.PendingChoiceOptionPick)
	if pick == nil || len(pick.PickOptions) != 2 {
		t.Fatalf("the Journal's note offered %+v, want Elf and Druid", pick)
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	// Elf is noted: an Elf Warrior leaves only Warrior, noted without a question.
	gap2B10CastCreature(t, g, me, "Elf Warrior", "Creature — Elf Warrior")
	effectstest.PassPriorityAroundTable(t, g)
	if got := gap2B10Notes(g, journal.InstanceID); !slices.Equal(got, []string{"Elf", "Warrior"}) {
		t.Fatalf("Journal notes = %v, want [Elf Warrior]", got)
	}

	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == volo {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	hand := me.Hand.Size()
	if err := effectstest.CrtActivateRaw(t, g, me.ID, volo, 0,
		[]game.TargetRef{{Kind: game.TargetCard, ID: journal.InstanceID}}, "C", "C"); err != nil {
		t.Fatalf("activate Volo: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("Volo drew %d, want 2 (two noted types)", got)
	}
}

func TestGap2B10VoloCanOnlyTargetAJournal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	volo := effectstest.B12Push(g, me.ID, "Volo, Itinerant Scholar", "Legendary Creature — Human Wizard", gap2B10Volo, 2, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", gap2B10BearTL, 2, 2)
	if err := effectstest.CrtActivateRaw(t, g, me.ID, volo, 0,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}}, "C", "C"); err == nil {
		t.Error("Volo targeted a permanent not named Volo's Journal")
	}
}

// --- Me, the Immortal -------------------------------------------------------

const gap2B10Me = "26ad696e-9f17-4110-9c59-b9211e662900"

func TestGap2B10MeTheImmortalPutsTheChosenCounterOnItself(t *testing.T) {
	spec, ok := effects.Lookup(gap2B10Me)
	if !ok || spec.Name != "Me, the Immortal" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Me, the Immortal: registered %v, %+v", ok, spec.Name)
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, "Me, the Immortal", "Legendary Creature — Human Rogue", gap2B10Me, 3, 3)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if c := gap2B10Card(t, g, id); c.Counters[game.CounterMenace] != 1 || len(c.Counters) != 1 {
		t.Errorf("counters = %v, want one menace counter", c.Counters)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, id, "menace") {
		t.Error("a menace counter did not give menace")
	}
}

func TestGap2B10MeTheImmortalKeepsItsCountersThroughAGraveyardRecast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Me, the Immortal", "Legendary Creature — Human Rogue", gap2B10Me, 3, 3)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(id, game.CounterPlusOne, 2); err != nil {
			t.Fatal(err)
		}
		if err := g.AddCounterForEffect(id, game.CounterVigilance, 1); err != nil {
			t.Fatal(err)
		}
	})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, id); err != nil {
		t.Fatal(err)
	}
	if !effectstest.InGraveyard(me, id) {
		t.Fatal("Me is not in the graveyard")
	}
	for _, c := range me.Graveyard.Cards {
		if c.InstanceID == id && (c.Counters[game.CounterPlusOne] != 2 || c.Counters[game.CounterVigilance] != 1) {
			t.Errorf("counters in the graveyard = %v", c.Counters)
		}
	}
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == id {
				me.Graveyard.Cards[i].ManaCost = "{2}{G}{U}{R}"
			}
		}
	})
	me.Hand.Cards = nil
	var hand []uuid.UUID
	for i := 0; i < 3; i++ {
		h := uuid.New()
		hand = append(hand, h)
		me.Hand.PushTop(game.Card{InstanceID: h, Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	effectstest.PayMana(t, g, me, "{C}{C}{G}{U}{R}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "discard_from_graveyard", AltCostIDs: hand[:1], Strict: true,
	}); err == nil {
		t.Fatal("one discarded card paid a two-card cost")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "discard_from_graveyard", AltCostIDs: hand[:2], Strict: true,
	}); err != nil {
		t.Fatalf("graveyard cast: %v", err)
	}
	if me.Hand.Size() != 1 {
		t.Errorf("hand %d, want 1 after discarding two", me.Hand.Size())
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, id) == nil {
		t.Fatal("Me did not resolve onto the battlefield")
	}
	if c := gap2B10Card(t, g, id); c.Counters[game.CounterPlusOne] != 2 || c.Counters[game.CounterVigilance] != 1 {
		t.Errorf("counters back on the battlefield = %v, want two +1/+1 and a vigilance", c.Counters)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, id); err != nil {
		t.Fatal(err)
	}
	for _, c := range me.Hand.Cards {
		if c.InstanceID == id && len(c.Counters) != 0 {
			t.Errorf("counters in the hand = %v, want none", c.Counters)
		}
	}
}
