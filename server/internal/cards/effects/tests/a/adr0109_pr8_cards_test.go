package testsa

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Every card declares the component its text prints, and is Full.
func TestADR0109PR8CardsDeclareTheirCosts(t *testing.T) {
	type want struct {
		random, top, library int
	}
	for oracle, w := range map[string]want{
		effectstest.P8StormscaleAnarch: {random: 1}, effectstest.P8MageIlVec: {random: 1}, effectstest.P8FreneticOgre: {random: 1},
		effectstest.P8CanyonDrake: {random: 1}, effectstest.P8CoralHelm: {random: 1}, effectstest.P8Pyromancy: {random: 1},
		effectstest.P8MeteorStorm: {random: 2}, effectstest.P8PardicSwordsmith: {random: 1}, effectstest.P8Stormbind: {random: 1},
		effectstest.P8Amok: {random: 1}, effectstest.P8Pyromania: {random: 1}, effectstest.P8HellBentRaider: {random: 1},
		effectstest.P8DwarvenStrike: {random: 1}, effectstest.P8DraconianCylix: {random: 1}, effectstest.P8OgreShaman: {random: 1},
		effectstest.P8PardicLancer:     {random: 1},
		effectstest.P8WhirlingCatapult: {library: 2}, effectstest.P8ArcSlogger: {library: 10}, effectstest.P8SeasonedTactic: {library: 4},
		effectstest.P8RoyalHerbalist: {library: 1}, effectstest.P8StormElemental: {library: 1}, effectstest.P8PhyrexianDevour: {library: 1},
		effectstest.P8Leashling: {top: 1}, effectstest.P8Penance: {top: 1},
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: completeness %v, want Full", spec.Name, spec.Completeness)
		}
		if len(spec.Activated) == 0 {
			t.Errorf("%s has no activated ability", spec.Name)
			continue
		}
		c := spec.Activated[0].Cost
		if c.RandomDiscardCount() != w.random || c.PutFromHandOnLibraryTop != w.top || c.ExileFromLibraryTop != w.library {
			t.Errorf("%s: cost random=%d top=%d library=%d, want %+v", spec.Name,
				c.RandomDiscardCount(), c.PutFromHandOnLibraryTop, c.ExileFromLibraryTop, w)
		}
	}
}

func p8Push(g *game.Game, me *game.Player, name, oracle, typeLine string) uuid.UUID {
	return effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: name, OracleID: oracle, TypeLine: typeLine, Power: 1, Toughness: 1})
}

// Pyromancy deals the discarded card's mana value; Stormscale Anarch
// deals 4 for a multicolored card and 2 otherwise (ADR 0109 §8's read
// of the random discard).
func TestPyromancyAndStormscaleAnarchReadTheRandomDiscard(t *testing.T) {
	g, me, opp := effectstest.P7Table(t)
	pyro := p8Push(g, me, "Pyromancy", effectstest.P8Pyromancy, "Enchantment")
	p7Hand(me, "Three Drop", "Sorcery", "{2}{R}")
	life := opp.Life
	effectstest.P7Activate(t, g, me, pyro, 0, game.ActivateAbilityParams{Targets: effectstest.P8Player(opp.ID)})
	if opp.Life != life-3 {
		t.Fatalf("Pyromancy discarding a three-drop: life %d, want %d", opp.Life, life-3)
	}

	for _, tc := range []struct {
		name   string
		colors []string
		want   int
	}{{"multicolored", []string{"B", "R"}, 4}, {"one colour", []string{"R"}, 2}} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp := effectstest.P7Table(t)
			anarch := p8Push(g, me, "Stormscale Anarch", effectstest.P8StormscaleAnarch, "Creature — Lizard Shaman")
			me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Gold Card", TypeLine: "Sorcery",
				ManaCost: "{B}{R}", Colors: tc.colors, Owner: me.ID, Controller: me.ID})
			life := opp.Life
			effectstest.P7Activate(t, g, me, anarch, 0, game.ActivateAbilityParams{Targets: effectstest.P8Player(opp.ID)})
			if opp.Life != life-tc.want {
				t.Errorf("life %d, want %d", opp.Life, life-tc.want)
			}
		})
	}
}

