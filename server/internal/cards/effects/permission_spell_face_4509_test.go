package effects

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// permission_spell_face_4509_test.go — #4509 at the card level: a cast
// permission phrased about the SPELL ("cast an artifact, instant, or
// sorcery spell", "a spell with mana value 5 or less") judges the face
// being cast (CR 601.3e, 709.3a, 712.11c, 715.3a), and one phrased about
// the CARD ("each instant and sorcery card in your graveyard", "nonland
// permanent cards in your graveyard") judges the card in its zone
// (CR 715.4, 712.8a). The engine half is game/permission_spell_face_test.go.

// creatureArtifactMDFC is the Tergrid, God of Fright // Tergrid's Lantern
// shape, uncatalogued: a creature front and an artifact back.
func creatureArtifactMDFC(owner uuid.UUID) game.Card {
	return mdfcCard(owner, "", game.LayoutModalDFC,
		game.Face{Name: "Fixture God", TypeLine: "Legendary Creature — God", ManaCost: "{2}{B}",
			Colors: []string{"B"}, Power: 4, Toughness: 5},
		game.Face{Name: "Fixture Lantern", TypeLine: "Legendary Artifact", ManaCost: "{1}{B}", Colors: []string{"B"}},
	)
}

// cheapFrontExpensiveBack is the Valki, God of Lies // Tibalt, Cosmic
// Impostor shape: a mana value 2 creature front and a mana value 7
// planeswalker back.
func cheapFrontExpensiveBack(owner uuid.UUID) game.Card {
	return mdfcCard(owner, "", game.LayoutModalDFC,
		game.Face{Name: "Fixture Valki", TypeLine: "Legendary Creature — God", ManaCost: "{1}{B}",
			Colors: []string{"B"}, Power: 2, Toughness: 1},
		game.Face{Name: "Fixture Tibalt", TypeLine: "Legendary Planeswalker — Tibalt", ManaCost: "{5}{B}{R}",
			Colors: []string{"B", "R"}, StartingLoyalty: 5},
	)
}

// expensiveFrontCheapBack is the Beanstalk Giant // Fertile Footsteps
// shape: a mana value 7 creature whose Adventure is a mana value 3
// sorcery.
func expensiveFrontCheapBack(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(), Layout: game.LayoutAdventure, Owner: owner, Controller: owner,
		Faces: []game.Face{
			{Name: "Fixture Beanstalk", TypeLine: "Creature — Giant", ManaCost: "{6}{G}", Colors: []string{"G"},
				Power: 7, Toughness: 7},
			{Name: "Fixture Footsteps", TypeLine: "Sorcery — Adventure", ManaCost: "{2}{G}", Colors: []string{"G"}},
		},
	}
	c.SetFace(0)
	return c
}

// castMoveFaces is every face the bot enumerator offers a cast of card
// as, in the order it lists them.
func castMoveFaces(g *game.Game, seat, card uuid.UUID) []int {
	var out []int
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type != legal.TypeCastSpell || m.Source != card {
			continue
		}
		var p struct {
			Face int `json:"face"`
		}
		if json.Unmarshal(m.Params, &p) == nil && !slices.Contains(out, p.Face) {
			out = append(out, p.Face)
		}
	}
	return out
}

// stackCardOf is the card on the stack with this ID.
func stackCardOf(g *game.Game, id uuid.UUID) (game.Card, bool) {
	for _, c := range g.Stack.Cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return game.Card{}, false
}

// permissionOver is the cast permission that opens card out of zone for
// seat right now, or nil.
func permissionOver(g *game.Game, seat, card uuid.UUID, zone game.ZoneKind) *game.CastPermission {
	var out *game.CastPermission
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(card); ok {
			out = g.CastPermissionForLocked(seat, c, zone)
		}
	})
	return out
}

