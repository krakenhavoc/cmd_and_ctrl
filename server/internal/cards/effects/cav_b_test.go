package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cav_b_test.go — tracker #3172, slice "cav-b": caveated cards whose
// missing mechanic exists now.

const dragToTheRootsOracle = "ef4c478c-7019-4ec0-8edf-2a3078a8e97a"

// Drag to the Roots costs {2} less with four or more card types among
// the cards in your graveyard — and only YOUR graveyard.
func TestDragToTheRootsDeliriumDiscount(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	price := func() int { return selfPricedMV(t, g, me, dragToTheRootsOracle, "Instant", "{2}{B}{G}") }

	if got := price(); got != 4 {
		t.Fatalf("empty graveyard: %d, want 4", got)
	}
	pushGraveyardCardTyped(me, "Bear", "Creature — Bear")
	pushGraveyardCardTyped(me, "Rock", "Artifact")
	pushGraveyardCardTyped(me, "Bolt", "Instant")
	if got := price(); got != 4 {
		t.Errorf("three card types: %d, want 4", got)
	}
	// An opponent's graveyard never counts.
	pushGraveyardCardTyped(opp, "Island", "Basic Land — Island")
	if got := price(); got != 4 {
		t.Errorf("fourth type only in an opponent's graveyard: %d, want 4", got)
	}
	pushGraveyardCardTyped(me, "Forest", "Basic Land — Forest")
	if got := price(); got != 2 {
		t.Errorf("four card types: %d, want 2", got)
	}
	// Repeats of a type are one type; subtypes and supertypes are not types.
	pushGraveyardCardTyped(me, "Bear 2", "Legendary Creature — Bear")
	if got := price(); got != 2 {
		t.Errorf("five cards, four types: %d, want 2", got)
	}
}

const starfallInvocationOracle = "7024532b-f99b-43a7-b0ed-5b3e7ec7592b"

// Starfall Invocation with the gift promised: the opponent draws, every
// creature dies, and the caster picks one of THEIR creature cards put
// into their graveyard this way to come back under their control.
func TestStarfallInvocationPromisedReturnsOneOfYours(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 3, 3)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	advanceToMain(t, g)
	oppHand := opp.Hand.Size()

	mustCastWithGift(t, g, "Starfall Invocation", "Sorcery", starfallInvocationOracle, nil, opp.ID)
	passPriorityAroundTable(t, g)

	if opp.Hand.Size() != oppHand+1 {
		t.Errorf("the opponent draws the gift: hand %d -> %d", oppHand, opp.Hand.Size())
	}
	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("two of the caster's creatures died: the caster owes a pick")
	}
	if len(pick.ChooseCards) != 2 {
		t.Fatalf("candidates: %d, want exactly the caster's two creature cards", len(pick.ChooseCards))
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{b}); err != nil {
		t.Fatalf("pick Bear B: %v", err)
	}
	if !onBattlefieldByID(g, b) {
		t.Error("the chosen creature card is back on the battlefield")
	}
	if onBattlefieldByID(g, a) || onBattlefieldByID(g, theirs) {
		t.Error("the other creatures stay dead")
	}
	if !me.Graveyard.Contains(a) || !opp.Graveyard.Contains(theirs) {
		t.Error("unpicked creatures stay in their owners' graveyards")
	}
}

// Without the promise the spell is a plain Wrath: nobody draws and
// nothing returns.
func TestStarfallInvocationUnpromisedIsPlainWrath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	advanceToMain(t, g)
	oppHand := opp.Hand.Size()

	mustCastWithGift(t, g, "Starfall Invocation", "Sorcery", starfallInvocationOracle, nil, uuid.Nil)
	passPriorityAroundTable(t, g)

	if opp.Hand.Size() != oppHand {
		t.Errorf("no gift, no draw: hand %d -> %d", oppHand, opp.Hand.Size())
	}
	if onBattlefieldByID(g, a) || !me.Graveyard.Contains(a) {
		t.Error("the creature dies and stays dead")
	}
	if latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID) != nil {
		t.Error("an unpromised gift offers no return")
	}
}

// A promised cast with no creature of the caster's dying returns
// nothing and asks for nothing; one candidate returns outright.
func TestStarfallInvocationPromisedWithOneOrNoCandidate(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	advanceToMain(t, g)
	mustCastWithGift(t, g, "Starfall Invocation", "Sorcery", starfallInvocationOracle, nil, opp.ID)
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID) != nil {
		t.Error("only an opponent's creature died: nothing to return")
	}

	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	only := pushVanillaCreature(g2, me2.ID, "Lone Bear", 2, 2)
	advanceToMain(t, g2)
	mustCastWithGift(t, g2, "Starfall Invocation", "Sorcery", starfallInvocationOracle, nil, opp2.ID)
	passPriorityAroundTable(t, g2)
	if !onBattlefieldByID(g2, only) {
		t.Error("a single candidate returns without a prompt")
	}
}

