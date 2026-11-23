package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_dmu_b16_cards_test.go — Dominaria United batch 16 (#5487): Ajani,
// Sleeper Agent; Jaya, Fiery Negotiator; Dihada, Binder of Wills; Meria,
// Scholar of Antiquity; Jhoira, Ageless Innovator; Weatherlight Compleated.

const (
	s82B16AjaniOracle   = "37714eb9-a39d-4534-a2d3-27908c418f8a"
	s82B16JayaOracle    = "2458aa66-5b20-4811-a4a4-8375ad0a6498"
	s82B16DihadaOracle  = "d326de4d-6157-4586-91b8-5668e1381ac3"
	s82B16MeriaOracle   = "24aaadda-ef1c-44eb-9b01-4847bf454cd5"
	s82B16JhoiraOracle  = "83b9a734-ee92-44c8-bf47-ca69b6d6405e"
	s82B16WeatherOracle = "060a7d77-61a8-4964-8fb6-35bd9a999517"
)

func TestS82DmuB16CardsAreRegisteredFull(t *testing.T) {
	for name, oracle := range map[string]string{
		"Ajani, Sleeper Agent":        s82B16AjaniOracle,
		"Jaya, Fiery Negotiator":      s82B16JayaOracle,
		"Dihada, Binder of Wills":     s82B16DihadaOracle,
		"Meria, Scholar of Antiquity": s82B16MeriaOracle,
		"Jhoira, Ageless Innovator":   s82B16JhoiraOracle,
		"Weatherlight Compleated":     s82B16WeatherOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// s82B16LibraryTop pushes cards onto the top of p's library, the last name on top.
func s82B16LibraryTop(p *game.Player, typeLine string, names ...string) []uuid.UUID {
	ids := make([]uuid.UUID, len(names))
	for i, n := range names {
		ids[i] = uuid.New()
		p.Library.PushTop(game.Card{InstanceID: ids[i], Name: n, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	}
	return ids
}

func s82B16Zone(g *game.Game, id uuid.UUID) game.ZoneKind {
	var k game.ZoneKind
	g.WithWriteLock(func() {
		if z := g.FindCardZoneForEffect(id); z != nil {
			k = z.Kind
		}
	})
	return k
}

// --- Ajani, Sleeper Agent -----------------------------------------------------------

func TestS82DmuB16AjaniCompleatedAndPlusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	ajani := effectstest.S82CastWalker(t, g, "Ajani, Sleeper Agent", s82B16AjaniOracle, "Legendary Planeswalker — Ajani",
		"{1}{G}{G/W/P}{W}", 4, 1, "CGW")
	if got := effectstest.LoyaltyOf(g, ajani); got != 2 {
		t.Errorf("loyalty %d after paying life for the hybrid Phyrexian symbol, want 2", got)
	}
	if me.Life != life-2 {
		t.Errorf("life %d -> %d, want 2 paid", life, me.Life)
	}

	// A creature on top goes to the hand.
	top := s82B16LibraryTop(me, "Creature — Bear", "Bear")[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, ajani, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if s82B16Zone(g, top) != game.ZoneHand {
		t.Errorf("the revealed creature is in zone %v, want the hand", s82B16Zone(g, top))
	}
	if got := effectstest.LoyaltyOf(g, ajani); got != 3 {
		t.Errorf("loyalty %d after +1, want 3", got)
	}
}

func TestS82DmuB16AjaniPlusOneNonCreatureMayGoToTheBottom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ajani := effectstest.S82Walker(g, me.ID, "Ajani, Sleeper Agent", s82B16AjaniOracle, 4)
	s82B16LibraryTop(me, "Creature — Bear", "Below")
	spell := s82B16LibraryTop(me, "Sorcery", "Spell")[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, ajani, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	if s82B16Zone(g, spell) != game.ZoneLibrary || me.Library.Cards[len(me.Library.Cards)-1].InstanceID != spell {
		t.Fatalf("declining left the card %v, want on top of the library", s82B16Zone(g, spell))
	}

	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	ajani = effectstest.S82Walker(g, me.ID, "Ajani, Sleeper Agent", s82B16AjaniOracle, 4)
	below := s82B16LibraryTop(me, "Creature — Bear", "Below")[0]
	spell = s82B16LibraryTop(me, "Sorcery", "Spell")[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, ajani, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	if me.Library.Cards[0].InstanceID != spell {
		t.Errorf("accepting left %s at the bottom, want the revealed card", me.Library.Cards[0].Name)
	}
	if s82B16Zone(g, below) != game.ZoneLibrary {
		t.Error("the card beneath moved")
	}
}

func TestS82DmuB16AjaniMinusThreeDistributesCountersAndGivesVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ajani := effectstest.S82Walker(g, me.ID, "Ajani, Sleeper Agent", s82B16AjaniOracle, 4)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, ajani, 1, game.ActivateAbilityParams{
		Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 2, b: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.BattlefieldCountersOf(g, a, game.CounterPlusOne); n != 2 {
		t.Errorf("Bear A has %d +1/+1 counters, want 2", n)
	}
	if n := effectstest.BattlefieldCountersOf(g, b, game.CounterPlusOne); n != 1 {
		t.Errorf("Bear B has %d +1/+1 counters, want 1", n)
	}
	for _, id := range []uuid.UUID{a, b} {
		if !effectstest.HasKeywordOf(g, id, "vigilance") {
			t.Errorf("creature %s has no vigilance", id)
		}
	}
	if got := effectstest.LoyaltyOf(g, ajani); got != 1 {
		t.Errorf("loyalty %d after −3, want 1", got)
	}
}

func TestS82DmuB16AjaniEmblemPoisonsOnCreatureAndPlaneswalkerSpellsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	ajani := effectstest.S82Walker(g, me.ID, "Ajani, Sleeper Agent", s82B16AjaniOracle, 6)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, ajani, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if len(effectstest.EmblemsOf(g, me.ID)) != 1 {
		t.Fatalf("emblems %v, want one", effectstest.EmblemsOf(g, me.ID))
	}
	effectstest.FloatForTest(g, me, "CG")
	effectstest.CastFromHandForTest(t, g, me, "Bear", "Creature — Bear", "{1}{G}", "", game.CastSpellParams{Strict: true})
	s82B16ResolveWithPoisonTarget(t, g, me, opp.ID)
	if got := opp.Counters[game.CounterPoison]; got != 2 {
		t.Errorf("poison %d after a creature spell, want 2", got)
	}
	effectstest.FloatForTest(g, me, "CU")
	effectstest.CastFromHandForTest(t, g, me, "Divination", "Sorcery", "{1}{U}", "", game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Counters[game.CounterPoison]; got != 2 {
		t.Errorf("poison %d after a sorcery, want still 2", got)
	}
	// A planeswalker spell triggers it too.
	effectstest.FloatForTest(g, me, "CCCC")
	effectstest.CastFromHandForTest(t, g, me, "Some Walker", "Legendary Planeswalker — Test", "{4}", "", game.CastSpellParams{Strict: true})
	s82B16ResolveWithPoisonTarget(t, g, me, opp.ID)
	if got := opp.Counters[game.CounterPoison]; got != 4 {
		t.Errorf("poison %d after a planeswalker spell, want 4", got)
	}
}

// s82B16ResolveWithPoisonTarget answers the emblem trigger's "target
// opponent" prompt with victim and resolves everything on the stack.
func s82B16ResolveWithPoisonTarget(t *testing.T, g *game.Game, me *game.Player, victim uuid.UUID) {
	t.Helper()
	for i := 0; i < 24; i++ {
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: victim}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if effectstest.StackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never emptied")
}

// --- Jaya, Fiery Negotiator ---------------------------------------------------------

func TestS82DmuB16JayaPlusOneMakesAProwessMonk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	jaya := effectstest.S82Walker(g, me.ID, "Jaya, Fiery Negotiator", s82B16JayaOracle, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, jaya, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	monk := effectstest.FindBattlefieldByName(g, "Monk")
	if monk == uuid.Nil {
		t.Fatal("no Monk token")
	}
	if p, th := effectstest.EffectivePTOf(t, g, monk); p != 1 || th != 1 {
		t.Errorf("Monk %d/%d, want 1/1", p, th)
	}
	if !effectstest.HasKeywordOf(g, monk, "prowess") {
		t.Error("the Monk has no prowess")
	}
	if got := effectstest.LoyaltyOf(g, jaya); got != 5 {
		t.Errorf("loyalty %d, want 5", got)
	}
}

func TestS82DmuB16JayaMinusOneLetsYouPlayOneOfTheTwoExiledCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	jaya := effectstest.S82Walker(g, me.ID, "Jaya, Fiery Negotiator", s82B16JayaOracle, 4)
	ids := s82B16LibraryTop(me, "Sorcery", "Card C", "Card B", "Card A")
	below := ids[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, jaya, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range ids[1:] {
		if s82B16Zone(g, id) != game.ZoneExile {
			t.Fatalf("a top card is in zone %v, want exile", s82B16Zone(g, id))
		}
	}
	if s82B16Zone(g, below) != game.ZoneLibrary {
		t.Error("the third card left the library")
	}
	chosen := ids[2]
	effectstest.AnswerChooseCards(t, g, me.ID, chosen)
	granted := map[uuid.UUID]bool{}
	for _, perm := range me.CastPermissions {
		for _, ref := range perm.Cards {
			granted[ref.ID] = true
		}
	}
	if !granted[chosen] || len(granted) != 1 {
		t.Errorf("play permission over %v, want only %s", granted, chosen)
	}
	if got := effectstest.LoyaltyOf(g, jaya); got != 3 {
		t.Errorf("loyalty %d, want 3", got)
	}
}

func TestS82DmuB16JayaMinusTwoDamagesEachAttackThisTurnByTheAttackerCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	jaya := effectstest.S82Walker(g, me.ID, "Jaya, Fiery Negotiator", s82B16JayaOracle, 4)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Their Ox", "Creature — Ox", 1, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, jaya, 2, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(victim)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyOf(g, jaya); got != 2 {
		t.Errorf("loyalty %d, want 2", got)
	}
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, victim)
	if !ok {
		t.Fatal("the Ox died to two damage with toughness 3")
	}
	if c.DamageMarked != 2 {
		t.Errorf("Ox has %d damage, want 2 (two attacking creatures)", c.DamageMarked)
	}
}

func TestS82DmuB16JayaMinusTwoDoesNotTriggerOnAnotherTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	jaya := effectstest.S82Walker(g, me.ID, "Jaya, Fiery Negotiator", s82B16JayaOracle, 4)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Their Ox", "Creature — Ox", 1, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, jaya, 2, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(victim)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B39NextTurnOf(t, g, 0)
	effectstest.DeclareAttack(t, g, opp.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, victim); ok && c.DamageMarked != 0 {
		t.Errorf("Ox has %d damage from an attack on a later turn, want 0", c.DamageMarked)
	}
}

func TestS82DmuB16JayaEmblemCopiesRedInstantsTwice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	jaya := effectstest.S82Walker(g, me.ID, "Jaya, Fiery Negotiator", s82B16JayaOracle, 8)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, jaya, 3, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if len(effectstest.EmblemsOf(g, me.ID)) != 1 {
		t.Fatalf("emblems %v, want one", effectstest.EmblemsOf(g, me.ID))
	}
	before := effectstest.LifeOf(g, victim.ID)
	s82B16CastRedBoltKeepingCopyTargets(t, g, me, victim.ID, 2)
	if got := effectstest.LifeOf(g, victim.ID); got != before-9 {
		t.Errorf("victim life %d -> %d, want −9 (a Bolt and two copies)", before, got)
	}
	// A nonred instant is not copied.
	before = effectstest.LifeOf(g, victim.ID)
	bolt := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", OracleID: effectstest.LightningBoltOracle,
			Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})
	})
	effectstest.FloatForTest(g, me, "R")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Fatal("a nonred spell was copied")
	}
	if got := effectstest.LifeOf(g, victim.ID); got != before-3 {
		t.Errorf("victim life %d -> %d, want −3 (no copies)", before, got)
	}
}

