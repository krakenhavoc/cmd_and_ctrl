package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// citys_blessing_cards_test.go — ascend and the city's blessing on the
// catalog (CR 702.131, #2696, ADR 0096's 2026-10-08 amendment). The
// engine half is pinned in game/citys_blessing_test.go.

const (
	waywardSwordtoothOracle = "3875aef0-3102-4fbf-be90-e4139f7a2348"
	ocelotPrideOracle       = "b4ce2c8f-a19c-461e-8a2e-0ec2bd3ebca3"
	twilightProphetOracle   = "a3afd8b9-d499-40f7-be41-f8fee6636a1d"
	andurilOracle           = "b7d7c131-b628-4b35-b40b-1087589ebdd0"
	vonasHungerOracle       = "da522912-1bc2-4d70-b426-94d10edb2afe"
	radiantDestinyOracle    = "068cbbe3-43c3-425e-80ec-d1ce5c02ea9b"
	detectiveOfTheMonthOra  = "10d1d6ca-95f8-48da-8798-37c8aadb1f1f"
	expelFromOrazcaOracle   = "a6e12be3-4166-4bd7-8254-e28b7c8588c7"
	kumenasAwakeningOracle  = "f7a8066e-f8a4-4202-8825-0327d11e58c3"
	secretsGoldenCityOracle = "c0dda0d0-1fae-4777-ba86-9fe7990bf3a8"
	slipperyScoundrelOracle = "3c7ea8d0-d866-4c02-bdfc-13bbc87c094a"
	goldenDemiseOracle      = "e4648fa3-0343-414b-b04d-07fe0d132145"
)

// grantBlessing gives a player the city's blessing the way the engine
// does (the designation and its event, which also invalidates the layer
// pass), without a board of ten permanents.
func grantBlessing(g *game.Game, p *game.Player) {
	g.WithWriteLock(func() {
		p.CitysBlessing = true
		g.EmitEvent(game.Event{Kind: game.EventCitysBlessing, Actor: p.ID})
	})
}

// blessingPermanents puts n vanilla permanents under controller.
func blessingPermanents(g *game.Game, controller uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		b12Permanent(g, controller, "Filler", "Artifact")
	}
}

// --- every ascend card carries the keyword the engine reads ---------

func TestEveryAscendCardDeclaresTheKeyword(t *testing.T) {
	for name, oracle := range map[string]string{
		"Wayward Swordtooth": waywardSwordtoothOracle, "Ocelot Pride": ocelotPrideOracle,
		"Twilight Prophet": twilightProphetOracle, "Andúril": andurilOracle,
		"Vona's Hunger": vonasHungerOracle, "Radiant Destiny": radiantDestinyOracle,
		"Detective of the Month": detectiveOfTheMonthOra, "Expel from Orazca": expelFromOrazcaOracle,
		"Kumena's Awakening": kumenasAwakeningOracle, "Secrets of the Golden City": secretsGoldenCityOracle,
		"Slippery Scoundrel": slipperyScoundrelOracle, "Golden Demise": goldenDemiseOracle,
		"Tendershoot Dryad":       "a336c10a-b5bd-47ff-ba2d-31e27af1e15a",
		"Illustrious Wanderglyph": "72adf091-4491-4afe-b893-b8d20ad38a86",
		"Arch of Orazca":          "3bb518ff-399b-4ce7-b9ad-a1d563dd7792", "Orazca Relic": "48b84b58-1a06-4bb4-be1f-ad3ca69e66dc",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", name)
			continue
		}
		if !hasKeyword(spec.PrintedKeywords, game.KeywordAscend) {
			t.Errorf("%s does not declare ascend", name)
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s completeness = %v, want full", name, spec.Completeness)
		}
	}
}

// --- Wayward Swordtooth ---------------------------------------------

