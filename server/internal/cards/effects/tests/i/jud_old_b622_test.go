package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// jud_old_b622_test.go — S82 OLD batch 622 (#5611): Judgment, threshold. Each
// test plays the real catalog card through the engine.

const (
	oldB622Vampire   = "1d33d49f-ba12-4bf6-878b-04cee5d5b48d"
	oldB622Werewolf  = "59681998-c395-4d37-a419-2a78cf894506"
	oldB622Scrounger = "92ac7531-10bb-47a2-9dd0-378d547b4ed6"
	oldB622Seraph    = "bcc7e32a-9cf9-4fdf-a0a4-e90ae42f1353"
	oldB622Warcraft  = "c152f829-0452-407b-8020-023bf3d75c74"
	oldB622Sandstorm = "dc5cb229-82c9-4a87-9a10-490350cbdfc0"
	oldB622Gorgon    = "f41b1a5b-bc99-4c51-9665-e251f04ad83f"
	oldB622Aven      = "f165e482-5b87-4062-87ac-bbe8303f74c9" // Battlewise Aven: prints flying
)

func TestOldB622EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB622Vampire: "Treacherous Vampire", oldB622Werewolf: "Treacherous Werewolf",
		oldB622Scrounger: "Battlefield Scrounger", oldB622Seraph: "Silver Seraph",
		oldB622Warcraft: "Aven Warcraft", oldB622Sandstorm: "Swirling Sandstorm", oldB622Gorgon: "Masked Gorgon",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB622TreacherousWerewolfGrowsAndCostsLifeOnDeathAtThreshold(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	wolf := effectstest.BlbB1Push(g, me.ID, "Treacherous Werewolf", "Creature — Werewolf Minion", oldB622Werewolf, 2, 2)
	// Seven cards in an opponent's graveyard are not threshold.
	effectstest.BlbMPFillGraveyard(g, opp, 7)
	effectstest.BlbMPFillGraveyard(g, me, 6)
	if p, tough := effectstest.PtOf(t, g, wolf); p != 2 || tough != 2 {
		t.Fatalf("six cards: %d/%d, want 2/2", p, tough)
	}
	effectstest.BlbMPFillGraveyard(g, me, 1)
	if p, tough := effectstest.PtOf(t, g, wolf); p != 4 || tough != 4 {
		t.Fatalf("seven cards: %d/%d, want 4/4", p, tough)
	}
	life := me.Life
	effectstest.P1Destroy(t, g, wolf)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - me.Life; got != 4 {
		t.Errorf("lost %d life when the threshold Werewolf died, want 4", got)
	}
}

func TestOldB622TreacherousWerewolfDiesQuietlyBelowThreshold(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	wolf := effectstest.BlbB1Push(g, me.ID, "Treacherous Werewolf", "Creature — Werewolf Minion", oldB622Werewolf, 2, 2)
	effectstest.BlbMPFillGraveyard(g, me, 5)
	life := me.Life
	effectstest.P1Destroy(t, g, wolf)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("lost %d life below threshold", life-me.Life)
	}
}

func oldB622AttackWithVampire(t *testing.T, g *game.Game, me, opp *game.Player) uuid.UUID {
	t.Helper()
	v := effectstest.PushDiesCreatureForTest(g, me.ID, "Treacherous Vampire", oldB622Vampire, "Creature — Vampire", 4, 4)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(v, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	return v
}

func oldB622Prompt(g *game.Game, who uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == who {
			return c
		}
	}
	return nil
}

