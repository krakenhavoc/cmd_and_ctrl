package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 797 (Mirage, #5633): the two guildmages, Ivory Charm, Goblin
// Elite Infantry, Crimson Roc, Barbed-Back Wurm, Azimaet Drake, Leering
// Gargoyle, Goblin Soothsayer and Barreling Attack.

const (
	oldB797Armorer   = "7b75b74c-6003-4b34-bae3-2c9e2f880042"
	oldB797Shaper    = "dfbce321-d3c1-4b76-a3fa-27f875015085"
	oldB797Charm     = "dded2874-e80f-401a-b700-52c54dd6da60"
	oldB797Infantry  = "a6a8861c-b6fe-45a6-86e4-9160c36362ff"
	oldB797Roc       = "acc7d3b8-60f2-4f3b-9cf9-b6b39627d8f6"
	oldB797Wurm      = "d60cbdf0-f20f-4306-be6b-eed90d2a2f77"
	oldB797Drake     = "a82d9c9b-ec17-4f51-a6bf-5984ca4433e5"
	oldB797Gargoyle  = "fc0907eb-a3fc-40ef-950d-b05aee9d5da7"
	oldB797Soothsay  = "a9277df9-1417-45bc-bebf-d9a7595f54c9"
	oldB797Barreling = "f44eab45-8ff9-40eb-9919-28fe5203fc48"
)

func TestOldB797EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB797Armorer: "Armorer Guildmage", oldB797Shaper: "Shaper Guildmage", oldB797Charm: "Ivory Charm",
		oldB797Infantry: "Goblin Elite Infantry", oldB797Roc: "Crimson Roc", oldB797Wurm: "Barbed-Back Wurm",
		oldB797Drake: "Azimaet Drake", oldB797Gargoyle: "Leering Gargoyle", oldB797Soothsay: "Goblin Soothsayer",
		oldB797Barreling: "Barreling Attack",
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

func oldB797Stats(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	return effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id)
}

