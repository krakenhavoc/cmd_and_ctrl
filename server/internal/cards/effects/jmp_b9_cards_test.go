package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// jmp_b9_cards_test.go — deck request #4894 (Jumpstart Azorius), cards
// batch 9: each card against the engine, with the printed refusals.

const (
	jmpB9RousingReadOracle    = "306549ad-84c7-4983-840b-9663799e9f40"
	jmpB9ScholarOracle        = "5f1e9098-f554-4505-974b-cef4b4b7b23d"
	jmpB9SerendibEfreetOracle = "dd520d83-297a-487f-b8f3-4997bbc056e0"
	jmpB9SkyTetherOracle      = "06548af1-4fda-435c-ac32-bbf7afafb34e"
	jmpB9SteelPlumeOracle     = "e7ed9c56-c388-4188-98e2-307d35ca3088"
	jmpB9TalrandsInvOracle    = "43355ac4-bf8c-48f6-a322-fafbc9d132d1"
	jmpB9TideSkimmerOracle    = "8d83dfb3-5662-4109-b465-bf5f257e58a5"
	jmpB9VoyagesEndOracle     = "1c2829b7-f73f-4023-95f3-9cd40b4125d7"
	jmpB9WindreaderOracle     = "5e038a3a-fbab-46c2-8773-1327d953da16"
	jmpB9WindstormDrakeOracle = "16977ebd-6384-4c28-b9c3-448563a07807"
)

func jmpB9Creature(g *game.Game, owner uuid.UUID, name string, power, toughness int, keywords ...string) uuid.UUID {
	return apaPush(g, owner, owner, game.Card{Name: name, TypeLine: "Creature — Test", Power: power, Toughness: toughness, Keywords: keywords})
}

func jmpB9Catalog(g *game.Game, owner uuid.UUID, name, oracle string, power, toughness int) uuid.UUID {
	return apaPush(g, owner, owner, game.Card{Name: name, OracleID: oracle, TypeLine: "Creature — Test", Power: power, Toughness: toughness})
}

func jmpB9PT(g *game.Game, id uuid.UUID) (int, int) {
	c := apaLive(g, id)
	return c.CurrentPower(), c.CurrentToughness()
}

// --- Rousing Read ---------------------------------------------------

func TestJmpB9RousingReadLootsOnEntryAndGivesPlusOneAndFlying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := jmpB9Creature(g, me.ID, "Bear", 2, 2)
	aura := castCatalogSpell(t, g, "Rousing Read", "Enchantment — Aura", jmpB9RousingReadOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Fatalf("hand %d -> %d before the discard, want +2", hand, me.Hand.Size())
	}
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("no discard prompt after the two draws")
	}
	discardFromHand(t, g, me.ID)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d after the loot, want +1", hand, me.Hand.Size())
	}
	if !g.Battlefield.Contains(aura) {
		t.Fatal("the Aura is not on the battlefield")
	}
	if p, tough := jmpB9PT(g, bear); p != 3 || tough != 3 {
		t.Errorf("enchanted creature is %d/%d, want 3/3", p, tough)
	}
	if !hasAbility(effectiveAbilities(t, g, bear), "flying") {
		t.Error("enchanted creature has no flying")
	}
}

func TestJmpB9RousingReadRefusesANonCreatureTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := pushTypedCard(g, me.ID, "Rock", "Artifact", "{1}")
	if err := castCatalogSpellErr(t, g, "Rousing Read", "Enchantment — Aura", jmpB9RousingReadOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}}); err == nil {
		t.Fatal("Rousing Read was cast enchanting an artifact")
	}
}

// --- Sky Tether -----------------------------------------------------

func TestJmpB9SkyTetherGivesDefenderAndRemovesFlying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	bird := jmpB9Creature(g, me.ID, "Bird", 2, 2, "flying")
	other := jmpB9Creature(g, me.ID, "Other Bird", 2, 2, "flying")
	castCatalogSpell(t, g, "Sky Tether", "Enchantment — Aura", jmpB9SkyTetherOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bird}})
	passPriorityAroundTable(t, g)
	abilities := effectiveAbilities(t, g, bird)
	if !hasAbility(abilities, "defender") || hasAbility(abilities, "flying") {
		t.Errorf("enchanted creature has %v, want defender and no flying", abilities)
	}
	if !hasAbility(effectiveAbilities(t, g, other), "flying") || hasAbility(effectiveAbilities(t, g, other), "defender") {
		t.Error("Sky Tether changed a creature it does not enchant")
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bird, opp.ID); err == nil {
		t.Error("the tethered creature was allowed to attack")
	}
	if err := g.DeclareAttacker(other, opp.ID); err != nil {
		t.Errorf("an untethered creature could not attack: %v", err)
	}
}

