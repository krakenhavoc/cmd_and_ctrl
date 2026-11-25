package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b099_cards_test.go — Jumpstart, S82 older sets batch 099 (#5510):
// damage cards.

const (
	jmpCompanion = "f3264224-43a4-4c6b-b363-2d147687dfc1"
	jmpSpitting  = "0c3bf4e1-d91e-4dd3-a800-e40971222c71"
	jmpRoar      = "843af5d3-447e-40e4-a562-f5a7ed09ca3e"
	jmpFlames    = "da8dd2ee-0ce3-4ce6-abb9-7444cefcbe83"
	jmpHound     = "d40fd63b-f55d-49f6-b4e5-a8d00f22e8e1"
	jmpProdder   = "745efa76-7063-4971-9d1b-b2b4cb336a57"
	jmpHellion   = "68c88a4c-8c2b-4e6f-b835-7e269eb073c8"
	jmpHeartfire = "2f6f5054-4a48-458e-969e-3a0f0e507354"
	jmpExcavator = "cdaf3711-5bec-41bf-b333-c5f34ccb12ef"
	jmpGyre      = "1888b473-e26a-4882-ae18-5a0a6caee410"
)

func TestS82OldB099EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		jmpCompanion: "Initiate's Companion", jmpSpitting: "Spitting Earth", jmpRoar: "Draconic Roar",
		jmpFlames: "Flames of the Raze-Boar", jmpHound: "Ashmouth Hound", jmpProdder: "Sin Prodder",
		jmpHellion: "Inferno Hellion", jmpHeartfire: "Heartfire", jmpExcavator: "Lightning-Core Excavator",
		jmpGyre: "Immolating Gyre",
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

func jmpMountain(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Mountain", TypeLine: "Basic Land — Mountain", Owner: owner, Controller: owner})
}

func TestS82OldB099InitiatesCompanionUntapsACreatureOrLandOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cat := effectstest.B12Push(g, me.ID, "Initiate's Companion", "Creature — Cat", jmpCompanion, 3, 1)
	land := jmp99Forest(g, me.ID)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, land).Tapped = true })
	life := opp.Life
	effectstest.AttackWith(t, g, opp.ID, cat)
	jmp99Settle(t, g, me.ID, land)
	if opp.Life != life-3 {
		t.Fatalf("the Companion dealt %d damage, want 3", life-opp.Life)
	}
	if effectstest.B12Card(t, g, land).Tapped {
		t.Error("the targeted land is still tapped")
	}
}

func TestS82OldB099SpittingEarthDealsDamageEqualToYourMountains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	jmpMountain(g, me.ID)
	jmpMountain(g, me.ID)
	jmpMountain(g, opp.ID)
	jmp99Forest(g, me.ID)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Ogre", 5, 5)
	effectstest.CastCatalogSpell(t, g, "Spitting Earth", "Sorcery", jmpSpitting, effectstest.CardRefs(big))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Pr6Marked(g, big); got != 2 {
		t.Errorf("the creature has %d damage, want 2 (your two Mountains, not theirs or your Forest)", got)
	}
}

func TestS82OldB099HeartfireNeedsASacrificeAndDealsFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	card := game.Card{InstanceID: uuid.New(), Name: "Heartfire", TypeLine: "Instant", OracleID: jmpHeartfire, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(card)
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Strict: true, Targets: target}); err == nil {
		t.Fatal("Heartfire was cast with nothing to sacrifice")
	}
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	life := opp.Life
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Targets: target, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed")
	}
	if opp.Life != life-4 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-4)
	}
}

