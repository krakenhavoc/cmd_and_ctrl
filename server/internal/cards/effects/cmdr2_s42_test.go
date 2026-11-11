package effects

import (
	"slices"

	"github.com/google/uuid"

	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const plasmCaptureOracle = "3673219d-d2b5-43ad-b0c8-58b583491158"

func TestPlasmCaptureCountersAndAddsManaAtNextFirstMain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := batch01OpponentCasts(t, g, opp, "Big Bolt", lightningBoltOracle, "{2}{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	castCatalogSpell(t, g, "Plasm Capture", "Instant", plasmCaptureOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(spell) {
		t.Fatal("the spell was not countered")
	}
	if n := len(g.DelayedTriggers); n != 1 || !g.DelayedTriggers[0].ControllerTurnOnly ||
		g.DelayedTriggers[0].At != game.StepPrecombatMain {
		t.Fatalf("want one delayed trigger on the caster's precombat main, got %+v", g.DelayedTriggers)
	}
	// This turn's postcombat main is not a first main phase.
	batch01AdvanceToStepOf(t, g, 0, game.StepPostcombatMain)
	if len(g.DelayedTriggers) != 1 {
		t.Fatal("the mana fired in the second main phase")
	}
	// The opponent's precombat main is not mine.
	batch01AdvanceToStepOf(t, g, 1, game.StepPrecombatMain)
	if len(g.DelayedTriggers) != 1 {
		t.Fatal("the mana fired on an opponent's first main phase")
	}
	batch01AdvanceToStepOf(t, g, 0, game.StepPrecombatMain)
	if len(g.DelayedTriggers) != 0 {
		t.Fatal("the mana did not fire on my next first main phase")
	}
	passPriorityAroundTable(t, g)
	for _, c := range []string{"W", "U", "U"} {
		pick := riderLatestManaPick(g, me.ID)
		if pick == nil {
			t.Fatalf("missing a colour pick for %s", c)
		}
		if err := g.ResolveManaChoice(pick.ID, me.ID, c); err != nil {
			t.Fatalf("ResolveManaChoice: %v", err)
		}
	}
	if me.ManaPool.Total() != 3 || msc12PoolOf(me, "W") != 1 || msc12PoolOf(me, "U") != 2 {
		t.Errorf("pool %v, want W U U (the countered spell's mana value, 3)", batch01PoolColors(me))
	}
}

const counterpointOracle = "dad4af15-5b26-417d-a41b-29e671e878f4"

func TestCounterpointCastsASpellOfEqualOrLowerManaValueFromYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := uuid.New()
	big := gyCard(me, "Big Sorcery", "Sorcery", "{3}{R}", "")
	me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear",
		ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	land := gyCard(me, "Forest", "Basic Land — Forest", "", "")
	spell := batch01OpponentCasts(t, g, opp, "Big Bolt", lightningBoltOracle, "{2}{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	castCatalogSpell(t, g, "Counterpoint", "Instant", counterpointOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(spell) {
		t.Fatal("the spell was not countered")
	}
	cast := func(id uuid.UUID) error {
		return g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", Strict: true,
			AlternativeCost: game.AltCostKeyGrantedCost})
	}
	if err := cast(big); err == nil {
		t.Error("a mana value 4 spell was cast under a cap of 3")
	}
	if err := cast(land); err == nil {
		t.Error("a land was cast")
	}
	if err := cast(bear); err != nil {
		t.Fatalf("cast the Bear free: %v", err)
	}
	if g.Stack.Size() == 0 {
		t.Fatal("the Bear is not on the stack")
	}
}

func TestCounterpointOffersNothingAboveTheCounteredSpellsManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := gyCard(me, "Big Sorcery", "Sorcery", "{3}{R}", "")
	spell := batch01OpponentCasts(t, g, opp, "Big Bolt", lightningBoltOracle, "{2}{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	castCatalogSpell(t, g, "Counterpoint", "Instant", counterpointOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	passPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, big, game.CastSpellParams{FromZone: "graveyard", Strict: true,
		AlternativeCost: game.AltCostKeyGrantedCost}); err == nil {
		t.Error("cast a mana value 4 spell under a cap of 3")
	}
}

const diluvianPrimordialOracle = "f21ce158-2925-4658-ab7b-b73718d31965"

