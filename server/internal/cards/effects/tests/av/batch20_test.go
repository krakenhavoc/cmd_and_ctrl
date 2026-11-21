package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Two are rows
// in cycle tables (a Guildgate and a Dominaria United dual), so a
// transposed row is invisible until someone plays that exact card.
func TestBatch20CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B20HornOfGreedOracle:          "Horn of Greed",
		effectstest.B20FierceEmpathOracle:         "Fierce Empath",
		effectstest.B20TradingPostOracle:          "Trading Post",
		effectstest.B20TributeMageOracle:          "Tribute Mage",
		effectstest.B20GruulGuildgateOracle:       "Gruul Guildgate",
		effectstest.B20KwainOracle:                "Kwain, Itinerant Meddler",
		effectstest.B20GrappleWithThePastOracle:   "Grapple with the Past",
		effectstest.B20FontOfMythosOracle:         "Font of Mythos",
		effectstest.B20MasterOfEtheriumOracle:     "Master of Etherium",
		effectstest.B20ContaminatedAquiferOracle:  "Contaminated Aquifer",
		effectstest.B20GreedOracle:                "Greed",
		effectstest.B20DesolateLighthouseOracle:   "Desolate Lighthouse",
		effectstest.B20VerdantSunsAvatarOracle:    "Verdant Sun's Avatar",
		effectstest.B20GrindingStationOracle:      "Grinding Station",
		effectstest.B20PrimevalBountyOracle:       "Primeval Bounty",
		effectstest.B20CemeteryReaperOracle:       "Cemetery Reaper",
		effectstest.B20ProsperOracle:              "Prosper, Tome-Bound",
		effectstest.B20TimelessLotusOracle:        "Timeless Lotus",
		effectstest.B20CircuitMenderOracle:        "Circuit Mender",
		effectstest.B20StarfallInvocationOracle:   "Starfall Invocation",
		effectstest.B20MercilessExecutionerOracle: "Merciless Executioner",
		effectstest.B20TriplicateTitanOracle:      "Triplicate Titan",
		effectstest.B20AssembleTheLegionOracle:    "Assemble the Legion",
		effectstest.B20JetmirOracle:               "Jetmir, Nexus of Revels",
		effectstest.B20WallOfBlossomsOracle:       "Wall of Blossoms",
		effectstest.B20ManabarbsOracle:            "Manabarbs",
		effectstest.B20BrokenBondOracle:           "Broken Bond",
		// #1213 closed the return-a-Forest cost and gave the
		// once-each-turn gate a tally to read, so the batch's second
		// declared skip is a registered card now.
		"3ecaefc8-ead2-47a3-a7ea-b030faab65a7": "Quirion Ranger",
	}
	if len(want) != 28 {
		t.Fatalf("the batch registers 28 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
	// Arboreal Grazer was the last declared skip here: its put-a-land-from-hand
	// prompt (#654) now exists, so the card is built (cmdr2 batch 16, #5368)
	// and is asserted in cmdr2_b16_cards_test.go instead.
	//
	// Quirion Ranger LEFT this list in #1213: her return-a-Forest cost
	// is a component now and her "Activate only once each turn" reads
	// the activation tally, so nothing about her is stronger than
	// printed any more.
}

// --- the lands -----------------------------------------------------

func TestB20CycleRowsEnterTappedAndProduceTheirColours(t *testing.T) {
	for _, row := range []struct{ oracle, name, produced string }{
		{effectstest.B20GruulGuildgateOracle, "Gruul Guildgate", "{R|G}"},
		{effectstest.B20ContaminatedAquiferOracle, "Contaminated Aquifer", "{U|B}"},
	} {
		spec, ok := effects.Lookup(row.oracle)
		if !ok {
			t.Errorf("%s not registered", row.name)
			continue
		}
		if len(spec.Replacements) != 1 {
			t.Errorf("%s: %d replacements, want 1 (enters tapped)", row.name, len(spec.Replacements))
		}
		if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != row.produced {
			t.Errorf("%s: mana abilities %+v, want one producing %s", row.name, spec.ManaAbilities, row.produced)
		}
		g := effectstest.NewCatalogGame(t)
		land := effectstest.PlayLandFromHand(t, g, row.name, row.oracle)
		effectstest.Top100AssertEnteredTapped(t, g, land, row.name)
		if effectstest.TapEventsFor(g, land) != 0 {
			t.Errorf("%s: enters-tapped is a replacement, not a tap", row.name)
		}
	}
}

func TestB20DesolateLighthouseTapsForColorlessOrLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lighthouse := effectstest.B12Push(g, me.ID, "Desolate Lighthouse", "Land", effectstest.B20DesolateLighthouseOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, lighthouse, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	// Tapped for mana, it cannot loot this turn: the two abilities
	// share the tap.
	me.ManaPool.EmptyPool()
	effectstest.B06AddMana(me, "U", "R", "C")
	if err := g.ActivateCatalogAbility(me.ID, lighthouse, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("a tapped Lighthouse cannot pay {T} again")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(lighthouse) })
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, lighthouse, 0, game.ActivateAbilityParams{})
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("drew %d, want 1 before the discard", got-hand)
	}
	if effectstest.DiscardOwed(g, me.ID) != 1 {
		t.Errorf("discard owed = %d, want 1", effectstest.DiscardOwed(g, me.ID))
	}
	if !effectstest.B20Tapped(t, g, lighthouse) {
		t.Error("the loot has a tap cost")
	}
}

