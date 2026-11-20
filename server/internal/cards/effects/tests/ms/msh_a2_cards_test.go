package testsms

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msh_a2_cards_test.go — Marvel Super Heroes, half A, second batch (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	mshA2AbominationOracle = "72101185-4336-480a-b4d7-f8ceb0007a51"
	mshA2DoombotOracle     = "f24efc60-fdac-4a5d-bcbc-784ec5ae6d96"
	mshA2BoroughOracle     = "6effbf59-3933-4a77-9a6a-fd3bedf45c26"
	mshA2DarkDeedOracle    = "e98cd00c-b812-4c6a-81d4-e3f94b64fc58"
	mshA2CallOracle        = "49a32f34-01a9-4f25-bf38-b4d46ca67b2d"
	mshA2CastleDoomOracle  = "9bd013df-ad75-4099-940b-1765c58faf26"
	mshA2CrowdOracle       = "06296d4d-a1b3-4e9a-a2fd-a0183bb892a1"
	mshA2QuinjetOracle     = "60105eb8-bdca-44b9-a970-68a7615526e8"
	mshA2CubeOracle        = "0866a630-d962-4aa9-a15f-08b2947540ef"
)

func TestMshA2EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		mshA2AbominationOracle: "Abomination, Terrifying Titan",
		mshA2DoombotOracle:     "Aerial Doombot",
		mshA2BoroughOracle:     "Borough Backup",
		mshA2DarkDeedOracle:    "Dark Deed",
		mshA2CallOracle:        "Call Damage Control",
		mshA2CastleDoomOracle:  "Castle Doom",
		mshA2CrowdOracle:       "Crowd of True Believers",
		mshA2QuinjetOracle:     "Dependable Quinjet",
		mshA2CubeOracle:        "Cosmic Cube",
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

// mshA2Named lists the battlefield permanents with a name.
func mshA2Named(g *game.Game, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			out = append(out, c)
		}
	}
	return out
}

// --- Aerial Doombot ------------------------------------------------------------------------

func TestMshA2AerialDoombotPowerUpCostsFiveTheTurnItEnteredAndGivesThreeCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.PowerUpMain(t, g)
	me := g.Seats[0]
	bot := effectstest.CastPowerUpCreature(t, g, me, "Aerial Doombot", "Artifact Creature — Robot Villain", "{U}", "{U}", mshA2DoombotOracle, 1)
	if got := effectstest.ChargedPowerUp(t, g, me.ID, bot); got != "{5}" {
		t.Fatalf("the turn it entered: %q, want {5} ({5}{U} less its {U})", got)
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, bot, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("power up: %v", err)
	}
	effectstest.ResolveTopOfStack(t, g)
	if got := effectstest.PlusOnes(g, bot); got != 3 {
		t.Errorf("+1/+1 counters = %d, want 3", got)
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{U}")
	if err := g.ActivateCatalogAbility(me.ID, bot, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrPowerUpUsed) {
		t.Errorf("second activation: err = %v, want ErrPowerUpUsed", err)
	}
}

func TestMshA2AerialDoombotPaysFullPriceTheNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.PowerUpMain(t, g)
	me := g.Seats[0]
	bot := effectstest.CastPowerUpCreature(t, g, me, "Aerial Doombot", "Artifact Creature — Robot Villain", "{U}", "{U}", mshA2DoombotOracle, 1)
	effectstest.PowerUpPassTo(t, g, 0)
	if got := effectstest.ChargedPowerUp(t, g, me.ID, bot); got != "{5}{U}" {
		t.Errorf("a turn later: %q, want {5}{U}", got)
	}
}

// --- Abomination, Terrifying Titan ---------------------------------------------------------

const mshA2AbominationType = "Legendary Creature — Gamma Villain"

