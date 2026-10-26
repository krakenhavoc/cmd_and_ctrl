package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sweep_04_restore_test.go — #4253's closure sweep, batch 04 (mairsil_the_pretender
// through paleontologists_pick_axe): the prompts that carried a Go closure are
// keyed now, so a table waiting on any of them is a restore point. Each test
// opens the prompt, takes a STRICT restore through JSON (restoreRoundTrip with
// strict, which fails on a table that is not a restore point) and answers on the
// restored table.

// sweep04OptionKey returns the open option pick of `chooser`, failing the test
// when it names a key other than `want`.
func sweep04OptionKey(t *testing.T, g *game.Game, chooser uuid.UUID, want string) *game.PendingChoice {
	t.Helper()
	c := latestOptionPickFor(g, chooser)
	if c == nil {
		t.Fatalf("no option pick for %s: %+v", chooser, g.PendingChoices)
	}
	if c.PickThen != want {
		t.Fatalf("the option pick names key %q, want %q", c.PickThen, want)
	}
	return c
}

// sweep04DiscardKey returns the open discard prompt of `player`, failing the
// test when it names a key other than `want`.
func sweep04DiscardKey(t *testing.T, g *game.Game, player uuid.UUID, want string) *game.PendingChoice {
	t.Helper()
	c := discardChoiceFor(g, player)
	if c == nil {
		t.Fatalf("no discard prompt for %s: %+v", player, g.PendingChoices)
	}
	if c.PickThen != want {
		t.Fatalf("the discard prompt names key %q, want %q", c.PickThen, want)
	}
	return c
}

// "You may return a card from your graveyard to your hand" is one keyed pick
// shared by every card that asks it (takeOneIntoYourHand).
func TestSweep04TakeOneIntoHandIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	source := pushVanillaCreature(g, me.ID, "Asker", 1, 1)
	back := b38Graveyard(g, me.ID, "Old Bear", "Creature — Bear", "{1}{G}", 2, 2)
	g.WithWriteLock(func() {
		item := &game.StackItem{Kind: game.StackItemTriggered, Controller: me.ID, Owner: me.ID, SourceCardID: source}
		if err := mayTakeOneIntoYourHand(NewContext(g, item), []uuid.UUID{back}, "Return a card?"); err != nil {
			t.Fatal(err)
		}
	})
	const key = "choose-cards/take-one-from-graveyard-into-hand"
	sweep00PickKey(t, g, me.ID, key)

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := sweep00PickKey(t, g, me.ID, key)
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{back}); err != nil {
		t.Fatalf("ResolveChooseCards on the restored table: %v", err)
	}
	if !me.Hand.Contains(back) || me.Graveyard.Contains(back) {
		t.Error("the card did not return to hand on the restored table")
	}
}

// Mairsil: the cage pick carries the trigger as data, and the card is caged
// with its counter on the restored table.
func TestSweep04MairsilCagePickIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	elves := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: elves, Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid",
		OracleID: blLlanowarOracle, Owner: me.ID, Controller: me.ID})
	castAndResolveCreature(t, g, "Mairsil, the Pretender", "Legendary Creature — Human Wizard", elMairsilOracle)
	passPriorityAroundTable(t, g)
	const key = "choose-cards/mairsil-cage"
	sweep00PickKey(t, g, me.ID, key)

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := sweep00PickKey(t, g, me.ID, key)
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{elves}); err != nil {
		t.Fatalf("ResolveChooseCards on the restored table: %v", err)
	}
	if n := elExiledCard(t, g, elves).Counters[mairsilCageCounter]; n != 1 {
		t.Errorf("the caged Elves have %d cage counters on the restored table, want 1", n)
	}
}

// Mount Doom: the creatures to keep, asked as the ability resolves, are a
// keyed pick whose answer destroys every other creature on the restored table.
func TestSweep04MountDoomKeepPickIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	doom := seedPermanentWithOracle(g, me.ID, "Mount Doom", "Legendary Land", b10MountDoomOracle)
	toMain(t, g)
	me.ManaPool.EmptyPool()
	b10AddMana(me, "C", "C", "C", "C", "C", "B", "R")
	ring := seedPermanentFor(g, me.ID, "The One Ring", "Legendary Artifact")
	keepMine := seedCreature(g, "Keep Mine", me.ID)
	loseMine := seedCreature(g, "Lose Mine", me.ID)
	loseTheirs := seedCreature(g, "Lose Theirs", opp.ID)
	if err := g.ActivateCatalogAbility(me.ID, doom, 1, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{ring},
	}); err != nil {
		t.Fatalf("wipe: %v", err)
	}
	passPriorityAroundTable(t, g)
	const key = "choose-cards/mount-doom-keep"
	sweep00PickKey(t, g, me.ID, key)

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := sweep00PickKey(t, g, me.ID, key)
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{keepMine}); err != nil {
		t.Fatalf("ResolveChooseCards on the restored table: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(keepMine) {
		t.Error("the kept creature did not survive on the restored table")
	}
	if g.Battlefield.Contains(loseMine) || g.Battlefield.Contains(loseTheirs) {
		t.Error("the creatures that were not kept were not destroyed on the restored table")
	}
}