func TestDiluvianPrimordialCastsFromEachOpponentsGraveyardAndExilesThem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var opps []*game.Player
	for _, p := range g.Seats {
		if p.ID != me.ID {
			opps = append(opps, p)
		}
	}
	a := gyCard(opps[0], "A's Sorcery", "Sorcery", "{4}{R}", "")
	creature := gyCard(opps[0], "A's Bear", "Creature — Bear", "{1}{G}", "")
	b := gyCard(opps[1], "B's Instant", "Instant", "{U}", "")
	mine := gyCard(me, "My Instant", "Instant", "{U}", "")
	castHoldCreature(t, g, "Diluvian Primordial", "Creature — Avatar", diluvianPrimordialOracle, 5, 5)

	ch := pickTargetPrompt(t, g)
	if slices.Contains(ch.PickTargetCards, creature) || !slices.Contains(ch.PickTargetCards, a) ||
		slices.Contains(ch.PickTargetCards, b) || ch.PickTargetMin != 0 || ch.PickTargetMax != 1 {
		t.Fatalf("first clause: only the first opponent's instants and sorceries, up to one: %v", ch.PickTargetCards)
	}
	answerPickTarget(t, g, a)
	answerPickTarget(t, g, b)
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{a, b} {
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", Strict: true,
			AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
			t.Fatalf("cast %s free: %v", id, err)
		}
	}
	if err := g.CastSpell(me.ID, mine, game.CastSpellParams{FromZone: "graveyard", Strict: true,
		AlternativeCost: game.AltCostKeyGrantedCost}); err == nil {
		t.Error("an untargeted card was cast")
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			t.Errorf("a spell cast this way was not exiled instead of going to a graveyard")
		}
	}
}

const direFleetDaredevilOracle = "7a3e50a5-c163-4c41-b62d-d52c233c55b7"

func TestDireFleetDaredevilExilesAndLetsYouCastWithAnyManaThenExilesIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	mineGY := gyCard(me, "My Instant", "Instant", "{U}", "")
	bolt := gyCard(opp, "Opp Bolt", "Instant", "{R}", "")
	creature := gyCard(opp, "Opp Bear", "Creature — Bear", "{1}{G}", "")
	castHoldCreature(t, g, "Dire Fleet Daredevil", "Creature — Human Pirate", direFleetDaredevilOracle, 2, 1)
	ch := pickTargetPrompt(t, g)
	if slices.Contains(ch.PickTargetCards, mineGY) || slices.Contains(ch.PickTargetCards, creature) ||
		!slices.Contains(ch.PickTargetCards, bolt) {
		t.Fatalf("only an opponent's instant or sorcery: %v", ch.PickTargetCards)
	}
	answerPickTarget(t, g, bolt)
	passPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(bolt); z == nil || z.Kind != game.ZoneExile {
		t.Fatal("the card was not exiled")
	}
	// Not free: a pool that cannot pay refuses. Mana of any type pays {R}.
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{FromZone: "exile", Strict: true, AlternativeCost: game.AltCostKeyGrantedCost,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Fatal("cast it for free with an empty pool")
	}
	floatForTest(g, me, "U")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{FromZone: "exile", Strict: true, AlternativeCost: game.AltCostKeyGrantedCost,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast it with blue mana: %v", err)
	}
	passPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(bolt); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("the spell was not exiled instead of put into a graveyard: %v", z)
	}
}

const missionBriefingOracle = "c3a68018-9eff-47e6-a612-182886d28fe3"

func TestMissionBriefingSurveilsThenLetsYouCastAChosenInstantOrSorceryThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedLibrary(me, "Bin Me", "Keep Me", "Third")
	old := gyCard(me, "Old Bolt", "Instant", "{R}", lightningBoltOracle)
	creature := gyCard(me, "Old Bear", "Creature — Bear", "{1}{G}", "")
	castCatalogSpell(t, g, "Mission Briefing", "Instant", missionBriefingOracle, nil)
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("surveil prompt %+v, want one over two cards", c)
	}
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{ids[0]}, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	answerChooseCards(t, g, me.ID, old)
	passPriorityAroundTable(t, g)

	cast := func(id uuid.UUID) error {
		return g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", Strict: true,
			AlternativeCost: game.AltCostKeyGrantedCost, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}})
	}
	if err := cast(creature); err == nil {
		t.Error("an unchosen card was cast")
	}
	if err := cast(old); err == nil {
		t.Fatal("cast the chosen card without paying its mana cost")
	}
	floatForTest(g, me, "R")
	if err := cast(old); err != nil {
		t.Fatalf("cast the chosen card: %v", err)
	}
	passPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(old); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("the spell was not exiled instead of put into the graveyard")
	}
}

