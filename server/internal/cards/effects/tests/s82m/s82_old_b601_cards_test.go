package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b601_cards_test.go — Betrayers of Kamigawa, S82 OLD batch 601
// (#5602). Each test plays the real catalog entry and asserts what its text
// says.

const (
	oldB601Akki      = "bf9c5c0d-86a5-467a-863c-653279f2e6e7"
	oldB601Hulk      = "ce4570eb-f637-45d1-afbc-f11357c09bf9"
	oldB601Child     = "f94002a1-582f-4b9d-a4f7-dae816f21c1d"
	oldB601Verdict   = "28ea1f58-3310-4cf1-a834-40f2cc3b6898"
	oldB601Nezumi    = "3eae5419-b686-4321-a89d-cb602c487bfc"
	oldB601Summons   = "6be663a5-70e2-435a-91ae-16168402787b"
	oldB601Kodama    = "88188cc1-3211-4bf8-a591-613b66192695"
	oldB601Orb       = "2c24b7c3-7c90-4109-a4aa-d9cae6febf34"
	oldB601Iwamori   = "35734003-0c94-4a0c-9441-ab989e73bb11"
	oldB601Scourge   = "540d1878-7afe-48a5-8b40-3dd1b06a024a"
	oldB601Springcal = "675a0eff-4da0-45cd-9371-fca96c313f4c"
	oldB601Meddling  = "6c7239b0-e67f-4848-b244-8be9915bd2dc"
)

func TestOldB601CardsAreFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB601Akki: "Akki Raider", oldB601Hulk: "Scaled Hulk", oldB601Child: "Child of Thorns",
		oldB601Verdict: "Terashi's Verdict", oldB601Nezumi: "Nezumi Shadow-Watcher", oldB601Summons: "Sosuke's Summons",
		oldB601Kodama: "Kodama of the Center Tree", oldB601Orb: "Orb of Dreams", oldB601Iwamori: "Iwamori of the Open Fist",
		oldB601Scourge: "Scourge of Numai", oldB601Springcal: "Sakura-Tribe Springcaller", oldB601Meddling: "Minamo's Meddling",
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

func TestOldB601AkkiRaiderGrowsForEveryLandThatDiesAnywhere(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	raider := effectstest.B12Push(g, me.ID, "Akki Raider", "Creature — Goblin Warrior", oldB601Akki, 2, 1)
	mine := effectstest.PushLand(g, me.ID, "Island")
	theirs := effectstest.PushLand(g, opp.ID, "Island")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.P1Destroy(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, raider); p != 2 {
		t.Errorf("a dying creature grew the Raider to %d power", p)
	}
	effectstest.P1Destroy(t, g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1Destroy(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, raider); p != 4 || tough != 1 {
		t.Errorf("the Raider is %d/%d, want 4/1 after two lands died", p, tough)
	}
	effectstest.DcCleanup(g)
	if p, _ := effectstest.EffectivePTOf(t, g, raider); p != 2 {
		t.Errorf("the bonus outlived the cleanup step (%d power)", p)
	}
}

func TestOldB601ScaledHulkGrowsForSpiritAndArcaneSpellsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hulk := effectstest.B12Push(g, me.ID, "Scaled Hulk", "Creature — Spirit", oldB601Hulk, 4, 4)
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "A Spirit", "Creature — Spirit", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "An Arcane Spell", "Sorcery — Arcane", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, hulk); p != 8 || tough != 8 {
		t.Errorf("the Hulk is %d/%d, want 8/8 (two Spirit or Arcane spells)", p, tough)
	}
	effectstest.DcCleanup(g)
	if p, tough := effectstest.EffectivePTOf(t, g, hulk); p != 4 || tough != 4 {
		t.Errorf("the bonus outlived the cleanup step (%d/%d)", p, tough)
	}
}

func TestOldB601ChildOfThornsSacrificesItselfToPumpATarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	child := effectstest.B12Push(g, me.ID, "Child of Thorns", "Creature — Spirit", oldB601Child, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, child, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(child) {
		t.Error("the Child was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the Bear is %d/%d, want 3/3", p, tough)
	}
	effectstest.DcCleanup(g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("the bonus outlived the cleanup step (%d/%d)", p, tough)
	}
}

func TestOldB601TerashisVerdictDestroysOnlyASmallAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	small := effectstest.B12Creature(g, me.ID, "Small Ogre", "Creature — Ogre", 3, 3)
	big := effectstest.B12Creature(g, me.ID, "Big Ogre", "Creature — Ogre", 4, 4)
	idle := effectstest.B12Creature(g, me.ID, "Idle Bear", "Creature — Bear", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, small, big)
	for _, bad := range []uuid.UUID{big, idle} {
		if _, err := effectstest.RprB19Cast(t, g, opp, "Terashi's Verdict", "Instant — Arcane", "{1}{W}", oldB601Verdict, "{W}{W}", game.CastSpellParams{Targets: effectstest.CardTarget(bad)}); err == nil {
			t.Fatal("a creature with power over 3, or one that is not attacking, was accepted as the target")
		}
		g.WithWriteLock(func() { opp.ManaPool = game.ManaPool{} })
	}
	if _, err := effectstest.RprB19Cast(t, g, opp, "Terashi's Verdict", "Instant — Arcane", "{1}{W}", oldB601Verdict, "{W}{W}", game.CastSpellParams{Targets: effectstest.CardTarget(small)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the small attacker was not destroyed")
	}
	if !g.Battlefield.Contains(big) || !g.Battlefield.Contains(idle) {
		t.Error("a creature that was not a legal target left the battlefield")
	}
}