func TestMshA2AbominationPowerUpGrowsThenFightsTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.PowerUpMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	titan := effectstest.CastPowerUpCreature(t, g, me, "Abomination, Terrifying Titan", mshA2AbominationType, "{3}{R/G}", "{C}{C}{C}{R}", mshA2AbominationOracle, 4)
	price := effectstest.ChargedPowerUp(t, g, me.ID, titan)
	if price != "{2}{R/G}" {
		t.Fatalf("the turn it entered: %q, want {2}{R/G} ({5}{R/G}{R/G} less {3}{R/G})", price)
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{R}")
	if err := g.ActivateCatalogAbility(me.ID, titan, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(victim)}); err != nil {
		t.Fatalf("power up: %v", err)
	}
	effectstest.ResolveTopOfStack(t, g)
	if got := effectstest.PlusOnes(g, titan); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	if g.Battlefield.Contains(victim) {
		t.Error("the 3/3 target survived a 5/5 fighting it")
	}
	if c := effectstest.B12Card(t, g, titan); c.DamageMarked != 3 {
		t.Errorf("Abomination marked damage = %d, want 3 (the fight is mutual)", c.DamageMarked)
	}
}

func TestMshA2AbominationPowerUpWithNoTargetOnlyGrows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.PowerUpMain(t, g)
	me := g.Seats[0]
	titan := effectstest.CastPowerUpCreature(t, g, me, "Abomination, Terrifying Titan", mshA2AbominationType, "{3}{R/G}", "{C}{C}{C}{G}", mshA2AbominationOracle, 4)
	effectstest.FloatMana(t, g, me, "{C}{C}{G}")
	if err := g.ActivateCatalogAbility(me.ID, titan, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("power up with up to one target declined: %v", err)
	}
	effectstest.ResolveTopOfStack(t, g)
	if got := effectstest.PlusOnes(g, titan); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
}

func TestMshA2AbominationCannotFightItsOwnSide(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.PowerUpMain(t, g)
	me := g.Seats[0]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	titan := effectstest.CastPowerUpCreature(t, g, me, "Abomination, Terrifying Titan", mshA2AbominationType, "{3}{R/G}", "{C}{C}{C}{R}", mshA2AbominationOracle, 4)
	effectstest.FloatMana(t, g, me, "{C}{C}{R}")
	if err := g.ActivateCatalogAbility(me.ID, titan, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(mine)}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("targeting my own creature: err = %v, want ErrIllegalTarget", err)
	}
}

// --- Borough Backup ------------------------------------------------------------------------

func TestMshA2BoroughBackupCreatesTwoVigilantHeroes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.CastCatalogSpell(t, g, "Borough Backup", "Sorcery", mshA2BoroughOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	heroes := mshA2Named(g, "Hero")
	if len(heroes) != 2 {
		t.Fatalf("%d Hero tokens, want 2", len(heroes))
	}
	for _, h := range heroes {
		if h.Power != 3 || h.Toughness != 2 || !h.IsToken() || len(h.Colors) != 1 || h.Colors[0] != "W" {
			t.Errorf("Hero token = %+v, want a 3/2 white token", h)
		}
		if !game.HasKeyword(&h, "vigilance") {
			t.Error("a Hero token lacks vigilance")
		}
	}
}

func TestMshA2BoroughBackupBasicLandcycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Plains", "Basic Land — Plains"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	id, _ := effectstest.CycleFromHand(t, g, "Borough Backup", "Sorcery", mshA2BoroughOracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled card is not in the graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the landcycling search did not ask")
	}
	if effectstest.HasID(c.DiscardOptions, ids[1]) {
		t.Error("a non-land card is offered")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(ids[0]) {
		t.Error("the basic land is not in hand")
	}
}

// --- Dark Deed -----------------------------------------------------------------------------

func TestMshA2DarkDeedKillsAFourToughnessCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	effectstest.CastCatalogSpell(t, g, "Dark Deed", "Instant", mshA2DarkDeedOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("a 4/4 survived -4/-4")
	}
}

func TestMshA2DarkDeedShrinksABiggerCreatureOnlyUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	wurm := effectstest.B12Creature(g, opp.ID, "Their Wurm", "Creature — Wurm", 6, 6)
	effectstest.CastCatalogSpell(t, g, "Dark Deed", "Instant", mshA2DarkDeedOracle, effectstest.CardRefs(wurm))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HobA4Power(t, g, wurm); got != 2 {
		t.Errorf("power %d, want 2 (6 - 4)", got)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if got := effectstest.HobA4Power(t, g, wurm); got != 6 {
		t.Errorf("power %d after the turn, want 6", got)
	}
}