const haloForagerOracle = "210d1077-6e60-4f06-a8a4-10b842979ac5"

func TestHaloForagerPaysXThenCastsAnInstantOrSorceryOfThatManaValueFromAnyGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	two := gyCard(opp, "Two Instant", "Instant", "{1}{R}", "")
	three := gyCard(me, "Three Sorcery", "Sorcery", "{2}{R}", "")
	bear := gyCard(me, "Two Bear", "Creature — Bear", "{1}{G}", "")
	advanceToMain(t, g)
	fillPool(me, 2)
	castHoldCreature(t, g, "Halo Forager", "Creature — Faerie Rogue", haloForagerOracle, 3, 1)
	passPriorityAroundTable(t, g)
	pa := answerNumber(t, g, me.ID, 2)
	if pa.Min != 0 {
		t.Fatalf("X prompt = %+v", pa)
	}
	answerPayUnless(t, g, me.ID, true)
	ch := pickTargetPrompt(t, g)
	if !slices.Contains(ch.PickTargetCards, two) || slices.Contains(ch.PickTargetCards, three) || slices.Contains(ch.PickTargetCards, bear) {
		t.Fatalf("only an instant or sorcery with mana value exactly 2: %v", ch.PickTargetCards)
	}
	answerPickTarget(t, g, two)
	passPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, two, game.CastSpellParams{FromZone: "graveyard", Strict: true,
		AlternativeCost: game.AltCostKeyGrantedCost, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast it free: %v", err)
	}
	passPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(two); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("the spell was not exiled instead of put into a graveyard")
	}
}

const bloodthirstyAdversaryOracle = "3b943c3e-1f18-4997-b6d0-28ef75aed88e"

func TestBloodthirstyAdversaryPaysThenCountersAndCopiesFromYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bolt := gyCard(me, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
	big := gyCard(me, "Big Sorcery", "Sorcery", "{2}{R}{R}", "")
	oppBolt := gyCard(opp, "Opp Bolt", "Instant", "{R}", lightningBoltOracle)
	advanceToMain(t, g)
	fillPoolColored(me, "R", 2)
	fillPool(me, 4)
	adv := castHoldCreature(t, g, "Bloodthirsty Adversary", "Creature — Vampire", bloodthirstyAdversaryOracle, 2, 2)
	passPriorityAroundTable(t, g)
	pa := answerNumber(t, g, me.ID, 2)
	if pa.Max != 2 {
		t.Fatalf("X prompt = %+v, want up to 2 ({2}{R} twice from 6 mana)", pa)
	}
	answerPayUnless(t, g, me.ID, true)
	ch := pickTargetPrompt(t, g)
	if !slices.Contains(ch.PickTargetCards, bolt) || slices.Contains(ch.PickTargetCards, big) ||
		slices.Contains(ch.PickTargetCards, oppBolt) || ch.PickTargetMax != 2 || ch.PickTargetMin != 0 {
		t.Fatalf("up to two of my instants and sorceries with mana value 3 or less: %v (%d..%d)",
			ch.PickTargetCards, ch.PickTargetMin, ch.PickTargetMax)
	}
	answerPickTarget(t, g, bolt)
	passPriorityAroundTable(t, g)

	c, ok := battlefieldCard(g, adv)
	if !ok || c.Counters[game.CounterPlusOne] != 2 {
		t.Errorf("Adversary counters = %v, want two +1/+1 counters", c.Counters)
	}
	if !inExile(g, bolt) {
		t.Fatal("the targeted card was not exiled")
	}
	cp := copyOf(t, g, "Lightning Bolt")
	before := opp.Life
	if err := castCopy(g, me.ID, cp.InstanceID, true, playerTarget(opp.ID)); err != nil {
		t.Fatalf("casting the copy free: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("life = %d, want %d", opp.Life, before-3)
	}
}

const finaleOfPromiseOracle = "d6fafb50-9531-4fbf-bb1e-ebb4dd39281c"

func castFinaleOfPromise(t *testing.T, g *game.Game, me *game.Player, x int, targets ...uuid.UUID) {
	t.Helper()
	advanceToMain(t, g)
	fillPoolColored(me, "R", 2)
	fillPool(me, x)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Finale of Promise", TypeLine: "Sorcery", ManaCost: "{X}{R}{R}",
		OracleID: finaleOfPromiseOracle, Owner: me.ID, Controller: me.ID})
	var refs []game.TargetRef
	for _, tg := range targets {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: tg})
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: x, Targets: refs, Strict: true}); err != nil {
		t.Fatalf("cast Finale of Promise: %v", err)
	}
}