// s82B16CastRedBoltKeepingCopyTargets casts a red Lightning Bolt at victim and
// keeps the target for each of the `copies` copies the emblem makes.
func s82B16CastRedBoltKeepingCopyTargets(t *testing.T, g *game.Game, me *game.Player, victim uuid.UUID, copies int) {
	t.Helper()
	bolt := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", OracleID: effectstest.LightningBoltOracle,
			Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	})
	effectstest.FloatForTest(g, me, "R")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: victim}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	answered := 0
	for i := 0; i < 48; i++ {
		if p := effectstest.TriggerOrderPromptFor(g, me.ID); p != nil {
			if err := g.ResolveTriggerOrder(p.ID, me.ID, append([]uuid.UUID(nil), p.TriggerOrderIDs...)); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: victim}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			answered++
			continue
		}
		if effectstest.StackFullyEmpty(g) {
			break
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if answered != copies {
		t.Fatalf("answered %d copy-target prompts, want %d", answered, copies)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// --- Dihada, Binder of Wills ---------------------------------------------------------

func TestS82DmuB16DihadaPlusTwoLastsUntilYourNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dihada := effectstest.S82Walker(g, me.ID, "Dihada, Binder of Wills", s82B16DihadaOracle, 5)
	hero := effectstest.B12Creature(g, me.ID, "Hero", "Legendary Creature — Human", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, dihada, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.S82TargetCard(plain)}); err == nil {
		t.Error("a nonlegendary creature was a legal target")
	}
	effectstest.S82Activate(t, g, me.ID, dihada, 0, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(hero)})
	effectstest.PassPriorityAroundTable(t, g)
	for _, kw := range []string{"vigilance", "lifelink", "indestructible"} {
		if !effectstest.HasKeywordOf(g, hero, kw) {
			t.Errorf("Hero lacks %s", kw)
		}
	}
	if got := effectstest.LoyaltyOf(g, dihada); got != 7 {
		t.Errorf("loyalty %d, want 7", got)
	}
	effectstest.B39NextTurnOf(t, g, 1)
	if !effectstest.HasKeywordOf(g, hero, "indestructible") {
		t.Error("the keywords ended before your next turn")
	}
	effectstest.B39NextTurnOf(t, g, 0)
	if effectstest.HasKeywordOf(g, hero, "indestructible") {
		t.Error("the keywords outlasted your next turn's start")
	}
}