// --- the artifacts -------------------------------------------------

func TestB20TimelessLotusEntersTappedAndTapsForAllFiveColours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lotus := effectstest.CastCatalogSpell(t, g, "Timeless Lotus", "Legendary Artifact", effectstest.B20TimelessLotusOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B20Tapped(t, g, lotus) {
		t.Fatal("Timeless Lotus enters tapped")
	}
	if effectstest.TapEventsFor(g, lotus) != 0 {
		t.Error("enters-tapped is a replacement, not a tap")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(lotus) })
	if err := g.ActivateManaAbility(me.ID, lotus, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	got := effectstest.Batch01PoolColors(me)
	want := []string{"W", "U", "B", "R", "G"}
	if len(got) != 5 {
		t.Fatalf("pool %v, want one of each colour", got)
	}
	for i, c := range want {
		if got[i] != c {
			t.Errorf("pool %v, want %v", got, want)
			break
		}
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceMana {
			t.Error("five fixed colours need no colour pick")
		}
	}
}

func TestB20GreedPaysBlackAndTwoLifeForACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	greed := effectstest.B12Push(g, me.ID, "Greed", "Enchantment", effectstest.B20GreedOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, greed, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("{B} must be paid")
	}
	effectstest.B06AddMana(me, "B")
	life, hand := me.Life, me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, greed, 0, game.ActivateAbilityParams{})
	if me.Life != life-2 {
		t.Errorf("life %d → %d, want -2", life, me.Life)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("drew %d, want 1", got-hand)
	}
	// Below 2 life the life component cannot be paid.
	me.Life = 1
	effectstest.B06AddMana(me, "B")
	if err := g.ActivateCatalogAbility(me.ID, greed, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("at 1 life there is no 2 life to pay")
	}
}

func TestB20FontOfMythosDrawsTwoMoreOnEveryDrawStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Font of Mythos", "Artifact", effectstest.B20FontOfMythosOracle, 0, 0)
	// Seat 0's own draw step is already behind the cursor
	// (newCatalogGame parks there, #692), so the next seat's draw
	// step is the first one the Font sees — and it is an OPPONENT's,
	// which is the point: each player's draw step, not just the
	// controller's.
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.AdvanceTo(t, g, game.StepDraw)
	active := g.Seats[g.Turn.ActiveSeat]
	if active.ID == me.ID {
		t.Fatal("expected an opponent's draw step")
	}
	hands := effectstest.B20Hands(g)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		want := hands[i]
		if p.ID == active.ID {
			want += 2
		}
		if got := p.Hand.Size(); got != want {
			t.Errorf("seat %d hand %d → %d, want %d", i, hands[i], got, want)
		}
	}
}

func TestB20GrindingStationMillsThreeAndMayUntapWhenAnArtifactEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	station := effectstest.B12Push(g, me.ID, "Grinding Station", "Artifact", effectstest.B20GrindingStationOracle, 0, 0)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.AdvanceToMain(t, g)
	library := opp.Library.Size()
	effectstest.B16Activate(t, g, me.ID, station, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{rock},
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact is sacrificed as the cost")
	}
	if got := opp.Library.Size(); got != library-3 {
		t.Errorf("opponent's library %d → %d, want -3", library, got)
	}
	if !effectstest.B20Tapped(t, g, station) {
		t.Fatal("the Station has a tap cost")
	}
	// An artifact entering — anyone's — offers the untap.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.TreasureToken(), 1) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B20Tapped(t, g, station) {
		t.Error("the Station untaps when an artifact enters")
	}
	// A nonartifact entering is silent.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("3/3 green Beast"), 1) })
	if effectstest.B18TriggerPromptCount(g, me.ID) != 0 {
		t.Error("a Beast is not an artifact")
	}
	// The Station can eat itself.
	effectstest.B16Activate(t, g, me.ID, station, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{station},
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	if g.Battlefield.Contains(station) {
		t.Error("'sacrifice an artifact' includes the Station itself")
	}
	if got := opp.Library.Size(); got != library-6 {
		t.Errorf("opponent's library %d → %d, want -6", library, got)
	}
}

