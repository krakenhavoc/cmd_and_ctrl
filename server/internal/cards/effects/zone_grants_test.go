package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// zone_grants_test.go — #3233, ADR 0093's amendment of 2026-10-13: a
// static that gives an ACTIVATED ABILITY to the cards in a hand or a
// graveyard. Satoru Umezawa's ninjutsu is the headline; Rhet-Tomb
// Mystic, Tectonic Reformation, Sedris and Solemn Doomguide ride the
// same seam. The engine half (the stamp, the readers) is exercised
// through the cards, because the seam is a layer-pass step whose only
// observable output is the rows an object has.

const (
	satoruUmezawaOracle       = "ed1d21ba-ddfe-4f82-9e37-93f9e2c7ba33"
	rhetTombMysticOracle      = "e445825a-e70e-4796-99be-4a7c4080b8cc"
	tectonicReformationOracle = "10da681b-50bd-4a1b-9934-f0c90e0af55c"
	sedrisOracle              = "48054407-b5e1-431c-ad03-c12a31d0d9a8"
	solemnDoomguideOracle     = "c3e4a6ce-4a36-43a4-8e80-3c16a9101b79"

	// zgPlainOracle is an oracle ID no card registers, so a hand card
	// that carries it has no abilities of its own: every row it shows
	// is a granted one.
	zgPlainOracle = "zone-grant-fixture-vanilla"
)

// zgRows reads a card's activated rows and their origins the way every
// consumer does, under the read lock (which recomputes a stale layer
// pass first).
func zgRows(g *game.Game, id uuid.UUID) (rows []game.ActivatedAbilityShape, origins game.AbilityOrigins) {
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			return
		}
		rows, origins = game.ActivatedAbilitiesWithOrigins(c)
	})
	return rows, origins
}

func zgLabels(rows []game.ActivatedAbilityShape) []string {
	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = r.Label
	}
	return out
}

func zgHandCard(p *game.Player, name, typeLine string) uuid.UUID {
	return pushCatalogHandCard(p, name, typeLine, zgPlainOracle)
}

func zgGraveyardCard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: zgPlainOracle,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

func zgLeaveBattlefield(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	var err error
	g.WithWriteLock(func() { err = g.DestroyPermanentForEffect(id) })
	if err != nil {
		t.Fatalf("destroy: %v", err)
	}
}

// The grant appears when Satoru enters and disappears when he leaves,
// reaches only creature cards in HIS CONTROLLER's hand, and never lands
// on a permanent.
func TestSatoruGivesCreatureCardsInHandNinjutsuWhileHeIsOut(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := zgHandCard(me, "Plain Creature", "Creature — Human Ninja")
	sorcery := zgHandCard(me, "Plain Sorcery", "Sorcery")
	theirs := zgHandCard(opp, "Their Creature", "Creature — Bear")

	if rows, _ := zgRows(g, mine); len(rows) != 0 {
		t.Fatalf("before Satoru, rows = %v, want none", zgLabels(rows))
	}

	satoru := b12Push(g, me.ID, "Satoru Umezawa", "Legendary Creature — Human Ninja", satoruUmezawaOracle, 2, 4)

	rows, origins := zgRows(g, mine)
	if len(rows) != 1 || !strings.HasPrefix(rows[0].Label, "Ninjutsu {2}{U}{B}") {
		t.Fatalf("rows with Satoru out = %v, want one Ninjutsu {2}{U}{B}", zgLabels(rows))
	}
	if got := origins.Ref(0); got != "grant:satoru-umezawa/ninjutsu:0:0" {
		t.Errorf("ref = %q, want grant:satoru-umezawa/ninjutsu:0:0", got)
	}
	if o := origins.At(0); o.GrantedBy != satoru || o.Activator != me.ID || !o.Granted() {
		t.Errorf("origin = %+v, want granted by Satoru for %s", o, me.ID)
	}
	if rows, _ := zgRows(g, sorcery); len(rows) != 0 {
		t.Errorf("a sorcery card has rows %v, want none", zgLabels(rows))
	}
	if rows, _ := zgRows(g, theirs); len(rows) != 0 {
		t.Errorf("an opponent's creature card has rows %v, want none — the grant says \"your hand\"", zgLabels(rows))
	}
	if rows, _ := zgRows(g, satoru); len(rows) != 0 {
		t.Errorf("Satoru himself has rows %v, want none — he is not a card in a hand", zgLabels(rows))
	}

	zgLeaveBattlefield(t, g, satoru)
	if rows, _ := zgRows(g, mine); len(rows) != 0 {
		t.Errorf("after Satoru left, rows = %v, want none (CR 113.6, 613.6)", zgLabels(rows))
	}
}