func TestS82DmuB16DihadaMinusThreeTakesLegendsAndTreasuresForTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dihada := effectstest.S82Walker(g, me.ID, "Dihada, Binder of Wills", s82B16DihadaOracle, 5)
	below := s82B16LibraryTop(me, "Creature — Bear", "Below")[0]
	rest1 := s82B16LibraryTop(me, "Sorcery", "Rest One")[0]
	leg1 := s82B16LibraryTop(me, "Legendary Creature — Human", "Legend One")[0]
	rest2 := s82B16LibraryTop(me, "Land", "Rest Two")[0]
	leg2 := s82B16LibraryTop(me, "Legendary Artifact", "Legend Two")[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, dihada, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, leg1, leg2)
	for _, id := range []uuid.UUID{leg1, leg2} {
		if s82B16Zone(g, id) != game.ZoneHand {
			t.Errorf("a legendary card is in %v, want the hand", s82B16Zone(g, id))
		}
	}
	for _, id := range []uuid.UUID{rest1, rest2} {
		if s82B16Zone(g, id) != game.ZoneGraveyard {
			t.Errorf("a nonlegendary card is in %v, want the graveyard", s82B16Zone(g, id))
		}
	}
	if s82B16Zone(g, below) != game.ZoneLibrary {
		t.Error("the fifth card left the library")
	}
	if got := effectstest.CountOnBattlefield(g, "Treasure", me.ID); got != 2 {
		t.Errorf("%d Treasures, want 2 for the two cards put into the graveyard", got)
	}
}

