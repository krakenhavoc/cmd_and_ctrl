package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b294_cards_test.go — Duel Decks: Blessed vs. Cursed cards, S82 OLD
// batch 294, played through the real catalog entries.

const (
	oldB294ElderCathar   = "ba9ac0a9-735d-4bca-9d24-586aca963872"
	oldB294Champion      = "faeaa3eb-6eb8-4621-98c7-0c6f58bbff85"
	oldB294TandemLookout = "1576f031-57ca-4894-b461-8ea57b6cefdc"
	oldB294AbattoirGhoul = "2b211adb-de44-4468-b65d-907a09aa7e9d"
	oldB294Pitchfork     = "409b44fc-fae2-4804-83a7-78cf95038f58"
	oldB294BondsOfFaith  = "79af045e-2c73-4003-8f62-b8611a225f08"
	oldB294Captain       = "a7b6cd8b-1bed-4742-8c97-8be4edbfc8ee"
	oldB294HumanFrailty  = "bd95c832-9ae9-404a-a2bc-0760a5b9ff56"
)

func TestOldB294CardsAreFullWithNoCaveats(t *testing.T) {
	for id, name := range map[string]string{
		oldB294ElderCathar: "Elder Cathar", oldB294Champion: "Champion of the Parish", oldB294TandemLookout: "Tandem Lookout",
		oldB294AbattoirGhoul: "Abattoir Ghoul", oldB294Pitchfork: "Sharpened Pitchfork", oldB294BondsOfFaith: "Bonds of Faith",
		oldB294Captain: "Captain of the Mists", oldB294HumanFrailty: "Human Frailty",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB294Counters(g *game.Game, id uuid.UUID) int {
	c, _ := g.LookupCardForEffect(id)
	return c.Counters[game.CounterPlusOne]
}

func oldB294PickTarget(t *testing.T, g *game.Game, me uuid.UUID, ref game.TargetRef) {
	t.Helper()
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.LatestPickTarget(g, me); pick != nil {
		if err := g.ResolvePickTarget(pick.ID, me, ref); err != nil {
			t.Fatalf("pick target: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
}

func TestOldB294ElderCatharPutsOneOrTwoCounters(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		want     int
	}{{"Creature — Bear", 1}, {"Creature — Human Soldier", 2}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		cathar := effectstest.B12Push(g, me.ID, "Elder Cathar", "Creature — Human Soldier", oldB294ElderCathar, 2, 2)
		friend := effectstest.B12Creature(g, me.ID, "Friend", tc.typeLine, 2, 2)
		effectstest.KillCreature(t, g, me.ID, cathar)
		oldB294PickTarget(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: friend})
		if got := oldB294Counters(g, friend); got != tc.want {
			t.Errorf("%s got %d counters, want %d", tc.typeLine, got, tc.want)
		}
	}
}

func TestOldB294ChampionAndCaptainWatchOtherHumansOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	champ := effectstest.B12Push(g, me.ID, "Champion of the Parish", "Creature — Human Soldier", oldB294Champion, 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB294Counters(g, champ); got != 0 {
		t.Fatalf("a Bear entering gave %d counters", got)
	}
	effectstest.CastCatalogSpell(t, g, "Villager", "Creature — Human Peasant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB294Counters(g, champ); got != 1 {
		t.Errorf("a Human entering gave %d counters, want 1", got)
	}
}

func TestOldB294CaptainUntapsForAHumanAndTapsOrUntapsAPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cap := effectstest.B12Push(g, me.ID, "Captain of the Mists", "Creature — Human Wizard", oldB294Captain, 2, 3)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == cap {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Villager", "Creature — Human Peasant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := g.LookupCardForEffect(cap)
	if c.Tapped {
		t.Error("the Captain stayed tapped after another Human entered")
	}
	spec, _ := effects.Lookup(oldB294Captain)
	if len(spec.Activated) != 1 || spec.Activated[0].Targets == nil {
		t.Errorf("the Captain's activated ability is missing or targetless")
	}
}

func TestOldB294TandemLookoutDrawsForEachPairedCreatureThatDamagesAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Wolf", "Creature — Wolf", 3, 3)
	lookout := effectstest.SbondPairCard(t, g, me, "Tandem Lookout", "Creature — Human Scout", oldB294TandemLookout, 2, 1, bear)
	for _, tc := range []struct {
		name string
		src  uuid.UUID
		want int
	}{{"the Lookout", lookout, 1}, {"its partner", bear, 1}, {"an unpaired creature", other, 0}} {
		before := me.Hand.Size()
		effectstest.DealNoncombatDamage(g, me.ID, tc.src, opp.ID)
		effectstest.B21SettleStack(t, g)
		if got := me.Hand.Size() - before; got != tc.want {
			t.Errorf("%s drew %d, want %d", tc.name, got, tc.want)
		}
	}
	before := me.Hand.Size()
	effectstest.DealNoncombatDamage(g, me.ID, bear, me.ID)
	effectstest.B21SettleStack(t, g)
	if got := me.Hand.Size() - before; got != 0 {
		t.Errorf("damage to yourself drew %d", got)
	}
}

func TestOldB294AbattoirGhoulGainsTheToughnessOfCreaturesItDamaged(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ghoul := effectstest.B12Push(g, me.ID, "Abattoir Ghoul", "Creature — Zombie", oldB294AbattoirGhoul, 3, 2)
	hit := effectstest.B12Creature(g, opp.ID, "Hit Bear", "Creature — Bear", 2, 5)
	missed := effectstest.B12Creature(g, opp.ID, "Missed Bear", "Creature — Bear", 2, 4)
	effectstest.DealNoncombatDamage(g, me.ID, ghoul, hit)
	before := me.Life
	effectstest.KillCreature(t, g, opp.ID, missed)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before {
		t.Fatalf("life %d after a creature the Ghoul never damaged died, want %d", me.Life, before)
	}
	effectstest.KillCreature(t, g, opp.ID, hit)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+5 {
		t.Errorf("life %d, want %d (the damaged creature's toughness)", me.Life, before+5)
	}
}

func oldB294Attach(g *game.Game, owner, host uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: owner, Controller: owner, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host},
	})
}

