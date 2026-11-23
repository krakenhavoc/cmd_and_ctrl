package testsso

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsd "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/d"
	cardsdi "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/di"
	cardse "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/e"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b4_cards_test.go — Secrets of Strixhaven Commander, fourth batch
// (#4834). Each test plays the real catalog entry through the engine and
// asserts what the printed text says.

func TestSocB4EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		cardsd.DeekahFractalTheoristOracleID:      "Deekah, Fractal Theorist",
		cardsd.DefacingDuskmageOracleID:           "Defacing Duskmage",
		cardsd.DefacingDuskmageOracleID + "#1":    "Vandal's Edit",
		"f61c1dc4-2f09-4b50-957f-ee656c659072":    "Dina, Essence Brewer",
		"2121bcb4-f5fe-447f-9112-c7e5b5c45baf":    "Dina, Soul Steeper",
		cardsdi.DirgurFocusmageOracleID:           "Dirgur Focusmage",
		cardsdi.DirgurFocusmageOracleID + "#1":    "Braingeyser",
		"ba069664-9f92-47c7-b556-9e5143a22579":    "Doomwake Giant",
		cardse.EccentricPestfinderOracleID:        "Eccentric Pestfinder",
		cardse.EccentricPestfinderOracleID + "#1": "Turn Stones",
		cardse.EiganjoDynastorianOracleID:         "Eiganjo Dynastorian",
		cardse.EiganjoDynastorianOracleID + "#1":  "Replenish",
		cardse.ElusiveOtterOracleID:               "Elusive Otter",
		cardse.ElusiveOtterOracleID + "#1":        "Grove's Bounty",
		"6d588e25-7b4d-43c0-b45b-1b0ffb277a78":    "Elementalist's Palette",
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

// socB4Fractals lists the Fractal tokens `owner` controls.
func socB4Fractals(g *game.Game, owner uuid.UUID) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == "Fractal" {
			out = append(out, c)
		}
	}
	return out
}

// --- Deekah, Fractal Theorist ---------------------------------------------

func TestSocB4DeekahMakesAFractalWithTheSpellsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Deekah, Fractal Theorist", "Legendary Creature — Human Wizard", cardsd.DeekahFractalTheoristOracleID, false)
	effectstest.SocB4Spell(t, g, "Three", "Instant", "{1}{U}{U}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	got := socB4Fractals(g, me.ID)
	if len(got) != 1 {
		t.Fatalf("Fractals: %d, want 1", len(got))
	}
	if got[0].Counters[game.CounterPlusOne] != 3 || got[0].CurrentPower() != 3 {
		t.Errorf("Fractal: %d counters, power %d, want 3 and 3", got[0].Counters[game.CounterPlusOne], got[0].CurrentPower())
	}
	if !effects.IsToken(got[0]) {
		t.Error("the Fractal is not a token")
	}
	// A creature spell is not magecraft.
	effectstest.SocB4Spell(t, g, "Bear", "Creature — Bear", "{1}{G}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(socB4Fractals(g, me.ID)); n != 1 {
		t.Errorf("a creature spell made a Fractal: %d", n)
	}
}

func TestSocB4DeekahFractalFromAZeroManaValueSpellDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Deekah, Fractal Theorist", "Legendary Creature — Human Wizard", cardsd.DeekahFractalTheoristOracleID, false)
	effectstest.SocB4Spell(t, g, "Free", "Instant", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(socB4Fractals(g, me.ID)); n != 0 {
		t.Errorf("a 0/0 Fractal survived: %d on the battlefield (CR 704.5f)", n)
	}
}

