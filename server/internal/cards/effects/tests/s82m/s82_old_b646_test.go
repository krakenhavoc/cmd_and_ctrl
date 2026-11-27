package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b646_test.go — Invasion, "Activated abilities" (S82 OLD batch 646, #5616): each
// card driven through a real activation, cast or trigger.

func oldB646AllFull(t *testing.T, ids ...string) {
	t.Helper()
	for _, id := range ids {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB646Ready(g *game.Game, id uuid.UUID) {
	effectstest.UntapForTest(g, id)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
}

// oldB646Perm seeds a ready catalog permanent for the active seat.
func oldB646Perm(t *testing.T, g *game.Game, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	id := effectstest.SeedPermanentWithOracle(g, g.Seats[0].ID, name, typeLine, oracle)
	oldB646Ready(g, id)
	return id
}

func oldB646Activate(t *testing.T, g *game.Game, src uuid.UUID, ability int, mana string, params game.ActivateAbilityParams) error {
	t.Helper()
	me := g.Seats[0]
	if mana != "" {
		effectstest.FloatMana(t, g, me, mana)
	}
	params.Strict = true
	return g.ActivateCatalogAbility(me.ID, src, ability, params)
}

func oldB646Effective(g *game.Game, id uuid.UUID) game.Card {
	c, _ := g.LookupCardForEffect(id)
	return c
}

func oldB646Target(id uuid.UUID) game.ActivateAbilityParams {
	return game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: id}}}
}

const (
	oldB646Transport = "6ab1bd6f-866b-4665-bd8b-627aafafb94d"
	oldB646Utopia    = "73a9122c-60d2-4fdc-bc7b-3576bcbca241"
	oldB646Grappler  = "74c5bfb0-1375-46e9-9a6c-1b9fada2913b"
	oldB646Crow      = "79bf98c8-1169-477d-838e-2ebc0fa396dc"
	oldB646Chameleon = "82c583ab-8158-4074-a3ff-b7a05ecc8ab1"
	oldB646Strossus  = "8f2ab0a8-9d60-493b-a284-9ed7fa07e5ea"
	oldB646Visionary = "8fc72080-8d4d-427d-ae30-065bc8ee9a28"
	oldB646Angel     = "9629ab26-f231-481d-8392-e2450ca02b67"
	oldB646Acolyte   = "969626f0-19d9-47ae-8766-4d59269e389c"
	oldB646Weaver    = "9777a54b-1c60-4d6b-bb83-c57cb3a1d7e6"
	oldB646Hunting   = "a1b7fbfc-c510-4b1c-a2b9-c1204c699393"
)

func TestS82OldB646AllAreFullWithNoCaveats(t *testing.T) {
	oldB646AllFull(t, oldB646Transport, oldB646Utopia, oldB646Grappler, oldB646Crow, oldB646Chameleon, oldB646Strossus,
		oldB646Visionary, oldB646Angel, oldB646Acolyte, oldB646Weaver, oldB646Hunting)
}

func oldB646Colored(g *game.Game, owner uuid.UUID, name, color string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear", Colors: []string{color},
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
}

func TestS82OldB646UtopiaTreeTapsForAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tree := oldB646Perm(t, g, "Utopia Tree", "Creature — Plant", oldB646Utopia)
	effectstest.TapForMana(t, g, me.ID, tree)
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil {
		t.Fatal("no colour pick was asked for")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "G"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if me.ManaPool.Total() != 1 {
		t.Errorf("the pool holds %d mana, want 1", me.ManaPool.Total())
	}
	if !effectstest.TapCostTapped(g, tree) {
		t.Error("the {T} did not tap the Tree")
	}
}

