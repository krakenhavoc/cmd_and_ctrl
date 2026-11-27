package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b756_cards_test.go — Tempest, batch 756 (S82 older sets, #5628):
// pump and combat tricks. Every test plays the real catalog entry through the
// engine.

const (
	oldB756Barbed   = "8fbca9ba-8b83-40ca-9f7e-5c483ba76675"
	oldB756Drake    = "93b6a170-a816-4556-aa27-a729231715e7"
	oldB756Warlord  = "ae63849d-7352-4a2b-9d64-f3005146bd22"
	oldB756Dread    = "b431a73b-cf4c-4154-810f-82b7c6099e66"
	oldB756Armor    = "c284b1fd-f238-478a-bcf1-c4b8f32f11f3"
	oldB756Armodon  = "ce7d672a-7cb9-40e5-a8ad-dd3ce0f8ecd3"
	oldB756Giant    = "f9558fa5-2513-47e0-9af9-85db5cae4146"
	oldB756Auratog  = "fe47535e-e50c-49ad-beda-46750d33e6f4"
	oldB756BearType = "Creature — Bear"
)

func TestOldB756EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		oldB756Barbed, oldB756Drake, oldB756Warlord, oldB756Dread, oldB756Armor,
		oldB756Armodon, oldB756Giant, oldB756Auratog,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func oldB756Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	return g, me, opp
}

func oldB756Activate(g *game.Game, me *game.Player, id uuid.UUID, params game.ActivateAbilityParams) error {
	params.Strict = true
	return g.ActivateCatalogAbility(me.ID, id, 0, params)
}

func oldB756PT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	return effectstest.EffectivePTOf(t, g, id)
}

func TestOldB756SliversGiveEverySliverAPump(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		power, tough int
	}{
		{"Barbed Sliver", oldB756Barbed, 1, 0},
		{"Armor Sliver", oldB756Armor, 0, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp := oldB756Table(t)
			effectstest.B12Push(g, me.ID, tc.name, "Creature — Sliver", tc.oracle, 2, 2)
			mine := effectstest.TspB1Sliver(g, me.ID, "My Sliver", 1, 1)
			theirs := effectstest.TspB1Sliver(g, opp.ID, "Their Sliver", 1, 1)
			bear := effectstest.B12Creature(g, me.ID, "Bear", oldB756BearType, 2, 2)
			if effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, bear), "{2}") {
				t.Fatal("a non-Sliver was granted the ability")
			}
			effectstest.FloatMana(t, g, me, "{C}{C}")
			effectstest.TspB1ActivateGranted(t, g, me.ID, mine, nil)
			effectstest.PassPriorityAroundTable(t, g)
			if p, tough := oldB756PT(t, g, mine); p != 1+tc.power || tough != 1+tc.tough {
				t.Errorf("my Sliver is %d/%d, want %d/%d", p, tough, 1+tc.power, 1+tc.tough)
			}
			if p, tough := oldB756PT(t, g, theirs); p != 1 || tough != 1 {
				t.Errorf("their Sliver is %d/%d, want 1/1 (only its controller can pay)", p, tough)
			}
		})
	}
}

func TestOldB756DreadOfNightShrinksEveryWhiteCreature(t *testing.T) {
	g, me, opp := oldB756Table(t)
	effectstest.B12Push(g, me.ID, "Dread of Night", "Enchantment", oldB756Dread, 0, 0)
	white := func(owner uuid.UUID, name string) uuid.UUID {
		return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: name, TypeLine: oldB756BearType, Power: 3, Toughness: 3,
			Colors: []string{"W"}, Owner: owner, Controller: owner,
		})
	}
	mine, theirs := white(me.ID, "My White Bear"), white(opp.ID, "Their White Bear")
	green := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Green Bear", TypeLine: oldB756BearType, Power: 3, Toughness: 3,
		Colors: []string{"G"}, Owner: opp.ID, Controller: opp.ID,
	})
	for name, tc := range map[string]struct {
		id   uuid.UUID
		want int
	}{"my white creature": {mine, 2}, "their white creature": {theirs, 2}, "a green creature": {green, 3}} {
		if p, tough := oldB756PT(t, g, tc.id); p != tc.want || tough != tc.want {
			t.Errorf("%s is %d/%d, want %d/%d", name, p, tough, tc.want, tc.want)
		}
	}
	weak := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "White Mouse", TypeLine: "Creature — Mouse", Power: 1, Toughness: 1,
		Colors: []string{"W"}, Owner: opp.ID, Controller: opp.ID,
	})
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(weak) {
		t.Error("a white 1/1 survived at 0 toughness")
	}
}

