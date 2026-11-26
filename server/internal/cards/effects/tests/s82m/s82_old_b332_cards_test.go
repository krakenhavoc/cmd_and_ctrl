package testss82m

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 Dragons of Tarkir batch 332 (#5551): mana and ramp, Auras.

const (
	dtk332Silumgar = "0b88268c-7fcb-4fd2-96cd-f93b37413283"
	dtk332Dromoka  = "1c587200-02cc-4c5f-86fe-d13cc7bdad8b"
	dtk332Ojutai   = "d6b4810f-2bee-4c83-b13d-c46fb7dd8e08"
	dtk332Kolaghan = "d8b79fed-9817-4b91-ac2e-e9fa955dd43c"
	dtk332Qarsi    = "e12681cc-c66e-4dac-882c-bd8638a2e12c"
	dtk332Gains    = "1ba45691-417b-4e6a-be45-6027e4f04599"
	dtk332Mockery  = "6a802318-9123-4058-8554-5023eace33f4"
	dtk332Reduce   = "7ebaeef0-4e4c-4bf0-85dd-07c6b6be0741"
	dtk332Encase   = "a61277d6-a27d-45d1-ba2a-7afe1b8f5ce1"
	dtk332Aerie    = "f2c408f6-3994-4a90-a9f5-f248eb976017"
)

func TestDtk332CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		dtk332Silumgar: "Silumgar Monument", dtk332Dromoka: "Dromoka Monument", dtk332Ojutai: "Ojutai Monument",
		dtk332Kolaghan: "Kolaghan Monument", dtk332Qarsi: "Qarsi Deceiver", dtk332Gains: "Illusory Gains",
		dtk332Mockery: "Mirror Mockery", dtk332Reduce: "Reduce in Stature", dtk332Encase: "Encase in Ice",
		dtk332Aerie: "Sheltered Aerie",
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

func TestDtk332MonumentsTapForTheirTwoColoursAndBecomeFourFourDragons(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		a, b         string
		colours      []string
	}{
		{"Silumgar Monument", dtk332Silumgar, "U", "B", []string{"U", "B"}},
		{"Dromoka Monument", dtk332Dromoka, "G", "W", []string{"G", "W"}},
		{"Ojutai Monument", dtk332Ojutai, "W", "U", []string{"W", "U"}},
		{"Kolaghan Monument", dtk332Kolaghan, "B", "R", []string{"B", "R"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.AdvanceToMain(t, g)
			mon := effectstest.PushCatalogPermanent(g, me.ID, tc.name, "Artifact", tc.oracle, false)
			// Its mana ability offers exactly its two colours.
			for _, colour := range []string{tc.a, tc.b} {
				if err := g.ActivateManaAbility(me.ID, mon, 0, game.ManaAbilityParams{Colors: []string{colour}}); err != nil {
					t.Fatalf("tap for %s: %v", colour, err)
				}
				if me.ManaPool.Total() != 1 || me.ManaPool.Units()[0].Color != colour {
					t.Fatalf("pool = %v, want one %s", me.ManaPool.Units(), colour)
				}
				g.WithWriteLock(func() { me.ManaPool = game.ManaPool{} })
				effectstest.Untap(g, mon)
			}
			// The animation: {4} and its two colours, no less.
			if err := g.ActivateCatalogAbility(me.ID, mon, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
				t.Fatal("animated with an empty pool")
			}
			effectstest.FloatForTest(g, me, "CCCC"+tc.a)
			if err := g.ActivateCatalogAbility(me.ID, mon, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
				t.Fatal("animated without the second colour")
			}
			effectstest.FloatForTest(g, me, tc.b)
			if err := g.ActivateCatalogAbility(me.ID, mon, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
				t.Fatalf("animate: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			ch := effectstest.LayeredCard(t, g, mon)
			if !ch.IsCreature() || !ch.IsArtifact() || !ch.HasSubtype("Dragon") {
				t.Errorf("type line %q, want an artifact Dragon creature", ch.TypeLine)
			}
			if p, tough := effectstest.EffectivePTOf(t, g, mon); p != 4 || tough != 4 {
				t.Errorf("%d/%d, want 4/4", p, tough)
			}
			if !effectstest.HasKeywordOf(g, mon, "flying") {
				t.Error("no flying")
			}
			if !slices.Equal(ch.EffectiveColors(), tc.colours) {
				t.Errorf("colours %v, want %v", ch.EffectiveColors(), tc.colours)
			}
			// Until end of turn only.
			effectstest.AdvanceToUpkeepOf(t, g, 1)
			if ch := effectstest.LayeredCard(t, g, mon); ch.IsCreature() {
				t.Error("still a creature after the turn ended")
			}
		})
	}
}

// dtk332QarsiMana floats n Qarsi Deceiver mana (each a restricted {C}) for the seat.
func dtk332QarsiMana(t *testing.T, g *game.Game, me *game.Player, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		qarsi := effectstest.PushCatalogPermanent(g, me.ID, "Qarsi Deceiver", "Creature — Snake Wizard", dtk332Qarsi, false)
		effectstest.Unsick(g, qarsi)
		if err := g.ActivateManaAbility(me.ID, qarsi, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("tap Qarsi Deceiver: %v", err)
		}
	}
	if me.ManaPool.Total() != n {
		t.Fatalf("pool has %d mana, want %d", me.ManaPool.Total(), n)
	}
	for _, u := range me.ManaPool.Units() {
		if len(u.Restrictions) == 0 {
			t.Fatalf("Qarsi's mana is unrestricted: %+v", u)
		}
	}
}

func TestDtk332QarsiDeceiverManaPaysAFaceDownCastButNotAFaceUpOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMainForCost(t, g)
	dtk332QarsiMana(t, g, me, 3)
	faceUp := effectstest.HandCardFull(me, "Test Bear", "Creature — Bear", "{3}", "", nil)
	if err := g.CastSpell(me.ID, faceUp, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("Qarsi's mana paid a creature spell cast face up")
	}
	morph := effectstest.HandCardFull(me, "Willbender", "Creature — Human Wizard", "{1}{U}", effectstest.WillbenderOracle, []string{"U"})
	if err := g.CastSpell(me.ID, morph, game.CastSpellParams{Strict: true, AlternativeCost: "morph"}); err != nil {
		t.Fatalf("Qarsi's mana did not pay the {3} of a face-down cast: %v", err)
	}
}

