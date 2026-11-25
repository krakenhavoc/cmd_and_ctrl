package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB237Claim   = "16da9e4b-c6ee-4e70-8653-31ca147f3b01"
	oldB237Appeal  = "3ecc4b57-3e0c-48b8-9952-e4b63b6bc45d"
	oldB237Driven  = "87fca901-678d-4c9a-be4a-9dce2a53074b"
	oldB237Grind   = "ab5b697b-ab62-4ee7-a10e-cc937fef3095"
	oldB237Leave   = "dff5d16f-3a41-4873-8188-076e77f18c75"
	oldB237Reason  = "e2b152ea-82d5-465e-8e13-3d62aacd6692"
	oldB237Aftermt = "Aftermath (Cast this spell only from your graveyard. Then exile it.)"
)

func TestOldB237CardsAreFullWithNoCaveats(t *testing.T) {
	want := map[string]string{
		oldB237Claim: "Claim", oldB237Claim + "#1": "Fame",
		oldB237Appeal: "Appeal", oldB237Appeal + "#1": "Authority",
		oldB237Driven: "Driven", oldB237Driven + "#1": "Despair",
		oldB237Grind: "Grind", oldB237Grind + "#1": "Dust",
		oldB237Leave: "Leave", oldB237Leave + "#1": "Chance",
		oldB237Reason: "Reason", oldB237Reason + "#1": "Believe",
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

var oldB237Costs = map[string]string{
	"Claim": "{B}", "Fame": "{1}{R}", "Appeal": "{G}", "Authority": "{1}{W}", "Driven": "{1}{G}", "Despair": "{1}{B}",
	"Grind": "{1}{B}", "Dust": "{3}{W}", "Leave": "{1}{W}", "Chance": "{3}{R}", "Reason": "{U}", "Believe": "{4}{G}",
}

// oldB237Card is a split card with its aftermath half second, seated in `zone`.
func oldB237Card(p *game.Player, oracle, front, frontType, back string, to *game.Zone) uuid.UUID {
	card := game.Card{InstanceID: uuid.New(), OracleID: oracle, Layout: game.LayoutSplit, Owner: p.ID, Controller: p.ID,
		Faces: []game.Face{
			{Name: front, TypeLine: frontType, ManaCost: oldB237Costs[front]},
			{Name: back, TypeLine: "Sorcery", ManaCost: oldB237Costs[back], OracleText: oldB237Aftermt},
		}}
	card.SetFace(0)
	to.PushTop(card)
	effectstest.B06AddMana(p, "W", "U", "B", "R", "G", "C", "C", "C", "C")
	return card.InstanceID
}

// oldB237CastBack casts the aftermath half from the graveyard and resolves it.
func oldB237CastBack(t *testing.T, g *game.Game, me *game.Player, oracle, front, frontType, back string, targets ...uuid.UUID) uuid.UUID {
	t.Helper()
	id := oldB237Card(me, oracle, front, frontType, back, me.Graveyard)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Face: 1, FromZone: "graveyard", Targets: effectstest.CardRefs(targets...)}); err != nil {
		t.Fatalf("cast %s from the graveyard: %v", back, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func oldB237CastFront(t *testing.T, g *game.Game, me *game.Player, oracle, front, frontType, back string, targets ...uuid.UUID) uuid.UUID {
	t.Helper()
	id := oldB237Card(me, oracle, front, frontType, back, me.Hand)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(targets...)}); err != nil {
		t.Fatalf("cast %s: %v", front, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestOldB237ClaimReturnsOnlyAnMV2OrLessCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	small := effectstest.B38Graveyard(g, me.ID, "Small", "Creature — Bear", "{1}{G}", 2, 2)
	big := effectstest.B38Graveyard(g, me.ID, "Big", "Creature — Ogre", "{2}{G}", 3, 3)
	id := oldB237Card(me, oldB237Claim, "Claim", "Sorcery", "Fame", me.Hand)
	effectstest.B06AddMana(me, "B")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(big)}); err == nil {
		t.Fatal("Claim accepted a creature card with mana value 3")
	}
	effectstest.B06AddMana(me, "B")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(small)}); err != nil {
		t.Fatalf("cast Claim: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) {
		t.Error("the mana value 2 creature did not return")
	}
	if g.Battlefield.Contains(big) {
		t.Error("the mana value 3 creature returned")
	}
}

func TestOldB237FameGivesPlusTwoAndHasteFromTheGraveyardThenExiles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	id := oldB237CastBack(t, g, me, oldB237Claim, "Claim", "Sorcery", "Fame", bear)
	p, tough := effectstest.B38PT(t, g, bear)
	if p != 4 || tough != 2 {
		t.Errorf("%d/%d, want 4/2", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("no haste")
	}
	if !g.Exile.Contains(id) {
		t.Error("the card was not exiled after aftermath")
	}
}

func TestOldB237AppealPumpsByTheCreaturesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, me.ID, "Another", "Creature — Bear", 1, 1)
	oldB237CastFront(t, g, me, oldB237Appeal, "Appeal", "Sorcery", "Authority", bear)
	p, tough := effectstest.B38PT(t, g, bear)
	if p != 5 || tough != 5 {
		t.Errorf("%d/%d, want 5/5 (+3/+3 for three creatures)", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("no trample")
	}
}

