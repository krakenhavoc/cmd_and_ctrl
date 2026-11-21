package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b02_cards_test.go — Mystery Booster Commander Edition creatures
// (S82, #5948, GAP batch 2). Each test plays the real catalog entry through
// the engine.

const (
	gapB02MuckslingerOracle = "56cacdcf-2f11-4969-9629-446068a15369"
	gapB02NephilimOracle    = "c1198033-47bc-4bee-a3a5-3917dae5cba3"
	gapB02HomerOracle       = "a13aa603-8709-43aa-b69a-ee74de79561f"
	gapB02UuugguOracle      = "df9bb1fc-135d-4efb-b143-81950fc95a27"
	gapB02OlagOracle        = "d117825e-fb7d-44dc-a17d-71bd610a6078"
	gapB02JesterOracle      = "60f3857d-2dcd-427e-aa0f-e2028bbc5acc"
	gapB02TresserhornOracle = "11cc638c-b1c5-4b73-8eac-00cac6ae32b6"
	gapB02TsaganOracle      = "3194c964-1693-40b6-b4f2-64fe056e7cd9"
	gapB02WeaverOracle      = "ed1a6d06-436b-45c0-979c-7802246eeb9e"
	gapB02LynaOracle        = "82321fa9-7a95-4c9d-988e-2140eaf1a083"
)

func TestGapB02EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		gapB02MuckslingerOracle: "The Muckslinger Gang",
		gapB02NephilimOracle:    "Nephilim Epochal",
		gapB02HomerOracle:       "Homer, the Hermit",
		gapB02UuugguOracle:      "Uugguu, the Omniplasm",
		gapB02OlagOracle:        "Olag and Miau, New Friends",
		gapB02JesterOracle:      "The Madcap Jester",
		gapB02TresserhornOracle: "Tresserhorn's Lord, Returned",
		gapB02TsaganOracle:      "Tsagan, Raider Warlord",
		gapB02WeaverOracle:      "The Weaver King",
		gapB02LynaOracle:        "Lyna, Veil of Vengeance",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func gapB02Active(g *game.Game) (*game.Player, *game.Player) {
	me := g.Seats[g.Turn.ActiveSeat]
	return me, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// gapB02Cast casts a permanent spell out of a floated pool under the strict
// gate and lets what it can resolve.
func gapB02Cast(t *testing.T, g *game.Game, me *game.Player, name, typeLine, cost, oracle, pool string) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	if pool != "" {
		floatForTest(g, me, pool)
	}
	id := castFromHandForTest(t, g, me, name, typeLine, cost, oracle, game.CastSpellParams{Strict: true})
	for i := 0; i < 12 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
		if latestPickTarget(g, me.ID) != nil {
			break
		}
	}
	return id
}

// --- Tsagan, Raider Warlord ---------------------------------------------

func TestGapB02TsaganPumpsForEachFirstOrDoubleStrikeCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := gapB02Active(g)
	advanceToMain(t, g)
	ts := pushSpeedCard(t, g, me, "Tsagan, Raider Warlord", "Legendary Creature — Dinosaur Berserker", gapB02TsaganOracle, 1, 4)
	striker := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Striker", TypeLine: "Creature — Soldier", Keywords: []string{"first strike"},
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	plain := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Plain", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	assertKeywords(t, g, ts, "double strike")
	declareAttack(t, g, opp.ID, ts)
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	// Tsagan (double strike) and Striker (first strike) count: +2/+0 each.
	if got := effectivePower(t, g, ts); got != 3 {
		t.Errorf("Tsagan power %d, want 3", got)
	}
	if got := effectivePower(t, g, striker); got != 4 {
		t.Errorf("Striker power %d, want 4", got)
	}
	if got := effectivePower(t, g, plain); got != 4 {
		t.Errorf("a creature with neither keyword still gets the pump: power %d, want 4", got)
	}
}

