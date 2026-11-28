package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b735_test.go — Portal Second Age, batch 735 (#5625): instants and
// sorceries, pump and combat tricks, removal, graveyard recursion, life gain.
// Each test plays the real catalog entry.

const (
	oldB735CruelEdict   = "10c585c4-bf5b-4d8f-94a9-e9a5036a688f"
	oldB735WarCry       = "a6c8c660-2b06-4395-8c86-dd540760d9ae"
	oldB735TidalSurge   = "be738992-77fe-498d-b219-e5da4ce5bf07"
	oldB735Razorclaw    = "3ee07899-e6d6-4dcf-8596-4d2420690d7c"
	oldB735Warrior      = "b3ae1481-9fa7-45dd-bbf5-de2d815c4227"
	oldB735TownSentry   = "ec1ac4df-b031-43b4-a2de-ca58a1599053"
	oldB735Arsonist     = "00af96af-5eae-4044-a2f7-a08cd0699d1e"
	oldB735JustFate     = "2619692a-07a0-4f50-88b0-4cec6dbc8600"
	oldB735ReturnOfNs   = "6082f66e-3641-4b09-bb04-86bf29ee1c6f"
	oldB735DejaVu       = "7408b9c5-7266-4627-be4e-b691cf5c622c"
	oldB735Harmony      = "6c943c9d-090b-41c1-9003-24c91601e589"
	oldB735Festival     = "a5a3cd2c-20ae-49a8-88fa-e8e3e7cc50db"
	oldB735BearType     = "Creature — Bear"
	oldB735SorceryType  = "Sorcery"
	oldB735NightstalkTy = "Creature — Nightstalker"
)

func TestOldB735EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB735CruelEdict: "Cruel Edict", oldB735WarCry: "Goblin War Cry", oldB735TidalSurge: "Tidal Surge",
		oldB735Razorclaw: "Razorclaw Bear", oldB735Warrior: "Norwood Warrior", oldB735TownSentry: "Town Sentry",
		oldB735Arsonist: "Ogre Arsonist", oldB735JustFate: "Just Fate",
		oldB735ReturnOfNs: "Return of the Nightstalkers", oldB735DejaVu: "Déjà Vu",
		oldB735Harmony: "Harmony of Nature", oldB735Festival: "Festival of Trokin",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v; want Full with none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB735Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func oldB735OppTarget(opp *game.Player) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
}

func TestOldB735CruelEdictMakesTheOpponentSacrificeACreatureOfTheirChoice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB735Seats(g)
	keep := effectstest.B12Creature(g, opp.ID, "Keeper", oldB735BearType, 3, 3)
	lose := effectstest.B12Creature(g, opp.ID, "Loser", oldB735BearType, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", oldB735BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Cruel Edict", oldB735SorceryType, oldB735CruelEdict, oldB735OppTarget(opp))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, opp.ID, lose)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, lose) || !effectstest.OnBattlefieldByID(g, keep) || !effectstest.OnBattlefieldByID(g, mine) {
		t.Errorf("loser on battlefield %v, keeper %v, caster's creature %v; want only the chosen one gone",
			effectstest.OnBattlefieldByID(g, lose), effectstest.OnBattlefieldByID(g, keep), effectstest.OnBattlefieldByID(g, mine))
	}
}

func TestOldB735CruelEdictRefusesYourselfAsTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB735Seats(g)
	if err := effectstest.CastCatalogSpellErr(t, g, "Cruel Edict", oldB735SorceryType, oldB735CruelEdict,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}); err == nil {
		t.Error("Cruel Edict targeted its own controller")
	}
}

