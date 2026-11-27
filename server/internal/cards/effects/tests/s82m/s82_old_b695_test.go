package testss82m

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b695_test.go — the ten Mercadian Masques enchantments (#5619),
// each played through the real catalog and the engine.

const (
	mm5CorneredMarket  = "1784bbaa-67b1-4bd1-bf77-53da2bfbf7fc"
	mm5VernalEquinox   = "2241ef7d-3e9d-44ee-88ac-e8ad6fb2444e"
	mm5UphillBattle    = "349a18bf-5c61-4fe6-a5e5-5e7cad6b4ba6"
	mm5GamePreserve    = "409e0a43-0e85-47ea-9d40-de0860b12952"
	mm5Statecraft      = "854ef7c2-52e9-486a-aab6-7e3b3a6185ea"
	mm5Putrefaction    = "8c4545f8-10b6-4a62-9bce-8c9ffa8b8471"
	mm5HighSeas        = "9100f213-f4c1-4b15-94f2-e280579894c0"
	mm5Intimidation    = "a345d7d1-5d30-49a0-b139-a46f9419c50f"
	mm5Squeeze         = "e18d8326-c3c3-43c2-ac13-6861c0ef3888"
	mm5MagistratesVeto = "f50f8cac-35bf-4d4d-88c0-1c0c581f862c"
)

func TestMm5B695Registered(t *testing.T) {
	for oracle, name := range map[string]string{
		mm5CorneredMarket: "Cornered Market", mm5VernalEquinox: "Vernal Equinox",
		mm5UphillBattle: "Uphill Battle", mm5GamePreserve: "Game Preserve",
		mm5Statecraft: "Statecraft", mm5Putrefaction: "Putrefaction",
		mm5HighSeas: "High Seas", mm5Intimidation: "Intimidation",
		mm5Squeeze: "Squeeze", mm5MagistratesVeto: "Magistrate's Veto",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

func TestMm5CorneredMarketBarsSpellsAndNonbasicLandsNamedLikeAPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushCatalogPermanent(g, me.ID, "Cornered Market", "Enchantment", mm5CorneredMarket, false)
	effectstest.PushCatalogPermanent(g, opp.ID, "Sol Ring", "Artifact", "", false)
	effectstest.B12Push(g, opp.ID, "Steam Vents", "Land — Island Mountain", "", 0, 0)
	effectstest.B12Push(g, opp.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	effectstest.B12Push(g, opp.ID, "Goblin", "Token Creature — Goblin", "", 1, 1)
	effectstest.AdvanceToMain(t, g)

	for _, p := range []*game.Player{me, opp} {
		_, err := effectstest.CastFromHandErr(g, p, "Sol Ring", "Instant", "", "", game.CastSpellParams{Strict: true})
		if err == nil || !strings.Contains(err.Error(), "same name") {
			t.Errorf("%s cast a spell named like a permanent: %v", p.Name, err)
		}
	}
	if _, err := effectstest.CastFromHandErr(g, me, "Goblin", "Instant", "", "", game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("a spell named like a TOKEN was refused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, err := effectstest.CastFromHandErr(g, me, "Mind Stone", "Instant", "", "", game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("an unrelated spell was refused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	vents := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: vents, Name: "Steam Vents", TypeLine: "Land — Island Mountain", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, vents, game.CastSpellParams{Strict: true}); err == nil || !strings.Contains(err.Error(), "same name") {
		t.Errorf("a nonbasic land named like a permanent was played: %v", err)
	}
	forest := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: forest, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, forest, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("a basic land was refused: %v", err)
	}
}

func TestMm5CorneredMarketLiftsWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	market := effectstest.PushCatalogPermanent(g, me.ID, "Cornered Market", "Enchantment", mm5CorneredMarket, false)
	effectstest.PushCatalogPermanent(g, me.ID, "Sol Ring", "Artifact", "", false)
	effectstest.AdvanceToMain(t, g)
	if _, err := effectstest.CastFromHandErr(g, me, "Sol Ring", "Instant", "", "", game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("the bar was not in force")
	}
	effectstest.B14Kill(g, market)
	if _, err := effectstest.CastFromHandErr(g, me, "Sol Ring", "Instant", "", "", game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("the bar outlived the enchantment: %v", err)
	}
}

func TestMm5VernalEquinoxGivesEveryPlayerFlashForCreaturesAndEnchantments(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.AdvanceToMainOf(t, g, seat)
	effectstest.AdvanceToStepInTurn(t, g, game.StepBeginCombat)
	cast := func(name, typeLine string) error {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: me.ID, Controller: me.ID})
		return g.CastSpell(me.ID, id, game.CastSpellParams{})
	}
	if err := cast("Grizzly Bears", "Creature — Bear"); err == nil {
		t.Fatal("cast a creature in combat with no flash")
	}
	// The Equinox belongs to another player: "any player" still gets it.
	effectstest.PushCatalogPermanent(g, g.Seats[(seat+1)%len(g.Seats)].ID, "Vernal Equinox", "Enchantment", mm5VernalEquinox, false)
	if err := cast("Grizzly Bears", "Creature — Bear"); err != nil {
		t.Errorf("a creature was refused at flash speed: %v", err)
	}
	if err := cast("Plain Enchantment", "Enchantment"); err != nil {
		t.Errorf("an enchantment was refused at flash speed: %v", err)
	}
	if err := cast("Plain Sorcery", "Sorcery"); err == nil {
		t.Error("a sorcery was given flash")
	}
	if err := cast("Plain Artifact", "Artifact"); err == nil {
		t.Error("an artifact was given flash")
	}
}

