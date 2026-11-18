package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mid_b14_cards_test.go — Midnight Hunt's Curses and Lynde (S82, #5497).

const (
	s82MidB14Unbinding   = "7e79cb6c-df9a-46d1-b3c6-240297e7d47d"
	s82MidB14Conformity  = "71de5f9a-73e4-4a93-b199-b280687f71eb"
	s82MidB14ShakenFaith = "297494f0-1f79-4f77-b68b-6ed858e21f4e"
	s82MidB14Webs        = "d063f1d4-eb6b-40fa-97b7-f15ea53397fb"
	s82MidB14Obsession   = "ca4b4bf2-ed34-444c-9088-dd1ea488ec4e"
	s82MidB14Silence     = "36e267c5-ea98-4e26-b2b4-3faa05493eb4"
	s82MidB14Surveil     = "4fd061f2-7269-4eea-bc7c-95a123893efa"
	s82MidB14Lynde       = "dc386f6c-bd4a-462a-8d42-0e7a552a4f17"
)

// s82MidB14Curse puts a named Curse onto the battlefield for `owner`,
// attached to `host`.
func s82MidB14Curse(g *game.Game, owner, host *game.Player, name, oracle string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: curseAuraTypeLine, OracleID: oracle,
		Owner: owner.ID, Controller: owner.ID,
		AttachedTo: game.TargetRef{Kind: game.TargetPlayer, ID: host.ID},
	})
}

func s82MidB14HandSize(p *game.Player) int { return len(p.Hand.Cards) }

// --- Curse of Obsession ----------------------------------------------------

func TestS82MidB14ObsessionDrawsTwoMoreAndDiscardsTheHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	s82MidB14Curse(g, me, opp, "Curse of Obsession", s82MidB14Obsession)

	// Another player's draw step is untouched.
	advanceToDrawStepOfSeat(t, g, 2)
	before := s82MidB14HandSize(other)
	passPriorityAroundTable(t, g)
	if got := s82MidB14HandSize(other); got != before {
		t.Errorf("a bystander's draw step moved the hand: %d -> %d", before, got)
	}

	advanceToDrawStepOfSeat(t, g, 1)
	before = s82MidB14HandSize(opp)
	passPriorityAroundTable(t, g)
	if got := s82MidB14HandSize(opp); got != before+2 {
		t.Errorf("enchanted player's hand = %d, want %d (two additional cards)", got, before+2)
	}

	advanceToEndStepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if got := s82MidB14HandSize(opp); got != 0 {
		t.Errorf("enchanted player's hand at their end step = %d, want 0", got)
	}
	if len(opp.Graveyard.Cards) < before+2 {
		t.Errorf("the discarded hand is not in the graveyard: %d cards", len(opp.Graveyard.Cards))
	}
}

// --- Curse of Unbinding ----------------------------------------------------

func TestS82MidB14UnbindingStealsTheFirstCreatureAndGravesTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	s82MidB14Curse(g, me, opp, "Curse of Unbinding", s82MidB14Unbinding)

	var bearID uuid.UUID
	g.WithWriteLock(func() {
		bearID = uuid.New()
		opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Deep Creature", TypeLine: "Creature — Test", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
		opp.Library.PushTop(game.Card{InstanceID: bearID, Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
		opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest A", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID})
		opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest B", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID})
	})
	libBefore := opp.Library.Size()

	advanceToUpkeepOfSeat(t, g, 1)
	passPriorityAroundTable(t, g)

	bear, ok := battlefieldCard(g, bearID)
	if !ok {
		t.Fatal("the first creature card was not put onto the battlefield")
	}
	if bear.Controller != me.ID || bear.Owner != opp.ID {
		t.Errorf("Bear controlled by %s, owned by %s; want the Curse's controller, owned by the enchanted player", bear.Controller, bear.Owner)
	}
	if got := libBefore - opp.Library.Size(); got != 3 {
		t.Errorf("library lost %d cards, want 3 (two revealed lands and the creature)", got)
	}
	lands := 0
	for _, c := range opp.Graveyard.Cards {
		if c.Name == "Forest A" || c.Name == "Forest B" {
			lands++
		}
	}
	if lands != 2 {
		t.Errorf("the rest of the revealed cards: %d lands in their graveyard, want 2", lands)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Deep Creature" {
			t.Error("the second creature card was put onto the battlefield too")
		}
	}
}

