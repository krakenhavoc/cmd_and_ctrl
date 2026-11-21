package testsso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b15_cards_test.go — Secrets of Strixhaven Commander, fifteenth batch
// (#4834). Each test plays the real catalog entry through the engine and
// asserts what the printed text says.

const (
	socB15VanguardOracle   = "5564c782-be40-49fc-abc0-6cdcbaaa575a"
	socB15WarsingerOracle  = "e286d1aa-bb74-4568-98b5-a693ac7a369d"
	socB15SalvoOracle      = "9c2fd3a3-96f2-4e84-85a9-77698c2f83aa"
	socB15TorrentOracle    = "ec2968fb-163d-4907-862b-b84a1f167833"
	socB15WaveOracle       = "24f7143c-d43a-4120-baa0-0235b15d8186"
	socB15PhantomOracle    = "6edd2a36-3d65-4bf9-bd82-b6d03a7e8fea"
	socB15CampusOracle     = "6cd58a88-6434-4c55-bf93-a739b5ed9bc1"
	socB15CommandOracle    = "7cc81fb3-495c-47cd-8c56-96197d87e9e0"
	socB15ZimoneOracle     = "6c044dd1-8938-4d62-a7b2-e2298de08f70"
	socB15CreatureTypeLine = "Creature — Bear"
)

func TestSocB15EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB15VanguardOracle:  "Vanguard of the Restless",
		socB15WarsingerOracle: "Venerable Warsinger",
		socB15SalvoOracle:     "Volcanic Salvo",
		socB15TorrentOracle:   "Volcanic Torrent",
		socB15WaveOracle:      "Wave of Reckoning",
		socB15PhantomOracle:   "White Orchid Phantom",
		socB15CampusOracle:    "Witherbloom Campus",
		socB15CommandOracle:   "Witherbloom Command",
		socB15ZimoneOracle:    "Zimone, All-Questioning",
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

// --- Vanguard of the Restless ---------------------------------------------

func TestSocB15VanguardPumpsYourSpiritsPerCommanderCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vanguard := effectstest.B12Push(g, me.ID, "Vanguard of the Restless", "Creature — Spirit Knight", socB15VanguardOracle, 2, 2)
	spirit := effectstest.B12Creature(g, me.ID, "Friendly Spirit", "Creature — Spirit", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", socB15CreatureTypeLine, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Spirit", "Creature — Spirit", 1, 1)
	if got := effectstest.EffectivePower(t, g, vanguard); got != 2 {
		t.Fatalf("Vanguard power %d before any commander cast, want 2", got)
	}

	// A real cast of a commander from the command zone bumps the tally.
	yuriko := effectstest.YurikoInCommandZone(me, 0)
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, yuriko, game.CastSpellParams{FromZone: "command"}); err != nil {
		t.Fatalf("cast the commander: %v", err)
	}
	for _, c := range []struct {
		name string
		id   uuid.UUID
		want int
	}{{"Vanguard", vanguard, 3}, {"your other Spirit", spirit, 2}, {"your Bear", bear, 2}, {"their Spirit", theirs, 1}} {
		if got := effectstest.EffectivePower(t, g, c.id); got != c.want {
			t.Errorf("%s power %d after one commander cast, want %d", c.name, got, c.want)
		}
	}
	if got := effectstest.EffectiveToughness(t, g, vanguard); got != 3 {
		t.Errorf("Vanguard toughness %d, want 3", got)
	}
}

func socB15VanguardInGraveyard(g *game.Game, owner *game.Player) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		owner.Graveyard.PushTop(game.Card{
			InstanceID: id, Name: "Vanguard of the Restless", TypeLine: "Creature — Spirit Knight",
			OracleID: socB15VanguardOracle, ManaCost: "{2}{W}", Power: 2, Toughness: 2,
			Owner: owner.ID, Controller: owner.ID,
		})
	})
	return id
}

