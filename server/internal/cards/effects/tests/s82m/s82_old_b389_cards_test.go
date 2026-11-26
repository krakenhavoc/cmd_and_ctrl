package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b389_cards_test.go — Theros, batch 389 (#5561, S82 OLD).
// Every test plays the real catalog entry through the engine.

const (
	oldB389Hunt     = "00afc724-be09-4e92-a50c-88011bd10ccd"
	oldB389Surge    = "0acbc9a6-0de0-4096-b35c-1a586a1312ea"
	oldB389Triton   = "55b72176-e428-4d2a-a701-237e8c8de22e"
	oldB389Adept    = "ab0707fb-b404-433d-b845-0d50d01cd72e"
	oldB389Valor    = "b4540f6a-bec8-4942-aebe-d3285b8a8134"
	oldB389Warcall  = "da04b438-aa49-4736-b0ed-131bbe70b566"
	oldB389Labyr    = "e6813885-b629-4bd4-b445-da8ab4363080"
	oldB389Griffin  = "fbab033e-3a8f-4c08-a445-46ae34c283eb"
	oldB389Free     = "{0}"
	oldB389Instant  = "Instant"
	oldB389Sorcery  = "Sorcery"
	oldB389BearType = "Creature — Bear"
)

func TestOldB389CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB389Hunt: "Hunt the Hunter", oldB389Surge: "Savage Surge", oldB389Triton: "Triton Tactics",
		oldB389Adept: "Flamespeaker Adept", oldB389Valor: "Battlewise Valor", oldB389Warcall: "Kragma Warcaller",
		oldB389Labyr: "Lost in a Labyrinth", oldB389Griffin: "Setessan Griffin",
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

// oldB389Cast casts a free spell strictly at the given targets where the
// game stands, from the active seat's hand.
func oldB389Cast(g *game.Game, name, typeLine, oracle string, targets ...uuid.UUID) error {
	p := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: oldB389Free, OracleID: oracle, Owner: p.ID, Controller: p.ID})
	})
	return g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTargets(targets...)})
}

func oldB389Main(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	return g, g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func oldB389Resolve(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
}

func oldB389Scry(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	scry := effectstest.LatestChoiceOfKind(g, game.PendingChoiceScry)
	if scry == nil {
		t.Fatal("the scry asked nothing")
	}
	if err := g.ResolveScry(scry.ID, p.ID, nil, scry.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	oldB389Resolve(t, g)
}

func TestOldB389HuntTheHunterPumpsThenFightsGreenCreatures(t *testing.T) {
	g, me, opp := oldB389Main(t)
	mine := effectstest.B21Push(g, me.ID, "Green Bear", oldB389BearType, "", 2, 2, "G")
	redMine := effectstest.B21Push(g, me.ID, "Red Bear", oldB389BearType, "", 2, 2, "R")
	theirs := effectstest.B21Push(g, opp.ID, "Their Green Bear", oldB389BearType, "", 3, 3, "G")
	redTheirs := effectstest.B21Push(g, opp.ID, "Their Red Bear", oldB389BearType, "", 3, 3, "R")
	for _, bad := range [][]uuid.UUID{{redMine, theirs}, {mine, redTheirs}, {theirs, mine}} {
		if err := oldB389Cast(g, "Hunt the Hunter", oldB389Sorcery, oldB389Hunt, bad...); err == nil {
			t.Fatalf("cast accepted the targets %v", bad)
		}
	}
	if err := oldB389Cast(g, "Hunt the Hunter", oldB389Sorcery, oldB389Hunt, mine, theirs); err != nil {
		t.Fatalf("cast: %v", err)
	}
	oldB389Resolve(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the 3/3 survived a fight with a 4/4")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the pumped 4/4 died to a 3/3")
	}
	if p := effectstest.EffectivePower(t, g, mine); p != 4 {
		t.Errorf("power %d, want 4", p)
	}
}

func TestOldB389SavageSurgePumpsAndUntaps(t *testing.T) {
	g, me, _ := oldB389Main(t)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB389BearType, 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardByID(g, bear).Tapped = true })
	if err := oldB389Cast(g, "Savage Surge", oldB389Instant, oldB389Surge); err == nil {
		t.Fatal("cast without a target")
	}
	if err := oldB389Cast(g, "Savage Surge", oldB389Instant, oldB389Surge, bear); err != nil {
		t.Fatalf("cast: %v", err)
	}
	oldB389Resolve(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 4 {
		t.Errorf("power %d, want 4", p)
	}
	if effectstest.B16Tapped(t, g, bear) {
		t.Error("the creature is still tapped")
	}
}