func TestWaywardSwordtoothAttacksOnlyWithTheBlessing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tooth := pushTestPermanent(g, me.ID, game.Card{Name: "Wayward Swordtooth", OracleID: waywardSwordtoothOracle,
		TypeLine: "Creature — Dinosaur", Power: 5, Toughness: 5})
	advanceToDeclareAttackersOf(t, g, 0)

	wantAttackRefused(t, g, tooth, opp.ID, "doesn't have the city's blessing")
	grantBlessing(g, me)
	if err := declareResult(g, tooth, opp.ID); err != nil {
		t.Errorf("attack with the city's blessing: %v", err)
	}
}

func TestWaywardSwordtoothBlocksOnlyWithTheBlessing(t *testing.T) {
	g := newCatalogGame(t)
	attacker, defender := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, attacker.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	tooth := b12Push(g, defender.ID, "Wayward Swordtooth", "Creature — Dinosaur", waywardSwordtoothOracle, 5, 5)
	b12Creature(g, defender.ID, "Elf", "Creature — Elf", 1, 1)
	brAttack(t, g, bear)

	if brOffers(brOffered(t, g, defender.ID), tooth, bear) {
		t.Error("the enumerator offers a block the verb refuses")
	}
	brRefusal(t, g.Clone().DeclareBlocker(tooth, bear), game.BlockReasonCantBlockAttacker)
	grantBlessing(g, defender)
	if !brOffers(brOffered(t, g, defender.ID), tooth, bear) {
		t.Error("the enumerator does not offer the now-legal block")
	}
	if err := g.DeclareBlocker(tooth, bear); err != nil {
		t.Errorf("block with the city's blessing: %v", err)
	}
}

// The Swordtooth is itself an ascend permanent: nine other permanents
// are enough, and the attack is open from then on.
func TestWaywardSwordtoothEarnsTheBlessingItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tooth := pushTestPermanent(g, me.ID, game.Card{Name: "Wayward Swordtooth", OracleID: waywardSwordtoothOracle,
		TypeLine: "Creature — Dinosaur", Power: 5, Toughness: 5})
	blessingPermanents(g, me.ID, 9)
	g.RunStateChecksForTest()
	if !me.CitysBlessing {
		t.Fatal("ten permanents with the Swordtooth's ascend did not give the blessing")
	}
	advanceToDeclareAttackersOf(t, g, 0)
	if err := declareResult(g, tooth, opp.ID); err != nil {
		t.Errorf("attack after earning the blessing: %v", err)
	}
}

// --- Secrets of the Golden City -------------------------------------

func TestSecretsOfTheGoldenCityDrawsTwoOrThree(t *testing.T) {
	for _, tc := range []struct {
		name   string
		blessd bool
		want   int
	}{{"without the blessing", false, 2}, {"with the blessing", true, 3}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			advanceToMain(t, g)
			if tc.blessd {
				grantBlessing(g, me)
			}
			before := me.Hand.Size()
			castCatalogSpell(t, g, "Secrets of the Golden City", "Sorcery", secretsGoldenCityOracle, nil)
			passPriorityAroundTable(t, g)
			// The helper puts the spell in hand and the cast takes it out
			// again, so the net growth is exactly the cards drawn.
			if got := me.Hand.Size() - before; got != tc.want {
				t.Errorf("hand grew by %d, want %d cards drawn", got, tc.want)
			}
		})
	}
}

// CR 702.131a: the spell earns the blessing as it resolves, BEFORE its
// other instructions — ten permanents and no blessing yet draws three.
func TestSecretsOfTheGoldenCityEarnsTheBlessingBeforeItDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	blessingPermanents(g, me.ID, 10)
	if me.CitysBlessing {
		t.Fatal("setup: the blessing arrived without an ascend source")
	}
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Secrets of the Golden City", "Sorcery", secretsGoldenCityOracle, nil)
	passPriorityAroundTable(t, g)
	if !me.CitysBlessing {
		t.Fatal("the resolving ascend sorcery did not give the blessing")
	}
	if got := me.Hand.Size() - before; got != 3 {
		t.Errorf("hand grew by %d, want 3 cards drawn", got)
	}
}

// --- Kumena's Awakening ---------------------------------------------

