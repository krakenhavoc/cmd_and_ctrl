package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b830_cards_test.go — Ice Age, batch 830 (S82 OLD, #5636): Auras.
// Every test plays the real catalog entry through the engine.

const (
	oldB830Conquer    = "be15e131-0753-4ceb-a36b-4f9172241080"
	oldB830Green      = "c5ca2df3-6fd9-478d-9bc6-48695a4600fb"
	oldB830Rite       = "ced4538a-4919-4ef8-8bf2-2de9e0388587"
	oldB830Lore       = "d2853bad-a751-45ee-8269-24b0cdbc5c6b"
	oldB830Fylgja     = "df0e5dd8-65e0-474c-ac32-6e0f6f16b8e6"
	oldB830MindWhip   = "e26f7a58-18ae-432b-bc3f-c410512eb6bf"
	oldB830SnowDevil  = "e743b4ec-65bf-4e8d-ad87-6746a2e6ba73"
	oldB830White      = "ec9c3906-8628-48c6-96a7-8c6748cba816"
	oldB830Errantry   = "fee455f2-fcc4-47b1-a533-92fa1ae8ae63"
	oldB830Bear       = "Creature — Bear"
	oldB830ForestType = "Basic Land — Forest"
)

func TestOldB830EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB830Conquer: "Conquer", oldB830Green: "Green Scarab", oldB830Rite: "Leshrac's Rite",
		oldB830Lore: "Forbidden Lore", oldB830Fylgja: "Fylgja", oldB830MindWhip: "Mind Whip",
		oldB830SnowDevil: "Snow Devil", oldB830White: "White Scarab", oldB830Errantry: "Errantry",
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

func oldB830ToUpkeepOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < 300; i++ {
		if g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat == seat && i > 0 {
			break
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}

func oldB830ScarabCase(t *testing.T, name, oracle, color string) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB830Bear, 2, 2)
	effectstest.Enchant(t, g, name, oracle, bear)
	if p, _ := effectstest.PtOf(t, g, bear); p != 2 {
		t.Fatalf("with no %s permanent the Bear has power %d, want 2", color, p)
	}
	mine := effectstest.B12Creature(g, me.ID, "Own", oldB830Bear, 1, 1)
	effectstest.AlliSetColors(g, mine, color)
	if p, _ := effectstest.PtOf(t, g, bear); p != 2 {
		t.Fatalf("the Aura's controller's own permanent pumped the Bear: power %d", p)
	}
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", oldB830Bear, 1, 1)
	effectstest.AlliSetColors(g, theirs, color)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
		t.Fatalf("with an opponent's permanent the Bear is %d/%d, want 4/4", p, tough)
	}
	other := effectstest.B12Creature(g, opp.ID, "Other", oldB830Bear, 2, 2)
	effectstest.AlliSetColors(g, other, "B")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(theirs, bear); err == nil {
		t.Errorf("a %s creature blocked the enchanted creature", color)
	}
	if err := g.DeclareBlocker(other, bear); err != nil {
		t.Errorf("a black creature could not block it: %v", err)
	}
}

func TestOldB830GreenScarab(t *testing.T) { oldB830ScarabCase(t, "Green Scarab", oldB830Green, "G") }
func TestOldB830WhiteScarab(t *testing.T) { oldB830ScarabCase(t, "White Scarab", oldB830White, "W") }

func TestOldB830ConquerTakesControlOfALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forest := effectstest.SeedLand(g, opp.ID, "Forest", oldB830ForestType, "")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Conquer", effectstest.AuraTypeLine, oldB830Conquer, effectstest.CardTargets(effectstest.B12Creature(g, opp.ID, "Bear", oldB830Bear, 2, 2))); err == nil {
		t.Fatal("Conquer enchanted a creature")
	}
	effectstest.Enchant(t, g, "Conquer", oldB830Conquer, forest)
	if c, ok := effectstest.BattlefieldCard(g, forest); !ok || c.Controller != me.ID {
		t.Fatalf("the Forest is controlled by %v, want the Aura's controller", c.Controller)
	}
}

func TestOldB830LeshracsRiteGrantsSwampwalk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB830Bear, 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, bear, "swampwalk") {
		t.Fatal("swampwalk before the Aura")
	}
	effectstest.Enchant(t, g, "Leshrac's Rite", oldB830Rite, bear)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "swampwalk") {
		t.Error("the enchanted creature lacks swampwalk")
	}
}

func TestOldB830ForbiddenLoreGivesTheLandATapPump(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := effectstest.SeedLand(g, me.ID, "Forest", oldB830ForestType, "")
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB830Bear, 2, 2)
	effectstest.Enchant(t, g, "Forbidden Lore", oldB830Lore, forest)
	effectstest.AdvanceToMain(t, g)
	idx, ref := effectstest.GrantedActivatedIndex(t, g, forest)
	if idx < 0 {
		t.Fatal("the land has no granted ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, forest, idx, game.ActivateAbilityParams{Strict: true, Ref: ref, Targets: effectstest.CardTargets(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 3 {
		t.Errorf("the Bear is %d/%d, want 4/3", p, tough)
	}
	if !effectstest.B16Tapped(t, g, forest) {
		t.Error("the land was not tapped for the cost")
	}
}

