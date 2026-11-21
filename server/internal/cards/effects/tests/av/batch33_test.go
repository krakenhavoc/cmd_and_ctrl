package testsav

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
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

// b33PoolCount counts the tokens of one colour in a player's pool.
func b33PoolCount(p *game.Player, color string) int {
	n := 0
	for _, tok := range p.ManaPool.Units() {
		if tok.Color == color {
			n++
		}
	}
	return n
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Copper Myr is
// a table row, so a transposed row is invisible until someone plays
// that exact card.
func TestBatch33CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B33VisionsOfBeyondOracle:     "Visions of Beyond",
		effectstest.B33AlelaOracle:               "Alela, Cunning Conqueror",
		effectstest.B33CoalitionRelicOracle:      "Coalition Relic",
		effectstest.B33BurningInquiryOracle:      "Burning Inquiry",
		effectstest.B33LathielOracle:             "Lathiel, the Bounteous Dawn",
		effectstest.B33RaidBombardmentOracle:     "Raid Bombardment",
		effectstest.B33BlastingStationOracle:     "Blasting Station",
		effectstest.B33MajaOracle:                "Maja, Bretagard Protector",
		effectstest.B33DalkovanEncampmentOracle:  "Dalkovan Encampment",
		effectstest.B33CountersquallOracle:       "Countersquall",
		effectstest.B33NestOfScarabsOracle:       "Nest of Scarabs",
		effectstest.B33RecklessBarbarianOracle:   "Reckless Barbarian",
		effectstest.B33CloudblazerOracle:         "Cloudblazer",
		effectstest.B33RakdosJoinsUpOracle:       "Rakdos Joins Up",
		effectstest.B33GoodFortuneUnicornOracle:  "Good-Fortune Unicorn",
		effectstest.B33OverseerOfTheDamnedOracle: "Overseer of the Damned",
		effectstest.B33PatronOfTheVeinOracle:     "Patron of the Vein",
		effectstest.B33FervorOracle:              "Fervor",
		effectstest.B33HermesOracle:              "Hermes, Overseer of Elpis",
		effectstest.B33SidisiOracle:              "Sidisi, Brood Tyrant",
		effectstest.B33FlowOfKnowledgeOracle:     "Flow of Knowledge",
		effectstest.B33CopperMyrOracle:           "Copper Myr",
		effectstest.B33PlanarEngineeringOracle:   "Planar Engineering",
		effectstest.B33DoublingCubeOracle:        "Doubling Cube",
		effectstest.B33StoneOfErechOracle:        "Stone of Erech",
		effectstest.B33GodEternalBontuOracle:     "God-Eternal Bontu",
	}
	if len(want) != 26 {
		t.Fatalf("the batch ships 26 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Visions of Beyond", "Instant", effectstest.B33VisionsOfBeyondOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("no full graveyard: hand %d → %d, want one drawn (the cast helper adds one, the cast removes it)", hand, got)
	}
	for i := 0; i < 19; i++ {
		effectstest.PushGraveyardCardForTest(opp, "Filler")
	}
	hand = me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Visions of Beyond", "Instant", effectstest.B33VisionsOfBeyondOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("nineteen cards is not twenty: hand %d → %d", hand, got)
	}
	effectstest.PushGraveyardCardForTest(opp, "Filler")
	hand = me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Visions of Beyond", "Instant", effectstest.B33VisionsOfBeyondOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("an opponent's twenty-card graveyard: hand %d → %d, want three drawn", hand, got)
	}
}

func TestB33BurningInquiryEveryPlayerDrawsThreeThenDiscardsThreeAtRandom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, p := range g.Seats {
		effectstest.FillHandTo(t, g, p, 4)
	}
	hands, graveyards := b33Hands(g), b33Graveyards(g)
	// The cast helper puts the spell in hand and the cast removes it,
	// so every seat's hand — the caster's included — ends where it
	// started: three drawn, three discarded.
	effectstest.CastCatalogSpell(t, g, "Burning Inquiry", "Sorcery", effectstest.B33BurningInquiryOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	effectstest.B12Permanent(g, me.ID, "Hallowed Fountain", "Land — Plains Island")
	effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Flow of Knowledge", "Instant", effectstest.B33FlowOfKnowledgeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("three Islands you control (one nonbasic): hand %d → %d, want +3", hand, got)
	}
	if got := effectstest.DiscardOwed(g, me.ID); got != 2 {
		t.Errorf("then discard two — %d discards pending", got)
	}
}

