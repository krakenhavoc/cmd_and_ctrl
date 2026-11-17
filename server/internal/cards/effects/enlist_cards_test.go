package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// enlist_cards_test.go — the eleven Dominaria United enlist cards
// (#5487, seam #5658). The engine half of enlist (CR 702.154) is
// game/enlist_test.go; these tests play the real catalog entries.

const (
	argivianCavalierOracle     = "c45090bd-a589-45a0-8388-025b8e0b3487"
	balduvianBerserkerOracle   = "3a7b1d1c-1724-4208-943b-650a5d773653"
	barkweaveCrusherOracle     = "cbac75c6-6006-4025-b893-8828965b5f97"
	benalishFaithbonderOracle  = "278db664-4235-40f0-a926-3074a0e038c1"
	coalitionSkyknightOracle   = "dc763f20-73e7-4db7-9e98-6f419b8fba84"
	coalitionWarbruteOracle    = "84d67d87-2696-4c2f-9f10-29d5832bfbb6"
	guardianOfNewBenaliaOracle = "33747a34-79cc-446f-8718-b499bd497687"
	hexbaneTortoiseOracle      = "f3b4097f-b84f-4645-81f1-018366e40eeb"
	keldonFlamesageOracle      = "0efedec3-b57c-4395-ad72-19d79560c99f"
	linebreakerBalothOracle    = "c3b395b3-fea0-41d9-a3f2-6c90f89d8bdb"
	yavimayaSteelcrusherOracle = "ad181ee7-756f-4002-9009-023ea6172991"
)

type enlistCard struct {
	name, typeLine, oracle string
	power, toughness       int
	keywords               []string
}

var enlistCards = []enlistCard{
	{"Argivian Cavalier", "Creature — Orc Knight", argivianCavalierOracle, 2, 2, nil},
	{"Balduvian Berserker", "Creature — Kor Berserker", balduvianBerserkerOracle, 1, 3, nil},
	{"Barkweave Crusher", "Creature — Elemental Warrior", barkweaveCrusherOracle, 2, 5, nil},
	{"Benalish Faithbonder", "Creature — Human Cleric", benalishFaithbonderOracle, 1, 3, []string{"vigilance"}},
	{"Coalition Skyknight", "Creature — Human Knight", coalitionSkyknightOracle, 2, 2, []string{"flying"}},
	{"Coalition Warbrute", "Creature — Minotaur Berserker", coalitionWarbruteOracle, 3, 4, []string{"trample"}},
	{"Guardian of New Benalia", "Creature — Human Soldier", guardianOfNewBenaliaOracle, 2, 2, nil},
	{"Hexbane Tortoise", "Creature — Turtle", hexbaneTortoiseOracle, 3, 2, []string{"ward"}},
	{"Keldon Flamesage", "Creature — Human Shaman", keldonFlamesageOracle, 2, 3, nil},
	{"Linebreaker Baloth", "Creature — Beast", linebreakerBalothOracle, 4, 5, nil},
	{"Yavimaya Steelcrusher", "Creature — Ape Warrior", yavimayaSteelcrusherOracle, 2, 2, nil},
}

func enlistCardByOracle(t *testing.T, oracle string) enlistCard {
	t.Helper()
	for _, c := range enlistCards {
		if c.oracle == oracle {
			return c
		}
	}
	t.Fatalf("no enlist card %s", oracle)
	return enlistCard{}
}

// enlistPush puts one of the cards onto seat 0's battlefield, able to
// attack.
func enlistPush(t *testing.T, g *game.Game, oracle string) uuid.UUID {
	t.Helper()
	c := enlistCardByOracle(t, oracle)
	return b12Push(g, g.Seats[0].ID, c.name, c.typeLine, c.oracle, c.power, c.toughness)
}

// enlistDeclare declares attacker at seat 1, enlisting `enlisted`.
func enlistDeclare(t *testing.T, g *game.Game, attacker, enlisted uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttackerDeclWith(game.AttackDeclaration{Attacker: attacker, Target: g.Seats[1].ID, Enlist: enlisted}, game.DeclareAttackersParams{}); err != nil {
		t.Fatalf("declare with enlist: %v", err)
	}
}

