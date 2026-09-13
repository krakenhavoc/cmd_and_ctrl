package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch19_test.go — card-level coverage for the card-coverage
// roadmap's batch 19 (#312, `edhrec_rank` 2032–2132): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, land play, attack or step change.
// Helpers from the earlier batch test files are reused by name; new
// ones are b19-prefixed.

const (
	b19TrinketMageOracle            = "ffc95093-24a2-4616-a44a-24788e8df9c8"
	b19BirthingPodOracle            = "f8b9dd54-0837-47f4-ad14-7a0322d46d5f"
	b19CancelOracle                 = "7d00fb28-ea6c-49a9-b4af-ffb38860a9a7"
	b19SelesnyaGuildgateOracle      = "75b235d3-595a-4859-be45-9559d8445db5"
	b19ElvishReclaimerOracle        = "94f01534-7bb3-4b1e-8f77-e84156408686"
	b19DeathreapRitualOracle        = "b26e3596-5b28-4eb6-b3e2-03f63d8c6d49"
	b19GrimBackwoodsOracle          = "5effaa94-7f87-4485-8959-473d584c5034"
	b19WindingConstrictorOracle     = "c9404d7d-a026-4082-9fcb-1ab571a136b5"
	b19GenerousPlundererOracle      = "91c835d1-22ca-4c90-9ba6-c8e01bbc0347"
	b19AzoriusGuildgateOracle       = "ad1712d8-809f-410c-8b91-ffe6fb8a69a1"
	b19WeaverOfHarmonyOracle        = "494b31b2-27ef-4ca1-ac72-c2fcfc8a23a1"
	b19RuthlessTechnomancerOracle   = "4e58ad76-37c7-4531-b207-6890b39a2679"
	b19DrownInDreamsOracle          = "d3cec4b5-bc93-44a2-a29d-3478f0a5dac6"
	b19IllustriousWanderglyphOracle = "72adf091-4491-4afe-b893-b8d20ad38a86"
	b19LiesaOracle                  = "efcaadbe-24e3-4dfc-b08c-a910f003d427"
	b19ErodeOracle                  = "2e467fab-e808-44d3-99bf-e3621baeb7cb"
	b19PerplexingTestOracle         = "8682266c-2b0f-496e-b1c6-1fb338feac24"
	b19ZendikarsRoilOracle          = "a842cc2b-52eb-4dc5-86c6-6575c2ed913d"
	b19CracklingDoomOracle          = "ee81f37b-2a81-46d7-8d2f-9091123846c4"
	b19RecklessImpulseOracle        = "584cf0fd-112a-4ca6-9c0e-1f3228a7d325"
	b19DiregrafColossusOracle       = "fd62ad01-601f-4250-bc2a-8ef3982e45c4"
	b19WrennsResolveOracle          = "a7d8c99c-cd44-4bad-82e6-7c7cff9bd89f"
)

// b19CastXModal casts an X spell with a mode choice from the active
// seat's hand — castXSpell and castModal in one, for Drown in Dreams.
func b19CastXModal(t *testing.T, g *game.Game, name, typeLine, oracle, manaCost string, x int, modes []int, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
		Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Modes: modes, Targets: targets, XValue: x}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// b19HasTriggerPromptFor reports whether an optional-trigger prompt
// is waiting on `chooser`.
func b19HasTriggerPromptFor(g *game.Game, chooser uuid.UUID) bool {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == chooser {
			return true
		}
	}
	return false
}

// b19TriggerQueuedFrom reports whether a triggered ability from
// `source` is waiting — harvested into PendingTriggers, or already on
// the stack.
func b19TriggerQueuedFrom(g *game.Game, source uuid.UUID) bool {
	for _, item := range g.PendingTriggers {
		if item != nil && item.SourceCardID == source {
			return true
		}
	}
	return triggerOnStack(g, source) != nil
}

