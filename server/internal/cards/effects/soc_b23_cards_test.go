package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b23_cards_test.go — Secrets of Strixhaven (sos), batch 23 (#4456).
// Each test plays the real catalog entry through the engine and asserts what
// the printed text says.

const (
	socB23DissectionOracle  = "73905f94-9665-4d82-8927-725d2c4fc4bd"
	socB23DivergentOracle   = "84681f12-2978-4f62-ab89-b579033c5192"
	socB23DuelOracle        = "262c5ac9-af41-42d1-a53a-8a671eedd5a6"
	socB23GlyphmageOracle   = "070aa6c7-4741-47de-b68a-629aa4201ac0"
	socB23EfflorescenceOrac = "181da74b-53fb-4e46-afc0-e7c990fda128"
	socB23MascotOracle      = "882f30e4-4ce1-4581-8d2d-d763aa13befe"
	socB23InterceptorOracle = "d97de156-5a6d-4dc6-953e-d4bad63cbad2"
	socB23EmbraceOracle     = "b9110be0-3f0c-449b-8baf-689e92fdb8f0"
	socB23IdeationOracle    = "24a29be4-5bc2-469b-b2ce-2fb305c3facb"
	socB23EmilOracle        = "9f800ebc-d6e4-4bfc-bebe-8baad51efa19"
	socB23BertaOracle       = "715f0832-696c-4111-b418-d9b6da403159"
)

func TestSocB23EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB23DissectionOracle:         "Dissection Practice",
		socB23DivergentOracle:          "Divergent Equation",
		socB23DuelOracle:               "Duel Tactics",
		socB23GlyphmageOracle:          "Eager Glyphmage",
		socB23EfflorescenceOrac:        "Efflorescence",
		socB23MascotOracle:             "Elemental Mascot",
		socB23InterceptorOracle:        "Elite Interceptor",
		socB23InterceptorOracle + "#1": "Rejoinder",
		socB23EmbraceOracle:            "Embrace the Paradox",
		socB23IdeationOracle:           "Emeritus of Ideation",
		socB23IdeationOracle + "#1":    "Ancestral Recall",
		socB23EmilOracle:               "Emil, Vastlands Roamer",
		socB23BertaOracle:              "Berta, Wise Extrapolator",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// socB23ZoneOf reads which zone a card is in.
func socB23ZoneOf(g *game.Game, id uuid.UUID) game.ZoneKind {
	return znrZoneOf(g, id)
}

// --- Dissection Practice ----------------------------------------------------

func socB23CastDissection(t *testing.T, g *game.Game, refs []game.TargetRef) error {
	t.Helper()
	return castCatalogSpellErr(t, g, "Dissection Practice", "Instant", socB23DissectionOracle, refs)
}

func TestSocB23DissectionPracticeDrainsOnePumpsOneAndShrinksOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	myLife, oppLife := me.Life, opp.Life
	refs := []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0},
		{Kind: game.TargetCard, ID: mine, Slot: 1},
		{Kind: game.TargetCard, ID: theirs, Slot: 2},
	}
	if err := socB23CastDissection(t, g, refs); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-1 || me.Life != myLife+1 {
		t.Errorf("life: opponent %d -> %d, you %d -> %d, want -1 and +1", oppLife, opp.Life, myLife, me.Life)
	}
	if p, tough := socB14PT(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("the grown creature is %d/%d, want 3/3", p, tough)
	}
	if p, tough := socB14PT(t, g, theirs); p != 1 || tough != 1 {
		t.Errorf("the shrunk creature is %d/%d, want 1/1", p, tough)
	}
	// Until end of turn only.
	endTurn(t, g)
	if p, tough := socB14PT(t, g, mine); p != 2 || tough != 2 {
		t.Errorf("the grown creature is %d/%d next turn, want 2/2", p, tough)
	}
	if p, tough := socB14PT(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("the shrunk creature is %d/%d next turn, want 2/2", p, tough)
	}
}