// pickResolutionExit answers the CR 616.1 choice between the replacements
// of a resolving spell's trip to the graveyard with the one whose words
// start with `prefix`.
func pickResolutionExit(t *testing.T, g *game.Game, chooser uuid.UUID, prefix string) {
	t.Helper()
	c := latestOptionPickFor(g, chooser)
	if c == nil {
		t.Fatalf("no CR 616.1 choice for %s: %+v", chooser, g.PendingChoices)
	}
	for i, o := range c.PickOptions {
		if strings.HasPrefix(o.Label, prefix) {
			if err := g.ResolveOptionPick(c.ID, chooser, i); err != nil {
				t.Fatalf("ResolveOptionPick(%q): %v", o.Label, err)
			}
			return
		}
	}
	t.Fatalf("no option starting %q among %+v", prefix, c.PickOptions)
}

// --- Bilbo, Thief in the Night -------------------------------------------

// "Cast an artifact, instant, or sorcery SPELL": Bonecrusher Giant in the
// graveyard is a creature card, and Stomp is an instant spell. Bilbo
// offers the card, and only as Stomp (the card's ruling of 2026-06-29).
// Stomp resolves, and its controller chooses between the Adventure's exile
// and Bilbo's replacement (CR 616.1): the Adventure keeps the creature
// castable from exile (CR 715.3d).
func TestBilboOffersAnAdventureAndCastsItAsTheAdventure(t *testing.T) {
	g, me, opp, bilbo := bilboTable(t)
	giant := bonecrusherGiant(me.ID)
	id := giant.InstanceID
	g.WithWriteLock(func() { me.Graveyard.PushTop(giant) })
	bilboOffer(t, g, me, opp, bilbo, id)

	floatForTest(g, me, "RR")
	if faces := castMoveFaces(g, me.ID, id); !slices.Equal(faces, []int{1}) {
		t.Fatalf("the bot is offered faces %v, want Stomp alone [1]", faces)
	}
	creature := game.CastSpellParams{FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedCost, Strict: true}
	if err := g.CastSpell(me.ID, id, creature); err == nil {
		t.Fatal("Bilbo let the creature half be cast: it is not an artifact, instant or sorcery spell")
	}
	stomp := creature
	stomp.Face = 1
	stomp.Targets = []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	if err := g.CastSpell(me.ID, id, stomp); err != nil {
		t.Fatalf("cast Stomp for {1}{R} less Bilbo's {1}: %v", err)
	}
	castInsideTheResolution(t, g, id)
	spell, ok := stackCardOf(g, id)
	if !ok || spell.ActiveFace != 1 || !spell.IsInstant() || spell.IsCreature() || spell.Name != "Stomp" {
		t.Fatalf("on the stack: %+v, want Stomp, an instant (CR 715.3b)", spell)
	}

	before := opp.Life
	for i := 0; i < 8 && g.Stack.Contains(id) && latestOptionPickFor(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	pickResolutionExit(t, g, me.ID, "Adventure")
	if opp.Life != before-2 {
		t.Errorf("opponent life %d, want %d", opp.Life, before-2)
	}
	if !inExile(g, id) {
		t.Fatal("the resolved Stomp is not in exile")
	}
	perm := g.CastPermissionOnCardByIDForEffect(id)
	if face, ok := perm.NamedFace(); !perm.Granted() || !ok || face != 0 {
		t.Errorf("CR 715.3d: no permission to cast the creature from exile: %+v", perm)
	}
}

// Countered, the Adventure never resolved, so CR 715.3d has nothing to
// say, and Bilbo's replacement exiles it: the creature cannot be cast
// later (the same ruling).
func TestBilboCounteredAdventureIsExiledWithoutTheCreatureGrant(t *testing.T) {
	g, me, opp, bilbo := bilboTable(t)
	giant := bonecrusherGiant(me.ID)
	id := giant.InstanceID
	g.WithWriteLock(func() { me.Graveyard.PushTop(giant) })
	bilboOffer(t, g, me, opp, bilbo, id)

	floatForTest(g, me, "R")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedCost, Strict: true, Face: 1,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("cast Stomp: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(id); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	if me.Graveyard.Contains(id) || !inExile(g, id) {
		t.Fatal("the countered Stomp went to the graveyard, want exile")
	}
	if perm := g.CastPermissionOnCardByIDForEffect(id); perm.Granted() {
		t.Errorf("a countered Adventure left a permission to cast the creature: %+v", perm)
	}
}

// CR 712.11b, 712.11c: a modal DFC is offered for its artifact back face,
// and cast only as that face.
func TestBilboOffersAModalDFCsArtifactBackFace(t *testing.T) {
	g, me, opp, bilbo := bilboTable(t)
	card := creatureArtifactMDFC(me.ID)
	id := card.InstanceID
	g.WithWriteLock(func() { me.Graveyard.PushTop(card) })
	bilboOffer(t, g, me, opp, bilbo, id)

	floatForTest(g, me, "BB")
	front := game.CastSpellParams{FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedCost, Strict: true}
	if err := g.CastSpell(me.ID, id, front); err == nil {
		t.Fatal("Bilbo let the creature front be cast")
	}
	back := front
	back.Face = 1
	if err := g.CastSpell(me.ID, id, back); err != nil {
		t.Fatalf("cast the artifact back face for {1}{B} less Bilbo's {1}: %v", err)
	}
	if spell, ok := stackCardOf(g, id); !ok || !spell.IsArtifact() || spell.IsCreature() || spell.Name != "Fixture Lantern" {
		t.Fatalf("on the stack: %+v, want the artifact back face", spell)
	}
}

// A Town whose Adventure is a sorcery is a land card, and its Adventure is
// a sorcery spell: Bilbo offers it, as the Adventure (CR 715.3a).
func TestBilboOffersATownForItsAdventure(t *testing.T) {
	g, me, opp, bilbo := bilboTable(t)
	town := landAdventureCard(me.ID, "", "Fixture Town", "Fixture Overture", "{1}{U}", []string{"U"})
	id := town.InstanceID
	g.WithWriteLock(func() { me.Graveyard.PushTop(town) })
	bilboOffer(t, g, me, opp, bilbo, id)

	floatForTest(g, me, "U")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedCost, Strict: true, Face: 1,
	}); err != nil {
		t.Fatalf("cast the Town's Adventure: %v", err)
	}
	if spell, ok := stackCardOf(g, id); !ok || !spell.IsSorcery() || spell.IsLand() {
		t.Fatalf("on the stack: %+v, want the sorcery Adventure", spell)
	}
}