func TestB20TradingPostFourAbilities(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.B20TradingPostOracle)
	if !ok {
		t.Fatal("Trading Post not registered")
	}
	if len(spec.Activated) != 4 || spec.Completeness != effects.CompletenessFull {
		t.Fatalf("all four abilities ship, got %d / %v", len(spec.Activated), spec.Completeness)
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	post := effectstest.B12Push(g, me.ID, "Trading Post", "Artifact", effectstest.B20TradingPostOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)

	// {1}, {T}, Discard a card: gain 4 life.
	effectstest.B06AddMana(me, "C")
	toss := effectstest.HandCardForTest(me, "Spare Card", "Sorcery", "")
	lifeBefore := me.Life
	effectstest.B16Activate(t, g, me.ID, post, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{toss}})
	if me.Life != lifeBefore+4 {
		t.Errorf("discard for life: %d → %d, want +4", lifeBefore, me.Life)
	}
	if me.Hand.Contains(toss) {
		t.Error("the discarded card leaves the hand")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(post) })

	// {1}, {T}, Pay 1 life: a Goat.
	effectstest.B06AddMana(me, "C")
	life := me.Life
	effectstest.B16Activate(t, g, me.ID, post, 1, game.ActivateAbilityParams{})
	if me.Life != life-1 {
		t.Errorf("life %d → %d, want -1", life, me.Life)
	}
	goat := effectstest.FindBattlefieldByName(g, "Goat")
	if goat == uuid.Nil {
		t.Fatal("no Goat")
	}
	if c, _ := effectstest.BattlefieldCard(g, goat); c.Power != 0 || c.Toughness != 1 || !c.HasColor("W") {
		t.Error("a 0/1 white Goat")
	}

	// {1}, {T}, Sacrifice a creature: the Goat pays for an artifact
	// card back.
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(post) })
	effectstest.B06AddMana(me, "C")
	relic := effectstest.B17GraveyardCard(me, "Sol Ring", "Artifact", "{1}")
	effectstest.B16Activate(t, g, me.ID, post, 2, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{goat},
		Targets:      effectstest.B16TargetCard(relic),
	})
	if g.Battlefield.Contains(goat) {
		t.Error("the Goat is sacrificed as the cost")
	}
	if !me.Hand.Contains(relic) {
		t.Error("the artifact card comes back to hand")
	}

	// {1}, {T}, Sacrifice an artifact: the Post eats itself for a card.
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(post) })
	effectstest.B06AddMana(me, "C")
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, post, 3, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{post}})
	if g.Battlefield.Contains(post) {
		t.Error("'sacrifice an artifact' includes the Post itself")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("drew %d, want 1", got-hand)
	}
}

func TestB20HornOfGreedDrawsForTheLandPlayerOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Horn of Greed", "Artifact", effectstest.B20HornOfGreedOracle, 0, 0)

	// My land drop: the Horn draws me a card (the land left the
	// hand, so net zero before the draw).
	hands := effectstest.B20Hands(g)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hands[0]+1 {
		t.Errorf("my hand %d → %d, want +1", hands[0], got)
	}
	if got := opp.Hand.Size(); got != hands[1] {
		t.Error("only the player who played the land draws")
	}

	// A land put onto the battlefield by an effect is not played.
	buried := effectstest.B17GraveyardCard(me, "Island", "Basic Land — Island", "")
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.ReturnFromGraveyardUnderControlForEffect(buried, game.ZoneBattlefield, me.ID) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Error("a land returned from the graveyard by an effect is not a land play")
	}

	// An opponent's land drop draws the OPPONENT. (Walk to their main
	// phase first — their draw step is not the Horn's doing.)
	effectstest.AangAdvanceToMain(t, g, 1)
	hands = effectstest.B20Hands(g)
	effectstest.B13PlayAs(t, g, 1, "Mountain", "Basic Land — Mountain", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size(); got != hands[1]+1 {
		t.Errorf("opponent's hand %d → %d, want +1", hands[1], got)
	}
	if got := me.Hand.Size(); got != hands[0] {
		t.Error("the Horn's controller does not draw for an opponent's land")
	}
}

