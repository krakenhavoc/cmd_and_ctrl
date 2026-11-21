package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB212Duodrone = "939b77f1-d7c4-4e52-b487-5866145ace01"
	gapB212Burakos  = "883ad210-917b-40e6-8682-31d6eb3ed12a"
	gapB212Marut    = "4ba37a9e-6b80-4099-bd79-736123d0554b"
	gapB212Doll     = "499dceab-1890-49c8-b35d-b059a1dc950f"
	gapB212Mastiff  = "6d00e24f-c89f-49fa-869e-d788021aa084"
	gapB212Slaad    = "788897f5-d49d-4294-8006-0009a3124ad0"
	gapB212Dragon   = "7d79fb4e-2ca0-404a-9663-1f5abce6e733"
	gapB212Devil    = "b73bbdcc-e872-4da5-8d80-0fc9e52418bf"
	gapB212Tyrant   = "ccd248ec-1029-47cd-b6b4-876505efd9dd"
	gapB212Rug      = "da1035c4-bafe-434e-b1bd-5b08960d4c91"
	gapB212Brain    = "e537bd23-7368-459b-badf-b7b7c112c88a"
)

func TestGapB212EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB212Duodrone: "Marching Duodrone", gapB212Burakos: "Burakos, Party Leader", gapB212Marut: "Marut",
		gapB212Doll: "Stuffy Doll", gapB212Mastiff: "Iron Mastiff", gapB212Slaad: "Gray Slaad",
		gapB212Slaad + "#1": "Entropic Decay", gapB212Dragon: "Ancient Brass Dragon", gapB212Devil: "Chain Devil",
		gapB212Tyrant: "Ghastly Death Tyrant", gapB212Rug: "Rug of Smothering", gapB212Brain: "Elder Brain",
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

func gapB212Attack(t *testing.T, g *game.Game, attacker uuid.UUID, defender *game.Player) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, defender.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
}

func gapB212PickMode(t *testing.T, g *game.Game, me uuid.UUID, mode int) {
	t.Helper()
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me)
	if c == nil {
		t.Fatalf("no mode_pick prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, me, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick(%d): %v", mode, err)
	}
}

func TestGapB212ChainDevilEdictsEveryPlayersNontokenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp1, opp2 := g.Seats[1], g.Seats[2]
	bear := effectstest.SeedCreature(g, "Bear", opp1.ID)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp2.ID, effects.RedGoblinToken(), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	devil := effectstest.CastCatalogSpell(t, g, "Chain Devil", "Creature — Devil", gapB212Devil, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("an opponent's only nontoken creature was not sacrificed")
	}
	if effectstest.FindBattlefieldByName(g, "Goblin") == uuid.Nil {
		t.Error("a token was sacrificed")
	}
	if g.Battlefield.Contains(devil) {
		t.Error("the caster's only nontoken creature is the Devil itself")
	}
}

func TestGapB212MarchingDuodrone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	duo := effectstest.PushCatalogPermanent(g, me.ID, "Marching Duodrone", "Artifact Creature — Construct", gapB212Duodrone, false)
	gapB212Attack(t, g, duo, opp)
	for _, s := range g.Seats {
		if n, _ := effectstest.CountTokensNamed(g, s.ID, "Treasure"); n != 1 {
			t.Errorf("%s has %d Treasures, want 1", s.Name, n)
		}
	}
}

func TestGapB212MarutMakesATreasurePerTreasureMana(t *testing.T) {
	for _, tc := range []struct {
		name     string
		treasure bool
		want     int
	}{{"cast with a Treasure's mana", true, 1}, {"cast with other mana", false, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.AdvanceToMain(t, g)
			if tc.treasure {
				g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 1) })
				treasure := effectstest.FindBattlefieldByName(g, "Treasure")
				if err := g.ActivateManaAbility(me.ID, treasure, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
					t.Fatalf("treasure mana: %v", err)
				}
				effectstest.FloatForTest(g, me, "CCCCCCC")
			} else {
				effectstest.FloatForTest(g, me, "CCCCCCCC")
			}
			id := uuid.New()
			me.Hand.PushTop(game.Card{InstanceID: id, Name: "Marut", TypeLine: "Artifact Creature — Construct", ManaCost: "{8}", OracleID: gapB212Marut, Power: 7, Toughness: 7, Owner: me.ID, Controller: me.ID})
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
				t.Fatalf("cast Marut: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			if n, _ := effectstest.CountTokensNamed(g, me.ID, "Treasure"); n != tc.want {
				t.Errorf("Treasures %d, want %d", n, tc.want)
			}
		})
	}
}

func TestGapB212RugOfSmotheringCostsLifePerSpellCastThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Rug of Smothering", "Artifact Creature — Construct", gapB212Rug, false)
	life := me.Life
	for i := 0; i < 2; i++ {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Free Spell", TypeLine: "Creature — Bear", ManaCost: "{0}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if me.Life != life-3 {
		t.Errorf("life %d, want %d (1 for the first spell, 2 for the second)", me.Life, life-3)
	}
}

