package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b660_test.go — Invasion, batch 660 (#5616): domain.

const (
	oldB660StrengthOfUnity    = "00606d5e-a134-4363-9ee6-05e03d58f234"
	oldB660PowerArmor         = "1423b8e6-9165-4a89-a6ed-18085f460bca"
	oldB660WanderingStream    = "1f592b83-7838-4466-8bcb-093ac7346c55"
	oldB660KavuScout          = "2bd91ce8-1797-4dd7-8150-0d4caa7dad61"
	oldB660ExoticCurse        = "344476b1-8710-4c7c-826d-efb7d13ceaa7"
	oldB660CollectiveRestrain = "51f561db-ca7f-4661-bc04-99b8af02b8c9"
	oldB660CollapsingBorders  = "9fea3f42-643f-43c8-a9a9-4516dd72af5b"
	oldB660TribalFlames       = "c82eda61-d195-49c9-8ff0-a78dfa650689"
	oldB660WayfaringGiant     = "e20e5ad9-36b5-4962-acc2-555d069c5daf"
)

func TestOldB660EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB660StrengthOfUnity: "Strength of Unity", oldB660PowerArmor: "Power Armor", oldB660WanderingStream: "Wandering Stream",
		oldB660KavuScout: "Kavu Scout", oldB660ExoticCurse: "Exotic Curse", oldB660CollectiveRestrain: "Collective Restraint",
		oldB660CollapsingBorders: "Collapsing Borders", oldB660TribalFlames: "Tribal Flames", oldB660WayfaringGiant: "Wayfaring Giant",
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

// oldB660Lands gives `owner` one basic land of each named type.
func oldB660Lands(g *game.Game, owner uuid.UUID, types ...string) {
	for _, ty := range types {
		effectstest.B12Permanent(g, owner, ty, "Basic Land — "+ty)
	}
}

func TestOldB660WayfaringGiantAndKavuScoutGrowWithTheBasicLandTypesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giant := effectstest.PushCatalogPermanent(g, me.ID, "Wayfaring Giant", "Creature — Giant", oldB660WayfaringGiant, false)
	scout := effectstest.PushCatalogPermanent(g, me.ID, "Kavu Scout", "Creature — Kavu Scout", oldB660KavuScout, false)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, giant).Power, effectstest.FindBattlefieldCardForTest(g, giant).Toughness = 1, 3
		effectstest.FindBattlefieldCardForTest(g, scout).Power, effectstest.FindBattlefieldCardForTest(g, scout).Toughness = 0, 2
	})
	check := func(label string, domain int) {
		t.Helper()
		if p, tough := effectstest.EffectivePTOf(t, g, giant); p != 1+domain || tough != 3+domain {
			t.Errorf("%s: Giant is %d/%d, want %d/%d", label, p, tough, 1+domain, 3+domain)
		}
		if p, tough := effectstest.EffectivePTOf(t, g, scout); p != domain || tough != 2 {
			t.Errorf("%s: Scout is %d/%d, want %d/2", label, p, tough, domain)
		}
	}
	check("no lands", 0)
	oldB660Lands(g, me.ID, "Forest", "Forest", "Island")
	check("two types from three lands", 2)
	oldB660Lands(g, opp.ID, "Swamp", "Mountain", "Plains")
	check("an opponent's lands do not count", 2)
	oldB660Lands(g, me.ID, "Swamp")
	check("a third type", 3)
	effectstest.B12Permanent(g, me.ID, "Nonbasic Plains", "Land — Plains")
	check("a nonbasic land still has its basic land type", 4)
}

