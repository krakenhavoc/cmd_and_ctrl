package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b561_cards_test.go — Coldsnap's activated-ability creatures, played
// from the real catalog entries (S82 OLD batch 561).

const (
	b561FrostwakerOracle = "2cff1f9e-11a6-460c-8f5f-4c02424357e3"
	b561DrifterOracle    = "63ec5617-0da6-4534-b502-34304b1ea02b"
	b561RimerunnerOracle = "8505f9a6-a510-43f9-a96d-b840b353e144"
	b561MusherOracle     = "89f87341-c18f-4dc7-ba2e-81305383231e"
	b561DeadOracle       = "cfc58887-859d-4ba5-8507-1fb1aa4d37c0"
	b561WindformOracle   = "dd66b8ab-14a9-455b-98f6-5356df72a522"
	b561YetiOracle       = "e606cba7-7e3c-4834-b2cd-56e51cedccb7"
	b561TaskmageOracle   = "f5d9248e-9cff-4f80-b6ed-7d368099e97b"
	b561WarlordOracle    = "f73d1525-ab0b-4772-9dfa-66e1c84c3ea4"
)

func TestS82OldB561CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b561FrostwakerOracle: "Balduvian Frostwaker",
		b561DrifterOracle:    "Squall Drifter",
		b561RimerunnerOracle: "Goblin Rimerunner",
		b561MusherOracle:     "Zombie Musher",
		b561DeadOracle:       "Rimebound Dead",
		b561WindformOracle:   "Adarkar Windform",
		b561YetiOracle:       "Ohran Yeti",
		b561TaskmageOracle:   "Rimewind Taskmage",
		b561WarlordOracle:    "Balduvian Warlord",
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

func b561Snow(g *game.Game, owner uuid.UUID, subtype string) uuid.UUID {
	return effectstest.BlbB1Push(g, owner, "Snow-Covered "+subtype, "Basic Snow Land — "+subtype, "", 0, 0)
}

func b561Plain(g *game.Game, owner uuid.UUID, subtype string) uuid.UUID {
	return effectstest.BlbB1Push(g, owner, subtype, "Basic Land — "+subtype, "", 0, 0)
}

func b561Target(id uuid.UUID) game.ActivateAbilityParams {
	return game.ActivateAbilityParams{Strict: true, AutoTap: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: id}}}
}

func b561Tapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c := effectstest.FindBattlefieldCardByID(g, id)
	if c == nil {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Tapped
}

func TestBalduvianFrostwakerAnimatesASnowLandForGood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.BlbB1Push(g, me.ID, "Balduvian Frostwaker", "Creature — Human Wizard", b561FrostwakerOracle, 1, 1)
	snow := b561Snow(g, me.ID, "Island")
	plain := b561Plain(g, me.ID, "Island")
	effectstest.AdvanceToMain(t, g)
	// A nonsnow land is not a legal target.
	if err := g.ActivateCatalogAbility(me.ID, id, 0, b561Target(plain)); err == nil {
		t.Fatal("targeted a nonsnow land: accepted")
	}
	// The cost is {U}: the snow Island pays it, or the plain one does.
	if err := g.ActivateCatalogAbility(me.ID, id, 0, b561Target(snow)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, snow); p != 2 || tough != 2 {
		t.Fatalf("the land is %d/%d, want 2/2", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, snow, "flying") {
		t.Error("the animated land lacks flying")
	}
	c := effectstest.FindBattlefieldCardByID(g, snow)
	if !c.IsCreature() || !c.IsLand() || !c.HasColor("U") {
		t.Errorf("the land is creature=%v land=%v blue=%v", c.IsCreature(), c.IsLand(), c.HasColor("U"))
	}
	// No duration: it is still a creature in later turns.
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if c := effectstest.FindBattlefieldCardByID(g, snow); !c.IsCreature() {
		t.Error("the land stopped being a creature")
	}
}

func TestSquallDrifterTapsTargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.BlbB1Push(g, me.ID, "Squall Drifter", "Snow Creature — Elemental", b561DrifterOracle, 1, 1)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, id, "flying") {
		t.Error("Squall Drifter lacks flying")
	}
	b561Plain(g, me.ID, "Plains")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, b561Target(victim)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !b561Tapped(t, g, victim) || !b561Tapped(t, g, id) {
		t.Errorf("victim tapped %v, drifter tapped %v, want both", b561Tapped(t, g, victim), b561Tapped(t, g, id))
	}
}

