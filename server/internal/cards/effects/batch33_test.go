package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch33_test.go — card-level coverage for the card-coverage
// roadmap's batch 33 (#396, `edhrec_rank` 3447–3548): the "no new
// machinery" group. One test per observable behaviour, driven through
// a real cast, activation or attack rather than by calling primitives.

const (
	b33VisionsOfBeyondOracle     = "c68964df-51ae-43b0-abea-74735016c13f"
	b33AlelaOracle               = "1cae5752-b4af-4a8f-8c8c-2493e163083b"
	b33CoalitionRelicOracle      = "008cb342-79f5-4df6-a6b7-0e9e22ed693f"
	b33BurningInquiryOracle      = "af98156b-3064-4f16-940e-10039241f2b0"
	b33LathielOracle             = "d3c56fc4-3611-41b1-952e-4c5311b1510b"
	b33RaidBombardmentOracle     = "1734e777-d34a-4526-ae24-ef034f147cc5"
	b33BlastingStationOracle     = "3a38d2d1-c4ff-4088-b1df-5feb9602ee2e"
	b33MajaOracle                = "0a2075b5-9609-433d-bcd2-e0a637456cf8"
	b33DalkovanEncampmentOracle  = "33a90122-7280-4481-9b97-5879194cae40"
	b33CountersquallOracle       = "6df620b8-1e54-4f63-8556-12c75e5679af"
	b33NestOfScarabsOracle       = "1f21cf59-6390-44b4-ab2e-ef290dfd8c85"
	b33RecklessBarbarianOracle   = "a621ae15-122a-4e11-bfcf-11dfc09784c0"
	b33CloudblazerOracle         = "f84d1291-1f82-4b67-a26e-b79624b4ce1d"
	b33RakdosJoinsUpOracle       = "6a47865c-8fa2-4cb2-aebf-8009c065395f"
	b33GoodFortuneUnicornOracle  = "1e0cdff3-3ec5-41fc-8053-f072bae156b3"
	b33OverseerOfTheDamnedOracle = "c1f085ce-5b52-45b3-aef0-f77f36b3da36"
	b33PatronOfTheVeinOracle     = "8dbc8fdb-36ce-4f30-b679-9c2029fcd9c6"
	b33FervorOracle              = "8e0cea9c-3110-4728-9378-76849e33bb90"
	b33HermesOracle              = "63e2cba7-ec1a-44bf-b915-d2ae75851430"
	b33SidisiOracle              = "3fad7072-21e3-446e-a28f-615038c8bfea"
	b33FlowOfKnowledgeOracle     = "b65f20de-52fa-4904-bbfe-7ba53ccd8ae5"
	b33CopperMyrOracle           = "8b52f30c-5e38-4333-88ab-901b37105b36"
	b33PlanarEngineeringOracle   = "dc48079a-b1c4-4c43-a989-418c78a264ec"
	b33DoublingCubeOracle        = "9afd8f12-0796-4500-aaa3-10b4a46ef6ec"
	b33StoneOfErechOracle        = "73dad679-1edb-41c9-9d43-56dc93c3e9fe"
	b33GodEternalBontuOracle     = "183891b0-b5ec-47f4-8d09-b9d3cfc4e7f1"

	// Contagion Clasp puts a -1/-1 counter on target creature as it
	// enters — the batch's -1/-1 counter source for Nest of Scarabs.
	b33ContagionClaspOracle = "43f2d81e-aa01-4fa9-9046-6a27a05dbd2d"
)

// b33Hands is every seat's hand size, seat order.
func b33Hands(g *game.Game) []int {
	out := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		out[i] = p.Hand.Size()
	}
	return out
}

// b33Graveyards is every seat's graveyard size, seat order.
func b33Graveyards(g *game.Game) []int {
	out := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		out[i] = p.Graveyard.Size()
	}
	return out
}

// b33Goaded reads a battlefield card's goad marker.
func b33Goaded(t *testing.T, g *game.Game, id uuid.UUID) uuid.UUID {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.GoadedBy
}

// b33PoolCount counts the tokens of one colour in a player's pool.
func b33PoolCount(p *game.Player, color string) int {
	n := 0
	for _, tok := range p.ManaPool {
		if tok.Color == color {
			n++
		}
	}
	return n
}

