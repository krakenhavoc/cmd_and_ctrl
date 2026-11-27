package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b655_cards_test.go — Invasion, batch 655 (S82, #5616). Each test
// plays the real catalog entry through the engine.

const (
	oldB655SimoonOracle      = "128e201a-f520-4917-a1a4-2f3836f1f92d"
	oldB655DromarOracle      = "35ad7f82-a1b1-4aa7-a8b0-3f99ca2cff52"
	oldB655ArcherOracle      = "5322a218-6a0c-4d86-835a-c87eb52aaf6a"
	oldB655DarigaazOracle    = "6cbafa7b-4448-4ad4-b4f2-0ab6e7c95b0b"
	oldB655TrevaOracle       = "72524f3c-674b-4ebf-9526-18737b354010"
	oldB655GiantOracle       = "7a2336ff-c8ed-4daa-b182-403c21bc8133"
	oldB655DartOracle        = "7bc53740-86e7-4e9d-aab6-b5ae62a42133"
	oldB655OverabundOracle   = "b1ee0503-be92-4a06-89cd-5dfa937930d3"
	oldB655GhituFireOracle   = "c15ae9d8-b423-4628-ad88-3f44515e0c77"
	oldB655WeaverOracle      = "c2cfbc1a-ae73-409c-b73a-5c8625365905"
	oldB655NightscapeOracle  = "e5f9bd6b-0aa6-4f61-b101-ec194af6d632"
	oldB655ThornscapeOracle  = "ea802eda-fc8d-45c7-b5a3-238a2cdc4dcb"
	oldB655RaysOracle        = "fd6730e2-3904-43cf-90c6-e3d9fe210bc5"
	oldB655PresenceOracle    = "ff9a5529-83bd-46a8-b20b-235a5fd26d93"
	oldB655CreatureTypeLine  = "Creature — Bear"
	oldB655LegendaryDragon   = "Legendary Creature — Dragon"
	oldB655CreatureWizardTyp = "Creature — Human Wizard"
)

func TestOldB655EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		oldB655SimoonOracle: "Simoon", oldB655DromarOracle: "Dromar, the Banisher", oldB655ArcherOracle: "Samite Archer",
		oldB655DarigaazOracle: "Darigaaz, the Igniter", oldB655TrevaOracle: "Treva, the Renewer",
		oldB655GiantOracle: "Callous Giant", oldB655DartOracle: "Lightning Dart", oldB655OverabundOracle: "Overabundance",
		oldB655GhituFireOracle: "Ghitu Fire", oldB655WeaverOracle: "Might Weaver", oldB655NightscapeOracle: "Nightscape Master",
		oldB655ThornscapeOracle: "Thornscape Master", oldB655RaysOracle: "Searing Rays", oldB655PresenceOracle: "Divine Presence",
	}
	for oracle, name := range want {
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

func oldB655Colored(g *game.Game, id uuid.UUID, colors ...string) uuid.UUID {
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, id).Colors = colors })
	return id
}

func oldB655Bear(g *game.Game, owner uuid.UUID, p, tough int, colors ...string) uuid.UUID {
	return oldB655Colored(g, effectstest.B12Creature(g, owner, "Bear", oldB655CreatureTypeLine, p, tough), colors...)
}

func oldB655Cast(t *testing.T, g *game.Game, p *game.Player, name, typeLine, cost, oracle, mana string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	effectstest.FloatForTest(g, p, mana)
	params.Strict = true
	return effectstest.CastFromHandForTest(t, g, p, name, typeLine, cost, oracle, params)
}

func oldB655Ref(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

func TestOldB655SimoonHitsOnlyTheTargetOpponentsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.AdvanceToMain(t, g)
	a := oldB655Bear(g, opp.ID, 2, 3)
	b := oldB655Bear(g, opp.ID, 2, 3)
	mine := oldB655Bear(g, me.ID, 2, 3)
	other := oldB655Bear(g, third.ID, 2, 3)
	oldB655Cast(t, g, me, "Simoon", "Instant", "{R}{G}", oldB655SimoonOracle, "RG", game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if got := effectstest.DamageMarkedOn(g, id); got != 1 {
			t.Errorf("the target opponent's creature has %d damage, want 1", got)
		}
	}
	for _, id := range []uuid.UUID{mine, other} {
		if got := effectstest.DamageMarkedOn(g, id); got != 0 {
			t.Errorf("a creature the target opponent does not control has %d damage", got)
		}
	}
}

