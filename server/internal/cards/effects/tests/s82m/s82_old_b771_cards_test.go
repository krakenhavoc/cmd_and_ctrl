package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b771_cards_test.go — Weatherlight damage cards, S82 OLD batch 771
// (#5629), played through the real catalog entries.

const (
	o771Firestorm   = "157813c8-af83-41b5-9ffb-4abd167c1bfc"
	o771CinderGiant = "3a8306ae-3162-4105-95f4-dacd10c7c051"
	o771Vines       = "3b6b565e-c0b5-48f9-b3e0-4d5d4b065402"
	o771Sawtooth    = "473c9f27-2091-4d87-b279-64b3662c9af8"
	o771SpinDark    = "69eeee95-cc69-41d1-833b-8895ef72df82"
	o771DingusStaff = "7127fc35-600c-4397-8c01-5327d3c4944b"
	o771LavaStorm   = "78d33345-a147-402f-9a40-1e8260d7d44e"
	o771AetherFlash = "a3c35742-e306-49b6-b042-db4f685c6f86"
	o771Downdraft   = "c2508190-5aca-4511-adc8-2072cd228119"
	o771Stalker     = "e5ef45e4-8c19-40bf-8756-c1b408dcb2fb"
)

func TestS82OldB771CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		o771Firestorm: "Firestorm", o771CinderGiant: "Cinder Giant", o771Vines: "Choking Vines",
		o771Sawtooth: "Sawtooth Ogre", o771SpinDark: "Spinning Darkness", o771DingusStaff: "Dingus Staff",
		o771LavaStorm: "Lava Storm", o771AetherFlash: "Aether Flash", o771Downdraft: "Downdraft",
		o771Stalker: "Urborg Stalker",
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

func TestS82OldB771FirestormDiscardsXAndHitsXTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	me.Hand.Cards = nil
	spell := effectstest.AvB24Hand(g, me, game.Card{Name: "Firestorm", TypeLine: "Instant", ManaCost: "{R}", OracleID: o771Firestorm})
	d1 := effectstest.HandCard(me, "Spare One", "Instant")
	d2 := effectstest.HandCard(me, "Spare Two", "Instant")
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	ogre := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	effectstest.FloatForTest(g, me, "R")
	life := opp.Life
	refs := []game.TargetRef{{Kind: game.TargetCard, ID: bear}, effectstest.PlayerTarget(opp.ID)}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, XValue: 2, DiscardIDs: []uuid.UUID{d1}, Targets: refs}); err == nil {
		t.Fatal("the cast was accepted with one discard at X=2")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, XValue: 2, DiscardIDs: []uuid.UUID{d1, d2}, Targets: refs[:1]}); err == nil {
		t.Fatal("the cast was accepted with one target at X=2")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, XValue: 2, DiscardIDs: []uuid.UUID{d1, d2}, Targets: refs}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(d1) || !me.Graveyard.Contains(d2) {
		t.Error("the cost did not discard two cards")
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the 2/2 target survived 2 damage")
	}
	if opp.Life != life-2 {
		t.Errorf("the player target is at %d, want %d", opp.Life, life-2)
	}
	if got := effectstest.DamageMarkedOn(g, ogre); got != 0 {
		t.Errorf("an untargeted creature has %d damage", got)
	}
}

func TestS82OldB771CinderGiantHurtsEachOtherCreatureYouControlAtUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giant := effectstest.B12Push(g, me.ID, "Cinder Giant", "Creature — Giant", o771CinderGiant, 5, 3)
	small := effectstest.B12Creature(g, me.ID, "My Rat", "Creature — Rat", 1, 1)
	big := effectstest.B12Creature(g, me.ID, "My Wall", "Creature — Wall", 0, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the 1/1 survived 2 damage")
	}
	if got := effectstest.DamageMarkedOn(g, big); got != 2 {
		t.Errorf("my 0/4 has %d damage, want 2", got)
	}
	if got := effectstest.DamageMarkedOn(g, giant); got != 0 {
		t.Errorf("the Giant hurt itself: %d damage", got)
	}
	if got := effectstest.DamageMarkedOn(g, theirs); got != 0 {
		t.Errorf("an opponent's creature took %d damage", got)
	}
}

