package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b590_cards_test.go — Saviors of Kamigawa, batch 590 (#5601). Each
// test plays the real catalog entry and asserts what its text says.

const (
	b590FeralLightningOracle = "acd17fbf-645b-4c4d-b2c6-44127e3b743d"
	b590BordersOracle        = "b377f717-d0c2-4db5-aba2-fa4bfa35de0f"
	b590ThoughtsOracle       = "bb9e1d12-ebcc-4a8a-baef-9f69638a51cd"
	b590ShockwaveOracle      = "e9db79fd-a5b6-4d59-9176-e713e2e9c708"
	b590RallyOracle          = "fd9aeda1-9e44-4bc6-aa0d-19d58918d1f4"
	b590TormentOracle        = "0f07d3d0-1f79-4d50-ab3e-f9dc5ef59482"
	b590FlamesOracle         = "5db128ed-69c0-41f3-b8b4-fb7cefdb236c"
	b590SwarmOracle          = "c62a9e10-cfc6-48cc-a3f7-68d624182a5a"
	b590DominionOracle       = "ede2a435-bf9a-4f87-b579-259cfdebc13f"
)

func TestOldB590EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		b590FeralLightningOracle: "Feral Lightning",
		b590BordersOracle:        "Shifting Borders",
		b590ThoughtsOracle:       "Thoughts of Ruin",
		b590ShockwaveOracle:      "Aether Shockwave",
		b590RallyOracle:          "Rally the Horde",
		b590TormentOracle:        "Neverending Torment",
		b590FlamesOracle:         "Undying Flames",
		b590SwarmOracle:          "Endless Swarm",
		b590DominionOracle:       "Eternal Dominion",
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
	if spec, _ := effects.Lookup(b590BordersOracle); spec.Splice == nil {
		t.Error("Shifting Borders has no splice")
	}
}

// b590Library replaces a player's library with the cards, top first.
func b590Library(p *game.Player, cards ...game.Card) []uuid.UUID {
	p.Library.Cards = nil
	ids := make([]uuid.UUID, len(cards))
	// PushLibraryCardForTest pushes to the bottom, so the first card ends on top.
	for i := range cards {
		cards[i].Owner, cards[i].Controller = p.ID, p.ID
		ids[i] = effectstest.PushLibraryCardForTest(p, cards[i])
	}
	return ids
}

func b590Land(name string) game.Card {
	return game.Card{Name: name, TypeLine: "Basic Land — Forest"}
}

func b590Spell(name, cost string) game.Card {
	return game.Card{Name: name, TypeLine: "Sorcery", ManaCost: cost}
}

func b590CountNamed(g *game.Game, owner uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == name {
			n++
		}
	}
	return n
}

func b590Opponent(g *game.Game) (me, opp *game.Player) {
	me = g.Seats[g.Turn.ActiveSeat]
	opp = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	return me, opp
}

func TestOldB590FeralLightningMakesThreeHastyElementalsAndExilesThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := b590Opponent(g)
	effectstest.CastCatalogSpell(t, g, "Feral Lightning", "Sorcery", b590FeralLightningOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var ids []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Elemental" {
			ids = append(ids, c.InstanceID)
		}
	}
	if len(ids) != 3 {
		t.Fatalf("%d Elementals, want 3", len(ids))
	}
	for _, id := range ids {
		if p, tough := effectstest.EffectivePTOf(t, g, id); p != 3 || tough != 1 {
			t.Errorf("an Elemental is %d/%d, want 3/1", p, tough)
		}
		if !effectstest.HasKeywordOnBattlefield(t, g, id, "haste") {
			t.Error("an Elemental lacks haste")
		}
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := b590CountNamed(g, me.ID, "Elemental"); n != 0 {
		t.Errorf("%d Elementals survived the beginning of the end step", n)
	}
}

func TestOldB590ShiftingBordersExchangesTwoLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b590Opponent(g)
	mine := effectstest.AangPushLand(g, me.ID, "Island", false)
	theirs := effectstest.AangPushLand(g, opp.ID, "Mountain", false)
	other := effectstest.AangPushLand(g, opp.ID, "Swamp", false)
	// Two lands under one player's control: nothing to exchange.
	effectstest.CastCatalogSpell(t, g, "Shifting Borders", "Instant — Arcane", b590BordersOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: theirs}, {Kind: game.TargetCard, ID: other}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ControllerOf(t, g, theirs) != opp.ID || effectstest.ControllerOf(t, g, other) != opp.ID {
		t.Fatal("two lands of one player were exchanged")
	}
	effectstest.CastCatalogSpell(t, g, "Shifting Borders", "Instant — Arcane", b590BordersOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ControllerOf(t, g, mine) != opp.ID || effectstest.ControllerOf(t, g, theirs) != me.ID {
		t.Error("the two lands did not change hands")
	}
	// A single target is not a legal cast.
	if err := g.CastSpell(me.ID, effectstest.PushHandCardWithManaCost(me, "Borders", "Instant — Arcane", "{3}{U}"),
		game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: other}}}); err == nil {
		t.Error("cast with one target: accepted")
	}
}

func TestOldB590ShiftingBordersCanBeSplicedOntoAnArcaneSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b590Opponent(g)
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.AangPushLand(g, me.ID, "Island", false)
	theirs := effectstest.AangPushLand(g, opp.ID, "Mountain", false)
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	vortex := effectstest.HandCardFull(me, "Consuming Vortex", "Instant — Arcane", "{3}{U}", effectstest.ConsumingVortexOracle, []string{"U"})
	borders := effectstest.HandCardFull(me, "Shifting Borders", "Instant — Arcane", "{3}{U}", b590BordersOracle, []string{"U"})
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{U}{C}{C}{C}{U}"); err != nil {
		t.Fatal(err)
	}
	spliced := game.TargetRef{Kind: game.TargetCard, ID: mine, Splice: 1}
	spliced2 := game.TargetRef{Kind: game.TargetCard, ID: theirs, Splice: 1}
	if err := g.CastSpell(me.ID, vortex, game.CastSpellParams{Strict: true, Splices: []uuid.UUID{borders},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}, spliced, spliced2}}); err != nil {
		t.Fatalf("splice Shifting Borders onto an Arcane spell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(bear) {
		t.Error("the host spell did not return the Bear")
	}
	if effectstest.ControllerOf(t, g, mine) != opp.ID || effectstest.ControllerOf(t, g, theirs) != me.ID {
		t.Error("the spliced text did not exchange the lands")
	}
	if !me.Hand.Contains(borders) {
		t.Error("Shifting Borders did not stay in hand")
	}
}

func TestOldB590ThoughtsOfRuinTakesALandPerCardInTheCastersHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b590Opponent(g)
	me.Hand.Cards = nil
	effectstest.FillHand(me, "A", "B")
	var myLands, theirLands []uuid.UUID
	for i := 0; i < 3; i++ {
		myLands = append(myLands, effectstest.AangPushLand(g, me.ID, "Forest", false))
		theirLands = append(theirLands, effectstest.AangPushLand(g, opp.ID, "Island", false))
	}
	effectstest.CastCatalogSpell(t, g, "Thoughts of Ruin", "Sorcery", b590ThoughtsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// Answer every land-choice prompt with the first candidates.
	for i := 0; i < 8; i++ {
		answered := false
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceChooseCards {
				n := c.ChooseMin
				if err := g.ResolveChooseCards(c.ID, c.Chooser, c.ChooseCards[:n]); err != nil {
					t.Fatalf("answer: %v", err)
				}
				answered = true
				break
			}
		}
		if !answered {
			break
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	left := func(ids []uuid.UUID) int {
		n := 0
		for _, id := range ids {
			if g.Battlefield.Contains(id) {
				n++
			}
		}
		return n
	}
	if left(myLands) != 1 || left(theirLands) != 1 {
		t.Errorf("lands left: yours %d, theirs %d, want 1 each (two cards in hand)", left(myLands), left(theirLands))
	}
}

func TestOldB590AetherShockwaveTapsSpiritsOrEverythingElse(t *testing.T) {
	for mode, wantSpiritTapped := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := b590Opponent(g)
		spirit := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Spirit", TypeLine: "Creature — Spirit", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
		bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
		land := effectstest.AangPushLand(g, me.ID, "Island", false)
		effectstest.CastCatalogSpellWithModes(t, g, "Aether Shockwave", "Instant", b590ShockwaveOracle, []int{mode})
		effectstest.PassPriorityAroundTable(t, g)
		tapped := func(id uuid.UUID) bool {
			c := effectstest.FindBattlefieldCardByID(g, id)
			return c != nil && c.Tapped
		}
		if tapped(spirit) != wantSpiritTapped || tapped(bear) == wantSpiritTapped {
			t.Errorf("mode %d: spirit tapped %v, bear tapped %v", mode, tapped(spirit), tapped(bear))
		}
		if tapped(land) {
			t.Errorf("mode %d tapped a land", mode)
		}
	}
}