// --- Serendib Efreet ------------------------------------------------

func TestJmpB9SerendibEfreetDealsOneToItsControllerEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	owner, other := g.Seats[1], g.Seats[2]
	efreet := jmpB9Catalog(g, owner.ID, "Serendib Efreet", jmpB9SerendibEfreetOracle, 3, 4)
	if !hasAbility(effectiveAbilities(t, g, efreet), "flying") {
		t.Error("Serendib Efreet has no flying")
	}
	life, otherLife := owner.Life, other.Life
	advanceToUpkeepOf(t, g, 1)
	if triggerOnStack(g, efreet) == nil {
		t.Fatal("no upkeep trigger on its controller's upkeep")
	}
	passPriorityAroundTable(t, g)
	if owner.Life != life-1 {
		t.Errorf("controller life %d -> %d, want -1", life, owner.Life)
	}
	if other.Life != otherLife {
		t.Error("the damage reached another player")
	}
}

func TestJmpB9SerendibEfreetIgnoresAnotherPlayersUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	efreet := jmpB9Catalog(g, owner.ID, "Serendib Efreet", jmpB9SerendibEfreetOracle, 3, 4)
	advanceToUpkeepOf(t, g, 2)
	if triggerOnStack(g, efreet) != nil {
		t.Fatal("the Efreet triggered on another player's upkeep")
	}
}

// The source left in response: the damage is still dealt (CR 608.2h).
func TestJmpB9SerendibEfreetDealsTheDamageAfterItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	efreet := jmpB9Catalog(g, owner.ID, "Serendib Efreet", jmpB9SerendibEfreetOracle, 3, 4)
	life := owner.Life
	advanceToUpkeepOf(t, g, 1)
	if triggerOnStack(g, efreet) == nil {
		t.Fatal("no upkeep trigger")
	}
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(efreet); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if owner.Life != life-1 {
		t.Errorf("controller life %d -> %d after the Efreet left, want -1", life, owner.Life)
	}
}

// --- Steel-Plume Marshal --------------------------------------------

func TestJmpB9SteelPlumeMarshalPumpsOtherAttackingFliers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	marshal := jmpB9Catalog(g, me.ID, "Steel-Plume Marshal", jmpB9SteelPlumeOracle, 3, 3)
	flier := jmpB9Creature(g, me.ID, "Flier", 1, 1, "flying")
	walker := jmpB9Creature(g, me.ID, "Walker", 1, 1)
	homeFlier := jmpB9Creature(g, me.ID, "Home Flier", 1, 1, "flying")
	declareAttack(t, g, opp.ID, marshal, flier, walker)
	if n := triggersOnStackFrom(g, marshal); n != 1 {
		t.Fatalf("Marshal triggers on the stack = %d, want 1", n)
	}
	passPriorityAroundTable(t, g)
	if p, tough := jmpB9PT(g, flier); p != 3 || tough != 3 {
		t.Errorf("attacking flier is %d/%d, want 3/3", p, tough)
	}
	if p, tough := jmpB9PT(g, walker); p != 1 || tough != 1 {
		t.Errorf("attacking non-flier is %d/%d, want 1/1", p, tough)
	}
	if p, tough := jmpB9PT(g, homeFlier); p != 1 || tough != 1 {
		t.Errorf("flier that stayed home is %d/%d, want 1/1", p, tough)
	}
	if p, tough := jmpB9PT(g, marshal); p != 3 || tough != 3 {
		t.Errorf("the Marshal is %d/%d, want 3/3 (other creatures only)", p, tough)
	}
}

