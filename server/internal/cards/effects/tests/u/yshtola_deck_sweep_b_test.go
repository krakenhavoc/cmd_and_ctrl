package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestYsBAlisaieDiscountsOnlyTheSecondSpellEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Alisaie Leveilleur", "Legendary Creature — Elf Wizard", effectstest.YsBAlisaieOracle, false)
	if got := effectstest.PriceInHand(t, g, me, "Cultivate", "Sorcery", "{3}{G}"); got != 4 {
		t.Errorf("first spell: %d, want 4 (undiscounted)", got)
	}
	effectstest.B24CastNoncreature(t, g, "First", "Instant", "{U}")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PriceInHand(t, g, me, "Cultivate", "Sorcery", "{3}{G}"); got != 2 {
		t.Errorf("second spell: %d, want 2 ({2} less)", got)
	}
	effectstest.B24CastNoncreature(t, g, "Second", "Instant", "{U}")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PriceInHand(t, g, me, "Cultivate", "Sorcery", "{3}{G}"); got != 4 {
		t.Errorf("third spell: %d, want 4 (undiscounted)", got)
	}
}

func TestYsBAlphinaudDrawsOnTheSecondSpellOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Alphinaud Leveilleur", "Legendary Creature — Elf Wizard", effectstest.YsBAlphinaudOracle, false)
	hand := me.Hand.Size()
	effectstest.B24CastNoncreature(t, g, "First", "Instant", "{U}")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatalf("the first spell drew: hand %d → %d", hand, me.Hand.Size())
	}
	effectstest.B24CastNoncreature(t, g, "Second", "Instant", "{U}")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("the second spell should draw one: hand %d → %d", hand, me.Hand.Size())
	}
	effectstest.B24CastNoncreature(t, g, "Third", "Instant", "{U}")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("the third spell drew: hand %d → %d", hand, me.Hand.Size())
	}
}

func TestYsBReapersScytheGrowsBySoulCountersAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	scythe := effectstest.CastCatalogSpell(t, g, "Reaper's Scythe", "Artifact — Equipment", effectstest.YsBScytheOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	hero := effectstest.YsBHero(t, g, scythe)
	v := effectstest.S58p6View(t, g, hero)
	if v.CurrentPower() != 1 || !v.HasSubtype("Assassin") {
		t.Fatalf("Hero is %d/%d with Assassin=%v before any counters", v.CurrentPower(), v.CurrentToughness(), v.HasSubtype("Assassin"))
	}

	g.WithWriteLock(func() {
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -3)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -1)
	})
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, scythe, "soul"); got != 2 {
		t.Fatalf("soul counters %d, want 2 (two players lost life)", got)
	}
	v = effectstest.S58p6View(t, g, hero)
	if v.CurrentPower() != 3 || v.CurrentToughness() != 3 {
		t.Errorf("Hero is %d/%d, want 3/3", v.CurrentPower(), v.CurrentToughness())
	}
}

func TestYsBSagesNoulithsMakesAClericHeroWithPlusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	nouliths := effectstest.CastCatalogSpell(t, g, "Sage's Nouliths", "Artifact — Equipment", effectstest.YsBNoulithsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	hero := effectstest.YsBHero(t, g, nouliths)
	v := effectstest.S58p6View(t, g, hero)
	if v.CurrentPower() != 2 || v.CurrentToughness() != 1 {
		t.Errorf("Hero is %d/%d, want 2/1", v.CurrentPower(), v.CurrentToughness())
	}
	if !v.HasSubtype("Cleric") || !v.HasSubtype("Hero") {
		t.Errorf("Hero subtypes %v, want Hero and Cleric", v.Effective().Subtypes)
	}
}

func TestYsBPlanisphereGrowsTheHeroOnNoncreatureSpellsAndTheThirdDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	planisphere := effectstest.CastCatalogSpell(t, g, "Astrologian's Planisphere", "Artifact — Equipment", effectstest.YsBPlanisphereOracl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	hero := effectstest.YsBHero(t, g, planisphere)
	if !effectstest.S58p6View(t, g, hero).HasSubtype("Wizard") {
		t.Error("the Hero is not a Wizard")
	}
	effectstest.B24CastNoncreature(t, g, "Spell", "Instant", "{U}")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, hero, game.CounterPlusOne); got != 1 {
		t.Fatalf("after a noncreature spell: %d counters, want 1", got)
	}
	// However many cards are drawn, only the third of the turn counts.
	for i := 0; i < 6; i++ {
		g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.CountersOn(g, hero, game.CounterPlusOne); got != 2 {
		t.Errorf("after six draws: %d counters, want 2 (one for the third card)", got)
	}
}