func TestOldB830FylgjaShieldsTheEnchantedCreatureFromOneDamageAtATime(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB830Bear, 2, 5)
	aura := effectstest.Enchant(t, g, "Fylgja", oldB830Fylgja, bear)
	if c, _ := effectstest.BattlefieldCard(g, aura); c.Counters["healing"] != 4 {
		t.Fatalf("Fylgja entered with %d healing counters, want 4", c.Counters["healing"])
	}
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("remove a counter: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, aura); c.Counters["healing"] != 3 {
		t.Errorf("after the ability Fylgja has %d healing counters, want 3", c.Counters["healing"])
	}
	bolt := effectstest.PutInHand(me, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", OracleID: effectstest.LightningBoltOracle, Colors: []string{"R"}, ManaCost: "{R}"})
	effectstest.FloatMana(t, g, me, "{R}")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("cast Bolt: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, bear); got != 2 {
		t.Errorf("the Bear has %d damage, want 2 (3 less the 1 prevented)", got)
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{W}")
	if err := g.ActivateCatalogAbility(me.ID, aura, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("put a counter: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, aura); c.Counters["healing"] != 4 {
		t.Errorf("after {2}{W} Fylgja has %d healing counters, want 4", c.Counters["healing"])
	}
}

func TestOldB830MindWhipDamagesAndTapsUnlessThePlayerPays(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[1]
		bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB830Bear, 2, 2)
		effectstest.Enchant(t, g, "Mind Whip", oldB830MindWhip, bear)
		life := effectstest.LifeOf(g, opp.ID)
		oldB830ToUpkeepOf(t, g, 1)
		effectstest.PassPriorityAroundTable(t, g)
		if pay {
			effectstest.FillPool(opp, 3)
		}
		effectstest.AnswerPayUnless(t, g, opp.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		wantLife, wantTapped := life-2, true
		if pay {
			wantLife, wantTapped = life, false
		}
		if got := effectstest.LifeOf(g, opp.ID); got != wantLife {
			t.Errorf("pay=%v: life %d, want %d", pay, got, wantLife)
		}
		if got := effectstest.B16Tapped(t, g, bear); got != wantTapped {
			t.Errorf("pay=%v: tapped = %v, want %v", pay, got, wantTapped)
		}
	}
}

func TestOldB830SnowDevilGivesFlyingAndFirstStrikeWhileBlockingWithASnowLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB830Bear, 2, 2)
	effectstest.Enchant(t, g, "Snow Devil", oldB830SnowDevil, bear)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Fatal("the enchanted creature lacks flying")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Fatal("first strike while not blocking")
	}
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", oldB830Bear, 1, 1)
	effectstest.SeedLand(g, me.ID, "Snow-Covered Forest", "Basic Snow Land — Forest", "")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(bear, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("a blocking enchanted creature lacks first strike while the Aura's controller has a snow land")
	}
}

func TestOldB830SnowDevilNeedsASnowLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB830Bear, 2, 2)
	effectstest.Enchant(t, g, "Snow Devil", oldB830SnowDevil, bear)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", oldB830Bear, 1, 1)
	effectstest.SeedLand(g, me.ID, "Forest", oldB830ForestType, "")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(bear, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	if effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("first strike without a snow land")
	}
}

func TestOldB830ErrantryPumpsAndLimitsTheCreatureToAttackingAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB830Bear, 2, 2)
	friend := effectstest.B12Creature(g, me.ID, "Friend", oldB830Bear, 2, 2)
	effectstest.Enchant(t, g, "Errantry", oldB830Errantry, bear)
	if p, _ := effectstest.PtOf(t, g, bear); p != 5 {
		t.Fatalf("the Bear has power %d, want 5", p)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("the enchanted creature could not attack alone: %v", err)
	}
	if err := g.DeclareAttacker(friend, opp.ID); err == nil {
		t.Error("another creature attacked beside the enchanted one")
	}
}

func TestOldB830ErrantryLetsOtherCreaturesAttackWhenTheEnchantedOneStaysHome(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB830Bear, 2, 2)
	friend := effectstest.B12Creature(g, me.ID, "Friend", oldB830Bear, 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other", oldB830Bear, 2, 2)
	effectstest.Enchant(t, g, "Errantry", oldB830Errantry, bear)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, c := range []uuid.UUID{friend, other} {
		if err := g.DeclareAttacker(c, opp.ID); err != nil {
			t.Fatalf("an unenchanted creature could not attack: %v", err)
		}
	}
	if err := g.DeclareAttacker(bear, opp.ID); err == nil {
		t.Error("the enchanted creature joined an attack")
	}
}
