package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_dmu_b05_cards_test.go — Dominaria United, S82 batch 5 (#5487).
// Each test plays the real catalog entries through the engine.

const (
	dmuB05Academy  = "0cd77ee1-408c-4c51-ac87-24f4fc50083d"
	dmuB05Infantry = "ba8883fc-db24-4c3d-ae36-05085089c9cc"
	dmuB05Djinn    = "dcf28f91-1bf4-48be-a3d0-e061f443fd9e"
	dmuB05Tura     = "4cb3a303-9ded-46d5-90e7-1cc1c65a75da"
	dmuB05Raff     = "41f3acbf-6092-435f-a5af-4d80dd6e1aef"
	dmuB05Rona     = "52e92296-7af9-4e68-8045-40b6d6b73910"
	dmuB05Vohar    = "4380df08-7be4-48ca-9783-d644fd2ea27d"
	dmuB05Hexcatch = "4246b503-0da8-4de7-9343-9dfb3c3672df"
	dmuB05Stenn    = "03d2d7dd-54b6-469e-8b1c-c48db8cc0925"
)

func TestDmuB05EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		dmuB05Academy: "Academy Wall", dmuB05Infantry: "Electrostatic Infantry",
		dmuB05Djinn: "Djinn of the Fountain", dmuB05Tura: "Tura Kennerüd, Skyknight",
		dmuB05Raff: "Raff, Weatherlight Stalwart", dmuB05Rona: "Rona, Sheoldred's Faithful",
		dmuB05Vohar: "Vohar, Vodalian Desecrator", dmuB05Hexcatch: "Vodalian Hexcatcher",
		dmuB05Stenn: "Stenn, Paranoid Partisan",
	} {
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

// dmuB05CastBolt casts a plain instant and a plain sorcery's worth of
// "spell cast" events for the active player and lets them resolve.
func dmuB05CastInstant(t *testing.T, g *game.Game) {
	t.Helper()
	castCatalogSpell(t, g, "Shock Test", "Instant", "", nil)
}

func dmuB05CastSorcery(t *testing.T, g *game.Game) {
	t.Helper()
	castCatalogSpell(t, g, "Ponder Test", "Sorcery", "", nil)
}

// --- Tura Kennerüd -------------------------------------------------------

func TestDmuB05TuraMakesASoldierPerInstantOrSorceryOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Tura Kennerüd, Skyknight", "Legendary Creature — Human Knight", dmuB05Tura, false)
	dmuB05CastInstant(t, g)
	passPriorityAroundTable(t, g)
	dmuB05CastSorcery(t, g)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Soldier"); n != 2 {
		t.Fatalf("an instant and a sorcery made %d Soldiers, want 2", n)
	}
	soldier := findBattlefieldByName(g, "Soldier")
	if card, _ := battlefieldCard(g, soldier); !card.HasColor("W") || !card.IsToken() {
		t.Error("the Soldier is a white token")
	}
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Soldier"); n != 2 {
		t.Errorf("a creature spell made a Soldier: %d", n)
	}
}

// --- Electrostatic Infantry ----------------------------------------------

func TestDmuB05InfantryGrowsPerInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	inf := cmdr2B05Card(g, me.ID, "Electrostatic Infantry", "Creature — Dwarf Wizard", dmuB05Infantry, 1, 2)
	dmuB05CastInstant(t, g)
	passPriorityAroundTable(t, g)
	dmuB05CastSorcery(t, g)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, inf, game.CounterPlusOne); n != 2 {
		t.Fatalf("%d +1/+1 counters after an instant and a sorcery, want 2", n)
	}
	castCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, inf, game.CounterPlusOne); n != 2 {
		t.Errorf("an artifact spell added a counter: %d", n)
	}
	if p := effectivePowerWithCounters(t, g, inf); p != 3 {
		t.Errorf("power %d, want 3", p)
	}
}