func TestOldB294SharpenedPitchforkGivesFirstStrikeAndPumpsHumans(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	human := effectstest.B12Creature(g, me.ID, "Human", "Creature — Human Soldier", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	oldB294Attach(g, me.ID, human, "Sharpened Pitchfork", "Artifact — Equipment", oldB294Pitchfork)
	oldB294Attach(g, me.ID, bear, "Sharpened Pitchfork", "Artifact — Equipment", oldB294Pitchfork)
	if p := effectstest.PowerOf(t, g, human); p != 3 {
		t.Errorf("Human power %d, want 3", p)
	}
	if p := effectstest.PowerOf(t, g, bear); p != 2 {
		t.Errorf("Bear power %d, want 2", p)
	}
	for _, id := range []uuid.UUID{human, bear} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "first strike") {
			t.Errorf("%s lacks first strike", id)
		}
	}
}

func TestOldB294BondsOfFaithPumpsAHumanAndGroundsAnythingElse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	human := effectstest.B12Creature(g, me.ID, "Human", "Creature — Human Soldier", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	oldB294Attach(g, me.ID, human, "Bonds of Faith", "Enchantment — Aura", oldB294BondsOfFaith)
	oldB294Attach(g, me.ID, bear, "Bonds of Faith", "Enchantment — Aura", oldB294BondsOfFaith)
	if p := effectstest.PowerOf(t, g, human); p != 4 {
		t.Errorf("Human power %d, want 4", p)
	}
	if p := effectstest.PowerOf(t, g, bear); p != 2 {
		t.Errorf("Bear power %d, want 2 (no bonus)", p)
	}
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if err := effectstest.DeclareResult(g, bear, opp.ID); err == nil {
		t.Error("the enchanted Bear attacked")
	}
	if err := effectstest.DeclareResult(g, human, opp.ID); err != nil {
		t.Errorf("the enchanted Human could not attack: %v", err)
	}
}

func TestOldB294HumanFrailtyDestroysOnlyHumans(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	human := effectstest.B12Creature(g, opp.ID, "Human", "Creature — Human Soldier", 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Human Frailty", "Instant", oldB294HumanFrailty, effectstest.CardRefs(bear)); err == nil {
		t.Error("targeted a non-Human creature")
	}
	effectstest.CastCatalogSpell(t, g, "Human Frailty", "Instant", oldB294HumanFrailty, effectstest.CardRefs(human))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(human) {
		t.Error("the Human survived")
	}
}
