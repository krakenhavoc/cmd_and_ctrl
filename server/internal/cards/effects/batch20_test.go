package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch20_test.go — card-level coverage for the card-coverage
// roadmap's batch 20 (#313, `edhrec_rank` 2133–2234): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, land play or step change.
// Helpers from the earlier batch test files are reused by name; new
// ones are b20-prefixed.

const (
	b20HornOfGreedOracle          = "b8181d53-1954-4f46-8670-8696440208e8"
	b20FierceEmpathOracle         = "5104053a-d394-4b00-82a4-60fd9f051a6e"
	b20TradingPostOracle          = "63788566-e25a-44bb-bb55-197e1b93b3e8"
	b20TributeMageOracle          = "c2b51295-8820-425f-a517-55231123c7de"
	b20GruulGuildgateOracle       = "d38476e9-2e47-4c0c-8129-483c0bd09ec0"
	b20KwainOracle                = "ce6a85b5-e249-4582-a6be-ce200da5fd53"
	b20GrappleWithThePastOracle   = "d14c633d-977a-4f78-85c2-88582b3c3670"
	b20FontOfMythosOracle         = "7194a262-5e7a-4c12-b271-2bc5e0799477"
	b20MasterOfEtheriumOracle     = "181e6b3e-c33e-49e7-acb8-67473289a856"
	b20ContaminatedAquiferOracle  = "c27b771d-b5ec-459a-a101-f078cb8d0184"
	b20GreedOracle                = "1ff62220-be95-4901-b8d8-812b9a1a1b0a"
	b20DesolateLighthouseOracle   = "aa6dbdf2-2379-4ff5-8a6c-70258784dc35"
	b20VerdantSunsAvatarOracle    = "6ee71432-a44d-495e-86ea-d495d891c331"
	b20GrindingStationOracle      = "0fcd476f-4db8-4293-9388-1678a0043c9e"
	b20PrimevalBountyOracle       = "f633c16d-d943-421c-ad18-af35db9ec9fc"
	b20CemeteryReaperOracle       = "70fa209f-5e79-44de-81ce-1d8d0d8c1006"
	b20ProsperOracle              = "1e9b0fd6-5aae-401a-8df5-d94ce6442696"
	b20TimelessLotusOracle        = "fec77e0a-433e-4d32-9260-e26a37f6ad23"
	b20CircuitMenderOracle        = "1665ca9f-176d-40f1-a4e9-42da4f1236e9"
	b20StarfallInvocationOracle   = "7024532b-f99b-43a7-b0ed-5b3e7ec7592b"
	b20MercilessExecutionerOracle = "c3c45d50-9038-41df-bb2f-9bc40071845b"
	b20TriplicateTitanOracle      = "56e05b17-aa1d-483d-857c-4cab4f12aa8a"
	b20AssembleTheLegionOracle    = "6f81bef3-6ca0-4cf4-aa99-7b2813eeef04"
	b20JetmirOracle               = "da72a4bc-ce6f-4b72-bc66-2ee33cfa87df"
	b20WallOfBlossomsOracle       = "ef4d5fb3-70a3-433d-a9d3-18b2beb8d79f"
	b20ManabarbsOracle            = "0f1afedd-c60f-454f-b84a-c8117aec0128"
	b20BrokenBondOracle           = "858e12e9-3eaa-40cf-9e22-f9ccdfe485b3"
)

// b20Lives snapshots every seat's life total.
func b20Lives(g *game.Game) []int {
	out := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		out[i] = p.Life
	}
	return out
}

// b20Hands snapshots every seat's hand size.
func b20Hands(g *game.Game) []int {
	out := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		out[i] = p.Hand.Size()
	}
	return out
}

// b20HandCard puts a card in a player's hand.
func b20HandCard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

