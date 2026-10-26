package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sweep_05_restore_test.go — #4253's closure sweep, batch 05 (path_to_exile
// through selective_obliteration): the option picks, card picks and discards
// that carried a Go closure are keyed now, so a table waiting on any of them
// is a restore point. Each test opens the prompt, takes a STRICT restore
// through JSON (restoreRoundTrip with strict, which fails on a table that is
// not a restore point) and answers on the restored table.

// sweep05OptionKey returns the open option pick of `chooser`, failing the
// test when it names a key other than `want`.
func sweep05OptionKey(t *testing.T, g *game.Game, chooser uuid.UUID, want string) *game.PendingChoice {
	t.Helper()
	c := latestOptionPickFor(g, chooser)
	if c == nil {
		t.Fatalf("no option pick for %s: %+v", chooser, g.PendingChoices)
	}
	if c.PickThen != want {
		t.Fatalf("the prompt names key %q, want %q", c.PickThen, want)
	}
	return c
}

// The repeated three-way punisher (Torment of Hailfire, Rottenmouth Viper):
// the question, the discard it can ask and the next opponent's question are
// each a restore point, and the opponents still to ask ride the prompts as
// data.
func TestSweep05TormentOfHailfireDiscardBranchIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	hand := opp.Hand.Size()

	castTormentOfHailfire(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)
	sweep05OptionKey(t, g, opp.ID, "option-pick/torment-of-hailfire-punisher")

	g = restoreRoundTrip(t, g, true)
	opp = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	sweep05OptionKey(t, g, opp.ID, "option-pick/torment-of-hailfire-punisher")
	answerOptionPick(t, g, opp.ID, 2) // discard a card
	sweep00PickKey(t, g, opp.ID, "choose-cards/torment-of-hailfire-punisher-discarded")

	g = restoreRoundTrip(t, g, true)
	opp = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	c := sweep00PickKey(t, g, opp.ID, "choose-cards/torment-of-hailfire-punisher-discarded")
	if err := g.ResolveChooseCards(c.ID, opp.ID, c.ChooseCards[:1]); err != nil {
		t.Fatalf("ResolveChooseCards on the restored table: %v", err)
	}
	if opp.Hand.Size() != hand-1 {
		t.Errorf("hand %d → %d on the restored table, want one discarded", hand, opp.Hand.Size())
	}
	// The next opponent is asked, with the same key.
	next := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	sweep05OptionKey(t, g, next.ID, "option-pick/torment-of-hailfire-punisher")
	drainRemainingTormentPrompts(t, g)
}

// The sacrifice branch is still a choose_cards prompt with a closure until
// SacrificeChoice is keyed (the primitives pass), so the table is a restore
// point again once the sacrifice is answered: the next opponent is asked
// with the same key.
func TestSweep05TormentOfHailfireSacrificeBranchThenNextOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	castTormentOfHailfire(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, opp.ID, 1) // sacrifice a nonland permanent
	answerChooseCards(t, g, opp.ID, bear)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the chosen permanent was not sacrificed")
	}

	g = restoreRoundTrip(t, g, true)
	next := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	sweep05OptionKey(t, g, next.ID, "option-pick/torment-of-hailfire-punisher")
	drainRemainingTormentPrompts(t, g)
}

// Rottenmouth Viper shares the punisher with its own keys: the life branch
// of the first opponent's question loses 4 life on the restored table.
func TestSweep05RottenmouthViperQuestionIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	life := opp.Life
	var queued error
	g.WithWriteLock(func() {
		item := &game.StackItem{Kind: game.StackItemTriggered, Controller: me.ID, Owner: me.ID}
		queued = rottenmouthViper.start(NewContext(g, item), 1)
	})
	if queued != nil {
		t.Fatal(queued)
	}
	sweep05OptionKey(t, g, opp.ID, "option-pick/rottenmouth-viper-punisher")

	g = restoreRoundTrip(t, g, true)
	opp = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	sweep05OptionKey(t, g, opp.ID, "option-pick/rottenmouth-viper-punisher")
	answerOptionPick(t, g, opp.ID, 0)
	if opp.Life != life-4 {
		t.Errorf("life %d → %d on the restored table, want -4", life, opp.Life)
	}
	drainRemainingTormentPrompts(t, g)
}

