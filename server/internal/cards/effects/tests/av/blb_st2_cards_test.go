package testsav

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blb_st2_cards_test.go — Bloomburrow cards starting with S or T, batch 2
// (#4467). Each test plays the real catalog entry and asserts what its
// text says.

const (
	blbST2BrewOracle        = "ac7d1986-8388-4e4c-973e-d90382b12acf"
	blbST2WarrenguardOracle = "5bf1dbb6-ed65-4968-8edd-b98e9bd19d4d"
	blbST2LooterOracle      = "c15cad77-e70f-42cc-a9f5-50e1fd8c9c96"
	blbST2CharterOracle     = "578ddb51-e128-43e5-a115-471f0cae1222"
	blbST2SwordOracle       = "2cdc1d82-30dc-4f0a-af20-33dde90269b1"
	blbST2SentinelOracle    = "899f0931-4e9b-4fa8-9f6d-d13a7de05334"
	blbST2PantryOracle      = "068b370f-a358-4244-8050-6c81680c55c7"
	blbST2TalentOracle      = "00c1788d-c52f-4839-9cd8-11fb86c418b6"
	blbST2SplitterOracle    = "4a9ae0d9-5ee7-4289-9333-ef0bcc3079ca"
	blbST2TrashOracle       = "e07700e7-218a-45a3-9ee4-de2cc25cdc02"
	blbST2ShuckerOracle     = "33f4fcd5-4eea-4146-a564-da5511a8ee0a"
	blbST2WarlockOracle     = "25a63171-a1a0-4c61-ac61-49bb59fe83c1"
	blbST2ScribeOracle      = "7d2a7741-1dff-4fac-92d8-abea24f23f97"
	blbST2TrainerOracle     = "d0be5969-0442-4557-a5fc-bee56e1835c2"
	blbST2MentorOracle      = "e75e40a5-a9ae-4789-96c1-0e19d1ce59c5"
)

func TestBlbST2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		blbST2BrewOracle:        "Sazacap's Brew",
		blbST2WarrenguardOracle: "Seasoned Warrenguard",
		blbST2LooterOracle:      "Shoreline Looter",
		blbST2CharterOracle:     "Star Charter",
		blbST2SwordOracle:       "Starforged Sword",
		blbST2SentinelOracle:    "Stickytongue Sentinel",
		blbST2PantryOracle:      "Stocking the Pantry",
		blbST2TalentOracle:      "Stormchaser's Talent",
		blbST2SplitterOracle:    "Stormsplitter",
		blbST2TrashOracle:       "Take Out the Trash",
		blbST2ShuckerOracle:     "Thought Shucker",
		blbST2WarlockOracle:     "Thought-Stalker Warlock",
		blbST2ScribeOracle:      "Three Tree Scribe",
		blbST2TrainerOracle:     "Thundertrap Trainer",
		blbST2MentorOracle:      "Tidecaller Mentor",
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

// blbST2Graveyard puts n cards into p's graveyard.
func blbST2Graveyard(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		effectstest.PushGraveyardCardWithTypeLineID(p, "Old Card", "Instant")
	}
}

// blbST2Token creates one 2/2 Bear token for `owner`.
func blbST2Token(t *testing.T, g *game.Game, owner uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(owner, effects.TokenCard("2/2 green Bear"), 1); err != nil {
			t.Fatalf("create token: %v", err)
		}
	})
}

// blbST2ChangeLife changes a player's life by delta, as an effect.
func blbST2ChangeLife(t *testing.T, g *game.Game, player uuid.UUID, delta int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, player, delta); err != nil {
			t.Fatalf("change life: %v", err)
		}
	})
}

// blbST2Move moves a battlefield permanent to `to`.
func blbST2Move(t *testing.T, g *game.Game, id uuid.UUID, to game.ZoneRef) {
	t.Helper()
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, to, id); err != nil {
		t.Fatalf("move %s: %v", id, err)
	}
}

// --- Sazacap's Brew ----------------------------------------------------------