// --- Call Damage Control -------------------------------------------------------------------

func mshA2Grave(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

func TestMshA2CallDamageControlReturnsTwoChosenCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := mshA2Grave(me, "Old Rock", "Artifact")
	bear := mshA2Grave(me, "Old Bear", "Creature — Bear")
	land := mshA2Grave(me, "Old Forest", "Basic Land — Forest")
	mshA2MustCall(t, g, []int{0, 1}, effectstest.CardRefs(rock, bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(rock) || !me.Hand.Contains(bear) {
		t.Error("the artifact and the creature were not both returned")
	}
	if !me.Graveyard.Contains(land) {
		t.Error("an unchosen land card left the graveyard")
	}
}

func TestMshA2CallDamageControlChoosesOneOrNone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := mshA2Grave(me, "Old Forest", "Basic Land — Forest")
	ench := mshA2Grave(me, "Old Aura", "Enchantment — Aura")
	mshA2MustCall(t, g, []int{3}, effectstest.CardRefs(land))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(land) || me.Hand.Contains(ench) {
		t.Error("only the land card should have returned")
	}
	// "Up to two" includes none.
	mshA2MustCall(t, g, []int{}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(ench) {
		t.Error("a spell that chose no bullet returned a card")
	}
}

func TestMshA2CallDamageControlRefusesThreeAndTheWrongType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := mshA2Grave(me, "Old Rock", "Artifact")
	bear := mshA2Grave(me, "Old Bear", "Creature — Bear")
	ench := mshA2Grave(me, "Old Aura", "Enchantment — Aura")
	if err := mshA2TryCall(t, g, []int{0, 1, 2}, effectstest.CardRefs(rock, bear, ench)); err == nil {
		t.Error("three bullets were accepted for a choose-up-to-two spell")
	}
	if err := mshA2TryCall(t, g, []int{1}, effectstest.CardRefs(rock)); err == nil {
		t.Error("an artifact card was accepted as the creature bullet's target")
	}
}

func mshA2MustCall(t *testing.T, g *game.Game, modes []int, targets []game.TargetRef) {
	t.Helper()
	if err := mshA2TryCall(t, g, modes, targets); err != nil {
		t.Fatalf("CastSpell Call Damage Control: %v", err)
	}
}

func mshA2TryCall(t *testing.T, g *game.Game, modes []int, targets []game.TargetRef) error {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Call Damage Control", TypeLine: "Sorcery",
		OracleID: mshA2CallOracle, Owner: active.ID, Controller: active.ID})
	effectstest.AdvanceToMain(t, g)
	return g.CastSpell(active.ID, id, game.CastSpellParams{Modes: modes, Targets: targets})
}

// --- Castle Doom ---------------------------------------------------------------------------

func TestMshA2CastleDoomAnyColorManaPaysOnlyForArtifactSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.PushCatalogPermanent(g, me.ID, "Castle Doom", "Land", mshA2CastleDoomOracle, false)
	if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	cost, _ := game.ParseCost("{U}")
	rock := game.Card{Name: "Test Rock", TypeLine: "Artifact", ManaCost: "{U}"}
	bear := game.Card{Name: "Test Bear", TypeLine: "Creature — Bear", ManaCost: "{U}"}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(rock)) {
		t.Error("the mana was refused for an artifact spell")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(bear)) {
		t.Error("the mana paid for a non-artifact spell: stronger than printed")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(rock)) {
		t.Error("the mana paid for an ability: it is for spells only")
	}
}

func TestMshA2CastleDoomSacrificesAnArtifactForADoombot(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.PushCatalogPermanent(g, me.ID, "Castle Doom", "Land", mshA2CastleDoomOracle, false)
	rock := effectstest.PushArtifactForTest(g, me.ID, "Mind Stone")
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed as a cost")
	}
	effectstest.ResolveTopOfStack(t, g)
	bots := mshA2Named(g, "Doombot")
	if len(bots) != 1 {
		t.Fatalf("%d Doombot tokens, want 1", len(bots))
	}
	b := bots[0]
	if b.Power != 3 || b.Toughness != 3 || !b.IsToken() || !b.IsCreature() || !b.IsArtifact() {
		t.Errorf("Doombot = %+v, want a 3/3 artifact creature token", b)
	}
	if !strings.Contains(b.TypeLine, "Robot Villain") {
		t.Errorf("Doombot type line %q lacks Robot Villain", b.TypeLine)
	}
}

