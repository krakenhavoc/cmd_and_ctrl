package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mom_b22_test.go — March of the Machine batch 22: the two Praetor
// Sagas (Urabrask, Vorinclex), Mutagen Connoisseur, Oculus Whelp, Overgrown
// Pest and the Invasions of Pyrulea and Ixalan (S82, #5477).

const (
	b22Pest        = "ce84891f-07e1-4c72-a7b8-93b6afd1293e"
	b22Connoisseur = "07b4f5b5-47ef-43a5-af71-51693d775d4b"
	b22Whelp       = "89ff1b0a-0cf3-455e-a381-1d93ef86a41b"
	b22Pyrulea     = "323437f5-cc6c-4579-b591-499476dffb3c"
	b22Ixalan      = "1e5a5a7d-2c34-4e56-96e9-ebcf7745a03c"
)

func b22Praetor(owner uuid.UUID, oracle, front, back string, transformed bool) game.Card {
	return praetorCard(owner, oracle, front, "Legendary Creature — Phyrexian Praetor", 4, 4, back, "Enchantment — Saga", transformed)
}

func b22Push(g *game.Game, c game.Card) uuid.UUID { return pushBattlefieldCardWithTimestamp(g, c) }

func b22Card(g *game.Game, id uuid.UUID) *game.Card { return battlefieldCardFor(g, id) }

func b22HasKeyword(g *game.Game, id uuid.UUID, kw string) bool {
	var c game.Card
	var ok bool
	g.ReadSnapshot(func() { c, ok = g.LookupCardForEffect(id) })
	return ok && game.HasKeyword(&c, kw)
}

func b22SetTally(g *game.Game, p uuid.UUID, tally game.CastTally) {
	g.WithWriteLock(func() {
		if g.SpellsCastThisTurn == nil {
			g.SpellsCastThisTurn = map[uuid.UUID]game.CastTally{}
		}
		g.SpellsCastThisTurn[p] = tally
	})
}

// b22EnterFromHand moves a hand card onto the battlefield so its enters
// triggers fire.
func b22EnterFromHand(t *testing.T, g *game.Game, owner, id uuid.UUID) {
	t.Helper()
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: owner}, game.ZoneRef{Kind: game.ZoneBattlefield}, id); err != nil {
		t.Fatalf("enter the battlefield: %v", err)
	}
}

func TestS82MomB22AllFacesAreRegisteredFull(t *testing.T) {
	for _, id := range []string{
		urabraskOracleID, urabraskOracleID + "#1",
		vorinclexOracleID, vorinclexOracleID + "#1",
		b22Pest, b22Connoisseur, b22Whelp,
		b22Pyrulea, b22Pyrulea + "#1", b22Ixalan, b22Ixalan + "#1",
	} {
		spec, ok := Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s (%s) is %v, want Full", spec.Name, id, spec.Completeness)
		}
		if len(spec.Caveats) != 0 {
			t.Errorf("%s has caveats %v", spec.Name, spec.Caveats)
		}
	}
}

// --- Mutagen Connoisseur / Oculus Whelp / Slabhorn: "transformed permanent" ---

func TestMutagenConnoisseurCountsOnlyYourTransformedPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	conn := b22Push(g, game.Card{InstanceID: uuid.New(), OracleID: b22Connoisseur, Name: "Mutagen Connoisseur",
		TypeLine: "Creature — Vedalken Mutant", Power: 0, Toughness: 5, Owner: me.ID, Controller: me.ID})
	if got := gaLayered(t, g, conn).Effective().Power; got != 0 {
		t.Fatalf("alone: power %d, want 0", got)
	}
	// A double-faced permanent with its front face up is not transformed.
	b22Push(g, b22Praetor(me.ID, urabraskOracleID, "Urabrask", "The Great Work", false))
	if got := gaLayered(t, g, conn).Effective().Power; got != 0 {
		t.Errorf("front-face Praetor: power %d, want 0", got)
	}
	b22Push(g, b22Praetor(me.ID, vorinclexOracleID, "Vorinclex", "The Grand Evolution", true))
	if got := gaLayered(t, g, conn).Effective().Power; got != 1 {
		t.Errorf("one transformed permanent: power %d, want 1", got)
	}
	// An opponent's transformed permanent does not count.
	b22Push(g, b22Praetor(opp.ID, jinGitaxiasOracleID, "Jin-Gitaxias", "The Great Synthesis", true))
	if got := gaLayered(t, g, conn).Effective().Power; got != 1 {
		t.Errorf("opponent's transformed permanent: power %d, want still 1", got)
	}
	if got := gaLayered(t, g, conn).Effective().Toughness; got != 5 {
		t.Errorf("toughness %d, want 5", got)
	}
}