// A creature card that ARRIVES in the hand while Satoru is out has the
// ability at once: the arrival alone moves nothing on the battlefield,
// so the layer version is bumped for it (layerVersionBump).
func TestACardDrawnWhileSatoruIsOutHasTheGrant(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Satoru Umezawa", "Legendary Creature — Human Ninja", satoruUmezawaOracle, 2, 4)
	// Make sure the pass has run once with the hand as it is, so the
	// next read is stale only because of the draw.
	zgRows(g, uuid.New())

	top := uuid.New()
	me.Library.PushTop(game.Card{
		InstanceID: top, Name: "Drawn Creature", TypeLine: "Creature — Human Ninja", OracleID: zgPlainOracle,
		Owner: me.ID, Controller: me.ID,
	})
	var err error
	g.WithWriteLock(func() { err = g.DrawNForEffect(me.ID, 1) })
	if err != nil {
		t.Fatalf("draw: %v", err)
	}
	if !me.Hand.Contains(top) {
		t.Fatal("the creature was not drawn")
	}
	if rows, _ := zgRows(g, top); len(rows) != 1 {
		t.Errorf("a creature drawn with Satoru out has rows %v, want the granted ninjutsu", zgLabels(rows))
	}
}

// Each Satoru is its own instance of the grant (CR 113.2c), numbered in
// the ref so an announcement names the row it meant.
func TestTwoSatorusGiveTwoNinjutsuRows(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	card := zgHandCard(me, "Plain Creature", "Creature — Human Ninja")
	b12Push(g, me.ID, "Satoru Umezawa", "Legendary Creature — Human Ninja", satoruUmezawaOracle, 2, 4)
	b12Push(g, me.ID, "Satoru Umezawa", "Legendary Creature — Human Ninja", satoruUmezawaOracle, 2, 4)

	rows, origins := zgRows(g, card)
	if len(rows) != 2 {
		t.Fatalf("rows = %v, want two", zgLabels(rows))
	}
	if origins.Ref(0) != "grant:satoru-umezawa/ninjutsu:0:0" || origins.Ref(1) != "grant:satoru-umezawa/ninjutsu:0:1" {
		t.Errorf("refs = %q, %q, want :0:0 and :0:1", origins.Ref(0), origins.Ref(1))
	}
}

// CR 613.6: a Satoru whose abilities have been removed grants nothing.
func TestASilencedSatoruGrantsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	card := zgHandCard(me, "Plain Creature", "Creature — Human Ninja")
	satoru := b12Push(g, me.ID, "Satoru Umezawa", "Legendary Creature — Human Ninja", satoruUmezawaOracle, 2, 4)
	if rows, _ := zgRows(g, card); len(rows) != 1 {
		t.Fatalf("rows with Satoru out = %v, want the granted ninjutsu", zgLabels(rows))
	}

	enchant(t, g, "Darksteel Mutation", darksteelMutationOracle, satoru)

	if rows, _ := zgRows(g, card); len(rows) != 0 {
		t.Errorf("rows under a Darksteel Mutation'd Satoru = %v, want none", zgLabels(rows))
	}
}