// Myr Battlesphere: the Myr to tap carry the attack trigger as data, so the
// restored answer pumps the Battlesphere and hits the player it attacked.
func TestSweep04MyrBattlespherePickIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sphere := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Myr Battlesphere", TypeLine: "Artifact Creature — Myr Construct",
		OracleID: myrBattlesphereOracle, Power: 4, Toughness: 7, Owner: me.ID, Controller: me.ID,
	})
	var myr []uuid.UUID
	for i := 0; i < 3; i++ {
		myr = append(myr, pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Myr", TypeLine: "Token Artifact Creature — Myr",
			Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
		}))
	}
	before := opp.Life
	declareAttack(t, g, opp.ID, sphere)
	passPriorityAroundTable(t, g)
	const key = "choose-cards/myr-battlesphere-tap"
	sweep00PickKey(t, g, me.ID, key)

	g = restoreRoundTrip(t, g, true)
	me, opp = g.Seats[0], g.Seats[1]
	c := sweep00PickKey(t, g, me.ID, key)
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{myr[0], myr[1]}); err != nil {
		t.Fatalf("ResolveChooseCards on the restored table: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("the defending player is at %d on the restored table, want %d (X = 2 damage)", opp.Life, before-2)
	}
	if p := effectivePower(t, g, sphere); p != 6 {
		t.Errorf("the Battlesphere has power %d on the restored table, want 6", p)
	}
}

// Nesting Grounds: which counter kind to move is a keyed option pick carrying
// the two permanents, and the counter lands on the second on the restored table.
func TestSweep04NestingGroundsKindIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	grounds := seedPermanentWithOracle(g, me.ID, "Nesting Grounds", "Land", nestingGroundsOracle)
	from := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Source", TypeLine: "Creature — Test",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterPlusOne: 1, "stun": 1},
	})
	to := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dest", TypeLine: "Creature — Test",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, grounds, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: from}, {Kind: game.TargetCard, ID: to}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	const key = "option-pick/nesting-grounds-counter-kind"
	sweep04OptionKey(t, g, me.ID, key)

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	sweep04OptionKey(t, g, me.ID, key)
	// Kinds are offered in name order: "+1/+1" before "stun".
	answerOptionPick(t, g, me.ID, 1)
	src, _ := battlefieldCard(g, from)
	dst, _ := battlefieldCard(g, to)
	if src.Counters["stun"] != 0 || dst.Counters["stun"] != 1 || src.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("counters source=%v dest=%v on the restored table, want the stun counter moved", src.Counters, dst.Counters)
	}
}

// Mana Vortex: the cast trigger's choice carries the spell and the player, and
// both branches run on the restored table.
func TestSweep04ManaVortexChoiceIsARestorePoint(t *testing.T) {
	const key = "option-pick/mana-vortex-sacrifice-or-counter"
	t.Run("sacrifice a land", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		myLand := apaPush(g, me.ID, me.ID, stCard("Island", "", "Basic Land — Island", 0, 0))
		// A land elsewhere keeps the Vortex from leaving for want of any.
		apaPush(g, g.Seats[1].ID, g.Seats[1].ID, stCard("Forest", "", "Basic Land — Forest", 0, 0))
		vortex := castCatalogSpell(t, g, "Mana Vortex", "Enchantment", stManaVortex, nil)
		passPriorityAroundTable(t, g)
		sweep04OptionKey(t, g, me.ID, key)

		g = restoreRoundTrip(t, g, true)
		me = g.Seats[0]
		sweep04OptionKey(t, g, me.ID, key)
		answerOptionPick(t, g, me.ID, 0)
		answerChooseCards(t, g, me.ID, myLand)
		passPriorityAroundTable(t, g)
		if onBattlefield(g, myLand) || !onBattlefield(g, vortex) {
			t.Errorf("on the restored table: land %v, vortex %v; want the land gone and the Vortex in play",
				onBattlefield(g, myLand), onBattlefield(g, vortex))
		}
	})
	t.Run("let it be countered", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		myLand := apaPush(g, me.ID, me.ID, stCard("Island", "", "Basic Land — Island", 0, 0))
		vortex := castCatalogSpell(t, g, "Mana Vortex", "Enchantment", stManaVortex, nil)
		passPriorityAroundTable(t, g)
		sweep04OptionKey(t, g, me.ID, key)

		g = restoreRoundTrip(t, g, true)
		me = g.Seats[0]
		answerOptionPick(t, g, me.ID, 1)
		passPriorityAroundTable(t, g)
		if onBattlefield(g, vortex) || !me.Graveyard.Contains(vortex) {
			t.Error("the Vortex was not countered on the restored table")
		}
		if !onBattlefield(g, myLand) {
			t.Error("the land was sacrificed though the Vortex was let go")
		}
	})
}

