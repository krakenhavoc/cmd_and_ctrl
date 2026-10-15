package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// yshtola_deck_sweep_b_test.go — the second slice of the "Y'shtola,
// Night's Blessed" deck request (#3657): Alisaie and Alphinaud
// Leveilleur, the job-select Equipment (Reaper's Scythe, Sage's
// Nouliths, Astrologian's Planisphere), Champions from Beyond, Eye of
// Nidhogg and Observed Stasis.

const (
	ysBAlisaieOracle    = "4bd1d079-314e-4fb5-9b43-c9c83e5528bc"
	ysBAlphinaudOracle  = "d16783a4-9980-451b-9af0-33866980bc8d"
	ysBScytheOracle     = "b0f7c2a9-914a-48c1-871c-6f72946d61d8"
	ysBNoulithsOracle   = "97248734-6e99-42a2-8f11-f3337b25f8ab"
	ysBPlanisphereOracl = "4dcd0583-cee0-46a6-ba43-93bca5eb0428"
	ysBChampionsOracle  = "57e4a1c0-4218-4a07-88e7-d6083884c2b6"
	ysBEyeOracle        = "c288f184-0beb-4cc9-9b0a-5f3c4b9e70d7"
	ysBStasisOracle     = "0c729806-3a23-491e-847f-23a394d27291"
)

func TestYsBAlisaieDiscountsOnlyTheSecondSpellEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Alisaie Leveilleur", "Legendary Creature — Elf Wizard", ysBAlisaieOracle, false)
	if got := priceInHand(t, g, me, "Cultivate", "Sorcery", "{3}{G}"); got != 4 {
		t.Errorf("first spell: %d, want 4 (undiscounted)", got)
	}
	b24CastNoncreature(t, g, "First", "Instant", "{U}")
	passPriorityAroundTable(t, g)
	if got := priceInHand(t, g, me, "Cultivate", "Sorcery", "{3}{G}"); got != 2 {
		t.Errorf("second spell: %d, want 2 ({2} less)", got)
	}
	b24CastNoncreature(t, g, "Second", "Instant", "{U}")
	passPriorityAroundTable(t, g)
	if got := priceInHand(t, g, me, "Cultivate", "Sorcery", "{3}{G}"); got != 4 {
		t.Errorf("third spell: %d, want 4 (undiscounted)", got)
	}
}

func TestYsBAlphinaudDrawsOnTheSecondSpellOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Alphinaud Leveilleur", "Legendary Creature — Elf Wizard", ysBAlphinaudOracle, false)
	hand := me.Hand.Size()
	b24CastNoncreature(t, g, "First", "Instant", "{U}")
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatalf("the first spell drew: hand %d → %d", hand, me.Hand.Size())
	}
	b24CastNoncreature(t, g, "Second", "Instant", "{U}")
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("the second spell should draw one: hand %d → %d", hand, me.Hand.Size())
	}
	b24CastNoncreature(t, g, "Third", "Instant", "{U}")
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("the third spell drew: hand %d → %d", hand, me.Hand.Size())
	}
}

// ysBHero returns the Hero token a job-select Equipment made and
// whether it is the one the Equipment is attached to.
func ysBHero(t *testing.T, g *game.Game, equipment uuid.UUID) uuid.UUID {
	t.Helper()
	eq, ok := battlefieldCard(g, equipment)
	if !ok {
		t.Fatal("the Equipment is not on the battlefield")
	}
	if eq.AttachedTo.Kind != game.TargetCard {
		t.Fatal("job select did not attach the Equipment to the Hero")
	}
	hero, ok := battlefieldCard(g, eq.AttachedTo.ID)
	if !ok || hero.Name != "Hero" {
		t.Fatalf("attached to %q, want a Hero token", hero.Name)
	}
	return hero.InstanceID
}

func TestYsBReapersScytheGrowsBySoulCountersAtTheEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	scythe := castCatalogSpell(t, g, "Reaper's Scythe", "Artifact — Equipment", ysBScytheOracle, nil)
	passPriorityAroundTable(t, g)
	hero := ysBHero(t, g, scythe)
	v := s58p6View(t, g, hero)
	if v.CurrentPower() != 1 || !v.HasSubtype("Assassin") {
		t.Fatalf("Hero is %d/%d with Assassin=%v before any counters", v.CurrentPower(), v.CurrentToughness(), v.HasSubtype("Assassin"))
	}

	g.WithWriteLock(func() {
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -3)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -1)
	})
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, scythe, "soul"); got != 2 {
		t.Fatalf("soul counters %d, want 2 (two players lost life)", got)
	}
	v = s58p6View(t, g, hero)
	if v.CurrentPower() != 3 || v.CurrentToughness() != 3 {
		t.Errorf("Hero is %d/%d, want 3/3", v.CurrentPower(), v.CurrentToughness())
	}
}

func TestYsBSagesNoulithsMakesAClericHeroWithPlusOne(t *testing.T) {
	g := newCatalogGame(t)
	nouliths := castCatalogSpell(t, g, "Sage's Nouliths", "Artifact — Equipment", ysBNoulithsOracle, nil)
	passPriorityAroundTable(t, g)
	hero := ysBHero(t, g, nouliths)
	v := s58p6View(t, g, hero)
	if v.CurrentPower() != 2 || v.CurrentToughness() != 1 {
		t.Errorf("Hero is %d/%d, want 2/1", v.CurrentPower(), v.CurrentToughness())
	}
	if !v.HasSubtype("Cleric") || !v.HasSubtype("Hero") {
		t.Errorf("Hero subtypes %v, want Hero and Cleric", v.Effective().Subtypes)
	}
}