// b20CastCreature puts a creature with a printed P/T in a player's
// hand and casts it from a main phase of their turn.
func b20CastCreature(t *testing.T, g *game.Game, p *game.Player, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// b20Tapped reads a battlefield card's tapped state, failing when it
// is not there.
func b20Tapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Tapped
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Two are rows
// in cycle tables (a Guildgate and a Dominaria United dual), so a
// transposed row is invisible until someone plays that exact card.
func TestBatch20CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b20HornOfGreedOracle:          "Horn of Greed",
		b20FierceEmpathOracle:         "Fierce Empath",
		b20TradingPostOracle:          "Trading Post",
		b20TributeMageOracle:          "Tribute Mage",
		b20GruulGuildgateOracle:       "Gruul Guildgate",
		b20KwainOracle:                "Kwain, Itinerant Meddler",
		b20GrappleWithThePastOracle:   "Grapple with the Past",
		b20FontOfMythosOracle:         "Font of Mythos",
		b20MasterOfEtheriumOracle:     "Master of Etherium",
		b20ContaminatedAquiferOracle:  "Contaminated Aquifer",
		b20GreedOracle:                "Greed",
		b20DesolateLighthouseOracle:   "Desolate Lighthouse",
		b20VerdantSunsAvatarOracle:    "Verdant Sun's Avatar",
		b20GrindingStationOracle:      "Grinding Station",
		b20PrimevalBountyOracle:       "Primeval Bounty",
		b20CemeteryReaperOracle:       "Cemetery Reaper",
		b20ProsperOracle:              "Prosper, Tome-Bound",
		b20TimelessLotusOracle:        "Timeless Lotus",
		b20CircuitMenderOracle:        "Circuit Mender",
		b20StarfallInvocationOracle:   "Starfall Invocation",
		b20MercilessExecutionerOracle: "Merciless Executioner",
		b20TriplicateTitanOracle:      "Triplicate Titan",
		b20AssembleTheLegionOracle:    "Assemble the Legion",
		b20JetmirOracle:               "Jetmir, Nexus of Revels",
		b20WallOfBlossomsOracle:       "Wall of Blossoms",
		b20ManabarbsOracle:            "Manabarbs",
		b20BrokenBondOracle:           "Broken Bond",
		// #1213 closed the return-a-Forest cost and gave the
		// once-each-turn gate a tally to read, so the batch's second
		// declared skip is a registered card now.
		"3ecaefc8-ead2-47a3-a7ea-b030faab65a7": "Quirion Ranger",
	}
	if len(want) != 28 {
		t.Fatalf("the batch registers 28 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
	// The remaining declared skip must NOT be registered: Arboreal
	// Grazer's put-a-land-from-hand prompt. A spec for it would ship
	// the card stronger than printed.
	//
	// Quirion Ranger LEFT this list in #1213: her return-a-Forest cost
	// is a component now and her "Activate only once each turn" reads
	// the activation tally, so nothing about her is stronger than
	// printed any more.
	for _, skipped := range []string{
		"d18a0815-59d3-4667-b52b-9acda741215e", // Arboreal Grazer
	} {
		if _, ok := Lookup(skipped); ok {
			t.Errorf("%s is a declared skip and must not be registered", skipped)
		}
	}
}

// --- the lands -----------------------------------------------------

func TestB20CycleRowsEnterTappedAndProduceTheirColours(t *testing.T) {
	for _, row := range []struct{ oracle, name, produced string }{
		{b20GruulGuildgateOracle, "Gruul Guildgate", "{R|G}"},
		{b20ContaminatedAquiferOracle, "Contaminated Aquifer", "{U|B}"},
	} {
		spec, ok := Lookup(row.oracle)
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
		g := newCatalogGame(t)
		land := playLandFromHand(t, g, row.name, row.oracle)
		top100AssertEnteredTapped(t, g, land, row.name)
		if tapEventsFor(g, land) != 0 {
			t.Errorf("%s: enters-tapped is a replacement, not a tap", row.name)
		}
	}
}

func TestB20DesolateLighthouseTapsForColorlessOrLoots(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lighthouse := b12Push(g, me.ID, "Desolate Lighthouse", "Land", b20DesolateLighthouseOracle, 0, 0)
	advanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, lighthouse, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	// Tapped for mana, it cannot loot this turn: the two abilities
	// share the tap.
	me.ManaPool.EmptyPool()
	b06AddMana(me, "U", "R", "C")
	if err := g.ActivateCatalogAbility(me.ID, lighthouse, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("a tapped Lighthouse cannot pay {T} again")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(lighthouse) })
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, lighthouse, 0, game.ActivateAbilityParams{})
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("drew %d, want 1 before the discard", got-hand)
	}
	if discardOwed(g, me.ID) != 1 {
		t.Errorf("discard owed = %d, want 1", discardOwed(g, me.ID))
	}
	if !b20Tapped(t, g, lighthouse) {
		t.Error("the loot has a tap cost")
	}
}

// --- the artifacts -------------------------------------------------

func TestB20TimelessLotusEntersTappedAndTapsForAllFiveColours(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lotus := castCatalogSpell(t, g, "Timeless Lotus", "Legendary Artifact", b20TimelessLotusOracle, nil)
	passPriorityAroundTable(t, g)
	if !b20Tapped(t, g, lotus) {
		t.Fatal("Timeless Lotus enters tapped")
	}
	if tapEventsFor(g, lotus) != 0 {
		t.Error("enters-tapped is a replacement, not a tap")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(lotus) })
	if err := g.ActivateManaAbility(me.ID, lotus, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	got := batch01PoolColors(me)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	greed := b12Push(g, me.ID, "Greed", "Enchantment", b20GreedOracle, 0, 0)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, greed, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("{B} must be paid")
	}
	b06AddMana(me, "B")
	life, hand := me.Life, me.Hand.Size()
	b16Activate(t, g, me.ID, greed, 0, game.ActivateAbilityParams{})
	if me.Life != life-2 {
		t.Errorf("life %d → %d, want -2", life, me.Life)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("drew %d, want 1", got-hand)
	}
	// Below 2 life the life component cannot be paid.
	me.Life = 1
	b06AddMana(me, "B")
	if err := g.ActivateCatalogAbility(me.ID, greed, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("at 1 life there is no 2 life to pay")
	}
}