func TestSocB23DissectionPracticeCreatureClausesAreOptionalAndMinusOneKills(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tiny := b12Creature(g, opp.ID, "Tiny", "Creature — Bear", 1, 1)
	oppLife := opp.Life
	// Only the opponent: the drain alone is a legal cast.
	if err := socB23CastDissection(t, g, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0}}); err != nil {
		t.Fatalf("cast with only the opponent: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-1 {
		t.Errorf("opponent life %d, want %d", opp.Life, oppLife-1)
	}
	// The -1/-1 alone kills a 1/1.
	refs := []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0},
		{Kind: game.TargetCard, ID: tiny, Slot: 2},
	}
	if err := socB23CastDissection(t, g, refs); err != nil {
		t.Fatalf("cast with the shrink only: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tiny) {
		t.Error("the 1/1 survived -1/-1")
	}
	_ = me
}

func TestSocB23DissectionPracticeNeedsAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if err := socB23CastDissection(t, g, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID, Slot: 0}}); err == nil {
		t.Error("the spell targeted its own caster as the opponent")
	}
}

// --- Divergent Equation -----------------------------------------------------

func socB23DivergentCard() game.Card {
	return game.Card{Name: "Divergent Equation", TypeLine: "Instant", ManaCost: "{X}{X}{U}", OracleID: socB23DivergentOracle}
}

func TestSocB23DivergentEquationReturnsUpToXInstantsAndSorceriesThenExilesItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	a := gyCard(me, "Spell A", "Instant", "{U}", "")
	b := gyCard(me, "Spell B", "Sorcery", "{1}{U}", "")
	c := gyCard(me, "Spell C", "Instant", "{R}", "")
	creature := gyCard(me, "Bear", "Creature — Bear", "{1}{G}", "")
	card := socB23DivergentCard()
	card.InstanceID = uuid.New()
	refs := []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}
	if err := socB19CastWithPool(g, me, card, "U", 5, game.CastSpellParams{XValue: 2, Targets: refs}); err != nil {
		t.Fatalf("cast X=2 with two targets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Error("the two targets did not return to hand")
	}
	if me.Graveyard.Contains(c) == false || me.Graveyard.Contains(creature) == false {
		t.Error("an untargeted card left the graveyard")
	}
	if got := socB23ZoneOf(g, card.InstanceID); got != game.ZoneExile {
		t.Errorf("Divergent Equation is in %v, want exile", got)
	}
	if me.Graveyard.Contains(card.InstanceID) {
		t.Error("Divergent Equation went to the graveyard")
	}
}

func TestSocB23DivergentEquationAtXZeroStillExilesItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	card := socB23DivergentCard()
	card.InstanceID = uuid.New()
	if err := socB19CastWithPool(g, me, card, "U", 1, game.CastSpellParams{XValue: 0}); err != nil {
		t.Fatalf("cast X=0: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := socB23ZoneOf(g, card.InstanceID); got != game.ZoneExile {
		t.Errorf("Divergent Equation is in %v, want exile", got)
	}
}

func TestSocB23DivergentEquationRefusesTooManyTargetsAndACreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	a := gyCard(me, "Spell A", "Instant", "{U}", "")
	b := gyCard(me, "Spell B", "Sorcery", "{1}{U}", "")
	creature := gyCard(me, "Bear", "Creature — Bear", "{1}{G}", "")
	card := socB23DivergentCard()
	card.InstanceID = uuid.New()
	if err := socB19CastWithPool(g, me, card, "U", 5, game.CastSpellParams{XValue: 1,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}}); err == nil {
		t.Error("two targets were accepted for X = 1")
	}
	card2 := socB23DivergentCard()
	card2.InstanceID = uuid.New()
	if err := socB19CastWithPool(g, me, card2, "U", 5, game.CastSpellParams{XValue: 1,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: creature}}}); err == nil {
		t.Error("a creature card was accepted as a target")
	}
}

// --- Duel Tactics -----------------------------------------------------------

func TestSocB23DuelTacticsDamagesTheCreatureAndItCantBlockThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	other := b12Creature(g, opp.ID, "Other Wall", "Creature — Wall", 0, 4)
	castCatalogSpell(t, g, "Duel Tactics", "Sorcery", socB23DuelOracle, cardRefs(wall))
	passPriorityAroundTable(t, g)
	attackIntoBlocks(t, g, opp.ID, bear)
	if err := g.DeclareBlocker(wall, bear); err == nil {
		t.Error("the targeted creature blocked")
	}
	if err := g.DeclareBlocker(other, bear); err != nil {
		t.Errorf("an untargeted creature could not block: %v", err)
	}
	if d := damageMarkedOn(g, wall); d != 1 {
		t.Errorf("damage on the target %d, want 1", d)
	}
}

