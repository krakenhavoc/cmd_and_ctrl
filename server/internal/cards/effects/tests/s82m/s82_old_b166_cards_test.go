package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b166_cards_test.go — Modern Horizons, S82 older sets, batch 166
// (#5519). Each test plays the real catalog entry.

const (
	oldB166RankOfficer = "52c73c88-3b87-4184-b8e1-cd2b5688800d"
	oldB166Besieged    = "674e2683-31c0-4fee-95fb-98b1201e41e7"
	oldB166Prayers     = "7b238254-51b4-4e68-a8c6-93ab1180012e"
	oldB166Throatseek  = "edb7bd88-971a-4fcc-9d76-d73f21a4926a"
	oldB166Mariner     = "7e4d1ce3-92f0-4acc-962b-8073d41d3673"
	oldB166Therapist   = "9b7870df-6bca-499f-bc6e-e57f2ddfe640"
	oldB166Golem       = "29b925e9-1e98-471c-9b23-e0372ac86485"
	oldB166Evasion     = "3e56606d-562e-4c77-9769-213997fea28e"
	oldB166Bastion     = "ae4a18ec-70a3-4d21-b9e5-b13ab4901600"
	oldB166Strix       = "c687af9b-2159-4965-a068-ce4db1ce801c"
	oldB166Bear        = "Creature — Bear"
)

func TestOldB166EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB166RankOfficer: "Rank Officer",
		oldB166Besieged:    "Mirrodin Besieged",
		oldB166Prayers:     "Answered Prayers",
		oldB166Throatseek:  "Throatseeker",
		oldB166Mariner:     "Unsettled Mariner",
		oldB166Therapist:   "Cabal Therapist",
		oldB166Golem:       "Icehide Golem",
		oldB166Evasion:     "Cunning Evasion",
		oldB166Bastion:     "Frostwalk Bastion",
		oldB166Strix:       "Blizzard Strix",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB166CountNamed(g *game.Game, name string, controller uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name && c.Controller == controller {
				n++
			}
		}
	})
	return n
}

// --- Rank Officer -----------------------------------------------------

func TestOldB166RankOfficerMakesAZombieOnlyIfYouDiscard(t *testing.T) {
	for _, discard := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		held := effectstest.HandCard(me, "Held", "Sorcery")
		id := effectstest.HandCardFull(me, "Rank Officer", "Creature — Zombie Soldier", "{3}{B}", oldB166RankOfficer, []string{"B"})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if discard {
			effectstest.AnswerDiscard(t, g, me.ID, held)
		} else {
			effectstest.AnswerDiscard(t, g, me.ID)
		}
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if discard {
			want = 1
		}
		if got := oldB166CountNamed(g, "Zombie", me.ID); got != want {
			t.Errorf("discard=%v: %d Zombie tokens, want %d", discard, got, want)
		}
		if me.Graveyard.Contains(held) != discard {
			t.Errorf("discard=%v: the card is in the graveyard = %v", discard, me.Graveyard.Contains(held))
		}
	}
}

func TestOldB166RankOfficerDrainsByExilingACreatureCardFromYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	officer := effectstest.B12Push(g, me.ID, "Rank Officer", "Creature — Zombie Soldier", oldB166RankOfficer, 3, 1)
	effectstest.B06AddMana(me, "B", "B")
	if err := g.ActivateCatalogAbility(me.ID, officer, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no creature card in the graveyard")
	}
	dead := effectstest.PushGraveyardCreature(g, me.ID, "Dead Bear", "{1}{G}")
	spell := effectstest.PushGraveyardCardTyped(me, "Spent Spell", "Sorcery")
	if err := g.ActivateCatalogAbility(me.ID, officer, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{spell}}); err == nil {
		t.Fatal("a sorcery card paid \"exile a creature card\"")
	}
	opp.Life = 40
	effectstest.B16Activate(t, g, me.ID, officer, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{dead}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 38 {
		t.Errorf("opponent life = %d, want 38", opp.Life)
	}
	if !g.Exile.Contains(dead) {
		t.Error("the creature card was not exiled as the cost")
	}
}

// --- Mirrodin Besieged ------------------------------------------------

func TestOldB166MirrodinBesiegedMirranMakesMyrForArtifactSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastSiege(t, g, "Mirrodin Besieged", oldB166Besieged, "Mirran")
	effectstest.CastCatalogSpell(t, g, "Trinket", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB166CountNamed(g, "Myr", me.ID); got != 1 {
		t.Errorf("%d Myr tokens after an artifact spell, want 1", got)
	}
	effectstest.CastCatalogSpell(t, g, "Plain Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB166CountNamed(g, "Myr", me.ID); got != 1 {
		t.Errorf("%d Myr tokens after a sorcery, want still 1", got)
	}
}

