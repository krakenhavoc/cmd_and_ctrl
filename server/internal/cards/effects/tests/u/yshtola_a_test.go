package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsmi "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/mi"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// yshtola_a_test.go — the Y'shtola deck request (#3248), slice A: the
// cards existing seams carry, plus the Kambal review.

const (
	ysCrushingOracle  = "77a1958e-ee24-4305-9dad-da8372803be8"
	ysUndermineOracle = "6bcfe240-c165-49f2-a90a-be05c05498a1"
	ysIrenicusOracle  = "a3faf698-1dac-45ef-9c43-c8cead489ffc"
	ysMindspliceOracl = "4d0760fb-4e04-4f29-a268-14c13c30232d"
	ysDelneyOracle    = "245d0ccf-87b6-460a-8b99-9e2079f2d375"
	ysGrahaTiaOracle  = "f558c758-8f12-45f9-8683-eee33ba1fb0d"
)

func TestYshtolaSliceAIsRegistered(t *testing.T) {
	for oracle, name := range map[string]string{
		ysCrushingOracle:         "Crushing Disappointment",
		ysUndermineOracle:        "Undermine",
		ysIrenicusOracle:         "Irenicus's Vile Duplication",
		ysMindspliceOracl:        "Mindsplice Apparatus",
		ysDelneyOracle:           "Delney, Streetwise Lookout",
		ysGrahaTiaOracle:         "G'raha Tia, Scion Reborn",
		effectstest.KambalOracle: "Kambal, Consul of Allocation",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s registered as %q", oracle, spec.Name)
		}
		if spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: Completeness = %v, want full", name, spec.Completeness)
		}
	}
}

func TestCrushingDisappointmentEachPlayerLosesTwoThenYouDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		life[p.ID] = p.Life
	}
	effectstest.CastCatalogSpell(t, g, "Crushing Disappointment", "Instant", ysCrushingOracle, nil)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if p.Life != life[p.ID]-2 {
			t.Errorf("%s: life %d → %d, want -2", p.Name, life[p.ID], p.Life)
		}
	}
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("caster hand %d → %d, want +2", hand, got)
	}
}

func TestUndermineCountersAndItsControllerLosesThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: victim, Name: "Doomed Spell", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(opp.ID, victim, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the victim: %v", err)
	}
	myLife, theirLife := me.Life, opp.Life
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Undermine", TypeLine: "Instant", OracleID: ysUndermineOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("cast Undermine: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(victim) {
		t.Error("the countered spell is not in its owner's graveyard")
	}
	if opp.Life != theirLife-3 {
		t.Errorf("victim's controller life %d → %d, want -3", theirLife, opp.Life)
	}
	if me.Life != myLife {
		t.Errorf("Undermine's caster lost life: %d → %d", myLife, me.Life)
	}
}

func TestIrenicussVileDuplicationMakesAFlyingNonlegendaryCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hero := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Legendary Bear", TypeLine: "Legendary Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Irenicus's Vile Duplication", "Sorcery", ysIrenicusOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: hero}})
	effectstest.PassPriorityAroundTable(t, g)
	var copies []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Legendary Bear" && c.InstanceID != hero {
			copies = append(copies, c)
		}
	}
	if len(copies) != 1 {
		t.Fatalf("token copies = %d, want 1 (the legend rule must not have eaten one)", len(copies))
	}
	tok := copies[0]
	if !effects.IsToken(tok) {
		t.Errorf("the copy is not a token: %q", tok.TypeLine)
	}
	if tok.IsLegendary() {
		t.Errorf("the copy is legendary: %q", tok.TypeLine)
	}
	if !effects.KeywordSliceContains(tok.Keywords, "flying") {
		t.Errorf("the copy has no flying: %v", tok.Keywords)
	}
	if !g.Battlefield.Contains(hero) {
		t.Error("the original left the battlefield")
	}
}

func TestMindspliceApparatusGrowsAndDiscountsInstantsAndSorceries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Mindsplice Apparatus", "Artifact", ysMindspliceOracl, false)
	if got := effectstest.PriceInHand(t, g, me, "Divination", "Sorcery", "{2}{U}"); got != 3 {
		t.Errorf("no counters: %d, want 3", got)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, cardsmi.MindspliceOilCounter, 2) })
	if got := effectstest.PriceInHand(t, g, me, "Divination", "Sorcery", "{2}{U}"); got != 1 {
		t.Errorf("two oil counters, sorcery: %d, want 1", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Counterspell", "Instant", "{U}{U}"); got != 2 {
		t.Errorf("two oil counters, {U}{U}: %d, want 2 (coloured pips stand)", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Golem", "Artifact Creature — Golem", "{4}"); got != 4 {
		t.Errorf("creature spell: %d, want 4 (undiscounted)", got)
	}
	if got := effectstest.PriceInHand(t, g, them, "Divination", "Sorcery", "{2}{U}"); got != 3 {
		t.Errorf("opponent's sorcery: %d, want 3 (undiscounted)", got)
	}

	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id && c.Counters[cardsmi.MindspliceOilCounter] != 3 {
			t.Errorf("oil counters after your upkeep = %d, want 3", c.Counters[cardsmi.MindspliceOilCounter])
		}
	}
}