// Meteor Storm needs two cards in hand and discards both.
func TestMeteorStormDiscardsTwoAtRandom(t *testing.T) {
	g, me, opp := effectstest.P7Table(t)
	storm := p8Push(g, me, "Meteor Storm", effectstest.P8MeteorStorm, "Enchantment")
	p7Hand(me, "Only Card", "Sorcery", "{R}")
	if err := g.ActivateCatalogAbility(me.ID, storm, 0, game.ActivateAbilityParams{Targets: effectstest.P8Player(opp.ID)}); !errors.Is(err, game.ErrCantPayRandomDiscard) {
		t.Fatalf("a hand of one: %v, want ErrCantPayRandomDiscard", err)
	}
	p7Hand(me, "Second Card", "Sorcery", "{R}")
	life := opp.Life
	effectstest.P7Activate(t, g, me, storm, 0, game.ActivateAbilityParams{Targets: effectstest.P8Player(opp.ID)})
	if opp.Life != life-4 || me.Hand.Size() != 0 || me.Graveyard.Size() != 2 {
		t.Errorf("life %d (want %d), hand %d, graveyard %d", opp.Life, life-4, me.Hand.Size(), me.Graveyard.Size())
	}
}

// The self-pumping random discarders change only their own source.
func TestRandomDiscardPumpsChangeTheirSource(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		power        int
		keywords     []string
	}{
		{"Frenetic Ogre", effectstest.P8FreneticOgre, 3, nil},
		{"Canyon Drake", effectstest.P8CanyonDrake, 2, nil},
		{"Pardic Swordsmith", effectstest.P8PardicSwordsmith, 2, nil},
		{"Pardic Lancer", effectstest.P8PardicLancer, 1, []string{"first strike"}},
		{"Dwarven Strike Force", effectstest.P8DwarvenStrike, 0, []string{"first strike", "haste"}},
		{"Hell-Bent Raider", effectstest.P8HellBentRaider, 0, []string{"protection from white"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := effectstest.P7Table(t)
			src := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature(tc.name, tc.oracle, 2, 2))
			p7Hand(me, "Fodder", "Sorcery", "{R}")
			effectstest.P7Activate(t, g, me, src, 0, game.ActivateAbilityParams{})
			live := effectstest.ApaLive(g, src)
			if got := live.CurrentPower(); got != 2+tc.power {
				t.Errorf("power %d, want %d", got, 2+tc.power)
			}
			for _, kw := range tc.keywords {
				if !game.HasKeyword(live, kw) {
					t.Errorf("no %s", kw)
				}
			}
			if me.Hand.Size() != 0 {
				t.Error("the card was not discarded")
			}
		})
	}
}

