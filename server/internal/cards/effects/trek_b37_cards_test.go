package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b37_cards_test.go — Star Trek Commander, batch 37 (#4974). Each
// test plays the real catalog entry and asserts what its text says.

const (
	trekB37GaghOracle          = "b5697e53-eb2f-482b-8b4a-1dad165ac16f"
	trekB37TenorOracle         = "d974fee0-7d20-465c-a7e3-bc361e241cd4"
	trekB37GloryOracle         = "22c53662-51b9-4a8c-8ec7-e8b456a268b7"
	trekB37TargOracle          = "1fc59e37-0dc0-4e24-af71-6bab4a2b56ca"
	trekB37GrilkaOracle        = "d8366080-dd0c-40cf-84e9-f628dfe8f45c"
	trekB37KangOracle          = "1ce8bc83-3cda-4ac5-ad9a-2a2a327c5193"
	trekB37KolothOracle        = "01b7ced9-5dcb-45ce-bcda-ffdc1afefb2a"
	trekB37OwnTheDayOracle     = "4a9bfeb2-b4b3-4c69-9739-a878c2704f46"
	trekB37QuartermasterOracle = "b7115ed6-526c-491d-828f-f60b18437361"
	trekB37RoLarenOracle       = "6f10f6bf-f889-4c3d-a6cf-04cf74f1b4bd"
)

func TestTrekB37EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB37GaghOracle:          "Gagh Slinger",
		trekB37TenorOracle:         "Galvanizing Tenor",
		trekB37GloryOracle:         "Glory to the Empire!",
		trekB37TargOracle:          "Gnarling Targ",
		trekB37GrilkaOracle:        "Grilka, Widowed Warrior",
		trekB37KangOracle:          "Kang, Suspicious Captain",
		trekB37KolothOracle:        "Koloth, Tribble Hunter",
		trekB37OwnTheDayOracle:     "Own the Day",
		trekB37QuartermasterOracle: "Qo'noS Quartermaster",
		trekB37RoLarenOracle:       "Ro Laren, Double Agent",
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

// --- fixtures ------------------------------------------------------------------

// trekB37Seats is the active seat and the three opponents in turn order.
func trekB37Seats(g *game.Game) (me *game.Player, opps []*game.Player) {
	me = g.Seats[g.Turn.ActiveSeat]
	for i := 1; i < len(g.Seats); i++ {
		opps = append(opps, g.Seats[(g.Turn.ActiveSeat+i)%len(g.Seats)])
	}
	return me, opps
}

// trekB37Lives is every player's life total, by ID.
func trekB37Lives(g *game.Game) map[uuid.UUID]int {
	out := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		out[p.ID] = p.Life
	}
	return out
}

// trekB37AssertOpponentsLost checks that every opponent of me lost n life
// since `before` and that me lost none.
func trekB37AssertOpponentsLost(t *testing.T, g *game.Game, me uuid.UUID, before map[uuid.UUID]int, n int) {
	t.Helper()
	for _, p := range g.Seats {
		want := before[p.ID] - n
		if p.ID == me {
			want = before[p.ID]
		}
		if p.Life != want {
			t.Errorf("seat %d life %d → %d, want %d", p.Seat, before[p.ID], p.Life, want)
		}
	}
}

// trekB37Declare declares each attacker at its defender and locks the
// declaration in, so the attack triggers go on the stack.
func trekB37Declare(t *testing.T, g *game.Game, attacks map[uuid.UUID]uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	for attacker, defender := range attacks {
		if err := g.DeclareAttacker(attacker, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
}

// trekB37CastFromHand puts a card in the active seat's hand, gives it
// `lands` untapped Mountains, and casts it strictly with auto-tap from a
// main phase. Returns the card's ID and the cast error.
func trekB37CastFromHand(t *testing.T, g *game.Game, c game.Card, lands int) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	if g.Turn.Step != game.StepPostcombatMain {
		advanceToMain(t, g)
	}
	for i := 0; i < lands; i++ {
		pushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")
	}
	c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
	me.Hand.PushTop(c)
	return c.InstanceID, g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true, AutoTap: true})
}

