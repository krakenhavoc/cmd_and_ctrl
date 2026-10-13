package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Caveat slice cav-g1 (#3172): tests for the cards whose caveat this slice
// cleared.

// --- Alchemist's Talent level 3 -------------------------------------

// "Whenever you cast a spell, if mana from a Treasure was spent to cast
// it, this Class deals damage equal to that spell's mana value to each
// opponent." The Treasure is gone before the trigger is judged, so the
// condition reads the snapshot (#1212). A spell paid with lands, and a
// Class still at level 2, deal nothing.
func TestAlchemistsTalentLevelThreeDamagesForTreasureMana(t *testing.T) {
	for _, tc := range []struct {
		name     string
		level3   bool
		treasure bool
		damage   int
	}{
		{"level 3, Treasure mana", true, true, 2},
		{"level 3, lands only", true, false, 0},
		{"level 2, Treasure mana", false, true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			id := pushClass(g, me.ID, "Alchemist's Talent", "Enchantment — Class", alchemistsTalentOracle)
			advanceToMain(t, g)
			if err := levelUpTo(t, g, me.ID, id, 0, "{1}{R}"); err != nil {
				t.Fatalf("level 2: %v", err)
			}
			if tc.level3 {
				if err := levelUpTo(t, g, me.ID, id, 1, "{4}{R}"); err != nil {
					t.Fatalf("level 3: %v", err)
				}
			}
			if tc.treasure {
				crackATreasureFor(t, g, me, "B")
				floatForTest(g, me, "B")
			} else {
				floatForTest(g, me, "BB")
			}
			var before []int
			for _, s := range g.Seats {
				before = append(before, s.Life)
			}

			castFromHandForTest(t, g, me, "Bear Cub", "Creature — Bear",
				"{1}{B}", "", game.CastSpellParams{Strict: true})
			passPriorityAroundTable(t, g)

			for i, s := range g.Seats {
				want := before[i]
				if s.ID != me.ID {
					want -= tc.damage
				}
				if s.Life != want {
					t.Errorf("seat %d life = %d, want %d", i, s.Life, want)
				}
			}
		})
	}
}

// --- Ayara, First of Locthwain ---------------------------------------

// The caveat said a black token only counts if its template records its
// colour, and named the Bitterblossom Faerie Rogue. The catalog's token
// table has recorded colours since #1127 (a test pins the table against
// each key's colour words), so a black Faerie Rogue drains on entering
// and is a legal sacrifice.
func TestAyaraCountsABlackFaerieRogueToken(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ayara := pushCatalogPermanent(g, me.ID, "Ayara, First of Locthwain", "Legendary Creature — Elf Noble", b07AyaraOracle, false)
	before := lifeOfOpponents(g)
	meBefore := me.Life

	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, FaerieRogueToken(), 1)
	})
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("opponent %d after a black Faerie Rogue: %d -> %d, want -1", i+1, b, got)
		}
	}
	if me.Life != meBefore+1 {
		t.Errorf("Ayara's controller %d -> %d, want +1", meBefore, me.Life)
	}

	rogue := findBattlefieldByName(g, "Faerie Rogue")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, ayara, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{rogue},
	}); err != nil {
		t.Fatalf("sacrificing the black token: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

// --- Firebending: "this mana lasts until end of combat" -------------

// The mark keeps the mana through every step of the combat phase and it
// is gone at the first boundary outside it. A control mana without the
// mark empties at the next step as usual.
func TestKeepManaUntilEndOfCombatLastsThroughCombatOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	mkAdd(t, g, me, "{R}{R}", game.AddManaOptions{Riders: []game.ManaSpendRider{KeepManaUntilEndOfCombat()}})
	mkAdd(t, g, me, "{G}", game.AddManaOptions{})

	mkAdvanceTo(t, g, game.StepBeginCombat)
	mkWantPool(t, me, map[string]int{"R": 2})
	mkAdvanceTo(t, g, game.StepEndCombat)
	mkWantPool(t, me, map[string]int{"R": 2})
	mkAdvanceTo(t, g, game.StepPostcombatMain)
	mkWantPool(t, me, map[string]int{})
}

// Avatar Roku's Dragon adds four red on attacking that is still there in
// the end of combat step and empties by the postcombat main phase.
func TestFirebendingManaLastsUntilEndOfCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := pushToken(g, me.ID, firebendingDragonToken())
	declareAttack(t, g, opp.ID, dragon)
	passPriorityAroundTable(t, g)
	mkWantPool(t, me, map[string]int{"R": 4})
	mkAdvanceTo(t, g, game.StepEndCombat)
	mkWantPool(t, me, map[string]int{"R": 4})
	mkAdvanceTo(t, g, game.StepPostcombatMain)
	mkWantPool(t, me, map[string]int{})
}

