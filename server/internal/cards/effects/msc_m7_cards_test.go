package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_m7_cards_test.go — Marvel Super Heroes Commander, M–Z batch 7 (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	mscM7PatriotOracle   = "6c44d62f-52de-446c-9824-6b474534a612"
	mscM7NakiaOracle     = "3f2ab9ff-10b8-494a-be7a-2fce172f62a3"
	mscM7OkoyeOracle     = "8fbfc5b1-2146-4d07-a5fe-5f025da31bd6"
	mscM7PortalOracle    = "8430ae03-2d4c-4031-aa29-4d53815ded78"
	mscM7SpiderRexOracle = "58b771d4-dfaf-4189-a5f6-4bc0fb4c0fa3"
	mscM7GwenOracle      = "2648559d-1d96-47d2-a769-cfa2cbca2be5"
	mscM7WebSpinnerOrcl  = "d586ef58-ca9e-4c6d-94d6-3b8d782b11e1"
	mscM7MobLookoutOrcl  = "47654b36-b215-4333-aa70-d88d627447e3"
	mscM7RamaTutOracle   = "2c02c8e3-943a-4c20-9dad-1408793a825e"
	mscM7StatureOracle   = "b1eee67f-30cf-4afb-8080-84ac282effc4"
	mscM7MjolnirOracle   = "8c944733-3fb5-4d1f-af1a-3fd31377eb9a"
)

func TestMscM7EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mscM7PatriotOracle:   "Patriot, Young Avenger",
		mscM7NakiaOracle:     "Nakia, Wakandan Operative",
		mscM7OkoyeOracle:     "Okoye, Mighty and Adored",
		mscM7PortalOracle:    "Negative Zone Portal",
		mscM7SpiderRexOracle: "Spider-Rex, Daring Dino",
		mscM7GwenOracle:      "Spider-Gwen, Free Spirit",
		mscM7WebSpinnerOrcl:  "Spider-Man, Web-Spinner",
		mscM7MobLookoutOrcl:  "Mob Lookout",
		mscM7RamaTutOracle:   "Pharaoh Rama-Tut",
		mscM7StatureOracle:   "Stature, Young Avenger",
		mscM7MjolnirOracle:   "Mjölnir's Might",
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

// mscM7Hand puts a card with the given types in a hand.
func mscM7Hand(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID,
	})
	return id
}

// --- Patriot, Young Avenger ------------------------------------------------------------

func TestMscM7PatriotAnthemOnlyWhileEquipped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	patriot := b12Push(g, me.ID, "Patriot, Young Avenger", "Legendary Creature — Human Hero", mscM7PatriotOracle, 3, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	sword := seedEquipment(g, me.ID, "Bonesplitter", bonesplitterOracle)

	if got := effectivePower(t, g, bear); got != 2 {
		t.Fatalf("unequipped Patriot gave the bear power %d, want 2", got)
	}
	if !hasEffectiveKeyword(t, g, patriot, "prowess") {
		t.Error("Patriot lacks prowess")
	}

	equipTo(t, g, me.ID, sword, patriot)
	if got := effectivePower(t, g, patriot); got != 5 {
		t.Errorf("equipped Patriot power %d, want 5 (3 + Bonesplitter's 2, no bonus for himself)", got)
	}
	if got := effectivePower(t, g, bear); got != 3 {
		t.Errorf("my other creature power %d, want 3", got)
	}
	if got := effectivePower(t, g, theirs); got != 2 {
		t.Errorf("an opponent's creature power %d, want 2", got)
	}

	// Moving the Equipment off Patriot turns the static off in the same pass.
	equipTo(t, g, me.ID, sword, bear)
	if got := effectivePower(t, g, patriot); got != 3 {
		t.Errorf("Patriot power %d after the sword moved, want 3", got)
	}
	if got := effectivePower(t, g, bear); got != 4 {
		t.Errorf("the equipped bear power %d, want 4 (2 + Bonesplitter's 2, no Patriot bonus)", got)
	}
}

// --- Nakia, Wakandan Operative -----------------------------------------------------------

// mscM7CastCreature casts a creature with real P/T from `p`'s hand during
// their main phase and lets it resolve.
func mscM7CastCreature(t *testing.T, g *game.Game, p *game.Player, name, oracle string, commander bool) uuid.UUID {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Bear", OracleID: oracle,
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID, IsCommander: commander,
	})
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	return id
}

