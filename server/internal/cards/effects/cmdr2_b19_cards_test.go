package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_b19_cards_test.go — Commander wave 2, batch 19 (#5368): simple Auras.

const (
	cmdr2VowOfDuty      = "1e7bddbd-e448-4aea-9d18-b44bd45837c5"
	cmdr2VowOfLightning = "bfed67f3-70d2-49c2-8f25-16e4d10fa5be"
)

func TestCmdr2B19EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		cmdr2VowOfDuty:      "Vow of Duty",
		cmdr2VowOfLightning: "Vow of Lightning",
	}
	for oracle, name := range want {
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

// cmdr2B19Attach puts an Aura of the catalog onto the battlefield attached
// to host, controlled by owner.
func cmdr2B19Attach(g *game.Game, owner, host uuid.UUID, name, oracle string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: auraTypeLine,
		OracleID: oracle, Owner: owner, Controller: owner, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host}})
}

func TestCmdr2B19VowsPumpGrantAndBarAttackingTheirController(t *testing.T) {
	for _, tc := range []struct{ name, oracle, kw string }{
		{"Vow of Duty", cmdr2VowOfDuty, "vigilance"},
		{"Vow of Lightning", cmdr2VowOfLightning, "first strike"},
	} {
		g := newCatalogGame(t)
		me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
		host := b12Creature(g, opp.ID, "Host", "Creature — Bear", 2, 2)
		other := b12Creature(g, opp.ID, "Other", "Creature — Bear", 2, 2)
		cmdr2B19Attach(g, me.ID, host, tc.name, tc.oracle)
		if p, to := effectivePTOf(t, g, host); p != 4 || to != 4 {
			t.Errorf("%s: host is %d/%d, want 4/4", tc.name, p, to)
		}
		if !hasEffectiveKeyword(t, g, host, tc.kw) {
			t.Errorf("%s: host lacks %s", tc.name, tc.kw)
		}
		var walker uuid.UUID
		g.WithWriteLock(func() {
			walker = pushPermanent(g, me.ID, game.Card{Name: "My Walker", TypeLine: "Legendary Planeswalker — Test",
				Counters: map[string]int{game.CounterLoyalty: 3}})
		})
		advanceToDeclareAttackersOf(t, g, 1)
		for _, target := range []uuid.UUID{me.ID, walker} {
			if err := g.Clone().DeclareAttacker(host, target); !errors.Is(err, game.ErrIllegalAttackTarget) {
				t.Errorf("%s: enchanted creature attacking %s: err = %v, want ErrIllegalAttackTarget", tc.name, target, err)
			}
		}
		if err := g.Clone().DeclareAttacker(host, third.ID); err != nil {
			t.Errorf("%s: enchanted creature attacking another player: %v", tc.name, err)
		}
		if err := g.Clone().DeclareAttacker(other, me.ID); err != nil {
			t.Errorf("%s: an unenchanted creature attacking me: %v", tc.name, err)
		}
	}
}

const (
	cmdr2VerdantEmbrace    = "f0bbd988-50b7-476c-9183-dfabd4a95079"
	cmdr2PollenbrightWings = "582c8030-1d41-4c63-8ed6-06aad398a96c"
)

func TestCmdr2B19VerdantEmbracePumpsAndMakesSaprolingsForTheHostsControllerEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := b12Creature(g, opp.ID, "Host", "Creature — Bear", 2, 2)
	cmdr2B19Attach(g, me.ID, host, "Verdant Embrace", cmdr2VerdantEmbrace)
	if p, to := effectivePTOf(t, g, host); p != 5 || to != 5 {
		t.Fatalf("host is %d/%d, want 5/5", p, to)
	}
	advanceTo(t, g, game.StepUpkeep)
	passPriorityAroundTable(t, g)
	if n := namedOnBattlefield(g, "Saproling", opp.ID); n != 1 {
		t.Errorf("%d Saprolings for the enchanted creature's controller on my upkeep, want 1", n)
	}
	if n := namedOnBattlefield(g, "Saproling", me.ID); n != 0 {
		t.Errorf("%d Saprolings for the Aura's controller, want 0", n)
	}
}