// b19SacrificeOptions reads the option list of a player's open
// sacrifice prompt.
func b19SacrificeOptions(g *game.Game, chooser uuid.UUID) []uuid.UUID {
	if c := sacrificeChoiceFor(g, chooser); c != nil {
		return c.SacrificeOptions
	}
	return nil
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. The two
// Guildgates are rows in the Guildgate table, so a transposed row is
// invisible until someone plays that exact card.
func TestBatch19CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b19TrinketMageOracle:            "Trinket Mage",
		b19BirthingPodOracle:            "Birthing Pod",
		b19CancelOracle:                 "Cancel",
		b19SelesnyaGuildgateOracle:      "Selesnya Guildgate",
		b19ElvishReclaimerOracle:        "Elvish Reclaimer",
		b19DeathreapRitualOracle:        "Deathreap Ritual",
		b19GrimBackwoodsOracle:          "Grim Backwoods",
		b19WindingConstrictorOracle:     "Winding Constrictor",
		b19GenerousPlundererOracle:      "Generous Plunderer",
		b19AzoriusGuildgateOracle:       "Azorius Guildgate",
		b19WeaverOfHarmonyOracle:        "Weaver of Harmony",
		b19RuthlessTechnomancerOracle:   "Ruthless Technomancer",
		b19DrownInDreamsOracle:          "Drown in Dreams",
		b19IllustriousWanderglyphOracle: "Illustrious Wanderglyph",
		b19LiesaOracle:                  "Liesa, Forgotten Archangel",
		b19ErodeOracle:                  "Erode",
		b19PerplexingTestOracle:         "Perplexing Test",
		b19ZendikarsRoilOracle:          "Zendikar's Roil",
		b19CracklingDoomOracle:          "Crackling Doom",
		b19RecklessImpulseOracle:        "Reckless Impulse",
		b19DiregrafColossusOracle:       "Diregraf Colossus",
		b19WrennsResolveOracle:          "Wrenn's Resolve",
	}
	if len(want) != 22 {
		t.Fatalf("the batch registers 22 cards, the table lists %d", len(want))
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
}

// --- the lands -----------------------------------------------------

func TestB19GuildgatesEnterTappedAndTapForTheirColours(t *testing.T) {
	for _, tc := range []struct{ name, oracle, produced string }{
		{"Selesnya Guildgate", b19SelesnyaGuildgateOracle, "{G|W}"},
		{"Azorius Guildgate", b19AzoriusGuildgateOracle, "{W|U}"},
	} {
		g := newCatalogGame(t)
		gate := playLandFromHand(t, g, tc.name, tc.oracle)
		top100AssertEnteredTapped(t, g, gate, tc.name)
		if tapEventsFor(g, gate) != 0 {
			t.Errorf("%s: enters-tapped is a replacement, not a tap", tc.name)
		}
		spec, _ := Lookup(tc.oracle)
		if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != tc.produced {
			t.Errorf("%s: mana abilities %+v, want one producing %s", tc.name, spec.ManaAbilities, tc.produced)
		}
	}
}

func TestB19GrimBackwoodsTapsForColorlessAndEatsACreatureToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	woods := b12Permanent(g, me.ID, "Grim Backwoods", "Land")
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == woods {
			g.Battlefield.Cards[i].OracleID = b19GrimBackwoodsOracle
		}
	}
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	advanceToMain(t, g)
	spec, _ := Lookup(b19GrimBackwoodsOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{C}" {
		t.Errorf("mana abilities %+v, want one producing {C}", spec.ManaAbilities)
	}
	hand := me.Hand.Size()
	b06AddMana(me, "C", "C", "B", "G")
	if err := g.ActivateCatalogAbility(me.ID, woods, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("the creature is a cost, not optional")
	}
	b16Activate(t, g, me.ID, woods, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}})
	if g.Battlefield.Contains(bear) {
		t.Error("the creature is sacrificed as a cost")
	}
	if !b16Tapped(t, g, woods) {
		t.Error("the land taps as its cost")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("drew %d, want 1", me.Hand.Size()-hand)
	}
}

// --- the spells ----------------------------------------------------

func TestB19CancelCountersASpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "", b16TargetPlayer(me.ID))
	castCatalogSpell(t, g, "Cancel", "Instant", b19CancelOracle, b16TargetCard(bolt))
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) || me.Life != 40 {
		t.Error("the Bolt is countered")
	}
}

