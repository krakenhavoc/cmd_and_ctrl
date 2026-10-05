package effects

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// devoid_test.go — #2152, CR 702.114a, through real casts. The engine
// half is game/devoid_test.go; this file is the catalog half: the
// devoid cards declare the keyword, and what the table sees is
// colourless — in hand, on the stack, on the battlefield and in the
// graveyard — while a layer-5 colour effect and a copy behave as the
// rules say.

// assertCardColourless fails unless the live card answers colourless
// through the reads card files use.
func assertCardColourless(t *testing.T, g *game.Game, where string, id uuid.UUID) {
	t.Helper()
	c := findCardAnywhere(t, g, id)
	if !c.IsColorless() || len(c.EffectiveColors()) != 0 || len(c.Effective().Colors) != 0 {
		t.Errorf("%s: %s colours = %v, want none", where, c.Name, c.EffectiveColors())
	}
	if !Colorless()(g, uuid.Nil, c) || OfColor("U")(g, uuid.Nil, c) || OfColor("R")(g, uuid.Nil, c) {
		t.Errorf("%s: %s fails the colour predicates a target clause uses", where, c.Name)
	}
}

// TestUginsBindingIsColourlessInEveryZone is the issue's own card: it
// used to be blue everywhere, which is what its caveat said.
func TestUginsBindingIsColourlessInEveryZone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushCatalogPermanent(g, opp.ID, "Grizzly Bears", "Creature — Bear", "", false)
	// The deck importer's shape: Scryfall's empty colour list, the blue
	// pip in the cost, and no keyword on the card itself — the catalog
	// entry is what says devoid here.
	binding := putInHand(me, game.Card{Name: "Ugin's Binding", TypeLine: "Instant",
		OracleID: uginsBindingOracle, ManaCost: "{2}{U}", Colors: []string{}})
	assertCardColourless(t, g, "hand", binding)

	toMain(t, g)
	if err := g.CastSpell(me.ID, binding, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if !g.Stack.Contains(binding) {
		t.Fatal("Ugin's Binding should be on the stack")
	}
	assertCardColourless(t, g, "stack", binding)

	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(binding) {
		t.Fatal("Ugin's Binding should be in the graveyard")
	}
	assertCardColourless(t, g, "graveyard", binding)
	if !opp.Hand.Contains(bear) {
		t.Error("the spell still resolves: the bear goes home")
	}
}

// TestADevoidSpellIsAColorlessSpellForUginsBinding is devoid seen by a
// payoff: Serpentine Spike costs {5}{R}{R} — mana value 7 — and is a
// colorless spell, so it wakes Ugin's Binding in the graveyard. Before
// #2152 it was red and nothing happened. The second case is a devoid
// card with no catalog entry, colourless through the keyword the deck
// importer stamps.
func TestADevoidSpellIsAColorlessSpellForUginsBinding(t *testing.T) {
	for _, tc := range []struct {
		name     string
		card     game.Card
		creature int // target creatures the spell needs
	}{
		{"catalog road: Serpentine Spike", game.Card{Name: "Serpentine Spike", TypeLine: "Sorcery",
			OracleID: p1g3SpikeOracle, ManaCost: "{5}{R}{R}", Colors: []string{}}, 3},
		{"import road: an uncatalogued devoid spell", game.Card{Name: "Fixture Devastation", TypeLine: "Sorcery",
			ManaCost: "{6}{R}", Colors: []string{}, Keywords: []string{game.KeywordDevoid}}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			pushGraveyardBinding(me)
			var targets []game.TargetRef
			for i := 0; i < tc.creature; i++ {
				id := pushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
				targets = append(targets, game.TargetRef{Kind: game.TargetCard, ID: id})
			}
			spell := putInHand(me, tc.card)
			toMain(t, g)
			if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Targets: targets}); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			assertCardColourless(t, g, "stack", spell)
			passPriorityUntilChoice(t, g)
			if latestChoiceOfKind(g, game.PendingChoiceConfirm) == nil {
				t.Error("a devoid spell with mana value 7 is a colorless spell: Ugin's Binding should offer its exile")
			}
		})
	}
}

// TestCeruleanWispsMakesADevoidCreatureBlue: devoid is applied first
// in layer 5 (CR 613.3), and "becomes blue" applies after it — so the
// Drone is blue until end of turn.
func TestCeruleanWispsMakesADevoidCreatureBlue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	drone := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Propagator Drone",
		TypeLine: "Creature — Eldrazi Drone", OracleID: oraclePropagatorDrone, ManaCost: "{1}{G}",
		Colors: []string{}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if got := battlefieldCardCopy(t, g, drone).EffectiveColors(); len(got) != 0 {
		t.Fatalf("Propagator Drone on the battlefield: colours = %v, want none", got)
	}
	castCatalogSpell(t, g, "Cerulean Wisps", "Instant", b43CeruleanWispsOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: drone}})
	passPriorityAroundTable(t, g)
	if got := battlefieldCardCopy(t, g, drone).EffectiveColors(); !reflect.DeepEqual(got, []string{"U"}) {
		t.Errorf("after Cerulean Wisps: colours = %v, want [U]", got)
	}
}

// TestTokenCopyOfADevoidCreatureIsColourless: devoid is a copiable
// value (CR 707.2) on both roads — the oracle ID a token copy takes
// carries the catalog's keyword, and Keywords carries the importer's.
// The Scarab God's "except it's a 4/4 black Zombie" provides a colour,
// so the colour-defining ability is not copied (CR 707.9d): black.
func TestTokenCopyOfADevoidCreatureIsColourless(t *testing.T) {
	for _, tc := range []struct {
		name   string
		card   game.Card
		except func(*game.Card)
		want   []string
	}{
		{"catalog road", game.Card{Name: "Propagator Drone", TypeLine: "Creature — Eldrazi Drone",
			OracleID: oraclePropagatorDrone, ManaCost: "{1}{G}", Colors: []string{}, Power: 2, Toughness: 2}, nil, nil},
		{"import road", game.Card{Name: "Fixture Drone", TypeLine: "Creature — Eldrazi Drone",
			ManaCost: "{2}{B}", Colors: []string{}, Keywords: []string{game.KeywordDevoid}, Power: 2, Toughness: 2}, nil, nil},
		{"except it's black", game.Card{Name: "Fixture Drone", TypeLine: "Creature — Eldrazi Drone",
			ManaCost: "{2}{U}", Colors: []string{}, Keywords: []string{game.KeywordDevoid}, Power: 2, Toughness: 2},
			scarabGodZombieException, []string{"B"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			src := tc.card
			src.InstanceID, src.Owner, src.Controller = uuid.New(), me.ID, me.ID
			me.Graveyard.PushTop(src)
			g.WithWriteLock(func() {
				if err := (CreateTokenCopy{Controller: me.ID, Copy: src.InstanceID, N: 1, Except: tc.except}).Apply(NewContext(g, nil)); err != nil {
					t.Fatalf("CreateTokenCopy.Apply: %v", err)
				}
			})
			token := findBattlefieldByName(g, src.Name)
			if token == uuid.Nil {
				t.Fatal("no token copy was created")
			}
			got := battlefieldCardCopy(t, g, token).EffectiveColors()
			if !reflect.DeepEqual(got, tc.want) && !(len(got) == 0 && len(tc.want) == 0) {
				t.Errorf("token copy colours = %v, want %v", got, tc.want)
			}
		})
	}
}