// --- permissions phrased about the card ----------------------------------

// Lier: "each instant and sorcery CARD in your graveyard has flashback".
// Bonecrusher Giant in a graveyard is a creature card (CR 715.4; the
// card's ruling of 2019-10-04), so it has no flashback, Stomp or not.
func TestLierGivesAnAdventurerCardNoFlashback(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMainForCost(t, g)
	pushCatalogPermanent(g, me.ID, "Lier, Disciple of the Drowned", "Legendary Creature — Human Wizard",
		"f224db3d-cdd2-43f2-b625-4b861efa6449", false)
	giant := bonecrusherGiant(me.ID)
	g.WithWriteLock(func() { me.Graveyard.PushTop(giant) })
	instant := gyCard(me, "Test Instant", "Instant", "{1}{U}", "")
	floatForTest(g, me, "UURR")

	if perm := permissionOver(g, me.ID, giant.InstanceID, game.ZoneGraveyard); perm != nil {
		t.Errorf("Bonecrusher Giant has flashback under Lier: %+v", perm)
	}
	if perm := permissionOver(g, me.ID, instant, game.ZoneGraveyard); perm == nil {
		t.Error("the instant card has no flashback under Lier")
	}
	if faces := castMoveFaces(g, me.ID, giant.InstanceID); len(faces) != 0 {
		t.Errorf("the bot is offered Bonecrusher Giant as faces %v", faces)
	}
	if faces := castMoveFaces(g, me.ID, instant); !slices.Equal(faces, []int{0}) {
		t.Errorf("the instant card's faces %v, want [0]", faces)
	}
}