// The upkeep tax of The Gitrog Monster (and Territorial Dispute): the
// sacrifice-a-land-or-this question is keyed, and the land it asks for is
// the engine's own sacrifice prompt.
func TestSweep05GitrogUpkeepQuestionIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	gitrog := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Gitrog Monster", TypeLine: "Legendary Creature — Frog Horror",
		OracleID: gitrogMonsterOracle, Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	land := pushPermanentForTest(g, me.ID, "Forest", "", "Basic Land — Forest")
	advanceToUpkeepOf(t, g, seat)
	passPriorityAroundTable(t, g)
	sweep05OptionKey(t, g, me.ID, "option-pick/sacrifice-this-unless-you-sacrifice-a-land")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[seat]
	sweep05OptionKey(t, g, me.ID, "option-pick/sacrifice-this-unless-you-sacrifice-a-land")
	answerOptionPick(t, g, me.ID, 0) // sacrifice a land
	// The land is asked as a choose_cards prompt that holds a closure until
	// SacrificeChoice is keyed, so the table is not restored here.
	answerChooseCards(t, g, me.ID, land)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) || !g.Battlefield.Contains(gitrog) {
		t.Errorf("land on the battlefield %v, Gitrog %v; want the land gone and the Gitrog kept",
			g.Battlefield.Contains(land), g.Battlefield.Contains(gitrog))
	}
}

// The other branch: the permanent itself is sacrificed.
func TestSweep05GitrogUpkeepSacrificingItselfAfterARestore(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	gitrog := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Gitrog Monster", TypeLine: "Legendary Creature — Frog Horror",
		OracleID: gitrogMonsterOracle, Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	pushPermanentForTest(g, me.ID, "Forest", "", "Basic Land — Forest")
	advanceToUpkeepOf(t, g, seat)
	passPriorityAroundTable(t, g)

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[seat]
	answerOptionPick(t, g, me.ID, 1) // sacrifice The Gitrog Monster
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(gitrog) {
		t.Error("The Gitrog Monster was not sacrificed on the restored table")
	}
}

// Plague Boiler's put-or-remove choice: the Boiler rides the prompt as its
// source, so the restored answer adds the counter.
func TestSweep05PlagueBoilerChoiceIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	boiler := apaPush(g, me.ID, me.ID, stCard("Plague Boiler", "fef502af-6e79-4c55-a86a-b45adb3fc64a", "Artifact", 0, 0))
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(boiler, "plague", 1); err != nil {
			t.Fatal(err)
		}
	})
	apaMana(me, "C", "B", "G")
	if err := g.ActivateCatalogAbility(me.ID, boiler, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	sweep05OptionKey(t, g, me.ID, "option-pick/plague-boiler-put-or-remove")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	sweep05OptionKey(t, g, me.ID, "option-pick/plague-boiler-put-or-remove")
	answerOptionPick(t, g, me.ID, 0) // put a counter
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, boiler); !ok || c.Counters["plague"] != 2 {
		t.Errorf("plague counters on the restored table = %+v, want 2", c.Counters)
	}
}

// Psychic Puppetry: the permanent rides the prompt as its carried ID.
func TestSweep05PsychicPuppetryChoiceIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	castCatalogSpell(t, g, "Psychic Puppetry", "Instant — Arcane", psychicPuppetryOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	sweep05OptionKey(t, g, me.ID, "option-pick/psychic-puppetry-tap-or-untap")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	answerOptionPick(t, g, me.ID, 0) // tap it
	if !b16Tapped(t, g, bear) {
		t.Error("the permanent was not tapped on the restored table")
	}
}

// Merrow Reejerey and Gandalf the Grey share the tap-or-untap choice.
func TestSweep05TapOrUntapThatPermanentIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Permanent(g, opp.ID, "Their Island", "Basic Land — Island")
	g.WithWriteLock(func() {
		item := &game.StackItem{Kind: game.StackItemTriggered, Controller: me.ID, Owner: me.ID}
		if err := tapOrUntapPermanent(NewContext(g, item), "Merrow Reejerey", victim); err != nil {
			t.Fatal(err)
		}
	})
	sweep05OptionKey(t, g, me.ID, "option-pick/tap-or-untap-that-permanent")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	answerOptionPick(t, g, me.ID, 0)
	if !b16Tapped(t, g, victim) {
		t.Error("the permanent was not tapped on the restored table")
	}
}

