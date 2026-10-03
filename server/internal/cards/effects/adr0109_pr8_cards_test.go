package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adr0109_pr8_cards_test.go — the cards ADR 0109 §7 (#1902) and owner
// decision 3 unblock: costs that put a card from a hand on top of a
// library, that exile the top of a library, and that discard at
// random. The engine half is library_costs_test.go.

const (
	p8StormscaleAnarch = "819a7235-b61a-49d7-a0ad-3d8bbf83091e"
	p8MageIlVec        = "adea7db6-634c-4c5d-be40-264b4acffc53"
	p8FreneticOgre     = "6eb4b649-c578-4b29-8da1-7c757412bec0"
	p8CanyonDrake      = "b5b46c99-1cc1-465a-a831-3fd6665ef560"
	p8CoralHelm        = "aa2970c8-f2ea-4e06-8b8f-ec89af0012a0"
	p8Pyromancy        = "71a176ce-cbe7-4cec-a2ca-58abf59ad964"
	p8MeteorStorm      = "03f96c23-de0b-4b85-81f8-febc850aa621"
	p8PardicSwordsmith = "946d770f-0a59-415b-b334-8f7353b96046"
	p8Stormbind        = "78f50668-36fc-4911-84f7-93667436b0c7"
	p8Amok             = "2ddbbe65-f928-4b8e-8c4c-a9ce82e2594d"
	p8Pyromania        = "2c89cf19-8dc6-422e-bbf8-4dd5a399c630"
	p8HellBentRaider   = "ee8ab29b-0749-463a-abbd-eb0b9013aec8"
	p8DwarvenStrike    = "66dd0e68-f5ec-45e5-991f-d588aa726387"
	p8DraconianCylix   = "0880b36a-6141-4955-b532-cf88daca0869"
	p8OgreShaman       = "04c7bb20-6e40-4e79-a0ec-ced920d3491e"
	p8PardicLancer     = "1bb2b090-19a4-40d9-823a-671f6eaca9f9"
	p8WhirlingCatapult = "0987b5e9-0012-4189-8745-45f10e9557f3"
	p8ArcSlogger       = "a28fc506-a98c-4b4d-a0f4-7d971489ccbb"
	p8SeasonedTactic   = "d7c2f27c-8c82-4283-a17d-f979ee636542"
	p8RoyalHerbalist   = "b3baf498-3be7-4a22-8fed-806d8d8ac748"
	p8StormElemental   = "f1fee486-660e-4356-896e-671e8b675ad8"
	p8PhyrexianDevour  = "28b1ed69-784c-4642-86c7-104792d5afc2"
	p8Leashling        = "8460abd9-ab5d-4a7e-8132-185b4a310190"
	p8Penance          = "7818d6e0-74d8-46ae-94f6-d2a349c9507b"
)

// Every card declares the component its text prints, and is Full.
func TestADR0109PR8CardsDeclareTheirCosts(t *testing.T) {
	type want struct {
		random, top, library int
	}
	for oracle, w := range map[string]want{
		p8StormscaleAnarch: {random: 1}, p8MageIlVec: {random: 1}, p8FreneticOgre: {random: 1},
		p8CanyonDrake: {random: 1}, p8CoralHelm: {random: 1}, p8Pyromancy: {random: 1},
		p8MeteorStorm: {random: 2}, p8PardicSwordsmith: {random: 1}, p8Stormbind: {random: 1},
		p8Amok: {random: 1}, p8Pyromania: {random: 1}, p8HellBentRaider: {random: 1},
		p8DwarvenStrike: {random: 1}, p8DraconianCylix: {random: 1}, p8OgreShaman: {random: 1},
		p8PardicLancer:     {random: 1},
		p8WhirlingCatapult: {library: 2}, p8ArcSlogger: {library: 10}, p8SeasonedTactic: {library: 4},
		p8RoyalHerbalist: {library: 1}, p8StormElemental: {library: 1}, p8PhyrexianDevour: {library: 1},
		p8Leashling: {top: 1}, p8Penance: {top: 1},
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != CompletenessFull {
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
	return apaPush(g, me.ID, me.ID, game.Card{Name: name, OracleID: oracle, TypeLine: typeLine, Power: 1, Toughness: 1})
}

func p8LibraryTop(g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	g.WithWriteLock(func() { p.Library.PushTop(c) })
	return c.InstanceID
}

func p8Player(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: id}}
}

// Pyromancy deals the discarded card's mana value; Stormscale Anarch
// deals 4 for a multicolored card and 2 otherwise (ADR 0109 §8's read
// of the random discard).
func TestPyromancyAndStormscaleAnarchReadTheRandomDiscard(t *testing.T) {
	g, me, opp := p7Table(t)
	pyro := p8Push(g, me, "Pyromancy", p8Pyromancy, "Enchantment")
	p7Hand(me, "Three Drop", "Sorcery", "{2}{R}")
	life := opp.Life
	p7Activate(t, g, me, pyro, 0, game.ActivateAbilityParams{Targets: p8Player(opp.ID)})
	if opp.Life != life-3 {
		t.Fatalf("Pyromancy discarding a three-drop: life %d, want %d", opp.Life, life-3)
	}

	for _, tc := range []struct {
		name   string
		colors []string
		want   int
	}{{"multicolored", []string{"B", "R"}, 4}, {"one colour", []string{"R"}, 2}} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp := p7Table(t)
			anarch := p8Push(g, me, "Stormscale Anarch", p8StormscaleAnarch, "Creature — Lizard Shaman")
			me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Gold Card", TypeLine: "Sorcery",
				ManaCost: "{B}{R}", Colors: tc.colors, Owner: me.ID, Controller: me.ID})
			life := opp.Life
			p7Activate(t, g, me, anarch, 0, game.ActivateAbilityParams{Targets: p8Player(opp.ID)})
			if opp.Life != life-tc.want {
				t.Errorf("life %d, want %d", opp.Life, life-tc.want)
			}
		})
	}
}

