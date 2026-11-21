package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// wishBoard seats the active player at main with a Sol Ring and the named
// basic lands pushed in order (the planner's tie-breaks follow it, so each
// test lists the lands the unwished plan would reach first).
func wishBoard(t *testing.T, lands ...string) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Sol Ring", "Artifact", effectstest.WishSolRingOracle, false)
	basics := map[string]string{
		"Plains": "Basic Land — Plains", "Island": "Basic Land — Island", "Swamp": "Basic Land — Swamp",
		"Mountain": "Basic Land — Mountain", "Forest": "Basic Land — Forest",
	}
	for _, name := range lands {
		if _, err := g.SpawnCards(me.ID, me.ID, game.ZoneBattlefield, game.Card{Name: name, TypeLine: basics[name]}, 1); err != nil {
			t.Fatalf("spawn %s: %v", name, err)
		}
	}
	return g, me, opp
}

func wishAutoCast(t *testing.T, g *game.Game, me *game.Player, name, typeLine, cost, oracle string, targets ...game.TargetRef) *game.StackItem {
	t.Helper()
	id := effectstest.CastFromHandForTest(t, g, me, name, typeLine, cost, oracle,
		game.CastSpellParams{Strict: true, AutoTap: true, FromZone: "hand", Targets: targets})
	item := g.StackMeta[id]
	if item == nil {
		t.Fatalf("%s is not on the stack", name)
	}
	return item
}

// The colour readers, through the printed cards. Each board is built so
// the unwished plan spends the Sol Ring (colourless first) or the wrong
// land, and the card's own declaration is what changes it.
func TestAutoPayGivesTheColourReadersWhatTheyRead(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, cost, oracle string
		lands                        []string
		wantSpent                    map[string]int
	}{
		{"Gruul Scrapper reads {R}", "Creature — Human Berserker", "{3}{G}", effectstest.WishGruulScrapper,
			[]string{"Forest", "Swamp", "Mountain"}, map[string]int{"R": 1}},
		{"Ribbons of Night reads {U}", "Sorcery", "{4}{B}", effectstest.WishRibbonsOracle,
			[]string{"Swamp", "Swamp", "Swamp", "Island"}, map[string]int{"U": 1}},
		{"Slaying Fire reads three {R}", "Instant", "{2}{R}", effectstest.SlayingFireOracle,
			[]string{"Swamp", "Mountain", "Mountain", "Mountain"}, map[string]int{"R": 3}},
		{"Moonhold reads {R} and {W}", "Instant", "{2}{R/W}", effectstest.WishMoonholdOracle,
			[]string{"Swamp", "Mountain", "Plains"}, map[string]int{"R": 1, "W": 1}},
		{"River's Grasp reads {U} and {B}", "Sorcery", "{3}{U/B}", effectstest.WishRiversGraspOracle,
			[]string{"Swamp", "Island", "Forest", "Forest"}, map[string]int{"U": 1, "B": 1}},
		{"Painful Truths converges", "Sorcery", "{2}{B}", effectstest.PainfulTruthsOracle,
			[]string{"Swamp", "Plains", "Island"}, map[string]int{"W": 1, "U": 1, "B": 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp := wishBoard(t, tc.lands...)
			// A creature for the spells that name one; the rest take the player.
			bear := effectstest.PushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
			targets := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
			switch tc.oracle {
			case effectstest.WishRibbonsOracle:
				targets = []game.TargetRef{{Kind: game.TargetCard, ID: bear}}
			case effectstest.WishRiversGraspOracle:
				targets = []game.TargetRef{{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetPlayer, ID: opp.ID}}
			}
			item := wishAutoCast(t, g, me, tc.name, tc.typeLine, tc.cost, tc.oracle, targets...)
			for colour, n := range tc.wantSpent {
				if got := item.Paid.SpentOfColor(colour); got < n {
					t.Errorf("{%s} spent = %d, want at least %d (%+v)", colour, got, n, item.Paid.Mana)
				}
			}
		})
	}
}

// The same boards with a spell that declares nothing keep the
// colourless-first instinct: the Sol Ring pays, and the readers' colours
// are not spent. This is the half that proves it is a WISH of the card and
// not a new blanket preference.
func TestAutoPayKeepsTheColourlessInstinctForASpellThatReadsNothing(t *testing.T) {
	g, me, _ := wishBoard(t, "Forest", "Swamp", "Mountain")
	item := wishAutoCast(t, g, me, "Plain Berserker", "Creature — Human Berserker", "{3}{G}", "")
	if got := item.Paid.SpentOfColor("R"); got != 0 {
		t.Errorf("{R} spent = %d on a spell that reads nothing, want 0 (%+v)", got, item.Paid.Mana)
	}
}

// A reader on a board that cannot make its colour is still cast: the
// wish is an ordering. Gruul Scrapper off two Forests and a Sol Ring
// pays {3}{G} and is an ordinary 3/2.
func TestAutoPayCastsAColourReaderThatHasNoSourceForItsColour(t *testing.T) {
	g, me, _ := wishBoard(t, "Forest", "Forest")
	item := wishAutoCast(t, g, me, "Gruul Scrapper", "Creature — Human Berserker", "{3}{G}", effectstest.WishGruulScrapper)
	if item.Paid.SpentOfColor("R") != 0 || item.Paid.ManaSpentCount() != 4 {
		t.Errorf("paid %+v, want four mana and no {R}", item.Paid.Mana)
	}
}