func TestJmpB9SteelPlumeMarshalDoesNotTriggerWhenOthersAttack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	marshal := jmpB9Catalog(g, me.ID, "Steel-Plume Marshal", jmpB9SteelPlumeOracle, 3, 3)
	flier := jmpB9Creature(g, me.ID, "Flier", 1, 1, "flying")
	declareAttack(t, g, opp.ID, flier)
	if n := triggersOnStackFrom(g, marshal); n != 0 {
		t.Fatalf("the Marshal triggered without attacking (%d)", n)
	}
	passPriorityAroundTable(t, g)
	if p, _ := jmpB9PT(g, flier); p != 1 {
		t.Errorf("flier power %d, want 1", p)
	}
}

// --- Talrand's Invocation -------------------------------------------

func TestJmpB9TalrandsInvocationMakesTwoFlyingDrakes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := tokensNamedControlledBy(g, "Drake", me.ID)
	castCatalogSpell(t, g, "Talrand's Invocation", "Sorcery", jmpB9TalrandsInvOracle, nil)
	if tokensNamedControlledBy(g, "Drake", me.ID) != before {
		t.Fatal("the Drakes arrived before the spell resolved")
	}
	passPriorityAroundTable(t, g)
	if got := tokensNamedControlledBy(g, "Drake", me.ID) - before; got != 2 {
		t.Fatalf("Drakes created = %d, want 2", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Drake" || c.Controller != me.ID {
			continue
		}
		live := apaLive(g, c.InstanceID)
		if live.CurrentPower() != 2 || live.CurrentToughness() != 2 || !game.HasKeyword(live, "flying") {
			t.Errorf("Drake is %d/%d flying=%v, want a 2/2 flier", live.CurrentPower(), live.CurrentToughness(), game.HasKeyword(live, "flying"))
		}
		if len(live.Colors) != 1 || live.Colors[0] != "U" {
			t.Errorf("Drake colors = %v, want blue", live.Colors)
		}
	}
}

// --- Tide Skimmer ---------------------------------------------------

func TestJmpB9TideSkimmerDrawsForTwoOrMoreAttackingFliers(t *testing.T) {
	cases := []struct {
		name          string
		skimmerAttack bool
		fliers        int
		walkers       int
		want          int
	}{
		{"Skimmer and one flier", true, 1, 0, 1},
		{"two fliers, Skimmer home", false, 2, 0, 1},
		{"Skimmer and three fliers", true, 3, 0, 1},
		{"Skimmer alone", true, 0, 0, 0},
		{"Skimmer and a non-flier", true, 0, 1, 0},
		{"one flier and two non-fliers", false, 1, 2, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
			skimmer := jmpB9Catalog(g, me.ID, "Tide Skimmer", jmpB9TideSkimmerOracle, 2, 3)
			var attackers []uuid.UUID
			if tc.skimmerAttack {
				attackers = append(attackers, skimmer)
			}
			for i := 0; i < tc.fliers; i++ {
				attackers = append(attackers, jmpB9Creature(g, me.ID, "Flier", 1, 1, "flying"))
			}
			for i := 0; i < tc.walkers; i++ {
				attackers = append(attackers, jmpB9Creature(g, me.ID, "Walker", 1, 1))
			}
			hand := me.Hand.Size()
			declareAttack(t, g, opp.ID, attackers...)
			if n := triggersOnStackFrom(g, skimmer); n != tc.want {
				t.Fatalf("Tide Skimmer triggers = %d, want %d", n, tc.want)
			}
			passPriorityAroundTable(t, g)
			if got := me.Hand.Size() - hand; got != tc.want {
				t.Errorf("cards drawn = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestJmpB9TideSkimmerIgnoresAnOpponentsFliers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	skimmer := jmpB9Catalog(g, opp.ID, "Tide Skimmer", jmpB9TideSkimmerOracle, 2, 3)
	a := jmpB9Creature(g, me.ID, "Flier A", 1, 1, "flying")
	b := jmpB9Creature(g, me.ID, "Flier B", 1, 1, "flying")
	declareAttack(t, g, opp.ID, a, b)
	if n := triggersOnStackFrom(g, skimmer); n != 0 {
		t.Fatalf("an opponent's attack triggered Tide Skimmer (%d)", n)
	}
}

// --- Voyage's End ---------------------------------------------------

func TestJmpB9VoyagesEndBouncesThenScries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	bear := jmpB9Creature(g, opp.ID, "Their Bear", 2, 2)
	castCatalogSpell(t, g, "Voyage's End", "Instant", jmpB9VoyagesEndOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Fatal("the creature was not returned to its owner's hand")
	}
	if scryChoiceFor(g, me.ID) == nil {
		t.Fatal("no scry 1 prompt after the bounce")
	}
	if c := scryChoiceFor(g, me.ID); len(c.ScryCards) != 1 {
		t.Errorf("scry looks at %d cards, want 1", len(c.ScryCards))
	}
	answerScryKeepAll(t, g, me.ID)
}

func TestJmpB9VoyagesEndRefusesANonCreatureAndFizzlesWhenTheTargetLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	rock := pushTypedCard(g, opp.ID, "Rock", "Artifact", "{1}")
	if err := castCatalogSpellErr(t, g, "Voyage's End", "Instant", jmpB9VoyagesEndOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}}); err == nil {
		t.Fatal("Voyage's End targeted an artifact")
	}
	bear := jmpB9Creature(g, opp.ID, "Their Bear", 2, 2)
	castCatalogSpell(t, g, "Voyage's End", "Instant", jmpB9VoyagesEndOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(bear); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) != nil {
		t.Error("a fizzled Voyage's End still scried")
	}
}