func blbST2CastBrew(t *testing.T, g *game.Game, targets []game.TargetRef, to uuid.UUID) (me *game.Player, fodder uuid.UUID) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	me = g.Seats[g.Turn.ActiveSeat]
	fodder = effectstest.HandCard(me, "Fodder", "Instant")
	id := effectstest.PushCatalogHandCard(me, "Sazacap's Brew", "Instant", blbST2BrewOracle)
	params := game.CastSpellParams{Targets: targets, DiscardIDs: []uuid.UUID{fodder}}
	if to != uuid.Nil {
		params.OptionalCosts = []int{effectstest.GiftIndex(t, blbST2BrewOracle)}
		params.GiftOpponent = to
	}
	if err := g.CastSpell(me.ID, id, params); err != nil {
		t.Fatalf("cast Sazacap's Brew: %v", err)
	}
	return me, fodder
}

func TestBlbST2SazacapsBrewPromisedDrawsAndPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	hand := me.Hand.Size()
	_, fodder := blbST2CastBrew(t, g, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: me.ID, Slot: 0}, {Kind: game.TargetCard, ID: bear, Slot: 1},
	}, opp.ID)
	if !me.Graveyard.Contains(fodder) {
		t.Fatal("the discard is an additional cost, paid on cast")
	}
	effectstest.PassPriorityAroundTable(t, g)
	// +1 fodder, +1 Brew, then -2 cast and discarded, +2 drawn.
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d, want %d", got, hand+2)
	}
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the bear is %d/%d, want 4/2", p, tough)
	}
	if n, tapped := effectstest.CountTokensNamed(g, opp.ID, "Fish"); n != 1 || tapped != 1 {
		t.Errorf("the opponent has %d Fish (%d tapped), want one tapped Fish", n, tapped)
	}
}

func TestBlbST2SazacapsBrewUnpromisedOnlyDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	hand := me.Hand.Size()
	blbST2CastBrew(t, g, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID, Slot: 0}}, uuid.Nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d, want %d", got, hand+2)
	}
	if p, _ := effectstest.PtOf(t, g, bear); p != 2 {
		t.Errorf("no gift: the bear has power %d, want 2", p)
	}
	if n, _ := effectstest.CountTokensNamed(g, opp.ID, "Fish"); n != 0 {
		t.Errorf("%d Fish without the gift, want 0", n)
	}
}

// --- Seasoned Warrenguard ----------------------------------------------------

func TestBlbST2SeasonedWarrenguardPumpsOnlyWhileYouControlAToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	guard := effectstest.BlbB1Push(g, me.ID, "Seasoned Warrenguard", "Creature — Rabbit Warrior", blbST2WarrenguardOracle, 1, 2)
	effectstest.DeclareAttack(t, g, opp.ID, guard)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, guard); p != 1 {
		t.Errorf("with no token the Warrenguard has power %d, want 1", p)
	}

	g = effectstest.NewCatalogGame(t)
	me, opp, _ = effectstest.BlbB1Seats(g)
	guard = effectstest.BlbB1Push(g, me.ID, "Seasoned Warrenguard", "Creature — Rabbit Warrior", blbST2WarrenguardOracle, 1, 2)
	blbST2Token(t, g, me.ID)
	effectstest.DeclareAttack(t, g, opp.ID, guard)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, guard); p != 3 || tough != 2 {
		t.Errorf("with a token the Warrenguard is %d/%d, want 3/2", p, tough)
	}
}

// --- Shoreline Looter --------------------------------------------------------

func TestBlbST2ShorelineLooterCantBeBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	looter := effectstest.B12Push(g, me.ID, "Shoreline Looter", "Creature — Rat Rogue", blbST2LooterOracle, 1, 1)
	effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.BrAttack(t, g, looter)
	if n := len(effectstest.BrOffered(t, g, opp.ID)); n != 0 {
		t.Errorf("the defender is offered %d blocks of an unblockable creature", n)
	}
}