func TestGapB02TsaganMaxSpeedGivesDeathtouchAndFirstStrike(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := gapB02Active(g)
	advanceToMain(t, g)
	ts := pushSpeedCard(t, g, me, "Tsagan, Raider Warlord", "Legendary Creature — Dinosaur Berserker", gapB02TsaganOracle, 1, 4)
	ally := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ally", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	if effectiveAbilitiesContain(t, g, ts, "deathtouch") || effectiveAbilitiesContain(t, g, ally, "first strike") {
		t.Fatal("max-speed abilities below max speed")
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	if !effectiveAbilitiesContain(t, g, ts, "deathtouch") {
		t.Error("Tsagan has no deathtouch at max speed")
	}
	if !effectiveAbilitiesContain(t, g, ally, "first strike") {
		t.Error("other creatures have no first strike at max speed")
	}
	if effectiveAbilitiesContain(t, g, ts, "first strike") {
		t.Error("Tsagan itself gets first strike from its own other-creatures clause")
	}
}

// --- Homer, the Hermit --------------------------------------------------

func TestGapB02HomerMillsTwiceTheCrabsForEachTargetedPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := gapB02Active(g)
	third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Homer, the Hermit", "Legendary Creature — Crab Druid", gapB02HomerOracle, false)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Starfish", TypeLine: "Creature — Starfish", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	oppLib, thirdLib := opp.Library.Size(), third.Library.Size()
	playLandFromHand(t, g, "Forest", "")
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for Homer's landfall")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	// Homer (a Crab) and the Starfish: X = 4.
	if got := oppLib - opp.Library.Size(); got != 4 {
		t.Errorf("the targeted player milled %d, want 4", got)
	}
	if got := thirdLib - third.Library.Size(); got != 0 {
		t.Errorf("an untargeted player milled %d", got)
	}
}

func TestGapB02HomerMayTargetNobody(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := gapB02Active(g)
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Homer, the Hermit", "Legendary Creature — Crab Druid", gapB02HomerOracle, false)
	oppLib := opp.Library.Size()
	playLandFromHand(t, g, "Forest", "")
	if pick := latestPickTarget(g, me.ID); pick != nil {
		if err := g.ResolvePickTargets(pick.ID, me.ID, nil); err != nil {
			t.Fatalf("ResolvePickTargets: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if got := oppLib - opp.Library.Size(); got != 0 {
		t.Errorf("milled %d with no targets", got)
	}
}

// --- Lyna, Veil of Vengeance --------------------------------------------

func TestGapB02LynaDrawsWhenAShadowCreatureConnects(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := gapB02Active(g)
	advanceToMain(t, g)
	lyna := pushCatalogPermanent(g, me.ID, "Lyna, Veil of Vengeance", "Legendary Creature — Soltari Cleric", gapB02LynaOracle, false)
	assertKeywords(t, g, lyna, "shadow")
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	hand := me.Hand.Size()
	dealCombatDamageToPlayer(g, bear, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("a creature without shadow drew: hand %d -> %d", hand, got)
	}
	dealCombatDamageToPlayer(g, lyna, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want +1 for the shadow creature", hand, got)
	}
}

func TestGapB02LynaShadowCounterAtSorcerySpeedMakesAShadowCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := gapB02Active(g)
	advanceToMain(t, g)
	lyna := pushCatalogPermanent(g, me.ID, "Lyna, Veil of Vengeance", "Legendary Creature — Soltari Cleric", gapB02LynaOracle, false)
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	floatForTest(g, me, "CCC")
	// Another target creature only: Lyna herself is not a legal target.
	if err := g.ActivateCatalogAbility(me.ID, lyna, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(lyna)}); err == nil {
		t.Error("Lyna targeted herself")
	}
	if err := g.ActivateCatalogAbility(me.ID, lyna, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	assertKeywords(t, g, bear, "shadow")
	hand := me.Hand.Size()
	dealCombatDamageToPlayer(g, bear, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("the shadow counter creature did not draw: hand %d -> %d", hand, got)
	}
	// Not at instant speed.
	advanceTo(t, g, game.StepBeginCombat)
	floatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, lyna, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(bear)}); err == nil {
		t.Error("activated outside a main phase")
	}
}