func TestB20HornOfGreedSeesALandPlayedFromExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Horn of Greed", "Artifact", effectstest.B20HornOfGreedOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	grant := game.CastPermission{Player: me.ID}
	swamp := effectstest.B20HandCard(me, "Swamp", "Basic Land — Swamp")
	g.WithWriteLock(func() { _ = g.ExileCardWithPermissionForEffect(swamp, grant) })
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, swamp, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("playing the exiled land: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d: a land played from exile is a land play", hand, got)
	}

	// #1326: a land RETURNED from exile by an effect is still not a
	// play, and the engine's own Event.Played marker (not a land-drop
	// tally guess) tells it apart from a later genuine play from
	// exile even in the same turn.
	plains := effectstest.B20HandCard(me, "Plains", "Basic Land — Plains")
	g.WithWriteLock(func() {
		_ = g.ExileCardForEffect(plains)
		_, _ = g.ReturnFromExileToBattlefieldForEffect(plains, me.ID, false)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d: a land returned from exile by an effect is not a land play", hand, got)
	}
	island := effectstest.B20HandCard(me, "Island", "Basic Land — Island")
	g.WithWriteLock(func() { _ = g.ExileCardWithPermissionForEffect(island, grant) })
	hand = me.Hand.Size()
	// #500: two land PLAYS in one turn, which the engine now refuses
	// by default. The subject here is the Horn's read of the play, not
	// CR 305.2, so the seat is given the second land play outright.
	me.LandDropsPerTurn++
	if err := g.CastSpell(me.ID, island, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("playing the second exiled land: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d: a second genuine land play from exile in the same turn still draws", hand, got)
	}
	if spec, _ := effects.Lookup(effectstest.B20HornOfGreedOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("the gap is closed (#1326): no caveat")
	}
}

func TestB20MasterOfEtheriumCountsArtifactsAndPumpsOtherArtifactCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	master := effectstest.B12Push(g, me.ID, "Master of Etherium", "Artifact Creature — Vedalken Wizard", effectstest.B20MasterOfEtheriumOracle, 0, 0)
	if effectstest.EffectivePower(t, g, master) != 1 || effectstest.EffectiveToughness(t, g, master) != 1 {
		t.Errorf("alone it counts itself: %d/%d, want 1/1", effectstest.EffectivePower(t, g, master), effectstest.EffectiveToughness(t, g, master))
	}
	myr := effectstest.B12Creature(g, me.ID, "Myr", "Artifact Creature — Myr", 1, 1)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Myr", "Artifact Creature — Myr", 1, 1)
	if got := effectstest.EffectivePower(t, g, master); got != 3 {
		t.Errorf("three artifacts: power %d, want 3", got)
	}
	if got := effectstest.EffectivePower(t, g, myr); got != 2 {
		t.Errorf("another artifact creature gets +1/+1: %d, want 2", got)
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("a nonartifact creature is untouched: %d, want 2", got)
	}
	if got := effectstest.EffectivePower(t, g, theirs); got != 1 {
		t.Errorf("an opponent's artifact creature is untouched: %d, want 1", got)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rock) })
	if got := effectstest.EffectivePower(t, g, master); got != 2 {
		t.Errorf("the count is live: %d, want 2", got)
	}
}

// --- the tutors ----------------------------------------------------

func TestB20FierceEmpathAndTributeMageOfferOnlyTheirFilter(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, want string
	}{
		{"Fierce Empath", "Creature — Elf", effectstest.B20FierceEmpathOracle, "Big Dragon"},
		{"Tribute Mage", "Creature — Human Wizard", effectstest.B20TributeMageOracle, "Arcane Signet"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.SeedSearchLibrary(me,
				game.Card{Name: "Big Dragon", TypeLine: "Creature — Dragon", ManaCost: "{4}{R}{R}"},
				game.Card{Name: "Small Elf", TypeLine: "Creature — Elf", ManaCost: "{G}"},
				game.Card{Name: "Arcane Signet", TypeLine: "Artifact", ManaCost: "{2}"},
				game.Card{Name: "Sol Ring", TypeLine: "Artifact", ManaCost: "{1}"},
				game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
			)
			effectstest.CastCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			c := effectstest.SearchChoiceFor(g, me.ID)
			if c == nil {
				t.Fatal("'you may search' opens the search prompt")
			}
			if len(c.SearchCards) != 1 || effectstest.SearchOptionNamed(g, c, tc.want) == uuid.Nil {
				t.Fatalf("the prompt offers exactly %q, got %d candidates", tc.want, len(c.SearchCards))
			}
			effectstest.AnswerSearchNamed(t, g, me.ID, tc.want)
			found := false
			for _, h := range me.Hand.Cards {
				if h.Name == tc.want {
					found = true
				}
			}
			if !found {
				t.Errorf("%s goes to hand", tc.want)
			}
		})
	}
}

// --- the draw and lifegain creatures -------------------------------

func TestB20WallOfBlossomsDefendsAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	wall := effectstest.CastCatalogSpell(t, g, "Wall of Blossoms", "Creature — Plant Wall", effectstest.B20WallOfBlossomsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d: added and cast (net 0), drew one (+1)", hand, got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, wall, "defender") {
		t.Error("defender is printed")
	}
}

func TestB20CircuitMenderGainsOnEntryAndDrawsOnAnyExit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	mender := effectstest.CastCatalogSpell(t, g, "Circuit Mender", "Artifact Creature — Insect", effectstest.B20CircuitMenderOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2", life, me.Life)
	}
	hand := me.Hand.Size()
	// Exile is not dying, and the Mender draws anyway.
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(mender) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d: leaving the battlefield draws", hand, got)
	}
}

func TestB20VerdantSunsAvatarGainsToughnessForItselfAndYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := me.Life
	// "This creature ... enters": the Avatar's own entry gains its
	// own toughness.
	effectstest.B20CastCreature(t, g, me, "Verdant Sun's Avatar", "Creature — Dinosaur Avatar", effectstest.B20VerdantSunsAvatarOracle, 5, 5)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+5 {
		t.Fatalf("its own entry: life %d → %d, want +5", life, me.Life)
	}
	// "Another creature you control": a 3/3 token.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("3/3 green Beast"), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+8 {
		t.Errorf("a 3/3 entering: life %d → %d, want +8", life, me.Life)
	}
	// Toughness is read as the trigger resolves: a creature that grew
	// while the trigger waited gains the grown amount.
	bear := effectstest.B20CastCreature(t, g, me, "Bear", "Creature — Bear", "", 2, 2)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, "+1/+1", 2) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+12 {
		t.Errorf("a 2/2 with two counters by resolution: life %d → %d, want +12", life, me.Life)
	}
	// An opponent's creature and a noncreature are silent.
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(opp.ID, effects.TokenCard("3/3 green Beast"), 1)
		_ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 1)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+12 {
		t.Error("an opponent's creature and a Treasure gain nothing")
	}
}

func TestB20KwainEachPlayerMayDrawAndOnlyDrawersGain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	kwain := effectstest.B12Push(g, me.ID, "Kwain, Itinerant Meddler", "Legendary Creature — Rabbit Wizard", effectstest.B20KwainOracle, 1, 3)
	empty := g.Seats[3]
	empty.Library.Cards = nil
	effectstest.AdvanceToMain(t, g)
	hands, lives := effectstest.B20Hands(g), effectstest.B20Lives(g)
	effectstest.B16Activate(t, g, me.ID, kwain, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	// Asked in APNAP order: seat 0 (the active player) first. Seat 1
	// declines, seat 2 draws, seat 3 (no library) says yes.
	answers := []bool{true, false, true, true}
	for i, yes := range answers {
		effectstest.AnswerMayChoice(t, g, g.Seats[i].ID, yes)
	}
	want := []struct{ hand, life int }{{1, 1}, {0, 0}, {1, 1}, {0, 0}}
	for i, p := range g.Seats {
		if p.ID == empty.ID {
			continue
		}
		if p.Hand.Size() != hands[i]+want[i].hand || p.Life != lives[i]+want[i].life {
			t.Errorf("seat %d: hand %d life %d, want hand %d life %d", i, p.Hand.Size(), p.Life,
				hands[i]+want[i].hand, lives[i]+want[i].life)
		}
	}
	if !empty.Eliminated && !empty.AttemptedEmptyDraw {
		t.Error("a player with no library who says yes tries to draw from nothing")
	}
	if empty.Life != lives[3] && !empty.Eliminated {
		t.Error("and gains no life for a card they did not draw")
	}
	if spec, _ := effects.Lookup(effectstest.B20KwainOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("the per-player 'may' is asked now")
	}
}

// --- the spells ----------------------------------------------------

func TestB20BrokenBondDestroysAnArtifactOrEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	bear := effectstest.B16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.CastCatalogSpell(t, g, "Broken Bond", "Sorcery", effectstest.B20BrokenBondOracle, effectstest.B16TargetCard(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Graveyard.Contains(rock) {
		t.Error("the artifact is destroyed")
	}
	if err := b09TryCast(t, g, "Broken Bond", "Sorcery", effectstest.B20BrokenBondOracle, effectstest.B16TargetCard(bear)); err == nil {
		t.Error("a creature is not an artifact or enchantment")
	}
	if spec, _ := effects.Lookup(effectstest.B20BrokenBondOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("the put-a-land clause landed with #654; nothing is deferred any more")
	}
}

func TestB20StarfallInvocationDestroysAllCreaturesAndDeclaresTheGift(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Starfall Invocation", "Sorcery", effectstest.B20StarfallInvocationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("every creature is destroyed")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("a noncreature survives")
	}
	// S30 (#470 / #446): the indestructible half of this assertion was
	// an engine gap, not a card gap, and it is closed. The gift landed
	// with #3172 (cav-b): the card is Full and the return is tested in
	// cav_b_test.go.
	spec, _ := effects.Lookup(effectstest.B20StarfallInvocationOracle)
	if spec.Completeness != effects.CompletenessFull || spec.Gift == nil {
		t.Error("Starfall Invocation offers its gift and has no caveat left")
	}
}

