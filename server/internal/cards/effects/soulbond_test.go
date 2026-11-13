package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soulbond_test.go — the soulbond seam (#5675, CR 702.95), driven
// through Thundering Mightmare: the two triggers, the choice, what ends
// a pair, and a restore.

const (
	sbondMightmareOracle = "bf9a6242-b608-46ef-a059-7e757dd52e9b"
	sbondBoltOracle      = "4457ed35-7c10-48c8-9776-456485fdf070"
)

// sbondHand puts a catalog creature card in a player's hand.
func sbondHand(p *game.Player, name, typeLine, oracle string, power, toughness int, keywords ...string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Keywords: keywords, Owner: p.ID, Controller: p.ID,
	})
	return id
}

// sbondCast casts a creature from hand and resolves it, leaving whatever
// it triggered on the stack.
func sbondCast(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID) {
	t.Helper()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	nghcPassOnce(t, g)
}

// sbondMightmare casts Thundering Mightmare and resolves it.
func sbondMightmare(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	id := sbondHand(me, "Thundering Mightmare", "Creature — Horse Spirit", sbondMightmareOracle, 3, 3, game.KeywordSoulbond)
	sbondCast(t, g, me, id)
	return id
}

// sbondUntilAsked passes priority, answering any trigger-order prompt in
// the offered order, until a soulbond question is open.
func sbondUntilAsked(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	for i := 0; i < 32; i++ {
		if latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) != nil {
			return
		}
		if triggerOrderPrompt(g) != nil {
			answerTriggerOrderInOfferedOrder(t, g)
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("no soulbond question opened")
}

// sbondPartner reads who a permanent is paired with.
func sbondPartner(g *game.Game, id uuid.UUID) uuid.UUID {
	var out uuid.UUID
	g.WithWriteLock(func() { out = g.PairedWithForEffect(id) })
	return out
}

// sbondPair pairs a freshly cast Mightmare with `partner` through its own
// enters trigger and returns the Mightmare.
func sbondPair(t *testing.T, g *game.Game, me *game.Player, partner uuid.UUID) uuid.UUID {
	t.Helper()
	mare := sbondMightmare(t, g, me)
	nghcPassOnce(t, g)
	answerOwnPermanents(t, g, me.ID, partner)
	passPriorityAroundTable(t, g)
	if sbondPartner(g, mare) != partner || sbondPartner(g, partner) != mare {
		t.Fatal("the two creatures are not paired with each other")
	}
	return mare
}

// sbondGranted counts the granted Mightmare triggers a permanent has.
func sbondGranted(g *game.Game, id uuid.UUID) int {
	n := 0
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		for i := range g.Battlefield.Cards {
			if c := &g.Battlefield.Cards[i]; c.InstanceID == id {
				for _, gr := range c.Effective().GrantedAbilities {
					if gr.Key == game.GrantKey(thunderingMightmareGrant) {
						n++
					}
				}
			}
		}
	})
	return n
}

// sbondOpponentCasts has an opponent cast a Lightning Bolt at a third
// player and lets everything resolve.
func sbondOpponentCasts(t *testing.T, g *game.Game, opp *game.Player) {
	t.Helper()
	bolt := handCardForTest(opp, "Lightning Bolt", "Instant", sbondBoltOracle)
	if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[2].ID}},
	}); err != nil {
		t.Fatalf("opponent casts Lightning Bolt: %v", err)
	}
	b21SettleStack(t, g)
}

// CR 702.95a, first ability: as the soulbond creature enters, its
// controller may pair it with another unpaired creature they control,
// chosen as the ability resolves; both then have the shared ability.
func TestSoulbondPairsTheEnteringCreatureWithAChosenCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceTo(t, g, game.StepPrecombatMain)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	wolf := b12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	relic := b12Creature(g, me.ID, "Relic", "Artifact", 0, 0)

	mare := sbondMightmare(t, g, me)
	nghcPassOnce(t, g)
	p := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil || p.ChooseMin != 0 || p.ChooseMax != 1 {
		t.Fatalf("soulbond asks \"you may pair\" over one creature with a floor of zero: %+v", p)
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range p.ChooseCards {
		offered[id] = true
	}
	if !offered[bear] || !offered[wolf] || offered[mare] || offered[theirs] || offered[relic] {
		t.Fatalf("the candidates are your other unpaired creatures: %v", p.ChooseCards)
	}
	answerOwnPermanents(t, g, me.ID, wolf)
	passPriorityAroundTable(t, g)

	if sbondPartner(g, mare) != wolf || sbondPartner(g, wolf) != mare {
		t.Fatal("Thundering Mightmare and the Wolf are paired with each other")
	}
	if sbondPartner(g, bear) != uuid.Nil {
		t.Error("the Bear was not chosen and is unpaired")
	}
	if sbondGranted(g, mare) != 1 || sbondGranted(g, wolf) != 1 || sbondGranted(g, bear) != 0 {
		t.Errorf("each paired creature has the ability once: mare %d, wolf %d, bear %d",
			sbondGranted(g, mare), sbondGranted(g, wolf), sbondGranted(g, bear))
	}
}