// trekB37Tokens is every token named `name` that p controls.
func trekB37Tokens(g *game.Game, p uuid.UUID, name string) []game.Card {
	var out []game.Card
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		for _, c := range g.Battlefield.Cards {
			if c.Controller == p && c.IsToken() && c.Name == name {
				out = append(out, c)
			}
		}
	})
	return out
}

// trekB37Goaders is everyone who has goaded the creature, any way.
func trekB37Goaders(t *testing.T, g *game.Game, id uuid.UUID) []uuid.UUID {
	t.Helper()
	var by []uuid.UUID
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			t.Fatalf("%s not found", id)
		}
		by = c.AllGoaders()
	})
	return by
}

// --- Gagh Slinger ------------------------------------------------------------------

func TestTrekB37GaghSlingerDiscountsKlingonAndWarriorSpellsOnce(t *testing.T) {
	cases := []struct {
		name, typeLine, cost string
		lands                int
		ok                   bool
	}{
		{"Klingon", "Creature — Klingon Scientist", "{2}{R}", 2, true},
		{"Warrior", "Creature — Human Warrior", "{2}{R}", 2, true},
		{"Klingon Warrior", "Creature — Klingon Warrior", "{3}{R}", 2, false}, // {1} less once, not twice
		{"Klingon Warrior", "Creature — Klingon Warrior", "{3}{R}", 3, true},
		{"Human", "Creature — Human Soldier", "{2}{R}", 2, false},
	}
	for _, tc := range cases {
		g := newCatalogGame(t)
		me, _ := trekB37Seats(g)
		b27Push(g, me.ID, "Gagh Slinger", "Creature — Klingon Warrior", trekB37GaghOracle, "{2}{R}", 3, 2, "R")
		_, err := trekB37CastFromHand(t, g, game.Card{Name: tc.name, TypeLine: tc.typeLine, ManaCost: tc.cost, Power: 1, Toughness: 1}, tc.lands)
		if (err == nil) != tc.ok {
			t.Errorf("%s %s with %d Mountains: cast error %v, want ok=%v", tc.typeLine, tc.cost, tc.lands, err, tc.ok)
		}
	}
}

func TestTrekB37GaghSlingerDoesNotDiscountAnOpponentsKlingon(t *testing.T) {
	g := newCatalogGame(t)
	_, opps := trekB37Seats(g)
	// The Slinger is an opponent's: the active seat's Klingon is full price.
	b27Push(g, opps[0].ID, "Gagh Slinger", "Creature — Klingon Warrior", trekB37GaghOracle, "{2}{R}", 3, 2, "R")
	if _, err := trekB37CastFromHand(t, g, game.Card{Name: "Klingon", TypeLine: "Creature — Klingon", ManaCost: "{2}{R}", Power: 1, Toughness: 1}, 2); err == nil {
		t.Error("an opponent's Gagh Slinger discounted your Klingon spell")
	}
}

func TestTrekB37GaghSlingerPingsWhenAnotherKlingonOrWarriorEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := trekB37Seats(g)
	b27Push(g, me.ID, "Gagh Slinger", "Creature — Klingon Warrior", trekB37GaghOracle, "{2}{R}", 3, 2, "R")
	advanceToMain(t, g)

	before := trekB37Lives(g)
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, TokenCard("2/1 red Klingon Warrior with haste"), 1)
	})
	passPriorityAroundTable(t, g)
	trekB37AssertOpponentsLost(t, g, me.ID, before, 1) // one Klingon Warrior: one trigger

	before = trekB37Lives(g)
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 black Ninja"), 1)
	})
	passPriorityAroundTable(t, g)
	trekB37AssertOpponentsLost(t, g, me.ID, before, 0)
}