func b22WhelpDies(t *testing.T, withTransformed bool) int {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	whelp := b22Push(g, game.Card{InstanceID: uuid.New(), OracleID: b22Whelp, Name: "Oculus Whelp",
		TypeLine: "Creature — Phyrexian Dragon", Power: 3, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if withTransformed {
		b22Push(g, b22Praetor(me.ID, vorinclexOracleID, "Vorinclex", "The Grand Evolution", true))
	}
	before := me.Hand.Size()
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{whelp}, game.DestroyOptions{}) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	return me.Hand.Size() - before
}

func TestOculusWhelpDrawsOnlyWhileYouControlATransformedPermanent(t *testing.T) {
	if got := b22WhelpDies(t, true); got != 1 {
		t.Errorf("dying with a transformed permanent drew %d, want 1", got)
	}
	if got := b22WhelpDies(t, false); got != 0 {
		t.Errorf("dying with none drew %d, want 0", got)
	}
}

func TestGargantuanSlabhornGrantsTrampleToOtherTransformedPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	slab := b22Push(g, praetorCard(me.ID, b22Pyrulea, "Invasion of Pyrulea", "Battle — Siege", 0, 0, "Gargantuan Slabhorn", "Creature — Beast", true))
	other := b22Push(g, b22Praetor(me.ID, vorinclexOracleID, "Vorinclex", "The Grand Evolution", true))
	front := b22Push(g, b22Praetor(me.ID, urabraskOracleID, "Urabrask", "The Great Work", false))
	foe := b22Push(g, b22Praetor(g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, jinGitaxiasOracleID, "Jin-Gitaxias", "The Great Synthesis", true))
	if !b22HasKeyword(g, other, "trample") {
		t.Error("another transformed permanent of yours has no trample")
	}
	if b22HasKeyword(g, front, "trample") {
		t.Error("a front-face permanent got trample")
	}
	if b22HasKeyword(g, foe, "trample") {
		t.Error("an opponent's transformed permanent got trample")
	}
	spec, _ := Lookup(b22Pyrulea + "#1")
	if len(spec.Triggered) != 2 {
		t.Errorf("Slabhorn has %d triggers, want its own ward and the grant", len(spec.Triggered))
	}
	_ = slab
}

// --- Belligerent Regisaur ---

func TestBelligerentRegisaurGainsIndestructibleWhenYouCastASpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	reg := b22Push(g, praetorCard(me.ID, b22Ixalan, "Invasion of Ixalan", "Battle — Siege", 0, 0, "Belligerent Regisaur", "Creature — Dinosaur", true))
	if b22HasKeyword(g, reg, "indestructible") {
		t.Fatal("indestructible before any spell")
	}
	castAndResolve(t, g, "Shock", "Instant", "{0}", "", nil)
	if !b22HasKeyword(g, reg, "indestructible") {
		t.Error("no indestructible after casting a spell")
	}
	// Until end of turn.
	runCleanup(t, g)
	if b22HasKeyword(g, reg, "indestructible") {
		t.Error("indestructible outlasted the turn")
	}
}

// --- Overgrown Pest ---

func b22DoubleFacedCard(owner uuid.UUID) game.Card {
	return game.Card{InstanceID: uuid.New(), Name: "Dual", Layout: game.LayoutTransform, Owner: owner, Controller: owner,
		Faces: []game.Face{{Name: "Dual", TypeLine: "Creature — Human"}, {Name: "Flip", TypeLine: "Creature — Werewolf"}}}
}