func TestB20FontOfMythosDrawsTwoMoreOnEveryDrawStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Font of Mythos", "Artifact", b20FontOfMythosOracle, 0, 0)
	// Seat 0's own draw step is already behind the cursor
	// (newCatalogGame parks there, #692), so the next seat's draw
	// step is the first one the Font sees — and it is an OPPONENT's,
	// which is the point: each player's draw step, not just the
	// controller's.
	advanceTo(t, g, game.StepEnd)
	advanceTo(t, g, game.StepDraw)
	active := g.Seats[g.Turn.ActiveSeat]
	if active.ID == me.ID {
		t.Fatal("expected an opponent's draw step")
	}
	hands := b20Hands(g)
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	station := b12Push(g, me.ID, "Grinding Station", "Artifact", b20GrindingStationOracle, 0, 0)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	advanceToMain(t, g)
	library := opp.Library.Size()
	b16Activate(t, g, me.ID, station, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{rock},
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact is sacrificed as the cost")
	}
	if got := opp.Library.Size(); got != library-3 {
		t.Errorf("opponent's library %d → %d, want -3", library, got)
	}
	if !b20Tapped(t, g, station) {
		t.Fatal("the Station has a tap cost")
	}
	// An artifact entering — anyone's — offers the untap.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, TreasureToken(), 1) })
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if b20Tapped(t, g, station) {
		t.Error("the Station untaps when an artifact enters")
	}
	// A nonartifact entering is silent.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("3/3 green Beast"), 1) })
	if b18TriggerPromptCount(g, me.ID) != 0 {
		t.Error("a Beast is not an artifact")
	}
	// The Station can eat itself.
	b16Activate(t, g, me.ID, station, 0, game.ActivateAbilityParams{
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

func TestB20TradingPostThreeAbilitiesAndTheDeclaredDiscardGap(t *testing.T) {
	spec, ok := Lookup(b20TradingPostOracle)
	if !ok {
		t.Fatal("Trading Post not registered")
	}
	if len(spec.Activated) != 3 || spec.Completeness != CompletenessCaveats {
		t.Fatalf("three abilities ship and the discard gap is declared, got %d / %v", len(spec.Activated), spec.Completeness)
	}
	g := newCatalogGame(t)
	me := g.Seats[0]
	post := b12Push(g, me.ID, "Trading Post", "Artifact", b20TradingPostOracle, 0, 0)
	advanceToMain(t, g)

	// {1}, {T}, Pay 1 life: a Goat.
	b06AddMana(me, "C")
	life := me.Life
	b16Activate(t, g, me.ID, post, 0, game.ActivateAbilityParams{})
	if me.Life != life-1 {
		t.Errorf("life %d → %d, want -1", life, me.Life)
	}
	goat := findBattlefieldByName(g, "Goat")
	if goat == uuid.Nil {
		t.Fatal("no Goat")
	}
	if c, _ := battlefieldCard(g, goat); c.Power != 0 || c.Toughness != 1 || !c.HasColor("W") {
		t.Error("a 0/1 white Goat")
	}

	// {1}, {T}, Sacrifice a creature: the Goat pays for an artifact
	// card back.
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(post) })
	b06AddMana(me, "C")
	relic := b17GraveyardCard(me, "Sol Ring", "Artifact", "{1}")
	b16Activate(t, g, me.ID, post, 1, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{goat},
		Targets:      b16TargetCard(relic),
	})
	if g.Battlefield.Contains(goat) {
		t.Error("the Goat is sacrificed as the cost")
	}
	if !me.Hand.Contains(relic) {
		t.Error("the artifact card comes back to hand")
	}

	// {1}, {T}, Sacrifice an artifact: the Post eats itself for a card.
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(post) })
	b06AddMana(me, "C")
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, post, 2, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{post}})
	if g.Battlefield.Contains(post) {
		t.Error("'sacrifice an artifact' includes the Post itself")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("drew %d, want 1", got-hand)
	}
}

