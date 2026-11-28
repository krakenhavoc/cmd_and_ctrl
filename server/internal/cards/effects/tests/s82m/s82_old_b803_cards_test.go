package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b803_cards_test.go — Mirage's removal, second batch, played from the
// real catalog entries (S82 OLD batch 803).

const (
	b803DomainOracle  = "87547bf0-02cf-4d83-9a1f-c82d2f1a22f1"
	b803AmuletOracle  = "945d5351-ce7d-488c-bc63-269ad82b74cb"
	b803CharmOracle   = "b81e5dbd-47a5-4994-91b7-e2a50c8e098d"
	b803TributeOracle = "c06b993b-a36a-412a-a732-d678a4179791"
	b803PurgeOracle   = "daa86c76-aacb-4244-9639-81f758c634b2"
	b803HeartOracle   = "f38b03dd-a200-48a6-ab65-fb2ac47456df"
	b803SeedsOracle   = "f5ebad5c-9675-4a37-b76c-760416083939"
	b803CloudOracle   = "f64ea1a7-0e10-4549-8056-25ebc2ce0175"
	b803Bear          = "Creature — Bear"
)

func TestS82OldB803CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b803DomainOracle:  "Tranquil Domain",
		b803AmuletOracle:  "Amulet of Unmaking",
		b803CharmOracle:   "Chaos Charm",
		b803TributeOracle: "Phyrexian Tribute",
		b803PurgeOracle:   "Kaervek's Purge",
		b803HeartOracle:   "Serene Heart",
		b803SeedsOracle:   "Seeds of Innocence",
		b803CloudOracle:   "Cinder Cloud",
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

func b803Cast(t *testing.T, g *game.Game, p *game.Player, name, typeLine, oracle string, params game.CastSpellParams) error {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: p.ID, Controller: p.ID})
	effectstest.AdvanceToMain(t, g)
	return g.CastSpell(p.ID, id, params)
}

func TestTranquilDomainDestroysEveryEnchantmentButAuras(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Permanent(g, me.ID, "Ward", "Enchantment")
	theirs := effectstest.B12Permanent(g, opp.ID, "Rite", "Enchantment")
	aura := effectstest.B12Permanent(g, opp.ID, "Pacifism", "Enchantment — Aura")
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	if err := b803Cast(t, g, me, "Tranquil Domain", "Instant", b803DomainOracle, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("a non-Aura enchantment survived")
	}
	if !g.Battlefield.Contains(aura) || !g.Battlefield.Contains(rock) {
		t.Error("an Aura or an artifact was destroyed")
	}
}

func TestSereneHeartDestroysOnlyAuras(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Permanent(g, me.ID, "Ward", "Enchantment")
	aura := effectstest.B12Permanent(g, opp.ID, "Pacifism", "Enchantment — Aura")
	aura2 := effectstest.B12Permanent(g, me.ID, "Lure", "Enchantment — Aura")
	if err := b803Cast(t, g, me, "Serene Heart", "Instant", b803HeartOracle, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) || g.Battlefield.Contains(aura2) {
		t.Error("an Aura survived")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("a non-Aura enchantment was destroyed")
	}
}

func TestAmuletOfUnmakingExilesAnArtifactCreatureOrLandAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	amulet := effectstest.PushCatalogPermanent(g, me.ID, "Amulet of Unmaking", "Artifact", b803AmuletOracle, false)
	ench := effectstest.B12Permanent(g, opp.ID, "Rite", "Enchantment")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", b803Bear, 2, 2)
	pay := func() {
		for i := 0; i < 5; i++ {
			me.ManaPool.AddMana(game.ManaToken{Color: "C"})
		}
	}
	effectstest.AdvanceToMain(t, g)
	pay()
	activate := func(target uuid.UUID) error {
		return g.ActivateCatalogAbility(me.ID, amulet, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(target)})
	}
	if err := activate(ench); err == nil {
		t.Fatal("targeted an enchantment: accepted")
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	pay()
	if err := activate(bear); err == nil {
		t.Fatal("activated outside a main phase: accepted")
	}
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	pay()
	if err := activate(bear); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(amulet) {
		t.Error("the Amulet was not exiled as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature was not exiled")
	}
	for _, c := range g.Exile.Cards {
		if c.InstanceID == bear {
			return
		}
	}
	t.Error("the creature is not in exile")
}

