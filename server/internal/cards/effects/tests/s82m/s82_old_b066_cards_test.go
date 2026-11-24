package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b066_cards_test.go — Kaldheim counters (S82, set OLD batch 066).

const (
	oldB066BorealOutrider     = "19ef24b7-25f5-4ae2-b892-3e15ed62d299"
	oldB066ResplendentMarshal = "1fb4c1c1-b10c-463f-a97d-eca9f978e371"
	oldB066CodespellCleric    = "4388af80-118f-4bf4-8fd7-e68c58398ee8"
	oldB066AscendantSpirit    = "784c039d-a9b2-4b22-a937-c8f7b475e96a"
	oldB066InfernalPet        = "831b7dad-ab5a-4980-a52d-65c5633d474a"
	oldB066BloodskyBerserker  = "ed04f170-a4d9-4151-b4ae-1200d523dc6b"
)

func TestOldB066EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB066BorealOutrider:     "Boreal Outrider",
		oldB066ResplendentMarshal: "Resplendent Marshal",
		oldB066CodespellCleric:    "Codespell Cleric",
		oldB066AscendantSpirit:    "Ascendant Spirit",
		oldB066InfernalPet:        "Infernal Pet",
		oldB066BloodskyBerserker:  "Bloodsky Berserker",
	} {
		effectstest.RequireFullNoCaveats(t, oracle, name)
	}
}

func oldB066Snow(g *game.Game, owner uuid.UUID, sub string) uuid.UUID {
	return effectstest.BlbB1Push(g, owner, "Snow-Covered "+sub, "Basic Snow Land — "+sub, "", 0, 0)
}

func oldB066Plain(g *game.Game, owner uuid.UUID, sub string) uuid.UUID {
	return effectstest.BlbB1Push(g, owner, sub, "Basic Land — "+sub, "", 0, 0)
}

// oldB066CastCreature casts a 2/2 creature spell with the given cost and
// colours from the active player's hand, paying with the lands in play.
func oldB066CastCreature(t *testing.T, g *game.Game, me *game.Player, cost string, colors []string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Test Bear", TypeLine: "Creature — Bear", ManaCost: cost,
		Colors: colors, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	effectstest.ToMainForCost(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestOldB066BorealOutriderNeedsSnowManaOfTheSpellsColor(t *testing.T) {
	cases := []struct {
		name  string
		lands func(g *game.Game, me uuid.UUID)
		cost  string
		color []string
		want  int
	}{
		{"snow Forests pay a green spell", func(g *game.Game, me uuid.UUID) {
			oldB066Snow(g, me, "Forest")
			oldB066Snow(g, me, "Forest")
		}, "{1}{G}", []string{"G"}, 1},
		{"plain Forests pay a green spell", func(g *game.Game, me uuid.UUID) {
			oldB066Plain(g, me, "Forest")
			oldB066Plain(g, me, "Forest")
		}, "{1}{G}", []string{"G"}, 0},
		{"snow mana of another colour only pays generic", func(g *game.Game, me uuid.UUID) {
			oldB066Plain(g, me, "Mountain")
			oldB066Snow(g, me, "Forest")
		}, "{1}{R}", []string{"R"}, 0},
		{"one snow mana of the colour is enough", func(g *game.Game, me uuid.UUID) {
			oldB066Snow(g, me, "Mountain")
			oldB066Plain(g, me, "Island")
		}, "{1}{R}", []string{"R"}, 1},
	}
	for _, tc := range cases {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.PushCatalogPermanent(g, me.ID, "Boreal Outrider", "Snow Creature — Elf Warrior", oldB066BorealOutrider, false)
		tc.lands(g, me.ID)
		bear := oldB066CastCreature(t, g, me, tc.cost, tc.color)
		if !effectstest.BattlefieldHas(g, bear) {
			t.Fatalf("%s: the creature did not resolve", tc.name)
		}
		if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != tc.want {
			t.Errorf("%s: %d +1/+1 counters, want %d", tc.name, n, tc.want)
		}
	}
}

func TestOldB066BorealOutriderIgnoresNoncreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Boreal Outrider", "Snow Creature — Elf Warrior", oldB066BorealOutrider, false)
	oldB066Snow(g, me.ID, "Forest")
	oldB066Snow(g, me.ID, "Forest")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Test Rite", TypeLine: "Enchantment", ManaCost: "{1}{G}",
		Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})
	effectstest.ToMainForCost(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, id, game.CounterPlusOne); n != 0 {
		t.Errorf("an enchantment got %d counters", n)
	}
}