// b33AttackingTokensNamed lists the tokens named `name` under
// `controller` that are attacking `defender`.
func b33AttackingTokensNamed(g *game.Game, controller, defender uuid.UUID, name string) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name && c.AttackingTarget == defender {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Copper Myr is
// a table row, so a transposed row is invisible until someone plays
// that exact card.
func TestBatch33CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b33VisionsOfBeyondOracle:     "Visions of Beyond",
		b33AlelaOracle:               "Alela, Cunning Conqueror",
		b33CoalitionRelicOracle:      "Coalition Relic",
		b33BurningInquiryOracle:      "Burning Inquiry",
		b33LathielOracle:             "Lathiel, the Bounteous Dawn",
		b33RaidBombardmentOracle:     "Raid Bombardment",
		b33BlastingStationOracle:     "Blasting Station",
		b33MajaOracle:                "Maja, Bretagard Protector",
		b33DalkovanEncampmentOracle:  "Dalkovan Encampment",
		b33CountersquallOracle:       "Countersquall",
		b33NestOfScarabsOracle:       "Nest of Scarabs",
		b33RecklessBarbarianOracle:   "Reckless Barbarian",
		b33CloudblazerOracle:         "Cloudblazer",
		b33RakdosJoinsUpOracle:       "Rakdos Joins Up",
		b33GoodFortuneUnicornOracle:  "Good-Fortune Unicorn",
		b33OverseerOfTheDamnedOracle: "Overseer of the Damned",
		b33PatronOfTheVeinOracle:     "Patron of the Vein",
		b33FervorOracle:              "Fervor",
		b33HermesOracle:              "Hermes, Overseer of Elpis",
		b33SidisiOracle:              "Sidisi, Brood Tyrant",
		b33FlowOfKnowledgeOracle:     "Flow of Knowledge",
		b33CopperMyrOracle:           "Copper Myr",
		b33PlanarEngineeringOracle:   "Planar Engineering",
		b33DoublingCubeOracle:        "Doubling Cube",
		b33StoneOfErechOracle:        "Stone of Erech",
		b33GodEternalBontuOracle:     "God-Eternal Bontu",
	}
	if len(want) != 26 {
		t.Fatalf("the batch ships 26 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
}

// --- spells ---------------------------------------------------------

func TestB33VisionsOfBeyondDrawsThreeOnceAGraveyardHoldsTwenty(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Visions of Beyond", "Instant", b33VisionsOfBeyondOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("no full graveyard: hand %d → %d, want one drawn (the cast helper adds one, the cast removes it)", hand, got)
	}
	for i := 0; i < 19; i++ {
		pushGraveyardCardForTest(opp, "Filler")
	}
	hand = me.Hand.Size()
	castCatalogSpell(t, g, "Visions of Beyond", "Instant", b33VisionsOfBeyondOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("nineteen cards is not twenty: hand %d → %d", hand, got)
	}
	pushGraveyardCardForTest(opp, "Filler")
	hand = me.Hand.Size()
	castCatalogSpell(t, g, "Visions of Beyond", "Instant", b33VisionsOfBeyondOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("an opponent's twenty-card graveyard: hand %d → %d, want three drawn", hand, got)
	}
}

func TestB33BurningInquiryEveryPlayerDrawsThreeThenDiscardsThreeAtRandom(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for _, p := range g.Seats {
		fillHandTo(t, g, p, 4)
	}
	hands, graveyards := b33Hands(g), b33Graveyards(g)
	// The cast helper puts the spell in hand and the cast removes it,
	// so every seat's hand — the caster's included — ends where it
	// started: three drawn, three discarded.
	castCatalogSpell(t, g, "Burning Inquiry", "Sorcery", b33BurningInquiryOracle, nil)
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if got := p.Hand.Size(); got != hands[i] {
			t.Errorf("seat %d: hand %d → %d, want three drawn and three discarded", i, hands[i], got)
		}
		want := graveyards[i] + 3
		if p.ID == me.ID {
			want++ // the Inquiry itself
		}
		if got := p.Graveyard.Size(); got != want {
			t.Errorf("seat %d: graveyard %d → %d, want %d", i, graveyards[i], got, want)
		}
	}
	if len(g.DiscardPending) != 0 {
		t.Error("the discards are at random — no prompt")
	}
}

func TestB33FlowOfKnowledgeDrawsPerIslandThenAsksForTwoDiscards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	b12Permanent(g, me.ID, "Hallowed Fountain", "Land — Plains Island")
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	b12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Flow of Knowledge", "Instant", b33FlowOfKnowledgeOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("three Islands you control (one nonbasic): hand %d → %d, want +3", hand, got)
	}
	if got := discardOwed(g, me.ID); got != 2 {
		t.Errorf("then discard two — %d discards pending", got)
	}
}

func TestB33CountersquallCountersANoncreatureSpellAndItsControllerLosesTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	if err := b09TryCast(t, g, "Countersquall", "Instant", b33CountersquallOracle, b16TargetCard(bear)); err == nil {
		t.Fatal("a creature spell is not a legal target")
	}
	passPriorityAroundTable(t, g)
	bolt := batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	life := opp.Life
	castCatalogSpell(t, g, "Countersquall", "Instant", b33CountersquallOracle, b16TargetCard(bolt))
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) {
		t.Fatal("the Bolt was not countered")
	}
	if me.Life != 40 {
		t.Errorf("the countered Bolt dealt damage: life %d", me.Life)
	}
	if opp.Life != life-2 {
		t.Errorf("the Bolt's controller loses 2: %d → %d", life, opp.Life)
	}
}

