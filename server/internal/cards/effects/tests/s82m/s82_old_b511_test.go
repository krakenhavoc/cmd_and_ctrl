package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b511_test.go — Shards of Alara, "Pump and combat tricks" (S82 OLD
// batch 511, #5588): each card driven through a real cast or activation.

const (
	oldB511EsperOracle    = "2551f383-7308-4670-9fda-406b20e372a5"
	oldB511NacatlOracle   = "35d3d96a-d303-4bf5-b972-f311331ec07a"
	oldB511ThoctarOracle  = "4a1b8db2-8814-45b3-ab8a-a1c8522d7fa1"
	oldB511KnightOracle   = "633e91ea-84d4-42ca-a8eb-603db340ebe7"
	oldB511SigilOracle    = "87b5d197-4400-47ff-89a7-9aeeb7c9a0b4"
	oldB511SurgeOracle    = "d8053319-b2a0-4df3-9787-3591d12cf1b9"
	oldB511NayaOracle     = "db4d2580-50f4-4986-8091-769b4d2f6d72"
	oldB511AsceticOracle  = "e5868cf5-9fdc-442d-95cb-3bc5e2727981"
	oldB511SorcererOracle = "f9921c9e-3248-488d-ad7b-8b064c5387e7"
)

func TestS82OldB511AllAreFullWithNoCaveats(t *testing.T) {
	for _, id := range []string{oldB511EsperOracle, oldB511NacatlOracle, oldB511ThoctarOracle, oldB511KnightOracle,
		oldB511SigilOracle, oldB511SurgeOracle, oldB511NayaOracle, oldB511AsceticOracle, oldB511SorcererOracle} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB511Perm(t *testing.T, g *game.Game, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	id := effectstest.SeedPermanentWithOracle(g, g.Seats[0].ID, name, typeLine, oracle)
	effectstest.UntapForTest(g, id)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].SummonedThisTurn = false
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 2, 2
			}
		}
	})
	return id
}

func oldB511Activate(t *testing.T, g *game.Game, src uuid.UUID, ability int, mana string, params game.ActivateAbilityParams) error {
	t.Helper()
	if mana != "" {
		effectstest.FloatMana(t, g, g.Seats[0], mana)
	}
	params.Strict = true
	return g.ActivateCatalogAbility(g.Seats[0].ID, src, ability, params)
}

func oldB511Target(id uuid.UUID) game.ActivateAbilityParams {
	return game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: id}}}
}