func TestB33CountersquallCountersANoncreatureSpellAndItsControllerLosesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	if err := b09TryCast(t, g, "Countersquall", "Instant", effectstest.B33CountersquallOracle, effectstest.B16TargetCard(bear)); err == nil {
		t.Fatal("a creature spell is not a legal target")
	}
	effectstest.PassPriorityAroundTable(t, g)
	bolt := effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Countersquall", "Instant", effectstest.B33CountersquallOracle, effectstest.B16TargetCard(bolt))
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	swamp := effectstest.B12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	forest := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	crypt := effectstest.B12Permanent(g, me.ID, "Blood Crypt", "Land — Swamp Mountain")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"},
		game.Card{Name: "Snow Swamp", TypeLine: "Basic Snow Land — Swamp"},
		game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Hallowed Fountain", TypeLine: "Land — Plains Island"},
	)
	effectstest.CastCatalogSpell(t, g, "Planar Engineering", "Sorcery", effectstest.B33PlanarEngineeringOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SacrificeChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no sacrifice prompt")
	}
	if effectstest.HasID(c.SacrificeOptions, bear) || !effectstest.HasID(c.SacrificeOptions, crypt) {
		t.Error("only lands are offered, nonbasics included")
	}
	effectstest.AnswerSacrifice(t, g, me.ID, swamp)
	effectstest.AnswerSacrifice(t, g, me.ID, crypt)
	if effectstest.SacrificeChoiceFor(g, me.ID) != nil {
		t.Error("two lands, not three")
	}
	if !g.Battlefield.Contains(forest) || !g.Battlefield.Contains(bear) {
		t.Error("the unchosen land and the creature stay")
	}
	s := effectstest.SearchChoiceFor(g, me.ID)
	if s == nil {
		t.Fatal("no search prompt")
	}
	if s.SearchMax != 4 || len(s.SearchCards) != 5 {
		t.Errorf("five basics offered (the snow basic included, the shockland not), up to four taken: %d offered, max %d", len(s.SearchCards), s.SearchMax)
	}
	effectstest.AnswerSearchByID(t, g, me.ID, ids[:4]...)
	for _, id := range ids[:4] {
		if !g.Battlefield.Contains(id) {
			t.Fatal("every chosen basic is put onto the battlefield")
		}
		if !effectstest.B16Tapped(t, g, id) {
			t.Error("the basics enter tapped")
		}
	}
	if me.Library.Size() != 2 {
		t.Errorf("library %d, want 2 (an Island and the shockland)", me.Library.Size())
	}
}

// --- artifacts ------------------------------------------------------

func TestB33CoalitionRelicBanksChargeCountersForYourFirstMainPhase(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	relic := effectstest.PushCatalogPermanent(g, me.ID, "Coalition Relic", "Artifact", effectstest.B33CoalitionRelicOracle, false)
	effectstest.AdvanceToMain(t, g)
	// Entering the main phase fires the trigger with no counters on
	// the Relic: it resolves and adds nothing.
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 || me.ManaPool.Total() != 0 {
		t.Fatal("no counters, no mana")
	}
	if err := g.ActivateManaAbility(me.ID, relic, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the mana ability: %v", err)
	}
	if pick := effectstest.RiderLatestManaPick(g, me.ID); pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("one mana of any colour, got %+v", pick)
	}
	effectstest.B10ResolveAllManaPicks(t, g, me.ID, "U")
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(relic) })
	effectstest.B16Activate(t, g, me.ID, relic, 0, game.ActivateAbilityParams{})
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(relic) })
	effectstest.B16Activate(t, g, me.ID, relic, 0, game.ActivateAbilityParams{})
	if got := effectstest.CounterCount(g, relic, "charge"); got != 2 {
		t.Fatalf("two activations, %d charge counters", got)
	}
	// The pool empties between steps; the trigger fires at the
	// start of the controller's next precombat main.
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	if me.ManaPool.Total() != 0 {
		t.Fatal("the pool is empty as the main phase begins")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, relic, "charge"); got != 0 {
		t.Errorf("the counters come off: %d left", got)
	}
	if n := effectstest.B10ResolveAllManaPicks(t, g, me.ID, "R"); n != 2 {
		t.Errorf("one colour pick per counter: %d", n)
	}
	if got := b33PoolCount(me, "R"); got != 2 {
		t.Errorf("two red mana in the pool, got %d", got)
	}
	// An opponent's main phase is silent; with no counters, so is yours.
	effectstest.AdvanceToPrecombatMainOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 {
		t.Error("an opponent's main phase is not yours")
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 || me.ManaPool.Total() != 0 {
		t.Error("no counters, no mana")
	}
}

func TestB33BlastingStationPingsForASacrificeAndMayUntapWhenACreatureEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	station := effectstest.PushCatalogPermanent(g, me.ID, "Blasting Station", "Artifact", effectstest.B33BlastingStationOracle, false)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	effectstest.B16Activate(t, g, me.ID, station, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetPlayer(opp.ID), SacrificeIDs: []uuid.UUID{fodder}})
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrifice is the cost")
	}
	if !effectstest.B16Tapped(t, g, station) {
		t.Error("the tap is the cost")
	}
	if opp.Life != life-1 {
		t.Errorf("1 damage to the target: %d → %d", life, opp.Life)
	}
	// A creature entering — anyone's — offers the untap.
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, station) {
		t.Error("the Station untaps")
	}
	// Declined, it stays tapped; an artifact entering offers nothing.
	effectstest.B16Tap(g, station)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.RedGoblinToken(), 1) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, station) {
		t.Error("declining leaves it tapped")
	}
	effectstest.CastCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OwedMays(g, me.ID) > 0 {
		t.Error("an artifact entering is not a creature")
	}
	_ = theirs
}

