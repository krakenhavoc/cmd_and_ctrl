package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b22_cards_test.go — Secrets of Strixhaven (sos), batch 22 (#4456).
// Each test plays the real catalog entry through the engine and asserts what
// the printed text says.

const (
	socB22TackleOracle     = "8c1c519a-6edd-4204-b825-b0feca5afa42"
	socB22ColossusOracle   = "0788fdd2-6fab-460c-b054-a736df670f96"
	socB22DuelistOracle    = "d6ea0e58-7b14-425b-9135-57e149c0e47e"
	socB22CostOracle       = "b45afde2-eaaf-4174-8082-d344efa54c44"
	socB22ColonyOracle     = "20f3388f-0931-4a09-bcda-96deb01e1787"
	socB22DaydreamOracle   = "094c4486-6cae-4268-850e-ff0082e5ad10"
	socB22VirtuosoOracle   = "416c70ab-f136-43e0-b5d5-9af8d4211c01"
	socB22DiaryOracle      = "98b43c61-4563-4a29-8cf0-fb8b34f337ce"
	socB22InventoryOracle  = "96742ebc-10a6-49e6-812a-631e5a60847d"
	socB22GuidanceOracle   = "a9abed25-9256-4fc7-8afb-edb4286e18f4"
	socB22BearType         = "Creature — Bear"
	socB22InstantTypeLine  = "Instant"
	socB22SorceryTypeLine  = "Sorcery"
	socB22ManaPoolForCasts = "G"
)

func TestSocB22EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB22TackleOracle:    "Chelonian Tackle",
		socB22ColossusOracle:  "Colossus of the Blood Age",
		socB22DuelistOracle:   "Conciliator's Duelist",
		socB22CostOracle:      "Cost of Brilliance",
		socB22ColonyOracle:    "Cuboid Colony",
		socB22DaydreamOracle:  "Daydream",
		socB22VirtuosoOracle:  "Deluge Virtuoso",
		socB22DiaryOracle:     "Diary of Dreams",
		socB22InventoryOracle: "Dig Site Inventory",
		socB22GuidanceOracle:  "Dina's Guidance",
	}
	for oracle, name := range want {
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

// socB22Damage is the damage marked on a battlefield permanent.
func socB22Damage(g *game.Game, id uuid.UUID) int {
	n := -1
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				n = c.DamageMarked
			}
		}
	})
	return n
}

// --- Chelonian Tackle -------------------------------------------------------

func socB22Tackle(t *testing.T, g *game.Game, mine uuid.UUID, theirs ...uuid.UUID) {
	t.Helper()
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: mine, Slot: 0}}
	for _, id := range theirs {
		targets = append(targets, game.TargetRef{Kind: game.TargetCard, ID: id, Slot: 1})
	}
	castCatalogSpell(t, g, "Chelonian Tackle", socB22SorceryTypeLine, socB22TackleOracle, targets)
	passPriorityAroundTable(t, g)
}

func TestSocB22TackleGrowsYourCreatureThenItFightsAndKillsTheSmallerOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", socB22BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", socB22BearType, 5, 2)
	socB22Tackle(t, g, mine, theirs)
	if g.Battlefield.Contains(theirs) {
		t.Error("their 5/2 survived a fight with a 2-power creature")
	}
	if !g.Battlefield.Contains(mine) {
		t.Fatal("your creature died: the +0/+10 should have kept it alive through 5 damage")
	}
	if p, tg := socB14PT(t, g, mine); p != 2 || tg != 12 {
		t.Errorf("your creature is %d/%d, want 2/12", p, tg)
	}
	if d := socB22Damage(g, mine); d != 5 {
		t.Errorf("damage marked on your creature %d, want 5", d)
	}
}

func TestSocB22TackleWithNoSecondTargetStillPumps(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", socB22BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", socB22BearType, 2, 2)
	socB22Tackle(t, g, mine)
	if p, tg := socB14PT(t, g, mine); p != 2 || tg != 12 {
		t.Errorf("your creature is %d/%d, want 2/12", p, tg)
	}
	if d := socB22Damage(g, theirs); d != 0 {
		t.Errorf("their creature took %d damage with no second target, want 0", d)
	}
}