func TestTrekB37GaghSlingerIgnoresItselfAndOpponentsKlingons(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := trekB37Seats(g)
	before := trekB37Lives(g)
	castAndResolveCreature(t, g, "Gagh Slinger", "Creature — Klingon Warrior", trekB37GaghOracle)
	passPriorityAroundTable(t, g)
	trekB37AssertOpponentsLost(t, g, me.ID, before, 0) // "another"

	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(opps[0].ID, TokenCard("2/1 red Klingon Warrior with haste"), 1)
	})
	passPriorityAroundTable(t, g)
	trekB37AssertOpponentsLost(t, g, me.ID, before, 0) // not one you control
}

// --- Galvanizing Tenor --------------------------------------------------------------

func TestTrekB37GalvanizingTenorGoadsUpToOnePerOpponentAttacked(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := trekB37Seats(g)
	tenor := b27Push(g, me.ID, "Galvanizing Tenor", "Creature — Klingon Performer", trekB37TenorOracle, "{3}{R}", 4, 3, "R")
	if !hasEffectiveKeyword(t, g, tenor, "menace") {
		t.Error("Galvanizing Tenor has no menace")
	}
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	a := b12Creature(g, opps[2].ID, "Foe A", "Creature — Ogre", 3, 3)
	b := b12Creature(g, opps[2].ID, "Foe B", "Creature — Ogre", 3, 3)
	c := b12Creature(g, opps[2].ID, "Foe C", "Creature — Ogre", 3, 3)
	trekB37Declare(t, g, map[uuid.UUID]uuid.UUID{tenor: opps[0].ID, bear: opps[1].ID})

	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the battlefront trigger")
	}
	three := []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}, {Kind: game.TargetCard, ID: c}}
	if err := g.ResolvePickTargets(pick.ID, me.ID, three); err == nil {
		t.Fatal("three targets accepted after attacking two opponents")
	}
	answerTriggerTargets(t, g, me.ID, a, b)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if by := trekB37Goaders(t, g, id); !slices.Contains(by, me.ID) {
			t.Errorf("%s goaded by %v, want the Tenor's controller", id, by)
		}
	}
	if by := trekB37Goaders(t, g, c); len(by) != 0 {
		t.Errorf("the untargeted creature is goaded by %v", by)
	}
}

func TestTrekB37GalvanizingTenorAttackingOneOpponentAllowsOneTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := trekB37Seats(g)
	tenor := b27Push(g, me.ID, "Galvanizing Tenor", "Creature — Klingon Performer", trekB37TenorOracle, "{3}{R}", 4, 3, "R")
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	a := b12Creature(g, opps[1].ID, "Foe A", "Creature — Ogre", 3, 3)
	b := b12Creature(g, opps[1].ID, "Foe B", "Creature — Ogre", 3, 3)
	trekB37Declare(t, g, map[uuid.UUID]uuid.UUID{tenor: opps[0].ID, bear: opps[0].ID})
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the battlefront trigger")
	}
	two := []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}
	if err := g.ResolvePickTargets(pick.ID, me.ID, two); err == nil {
		t.Fatal("two targets accepted after attacking one opponent")
	}
	answerTriggerTargets(t, g, me.ID, a)
	passPriorityAroundTable(t, g)
	if by := trekB37Goaders(t, g, a); !slices.Contains(by, me.ID) {
		t.Errorf("the target is goaded by %v", by)
	}
}

// --- Glory to the Empire! ------------------------------------------------------------