func TestOldB601NezumiShadowWatcherDestroysOnlyANinja(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	watcher := effectstest.B12Push(g, me.ID, "Nezumi Shadow-Watcher", "Creature — Rat Warrior", oldB601Nezumi, 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	ninja := effectstest.B12Creature(g, opp.ID, "Ninja", "Creature — Human Ninja", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, watcher, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Fatal("a non-Ninja was accepted as the target")
	}
	if !g.Battlefield.Contains(watcher) {
		t.Fatal("the Shadow-Watcher was sacrificed for a refused activation")
	}
	if err := g.ActivateCatalogAbility(me.ID, watcher, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(ninja)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(watcher) || g.Battlefield.Contains(ninja) {
		t.Error("the Shadow-Watcher or the Ninja is still on the battlefield")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the Bear left the battlefield")
	}
}

func TestOldB601OrbOfDreamsMakesEveryPermanentEnterTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	orb := effectstest.CastAndResolveCreature(t, g, "Orb of Dreams", "Artifact", oldB601Orb)
	if c, ok := effectstest.BattlefieldCard(g, orb); !ok || !c.Tapped {
		t.Errorf("the Orb itself: on battlefield=%v tapped=%v, want it to enter tapped", ok, c.Tapped)
	}
	bear := effectstest.CastAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	if c, ok := effectstest.BattlefieldCard(g, bear); !ok || !c.Tapped {
		t.Errorf("a creature of mine: on battlefield=%v tapped=%v, want tapped", ok, c.Tapped)
	}
	land := effectstest.PlayLandFromHand(t, g, "Forest", "")
	if c, ok := effectstest.BattlefieldCard(g, land); !ok || !c.Tapped {
		t.Errorf("a land: on battlefield=%v tapped=%v, want tapped", ok, c.Tapped)
	}
	_ = me
	_ = opp
}

func TestOldB601ScourgeOfNumaiCostsLifeUnlessYouControlAnOgre(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Scourge of Numai", "Creature — Demon Spirit", oldB601Scourge, 4, 4)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	life := me.Life
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-2 {
		t.Errorf("no Ogre: life %d → %d, want -2", life, me.Life)
	}
	effectstest.B12Creature(g, me.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	life = me.Life
	if g.Stack.Size() != 0 {
		t.Error("the trigger went on the stack although I control an Ogre")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("an Ogre: life %d → %d, want no loss", life, me.Life)
	}
}

func TestOldB601SakuraTribeSpringcallerKeepsItsManaThroughTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Sakura-Tribe Springcaller", "Creature — Snake Shaman", oldB601Springcal, 2, 4)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(me.ManaPool.Units()); n != 1 {
		t.Fatalf("%d mana after the upkeep trigger, want 1", n)
	}
	effectstest.AdvanceToMain(t, g)
	if n := len(me.ManaPool.Units()); n != 1 {
		t.Errorf("%d mana in the main phase, want the {G} kept through the steps", n)
	}
}

func TestOldB601IwamoriLetsEachOpponentPutALegendaryCreatureIn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, o1, o2 := g.Seats[0], g.Seats[1], g.Seats[2]
	legend := func(p *game.Player, name string) game.Card {
		c := game.NewCard(name, p.ID)
		c.TypeLine = "Legendary Creature — Human"
		p.Hand.PushTop(c)
		return c
	}
	plain := game.NewCard("Plain Bear", o1.ID)
	plain.TypeLine = "Creature — Bear"
	o1.Hand.PushTop(plain)
	l1 := legend(o1, "Legend One")
	l2 := legend(o2, "Legend Two")
	effectstest.CastAndResolveCreature(t, g, "Iwamori of the Open Fist", "Legendary Creature — Human Monk", oldB601Iwamori)
	effectstest.PassPriorityAroundTable(t, g)
	for _, who := range []struct {
		p    *game.Player
		card game.Card
	}{{o1, l1}, {o2, l2}} {
		effectstest.AnswerChooseCards(t, g, who.p.ID, who.card.InstanceID)
		if c, ok := effectstest.BattlefieldCard(g, who.card.InstanceID); !ok || c.Controller != who.p.ID {
			t.Errorf("%s: on battlefield=%v, controller %v, want under its owner's control", who.card.Name, ok, c.Controller)
		}
	}
	if g.Battlefield.Contains(plain.InstanceID) {
		t.Error("a nonlegendary creature card was put onto the battlefield")
	}
	_ = me
}