func TestS82OldB099LightningCoreExcavatorSacrificesForThreeDamage(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	ex := effectstest.PushCatalogPermanent(g, me.ID, "Lightning-Core Excavator", "Artifact Creature — Golem", jmpExcavator, false)
	effectstest.AddPool(t, g, me.ID, "{C}{C}{C}{C}{C}")
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, ex, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(ex) {
		t.Error("the Excavator is still on the battlefield after paying the sacrifice cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-3)
	}
}

func TestS82OldB099LightningCoreExcavatorNeedsFiveMana(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	ex := effectstest.PushCatalogPermanent(g, me.ID, "Lightning-Core Excavator", "Artifact Creature — Golem", jmpExcavator, false)
	effectstest.AddPool(t, g, me.ID, "{C}{C}{C}{C}")
	err := g.ActivateCatalogAbility(me.ID, ex, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.RefusedForMana(t, err, "four mana for the Excavator")
}

func TestS82OldB099ImmolatingGyreHitsOnlyWhatYouDontControlForTheSpellsInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	for _, tl := range []string{"Instant", "Sorcery", "Instant", "Creature — Bear"} {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: tl, Owner: me.ID, Controller: me.ID})
	}
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 3)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Ogre", 2, 4)
	walker := effectstest.PushWalkerForTest(g, opp.ID, "Test Walker", "", 5)
	effectstest.CastCatalogSpell(t, g, "Immolating Gyre", "Sorcery", jmpGyre, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("a 3-toughness creature survived X=3")
	}
	if got := effectstest.Pr6Marked(g, big); got != 3 {
		t.Errorf("the 4-toughness creature has %d damage, want 3", got)
	}
	if got := effectstest.LoyaltyOf(g, walker); got != 2 {
		t.Errorf("the planeswalker has %d loyalty, want 2", got)
	}
	if got := effectstest.Pr6Marked(g, mine); got != 0 {
		t.Errorf("your own creature has %d damage, want 0", got)
	}
	if opp.Life != 40 {
		t.Errorf("a player was damaged: life %d", opp.Life)
	}
}

func TestS82OldB099AshmouthHoundPokesWhatItBlocksOrIsBlockedBy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hound := effectstest.B12Push(g, opp.ID, "Ashmouth Hound", "Creature — Elemental Dog", jmpHound, 2, 1)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 4)
	bystander := effectstest.B12Creature(g, me.ID, "Bystander", "Creature — Bear", 2, 4)
	effectstest.WoeS2Block(g, opp.ID, hound, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Pr6Marked(g, attacker); got != 1 {
		t.Errorf("the blocked creature has %d damage, want 1", got)
	}
	if got := effectstest.Pr6Marked(g, bystander); got != 0 {
		t.Errorf("a creature outside the block has %d damage, want 0", got)
	}
	mine := effectstest.B12Push(g, me.ID, "Ashmouth Hound", "Creature — Elemental Dog", jmpHound, 2, 1)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 4)
	effectstest.WoeS2Block(g, opp.ID, blocker, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Pr6Marked(g, blocker); got != 1 {
		t.Errorf("a creature blocking the Hound has %d damage, want 1", got)
	}
}

func TestS82OldB099InfernoHellionShufflesItselfAwayAfterAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hellion := effectstest.B12Push(g, me.ID, "Inferno Hellion", "Creature — Hellion", jmpHellion, 7, 3)
	lib := me.Library.Size()
	effectstest.AttackWith(t, g, opp.ID, hellion)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(hellion) {
		t.Error("the Hellion that attacked is still on the battlefield")
	}
	if me.Library.Size() != lib+1 {
		t.Errorf("library %d, want %d (the Hellion shuffled in)", me.Library.Size(), lib+1)
	}
}

func TestS82OldB099InfernoHellionShufflesItselfAwayAfterBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	hellion := effectstest.B12Push(g, opp.ID, "Inferno Hellion", "Creature — Hellion", jmpHellion, 7, 3)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 1, 1)
	effectstest.WoeS2Block(g, opp.ID, hellion, attacker)
	lib := opp.Library.Size()
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(hellion) || opp.Library.Size() != lib+1 {
		t.Errorf("the blocking Hellion was not shuffled into its owner's library (library %d, was %d)", opp.Library.Size(), lib)
	}
}

func TestS82OldB099InfernoHellionStaysIfItDidNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hellion := effectstest.B12Push(g, me.ID, "Inferno Hellion", "Creature — Hellion", jmpHellion, 7, 3)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(hellion) {
		t.Error("the Hellion left without having attacked or blocked")
	}
}

func TestS82OldB099FlamesOfTheRazeBoarSplashesOnlyWithAPowerFourCreature(t *testing.T) {
	for _, bigger := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		power := 3
		if bigger {
			power = 4
		}
		effectstest.B12Creature(g, me.ID, "Mine", "Creature — Ogre", power, 4)
		mineToo := effectstest.B12Creature(g, me.ID, "Mine too", "Creature — Bear", 1, 5)
		target := effectstest.B12Creature(g, opp.ID, "Target", "Creature — Ogre", 1, 9)
		other := effectstest.B12Creature(g, opp.ID, "Other", "Creature — Bear", 1, 9)
		third := effectstest.B12Creature(g, g.Seats[2].ID, "Third", "Creature — Bear", 1, 9)
		effectstest.CastCatalogSpell(t, g, "Flames of the Raze-Boar", "Instant", jmpFlames, effectstest.CardRefs(target))
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.Pr6Marked(g, target); got != 4 {
			t.Errorf("bigger=%v: the target has %d damage, want 4", bigger, got)
		}
		wantOther := 0
		if bigger {
			wantOther = 2
		}
		if got := effectstest.Pr6Marked(g, other); got != wantOther {
			t.Errorf("bigger=%v: the target's controller's other creature has %d damage, want %d", bigger, got, wantOther)
		}
		for _, id := range []uuid.UUID{mineToo, third} {
			if got := effectstest.Pr6Marked(g, id); got != 0 {
				t.Errorf("bigger=%v: an uninvolved creature has %d damage", bigger, got)
			}
		}
	}
}