// --- the edict and the dies trigger --------------------------------

func TestB20MercilessExecutionerMakesEveryoneSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	executioner := effectstest.CastCatalogSpell(t, g, "Merciless Executioner", "Creature — Orc Warrior", effectstest.B20MercilessExecutionerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// Each player has exactly one creature — the Executioner is its own
	// controller's — so every pick is forced and nobody is asked (ADR
	// 0175, #3664).
	if effectstest.SacrificeChoiceFor(g, g.Seats[2].ID) != nil {
		t.Error("a player with no creature is skipped, not prompted")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(executioner) || g.Battlefield.Contains(theirs) {
		t.Error("the sacrifices happen")
	}
}

func TestB20TriplicateTitanLeavesThreeDifferentGolems(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	titan := effectstest.B12Push(g, me.ID, "Triplicate Titan", "Artifact Creature — Golem", effectstest.B20TriplicateTitanOracle, 9, 9)
	for _, kw := range []string{"flying", "vigilance", "trample"} {
		if !effectstest.HasEffectiveKeyword(t, g, titan, kw) {
			t.Errorf("%s is printed", kw)
		}
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B18Kill(t, g, titan)
	golems := effectstest.BattlefieldIDsNamed(g, "Golem")
	if len(golems) != 3 {
		t.Fatalf("%d Golems, want 3", len(golems))
	}
	seen := map[string]int{}
	for _, id := range golems {
		c, _ := effectstest.BattlefieldCard(g, id)
		if c.Power != 3 || c.Toughness != 3 || !c.IsArtifact() || !c.IsCreature() {
			t.Errorf("a 3/3 artifact creature, got %s %d/%d", c.TypeLine, c.Power, c.Toughness)
		}
		for _, kw := range []string{"flying", "vigilance", "trample"} {
			if effectstest.HasEffectiveKeyword(t, g, id, kw) {
				seen[kw]++
			}
		}
	}
	for _, kw := range []string{"flying", "vigilance", "trample"} {
		if seen[kw] != 1 {
			t.Errorf("exactly one Golem has %s, got %d", kw, seen[kw])
		}
	}
	// Exile is not dying.
	titan2 := effectstest.B12Push(g, me.ID, "Triplicate Titan", "Artifact Creature — Golem", effectstest.B20TriplicateTitanOracle, 9, 9)
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(titan2) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.BattlefieldIDsNamed(g, "Golem")); n != 3 {
		t.Errorf("%d Golems after an exile, want still 3", n)
	}
}

// --- the enchantments ----------------------------------------------

func TestB20AssembleTheLegionMustersOneMoreSoldierEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	legion := effectstest.B12Push(g, me.ID, "Assemble the Legion", "Enchantment", effectstest.B20AssembleTheLegionOracle, 0, 0)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if effectstest.B16CountNamed(g, "Soldier") != 0 {
		t.Fatal("an opponent's upkeep is not yours")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B18Counters(t, g, legion, "muster"); n != 1 {
		t.Errorf("%d muster counters, want 1", n)
	}
	if n := effectstest.B16CountNamed(g, "Soldier"); n != 1 {
		t.Fatalf("%d Soldiers after the first upkeep, want 1", n)
	}
	soldier := effectstest.FindBattlefieldByName(g, "Soldier")
	if !effectstest.HasEffectiveKeyword(t, g, soldier, "haste") {
		t.Error("the Soldier has haste")
	}
	if c, _ := effectstest.BattlefieldCard(g, soldier); !c.HasColor("R") || !c.HasColor("W") {
		t.Error("the Soldier is red and white")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B18Counters(t, g, legion, "muster"); n != 2 {
		t.Errorf("%d muster counters, want 2", n)
	}
	if n := effectstest.B16CountNamed(g, "Soldier"); n != 3 {
		t.Errorf("%d Soldiers after the second upkeep, want 3 (1 + 2)", n)
	}
}

func TestB20PrimevalBountyThreeTriggers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Primeval Bounty", "Enchantment", effectstest.B20PrimevalBountyOracle, 0, 0)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	// A creature spell: a Beast.
	effectstest.CastCatalogSpell(t, g, "Grizzly", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16CountNamed(g, "Beast") != 1 {
		t.Error("a creature spell makes a 3/3 Beast")
	}
	// A noncreature spell: three counters on a creature you control,
	// picked when the trigger fires.
	effectstest.CastCatalogSpell(t, g, "Shock", "Instant", "", nil)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B18Counters(t, g, bear, "+1/+1"); n != 3 {
		t.Errorf("%d +1/+1 counters, want 3", n)
	}
	if effectstest.B16CountNamed(g, "Beast") != 1 {
		t.Error("a noncreature spell makes no Beast")
	}
	// Landfall: 3 life.
	life := me.Life
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3", life, me.Life)
	}
}

func TestB20ManabarbsPingsWhoeverTapsALandForMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Manabarbs", "Enchantment", effectstest.B20ManabarbsOracle, 0, 0)
	mine := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := effectstest.SeedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island")
	effectstest.AdvanceToMain(t, g)
	lives := effectstest.B20Lives(g)
	if err := g.ActivateManaAbility(me.ID, mine, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != lives[0]-1 {
		t.Errorf("my life %d → %d, want -1: the controller is a player too", lives[0], me.Life)
	}
	if err := g.ActivateManaAbility(opp.ID, theirs, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("opponent ActivateManaAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != lives[1]-1 {
		t.Errorf("opponent's life %d → %d, want -1", lives[1], opp.Life)
	}
	if me.Life != lives[0]-1 {
		t.Error("the damage goes to the player who tapped, not the controller")
	}
	// A nonland mana source is silent.
	ring := effectstest.B12Push(g, me.ID, "Sol Ring", "Artifact", "6ad8011d-3471-4369-9d68-b264cc027487", 0, 0)
	if err := g.ActivateManaAbility(me.ID, ring, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("Sol Ring: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != lives[0]-1 {
		t.Error("an artifact tapped for mana is not a land")
	}
}

func TestB20JetmirScalesWithThreeSixAndNineCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	jetmir := effectstest.B12Push(g, me.ID, "Jetmir, Nexus of Revels", "Legendary Creature — Cat Demon", effectstest.B20JetmirOracle, 5, 4)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	if effectstest.EffectivePower(t, g, bear) != 2 || effectstest.HasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Fatal("two creatures: no bonus yet")
	}
	effectstest.B16Creature(g, me.ID, "Bear 3", "Creature — Bear", 2, 2, "G")
	if effectstest.EffectivePower(t, g, bear) != 3 || !effectstest.HasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Errorf("three creatures: +1/+0 and vigilance, got %d", effectstest.EffectivePower(t, g, bear))
	}
	if effectstest.EffectivePower(t, g, jetmir) != 6 || effectstest.EffectiveToughness(t, g, jetmir) != 4 {
		t.Errorf("Jetmir counts and buffs himself: %d/%d, want 6/4", effectstest.EffectivePower(t, g, jetmir), effectstest.EffectiveToughness(t, g, jetmir))
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("trample waits for six")
	}
	for i := 4; i <= 6; i++ {
		effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	}
	if effectstest.EffectivePower(t, g, bear) != 4 || !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Errorf("six creatures: +2/+0 and trample, got %d", effectstest.EffectivePower(t, g, bear))
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "double strike") {
		t.Error("double strike waits for nine")
	}
	for i := 7; i <= 9; i++ {
		effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	}
	if effectstest.EffectivePower(t, g, bear) != 5 || !effectstest.HasEffectiveKeyword(t, g, bear, "double strike") {
		t.Errorf("nine creatures: +3/+0 and double strike, got %d", effectstest.EffectivePower(t, g, bear))
	}
	// An opponent's creature is not "creatures you control", and does
	// not count.
	theirs := effectstest.B16Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	if effectstest.EffectivePower(t, g, theirs) != 2 {
		t.Error("an opponent's creature is untouched")
	}
	// Losing one drops the line at once.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(jetmir) })
	if effectstest.EffectivePower(t, g, bear) != 2 {
		t.Errorf("without Jetmir the bonus is gone: %d", effectstest.EffectivePower(t, g, bear))
	}
}