func TestTrekB37GloryToTheEmpireMakesAWarriorPerOpponentAndGrantsMelee(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := trekB37Seats(g)
	opps[2].Eliminated = true // two opponents left
	theirs := pushVanillaCreature(g, opps[0].ID, "Their Bear", 2, 2)
	castCatalogSpell(t, g, "Glory to the Empire!", "Enchantment", trekB37GloryOracle, nil)
	passPriorityAroundTable(t, g)
	warriors := trekB37Tokens(g, me.ID, "Klingon Warrior")
	if len(warriors) != 2 {
		t.Fatalf("%d Klingon Warriors, want 2 (one per opponent still in the game)", len(warriors))
	}
	w := warriors[0]
	if w.Power != 2 || w.Toughness != 1 || !slices.Equal(w.Colors, []string{"R"}) || !w.HasSubtype("Klingon") || !w.HasSubtype("Warrior") {
		t.Errorf("token = %s %d/%d %v", w.TypeLine, w.Power, w.Toughness, w.Colors)
	}
	for _, kw := range []string{"haste", game.KeywordMelee} {
		if !slices.Contains(effectiveAbilities(t, g, w.InstanceID), kw) {
			t.Errorf("the Klingon Warrior has no %s", kw)
		}
	}
	if slices.Contains(effectiveAbilities(t, g, theirs), game.KeywordMelee) {
		t.Error("an opponent's creature has melee")
	}

	// Melee: the two hasty Warriors attack the two opponents, each gets +2/+2.
	trekB37Declare(t, g, map[uuid.UUID]uuid.UUID{warriors[0].InstanceID: opps[0].ID, warriors[1].InstanceID: opps[1].ID})
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{warriors[0].InstanceID, warriors[1].InstanceID} {
		if c := layeredCard(t, g, id).Effective(); c.Power != 4 || c.Toughness != 3 {
			t.Errorf("melee Warrior is %d/%d, want 4/3 after attacking two opponents", c.Power, c.Toughness)
		}
	}
}

// --- Gnarling Targ -------------------------------------------------------------------

func TestTrekB37GnarlingTargPumpsKlingonsAndWarriorsOncePerAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := trekB37Seats(g)
	targ := b27Push(g, me.ID, "Gnarling Targ", "Creature — Boar", trekB37TargOracle, "{4}{R}", 4, 5, "R")
	if !hasEffectiveKeyword(t, g, targ, "trample") {
		t.Error("Gnarling Targ has no trample")
	}
	klingon := b12Creature(g, me.ID, "Klingon", "Creature — Klingon Scientist", 1, 1)
	warrior := b12Creature(g, me.ID, "Warrior", "Creature — Human Warrior", 2, 2)
	human := pushVanillaCreature(g, me.ID, "Human", 2, 2)
	theirs := b12Creature(g, opps[0].ID, "Their Klingon", "Creature — Klingon", 2, 2)
	// Two attackers, neither a Klingon: one trigger, and the stay-at-home Klingon is pumped.
	trekB37Declare(t, g, map[uuid.UUID]uuid.UUID{human: opps[0].ID, warrior: opps[1].ID})
	passPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{klingon: 3, warrior: 4, human: 2, targ: 4, theirs: 2} {
		if got := layeredCard(t, g, id).Effective().Power; got != want {
			t.Errorf("%s power %d, want %d", layeredCard(t, g, id).Name, got, want)
		}
	}
	if !hasEffectiveKeyword(t, g, klingon, "trample") || !hasEffectiveKeyword(t, g, warrior, "trample") {
		t.Error("a Klingon or Warrior did not gain trample")
	}
	if hasEffectiveKeyword(t, g, human, "trample") {
		t.Error("the Human gained trample")
	}
	advanceTo(t, g, game.StepCleanup)
	if got := layeredCard(t, g, klingon).Effective().Power; got != 1 {
		t.Errorf("the pump outlived the turn: power %d", got)
	}
}

// --- Grilka, Widowed Warrior -----------------------------------------------------------