// Eureka Moment (and the other cards that print the clause): the put a land
// pick is keyed, and "tapped" is the one carried number.
func TestSweep05PutALandFromHandIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := handCardForTest(me, "Forest", "Basic Land — Forest", "")
	castCatalogSpell(t, g, "Eureka Moment", "Instant", putFromHandEurekaOracle, nil)
	passPriorityAroundTable(t, g)
	sweep00PickKey(t, g, me.ID, "choose-cards/put-from-hand-onto-battlefield")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := sweep00PickKey(t, g, me.ID, "choose-cards/put-from-hand-onto-battlefield")
	if c.ChooseMin != 0 || c.ChooseMax != 1 || c.ChooseDestination != game.ChooseOntoBattlefield {
		t.Errorf("the restored pick is %d..%d to %q, want the optional one to the battlefield",
			c.ChooseMin, c.ChooseMax, c.ChooseDestination)
	}
	answerChooseCards(t, g, me.ID, land)
	if !g.Battlefield.Contains(land) || tappedOnBattlefield(t, g, land) {
		t.Error("the land did not arrive untapped on the restored table")
	}
}

// The tapped clause (Insidious Fungus, Arboreal Grazer): the number rides.
func TestSweep05PutALandFromHandTappedIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := handCardForTest(me, "Forest", "Basic Land — Forest", "")
	g.WithWriteLock(func() {
		item := &game.StackItem{Kind: game.StackItemTriggered, Controller: me.ID, Owner: me.ID}
		if err := MayPutALandFromHandTapped("Arboreal Grazer").Apply(NewContext(g, item)); err != nil {
			t.Fatal(err)
		}
	})
	sweep00PickKey(t, g, me.ID, "choose-cards/put-from-hand-onto-battlefield")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	answerChooseCards(t, g, me.ID, land)
	if !g.Battlefield.Contains(land) || !tappedOnBattlefield(t, g, land) {
		t.Error("the land did not arrive tapped on the restored table")
	}
}

// Rejoin the Fight: each opponent chooses in turn, and the opponents still
// to ask and the cards chosen so far ride each prompt as data, so the table
// is a restore point between any two choices.
func TestSweep05RejoinTheFightWalkIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := deck3655GraveyardCreature(me, "Alpha", 1)
	b := deck3655GraveyardCreature(me, "Beta", 1)
	c := deck3655GraveyardCreature(me, "Gamma", 1)

	castCatalogSpell(t, g, "Rejoin the Fight", "Sorcery", deck3655RejoinOracle, nil)
	passPriorityAroundTable(t, g)

	picks := []uuid.UUID{a, b, c}
	for i, id := range picks {
		g = restoreRoundTrip(t, g, true)
		chooser := g.Seats[i+1]
		prompt := sweep00PickKey(t, g, chooser.ID, "choose-cards/rejoin-the-fight-pick")
		for _, offered := range prompt.ChooseCards {
			for _, taken := range picks[:i] {
				if offered == taken {
					t.Errorf("opponent %d was offered a card already chosen after a restore", i+1)
				}
			}
		}
		answerChooseCards(t, g, chooser.ID, id)
		if i < 2 && g.Battlefield.Contains(id) {
			t.Fatal("nothing returns until the last opponent has chosen")
		}
	}
	passPriorityAroundTable(t, g)
	for _, id := range picks {
		if !g.Battlefield.Contains(id) {
			t.Errorf("chosen card %s did not return on the restored table", id)
		}
	}
	for _, card := range g.Battlefield.Cards {
		if (card.InstanceID == a || card.InstanceID == b || card.InstanceID == c) && card.Controller != g.Seats[0].ID {
			t.Errorf("%s returned under %s, not the caster", card.Name, card.Controller)
		}
	}
}

// Reckoner Shakedown: choosing no card puts two counters on a creature or
// Vehicle you pick, and that pick is a restore point.
func TestSweep05ReckonerShakedownCounterPickIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	pushCreatureToBattlefieldForTest(g, caster.ID, "My Bear")
	vehicle := pushCatalogPermanent(g, caster.ID, "Smuggler's Copter", "Artifact — Vehicle", "", false)
	revealHand(victim, rhBolt())

	castAt(t, g, "Reckoner Shakedown", "Sorcery", reckonerShakedownOracle, pbPlayer(victim.ID))
	pick := openVariant(t, g)
	if err := g.ResolvePendingChoice(pick.ID, caster.ID, nil); err != nil {
		t.Fatalf("choose nothing: %v", err)
	}
	sweep00PickKey(t, g, caster.ID, "choose-cards/reckoner-shakedown-counters")

	g = restoreRoundTrip(t, g, true)
	caster = g.Seats[0]
	sweep00PickKey(t, g, caster.ID, "choose-cards/reckoner-shakedown-counters")
	answerChooseCards(t, g, caster.ID, vehicle)
	if v := rvBattlefieldCard(g, vehicle); v == nil || v.Counters["+1/+1"] != 2 {
		t.Error("the Vehicle did not get two +1/+1 counters on the restored table")
	}
}