func TestGoblinRimerunnerStopsABlockerAndGainsHasteForSnow(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.BlbB1Push(g, me.ID, "Goblin Rimerunner", "Snow Creature — Goblin Warrior", b561RimerunnerOracle, 2, 2)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	// The {S} ability is refused with nonsnow mana.
	b561Plain(g, me.ID, "Mountain")
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err == nil {
		t.Fatal("{S} paid from a plain Mountain: accepted")
	}
	b561Snow(g, me.ID, "Mountain")
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("{S} off a snow land: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, id, "haste") {
		t.Error("no haste after the {S} ability")
	}
	// With haste it can pay its own {T}.
	if err := g.ActivateCatalogAbility(me.ID, id, 0, b561Target(victim)); err != nil {
		t.Fatalf("{T}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.RestrictionsOf(t, g, victim)&game.CantBlock == 0 {
		t.Error("the target can still block")
	}
}

func TestRegeneratingSnowCreaturesPayWithSnowOnly(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Zombie Musher", "Snow Creature — Zombie", b561MusherOracle},
		{"Rimebound Dead", "Snow Creature — Skeleton", b561DeadOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			id := effectstest.BlbB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			effectstest.AdvanceToMain(t, g)
			plain := b561Plain(g, me.ID, "Swamp")
			if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err == nil {
				t.Fatal("{S} paid from a plain Swamp: accepted")
			}
			if b561Tapped(t, g, plain) {
				t.Fatal("a refused activation tapped the Swamp")
			}
			b561Snow(g, me.ID, "Swamp")
			if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err != nil {
				t.Fatalf("{S} off a snow land: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			// The shield is up: destroying the creature regenerates it.
			g.WithWriteLock(func() {
				if err := g.DestroyPermanentForEffect(id); err != nil {
					t.Fatalf("destroy: %v", err)
				}
			})
			c := effectstest.FindBattlefieldCardByID(g, id)
			if c == nil || !c.Tapped {
				t.Fatal("the creature did not regenerate (gone or untapped)")
			}
		})
	}
	spec, _ := effects.Lookup(b561MusherOracle)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "snow landwalk" {
		t.Errorf("Zombie Musher keywords %v, want snow landwalk", spec.PrintedKeywords)
	}
}

func TestAdarkarWindformRemovesFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.BlbB1Push(g, me.ID, "Adarkar Windform", "Snow Creature — Illusion", b561WindformOracle, 3, 3)
	bird := effectstest.PushSizedCreature(g, opp.ID, "Bird", 1, 1, "flying")
	effectstest.AdvanceToMain(t, g)
	b561Plain(g, me.ID, "Island")
	b561Plain(g, me.ID, "Island")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, b561Target(bird)); err == nil {
		t.Fatal("{1}{S} paid with two plain Islands: accepted")
	}
	b561Snow(g, me.ID, "Island")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, b561Target(bird)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasKeywordOnBattlefield(t, g, bird, "flying") {
		t.Error("the Bird still has flying")
	}
}

func TestOhranYetiGivesFirstStrikeToSnowCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.BlbB1Push(g, me.ID, "Ohran Yeti", "Snow Creature — Yeti", b561YetiOracle, 3, 3)
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	b561Snow(g, me.ID, "Mountain")
	b561Plain(g, me.ID, "Mountain")
	b561Plain(g, me.ID, "Mountain")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, b561Target(bear)); err == nil {
		t.Fatal("targeted a nonsnow creature: accepted")
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, b561Target(id)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, id, "first strike") {
		t.Error("the Yeti has no first strike")
	}
}

func TestRimewindTaskmageNeedsFourSnowPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.BlbB1Push(g, me.ID, "Rimewind Taskmage", "Creature — Human Wizard", b561TaskmageOracle, 1, 2)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 3; i++ {
		b561Snow(g, me.ID, "Island")
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, b561Target(victim)); err == nil {
		t.Fatal("activated with three snow permanents: accepted")
	}
	b561Snow(g, me.ID, "Island")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, b561Target(victim)); err != nil {
		t.Fatalf("activate with four snow permanents: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !b561Tapped(t, g, victim) {
		t.Error("the Bear was not tapped")
	}
}

func TestBalduvianWarlordReassignsABlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	warlord := effectstest.BlbB1Push(g, me.ID, "Balduvian Warlord", "Creature — Human Barbarian", b561WarlordOracle, 3, 2)
	first := effectstest.PushVanillaCreature(g, me.ID, "First Bear", 3, 3)
	second := effectstest.PushVanillaCreature(g, me.ID, "Second Bear", 2, 2)
	wall := effectstest.PushVanillaCreature(g, opp.ID, "Wall", 0, 5)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, a := range []uuid.UUID{first, second} {
		if err := g.DeclareAttacker(a, opp.ID); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, first); err != nil {
		t.Fatal(err)
	}
	effectstest.FinishBlockDeclarations(t, g)
	if err := g.ActivateCatalogAbility(me.ID, warlord, 0, b561Target(first)); err == nil {
		t.Fatal("targeted an attacker that is not blocking: accepted")
	}
	if err := g.ActivateCatalogAbility(me.ID, warlord, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: wall}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.AnswerPermanentPicks(t, g, me.ID, map[uuid.UUID][]uuid.UUID{me.ID: {second}}); got == 0 {
		t.Fatal("no attacker prompt")
	}
	var blockedFirst, blockedSecond bool
	g.WithWriteLock(func() {
		blockedFirst = g.BlockedAttackerForEffect(first)
		blockedSecond = g.BlockedAttackerForEffect(second)
	})
	if blockedFirst || !blockedSecond {
		t.Errorf("first blocked %v, second blocked %v; want the first unblocked and the second blocked", blockedFirst, blockedSecond)
	}
}