func TestB20HornOfGreedDrawsForTheLandPlayerOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Horn of Greed", "Artifact", b20HornOfGreedOracle, 0, 0)

	// My land drop: the Horn draws me a card (the land left the
	// hand, so net zero before the draw).
	hands := b20Hands(g)
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hands[0]+1 {
		t.Errorf("my hand %d → %d, want +1", hands[0], got)
	}
	if got := opp.Hand.Size(); got != hands[1] {
		t.Error("only the player who played the land draws")
	}

	// A land put onto the battlefield by an effect is not played.
	buried := b17GraveyardCard(me, "Island", "Basic Land — Island", "")
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.ReturnFromGraveyardUnderControlForEffect(buried, game.ZoneBattlefield, me.ID) })
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Error("a land returned from the graveyard by an effect is not a land play")
	}

	// An opponent's land drop draws the OPPONENT. (Walk to their main
	// phase first — their draw step is not the Horn's doing.)
	aangAdvanceToMain(t, g, 1)
	hands = b20Hands(g)
	b13PlayAs(t, g, 1, "Mountain", "Basic Land — Mountain", "")
	passPriorityAroundTable(t, g)
	if got := opp.Hand.Size(); got != hands[1]+1 {
		t.Errorf("opponent's hand %d → %d, want +1", hands[1], got)
	}
	if got := me.Hand.Size(); got != hands[0] {
		t.Error("the Horn's controller does not draw for an opponent's land")
	}
}

func TestB20HornOfGreedSeesALandPlayedFromExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Horn of Greed", "Artifact", b20HornOfGreedOracle, 0, 0)
	advanceToMain(t, g)
	grant := game.CastPermission{Player: me.ID}
	swamp := b20HandCard(me, "Swamp", "Basic Land — Swamp")
	g.WithWriteLock(func() { _ = g.ExileCardWithPermissionForEffect(swamp, grant) })
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, swamp, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("playing the exiled land: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d: a land played from exile is a land play", hand, got)
	}

	// #1326: a land RETURNED from exile by an effect is still not a
	// play, and the engine's own Event.Played marker (not a land-drop
	// tally guess) tells it apart from a later genuine play from
	// exile even in the same turn.
	plains := b20HandCard(me, "Plains", "Basic Land — Plains")
	g.WithWriteLock(func() {
		_ = g.ExileCardForEffect(plains)
		_, _ = g.ReturnFromExileToBattlefieldForEffect(plains, me.ID, false)
	})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d: a land returned from exile by an effect is not a land play", hand, got)
	}
	island := b20HandCard(me, "Island", "Basic Land — Island")
	g.WithWriteLock(func() { _ = g.ExileCardWithPermissionForEffect(island, grant) })
	hand = me.Hand.Size()
	// #500: two land PLAYS in one turn, which the engine now refuses
	// by default. The subject here is the Horn's read of the play, not
	// CR 305.2, so the seat is given the second land play outright.
	me.LandDropsPerTurn++
	if err := g.CastSpell(me.ID, island, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("playing the second exiled land: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d: a second genuine land play from exile in the same turn still draws", hand, got)
	}
	if spec, _ := Lookup(b20HornOfGreedOracle); spec.Completeness != CompletenessFull {
		t.Error("the gap is closed (#1326): no caveat")
	}
}

func TestB20MasterOfEtheriumCountsArtifactsAndPumpsOtherArtifactCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	master := b12Push(g, me.ID, "Master of Etherium", "Artifact Creature — Vedalken Wizard", b20MasterOfEtheriumOracle, 0, 0)
	if effectivePower(t, g, master) != 1 || effectiveToughness(t, g, master) != 1 {
		t.Errorf("alone it counts itself: %d/%d, want 1/1", effectivePower(t, g, master), effectiveToughness(t, g, master))
	}
	myr := b12Creature(g, me.ID, "Myr", "Artifact Creature — Myr", 1, 1)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := b12Creature(g, opp.ID, "Their Myr", "Artifact Creature — Myr", 1, 1)
	if got := effectivePower(t, g, master); got != 3 {
		t.Errorf("three artifacts: power %d, want 3", got)
	}
	if got := effectivePower(t, g, myr); got != 2 {
		t.Errorf("another artifact creature gets +1/+1: %d, want 2", got)
	}
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("a nonartifact creature is untouched: %d, want 2", got)
	}
	if got := effectivePower(t, g, theirs); got != 1 {
		t.Errorf("an opponent's artifact creature is untouched: %d, want 1", got)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rock) })
	if got := effectivePower(t, g, master); got != 2 {
		t.Errorf("the count is live: %d, want 2", got)
	}
}

// --- the tutors ----------------------------------------------------