func TestS82DmuB16DihadaMinusThreeWithNoLegendsBinsAllFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dihada := effectstest.S82Walker(g, me.ID, "Dihada, Binder of Wills", s82B16DihadaOracle, 5)
	ids := s82B16LibraryTop(me, "Sorcery", "A", "B", "C", "D")
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, dihada, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range ids {
		if s82B16Zone(g, id) != game.ZoneGraveyard {
			t.Errorf("a card is in %v, want the graveyard", s82B16Zone(g, id))
		}
	}
	if got := effectstest.CountOnBattlefield(g, "Treasure", me.ID); got != 4 {
		t.Errorf("%d Treasures, want 4", got)
	}
}

func TestS82DmuB16DihadaMinusElevenStealsUntapsAndHastesNonlandPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	dihada := effectstest.S82Walker(g, me.ID, "Dihada, Binder of Wills", s82B16DihadaOracle, 11)
	ox := effectstest.B12Creature(g, opp.ID, "Their Ox", "Creature — Ox", 3, 3)
	rock := effectstest.DftABPush(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)
	land := effectstest.DftABPush(g, opp.ID, "Their Forest", "Basic Land — Forest", "", 0, 0)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{ox, rock, land} {
			effectstest.BattlefieldCardPtr(t, g, id).Tapped = true
		}
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, dihada, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{ox, rock} {
		c, _ := effectstest.BattlefieldCard(g, id)
		if c.Controller != me.ID {
			t.Errorf("%s is controlled by %v, want Dihada's controller", c.Name, c.Controller)
		}
		if c.Tapped {
			t.Errorf("%s is still tapped", c.Name)
		}
		if !effectstest.HasKeywordOf(g, id, "haste") {
			t.Errorf("%s has no haste", c.Name)
		}
	}
	if c, _ := effectstest.BattlefieldCard(g, land); c.Controller != opp.ID || !c.Tapped {
		t.Errorf("the land was touched: controller %v tapped %v", c.Controller, c.Tapped)
	}
	effectstest.B39NextTurnOf(t, g, 1)
	if c, _ := effectstest.BattlefieldCard(g, ox); c.Controller != opp.ID {
		t.Error("control did not return at end of turn")
	}
}