func TestMscM7NakiaMakesYouMonarchWhenYourCommanderEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Nakia, Wakandan Operative", "Legendary Creature — Human Warrior Hero", mscM7NakiaOracle, false)

	mscM7CastCreature(t, g, me, "Plain Bear", "", false)
	if g.Monarch == me.ID {
		t.Fatal("a creature that is not my commander made me the monarch")
	}
	mscM7CastCreature(t, g, me, "My General", "", true)
	if g.Monarch != me.ID {
		t.Errorf("monarch is %v after my commander entered, want me", g.Monarch)
	}
}

func TestMscM7NakiaIgnoresAnotherPlayersCommander(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Nakia, Wakandan Operative", "Legendary Creature — Human Warrior Hero", mscM7NakiaOracle, false)
	advanceToMainOf(t, g, 1)

	mscM7CastCreature(t, g, opp, "Their General", "", true)
	if g.Monarch == me.ID {
		t.Error("an opponent's commander entering made me the monarch")
	}
}

func TestMscM7NakiaIsHerOwnCommanderTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	mscM7CastCreature(t, g, me, "Nakia, Wakandan Operative", mscM7NakiaOracle, true)
	if g.Monarch != me.ID {
		t.Errorf("Nakia entering as my commander did not make me the monarch (monarch %v)", g.Monarch)
	}
}

func TestMscM7NakiaPutsTwoCountersOnACreatureOrAVehicleAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	nakia := pushCatalogPermanent(g, me.ID, "Nakia, Wakandan Operative", "Legendary Creature — Human Warrior Hero", mscM7NakiaOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	van := pushPermanent(g, me.ID, game.Card{Name: "Cargo Van", TypeLine: "Artifact — Vehicle", Power: 3, Toughness: 3})
	trinket := pushPermanent(g, me.ID, game.Card{Name: "Trinket", TypeLine: "Artifact"})

	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, nakia, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: trinket}},
	}); err == nil {
		t.Fatal("two counters were put on an artifact that is neither a creature nor a Vehicle")
	}
	if err := g.ActivateCatalogAbility(me.ID, nakia, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate on a creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, bear); got != 2 {
		t.Errorf("the bear has %d counters, want 2", got)
	}

	battlefieldCardPtr(t, g, nakia).Tapped = false
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, nakia, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: van}},
	}); err != nil {
		t.Fatalf("activate on a Vehicle: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, van); got != 2 {
		t.Errorf("the Vehicle has %d counters, want 2", got)
	}
}

func TestMscM7NakiaRefusesBelowSorcerySpeedAndWithoutMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	nakia := pushCatalogPermanent(g, me.ID, "Nakia, Wakandan Operative", "Legendary Creature — Human Warrior Hero", mscM7NakiaOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	target := []game.TargetRef{{Kind: game.TargetCard, ID: bear}}

	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, nakia, 0, game.ActivateAbilityParams{Strict: true, Targets: target}); err == nil {
		t.Fatal("the {2} ability was accepted with one mana")
	}

	advanceTo(t, g, game.StepBeginCombat)
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, nakia, 0, game.ActivateAbilityParams{Strict: true, Targets: target}); err == nil {
		t.Fatal("the ability was activated in combat, not at sorcery speed")
	}
	if got := plusOneCounters(g, bear); got != 0 {
		t.Errorf("the bear has %d counters after refused activations, want 0", got)
	}
}

// --- Okoye, Mighty and Adored ----------------------------------------------------------------

// mscM7OkoyeTable casts Okoye (the enters trigger makes me the monarch),
// walks to the beginning of combat and puts the counter on `target`.
func mscM7OkoyeTable(t *testing.T) (g *game.Game, me *game.Player, bear, other uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[0]
	advanceToMain(t, g)
	bear = b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other = b12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "Okoye, Mighty and Adored", "Legendary Creature — Human Warrior Hero", mscM7OkoyeOracle, 3, 3)
	if g.Monarch != me.ID {
		t.Fatalf("Okoye entering did not make me the monarch (monarch %v)", g.Monarch)
	}
	advanceTo(t, g, game.StepBeginCombat)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: bear}, false)
	if got := plusOneCounters(g, bear); got != 1 {
		t.Fatalf("the targeted bear has %d counters, want 1", got)
	}
	if got := plusOneCounters(g, other); got != 0 {
		t.Fatalf("the other bear has %d counters, want 0", got)
	}
	return g, me, bear, other
}

func mscM7MakeMonarch(g *game.Game, p uuid.UUID) {
	g.WithWriteLock(func() { g.Monarch = p })
}