func TestS82OldB771ChokingVinesBlocksXAttackersAndDamagesThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	attacker, me := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, attacker.ID, "Attacker A", "Creature — Bear", 2, 3)
	b := effectstest.B12Creature(g, attacker.ID, "Attacker B", "Creature — Bear", 2, 3)
	c := effectstest.B12Creature(g, attacker.ID, "Attacker C", "Creature — Bear", 2, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	effectstest.DeclareAttack(t, g, me.ID, a, b, c)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if _, err := o771CastVines(t, g, me, 2, effectstest.CardRefs(a)); err == nil {
		t.Fatal("one target at X=2 was accepted")
	}
	if _, err := o771CastVines(t, g, me, 2, effectstest.CardRefs(a, c)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, a); got != 1 {
		t.Errorf("target A has %d damage, want 1", got)
	}
	if g.Battlefield.Contains(c) {
		t.Error("the 2/1 target survived 1 damage")
	}
	if got := effectstest.DamageMarkedOn(g, b); got != 0 {
		t.Errorf("the untargeted attacker has %d damage", got)
	}
	if !g.BlockedAttackerForEffect(a) {
		t.Error("the targeted attacker did not become blocked")
	}
}

func TestS82OldB771SawtoothOgreDamagesWhateverBlocksOrIsBlockedByItAtEndOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := effectstest.B12Push(g, me.ID, "Sawtooth Ogre", "Creature — Ogre", o771Sawtooth, 3, 3)
	wall := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 5)
	effectstest.DeclareAttack(t, g, opp.ID, ogre)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, ogre); err != nil {
		t.Fatalf("block: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	// 3 combat damage was cleared at end of turn only; the wall also took the
	// delayed 1, so it is 4 total this turn.
	if got := effectstest.DamageMarkedOn(g, wall); got != 4 {
		t.Errorf("the blocker has %d damage, want 4 (3 combat and 1 from the trigger)", got)
	}
}

func TestS82OldB771SpinningDarknessExilesTheTopThreeBlackCardsOrPays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	me.Hand.Cards = nil
	me.Graveyard.Cards = nil
	var black []uuid.UUID
	for i, cost := range []string{"{B}", "{U}", "{B}", "{B}", "{B}"} {
		id := uuid.New()
		colors := []string{"B"}
		if cost == "{U}" {
			colors = []string{"U"}
		}
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Card " + string(rune('A'+i)), TypeLine: "Sorcery", ManaCost: cost, Colors: colors, Owner: me.ID, Controller: me.ID})
		if cost == "{B}" {
			black = append(black, id)
		}
	}
	spell := effectstest.AvB24Hand(g, me, game.Card{Name: "Spinning Darkness", TypeLine: "Instant", ManaCost: "{4}{B}{B}", OracleID: o771SpinDark})
	effectstest.AdvanceToMain(t, g)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 3)
	black2 := effectstest.B12Push(g, opp.ID, "Their Zombie", "Creature — Zombie", "", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == black2 {
				g.Battlefield.Cards[i].Colors = []string{"B"}
			}
		}
	})
	// A black creature is not a legal target.
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(black2), AlternativeCost: "exile-top-three-black", AltCostIDs: black[1:]}); err == nil {
		t.Fatal("a black creature was accepted as the target")
	}
	// The cards below the top three are refused.
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(victim), AlternativeCost: "exile-top-three-black", AltCostIDs: black[:3]}); err == nil {
		t.Fatal("the three lowest black cards paid the cost")
	}
	life := me.Life
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(victim), AlternativeCost: "exile-top-three-black", AltCostIDs: black[1:]}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	for _, id := range black[1:] {
		if me.Graveyard.Contains(id) {
			t.Errorf("black card %s was not exiled", id)
		}
	}
	if !me.Graveyard.Contains(black[0]) {
		t.Error("a card below the top three black ones was exiled")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the 2/3 target survived 3 damage")
	}
	if me.Life != life+3 {
		t.Errorf("life is %d, want %d", me.Life, life+3)
	}
}

func TestS82OldB771DingusStaffHurtsTheControllerOfEachCreatureThatDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Dingus Staff", "Artifact", o771DingusStaff, false)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	myLife, theirLife := me.Life, opp.Life
	effectstest.B14Kill(g, theirs)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirLife-2 || me.Life != myLife {
		t.Errorf("after their creature died: they are at %d (want %d), I am at %d (want %d)", opp.Life, theirLife-2, me.Life, myLife)
	}
	effectstest.B14Kill(g, mine)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife-2 {
		t.Errorf("after my creature died I am at %d, want %d", me.Life, myLife-2)
	}
}

func TestS82OldB771LavaStormHitsAttackersOrBlockers(t *testing.T) {
	for _, tc := range []struct {
		name         string
		mode         int
		attackerHurt int
		blockerHurt  int
	}{{"attackers", 0, 2, 0}, {"blockers", 1, 0, 2}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Ogre", 0, 5)
		blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Wall", 0, 5)
		idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Wall", 0, 5)
		effectstest.DeclareAttack(t, g, opp.ID, attacker)
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		if err := g.DeclareBlocker(blocker, attacker); err != nil {
			t.Fatalf("block: %v", err)
		}
		effectstest.B14CastAt(t, g, me, "Lava Storm", "Instant", o771LavaStorm, nil, game.CastSpellParams{Strict: true, Modes: []int{tc.mode}})
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.DamageMarkedOn(g, attacker); got != tc.attackerHurt {
			t.Errorf("%s: the attacker has %d damage, want %d", tc.name, got, tc.attackerHurt)
		}
		if got := effectstest.DamageMarkedOn(g, blocker); got != tc.blockerHurt {
			t.Errorf("%s: the blocker has %d damage, want %d", tc.name, got, tc.blockerHurt)
		}
		if got := effectstest.DamageMarkedOn(g, idle); got != 0 {
			t.Errorf("%s: an idle creature has %d damage", tc.name, got)
		}
	}
}

