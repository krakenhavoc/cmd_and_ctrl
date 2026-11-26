package testss82m

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Commander 2016, batch 257 (#5537).

const (
	oldB257AncientExcavation = "46db3dc2-9eb1-4be3-96f9-0013119fcd97"
	oldB257HorizonChimera    = "a1ed77f9-51ac-4cca-a9ce-d3874f3948ad"
	oldB257WhimsOfTheFates   = "7d359434-eb1b-478e-a745-10e2c564d89f"
	oldB257ManifoldInsights  = "9c793037-6ba8-400a-b24d-6032141e8dae"
	oldB257FrenziedFugue     = "b3f9bd29-8dd6-4915-8ffd-02e02f26fed8"
	oldB257CurseOfVengeance  = "b9b1c644-041a-4b61-9d6d-bc73a3a95bb8"
	oldB257ChiefEngineer     = "02dbb6f8-ef84-4156-87dc-34033e64fdb7"
	oldB257Sydri             = "4e92d36b-1a35-4fa9-87ea-10eace5a3cc7"
	oldB257QuirionExplorer   = "766c8a68-9c51-44ea-903c-1f5cb30941d1"
	oldB257KorozdaGuildmage  = "b209b4ae-ceba-4a98-b307-222d3c78f8a1"
	oldB257TrashForTreasure  = "e69279f6-50a1-4db5-83be-30a2ea7fc6a9"
)

func TestOldB257AllCardsAreCatalogued(t *testing.T) {
	for _, id := range []string{oldB257AncientExcavation, oldB257HorizonChimera, oldB257WhimsOfTheFates,
		oldB257ManifoldInsights, oldB257FrenziedFugue, oldB257CurseOfVengeance, oldB257ChiefEngineer,
		oldB257Sydri, oldB257QuirionExplorer, oldB257KorozdaGuildmage, oldB257TrashForTreasure} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: catalogued %v, completeness %v, want Full", id, ok, spec.Completeness)
		}
	}
}

// Draw cards equal to the hand, then discard that many; the drawn cards are
// legal discards.
func TestOldB257AncientExcavationDrawsThenDiscardsThatMany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	for i := 0; i < 3; i++ {
		effectstest.PushHandCardOfType(me, "Filler", "Sorcery")
	}
	before := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Ancient Excavation", "Instant", oldB257AncientExcavation, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// The hand as it resolved (the spell is on the stack, not in it) is drawn
	// again, then that many are discarded.
	if got := me.Hand.Size(); got != 2*before {
		t.Fatalf("hand before the discard = %d, want %d", got, 2*before)
	}
	p := effectstest.DiscardChoiceFor(g, me.ID)
	if p == nil || p.Count != before {
		t.Fatalf("discard prompt = %+v, want %d cards", p, before)
	}
	var ids []uuid.UUID
	for _, c := range me.Hand.Cards[:before] {
		ids = append(ids, c.InstanceID)
	}
	effectstest.AnswerDiscard(t, g, me.ID, ids...)
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand after the discard = %d, want %d", got, before)
	}
}

func TestOldB257AncientExcavationBasicLandcycling(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushCatalogHandCard(me, "Ancient Excavation", "Instant", oldB257AncientExcavation)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
	)
	if err := effectstest.HobB4Activate(t, g, me.ID, id, 0, "{C}{C}", nil); err != nil {
		t.Fatalf("landcycling: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) || !me.Graveyard.Contains(id) {
		t.Errorf("Forest in hand %v, Excavation in graveyard %v", me.Hand.Contains(ids[1]), me.Graveyard.Contains(id))
	}
}

// One life per card drawn, and none for a card an opponent draws.
func TestOldB257HorizonChimeraGainsOneLifePerDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Horizon Chimera", "Creature — Chimera", oldB257HorizonChimera, false)
	life := me.Life
	if err := drawFor(g, me.ID, 2); err != nil {
		t.Fatalf("draw: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life = %d, want %d", me.Life, life+2)
	}
	if err := drawFor(g, them.ID, 1); err != nil {
		t.Fatalf("draw: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life after an opponent's draw = %d, want %d", me.Life, life+2)
	}
	spec, _ := effects.Lookup(oldB257HorizonChimera)
	if !slices.Contains(spec.PrintedKeywords, "flash") {
		t.Errorf("printed keywords %v lack flash", spec.PrintedKeywords)
	}
}