// Orcish Medicine: the keyword choice carries its target, and the Army follows
// the grant on the restored table.
func TestSweep04OrcishMedicineKeywordIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushDiesCreatureForTest(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Orcish Medicine", "Instant", orcishMedicineOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	const key = "option-pick/orcish-medicine-keyword"
	sweep04OptionKey(t, g, me.ID, key)

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	sweep04OptionKey(t, g, me.ID, key)
	answerOptionPick(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)
	if c := findBattlefieldCardByID(g, bear); c == nil || !game.HasKeyword(c, "indestructible") {
		t.Error("the Bear did not gain indestructible on the restored table")
	}
	if army := armyOf(t, g, me.ID); countersOn(g, army.InstanceID, game.CounterPlusOne) != 1 {
		t.Error("the Army did not get its counter on the restored table")
	}
}

// Monumental Henge: the look is a registered take, so the pick holds no
// closure; the revealed card reaches the hand and the rest go under the library
// on the restored table.
func TestSweep04MonumentalHengeTakeIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	henge := seedPermanentWithOracle(g, me.ID, "Monumental Henge", "Land", samiMonumentalHengeOracle)
	ids := samiLibraryTopFirst(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Shock", TypeLine: "Instant"},
		game.Card{Name: "Wurm", TypeLine: "Artifact Creature — Wurm"},
		game.Card{Name: "Lone Hero", TypeLine: "Legendary Creature — Human"},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Sixth", TypeLine: "Instant"},
	)
	samiMainPhase(t, g)
	libBefore := me.Library.Size()
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{W}{W}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, henge, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	const key = "choose-cards/take-reveal-rest-bottom-random"
	sweep00PickKey(t, g, me.ID, key)

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	c := sweep00PickKey(t, g, me.ID, key)
	if !hasID(c.ChooseCards, ids[2]) || !hasID(c.ChooseCards, ids[3]) || hasID(c.ChooseCards, ids[0]) {
		t.Errorf("the restored pick offers %v, want exactly the two historic cards", c.ChooseCards)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{ids[3]}); err != nil {
		t.Fatalf("ResolveChooseCards on the restored table: %v", err)
	}
	if !me.Hand.Contains(ids[3]) {
		t.Error("the chosen legendary did not reach the hand on the restored table")
	}
	if me.Library.Size() != libBefore-1 {
		t.Fatalf("library = %d, want %d", me.Library.Size(), libBefore-1)
	}
	if top := me.Library.Cards[me.Library.Size()-1].InstanceID; top != ids[5] {
		t.Error("the sixth card must be on top; the other four went under it")
	}
}