func TestOldB237AuthorityTapsUpToTwoOpposingCreaturesAndGivesVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	oldB237CastBack(t, g, me, oldB237Appeal, "Appeal", "Sorcery", "Authority", a, b)
	for _, id := range []uuid.UUID{a, b} {
		if !effectstest.B12Card(t, g, id).Tapped {
			t.Error("an opposing creature was not tapped")
		}
	}
	if !effectstest.HasEffectiveKeyword(t, g, mine, "vigilance") {
		t.Error("your creature did not gain vigilance")
	}
	if effectstest.HasEffectiveKeyword(t, g, a, "vigilance") {
		t.Error("an opposing creature gained vigilance")
	}
}

func TestOldB237DrivenDrawsOnCombatDamageAndDespairDiscards(t *testing.T) {
	for _, back := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		effectstest.B64ReadyToAttack(g, bear)
		for i := 0; i < 3; i++ {
			me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
			opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spare", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
		}
		effectstest.AdvanceToMain(t, g)
		kw := "trample"
		if back {
			kw = "menace"
			oldB237CastBack(t, g, me, oldB237Driven, "Driven", "Sorcery", "Despair")
		} else {
			oldB237CastFront(t, g, me, oldB237Driven, "Driven", "Sorcery", "Despair")
		}
		if !effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Fatalf("back=%v: the creature did not gain %s", back, kw)
		}
		myHand, oppHand := me.Hand.Size(), opp.Hand.Size()
		effectstest.LciB1Attack(t, g, bear, opp.ID)
		effectstest.AdvanceTo(t, g, game.StepCombatDamage)
		effectstest.PassPriorityAroundTable(t, g)
		if back {
			if effectstest.DiscardChoiceFor(g, opp.ID) == nil {
				t.Fatal("Despair: the damaged player was not asked to discard")
			}
			effectstest.DiscardFromHand(t, g, opp.ID)
			effectstest.PassPriorityAroundTable(t, g)
			if opp.Hand.Size() != oppHand-1 {
				t.Errorf("Despair: opponent hand %d, want %d", opp.Hand.Size(), oppHand-1)
			}
		} else if me.Hand.Size() != myHand+1 {
			t.Errorf("Driven: hand %d, want %d", me.Hand.Size(), myHand+1)
		}
	}
}

func TestOldB237GrindPutsMinusCountersAndDustExilesOnlyCreaturesWithThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 3, 3)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 3, 3)
	clean := effectstest.B12Creature(g, opp.ID, "Clean", "Creature — Bear", 3, 3)
	oldB237CastFront(t, g, me, oldB237Grind, "Grind", "Sorcery", "Dust", a, b)
	for _, id := range []uuid.UUID{a, b} {
		if n := effectstest.B12Counter(t, g, id, game.CounterMinusOne); n != 1 {
			t.Errorf("-1/-1 counters = %d, want 1", n)
		}
	}
	// Dust refuses a creature with no counter.
	id := oldB237Card(me, oldB237Grind, "Grind", "Sorcery", "Dust", me.Graveyard)
	effectstest.B06AddMana(me, "W", "C", "C", "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Face: 1, FromZone: "graveyard", Targets: effectstest.CardRefs(clean)}); err == nil {
		t.Fatal("Dust accepted a creature with no -1/-1 counter")
	}
	effectstest.B06AddMana(me, "W", "C", "C", "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Face: 1, FromZone: "graveyard", Targets: effectstest.CardRefs(a, b)}); err != nil {
		t.Fatalf("cast Dust: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("a creature with a counter was not exiled")
	}
	if !g.Battlefield.Contains(clean) {
		t.Error("the creature with no counter was exiled")
	}
	if !g.Exile.Contains(id) {
		t.Error("Grind // Dust was not exiled after aftermath")
	}
}

func TestOldB237LeaveReturnsOnlyPermanentsYouOwn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	id := oldB237Card(me, oldB237Leave, "Leave", "Instant", "Chance", me.Hand)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(theirs)}); err == nil {
		t.Fatal("Leave accepted a permanent you do not own")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(a, b)}); err != nil {
		t.Fatalf("cast Leave: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("an owned permanent stayed on the battlefield")
	}
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Error("the permanents are not in hand")
	}
}

func TestOldB237ChanceDiscardsThenDrawsThatMany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 4; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	for i := 0; i < 3; i++ {
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spare", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	hand := me.Hand.Size()
	id := oldB237Card(me, oldB237Leave, "Leave", "Instant", "Chance", me.Graveyard)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Face: 1, FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast Chance: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("no discard prompt")
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d after discarding one and drawing one", me.Hand.Size(), hand)
	}
	if !g.Exile.Contains(id) {
		t.Error("Chance was not exiled after aftermath")
	}
}

func TestOldB237BelieveTakesANonCreatureAndOffersACreature(t *testing.T) {
	// Non-creature on top: it goes to your hand.
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	spell := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: spell, Name: "Shock", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	oldB237CastBack(t, g, me, oldB237Reason, "Reason", "Sorcery", "Believe")
	if !me.Hand.Contains(spell) {
		t.Error("a non-creature top card did not go to hand")
	}

	// Creature on top, put onto the battlefield.
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	bear := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	oldB237CastBack(t, g, me, oldB237Reason, "Reason", "Sorcery", "Believe")
	effectstest.SlChoose(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the chosen creature card did not enter the battlefield")
	}
}
