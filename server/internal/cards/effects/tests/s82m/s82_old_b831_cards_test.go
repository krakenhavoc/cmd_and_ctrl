package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b831_cards_test.go — Ice Age, batch 831 (S82 OLD, #5636): activated
// abilities. Every test plays the real catalog entry through the engine.

const (
	oldB831Crown      = "0b772ef2-b760-400b-8e5b-f4b98e6c935d"
	oldB831Bottle     = "1e381361-7404-4654-bfb4-5ea65f6e4dce"
	oldB831Merieke    = "288fdfbb-1236-4732-afd7-e7eaba860007"
	oldB831Spider     = "2c87565b-3902-4b11-b894-6922f66008ba"
	oldB831Hound      = "2fe8512f-826c-400d-ba6d-1de4f315b617"
	oldB831Sunstone   = "300ad0f7-92a7-450d-a9de-b658836fe3cb"
	oldB831Conjurer   = "3eb21641-90e1-4865-b7be-a21d52b7b56b"
	oldB831Sleigh     = "40960ec8-e3be-45f2-8d27-316d82c69123"
	oldB831Augury     = "42cffe5d-6e6b-4ac4-b539-36a7b3bc9a65"
	oldB831Bear       = "Creature — Bear"
	oldB831SnowForest = "Basic Snow Land — Forest"
)

func TestOldB831EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB831Crown: "Crown of the Ages", oldB831Bottle: "Elkin Bottle", oldB831Merieke: "Merieke Ri Berit",
		oldB831Spider: "Giant Trap Door Spider", oldB831Hound: "Snow Hound", oldB831Sunstone: "Sunstone",
		oldB831Conjurer: "Balduvian Conjurer", oldB831Sleigh: "Arcum's Sleigh", oldB831Augury: "Elemental Augury",
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

// oldB831Activate activates ability 0 of card with the mana floated.
func oldB831Activate(t *testing.T, g *game.Game, me *game.Player, card uuid.UUID, mana string, targets []game.TargetRef) error {
	t.Helper()
	if mana != "" {
		effectstest.FloatMana(t, g, me, mana)
	}
	err := g.ActivateCatalogAbility(me.ID, card, 0, game.ActivateAbilityParams{Strict: true, Targets: targets})
	if err == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	return err
}

func TestOldB831CrownOfTheAgesMovesAnAuraToAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	crown := effectstest.B12Push(g, me.ID, "Crown of the Ages", "Artifact", oldB831Crown, 0, 0)
	a := effectstest.B12Creature(g, me.ID, "Bear A", oldB831Bear, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", oldB831Bear, 2, 2)
	aura := effectstest.Enchant(t, g, "Imposing Visage", "6f67058b-ed14-4e3c-9af3-d61570870e36", a)
	land := effectstest.SeedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	if err := oldB831Activate(t, g, me, crown, "{C}{C}{C}", effectstest.CardTargets(aura)); err == nil {
		t.Fatal("activated with three of the four mana")
	}
	_ = land
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if err := oldB831Activate(t, g, me, crown, "{C}{C}{C}{C}", effectstest.CardTargets(effectstest.B12Push(g, me.ID, "Plain Artifact", "Artifact", "", 0, 0))); err == nil {
		t.Fatal("targeted something that is not an Aura on a creature")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if err := oldB831Activate(t, g, me, crown, "{C}{C}{C}{C}", effectstest.CardTargets(aura)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, b)
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, aura); !ok || c.AttachedTo.ID != b {
		t.Errorf("the Aura is attached to %v, want the other Bear", c.AttachedTo.ID)
	}
	if !effectstest.B16Tapped(t, g, crown) {
		t.Error("the Crown was not tapped")
	}
}

func TestOldB831ElkinBottleExilesTheTopCardUntilTheNextUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	top := effectstest.SeedLibrary(me, "Top", "Next")
	effectstest.AdvanceToMain(t, g)
	bottle := effectstest.B12Push(g, me.ID, "Elkin Bottle", "Artifact", oldB831Bottle, 0, 0)
	if err := oldB831Activate(t, g, me, bottle, "{C}{C}{C}", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !g.Exile.Contains(top[0]) || g.Exile.Contains(top[1]) {
		t.Fatal("exactly the top card should be exiled")
	}
	if !effectstest.PermissionLive(g, effectstest.ExiledPermission(g, top[0]), me.ID) {
		t.Fatal("the exiled card is not playable")
	}
	for i := 0; i < 300; i++ {
		if g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat == 0 && i > 0 {
			break
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if effectstest.PermissionLive(g, effectstest.ExiledPermission(g, top[0]), me.ID) {
		t.Error("the card is still playable after the next upkeep began")
	}
}

func TestOldB831MeriekeStealsACreatureAndDestroysItWhenSheUntapsOrLeaves(t *testing.T) {
	for _, leave := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		merieke := effectstest.B12Push(g, me.ID, "Merieke Ri Berit", "Legendary Creature — Human", oldB831Merieke, 1, 1)
		victim := effectstest.B12Creature(g, opp.ID, "Victim", oldB831Bear, 2, 2)
		if err := oldB831Activate(t, g, me, merieke, "", effectstest.CardTargets(victim)); err != nil {
			t.Fatalf("activate: %v", err)
		}
		if c, ok := effectstest.BattlefieldCard(g, victim); !ok || c.Controller != me.ID {
			t.Fatalf("the Victim is controlled by %v, want Merieke's controller", c.Controller)
		}
		if leave {
			if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, merieke); err != nil {
				t.Fatalf("MoveCardByID: %v", err)
			}
		} else if err := g.TapCard(merieke, false); err != nil {
			t.Fatalf("untap: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effects.OnBattlefield(g, victim) {
			t.Errorf("leave=%v: the stolen creature was not destroyed", leave)
		}
	}
}

func TestOldB831MeriekeDoesntUntapInTheUntapStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	merieke := effectstest.B12Push(g, me.ID, "Merieke Ri Berit", "Legendary Creature — Human", oldB831Merieke, 1, 1)
	effectstest.TapForTest(t, g, merieke)
	for i := 0; i < 300; i++ {
		if g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat == 0 && i > 0 {
			break
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if !effectstest.B16Tapped(t, g, merieke) {
		t.Error("Merieke untapped during her controller's untap step")
	}
}

func TestOldB831GiantTrapDoorSpiderExilesItselfAndAnAttackerWithoutFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", oldB831Bear, 2, 2)
	flier := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Flier", TypeLine: oldB831Bear,
		Power: 2, Toughness: 2, Keywords: []string{"flying"}, Owner: me.ID, Controller: me.ID})
	spider := effectstest.B12Push(g, opp.ID, "Giant Trap Door Spider", oldB831Bear, oldB831Spider, 2, 3)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, a := range []uuid.UUID{attacker, flier} {
		if err := g.DeclareAttacker(a, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.FloatMana(t, g, opp, "{C}{R}{G}")
	if err := g.ActivateCatalogAbility(opp.ID, spider, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(flier)}); err == nil {
		t.Fatal("targeted an attacker with flying")
	}
	if err := g.ActivateCatalogAbility(opp.ID, spider, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(attacker)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, attacker) || effects.OnBattlefield(g, spider) {
		t.Error("the Spider and the attacker were not both exiled")
	}
	if !g.Exile.Contains(attacker) || !g.Exile.Contains(spider) {
		t.Error("they were not exiled")
	}
}

func TestOldB831SnowHoundReturnsItselfAndAGreenOrBlueCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	hound := effectstest.B12Push(g, me.ID, "Snow Hound", "Creature — Dog", oldB831Hound, 1, 1)
	green := effectstest.B12Creature(g, me.ID, "Green Bear", oldB831Bear, 2, 2)
	effectstest.AlliSetColors(g, green, "G")
	red := effectstest.B12Creature(g, me.ID, "Red Bear", oldB831Bear, 2, 2)
	effectstest.AlliSetColors(g, red, "R")
	if err := oldB831Activate(t, g, me, hound, "{C}", effectstest.CardTargets(red)); err == nil {
		t.Fatal("targeted a red creature")
	}
	if err := oldB831Activate(t, g, me, hound, "{C}", effectstest.CardTargets(green)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if effects.OnBattlefield(g, hound) || effects.OnBattlefield(g, green) {
		t.Error("the Hound and the green creature were not both returned")
	}
	if !effects.OnBattlefield(g, red) {
		t.Error("the red creature left")
	}
}

func TestOldB831SunstonePreventsAllCombatDamageForASnowLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sunstone := effectstest.B12Push(g, me.ID, "Sunstone", "Artifact", oldB831Sunstone, 0, 0)
	plain := effectstest.SeedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	snow := effectstest.SeedLand(g, me.ID, "Snow-Covered Forest", oldB831SnowForest, "")
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", oldB831Bear, 3, 3)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	_ = plain
	life := effectstest.LifeOf(g, opp.ID)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, sunstone, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{plain}}); err == nil {
		t.Fatal("sacrificed a nonsnow land")
	}
	if err := g.ActivateCatalogAbility(me.ID, sunstone, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{snow}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, snow) {
		t.Error("the snow land was not sacrificed")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if got := effectstest.LifeOf(g, opp.ID); got != life {
		t.Errorf("the defender lost %d life; combat damage was not prevented", life-got)
	}
}

