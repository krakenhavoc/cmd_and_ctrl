package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b391_cards_test.go — Theros, life gain and loss, batch 391
// (#5561, S82 OLD). Every test plays the real catalog entry through the
// engine.

const (
	oldB391Scholar  = "18c5e7a0-4397-48bb-b993-1c3e3b08be2c"
	oldB391Maneuver = "28a24aeb-1ede-4331-adae-6b2dfcd3e4da"
	oldB391Boon     = "55821968-baf2-4fa6-8376-9ec305e27b3b"
	oldB391Elder    = "60e3ccf2-63f5-4435-a752-1d82fefb55d7"
	oldB391Soldier  = "a97f3131-586b-440f-a646-50d65e12e13f"
	oldB391Disciple = "aaefdb6f-306e-4fdc-9e49-af7a63e70f06"
	oldB391Sentry   = "fd4d83fa-070c-46bf-a620-041f69acdea9"
)

func TestOldB391CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB391Scholar: "Scholar of Athreos", oldB391Maneuver: "Cutthroat Maneuver", oldB391Boon: "Boon of Erebos",
		oldB391Elder: "Lagonna-Band Elder", oldB391Soldier: "Soldier of the Pantheon",
		oldB391Disciple: "Nylea's Disciple", oldB391Sentry: "Sentry of the Underworld",
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

func oldB391Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	return g, g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// oldB391Cast casts a card from the given player's hand where the game
// stands (no mana is asked of a non-strict cast) and returns its id.
func oldB391Cast(t *testing.T, g *game.Game, p *game.Player, name, typeLine, cost, oracle string, colors []string, strict bool, targets ...uuid.UUID) (uuid.UUID, error) {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle,
			Colors: colors, Owner: p.ID, Controller: p.ID})
	})
	err := g.CastSpell(p.ID, id, game.CastSpellParams{Strict: strict, Targets: effectstest.CardTargets(targets...)})
	return id, err
}

func oldB391Resolve(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
}

func TestOldB391ScholarOfAthreosDrainsEachOpponentAndGainsTheTotal(t *testing.T) {
	g, me, _ := oldB391Table(t)
	scholar := effectstest.PushCatalogPermanent(g, me.ID, "Scholar of Athreos", "Creature — Human Cleric", oldB391Scholar, false)
	effectstest.B06AddMana(me, "B", "B", "B")
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	effectstest.B16Activate(t, g, me.ID, scholar, 0, game.ActivateAbilityParams{})
	oldB391Resolve(t, g)
	opponents := 0
	for _, p := range g.Seats {
		if p.ID == me.ID {
			continue
		}
		opponents++
		if p.Life != lives[p.ID]-1 {
			t.Errorf("an opponent is at %d, want %d", p.Life, lives[p.ID]-1)
		}
	}
	if me.Life != lives[me.ID]+opponents {
		t.Errorf("life %d, want %d (one for each of %d opponents)", me.Life, lives[me.ID]+opponents, opponents)
	}
}

func TestOldB391CutthroatManeuverPumpsAndGrantsLifelinkToUpToTwo(t *testing.T) {
	g, me, opp := oldB391Table(t)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "C", "Creature — Bear", 2, 2)
	if _, err := oldB391Cast(t, g, me, "Cutthroat Maneuver", "Instant", "{0}", oldB391Maneuver, nil, true, a, b, c); err == nil {
		t.Fatal("three targets were accepted")
	}
	if _, err := oldB391Cast(t, g, me, "Cutthroat Maneuver", "Instant", "{0}", oldB391Maneuver, nil, true, a, b); err != nil {
		t.Fatalf("cast: %v", err)
	}
	oldB391Resolve(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if p, tough := effectstest.EffectivePTOf(t, g, id); p != 3 || tough != 3 {
			t.Errorf("%s is %d/%d, want 3/3", id, p, tough)
		}
		if !effectstest.HasKeywordOnBattlefield(t, g, id, "lifelink") {
			t.Errorf("%s has no lifelink", id)
		}
	}
	if p, _ := effectstest.EffectivePTOf(t, g, c); p != 2 {
		t.Errorf("the untargeted creature is pumped to %d", p)
	}
}