// --- The Muckslinger Gang -----------------------------------------------

// gapB02CardPower is the post-layer power of a card in a hand, library,
// graveyard or on the stack.
func gapB02CardPower(g *game.Game, id uuid.UUID) (int, bool) {
	var power int
	var found bool
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		for _, p := range g.Seats {
			for _, z := range []*game.Zone{p.Hand, p.Library, p.Graveyard} {
				for _, c := range z.Cards {
					if c.InstanceID == id {
						power, found = c.CurrentPower(), true
					}
				}
			}
		}
	})
	return power, found
}

func TestGapB02MuckslingerPumpsOtherCreaturesUnlessAttackingOrBlocking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := gapB02Active(g)
	advanceToMain(t, g)
	gang := pushCatalogPermanent(g, me.ID, "The Muckslinger Gang", "Legendary Creature — Goblin Lizard Beast", gapB02MuckslingerOracle, false)
	// pushCatalogPermanent seeds a 1/1; the catalog's 3/3 comes from the printed card.
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if got := effectivePower(t, g, bear); got != 5 {
		t.Errorf("another creature I control has power %d, want 5", got)
	}
	if got := effectivePower(t, g, gang); got != 1 {
		t.Errorf("The Muckslinger Gang pumps itself: power %d, want its own 1", got)
	}
	if got := effectivePower(t, g, theirs); got != 2 {
		t.Errorf("an opponent's creature has power %d, want 2", got)
	}
	declareAttack(t, g, opp.ID, attacker)
	if got := effectivePower(t, g, attacker); got != 2 {
		t.Errorf("an attacking creature has power %d, want 2", got)
	}
	if got := effectivePower(t, g, bear); got != 5 {
		t.Errorf("a creature that is not attacking has power %d, want 5", got)
	}
}

func TestGapB02MuckslingerPumpsCreatureCardsOffTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := gapB02Active(g)
	advanceToMain(t, g)
	handBear, libBear, yardBear, oppBear, land := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: handBear, Name: "Hand Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: libBear, Name: "Library Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: yardBear, Name: "Yard Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	opp.Hand.PushTop(game.Card{InstanceID: oppBear, Name: "Their Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	me.Hand.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	gang := pushCatalogPermanent(g, me.ID, "The Muckslinger Gang", "Legendary Creature — Goblin Lizard Beast", gapB02MuckslingerOracle, false)
	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want int
	}{
		{"a creature card in my hand", handBear, 5},
		{"a creature card in my library", libBear, 5},
		{"a creature card in my graveyard", yardBear, 5},
		{"a creature card in an opponent's hand", oppBear, 2},
		{"a land card in my hand", land, 0},
	} {
		if got, ok := gapB02CardPower(g, tc.id); !ok || got != tc.want {
			t.Errorf("%s: power %d (found %v), want %d", tc.name, got, ok, tc.want)
		}
	}
	// The bonus goes with the Gang.
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, gang); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	if got, _ := gapB02CardPower(g, handBear); got != 2 {
		t.Errorf("a creature card in hand keeps the bonus after the Gang left: power %d", got)
	}
}

// --- Nephilim Epochal ---------------------------------------------------

func TestGapB02NephilimEpochalTutorsANephilim(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := gapB02Active(g)
	advanceToMain(t, g)
	nephilim := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: nephilim, Name: "Yore-Tiller Nephilim", TypeLine: "Creature — Nephilim", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	gapB02Cast(t, g, me, "Nephilim Epochal", "Legendary Creature — Elder Nephilim", "{U/R}{R/G}{G/W}{W/B}", gapB02NephilimOracle, "URGW")
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, me.ID) == nil {
		t.Fatal("no search prompt")
	}
	answerSearchByID(t, g, me.ID, nephilim)
	passPriorityAroundTable(t, g)
	inHand := false
	for _, c := range me.Hand.Cards {
		if c.InstanceID == nephilim {
			inHand = true
		}
	}
	if !inHand {
		t.Error("the Nephilim card is not in hand")
	}
}

