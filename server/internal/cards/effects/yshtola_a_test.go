package effects

import (
	"testing"

	"github.com/google/uuid"

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
		ysCrushingOracle:  "Crushing Disappointment",
		ysUndermineOracle: "Undermine",
		ysIrenicusOracle:  "Irenicus's Vile Duplication",
		ysMindspliceOracl: "Mindsplice Apparatus",
		ysDelneyOracle:    "Delney, Streetwise Lookout",
		ysGrahaTiaOracle:  "G'raha Tia, Scion Reborn",
		kambalOracle:      "Kambal, Consul of Allocation",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s registered as %q", oracle, spec.Name)
		}
		want := CompletenessFull
		if oracle == ysGrahaTiaOracle {
			want = CompletenessCaveats
		}
		if spec.Completeness != want {
			t.Errorf("%s: Completeness = %v, want %v", name, spec.Completeness, want)
		}
	}
}

func TestCrushingDisappointmentEachPlayerLosesTwoThenYouDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		life[p.ID] = p.Life
	}
	castCatalogSpell(t, g, "Crushing Disappointment", "Instant", ysCrushingOracle, nil)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: victim, Name: "Doomed Spell", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	advanceToMain(t, g)
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
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	hero := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Legendary Bear", TypeLine: "Legendary Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	castCatalogSpell(t, g, "Irenicus's Vile Duplication", "Sorcery", ysIrenicusOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: hero}})
	passPriorityAroundTable(t, g)
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
	if !IsToken(tok) {
		t.Errorf("the copy is not a token: %q", tok.TypeLine)
	}
	if tok.IsLegendary() {
		t.Errorf("the copy is legendary: %q", tok.TypeLine)
	}
	if !keywordSliceContains(tok.Keywords, "flying") {
		t.Errorf("the copy has no flying: %v", tok.Keywords)
	}
	if !g.Battlefield.Contains(hero) {
		t.Error("the original left the battlefield")
	}
}

func TestMindspliceApparatusGrowsAndDiscountsInstantsAndSorceries(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	id := pushCatalogPermanent(g, me.ID, "Mindsplice Apparatus", "Artifact", ysMindspliceOracl, false)
	if got := priceInHand(t, g, me, "Divination", "Sorcery", "{2}{U}"); got != 3 {
		t.Errorf("no counters: %d, want 3", got)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, mindspliceOilCounter, 2) })
	if got := priceInHand(t, g, me, "Divination", "Sorcery", "{2}{U}"); got != 1 {
		t.Errorf("two oil counters, sorcery: %d, want 1", got)
	}
	if got := priceInHand(t, g, me, "Counterspell", "Instant", "{U}{U}"); got != 2 {
		t.Errorf("two oil counters, {U}{U}: %d, want 2 (coloured pips stand)", got)
	}
	if got := priceInHand(t, g, me, "Golem", "Artifact Creature — Golem", "{4}"); got != 4 {
		t.Errorf("creature spell: %d, want 4 (undiscounted)", got)
	}
	if got := priceInHand(t, g, them, "Divination", "Sorcery", "{2}{U}"); got != 3 {
		t.Errorf("opponent's sorcery: %d, want 3 (undiscounted)", got)
	}

	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id && c.Counters[mindspliceOilCounter] != 3 {
			t.Errorf("oil counters after your upkeep = %d, want 3", c.Counters[mindspliceOilCounter])
		}
	}
}

func TestDelneysCreaturesCantBeBlockedByPowerThreeOrMore(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Delney, Streetwise Lookout", "Legendary Creature — Human Scout", ysDelneyOracle, 2, 2)
	small := b12Creature(g, me.ID, "Small Attacker", "Creature — Bear", 2, 2)
	large := b12Creature(g, me.ID, "Large Attacker", "Creature — Giant", 3, 3)
	bigBlocker := b12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	smallBlocker := b12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	brAttack(t, g, small, large)

	offered := brOffered(t, g, opp.ID)
	if brOffers(offered, bigBlocker, small) || !brOffers(offered, smallBlocker, small) || !brOffers(offered, bigBlocker, large) {
		t.Errorf("the enumerator disagrees with the rule: %v", offered)
	}
	brRefusal(t, g.DeclareBlocker(bigBlocker, small), game.BlockReasonCantBeBlockedBy)
	if err := g.DeclareBlocker(bigBlocker, large); err != nil {
		t.Fatalf("a power-3 attacker is outside the rule: %v", err)
	}
	if err := g.DeclareBlocker(smallBlocker, small); err != nil {
		t.Fatalf("a power-2 blocker may block: %v", err)
	}
}