func oldB655Dragon(t *testing.T, g *game.Game, me *game.Player, name, oracle string, colors ...string) uuid.UUID {
	t.Helper()
	id := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: name, OracleID: oracle, TypeLine: oldB655LegendaryDragon, Power: 6, Toughness: 6, Keywords: []string{"flying"}})
	return oldB655Colored(g, id, colors...)
}

func TestOldB655DromarBouncesEveryCreatureOfTheChosenColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dromar := oldB655Dragon(t, g, me, "Dromar, the Banisher", oldB655DromarOracle, "W", "U", "B")
	greenA := oldB655Bear(g, opp.ID, 2, 2, "G")
	greenB := oldB655Bear(g, me.ID, 2, 2, "G")
	red := oldB655Bear(g, opp.ID, 2, 2, "R")
	effectstest.AttackWith(t, g, opp.ID, dromar)
	effectstest.ApaMana(me, "C", "C", "U")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "G")
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{greenA, greenB} {
		if g.Battlefield.Contains(id) {
			t.Error("a green creature stayed on the battlefield")
		}
	}
	if !g.Battlefield.Contains(red) || !g.Battlefield.Contains(dromar) {
		t.Error("a creature of another colour was returned")
	}
}

func TestOldB655DromarDeclinedBouncesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dromar := oldB655Dragon(t, g, me, "Dromar, the Banisher", oldB655DromarOracle, "W", "U", "B")
	green := oldB655Bear(g, opp.ID, 2, 2, "G")
	effectstest.AttackWith(t, g, opp.ID, dromar)
	effectstest.ApaMana(me, "C", "C", "U")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(green) {
		t.Error("a declined payment still bounced a creature")
	}
}

func TestOldB655DarigaazDealsDamageForTheChosenColourInTheirHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dar := oldB655Dragon(t, g, me, "Darigaaz, the Igniter", oldB655DarigaazOracle, "B", "R", "G")
	for _, colors := range [][]string{{"R"}, {"R"}, {"R", "G"}, {"G"}, nil} {
		opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Sorcery", Colors: colors, Owner: opp.ID, Controller: opp.ID})
	}
	life := opp.Life
	effectstest.AttackWith(t, g, opp.ID, dar)
	effectstest.ApaMana(me, "C", "C", "R")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "R")
	effectstest.PassPriorityAroundTable(t, g)
	// 6 combat damage, then 3 red cards in the hand.
	if opp.Life != life-6-3 {
		t.Errorf("life %d, want %d (6 combat + 3 red cards)", opp.Life, life-9)
	}
}

func TestOldB655TrevaGainsLifeForEachPermanentOfTheChosenColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	treva := oldB655Dragon(t, g, me, "Treva, the Renewer", oldB655TrevaOracle, "G", "W", "U")
	oldB655Bear(g, opp.ID, 2, 2, "G")
	effectstest.B12Permanent(g, me.ID, "Mystery", "Artifact")
	oldB655Colored(g, effectstest.B12Permanent(g, opp.ID, "Green Wall", "Enchantment"), "G")
	life := me.Life
	effectstest.AttackWith(t, g, opp.ID, treva)
	effectstest.ApaMana(me, "C", "C", "W")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "G")
	effectstest.PassPriorityAroundTable(t, g)
	// Treva, the green Bear and the green enchantment: 3.
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestOldB655SamiteArcherShieldsAndPingsButOnlyOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	archer := effectstest.B12Push(g, me.ID, "Samite Archer", "Creature — Human Cleric Archer", oldB655ArcherOracle, 1, 1)
	life := opp.Life
	effectstest.ApaActivate(t, g, me, archer, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if opp.Life != life-1 {
		t.Errorf("life %d, want %d after the ping", opp.Life, life-1)
	}
	if err := g.ActivateCatalogAbility(me.ID, archer, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err == nil {
		t.Error("the Archer used a second tap ability in one turn")
	}
}

func TestOldB655SamiteArcherPreventsTheNextDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	archer := effectstest.B12Push(g, me.ID, "Samite Archer", "Creature — Human Cleric Archer", oldB655ArcherOracle, 1, 1)
	bear := oldB655Bear(g, me.ID, 2, 5)
	effectstest.ApaActivate(t, g, me, archer, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{oldB655Ref(bear)}})
	effectstest.P2Ping(t, g, bear, bear, 3)
	if got := effectstest.DamageMarkedOn(g, bear); got != 2 {
		t.Errorf("damage %d, want 2 (1 of 3 prevented)", got)
	}
}