func TestB33PlanarEngineeringSacrificesTwoLandsAndFetchesFourBasicsTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	swamp := b12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	forest := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	crypt := b12Permanent(g, me.ID, "Blood Crypt", "Land — Swamp Mountain")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	ids := seedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"},
		game.Card{Name: "Snow Swamp", TypeLine: "Basic Snow Land — Swamp"},
		game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Hallowed Fountain", TypeLine: "Land — Plains Island"},
	)
	castCatalogSpell(t, g, "Planar Engineering", "Sorcery", b33PlanarEngineeringOracle, nil)
	passPriorityAroundTable(t, g)
	c := sacrificeChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no sacrifice prompt")
	}
	if hasID(c.SacrificeOptions, bear) || !hasID(c.SacrificeOptions, crypt) {
		t.Error("only lands are offered, nonbasics included")
	}
	answerSacrifice(t, g, me.ID, swamp)
	answerSacrifice(t, g, me.ID, crypt)
	if sacrificeChoiceFor(g, me.ID) != nil {
		t.Error("two lands, not three")
	}
	if !g.Battlefield.Contains(forest) || !g.Battlefield.Contains(bear) {
		t.Error("the unchosen land and the creature stay")
	}
	s := searchChoiceFor(g, me.ID)
	if s == nil {
		t.Fatal("no search prompt")
	}
	if s.SearchMax != 4 || len(s.SearchCards) != 5 {
		t.Errorf("five basics offered (the snow basic included, the shockland not), up to four taken: %d offered, max %d", len(s.SearchCards), s.SearchMax)
	}
	answerSearchByID(t, g, me.ID, ids[:4]...)
	for _, id := range ids[:4] {
		if !g.Battlefield.Contains(id) {
			t.Fatal("every chosen basic is put onto the battlefield")
		}
		if !b16Tapped(t, g, id) {
			t.Error("the basics enter tapped")
		}
	}
	if me.Library.Size() != 2 {
		t.Errorf("library %d, want 2 (an Island and the shockland)", me.Library.Size())
	}
}

// --- artifacts ------------------------------------------------------

func TestB33CoalitionRelicBanksChargeCountersForYourFirstMainPhase(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	relic := pushCatalogPermanent(g, me.ID, "Coalition Relic", "Artifact", b33CoalitionRelicOracle, false)
	advanceToMain(t, g)
	// Entering the main phase fires the trigger with no counters on
	// the Relic: it resolves and adds nothing.
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 || len(me.ManaPool) != 0 {
		t.Fatal("no counters, no mana")
	}
	if err := g.ActivateManaAbility(me.ID, relic, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the mana ability: %v", err)
	}
	if pick := riderLatestManaPick(g, me.ID); pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("one mana of any colour, got %+v", pick)
	}
	b10ResolveAllManaPicks(t, g, me.ID, "U")
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(relic) })
	b16Activate(t, g, me.ID, relic, 0, game.ActivateAbilityParams{})
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(relic) })
	b16Activate(t, g, me.ID, relic, 0, game.ActivateAbilityParams{})
	if got := counterCount(g, relic, "charge"); got != 2 {
		t.Fatalf("two activations, %d charge counters", got)
	}
	// The pool empties between steps; the trigger fires at the
	// start of the controller's next precombat main.
	advanceToPrecombatMainOf(t, g, 0)
	if len(me.ManaPool) != 0 {
		t.Fatal("the pool is empty as the main phase begins")
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, relic, "charge"); got != 0 {
		t.Errorf("the counters come off: %d left", got)
	}
	if n := b10ResolveAllManaPicks(t, g, me.ID, "R"); n != 2 {
		t.Errorf("one colour pick per counter: %d", n)
	}
	if got := b33PoolCount(me, "R"); got != 2 {
		t.Errorf("two red mana in the pool, got %d", got)
	}
	// An opponent's main phase is silent; with no counters, so is yours.
	advanceToPrecombatMainOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 {
		t.Error("an opponent's main phase is not yours")
	}
	advanceToPrecombatMainOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 || len(me.ManaPool) != 0 {
		t.Error("no counters, no mana")
	}
}

func TestB33BlastingStationPingsForASacrificeAndMayUntapWhenACreatureEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	station := pushCatalogPermanent(g, me.ID, "Blasting Station", "Artifact", b33BlastingStationOracle, false)
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	life := opp.Life
	b16Activate(t, g, me.ID, station, 0, game.ActivateAbilityParams{Targets: b16TargetPlayer(opp.ID), SacrificeIDs: []uuid.UUID{fodder}})
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrifice is the cost")
	}
	if !b16Tapped(t, g, station) {
		t.Error("the tap is the cost")
	}
	if opp.Life != life-1 {
		t.Errorf("1 damage to the target: %d → %d", life, opp.Life)
	}
	// A creature entering — anyone's — offers the untap.
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if b16Tapped(t, g, station) {
		t.Error("the Station untaps")
	}
	// Declined, it stays tapped; an artifact entering offers nothing.
	b16Tap(g, station)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, RedGoblinToken(), 1) })
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, station) {
		t.Error("declining leaves it tapped")
	}
	castCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	if latestTriggerPrompt(g, me.ID) != nil {
		t.Error("an artifact entering is not a creature")
	}
	_ = theirs
}