func TestDtk332QarsiDeceiverManaPaysAMorphCostToTurnFaceUpOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.CastWithAltCost(t, g, "Willbender", "Creature — Human Wizard", effectstest.WillbenderOracle, "morph")
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindCardAnywhere(t, g, id); !c.FaceDownIsPermanent() {
		t.Fatal("the Willbender did not enter face down")
	}
	// Willbender's morph cost is {1}{U}: Qarsi's {C}s can pay the {1}, never the {U}.
	dtk332QarsiMana(t, g, me, 2)
	if err := g.PerformSpecialAction(me.ID, id, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true}); err == nil {
		t.Fatal("two {C} paid a {1}{U} morph cost")
	}
	effectstest.FloatForTest(g, me, "U")
	if err := g.PerformSpecialAction(me.ID, id, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn face up with Qarsi's {C} and an unrestricted {U}: %v", err)
	}
	if up := effectstest.FindCardAnywhere(t, g, id); up.FaceDown {
		t.Error("still face down")
	}
}

func TestDtk332QarsiDeceiverManaPaysNothingElse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	dtk332QarsiMana(t, g, me, 1)
	g.WithWriteLock(func() { me.ManaPool.AddMana(effectstest.ManaTokens("U")...) })
	// An ordinary cost that needs the {C}: a face-up {1}{U} spell.
	spell := effectstest.HandCardFull(me, "Plain Spell", "Sorcery", "{1}{U}", "", []string{"U"})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("Qarsi's {C} paid a face-up spell")
	}
}

func TestDtk332ReduceInStatureSetsBasePTUnderCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 5, 5)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(bear, game.CounterPlusOne, 2); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.Enchant(t, g, "Reduce in Stature", dtk332Reduce, bear)
	if p, tough := effectstest.PtOf(t, g, bear); p != 2 || tough != 4 {
		t.Errorf("P/T = %d/%d, want 2/4 (base 0/2 plus two +1/+1 counters)", p, tough)
	}
}