func TestS82OldB646SelfGrantRows(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle, mana, keyword string }{
		{"Viashino Grappler", "Creature — Lizard", oldB646Grappler, "{G}", "trample"},
		{"Glimmering Angel", "Creature — Angel", oldB646Angel, "{U}", "shroud"},
	} {
		g := effectstest.NewCatalogGame(t)
		id := oldB646Perm(t, g, tc.name, tc.typeLine, tc.oracle)
		if effectstest.P1g2HasKeyword(g, id, tc.keyword) {
			t.Fatalf("%s starts with %s", tc.name, tc.keyword)
		}
		if err := oldB646Activate(t, g, id, 0, "", game.ActivateAbilityParams{}); err == nil {
			t.Errorf("%s: activated without mana", tc.name)
		}
		if err := oldB646Activate(t, g, id, 0, tc.mana, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("%s: activate: %v", tc.name, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.P1g2HasKeyword(g, id, tc.keyword) {
			t.Errorf("%s did not gain %s", tc.name, tc.keyword)
		}
	}
}

func TestS82OldB646MetathranTransportBlueAndCantBeBlockedByBlue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tr := oldB646Perm(t, g, "Metathran Transport", "Creature — Metathran", oldB646Transport)
	if len(oldB646Effective(g, tr).Colors) == 0 {
		// The seeded card has no printed colour: give it blue so the block rule has a subject.
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == tr {
					g.Battlefield.Cards[i].Colors = []string{"U"}
				}
			}
		})
	}
	red := oldB646Colored(g, me.ID, "Imp", "R")
	if err := oldB646Activate(t, g, tr, 0, "", oldB646Target(red)); err == nil {
		t.Fatal("activated without paying {U}")
	}
	if err := oldB646Activate(t, g, tr, 0, "{U}", oldB646Target(red)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := oldB646Effective(g, red)
	if !c.HasColor("U") || c.HasColor("R") {
		t.Errorf("colours %v, want only blue", c.Colors)
	}
	if !effectstest.P1g2HasKeyword(g, tr, "flying") {
		t.Error("no flying")
	}
	spec, _ := effects.Lookup(oldB646Transport)
	if len(spec.BlockRules) != 1 {
		t.Errorf("block rules %d, want the blue-creature restriction", len(spec.BlockRules))
	}
}

func TestS82OldB646RainbowCrowAndChameleonChooseTheirColor(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle, mana string }{
		{"Rainbow Crow", "Creature — Bird", oldB646Crow, "{C}"},
		{"Kavu Chameleon", "Creature — Kavu", oldB646Chameleon, "{G}"},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		id := oldB646Perm(t, g, tc.name, tc.typeLine, tc.oracle)
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == id {
					g.Battlefield.Cards[i].Colors = []string{"G"}
				}
			}
		})
		if err := oldB646Activate(t, g, id, 0, "", game.ActivateAbilityParams{}); err == nil {
			t.Errorf("%s: activated without mana", tc.name)
		}
		if err := oldB646Activate(t, g, id, 0, tc.mana, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("%s: activate: %v", tc.name, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerColor(t, g, me.ID, "R")
		c := oldB646Effective(g, id)
		if !c.HasColor("R") || c.HasColor("G") {
			t.Errorf("%s: colours %v, want only red", tc.name, c.Colors)
		}
	}
}

func TestS82OldB646KavuChameleonCantBeCountered(t *testing.T) {
	spec, ok := effects.Lookup(oldB646Chameleon)
	if !ok || !spec.CantBeCountered {
		t.Fatal("Kavu Chameleon is not registered as uncounterable")
	}
}

func TestS82OldB646TidalVisionaryRecolorsATarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	vis := oldB646Perm(t, g, "Tidal Visionary", "Creature — Merfolk Wizard", oldB646Visionary)
	ogre := oldB646Colored(g, opp.ID, "Ogre", "R")
	_ = me
	if err := oldB646Activate(t, g, vis, 0, "", oldB646Target(ogre)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "W")
	c := oldB646Effective(g, ogre)
	if !c.HasColor("W") || c.HasColor("R") {
		t.Errorf("colours %v, want only white", c.Colors)
	}
	if !effectstest.TapCostTapped(g, vis) {
		t.Error("the {T} did not tap the Visionary")
	}
}

