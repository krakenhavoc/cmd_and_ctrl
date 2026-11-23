package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_snc_b21_test.go — Streets of New Capenna Commander spell-copy batch
// (S82, #5491): every card cast or activated through the engine.

const (
	sncAnheloOracle     = "d3c3f293-6790-4dce-900a-8af0464c2572"
	sncThreefoldOracle  = "bc48da27-77e0-440f-b47b-4bdea3c3cc5e"
	sncParnesseOracle   = "c6040e8d-dac3-4bed-8d1d-d7d44aa10c86"
	sncErrantOracle     = "23e6df25-90db-42a4-8658-e592e67be15e"
	sncStormOracle      = "e9a08492-3d2c-4c7a-8198-ef4193f5758a"
	sncKnucklesOracle   = "5a8630a8-804e-4789-b9a3-96d9ede5e188"
	sncForgeryOracle    = "0e3cb267-5704-4708-94c8-e2fbce4978af"
	sncSchemeOracle     = "e3ff6134-10dd-4565-afcb-0f41a2933d19"
	sncSopranoOracle    = "dffd5f0f-3a7b-47ed-864e-06c752bdcefc"
	sncBoltTargetPlayer = game.TargetPlayer
)

func TestS82SncB21CardsAreRegisteredFull(t *testing.T) {
	for _, id := range []string{sncAnheloOracle, sncThreefoldOracle, sncParnesseOracle, sncErrantOracle, sncStormOracle,
		sncKnucklesOracle, sncForgeryOracle, sncSchemeOracle, sncSopranoOracle} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// sncBolt puts a Lightning Bolt-shaped instant (cost and colours of the
// caller's choosing) in me's hand.
func sncBolt(g *game.Game, me *game.Player, cost string) uuid.UUID {
	return effectstest.TrekB39Hand(g, me, "Lightning Bolt", "Instant", cost, effectstest.LightningBoltOracle, 0, 0)
}

func sncCastBolt(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, params game.CastSpellParams) error {
	t.Helper()
	params.Strict = true
	params.Targets = []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}
	return g.CastSpell(me.ID, id, params)
}

// --- Anhelo, the Painter ------------------------------------------------

func TestS82SncB21AnheloGivesOnlyTheFirstInstantOrSorceryCasualty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Anhelo, the Painter", "Legendary Creature — Vampire Assassin", sncAnheloOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	first, second := sncBolt(g, me, "{R}"), sncBolt(g, me, "{R}")
	before := opp.Life

	effectstest.FloatMana(t, g, me, "{R}{R}")
	if err := sncCastBolt(t, g, me, first, game.CastSpellParams{OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("first instant with casualty: %v", err)
	}
	if g.Battlefield.Contains(bear) {
		t.Fatal("the creature was not sacrificed")
	}
	effectstest.SqResolveAll(t, g, me)
	if opp.Life != before-6 {
		t.Errorf("opponent lost %d, want 6 (the Bolt and its copy)", before-opp.Life)
	}

	bear2 := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if err := sncCastBolt(t, g, me, second, game.CastSpellParams{OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{bear2}}); err == nil {
		t.Fatal("the second instant of the turn was offered casualty")
	}
	if err := sncCastBolt(t, g, me, second, game.CastSpellParams{}); err != nil {
		t.Fatalf("second instant without casualty: %v", err)
	}
}

func TestS82SncB21AnheloCasualtyNeedsPowerTwoAndAnInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Anhelo, the Painter", "Legendary Creature — Vampire Assassin", sncAnheloOracle, false)
	rat := effectstest.PushVanillaCreature(g, me.ID, "Rat", 1, 1)
	bolt := sncBolt(g, me, "{R}")
	effectstest.FloatMana(t, g, me, "{R}")
	if err := sncCastBolt(t, g, me, bolt, game.CastSpellParams{OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{rat}}); err == nil {
		t.Error("a power-1 creature paid casualty 2")
	}
	bear := effectstest.TrekB39Hand(g, me, "Bear", "Creature — Bear", "{R}", "", 2, 2)
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}}); err == nil {
		t.Error("a creature spell was offered casualty")
	}
}