func TestOldB756DrakeGivesTheCreatureItBlocksOrIsBlockedByAPump(t *testing.T) {
	g, me, opp := oldB756Table(t)
	drake := effectstest.B12Push(g, me.ID, "Flailing Drake", "Creature — Drake", oldB756Drake, 2, 3)
	attacker := effectstest.B12Creature(g, opp.ID, "Attacker", oldB756BearType, 2, 2)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", oldB756BearType, 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", oldB756BearType, 2, 2)
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, drake), "flying") {
		t.Error("the Drake does not fly")
	}
	effectstest.WoeS2Block(g, me.ID, drake, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.WoeS2Block(g, opp.ID, blocker, drake)
	effectstest.PassPriorityAroundTable(t, g)
	for name, tc := range map[string]struct {
		id   uuid.UUID
		want int
	}{"the creature it blocked": {attacker, 3}, "the creature that blocked it": {blocker, 3}, "a bystander": {bystander, 2}, "the Drake": {drake, 2}} {
		if p, tough := oldB756PT(t, g, tc.id); p != tc.want || tough != tc.want+oldB756BoolToInt(tc.id == drake) {
			t.Errorf("%s is %d/%d, want %d power", name, p, tough, tc.want)
		}
	}
}

func oldB756BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func TestOldB756WarlordGivesEachOtherAttackerPower(t *testing.T) {
	g, me, _ := oldB756Table(t)
	warlord := effectstest.B12Push(g, me.ID, "Renegade Warlord", "Creature — Human Soldier", oldB756Warlord, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Attacking Bear", oldB756BearType, 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle Bear", oldB756BearType, 2, 2)
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, warlord), "first strike") {
		t.Error("the Warlord lacks first strike")
	}
	effectstest.AttackAndTrigger(t, g, warlord, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := oldB756PT(t, g, bear); p != 3 {
		t.Errorf("the other attacker has %d power, want 3", p)
	}
	if p, _ := oldB756PT(t, g, warlord); p != 3 {
		t.Errorf("the Warlord has %d power, want 3 (it gives the bonus to the others)", p)
	}
	if p, _ := oldB756PT(t, g, idle); p != 2 {
		t.Errorf("a creature that did not attack has %d power, want 2", p)
	}
}

func TestOldB756ArmodonRagesOncePerTurnThenDiesAtTheEndStep(t *testing.T) {
	g, me, _ := oldB756Table(t)
	armodon := effectstest.B12Push(g, me.ID, "Crazed Armodon", "Creature — Elephant", oldB756Armodon, 3, 3)
	effectstest.FloatForTest(g, me, "GG")
	if err := oldB756Activate(g, me, armodon, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if err := oldB756Activate(g, me, armodon, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("a second activation in one turn was accepted")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := oldB756PT(t, g, armodon); p != 6 {
		t.Errorf("the Armodon has %d power, want 6", p)
	}
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, armodon), "trample") {
		t.Error("the Armodon did not gain trample")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(armodon) {
		t.Error("the Armodon was not destroyed at the beginning of the end step")
	}
}

func TestOldB756GiantTradesToughnessForPower(t *testing.T) {
	g, me, _ := oldB756Table(t)
	giant := effectstest.B12Push(g, me.ID, "Flowstone Giant", "Creature — Giant", oldB756Giant, 3, 3)
	effectstest.FloatForTest(g, me, "R")
	if err := oldB756Activate(g, me, giant, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := oldB756PT(t, g, giant); p != 5 || tough != 1 {
		t.Errorf("the Giant is %d/%d, want 5/1", p, tough)
	}
	effectstest.FloatForTest(g, me, "R")
	if err := oldB756Activate(g, me, giant, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate again: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(giant) {
		t.Error("the Giant survived at -1 toughness")
	}
}

func TestOldB756AuratogSacrificesAnEnchantment(t *testing.T) {
	g, me, _ := oldB756Table(t)
	atog := effectstest.B12Push(g, me.ID, "Auratog", "Creature — Atog", oldB756Auratog, 1, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB756BearType, 2, 2)
	if err := oldB756Activate(g, me, atog, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("the Auratog sacrificed a creature")
	}
	ench := effectstest.B12Permanent(g, me.ID, "Pacifism-like", "Enchantment")
	if err := oldB756Activate(g, me, atog, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{ench}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) {
		t.Error("the enchantment was not sacrificed")
	}
	if p, tough := oldB756PT(t, g, atog); p != 3 || tough != 4 {
		t.Errorf("the Auratog is %d/%d, want 3/4", p, tough)
	}
}