func TestGapB02NephilimEpochalGrowsEachNephilimByTheOthers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := gapB02Active(g)
	advanceToMain(t, g)
	epochal := b12Push(g, me.ID, "Nephilim Epochal", "Legendary Creature — Elder Nephilim", gapB02NephilimOracle, 4, 4)
	if got := effectivePower(t, g, epochal); got != 4 {
		t.Errorf("alone it is %d power, want 4", got)
	}
	other := b12Creature(g, me.ID, "Dune-Brood Nephilim", "Creature — Nephilim", 2, 2)
	third := b12Creature(g, me.ID, "Glint-Eye Nephilim", "Creature — Nephilim", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Nephilim", "Creature — Nephilim", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for name, tc := range map[string]struct {
		id   uuid.UUID
		want int
	}{
		"Epochal":        {epochal, 6},
		"another":        {other, 4},
		"the third":      {third, 4},
		"an opponent":    {theirs, 2},
		"a non-Nephilim": {bear, 2},
	} {
		if got := effectivePower(t, g, tc.id); got != tc.want {
			t.Errorf("%s: power %d, want %d", name, got, tc.want)
		}
		if got := effectiveToughness(t, g, tc.id); got != tc.want {
			t.Errorf("%s: toughness %d, want %d", name, got, tc.want)
		}
	}
}

func TestGapB02NephilimCantBeBlockedByMonocoloredCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := gapB02Active(g)
	advanceToMain(t, g)
	epochal := b12Push(g, me.ID, "Nephilim Epochal", "Legendary Creature — Elder Nephilim", gapB02NephilimOracle, 4, 4)
	mono := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Mono", TypeLine: "Creature — Bear", Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	multi := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Multi", TypeLine: "Creature — Bear", Colors: []string{"G", "W"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	colorless := b12Creature(g, opp.ID, "Colorless", "Artifact Creature — Golem", 2, 2)
	declareAttack(t, g, opp.ID, epochal)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(mono, epochal); err == nil {
		t.Error("a monocolored creature blocked")
	}
	if err := g.DeclareBlocker(multi, epochal); err != nil {
		t.Errorf("a multicolored creature could not block: %v", err)
	}
	if err := g.DeclareBlocker(colorless, epochal); err != nil {
		t.Errorf("a colorless creature could not block: %v", err)
	}
}

// --- Uugguu, the Omniplasm ----------------------------------------------

func TestGapB02UuugguCopiesADyingNontokenOozeAsTwoTwoTwos(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := gapB02Active(g)
	advanceToMain(t, g)
	b12Push(g, me.ID, "Uugguu, the Omniplasm", "Legendary Creature — Ooze", gapB02UuugguOracle, 2, 2)
	ooze := b12Push(g, me.ID, "Big Ooze", "Legendary Creature — Ooze", "oracle-big-ooze", 6, 6)
	killCreature(t, g, me.ID, ooze)
	passPriorityAroundTable(t, g)
	toks := dmuB17Tokens(g, "Big Ooze")
	if len(toks) != 2 {
		t.Fatalf("%d Big Ooze tokens, want 2", len(toks))
	}
	for _, tok := range toks {
		if tok.HasSupertype("Legendary") {
			t.Error("a copy is legendary")
		}
		if p, tg := effectivePower(t, g, tok.InstanceID), effectiveToughness(t, g, tok.InstanceID); p != 2 || tg != 2 {
			t.Errorf("a copy is %d/%d, want 2/2", p, tg)
		}
		if !tok.HasSubtype("Ooze") {
			t.Errorf("a copy is not an Ooze: %q", tok.TypeLine)
		}
	}
}