// --- Benevolent Hydra ------------------------------------------------

// "Another target creature you control": a second Benevolent Hydra is a
// legal target (the exclusion is by instance, not by name), and the first
// Hydra's replacement makes the placement two counters.
func TestBenevolentHydraMovesACounterOntoAnotherHydra(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hydra := pushCatalogPermanent(g, me.ID, "Benevolent Hydra", "Creature — Hydra", ccBenevolentHydra, false)
	setCounters(g, hydra, map[string]int{game.CounterPlusOne: 3})
	second := pushCatalogPermanent(g, me.ID, "Benevolent Hydra", "Creature — Hydra", ccBenevolentHydra, false)
	advanceToMain(t, g)

	b16Activate(t, g, me.ID, hydra, 0, game.ActivateAbilityParams{Targets: b16TargetCard(second)})
	if got := counterCount(g, hydra, game.CounterPlusOne); got != 2 {
		t.Errorf("the first Hydra has %d counters, want 2", got)
	}
	if got := counterCount(g, second, game.CounterPlusOne); got != 2 {
		t.Errorf("the second Hydra has %d counters, want 2 (one, plus the first Hydra's one)", got)
	}
}

// --- Beamtown Beatstick ----------------------------------------------

// "Whenever equipped creature deals combat damage to a player or battle,
// create a Treasure token." The battle half: an equipped attacker
// hitting a battle makes the Treasure.
func TestBeamtownBeatstickMakesATreasureForCombatDamageToABattle(t *testing.T) {
	g := newCatalogGame(t)
	me, owner, protector := g.Seats[0], g.Seats[1], g.Seats[2]
	beatstick := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Beamtown Beatstick", TypeLine: "Artifact — Equipment",
		OracleID: "035e33ab-adfe-4831-bdbc-52f4ddc7ffe8", Owner: me.ID, Controller: me.ID,
	})
	attacker := pushAnnihilatorCreature(g, me.ID)
	battle := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Siege", TypeLine: "Battle — Siege",
		Owner: owner.ID, Controller: owner.ID, ProtectorPlayerID: protector.ID,
		Counters: map[string]int{game.CounterDefense: 9},
	})
	advanceToMain(t, g)
	equipTo(t, g, me.ID, beatstick, attacker)

	declareAttack(t, g, battle, attacker)
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)

	if findBattlefieldByName(g, "Treasure") == uuid.Nil {
		t.Error("combat damage to a battle made no Treasure")
	}
}

// --- Buster Sword ------------------------------------------------------

// "Whenever equipped creature deals combat damage to a player, draw a
// card, then you may cast a spell from your hand with mana value less
// than or equal to that damage without paying its mana cost." The creature
// is 2/2 with the sword's +3/+2, so five damage lands: a mana value 5
// spell is offered, a mana value 6 spell and a land are not.
func TestBusterSwordDrawsThenOffersAFreeCastCappedByTheDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sword := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Buster Sword", TypeLine: "Artifact — Equipment",
		OracleID: "5e060d58-4d6e-425c-b7d4-727669fcce5b", Owner: me.ID, Controller: me.ID,
	})
	attacker := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	fits, tooBig, land := uuid.New(), uuid.New(), uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: fits, Name: "Fits", TypeLine: "Sorcery", ManaCost: "{4}{R}", Owner: me.ID, Controller: me.ID})
	me.Hand.PushTop(game.Card{InstanceID: tooBig, Name: "Too Big", TypeLine: "Sorcery", ManaCost: "{5}{R}", Owner: me.ID, Controller: me.ID})
	me.Hand.PushTop(game.Card{InstanceID: land, Name: "Test Land", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	equipTo(t, g, me.ID, sword, attacker)
	handBefore := me.Hand.Size()
	lifeBefore := opp.Life

	declareAttack(t, g, opp.ID, attacker)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)

	if opp.Life != lifeBefore-5 {
		t.Fatalf("opponent life %d -> %d, want five damage", lifeBefore, opp.Life)
	}
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatalf("no free-cast pick was offered: %+v", g.PendingChoices)
	}
	if !hasID(pick.ChooseCards, fits) || hasID(pick.ChooseCards, tooBig) || hasID(pick.ChooseCards, land) {
		t.Errorf("offered %v, want only the mana value 5 spell", pick.ChooseCards)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{fits}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if perm := g.CastPermissionOnCardByIDForEffect(fits); perm == nil || perm.Cost != "{0}" {
		t.Errorf("the chosen card must carry a free-cast permission: %+v", perm)
	}
}

