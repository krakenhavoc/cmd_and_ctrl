package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b526_cards_test.go — Eventide, S82 older sets batch 526 (#5590).

const (
	eve2Recruit  = "03a0f7ac-de8a-43cf-be15-07d31e9d6188"
	eve2Incurs   = "3e734759-d43a-4bbe-baff-26d866ff0923"
	eve2Invert   = "ea5ef11d-635f-4f62-908d-6d14465eb921"
	eve2Riverf   = "16948dd9-959b-4972-b337-37ecc2b141bd"
	eve2Shorec   = "3da837db-b3e9-46e4-939d-54ab43af94d9"
	eve2Nightsky = "ed186aea-e0dc-4d88-a619-748b2ebf0c52"
	eve2Dream    = "55be0fd0-b371-4b9e-920f-3123bb37a1e0"
	eve2Idle     = "a2a40a19-8704-42fe-b9a6-5ae37c110303"
	eve2Horizons = "ca945c1e-efa3-4bb2-bf5a-379b93409855"
	eve2Drain    = "84e89aac-439c-4e09-b7d9-2ee047c224c5"
	eve2Giant    = "be109ed5-9e3e-40b3-9bf0-7caf7cdbb958"
	eve2Reap     = "d65f701d-31a1-4275-a580-3e1799ebf8b9"
)

func TestS82OldB526EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		eve2Recruit: "Unwilling Recruit", eve2Incurs: "Nightmare Incursion", eve2Invert: "Invert the Skies",
		eve2Riverf: "Riverfall Mimic", eve2Shorec: "Shorecrasher Mimic", eve2Nightsky: "Nightsky Mimic",
		eve2Dream: "Dream Thief", eve2Idle: "Idle Thoughts", eve2Horizons: "Endless Horizons",
		eve2Drain: "Drain the Well", eve2Giant: "Marshdrinker Giant", eve2Reap: "Soul Reap",
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

// eve2Cast casts a card from my hand for real, with the printed cost, colours
// and the pool floated, in a main phase, and settles the stack.
func eve2Cast(t *testing.T, g *game.Game, me *game.Player, name, typeLine, oracle, cost string, colors []string, pool string, targets []game.TargetRef, x int) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: cost,
		Colors: colors, Owner: me.ID, Controller: me.ID})
	if pool != "" {
		effectstest.AddPool(t, g, me.ID, pool)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: targets, XValue: x}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestS82OldB526UnwillingRecruitStealsUntapsAndPumpsByX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	bear := effectstest.B16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.TapForTest(t, g, bear)
	eve2Cast(t, g, me, "Unwilling Recruit", "Sorcery", eve2Recruit, "{X}{R}{R}{R}", []string{"R"}, "{C}{C}{R}{R}{R}", effectstest.CardRefs(bear), 2)
	c, _ := effectstest.BattlefieldCard(g, bear)
	if c.Controller != me.ID {
		t.Fatal("control did not change")
	}
	if c.Tapped {
		t.Error("the creature was not untapped")
	}
	if p, _ := eve2Stats(t, g, bear); p != 4 {
		t.Errorf("power %d, want 2+X=4", p)
	}
	effectstest.AssertKeywords(t, g, bear, "haste")
}

func TestS82OldB526NightmareIncursionSearchesUpToYourSwamps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	for i := 0; i < 2; i++ {
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Swamp", TypeLine: "Basic Land — Swamp", Owner: me.ID, Controller: me.ID})
	}
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Swamp", TypeLine: "Basic Land — Swamp", Owner: opp.ID, Controller: opp.ID})
	var ids []uuid.UUID
	for i := 0; i < 4; i++ {
		ids = append(ids, effectstest.SeedLibraryTop(opp, "Target Card", "Sorcery", ""))
	}
	eve2Cast(t, g, me, "Nightmare Incursion", "Sorcery", eve2Incurs, "{5}{B}", []string{"B"}, "{C}{C}{C}{C}{C}{B}", []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, 0)
	pc := effectstest.ChkB1OpenSearch(g)
	if pc == nil {
		t.Fatal("no search prompt")
	}
	if err := g.ResolveSearchLibrary(pc.ID, pc.Chooser, ids[:3]); err == nil {
		t.Fatal("took three cards with two Swamps (the opponent's Swamp does not count)")
	}
	if err := g.ResolveSearchLibrary(pc.ID, pc.Chooser, ids[:2]); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	exiled := 0
	for _, c := range g.Exile.Cards {
		if c.Owner == opp.ID {
			exiled++
		}
	}
	if exiled != 2 {
		t.Errorf("%d of the opponent's cards were exiled, want 2", exiled)
	}
}