// The grant is data on the card: a Clone keeps it, and so does a
// snapshot round trip (the restore recomputes the layers).
func TestHandGrantsSurviveCloneAndSnapshotRestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	card := zgHandCard(me, "Plain Creature", "Creature — Human Ninja")
	b12Push(g, me.ID, "Satoru Umezawa", "Legendary Creature — Human Ninja", satoruUmezawaOracle, 2, 4)
	if rows, _ := zgRows(g, card); len(rows) != 1 {
		t.Fatalf("setup: rows = %v", zgLabels(rows))
	}

	if rows, _ := zgRows(g.Clone(), card); len(rows) != 1 {
		t.Errorf("clone: rows = %v, want the granted ninjutsu", zgLabels(rows))
	}
	restored := restoreRoundTrip(t, g, false)
	if rows, origins := zgRows(restored, card); len(rows) != 1 || origins.Ref(0) != "grant:satoru-umezawa/ninjutsu:0:0" {
		t.Errorf("restored: rows = %v, want the granted ninjutsu", zgLabels(rows))
	}
}

// zgNinjutsuBoard is ninjutsuBoard for a plain creature card, with
// Satoru out: one attacker swinging at seat 1, the cursor in the
// declare-blockers step, and the granted ninjutsu's price floating.
func zgNinjutsuBoard(t *testing.T, g *game.Game) (ninja, attacker, satoru uuid.UUID, me, opp *game.Player) {
	t.Helper()
	me, opp = g.Seats[0], g.Seats[1]
	attacker = b12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
	satoru = b12Push(g, me.ID, "Satoru Umezawa", "Legendary Creature — Human Ninja", satoruUmezawaOracle, 2, 4)
	ninja = pushNinjaToHand(me, "Plain Ninja", "Creature — Human Ninja", zgPlainOracle, 2, 2)
	declareAttack(t, g, opp.ID, attacker)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{U}{B}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	return ninja, attacker, satoru, me, opp
}

// The headline. A creature card with no ninjutsu of its own, in the
// hand of Satoru's controller, activates the GRANTED ninjutsu: the
// unblocked attacker goes back to hand as the cost and the creature
// arrives from hand tapped and attacking the same player.
func TestAGrantedNinjutsuReturnsTheAttackerAndPutsTheCreatureInAttacking(t *testing.T) {
	g := newCatalogGame(t)
	ninja, attacker, _, me, opp := zgNinjutsuBoard(t, g)
	seeded := seedLibrary(me, "Look1", "Look2", "Look3", "Filler1")

	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{
		Ref:       "grant:satoru-umezawa/ninjutsu:0:0",
		ReturnIDs: []uuid.UUID{attacker},
	}); err != nil {
		t.Fatalf("activate the granted ninjutsu: %v", err)
	}
	if !me.Hand.Contains(attacker) {
		t.Error("the unblocked attacker was not returned to hand")
	}

	// Satoru's own trigger goes on the stack above the ability and
	// resolves first: look at three, take one.
	passPriorityAroundTable(t, g)
	choice := chooseCardsChoiceFor(g, me.ID)
	if choice == nil {
		t.Fatal("no look-at-three prompt — Satoru did not trigger off the granted ninjutsu")
	}
	if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{seeded[0]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	order := putInLibraryChoiceFor(g, me.ID)
	if order == nil {
		t.Fatal("no order prompt for the two cards left on the bottom")
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, []uuid.UUID{seeded[2], seeded[1]}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if !me.Hand.Contains(seeded[0]) {
		t.Error("the chosen card did not reach the hand")
	}

	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ninja) {
		t.Fatal("the creature did not arrive on the battlefield")
	}
	c, ok := g.LookupCardForEffect(ninja)
	if !ok {
		t.Fatal("cannot look the creature up")
	}
	if !c.Tapped {
		t.Error("the creature entered untapped — ninjutsu puts it onto the battlefield tapped")
	}
	if c.AttackingTarget != opp.ID {
		t.Errorf("the creature attacks %s, want %s", c.AttackingTarget, opp.ID)
	}
	if n := attackEventsFor(g, ninja); n != 0 {
		t.Errorf("%d EventAttack for the ninja, want 0 (CR 506.3c)", n)
	}
}