func TestB20FierceEmpathAndTributeMageOfferOnlyTheirFilter(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, want string
	}{
		{"Fierce Empath", "Creature — Elf", b20FierceEmpathOracle, "Big Dragon"},
		{"Tribute Mage", "Creature — Human Wizard", b20TributeMageOracle, "Arcane Signet"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			seedSearchLibrary(me,
				game.Card{Name: "Big Dragon", TypeLine: "Creature — Dragon", ManaCost: "{4}{R}{R}"},
				game.Card{Name: "Small Elf", TypeLine: "Creature — Elf", ManaCost: "{G}"},
				game.Card{Name: "Arcane Signet", TypeLine: "Artifact", ManaCost: "{2}"},
				game.Card{Name: "Sol Ring", TypeLine: "Artifact", ManaCost: "{1}"},
				game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
			)
			castCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
			passPriorityAroundTable(t, g)
			c := searchChoiceFor(g, me.ID)
			if c == nil {
				t.Fatal("'you may search' opens the search prompt")
			}
			if len(c.SearchCards) != 1 || searchOptionNamed(g, c, tc.want) == uuid.Nil {
				t.Fatalf("the prompt offers exactly %q, got %d candidates", tc.want, len(c.SearchCards))
			}
			answerSearchNamed(t, g, me.ID, tc.want)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	wall := castCatalogSpell(t, g, "Wall of Blossoms", "Creature — Plant Wall", b20WallOfBlossomsOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d: added and cast (net 0), drew one (+1)", hand, got)
	}
	if !hasEffectiveKeyword(t, g, wall, "defender") {
		t.Error("defender is printed")
	}
}

func TestB20CircuitMenderGainsOnEntryAndDrawsOnAnyExit(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	mender := castCatalogSpell(t, g, "Circuit Mender", "Artifact Creature — Insect", b20CircuitMenderOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2", life, me.Life)
	}
	hand := me.Hand.Size()
	// Exile is not dying, and the Mender draws anyway.
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(mender) })
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d: leaving the battlefield draws", hand, got)
	}
}

func TestB20VerdantSunsAvatarGainsToughnessForItselfAndYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := me.Life
	// "This creature ... enters": the Avatar's own entry gains its
	// own toughness.
	b20CastCreature(t, g, me, "Verdant Sun's Avatar", "Creature — Dinosaur Avatar", b20VerdantSunsAvatarOracle, 5, 5)
	passPriorityAroundTable(t, g)
	if me.Life != life+5 {
		t.Fatalf("its own entry: life %d → %d, want +5", life, me.Life)
	}
	// "Another creature you control": a 3/3 token.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("3/3 green Beast"), 1) })
	passPriorityAroundTable(t, g)
	if me.Life != life+8 {
		t.Errorf("a 3/3 entering: life %d → %d, want +8", life, me.Life)
	}
	// Toughness is read as the trigger resolves: a creature that grew
	// while the trigger waited gains the grown amount.
	bear := b20CastCreature(t, g, me, "Bear", "Creature — Bear", "", 2, 2)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, "+1/+1", 2) })
	passPriorityAroundTable(t, g)
	if me.Life != life+12 {
		t.Errorf("a 2/2 with two counters by resolution: life %d → %d, want +12", life, me.Life)
	}
	// An opponent's creature and a noncreature are silent.
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(opp.ID, TokenCard("3/3 green Beast"), 1)
		_ = g.CreateTokenForEffect(me.ID, TreasureToken(), 1)
	})
	passPriorityAroundTable(t, g)
	if me.Life != life+12 {
		t.Error("an opponent's creature and a Treasure gain nothing")
	}
}

func TestB20KwainEveryoneDrawsAndGainsExceptAnEmptyLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	kwain := b12Push(g, me.ID, "Kwain, Itinerant Meddler", "Legendary Creature — Rabbit Wizard", b20KwainOracle, 1, 3)
	empty := g.Seats[3]
	empty.Library.Cards = nil
	advanceToMain(t, g)
	hands, lives := b20Hands(g), b20Lives(g)
	b16Activate(t, g, me.ID, kwain, 0, game.ActivateAbilityParams{})
	for i, p := range g.Seats {
		wantHand, wantLife := hands[i]+1, lives[i]+1
		if p.ID == empty.ID {
			wantHand, wantLife = hands[i], lives[i]
		}
		if p.Hand.Size() != wantHand || p.Life != wantLife {
			t.Errorf("seat %d: hand %d life %d, want hand %d life %d", i, p.Hand.Size(), p.Life, wantHand, wantLife)
		}
	}
	if empty.AttemptedEmptyDraw {
		t.Error("a player with no library is treated as declining, not as drawing from nothing")
	}
	if spec, _ := Lookup(b20KwainOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the missing 'may' must be declared")
	}
}

// --- the spells ----------------------------------------------------

