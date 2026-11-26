package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b111_cards_test.go — Ikoria: Lair of Behemoths, batch 111 (#5512).
// Each mutate trigger is driven through the engine (ADR 0223), over or under.

type o111Card struct {
	oracle, name, typeLine, cost, mutate string
	power, toughness                     int
	keywords                             []string
}

var (
	o111Octopus    = o111Card{"0b718d21-6873-4613-ad15-660d30527fcd", "Sea-Dasher Octopus", "Creature — Octopus", "{1}{U}{U}", "{1}{U}", 2, 2, []string{"flash", "mutate"}}
	o111Shoreshark = o111Card{"38630f89-bc22-49b2-b97e-d31a5169decc", "Pouncing Shoreshark", "Creature — Shark Beast", "{4}{U}", "{3}{U}", 4, 3, []string{"flash", "mutate"}}
	o111Liger      = o111Card{"38f5e6bf-18a3-4ff3-9726-a25a9715a2d9", "Huntmaster Liger", "Creature — Cat", "{3}{W}", "{2}{W}", 3, 4, []string{"mutate"}}
	o111Drakkis    = o111Card{"3ec271d8-8264-4574-8c4c-e57b35872b1e", "Lore Drakkis", "Creature — Lizard Beast", "{1}{U}{R}", "{U/R}{U/R}", 2, 3, []string{"mutate"}}
	o111Hemophage  = o111Card{"46bfdcec-7cdc-413e-aa77-ed98b64f147b", "Insatiable Hemophage", "Creature — Nightmare", "{3}{B}", "{2}{B}", 3, 3, []string{"deathtouch", "mutate"}}
	o111DirgeBat   = o111Card{"4746828f-46ea-439f-88de-5e54cc536545", "Dirge Bat", "Creature — Bat", "{2}{B}{B}", "{4}{B}{B}", 3, 3, []string{"flash", "flying", "mutate"}}
	o111Greathorn  = o111Card{"4c6c8155-efa0-4c22-b255-34974a80c496", "Migratory Greathorn", "Creature — Beast", "{3}{G}", "{2}{G}", 3, 4, []string{"mutate"}}
	o111Gnarr      = o111Card{"734be5ec-6148-48b9-9021-380e9b62ec23", "Trumpeting Gnarr", "Creature — Beast", "{1}{G}{U}", "{3}{G/U}{G/U}", 3, 3, []string{"mutate"}}
	o111Vadrok     = o111Card{"775c41fa-2316-43c9-af34-ac9c9b0db7e2", "Vadrok, Apex of Thunder", "Legendary Creature — Elemental Dinosaur Cat", "{U}{R}{W}", "{1}{W/U}{R}{R}", 3, 3, []string{"flying", "first strike", "mutate"}}
	o111Necro      = o111Card{"7df82bf8-b232-48e1-858b-1571c48b6f46", "Necropanther", "Creature — Cat Nightmare", "{1}{W}{B}", "{2}{W/B}{W/B}", 3, 3, []string{"mutate"}}
	o111Starrix    = o111Card{"8432df19-dbf4-457b-8c0b-1f4b3a0a37d2", "Auspicious Starrix", "Creature — Elk Beast", "{4}{G}", "{5}{G}", 6, 6, []string{"mutate"}}
)

var o111All = []o111Card{o111Octopus, o111Shoreshark, o111Liger, o111Drakkis, o111Hemophage, o111DirgeBat, o111Greathorn, o111Gnarr, o111Vadrok, o111Necro, o111Starrix}

func TestS82OldB111EveryCardIsFullWithItsMutateOffer(t *testing.T) {
	for _, tc := range o111All {
		spec, ok := effects.Lookup(tc.oracle)
		if !ok || spec.Name != tc.name {
			t.Errorf("%s is not registered: %+v", tc.name, spec.Name)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 || len(spec.Triggered) == 0 {
			t.Errorf("%s: want Full, no caveat and a trigger; got %v %q %d", tc.name, spec.Completeness, spec.Caveats, len(spec.Triggered))
		}
		alt := game.AlternativeCostByKey(tc.oracle, game.MutateKey)
		if alt == nil || alt.ManaCost != tc.mutate || alt.Targets == nil || !alt.Targets.OwnedBySourceOwner {
			t.Errorf("%s: mutate offer %+v, want %s targeting a creature with the spell's owner", tc.name, alt, tc.mutate)
		}
	}
}

func (c o111Card) card() game.Card {
	return game.Card{Name: c.name, TypeLine: c.typeLine, OracleID: c.oracle, ManaCost: c.cost,
		Keywords: c.keywords, Power: c.power, Toughness: c.toughness}
}

// o111Mutate casts `c` for its mutate cost onto host and answers over (false)
// or under (true); the trigger is left for the caller to settle.
func o111Mutate(t *testing.T, g *game.Game, me *game.Player, c o111Card, host uuid.UUID, under bool) uuid.UUID {
	t.Helper()
	id := effectstest.RprMutatorInHand(t, g, me.ID, c.card())
	if err := effectstest.CastMutating(g, me.ID, id, host, false); err != nil {
		t.Fatalf("cast %s mutating: %v", c.name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	idx := 0
	if under {
		idx = 1
	}
	effectstest.AnswerOptionPick(t, g, me.ID, idx)
	return id
}

func o111Ref(id uuid.UUID) *game.TargetRef { return &game.TargetRef{Kind: game.TargetCard, ID: id} }

func o111Opp(g *game.Game) *game.Player { return g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)] }

func o111Perm(g *game.Game, owner uuid.UUID, name, typeLine string, p, tough int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Power: p, Toughness: tough, Owner: owner, Controller: owner})
}