// --- Commander's Plate ---------------------------------------------------

// "Equipped creature gets +3/+3 and has protection from each color that's
// not in your commander's color identity." With a white-blue commander the
// wearer gets protection from black, red and green; with no commander the
// identity is colourless and it gets all five.
func TestCommandersPlateGrantsProtectionFromColorsOutsideTheIdentity(t *testing.T) {
	for _, tc := range []struct {
		name     string
		identity []string // nil: no commander at all
		want     []string
		not      []string
	}{
		{"white-blue commander", []string{"W", "U"}, []string{"black", "red", "green"}, []string{"white", "blue"}},
		{"no commander", nil, []string{"white", "blue", "black", "red", "green"}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			if tc.identity != nil {
				pushBattlefieldCardWithTimestamp(g, game.Card{
					InstanceID: uuid.New(), Name: "General", TypeLine: "Legendary Creature — Human Soldier",
					Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, IsCommander: true,
					ColorIdentity: tc.identity,
				})
			}
			advanceToMain(t, g)
			bear := seedBear(g, me.ID)
			plate := seedEquipment(g, me.ID, "Commander's Plate", commandersPlateOracle)
			if hasEffectiveKeyword(t, g, bear, "protection from black") {
				t.Fatal("protection before equipping")
			}
			equipTo(t, g, me.ID, plate, bear)
			for _, colour := range tc.want {
				if !hasEffectiveKeyword(t, g, bear, "protection from "+colour) {
					t.Errorf("missing protection from %s", colour)
				}
			}
			for _, colour := range tc.not {
				if hasEffectiveKeyword(t, g, bear, "protection from "+colour) {
					t.Errorf("protection from %s is in the commander's identity", colour)
				}
			}
		})
	}
}

// --- Command the Stage -----------------------------------------------------

// "At the beginning of each upkeep, if an opponent was dealt noncombat
// damage last turn, return this card from your graveyard to your hand."
// The card returns at the next upkeep after the damage, and not when
// nobody was hurt.
func TestCommandTheStageReturnsAfterNoncombatDamageToAnOpponent(t *testing.T) {
	for _, tc := range []struct {
		name     string
		damage   bool
		returned bool
	}{
		{"noncombat damage last turn", true, true},
		{"no damage last turn", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			stage := uuid.New()
			me.Graveyard.PushTop(game.Card{
				InstanceID: stage, Name: "Command the Stage", TypeLine: "Sorcery",
				OracleID: "13bfc51f-d079-40ef-a4ba-47d06d3150a3", Owner: me.ID, Controller: me.ID,
			})
			source := pushVanillaCreature(g, me.ID, "Pinger", 2, 2)
			advanceToMain(t, g)
			if tc.damage {
				g.WithWriteLock(func() {
					_ = g.DealDamageToPlayerForEffect(source, opp.ID, 2)
				})
			}
			// Roll into the next turn's upkeep: the trigger is watched
			// from the graveyard, and on it the card goes to hand.
			next := (g.Turn.ActiveSeat + 1) % len(g.Seats)
			for i := 0; i < 40 && !(g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat == next); i++ {
				if _, err := g.AdvanceStep(); err != nil {
					t.Fatalf("AdvanceStep: %v", err)
				}
			}
			passPriorityAroundTable(t, g)
			if got := me.Hand.Contains(stage); got != tc.returned {
				t.Errorf("returned to hand = %v, want %v", got, tc.returned)
			}
			if got := me.Graveyard.Contains(stage); got == tc.returned {
				t.Errorf("still in graveyard = %v, want %v", got, !tc.returned)
			}
		})
	}
}

// --- Creeping Peeper -------------------------------------------------------