func TestSocB15VanguardReturnsFromTheGraveyardWhenASpiritEntersForThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vanguard := socB15VanguardInGraveyard(g, me)
	effectstest.CastAndResolveCreature(t, g, "Test Spirit", "Creature — Spirit", "")
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		effectstest.DrcSettle(t, g, me.ID)
	}
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("a Spirit entering offered no payment")
	}
	effectstest.B06AddMana(me, "W", "W", "W")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	c, ok := effectstest.BattlefieldCard(g, vanguard)
	if !ok || c.Controller != me.ID {
		t.Fatal("Vanguard did not return to the battlefield")
	}
	if me.Graveyard.Contains(vanguard) {
		t.Error("Vanguard is still in the graveyard")
	}
}

func TestSocB15VanguardStaysWhenTheCostIsDeclinedOrTheCreatureIsNotASpirit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vanguard := socB15VanguardInGraveyard(g, me)

	effectstest.CastAndResolveCreature(t, g, "Plain Bear", socB15CreatureTypeLine, "")
	effectstest.DrcSettle(t, g, me.ID)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("a Bear woke Vanguard")
	}

	effectstest.B12Creature(g, opp.ID, "Their Spirit", "Creature — Spirit", 1, 1)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: g.Battlefield.Cards[len(g.Battlefield.Cards)-1].InstanceID})
	})
	effectstest.DrcSettle(t, g, me.ID)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("an opponent's Spirit woke Vanguard")
	}

	effectstest.CastAndResolveCreature(t, g, "Test Spirit", "Creature — Spirit", "")
	effectstest.DrcSettle(t, g, me.ID)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("no payment offered for my Spirit")
	}
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	if g.Battlefield.Contains(vanguard) || !me.Graveyard.Contains(vanguard) {
		t.Error("declined but Vanguard moved")
	}
}

// --- Venerable Warsinger --------------------------------------------------

func TestSocB15WarsingerReturnsACreatureCardUpToTheDamageDealt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	warsinger := effectstest.PushDiesCreatureForTest(g, me.ID, "Venerable Warsinger", socB15WarsingerOracle, "Creature — Spirit Cleric", 3, 3)
	one := effectstest.PushGraveyardPermanent(me, "One Drop", socB15CreatureTypeLine, "{G}")
	three := effectstest.PushGraveyardPermanent(me, "Three Drop", socB15CreatureTypeLine, "{1}{U}{U}")
	five := effectstest.PushGraveyardPermanent(me, "Five Drop", socB15CreatureTypeLine, "{3}{W}{W}")
	spell := effectstest.PushGraveyardPermanent(me, "Cheap Sorcery", "Sorcery", "{R}")
	theirs := effectstest.PushGraveyardPermanent(opp, "Their Two Drop", socB15CreatureTypeLine, "{1}{B}")
	effectstest.AdvanceToMain(t, g)

	effectstest.DealCombatDamageToPlayer(g, warsinger, opp.ID, 3)
	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no target prompt after the damage")
	}
	for _, c := range []struct {
		name string
		id   uuid.UUID
		in   bool
	}{{"mana value 1", one, true}, {"mana value 3", three, true}, {"mana value 5", five, false},
		{"a sorcery card", spell, false}, {"an opponent's creature card", theirs, false}} {
		if got := effectstest.HasID(prompt.PickTargetCards, c.id); got != c.in {
			t.Errorf("%s offered %v, want %v", c.name, got, c.in)
		}
	}
	effectstest.PickCard(t, g, me.ID, three)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(three) || me.Graveyard.Contains(three) {
		t.Error("the mana value 3 creature card did not come back")
	}
	if g.Battlefield.Contains(five) || g.Battlefield.Contains(one) {
		t.Error("another card came back")
	}
}

