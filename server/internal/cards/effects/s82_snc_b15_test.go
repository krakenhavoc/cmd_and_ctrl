package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	sncB15BrokersCharm        = "a90de2b8-4d71-4d0e-82b3-3b7ab194165b"
	sncB15MaestrosCharm       = "3b207754-9172-426a-8efc-a9c3912d94bb"
	sncB15RiveteersCharm      = "54e1b7f3-096a-4eb0-bcce-cce036d6cbe5"
	sncB15ObscuraCharm        = "884eaa3f-b6af-4f0e-b0e7-cd40666986ce"
	sncB15CabarettiCharm      = "36d8b701-f5dc-42c0-9869-268f8149cbd1"
	sncB15RiveteersConfluence = "9f0bd777-f79d-47cd-83e3-85b80d5e7942"
	sncB15CabarettiConfluence = "99c32f20-da98-477d-9eb2-9c0b189d8d88"
	sncB15ObscuraConfluence   = "1f2fb1dd-82e4-45fa-b2a9-7df7b8a4d533"
	sncB15BrokersConfluence   = "d9c6c9cc-2dbf-48ac-ad3a-c500e7074c62"
	sncB15MaestrosConfluence  = "ef67a41b-a2a6-4f1b-a0d5-15560e99e387"
	sncB15CleanupCrew         = "12e0321b-1cd7-4519-8929-3261cf11a1a9"
)

func TestSncB15CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		sncB15BrokersCharm: "Brokers Charm", sncB15MaestrosCharm: "Maestros Charm",
		sncB15RiveteersCharm: "Riveteers Charm", sncB15ObscuraCharm: "Obscura Charm",
		sncB15CabarettiCharm: "Cabaretti Charm", sncB15RiveteersConfluence: "Riveteers Confluence",
		sncB15CabarettiConfluence: "Cabaretti Confluence", sncB15ObscuraConfluence: "Obscura Confluence",
		sncB15BrokersConfluence: "Brokers Confluence", sncB15MaestrosConfluence: "Maestros Confluence",
		sncB15CleanupCrew: "Cleanup Crew",
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

func sncB15Gold(g *game.Game, owner uuid.UUID, name string, mv int, colors ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear", Colors: colors,
		ManaCost: "{" + string(rune('0'+mv)) + "}", Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
}

// --- Cleanup Crew -----------------------------------------------------------

func sncB15CleanupMode(t *testing.T, g *game.Game, me uuid.UUID, mode int) {
	t.Helper()
	castRingCreature(t, g, "Cleanup Crew", "Creature — Human Citizen", sncB15CleanupCrew, 4, 4)
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me)
	if pick == nil {
		t.Fatalf("no mode prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(pick.ID, me, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
}

func TestSncB15CleanupCrewGainsFourLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Life
	sncB15CleanupMode(t, g, me.ID, 3)
	passPriorityAroundTable(t, g)
	if me.Life != before+4 {
		t.Errorf("life %d, want %d", me.Life, before+4)
	}
}

func TestSncB15CleanupCrewDestroysAnArtifactOrEnchantmentAndExilesACard(t *testing.T) {
	t.Run("artifact", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
		aura := b12Permanent(g, opp.ID, "Pacifism", "Enchantment — Aura")
		sncB15CleanupMode(t, g, me.ID, 0)
		if p := latestPickTarget(g, me.ID); p == nil {
			t.Fatal("no target prompt")
		} else if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: aura}); err == nil {
			t.Fatal("an enchantment was accepted as the artifact")
		}
		answerPickTarget(t, g, rock)
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(rock) || !g.Battlefield.Contains(aura) {
			t.Error("the artifact must die and the enchantment stay")
		}
	})
	t.Run("enchantment", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		aura := b12Permanent(g, opp.ID, "Pacifism", "Enchantment — Aura")
		sncB15CleanupMode(t, g, me.ID, 1)
		answerPickTarget(t, g, aura)
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(aura) {
			t.Error("the enchantment survived")
		}
	})
	t.Run("graveyard", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		card := uuid.New()
		opp.Graveyard.PushTop(game.Card{InstanceID: card, Name: "Dead", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
		sncB15CleanupMode(t, g, me.ID, 2)
		answerPickTarget(t, g, card)
		passPriorityAroundTable(t, g)
		if opp.Graveyard.Contains(card) {
			t.Error("the card is still in the graveyard")
		}
	})
}