func TestBlbST2ShorelineLooterLootsWithoutThreshold(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "Drawn")
	fodder := effectstest.HandCard(me, "Fodder", "Instant")
	looter := effectstest.B12Push(g, me.ID, "Shoreline Looter", "Creature — Rat Rogue", blbST2LooterOracle, 1, 1)
	blbST2Graveyard(me, 6)
	hand := me.Hand.Size()
	effectstest.DealCombatDamageToPlayer(g, looter, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("six cards in the graveyard: the draw should be followed by a discard")
	}
	effectstest.AnswerDiscard(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(fodder) || me.Hand.Size() != hand {
		t.Errorf("hand %d → %d, want one drawn and one discarded", hand, me.Hand.Size())
	}
}

func TestBlbST2ShorelineLooterJustDrawsAtThreshold(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "Drawn")
	effectstest.HandCard(me, "Fodder", "Instant")
	looter := effectstest.B12Push(g, me.ID, "Shoreline Looter", "Creature — Rat Rogue", blbST2LooterOracle, 1, 1)
	blbST2Graveyard(me, 7)
	hand := me.Hand.Size()
	effectstest.DealCombatDamageToPlayer(g, looter, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) != nil {
		t.Fatal("seven cards in the graveyard: no discard")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, me.Hand.Size())
	}
}

// --- Star Charter ------------------------------------------------------------

func blbST2CharterTable(t *testing.T) (*game.Game, *game.Player, []uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SamiLibraryTopFirst(me,
		game.Card{Name: "Big", TypeLine: "Creature — Beast", Power: 4, Toughness: 4},
		game.Card{Name: "Small", TypeLine: "Creature — Bat", Power: 3, Toughness: 1},
		game.Card{Name: "Spell", TypeLine: "Instant"},
		game.Card{Name: "Land", TypeLine: "Basic Land — Plains"},
		game.Card{Name: "Fifth", TypeLine: "Instant"},
	)
	charter := effectstest.B12Push(g, me.ID, "Star Charter", "Creature — Bat Cleric", blbST2CharterOracle, 3, 1)
	effectstest.AssertKeywords(t, g, charter, "flying")
	effectstest.AdvanceToMain(t, g)
	return g, me, ids
}

func TestBlbST2StarCharterTakesASmallCreatureAfterALifeChange(t *testing.T) {
	g, me, ids := blbST2CharterTable(t)
	blbST2ChangeLife(t, g, me.ID, -1)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no look-and-take prompt at your end step")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{ids[0]}); err == nil && me.Hand.Contains(ids[0]) {
		t.Fatal("a creature with power 4 was taken")
	}
	if c := effectstest.ChooseCardsChoiceFor(g, me.ID); c != nil {
		effectstest.AnswerChooseCards(t, g, me.ID, ids[1])
	}
	if !me.Hand.Contains(ids[1]) {
		t.Fatal("the power 3 creature did not reach the hand")
	}
	if top := me.Library.Cards[me.Library.Size()-1].InstanceID; top != ids[4] {
		t.Error("the fifth card must now be on top; the other three went to the bottom")
	}
}

func TestBlbST2StarCharterNeedsALifeChange(t *testing.T) {
	g, me, _ := blbST2CharterTable(t)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("no life gained or lost: the trigger should not happen")
	}
}

// --- Starforged Sword --------------------------------------------------------

func TestBlbST2StarforgedSwordPromisedAttachesOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bird := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird", Power: 1, Toughness: 1,
		Keywords: []string{"flying"}, Owner: me.ID, Controller: me.ID,
	})
	if !effectstest.HasKeywordOnBattlefield(t, g, bird, "flying") {
		t.Fatal("the bird should fly before it is equipped")
	}
	id := effectstest.MustCastWithGift(t, g, "Starforged Sword", "Artifact — Equipment", blbST2SwordOracle, nil, opp.ID)
	effectstest.ResolveCreatureSpell(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, bird)
	effectstest.AnswerTriggerOrderLastQueuedFirst(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AttachedTo(t, g, id, bird) {
		t.Fatal("the promised Sword attaches to the target")
	}
	if p, tough := effectstest.PtOf(t, g, bird); p != 4 || tough != 4 {
		t.Errorf("the equipped creature is %d/%d, want 4/4", p, tough)
	}
	if effectstest.HasKeywordOnBattlefield(t, g, bird, "flying") {
		t.Error("the equipped creature loses flying")
	}
	if n, tapped := effectstest.CountTokensNamed(g, opp.ID, "Fish"); n != 1 || tapped != 1 {
		t.Errorf("the opponent has %d Fish (%d tapped), want one tapped Fish", n, tapped)
	}
}

