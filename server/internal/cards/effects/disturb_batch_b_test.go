package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// disturb_batch_b_test.go — the second batch of disturb cards (#1855,
// ADR 0107 §4): each card's disturb cast lands its back face, and each
// non-keyword ability on either face does what the card prints. The
// cast and the exile clause themselves are pinned in disturb_test.go.

func catgeistRow() cards.Card {
	return disturbRow(mischievousCatgeistOracleID, []string{"U"},
		disturbFace{"Mischievous Catgeist", "Creature — Cat Spirit", "{1}{U}", "1", "1"},
		disturbFace{"Catlike Curiosity", "Enchantment — Aura", "", "", ""})
}

func distractingGeistRow() cards.Card {
	return disturbRow(distractingGeistOracleID, []string{"W"},
		disturbFace{"Distracting Geist", "Creature — Spirit", "{2}{W}", "2", "1"},
		disturbFace{"Clever Distraction", "Enchantment — Aura", "", "", ""})
}

func gutterSkulkerRow() cards.Card {
	return disturbRow(gutterSkulkerOracleID, []string{"U"},
		disturbFace{"Gutter Skulker", "Creature — Spirit", "{3}{U}", "3", "3"},
		disturbFace{"Gutter Shortcut", "Enchantment — Aura", "", "", ""})
}

func covetousCastawayRow() cards.Card {
	return disturbRow(covetousCastawayOracleID, []string{"U"},
		disturbFace{"Covetous Castaway", "Creature — Human", "{1}{U}", "1", "3"},
		disturbFace{"Ghostly Castigator", "Creature — Spirit", "", "3", "4"})
}

func devotedGrafkeeperRow() cards.Card {
	return disturbRow(devotedGrafkeeperOracleID, []string{"W", "U"},
		disturbFace{"Devoted Grafkeeper", "Creature — Human Peasant", "{W}{U}", "2", "1"},
		disturbFace{"Departed Soulkeeper", "Creature — Spirit", "", "3", "1"})
}

func chaplainOfAlmsRow() cards.Card {
	return disturbRow(chaplainOfAlmsOracleID, []string{"W"},
		disturbFace{"Chaplain of Alms", "Creature — Human Cleric", "{W}", "1", "1"},
		disturbFace{"Chapel Shieldgeist", "Creature — Spirit Cleric", "", "2", "1"})
}

// frontFaceInPlay runs a row down the import road onto the battlefield
// front face up, ready to attack and block.
func frontFaceInPlay(g *game.Game, row cards.Card, p *game.Player) uuid.UUID {
	c := deck.ToGameCard(row, false)
	c.Owner, c.Controller = p.ID, p.ID
	return pushBattlefieldCardWithTimestamp(g, c)
}

// disturbOnto disturbs an Aura card onto `host` and resolves it,
// failing unless the back face is attached to the host.
func disturbOnto(t *testing.T, g *game.Game, p *game.Player, row cards.Card, host uuid.UUID) uuid.UUID {
	t.Helper()
	id := importToGraveyard(t, g, row, p)
	disturb(t, g, p, id, game.TargetRef{Kind: game.TargetCard, ID: host})
	passPriorityAroundTable(t, g)
	zone, aura := cardWhere(g, p, id)
	if zone != "battlefield" || aura.ActiveFace != 1 || aura.AttachedTo.ID != host {
		t.Fatalf("%s: disturbed into %s as face %d (%q) attached to %v, want its back face on the host",
			row.Name, zone, aura.ActiveFace, aura.Name, aura.AttachedTo.ID)
	}
	return id
}

// disturbIntoPlay disturbs a creature card and resolves it, failing
// unless its back face is on the battlefield.
func disturbIntoPlay(t *testing.T, g *game.Game, p *game.Player, row cards.Card, back string) uuid.UUID {
	t.Helper()
	id := importToGraveyard(t, g, row, p)
	disturb(t, g, p, id)
	passPriorityAroundTable(t, g)
	if zone, c := cardWhere(g, p, id); zone != "battlefield" || c.Name != back {
		t.Fatalf("%s: disturbed into %s as %q, want %s on the battlefield", row.Name, zone, c.Name, back)
	}
	return id
}

