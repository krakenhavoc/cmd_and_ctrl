package testss82m

import (
	"strconv"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b580_test.go — Ravnica: City of Guilds, batch 580 (#5600). Every test
// plays the real catalog entry through the engine.

const (
	oldB580Shaman  = "5458689c-48e0-48b6-bda8-048fa4718734"
	oldB580Hound   = "762e01d8-c16a-44b8-8ba0-052d1a2f312f"
	oldB580Circu   = "c87d547b-00a4-4fdd-ade4-f3f032e5ea3b"
	oldB580Oaf     = "d82bde29-1576-4736-a755-df9c37765202"
	oldB580Zephyr  = "ff640343-6980-4668-a41f-0f01b0bbc0f2"
	oldB580Phalanx = "1a6ed882-ea6c-4dc2-8614-541e4a56b43c"
	oldB580Usher   = "3ad6e495-d0e6-4b07-8f91-fe057e173975"
	oldB580Grozoth = "8ecbc87e-a63e-4afc-b5eb-2fc3dcd685f5"
	oldB580Perplex = "904f9e0c-cb7f-4611-8bf7-2e408e8ab237"
	oldB580Bear    = "Creature — Bear"
)

func TestOldB580EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB580Shaman: "Stoneshaker Shaman", oldB580Hound: "Gate Hound", oldB580Circu: "Circu, Dimir Lobotomist",
		oldB580Oaf: "Indentured Oaf", oldB580Zephyr: "Zephyr Spirit", oldB580Phalanx: "Netherborn Phalanx",
		oldB580Usher: "Ethereal Usher", oldB580Grozoth: "Grozoth", oldB580Perplex: "Perplex",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB580StoneshakerShamanTakesAnUntappedLandFromTheActivePlayerAtEachEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Stoneshaker Shaman", "Creature — Human Shaman", oldB580Shaman, 1, 1)
	tapped := effectstest.PushLand(g, me.ID, "Tapped Island")
	effectstest.ArnTap(t, g, tapped)
	untapped := effectstest.PushLand(g, me.ID, "Untapped Island")
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, untapped)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(untapped) {
		t.Error("the untapped land was not sacrificed at its controller's end step")
	}
	if !g.Battlefield.Contains(tapped) {
		t.Error("the tapped land was sacrificed")
	}

	// The opponent's end step: they sacrifice, though the Shaman is not theirs.
	theirs := effectstest.PushLand(g, opp.ID, "Their Island")
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the opponent's untapped land was not sacrificed at their end step")
	}
}

func TestOldB580GateHoundGivesYourCreaturesVigilanceOnlyWhileEnchanted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	hound := effectstest.B12Push(g, me.ID, "Gate Hound", "Creature — Dog", oldB580Hound, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB580Bear, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", oldB580Bear, 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, bear, "vigilance") || effectstest.HasEffectiveKeyword(t, g, hound, "vigilance") {
		t.Fatal("vigilance without an Aura on the Hound")
	}
	aura := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Pacifism",
		TypeLine: "Enchantment — Aura", Owner: opp.ID, Controller: opp.ID})
	effectstest.AttachForTest(t, g, aura, hound)
	for name, id := range map[string]uuid.UUID{"the Hound": hound, "another creature": bear} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "vigilance") {
			t.Errorf("%s has no vigilance while the Hound is enchanted", name)
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "vigilance") {
		t.Error("an opponent's creature gained vigilance")
	}
	late := effectstest.B12Creature(g, me.ID, "Late", oldB580Bear, 1, 1)
	if !effectstest.HasEffectiveKeyword(t, g, late, "vigilance") {
		t.Error("a creature that arrived later has no vigilance")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })
	g.RunStateChecksForTest()
	if effectstest.HasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("vigilance stayed after the Aura left")
	}
}