func TestB33DoublingCubeDoublesWhatIsLeftAfterPayingThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cube := pushCatalogPermanent(g, me.ID, "Doubling Cube", "Artifact", b33DoublingCubeOracle, false)
	advanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, cube, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("the {3} is a cost — nothing to pay it with")
	}
	b06AddMana(me, "C", "C", "C", "R", "R", "G")
	if err := g.ActivateManaAbility(me.ID, cube, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if !b16Tapped(t, g, cube) {
		t.Error("the tap is the cost")
	}
	if got := len(me.ManaPool); got != 6 {
		t.Fatalf("three paid, three left, three added: pool %v", poolColors(me))
	}
	if r, gr, c := b33PoolCount(me, "R"), b33PoolCount(me, "G"), b33PoolCount(me, "C"); r != 4 || gr != 2 || c != 0 {
		t.Errorf("the colourless paid the {3} and the rest doubled: pool %v", poolColors(me))
	}
	// An empty pool after paying adds nothing.
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(cube) })
	me.ManaPool = nil
	b06AddMana(me, "C", "C", "C")
	if err := g.ActivateManaAbility(me.ID, cube, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("nothing left to double: pool %v", poolColors(me))
	}
}

func TestB33StoneOfErechExilesOpposingDeathsAndEatsAGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stone := pushCatalogPermanent(g, me.ID, "Stone of Erech", "Legendary Artifact", b33StoneOfErechOracle, false)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	b25Destroy(g, theirs)
	b25Destroy(g, mine)
	b25Destroy(g, rock)
	if !g.Exile.Contains(theirs) {
		t.Error("an opponent's creature dying is exiled instead")
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("your own creature dies as normal")
	}
	if !opp.Graveyard.Contains(rock) {
		t.Error("an opponent's noncreature dies as normal")
	}
	old := pushGraveyardCardForTest(opp, "Old Card")
	yard := opp.Graveyard.Size()
	advanceToMain(t, g)
	b06AddMana(me, "C", "C")
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, stone, 0, game.ActivateAbilityParams{Targets: b16TargetPlayer(opp.ID)})
	if opp.Graveyard.Size() != 0 {
		t.Errorf("the target player's graveyard (%d cards) is exiled: %d left", yard, opp.Graveyard.Size())
	}
	// "Exile" names a destination — an emptied graveyard alone is
	// equally consistent with the cards having been deleted.
	for _, id := range []uuid.UUID{rock, old} {
		if !g.Exile.Contains(id) {
			t.Errorf("graveyard card %s left the graveyard but was not exiled", id)
		}
	}
	if me.Hand.Size() != hand+1 {
		t.Error("then draw a card")
	}
	if g.Battlefield.Contains(stone) {
		t.Error("the Stone is sacrificed")
	}
}

// --- lands and mana -------------------------------------------------

func TestB33CopperMyrTapsForGreen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	myr := pushCatalogPermanent(g, me.ID, "Copper Myr", "Artifact Creature — Myr", b33CopperMyrOracle, false)
	if err := g.ActivateManaAbility(me.ID, myr, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool %v, want [G]", got)
	}
}

func TestB33RecklessBarbarianSacrificesForTwoRed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	barbarian := pushCatalogPermanent(g, me.ID, "Reckless Barbarian", "Creature — Dragon Barbarian", b33RecklessBarbarianOracle, true)
	if err := g.ActivateManaAbility(me.ID, barbarian, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("no tap in the cost, so summoning sickness does not apply: %v", err)
	}
	if g.Battlefield.Contains(barbarian) || !me.Graveyard.Contains(barbarian) {
		t.Error("the Barbarian is sacrificed")
	}
	if got := poolColors(me); len(got) != 2 || got[0] != "R" || got[1] != "R" {
		t.Errorf("pool %v, want [R R]", got)
	}
}