func TestOldB660DomainAurasPumpAndShrinkByTheControllersDomain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB660Lands(g, me.ID, "Forest", "Island")
	oldB660Lands(g, opp.ID, "Swamp", "Mountain", "Plains")
	host := effectstest.AuraBear(g, opp.ID)
	effectstest.AuraCast(t, g, "Strength of Unity", oldB660StrengthOfUnity, host)
	if p, tough := effectstest.EffectivePTOf(t, g, host); p != 4 || tough != 4 {
		t.Errorf("host is %d/%d, want 4/4 (+2/+2 from my two basic land types, not their three)", p, tough)
	}
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	oldB660Lands(g2, me2.ID, "Forest", "Island", "Swamp")
	host2 := effectstest.AuraBear(g2, g2.Seats[1].ID)
	effectstest.AuraCast(t, g2, "Exotic Curse", oldB660ExoticCurse, host2)
	if effectstest.OnBattlefieldByID(g2, host2) {
		t.Error("a 2/2 under -3/-3 should have died")
	}
	g3 := effectstest.NewCatalogGame(t)
	oldB660Lands(g3, g3.Seats[0].ID, "Forest")
	host3 := effectstest.AuraBear(g3, g3.Seats[1].ID)
	effectstest.AuraCast(t, g3, "Exotic Curse", oldB660ExoticCurse, host3)
	if p, tough := effectstest.EffectivePTOf(t, g3, host3); p != 1 || tough != 1 {
		t.Errorf("host is %d/%d, want 1/1", p, tough)
	}
}

func TestOldB660WanderingStreamGainsTwoPerBasicLandType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB660Lands(g, me.ID, "Forest", "Island", "Island")
	before := me.Life
	effectstest.CastCatalogSpell(t, g, "Wandering Stream", "Sorcery", oldB660WanderingStream, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+4 {
		t.Errorf("life %d→%d, want +4 for two types", before, me.Life)
	}
}

func TestOldB660TribalFlamesDealsDamageEqualToTheDomain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	oldB660Lands(g, me.ID, "Mountain", "Forest", "Swamp")
	before := opp.Life
	effectstest.CastCatalogSpell(t, g, "Tribal Flames", "Sorcery", oldB660TribalFlames, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("opponent life %d→%d, want -3", before, opp.Life)
	}
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Tribal Flames", "Sorcery", oldB660TribalFlames, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, bear) {
		t.Error("a 3/3 survived 3 damage")
	}
}

func TestOldB660PowerArmorPumpsByTheDomainUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := oldB660Seats(g)
	oldB660Lands(g, me.ID, "Mountain", "Forest")
	armor := effectstest.PushCatalogPermanent(g, me.ID, "Power Armor", "Artifact", oldB660PowerArmor, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := oldB660Activate(g, me.ID, armor, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Fatal("activated without paying {3}")
	}
	effectstest.FloatMana(t, g, me, "{C}")
	if err := oldB660Activate(g, me.ID, armor, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("Bear is %d/%d, want 4/4", p, tough)
	}
}

func TestOldB660CollectiveRestraintTaxesEachAttackerByTheDefendersDomain(t *testing.T) {
	g, bear, opp := effectstest.TaxTable(t)
	effectstest.PushCatalogPermanent(g, opp.ID, "Collective Restraint", "Enchantment", oldB660CollectiveRestrain, false)
	if p := g.PriceAttackDeclaration([]game.AttackDeclaration{{Attacker: bear, Target: opp.ID}}); !p.IsFree() {
		t.Errorf("with no lands the tax is %q, want nothing", p.Cost)
	}
	oldB660Lands(g, opp.ID, "Forest", "Island", "Swamp")
	oldB660Lands(g, g.Seats[0].ID, "Mountain", "Plains")
	if p := g.PriceAttackDeclaration([]game.AttackDeclaration{{Attacker: bear, Target: opp.ID}}); p.Total.Generic != 3 {
		t.Errorf("tax %d, want {3} (the defender's three types, not the attacker's)", p.Total.Generic)
	}
}

func TestOldB660CollapsingBordersGivesEachPlayerLifeForTheirDomainThenDamagesThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Collapsing Borders", "Enchantment", oldB660CollapsingBorders, false)
	oldB660Lands(g, opp.ID, "Forest", "Island")
	oldB660Lands(g, me.ID, "Mountain", "Mountain", "Swamp", "Plains", "Forest", "Island")
	before := opp.Life
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before+2-3 {
		t.Errorf("opponent life %d→%d in their upkeep, want +2 (their two types) then -3", before, opp.Life)
	}
	myBefore := me.Life
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myBefore+5-3 {
		t.Errorf("my life %d→%d in my upkeep, want +5 (my five types) then -3", myBefore, me.Life)
	}
}

func oldB660Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func oldB660Activate(g *game.Game, who, src uuid.UUID, idx int, p game.ActivateAbilityParams) error {
	p.Strict = true
	return g.ActivateCatalogAbility(who, src, idx, p)
}