func TestB33DoublingCubeDoublesWhatIsLeftAfterPayingThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cube := effectstest.PushCatalogPermanent(g, me.ID, "Doubling Cube", "Artifact", effectstest.B33DoublingCubeOracle, false)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, cube, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("the {3} is a cost — nothing to pay it with")
	}
	effectstest.B06AddMana(me, "C", "C", "C", "R", "R", "G")
	if err := g.ActivateManaAbility(me.ID, cube, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if !effectstest.B16Tapped(t, g, cube) {
		t.Error("the tap is the cost")
	}
	if got := me.ManaPool.Total(); got != 6 {
		t.Fatalf("three paid, three left, three added: pool %v", effectstest.PoolColors(me))
	}
	if r, gr, c := b33PoolCount(me, "R"), b33PoolCount(me, "G"), b33PoolCount(me, "C"); r != 4 || gr != 2 || c != 0 {
		t.Errorf("the colourless paid the {3} and the rest doubled: pool %v", effectstest.PoolColors(me))
	}
	// An empty pool after paying adds nothing.
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(cube) })
	me.ManaPool = nil
	effectstest.B06AddMana(me, "C", "C", "C")
	if err := g.ActivateManaAbility(me.ID, cube, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("nothing left to double: pool %v", effectstest.PoolColors(me))
	}
}

func TestB33StoneOfErechExilesOpposingDeathsAndEatsAGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stone := effectstest.PushCatalogPermanent(g, me.ID, "Stone of Erech", "Legendary Artifact", effectstest.B33StoneOfErechOracle, false)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.B25Destroy(g, theirs)
	effectstest.B25Destroy(g, mine)
	effectstest.B25Destroy(g, rock)
	if !g.Exile.Contains(theirs) {
		t.Error("an opponent's creature dying is exiled instead")
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("your own creature dies as normal")
	}
	if !opp.Graveyard.Contains(rock) {
		t.Error("an opponent's noncreature dies as normal")
	}
	old := effectstest.PushGraveyardCardForTest(opp, "Old Card")
	yard := opp.Graveyard.Size()
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C")
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, stone, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetPlayer(opp.ID)})
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

// #1337: the same stale caveat Liesa's file carried, closed the same
// way — CR 903.9a's command-zone move is a state-based "may" AFTER
// the exile, so an opponent's dying commander is exiled like any
// other of their creatures and its owner separately decides. Since
// ADR 0115 that is literally the engine's order: no CR 616 ordering
// prompt, the Stone's replacement exiles it, then CR 903.9a asks.
func TestB33StoneOfErechExilesAnOpponentsDyingCommanderThenAsksItsOwner(t *testing.T) {
	for _, takeCommandZone := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.PushCatalogPermanent(g, me.ID, "Stone of Erech", "Legendary Artifact", effectstest.B33StoneOfErechOracle, false)
		commander := uuid.New()
		g.Battlefield.PushTop(game.Card{
			InstanceID: commander, Name: "Their Commander", TypeLine: "Legendary Creature — Human Wizard",
			Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID, IsCommander: true,
		})

		g.WithWriteLock(func() {
			if err := g.DestroyPermanentForEffect(commander); err != nil {
				t.Fatalf("DestroyPermanentForEffect: %v", err)
			}
		})
		for _, c := range g.PendingChoices {
			if c != nil && (c.Kind == game.PendingChoiceReplacementOrder || c.Kind == game.PendingChoiceOptionalReplacement) {
				t.Fatalf("unexpected %s prompt: only the Stone's replacement applies to a dying commander", c.Kind)
			}
		}
		if opp.Graveyard.Contains(commander) {
			t.Fatal("the commander hit the graveyard under Stone of Erech")
		}
		if !effectstest.ExileHas(g, commander) {
			t.Fatal("the Stone did not exile the dying commander")
		}
		if effectstest.CommanderReturnPromptFor(g, opp.ID) == nil {
			t.Fatal("the commander's owner was never offered the command zone")
		}
		effectstest.AnswerCommanderReturn(t, g, opp.ID, takeCommandZone)
		if takeCommandZone {
			if !opp.Command.Contains(commander) {
				t.Fatal("the owner took the command zone but the commander is not there")
			}
			continue
		}
		if !effectstest.ExileHas(g, commander) {
			t.Fatal("declined commander is not in exile")
		}
	}
}

// --- lands and mana -------------------------------------------------

func TestB33CopperMyrTapsForGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	myr := effectstest.PushCatalogPermanent(g, me.ID, "Copper Myr", "Artifact Creature — Myr", effectstest.B33CopperMyrOracle, false)
	if err := g.ActivateManaAbility(me.ID, myr, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool %v, want [G]", got)
	}
}