func o111Grave(p *game.Player, name, typeLine, cost string, power int) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		Power: power, Toughness: 1, Owner: p.ID, Controller: p.ID})
	return id
}

func o111Zone(g *game.Game, id uuid.UUID) game.ZoneKind {
	if z := g.FindCardZoneForEffect(id); z != nil {
		return z.Kind
	}
	return ""
}

// A cast for the mana cost does not mutate, so nothing triggers.
func TestS82OldB111CastNormallyDoesNotTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.RprMutatorInHand(t, g, me.ID, o111Gnarr.card())
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Beast", me.ID); n != 0 {
		t.Errorf("%d Beast tokens from a cast that did not mutate", n)
	}
}

// --- Sea-Dasher Octopus -----------------------------------------------------------------

func TestS82OldB111SeaDasherOctopusDrawsWhenItDealsCombatDamageFromUnder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], o111Opp(g)
	elves := effectstest.LlanowarOn(g, me.ID)
	o111Mutate(t, g, me, o111Octopus, elves, true)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if !effectstest.HasKeywordOf(g, elves, "flash") {
		t.Error("the merged creature lacks the Octopus's flash")
	}
	before := me.Hand.Size()
	effectstest.DealCombatDamageToPlayer(g, elves, opp.ID, 1)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if me.Hand.Size() != before+1 {
		t.Errorf("hand %d -> %d, want one more card", before, me.Hand.Size())
	}
}

// --- Pouncing Shoreshark ----------------------------------------------------------------

func TestS82OldB111PouncingSharkBouncesAnOpponentsCreatureOnlyIfYouAgree(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], o111Opp(g)
		mine := o111Perm(g, me.ID, "Mine", "Creature — Bear", 2, 2)
		theirs := o111Perm(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
		elves := effectstest.LlanowarOn(g, me.ID)
		o111Mutate(t, g, me, o111Shoreshark, elves, false)
		effectstest.AssistCardSettle(t, g, me.ID, o111Ref(theirs), yes)
		inHand := o111Zone(g, theirs) == game.ZoneHand
		if inHand != yes {
			t.Errorf("answer %v: their creature in hand = %v", yes, inHand)
		}
		if o111Zone(g, mine) != game.ZoneBattlefield {
			t.Errorf("answer %v: my own creature moved", yes)
		}
	}
}

// --- Huntmaster Liger -------------------------------------------------------------------

func TestS82OldB111HuntmasterLigerPumpsTheOthersByTheTimesItHasMutated(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], o111Opp(g)
	bear := o111Perm(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := o111Perm(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	elves := effectstest.LlanowarOn(g, me.ID)
	o111Mutate(t, g, me, o111Liger, elves, false)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if p, _ := effectstest.PtOf(t, g, bear); p != 3 {
		t.Errorf("after one mutation the Bear is %d power, want 3", p)
	}
	if p, _ := effectstest.PtOf(t, g, elves); p != 3 {
		t.Errorf("the Liger itself is %d power, want its own 3", p)
	}
	if p, _ := effectstest.PtOf(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature is %d power, want 2", p)
	}
	// A second mutation of the same permanent: X is 2, on top of the first +1.
	o111Mutate(t, g, me, o111Liger, elves, false)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	// Two Ligers are now in the permanent and each one's ability triggers, each
	// reading X = 2: 2/2 + 1 (the first mutation) + 2 + 2.
	if p, tough := effectstest.PtOf(t, g, bear); p != 7 || tough != 7 {
		t.Errorf("after two mutations the Bear is %d/%d, want 7/7 (+1/+1 once, then +2/+2 from each of two Ligers)", p, tough)
	}
}

// --- Lore Drakkis -----------------------------------------------------------------------

func TestS82OldB111LoreDrakkisReturnsAnInstantOrSorceryOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bolt := o111Grave(me, "Bolt", "Instant", "{R}", 0)
	bear := o111Grave(me, "Dead Bear", "Creature — Bear", "{1}{G}", 2)
	elves := effectstest.LlanowarOn(g, me.ID)
	o111Mutate(t, g, me, o111Drakkis, elves, false)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil && effectstest.HasID(pick.PickTargetCards, bear) {
		t.Error("a creature card was offered")
	}
	effectstest.AssistCardSettle(t, g, me.ID, o111Ref(bolt), true)
	if o111Zone(g, bolt) != game.ZoneHand {
		t.Errorf("the instant is in %v, want hand", o111Zone(g, bolt))
	}
	if o111Zone(g, bear) != game.ZoneGraveyard {
		t.Errorf("the creature card is in %v, want the graveyard", o111Zone(g, bear))
	}
}

