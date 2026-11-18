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

	pt := func(id uuid.UUID) [2]int { return [2]int{effectivePower(t, g, id), effectiveToughness(t, g, id)} }
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
	g.WithWriteLock(func() { _ = g.AddCounterByForEffect(opp.ID, elf, game.CounterPlusOne, 1) })
	if got := pt(elf); got != [2]int{4, 4} {
		t.Errorf("Elf with a counter is %v, want 4/4", got)
	}

	// The Curse leaving ends it: a 1/1 with a counter.
	b15Destroy(t, g, curse)
	if got := pt(elf); got != [2]int{2, 2} {
		t.Errorf("Elf after the Curse left is %v, want 2/2", got)
	}
}