func TestMm5UphillBattleTapsOnlyCreaturesCastByOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushCatalogPermanent(g, opp.ID, "Uphill Battle", "Enchantment", mm5UphillBattle, false)
	bear := effectstest.B12PlayFromHand(t, g, "Grizzly Bears", "Creature — Bear", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, bear) {
		t.Error("an opponent's creature entered untapped")
	}
	// Its own controller's creatures are untouched.
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g2, me2.ID, "Uphill Battle", "Enchantment", mm5UphillBattle, false)
	own := effectstest.B12PlayFromHand(t, g2, "Grizzly Bears", "Creature — Bear", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g2)
	if effectstest.B16Tapped(t, g2, own) {
		t.Error("the controller's own creature entered tapped")
	}
	_ = me
}

func TestMm5StatecraftPreventsCombatDamageToAndByYourCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushCatalogPermanent(g, me.ID, "Statecraft", "Enchantment", mm5Statecraft, false)
	mine := effectstest.B12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, mine)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(theirs, mine); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.FinishBlockDeclarations(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	g.RunStateChecksForTest()
	if effectstest.DamageMarkedOn(g, mine) != 0 {
		t.Errorf("my creature was dealt %d combat damage", effectstest.DamageMarkedOn(g, mine))
	}
	if effectstest.DamageMarkedOn(g, theirs) != 0 {
		t.Errorf("their creature was dealt %d combat damage by my creature", effectstest.DamageMarkedOn(g, theirs))
	}
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), mine, 3) })
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(mine) {
		t.Error("noncombat damage was prevented")
	}
}

func TestMm5StatecraftLetsAnUnblockedCreatureDealNothingToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushCatalogPermanent(g, me.ID, "Statecraft", "Enchantment", mm5Statecraft, false)
	mine := effectstest.B12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 3, 3)
	before := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, mine)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	g.RunStateChecksForTest()
	if opp.Life != before {
		t.Errorf("opponent life %d, want %d: my creature's combat damage is prevented", opp.Life, before)
	}
}