func TestS82OldB526InvertTheSkiesReadsTheColoursSpent(t *testing.T) {
	for _, row := range []struct {
		pool             string
		oppFlies, youFly bool
	}{{"{C}{C}{C}{G}", false, false}, {"{C}{C}{C}{U}", true, true}, {"{C}{C}{G}{U}", false, true}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
		theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
			Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
		mine := effectstest.B16Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
		eve2Cast(t, g, me, "Invert the Skies", "Instant", eve2Invert, "{3}{G/U}", []string{"G", "U"}, row.pool, nil, 0)
		_ = row.oppFlies
		hasFlying := func(id uuid.UUID) bool {
			return effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, id), "flying")
		}
		wantTheirs := row.pool == "{C}{C}{C}{U}"
		if hasFlying(theirs) != wantTheirs {
			t.Errorf("%s: their flier has flying %v, want %v", row.pool, hasFlying(theirs), wantTheirs)
		}
		if hasFlying(mine) != row.youFly {
			t.Errorf("%s: my creature has flying %v, want %v", row.pool, hasFlying(mine), row.youFly)
		}
	}
}

func TestS82OldB526MimicsNeedASpellOfBothColours(t *testing.T) {
	for _, row := range []struct {
		name, oracle string
		colors       []string
		pt           [2]int
		kw           string
	}{
		{"Riverfall Mimic", eve2Riverf, []string{"U", "R"}, [2]int{3, 3}, ""},
		{"Shorecrasher Mimic", eve2Shorec, []string{"G", "U"}, [2]int{5, 3}, "trample"},
		{"Nightsky Mimic", eve2Nightsky, []string{"W", "B"}, [2]int{4, 4}, "flying"},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		mimic := effectstest.B12Push(g, me.ID, row.name, "Creature — Shapeshifter", row.oracle, 2, 1)
		// One of the two colours: nothing.
		eve2Cast(t, g, me, "Mono", "Instant", "", "", []string{row.colors[0]}, "", nil, 0)
		if p, _ := eve2Stats(t, g, mimic); p != 2 {
			t.Fatalf("%s: a one-colour spell changed the Mimic to power %d", row.name, p)
		}
		eve2Cast(t, g, me, "Gold", "Instant", "", "", row.colors, "", nil, 0)
		if p, tough := eve2Stats(t, g, mimic); p != row.pt[0] || tough != row.pt[1] {
			t.Errorf("%s is %d/%d, want %d/%d", row.name, p, tough, row.pt[0], row.pt[1])
		}
		if row.kw != "" {
			effectstest.AssertKeywords(t, g, mimic, row.kw)
		} else {
			effectstest.AssertRestrictions(t, g, mimic, game.CantBeBlocked)
		}
	}
}

func TestS82OldB526DreamThiefDrawsOnlyAfterAnotherBlueSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := len(me.Hand.Cards)
	eve2Cast(t, g, me, "Dream Thief", "Creature — Faerie Rogue", eve2Dream, "{2}{U}", []string{"U"}, "{C}{C}{U}", nil, 0)
	if got := len(me.Hand.Cards); got != before {
		t.Errorf("Dream Thief alone drew: hand %d → %d (its own cast is not another blue spell)", before, got)
	}
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	eve2Cast(t, g2, me2, "Blue Spell", "Instant", "", "", []string{"U"}, "", nil, 0)
	before = len(me2.Hand.Cards)
	eve2Cast(t, g2, me2, "Dream Thief", "Creature — Faerie Rogue", eve2Dream, "{2}{U}", []string{"U"}, "{C}{C}{U}", nil, 0)
	if got := len(me2.Hand.Cards); got != before+1 {
		t.Errorf("after another blue spell the hand is %d, want %d", got, before+1)
	}
	// A red spell does not count.
	g3 := effectstest.NewCatalogGame(t)
	me3 := g3.Seats[g3.Turn.ActiveSeat]
	eve2Cast(t, g3, me3, "Red Spell", "Instant", "", "", []string{"R"}, "", nil, 0)
	before = len(me3.Hand.Cards)
	eve2Cast(t, g3, me3, "Dream Thief", "Creature — Faerie Rogue", eve2Dream, "{2}{U}", []string{"U"}, "{C}{C}{U}", nil, 0)
	if got := len(me3.Hand.Cards); got != before {
		t.Errorf("a red spell made Dream Thief draw: %d → %d", before, got)
	}
}

func TestS82OldB526IdleThoughtsDrawsOnlyWithAnEmptyHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	thoughts := effectstest.B12Push(g, me.ID, "Idle Thoughts", "Enchantment", eve2Idle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	me.Hand.Cards = nil
	n := len(me.Library.Cards)
	effectstest.AddPool(t, g, me.ID, "{C}{C}")
	effectstest.B16Activate(t, g, me.ID, thoughts, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != 1 || len(me.Library.Cards) != n-1 {
		t.Fatalf("with an empty hand: hand %d, library %d → %d", len(me.Hand.Cards), n, len(me.Library.Cards))
	}
	effectstest.AddPool(t, g, me.ID, "{C}{C}")
	effectstest.B16Activate(t, g, me.ID, thoughts, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != 1 {
		t.Errorf("drew with a card in hand: hand %d", len(me.Hand.Cards))
	}
}

func TestS82OldB526DrainTheWellDestroysALandAndGainsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID})
	life := me.Life
	eve2Cast(t, g, me, "Drain the Well", "Sorcery", eve2Drain, "{2}{B/G}{B/G}", []string{"B", "G"}, "{C}{C}{B}{G}", effectstest.CardRefs(land), 0)
	if _, ok := effectstest.BattlefieldCard(g, land); ok {
		t.Error("the land was not destroyed")
	}
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2", life, me.Life)
	}
}