func TestB19ErodeDestroysAndOffersItsControllerABasic(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	seedSearchLibrary(opp,
		searchTestLand("Plains", "Basic Land — Plains"),
		searchTestLand("Island", "Basic Land — Island"),
		searchTestLand("Command Tower", "Land"),
	)
	castCatalogSpell(t, g, "Erode", "Instant", b19ErodeOracle, b16TargetCard(theirs))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Graveyard.Contains(theirs) {
		t.Fatal("the target is destroyed")
	}
	c := searchChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("ITS CONTROLLER may search — the opponent gets the prompt")
	}
	if searchChoiceFor(g, me.ID) != nil {
		t.Error("the caster searches nothing")
	}
	if searchOptionNamed(g, c, "Command Tower") != uuid.Nil {
		t.Error("only basic land cards are offered")
	}
	answerSearchNamed(t, g, opp.ID, "Island")
	island := findBattlefieldByName(g, "Island")
	if island == uuid.Nil || controllerOf(t, g, island) != opp.ID || !b16Tapped(t, g, island) {
		t.Error("the basic enters tapped under the opponent's control")
	}

	// A planeswalker is a legal target; an artifact is not.
	walker := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Jace", TypeLine: "Legendary Planeswalker — Jace", Owner: opp.ID, Controller: opp.ID,
		Counters: map[string]int{game.CounterLoyalty: 3},
	})
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	seedSearchLibrary(opp, searchTestLand("Swamp", "Basic Land — Swamp"))
	id := handCardFull(me, "Erode", "Instant", "", b19ErodeOracle, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: b16TargetCard(rock)}); err == nil {
		t.Fatal("an artifact is not a creature or planeswalker")
	}
	castCatalogSpell(t, g, "Erode", "Instant", b19ErodeOracle, b16TargetCard(walker))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(walker) {
		t.Error("the planeswalker is destroyed")
	}
	// "May" search: even one candidate is offered rather than taken.
	if searchChoiceFor(g, opp.ID) == nil {
		t.Fatal("the search is optional, so it always asks")
	}
	answerSearchNamed(t, g, opp.ID, "Swamp")
	if findBattlefieldByName(g, "Swamp") == uuid.Nil {
		t.Error("the basic lands")
	}
}

func TestB19PerplexingTestBouncesTokensOrNontokens(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, RedGoblinToken(), 2) })
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1) })
	castModal(t, g, "Perplexing Test", "Instant", b19PerplexingTestOracle, []int{0}, nil)
	passPriorityAroundTable(t, g)
	if b16CountNamed(g, "Goblin") != 0 {
		t.Error("mode one returns every creature token")
	}
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(theirs) {
		t.Error("mode one leaves nontoken creatures alone")
	}
	hand, theirHand := me.Hand.Size(), opp.Hand.Size()
	castModal(t, g, "Perplexing Test", "Instant", b19PerplexingTestOracle, []int{1}, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("mode two returns every nontoken creature")
	}
	if me.Hand.Size() != hand+1 || opp.Hand.Size() != theirHand+1 {
		t.Error("to their owners' hands")
	}
}

func TestB19CracklingDoomPingsAndEdictsTheBiggest(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other, third := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	mine := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	small := b16Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 1, 1, "R")
	bigA := b16Creature(g, opp.ID, "Their Wurm", "Creature — Wurm", 4, 4, "G")
	bigB := b16Creature(g, opp.ID, "Their Pumped Bear", "Creature — Bear", 2, 2, "G")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bigB, "+1/+1", 2) })
	only := b16Creature(g, other.ID, "Lone Elf", "Creature — Elf", 1, 1, "G")
	before := b17Life(g)
	castCatalogSpell(t, g, "Crackling Doom", "Instant", b19CracklingDoomOracle, nil)
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		want := before[i]
		if i != 0 {
			want -= 2
		}
		if p.Life != want {
			t.Errorf("seat %d life %d, want %d", i, p.Life, want)
		}
	}
	options := b19SacrificeOptions(g, opp.ID)
	if len(options) != 2 || !hasID(options, bigA) || !hasID(options, bigB) || hasID(options, small) {
		t.Errorf("the opponent may sacrifice only a creature tied for greatest power (counters count): %v", options)
	}
	if b19SacrificeOptions(g, other.ID) == nil || sacrificeChoiceFor(g, third.ID) != nil || sacrificeChoiceFor(g, me.ID) != nil {
		t.Error("each opponent with a creature sacrifices; the caster and a creatureless opponent do not")
	}
	answerSacrifice(t, g, opp.ID, bigB)
	answerSacrifice(t, g, other.ID, only)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bigB) || g.Battlefield.Contains(only) || !g.Battlefield.Contains(bigA) || !g.Battlefield.Contains(mine) {
		t.Error("the chosen creatures are sacrificed and nothing else")
	}
}

func TestB19DrownInDreamsDrawsXOrMillsTwiceX(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hand := me.Hand.Size()
	b19CastXModal(t, g, "Drown in Dreams", "Instant", b19DrownInDreamsOracle, "{X}{2}{U}", 3, []int{0}, b16TargetPlayer(me.ID))
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 {
		t.Errorf("X=3 draws 3, drew %d", me.Hand.Size()-hand)
	}
	library := opp.Library.Size()
	b19CastXModal(t, g, "Drown in Dreams", "Instant", b19DrownInDreamsOracle, "{X}{2}{U}", 3, []int{1}, b16TargetPlayer(opp.ID))
	passPriorityAroundTable(t, g)
	if opp.Library.Size() != library-6 || opp.Graveyard.Size() < 6 {
		t.Errorf("X=3 mills 6, milled %d", library-opp.Library.Size())
	}
	// The declared gap: never both.
	id := handCardFull(me, "Drown in Dreams", "Instant", "{X}{2}{U}", b19DrownInDreamsOracle, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0, 1}, Targets: b16TargetPlayer(me.ID), XValue: 1}); err == nil {
		t.Error("choose one — both modes is refused (the commander rider is the declared gap)")
	}
	if spec, _ := Lookup(b19DrownInDreamsOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the choose-both gap must be declared")
	}
}