func TestS82MidB14UnbindingWithNoCreatureGravesTheWholeLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	s82MidB14Curse(g, me, opp, "Curse of Unbinding", s82MidB14Unbinding)
	size := opp.Library.Size()
	advanceToUpkeepOfSeat(t, g, 1)
	passPriorityAroundTable(t, g)
	if opp.Library.Size() != 0 {
		t.Errorf("library = %d cards, want 0: no creature card to stop at", opp.Library.Size())
	}
	if len(opp.Graveyard.Cards) < size {
		t.Errorf("graveyard holds %d cards, want at least %d", len(opp.Graveyard.Cards), size)
	}
}

// --- Curse of Shaken Faith -------------------------------------------------

func TestS82MidB14ShakenFaithPunishesEverySpellButTheFirst(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	s82MidB14Curse(g, me, opp, "Curse of Shaken Faith", s82MidB14ShakenFaith)
	advanceToMainOf(t, g, 1)

	life := opp.Life
	cast := func() {
		castCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
		passPriorityAroundTable(t, g)
	}
	cast()
	if opp.Life != life {
		t.Fatalf("the first spell dealt damage: %d -> %d", life, opp.Life)
	}
	cast()
	if opp.Life != life-2 {
		t.Errorf("after the second spell life = %d, want %d", opp.Life, life-2)
	}
	cast()
	if opp.Life != life-4 {
		t.Errorf("after the third spell life = %d, want %d", opp.Life, life-4)
	}
}

func TestS82MidB14ShakenFaithIgnoresOtherPlayersAndPunishesACopy(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	s82MidB14Curse(g, me, opp, "Curse of Shaken Faith", s82MidB14ShakenFaith)

	// A copy made by the enchanted player is "copies a spell", whatever the cast count.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventSpellCopied, Actor: opp.ID, CardID: uuid.New()})
	})
	life := opp.Life
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("a copy by the enchanted player: life = %d, want %d", opp.Life, life-2)
	}
	// A copy made by anyone else does not.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventSpellCopied, Actor: other.ID, CardID: uuid.New()})
	})
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("another player's copy hurt the enchanted player: %d", opp.Life)
	}
}

// --- Curse of Clinging Webs ------------------------------------------------

func TestS82MidB14ClingingWebsExilesTheDeadAndMakesASpider(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	s82MidB14Curse(g, me, opp, "Curse of Clinging Webs", s82MidB14Webs)
	victim := pushVanillaCreature(g, opp.ID, "Victim", 2, 2)
	mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)

	b15Destroy(t, g, mine)
	passPriorityAroundTable(t, g)
	if spiders := battlefieldIDsNamed(g, "Spider"); len(spiders) != 0 {
		t.Fatalf("a creature the Curse's controller controls made a Spider: %d", len(spiders))
	}

	b15Destroy(t, g, victim)
	passPriorityAroundTable(t, g)
	if !s82MidB14InExile(g, "Victim") {
		t.Error("the dead creature was not exiled")
	}
	spiders := battlefieldIDsNamed(g, "Spider")
	if len(spiders) != 1 {
		t.Fatalf("Spiders = %d, want 1", len(spiders))
	}
	sp, _ := battlefieldCard(g, spiders[0])
	if sp.Controller != me.ID || sp.Power != 1 || sp.Toughness != 2 || !game.HasKeyword(&sp, "reach") {
		t.Errorf("Spider = %+v, want a 1/2 reach token for the Curse's controller", sp)
	}
}