func TestKumenasAwakeningDrawsForEveryoneOrOnlyYou(t *testing.T) {
	for _, tc := range []struct {
		name   string
		blessd bool
	}{{"without the blessing", false}, {"with the blessing", true}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			pushPermanentForTest(g, me.ID, "Kumena's Awakening", kumenasAwakeningOracle, "Enchantment")
			if tc.blessd {
				grantBlessing(g, me)
			}
			before := make([]int, len(g.Seats))
			for i, p := range g.Seats {
				before[i] = p.Hand.Size()
			}
			// The next upkeep of seat 0 after the setup turn.
			advanceToUpkeepOf(t, g, 1)
			advanceToUpkeepOf(t, g, 0)
			for i, p := range g.Seats {
				before[i] = p.Hand.Size()
			}
			passPriorityAroundTable(t, g)
			for i, p := range g.Seats {
				got := p.Hand.Size() - before[i]
				want := 1
				if tc.blessd && i != 0 {
					want = 0
				}
				if got != want {
					t.Errorf("seat %d drew %d, want %d", i, got, want)
				}
			}
		})
	}
}

// --- Twilight Prophet -----------------------------------------------

func TestTwilightProphetDoesNothingWithoutTheBlessing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	pushPermanentForTest(g, me.ID, "Twilight Prophet", twilightProphetOracle, "Creature — Vampire Cleric")
	top := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: top, Name: "Flipped", TypeLine: "Creature — Bear", ManaCost: "{2}{B}", Owner: me.ID, Controller: me.ID})
	hand, life := me.Hand.Size(), me.Life
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(top) || me.Hand.Size() != hand || me.Life != life {
		t.Error("the upkeep trigger did something without the city's blessing")
	}
}

func TestTwilightProphetDrainsForTheManaValueOfTheRevealedCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	pushPermanentForTest(g, me.ID, "Twilight Prophet", twilightProphetOracle, "Creature — Vampire Cleric")
	grantBlessing(g, me)
	top := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: top, Name: "Flipped", TypeLine: "Creature — Bear", ManaCost: "{2}{B}", Owner: me.ID, Controller: me.ID})
	hand, life := me.Hand.Size(), me.Life
	others := map[int]int{}
	for i, p := range g.Seats {
		others[i] = p.Life
	}

	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)

	if !me.Hand.Contains(top) {
		t.Fatal("the revealed card did not reach the hand")
	}
	// The draw step comes after the upkeep, so exactly the flipped card.
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want exactly one card", hand, me.Hand.Size())
	}
	if me.Life != life+3 {
		t.Errorf("life %d -> %d, want +3 (the revealed card's mana value)", life, me.Life)
	}
	for i, p := range g.Seats {
		if i == 1 {
			continue
		}
		if p.Life != others[i]-3 {
			t.Errorf("seat %d life %d -> %d, want -3", i, others[i], p.Life)
		}
	}
}

// --- Ocelot Pride ---------------------------------------------------

func ocelotSetup(t *testing.T, blessed bool) (*game.Game, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushPermanentForTest(g, me.ID, "Ocelot Pride", ocelotPrideOracle, "Creature — Cat")
	advanceToMain(t, g)
	// A token that entered earlier this turn.
	g.WithWriteLock(func() {
		if _, err := g.CreateTokensForEffect(me.ID, TokenCard("1/1 white Cat with lifelink"), 1, game.TokenEntryOptions{}); err != nil {
			t.Fatalf("create token: %v", err)
		}
	})
	if blessed {
		grantBlessing(g, me)
	}
	return g, me
}

func ocelotTokens(g *game.Game, me *game.Player) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && IsToken(c) {
			n++
		}
	}
	return n
}

func TestOcelotPrideMakesACatOnlyIfYouGainedLife(t *testing.T) {
	g, me := ocelotSetup(t, false)
	before := ocelotTokens(g, me)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := ocelotTokens(g, me); got != before {
		t.Errorf("tokens %d -> %d with no life gained this turn, want no change", before, got)
	}
}