// enlistSettle locks the declaration in and resolves what it put on the
// stack. A trigger-order prompt is answered so the trigger whose label
// contains `first` resolves first, and a scry keeps the cards on top. It
// stops at any other prompt.
func enlistSettle(t *testing.T, g *game.Game, first string) {
	t.Helper()
	g.RunStateChecksForTest()
	for i := 0; i < 64; i++ {
		if c := latestChoiceOfKind(g, game.PendingChoiceTriggerOrder); c != nil {
			order := append([]uuid.UUID(nil), c.TriggerOrderIDs...)
			for j, id := range order {
				if first != "" && strings.Contains(enlistItemLabel(g, id), first) {
					order[0], order[j] = order[j], order[0]
					break
				}
			}
			if err := g.ResolveTriggerOrder(c.ID, c.Chooser, order); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		if c := scryChoiceFor(g, g.Seats[0].ID); c != nil {
			if err := g.ResolveScry(c.ID, c.Chooser, nil, c.ScryCards); err != nil {
				t.Fatalf("ResolveScry: %v", err)
			}
			continue
		}
		if len(g.PendingChoices) > 0 || stackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) {
				continue
			}
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the declaration's triggers never settled")
}

func enlistItemLabel(g *game.Game, id uuid.UUID) string {
	for _, it := range g.PendingTriggers {
		if it != nil && it.ID == id {
			return it.Label
		}
	}
	if it := g.StackMeta[id]; it != nil {
		return it.Label
	}
	return ""
}

func TestEnlistCardsAreRegisteredFullWithEnlist(t *testing.T) {
	for _, c := range enlistCards {
		spec, ok := Lookup(c.oracle)
		if !ok {
			t.Errorf("%s is not registered", c.name)
			continue
		}
		if spec.Name != c.name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", c.name, spec.Name, spec.Completeness, spec.Caveats)
		}
		g := newCatalogGame(t)
		id := enlistPush(t, g, c.oracle)
		assertKeywords(t, g, id, append([]string{game.KeywordEnlist}, c.keywords...)...)
	}
}

// Every card: enlisting a 3/3 taps it as the declaration locks in and
// adds 3 to the attacker's power until end of turn; the attacker is
// tapped by the attack unless it has vigilance.
func TestEnlistCardsAddTheEnlistedCreaturesPower(t *testing.T) {
	for _, c := range enlistCards {
		t.Run(c.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			emptyLibraryForEnlist(g, me)
			attacker := enlistPush(t, g, c.oracle)
			elk := b12Creature(g, me.ID, "Elk", "Creature — Elk", 3, 3)
			enlistDeclare(t, g, attacker, elk)
			if tappedNow(g, elk) {
				t.Fatal("the enlisted creature is tapped before the declaration locks in")
			}
			enlistSettle(t, g, "Enlist")
			if !tappedNow(g, elk) {
				t.Fatal("the enlisted creature was not tapped")
			}
			if c, _ := g.LookupCardForEffect(elk); c.AttackingTarget != uuid.Nil {
				t.Error("the enlisted creature is attacking")
			}
			if got := hobA4Power(t, g, attacker); got != c.power+3 {
				t.Errorf("power %d, want %d (+3 from the enlisted Elk)", got, c.power+3)
			}
			if _, tough := sauronPT(t, g, attacker); tough != c.toughness {
				t.Errorf("toughness %d, want %d (+X/+0)", tough, c.toughness)
			}
			vigilant := c.oracle == benalishFaithbonderOracle
			if tappedNow(g, attacker) == vigilant {
				t.Errorf("attacker tapped = %v, want %v", !vigilant, !vigilant)
			}
			// Until end of turn.
			advanceToUpkeepOf(t, g, 1)
			if got := hobA4Power(t, g, attacker); got != c.power {
				t.Errorf("power %d on the next turn, want %d", got, c.power)
			}
		})
	}
}

// emptyLibraryForEnlist empties the library, so Keldon Flamesage's attack
// trigger looks at nothing in the shared enlist test.
func emptyLibraryForEnlist(g *game.Game, p *game.Player) {
	g.WithWriteLock(func() { p.Library.Cards = nil })
}