func TestYsBPlanisphereGrowsTheHeroOnNoncreatureSpellsAndTheThirdDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	planisphere := castCatalogSpell(t, g, "Astrologian's Planisphere", "Artifact — Equipment", ysBPlanisphereOracl, nil)
	passPriorityAroundTable(t, g)
	hero := ysBHero(t, g, planisphere)
	if !s58p6View(t, g, hero).HasSubtype("Wizard") {
		t.Error("the Hero is not a Wizard")
	}
	b24CastNoncreature(t, g, "Spell", "Instant", "{U}")
	passPriorityAroundTable(t, g)
	if got := countersOn(g, hero, game.CounterPlusOne); got != 1 {
		t.Fatalf("after a noncreature spell: %d counters, want 1", got)
	}
	// However many cards are drawn, only the third of the turn counts.
	for i := 0; i < 6; i++ {
		g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
		passPriorityAroundTable(t, g)
	}
	if got := countersOn(g, hero, game.CounterPlusOne); got != 2 {
		t.Errorf("after six draws: %d counters, want 2 (one for the third card)", got)
	}
}

func TestYsBChampionsFromBeyondMakesXHeroesAndPartiesUp(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castXSpell(t, g, "Champions from Beyond", "Enchantment", ysBChampionsOracle, "{X}{W}{W}", 5, nil)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Hero"); n != 5 {
		t.Fatalf("Hero tokens %d, want 5", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Hero" {
			g.WithWriteLock(func() {
				for i := range g.Battlefield.Cards {
					g.Battlefield.Cards[i].SummonedThisTurn = false
				}
			})
			break
		}
	}
	var heroes []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Hero" {
			heroes = append(heroes, c.InstanceID)
		}
	}
	hand := me.Hand.Size()
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range heroes[:4] {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatal(err)
		}
	}
	lockInAttacks(t, g)
	drcSettle(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Fatalf("Light Party drew before the scry was answered: %d", me.Hand.Size()-hand)
	}
	answerScryKeepAll(t, g, me.ID)
	drcSettle(t, g, me.ID)
	if me.Hand.Size() != hand+1 {
		t.Errorf("Light Party: drew %d, want 1", me.Hand.Size()-hand)
	}
	for _, id := range heroes[:4] {
		if s58p6View(t, g, id).CurrentPower() != 1 {
			t.Error("Full Party fired with only four attackers")
		}
	}
}

func TestYsBChampionsFullPartyPumpsEightAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Champions from Beyond", "Enchantment", ysBChampionsOracle, false)
	var bears []uuid.UUID
	for i := 0; i < 8; i++ {
		bears = append(bears, b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range bears {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatal(err)
		}
	}
	lockInAttacks(t, g)
	drcSettle(t, g, me.ID)
	for i := 0; i < 4 && len(g.PendingChoices) > 0; i++ {
		answerScryKeepAll(t, g, me.ID)
		drcSettle(t, g, me.ID)
	}
	for _, id := range bears {
		v := s58p6View(t, g, id)
		if v.CurrentPower() != 6 || v.CurrentToughness() != 6 {
			t.Errorf("attacker is %d/%d, want 6/6", v.CurrentPower(), v.CurrentToughness())
		}
	}
}

func TestYsBEyeOfNidhoggMakesADragonGoadedAndComesBack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	eye := castCatalogSpell(t, g, "Eye of Nidhogg", "Legendary Enchantment — Aura", ysBEyeOracle, b16TargetCard(bear))
	passPriorityAroundTable(t, g)

	v := s58p6View(t, g, bear)
	if v.CurrentPower() != 4 || v.CurrentToughness() != 2 {
		t.Errorf("enchanted creature is %d/%d, want 4/2", v.CurrentPower(), v.CurrentToughness())
	}
	if !v.HasSubtype("Dragon") || v.HasSubtype("Bear") {
		t.Errorf("subtypes %v, want only Dragon", v.Effective().Subtypes)
	}
	if !v.HasColor("B") {
		t.Error("enchanted creature is not black")
	}
	if !hasAbility(effectiveAbilities(t, g, bear), "flying") || !hasAbility(effectiveAbilities(t, g, bear), "deathtouch") {
		t.Error("enchanted creature lacks flying or deathtouch")
	}
	if !v.Goaded() {
		t.Error("enchanted creature is not goaded")
	}

	killCreature(t, g, me.ID, bear)
	drcSettle(t, g, me.ID)
	if !me.Hand.Contains(eye) {
		t.Error("Eye of Nidhogg did not return to its owner's hand")
	}
}

func TestYsBObservedStasisDrawsPerTappedCreatureAndSilencesTheHost(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := b12Push(g, opp.ID, "Host", "Creature — Bear", "", 2, 2)
	t1 := b12Creature(g, opp.ID, "Tapped One", "Creature — Bear", 1, 1)
	t2 := b12Creature(g, opp.ID, "Tapped Two", "Creature — Bear", 1, 1)
	b12Creature(g, opp.ID, "Untapped", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == t1 || g.Battlefield.Cards[i].InstanceID == t2 {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Observed Stasis", "Enchantment — Aura", ysBStasisOracle, b16TargetCard(host))
	passPriorityAroundTable(t, g)
	// The Aura left the hand, then two tapped creatures draw two.
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d → %d, want two draws (from %d, the Aura left)", hand, got, hand)
	}
	assertRestrictions(t, g, host, game.CantAttackOrBlock)
}
