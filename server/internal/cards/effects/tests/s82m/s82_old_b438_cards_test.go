package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b438_cards_test.go — Avacyn Restored, batch 438 (S82, #5570).
// Each test plays the real catalog entry through the engine.

const (
	b438Descent   = "88d42e0e-3921-4873-a53b-146220341f76"
	b438Atlas     = "8d907a1b-f301-46a3-aa1c-25fb7f29c949"
	b438Repo      = "cbbc530f-e51b-423e-b219-23bb580fd834"
	b438Second    = "b036eac6-60db-48ca-b6e1-b42afac959bd"
	b438Snatch    = "b5fcfcf5-8af5-4395-b989-427fa92e0dd2"
	b438Outwit    = "eac02e36-15c8-4b6d-b9b4-da549a722336"
	b438Yew       = "0333cb81-5f44-4f31-9ee7-725e4f6af7b3"
	b438Justiciar = "9edcf49b-2767-48bd-8fbe-968527cbddf4"
	b438Gallows   = "0d80841e-7a33-4b7f-b061-fa2f87f596de"
	b438Rising    = "10064a1d-66f4-4018-94c2-c727e7838850"
	b438Bracers   = "24d373cd-245a-4051-b187-ba8cc83b0c48"
	b438Trident   = "b74992b8-2857-403b-88cf-f2718873571a"
)

func TestB438EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b438Descent: "Descent into Madness", b438Atlas: "Otherworld Atlas", b438Repo: "Druids' Repository",
		b438Second: "Second Guess", b438Snatch: "Geist Snatch", b438Outwit: "Outwit", b438Yew: "Yew Spirit",
		b438Justiciar: "Holy Justiciar", b438Gallows: "Gallows at Willow Hill", b438Rising: "Demonic Rising",
		b438Bracers: "Bladed Bracers", b438Trident: "Tormentor's Trident",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func b438Effective(g *game.Game, id uuid.UUID) (c game.Characteristic) {
	g.ReadSnapshot(func() {
		if card := effectstest.FindBattlefieldCardForTest(g, id); card != nil {
			c = card.Effective()
		}
	})
	return c
}

func b438Seats(g *game.Game) (*game.Player, *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// b438Cast casts a card strictly from me's hand, paying nothing when cost is "".
func b438Cast(t *testing.T, g *game.Game, me *game.Player, name, typeLine, cost, oracle string, targets []game.TargetRef) (uuid.UUID, error) {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle, Owner: me.ID, Controller: me.ID})
	})
	return id, g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: targets})
}

func b438Spirits(g *game.Game, owner uuid.UUID, name string) int {
	return effectstest.CountBattlefieldNamed(g, owner, name)
}

func TestB438SecondGuessCountersOnlyTheSecondSpellCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b438Seats(g)
	effectstest.AdvanceToMain(t, g)
	first, err := b438Cast(t, g, me, "First Spell", "Instant", "", "", effectstest.PlayerRefs(opp.ID))
	if err != nil {
		t.Fatal(err)
	}
	second, err := b438Cast(t, g, me, "Second Spell", "Instant", "", "", effectstest.PlayerRefs(opp.ID))
	if err != nil {
		t.Fatal(err)
	}
	effectstest.ApaMana(me, "U", "C", "U", "C")
	if _, err := b438Cast(t, g, me, "Second Guess", "Instant", "{1}{U}", b438Second, effectstest.CardTarget(first)); err == nil {
		t.Fatal("Second Guess targeted the first spell cast this turn")
	}
	if _, err := b438Cast(t, g, me, "Second Guess", "Instant", "{1}{U}", b438Second, effectstest.CardTarget(second)); err != nil {
		t.Fatalf("Second Guess on the second spell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.P7Countered(g, second) {
		t.Error("the second spell was not countered")
	}
	if effectstest.P7Countered(g, first) {
		t.Error("the first spell was countered")
	}
}