func oldB066CastFree(t *testing.T, g *game.Game, c game.Card) uuid.UUID {
	t.Helper()
	id := effectstest.TdCastFromHand(t, g, c, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func oldB066Cleric() game.Card {
	return game.Card{Name: "Codespell Cleric", TypeLine: "Creature — Human Cleric", OracleID: oldB066CodespellCleric, Power: 1, Toughness: 1}
}

func oldB066Bear(name string) game.Card {
	return game.Card{Name: name, TypeLine: "Creature — Bear", Power: 2, Toughness: 2}
}

func oldB066PickCreature(t *testing.T, g *game.Game, chooser, id uuid.UUID) {
	t.Helper()
	for i := 0; i < 4; i++ {
		if p := effectstest.LatestPickTarget(g, chooser); p != nil {
			if err := g.ResolvePickTarget(p.ID, chooser, game.TargetRef{Kind: game.TargetCard, ID: id}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.LatestPickTarget(g, chooser) == nil && effectstest.StackFullyEmpty(g) {
			return
		}
	}
}

func TestOldB066CodespellClericOnlyAsTheSecondSpell(t *testing.T) {
	for _, order := range []int{1, 2, 3} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		pal := effectstest.PushVanillaCreature(g, me.ID, "Pal", 2, 2)
		for i := 1; i < order; i++ {
			oldB066CastFree(t, g, oldB066Bear("Filler"))
		}
		cleric := oldB066CastFree(t, g, oldB066Cleric())
		if order == 2 {
			oldB066PickCreature(t, g, me.ID, pal)
		} else if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			t.Fatalf("order %d: the entry trigger asked for a target", order)
		}
		if !effectstest.BattlefieldHas(g, cleric) {
			t.Fatalf("order %d: the Cleric is not on the battlefield", order)
		}
		want := 0
		if order == 2 {
			want = 1
		}
		if n := effectstest.CountersOn(g, pal, game.CounterPlusOne); n != want {
			t.Errorf("order %d: %d counters on the target, want %d", order, n, want)
		}
	}
}

func TestOldB066CodespellClericNotCastDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oldB066CastFree(t, g, oldB066Bear("Filler"))
	effectstest.PushVanillaCreature(g, me.ID, "Pal", 2, 2)
	effectstest.PushCatalogPermanent(g, me.ID, "Codespell Cleric", "Creature — Human Cleric", oldB066CodespellCleric, false)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Error("a Cleric that was never cast triggered")
	}
}

func oldB066Elf(name string) game.Card {
	return game.Card{Name: name, TypeLine: "Creature — Elf Warrior", Power: 2, Toughness: 2}
}

func TestOldB066ResplendentMarshalEntersAndExilesForCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	grave := effectstest.PushGraveyardCardTyped(me, "Dead Elf", "Creature — Elf Druid")
	elf := effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf Warrior", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Their Elf", "Creature — Elf", 2, 2)
	marshal := effectstest.TdCastFromHand(t, g, game.Card{Name: "Resplendent Marshal", TypeLine: "Creature — Angel Warrior",
		OracleID: oldB066ResplendentMarshal, Power: 3, Toughness: 3}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choice to exile a creature card")
	}
	if effectstest.HasID(pick.ChooseCards, marshal) {
		t.Fatal("the Marshal itself is offered")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, grave)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(grave) {
		t.Error("the chosen card was not exiled")
	}
	if n := effectstest.CountersOn(g, elf, game.CounterPlusOne); n != 1 {
		t.Errorf("the Elf has %d counters, want 1", n)
	}
	for name, id := range map[string]uuid.UUID{"the Bear": bear, "an opponent's Elf": theirs, "the Marshal": marshal} {
		if n := effectstest.CountersOn(g, id, game.CounterPlusOne); n != 0 {
			t.Errorf("%s has %d counters, want 0", name, n)
		}
	}
}

