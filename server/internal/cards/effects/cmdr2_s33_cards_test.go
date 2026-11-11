package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	s33BackInTownOracle    = "ce723db6-2a80-42a5-a176-cd322b4e1e27"
	s33WeRideAtDawnOracle  = "5675ac63-b69b-41b9-a7ed-4085a4e431b2"
	s33SellSwordOracle     = "60c77b18-c111-4e16-ad19-2ef32582cd9e"
	s33GraverobberOracle   = "960cff3b-678c-404e-a643-ac5e91e6b50c"
	s33TellerOracle        = "bd1ccc6f-d281-413e-889e-1e022c3f2a99"
	s33ApothecaryOracle    = "09ecb8df-2a17-4e5a-814c-474bce54794d"
	s33StockpileOracle     = "feaf9dcb-d845-4cf9-b35d-cd58175162e7"
	s33RetreatOracle       = "0158634a-08a3-4eb8-8216-bf85161fd168"
	s33LifeInsuranceOracle = "50779f73-c538-4364-aca4-4ab069a00814"
)

func TestS33CardsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s33BackInTownOracle:    "Back in Town",
		s33LifeInsuranceOracle: "Life Insurance",
		s33RetreatOracle:       "Discreet Retreat",
		s33StockpileOracle:     "Glittering Stockpile",
		s33ApothecaryOracle:    "Bloodroot Apothecary",
		s33TellerOracle:        "Misfortune Teller",
		s33GraverobberOracle:   "Charred Graverobber",
		s33SellSwordOracle:     "Angelic Sell-Sword",
		s33WeRideAtDawnOracle:  "We Ride at Dawn",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestS33LifeInsuranceTreasureForEachNontokenCreatureDeath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Life Insurance", "Enchantment", s33LifeInsuranceOracle, 0, 0)
	tok := socB2Token(g, me.ID, "Goblin", "Goblin", 1, 1)
	b27Kill(g, tok)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Treasure"); n != 0 {
		t.Fatalf("%d Treasures from a token dying, want 0", n)
	}
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := me.Life
	b27Kill(g, theirs)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("%d Treasures after an opponent's creature died, want 1", n)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b27Kill(g, mine)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Treasure"); n != 2 {
		t.Errorf("%d Treasures after two deaths, want 2", n)
	}
	_ = game.ZoneBattlefield
}

func TestS33BackInTownReturnsXOutlawCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pirate := pushGraveyardPermanent(me, "Pirate", "Creature — Human Pirate", "{1}{U}")
	rogue := pushGraveyardPermanent(me, "Rogue", "Creature — Human Rogue", "{1}{B}")
	bear := pushGraveyardPermanent(me, "Bear", "Creature — Bear", "{1}{G}")
	rock := pushGraveyardPermanent(me, "Rock", "Artifact — Mercenary", "{1}")

	castXSpell(t, g, "Back in Town", "Sorcery", s33BackInTownOracle, "{X}{2}{B}", 2, cardRefs(pirate, rogue))
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(pirate) || !g.Battlefield.Contains(rogue) {
		t.Error("a targeted outlaw creature card did not return")
	}
	if !me.Graveyard.Contains(bear) || !me.Graveyard.Contains(rock) {
		t.Error("an untargeted card left the graveyard")
	}
}

func TestS33BackInTownTargetsExactlyXOutlawCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pirate := pushGraveyardPermanent(me, "Pirate", "Creature — Human Pirate", "{1}{U}")
	bear := pushGraveyardPermanent(me, "Bear", "Creature — Bear", "{1}{G}")
	rock := pushGraveyardPermanent(me, "Rock", "Artifact — Mercenary", "{1}")
	advanceToMain(t, g)
	cast := func(x int, targets ...uuid.UUID) error {
		floatMana(t, g, me, "{B}{C}{C}"+string(slices.Repeat([]byte("{C}"), x)))
		_, err := castFromHandErr(g, me, "Back in Town", "Sorcery", "{X}{2}{B}", s33BackInTownOracle, game.CastSpellParams{Strict: true, XValue: x, Targets: cardRefs(targets...)})
		g.WithWriteLock(func() { me.ManaPool = game.ManaPool{} })
		return err
	}
	if err := cast(1, bear); err == nil {
		t.Error("a non-outlaw creature card was a legal target")
	}
	if err := cast(1, rock); err == nil {
		t.Error("a noncreature outlaw card was a legal target")
	}
	if err := cast(2, pirate); err == nil {
		t.Error("one target was accepted for X = 2")
	}
	if err := cast(1, pirate); err != nil {
		t.Fatalf("X = 1 with an outlaw creature card was refused: %v", err)
	}
}

