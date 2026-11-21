package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// equipment_legends_test.go covers the four big legendary Equipment
// that are not living weapons. Only what is new is asserted — a plain
// pump-and-grant is already pinned by Loxodon Warhammer.

const (
	aettirAndPriwenOracle = "19d8937e-23fb-4758-a306-0e7af9bb44c4"
	excaliburEdenOracle   = "9fb6bd72-031b-40a1-83c5-8a1c82f84e12"
	ultimaWeaponOracle    = "0dc4bf39-67d6-44e5-8993-dd18b20cfdee"
	mithrilCoatOracle     = "3dc364f3-3094-4660-80ca-9418588c7fde"
)

// --- Aettir and Priwen: a base P/T counted from the life total -----

// The three things layer 7b buys, in one test: the base tracks the
// life total, it MOVES when the total does, and +1/+1 counters land
// on top of it rather than being erased by it.
func TestAettirAndPriwenSetsBasePTFromTheLifeTotal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	aettir := effectstest.SeedEquipment(g, me.ID, "Aettir and Priwen", aettirAndPriwenOracle)

	me.Life = 31
	effectstest.EquipTo(t, g, me.ID, aettir, bear)

	if got := effectstest.EffectivePower(t, g, bear); got != 31 {
		t.Errorf("power %d, want 31 — the printed 2/2 is REPLACED, not added to", got)
	}
	if got := effectstest.EffectiveToughness(t, g, bear); got != 31 {
		t.Errorf("toughness %d, want 31", got)
	}

	// Through the real mutation, because the point of the assertion
	// is that a life change invalidates the layer cache at all.
	if _, err := g.ChangePlayerLife(me.ID, -19); err != nil {
		t.Fatalf("ChangePlayerLife: %v", err)
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 12 {
		t.Errorf("power %d after losing life, want 12 — the value is re-read, not captured", got)
	}

	// 7b SETS, so counters are added to the result rather than erased
	// by it. Counter math lives in CurrentPower, not in the layer
	// pass, so this is the read that answers "how big is it really".
	if err := g.AddCounter(bear, game.CounterPlusOne, 2); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	card := effectstest.BattlefieldCardCopy(t, g, bear)
	if got := card.CurrentPower(); got != 14 {
		t.Errorf("power %d with two +1/+1 counters, want 14 — counters apply ON TOP of the set base", got)
	}
	if got := card.CurrentToughness(); got != 14 {
		t.Errorf("toughness %d with two +1/+1 counters, want 14", got)
	}
}

// --- Excalibur: the self cost reduction ----------------------------

// "Total mana value of historic permanents you control" — summed, not
// counted, an opponent's board ignored, and a non-historic permanent
// of yours contributing nothing.
func TestExcaliburCostsLessForHistoricManaValueYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	if got := effectstest.SelfPricedMV(t, g, me, excaliburEdenOracle, "Legendary Artifact — Equipment", "{12}"); got != 12 {
		t.Errorf("Excalibur on an empty board: %d, want 12", got)
	}

	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Sol Ring", TypeLine: "Artifact", ManaCost: "{1}"})
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Commander", TypeLine: "Legendary Creature — Human", ManaCost: "{3}{W}{W}", Power: 3, Toughness: 3})
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Theirs", TypeLine: "Legendary Artifact", ManaCost: "{6}"})

	// {1} + {3}{W}{W} = 6. The Bear is not historic and the
	// opponent's six-drop is not yours.
	if got := effectstest.SelfPricedMV(t, g, me, excaliburEdenOracle, "Legendary Artifact — Equipment", "{12}"); got != 6 {
		t.Errorf("Excalibur with six historic mana value: %d, want 6", got)
	}

	// A reduction floors at zero generic and never goes negative.
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Colossus", TypeLine: "Legendary Artifact Creature — Golem", ManaCost: "{20}", Power: 10, Toughness: 10})
	if got := effectstest.SelfPricedMV(t, g, me, excaliburEdenOracle, "Legendary Artifact — Equipment", "{12}"); got != 0 {
		t.Errorf("Excalibur past twelve historic mana value: %d, want 0", got)
	}
}

func TestExcaliburEquipsOnlyALegendaryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	plain := effectstest.SeedBear(g, me.ID)
	legend := effectstest.SeedLegendaryCreature(g, me.ID, "Commander")
	sword := effectstest.SeedEquipment(g, me.ID, "Excalibur, Sword of Eden", excaliburEdenOracle)

	if err := g.ActivateCatalogAbility(me.ID, sword, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: plain}},
	}); err == nil {
		t.Error("equipped a non-legendary creature; the clause is 'equip legendary creature'")
	}

	effectstest.EquipTo(t, g, me.ID, sword, legend)
	if got := effectstest.EffectivePower(t, g, legend); got != 12 {
		t.Errorf("power %d, want 12 — a 2/2 plus +10/+0", got)
	}
	if got := effectstest.EffectiveToughness(t, g, legend); got != 2 {
		t.Errorf("toughness %d, want 2 — Excalibur is +10/+0", got)
	}
	if ab := effectstest.EffectiveAbilities(t, g, legend); !effectstest.ContainsString(ab, "vigilance") {
		t.Errorf("abilities %v missing vigilance", ab)
	}
}

// --- Ultima Weapon: a targeted attack trigger, opponents only ------

func TestUltimaWeaponDestroysAnOpponentsCreatureOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	mine := effectstest.SeedBear(g, me.ID)
	theirs := effectstest.SeedBear(g, opp.ID)
	weapon := effectstest.SeedEquipment(g, me.ID, "Ultima Weapon", ultimaWeaponOracle)
	effectstest.EquipTo(t, g, me.ID, weapon, bear)

	if got := effectstest.EffectivePower(t, g, bear); got != 9 {
		t.Fatalf("power %d, want 9", got)
	}

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)

	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no pick_target prompt from Ultima Weapon")
	}
	if !effectstest.HasID(prompt.PickTargetCards, theirs) {
		t.Fatalf("legal set %v missing the opponent's creature", prompt.PickTargetCards)
	}
	if effectstest.HasID(prompt.PickTargetCards, mine) || effectstest.HasID(prompt.PickTargetCards, bear) {
		t.Fatalf("legal set %v includes a creature YOU control", prompt.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(theirs) {
		t.Error("Ultima Weapon did not destroy the chosen creature")
	}
}

// --- Mithril Coat: attach on entry, and indestructible -------------

func TestMithrilCoatAttachesItselfOnEntryAndGrantsIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	legend := effectstest.SeedLegendaryCreature(g, me.ID, "Commander")

	coat := effectstest.CastCatalogSpell(t, g, "Mithril Coat", "Legendary Artifact — Equipment", mithrilCoatOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickCard(t, g, me.ID, legend)
		effectstest.PassPriorityAroundTable(t, g)
	}

	if host := effectstest.AttachmentHostOf(t, g, coat); host.Kind != game.TargetCard || host.ID != legend {
		t.Fatalf("Mithril Coat AttachedTo = %+v, want the legendary creature %s", host, legend)
	}
	if ab := effectstest.EffectiveAbilities(t, g, legend); !effectstest.ContainsString(ab, "indestructible") {
		t.Errorf("abilities %v missing indestructible", ab)
	}
	if ab := effectstest.EffectiveAbilities(t, g, coat); !effectstest.ContainsString(ab, "indestructible") {
		t.Errorf("the Coat's own abilities %v missing indestructible", ab)
	}
	if ab := effectstest.EffectiveAbilities(t, g, coat); !effectstest.ContainsString(ab, "flash") {
		t.Errorf("the Coat's own abilities %v missing flash", ab)
	}
}

// The trigger targets a LEGENDARY creature you control. With nothing
// legal on the board it is removed with no prompt (CR 603.3d) and the
// Coat simply sits there unattached.
func TestMithrilCoatWithNoLegendaryCreatureAttachesToNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedBear(g, me.ID)

	coat := effectstest.CastCatalogSpell(t, g, "Mithril Coat", "Legendary Artifact — Equipment", mithrilCoatOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Fatalf("a pick_target prompt with no legal target: %+v", p.PickTargetCards)
	}
	if host := effectstest.AttachmentHostOf(t, g, coat); host.Kind != "" && host.ID != uuid.Nil {
		t.Errorf("Mithril Coat attached to %+v with no legendary creature on the board", host)
	}
}