func TestOldB622TreacherousVampireIsSacrificedUnlessYouExileACard(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp, _ := effectstest.BlbB1Seats(g)
		fuel := effectstest.PushGraveyardCardTyped(me, "Fuel", "Sorcery")
		v := oldB622AttackWithVampire(t, g, me, opp)
		effectstest.PassPriorityAroundTable(t, g)
		c := oldB622Prompt(g, me.ID)
		if c == nil {
			t.Fatalf("pay=%v: no prompt for the Vampire's controller", pay)
		}
		var ids []uuid.UUID
		if pay {
			ids = []uuid.UUID{fuel}
		}
		if err := g.ResolvePayUnlessWithCards(c.ID, me.ID, pay, ids); err != nil {
			t.Fatalf("pay=%v: %v", pay, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if pay && (!g.Battlefield.Contains(v) || !g.Exile.Contains(fuel)) {
			t.Errorf("paid: the Vampire stayed=%v, fuel exiled=%v", g.Battlefield.Contains(v), g.Exile.Contains(fuel))
		}
		if !pay && g.Battlefield.Contains(v) {
			t.Error("declined: the Vampire was not sacrificed")
		}
	}
}

func TestOldB622TreacherousVampireWithAnEmptyGraveyardCannotPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	v := oldB622AttackWithVampire(t, g, me, opp)
	effectstest.PassPriorityAroundTable(t, g)
	if c := oldB622Prompt(g, me.ID); c != nil {
		if err := g.ResolvePayUnlessWithCards(c.ID, me.ID, true, nil); err != nil {
			_ = g.ResolvePayUnless(c.ID, me.ID, false)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(v) {
		t.Error("with nothing to exile the Vampire survived its attack")
	}
}

func TestOldB622TreacherousVampireAtThresholdIsBiggerAndCostsSixLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	v := effectstest.BlbB1Push(g, me.ID, "Treacherous Vampire", "Creature — Vampire", oldB622Vampire, 4, 4)
	if !effectstest.HasKeywordOnBattlefield(t, g, v, "flying") {
		t.Fatal("no flying")
	}
	effectstest.BlbMPFillGraveyard(g, me, 7)
	if p, tough := effectstest.PtOf(t, g, v); p != 6 || tough != 6 {
		t.Fatalf("at threshold: %d/%d, want 6/6", p, tough)
	}
	life := me.Life
	effectstest.P1Destroy(t, g, v)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - me.Life; got != 6 {
		t.Errorf("lost %d life, want 6", got)
	}
}

func TestOldB622SilverSeraphPumpsOnlyYourOtherCreaturesAtThreshold(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	seraph := effectstest.BlbB1Push(g, me.ID, "Silver Seraph", "Creature — Angel", oldB622Seraph, 6, 6)
	mine := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.BlbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.BlbMPFillGraveyard(g, me, 6)
	if p, _ := effectstest.PtOf(t, g, mine); p != 2 {
		t.Fatalf("below threshold the Bear is %d power, want 2", p)
	}
	effectstest.BlbMPFillGraveyard(g, me, 1)
	if p, tough := effectstest.PtOf(t, g, mine); p != 4 || tough != 4 {
		t.Errorf("at threshold your Bear is %d/%d, want 4/4", p, tough)
	}
	if p, _ := effectstest.PtOf(t, g, theirs); p != 2 {
		t.Errorf("their Bear got the bonus (%d power)", p)
	}
	if p, _ := effectstest.PtOf(t, g, seraph); p != 6 {
		t.Errorf("the Seraph pumped itself (%d power)", p)
	}
}

func TestOldB622MaskedGorgonGivesGreenAndWhiteCreaturesProtectionFromGorgons(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	gorgon := effectstest.BlbB1Push(g, me.ID, "Masked Gorgon", "Creature — Gorgon", oldB622Gorgon, 5, 5)
	green := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Green Bear", TypeLine: "Creature — Bear", Colors: []string{"G"},
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	white := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "White Knight", TypeLine: "Creature — Knight", Colors: []string{"W"},
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	red := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Red Goblin", TypeLine: "Creature — Goblin", Colors: []string{"R"},
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	const prot = "protection from Gorgons"
	if !effectstest.HasKeywordOnBattlefield(t, g, green, prot) || !effectstest.HasKeywordOnBattlefield(t, g, white, prot) {
		t.Error("a green or white creature lacks protection from Gorgons")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, red, prot) {
		t.Error("a red creature has protection from Gorgons")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, gorgon, game.ProtectionFromColor("G")) {
		t.Error("the Gorgon has protection from green below threshold")
	}
	effectstest.BlbMPFillGraveyard(g, me, 7)
	for _, c := range []string{"G", "W"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, gorgon, game.ProtectionFromColor(c)) {
			t.Errorf("no protection from %s at threshold", c)
		}
	}
}