func TestS33WeRideAtDawnGivesLegendaryCreatureSpellsConvokeOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	b12Push(g, me.ID, "We Ride at Dawn", "Enchantment", s33WeRideAtDawnOracle, 0, 0)
	elves := cgTwoGreenCreatures(g, me.ID)

	plain := cgBear(me.ID)
	g.WithWriteLock(func() { me.Hand.PushTop(plain) })
	if err := g.CastSpell(me.ID, plain.InstanceID, game.CastSpellParams{Strict: true, TapIDs: elves}); err == nil {
		t.Fatal("a nonlegendary creature spell was convoked")
	}
	legend := cgBear(me.ID)
	legend.Name, legend.TypeLine = "Legend", "Legendary Creature — Elf"
	g.WithWriteLock(func() { me.Hand.PushTop(legend) })
	if err := g.CastSpell(me.ID, legend.InstanceID, game.CastSpellParams{Strict: true, TapIDs: elves}); err != nil {
		t.Fatalf("convoking a legendary creature spell: %v", err)
	}
	if !cgTapped(g, elves) {
		t.Error("the convoking creatures were not tapped")
	}
}

func TestS33WeRideAtDawnMakesAMercenaryWhenYourCommanderAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Push(g, me.ID, "We Ride at Dawn", "Enchantment", s33WeRideAtDawnOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	cmdr := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Boss", TypeLine: "Legendary Creature — Human",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, IsCommander: true,
	})
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Mercenary"); n != 0 {
		t.Fatalf("%d Mercenaries after a non-commander attacked, want 0", n)
	}
	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	advanceToMain(t, g2)
	b12Push(g2, me2.ID, "We Ride at Dawn", "Enchantment", s33WeRideAtDawnOracle, 0, 0)
	cmdr2 := pushBattlefieldCardWithTimestamp(g2, game.Card{
		InstanceID: uuid.New(), Name: "Boss", TypeLine: "Legendary Creature — Human",
		Power: 2, Toughness: 2, Owner: me2.ID, Controller: me2.ID, IsCommander: true,
	})
	declareAttack(t, g2, opp2.ID, cmdr2)
	passPriorityAroundTable(t, g2)
	if n, _ := countTokensNamed(g2, me2.ID, "Mercenary"); n != 1 {
		t.Errorf("%d Mercenaries after the commander attacked, want 1", n)
	}
	_ = cmdr
}

func TestS33AngelicSellSwordMakesAMercenaryForItselfAndOtherNontokenCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12PlayFromHand(t, g, "Angelic Sell-Sword", "Creature — Angel Mercenary", s33SellSwordOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Mercenary"); n != 1 {
		t.Fatalf("%d Mercenaries after the Sell-Sword entered, want 1", n)
	}
	b12PlayFromHand(t, g, "Bear", "Creature — Bear", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Mercenary"); n != 2 {
		t.Fatalf("%d Mercenaries after another nontoken creature entered, want 2", n)
	}
	// The Mercenary tokens themselves entered without making more.
}

func TestS33AngelicSellSwordDrawsOnlyAtPowerSixOrMore(t *testing.T) {
	for _, tc := range []struct {
		power, draws int
	}{{4, 0}, {6, 1}} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		advanceToMain(t, g)
		sword := b12Push(g, me.ID, "Angelic Sell-Sword", "Creature — Angel Mercenary", s33SellSwordOracle, tc.power, 4)
		hand := me.Hand.Size()
		declareAttack(t, g, opp.ID, sword)
		passPriorityAroundTable(t, g)
		if got := me.Hand.Size() - hand; got != tc.draws {
			t.Errorf("power %d: drew %d, want %d", tc.power, got, tc.draws)
		}
	}
}

func TestS33CharredGraverobberDeclaresEscapeWithACounter(t *testing.T) {
	if !game.CardCastableFromZone(s33GraverobberOracle, game.ZoneGraveyard) {
		t.Fatal("the graveyard is not a castable zone")
	}
	offers := game.AlternativeCostsOfferedFromZone(s33GraverobberOracle, game.ZoneGraveyard)
	if len(offers) != 1 || offers[0].Key != "escape" {
		t.Fatalf("graveyard offers = %+v, want one escape", offers)
	}
	o := offers[0]
	if o.ManaCost != "{3}{B}{B}" || o.ExileFromGraveyard == nil || o.ExileFromGraveyard.Min != 4 || o.ExileFromGraveyard.Max != 4 {
		t.Errorf("escape offer = %+v, want {3}{B}{B} and exile four", o)
	}
	if o.EntersWithCounterCount != 1 || o.EntersWithCounterName != "+1/+1" {
		t.Errorf("escapes with %d %q counters, want one +1/+1", o.EntersWithCounterCount, o.EntersWithCounterName)
	}
}