func TestB19RecklessImpulseGrantsPlayUntilTheEndOfYourNextTurn(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		seat         int
	}{
		{"Reckless Impulse", b19RecklessImpulseOracle, 0},
		{"Wrenn's Resolve", b19WrennsResolveOracle, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[tc.seat]
			advanceToMainOf(t, g, tc.seat)
			ids := seedSearchLibrary(me,
				game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
				game.Card{Name: "Shock", TypeLine: "Instant"},
				game.Card{Name: "Deep Card", TypeLine: "Sorcery"},
			)
			forest, shock, deep := ids[0], ids[1], ids[2]
			castCatalogSpell(t, g, tc.name, "Sorcery", tc.oracle, nil)
			passPriorityAroundTable(t, g)
			if !g.Exile.Contains(forest) || !g.Exile.Contains(shock) || !me.Library.Contains(deep) {
				t.Fatal("the top two cards are exiled")
			}
			round := g.Turn.Number
			for _, id := range []uuid.UUID{forest, shock} {
				perm := exiledPermission(g, id)
				if perm.Player != me.ID || perm.CastOnly || !perm.Active(me.ID, round) {
					t.Errorf("%+v: the caster may PLAY it now", perm)
				}
			}
			// The land can be played this turn.
			if err := g.CastSpell(me.ID, forest, game.CastSpellParams{FromZone: "exile"}); err != nil {
				t.Fatalf("playing the exiled land this turn: %v", err)
			}
			// Through every opponent's turn the grant is still live.
			advanceToMainOf(t, g, (tc.seat+1)%4)
			if !exiledPermission(g, shock).Active(me.ID, g.Turn.Number) {
				t.Fatal("the grant survives the caster's own cleanup")
			}
			advanceToMainOf(t, g, (tc.seat+3)%4)
			if !exiledPermission(g, shock).Active(me.ID, g.Turn.Number) {
				t.Fatal("the grant survives seat 0's cleanup of the next round too")
			}
			// Your next upkeep: the delayed trigger pins the grant to
			// this turn.
			advanceToUpkeepOf(t, g, tc.seat)
			if stackFullyEmpty(g) {
				t.Fatal("the delayed trigger fires at the beginning of your next upkeep")
			}
			if exiledPermission(g, shock).UntilTurn == g.Turn.Number {
				t.Fatal("the trigger uses the stack: nothing is re-stamped before it resolves")
			}
			passPriorityAroundTable(t, g)
			perm := exiledPermission(g, shock)
			if perm.UntilTurn != g.Turn.Number || !perm.Active(me.ID, g.Turn.Number) {
				t.Errorf("after the upkeep trigger the grant is %+v, want live through this turn only", perm)
			}
			if !g.Battlefield.Contains(forest) {
				t.Error("the land played last turn is untouched by the re-stamp")
			}
			// Still castable at your end step; gone on the next player's turn.
			advanceToEndStepOf(t, g, tc.seat)
			if !exiledPermission(g, shock).Active(me.ID, g.Turn.Number) {
				t.Error("the end of your next turn is still your next turn")
			}
			advanceToMainOf(t, g, (tc.seat+1)%4)
			if exiledPermission(g, shock).Granted() {
				t.Error("the grant lapses at your next turn's cleanup")
			}
			if !g.Exile.Contains(shock) {
				t.Error("the card stays exiled")
			}
		})
	}
}

// --- the creatures: ETB and search ---------------------------------

func TestB19TrinketMageTutorsACheapArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		game.Card{Name: "Sol Ring", TypeLine: "Artifact", ManaCost: "{1}"},
		game.Card{Name: "Mox Opal", TypeLine: "Legendary Artifact", ManaCost: ""},
		game.Card{Name: "Mind Stone", TypeLine: "Artifact", ManaCost: "{2}"},
		game.Card{Name: "Myr", TypeLine: "Artifact Creature — Myr", ManaCost: "{1}"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{G}"},
	)
	castCatalogSpell(t, g, "Trinket Mage", "Creature — Human Wizard", b19TrinketMageOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("you may search: the chooser asks")
	}
	for _, name := range []string{"Sol Ring", "Mox Opal", "Myr"} {
		if searchOptionNamed(g, c, name) == uuid.Nil {
			t.Errorf("%s (an artifact with mana value 1 or less) is offered", name)
		}
	}
	if searchOptionNamed(g, c, "Mind Stone") != uuid.Nil || searchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("mana value 2, and a non-artifact, are not")
	}
	answerSearchNamed(t, g, me.ID, "Sol Ring")
	if !b02bHandHasNamed(me, "Sol Ring") {
		t.Error("the pick goes to hand")
	}
	if me.Library.Size() != 4 {
		t.Errorf("library %d, want 4", me.Library.Size())
	}
}