func TestOldB166MirrodinBesiegedPhyrexianOnlyLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastSiege(t, g, "Mirrodin Besieged", oldB166Besieged, "Phyrexian")
	effectstest.CastCatalogSpell(t, g, "Trinket", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB166CountNamed(g, "Myr", me.ID); got != 0 {
		t.Fatalf("the Phyrexian choice made %d Myr", got)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AnswerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand-1 {
		t.Errorf("hand %d -> %d, want one discarded", hand, me.Hand.Size())
	}
	if opp.Eliminated {
		t.Error("the opponent lost with fewer than fifteen artifact cards in my graveyard")
	}
}

func TestOldB166MirrodinBesiegedPhyrexianWinsWithFifteenArtifactCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastSiege(t, g, "Mirrodin Besieged", oldB166Besieged, "Phyrexian")
	for i := 0; i < 15; i++ {
		effectstest.PushGraveyardTyped(me, "Gizmo", "Artifact")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Eliminated {
		t.Error("the targeted opponent did not lose with fifteen artifact cards")
	}
	if me.Eliminated {
		t.Error("I lost")
	}
}

// --- Answered Prayers -------------------------------------------------

func TestOldB166AnsweredPrayersGainsLifeAndBecomesAnAngel(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	prayers := effectstest.B12Push(g, me.ID, "Answered Prayers", "Enchantment", oldB166Prayers, 0, 0)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Bear", oldB166Bear, "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d -> %d, want +1", life, me.Life)
	}
	c := effectstest.LayeredCard(t, g, prayers)
	if !c.IsCreature() || c.CurrentPower() != 3 || c.CurrentToughness() != 3 || !c.HasSubtype("Angel") || !game.HasKeyword(&c, "flying") {
		t.Errorf("Answered Prayers is creature=%v %d/%d angel=%v flying=%v, want a 3/3 Angel creature with flying",
			c.IsCreature(), c.CurrentPower(), c.CurrentToughness(), c.HasSubtype("Angel"), game.HasKeyword(&c, "flying"))
	}
	if !c.HasCardType("enchantment") {
		t.Error("it lost its enchantment type")
	}
	// A second creature while it already is one only gains life.
	effectstest.CastCatalogSpell(t, g, "Bear 2", oldB166Bear, "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life = %d, want %d", me.Life, life+2)
	}
	// An opponent's creature entering does not trigger it.
	effectstest.EnterFromHand(t, g, g.Seats[1].ID, "Their Bear", oldB166Bear, "")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("an opponent's creature gained me life: %d", me.Life)
	}
}

func TestOldB166AnsweredPrayersAngelEndsAtEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	prayers := effectstest.B12Push(g, me.ID, "Answered Prayers", "Enchantment", oldB166Prayers, 0, 0)
	effectstest.CastCatalogSpell(t, g, "Bear", oldB166Bear, "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if c := effectstest.LayeredCard(t, g, prayers); c.IsCreature() {
		t.Error("it was still a creature next turn")
	}
}

// --- Throatseeker -----------------------------------------------------

func TestOldB166ThroatseekerGivesLifelinkToUnblockedAttackingNinjas(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seeker := effectstest.B12Push(g, me.ID, "Throatseeker", "Creature — Vampire Ninja", oldB166Throatseek, 3, 2)
	ninja := effectstest.B12Creature(g, me.ID, "Other Ninja", "Creature — Human Ninja", 2, 5)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB166Bear, 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", oldB166Bear, 1, 5)
	lifelink := func(id uuid.UUID) bool {
		c := effectstest.LayeredCard(t, g, id)
		return game.HasKeyword(&c, "lifelink")
	}
	if lifelink(seeker) {
		t.Fatal("a Ninja that is not attacking has lifelink")
	}
	life := me.Life
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{seeker, ninja, bear} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, ninja); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInCombatDeclaration(t, g)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	// Only the unblocked Throatseeker's 3 damage gains life: the blocked Ninja
	// and the unblocked non-Ninja bear have no lifelink.
	if me.Life != life+3 {
		t.Errorf("life %d -> %d, want +3 from the Throatseeker alone", life, me.Life)
	}
}

// --- Unsettled Mariner ------------------------------------------------