// Whims of the Fates: every player separates their permanents into three piles
// and sacrifices one at random; a player with one permanent loses it or keeps
// it as the pile falls, and the table ends with no prompt open.
func TestOldB257WhimsOfTheFatesSacrificesOnePileEach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	perms := make([][]uuid.UUID, len(g.Seats))
	for i, p := range g.Seats {
		for _, n := range []string{"Rock A", "Rock B", "Rock C"} {
			perms[i] = append(perms[i], effectstest.PushArtifactToBattlefieldForTest(g, p.ID, n, "Artifact", ""))
		}
	}
	effectstest.CastCatalogSpell(t, g, "Whims of the Fates", "Sorcery", oldB257WhimsOfTheFates, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// Starting with the caster, each player puts one rock in each pile.
	for i := range g.Seats {
		seat := g.Seats[(g.Turn.ActiveSeat+i)%len(g.Seats)]
		x := perms[(g.Turn.ActiveSeat+i)%len(g.Seats)]
		c := openRevealPick(g)
		if c == nil || c.Chooser != seat.ID {
			t.Fatalf("want %s asked for a pile, got %+v", seat.Name, c)
		}
		if err := g.ResolveRevealPick(c.ID, seat.ID, []uuid.UUID{x[0]}); err != nil {
			t.Fatal(err)
		}
		c = openRevealPick(g)
		if c == nil || c.Chooser != seat.ID {
			t.Fatalf("want %s asked for a second pile, got %+v", seat.Name, c)
		}
		if err := g.ResolveRevealPick(c.ID, seat.ID, []uuid.UUID{x[1]}); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		left := 0
		for _, id := range perms[i] {
			if g.Battlefield.Contains(id) {
				left++
			}
		}
		if left != 2 {
			t.Errorf("%s has %d rocks left, want 2 (one pile of one each)", p.Name, left)
		}
	}
}

func openRevealPick(g *game.Game) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		if c := g.PendingChoices[i]; c != nil && c.Kind == game.PendingChoiceRevealPick {
			return c
		}
	}
	return nil
}

// Manifold Insights: the next opponent in turn order chooses first, each
// opponent a different nonland card; the controller takes them, and the rest
// (the unchosen cards and lands) go to the bottom.
func TestOldB257ManifoldInsightsEachOpponentChoosesADifferentNonlandCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	var cards []game.Card
	for i := 0; i < 6; i++ {
		cards = append(cards, game.Card{Name: "Spell", TypeLine: "Sorcery"})
	}
	cards = append(cards, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	ids := effectstest.SeedSearchLibrary(me, cards...)
	handBefore := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Manifold Insights", "Sorcery", oldB257ManifoldInsights, nil)
	effectstest.PassPriorityAroundTable(t, g)
	opps := effects.OpponentsInTurnOrder(g, me.ID)
	if len(opps) != 3 {
		t.Fatalf("opponents = %d, want 3", len(opps))
	}
	var picked []uuid.UUID
	for i, o := range opps {
		c := effectstest.ChooseCardsChoiceFor(g, o)
		if c == nil {
			t.Fatalf("opponent %d was not asked", i)
		}
		for _, id := range c.ChooseCards {
			for _, p := range picked {
				if id == p {
					t.Errorf("opponent %d is offered a card an earlier opponent took", i)
				}
			}
			if cc, ok := g.LookupCardForEffect(id); ok && cc.IsLand() {
				t.Errorf("opponent %d is offered a land", i)
			}
		}
		pick := c.ChooseCards[0]
		effectstest.AnswerChooseCards(t, g, o, pick)
		picked = append(picked, pick)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != handBefore+3 {
		t.Errorf("hand = %d, want %d (three chosen cards)", got, handBefore+3)
	}
	for _, p := range picked {
		if !me.Hand.Contains(p) {
			t.Errorf("chosen card %v is not in hand", p)
		}
	}
	if got := me.Library.Size(); got != len(ids)-3 {
		t.Errorf("library = %d, want %d", got, len(ids)-3)
	}
}

// Quirion Explorer taps for a colour an opposing land could produce.
func TestOldB257QuirionExplorerReadsOpposingLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	elf := effectstest.SeedPermanentWithOracle(g, me.ID, "Quirion Explorer", "Creature — Elf Druid Scout", oldB257QuirionExplorer)
	effectstest.SeedManaLand(g, them.ID, "Island", "Basic Land — Island", "U")
	effectstest.SeedManaLand(g, me.ID, "Forest", "Basic Land — Forest", "G")
	if err := g.ActivateManaAbility(me.ID, elf, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if pick := effectstest.ManaPickFor(g, me.ID); pick != nil {
		if len(pick.ColorOptions) != 1 || pick.ColorOptions[0] != "U" {
			t.Errorf("options = %v, want only {U}", pick.ColorOptions)
		}
	} else if me.ManaPool.Total() != 1 {
		t.Errorf("pool = %v, want one mana", me.ManaPool)
	}
}

func TestOldB257QuirionExplorerWithNoOpposingLandsMakesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	elf := effectstest.SeedPermanentWithOracle(g, me.ID, "Quirion Explorer", "Creature — Elf Druid Scout", oldB257QuirionExplorer)
	if err := g.ActivateManaAbility(me.ID, elf, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.ManaPool.Total() != 0 || effectstest.ManaPickFor(g, me.ID) != nil {
		t.Errorf("pool = %v; want nothing produced", me.ManaPool)
	}
}

// Sydri's first ability animates a noncreature artifact as big as its mana
// value until end of turn; a creature is not a legal target.
func TestOldB257SydriAnimatesAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	sydri := effectstest.PushCatalogPermanent(g, me.ID, "Sydri, Galvanic Genius", "Legendary Creature — Human Artificer", oldB257Sydri, false)
	rock := effectstest.PushArtifactToBattlefieldForTest(g, me.ID, "Big Rock", "Artifact", "")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == rock {
				g.Battlefield.Cards[i].ManaCost = "{4}"
			}
		}
	})
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Artifact Creature — Bear", 2, 2)
	if err := effectstest.HobB4Activate(t, g, me.ID, sydri, 0, "{U}", effectstest.CardTarget(bear)); err == nil {
		t.Fatal("animated a creature")
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, sydri, 0, "{U}", effectstest.CardTarget(rock)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, rock); p != 4 || tough != 4 {
		t.Errorf("rock is %d/%d, want 4/4", p, tough)
	}
	effectstest.AdvanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if p, _ := effectstest.PtOf(t, g, rock); p == 4 {
		t.Error("the animation outlasted the turn")
	}
}