func TestMshA2CastleDoomIsSorcerySpeedAndNeedsAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.PushCatalogPermanent(g, me.ID, "Castle Doom", "Land", mshA2CastleDoomOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Error("a creature that is not an artifact paid the sacrifice")
	}
	rock := effectstest.PushArtifactForTest(g, me.ID, "Mind Stone")
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}}); err == nil {
		t.Error("activated outside a main phase")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("a refused activation still sacrificed the artifact")
	}
}

// --- Crowd of True Believers ---------------------------------------------------------------

func TestMshA2CrowdPumpsTheLoneAttackerAndGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	crowd := effectstest.PushCatalogPermanent(g, me.ID, "Crowd of True Believers", "Creature — Human Citizen", mshA2CrowdOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, crowd, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HobA4Power(t, g, bear); got != 3 {
		t.Errorf("power %d, want 3", got)
	}
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
}

func TestMshA2CrowdNeedsAnAloneAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	crowd := effectstest.PushCatalogPermanent(g, me.ID, "Crowd of True Believers", "Creature — Human Citizen", mshA2CrowdOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	wolf := effectstest.B12Creature(g, me.ID, "My Wolf", "Creature — Wolf", 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle Ox", "Creature — Ox", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, bear, wolf)
	if err := g.ActivateCatalogAbility(me.ID, crowd, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("one of two attackers: err = %v, want ErrIllegalTarget", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, crowd, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(idle)}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("a creature that is not attacking: err = %v, want ErrIllegalTarget", err)
	}
}

// --- Dependable Quinjet --------------------------------------------------------------------

func TestMshA2QuinjetTapsForAnyColorWhileNotACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	jet := effectstest.PushVehicleForTest(g, me.ID, "Dependable Quinjet", mshA2QuinjetOracle, 3, 3)
	if err := g.ActivateManaAbility(me.ID, jet, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("pool %v, want one {R}", got)
	}
	if c := effectstest.B12Card(t, g, jet); !c.Tapped {
		t.Error("the Quinjet did not tap for mana")
	}
}

func TestMshA2QuinjetCrewsAtFourPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	jet := effectstest.PushVehicleForTest(g, me.ID, "Dependable Quinjet", mshA2QuinjetOracle, 3, 3)
	small := effectstest.PushCrewerForTest(g, me.ID, "Small", 3)
	if err := g.ActivateCatalogAbility(me.ID, jet, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{small}}); err == nil {
		t.Error("crewed with total power 3")
	}
	extra := effectstest.PushCrewerForTest(g, me.ID, "Extra", 1)
	effectstest.CrewBy(t, g, me.ID, jet, small, extra)
	effectstest.ResolveTopOfStack(t, g)
	if c := effectstest.B12Card(t, g, jet); !c.IsCreature() {
		t.Error("the Quinjet did not become a creature")
	}
}

// --- Cosmic Cube ---------------------------------------------------------------------------

// mshA2CubeLibrary stacks, top first: a land, a 2-drop, an 8-drop, a
// 1-drop and a 7-mana sorcery, and a deep card below them.
func mshA2CubeLibrary(me *game.Player) (bear, wurm, elk, sorcery, land, deep uuid.UUID) {
	deep = effectstest.SunbirdLibrary(me, game.Card{Name: "Deep Card", TypeLine: "Basic Land — Plains"})[0]
	ids := effectstest.SunbirdLibrary(me,
		game.Card{Name: "Top Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Cheap Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Huge Wurm", TypeLine: "Creature — Wurm", ManaCost: "{7}{G}", Power: 8, Toughness: 8},
		game.Card{Name: "Cheap Elk", TypeLine: "Creature — Elk", ManaCost: "{G}"},
		game.Card{Name: "Big Sorcery", TypeLine: "Sorcery", ManaCost: "{6}{R}"},
		game.Card{Name: "Sixth Card", TypeLine: "Artifact", ManaCost: "{3}"},
	)
	return ids[1], ids[2], ids[3], ids[4], ids[0], deep
}