func TestOvergrownPestTakesALandOrADoubleFacedCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dfc := b22DoubleFacedCard(me.ID)
	ids := hornTopFive(me, []struct{ name, typeLine string }{
		{"Bear", "Creature — Bear"}, {"Forest", "Basic Land — Forest"}, {"Opt", "Instant"}, {"Dual", "Creature — Human"}, {"Elf", "Creature — Elf"},
	})
	// Replace the fourth card with a real double-faced one.
	for i := range me.Library.Cards {
		if me.Library.Cards[i].InstanceID == ids[3] {
			dfc.InstanceID = ids[3]
			me.Library.Cards[i] = dfc
		}
	}
	pest := pushCatalogHandCard(me, "Overgrown Pest", "Creature — Pest", b22Pest)
	b22EnterFromHand(t, g, me.ID, pest)
	passPriorityAroundTable(t, g)
	prompt := latestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("no look prompt")
	}
	want := map[uuid.UUID]bool{ids[1]: true, ids[3]: true}
	if len(prompt.ChooseCards) != 2 {
		t.Fatalf("offered %d cards, want the land and the double-faced card", len(prompt.ChooseCards))
	}
	for _, id := range prompt.ChooseCards {
		if !want[id] {
			t.Errorf("offered %s, which is neither a land nor double-faced", id)
		}
	}
	if prompt.ChooseMin != 0 || prompt.ChooseMax != 1 {
		t.Errorf("bounds [%d,%d], want [0,1]", prompt.ChooseMin, prompt.ChooseMax)
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{ids[3]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(ids[3]) {
		t.Error("the double-faced card is not in hand")
	}
	if !hornRevealed(g, ids[3]) {
		t.Error("the taken card was not revealed")
	}
	bottom := me.Library.Cards[:4]
	for _, id := range []uuid.UUID{ids[0], ids[1], ids[2], ids[4]} {
		found := false
		for _, c := range bottom {
			found = found || c.InstanceID == id
		}
		if !found {
			t.Errorf("%s was not put on the bottom", id)
		}
	}
}

// --- Invasion of Pyrulea ---

func TestInvasionOfPyruleaDrawsOnALandOrDoubleFacedTop(t *testing.T) {
	for _, tc := range []struct {
		name string
		top  func(owner uuid.UUID) game.Card
		draw int
	}{
		{"land", func(o uuid.UUID) game.Card {
			return game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: o, Controller: o}
		}, 1},
		{"double-faced", b22DoubleFacedCard, 1},
		{"other", func(o uuid.UUID) game.Card {
			return game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Owner: o, Controller: o}
		}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			top := tc.top(me.ID)
			me.Library.PushTop(top)
			before := me.Hand.Size()
			protector := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			id := importToHand(cards.Card{
				ID: uuid.New(), Name: "Invasion of Pyrulea // Gargantuan Slabhorn", Layout: "transform",
				TypeLine: "Battle — Siege // Creature — Beast", OracleID: uuid.MustParse(b22Pyrulea),
				CardFaces: []cards.CardFace{
					{Name: "Invasion of Pyrulea", TypeLine: "Battle — Siege", ManaCost: "{G}{U}", Defense: "4", Colors: []string{"G", "U"}},
					{Name: "Gargantuan Slabhorn", TypeLine: "Creature — Beast", Power: "4", Toughness: "4", Colors: []string{"G", "U"}},
				},
			}, me)
			b22EnterFromHand(t, g, me.ID, id)
			answerProtectorPrompt(t, g, me.ID, protector.ID)
			passPriorityAroundTable(t, g)
			answerScryKeepAll(t, g, me.ID)
			passPriorityAroundTable(t, g)
			if got := me.Hand.Size() - before; got != tc.draw {
				t.Errorf("drew %d, want %d", got, tc.draw)
			}
			if tc.draw == 1 && !me.Hand.Contains(top.InstanceID) {
				t.Error("the revealed card was not the one drawn")
			}
			if !hornRevealed(g, top.InstanceID) {
				t.Error("the top card was not revealed")
			}
		})
	}
}

// --- Urabrask // The Great Work ---

func b22Urabrask(g *game.Game, me *game.Player) uuid.UUID {
	return b22Push(g, b22Praetor(me.ID, urabraskOracleID, "Urabrask", "The Great Work", false))
}

func TestUrabraskDamagesAnOpponentAndAddsRedPerInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b22Urabrask(g, me)
	toMainForCost(t, g)
	life := opp.Life
	if _, err := castWithTapParams(t, g, "Shock", "Instant", "{0}", "", game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	answerTriggerTargets2(t, g, me.ID, opp.ID)
	// Resolve only the trigger, so the mana is seen before the phase ends.
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	for i := 0; i < 4 && opp.Life == life; i++ {
		_ = g.PassPriority()
	}
	if opp.Life != life-1 {
		t.Errorf("opponent life %d -> %d, want -1", life, opp.Life)
	}
	if me.ManaPool.CountColor("R") != 1 {
		t.Errorf("red mana in pool %d, want 1", me.ManaPool.CountColor("R"))
	}
}

func TestUrabraskIgnoresCreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b22Urabrask(g, me)
	if _, err := castWithTapParams(t, g, "Bear", "Creature — Bear", "{0}", "", game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a creature spell raised Urabrask's trigger")
	}
}

func TestUrabraskTransformsOnlyAfterThreeInstantsOrSorceries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	urb := b22Urabrask(g, me)
	toMainForCost(t, g)
	floatForTest(g, me, "R")
	b22SetTally(g, me.ID, game.CastTally{InstantSorcery: 2, Total: 2})
	if err := g.ActivateCatalogAbility(me.ID, urb, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("transformed after only two instants or sorceries")
	}
	b22SetTally(g, me.ID, game.CastTally{InstantSorcery: 3, Total: 3})
	if err := g.ActivateCatalogAbility(me.ID, urb, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate with three: %v", err)
	}
	passPriorityAroundTable(t, g)
	saga := praetorOnBattlefield(g, urabraskOracleID)
	if saga == nil || saga.Name != "The Great Work" {
		t.Fatalf("Urabrask did not return transformed: %+v", saga)
	}
}

// b22ToSagaChapter flips Urabrask or Vorinclex into its Saga and returns it.
func b22Flip(t *testing.T, g *game.Game, me *game.Player, oracle, front, back, mana string, tally bool) uuid.UUID {
	t.Helper()
	id := b22Push(g, b22Praetor(me.ID, oracle, front, back, false))
	toMainForCost(t, g)
	floatForTest(g, me, mana)
	if tally {
		b22SetTally(g, me.ID, game.CastTally{InstantSorcery: 3, Total: 3})
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("flip %s: %v", front, err)
	}
	for i := 0; i < 12; i++ {
		if saga := praetorOnBattlefield(g, oracle); saga != nil && saga.Name == back {
			return saga.InstanceID
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatalf("%s chapter I never reached the stack", back)
	return uuid.Nil
}

func TestTheGreatWorkChapterOneHitsOneOpponentAndTheirCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	mine := pushCreatureToBattlefieldForTest(g, me.ID, "Mine")
	theirs := pushCreatureToBattlefieldForTest(g, opp.ID, "Theirs")
	bystander := pushCreatureToBattlefieldForTest(g, other.ID, "Bystander")
	for _, id := range []uuid.UUID{mine, theirs, bystander} {
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == id {
					g.Battlefield.Cards[i].Toughness = 5
				}
			}
		})
	}
	life, otherLife := opp.Life, other.Life
	b22Flip(t, g, me, urabraskOracleID, "Urabrask", "The Great Work", "R", true)
	answerTriggerTargets2(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("target opponent life %d -> %d, want -3", life, opp.Life)
	}
	if other.Life != otherLife {
		t.Errorf("a different opponent lost life: %d -> %d", otherLife, other.Life)
	}
	damage := func(id uuid.UUID) int { return b22Card(g, id).DamageMarked }
	if damage(theirs) != 3 {
		t.Errorf("their creature has %d damage, want 3", damage(theirs))
	}
	if damage(mine) != 0 || damage(bystander) != 0 {
		t.Errorf("my creature %d / bystander %d damage, want 0 and 0", damage(mine), damage(bystander))
	}
}

func TestTheGreatWorkChapterTwoMakesThreeTreasures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b22Flip(t, g, me, urabraskOracleID, "Urabrask", "The Great Work", "R", true)
	answerTriggerTargets2(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	saga := praetorOnBattlefield(g, urabraskOracleID)
	b22SetLore(g, saga, 1)
	advanceToPrecombatMainOf(t, g, g.Turn.ActiveSeat)
	passPriorityAroundTable(t, g)
	if n := countNamed(g, "Treasure"); n != 3 {
		t.Errorf("made %d Treasures, want 3", n)
	}
}