func TestOldB391BoonOfErebosPumpsRegeneratesAndCostsTwoLife(t *testing.T) {
	g, me, _ := oldB391Table(t)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	life := me.Life
	if _, err := oldB391Cast(t, g, me, "Boon of Erebos", "Instant", "{0}", oldB391Boon, nil, true); err == nil {
		t.Fatal("cast without a target")
	}
	if _, err := oldB391Cast(t, g, me, "Boon of Erebos", "Instant", "{0}", oldB391Boon, nil, true, bear); err != nil {
		t.Fatalf("cast: %v", err)
	}
	oldB391Resolve(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 4 {
		t.Errorf("power %d, want 4", p)
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
	effectstest.B27Kill(g, bear)
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(bear) {
		t.Error("the regeneration shield did not save the creature")
	}
	if !effectstest.B16Tapped(t, g, bear) {
		t.Error("a regenerated creature is untapped")
	}
}

func TestOldB391LagonnaBandElderGainsThreeOnlyWithAnEnchantment(t *testing.T) {
	for _, withEnchantment := range []bool{false, true} {
		g, me, _ := oldB391Table(t)
		if withEnchantment {
			effectstest.B21Push(g, me.ID, "Test Enchantment", "Enchantment", "", 0, 0)
		}
		life := me.Life
		if _, err := oldB391Cast(t, g, me, "Lagonna-Band Elder", "Creature — Centaur Advisor", "{2}{W}", oldB391Elder, []string{"W"}, false); err != nil {
			t.Fatalf("cast: %v", err)
		}
		oldB391Resolve(t, g)
		want := life
		if withEnchantment {
			want += 3
		}
		if me.Life != want {
			t.Errorf("with an enchantment %v: life %d, want %d", withEnchantment, me.Life, want)
		}
	}
}

func TestOldB391SoldierOfThePantheonGainsLifeForOpposingMulticoloredSpellsOnly(t *testing.T) {
	g, me, opp := oldB391Table(t)
	soldier := effectstest.PushCatalogPermanent(g, me.ID, "Soldier of the Pantheon", "Creature — Human Soldier", oldB391Soldier, false)
	if !effectstest.HasKeywordOnBattlefield(t, g, soldier, "protection from multicolored") {
		t.Error("no protection from multicolored")
	}
	life := me.Life
	if _, err := oldB391Cast(t, g, opp, "Mono Spell", "Instant", "{0}", "", []string{"R"}, false); err != nil {
		t.Fatalf("cast: %v", err)
	}
	oldB391Resolve(t, g)
	if me.Life != life {
		t.Errorf("a mono-colored spell gained %d life", me.Life-life)
	}
	if _, err := oldB391Cast(t, g, opp, "Gold Spell", "Instant", "{0}", "", []string{"R", "G"}, false); err != nil {
		t.Fatalf("cast: %v", err)
	}
	oldB391Resolve(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d after an opposing multicolored spell, want %d", me.Life, life+1)
	}
	if _, err := oldB391Cast(t, g, me, "Own Gold Spell", "Instant", "{0}", "", []string{"R", "G"}, false); err != nil {
		t.Fatalf("cast: %v", err)
	}
	oldB391Resolve(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d after its controller's own multicolored spell, want %d", me.Life, life+1)
	}
}

func TestOldB391NyleasDiscipleGainsLifeEqualToGreenDevotion(t *testing.T) {
	g, me, _ := oldB391Table(t)
	effectstest.B27Push(g, me.ID, "Green Bear", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	effectstest.B27Push(g, me.ID, "Red Bear", "Creature — Bear", "", "{R}{R}", 2, 2, "R")
	life := me.Life
	if _, err := oldB391Cast(t, g, me, "Nylea's Disciple", "Creature — Centaur Archer", "{2}{G}{G}", oldB391Disciple, []string{"G"}, false); err != nil {
		t.Fatalf("cast: %v", err)
	}
	oldB391Resolve(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d (devotion 1 from the Bear and 2 from the Disciple)", me.Life, life+3)
	}
}

func TestOldB391SentryOfTheUnderworldRegeneratesForThreeLife(t *testing.T) {
	g, me, _ := oldB391Table(t)
	sentry := effectstest.PushCatalogPermanent(g, me.ID, "Sentry of the Underworld", "Creature — Griffin Skeleton", oldB391Sentry, false)
	for _, kw := range []string{"flying", "vigilance"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, sentry, kw) {
			t.Errorf("the Sentry lacks %s", kw)
		}
	}
	effectstest.B06AddMana(me, "W", "B")
	life := me.Life
	effectstest.B16Activate(t, g, me.ID, sentry, 0, game.ActivateAbilityParams{})
	oldB391Resolve(t, g)
	if me.Life != life-3 {
		t.Errorf("life %d, want %d", me.Life, life-3)
	}
	effectstest.B27Kill(g, sentry)
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(sentry) {
		t.Error("the regeneration shield did not save the Sentry")
	}
	me.Life = 2
	effectstest.B06AddMana(me, "W", "B")
	if err := g.ActivateCatalogAbility(me.ID, sentry, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("the ability was activated at 2 life with 3 to pay")
	}
}