func TestS33CharredGraverobberReturnsAnOutlawCardFromYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pirate := seedGraveyardCard(t, g, "Dead Pirate", "Creature — Human Pirate", "")
	bear := seedGraveyardCard(t, g, "Dead Bear", "Creature — Bear", "")
	castCatalogSpell(t, g, "Charred Graverobber", "Creature — Skeleton Mercenary", s33GraverobberOracle, nil)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: pirate}, true)
	if !me.Hand.Contains(pirate) {
		t.Error("the outlaw card did not return to hand")
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("a non-outlaw card left the graveyard")
	}
}

func TestS33MisfortuneTellerExilesAndPaysByCardType(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine          string
		rogues, treasures, gain int
	}{
		{"Dead Bear", "Creature — Bear", 1, 0, 0},
		{"Dead Forest", "Basic Land — Forest", 0, 1, 0},
		{"Dead Dryad", "Creature Land — Dryad", 1, 1, 0},
		{"Dead Rock", "Artifact", 0, 0, 3},
	} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		card := seedGraveyardCard(t, g, tc.name, tc.typeLine, "")
		life := me.Life
		castCatalogSpell(t, g, "Misfortune Teller", "Creature — Human Warlock", s33TellerOracle, nil)
		assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: card}, true)
		if me.Graveyard.Contains(card) {
			t.Errorf("%s: the card was not exiled", tc.name)
		}
		if n, _ := countTokensNamed(g, me.ID, "Rogue"); n != tc.rogues {
			t.Errorf("%s: %d Rogues, want %d", tc.name, n, tc.rogues)
		}
		if n, _ := countTokensNamed(g, me.ID, "Treasure"); n != tc.treasures {
			t.Errorf("%s: %d Treasures, want %d", tc.name, n, tc.treasures)
		}
		if me.Life != life+tc.gain {
			t.Errorf("%s: life %d, want %d", tc.name, me.Life, life+tc.gain)
		}
	}
}

func TestS33MisfortuneTellerTriggersOnCombatDamageToAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	teller := b12Push(g, me.ID, "Misfortune Teller", "Creature — Human Warlock", s33TellerOracle, 3, 1)
	card := seedGraveyardCard(t, g, "Dead Rock", "Artifact", "")
	life := me.Life
	dealNoncombatDamage(g, me.ID, teller, opp.ID)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(card) {
		t.Fatal("noncombat damage triggered it")
	}
	dealCombatDamage(g, me.ID, teller, opp.ID)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: card}, true)
	if me.Graveyard.Contains(card) || me.Life != life+3 {
		t.Errorf("combat damage did not trigger: card in graveyard %v, life %d want %d", me.Graveyard.Contains(card), me.Life, life+3)
	}
}

func TestS33BloodrootApothecaryGivesYouAndTargetOpponentATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castCatalogSpell(t, g, "Bloodroot Apothecary", "Creature — Squirrel Druid", s33ApothecaryOracle, nil)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}, true)
	if n, _ := countTokensNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("you have %d Treasures, want 1", n)
	}
	if n, _ := countTokensNamed(g, opp.ID, "Treasure"); n != 1 {
		t.Errorf("the opponent has %d Treasures, want 1", n)
	}
	if n, _ := countTokensNamed(g, g.Seats[2].ID, "Treasure"); n != 0 {
		t.Errorf("a bystander has %d Treasures, want 0", n)
	}
}

