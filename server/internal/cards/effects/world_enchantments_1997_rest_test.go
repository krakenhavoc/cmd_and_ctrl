package effects

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// world_enchantments_1997_rest_test.go — #1997, the last three world
// enchantments: Null Chamber's two names, The Abyss's target chosen by
// the upkeep player, and Arboria's "during their last turn".

const (
	nullChamberOracle     = "cb75aca6-dfc7-460d-b90c-80b7988f39a8"
	theAbyssOracle        = "a02a7816-c967-4503-bb08-f8db44915250"
	magusOfTheAbyssOracle = "66958c48-3ac7-4f41-aa3e-adacaa24ea4f"
	arboriaOracle         = "acb3e93c-a1d3-458f-b8c3-c426cd359fa4"
)

// --- Null Chamber ----------------------------------------------------

// ncEnter puts a Null Chamber from `p`'s hand onto the battlefield, so
// its as-enters questions are asked as they are in play.
func ncEnter(t *testing.T, g *game.Game, p *game.Player) uuid.UUID {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: "Null Chamber", OracleID: nullChamberOracle,
		TypeLine: "World Enchantment", ManaCost: "{3}{W}", Owner: p.ID, Controller: p.ID})
	var err error
	g.WithWriteLock(func() {
		err = g.PutCardsFromHandOntoBattlefieldKeyedForEffect([]uuid.UUID{id}, game.HandEntryOptions{}, game.KeyedTail{})
	})
	if err != nil || !g.Battlefield.Contains(id) {
		t.Fatalf("setup: Null Chamber did not enter (err %v)", err)
	}
	return id
}

// ncNamePrompts lists the open card-name prompts in queue order.
func ncNamePrompts(g *game.Game) []*game.PendingChoice {
	var out []*game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceCardName {
			out = append(out, c)
		}
	}
	return out
}

// ncPickOpponent answers the controller's "which opponent" question.
func ncPickOpponent(t *testing.T, g *game.Game, chooser, opponent uuid.UUID) {
	t.Helper()
	c := latestOptionPickFor(g, chooser)
	if c == nil {
		t.Fatal("no choose-an-opponent question is open")
	}
	for i, o := range c.PickOptions {
		if o.Player == opponent {
			if err := g.ResolveOptionPick(c.ID, chooser, i); err != nil {
				t.Fatalf("ResolveOptionPick: %v", err)
			}
			return
		}
	}
	t.Fatalf("the opponent %s is not offered: %+v", opponent, c.PickOptions)
}

func ncHandCard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: "{0}", Owner: p.ID, Controller: p.ID})
	return id
}

// The controller chooses which opponent names a card (the 2013 ruling),
// the two names are asked in APNAP order, a basic land card's name is
// refused, and the table is a restore point while the questions are
// open.
func TestNullChamberAsksYouAndTheOpponentYouChoose(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	them := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	chamber := ncEnter(t, g, me)
	if len(ncNamePrompts(g)) != 0 {
		t.Fatal("a name was asked before the opponent was chosen")
	}
	// The keyed opponent question is plain data: restore across it.
	g = restoreRoundTrip(t, g, true)
	ncPickOpponent(t, g, me.ID, them.ID)

	prompts := ncNamePrompts(g)
	if len(prompts) != 2 || prompts[0].Chooser != me.ID || prompts[1].Chooser != them.ID {
		t.Fatalf("name prompts %+v, want the controller's then the chosen opponent's (APNAP)", prompts)
	}
	if !prompts[0].NotBasicLandName || !prompts[1].NotBasicLandName || prompts[0].SecondName || !prompts[1].SecondName {
		t.Fatalf("prompt flags wrong: %+v / %+v", prompts[0], prompts[1])
	}
	for _, basic := range []string{"Forest", "snow-covered island", " Wastes "} {
		if err := g.ResolveCardNameChoice(prompts[0].ID, me.ID, basic); !errors.Is(err, game.ErrInvalidParam) {
			t.Errorf("naming %q: err = %v, want a refusal (CR 201.4a)", basic, err)
		}
	}
	actNameFor(t, g, me.ID, "Strip Mine")
	g = restoreRoundTrip(t, g, true)
	actNameFor(t, g, them.ID, "Lightning Bolt")

	c := findBattlefieldCardForTest(g, chamber)
	if c.ChosenName != "Strip Mine" || c.ChosenSecondName != "Lightning Bolt" {
		t.Fatalf("names %q / %q, want Strip Mine / Lightning Bolt", c.ChosenName, c.ChosenSecondName)
	}
}