// A stale ref is refused before anything is paid, as for any grant.
func TestAGrantedNinjutsuRefusesAStaleRef(t *testing.T) {
	g := newCatalogGame(t)
	ninja, attacker, _, me, _ := zgNinjutsuBoard(t, g)

	err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{
		Ref:       "grant:satoru-umezawa/ninjutsu:0:1",
		ReturnIDs: []uuid.UUID{attacker},
	})
	if err == nil {
		t.Fatal("a ref naming a second Satoru's row was accepted with only one out")
	}
	if !g.Battlefield.Contains(attacker) {
		t.Error("the refused activation returned the attacker anyway")
	}
}

// The look trigger is once each turn: a second ninjutsu in the same
// turn, granted or not, draws no second look. Its first firing is also
// proven for a PRINTED ninjutsu (Ninja of the Deep Hours), since the
// trigger watches the activation, not the grant.
func TestSatorusLookTriggersOncePerTurnOffAnyNinjutsu(t *testing.T) {
	g := newCatalogGame(t)
	ninja, attacker, satoru, me, _ := zgNinjutsuBoard(t, g)
	second := b12Creature(g, me.ID, "Second Rat", "Creature — Rat", 1, 1)
	_ = second
	printed := pushNinjaToHand(me, "Ninja of the Deep Hours", "Creature — Human Ninja", ninjaOfTheDeepHoursOracle, 2, 2)
	seedLibrary(me, "Look1", "Look2", "Look3", "Filler1")

	// First: the PRINTED ninjutsu triggers Satoru.
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, printed, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{attacker},
	}); err != nil {
		t.Fatalf("activate the printed ninjutsu: %v", err)
	}
	if triggerOnStack(g, satoru) == nil {
		t.Fatal("Satoru did not trigger off a printed ninjutsu")
	}
	passPriorityAroundTable(t, g)
	if choice := chooseCardsChoiceFor(g, me.ID); choice == nil {
		t.Fatal("no look prompt for the first ninjutsu")
	} else if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{choice.ChooseCards[0]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if order := putInLibraryChoiceFor(g, me.ID); order != nil {
		ids := append([]uuid.UUID(nil), order.ScryCards...)
		if err := g.ResolvePutInLibrary(order.ID, me.ID, ids, nil); err != nil {
			t.Fatalf("ResolvePutInLibrary: %v", err)
		}
	}
	passPriorityAroundTable(t, g)

	// Second, same turn: the granted ninjutsu on the plain creature. The
	// Ninja of the Deep Hours is now the unblocked attacker.
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{U}{B}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{printed},
	}); err != nil {
		t.Fatalf("activate the granted ninjutsu: %v", err)
	}
	if triggerOnStack(g, satoru) != nil {
		t.Error("Satoru triggered a second time in one turn — \"this ability triggers only once each turn\"")
	}
}

// The grant also covers the graveyard, with the recipient test the card
// prints: Sedris gives every creature card, Solemn Doomguide only four
// of the creature types.
func TestGraveyardGrantsUnearth(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cleric := zgGraveyardCard(me, "Plain Cleric", "Creature — Human Cleric")
	bear := zgGraveyardCard(me, "Plain Bear", "Creature — Bear")
	land := zgGraveyardCard(me, "Plain Land", "Land")

	sedris := b12Push(g, me.ID, "Sedris, the Traitor King", "Legendary Creature — Zombie Warrior", sedrisOracle, 5, 5)
	for _, tc := range []struct {
		id   uuid.UUID
		name string
		want int
	}{{cleric, "cleric", 1}, {bear, "bear", 1}, {land, "land", 0}} {
		rows, _ := zgRows(g, tc.id)
		if len(rows) != tc.want || (tc.want == 1 && !strings.HasPrefix(rows[0].Label, "Unearth {2}{B}")) {
			t.Errorf("Sedris: %s rows = %v, want %d Unearth {2}{B}", tc.name, zgLabels(rows), tc.want)
		}
	}
	zgLeaveBattlefield(t, g, sedris)
	if rows, _ := zgRows(g, bear); len(rows) != 0 {
		t.Errorf("after Sedris left, rows = %v, want none", zgLabels(rows))
	}

	b12Push(g, me.ID, "Solemn Doomguide", "Creature — Tiefling Cleric", solemnDoomguideOracle, 3, 3)
	for _, tc := range []struct {
		id   uuid.UUID
		name string
		want int
	}{{cleric, "cleric", 1}, {bear, "bear", 0}} {
		rows, _ := zgRows(g, tc.id)
		if len(rows) != tc.want || (tc.want == 1 && !strings.HasPrefix(rows[0].Label, "Unearth {1}{B}")) {
			t.Errorf("Solemn Doomguide: %s rows = %v, want %d Unearth {1}{B}", tc.name, zgLabels(rows), tc.want)
		}
	}
}

