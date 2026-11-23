package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsch "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/ch"
	cardsco "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/co"
	cardsd "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/d"
	cardsgr "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/gr"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// disturb_batch_b_test.go — the second batch of disturb cards (#1855,
// ADR 0107 §4): each card's disturb cast lands its back face, and each
// non-keyword ability on either face does what the card prints. The
// cast and the exile clause themselves are pinned in disturb_test.go.

func catgeistRow() cards.Card {
	return effectstest.DisturbRow(effects.MischievousCatgeistOracleID, []string{"U"},
		effectstest.DisturbFace{Name: "Mischievous Catgeist", TypeLine: "Creature — Cat Spirit", Cost: "{1}{U}", Power: "1", Toughness: "1"},
		effectstest.DisturbFace{Name: "Catlike Curiosity", TypeLine: "Enchantment — Aura", Cost: "", Power: "", Toughness: ""})
}

func distractingGeistRow() cards.Card {
	return effectstest.DisturbRow(effects.DistractingGeistOracleID, []string{"W"},
		effectstest.DisturbFace{Name: "Distracting Geist", TypeLine: "Creature — Spirit", Cost: "{2}{W}", Power: "2", Toughness: "1"},
		effectstest.DisturbFace{Name: "Clever Distraction", TypeLine: "Enchantment — Aura", Cost: "", Power: "", Toughness: ""})
}

func gutterSkulkerRow() cards.Card {
	return effectstest.DisturbRow(cardsgr.GutterSkulkerOracleID, []string{"U"},
		effectstest.DisturbFace{Name: "Gutter Skulker", TypeLine: "Creature — Spirit", Cost: "{3}{U}", Power: "3", Toughness: "3"},
		effectstest.DisturbFace{Name: "Gutter Shortcut", TypeLine: "Enchantment — Aura", Cost: "", Power: "", Toughness: ""})
}

func covetousCastawayRow() cards.Card {
	return effectstest.DisturbRow(cardsco.CovetousCastawayOracleID, []string{"U"},
		effectstest.DisturbFace{Name: "Covetous Castaway", TypeLine: "Creature — Human", Cost: "{1}{U}", Power: "1", Toughness: "3"},
		effectstest.DisturbFace{Name: "Ghostly Castigator", TypeLine: "Creature — Spirit", Cost: "", Power: "3", Toughness: "4"})
}

func devotedGrafkeeperRow() cards.Card {
	return effectstest.DisturbRow(cardsd.DevotedGrafkeeperOracleID, []string{"W", "U"},
		effectstest.DisturbFace{Name: "Devoted Grafkeeper", TypeLine: "Creature — Human Peasant", Cost: "{W}{U}", Power: "2", Toughness: "1"},
		effectstest.DisturbFace{Name: "Departed Soulkeeper", TypeLine: "Creature — Spirit", Cost: "", Power: "3", Toughness: "1"})
}

func chaplainOfAlmsRow() cards.Card {
	return effectstest.DisturbRow(cardsch.ChaplainOfAlmsOracleID, []string{"W"},
		effectstest.DisturbFace{Name: "Chaplain of Alms", TypeLine: "Creature — Human Cleric", Cost: "{W}", Power: "1", Toughness: "1"},
		effectstest.DisturbFace{Name: "Chapel Shieldgeist", TypeLine: "Creature — Spirit Cleric", Cost: "", Power: "2", Toughness: "1"})
}

// frontFaceInPlay runs a row down the import road onto the battlefield
// front face up, ready to attack and block.
func frontFaceInPlay(g *game.Game, row cards.Card, p *game.Player) uuid.UUID {
	c := deck.ToGameCard(row, false)
	c.Owner, c.Controller = p.ID, p.ID
	return effectstest.PushBattlefieldCardWithTimestamp(g, c)
}

// disturbOnto disturbs an Aura card onto `host` and resolves it,
// failing unless the back face is attached to the host.
func disturbOnto(t *testing.T, g *game.Game, p *game.Player, row cards.Card, host uuid.UUID) uuid.UUID {
	t.Helper()
	id := effectstest.ImportToGraveyard(t, g, row, p)
	disturb(t, g, p, id, game.TargetRef{Kind: game.TargetCard, ID: host})
	effectstest.PassPriorityAroundTable(t, g)
	zone, aura := effectstest.CardWhere(g, p, id)
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
	id := effectstest.ImportToGraveyard(t, g, row, p)
	disturb(t, g, p, id)
	effectstest.PassPriorityAroundTable(t, g)
	if zone, c := effectstest.CardWhere(g, p, id); zone != "battlefield" || c.Name != back {
		t.Fatalf("%s: disturbed into %s as %q, want %s on the battlefield", row.Name, zone, c.Name, back)
	}
	return id
}