// Spells with either name can't be cast and lands with either name
// can't be played, by any player, the controller included; other cards
// are untouched; both names survive a restore.
func TestNullChamberBansBothNamesForEveryone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ncEnter(t, g, me)
	ncPickOpponent(t, g, me.ID, them.ID)
	actNameFor(t, g, me.ID, "Strip Mine")
	actNameFor(t, g, them.ID, "lightning bolt")
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)

	bolt := ncHandCard(me, "Lightning Bolt", "Instant")
	assertCantCast(t, g.CastSpell(me.ID, bolt, game.CastSpellParams{}), "chosen names")
	mine := ncHandCard(me, "Strip Mine", "Land")
	lpAssertRefused(t, g, me, mine, g.CastSpell(me.ID, mine, game.CastSpellParams{}), "Lands with the chosen names can't be played.")
	if err := lpGate(g, g.Seats[(g.Turn.ActiveSeat+3)%len(g.Seats)], "Strip Mine", game.ZoneHand); err == nil {
		t.Error("an opponent may play a land with a chosen name")
	}
	shock := ncHandCard(me, "Shock", "Instant")
	if err := g.CastSpell(me.ID, shock, game.CastSpellParams{}); errors.Is(err, game.ErrCantCast) {
		t.Errorf("a spell with another name was refused: %v", err)
	}
}

// Before anyone answers, nothing is banned; with no opponent left only
// the controller names a card, and nobody is asked which opponent.
func TestNullChamberWithOneOpponentAsksNoQuestionAndBansNothingYet(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	g.WithWriteLock(func() {
		for _, p := range g.Seats {
			if p.ID != me.ID && p.ID != opp.ID {
				p.Eliminated = true
			}
		}
	})
	ncEnter(t, g, me)
	if latestOptionPickFor(g, me.ID) != nil {
		t.Fatal("with one opponent, nobody should be asked which")
	}
	prompts := ncNamePrompts(g)
	if len(prompts) != 2 || prompts[1].Chooser != opp.ID {
		t.Fatalf("name prompts %+v, want the controller's and the one opponent's", prompts)
	}
	if lpGate(g, me, "Strip Mine", game.ZoneHand) != nil {
		t.Error("an unanswered name banned a land")
	}
}

// The bot never offers a basic land card's name to a prompt that
// refuses one, though an opponent's Forest has a mana ability.
func TestNullChamberPromptOffersTheBotNoBasicLandName(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	apaPush(g, opp.ID, opp.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	ncEnter(t, g, me)
	ncPickOpponent(t, g, me.ID, opp.ID)
	offered := 0
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Kind != legal.KindChoice {
			continue
		}
		var p struct {
			CardName string `json:"card_name"`
		}
		_ = json.Unmarshal(m.Params, &p)
		if p.CardName == "" {
			continue
		}
		offered++
		if game.IsBasicLandCardName(p.CardName) {
			t.Errorf("the bot is offered the basic land name %q", p.CardName)
		}
	}
	if offered == 0 {
		t.Fatal("the bot was offered no card name at all")
	}
}

// --- The Abyss -----------------------------------------------------