func TestSocB22TackleWearsOffAtEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := b12Creature(g, me.ID, "Mine", socB22BearType, 2, 2)
	socB22Tackle(t, g, mine)
	endTurn(t, g)
	if _, tg := socB14PT(t, g, mine); tg != 2 {
		t.Errorf("toughness %d on the next turn, want it back at 2", tg)
	}
}

func TestSocB22TackleRefusesYourOwnCreatureAsTheFightTargetAndAnOpponentsAsTheFirst(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", socB22BearType, 2, 2)
	ally := b12Creature(g, me.ID, "Ally", socB22BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", socB22BearType, 2, 2)
	advanceToMain(t, g)
	card := game.Card{Name: "Chelonian Tackle", TypeLine: socB22SorceryTypeLine, OracleID: socB22TackleOracle}
	card.Owner, card.Controller, card.InstanceID = me.ID, me.ID, uuid.New()
	me.Hand.PushTop(card)
	err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: ally, Slot: 1}}})
	if err == nil {
		t.Error("a creature you control was accepted as the one an opponent controls")
	}
	card2 := card
	card2.InstanceID = uuid.New()
	me.Hand.PushTop(card2)
	err = g.CastSpell(me.ID, card2.InstanceID, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: theirs, Slot: 0}}})
	if err == nil {
		t.Error("an opponent's creature was accepted as the one you control")
	}
}

// --- Colossus of the Blood Age ----------------------------------------------

func TestSocB22ColossusDrainsEachOpponentForThreeWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}
	castCatalogSpell(t, g, "Colossus of the Blood Age", "Artifact Creature — Construct", socB22ColossusOracle, nil)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := before[p.ID] - 3
		if p.ID == me.ID {
			want = before[p.ID] + 3
		}
		if p.Life != want {
			t.Errorf("seat %s life %d, want %d", p.Name, p.Life, want)
		}
	}
}

func socB22KillColossus(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	colossus := b12Push(g, me.ID, "Colossus of the Blood Age", "Artifact Creature — Construct", socB22ColossusOracle, 6, 6)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Ajani's Response", socB22InstantTypeLine, socB19ResponseOracle, cardRefs(colossus))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(colossus) {
		t.Fatal("the Colossus survived")
	}
	return colossus
}

func TestSocB22ColossusDiesDiscardAnyNumberThenDrawThatManyPlusOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	a := ssCard(me.Hand, me.ID, "Card A", socB22SorceryTypeLine, 0, 0)
	b := ssCard(me.Hand, me.ID, "Card B", socB22InstantTypeLine, 0, 0)
	ssCard(me.Hand, me.ID, "Card C", socB22InstantTypeLine, 0, 0)
	socB22KillColossus(t, g, me)
	c := discardChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the dies trigger asks which cards to discard")
	}
	if c.ChooseMin != 0 || c.ChooseMax != 3 {
		t.Errorf("any number: bounds [%d,%d], want [0,3]", c.ChooseMin, c.ChooseMax)
	}
	answerDiscard(t, g, me.ID, a, b)
	if !me.Graveyard.Contains(a) || !me.Graveyard.Contains(b) {
		t.Fatal("the two chosen cards were not discarded")
	}
	if got := me.Hand.Size(); got != 1+3 {
		t.Errorf("hand %d, want 4 (one kept, three drawn for two discarded)", got)
	}
}

func TestSocB22ColossusDiesDiscardingNothingStillDrawsOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	ssCard(me.Hand, me.ID, "Card A", socB22SorceryTypeLine, 0, 0)
	socB22KillColossus(t, g, me)
	answerDiscard(t, g, me.ID)
	if got := me.Hand.Size(); got != 2 {
		t.Errorf("hand %d, want 2 (kept one, drew one)", got)
	}
}

func TestSocB22ColossusDiesWithAnEmptyHandDrawsOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	socB22KillColossus(t, g, me)
	if got := me.Hand.Size(); got != 1 {
		t.Errorf("hand %d, want 1 (nothing to discard, one drawn)", got)
	}
}

// --- Conciliator's Duelist --------------------------------------------------

