package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bro_b16_test.go — The Brothers' War, batch 16 (#5482): the three
// Commands, Urza, Chief Artificer, Fade from History, The Fall of Kroog,
// Calamity's Wake, Defabricate, Over the Top and Hurkyl's Final Meditation,
// driven through the engine.

const (
	broB16UrzasCommand = "f5c30dec-072c-4f02-b5d8-5eccfd1cffcf"
	broB16KaylasCmd    = "3d937cd9-b1e1-406d-9e29-3b9d3ceea3be"
	broB16GixsCommand  = "9270eb03-4600-40cc-83a5-ea987d5d4978"
	broB16UrzaChief    = "9deced63-80d0-419e-a5ac-71b06ba1459e"
	broB16Fade         = "533767ca-7208-4c88-b160-375479b68285"
	broB16Kroog        = "bfd8d447-7253-40d0-a153-78e9700045d7"
	broB16Wake         = "a6436330-f35a-4d31-9230-03e0c62f9e7e"
	broB16Defabricate  = "60fc5eb0-95d6-49b0-a538-945ef599e484"
	broB16OverTheTop   = "b82d8bdf-e292-490a-aee7-d830cc2c29df"
	broB16Hurkyl       = "a2099033-42d9-47ac-90c7-b41b55b7dadd"
)

func TestBroB16EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		broB16UrzasCommand: "Urza's Command",
		broB16KaylasCmd:    "Kayla's Command",
		broB16GixsCommand:  "Gix's Command",
		broB16UrzaChief:    "Urza, Chief Artificer",
		broB16Fade:         "Fade from History",
		broB16Kroog:        "The Fall of Kroog",
		broB16Wake:         "Calamity's Wake",
		broB16Defabricate:  "Defabricate",
		broB16OverTheTop:   "Over the Top",
		broB16Hurkyl:       "Hurkyl's Final Meditation",
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

// broB16Tokens counts a player's battlefield tokens of one name.
func broB16Tokens(g *game.Game, p uuid.UUID, name string) (n, tapped int) {
	return effectstest.CountTokensNamed(g, p, name)
}

func broB16PT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	var p, tough int
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			t.Fatalf("card %s is gone", id)
		}
		p, tough = c.CurrentPower(), c.CurrentToughness()
	})
	return p, tough
}

func broB16Counters(g *game.Game, id uuid.UUID) int {
	c, _ := effectstest.BattlefieldCard(g, id)
	return c.Counters[game.CounterPlusOne]
}

// --- Kayla's Command ------------------------------------------------

func TestBroB16KaylasCommandMakesAConstructAndGivesDoubleStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastModal(t, g, "Kayla's Command", "Sorcery", broB16KaylasCmd, []int{0, 1},
		[]game.TargetRef{effectstest.ModeRef(game.TargetCard, bear, 1, 0)})
	effectstest.PassPriorityAroundTable(t, g)

	if n, _ := broB16Tokens(g, me.ID, "Construct"); n != 1 {
		t.Fatalf("Constructs = %d, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Construct" {
			if p, tough := broB16PT(t, g, c.InstanceID); p != 2 || tough != 2 {
				t.Errorf("Construct is %d/%d, want 2/2", p, tough)
			}
			if len(c.Colors) != 0 || !c.IsArtifact() {
				t.Errorf("Construct should be a colorless artifact creature, got colors %v", c.Colors)
			}
		}
	}
	if got := broB16Counters(g, bear); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	if !effectstest.P1g2HasKeyword(g, bear, "double strike") {
		t.Error("the creature should have double strike")
	}
}

func TestBroB16KaylasCommandFetchesAPlainsAndGainsLifeThenScries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Forest A", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Plains A", TypeLine: "Basic Land — Plains"},
		game.Card{Name: "Hallowed Test", TypeLine: "Land — Plains Island"},
		game.Card{Name: "Filler", TypeLine: "Creature — Test"},
	)
	life := me.Life
	effectstest.CastModal(t, g, "Kayla's Command", "Sorcery", broB16KaylasCmd, []int{2, 3}, nil)
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Forest A") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Hallowed Test") != uuid.Nil {
		t.Error("only a basic Plains card may be offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Plains A")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B02bHandHasNamed(me, "Plains A") {
		t.Error("the Plains did not reach the hand")
	}
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2", life, me.Life)
	}
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Fatal("scry 2 asked nothing")
	}
}