// The targeted random discarders reach their targets.
func TestRandomDiscardTargetedAbilities(t *testing.T) {
	g, me, opp := effectstest.P7Table(t)
	bear := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Bear", "", 2, 2))
	helm := p8Push(g, me, "Coral Helm", effectstest.P8CoralHelm, "Artifact")
	amok := p8Push(g, me, "Amok", effectstest.P8Amok, "Enchantment")
	for i := 0; i < 2; i++ {
		p7Hand(me, "Fodder", "Sorcery", "{R}")
	}
	effectstest.P7Activate(t, g, me, helm, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	effectstest.P7Activate(t, g, me, amok, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	live := effectstest.ApaLive(g, bear)
	if live.CurrentPower() != 5 || live.Counters["+1/+1"] != 1 {
		t.Errorf("bear power %d with %d counters, want 5 with one", live.CurrentPower(), live.Counters["+1/+1"])
	}
	for _, tc := range []struct {
		name, oracle string
		amount       int
		row          int
	}{
		{"Mage il-Vec", effectstest.P8MageIlVec, 1, 0},
		{"Stormbind", effectstest.P8Stormbind, 2, 0},
		{"Ogre Shaman", effectstest.P8OgreShaman, 2, 0},
		{"Pyromania", effectstest.P8Pyromania, 1, 0},
		{"Pyromania", effectstest.P8Pyromania, 1, 1},
	} {
		g, me, opp := effectstest.P7Table(t)
		src := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature(tc.name, tc.oracle, 2, 2))
		p7Hand(me, "Fodder", "Sorcery", "{R}")
		life := opp.Life
		effectstest.P7Activate(t, g, me, src, tc.row, game.ActivateAbilityParams{Targets: effectstest.P8Player(opp.ID)})
		if opp.Life != life-tc.amount {
			t.Errorf("%s row %d: life %d, want %d", tc.name, tc.row, opp.Life, life-tc.amount)
		}
	}
}

// Draconian Cylix regenerates its target.
func TestDraconianCylixRegenerates(t *testing.T) {
	g, me, _ := effectstest.P7Table(t)
	bear := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Bear", "", 2, 2))
	cylix := p8Push(g, me, "Draconian Cylix", effectstest.P8DraconianCylix, "Artifact")
	p7Hand(me, "Fodder", "Sorcery", "{R}")
	effectstest.P7Activate(t, g, me, cylix, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the regenerated creature was destroyed")
	}
}

// Arc-Slogger needs ten cards in the library and exiles them.
func TestArcSloggerExilesTenAndDealsTwo(t *testing.T) {
	g, me, opp := effectstest.P7Table(t)
	slogger := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Arc-Slogger", effectstest.P8ArcSlogger, 4, 5))
	effectstest.LcLibrary(g, me, 9)
	if err := g.ActivateCatalogAbility(me.ID, slogger, 0, game.ActivateAbilityParams{Targets: effectstest.P8Player(opp.ID)}); !errors.Is(err, game.ErrCantPayLibraryCost) {
		t.Fatalf("nine cards: %v, want ErrCantPayLibraryCost", err)
	}
	effectstest.LcLibrary(g, me, 12)
	life := opp.Life
	effectstest.P7Activate(t, g, me, slogger, 0, game.ActivateAbilityParams{Targets: effectstest.P8Player(opp.ID)})
	if opp.Life != life-2 || me.Library.Size() != 2 {
		t.Errorf("life %d (want %d), library %d (want 2)", opp.Life, life-2, me.Library.Size())
	}
}

// Whirling Catapult hits each creature with flying and each player.
func TestWhirlingCatapult(t *testing.T) {
	g, me, opp := effectstest.P7Table(t)
	bird := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Bird", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	bear := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Bear", "", 2, 2))
	cat := p8Push(g, me, "Whirling Catapult", effectstest.P8WhirlingCatapult, "Artifact")
	effectstest.LcLibrary(g, me, 2)
	myLife, theirLife := me.Life, opp.Life
	effectstest.P7Activate(t, g, me, cat, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(bird) || !g.Battlefield.Contains(bear) {
		t.Error("the flyer should die and the bear survive")
	}
	if me.Life != myLife-1 || opp.Life != theirLife-1 {
		t.Error("each player takes 1")
	}
}

// Royal Herbalist gains a life for the top card.
func TestRoyalHerbalist(t *testing.T) {
	g, me, _ := effectstest.P7Table(t)
	herb := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Royal Herbalist", effectstest.P8RoyalHerbalist, 1, 1))
	effectstest.LcLibrary(g, me, 1)
	life := me.Life
	effectstest.P7Activate(t, g, me, herb, 0, game.ActivateAbilityParams{})
	if me.Life != life+1 || me.Library.Size() != 0 {
		t.Errorf("life %d (want %d), library %d", me.Life, life+1, me.Library.Size())
	}
}

// Storm Elemental: the first row taps a flyer; the second pumps only
// for a snow land (CR 400.7j).
func TestStormElemental(t *testing.T) {
	g, me, opp := effectstest.P7Table(t)
	elemental := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Storm Elemental", OracleID: effectstest.P8StormElemental,
		TypeLine: "Creature — Elemental", Power: 3, Toughness: 4})
	bird := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Bird", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	effectstest.LcLibrary(g, me, 3)
	effectstest.P7Activate(t, g, me, elemental, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bird)})
	if !effectstest.ApaLive(g, bird).Tapped {
		t.Error("the flyer is not tapped")
	}
	effectstest.P8LibraryTop(g, me, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
	effectstest.P7Activate(t, g, me, elemental, 1, game.ActivateAbilityParams{})
	if effectstest.ApaLive(g, elemental).CurrentPower() != 3 {
		t.Error("a nonsnow land pumped it")
	}
	effectstest.P8LibraryTop(g, me, game.Card{Name: "Snow-Covered Island", TypeLine: "Basic Snow Land — Island"})
	effectstest.P7Activate(t, g, me, elemental, 1, game.ActivateAbilityParams{})
	if got := effectstest.ApaLive(g, elemental); got.CurrentPower() != 4 || got.CurrentToughness() != 5 {
		t.Errorf("after a snow land: %d/%d, want 4/5", got.CurrentPower(), got.CurrentToughness())
	}
}