func TestB20GrappleWithThePastMillsThreeThenReturnsThePick(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b17GraveyardCard(me, "Bear", "Creature — Bear", "{1}{G}")
	library, yard := me.Library.Size(), me.Graveyard.Size()
	castCatalogSpell(t, g, "Grapple with the Past", "Instant", b20GrappleWithThePastOracle, b16TargetCard(bear))
	passPriorityAroundTable(t, g)
	if got := me.Library.Size(); got != library-3 {
		t.Errorf("library %d → %d, want -3", library, got)
	}
	if !me.Hand.Contains(bear) {
		t.Error("the picked creature card comes back to hand")
	}
	// The spell itself and the three milled cards are in the yard;
	// the Bear left it.
	if got := me.Graveyard.Size(); got != yard+3 {
		t.Errorf("graveyard %d → %d, want +3 (three milled, one returned, one spell)", yard, got)
	}
	// With nothing picked it simply mills.
	castCatalogSpell(t, g, "Grapple with the Past", "Instant", b20GrappleWithThePastOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Library.Size(); got != library-6 {
		t.Errorf("library %d → %d, want -6", library, got)
	}
	// An artifact card is not a legal pick.
	rock := b17GraveyardCard(me, "Sol Ring", "Artifact", "{1}")
	if err := b09TryCast(t, g, "Grapple with the Past", "Instant", b20GrappleWithThePastOracle, b16TargetCard(rock)); err == nil {
		t.Error("only a creature or land card may be returned")
	}
	if spec, _ := Lookup(b20GrappleWithThePastOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the pick-before-mill gap must be declared")
	}
}

func TestB20BrokenBondDestroysAnArtifactOrEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	bear := b16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2, "G")
	castCatalogSpell(t, g, "Broken Bond", "Sorcery", b20BrokenBondOracle, b16TargetCard(rock))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Graveyard.Contains(rock) {
		t.Error("the artifact is destroyed")
	}
	if err := b09TryCast(t, g, "Broken Bond", "Sorcery", b20BrokenBondOracle, b16TargetCard(bear)); err == nil {
		t.Error("a creature is not an artifact or enchantment")
	}
	if spec, _ := Lookup(b20BrokenBondOracle); spec.Completeness != CompletenessFull {
		t.Error("the put-a-land clause landed with #654; nothing is deferred any more")
	}
}

func TestB20StarfallInvocationDestroysAllCreaturesAndDeclaresTheGift(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	castCatalogSpell(t, g, "Starfall Invocation", "Sorcery", b20StarfallInvocationOracle, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("every creature is destroyed")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("a noncreature survives")
	}
	// S30 (#470 / #446): the indestructible half of this assertion was
	// an engine gap, not a card gap, and it is closed. The gift is
	// still the one clause this card does not offer.
	spec, _ := Lookup(b20StarfallInvocationOracle)
	if spec.Completeness != CompletenessCaveats || len(spec.Caveats) != 1 {
		t.Error("the un-offered gift is declared, and it is the only caveat left")
	}
}

// --- the edict and the dies trigger --------------------------------

func TestB20MercilessExecutionerMakesEveryoneSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	executioner := castCatalogSpell(t, g, "Merciless Executioner", "Creature — Orc Warrior", b20MercilessExecutionerOracle, nil)
	passPriorityAroundTable(t, g)
	if sacrificeChoiceFor(g, me.ID) == nil || sacrificeChoiceFor(g, opp.ID) == nil {
		t.Fatal("each player with a creature is asked")
	}
	if sacrificeChoiceFor(g, g.Seats[2].ID) != nil {
		t.Error("a player with no creature is skipped, not prompted")
	}
	// Feeding it to itself is the printed line.
	answerSacrifice(t, g, me.ID, executioner)
	answerSacrifice(t, g, opp.ID, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(executioner) || g.Battlefield.Contains(theirs) {
		t.Error("the sacrifices happen")
	}
}

func TestB20TriplicateTitanLeavesThreeDifferentGolems(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	titan := b12Push(g, me.ID, "Triplicate Titan", "Artifact Creature — Golem", b20TriplicateTitanOracle, 9, 9)
	for _, kw := range []string{"flying", "vigilance", "trample"} {
		if !hasEffectiveKeyword(t, g, titan, kw) {
			t.Errorf("%s is printed", kw)
		}
	}
	advanceToMain(t, g)
	b18Kill(t, g, titan)
	golems := battlefieldIDsNamed(g, "Golem")
	if len(golems) != 3 {
		t.Fatalf("%d Golems, want 3", len(golems))
	}
	seen := map[string]int{}
	for _, id := range golems {
		c, _ := battlefieldCard(g, id)
		if c.Power != 3 || c.Toughness != 3 || !c.IsArtifact() || !c.IsCreature() {
			t.Errorf("a 3/3 artifact creature, got %s %d/%d", c.TypeLine, c.Power, c.Toughness)
		}
		for _, kw := range []string{"flying", "vigilance", "trample"} {
			if hasEffectiveKeyword(t, g, id, kw) {
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
	titan2 := b12Push(g, me.ID, "Triplicate Titan", "Artifact Creature — Golem", b20TriplicateTitanOracle, 9, 9)
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(titan2) })
	passPriorityAroundTable(t, g)
	if n := len(battlefieldIDsNamed(g, "Golem")); n != 3 {
		t.Errorf("%d Golems after an exile, want still 3", n)
	}
}

// --- the enchantments ----------------------------------------------

func TestB20AssembleTheLegionMustersOneMoreSoldierEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	legion := b12Push(g, me.ID, "Assemble the Legion", "Enchantment", b20AssembleTheLegionOracle, 0, 0)
	advanceToUpkeepOf(t, g, 1)
	if b16CountNamed(g, "Soldier") != 0 {
		t.Fatal("an opponent's upkeep is not yours")
	}
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if n := b18Counters(t, g, legion, "muster"); n != 1 {
		t.Errorf("%d muster counters, want 1", n)
	}
	if n := b16CountNamed(g, "Soldier"); n != 1 {
		t.Fatalf("%d Soldiers after the first upkeep, want 1", n)
	}
	soldier := findBattlefieldByName(g, "Soldier")
	if !hasEffectiveKeyword(t, g, soldier, "haste") {
		t.Error("the Soldier has haste")
	}
	if c, _ := battlefieldCard(g, soldier); !c.HasColor("R") || !c.HasColor("W") {
		t.Error("the Soldier is red and white")
	}
	advanceToUpkeepOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if n := b18Counters(t, g, legion, "muster"); n != 2 {
		t.Errorf("%d muster counters, want 2", n)
	}
	if n := b16CountNamed(g, "Soldier"); n != 3 {
		t.Errorf("%d Soldiers after the second upkeep, want 3 (1 + 2)", n)
	}
}