func TestOldB735GoblinWarCryLeavesOnlyTheChosenCreatureAbleToBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB735Seats(g)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", oldB735BearType, 2, 2)
	chosen := effectstest.B12Creature(g, opp.ID, "Chosen", oldB735BearType, 2, 2)
	other1 := effectstest.B12Creature(g, opp.ID, "Other 1", oldB735BearType, 2, 2)
	other2 := effectstest.B12Creature(g, opp.ID, "Other 2", oldB735BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Goblin War Cry", oldB735SorceryType, oldB735WarCry, oldB735OppTarget(opp))
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, opp.ID)
	if pick == nil {
		t.Fatal("the target opponent was not asked to choose a creature")
	}
	if err := g.ResolveChooseCards(pick.ID, opp.ID, []uuid.UUID{chosen}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AttackIntoBlocks(t, g, opp.ID, attacker)
	for name, id := range map[string]uuid.UUID{"first other creature": other1, "second other creature": other2} {
		if err := g.Clone().DeclareBlocker(id, attacker); err == nil {
			t.Errorf("the %s blocked", name)
		}
	}
	if err := g.DeclareBlocker(chosen, attacker); err != nil {
		t.Errorf("the chosen creature was refused: %v", err)
	}
}

func TestOldB735TidalSurgeTapsUpToThreeCreaturesWithoutFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := oldB735Seats(g)
	a := effectstest.B12Creature(g, opp.ID, "A", oldB735BearType, 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", oldB735BearType, 2, 2)
	flier := effectstest.BrCreature(g, opp.ID, "Flier", "Creature — Bird", 1, 1, "flying")
	effectstest.CastCatalogSpell(t, g, "Tidal Surge", oldB735SorceryType, oldB735TidalSurge,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.IsTapped(g, a) || !effectstest.IsTapped(g, b) {
		t.Error("a chosen creature was not tapped")
	}
	if effectstest.IsTapped(g, flier) {
		t.Error("an unchosen creature was tapped")
	}
	g2 := effectstest.NewCatalogGame(t)
	_, opp2 := oldB735Seats(g2)
	fl := effectstest.BrCreature(g2, opp2.ID, "Flier", "Creature — Bird", 1, 1, "flying")
	if err := effectstest.CastCatalogSpellErr(t, g2, "Tidal Surge", oldB735SorceryType, oldB735TidalSurge,
		[]game.TargetRef{{Kind: game.TargetCard, ID: fl}}); err == nil {
		t.Error("Tidal Surge targeted a creature with flying")
	}
	g3 := effectstest.NewCatalogGame(t)
	_, opp3 := oldB735Seats(g3)
	var four []game.TargetRef
	for i := 0; i < 4; i++ {
		four = append(four, game.TargetRef{Kind: game.TargetCard, ID: effectstest.B12Creature(g3, opp3.ID, "Bear", oldB735BearType, 2, 2)})
	}
	if err := effectstest.CastCatalogSpellErr(t, g3, "Tidal Surge", oldB735SorceryType, oldB735TidalSurge, four); err == nil {
		t.Error("Tidal Surge took four targets")
	}
}

func TestOldB735BecomesBlockedBearsGrowOnce(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		grow         int
	}{{"Razorclaw Bear", oldB735Razorclaw, 2}, {"Norwood Warrior", oldB735Warrior, 1}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := oldB735Seats(g)
		bear := effectstest.B12Push(g, me.ID, tc.name, oldB735BearType, tc.oracle, 3, 3)
		b1 := effectstest.B12Creature(g, opp.ID, "Blocker 1", oldB735BearType, 1, 1)
		b2 := effectstest.B12Creature(g, opp.ID, "Blocker 2", oldB735BearType, 1, 1)
		effectstest.AttackIntoBlocks(t, g, opp.ID, bear)
		for _, b := range []uuid.UUID{b1, b2} {
			if err := g.DeclareBlocker(b, bear); err != nil {
				t.Fatalf("DeclareBlocker: %v", err)
			}
		}
		effectstest.LockInBlocks(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3+tc.grow || tough != 3+tc.grow {
			t.Errorf("%s is %d/%d after being double-blocked, want %d/%d", tc.name, p, tough, 3+tc.grow, 3+tc.grow)
		}
	}
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB735Seats(g)
	unblocked := effectstest.B12Push(g, me.ID, "Razorclaw Bear", oldB735BearType, oldB735Razorclaw, 3, 3)
	effectstest.AttackIntoBlocks(t, g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, unblocked)
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, unblocked); p != 3 {
		t.Errorf("an unblocked Razorclaw Bear is %d power, want 3", p)
	}
}