func TestS82OldB511EsperBattlemagePreventsTwoAndShrinksACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mage := oldB511Perm(t, g, "Esper Battlemage", "Artifact Creature — Human Wizard", oldB511EsperOracle)
	if err := oldB511Activate(t, g, mage, 0, "{W}", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("prevent: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	life := me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(opp.ID, me.ID, 5) })
	if me.Life != life-3 {
		t.Errorf("life %d -> %d, want -3 after preventing 2 of 5", life, me.Life)
	}
	// The mage is tapped now; untap and shrink a creature.
	effectstest.UntapForTest(g, mage)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := oldB511Activate(t, g, mage, 1, "{B}", oldB511Target(bear)); err != nil {
		t.Fatalf("-1/-1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, bear); p != 1 || tt != 1 {
		t.Errorf("bear %d/%d, want 1/1", p, tt)
	}
}

func TestS82OldB511WildNacatlGrowsWithMountainAndPlains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cat := effectstest.B12Push(g, me.ID, "Wild Nacatl", "Creature — Cat Warrior", oldB511NacatlOracle, 1, 1)
	effectstest.B12Permanent(g, opp.ID, "Their Mountain", "Basic Land — Mountain")
	if p, tt := effectstest.PtOf(t, g, cat); p != 1 || tt != 1 {
		t.Errorf("with only an opponent's Mountain: %d/%d, want 1/1", p, tt)
	}
	effectstest.B12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	if p, tt := effectstest.PtOf(t, g, cat); p != 2 || tt != 2 {
		t.Errorf("with a Mountain: %d/%d, want 2/2", p, tt)
	}
	plains := effectstest.B12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	if p, tt := effectstest.PtOf(t, g, cat); p != 3 || tt != 3 {
		t.Errorf("with both: %d/%d, want 3/3", p, tt)
	}
	effectstest.KillCreature(t, g, me.ID, plains)
	if p, tt := effectstest.PtOf(t, g, cat); p != 2 || tt != 2 {
		t.Errorf("after the Plains left: %d/%d, want 2/2", p, tt)
	}
}

func TestS82OldB511CavernThoctarPumpsPowerOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	thoctar := oldB511Perm(t, g, "Cavern Thoctar", "Creature — Beast", oldB511ThoctarOracle)
	for i := 0; i < 2; i++ {
		if err := oldB511Activate(t, g, thoctar, 0, "{C}{R}", game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if p, tt := effectstest.PtOf(t, g, thoctar); p != 4 || tt != 2 {
		t.Errorf("after two activations %d/%d, want 4/2 (from a 2/2 test body)", p, tt)
	}
}

func TestS82OldB511KnightOfTheSkywardEyeActivatesOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	knight := oldB511Perm(t, g, "Knight of the Skyward Eye", "Creature — Human Knight", oldB511KnightOracle)
	if err := oldB511Activate(t, g, knight, 0, "{C}{C}{C}{G}", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("first: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, knight); p != 5 || tt != 5 {
		t.Errorf("pumped %d/%d, want 5/5", p, tt)
	}
	if err := oldB511Activate(t, g, knight, 0, "{C}{C}{C}{G}", game.ActivateAbilityParams{}); err == nil {
		t.Error("a second activation in the same turn was accepted")
	}
}

func TestS82OldB511SigilBlessingBoostsTheTargetMoreThanTheOthers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Sigil Blessing", "Instant", oldB511SigilOracle, effectstest.B16TargetCard(theirs)); err == nil {
		t.Fatal("Sigil Blessing targeted a creature its caster does not control")
	}
	effectstest.CastCatalogSpell(t, g, "Sigil Blessing", "Instant", oldB511SigilOracle, effectstest.B16TargetCard(a))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, a); p != 5 || tt != 5 {
		t.Errorf("target %d/%d, want 5/5", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, b); p != 3 || tt != 3 {
		t.Errorf("other creature %d/%d, want 3/3", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, theirs); p != 2 || tt != 2 {
		t.Errorf("opponent's creature %d/%d, want 2/2", p, tt)
	}
}

func TestS82OldB511SangriteSurgeGivesPumpAndDoubleStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Sangrite Surge", "Sorcery", oldB511SurgeOracle, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, bear); p != 5 || tt != 5 {
		t.Errorf("target %d/%d, want 5/5", p, tt)
	}
	if !effectstest.P1g2HasKeyword(g, bear, "double strike") {
		t.Error("no double strike")
	}
}

func TestS82OldB511NayaBattlemagePumpsOrTaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	mage := oldB511Perm(t, g, "Naya Battlemage", "Creature — Human Shaman", oldB511NayaOracle)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := oldB511Activate(t, g, mage, 0, "{R}", oldB511Target(bear)); err != nil {
		t.Fatalf("+2/+0: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, bear); p != 4 || tt != 2 {
		t.Errorf("bear %d/%d, want 4/2", p, tt)
	}
	effectstest.UntapForTest(g, mage)
	if err := oldB511Activate(t, g, mage, 1, "{W}", oldB511Target(bear)); err != nil {
		t.Fatalf("tap: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := g.LookupCardForEffect(bear); !c.Tapped {
		t.Error("the targeted creature is not tapped")
	}
}

func TestS82OldB511TopanAsceticTapsAnyUntappedCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	asc := oldB511Perm(t, g, "Topan Ascetic", "Creature — Human Monk", oldB511AsceticOracle)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 1, 1)
	if err := oldB511Activate(t, g, asc, 0, "", game.ActivateAbilityParams{TapIDs: []uuid.UUID{other}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, asc); p != 3 || tt != 3 {
		t.Errorf("ascetic %d/%d, want 3/3", p, tt)
	}
	if c, _ := g.LookupCardForEffect(other); !c.Tapped {
		t.Error("the paying creature is not tapped")
	}
	if c, _ := g.LookupCardForEffect(asc); c.Tapped {
		t.Error("the Ascetic tapped though another creature paid")
	}
}

func TestS82OldB511SightedCasteSorcererGainsShroud(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	sorc := oldB511Perm(t, g, "Sighted-Caste Sorcerer", "Creature — Human Wizard Sorcerer", oldB511SorcererOracle)
	if effectstest.P1g2HasKeyword(g, sorc, "shroud") {
		t.Fatal("starts with shroud")
	}
	if err := oldB511Activate(t, g, sorc, 0, "{U}", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.P1g2HasKeyword(g, sorc, "shroud") {
		t.Error("did not gain shroud")
	}
	if !effectstest.P1g2HasKeyword(g, sorc, "exalted") {
		t.Error("exalted is not printed on the card")
	}
}
