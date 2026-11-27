package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b656_cards_test.go — Invasion, batch 656 (S82, #5616). Each test
// plays the real catalog entry through the engine.

const (
	oldB656BarbarianOracle = "04576b4d-2977-4cdf-b4a5-b0b47995e962"
	oldB656SerpentOracle   = "23875679-5c08-44ff-90b3-13a1a758e5f7"
	oldB656SapphireOracle  = "500aff17-0d69-46ac-a4f9-2fc6db061ce7"
	oldB656MongooseOracle  = "588a5bac-dcad-410e-8878-0230500c2e3a"
	oldB656VodalianOracle  = "58e4f299-bb32-4727-8498-a364166adc38"
	oldB656LlanowarOracle  = "6b26951f-57be-4524-ae0e-a102e84e81ef"
	oldB656RaiderOracle    = "79dcc432-e8f5-4f2b-bff8-5ae964714666"
	oldB656GalinaOracle    = "9a484eb3-ae4a-465b-9dbc-f7a42a036e07"
	oldB656NishobaOracle   = "ce82aa79-bb50-47bc-84f5-eae30249c7a5"
	oldB656RagingOracle    = "d126bb08-be60-4046-91f1-65672ef42c63"
	oldB656SlayerOracle    = "da74cb57-f6de-4d05-a339-6412beb8df25"
	oldB656RubyOracle      = "ec6195a0-2701-40a4-882f-eb563668791a"
	oldB656ShivanOracle    = "fb3d438e-5e04-45a0-8d16-db63f0faff75"
	oldB656OptOracle       = "713332c1-5bd8-400f-bfff-c1ca0697a043"
	oldB656PyroclasmOracle = "e4bcd4ea-e7cd-4471-8f3b-18bb51d3d70c"
	oldB656CancelOracle    = "7d00fb28-ea6c-49a9-b4af-ffb38860a9a7"
	oldB656ForestOracle    = "b34bb2dc-c1af-4d77-b0b3-a0fb342a5fc6"
)

type oldB656Card struct {
	name, typeLine, oracle string
	power, toughness       int
	keywords               []string
}

var oldB656Cards = []oldB656Card{
	{"Yavimaya Barbarian", "Creature — Elf Barbarian", oldB656BarbarianOracle, 2, 2, []string{"protection from blue"}},
	{"Slinking Serpent", "Creature — Serpent", oldB656SerpentOracle, 2, 3, []string{"forestwalk"}},
	{"Sapphire Leech", "Creature — Leech", oldB656SapphireOracle, 2, 2, []string{"flying"}},
	{"Blurred Mongoose", "Creature — Mongoose", oldB656MongooseOracle, 2, 1, []string{"shroud"}},
	{"Vodalian Zombie", "Creature — Merfolk Zombie", oldB656VodalianOracle, 2, 2, []string{"protection from green"}},
	{"Llanowar Knight", "Creature — Elf Knight", oldB656LlanowarOracle, 2, 2, []string{"protection from black"}},
	{"Shoreline Raider", "Creature — Merfolk", oldB656RaiderOracle, 2, 2, []string{"protection from Kavu"}},
	{"Galina's Knight", "Creature — Merfolk Knight", oldB656GalinaOracle, 2, 2, []string{"protection from red"}},
	{"Sabertooth Nishoba", "Creature — Cat Beast Warrior", oldB656NishobaOracle, 5, 5, []string{"trample", "protection from blue", "protection from red"}},
	{"Raging Kavu", "Creature — Kavu", oldB656RagingOracle, 3, 1, []string{"flash", "haste"}},
	{"Phyrexian Slayer", "Creature — Phyrexian Minion", oldB656SlayerOracle, 2, 2, []string{"flying"}},
	{"Ruby Leech", "Creature — Leech", oldB656RubyOracle, 2, 2, []string{"first strike"}},
	{"Shivan Zombie", "Creature — Phyrexian Barbarian Zombie", oldB656ShivanOracle, 2, 2, []string{"protection from white"}},
}

func TestOldB656EveryCardIsRegisteredFullWithItsKeywords(t *testing.T) {
	for _, c := range oldB656Cards {
		spec, ok := effects.Lookup(c.oracle)
		if !ok {
			t.Errorf("%s is not registered", c.name)
			continue
		}
		if spec.Name != c.name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", c.name, spec.Name, spec.Completeness, spec.Caveats)
		}
		g := effectstest.NewCatalogGame(t)
		id := effectstest.PushCatalogPermanent(g, g.Seats[0].ID, c.name, c.typeLine, c.oracle, false)
		for _, kw := range c.keywords {
			if !effectstest.HasKeywordOnBattlefield(t, g, id, kw) {
				t.Errorf("%s does not have %q on the battlefield", c.name, kw)
			}
		}
	}
}

func oldB656Defender(g *game.Game, owner uuid.UUID, keywords []string, colors ...string) uuid.UUID {
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Blocker", TypeLine: "Creature — Kavu", Power: 2, Toughness: 2,
		Keywords: keywords, Colors: colors, Owner: owner, Controller: owner,
	})
	return id
}

func oldB656Attacker(t *testing.T, g *game.Game, me *game.Player, c oldB656Card) uuid.UUID {
	t.Helper()
	return effectstest.PushCatalogPermanent(g, me.ID, c.name, c.typeLine, c.oracle, false)
}

func oldB656BlockRefused(t *testing.T, g *game.Game, blocker, attacker uuid.UUID) bool {
	t.Helper()
	return g.DeclareBlocker(blocker, attacker) != nil
}