// "Spend this mana only to cast an enchantment spell, unlock a door, or
// turn a permanent face up." Peeper's {U} pays half of Willbender's {1}{U}
// face-up cost; the same restricted mana does not pay it when the
// restriction is only the cast clause.
func TestCreepingPeeperManaTurnsAPermanentFaceUp(t *testing.T) {
	const peeperOracle = "acf22340-7ff9-443f-a846-c44bb5a3f4e7"
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	id := castWithAltCost(t, g, "Willbender", "Creature — Human Wizard", willbenderOracle, "morph")
	passPriorityAroundTable(t, g)
	if perm := findCardAnywhere(t, g, id); !perm.FaceDownIsPermanent() {
		t.Fatal("the Willbender did not enter face down")
	}

	peeper := pushCatalogPermanent(g, active.ID, "Creeping Peeper", "Creature — Eye", peeperOracle, false)
	if err := g.ActivateManaAbility(active.ID, peeper, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Peeper: %v", err)
	}
	if len(active.ManaPool) != 1 || len(active.ManaPool[0].Restrictions) == 0 {
		t.Fatalf("the Peeper's mana is not restricted: %+v", active.ManaPool)
	}

	// Cast-only restricted mana of the same colour must not pay for it.
	g.WithWriteLock(func() {
		active.ManaPool = game.ManaPool{
			{Color: "U", Restrictions: []string{game.ManaRestrictCast}},
			{Color: "C"},
		}
	})
	if err := g.PerformSpecialAction(active.ID, id, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true}); err == nil {
		t.Fatal("mana restricted to casting turned a permanent face up")
	}

	// The Peeper's own mana does.
	g.WithWriteLock(func() {
		active.ManaPool = game.ManaPool{}
	})
	peeper2 := pushCatalogPermanent(g, active.ID, "Creeping Peeper", "Creature — Eye", peeperOracle, false)
	if err := g.ActivateManaAbility(active.ID, peeper2, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the second Peeper: %v", err)
	}
	g.WithWriteLock(func() { active.ManaPool.AddMana(manaTokens("C")...) })
	if err := g.PerformSpecialAction(active.ID, id, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn face up with the Peeper's mana: %v", err)
	}
	if up := findCardAnywhere(t, g, id); up.FaceDown {
		t.Error("still face down")
	}
}

// --- Doppelgang and Electroduplicate: token copies run the as-enters clause ---

const automatonOracle = "53c730c6-2f8c-4af8-b400-b9d573a71e60"

// Both caveats said a token copy skips an on-enter hook. Since the token
// runs the ordinary entry pipeline (#762) the copied card's "as enters,
// choose a creature type" prompt reaches the copy.
func TestDoppelgangTokenCopyAsksForTheCopiedCardsEntryChoice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	original := pushTokenReplacementCard(g, automatonOracle, "Adaptive Automaton", "Artifact Creature — Construct", me.ID)

	castXSpell(t, g, "Doppelgang", "Sorcery", b28DoppelgangOracle, "{X}{X}{X}{G}{U}", 1, []game.TargetRef{
		{Kind: game.TargetCard, ID: original},
	})
	passPriorityAroundTable(t, g)

	if got := b28TokensNamed(g, me.ID, "Adaptive Automaton"); got != 1 {
		t.Fatalf("token copies of the Automaton: %d, want 1", got)
	}
	if len(g.PendingChoices) == 0 {
		t.Error("the token copy did not ask for its creature type")
	}
}

func TestElectroduplicateTokenCopyAsksForTheCopiedCardsEntryChoice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	original := pushTokenReplacementCard(g, automatonOracle, "Adaptive Automaton", "Artifact Creature — Construct", me.ID)

	castCatalogSpell(t, g, "Electroduplicate", "Sorcery", "112f2b3f-32e7-40b1-b80e-0b99184a840c", b16TargetCard(original))
	passPriorityAroundTable(t, g)

	if got := b28TokensNamed(g, me.ID, "Adaptive Automaton"); got != 1 {
		t.Fatalf("token copies of the Automaton: %d, want 1", got)
	}
	if len(g.PendingChoices) == 0 {
		t.Error("the token copy did not ask for its creature type")
	}
}

// --- Esika, God of the Tree --------------------------------------------------

// "Other legendary creatures you control have vigilance and '{T}: Add one
// mana of any color.'" Another legendary creature you control gets both; a
// nonlegendary creature, an opponent's legendary creature and Esika's own
// (printed) mana ability are not a grant.
func TestEsikaGrantsOtherLegendaryCreaturesVigilanceAndMana(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Esika, God of the Tree", "Legendary Creature — God", esikaGodOfTheTreeOracleID, false)
	general := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "General", TypeLine: "Legendary Creature — Human Soldier",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their General", TypeLine: "Legendary Creature — Human Soldier",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})

	if !hasEffectiveKeyword(t, g, general, "vigilance") {
		t.Error("another legendary creature you control did not gain vigilance")
	}
	if _, _, ok := grantedManaAbilityFor(g, general); !ok {
		t.Error("another legendary creature you control did not gain the mana ability")
	}
	for name, id := range map[string]uuid.UUID{"a nonlegendary creature": bear, "an opponent's legendary creature": theirs} {
		if hasEffectiveKeyword(t, g, id, "vigilance") {
			t.Errorf("%s gained vigilance", name)
		}
		if _, _, ok := grantedManaAbilityFor(g, id); ok {
			t.Errorf("%s gained the mana ability", name)
		}
	}
}