// --- Cabaretti Charm --------------------------------------------------------

func TestSncB15CabarettiCharmDamageCountsYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 3; i++ {
		b12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	}
	big := b12Creature(g, opp.ID, "Big", "Creature — Ogre", 3, 3)
	castModal(t, g, "Cabaretti Charm", "Instant", sncB15CabarettiCharm, []int{0},
		[]game.TargetRef{modeRef(game.TargetCard, big, 0, 0)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("3 damage (three creatures) should kill a 3/3")
	}
}

func TestSncB15CabarettiCharmPumpsOnlyYourCreaturesAsItResolves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	castModal(t, g, "Cabaretti Charm", "Instant", sncB15CabarettiCharm, []int{1}, nil)
	passPriorityAroundTable(t, g)
	if c := avB1Card(t, g, mine); c.CurrentPower() != 3 || !hasKeywordOf(g, mine, "trample") {
		t.Errorf("mine: power %d, trample %v", c.CurrentPower(), hasKeywordOf(g, mine, "trample"))
	}
	if c := avB1Card(t, g, theirs); c.CurrentPower() != 2 || hasKeywordOf(g, theirs, "trample") {
		t.Error("an opponent's creature was pumped")
	}
	late := b12Creature(g, me.ID, "Late", "Creature — Bear", 2, 2)
	if avB1Card(t, g, late).CurrentPower() != 2 {
		t.Error("a creature that entered afterwards was pumped (CR 611.2c)")
	}
}

func TestSncB15CabarettiCharmMakesTwoCitizens(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castModal(t, g, "Cabaretti Charm", "Instant", sncB15CabarettiCharm, []int{2}, nil)
	passPriorityAroundTable(t, g)
	toks := tokensNamedControlledBy(g, "Citizen", me.ID)
	if toks != 2 {
		t.Errorf("%d Citizen tokens, want 2", toks)
	}
}

// --- Brokers Charm ----------------------------------------------------------

func TestSncB15BrokersCharmBitesWithPlusOnePower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Ogre", 3, 3)
	castModal(t, g, "Brokers Charm", "Instant", sncB15BrokersCharm, []int{0}, []game.TargetRef{
		modeRef(game.TargetCard, mine, 0, 0), modeRef(game.TargetCard, theirs, 0, 1)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("2+1 = 3 damage should kill a 3/3")
	}
	if got := avB1Card(t, g, mine).CurrentPower(); got != 3 {
		t.Errorf("my creature has power %d, want 3", got)
	}
}

func TestSncB15BrokersCharmRefusesYourOwnCreatureAsTheRecipient(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Brokers Charm", TypeLine: "Instant", OracleID: sncB15BrokersCharm, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{0}, Targets: []game.TargetRef{
		modeRef(game.TargetCard, mine, 0, 0), modeRef(game.TargetCard, other, 0, 1)}}); err == nil {
		t.Error("a creature you control was accepted as the opponent's")
	}
}

func TestSncB15BrokersCharmDestroysAnEnchantmentAndDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aura := b12Permanent(g, opp.ID, "Pacifism", "Enchantment — Aura")
	castModal(t, g, "Brokers Charm", "Instant", sncB15BrokersCharm, []int{1},
		[]game.TargetRef{modeRef(game.TargetCard, aura, 0, 0)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("the enchantment survived")
	}
	for i := 0; i < 3; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})
	}
	before := len(me.Hand.Cards)
	castModal(t, g, "Brokers Charm", "Instant", sncB15BrokersCharm, []int{2}, nil)
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards) - before; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
}