func TestMscM7OkoyeGrantsDoubleStrikeAndTrampleWhenTheCreatureAttacksTheMonarch(t *testing.T) {
	g, me, bear, other := mscM7OkoyeTable(t)
	opp := g.Seats[1]
	mscM7MakeMonarch(g, opp.ID)

	declareAttack(t, g, opp.ID, bear, other)
	passPriorityAroundTable(t, g)
	for _, kw := range []string{"double strike", "trample"} {
		if !hasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the counter's creature attacked the monarch and lacks %s", kw)
		}
		if hasEffectiveKeyword(t, g, other, kw) {
			t.Errorf("a creature Okoye did not target gained %s", kw)
		}
	}
	_ = me
}

func TestMscM7OkoyeGivesNothingWhenTheDefenderIsNotTheMonarch(t *testing.T) {
	g, _, bear, _ := mscM7OkoyeTable(t)
	opp, third := g.Seats[1], g.Seats[2]
	mscM7MakeMonarch(g, opp.ID)

	declareAttack(t, g, third.ID, bear)
	passPriorityAroundTable(t, g)
	for _, kw := range []string{"double strike", "trample"} {
		if hasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the creature attacked a player who is not the monarch and gained %s", kw)
		}
	}
}

func TestMscM7OkoyeGivesNothingWhenIAmTheMonarchAndAttackSomeoneElse(t *testing.T) {
	g, _, bear, _ := mscM7OkoyeTable(t)
	opp := g.Seats[1]

	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if hasEffectiveKeyword(t, g, bear, "double strike") {
		t.Error("the creature gained double strike attacking a player who is not the monarch (I hold the crown)")
	}
}

// --- Negative Zone Portal ----------------------------------------------------------------------

func mscM7PortalTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	seedLibrary(me, "A", "B", "C", "D", "E", "F")
	portal := pushCatalogPermanentWithID(g, breechesSourceID, me.ID, "Negative Zone Portal", "Artifact", mscM7PortalOracle, false)
	return g, me, opp, portal
}

// mscM7PortalExile activates the Portal on `card` (untapping it first).
func mscM7PortalExile(t *testing.T, g *game.Game, me *game.Player, portal, card uuid.UUID) error {
	t.Helper()
	battlefieldCardPtr(t, g, portal).Tapped = false
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, portal, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: card}},
	}); err != nil {
		return err
	}
	passPriorityAroundTable(t, g)
	return nil
}

func TestMscM7PortalExilesAnOpposingGraveyardCardAndDrawsForACreature(t *testing.T) {
	g, me, opp, portal := mscM7PortalTable(t)
	creature := pushGraveyardPermanent(opp, "Dead Bear", "Creature — Bear", "")
	instant := pushGraveyardPermanent(opp, "Dead Bolt", "Instant", "")
	mine := pushGraveyardPermanent(me, "My Dead Bear", "Creature — Bear", "")

	if err := mscM7PortalExile(t, g, me, portal, mine); err == nil {
		t.Fatal("the Portal accepted a card from my own graveyard")
	}
	if !me.Graveyard.Contains(mine) {
		t.Fatal("my own graveyard card left the graveyard")
	}

	hand := me.Hand.Size()
	if err := mscM7PortalExile(t, g, me, portal, instant); err != nil {
		t.Fatalf("exile an instant: %v", err)
	}
	if opp.Graveyard.Contains(instant) || g.FindCardZoneForEffect(instant).Kind != game.ZoneExile {
		t.Error("the instant card was not exiled")
	}
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("drew %d cards for a noncreature card, want 0", got)
	}

	if err := mscM7PortalExile(t, g, me, portal, creature); err != nil {
		t.Fatalf("exile a creature: %v", err)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards for a creature card, want 1", got)
	}
	linked := 0
	g.ReadSnapshot(func() {
		c := battlefieldCardPtr(t, g, portal)
		linked = len(g.ExiledWithForEffect(game.PermissionCardRef{ID: c.InstanceID, Epoch: c.ObjectEpoch}))
	})
	if linked != 2 {
		t.Errorf("%d cards are exiled with the Portal, want 2", linked)
	}
}

func TestMscM7PortalNeedsTwoManaAndATap(t *testing.T) {
	g, me, opp, portal := mscM7PortalTable(t)
	card := pushGraveyardPermanent(opp, "Dead Bear", "Creature — Bear", "")
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, portal, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: card}},
	}); err == nil {
		t.Fatal("the {2} ability was accepted with one mana")
	}
	if !opp.Graveyard.Contains(card) {
		t.Error("the card left the graveyard on a refused activation")
	}
}