func TestGapB02UuugguIgnoresTokensNonOozesItselfAndOpposingOozes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := gapB02Active(g)
	advanceToMain(t, g)
	uug := b12Push(g, me.ID, "Uugguu, the Omniplasm", "Legendary Creature — Ooze", gapB02UuugguOracle, 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Ooze", "Creature — Ooze", 2, 2)
	token := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Ooze Token", TypeLine: "Token Creature — Ooze", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	killCreature(t, g, me.ID, bear)
	killCreature(t, g, opp.ID, theirs)
	killCreature(t, g, me.ID, token)
	killCreature(t, g, me.ID, uug)
	passPriorityAroundTable(t, g)
	for _, name := range []string{"Bear", "Their Ooze", "Ooze Token", "Uugguu, the Omniplasm"} {
		if got := len(dmuB17Tokens(g, name)); got != 0 {
			t.Errorf("%d %s copies, want none", got, name)
		}
	}
}

// --- Olag and Miau, New Friends -----------------------------------------

func TestGapB02OlagAndMiauCountersTheOtherTribeWhenOneEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := gapB02Active(g)
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Olag and Miau, New Friends", "Legendary Creature — Zombie Cat", gapB02OlagOracle, false)
	cat := b12Creature(g, me.ID, "Alley Cat", "Creature — Cat", 1, 1)
	zombie := b12Creature(g, me.ID, "Walker", "Creature — Zombie", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	gapB02Cast(t, g, me, "New Zombie", "Creature — Zombie", "{0}", "", "")
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for a Zombie entering")
	}
	for _, c := range pick.PickTargetCards {
		if c == zombie || c == bear {
			t.Error("offered something that is not a Cat")
		}
	}
	answerPickTarget(t, g, cat)
	passPriorityAroundTable(t, g)
	if got := battlefieldCardCopy(t, g, cat).CurrentPower(); got != 2 {
		t.Errorf("the Cat has power %d, want 2", got)
	}
	if got := battlefieldCardCopy(t, g, bear).CurrentPower(); got != 1 {
		t.Errorf("a Bear got a counter: power %d", got)
	}
}

func TestGapB02OlagAndMiauDrainsForTappedCatsAndZombies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := gapB02Active(g)
	advanceToMain(t, g)
	olag := pushCatalogPermanent(g, me.ID, "Olag and Miau, New Friends", "Legendary Creature — Zombie Cat", gapB02OlagOracle, false)
	cat := b12Creature(g, me.ID, "Alley Cat", "Creature — Cat", 1, 1)
	zombie := b12Creature(g, me.ID, "Walker", "Creature — Zombie", 1, 1)
	b12Creature(g, me.ID, "Untapped Cat", "Creature — Cat", 1, 1)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			switch g.Battlefield.Cards[i].InstanceID {
			case olag, cat, zombie:
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	life := opp.Life
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	// Olag and Miau is both a Cat and a Zombie and counts once.
	if opp.Life != life-3 {
		t.Errorf("opponent life %d -> %d, want 3 lost", life, opp.Life)
	}
}

// --- Tresserhorn's Lord, Returned ---------------------------------------

// gapB02CastLord casts Tresserhorn's Lord and answers its target prompt
// with `opp`, leaving the trigger to resolve.
func gapB02CastLord(t *testing.T, g *game.Game, me, opp *game.Player) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	floatForTest(g, me, "CCUBR")
	id := castFromHandForTest(t, g, me, "Tresserhorn's Lord, Returned", "Legendary Creature — Zombie Noble", "{2}{U}{B}{R}",
		gapB02TresserhornOracle, game.CastSpellParams{Strict: true})
	for i := 0; i < 12 && g.Stack.Size() > 0; i++ {
		if latestPickTarget(g, me.ID) != nil {
			answerPickTargetPlayer(t, g, opp.ID)
		}
		if latestOptionPickFor(g, me.ID) != nil {
			break
		}
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	if latestPickTarget(g, me.ID) != nil {
		answerPickTargetPlayer(t, g, opp.ID)
		passPriorityAroundTable(t, g)
	}
	return id
}

func gapB02OnBattlefield(g *game.Game, id uuid.UUID) bool {
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				found = true
			}
		}
	})
	return found
}