// A granted unearth works from the graveyard: the card returns to the
// battlefield with haste and is exiled at the end step.
func TestAGrantedUnearthReturnsTheCreatureFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMain(t, g)
	bear := zgGraveyardCard(me, "Plain Bear", "Creature — Bear")
	b12Push(g, me.ID, "Sedris, the Traitor King", "Legendary Creature — Zombie Warrior", sedrisOracle, 5, 5)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{B}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}

	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{
		Ref: "grant:sedris-the-traitor-king/unearth:0:0",
	}); err != nil {
		t.Fatalf("activate the granted unearth: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the creature did not return to the battlefield")
	}
}

// Rhet-Tomb Mystic and Tectonic Reformation grant ordinary cycling, and
// the cycled card is a cycling source in every way a printed one is.
func TestGrantedCyclingCyclesTheCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMain(t, g)
	creature := zgHandCard(me, "Plain Creature", "Creature — Bear")
	land := zgHandCard(me, "Plain Land", "Land")
	b12Push(g, me.ID, "Rhet-Tomb Mystic", "Creature — Zombie Bird Wizard", rhetTombMysticOracle, 1, 2)
	b12Push(g, me.ID, "Tectonic Reformation", "Enchantment", tectonicReformationOracle, 0, 0)

	if rows, _ := zgRows(g, creature); len(rows) != 1 || !strings.HasPrefix(rows[0].Label, "Cycling {1}{U}") {
		t.Fatalf("creature rows = %v, want Cycling {1}{U}", zgLabels(rows))
	}
	if rows, _ := zgRows(g, land); len(rows) != 1 || !strings.HasPrefix(rows[0].Label, "Cycling {R}") {
		t.Fatalf("land rows = %v, want Cycling {R}", zgLabels(rows))
	}

	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	handBefore := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{
		Ref: "grant:tectonic-reformation/cycling:0:0",
	}); err != nil {
		t.Fatalf("cycle the land: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(land) || !me.Graveyard.Contains(land) {
		t.Error("the cycled land is not in the graveyard")
	}
	if got := me.Hand.Size(); got != handBefore {
		t.Errorf("hand %d -> %d, want the same (discard one, draw one)", handBefore, got)
	}
	// Mystic's grant is off the land: it is not a creature card.
	for _, c := range me.Graveyard.Cards {
		if c.InstanceID == land {
			if rows, _ := zgRows(g, land); len(rows) != 0 {
				t.Errorf("a land in the graveyard has rows %v, want none (hand grant only)", zgLabels(rows))
			}
		}
	}
}

// The satoru label helper agrees with the two constructors that build
// the label, so the trigger cannot drift from the keyword.
func TestANinjutsuAbilityReadsBothConstructorsLabels(t *testing.T) {
	for _, ab := range []ActivatedAbility{Ninjutsu("{1}{U}"), CommanderNinjutsu("{1}{U}")} {
		if !ANinjutsuAbility(game.Event{Label: ab.Label}, nil, game.Characteristic{}, nil) {
			t.Errorf("%q is not read as a ninjutsu ability", ab.Label)
		}
	}
	for _, label := range []string{"Cycling {2}", "Unearth {B}", ""} {
		if ANinjutsuAbility(game.Event{Label: label}, nil, game.Characteristic{}, nil) {
			t.Errorf("%q is read as a ninjutsu ability", label)
		}
	}
}

