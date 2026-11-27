package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB474Skyfisher = "054a0193-861e-4644-ab0a-8aa33407abc8"
	oldB474Aeronaut  = "1adb733c-d18a-4019-9229-0ad701a15b1a"
	oldB474Noose     = "456f2991-ab64-41bc-9bd1-a1b1e7993ed6"
	oldB474Harrier   = "546f037b-aea8-4c7a-9b66-373f1be0bac6"
	oldB474Marshal   = "0c7239bf-dc8a-4d79-867e-7a4225568c49"
	oldB474Crusade   = "4692740f-be90-459f-8d90-c4ae71771595"
	oldB474Barrier   = "669701a5-fd7a-43f8-93f4-da7a8c4dd3dc"
	oldB474Extract   = "ac2711b7-318d-478a-b69c-d44baff46f2d"
	oldB474Biskelion = "c8552706-83f2-41d7-8068-75891abb8bde"
	oldB474Esperzoa  = "c9a188f9-2e65-4a66-826f-0150f25730af"
	oldB474Stones    = "f3658894-3d3d-4cd4-b0ac-c53e1d08747c"
)

func TestOldB474CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB474Skyfisher: "Kor Skyfisher", oldB474Aeronaut: "Kor Aeronaut", oldB474Noose: "Trip Noose",
		oldB474Harrier: "Goldmeadow Harrier", oldB474Marshal: "Seasoned Marshal", oldB474Crusade: "Crusade",
		oldB474Barrier: "Razor Barrier", oldB474Extract: "Moonglove Extract", oldB474Biskelion: "Serrated Biskelion",
		oldB474Esperzoa: "Esperzoa", oldB474Stones: "Stalking Stones",
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

func TestOldB474SkyfisherReturnsAPermanentYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	id := effectstest.CastWithCost(t, g, "Kor Skyfisher", "Creature — Kor Soldier", "{1}{W}", oldB474Skyfisher)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil {
		t.Fatal("Kor Skyfisher queued no return pick")
	}
	if err := g.ResolveOwnPermanents(p.ID, me.ID, []uuid.UUID{rock}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if !me.Hand.Contains(rock) {
		t.Error("the permanent was not returned to hand")
	}
	if !effectstest.HasEffectiveKeyword(t, g, id, "flying") {
		t.Error("Kor Skyfisher has no flying")
	}
}

func TestOldB474AeronautGrantsFlyingOnlyWhenKicked(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		var optional []int
		if kicked {
			optional = []int{0}
		}
		id, err := effectstest.CastWithOptionalCosts(t, g, "Kor Aeronaut", "Creature — Kor Soldier", oldB474Aeronaut, nil, optional, nil)
		if err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if kicked {
			if ch := effectstest.LatestPickTarget(g, me.ID); ch == nil {
				t.Fatal("the kicked enters trigger asked for no target")
			}
			effectstest.AnswerPickTarget(t, g, bear)
			effectstest.PassPriorityAroundTable(t, g)
		} else if effectstest.LatestPickTarget(g, me.ID) != nil {
			t.Fatal("an unkicked Kor Aeronaut asked for a target")
		}
		if got := effectstest.HasEffectiveKeyword(t, g, bear, "flying"); got != kicked {
			t.Errorf("kicked=%v: the Bear has flying = %v", kicked, got)
		}
		if !effectstest.HasEffectiveKeyword(t, g, id, "flying") {
			t.Errorf("kicked=%v: Kor Aeronaut has no flying", kicked)
		}
	}
}

func TestOldB474NooseAndHarrierTapTheTargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	noose := effectstest.B12Permanent(g, me.ID, "Trip Noose", "Artifact")
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, noose).OracleID = oldB474Noose })
	harrier := effectstest.B12Push(g, me.ID, "Goldmeadow Harrier", "Creature — Kithkin Soldier", oldB474Harrier, 1, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, opp.ID, "Their Wolf", "Creature — Wolf", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.B16Activate(t, g, me.ID, noose, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(victim)})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.ApaLive(g, victim).Tapped || effectstest.ApaLive(g, other).Tapped {
		t.Error("Trip Noose did not tap exactly the target")
	}
	if !effectstest.ApaLive(g, noose).Tapped {
		t.Error("Trip Noose's {T} was not paid")
	}
	// Goldmeadow Harrier needs {W} and its own tap.
	if err := g.ActivateCatalogAbility(me.ID, harrier, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(other)}); err == nil {
		t.Fatal("Goldmeadow Harrier activated without {W}")
	}
	effectstest.FloatMana(t, g, me, "{W}")
	effectstest.B16Activate(t, g, me.ID, harrier, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(other)})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.ApaLive(g, other).Tapped || !effectstest.ApaLive(g, harrier).Tapped {
		t.Error("Goldmeadow Harrier did not tap the target and itself")
	}
}

func TestOldB474MarshalMayTapTheTargetWhenItAttacks(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.SeatsForCombat(g)
		marshal := effectstest.B12Push(g, me.ID, "Seasoned Marshal", "Creature — Human Soldier", oldB474Marshal, 2, 2)
		blocker := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(marshal, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		effectstest.LockInAttacks(t, g)
		effectstest.AnswerPickTarget(t, g, blocker)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.ApaLive(g, blocker).Tapped; got != yes {
			t.Errorf("answered %v: the target is tapped = %v", yes, got)
		}
	}
}