func TestOldB797ArmorerGuildmageBoostsPowerWithBlackAndToughnessWithGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mage := effectstest.B12Push(g, me.ID, "Armorer Guildmage", "Creature — Human Wizard", oldB797Armorer, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CrtActivateRaw(t, g, me.ID, mage, 0, effectstest.LtCardTarget(bear)); err == nil {
		t.Fatal("activated {B}, {T} with no mana")
	}
	if err := effectstest.CrtActivateRaw(t, g, me.ID, mage, 0, effectstest.LtCardTarget(bear), "B"); err != nil {
		t.Fatalf("activate +1/+0: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := oldB797Stats(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("after the black ability %d/%d, want 3/2", p, tough)
	}
	if !effectstest.PermanentTapped(g, mage) {
		t.Error("the Guildmage did not tap")
	}
	effectstest.P1g3Untap(g, mage)
	if err := effectstest.CrtActivateRaw(t, g, me.ID, mage, 1, effectstest.LtCardTarget(bear), "G"); err != nil {
		t.Fatalf("activate +0/+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := oldB797Stats(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("after the green ability %d/%d, want 3/3", p, tough)
	}
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	if p, tough := oldB797Stats(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("after the turn %d/%d, want 2/2", p, tough)
	}
}

func TestOldB797ShaperGuildmageGrantsFirstStrikeAndPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mage := effectstest.B12Push(g, me.ID, "Shaper Guildmage", "Creature — Human Wizard", oldB797Shaper, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CrtActivateRaw(t, g, me.ID, mage, 0, effectstest.LtCardTarget(bear), "B"); err == nil {
		t.Fatal("paid {W} with black mana")
	}
	if err := effectstest.CrtActivateRaw(t, g, me.ID, mage, 0, effectstest.LtCardTarget(bear), "W"); err != nil {
		t.Fatalf("activate first strike: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("no first strike")
	}
	effectstest.P1g3Untap(g, mage)
	if err := effectstest.CrtActivateRaw(t, g, me.ID, mage, 1, effectstest.LtCardTarget(bear), "B"); err != nil {
		t.Fatalf("activate +1/+0: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := oldB797Stats(t, g, bear); p != 3 {
		t.Errorf("power %d, want 3", p)
	}
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	if effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("first strike outlasted the turn")
	}
}

func TestOldB797IvoryCharmModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 4, 4)

	effectstest.P1g3CastModal(t, g, "Ivory Charm", "Instant", oldB797Charm, []int{0})
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := oldB797Stats(t, g, mine); p != 1 {
		t.Errorf("my power %d, want 1", p)
	}
	if p, _ := oldB797Stats(t, g, theirs); p != 2 {
		t.Errorf("their power %d, want 2", p)
	}

	effectstest.P1g3CastModal(t, g, "Ivory Charm", "Instant", oldB797Charm, []int{1}, effectstest.P1g3ModeCard(theirs, 0))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.PermanentTapped(g, theirs) {
		t.Error("mode two did not tap the creature")
	}

	effectstest.P1g3CastModal(t, g, "Ivory Charm", "Instant", oldB797Charm, []int{2}, effectstest.P1g3ModeCard(mine, 0))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Damage(t, g, theirs, mine, 3)
	if got := effectstest.DamageMarkedOn(g, mine); got != 2 {
		t.Errorf("damage marked %d, want 2 (one prevented of 3)", got)
	}
}

func oldB797Combat(t *testing.T, g *game.Game, attacker uuid.UUID, blockers ...uuid.UUID) {
	t.Helper()
	effectstest.DeclareAttack(t, g, g.Seats[1].ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	for _, b := range blockers {
		if err := g.DeclareBlocker(b, attacker); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB797GoblinEliteInfantryShrinksWhenItBlocksAndWhenBlocked(t *testing.T) {
	// Becomes blocked, by two creatures: one trigger.
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	inf := effectstest.B12Push(g, me.ID, "Goblin Elite Infantry", "Creature — Goblin Warrior", oldB797Infantry, 2, 2)
	b1 := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	b2 := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	oldB797Combat(t, g, inf, b1, b2)
	if p, tough := oldB797Stats(t, g, inf); p != 1 || tough != 1 {
		t.Errorf("blocked by two: %d/%d, want 1/1 (one trigger)", p, tough)
	}

	// Blocks.
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	att := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 0, 9)
	inf = effectstest.B12Push(g, opp.ID, "Goblin Elite Infantry", "Creature — Goblin Warrior", oldB797Infantry, 2, 2)
	oldB797Combat(t, g, att, inf)
	if p, tough := oldB797Stats(t, g, inf); p != 1 || tough != 1 {
		t.Errorf("blocking: %d/%d, want 1/1", p, tough)
	}
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	if p, tough := oldB797Stats(t, g, inf); p != 2 || tough != 2 {
		t.Errorf("after the turn %d/%d, want 2/2", p, tough)
	}
}

func TestOldB797CrimsonRocBoostsOnlyWhenBlockingANonflyer(t *testing.T) {
	for _, flyer := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		var att uuid.UUID
		if flyer {
			att = effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird", Power: 1, Toughness: 4, Keywords: []string{"flying"}, Owner: me.ID, Controller: me.ID})
		} else {
			att = effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 4)
		}
		roc := effectstest.B12Push(g, opp.ID, "Crimson Roc", "Creature — Bird", oldB797Roc, 2, 2)
		oldB797Combat(t, g, att, roc)
		p, _ := oldB797Stats(t, g, roc)
		fs := effectstest.HasEffectiveKeyword(t, g, roc, "first strike")
		if flyer && (p != 2 || fs) {
			t.Errorf("blocking a flyer: power %d first strike %v, want 2 and none", p, fs)
		}
		if !flyer && (p != 3 || !fs) {
			t.Errorf("blocking a nonflyer: power %d first strike %v, want 3 and first strike", p, fs)
		}
	}
}

func TestOldB797BarbedBackWurmShrinksOnlyGreenBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wurm := effectstest.B12Push(g, me.ID, "Barbed-Back Wurm", "Creature — Wurm", oldB797Wurm, 4, 3)
	green := effectstest.B16Creature(g, opp.ID, "Elf", "Creature — Elf", 2, 2, "G")
	red := effectstest.B16Creature(g, opp.ID, "Imp", "Creature — Imp", 2, 2, "R")
	effectstest.CrtBlock(g, green, wurm, opp.ID)
	effectstest.CrtBlock(g, red, wurm, opp.ID)
	if err := effectstest.CrtActivateRaw(t, g, me.ID, wurm, 0, effectstest.LtCardTarget(red), "B"); err == nil {
		t.Error("targeted a blocker that is not green")
	}
	if err := effectstest.CrtActivateRaw(t, g, me.ID, wurm, 0, effectstest.LtCardTarget(green), "B"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := oldB797Stats(t, g, green); p != 1 || tough != 1 {
		t.Errorf("green blocker %d/%d, want 1/1", p, tough)
	}
	bystander := effectstest.B16Creature(g, opp.ID, "Elf", "Creature — Elf", 2, 2, "G")
	if err := effectstest.CrtActivateRaw(t, g, me.ID, wurm, 0, effectstest.LtCardTarget(bystander), "B"); err == nil {
		t.Error("targeted a green creature that is not blocking it")
	}
}

func TestOldB797AzimaetDrakePumpsOnlyOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	drake := effectstest.B12Push(g, me.ID, "Azimaet Drake", "Creature — Drake", oldB797Drake, 1, 3)
	if !effectstest.HasEffectiveKeyword(t, g, drake, "flying") {
		t.Error("no flying")
	}
	if err := effectstest.CrtActivateRaw(t, g, me.ID, drake, 0, nil, "U"); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := oldB797Stats(t, g, drake); p != 2 {
		t.Errorf("power %d, want 2", p)
	}
	if err := effectstest.CrtActivateRaw(t, g, me.ID, drake, 0, nil, "U"); err == nil {
		t.Error("a second activation in the same turn was allowed")
	}
}

