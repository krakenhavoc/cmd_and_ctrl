package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// granted_trigger_doubler_test.go — #4137: a CR 603.2d trigger doubler
// GIVEN to a permanent by a layer-6 grant (ADR 0018's trigger doubling
// addendum, amended 2026-10-19; ADR 0093). Dungeon Delver gives each
// commander creature its owner has Hama Pashar's doubler; The Masamune
// gives the equipped creature a creature-dying doubler.

const (
	dungeonDelverOracle = "0073d731-e60d-433a-ac35-a22af855da20"
	theMasamuneOracle   = "89eb4c39-35ae-405b-8af4-779880cb11c9"
	filigreeFamiliarOrc = "b544f690-e4bf-4a5b-984d-9256518fd574"
)

// gdGoblinLair has `player` venture to the Goblin Lair room (the room
// makes a Goblin token) and returns how many Goblins they control.
func gdGoblinLair(t *testing.T, g *game.Game, player uuid.UUID) int {
	t.Helper()
	before := frTokens(g, player, "Goblin")
	frVenture(t, g, player, LostMineOfPhandelverKey)
	frVenture(t, g, player, "goblin-lair")
	return frTokens(g, player, "Goblin") - before
}

func TestDungeonDelverIsFull(t *testing.T) {
	spec, ok := Lookup(dungeonDelverOracle)
	if !ok || spec.Name != "Dungeon Delver" {
		t.Fatalf("registered=%v name=%q", ok, spec.Name)
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("%v %q, want Full with no caveat", spec.Completeness, spec.Caveats)
	}
}

// The Background alone doubles nothing: the doubler lives on the
// commander creatures it is granted to.
func TestDungeonDelverWithoutACommanderCreatureDoublesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b21Push(g, me.ID, "Dungeon Delver", "Legendary Enchantment — Background", dungeonDelverOracle, 0, 0, "U")
	if n := gdGoblinLair(t, g, me.ID); n != 1 {
		t.Fatalf("Goblin Lair made %d Goblins, want 1", n)
	}
}

func TestDungeonDelverOneCommanderCreatureIsOneExtraTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b21Push(g, me.ID, "Dungeon Delver", "Legendary Enchantment — Background", dungeonDelverOracle, 0, 0, "U")
	b21Commander(g, me.ID, "My Commander", "Legendary Creature — Human Wizard", 3, 3)
	if n := gdGoblinLair(t, g, me.ID); n != 2 {
		t.Fatalf("Goblin Lair made %d Goblins, want 2", n)
	}
}

// Each commander creature carries its own doubler and they add
// (CR 603.2d): two partners are two additional triggers, three Goblins.
func TestDungeonDelverTwoCommanderCreaturesAreTwoExtraTriggers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b21Push(g, me.ID, "Dungeon Delver", "Legendary Enchantment — Background", dungeonDelverOracle, 0, 0, "U")
	b21Commander(g, me.ID, "Partner A", "Legendary Creature — Human Wizard", 3, 3)
	b21Commander(g, me.ID, "Partner B", "Legendary Creature — Elf Rogue", 2, 2)
	if n := gdGoblinLair(t, g, me.ID); n != 3 {
		t.Fatalf("Goblin Lair made %d Goblins, want 3 (one trigger and one extra per commander)", n)
	}
}

// A creature that is not a commander, and a commander an opponent
// owns, do not carry the doubler.
func TestDungeonDelverOnlyCommanderCreaturesYouOwnDouble(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b21Push(g, me.ID, "Dungeon Delver", "Legendary Enchantment — Background", dungeonDelverOracle, 0, 0, "U")
	b21Push(g, me.ID, "Plain Bear", "Creature — Bear", "", 2, 2, "G")
	b21Commander(g, opp.ID, "Their Commander", "Legendary Creature — Human Wizard", 3, 3)
	if n := gdGoblinLair(t, g, me.ID); n != 1 {
		t.Fatalf("Goblin Lair made %d Goblins, want 1", n)
	}
}

// The doubler is Hama Pashar's: an opponent's dungeon is not "a dungeon
// you own", so their rooms stay single even with my commander out.
func TestDungeonDelverDoesNotDoubleAnOpponentsDungeon(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b21Push(g, me.ID, "Dungeon Delver", "Legendary Enchantment — Background", dungeonDelverOracle, 0, 0, "U")
	b21Commander(g, me.ID, "My Commander", "Legendary Creature — Human Wizard", 3, 3)
	if n := gdGoblinLair(t, g, opp.ID); n != 1 {
		t.Fatalf("an opponent's Goblin Lair made %d Goblins, want 1", n)
	}
}