func TestBroB16KaylasCommandRefusesATargetForTheWrongBullet(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	active := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Kayla's Command", TypeLine: "Sorcery",
		OracleID: broB16KaylasCmd, Owner: active.ID, Controller: active.ID})
	err := g.CastSpell(active.ID, id, game.CastSpellParams{Strict: true, Modes: []int{0, 1},
		Targets: []game.TargetRef{effectstest.ModeRef(game.TargetCard, theirs, 1, 0)}})
	if err != game.ErrIllegalTarget {
		t.Errorf("counter on a creature you don't control: %v, want ErrIllegalTarget", err)
	}
	err = g.CastSpell(active.ID, id, game.CastSpellParams{Strict: true, Modes: []int{0}})
	if err != game.ErrInvalidParam {
		t.Errorf("choose two with one mode: %v, want ErrInvalidParam", err)
	}
}

// --- Gix's Command --------------------------------------------------

func TestBroB16GixsCommandCountersApplyBeforeTheDestroy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Mouse", "Creature — Mouse", 1, 1)
	small := effectstest.B12Creature(g, opp.ID, "Their Imp", "Creature — Imp", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	effectstest.CastModal(t, g, "Gix's Command", "Sorcery", broB16GixsCommand, []int{0, 1},
		[]game.TargetRef{effectstest.ModeRef(game.TargetCard, mine, 0, 0)})
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(mine) {
		t.Fatal("the creature with two new counters (power 3) should survive the destroy")
	}
	if got := broB16Counters(g, mine); got != 2 {
		t.Errorf("counters = %d, want 2", got)
	}
	if !effectstest.P1g2HasKeyword(g, mine, "lifelink") {
		t.Error("the creature should have lifelink")
	}
	if g.Battlefield.Contains(small) {
		t.Error("power 2 or less should have been destroyed")
	}
	if !g.Battlefield.Contains(big) {
		t.Error("a power 3 creature should survive")
	}
}