func TestB438OutwitCountersOnlyASpellThatTargetsAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b438Seats(g)
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	atBear, err := b438Cast(t, g, me, "Shock A", "Instant", "", "", effectstest.CardTarget(bear))
	if err != nil {
		t.Fatal(err)
	}
	atFace, err := b438Cast(t, g, me, "Shock B", "Instant", "", "", effectstest.PlayerRefs(opp.ID))
	if err != nil {
		t.Fatal(err)
	}
	effectstest.ApaMana(me, "U", "U")
	if _, err := b438Cast(t, g, me, "Outwit", "Instant", "{U}", b438Outwit, effectstest.CardTarget(atBear)); err == nil {
		t.Fatal("Outwit targeted a spell that targets only a creature")
	}
	if _, err := b438Cast(t, g, me, "Outwit", "Instant", "{U}", b438Outwit, effectstest.CardTarget(atFace)); err != nil {
		t.Fatalf("Outwit on a spell that targets a player: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.P7Countered(g, atFace) {
		t.Error("the spell that targeted a player was not countered")
	}
	if effectstest.P7Countered(g, atBear) {
		t.Error("the spell that targeted a creature was countered")
	}
}

func TestB438GeistSnatchCountersACreatureSpellAndMakesASpirit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b438Seats(g)
	effectstest.AdvanceToMain(t, g)
	burn, err := b438Cast(t, g, me, "Burn", "Instant", "", "", effectstest.PlayerRefs(opp.ID))
	if err != nil {
		t.Fatal(err)
	}
	effectstest.ApaMana(me, "U", "U", "C", "C")
	if _, err := b438Cast(t, g, me, "Geist Snatch", "Instant", "{2}{U}{U}", b438Snatch, effectstest.CardTarget(burn)); err == nil {
		t.Fatal("Geist Snatch targeted a noncreature spell")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2, _ := b438Seats(g2)
	effectstest.AdvanceToMain(t, g2)
	bearSpell, err := b438Cast(t, g2, me2, "Bear Cub", "Creature — Bear", "", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	effectstest.ApaMana(me2, "U", "U", "C", "C")
	if _, err := b438Cast(t, g2, me2, "Geist Snatch", "Instant", "{2}{U}{U}", b438Snatch, effectstest.CardTarget(bearSpell)); err != nil {
		t.Fatalf("Geist Snatch: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if !effectstest.P7Countered(g2, bearSpell) {
		t.Error("the creature spell was not countered")
	}
	if n := b438Spirits(g2, me2.ID, "Spirit"); n != 1 {
		t.Errorf("Spirit tokens %d, want 1 (Geist Snatch's controller gets it)", n)
	}
}

func TestB438YewSpiritDoublesItsPowerEachTime(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := b438Seats(g)
	effectstest.AdvanceToMain(t, g)
	yew := effectstest.B12Push(g, me.ID, "Yew Spirit", "Creature — Spirit Treefolk", b438Yew, 3, 3)
	effectstest.RavB2MustActivate(t, g, me.ID, yew, 0, game.ActivateAbilityParams{}, "C", "C", "G", "G")
	if p, tough := effectstest.PtOf(t, g, yew); p != 6 || tough != 6 {
		t.Fatalf("after one activation %d/%d, want 6/6", p, tough)
	}
	effectstest.RavB2MustActivate(t, g, me.ID, yew, 0, game.ActivateAbilityParams{}, "C", "C", "G", "G")
	if p, tough := effectstest.PtOf(t, g, yew); p != 12 || tough != 12 {
		t.Errorf("after two activations %d/%d, want 12/12 (X is its power as the ability resolves)", p, tough)
	}
}

func TestB438HolyJusticiarTapsAndExilesAZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b438Seats(g)
	effectstest.AdvanceToMain(t, g)
	j := effectstest.B12Push(g, me.ID, "Holy Justiciar", "Creature — Human Cleric", b438Justiciar, 2, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	zombie := effectstest.B12Creature(g, opp.ID, "Walker", "Creature — Zombie", 2, 2)
	effectstest.RavB2MustActivate(t, g, me.ID, j, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)}, "C", "C", "W")
	if !effectstest.TappedOnBattlefield(t, g, bear) || effectstest.FindBattlefieldCardForTest(g, bear) == nil {
		t.Error("a non-Zombie should be tapped and stay")
	}
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, j).Tapped = false })
	effectstest.RavB2MustActivate(t, g, me.ID, j, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(zombie)}, "C", "C", "W")
	if !effectstest.InExile(g, zombie) {
		t.Error("the Zombie was not exiled")
	}
}