// mscM7PortalUpkeep exiles `creatures` creature cards and one instant with
// the Portal, burns `burn` flips off the Portal's stream, then walks to my
// next upkeep. It returns the coin prompt (nil when none was asked).
func mscM7PortalUpkeep(t *testing.T, creatures, burn int) (g *game.Game, me, opp *game.Player, portal uuid.UUID, exiled []uuid.UUID, call *game.PendingChoice) {
	t.Helper()
	g, me, opp, portal = mscM7PortalTable(t)
	g.SetRNGKeyForTest(breechesRNGKey())
	for i := 0; i < creatures; i++ {
		exiled = append(exiled, pushGraveyardPermanent(opp, "Dead Bear", "Creature — Bear", ""))
	}
	exiled = append(exiled, pushGraveyardPermanent(opp, "Dead Bolt", "Instant", ""))
	for _, id := range exiled {
		if err := mscM7PortalExile(t, g, me, portal, id); err != nil {
			t.Fatalf("exile with the Portal: %v", err)
		}
	}
	advanceToUpkeepOf(t, g, 0)
	g.WithWriteLock(func() {
		for i := 0; i < burn; i++ {
			_, _ = g.FlipCoinsForEffect(game.RandomDraw{Player: me.ID, Source: portal}, 1)
		}
	})
	passPriorityAroundTable(t, g)
	call = latestChoiceOfKind(g, game.PendingChoiceCoinCall)
	return g, me, opp, portal, exiled, call
}

func TestMscM7PortalDoesNotFlipBelowFourCreatureCards(t *testing.T) {
	g, _, _, portal, _, call := mscM7PortalUpkeep(t, 3, 0)
	if call != nil {
		t.Fatal("the Portal flipped a coin with three creature cards (and an instant) exiled with it")
	}
	if !g.Battlefield.Contains(portal) {
		t.Error("the Portal is gone")
	}
}

func TestMscM7PortalFlipLoseSacrificesAndReturnsOneExiledCardWinKeepsIt(t *testing.T) {
	sawWin, sawLoss := false, false
	for burn := 0; burn < 16 && !(sawWin && sawLoss); burn++ {
		g, me, opp, portal, exiled, call := mscM7PortalUpkeep(t, 4, burn)
		if call == nil {
			t.Fatalf("burn %d: no coin prompt with four creature cards exiled", burn)
		}
		if err := g.ResolveCoinCall(call.ID, me.ID, "heads"); err != nil {
			t.Fatalf("ResolveCoinCall: %v", err)
		}
		flips := randomEvents(g, game.EventFlipCoin)
		won := flips[len(flips)-1].Won
		passPriorityAroundTable(t, g)
		inExile := 0
		for _, id := range exiled {
			if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneExile {
				inExile++
			}
		}
		if won {
			sawWin = true
			if !g.Battlefield.Contains(portal) {
				t.Error("a won flip sacrificed the Portal")
			}
			if inExile != len(exiled) {
				t.Errorf("a won flip returned a card: %d of %d still exiled", inExile, len(exiled))
			}
			continue
		}
		sawLoss = true
		if g.Battlefield.Contains(portal) {
			t.Error("a lost flip left the Portal on the battlefield")
		}
		if inExile != len(exiled)-1 {
			t.Errorf("a lost flip returned %d cards, want exactly 1", len(exiled)-inExile)
		}
		back := 0
		for _, id := range exiled {
			if opp.Hand.Contains(id) {
				back++
			}
			if me.Hand.Contains(id) {
				t.Error("the card went to the Portal controller's hand, not its owner's")
			}
		}
		if back != 1 {
			t.Errorf("%d exiled cards are in their owner's hand, want 1", back)
		}
	}
	if !sawWin || !sawLoss {
		t.Fatalf("both faces must be exercised: saw a win=%v, saw a loss=%v", sawWin, sawLoss)
	}
}

// --- Spider-Rex and Pharaoh Rama-Tut: ward {2} --------------------------------------------

func mscM7WardDeclined(t *testing.T, name, typeLine, oracle string) {
	t.Helper()
	g, _, opp, warded := wardTable(t, name, typeLine, oracle)
	blade := castAtWardedCreature(t, g, opp, warded)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatalf("%s: no ward payment prompt for the spell's controller", name)
	}
	answerPayUnless(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(warded) {
		t.Errorf("%s died to a spell whose controller declined the ward {2}", name)
	}
	if !opp.Graveyard.Contains(blade) {
		t.Errorf("%s: the countered spell is not in its owner's graveyard", name)
	}
}

