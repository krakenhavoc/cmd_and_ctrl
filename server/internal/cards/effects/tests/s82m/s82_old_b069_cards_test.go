package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kaldheim batch 069 (S82 OLD, #5505).

const (
	oldB069Harald    = "0089b07c-7f47-4a9a-9fb7-cc38ec9d7bcd"
	oldB069Cyclone   = "2088e791-9ce5-4e97-a4ce-f51adaa3f42b"
	oldB069Koll      = "344dc8b0-302a-4721-99b6-5a6626a349c1"
	oldB069Thornmant = "39f17e13-018c-4a00-ab25-1ac5a904de63"
	oldB069Sigrid    = "84d37104-2681-42c1-9d19-91a2ebe6a435"
	oldB069Rally     = "eeb07c8a-21d1-43fa-a5f3-ee3fe328a671"
)

func TestOldB069CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB069Harald: "Harald, King of Skemfar", oldB069Cyclone: "Cyclone Summoner", oldB069Koll: "Koll, the Forgemaster",
		oldB069Thornmant: "Thornmantle Striker", oldB069Sigrid: "Sigrid, God-Favored", oldB069Rally: "Rally the Ranks",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB069Attach(g *game.Game, owner, host uuid.UUID, typeLine string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Attachment", TypeLine: typeLine,
		Owner: owner, Controller: owner, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host},
	})
}

// --- Rally the Ranks -----------------------------------------------------

func TestOldB069RallyTheRanksPumpsYourCreaturesOfTheChosenType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	elf := effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf Warrior", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 2, 2)
	effectstest.PushNamedTribePermanent(t, g, me.ID, "Rally the Ranks", "Enchantment", oldB069Rally, "Elf")
	if got := effectstest.EffectivePower(t, g, elf); got != 3 {
		t.Errorf("Elf power %d, want 3", got)
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("Bear power %d, want 2", got)
	}
	if got := effectstest.EffectivePower(t, g, theirs); got != 2 {
		t.Errorf("an opponent's Elf has power %d, want 2", got)
	}
}

// --- Harald, King of Skemfar -------------------------------------------

func TestOldB069HaraldTakesAnElfWarriorOrTyvarAndBottomsTheRest(t *testing.T) {
	for _, c := range []struct{ name, typeLine string }{
		{"Elf Card", "Creature — Elf Druid"},
		{"Warrior Card", "Creature — Human Warrior"},
		{"Tyvar Card", "Legendary Planeswalker — Tyvar"},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			deep := effectstest.SeedLibraryTop(me, "Deep Elf", "Creature — Elf", "{G}")
			hit := effectstest.SeedLibraryTop(me, c.name, c.typeLine, "{G}")
			var others []uuid.UUID
			for _, n := range []string{"Bear A", "Bear B", "Forest", "Bear C"} {
				tl := "Creature — Bear"
				if n == "Forest" {
					tl = "Basic Land — Forest"
				}
				others = append(others, effectstest.SeedLibraryTop(me, n, tl, ""))
			}
			hand := me.Hand.Size()
			effectstest.CastCatalogSpell(t, g, "Harald, King of Skemfar", "Legendary Creature — Elf Warrior", oldB069Harald, nil)
			effectstest.PassPriorityAroundTable(t, g)
			p := effectstest.LatestChooseCardsFor(g, me.ID)
			if p == nil {
				t.Fatal("no choose_cards prompt")
			}
			if len(p.ChooseCards) != 1 || p.ChooseCards[0] != hit {
				t.Fatalf("offered %v, want only the qualifying card among the top five", p.ChooseCards)
			}
			if err := g.ResolveChooseCards(p.ID, me.ID, []uuid.UUID{hit}); err != nil {
				t.Fatalf("ResolveChooseCards: %v", err)
			}
			if !me.Hand.Contains(hit) || me.Hand.Size() != hand+1 {
				t.Error("the card did not reach the hand")
			}
			n := len(me.Library.Cards)
			if me.Library.Cards[n-1].InstanceID != deep {
				t.Errorf("the card that was sixth is not on top: %s", me.Library.Cards[n-1].Name)
			}
			bottom := map[uuid.UUID]bool{}
			for _, c := range me.Library.Cards[:4] {
				bottom[c.InstanceID] = true
			}
			for _, id := range others {
				if !bottom[id] {
					t.Error("a card that was not taken did not go to the bottom")
				}
			}
		})
	}
}