// Countered in response to the trigger, the spell is read as it last
// existed on the stack, its X included (CR 608.2h, CR 202.3e).
func TestSocB4DeekahReadsACounteredSpellsLastKnownManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Deekah, Fractal Theorist", "Legendary Creature — Human Wizard", cardsd.DeekahFractalTheoristOracleID, false)
	spell := effectstest.SocB4Spell(t, g, "X Spell", "Sorcery", "{X}{U}", game.CastSpellParams{XValue: 3})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneStack}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, spell); err != nil {
		t.Fatalf("counter the spell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	got := socB4Fractals(g, me.ID)
	if len(got) != 1 || got[0].Counters[game.CounterPlusOne] != 4 {
		t.Fatalf("Fractals %d, want one with 4 counters (X=3 plus {U})", len(got))
	}
}

func TestSocB4DeekahMakesATokenUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	deekah := effectstest.PushCatalogPermanent(g, me.ID, "Deekah, Fractal Theorist", "Legendary Creature — Human Wizard", cardsd.DeekahFractalTheoristOracleID, false)
	effectstest.AdvanceToMain(t, g)
	nontoken := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	err := g.ActivateCatalogAbility(me.ID, deekah, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(nontoken)})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("a nontoken creature: err = %v, want ErrIllegalTarget", err)
	}
	var token uuid.UUID
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, effects.TokenCard("2/2 black Zombie"), 1)
	})
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Zombie" && c.Controller == me.ID {
			token = c.InstanceID
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, deekah, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(token)}); err != nil {
		t.Fatalf("activate on a token: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertRestrictions(t, g, token, game.CantBeBlocked)
}

// --- Defacing Duskmage // Vandal's Edit -----------------------------------

func TestSocB4DefacingDuskmagePreparesOnAnOpponentsSecondDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	c := sos1Prep(me.ID, cardsd.DefacingDuskmageOracleID, []string{"deathtouch"}, "Defacing Duskmage", "Creature — Dog Warlock", "{W}{B}", 2, 2,
		"Vandal's Edit", "Instant", "{1}{W}{B}")
	effectstest.PushBattlefieldCardWithTimestamp(g, c)
	before := g.TurnTallyFor(opp.ID).CardsDrawn
	for i := before; i < 2; i++ {
		if err := g.DrawCard(opp.ID); err != nil {
			t.Fatalf("draw: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got, want := g.IsPreparedForEffect(c.InstanceID), i+1 == 2; got != want {
			t.Fatalf("after the opponent's card %d: prepared = %v, want %v", i+1, got, want)
		}
	}
	hand := me.Hand.Size()
	lives := []int{me.Life, opp.Life, g.Seats[2].Life}
	effectstest.FraBCastCopy(t, g, me, "Vandal's Edit")
	if me.Hand.Size() != hand+2 {
		t.Errorf("Vandal's Edit drew %d, want 2", me.Hand.Size()-hand)
	}
	for i, p := range []*game.Player{me, opp, g.Seats[2]} {
		if p.Life != lives[i]-2 {
			t.Errorf("seat %d life %d -> %d, want -2", i, lives[i], p.Life)
		}
	}
	// Its own controller's second card does not prepare it.
	if g.IsPreparedForEffect(c.InstanceID) {
		t.Error("still prepared after the copy was cast")
	}
}

// --- Dina, Essence Brewer -------------------------------------------------

func TestSocB4DinaEssenceBrewerGrowsAndDrawsOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	dina := effectstest.PushCatalogPermanent(g, me.ID, "Dina, Essence Brewer", "Legendary Creature — Dryad Druid", "f61c1dc4-2f09-4b50-957f-ee656c659072", false)
	fodder := effectstest.B12Creature(g, me.ID, "Ox", "Creature — Ox", 3, 3)
	other := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	target := effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	hand, life := me.Hand.Size(), me.Life
	if err := g.ActivateCatalogAbility(me.ID, dina, 0, game.ActivateAbilityParams{
		Targets: effectstest.B16TargetCard(target), SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d -> %d, want +3", life, me.Life)
	}
	if n := effectstest.B12Card(t, g, target).Counters[game.CounterPlusOne]; n != 3 {
		t.Errorf("target has %d counters, want 3", n)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("the sacrifice drew %d, want 1", me.Hand.Size()-hand)
	}
	// A second sacrifice the same turn draws nothing.
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(other) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("the second sacrifice drew: hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

// --- Dina, Soul Steeper ---------------------------------------------------

func TestSocB4DinaSoulSteeperDrainsAndPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	dina := effectstest.PushCatalogPermanent(g, me.ID, "Dina, Soul Steeper", "Legendary Creature — Dryad Druid", "2121bcb4-f5fe-447f-9112-c7e5b5c45baf", false)
	var lives []int
	for _, p := range g.Seats[1:] {
		lives = append(lives, p.Life)
	}
	sos1GainLife(t, g)
	for i, p := range g.Seats[1:] {
		if p.Life != lives[i]-1 {
			t.Errorf("opponent %d life %d -> %d, want -1", i+1, lives[i], p.Life)
		}
	}
	fodder := effectstest.B12Creature(g, me.ID, "Ox", "Creature — Ox", 4, 4)
	power := effectstest.B12Card(t, g, dina).CurrentPower()
	if err := g.ActivateCatalogAbility(me.ID, dina, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	d := effectstest.B12Card(t, g, dina)
	if d.CurrentPower() != power+4 {
		t.Errorf("Dina's power %d -> %d, want +4", power, d.CurrentPower())
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the Ox was not sacrificed")
	}
}

// --- Dirgur Focusmage // Braingeyser --------------------------------------

func TestSocB4DirgurFocusmagePreparesOnABigSpellFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	c := sos1Prep(me.ID, cardsdi.DirgurFocusmageOracleID, nil, "Dirgur Focusmage", "Creature — Djinn Monk", "{2}{U}", 1, 4,
		"Braingeyser", "Sorcery", "{X}{U}{U}")
	effectstest.PushBattlefieldCardWithTimestamp(g, c)
	effectstest.SocB4Spell(t, g, "Four", "Sorcery", "{3}{U}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("a mana value 4 spell prepared it")
	}
	// X counts on the stack: {X}{U} with X = 4 is mana value 5.
	effectstest.SocB4Spell(t, g, "Big X", "Sorcery", "{X}{U}", game.CastSpellParams{XValue: 4})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("a mana value 5 spell from hand did not prepare it")
	}
	hand := opp.Hand.Size()
	cp, ok := effectstest.PrepareCopyOf(g, "Braingeyser")
	if !ok {
		t.Fatal("no copy of Braingeyser in exile")
	}
	if err := g.CastSpell(me.ID, cp.InstanceID, game.CastSpellParams{
		FromZone: string(game.ZoneExile), XValue: 3,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("cast Braingeyser: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size() - hand; got != 3 {
		t.Errorf("Braingeyser drew %d, want 3", got)
	}
}

func TestSocB4DirgurFocusmageReducesInstantsAndSorceries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	c := sos1Prep(me.ID, cardsdi.DirgurFocusmageOracleID, nil, "Dirgur Focusmage", "Creature — Djinn Monk", "{2}{U}", 1, 4,
		"Braingeyser", "Sorcery", "{X}{U}{U}")
	effectstest.PushBattlefieldCardWithTimestamp(g, c)
	effectstest.AdvanceToMain(t, g)
	me.ManaPool.EmptyPool()
	effectstest.FillPoolColored(me, "U", 1)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Two", TypeLine: "Instant", ManaCost: "{1}{U}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("a {1}{U} instant with {U} under the reduction: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	me.ManaPool.EmptyPool()
	effectstest.FillPoolColored(me, "G", 1)
	bear := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a creature spell was reduced too")
	}
}

// --- Doomwake Giant -------------------------------------------------------

func TestSocB4DoomwakeGiantShrinksOpponentsOnEachConstellation(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	small := effectstest.B12Creature(g, opp.ID, "Squirrel", "Creature — Squirrel", 1, 1)
	big := effectstest.B12Creature(g, opp.ID, "Troll", "Creature — Troll", 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	giant := game.Card{InstanceID: uuid.New(), OracleID: "ba069664-9f92-47c7-b556-9e5143a22579", Name: "Doomwake Giant",
		TypeLine: "Enchantment Creature — Giant", ManaCost: "{4}{B}", Power: 4, Toughness: 6, Owner: me.ID, Controller: me.ID}
	effectstest.CastFromHandAndResolve(t, g, me, giant)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the 1/1 survived -1/-1")
	}
	if p := effectstest.B12Card(t, g, big).CurrentPower(); p != 2 {
		t.Errorf("the Troll's power is %d, want 2", p)
	}
	if p := effectstest.B12Card(t, g, mine).CurrentPower(); p != 2 {
		t.Errorf("your own Bear shrank to %d", p)
	}
	ench := game.Card{InstanceID: uuid.New(), Name: "Glow", TypeLine: "Enchantment", ManaCost: "{W}", Owner: me.ID, Controller: me.ID}
	effectstest.CastFromHandAndResolve(t, g, me, ench)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.B12Card(t, g, big).CurrentPower(); p != 1 {
		t.Errorf("after another enchantment the Troll's power is %d, want 1", p)
	}
}

// --- Eccentric Pestfinder // Turn Stones ----------------------------------

func TestSocB4EccentricPestfinderPreparesAtEndStepAfterLifeGain(t *testing.T) {
	for _, gain := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		c := sos1Prep(me.ID, cardse.EccentricPestfinderOracleID, []string{"trample"}, "Eccentric Pestfinder", "Creature — Troll Druid", "{2}{B}{G}", 5, 5,
			"Turn Stones", "Sorcery", "{B}{G}")
		effectstest.PushBattlefieldCardWithTimestamp(g, c)
		if gain {
			sos1GainLife(t, g)
		}
		effectstest.AdvanceToEndStep(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if got := g.IsPreparedForEffect(c.InstanceID); got != gain {
			t.Errorf("gained life %v: prepared = %v", gain, got)
		}
	}
}

func TestSocB4TurnStonesMakesAPestPerOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	c := sos1Prep(me.ID, cardse.EccentricPestfinderOracleID, []string{"trample"}, "Eccentric Pestfinder", "Creature — Troll Druid", "{2}{B}{G}", 5, 5,
		"Turn Stones", "Sorcery", "{B}{G}")
	effectstest.PushBattlefieldCardWithTimestamp(g, c)
	if _, err := g.BecomePreparedForEffect(c.InstanceID); err != nil {
		t.Fatalf("prepare: %v", err)
	}
	effectstest.FraBCastCopy(t, g, me, "Turn Stones")
	n := 0
	for _, p := range g.Battlefield.Cards {
		if p.Controller == me.ID && p.Name == "Pest" {
			n++
		}
	}
	if n != 3 {
		t.Errorf("Pests: %d, want 3 (one per opponent)", n)
	}
}

// --- Eiganjo Dynastorian // Replenish -------------------------------------

func TestSocB4EiganjoDynastorianPreparesOnTwoAttackers(t *testing.T) {
	for _, attackers := range []int{1, 2} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		c := sos1Prep(me.ID, cardse.EiganjoDynastorianOracleID, []string{"vigilance"}, "Eiganjo Dynastorian", "Creature — Fox Advisor", "{2}{W}", 2, 3,
			"Replenish", "Sorcery", "{3}{W}")
		effectstest.PushBattlefieldCardWithTimestamp(g, c)
		var ids []uuid.UUID
		for i := 0; i < attackers; i++ {
			ids = append(ids, effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
		}
		effectstest.DeclareAttack(t, g, opp.ID, ids...)
		effectstest.PassPriorityAroundTable(t, g)
		if got, want := g.IsPreparedForEffect(c.InstanceID), attackers >= 2; got != want {
			t.Errorf("%d attackers: prepared = %v, want %v", attackers, got, want)
		}
	}
}

func TestSocB4EiganjoReplenishReturnsEnchantments(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	c := sos1Prep(me.ID, cardse.EiganjoDynastorianOracleID, []string{"vigilance"}, "Eiganjo Dynastorian", "Creature — Fox Advisor", "{2}{W}", 2, 3,
		"Replenish", "Sorcery", "{3}{W}")
	effectstest.PushBattlefieldCardWithTimestamp(g, c)
	if _, err := g.BecomePreparedForEffect(c.InstanceID); err != nil {
		t.Fatalf("prepare: %v", err)
	}
	ench := game.Card{InstanceID: uuid.New(), Name: "Glow", TypeLine: "Enchantment", Owner: me.ID, Controller: me.ID}
	me.Graveyard.PushTop(ench)
	effectstest.FraBCastCopy(t, g, me, "Replenish")
	if !g.Battlefield.Contains(ench.InstanceID) && effectstest.FindBattlefieldByName(g, "Glow") == uuid.Nil {
		t.Error("Replenish did not return the enchantment")
	}
}

// --- Elusive Otter // Grove's Bounty --------------------------------------

func TestSocB4ElusiveOtterCantBeBlockedBySmallerCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	otter := effectstest.YsAAdventure(me.ID, cardse.ElusiveOtterOracleID,
		game.Face{Name: "Elusive Otter", TypeLine: "Creature — Otter", ManaCost: "{U}", Power: 1, Toughness: 1},
		game.Face{Name: "Grove's Bounty", TypeLine: "Sorcery — Adventure", ManaCost: "{X}{G}"})
	effectstest.PushBattlefieldCardWithTimestamp(g, otter)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	peer := effectstest.B12Creature(g, opp.ID, "Peer", "Creature — Otter", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(otter.InstanceID, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, otter.InstanceID); err == nil {
		t.Error("a power 0 creature blocked the power 1 Otter")
	}
	if err := g.DeclareBlocker(peer, otter.InstanceID); err != nil {
		t.Errorf("an equal-power creature could not block: %v", err)
	}
}

func TestSocB4GrovesBountyDistributesXCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	otter := effectstest.YsAAdventure(me.ID, cardse.ElusiveOtterOracleID,
		game.Face{Name: "Elusive Otter", TypeLine: "Creature — Otter", ManaCost: "{U}", Power: 1, Toughness: 1},
		game.Face{Name: "Grove's Bounty", TypeLine: "Sorcery — Adventure", ManaCost: "{X}{G}"})
	me.Hand.PushTop(otter)
	if err := g.CastSpell(me.ID, otter.InstanceID, game.CastSpellParams{
		Face: 1, XValue: 3,
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}},
		Distribution: map[uuid.UUID]int{a: 2, b: 1},
	}); err != nil {
		t.Fatalf("cast Grove's Bounty: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if na, nb := effectstest.B12Card(t, g, a).Counters[game.CounterPlusOne], effectstest.B12Card(t, g, b).Counters[game.CounterPlusOne]; na != 2 || nb != 1 {
		t.Errorf("counters %d and %d, want 2 and 1", na, nb)
	}
	if !g.Exile.Contains(otter.InstanceID) {
		t.Error("the adventure did not go to exile")
	}
}