// --- Windreader Sphinx ----------------------------------------------

func TestJmpB9WindreaderSphinxTriggersForEachAttackingFlierOfAnyPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	third := g.Seats[(g.Turn.ActiveSeat+2)%4]
	sphinx := jmpB9Catalog(g, opp.ID, "Windreader Sphinx", jmpB9WindreaderOracle, 3, 7)
	a := jmpB9Creature(g, me.ID, "Flier A", 1, 1, "flying")
	b := jmpB9Creature(g, me.ID, "Flier B", 1, 1, "flying")
	walker := jmpB9Creature(g, me.ID, "Walker", 1, 1)
	declareAttack(t, g, third.ID, a, b, walker)
	if n := triggersOnStackFrom(g, sphinx); n != 2 {
		t.Fatalf("Windreader Sphinx triggers = %d, want one per attacking flier (2)", n)
	}
	hand := opp.Hand.Size()
	ask := awaitTriggerPrompt(t, g, opp.ID)
	if err := g.ResolveTriggerPrompt(ask.ID, opp.ID, true); err != nil {
		t.Fatalf("ResolveTriggerPrompt yes: %v", err)
	}
	ask = awaitTriggerPrompt(t, g, opp.ID)
	if err := g.ResolveTriggerPrompt(ask.ID, opp.ID, false); err != nil {
		t.Fatalf("ResolveTriggerPrompt no: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := opp.Hand.Size() - hand; got != 1 {
		t.Errorf("the Sphinx's controller drew %d, want 1 (one yes, one no)", got)
	}
}

func TestJmpB9WindreaderSphinxIgnoresAttackersWithoutFlying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	sphinx := jmpB9Catalog(g, me.ID, "Windreader Sphinx", jmpB9WindreaderOracle, 3, 7)
	walker := jmpB9Creature(g, me.ID, "Walker", 1, 1)
	declareAttack(t, g, opp.ID, walker)
	if n := triggersOnStackFrom(g, sphinx); n != 0 {
		t.Fatalf("a non-flier triggered the Sphinx (%d)", n)
	}
}

// --- Windstorm Drake ------------------------------------------------

func TestJmpB9WindstormDrakePumpsOtherFliersYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drake := jmpB9Catalog(g, me.ID, "Windstorm Drake", jmpB9WindstormDrakeOracle, 3, 3)
	flier := jmpB9Creature(g, me.ID, "Flier", 1, 1, "flying")
	walker := jmpB9Creature(g, me.ID, "Walker", 1, 1)
	theirs := jmpB9Creature(g, opp.ID, "Their Flier", 1, 1, "flying")
	if p, tough := jmpB9PT(g, flier); p != 2 || tough != 1 {
		t.Errorf("your flier is %d/%d, want 2/1", p, tough)
	}
	if p, _ := jmpB9PT(g, walker); p != 1 {
		t.Errorf("your non-flier has power %d, want 1", p)
	}
	if p, _ := jmpB9PT(g, theirs); p != 1 {
		t.Errorf("an opponent's flier has power %d, want 1", p)
	}
	if p, _ := jmpB9PT(g, drake); p != 3 {
		t.Errorf("the Drake has power %d, want 3 (other creatures only)", p)
	}
}