// Past in Flames: "each instant and sorcery CARD in your graveyard gains
// flashback". The set is the cards as they are in the graveyard: Stomp
// gains nothing, because Bonecrusher Giant there is a creature card
// (CR 715.4), and nor does a creature card whose back face is a sorcery
// (CR 712.8a); an instant // sorcery split card does (CR 709.4c).
func TestPastInFlamesGrantsByTheCardInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	giant := bonecrusherGiant(me.ID)
	mdfc := mdfcCard(me.ID, "", game.LayoutModalDFC,
		game.Face{Name: "Fixture Glyphweaver", TypeLine: "Creature — Human Cleric", ManaCost: "{2}{W}",
			Colors: []string{"W"}, Power: 2, Toughness: 2},
		game.Face{Name: "Fixture Vanity", TypeLine: "Sorcery", ManaCost: "{5}{B}{B}", Colors: []string{"B"}},
	)
	split := game.Card{InstanceID: uuid.New(), Layout: game.LayoutSplit, Owner: me.ID, Controller: me.ID,
		Faces: []game.Face{
			{Name: "Fixture Commit", TypeLine: "Instant", ManaCost: "{3}{U}", Colors: []string{"U"}},
			{Name: "Fixture Memory", TypeLine: "Sorcery", ManaCost: "{4}{U}{U}", Colors: []string{"U"}},
		}}
	split.SettleImported()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(giant)
		me.Graveyard.PushTop(mdfc)
		me.Graveyard.PushTop(split)
	})

	castCatalogSpell(t, g, "Past in Flames", "Sorcery", pastInFlamesOracle, nil)
	passPriorityAroundTable(t, g)

	if perm := permissionOver(g, me.ID, giant.InstanceID, game.ZoneGraveyard); perm != nil {
		t.Errorf("Bonecrusher Giant gained flashback: %+v", perm)
	}
	if perm := permissionOver(g, me.ID, mdfc.InstanceID, game.ZoneGraveyard); perm != nil {
		t.Errorf("a creature-front modal DFC gained flashback: %+v", perm)
	}
	if perm := permissionOver(g, me.ID, split.InstanceID, game.ZoneGraveyard); perm == nil {
		t.Error("an instant // sorcery split card gained no flashback")
	}
}

// Six: "nonland permanent CARDS in your graveyard have retrace". Bonecrusher
// Giant is a creature card there, so it has retrace, and retrace names no
// type: it may be cast as Stomp (CR 715.3).
func TestSixRetracesAnAdventurerCardAsItsAdventure(t *testing.T) {
	g, me, _, land := sixTable(t)
	opp := g.Seats[1]
	giant := bonecrusherGiant(me.ID)
	g.WithWriteLock(func() { me.Graveyard.PushTop(giant) })

	var faces []int
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(giant.InstanceID)
		if perm := g.CastPermissionForLocked(me.ID, c, game.ZoneGraveyard); perm != nil {
			faces = game.CastableFacesUnder(c, perm, me.ID)
		}
	})
	if !slices.Equal(faces, []int{0, 1}) {
		t.Fatalf("retrace opens faces %v, want both [0 1]", faces)
	}
	payMana(t, g, me, "{R}{R}")
	if err := g.CastSpell(me.ID, giant.InstanceID, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "retrace", AltCostIDs: []uuid.UUID{land}, Strict: true, Face: 1,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("retrace Stomp: %v", err)
	}
	if spell, ok := stackCardOf(g, giant.InstanceID); !ok || spell.Name != "Stomp" {
		t.Fatalf("on the stack: %+v, want Stomp", spell)
	}
}