func TestOldB257SydriGrantsDeathtouchAndLifelinkToAnArtifactCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	sydri := effectstest.PushCatalogPermanent(g, me.ID, "Sydri, Galvanic Genius", "Legendary Creature — Human Artificer", oldB257Sydri, false)
	golem := effectstest.B12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.HobB4Activate(t, g, me.ID, sydri, 1, "{W}{B}", effectstest.CardTarget(bear)); err == nil {
		t.Fatal("targeted a nonartifact creature")
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, sydri, 1, "{W}{B}", effectstest.CardTarget(golem)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, kw := range []string{"deathtouch", "lifelink"} {
		if !effectstest.HasEffectiveKeyword(t, g, golem, kw) {
			t.Errorf("the golem has no %s", kw)
		}
	}
	effectstest.AdvanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if effectstest.HasEffectiveKeyword(t, g, golem, "deathtouch") {
		t.Error("deathtouch outlasted the turn")
	}
}

// Korozda Guildmage's pump grants intimidate.
func TestOldB257KorozdaGuildmagePump(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	gm := effectstest.PushCatalogPermanent(g, me.ID, "Korozda Guildmage", "Creature — Elf Shaman", oldB257KorozdaGuildmage, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.HobB4Activate(t, g, me.ID, gm, 0, "{C}{B}{G}", effectstest.CardTarget(bear)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("bear is %d/%d, want 3/3", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "intimidate") {
		t.Error("no intimidate")
	}
}

// The second ability sacrifices a nontoken creature and makes Saprolings equal
// to its toughness as it last was; a token cannot be the sacrifice.
func TestOldB257KorozdaGuildmageSacrificeMakesSaprolings(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	gm := effectstest.PushCatalogPermanent(g, me.ID, "Korozda Guildmage", "Creature — Elf Shaman", oldB257KorozdaGuildmage, false)
	wall := effectstest.B12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 4)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{B}{G}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, gm, 1, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{wall}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wall) {
		t.Error("the wall was not sacrificed")
	}
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Saproling" && c.Controller == me.ID {
			n++
		}
	}
	if n != 4 {
		t.Errorf("Saprolings = %d, want 4 (the wall's toughness)", n)
	}
}