// --- Insatiable Hemophage ---------------------------------------------------------------

func TestS82OldB111InsatiableHemophageDrainsByTheTimesItHasMutated(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], o111Opp(g)
	elves := effectstest.LlanowarOn(g, me.ID)
	life, theirs := me.Life, opp.Life
	o111Mutate(t, g, me, o111Hemophage, elves, true)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if me.Life != life+1 || opp.Life != theirs-1 {
		t.Errorf("first mutation: life %d (want %d), opponent %d (want %d)", me.Life, life+1, opp.Life, theirs-1)
	}
	if !effectstest.HasKeywordOf(g, elves, "deathtouch") {
		t.Error("the merged creature lacks deathtouch")
	}
	o111Mutate(t, g, me, o111Hemophage, elves, false)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	// Two Hemophages are in the permanent now and each one's ability triggers
	// on the second merge, each reading X = 2: 1 + 2 + 2.
	if me.Life != life+5 || opp.Life != theirs-5 {
		t.Errorf("second mutation: life %d (want %d), opponent %d (want %d)", me.Life, life+5, opp.Life, theirs-5)
	}
}

// --- Dirge Bat --------------------------------------------------------------------------

func TestS82OldB111DirgeBatDestroysACreatureOrPlaneswalkerAnOpponentControls(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], o111Opp(g)
	mine := o111Perm(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: 4}, Owner: opp.ID, Controller: opp.ID})
	elves := effectstest.LlanowarOn(g, me.ID)
	o111Mutate(t, g, me, o111DirgeBat, elves, false)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if effectstest.HasID(pick.PickTargetCards, mine) || !effectstest.HasID(pick.PickTargetCards, walker) {
		t.Errorf("offered %v: want the opponent's planeswalker and not my own creature", pick.PickTargetCards)
	}
	effectstest.AssistCardSettle(t, g, me.ID, o111Ref(walker), true)
	if o111Zone(g, walker) != game.ZoneGraveyard {
		t.Errorf("the planeswalker is in %v, want the graveyard", o111Zone(g, walker))
	}
	if !effectstest.HasKeywordOf(g, elves, "flying") {
		t.Error("the merged creature lacks flying")
	}
}

// --- Migratory Greathorn ----------------------------------------------------------------

func TestS82OldB111MigratoryGreathornFetchesABasicLandTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := effectstest.PushLibraryTyped(me, "Forest", "Basic Land — Forest")
	effectstest.PushLibraryTyped(me, "Nonbasic", "Land")
	elves := effectstest.LlanowarOn(g, me.ID)
	o111Mutate(t, g, me, o111Greathorn, elves, false)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, forest)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.AangCardOnBF(g, forest)
	if !ok || !c.Tapped || c.Controller != me.ID {
		t.Errorf("the basic land is not on the battlefield tapped under my control: %+v (%v)", c, ok)
	}
}

// --- Trumpeting Gnarr -------------------------------------------------------------------

func TestS82OldB111TrumpetingGnarrMakesABeast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elves := effectstest.LlanowarOn(g, me.ID)
	o111Mutate(t, g, me, o111Gnarr, elves, true)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	beasts := effectstest.TokensNamed(g, "Beast")
	if len(beasts) != 1 {
		t.Fatalf("%d Beast tokens, want 1", len(beasts))
	}
	if p, tough := effectstest.PtOf(t, g, beasts[0].InstanceID); p != 3 || tough != 3 {
		t.Errorf("the Beast is %d/%d, want 3/3", p, tough)
	}
}

// --- Vadrok, Apex of Thunder ------------------------------------------------------------