func TestSocB15WarsingerReadsXFromEachDamageEvent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	warsinger := effectstest.PushDiesCreatureForTest(g, me.ID, "Venerable Warsinger", socB15WarsingerOracle, "Creature — Spirit Cleric", 3, 3)
	one := effectstest.PushGraveyardPermanent(me, "One Drop", socB15CreatureTypeLine, "{G}")
	five := effectstest.PushGraveyardPermanent(me, "Five Drop", socB15CreatureTypeLine, "{3}{W}{W}")
	effectstest.AdvanceToMain(t, g)

	effectstest.DealCombatDamageToPlayer(g, warsinger, opp.ID, 1)
	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil || !effectstest.HasID(prompt.PickTargetCards, one) || effectstest.HasID(prompt.PickTargetCards, five) {
		t.Fatalf("after 1 damage the offer is %v, want only the mana value 1 card", prompt)
	}
	effectstest.PickCard(t, g, me.ID, one)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(one) || !me.Graveyard.Contains(one) {
		t.Error("declined but the card moved")
	}

	effectstest.DealCombatDamageToPlayer(g, warsinger, opp.ID, 5)
	prompt = effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil || !effectstest.HasID(prompt.PickTargetCards, five) {
		t.Fatalf("after 5 damage the offer is %v, want the mana value 5 card too", prompt)
	}
}

func TestSocB15WarsingerAsksNothingWithNoCreatureCardInRange(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	warsinger := effectstest.PushDiesCreatureForTest(g, me.ID, "Venerable Warsinger", socB15WarsingerOracle, "Creature — Spirit Cleric", 3, 3)
	effectstest.PushGraveyardPermanent(me, "Five Drop", socB15CreatureTypeLine, "{3}{W}{W}")
	effectstest.AdvanceToMain(t, g)
	effectstest.DealCombatDamageToPlayer(g, warsinger, opp.ID, 2)
	if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.TriggersOnStackFrom(g, warsinger) != 0 {
		t.Error("the trigger went on the stack with no creature card of mana value 2 or less")
	}
}

// --- Volcanic Salvo -------------------------------------------------------

func TestSocB15SalvoDealsSixToEachOfUpToTwoTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 6, 6)
	bigger := effectstest.B12Creature(g, opp.ID, "Their Colossus", "Creature — Giant", 7, 7)
	walker := effectstest.CcrLoyaltyWalker(g, opp.ID, 5, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "RR")
	// Total power 10: the {10} is gone and {R}{R} is the whole cost.
	effectstest.B12Creature(g, me.ID, "My Giant", "Creature — Giant", 6, 6)
	effectstest.B12Creature(g, me.ID, "My Bear", socB15CreatureTypeLine, 4, 4)
	socB8Cast(t, g, game.Card{Name: "Volcanic Salvo", TypeLine: "Sorcery", ManaCost: "{10}{R}{R}", OracleID: socB15SalvoOracle},
		game.CastSpellParams{Strict: true, Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: big}, {Kind: game.TargetCard, ID: walker},
		}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("a 6/6 survived 6 damage")
	}
	if g.Battlefield.Contains(walker) {
		t.Error("a 5-loyalty planeswalker survived 6 damage")
	}
	if !g.Battlefield.Contains(bigger) || effectstest.DamageMarkedOn(g, bigger) != 0 {
		t.Error("an untargeted creature was damaged")
	}
}

func TestSocB15SalvoCostsLessByTotalPowerAndNeverBelowRR(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := effectstest.B12Creature(g, opp.ID, "Their Bear", socB15CreatureTypeLine, 2, 2)
	effectstest.B12Creature(g, me.ID, "My Bear", socB15CreatureTypeLine, 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "RR")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Volcanic Salvo", TypeLine: "Sorcery", ManaCost: "{10}{R}{R}",
		OracleID: socB15SalvoOracle, Owner: me.ID, Controller: me.ID})
	// Power 3 takes only {3} off: {7}{R}{R} is out of reach of two mana.
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(target)}); err == nil {
		t.Fatal("cast {10}{R}{R} for {R}{R} with a total power of 3")
	}
}

// --- Volcanic Torrent -----------------------------------------------------