func TestOldB066ResplendentMarshalMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	grave := effectstest.PushGraveyardCardTyped(me, "Dead Elf", "Creature — Elf Druid")
	elf := effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf Warrior", 2, 2)
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Resplendent Marshal", TypeLine: "Creature — Angel Warrior",
		OracleID: oldB066ResplendentMarshal, Power: 3, Toughness: 3}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(grave) || effectstest.CountersOn(g, elf, game.CounterPlusOne) != 0 {
		t.Error("declining still exiled or pumped")
	}
}

func TestOldB066ResplendentMarshalDiesToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	grave := effectstest.PushGraveyardCardTyped(me, "Dead Warrior", "Creature — Human Warrior")
	friend := effectstest.B12Creature(g, me.ID, "Friend", "Creature — Human Soldier", 2, 2)
	marshal := effectstest.B12Push(g, me.ID, "Resplendent Marshal", "Creature — Angel Warrior", oldB066ResplendentMarshal, 3, 3)
	effectstest.P1Destroy(t, g, marshal)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("the dies half did not ask")
	}
	if effectstest.HasID(pick.ChooseCards, marshal) || !effectstest.HasID(pick.ChooseCards, grave) {
		t.Fatalf("offered %v, want only the other creature card", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, grave)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, friend, game.CounterPlusOne); n != 1 {
		t.Errorf("a Human Soldier got %d counters from an exiled Human Warrior, want 1", n)
	}
}

func TestOldB066InfernalPetAndBerserkerWantTheSecondSpell(t *testing.T) {
	for _, kind := range []string{"pet", "berserker"} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		var id uuid.UUID
		counters, keyword := 1, "flying"
		if kind == "pet" {
			id = effectstest.B12Push(g, me.ID, "Infernal Pet", "Creature — Imp", oldB066InfernalPet, 2, 2)
		} else {
			id = effectstest.B12Push(g, me.ID, "Bloodsky Berserker", "Creature — Human Berserker", oldB066BloodskyBerserker, 1, 1)
			counters, keyword = 2, "menace"
		}
		oldB066CastFree(t, g, oldB066Bear("First"))
		if n := effectstest.CountersOn(g, id, game.CounterPlusOne); n != 0 || effectstest.HasKeywordOnBattlefield(t, g, id, keyword) {
			t.Fatalf("%s: the first spell triggered it", kind)
		}
		oldB066CastFree(t, g, oldB066Bear("Second"))
		if n := effectstest.CountersOn(g, id, game.CounterPlusOne); n != counters {
			t.Errorf("%s: %d counters after the second spell, want %d", kind, n, counters)
		}
		if !effectstest.HasKeywordOnBattlefield(t, g, id, keyword) {
			t.Errorf("%s: no %s until end of turn", kind, keyword)
		}
		oldB066CastFree(t, g, oldB066Bear("Third"))
		if n := effectstest.CountersOn(g, id, game.CounterPlusOne); n != counters {
			t.Errorf("%s: the third spell added counters (%d)", kind, n)
		}
		effectstest.AdvanceToNextSeatsTurn(t, g)
		if effectstest.HasKeywordOnBattlefield(t, g, id, keyword) {
			t.Errorf("%s: %s lasted past the turn", kind, keyword)
		}
	}
}