func TestDelneysCreaturesCantBeBlockedByPowerThreeOrMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Delney, Streetwise Lookout", "Legendary Creature — Human Scout", ysDelneyOracle, 2, 2)
	small := effectstest.B12Creature(g, me.ID, "Small Attacker", "Creature — Bear", 2, 2)
	large := effectstest.B12Creature(g, me.ID, "Large Attacker", "Creature — Giant", 3, 3)
	bigBlocker := effectstest.B12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	smallBlocker := effectstest.B12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.BrAttack(t, g, small, large)

	offered := effectstest.BrOffered(t, g, opp.ID)
	if effectstest.BrOffers(offered, bigBlocker, small) || !effectstest.BrOffers(offered, smallBlocker, small) || !effectstest.BrOffers(offered, bigBlocker, large) {
		t.Errorf("the enumerator disagrees with the rule: %v", offered)
	}
	effectstest.BrRefusal(t, g.DeclareBlocker(bigBlocker, small), game.BlockReasonCantBeBlockedBy)
	if err := g.DeclareBlocker(bigBlocker, large); err != nil {
		t.Fatalf("a power-3 attacker is outside the rule: %v", err)
	}
	if err := g.DeclareBlocker(smallBlocker, small); err != nil {
		t.Fatalf("a power-2 blocker may block: %v", err)
	}
}

func TestDelneyDoublesTriggersOfYourSmallCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Delney, Streetwise Lookout", "Legendary Creature — Human Scout", ysDelneyOracle, 2, 2)
	effectstest.PushCatalogPermanent(g, me.ID, "Young Pyromancer", "Creature — Human Shaman", effectstest.B07YoungPyromancerOracle, false)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Gl1bPyromancerTokens(g, me.ID); got != 2 {
		t.Errorf("Elementals from a power-1 Pyromancer = %d, want 2", got)
	}

	big := effectstest.NewCatalogGame(t)
	me, opp = big.Seats[0], big.Seats[1]
	effectstest.B12Push(big, me.ID, "Delney, Streetwise Lookout", "Legendary Creature — Human Scout", ysDelneyOracle, 2, 2)
	effectstest.PushBattlefieldCardWithTimestamp(big, game.Card{
		InstanceID: uuid.New(), Name: "Young Pyromancer", TypeLine: "Creature — Human Shaman",
		OracleID: effectstest.B07YoungPyromancerOracle, Power: 3, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	effectstest.CastCatalogSpell(t, big, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, big)
	if got := effectstest.Gl1bPyromancerTokens(big, me.ID); got != 1 {
		t.Errorf("Elementals from a power-3 Pyromancer = %d, want 1", got)
	}
}

// ysCastNoncreature casts a sorcery with the given mana cost from me's
// hand and settles priority up to G'raha's prompt.
func ysCastNoncreature(t *testing.T, g *game.Game, me *game.Player, manaCost string) {
	t.Helper()
	c := game.Card{InstanceID: uuid.New(), Name: "Ponder Deeply", TypeLine: "Sorcery", ManaCost: manaCost, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func ysGrahaPromptOpen(g *game.Game) bool {
	return effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm) != nil
}

// ysGrahaAnswer answers G'raha's resolution-time prompt and settles
// the stack.
func ysGrahaAnswer(t *testing.T, g *game.Game, me *game.Player, pay bool) {
	t.Helper()
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil || ask.Chooser != me.ID {
		t.Fatalf("expected G'raha Tia's pay prompt, got %+v", g.PendingChoices)
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, pay); err != nil {
		t.Fatalf("ResolveConfirm(%v): %v", pay, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestGrahaTiaPaysXLifeAtResolutionForAHeroOnceEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "G'raha Tia, Scion Reborn", "Legendary Creature — Cat Wizard", ysGrahaTiaOracle, false)
	effectstest.AdvanceToMain(t, g)
	life := me.Life

	// The "may" is not asked as the trigger goes on the stack.
	ysCastNoncreature(t, g, me, "{2}{U}")
	if me.Life != life {
		t.Fatalf("life paid before resolution: %d → %d", life, me.Life)
	}
	if effectstest.LatestChoiceOfKind(g, game.PendingChoiceTriggerPrompt) != nil {
		t.Fatal("the trigger was asked about when it went on the stack")
	}
	ysGrahaAnswer(t, g, me, true)
	if me.Life != life-3 {
		t.Errorf("life %d → %d, want -3 for a mana value 3 spell", life, me.Life)
	}
	heroes := effectstest.Gl1bNamed(g, me.ID, "Hero")
	if len(heroes) != 1 {
		t.Fatalf("Hero tokens = %d, want 1", len(heroes))
	}
	if got := heroes[0].Counters[game.CounterPlusOne]; got != 3 {
		t.Errorf("Hero +1/+1 counters = %d, want 3", got)
	}

	// Done once this turn: the next noncreature spell offers nothing.
	ysCastNoncreature(t, g, me, "{1}{U}")
	if ysGrahaPromptOpen(g) {
		t.Fatal("G'raha Tia offered a second Hero this turn")
	}
}

func TestGrahaTiaDeclinedStaysAvailableAndIgnoresCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "G'raha Tia, Scion Reborn", "Legendary Creature — Cat Wizard", ysGrahaTiaOracle, false)
	effectstest.AdvanceToMain(t, g)

	creature := game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(creature)
	if err := g.CastSpell(me.ID, creature.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if ysGrahaPromptOpen(g) {
		t.Fatal("a creature spell triggered Throw Wide the Gates")
	}

	ysCastNoncreature(t, g, me, "{1}{U}")
	ysGrahaAnswer(t, g, me, false)
	if n := len(effectstest.Gl1bNamed(g, me.ID, "Hero")); n != 0 {
		t.Fatalf("declined, yet %d Hero tokens", n)
	}
	life := me.Life
	ysCastNoncreature(t, g, me, "{1}{U}")
	ysGrahaAnswer(t, g, me, true)
	if me.Life != life-2 || len(effectstest.Gl1bNamed(g, me.ID, "Hero")) != 1 {
		t.Errorf("after a decline the next spell should still pay out: life %d → %d, heroes %d", life, me.Life, len(effectstest.Gl1bNamed(g, me.ID, "Hero")))
	}
}

// A payment that can't be made is not offered and does nothing, and it
// does not use up the turn: the next, cheaper spell still pays out.
func TestGrahaTiaCannotPayDoesNothingAndKeepsTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "G'raha Tia, Scion Reborn", "Legendary Creature — Cat Wizard", ysGrahaTiaOracle, false)
	effectstest.AdvanceToMain(t, g)
	me.Life = 2

	ysCastNoncreature(t, g, me, "{2}{U}")
	if ysGrahaPromptOpen(g) {
		t.Fatal("asked to pay 3 life with 2 life")
	}
	if me.Life != 2 || len(effectstest.Gl1bNamed(g, me.ID, "Hero")) != 0 {
		t.Fatalf("unpayable: life %d, heroes %d, want 2 and 0", me.Life, len(effectstest.Gl1bNamed(g, me.ID, "Hero")))
	}

	ysCastNoncreature(t, g, me, "{U}")
	ysGrahaAnswer(t, g, me, true)
	if me.Life != 1 || len(effectstest.Gl1bNamed(g, me.ID, "Hero")) != 1 {
		t.Errorf("a payable spell after an unpayable one: life %d, heroes %d, want 1 and 1", me.Life, len(effectstest.Gl1bNamed(g, me.ID, "Hero")))
	}
}