// Ripples of Potential: the permanents to phase out are chosen on a keyed
// pick whose source is the spell.
func TestSweep05RipplesOfPotentialPhaseOutPickIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := seedCreatureWithCounters(g, me.ID, map[string]int{game.CounterPlusOne: 1})
	b := seedCreatureWithCounters(g, me.ID, map[string]int{game.CounterPlusOne: 1})

	castRipples(t, g)
	answerProliferate(t, g, a, b)
	if c := ripplesChoosePrompt(g); c == nil {
		t.Fatal("no choose prompt after the proliferate")
	}

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	c := sweep00PickKey(t, g, me.ID, "choose-cards/ripples-of-potential-phase-out")
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{a}); err != nil {
		t.Fatalf("ResolveChooseCards on the restored table: %v", err)
	}
	if !phasedOutInCatalogGame(g, a) || phasedOutInCatalogGame(g, b) {
		t.Errorf("phased out: a %v, b %v on the restored table; want only a",
			phasedOutInCatalogGame(g, a), phasedOutInCatalogGame(g, b))
	}
}

// Seasoned Cryomancer: the loot's discard is a keyed discard-then, and the
// reflexive stun trigger counts the nonland cards that really left the hand.
func TestSweep05SeasonedCryomancerDiscardIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, foe := g.Seats[g.Turn.ActiveSeat], rfCrEFoe(g)
	me.Hand.Cards = nil
	for i := 0; i < 4; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spell", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	}
	land := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: land, Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	victim := rfCrAPush(g, foe.ID, "Victim", "Creature — Bear", "", 2, 2)
	rfCrACast(t, g, "Seasoned Cryomancer", "Creature — Human Wizard", oracleCryomancer, 2, 2)
	passPriorityAroundTable(t, g)
	spellInHand := uuid.Nil
	for _, c := range me.Hand.Cards {
		if c.Name == "Spell" {
			spellInHand = c.InstanceID
		}
	}
	if spellInHand == uuid.Nil {
		t.Fatal("no Spell was drawn")
	}
	sweep00PickKey(t, g, me.ID, "choose-cards/seasoned-cryomancer-discarded")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	sweep00PickKey(t, g, me.ID, "choose-cards/seasoned-cryomancer-discarded")
	answerDiscard(t, g, me.ID, land, spellInHand)
	passPriorityAroundTable(t, g)
	pickTriggerTarget(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	v, _ := battlefieldCard(g, victim)
	if !v.Tapped || v.Counters[game.CounterStun] != 1 {
		t.Fatalf("the chosen creature on the restored table: tapped %v, stun %d", v.Tapped, v.Counters[game.CounterStun])
	}
}

// Keys to the House and Ghostly Keybearer: the lock-or-unlock choice of a
// Room with both doors to choose from carries the Room as data, and the
// door and the action are the option's label.
func TestSweep05ChooseADoorIsARestorePoint(t *testing.T) {
	registerTestRoom(t, map[string]int{})
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	c := testRoomCard(me.ID)
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 0}); err != nil {
		t.Fatal(err)
	}
	settleOrdering(t, g, me.ID)
	item := &game.StackItem{Controller: me.ID, SourceCardID: c.InstanceID}
	g.WithWriteLock(func() {
		if err := (UnlockADoor{Room: c.InstanceID}).Apply(NewContext(g, item)); err != nil {
			t.Fatal(err)
		}
		if err := (LockOrUnlockADoor{Room: c.InstanceID}).Apply(NewContext(g, item)); err != nil {
			t.Fatal(err)
		}
	})
	sweep05OptionKey(t, g, me.ID, "option-pick/choose-a-door")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	sweep05OptionKey(t, g, me.ID, "option-pick/choose-a-door")
	answerOptionPick(t, g, me.ID, 0) // lock the first door
	if IsFullyUnlocked(g, c.InstanceID) {
		t.Error("the Room is still fully unlocked after the restored table locked a door")
	}
}

