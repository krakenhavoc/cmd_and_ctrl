package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// disturb_test.go — ADR 0107 §4 (#1855), disturb end to end on real
// cards: the cast out of the graveyard (game/disturb_test.go pins the
// engine half on a fixture), the back face's "If [this] would be put
// into a graveyard from anywhere, exile it instead" (DisturbedExile),
// and the cards that need more than the cast.
//
// Every card here goes down the import road (deck.ToGameCard), for the
// reason siege_transform_test.go gives: a disturb card's whole point is
// its back face, which only the importer puts on game.Card.Faces.

// disturbFace is one printed face of a disturb row.
type disturbFace struct {
	name, typeLine, cost, power, toughness string
}

// disturbRow is the Scryfall record of a disturb card in the shape the
// bulk dump ships every `transform` card: a null top-level cost and the
// real data on the faces.
func disturbRow(oracleID string, colors []string, front, back disturbFace) cards.Card {
	face := func(f disturbFace) cards.CardFace {
		return cards.CardFace{Name: f.name, TypeLine: f.typeLine, ManaCost: f.cost, Power: f.power, Toughness: f.toughness, Colors: colors}
	}
	return cards.Card{
		ID:        uuid.New(),
		Name:      front.name + " // " + back.name,
		Layout:    "transform",
		TypeLine:  front.typeLine + " // " + back.typeLine,
		OracleID:  uuid.MustParse(oracleID),
		CardFaces: []cards.CardFace{face(front), face(back)},
	}
}

func baithookAnglerRow() cards.Card {
	return disturbRow(baithookAnglerOracleID, []string{"U"},
		disturbFace{"Baithook Angler", "Creature — Human Peasant", "{1}{U}", "2", "1"},
		disturbFace{"Hook-Haunt Drifter", "Creature — Spirit", "", "1", "2"})
}

func drogskolInfantryRow() cards.Card {
	return disturbRow(drogskolInfantryOracleID, []string{"W"},
		disturbFace{"Drogskol Infantry", "Creature — Spirit Soldier", "{1}{W}", "2", "2"},
		disturbFace{"Drogskol Armaments", "Enchantment — Aura", "", "", ""})
}

func malevolentHermitRow() cards.Card {
	return disturbRow(malevolentHermitOracleID, []string{"U"},
		disturbFace{"Malevolent Hermit", "Creature — Human Wizard", "{1}{U}", "2", "1"},
		disturbFace{"Benevolent Geist", "Creature — Spirit Wizard", "", "2", "2"})
}

// importToGraveyard runs a row down the import road into a player's
// graveyard, front face up (CR 712.8a), and walks to a main phase.
func importToGraveyard(t *testing.T, g *game.Game, row cards.Card, p *game.Player) uuid.UUID {
	t.Helper()
	c := deck.ToGameCard(row, false)
	c.Owner, c.Controller = p.ID, p.ID
	p.Graveyard.PushTop(c)
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return c.InstanceID
}

// disturb casts a card out of its owner's graveyard for its disturb
// cost, leaving the face unset as the client's graveyard menu does.
func disturb(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID, targets ...game.TargetRef) {
	t.Helper()
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "disturb", Targets: targets}); err != nil {
		t.Fatalf("disturb: %v", err)
	}
}

// cardWhere returns the zone a card is in and the card itself.
func cardWhere(g *game.Game, p *game.Player, id uuid.UUID) (string, game.Card) {
	for name, z := range map[string]*game.Zone{
		"battlefield": g.Battlefield, "stack": g.Stack, "exile": g.Exile,
		"graveyard": p.Graveyard, "hand": p.Hand, "library": p.Library,
	} {
		for _, c := range z.Cards {
			if c.InstanceID == id {
				return name, c
			}
		}
	}
	return "", game.Card{}
}

