package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mom_b23_test.go — March of the Machine Sieges, S82 batch 23:
// Invasion of Alara, Amonkhet, Arcavios, Gobakhan, Ikoria, Kaldheim and
// Ravnica with their back faces, and Begin the Invasion.

const beginTheInvasionOracleID = "4feca1da-dcc8-440c-8933-bab91782b015"

func TestS82MomB23EveryCardIsRegisteredFull(t *testing.T) {
	for key, name := range map[string]string{
		invasionOfAlaraOracleID:           "Invasion of Alara",
		invasionOfAlaraOracleID + "#1":    "Awaken the Maelstrom",
		invasionOfAmonkhetOracleID:        "Invasion of Amonkhet",
		invasionOfAmonkhetOracleID + "#1": "Lazotep Convert",
		invasionOfArcaviosOracleID:        "Invasion of Arcavios",
		invasionOfArcaviosOracleID + "#1": "Invocation of the Founders",
		invasionOfGobakhanOracleID:        "Invasion of Gobakhan",
		invasionOfGobakhanOracleID + "#1": "Lightshield Array",
		invasionOfIkoriaOracleID:          "Invasion of Ikoria",
		invasionOfIkoriaOracleID + "#1":   "Zilortha, Apex of Ikoria",
		invasionOfKaldheimOracleID:        "Invasion of Kaldheim",
		invasionOfKaldheimOracleID + "#1": "Pyre of the World Tree",
		invasionOfRavnicaOracleID:         "Invasion of Ravnica",
		invasionOfRavnicaOracleID + "#1":  "Guildpact Paragon",
		beginTheInvasionOracleID:          "Begin the Invasion",
	} {
		spec, ok := Lookup(key)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- fixtures ------------------------------------------------------------------

func b23Row(oracleID, front, back, cost, defense, backType, power, toughness string) cards.Card {
	return siegeRow(oracleID, front, back, cost, defense, backType, power, toughness)
}

// b23Enter puts a Siege onto the battlefield from the active seat's hand and
// names the next seat its protector. The enters trigger is left on the stack.
func b23Enter(t *testing.T, g *game.Game, row cards.Card) (id uuid.UUID, owner, protector *game.Player) {
	t.Helper()
	seat := g.Turn.ActiveSeat
	owner = g.Seats[seat]
	protector = g.Seats[(seat+1)%len(g.Seats)]
	id = importToHand(row, owner)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: owner.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, id); err != nil {
		t.Fatalf("move the Siege to the battlefield: %v", err)
	}
	answerProtectorPrompt(t, g, owner.ID, protector.ID)
	return id, owner, protector
}

func b23Permission(g *game.Game, id uuid.UUID) (perm game.CastPermission, ok bool) {
	g.ReadSnapshot(func() {
		if p := g.CastPermissionOnCardByIDForEffect(id); p != nil {
			perm, ok = *p, true
		}
	})
	return perm, ok
}

// b23Library replaces a player's library with cs, the first card listed on
// top, and returns their ids in that order.
func b23Library(p *game.Player, cs ...game.Card) []uuid.UUID {
	p.Library.Cards = nil
	var ids []uuid.UUID
	for _, c := range cs {
		c.Owner, c.Controller = p.ID, p.ID
		if c.InstanceID == uuid.Nil {
			c.InstanceID = uuid.New()
		}
		ids = append(ids, pushLibraryCardForTest(p, c))
	}
	return ids
}

// --- Invasion of Kaldheim // Pyre of the World Tree ----------------------------

func kaldheimRow() cards.Card {
	return b23Row(invasionOfKaldheimOracleID, "Invasion of Kaldheim", "Pyre of the World Tree", "{3}{R}", "4", "Enchantment", "", "")
}

func TestInvasionOfKaldheimExilesTheHandDrawsThatManyAndLetsYouPlayThemUntilYourNextTurnEnds(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	owner.Hand.Cards = nil
	hand := []uuid.UUID{
		pushCatalogHandCard(owner, "Forest", "Basic Land — Forest", ""),
		pushCatalogHandCard(owner, "Bear", "Creature — Bear", ""),
		pushCatalogHandCard(owner, "Bolt", "Instant", ""),
	}
	b23Library(owner, game.Card{Name: "A", TypeLine: "Instant"}, game.Card{Name: "B", TypeLine: "Instant"}, game.Card{Name: "C", TypeLine: "Instant"}, game.Card{Name: "D", TypeLine: "Instant"})
	libBefore := owner.Library.Size()
	_, _, _ = b23Enter(t, g, kaldheimRow())
	passPriorityAroundTable(t, g)

	for _, id := range hand {
		if !g.Exile.Contains(id) {
			t.Errorf("%s was not exiled", id)
		}
		if owner.Hand.Contains(id) {
			t.Errorf("%s is still in hand", id)
		}
		perm, ok := b23Permission(g, id)
		if !ok {
			t.Errorf("%s has no play permission", id)
			continue
		}
		if perm.Player != owner.ID {
			t.Errorf("permission held by %v, want the Siege's controller", perm.Player)
		}
		if perm.Duration.Kind != game.UntilEndOfTurn || perm.Duration.ExpiresAfterTurnsBegun == 0 {
			t.Errorf("the window is %+v; want until the end of your next turn", perm.Duration)
		}
	}
	if got := owner.Hand.Size(); got != 3 {
		t.Errorf("hand = %d cards, want 3 drawn", got)
	}
	if got := libBefore - owner.Library.Size(); got != 3 {
		t.Errorf("drew %d, want 3", got)
	}
}

func TestInvasionOfKaldheimWithAnEmptyHandDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	owner.Hand.Cards = nil
	libBefore := owner.Library.Size()
	_, _, _ = b23Enter(t, g, kaldheimRow())
	passPriorityAroundTable(t, g)
	if owner.Hand.Size() != 0 || owner.Library.Size() != libBefore {
		t.Errorf("an empty hand drew: hand %d, library %d -> %d", owner.Hand.Size(), libBefore, owner.Library.Size())
	}
}