// --- Angel's Grace: the life floor -------------------------------------------

// "Until end of turn, damage that would reduce your life total to less
// than 1 reduces it to 1 instead." The damage is still dealt in full (the
// event carries all of it, the part that cost no life rides it as
// DamageNotLifeLoss) and the floor ends with the turn.
func TestAngelsGraceDamageStopsAtOneLifeAndTheFloorEndsWithTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	source := pushVanillaCreature(g, opp.ID, "Their Burner", 2, 2)
	castCatalogSpell(t, g, "Angel's Grace", "Instant", angelsGraceOracle, nil)
	passPriorityAroundTable(t, g)

	me.Life = 5
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(source, me.ID, 20) })
	if me.Life != 1 {
		t.Fatalf("life after 20 damage from 5 under the floor = %d, want 1", me.Life)
	}
	var last game.Event
	for _, ev := range g.Events {
		if ev.Kind == game.EventDealDamage && ev.Target == me.ID {
			last = ev
		}
	}
	if last.Amount != 20 || last.DamageNotLifeLoss != 16 {
		t.Errorf("damage event = %d dealt, %d cost no life; want 20 dealt and 16 that cost none", last.Amount, last.DamageNotLifeLoss)
	}
	// At 1 life further damage costs nothing more.
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(source, me.ID, 3) })
	if me.Life != 1 {
		t.Errorf("life at the floor after more damage = %d, want 1", me.Life)
	}

	advanceToUpkeepOf(t, g, 1)
	if me.Eliminated {
		t.Fatal("eliminated on the way to the next turn")
	}
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(source, me.ID, 3) })
	if me.Life != -2 {
		t.Errorf("life after the turn ended = %d, want the floor gone (-2)", me.Life)
	}
}

// --- Arlinn, the Pack's Hope: the +1 counter half ------------------------------

// "+1: Until your next turn, you may cast creature spells as though they
// had flash, and each creature you control enters with an additional +1/+1
// counter on it." Tokens and cast creatures both get the counter, an
// opponent's creatures do not, and the grant ends when your next turn
// begins.
func TestArlinnThePacksHopePlusOneGivesEnteringCreaturesACounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	arlinn := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: arlinn, Name: "Arlinn, the Pack's Hope", TypeLine: "Legendary Planeswalker — Arlinn",
		OracleID: dnaArlinn, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterLoyalty: 4},
	})
	if err := g.ActivateCatalogAbility(me.ID, arlinn, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	passPriorityAroundTable(t, g)

	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, TokenCard("2/2 green Wolf"), 1)
		_ = g.CreateTokenForEffect(opp.ID, TokenCard("2/2 green Wolf"), 1)
	})
	mine, theirs := findBattlefieldControlledByName(g, me.ID, "Wolf"), findBattlefieldControlledByName(g, opp.ID, "Wolf")
	if mine == uuid.Nil || theirs == uuid.Nil {
		t.Fatalf("missing Wolves: mine %v theirs %v", mine, theirs)
	}
	if got := counterCount(g, mine, game.CounterPlusOne); got != 1 {
		t.Errorf("my token entered with %d +1/+1 counters, want 1", got)
	}
	if got := counterCount(g, theirs, game.CounterPlusOne); got != 0 {
		t.Errorf("an opponent's token entered with %d counters, want 0", got)
	}

	// Until YOUR next turn: gone once it begins.
	advanceToUpkeepOfSeat(t, g, 0)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 green Elf Warrior"), 1) })
	elf := findBattlefieldControlledByName(g, me.ID, "Elf Warrior")
	if elf == uuid.Nil {
		t.Fatal("no Elf Warrior token")
	}
	if got := counterCount(g, elf, game.CounterPlusOne); got != 0 {
		t.Errorf("a creature entering on my next turn got %d counters, want 0", got)
	}
}

// findBattlefieldControlledByName is findBattlefieldByName restricted to
// one controller.
func findBattlefieldControlledByName(g *game.Game, controller uuid.UUID, name string) uuid.UUID {
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == controller {
			return c.InstanceID
		}
	}
	return uuid.Nil
}