// Attacking without enlisting taps nothing and adds nothing, and a
// creature that came under your control this turn can't be enlisted.
func TestEnlistIsOptionalAndNeedsACreatureWithoutSummoningSickness(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	crusher := enlistPush(t, g, barkweaveCrusherOracle)
	elk := b12Creature(g, me.ID, "Elk", "Creature — Elk", 3, 3)
	fresh := b12Creature(g, me.ID, "Fresh Elk", "Creature — Elk", 3, 3)
	battlefieldCardPtr(t, g, fresh).SummonedThisTurn = true
	advanceTo(t, g, game.StepDeclareAttackers)
	err := g.DeclareAttackerDeclWith(game.AttackDeclaration{Attacker: crusher, Target: g.Seats[1].ID, Enlist: fresh}, game.DeclareAttackersParams{})
	if !errors.Is(err, game.ErrCantEnlist) {
		t.Fatalf("enlisting a summoning-sick creature: %v, want ErrCantEnlist", err)
	}
	if err := g.DeclareAttacker(crusher, g.Seats[1].ID); err != nil {
		t.Fatalf("attack without enlisting: %v", err)
	}
	enlistSettle(t, g, "")
	if tappedNow(g, elk) || hobA4Power(t, g, crusher) != 2 {
		t.Error("an attack that enlisted nothing tapped a creature or added power")
	}
}

// --- Argivian Cavalier -------------------------------------------------

func TestArgivianCavalierMakesASoldierItCantEnlistThatTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	c := enlistCardByOracle(t, argivianCavalierOracle)
	b12PlayFromHand(t, g, c.name, c.typeLine, c.oracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if n := mshB1Tokens(g, me.ID, "Soldier"); n != 1 {
		t.Fatalf("%d Soldier tokens, want 1", n)
	}
	soldier := mshB1Token(t, g, me.ID, "Soldier")
	if soldier.Power != 1 || soldier.Toughness != 1 || len(soldier.Colors) != 1 || soldier.Colors[0] != "W" {
		t.Errorf("the token is %+v, want a white 1/1", soldier)
	}
	// The Soldier arrived this turn: another enlister can't tap it.
	crusher := enlistPush(t, g, barkweaveCrusherOracle)
	advanceTo(t, g, game.StepDeclareAttackers)
	err := g.DeclareAttackerDeclWith(game.AttackDeclaration{Attacker: crusher, Target: g.Seats[1].ID, Enlist: soldier.InstanceID}, game.DeclareAttackersParams{})
	if !errors.Is(err, game.ErrCantEnlist) {
		t.Errorf("enlisting the new Soldier: %v, want ErrCantEnlist", err)
	}
}

// --- Balduvian Berserker -----------------------------------------------

func TestBalduvianBerserkerDiesDealingItsEnlistedPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	berserker := enlistPush(t, g, balduvianBerserkerOracle)
	elk := b12Creature(g, me.ID, "Elk", "Creature — Elk", 3, 3)
	enlistDeclare(t, g, berserker, elk)
	enlistSettle(t, g, "")
	before := opp.Life
	eerieDestroy(t, g, berserker)
	eerieSettle(t, g, eerieAnswers{Target: game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}})
	if opp.Life != before-4 {
		t.Fatalf("opponent life %d -> %d, want 4 damage (1 + the enlisted Elk's 3)", before, opp.Life)
	}
}

func TestBalduvianBerserkerUnpumpedDealsOneToACreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	berserker := enlistPush(t, g, balduvianBerserkerOracle)
	victim := b12Creature(g, g.Seats[1].ID, "Squire", "Creature — Human", 1, 1)
	eerieDestroy(t, g, berserker)
	eerieSettle(t, g, eerieAnswers{Target: game.TargetRef{Kind: game.TargetCard, ID: victim}})
	if g.Battlefield.Contains(victim) {
		t.Error("1 damage should kill the 1/1")
	}
}

// --- Guardian of New Benalia -------------------------------------------