func TestB33RecklessBarbarianSacrificesForTwoRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	barbarian := effectstest.PushCatalogPermanent(g, me.ID, "Reckless Barbarian", "Creature — Dragon Barbarian", effectstest.B33RecklessBarbarianOracle, true)
	if err := g.ActivateManaAbility(me.ID, barbarian, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("no tap in the cost, so summoning sickness does not apply: %v", err)
	}
	if g.Battlefield.Contains(barbarian) || !me.Graveyard.Contains(barbarian) {
		t.Error("the Barbarian is sacrificed")
	}
	if got := effectstest.PoolColors(me); len(got) != 2 || got[0] != "R" || got[1] != "R" {
		t.Errorf("pool %v, want [R R]", got)
	}
}

func TestB33DalkovanEncampmentEntersTappedWithoutASwampOrMountain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	land := effectstest.B12PlayFromHand(t, g, "Dalkovan Encampment", "Land", effectstest.B33DalkovanEncampmentOracle, game.CastSpellParams{})
	if !effectstest.B16Tapped(t, g, land) {
		t.Error("with no Swamp or Mountain it enters tapped")
	}
	g2 := effectstest.NewCatalogGame(t)
	effectstest.B12Permanent(g2, g2.Seats[0].ID, "Blood Crypt", "Land — Swamp Mountain")
	land2 := effectstest.B12PlayFromHand(t, g2, "Dalkovan Encampment", "Land", effectstest.B33DalkovanEncampmentOracle, game.CastSpellParams{})
	if effectstest.B16Tapped(t, g2, land2) {
		t.Error("with a Mountain it enters untapped")
	}
	if err := g2.ActivateManaAbility(g2.Seats[0].ID, land2, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.PoolColors(g2.Seats[0]); len(got) != 1 || got[0] != "W" {
		t.Errorf("pool %v, want [W]", got)
	}
}

func TestB33DalkovanEncampmentPrimedMakesTwoAttackingWarriorsPerActivation(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.PushCatalogPermanent(g, me.ID, "Dalkovan Encampment", "Land", effectstest.B33DalkovanEncampmentOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	elf := effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	// Unprimed, an attack makes nothing.
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Warrior"); n != 0 {
		t.Fatalf("no activation this turn: %d Warriors", n)
	}
	// Primed twice on the next turn: two attackers, one trigger,
	// four Warriors.
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.B06AddMana(me, "W", "C", "C")
	effectstest.B16Activate(t, g, me.ID, land, 0, game.ActivateAbilityParams{})
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	effectstest.B06AddMana(me, "W", "C", "C")
	effectstest.B16Activate(t, g, me.ID, land, 0, game.ActivateAbilityParams{})
	effectstest.DeclareAttack(t, g, opp.ID, bear, elf)
	effectstest.PassAndSettleAttackSplit(t, g, me.ID)
	warriors := effectstest.B33AttackingTokensNamed(g, me.ID, opp.ID, "Warrior")
	if len(warriors) != 4 {
		t.Fatalf("two activations, two attackers: %d Warriors tapped and attacking, want 4", len(warriors))
	}
	for _, id := range warriors {
		if !effectstest.B16Tapped(t, g, id) {
			t.Error("the Warriors are tapped")
		}
		if c, _ := effectstest.BattlefieldCard(g, id); c.Power != 1 || c.Toughness != 1 || !c.HasColor("R") {
			t.Error("1/1 red Warriors")
		}
	}
	// Sacrificed at the beginning of the next end step.
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Warrior"); n != 0 {
		t.Errorf("the Warriors are sacrificed at the end step: %d left", n)
	}
	if !g.Battlefield.Contains(bear) || !g.Battlefield.Contains(elf) {
		t.Error("only the Warriors")
	}
	// The next turn is not "this turn".
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Warrior"); n != 0 {
		t.Errorf("the priming ends with the turn: %d Warriors", n)
	}
}

// --- enchantments ---------------------------------------------------

func TestB33FervorGivesYourCreaturesHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fervor := effectstest.B12Push(g, me.ID, "Fervor", "Enchantment", effectstest.B33FervorOracle, 0, 0)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, mine, "haste") {
		t.Error("a creature you control has haste")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("an opponent's creature does not")
	}
	if effectstest.HasEffectiveKeyword(t, g, fervor, "haste") {
		t.Error("Fervor is not a creature")
	}
	fresh := effectstest.CastAndResolveCreature(t, g, "Hasty Bear", "Creature — Bear", "")
	if effectstest.SummoningSickOf(t, g, fresh) {
		t.Error("a creature that entered this turn can attack")
	}
}

func TestB33RaidBombardmentPingsForEachSmallAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Raid Bombardment", "Enchantment", effectstest.B33RaidBombardmentOracle, 0, 0)
	small := effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	two := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, me.ID, "Ogre", "Creature — Ogre", 3, 3)
	life := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, small, two, big)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("two attackers with power 2 or less: %d → %d, want -2 before damage", life, opp.Life)
	}
}