func TestDmuB05InfantryIgnoresAnOpponentsSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	inf := cmdr2B05Card(g, me.ID, "Electrostatic Infantry", "Creature — Dwarf Wizard", dmuB05Infantry, 1, 2)
	opp := g.Seats[1]
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventCast, Actor: opp.ID, CardID: uuid.New()}) })
	passPriorityAroundTable(t, g)
	if n := countersOn(g, inf, game.CounterPlusOne); n != 0 {
		t.Errorf("an opponent's cast put %d counters on the Infantry", n)
	}
}

// --- Academy Wall --------------------------------------------------------

func TestDmuB05AcademyWallLootsOnlyOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	msc18Library(me, 8)
	cmdr2B05Card(g, me.ID, "Academy Wall", "Creature — Wall", dmuB05Academy, 0, 5)
	advanceToMain(t, g)
	dmuB05CastInstant(t, g)
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	answerMayChoice(t, g, me.ID, true)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d, want %d after the draw", me.Hand.Size(), hand+1)
	}
	discardFromHand(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d after the discard, want %d", me.Hand.Size(), hand)
	}
	passPriorityAroundTable(t, g)
	dmuB05CastSorcery(t, g)
	passPriorityAroundTable(t, g)
	if latestConfirmFor(g, me.ID) != nil {
		t.Error("a second instant or sorcery in one turn triggered the Wall again")
	}
}

func TestDmuB05AcademyWallMayDecline(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	msc18Library(me, 8)
	cmdr2B05Card(g, me.ID, "Academy Wall", "Creature — Wall", dmuB05Academy, 0, 5)
	advanceToMain(t, g)
	dmuB05CastSorcery(t, g)
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	answerMayChoice(t, g, me.ID, false)
	if me.Hand.Size() != hand {
		t.Errorf("declining changed the hand: %d → %d", hand, me.Hand.Size())
	}
}

// --- Djinn of the Fountain -----------------------------------------------

func dmuB05DjinnMode(t *testing.T, g *game.Game, me *game.Player, mode int) {
	t.Helper()
	dmuB05CastInstant(t, g)
	c := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		passPriorityAroundTable(t, g)
		c = latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	}
	if c == nil {
		t.Fatal("no mode prompt for the Djinn")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestDmuB05DjinnPumpsUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	djinn := cmdr2B05Card(g, me.ID, "Djinn of the Fountain", "Creature — Djinn", dmuB05Djinn, 4, 4)
	dmuB05DjinnMode(t, g, me, 0)
	if p, tough := effectivePTOf(t, g, djinn); p != 5 || tough != 5 {
		t.Errorf("P/T %d/%d, want 5/5", p, tough)
	}
}

func TestDmuB05DjinnBlinksAtTheNextEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	djinn := cmdr2B05Card(g, me.ID, "Djinn of the Fountain", "Creature — Djinn", dmuB05Djinn, 4, 4)
	advanceToMain(t, g)
	dmuB05DjinnMode(t, g, me, 1)
	if zone, _ := cardWhere(g, me, djinn); zone != "exile" {
		t.Fatalf("the Djinn is in %s, want exile", zone)
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Djinn of the Fountain"); n != 1 {
		t.Errorf("%d Djinns on the battlefield at the end step, want 1", n)
	}
}

func TestDmuB05DjinnScries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	msc18Library(me, 6)
	cmdr2B05Card(g, me.ID, "Djinn of the Fountain", "Creature — Djinn", dmuB05Djinn, 4, 4)
	dmuB05DjinnMode(t, g, me, 2)
	if latestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID) == nil && len(g.PendingChoices) == 0 {
		t.Error("the scry mode asked nothing")
	}
}

// --- Raff, Weatherlight Stalwart -----------------------------------------

func TestDmuB05RaffTapsTwoCreaturesToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	msc18Library(me, 6)
	cmdr2B05Card(g, me.ID, "Raff, Weatherlight Stalwart", "Legendary Creature — Human Wizard", dmuB05Raff, 1, 3)
	a := cmdr2B05Card(g, me.ID, "Bear A", "Creature — Bear", "", 2, 2)
	b := cmdr2B05Card(g, me.ID, "Bear B", "Creature — Bear", "", 2, 2)
	dmuB05CastInstant(t, g)
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	answerChooseCards(t, g, me.ID, a, b)
	if !b16Tapped(t, g, a) || !b16Tapped(t, g, b) {
		t.Error("the two Bears were not tapped")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d after paying the tap", me.Hand.Size(), hand+1)
	}
}

func TestDmuB05RaffDecliningTapsNothingAndDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	msc18Library(me, 6)
	cmdr2B05Card(g, me.ID, "Raff, Weatherlight Stalwart", "Legendary Creature — Human Wizard", dmuB05Raff, 1, 3)
	a := cmdr2B05Card(g, me.ID, "Bear A", "Creature — Bear", "", 2, 2)
	cmdr2B05Card(g, me.ID, "Bear B", "Creature — Bear", "", 2, 2)
	dmuB05CastSorcery(t, g)
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	answerChooseCards(t, g, me.ID)
	if b16Tapped(t, g, a) || me.Hand.Size() != hand {
		t.Error("declining tapped a creature or drew")
	}
}

func TestDmuB05RaffNeedsTwoUntappedCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	msc18Library(me, 6)
	cmdr2B05Card(g, me.ID, "Raff, Weatherlight Stalwart", "Legendary Creature — Human Wizard", dmuB05Raff, 1, 3)
	dmuB05CastInstant(t, g)
	passPriorityAroundTable(t, g)
	if c := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID); c != nil {
		t.Error("Raff alone was offered a tap-two choice")
	}
}

func TestDmuB05RaffPumpsAndGrantsVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	raff := cmdr2B05Card(g, me.ID, "Raff, Weatherlight Stalwart", "Legendary Creature — Human Wizard", dmuB05Raff, 1, 3)
	bear := cmdr2B05Card(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	opp := cmdr2B05Card(g, g.Seats[1].ID, "Opposing Bear", "Creature — Bear", "", 2, 2)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, raff, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no mana")
	}
	floatForTest(g, me, "CCCWW")
	b16Activate(t, g, me.ID, raff, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("Bear P/T %d/%d, want 3/3", p, tough)
	}
	if p, _ := effectivePTOf(t, g, raff); p != 2 {
		t.Errorf("Raff power %d, want 2", p)
	}
	if !effectiveAbilitiesContain(t, g, bear, "vigilance") {
		t.Error("the Bear did not gain vigilance")
	}
	if p, _ := effectivePTOf(t, g, opp); p != 2 || effectiveAbilitiesContain(t, g, opp, "vigilance") {
		t.Error("an opponent's creature was pumped")
	}
}

// --- Vohar, Vodalian Desecrator ------------------------------------------

func dmuB05VoharLoot(t *testing.T, discard game.Card) (g *game.Game, me, opp *game.Player) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	msc18Library(me, 8)
	me.Hand.Cards = nil
	discard.InstanceID = uuid.New()
	discard.Owner, discard.Controller = me.ID, me.ID
	me.Hand.PushTop(discard)
	vohar := cmdr2B05Card(g, me.ID, "Vohar, Vodalian Desecrator", "Legendary Creature — Phyrexian Merfolk Wizard", dmuB05Vohar, 1, 2)
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, vohar, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID, discard.InstanceID)
	passPriorityAroundTable(t, g)
	return g, me, opp
}