func TestChaosCharmEachMode(t *testing.T) {
	// Mode 0: destroy target Wall.
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", b803Bear, 2, 2)
	if err := b803Cast(t, g, me, "Chaos Charm", "Instant", b803CharmOracle, game.CastSpellParams{Strict: true, Modes: []int{0}, Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Fatal("mode 0 targeted a non-Wall: accepted")
	}
	if err := b803Cast(t, g, me, "Chaos Charm", "Instant", b803CharmOracle, game.CastSpellParams{Strict: true, Modes: []int{0}, Targets: effectstest.CardTarget(wall)}); err != nil {
		t.Fatalf("mode 0: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wall) {
		t.Error("mode 0: the Wall survived")
	}

	// Mode 1: one damage.
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	bear = effectstest.B12Creature(g, opp.ID, "Bear", b803Bear, 2, 3)
	if err := b803Cast(t, g, me, "Chaos Charm", "Instant", b803CharmOracle, game.CastSpellParams{Strict: true, Modes: []int{1}, Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("mode 1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, bear); got != 1 {
		t.Errorf("mode 1: %d damage, want 1", got)
	}

	// Mode 2: haste.
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	bear = effectstest.B12Creature(g, me.ID, "Bear", b803Bear, 2, 2)
	if err := b803Cast(t, g, me, "Chaos Charm", "Instant", b803CharmOracle, game.CastSpellParams{Strict: true, Modes: []int{2}, Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("mode 2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "haste") {
		t.Error("mode 2: the creature did not gain haste")
	}
	_ = opp
}

func TestPhyrexianTributeNeedsTwoCreaturesAndDestroysAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "A", b803Bear, 1, 1)
	b := effectstest.B12Creature(g, me.ID, "B", b803Bear, 1, 1)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	if err := b803Cast(t, g, me, "Phyrexian Tribute", "Sorcery", b803TributeOracle, game.CastSpellParams{Strict: true, SacrificeIDs: []uuid.UUID{a}, Targets: effectstest.CardTarget(rock)}); err == nil {
		t.Fatal("cast sacrificing one creature: accepted")
	}
	if err := b803Cast(t, g, me, "Phyrexian Tribute", "Sorcery", b803TributeOracle, game.CastSpellParams{Strict: true, SacrificeIDs: []uuid.UUID{a, b}, Targets: effectstest.CardTarget(rock)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Errorf("artifact %v, creatures %v %v, want all gone", g.Battlefield.Contains(rock), g.Battlefield.Contains(a), g.Battlefield.Contains(b))
	}
}

func TestKaerveksPurgeNeedsManaValueXAndHitsTheController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	three := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Three", TypeLine: b803Bear, ManaCost: "{2}{G}", Power: 4, Toughness: 4})
	two := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Two", TypeLine: b803Bear, ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	if err := b803Cast(t, g, me, "Kaervek's Purge", "Sorcery", b803PurgeOracle, game.CastSpellParams{Strict: true, XValue: 3, Targets: effectstest.CardTarget(two)}); err == nil {
		t.Fatal("X=3 targeted a mana value 2 creature: accepted")
	}
	life := opp.Life
	if err := b803Cast(t, g, me, "Kaervek's Purge", "Sorcery", b803PurgeOracle, game.CastSpellParams{Strict: true, XValue: 3, Targets: effectstest.CardTarget(three)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(three) {
		t.Error("the creature survived")
	}
	if opp.Life != life-4 {
		t.Errorf("life %d -> %d, want -4 (its power)", life, opp.Life)
	}
}

func TestSeedsOfInnocenceDestroysArtifactsAndPaysTheirControllers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Mine", TypeLine: "Artifact", ManaCost: "{3}"})
	theirs := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Theirs", TypeLine: "Artifact", ManaCost: "{1}"})
	theirs2 := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Theirs 2", TypeLine: "Artifact Creature — Golem", ManaCost: "{4}", Power: 2, Toughness: 2})
	other := effectstest.B12Permanent(g, opp.ID, "Rite", "Enchantment")
	myLife, oppLife := me.Life, opp.Life
	if err := b803Cast(t, g, me, "Seeds of Innocence", "Sorcery", b803SeedsOracle, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirs, theirs2} {
		if g.Battlefield.Contains(id) {
			t.Error("an artifact survived")
		}
	}
	if !g.Battlefield.Contains(other) {
		t.Error("an enchantment was destroyed")
	}
	if me.Life != myLife+3 || opp.Life != oppLife+5 {
		t.Errorf("life: caster %d -> %d (want +3), opponent %d -> %d (want +5)", myLife, me.Life, oppLife, opp.Life)
	}
}

func TestCinderCloudHurtsTheControllerOfADyingWhiteCreature(t *testing.T) {
	for _, tc := range []struct {
		colour string
		dmg    int
	}{{"W", 3}, {"G", 0}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		victim := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Victim", TypeLine: b803Bear, Colors: []string{tc.colour}, Power: 3, Toughness: 3})
		life := opp.Life
		if err := b803Cast(t, g, me, "Cinder Cloud", "Instant", b803CloudOracle, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(victim)}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(victim) {
			t.Errorf("%s: the creature survived", tc.colour)
		}
		if opp.Life != life-tc.dmg {
			t.Errorf("%s: life %d -> %d, want -%d", tc.colour, life, opp.Life, tc.dmg)
		}
	}
}
