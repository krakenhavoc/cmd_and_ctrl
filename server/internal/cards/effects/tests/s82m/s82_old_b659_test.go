package testss82m

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b659_test.go — Invasion, batch 659 (#5616): life gain and loss.

const (
	oldB659Rewards      = "064c5a14-c2f2-4b7a-8e0b-01b9995e8f89"
	oldB659ThunderApp   = "34e325f0-b66f-44b9-a25c-da88357dbac4"
	oldB659Vapors       = "57104a79-cfb6-4330-8ea1-7726534ada56"
	oldB659Atalya       = "84587724-af96-4396-a0b3-1133d492d915"
	oldB659ContestID    = "b0a6bf69-7e21-4d5c-ac72-5946d84fc2ca"
	oldB659VileConsump  = "b6a005db-e7e2-4b88-bc8a-c257bcc31cf5"
	oldB659StormMaster  = "bfcd6c5d-7189-4d84-b0df-0a5ce5f43528"
	oldB659StormApp     = "cf2bacd3-b1ac-4863-b9c7-1e0a8d4854f3"
	oldB659Lens         = "d2ce3832-a12d-4c7a-9bc2-51dc92764994"
	oldB659ThunderMstr  = "e1db1798-17d1-4ef3-80d7-3f9115a01f7b"
	oldB659CreatureLine = "Creature — Human Wizard"
)

func TestOldB659EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB659Rewards: "Rewards of Diversity", oldB659ThunderApp: "Thunderscape Apprentice", oldB659Vapors: "Reviving Vapors",
		oldB659Atalya: "Atalya, Samite Master", oldB659ContestID: "Mages' Contest", oldB659VileConsump: "Vile Consumption",
		oldB659StormMaster: "Stormscape Master", oldB659StormApp: "Stormscape Apprentice", oldB659Lens: "Phyrexian Lens",
		oldB659ThunderMstr: "Thunderscape Master",
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

func oldB659Wizard(t *testing.T, g *game.Game, owner uuid.UUID, name, oracle string) uuid.UUID {
	t.Helper()
	id := effectstest.PushCatalogPermanent(g, owner, name, oldB659CreatureLine, oracle, false)
	return id
}

func TestOldB659RewardsOfDiversityGainsFourForAnOpponentsMulticolouredSpell(t *testing.T) {
	for _, tc := range []struct {
		name   string
		colors []string
		opp    bool
		want   int
	}{
		{"opponent's multicoloured", []string{"R", "G"}, true, 4},
		{"opponent's mono", []string{"R"}, true, 0},
		{"my own multicoloured", []string{"R", "G"}, false, 0},
	} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me, opp := oldB659Seats(g)
		effectstest.PushCatalogPermanent(g, me.ID, "Rewards of Diversity", "Enchantment", oldB659Rewards, false)
		caster := me
		if tc.opp {
			caster = opp
		}
		id := uuid.New()
		caster.Hand.PushTop(game.Card{InstanceID: id, Name: "Spell", TypeLine: "Instant", Colors: tc.colors, Owner: caster.ID, Controller: caster.ID})
		before := me.Life
		if err := g.CastSpell(caster.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("%s: cast: %v", tc.name, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if me.Life != before+tc.want {
			t.Errorf("%s: life %d→%d, want %+d", tc.name, before, me.Life, tc.want)
		}
	}
}

func TestOldB659PhyrexianLensPaysOneLifeForOneManaOfAnyColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lens := effectstest.PushCatalogPermanent(g, me.ID, "Phyrexian Lens", "Artifact", oldB659Lens, false)
	before := me.Life
	if err := g.ActivateManaAbility(me.ID, lens, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"U"}) {
		t.Errorf("pool %v, want [U]", got)
	}
	if me.Life != before-1 {
		t.Errorf("life %d→%d, want -1", before, me.Life)
	}
	if err := g.ActivateManaAbility(me.ID, lens, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err == nil {
		t.Error("a tapped Lens made mana again")
	}
}

func TestOldB659ApprenticesAndMastersUseTheirColourAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := oldB659Seats(g)
	thunderApp := oldB659Wizard(t, g, me.ID, "Thunderscape Apprentice", oldB659ThunderApp)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	// {B},{T}: target player loses 1 life.
	effectstest.FloatMana(t, g, me, "{B}{G}")
	before := opp.Life
	if err := oldB659Activate(g, me.ID, thunderApp, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("drain: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("opponent life %d→%d, want -1", before, opp.Life)
	}
	if err := oldB659Activate(g, me.ID, thunderApp, 1, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Error("the tapped Apprentice used its second ability")
	}
	effectstest.Untap(g, thunderApp)
	if err := oldB659Activate(g, me.ID, thunderApp, 1, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("pump: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("Bear is %d/%d, want 3/3", p, tough)
	}

	// Thunderscape Master: {B}{B},{T}: drain 2; {G}{G},{T}: creatures you control +2/+2.
	g = effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp = oldB659Seats(g)
	master := oldB659Wizard(t, g, me.ID, "Thunderscape Master", oldB659ThunderMstr)
	bear = effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{B}{B}")
	myLife, theirLife := me.Life, opp.Life
	if err := oldB659Activate(g, me.ID, master, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("master drain: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirLife-2 || me.Life != myLife+2 {
		t.Errorf("life me %d→%d, opponent %d→%d, want +2 and -2", myLife, me.Life, theirLife, opp.Life)
	}
	effectstest.Untap(g, master)
	effectstest.FloatMana(t, g, me, "{G}{G}")
	if err := oldB659Activate(g, me.ID, master, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("master pump: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 4 {
		t.Errorf("my Bear has power %d, want 4", p)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, theirs); p != 2 {
		t.Errorf("their Bear has power %d, want 2 (only my creatures)", p)
	}

	// Stormscape Apprentice: {W},{T}: tap target creature; {B},{T}: lose 1 life.
	g = effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp = oldB659Seats(g)
	stormApp := oldB659Wizard(t, g, me.ID, "Stormscape Apprentice", oldB659StormApp)
	theirs = effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{W}{B}")
	if err := oldB659Activate(g, me.ID, stormApp, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(theirs)}); err != nil {
		t.Fatalf("tap: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B13Tapped(t, g, theirs) {
		t.Error("their Bear was not tapped")
	}
	effectstest.Untap(g, stormApp)
	before = opp.Life
	if err := oldB659Activate(g, me.ID, stormApp, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("drain: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("opponent life %d→%d, want -1", before, opp.Life)
	}

	// Stormscape Master: {W}{W},{T}: protection from the colour of your choice; {B}{B},{T}: drain 2.
	g = effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp = oldB659Seats(g)
	stormMaster := oldB659Wizard(t, g, me.ID, "Stormscape Master", oldB659StormMaster)
	bear = effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{W}{W}{B}{B}")
	if err := oldB659Activate(g, me.ID, stormMaster, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("protection: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "R")
	if c, _ := effectstest.BattlefieldCard(g, bear); !game.HasKeyword(&c, game.ProtectionFromColor("R")) {
		t.Error("the Bear did not gain protection from red")
	}
	effectstest.Untap(g, stormMaster)
	myLife, theirLife = me.Life, opp.Life
	if err := oldB659Activate(g, me.ID, stormMaster, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("drain: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirLife-2 || me.Life != myLife+2 {
		t.Errorf("life me %d→%d, opponent %d→%d, want +2 and -2", myLife, me.Life, theirLife, opp.Life)
	}
}

func TestOldB659AtalyaSpendsOnlyWhiteOnXToPreventDamageOrGainLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := oldB659Seats(g)
	atalya := effectstest.PushCatalogPermanent(g, me.ID, "Atalya, Samite Master", "Legendary Creature — Human Cleric", oldB659Atalya, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{R}{R}")
	if err := oldB659Activate(g, me.ID, atalya, 0, game.ActivateAbilityParams{XValue: 2, Modes: []int{1}}); err == nil {
		t.Fatal("X was paid with red mana")
	}
	effectstest.FloatMana(t, g, me, "{W}{W}")
	before := me.Life
	if err := oldB659Activate(g, me.ID, atalya, 0, game.ActivateAbilityParams{XValue: 2, Modes: []int{1}}); err != nil {
		t.Fatalf("gain life: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+2 {
		t.Errorf("life %d→%d, want +2", before, me.Life)
	}
	effectstest.Untap(g, atalya)
	effectstest.FloatMana(t, g, me, "{W}{W}")
	if err := oldB659Activate(g, me.ID, atalya, 0, game.ActivateAbilityParams{XValue: 2, Modes: []int{0}, Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("prevent: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DealDamageTo(t, g, uuid.New(), bear, 2)
	if !effectstest.OnBattlefieldByID(g, bear) {
		t.Error("the Bear died though 2 damage was to be prevented")
	}
}

func TestOldB659VileConsumptionAsksEachCreaturesControllerToPayOneLifeOrSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, opp.ID, "Vile Consumption", "Enchantment", oldB659VileConsump, false)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	// The game starts on seat 0's turn, so the opponent's upkeep comes first: only
	// their creature asks, and declining sacrifices it.
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("the opponent was not asked in their upkeep")
	}
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("I was asked in the opponent's upkeep")
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, theirs) {
		t.Error("their Bear was not sacrificed after declining")
	}
	// Then mine: paying 1 life keeps it.
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("I was not asked to pay 1 life in my upkeep")
	}
	before := me.Life
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before-1 || !effectstest.OnBattlefieldByID(g, mine) {
		t.Errorf("paying: life %d→%d, My Bear on battlefield %v; want -1 and kept", before, me.Life, effectstest.OnBattlefieldByID(g, mine))
	}
}

func TestOldB659RevivingVaporsTakesOneGainsItsManaValueAndBinsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "Filler")
	two := effectstest.SeedLibraryTop(me, "Two", "Creature — Bear", "{1}{G}")
	five := effectstest.SeedLibraryTop(me, "Five", "Creature — Dragon", "{3}{R}{R}")
	zero := effectstest.SeedLibraryTop(me, "Zero", "Land", "")
	effectstest.CastCatalogSpell(t, g, "Reviving Vapors", "Instant", oldB659Vapors, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var c *game.PendingChoice
	for _, pc := range g.PendingChoices {
		if pc != nil && pc.Kind == game.PendingChoiceRevealPick && pc.Chooser == me.ID {
			c = pc
		}
	}
	if c == nil || len(c.ChooseCards) != 3 {
		t.Fatalf("want a pick among the three revealed cards, got %+v", c)
	}
	before := me.Life
	if err := g.ResolveRevealPick(c.ID, me.ID, []uuid.UUID{five}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	if !me.Hand.Contains(five) {
		t.Error("the picked card is not in hand")
	}
	if me.Life != before+5 {
		t.Errorf("life %d→%d, want +5 (the picked card's mana value)", before, me.Life)
	}
	if !me.Graveyard.Contains(two) || !me.Graveyard.Contains(zero) {
		t.Error("the other revealed cards did not go to the graveyard")
	}
}

func oldB659Contest(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := oldB659Seats(g)
	spell := effectstest.Batch01OpponentCasts(t, g, opp, "Their Spell", "", "", nil)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Mages' Contest", TypeLine: "Instant", OracleID: oldB659ContestID, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardTarget(spell)}); err != nil {
		t.Fatalf("cast Mages' Contest: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, opp, spell
}

func TestOldB659MagesContestCountersTheSpellWhenTheCasterWinsTheBidding(t *testing.T) {
	g, me, opp, spell := oldB659Contest(t)
	before := me.Life
	oldB585Answer(t, g, opp.ID, 0) // the opponent lets my opening bid of 1 stand
	// Every other seat is not a bidder: nothing more is asked.
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before-1 {
		t.Errorf("my life %d→%d, want -1 (the winning bid)", before, me.Life)
	}
	if !oldB659Countered(g, spell) {
		t.Error("the targeted spell was not countered")
	}
}

func TestOldB659MagesContestDoesNotCounterWhenTheSpellsControllerWins(t *testing.T) {
	g, me, opp, spell := oldB659Contest(t)
	myLife, theirLife := me.Life, opp.Life
	oldB585Answer(t, g, opp.ID, 1) // top it
	oldB585Bid(t, g, opp.ID, 2)
	oldB585Answer(t, g, me.ID, 0) // let it stand
	if opp.Life != theirLife-2 || me.Life != myLife {
		t.Errorf("life me %d→%d, opponent %d→%d, want only the opponent to lose 2", myLife, me.Life, theirLife, opp.Life)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, spell) == game.ZoneStack {
		t.Error("the spell was still on the stack after the bidding")
	}
	if oldB659Countered(g, spell) {
		t.Error("the spell was countered though its controller won the bidding")
	}
}

func oldB659Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func oldB659Activate(g *game.Game, who, src uuid.UUID, idx int, p game.ActivateAbilityParams) error {
	p.Strict = true
	return g.ActivateCatalogAbility(who, src, idx, p)
}

// oldB659Countered reports whether the log shows `spell` countered.
func oldB659Countered(g *game.Game, spell uuid.UUID) bool {
	for _, ev := range g.Events {
		if ev.Kind == game.EventCounterSpell && (ev.CardID == spell || ev.StackItemID == spell) {
			return true
		}
	}
	return false
}