func TestGapB02TresserhornPaysTheWholeCostAndKeepsTheLord(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := gapB02Active(g)
	var fodder []uuid.UUID
	for i := 0; i < 3; i++ {
		fodder = append(fodder, b12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1))
	}
	spare := b12Creature(g, me.ID, "Spare", "Creature — Bear", 1, 1)
	life, oppHand, myHand := me.Life, opp.Hand.Size(), me.Hand.Size()
	lord := gapB02CastLord(t, g, me, opp)
	if latestOptionPickFor(g, me.ID) == nil {
		t.Fatal("no choice between sacrificing the Lord and paying")
	}
	answerOptionPick(t, g, me.ID, 1)
	answerChooseCards(t, g, me.ID, fodder...)
	passPriorityAroundTable(t, g)
	if !gapB02OnBattlefield(g, lord) {
		t.Error("the Lord was sacrificed despite the payment")
	}
	for _, id := range fodder {
		if gapB02OnBattlefield(g, id) {
			t.Error("a chosen creature was not sacrificed")
		}
	}
	if !gapB02OnBattlefield(g, spare) {
		t.Error("an unchosen creature was sacrificed")
	}
	if me.Life != life-3 {
		t.Errorf("life %d -> %d, want 3 paid", life, me.Life)
	}
	if got := opp.Hand.Size() - oppHand; got != 3 {
		t.Errorf("the target opponent drew %d, want 3", got)
	}
	// Whenever you sacrifice a creature, draw a card: three sacrifices.
	if got := me.Hand.Size() - myHand; got != 3 {
		t.Errorf("I drew %d, want 3 from the three sacrifices", got)
	}
}

func TestGapB02TresserhornSacrificesItselfWhenYouChooseNotToPay(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := gapB02Active(g)
	for i := 0; i < 3; i++ {
		b12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	}
	life, oppHand, myHand := me.Life, opp.Hand.Size(), me.Hand.Size()
	lord := gapB02CastLord(t, g, me, opp)
	answerOptionPick(t, g, me.ID, 0)
	passPriorityAroundTable(t, g)
	if gapB02OnBattlefield(g, lord) {
		t.Error("the Lord stayed without a payment")
	}
	if me.Life != life || opp.Hand.Size() != oppHand {
		t.Errorf("a cost was paid anyway: life %d -> %d, opponent hand %d -> %d", life, me.Life, oppHand, opp.Hand.Size())
	}
	// Sacrificing the Lord is a sacrifice of a creature and draws.
	if got := me.Hand.Size() - myHand; got != 1 {
		t.Errorf("I drew %d, want 1 for sacrificing the Lord", got)
	}
}

func TestGapB02TresserhornIsSacrificedWithTooFewCreaturesOrLife(t *testing.T) {
	t.Run("two other creatures and the Lord is three, so the Lord may be one of them", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := gapB02Active(g)
		a := b12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
		b := b12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
		lord := gapB02CastLord(t, g, me, opp)
		if latestOptionPickFor(g, me.ID) == nil {
			t.Fatal("the payment was not offered with exactly three creatures")
		}
		answerOptionPick(t, g, me.ID, 1)
		answerChooseCards(t, g, me.ID, a, b, lord)
		passPriorityAroundTable(t, g)
		if gapB02OnBattlefield(g, lord) {
			t.Error("the Lord was sacrificed to its own cost but stayed")
		}
	})
	t.Run("one other creature", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := gapB02Active(g)
		b12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
		oppHand := opp.Hand.Size()
		lord := gapB02CastLord(t, g, me, opp)
		passPriorityAroundTable(t, g)
		if latestOptionPickFor(g, me.ID) != nil {
			t.Error("offered a payment that can't be made")
		}
		if gapB02OnBattlefield(g, lord) || opp.Hand.Size() != oppHand {
			t.Error("the Lord stayed or the opponent drew")
		}
	})
	t.Run("life 2", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := gapB02Active(g)
		for i := 0; i < 3; i++ {
			b12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
		}
		me.Life = 2
		lord := gapB02CastLord(t, g, me, opp)
		passPriorityAroundTable(t, g)
		if latestOptionPickFor(g, me.ID) != nil {
			t.Error("offered a payment of 3 life at 2 life")
		}
		if gapB02OnBattlefield(g, lord) {
			t.Error("the Lord stayed")
		}
	})
}