func TestOldB580CircuExilesFromTargetLibrariesAndBarsThoseNamesToOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Circu, Dimir Lobotomist", "Legendary Creature — Human Wizard", oldB580Circu, 2, 3)
	zap := effectstest.SeedLibraryTop(opp, "Zap Beast", "Creature — Beast", "{1}{R}")
	cast := func(name string, colors ...string) {
		t.Helper()
		spell := effectstest.HandCardFull(me, name, "Instant", "{U}", "", colors)
		effectstest.FloatMana(t, g, me, "{U}{B}")
		if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("cast %s: %v", name, err)
		}
		for i := 0; i < 6; i++ {
			if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
				if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
					t.Fatalf("pick target: %v", err)
				}
				continue
			}
			if effectstest.PendingOfKind(g, game.PendingChoiceTriggerOrder) != nil {
				effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID)
				continue
			}
			effectstest.PassPriorityAroundTable(t, g)
		}
	}
	cast("Blue Trick", "U")
	if !effectstest.InExile(g, zap) {
		t.Fatal("the top card of the target's library was not exiled by a blue spell")
	}
	// A black spell exiles another; one that is both exiles two.
	second := effectstest.SeedLibraryTop(opp, "Second Card", "Creature — Beast", "{2}")
	cast("Black Trick", "B")
	if !effectstest.InExile(g, second) {
		t.Error("a black spell did not exile the top card")
	}
	third := effectstest.SeedLibraryTop(opp, "Third Card", "Creature — Beast", "{3}")
	fifth := effectstest.SeedLibraryTop(opp, "Fifth Card", "Creature — Beast", "{4}")
	cast("Both Trick", "U", "B")
	if !effectstest.InExile(g, fifth) || !effectstest.InExile(g, third) {
		t.Error("a blue-and-black spell did not trigger both abilities")
	}
	cast("Red Trick", "R")
	if len(oppExiledWithCount(g)) != 4 {
		t.Errorf("%d cards exiled with Circu, want 4 (a red spell triggers nothing)", len(oppExiledWithCount(g)))
	}

	// The opponent cannot cast a spell named like an exiled card; another is fine.
	effectstest.AdvanceToMainOf(t, g, 1)
	banned := effectstest.HandCardFull(opp, "Zap Beast", "Creature — Beast", "{1}{R}", "", []string{"R"})
	free := effectstest.HandCardFull(opp, "Free Beast", "Creature — Beast", "{1}{R}", "", []string{"R"})
	effectstest.FloatMana(t, g, opp, "{R}{R}{R}{R}")
	if err := g.CastSpell(opp.ID, banned, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("the opponent cast a spell with the name of a card exiled with Circu")
	}
	if err := g.CastSpell(opp.ID, free, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("the opponent could not cast an unrelated spell: %v", err)
	}
}

func oppExiledWithCount(g *game.Game) []game.Card {
	var out []game.Card
	for _, c := range g.Exile.Cards {
		if c.ExiledWith.ID != uuid.Nil {
			out = append(out, c)
		}
	}
	return out
}

func TestOldB580IndenturedOafPreventsItsDamageToRedCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	oaf := effectstest.B27Push(g, me.ID, "Indentured Oaf", "Creature — Ogre Warrior", oldB580Oaf, "{3}{R}", 4, 3, "R")
	red := effectstest.B27Push(g, opp.ID, "Red Bear", oldB580Bear, "", "{1}{R}", 5, 5, "R")
	green := effectstest.B27Push(g, opp.ID, "Green Bear", oldB580Bear, "", "{1}{G}", 5, 5, "G")
	other := effectstest.B27Push(g, opp.ID, "Other Red", oldB580Bear, "", "{1}{R}", 5, 5, "R")
	effectstest.B27Damage(g, oaf, red, 4)
	if got := effectstest.Pr6Marked(g, red); got != 0 {
		t.Errorf("the Oaf dealt %d to a red creature, want 0", got)
	}
	effectstest.B27Damage(g, oaf, green, 4)
	if got := effectstest.Pr6Marked(g, green); got != 4 {
		t.Errorf("the Oaf dealt %d to a green creature, want 4", got)
	}
	effectstest.B27Damage(g, other, red, 3)
	if got := effectstest.Pr6Marked(g, red); got != 3 {
		t.Errorf("another source dealt %d to a red creature, want 3", got)
	}
	life := effectstest.LifeOf(g, opp.ID)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(oaf, opp.ID, 4) })
	if got := effectstest.LifeOf(g, opp.ID); got != life-4 {
		t.Errorf("the Oaf dealt %d to a player, want 4", life-got)
	}
}