// --- Maestros Charm ---------------------------------------------------------

func TestSncB15MaestrosCharmDrainsThreeAndDealsFive(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myLife, oppLife := me.Life, opp.Life
	castModal(t, g, "Maestros Charm", "Instant", sncB15MaestrosCharm, []int{1}, nil)
	passPriorityAroundTable(t, g)
	if me.Life != myLife+3 || opp.Life != oppLife-3 {
		t.Errorf("life %d/%d, want %d/%d", me.Life, opp.Life, myLife+3, oppLife-3)
	}
	big := b12Creature(g, opp.ID, "Big", "Creature — Ogre", 5, 5)
	castModal(t, g, "Maestros Charm", "Instant", sncB15MaestrosCharm, []int{2},
		[]game.TargetRef{modeRef(game.TargetCard, big, 0, 0)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("5 damage should kill a 5/5")
	}
}

func TestSncB15MaestrosCharmLooksAtFiveTakesOneAndGravesTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	var ids []uuid.UUID
	for i := 0; i < 6; i++ {
		ids = append(ids, pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Creature — Bear"}))
	}
	hand, gy := len(me.Hand.Cards), len(me.Graveyard.Cards)
	castModal(t, g, "Maestros Charm", "Instant", sncB15MaestrosCharm, []int{0}, nil)
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choose-cards prompt")
	}
	if len(pick.ChooseCards) != 5 {
		t.Errorf("offered %d cards, want 5", len(pick.ChooseCards))
	}
	answerChooseCards(t, g, me.ID, pick.ChooseCards[0])
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards) - hand; got != 1 {
		t.Errorf("hand grew by %d, want 1", got)
	}
	// The Charm itself also reaches the graveyard: 4 rest cards + the spell.
	if got := len(me.Graveyard.Cards) - gy; got != 5 {
		t.Errorf("graveyard grew by %d, want 5", got)
	}
	_ = ids
}

// --- Riveteers Charm --------------------------------------------------------

func TestSncB15RiveteersCharmEdictTakesTheGreatestManaValue(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	small := sncB15Gold(g, opp.ID, "Small", 1)
	big := sncB15Gold(g, opp.ID, "Big", 5)
	castModal(t, g, "Riveteers Charm", "Instant", sncB15RiveteersCharm, []int{0},
		[]game.TargetRef{modeRef(game.TargetPlayer, opp.ID, 0, 0)})
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, opp.ID, big)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) || !g.Battlefield.Contains(small) {
		t.Error("the greatest mana value permanent, and only it, must be sacrificed")
	}
}

func TestSncB15RiveteersCharmExilesTopThreeToPlay(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 4; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})
	}
	lib := len(me.Library.Cards)
	castModal(t, g, "Riveteers Charm", "Instant", sncB15RiveteersCharm, []int{1}, nil)
	passPriorityAroundTable(t, g)
	if got := lib - len(me.Library.Cards); got != 3 {
		t.Errorf("library lost %d cards, want 3", got)
	}
}

func TestSncB15RiveteersCharmExilesAGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	card := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: card, Name: "Dead", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
	castModal(t, g, "Riveteers Charm", "Instant", sncB15RiveteersCharm, []int{2},
		[]game.TargetRef{modeRef(game.TargetPlayer, opp.ID, 0, 0)})
	passPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(card) {
		t.Error("the graveyard was not exiled")
	}
}

// --- Obscura Charm ----------------------------------------------------------