// pickIfAsked answers the chooser's open pick_target prompt, if any.
func pickIfAsked(t *testing.T, g *game.Game, chooser uuid.UUID, ids ...uuid.UUID) {
	t.Helper()
	p := latestPickTarget(g, chooser)
	if p == nil {
		return
	}
	refs := make([]game.TargetRef, 0, len(ids))
	for _, id := range ids {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	if err := g.ResolvePickTargets(p.ID, chooser, refs); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
}

func tappedNow(g *game.Game, id uuid.UUID) bool {
	c, ok := g.LookupCardForEffect(id)
	return ok && c.Tapped
}

// --- Mischievous Catgeist // Catlike Curiosity -------------------

func TestMischievousCatgeistDrawsOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cat := frontFaceInPlay(g, catgeistRow(), me)
	hand := me.Hand.Size()
	attackWith(t, g, opp.ID, cat)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("Mischievous Catgeist connected and drew %d, want 1", got)
	}
}

func TestCatlikeCuriosityGivesTheEnchantedCreatureTheDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := auraBear(g, me.ID)
	disturbOnto(t, g, me, catgeistRow(), bear)
	hand := me.Hand.Size()
	attackWith(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("the creature enchanted by Catlike Curiosity connected and drew %d, want 1", got)
	}
}

// --- Distracting Geist // Clever Distraction ---------------------

func TestDistractingGeistTapsADefendersCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	geist := frontFaceInPlay(g, distractingGeistRow(), me)
	victim := brCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	bystander := brCreature(g, g.Seats[2].ID, "Other Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, geist)
	pickIfAsked(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if !tappedNow(g, victim) {
		t.Error("Distracting Geist's attack did not tap the defending player's creature")
	}
	if tappedNow(g, bystander) {
		t.Error("a creature of a player who is not defending was tapped")
	}
}

func TestCleverDistractionGivesTheEnchantedCreatureTheTap(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := auraBear(g, me.ID)
	disturbOnto(t, g, me, distractingGeistRow(), bear)
	victim := brCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, bear)
	pickIfAsked(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if !tappedNow(g, victim) {
		t.Error("the creature enchanted by Clever Distraction attacked and tapped nothing")
	}
}

// --- Gutter Skulker // Gutter Shortcut ---------------------------

func TestGutterSkulkerIsUnblockableOnlyWhileAttackingAlone(t *testing.T) {
	// Alone: no block.
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	skulker := frontFaceInPlay(g, gutterSkulkerRow(), me)
	wall := brCreature(g, opp.ID, "Wall", "Creature — Wall", 0, 8)
	brAttack(t, g, skulker)
	brRefusal(t, g.DeclareBlocker(wall, skulker), game.BlockReasonCantBeBlocked)

	// With a second attacker it is not attacking alone.
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	skulker = frontFaceInPlay(g, gutterSkulkerRow(), me)
	friend := brCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	wall = brCreature(g, opp.ID, "Wall", "Creature — Wall", 0, 8)
	brAttack(t, g, skulker, friend)
	if err := g.DeclareBlocker(wall, skulker); err != nil {
		t.Errorf("Gutter Skulker attacking beside another creature could not be blocked: %v", err)
	}
}

func TestGutterShortcutMakesTheEnchantedCreatureUnblockableAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := auraBear(g, me.ID)
	disturbOnto(t, g, me, gutterSkulkerRow(), bear)
	wall := brCreature(g, opp.ID, "Wall", "Creature — Wall", 0, 8)
	brAttack(t, g, bear)
	brRefusal(t, g.DeclareBlocker(wall, bear), game.BlockReasonCantBeBlocked)
}

// --- Covetous Castaway // Ghostly Castigator ---------------------

func TestCovetousCastawayMillsThreeWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castaway := frontFaceInPlay(g, covetousCastawayRow(), me)
	library := me.Library.Size()
	destroy(t, g, castaway)
	passPriorityAroundTable(t, g)
	if got := library - me.Library.Size(); got != 3 {
		t.Errorf("Covetous Castaway's death milled %d cards, want 3", got)
	}
	if zone, _ := cardWhere(g, me, castaway); zone != "graveyard" {
		t.Errorf("the front face died into %s, want the graveyard", zone)
	}
}