func TestDmuB05VoharDrainsWhenYouDiscardAnInstant(t *testing.T) {
	g := newCatalogGame(t)
	_ = g
	for _, tc := range []struct {
		name      string
		card      game.Card
		wantDrain bool
	}{
		{"instant", game.Card{Name: "Bolt", TypeLine: "Instant"}, true},
		{"sorcery", game.Card{Name: "Ponder", TypeLine: "Sorcery"}, true},
		{"creature", game.Card{Name: "Bear", TypeLine: "Creature — Bear"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			meLife, oppLife := me.Life, opp.Life
			_, me2, opp2 := dmuB05VoharLoot(t, tc.card)
			_ = g
			_, _ = meLife, oppLife
			if tc.wantDrain {
				if me2.Life != 41 || opp2.Life != 39 {
					t.Errorf("life %d / %d, want 41 / 39", me2.Life, opp2.Life)
				}
			} else if me2.Life != 40 || opp2.Life != 40 {
				t.Errorf("life %d / %d, want 40 / 40", me2.Life, opp2.Life)
			}
		})
	}
}

func TestDmuB05VoharSacrificeCastsAGraveyardSpellThenExilesIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	vohar := cmdr2B05Card(g, me.ID, "Vohar, Vodalian Desecrator", "Legendary Creature — Phyrexian Merfolk Wizard", dmuB05Vohar, 1, 2)
	advanceToMain(t, g)
	spell := graveyardCardOf(g, me, "Ponder Test", "Sorcery", "{U}", "")
	floatForTest(g, me, "CC")
	b16Activate(t, g, me.ID, vohar, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell}},
	})
	passPriorityAroundTable(t, g)
	if zone, _ := cardWhere(g, me, vohar); zone == "battlefield" {
		t.Fatal("Vohar was not sacrificed")
	}
	floatForTest(g, me, "U")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedCost, Strict: true}); err != nil {
		t.Fatalf("casting the granted spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if zone, _ := cardWhere(g, me, spell); zone != "exile" {
		t.Errorf("the spell is in %s, want exile", zone)
	}
}

func TestDmuB05VoharGrantsNothingToACreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	vohar := cmdr2B05Card(g, me.ID, "Vohar, Vodalian Desecrator", "Legendary Creature — Phyrexian Merfolk Wizard", dmuB05Vohar, 1, 2)
	advanceToMain(t, g)
	bear := graveyardCardOf(g, me, "Bear", "Creature — Bear", "{1}{G}", "")
	floatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, vohar, 1, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Error("a creature card was a legal target")
	}
}

// --- Vodalian Hexcatcher -------------------------------------------------

func TestDmuB05HexcatcherPumpsOtherMerfolkOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hex := cmdr2B05Card(g, me.ID, "Vodalian Hexcatcher", "Creature — Merfolk Wizard", dmuB05Hexcatch, 1, 1)
	fish := cmdr2B05Card(g, me.ID, "Fish", "Creature — Merfolk", "", 1, 1)
	bear := cmdr2B05Card(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := cmdr2B05Card(g, g.Seats[1].ID, "Their Fish", "Creature — Merfolk", "", 1, 1)
	for _, tc := range []struct {
		id   uuid.UUID
		p, t int
	}{{hex, 1, 1}, {fish, 2, 2}, {bear, 2, 2}, {theirs, 1, 1}} {
		if p, tough := effectivePTOf(t, g, tc.id); p != tc.p || tough != tc.t {
			t.Errorf("%s: P/T %d/%d, want %d/%d", dmuB05NameOf(g, tc.id), p, tough, tc.p, tc.t)
		}
	}
}

func dmuB05NameOf(g *game.Game, id uuid.UUID) string {
	if c, ok := battlefieldCard(g, id); ok {
		return c.Name
	}
	return id.String()
}