// --- The Madcap Jester --------------------------------------------------

// gapB02JesterAttacks puts The Madcap Jester on the battlefield with a known
// top card in every opponent's library, attacks with it and returns those
// top cards.
func gapB02JesterAttacks(t *testing.T, g *game.Game, me *game.Player) map[uuid.UUID]uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	jester := pushCatalogPermanent(g, me.ID, "The Madcap Jester", "Legendary Creature — Clown Wizard", gapB02JesterOracle, false)
	tops := map[uuid.UUID]uuid.UUID{}
	for _, p := range g.Seats {
		if p.ID == me.ID {
			continue
		}
		id := uuid.New()
		p.Library.PushTop(game.Card{InstanceID: id, Name: "Top of " + p.Name, TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
		tops[p.ID] = id
	}
	declareAttack(t, g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, jester)
	lockInAttacks(t, g)
	return tops
}

func gapB02InExile(g *game.Game, id uuid.UUID) bool {
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Exile.Cards {
			if c.InstanceID == id {
				found = true
			}
		}
	})
	return found
}

func TestGapB02JesterExilesEachOpponentsTopCardAndLetsYouPlayThemWithAnyMana(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := gapB02Active(g)
	tops := gapB02JesterAttacks(t, g, me)
	passPriorityAroundTable(t, g)
	if pr := latestTriggerPromptFor(g, me.ID); pr != nil {
		if err := g.ResolveTriggerPrompt(pr.ID, me.ID, true); err != nil {
			t.Fatalf("ResolveTriggerPrompt: %v", err)
		}
	} else {
		t.Fatal("no \"you may play those cards\" offer")
	}
	passPriorityAroundTable(t, g)
	for owner, id := range tops {
		if !gapB02InExile(g, id) {
			t.Errorf("the top card of %s was not exiled", owner)
			continue
		}
		perm := g.CastPermissionOnCardByIDForEffect(id)
		if perm == nil || perm.Player != me.ID || !perm.AnyType {
			t.Errorf("no any-type permission for me over the card from %s: %+v", owner, perm)
		}
	}
	// One card is castable with the wrong colours: mana of any type can pay.
	var target uuid.UUID
	for _, id := range tops {
		target = id
		break
	}
	advanceTo(t, g, game.StepPostcombatMain)
	floatForTest(g, me, "RB")
	if err := g.CastSpell(me.ID, target, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Errorf("could not cast the exiled {1}{G} creature with {R}{B}: %v", err)
	}
}