func TestGhostlyCastigatorShufflesChosenGraveyardCardsAway(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := seedGraveyardCard(t, g, "Card A", "Sorcery", "")
	b := seedGraveyardCard(t, g, "Card B", "Sorcery", "")
	kept := seedGraveyardCard(t, g, "Card C", "Sorcery", "")
	id := importToGraveyard(t, g, covetousCastawayRow(), me)
	library := me.Library.Size()

	disturb(t, g, me, id)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	pickIfAsked(t, g, me.ID, a, b)
	passPriorityAroundTable(t, g)

	if zone, c := cardWhere(g, me, id); zone != "battlefield" || c.Name != "Ghostly Castigator" {
		t.Fatalf("the disturb cast left the card in %s as %q", zone, c.Name)
	}
	for _, card := range []uuid.UUID{a, b} {
		if zone, _ := cardWhere(g, me, card); zone != "library" {
			t.Errorf("a chosen graveyard card is in %s, want the library", zone)
		}
	}
	if zone, _ := cardWhere(g, me, kept); zone != "graveyard" {
		t.Errorf("an unchosen graveyard card is in %s, want the graveyard", zone)
	}
	if got := me.Library.Size() - library; got != 2 {
		t.Errorf("the library grew by %d, want 2", got)
	}
}

// --- Devoted Grafkeeper // Departed Soulkeeper -------------------

func TestDevotedGrafkeeperMillsTwoAsItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := importToHand(devotedGrafkeeperRow(), me)
	advanceTo(t, g, game.StepPrecombatMain)
	library := me.Library.Size()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Devoted Grafkeeper: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := library - me.Library.Size(); got != 2 {
		t.Errorf("Devoted Grafkeeper's entry milled %d, want 2", got)
	}
}

func TestDevotedGrafkeeperTapsWhenYouCastFromYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	frontFaceInPlay(g, devotedGrafkeeperRow(), me)
	victim := brCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := brCreature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	// From hand: no trigger.
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", "4457ed35-7c10-48c8-9776-456485fdf070",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Fatal("a spell cast from hand triggered Devoted Grafkeeper")
	}
	passPriorityAroundTable(t, g)

	// From the graveyard: tap target creature you don't control.
	looting := seedGraveyardCard(t, g, "Faithless Looting", "Sorcery", "3d6fa57a-aa53-4b5c-b8af-a7612c823117")
	if err := g.CastSpell(me.ID, looting, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("a flashback cast did not trigger Devoted Grafkeeper")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
		t.Error("Devoted Grafkeeper's trigger accepted a creature its controller controls")
	}
	pickIfAsked(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if !tappedNow(g, victim) {
		t.Error("Devoted Grafkeeper's trigger did not tap the opponent's creature")
	}
}

func TestDepartedSoulkeeperBlocksOnlyFlyers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	soulkeeper := disturbIntoPlay(t, g, me, devotedGrafkeeperRow(), "Departed Soulkeeper")
	walker := brCreature(g, opp.ID, "Walker", "Creature — Bear", 2, 2)
	flyer := brCreature(g, opp.ID, "Flyer", "Creature — Bird", 2, 2, "flying")

	advanceToDeclareAttackersOf(t, g, 1)
	for _, a := range []uuid.UUID{walker, flyer} {
		if err := g.DeclareAttacker(a, me.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	brRefusal(t, g.DeclareBlocker(soulkeeper, walker), game.BlockReasonCantBlockAttacker)
	if err := g.DeclareBlocker(soulkeeper, flyer); err != nil {
		t.Errorf("Departed Soulkeeper could not block a flyer: %v", err)
	}
}

// --- Chaplain of Alms // Chapel Shieldgeist -----------------------

func TestChaplainOfAlmsHasWard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	chaplain := frontFaceInPlay(g, chaplainOfAlmsRow(), me)
	if c, _ := g.LookupCardForEffect(chaplain); !game.HasKeyword(&c, "first strike") {
		t.Error("Chaplain of Alms has no first strike")
	}
	castAtWardedCreature(t, g, opp, chaplain)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward {1} payment for a spell targeting Chaplain of Alms")
	}
}

func TestChapelShieldgeistGivesEachCreatureYouControlWard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := auraBear(g, me.ID)
	disturbIntoPlay(t, g, me, chaplainOfAlmsRow(), "Chapel Shieldgeist")
	blade := castAtWardedCreature(t, g, opp, bear)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward {1} payment for a spell targeting a creature under Chapel Shieldgeist")
	}
	answerPayUnless(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("a declined ward did not counter the spell")
	}
	if !opp.Graveyard.Contains(blade) {
		t.Error("the countered spell is not in its owner's graveyard")
	}
}