func TestTrekB37GrilkaStealsUntapsAndMakesTappedTreasuresOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := trekB37Seats(g)
	ogre := b12Creature(g, opps[0].ID, "Ogre", "Creature — Ogre", 3, 3)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(ogre) })
	castAndResolveCreature(t, g, "Grilka, Widowed Warrior", "Legendary Creature — Klingon Warrior", trekB37GrilkaOracle)
	pickCard(t, g, me.ID, ogre)
	passPriorityAroundTable(t, g)

	c := layeredCard(t, g, ogre)
	if c.Controller != me.ID || c.Tapped {
		t.Fatalf("the Ogre is controlled by %s, tapped %v: want yours and untapped", c.Controller, c.Tapped)
	}
	for _, kw := range []string{"trample", "haste"} {
		if !hasEffectiveKeyword(t, g, ogre, kw) {
			t.Errorf("the stolen creature has no %s", kw)
		}
	}
	trekB37Declare(t, g, map[uuid.UUID]uuid.UUID{ogre: opps[1].ID})
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepEndCombat)
	treasures := trekB37Tokens(g, me.ID, "Treasure")
	if len(treasures) != 3 {
		t.Fatalf("%d Treasures, want 3 (the 3 damage it dealt)", len(treasures))
	}
	for _, tr := range treasures {
		if !tr.Tapped {
			t.Error("a Treasure entered untapped")
		}
	}
	advanceTo(t, g, game.StepUpkeep) // the next turn
	if c := layeredCard(t, g, ogre); c.Controller != opps[0].ID {
		t.Errorf("the Ogre is still controlled by %s after the turn", c.Controller)
	}
}

func TestTrekB37GrilkaCantTakeYourOwnCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := trekB37Seats(g)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Ogre", 3, 3)
	b12Creature(g, opps[0].ID, "Theirs", "Creature — Ogre", 3, 3)
	castAndResolveCreature(t, g, "Grilka, Widowed Warrior", "Legendary Creature — Klingon Warrior", trekB37GrilkaOracle)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for Grilka's enters trigger")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
		t.Error("Grilka targeted a creature you control")
	}
}

// --- Kang, Suspicious Captain -----------------------------------------------------------

func TestTrekB37KangGoadsHimselfAndTheTargetWhileYouControlHim(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := trekB37Seats(g)
	ogre := b12Creature(g, opps[0].ID, "Ogre", "Creature — Ogre", 3, 3)
	kang := castAndResolveCreature(t, g, "Kang, Suspicious Captain", "Legendary Creature — Klingon Warrior Officer", trekB37KangOracle)
	pickCard(t, g, me.ID, ogre)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{kang, ogre} {
		if by := trekB37Goaders(t, g, id); !slices.Equal(by, []uuid.UUID{me.ID}) {
			t.Errorf("%s goaded by %v, want only Kang's controller", id, by)
		}
	}
	// No "until your next turn" end: it lasts while you control Kang.
	b15Destroy(t, g, kang)
	if by := trekB37Goaders(t, g, ogre); len(by) != 0 {
		t.Errorf("the Ogre is still goaded by %v after Kang left", by)
	}
}

func TestTrekB37KangCountersWhenGoadedCreaturesAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := trekB37Seats(g)
	kang := b27Push(g, me.ID, "Kang, Suspicious Captain", "Legendary Creature — Klingon Warrior Officer", trekB37KangOracle, "{3}{R}", 4, 4, "R")
	a := pushVanillaCreature(g, me.ID, "Goaded A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Goaded B", 2, 2)
	g.WithWriteLock(func() {
		g.GoadForEffect(a, opps[2].ID)
		g.GoadForEffect(b, opps[2].ID)
	})
	trekB37Declare(t, g, map[uuid.UUID]uuid.UUID{a: opps[0].ID, b: opps[0].ID})
	answerTriggerTargets(t, g, me.ID, kang)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Fatal("a second trigger for the second goaded attacker")
	}
	if got := dnaPower(t, g, kang); got != 5 {
		t.Errorf("Kang's power %d, want 5 (one +1/+1 counter)", got)
	}
}

func TestTrekB37KangIgnoresAnAttackWithNoGoadedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := trekB37Seats(g)
	b27Push(g, me.ID, "Kang, Suspicious Captain", "Legendary Creature — Klingon Warrior Officer", trekB37KangOracle, "{3}{R}", 4, 4, "R")
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	trekB37Declare(t, g, map[uuid.UUID]uuid.UUID{bear: opps[0].ID})
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Error("Kang triggered for an attack with no goaded creature")
	}
}

// --- Koloth, Tribble Hunter --------------------------------------------------------------