// The upkeep player chooses (the 2004 ruling), among the nonartifact
// creatures they control, for the Abyss's controller; the creature is
// destroyed through a regeneration shield, and a restore with the
// trigger on the stack keeps all of it.
func TestTheAbyssTargetIsChosenByTheUpkeepPlayer(t *testing.T) {
	g := newCatalogGame(t)
	owner, victim, bystander := g.Seats[3], g.Seats[1], g.Seats[2]
	abyss := pushPermanentForTest(g, owner.ID, "The Abyss", theAbyssOracle, "World Enchantment")
	bear := b12Creature(g, victim.ID, "Bear", "Creature — Bear", 2, 2)
	elk := b12Creature(g, victim.ID, "Elk", "Creature — Elk", 3, 3)
	golem := b12Creature(g, victim.ID, "Golem", "Artifact Creature — Golem", 3, 3)
	other := b12Creature(g, bystander.ID, "Wolf", "Creature — Wolf", 2, 2)

	advanceToUpkeepOf(t, g, 1)
	ch := pickTargetPrompt(t, g)
	if ch.Chooser != victim.ID || ch.PickTargetFor != owner.ID {
		t.Fatalf("prompt for %s on behalf of %s, want the upkeep player for the Abyss's controller", ch.Chooser, ch.PickTargetFor)
	}
	legalSet := map[uuid.UUID]bool{}
	for _, id := range ch.PickTargetCards {
		legalSet[id] = true
	}
	if !legalSet[bear] || !legalSet[elk] || legalSet[golem] || legalSet[other] || len(legalSet) != 2 {
		t.Fatalf("legal set %v, want only the upkeep player's two nonartifact creatures", ch.PickTargetCards)
	}
	if err := g.ResolvePickTarget(ch.ID, owner.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); !errors.Is(err, game.ErrNotTheChooser) {
		t.Fatalf("the Abyss's controller answered the upkeep player's prompt: %v", err)
	}
	answerPickTarget(t, g, elk)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, elk).RegenerationShields = 1 })
	if triggerOnStack(g, abyss) == nil {
		t.Fatal("the trigger is not on the stack after the target was chosen")
	}
	restored := restoreRoundTrip(t, g, true)
	passPriorityAroundTable(t, restored)
	if _, ok := battlefieldCard(restored, elk); ok {
		t.Error("the chosen creature survived (it can't be regenerated)")
	}
	if _, ok := battlefieldCard(restored, bear); !ok {
		t.Error("a creature nobody chose was destroyed")
	}
}

// On its controller's own upkeep the controller chooses, for itself;
// a Magus that is its controller's only nonartifact creature is the only
// choice.
func TestMagusOfTheAbyssOnItsControllersUpkeepMustChooseItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	magus := pushCatalogPermanent(g, me.ID, "Magus of the Abyss", "Creature — Human Wizard", magusOfTheAbyssOracle, false)
	advanceToUpkeepOf(t, g, 1)
	ch := pickTargetPrompt(t, g)
	if ch.Chooser != me.ID || ch.PickTargetFor != uuid.Nil {
		t.Fatalf("prompt for %s on behalf of %s, want the controller for themselves", ch.Chooser, ch.PickTargetFor)
	}
	if len(ch.PickTargetCards) != 1 || ch.PickTargetCards[0] != magus {
		t.Fatalf("legal set %v, want only the Magus", ch.PickTargetCards)
	}
	answerPickTarget(t, g, magus)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, magus); ok {
		t.Error("the Magus chose itself and survived")
	}
}

// An upkeep player whose only creature has hexproof has no legal choice
// for an opponent's Abyss (legality is judged for the controller), so
// the trigger is removed (CR 603.3d) and nobody is asked.
func TestTheAbyssHexproofLeavesTheUpkeepPlayerNoChoice(t *testing.T) {
	g := newCatalogGame(t)
	owner, victim := g.Seats[3], g.Seats[1]
	abyss := pushPermanentForTest(g, owner.ID, "The Abyss", theAbyssOracle, "World Enchantment")
	hexed := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Slippery Bogle",
		TypeLine: "Creature — Beast", Power: 1, Toughness: 1, Owner: victim.ID, Controller: victim.ID,
		Keywords: []string{"hexproof"}})
	advanceToUpkeepOf(t, g, 1)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			t.Fatalf("a prompt was opened with no legal choice: %+v", c)
		}
	}
	if triggerOnStack(g, abyss) != nil {
		t.Error("the trigger reached the stack with no legal target")
	}
	if _, ok := battlefieldCard(g, hexed); !ok {
		t.Error("the hexproof creature was destroyed")
	}
}

// --- Arboria -------------------------------------------------------

// arbTargets is who the enumerator offers `attacker` at, as ids.
func arbTargets(t *testing.T, g *game.Game, seat, attacker uuid.UUID) map[string]bool {
	t.Helper()
	return bloodthirsterTargets(t, g, seat, attacker)
}