// --- Mischievous Catgeist // Catlike Curiosity -------------------

func TestMischievousCatgeistDrawsOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cat := frontFaceInPlay(g, catgeistRow(), me)
	hand := me.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, cat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("Mischievous Catgeist connected and drew %d, want 1", got)
	}
}

func TestCatlikeCuriosityGivesTheEnchantedCreatureTheDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.AuraBear(g, me.ID)
	disturbOnto(t, g, me, catgeistRow(), bear)
	hand := me.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("the creature enchanted by Catlike Curiosity connected and drew %d, want 1", got)
	}
}

// --- Distracting Geist // Clever Distraction ---------------------

func TestDistractingGeistTapsADefendersCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	geist := frontFaceInPlay(g, distractingGeistRow(), me)
	victim := effectstest.BrCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	bystander := effectstest.BrCreature(g, g.Seats[2].ID, "Other Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, geist)
	effectstest.PickIfAsked(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedNow(g, victim) {
		t.Error("Distracting Geist's attack did not tap the defending player's creature")
	}
	if effectstest.TappedNow(g, bystander) {
		t.Error("a creature of a player who is not defending was tapped")
	}
}

func TestCleverDistractionGivesTheEnchantedCreatureTheTap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.AuraBear(g, me.ID)
	disturbOnto(t, g, me, distractingGeistRow(), bear)
	victim := effectstest.BrCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PickIfAsked(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedNow(g, victim) {
		t.Error("the creature enchanted by Clever Distraction attacked and tapped nothing")
	}
}

// --- Gutter Skulker // Gutter Shortcut ---------------------------

func TestGutterSkulkerIsUnblockableOnlyWhileAttackingAlone(t *testing.T) {
	// Alone: no block.
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	skulker := frontFaceInPlay(g, gutterSkulkerRow(), me)
	wall := effectstest.BrCreature(g, opp.ID, "Wall", "Creature — Wall", 0, 8)
	effectstest.BrAttack(t, g, skulker)
	effectstest.BrRefusal(t, g.DeclareBlocker(wall, skulker), game.BlockReasonCantBeBlocked)

	// With a second attacker it is not attacking alone.
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	skulker = frontFaceInPlay(g, gutterSkulkerRow(), me)
	friend := effectstest.BrCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	wall = effectstest.BrCreature(g, opp.ID, "Wall", "Creature — Wall", 0, 8)
	effectstest.BrAttack(t, g, skulker, friend)
	if err := g.DeclareBlocker(wall, skulker); err != nil {
		t.Errorf("Gutter Skulker attacking beside another creature could not be blocked: %v", err)
	}
}

func TestGutterShortcutMakesTheEnchantedCreatureUnblockableAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.AuraBear(g, me.ID)
	disturbOnto(t, g, me, gutterSkulkerRow(), bear)
	wall := effectstest.BrCreature(g, opp.ID, "Wall", "Creature — Wall", 0, 8)
	effectstest.BrAttack(t, g, bear)
	effectstest.BrRefusal(t, g.DeclareBlocker(wall, bear), game.BlockReasonCantBeBlocked)
}

// --- Covetous Castaway // Ghostly Castigator ---------------------

func TestCovetousCastawayMillsThreeWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	castaway := frontFaceInPlay(g, covetousCastawayRow(), me)
	library := me.Library.Size()
	effectstest.Destroy(t, g, castaway)
	effectstest.PassPriorityAroundTable(t, g)
	if got := library - me.Library.Size(); got != 3 {
		t.Errorf("Covetous Castaway's death milled %d cards, want 3", got)
	}
	if zone, _ := effectstest.CardWhere(g, me, castaway); zone != "graveyard" {
		t.Errorf("the front face died into %s, want the graveyard", zone)
	}
}