func TestBroB16GixsCommandReturnsUpToTwoThenEveryOpponentSacrificesTheirBiggest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B17GraveyardCard(me, "Old Ogre", "Creature — Ogre", "{3}")
	b := effectstest.B17GraveyardCard(me, "Old Imp", "Creature — Imp", "{1}")
	c := effectstest.B17GraveyardCard(me, "Old Wolf", "Creature — Wolf", "{2}")
	spell := effectstest.B17GraveyardCard(me, "Old Bolt", "Instant", "{1}")
	oppBig := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 5, 5)
	oppSmall := effectstest.B12Creature(g, opp.ID, "Their Imp", "Creature — Imp", 3, 3)
	thirdBig := effectstest.B12Creature(g, third.ID, "Third Giant", "Creature — Giant", 4, 4)
	mineOnBF := effectstest.B12Creature(g, me.ID, "My Giant", "Creature — Giant", 9, 9)

	effectstest.CastModal(t, g, "Gix's Command", "Sorcery", broB16GixsCommand, []int{2, 3}, nil)
	effectstest.PassPriorityAroundTable(t, g)

	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no return prompt")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 2 {
		t.Errorf("bounds %d..%d, want 0..2", pick.ChooseMin, pick.ChooseMax)
	}
	for _, id := range pick.ChooseCards {
		if id == spell {
			t.Error("a noncreature card was offered")
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) || me.Hand.Contains(c) {
		t.Error("exactly the two chosen creature cards should be in hand")
	}
	// Each opponent now sacrifices a creature with the greatest power
	// among creatures THEY control (one choice here, so it may be automatic).
	for _, who := range []uuid.UUID{opp.ID, third.ID} {
		if effectstest.SacrificeChoiceFor(g, who) != nil {
			effectstest.AnswerSacrifice(t, g, who, map[uuid.UUID]uuid.UUID{opp.ID: oppBig, third.ID: thirdBig}[who])
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(oppBig) || g.Battlefield.Contains(thirdBig) {
		t.Error("each opponent should have sacrificed their greatest-power creature")
	}
	if !g.Battlefield.Contains(oppSmall) {
		t.Error("the smaller creature must stay")
	}
	if !g.Battlefield.Contains(mineOnBF) {
		t.Error("the caster sacrifices nothing")
	}
}

// --- Urza's Command -------------------------------------------------

func TestBroB16UrzasCommandMakesTappedPowerstoneAndConstruct(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastModal(t, g, "Urza's Command", "Instant", broB16UrzasCommand, []int{1, 2}, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if n, tapped := broB16Tokens(g, me.ID, "Powerstone"); n != 1 || tapped != 1 {
		t.Errorf("Powerstones %d (tapped %d), want 1 tapped", n, tapped)
	}
	n, tapped := broB16Tokens(g, me.ID, "Construct")
	if n != 1 || tapped != 1 {
		t.Fatalf("Constructs %d (tapped %d), want 1 tapped", n, tapped)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Construct" {
			// Powerstone and the Construct itself are two artifacts.
			if p, tough := broB16PT(t, g, c.InstanceID); p != 2 || tough != 2 {
				t.Errorf("Construct is %d/%d, want 2/2", p, tough)
			}
		}
	}
}

func TestBroB16UrzasCommandShrinksOpposingCreaturesAndScriesThenDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	hand := me.Hand.Size()
	effectstest.CastModal(t, g, "Urza's Command", "Instant", broB16UrzasCommand, []int{0, 3}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Fatal("scry 1 asked nothing")
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)

	if p, _ := broB16PT(t, g, theirs); p != 1 {
		t.Errorf("their creature power %d, want 1", p)
	}
	if p, _ := broB16PT(t, g, mine); p != 3 {
		t.Errorf("my creature power %d, want 3", p)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want +1", hand, me.Hand.Size())
	}
}

// --- Urza, Chief Artificer ------------------------------------------

func TestBroB16UrzaChiefArtificerAffinityMenaceAndEndStepConstruct(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	rock := effectstest.B12Creature(g, me.ID, "Artifact Bear", "Artifact Creature — Bear", 2, 2)
	rock2 := effectstest.B12Creature(g, me.ID, "Artifact Bear Two", "Artifact Creature — Bear", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	effectstest.B12Permanent(g, me.ID, "Plain Rock", "Artifact")

	// {3}{W}{U}{B} with two artifact creatures: {1}{W}{U}{B}. The plain
	// artifact does not count.
	if _, err := effectstest.CastFromHandErr(g, me, "Urza, Chief Artificer", "Legendary Creature — Human Artificer",
		"{3}{W}{U}{B}", broB16UrzaChief, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast with an empty pool should be refused")
	}
	effectstest.B06AddMana(me, "W", "U", "B", "C")
	urza, err := effectstest.CastFromHandErr(g, me, "Urza, Chief Artificer", "Legendary Creature — Human Artificer",
		"{3}{W}{U}{B}", broB16UrzaChief, game.CastSpellParams{Strict: true})
	if err != nil {
		t.Fatalf("cast with two artifact creatures and four mana: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(urza) {
		t.Fatal("Urza did not enter")
	}
	if !effectstest.P1g2HasKeyword(g, rock, "menace") || !effectstest.P1g2HasKeyword(g, rock2, "menace") {
		t.Error("artifact creatures you control should have menace")
	}
	if effectstest.P1g2HasKeyword(g, plain, "menace") || effectstest.P1g2HasKeyword(g, urza, "menace") {
		t.Error("nonartifact creatures should not have menace")
	}

	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := broB16Tokens(g, me.ID, "Construct"); n != 1 {
		t.Errorf("Constructs after the end step = %d, want 1", n)
	}
}

// --- Fade from History ----------------------------------------------

func TestBroB16FadeFromHistoryGivesBearsToTheOwnersThenDestroysThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	myAura := effectstest.B12Permanent(g, me.ID, "My Ward", "Enchantment")
	theirRock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	theirGolem := effectstest.B12Creature(g, opp.ID, "Their Golem", "Artifact Creature — Golem", 3, 3)
	thirdBear := effectstest.B12Creature(g, third.ID, "Third Bear", "Creature — Bear", 2, 2)

	effectstest.CastCatalogSpell(t, g, "Fade from History", "Sorcery", broB16Fade, nil)
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{myAura, theirRock, theirGolem} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s survived", id)
		}
	}
	if !g.Battlefield.Contains(thirdBear) {
		t.Error("a creature is not an artifact or enchantment")
	}
	for _, tc := range []struct {
		p    *game.Player
		want int
	}{{me, 1}, {opp, 1}, {third, 0}, {g.Seats[3], 0}} {
		if n, _ := broB16Tokens(g, tc.p.ID, "Bear"); n != tc.want {
			t.Errorf("%s has %d Bears, want %d", tc.p.Name, n, tc.want)
		}
	}
}