func TestOldB590RallyTheHordeExilesInThreesUntilALandEndsAGroup(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := b590Opponent(g)
	// Group one: spell, land, spell (last is not a land) -> repeat.
	// Group two: spell, spell, land (last is a land) -> stop.
	// The rest stays in the library.
	ids := b590Library(me,
		b590Spell("S1", "{1}"), b590Land("L1"), b590Spell("S2", "{1}"),
		b590Spell("S3", "{1}"), b590Spell("S4", "{1}"), b590Land("L2"),
		b590Spell("Untouched", "{1}"))
	effectstest.CastCatalogSpell(t, g, "Rally the Horde", "Sorcery", b590RallyOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i, id := range ids[:6] {
		if !exiledHas(g, me, id) {
			t.Errorf("card %d was not exiled", i)
		}
	}
	if exiledHas(g, me, ids[6]) || !me.Library.Contains(ids[6]) {
		t.Error("a seventh card was exiled after the group ended on a land")
	}
	if n := b590CountNamed(g, me.ID, "Warrior"); n != 4 {
		t.Errorf("%d Warriors, want 4 (one per nonland card exiled)", n)
	}
}

func exiledHas(g *game.Game, p *game.Player, id uuid.UUID) bool {
	z := g.FindCardZoneForEffect(id)
	return z != nil && z.Kind == game.ZoneExile
}

func TestOldB590RallyTheHordeStopsAtOnceWhenTheThirdCardIsALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := b590Opponent(g)
	ids := b590Library(me, b590Spell("S1", "{1}"), b590Spell("S2", "{1}"), b590Land("L1"), b590Spell("Untouched", "{1}"))
	effectstest.CastCatalogSpell(t, g, "Rally the Horde", "Sorcery", b590RallyOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if exiledHas(g, me, ids[3]) {
		t.Error("it repeated although the last card exiled was a land")
	}
	if n := b590CountNamed(g, me.ID, "Warrior"); n != 2 {
		t.Errorf("%d Warriors, want 2", n)
	}
}

func TestOldB590NeverendingTormentTakesAsManyCardsAsYourHandAndEpicCopiesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b590Opponent(g)
	seat := g.Turn.ActiveSeat
	me.Hand.Cards = nil
	effectstest.FillHand(me, "A", "B")
	ids := effectstest.SeedSearchLibrary(opp,
		game.Card{Name: "O1", TypeLine: "Instant"}, game.Card{Name: "O2", TypeLine: "Instant"},
		game.Card{Name: "O3", TypeLine: "Instant"}, game.Card{Name: "O4", TypeLine: "Instant"})
	effectstest.CastCatalogSpell(t, g, "Neverending Torment", "Sorcery", b590TormentOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SearchChoiceFor(g, me.ID); c == nil {
		t.Fatal("no search prompt")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, ids[0], ids[2])
	effectstest.PassPriorityAroundTable(t, g)
	if !exiledHas(g, opp, ids[0]) || !exiledHas(g, opp, ids[2]) {
		t.Error("the chosen cards were not exiled")
	}
	if opp.Library.Size() != 2 {
		t.Errorf("the library holds %d cards, want 2", opp.Library.Size())
	}
	// Epic: no more spells for the caster, and a copy at the next upkeep.
	probe := game.Card{Name: "Probe", TypeLine: "Sorcery"}
	if err := effectstest.CastGateFor(g, me.ID, probe, game.ZoneHand); err == nil {
		t.Error("the caster can still cast spells after epic")
	}
	b590UpkeepCopy(t, g, seat, &game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	effectstest.PassPriorityAroundTable(t, g)
	// The copy counts the hand again (two cards) and takes what the library
	// holds: the opponent drew one of the two remaining cards in their own
	// draw step on the way, so one is left, and the forced take asks nothing.
	if opp.Library.Size() != 0 {
		t.Errorf("the library holds %d cards, want 0", opp.Library.Size())
	}
	if n := b590ExiledCount(g, opp, ids[1], ids[3]); n != 1 {
		t.Errorf("%d of the two remaining cards were exiled by the copy, want 1 (the other was drawn)", n)
	}
}

func b590ExiledCount(g *game.Game, p *game.Player, ids ...uuid.UUID) int {
	n := 0
	for _, id := range ids {
		if exiledHas(g, p, id) {
			n++
		}
	}
	return n
}

// b590UpkeepCopy takes the table to the seat's next upkeep and resolves the
// epic trigger; when the copy offers a new target it is answered with `pick`.
func b590UpkeepCopy(t *testing.T, g *game.Game, seat int, pick *game.TargetRef) {
	t.Helper()
	effectstest.AdvanceToUpkeepOfSeat(t, g, seat)
	me := g.Seats[seat]
	for i := 0; i < 16; i++ {
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			if pick == nil {
				t.Fatal("an unexpected target prompt")
			}
			if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{*pick}); err != nil {
				t.Fatalf("pick the copy's target: %v", err)
			}
			return
		}
		if pick == nil && effectstest.StackFullyEmpty(g) && i > 0 {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("pass: %v", err)
		}
	}
	if pick != nil {
		t.Fatal("the epic copy offered no new target")
	}
}