// --- Cemetery Reaper -----------------------------------------------

func TestB20CemeteryReaperLordsZombiesAndMakesThemFromGraveyards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	reaper := effectstest.B12Push(g, me.ID, "Cemetery Reaper", "Creature — Zombie", effectstest.B20CemeteryReaperOracle, 2, 2)
	zombie := effectstest.B16Creature(g, me.ID, "Shambler", "Creature — Zombie", 2, 2, "B")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Zombie", "Creature — Zombie", 2, 2, "B")
	if effectstest.EffectivePower(t, g, zombie) != 3 {
		t.Errorf("other Zombies you control get +1/+1: %d", effectstest.EffectivePower(t, g, zombie))
	}
	if effectstest.EffectivePower(t, g, reaper) != 2 || effectstest.EffectivePower(t, g, theirs) != 2 {
		t.Error("the Reaper itself and an opponent's Zombie are untouched")
	}
	effectstest.AdvanceToMain(t, g)
	dead := effectstest.B17GraveyardCard(opp, "Their Bear", "Creature — Bear", "{1}{G}")
	effectstest.B06AddMana(me, "B", "C", "C")
	effectstest.B16Activate(t, g, me.ID, reaper, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(dead)})
	if !effectstest.InExile(g, dead) {
		t.Error("the creature card is exiled from the opponent's graveyard")
	}
	if n := effectstest.B16CountNamed(g, "Zombie"); n != 1 {
		t.Fatalf("%d Zombie tokens, want 1", n)
	}
	token := effectstest.FindBattlefieldByName(g, "Zombie")
	if effectstest.EffectivePower(t, g, token) != 3 {
		t.Error("the new Zombie is a 2/2 that the Reaper lords to 3/3")
	}
	if !effectstest.B20Tapped(t, g, reaper) {
		t.Error("the ability has a tap cost")
	}
}