// Trash for Treasure: sacrifice an artifact, return an artifact card.
func TestOldB257TrashForTreasureReturnsAnArtifactCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	fodder := effectstest.PushArtifactToBattlefieldForTest(g, me.ID, "Fodder", "Artifact", "")
	bear := effectstest.WthGraveyardCard(me, "Bear", "Creature — Bear")
	id := effectstest.PushCatalogHandCard(me, "Trash for Treasure", "Sorcery", oldB257TrashForTreasure)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardTarget(bear), SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err == nil {
		t.Fatal("targeted a nonartifact card")
	}
	rock := effectstest.WthGraveyardCard(me, "Old Rock", "Artifact")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardTarget(rock), Strict: true}); err == nil {
		t.Fatal("cast without sacrificing an artifact")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardTarget(rock), SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the artifact was not sacrificed")
	}
	if _, ok := effectstest.VisBattlefieldNamed(g, me.ID, "Old Rock"); !ok {
		t.Error("the artifact card did not return")
	}
}

// Chief Engineer gives artifact spells convoke, and not other spells.
func TestOldB257ChiefEngineerGrantsConvokeToArtifactSpells(t *testing.T) {
	spec, ok := effects.Lookup(oldB257ChiefEngineer)
	if !ok || len(spec.GrantedTapCosts) != 1 {
		t.Fatalf("granted tap costs = %d, want 1", len(spec.GrantedTapCosts))
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	grant := spec.GrantedTapCosts[0]
	art := game.Card{Name: "Rock", TypeLine: "Artifact"}
	creature := game.Card{Name: "Bear", TypeLine: "Creature — Bear"}
	if !grant.Spells(g, me.ID, art, game.ZoneHand) {
		t.Error("an artifact spell does not have convoke")
	}
	if grant.Spells(g, me.ID, creature, game.ZoneHand) {
		t.Error("a nonartifact spell has convoke")
	}
}

// Frenzied Fugue: control until end of turn, untap, haste; it goes back, and
// returns at your upkeep.
func TestOldB257FrenziedFugueStealsOnEnteringAndAtUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	var them *game.Player
	for _, s := range g.Seats {
		if s.ID != me.ID {
			them = s
			break
		}
	}
	bear := effectstest.B12Creature(g, them.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.CastCatalogSpell(t, g, "Frenzied Fugue", "Enchantment — Aura", oldB257FrenziedFugue, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, bear); got != me.ID {
		t.Fatalf("controller = %v, want the caster", got)
	}
	if effectstest.TappedOnBattlefield(t, g, bear) {
		t.Error("the bear is still tapped")
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("the bear has no haste")
	}
	// The turn passes to an opponent: control returns.
	effectstest.AdvanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if got := effectstest.ControllerOf(t, g, bear); got != them.ID {
		t.Fatalf("controller after the turn = %v, want the owner", got)
	}
	// Back at my upkeep it triggers again.
	seat := 0
	for i, s := range g.Seats {
		if s.ID == me.ID {
			seat = i
		}
	}
	effectstest.AdvanceToMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, bear); got != me.ID {
		t.Errorf("controller after my upkeep = %v, want me", got)
	}
}

// Curse of Vengeance: a spite counter per spell the enchanted player casts;
// when they lose, X life and X cards.
func TestOldB257CurseOfVengeanceCountsSpellsAndPaysOnTheirLoss(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	curse := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: curse, Name: "Curse of Vengeance", TypeLine: "Enchantment — Aura Curse", OracleID: oldB257CurseOfVengeance,
			Owner: me.ID, Controller: me.ID, AttachedTo: game.TargetRef{Kind: game.TargetPlayer, ID: them.ID},
		})
	})
	for i := 0; i < 2; i++ {
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventCast, Actor: them.ID, CardID: uuid.New()})
		})
		effectstest.PassPriorityAroundTable(t, g)
	}
	// A spell cast by someone else adds nothing.
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventCast, Actor: me.ID, CardID: uuid.New()}) })
	effectstest.PassPriorityAroundTable(t, g)
	var counters int
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == curse {
				counters = c.Counters["spite"]
			}
		}
	})
	if counters != 2 {
		t.Fatalf("spite counters = %d, want 2", counters)
	}
	life, hand := me.Life, me.Hand.Size()
	if err := g.Concede(them.ID); err != nil {
		t.Fatalf("concede: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life = %d, want %d", me.Life, life+2)
	}
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand = %d, want %d", got, hand+2)
	}
}

func TestOldB257WhimsIsASacrificeSweep(t *testing.T) {
	spec, _ := effects.Lookup(oldB257WhimsOfTheFates)
	if spec.Purpose.Sweep.How != game.SweepSacrifice {
		t.Errorf("sweep = %+v, want a sacrifice", spec.Purpose.Sweep)
	}
}

func drawFor(g *game.Game, who uuid.UUID, n int) error {
	var err error
	g.WithWriteLock(func() { err = g.DrawNForEffect(who, n) })
	return err
}