func TestB33NestOfScarabsMakesAnInsectPerMinusCounterYouPut(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Nest of Scarabs", "Enchantment", effectstest.B33NestOfScarabsOracle, 0, 0)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 3, 3)
	other := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	// Your Clasp's counter on their creature: yours.
	effectstest.CastCatalogSpell(t, g, "Contagion Clasp", "Artifact", effectstest.B33ContagionClaspOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Insect"); n != 1 {
		t.Fatalf("one -1/-1 counter you put: %d Insects", n)
	}
	// A counter placed outside any resolution — a +1/+1 counter, or
	// an opponent's -1/-1 counter on your creature — is not yours.
	effectstest.B28AddCounter(t, g, other, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B13PlayAs(t, g, 1, "Contagion Clasp", "Artifact", effectstest.B33ContagionClaspOracle)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, opp.ID)
	effectstest.PickCard(t, g, opp.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Insect"); n != 1 {
		t.Errorf("an opponent putting the counter is not you: %d Insects", n)
	}
	if got := effectstest.CounterCount(g, mine, "-1/-1"); got != 1 {
		t.Fatalf("the opponent's counter landed: %d", got)
	}
}

// --- creatures ------------------------------------------------------

func TestB33CloudblazerGainsTwoAndDrawsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life, hand := me.Life, me.Hand.Size()
	blazer := effectstest.CastCatalogSpell(t, g, "Cloudblazer", "Creature — Human Scout", effectstest.B33CloudblazerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2", life, me.Life)
	}
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d → %d, want +2 (the cast helper adds one, the cast removes it)", hand, got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, blazer, "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}
}

func TestB33GoodFortuneUnicornCountersEachOtherCreatureYouControlEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	unicorn := effectstest.CastCatalogSpell(t, g, "Good-Fortune Unicorn", "Creature — Unicorn", effectstest.B33GoodFortuneUnicornOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, unicorn, "+1/+1"); got != 0 {
		t.Errorf("the Unicorn itself is not \"another\": %d counters", got)
	}
	bear := effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, "+1/+1"); got != 1 {
		t.Errorf("another creature entering gets a counter: %d", got)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, effectstest.FindBattlefieldByName(g, "Goblin"), "+1/+1"); got != 1 {
		t.Errorf("a token counts: %d", got)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.BlackZombieToken(), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, effectstest.FindBattlefieldByName(g, "Zombie"), "+1/+1"); got != 0 {
		t.Errorf("an opponent's creature does not: %d", got)
	}
}

func TestB33MajaPumpsOthersAndMakesAWarriorOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	maja := effectstest.B12Push(g, me.ID, "Maja, Bretagard Protector", "Legendary Creature — Human Warrior", effectstest.B33MajaOracle, 2, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if p := effectstest.EffectivePower(t, g, bear); p != 3 {
		t.Errorf("another creature you control gets +1/+1: power %d", p)
	}
	if p := effectstest.EffectivePower(t, g, maja); p != 2 {
		t.Errorf("Maja herself does not: power %d", p)
	}
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Human Warrior"); n != 1 {
		t.Fatalf("a land entering makes a Human Warrior: %d", n)
	}
	warrior := effectstest.FindBattlefieldByName(g, "Human Warrior")
	if p, tt := effectstest.EffectivePower(t, g, warrior), effectstest.EffectiveToughness(t, g, warrior); p != 2 || tt != 2 {
		t.Errorf("the 1/1 token under the anthem is %d/%d, want 2/2", p, tt)
	}
	if c, _ := effectstest.BattlefieldCard(g, warrior); !c.HasColor("W") || !c.HasSubtype("Human") || !c.HasSubtype("Warrior") {
		t.Error("a white Human Warrior")
	}
	// A creature entering is not landfall; an opponent's land is not yours.
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B13PlayAs(t, g, 1, "Swamp", "Basic Land — Swamp", "")
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Human Warrior"); n != 1 {
		t.Errorf("still one Warrior, got %d", n)
	}
}

func TestB33OverseerOfTheDamnedMayDestroyOnEntryAndZombifiesOpposingDeaths(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	overseer := effectstest.CastCatalogSpell(t, g, "Overseer of the Damned", "Creature — Demon", effectstest.B33OverseerOfTheDamnedOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("the chosen creature is destroyed")
	}
	if !effectstest.HasEffectiveKeyword(t, g, overseer, "flying") {
		t.Error("flying")
	}
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Zombie"); n != 1 {
		t.Fatalf("the opponent's nontoken creature dying makes a Zombie: %d", n)
	}
	zombie := effectstest.FindBattlefieldByName(g, "Zombie")
	if !effectstest.B16Tapped(t, g, zombie) {
		t.Error("the Zombie enters tapped")
	}
	if c, _ := effectstest.BattlefieldCard(g, zombie); c.Power != 2 || c.Toughness != 2 || !c.HasColor("B") {
		t.Error("a 2/2 black Zombie")
	}
	// A token dying, or your own creature dying, makes nothing.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.RedGoblinToken(), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B18Kill(t, g, effectstest.FindBattlefieldByName(g, "Goblin"))
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.B18Kill(t, g, mine)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Zombie"); n != 1 {
		t.Errorf("a token and your own creature are not nontoken opposing creatures: %d Zombies", n)
	}
}