func TestYsBChampionsFromBeyondMakesXHeroesAndPartiesUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastXSpell(t, g, "Champions from Beyond", "Enchantment", effectstest.YsBChampionsOracle, "{X}{W}{W}", 5, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Hero"); n != 5 {
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
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range heroes[:4] {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.DrcSettle(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Fatalf("Light Party drew before the scry was answered: %d", me.Hand.Size()-hand)
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.DrcSettle(t, g, me.ID)
	if me.Hand.Size() != hand+1 {
		t.Errorf("Light Party: drew %d, want 1", me.Hand.Size()-hand)
	}
	for _, id := range heroes[:4] {
		if effectstest.S58p6View(t, g, id).CurrentPower() != 1 {
			t.Error("Full Party fired with only four attackers")
		}
	}
}

func TestYsBChampionsFullPartyPumpsEightAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Champions from Beyond", "Enchantment", effectstest.YsBChampionsOracle, false)
	var bears []uuid.UUID
	for i := 0; i < 8; i++ {
		bears = append(bears, effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range bears {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.DrcSettle(t, g, me.ID)
	for i := 0; i < 4 && len(g.PendingChoices) > 0; i++ {
		effectstest.AnswerScryKeepAll(t, g, me.ID)
		effectstest.DrcSettle(t, g, me.ID)
	}
	for _, id := range bears {
		v := effectstest.S58p6View(t, g, id)
		if v.CurrentPower() != 6 || v.CurrentToughness() != 6 {
			t.Errorf("attacker is %d/%d, want 6/6", v.CurrentPower(), v.CurrentToughness())
		}
	}
}

func TestYsBEyeOfNidhoggMakesADragonGoadedAndComesBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	eye := effectstest.CastCatalogSpell(t, g, "Eye of Nidhogg", "Legendary Enchantment — Aura", effectstest.YsBEyeOracle, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)

	v := effectstest.S58p6View(t, g, bear)
	if v.CurrentPower() != 4 || v.CurrentToughness() != 2 {
		t.Errorf("enchanted creature is %d/%d, want 4/2", v.CurrentPower(), v.CurrentToughness())
	}
	if !v.HasSubtype("Dragon") || v.HasSubtype("Bear") {
		t.Errorf("subtypes %v, want only Dragon", v.Effective().Subtypes)
	}
	if !v.HasColor("B") {
		t.Error("enchanted creature is not black")
	}
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, bear), "flying") || !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, bear), "deathtouch") {
		t.Error("enchanted creature lacks flying or deathtouch")
	}
	if !v.Goaded() {
		t.Error("enchanted creature is not goaded")
	}

	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.DrcSettle(t, g, me.ID)
	if !me.Hand.Contains(eye) {
		t.Error("Eye of Nidhogg did not return to its owner's hand")
	}
}

func TestYsBObservedStasisDrawsPerTappedCreatureAndSilencesTheHost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Push(g, opp.ID, "Host", "Creature — Bear", "", 2, 2)
	t1 := effectstest.B12Creature(g, opp.ID, "Tapped One", "Creature — Bear", 1, 1)
	t2 := effectstest.B12Creature(g, opp.ID, "Tapped Two", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, opp.ID, "Untapped", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == t1 || g.Battlefield.Cards[i].InstanceID == t2 {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Observed Stasis", "Enchantment — Aura", effectstest.YsBStasisOracle, effectstest.B16TargetCard(host))
	effectstest.PassPriorityAroundTable(t, g)
	// The Aura left the hand, then two tapped creatures draw two.
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d → %d, want two draws (from %d, the Aura left)", hand, got, hand)
	}
	effectstest.AssertRestrictions(t, g, host, game.CantAttackOrBlock)
}