func TestOldB580ZephyrSpiritReturnsToHandWhenItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", oldB580Bear, 3, 3)
	spirit := effectstest.B12Push(g, opp.ID, "Zephyr Spirit", "Creature — Spirit", oldB580Zephyr, 0, 6)
	effectstest.Pr7bAttack(t, g, opp.ID, attacker)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{spirit: attacker})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(spirit) {
		t.Error("the Spirit stayed on the battlefield after blocking")
	}
	if !opp.Hand.Contains(spirit) {
		t.Error("the Spirit is not in its owner's hand")
	}
}

func TestOldB580NetherbornPhalanxDrainsEachOpponentPerCreatureTheyControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b, c := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 3; i++ {
		effectstest.B12Creature(g, a.ID, "A creature", oldB580Bear, 1, 1)
	}
	effectstest.B12Creature(g, b.ID, "B creature", oldB580Bear, 1, 1)
	effectstest.B12Creature(g, me.ID, "My creature", oldB580Bear, 1, 1)
	effectstest.B12Push(g, c.ID, "A land", "Basic Land — Island", "", 0, 0)
	lives := []int{effectstest.LifeOf(g, me.ID), effectstest.LifeOf(g, a.ID), effectstest.LifeOf(g, b.ID), effectstest.LifeOf(g, c.ID)}
	card := effectstest.HandCardFull(me, "Netherborn Phalanx", "Creature — Horror", "{5}{B}", oldB580Phalanx, []string{"B"})
	effectstest.FloatMana(t, g, me, "{B}{B}{B}{B}{B}{B}")
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	want := []int{lives[0], lives[1] - 3, lives[2] - 1, lives[3]}
	for i, p := range g.Seats {
		if got := effectstest.LifeOf(g, p.ID); got != want[i] {
			t.Errorf("seat %d life %d, want %d", i, got, want[i])
		}
	}
}

func oldB580Transmute(t *testing.T, oracle, name, typeLine, cost, mana string, idx int, mv int) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	card := effectstest.HandCardFull(me, name, typeLine, cost, oracle, nil)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Match", TypeLine: "Sorcery", ManaCost: manaCostOf(mv)},
		game.Card{Name: "Other", TypeLine: "Sorcery", ManaCost: manaCostOf(mv + 1)})
	effectstest.FloatMana(t, g, me, mana)
	if err := g.ActivateCatalogAbility(me.ID, card, idx, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("%s transmute: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, ids[0])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[0]) {
		t.Errorf("%s: the mana value %d card is not in hand", name, mv)
	}
	if me.Hand.Contains(card) {
		t.Errorf("%s: the card was not discarded", name)
	}
}

func manaCostOf(n int) string { return "{" + strconv.Itoa(n) + "}" }

func TestOldB580TransmuteFindsACardOfTheSameManaValue(t *testing.T) {
	oldB580Transmute(t, oldB580Phalanx, "Netherborn Phalanx", "Creature — Horror", "{5}{B}", "{B}{B}{B}", 0, 6)
	oldB580Transmute(t, oldB580Usher, "Ethereal Usher", "Creature — Spirit", "{5}{U}", "{U}{U}{U}", 1, 6)
	oldB580Transmute(t, oldB580Grozoth, "Grozoth", "Creature — Leviathan", "{6}{U}{U}{U}", "{U}{U}{U}", 1, 9)
	oldB580Transmute(t, oldB580Perplex, "Perplex", "Instant", "{1}{U}{B}", "{U}{U}{B}", 0, 3)
}