func TestOldB474CrusadeGivesEveryWhiteCreaturePlusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Permanent(g, me.ID, "Crusade", "Enchantment")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].Name == "Crusade" {
				g.Battlefield.Cards[i].OracleID = oldB474Crusade
			}
		}
	})
	mine := effectstest.B16Creature(g, me.ID, "White Knight", "Creature — Human Knight", 2, 2, "W")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Cleric", "Creature — Human Cleric", 1, 1, "W")
	green := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	for id, want := range map[uuid.UUID]int{mine: 3, theirs: 2, green: 2} {
		if p, _ := effectstest.PtOf(t, g, id); p != want {
			t.Errorf("power %d, want %d", p, want)
		}
	}
}

func TestOldB474RazorBarrierGivesTheChosenProtection(t *testing.T) {
	for i, want := range []string{"protection from artifacts", game.ProtectionFromColor("W"), game.ProtectionFromColor("G")} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		idx := []int{0, 1, 5}[i]
		effectstest.AdvanceToMain(t, g)
		if err := effectstest.CastCatalogSpellErr(t, g, "Razor Barrier", "Instant", oldB474Barrier, effectstest.CardRefs(theirs)); err == nil {
			t.Fatal("an opponent's permanent was a legal target")
		}
		effectstest.CastCatalogSpell(t, g, "Razor Barrier", "Instant", oldB474Barrier, effectstest.CardRefs(mine))
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerOptionPick(t, g, me.ID, idx)
		effectstest.PassPriorityAroundTable(t, g)
		for _, kw := range []string{"protection from artifacts", game.ProtectionFromColor("W"), game.ProtectionFromColor("G"), game.ProtectionFromColor("R")} {
			if got := effectstest.HasEffectiveKeyword(t, g, mine, kw); got != (kw == want) {
				t.Errorf("answer %d: %s = %v", idx, kw, got)
			}
		}
	}
}

func TestOldB474MoongloveExtractDealsTwoToAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, me.ID, "Moonglove Extract", "Artifact")
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, rock).OracleID = oldB474Extract })
	effectstest.AdvanceToMain(t, g)
	before := opp.Life
	effectstest.B16Activate(t, g, me.ID, rock, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("life %d → %d, want 2 damage", before, opp.Life)
	}
	if _, ok := effectstest.BattlefieldCard(g, rock); ok {
		t.Error("the Extract was not sacrificed")
	}
}

func TestOldB474BiskelionPutsACounterOnItselfAndOnTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bis := effectstest.B27Push(g, me.ID, "Serrated Biskelion", "Artifact Creature — Construct", oldB474Biskelion, "{3}", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, bis, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(theirs)})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bis); p != 1 || tough != 1 {
		t.Errorf("the Biskelion is %d/%d, want 1/1", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, theirs); p != 1 || tough != 1 {
		t.Errorf("the target is %d/%d, want 1/1", p, tough)
	}
	if !effectstest.ApaLive(g, bis).Tapped {
		t.Error("the {T} cost was not paid")
	}
}

func TestOldB474EsperzoaReturnsAnArtifactAtYourUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	zoa := effectstest.B27Push(g, me.ID, "Esperzoa", "Artifact Creature — Jellyfish", oldB474Esperzoa, "{2}{U}", 4, 3)
	rock := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	start := g.Turn.Seq
	for g.Turn.ActiveSeat != 0 || g.Turn.Step != game.StepUpkeep || g.Turn.Seq <= start {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil {
		t.Fatal("Esperzoa queued no return pick")
	}
	if err := g.ResolveOwnPermanents(p.ID, me.ID, []uuid.UUID{bear}); err == nil {
		t.Error("a nonartifact creature was accepted")
	}
	if err := g.ResolveOwnPermanents(p.ID, me.ID, []uuid.UUID{rock}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if !me.Hand.Contains(rock) {
		t.Error("the artifact was not returned to hand")
	}
	if _, ok := effectstest.BattlefieldCard(g, zoa); !ok {
		t.Error("Esperzoa left the battlefield")
	}
}

func TestOldB474StalkingStonesBecomeAnIndefiniteCreatureLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.B12Permanent(g, me.ID, "Stalking Stones", "Land")
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, land).OracleID = oldB474Stones })
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{C}")
	effectstest.B16Activate(t, g, me.ID, land, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	live := effectstest.ApaLive(g, land)
	if !live.IsLand() || !live.IsCreature() || !live.IsArtifact() || !live.HasSubtype("Elemental") {
		t.Fatalf("Stalking Stones is %q, want a land artifact creature Elemental", live.TypeLine)
	}
	if p, tough := effectstest.PtOf(t, g, land); p != 3 || tough != 3 {
		t.Errorf("Stalking Stones is %d/%d, want 3/3", p, tough)
	}
	// "This effect lasts indefinitely": still a creature next turn.
	effectstest.AdvanceToMainOf(t, g, 1)
	live = effectstest.ApaLive(g, land)
	if !live.IsCreature() || !live.HasSubtype("Elemental") {
		t.Error("the effect ended with the turn")
	}
}
