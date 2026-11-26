package testss82m

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB393Charlatan = "11dfe74c-bc8e-4d58-b526-41ea5b409135"
	oldB393Pyxis     = "4b746fe1-4c11-4dfd-bdc9-edd394478861"
	oldB393Warden    = "805d1f48-dd91-4034-b42c-3f1f1cb75cc9"
	oldB393Griffin   = "8356b1fe-02f8-4bb9-bcdb-bb7454eb18ed"
	oldB393Piper     = "d15eed8e-6c2c-4427-9f72-a46e90c562b9"
	oldB393Phalanx   = "f8d64a8e-1bff-4960-bc36-1fbb4fc18c8f"
)

func TestOldB393AreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB393Charlatan: "Meletis Charlatan", oldB393Pyxis: "Pyxis of Pandemonium",
		oldB393Warden: "Ephara's Warden", oldB393Griffin: "Decorated Griffin",
		oldB393Piper: "Satyr Piper", oldB393Phalanx: "Returned Phalanx",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB393EpharasWardenTapsOnlyPowerThreeOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	warden := effectstest.PushCatalogPermanent(g, me.ID, "Ephara's Warden", "Creature — Human Cleric", oldB393Warden, false)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 3, 3)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Ogre", 4, 4)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, warden, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(big), Strict: true}); err == nil {
		t.Fatal("tapped a creature with power 4")
	}
	if err := g.ActivateCatalogAbility(me.ID, warden, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(small), Strict: true}); err != nil {
		t.Fatalf("tap: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, small) {
		t.Error("the small creature is not tapped")
	}
	if !effectstest.B16Tapped(t, g, warden) {
		t.Error("the Warden did not tap for its cost")
	}
}

func TestOldB393SatyrPiperMakesTheTargetMustBeBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	piper := effectstest.PushCatalogPermanent(g, me.ID, "Satyr Piper", "Creature — Satyr Rogue", oldB393Piper, false)
	bear := effectstest.ReqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	effectstest.ReqCreature(g, opp.ID, "Wall A", "", 0, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "G", 4)
	if err := g.ActivateCatalogAbility(me.ID, piper, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ReqToBlockers(t, g, bear)
	br := effectstest.ReqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); !strings.Contains(got, "Bear must be blocked") || !strings.Contains(got, "Satyr Piper") {
		t.Errorf("sentence = %q", got)
	}
}

func TestOldB393ReturnedPhalanxMayAttackOnlyAfterThePayment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	phalanx := effectstest.PushCatalogPermanent(g, me.ID, "Returned Phalanx", "Creature — Zombie Soldier", oldB393Phalanx, false)
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, phalanx)
		c.Power, c.Toughness = 3, 3
	})
	if !effectstest.HasEffectiveKeyword(t, g, phalanx, "defender") {
		t.Fatal("no defender")
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.DeclareAttacker(phalanx, opp.ID); err == nil {
		t.Fatal("a defender attacked")
	}
	effectstest.FillPoolColored(me, "U", 2)
	if err := g.ActivateCatalogAbility(me.ID, phalanx, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, phalanx)
	if c, ok := effectstest.BattlefieldCard(g, phalanx); !ok || c.AttackingTarget != opp.ID {
		t.Error("the Phalanx is not attacking")
	}
	if !effectstest.HasEffectiveKeyword(t, g, phalanx, "defender") {
		t.Error("the keyword should stay")
	}
}