// --- The Fall of Kroog ----------------------------------------------

func TestBroB16TheFallOfKroogDestroysALandAndDamagesThePlayerAndTheirCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.B12Permanent(g, opp.ID, "Their Mountain", "Basic Land — Mountain")
	myLand := effectstest.B12Permanent(g, me.ID, "My Mountain", "Basic Land — Mountain")
	imp := effectstest.B12Creature(g, opp.ID, "Their Imp", "Creature — Imp", 1, 1)
	ogre := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 1, 1)
	life := opp.Life

	effectstest.CastCatalogSpell(t, g, "The Fall of Kroog", "Sorcery", broB16Kroog, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0},
		{Kind: game.TargetCard, ID: land, Slot: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(land) {
		t.Error("the target land should be destroyed")
	}
	if !g.Battlefield.Contains(myLand) {
		t.Error("my land must stay")
	}
	if opp.Life != life-3 {
		t.Errorf("life %d → %d, want -3", life, opp.Life)
	}
	if g.Battlefield.Contains(imp) {
		t.Error("the 1/1 should die to 1 damage")
	}
	if !g.Battlefield.Contains(ogre) {
		t.Error("the 3/3 survives 1 damage")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("my creature takes no damage")
	}
}

func TestBroB16TheFallOfKroogTheLandMustBeThatPlayersAndTheTargetAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	otherLand := effectstest.B12Permanent(g, third.ID, "Third Mountain", "Basic Land — Mountain")
	myLand := effectstest.B12Permanent(g, me.ID, "My Mountain", "Basic Land — Mountain")
	for name, targets := range map[string][]game.TargetRef{
		"another player's land": {{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0}, {Kind: game.TargetCard, ID: otherLand, Slot: 1}},
		"yourself":              {{Kind: game.TargetPlayer, ID: me.ID, Slot: 0}, {Kind: game.TargetCard, ID: myLand, Slot: 1}},
	} {
		if err := effectstest.CastCatalogSpellErrStrict(t, g, "The Fall of Kroog", "Sorcery", broB16Kroog, targets); err != game.ErrIllegalTarget {
			t.Errorf("%s: %v, want ErrIllegalTarget", name, err)
		}
	}
}

// --- Calamity's Wake ------------------------------------------------

func TestBroB16CalamitysWakeExilesGraveyardsBansNoncreatureSpellsAndExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B17GraveyardCard(me, "My Corpse", "Creature — Test", "{1}")
	theirs := effectstest.B17GraveyardCard(opp, "Their Corpse", "Creature — Test", "{1}")
	id := effectstest.CastCatalogSpell(t, g, "Calamity's Wake", "Instant", broB16Wake, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if me.Graveyard.Contains(mine) || opp.Graveyard.Contains(theirs) || !g.Exile.Contains(mine) || !g.Exile.Contains(theirs) {
		t.Error("every graveyard should be exiled")
	}
	if !g.Exile.Contains(id) || me.Graveyard.Contains(id) {
		t.Error("Calamity's Wake should exile itself instead of going to the graveyard")
	}
	for _, p := range []*game.Player{me, opp} {
		if _, err := effectstest.CastFromHandErr(g, p, "Some Instant", "Instant", "", "", game.CastSpellParams{Strict: true}); err == nil {
			t.Errorf("%s cast a noncreature spell under the ban", p.Name)
		}
	}
	// A creature spell is still castable.
	if _, err := effectstest.CastFromHandErr(g, me, "Some Bear", "Creature — Bear", "", "", game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("a creature spell under the ban: %v", err)
	}
}

// --- Defabricate ----------------------------------------------------

func TestBroB16DefabricateExilesACounteredArtifactSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	rock := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: rock, Name: "Test Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.CastModal(t, g, "Defabricate", "Instant", broB16Defabricate, []int{0},
		[]game.TargetRef{effectstest.ModeRef(game.TargetCard, rock, 0, 0)})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(rock) || me.Graveyard.Contains(rock) || g.Battlefield.Contains(rock) {
		t.Error("the countered artifact spell should be in exile, not the graveyard")
	}
}