func TestCmdr2B19PollenbrightWingsGivesFlyingAndMakesThatManySaprolings(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	host := b12Creature(g, me.ID, "Host", "Creature — Bear", 3, 3)
	other := b12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	cmdr2B19Attach(g, me.ID, host, "Pollenbright Wings", cmdr2PollenbrightWings)
	if !hasEffectiveKeyword(t, g, host, "flying") {
		t.Fatal("the enchanted creature lacks flying")
	}
	dealCombatDamageToPlayer(g, other, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if n := namedOnBattlefield(g, "Saproling", me.ID); n != 0 {
		t.Fatalf("%d Saprolings after an unenchanted creature connected", n)
	}
	dealCombatDamageToPlayer(g, host, opp.ID, 3)
	passPriorityAroundTable(t, g)
	if n := namedOnBattlefield(g, "Saproling", me.ID); n != 3 {
		t.Errorf("%d Saprolings, want 3 (the damage dealt)", n)
	}
}

const cmdr2WarbriarBlessing = "4a58954a-1d20-48b8-847c-3615de67e122"

func TestCmdr2B19WarbriarBlessingPumpsToughnessAndFightsOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ecl6Seats(g)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	ecl6Cast(t, g, me, "Warbriar Blessing", "Enchantment — Aura", cmdr2WarbriarBlessing, 0, 0,
		game.CastSpellParams{Targets: cardRefs(mine)})
	ecl6ResolveSpell(t, g)
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, mine); p != 2 || tt != 4 {
		t.Errorf("enchanted creature is %d/%d, want 2/4", p, tt)
	}
	if len(battlefieldIDsNamed(g, "Theirs")) != 0 {
		t.Error("the opposing creature survived the fight")
	}
	if len(battlefieldIDsNamed(g, "Mine")) != 1 {
		t.Error("the enchanted creature died")
	}
}

func TestCmdr2B19WarbriarBlessingOnlyEnchantsYourOwnCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ecl6Seats(g)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 3, 3)
	advanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Warbriar Blessing", TypeLine: "Enchantment — Aura",
			OracleID: cmdr2WarbriarBlessing, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: cardRefs(theirs)}); err == nil {
		t.Error("enchanting an opponent's creature was allowed")
	}
}

const cmdr2SpectralSteel = "95f74c0d-3f89-4944-997e-1bdc408bae71"

func TestCmdr2B19SpectralSteelPumpsAndReturnsAnotherAuraOrEquipmentFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	host := b12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	cmdr2B19Attach(g, me.ID, host, "Spectral Steel", cmdr2SpectralSteel)
	if p, tt := ptOf(t, g, host); p != 4 || tt != 4 {
		t.Fatalf("enchanted creature is %d/%d, want 4/4", p, tt)
	}
	toMainForCost(t, g)
	steel := pushCatalogGraveyardCard(me, "Spectral Steel", "Enchantment — Aura", cmdr2SpectralSteel, 0, 0)
	aura := pushCatalogGraveyardCard(me, "Gy Aura", "Enchantment — Aura", "", 0, 0)
	gear := pushCatalogGraveyardCard(me, "Gy Blade", "Artifact — Equipment", "", 0, 0)
	bear := pushCatalogGraveyardCard(me, "Gy Bear", "Creature — Bear", "", 2, 2)
	floatForTest(g, me, "WC")
	for _, bad := range []uuid.UUID{steel, bear} {
		if err := g.ActivateCatalogAbility(me.ID, steel, 0, game.ActivateAbilityParams{Targets: cardTarget(bad)}); err == nil {
			t.Fatalf("%s was accepted; the target is another Aura or Equipment card", bad)
		}
	}
	b16Activate(t, g, me.ID, steel, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(aura)})
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(aura) {
		t.Error("the Aura card did not return to hand")
	}
	if !g.Exile.Contains(steel) {
		t.Error("Spectral Steel was not exiled")
	}
	_ = gear
}