// Phyrexian Devourer grows by the exiled card's mana value and is
// sacrificed by its state trigger at power 7 (CR 603.8).
func TestPhyrexianDevourer(t *testing.T) {
	g, me, _ := effectstest.P7Table(t)
	dev := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Phyrexian Devourer", OracleID: effectstest.P8PhyrexianDevour,
		TypeLine: "Artifact Creature — Phyrexian Construct", Power: 1, Toughness: 1})
	effectstest.LcLibrary(g, me, 1)
	effectstest.P8LibraryTop(g, me, game.Card{Name: "Three Drop", TypeLine: "Sorcery", ManaCost: "{2}{R}"})
	effectstest.P7Activate(t, g, me, dev, 0, game.ActivateAbilityParams{})
	if got := effectstest.ApaLive(g, dev).Counters["+1/+1"]; got != 3 {
		t.Fatalf("counters %d, want 3", got)
	}
	effectstest.P8LibraryTop(g, me, game.Card{Name: "Four Drop", TypeLine: "Sorcery", ManaCost: "{3}{R}"})
	effectstest.P7Activate(t, g, me, dev, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dev) {
		t.Fatal("a Devourer with power 8 was not sacrificed")
	}
}

// Leashling returns itself, paid with a card put on top of the library.
func TestLeashling(t *testing.T) {
	g, me, _ := effectstest.P7Table(t)
	leash := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Leashling", effectstest.P8Leashling, 3, 3))
	card := p7Hand(me, "Card", "Sorcery", "{R}")
	effectstest.P7Activate(t, g, me, leash, 0, game.ActivateAbilityParams{TopIDs: []uuid.UUID{card}})
	if !me.Hand.Contains(leash) || g.Battlefield.Contains(leash) {
		t.Error("Leashling did not return to its owner's hand")
	}
	if top, _ := me.Library.Top(); top.InstanceID != card {
		t.Error("the paid card is not on top of the library")
	}
}

// Penance and Seasoned Tactician pay their costs and ask for the
// source as the shield is made (CR 609.7a).
func TestPenanceAndSeasonedTacticianMakeTheirShields(t *testing.T) {
	g, me, _ := effectstest.P7Table(t)
	pen := p8Push(g, me, "Penance", effectstest.P8Penance, "Enchantment")
	card := p7Hand(me, "Card", "Sorcery", "{R}")
	if err := g.ActivateCatalogAbility(me.ID, pen, 0, game.ActivateAbilityParams{TopIDs: []uuid.UUID{card}}); err != nil {
		t.Fatalf("Penance: %v", err)
	}
	if top, _ := me.Library.Top(); top.InstanceID != card {
		t.Error("Penance's card is not on top of the library")
	}

	g, me, _ = effectstest.P7Table(t)
	tac := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Seasoned Tactician", effectstest.P8SeasonedTactic, 1, 3))
	effectstest.LcLibrary(g, me, 3)
	if err := g.ActivateCatalogAbility(me.ID, tac, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrCantPayLibraryCost) {
		t.Fatalf("three cards: %v, want ErrCantPayLibraryCost", err)
	}
	effectstest.LcLibrary(g, me, 4)
	if err := g.ActivateCatalogAbility(me.ID, tac, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Seasoned Tactician: %v", err)
	}
	if me.Library.Size() != 0 || len(effectstest.OnlyStackItemOf(t, g).Paid.Exiled) != 4 {
		t.Error("the top four cards were not exiled and recorded")
	}
}