// --- source-side riders, through the printed cards -----------------------

// Path of Ancestry scries on a creature spell sharing a type with the
// commander; auto-pay taps it for that spell and leaves it for any other.
func TestAutoPayTapsPathOfAncestryForTheSpellItScriesOn(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		wantPath       bool
	}{
		{"an Elf shares the commander's type", "Creature — Elf Warrior", true},
		{"a Goblin does not", "Creature — Goblin", false},
		{"a sorcery is not a creature spell", "Sorcery", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.AdvanceToMain(t, g)
			g.WithWriteLock(func() {
				me.Command.PushTop(game.Card{
					InstanceID: uuid.New(), Name: "Lathril, Blade of the Elves",
					TypeLine: "Legendary Creature — Elf Noble", ManaCost: "{2}{B}{G}",
					Owner: me.ID, Controller: me.ID, IsCommander: true,
				})
			})
			// The Forest arrives first, so an unwished plan's tie-break
			// would reach for it; Path is the more flexible source.
			forests, err := g.SpawnCards(me.ID, me.ID, game.ZoneBattlefield, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}, 1)
			if err != nil {
				t.Fatalf("spawn Forest: %v", err)
			}
			path := effectstest.SeedPermanentWithOracle(g, me.ID, "Path of Ancestry", "Land", effectstest.PathOfAncestryOracle)

			wishAutoCast(t, g, me, "Test spell", tc.typeLine, "{G}", "")

			pathTapped, forestTapped := effectstest.PermanentTapped(g, path), effectstest.PermanentTapped(g, forests[0])
			if pathTapped != tc.wantPath || forestTapped == tc.wantPath {
				t.Errorf("Path tapped = %v, Forest tapped = %v; want Path %v", pathTapped, forestTapped, tc.wantPath)
			}
			if tc.wantPath {
				effectstest.PassPriorityAroundTable(t, g)
				scried := false
				for _, choice := range g.PendingChoices {
					if choice != nil && choice.Kind == game.PendingChoiceScry {
						scried = true
					}
				}
				if !scried {
					t.Error("Path was tapped for the spell but its scry never fired")
				}
			}
		})
	}
}

// Carnelian Orb of Dragonkind gives a Dragon spell haste; auto-pay taps it
// for a Dragon and leaves it for anything else.
func TestAutoPayTapsCarnelianOrbForADragon(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		wantOrb        bool
	}{
		{"a Dragon", "Creature — Dragon", true},
		{"a Goblin", "Creature — Goblin", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.AdvanceToMain(t, g)
			mountains, err := g.SpawnCards(me.ID, me.ID, game.ZoneBattlefield, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"}, 1)
			if err != nil {
				t.Fatalf("spawn Mountain: %v", err)
			}
			orb := effectstest.PushCatalogPermanent(g, me.ID, "Carnelian Orb of Dragonkind", "Artifact", effectstest.GmrCarnelianOrbOracle, false)

			wishAutoCast(t, g, me, "Test creature", tc.typeLine, "{R}", "")

			if got := effectstest.PermanentTapped(g, orb); got != tc.wantOrb {
				t.Errorf("Orb tapped = %v, want %v", got, tc.wantOrb)
			}
			if got := effectstest.PermanentTapped(g, mountains[0]); got == tc.wantOrb {
				t.Errorf("Mountain tapped = %v, want %v", got, !tc.wantOrb)
			}
		})
	}
}

// The thrift tiers still outrank a rider: Generator Servant eats itself
// for its mana, so a creature spell it would haste is paid from the
// Forest while a Forest is there to pay.
func TestAutoPayDoesNotSacrificeGeneratorServantJustForTheHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	servant := effectstest.PushCatalogPermanent(g, me.ID, "Generator Servant", "Creature — Elemental", effectstest.WishGeneratorServantID, false)
	if _, err := g.SpawnCards(me.ID, me.ID, game.ZoneBattlefield, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}, 2); err != nil {
		t.Fatalf("spawn Forests: %v", err)
	}

	wishAutoCast(t, g, me, "Test creature", "Creature — Bear", "{1}{G}", "")

	if !g.Battlefield.Contains(servant) {
		t.Error("auto-pay sacrificed the Generator Servant for a cost two Forests pay")
	}
}

// The new declaration is on the cards the issue names, and on no card
// that already pays the colour with a printed pip.
func TestWantsColorsIsDeclaredOnTheReadersOfAnUnrequiredColour(t *testing.T) {
	for oracle, want := range map[string][]string{
		effectstest.WishGruulScrapper:     {"R"},
		effectstest.WishRibbonsOracle:     {"U"},
		effectstest.SlayingFireOracle:     {"R"},
		effectstest.WishMoonholdOracle:    {"R", "W"},
		effectstest.WishRiversGraspOracle: {"U", "B"},
		effectstest.WishBorosFuryShield:   {"R"},
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Fatalf("%s is not in the catalog", oracle)
		}
		if len(spec.WantsColors) != len(want) {
			t.Errorf("%s WantsColors = %v, want %v", spec.Name, spec.WantsColors, want)
			continue
		}
		for i := range want {
			if spec.WantsColors[i] != want[i] {
				t.Errorf("%s WantsColors = %v, want %v", spec.Name, spec.WantsColors, want)
			}
		}
	}
}