func TestTheGreatWorkChapterThreeCastsFromAnyGraveyardAndExiles(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	inOpp := pushGraveyardTyped(opp, "Opt", "Instant")
	inMine := pushGraveyardTyped(me, "Ponder", "Sorcery")
	creature := pushGraveyardTyped(opp, "Bear", "Creature — Bear")
	b22Flip(t, g, me, urabraskOracleID, "Urabrask", "The Great Work", "R", true)
	answerTriggerTargets2(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	saga := praetorOnBattlefield(g, urabraskOracleID)
	b22SetLore(g, saga, 2)
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)

	// The Saga flickered and is the front face again.
	if c := praetorOnBattlefield(g, urabraskOracleID); c == nil || c.Name != "Urabrask" {
		t.Fatalf("the Saga did not return front face up: %+v", c)
	}
	if err := g.CastSpell(me.ID, creature, game.CastSpellParams{FromZone: "graveyard", Strict: true, AlternativeCost: game.AltCostKeyGrantedCost}); err == nil {
		t.Error("a creature card was castable from a graveyard")
	}
	if err := g.CastSpell(me.ID, inOpp, game.CastSpellParams{FromZone: "graveyard", Strict: true, AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("cast an instant from an opponent's graveyard: %v", err)
	}
	// Urabrask's own cast trigger from the first spell needs a target.
	if latestPickTarget(g, me.ID) != nil {
		answerTriggerTargets2(t, g, me.ID, opp.ID)
	}
	passPriorityAroundTable(t, g)
	if !inExile(g, inOpp) {
		t.Error("the spell cast from the graveyard did not go to exile")
	}
	if err := g.CastSpell(me.ID, inMine, game.CastSpellParams{FromZone: "graveyard", Strict: true, AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("cast a sorcery from your own graveyard: %v", err)
	}
	if latestPickTarget(g, me.ID) != nil {
		answerTriggerTargets2(t, g, me.ID, opp.ID)
	}
	passPriorityAroundTable(t, g)
	// Only until end of turn.
	runCleanup(t, g)
	later := pushGraveyardTyped(opp, "Shock", "Instant")
	if err := g.CastSpell(me.ID, later, game.CastSpellParams{FromZone: "graveyard", Strict: true, AlternativeCost: game.AltCostKeyGrantedCost}); err == nil {
		t.Error("the permission outlasted the turn")
	}
}

// --- Vorinclex // The Grand Evolution ---

func TestVorinclexSearchesForUpToTwoForests(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	f1 := pushLibraryTyped(me, "Forest", "Basic Land — Forest")
	f2 := pushLibraryTyped(me, "Stomping Ground", "Land — Mountain Forest")
	pushLibraryTyped(me, "Mountain", "Basic Land — Mountain")
	vor := pushCatalogHandCard(me, "Vorinclex", "Legendary Creature — Phyrexian Praetor", vorinclexOracleID)
	b22EnterFromHand(t, g, me.ID, vor)
	passPriorityAroundTable(t, g)
	prompt := searchChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no search prompt")
	}
	if err := g.ResolveSearchLibrary(prompt.ID, me.ID, []uuid.UUID{f1, f2}); err != nil {
		t.Fatalf("answer search: %v", err)
	}
	if !me.Hand.Contains(f1) || !me.Hand.Contains(f2) {
		t.Error("the two Forest cards are not in hand")
	}
}

func TestTheGrandEvolutionChapterOneMillsTenAndPutsTwoCreaturesOntoTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var bears []uuid.UUID
	for i := 0; i < 10; i++ {
		if i < 3 {
			bears = append(bears, pushLibraryTyped(me, "Bear", "Creature — Bear"))
		} else {
			pushLibraryTyped(me, "Island", "Basic Land — Island")
		}
	}
	libBefore := me.Library.Size()
	b22Flip(t, g, me, vorinclexOracleID, "Vorinclex", "The Grand Evolution", "GGGGGGGG", false)
	passPriorityAroundTable(t, g)
	if got := libBefore - me.Library.Size(); got != 10 {
		t.Fatalf("milled %d cards, want 10", got)
	}
	prompt := latestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("no prompt for the milled creature cards")
	}
	if len(prompt.ChooseCards) != 3 || prompt.ChooseMin != 0 || prompt.ChooseMax != 2 {
		t.Fatalf("offered %d cards in [%d,%d], want 3 creatures, up to two", len(prompt.ChooseCards), prompt.ChooseMin, prompt.ChooseMax)
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, bears[:2]); err != nil {
		t.Fatalf("answer: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range bears[:2] {
		if b22Card(g, id) == nil {
			t.Errorf("%s did not enter the battlefield", id)
		}
	}
	if b22Card(g, bears[2]) != nil {
		t.Error("a third creature entered")
	}
}