func TestOldB393DecoratedGriffinPreventsOnlyOneCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Ogre", "Creature — Ogre", 3, 3)
	griffin := effectstest.PushCatalogPermanent(g, opp.ID, "Decorated Griffin", "Creature — Griffin", oldB393Griffin, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(opp, "W", 2)
	if err := g.ActivateCatalogAbility(opp.ID, griffin, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	life := opp.Life
	effectstest.AttackWith(t, g, opp.ID, attacker)
	effectstest.InitSettle(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d, want %d (3 combat damage, 1 prevented)", opp.Life, life-2)
	}
}

func TestOldB393DecoratedGriffinDoesNotPreventNoncombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	griffin := effectstest.PushCatalogPermanent(g, opp.ID, "Decorated Griffin", "Creature — Griffin", oldB393Griffin, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(opp, "W", 2)
	if err := g.ActivateCatalogAbility(opp.ID, griffin, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	life := opp.Life
	_ = me
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.Pr6Player(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d, want %d (a Bolt is not combat damage)", opp.Life, life-3)
	}
}

func TestOldB393MeletisCharlatanLetsTheSpellsControllerCopyIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, mage, third := g.Seats[0], g.Seats[1], g.Seats[2]
	charlatan := effectstest.PushCatalogPermanent(g, mage.ID, "Meletis Charlatan", "Creature — Human Wizard", oldB393Charlatan, false)
	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.Pr6Player(third.ID))
	effectstest.FillPoolColored(mage, "U", 3)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass: %v", err)
	}
	if err := g.ActivateCatalogAbility(mage.ID, charlatan, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bolt), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	// Resolve the ability, then the copy's controller (the Bolt's caster) is
	// asked for new targets.
	for i := 0; i < 8 && effectstest.LatestPickTarget(g, caster.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	pick := effectstest.LatestPickTarget(g, caster.ID)
	if pick == nil {
		t.Fatal("the copy's controller, the Bolt's caster, was not asked for new targets")
	}
	if err := g.ResolvePickTarget(pick.ID, caster.ID, game.TargetRef{Kind: game.TargetPlayer, ID: mage.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	life3, lifeMage := third.Life, mage.Life
	effectstest.PassPriorityAroundTable(t, g)
	if third.Life != life3-3 || mage.Life != lifeMage-3 {
		t.Errorf("lives %d / %d, want %d / %d (the Bolt and its copy re-aimed at the Charlatan's controller)", third.Life, mage.Life, life3-3, lifeMage-3)
	}
}

func TestOldB393MeletisCharlatanRefusesAPermanentSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	charlatan := effectstest.PushCatalogPermanent(g, me.ID, "Meletis Charlatan", "Creature — Human Wizard", oldB393Charlatan, false)
	bear := effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.FillPoolColored(me, "U", 3)
	if err := g.ActivateCatalogAbility(me.ID, charlatan, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear), Strict: true}); err == nil {
		t.Fatal("targeted a creature spell")
	}
}

func oldB393ExileIDsOf(g *game.Game, owner uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Exile.Cards {
		if c.Owner == owner {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

func TestOldB393PyxisExilesFaceDownThenReleasesOnlyPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	a, b, c := g.Seats[0], g.Seats[1], g.Seats[2]
	pyxis := effectstest.PushCatalogPermanent(g, a.ID, "Pyxis of Pandemonium", "Artifact", oldB393Pyxis, false)
	creature := effectstest.TopOfLibrary(a, "Top Bear", "Creature — Bear", "{1}{G}")
	instant := effectstest.TopOfLibrary(b, "Top Bolt", "Instant", "{R}")
	land := effectstest.TopOfLibrary(c, "Top Forest", "Land — Forest", "")
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(a.ID, pyxis, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{T}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{creature, instant, land} {
		if !effectstest.ExileHas(g, id) {
			t.Fatalf("card %s was not exiled", id)
		}
		var down bool
		g.WithWriteLock(func() {
			for _, ec := range g.Exile.Cards {
				if ec.InstanceID == id {
					down = ec.FaceDown
				}
			}
		})
		if !down {
			t.Errorf("card %s is not face down", id)
		}
	}
	// An unrelated exiled card is not released.
	strayCard := game.NewCard("Stray", b.ID)
	strayCard.TypeLine = "Creature — Bear"
	stray := strayCard.InstanceID
	g.Exile.PushTop(strayCard)
	effectstest.P1g3Untap(g, pyxis)
	effectstest.FillPool(a, 7)
	if err := g.ActivateCatalogAbility(a.ID, pyxis, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{7},{T}, sacrifice: %v", err)
	}
	if g.Battlefield.Contains(pyxis) {
		t.Error("the Pyxis is sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(creature) && !oldB393OnBattlefieldUnderOwner(g, a.ID, "Top Bear") {
		t.Error("the creature card did not enter under its owner's control")
	}
	if !oldB393OnBattlefieldUnderOwner(g, c.ID, "Top Forest") {
		t.Error("the land card did not enter under its owner's control")
	}
	if !effectstest.ExileHas(g, instant) {
		t.Error("the instant card should stay in exile")
	}
	g.WithWriteLock(func() {
		for _, ec := range g.Exile.Cards {
			if ec.InstanceID == instant && ec.FaceDown {
				t.Error("the instant card should be turned face up")
			}
		}
	})
	if !effectstest.ExileHas(g, stray) {
		t.Error("a card not exiled with the Pyxis was released")
	}
}

func oldB393OnBattlefieldUnderOwner(g *game.Game, owner uuid.UUID, name string) bool {
	found := false
	g.WithWriteLock(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name && c.Controller == owner && c.Owner == owner {
				found = true
			}
		}
	})
	return found
}