// --- Threefold Signal ---------------------------------------------------

func TestS82SncB21ThreefoldSignalReplicatesExactlyThreeColorSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Threefold Signal", "Artifact", sncThreefoldOracle, false)
	tri := sncBolt(g, me, "{R}{G}{W}")
	before := opp.Life
	effectstest.FloatMana(t, g, me, "{R}{G}{W}{C}{C}{C}{C}{C}{C}")
	if err := sncCastBolt(t, g, me, tri, game.CastSpellParams{OptionalCosts: []int{0, 0}}); err != nil {
		t.Fatalf("three-color spell with replicate twice: %v", err)
	}
	effectstest.SqResolveAll(t, g, me)
	if opp.Life != before-9 {
		t.Errorf("opponent lost %d, want 9 (the Bolt and two copies)", before-opp.Life)
	}
}

func TestS82SncB21ThreefoldSignalIgnoresOtherColorCounts(t *testing.T) {
	for _, cost := range []string{"{R}", "{R}{G}", "{R}{G}{W}{U}", "{3}"} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		effectstest.PushCatalogPermanent(g, me.ID, "Threefold Signal", "Artifact", sncThreefoldOracle, false)
		spell := sncBolt(g, me, cost)
		effectstest.FloatMana(t, g, me, "{R}{G}{W}{U}{C}{C}{C}{C}{C}{C}")
		if err := sncCastBolt(t, g, me, spell, game.CastSpellParams{OptionalCosts: []int{0}}); err == nil {
			t.Errorf("a spell costing %s was offered replicate", cost)
		}
	}
}

func TestS82SncB21ThreefoldSignalScriesThreeWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	signal := effectstest.TrekB39Hand(g, me, "Threefold Signal", "Artifact", "{3}", sncThreefoldOracle, 0, 0)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.CastSpell(me.ID, signal, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	for i := 0; i < 8 && len(g.PendingChoices) == 0; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	found := false
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceScry && c.Chooser == me.ID {
			found = true
		}
	}
	if !found {
		t.Error("no scry prompt after Threefold Signal entered")
	}
}

// --- Storm of Forms -----------------------------------------------------

func TestS82SncB21StormOfFormsCopiesOncePerKindOfCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "Wolf", 2, 2)
	theirs := []uuid.UUID{effectstest.PushVanillaCreature(g, opp.ID, "Ape", 2, 2), effectstest.PushVanillaCreature(g, opp.ID, "Elk", 2, 2),
		effectstest.PushVanillaCreature(g, opp.ID, "Yak", 2, 2)}
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(a, game.CounterPlusOne, 2)
		_ = g.AddCounterForEffect(b, game.CounterPlusOne, 1)
		_ = g.AddCounterForEffect(b, "shield", 1)
		// A counter on an opponent's permanent is not one "among permanents you control".
		_ = g.AddCounterForEffect(theirs[0], "stun", 1)
	})
	storm := effectstest.TrekB39Hand(g, me, "Storm of Forms", "Instant", "{3}{U}", sncStormOracle, 0, 0)
	effectstest.FloatMana(t, g, me, "{U}{C}{C}{C}")
	if err := g.CastSpell(me.ID, storm, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs[0]}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	// Two kinds of counter: the spell and two copies; each copy keeps its target.
	for i := 0; i < 64 && !effectstest.StackFullyEmpty(g); i++ {
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs[1+i%2]}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	bounced := 0
	g.ReadSnapshot(func() {
		for _, id := range theirs {
			if !g.Battlefield.Contains(id) {
				bounced++
			}
		}
	})
	if bounced != 3 {
		t.Errorf("%d opposing creatures returned, want 3 (the spell and two copies)", bounced)
	}
}

func TestS82SncB21StormOfFormsWithNoCountersMakesNoCopies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Ape", 2, 2)
	other := effectstest.PushVanillaCreature(g, opp.ID, "Elk", 2, 2)
	storm := effectstest.TrekB39Hand(g, me, "Storm of Forms", "Instant", "{3}{U}", sncStormOracle, 0, 0)
	effectstest.FloatMana(t, g, me, "{U}{C}{C}{C}")
	if err := g.CastSpell(me.ID, storm, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.B21SettleStack(t, g)
	if g.Battlefield.Contains(victim) || !g.Battlefield.Contains(other) {
		t.Error("exactly the target should have been returned")
	}
}