func TestOldB069HaraldOffersNothingWhenNoCardQualifies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 4; i++ {
		effectstest.SeedLibraryTop(me, "Bear", "Creature — Bear", "{G}")
	}
	effectstest.CastCatalogSpell(t, g, "Harald, King of Skemfar", "Legendary Creature — Elf Warrior", oldB069Harald, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestChooseCardsFor(g, me.ID); p != nil && len(p.ChooseCards) > 0 {
		t.Errorf("offered %v with no Elf, Warrior or Tyvar", p.ChooseCards)
	}
}

// --- Cyclone Summoner --------------------------------------------------

func TestOldB069CycloneSummonerBouncesEverythingButGiantsWizardsAndLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	giant := effectstest.B12Creature(g, me.ID, "Giant", "Creature — Giant", 2, 2)
	wizard := effectstest.B12Creature(g, opp.ID, "Wizard", "Creature — Human Wizard", 2, 2)
	land := effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	rock := effectstest.PushCatalogPermanent(g, opp.ID, "Rock", "Artifact", "", false)
	ench := effectstest.PushCatalogPermanent(g, opp.ID, "Ward", "Enchantment", "", false)
	var tok uuid.UUID
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(opp.ID, game.Card{Name: "Goblin", TypeLine: "Token Creature — Goblin", Power: 1, Toughness: 1}, 1)
	})
	effectstest.CastCatalogSpell(t, g, "Cyclone Summoner", "Creature — Giant Wizard", oldB069Cyclone, nil)
	effectstest.PassPriorityAroundTable(t, g)
	_ = tok
	for _, id := range []uuid.UUID{giant, wizard, land} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("a Giant, Wizard or land was returned")
		}
	}
	if g.Battlefield.Contains(bear) || !me.Hand.Contains(bear) {
		t.Error("the Bear was not returned to its owner's hand")
	}
	if g.Battlefield.Contains(rock) || !opp.Hand.Contains(rock) {
		t.Error("the artifact was not returned to its owner's hand")
	}
	if g.Battlefield.Contains(ench) || !opp.Hand.Contains(ench) {
		t.Error("the enchantment was not returned to its owner's hand")
	}
	if n := effectstest.CountBattlefieldNamed(g, opp.ID, "Goblin"); n != 0 {
		t.Error("the token was not returned")
	}
	if n := effectstest.NamedOnBattlefield(g, "Cyclone Summoner", me.ID); n != 1 {
		t.Errorf("%d Summoners on the battlefield, want 1 (a Giant Wizard stays)", n)
	}
}

func TestOldB069CycloneSummonerDoesNothingWhenNotCastFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	id := effectstest.B38Graveyard(g, me.ID, "Cyclone Summoner", "Creature — Giant Wizard", "{5}{U}{U}", 7, 7)
	g.WithWriteLock(func() {
		if c, ok := g.LookupCardForEffect(id); ok {
			c.OracleID = oldB069Cyclone
		}
	})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, id); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("a Summoner that was not cast from the hand returned the Bear")
	}
}

// --- Koll, the Forgemaster ---------------------------------------------

func TestOldB069KollReturnsAnEnchantedOrEquippedNontokenCreatureThatDies(t *testing.T) {
	for _, tl := range []string{"Artifact — Equipment", "Enchantment — Aura"} {
		t.Run(tl, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.B12Push(g, me.ID, "Koll, the Forgemaster", "Legendary Creature — Dwarf Warrior", oldB069Koll, 2, 2)
			bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			oldB069Attach(g, me.ID, bear, tl)
			g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
			g.RunStateChecksForTest()
			effectstest.PassPriorityAroundTable(t, g)
			if !me.Hand.Contains(bear) {
				t.Error("the creature did not return to its owner's hand")
			}
		})
	}
}

func TestOldB069KollLeavesABareCreatureATokenAndKollHimselfInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	koll := effectstest.B12Push(g, me.ID, "Koll, the Forgemaster", "Legendary Creature — Dwarf Warrior", oldB069Koll, 2, 2)
	bare := effectstest.B12Creature(g, me.ID, "Bare", "Creature — Bear", 2, 2)
	oldB069Attach(g, me.ID, koll, "Artifact — Equipment")
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bare) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(bare) || !me.Graveyard.Contains(bare) {
		t.Error("an unattached creature came back")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(koll) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(koll) {
		t.Error("Koll returned himself: the ability says another creature")
	}
}