// --- Meria, Scholar of Antiquity -----------------------------------------------------

func TestS82DmuB16MeriaTapsNontokenArtifactsForManaAndExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	meria := effectstest.DftABPush(g, me.ID, "Meria, Scholar of Antiquity", "Legendary Creature — Elf Artificer", s82B16MeriaOracle, 3, 3)
	r1 := effectstest.DftABPush(g, me.ID, "Rock One", "Artifact", "", 0, 0)
	r2 := effectstest.DftABPush(g, me.ID, "Rock Two", "Artifact", "", 0, 0)
	r3 := effectstest.DftABPush(g, me.ID, "Rock Three", "Artifact", "", 0, 0)
	token := effectstest.DftABPush(g, me.ID, "Clue Token", "Token Artifact — Clue", "", 0, 0)
	effectstest.AdvanceToMain(t, g)

	if err := g.ActivateManaAbility(me.ID, meria, 0, game.ManaAbilityParams{TapIDs: []uuid.UUID{token}}); err == nil {
		t.Error("a token artifact paid the mana ability")
	}
	if err := g.ActivateManaAbility(me.ID, meria, 0, game.ManaAbilityParams{TapIDs: []uuid.UUID{r1}}); err != nil {
		t.Fatalf("tap a nontoken artifact for {G}: %v", err)
	}
	if me.ManaPool.Total() != 1 {
		t.Errorf("pool %d, want {G}", me.ManaPool.Total())
	}
	if c, _ := effectstest.BattlefieldCard(g, r1); !c.Tapped {
		t.Error("the artifact is untapped")
	}

	top := s82B16LibraryTop(me, "Sorcery", "Top")[0]
	if err := g.ActivateCatalogAbility(me.ID, meria, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{r2, token}}); err == nil {
		t.Error("a token artifact paid the two-artifact cost")
	}
	effectstest.B16Activate(t, g, me.ID, meria, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{r2, r3}})
	effectstest.PassPriorityAroundTable(t, g)
	if s82B16Zone(g, top) != game.ZoneExile {
		t.Fatalf("the top card is in %v, want exile", s82B16Zone(g, top))
	}
	granted := false
	for _, perm := range me.CastPermissions {
		for _, ref := range perm.Cards {
			granted = granted || ref.ID == top
		}
	}
	if !granted {
		t.Error("no permission to play the exiled card")
	}
}

// --- Jhoira, Ageless Innovator -------------------------------------------------------

func TestS82DmuB16JhoiraPutsArtifactsUpToTheGrowingCounterCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	jhoira := effectstest.DftABPush(g, me.ID, "Jhoira, Ageless Innovator", "Legendary Creature — Human Artificer", s82B16JhoiraOracle, 2, 3)
	three := effectstest.HandCardForTest(me, "Three Rock", "Artifact", "")
	two := effectstest.HandCardForTest(me, "Two Rock", "Artifact", "")
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			switch me.Hand.Cards[i].InstanceID {
			case three:
				me.Hand.Cards[i].ManaCost = "{3}"
			case two:
				me.Hand.Cards[i].ManaCost = "{2}"
			}
		}
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, jhoira, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.BattlefieldCountersOf(g, jhoira, "ingenuity"); got != 2 {
		t.Fatalf("%d ingenuity counters, want 2", got)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, two)
	if s82B16Zone(g, two) != game.ZoneBattlefield {
		t.Errorf("the mana value 2 artifact is in %v, want the battlefield", s82B16Zone(g, two))
	}
	if s82B16Zone(g, three) != game.ZoneHand {
		t.Error("the mana value 3 artifact left the hand on the first activation")
	}

	g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, jhoira).Tapped = false })
	effectstest.B16Activate(t, g, me.ID, jhoira, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.BattlefieldCountersOf(g, jhoira, "ingenuity"); got != 4 {
		t.Fatalf("%d ingenuity counters, want 4", got)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, three)
	if s82B16Zone(g, three) != game.ZoneBattlefield {
		t.Errorf("the mana value 3 artifact is in %v, want the battlefield", s82B16Zone(g, three))
	}
}