func TestOldB831BalduvianConjurerAnimatesASnowLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	conj := effectstest.B12Push(g, me.ID, "Balduvian Conjurer", "Creature — Human Wizard", oldB831Conjurer, 0, 2)
	plain := effectstest.SeedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	snow := effectstest.SeedLand(g, me.ID, "Snow-Covered Forest", oldB831SnowForest, "")
	if err := oldB831Activate(t, g, me, conj, "", effectstest.CardTargets(plain)); err == nil {
		t.Fatal("targeted a nonsnow land")
	}
	if err := oldB831Activate(t, g, me, conj, "", effectstest.CardTargets(snow)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if p, tough := effectstest.PtOf(t, g, snow); p != 2 || tough != 2 {
		t.Errorf("the land is %d/%d, want 2/2", p, tough)
	}
	c, _ := effectstest.BattlefieldCard(g, snow)
	if !c.IsCreature() || !c.IsLand() {
		t.Errorf("the land is creature=%v land=%v, want both", c.IsCreature(), c.IsLand())
	}
}

func TestOldB831ArcumsSleighNeedsCombatAndADefendingSnowLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sleigh := effectstest.B12Push(g, me.ID, "Arcum's Sleigh", "Artifact", oldB831Sleigh, 0, 0)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", oldB831Bear, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, sleigh, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(attacker)}); err == nil {
		t.Fatal("activated outside combat")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, sleigh, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(attacker)}); err == nil {
		t.Fatal("activated while the defending player has no snow land")
	}
	effectstest.SeedLand(g, opp.ID, "Snow-Covered Forest", oldB831SnowForest, "")
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, sleigh, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(attacker)}); err != nil {
		t.Fatalf("activate with a defending snow land: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, attacker, "vigilance") {
		t.Error("the creature did not gain vigilance")
	}
}

func TestOldB831ElementalAuguryOffersTheTopThreeOfATargetLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(opp, "A", "B", "C", "D")
	augury := effectstest.B12Push(g, me.ID, "Elemental Augury", "Enchantment", oldB831Augury, 0, 0)
	if err := oldB831Activate(t, g, me, augury, "{C}{C}{C}", []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID) == nil && len(g.PendingChoices) == 0 {
		t.Error("the controller was not asked to order the top three cards")
	}
}