func TestB19RuthlessTechnomancerTradesACreatureForTreasures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	beast := b16Creature(g, me.ID, "Beast", "Creature — Beast", 4, 4, "G")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(beast, "+1/+1", 1) })
	mancer := castAndResolveCreature(t, g, "Ruthless Technomancer", "Creature — Human Wizard", b19RuthlessTechnomancerOracle)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if hasID(p.PickTargetCards, mancer) || !hasID(p.PickTargetCards, beast) {
		t.Error("ANOTHER creature you control: the Technomancer itself is not offered")
	}
	pickCard(t, g, me.ID, beast)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(beast) {
		t.Fatal("the chosen creature is sacrificed")
	}
	if n := b16CountNamed(g, "Treasure"); n != 5 {
		t.Errorf("a 4/4 with a +1/+1 counter: 5 Treasures, got %d", n)
	}

	// Declined: nothing happens.
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	castAndResolveCreature(t, g, "Ruthless Technomancer", "Creature — Human Wizard", b19RuthlessTechnomancerOracle)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) || b16CountNamed(g, "Treasure") != 5 {
		t.Error("declining sacrifices nothing and makes nothing")
	}
	if spec, _ := Lookup(b19RuthlessTechnomancerOracle); spec.Completeness != CompletenessCaveats || len(spec.Activated) != 0 {
		t.Error("the omitted reanimation ability must be declared")
	}
}

func TestB19BirthingPodChainsUpOneManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pod := b12Permanent(g, me.ID, "Birthing Pod", "Artifact")
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == pod {
			g.Battlefield.Cards[i].OracleID = b19BirthingPodOracle
		}
	}
	two := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Two-Drop", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	seedSearchLibrary(me,
		game.Card{Name: "Three-Drop", TypeLine: "Creature — Elf", ManaCost: "{2}{G}"},
		game.Card{Name: "Other Three", TypeLine: "Creature — Wurm", ManaCost: "{1}{G}{G}"},
		game.Card{Name: "Four-Drop", TypeLine: "Creature — Giant", ManaCost: "{3}{G}"},
		game.Card{Name: "Three Rock", TypeLine: "Artifact", ManaCost: "{3}"},
	)
	spec, _ := Lookup(b19BirthingPodOracle)
	if len(spec.Activated) != 1 || !spec.Activated[0].SorcerySpeed {
		t.Fatal("one ability, sorcery speed")
	}
	advanceToMain(t, g)
	b06AddMana(me, "C", "G")
	b16Activate(t, g, me.ID, pod, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{two}})
	if g.Battlefield.Contains(two) || !b16Tapped(t, g, pod) {
		t.Fatal("the creature is sacrificed and the Pod taps")
	}
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("two creatures at mana value 3: the searcher chooses")
	}
	if searchOptionNamed(g, c, "Four-Drop") != uuid.Nil || searchOptionNamed(g, c, "Three Rock") != uuid.Nil {
		t.Error("only creature cards with mana value exactly 3 are offered")
	}
	answerSearchNamed(t, g, me.ID, "Other Three")
	wurm := findBattlefieldByName(g, "Other Three")
	if wurm == uuid.Nil || b16Tapped(t, g, wurm) {
		t.Error("the pick enters the battlefield untapped")
	}
	if spec.Completeness != CompletenessCaveats {
		t.Error("the Phyrexian mana gap must be declared")
	}
}