// Narset of the Ancient Way's -2: the discard is a keyed discard whose rest
// raises the reflexive trigger, so the damage is dealt on the restored table.
func TestSweep04NarsetDiscardIsARestorePoint(t *testing.T) {
	g, me, opp := spendTable(t)
	n := guffbWalker(g, me.ID, "Narset of the Ancient Way", "Legendary Planeswalker — Narset", guffbNarsetOracle, 4)
	pitch := handCardOf(g, me, "Pitch", "Sorcery", "{1}{R}{R}", "")
	victim := pushVanillaCreature(g, opp.ID, "Victim", 4, 5)
	seedSearchLibrary(me, searchTestLand("Drawn", "Instant"))
	b16Activate(t, g, me.ID, n, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	const key = "choose-cards/narset-of-the-ancient-way-discarded"
	sweep04DiscardKey(t, g, me.ID, key)

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	sweep04DiscardKey(t, g, me.ID, key)
	answerDiscard(t, g, me.ID, pitch)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	var damage int
	g.WithWriteLock(func() { damage = findBattlefieldCardForTest(g, victim).DamageMarked })
	if damage != 3 {
		t.Errorf("damage %d on the restored table, want 3 (the discarded card's mana value)", damage)
	}
}

// Nezumi Shortfang: the discard carries the ability's item as data, so the
// flip is still checked against the right object on the restored table.
func TestSweep04NezumiShortfangDiscardIsARestorePoint(t *testing.T) {
	g, me, opp := flipMain(t)
	id := flipEntered(t, g, nezumiShortfangRow(), me)
	hand := flipFillHand(g, opp, 1)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	const key = "choose-cards/nezumi-shortfang-discarded"
	sweep04DiscardKey(t, g, opp.ID, key)

	g = restoreRoundTrip(t, g, true)
	opp = g.Seats[1]
	sweep04DiscardKey(t, g, opp.ID, key)
	answerDiscard(t, g, opp.ID, hand[0])
	passPriorityAroundTable(t, g)
	wantFlipped(t, g, id, "Stabwhisker the Odious", []string{"B"}, 2)
}

// Mog, Moogle Warrior: each player's discard is a keyed discard carrying the
// trigger as data and the players still to ask as the carried IDs. The table is
// restored while the first and the second player's discards are open, and the
// dance still ends with the Moogle and the counters.
func TestSweep04MogDiscardIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mog := pushCatalogPermanent(g, me.ID, "Mog, Moogle Warrior", "Legendary Creature — Moogle Warrior", deck3655MogOracle, false)
	myCreature := handCardForTest(me, "My Bear", "Creature — Bear", "")
	theirSpell := handCardForTest(opp, "Their Bolt", "Instant", "")
	myHand, theirHand := me.Hand.Size(), opp.Hand.Size()

	advanceToStepOf(t, g, 0, game.StepEnd)
	if triggerOnStack(g, mog) == nil {
		t.Fatal("no Dance trigger at my end step")
	}
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	const key = "choose-cards/mog-moogle-warrior-discarded"
	sweep04DiscardKey(t, g, me.ID, key)

	g = restoreRoundTrip(t, g, true)
	sweep04DiscardKey(t, g, me.ID, key)
	answerDiscard(t, g, me.ID, myCreature)

	g = restoreRoundTrip(t, g, true)
	answerMayChoice(t, g, opp.ID, true)
	g = restoreRoundTrip(t, g, true)
	me, opp = g.Seats[0], g.Seats[1]
	sweep04DiscardKey(t, g, opp.ID, key)
	answerDiscard(t, g, opp.ID, theirSpell)
	for _, p := range g.Seats[2:] {
		answerMayChoice(t, g, p.ID, false)
	}
	passPriorityAroundTable(t, g)

	if me.Hand.Size() != myHand || opp.Hand.Size() != theirHand {
		t.Errorf("hands %d and %d on the restored table, want %d and %d (one discarded, one drawn)",
			me.Hand.Size(), opp.Hand.Size(), myHand, theirHand)
	}
	var moogles, counters int
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.HasSubtype("Moogle") {
			moogles++
			counters += c.Counters[game.CounterPlusOne]
		}
	}
	if moogles != 2 || counters != 2 {
		t.Errorf("Moogles = %d and counters = %d on the restored table, want 2 and 2", moogles, counters)
	}
}

// Nautiloid Ship: the creature card exiled with it is chosen by a keyed pick
// whose only data is the chooser, and it arrives under their control on the
// restored table.
func TestSweep04NautiloidShipPickIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := pushGraveyardCardForTest(opp, "Their Dragon")
	ship := s58Cast(t, g, game.Card{Name: "Nautiloid Ship", TypeLine: s58NautiloidShipTypeLine, OracleID: s58NautiloidShipOracle,
		ManaCost: "{4}", Power: 5, Toughness: 5}, nil)
	for i := 0; i < 8 && latestPickTarget(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == ship {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	}
	crewer := pushCrewerForTest(g, me.ID, "Crewer", 3)
	if err := g.ActivateCatalogAbility(me.ID, ship, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{crewer}}); err != nil {
		t.Fatalf("crew: %v", err)
	}
	passPriorityAroundTable(t, g)
	attackWith(t, g, opp.ID, ship)
	passPriorityAroundTable(t, g)
	const key = "choose-cards/nautiloid-ship-return"
	sweep00PickKey(t, g, me.ID, key)

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := sweep00PickKey(t, g, me.ID, key)
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{dragon}); err != nil {
		t.Fatalf("ResolveChooseCards on the restored table: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := s58Named(g, me.ID, "Their Dragon"); len(got) != 1 {
		t.Errorf("Their Dragon under my control = %d on the restored table, want 1", len(got))
	}
}