func TestB33PatronOfTheVeinDestroysOnEntryThenExilesAndGrowsTheVampires(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	vampire := effectstest.B12Creature(g, me.ID, "Vampire Nighthawk", "Creature — Vampire Shaman", 2, 3)
	patron := effectstest.CastCatalogSpell(t, g, "Patron of the Vein", "Creature — Vampire Shaman", effectstest.B33PatronOfTheVeinOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if effectstest.HasID(p.PickTargetCards, mine) || !effectstest.HasID(p.PickTargetCards, theirs) {
		t.Error("only an opponent's creature is offered")
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("the chosen creature is destroyed")
	}
	if !g.Exile.Contains(theirs) {
		t.Error("the dead creature is exiled from the graveyard")
	}
	if got := effectstest.CounterCount(g, vampire, "+1/+1"); got != 1 {
		t.Errorf("each Vampire you control gets a counter: the Nighthawk has %d", got)
	}
	if got := effectstest.CounterCount(g, patron, "+1/+1"); got != 1 {
		t.Errorf("the Patron is a Vampire too: %d", got)
	}
	if got := effectstest.CounterCount(g, mine, "+1/+1"); got != 0 {
		t.Errorf("a non-Vampire gets nothing: %d", got)
	}
	effectstest.B18Kill(t, g, mine)
	if got := effectstest.CounterCount(g, vampire, "+1/+1"); got != 1 {
		t.Errorf("your own creature dying is silent: %d", got)
	}
}

func TestB33HermesMakesBirdsOnNoncreatureSpellsAndScriesWhenTheyAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Hermes, Overseer of Elpis", "Legendary Creature — Elder Wizard", effectstest.B33HermesOracle, 2, 4)
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Bird"); n != 0 {
		t.Fatalf("a creature spell makes nothing: %d Birds", n)
	}
	effectstest.CastCatalogSpell(t, g, "Opt", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Bird"); n != 2 {
		t.Fatalf("two noncreature spells, %d Birds", n)
	}
	birds := effectstest.BattlefieldIDsNamed(g, "Bird")
	if c, _ := effectstest.BattlefieldCard(g, birds[0]); c.Power != 1 || c.Toughness != 1 || !c.HasColor("U") {
		t.Error("a 1/1 blue Bird")
	}
	effectstest.AssertKeywords(t, g, birds[0], "flying", "vigilance")
	// Two Birds attacking is one scry 2.
	for _, id := range birds {
		effectstest.B10Awake(g, id)
	}
	effectstest.DeclareAttack(t, g, opp.ID, birds...)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry prompt")
	}
	if len(c.ScryCards) != 2 {
		t.Errorf("scry 2, looking at %d", len(c.ScryCards))
	}
	effectstest.B26AnswerScryAllTop(t, g, me.ID)
	if effectstest.ScryChoiceFor(g, me.ID) != nil {
		t.Error("one or more Birds is one trigger")
	}
}

func TestB33SidisiMillsOnEntryAndAttackAndZombifiesMilledCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// seedSearchLibrary pushes each card under the previous, so the
	// first listed is on top.
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Wurm", TypeLine: "Creature — Wurm"},
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp"},
		game.Card{Name: "Elf", TypeLine: "Creature — Elf"},
	)
	sidisi := effectstest.CastCatalogSpell(t, g, "Sidisi, Brood Tyrant", "Legendary Creature — Snake Shaman", effectstest.B33SidisiOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != 4 {
		t.Fatalf("the entry mills three: library %d", me.Library.Size())
	}
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Zombie"); n != 1 {
		t.Fatalf("two creature cards in one mill is ONE Zombie: %d", n)
	}
	// The attack mills three more: a Bolt and two lands, no Zombie.
	effectstest.B10Awake(g, sidisi)
	effectstest.DeclareAttack(t, g, opp.ID, sidisi)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != 1 {
		t.Fatalf("the attack mills three: library %d", me.Library.Size())
	}
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Zombie"); n != 1 {
		t.Errorf("no creature card milled, no Zombie: %d", n)
	}
	// A separate mill later is a separate Zombie; an opponent's mill
	// of a creature card is not your graveyard.
	g.WithWriteLock(func() { _ = g.MillNForEffect(me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Zombie"); n != 2 {
		t.Errorf("the Elf milled later: %d Zombies, want 2", n)
	}
	effectstest.SeedSearchLibrary(opp, game.Card{Name: "Their Bear", TypeLine: "Creature — Bear"})
	g.WithWriteLock(func() { _ = g.MillNForEffect(opp.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Zombie"); n != 2 {
		t.Errorf("an opponent's mill: %d Zombies, want 2", n)
	}
}

func TestB33RakdosJoinsUpReanimatesWithTwoCountersAndPunishesLegendaryDeaths(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := effectstest.B17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	effectstest.B17GraveyardCard(me, "Dead Rock", "Artifact", "{1}")
	effectstest.CastCatalogSpell(t, g, "Rakdos Joins Up", "Legendary Enchantment", effectstest.B33RakdosJoinsUpOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if len(p.PickTargetCards) != 1 || p.PickTargetCards[0] != dead {
		t.Errorf("only the creature card is offered: %v", p.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, dead)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) || effectstest.ControllerOf(t, g, dead) != me.ID {
		t.Fatal("the creature card returns under your control")
	}
	if got := effectstest.CounterCount(g, dead, "+1/+1"); got != 2 {
		t.Errorf("with two +1/+1 counters: %d", got)
	}
	// A legendary creature you control dying: its power at the
	// opponent you pick, counters included.
	legend := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Legend", TypeLine: "Legendary Creature — Human",
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	effectstest.B28AddCounter(t, g, legend, 1)
	life := opp.Life
	effectstest.B25Destroy(g, legend)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("a 3/3 with a counter deals 4: %d → %d", life, opp.Life)
	}
	// A nonlegendary creature dying is silent.
	effectstest.B18Kill(t, g, dead)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("the Bear is not legendary")
	}
}

// TestB33RakdosJoinsUpDamageIncludesAnthemBonusOnTheDeadLegend closes
// the caveat #1565 audited: the dead legend's power now comes from
// #1379's resolution-time LKI (ctx.TriggeringPermanent,
// PowerForComparison), which folds in a static bonus from another
// permanent — not just printed power and counters, as the old
// b17LastKnownPowerOffBattlefield-only read did.
func TestB33RakdosJoinsUpDamageIncludesAnthemBonusOnTheDeadLegend(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastCatalogSpell(t, g, "Rakdos Joins Up", "Legendary Enchantment", effectstest.B33RakdosJoinsUpOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Glorious Anthem", TypeLine: "Enchantment",
		OracleID: effectstest.GloriousAnthemOracle, Owner: me.ID, Controller: me.ID,
	})
	legend := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Legend", TypeLine: "Legendary Creature — Human",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	if got := effectstest.EffectivePower(t, g, legend); got != 3 {
		t.Fatalf("effective power before death = %d, want 3 (the anthem's +1/+1 applied)", got)
	}

	life := opp.Life
	effectstest.B25Destroy(g, legend)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("a 2/2 with an anthem's +1/+1 should deal 3, not just its printed 2: %d → %d", life, opp.Life)
	}
}

