package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b410_test.go — Gatecrash, batch 410 (#5565): each test plays the real
// catalog entry through the real cast, activation and trigger paths.

const (
	oldB410Specialist = "3c334eb3-e3fd-45b0-a36d-4681c9e72ac5"
	oldB410Illness    = "51abc95c-bd92-46f7-a0a1-6e9f8e0dda87"
	oldB410HoldGates  = "7e6fcbc0-691a-4bd7-b063-debc86f4506a"
	oldB410Smog       = "8e51a07d-8618-4908-8798-bd28bc9b5bbb"
	oldB410Swing      = "bd9ab645-d8ed-4cdd-8d8e-53383b5501ba"
	oldB410Aerial     = "d4ccc319-db23-4ef4-8129-317426f3a794"
	oldB410Furious    = "e69a6fee-c4c0-4b31-8899-6e0935a6cded"
	oldB410Sunhome    = "f211d21f-7d78-4435-a592-a6896023af3b"
	oldB410Shade      = "ff0647bc-87d1-45f0-ad5c-63c29174f404"
)

func TestOldB410CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB410Specialist: "Incursion Specialist",
		oldB410Illness:    "Illness in the Ranks",
		oldB410HoldGates:  "Hold the Gates",
		oldB410Smog:       "Smog Elemental",
		oldB410Swing:      "Executioner's Swing",
		oldB410Aerial:     "Aerial Maneuver",
		oldB410Furious:    "Furious Resistance",
		oldB410Sunhome:    "Sunhome Guildmage",
		oldB410Shade:      "Gateway Shade",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

func oldB410Bear(name string) game.Card {
	return game.Card{Name: name, TypeLine: "Creature — Bear", Power: 2, Toughness: 2}
}

func oldB410CastFree(t *testing.T, g *game.Game, c game.Card) {
	t.Helper()
	effectstest.TdCastFromHand(t, g, c, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
}

// --- Incursion Specialist ---

func TestOldB410IncursionSpecialistWantsTheSecondSpellEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, "Incursion Specialist", "Creature — Human Wizard", oldB410Specialist, 1, 3)
	oldB410CastFree(t, g, oldB410Bear("First"))
	if p := effectstest.EffectivePower(t, g, id); p != 1 {
		t.Fatalf("power %d after the first spell, want 1", p)
	}
	if effectstest.EffectiveOf(t, g, id).Restrictions&game.CantBeBlocked != 0 {
		t.Fatal("unblockable after the first spell")
	}
	oldB410CastFree(t, g, oldB410Bear("Second"))
	if p := effectstest.EffectivePower(t, g, id); p != 3 {
		t.Errorf("power %d after the second spell, want 3", p)
	}
	if effectstest.EffectiveOf(t, g, id).Restrictions&game.CantBeBlocked == 0 {
		t.Error("not unblockable after the second spell")
	}
	oldB410CastFree(t, g, oldB410Bear("Third"))
	if p := effectstest.EffectivePower(t, g, id); p != 3 {
		t.Errorf("power %d after the third spell, want 3", p)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if p := effectstest.EffectivePower(t, g, id); p != 1 {
		t.Errorf("power %d next turn, want 1", p)
	}
	if effectstest.EffectiveOf(t, g, id).Restrictions&game.CantBeBlocked != 0 {
		t.Error("still unblockable next turn")
	}
}

// --- Illness in the Ranks ---

func TestOldB410IllnessInTheRanksShrinksEveryCreatureToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Illness in the Ranks", "Enchantment", oldB410Illness, 0, 0)
	mine := effectstest.B12Creature(g, me.ID, "My Token", "Token Creature — Soldier", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Token", "Token Creature — Soldier", 3, 3)
	card := effectstest.B12Creature(g, opp.ID, "Real Bear", "Creature — Bear", 3, 3)
	for id, want := range map[uuid.UUID]int{mine: 2, theirs: 2, card: 3} {
		if p, tg := effectstest.EffectivePTOf(t, g, id); p != want || tg != want {
			t.Errorf("%s is %d/%d, want %d/%d", id, p, tg, want, want)
		}
	}
}

// --- Hold the Gates ---

func TestOldB410HoldTheGatesCountsGatesAndGrantsVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Hold the Gates", "Enchantment", oldB410HoldGates, 0, 0)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	if _, tg := effectstest.EffectivePTOf(t, g, mine); tg != 2 {
		t.Errorf("toughness %d with no Gates, want 2", tg)
	}
	effectstest.B12Permanent(g, me.ID, "Gate A", "Land — Gate")
	effectstest.B12Permanent(g, me.ID, "Gate B", "Land — Gate")
	effectstest.B12Permanent(g, opp.ID, "Their Gate", "Land — Gate")
	effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	if p, tg := effectstest.EffectivePTOf(t, g, mine); p != 2 || tg != 4 {
		t.Errorf("mine is %d/%d with two Gates, want 2/4", p, tg)
	}
	if _, tg := effectstest.EffectivePTOf(t, g, theirs); tg != 2 {
		t.Errorf("their creature got +0/+%d", tg-2)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, mine, "vigilance") {
		t.Error("my creature has no vigilance")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, theirs, "vigilance") {
		t.Error("an opponent's creature has vigilance")
	}
}