func TestMshA2CosmicCubeOffersOnlySpellsUpToTheGreatestAttackingPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Cosmic Cube", "Artifact", mshA2CubeOracle, false)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 1, 1)
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Bear", 2, 2)
	bear, wurm, elk, sorcery, land, deep := mshA2CubeLibrary(me)
	effectstest.DeclareAttack(t, g, opp.ID, small, big)
	effectstest.PassPriorityUntilChoice(t, g)
	c := effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast)
	for _, id := range []uuid.UUID{bear, elk} {
		if !effectstest.HasID(c.DiscardOptions, id) {
			t.Errorf("a spell with mana value <= 2 was not offered: %v", id)
		}
	}
	for name, id := range map[string]uuid.UUID{"wurm": wurm, "sorcery": sorcery, "land": land, "deep": deep} {
		if effectstest.HasID(c.DiscardOptions, id) {
			t.Errorf("%s was offered: over the greatest attacking power (2) or not a spell", name)
		}
	}
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, bear, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	effectstest.CastInsideTheResolution(t, g, bear)
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
		t.Error("the offer outlived its one cast")
	}
	for _, id := range []uuid.UUID{wurm, elk, sorcery, land} {
		if i := effectstest.LibraryIndex(me, id); i < 0 || i > 4 {
			t.Errorf("an uncast looked-at card goes to the bottom, found index %d", i)
		}
	}
	if i := effectstest.LibraryIndex(me, deep); i < 0 {
		t.Error("the seventh card vanished")
	}
}

func TestMshA2CosmicCubeTriggersOncePerAttackNotPerCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	cube := effectstest.PushCatalogPermanent(g, me.ID, "Cosmic Cube", "Artifact", mshA2CubeOracle, false)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 3, 3)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	if got := effectstest.TriggersOnStackFrom(g, cube); got != 1 {
		t.Errorf("%d Cosmic Cube triggers, want 1", got)
	}
}

func TestMshA2CosmicCubeIgnoresAnOpponentsAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cube := effectstest.PushCatalogPermanent(g, opp.ID, "Cosmic Cube", "Artifact", mshA2CubeOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	if got := effectstest.TriggersOnStackFrom(g, cube); got != 0 {
		t.Errorf("an opponent's Cosmic Cube triggered on my attack: %d", got)
	}
}

// --- Giant-Sized Flying Ant ----------------------------------------------------------------

const mshA2AntOracle = "233e3b4b-faaf-451d-8a2e-e9f5a6ced671"

func TestMshA2GiantSizedFlyingAntIsRegisteredFullWithFlashAndFlying(t *testing.T) {
	spec, ok := effects.Lookup(mshA2AntOracle)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Giant-Sized Flying Ant: registered %v, spec %+v", ok, spec)
	}
	if len(spec.PrintedKeywords) != 2 || spec.PrintedKeywords[0] != "flash" || spec.PrintedKeywords[1] != "flying" {
		t.Errorf("keywords %v, want flash and flying", spec.PrintedKeywords)
	}
}

func TestMshA2GiantSizedFlyingAntTapsAnOpponentsPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	theirLand := effectstest.PushCatalogPermanent(g, opp.ID, "Their Forest", "Basic Land — Forest", "", false)
	effectstest.CastHoldCreature(t, g, "Giant-Sized Flying Ant", "Creature — Insect", mshA2AntOracle, 3, 2)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enter trigger did not ask for a mode")
	}
	effectstest.AnswerMode(t, g, c, me.ID, 0)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if !effectstest.HasID(p.PickTargetCards, theirs) || effectstest.HasID(p.PickTargetCards, theirLand) {
		t.Error("only nonland permanents are offered")
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedForTest(t, g, theirs) {
		t.Error("the chosen permanent is not tapped")
	}
}

func TestMshA2GiantSizedFlyingAntUntapsAPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(mine) })
	effectstest.CastHoldCreature(t, g, "Giant-Sized Flying Ant", "Creature — Insect", mshA2AntOracle, 3, 2)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enter trigger did not ask for a mode")
	}
	effectstest.AnswerMode(t, g, c, me.ID, 1)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TappedForTest(t, g, mine) {
		t.Error("the chosen permanent is still tapped")
	}
}