func TestOldB580EtherealUsherMakesATargetCreatureUnblockableThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	usher := effectstest.B12Push(g, me.ID, "Ethereal Usher", "Creature — Spirit", oldB580Usher, 2, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB580Bear, 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", oldB580Bear, 2, 2)
	effectstest.FloatMana(t, g, me, "{U}")
	if err := g.ActivateCatalogAbility(me.ID, usher, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertRestrictions(t, g, bear, game.CantBeBlocked)
	if c := effectstest.LayeredCard(t, g, usher); !c.Tapped {
		t.Error("the Usher did not tap")
	}
	_ = blocker
	effectstest.FloatMana(t, g, me, "{U}")
	if err := g.ActivateCatalogAbility(me.ID, usher, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Error("activated twice while tapped")
	}
}

func TestOldB580GrozothFetchesEveryManaValueNineCardAndCanLoseDefender(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Nine A", TypeLine: "Creature — Giant", ManaCost: "{9}"},
		game.Card{Name: "Nine B", TypeLine: "Creature — Giant", ManaCost: "{6}{U}{U}{U}"},
		game.Card{Name: "Eight", TypeLine: "Creature — Giant", ManaCost: "{8}"})
	card := effectstest.HandCardFull(me, "Grozoth", "Creature — Leviathan", "{6}{U}{U}{U}", oldB580Grozoth, []string{"U"})
	effectstest.FloatMana(t, g, me, "{U}{U}{U}{U}{U}{U}{U}{U}{U}")
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, ids[0], ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[0]) || !me.Hand.Contains(ids[1]) {
		t.Error("the two mana value 9 cards did not reach the hand")
	}
	if me.Hand.Contains(ids[2]) {
		t.Error("the mana value 8 card was fetched")
	}
	if !effectstest.HasEffectiveKeyword(t, g, card, "defender") {
		t.Fatal("Grozoth has no defender")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, card, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate {4}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasEffectiveKeyword(t, g, card, "defender") {
		t.Error("Grozoth kept defender after {4}")
	}
	effectstest.RunCleanup(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, card, "defender") {
		t.Error("defender did not come back after the turn")
	}
}

func oldB580PayPrompt(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		if c := g.PendingChoices[i]; c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

func TestOldB580PerplexCountersASpellUnlessItsControllerDiscardsTheirHand(t *testing.T) {
	for _, c := range []struct {
		name    string
		held    int
		discard bool
		stands  bool
	}{
		{"discard the whole hand", 2, true, true},
		{"decline", 2, false, false},
		{"empty hand pays nothing", 0, true, true},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		opp := g.Seats[1]
		effectstest.AdvanceToMainOf(t, g, 1)
		opp.Hand.Cards = nil
		var held []uuid.UUID
		for i := 0; i < c.held; i++ {
			held = append(held, effectstest.HandSpell(opp, "Held", "Sorcery", ""))
		}
		bear := effectstest.HandCardFull(opp, "Test Bear", oldB580Bear, "{1}{G}", "", []string{"G"})
		effectstest.FloatMana(t, g, opp, "{G}{G}")
		if err := g.CastSpell(opp.ID, bear, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("%s: cast the bear: %v", c.name, err)
		}
		perplex := effectstest.HandCardFull(me, "Perplex", "Instant", "{1}{U}{B}", oldB580Perplex, []string{"U", "B"})
		effectstest.FloatMana(t, g, me, "{U}{U}{B}")
		if err := g.CastSpell(me.ID, perplex, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err != nil {
			t.Fatalf("%s: cast Perplex: %v", c.name, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		prompt := oldB580PayPrompt(g, opp.ID)
		if prompt == nil {
			t.Fatalf("%s: the spell's controller was not asked", c.name)
		}
		if oldB580PayPrompt(g, me.ID) != nil {
			t.Errorf("%s: Perplex's controller was asked", c.name)
		}
		switch {
		case c.discard && c.held > 0:
			if err := g.ResolvePayUnlessWithCards(prompt.ID, opp.ID, true, held[:1]); err == nil {
				t.Fatalf("%s: discarding only part of the hand was accepted", c.name)
			}
			if err := g.ResolvePayUnlessWithCards(prompt.ID, opp.ID, true, held); err != nil {
				t.Fatalf("%s: discard the hand: %v", c.name, err)
			}
		case c.discard:
			effectstest.AnswerPayUnless(t, g, opp.ID, true)
		default:
			effectstest.AnswerPayUnless(t, g, opp.ID, false)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := g.Battlefield.Contains(bear); got != c.stands {
			t.Errorf("%s: the bear resolved = %v, want %v (zone %s)", c.name, got, c.stands, effectstest.B12ZoneOf(g, bear))
		}
		wantHand := c.held
		if c.discard {
			wantHand = 0
		}
		if got := opp.Hand.Size(); got != wantHand {
			t.Errorf("%s: hand %d, want %d", c.name, got, wantHand)
		}
	}
}