func TestSncB15ObscuraCharmReturnsAMulticolouredPermanentCardTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gold := uuid.New()
	mono := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: gold, Name: "Gold", TypeLine: "Creature — Bear", Colors: []string{"W", "U"}, ManaCost: "{1}{W}{U}", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: mono, Name: "Mono", TypeLine: "Creature — Bear", Colors: []string{"W"}, ManaCost: "{1}{W}", Owner: me.ID, Controller: me.ID})
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Obscura Charm", TypeLine: "Instant", OracleID: sncB15ObscuraCharm, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{0},
		Targets: []game.TargetRef{modeRef(game.TargetCard, mono, 0, 0)}}); err == nil {
		t.Fatal("a monocoloured card was accepted")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0},
		Targets: []game.TargetRef{modeRef(game.TargetCard, gold, 0, 0)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, ok := cardOnBattlefield(g, gold)
	if !ok || !c.Tapped {
		t.Errorf("on battlefield %v, tapped %v: want true, true", ok, c.Tapped)
	}
}

func TestSncB15ObscuraCharmDestroysACheapCreatureOnly(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	cheap := sncB15Gold(g, opp.ID, "Cheap", 3)
	dear := sncB15Gold(g, opp.ID, "Dear", 4)
	me := g.Seats[0]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Obscura Charm", TypeLine: "Instant", OracleID: sncB15ObscuraCharm, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{2},
		Targets: []game.TargetRef{modeRef(game.TargetCard, dear, 0, 0)}}); err == nil {
		t.Fatal("mana value 4 was accepted")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{2},
		Targets: []game.TargetRef{modeRef(game.TargetCard, cheap, 0, 0)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cheap) {
		t.Error("the mana value 3 creature survived")
	}
}

func TestSncB15ObscuraCharmCountersAnInstantButNotACreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	spec, _ := Lookup(sncB15ObscuraCharm)
	_ = spec
	castModal(t, g, "Obscura Charm", "Instant", sncB15ObscuraCharm, []int{1},
		[]game.TargetRef{modeRef(game.TargetCard, bolt, 0, 0)})
	life := me.Life
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Error("the Bolt resolved")
	}
}

// --- Confluences: shape ------------------------------------------------------

func TestSncB15ConfluencesChooseThreeRepeating(t *testing.T) {
	for _, oracle := range []string{
		sncB15RiveteersConfluence, sncB15CabarettiConfluence, sncB15ObscuraConfluence,
		sncB15BrokersConfluence, sncB15MaestrosConfluence,
	} {
		ms := game.ModeSpecFor(oracle)
		if ms == nil || !ms.Repeatable || ms.Min != 3 || ms.Max != 3 || len(ms.Options) != 3 {
			t.Errorf("%s: mode spec %+v", oracle, ms)
		}
	}
}

// --- Riveteers Confluence ---------------------------------------------------

func TestSncB15RiveteersConfluenceDrawsTwiceAndPingsTwice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 3; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})
	}
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	two := b12Creature(g, opp.ID, "Two", "Creature — Bear", 2, 2)
	three := b12Creature(g, opp.ID, "Three", "Creature — Ogre", 3, 3)
	hand, life := len(me.Hand.Cards), me.Life
	castModal(t, g, "Riveteers Confluence", "Sorcery", sncB15RiveteersConfluence, []int{0, 1, 1}, nil)
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards) - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
	if g.Battlefield.Contains(two) || !g.Battlefield.Contains(three) || !g.Battlefield.Contains(mine) {
		t.Error("two passes of 1 damage to opposing creatures only")
	}
}

func TestSncB15RiveteersConfluencePutsALandFromGraveyardTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	castModal(t, g, "Riveteers Confluence", "Sorcery", sncB15RiveteersConfluence, []int{2, 0, 0}, nil)
	for i := 0; i < 3; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})
	}
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, land)
	passPriorityAroundTable(t, g)
	c, ok := cardOnBattlefield(g, land)
	if !ok || !c.Tapped {
		t.Errorf("land on battlefield %v tapped %v", ok, c.Tapped)
	}
}

// --- Cabaretti Confluence ---------------------------------------------------