func TestOldB389BattlewiseValorPumpsAndScries(t *testing.T) {
	g, me, _ := oldB389Main(t)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB389BearType, 2, 2)
	if err := oldB389Cast(g, "Battlewise Valor", oldB389Instant, oldB389Valor, bear); err != nil {
		t.Fatalf("cast: %v", err)
	}
	oldB389Resolve(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 4 {
		t.Errorf("power %d, want 4", p)
	}
	oldB389Scry(t, g, me)
}

func TestOldB389LostInALabyrinthShrinksAndScries(t *testing.T) {
	g, me, opp := oldB389Main(t)
	elk := effectstest.B12Creature(g, opp.ID, "Elk", "Creature — Elk", 5, 5)
	if err := oldB389Cast(g, "Lost in a Labyrinth", oldB389Instant, oldB389Labyr, elk); err != nil {
		t.Fatalf("cast: %v", err)
	}
	oldB389Resolve(t, g)
	if p := effectstest.EffectivePower(t, g, elk); p != 2 {
		t.Errorf("power %d, want 2", p)
	}
	oldB389Scry(t, g, me)
}

func TestOldB389SetessanGriffinPumpsOnlyOnceEachTurn(t *testing.T) {
	g, me, _ := oldB389Main(t)
	griffin := effectstest.PushCatalogPermanent(g, me.ID, "Setessan Griffin", "Creature — Griffin", oldB389Griffin, false)
	effectstest.B06AddMana(me, "G", "G", "G", "G", "G", "G", "G", "G")
	effectstest.B16Activate(t, g, me.ID, griffin, 0, game.ActivateAbilityParams{})
	oldB389Resolve(t, g)
	if p := effectstest.EffectivePower(t, g, griffin); p != 3 {
		t.Errorf("power %d, want 3 (a 1/1 test body plus 2)", p)
	}
	effectstest.AcRefused(t, g, me.ID, griffin, 0, game.ActivateAbilityParams{})
}

func TestOldB389FlamespeakerAdeptPumpsWhenYouScry(t *testing.T) {
	g, me, opp := oldB389Main(t)
	adept := effectstest.PushCatalogPermanent(g, me.ID, "Flamespeaker Adept", "Creature — Human Shaman", oldB389Adept, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB389BearType, 2, 2)
	if err := oldB389Cast(g, "Battlewise Valor", oldB389Instant, oldB389Valor, bear); err != nil {
		t.Fatalf("cast: %v", err)
	}
	oldB389Resolve(t, g)
	oldB389Scry(t, g, me)
	oldB389Resolve(t, g)
	if p := effectstest.EffectivePower(t, g, adept); p != 3 {
		t.Errorf("power %d, want 3 (a 1/1 test body plus 2)", p)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, adept, "first strike") {
		t.Error("no first strike")
	}
	_ = opp
}