func TestGapB02JesterGrantsNothingWhenYouDecline(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := gapB02Active(g)
	tops := gapB02JesterAttacks(t, g, me)
	passPriorityAroundTable(t, g)
	if pr := latestTriggerPromptFor(g, me.ID); pr != nil {
		if err := g.ResolveTriggerPrompt(pr.ID, me.ID, false); err != nil {
			t.Fatalf("ResolveTriggerPrompt: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	for owner, id := range tops {
		if !gapB02InExile(g, id) {
			t.Errorf("the top card of %s was not exiled", owner)
		}
		if g.CastPermissionOnCardByIDForEffect(id) != nil {
			t.Errorf("a permission over the card from %s despite declining", owner)
		}
	}
}

func TestGapB02JesterIgnoresExilesThatAreNotItsControllersAbility(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := gapB02Active(g)
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "The Madcap Jester", "Legendary Creature — Clown Wizard", gapB02JesterOracle, false)
	// A sandbox/rule move of an opponent's library card into exile, with no
	// spell or ability of mine behind it.
	id := uuid.New()
	opp.Library.PushTop(game.Card{InstanceID: id, Name: "Stray", TypeLine: "Creature — Bear", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneLibrary, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneExile}, id); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	passPriorityAroundTable(t, g)
	if latestTriggerPromptFor(g, me.ID) != nil {
		t.Error("offered to play a card no spell or ability of mine exiled")
	}
	if g.CastPermissionOnCardByIDForEffect(id) != nil {
		t.Error("a permission over a card no spell or ability of mine exiled")
	}
}

// --- The Weaver King ----------------------------------------------------

func TestGapB02WeaverKingMillsThenTakesAMilledCreatureFromEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, a := gapB02Active(g)
	b := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	advanceToMain(t, g)
	weaver := pushCatalogPermanent(g, me.ID, "The Weaver King", "Legendary Creature — Horror Wizard", gapB02WeaverOracle, false)
	assertKeywords(t, g, weaver, "shadow")
	card := func(p *game.Player, name, typeLine string) uuid.UUID {
		id := uuid.New()
		p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
		return id
	}
	// B mills one creature earlier this turn: it counts.
	cb := card(b, "B's Milled Creature", "Creature — Bear")
	g.WithWriteLock(func() { _ = g.MillNForEffect(b.ID, 1) })
	// A's graveyard already holds a creature that arrived before the turn.
	old := uuid.New()
	a.Graveyard.PushTop(game.Card{InstanceID: old, Name: "Old Creature", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: a.ID, Controller: a.ID})
	c2 := card(a, "A's Second Creature", "Creature — Bear")
	l1 := card(a, "A's Land", "Basic Land — Forest")
	c1 := card(a, "A's First Creature", "Creature — Bear")
	dealCombatDamageToPlayer(g, weaver, a.ID, 3)
	passPriorityAroundTable(t, g)
	pr := chooseCardsChoiceFor(g, me.ID)
	if pr == nil {
		t.Fatal("no pick for the damaged opponent's graveyard")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pr.ChooseCards {
		offered[id] = true
	}
	if !offered[c1] || !offered[c2] || offered[l1] || offered[old] || len(offered) != 2 {
		t.Fatalf("offered %v, want only the two creatures milled this turn", pr.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, c2)
	pr = chooseCardsChoiceFor(g, me.ID)
	if pr == nil || len(pr.ChooseCards) != 1 || pr.ChooseCards[0] != cb {
		t.Fatalf("no pick of the other opponent's milled creature: %+v", pr)
	}
	answerChooseCards(t, g, me.ID, cb)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{c2, cb} {
		c := battlefieldCardCopy(t, g, id)
		if c.Controller != me.ID {
			t.Errorf("%s is under %s's control, want mine", c.Name, c.Controller)
		}
	}
	for _, c := range a.Graveyard.Cards {
		if c.InstanceID == c1 {
			return
		}
	}
	t.Error("the unchosen creature is not still in the graveyard")
}

func TestGapB02WeaverKingSkipsOpponentsWithNothingMilledThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, a := gapB02Active(g)
	advanceToMain(t, g)
	weaver := pushCatalogPermanent(g, me.ID, "The Weaver King", "Legendary Creature — Horror Wizard", gapB02WeaverOracle, false)
	// Every card A has is a land: nothing to take.
	for i := 0; i < 3; i++ {
		a.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Land", TypeLine: "Basic Land — Forest", Owner: a.ID, Controller: a.ID})
	}
	// A creature that died this turn is not "put there from their library".
	died := b12Creature(g, a.ID, "Died", "Creature — Bear", 1, 1)
	killCreature(t, g, a.ID, died)
	lib := a.Library.Size()
	dealCombatDamageToPlayer(g, weaver, a.ID, 2)
	passPriorityAroundTable(t, g)
	if got := lib - a.Library.Size(); got != 2 {
		t.Errorf("milled %d, want 2", got)
	}
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("asked to pick with no creature card milled")
	}
	if gapB02OnBattlefield(g, died) {
		t.Error("took a creature that died rather than was milled")
	}
}