func TestOldB590UndyingFlamesDealsDamageEqualToTheFirstNonlandCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b590Opponent(g)
	ids := b590Library(me, b590Land("L1"), b590Land("L2"), b590Spell("Big", "{3}{G}{G}"), b590Spell("After", "{1}"))
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Undying Flames", "Sorcery", b590FlamesOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range ids[:3] {
		if !exiledHas(g, me, id) {
			t.Error("a card of the run is not in exile")
		}
	}
	if exiledHas(g, me, ids[3]) {
		t.Error("it exiled past the first nonland card")
	}
	if opp.Life != life-5 {
		t.Errorf("the opponent is at %d, want %d (mana value 5)", opp.Life, life-5)
	}
	probe := game.Card{Name: "Probe", TypeLine: "Sorcery"}
	if err := effectstest.CastGateFor(g, me.ID, probe, game.ZoneHand); err == nil {
		t.Error("the caster can still cast spells after epic")
	}
}

func TestOldB590UndyingFlamesWithNoNonlandCardDealsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b590Opponent(g)
	b590Library(me, b590Land("L1"), b590Land("L2"))
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Undying Flames", "Sorcery", b590FlamesOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life || me.Library.Size() != 0 {
		t.Errorf("life %d -> %d, library %d: want no damage and an empty library", life, opp.Life, me.Library.Size())
	}
}

func TestOldB590EndlessSwarmMakesASnakePerCardAndEpicRepeatsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := b590Opponent(g)
	seat := g.Turn.ActiveSeat
	me.Hand.Cards = nil
	effectstest.FillHand(me, "A", "B", "C")
	effectstest.CastCatalogSpell(t, g, "Endless Swarm", "Sorcery", b590SwarmOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := b590CountNamed(g, me.ID, "Snake"); n != 3 {
		t.Fatalf("%d Snakes, want 3", n)
	}
	me.Hand.Cards = nil
	effectstest.FillHand(me, "D", "E")
	b590UpkeepCopy(t, g, seat, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := b590CountNamed(g, me.ID, "Snake"); n < 5 {
		t.Errorf("%d Snakes after the upkeep copy, want at least 5", n)
	}
}

func TestOldB590EternalDominionPutsTheFoundCardOntoTheBattlefieldUnderYourControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b590Opponent(g)
	ids := effectstest.SeedSearchLibrary(opp,
		game.Card{Name: "A Sorcery", TypeLine: "Sorcery"},
		game.Card{Name: "A Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
		game.Card{Name: "An Instant", TypeLine: "Instant"})
	effectstest.CastCatalogSpell(t, g, "Eternal Dominion", "Sorcery", b590DominionOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "A Sorcery") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "An Instant") != uuid.Nil {
		t.Error("the search offers an instant or a sorcery")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ids[1]) || effectstest.ControllerOf(t, g, ids[1]) != me.ID {
		t.Error("the Bear is not on the battlefield under your control")
	}
	probe := game.Card{Name: "Probe", TypeLine: "Sorcery"}
	if err := effectstest.CastGateFor(g, me.ID, probe, game.ZoneHand); err == nil {
		t.Error("the caster can still cast spells after epic")
	}
}