func TestOldB601KodamaCountsSpiritsAndSoulshiftsForThatMany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	kodama := effectstest.B12Push(g, me.ID, "Kodama of the Center Tree", "Legendary Creature — Spirit", oldB601Kodama, 0, 0)
	if p, tough := effectstest.EffectivePTOf(t, g, kodama); p != 1 || tough != 1 {
		t.Errorf("alone the Kodama is %d/%d, want 1/1", p, tough)
	}
	effectstest.B12Push(g, me.ID, "Spirit A", "Creature — Spirit", "", 1, 1)
	effectstest.B12Push(g, me.ID, "Spirit B", "Creature — Spirit", "", 1, 1)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if p, tough := effectstest.EffectivePTOf(t, g, kodama); p != 3 || tough != 3 {
		t.Errorf("with two more Spirits the Kodama is %d/%d, want 3/3", p, tough)
	}
	grave := func(name, typeLine, cost string) uuid.UUID {
		id := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: me.ID, Controller: me.ID})
		return id
	}
	fits := grave("Fits", "Creature — Spirit", "{1}{G}")
	big := grave("Too Big", "Creature — Spirit", "{2}{G}")
	bear := grave("Bear Card", "Creature — Bear", "{G}")
	effectstest.KillCreature(t, g, me.ID, kodama)
	pick := effectstest.PassUntilPickTarget(t, g, me.ID)
	if pick == nil {
		t.Fatal("no soulshift target prompt")
	}
	if !effectstest.HasID(pick.PickTargetCards, fits) || effectstest.HasID(pick.PickTargetCards, big) || effectstest.HasID(pick.PickTargetCards, bear) {
		t.Errorf("soulshift X offers %v, want only the mana value 2 Spirit %v", pick.PickTargetCards, fits)
	}
	effectstest.PickCard(t, g, me.ID, fits)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(fits) {
		t.Error("the Spirit with mana value 2 did not return (two Spirits remain)")
	}
	if me.Hand.Contains(big) || me.Hand.Contains(bear) {
		t.Error("a card that is too big or not a Spirit returned")
	}
}

func TestOldB601SosukesSummonsMakesSnakesAndRecursFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Sosuke's Summons", "Sorcery", oldB601Summons, nil)
	effectstest.PassPriorityAroundTable(t, g)
	snakes := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsToken() && c.HasSubtype("Snake") {
			snakes++
		}
	}
	if snakes != 2 {
		t.Fatalf("%d Snake tokens, want 2", snakes)
	}
	var summons uuid.UUID
	for _, c := range me.Graveyard.Cards {
		if c.OracleID == oldB601Summons {
			summons = c.InstanceID
		}
	}
	if summons == uuid.Nil {
		t.Fatal("the Summons is not in the graveyard")
	}
	if me.Hand.Contains(summons) {
		t.Fatal("the Summons recurred off its own tokens")
	}
	effectstest.CastAndResolveCreature(t, g, "Plain Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(summons) {
		t.Error("a non-Snake recurred the Summons")
	}
	effectstest.CastAndResolveCreature(t, g, "Real Snake", "Creature — Snake", "")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(summons) {
		t.Error("a nontoken Snake did not return the Summons to hand")
	}
}

func TestOldB601MinamoMeddlingCountersAndDiscardsSplicedNames(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	vortex := effectstest.HandCardFull(me, "Consuming Vortex", "Instant — Arcane", "{1}{U}", effectstest.ConsumingVortexOracle, []string{"U"})
	ray := effectstest.HandCardFull(me, "Glacial Ray", "Instant — Arcane", "{1}{R}", effectstest.GlacialRayOracle, []string{"R"})
	ray2 := effectstest.HandCardFull(me, "Glacial Ray", "Instant — Arcane", "{1}{R}", effectstest.GlacialRayOracle, []string{"R"})
	other := effectstest.HandCardFull(me, "Other Card", "Instant", "{1}", "", nil)
	opp := g.Seats[1]
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{U}{C}{R}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, vortex, game.CastSpellParams{Strict: true, Splices: []uuid.UUID{ray},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetPlayer, ID: opp.ID, Splice: 1}}}); err != nil {
		t.Fatalf("splice: %v", err)
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{U}{U}"); err != nil {
		t.Fatal(err)
	}
	meddling := effectstest.HandCardFull(me, "Minamo's Meddling", "Instant", "{2}{U}{U}", oldB601Meddling, []string{"U"})
	if err := g.CastSpell(me.ID, meddling, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: vortex}}}); err != nil {
		t.Fatalf("cast Meddling: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the countered spell still returned the Bear")
	}
	if me.Hand.Contains(ray) || me.Hand.Contains(ray2) {
		t.Error("a card named like the spliced Glacial Ray stayed in hand")
	}
	if !me.Hand.Contains(other) {
		t.Error("a card with another name was discarded")
	}
}