// --- Smog Elemental ---

func TestOldB410SmogElementalShrinksOnlyOpposingFliers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	smog := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Smog Elemental", OracleID: oldB410Smog,
		TypeLine: "Creature — Elemental", Power: 3, Toughness: 3, Keywords: []string{"flying"}})
	mine := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "My Bird", TypeLine: "Creature — Bird", Power: 2, Toughness: 2, Keywords: []string{"flying"}})
	bird := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Their Bird", TypeLine: "Creature — Bird", Power: 2, Toughness: 2, Keywords: []string{"flying"}})
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	for id, want := range map[uuid.UUID]int{smog: 3, mine: 2, bird: 1, bear: 2} {
		if p, tg := effectstest.EffectivePTOf(t, g, id); p != want || tg != want {
			t.Errorf("%s is %d/%d, want %d/%d", id, p, tg, want, want)
		}
	}
}

// --- Executioner's Swing ---

func TestOldB410ExecutionersSwingNeedsACreatureThatDealtDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Giant", 6, 6)
	hitter := effectstest.B12Creature(g, opp.ID, "Hitter", "Creature — Giant", 6, 6)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Executioner's Swing", "Instant", oldB410Swing, effectstest.CardRefs(idle)); err == nil {
		t.Fatal("targeted a creature that has not dealt damage")
	}
	effectstest.DealNoncombatDamage(g, opp.ID, hitter, g.Seats[0].ID)
	effectstest.CastCatalogSpell(t, g, "Executioner's Swing", "Instant", oldB410Swing, effectstest.CardRefs(hitter))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.EffectivePTOf(t, g, hitter); p != 1 || tg != 1 {
		t.Errorf("Hitter is %d/%d, want 1/1", p, tg)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if p, tg := effectstest.EffectivePTOf(t, g, hitter); p != 6 || tg != 6 {
		t.Errorf("Hitter is %d/%d next turn, want 6/6", p, tg)
	}
}

func TestOldB410ExecutionersSwingKillsAFragileCreatureThatDealtDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	hitter := effectstest.B12Creature(g, opp.ID, "Hitter", "Creature — Elf", 2, 5)
	effectstest.DealNoncombatDamage(g, opp.ID, hitter, g.Seats[0].ID)
	effectstest.CastCatalogSpell(t, g, "Executioner's Swing", "Instant", oldB410Swing, effectstest.CardRefs(hitter))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(hitter) {
		t.Error("a 2/5 with -5/-5 survived")
	}
}

// --- Aerial Maneuver ---

func TestOldB410AerialManeuverPumpsAndGrantsFlyingAndFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Aerial Maneuver", "Instant", oldB410Aerial, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.EffectivePTOf(t, g, bear); p != 3 || tg != 3 {
		t.Errorf("Bear is %d/%d, want 3/3", p, tg)
	}
	for _, kw := range []string{"flying", "first strike"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, bear, kw) {
			t.Errorf("Bear lacks %s", kw)
		}
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Error("flying lasted past the turn")
	}
}

// --- Furious Resistance ---

func TestOldB410FuriousResistanceNeedsABlockingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := effectstest.B12Creature(g, me.ID, "Walker", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 3)
	idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Bear", 2, 3)
	effectstest.BrAttack(t, g, walker)
	if err := g.DeclareBlocker(blocker, walker); err != nil {
		t.Fatalf("block: %v", err)
	}
	spell := effectstest.AvB24Hand(g, opp, game.Card{Name: "Furious Resistance", TypeLine: "Instant", OracleID: oldB410Furious})
	if err := g.CastSpell(opp.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(idle)}); err == nil {
		t.Fatal("targeted a creature that is not blocking")
	}
	if err := g.CastSpell(opp.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(blocker)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.EffectivePTOf(t, g, blocker); p != 5 || tg != 3 {
		t.Errorf("Blocker is %d/%d, want 5/3", p, tg)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, blocker, "first strike") {
		t.Error("Blocker lacks first strike")
	}
}

// --- Sunhome Guildmage ---

func oldB410Pay(t *testing.T, g *game.Game, who uuid.UUID, mana ...string) {
	t.Helper()
	for _, m := range mana {
		if err := g.AddManaForEffect(who, uuid.Nil, m); err != nil {
			t.Fatalf("mana: %v", err)
		}
	}
}

func TestOldB410SunhomeGuildmagePumpsOnlyYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	mage := effectstest.B12Push(g, me.ID, "Sunhome Guildmage", "Creature — Human Wizard", oldB410Sunhome, 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	oldB410Pay(t, g, me.ID, "{C}", "{R}")
	if err := g.ActivateCatalogAbility(me.ID, mage, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without {W}")
	}
	me.ManaPool = game.ManaPool{}
	for i := 0; i < 2; i++ {
		oldB410Pay(t, g, me.ID, "{C}", "{R}", "{W}")
		if err := g.ActivateCatalogAbility(me.ID, mage, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	for id, want := range map[uuid.UUID]int{mage: 4, other: 4, theirs: 2} {
		if p, _ := effectstest.EffectivePTOf(t, g, id); p != want {
			t.Errorf("%s power %d, want %d", id, p, want)
		}
	}
}

func TestOldB410SunhomeGuildmageMakesAHasteSoldier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mage := effectstest.B12Push(g, me.ID, "Sunhome Guildmage", "Creature — Human Wizard", oldB410Sunhome, 2, 2)
	oldB410Pay(t, g, me.ID, "{C}", "{C}", "{R}", "{W}")
	if err := g.ActivateCatalogAbility(me.ID, mage, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var soldier *game.Card
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsToken() && c.Name == "Soldier" && c.Controller == me.ID {
			c := c
			soldier = &c
		}
	}
	if soldier == nil {
		t.Fatal("no Soldier token")
	}
	if soldier.Power != 1 || soldier.Toughness != 1 || len(soldier.Colors) != 2 {
		t.Errorf("token is %d/%d with colors %v, want 1/1 red and white", soldier.Power, soldier.Toughness, soldier.Colors)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, soldier.InstanceID, "haste") {
		t.Error("the Soldier has no haste")
	}
}

// --- Gateway Shade ---

func TestOldB410GatewayShadePumpsWithBlackManaOrByTappingAGate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	shade := effectstest.B12Push(g, me.ID, "Gateway Shade", "Creature — Shade", oldB410Shade, 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, shade, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("tapped a Gate that does not exist")
	}
	gate := effectstest.PushCatalogPermanent(g, me.ID, "Test Gate", "Land — Gate", "", false)
	if err := g.ActivateCatalogAbility(me.ID, shade, 1, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{gate}}); err != nil {
		t.Fatalf("tap a Gate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, gate) {
		t.Error("the Gate is untapped")
	}
	if p, tg := effectstest.EffectivePTOf(t, g, shade); p != 3 || tg != 3 {
		t.Errorf("Shade is %d/%d after the Gate, want 3/3", p, tg)
	}
	oldB410Pay(t, g, me.ID, "{B}")
	if err := g.ActivateCatalogAbility(me.ID, shade, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{B}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.EffectivePTOf(t, g, shade); p != 4 || tg != 4 {
		t.Errorf("Shade is %d/%d after {B}, want 4/4", p, tg)
	}
}
