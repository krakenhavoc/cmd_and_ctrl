package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fdn_de_cards_test.go — Foundations, D–E (#4465): the cards from that
// stretch of the alphabet that earlier batches left and that existing
// seams express. Each test plays the real catalog entry through the
// engine and asserts what its text says.

const (
	fdnDEDeadlyPlotOracle         = "6178715c-870d-4710-b758-66e080804ee3"
	fdnDEDeadlyRiposteOracle      = "e341478e-c3f8-45d9-9430-abe3a58fa2fc"
	fdnDEDevoutDecreeOracle       = "92f8c8b1-1fd5-4e80-b8a9-51671937aaf2"
	fdnDEDragonlordsServantOracle = "faf81ae0-0b35-45ef-a50d-6cecf0c0eeeb"
	fdnDEDreadwingScavengerOracle = "bbfd0de0-3558-4833-bae4-8efae44b7fff"
	fdnDEDriverOfTheDeadOracle    = "7c179e54-3beb-4761-bddf-41fa98b082db"
	fdnDEDwynensEliteOracle       = "3d4a7dd3-9258-4bd1-adc4-08f0205e196a"
	fdnDEEagerTrufflesnoutOracle  = "7f2ba562-8d5e-4915-967e-54916c3c5876"
	fdnDEElfswornGiantOracle      = "f6abdf34-cfdc-465c-b067-cdebb659406e"
	fdnDEElvishRegrowerOracle     = "adac526d-1a85-431a-a302-0a4b8f4a0c44"
	fdnDEEnigmaDrakeOracle        = "62007675-8a0d-4211-83b2-0daf3641dedc"
	fdnDEEruditeWizardOracle      = "7d4b6613-9000-4f53-9737-fe0338f3f629"
)

func TestFdnDEEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		fdnDEDeadlyPlotOracle:         "Deadly Plot",
		fdnDEDeadlyRiposteOracle:      "Deadly Riposte",
		fdnDEDevoutDecreeOracle:       "Devout Decree",
		fdnDEDragonlordsServantOracle: "Dragonlord's Servant",
		fdnDEDreadwingScavengerOracle: "Dreadwing Scavenger",
		fdnDEDriverOfTheDeadOracle:    "Driver of the Dead",
		fdnDEDwynensEliteOracle:       "Dwynen's Elite",
		fdnDEEagerTrufflesnoutOracle:  "Eager Trufflesnout",
		fdnDEElfswornGiantOracle:      "Elfsworn Giant",
		fdnDEElvishRegrowerOracle:     "Elvish Regrower",
		fdnDEEnigmaDrakeOracle:        "Enigma Drake",
		fdnDEEruditeWizardOracle:      "Erudite Wizard",
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

func TestFdnDEKeywordBodiesCarryTheirPrintedKeywords(t *testing.T) {
	for _, c := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		keyword                string
	}{
		{"Dreadwing Scavenger", "Creature — Nightmare Bird", fdnDEDreadwingScavengerOracle, 2, 2, "flying"},
		{"Eager Trufflesnout", "Creature — Boar", fdnDEEagerTrufflesnoutOracle, 4, 2, "trample"},
		{"Elfsworn Giant", "Creature — Giant", fdnDEElfswornGiantOracle, 5, 3, "reach"},
		{"Enigma Drake", "Creature — Drake", fdnDEEnigmaDrakeOracle, 0, 4, "flying"},
	} {
		g := newCatalogGame(t)
		id := b12Push(g, g.Seats[0].ID, c.name, c.typeLine, c.oracle, c.power, c.toughness)
		assertKeywords(t, g, id, c.keyword)
	}
}

// fdnDETapped seeds a tapped vanilla creature.
func fdnDETapped(g *game.Game, owner uuid.UUID, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tapped Bear", TypeLine: "Creature — Bear",
		Power: power, Toughness: toughness, Tapped: true, Owner: owner, Controller: owner,
	})
}

// --- Deadly Plot ----------------------------------------------------------------

func TestFdnDEDeadlyPlotDestroysACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	id := handCardFull(me, "Deadly Plot", "Instant", "{3}{B}", fdnDEDeadlyPlotOracle, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0}, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: bear, Mode: 0},
	}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Graveyard.Contains(bear) {
		t.Error("the creature was not destroyed")
	}
}

func TestFdnDEDeadlyPlotReturnsAZombieTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := pushGraveyardPermanent(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	zombie := pushGraveyardPermanent(me, "Dead Zombie", "Creature — Zombie", "{1}{B}")
	id := handCardFull(me, "Deadly Plot", "Instant", "{3}{B}", fdnDEDeadlyPlotOracle, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{1}, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: bear, Mode: 0},
	}}); err == nil {
		t.Fatal("a non-Zombie creature card was accepted")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{1}, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: zombie, Mode: 0},
	}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(zombie) {
		t.Fatal("the Zombie did not return to the battlefield")
	}
	if !isTapped(g, zombie) {
		t.Error("the Zombie returned untapped")
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("the other card left the graveyard")
	}
}

// --- Deadly Riposte -------------------------------------------------------------

func TestFdnDEDeadlyRiposteHitsATappedCreatureAndGainsTwo(t *testing.T) {
	g := newCatalogGame(t)
	untapped := b12Creature(g, g.Seats[1].ID, "Untapped Bear", "Creature — Bear", 3, 3)
	if err := castCatalogSpellErr(t, g, "Deadly Riposte", "Instant", fdnDEDeadlyRiposteOracle, cardRefs(untapped)); err == nil {
		t.Fatal("an untapped creature was accepted as the target")
	}
	g = newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tapped := fdnDETapped(g, opp.ID, 3, 3)
	life := me.Life
	castCatalogSpell(t, g, "Deadly Riposte", "Instant", fdnDEDeadlyRiposteOracle, cardRefs(tapped))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) {
		t.Error("the 3/3 survived 3 damage")
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

// --- Devout Decree --------------------------------------------------------------

func TestFdnDEDevoutDecreeExilesABlackOrRedCreatureThenScries(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	green := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Green Bear", TypeLine: "Creature — Bear",
		Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	if err := castCatalogSpellErr(t, g, "Devout Decree", "Sorcery", fdnDEDevoutDecreeOracle, cardRefs(green)); err == nil {
		t.Fatal("a green creature was accepted as the target")
	}
	for _, color := range []string{"B", "R"} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		target := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Creature", TypeLine: "Creature — Horror",
			Colors: []string{color}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
		seedLibrary(me, "Top")
		castCatalogSpell(t, g, "Devout Decree", "Sorcery", fdnDEDevoutDecreeOracle, cardRefs(target))
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(target) || !g.Exile.Contains(target) {
			t.Errorf("%s: the target was not exiled", color)
		}
		sc := scryChoiceFor(g, me.ID)
		if sc == nil {
			t.Fatalf("%s: no scry 1 after the exile", color)
		}
		if len(sc.ScryCards) != 1 {
			t.Errorf("%s: scry looks at %d cards, want 1", color, len(sc.ScryCards))
		}
	}
}

// --- Dragonlord's Servant -------------------------------------------------------

func TestFdnDEDragonlordsServantDiscountsYourDragonSpellsByOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Dragonlord's Servant", "Creature — Goblin Shaman", fdnDEDragonlordsServantOracle, false)
	if got := priceInHand(t, g, me, "Big Dragon", "Creature — Dragon", "{4}{R}{R}"); got != 5 {
		t.Errorf("{4}{R}{R} Dragon: %d, want 5 ({3}{R}{R})", got)
	}
	if got := priceInHand(t, g, me, "Coloured Dragon", "Creature — Dragon", "{R}{R}"); got != 2 {
		t.Errorf("{R}{R} Dragon: %d, want 2 (coloured pips untouched)", got)
	}
	if got := priceInHand(t, g, me, "Grizzly Bears", "Creature — Bear", "{2}{G}"); got != 3 {
		t.Errorf("a non-Dragon creature: %d, want 3", got)
	}
	if got := priceInHand(t, g, opp, "Their Dragon", "Creature — Dragon", "{4}{R}{R}"); got != 6 {
		t.Errorf("an opponent's Dragon: %d, want 6", got)
	}
}