func TestDelneyDoublesTriggersOfYourSmallCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Delney, Streetwise Lookout", "Legendary Creature — Human Scout", ysDelneyOracle, 2, 2)
	pushCatalogPermanent(g, me.ID, "Young Pyromancer", "Creature — Human Shaman", b07YoungPyromancerOracle, false)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if got := gl1bPyromancerTokens(g, me.ID); got != 2 {
		t.Errorf("Elementals from a power-1 Pyromancer = %d, want 2", got)
	}

	big := newCatalogGame(t)
	me, opp = big.Seats[0], big.Seats[1]
	b12Push(big, me.ID, "Delney, Streetwise Lookout", "Legendary Creature — Human Scout", ysDelneyOracle, 2, 2)
	pushBattlefieldCardWithTimestamp(big, game.Card{
		InstanceID: uuid.New(), Name: "Young Pyromancer", TypeLine: "Creature — Human Shaman",
		OracleID: b07YoungPyromancerOracle, Power: 3, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	castCatalogSpell(t, big, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, big)
	if got := gl1bPyromancerTokens(big, me.ID); got != 1 {
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
	passPriorityAroundTable(t, g)
}

func ysGrahaPromptOpen(g *game.Game) bool {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt {
			return true
		}
	}
	return false
}

func TestGrahaTiaPaysXLifeForAHeroOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "G'raha Tia, Scion Reborn", "Legendary Creature — Cat Wizard", ysGrahaTiaOracle, false)
	advanceToMain(t, g)
	life := me.Life

	ysCastNoncreature(t, g, me, "{2}{U}")
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Life != life-3 {
		t.Errorf("life %d → %d, want -3 for a mana value 3 spell", life, me.Life)
	}
	heroes := gl1bNamed(g, me.ID, "Hero")
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "G'raha Tia, Scion Reborn", "Legendary Creature — Cat Wizard", ysGrahaTiaOracle, false)
	advanceToMain(t, g)

	creature := game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(creature)
	if err := g.CastSpell(me.ID, creature.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if ysGrahaPromptOpen(g) {
		t.Fatal("a creature spell triggered Throw Wide the Gates")
	}

	ysCastNoncreature(t, g, me, "{1}{U}")
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if n := len(gl1bNamed(g, me.ID, "Hero")); n != 0 {
		t.Fatalf("declined, yet %d Hero tokens", n)
	}
	life := me.Life
	ysCastNoncreature(t, g, me, "{1}{U}")
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Life != life-2 || len(gl1bNamed(g, me.ID, "Hero")) != 1 {
		t.Errorf("after a decline the next spell should still pay out: life %d → %d, heroes %d", life, me.Life, len(gl1bNamed(g, me.ID, "Hero")))
	}
}

func TestKambalDrainsAnOpponentsNoncreatureSpellButNotACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Kambal, Consul of Allocation", "Legendary Creature — Human Advisor", kambalOracle, false)
	myLife, theirLife := me.Life, opp.Life
	castFor(t, g, opp, "Their Instant", "Instant", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if opp.Life != theirLife-2 || me.Life != myLife+2 {
		t.Errorf("noncreature spell: opponent %d → %d (want -2), me %d → %d (want +2)", theirLife, opp.Life, myLife, me.Life)
	}
	myLife, theirLife = me.Life, opp.Life
	bear := game.Card{InstanceID: uuid.New(), Name: "Their Bear", TypeLine: "Creature — Bear", Keywords: []string{"flash"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID}
	opp.Hand.PushTop(bear)
	if err := g.CastSpell(opp.ID, bear.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the flash creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != theirLife || me.Life != myLife {
		t.Errorf("creature spell drained: opponent %d → %d, me %d → %d", theirLife, opp.Life, myLife, me.Life)
	}
}