// "You may": declining pairs nothing, and an unpaired Mightmare gives
// nobody its ability.
func TestSoulbondDecliningPairsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	mare := sbondMightmare(t, g, me)
	nghcPassOnce(t, g)
	answerOwnPermanents(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if sbondPartner(g, mare) != uuid.Nil || sbondPartner(g, bear) != uuid.Nil {
		t.Error("declining soulbond pairs nothing")
	}
	if sbondGranted(g, mare) != 0 || sbondGranted(g, bear) != 0 {
		t.Error("an unpaired Thundering Mightmare grants nothing")
	}
}

// CR 702.95a's intervening "if": with no other creature, the first
// ability does not trigger at all.
func TestSoulbondDoesNotTriggerWithNothingToPairWith(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	sbondMightmare(t, g, me)
	if !stackFullyEmpty(g) {
		t.Error("soulbond triggered with no other creature to pair with")
	}
}

// CR 702.95a, second ability: a creature that enters later may be
// paired with the unpaired soulbond creature, and the question names
// that creature alone.
func TestSoulbondPairsWithACreatureThatEntersLater(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	mare := sbondMightmare(t, g, me)
	older := enterFromHand(t, g, me.ID, "Older Bear", "Creature — Bear", "")
	// The Older Bear's entry asked too; decline it, so the Mightmare is
	// still unpaired when the next creature arrives.
	nghcPassOnce(t, g)
	answerOwnPermanents(t, g, me.ID)
	passPriorityAroundTable(t, g)

	bear := sbondHand(me, "Bear", "Creature — Bear", "", 2, 2)
	sbondCast(t, g, me, bear)
	nghcPassOnce(t, g)
	p := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil || len(p.ChooseCards) != 1 || p.ChooseCards[0] != bear || p.ChooseMin != 0 {
		t.Fatalf("the second ability offers the creature that entered, and may be declined: %+v", p)
	}
	answerOwnPermanents(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if sbondPartner(g, mare) != bear || sbondPartner(g, bear) != mare {
		t.Error("the creature that entered is paired with Thundering Mightmare")
	}
	if sbondPartner(g, older) != uuid.Nil {
		t.Error("the declined creature stays unpaired")
	}
}

// "Another creature you control": an opponent's creature entering does
// not trigger an unpaired soulbond creature.
func TestSoulbondIgnoresAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceTo(t, g, game.StepPrecombatMain)
	sbondMightmare(t, g, me)
	enterFromHand(t, g, opp.ID, "Their Wolf", "Creature — Wolf", "")
	enterFromHand(t, g, me.ID, "Relic", "Artifact", "")
	if !stackFullyEmpty(g) {
		t.Error("soulbond triggered for a creature its controller does not control, or for a noncreature")
	}
}

// CR 702.95d and the intervening "if" (CR 603.4): a paired soulbond
// creature does not trigger for the next creature, and an opponent's
// creature never triggers it.
func TestSoulbondAPairedCreatureIsNotOfferedAgain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceTo(t, g, game.StepPrecombatMain)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	mare := sbondPair(t, g, me, bear)

	enterFromHand(t, g, me.ID, "Wolf", "Creature — Wolf", "")
	enterFromHand(t, g, opp.ID, "Their Wolf", "Creature — Wolf", "")
	if !stackFullyEmpty(g) {
		t.Error("a paired soulbond creature triggered for another creature")
	}
	if sbondPartner(g, mare) != bear {
		t.Error("the pair changed")
	}
}