func TestPyreOfTheWorldTreeDiscardsALandToShockAndImpulsesTheTopCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pyre := pushCatalogPermanent(g, me.ID, "Pyre of the World Tree", "Enchantment", invasionOfKaldheimOracleID+"#1", false)
	forest := pushCatalogHandCard(me, "Forest", "Basic Land — Forest", "")
	bear := pushCatalogHandCard(me, "Bear", "Creature — Bear", "")
	top := b23Library(me, game.Card{Name: "Top", TypeLine: "Sorcery"}, game.Card{Name: "Next", TypeLine: "Sorcery"})[0]

	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	if err := g.ActivateCatalogAbility(me.ID, pyre, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{bear}, Targets: target, Strict: true}); err == nil {
		t.Fatal("a nonland card was accepted as the discarded land card")
	}
	before := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, pyre, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{forest}, Targets: target, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Graveyard.Contains(forest) {
		t.Error("the land was not discarded as the cost")
	}
	passPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 2 {
		t.Errorf("damage = %d, want 2", got)
	}
	if !g.Exile.Contains(top) {
		t.Fatal("the top card was not exiled by the discard trigger")
	}
	if perm, ok := b23Permission(g, top); !ok || perm.Player != me.ID {
		t.Errorf("the exiled card should be playable by its controller this turn: %+v %v", perm, ok)
	}
}

// --- Invasion of Gobakhan // Lightshield Array ----------------------------------

func gobakhanRow() cards.Card {
	return b23Row(invasionOfGobakhanOracleID, "Invasion of Gobakhan", "Lightshield Array", "{1}{W}", "3", "Enchantment", "", "")
}

func TestInvasionOfGobakhanTaxesANonlandCardItsOwnerMayPlay(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	var ids []uuid.UUID
	g.WithWriteLock(func() {
		ids = revealHand(opp, rhForest(), game.Card{Name: "Spark", TypeLine: "Instant", ManaCost: "{R}"})
	})
	forest, spark := ids[0], ids[1]
	_, _, _ = b23Enter(t, g, gobakhanRow())
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	pick := openVariant(t, g)
	if pick.Chooser != me.ID || slices.Contains(pick.DiscardOptions, forest) || !slices.Contains(pick.DiscardOptions, spark) {
		t.Fatalf("pick: chooser %v options %v; want the Siege's controller choosing among nonland cards", pick.Chooser, pick.DiscardOptions)
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{spark}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	monSettle(t, g)
	if !g.Exile.Contains(spark) {
		t.Fatal("the chosen card was not exiled")
	}
	perm, ok := b23Permission(g, spark)
	if !ok || perm.Player != opp.ID || perm.CostMore != 2 {
		t.Errorf("permission = %+v (%v); want its owner's, costing {2} more", perm, ok)
	}
}

func TestLightshieldArrayCountersEachCreatureThatAttackedAtYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Lightshield Array", "Enchantment", invasionOfGobakhanOracleID+"#1", false)
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Test", 2, 2)
	idle := b12Creature(g, me.ID, "Idle", "Creature — Test", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)

	if got := battlefieldCardFor(g, attacker).Counters[game.CounterPlusOne]; got != 1 {
		t.Errorf("the attacker has %d +1/+1 counters, want 1", got)
	}
	if got := battlefieldCardFor(g, idle).Counters[game.CounterPlusOne]; got != 0 {
		t.Errorf("a creature that did not attack has %d counters, want none", got)
	}
}