func TestBroB16DefabricateRefusesACreatureSpellAndCountersAnAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Test Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Defabricate", TypeLine: "Instant", OracleID: broB16Defabricate, Owner: me.ID, Controller: me.ID})
	})
	err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{0},
		Targets: []game.TargetRef{effectstest.ModeRef(game.TargetCard, bear, 0, 0)}})
	if err != game.ErrIllegalTarget {
		t.Fatalf("a creature spell: %v, want ErrIllegalTarget", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	src := effectstest.PushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	fired := false
	ability := effectstest.TaAnnounce(t, g, opp.ID, src, "Their Rock — do a thing", nil, &fired)
	effectstest.CastModal(t, g, "Defabricate", "Instant", broB16Defabricate, []int{1},
		[]game.TargetRef{effectstest.ModeRef(game.TargetCard, ability, 0, 0)})
	effectstest.PassPriorityAroundTable(t, g)
	if fired {
		t.Error("the countered ability resolved")
	}
	if !g.Battlefield.Contains(src) {
		t.Error("the ability's source stays")
	}
}

// --- Over the Top ---------------------------------------------------

func TestBroB16OverTheTopRevealsPerNonlandPermanentAndPutsPermanentsOntoTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	// I control two nonland permanents and a land; they control one.
	effectstest.B12Creature(g, me.ID, "Mine A", "Creature — Test", 1, 1)
	effectstest.B12Creature(g, me.ID, "Mine B", "Creature — Test", 1, 1)
	effectstest.B12Permanent(g, me.ID, "My Land", "Basic Land — Forest")
	effectstest.B12Creature(g, opp.ID, "Theirs A", "Creature — Test", 1, 1)
	mine := effectstest.MomB08Library(me, "My Bear", "!My Bolt", "My Wolf")
	theirs := effectstest.MomB08Library(opp, "Their Bear", "Their Wolf")

	effectstest.CastCatalogSpell(t, g, "Over the Top", "Sorcery", broB16OverTheTop, nil)
	effectstest.PassPriorityAroundTable(t, g)

	// I revealed 2 (Bear, Bolt): the creature enters, the instant goes to the graveyard; the third stays.
	if !g.Battlefield.Contains(mine[0]) {
		t.Error("my revealed permanent card should enter")
	}
	if !me.Graveyard.Contains(mine[1]) {
		t.Error("my revealed nonpermanent card should go to the graveyard")
	}
	if !me.Library.Contains(mine[2]) {
		t.Error("the unrevealed card stays in the library")
	}
	// They revealed 1 (counted before anything entered, not 2).
	if !g.Battlefield.Contains(theirs[0]) {
		t.Error("their revealed permanent card should enter")
	}
	if !opp.Library.Contains(theirs[1]) {
		t.Error("their count was one: the second card stays (the counts are read before anything enters)")
	}
}

// --- Hurkyl's Final Meditation --------------------------------------

func TestBroB16HurkylsFinalMeditationBouncesNonlandPermanentsAndEndsTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	land := effectstest.B12Permanent(g, opp.ID, "Their Island", "Basic Land — Island")
	turn := g.Turn.Seq

	effectstest.B06AddMana(me, "U", "U", "U", "C", "C", "C", "C")
	effectstest.CastFromHandForTest(t, g, me, "Hurkyl's Final Meditation", "Instant", "{4}{U}{U}{U}", broB16Hurkyl, game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DiscardIfOwed(t, g)

	if !me.Hand.Contains(mine) || !opp.Hand.Contains(theirs) {
		t.Error("nonland permanents should return to their owners' hands")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("lands stay")
	}
	if g.Turn.Seq != turn+1 {
		t.Errorf("turn %d, want the turn to have ended (%d)", g.Turn.Seq, turn+1)
	}
}

func TestBroB16HurkylsFinalMeditationCostsThreeMoreOnOtherPlayersTurns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1] // not the active player
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(opp, "U", "U", "U", "C", "C", "C", "C")
	if _, err := effectstest.CastFromHandErr(g, opp, "Hurkyl's Final Meditation", "Instant", "{4}{U}{U}{U}", broB16Hurkyl,
		game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("seven mana is not enough on another player's turn ({3} more)")
	}
	effectstest.B06AddMana(opp, "C", "C", "C")
	if _, err := effectstest.CastFromHandErr(g, opp, "Hurkyl's Final Meditation", "Instant", "{4}{U}{U}{U}", broB16Hurkyl,
		game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("ten mana on another player's turn: %v", err)
	}
}