func TestTrekB37KolothGivesEveryoneTwoGoadedTribbles(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := trekB37Seats(g)
	koloth := b27Push(g, me.ID, "Koloth, Tribble Hunter", "Legendary Creature — Klingon Warrior Officer", trekB37KolothOracle, "{3}{R}", 3, 3, "R")
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	for _, p := range append([]*game.Player{me}, opps...) {
		tribbles := trekB37Tokens(g, p.ID, "Tribble")
		if len(tribbles) != 2 {
			t.Fatalf("seat %d has %d Tribbles, want 2", p.Seat, len(tribbles))
		}
		c := layeredCard(t, g, tribbles[0].InstanceID)
		if e := c.Effective(); e.Power != 2 || e.Toughness != 1 || !slices.Equal(e.Colors, []string{"W"}) || !c.HasSubtype("Tribble") {
			t.Errorf("seat %d's Tribble is %s %d/%d %v, want a 2/1 white Tribble", p.Seat, c.TypeLine, e.Power, e.Toughness, e.Colors)
		}
		if by := trekB37Goaders(t, g, c.InstanceID); !slices.Equal(by, []uuid.UUID{me.ID}) {
			t.Errorf("seat %d's Tribble is goaded by %v, want Koloth's controller", p.Seat, by)
		}
	}
	// The goad and the pump are Koloth's: they end when he leaves.
	b15Destroy(t, g, koloth)
	tribble := trekB37Tokens(g, opps[0].ID, "Tribble")[0]
	if c := layeredCard(t, g, tribble.InstanceID).Effective(); c.Power != 1 {
		t.Errorf("Tribble power %d after Koloth left, want 1", c.Power)
	}
	if by := trekB37Goaders(t, g, tribble.InstanceID); len(by) != 0 {
		t.Errorf("Tribble still goaded by %v after Koloth left", by)
	}
}

func TestTrekB37KolothDoesNothingAtAnOpponentsEndStep(t *testing.T) {
	g := newCatalogGame(t)
	_, opps := trekB37Seats(g)
	b27Push(g, opps[0].ID, "Koloth, Tribble Hunter", "Legendary Creature — Klingon Warrior Officer", trekB37KolothOracle, "{3}{R}", 3, 3, "R")
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if n := len(trekB37Tokens(g, p.ID, "Tribble")); n != 0 {
			t.Errorf("seat %d has %d Tribbles after an opponent's end step", p.Seat, n)
		}
	}
}

func TestTrekB37KolothTribblesCantBlock(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := trekB37Seats(g)
	b27Push(g, me.ID, "Koloth, Tribble Hunter", "Legendary Creature — Klingon Warrior Officer", trekB37KolothOracle, "{3}{R}", 3, 3, "R")
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opps[0].ID, TribbleToken(), 1) })
	tribble := trekB37Tokens(g, opps[0].ID, "Tribble")[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	trekB37Declare(t, g, map[uuid.UUID]uuid.UUID{bear: opps[0].ID})
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(tribble.InstanceID, bear); err == nil {
		t.Error("a Tribble blocked")
	}
}

// --- Own the Day -----------------------------------------------------------------------

func trekB37OwnTheDay() game.Card {
	return game.Card{Name: "Own the Day", TypeLine: "Sorcery", ManaCost: "{4}{R}{R}", OracleID: trekB37OwnTheDayOracle, Colors: []string{"R"}}
}

func TestTrekB37OwnTheDayCostsLessPerOpponentAttackedAndUntapsAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := trekB37Seats(g)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	c := pushVanillaCreature(g, me.ID, "Bear C", 2, 2)
	trekB37Declare(t, g, map[uuid.UUID]uuid.UUID{a: opps[0].ID, b: opps[1].ID, c: opps[1].ID})
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepPostcombatMain)
	// Two opponents attacked: {4}{R}{R} costs {2}{R}{R}.
	if _, err := trekB37CastFromHand(t, g, trekB37OwnTheDay(), 3); err == nil {
		t.Fatal("cast for three mana after attacking two opponents")
	}
	if _, err := trekB37CastFromHand(t, g, trekB37OwnTheDay(), 1); err != nil {
		t.Fatalf("cast for four mana after attacking two opponents: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if isTapped(g, id) {
			t.Errorf("%s attacked this turn and is still tapped", id)
		}
	}
	if next := stepOnce(t, g); next.Step != game.StepBeginCombat {
		t.Errorf("after Own the Day's main phase: %+v, want an additional combat", next)
	}
}