func TestOcelotPrideWithoutTheBlessingMakesOneCat(t *testing.T) {
	g, me := ocelotSetup(t, false)
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1); err != nil {
			t.Fatal(err)
		}
	})
	before := ocelotTokens(g, me)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := ocelotTokens(g, me); got != before+1 {
		t.Errorf("tokens %d -> %d, want one new Cat", before, got)
	}
}

// With the blessing the Cat is made first and then every token that
// entered this turn is copied: the earlier token and the new Cat, each
// once, and the copies are not copied in turn.
func TestOcelotPrideWithTheBlessingCopiesEveryTokenThatEnteredThisTurn(t *testing.T) {
	g, me := ocelotSetup(t, true)
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1); err != nil {
			t.Fatal(err)
		}
	})
	before := ocelotTokens(g, me) // the earlier token
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	// earlier token + new Cat, and a copy of each: before + 1 + (before + 1).
	if want := 2 * (before + 1); ocelotTokens(g, me) != want {
		t.Errorf("tokens = %d, want %d", ocelotTokens(g, me), want)
	}
}

// --- Andúril, Narsil Reforged ---------------------------------------

func TestAndurilPutsOneOrTwoCountersOnEachCreature(t *testing.T) {
	for _, tc := range []struct {
		name   string
		blessd bool
		want   int
	}{{"without the blessing", false, 1}, {"with the blessing", true, 2}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			advanceToMain(t, g)
			bear := seedBear(g, me.ID)
			other := seedBear(g, me.ID)
			theirs := seedBear(g, opp.ID)
			sword := seedEquipment(g, me.ID, "Andúril, Narsil Reforged", andurilOracle)
			equipTo(t, g, me.ID, sword, bear)
			if tc.blessd {
				grantBlessing(g, me)
			}
			advanceTo(t, g, game.StepDeclareAttackers)
			if err := g.DeclareAttacker(bear, opp.ID); err != nil {
				t.Fatalf("DeclareAttacker: %v", err)
			}
			lockInAttacks(t, g)
			passPriorityAroundTable(t, g)
			for _, id := range []uuid.UUID{bear, other} {
				if got := cardByID(g, id).Counters["+1/+1"]; got != tc.want {
					t.Errorf("my creature has %d +1/+1 counters, want %d", got, tc.want)
				}
			}
			if got := cardByID(g, theirs).Counters["+1/+1"]; got != 0 {
				t.Errorf("an opponent's creature got %d counters", got)
			}
		})
	}
}

// --- Vona's Hunger --------------------------------------------------

// vonaAnswer answers every outstanding sacrifice prompt with the first
// ChooseMin candidates and reports how many prompts there were.
func vonaAnswer(t *testing.T, g *game.Game) int {
	t.Helper()
	asked := 0
	for i := 0; i < 8; i++ {
		var choice *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceOwnPermanents {
				choice = c
				break
			}
		}
		if choice == nil {
			return asked
		}
		asked++
		if err := g.ResolveOwnPermanents(choice.ID, choice.Chooser, choice.ChooseCards[:choice.ChooseMin]); err != nil {
			t.Fatalf("ResolveOwnPermanents: %v", err)
		}
	}
	return asked
}

func creaturesOf(g *game.Game, p *game.Player) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == p.ID && c.IsCreature() {
			n++
		}
	}
	return n
}