func TestBlbST2StarforgedSwordUnpromisedStaysUnattached(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	id := effectstest.MustCastWithGift(t, g, "Starforged Sword", "Artifact — Equipment", blbST2SwordOracle, nil, uuid.Nil)
	effectstest.ResolveCreatureSpell(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.TriggerOnStack(g, id) != nil {
		t.Fatal("an unpromised Sword puts no trigger on the stack (CR 603.4)")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, id); !ok || c.IsAttached() {
		t.Error("the unpromised Sword enters unattached")
	}
	if n, _ := effectstest.CountTokensNamed(g, opp.ID, "Fish"); n != 0 {
		t.Errorf("%d Fish without the gift, want 0", n)
	}
}

// --- Stickytongue Sentinel ---------------------------------------------------

func TestBlbST2StickytongueSentinelReturnsAnotherPermanentYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	sentinel := effectstest.CastHoldCreature(t, g, "Stickytongue Sentinel", "Creature — Frog Warrior", blbST2SentinelOracle, 3, 3)
	effectstest.AssertKeywords(t, g, sentinel, "reach")
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	for _, bad := range []uuid.UUID{theirs, sentinel} {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Fatalf("%s was accepted as a target", bad)
		}
	}
	effectstest.AnswerPickTarget(t, g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || !me.Hand.Contains(mine) {
		t.Error("your other permanent was not returned to your hand")
	}
	if !g.Battlefield.Contains(sentinel) || !g.Battlefield.Contains(theirs) {
		t.Error("the Sentinel or the opponent's creature left the battlefield")
	}
}

// --- Stocking the Pantry -----------------------------------------------------

func TestBlbST2StockingThePantryStoresSupplyAndDrawsOffIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	pantry := effectstest.PushCatalogPermanent(g, me.ID, "Stocking the Pantry", "Enchantment", blbST2PantryOracle, false)
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.FillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, pantry, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated with no supply counter")
	}
	effectstest.TsPut(t, g, me.ID, theirs, game.CounterPlusOne, 1)
	effectstest.TsPut(t, g, opp.ID, mine, game.CounterPlusOne, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, pantry, "supply"); n != 0 {
		t.Fatalf("%d supply counters from counters on an opponent's creature or put by an opponent, want 0", n)
	}
	effectstest.TsPut(t, g, me.ID, mine, game.CounterPlusOne, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, pantry, "supply"); n != 1 {
		t.Fatalf("%d supply counters, want 1 for one batch on one creature", n)
	}
	effectstest.SeedLibrary(me, "Drawn")
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, pantry, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want a card drawn", hand, me.Hand.Size())
	}
	if n := effectstest.CountersOn(g, pantry, "supply"); n != 0 {
		t.Errorf("%d supply counters left, want 0", n)
	}
}

// --- Stormchaser's Talent ----------------------------------------------------

func TestBlbST2StormchasersTalentMakesAnOtterOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Stormchaser's Talent", "Enchantment — Class", blbST2TalentOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	otters := effectstest.SocB8Named(g, me.ID, "Otter")
	if len(otters) != 1 {
		t.Fatalf("%d Otters, want 1", len(otters))
	}
	o := effectstest.LayeredCard(t, g, otters[0].InstanceID)
	if o.CurrentPower() != 1 || o.CurrentToughness() != 1 || !o.HasColor("U") || !o.HasColor("R") {
		t.Errorf("the Otter is %d/%d %v, want a 1/1 blue and red Otter", o.CurrentPower(), o.CurrentToughness(), o.Colors)
	}
	effectstest.AssertKeywords(t, g, otters[0].InstanceID, "prowess")
}