// The doubler is the commander's own ability, so a later removal of all
// its abilities takes it (CR 613.6) while the other commander keeps its.
func TestDungeonDelverDoublerIsLostToALaterAbilityRemoval(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b21Push(g, me.ID, "Dungeon Delver", "Legendary Enchantment — Background", dungeonDelverOracle, 0, 0, "U")
	a := b21Commander(g, me.ID, "Partner A", "Legendary Creature — Human Wizard", 3, 3)
	b21Commander(g, me.ID, "Partner B", "Legendary Creature — Elf Rogue", 2, 2)
	pushAuraOnLand(t, g, me.ID, "Darksteel Mutation", darksteelMutationOracle, a)
	if n := gdGoblinLair(t, g, me.ID); n != 2 {
		t.Fatalf("Goblin Lair made %d Goblins, want 2 (the silenced commander doubles nothing)", n)
	}
}

// A stolen commander is still owned by me, and its controller is the
// thief: "you" in the granted ability is the thief, so it doubles the
// thief's dungeons and not mine (Agent of the Iron Throne's rule).
func TestDungeonDelverStolenCommanderDoublesTheThiefsDungeons(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b21Push(g, me.ID, "Dungeon Delver", "Legendary Enchantment — Background", dungeonDelverOracle, 0, 0, "U")
	commander := b21Commander(g, me.ID, "My Commander", "Legendary Creature — Human Wizard", 3, 3)
	pushAuraOnLand(t, g, opp.ID, "Mind Control", mindControlOracle, commander)
	if n := gdGoblinLair(t, g, me.ID); n != 1 {
		t.Fatalf("my Goblin Lair made %d Goblins, want 1 (the thief holds the doubler)", n)
	}
	if n := gdGoblinLair(t, g, opp.ID); n != 2 {
		t.Fatalf("the thief's Goblin Lair made %d Goblins, want 2", n)
	}
}

// Delver leaving takes the grant, and so the doubler, with it.
func TestDungeonDelverLeavingEndsTheDoubler(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	delver := b21Push(g, me.ID, "Dungeon Delver", "Legendary Enchantment — Background", dungeonDelverOracle, 0, 0, "U")
	b21Commander(g, me.ID, "My Commander", "Legendary Creature — Human Wizard", 3, 3)
	b18Kill(t, g, delver)
	if n := gdGoblinLair(t, g, me.ID); n != 1 {
		t.Fatalf("Goblin Lair made %d Goblins, want 1", n)
	}
}

// --- The Masamune -------------------------------------------------------------

func TestTheMasamuneIsFull(t *testing.T) {
	spec, ok := Lookup(theMasamuneOracle)
	if !ok || spec.Name != "The Masamune" {
		t.Fatalf("registered=%v name=%q", ok, spec.Name)
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("%v %q, want Full with no caveat", spec.Completeness, spec.Caveats)
	}
}

func gdFamiliar(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushDiesCreatureForTest(g, owner, "Filigree Familiar", filigreeFamiliarOrc, "Artifact Creature — Fox", 2, 2)
}

// gdMasamune puts The Masamune on the battlefield attached to host.
func gdMasamune(t *testing.T, g *game.Game, owner, host uuid.UUID) uuid.UUID {
	t.Helper()
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Masamune", TypeLine: "Legendary Artifact — Equipment", OracleID: theMasamuneOracle,
		Owner: owner, Controller: owner,
	})
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(id, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
	})
	return id
}

// The equipped creature's own dies trigger triggers an additional time.
func TestTheMasamuneDoublesTheEquippedCreaturesDiesTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	host := gdFamiliar(g, me.ID)
	gdMasamune(t, g, me.ID, host)
	before := me.Hand.Size()
	b18Kill(t, g, host)
	if got := me.Hand.Size() - before; got != 2 {
		t.Fatalf("an equipped Filigree Familiar drew %d cards when it died, want 2", got)
	}
}

// The same trigger on a creature that is not equipped is not doubled,
// and neither is an equipped creature's once the Masamune is unattached.
func TestTheMasamuneDoesNotDoubleAnyOtherCreaturesTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	equipped := gdFamiliar(g, me.ID)
	plain := gdFamiliar(g, me.ID)
	masamune := gdMasamune(t, g, me.ID, equipped)
	before := me.Hand.Size()
	b18Kill(t, g, plain)
	if got := me.Hand.Size() - before; got != 1 {
		t.Fatalf("an unequipped Filigree Familiar drew %d cards, want 1: the doubler is the host's, not the table's", got)
	}
	g.WithWriteLock(func() {
		if err := g.UnattachForEffect(masamune); err != nil {
			t.Fatalf("UnattachForEffect: %v", err)
		}
	})
	before = me.Hand.Size()
	b18Kill(t, g, equipped)
	if got := me.Hand.Size() - before; got != 1 {
		t.Fatalf("a creature the Masamune let go drew %d cards, want 1", got)
	}
}