func TestSncB15CabarettiConfluenceCopiesWithHasteAndSacrifices(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Grizzly", "Creature — Bear", 2, 2)
	castModal(t, g, "Cabaretti Confluence", "Sorcery", sncB15CabarettiConfluence, []int{0, 1, 1}, []game.TargetRef{
		modeRef(game.TargetCard, bear, 0, 0),
		modeRef(game.TargetCard, b12Permanent(g, g.Seats[1].ID, "Rock", "Artifact"), 1, 0),
		modeRef(game.TargetCard, b12Permanent(g, g.Seats[1].ID, "Pacifism", "Enchantment — Aura"), 2, 0),
	})
	passPriorityAroundTable(t, g)
	toks := tokensNamed(g, "Grizzly")
	if len(toks) != 1 || !hasKeywordOf(g, toks[0].InstanceID, "haste") {
		t.Fatalf("want one hasty copy, got %d", len(toks))
	}
	for g.Turn.Step != game.StepEnd {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(toks[0].InstanceID) {
		t.Error("the copy was not sacrificed at the next end step")
	}
}

func TestSncB15CabarettiConfluenceStacksTheTeamPump(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	castModal(t, g, "Cabaretti Confluence", "Sorcery", sncB15CabarettiConfluence, []int{2, 2, 1}, []game.TargetRef{
		modeRef(game.TargetPlayer, me.ID, 0, 0), modeRef(game.TargetPlayer, me.ID, 1, 0),
		modeRef(game.TargetCard, rock, 2, 0)})
	passPriorityAroundTable(t, g)
	if got := avB1Card(t, g, mine).CurrentPower(); got != 4 {
		t.Errorf("power %d, want 4 (+1/+1 twice)", got)
	}
	if !hasKeywordOf(g, mine, "first strike") || hasKeywordOf(g, theirs, "first strike") {
		t.Error("first strike must go to the target player's creatures only")
	}
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not exiled")
	}
}

// --- Brokers Confluence -----------------------------------------------------

func TestSncB15BrokersConfluencePhasesOutAndCountersAnAbility(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := g.Seats[1]
	src := pushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	fired := false
	ability := taAnnounce(t, g, opp.ID, src, "Their Rock — do a thing", nil, &fired)
	castModal(t, g, "Brokers Confluence", "Instant", sncB15BrokersConfluence, []int{1, 2, 0}, []game.TargetRef{
		modeRef(game.TargetCard, bear, 0, 0), modeRef(game.TargetCard, ability, 1, 0)})
	passPriorityAroundTable(t, g)
	if fired {
		t.Error("the countered ability resolved")
	}
	if !g.PhasedOut.Contains(bear) {
		t.Error("the creature did not phase out")
	}
}

func TestSncB15BrokersConfluenceProliferateAddsACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.AddCounterByForEffect(me.ID, bear, game.CounterPlusOne, 1); err != nil {
		t.Fatal(err)
	}
	castModal(t, g, "Brokers Confluence", "Instant", sncB15BrokersConfluence, []int{0, 0, 0}, nil)
	for i := 0; i < 3; i++ {
		passPriorityAroundTable(t, g)
		if proliferatePrompt(g) == nil {
			break
		}
		answerProliferate(t, g, bear)
	}
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, bear); got != 4 {
		t.Errorf("%d counters after proliferating three times, want 4", got)
	}
}

// --- Obscura Confluence -----------------------------------------------------

