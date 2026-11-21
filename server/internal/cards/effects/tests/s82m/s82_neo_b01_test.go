package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	neoB01Fang       = "9e187d43-ae75-4589-81fc-476ae95ba5b9"
	neoB01Disciple   = "316f7091-0c75-4510-a89a-6b58bb7f0627"
	neoB01BruteSuit  = "bc64e396-7fa1-4a4e-a528-352737714a9c"
	neoB01Dragonfly  = "f242d66c-011c-4b5b-a7d6-5e0bce257f73"
	neoB01Sentinel   = "78d9faff-2ad9-4cf9-a239-27c50435d007"
	neoB01Colossus   = "9cef3013-6a84-4edd-9a47-e842df2200e3"
	neoB01Infiltrate = "6a45d09a-6af3-45fb-8f7c-cf630cb37f24"
	neoB01Patrol     = "c4073754-7d92-4976-8e7d-d6eb85b91608"
	neoB01Trainee    = "cb9fa34d-1d5f-461c-acad-bc573d41ceb2"
	neoB01Bladeb     = "408512c7-5016-4df4-b708-b7467bcdf0ae"
	neoB01Puzzle     = "8ee2371c-e2c3-412e-b9e9-1ec511c9d537"
	neoB01Songshaper = "d2c7d10e-86d9-4e2d-a0da-2e39c172aae2"
	neoB01Hoverbike  = "9bf2f0d9-57fb-4e44-90a5-260d735076a0"
)

func TestNeoB01EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		neoB01Fang: "Fang of Shigeki", neoB01Disciple: "Golden-Tail Disciple", neoB01BruteSuit: "Brute Suit",
		neoB01Dragonfly: "Dragonfly Suit", neoB01Sentinel: "Futurist Sentinel", neoB01Colossus: "Thundersteel Colossus",
		neoB01Infiltrate: "Inkrise Infiltrator", neoB01Patrol: "Mothrider Patrol", neoB01Trainee: "Jukai Trainee",
		neoB01Bladeb: "Nezumi Bladeblesser", neoB01Puzzle: "Moonfolk Puzzlemaker", neoB01Songshaper: "Towashi Songshaper",
		neoB01Hoverbike: "High-Speed Hoverbike",
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

func TestNeoB01PrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		oracle, name, typeLine string
		p, tough               int
		kws                    []string
	}{
		{neoB01Fang, "Fang of Shigeki", "Enchantment Creature — Snake Ninja", 1, 1, []string{"deathtouch"}},
		{neoB01Disciple, "Golden-Tail Disciple", "Enchantment Creature — Fox Monk", 2, 3, []string{"lifelink"}},
		{neoB01Infiltrate, "Inkrise Infiltrator", "Creature — Human Ninja", 1, 2, []string{"flying"}},
		{neoB01Patrol, "Mothrider Patrol", "Creature — Fox Warrior", 1, 1, []string{"flying"}},
		{neoB01Puzzle, "Moonfolk Puzzlemaker", "Artifact Creature — Moonfolk Wizard", 1, 4, []string{"flying"}},
	} {
		g := effectstest.NewCatalogGame(t)
		id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, tc.p, tc.tough)
		for _, kw := range tc.kws {
			if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
				t.Errorf("%s lacks %s", tc.name, kw)
			}
		}
	}
}