// Ghostly Dancers: unlock a locked door of a Room you control. The Room is
// the option's own card, the door its label.
func TestSweep05UnlockALockedDoorOfARoomIsARestorePoint(t *testing.T) {
	registerTestRoom(t, map[string]int{})
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	var rooms []uuid.UUID
	for i := 0; i < 2; i++ {
		c := testRoomCard(me.ID)
		me.Hand.PushTop(c)
		if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 0}); err != nil {
			t.Fatal(err)
		}
		settleOrdering(t, g, me.ID)
		rooms = append(rooms, c.InstanceID)
	}
	item := &game.StackItem{Controller: me.ID}
	g.WithWriteLock(func() {
		if err := (UnlockALockedDoorOfARoomYouControl{}).Apply(NewContext(g, item)); err != nil {
			t.Fatal(err)
		}
	})
	sweep05OptionKey(t, g, me.ID, "option-pick/unlock-a-locked-door-of-a-room-you-control")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	pick := sweep05OptionKey(t, g, me.ID, "option-pick/unlock-a-locked-door-of-a-room-you-control")
	index := -1
	for i, o := range pick.PickOptions {
		if len(o.Cards) == 1 && o.Cards[0] == rooms[1] {
			index = i
		}
	}
	if index < 0 {
		t.Fatalf("no option names the second Room: %+v", pick.PickOptions)
	}
	answerOptionPick(t, g, me.ID, index)
	if !IsFullyUnlocked(g, rooms[1]) || IsFullyUnlocked(g, rooms[0]) {
		t.Errorf("fully unlocked: first %v, second %v on the restored table; want only the second",
			IsFullyUnlocked(g, rooms[0]), IsFullyUnlocked(g, rooms[1]))
	}
}

// Realmbreaker, the Invasion Tree: the land stolen from the graveyard is a
// keyed pick that carries the asking item, so the restored answer returns it
// tapped under the activator's control, with its exile-instead rider.
func TestSweep05RealmbreakerLandPickIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tree := seedPermanentWithOracle(g, me.ID, "Realmbreaker, the Invasion Tree", "Legendary Artifact", rbRealmbreakerOracle)
	land := pushGraveyardCardTyped(opp, "Their Mountain", "Basic Land — Mountain")
	advanceToMain(t, g)

	dcMana(t, g, me.ID, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, tree, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	sweep00PickKey(t, g, me.ID, "choose-cards/realmbreaker-steal-land")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := sweep00PickKey(t, g, me.ID, "choose-cards/realmbreaker-steal-land")
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{land}); err != nil {
		t.Fatalf("ResolveChooseCards on the restored table: %v", err)
	}
	card, ok := battlefieldCard(g, land)
	if !ok || card.Controller != me.ID || !card.Tapped {
		t.Fatalf("the land on the restored table: on the battlefield %v, %+v; want it tapped under the activator", ok, card)
	}
	if err := g.DestroyPermanentForEffect(land); err != nil {
		t.Fatalf("destroy: %v", err)
	}
	if !g.Exile.Contains(land) {
		t.Error("the stolen land did not go to exile instead of leaving, after a restore")
	}
}

// Rinoa, Angel Wing: the dead attacker returns tapped with a flying counter,
// once each turn, from a keyed pick whose source is Rinoa.
func TestSweep05RinoaReturnPickIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := sqActive(g)
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Rinoa, Angel Wing", "Legendary Creature — Human Rebel Warlock", sqRinoaOracle, false)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	declareAttack(t, g, opp.ID, a)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(a) })
	passPriorityAroundTable(t, g)
	sweep00PickKey(t, g, me.ID, "choose-cards/rinoa-angel-wing-return")

	g = restoreRoundTrip(t, g, true)
	me, _ = sqActive(g)
	sweep00PickKey(t, g, me.ID, "choose-cards/rinoa-angel-wing-return")
	sqAnswerCards(t, g, me.ID, a)
	passPriorityAroundTable(t, g)
	card, ok := battlefieldCard(g, a)
	if !ok || !card.Tapped || card.Counters[game.CounterFlying] != 1 {
		t.Errorf("the attacker on the restored table: on the battlefield %v, tapped %v, flying counters %d; want tapped with one",
			ok, card.Tapped, card.Counters[game.CounterFlying])
	}
}