func TestDmuB05HexcatcherSacrificeCountersANoncreatureSpellUnlessPaid(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cmdr2B05Card(g, me.ID, "Vodalian Hexcatcher", "Creature — Merfolk Wizard", dmuB05Hexcatch, 1, 1)
	fish := cmdr2B05Card(g, me.ID, "Fish", "Creature — Merfolk", "", 1, 1)
	hex := findBattlefieldByName(g, "Vodalian Hexcatcher")
	spell := castCatalogSpell(t, g, "Ponder Test", "Sorcery", "", nil)
	if err := g.ActivateCatalogAbility(me.ID, hex, 0, game.ActivateAbilityParams{
		Strict:       true,
		SacrificeIDs: []uuid.UUID{fish},
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: spell}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if zone, _ := cardWhere(g, me, fish); zone == "battlefield" {
		t.Fatal("the Merfolk was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	// Nothing to pay {1} with: the spell is countered.
	if c := latestChoiceOfKindFor(g, game.PendingChoicePayUnless, me.ID); c != nil {
		if err := g.ResolvePayUnless(c.ID, me.ID, false); err != nil {
			t.Fatalf("ResolvePayUnless: %v", err)
		}
		passPriorityAroundTable(t, g)
	} else {
		t.Fatal("no pay-unless prompt for the spell's controller")
	}
	if zone, _ := cardWhere(g, me, spell); zone != "graveyard" {
		t.Errorf("the spell is in %s, want the graveyard (countered)", zone)
	}
}

func TestDmuB05HexcatcherCantTargetACreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hex := cmdr2B05Card(g, me.ID, "Vodalian Hexcatcher", "Creature — Merfolk Wizard", dmuB05Hexcatch, 1, 1)
	bear := castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	if err := g.ActivateCatalogAbility(me.ID, hex, 0, game.ActivateAbilityParams{
		Strict:       true,
		SacrificeIDs: []uuid.UUID{hex},
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Error("a creature spell was a legal target")
	}
}

// --- Rona, Sheoldred's Faithful ------------------------------------------

func TestDmuB05RonaDrainsEachOpponentPerInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Rona, Sheoldred's Faithful", "Legendary Creature — Human Wizard", dmuB05Rona, false)
	dmuB05CastInstant(t, g)
	passPriorityAroundTable(t, g)
	dmuB05CastSorcery(t, g)
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats[1:] {
		if p.Life != 38 {
			t.Errorf("%s life %d, want 38", p.Name, p.Life)
		}
	}
	if me.Life != 40 {
		t.Errorf("Rona's controller lost life: %d", me.Life)
	}
}

func TestDmuB05RonaCastsFromTheGraveyardByDiscardingTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	rona := graveyardCardOf(g, me, "Rona, Sheoldred's Faithful", "Legendary Creature — Human Wizard", "{1}{U}{B}{B}", dmuB05Rona)
	me.Hand.Cards = nil
	var hand []uuid.UUID
	for i := 0; i < 3; i++ {
		id := uuid.New()
		hand = append(hand, id)
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	payMana(t, g, me, "{C}{U}{B}{B}")
	if err := g.CastSpell(me.ID, rona, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "discard_from_graveyard", AltCostIDs: hand[:1], Strict: true,
	}); err == nil {
		t.Fatal("one discarded card paid a two-card cost")
	}
	if !inGraveyard(me, rona) || me.Hand.Size() != 3 {
		t.Fatal("a refused cast moved a card")
	}
	if err := g.CastSpell(me.ID, rona, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "discard_from_graveyard", AltCostIDs: hand[:2], Strict: true,
	}); err != nil {
		t.Fatalf("graveyard cast: %v", err)
	}
	if me.Hand.Size() != 1 {
		t.Errorf("hand %d, want 1 after discarding two", me.Hand.Size())
	}
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, rona) == nil {
		t.Error("Rona did not resolve onto the battlefield")
	}
}

func TestDmuB05RonaCantBeCastFromTheGraveyardWithOneCardInHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	rona := graveyardCardOf(g, me, "Rona, Sheoldred's Faithful", "Legendary Creature — Human Wizard", "{1}{U}{B}{B}", dmuB05Rona)
	me.Hand.Cards = nil
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	payMana(t, g, me, "{C}{U}{B}{B}")
	if err := g.CastSpell(me.ID, rona, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "discard_from_graveyard", AltCostIDs: []uuid.UUID{id}, Strict: true,
	}); err == nil {
		t.Error("cast with one card in hand")
	}
}

// --- Stenn, Paranoid Partisan --------------------------------------------