func TestGapB212GhastlyDeathTyrantModes(t *testing.T) {
	t.Run("disintegration ray", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		aura := effectstest.PushCatalogPermanent(g, opp.ID, "Their Aura", "Enchantment", "", false)
		g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, aura).ManaCost = "{2}{W}" })
		life := me.Life
		effectstest.CastAndResolveCreature(t, g, "Ghastly Death Tyrant", "Creature — Beholder Skeleton", gapB212Tyrant)
		gapB212PickMode(t, g, me.ID, 0)
		if effectstest.LatestPickTarget(g, me.ID) != nil {
			effectstest.PickTriggerTarget(t, g, me.ID, aura)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(aura) {
			t.Error("the enchantment was not destroyed")
		}
		if me.Life != life-3 {
			t.Errorf("life %d, want %d", me.Life, life-3)
		}
	})
	t.Run("death ray", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		effectstest.CastAndResolveCreature(t, g, "Ghastly Death Tyrant", "Creature — Beholder Skeleton", gapB212Tyrant)
		gapB212PickMode(t, g, me.ID, 1)
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.EffectiveAbilitiesContain(t, g, bear, "deathtouch") {
			t.Error("no deathtouch on your creature")
		}
	})
}

func TestGapB212BurakosPartyLeaderCountsAsAWholeParty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	burakos := effectstest.CastHoldCreature(t, g, "Burakos, Party Leader", "Legendary Creature — Orc", gapB212Burakos, 2, 4)
	g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, burakos).SummonedThisTurn = false })
	for _, st := range []string{"Cleric", "Rogue", "Warrior", "Wizard"} {
		if c, _ := g.LookupCardForEffect(burakos); !c.HasSubtype(st) {
			t.Errorf("Burakos is not a %s", st)
		}
	}
	life := opp.Life
	gapB212Attack(t, g, burakos, opp)
	// Burakos alone is one creature in the party.
	if opp.Life != life-1 {
		t.Errorf("defending player life %d, want %d", opp.Life, life-1)
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("Treasures %d, want 1", n)
	}
}

func TestGapB212GraySlaadGainsMenaceAndDeathtouchWithFourCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	slaad := effectstest.PushCatalogPermanent(g, me.ID, "Gray Slaad", "Creature — Frog Horror", gapB212Slaad, false)
	if effectstest.EffectiveAbilitiesContain(t, g, slaad, "menace") {
		t.Fatal("menace with an empty graveyard")
	}
	for i := 0; i < 4; i++ {
		bear := effectstest.B12Creature(g, me.ID, "Dead Bear", "Creature — Bear", 1, 1)
		effectstest.KillCreature(t, g, me.ID, bear)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, slaad, "menace") || !effectstest.EffectiveAbilitiesContain(t, g, slaad, "deathtouch") {
		t.Error("no menace and deathtouch with four creature cards in the graveyard")
	}
}

func TestGapB212EntropicDecayMillsFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")
	before := me.Graveyard.Size()
	effectstest.CastCatalogSpell(t, g, "Entropic Decay", "Sorcery — Adventure", gapB212Slaad+"#1", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Graveyard.Size() - before; got < 4 {
		t.Errorf("milled %d, want 4", got)
	}
}

func TestGapB212StuffyDollReflectsDamageToTheChosenPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	doll := effectstest.CastHoldCreature(t, g, "Stuffy Doll", "Artifact Creature — Construct", gapB212Doll, 0, 1)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(doll) {
		t.Fatal("the Doll is not on the battlefield")
	}
	life := opp.Life
	effectstest.UntapForTest(g, doll)
	g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, doll).SummonedThisTurn = false })
	if err := g.ActivateCatalogAbility(me.ID, doll, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("chosen player's life %d, want %d", opp.Life, life-1)
	}
	if !g.Battlefield.Contains(doll) {
		t.Error("the indestructible Doll died of its own damage")
	}
}

func TestGapB212IronMastiffRollsAndHitsSomeone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	mastiff := effectstest.PushCatalogPermanent(g, me.ID, "Iron Mastiff", "Artifact Creature — Dog", gapB212Mastiff, false)
	before := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		before[s.ID] = s.Life
	}
	gapB212Attack(t, g, mastiff, opp)
	lost := 0
	for _, s := range g.Seats {
		if s.Life < before[s.ID] {
			lost++
		}
	}
	if lost == 0 {
		t.Error("nobody was dealt damage by the Mastiff's trigger")
	}
}

func TestGapB212AncientBrassDragonReanimatesWithinTheRoll(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	dragon := effectstest.PushCatalogPermanent(g, me.ID, "Ancient Brass Dragon", "Creature — Elder Dragon", gapB212Dragon, false)
	free := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: free, Name: "Zero Drop", TypeLine: "Creature — Bear", ManaCost: "{0}", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	gapB212Attack(t, g, dragon, opp)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickTriggerTarget(t, g, me.ID, free)
		effectstest.PassPriorityAroundTable(t, g)
	}
	found := false
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Zero Drop" && c.Controller == me.ID {
			found = true
		}
	}
	if !found {
		t.Error("a mana value 0 creature from a graveyard was not put onto the battlefield under your control")
	}
}

func TestGapB212ElderBrainExilesTheHandAndLetsYouPlayIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	brain := effectstest.PushCatalogPermanent(g, me.ID, "Elder Brain", "Creature — Horror", gapB212Brain, false)
	hand := opp.Hand.Size()
	if hand == 0 {
		t.Fatal("the opponent has no hand to exile")
	}
	effectstest.SeedLibrary(opp, "A", "B", "C", "D", "E", "F", "G", "H", "I")
	gapB212Attack(t, g, brain, opp)
	if got := opp.Hand.Size(); got != hand {
		t.Errorf("opponent's hand %d after drawing that many, want %d", got, hand)
	}
	linked := 0
	g.WithWriteLock(func() {
		ref := effectstest.ObjectOf(t, g, brain)
		linked = len(g.ExiledWithForEffect(game.PermissionCardRef(ref)))
	})
	if linked != hand {
		t.Errorf("%d cards exiled with the Elder Brain, want %d", linked, hand)
	}
}