func TestSocB22DuelistDrawsThenEveryPlayerLosesOneLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}
	castCatalogSpell(t, g, "Conciliator's Duelist", "Creature — Kor Warlock", socB22DuelistOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d (one card drawn)", got, hand+1)
	}
	for _, p := range g.Seats {
		if p.Life != before[p.ID]-1 {
			t.Errorf("seat %s life %d, want %d", p.Name, p.Life, before[p.ID]-1)
		}
	}
}

// socB22DuelistOnBoard puts the Duelist out and returns it with a creature
// of each controller to blink.
func socB22DuelistOnBoard(g *game.Game) (duelist, mine, theirs uuid.UUID) {
	me, opp := g.Seats[0], g.Seats[1]
	duelist = b12Push(g, me.ID, "Conciliator's Duelist", "Creature — Kor Warlock", socB22DuelistOracle, 4, 3)
	mine = b12Creature(g, me.ID, "Mine", socB22BearType, 2, 2)
	theirs = b12Creature(g, opp.ID, "Theirs", socB22BearType, 2, 2)
	return duelist, mine, theirs
}

func TestSocB22DuelistReparteeExilesTheCreatureAndReturnsItAtTheNextEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	duelist, _, theirs := socB22DuelistOnBoard(g)
	advanceToMain(t, g)
	// Ajani's Response targets a creature, so Repartee triggers.
	castCatalogSpell(t, g, "Ajani's Response", socB22InstantTypeLine, socB19ResponseOracle, cardRefs(theirs))
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, theirs)
	if n := triggersOnStackFrom(g, duelist); n != 1 {
		t.Fatalf("Repartee triggers %d, want 1", n)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("their creature is still on the battlefield")
	}
	// Ajani's Response would destroy it, but the trigger resolved first and
	// exiled it, so the destroy fizzled and nothing died.
	if z := g.FindCardZoneForEffect(theirs); z == nil || z.Kind != game.ZoneExile {
		t.Fatalf("their creature is in %v, want exile", z)
	}
	for g.Turn.Step != game.StepEnd {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Theirs" && c.Controller == g.Seats[1].ID {
			found = true
		}
	}
	if !found {
		t.Error("their creature did not return under its owner's control at the beginning of the next end step")
	}
}

func TestSocB22DuelistReparteeIgnoresASpellThatTargetsNoCreature(t *testing.T) {
	g := newCatalogGame(t)
	duelist, _, _ := socB22DuelistOnBoard(g)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Sign in Blood", socB22SorceryTypeLine, "c6207f6a-a624-4754-88f5-dbe700c841ff",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[0].ID}})
	if n := triggersOnStackFrom(g, duelist); n != 0 {
		t.Errorf("Repartee triggered %d times for a spell that targets only a player", n)
	}
}

func TestSocB22DuelistReparteeCanChooseNoTargetAndThenExilesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	duelist, mine, theirs := socB22DuelistOnBoard(g)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Ajani's Response", socB22InstantTypeLine, socB19ResponseOracle, cardRefs(theirs))
	b04WaitForPick(t, g, me.ID)
	// "Up to one": decline to pick. The trigger exiles nothing.
	p := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
		t.Fatalf("declining the pick: %v", err)
	}
	_ = duelist
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) {
		t.Error("your own creature left the battlefield with no target chosen")
	}
	if z := g.FindCardZoneForEffect(theirs); z != nil && z.Kind == game.ZoneExile {
		t.Error("their creature was exiled although no target was chosen")
	}
}

// --- Cost of Brilliance -----------------------------------------------------

func TestSocB22CostOfBrillianceDrawsTwoLosesTwoAndPutsACounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", socB22BearType, 2, 2)
	hand, life := opp.Hand.Size(), opp.Life
	castCatalogSpell(t, g, "Cost of Brilliance", socB22SorceryTypeLine, socB22CostOracle, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0},
		{Kind: game.TargetCard, ID: bear, Slot: 1},
	})
	passPriorityAroundTable(t, g)
	if got := opp.Hand.Size(); got != hand+2 {
		t.Errorf("target player's hand %d, want %d", got, hand+2)
	}
	if opp.Life != life-2 {
		t.Errorf("target player's life %d, want %d", opp.Life, life-2)
	}
	if n := countersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("+1/+1 counters on the creature %d, want 1 (any creature, not only yours)", n)
	}
	_ = me
}