func TestNeoB01VehiclesAreInertUntilCrewedAndNeedEnoughPower(t *testing.T) {
	for _, tc := range []struct {
		oracle, name string
		p, tough     int
		crew         int
		kws          []string
	}{
		{neoB01BruteSuit, "Brute Suit", 4, 3, 1, []string{"vigilance"}},
		{neoB01Dragonfly, "Dragonfly Suit", 3, 2, 1, []string{"flying"}},
		{neoB01Sentinel, "Futurist Sentinel", 6, 6, 3, nil},
		{neoB01Colossus, "Thundersteel Colossus", 7, 7, 2, []string{"trample", "haste"}},
		{neoB01Hoverbike, "High-Speed Hoverbike", 2, 2, 1, []string{"flying", "flash"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			v := effectstest.B12Push(g, me.ID, tc.name, "Artifact — Vehicle", tc.oracle, tc.p, tc.tough)
			weak := effectstest.B12Creature(g, me.ID, "Weak", "Creature — Test", tc.crew-1, 1)
			strong := effectstest.B12Creature(g, me.ID, "Strong", "Creature — Test", tc.crew, 1)
			effectstest.AdvanceToMain(t, g)
			if effectstest.IsCreatureNow(g, v) {
				t.Fatal("a Vehicle is a creature before it is crewed")
			}
			for _, kw := range tc.kws {
				if !effectstest.HasEffectiveKeyword(t, g, v, kw) {
					t.Errorf("lacks %s", kw)
				}
			}
			spec, _ := effects.Lookup(tc.oracle)
			idx := len(spec.Activated) - 1
			if err := g.ActivateCatalogAbility(me.ID, v, idx, game.ActivateAbilityParams{Strict: true, CrewIDs: []uuid.UUID{weak}}); err == nil && tc.crew > 1 {
				t.Fatal("crewed with too little power")
			}
			if err := g.ActivateCatalogAbility(me.ID, v, idx, game.ActivateAbilityParams{Strict: true, CrewIDs: []uuid.UUID{strong}}); err != nil {
				t.Fatalf("crew %d: %v", tc.crew, err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if !effectstest.IsCreatureNow(g, v) {
				t.Error("the crewed Vehicle is not a creature")
			}
			if p, th := effectstest.EffectivePTOf(t, g, v); p != tc.p || th != tc.tough {
				t.Errorf("P/T %d/%d, want %d/%d", p, th, tc.p, tc.tough)
			}
		})
	}
}

func TestNeoB01HoverbikeTapsUpToOneCreatureWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.FloatForTest(g, me, "CC")
	effectstest.CastFromHandForTest(t, g, me, "High-Speed Hoverbike", "Artifact — Vehicle", "{2}", neoB01Hoverbike,
		game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the tap")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(victim)); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, victim); !tapped {
		t.Error("the targeted creature was not tapped")
	}
}

func TestNeoB01HoverbikeEntersWithNoTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.FloatForTest(g, me, "CC")
	id := effectstest.CastFromHandForTest(t, g, me, "High-Speed Hoverbike", "Artifact — Vehicle", "{2}", neoB01Hoverbike,
		game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Error("the Hoverbike did not enter with no target")
	}
}

func TestNeoB01JukaiTraineeGrowsWhenItBlocksOrBecomesBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tr := effectstest.B12Push(g, me.ID, "Jukai Trainee", "Creature — Human Samurai", neoB01Trainee, 2, 2)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 1, 1)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBlock, Actor: me.ID, CardID: tr, Target: a, Amount: 1})
		g.EmitEvent(game.Event{Kind: game.EventBlock, Actor: me.ID, CardID: tr, Target: b, Amount: 2})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, tr); p != 3 || th != 3 {
		t.Fatalf("after blocking two attackers it is %d/%d, want 3/3: blocks triggers once", p, th)
	}

	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	tr = effectstest.B12Push(g, me.ID, "Jukai Trainee", "Creature — Human Samurai", neoB01Trainee, 2, 2)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBecomesBlocked, Actor: opp.ID, Source: tr, CardID: tr, Target: tr})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, tr); p != 3 || th != 3 {
		t.Errorf("after becoming blocked it is %d/%d, want 3/3", p, th)
	}
}