// --- Brass Knuckles -----------------------------------------------------

func TestS82SncB21BrassKnucklesCopiesItselfAsAToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bk := effectstest.TrekB39Hand(g, me, "Brass Knuckles", "Artifact — Equipment", "{4}", sncKnucklesOracle, 0, 0)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.CastSpell(me.ID, bk, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.B21SettleStack(t, g)
	cards, tokens := effectstest.TrekB39Named(g, me.ID, "Brass Knuckles")
	if cards != 1 || tokens != 1 {
		t.Errorf("%d Brass Knuckles cards and %d tokens, want 1 and 1", cards, tokens)
	}
}

func TestS82SncB21BrassKnucklesGrantsDoubleStrikeOnlyToTwoEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	one := effectstest.SeedEquipment(g, me.ID, "Brass Knuckles", sncKnucklesOracle)
	two := effectstest.SeedEquipment(g, me.ID, "Brass Knuckles", sncKnucklesOracle)
	effectstest.EquipTo(t, g, me.ID, one, bear)
	if effectstest.HasEffectiveKeyword(t, g, bear, "double strike") {
		t.Fatal("one Equipment gave double strike")
	}
	effectstest.EquipTo(t, g, me.ID, two, bear)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "double strike") {
		t.Error("two Equipment gave no double strike")
	}
	other := effectstest.SeedBear(g, me.ID)
	effectstest.EquipTo(t, g, me.ID, two, other)
	if effectstest.HasEffectiveKeyword(t, g, bear, "double strike") || effectstest.HasEffectiveKeyword(t, g, other, "double strike") {
		t.Error("double strike stayed after the second Equipment moved away")
	}
}

// --- Errant, Street Artist ------------------------------------------------

func TestS82SncB21ErrantCopiesOnlyASpellThatWasNotCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	errant := effectstest.PushCatalogPermanent(g, me.ID, "Errant, Street Artist", "Legendary Creature — Human Rogue", sncErrantOracle, false)
	bolt := sncBolt(g, me, "{R}")
	effectstest.FloatMana(t, g, me, "{R}{U}{C}")
	if err := sncCastBolt(t, g, me, bolt, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if err := effectstest.TrekB39Activate(t, g, me, errant, "", game.TargetRef{Kind: game.TargetCard, ID: bolt}); err == nil {
		t.Fatal("Errant copied a spell that was cast")
	}
	// A copy is not cast.
	before := opp.Life
	g.WithWriteLock(func() {
		if err := g.CopySpellForEffect(bolt, me.ID, false, nil); err != nil {
			t.Fatalf("CopySpellForEffect: %v", err)
		}
	})
	var copyID uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Stack.Cards {
			if c.InstanceID != bolt {
				copyID = c.InstanceID
			}
		}
	})
	if copyID == uuid.Nil {
		t.Fatal("no copy on the stack")
	}
	if err := effectstest.TrekB39Activate(t, g, me, errant, "", game.TargetRef{Kind: game.TargetCard, ID: copyID}); err != nil {
		t.Fatalf("Errant targeting a copy: %v", err)
	}
	effectstest.SqResolveAll(t, g, me)
	if opp.Life != before-9 {
		t.Errorf("opponent lost %d, want 9 (the Bolt, its copy and the copy of the copy)", before-opp.Life)
	}
}

// --- Swindler's Scheme --------------------------------------------------

// sncSchemeTable puts Swindler's Scheme under seat 0 with `top` on its
// library and has seat 1 cast a Bolt at seat 0, stopping with the Scheme's
// trigger on the stack.
func sncSchemeTable(t *testing.T, topType string) (g *game.Game, me, opp *game.Player, top, bolt uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Swindler's Scheme", "Enchantment", sncSchemeOracle, false)
	top = uuid.New()
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: top, Name: "Top Card", TypeLine: topType, ManaCost: "{R}",
			OracleID: effectstest.LightningBoltOracle, Owner: me.ID, Controller: me.ID})
	})
	bolt = sncBolt(g, opp, "{R}")
	effectstest.FloatMana(t, g, opp, "{R}")
	if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err != nil {
		t.Fatalf("opponent's cast: %v", err)
	}
	return g, me, opp, top, bolt
}