func TestOldB389FlamespeakerAdeptIgnoresAnOpponentsScry(t *testing.T) {
	g, me, opp := oldB389Main(t)
	adept := effectstest.PushCatalogPermanent(g, me.ID, "Flamespeaker Adept", "Creature — Human Shaman", oldB389Adept, false)
	oppBear := effectstest.B12Creature(g, opp.ID, "Bear", oldB389BearType, 2, 2)
	effectstest.B13OpponentCasts(t, g, opp, "Battlewise Valor", oldB389Instant, oldB389Valor, oldB389Free, effectstest.CardTargets(oppBear))
	oldB389Resolve(t, g)
	if scry := effectstest.LatestChoiceOfKind(g, game.PendingChoiceScry); scry != nil {
		if err := g.ResolveScry(scry.ID, opp.ID, nil, scry.ScryCards); err != nil {
			t.Fatalf("ResolveScry: %v", err)
		}
	}
	oldB389Resolve(t, g)
	if p := effectstest.EffectivePower(t, g, adept); p != 1 {
		t.Errorf("power %d, want 1: an opponent's scry must not pump it", p)
	}
}

func TestOldB389KragmaWarcallerGivesMinotaursHasteAndPumpsAttackers(t *testing.T) {
	g, me, opp := oldB389Main(t)
	warcaller := effectstest.PushCatalogPermanent(g, me.ID, "Kragma Warcaller", "Creature — Minotaur Warrior", oldB389Warcall, false)
	minotaur := effectstest.B12Creature(g, me.ID, "Plains Minotaur", "Creature — Minotaur", 2, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB389BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Minotaur", "Creature — Minotaur", 2, 3)
	if !effectstest.HasKeywordOnBattlefield(t, g, minotaur, "haste") || !effectstest.HasKeywordOnBattlefield(t, g, warcaller, "haste") {
		t.Error("your Minotaurs do not have haste")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "haste") || effectstest.HasKeywordOnBattlefield(t, g, theirs, "haste") {
		t.Error("a non-Minotaur or an opponent's Minotaur has haste")
	}
	effectstest.DeclareAttack(t, g, opp.ID, minotaur, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, minotaur); p != 4 {
		t.Errorf("attacking Minotaur power %d, want 4", p)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("attacking Bear power %d, want 2", p)
	}
}

func TestOldB389TritonTacticsTapsAndHoldsWhatTheyBlocked(t *testing.T) {
	g, me, opp := oldB389Main(t)
	attacker := effectstest.Pr7bCreature(g, me.ID, "Vigilant Attacker", 1, "{2}", nil, "vigilance")
	other := effectstest.B12Creature(g, me.ID, "Unblocked Attacker", oldB389BearType, 1, 4)
	blocker := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{attacker, other} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatal(err)
	}
	if err := oldB389Cast(g, "Triton Tactics", oldB389Instant, oldB389Triton, blocker, other, attacker); err == nil {
		t.Fatal("three targets were accepted")
	}
	if err := oldB389Cast(g, "Triton Tactics", oldB389Instant, oldB389Triton, blocker); err != nil {
		t.Fatalf("cast: %v", err)
	}
	oldB389Resolve(t, g)
	if toughness := effectstest.EffectiveToughness(t, g, blocker); toughness != 4 {
		t.Errorf("blocker toughness %d, want 4", toughness)
	}
	if effectstest.B16Tapped(t, g, attacker) {
		t.Fatal("the vigilant attacker is already tapped")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, attacker) {
		t.Error("the creature the pumped blocker blocked was not tapped at end of combat")
	}
	if effectstest.B16Tapped(t, g, blocker) {
		t.Error("the blocker itself was tapped")
	}
}

func TestOldB389TritonTacticsUntapsThePumpedCreatures(t *testing.T) {
	g, me, _ := oldB389Main(t)
	a := effectstest.B12Creature(g, me.ID, "A", oldB389BearType, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", oldB389BearType, 2, 2)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardByID(g, a).Tapped = true
		effectstest.FindBattlefieldCardByID(g, b).Tapped = true
	})
	if err := oldB389Cast(g, "Triton Tactics", oldB389Instant, oldB389Triton, a, b); err != nil {
		t.Fatalf("cast: %v", err)
	}
	oldB389Resolve(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if effectstest.B16Tapped(t, g, id) {
			t.Errorf("%s is still tapped", id)
		}
		if toughness := effectstest.EffectiveToughness(t, g, id); toughness != 5 {
			t.Errorf("toughness %d, want 5", toughness)
		}
	}
}
