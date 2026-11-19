package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b04_test.go — S82 GAP batch 4 (Mystery Booster Commander
// Edition creatures, #5948). Each test plays the real catalog entry.

const (
	gapB04JabsOracle     = "723959e1-8356-4fdb-99e2-818df317dc72"
	gapB04FlitwingOracle = "e8a9dddd-4527-46cd-8b7e-4f83c5b51e48"
	gapB04ChiefOracle    = "6b4544d8-0579-4621-8660-4adb3b2a90be"
	gapB04BalefangOracle = "195d1f2a-0d71-475b-9dd1-0c8efe64ad8f"
	gapB04GobyOracle     = "2d29c40a-4e13-4d99-bb5a-e991e385b8c2"
	gapB04JovenOracle    = "11db8545-eca6-43f5-b9e8-f302acef53a5"
	gapB04MatocOracle    = "bbe79ca2-8f05-4f1b-b5d3-84724d3eff3a"
)

// gapB04Answer answers the pending optional-replacement prompt, if any.
func gapB04Answer(t *testing.T, g *game.Game, apply bool) bool {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionalReplacement {
			if err := g.ResolveOptionalReplacement(c.ID, c.Chooser, apply); err != nil {
				t.Fatalf("ResolveOptionalReplacement: %v", err)
			}
			return true
		}
	}
	return false
}

func TestGapB04JabsAndFlitwingAreFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB04JabsOracle:     "Jabs, Mistress of Mockery",
		gapB04FlitwingOracle: "Flitwing, Lyev Detective",
		gapB04ChiefOracle:    "Chief Magistrate of Mercadia",
		gapB04BalefangOracle: "Balefang the Unslayable",
		gapB04GobyOracle:     "Grandmother Goby",
		gapB04JovenOracle:    "Joven and Chandler",
		gapB04MatocOracle:    "Matoc, Lavamancer",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v spec=%+v", name, ok, spec)
		}
	}
}

// --- Jabs ---------------------------------------------------------------

func TestGapB04JabsDoublesTriggersOfPermanentsYouControlButDontOwn(t *testing.T) {
	foods := func(withJabs bool, owner, controller int) int {
		g := newCatalogGame(t)
		me := g.Seats[0]
		if withJabs {
			pushCatalogPermanent(g, me.ID, "Jabs, Mistress of Mockery", "Legendary Creature — Human Rogue", gapB04JabsOracle, false)
		}
		pigeon := pushCatalogPermanent(g, g.Seats[owner].ID, "City Pigeon", "Creature — Bird", msc17PigeonOracle, false)
		if controller != owner {
			g.WithWriteLock(func() {
				g.GainControlForEffect(g.Seats[controller].ID, pigeon, g.Seats[controller].ID, g.UntilEndOfTurnDuration(), "test theft")
			})
		}
		b27Kill(g, pigeon)
		passPriorityAroundTable(t, g)
		return countNamed(g, "Food")
	}
	if got := foods(false, 1, 0); got != 1 {
		t.Fatalf("baseline: %d Food, want 1", got)
	}
	if got := foods(true, 1, 0); got != 2 {
		t.Errorf("a stolen permanent's trigger with Jabs made %d Food, want 2", got)
	}
	if got := foods(true, 0, 0); got != 1 {
		t.Errorf("a permanent its controller owns made %d Food with Jabs, want 1", got)
	}
	if got := foods(true, 1, 1); got != 1 {
		t.Errorf("an opponent's own permanent made %d Food with Jabs, want 1", got)
	}
}

// --- Flitwing -----------------------------------------------------------

func TestGapB04FlitwingMayMakeCluesInsteadOfTokens(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	pushCatalogPermanent(g, me, "Flitwing, Lyev Detective", "Legendary Creature — Faerie Detective", gapB04FlitwingOracle, false)
	makeTokens(t, g, me, RedGoblinToken(), 3)
	if !gapB04Answer(t, g, true) {
		t.Fatal("no optional-replacement prompt for Flitwing")
	}
	if clues, goblins := onBattlefieldNamed(g, "Clue"), onBattlefieldNamed(g, "Goblin"); clues != 3 || goblins != 0 {
		t.Errorf("accepted: %d Clues and %d Goblins, want 3 and 0", clues, goblins)
	}
}

func TestGapB04FlitwingDeclinedLeavesTheTokens(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	pushCatalogPermanent(g, me, "Flitwing, Lyev Detective", "Legendary Creature — Faerie Detective", gapB04FlitwingOracle, false)
	makeTokens(t, g, me, RedGoblinToken(), 2)
	if !gapB04Answer(t, g, false) {
		t.Fatal("no optional-replacement prompt for Flitwing")
	}
	if clues, goblins := onBattlefieldNamed(g, "Clue"), onBattlefieldNamed(g, "Goblin"); clues != 0 || goblins != 2 {
		t.Errorf("declined: %d Clues and %d Goblins, want 0 and 2", clues, goblins)
	}
}

