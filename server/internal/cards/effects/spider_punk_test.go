package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spider_punk_test.go — Spider-Punk (#2014), every line of it through
// the real catalog. The ability half of the counter gate is pinned with
// stubs in game/cant_be_countered_test.go.

const spiderPunkOracle = "03bf7a90-78c1-4e83-8553-e22d85e49876"

func pushSpiderPunk(g *game.Game, controller uuid.UUID) uuid.UUID {
	return pushCatalogPermanent(g, controller, "Spider-Punk", "Legendary Creature — Spider Human Hero", spiderPunkOracle, false)
}

// The headline: a Stifle aimed at an opponent's triggered ability is
// still a legal target, resolves and does nothing, and the ability
// resolves after it. The static has no "you", so it covers the
// opponent's ability as well as its controller's.
func TestSpiderPunkStopsStifle(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	pushSpiderPunk(g, me.ID)
	src := pushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	fired := false
	ability := taAnnounce(t, g, opp.ID, src, "Their Rock — do a thing", nil, &fired)

	if !g.AbilityCantBeCounteredForEffect(ability) {
		t.Error("the stack chip should say the ability can't be countered")
	}
	if !taLegalTargets(g, me.ID, stifleOracle)[ability] {
		t.Fatal("an ability that can't be countered is still a legal target (CR 101.2)")
	}
	stifle := castCatalogSpell(t, g, "Stifle", "Instant", stifleOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: ability}})
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(stifle) {
		t.Error("Stifle resolves and does nothing: it should be in the graveyard")
	}
	passPriorityAroundTable(t, g)
	if !fired {
		t.Error("Spider-Punk's ability shield did not stop Stifle")
	}
}

// "Spells and abilities can't be countered": every player's spells.
func TestSpiderPunkShieldsEveryPlayersSpells(t *testing.T) {
	for _, s := range []csSpell{{card: csBlueInstant}, {card: csBear, caster: 1, controller: 1}} {
		g := newCatalogGame(t)
		pushSpiderPunk(g, g.Seats[0].ID)
		id := csPushSpell(g, s)
		if !counterByEffect(t, g, id) {
			t.Errorf("%s (seat %d's) was countered under Spider-Punk", s.card.Name, s.controller)
		}
	}
}

// "Other Spiders you control have riot": a Spider of mine is asked as
// it enters, a non-Spider and an opponent's Spider are not, and
// Spider-Punk itself is asked once, for its printed riot only.
func TestSpiderPunkGrantsRiotToOtherSpiders(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]

	punk := castCatalogSpell(t, g, "Spider-Punk", "Legendary Creature — Spider Human Hero", spiderPunkOracle, nil)
	passPriorityAroundTable(t, g)
	answerRiotFor(t, g, true)
	if !g.Battlefield.Contains(punk) {
		t.Fatal("Spider-Punk was asked for riot more than once: its grant is for OTHER Spiders")
	}
	passPriorityAroundTable(t, g)

	spider := castCatalogSpell(t, g, "Giant Spider", "Creature — Spider", "", nil)
	passPriorityAroundTable(t, g)
	answerRiotFor(t, g, false)
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, spider, "haste") {
		t.Error("my Spider's riot, haste answer: no haste")
	}

	bear := castCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if pendingOfKind(g, game.PendingChoiceEntryRiot) != nil {
		t.Fatal("a non-Spider was asked for riot")
	}
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the bear did not enter")
	}

	theirs := riotPush(g, opp.ID, "Their Spider", "Creature — Spider", "", 2, 4, 0)
	if hasKeywordOnBattlefield(t, g, theirs, game.KeywordRiot) {
		t.Error("an opponent's Spider has riot from my Spider-Punk")
	}
	if !hasKeywordOnBattlefield(t, g, spider, game.KeywordRiot) || hasKeywordOnBattlefield(t, g, bear, game.KeywordRiot) {
		t.Errorf("riot on the battlefield: Spider %v, Bear %v; want true, false",
			hasKeywordOnBattlefield(t, g, spider, game.KeywordRiot), hasKeywordOnBattlefield(t, g, bear, game.KeywordRiot))
	}
	if me.ID == opp.ID {
		t.Fatal("setup: two seats")
	}
}

// The riot grant names a creature type, which a kindred permanent
// shares (CR 205.3m): a noncreature Spider of mine has riot too.
func TestSpiderPunkGrantsRiotToAKindredSpider(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushSpiderPunk(g, me.ID)
	core := riotPush(g, me.ID, "Kindred Web", "Kindred Artifact — Spider", "", 0, 0, 0)
	if !hasKeywordOnBattlefield(t, g, core, game.KeywordRiot) {
		t.Error("a kindred Spider I control does not have riot")
	}
}

// "Damage can't be prevented": protection stops nothing.
func TestSpiderPunkDamageCantBePrevented(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushSpiderPunk(g, me.ID)
	knight := pr6ProRed(g, opp.ID, 9)
	pinger := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Red Pinger", TypeLine: "Creature — Goblin",
		Power: 1, Toughness: 1, Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(pinger, knight, 2)
	})
	if got := damageMarkedOn(g, knight); got != 2 {
		t.Errorf("knight has %d damage, want 2 through protection", got)
	}
}