func TestBlbST2StormchasersTalentLevelsReturnASpellThenMakeOtters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	talent := effectstest.PushClass(g, me.ID, "Stormchaser's Talent", "Enchantment — Class", blbST2TalentOracle)
	bolt := effectstest.PushGraveyardCardWithTypeLineID(me, "Old Bolt", "Instant")
	bear := effectstest.PushGraveyardCardWithTypeLineID(me, "Old Bear", "Creature — Bear")
	effectstest.ToMainForCost(t, g)
	effectstest.SocB4Spell(t, g, "Early", "Instant", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.SocB8Named(g, me.ID, "Otter")); n != 0 {
		t.Fatalf("%d Otters from a spell at level 1, want 0", n)
	}

	if err := effectstest.LevelUpTo(t, g, me.ID, talent, 0, "{3}{U}"); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target asked for at level 2")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(bear)); err == nil {
		t.Fatal("a creature card was accepted")
	}
	effectstest.PickTriggerTarget(t, g, me.ID, bolt)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bolt) {
		t.Fatal("the instant is not back in hand")
	}

	effectstest.ToMainForCost(t, g)
	if err := effectstest.LevelUpTo(t, g, me.ID, talent, 1, "{5}{U}"); err != nil {
		t.Fatalf("level 3: %v", err)
	}
	effectstest.SocB4Spell(t, g, "Late", "Sorcery", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.SocB8Named(g, me.ID, "Otter")); n != 1 {
		t.Errorf("%d Otters from a sorcery at level 3, want 1", n)
	}
}

// --- Stormsplitter -----------------------------------------------------------

func TestBlbST2StormsplitterCopiesItselfAndExilesTheCopyAtEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	splitter := effectstest.B12Push(g, me.ID, "Stormsplitter", "Creature — Otter Wizard", blbST2SplitterOracle, 1, 4)
	effectstest.AssertKeywords(t, g, splitter, "haste")
	effectstest.SocB4Spell(t, g, "Bear Spell", "Creature — Bear", "{1}{G}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.SocB8Named(g, me.ID, "Stormsplitter")); n != 1 {
		t.Fatalf("%d Stormsplitters after a creature spell, want 1", n)
	}
	effectstest.SocB4Spell(t, g, "Shock", "Instant", "{R}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	got := effectstest.SocB8Named(g, me.ID, "Stormsplitter")
	if len(got) != 2 {
		t.Fatalf("%d Stormsplitters after an instant, want 2", len(got))
	}
	var token uuid.UUID
	for _, c := range got {
		if effects.IsToken(c) {
			token = c.InstanceID
		}
	}
	if token == uuid.Nil {
		t.Fatal("no token copy")
	}
	if c := effectstest.LayeredCard(t, g, token); c.CurrentPower() != 1 || c.CurrentToughness() != 4 {
		t.Errorf("the copy is %d/%d, want 1/4", c.CurrentPower(), c.CurrentToughness())
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(token) {
		t.Error("the token copy should be exiled at the beginning of the next end step")
	}
	if !g.Battlefield.Contains(splitter) {
		t.Error("the Stormsplitter itself stays")
	}
}

// --- Take Out the Trash ------------------------------------------------------

func TestBlbST2TakeOutTheTrashDealsThreeAndRummagesWithARaccoon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.SeedLibrary(me, "Drawn")
	fodder := effectstest.HandCard(me, "Fodder", "Instant")
	effectstest.BlbB1Push(g, me.ID, "Raccoon", "Creature — Raccoon Rogue", "", 1, 1)
	victim := effectstest.BlbB1Push(g, opp.ID, "Victim", "Creature — Bear", "", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Take Out the Trash", "Instant", blbST2TrashOracle, effectstest.CardTarget(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, victim); got != 3 {
		t.Errorf("the 3/3 has %d damage marked, want 3", got)
	}
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("with a Raccoon you may discard a card")
	}
	hand := me.Hand.Size()
	effectstest.AnswerDiscard(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("3 damage should kill the 3/3")
	}
	if !me.Graveyard.Contains(fodder) || me.Hand.Size() != hand {
		t.Errorf("hand %d → %d, want one discarded and one drawn", hand, me.Hand.Size())
	}
}