func TestOldB069KollPumpsOnlyEnchantedOrEquippedCreatureTokensYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Koll, the Forgemaster", "Legendary Creature — Dwarf Warrior", oldB069Koll, 2, 2)
	mk := func(owner uuid.UUID, token bool) uuid.UUID {
		tl := "Creature — Bear"
		if token {
			tl = "Token Creature — Bear"
		}
		return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: tl,
			Power: 1, Toughness: 1, Owner: owner, Controller: owner})
	}
	armed, bare, nontoken, theirs := mk(me.ID, true), mk(me.ID, true), mk(me.ID, false), mk(opp.ID, true)
	oldB069Attach(g, me.ID, armed, "Artifact — Equipment")
	oldB069Attach(g, me.ID, nontoken, "Artifact — Equipment")
	oldB069Attach(g, opp.ID, theirs, "Enchantment — Aura")
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	for id, want := range map[uuid.UUID]int{armed: 2, bare: 1, nontoken: 1, theirs: 1} {
		if got := effectstest.EffectivePower(t, g, id); got != want {
			t.Errorf("power %d, want %d", got, want)
		}
	}
}

// --- Thornmantle Striker -----------------------------------------------

func oldB069StrikerMode(t *testing.T, g *game.Game, mode int, target uuid.UUID) {
	t.Helper()
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Thornmantle Striker", "Creature — Elf Rogue", oldB069Thornmant, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no mode prompt")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PickCard(t, g, me.ID, target)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB069ThornmantleStrikerShrinksByTheNumberOfElves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf Warrior", 1, 1)
	effectstest.B12Creature(g, me.ID, "Changeling", "Creature — Shapeshifter Changeling", 1, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 5, 5)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 5, 5)
	_ = mine
	oldB069StrikerMode(t, g, 1, victim)
	// The Elf and the Striker itself: two Elves (the Changeling type line is not an Elf).
	if pw, tough := effectstest.EffectivePTOf(t, g, victim); pw != 3 || tough != 3 {
		t.Errorf("Ogre is %d/%d, want 3/3 (-2/-2)", pw, tough)
	}
}

func TestOldB069ThornmantleStrikerRefusesItsOwnCreatureForTheShrinkMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 5, 5)
	effectstest.B12Creature(g, g.Seats[1].ID, "Ogre", "Creature — Ogre", 5, 5)
	effectstest.CastCatalogSpell(t, g, "Thornmantle Striker", "Creature — Elf Rogue", oldB069Thornmant, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no mode prompt")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil && effectstest.HasID(p.PickTargetCards, mine) {
		t.Fatal("the creature you control was offered for -X/-X")
	}
}

func TestOldB069ThornmantleStrikerRemovesXCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf Warrior", 1, 1)
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact",
		Owner: opp.ID, Controller: opp.ID, Counters: map[string]int{"charge": 5}})
	oldB069StrikerMode(t, g, 0, rock)
	// Two Elves: the prompt asks for exactly two of the five.
	for i := 0; i < 3; i++ {
		var q *game.PendingChoice
		for j := len(g.PendingChoices) - 1; j >= 0; j-- {
			if c := g.PendingChoices[j]; c != nil && c.Kind == game.PendingChoicePayAmount && c.Chooser == me.ID {
				q = c
				break
			}
		}
		if q == nil {
			break
		}
		if err := g.ResolvePayAmount(q.ID, me.ID, q.PayAmount.Max); err != nil {
			t.Fatalf("ResolvePayAmount: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if n := effectstest.CountersOn(g, rock, "charge"); n != 3 {
		t.Errorf("%d counters left, want 3 (two Elves, two removed)", n)
	}
}

// --- Sigrid, God-Favored -----------------------------------------------

func TestOldB069SigridExilesAnAttackingCreatureUntilShesGone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	attacker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Raider", TypeLine: "Creature — Ogre",
		Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID, AttackingTarget: me.ID})
	idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Ogre", 3, 3)
	sigrid := effectstest.CastCatalogSpell(t, g, "Sigrid, God-Favored", "Legendary Creature — Human Warrior", oldB069Sigrid, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	if err := g.ResolvePickTargets(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, effectstest.CardRefs(idle)); err == nil {
		t.Fatal("Sigrid targeted a creature that is neither attacking nor blocking")
	}
	effectstest.PickCard(t, g, me.ID, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(attacker) {
		t.Fatal("the attacking creature was not exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(sigrid) })
	g.RunStateChecksForTest()
	if n := effectstest.NamedOnBattlefield(g, "Raider", opp.ID); n != 1 {
		t.Errorf("after Sigrid left, %d of the exiled creature under its owner's control, want 1", n)
	}
}

func TestOldB069SigridHasProtectionFromGodCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sigrid := effectstest.B12Push(g, me.ID, "Sigrid, God-Favored", "Legendary Creature — Human Warrior", oldB069Sigrid, 2, 2)
	effectstest.AssertKeywords(t, g, sigrid, "first strike", "flash", "protection from God creatures")
}
