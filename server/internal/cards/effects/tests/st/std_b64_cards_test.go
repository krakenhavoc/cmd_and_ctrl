package testsst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b64_cards_test.go — Lorwyn Eclipsed leftovers, batch 64 (#4458).

const (
	stdB64AssistanceOracle = "c452e748-8a2a-4f68-8869-e294925a8b0b"
	stdB64SpriteOracle     = "d0a38e8c-c42c-4c99-8a44-d390ff2511ff"
	stdB64DistracterOracle = "6c0f6673-bc24-4c84-b4e7-736283ffd60e"
	stdB64FarewellOracle   = "7e1da4a1-5922-4a2e-8bbc-25fbcbf20935"
	stdB64BiddingOracle    = "bdb9e933-f679-4930-b669-0d123db09d16"
	stdB64DoseOracle       = "f852bc33-e2f8-43be-9e6e-7badd1f894ef"
	stdB64AuntieOracle     = "cc9523da-f48e-49ab-9989-a4031dafb426"
	stdB64ExtractorOracle  = "980affa7-6308-4acd-af64-f804e574a89b"
	stdB64MoonshadowOracle = "a3635b3f-839c-4fe5-83a5-7c70e8630184"
)

func TestStdB64EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		stdB64AssistanceOracle: "Unexpected Assistance",
		stdB64SpriteOracle:     "Unwelcome Sprite",
		stdB64DistracterOracle: "Wanderwine Distracter",
		stdB64FarewellOracle:   "Wanderwine Farewell",
		stdB64BiddingOracle:    "Bloodline Bidding",
		stdB64DoseOracle:       "Dose of Dawnglow",
		stdB64AuntieOracle:     "Heirloom Auntie",
		stdB64ExtractorOracle:  "Moonglove Extractor",
		stdB64MoonshadowOracle: "Moonshadow",
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
	for _, oracle := range []string{stdB64AssistanceOracle, stdB64FarewellOracle, stdB64BiddingOracle} {
		spec, _ := effects.Lookup(oracle)
		if spec.TapCost == nil {
			t.Errorf("%s has no convoke tap cost", spec.Name)
		}
	}
}

func TestStdB64UnexpectedAssistanceDrawsThreeThenDiscardsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Unexpected Assistance", "Instant", stdB64AssistanceOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+3)
	}
	if effectstest.DiscardOwed(g, me.ID) != 1 {
		t.Errorf("owes %d discards, want 1", effectstest.DiscardOwed(g, me.ID))
	}
}

func TestStdB64UnwelcomeSpriteSurveilsOnlyOnAnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// The Sprite is the opponent's: the spell below is cast by the seat
	// whose turn it is not.
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Unwelcome Sprite", TypeLine: "Creature — Faerie Rogue",
		OracleID: stdB64SpriteOracle, Power: 2, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Sprite", TypeLine: "Creature — Faerie Rogue",
		OracleID: stdB64SpriteOracle, Power: 2, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	effectstest.SeedLibrary(opp, "A", "B", "C")
	// Seat 0's own turn: its own Sprite does not trigger on its own cast.
	effectstest.CastCatalogSpell(t, g, "Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	if effectstest.SurveilChoiceFor(g, me.ID) != nil {
		t.Fatal("surveil on a spell cast during your own turn")
	}
	effectstest.PassPriorityAroundTable(t, g)
	// Now the opponent casts during seat 0's turn: its Sprite triggers.
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	effectstest.PushBattlefieldCardWithTimestamp(g2, game.Card{
		InstanceID: uuid.New(), Name: "Unwelcome Sprite", TypeLine: "Creature — Faerie Rogue",
		OracleID: stdB64SpriteOracle, Power: 2, Toughness: 1, Owner: opp2.ID, Controller: opp2.ID,
	})
	effectstest.SeedLibrary(opp2, "A", "B", "C")
	effectstest.Batch01OpponentCasts(t, g2, opp2, "Bolt", effectstest.LightningBoltOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me2.ID}})
	effectstest.PassPriorityAroundTable(t, g2)
	sc := effectstest.SurveilChoiceFor(g2, opp2.ID)
	if sc == nil {
		t.Fatal("no surveil after casting a spell during an opponent's turn")
	}
	if len(sc.ScryCards) != 2 {
		t.Errorf("surveil looks at %d cards, want 2", len(sc.ScryCards))
	}
}