func TestNeoB01BladeblesserGainsEachKeywordOnlyWithItsPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bb := effectstest.B12Push(g, me.ID, "Nezumi Bladeblesser", "Creature — Rat Samurai", neoB01Bladeb, 3, 2)
	effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.B12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	if effectstest.HasEffectiveKeyword(t, g, bb, "deathtouch") || effectstest.HasEffectiveKeyword(t, g, bb, "menace") {
		t.Fatal("an opponent's permanents grant a keyword")
	}
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	if !effectstest.HasEffectiveKeyword(t, g, bb, "deathtouch") || effectstest.HasEffectiveKeyword(t, g, bb, "menace") {
		t.Error("an artifact gives deathtouch and not menace")
	}
	ench := effectstest.B12Permanent(g, me.ID, "My Aura", "Enchantment")
	if !effectstest.HasEffectiveKeyword(t, g, bb, "deathtouch") || !effectstest.HasEffectiveKeyword(t, g, bb, "menace") {
		t.Error("an artifact and an enchantment give both")
	}
	effectstest.B15Destroy(t, g, rock)
	if effectstest.HasEffectiveKeyword(t, g, bb, "deathtouch") || !effectstest.HasEffectiveKeyword(t, g, bb, "menace") {
		t.Error("deathtouch ends with the last artifact, menace stays")
	}
	effectstest.B15Destroy(t, g, ench)
	if effectstest.HasEffectiveKeyword(t, g, bb, "menace") {
		t.Error("menace ends with the last enchantment")
	}
}

func TestNeoB01PuzzlemakerScriesWhenItBecomesTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pz := effectstest.B12Push(g, me.ID, "Moonfolk Puzzlemaker", "Artifact Creature — Moonfolk Wizard", neoB01Puzzle, 1, 4)
	effectstest.AdvanceToMain(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) != nil {
		t.Fatal("a scry is owed before it has tapped")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(pz, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	sc := effectstest.ScryChoiceFor(g, me.ID)
	if sc == nil || len(sc.ScryCards) != 1 {
		t.Fatalf("attacking tapped it: a scry 1 is owed, got %+v", sc)
	}
}

func TestNeoB01InkriseInfiltratorPumpsTwoTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ink := effectstest.B12Push(g, me.ID, "Inkrise Infiltrator", "Creature — Human Ninja", neoB01Infiltrate, 1, 2)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.UsgB2Activate(t, g, me, ink, "CCC", nil); err == nil {
		t.Fatal("activated without {B}")
	}
	if err := effectstest.UsgB2Activate(t, g, me, ink, "CCCB", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, ink); p != 3 || th != 4 {
		t.Errorf("P/T %d/%d, want 3/4", p, th)
	}
}

func TestNeoB01MothriderPatrolTapsATargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	patrol := effectstest.B12Push(g, me.ID, "Mothrider Patrol", "Creature — Fox Warrior", neoB01Patrol, 1, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.UsgB2Activate(t, g, me, patrol, "CCW", effectstest.CardRefs(victim)); err == nil {
		t.Fatal("activated without enough mana")
	}
	if err := effectstest.UsgB2Activate(t, g, me, patrol, "CCCW", effectstest.CardRefs(victim)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, victim); !tapped {
		t.Error("the target was not tapped")
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, patrol); !tapped {
		t.Error("the Patrol did not tap as a cost")
	}
}

func TestNeoB01SongshaperGrowsOnlyForAnotherArtifactYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	song := effectstest.B12Push(g, me.ID, "Towashi Songshaper", "Artifact Creature — Human Artificer", neoB01Songshaper, 2, 2)
	effectstest.AdvanceToMain(t, g)
	enter := func(owner uuid.UUID, typeLine string) {
		id := effectstest.B12Permanent(g, owner, "Thing", typeLine)
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventETB, Actor: owner, CardID: id})
		})
		effectstest.PassPriorityAroundTable(t, g)
	}
	enter(opp.ID, "Artifact")
	enter(me.ID, "Enchantment")
	if p, _ := effectstest.EffectivePTOf(t, g, song); p != 2 {
		t.Fatalf("power %d, want 2 after non-qualifying entries", p)
	}
	enter(me.ID, "Artifact")
	if p, th := effectstest.EffectivePTOf(t, g, song); p != 3 || th != 2 {
		t.Errorf("P/T %d/%d, want 3/2", p, th)
	}
	enter(me.ID, "Artifact Creature — Robot")
	if p, _ := effectstest.EffectivePTOf(t, g, song); p != 4 {
		t.Errorf("power %d, want 4 after a second artifact", p)
	}
}