func TestB19ElvishReclaimerGrowsAndCropRotates(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	reclaimer := b12Push(g, me.ID, "Elvish Reclaimer", "Creature — Elf Warrior", b19ElvishReclaimerOracle, 1, 2)
	forest := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	b17GraveyardCard(me, "Dead Forest", "Basic Land — Forest", "")
	b17GraveyardCard(me, "Dead Swamp", "Basic Land — Swamp", "")
	g.WithWriteLock(func() { g.BumpLayerVersionForTest() })
	if effectivePower(t, g, reclaimer) != 1 {
		t.Fatal("two land cards in the graveyard: still 1/2")
	}
	seedSearchLibrary(me,
		searchTestLand("Gaea's Cradle", "Legendary Land"),
		searchTestLand("Plains", "Basic Land — Plains"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	advanceToMain(t, g)
	b06AddMana(me, "C", "C")
	b16Activate(t, g, me.ID, reclaimer, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{forest}})
	if g.Battlefield.Contains(forest) || !b16Tapped(t, g, reclaimer) {
		t.Fatal("the land is sacrificed and the Reclaimer taps")
	}
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("two land cards: the searcher chooses")
	}
	if searchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("only land cards are offered")
	}
	answerSearchNamed(t, g, me.ID, "Gaea's Cradle")
	cradle := findBattlefieldByName(g, "Gaea's Cradle")
	if cradle == uuid.Nil || !b16Tapped(t, g, cradle) {
		t.Error("any land card, onto the battlefield tapped")
	}
	// The sacrificed Forest is the third land card in the graveyard.
	if got := effectivePower(t, g, reclaimer); got != 3 {
		t.Errorf("three land cards in the graveyard: 3/4, got power %d", got)
	}
	if got := effectiveToughness(t, g, reclaimer); got != 4 {
		t.Errorf("toughness %d, want 4", got)
	}
}

func TestB19DiregrafColossusCountsZombiesAndMakesTappedOnes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b17GraveyardCard(me, "Dead Zombie", "Creature — Zombie", "")
	b17GraveyardCard(me, "Dead Changeling", "Creature — Shapeshifter", "")
	b17GraveyardCard(me, "Dead Bear", "Creature — Bear", "")
	b17GraveyardCard(opp, "Their Dead Zombie", "Creature — Zombie", "")
	colossus := castCatalogSpell(t, g, "Diregraf Colossus", "Creature — Zombie Giant", b19DiregrafColossusOracle, nil)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, colossus, "+1/+1"); n != 1 {
		t.Fatalf("one Zombie card in YOUR graveyard: 1 counter, got %d", n)
	}
	castCatalogSpell(t, g, "Gravecrawler", "Creature — Zombie", "", nil)
	if triggerOnStack(g, colossus) == nil {
		t.Fatal("casting a Zombie spell triggers")
	}
	passPriorityAroundTable(t, g)
	zombies := battlefieldIDsNamed(g, "Zombie")
	if len(zombies) != 1 {
		t.Fatalf("one tapped 2/2 Zombie token, got %d", len(zombies))
	}
	if !b16Tapped(t, g, zombies[0]) || effectivePower(t, g, zombies[0]) != 2 {
		t.Error("the token is tapped and 2/2")
	}
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	if triggerOnStack(g, colossus) != nil {
		t.Error("a non-Zombie spell does not trigger")
	}
	passPriorityAroundTable(t, g)
}

// --- the creatures: statics and replacements -----------------------

func TestB19WindingConstrictorAddsOneCounterOfEachKind(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Winding Constrictor", "Creature — Snake", b19WindingConstrictorOracle, 2, 3)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	rock := b12Permanent(g, me.ID, "Coffers", "Artifact")
	shrine := b12Permanent(g, me.ID, "Shrine", "Enchantment")
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(bear, "+1/+1", 1)
		_ = g.AddCounterForEffect(rock, "charge", 2)
		_ = g.AddCounterForEffect(shrine, "verse", 1)
		_ = g.AddCounterForEffect(theirs, "+1/+1", 1)
	})
	if counterCount(g, bear, "+1/+1") != 2 {
		t.Errorf("one +1/+1 counter becomes two, got %d", counterCount(g, bear, "+1/+1"))
	}
	if counterCount(g, rock, "charge") != 3 {
		t.Errorf("two charge counters on an artifact become three, got %d", counterCount(g, rock, "charge"))
	}
	if counterCount(g, shrine, "verse") != 1 || counterCount(g, theirs, "+1/+1") != 1 {
		t.Error("an enchantment, and an opponent's creature, get the printed number")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, "+1/+1", -1) })
	if counterCount(g, bear, "+1/+1") != 1 {
		t.Errorf("a removal is not counters being put on: %d, want 1", counterCount(g, bear, "+1/+1"))
	}
	// "Enters with" counters run the same pipeline: a Diregraf
	// Colossus with two Zombies in the graveyard enters with three.
	b17GraveyardCard(me, "Dead Zombie", "Creature — Zombie", "")
	b17GraveyardCard(me, "Dead Ghoul", "Creature — Zombie", "")
	colossus := castCatalogSpell(t, g, "Diregraf Colossus", "Creature — Zombie Giant", b19DiregrafColossusOracle, nil)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, colossus, "+1/+1"); n != 3 {
		t.Errorf("enters with 2+1 counters, got %d", n)
	}
	if spec, _ := Lookup(b19WindingConstrictorOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the player-counter gap must be declared")
	}
}