// --- Dreadwing Scavenger --------------------------------------------------------

func TestFdnDEDreadwingScavengerLootsOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Hand.Size()
	yard := me.Graveyard.Size()
	castHoldCreature(t, g, "Dreadwing Scavenger", "Creature — Nightmare Bird", fdnDEDreadwingScavengerOracle, 2, 2)
	discardFromHand(t, g, me.ID)
	g.SettleResolution()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand %d, want %d (draw one, discard one)", got, before)
	}
	if got := me.Graveyard.Size(); got != yard+1 {
		t.Errorf("graveyard %d, want %d", got, yard+1)
	}
}

func TestFdnDEDreadwingScavengerLootsWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bird := b12Push(g, me.ID, "Dreadwing Scavenger", "Creature — Nightmare Bird", fdnDEDreadwingScavengerOracle, 2, 2)
	hand := me.Hand.Size()
	finABAttack(t, g, opp.ID, bird)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d → %d, want the draw first", hand, me.Hand.Size())
	}
	discard := me.Hand.Cards[0].InstanceID
	answerDiscard(t, g, me.ID, discard)
	if me.Hand.Size() != hand || !me.Graveyard.Contains(discard) {
		t.Errorf("hand %d, want %d after the discard", me.Hand.Size(), hand)
	}
}

func TestFdnDEDreadwingScavengerThresholdGrowsAndGrantsDeathtouch(t *testing.T) {
	for _, n := range []int{6, 7} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		bird := b12Push(g, me.ID, "Dreadwing Scavenger", "Creature — Nightmare Bird", fdnDEDreadwingScavengerOracle, 2, 2)
		blbMPFillGraveyard(g, me, n)
		threshold := n >= 7
		wantP, wantT := 2, 2
		if threshold {
			wantP, wantT = 3, 3
		}
		if p, tough := ptOf(t, g, bird); p != wantP || tough != wantT {
			t.Errorf("%d cards: %d/%d, want %d/%d", n, p, tough, wantP, wantT)
		}
		if got := hasEffectiveKeyword(t, g, bird, "deathtouch"); got != threshold {
			t.Errorf("%d cards: deathtouch %v, want %v", n, got, threshold)
		}
	}
}

// --- Driver of the Dead ---------------------------------------------------------

func TestFdnDEDriverOfTheDeadReturnsACheapCreatureWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	driver := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Driver of the Dead",
		TypeLine: "Creature — Vampire", OracleID: fdnDEDriverOfTheDeadOracle, ManaCost: "{3}{B}",
		Power: 3, Toughness: 2, Owner: me.ID, Controller: me.ID})
	cheap := pushGraveyardPermanent(me, "Cheap Bear", "Creature — Bear", "{1}{G}")
	pricey := pushGraveyardPermanent(me, "Pricey Bear", "Creature — Bear", "{2}{G}")
	b15Destroy(t, g, driver)
	b04WaitForPick(t, g, me.ID)
	pick := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: pricey}); err == nil {
		t.Fatal("a mana value 3 creature card was accepted")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: driver}); err == nil {
		t.Fatal("the Driver itself (mana value 4) was accepted")
	}
	pickTriggerTarget(t, g, me.ID, cheap)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) || controllerOf(t, g, cheap) != me.ID {
		t.Error("the mana value 2 creature card did not return under your control")
	}
	if !me.Graveyard.Contains(pricey) {
		t.Error("the mana value 3 creature card left the graveyard")
	}
}

// --- Dwynen's Elite -------------------------------------------------------------