func TestSocB15TorrentDealsDamageEqualToSpellsCastThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := effectstest.B12Creature(g, opp.ID, "Their Small", "Creature — Bear", 3, 3)
	large := effectstest.B12Creature(g, opp.ID, "Their Large", "Creature — Giant", 5, 5)
	walker := effectstest.CcrLoyaltyWalker(g, opp.ID, 3, 0)
	mine := effectstest.B12Creature(g, me.ID, "My Large", "Creature — Giant", 5, 5)
	effectstest.SeedCheapLibrary(g, me, 1)

	effectstest.CastWithCost(t, g, "Prior Spell", "Instant", "{U}", "")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastWithCost(t, g, "Volcanic Torrent", "Sorcery", "{4}{R}", socB15TorrentOracle)
	effectstest.PassPriorityAroundTable(t, g) // the cascade trigger resolves
	// Cascade finds the cheap instant and casts it: a third spell this turn.
	c := effectstest.KeywordCastPromptFor(g, me.ID)
	if c == nil {
		t.Fatal("no cascade offer")
	}
	effectstest.AnswerKeywordCast(t, g, c, true)
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		effectstest.PassPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(small) {
		t.Error("the 3/3 survived 3 damage")
	}
	if g.Battlefield.Contains(walker) {
		t.Error("the 3-loyalty planeswalker survived 3 damage")
	}
	if !g.Battlefield.Contains(large) || effectstest.DamageMarkedOn(g, large) != 3 {
		t.Errorf("the 5/5 has %d damage marked, want 3", effectstest.DamageMarkedOn(g, large))
	}
	if effectstest.DamageMarkedOn(g, mine) != 0 {
		t.Error("my own creature was damaged")
	}
}

// --- Wave of Reckoning ----------------------------------------------------

func TestSocB15WaveMakesEachCreatureDamageItselfByItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := effectstest.B12Creature(g, me.ID, "Small", socB15CreatureTypeLine, 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 6, 6)
	sturdy := effectstest.B12Creature(g, opp.ID, "Sturdy", "Creature — Giant", 4, 5)
	wall := effectstest.B12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 4)
	touch := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Toucher", TypeLine: "Creature — Snake", Power: 1, Toughness: 9,
		Keywords: []string{"deathtouch"}, Owner: me.ID, Controller: me.ID,
	})
	link := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Linker", TypeLine: "Creature — Cleric", Power: 3, Toughness: 4,
		Keywords: []string{"lifelink"}, Owner: opp.ID, Controller: opp.ID,
	})
	oppLife := opp.Life
	effectstest.CastCatalogSpell(t, g, "Wave of Reckoning", "Sorcery", socB15WaveOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range []struct {
		name string
		id   uuid.UUID
		dead bool
	}{{"the 2/2", small, true}, {"the 6/6", big, true}, {"the 4/5", sturdy, false},
		{"the 0/4", wall, false}, {"the deathtouch 1/9", touch, true}, {"the lifelink 3/4", link, false}} {
		if gone := !g.Battlefield.Contains(c.id); gone != c.dead {
			t.Errorf("%s dead=%v, want %v", c.name, gone, c.dead)
		}
	}
	if got := effectstest.DamageMarkedOn(g, sturdy); got != 4 {
		t.Errorf("the 4/5 has %d damage, want 4", got)
	}
	if got := effectstest.DamageMarkedOn(g, wall); got != 0 {
		t.Errorf("the 0/4 has %d damage, want 0", got)
	}
	if got := opp.Life - oppLife; got != 3 {
		t.Errorf("the lifelink creature's controller gained %d, want 3", got)
	}
}

// --- White Orchid Phantom -------------------------------------------------