func TestOldB735TownSentryGetsPlusZeroPlusTwoWhenItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB735Seats(g)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", oldB735BearType, 2, 2)
	sentry := effectstest.B12Push(g, opp.ID, "Town Sentry", "Creature — Human Soldier", oldB735TownSentry, 2, 2)
	effectstest.AttackIntoBlocks(t, g, opp.ID, attacker)
	if err := g.DeclareBlocker(sentry, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, sentry); p != 2 || tough != 4 {
		t.Errorf("the Sentry is %d/%d after blocking, want 2/4", p, tough)
	}
}

func TestOldB735OgreArsonistDestroysTargetLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB735Seats(g)
	land := effectstest.PushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	mine := effectstest.PushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")
	effectstest.CastCatalogSpell(t, g, "Ogre Arsonist", "Creature — Ogre", oldB735Arsonist, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerTriggerTargets(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, land) {
		t.Error("the targeted land survived")
	}
	if !effectstest.OnBattlefieldByID(g, mine) {
		t.Error("an untargeted land was destroyed")
	}
}

func TestOldB735JustFateDestroysAnAttackerOnlyWhenYouWereAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	attacker := effectstest.SfPush(g, me.ID, "Attacker", "Creature — Test", 3, []string{"R"})
	home := effectstest.SfPush(g, me.ID, "Home", "Creature — Test", 1, []string{"R"})
	if err := effectstest.CastCatalogSpellErr(t, g, "Just Fate", "Instant", oldB735JustFate, effectstest.CardTarget(attacker)); err == nil {
		t.Fatal("cast outside the declare attackers step")
	}
	effectstest.SfAttackHold(t, g, opp.ID, attacker)
	card := effectstest.SfInstant("Just Fate", oldB735JustFate)
	card.InstanceID, card.Owner, card.Controller = uuid.New(), third.ID, third.ID
	third.Hand.PushTop(card)
	if err := g.CastSpell(third.ID, card.InstanceID, game.CastSpellParams{Targets: effectstest.CardTarget(attacker)}); err == nil {
		t.Fatal("cast by a player nobody attacked")
	}
	card2 := effectstest.SfInstant("Just Fate", oldB735JustFate)
	card2.InstanceID, card2.Owner, card2.Controller = uuid.New(), opp.ID, opp.ID
	opp.Hand.PushTop(card2)
	if err := g.CastSpell(opp.ID, card2.InstanceID, game.CastSpellParams{Targets: effectstest.CardTarget(home), Strict: true}); err == nil {
		t.Fatal("Just Fate targeted a creature that isn't attacking")
	}
	if err := g.CastSpell(opp.ID, card2.InstanceID, game.CastSpellParams{Targets: effectstest.CardTarget(attacker), Strict: true}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, attacker) {
		t.Error("the attacker survived")
	}
	if !effectstest.OnBattlefieldByID(g, home) {
		t.Error("another creature was destroyed")
	}
}