func TestB438GallowsDestroysAndGivesTheVictimsControllerASpirit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b438Seats(g)
	effectstest.AdvanceToMain(t, g)
	gal := effectstest.B12Push(g, me.ID, "Gallows at Willow Hill", "Artifact", b438Gallows, 0, 0)
	h1 := effectstest.B12Creature(g, me.ID, "Villager 1", "Creature — Human", 1, 1)
	h2 := effectstest.B12Creature(g, me.ID, "Villager 2", "Creature — Human", 1, 1)
	h3 := effectstest.B12Creature(g, me.ID, "Villager 3", "Creature — Human", 1, 1)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	effectstest.ApaMana(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, gal, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(ogre), TapIDs: []uuid.UUID{h1, h2}}); err == nil {
		t.Fatal("paid with only two Humans")
	}
	if err := g.ActivateCatalogAbility(me.ID, gal, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(ogre), TapIDs: []uuid.UUID{h1, h2, h3}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if effectstest.FindBattlefieldCardForTest(g, ogre) != nil {
		t.Error("the Ogre survived")
	}
	if n := b438Spirits(g, opp.ID, "Spirit"); n != 1 {
		t.Errorf("the Ogre's controller has %d Spirit tokens, want 1", n)
	}
	if n := b438Spirits(g, me.ID, "Spirit"); n != 0 {
		t.Errorf("the Gallows' controller got %d Spirit tokens, want 0", n)
	}
}

func TestB438DemonicRisingNeedsExactlyOneCreature(t *testing.T) {
	for _, extra := range []int{0, 1} {
		g := effectstest.NewCatalogGame(t)
		me, _ := b438Seats(g)
		effectstest.AdvanceToMain(t, g)
		effectstest.B12Push(g, me.ID, "Demonic Rising", "Enchantment", b438Rising, 0, 0)
		effectstest.B12Creature(g, me.ID, "Solo", "Creature — Bear", 2, 2)
		for i := 0; i < extra; i++ {
			effectstest.B12Creature(g, me.ID, "Friend", "Creature — Bear", 2, 2)
		}
		effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
		effectstest.PassPriorityAroundTable(t, g)
		want := 1 - extra
		if got := b438Spirits(g, me.ID, "Demon"); got != want {
			t.Errorf("with %d creatures: %d Demons, want %d", 1+extra, got, want)
		}
	}
}