func TestB20PrimevalBountyThreeTriggers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Primeval Bounty", "Enchantment", b20PrimevalBountyOracle, 0, 0)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	// A creature spell: a Beast.
	castCatalogSpell(t, g, "Grizzly", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if b16CountNamed(g, "Beast") != 1 {
		t.Error("a creature spell makes a 3/3 Beast")
	}
	// A noncreature spell: three counters on a creature you control,
	// picked when the trigger fires.
	castCatalogSpell(t, g, "Shock", "Instant", "", nil)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if n := b18Counters(t, g, bear, "+1/+1"); n != 3 {
		t.Errorf("%d +1/+1 counters, want 3", n)
	}
	if b16CountNamed(g, "Beast") != 1 {
		t.Error("a noncreature spell makes no Beast")
	}
	// Landfall: 3 life.
	life := me.Life
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3", life, me.Life)
	}
}

func TestB20ManabarbsPingsWhoeverTapsALandForMana(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Manabarbs", "Enchantment", b20ManabarbsOracle, 0, 0)
	mine := seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := seedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island")
	advanceToMain(t, g)
	lives := b20Lives(g)
	if err := g.ActivateManaAbility(me.ID, mine, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != lives[0]-1 {
		t.Errorf("my life %d → %d, want -1: the controller is a player too", lives[0], me.Life)
	}
	if err := g.ActivateManaAbility(opp.ID, theirs, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("opponent ActivateManaAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != lives[1]-1 {
		t.Errorf("opponent's life %d → %d, want -1", lives[1], opp.Life)
	}
	if me.Life != lives[0]-1 {
		t.Error("the damage goes to the player who tapped, not the controller")
	}
	// A nonland mana source is silent.
	ring := b12Push(g, me.ID, "Sol Ring", "Artifact", "6ad8011d-3471-4369-9d68-b264cc027487", 0, 0)
	if err := g.ActivateManaAbility(me.ID, ring, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("Sol Ring: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != lives[0]-1 {
		t.Error("an artifact tapped for mana is not a land")
	}
}

func TestB20JetmirScalesWithThreeSixAndNineCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	jetmir := b12Push(g, me.ID, "Jetmir, Nexus of Revels", "Legendary Creature — Cat Demon", b20JetmirOracle, 5, 4)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	if effectivePower(t, g, bear) != 2 || hasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Fatal("two creatures: no bonus yet")
	}
	b16Creature(g, me.ID, "Bear 3", "Creature — Bear", 2, 2, "G")
	if effectivePower(t, g, bear) != 3 || !hasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Errorf("three creatures: +1/+0 and vigilance, got %d", effectivePower(t, g, bear))
	}
	if effectivePower(t, g, jetmir) != 6 || effectiveToughness(t, g, jetmir) != 4 {
		t.Errorf("Jetmir counts and buffs himself: %d/%d, want 6/4", effectivePower(t, g, jetmir), effectiveToughness(t, g, jetmir))
	}
	if hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("trample waits for six")
	}
	for i := 4; i <= 6; i++ {
		b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	}
	if effectivePower(t, g, bear) != 4 || !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Errorf("six creatures: +2/+0 and trample, got %d", effectivePower(t, g, bear))
	}
	if hasEffectiveKeyword(t, g, bear, "double strike") {
		t.Error("double strike waits for nine")
	}
	for i := 7; i <= 9; i++ {
		b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	}
	if effectivePower(t, g, bear) != 5 || !hasEffectiveKeyword(t, g, bear, "double strike") {
		t.Errorf("nine creatures: +3/+0 and double strike, got %d", effectivePower(t, g, bear))
	}
	// An opponent's creature is not "creatures you control", and does
	// not count.
	theirs := b16Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	if effectivePower(t, g, theirs) != 2 {
		t.Error("an opponent's creature is untouched")
	}
	// Losing one drops the line at once.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(jetmir) })
	if effectivePower(t, g, bear) != 2 {
		t.Errorf("without Jetmir the bonus is gone: %d", effectivePower(t, g, bear))
	}
}