func TestSocB23DuelTacticsFlashbackCastsFromTheGraveyardAndExiles(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	id := seedGraveyardCard(t, g, "Duel Tactics", "Sorcery", socB23DuelOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback", Targets: cardRefs(wall),
	}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := socB23ZoneOf(g, id); got != game.ZoneExile {
		t.Errorf("the flashed-back Duel Tactics is in %v, want exile", got)
	}
	if d := damageMarkedOn(g, wall); d != 1 {
		t.Errorf("damage on the target %d, want 1", d)
	}
}

func TestSocB23DuelTacticsKillsAOneToughnessCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	tiny := b12Creature(g, opp.ID, "Tiny", "Creature — Bear", 1, 1)
	castCatalogSpell(t, g, "Duel Tactics", "Sorcery", socB23DuelOracle, cardRefs(tiny))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tiny) {
		t.Error("the 1/1 survived 1 damage")
	}
}

// --- Eager Glyphmage --------------------------------------------------------

func socB23Inklings(g *game.Game, owner uuid.UUID) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == "Inkling" {
			out = append(out, c)
		}
	}
	return out
}

func TestSocB23EagerGlyphmageMakesAnInklingWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	c := game.Card{InstanceID: uuid.New(), Name: "Eager Glyphmage", TypeLine: "Creature — Cat Cleric", OracleID: socB23GlyphmageOracle,
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID}
	castFromHandAndResolve(t, g, me, c)
	ink := socB23Inklings(g, me.ID)
	if len(ink) != 1 {
		t.Fatalf("Inklings %d, want 1", len(ink))
	}
	if ink[0].Power != 1 || ink[0].Toughness != 1 || !hasEffectiveKeyword(t, g, ink[0].InstanceID, "flying") {
		t.Errorf("the Inkling is %d/%d, want a 1/1 flier", ink[0].Power, ink[0].Toughness)
	}
	if len(ink[0].Colors) != 2 {
		t.Errorf("the Inkling's colors %v, want white and black", ink[0].Colors)
	}
}

// --- Efflorescence ----------------------------------------------------------

func TestSocB23EfflorescencePutsTwoCountersAndNothingElseWithoutLifeGain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Efflorescence", "Instant", socB23EfflorescenceOrac, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if n := countersOn(g, bear, game.CounterPlusOne); n != 2 {
		t.Errorf("counters %d, want 2", n)
	}
	if hasEffectiveKeyword(t, g, bear, "trample") || hasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("the creature gained keywords without Infusion")
	}
}

func TestSocB23EfflorescenceInfusionGivesTrampleAndIndestructibleUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	sos1GainLife(t, g)
	castCatalogSpell(t, g, "Efflorescence", "Instant", socB23EfflorescenceOrac, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if n := countersOn(g, bear, game.CounterPlusOne); n != 2 {
		t.Errorf("counters %d, want 2", n)
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") || !hasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("the creature did not gain trample and indestructible")
	}
	endTurn(t, g)
	if hasEffectiveKeyword(t, g, bear, "trample") || hasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("the keywords outlasted the turn")
	}
	if n := countersOn(g, bear, game.CounterPlusOne); n != 2 {
		t.Errorf("counters %d after the turn, want to keep 2", n)
	}
}

func TestSocB23EfflorescenceLifeGainedByAnOpponentDoesNotCount(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, 3); err != nil {
		t.Fatalf("gain: %v", err)
	}
	castCatalogSpell(t, g, "Efflorescence", "Instant", socB23EfflorescenceOrac, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("an opponent's life gain turned Infusion on")
	}
}

// --- Elemental Mascot -------------------------------------------------------