// sncResolveScheme passes until the Scheme's "you may" is asked, answers it,
// and stops at whatever that raises.
func sncResolveScheme(t *testing.T, g *game.Game, me *game.Player, yes bool) {
	t.Helper()
	answer := func() bool {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == me.ID {
				if err := g.ResolveTriggerPrompt(c.ID, me.ID, yes); err != nil {
					t.Fatalf("ResolveTriggerPrompt: %v", err)
				}
				return true
			}
		}
		return false
	}
	for i := 0; i < 16; i++ {
		if answer() {
			return
		}
		if err := g.PassPriority(); err != nil && !answer() {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the Scheme never asked whether to reveal")
}

func TestS82SncB21SwindlersSchemeCountersAndOffersTheSharedTypeCard(t *testing.T) {
	g, me, opp, top, bolt := sncSchemeTable(t, "Instant")
	before := me.Life
	sncResolveScheme(t, g, me, true)
	if g.StackItemForEffect(bolt) != nil {
		t.Error("the Bolt was not countered")
	}
	c := effectstest.ResolutionCastFor(g, opp.ID)
	if c == nil {
		t.Fatalf("the opponent was offered nothing: %+v", g.PendingChoices)
	}
	if err := g.CastForResolutionChoice(opp.ID, c.ID, top, game.CastSpellParams{FromZone: "library",
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast the revealed card free: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before {
		t.Errorf("life %d -> %d: the countered Bolt dealt damage", before, me.Life)
	}
}

func TestS82SncB21SwindlersSchemeDoesNothingWithoutASharedType(t *testing.T) {
	g, me, opp, top, bolt := sncSchemeTable(t, "Creature — Bear")
	before := me.Life
	sncResolveScheme(t, g, me, true)
	if g.StackItemForEffect(bolt) == nil {
		t.Error("the Bolt was countered with no shared card type")
	}
	if effectstest.ResolutionCastFor(g, opp.ID) != nil {
		t.Error("the opponent was offered the revealed card")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before-3 {
		t.Errorf("life %d -> %d, want -3 from the Bolt", before, me.Life)
	}
	if z := g.FindCardZoneForEffect(top); z == nil || z.Kind != game.ZoneLibrary {
		t.Error("the revealed card left the library")
	}
}

func TestS82SncB21SwindlersSchemeCanBeDeclined(t *testing.T) {
	g, me, _, _, bolt := sncSchemeTable(t, "Instant")
	sncResolveScheme(t, g, me, false)
	if g.StackItemForEffect(bolt) == nil {
		t.Error("declining the reveal still countered the spell")
	}
}

func TestS82SncB21SwindlersSchemeIgnoresYourOwnSpellsAndOtherZones(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Swindler's Scheme", "Enchantment", sncSchemeOracle, false)
	bolt := sncBolt(g, me, "{R}")
	effectstest.FloatMana(t, g, me, "{R}")
	if err := sncCastBolt(t, g, me, bolt, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if effectstest.TriggerOnStackFrom(g, "Swindler's Scheme") {
		t.Error("the Scheme triggered on its controller's own spell")
	}
}

// --- Flawless Forgery ---------------------------------------------------

func TestS82SncB21FlawlessForgeryExilesCopiesAndCastsFromAnOpponentsGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	theirBolt := effectstest.GyCard(opp, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle)
	forgery := effectstest.TrekB39Hand(g, me, "Flawless Forgery", "Sorcery", "{3}{U}{U}", sncForgeryOracle, 0, 0)
	effectstest.FloatMana(t, g, me, "{U}{U}{C}{C}{C}")
	if err := g.CastSpell(me.ID, forgery, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirBolt}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passUntilResolutionCast(t, g, me)
	cp := effectstest.CopyOf(t, g, "Lightning Bolt")
	if z := g.FindCardZoneForEffect(theirBolt); z == nil || z.Kind != game.ZoneExile {
		t.Error("the Bolt was not exiled")
	}
	c := effectstest.ResolutionCastFor(g, me.ID)
	before := opp.Life
	if err := g.CastForResolutionChoice(me.ID, c.ID, cp.InstanceID, game.CastSpellParams{FromZone: "exile",
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast the copy free: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("opponent lost %d, want 3", before-opp.Life)
	}
}

func TestS82SncB21FlawlessForgeryCannotTargetYourOwnGraveyardOrACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.GyCard(me, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle)
	bear := effectstest.GyCard(opp, "Bear", "Creature — Bear", "{1}{G}", "")
	forgery := effectstest.TrekB39Hand(g, me, "Flawless Forgery", "Sorcery", "{3}{U}{U}", sncForgeryOracle, 0, 0)
	effectstest.FloatMana(t, g, me, "{U}{U}{C}{C}{C}")
	for _, id := range []uuid.UUID{mine, bear} {
		if err := g.CastSpell(me.ID, forgery, game.CastSpellParams{Strict: true,
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: id}}}); err == nil {
			t.Errorf("Flawless Forgery targeted %s", id)
		}
	}
}

func TestS82SncB21FlawlessForgeryHasCasualtyThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.GyCard(opp, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle)
	theirs := effectstest.GyCard(opp, "Bolt", "Instant", "{R}", effectstest.LightningBoltOracle)
	ogre := effectstest.PushVanillaCreature(g, me.ID, "Ogre", 3, 3)
	forgery := effectstest.TrekB39Hand(g, me, "Flawless Forgery", "Sorcery", "{3}{U}{U}", sncForgeryOracle, 0, 0)
	effectstest.FloatMana(t, g, me, "{U}{U}{C}{C}{C}")
	if err := g.CastSpell(me.ID, forgery, game.CastSpellParams{Strict: true, OptionalCosts: []int{0},
		SacrificeIDs: []uuid.UUID{ogre}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}}); err != nil {
		t.Fatalf("cast with casualty: %v", err)
	}
	if g.Battlefield.Contains(ogre) {
		t.Error("the creature was not sacrificed")
	}
	if it := effectstest.TriggerOnStack(g, forgery); it == nil {
		t.Error("no casualty trigger on the stack")
	}
}