func TestB33DalkovanEncampmentEntersTappedWithoutASwampOrMountain(t *testing.T) {
	g := newCatalogGame(t)
	land := b12PlayFromHand(t, g, "Dalkovan Encampment", "Land", b33DalkovanEncampmentOracle, game.CastSpellParams{})
	if !b16Tapped(t, g, land) {
		t.Error("with no Swamp or Mountain it enters tapped")
	}
	g2 := newCatalogGame(t)
	b12Permanent(g2, g2.Seats[0].ID, "Blood Crypt", "Land — Swamp Mountain")
	land2 := b12PlayFromHand(t, g2, "Dalkovan Encampment", "Land", b33DalkovanEncampmentOracle, game.CastSpellParams{})
	if b16Tapped(t, g2, land2) {
		t.Error("with a Mountain it enters untapped")
	}
	if err := g2.ActivateManaAbility(g2.Seats[0].ID, land2, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(g2.Seats[0]); len(got) != 1 || got[0] != "W" {
		t.Errorf("pool %v, want [W]", got)
	}
}

func TestB33DalkovanEncampmentPrimedMakesTwoAttackingWarriorsPerActivation(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := pushCatalogPermanent(g, me.ID, "Dalkovan Encampment", "Land", b33DalkovanEncampmentOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	elf := b12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	// Unprimed, an attack makes nothing.
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Warrior"); n != 0 {
		t.Fatalf("no activation this turn: %d Warriors", n)
	}
	// Primed twice on the next turn: two attackers, one trigger,
	// four Warriors.
	advanceToPrecombatMainOf(t, g, 0)
	b06AddMana(me, "W", "C", "C")
	b16Activate(t, g, me.ID, land, 0, game.ActivateAbilityParams{})
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	b06AddMana(me, "W", "C", "C")
	b16Activate(t, g, me.ID, land, 0, game.ActivateAbilityParams{})
	declareAttack(t, g, opp.ID, bear, elf)
	passPriorityAroundTable(t, g)
	warriors := b33AttackingTokensNamed(g, me.ID, opp.ID, "Warrior")
	if len(warriors) != 4 {
		t.Fatalf("two activations, two attackers: %d Warriors tapped and attacking, want 4", len(warriors))
	}
	for _, id := range warriors {
		if !b16Tapped(t, g, id) {
			t.Error("the Warriors are tapped")
		}
		if c, _ := battlefieldCard(g, id); c.Power != 1 || c.Toughness != 1 || !c.HasColor("R") {
			t.Error("1/1 red Warriors")
		}
	}
	// Sacrificed at the beginning of the next end step.
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Warrior"); n != 0 {
		t.Errorf("the Warriors are sacrificed at the end step: %d left", n)
	}
	if !g.Battlefield.Contains(bear) || !g.Battlefield.Contains(elf) {
		t.Error("only the Warriors")
	}
	// The next turn is not "this turn".
	advanceToPrecombatMainOf(t, g, 0)
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Warrior"); n != 0 {
		t.Errorf("the priming ends with the turn: %d Warriors", n)
	}
}

// --- enchantments ---------------------------------------------------

func TestB33FervorGivesYourCreaturesHaste(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fervor := b12Push(g, me.ID, "Fervor", "Enchantment", b33FervorOracle, 0, 0)
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if !hasEffectiveKeyword(t, g, mine, "haste") {
		t.Error("a creature you control has haste")
	}
	if hasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("an opponent's creature does not")
	}
	if hasEffectiveKeyword(t, g, fervor, "haste") {
		t.Error("Fervor is not a creature")
	}
	fresh := castAndResolveCreature(t, g, "Hasty Bear", "Creature — Bear", "")
	if summoningSickOf(t, g, fresh) {
		t.Error("a creature that entered this turn can attack")
	}
}

func TestB33RaidBombardmentPingsForEachSmallAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Raid Bombardment", "Enchantment", b33RaidBombardmentOracle, 0, 0)
	small := b12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	two := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	big := b12Creature(g, me.ID, "Ogre", "Creature — Ogre", 3, 3)
	life := opp.Life
	declareAttack(t, g, opp.ID, small, two, big)
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("two attackers with power 2 or less: %d → %d, want -2 before damage", life, opp.Life)
	}
}

func TestB33NestOfScarabsMakesAnInsectPerMinusCounterYouPut(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Nest of Scarabs", "Enchantment", b33NestOfScarabsOracle, 0, 0)
	theirs := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	mine := b12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 3, 3)
	other := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	// Your Clasp's counter on their creature: yours.
	castCatalogSpell(t, g, "Contagion Clasp", "Artifact", b33ContagionClaspOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Insect"); n != 1 {
		t.Fatalf("one -1/-1 counter you put: %d Insects", n)
	}
	// A counter placed outside any resolution — a +1/+1 counter, or
	// an opponent's -1/-1 counter on your creature — is not yours.
	b28AddCounter(t, g, other, 1)
	passPriorityAroundTable(t, g)
	b13PlayAs(t, g, 1, "Contagion Clasp", "Artifact", b33ContagionClaspOracle)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, opp.ID)
	pickCard(t, g, opp.ID, mine)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Insect"); n != 1 {
		t.Errorf("an opponent putting the counter is not you: %d Insects", n)
	}
	if got := counterCount(g, mine, "-1/-1"); got != 1 {
		t.Fatalf("the opponent's counter landed: %d", got)
	}
}

// --- creatures ------------------------------------------------------

func TestB33CloudblazerGainsTwoAndDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life, hand := me.Life, me.Hand.Size()
	blazer := castCatalogSpell(t, g, "Cloudblazer", "Creature — Human Scout", b33CloudblazerOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2", life, me.Life)
	}
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d → %d, want +2 (the cast helper adds one, the cast removes it)", hand, got)
	}
	if !hasEffectiveKeyword(t, g, blazer, "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}
}

func TestB33GoodFortuneUnicornCountersEachOtherCreatureYouControlEntering(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	unicorn := castCatalogSpell(t, g, "Good-Fortune Unicorn", "Creature — Unicorn", b33GoodFortuneUnicornOracle, nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, unicorn, "+1/+1"); got != 0 {
		t.Errorf("the Unicorn itself is not \"another\": %d counters", got)
	}
	bear := castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, bear, "+1/+1"); got != 1 {
		t.Errorf("another creature entering gets a counter: %d", got)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1) })
	passPriorityAroundTable(t, g)
	if got := counterCount(g, findBattlefieldByName(g, "Goblin"), "+1/+1"); got != 1 {
		t.Errorf("a token counts: %d", got)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, BlackZombieToken(), 1) })
	passPriorityAroundTable(t, g)
	if got := counterCount(g, findBattlefieldByName(g, "Zombie"), "+1/+1"); got != 0 {
		t.Errorf("an opponent's creature does not: %d", got)
	}
}