func TestGhostlyCastigatorShufflesChosenGraveyardCardsAway(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.SeedGraveyardCard(t, g, "Card A", "Sorcery", "")
	b := effectstest.SeedGraveyardCard(t, g, "Card B", "Sorcery", "")
	kept := effectstest.SeedGraveyardCard(t, g, "Card C", "Sorcery", "")
	id := effectstest.ImportToGraveyard(t, g, covetousCastawayRow(), me)
	library := me.Library.Size()

	disturb(t, g, me, id)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickIfAsked(t, g, me.ID, a, b)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	if zone, c := effectstest.CardWhere(g, me, id); zone != "battlefield" || c.Name != "Ghostly Castigator" {
		t.Fatalf("the disturb cast left the card in %s as %q", zone, c.Name)
	}
	for _, card := range []uuid.UUID{a, b} {
		if zone, _ := effectstest.CardWhere(g, me, card); zone != "library" {
			t.Errorf("a chosen graveyard card is in %s, want the library", zone)
		}
	}
	if zone, _ := effectstest.CardWhere(g, me, kept); zone != "graveyard" {
		t.Errorf("an unchosen graveyard card is in %s, want the graveyard", zone)
	}
	if got := me.Library.Size() - library; got != 2 {
		t.Errorf("the library grew by %d, want 2", got)
	}
}

// --- Devoted Grafkeeper // Departed Soulkeeper -------------------

func TestDevotedGrafkeeperMillsTwoAsItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.ImportToHand(devotedGrafkeeperRow(), me)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	library := me.Library.Size()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Devoted Grafkeeper: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := library - me.Library.Size(); got != 2 {
		t.Errorf("Devoted Grafkeeper's entry milled %d, want 2", got)
	}
}

func TestDevotedGrafkeeperTapsWhenYouCastFromYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	frontFaceInPlay(g, devotedGrafkeeperRow(), me)
	victim := effectstest.BrCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.BrCreature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	// From hand: no trigger.
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", "4457ed35-7c10-48c8-9776-456485fdf070",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Fatal("a spell cast from hand triggered Devoted Grafkeeper")
	}
	effectstest.PassPriorityAroundTable(t, g)

	// From the graveyard: tap target creature you don't control.
	looting := effectstest.SeedGraveyardCard(t, g, "Faithless Looting", "Sorcery", "3d6fa57a-aa53-4b5c-b8af-a7612c823117")
	if err := g.CastSpell(me.ID, looting, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("a flashback cast did not trigger Devoted Grafkeeper")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
		t.Error("Devoted Grafkeeper's trigger accepted a creature its controller controls")
	}
	effectstest.PickIfAsked(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedNow(g, victim) {
		t.Error("Devoted Grafkeeper's trigger did not tap the opponent's creature")
	}
}

func TestDepartedSoulkeeperBlocksOnlyFlyers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	soulkeeper := disturbIntoPlay(t, g, me, devotedGrafkeeperRow(), "Departed Soulkeeper")
	walker := effectstest.BrCreature(g, opp.ID, "Walker", "Creature — Bear", 2, 2)
	flyer := effectstest.BrCreature(g, opp.ID, "Flyer", "Creature — Bird", 2, 2, "flying")

	effectstest.AdvanceToDeclareAttackersOf(t, g, 1)
	for _, a := range []uuid.UUID{walker, flyer} {
		if err := g.DeclareAttacker(a, me.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.BrRefusal(t, g.DeclareBlocker(soulkeeper, walker), game.BlockReasonCantBlockAttacker)
	if err := g.DeclareBlocker(soulkeeper, flyer); err != nil {
		t.Errorf("Departed Soulkeeper could not block a flyer: %v", err)
	}
}

// --- Chaplain of Alms // Chapel Shieldgeist -----------------------

func TestChaplainOfAlmsHasWard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	chaplain := frontFaceInPlay(g, chaplainOfAlmsRow(), me)
	if c, _ := g.LookupCardForEffect(chaplain); !game.HasKeyword(&c, "first strike") {
		t.Error("Chaplain of Alms has no first strike")
	}
	effectstest.CastAtWardedCreature(t, g, opp, chaplain)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward {1} payment for a spell targeting Chaplain of Alms")
	}
}

func TestChapelShieldgeistGivesEachCreatureYouControlWard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.AuraBear(g, me.ID)
	disturbIntoPlay(t, g, me, chaplainOfAlmsRow(), "Chapel Shieldgeist")
	blade := effectstest.CastAtWardedCreature(t, g, opp, bear)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward {1} payment for a spell targeting a creature under Chapel Shieldgeist")
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("a declined ward did not counter the spell")
	}
	if !opp.Graveyard.Contains(blade) {
		t.Error("the countered spell is not in its owner's graveyard")
	}
}