func TestOldB166UnsettledMarinerTaxesOpponentsTargetingYouOrYourPermanents(t *testing.T) {
	for _, tc := range []struct {
		name      string
		pay       bool
		atPlayer  bool
		wantAlive bool
	}{
		{"at its own controller, unpaid", false, true, true},
		{"at a permanent, unpaid", false, false, true},
		{"at a permanent, paid", true, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.B12Push(g, me.ID, "Unsettled Mariner", "Creature — Shapeshifter", oldB166Mariner, 2, 2)
			bear := effectstest.B12Creature(g, me.ID, "Bear", oldB166Bear, 2, 2)
			effectstest.AdvanceToMain(t, g)
			effectstest.B06AddMana(opp, "R", "R", "R")
			target := effectstest.B16TargetCard(bear)
			if tc.atPlayer {
				target = effectstest.B16TargetPlayer(me.ID)
			}
			life := me.Life
			bolt := effectstest.HandCardFull(opp, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle, []string{"R"})
			if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{Targets: target}); err != nil {
				t.Fatal(err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if !effectstest.HasPayUnlessFor(g, opp.ID) {
				t.Fatal("the opponent was not asked to pay {1}")
			}
			effectstest.AnswerPayUnless(t, g, opp.ID, tc.pay)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.BattlefieldHas(g, bear); !tc.atPlayer && got != tc.wantAlive {
				t.Errorf("Bear alive = %v, want %v", got, tc.wantAlive)
			}
			if tc.atPlayer && me.Life != life {
				t.Errorf("life %d -> %d, the Bolt should have been countered", life, me.Life)
			}
		})
	}
}

func TestOldB166UnsettledMarinerIgnoresYourOwnSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Unsettled Mariner", "Creature — Shapeshifter", oldB166Mariner, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB166Bear, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetCard(bear))
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("the Mariner taxed its own controller")
	}
}

// oldB166BeginPrecombatMain emits the begin-of-precombat-main event for a seat,
// as Muerra's and Ripples of Undeath's tests do, and settles the stack.
func oldB166BeginPrecombatMain(t *testing.T, g *game.Game, who uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventBeginPrecombatMain, Actor: who}) })
	effectstest.PassPriorityAroundTable(t, g)
}

// --- Cabal Therapist --------------------------------------------------

func TestOldB166CabalTherapistNamesACardAndTheTargetDiscardsEveryCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	therapist := effectstest.B12Push(g, me.ID, "Cabal Therapist", "Creature — Horror", oldB166Therapist, 1, 1)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", oldB166Bear, 1, 1)
	opp.Hand.Cards = nil
	one := effectstest.HandCard(opp, "Counterspell", "Instant")
	two := effectstest.HandCard(opp, "Counterspell", "Instant")
	other := effectstest.HandCard(opp, "Opt", "Instant")
	land := effectstest.HandCard(opp, "Island", "Basic Land — Island")
	oldB166BeginPrecombatMain(t, g, me.ID)
	effectstest.AnswerMayCounter(t, g, me.ID, fodder)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ActNameFor(t, g, me.ID, "counterspell")
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(one) || !opp.Graveyard.Contains(two) {
		t.Error("not every card with the named name was discarded")
	}
	if !opp.Hand.Contains(other) || !opp.Hand.Contains(land) {
		t.Error("a card with another name was discarded")
	}
	if !effectstest.BattlefieldHas(g, therapist) || effectstest.BattlefieldHas(g, fodder) {
		t.Error("the sacrificed creature is still there, or the Therapist is gone")
	}
}

func TestOldB166CabalTherapistNamingALandDiscardsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Cabal Therapist", "Creature — Horror", oldB166Therapist, 1, 1)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", oldB166Bear, 1, 1)
	opp.Hand.Cards = nil
	land := effectstest.HandCard(opp, "Island", "Basic Land — Island")
	oldB166BeginPrecombatMain(t, g, me.ID)
	effectstest.AnswerMayCounter(t, g, me.ID, fodder)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ActNameFor(t, g, me.ID, "Island")
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(land) {
		t.Error("a land was discarded for a name that is not a nonland card name")
	}
}

func TestOldB166CabalTherapistWithoutASacrificeDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Cabal Therapist", "Creature — Horror", oldB166Therapist, 1, 1)
	opp.Hand.Cards = nil
	card := effectstest.HandCard(opp, "Opt", "Instant")
	oldB166BeginPrecombatMain(t, g, me.ID)
	effectstest.AnswerMayCounter(t, g, me.ID, uuid.Nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(card) || effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a reflexive trigger followed a declined sacrifice")
	}
}

// --- Icehide Golem ----------------------------------------------------