// The doubler's question, asked directly: the host's own trigger and an
// emblem its controller owns, caused by a creature dying, and nothing else.
func TestTheMasamuneDoublerNamesThisCreatureAndYourEmblems(t *testing.T) {
	me, opp := uuid.New(), uuid.New()
	host := game.Card{InstanceID: uuid.New(), Controller: me, Owner: me}
	hostLKI := game.Characteristic{Controller: me}
	other := game.Card{InstanceID: uuid.New(), Controller: me, Owner: me}
	myEmblem := game.Card{InstanceID: uuid.New(), OracleID: game.EmblemKeyPrefix + "x", Controller: me, Owner: me}
	theirEmblem := game.Card{InstanceID: uuid.New(), OracleID: game.EmblemKeyPrefix + "x", Controller: opp, Owner: opp}
	dies := game.Event{Kind: game.EventLTB, NewZone: game.ZoneGraveyard}
	cases := []struct {
		name   string
		source game.Card
		ev     game.Event
		spell  bool
		want   bool
	}{
		{"the host's own trigger", host, dies, false, true},
		{"another permanent's trigger", other, dies, false, false},
		{"an emblem you own", myEmblem, dies, false, true},
		{"an opponent's emblem", theirEmblem, dies, false, false},
		{"a cause that is not a death", host, game.Event{Kind: game.EventETB}, false, false},
		{"a spell's trigger", host, dies, true, false},
	}
	d := masamuneDoubler()
	for _, tc := range cases {
		q := game.TriggerDoublingQuery{Event: tc.ev, Doubler: host, DoublerLKI: hostLKI, Source: tc.source, FromSpell: tc.spell, HasSubject: true}
		if got := d.Applies(nil, q); got != tc.want {
			t.Errorf("%s: Applies = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// First strike and "must be blocked if able" apply only while the
// equipped creature is attacking.
func TestTheMasamuneGivesFirstStrikeAndMustBeBlockedWhileAttacking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	gdMasamune(t, g, me.ID, host)
	mustBeBlocked := func() bool {
		var out bool
		g.ReadSnapshot(func() {
			for _, r := range findBattlefieldCardForTest(g, host).Effective().BlockRequirements {
				if r.Kind == game.BlockRequirementMustBeBlocked {
					out = true
				}
			}
		})
		return out
	}
	if effectiveAbilitiesContain(t, g, host, "first strike") || mustBeBlocked() {
		t.Fatal("an equipped creature that is not attacking has first strike or must be blocked")
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(host, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	if !effectiveAbilitiesContain(t, g, host, "first strike") {
		t.Error("the attacking equipped creature has no first strike")
	}
	if !mustBeBlocked() {
		t.Error("the attacking equipped creature does not have to be blocked")
	}
	if err := g.ClearCombat(); err != nil {
		t.Fatalf("ClearCombat: %v", err)
	}
	if effectiveAbilitiesContain(t, g, host, "first strike") || mustBeBlocked() {
		t.Error("the creature kept the attack-only text after combat")
	}
}

// A bundle's registration guard: a granted doubler is the recipient's, so
// an ActiveWhen gate is the grantor's to declare, and a doubler that never
// asks anything would silently do nothing.
func TestCheckGrantsRefusesAMalformedGrantedDoubler(t *testing.T) {
	asks := func(*game.Game, game.TriggerDoublingQuery) bool { return true }
	cases := map[string]AbilityGrant{
		"ActiveWhen": {Key: "test-4137/gated", Text: "x", TriggerDoublers: []game.TriggerDoubler{{Applies: asks, ActiveWhen: ChosenIs("Mardu")}}},
		"no Applies": {Key: "test-4137/blank", Text: "x", TriggerDoublers: []game.TriggerDoubler{{}}},
	}
	for name, gr := range cases {
		func() {
			defer func() {
				r := recover()
				if r == nil {
					t.Errorf("%s: checkGrants accepted the bundle", name)
					return
				}
				if msg, _ := r.(string); !strings.Contains(msg, "trigger doubler") {
					t.Errorf("%s: panic %v does not name the trigger doubler", name, r)
				}
			}()
			checkGrants("Test Card", []AbilityGrant{gr})
		}()
	}
}

// A bundle that holds only a doubler is a real grant, not an empty one,
// and a zone grant (activated abilities only) refuses it.
func TestCheckGrantsAcceptsADoublerOnlyBundleButNotInAZone(t *testing.T) {
	asks := func(*game.Game, game.TriggerDoublingQuery) bool { return true }
	checkGrants("Test Card", []AbilityGrant{{Key: "test-4137/ok", Text: "x", TriggerDoublers: []game.TriggerDoubler{{Applies: asks}}}})
	defer func() {
		if recover() == nil {
			t.Error("a hand-zone bundle with a doubler was accepted")
		}
	}()
	checkGrants("Test Card", []AbilityGrant{{Key: "test-4137/zone", Zone: game.ZoneHand, Text: "x", TriggerDoublers: []game.TriggerDoubler{{Applies: asks}}}})
}