func TestGapB04FlitwingIgnoresAnOpponentsTokens(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	pushCatalogPermanent(g, me, "Flitwing, Lyev Detective", "Legendary Creature — Faerie Detective", gapB04FlitwingOracle, false)
	makeTokens(t, g, opp, RedGoblinToken(), 2)
	if gapB04Answer(t, g, true) {
		t.Error("Flitwing asked about an opponent's tokens")
	}
	if goblins := onBattlefieldNamed(g, "Goblin"); goblins != 2 {
		t.Errorf("%d Goblins, want 2", goblins)
	}
}

// --- Chief Magistrate of Mercadia ---------------------------------------

func TestGapB04ChiefMagistrateCrownsYouWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	monCrown(t, g, g.Seats[1].ID)
	msc17Play(t, g, me, game.Card{Name: "Chief Magistrate of Mercadia", TypeLine: "Legendary Creature — Human Noble", OracleID: gapB04ChiefOracle, Power: 3, Toughness: 3})
	monSettle(t, g)
	if g.Monarch != me.ID {
		t.Errorf("monarch = %v, want the Chief's controller", g.Monarch)
	}
}

func TestGapB04ChiefMagistrateUpkeepCopiesEveryCreatureTokenForTheMonarch(t *testing.T) {
	for _, tc := range []struct {
		name         string
		monarch      int
		goblins, cat int
	}{{"monarch", 1, 2, 2}, {"not the monarch", 0, 1, 1}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[1]
			b12Push(g, me.ID, "Chief Magistrate of Mercadia", "Legendary Creature — Human Noble", gapB04ChiefOracle, 3, 3)
			makeTokens(t, g, me.ID, TokenCard("1/1 white Cat"), 1)
			// A noncreature token is not copied.
			makeTokens(t, g, me.ID, TreasureToken(), 1)
			monCrown(t, g, g.Seats[tc.monarch].ID)
			advanceToUpkeepOf(t, g, 1)
			monSettle(t, g)
			if n := monTokens(g, me.ID, "Goblin"); n != tc.goblins {
				t.Errorf("%d Goblins, want %d", n, tc.goblins)
			}
			if n := monTokens(g, me.ID, "Cat"); n != tc.cat {
				t.Errorf("%d Cats, want %d", n, tc.cat)
			}
			if n := monTokens(g, me.ID, "Treasure"); n != 1 {
				t.Errorf("%d Treasures, want 1 (not a creature)", n)
			}
			for _, c := range g.Battlefield.Cards {
				if c.Name == "Goblin" && (c.Power != 2 || c.Toughness != 1 || !effectiveAbilitiesContainCard(c, "haste")) {
					t.Errorf("Goblin is %d/%d %v, want a 2/1 with haste", c.Power, c.Toughness, c.Keywords)
				}
			}
		})
	}
}

func effectiveAbilitiesContainCard(c game.Card, kw string) bool {
	for _, k := range c.Keywords {
		if k == kw {
			return true
		}
	}
	return false
}

// --- Balefang the Unslayable --------------------------------------------

func TestGapB04BalefangGivesAnOpponentAGoadedTappedBaneslayerToken(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	msc17Play(t, g, me, game.Card{Name: "Balefang the Unslayable", TypeLine: "Legendary Creature — Demon Dragon", OracleID: gapB04BalefangOracle, Power: 5, Toughness: 5})
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
			t.Fatalf("pick: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	found := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Baneslayer Angel" {
			continue
		}
		found++
		if c.Controller != opp.ID || !IsToken(c) || !c.Tapped || c.Power != 5 || c.Toughness != 5 || !c.HasColor("W") {
			t.Errorf("token = %+v, want a tapped white 5/5 token under the opponent", c)
		}
		if !c.IsGoadedBy(me.ID) {
			t.Error("the token is not goaded by Balefang's controller")
		}
	}
	if found != 1 {
		t.Fatalf("%d Baneslayer Angel tokens, want 1", found)
	}
}

func TestGapB04BalefangStopsAllLifeGain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Balefang the Unslayable", "Legendary Creature — Demon Dragon", gapB04BalefangOracle, 5, 5)
	for _, p := range []*game.Player{me, opp} {
		before := p.Life
		g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, p.ID, 3) })
		if p.Life != before {
			t.Errorf("%s's life went %d to %d with Balefang out", p.ID, before, p.Life)
		}
	}
}

// --- Grandmother Goby ---------------------------------------------------