func TestOldB797LeeringGargoyleTradesPowerAndFlyingForToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gar := effectstest.B12Push(g, me.ID, "Leering Gargoyle", "Creature — Gargoyle", oldB797Gargoyle, 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, gar, "flying") {
		t.Fatal("no flying")
	}
	if err := effectstest.CrtActivateRaw(t, g, me.ID, gar, 0, nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	p, tough := oldB797Stats(t, g, gar)
	if p != 0 || tough != 4 {
		t.Errorf("%d/%d, want 0/4", p, tough)
	}
	if effectstest.HasEffectiveKeyword(t, g, gar, "flying") {
		t.Error("still flying")
	}
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	if !effectstest.HasEffectiveKeyword(t, g, gar, "flying") {
		t.Error("flying did not come back")
	}
}

func TestOldB797GoblinSoothsayerBoostsRedCreaturesAfterSacrificingAGoblin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sooth := effectstest.B12Push(g, me.ID, "Goblin Soothsayer", "Creature — Goblin Shaman", oldB797Soothsay, 1, 1)
	red := effectstest.B16Creature(g, me.ID, "Imp", "Creature — Imp", 2, 2, "R")
	theirsRed := effectstest.B16Creature(g, opp.ID, "Their Imp", "Creature — Imp", 2, 2, "R")
	green := effectstest.B16Creature(g, me.ID, "Elf", "Creature — Elf", 2, 2, "G")
	gob := effectstest.B16Creature(g, me.ID, "Goblin Token", "Creature — Goblin", 1, 1, "R")
	effectstest.B06AddMana(me, "R")
	if err := g.ActivateCatalogAbility(me.ID, sooth, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{red}}); err == nil {
		t.Error("sacrificed a non-Goblin")
	}
	if err := g.ActivateCatalogAbility(me.ID, sooth, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{gob}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.CardOnBattlefield(g, gob); ok {
		t.Error("a Goblin was not sacrificed")
	}
	if p, tough := oldB797Stats(t, g, red); p != 3 || tough != 3 {
		t.Errorf("my red creature %d/%d, want 3/3", p, tough)
	}
	if p, tough := oldB797Stats(t, g, theirsRed); p != 3 || tough != 3 {
		t.Errorf("their red creature %d/%d, want 3/3", p, tough)
	}
	if p, tough := oldB797Stats(t, g, green); p != 2 || tough != 2 {
		t.Errorf("my green creature %d/%d, want 2/2", p, tough)
	}
}

func TestOldB797BarrelingAttackGrantsTrampleAndGrowsPerBlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	att := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b1 := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	b2 := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	b3 := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	effectstest.CastCatalogSpell(t, g, "Barreling Attack", "Instant", oldB797Barreling, effectstest.LtCardTarget(att))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, att, "trample") {
		t.Fatal("no trample")
	}
	if p, _ := oldB797Stats(t, g, att); p != 2 {
		t.Fatalf("boosted before it was blocked: power %d", p)
	}
	oldB797Combat(t, g, att, b1, b2, b3)
	if p, tough := oldB797Stats(t, g, att); p != 5 || tough != 5 {
		t.Errorf("blocked by three: %d/%d, want 5/5", p, tough)
	}
}

func TestOldB797BarrelingAttackDoesNothingIfTheCreatureIsNeverBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	att := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Barreling Attack", "Instant", oldB797Barreling, effectstest.LtCardTarget(att))
	effectstest.PassPriorityAroundTable(t, g)
	oldB797Combat(t, g, att)
	if p, _ := oldB797Stats(t, g, att); p != 2 {
		t.Errorf("unblocked power %d, want 2", p)
	}
}