func TestSocB22CostOfBrillianceOnYourselfWithNoCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand, life := me.Hand.Size(), me.Life
	castCatalogSpell(t, g, "Cost of Brilliance", socB22SorceryTypeLine, socB22CostOracle, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: me.ID, Slot: 0},
	})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d, want %d", got, hand+2)
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
}

// --- Cuboid Colony ----------------------------------------------------------

func TestSocB22ColonyHasFlashFlyingTrampleAndIncrement(t *testing.T) {
	spec, ok := Lookup(socB22ColonyOracle)
	if !ok {
		t.Fatal("Cuboid Colony is not registered")
	}
	want := map[string]bool{"flash": false, "flying": false, "trample": false, KeywordIncrement: false}
	for _, k := range spec.PrintedKeywords {
		if _, ok := want[k]; ok {
			want[k] = true
		}
	}
	for k, seen := range want {
		if !seen {
			t.Errorf("missing printed keyword %q", k)
		}
	}
}

func TestSocB22ColonyIncrementsWhenTheManaSpentExceedsItsPowerOrToughness(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	colony := b12Push(g, me.ID, "Cuboid Colony", "Creature — Insect", socB22ColonyOracle, 1, 1)
	advanceToMain(t, g)
	cast := func(cost string, n int) {
		t.Helper()
		if err := socB19CastWithPool(g, me, game.Card{Name: "Spell", TypeLine: socB22SorceryTypeLine, ManaCost: cost}, socB22ManaPoolForCasts, n, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast %s: %v", cost, err)
		}
		passPriorityAroundTable(t, g)
	}
	cast("{G}", 1) // 1 is not greater than 1/1
	if n := countersOn(g, colony, game.CounterPlusOne); n != 0 {
		t.Fatalf("counters %d after a one-mana spell, want 0", n)
	}
	cast("{1}{G}", 2) // 2 > 1
	if n := countersOn(g, colony, game.CounterPlusOne); n != 1 {
		t.Fatalf("counters %d after a two-mana spell, want 1", n)
	}
	cast("{1}{G}", 2) // 2 is not greater than 2/2
	if n := countersOn(g, colony, game.CounterPlusOne); n != 1 {
		t.Fatalf("counters %d after a second two-mana spell, want still 1", n)
	}
}

// --- Daydream ---------------------------------------------------------------

func TestSocB22DaydreamBlinksYourCreatureBackWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", socB22BearType, 2, 2)
	castCatalogSpell(t, g, "Daydream", socB22SorceryTypeLine, socB22DaydreamOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	var back *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Bear" && g.Battlefield.Cards[i].Controller == me.ID {
			back = &g.Battlefield.Cards[i]
		}
	}
	if back == nil {
		t.Fatal("the Bear did not return to the battlefield")
	}
	if back.InstanceID == bear {
		t.Error("the returned card kept its old object identity (CR 400.7)")
	}
	if n := countersOn(g, back.InstanceID, game.CounterPlusOne); n != 1 {
		t.Errorf("+1/+1 counters on the returned card %d, want 1", n)
	}
}

func TestSocB22DaydreamReturnsAStolenCreatureToItsOwner(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := uuid.New()
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: id, Name: "Stolen", TypeLine: socB22BearType,
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: me.ID})
	castCatalogSpell(t, g, "Daydream", socB22SorceryTypeLine, socB22DaydreamOracle, cardRefs(id))
	passPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Stolen" && c.Controller != opp.ID {
			t.Errorf("Stolen came back under %s's control, want its owner's", c.Controller)
		}
	}
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Stolen" {
			found = true
		}
	}
	if !found {
		t.Fatal("Stolen did not return")
	}
}

func TestSocB22DaydreamFlashbackBlinksAgainAndExilesTheCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", socB22BearType, 2, 2)
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Daydream", TypeLine: socB22SorceryTypeLine, ManaCost: "{W}",
		OracleID: socB22DaydreamOracle, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	fillPoolColored(me, "W", 3)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: string(game.ZoneGraveyard), AlternativeCost: "flashback",
		Strict: true, Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("flash it back: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(id) {
		t.Error("the flashed-back Daydream returned to the graveyard instead of being exiled")
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the original Bear object is still on the battlefield")
	}
}