func TestS82OldB526MarshdrinkerGiantDestroysAnOpposingIslandOrSwamp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	forest := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID})
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	swamp := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Swamp", TypeLine: "Basic Land — Swamp", Owner: opp.ID, Controller: opp.ID})
	eve2Cast(t, g, me, "Marshdrinker Giant", "Creature — Giant Warrior", eve2Giant, "{3}{G}{G}", []string{"G"}, "{C}{C}{C}{G}{G}", nil, 0)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.B17PickCards(t, g, me.ID, swamp)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if _, ok := effectstest.BattlefieldCard(g, swamp); ok {
		t.Error("the opposing Swamp survived")
	}
	for _, id := range []uuid.UUID{forest, mine} {
		if _, ok := effectstest.BattlefieldCard(g, id); !ok {
			t.Error("a land that is not an opposing Island or Swamp was destroyed")
		}
	}
}

func TestS82OldB526SoulReapDestroysNongreenAndDrainsAfterAnotherBlackSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	green := effectstest.B16Creature(g, opp.ID, "Elf", "Creature — Elf", 1, 1, "G")
	if err := eve2TryCast(t, g, me, eve2Reap, green); err == nil {
		t.Fatal("Soul Reap targeted a green creature")
	}
	victim := effectstest.B16Creature(g, opp.ID, "Zombie", "Creature — Zombie", 2, 2, "B")
	life := opp.Life
	eve2Cast(t, g, me, "Soul Reap", "Sorcery", eve2Reap, "{1}{B}", []string{"B"}, "{C}{B}", effectstest.CardRefs(victim), 0)
	if _, ok := effectstest.BattlefieldCard(g, victim); ok {
		t.Error("the creature survived")
	}
	if opp.Life != life {
		t.Errorf("Soul Reap alone cost %d life", life-opp.Life)
	}
	// With another black spell cast first the controller loses 3.
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[g2.Turn.ActiveSeat], g2.Seats[1]
	v2 := effectstest.B16Creature(g2, opp2.ID, "Zombie", "Creature — Zombie", 2, 2, "B")
	eve2Cast(t, g2, me2, "Dark Spell", "Instant", "", "", []string{"B"}, "", nil, 0)
	life = opp2.Life
	eve2Cast(t, g2, me2, "Soul Reap", "Sorcery", eve2Reap, "{1}{B}", []string{"B"}, "{C}{B}", effectstest.CardRefs(v2), 0)
	if opp2.Life != life-3 {
		t.Errorf("life %d → %d, want -3", life, opp2.Life)
	}
}

func TestS82OldB526EndlessHorizonsExilesPlainsAndReturnsOneEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	var plains []uuid.UUID
	for i := 0; i < 3; i++ {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: "Plains", TypeLine: "Basic Land — Plains", Owner: me.ID, Controller: me.ID})
		plains = append(plains, id)
	}
	eve2Cast(t, g, me, "Endless Horizons", "Enchantment", eve2Horizons, "{3}{W}", []string{"W"}, "{C}{C}{C}{W}", nil, 0)
	pc := effectstest.ChkB1OpenSearch(g)
	if pc == nil {
		t.Fatal("no search prompt")
	}
	if err := g.ResolveSearchLibrary(pc.ID, pc.Chooser, plains[:2]); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	inExile := func(id uuid.UUID) bool {
		for _, c := range g.Exile.Cards {
			if c.InstanceID == id {
				return true
			}
		}
		return false
	}
	if !inExile(plains[0]) || !inExile(plains[1]) || inExile(plains[2]) {
		t.Fatalf("exile holds the wrong Plains: %v %v %v", inExile(plains[0]), inExile(plains[1]), inExile(plains[2]))
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("no upkeep prompt for a card exiled with Horizons")
	}
	hand := len(me.Hand.Cards)
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{plains[0]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if len(me.Hand.Cards) != hand+1 || inExile(plains[0]) || !inExile(plains[1]) {
		t.Errorf("the pick did not move to hand: hand %d → %d", hand, len(me.Hand.Cards))
	}
}

// eve2TryCast tries a strict cast of Soul Reap at the creature and returns the refusal.
func eve2TryCast(t *testing.T, g *game.Game, me *game.Player, oracle string, target uuid.UUID) error {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Soul Reap", TypeLine: "Sorcery", OracleID: oracle, ManaCost: "{1}{B}",
		Colors: []string{"B"}, Owner: me.ID, Controller: me.ID})
	effectstest.AddPool(t, g, me.ID, "{C}{B}")
	return g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(target), Strict: true})
}

func eve2Stats(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	return effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id)
}