// #1337: the sacrifice is a resolution-time choice (ChoosePermanents,
// Sacrifice mode), not a target picked when the trigger goes on the
// stack, so the prompt is an own_permanents choose-cards pick rather
// than a pick_target one, and it only appears once the trigger
// resolves.
func TestB33GodEternalBontuSacrificesTheChosenPermanentsToDrawAndReturnsThirdFromTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	land := effectstest.B12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	bontu := effectstest.CastCatalogSpell(t, g, "God-Eternal Bontu", "Legendary Creature — Zombie God", effectstest.B33GodEternalBontuOracle, nil)
	if effectstest.PendingOfKind(g, game.PendingChoicePickTarget) != nil {
		t.Fatal("the sacrifice asked for a target; the printed text names none")
	}
	effectstest.PassPriorityAroundTable(t, g)

	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil {
		t.Fatal("God-Eternal Bontu queued no sacrifice pick")
	}
	if effectstest.HasID(p.ChooseCards, bontu) || effectstest.HasID(p.ChooseCards, theirs) {
		t.Error("Bontu himself and an opponent's permanent are not offered")
	}
	if !effectstest.HasID(p.ChooseCards, rock) || !effectstest.HasID(p.ChooseCards, land) || !effectstest.HasID(p.ChooseCards, bear) {
		t.Error("every other permanent you control is offered")
	}
	if p.ChooseMin != 0 {
		t.Errorf("any number: min %d, want 0", p.ChooseMin)
	}
	hand := me.Hand.Size()
	if err := g.ResolveOwnPermanents(p.ID, me.ID, []uuid.UUID{rock, land}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if g.Battlefield.Contains(rock) || g.Battlefield.Contains(land) {
		t.Error("the chosen permanents are sacrificed")
	}
	if !g.Battlefield.Contains(bear) || !g.Battlefield.Contains(bontu) {
		t.Error("the unchosen stay")
	}
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("then draw that many: hand %d → %d", hand, got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bontu, "menace") {
		t.Error("menace")
	}
	// Dies: may go third from the top.
	effectstest.B25Destroy(g, bontu)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(bontu) {
		t.Fatal("Bontu should have left the graveyard")
	}
	names := effectstest.LibraryTopNames(me, 3)
	if len(names) != 3 || names[2] != "God-Eternal Bontu" || names[0] == "God-Eternal Bontu" {
		t.Errorf("third from the top, got %v", names)
	}
}