func TestB19WeaverOfHarmonyBuffsOtherEnchantmentCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	weaver := b12Push(g, me.ID, "Weaver of Harmony", "Enchantment Creature — Snake Druid", b19WeaverOfHarmonyOracle, 2, 2)
	kami := b16Creature(g, me.ID, "Kami", "Enchantment Creature — Spirit", 1, 1, "W")
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	shrine := b12Permanent(g, me.ID, "Shrine", "Enchantment")
	theirs := b16Creature(g, opp.ID, "Their Kami", "Enchantment Creature — Spirit", 1, 1, "W")
	if effectivePower(t, g, kami) != 2 || effectiveToughness(t, g, kami) != 2 {
		t.Error("another enchantment creature you control gets +1/+1")
	}
	if effectivePower(t, g, weaver) != 2 || effectivePower(t, g, bear) != 2 || effectivePower(t, g, theirs) != 1 {
		t.Error("not itself, not a plain creature, not an opponent's")
	}
	_ = shrine
	if spec, _ := Lookup(b19WeaverOfHarmonyOracle); spec.Completeness != CompletenessCaveats || len(spec.Activated) != 0 {
		t.Error("the omitted copy ability must be declared")
	}
}

func TestB19IllustriousWanderglyphMakesGnomesAndLordsThemAtTen(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	glyph := b12Push(g, me.ID, "Illustrious Wanderglyph", "Artifact Creature — Golem", b19IllustriousWanderglyphOracle, 2, 2)
	myr := b16Creature(g, me.ID, "Myr", "Artifact Creature — Myr", 1, 1)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirMyr := b16Creature(g, opp.ID, "Their Myr", "Artifact Creature — Myr", 1, 1)
	// Four permanents: no blessing.
	if effectivePower(t, g, myr) != 1 {
		t.Fatal("below ten permanents: no bonus")
	}
	// Each upkeep, anyone's, makes a Gnome for the controller.
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Gnome"); n != 1 {
		t.Fatalf("an opponent's upkeep: 1 Gnome, got %d", n)
	}
	gnome := findBattlefieldByName(g, "Gnome")
	if controllerOf(t, g, gnome) != me.ID {
		t.Error("the Gnome is the Wanderglyph's controller's")
	}
	for i := 0; i < 5; i++ {
		b12Permanent(g, me.ID, "Land", "Basic Land — Forest")
	}
	// Glyph, Myr, Bear, Gnome, five lands: nine. One more.
	if effectivePower(t, g, myr) != 1 {
		t.Fatal("nine permanents: still no bonus")
	}
	b12Permanent(g, me.ID, "Last Land", "Basic Land — Forest")
	if effectivePower(t, g, myr) != 3 || effectivePower(t, g, gnome) != 3 {
		t.Error("ten permanents: other artifact creatures you control get +2/+2")
	}
	if effectivePower(t, g, glyph) != 2 || effectivePower(t, g, bear) != 2 || effectivePower(t, g, theirMyr) != 1 {
		t.Error("not itself, not a non-artifact, not an opponent's")
	}
	if spec, _ := Lookup(b19IllustriousWanderglyphOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the blessing gap must be declared")
	}
}

// --- the creatures: triggers ---------------------------------------

func TestB19ZendikarsRoilMakesAnElementalPerLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Zendikar's Roil", "Enchantment", b19ZendikarsRoilOracle, 0, 0)
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	elementals := battlefieldIDsNamed(g, "Elemental")
	if len(elementals) != 1 {
		t.Fatalf("a land you control entering: one Elemental, got %d", len(elementals))
	}
	if c := cardByID(g, elementals[0]); effectivePower(t, g, elementals[0]) != 2 || len(c.Colors) != 1 || c.Colors[0] != "G" || controllerOf(t, g, elementals[0]) != me.ID {
		t.Error("a 2/2 green Elemental under your control")
	}
	advanceToMainOf(t, g, 1)
	playLandFromHand(t, g, "Their Forest", "")
	passPriorityAroundTable(t, g)
	if b16CountNamed(g, "Elemental") != 1 {
		t.Error("an opponent's land is not yours")
	}
	_ = opp
}