// --- Weatherlight Compleated ---------------------------------------------------------

func TestS82DmuB16WeatherlightBecomesAPhyrexianCreatureAtFourCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ship := effectstest.DftABPush(g, me.ID, "Weatherlight Compleated", "Legendary Artifact — Vehicle", s82B16WeatherOracle, 5, 5)
	g.RunStateChecksForTest()
	if effectstest.HasEffectiveType(t, g, ship, "Creature") {
		t.Fatal("a Vehicle with no counters is a creature")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ship, "phyresis", 3) })
	if effectstest.HasEffectiveType(t, g, ship, "Creature") {
		t.Fatal("three counters made it a creature")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ship, "phyresis", 1) })
	g.RunStateChecksForTest()
	if !effectstest.HasEffectiveType(t, g, ship, "Creature") {
		t.Fatalf("four counters did not make it a creature (counters %d)", effectstest.BattlefieldCountersOf(g, ship, "phyresis"))
	}
	if !effectstest.HasEffectiveType(t, g, ship, "Artifact") {
		t.Error("it lost its artifact type")
	}
	found := false
	for _, st := range effectstest.LayeredCard(t, g, ship).Effective().Subtypes {
		found = found || st == "Phyrexian"
	}
	if !found {
		t.Error("it is not a Phyrexian")
	}
	if p, th := effectstest.EffectivePTOf(t, g, ship); p != 5 || th != 5 {
		t.Errorf("%d/%d, want 5/5", p, th)
	}
}

// s82B16WeatherlightDeath kills a fresh creature and settles the trigger,
// returning the scry prompt it queued, if any.
func s82B16WeatherlightDeath(t *testing.T, g *game.Game, me *game.Player) *game.PendingChoice {
	t.Helper()
	bear := effectstest.B12Creature(g, me.ID, "Doomed", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	effectstest.PassPriorityAroundTable(t, g)
	return effectstest.LatestChoiceOfKind(g, game.PendingChoiceScry)
}

func TestS82DmuB16WeatherlightPutsACounterThenScriesBelowSeven(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ship := effectstest.DftABPush(g, me.ID, "Weatherlight Compleated", "Legendary Artifact — Vehicle", s82B16WeatherOracle, 5, 5)
	g.RunStateChecksForTest()
	s82B16LibraryTop(me, "Creature — Bear", "A", "B", "C")
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	for i := 1; i <= 2; i++ {
		sc := s82B16WeatherlightDeath(t, g, me)
		if sc == nil {
			t.Fatalf("death %d: no scry", i)
		}
		if err := g.ResolveScry(sc.ID, me.ID, nil, sc.ScryCards); err != nil {
			t.Fatalf("ResolveScry: %v", err)
		}
		if got := effectstest.BattlefieldCountersOf(g, ship, "phyresis"); got != i {
			t.Fatalf("after death %d: %d phyresis counters", i, got)
		}
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, drew with fewer than seven counters", hand, me.Hand.Size())
	}
}

func TestS82DmuB16WeatherlightDrawsAtSevenCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ship := effectstest.DftABPush(g, me.ID, "Weatherlight Compleated", "Legendary Artifact — Vehicle", s82B16WeatherOracle, 5, 5)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ship, "phyresis", 6) })
	g.RunStateChecksForTest()
	s82B16LibraryTop(me, "Creature — Bear", "A", "B", "C")
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	if sc := s82B16WeatherlightDeath(t, g, me); sc != nil {
		t.Fatal("scried with seven counters")
	}
	if got := effectstest.BattlefieldCountersOf(g, ship, "phyresis"); got != 7 {
		t.Errorf("%d phyresis counters, want 7", got)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want a draw", hand, me.Hand.Size())
	}
}

// Weatherlight Compleated dying itself (a creature at four counters) sees
// its own death, cannot take the counter, and reads the counters it last had.
func TestS82DmuB16WeatherlightDyingItselfReadsItsLastKnownCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ship := effectstest.DftABPush(g, me.ID, "Weatherlight Compleated", "Legendary Artifact — Vehicle", s82B16WeatherOracle, 5, 5)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ship, "phyresis", 7) })
	g.RunStateChecksForTest()
	s82B16LibraryTop(me, "Creature — Bear", "A", "B", "C")
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(ship) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want a draw from its own death with seven counters last known", hand, me.Hand.Size())
	}
}