// CR 603.4 at resolution: two soulbond abilities race for one creature;
// the second to resolve finds it paired and asks nothing.
func TestSoulbondSecondAbilityDoesNothingOnceTheCreatureIsPaired(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	first := sbondMightmare(t, g, me)
	second := sbondHand(me, "Thundering Mightmare", "Creature — Horse Spirit", sbondMightmareOracle, 3, 3, game.KeywordSoulbond)
	sbondCast(t, g, me, second)
	// Two abilities: the newcomer's own, and the first Mightmare's
	// "another creature enters". Both name the same pair.
	sbondUntilAsked(t, g, me)
	answerOwnPermanents(t, g, me.ID, latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID).ChooseCards[0])
	b21SettleStack(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) != nil {
		t.Fatal("the second soulbond ability asked although both creatures were paired")
	}
	if sbondPartner(g, first) != second || sbondPartner(g, second) != first {
		t.Fatal("the two Mightmares are paired with each other")
	}
	// Each grants to both, so each has two instances (CR 113.2c).
	if sbondGranted(g, first) != 2 || sbondGranted(g, second) != 2 {
		t.Errorf("two paired Mightmares each have the ability twice: %d, %d", sbondGranted(g, first), sbondGranted(g, second))
	}
}

// The granted ability is each creature's own: an opponent's spell puts
// a counter on both, and your own spell on neither.
func TestThunderingMightmareBothCreaturesGrowWhenAnOpponentCasts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceTo(t, g, game.StepPrecombatMain)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
	mare := sbondPair(t, g, me, bear)

	sbondOpponentCasts(t, g, opp)
	if counterCount(g, mare, game.CounterPlusOne) != 1 || counterCount(g, bear, game.CounterPlusOne) != 1 {
		t.Errorf("each paired creature gets a +1/+1 counter: mare %d, bear %d",
			counterCount(g, mare, game.CounterPlusOne), counterCount(g, bear, game.CounterPlusOne))
	}
	if counterCount(g, other, game.CounterPlusOne) != 0 {
		t.Error("an unpaired creature got a counter")
	}
	sbondOpponentCasts(t, g, me)
	if counterCount(g, mare, game.CounterPlusOne) != 1 || counterCount(g, bear, game.CounterPlusOne) != 1 {
		t.Error("your own spell is not an opponent's")
	}
}

// CR 702.95e: the pair ends when either creature leaves the
// battlefield, and the one left behind loses the shared ability.
func TestSoulbondEndsWhenACreatureLeavesTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceTo(t, g, game.StepPrecombatMain)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	mare := sbondPair(t, g, me, bear)

	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, bear); err != nil {
		t.Fatalf("move the Bear: %v", err)
	}
	if sbondPartner(g, mare) != uuid.Nil || sbondGranted(g, mare) != 0 {
		t.Error("Thundering Mightmare is unpaired once its partner has left")
	}
	sbondOpponentCasts(t, g, opp)
	if counterCount(g, mare, game.CounterPlusOne) != 0 {
		t.Error("an unpaired Mightmare still had the ability")
	}
	// It is unpaired, so the next creature may be paired with it.
	wolf := enterFromHand(t, g, me.ID, "Wolf", "Creature — Wolf", "")
	nghcPassOnce(t, g)
	answerOwnPermanents(t, g, me.ID, wolf)
	passPriorityAroundTable(t, g)
	if sbondPartner(g, mare) != wolf {
		t.Error("an unpaired soulbond creature can be paired again")
	}
}

// CR 702.95e: another player gaining control of either creature ends
// the pair for good, even when control comes back.
func TestSoulbondEndsWhenAnotherPlayerGainsControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceTo(t, g, game.StepPrecombatMain)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	mare := sbondPair(t, g, me, bear)

	stealForTest(t, g, bear, opp.ID)
	if sbondPartner(g, mare) != uuid.Nil || sbondPartner(g, bear) != uuid.Nil {
		t.Fatal("a stolen creature and its old partner are both unpaired")
	}
	if sbondGranted(g, mare) != 0 || sbondGranted(g, bear) != 0 {
		t.Error("the shared ability went with the pair")
	}
	endTurn(t, g)
	if sbondPartner(g, mare) != uuid.Nil || sbondGranted(g, bear) != 0 {
		t.Error("the pair came back when control did")
	}
}