func TestLightshieldArraySacrificeGrantsHexproofAndIndestructibleToYourCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	array := pushCatalogPermanent(g, me.ID, "Lightshield Array", "Enchantment", invasionOfGobakhanOracleID+"#1", false)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Test", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Test", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, array, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(array) {
		t.Error("the enchantment was not sacrificed")
	}
	for _, kw := range []string{"hexproof", "indestructible"} {
		if !hasKeywordOnBattlefield(t, g, mine, kw) {
			t.Errorf("your creature lacks %s", kw)
		}
		if hasKeywordOnBattlefield(t, g, theirs, kw) {
			t.Errorf("an opponent's creature gained %s", kw)
		}
	}
}

// --- Invasion of Amonkhet // Lazotep Convert ------------------------------------

func amonkhetRow() cards.Card {
	return b23Row(invasionOfAmonkhetOracleID, "Invasion of Amonkhet", "Lazotep Convert", "{1}{U}{B}", "4", "Creature — Zombie", "4", "4")
}

func TestInvasionOfAmonkhetMillsEveryoneThenOpponentsDiscardAndYouDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var opps []*game.Player
	for _, p := range g.Seats {
		if p.ID != me.ID {
			opps = append(opps, p)
		}
	}
	gy := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		var fill []game.Card
		for i := 0; i < 5; i++ {
			fill = append(fill, game.Card{Name: "Filler", TypeLine: "Sorcery"})
		}
		b23Library(p, fill...)
		gy[p.ID] = p.Graveyard.Size()
	}
	for _, p := range opps {
		p.Hand.Cards = nil
		pushCatalogHandCard(p, "Card1", "Sorcery", "")
		pushCatalogHandCard(p, "Card2", "Sorcery", "")
	}
	_, _, _ = b23Enter(t, g, amonkhetRow())
	myHand := me.Hand.Size()
	libBefore := me.Library.Size()
	passPriorityAroundTable(t, g)
	for _, p := range opps {
		answerDiscard(t, g, p.ID, p.Hand.Cards[0].InstanceID)
	}
	passPriorityAroundTable(t, g)

	for _, p := range opps {
		if got := p.Graveyard.Size() - gy[p.ID]; got != 4 {
			t.Errorf("%s graveyard grew by %d, want 3 milled + 1 discarded", p.Name, got)
		}
	}
	if got := me.Graveyard.Size() - gy[me.ID]; got != 3 {
		t.Errorf("your graveyard grew by %d, want 3 milled", got)
	}
	for _, p := range opps {
		if p.Hand.Size() != 1 {
			t.Errorf("%s has %d cards in hand, want 1 after discarding", p.Name, p.Hand.Size())
		}
	}
	if got := me.Hand.Size() - myHand; got != 1 {
		t.Errorf("you drew %d, want 1", got)
	}
	if got := libBefore - me.Library.Size(); got != 3+1 {
		t.Errorf("your library lost %d cards, want 3 milled and 1 drawn", got)
	}
}

func TestLazotepConvertEntersAsA4x4BlackZombieCopyOfACreatureCardInAGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	giant := graveyardCreature(opp, "Hill Giant", "Giant", 3, 3)
	g.WithWriteLock(func() {
		for i := range opp.Graveyard.Cards {
			if opp.Graveyard.Cards[i].InstanceID == giant {
				opp.Graveyard.Cards[i].Colors = []string{"G"}
			}
		}
	})
	convert := castGraveyardCopier(t, g, "Lazotep Convert", "Creature — Zombie", invasionOfAmonkhetOracleID+"#1", 4, 4)
	resolveWithCopyChoice(t, g, giant)

	got := copyBattlefieldCard(t, g, convert)
	if got.Name != "Hill Giant" {
		t.Errorf("name = %q, want a copy of Hill Giant", got.Name)
	}
	if got.CurrentPower() != 4 || got.CurrentToughness() != 4 {
		t.Errorf("size = %d/%d, want 4/4", got.CurrentPower(), got.CurrentToughness())
	}
	for _, sub := range []string{"Giant", "Zombie"} {
		if !got.HasSubtype(sub) {
			t.Errorf("the copy should be a %s", sub)
		}
	}
	colors := got.EffectiveColors()
	if !slices.Contains(colors, "B") || !slices.Contains(colors, "G") {
		t.Errorf("colors = %v, want black in addition to the copied green", colors)
	}
	if !opp.Graveyard.Contains(giant) {
		t.Error("the card stays in the graveyard")
	}
}

// --- Invasion of Ravnica // Guildpact Paragon -----------------------------------

func ravnicaRow() cards.Card {
	return b23Row(invasionOfRavnicaOracleID, "Invasion of Ravnica", "Guildpact Paragon", "{5}", "4", "Artifact Creature — Construct", "5", "5")
}