func TestGapB04GobyCountersYourNoncreatureSpellAndMakesAMerfolkPerMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Grandmother Goby", "Legendary Creature — Merfolk Wizard", gapB04GobyOracle, 2, 2)
	advanceToMain(t, g)
	id := uuid.New()
	if err := socB19CastWithPool(g, me, game.Card{InstanceID: id, Name: "Three", TypeLine: "Sorcery", ManaCost: "{2}{G}"}, "G", 3, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := monTokens(g, me.ID, "Merfolk"); n != 3 {
		t.Errorf("%d Merfolk, want one per mana spent (3)", n)
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the spell was not countered into its owner's graveyard")
	}
}

func TestGapB04GobyIgnoresCreatureSpellsAndOpponentsSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Grandmother Goby", "Legendary Creature — Merfolk Wizard", gapB04GobyOracle, 2, 2)
	advanceToMain(t, g)
	if err := socB19CastWithPool(g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, ManaCost: "{1}{G}"}, "G", 2, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := monTokens(g, me.ID, "Merfolk"); n != 0 {
		t.Errorf("%d Merfolk from a creature spell, want 0", n)
	}
	if countNamed(g, "Bear") != 1 {
		t.Error("the creature spell was countered")
	}
}

// --- Joven and Chandler -------------------------------------------------

func TestGapB04JovenGrowsByTheManaValueOfAnyArtifactThatDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	joven := b12Push(g, me.ID, "Joven and Chandler", "Legendary Creature — Human Rogue", gapB04JovenOracle, 3, 3)
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Big Rock", TypeLine: "Artifact", ManaCost: "{4}", Owner: opp.ID, Controller: opp.ID})
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, ManaCost: "{1}{G}", Owner: opp.ID, Controller: opp.ID})
	b27Kill(g, rock)
	passPriorityAroundTable(t, g)
	if p := effectivePowerWithCounters(t, g, joven); p != 7 {
		t.Errorf("power %d after a mana value 4 artifact died, want 3+4", p)
	}
	// A nonartifact creature dying adds nothing.
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bear" {
			b27Kill(g, c.InstanceID)
		}
	}
	passPriorityAroundTable(t, g)
	if p := effectivePowerWithCounters(t, g, joven); p != 7 {
		t.Errorf("power %d after a nonartifact died, want 7", p)
	}
}

func TestGapB04JovenStealsAnArtifactUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	joven := b12Push(g, me.ID, "Joven and Chandler", "Legendary Creature — Human Rogue", gapB04JovenOracle, 3, 3)
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", ManaCost: "{2}", Owner: opp.ID, Controller: opp.ID, Tapped: true})
	advanceToMain(t, g)
	fillPoolColored(me, "R", 3)
	if err := g.ActivateCatalogAbility(me.ID, joven, 0, game.ActivateAbilityParams{Targets: cardRefs(rock), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, ok := g.LookupCardForEffect(rock)
	if !ok || c.Controller != me.ID || c.Tapped {
		t.Errorf("artifact = %+v, want it under my control and untapped", c)
	}
}

// --- Matoc, Lavamancer --------------------------------------------------

func TestGapB04MatocExilesTheTopCardWhenASourceDealsExactlyTwo(t *testing.T) {
	for _, tc := range []struct {
		name   string
		damage int
		want   bool
	}{{"two", 2, true}, {"three", 3, false}, {"one", 1, false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			b12Push(g, me.ID, "Matoc, Lavamancer", "Legendary Creature — Human Barbarian Wizard", gapB04MatocOracle, 2, 3)
			src := b12Creature(g, me.ID, "Pinger", "Creature — Human", 1, 1)
			victim := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
			top := stackLibrary(me, "Top Card", "Sorcery", "{R}")
			b27Damage(g, src, victim, tc.damage)
			passPriorityAroundTable(t, g)
			if got := g.Exile.Contains(top); got != tc.want {
				t.Errorf("top card exiled = %v after %d damage, want %v", got, tc.damage, tc.want)
			}
			if tc.want {
				if perm := exiledPermission(g, top); perm.Player != me.ID {
					t.Errorf("permission holder = %v, want Matoc's controller", perm.Player)
				}
			}
		})
	}
}

func TestGapB04MatocIgnoresAnOpponentsDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Matoc, Lavamancer", "Legendary Creature — Human Barbarian Wizard", gapB04MatocOracle, 2, 3)
	src := b12Creature(g, opp.ID, "Their Pinger", "Creature — Human", 1, 1)
	victim := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	top := stackLibrary(me, "Top Card", "Sorcery", "{R}")
	b27Damage(g, src, victim, 2)
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(top) {
		t.Error("an opponent's 2 damage triggered Matoc")
	}
}