func TestB438BladedBracersGivesVigilanceToHumansAndAngelsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := b438Seats(g)
	effectstest.AdvanceToMain(t, g)
	br := effectstest.B12Push(g, me.ID, "Bladed Bracers", "Artifact — Equipment", b438Bracers, 0, 0)
	human := effectstest.B12Creature(g, me.ID, "Villager", "Creature — Human", 1, 1)
	angel := effectstest.B12Creature(g, me.ID, "Seraph", "Creature — Angel", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for _, c := range []struct {
		id       uuid.UUID
		vigilant bool
		p        int
	}{{human, true, 2}, {angel, true, 3}, {bear, false, 3}} {
		if err := effectstest.StrictEquip(t, g, me.ID, br, c.id, "{C}{C}"); err != nil {
			t.Fatalf("equip: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if p, tough := effectstest.PtOf(t, g, c.id); p != c.p || tough != c.p {
			t.Errorf("equipped %d/%d, want %d/%d", p, tough, c.p, c.p)
		}
		if got := effectstest.HasEffectiveKeyword(t, g, c.id, "vigilance"); got != c.vigilant {
			t.Errorf("vigilance = %v, want %v", got, c.vigilant)
		}
	}
}

func TestB438TormentorsTridentPumpsAndForcesAnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := b438Seats(g)
	effectstest.AdvanceToMain(t, g)
	tr := effectstest.B12Push(g, me.ID, "Tormentor's Trident", "Artifact — Equipment", b438Trident, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.StrictEquip(t, g, me.ID, tr, bear, "{C}{C}{C}"); err != nil {
		t.Fatalf("equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 5 || tough != 2 {
		t.Errorf("equipped Bear %d/%d, want 5/2", p, tough)
	}
	if n := len(b438Effective(g, bear).AttackRequirements); n != 1 {
		t.Errorf("attack requirements %d, want 1 (attacks each combat if able)", n)
	}
}

func TestB438DruidsRepositoryCountsEachAttackerAndSpendsCountersForMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b438Seats(g)
	effectstest.AdvanceToMain(t, g)
	repo := effectstest.B12Push(g, me.ID, "Druids' Repository", "Enchantment", b438Repo, 0, 0)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{a, b} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, repo, game.CounterCharge); got != 2 {
		t.Fatalf("charge counters %d, want 2 (one per attacking creature)", got)
	}
	if err := g.ActivateManaAbility(me.ID, repo, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if effectstest.LatestChoiceOfKind(g, game.PendingChoiceMana) != nil {
		effectstest.ResolveManaPickAnyColor(t, g, me.ID)
	}
	if got := effectstest.CounterCount(g, repo, game.CounterCharge); got != 1 {
		t.Errorf("charge counters %d, want 1 after spending one", got)
	}
	if me.ManaPool.Total() != 1 {
		t.Errorf("pool %v, want one mana", effectstest.PoolColors(me))
	}
}

func TestB438OtherworldAtlasDrawsForEveryCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := b438Seats(g)
	effectstest.AdvanceToMain(t, g)
	atlas := effectstest.B12Push(g, me.ID, "Otherworld Atlas", "Artifact — Book", b438Atlas, 0, 0)
	untap := func() { g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, atlas).Tapped = false }) }
	for i := 0; i < 2; i++ {
		effectstest.RavB2MustActivate(t, g, me.ID, atlas, 0, game.ActivateAbilityParams{})
		untap()
	}
	if got := effectstest.CounterCount(g, atlas, game.CounterCharge); got != 2 {
		t.Fatalf("charge counters %d, want 2", got)
	}
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = len(p.Hand.Cards)
	}
	effectstest.RavB2MustActivate(t, g, me.ID, atlas, 1, game.ActivateAbilityParams{})
	for _, p := range g.Seats {
		if got := len(p.Hand.Cards) - before[p.ID]; got != 2 {
			t.Errorf("%s drew %d cards, want 2 (one per charge counter)", p.Name, got)
		}
	}
}

func TestB438DescentIntoMadnessGrowsEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := b438Seats(g)
	effectstest.AdvanceToMain(t, g)
	d := effectstest.B12Push(g, me.ID, "Descent into Madness", "Enchantment", b438Descent, 0, 0)
	seat := g.Turn.ActiveSeat
	for _, p := range g.Seats {
		g.WithWriteLock(func() { p.Hand.Cards = nil })
	}
	stones := []uuid.UUID{}
	for i := 0; i < 3; i++ {
		stones = append(stones, effectstest.B12Push(g, me.ID, "Stone", "Artifact", "", 0, 0))
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, seat)
	if g.Turn.ActiveSeat == seat {
		effectstest.PassPriorityUntilChoice(t, g)
	}
	if got := effectstest.CounterCount(g, d, "despair"); got != 1 {
		t.Fatalf("despair counters %d, want 1", got)
	}
	prompt := effectstest.LatestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("the controller was not asked to exile")
	}
	if prompt.ChooseMin != 1 || prompt.ChooseMax != 1 {
		t.Errorf("X = %d..%d, want exactly 1 on the first upkeep", prompt.ChooseMin, prompt.ChooseMax)
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{stones[0]}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	for _, p := range g.Seats {
		if p.ID == me.ID {
			continue
		}
		if pr := effectstest.LatestChooseCards(g, p.ID); pr != nil {
			if err := g.ResolveChooseCards(pr.ID, p.ID, pr.ChooseCards[:pr.ChooseMin]); err != nil {
				t.Fatalf("%s picks: %v", p.Name, err)
			}
		}
	}
	if !effectstest.InExile(g, stones[0]) || effectstest.InExile(g, stones[1]) {
		t.Error("exactly the chosen permanent should be exiled")
	}
}