// Dim Oubliette's creature card from the graveyard.
func TestSweep05RoomChosenCreatureReturnIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	one := pushGraveyardPermanent(me, "Old Bear", "Creature — Bear", "{1}{G}")
	two := pushGraveyardPermanent(me, "Old Wolf", "Creature — Wolf", "{1}{G}")
	g.WithWriteLock(func() {
		item := &game.StackItem{Kind: game.StackItemSpell, Controller: me.ID, Owner: me.ID}
		if err := roomsBReturnAChosenCreatureCard(NewContext(g, item)); err != nil {
			t.Fatal(err)
		}
	})
	sweep00PickKey(t, g, me.ID, "choose-cards/rooms-b-return-chosen-creature")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	answerChooseCards(t, g, me.ID, two)
	if !onBattlefield(g, two) || onBattlefield(g, one) {
		t.Errorf("on the battlefield after the restore: chosen %v, other %v; want only the chosen creature",
			onBattlefield(g, two), onBattlefield(g, one))
	}
}

// Rickety Gazebo (and the Rooms that take milled cards): up to N matching
// cards go to the hand from a keyed pick.
func TestSweep05RoomTakeUpToFromMilledIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	one := pushGraveyardPermanent(me, "Milled Bear", "Creature — Bear", "{1}{G}")
	two := pushGraveyardPermanent(me, "Milled Wolf", "Creature — Wolf", "{1}{G}")
	g.WithWriteLock(func() {
		item := &game.StackItem{Kind: game.StackItemSpell, Controller: me.ID, Owner: me.ID}
		any := func(_ *game.Game, _ uuid.UUID, _ game.Card) bool { return true }
		if err := mayTakeUpToFromAmongThem(NewContext(g, item), []uuid.UUID{one, two}, 2, any, "Take up to two"); err != nil {
			t.Fatal(err)
		}
	})
	sweep00PickKey(t, g, me.ID, "choose-cards/rooms-d-take-up-to-from-milled")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	answerChooseCards(t, g, me.ID, one, two)
	if !me.Hand.Contains(one) || !me.Hand.Contains(two) {
		t.Error("the two chosen cards were not put into the hand on the restored table")
	}
}

// Jace, Multiverse Architect (and Identity Echo's loot): the card tucked on
// the bottom of the library.
func TestSweep05PutACardFromHandOnTheBottomIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	keep := handCardForTest(me, "Keep", "Instant", "")
	tuck := handCardForTest(me, "Tuck", "Instant", "")
	g.WithWriteLock(func() {
		item := &game.StackItem{Kind: game.StackItemSpell, Controller: me.ID, Owner: me.ID}
		if err := rfPutCardFromHandOnBottom(NewContext(g, item), "Put a card from your hand on the bottom of your library"); err != nil {
			t.Fatal(err)
		}
	})
	sweep00PickKey(t, g, me.ID, "choose-cards/reality-fracture-hand-card-on-bottom")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	answerChooseCards(t, g, me.ID, tuck)
	if me.Hand.Contains(tuck) || !me.Hand.Contains(keep) {
		t.Error("the chosen card did not leave the hand on the restored table")
	}
	if n := len(me.Library.Cards); n == 0 || me.Library.Cards[0].InstanceID != tuck {
		t.Error("the chosen card is not on the bottom of the library on the restored table")
	}
}

// Rambling Possum: the saddlers to return are a keyed pick that carries the
// asking item, so the restored answer bounces them.
func TestSweep05RamblingPossumReturnPickIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mount := pushMountForTest(g, me.ID, "Rambling Possum", ramblingPossumOracle, 2)
	saddler := pushCrewerForTest(g, me.ID, "Saddler", 1)
	toMain(t, g)
	if err := saddleAbility(t, g, me.ID, mount, saddler); err != nil {
		t.Fatalf("saddle: %v", err)
	}
	passPriorityAroundTable(t, g)
	declareAttack(t, g, opp.ID, mount)
	passPriorityAroundTable(t, g)
	sweep00PickKey(t, g, me.ID, "choose-cards/rambling-possum-return-saddlers")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	sweep00PickKey(t, g, me.ID, "choose-cards/rambling-possum-return-saddlers")
	answerChooseCards(t, g, me.ID, saddler)
	if !me.Hand.Contains(saddler) || g.Battlefield.Contains(saddler) {
		t.Error("the saddler was not returned to its owner's hand on the restored table")
	}
}