// --- Cemetery Reaper -----------------------------------------------

func TestB20CemeteryReaperLordsZombiesAndMakesThemFromGraveyards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	reaper := b12Push(g, me.ID, "Cemetery Reaper", "Creature — Zombie", b20CemeteryReaperOracle, 2, 2)
	zombie := b16Creature(g, me.ID, "Shambler", "Creature — Zombie", 2, 2, "B")
	theirs := b16Creature(g, opp.ID, "Their Zombie", "Creature — Zombie", 2, 2, "B")
	if effectivePower(t, g, zombie) != 3 {
		t.Errorf("other Zombies you control get +1/+1: %d", effectivePower(t, g, zombie))
	}
	if effectivePower(t, g, reaper) != 2 || effectivePower(t, g, theirs) != 2 {
		t.Error("the Reaper itself and an opponent's Zombie are untouched")
	}
	advanceToMain(t, g)
	dead := b17GraveyardCard(opp, "Their Bear", "Creature — Bear", "{1}{G}")
	b06AddMana(me, "B", "C", "C")
	b16Activate(t, g, me.ID, reaper, 0, game.ActivateAbilityParams{Targets: b16TargetCard(dead)})
	if !inExile(g, dead) {
		t.Error("the creature card is exiled from the opponent's graveyard")
	}
	if n := b16CountNamed(g, "Zombie"); n != 1 {
		t.Fatalf("%d Zombie tokens, want 1", n)
	}
	token := findBattlefieldByName(g, "Zombie")
	if effectivePower(t, g, token) != 3 {
		t.Error("the new Zombie is a 2/2 that the Reaper lords to 3/3")
	}
	if !b20Tapped(t, g, reaper) {
		t.Error("the ability has a tap cost")
	}
}

// --- Prosper, Tome-Bound -------------------------------------------

func TestB20ProsperExilesAtYourEndStepAndPaysTreasureForPlaysFromExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	prosper := b12Push(g, me.ID, "Prosper, Tome-Bound", "Legendary Creature — Tiefling Warlock", b20ProsperOracle, 1, 4)
	if !hasEffectiveKeyword(t, g, prosper, "deathtouch") {
		t.Error("deathtouch is printed")
	}
	// The first turn's end step exiles whatever filler is on top;
	// seed the real library after it so the walk below is about the
	// second round. The Filler is what seat 0's next draw step takes,
	// leaving the Forest on top for the end step.
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if n := len(g.Exile.Cards); n != 1 {
		t.Fatalf("%d cards in exile after your end step, want 1", n)
	}
	ids := seedSearchLibrary(me,
		game.Card{Name: "Filler", TypeLine: "Sorcery"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Deep Card", TypeLine: "Sorcery"},
	)
	forest := ids[1]
	// An opponent's end step is not yours.
	advanceToEndStepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(forest) {
		t.Fatal("Mystic Arcanum fires on the controller's own end step only")
	}
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(forest) {
		t.Fatal("the top card is exiled at your end step")
	}
	perm := exiledPermission(g, forest)
	if perm.Player != me.ID || perm.CastOnly || !permissionLive(g, perm, me.ID) {
		t.Fatalf("grant %+v: the controller may PLAY it", perm)
	}
	// Through the opponents' turns the grant survives, and it reaches
	// the END of the controller's next turn (#945: one duration, no
	// upkeep re-stamp).
	advanceToMainOf(t, g, 3)
	if !permissionLive(g, exiledPermission(g, forest), me.ID) {
		t.Fatal("the grant survives every opponent's cleanup")
	}
	// Playing the land from exile is "playing a card from exile":
	// Pact Boon makes a Treasure.
	advanceToMainOf(t, g, 0)
	if p := exiledPermission(g, forest); !permissionLive(g, p, me.ID) {
		t.Fatalf("on your next turn the grant is %+v, want live", p)
	}
	if err := g.CastSpell(me.ID, forest, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("playing the exiled land: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(forest) {
		t.Fatal("the land is played")
	}
	if n := b16CountNamed(g, "Treasure"); n != 1 {
		t.Errorf("%d Treasures, want 1", n)
	}
	// A spell cast from exile under any grant does too.
	shock := b20HandCard(me, "Shock", "Instant")
	g.WithWriteLock(func() {
		_ = g.ExileCardWithPermissionForEffect(shock, game.CastPermission{Player: me.ID})
	})
	if err := g.CastSpell(me.ID, shock, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("casting from exile: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Treasure"); n != 2 {
		t.Errorf("%d Treasures, want 2", n)
	}
	// A spell cast from hand is not.
	castCatalogSpell(t, g, "Bolt", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Treasure"); n != 2 {
		t.Error("a cast from hand makes no Treasure")
	}
	// A card exiled at the end step with an empty library: no grant,
	// no crash.
	me.Library.Cards = nil
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
}