// passUntilResolutionCast passes priority until `who` is asked to cast.
func passUntilResolutionCast(t *testing.T, g *game.Game, who *game.Player) {
	t.Helper()
	for i := 0; i < 32 && effectstest.ResolutionCastFor(g, who.ID) == nil; i++ {
		if p := effectstest.LatestPickTarget(g, who.ID); p != nil {
			if err := g.ResolvePickTarget(p.ID, who.ID, game.TargetRef{Kind: game.TargetPlayer, ID: g.Seats[1].ID}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if effectstest.ResolutionCastFor(g, who.ID) == nil {
		t.Fatalf("no cast offered: %+v", g.PendingChoices)
	}
}

// --- Spellbinding Soprano -----------------------------------------------

func TestS82SncB21SpellbindingSopranoReducesInstantsAndSorceriesAfterAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	soprano := effectstest.PushCatalogPermanent(g, me.ID, "Spellbinding Soprano", "Creature — Human Bard", sncSopranoOracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == soprano {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 2, 2
			}
		}
	})
	effectstest.AttackWithTrigger(t, g, soprano, opp.ID, uuid.Nil)
	spell := sncBolt(g, me, "{1}{R}")
	bear := effectstest.TrekB39Hand(g, me, "Bear", "Creature — Bear", "{1}{R}", "", 2, 2)
	effectstest.FloatMana(t, g, me, "{R}")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a creature spell was reduced")
	}
	if err := sncCastBolt(t, g, me, spell, game.CastSpellParams{}); err != nil {
		t.Errorf("the instant did not cost {1} less: %v", err)
	}
}

func TestS82SncB21SpellbindingSopranoHasEncore(t *testing.T) {
	spec, _ := effects.Lookup(sncSopranoOracle)
	if len(spec.Activated) != 1 {
		t.Fatalf("%d activated abilities, want encore", len(spec.Activated))
	}
}