func TestVonasHungerIsAnEdictWithoutTheBlessingAndHalfWithIt(t *testing.T) {
	for _, tc := range []struct {
		name   string
		blessd bool
		left   [3]int // survivors of seats 1..3 starting with 3, 1, 0 creatures... and 5
	}{
		{"edict", false, [3]int{3, 0, 4}},
		{"half rounded up", true, [3]int{2, 0, 2}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			advanceToMain(t, g)
			// Seat 1: five creatures (3 survive a half, 4 an edict);
			// seat 2: none; seat 3: three (2 left after losing a half
			// rounded up; 2 after an edict).
			for i := 0; i < 5; i++ {
				pushCreatureToBattlefieldForTest(g, g.Seats[1].ID, "A")
			}
			for i := 0; i < 3; i++ {
				pushCreatureToBattlefieldForTest(g, g.Seats[3].ID, "C")
			}
			mine := pushCreatureToBattlefieldForTest(g, me.ID, "Mine")
			if tc.blessd {
				grantBlessing(g, me)
			}
			castCatalogSpell(t, g, "Vona's Hunger", "Instant", vonasHungerOracle, nil)
			passPriorityAroundTable(t, g)
			asked := vonaAnswer(t, g)
			passPriorityAroundTable(t, g)
			if asked != 2 {
				t.Errorf("%d players were asked, want the two opponents with creatures", asked)
			}
			wantA, wantC := 4, 2 // an edict: one each
			if tc.blessd {
				wantA, wantC = 2, 1 // half rounded up: three of five, two of three
			}
			if got := creaturesOf(g, g.Seats[1]); got != wantA {
				t.Errorf("seat 1 has %d creatures, want %d", got, wantA)
			}
			if got := creaturesOf(g, g.Seats[3]); got != wantC {
				t.Errorf("seat 3 has %d creatures, want %d", got, wantC)
			}
			if !g.Battlefield.Contains(mine) {
				t.Error("the caster's own creature was sacrificed")
			}
		})
	}
}

// --- Golden Demise --------------------------------------------------

func TestGoldenDemiseShrinksEveryoneOrOnlyYourOpponents(t *testing.T) {
	for _, tc := range []struct {
		name         string
		blessd       bool
		wantMine     int
		wantTheirsLt int
	}{{"without the blessing", false, 1, 1}, {"with the blessing", true, 3, 1}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			advanceToMain(t, g)
			mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 5)
			theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 3, 5)
			if tc.blessd {
				grantBlessing(g, me)
			}
			castCatalogSpell(t, g, "Golden Demise", "Sorcery", goldenDemiseOracle, nil)
			passPriorityAroundTable(t, g)
			if got := effectivePower(t, g, mine); got != tc.wantMine {
				t.Errorf("my creature's power %d, want %d", got, tc.wantMine)
			}
			if got := effectivePower(t, g, theirs); got != tc.wantTheirsLt {
				t.Errorf("their creature's power %d, want %d", got, tc.wantTheirsLt)
			}
		})
	}
}

// --- Slippery Scoundrel ---------------------------------------------

func TestSlipperyScoundrelIsHexproofAndUnblockableWithTheBlessing(t *testing.T) {
	g := newCatalogGame(t)
	attacker, defender := g.Seats[0], g.Seats[1]
	scoundrel := b12Push(g, attacker.ID, "Slippery Scoundrel", "Creature — Human Pirate", slipperyScoundrelOracle, 2, 2)
	blocker := b12Creature(g, defender.ID, "Wall", "Creature — Wall", 0, 4)
	if hasString(effectiveAbilities(t, g, scoundrel), "hexproof") {
		t.Fatal("hexproof without the blessing")
	}
	grantBlessing(g, attacker)
	if !hasString(effectiveAbilities(t, g, scoundrel), "hexproof") {
		t.Error("no hexproof with the blessing")
	}
	brAttack(t, g, scoundrel)
	brRefusal(t, g.Clone().DeclareBlocker(blocker, scoundrel), game.BlockReasonCantBeBlocked)
}

func TestSlipperyScoundrelCanBeBlockedWithoutTheBlessing(t *testing.T) {
	g := newCatalogGame(t)
	attacker, defender := g.Seats[0], g.Seats[1]
	scoundrel := b12Push(g, attacker.ID, "Slippery Scoundrel", "Creature — Human Pirate", slipperyScoundrelOracle, 2, 2)
	blocker := b12Creature(g, defender.ID, "Wall", "Creature — Wall", 0, 4)
	brAttack(t, g, scoundrel)
	if err := g.DeclareBlocker(blocker, scoundrel); err != nil {
		t.Errorf("block without the blessing: %v", err)
	}
}

// --- Detective of the Month -----------------------------------------