// --- Scholar of the Lost Trove --------------------------------------

func jmpB9ScholarTable(t *testing.T) (*game.Game, *game.Player) {
	g := newCatalogGame(t)
	return g, g.Seats[g.Turn.ActiveSeat]
}

func jmpB9GraveCard(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID})
	return id
}

// jmpB9ScholarToOffer casts the Scholar, targets `pick` with its enters
// trigger and resolves the trigger to its cast offer.
func jmpB9ScholarToOffer(t *testing.T, g *game.Game, me *game.Player, pick uuid.UUID) {
	t.Helper()
	castCatalogSpell(t, g, "Scholar of the Lost Trove", "Creature — Sphinx", jmpB9ScholarOracle, nil)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, pick)
	passPriorityAroundTable(t, g)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, pick)
}

func TestJmpB9ScholarCastsAnInstantForFreeAndExilesIt(t *testing.T) {
	g, me := jmpB9ScholarTable(t)
	instant := jmpB9GraveCard(me, "Old Instant", "Instant", "{3}{U}{U}")
	jmpB9GraveCard(me, "Old Relic", "Artifact", "{4}")
	jmpB9ScholarToOffer(t, g, me, instant)
	pool := me.ManaPool.Total()
	if err := castThroughResolutionCast(t, g, me.ID, instant, game.CastSpellParams{FromZone: "graveyard", Strict: true,
		AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("casting the instant for free: %v", err)
	}
	if me.ManaPool.Total() != pool {
		t.Error("the free cast spent mana")
	}
	castInsideTheResolution(t, g, instant)
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(instant) || !inExile(g, instant) {
		t.Error("the instant did not go to exile instead of the graveyard")
	}
}

func TestJmpB9ScholarCastsAnArtifactForFreeOntoTheBattlefield(t *testing.T) {
	g, me := jmpB9ScholarTable(t)
	relic := jmpB9GraveCard(me, "Old Relic", "Artifact", "{4}")
	jmpB9ScholarToOffer(t, g, me, relic)
	if err := castThroughResolutionCast(t, g, me.ID, relic, game.CastSpellParams{FromZone: "graveyard", Strict: true,
		AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("casting the artifact for free: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(relic) {
		t.Error("the artifact cast this way did not resolve onto the battlefield")
	}
	if inExile(g, relic) {
		t.Error("the artifact was exiled")
	}
}

func TestJmpB9ScholarDeclineLeavesTheCard(t *testing.T) {
	g, me := jmpB9ScholarTable(t)
	sorcery := jmpB9GraveCard(me, "Old Sorcery", "Sorcery", "{2}{U}")
	jmpB9ScholarToOffer(t, g, me, sorcery)
	declineResolutionCast(t, g, me.ID)
	if !me.Graveyard.Contains(sorcery) {
		t.Error("the declined card left the graveyard")
	}
	if perm := g.CastPermissionOnCardByIDForEffect(sorcery); perm != nil {
		t.Errorf("the decline left a permission: %+v", perm)
	}
}

// A creature card, a land, and an opponent's instant are not legal
// targets, so with only those the trigger has nothing to target.
func TestJmpB9ScholarRefusesCreaturesLandsAndOpponentsCards(t *testing.T) {
	g, me := jmpB9ScholarTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	jmpB9GraveCard(me, "Old Bear", "Creature — Bear", "{1}{G}")
	jmpB9GraveCard(me, "Old Land", "Land", "")
	jmpB9GraveCard(opp, "Their Instant", "Instant", "{U}")
	castCatalogSpell(t, g, "Scholar of the Lost Trove", "Creature — Sphinx", jmpB9ScholarOracle, nil)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Fatalf("the trigger offered targets %v with no legal one", p.PickTargetCards)
	}
	if resolutionCastOfferFor(g, me.ID) != nil {
		t.Fatal("a cast was offered with no legal target")
	}
}