// Underworld Breach: "each nonland CARD in your graveyard has escape". A
// Town is a land card there, so it has no escape, though its Adventure is
// a sorcery.
func TestUnderworldBreachGivesATownNoEscape(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMainForCost(t, g)
	pushCatalogPermanent(g, me.ID, "Underworld Breach", "Enchantment", "27e0948b-9916-473b-8d8c-a51bdfbc7457", false)
	town := landAdventureCard(me.ID, "", "Fixture Town", "Fixture Overture", "{1}{U}", []string{"U"})
	g.WithWriteLock(func() { me.Graveyard.PushTop(town) })
	for i := 0; i < 3; i++ {
		gyCard(me, "Fuel", "Sorcery", "{1}", "")
	}
	giant := bonecrusherGiant(me.ID)
	g.WithWriteLock(func() { me.Graveyard.PushTop(giant) })

	if perm := permissionOver(g, me.ID, town.InstanceID, game.ZoneGraveyard); perm != nil {
		t.Errorf("the Town has escape under Underworld Breach: %+v", perm)
	}
	// A creature card does, and escape names no type either.
	var faces []int
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(giant.InstanceID)
		if perm := g.CastPermissionForLocked(me.ID, c, game.ZoneGraveyard); perm != nil {
			faces = game.CastableFacesUnder(c, perm, me.ID)
		}
	})
	if !slices.Equal(faces, []int{0, 1}) {
		t.Errorf("Bonecrusher Giant's escape faces %v, want both [0 1]", faces)
	}
}

// --- permissions phrased about the spell ---------------------------------

// Vizier of the Menagerie: "you may cast creature SPELLS from the top of
// your library". An adventurer card on top is cast as the creature and not
// as its Adventure (the Giant Killer ruling of 2019-10-04).
func TestVizierCastsOnlyTheCreatureHalfOfAnAdventurerCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	toMainForCost(t, g)
	pushCatalogPermanent(g, me.ID, "Vizier of the Menagerie", "Creature — Snake Cleric",
		"4f890d42-e4f1-4eed-aa6c-718865de8927", false)
	giant := bonecrusherGiant(me.ID)
	g.WithWriteLock(func() { me.Library.PushTop(giant) })

	floatForTest(g, me, "RRR")
	if faces := castMoveFaces(g, me.ID, giant.InstanceID); !slices.Equal(faces, []int{0}) {
		t.Fatalf("the bot is offered faces %v, want the creature alone [0]", faces)
	}
	if err := g.CastSpell(me.ID, giant.InstanceID, game.CastSpellParams{
		FromZone: "library", Strict: true, Face: 1,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err == nil {
		t.Fatal("Stomp was cast from the top of the library under a creature-spells permission")
	}
	if err := g.CastSpell(me.ID, giant.InstanceID, game.CastSpellParams{FromZone: "library", Strict: true}); err != nil {
		t.Fatalf("cast the creature from the top of the library: %v", err)
	}
}

// Crucible of Worlds: "you may play LANDS from your graveyard". A modal
// DFC whose back face is a land is played as that face (the MDFC ruling of
// 2020-09-25: "any face that fits the criteria of that effect").
func TestCrucibleOfWorldsPlaysAModalDFCsLandFace(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMainForCost(t, g)
	pushCatalogPermanent(g, me.ID, "Crucible of Worlds", "Artifact", "33c722cf-b4bf-431f-aefd-ee96241a7fbf", false)
	card := seaGateRestoration(me.ID)
	id := card.InstanceID
	g.WithWriteLock(func() { me.Graveyard.PushTop(card) })

	if faces := castMoveFaces(g, me.ID, id); !slices.Equal(faces, []int{1}) {
		t.Fatalf("the bot is offered faces %v, want the land face alone [1]", faces)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", Face: 1}); err != nil {
		t.Fatalf("play Sea Gate, Reborn from the graveyard: %v", err)
	}
	// Sea Gate, Reborn asks "you may pay 3 life; if you don't, it
	// enters tapped".
	answerEntryPayLife(t, g, me.ID, false)
	land, ok := battlefieldCard(g, id)
	if !ok || land.ActiveFace != 1 || !land.IsLand() {
		t.Fatalf("on the battlefield: %+v, want Sea Gate, Reborn", land)
	}
}

// Surtland Elementalist: "you may cast an instant or sorcery SPELL from
// your hand without paying its mana cost". Bonecrusher Giant in hand is
// offered for Stomp, free; the free cast of the creature is not offered.
func TestSurtlandCastsAnAdventureFreeFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	giant := bonecrusherGiant(me.ID)
	id := giant.InstanceID
	g.WithWriteLock(func() { me.Hand.PushTop(giant) })
	bear := handCardFull(me, "Bear", "Creature — Bear", "{1}{G}", "", []string{"G"})
	surtland := pushCatalogPermanent(g, me.ID, "Surtland Elementalist", "Creature — Giant Wizard", surtlandOracle, false)

	declareAttack(t, g, opp.ID, surtland)
	passPriorityAroundTable(t, g)
	c := mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, id)
	if slices.Contains(c.DiscardOptions, bear) {
		t.Fatal("a creature card with no instant or sorcery spell is offered")
	}
	stomp := game.CastSpellParams{AlternativeCost: game.AltCostKeyGrantedCost, Strict: true, Face: 1,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}
	creature := game.CastSpellParams{AlternativeCost: game.AltCostKeyGrantedCost, Strict: true}
	if err := g.CastSpell(me.ID, id, creature); err == nil {
		t.Fatal("the creature half was cast free under an instant-or-sorcery offer")
	}
	if err := g.CastSpell(me.ID, id, stomp); err != nil {
		t.Fatalf("cast Stomp free: %v", err)
	}
	if spell, ok := stackCardOf(g, id); !ok || spell.Name != "Stomp" {
		t.Fatalf("on the stack: %+v, want Stomp", spell)
	}
}