func TestOldB655CallousGiantPreventsSmallDamageOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giant := effectstest.B12Push(g, me.ID, "Callous Giant", "Creature — Giant", oldB655GiantOracle, 4, 4)
	src := oldB655Bear(g, opp.ID, 1, 10)
	effectstest.P2Ping(t, g, src, giant, 3)
	effectstest.P2Ping(t, g, src, giant, 2)
	effectstest.P2Ping(t, g, src, giant, 1)
	if got := effectstest.DamageMarkedOn(g, giant); got != 0 {
		t.Errorf("damage %d from sources of 3 or less, want all prevented", got)
	}
	effectstest.P2Ping(t, g, src, giant, 4)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(giant) {
		t.Error("4 damage was prevented")
	}
}

func TestOldB655DivinePresenceCapsBigDamageAtThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Divine Presence", "Enchantment", oldB655PresenceOracle, false)
	tough := oldB655Bear(g, opp.ID, 1, 30)
	src := oldB655Bear(g, me.ID, 1, 30)
	effectstest.P2Ping(t, g, src, tough, 7)
	if got := effectstest.DamageMarkedOn(g, tough); got != 3 {
		t.Errorf("a 7-damage hit left %d damage, want 3", got)
	}
	effectstest.P2Ping(t, g, src, tough, 2)
	if got := effectstest.DamageMarkedOn(g, tough); got != 5 {
		t.Errorf("a 2-damage hit changed the total to %d, want 5", got)
	}
	life := opp.Life
	effectstest.P2Ping(t, g, src, opp.ID, 5)
	if opp.Life != life-3 {
		t.Errorf("life %d, want %d (5 damage to a player becomes 3)", opp.Life, life-3)
	}
}

func TestOldB655LightningDartHitsWhiteAndBlueHarder(t *testing.T) {
	for _, tc := range []struct {
		colors []string
		want   int
	}{{[]string{"W"}, 4}, {[]string{"U"}, 4}, {[]string{"G"}, 1}, {nil, 1}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		bear := oldB655Bear(g, g.Seats[1].ID, 2, 9, tc.colors...)
		oldB655Cast(t, g, me, "Lightning Dart", "Instant", "{1}{R}", oldB655DartOracle, "CR", game.CastSpellParams{Targets: []game.TargetRef{oldB655Ref(bear)}})
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.DamageMarkedOn(g, bear); got != tc.want {
			t.Errorf("colours %v: %d damage, want %d", tc.colors, got, tc.want)
		}
	}
}

func TestOldB655OverabundanceDoublesAnyPlayersLandAndDamagesThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Overabundance", "Enchantment", oldB655OverabundOracle, false)
	for _, p := range []*game.Player{me, opp} {
		forest := effectstest.PushCatalogPermanent(g, p.ID, "Forest", "Basic Land — Forest", "b34bb2dc-c1af-4d77-b0b3-a0fb342a5fc6", false)
		life := p.Life
		effectstest.TapForMana(t, g, p.ID, forest)
		if n := effectstest.PoolCount(p, "G"); n != 2 {
			t.Errorf("%s has %d green mana, want 2 (the land's and Overabundance's)", p.Name, n)
		}
		if p.Life != life-1 {
			t.Errorf("%s life %d, want %d", p.Name, p.Life, life-1)
		}
	}
}

func TestOldB655GhituFireDealsXAndHasFlashForTwoMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	oldB655Cast(t, g, me, "Ghitu Fire", "Sorcery", "{X}{R}", oldB655GhituFireOracle, "CCCR", game.CastSpellParams{XValue: 3, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d, want %d", opp.Life, life-3)
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Ghitu Fire", TypeLine: "Sorcery", ManaCost: "{X}{R}", OracleID: oldB655GhituFireOracle, Owner: me.ID, Controller: me.ID})
	})
	effectstest.FloatForTest(g, me, "CCR")
	targets := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 2, Targets: targets}); err == nil {
		t.Fatal("cast at instant speed without paying {2} more")
	}
	effectstest.FloatForTest(g, me, "CCCCR")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 2, OptionalCosts: []int{0}, Targets: targets}); err != nil {
		t.Fatalf("cast at instant speed paying {2} more: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3-2 {
		t.Errorf("life %d, want %d", opp.Life, life-5)
	}
}