func TestFdnDEDwynensEliteMakesAnElfOnlyWithAnotherElf(t *testing.T) {
	for _, withElf := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		if withElf {
			b12Creature(g, me.ID, "Other Elf", "Creature — Elf Druid", 1, 1)
		}
		castHoldCreature(t, g, "Dwynen's Elite", "Creature — Elf Warrior", fdnDEDwynensEliteOracle, 2, 2)
		passPriorityAroundTable(t, g)
		want := 0
		if withElf {
			want = 1
		}
		if got := b27CountNamed(g, me.ID, "Elf Warrior"); got != want {
			t.Errorf("another Elf %v: %d Elf Warrior tokens, want %d", withElf, got, want)
		}
	}
}

// --- Eager Trufflesnout ---------------------------------------------------------

func TestFdnDEEagerTrufflesnoutMakesAFoodWhenItConnects(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	boar := b12Push(g, me.ID, "Eager Trufflesnout", "Creature — Boar", fdnDEEagerTrufflesnoutOracle, 4, 2)
	dealNoncombatDamage(g, me.ID, boar, opp.ID)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Food"); got != 0 {
		t.Fatalf("%d Food after noncombat damage, want 0", got)
	}
	dealCombatDamage(g, me.ID, boar, opp.ID)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Food"); got != 1 {
		t.Errorf("%d Food after combat damage to a player, want 1", got)
	}
}

// --- Elfsworn Giant -------------------------------------------------------------

func TestFdnDEElfswornGiantMakesAnElfWarriorPerLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Elfsworn Giant", "Creature — Giant", fdnDEElfswornGiantOracle, 5, 3)
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	elves := battlefieldIDsNamed(g, "Elf Warrior")
	if len(elves) != 2 {
		t.Fatalf("two lands entered: %d Elf Warriors, want 2", len(elves))
	}
	if effectivePower(t, g, elves[0]) != 1 || controllerOf(t, g, elves[0]) != me.ID {
		t.Error("want a 1/1 Elf Warrior under your control")
	}
}

// --- Elvish Regrower ------------------------------------------------------------

func TestFdnDEElvishRegrowerReturnsAPermanentCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sorcery := pushGraveyardCardTyped(me, "Old Sorcery", "Sorcery")
	rock := pushGraveyardPermanent(me, "Dead Rock", "Artifact", "{1}")
	castHoldCreature(t, g, "Elvish Regrower", "Creature — Elf Druid", fdnDEElvishRegrowerOracle, 4, 3)
	b04WaitForPick(t, g, me.ID)
	pick := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: sorcery}); err == nil {
		t.Fatal("a sorcery card was accepted")
	}
	pickTriggerTarget(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(rock) || me.Graveyard.Contains(rock) {
		t.Error("the artifact card did not return to your hand")
	}
	if !me.Graveyard.Contains(sorcery) {
		t.Error("the sorcery card left your graveyard")
	}
}

// --- Enigma Drake ---------------------------------------------------------------

func TestFdnDEEnigmaDrakeCountsInstantsAndSorceriesInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushGraveyardCardTyped(me, "Old Instant", "Instant")
	pushGraveyardCardTyped(me, "Another Instant", "Instant")
	pushGraveyardCardTyped(me, "Old Sorcery", "Sorcery")
	pushGraveyardPermanent(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	pushGraveyardCardTyped(opp, "Their Sorcery", "Sorcery")
	drake := b12Push(g, me.ID, "Enigma Drake", "Creature — Drake", fdnDEEnigmaDrakeOracle, 0, 4)
	if p, tough := ptOf(t, g, drake); p != 3 || tough != 4 {
		t.Errorf("Enigma Drake is %d/%d, want 3/4", p, tough)
	}
}

// --- Erudite Wizard -------------------------------------------------------------

func TestFdnDEEruditeWizardGrowsOnTheSecondDrawOnly(t *testing.T) {
	hobA5SecondDrawGrows(t, "Erudite Wizard", "Creature — Human Wizard", fdnDEEruditeWizardOracle)
}