// A round of turns: seat 1 plays a land, seat 2 casts an instant, seat 3
// does nothing. On seat 0's next turn its creature may attack seats 1
// and 2 and seat 3's planeswalker, but not seat 3; on its first turn,
// before anyone had a last turn, it could attack only the planeswalker.
// The engine, the enumerator and a restore agree.
func TestArboriaProtectsOnlyAPlayerWhoDidNothingLastTurn(t *testing.T) {
	g := newCatalogGame(t)
	if g.Turn.ActiveSeat != 0 {
		t.Fatalf("setup: seat %d is active, want seat 0", g.Turn.ActiveSeat)
	}
	s0, s1, s2, s3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	pushPermanentForTest(g, s3.ID, "Arboria", arboriaOracle, "World Enchantment")
	bear := b12Creature(g, s0.ID, "Bear", "Creature — Bear", 2, 2)
	walker := pushWalkerForTest(g, s3.ID, "Test Walker", "", 3)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, bear).SummonedThisTurn = false })

	advanceToDeclareAttackersOf(t, g, 0)
	first := arbTargets(t, g, s0.ID, bear)
	if len(first) != 1 || !first[walker.String()] {
		t.Fatalf("first turn offers %v, want only the planeswalker", first)
	}

	advanceToMainOf(t, g, 1)
	land := lpHandLand(s1, "Forest")
	if err := g.CastSpell(s1.ID, land, game.CastSpellParams{}); err != nil {
		t.Fatalf("seat 1's land play: %v", err)
	}
	advanceToMainOf(t, g, 2)
	shock := ncHandCard(s2, "Shock", "Instant")
	if err := g.CastSpell(s2.ID, shock, game.CastSpellParams{}); err != nil {
		t.Fatalf("seat 2's spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	advanceToMainOf(t, g, 3)

	advanceToDeclareAttackersOf(t, g, 0)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, bear).Tapped = false })
	got := arbTargets(t, g, s0.ID, bear)
	if !got[s1.ID.String()] || !got[s2.ID.String()] || !got[walker.String()] || got[s3.ID.String()] || len(got) != 3 {
		t.Fatalf("offered %v, want seats 1 and 2 and the planeswalker, not seat 3", got)
	}
	restored := restoreRoundTrip(t, g, true)
	err := restored.Clone().DeclareAttacker(bear, s3.ID)
	var refused *game.AttackTargetRestrictionError
	if !errors.As(err, &refused) || !strings.Contains(refused.Sentence(), "didn't cast a spell or put a nontoken permanent onto the battlefield during their last turn") {
		t.Fatalf("attacking the idle player after a restore: err = %v, want Arboria's refusal", err)
	}
	if err := restored.DeclareAttacker(bear, s1.ID); err != nil {
		t.Fatalf("attacking the player who played a land: %v", err)
	}
}

// "Who put a permanent onto the battlefield, as opposed to under whose
// control it entered": Chaos Warp's "they put it onto the battlefield"
// is the owner's put, not the caster's.
func TestChaosWarpIsTheOwnersPut(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	them.Library.Cards = nil
	bear := b12Creature(g, them.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Chaos Warp", "Instant", plChaosWarpOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("setup: the warped creature is the only card in the library and comes back")
	}
	if n := g.TurnTallyFor(them.ID).NontokenPermanentsPut; n != 1 {
		t.Errorf("the owner put %d permanents, want 1", n)
	}
	if n := g.TurnTallyFor(me.ID).NontokenPermanentsPut; n != 0 {
		t.Errorf("the caster put %d permanents, want 0: the owner did", n)
	}
}

// Living Death's "each player … puts all cards they exiled this way onto
// the battlefield" is each player's own put.
func TestLivingDeathIsEachPlayersPut(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushGraveyardCardTyped(me, "Bear", "Creature — Bear")
	pushGraveyardCardTyped(them, "Elk", "Creature — Elk")
	pushGraveyardCardTyped(them, "Wolf", "Creature — Wolf")
	castCatalogSpell(t, g, "Living Death", "Sorcery", b03LivingDeathOracle, nil)
	passPriorityAroundTable(t, g)
	if n := g.TurnTallyFor(them.ID).NontokenPermanentsPut; n != 2 {
		t.Errorf("seat 1 put %d permanents, want its own 2", n)
	}
	if n := g.TurnTallyFor(me.ID).NontokenPermanentsPut; n != 1 {
		t.Errorf("the caster put %d permanents, want only its own 1", n)
	}
}