func dmuB05StennType(t *testing.T, g *game.Game, me *game.Player, word string) uuid.UUID {
	t.Helper()
	id := castCatalogSpell(t, g, "Stenn, Paranoid Partisan", "Legendary Creature — Human Wizard", dmuB05Stenn, nil)
	passPriorityAroundTable(t, g)
	answerAnchorWord(t, g, me.ID, word)
	return id
}

func TestDmuB05StennDiscountsOnlyTheChosenType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := dmuB05StennType(t, g, me, "instant")
	if got := g.ChosenOptionOf(id); got != "instant" {
		t.Fatalf("Stenn chose %q, want instant", got)
	}
	if got := priceInHand(t, g, me, "Lightning Bolt", "Instant", "{1}{R}"); got != 1 {
		t.Errorf("instant: %d, want 1", got)
	}
	if got := priceInHand(t, g, me, "Divination", "Sorcery", "{2}{U}"); got != 3 {
		t.Errorf("sorcery (not chosen): %d, want 3", got)
	}
	if got := priceInHand(t, g, me, "Grizzly Bears", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("creature: %d, want 2", got)
	}
}

func TestDmuB05StennOffersNeitherCreatureNorLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Stenn, Paranoid Partisan", "Legendary Creature — Human Wizard", dmuB05Stenn, nil)
	passPriorityAroundTable(t, g)
	c := pendingOfKind(g, game.PendingChoiceOptionPick)
	if c == nil {
		t.Fatal("no as-enters choice")
	}
	_ = me
	for _, o := range c.PickOptions {
		if o.Label == "creature" || o.Label == "land" {
			t.Errorf("%q is offered", o.Label)
		}
	}
	if len(c.PickOptions) == 0 {
		t.Error("no options offered")
	}
}

func TestDmuB05StennBlinksAtTheNextEndStepAndChoosesAgain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	id := dmuB05StennType(t, g, me, "instant")
	floatForTest(g, me, "CCWU")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if zone, _ := cardWhere(g, me, id); zone != "exile" {
		t.Fatalf("Stenn is in %s, want exile", zone)
	}
	if got := priceInHand(t, g, me, "Lightning Bolt", "Instant", "{1}{R}"); got != 2 {
		t.Errorf("a discount outlived Stenn: %d, want 2", got)
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Stenn, Paranoid Partisan"); n != 1 {
		t.Fatalf("%d Stennes on the battlefield after the end step, want 1", n)
	}
	answerAnchorWord(t, g, me.ID, "sorcery")
	if got := priceInHand(t, g, me, "Divination", "Sorcery", "{2}{U}"); got != 2 {
		t.Errorf("the new choice does not discount: %d, want 2", got)
	}
}

func TestDmuB05HexcatcherCounterFailsWhenTheControllerPays(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hex := cmdr2B05Card(g, me.ID, "Vodalian Hexcatcher", "Creature — Merfolk Wizard", dmuB05Hexcatch, 1, 1)
	spell := castCatalogSpell(t, g, "Ponder Test", "Sorcery", "", nil)
	floatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, hex, 0, game.ActivateAbilityParams{
		Strict:       true,
		SacrificeIDs: []uuid.UUID{hex},
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: spell}},
	}); err != nil {
		t.Fatalf("activate (sacrificing the Hexcatcher itself): %v", err)
	}
	passPriorityAroundTable(t, g)
	c := latestChoiceOfKindFor(g, game.PendingChoicePayUnless, me.ID)
	if c == nil {
		t.Fatal("no pay-unless prompt")
	}
	if err := g.ResolvePayUnless(c.ID, me.ID, true); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	passPriorityAroundTable(t, g)
	// The {1} was spent, so the counter did nothing: the sorcery resolved.
	if len(me.ManaPool) != 0 {
		t.Errorf("the {1} was not paid: pool %v", me.ManaPool)
	}
	if g.Stack.Size() != 0 {
		t.Errorf("stack size %d after paying, want the spell resolved", g.Stack.Size())
	}
}