func TestS82OldB646CrimsonAcolyteGivesProtectionFromRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ac := oldB646Perm(t, g, "Crimson Acolyte", "Creature — Human Cleric", oldB646Acolyte)
	bear := oldB646Colored(g, me.ID, "Bear", "G")
	if err := oldB646Activate(t, g, ac, 0, "", oldB646Target(bear)); err == nil {
		t.Fatal("activated without paying {W}")
	}
	if err := oldB646Activate(t, g, ac, 0, "{W}", oldB646Target(bear)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.P1g2HasKeyword(g, bear, game.ProtectionFromColor("R")) {
		t.Error("no protection from red")
	}
	if !effectstest.P1g2HasKeyword(g, ac, game.ProtectionFromColor("R")) {
		t.Error("the Acolyte lacks its printed protection from red")
	}
}

func TestS82OldB646RageWeaverHastesBlackOrGreenOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rw := oldB646Perm(t, g, "Rage Weaver", "Creature — Human Wizard", oldB646Weaver)
	red := oldB646Colored(g, me.ID, "Imp", "R")
	black := oldB646Colored(g, me.ID, "Zombie", "B")
	if err := oldB646Activate(t, g, rw, 0, "{C}{C}", oldB646Target(red)); err == nil {
		t.Fatal("a red creature was a legal target")
	}
	if err := oldB646Activate(t, g, rw, 0, "", oldB646Target(black)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.P1g2HasKeyword(g, black, "haste") {
		t.Error("the black creature did not gain haste")
	}
}

func TestS82OldB646DevouringStrossusSacrificesAtUpkeepAndRegenerates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	str := effectstest.B12Push(g, me.ID, "Devouring Strossus", "Creature — Phyrexian Horror", oldB646Strossus, 9, 9)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	oldB646Ready(g, str)
	// Sacrifice a creature: Regenerate.
	if err := oldB646Activate(t, g, str, 0, "", game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("regenerate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature stayed")
	}
	if !(g.RegenerationShieldsOn(str) > 0) {
		t.Error("the Strossus has no regeneration shield")
	}
	// Upkeep: sacrifice a creature.
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	str2 := effectstest.B12Push(g2, me2.ID, "Devouring Strossus", "Creature — Phyrexian Horror", oldB646Strossus, 9, 9)
	chaff := effectstest.B12Creature(g2, me2.ID, "Chaff", "Creature — Bear", 1, 1)
	effectstest.AdvanceToUpkeepOfSeat(t, g2, 0)
	effectstest.PassPriorityAroundTable(t, g2)
	effectstest.AnswerSacrifice(t, g2, me2.ID, chaff)
	if g2.Battlefield.Contains(chaff) {
		t.Error("the Bear was not sacrificed at upkeep")
	}
	if !g2.Battlefield.Contains(str2) {
		t.Error("the Strossus was sacrificed although another creature was chosen")
	}
}

func TestS82OldB646HuntingKavuExilesItselfAndAnAttackerWithoutFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	kavu := oldB646Perm(t, g, "Hunting Kavu", "Creature — Kavu", oldB646Hunting)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	bird := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird", Keywords: []string{"flying"},
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})
	idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Ogre", 3, 3)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.InstanceID == ogre || c.InstanceID == bird {
				c.AttackingTarget = me.ID
			}
		}
	})
	if err := oldB646Activate(t, g, kavu, 0, "{C}{R}{G}", oldB646Target(bird)); err == nil {
		t.Fatal("an attacker with flying was a legal target")
	}
	if err := oldB646Activate(t, g, kavu, 0, "", oldB646Target(idle)); err == nil {
		t.Fatal("a creature that is not attacking was a legal target")
	}
	if err := oldB646Activate(t, g, kavu, 0, "", oldB646Target(ogre)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) || g.Battlefield.Contains(kavu) {
		t.Error("the Kavu and the attacker were not both exiled")
	}
	if !g.Exile.Contains(ogre) || !g.Exile.Contains(kavu) {
		t.Error("they are not in exile")
	}
}