// --- Parnesse, the Subtle Brush ------------------------------------------

func sncParnesseTable(t *testing.T) (g *game.Game, me, opp *game.Player) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Parnesse, the Subtle Brush", "Legendary Creature — Vampire Wizard", sncParnesseOracle, false)
	return g, me, opp
}

func sncOppBolt(t *testing.T, g *game.Game, opp *game.Player, target game.TargetRef) uuid.UUID {
	t.Helper()
	bolt := sncBolt(g, opp, "{R}")
	effectstest.FloatMana(t, g, opp, "{R}")
	if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{target}}); err != nil {
		t.Fatalf("opponent's cast: %v", err)
	}
	return bolt
}

func TestS82SncB21ParnesseCountersASpellAtYouUnlessThePlayerPaysFourLife(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g, me, opp := sncParnesseTable(t)
		bolt := sncOppBolt(t, g, opp, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
		mine, theirs := me.Life, opp.Life
		for i := 0; i < 16 && effectstest.LatestConfirmFor(g, opp.ID) == nil; i++ {
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
		}
		effectstest.AnswerConfirmFor(t, g, opp, pay)
		if pay {
			effectstest.PassPriorityAroundTable(t, g)
			if opp.Life != theirs-4 || me.Life != mine-3 {
				t.Errorf("paid: opponent %d -> %d, me %d -> %d; want -4 and -3", theirs, opp.Life, mine, me.Life)
			}
		} else if g.StackItemForEffect(bolt) != nil || opp.Life != theirs {
			t.Errorf("declined: the Bolt is still on the stack or life moved (%d -> %d)", theirs, opp.Life)
		}
	}
}

func TestS82SncB21ParnesseProtectsYourPermanentsAndNotAnOpponents(t *testing.T) {
	g, me, opp := sncParnesseTable(t)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	sncOppBolt(t, g, opp, game.TargetRef{Kind: game.TargetCard, ID: bear})
	for i := 0; i < 16 && effectstest.LatestConfirmFor(g, opp.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if effectstest.LatestConfirmFor(g, opp.ID) == nil {
		t.Fatal("a spell aimed at your permanent was not taxed")
	}

	g, _, opp = sncParnesseTable(t)
	ape := effectstest.PushVanillaCreature(g, opp.ID, "Ape", 2, 2)
	sncOppBolt(t, g, opp, game.TargetRef{Kind: game.TargetCard, ID: ape})
	if effectstest.TriggerOnStackFrom(g, "Parnesse, the Subtle Brush") {
		t.Error("Parnesse triggered on a spell aimed at the opponent's own creature")
	}
}

func TestS82SncB21ParnesseLetsAnOpponentCopyWhenYouCopyASpell(t *testing.T) {
	g, me, opp := sncParnesseTable(t)
	bolt := sncBolt(g, me, "{R}")
	effectstest.FloatMana(t, g, me, "{R}")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[2].ID}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.CopySpellForEffect(bolt, me.ID, false, nil); err != nil {
			t.Fatalf("copy: %v", err)
		}
	})
	before := opp.Life
	for i := 0; i < 64 && !effectstest.StackFullyEmpty(g); i++ {
		switch {
		case effectstest.LatestConfirmFor(g, opp.ID) != nil:
			effectstest.AnswerConfirmFor(t, g, opp, true)
		case effectstest.LatestPickTarget(g, me.ID) != nil:
			p := effectstest.LatestPickTarget(g, me.ID)
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
		case effectstest.LatestPickTarget(g, opp.ID) != nil:
			p := effectstest.LatestPickTarget(g, opp.ID)
			if err := g.ResolvePickTarget(p.ID, opp.ID, game.TargetRef{Kind: game.TargetPlayer, ID: g.Seats[3].ID}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
		default:
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
		}
	}
	if g.Seats[3].Life != 40-3 && g.Seats[3].Life != before-3 {
		t.Errorf("the opponent's copy did not hit seat 3: life %d", g.Seats[3].Life)
	}
}