func TestS82MidB14ClingingWebsIgnoresTokens(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	s82MidB14Curse(g, me, opp, "Curse of Clinging Webs", s82MidB14Webs)
	tok := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Token Creature — Goblin", Power: 1, Toughness: 1,
		Owner: opp.ID, Controller: opp.ID,
	})
	b15Destroy(t, g, tok)
	passPriorityAroundTable(t, g)
	if spiders := battlefieldIDsNamed(g, "Spider"); len(spiders) != 0 {
		t.Errorf("a dying token made %d Spiders", len(spiders))
	}
}

func s82MidB14InExile(g *game.Game, name string) bool {
	for _, c := range g.Exile.Cards {
		if c.Name == name {
			return true
		}
	}
	return false
}

// --- Curse of Conformity ---------------------------------------------------

func TestS82MidB14ConformityMakesNonlegendaryCreaturesThreeThreeAndTypeless(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	curse := s82MidB14Curse(g, me, opp, "Curse of Conformity", s82MidB14Conformity)

	push := func(owner *game.Player, name, tl string, p, th int) uuid.UUID {
		return pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: name, TypeLine: tl, Power: p, Toughness: th,
			Owner: owner.ID, Controller: owner.ID,
		})
	}
	elf := push(opp, "Elf", "Creature — Elf Druid", 1, 1)
	king := push(opp, "King", "Legendary Creature — Human Noble", 1, 1)
	theirs := push(other, "Bystander", "Creature — Bear", 2, 2)
	mine := push(me, "Mine", "Creature — Bear", 2, 2)

	pt := func(id uuid.UUID) [2]int {
		c := layeredCard(t, g, id)
		return [2]int{c.CurrentPower(), c.CurrentToughness()}
	}
	if got := pt(elf); got != [2]int{3, 3} {
		t.Errorf("Elf is %v, want 3/3", got)
	}
	if subs := effectiveSubtypes(t, g, elf); len(subs) != 0 {
		t.Errorf("Elf kept its creature types: %v", subs)
	}
	if got := pt(king); got != [2]int{1, 1} {
		t.Errorf("a legendary creature changed to %v", got)
	}
	if subs := effectiveSubtypes(t, g, king); len(subs) != 2 {
		t.Errorf("a legendary creature lost its types: %v", subs)
	}
	if got := pt(theirs); got != [2]int{2, 2} {
		t.Errorf("a bystander's creature changed to %v", got)
	}
	if got := pt(mine); got != [2]int{2, 2} {
		t.Errorf("the Curse controller's creature changed to %v", got)
	}

	// A counter still applies on top of the base 3/3 (CR 613.4b, 613.4d).
	boosted := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Boosted", TypeLine: "Creature — Elf", Power: 1, Toughness: 1,
		Counters: map[string]int{game.CounterPlusOne: 1}, Owner: opp.ID, Controller: opp.ID,
	})
	if got := pt(boosted); got != [2]int{4, 4} {
		t.Errorf("a creature with a counter is %v, want 4/4", got)
	}

	// The Curse leaving ends it.
	b15Destroy(t, g, curse)
	if got := pt(elf); got != [2]int{1, 1} {
		t.Errorf("Elf after the Curse left is %v, want 1/1", got)
	}
	if got := pt(boosted); got != [2]int{2, 2} {
		t.Errorf("a creature with a counter after the Curse left is %v, want 2/2", got)
	}
}

// --- Curse of Silence ------------------------------------------------------

func TestS82MidB14SilenceTaxesAndOffersASacrificeForACard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	curse := castCatalogSpell(t, g, "Curse of Silence", curseAuraTypeLine, s82MidB14Silence,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	actNameFor(t, g, me.ID, "Lightning Bolt")
	if host := attachmentHostOf(t, g, curse); host.Kind != game.TargetPlayer || host.ID != opp.ID {
		t.Fatalf("attached to %+v, want the opponent", host)
	}

	price := func(p *game.Player, name string) string {
		var total string
		g.ReadSnapshot(func() {
			c := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Instant", ManaCost: "{R}", Owner: p.ID, Controller: p.ID}
			pr, err := g.PriceCastForEffect(p.ID, c, game.CastSpellParams{})
			if err != nil {
				t.Fatalf("PriceCast: %v", err)
			}
			total = pr.Total.String()
		})
		return total
	}
	if got := price(opp, "Lightning Bolt"); got != "{2}{R}" {
		t.Errorf("the enchanted player's named spell costs %s, want {2}{R}", got)
	}
	if got := price(opp, "Shock"); got != "{R}" {
		t.Errorf("an unnamed spell costs %s, want {R}", got)
	}
	if got := price(me, "Lightning Bolt"); got != "{R}" {
		t.Errorf("the Curse's controller's named spell costs %s, want {R}", got)
	}
}