func TestTrekB37OwnTheDayIsFullPriceWithoutAnAttack(t *testing.T) {
	g := newCatalogGame(t)
	if _, err := trekB37CastFromHand(t, g, trekB37OwnTheDay(), 5); err == nil {
		t.Error("Own the Day cast for five mana without attacking anyone")
	}
}

// --- Qo'noS Quartermaster ------------------------------------------------------------------

func TestTrekB37QuartermasterAddsAnyColorManaThatLastsTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := trekB37Seats(g)
	qm := b27Push(g, me.ID, "Qo'noS Quartermaster", "Creature — Klingon Warrior", trekB37QuartermasterOracle, "{1}{R}", 2, 2, "R")
	if !hasEffectiveKeyword(t, g, qm, "menace") {
		t.Error("Qo'noS Quartermaster has no menace")
	}
	trekB37Declare(t, g, map[uuid.UUID]uuid.UUID{qm: opps[0].ID})
	passPriorityAroundTable(t, g)
	if b10ResolveAllManaPicks(t, g, me.ID, "B") != 1 {
		t.Fatal("the attack trigger raised no colour pick")
	}
	advanceTo(t, g, game.StepPostcombatMain)
	if got := len(me.ManaPool); got != 1 {
		t.Fatalf("mana pool = %d in the second main phase, want the kept mana", got)
	}
}

// --- Ro Laren, Double Agent ----------------------------------------------------------------

func TestTrekB37RoLarenPingsOnTheSecondDrawOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := trekB37Seats(g)
	b27Push(g, me.ID, "Ro Laren, Double Agent", "Legendary Creature — Bajoran Rogue Officer", trekB37RoLarenOracle, "{2}{R}", 3, 2, "R")
	advanceToMain(t, g) // the draw step's card was the first
	before := trekB37Lives(g)
	g.WithWriteLock(func() { _ = DrawCards{Player: me.ID, N: 1}.Apply(NewContext(g, &game.StackItem{Controller: me.ID})) })
	passPriorityAroundTable(t, g)
	trekB37AssertOpponentsLost(t, g, me.ID, before, 2)
	before = trekB37Lives(g)
	g.WithWriteLock(func() { _ = DrawCards{Player: me.ID, N: 1}.Apply(NewContext(g, &game.StackItem{Controller: me.ID})) })
	passPriorityAroundTable(t, g)
	trekB37AssertOpponentsLost(t, g, me.ID, before, 0)
}

func TestTrekB37RoLarenLootsForOneAndAPinch(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := trekB37Seats(g)
	ro := b27Push(g, me.ID, "Ro Laren, Double Agent", "Legendary Creature — Bajoran Rogue Officer", trekB37RoLarenOracle, "{2}{R}", 3, 2, "R")
	advanceToMain(t, g)
	discard := me.Hand.Cards[0].InstanceID
	fillPool(me, 1)
	hand := me.Hand.Size()
	before := trekB37Lives(g)
	if err := g.ActivateCatalogAbility(me.ID, ro, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{discard}}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(discard) {
		t.Error("the discarded card is not in the graveyard")
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d → %d, want the same size (discard one, draw one)", hand, got)
	}
	// The loot's draw is the turn's second card: Seek out triggers.
	trekB37AssertOpponentsLost(t, g, me.ID, before, 2)
	// {T} is part of the cost: a tapped Ro Laren can't loot again.
	fillPool(me, 1)
	again := me.Hand.Cards[0].InstanceID
	if err := g.ActivateCatalogAbility(me.ID, ro, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{again}}); err == nil {
		t.Error("a tapped Ro Laren activated her {T} ability")
	}
}