func TestOldB622AvenWarcraftPumpsToughnessAndAtThresholdGivesProtection(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.BlbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Aven Warcraft", "Instant", oldB622Warcraft, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 2 || tough != 4 {
		t.Fatalf("the Bear is %d/%d, want 2/4", p, tough)
	}
	if _, tough := effectstest.PtOf(t, g, theirs); tough != 2 {
		t.Errorf("their Bear got the toughness (%d)", tough)
	}
	if effectstest.PendingOfKind(g, game.PendingChoiceColor) != nil {
		t.Fatal("a colour was asked for below threshold")
	}

	// The spell above is in the graveyard now: six more is threshold.
	effectstest.BlbMPFillGraveyard(g, me, 6)
	warcraft := effectstest.PushCatalogHandCard(me, "Aven Warcraft", "Instant", oldB622Warcraft)
	if err := g.CastSpell(me.ID, warcraft, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "R")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, game.ProtectionFromColor("R")) {
		t.Error("the Bear did not gain protection from red")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, theirs, game.ProtectionFromColor("R")) {
		t.Error("their Bear gained protection")
	}
}

func TestOldB622SwirlingSandstormNeedsThresholdAndSparesFliers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	bear := effectstest.BlbB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Swirling Sandstorm", "Sorcery", oldB622Sandstorm, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the Bear died without threshold")
	}

	// The Sandstorm is in the graveyard now: six more is seven.
	effectstest.BlbMPFillGraveyard(g, me, 6)
	flier := effectstest.BlbB1Push(g, opp.ID, "Battlewise Aven", "Creature — Bird Soldier", oldB622Aven, 2, 2)
	sandstorm := effectstest.PushCatalogHandCard(me, "Swirling Sandstorm", "Sorcery", oldB622Sandstorm)
	if err := g.CastSpell(me.ID, sandstorm, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the ground creature survived 5 damage")
	}
	if !g.Battlefield.Contains(flier) {
		t.Error("the flier was dealt damage")
	}
}

func TestOldB622BattlefieldScroungerNeedsThresholdAndOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	s := effectstest.BlbB1Push(g, me.ID, "Battlefield Scrounger", "Creature — Centaur", oldB622Scrounger, 3, 3)
	var pile []uuid.UUID
	for i := 0; i < 6; i++ {
		pile = append(pile, effectstest.PushGraveyardCardTyped(me, "Old", "Sorcery"))
	}
	if err := g.ActivateCatalogAbility(me.ID, s, 0, game.ActivateAbilityParams{ExileIDs: pile[:3]}); err == nil {
		t.Fatal("activated with six cards in the graveyard")
	}
	pile = append(pile, effectstest.PushGraveyardCardTyped(me, "Old", "Sorcery"))
	lib := me.Library.Size()
	effectstest.B16Activate(t, g, me.ID, s, 0, game.ActivateAbilityParams{ExileIDs: pile[:3]})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, s); p != 6 || tough != 6 {
		t.Errorf("the Scrounger is %d/%d, want 6/6", p, tough)
	}
	if me.Library.Size() != lib+3 {
		t.Errorf("library %d, want %d", me.Library.Size(), lib+3)
	}
	for _, id := range pile[:3] {
		if me.Graveyard.Contains(id) || g.Exile.Contains(id) {
			t.Error("a paid card is not in the library")
		}
	}
	// Four cards are left: below threshold, and already used this turn.
	for i := 0; i < 4; i++ {
		pile = append(pile, effectstest.PushGraveyardCardTyped(me, "More", "Sorcery"))
	}
	if err := g.ActivateCatalogAbility(me.ID, s, 0, game.ActivateAbilityParams{ExileIDs: pile[3:6]}); err == nil {
		t.Error("a second activation in one turn was accepted")
	}
}