func TestSocB15PhantomDestroysANonbasicLandAndItsControllerFetchesATappedBasic(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	coffers := effectstest.SeedLandOnBattlefield(g, opp.ID, "Cabal Coffers", "Land")
	swamp := effectstest.SeedLandOnBattlefield(g, opp.ID, "Swamp", "Basic Land — Swamp")
	basic := effectstest.B07LibraryLand(opp, "Swamp", "Basic Land — Swamp")
	effectstest.B07LibraryLand(opp, "Swamp", "Basic Land — Swamp")

	effectstest.CastAndResolveCreature(t, g, "White Orchid Phantom", "Creature — Spirit Knight", socB15PhantomOracle)
	prompt := effectstest.LatestPickTarget(g, g.Seats[0].ID)
	if prompt == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if !effectstest.HasID(prompt.PickTargetCards, coffers) || effectstest.HasID(prompt.PickTargetCards, swamp) {
		t.Fatalf("offered %v, want the nonbasic land and not the basic one", prompt.PickTargetCards)
	}
	effectstest.PickCard(t, g, g.Seats[0].ID, coffers)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(coffers) || !opp.Graveyard.Contains(coffers) {
		t.Error("the nonbasic land was not destroyed")
	}
	if effectstest.SearchChoiceFor(g, g.Seats[0].ID) != nil {
		t.Error("the Phantom's controller searched; the land's controller does")
	}
	effectstest.AnswerSearchByID(t, g, opp.ID, basic)
	card, ok := effectstest.BattlefieldCard(g, basic)
	if !ok {
		t.Fatal("the basic land did not reach the battlefield")
	}
	if !card.Tapped {
		t.Error("the basic land entered untapped; the card says tapped")
	}
}

func TestSocB15PhantomAsksNothingWithNoNonbasicLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLandOnBattlefield(g, opp.ID, "Swamp", "Basic Land — Swamp")
	effectstest.B07LibraryLand(opp, "Swamp", "Basic Land — Swamp")
	effectstest.CastAndResolveCreature(t, g, "White Orchid Phantom", "Creature — Spirit Knight", socB15PhantomOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.SearchChoiceFor(g, opp.ID) != nil {
		t.Error("a prompt opened with no nonbasic land to destroy")
	}
}

// --- Witherbloom Campus ---------------------------------------------------

func TestSocB15WitherbloomCampusEntersTappedMakesBOrGAndScries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "Top", "Next")
	id := effectstest.PlayLandFromHand(t, g, "Witherbloom Campus", socB15CampusOracle)
	effectstest.PassPriorityAroundTable(t, g)
	card, ok := effectstest.BattlefieldCard(g, id)
	if !ok || !card.Tapped {
		t.Fatalf("on battlefield %v, tapped %v; want it to enter tapped", ok, card.Tapped)
	}
	effectstest.UntapForTest(g, id)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err == nil {
		t.Fatal("the land made {R}")
	}
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("{T}: Add {G}: %v", err)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })

	effectstest.UntapForTest(g, id)
	effectstest.FloatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("scry activated with three mana")
	}
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{4},{T}: scry 1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("scry prompt %+v, want one card", c)
	}
	top := c.ScryCards[0]
	if err := g.ResolveScry(c.ID, me.ID, []uuid.UUID{top}, nil); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	if bottom, err := me.Library.Bottom(); err != nil || bottom.InstanceID != top {
		t.Error("the scried card is not on the bottom")
	}
}

// --- Witherbloom Command --------------------------------------------------

func socB15CastCommand(t *testing.T, g *game.Game, modes []int, targets []game.TargetRef) error {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Witherbloom Command", TypeLine: "Sorcery", ManaCost: "{B}{G}",
		OracleID: socB15CommandOracle, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	return g.CastSpell(me.ID, id, game.CastSpellParams{Modes: modes, Targets: targets})
}