func TestDetectiveOfTheMonthMakesDetectivesUnblockableWithTheBlessing(t *testing.T) {
	g := newCatalogGame(t)
	attacker, defender := g.Seats[0], g.Seats[1]
	b12Push(g, attacker.ID, "Detective of the Month", "Creature — Human Detective", detectiveOfTheMonthOra, 2, 3)
	sleuth := b12Creature(g, attacker.ID, "Sleuth", "Creature — Human Detective", 2, 2)
	bear := b12Creature(g, attacker.ID, "Bear", "Creature — Bear", 2, 2)
	blocker := b12Creature(g, defender.ID, "Wall", "Creature — Wall", 0, 4)
	grantBlessing(g, attacker)
	brAttack(t, g, sleuth, bear)
	brRefusal(t, g.Clone().DeclareBlocker(blocker, sleuth), game.BlockReasonCantBeBlocked)
	if err := g.DeclareBlocker(blocker, bear); err != nil {
		t.Errorf("a non-Detective is still blockable: %v", err)
	}
}

func TestDetectiveOfTheMonthMakesADetectiveOnTheSecondDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pushPermanentForTest(g, me.ID, "Detective of the Month", detectiveOfTheMonthOra, "Creature — Human Detective")
	g.WithWriteLock(func() {
		if err := g.DrawNForEffect(me.ID, 2); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Detective"); n != 1 {
		t.Errorf("Detective tokens = %d, want 1", n)
	}
}

// --- Radiant Destiny ------------------------------------------------

func TestRadiantDestinyGivesVigilanceWithTheBlessing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	elf := b12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	pushNamedTribePermanent(t, g, me.ID, "Radiant Destiny", "Enchantment", radiantDestinyOracle, "Elf")
	if got := effectivePower(t, g, elf); got != 2 {
		t.Errorf("an Elf's power %d, want 2", got)
	}
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("a Bear's power %d, want 2 (not the chosen type)", got)
	}
	if hasString(effectiveAbilities(t, g, elf), "vigilance") {
		t.Error("vigilance without the blessing")
	}
	grantBlessing(g, me)
	if !hasString(effectiveAbilities(t, g, elf), "vigilance") {
		t.Error("no vigilance with the blessing")
	}
	if hasString(effectiveAbilities(t, g, bear), "vigilance") {
		t.Error("a Bear got vigilance")
	}
}

// --- Expel from Orazca ----------------------------------------------

func TestExpelFromOrazcaBouncesWithoutTheBlessing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := pushPermanentForTest(g, opp.ID, "Their Relic", "", "Artifact")
	castCatalogSpell(t, g, "Expel from Orazca", "Instant", expelFromOrazcaOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(target) || !opp.Hand.Contains(target) {
		t.Error("the permanent was not returned to its owner's hand")
	}
	if latestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID) != nil {
		t.Error("asked a question without the blessing")
	}
}

func TestExpelFromOrazcaMayTuckWithTheBlessing(t *testing.T) {
	for _, tc := range []struct {
		name   string
		accept bool
	}{{"tucked", true}, {"declined", false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			grantBlessing(g, me)
			target := pushPermanentForTest(g, opp.ID, "Their Relic", "", "Artifact")
			libBefore := opp.Library.Size()
			castCatalogSpell(t, g, "Expel from Orazca", "Instant", expelFromOrazcaOracle,
				[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
			passPriorityAroundTable(t, g)
			answerMayChoice(t, g, me.ID, tc.accept)
			passPriorityAroundTable(t, g)
			if g.Battlefield.Contains(target) {
				t.Fatal("the permanent stayed on the battlefield")
			}
			if tc.accept {
				if !opp.Library.Contains(target) || opp.Library.Size() != libBefore+1 {
					t.Error("the permanent is not on top of its owner's library")
				}
				if c := opp.Library.Cards[len(opp.Library.Cards)-1]; c.InstanceID != target {
					t.Error("the permanent is in the library but not on top")
				}
			} else if !opp.Hand.Contains(target) {
				t.Error("declining did not return it to hand")
			}
		})
	}
}