const hopefulInitiateOracle = "317169b8-9014-48e6-862b-ca21b706846e"

// Training: the Initiate grows only when a creature with GREATER power
// attacks with it, whichever order the attackers were declared in.
func TestHopefulInitiateTrainsOnlyBesideAGreaterPower(t *testing.T) {
	cases := []struct {
		name          string
		otherPower    int
		initiateFirst bool
		withOther     bool
		want          int
	}{
		{"bigger creature, initiate declared first", 3, true, true, 1},
		{"bigger creature, initiate declared second", 3, false, true, 1},
		{"equal power does not count", 1, true, true, 0},
		{"smaller creature does not count", 0, true, true, 0},
		{"attacking alone", 3, true, false, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			init := pushDiesCreatureForTest(g, me.ID, "Hopeful Initiate", hopefulInitiateOracle, "Creature — Human Warlock", 1, 1)
			other := pushVanillaCreature(g, me.ID, "Other", tc.otherPower, 3)
			var order []uuid.UUID
			switch {
			case !tc.withOther:
				order = []uuid.UUID{init}
			case tc.initiateFirst:
				order = []uuid.UUID{init, other}
			default:
				order = []uuid.UUID{other, init}
			}
			attackWith(t, g, opp.ID, order...)
			passPriorityAroundTable(t, g)
			if got := counterCount(g, init, game.CounterPlusOne); got != tc.want {
				t.Errorf("+1/+1 counters on the Initiate: %d, want %d", got, tc.want)
			}
		})
	}
}

// Only creatures attacking count: a bigger creature that stays home
// does not train the Initiate.
func TestHopefulInitiateIgnoresABiggerCreatureThatStaysHome(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	init := pushDiesCreatureForTest(g, me.ID, "Hopeful Initiate", hopefulInitiateOracle, "Creature — Human Warlock", 1, 1)
	pushVanillaCreature(g, me.ID, "Home", 5, 5)
	attackWith(t, g, opp.ID, init)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, init, game.CounterPlusOne); got != 0 {
		t.Errorf("counters: %d, want 0", got)
	}
}

const (
	asininAnticsOracle  = "0a0a051c-59ed-4286-b842-a34da7ef15d5"
	oakshadeStalkerOrcl = "1b639537-a8fe-4615-b937-afa478fdec0f"
)

// "As though it had flash if you pay {2} more": the printed cost is
// still sorcery-speed only; the surcharge offer opens instant speed for
// that cast alone and charges the printed cost plus {2}.
func TestFlashForTwoMoreOpensInstantSpeedOnlyThroughTheSurcharge(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine, printed string
		pool                            []string // exactly printed + {2}
	}{
		{"Asinine Antics", asininAnticsOracle, "Sorcery", "{2}{U}{U}", []string{"U", "U", "C", "C", "C", "C"}},
		{"Oakshade Stalker", oakshadeStalkerOrcl, "Creature — Human Ranger Werewolf", "{2}{G}", []string{"G", "C", "C", "C", "C"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := p7Table(t)
			card := eacHandCard(g, me, tc.name, tc.typeLine, tc.printed)
			g.WithWriteLock(func() {
				for i := range me.Hand.Cards {
					if me.Hand.Cards[i].InstanceID == card {
						me.Hand.Cards[i].OracleID = tc.oracle
					}
				}
			})
			for g.Turn.Step != game.StepBeginCombat {
				if _, err := g.AdvanceStep(); err != nil {
					t.Fatalf("AdvanceStep: %v", err)
				}
			}
			if err := g.CastSpell(me.ID, card, game.CastSpellParams{}); !errors.Is(err, game.ErrSorcerySpeedRequired) {
				t.Fatalf("printed cost at instant speed: err = %v, want ErrSorcerySpeedRequired", err)
			}
			fund := func(colors []string) {
				g.WithWriteLock(func() {
					me.ManaPool = nil
					for _, c := range colors {
						me.ManaPool.AddMana(game.ManaToken{Color: c})
					}
				})
			}
			// One short of the surcharge price is refused.
			fund(tc.pool[1:])
			if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true, AlternativeCost: "flash-for-more"}); err == nil {
				t.Fatal("the surcharge offer cast one mana short was accepted")
			}
			fund(tc.pool)
			if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true, AlternativeCost: "flash-for-more"}); err != nil {
				t.Fatalf("flash cast for {2} more: %v", err)
			}
			if len(me.ManaPool) != 0 {
				t.Errorf("mana left in the pool: %d, want 0 (printed cost + {2} was spent)", len(me.ManaPool))
			}
		})
	}
}