// A disturbed Hook-Haunt Drifter is the back face on the battlefield,
// with the back face's flying, and exiled rather than put into a
// graveyard when it dies.
func TestDisturbedCreatureIsExiledWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := importToGraveyard(t, g, baithookAnglerRow(), me)

	disturb(t, g, me, id)
	passPriorityAroundTable(t, g)
	zone, c := cardWhere(g, me, id)
	if zone != "battlefield" || c.ActiveFace != 1 || c.Name != "Hook-Haunt Drifter" {
		t.Fatalf("after the disturb cast the card is in %s as face %d (%q), want the back face on the battlefield", zone, c.ActiveFace, c.Name)
	}
	if !game.HasKeyword(&c, "flying") {
		t.Error("Hook-Haunt Drifter has no flying")
	}
	if c.ManaValue() != 2 {
		t.Errorf("Hook-Haunt Drifter's mana value = %d, want 2 from Baithook Angler's {1}{U} (CR 712.8e)", c.ManaValue())
	}

	destroy(t, g, id)
	if zone, c := cardWhere(g, me, id); zone != "exile" {
		t.Fatalf("the destroyed disturbed creature went to %s, want exile", zone)
	} else if c.ActiveFace != 0 {
		t.Errorf("the exiled card is face %d, want front face up (CR 712.8a)", c.ActiveFace)
	}
}

// Scryfall's keyword list is the union over both faces ("Flying",
// "Disturb" for Baithook Angler). The importer splits it per face, so
// the front face is not a flier from hand and the disturbed back face
// does not carry the front face's disturb (CR 712.8d, 712.8e).
func TestDisturbCardKeywordsBelongToTheirFace(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	row := baithookAnglerRow()
	row.Keywords = []string{"Flying", "Disturb"}
	row.CardFaces[0].OracleText = "Disturb {1}{U} (You may cast this card from your graveyard transformed for its disturb cost.)"
	row.CardFaces[1].OracleText = "Flying\nIf Hook-Haunt Drifter would be put into a graveyard from anywhere, exile it instead."
	front := deck.ToGameCard(row, false)
	if game.HasKeyword(&front, "flying") {
		t.Errorf("Baithook Angler flies from hand: keywords %v", front.Keywords)
	}

	id := importToGraveyard(t, g, row, me)
	disturb(t, g, me, id)
	passPriorityAroundTable(t, g)
	_, c := cardWhere(g, me, id)
	if !game.HasKeyword(&c, "flying") {
		t.Errorf("Hook-Haunt Drifter has no flying: %v", c.Effective().Abilities)
	}
	for _, a := range c.Effective().Abilities {
		if a == "disturb" {
			t.Errorf("Hook-Haunt Drifter carries the front face's disturb: %v", c.Effective().Abilities)
		}
	}
}

// The clause is the BACK face's, so it is read only while that face is
// up (CR 712.8a). A Baithook Angler cast from hand, front face up, dies
// into its owner's graveyard — where it can be disturbed.
func TestFrontFaceGoesToTheGraveyardAsUsual(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := importToHand(baithookAnglerRow(), me)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Baithook Angler: %v", err)
	}
	passPriorityAroundTable(t, g)
	destroy(t, g, id)
	if zone, _ := cardWhere(g, me, id); zone != "graveyard" {
		t.Fatalf("a dead Baithook Angler went to %s, want the graveyard", zone)
	}
	// And from there, disturb.
	disturb(t, g, me, id)
	if zone, c := cardWhere(g, me, id); zone != "stack" || c.ActiveFace != 1 {
		t.Errorf("the disturb cast left the card in %s as face %d", zone, c.ActiveFace)
	}
}

// "From anywhere" includes the stack: a disturbed spell that is
// countered is exiled, not returned to the graveyard to be disturbed
// again.
func TestCounteredDisturbedSpellIsExiled(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := importToGraveyard(t, g, baithookAnglerRow(), me)
	disturb(t, g, me, id)
	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(id); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	if zone, _ := cardWhere(g, me, id); zone != "exile" {
		t.Fatalf("the countered disturbed spell went to %s, want exile", zone)
	}
}