func TestStdB64WanderwineDistracterShrinksAnOpponentsCreatureWhenTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.AdvanceToMain(t, g)
	d := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Wanderwine Distracter", TypeLine: "Creature — Merfolk Wizard",
		OracleID: stdB64DistracterOracle, Power: 4, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 5, 5)
	effectstest.TapForTest(t, g, d)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if effectstest.HasID(pick.PickTargetCards, mine) {
		t.Error("offered my own creature")
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, theirs); p != 2 || tt != 5 {
		t.Errorf("their creature is %d/%d, want 2/5", p, tt)
	}
}

func TestStdB64WanderwineFarewellMakesATokenPerReturnedPermanentWithAMerfolk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, me.ID, "Merfolk Friend", "Creature — Merfolk Wizard", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	land := effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	effectstest.CastCatalogSpell(t, g, "Wanderwine Farewell", "Kindred Sorcery — Merfolk", stdB64FarewellOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}, {Kind: game.TargetCard, ID: rock}})
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(theirs) || !opp.Hand.Contains(rock) {
		t.Fatal("the targets were not returned")
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Merfolk"); n != 2 {
		t.Errorf("%d Merfolk tokens, want one per permanent returned (2)", n)
	}
	if err := effectstest.CastCatalogSpellErr(t, g, "Wanderwine Farewell", "Kindred Sorcery — Merfolk", stdB64FarewellOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land}}); err == nil {
		t.Error("a land was accepted as a nonland permanent target")
	}
}

func TestStdB64WanderwineFarewellMakesNoTokensWithoutAMerfolk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Wanderwine Farewell", "Kindred Sorcery — Merfolk", stdB64FarewellOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(theirs) {
		t.Fatal("the target was not returned")
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Merfolk"); n != 0 {
		t.Errorf("%d Merfolk tokens with no Merfolk, want 0", n)
	}
}

func TestStdB64BloodlineBiddingReturnsEveryCardOfTheChosenType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elf1 := effectstest.Batch01GraveyardCard(me, "Elf One", "Creature — Elf Warrior")
	elf2 := effectstest.Batch01GraveyardCard(me, "Elf Two", "Creature — Elf Druid")
	goblin := effectstest.Batch01GraveyardCard(me, "Goblin", "Creature — Goblin")
	theirElf := effectstest.Batch01GraveyardCard(opp, "Their Elf", "Creature — Elf")
	effectstest.CastCatalogSpell(t, g, "Bloodline Bidding", "Sorcery", stdB64BiddingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerCreatureType(t, g, me.ID, "Elf")
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{elf1, elf2} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("Elf card %s did not return", id)
		}
	}
	if g.Battlefield.Contains(goblin) || g.Battlefield.Contains(theirElf) {
		t.Error("a card of another type, or an opponent's, returned")
	}
}

func TestStdB64DoseOfDawnglowBlightsOutsideYourMainPhase(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 4, 4)
	dead := effectstest.Batch01GraveyardCard(me, "Dead Ogre", "Creature — Ogre")
	effectstest.AdvanceToMain(t, g)
	// Move off the main phase: cast it in combat.
	for g.Turn.Step != game.StepBeginCombat {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Dose of Dawnglow", TypeLine: "Instant", OracleID: stdB64DoseOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dead}}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		t.Fatal("the creature card did not return")
	}
	p := effectstest.BlightPrompt(g, me.ID)
	if p == nil {
		t.Fatal("no blight outside a main phase")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, bear)
	effectstest.MonSettle(t, g)
	if got := effectstest.MinusCounters(g, bear); got != 2 {
		t.Errorf("-1/-1 counters %d, want 2", got)
	}
}