func TestMscM7SpiderRexHasReachTrampleAndWard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rex := b12Push(g, me.ID, "Spider-Rex, Daring Dino", "Legendary Creature — Spider Dinosaur Hero", mscM7SpiderRexOracle, 6, 6)
	for _, kw := range []string{"reach", "trample"} {
		if !hasEffectiveKeyword(t, g, rex, kw) {
			t.Errorf("Spider-Rex lacks %s", kw)
		}
	}
	mscM7WardDeclined(t, "Spider-Rex, Daring Dino", "Legendary Creature — Spider Dinosaur Hero", mscM7SpiderRexOracle)
}

func TestMscM7RamaTutHasWard(t *testing.T) {
	mscM7WardDeclined(t, "Pharaoh Rama-Tut", "Legendary Creature — Human Noble Villain", mscM7RamaTutOracle)
}

// --- Pharaoh Rama-Tut: connive on a noncreature cast -------------------------------------------

func TestMscM7RamaTutConnivesOnYourNoncreatureSpellOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	seedLibrary(me, "A", "B", "C", "D")
	tut := b12Push(g, me.ID, "Pharaoh Rama-Tut", "Legendary Creature — Human Noble Villain", mscM7RamaTutOracle, 4, 4)

	castFree(t, g, me, "Creature Spell", "Creature — Bear")
	if discardChoiceFor(g, me.ID) != nil {
		t.Fatal("a discard prompt is open after a creature spell")
	}

	castFree(t, g, me, "Instant Spell", "Instant")
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("a noncreature spell did not make Rama-Tut connive")
	}
	drawn := me.Hand.Cards[me.Hand.Size()-1].InstanceID
	answerDiscard(t, g, me.ID, drawn)
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, tut); got != 1 {
		t.Errorf("Rama-Tut has %d counters after discarding a nonland card, want 1", got)
	}
}

// --- Spider-Gwen, Free Spirit --------------------------------------------------------------------

func TestMscM7GwenMayDiscardToDrawWhenTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	seedLibrary(me, "A", "B", "C")
	gwen := pushCatalogPermanent(g, me.ID, "Spider-Gwen, Free Spirit", "Legendary Creature — Spider Human Hero", mscM7GwenOracle, false)
	pitch := mscM7Hand(me, "Pitched Card", "Sorcery")

	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(gwen) })
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID, pitch)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d after discard-and-draw, want %d", got, hand)
	}
	if !me.Graveyard.Contains(pitch) {
		t.Error("the discarded card is not in the graveyard")
	}
}

func TestMscM7GwenDrawsNothingWhenYouDeclineTheDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	seedLibrary(me, "A", "B", "C")
	gwen := pushCatalogPermanent(g, me.ID, "Spider-Gwen, Free Spirit", "Legendary Creature — Spider Human Hero", mscM7GwenOracle, false)
	mscM7Hand(me, "Kept Card", "Sorcery")

	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(gwen) })
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d after declining, want %d (no discard, no draw)", got, hand)
	}
	if me.Graveyard.Size() != 0 {
		t.Error("a card was discarded though the player declined")
	}
}

func TestMscM7GwenTriggersWhenSheAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	seedLibrary(me, "A", "B", "C")
	gwen := b12Creature(g, me.ID, "Spider-Gwen, Free Spirit", "Legendary Creature — Spider Human Hero", 2, 3)
	battlefieldCardPtr(t, g, gwen).OracleID = mscM7GwenOracle
	mscM7Hand(me, "Pitched Card", "Sorcery")

	declareAttack(t, g, opp.ID, gwen)
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, me.ID) == nil {
		t.Error("attacking taps Spider-Gwen but no discard prompt appeared")
	}
}

// --- Spider-Man, Web-Spinner --------------------------------------------------------------------------

func TestMscM7WebSpinnerStopsOneCreatureFromBlocking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	spider := b12Push(g, me.ID, "Spider-Man, Web-Spinner", "Legendary Creature — Spider Human Hero", mscM7WebSpinnerOrcl, 2, 3)
	battlefieldCardPtr(t, g, spider).SummonedThisTurn = false
	webbed := b12Creature(g, opp.ID, "Webbed Guard", "Creature — Soldier", 2, 2)
	free := b12Creature(g, opp.ID, "Free Guard", "Creature — Soldier", 2, 2)
	for _, kw := range []string{"double strike", "haste"} {
		if !hasEffectiveKeyword(t, g, spider, kw) {
			t.Errorf("Spider-Man lacks %s", kw)
		}
	}

	declareAttack(t, g, opp.ID, spider)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: webbed}, false)
	assertRestrictions(t, g, webbed, game.CantBlock)
	assertRestrictions(t, g, free, 0)
}