func TestS82OldB111VadrokCastsACheapNoncreatureCardFromTheGraveyardFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	big := o111Grave(me, "Expensive", "Sorcery", "{2}{R}{R}", 0)
	bear := o111Grave(me, "Dead Bear", "Creature — Bear", "{1}", 2)
	cheap := o111Grave(me, "Cheap Rock", "Artifact", "{2}", 0)
	elves := effectstest.LlanowarOn(g, me.ID)
	o111Mutate(t, g, me, o111Vadrok, elves, false)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if effectstest.HasID(pick.PickTargetCards, big) || effectstest.HasID(pick.PickTargetCards, bear) || !effectstest.HasID(pick.PickTargetCards, cheap) {
		t.Errorf("offered %v: want only the noncreature card with mana value 3 or less", pick.PickTargetCards)
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, *o111Ref(cheap)); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, cheap, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if o111Zone(g, cheap) != game.ZoneBattlefield {
		t.Errorf("the free-cast artifact is in %v, want the battlefield", o111Zone(g, cheap))
	}
	if !effectstest.HasKeywordOf(g, elves, "first strike") || !effectstest.HasKeywordOf(g, elves, "flying") {
		t.Error("the merged creature lacks flying or first strike")
	}
}

// --- Necropanther -----------------------------------------------------------------------

func TestS82OldB111NecropantherReturnsACreatureOfManaValueThreeOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	small := o111Grave(me, "Small", "Creature — Bear", "{1}{G}", 2)
	big := o111Grave(me, "Big", "Creature — Bear", "{3}{G}", 4)
	spell := o111Grave(me, "Spell", "Instant", "{R}", 0)
	elves := effectstest.LlanowarOn(g, me.ID)
	o111Mutate(t, g, me, o111Necro, elves, false)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if effectstest.HasID(pick.PickTargetCards, big) || effectstest.HasID(pick.PickTargetCards, spell) {
		t.Errorf("offered %v: want only creature cards with mana value 3 or less", pick.PickTargetCards)
	}
	effectstest.AssistCardSettle(t, g, me.ID, o111Ref(small), true)
	if c, ok := effectstest.AangCardOnBF(g, small); !ok || c.Controller != me.ID {
		t.Errorf("the small creature is not on the battlefield under my control (%v)", ok)
	}
}

// --- Auspicious Starrix -----------------------------------------------------------------

func TestS82OldB111AuspiciousStarrixExilesUntilXPermanentCardsAndPutsThemOntoTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	// Top down: instant, creature, instant, creature, creature.
	c3 := effectstest.PushLibraryTyped(me, "Below", "Creature — Bear")
	effectstest.PushLibraryTyped(me, "Second", "Creature — Bear")
	i2 := effectstest.PushLibraryTyped(me, "Shock Two", "Instant")
	effectstest.PushLibraryTyped(me, "First", "Creature — Bear")
	i1 := effectstest.PushLibraryTyped(me, "Shock One", "Instant")
	onBattlefield := func(name string) bool {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name && c.Controller == me.ID {
				return true
			}
		}
		return false
	}
	elves := effectstest.LlanowarOn(g, me.ID)
	o111Mutate(t, g, me, o111Starrix, elves, false)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if !onBattlefield("First") {
		t.Error("the first permanent card did not enter")
	}
	if onBattlefield("Second") {
		t.Error("a second permanent card entered after one mutation")
	}
	if o111Zone(g, i1) != game.ZoneExile {
		t.Errorf("the instant above the creature is in %v, want exile", o111Zone(g, i1))
	}
	if o111Zone(g, i2) != game.ZoneLibrary || o111Zone(g, c3) != game.ZoneLibrary {
		t.Error("the run went past the first permanent card")
	}
	// A second mutation: two Starrixes are in the permanent and each one's
	// ability triggers with X = 2, so four more permanent cards are asked for;
	// the library holds two, which enter.
	o111Mutate(t, g, me, o111Starrix, elves, false)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if !onBattlefield("Second") || !onBattlefield("Below") {
		t.Error("the second mutation did not put the remaining permanent cards onto the battlefield")
	}
}

func TestS82OldB111AuspiciousStarrixWithNoPermanentCardsExilesTheWholeLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Library.Cards = nil })
	only := effectstest.PushLibraryTyped(me, "Only Instant", "Instant")
	elves := effectstest.LlanowarOn(g, me.ID)
	o111Mutate(t, g, me, o111Starrix, elves, false)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if o111Zone(g, only) != game.ZoneExile {
		t.Errorf("the only card is in %v, want exile", o111Zone(g, only))
	}
}