func TestFinaleOfPromiseCastsAnInstantAndASorceryFreeAndExilesThem(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bolt := gyCard(me, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
	sorc := gyCard(me, "Small Sorcery", "Sorcery", "{1}{R}", "")
	big := gyCard(me, "Big Sorcery", "Sorcery", "{3}{R}", "")
	oppBolt := gyCard(opp, "Opp Instant", "Instant", "{R}", lightningBoltOracle)
	advanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Finale of Promise", TypeLine: "Sorcery", ManaCost: "{X}{R}{R}",
		OracleID: finaleOfPromiseOracle, Owner: me.ID, Controller: me.ID})
	fillPoolColored(me, "R", 2)
	fillPool(me, 3)
	for _, bad := range []uuid.UUID{big, oppBolt} {
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 3, Strict: true,
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bad}}}); err == nil {
			t.Errorf("targeted a card above X or in another graveyard: %v", bad)
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 3, Strict: true, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: bolt}, {Kind: game.TargetCard, ID: sorc}}}); err != nil {
		t.Fatalf("cast Finale of Promise: %v", err)
	}
	passPriorityAroundTable(t, g)
	free := func(card uuid.UUID, targets ...game.TargetRef) error {
		return g.CastSpell(me.ID, card, game.CastSpellParams{FromZone: "graveyard", Strict: true,
			AlternativeCost: game.AltCostKeyGrantedCost, Targets: targets})
	}
	if err := free(bolt, playerTarget(opp.ID)); err != nil {
		t.Fatalf("cast the instant free: %v", err)
	}
	if err := free(sorc); err != nil {
		t.Fatalf("cast the sorcery free (its own permission): %v", err)
	}
	if err := free(big); err == nil {
		t.Error("cast an untargeted card")
	}
	passPriorityAroundTable(t, g)
	for _, c := range []uuid.UUID{bolt, sorc} {
		if z := g.FindCardZoneForEffect(c); z == nil || z.Kind != game.ZoneExile {
			t.Errorf("a spell cast this way was not exiled instead of put into the graveyard")
		}
	}
}

func TestFinaleOfPromiseCopiesEachSpellTwiceAtXTenOrMore(t *testing.T) {
	for _, tc := range []struct{ x, wantCopies int }{{9, 0}, {10, 2}} {
		g := newCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
		bolt := gyCard(me, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
		castFinaleOfPromise(t, g, me, tc.x, bolt)
		passPriorityAroundTable(t, g)
		if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{FromZone: "graveyard", Strict: true,
			AlternativeCost: game.AltCostKeyGrantedCost, Targets: []game.TargetRef{playerTarget(opp.ID)}}); err != nil {
			t.Fatalf("X=%d: cast the instant free: %v", tc.x, err)
		}
		// Each copy offers new targets (CR 707.10c): keep the opponent.
		prompts := 0
		for c := copyPromptFor(g, me.ID); c != nil && prompts < 4; c = copyPromptFor(g, me.ID) {
			prompts++
			if err := g.ResolvePickTarget(c.ID, me.ID, playerTarget(opp.ID)); err != nil {
				t.Fatalf("X=%d: new targets for a copy: %v", tc.x, err)
			}
		}
		if prompts != tc.wantCopies {
			t.Errorf("X=%d: asked for new targets %d times, want %d", tc.x, prompts, tc.wantCopies)
		}
		if got := stackCopies(g); got != tc.wantCopies {
			t.Errorf("X=%d: %d copies on the stack, want %d", tc.x, got, tc.wantCopies)
		}
		before := opp.Life
		passPriorityAroundTable(t, g)
		if want := before - 3*(1+tc.wantCopies); opp.Life != want {
			t.Errorf("X=%d: life %d, want %d", tc.x, opp.Life, want)
		}
	}
}