// --- Elementalist's Palette -----------------------------------------------

func TestSocB4ElementalistsPaletteChargesOnXSpellsAndPaysOnlyXCosts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	palette := effectstest.PushCatalogPermanent(g, me.ID, "Elementalist's Palette", "Artifact", "6d588e25-7b4d-43c0-b45b-1b0ffb277a78", false)
	effectstest.SocB4Spell(t, g, "Plain", "Sorcery", "{2}{U}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Card(t, g, palette).Counters[game.CounterCharge]; n != 0 {
		t.Fatalf("a spell without {X} charged it: %d", n)
	}
	effectstest.SocB4Spell(t, g, "X Spell", "Sorcery", "{X}{U}", game.CastSpellParams{XValue: 1})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Card(t, g, palette).Counters[game.CounterCharge]; n != 2 {
		t.Fatalf("charge counters %d, want 2", n)
	}

	me.ManaPool.EmptyPool()
	if err := g.ActivateManaAbility(me.ID, palette, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for {C}: %v", err)
	}
	if got := me.ManaPool.Total(); got != 2 {
		t.Fatalf("pool has %d mana, want 2: %+v", got, me.ManaPool)
	}
	effectstest.FillPoolColored(me, "U", 1)
	plain := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: plain, Name: "Plain", TypeLine: "Sorcery", ManaCost: "{2}{U}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, plain, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("the {X}-only mana paid a cost without {X}")
	}
	xs := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: xs, Name: "X Spell", TypeLine: "Sorcery", ManaCost: "{X}{U}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, xs, game.CastSpellParams{Strict: true, XValue: 2}); err != nil {
		t.Fatalf("the {X}-only mana could not pay an {X} spell: %v", err)
	}
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("pool has %d mana left, want 0", got)
	}
}