func TestSocB15CommandMillsThenReturnsALandAndDrains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(opp, "A", "B", "C", "D")
	land := effectstest.PushGraveyardPermanent(me, "Old Forest", "Basic Land — Forest", "")
	nonland := effectstest.PushGraveyardPermanent(me, "Old Bear", socB15CreatureTypeLine, "{1}{G}")
	libBefore := opp.Library.Size()
	myLife, oppLife := me.Life, opp.Life

	if err := socB15CastCommand(t, g, []int{0, 3}, []game.TargetRef{
		effectstest.ModeRef(game.TargetPlayer, opp.ID, 0, 0),
		effectstest.ModeRef(game.TargetPlayer, opp.ID, 1, 0),
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := libBefore - opp.Library.Size(); got != 3 {
		t.Errorf("opponent milled %d, want 3", got)
	}
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no land to return was asked for")
	}
	if !effectstest.HasID(pick.ChooseCards, land) || effectstest.HasID(pick.ChooseCards, nonland) {
		t.Fatalf("offered %v, want only the land card", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, land)
	if !me.Hand.Contains(land) || me.Graveyard.Contains(land) {
		t.Error("the land card did not go to my hand")
	}
	if opp.Life != oppLife-2 || me.Life != myLife+2 {
		t.Errorf("life: opponent %d (was %d), me %d (was %d); want -2 and +2", opp.Life, oppLife, me.Life, myLife)
	}
}

func TestSocB15CommandCanReturnTheLandItJustMilled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil // exactly three cards, all lands
	owned := func(name string) uuid.UUID {
		return effectstest.PushLibraryCardForTest(me, game.Card{Name: name, TypeLine: "Basic Land — Plains", Owner: me.ID, Controller: me.ID})
	}
	milled := owned("Milled Plains")
	owned("Another Plains")
	owned("Third Plains")
	victim := g.Seats[1]
	if err := socB15CastCommand(t, g, []int{0, 3}, []game.TargetRef{
		effectstest.ModeRef(game.TargetPlayer, me.ID, 0, 0),
		effectstest.ModeRef(game.TargetPlayer, victim.ID, 1, 0),
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(milled) {
		t.Fatalf("the land was not milled")
	}
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil || !effectstest.HasID(pick.ChooseCards, milled) {
		t.Fatalf("the land the spell milled is not offered: %+v", pick)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, milled)
	if !me.Hand.Contains(milled) {
		t.Error("the milled land did not reach my hand")
	}
}

func TestSocB15CommandDestroysACheapNoncreatureNonlandPermanentOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.PushTypedCard(g, opp.ID, "Cheap Rock", "Artifact", "{1}")
	dear := effectstest.PushTypedCard(g, opp.ID, "Dear Rock", "Artifact", "{3}")
	creature := effectstest.PushTypedCard(g, opp.ID, "Cheap Creature", "Artifact Creature — Golem", "{1}")
	land := effectstest.SeedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island")
	for _, bad := range []uuid.UUID{dear, creature, land} {
		if err := socB15CastCommand(t, g, []int{1, 3}, []game.TargetRef{
			effectstest.ModeRef(game.TargetCard, bad, 0, 0),
			effectstest.ModeRef(game.TargetPlayer, opp.ID, 1, 0),
		}); err == nil {
			t.Errorf("the destroy mode accepted %v", bad)
		}
	}
	if err := socB15CastCommand(t, g, []int{1, 3}, []game.TargetRef{
		effectstest.ModeRef(game.TargetCard, rock, 0, 0),
		effectstest.ModeRef(game.TargetPlayer, opp.ID, 1, 0),
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	_ = me
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Graveyard.Contains(rock) {
		t.Error("the cheap artifact was not destroyed")
	}
}

func TestSocB15CommandShrinksACreatureUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	tough := effectstest.B12Creature(g, opp.ID, "Tough", "Creature — Giant", 5, 4)
	frail := effectstest.B12Creature(g, opp.ID, "Frail", socB15CreatureTypeLine, 2, 1)
	if err := socB15CastCommand(t, g, []int{2, 3}, []game.TargetRef{
		effectstest.ModeRef(game.TargetCard, tough, 0, 0),
		effectstest.ModeRef(game.TargetPlayer, opp.ID, 1, 0),
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.EffectivePower(t, g, tough), effectstest.EffectiveToughness(t, g, tough); p != 2 || tt != 3 {
		t.Errorf("the 5/4 is %d/%d, want 2/3", p, tt)
	}
	if err := socB15CastCommand(t, g, []int{2, 3}, []game.TargetRef{
		effectstest.ModeRef(game.TargetCard, frail, 0, 0),
		effectstest.ModeRef(game.TargetPlayer, opp.ID, 1, 0),
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(frail) {
		t.Error("a creature with toughness 1 survived -3/-1")
	}
}

// --- Zimone, All-Questioning ----------------------------------------------

func socB15Primos(g *game.Game, owner uuid.UUID) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == "Primo, the Indivisible" {
			out = append(out, c)
		}
	}
	return out
}

// socB15ZimoneBoard puts Zimone and `others` lands on the board, then plays
// one more land this turn: `others`+1 lands in all.
func socB15ZimoneBoard(t *testing.T, g *game.Game, others int) {
	t.Helper()
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Zimone, All-Questioning", "Legendary Creature — Human Wizard", socB15ZimoneOracle, false)
	for i := 0; i < others; i++ {
		effectstest.SeedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")
	}
	effectstest.PlayLandFromHand(t, g, "Forest", "")
}

func TestSocB15ZimoneMakesPrimoWithAsManyCountersAsLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	socB15ZimoneBoard(t, g, 2) // three lands: prime
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	primos := socB15Primos(g, me.ID)
	if len(primos) != 1 {
		t.Fatalf("Primos %d, want 1", len(primos))
	}
	p := primos[0]
	if !p.IsLegendary() || !p.IsToken() || p.Counters[game.CounterPlusOne] != 3 {
		t.Errorf("Primo %+v, want a legendary token with three +1/+1 counters", p)
	}
	if got := effectstest.PowerOf(t, g, p.InstanceID); got != 3 {
		t.Errorf("Primo power %d, want 3", got)
	}
}

func TestSocB15ZimoneDoesNothingOnANonPrimeLandCountOrWithoutALandEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	socB15ZimoneBoard(t, g, 3) // four lands: not prime
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(socB15Primos(g, me.ID)); n != 0 {
		t.Errorf("Primos %d with four lands, want 0", n)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.PushCatalogPermanent(g2, me2.ID, "Zimone, All-Questioning", "Legendary Creature — Human Wizard", socB15ZimoneOracle, false)
	for i := 0; i < 3; i++ { // three lands, none entered this turn
		effectstest.SeedLandOnBattlefield(g2, me2.ID, "Island", "Basic Land — Island")
	}
	effectstest.AdvanceTo(t, g2, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g2)
	if n := len(socB15Primos(g2, me2.ID)); n != 0 {
		t.Errorf("Primos %d with no land entering, want 0", n)
	}
}

func TestSocB15ZimoneRechecksTheLandCountOnResolution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	socB15ZimoneBoard(t, g, 4) // five lands: prime
	effectstest.AdvanceTo(t, g, game.StepEnd)
	if effectstest.StackFullyEmpty(g) {
		t.Fatal("the end step trigger did not go on the stack")
	}
	// A land leaves in response: four lands, not prime (CR 603.4).
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsLand() {
			effectstest.KillCreature(t, g, me.ID, c.InstanceID)
			break
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(socB15Primos(g, me.ID)); n != 0 {
		t.Errorf("Primos %d after the land count stopped being prime, want 0", n)
	}
}

func TestSocB15ZimoneCountsLandsAsItResolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	socB15ZimoneBoard(t, g, 2) // three lands
	effectstest.AdvanceTo(t, g, game.StepEnd)
	if effectstest.StackFullyEmpty(g) {
		t.Fatal("the end step trigger did not go on the stack")
	}
	effectstest.SeedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")
	effectstest.SeedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island") // five: still prime
	effectstest.PassPriorityAroundTable(t, g)
	primos := socB15Primos(g, me.ID)
	if len(primos) != 1 || primos[0].Counters[game.CounterPlusOne] != 5 {
		t.Fatalf("Primos %v, want one with five +1/+1 counters", primos)
	}
}

func TestSocB15IsPrime(t *testing.T) {
	prime := map[int]bool{2: true, 3: true, 5: true, 7: true, 11: true, 13: true, 17: true, 19: true, 23: true, 29: true, 31: true}
	for n := -1; n <= 32; n++ {
		if effects.SocB15IsPrime(n) != prime[n] {
			t.Errorf("socB15IsPrime(%d) = %v", n, effects.SocB15IsPrime(n))
		}
	}
}