// A disturbed Aura targets as it is cast, through the back face's
// EnchantCreature clause, pumps its creature, and is exiled when it
// falls off (CR 704.5m puts it into a graveyard; the back face's
// clause replaces that).
func TestDisturbedAuraEnchantsAndIsExiledWhenItFallsOff(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := importToGraveyard(t, g, drogskolInfantryRow(), me)
	bear := auraBear(g, me.ID)

	// No target: the Aura spell has one to name (CR 303.4a).
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "disturb"}); err == nil {
		t.Fatal("a disturbed Aura was cast with no creature to enchant")
	}
	disturb(t, g, me, id, game.TargetRef{Kind: game.TargetCard, ID: bear})
	passPriorityAroundTable(t, g)
	zone, aura := cardWhere(g, me, id)
	if zone != "battlefield" || aura.Name != "Drogskol Armaments" || aura.AttachedTo.ID != bear {
		t.Fatalf("the Aura is in %s as %q attached to %v, want Drogskol Armaments on the bear", zone, aura.Name, aura.AttachedTo.ID)
	}
	_, host := cardWhere(g, me, bear)
	if eff := host.Effective(); eff.Power != 4 || eff.Toughness != 4 {
		t.Errorf("the enchanted bear is %d/%d, want 4/4", eff.Power, eff.Toughness)
	}

	killCreature(t, g, me.ID, bear)
	if zone, _ := cardWhere(g, me, id); zone != "exile" {
		t.Fatalf("the fallen-off disturbed Aura went to %s, want exile", zone)
	}
}

// Malevolent Hermit, the row's Waiting card: the sacrifice is the cost,
// so the Hermit is in the graveyard at once, and Benevolent Geist's
// static is the counter gate ADR 0106 PR 2 built.
func TestMalevolentHermitDisturbsIntoBenevolentGeist(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := importToGraveyard(t, g, malevolentHermitRow(), me)
	disturb(t, g, me, id)
	passPriorityAroundTable(t, g)
	zone, geist := cardWhere(g, me, id)
	if zone != "battlefield" || geist.Name != "Benevolent Geist" {
		t.Fatalf("the Hermit's disturb left it in %s as %q", zone, geist.Name)
	}

	// A noncreature spell this player controls can't be countered while
	// the Geist is out — and can once it has gone.
	counterABolt := func() string {
		bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", "4457ed35-7c10-48c8-9776-456485fdf070",
			[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID}})
		g.WithWriteLock(func() { _ = g.CounterTargetForEffect(bolt) })
		zone, _ := cardWhere(g, me, bolt)
		passPriorityAroundTable(t, g)
		return zone
	}
	if zone := counterABolt(); zone != "stack" {
		t.Errorf("a noncreature spell under Benevolent Geist was countered (now in %s)", zone)
	}
	destroy(t, g, id)
	if zone, _ := cardWhere(g, me, id); zone != "exile" {
		t.Errorf("the destroyed Benevolent Geist went to %s, want exile", zone)
	}
	if zone := counterABolt(); zone != "graveyard" {
		t.Errorf("with Benevolent Geist gone, the countered bolt is in %s, want the graveyard", zone)
	}
}

// Every offer in the catalog that casts a face names a face the card
// has an entry for: a disturb card file with no back-face entry would
// cast a back face the engine can run nothing of.
func TestEveryFaceCastingOfferHasABackFaceEntry(t *testing.T) {
	n := 0
	for _, spec := range All() {
		key := spec.OracleID
		for _, ac := range spec.AlternativeCosts {
			if ac.CastsFace == 0 {
				continue
			}
			n++
			back := game.CatalogKeyForFace(key, ac.CastsFace)
			if !Has(back) {
				t.Errorf("%s offers %q casting face %d, but %s is not registered", spec.Name, ac.Key, ac.CastsFace, back)
			}
		}
	}
	if n == 0 {
		t.Fatal("no catalog card offers a face-casting cost; the probe found nothing to check")
	}
}

// Register refuses the shapes the cast path cannot read.
func TestRegisterRefusesAFaceCastingOfferOnABackFace(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil || !strings.Contains(r.(string), "back face") {
			t.Errorf("recover() = %v, want a refusal naming the back face", r)
		}
	}()
	checkCastsFace(Spec{OracleID: "x#1", Name: "Back"}, Disturb("{1}"))
}