func TestOldB166IcehideGolemNeedsSnowMana(t *testing.T) {
	for _, snow := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.ToMainForCost(t, g)
		if snow {
			effectstest.BlbB1Push(g, me.ID, "Snow-Covered Island", "Basic Snow Land — Island", "", 0, 0)
		} else {
			effectstest.BlbB1Push(g, me.ID, "Island", "Basic Land — Island", "", 0, 0)
		}
		id := effectstest.HandCardFull(me, "Icehide Golem", "Snow Artifact Creature — Golem", "{S}", oldB166Golem, nil)
		err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AutoTap: true})
		if (err == nil) != snow {
			t.Errorf("snow=%v: cast error = %v", snow, err)
		}
		if snow {
			effectstest.PassPriorityAroundTable(t, g)
			if !effectstest.BattlefieldHas(g, id) {
				t.Error("the Golem did not resolve")
			}
		}
	}
}

// --- Cunning Evasion --------------------------------------------------

func TestOldB166CunningEvasionBouncesABlockedCreatureOfYours(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.B12Push(g, me.ID, "Cunning Evasion", "Enchantment", oldB166Evasion, 0, 0)
		attacker := effectstest.B12Creature(g, me.ID, "Attacker", oldB166Bear, 2, 2)
		blocker := effectstest.B12Creature(g, opp.ID, "Blocker", oldB166Bear, 2, 2)
		effectstest.Pr7aBlockedCombat(t, g, opp.ID, attacker, blocker)
		effectstest.LockInCombatDeclaration(t, g)
		effectstest.AssistCardSettle(t, g, me.ID, nil, yes)
		if got := me.Hand.Contains(attacker); got != yes {
			t.Errorf("answered %v: the attacker is in my hand = %v", yes, got)
		}
	}
}

// --- Frostwalk Bastion ------------------------------------------------

func TestOldB166FrostwalkBastionAnimatesWithSnowManaAndFreezesWhatItFights(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bastion := effectstest.B12Push(g, me.ID, "Frostwalk Bastion", "Snow Land", oldB166Bastion, 0, 0)
	effectstest.BlbB1Push(g, me.ID, "Snow-Covered Island", "Basic Snow Land — Island", "", 0, 0)
	effectstest.BlbB1Push(g, me.ID, "Snow-Covered Forest", "Basic Snow Land — Forest", "", 0, 0)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", oldB166Bear, 2, 5)
	effectstest.B16Activate(t, g, me.ID, bastion, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, bastion)
	if !c.IsCreature() || !c.IsLand() || !c.HasCardType("artifact") || c.CurrentPower() != 2 || c.CurrentToughness() != 3 || !c.HasSubtype("Construct") {
		t.Fatalf("Bastion is creature=%v land=%v artifact=%v %d/%d", c.IsCreature(), c.IsLand(), c.HasCardType("artifact"), c.CurrentPower(), c.CurrentToughness())
	}
	effectstest.DealCombatDamage(g, me.ID, bastion, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HmlTapped(t, g, victim) {
		t.Error("the creature it hit is not tapped")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceTo(t, g, game.StepUpkeep)
	if !effectstest.HmlTapped(t, g, victim) {
		t.Error("it untapped during its controller's next untap step")
	}
}

func TestOldB166FrostwalkBastionNeedsSnowManaToAnimate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bastion := effectstest.B12Push(g, me.ID, "Frostwalk Bastion", "Snow Land", oldB166Bastion, 0, 0)
	// The Bastion is itself a snow source, so it is tapped out of the way and
	// two mana that is not snow float instead.
	effectstest.AssistCardTapped(t, g, bastion)
	effectstest.B06AddMana(me, "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, bastion, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("animated with no snow mana")
	}
}

// --- Blizzard Strix ---------------------------------------------------

func TestOldB166BlizzardStrixBlinksATargetOnlyWithAnotherSnowPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB166Bear, 2, 2)
	effectstest.ToMainForCost(t, g)
	// No other snow permanent: nothing happens, and no target is asked.
	effectstest.CastCatalogSpell(t, g, "Blizzard Strix", "Snow Creature — Bird", oldB166Strix, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || !effectstest.BattlefieldHas(g, bear) {
		t.Fatal("the Strix triggered with no other snow permanent")
	}
	effectstest.B12Permanent(g, me.ID, "Snow-Covered Island", "Basic Snow Land — Island")
	effectstest.CastCatalogSpell(t, g, "Blizzard Strix", "Snow Creature — Bird", oldB166Strix, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB166CountNamed(g, "Bear", opp.ID) != 0 || !g.Exile.Contains(bear) {
		t.Fatal("the target was not exiled")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB166CountNamed(g, "Bear", opp.ID); got != 1 {
		t.Errorf("%d Bears under its owner's control at the end step, want 1", got)
	}
}