func TestGuardianOfNewBenaliaScriesWhenItEnlists(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	guardian := enlistPush(t, g, guardianOfNewBenaliaOracle)
	elk := b12Creature(g, me.ID, "Elk", "Creature — Elk", 3, 3)
	enlistDeclare(t, g, guardian, elk)
	g.RunStateChecksForTest()
	labels := map[string]bool{}
	for _, it := range g.PendingTriggers {
		if it != nil && it.SourceCardID == guardian {
			labels[it.Label] = true
		}
	}
	for _, it := range g.StackMeta {
		if it != nil && it.SourceCardID == guardian {
			labels[it.Label] = true
		}
	}
	if len(labels) != 2 || !labels["Guardian of New Benalia — scry 2"] {
		t.Fatalf("triggers from the Guardian = %v, want the enlist pump and the scry", labels)
	}
	// Resolve the scry and look at what it offered.
	for i := 0; i < 16 && scryChoiceFor(g, me.ID) == nil; i++ {
		if c := latestChoiceOfKind(g, game.PendingChoiceTriggerOrder); c != nil {
			if err := g.ResolveTriggerOrder(c.ID, c.Chooser, c.TriggerOrderIDs); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		if err := g.PassPriority(); err != nil && !errors.Is(err, game.ErrChoicePending) {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	c := scryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("scry prompt = %+v, want two cards", c)
	}
	enlistSettle(t, g, "")
	if got := hobA4Power(t, g, guardian); got != 5 {
		t.Errorf("Guardian power %d, want 5", got)
	}
}

func TestGuardianOfNewBenaliaDoesNotScryWithoutEnlisting(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	guardian := enlistPush(t, g, guardianOfNewBenaliaOracle)
	b12Creature(g, me.ID, "Elk", "Creature — Elk", 3, 3)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(guardian, g.Seats[1].ID); err != nil {
		t.Fatal(err)
	}
	g.RunStateChecksForTest()
	if !stackFullyEmpty(g) || scryChoiceFor(g, me.ID) != nil {
		t.Error("an attack that enlisted nothing triggered the Guardian")
	}
}

func TestGuardianOfNewBenaliaDiscardsForIndestructibleAndTaps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMainForCost(t, g)
	guardian := enlistPush(t, g, guardianOfNewBenaliaOracle)
	emptyHandForTest(g, me)
	if err := g.ActivateCatalogAbility(me.ID, guardian, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with an empty hand")
	}
	ids := fillHandForTest(g, me, 2)
	if err := g.ActivateCatalogAbility(me.ID, guardian, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{ids[0]}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Hand.Size() != 1 {
		t.Errorf("hand %d after the discard cost, want 1", me.Hand.Size())
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, guardian, "indestructible") || !tappedNow(g, guardian) {
		t.Error("the Guardian should be indestructible and tapped")
	}
	// A tapped Guardian can activate it again (no {T} in the cost).
	if err := g.ActivateCatalogAbility(me.ID, guardian, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{ids[1]}}); err != nil {
		t.Fatalf("activate while tapped: %v", err)
	}
}

// --- Hexbane Tortoise --------------------------------------------------

func TestHexbaneTortoiseWardTwoCountersAnUnpaidSpell(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	tortoise := enlistPush(t, g, hexbaneTortoiseOracle)
	advanceToMain(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	blade := castAtWardedCreature(t, g, opp, tortoise)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	c := latestChoiceOfKind(g, game.PendingChoicePayUnless)
	if c == nil || c.Chooser != opp.ID || c.PayCost != "{2}" {
		t.Fatalf("ward prompt = %+v, want {2} asked of the spell's controller", c)
	}
	answerPayUnless(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(tortoise) || !opp.Graveyard.Contains(blade) {
		t.Error("a declined ward {2} counters the spell and the Tortoise lives")
	}
}

// --- Keldon Flamesage --------------------------------------------------

// The enlist trigger resolves first, so X is 2 + 3 = 5: five cards are
// looked at, an instant or sorcery with mana value 5 or less may be
// exiled, the rest go to the bottom and the exiled card is cast free.
func TestKeldonFlamesageLooksAtItsEnlistedPowerAndCastsFree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sage := enlistPush(t, g, keldonFlamesageOracle)
	elk := b12Creature(g, me.ID, "Elk", "Creature — Elk", 3, 3)
	ids := sunbirdLibrary(me,
		game.Card{Name: "Small Spell", TypeLine: "Sorcery", ManaCost: "{4}{R}"},
		game.Card{Name: "Big Spell", TypeLine: "Instant", ManaCost: "{5}{R}"},
		game.Card{Name: "A Creature", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "A Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Cheap Spell", TypeLine: "Instant", ManaCost: "{R}"},
		game.Card{Name: "Too Deep", TypeLine: "Instant", ManaCost: "{R}"},
	)
	small, big, bear, forest, cheap, deep := ids[0], ids[1], ids[2], ids[3], ids[4], ids[5]
	enlistDeclare(t, g, sage, elk)
	enlistSettle(t, g, "Enlist")
	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the Flamesage should offer the instants and sorceries among the top five")
	}
	if len(c.ChooseCards) != 2 || !hasID(c.ChooseCards, small) || !hasID(c.ChooseCards, cheap) {
		t.Errorf("offered %v: want the two spells with mana value 5 or less among the top five", c.ChooseCards)
	}
	for _, id := range []uuid.UUID{big, bear, forest, deep} {
		if hasID(c.ChooseCards, id) {
			t.Errorf("offered a card it must not: %v", id)
		}
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{small}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !inExile(g, small) {
		t.Fatal("the chosen card is exiled")
	}
	n := len(me.Library.Cards)
	for _, id := range []uuid.UUID{big, bear, forest, cheap} {
		if i := libraryIndex(me, id); i < 0 || i > 3 {
			t.Errorf("the rest go on the bottom of the library, found at index %d of %d", i, n)
		}
	}
	if i := libraryIndex(me, deep); i != n-1 {
		t.Errorf("the sixth card stays on top, found at index %d of %d", i, n)
	}
	if err := castFromExile(t, g, me.ID, small); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(small) {
		t.Error("the free sorcery should resolve into the graveyard")
	}
}

// Without enlisting X is 2; declining the exile bottoms both cards, and
// with no instant or sorcery among them nothing is asked.
func TestKeldonFlamesageUnenlistedLooksAtTwoAndMayDecline(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sage := enlistPush(t, g, keldonFlamesageOracle)
	ids := sunbirdLibrary(me,
		game.Card{Name: "Cheap Spell", TypeLine: "Instant", ManaCost: "{1}{R}"},
		game.Card{Name: "Three Drop", TypeLine: "Sorcery", ManaCost: "{2}{R}"},
		game.Card{Name: "Third", TypeLine: "Instant", ManaCost: "{R}"},
	)
	cheap, three, third := ids[0], ids[1], ids[2]
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(sage, g.Seats[1].ID); err != nil {
		t.Fatal(err)
	}
	enlistSettle(t, g, "")
	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil || len(c.ChooseCards) != 1 || c.ChooseCards[0] != cheap {
		t.Fatalf("offered %+v, want only the mana value 2 instant of the top two", c)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if inExile(g, cheap) || inExile(g, three) {
		t.Error("declining exiles nothing")
	}
	for _, id := range []uuid.UUID{cheap, three} {
		if i := libraryIndex(me, id); i < 0 || i > 1 {
			t.Errorf("a declined look puts both cards on the bottom, found at %d", i)
		}
	}
	if i := libraryIndex(me, third); i != len(me.Library.Cards)-1 {
		t.Errorf("the third card stays on top, found at %d", i)
	}
}

// --- Linebreaker Baloth ------------------------------------------------

func TestLinebreakerBalothCantBeBlockedByPowerTwoOrLess(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	baloth := enlistPush(t, g, linebreakerBalothOracle)
	small := b12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	big := b12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	brAttack(t, g, baloth)
	brRefusal(t, g.DeclareBlocker(small, baloth), game.BlockReasonCantBeBlockedBy)
	if err := g.DeclareBlocker(big, baloth); err != nil {
		t.Fatalf("a power 3 creature blocks: %v", err)
	}
}

// --- Yavimaya Steelcrusher ---------------------------------------------

func TestYavimayaSteelcrusherSacrificesToDestroyAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	toMainForCost(t, g)
	ape := enlistPush(t, g, yavimayaSteelcrusherOracle)
	rock := b12Permanent(g, opp.ID, "Mind Stone", "Artifact")
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	at := func(id uuid.UUID) game.ActivateAbilityParams {
		return game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: id}}}
	}
	if err := g.ActivateCatalogAbility(me.ID, ape, 0, at(rock)); err == nil {
		t.Fatal("activated without {1}")
	}
	fillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, ape, 0, at(bear)); err == nil {
		t.Fatal("a creature is not an artifact")
	}
	if err := g.ActivateCatalogAbility(me.ID, ape, 0, at(rock)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(ape) || !me.Graveyard.Contains(ape) {
		t.Error("the Steelcrusher is sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact should be destroyed")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature was never a target")
	}
}