// Meteor Storm needs two cards in hand and discards both.
func TestMeteorStormDiscardsTwoAtRandom(t *testing.T) {
	g, me, opp := p7Table(t)
	storm := p8Push(g, me, "Meteor Storm", p8MeteorStorm, "Enchantment")
	p7Hand(me, "Only Card", "Sorcery", "{R}")
	if err := g.ActivateCatalogAbility(me.ID, storm, 0, game.ActivateAbilityParams{Targets: p8Player(opp.ID)}); !errors.Is(err, game.ErrCantPayRandomDiscard) {
		t.Fatalf("a hand of one: %v, want ErrCantPayRandomDiscard", err)
	}
	p7Hand(me, "Second Card", "Sorcery", "{R}")
	life := opp.Life
	p7Activate(t, g, me, storm, 0, game.ActivateAbilityParams{Targets: p8Player(opp.ID)})
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
		{"Frenetic Ogre", p8FreneticOgre, 3, nil},
		{"Canyon Drake", p8CanyonDrake, 2, nil},
		{"Pardic Swordsmith", p8PardicSwordsmith, 2, nil},
		{"Pardic Lancer", p8PardicLancer, 1, []string{"first strike"}},
		{"Dwarven Strike Force", p8DwarvenStrike, 0, []string{"first strike", "haste"}},
		{"Hell-Bent Raider", p8HellBentRaider, 0, []string{"protection from white"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := p7Table(t)
			src := apaPush(g, me.ID, me.ID, apaCreature(tc.name, tc.oracle, 2, 2))
			p7Hand(me, "Fodder", "Sorcery", "{R}")
			p7Activate(t, g, me, src, 0, game.ActivateAbilityParams{})
			live := apaLive(g, src)
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
	g, me, opp := p7Table(t)
	bear := apaPush(g, opp.ID, opp.ID, apaCreature("Bear", "", 2, 2))
	helm := p8Push(g, me, "Coral Helm", p8CoralHelm, "Artifact")
	amok := p8Push(g, me, "Amok", p8Amok, "Enchantment")
	for i := 0; i < 2; i++ {
		p7Hand(me, "Fodder", "Sorcery", "{R}")
	}
	p7Activate(t, g, me, helm, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
	p7Activate(t, g, me, amok, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
	live := apaLive(g, bear)
	if live.CurrentPower() != 5 || live.Counters["+1/+1"] != 1 {
		t.Errorf("bear power %d with %d counters, want 5 with one", live.CurrentPower(), live.Counters["+1/+1"])
	}
	for _, tc := range []struct {
		name, oracle string
		amount       int
		row          int
	}{
		{"Mage il-Vec", p8MageIlVec, 1, 0},
		{"Stormbind", p8Stormbind, 2, 0},
		{"Ogre Shaman", p8OgreShaman, 2, 0},
		{"Pyromania", p8Pyromania, 1, 0},
		{"Pyromania", p8Pyromania, 1, 1},
	} {
		g, me, opp := p7Table(t)
		src := apaPush(g, me.ID, me.ID, apaCreature(tc.name, tc.oracle, 2, 2))
		p7Hand(me, "Fodder", "Sorcery", "{R}")
		life := opp.Life
		p7Activate(t, g, me, src, tc.row, game.ActivateAbilityParams{Targets: p8Player(opp.ID)})
		if opp.Life != life-tc.amount {
			t.Errorf("%s row %d: life %d, want %d", tc.name, tc.row, opp.Life, life-tc.amount)
		}
	}
}

// Draconian Cylix regenerates its target.
func TestDraconianCylixRegenerates(t *testing.T) {
	g, me, _ := p7Table(t)
	bear := apaPush(g, me.ID, me.ID, apaCreature("Bear", "", 2, 2))
	cylix := p8Push(g, me, "Draconian Cylix", p8DraconianCylix, "Artifact")
	p7Hand(me, "Fodder", "Sorcery", "{R}")
	p7Activate(t, g, me, cylix, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the regenerated creature was destroyed")
	}
}

// Arc-Slogger needs ten cards in the library and exiles them.
func TestArcSloggerExilesTenAndDealsTwo(t *testing.T) {
	g, me, opp := p7Table(t)
	slogger := apaPush(g, me.ID, me.ID, apaCreature("Arc-Slogger", p8ArcSlogger, 4, 5))
	lcLibrary(g, me, 9)
	if err := g.ActivateCatalogAbility(me.ID, slogger, 0, game.ActivateAbilityParams{Targets: p8Player(opp.ID)}); !errors.Is(err, game.ErrCantPayLibraryCost) {
		t.Fatalf("nine cards: %v, want ErrCantPayLibraryCost", err)
	}
	lcLibrary(g, me, 12)
	life := opp.Life
	p7Activate(t, g, me, slogger, 0, game.ActivateAbilityParams{Targets: p8Player(opp.ID)})
	if opp.Life != life-2 || me.Library.Size() != 2 {
		t.Errorf("life %d (want %d), library %d (want 2)", opp.Life, life-2, me.Library.Size())
	}
}

// Whirling Catapult hits each creature with flying and each player.
func TestWhirlingCatapult(t *testing.T) {
	g, me, opp := p7Table(t)
	bird := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Bird", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	bear := apaPush(g, opp.ID, opp.ID, apaCreature("Bear", "", 2, 2))
	cat := p8Push(g, me, "Whirling Catapult", p8WhirlingCatapult, "Artifact")
	lcLibrary(g, me, 2)
	myLife, theirLife := me.Life, opp.Life
	p7Activate(t, g, me, cat, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(bird) || !g.Battlefield.Contains(bear) {
		t.Error("the flyer should die and the bear survive")
	}
	if me.Life != myLife-1 || opp.Life != theirLife-1 {
		t.Error("each player takes 1")
	}
}

// Royal Herbalist gains a life for the top card.
func TestRoyalHerbalist(t *testing.T) {
	g, me, _ := p7Table(t)
	herb := apaPush(g, me.ID, me.ID, apaCreature("Royal Herbalist", p8RoyalHerbalist, 1, 1))
	lcLibrary(g, me, 1)
	life := me.Life
	p7Activate(t, g, me, herb, 0, game.ActivateAbilityParams{})
	if me.Life != life+1 || me.Library.Size() != 0 {
		t.Errorf("life %d (want %d), library %d", me.Life, life+1, me.Library.Size())
	}
}

// Storm Elemental: the first row taps a flyer; the second pumps only
// for a snow land (CR 400.7j).
func TestStormElemental(t *testing.T) {
	g, me, opp := p7Table(t)
	elemental := apaPush(g, me.ID, me.ID, game.Card{Name: "Storm Elemental", OracleID: p8StormElemental,
		TypeLine: "Creature — Elemental", Power: 3, Toughness: 4})
	bird := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Bird", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	lcLibrary(g, me, 3)
	p7Activate(t, g, me, elemental, 0, game.ActivateAbilityParams{Targets: cardRefs(bird)})
	if !apaLive(g, bird).Tapped {
		t.Error("the flyer is not tapped")
	}
	p8LibraryTop(g, me, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
	p7Activate(t, g, me, elemental, 1, game.ActivateAbilityParams{})
	if apaLive(g, elemental).CurrentPower() != 3 {
		t.Error("a nonsnow land pumped it")
	}
	p8LibraryTop(g, me, game.Card{Name: "Snow-Covered Island", TypeLine: "Basic Snow Land — Island"})
	p7Activate(t, g, me, elemental, 1, game.ActivateAbilityParams{})
	if got := apaLive(g, elemental); got.CurrentPower() != 4 || got.CurrentToughness() != 5 {
		t.Errorf("after a snow land: %d/%d, want 4/5", got.CurrentPower(), got.CurrentToughness())
	}
}

// Phyrexian Devourer grows by the exiled card's mana value and is
// sacrificed by its state trigger at power 7 (CR 603.8).
func TestPhyrexianDevourer(t *testing.T) {
	g, me, _ := p7Table(t)
	dev := apaPush(g, me.ID, me.ID, game.Card{Name: "Phyrexian Devourer", OracleID: p8PhyrexianDevour,
		TypeLine: "Artifact Creature — Phyrexian Construct", Power: 1, Toughness: 1})
	lcLibrary(g, me, 1)
	p8LibraryTop(g, me, game.Card{Name: "Three Drop", TypeLine: "Sorcery", ManaCost: "{2}{R}"})
	p7Activate(t, g, me, dev, 0, game.ActivateAbilityParams{})
	if got := apaLive(g, dev).Counters["+1/+1"]; got != 3 {
		t.Fatalf("counters %d, want 3", got)
	}
	p8LibraryTop(g, me, game.Card{Name: "Four Drop", TypeLine: "Sorcery", ManaCost: "{3}{R}"})
	p7Activate(t, g, me, dev, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dev) {
		t.Fatal("a Devourer with power 8 was not sacrificed")
	}
}

// Leashling returns itself, paid with a card put on top of the library.
func TestLeashling(t *testing.T) {
	g, me, _ := p7Table(t)
	leash := apaPush(g, me.ID, me.ID, apaCreature("Leashling", p8Leashling, 3, 3))
	card := p7Hand(me, "Card", "Sorcery", "{R}")
	p7Activate(t, g, me, leash, 0, game.ActivateAbilityParams{TopIDs: []uuid.UUID{card}})
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
	g, me, _ := p7Table(t)
	pen := p8Push(g, me, "Penance", p8Penance, "Enchantment")
	card := p7Hand(me, "Card", "Sorcery", "{R}")
	if err := g.ActivateCatalogAbility(me.ID, pen, 0, game.ActivateAbilityParams{TopIDs: []uuid.UUID{card}}); err != nil {
		t.Fatalf("Penance: %v", err)
	}
	if top, _ := me.Library.Top(); top.InstanceID != card {
		t.Error("Penance's card is not on top of the library")
	}

	g, me, _ = p7Table(t)
	tac := apaPush(g, me.ID, me.ID, apaCreature("Seasoned Tactician", p8SeasonedTactic, 1, 3))
	lcLibrary(g, me, 3)
	if err := g.ActivateCatalogAbility(me.ID, tac, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrCantPayLibraryCost) {
		t.Fatalf("three cards: %v, want ErrCantPayLibraryCost", err)
	}
	lcLibrary(g, me, 4)
	if err := g.ActivateCatalogAbility(me.ID, tac, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Seasoned Tactician: %v", err)
	}
	if me.Library.Size() != 0 || len(onlyStackItemOf(t, g).Paid.Exiled) != 4 {
		t.Error("the top four cards were not exiled and recorded")
	}
}