// Life lost between the trigger and its answer: saying yes no longer
// pays, so no Hero, and the turn's use is still there.
func TestGrahaTiaLifeGoneBeforeTheYesPaysNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "G'raha Tia, Scion Reborn", "Legendary Creature — Cat Wizard", ysGrahaTiaOracle, false)
	effectstest.AdvanceToMain(t, g)

	ysCastNoncreature(t, g, me, "{2}{U}")
	me.Life = 1
	ysGrahaAnswer(t, g, me, true)
	if me.Life != 1 || len(effectstest.Gl1bNamed(g, me.ID, "Hero")) != 0 {
		t.Fatalf("life %d, heroes %d, want 1 and 0", me.Life, len(effectstest.Gl1bNamed(g, me.ID, "Hero")))
	}
	me.Life = 2
	ysCastNoncreature(t, g, me, "{U}")
	ysGrahaAnswer(t, g, me, true)
	if me.Life != 1 || len(effectstest.Gl1bNamed(g, me.ID, "Hero")) != 1 {
		t.Errorf("turn's use was spent by an unpaid answer: life %d, heroes %d", me.Life, len(effectstest.Gl1bNamed(g, me.ID, "Hero")))
	}
}

func TestKambalDrainsAnOpponentsNoncreatureSpellButNotACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Kambal, Consul of Allocation", "Legendary Creature — Human Advisor", effectstest.KambalOracle, false)
	myLife, theirLife := me.Life, opp.Life
	effectstest.CastFor(t, g, opp, "Their Instant", "Instant", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirLife-2 || me.Life != myLife+2 {
		t.Errorf("noncreature spell: opponent %d → %d (want -2), me %d → %d (want +2)", theirLife, opp.Life, myLife, me.Life)
	}
	myLife, theirLife = me.Life, opp.Life
	bear := game.Card{InstanceID: uuid.New(), Name: "Their Bear", TypeLine: "Creature — Bear", Keywords: []string{"flash"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID}
	opp.Hand.PushTop(bear)
	if err := g.CastSpell(opp.ID, bear.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the flash creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirLife || me.Life != myLife {
		t.Errorf("creature spell drained: opponent %d → %d, me %d → %d", theirLife, opp.Life, myLife, me.Life)
	}
}