func TestTheGrandEvolutionChapterTwoDistributesSevenCounters(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	a := pushCreatureToBattlefieldForTest(g, me.ID, "A")
	b := pushCreatureToBattlefieldForTest(g, me.ID, "B")
	b22Flip(t, g, me, vorinclexOracleID, "Vorinclex", "The Grand Evolution", "GGGGGGGG", false)
	passPriorityAroundTable(t, g)
	if prompt := latestChooseCards(g, me.ID); prompt != nil {
		_ = g.ResolveChooseCards(prompt.ID, me.ID, nil)
	}
	passPriorityAroundTable(t, g)
	saga := praetorOnBattlefield(g, vorinclexOracleID)
	b22SetLore(g, saga, 1)
	advanceToPrecombatMainOf(t, g, seat)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("chapter II asked for no targets")
	}
	if err := g.ResolvePickTargetsDivided(pick.ID, me.ID,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}},
		map[uuid.UUID]int{a: 4, b: 3}); err != nil {
		t.Fatalf("answer: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := b22Card(g, a).Counters[game.CounterPlusOne]; got != 4 {
		t.Errorf("A has %d counters, want 4", got)
	}
	if got := b22Card(g, b).Counters[game.CounterPlusOne]; got != 3 {
		t.Errorf("B has %d counters, want 3", got)
	}
}

func TestTheGrandEvolutionChapterThreeGrantsTheFightAbilityAndFlickers(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	mine := pushCreatureToBattlefieldForTest(g, me.ID, "Fighter")
	theirs := pushCreatureToBattlefieldForTest(g, opp.ID, "Victim")
	b22Flip(t, g, me, vorinclexOracleID, "Vorinclex", "The Grand Evolution", "GGGGGGGG", false)
	passPriorityAroundTable(t, g)
	if prompt := latestChooseCards(g, me.ID); prompt != nil {
		_ = g.ResolveChooseCards(prompt.ID, me.ID, nil)
	}
	passPriorityAroundTable(t, g)
	saga := praetorOnBattlefield(g, vorinclexOracleID)
	b22SetLore(g, saga, 2)
	advanceToPrecombatMainOf(t, g, seat)
	if pick := latestPickTarget(g, me.ID); pick != nil {
		_ = g.ResolvePickTargets(pick.ID, me.ID, nil)
	}
	passPriorityAroundTable(t, g)
	b22SetLore(g, saga, 2)
	passPriorityAroundTable(t, g)

	if c := praetorOnBattlefield(g, vorinclexOracleID); c == nil || c.Name != "Vorinclex" {
		t.Fatalf("the Saga did not return front face up: %+v", c)
	}
	abilities := game.ActivatedAbilitiesForCard(*b22Card(g, mine))
	_ = abilities
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			switch g.Battlefield.Cards[i].InstanceID {
			case mine:
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 3, 3
			case theirs:
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 1, 2
			}
		}
	})
	floatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, mine, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}}); err != nil {
		t.Fatalf("activate the granted fight ability: %v", err)
	}
	if pick := latestPickTarget(g, me.ID); pick != nil {
		if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}); err != nil {
			t.Fatalf("pick the fight target: %v", err)
		}
	}
	// The returned Vorinclex searches again as it enters; take nothing.
	if search := searchChoiceFor(g, me.ID); search != nil {
		if err := g.ResolveSearchLibrary(search.ID, me.ID, nil); err != nil {
			t.Fatalf("decline the search: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if b22Card(g, theirs) != nil {
		t.Error("the fought creature survived 3 damage")
	}
	// Both creatures are 2/2 here, so the fight kills both.
	if b22Card(g, mine) != nil {
		t.Errorf("the fighter survived the fight, it should have taken 2 damage")
	}
}

// pushLibraryTyped puts a card on top of a library.
func pushLibraryTyped(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

func b22SetLore(g *game.Game, saga *game.Card, n int) {
	g.WithWriteLock(func() {
		if saga.Counters == nil {
			saga.Counters = map[string]int{}
		}
		saga.Counters[game.CounterLore] = n
	})
}