// --- Prosper, Tome-Bound -------------------------------------------

func TestB20ProsperExilesAtYourEndStepAndPaysTreasureForPlaysFromExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	prosper := effectstest.B12Push(g, me.ID, "Prosper, Tome-Bound", "Legendary Creature — Tiefling Warlock", effectstest.B20ProsperOracle, 1, 4)
	if !effectstest.HasEffectiveKeyword(t, g, prosper, "deathtouch") {
		t.Error("deathtouch is printed")
	}
	// The first turn's end step exiles whatever filler is on top;
	// seed the real library after it so the walk below is about the
	// second round. The Filler is what seat 0's next draw step takes,
	// leaving the Forest on top for the end step.
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(g.Exile.Cards); n != 1 {
		t.Fatalf("%d cards in exile after your end step, want 1", n)
	}
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Filler", TypeLine: "Sorcery"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Deep Card", TypeLine: "Sorcery"},
	)
	forest := ids[1]
	// An opponent's end step is not yours.
	effectstest.AdvanceToEndStepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile.Contains(forest) {
		t.Fatal("Mystic Arcanum fires on the controller's own end step only")
	}
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(forest) {
		t.Fatal("the top card is exiled at your end step")
	}
	perm := effectstest.ExiledPermission(g, forest)
	if perm.Player != me.ID || perm.CastOnly || !effectstest.PermissionLive(g, perm, me.ID) {
		t.Fatalf("grant %+v: the controller may PLAY it", perm)
	}
	// Through the opponents' turns the grant survives, and it reaches
	// the END of the controller's next turn (#945: one duration, no
	// upkeep re-stamp).
	effectstest.AdvanceToMainOf(t, g, 3)
	if !effectstest.PermissionLive(g, effectstest.ExiledPermission(g, forest), me.ID) {
		t.Fatal("the grant survives every opponent's cleanup")
	}
	// Playing the land from exile is "playing a card from exile":
	// Pact Boon makes a Treasure.
	effectstest.AdvanceToMainOf(t, g, 0)
	if p := effectstest.ExiledPermission(g, forest); !effectstest.PermissionLive(g, p, me.ID) {
		t.Fatalf("on your next turn the grant is %+v, want live", p)
	}
	if err := g.CastSpell(me.ID, forest, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("playing the exiled land: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(forest) {
		t.Fatal("the land is played")
	}
	if n := effectstest.B16CountNamed(g, "Treasure"); n != 1 {
		t.Errorf("%d Treasures, want 1", n)
	}
	// A spell cast from exile under any grant does too.
	shock := effectstest.B20HandCard(me, "Shock", "Instant")
	g.WithWriteLock(func() {
		_ = g.ExileCardWithPermissionForEffect(shock, game.CastPermission{Player: me.ID})
	})
	if err := g.CastSpell(me.ID, shock, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("casting from exile: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B16CountNamed(g, "Treasure"); n != 2 {
		t.Errorf("%d Treasures, want 2", n)
	}
	// A spell cast from hand is not.
	effectstest.CastCatalogSpell(t, g, "Bolt", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B16CountNamed(g, "Treasure"); n != 2 {
		t.Error("a cast from hand makes no Treasure")
	}
	// A card exiled at the end step with an empty library: no grant,
	// no crash.
	me.Library.Cards = nil
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
}