// Rishkar's Expertise: "you may cast a spell with mana value 5 or less from
// your hand without paying its mana cost". The mana value is the spell's:
// a mana value 2 front whose back is a mana value 7 planeswalker is cast
// free only as its front, and a mana value 7 creature whose Adventure is
// mana value 3 is cast free as its Adventure (CR 601.3e, 712.11c, 715.3a).
func TestRishkarsExpertiseJudgesTheSpellsManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	valki := cheapFrontExpensiveBack(me.ID)
	beanstalk := expensiveFrontCheapBack(me.ID)
	g.WithWriteLock(func() {
		me.Hand.PushTop(valki)
		me.Hand.PushTop(beanstalk)
	})

	castCatalogSpell(t, g, "Rishkar's Expertise", "Sorcery", "97407cd0-2bd2-4074-94d3-4ec3d243fa78", nil)
	passPriorityAroundTable(t, g)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, valki.InstanceID, beanstalk.InstanceID)
	free := func(face int) game.CastSpellParams {
		return game.CastSpellParams{AlternativeCost: game.AltCostKeyGrantedCost, Strict: true, Face: face}
	}
	if err := g.CastSpell(me.ID, valki.InstanceID, free(1)); err == nil {
		t.Fatal("the mana value 7 back face was cast free under a mana value 5 cap")
	}
	if err := g.CastSpell(me.ID, beanstalk.InstanceID, free(0)); err == nil {
		t.Fatal("the mana value 7 creature was cast free under a mana value 5 cap")
	}
	if err := g.CastSpell(me.ID, beanstalk.InstanceID, free(1)); err != nil {
		t.Fatalf("cast the mana value 3 Adventure free: %v", err)
	}
	if spell, ok := stackCardOf(g, beanstalk.InstanceID); !ok || spell.Name != "Fixture Footsteps" {
		t.Fatalf("on the stack: %+v, want the Adventure", spell)
	}
}