func TestS33BloodrootApothecaryPoisonsAnOpponentSacrificingANoncreatureToken(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Bloodroot Apothecary", "Creature — Squirrel Druid", s33ApothecaryOracle, 3, 3)
	sac := func(owner uuid.UUID, id uuid.UUID) {
		if err := (SacrificePermanent{Target: id}).Apply(ctxFor(g, &game.StackItem{Controller: owner})); err != nil {
			t.Fatalf("sacrifice: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	poison := func(p *game.Player) int { return p.Counters[game.CounterPoison] }
	sac(opp.ID, socB2Token(g, opp.ID, "Goblin", "Goblin", 1, 1))
	if poison(opp) != 0 {
		t.Fatalf("a creature token gave %d poison", poison(opp))
	}
	sac(opp.ID, b12Permanent(g, opp.ID, "Rock", "Artifact"))
	if poison(opp) != 0 {
		t.Fatalf("a nontoken permanent gave %d poison", poison(opp))
	}
	mint := func(owner uuid.UUID) uuid.UUID {
		if err := (CreateToken{Controller: owner, Template: TreasureToken(), N: 1}).Apply(ctxFor(g, &game.StackItem{Controller: owner})); err != nil {
			t.Fatalf("create: %v", err)
		}
		return findBattlefieldByName(g, "Treasure")
	}
	sac(me.ID, mint(me.ID))
	if poison(me) != 0 {
		t.Fatalf("your own Treasure gave you %d poison", poison(me))
	}
	sac(opp.ID, mint(opp.ID))
	if poison(opp) != 2 {
		t.Errorf("an opponent's sacrificed Treasure gave %d poison, want 2", poison(opp))
	}
}

func TestS33GlitteringStockpileBanksStashCountersAndCashesThemIn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	stockpile := pushCatalogPermanent(g, me.ID, "Glittering Stockpile", "Artifact — Treasure", s33StockpileOracle, false)
	untap := func() { g.WithWriteLock(func() { _ = g.UntapTargetForEffect(stockpile) }) }
	stash := func() int {
		c, _ := g.LookupCardForEffect(stockpile)
		return c.Counters["stash"]
	}
	for i := 1; i <= 3; i++ {
		if err := g.ActivateManaAbility(me.ID, stockpile, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("tap for {R} #%d: %v", i, err)
		}
		if stash() != i {
			t.Fatalf("after %d taps: %d stash counters", i, stash())
		}
		untap()
	}
	if me.ManaPool.CountColor("R") != 3 || me.ManaPool.Total() != 3 {
		t.Fatalf("pool = %v, want {R}{R}{R}", me.ManaPool)
	}
	g.WithWriteLock(func() { me.ManaPool = game.ManaPool{} })
	if err := g.ActivateManaAbility(me.ID, stockpile, 1, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("sacrifice for mana: %v", err)
	}
	if got := me.ManaPool.CountColor("G"); got != 3 || me.ManaPool.Total() != 3 {
		t.Errorf("sacrificing with 3 stash counters added %d mana, want 3", got)
	}
	if onBattlefield(g, stockpile) {
		t.Error("the artifact was not sacrificed")
	}
}

func s33RetreatTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	land := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Plain Land", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	pushAuraOnLand(t, g, me.ID, "Discreet Retreat", s33RetreatOracle, land)
	return g, me, land
}

func TestS33DiscreetRetreatMakesTwoManaOfOneColorForOutlawsOnly(t *testing.T) {
	g, me, land := s33RetreatTable(t)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("tap the enchanted land: %v", err)
	}
	if me.ManaPool.CountColor("G") != 2 || me.ManaPool.Total() != 2 {
		t.Fatalf("pool = %v, want {G}{G}", me.ManaPool)
	}
	bear := cgBear(me.ID)
	pirate := cgBear(me.ID)
	pirate.Name, pirate.TypeLine = "Pirate", "Creature — Human Pirate"
	g.WithWriteLock(func() { me.Hand.PushTop(bear); me.Hand.PushTop(pirate) })
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("the restricted mana paid for a non-outlaw spell")
	}
	if err := g.CastSpell(me.ID, pirate.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the restricted mana did not pay for an outlaw spell: %v", err)
	}
}

func TestS33DiscreetRetreatDrawsOnTheFirstOutlawSpellEachTurnOnly(t *testing.T) {
	g, me, _ := s33RetreatTable(t)
	outlaw := func(name, subtype string) game.Card {
		c := cgBear(me.ID)
		c.Name, c.TypeLine, c.ManaCost = name, "Creature — Human "+subtype, ""
		return c
	}
	plain := cgBear(me.ID)
	plain.ManaCost = ""
	rogue, warlock := outlaw("Rogue", "Rogue"), outlaw("Warlock", "Warlock")
	g.WithWriteLock(func() { me.Hand.PushTop(plain); me.Hand.PushTop(rogue); me.Hand.PushTop(warlock) })
	cast := func(c game.Card) {
		t.Helper()
		if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("cast %s: %v", c.Name, err)
		}
		passPriorityAroundTable(t, g)
	}
	hand, life := me.Hand.Size(), me.Life
	cast(plain)
	if me.Hand.Size() != hand-1 || me.Life != life {
		t.Fatalf("a non-outlaw spell triggered it: hand %d→%d, life %d→%d", hand, me.Hand.Size(), life, me.Life)
	}
	cast(rogue)
	if me.Hand.Size() != hand-2+1 || me.Life != life-1 {
		t.Fatalf("the first outlaw spell: hand %d→%d, life %d→%d, want a draw and 1 life", hand, me.Hand.Size(), life, me.Life)
	}
	cast(warlock)
	if me.Hand.Size() != hand-3+1 || me.Life != life-1 {
		t.Errorf("the second outlaw spell triggered it: hand %d, life %d", me.Hand.Size(), me.Life)
	}
}