// The boot-time guards: a bundle given to cards outside the battlefield
// holds only activated abilities that function from that zone, and a
// static's zone must match its bundle's.
func TestZoneGrantRegistrationRefusals(t *testing.T) {
	hand := []game.ZoneKind{game.ZoneHand}
	ok := ActivatedAbility{Label: "x", Zones: hand}
	expectRegisterPanic(t, "more than activated", func() {
		checkGrants("Fixture", []AbilityGrant{{Key: "fixture/z1", Zone: game.ZoneHand, Text: "x", Activated: []ActivatedAbility{ok},
			Mana: []ManaAbility{{Cost: ManaAbilityCost{Tap: true}, Produced: "{G}"}}}})
	})
	expectRegisterPanic(t, "does not function from hand", func() {
		checkGrants("Fixture", []AbilityGrant{{Key: "fixture/z2", Zone: game.ZoneHand, Text: "x",
			Activated: []ActivatedAbility{{Label: "x"}}}})
	})
	expectRegisterPanic(t, "needs a permanent", func() {
		checkGrants("Fixture", []AbilityGrant{{Key: "fixture/z3", Zone: game.ZoneHand, Text: "x",
			Activated: []ActivatedAbility{{Label: "x", Zones: hand, Cost: TapCost()}}}})
	})
	expectRegisterPanic(t, "walks only", func() {
		checkGrants("Fixture", []AbilityGrant{{Key: "fixture/z4", Zone: game.ZoneLibrary, Text: "x", Activated: []ActivatedAbility{ok}}})
	})
	// The static half.
	s := GrantAbilitiesInZone(game.ZoneHand, nil, "fixture/z1")
	expectRegisterPanic(t, "no AppliesTo", func() { checkStaticGrants("Fixture", "static", []game.StaticAbility{s}) })
	s = GrantAbilitiesInZone(game.ZoneLibrary, CardsYouOwn(creatureCard), "fixture/z1")
	expectRegisterPanic(t, "walks only", func() { checkStaticGrants("Fixture", "static", []game.StaticAbility{s}) })
	s = GrantAbilitiesInZone(game.ZoneHand, CardsYouOwn(creatureCard), "fixture/z1")
	expectRegisterPanic(t, "only a battlefield permanent", func() { checkStaticGrants("Fixture", "emblem static", []game.StaticAbility{s}) })
	checkStaticGrants("Fixture", "static", []game.StaticAbility{s}) // the good shape is accepted

	// A bundle for the wrong zone is caught by the catalog-wide walk.
	problems, _ := grantKeyProblems(map[string]*game.CardDef{
		"card-a":                         {Static: []game.StaticAbility{GrantAbilitiesInZone(game.ZoneGraveyard, nil, "fixture/in-hand")}},
		game.GrantKey("fixture/in-hand"): {Activated: []game.ActivatedAbilityShape{{Label: "Ninjutsu {1}", Zones: hand}}},
	})
	if len(problems) != 1 || !strings.Contains(problems[0], "does not function from there") {
		t.Errorf("problems = %v, want the zone mismatch", problems)
	}
}

// A zone change ends the grant on the spot (CR 400.7), without waiting
// for the next layer pass: the card that leaves the hand is a new
// object, and a stale row on it would be offered before the pass ran.
func TestAZoneGrantEndsWhenTheCardChangesZone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	card := zgHandCard(me, "Plain Creature", "Creature — Human Ninja")
	b12Push(g, me.ID, "Satoru Umezawa", "Legendary Creature — Human Ninja", satoruUmezawaOracle, 2, 4)
	if rows, _ := zgRows(g, card); len(rows) != 1 {
		t.Fatalf("setup: rows = %v, want the granted ninjutsu", zgLabels(rows))
	}

	moved, err := game.MoveCard(me.Hand, me.Graveyard, card)
	if err != nil {
		t.Fatalf("MoveCard: %v", err)
	}
	if rows := game.ActivatedAbilitiesForCard(moved); len(rows) != 0 {
		t.Errorf("the card that left the hand still has rows %v", zgLabels(rows))
	}
}
