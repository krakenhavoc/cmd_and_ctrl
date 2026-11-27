package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b563_cards_test.go — Coldsnap's cantrips and library cards, played
// from the real catalog entries (S82 OLD batch 563).

const (
	b563RageOracle     = "1bc7d73b-0f54-4807-b13a-3def0c5e5616"
	b563NorthOracle    = "30aefabc-5fd2-451b-aa93-8e5a972c7691"
	b563MeltingOracle  = "850fe8ad-f6b8-4d01-a94b-2b69adc3d8f0"
	b563FluxOracle     = "90fe98df-4f87-49ff-b562-0ed51d1d3a1c"
	b563AurochsOracle  = "b07bd1cd-4000-4e6b-b16b-283c76593377"
	b563DrelnochOracle = "c25465a9-a0f3-48be-8bdf-4aa861da7d3e"
	b563ManeuverOracle = "d7b5fb4b-8f3c-4143-9fb0-2e2544612528"
)

func TestS82OldB563CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b563RageOracle:     "Balduvian Rage",
		b563NorthOracle:    "Into the North",
		b563MeltingOracle:  "Mystic Melting",
		b563FluxOracle:     "Thermal Flux",
		b563AurochsOracle:  "Aurochs Herd",
		b563DrelnochOracle: "Drelnoch",
		b563ManeuverOracle: "Swift Maneuver",
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

// b563DrawsAtTheNextUpkeep asserts `p` draws one card at the next turn's upkeep
// and not before.
func b563DrawsAtTheNextUpkeep(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	before := p.Hand.Size()
	next := (g.Turn.ActiveSeat + 1) % len(g.Seats)
	effectstest.AdvanceToUpkeepOf(t, g, next)
	if p.Hand.Size() != before {
		t.Fatalf("drew before the next upkeep's trigger resolved: hand %d -> %d", before, p.Hand.Size())
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p.Hand.Size() != before+1 {
		t.Errorf("hand %d -> %d at the next upkeep, want +1", before, p.Hand.Size())
	}
}

func TestBalduvianRagePumpsAnAttackerByXAndDrawsLater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	atk := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	idle := effectstest.PushVanillaCreature(g, me.ID, "Idle", 2, 2)
	effectstest.Pr7bAttack(t, g, opp.ID, atk)

	cast := func(target uuid.UUID, x int) error {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Balduvian Rage", TypeLine: "Instant", OracleID: b563RageOracle, Owner: me.ID, Controller: me.ID})
		return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: x, Targets: effectstest.CardTarget(target)})
	}
	if err := cast(idle, 3); err == nil {
		t.Fatal("targeted a creature that is not attacking: accepted")
	}
	if err := cast(atk, 3); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.ResolveTopOfStack(t, g)
	if p, tough := effectstest.PtOf(t, g, atk); p != 5 || tough != 2 {
		t.Errorf("the attacker is %d/%d, want 5/2", p, tough)
	}
	b563DrawsAtTheNextUpkeep(t, g, me)
}

func TestIntoTheNorthFindsASnowLandTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	plain := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
	snow := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Snow-Covered Forest", TypeLine: "Basic Snow Land — Forest"})
	_ = plain
	effectstest.CastCatalogSpell(t, g, "Into the North", "Sorcery", b563NorthOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SearchChoiceFor(g, me.ID); c != nil {
		if id := effectstest.SearchOptionNamed(g, c, "Mountain"); id != uuid.Nil {
			t.Fatal("a nonsnow land is offered")
		}
		effectstest.AnswerSearchNamed(t, g, me.ID, "Snow-Covered Forest")
	} else {
		t.Fatal("no search prompt")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.FindBattlefieldCardByID(g, snow)
	if c == nil {
		t.Fatal("the snow land is not on the battlefield")
	}
	if !c.Tapped {
		t.Error("the snow land did not enter tapped")
	}
}

func TestMysticMeltingDestroysAnArtifactOrEnchantmentAndDrawsLater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Mystic Melting", "Instant", b563MeltingOracle, effectstest.CardTarget(bear)); err == nil {
		t.Fatal("targeted a creature: accepted")
	}
	effectstest.CastCatalogSpell(t, g, "Mystic Melting", "Instant", b563MeltingOracle, effectstest.CardTarget(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
	b563DrawsAtTheNextUpkeep(t, g, me)
}

func TestThermalFluxBothModes(t *testing.T) {
	for mode := 0; mode < 2; mode++ {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.SeedLibrary(me, "A", "B", "C", "D")
		plain := effectstest.BlbB1Push(g, me.ID, "Mountain", "Basic Land — Mountain", "", 0, 0)
		snow := effectstest.BlbB1Push(g, me.ID, "Snow-Covered Forest", "Basic Snow Land — Forest", "", 0, 0)
		target, other := plain, snow
		if mode == 1 {
			target, other = snow, plain
		}
		effectstest.AdvanceToMain(t, g)
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Thermal Flux", TypeLine: "Instant", OracleID: b563FluxOracle, Owner: me.ID, Controller: me.ID})
		// The mode's own target filter refuses the other kind of permanent.
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{mode}, Targets: effectstest.CardTarget(other)}); err == nil {
			t.Fatalf("mode %d: targeted the wrong kind of permanent: accepted", mode)
		}
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{mode}, Targets: effectstest.CardTarget(target)}); err != nil {
			t.Fatalf("mode %d: cast: %v", mode, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		var isSnow, isLand bool
		g.WithWriteLock(func() {
			g.RecomputeLayersIfStaleLocked()
			c := effectstest.FindBattlefieldCardByID(g, target)
			isSnow, isLand = c.HasSupertype("snow"), c.IsLand()
		})
		if !isLand {
			t.Errorf("mode %d: the permanent stopped being a land", mode)
		}
		if isSnow != (mode == 0) {
			t.Errorf("mode %d: snow = %v", mode, isSnow)
		}
		b563DrawsAtTheNextUpkeep(t, g, me)
	}
}

func TestThermalFluxSnowLastsOnlyUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	plain := effectstest.BlbB1Push(g, me.ID, "Mountain", "Basic Land — Mountain", "", 0, 0)
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Thermal Flux", TypeLine: "Instant", OracleID: b563FluxOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{0}, Targets: effectstest.CardTarget(plain)}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if effectstest.FindBattlefieldCardByID(g, plain).HasSupertype("snow") {
		t.Error("the land is still snow on the next turn")
	}
}

func TestSwiftManeuverPreventsTwoAndDrawsLater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Swift Maneuver", "Instant", b563ManeuverOracle, []game.TargetRef{effectstest.PlayerTarget(me.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(uuid.New(), me.ID, 5) })
	if got := life - me.Life; got != 3 {
		t.Errorf("5 damage into a 2-shield: lost %d life, want 3", got)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(uuid.New(), me.ID, 1) })
	if got := life - me.Life; got != 4 {
		t.Errorf("the shield is spent: lost %d life, want 4", got)
	}
	b563DrawsAtTheNextUpkeep(t, g, me)
}

// b563Combat attacks with `attacker`, has `blocker` (when set) block it and
// answers every prompt the combat raises: "yes" to a "you may".
func b563Combat(t *testing.T, g *game.Game, attacker, blocker uuid.UUID, yes bool) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	attackers := []uuid.UUID{attacker}
	if blocker == uuid.Nil {
		attackers = []uuid.UUID{attacker}
	}
	for _, a := range attackers {
		if err := g.DeclareAttacker(a, opp.ID); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if blocker != uuid.Nil {
		if err := g.DeclareBlocker(blocker, attacker); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.FinishBlockDeclarations(t, g)
	for i := 0; i < 30; i++ {
		var open *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil {
				open = c
				break
			}
		}
		if open == nil {
			if len(g.StackMeta) == 0 && len(g.PendingTriggers) == 0 && i > 2 {
				return
			}
			if err := g.PassPriority(); err != nil {
				return
			}
			continue
		}
		switch open.Kind {
		case game.PendingChoiceTriggerPrompt:
			if err := g.ResolveTriggerPrompt(open.ID, me.ID, yes); err != nil {
				t.Fatalf("trigger prompt: %v", err)
			}
		default:
			t.Fatalf("unexpected prompt %s", open.Kind)
		}
	}
}

func TestDrelnochDrawsTwoWhenBlockedIfYouMay(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")
		id := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Drelnoch", TypeLine: "Creature — Yeti Mutant", OracleID: b563DrelnochOracle, Power: 3, Toughness: 3})
		wall := effectstest.PushVanillaCreature(g, opp.ID, "Wall", 0, 9)
		hand := me.Hand.Size()
		b563Combat(t, g, id, wall, yes)
		want := 0
		if yes {
			want = 2
		}
		if got := me.Hand.Size() - hand; got != want {
			t.Errorf("yes=%v: drew %d cards, want %d", yes, got, want)
		}
	}
}

func TestAurochsHerdSearchesForAnAurochs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pup := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Aurochs Pup", TypeLine: "Creature — Aurochs"})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	effectstest.CastCatalogSpell(t, g, "Aurochs Herd", "Creature — Aurochs", b563AurochsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("a non-Aurochs card is offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Aurochs Pup")
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(pup) {
		t.Error("the Aurochs card is not in hand")
	}
}

func TestAurochsHerdGrowsForEachOtherAttackingAurochs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	herd := func() uuid.UUID {
		return effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Aurochs Herd", TypeLine: "Creature — Aurochs", OracleID: b563AurochsOracle, Power: 4, Toughness: 4, Keywords: []string{"trample"}})
	}
	a, b := herd(), herd()
	bear := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{a, b, bear} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.LockInAttacks(t, g)
	if !effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID) {
		t.Fatal("the two Aurochs triggers were not offered for ordering")
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if p, _ := effectstest.PtOf(t, g, id); p != 5 {
			t.Errorf("an Aurochs Herd has power %d, want 5 (4 plus the other Aurochs)", p)
		}
	}
	if p, _ := effectstest.PtOf(t, g, bear); p != 2 {
		t.Errorf("the Bear has power %d, want 2", p)
	}
}