func TestInvasionOfRavnicaExilesANonlandPermanentThatIsNotExactlyTwoColors(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	mono := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Mono", TypeLine: "Creature — Test", ManaCost: "{R}", Colors: []string{"R"}, Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	gold := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Gold", TypeLine: "Creature — Test", ManaCost: "{R}{G}", Colors: []string{"R", "G"}, Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", ManaCost: "{2}", Owner: opp.ID, Controller: opp.ID})
	land := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island", Owner: opp.ID, Controller: opp.ID})
	_, owner, _ := b23Enter(t, g, ravnicaRow())

	var prompt *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			prompt = c
		}
	}
	if prompt == nil {
		t.Fatal("the enters trigger did not ask for a target")
	}
	for _, bad := range []uuid.UUID{gold, land} {
		if slices.Contains(prompt.PickTargetCards, bad) {
			t.Errorf("%s offered as a target", bad)
		}
	}
	for _, good := range []uuid.UUID{mono, rock} {
		if !slices.Contains(prompt.PickTargetCards, good) {
			t.Errorf("%s not offered as a target", good)
		}
	}
	if err := g.ResolvePickTarget(prompt.ID, owner.ID, game.TargetRef{Kind: game.TargetCard, ID: gold}); err == nil {
		t.Error("an exactly-two-colour permanent was accepted as the target")
	}
	if err := g.ResolvePickTarget(prompt.ID, owner.ID, game.TargetRef{Kind: game.TargetCard, ID: mono}); err != nil {
		t.Fatalf("target the mono-coloured permanent: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(mono) {
		t.Error("the target was not exiled")
	}
	for _, id := range []uuid.UUID{gold, rock, land} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s left the battlefield", id)
		}
	}
}

func TestGuildpactParagonLooksForAnExactlyTwoColorCardWhenYouCastOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	paragon := pushCatalogPermanent(g, me.ID, "Guildpact Paragon", "Artifact Creature — Construct", invasionOfRavnicaOracleID+"#1", false)
	// Top first: the top six are Deep..X2; X3 is the seventh card.
	ids := b23Library(me,
		game.Card{Name: "Deep", TypeLine: "Sorcery"},
		game.Card{Name: "Gold", TypeLine: "Sorcery", ManaCost: "{W}{U}", Colors: []string{"W", "U"}},
		game.Card{Name: "Mono", TypeLine: "Sorcery", ManaCost: "{W}", Colors: []string{"W"}},
		game.Card{Name: "Three", TypeLine: "Sorcery", ManaCost: "{W}{U}{B}", Colors: []string{"W", "U", "B"}},
		game.Card{Name: "X1", TypeLine: "Sorcery"}, game.Card{Name: "X2", TypeLine: "Sorcery"}, game.Card{Name: "X3", TypeLine: "Sorcery"},
	)
	gold := ids[1]

	mono := pushCatalogHandCard(me, "Mono Spell", "Sorcery", "")
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == mono {
				me.Hand.Cards[i].ManaCost, me.Hand.Cards[i].Colors = "{W}", []string{"W"}
			}
		}
	})
	if err := g.CastSpell(me.ID, mono, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast a one-colour spell: %v", err)
	}
	if triggerOnStack(g, paragon) != nil {
		t.Fatal("a one-colour spell triggered Guildpact Paragon")
	}
	passPriorityAroundTable(t, g)

	two := pushCatalogHandCard(me, "Two Spell", "Sorcery", "")
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == two {
				me.Hand.Cards[i].ManaCost, me.Hand.Cards[i].Colors = "{W}{U}", []string{"W", "U"}
			}
		}
	})
	if err := g.CastSpell(me.ID, two, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast a two-colour spell: %v", err)
	}
	if triggerOnStack(g, paragon) == nil {
		t.Fatal("a two-colour spell did not trigger Guildpact Paragon")
	}
	passPriorityAroundTable(t, g)
	choice := pendingChooseCards(t, g, me.ID)
	if len(choice.ChooseCards) != 1 || choice.ChooseCards[0] != gold {
		t.Fatalf("offered %v, want only the exactly-two-colour card among the top six", choice.ChooseCards)
	}
	if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{gold}); err != nil {
		t.Fatalf("take it: %v", err)
	}
	if !me.Hand.Contains(gold) {
		t.Error("the card did not go into the hand")
	}
	if got := me.Library.Cards[len(me.Library.Cards)-1].InstanceID; got != ids[6] {
		t.Error("the seventh card, not looked at, is no longer on top")
	}
	bottom := me.Library.Cards[:5]
	for _, id := range []uuid.UUID{ids[0], ids[2], ids[3], ids[4], ids[5]} {
		if !slices.ContainsFunc(bottom, func(c game.Card) bool { return c.InstanceID == id }) {
			t.Errorf("%s was not put on the bottom", id)
		}
	}
}