// --- Deluge Virtuoso --------------------------------------------------------

func TestSocB22VirtuosoTapsAndStunsTargetOpponentCreatureWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Theirs", socB22BearType, 2, 2)
	castAndResolveCreature(t, g, "Deluge Virtuoso", "Creature — Human Wizard", socB22VirtuosoOracle)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	c, ok := g.LookupCardForEffect(theirs)
	if !ok || !c.Tapped {
		t.Error("their creature is not tapped")
	}
	if n := countersOn(g, theirs, game.CounterStun); n != 1 {
		t.Errorf("stun counters %d, want 1", n)
	}
}

func TestSocB22VirtuosoCannotTargetYourOwnCreatureWithItsEntersTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := b12Creature(g, me.ID, "Mine", socB22BearType, 2, 2)
	castAndResolveCreature(t, g, "Deluge Virtuoso", "Creature — Human Wizard", socB22VirtuosoOracle)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
			t.Error("your own creature was accepted as a target for \"an opponent controls\"")
		}
	}
}

func TestSocB22VirtuosoOpusIsPlusOneOrPlusTwoByTheManaSpent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	v := b12Push(g, me.ID, "Deluge Virtuoso", "Creature — Human Wizard", socB22VirtuosoOracle, 2, 2)
	advanceToMain(t, g)
	if err := socB19CastWithPool(g, me, game.Card{Name: "Small", TypeLine: socB22SorceryTypeLine, ManaCost: "{3}{G}"}, "G", 4, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tg := socB14PT(t, g, v); p != 3 || tg != 3 {
		t.Errorf("after a four-mana sorcery it is %d/%d, want 3/3", p, tg)
	}
	if err := socB19CastWithPool(g, me, game.Card{Name: "Big", TypeLine: socB22InstantTypeLine, ManaCost: "{4}{G}"}, "G", 5, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tg := socB14PT(t, g, v); p != 5 || tg != 5 {
		t.Errorf("after a five-mana instant as well it is %d/%d, want 5/5 (+1/+1 then +2/+2, never both for one spell)", p, tg)
	}
	endTurn(t, g)
	if p, tg := socB14PT(t, g, v); p != 2 || tg != 2 {
		t.Errorf("next turn it is %d/%d, want 2/2", p, tg)
	}
}

func TestSocB22VirtuosoOpusIgnoresACreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	v := b12Push(g, me.ID, "Deluge Virtuoso", "Creature — Human Wizard", socB22VirtuosoOracle, 2, 2)
	advanceToMain(t, g)
	if err := socB19CastWithPool(g, me, game.Card{Name: "Bear", TypeLine: socB22BearType, ManaCost: "{1}{G}", Power: 2, Toughness: 2}, "G", 2, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tg := socB14PT(t, g, v); p != 2 || tg != 2 {
		t.Errorf("after a creature spell it is %d/%d, want 2/2", p, tg)
	}
}

// --- Diary of Dreams --------------------------------------------------------

func socB22Diary(g *game.Game, me *game.Player) uuid.UUID {
	return b12Push(g, me.ID, "Diary of Dreams", "Artifact — Book", socB22DiaryOracle, 0, 0)
}

func TestSocB22DiaryGetsAPageCounterForEachInstantOrSorceryYouCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	diary := socB22Diary(g, me)
	advanceToMain(t, g)
	if err := socB19CastWithPool(g, me, game.Card{Name: "Sorcery", TypeLine: socB22SorceryTypeLine, ManaCost: "{G}"}, "G", 1, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if err := socB19CastWithPool(g, me, game.Card{Name: "Instant", TypeLine: socB22InstantTypeLine, ManaCost: "{G}"}, "G", 1, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if err := socB19CastWithPool(g, me, game.Card{Name: "Bear", TypeLine: socB22BearType, ManaCost: "{G}", Power: 1, Toughness: 1}, "G", 1, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, diary, "page"); n != 2 {
		t.Errorf("page counters %d, want 2 (the creature spell does not count)", n)
	}
}

func TestSocB22DiaryDrawCostsOneLessPerPageCounterAndStillTaps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	diary := socB22Diary(g, me)
	advanceToMain(t, g)
	// No counters: {5} is needed, and four is not enough.
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	fillPoolColored(me, "G", 4)
	if err := g.ActivateCatalogAbility(me.ID, diary, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with four mana and no page counters")
	}
	addCountersForTest(g, diary, "page", 3)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	fillPoolColored(me, "G", 2)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, diary, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate with two mana and three page counters: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
	c, _ := g.LookupCardForEffect(diary)
	if !c.Tapped {
		t.Error("the Diary did not tap for its cost")
	}
}

func TestSocB22DiaryCostNeverGoesBelowTheTapWithManyCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	diary := socB22Diary(g, me)
	addCountersForTest(g, diary, "page", 7)
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, diary, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate for free with seven page counters: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

// --- Dig Site Inventory -----------------------------------------------------

func TestSocB22InventoryPutsACounterAndGivesVigilanceUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", socB22BearType, 2, 2)
	castCatalogSpell(t, g, "Dig Site Inventory", socB22SorceryTypeLine, socB22InventoryOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if n := countersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("+1/+1 counters %d, want 1", n)
	}
	if !socB22HasKeyword(g, bear, "vigilance") {
		t.Error("the creature did not gain vigilance")
	}
	endTurn(t, g)
	if socB22HasKeyword(g, bear, "vigilance") {
		t.Error("vigilance lasted past the end of the turn")
	}
	if n := countersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("the counter is permanent: %d, want 1", n)
	}
}