// CR 702.95e: a creature that stops being a creature is unpaired.
func TestSoulbondEndsWhenACreatureStopsBeingACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	mare := sbondPair(t, g, me, bear)

	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(bear),
			[]game.Mod{game.SetTypesMod([]string{"Artifact"}, nil)}, g.UntilEndOfTurnDuration(), "test — becomes an artifact")
	})
	if sbondPartner(g, mare) != uuid.Nil || sbondPartner(g, bear) != uuid.Nil {
		t.Fatal("a permanent that is no longer a creature is unpaired, and so is its partner")
	}
	endTurn(t, g)
	if sbondPartner(g, mare) != uuid.Nil {
		t.Error("the pair came back when the Bear was a creature again")
	}
}

// CR 702.95c: a creature that is gone, or no longer the controller's,
// as the ability resolves is not paired.
func TestSoulbondDoesNotPairACreatureStolenInResponse(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceTo(t, g, game.StepPrecombatMain)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	mare := sbondMightmare(t, g, me)
	// The ability is on the stack; the only candidate changes hands.
	stealForTest(t, g, bear, opp.ID)
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) != nil {
		t.Error("soulbond offered a creature its controller no longer controls")
	}
	if sbondPartner(g, mare) != uuid.Nil || sbondPartner(g, bear) != uuid.Nil {
		t.Error("nothing is paired")
	}
	var paired bool
	g.WithWriteLock(func() { paired = g.PairForEffect(me.ID, mare, bear) })
	if paired {
		t.Error("PairForEffect paired a creature another player controls")
	}
}

// A soulbond creature that has lost its abilities does not ask.
func TestSoulbondWithoutTheKeywordDoesNotTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	mare := sbondMightmare(t, g, me)
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(mare),
			[]game.Mod{game.LoseAllAbilitiesMod()}, game.IndefiniteDuration(), "test — loses all abilities")
	})
	enterFromHand(t, g, me.ID, "Bear", "Creature — Bear", "")
	if !stackFullyEmpty(g) {
		t.Error("a creature with no soulbond asked to pair")
	}
}

// The pair is carried by a restore point, with the question open and
// after it is answered.
func TestSoulbondSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	mare := sbondMightmare(t, g, me)
	nghcPassOnce(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) == nil {
		t.Fatal("no soulbond question is open")
	}

	// Restored while the question waits: the answer still pairs.
	g = restoreThroughJSON(t, g)
	me = g.Seats[0]
	answerOwnPermanents(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if sbondPartner(g, mare) != bear || sbondPartner(g, bear) != mare {
		t.Fatal("the restored question did not pair the creatures")
	}

	// Restored again with the pair made: it is still there and works.
	g = restoreThroughJSON(t, g)
	opp := g.Seats[1]
	if sbondPartner(g, mare) != bear || sbondPartner(g, bear) != mare {
		t.Fatal("the pair did not survive a restore")
	}
	if sbondGranted(g, mare) != 1 || sbondGranted(g, bear) != 1 {
		t.Error("the restored pair lost its shared ability")
	}
	sbondOpponentCasts(t, g, opp)
	if counterCount(g, mare, game.CounterPlusOne) != 1 || counterCount(g, bear, game.CounterPlusOne) != 1 {
		t.Error("the restored pair's ability did not trigger")
	}
}

// CR 702.26d: phasing out is not leaving the battlefield and changes no
// controller, so the pair is kept. While the partner is phased out it
// is treated as though it does not exist, and nothing is shared.
func TestSoulbondIsKeptThroughPhasing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	mare := sbondPair(t, g, me, bear)

	g.WithWriteLock(func() {
		if err := g.PhaseOutForEffect(uuid.Nil, bear); err != nil {
			t.Fatalf("PhaseOutForEffect: %v", err)
		}
	})
	if g.Battlefield.Contains(bear) {
		t.Fatal("the Bear did not phase out")
	}
	if sbondPartner(g, mare) != uuid.Nil || sbondGranted(g, mare) != 0 {
		t.Error("with its partner phased out, the Mightmare shares nothing")
	}
	// The Bear phases in during its controller's next untap step.
	for i := 0; i < len(g.Seats); i++ {
		endTurn(t, g)
	}
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the Bear did not phase back in")
	}
	if sbondPartner(g, mare) != bear || sbondPartner(g, bear) != mare {
		t.Error("the pair did not survive phasing")
	}
	if sbondGranted(g, mare) != 1 || sbondGranted(g, bear) != 1 {
		t.Error("the pair's shared ability is back")
	}
}