func TestS82OldB099FlamesOfTheRazeBoarTargetsOnlyAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	card := game.Card{InstanceID: uuid.New(), Name: "Flames of the Raze-Boar", TypeLine: "Instant", OracleID: jmpFlames, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(card)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(mine)}); err == nil {
		t.Error("the spell was cast at your own creature")
	}
}

func jmpRoarCast(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID, params game.CastSpellParams) {
	t.Helper()
	card := game.Card{InstanceID: uuid.New(), Name: "Draconic Roar", TypeLine: "Instant", OracleID: jmpRoar, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(card)
	params.Targets = effectstest.CardRefs(target)
	if err := g.CastSpell(me.ID, card.InstanceID, params); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestS82OldB099DraconicRoarHurtsTheControllerOnlyWithADragon(t *testing.T) {
	for _, tc := range []string{"nothing", "controlled", "revealed"} {
		t.Run(tc, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Ogre", 1, 9)
			params := game.CastSpellParams{}
			switch tc {
			case "controlled":
				effectstest.B12Creature(g, me.ID, "Dragon", "Creature — Dragon", 4, 4)
			case "revealed":
				dragon := effectstest.PutInHand(me, game.Card{Name: "Hand Dragon", TypeLine: "Creature — Dragon", Power: 4, Toughness: 4})
				params = game.CastSpellParams{OptionalCosts: []int{0}, RevealIDs: []uuid.UUID{dragon}}
			}
			life := opp.Life
			jmpRoarCast(t, g, me, victim, params)
			if got := effectstest.Pr6Marked(g, victim); got != 3 {
				t.Errorf("the creature has %d damage, want 3", got)
			}
			want := life
			if tc != "nothing" {
				want = life - 3
			}
			if opp.Life != want {
				t.Errorf("%s: the creature's controller is at %d, want %d", tc, opp.Life, want)
			}
		})
	}
}

func TestS82OldB099SinProdderGivesYouTheCardOrHitsTheOpponentWhoTakesIt(t *testing.T) {
	for _, take := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.B12Push(g, me.ID, "Sin Prodder", "Creature — Devil", jmpProdder, 3, 2)
		top := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: top, Name: "Top", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Owner: me.ID, Controller: me.ID})
		hand, gy, life := me.Hand.Size(), me.Graveyard.Size(), opp.Life
		effectstest.AppToUpkeepOf(t, g, 0)
		effectstest.PassPriorityAroundTable(t, g)
		who := uuid.Nil
		if take {
			who = opp.ID
		}
		asked := effectstest.AppAnswerWith(t, g, effectstest.AppOnly(who))
		for _, id := range asked {
			if id == me.ID {
				t.Errorf("take=%v: the controller was asked", take)
			}
		}
		if take {
			if !me.Graveyard.Contains(top) || me.Graveyard.Size() != gy+1 {
				t.Errorf("the card did not go to the graveyard")
			}
			if opp.Life != life-3 {
				t.Errorf("the opponent is at %d, want %d (the card's mana value)", opp.Life, life-3)
			}
			continue
		}
		if !me.Hand.Contains(top) || me.Hand.Size() != hand+1 {
			t.Errorf("the card did not go to your hand")
		}
		if opp.Life != life {
			t.Errorf("the opponent lost life: %d", opp.Life)
		}
	}
}

func jmp99Forest(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: owner, Controller: owner})
}

// jmp99Settle passes priority around the table, answering a target pick with
// `want` when the trigger asks for one, then settles the rest.
func jmp99Settle(t *testing.T, g *game.Game, chooser uuid.UUID, want uuid.UUID) {
	t.Helper()
	effectstest.PassPriorityAroundTable(t, g)
	if want != uuid.Nil && effectstest.LatestPickTarget(g, chooser) != nil {
		effectstest.PickCard(t, g, chooser, want)
		effectstest.PassPriorityAroundTable(t, g)
	}
}