func oldB066Activate(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, idx, snow int) error {
	t.Helper()
	for i := 0; i < snow; i++ {
		oldB066Snow(g, me.ID, "Island")
	}
	err := g.ActivateCatalogAbility(me.ID, id, idx, game.ActivateAbilityParams{Strict: true, AutoTap: true})
	if err == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	return err
}

func TestOldB066AscendantSpiritClimbsThroughItsStages(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.BlbB1Push(g, me.ID, "Ascendant Spirit", "Snow Creature — Spirit", oldB066AscendantSpirit, 1, 1)
	effectstest.AdvanceToMain(t, g)
	// Plain lands cannot pay {S}.
	oldB066Plain(g, me.ID, "Island")
	oldB066Plain(g, me.ID, "Island")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err == nil {
		t.Fatal("{S}{S} was paid with nonsnow lands")
	}
	// The later stages do nothing before their type is there.
	if err := oldB066Activate(t, g, me, id, 1, 3); err != nil {
		t.Fatalf("stage 2: %v", err)
	}
	if p, tough := effectstest.PtOf(t, g, id); p != 1 || tough != 1 || effectstest.CountersOn(g, id, game.CounterFlying) != 0 {
		t.Fatalf("a non-Warrior stage 2 did something: %d/%d", p, tough)
	}
	if err := oldB066Activate(t, g, me, id, 2, 4); err != nil {
		t.Fatalf("stage 3: %v", err)
	}
	if p, tough := effectstest.PtOf(t, g, id); p != 1 || tough != 1 {
		t.Fatalf("a non-Angel stage 3 did something: %d/%d", p, tough)
	}
	if err := oldB066Activate(t, g, me, id, 0, 2); err != nil {
		t.Fatalf("stage 1: %v", err)
	}
	c, _ := effectstest.BattlefieldCard(g, id)
	if p, tough := effectstest.PtOf(t, g, id); p != 2 || tough != 3 || !c.HasSubtype("Spirit") || !c.HasSubtype("Warrior") {
		t.Fatalf("stage 1: %d/%d %v", p, tough, c.Effective().Subtypes)
	}
	if err := oldB066Activate(t, g, me, id, 1, 3); err != nil {
		t.Fatalf("stage 2: %v", err)
	}
	c, _ = effectstest.BattlefieldCard(g, id)
	if p, tough := effectstest.PtOf(t, g, id); p != 4 || tough != 4 || !c.HasSubtype("Angel") {
		t.Fatalf("stage 2: %d/%d %v", p, tough, c.Effective().Subtypes)
	}
	if effectstest.CountersOn(g, id, game.CounterFlying) != 1 || !effectstest.HasKeywordOnBattlefield(t, g, id, "flying") {
		t.Error("no flying counter or no flying after stage 2")
	}
	if err := oldB066Activate(t, g, me, id, 2, 4); err != nil {
		t.Fatalf("stage 3: %v", err)
	}
	if p, tough := effectstest.PtOf(t, g, id); p != 6 || tough != 6 {
		t.Errorf("stage 3: %d/%d, want 6/6 (4/4 with two counters)", p, tough)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if p, _ := effectstest.PtOf(t, g, id); p != 6 {
		t.Error("the changes ended with the turn")
	}
}

func TestOldB066AscendantSpiritDrawsOnCombatDamageAfterStageThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.BlbB1Push(g, me.ID, "Ascendant Spirit", "Snow Creature — Spirit", oldB066AscendantSpirit, 1, 1)
	effectstest.AdvanceToMain(t, g)
	for _, step := range []struct{ idx, snow int }{{0, 2}, {1, 3}, {2, 4}} {
		if err := oldB066Activate(t, g, me, id, step.idx, step.snow); err != nil {
			t.Fatalf("stage %d: %v", step.idx, err)
		}
	}
	for i := 0; i < 6; i++ {
		effectstest.PushVanillaCreature(g, me.ID, "Padding", 1, 1)
	}
	before := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, id)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d -> %d, want one card drawn", before, got)
	}
}