func TestB33MajaPumpsOthersAndMakesAWarriorOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	maja := b12Push(g, me.ID, "Maja, Bretagard Protector", "Legendary Creature — Human Warrior", b33MajaOracle, 2, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if p := effectivePower(t, g, bear); p != 3 {
		t.Errorf("another creature you control gets +1/+1: power %d", p)
	}
	if p := effectivePower(t, g, maja); p != 2 {
		t.Errorf("Maja herself does not: power %d", p)
	}
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Human Warrior"); n != 1 {
		t.Fatalf("a land entering makes a Human Warrior: %d", n)
	}
	warrior := findBattlefieldByName(g, "Human Warrior")
	if p, tt := effectivePower(t, g, warrior), effectiveToughness(t, g, warrior); p != 2 || tt != 2 {
		t.Errorf("the 1/1 token under the anthem is %d/%d, want 2/2", p, tt)
	}
	if c, _ := battlefieldCard(g, warrior); !c.HasColor("W") || !c.HasSubtype("Human") || !c.HasSubtype("Warrior") {
		t.Error("a white Human Warrior")
	}
	// A creature entering is not landfall; an opponent's land is not yours.
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	b13PlayAs(t, g, 1, "Swamp", "Basic Land — Swamp", "")
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Human Warrior"); n != 1 {
		t.Errorf("still one Warrior, got %d", n)
	}
}

func TestB33OverseerOfTheDamnedMayDestroyOnEntryAndZombifiesOpposingDeaths(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	overseer := castCatalogSpell(t, g, "Overseer of the Damned", "Creature — Demon", b33OverseerOfTheDamnedOracle, nil)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("the chosen creature is destroyed")
	}
	if !hasEffectiveKeyword(t, g, overseer, "flying") {
		t.Error("flying")
	}
	if n := countBattlefieldNamed(g, me.ID, "Zombie"); n != 1 {
		t.Fatalf("the opponent's nontoken creature dying makes a Zombie: %d", n)
	}
	zombie := findBattlefieldByName(g, "Zombie")
	if !b16Tapped(t, g, zombie) {
		t.Error("the Zombie enters tapped")
	}
	if c, _ := battlefieldCard(g, zombie); c.Power != 2 || c.Toughness != 2 || !c.HasColor("B") {
		t.Error("a 2/2 black Zombie")
	}
	// A token dying, or your own creature dying, makes nothing.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, RedGoblinToken(), 1) })
	passPriorityAroundTable(t, g)
	b18Kill(t, g, findBattlefieldByName(g, "Goblin"))
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b18Kill(t, g, mine)
	if n := countBattlefieldNamed(g, me.ID, "Zombie"); n != 1 {
		t.Errorf("a token and your own creature are not nontoken opposing creatures: %d Zombies", n)
	}
}

func TestB33PatronOfTheVeinDestroysOnEntryThenExilesAndGrowsTheVampires(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	vampire := b12Creature(g, me.ID, "Vampire Nighthawk", "Creature — Vampire Shaman", 2, 3)
	patron := castCatalogSpell(t, g, "Patron of the Vein", "Creature — Vampire Shaman", b33PatronOfTheVeinOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if hasID(p.PickTargetCards, mine) || !hasID(p.PickTargetCards, theirs) {
		t.Error("only an opponent's creature is offered")
	}
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("the chosen creature is destroyed")
	}
	if !g.Exile.Contains(theirs) {
		t.Error("the dead creature is exiled from the graveyard")
	}
	if got := counterCount(g, vampire, "+1/+1"); got != 1 {
		t.Errorf("each Vampire you control gets a counter: the Nighthawk has %d", got)
	}
	if got := counterCount(g, patron, "+1/+1"); got != 1 {
		t.Errorf("the Patron is a Vampire too: %d", got)
	}
	if got := counterCount(g, mine, "+1/+1"); got != 0 {
		t.Errorf("a non-Vampire gets nothing: %d", got)
	}
	b18Kill(t, g, mine)
	if got := counterCount(g, vampire, "+1/+1"); got != 1 {
		t.Errorf("your own creature dying is silent: %d", got)
	}
}