// --- Mob Lookout -------------------------------------------------------------------------------------

func TestMscM7MobLookoutConnivesTheChosenCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	seedLibrary(me, "A", "B", "C")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	castHoldCreature(t, g, "Mob Lookout", "Creature — Human Rogue Villain", mscM7MobLookoutOrcl, 0, 3)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("the connive did not ask for a discard")
	}
	drawn := me.Hand.Cards[me.Hand.Size()-1].InstanceID
	answerDiscard(t, g, me.ID, drawn)
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, bear); got != 1 {
		t.Errorf("the bear has %d counters after a nonland discard, want 1", got)
	}
}

func TestMscM7MobLookoutLandDiscardAddsNoCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	seedLibrary(me, "A", "B", "C")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	land := mscM7Hand(me, "Spare Forest", "Basic Land — Forest")

	castHoldCreature(t, g, "Mob Lookout", "Creature — Human Rogue Villain", mscM7MobLookoutOrcl, 0, 3)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID, land)
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, bear); got != 0 {
		t.Errorf("the bear has %d counters after discarding a land, want 0", got)
	}
}

// --- Stature, Young Avenger --------------------------------------------------------------------------------

func TestMscM7StatureBecomesFourByFourOnYourNoncreatureCast(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	stature := b12Push(g, me.ID, "Stature, Young Avenger", "Legendary Creature — Human Hero", mscM7StatureOracle, 2, 2)

	castFree(t, g, me, "Creature Spell", "Creature — Bear")
	if got := effectivePower(t, g, stature); got != 2 {
		t.Fatalf("a creature spell changed Stature to power %d", got)
	}
	castFree(t, g, opp, "Their Instant", "Instant")
	if got := effectivePower(t, g, stature); got != 2 {
		t.Fatalf("an opponent's spell changed Stature to power %d", got)
	}

	castFree(t, g, me, "Instant Spell", "Instant")
	if p, tough := effectivePower(t, g, stature), effectiveToughness(t, g, stature); p != 4 || tough != 4 {
		t.Errorf("Stature is %d/%d after my noncreature spell, want 4/4", p, tough)
	}
	// Base P/T are set in layer 7b, so a counter still lands on top.
	g.WithWriteLock(func() {
		_ = AddCounter{Target: stature, Kind: game.CounterPlusOne, N: 1}.Apply(
			NewContext(g, &game.StackItem{ID: uuid.New(), Controller: me.ID, SourceCardID: stature}))
	})
	if p := hobA4Power(t, g, stature); p != 5 {
		t.Errorf("Stature with a counter has power %d, want 5", p)
	}

	advanceToUpkeepOf(t, g, 1)
	if p := hobA4Power(t, g, stature); p != 3 {
		t.Errorf("after end of turn Stature has power %d, want 3 (2 + the counter)", p)
	}
}

// --- Mjölnir's Might -------------------------------------------------------------------------------------------

func TestMscM7MjolnirDealsFourToAPlayerAndExilesTheTopCardPlayable(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	top := seedLibrary(me, "Top Card", "Second Card")[0]
	life := opp.Life
	hand := me.Hand.Size()

	castCatalogSpell(t, g, "Mjölnir's Might", "Sorcery", mscM7MjolnirOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)

	if opp.Life != life-4 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-4)
	}
	if z := g.FindCardZoneForEffect(top); z == nil || z.Kind != game.ZoneExile {
		t.Fatalf("the top card is not exiled: %+v", z)
	}
	if !b12ExiledPlayableBy(t, g, top, me.ID) {
		t.Error("I cannot play the exiled card")
	}
	if b12ExiledPlayableBy(t, g, top, opp.ID) {
		t.Error("an opponent can play my exiled card")
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand changed by %d, want 0", me.Hand.Size()-hand)
	}
}

func TestMscM7MjolnirRefusesACreatureTarget(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	advanceToMain(t, g)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Mjölnir's Might", "Sorcery", mscM7MjolnirOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err == nil {
		t.Error("Mjölnir's Might accepted a creature as its target player")
	}
}