func TestS82MidB14SilenceSacrificeDrawsOnlyIfYouSacrificeIt(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		curse := castCatalogSpell(t, g, "Curse of Silence", curseAuraTypeLine, s82MidB14Silence,
			[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
		passPriorityAroundTable(t, g)
		actNameFor(t, g, me.ID, "Test Sorcery")

		advanceToMainOf(t, g, 1)
		cardID := uuid.New()
		opp.Hand.PushTop(game.Card{InstanceID: cardID, Name: "Test Sorcery", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
		if err := g.AddManaForEffect(opp.ID, uuid.Nil, "{C}{C}"); err != nil {
			t.Fatal(err)
		}
		if err := g.CastSpell(opp.ID, cardID, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		before := s82MidB14HandSize(me)
		passPriorityAroundTable(t, g)
		answerLatestTriggerPrompt(t, g, me.ID, accept)
		passPriorityAroundTable(t, g)
		_, onBF := battlefieldCard(g, curse)
		gained := s82MidB14HandSize(me) - before
		if accept && (onBF || gained != 1) {
			t.Errorf("accepted: on battlefield %v, cards drawn %d; want sacrificed and 1 drawn", onBF, gained)
		}
		if !accept && (!onBF || gained != 0) {
			t.Errorf("declined: on battlefield %v, cards drawn %d; want it kept and none drawn", onBF, gained)
		}
	}
}

// --- Curse of Surveillance -------------------------------------------------

func TestS82MidB14SurveillanceDrawsForEachCurseOnThatPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, a, b := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	s82MidB14Curse(g, me, opp, "Curse of Surveillance", s82MidB14Surveil)
	s82MidB14Curse(g, a, opp, "Curse of Opulence", curseOpulenceOracle)
	s82MidB14Curse(g, a, b, "Curse of Opulence", curseOpulenceOracle) // not on the enchanted player

	advanceToUpkeepOfSeat(t, g, 1)
	monSettle(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	before := map[uuid.UUID]int{me.ID: s82MidB14HandSize(me), a.ID: s82MidB14HandSize(a), opp.ID: s82MidB14HandSize(opp), b.ID: s82MidB14HandSize(b)}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: me.ID}, {Kind: game.TargetPlayer, ID: a.ID},
	}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	monSettle(t, g)
	for _, p := range []*game.Player{me, a} {
		if got := s82MidB14HandSize(p) - before[p.ID]; got != 2 {
			t.Errorf("%s drew %d, want 2 (two Curses attached to the enchanted player)", p.Name, got)
		}
	}
	if got := s82MidB14HandSize(b) - before[b.ID]; got != 0 {
		t.Errorf("an untargeted player drew %d", got)
	}
	if got := s82MidB14HandSize(opp) - before[opp.ID]; got != 0 {
		t.Errorf("the enchanted player drew %d", got)
	}
}

func TestS82MidB14SurveillanceCannotTargetTheEnchantedPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	s82MidB14Curse(g, me, opp, "Curse of Surveillance", s82MidB14Surveil)
	advanceToUpkeepOfSeat(t, g, 1)
	monSettle(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err == nil {
		t.Error("the enchanted player was accepted as a target")
	}
}

// --- Lynde, Cheerful Tormentor ---------------------------------------------

func TestS82MidB14LyndeReturnsADeadCurseAttachedToYouAtTheNextEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, me.ID, "Lynde, Cheerful Tormentor", s82MidB14Lynde, "Legendary Creature — Human Warlock")
	curse := s82MidB14Curse(g, me, opp, "Curse of Opulence", curseOpulenceOracle)

	b15Destroy(t, g, curse)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, curse); ok {
		t.Fatal("the Curse came back at once; it returns at the beginning of the next end step")
	}
	advanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, curse)
	if !ok {
		for _, e := range g.Events {
			t.Logf("event %s %s", e.Kind, e.Label)
		}
		_, inGY := cardInZone(me.Graveyard, curse)
		t.Logf("in graveyard: %v; pending: %d; stack: %d", inGY, len(g.PendingChoices), len(g.Stack.Cards))
		t.Fatal("the Curse did not return at the end step")
	}
	if c.AttachedTo.Kind != game.TargetPlayer || c.AttachedTo.ID != me.ID {
		t.Errorf("the Curse is attached to %+v, want Lynde's controller", c.AttachedTo)
	}
}