func TestMm5PutrefactionMakesTheCasterDiscardForGreenOrWhite(t *testing.T) {
	for _, tc := range []struct {
		name    string
		colors  []string
		discard bool
	}{
		{"green", []string{"G"}, true},
		{"white", []string{"W"}, true},
		{"green and blue", []string{"G", "U"}, true},
		{"blue", []string{"U"}, false},
		{"colorless", nil, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.PushCatalogPermanent(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, "Putrefaction", "Enchantment", mm5Putrefaction, false)
			effectstest.AdvanceToMain(t, g)
			held := uuid.New()
			me.Hand.PushTop(game.Card{InstanceID: held, Name: "Held Card", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
			spell := game.Card{Name: "Colored Thing", TypeLine: "Instant", ManaCost: "{1}", Colors: tc.colors}
			if err := effectstest.SocB19CastWithPool(g, me, spell, "C", 1, game.CastSpellParams{}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if !tc.discard {
				if effectstest.DiscardChoiceFor(g, me.ID) != nil || !me.Hand.Contains(held) {
					t.Fatal("a spell that is neither green nor white made the caster discard")
				}
				return
			}
			if effectstest.DiscardChoiceFor(g, me.ID) == nil {
				t.Fatal("the caster was not asked to discard")
			}
			effectstest.AnswerDiscard(t, g, me.ID, held)
			if me.Hand.Contains(held) {
				t.Error("the chosen card was not discarded")
			}
		})
	}
}

// mm5Price is the mana value of a spell as the cost modifiers price it.
func mm5Price(t *testing.T, g *game.Game, seat *game.Player, name, typeLine, cost string, colors ...string) int {
	t.Helper()
	base, err := game.ParseCost(cost)
	if err != nil {
		t.Fatalf("ParseCost(%q): %v", cost, err)
	}
	out, err := g.ApplyCostModifiers(base, game.CostQuery{
		Card: game.Card{
			InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: cost,
			Colors: colors, Owner: seat.ID, Controller: seat.ID,
		},
		Controller: seat.ID,
		FromZone:   game.ZoneHand,
	})
	if err != nil {
		t.Fatalf("ApplyCostModifiers: %v", err)
	}
	return out.ManaValue()
}

func TestMm5HighSeasTaxesRedAndGreenCreatureSpellsByOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "High Seas", "Enchantment", mm5HighSeas, false)
	for _, p := range []*game.Player{me, opp} {
		if got := mm5Price(t, g, p, "Red Bear", "Creature — Bear", "{1}{R}", "R"); got != 3 {
			t.Errorf("%s red creature: %d, want 3", p.Name, got)
		}
		if got := mm5Price(t, g, p, "Green Bear", "Creature — Bear", "{1}{G}", "G"); got != 3 {
			t.Errorf("%s green creature: %d, want 3", p.Name, got)
		}
	}
	if got := mm5Price(t, g, me, "Gruul Bear", "Creature — Bear", "{R}{G}", "R", "G"); got != 3 {
		t.Errorf("red-green creature: %d, want 3 (one more, not two)", got)
	}
	if got := mm5Price(t, g, me, "Blue Bear", "Creature — Bear", "{1}{U}", "U"); got != 2 {
		t.Errorf("blue creature: %d, want 2", got)
	}
	if got := mm5Price(t, g, me, "Red Bolt", "Instant", "{R}", "R"); got != 1 {
		t.Errorf("red instant: %d, want 1", got)
	}
	if got := mm5Price(t, g, me, "Green Aura", "Enchantment — Aura", "{G}", "G"); got != 1 {
		t.Errorf("green noncreature: %d, want 1", got)
	}
}

func TestMm5SqueezeTaxesSorceriesByThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Squeeze", "Enchantment", mm5Squeeze, false)
	for _, p := range []*game.Player{me, opp} {
		if got := mm5Price(t, g, p, "Cultivate", "Sorcery", "{2}{G}", "G"); got != 6 {
			t.Errorf("%s sorcery: %d, want 6", p.Name, got)
		}
	}
	if got := mm5Price(t, g, me, "Bolt", "Instant", "{R}", "R"); got != 1 {
		t.Errorf("instant: %d, want 1", got)
	}
	if got := mm5Price(t, g, me, "Bears", "Creature — Bear", "{1}{G}", "G"); got != 2 {
		t.Errorf("creature: %d, want 2", got)
	}
}

func TestMm5IntimidationGivesFearToYourCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Intimidation", "Enchantment", mm5Intimidation, false)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AssertKeywords(t, g, mine, "fear")
	c, _ := effectstest.BattlefieldCard(g, theirs)
	if game.HasKeyword(&c, "fear") {
		t.Error("an opponent's creature has fear")
	}
	green := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Green Wall", TypeLine: "Creature — Wall", Colors: []string{"G"}, Power: 0, Toughness: 4, Owner: opp.ID, Controller: opp.ID})
	black := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Black Wall", TypeLine: "Creature — Wall", Colors: []string{"B"}, Power: 0, Toughness: 4, Owner: opp.ID, Controller: opp.ID})
	artifact := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Golem", TypeLine: "Artifact Creature — Golem", Power: 1, Toughness: 4, Owner: opp.ID, Controller: opp.ID})
	effectstest.DeclareAttack(t, g, opp.ID, mine)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.BrRefusal(t, g.DeclareBlocker(green, mine), game.BlockReasonFear)
	if err := g.DeclareBlocker(black, mine); err != nil {
		t.Errorf("a black creature could not block: %v", err)
	}
	if err := g.DeclareBlocker(artifact, mine); err != nil {
		t.Errorf("an artifact creature could not block: %v", err)
	}
}

func TestMm5MagistratesVetoStopsWhiteAndBlueCreaturesBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Magistrate's Veto", "Enchantment", mm5MagistratesVeto, false)
	atk := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Ogre", 3, 3)
	mk := func(name string, colors ...string) uuid.UUID {
		return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Wall", Colors: colors, Power: 0, Toughness: 4, Owner: opp.ID, Controller: opp.ID})
	}
	white, blue, both, green, colorless := mk("White", "W"), mk("Blue", "U"), mk("Both", "W", "U"), mk("Green", "G"), mk("Colorless")
	effectstest.DeclareAttack(t, g, opp.ID, atk)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	for name, id := range map[string]uuid.UUID{"white": white, "blue": blue, "white-blue": both} {
		if err := g.DeclareBlocker(id, atk); err == nil {
			t.Errorf("a %s creature blocked", name)
		}
	}
	for name, id := range map[string]uuid.UUID{"green": green, "colorless": colorless} {
		if err := g.DeclareBlocker(id, atk); err != nil {
			t.Errorf("a %s creature could not block: %v", name, err)
		}
	}
}

func TestMm5GamePreservePutsEveryRevealedCreatureOntoTheBattlefield(t *testing.T) {
	for _, tc := range []struct {
		name      string
		spoiler   bool
		wantEnter bool
	}{
		{"all creatures", false, true},
		{"one noncreature", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			seat := g.Turn.ActiveSeat
			me := g.Seats[seat]
			effectstest.PushCatalogPermanent(g, me.ID, "Game Preserve", "Enchantment", mm5GamePreserve, false)
			// Park the game on the last seat's turn so no draw step falls between
			// the seeding and Game Preserve's upkeep.
			effectstest.AdvanceToMainOf(t, g, len(g.Seats)-1)
			tops := make([]uuid.UUID, len(g.Seats))
			for i, p := range g.Seats {
				c := game.Card{InstanceID: uuid.New(), Name: "Top Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID}
				if tc.spoiler && i == len(g.Seats)-1 {
					c.Name, c.TypeLine = "Top Spell", "Sorcery"
				}
				tops[i] = c.InstanceID
				p.Library.PushTop(c)
			}
			effectstest.AdvanceToUpkeepOf(t, g, seat)
			effectstest.PassPriorityAroundTable(t, g)
			for i, id := range tops {
				if got := g.Battlefield.Contains(id); got != tc.wantEnter {
					t.Errorf("seat %d's top card on the battlefield = %v, want %v", i, got, tc.wantEnter)
				}
				if tc.wantEnter {
					c, ok := effectstest.BattlefieldCard(g, id)
					if !ok || c.Controller != g.Seats[i].ID {
						t.Errorf("seat %d's creature is not under its owner's control", i)
					}
				}
			}
		})
	}
}