func TestB33HermesMakesBirdsOnNoncreatureSpellsAndScriesWhenTheyAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Hermes, Overseer of Elpis", "Legendary Creature — Elder Wizard", b33HermesOracle, 2, 4)
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Bird"); n != 0 {
		t.Fatalf("a creature spell makes nothing: %d Birds", n)
	}
	castCatalogSpell(t, g, "Opt", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Bird"); n != 2 {
		t.Fatalf("two noncreature spells, %d Birds", n)
	}
	birds := battlefieldIDsNamed(g, "Bird")
	if c, _ := battlefieldCard(g, birds[0]); c.Power != 1 || c.Toughness != 1 || !c.HasColor("U") {
		t.Error("a 1/1 blue Bird")
	}
	assertKeywords(t, g, birds[0], "flying", "vigilance")
	// Two Birds attacking is one scry 2.
	for _, id := range birds {
		b10Awake(g, id)
	}
	declareAttack(t, g, opp.ID, birds...)
	passPriorityAroundTable(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry prompt")
	}
	if len(c.ScryCards) != 2 {
		t.Errorf("scry 2, looking at %d", len(c.ScryCards))
	}
	b26AnswerScryAllTop(t, g, me.ID)
	if scryChoiceFor(g, me.ID) != nil {
		t.Error("one or more Birds is one trigger")
	}
}

func TestB33SidisiMillsOnEntryAndAttackAndZombifiesMilledCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// seedSearchLibrary pushes each card under the previous, so the
	// first listed is on top.
	seedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Wurm", TypeLine: "Creature — Wurm"},
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp"},
		game.Card{Name: "Elf", TypeLine: "Creature — Elf"},
	)
	sidisi := castCatalogSpell(t, g, "Sidisi, Brood Tyrant", "Legendary Creature — Snake Shaman", b33SidisiOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Library.Size() != 4 {
		t.Fatalf("the entry mills three: library %d", me.Library.Size())
	}
	if n := countBattlefieldNamed(g, me.ID, "Zombie"); n != 1 {
		t.Fatalf("two creature cards in one mill is ONE Zombie: %d", n)
	}
	// The attack mills three more: a Bolt and two lands, no Zombie.
	b10Awake(g, sidisi)
	declareAttack(t, g, opp.ID, sidisi)
	passPriorityAroundTable(t, g)
	if me.Library.Size() != 1 {
		t.Fatalf("the attack mills three: library %d", me.Library.Size())
	}
	if n := countBattlefieldNamed(g, me.ID, "Zombie"); n != 1 {
		t.Errorf("no creature card milled, no Zombie: %d", n)
	}
	// A separate mill later is a separate Zombie; an opponent's mill
	// of a creature card is not your graveyard.
	g.WithWriteLock(func() { _ = g.MillNForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Zombie"); n != 2 {
		t.Errorf("the Elf milled later: %d Zombies, want 2", n)
	}
	seedSearchLibrary(opp, game.Card{Name: "Their Bear", TypeLine: "Creature — Bear"})
	g.WithWriteLock(func() { _ = g.MillNForEffect(opp.ID, 1) })
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Zombie"); n != 2 {
		t.Errorf("an opponent's mill: %d Zombies, want 2", n)
	}
}

func TestB33RakdosJoinsUpReanimatesWithTwoCountersAndPunishesLegendaryDeaths(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := b17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	b17GraveyardCard(me, "Dead Rock", "Artifact", "{1}")
	castCatalogSpell(t, g, "Rakdos Joins Up", "Legendary Enchantment", b33RakdosJoinsUpOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if len(p.PickTargetCards) != 1 || p.PickTargetCards[0] != dead {
		t.Errorf("only the creature card is offered: %v", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, dead)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) || controllerOf(t, g, dead) != me.ID {
		t.Fatal("the creature card returns under your control")
	}
	if got := counterCount(g, dead, "+1/+1"); got != 2 {
		t.Errorf("with two +1/+1 counters: %d", got)
	}
	// A legendary creature you control dying: its power at the
	// opponent you pick, counters included.
	legend := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Legend", TypeLine: "Legendary Creature — Human",
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	b28AddCounter(t, g, legend, 1)
	life := opp.Life
	b25Destroy(g, legend)
	b04WaitForPick(t, g, me.ID)
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("a 3/3 with a counter deals 4: %d → %d", life, opp.Life)
	}
	// A nonlegendary creature dying is silent.
	b18Kill(t, g, dead)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("the Bear is not legendary")
	}
}

func TestB33GodEternalBontuSacrificesTheChosenPermanentsToDrawAndReturnsThirdFromTop(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	land := b12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	bontu := castCatalogSpell(t, g, "God-Eternal Bontu", "Legendary Creature — Zombie God", b33GodEternalBontuOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if hasID(p.PickTargetCards, bontu) || hasID(p.PickTargetCards, theirs) {
		t.Error("Bontu himself and an opponent's permanent are not offered")
	}
	if !hasID(p.PickTargetCards, rock) || !hasID(p.PickTargetCards, land) || !hasID(p.PickTargetCards, bear) {
		t.Error("every other permanent you control is offered")
	}
	if p.PickTargetMin != 0 || p.PickTargetMax != 0 {
		t.Errorf("any number: min %d max %d", p.PickTargetMin, p.PickTargetMax)
	}
	hand := me.Hand.Size()
	b17PickCards(t, g, me.ID, rock, land)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || g.Battlefield.Contains(land) {
		t.Error("the chosen permanents are sacrificed")
	}
	if !g.Battlefield.Contains(bear) || !g.Battlefield.Contains(bontu) {
		t.Error("the unchosen stay")
	}
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("then draw that many: hand %d → %d", hand, got)
	}
	if !hasEffectiveKeyword(t, g, bontu, "menace") {
		t.Error("menace")
	}
	// Dies: may go third from the top.
	b25Destroy(g, bontu)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(bontu) {
		t.Fatal("Bontu should have left the graveyard")
	}
	names := libraryTopNames(me, 3)
	if len(names) != 3 || names[2] != "God-Eternal Bontu" || names[0] == "God-Eternal Bontu" {
		t.Errorf("third from the top, got %v", names)
	}
}