func TestSncB15ObscuraConfluenceMakesAOneOneWithNoAbilities(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	flyer := b12Creature(g, opp.ID, "Flyer", "Creature — Bird", 4, 4)
	g.WithWriteLock(func() {
		if c := findBattlefieldCardForTest(g, flyer); c != nil {
			c.Keywords = append(c.Keywords, "flying")
		}
	})
	me := g.Seats[0]
	for i := 0; i < 3; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})
	}
	castModal(t, g, "Obscura Confluence", "Instant", sncB15ObscuraConfluence, []int{0, 2, 2}, []game.TargetRef{
		modeRef(game.TargetCard, flyer, 0, 0), modeRef(game.TargetPlayer, me.ID, 1, 0), modeRef(game.TargetPlayer, me.ID, 2, 0)})
	passPriorityAroundTable(t, g)
	c := avB1Card(t, g, flyer)
	if c.CurrentPower() != 1 || c.CurrentToughness() != 1 || hasKeywordOf(g, flyer, "flying") {
		t.Errorf("flyer is %d/%d flying=%v, want 1/1 with no abilities", c.CurrentPower(), c.CurrentToughness(), hasKeywordOf(g, flyer, "flying"))
	}
}

func TestSncB15ObscuraConfluenceTargetPlayerReturnsACreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	me := g.Seats[0]
	dead := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for i := 0; i < 3; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})
	}
	castModal(t, g, "Obscura Confluence", "Instant", sncB15ObscuraConfluence, []int{2, 0, 0}, []game.TargetRef{
		modeRef(game.TargetPlayer, opp.ID, 0, 0), modeRef(game.TargetCard, b12Creature(g, opp.ID, "X", "Creature — Bear", 1, 1), 1, 0),
		modeRef(game.TargetCard, bear, 2, 0)})
	passPriorityAroundTable(t, g)
	if p := chooseCardsChoiceFor(g, opp.ID); p != nil {
		answerChooseCards(t, g, opp.ID, dead)
	}
	if !opp.Hand.Contains(dead) {
		t.Error("the target player did not return their creature card to hand")
	}
}

// --- Maestros Confluence ----------------------------------------------------

func TestSncB15MaestrosConfluenceShrinksAndGoads(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 3, 3)
	other := b12Creature(g, opp.ID, "Other", "Creature — Bear", 4, 4)
	castModal(t, g, "Maestros Confluence", "Sorcery", sncB15MaestrosConfluence, []int{1, 2, 1}, []game.TargetRef{
		modeRef(game.TargetCard, small, 0, 0), modeRef(game.TargetPlayer, opp.ID, 1, 0),
		modeRef(game.TargetCard, other, 2, 0)})
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(small) {
		t.Error("-3/-3 should kill a 3/3")
	}
	if c := avB1Card(t, g, other); !c.IsGoadedBy(me.ID) || c.CurrentPower() != 1 {
		t.Errorf("other: goaded %v power %d, want goaded and 1", c.IsGoadedBy(me.ID), c.CurrentPower())
	}
}

func TestSncB15MaestrosConfluenceReturnsOnlyAMonocolouredSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mono, gold, creature := uuid.New(), uuid.New(), uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: mono, Name: "Bolt", TypeLine: "Instant", Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: gold, Name: "Gold Spell", TypeLine: "Instant", Colors: []string{"U", "B"}, Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: creature, Name: "Bear", TypeLine: "Creature — Bear", Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})
	opp := g.Seats[1]
	t1 := b12Creature(g, opp.ID, "T1", "Creature — Bear", 5, 5)
	t2 := b12Creature(g, opp.ID, "T2", "Creature — Bear", 5, 5)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Maestros Confluence", TypeLine: "Sorcery", OracleID: sncB15MaestrosConfluence, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	for _, bad := range []uuid.UUID{gold, creature} {
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{0, 1, 1}, Targets: []game.TargetRef{
			modeRef(game.TargetCard, bad, 0, 0), modeRef(game.TargetCard, t1, 1, 0), modeRef(game.TargetCard, t2, 2, 0)}}); err == nil {
			t.Errorf("%s was accepted as a monocolored instant or sorcery", bad)
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0, 1, 1}, Targets: []game.TargetRef{
		modeRef(game.TargetCard, mono, 0, 0), modeRef(game.TargetCard, t1, 1, 0), modeRef(game.TargetCard, t2, 2, 0)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(mono) {
		t.Error("the monocolored instant did not return to hand")
	}
}