func TestStdB64DoseOfDawnglowDoesNotBlightInYourMainPhase(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 4, 4)
	dead := effectstest.Batch01GraveyardCard(me, "Dead Ogre", "Creature — Ogre")
	effectstest.CastCatalogSpell(t, g, "Dose of Dawnglow", "Instant", stdB64DoseOracle, []game.TargetRef{{Kind: game.TargetCard, ID: dead}})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		t.Fatal("the creature card did not return")
	}
	if effectstest.BlightPrompt(g, me.ID) != nil {
		t.Error("blighted during your own main phase")
	}
}

func TestStdB64HeirloomAuntieEntersWithCountersAndLosesOneWhenAnotherCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	other := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	auntie := effectstest.RfdCast(t, g, "Heirloom Auntie", "Creature — Goblin Warlock", stdB64AuntieOracle, 4, 4)
	if got := effectstest.MinusCounters(g, auntie); got != 2 {
		t.Fatalf("enters with %d -1/-1 counters, want 2", got)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(other) })
	effectstest.PassPriorityAroundTable(t, g)
	sc := effectstest.SurveilChoiceFor(g, me.ID)
	if sc == nil {
		t.Fatal("no surveil")
	}
	if got := effectstest.MinusCounters(g, auntie); got != 2 {
		t.Errorf("the counter came off before the surveil was answered: %d", got)
	}
	if err := g.ResolveSurveil(sc.ID, me.ID, sc.ScryCards, nil); err != nil {
		t.Fatalf("answer surveil: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.MinusCounters(g, auntie); got != 1 {
		t.Errorf("%d -1/-1 counters after the surveil, want 1", got)
	}
}

func TestStdB64HeirloomAuntieIgnoresItsOwnDeathAndAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	auntie := effectstest.RfdCast(t, g, "Heirloom Auntie", "Creature — Goblin Warlock", stdB64AuntieOracle, 4, 4)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SurveilChoiceFor(g, me.ID) != nil {
		t.Error("surveilled for an opponent's creature")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(auntie) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SurveilChoiceFor(g, me.ID) != nil {
		t.Error("surveilled for its own death")
	}
}

func TestStdB64MoonglovExtractorDrawsAndLosesLifeOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	ext := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Moonglove Extractor", TypeLine: "Creature — Elf Warlock",
		OracleID: stdB64ExtractorOracle, Power: 2, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	hand, life := me.Hand.Size(), me.Life
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(ext, g.Seats[(g.Turn.ActiveSeat+1)%4].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Errorf("hand %d life %d, want %d and %d", me.Hand.Size(), me.Life, hand+1, life-1)
	}
}

func TestStdB64MoonshadowEntersWithSixCountersAndRemovesOneForPermanentCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	shadow := effectstest.RfdCast(t, g, "Moonshadow", "Creature — Elemental", stdB64MoonshadowOracle, 7, 7)
	if got := effectstest.MinusCounters(g, shadow); got != 6 {
		t.Fatalf("enters with %d -1/-1 counters, want 6", got)
	}
	if p, tt := effectstest.PtOf(t, g, shadow); p != 1 || tt != 1 {
		t.Errorf("Moonshadow is %d/%d, want 1/1", p, tt)
	}
	// A permanent card dies: one counter comes off.
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.MinusCounters(g, shadow); got != 5 {
		t.Errorf("%d counters after a creature card died, want 5", got)
	}
	// Two permanent cards milled together are one batch: one counter.
	effectstest.SeedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	effectstest.SeedLibraryTop(me, "Rock", "Artifact", "{2}")
	g.WithWriteLock(func() { _ = g.MillNForEffect(me.ID, 2) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.MinusCounters(g, shadow); got != 4 {
		t.Errorf("%d counters after two cards milled at once, want 4", got)
	}
	// An instant is not a permanent card.
	effectstest.SeedLibraryTop(me, "Bolt", "Instant", "{R}")
	g.WithWriteLock(func() { _ = g.MillNForEffect(me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.MinusCounters(g, shadow); got != 4 {
		t.Errorf("%d counters after an instant was milled, want 4", got)
	}
}