func TestSocB23MascotPumpsOnEveryInstantOrSorceryAndExilesOnlyOnFiveMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mascot := b12Push(g, me.ID, "Elemental Mascot", "Creature — Elemental Bird", socB23MascotOracle, 1, 4)
	top := seedLibraryTop(me, "Impulse", "Sorcery", "{R}")
	advanceToMain(t, g)
	if err := socB19CastWithPool(g, me, game.Card{Name: "Small", TypeLine: "Instant", ManaCost: "{R}"}, "R", 1, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, _ := socB14PT(t, g, mascot); p != 2 {
		t.Errorf("power %d after a one-mana instant, want 2", p)
	}
	if got := socB23ZoneOf(g, top); got != game.ZoneLibrary {
		t.Errorf("the top card is in %v after a one-mana spell, want the library", got)
	}
	if err := socB19CastWithPool(g, me, game.Card{Name: "Big", TypeLine: "Sorcery", ManaCost: "{4}{R}"}, "R", 5, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, _ := socB14PT(t, g, mascot); p != 3 {
		t.Errorf("power %d after a five-mana sorcery, want 3", p)
	}
	if got := socB23ZoneOf(g, top); got != game.ZoneExile {
		t.Fatalf("the top card is in %v after a five-mana spell, want exile", got)
	}
	perm := exiledPermission(g, top)
	if perm.Player != me.ID || perm.CastOnly {
		t.Errorf("permission %+v, want a play permission for the controller", perm)
	}
	if !permissionLive(g, perm, me.ID) {
		t.Error("the permission is not live for its controller")
	}
	// Until the end of your NEXT turn: it survives the rest of the round.
	for i := 0; i < 4; i++ {
		endTurn(t, g)
	}
	if g.CastPermissionOnCardByIDForEffect(top) == nil {
		t.Error("the permission lapsed before the end of the controller's next turn")
	}
	endTurn(t, g)
	if g.CastPermissionOnCardByIDForEffect(top) != nil {
		t.Error("the permission outlasted the end of the controller's next turn")
	}
}

func TestSocB23MascotIgnoresACreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mascot := b12Push(g, me.ID, "Elemental Mascot", "Creature — Elemental Bird", socB23MascotOracle, 1, 4)
	advanceToMain(t, g)
	if err := socB19CastWithPool(g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{4}{R}", Power: 2, Toughness: 2}, "R", 5, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, _ := socB14PT(t, g, mascot); p != 1 {
		t.Errorf("power %d after a creature spell, want 1", p)
	}
}

// --- Elite Interceptor // Rejoinder ----------------------------------------

func socB23CastInterceptor(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	advanceTo(t, g, game.StepPrecombatMain)
	c := sos1Prep(me.ID, socB23InterceptorOracle, nil, "Elite Interceptor", "Creature — Human Wizard", "{W}", 1, 2,
		"Rejoinder", "Sorcery", "{1}{W}")
	castFromHandAndResolve(t, g, me, c)
	return c.InstanceID
}

func socB23CastRejoinder(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) {
	t.Helper()
	cp, ok := prepareCopyOf(g, "Rejoinder")
	if !ok {
		t.Fatal("no Rejoinder copy in exile")
	}
	if err := g.CastSpell(me.ID, cp.InstanceID, game.CastSpellParams{FromZone: string(game.ZoneExile), Targets: cardRefs(target)}); err != nil {
		t.Fatalf("cast Rejoinder: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestSocB23InterceptorEntersPreparedAndRejoinderTapsThenDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	inter := socB23CastInterceptor(t, g, me)
	if !g.IsPreparedForEffect(inter) {
		t.Fatal("the Interceptor did not enter prepared")
	}
	hand := me.Hand.Size()
	socB23CastRejoinder(t, g, me, bear)
	answerOptionPick(t, g, me.ID, 0)
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, bear); !c.Tapped {
		t.Error("the creature is not tapped")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, me.Hand.Size())
	}
	if g.IsPreparedForEffect(inter) {
		t.Error("still prepared after casting the copy")
	}
}

func TestSocB23RejoinderUntapsOrLeavesTheCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	socB23CastInterceptor(t, g, me)
	hand := me.Hand.Size()
	socB23CastRejoinder(t, g, me, bear)
	answerOptionPick(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, bear); c.Tapped {
		t.Error("the creature is still tapped")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, me.Hand.Size())
	}
}