func TestDtk332EncaseInIceTapsLocksAndOnlyEnchantsRedOrGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	mk := func(name string, colors ...string) uuid.UUID {
		return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Test",
			Colors: colors, Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID})
	}
	blue := mk("Blue Guy", "U")
	green := mk("Green Guy", "G")
	red := mk("Red Guy", "R")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Encase in Ice", effectstest.AuraTypeLine, dtk332Encase, effectstest.B16TargetCard(blue)); err == nil {
		t.Fatal("Encase in Ice enchanted a blue creature")
	}
	effectstest.Enchant(t, g, "Encase in Ice", dtk332Encase, green)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, green); !c.Tapped {
		t.Fatal("the enchanted creature is not tapped")
	}
	effectstest.Enchant(t, g, "Encase in Ice", dtk332Encase, red)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, red); !c.Tapped {
		t.Fatal("the red creature is not tapped")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if c, _ := effectstest.BattlefieldCard(g, green); !c.Tapped {
		t.Error("the green creature untapped in its controller's untap step")
	}
	if c, _ := effectstest.BattlefieldCard(g, red); !c.Tapped {
		t.Error("the red creature untapped in its controller's untap step")
	}
}

func TestDtk332ShelteredAerieMakesTheLandTapForTwoManaOfOneColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
		Owner: me.ID, Controller: me.ID})
	effectstest.AuraCast(t, g, "Sheltered Aerie", dtk332Aerie, land)
	if err := effectstest.TapGrantedMana(t, g, me.ID, land, "U"); err != nil {
		t.Fatalf("tap the enchanted land: %v", err)
	}
	if me.ManaPool.Total() != 2 {
		t.Fatalf("pool has %d mana, want 2", me.ManaPool.Total())
	}
	for _, u := range me.ManaPool.Units() {
		if u.Color != "U" {
			t.Errorf("mana of colour %q, want both U", u.Color)
		}
	}
}

func TestDtk332IllusoryGainsStealsAndMovesToEachNewOpposingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	first := effectstest.PushVanillaCreature(g, opp.ID, "First", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.AuraCast(t, g, "Illusory Gains", dtk332Gains, first)
	if c := effectstest.LayeredCard(t, g, first); c.Controller != me.ID {
		t.Fatal("I do not control the enchanted creature")
	}
	// A creature entering under my control does not move it.
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Mine", TypeLine: "Creature — Bear", Power: 1, Toughness: 1})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, first); c.Controller != me.ID {
		t.Error("my own creature entering moved the Aura")
	}
	// A creature entering under an opponent's control does.
	second := effectstest.EnterCard(t, g, opp.ID, game.Card{Name: "Second", TypeLine: "Creature — Bear", Power: 3, Toughness: 3})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, second); c.Controller != me.ID {
		t.Error("the new creature did not come under my control")
	}
	if c := effectstest.LayeredCard(t, g, first); c.Controller != opp.ID {
		t.Error("the first creature did not go back to its owner")
	}
}

func TestDtk332MirrorMockeryCopiesTheAttackerAndExilesTheCopyAtEndOfCombat(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		host := effectstest.PushCatalogPermanent(g, me.ID, "Copy Target", "Creature — Test", "", false)
		effectstest.Unsick(g, host)
		effectstest.AuraCast(t, g, "Mirror Mockery", dtk332Mockery, host)
		effectstest.DeclareAttack(t, g, opp.ID, host)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerConfirmFor(t, g, me, yes)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if yes {
			want = 1
		}
		if got := effectstest.B28TokensNamed(g, me.ID, "Copy Target"); got != want {
			t.Fatalf("yes=%v: %d token copies, want %d", yes, got, want)
		}
		effectstest.AdvanceTo(t, g, game.StepEndCombat)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.B28TokensNamed(g, me.ID, "Copy Target"); got != 0 {
			t.Errorf("yes=%v: %d token copies after end of combat, want 0", yes, got)
		}
		if !g.Battlefield.Contains(host) {
			t.Errorf("yes=%v: the attacking creature itself left the battlefield", yes)
		}
	}
}