func TestB33LathielDealsTheLifeGainedOutAsCountersAtTheEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lathiel := b12Push(g, me.ID, "Lathiel, the Bounteous Dawn", "Legendary Creature — Unicorn", b33LathielOracle, 2, 2)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if !hasEffectiveKeyword(t, g, lathiel, "lifelink") {
		t.Error("lifelink")
	}
	// No life gained: no trigger at the end step.
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("no life gained this turn, no trigger")
	}
	// Three life on an opponent's turn: three counters around the
	// two creatures picked — two on the first, one on the second.
	advanceToMainOf(t, g, 1)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	advanceToEndStepOf(t, g, 1)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if hasID(p.PickTargetCards, lathiel) {
		t.Error("Lathiel herself is not \"other\"")
	}
	if !hasID(p.PickTargetCards, a) || !hasID(p.PickTargetCards, b) {
		t.Error("any creature, an opponent's included, is offered")
	}
	b17PickCards(t, g, me.ID, a, b)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, a, "+1/+1"); got != 2 {
		t.Errorf("the first picked gets the remainder: %d, want 2", got)
	}
	if got := counterCount(g, b, "+1/+1"); got != 1 {
		t.Errorf("the second picked: %d, want 1", got)
	}
	// One creature picked gets every counter.
	advanceToMainOf(t, g, 2)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2) })
	advanceToEndStepOf(t, g, 2)
	b04WaitForPick(t, g, me.ID)
	b17PickCards(t, g, me.ID, a)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, a, "+1/+1"); got != 4 {
		t.Errorf("two more on the one creature picked: %d, want 4", got)
	}
}

func TestB33AlelaMakesAFaerieOnYourFirstSpellOfAnOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Alela, Cunning Conqueror", "Legendary Creature — Faerie Warlock", b33AlelaOracle, 2, 4)
	castCatalogSpell(t, g, "Shock", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Faerie Rogue"); n != 0 {
		t.Fatalf("your own turn: %d Faeries", n)
	}
	advanceToNextSeatsTurn(t, g)
	advanceTo(t, g, game.StepPrecombatMain)
	b22CastInstantAs(t, g, me, "Opt")
	passPriorityAroundTable(t, g)
	b22CastInstantAs(t, g, me, "Brainstorm")
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Faerie Rogue"); n != 1 {
		t.Fatalf("the first spell of an opponent's turn, once: %d Faeries", n)
	}
	faerie := findBattlefieldByName(g, "Faerie Rogue")
	assertKeywords(t, g, faerie, "flying")
	if c, _ := battlefieldCard(g, faerie); c.Power != 1 || c.Toughness != 1 || !c.HasSubtype("Faerie") {
		t.Error("a 1/1 Faerie Rogue")
	}
}

func TestB33AlelaGoadsACreatureOfThePlayerHerFaeriesHit(t *testing.T) {
	g := newCatalogGame(t)
	me, victim, other := g.Seats[0], g.Seats[1], g.Seats[2]
	alela := b12Push(g, me.ID, "Alela, Cunning Conqueror", "Legendary Creature — Faerie Warlock", b33AlelaOracle, 2, 4)
	sprite := b12Creature(g, me.ID, "Sprite", "Creature — Faerie", 1, 1)
	theirs := b12Creature(g, victim.ID, "Their Bear", "Creature — Bear", 2, 2)
	otherBear := b12Creature(g, other.ID, "Other Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	attackWith(t, g, victim.ID, alela, sprite)
	if victim.Life != 40-3 {
		t.Fatalf("two Faeries connected for 3: life %d", victim.Life)
	}
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if !hasID(p.PickTargetCards, theirs) {
		t.Error("the hit player's creature is offered")
	}
	if hasID(p.PickTargetCards, otherBear) || hasID(p.PickTargetCards, mine) {
		t.Error("another player's creature and your own are not offered")
	}
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("one or more Faeries is one trigger")
	}
	if b33Goaded(t, g, theirs) != me.ID {
		t.Fatal("the chosen creature is goaded by you")
	}
	// Until your next turn: still goaded through the opponents'
	// turns, cleared at your upkeep.
	advanceToUpkeepOf(t, g, 2)
	passPriorityAroundTable(t, g)
	if b33Goaded(t, g, theirs) != me.ID {
		t.Error("still goaded on another opponent's turn")
	}
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if b33Goaded(t, g, theirs) != uuid.Nil {
		t.Error("the goad ends at the beginning of your next turn")
	}
	// A non-Faerie connecting does not goad.
	attackWith(t, g, victim.ID, mine)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a Bear is not a Faerie")
	}
}