func TestS82OldB771AetherFlashDamagesEachCreatureThatEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Aether Flash", "Enchantment", o771AetherFlash, false)
	big := effectstest.CastHoldCreature(t, g, "Big Ogre", "Creature — Ogre", "", 4, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, big); got != 2 {
		t.Errorf("the entering 4/4 has %d damage, want 2", got)
	}
	small := effectstest.CastHoldCreature(t, g, "Small Rat", "Creature — Rat", "", 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the entering 1/1 survived 2 damage")
	}
}

func TestS82OldB771DowndraftRemovesFlyingOrSacrificesToHitFliers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	draft := effectstest.PushCatalogPermanent(g, me.ID, "Downdraft", "Enchantment", o771Downdraft, false)
	bird := effectstest.B12Creature(g, opp.ID, "Their Bird", "Creature — Bird", 1, 3)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bird {
				g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, "flying")
			}
		}
	})
	ground := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	if !effectstest.HasEffectiveKeyword(t, g, bird, "flying") {
		t.Fatal("the bird does not fly")
	}
	effectstest.AddPool(t, g, me.ID, "{G}")
	if err := g.ActivateCatalogAbility(me.ID, draft, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bird)}); err != nil {
		t.Fatalf("activate the first ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasEffectiveKeyword(t, g, bird, "flying") {
		t.Error("the bird kept flying")
	}
	// Flying is gone, so the sacrifice ability finds no fliers until it returns.
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.AdvanceToMainOf(t, g, 0)
	if !effectstest.HasEffectiveKeyword(t, g, bird, "flying") {
		t.Fatal("flying did not come back at end of turn")
	}
	if err := g.ActivateCatalogAbility(me.ID, draft, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the second ability: %v", err)
	}
	if g.Battlefield.Contains(draft) {
		t.Error("the Downdraft was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, bird); got != 2 {
		t.Errorf("the flier has %d damage, want 2", got)
	}
	if got := effectstest.DamageMarkedOn(g, ground); got != 0 {
		t.Errorf("the ground creature has %d damage", got)
	}
}

func TestS82OldB771UrborgStalkerHurtsAPlayerWithANonblackNonlandPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	stalker := effectstest.B12Push(g, me.ID, "Urborg Stalker", "Creature — Horror", o771Stalker, 2, 4)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == stalker {
				g.Battlefield.Cards[i].ManaCost = "{3}{B}"
			}
		}
	})
	// The opponent controls a colorless artifact (nonblack) and a land; the third
	// player controls only a land.
	effectstest.PushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	effectstest.PushCatalogPermanent(g, opp.ID, "Their Swamp", "Basic Land — Swamp", "", false)
	effectstest.PushCatalogPermanent(g, third.ID, "Their Island", "Basic Land — Island", "", false)
	mine, theirs, thirds := me.Life, opp.Life, third.Life
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirs-1 {
		t.Errorf("the opponent is at %d, want %d (a colorless artifact is nonblack)", opp.Life, theirs-1)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if third.Life != thirds {
		t.Errorf("a player with only a land took damage: %d, want %d", third.Life, thirds)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mine {
		t.Errorf("a player controlling only a black permanent took damage: %d, want %d", me.Life, mine)
	}
}

func TestS82OldB771ChokingVinesIsCastableOnlyDuringTheDeclareBlockersStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B12Creature(g, me.ID, "Attacker A", "Creature — Bear", 2, 3)
	effectstest.AdvanceToMain(t, g)
	if _, err := o771CastVines(t, g, me, 1, effectstest.CardRefs(a)); err == nil {
		t.Error("cast in the main phase was accepted")
	}
}

// o771CastVines puts Choking Vines in the player's hand, floats the mana and
// casts it for X.
func o771CastVines(t *testing.T, g *game.Game, p *game.Player, x int, targets []game.TargetRef) (uuid.UUID, error) {
	t.Helper()
	id := effectstest.AvB24Hand(g, p, game.Card{Name: "Choking Vines", TypeLine: "Instant", ManaCost: "{X}{G}", OracleID: o771Vines})
	effectstest.FloatForTest(g, p, "GCC")
	err := g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true, XValue: x, Targets: targets})
	return id, err
}