func socB22HasKeyword(g *game.Game, id uuid.UUID, kw string) bool {
	c, ok := g.LookupCardForEffect(id)
	return ok && game.HasKeyword(&c, kw)
}

func TestSocB22InventoryFlashbackWorksOnceAndExilesIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", socB22BearType, 2, 2)
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Dig Site Inventory", TypeLine: socB22SorceryTypeLine, ManaCost: "{W}",
		OracleID: socB22InventoryOracle, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	fillPoolColored(me, "W", 1)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: string(game.ZoneGraveyard), AlternativeCost: "flashback",
		Strict: true, Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("flash it back: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("+1/+1 counters %d, want 1", n)
	}
	if me.Graveyard.Contains(id) {
		t.Error("the flashed-back card returned to the graveyard instead of being exiled")
	}
}

func TestSocB22InventoryRefusesAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Theirs", socB22BearType, 2, 2)
	err := castCatalogSpellErr(t, g, "Dig Site Inventory", socB22SorceryTypeLine, socB22InventoryOracle, cardRefs(theirs))
	if err == nil {
		t.Error("an opponent's creature was accepted as the target")
	}
}

// --- Dina's Guidance --------------------------------------------------------

func TestSocB22GuidanceToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	wurm := stapleLibraryCard(me, "Wurm", "Creature — Wurm")
	stapleLibraryCard(me, "Rock", "Artifact")
	castCatalogSpell(t, g, "Dina's Guidance", socB22InstantTypeLine, socB22GuidanceOracle, nil)
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, wurm)
	answerOptionPick(t, g, me.ID, 0)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(wurm) {
		t.Error("the creature card did not reach the hand")
	}
	if me.Library.Contains(wurm) {
		t.Error("the creature card is still in the library")
	}
}

func TestSocB22GuidanceToGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	wurm := stapleLibraryCard(me, "Wurm", "Creature — Wurm")
	castCatalogSpell(t, g, "Dina's Guidance", socB22InstantTypeLine, socB22GuidanceOracle, nil)
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, wurm)
	answerOptionPick(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(wurm) {
		t.Error("the creature card did not reach the graveyard")
	}
	if me.Hand.Contains(wurm) {
		t.Error("the creature card went to the hand as well")
	}
}

func TestSocB22GuidanceOnlyOffersCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rock := stapleLibraryCard(me, "Rock", "Artifact")
	wurm := stapleLibraryCard(me, "Wurm", "Creature — Wurm")
	castCatalogSpell(t, g, "Dina's Guidance", socB22InstantTypeLine, socB22GuidanceOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range c.SearchCards {
		offered[id] = true
	}
	if offered[rock] {
		t.Error("a noncreature card was offered")
	}
	if !offered[wurm] {
		t.Error("the creature card was not offered")
	}
}