func TestOldB656ProtectionStopsBlockersOfThatColourAndType(t *testing.T) {
	for _, tc := range []struct {
		card     oldB656Card
		colors   []string
		typeLine string
		refused  bool
	}{
		{oldB656Cards[0], []string{"U"}, "Creature — Merfolk", true},
		{oldB656Cards[0], []string{"G"}, "Creature — Elf", false},
		{oldB656Cards[6], nil, "Creature — Kavu", true},
		{oldB656Cards[6], nil, "Creature — Elf", false},
		{oldB656Cards[7], []string{"R"}, "Creature — Goblin", true},
		{oldB656Cards[12], []string{"W"}, "Creature — Human", true},
		{oldB656Cards[8], []string{"U"}, "Creature — Merfolk", true},
		{oldB656Cards[8], []string{"R"}, "Creature — Goblin", true},
		{oldB656Cards[8], []string{"G"}, "Creature — Elf", false},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		atk := oldB656Attacker(t, g, me, tc.card)
		blocker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Blocker", TypeLine: tc.typeLine, Power: 2, Toughness: 2,
			Colors: tc.colors, Owner: opp.ID, Controller: opp.ID,
		})
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(atk, opp.ID); err != nil {
			t.Fatalf("%s: DeclareAttacker: %v", tc.card.name, err)
		}
		effectstest.LockInAttacks(t, g)
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		if got := oldB656BlockRefused(t, g, blocker, atk); got != tc.refused {
			t.Errorf("%s blocked by %v %s: refused %v, want %v", tc.card.name, tc.colors, tc.typeLine, got, tc.refused)
		}
	}
}

func TestOldB656SlinkingSerpentCannotBeBlockedWhileTheyControlAForest(t *testing.T) {
	for _, forest := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		atk := oldB656Attacker(t, g, me, oldB656Cards[1])
		if forest {
			effectstest.PushCatalogPermanent(g, opp.ID, "Forest", "Basic Land — Forest", oldB656ForestOracle, false)
		}
		blocker := oldB656Defender(g, opp.ID, nil)
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(atk, opp.ID); err != nil {
			t.Fatal(err)
		}
		effectstest.LockInAttacks(t, g)
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		if got := oldB656BlockRefused(t, g, blocker, atk); got != forest {
			t.Errorf("Forest %v: block refused %v, want %v", forest, got, forest)
		}
	}
}

func TestOldB656PhyrexianSlayerDestroysAWhiteBlockerEvenIfItRegenerates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	slayer := oldB656Attacker(t, g, me, oldB656Cards[10])
	white := oldB656Defender(g, opp.ID, []string{"flying"}, "W")
	other := oldB656Defender(g, opp.ID, []string{"flying"}, "G")
	effectstest.P1ShieldCreature(g, white)
	effectstest.Pr7bAttack(t, g, opp.ID, slayer)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{white: slayer, other: slayer})
	if g.Battlefield.Contains(white) {
		t.Error("the white blocker survived, regeneration shield or not")
	}
}

func TestOldB656PhyrexianSlayerLeavesANonWhiteBlockerAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	slayer := oldB656Attacker(t, g, me, oldB656Cards[10])
	green := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Green Bird", TypeLine: "Creature — Bird", Power: 0, Toughness: 9,
		Keywords: []string{"flying"}, Colors: []string{"G"}, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.Pr7bAttack(t, g, opp.ID, slayer)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{green: slayer})
	if !g.Battlefield.Contains(green) {
		t.Error("a green blocker was destroyed by the white-only trigger")
	}
}

func TestOldB656LeechesTaxYourOwnSpellsOfTheirColourOnly(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, cost, oracle, mana string
		opponent, refused                  bool
	}{
		{"Opt", "Instant", "{U}", oldB656OptOracle, "U", false, true},
		{"Opt", "Instant", "{U}", oldB656OptOracle, "UU", false, false},
		{"Pyroclasm", "Sorcery", "{1}{R}", oldB656PyroclasmOracle, "CR", false, true},
		{"Pyroclasm", "Sorcery", "{1}{R}", oldB656PyroclasmOracle, "CRR", false, false},
		// "You cast": an opponent's blue spell is not taxed.
		{"Opt", "Instant", "{U}", oldB656OptOracle, "U", true, false},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		effectstest.PushCatalogPermanent(g, me.ID, "Sapphire Leech", "Creature — Leech", oldB656SapphireOracle, false)
		effectstest.PushCatalogPermanent(g, me.ID, "Ruby Leech", "Creature — Leech", oldB656RubyOracle, false)
		effectstest.AdvanceToMain(t, g)
		caster := me
		if tc.opponent {
			caster = opp
		}
		effectstest.FloatForTest(g, caster, tc.mana)
		id := uuid.New()
		g.WithWriteLock(func() {
			caster.Hand.PushTop(game.Card{InstanceID: id, Name: tc.name, TypeLine: tc.typeLine, ManaCost: tc.cost, OracleID: tc.oracle, Owner: caster.ID, Controller: caster.ID})
		})
		err := g.CastSpell(caster.ID, id, game.CastSpellParams{Strict: true})
		if (err != nil) != tc.refused {
			t.Errorf("%s with %s (opponent %v): error %v, want refused %v", tc.name, tc.mana, tc.opponent, err, tc.refused)
		}
	}
}

func TestOldB656BlurredMongooseCannotBeCounteredAndHasShroud(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CG")
	mongoose := effectstest.CastFromHandForTest(t, g, me, "Blurred Mongoose", "Creature — Mongoose", "{1}{G}", oldB656MongooseOracle, game.CastSpellParams{Strict: true})
	effectstest.FloatForTest(g, opp, "CUU")
	effectstest.CastFromHandForTest(t, g, opp, "Cancel", "Instant", "{1}{U}{U}", oldB656CancelOracle, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mongoose}}})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mongoose) {
		t.Error("Cancel countered Blurred Mongoose")
	}
}