func TestBlbST2TakeOutTheTrashOffersNoRummageWithoutARaccoon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.HandCard(me, "Fodder", "Instant")
	victim := effectstest.BlbB1Push(g, opp.ID, "Victim", "Creature — Bear", "", 4, 4)
	effectstest.CastCatalogSpell(t, g, "Take Out the Trash", "Instant", blbST2TrashOracle, effectstest.CardTarget(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) != nil {
		t.Error("no Raccoon: no discard offered")
	}
	if got := effectstest.DamageMarkedOn(g, victim); got != 3 {
		t.Errorf("the 4/4 has %d damage marked, want 3", got)
	}
}

// --- Thought Shucker ---------------------------------------------------------

func TestBlbST2ThoughtShuckerNeedsThresholdAndActivatesOnlyOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B")
	shucker := effectstest.PushCatalogPermanent(g, me.ID, "Thought Shucker", "Creature — Rat Rogue", blbST2ShuckerOracle, false)
	effectstest.AdvanceToMain(t, g)
	blbST2Graveyard(me, 6)
	effectstest.FillPool(me, 1)
	effectstest.FillPoolColored(me, "U", 1)
	if err := g.ActivateCatalogAbility(me.ID, shucker, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated with six cards in the graveyard")
	}
	blbST2Graveyard(me, 1)
	hand := effectstest.HandSizeOf(g, me)
	if err := g.ActivateCatalogAbility(me.ID, shucker, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate at threshold: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandSizeOf(g, me); got != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, got)
	}
	if got := effectstest.PlusOnes(g, shucker); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	effectstest.FillPool(me, 1)
	effectstest.FillPoolColored(me, "U", 1)
	if err := g.ActivateCatalogAbility(me.ID, shucker, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrAbilityActivatedOnce) {
		t.Errorf("second activation: err = %v, want ErrAbilityActivatedOnce", err)
	}
}

// --- Thought-Stalker Warlock -------------------------------------------------

func blbST2CastWarlock(t *testing.T, g *game.Game, me, opp *game.Player) {
	t.Helper()
	warlock := effectstest.CastHoldCreature(t, g, "Thought-Stalker Warlock", "Creature — Lizard Warlock", blbST2WarlockOracle, 2, 2)
	effectstest.AssertKeywords(t, g, warlock, "menace")
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err == nil {
		t.Fatal("you were accepted as the target opponent")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("target the opponent: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestBlbST2ThoughtStalkerWarlockYouChooseANonlandWhenTheyLostLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	land := effectstest.HandCard(opp, "Their Land", "Basic Land — Swamp")
	spell := effectstest.HandCard(opp, "Their Spell", "Sorcery")
	blbST2ChangeLife(t, g, opp.ID, -2)
	blbST2CastWarlock(t, g, me, opp)
	var choice *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == me.ID && c.FromPlayer == opp.ID {
			choice = c
		}
	}
	if choice == nil {
		t.Fatalf("no revealed-hand pick for the Warlock's controller: %+v", g.PendingChoices)
	}
	if err := g.ResolvePendingChoice(choice.ID, me.ID, []uuid.UUID{land}); err == nil {
		t.Fatal("a land was accepted")
	}
	if err := g.ResolvePendingChoice(choice.ID, me.ID, []uuid.UUID{spell}); err != nil {
		t.Fatalf("pick the sorcery: %v", err)
	}
	g.SettleResolution()
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(spell) || !opp.Hand.Contains(land) {
		t.Error("the chosen nonland card should be discarded and the land kept")
	}
}

func TestBlbST2ThoughtStalkerWarlockTheyDiscardOtherwise(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	land := effectstest.HandCard(opp, "Their Land", "Basic Land — Swamp")
	effectstest.HandCard(opp, "Their Spell", "Sorcery")
	blbST2CastWarlock(t, g, me, opp)
	if effectstest.DiscardChoiceFor(g, opp.ID) == nil {
		t.Fatalf("no discard for the opponent who lost no life: %+v", g.PendingChoices)
	}
	effectstest.AnswerDiscard(t, g, opp.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(land) {
		t.Error("the opponent discards the card of their choice")
	}
}

// --- Three Tree Scribe -------------------------------------------------------

func TestBlbST2ThreeTreeScribeCountsCreaturesLeavingWithoutDying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	scribe := effectstest.B12Push(g, me.ID, "Three Tree Scribe", "Creature — Frog Druid", blbST2ScribeOracle, 2, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	ox := effectstest.B12Creature(g, me.ID, "Ox", "Creature — Ox", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)

	// A death is not "without dying", and an opponent's creature is not yours.
	blbST2Move(t, g, ox, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID})
	blbST2Move(t, g, theirs, game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a death or an opponent's bounce triggered the Scribe")
	}

	blbST2Move(t, g, bear, game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID})
	effectstest.PickTriggerTarget(t, g, me.ID, scribe)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, scribe); got != 1 {
		t.Fatalf("the Scribe has %d +1/+1 counters after a bounce, want 1", got)
	}
}

func TestBlbST2ThreeTreeScribeCountsItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	scribe := effectstest.B12Push(g, me.ID, "Three Tree Scribe", "Creature — Frog Druid", blbST2ScribeOracle, 2, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	blbST2Move(t, g, scribe, game.ZoneRef{Kind: game.ZoneExile})
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, bear); got != 1 {
		t.Errorf("the bear has %d +1/+1 counters after the Scribe was exiled, want 1", got)
	}
}

// --- Thundertrap Trainer -----------------------------------------------------

func TestBlbST2ThundertrapTrainerTakesANoncreatureNonlandCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SamiLibraryTopFirst(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Ring", TypeLine: "Artifact"},
		game.Card{Name: "Shock", TypeLine: "Instant"},
		game.Card{Name: "Fifth", TypeLine: "Instant"},
	)
	effectstest.CastHoldCreature(t, g, "Thundertrap Trainer", "Creature — Otter Wizard", blbST2TrainerOracle, 1, 2)
	c := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no look-and-take prompt on entry")
	}
	for _, bad := range []uuid.UUID{ids[0], ids[1]} {
		if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{bad}); err == nil && me.Hand.Contains(bad) {
			t.Fatalf("%s was taken", bad)
		}
	}
	if c := effectstest.ChooseCardsChoiceFor(g, me.ID); c != nil {
		effectstest.AnswerChooseCards(t, g, me.ID, ids[2])
	}
	if !me.Hand.Contains(ids[2]) {
		t.Fatal("the artifact did not reach the hand")
	}
	if top := me.Library.Cards[me.Library.Size()-1].InstanceID; top != ids[4] {
		t.Error("the fifth card must now be on top; the other three went to the bottom")
	}
}

// --- Tidecaller Mentor -------------------------------------------------------

func TestBlbST2TidecallerMentorBouncesAtThreshold(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	blbST2Graveyard(me, 7)
	mentor := effectstest.CastHoldCreature(t, g, "Tidecaller Mentor", "Creature — Rat Wizard", blbST2MentorOracle, 3, 3)
	effectstest.AssertKeywords(t, g, mentor, "menace")
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.AnswerPickTarget(t, g, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Hand.Contains(rock) {
		t.Error("the nonland permanent should return to its owner's hand")
	}
}

func TestBlbST2TidecallerMentorDoesNothingBelowThreshold(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	blbST2Graveyard(me, 6)
	effectstest.CastHoldCreature(t, g, "Tidecaller Mentor", "Creature — Rat Wizard", blbST2MentorOracle, 3, 3)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("six cards in the graveyard: the trigger should not happen")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) {
		t.Error("nothing returns below threshold")
	}
}