// At sorcery speed on your own turn the printed cost still works and
// costs the printed amount.
func TestFlashForTwoMoreLeavesThePrintedCastAlone(t *testing.T) {
	g, me, _ := p7Table(t)
	card := eacHandCard(g, me, "Asinine Antics", "Sorcery", "{2}{U}{U}")
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == card {
				me.Hand.Cards[i].OracleID = asininAnticsOracle
			}
		}
		for _, c := range []string{"U", "U", "C", "C"} {
			me.ManaPool.AddMana(game.ManaToken{Color: c})
		}
	})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("printed cast in a main phase: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("mana left: %d, want 0", len(me.ManaPool))
	}
}

const vinesOfVastwoodOracle = "8998d211-5ca6-41d2-bc64-f51f70bd41e5"

// Vines of Vastwood: the creature can't be targeted by the caster's
// opponents either way; kicked ({G} more) it also gets +4/+4.
func TestVinesOfVastwoodKickerAddsPlusFourPlusFour(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
		advanceToMain(t, g)
		id := pushCatalogHandCard(me, "Vines of Vastwood", "Instant", vinesOfVastwoodOracle)
		params := game.CastSpellParams{Targets: cardTarget(bear)}
		if kicked {
			idx := -1
			for i, oc := range game.OptionalCostsFor(vinesOfVastwoodOracle) {
				if oc.Key == game.KickerKey {
					idx = i
				}
			}
			if idx < 0 {
				t.Fatal("Vines offers no kicker")
			}
			params.OptionalCosts = []int{idx}
		}
		if err := g.CastSpell(me.ID, id, params); err != nil {
			t.Fatalf("kicked=%v: %v", kicked, err)
		}
		passPriorityAroundTable(t, g)
		if eotHasAbility(effectiveAbilities(t, g, bear), "hexproof") {
			t.Errorf("kicked=%v: Vines is not hexproof (#3860), the keyword must not appear", kicked)
		}
		want := 2
		if kicked {
			want = 6
		}
		if got := effectivePower(t, g, bear); got != want {
			t.Errorf("kicked=%v: power %d, want %d", kicked, got, want)
		}
	}
}

const restartSequenceOracle = "c415bfdd-3d42-4f94-9f93-7e310becbc8b"

// Restart Sequence: freerunning {1}{B} is offered only after an Assassin
// or a commander under your control dealt combat damage to a player this
// turn; the printed {3}{B} always works.
func TestRestartSequenceFreerunningNeedsAnAssassinOrCommander(t *testing.T) {
	for _, tc := range []struct {
		name      string
		typeLine  string
		commander bool
		attack    bool
		want      bool
	}{
		{"assassin connects", "Creature — Human Assassin", false, true, true},
		{"commander connects", "Creature — Human Warrior", true, true, true},
		{"other creature connects", "Creature — Human Warrior", false, true, false},
		{"assassin stays home", "Creature — Human Assassin", false, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			dealer := pushDiesCreatureForTest(g, me.ID, "Dealer", "", tc.typeLine, 2, 2)
			g.WithWriteLock(func() {
				for i := range g.Battlefield.Cards {
					if g.Battlefield.Cards[i].InstanceID == dealer {
						g.Battlefield.Cards[i].IsCommander = tc.commander
					}
				}
			})
			dead := pushGraveyardCreature(g, me.ID, "Fallen", "{2}")
			if tc.attack {
				attackWith(t, g, opp.ID, dealer)
			}
			advanceTo(t, g, game.StepPostcombatMain)
			spell := handCardFull(me, "Restart Sequence", "Sorcery", "{3}{B}", restartSequenceOracle, []string{"B"})
			params := game.CastSpellParams{Strict: true, AlternativeCost: "freerunning", Targets: cardTarget(dead)}
			floatForTest(g, me, "CB")
			err := g.CastSpell(me.ID, spell, params)
			if tc.want && err != nil {
				t.Fatalf("freerunning was refused: %v", err)
			}
			if !tc.want && err == nil {
				t.Fatal("freerunning was accepted without an Assassin or commander connecting")
			}
			if !tc.want {
				// The printed cost still works.
				floatForTest(g, me, "CCB")
				params.AlternativeCost = ""
				if err := g.CastSpell(me.ID, spell, params); err != nil {
					t.Fatalf("printed {3}{B} cast: %v", err)
				}
			}
			passPriorityAroundTable(t, g)
			if !onBattlefieldByID(g, dead) {
				t.Error("the creature card should have been reanimated")
			}
		})
	}
}