func TestOldB655MightWeaverGivesTrampleToARedOrWhiteCreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	weaver := effectstest.B12Push(g, me.ID, "Might Weaver", oldB655CreatureWizardTyp, oldB655WeaverOracle, 2, 1)
	red := oldB655Bear(g, me.ID, 2, 2, "R")
	green := oldB655Bear(g, me.ID, 2, 2, "G")
	effectstest.FloatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, weaver, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{oldB655Ref(green)}}); err == nil {
		t.Fatal("targeted a green creature")
	}
	effectstest.ApaActivate(t, g, me, weaver, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{oldB655Ref(red)}})
	if !effectstest.HasKeywordOnBattlefield(t, g, red, "trample") {
		t.Error("the red creature did not gain trample")
	}
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	if effectstest.HasKeywordOnBattlefield(t, g, red, "trample") {
		t.Error("trample lasts past the turn")
	}
}

func TestOldB655NightscapeMasterBouncesOrShoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	master := effectstest.B12Push(g, me.ID, "Nightscape Master", "Creature — Zombie Wizard", oldB655NightscapeOracle, 2, 2)
	victim := oldB655Bear(g, opp.ID, 2, 5)
	effectstest.ApaMana(me, "R", "R")
	effectstest.ApaActivate(t, g, me, master, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{oldB655Ref(victim)}})
	if got := effectstest.DamageMarkedOn(g, victim); got != 2 {
		t.Errorf("damage %d, want 2", got)
	}
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	effectstest.AdvanceToMain(t, g2)
	master2 := effectstest.B12Push(g2, me2.ID, "Nightscape Master", "Creature — Zombie Wizard", oldB655NightscapeOracle, 2, 2)
	victim2 := oldB655Bear(g2, opp2.ID, 2, 5)
	effectstest.ApaMana(me2, "U", "U")
	effectstest.ApaActivate(t, g2, me2, master2, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{oldB655Ref(victim2)}})
	if g2.Battlefield.Contains(victim2) {
		t.Error("the creature was not returned to its owner's hand")
	}
}

func TestOldB655ThornscapeMasterShootsAndProtects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	master := effectstest.B12Push(g, me.ID, "Thornscape Master", oldB655CreatureWizardTyp, oldB655ThornscapeOracle, 2, 2)
	victim := oldB655Bear(g, opp.ID, 2, 5)
	effectstest.ApaMana(me, "R", "R")
	effectstest.ApaActivate(t, g, me, master, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{oldB655Ref(victim)}})
	if got := effectstest.DamageMarkedOn(g, victim); got != 2 {
		t.Errorf("damage %d, want 2", got)
	}
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.AdvanceToMain(t, g2)
	master2 := effectstest.B12Push(g2, me2.ID, "Thornscape Master", oldB655CreatureWizardTyp, oldB655ThornscapeOracle, 2, 2)
	friend := oldB655Bear(g2, me2.ID, 2, 2, "G")
	effectstest.ApaMana(me2, "W", "W")
	if err := g2.ActivateCatalogAbility(me2.ID, master2, 1, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{oldB655Ref(friend)}}); err != nil {
		t.Fatalf("activate protection: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	effectstest.AnswerColor(t, g2, me2.ID, "R")
	effectstest.PassPriorityAroundTable(t, g2)
	if !effectstest.HasKeywordOnBattlefield(t, g2, friend, "protection from red") {
		t.Error("the creature did not gain protection from red")
	}
}

func TestOldB655SearingRaysCountsEachPlayersCreaturesOfTheColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.AdvanceToMain(t, g)
	oldB655Bear(g, me.ID, 1, 1, "G")
	oldB655Bear(g, opp.ID, 1, 1, "G")
	oldB655Bear(g, opp.ID, 1, 1, "G")
	oldB655Bear(g, opp.ID, 1, 1, "R")
	oldB655Bear(g, third.ID, 1, 1, "R")
	lives := []int{me.Life, opp.Life, third.Life, g.Seats[3].Life}
	oldB655Cast(t, g, me, "Searing Rays", "Sorcery", "{2}{R}", oldB655RaysOracle, "CCR", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "G")
	effectstest.PassPriorityAroundTable(t, g)
	want := []int{lives[0] - 1, lives[1] - 2, lives[2], lives[3]}
	for i, p := range g.Seats {
		if p.Life != want[i] {
			t.Errorf("seat %d life %d, want %d", i, p.Life, want[i])
		}
	}
}