func TestSocB23RejoinderMayLeaveTheCreatureAndStillDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	socB23CastInterceptor(t, g, me)
	hand := me.Hand.Size()
	socB23CastRejoinder(t, g, me, bear)
	answerOptionPick(t, g, me.ID, 2)
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, bear); c.Tapped {
		t.Error("the creature was tapped though the caster declined")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, me.Hand.Size())
	}
}

// --- Embrace the Paradox ----------------------------------------------------

func TestSocB23EmbraceTheParadoxDrawsThreeThenPutsALandTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	for _, n := range []string{"Spell A", "Spell B"} {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: n, TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	land := seedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Embrace the Paradox", "Instant", socB23EmbraceOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 {
		t.Fatalf("hand %d -> %d, want three drawn", hand, me.Hand.Size())
	}
	answerChooseCards(t, g, me.ID, land)
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, land)
	if !ok {
		t.Fatal("the land is not on the battlefield")
	}
	if !c.Tapped {
		t.Error("the land entered untapped")
	}
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+2)
	}
}

func TestSocB23EmbraceTheParadoxLandIsOptionalAndNoLandAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	land := seedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	for _, n := range []string{"Spell A", "Spell B"} {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: n, TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	castCatalogSpell(t, g, "Embrace the Paradox", "Instant", socB23EmbraceOracle, nil)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land entered although the player declined")
	}
	if !me.Hand.Contains(land) {
		t.Error("the declined land left the hand")
	}
}

// --- Emeritus of Ideation // Ancestral Recall ------------------------------

func socB23IdeationCard(owner uuid.UUID) game.Card {
	return sos1Prep(owner, socB23IdeationOracle, []string{"flying"}, "Emeritus of Ideation", "Creature — Human Wizard", "{3}{U}{U}", 5, 5,
		"Ancestral Recall", "Instant", "{U}")
}

func TestSocB23IdeationEntersPreparedAndAncestralRecallDrawsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	c := socB23IdeationCard(me.ID)
	castFromHandAndResolve(t, g, me, c)
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("it did not enter prepared")
	}
	cp, ok := prepareCopyOf(g, "Ancestral Recall")
	if !ok {
		t.Fatal("no Ancestral Recall copy in exile")
	}
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, cp.InstanceID, game.CastSpellParams{FromZone: string(game.ZoneExile), Targets: []game.TargetRef{playerTarget(me.ID)}}); err != nil {
		t.Fatalf("cast Ancestral Recall: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 {
		t.Errorf("hand %d -> %d, want three drawn", hand, me.Hand.Size())
	}
	if g.IsPreparedForEffect(c.InstanceID) {
		t.Error("still prepared after casting the copy")
	}
}

// socB23IdeationAttacking puts an unprepared Emeritus on the battlefield with
// `n` cards in its controller's graveyard and declares it as an attacker.
func socB23IdeationAttacking(t *testing.T, g *game.Game, me, opp *game.Player, n int) (uuid.UUID, []uuid.UUID) {
	t.Helper()
	c := socB23IdeationCard(me.ID)
	id := pushBattlefieldCardWithTimestamp(g, c)
	var cards []uuid.UUID
	for i := 0; i < n; i++ {
		cards = append(cards, gyCard(me, "Filler", "Sorcery", "{R}", ""))
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(id, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	return id, cards
}

func TestSocB23IdeationAttackExilesEightCardsAndBecomesPrepared(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id, cards := socB23IdeationAttacking(t, g, me, opp, 9)
	if g.IsPreparedForEffect(id) {
		t.Fatal("prepared before the trigger")
	}
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, cards[:8]...)
	passPriorityAroundTable(t, g)
	if !g.IsPreparedForEffect(id) {
		t.Fatal("not prepared after exiling eight cards")
	}
	for _, c := range cards[:8] {
		if got := socB23ZoneOf(g, c); got != game.ZoneExile {
			t.Errorf("a chosen card is in %v, want exile", got)
		}
	}
	if !me.Graveyard.Contains(cards[8]) {
		t.Error("the ninth card left the graveyard")
	}
	if _, ok := prepareCopyOf(g, "Ancestral Recall"); !ok {
		t.Error("no Ancestral Recall copy in exile")
	}
}

func TestSocB23IdeationAttackDeclinedOrShortDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id, cards := socB23IdeationAttacking(t, g, me, opp, 8)
	passPriorityAroundTable(t, g)
	// Fewer than eight is not an answer the prompt accepts.
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt for eight cards in the graveyard")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, cards[:3]); err == nil {
		t.Fatal("three cards were accepted for an exile-eight")
	}
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if g.IsPreparedForEffect(id) {
		t.Error("prepared though the player declined")
	}
	if n := len(me.Graveyard.Cards); n != 8 {
		t.Errorf("graveyard %d, want 8 untouched", n)
	}
}

func TestSocB23IdeationAttackWithSevenCardsAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id, _ := socB23IdeationAttacking(t, g, me, opp, 7)
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a prompt opened with only seven cards in the graveyard")
	}
	if g.IsPreparedForEffect(id) {
		t.Error("prepared with seven cards")
	}
}

func TestSocB23IdeationAlreadyPreparedMayStillExileAndStaysOnePrepare(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	c := socB23IdeationCard(me.ID)
	id := pushBattlefieldCardWithTimestamp(g, c)
	g.WithWriteLock(func() {
		if _, err := g.BecomePreparedForEffect(id); err != nil {
			t.Fatalf("prepare: %v", err)
		}
	})
	var cards []uuid.UUID
	for i := 0; i < 8; i++ {
		cards = append(cards, gyCard(me, "Filler", "Sorcery", "{R}", ""))
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(id, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, cards...)
	passPriorityAroundTable(t, g)
	if got := socB23ZoneOf(g, cards[0]); got != game.ZoneExile {
		t.Errorf("a chosen card is in %v, want exile", got)
	}
	if n := countPrepareCopies(g); n != 1 {
		t.Errorf("prepare copies %d, want exactly 1", n)
	}
}

// --- Emil, Vastlands Roamer -------------------------------------------------

func TestSocB23EmilGivesTrampleToYourCreaturesWithPlusOneCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	emil := b12Push(g, me.ID, "Emil, Vastlands Roamer", "Legendary Creature — Elf Druid", socB23EmilOracle, 3, 3)
	grown := b12Creature(g, me.ID, "Grown", "Creature — Bear", 2, 2)
	plain := b12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	addCountersForTest(g, grown, game.CounterPlusOne, 1)
	addCountersForTest(g, theirs, game.CounterPlusOne, 1)
	if !hasEffectiveKeyword(t, g, grown, "trample") {
		t.Error("a creature with a +1/+1 counter has no trample")
	}
	if hasEffectiveKeyword(t, g, plain, "trample") {
		t.Error("a creature without a counter has trample")
	}
	if hasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("an opponent's creature with a counter has trample")
	}
	if hasEffectiveKeyword(t, g, emil, "trample") {
		t.Error("Emil has trample with no counter on it")
	}
	addCountersForTest(g, emil, game.CounterPlusOne, 1)
	if !hasEffectiveKeyword(t, g, emil, "trample") {
		t.Error("Emil with a counter has no trample")
	}
}

func TestSocB23EmilMakesAFractalWithOneCounterPerDifferentlyNamedLand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	emil := b12Push(g, me.ID, "Emil, Vastlands Roamer", "Legendary Creature — Elf Druid", socB23EmilOracle, 3, 3)
	for _, n := range []string{"Forest", "Forest", "Forest", "Island", "Command Tower"} {
		b12Permanent(g, me.ID, n, "Land")
	}
	b12Permanent(g, opp.ID, "Mountain", "Basic Land — Mountain")
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	fillPoolColored(me, "G", 5)
	if err := g.ActivateCatalogAbility(me.ID, emil, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	fractals := socB4Fractals(g, me.ID)
	if len(fractals) != 1 {
		t.Fatalf("Fractals %d, want 1", len(fractals))
	}
	if n := countersOn(g, fractals[0].InstanceID, game.CounterPlusOne); n != 3 {
		t.Errorf("counters %d, want 3 (Forest, Island, Command Tower)", n)
	}
	if c, _ := battlefieldCard(g, emil); !c.Tapped {
		t.Error("Emil did not tap")
	}
}

func TestSocB23EmilCantActivateWithoutTheMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	emil := b12Push(g, me.ID, "Emil, Vastlands Roamer", "Legendary Creature — Elf Druid", socB23EmilOracle, 3, 3)
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	fillPoolColored(me, "G", 4)
	if err := g.ActivateCatalogAbility(me.ID, emil, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the ability activated for {4}{G} with four mana")
	}
}

// --- Berta, Wise Extrapolator ----------------------------------------------

func socB23Berta(g *game.Game, me *game.Player) uuid.UUID {
	return b12Push(g, me.ID, "Berta, Wise Extrapolator", "Legendary Creature — Frog Druid", socB23BertaOracle, 1, 4)
}

func TestSocB23BertaIncrementPutsACounterAndThatCounterMakesMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	berta := socB23Berta(g, me)
	advanceToMain(t, g)
	// Two mana is greater than Berta's power of 1.
	if err := socB19CastWithPool(g, me, game.Card{Name: "Spell", TypeLine: "Sorcery", ManaCost: "{1}{G}"}, "G", 2, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, berta, game.CounterPlusOne); n != 1 {
		t.Fatalf("counters %d, want 1", n)
	}
	socB23AnswerAnyColor(t, g, me, "G")
	if got := me.ManaPool.CountColor("G"); got != 1 {
		t.Errorf("green mana in the pool %d, want the one added by the counter trigger", got)
	}
	if got := me.ManaPool.Total(); got != 1 {
		t.Errorf("mana in the pool %d, want exactly 1", got)
	}
}

// socB23AnswerAnyColor answers the open "add one mana of any color" pick.
func socB23AnswerAnyColor(t *testing.T, g *game.Game, me *game.Player, color string) {
	t.Helper()
	pick := latestChoiceOfKindFor(g, game.PendingChoiceMana, me.ID)
	if pick == nil {
		t.Fatal("no colour pick for the any-colour mana")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, color); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
}

func TestSocB23BertaOneManaSpellDoesNotIncrementAndOtherCountersMakeNoMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	berta := socB23Berta(g, me)
	advanceToMain(t, g)
	// One mana is not greater than power 1 or toughness 4.
	if err := socB19CastWithPool(g, me, game.Card{Name: "Spell", TypeLine: "Sorcery", ManaCost: "{G}"}, "G", 1, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, berta, game.CounterPlusOne); n != 0 {
		t.Errorf("counters %d, want 0", n)
	}
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("mana in the pool %d, want 0", got)
	}
	// A counter of another kind is not a +1/+1 counter.
	addCountersForTest(g, berta, "stun", 1)
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceMana, me.ID) != nil || me.ManaPool.Total() != 0 {
		t.Errorf("a stun counter made mana (pool %d)", me.ManaPool.Total())
	}
	// A +1/+1 counter from any source does make mana, once however many.
	addCountersForTest(g, berta, game.CounterPlusOne, 2)
	passPriorityAroundTable(t, g)
	socB23AnswerAnyColor(t, g, me, "U")
	if got := me.ManaPool.Total(); got != 1 {
		t.Errorf("mana in the pool %d after one placement of two counters, want 1", got)
	}
}

func TestSocB23BertaXAbilityMakesAFractalWithXCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	berta := socB23Berta(g, me)
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	fillPoolColored(me, "G", 3)
	if err := g.ActivateCatalogAbility(me.ID, berta, 0, game.ActivateAbilityParams{Strict: true, XValue: 3}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	fractals := socB4Fractals(g, me.ID)
	if len(fractals) != 1 {
		t.Fatalf("Fractals %d, want 1", len(fractals))
	}
	if n := countersOn(g, fractals[0].InstanceID, game.CounterPlusOne); n != 3 {
		t.Errorf("counters %d, want 3", n)
	}
	if c, _ := battlefieldCard(g, berta); !c.Tapped {
		t.Error("Berta did not tap")
	}
	if n := countersOn(g, berta, game.CounterPlusOne); n != 0 {
		t.Errorf("Berta got %d counters from her own ability", n)
	}
}