func TestS82MidB14LyndeIgnoresAnotherPlayersCurseAndNonCurses(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, me.ID, "Lynde, Cheerful Tormentor", s82MidB14Lynde, "Legendary Creature — Human Warlock")
	theirs := s82MidB14Curse(g, opp, me, "Curse of Opulence", curseOpulenceOracle)
	aura := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Pacifism", TypeLine: "Enchantment — Aura", Owner: me.ID, Controller: me.ID,
		AttachedTo: game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID},
	})
	b15Destroy(t, g, theirs)
	b15Destroy(t, g, aura)
	advanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, theirs); ok {
		t.Error("an opponent's Curse returned: it went to their graveyard, not Lynde's controller's")
	}
	if _, ok := battlefieldCard(g, aura); ok {
		t.Error("a non-Curse Aura returned")
	}
}

func TestS82MidB14LyndeDoesNotFollowACurseThatLeftTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, me.ID, "Lynde, Cheerful Tormentor", s82MidB14Lynde, "Legendary Creature — Human Warlock")
	curse := s82MidB14Curse(g, me, opp, "Curse of Opulence", curseOpulenceOracle)
	b15Destroy(t, g, curse)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(curse); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	advanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, curse); ok {
		t.Error("a Curse exiled from the graveyard came back")
	}
}

func TestS82MidB14LyndeUpkeepMovesACurseToAnOpponentAndDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, me.ID, "Lynde, Cheerful Tormentor", s82MidB14Lynde, "Legendary Creature — Human Warlock")
	curse := s82MidB14Curse(g, opp, me, "Curse of Opulence", curseOpulenceOracle)
	advanceToUpkeepOfSeat(t, g, 0)
	before := s82MidB14HandSize(me)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	pick := latestOptionPickFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pair prompt")
	}
	idx := -1
	for i, o := range pick.PickOptions {
		if o.Player == opp.ID {
			idx = i
		}
	}
	if idx < 1 {
		t.Fatalf("the opponent is not offered: %+v", pick.PickOptions)
	}
	answerOptionPick(t, g, me.ID, idx)
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, curse)
	if c.AttachedTo.Kind != game.TargetPlayer || c.AttachedTo.ID != opp.ID {
		t.Errorf("the Curse is attached to %+v, want the opponent", c.AttachedTo)
	}
	if got := s82MidB14HandSize(me) - before; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
}

func TestS82MidB14LyndeUpkeepDeclineAndNoCurseDrawNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, me.ID, "Lynde, Cheerful Tormentor", s82MidB14Lynde, "Legendary Creature — Human Warlock")
	curse := s82MidB14Curse(g, opp, me, "Curse of Opulence", curseOpulenceOracle)
	advanceToUpkeepOfSeat(t, g, 0)
	before := s82MidB14HandSize(me)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, curse); c.AttachedTo.ID != me.ID {
		t.Error("the Curse moved although the player declined")
	}
	if s82MidB14HandSize(me) != before {
		t.Error("declining drew cards")
	}
}