func TestB33LathielDistributesTheLifeGainedAsCountersAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lathiel := effectstest.B12Push(g, me.ID, "Lathiel, the Bounteous Dawn", "Legendary Creature — Unicorn", effectstest.B33LathielOracle, 2, 2)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, lathiel, "lifelink") {
		t.Error("lifelink")
	}
	// No life gained: no trigger at the end step.
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("no life gained this turn, no trigger")
	}
	// Three life on an opponent's turn: the controller divides the
	// three counters as they choose (#1657) — one on the first
	// creature picked, two on the second.
	effectstest.AdvanceToMainOf(t, g, 1)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	effectstest.AdvanceToEndStepOf(t, g, 1)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if effectstest.HasID(p.PickTargetCards, lathiel) {
		t.Error("Lathiel herself is not \"other\"")
	}
	if !effectstest.HasID(p.PickTargetCards, a) || !effectstest.HasID(p.PickTargetCards, b) {
		t.Error("any creature, an opponent's included, is offered")
	}
	effectstest.B17PickCardsDivided(t, g, me.ID, map[uuid.UUID]int{a: 1, b: 2}, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, a, "+1/+1"); got != 1 {
		t.Errorf("the first picked: %d, want the 1 announced for it", got)
	}
	if got := effectstest.CounterCount(g, b, "+1/+1"); got != 2 {
		t.Errorf("the second picked: %d, want the 2 announced for it", got)
	}
	// One creature picked, no share named, gets every counter.
	effectstest.AdvanceToMainOf(t, g, 2)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2) })
	effectstest.AdvanceToEndStepOf(t, g, 2)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.B17PickCards(t, g, me.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, a, "+1/+1"); got != 3 {
		t.Errorf("two more on the one creature picked: %d, want 3", got)
	}
}

func TestB33AlelaMakesAFaerieOnYourFirstSpellOfAnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Alela, Cunning Conqueror", "Legendary Creature — Faerie Warlock", effectstest.B33AlelaOracle, 2, 4)
	effectstest.CastCatalogSpell(t, g, "Shock", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Faerie Rogue"); n != 0 {
		t.Fatalf("your own turn: %d Faeries", n)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.B22CastInstantAs(t, g, me, "Opt")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B22CastInstantAs(t, g, me, "Brainstorm")
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Faerie Rogue"); n != 1 {
		t.Fatalf("the first spell of an opponent's turn, once: %d Faeries", n)
	}
	faerie := effectstest.FindBattlefieldByName(g, "Faerie Rogue")
	effectstest.AssertKeywords(t, g, faerie, "flying")
	if c, _ := effectstest.BattlefieldCard(g, faerie); c.Power != 1 || c.Toughness != 1 || !c.HasSubtype("Faerie") {
		t.Error("a 1/1 Faerie Rogue")
	}
}

func TestB33AlelaGoadsACreatureOfThePlayerHerFaeriesHit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim, other := g.Seats[0], g.Seats[1], g.Seats[2]
	alela := effectstest.B12Push(g, me.ID, "Alela, Cunning Conqueror", "Legendary Creature — Faerie Warlock", effectstest.B33AlelaOracle, 2, 4)
	sprite := effectstest.B12Creature(g, me.ID, "Sprite", "Creature — Faerie", 1, 1)
	theirs := effectstest.B12Creature(g, victim.ID, "Their Bear", "Creature — Bear", 2, 2)
	otherBear := effectstest.B12Creature(g, other.ID, "Other Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AttackWith(t, g, victim.ID, alela, sprite)
	if victim.Life != 40-3 {
		t.Fatalf("two Faeries connected for 3: life %d", victim.Life)
	}
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if !effectstest.HasID(p.PickTargetCards, theirs) {
		t.Error("the hit player's creature is offered")
	}
	if effectstest.HasID(p.PickTargetCards, otherBear) || effectstest.HasID(p.PickTargetCards, mine) {
		t.Error("another player's creature and your own are not offered")
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("one or more Faeries is one trigger")
	}
	if effectstest.B33Goaded(t, g, theirs) != me.ID {
		t.Fatal("the chosen creature is goaded by you")
	}
	// #1571: the goad is enforced (CR 701.15b). On the victim's turn
	// the Bear must attack, and a player other than you: the pass that
	// ends the declaration is refused while it stays home, an attack
	// on you is refused while another opponent is open, and one on the
	// other opponent is the answer.
	for i := 0; i < 40 && (g.Turn.Step != game.StepDeclareAttackers || g.Turn.ActiveSeat != 1); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep toward the victim's declare attackers: %v", err)
		}
	}
	if err := g.PassPriority(); !errors.Is(err, game.ErrAttackRequirement) {
		t.Fatalf("pass with the goaded Bear at home = %v, want ErrAttackRequirement", err)
	}
	if err := g.DeclareAttacker(theirs, me.ID); !errors.Is(err, game.ErrAttackRequirement) {
		t.Fatalf("the goaded Bear at its goader = %v, want ErrAttackRequirement", err)
	}
	if err := g.DeclareAttacker(theirs, other.ID); err != nil {
		t.Fatalf("the goaded Bear at the other opponent: %v", err)
	}
	// Until your next turn: still goaded through the opponents'
	// turns, cleared at your upkeep.
	effectstest.AdvanceToUpkeepOf(t, g, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B33Goaded(t, g, theirs) != me.ID {
		t.Error("still goaded on another opponent's turn")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B33Goaded(t, g, theirs) != uuid.Nil {
		t.Error("the goad ends at the beginning of your next turn")
	}
	// A non-Faerie connecting does not goad.
	effectstest.AttackWith(t, g, victim.ID, mine)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a Bear is not a Faerie")
	}
}