func TestB19DeathreapRitualDrawsAtEndStepsAfterADeath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Deathreap Ritual", "Enchantment", b19DeathreapRitualOracle, 0, 0)
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	// No death this turn: the end step passes without a prompt.
	advanceToEndStepOf(t, g, 0)
	if b19HasTriggerPromptFor(g, me.ID) {
		t.Fatal("no creature died: no trigger (intervening if)")
	}
	// A death on an opponent's turn: their end step asks you.
	advanceToMainOf(t, g, 1)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	hand := me.Hand.Size()
	advanceToEndStepOf(t, g, 1)
	if !b19HasTriggerPromptFor(g, me.ID) {
		t.Fatal("a creature died this turn: each end step asks the controller")
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("drew %d, want 1", me.Hand.Size()-hand)
	}
	// The next turn starts clean.
	advanceToEndStepOf(t, g, 2)
	if b19HasTriggerPromptFor(g, me.ID) {
		t.Error("the tally resets with the turn")
	}
}

func TestB19GenerousPlundererGiftsTreasureAndPunishesArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	plunderer := b12Push(g, me.ID, "Generous Plunderer", "Creature — Human Rogue", b19GenerousPlundererOracle, 2, 2)
	if !hasEffectiveKeyword(t, g, plunderer, "menace") {
		t.Error("menace")
	}
	b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	b12Permanent(g, opp.ID, "Their Signet", "Artifact")
	// Your NEXT upkeep (the game opens in this one): may make a
	// Treasure; the chosen opponent gets a tapped one.
	advanceToMain(t, g)
	advanceToUpkeepOf(t, g, 0)
	if !b19HasTriggerPromptFor(g, me.ID) {
		t.Fatal("your upkeep asks")
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if hasID(p.PickTargetPlayers, me.ID) || !hasID(p.PickTargetPlayers, opp.ID) {
		t.Error("target OPPONENT")
	}
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	treasures := battlefieldIDsNamed(g, "Treasure")
	if len(treasures) != 2 {
		t.Fatalf("two Treasures, got %d", len(treasures))
	}
	for _, id := range treasures {
		switch controllerOf(t, g, id) {
		case me.ID:
			if b16Tapped(t, g, id) {
				t.Error("yours is untapped")
			}
		case opp.ID:
			if !b16Tapped(t, g, id) {
				t.Error("theirs is tapped")
			}
		default:
			t.Error("a Treasure for someone else")
		}
	}
	// The attack: damage equal to the defending player's artifacts
	// (two rocks and the tapped Treasure).
	advanceToMain(t, g)
	theirLife, otherLife := opp.Life, other.Life
	declareAttack(t, g, opp.ID, plunderer)
	passPriorityAroundTable(t, g)
	if opp.Life != theirLife-3 {
		t.Errorf("three artifacts: 3 damage, took %d", theirLife-opp.Life)
	}
	if other.Life != otherLife {
		t.Error("only the defending player")
	}
	if spec, _ := Lookup(b19GenerousPlundererOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the folded reflexive trigger must be declared")
	}
}

func TestB19LiesaReturnsYoursAndExilesTheirs(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	liesa := b12Push(g, me.ID, "Liesa, Forgotten Archangel", "Legendary Creature — Angel", b19LiesaOracle, 4, 5)
	if !hasEffectiveKeyword(t, g, liesa, "flying") || !hasEffectiveKeyword(t, g, liesa, "lifelink") {
		t.Error("flying, lifelink")
	}
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	theirOther := b16Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1, "G")
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1) })
	goblin := findBattlefieldByName(g, "Goblin")
	advanceToMain(t, g)

	// Yours: dies, comes back at the next end step.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	if !b19TriggerQueuedFrom(g, liesa) {
		t.Fatal("another nontoken creature you control died: the trigger goes on the stack")
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(bear) {
		t.Fatal("the card waits in the graveyard until the end step")
	}
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("returned to its owner's hand at the beginning of the end step")
	}
	// A token of yours: no trigger.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(goblin) })
	if b19TriggerQueuedFrom(g, liesa) {
		t.Error("a token is not a nontoken creature")
	}
	passPriorityAroundTable(t, g)

	// Theirs: exiled instead of dying, by destruction and by sacrifice.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	if !g.Exile.Contains(theirs) || opp.Graveyard.Contains(theirs) {
		t.Error("an opponent's creature that would die is exiled instead")
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(theirOther) })
	if !g.Exile.Contains(theirOther) || opp.Graveyard.Contains(theirOther) {
		t.Error("sacrifice is dying too")
	}
	if b19TriggerQueuedFrom(g, liesa) {
		t.Error("their creatures are not yours")
	}
	passPriorityAroundTable(t, g)
	// Their noncreature permanent still goes to the graveyard.
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rock) })
	if !opp.Graveyard.Contains(rock) {
		t.Error("only creatures")
	}
}