func TestOldB735ReturnOfTheNightstalkersReturnsThemThenDestroysYourSwamps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB735Seats(g)
	ns1 := effectstest.PushGraveyardCardTyped(me, "Dead Nightstalker", oldB735NightstalkTy)
	ns2 := effectstest.PushGraveyardCardTyped(me, "Other Nightstalker", oldB735NightstalkTy)
	bear := effectstest.PushGraveyardCardTyped(me, "Dead Bear", oldB735BearType)
	theirs := effectstest.PushGraveyardCardTyped(opp, "Their Nightstalker", oldB735NightstalkTy)
	swamp := effectstest.PushLandFor(g, me.ID, "Swamp", "Basic Land — Swamp")
	forest := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	theirSwamp := effectstest.PushLandFor(g, opp.ID, "Swamp", "Basic Land — Swamp")
	effectstest.CastCatalogSpell(t, g, "Return of the Nightstalkers", oldB735SorceryType, oldB735ReturnOfNs, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{ns1, ns2} {
		if !effectstest.OnBattlefieldByID(g, id) {
			t.Error("one of your Nightstalker cards did not return")
		}
	}
	if effectstest.OnBattlefieldByID(g, bear) || effectstest.OnBattlefieldByID(g, theirs) {
		t.Error("a card that is not your Nightstalker returned")
	}
	if effectstest.OnBattlefieldByID(g, swamp) {
		t.Error("your Swamp was not destroyed")
	}
	if !effectstest.OnBattlefieldByID(g, forest) || !effectstest.OnBattlefieldByID(g, theirSwamp) {
		t.Error("a Forest or an opponent's Swamp was destroyed")
	}
}

func TestOldB735DejaVuReturnsASorceryCardOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB735Seats(g)
	spell := effectstest.PushGraveyardCardTyped(me, "Old Sorcery", oldB735SorceryType)
	instant := effectstest.PushGraveyardCardTyped(me, "Old Instant", "Instant")
	if err := effectstest.CastCatalogSpellErr(t, g, "Déjà Vu", oldB735SorceryType, oldB735DejaVu, effectstest.CardTarget(instant)); err == nil {
		t.Fatal("Déjà Vu targeted an instant card")
	}
	effectstest.CastCatalogSpell(t, g, "Déjà Vu", oldB735SorceryType, oldB735DejaVu, effectstest.CardTarget(spell))
	effectstest.PassPriorityAroundTable(t, g)
	inHand, inYard := false, false
	for _, c := range me.Hand.Cards {
		if c.InstanceID == spell {
			inHand = true
		}
	}
	for _, c := range me.Graveyard.Cards {
		if c.InstanceID == spell {
			inYard = true
		}
	}
	if !inHand || inYard {
		t.Error("the sorcery card did not return to hand")
	}
}

func TestOldB735HarmonyOfNatureGainsFourLifePerCreatureTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB735Seats(g)
	a := effectstest.B12Creature(g, me.ID, "A", oldB735BearType, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", oldB735BearType, 2, 2)
	c := effectstest.B12Creature(g, me.ID, "C", oldB735BearType, 2, 2)
	before := me.Life
	effectstest.CastCatalogSpell(t, g, "Harmony of Nature", oldB735SorceryType, oldB735Harmony, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt for the creatures to tap")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if got := me.Life - before; got != 8 {
		t.Errorf("gained %d life for two tapped creatures, want 8", got)
	}
	if !effectstest.IsTapped(g, a) || !effectstest.IsTapped(g, b) || effectstest.IsTapped(g, c) {
		t.Error("exactly the two chosen creatures should be tapped")
	}
}

func TestOldB735FestivalOfTrokinGainsTwoLifePerCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB735Seats(g)
	effectstest.B12Creature(g, me.ID, "A", oldB735BearType, 2, 2)
	effectstest.B12Creature(g, me.ID, "B", oldB735BearType, 2, 2)
	effectstest.B12Creature(g, me.ID, "C", oldB735BearType, 2, 2)
	effectstest.B12Creature(g, opp.ID, "Theirs", oldB735BearType, 2, 2)
	before := me.Life
	effectstest.CastCatalogSpell(t, g, "Festival of Trokin", oldB735SorceryType, oldB735Festival, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Life - before; got != 6 {
		t.Errorf("gained %d life with three creatures, want 6", got)
	}
}
