package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// b19SacrificeOptions reads the option list of a player's open
// sacrifice prompt.
func b19SacrificeOptions(g *game.Game, chooser uuid.UUID) []uuid.UUID {
	if c := effectstest.SacrificeChoiceFor(g, chooser); c != nil {
		return effectstest.SacrificeOptions(c)
	}
	return nil
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. The two
// Guildgates are rows in the Guildgate table, so a transposed row is
// invisible until someone plays that exact card.
func TestBatch19CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B19TrinketMageOracle:            "Trinket Mage",
		effectstest.B19BirthingPodOracle:            "Birthing Pod",
		effectstest.B19CancelOracle:                 "Cancel",
		effectstest.B19SelesnyaGuildgateOracle:      "Selesnya Guildgate",
		effectstest.B19ElvishReclaimerOracle:        "Elvish Reclaimer",
		effectstest.B19DeathreapRitualOracle:        "Deathreap Ritual",
		effectstest.B19GrimBackwoodsOracle:          "Grim Backwoods",
		effectstest.B19WindingConstrictorOracle:     "Winding Constrictor",
		effectstest.B19GenerousPlundererOracle:      "Generous Plunderer",
		effectstest.B19AzoriusGuildgateOracle:       "Azorius Guildgate",
		effectstest.B19WeaverOfHarmonyOracle:        "Weaver of Harmony",
		effectstest.B19RuthlessTechnomancerOracle:   "Ruthless Technomancer",
		effectstest.B19DrownInDreamsOracle:          "Drown in Dreams",
		effectstest.B19IllustriousWanderglyphOracle: "Illustrious Wanderglyph",
		effectstest.B19LiesaOracle:                  "Liesa, Forgotten Archangel",
		effectstest.B19ErodeOracle:                  "Erode",
		effectstest.B19PerplexingTestOracle:         "Perplexing Test",
		effectstest.B19ZendikarsRoilOracle:          "Zendikar's Roil",
		effectstest.B19CracklingDoomOracle:          "Crackling Doom",
		effectstest.B19RecklessImpulseOracle:        "Reckless Impulse",
		effectstest.B19DiregrafColossusOracle:       "Diregraf Colossus",
		effectstest.B19WrennsResolveOracle:          "Wrenn's Resolve",
	}
	if len(want) != 22 {
		t.Fatalf("the batch registers 22 cards, the table lists %d", len(want))
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
}

// --- the lands -----------------------------------------------------

func TestB19GuildgatesEnterTappedAndTapForTheirColours(t *testing.T) {
	for _, tc := range []struct{ name, oracle, produced string }{
		{"Selesnya Guildgate", effectstest.B19SelesnyaGuildgateOracle, "{G|W}"},
		{"Azorius Guildgate", effectstest.B19AzoriusGuildgateOracle, "{W|U}"},
	} {
		g := effectstest.NewCatalogGame(t)
		gate := effectstest.PlayLandFromHand(t, g, tc.name, tc.oracle)
		effectstest.Top100AssertEnteredTapped(t, g, gate, tc.name)
		if effectstest.TapEventsFor(g, gate) != 0 {
			t.Errorf("%s: enters-tapped is a replacement, not a tap", tc.name)
		}
		spec, _ := effects.Lookup(tc.oracle)
		if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != tc.produced {
			t.Errorf("%s: mana abilities %+v, want one producing %s", tc.name, spec.ManaAbilities, tc.produced)
		}
	}
}

func TestB19GrimBackwoodsTapsForColorlessAndEatsACreatureToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	woods := effectstest.B12Permanent(g, me.ID, "Grim Backwoods", "Land")
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == woods {
			g.Battlefield.Cards[i].OracleID = effectstest.B19GrimBackwoodsOracle
		}
	}
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	spec, _ := effects.Lookup(effectstest.B19GrimBackwoodsOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{C}" {
		t.Errorf("mana abilities %+v, want one producing {C}", spec.ManaAbilities)
	}
	hand := me.Hand.Size()
	effectstest.B06AddMana(me, "C", "C", "B", "G")
	if err := g.ActivateCatalogAbility(me.ID, woods, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("the creature is a cost, not optional")
	}
	effectstest.B16Activate(t, g, me.ID, woods, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}})
	if g.Battlefield.Contains(bear) {
		t.Error("the creature is sacrificed as a cost")
	}
	if !effectstest.B16Tapped(t, g, woods) {
		t.Error("the land taps as its cost")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("drew %d, want 1", me.Hand.Size()-hand)
	}
}

// --- the spells ----------------------------------------------------

func TestB19CancelCountersASpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "", effectstest.B16TargetPlayer(me.ID))
	effectstest.CastCatalogSpell(t, g, "Cancel", "Instant", effectstest.B19CancelOracle, effectstest.B16TargetCard(bolt))
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) || me.Life != 40 {
		t.Error("the Bolt is countered")
	}
}

func TestB19ErodeDestroysAndOffersItsControllerABasic(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	effectstest.SeedSearchLibrary(opp,
		effectstest.SearchTestLand("Plains", "Basic Land — Plains"),
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
		effectstest.SearchTestLand("Command Tower", "Land"),
	)
	effectstest.CastCatalogSpell(t, g, "Erode", "Instant", effectstest.B19ErodeOracle, effectstest.B16TargetCard(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Graveyard.Contains(theirs) {
		t.Fatal("the target is destroyed")
	}
	c := effectstest.SearchChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("ITS CONTROLLER may search — the opponent gets the prompt")
	}
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Error("the caster searches nothing")
	}
	if effectstest.SearchOptionNamed(g, c, "Command Tower") != uuid.Nil {
		t.Error("only basic land cards are offered")
	}
	effectstest.AnswerSearchNamed(t, g, opp.ID, "Island")
	island := effectstest.FindBattlefieldByName(g, "Island")
	if island == uuid.Nil || effectstest.ControllerOf(t, g, island) != opp.ID || !effectstest.B16Tapped(t, g, island) {
		t.Error("the basic enters tapped under the opponent's control")
	}

	// A planeswalker is a legal target; an artifact is not.
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Jace", TypeLine: "Legendary Planeswalker — Jace", Owner: opp.ID, Controller: opp.ID,
		Counters: map[string]int{game.CounterLoyalty: 3},
	})
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.SeedSearchLibrary(opp, effectstest.SearchTestLand("Swamp", "Basic Land — Swamp"))
	id := effectstest.HandCardFull(me, "Erode", "Instant", "", effectstest.B19ErodeOracle, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.B16TargetCard(rock)}); err == nil {
		t.Fatal("an artifact is not a creature or planeswalker")
	}
	effectstest.CastCatalogSpell(t, g, "Erode", "Instant", effectstest.B19ErodeOracle, effectstest.B16TargetCard(walker))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(walker) {
		t.Error("the planeswalker is destroyed")
	}
	// "May" search: even one candidate is offered rather than taken.
	if effectstest.SearchChoiceFor(g, opp.ID) == nil {
		t.Fatal("the search is optional, so it always asks")
	}
	effectstest.AnswerSearchNamed(t, g, opp.ID, "Swamp")
	if effectstest.FindBattlefieldByName(g, "Swamp") == uuid.Nil {
		t.Error("the basic lands")
	}
}

func TestB19PerplexingTestBouncesTokensOrNontokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.RedGoblinToken(), 2) })
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 1) })
	effectstest.CastModal(t, g, "Perplexing Test", "Instant", effectstest.B19PerplexingTestOracle, []int{0}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16CountNamed(g, "Goblin") != 0 {
		t.Error("mode one returns every creature token")
	}
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(theirs) {
		t.Error("mode one leaves nontoken creatures alone")
	}
	hand, theirHand := me.Hand.Size(), opp.Hand.Size()
	effectstest.CastModal(t, g, "Perplexing Test", "Instant", effectstest.B19PerplexingTestOracle, []int{1}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("mode two returns every nontoken creature")
	}
	if me.Hand.Size() != hand+1 || opp.Hand.Size() != theirHand+1 {
		t.Error("to their owners' hands")
	}
}

func TestB19CracklingDoomPingsAndEdictsTheBiggest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other, third := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	mine := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	small := effectstest.B16Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 1, 1, "R")
	bigA := effectstest.B16Creature(g, opp.ID, "Their Wurm", "Creature — Wurm", 4, 4, "G")
	bigB := effectstest.B16Creature(g, opp.ID, "Their Pumped Bear", "Creature — Bear", 2, 2, "G")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bigB, "+1/+1", 2) })
	only := effectstest.B16Creature(g, other.ID, "Lone Elf", "Creature — Elf", 1, 1, "G")
	before := effectstest.B17Life(g)
	effectstest.CastCatalogSpell(t, g, "Crackling Doom", "Instant", effectstest.B19CracklingDoomOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
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
	if len(options) != 2 || !effectstest.HasID(options, bigA) || !effectstest.HasID(options, bigB) || effectstest.HasID(options, small) {
		t.Errorf("the opponent may sacrifice only a creature tied for greatest power (counters count): %v", options)
	}
	// #3664: one opponent chooses at a time; the Lone Elf's controller
	// has no choice to make, so their pick is collected unasked.
	if effectstest.SacrificeChoiceFor(g, other.ID) != nil || effectstest.SacrificeChoiceFor(g, third.ID) != nil || effectstest.SacrificeChoiceFor(g, me.ID) != nil {
		t.Error("only the opponent with a real choice is asked, and only one seat at a time")
	}
	if !g.Battlefield.Contains(only) {
		t.Error("nothing is sacrificed until every opponent has chosen")
	}
	effectstest.AnswerSacrifice(t, g, opp.ID, bigB)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bigB) || g.Battlefield.Contains(only) || !g.Battlefield.Contains(bigA) || !g.Battlefield.Contains(mine) {
		t.Error("the chosen creatures are sacrificed and nothing else")
	}
}

func TestB19DrownInDreamsDrawsXOrMillsTwiceX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hand := me.Hand.Size()
	effectstest.B19CastXModal(t, g, "Drown in Dreams", "Instant", effectstest.B19DrownInDreamsOracle, "{X}{2}{U}", 3, []int{0}, effectstest.B16TargetPlayer(me.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 {
		t.Errorf("X=3 draws 3, drew %d", me.Hand.Size()-hand)
	}
	library := opp.Library.Size()
	effectstest.B19CastXModal(t, g, "Drown in Dreams", "Instant", effectstest.B19DrownInDreamsOracle, "{X}{2}{U}", 3, []int{1}, effectstest.B16TargetPlayer(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != library-6 || opp.Graveyard.Size() < 6 {
		t.Errorf("X=3 mills 6, milled %d", library-opp.Library.Size())
	}
	// No commander on the battlefield: both is refused (#1590 — the
	// rider's condition is false, so the bound stays one).
	id := effectstest.HandCardFull(me, "Drown in Dreams", "Instant", "{X}{2}{U}", effectstest.B19DrownInDreamsOracle, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0, 1}, Targets: append(effectstest.B16TargetPlayer(me.ID), game.TargetRef{Kind: game.TargetPlayer, ID: me.ID, Mode: 1}), XValue: 1}); err == nil {
		t.Error("without a commander, both modes is refused")
	}
}

func TestB19RecklessImpulseGrantsPlayUntilTheEndOfYourNextTurn(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		seat         int
	}{
		{"Reckless Impulse", effectstest.B19RecklessImpulseOracle, 0},
		{"Wrenn's Resolve", effectstest.B19WrennsResolveOracle, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[tc.seat]
			effectstest.AdvanceToMainOf(t, g, tc.seat)
			ids := effectstest.SeedSearchLibrary(me,
				game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
				game.Card{Name: "Shock", TypeLine: "Instant"},
				game.Card{Name: "Deep Card", TypeLine: "Sorcery"},
			)
			forest, shock, deep := ids[0], ids[1], ids[2]
			effectstest.CastCatalogSpell(t, g, tc.name, "Sorcery", tc.oracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			if !g.Exile.Contains(forest) || !g.Exile.Contains(shock) || !me.Library.Contains(deep) {
				t.Fatal("the top two cards are exiled")
			}
			for _, id := range []uuid.UUID{forest, shock} {
				perm := effectstest.ExiledPermission(g, id)
				if perm.Player != me.ID || perm.CastOnly || !effectstest.PermissionLive(g, perm, me.ID) {
					t.Errorf("%+v: the caster may PLAY it now", perm)
				}
			}
			// #945: the duration IS the clause, so the card schedules
			// nothing. Before the swap this queued a CR 603.7 delayed
			// trigger in the caster's next upkeep whose only job was
			// to re-stamp the window.
			if len(g.DelayedTriggers) != 0 {
				t.Errorf("%d delayed triggers scheduled, want none", len(g.DelayedTriggers))
			}
			// The land can be played this turn.
			if err := g.CastSpell(me.ID, forest, game.CastSpellParams{FromZone: "exile"}); err != nil {
				t.Fatalf("playing the exiled land this turn: %v", err)
			}
			// Through every opponent's turn the grant is still live.
			effectstest.AdvanceToMainOf(t, g, (tc.seat+1)%4)
			if !effectstest.PermissionLive(g, effectstest.ExiledPermission(g, shock), me.ID) {
				t.Fatal("the grant survives the caster's own cleanup")
			}
			effectstest.AdvanceToMainOf(t, g, (tc.seat+3)%4)
			if !effectstest.PermissionLive(g, effectstest.ExiledPermission(g, shock), me.ID) {
				t.Fatal("the grant survives seat 0's cleanup of the next round too")
			}
			// Your next turn, which the clause reaches the END of.
			// #945: one duration says so — ADR 0063's seat-turn
			// counter — where this used to need a backstop window and
			// a delayed trigger in your upkeep to pull it back.
			effectstest.AdvanceToMainOf(t, g, tc.seat)
			perm := effectstest.ExiledPermission(g, shock)
			if perm.Duration.Kind != game.UntilEndOfTurn || !effectstest.PermissionLive(g, perm, me.ID) {
				t.Errorf("on your next turn the grant is %+v, want live", perm)
			}
			if !g.Battlefield.Contains(forest) {
				t.Error("the land played last turn is untouched")
			}
			// Still castable at your end step; gone on the next player's turn.
			effectstest.AdvanceToEndStepOf(t, g, tc.seat)
			if !effectstest.PermissionLive(g, effectstest.ExiledPermission(g, shock), me.ID) {
				t.Error("the end of your next turn is still your next turn")
			}
			effectstest.AdvanceToMainOf(t, g, (tc.seat+1)%4)
			if effectstest.ExiledPermission(g, shock).Granted() {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Sol Ring", TypeLine: "Artifact", ManaCost: "{1}"},
		game.Card{Name: "Mox Opal", TypeLine: "Legendary Artifact", ManaCost: ""},
		game.Card{Name: "Mind Stone", TypeLine: "Artifact", ManaCost: "{2}"},
		game.Card{Name: "Myr", TypeLine: "Artifact Creature — Myr", ManaCost: "{1}"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{G}"},
	)
	effectstest.CastCatalogSpell(t, g, "Trinket Mage", "Creature — Human Wizard", effectstest.B19TrinketMageOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("you may search: the chooser asks")
	}
	for _, name := range []string{"Sol Ring", "Mox Opal", "Myr"} {
		if effectstest.SearchOptionNamed(g, c, name) == uuid.Nil {
			t.Errorf("%s (an artifact with mana value 1 or less) is offered", name)
		}
	}
	if effectstest.SearchOptionNamed(g, c, "Mind Stone") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("mana value 2, and a non-artifact, are not")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Sol Ring")
	if !effectstest.B02bHandHasNamed(me, "Sol Ring") {
		t.Error("the pick goes to hand")
	}
	if me.Library.Size() != 4 {
		t.Errorf("library %d, want 4", me.Library.Size())
	}
}

func TestB19RuthlessTechnomancerTradesACreatureForTreasures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	beast := effectstest.B16Creature(g, me.ID, "Beast", "Creature — Beast", 4, 4, "G")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(beast, "+1/+1", 1) })
	mancer := effectstest.CastAndResolveCreature(t, g, "Ruthless Technomancer", "Creature — Human Wizard", effectstest.B19RuthlessTechnomancerOracle)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil {
		t.Fatal("the sacrifice is asked as the trigger resolves")
	}
	if p.ChooseMin != 0 || p.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want 0..1 — \"you MAY sacrifice\"", p.ChooseMin, p.ChooseMax)
	}
	if effectstest.HasID(p.ChooseCards, mancer) || !effectstest.HasID(p.ChooseCards, beast) {
		t.Error("ANOTHER creature you control: the Technomancer itself is not offered")
	}
	effectstest.AnswerOwnPermanents(t, g, me.ID, beast)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(beast) {
		t.Fatal("the chosen creature is sacrificed")
	}
	if n := effectstest.B16CountNamed(g, "Treasure"); n != 5 {
		t.Errorf("a 4/4 with a +1/+1 counter: 5 Treasures, got %d", n)
	}

	// Declined: nothing happens.
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.CastAndResolveCreature(t, g, "Ruthless Technomancer", "Creature — Human Wizard", effectstest.B19RuthlessTechnomancerOracle)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) || effectstest.B16CountNamed(g, "Treasure") != 5 {
		t.Error("declining sacrifices nothing and makes nothing")
	}
	// The sacrifice is chosen at resolution and the reanimation ability is
	// registered, so nothing is left to declare.
	if spec, _ := effects.Lookup(effectstest.B19RuthlessTechnomancerOracle); spec.Completeness != effects.CompletenessFull || len(spec.Activated) != 1 || len(spec.Caveats) != 0 {
		t.Error("the card is complete: the reanimation ability is registered and the sacrifice is chosen at resolution")
	}
}

func TestB19BirthingPodChainsUpOneManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	pod := effectstest.B12Permanent(g, me.ID, "Birthing Pod", "Artifact")
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == pod {
			g.Battlefield.Cards[i].OracleID = effectstest.B19BirthingPodOracle
		}
	}
	two := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Two-Drop", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Three-Drop", TypeLine: "Creature — Elf", ManaCost: "{2}{G}"},
		game.Card{Name: "Other Three", TypeLine: "Creature — Wurm", ManaCost: "{1}{G}{G}"},
		game.Card{Name: "Four-Drop", TypeLine: "Creature — Giant", ManaCost: "{3}{G}"},
		game.Card{Name: "Three Rock", TypeLine: "Artifact", ManaCost: "{3}"},
	)
	spec, _ := effects.Lookup(effectstest.B19BirthingPodOracle)
	if len(spec.Activated) != 1 || !spec.Activated[0].SorcerySpeed {
		t.Fatal("one ability, sorcery speed")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "G")
	effectstest.B16Activate(t, g, me.ID, pod, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{two}})
	if g.Battlefield.Contains(two) || !effectstest.B16Tapped(t, g, pod) {
		t.Fatal("the creature is sacrificed and the Pod taps")
	}
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("two creatures at mana value 3: the searcher chooses")
	}
	if effectstest.SearchOptionNamed(g, c, "Four-Drop") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Three Rock") != uuid.Nil {
		t.Error("only creature cards with mana value exactly 3 are offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Other Three")
	wurm := effectstest.FindBattlefieldByName(g, "Other Three")
	if wurm == uuid.Nil || effectstest.B16Tapped(t, g, wurm) {
		t.Error("the pick enters the battlefield untapped")
	}
	// #917 / #916 closed both halves of the Phyrexian gap this card
	// declared: the activation announces the life payment and the
	// board offers it, so nothing about {G/P} is simplified any more.
	if spec.Completeness != effects.CompletenessFull {
		t.Error("the Phyrexian mana gap is closed; Birthing Pod is complete")
	}
}

func TestB19ElvishReclaimerGrowsAndCropRotates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	reclaimer := effectstest.B12Push(g, me.ID, "Elvish Reclaimer", "Creature — Elf Warrior", effectstest.B19ElvishReclaimerOracle, 1, 2)
	forest := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.B17GraveyardCard(me, "Dead Forest", "Basic Land — Forest", "")
	effectstest.B17GraveyardCard(me, "Dead Swamp", "Basic Land — Swamp", "")
	g.WithWriteLock(func() { g.BumpLayerVersionForTest() })
	if effectstest.EffectivePower(t, g, reclaimer) != 1 {
		t.Fatal("two land cards in the graveyard: still 1/2")
	}
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Gaea's Cradle", "Legendary Land"),
		effectstest.SearchTestLand("Plains", "Basic Land — Plains"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C")
	effectstest.B16Activate(t, g, me.ID, reclaimer, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{forest}})
	if g.Battlefield.Contains(forest) || !effectstest.B16Tapped(t, g, reclaimer) {
		t.Fatal("the land is sacrificed and the Reclaimer taps")
	}
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("two land cards: the searcher chooses")
	}
	if effectstest.SearchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("only land cards are offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Gaea's Cradle")
	cradle := effectstest.FindBattlefieldByName(g, "Gaea's Cradle")
	if cradle == uuid.Nil || !effectstest.B16Tapped(t, g, cradle) {
		t.Error("any land card, onto the battlefield tapped")
	}
	// The sacrificed Forest is the third land card in the graveyard.
	if got := effectstest.EffectivePower(t, g, reclaimer); got != 3 {
		t.Errorf("three land cards in the graveyard: 3/4, got power %d", got)
	}
	if got := effectstest.EffectiveToughness(t, g, reclaimer); got != 4 {
		t.Errorf("toughness %d, want 4", got)
	}
}

func TestB19DiregrafColossusCountsZombiesAndMakesTappedOnes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B17GraveyardCard(me, "Dead Zombie", "Creature — Zombie", "")
	effectstest.B17GraveyardCard(me, "Dead Changeling", "Creature — Shapeshifter", "")
	effectstest.B17GraveyardCard(me, "Dead Bear", "Creature — Bear", "")
	effectstest.B17GraveyardCard(opp, "Their Dead Zombie", "Creature — Zombie", "")
	colossus := effectstest.CastCatalogSpell(t, g, "Diregraf Colossus", "Creature — Zombie Giant", effectstest.B19DiregrafColossusOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, colossus, "+1/+1"); n != 1 {
		t.Fatalf("one Zombie card in YOUR graveyard: 1 counter, got %d", n)
	}
	effectstest.CastCatalogSpell(t, g, "Gravecrawler", "Creature — Zombie", "", nil)
	if effectstest.TriggerOnStack(g, colossus) == nil {
		t.Fatal("casting a Zombie spell triggers")
	}
	effectstest.PassPriorityAroundTable(t, g)
	zombies := effectstest.BattlefieldIDsNamed(g, "Zombie")
	if len(zombies) != 1 {
		t.Fatalf("one tapped 2/2 Zombie token, got %d", len(zombies))
	}
	if !effectstest.B16Tapped(t, g, zombies[0]) || effectstest.EffectivePower(t, g, zombies[0]) != 2 {
		t.Error("the token is tapped and 2/2")
	}
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	if effectstest.TriggerOnStack(g, colossus) != nil {
		t.Error("a non-Zombie spell does not trigger")
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// --- the creatures: statics and replacements -----------------------

func TestB19WindingConstrictorAddsOneCounterOfEachKind(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Winding Constrictor", "Creature — Snake", effectstest.B19WindingConstrictorOracle, 2, 3)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	rock := effectstest.B12Permanent(g, me.ID, "Coffers", "Artifact")
	shrine := effectstest.B12Permanent(g, me.ID, "Shrine", "Enchantment")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(bear, "+1/+1", 1)
		_ = g.AddCounterForEffect(rock, "charge", 2)
		_ = g.AddCounterForEffect(shrine, "verse", 1)
		_ = g.AddCounterForEffect(theirs, "+1/+1", 1)
	})
	if effectstest.CounterCount(g, bear, "+1/+1") != 2 {
		t.Errorf("one +1/+1 counter becomes two, got %d", effectstest.CounterCount(g, bear, "+1/+1"))
	}
	if effectstest.CounterCount(g, rock, "charge") != 3 {
		t.Errorf("two charge counters on an artifact become three, got %d", effectstest.CounterCount(g, rock, "charge"))
	}
	if effectstest.CounterCount(g, shrine, "verse") != 1 || effectstest.CounterCount(g, theirs, "+1/+1") != 1 {
		t.Error("an enchantment, and an opponent's creature, get the printed number")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, "+1/+1", -1) })
	if effectstest.CounterCount(g, bear, "+1/+1") != 1 {
		t.Errorf("a removal is not counters being put on: %d, want 1", effectstest.CounterCount(g, bear, "+1/+1"))
	}
	// "Enters with" counters run the same pipeline: a Diregraf
	// Colossus with two Zombies in the graveyard enters with three.
	effectstest.B17GraveyardCard(me, "Dead Zombie", "Creature — Zombie", "")
	effectstest.B17GraveyardCard(me, "Dead Ghoul", "Creature — Zombie", "")
	colossus := effectstest.CastCatalogSpell(t, g, "Diregraf Colossus", "Creature — Zombie Giant", effectstest.B19DiregrafColossusOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, colossus, "+1/+1"); n != 3 {
		t.Errorf("enters with 2+1 counters, got %d", n)
	}
	if spec, _ := effects.Lookup(effectstest.B19WindingConstrictorOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("the card is complete since the player-counter half landed (#3172)")
	}
}

func TestB19WeaverOfHarmonyBuffsOtherEnchantmentCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	weaver := effectstest.B12Push(g, me.ID, "Weaver of Harmony", "Enchantment Creature — Snake Druid", effectstest.B19WeaverOfHarmonyOracle, 2, 2)
	kami := effectstest.B16Creature(g, me.ID, "Kami", "Enchantment Creature — Spirit", 1, 1, "W")
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	shrine := effectstest.B12Permanent(g, me.ID, "Shrine", "Enchantment")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Kami", "Enchantment Creature — Spirit", 1, 1, "W")
	if effectstest.EffectivePower(t, g, kami) != 2 || effectstest.EffectiveToughness(t, g, kami) != 2 {
		t.Error("another enchantment creature you control gets +1/+1")
	}
	if effectstest.EffectivePower(t, g, weaver) != 2 || effectstest.EffectivePower(t, g, bear) != 2 || effectstest.EffectivePower(t, g, theirs) != 1 {
		t.Error("not itself, not a plain creature, not an opponent's")
	}
	_ = shrine
	// #1223 built the ability-copy seam the Weaver was waiting on, so
	// the copy activation is there and the caveat is gone.
	if spec, _ := effects.Lookup(effectstest.B19WeaverOfHarmonyOracle); spec.Completeness != effects.CompletenessFull || len(spec.Activated) != 1 {
		t.Error("the copy ability ships with #1223")
	}
}

func TestB19IllustriousWanderglyphMakesGnomesAndLordsThemAtTen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	glyph := effectstest.B12Push(g, me.ID, "Illustrious Wanderglyph", "Artifact Creature — Golem", effectstest.B19IllustriousWanderglyphOracle, 2, 2)
	myr := effectstest.B16Creature(g, me.ID, "Myr", "Artifact Creature — Myr", 1, 1)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirMyr := effectstest.B16Creature(g, opp.ID, "Their Myr", "Artifact Creature — Myr", 1, 1)
	// Four permanents: no blessing.
	if effectstest.EffectivePower(t, g, myr) != 1 {
		t.Fatal("below ten permanents: no bonus")
	}
	// Each upkeep, anyone's, makes a Gnome for the controller.
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B16CountNamed(g, "Gnome"); n != 1 {
		t.Fatalf("an opponent's upkeep: 1 Gnome, got %d", n)
	}
	gnome := effectstest.FindBattlefieldByName(g, "Gnome")
	if effectstest.ControllerOf(t, g, gnome) != me.ID {
		t.Error("the Gnome is the Wanderglyph's controller's")
	}
	for i := 0; i < 5; i++ {
		effectstest.B12Permanent(g, me.ID, "Land", "Basic Land — Forest")
	}
	// Glyph, Myr, Bear, Gnome, five lands: nine. One more.
	if effectstest.EffectivePower(t, g, myr) != 1 {
		t.Fatal("nine permanents: still no bonus")
	}
	effectstest.B12Permanent(g, me.ID, "Last Land", "Basic Land — Forest")
	g.RunStateChecksForTest()
	if effectstest.EffectivePower(t, g, myr) != 3 || effectstest.EffectivePower(t, g, gnome) != 3 {
		t.Error("ten permanents: other artifact creatures you control get +2/+2")
	}
	if effectstest.EffectivePower(t, g, glyph) != 2 || effectstest.EffectivePower(t, g, bear) != 2 || effectstest.EffectivePower(t, g, theirMyr) != 1 {
		t.Error("not itself, not a non-artifact, not an opponent's")
	}
	// The city's blessing is kept (CR 702.131c, #2696).
	g.WithWriteLock(func() {
		removed := 0
		for i := len(g.Battlefield.Cards) - 1; i >= 0 && removed < 6; i-- {
			if c := g.Battlefield.Cards[i]; c.Controller == me.ID && c.Name == "Land" {
				_, _ = g.Battlefield.Remove(c.InstanceID)
				removed++
			}
		}
		g.BumpLayerVersionForTest()
	})
	if effectstest.EffectivePower(t, g, myr) != 3 {
		t.Error("back under ten: the +2/+2 stays, the blessing is kept")
	}
	if spec, _ := effects.Lookup(effectstest.B19IllustriousWanderglyphOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("the blessing is a real designation now, so the card is full")
	}
}

// --- the creatures: triggers ---------------------------------------

func TestB19ZendikarsRoilMakesAnElementalPerLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Zendikar's Roil", "Enchantment", effectstest.B19ZendikarsRoilOracle, 0, 0)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	elementals := effectstest.BattlefieldIDsNamed(g, "Elemental")
	if len(elementals) != 1 {
		t.Fatalf("a land you control entering: one Elemental, got %d", len(elementals))
	}
	if c := effectstest.CardByID(g, elementals[0]); effectstest.EffectivePower(t, g, elementals[0]) != 2 || len(c.Colors) != 1 || c.Colors[0] != "G" || effectstest.ControllerOf(t, g, elementals[0]) != me.ID {
		t.Error("a 2/2 green Elemental under your control")
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.PlayLandFromHand(t, g, "Their Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16CountNamed(g, "Elemental") != 1 {
		t.Error("an opponent's land is not yours")
	}
	_ = opp
}

func TestB19DeathreapRitualDrawsAtEndStepsAfterADeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Deathreap Ritual", "Enchantment", effectstest.B19DeathreapRitualOracle, 0, 0)
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	// No death this turn: the end step passes without a prompt.
	effectstest.AdvanceToEndStepOf(t, g, 0)
	if effectstest.B19HasTriggerPromptFor(g, me.ID) {
		t.Fatal("no creature died: no trigger (intervening if)")
	}
	// A death on an opponent's turn: their end step asks you.
	effectstest.AdvanceToMainOf(t, g, 1)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	hand := me.Hand.Size()
	effectstest.AdvanceToEndStepOf(t, g, 1)
	if !effectstest.B19HasTriggerPromptFor(g, me.ID) {
		t.Fatal("a creature died this turn: each end step asks the controller")
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("drew %d, want 1", me.Hand.Size()-hand)
	}
	// The next turn starts clean.
	effectstest.AdvanceToEndStepOf(t, g, 2)
	if effectstest.B19HasTriggerPromptFor(g, me.ID) {
		t.Error("the tally resets with the turn")
	}
}

func TestB19GenerousPlundererGiftsTreasureAndPunishesArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	plunderer := effectstest.B12Push(g, me.ID, "Generous Plunderer", "Creature — Human Rogue", effectstest.B19GenerousPlundererOracle, 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, plunderer, "menace") {
		t.Error("menace")
	}
	effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.B12Permanent(g, opp.ID, "Their Signet", "Artifact")
	// Your NEXT upkeep (the game opens in this one): may make a
	// Treasure; the chosen opponent gets a tapped one.
	effectstest.AdvanceToMain(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if !effectstest.B19HasTriggerPromptFor(g, me.ID) {
		t.Fatal("your upkeep asks")
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.B04WaitForPick(t, g, me.ID)
	// CR 603.12 / 603.3d (#636): the opponent is picked for the
	// REFLEXIVE trigger, which exists only because the upkeep ability
	// already resolved — so your Treasure is on the battlefield by the
	// time the prompt opens, and theirs is not.
	if got := len(effectstest.BattlefieldIDsNamed(g, "Treasure")); got != 1 {
		t.Errorf("%d Treasures when the reflexive trigger asks for its target, want just yours", got)
	}
	p := effectstest.LatestPickTarget(g, me.ID)
	if effectstest.HasID(p.PickTargetPlayers, me.ID) || !effectstest.HasID(p.PickTargetPlayers, opp.ID) {
		t.Error("target OPPONENT")
	}
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	// And it is a stack item of its own, with a response window: the
	// second Treasure has not been made yet.
	if effectstest.TriggerOnStack(g, plunderer) == nil {
		t.Error("the reflexive trigger should be on the stack, not folded into the upkeep ability")
	}
	if got := len(effectstest.BattlefieldIDsNamed(g, "Treasure")); got != 1 {
		t.Errorf("%d Treasures while the reflexive trigger is still on the stack, want 1", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	treasures := effectstest.BattlefieldIDsNamed(g, "Treasure")
	if len(treasures) != 2 {
		t.Fatalf("two Treasures, got %d", len(treasures))
	}
	for _, id := range treasures {
		switch effectstest.ControllerOf(t, g, id) {
		case me.ID:
			if effectstest.B16Tapped(t, g, id) {
				t.Error("yours is untapped")
			}
		case opp.ID:
			if !effectstest.B16Tapped(t, g, id) {
				t.Error("theirs is tapped")
			}
		default:
			t.Error("a Treasure for someone else")
		}
	}
	// The attack: damage equal to the defending player's artifacts
	// (two rocks and the tapped Treasure).
	effectstest.AdvanceToMain(t, g)
	theirLife, otherLife := opp.Life, other.Life
	effectstest.DeclareAttack(t, g, opp.ID, plunderer)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirLife-3 {
		t.Errorf("three artifacts: 3 damage, took %d", theirLife-opp.Life)
	}
	if other.Life != otherLife {
		t.Error("only the defending player")
	}
	if spec, _ := effects.Lookup(effectstest.B19GenerousPlundererOracle); spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Error("the reflexive trigger (#636) closed the card's only gap")
	}
}

func TestB19LiesaReturnsYoursAndExilesTheirs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	liesa := effectstest.B12Push(g, me.ID, "Liesa, Forgotten Archangel", "Legendary Creature — Angel", effectstest.B19LiesaOracle, 4, 5)
	if !effectstest.HasEffectiveKeyword(t, g, liesa, "flying") || !effectstest.HasEffectiveKeyword(t, g, liesa, "lifelink") {
		t.Error("flying, lifelink")
	}
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	theirOther := effectstest.B16Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1, "G")
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 1) })
	goblin := effectstest.FindBattlefieldByName(g, "Goblin")
	effectstest.AdvanceToMain(t, g)

	// Yours: dies, comes back at the next end step.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	if !effectstest.B19TriggerQueuedFrom(g, liesa) {
		t.Fatal("another nontoken creature you control died: the trigger goes on the stack")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(bear) {
		t.Fatal("the card waits in the graveyard until the end step")
	}
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("returned to its owner's hand at the beginning of the end step")
	}
	// A token of yours: no trigger.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(goblin) })
	if effectstest.B19TriggerQueuedFrom(g, liesa) {
		t.Error("a token is not a nontoken creature")
	}
	effectstest.PassPriorityAroundTable(t, g)

	// Theirs: exiled instead of dying, by destruction and by sacrifice.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	if !g.Exile.Contains(theirs) || opp.Graveyard.Contains(theirs) {
		t.Error("an opponent's creature that would die is exiled instead")
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(theirOther) })
	if !g.Exile.Contains(theirOther) || opp.Graveyard.Contains(theirOther) {
		t.Error("sacrifice is dying too")
	}
	if effectstest.B19TriggerQueuedFrom(g, liesa) {
		t.Error("their creatures are not yours")
	}
	effectstest.PassPriorityAroundTable(t, g)
	// Their noncreature permanent still goes to the graveyard.
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rock) })
	if !opp.Graveyard.Contains(rock) {
		t.Error("only creatures")
	}
}

// #1337: CR 903.9a makes a commander's trip to the command zone a
// state-based "may" AFTER the move happens, so an opponent's dying
// commander is exiled by Liesa exactly like any other of their
// creatures, and its OWNER (not Liesa's controller) then separately
// decides whether to send it to the command zone instead — the same
// reasoning that closed Cosmic Intervention's identical stale caveat.
// "An opponent's dying commander still goes to the command zone" was
// never true once the two replacements were both registered normally
// and met in the CR 616 apply loop.
//
// ADR 0115: there is no CR 616 ordering any more. Liesa's replacement is
// the only one that applies, the commander is exiled, and the CR 903.9a
// state-based action then asks its owner.
func TestB19LiesaExilesAnOpponentsDyingCommanderThenAsksItsOwner(t *testing.T) {
	for _, takeCommandZone := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.B12Push(g, me.ID, "Liesa, Forgotten Archangel", "Legendary Creature — Angel", effectstest.B19LiesaOracle, 4, 5)
		commander := uuid.New()
		g.Battlefield.PushTop(game.Card{
			InstanceID: commander, Name: "Their Commander", TypeLine: "Legendary Creature — Human Wizard",
			Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID, IsCommander: true,
		})

		g.WithWriteLock(func() {
			if err := g.DestroyPermanentForEffect(commander); err != nil {
				t.Fatalf("DestroyPermanentForEffect: %v", err)
			}
		})
		for _, c := range g.PendingChoices {
			if c != nil && (c.Kind == game.PendingChoiceReplacementOrder || c.Kind == game.PendingChoiceOptionalReplacement) {
				t.Fatalf("unexpected %s prompt: only Liesa's replacement applies to a dying commander", c.Kind)
			}
		}
		if opp.Graveyard.Contains(commander) {
			t.Fatal("the commander hit the graveyard under Liesa")
		}
		if !effectstest.ExileHas(g, commander) {
			t.Fatal("Liesa did not exile the dying commander")
		}
		if effectstest.CommanderReturnPromptFor(g, opp.ID) == nil {
			t.Fatal("the commander's owner was never offered the command zone")
		}
		effectstest.AnswerCommanderReturn(t, g, opp.ID, takeCommandZone)
		if takeCommandZone {
			if !opp.Command.Contains(commander) {
				t.Fatal("the owner took the command zone but the commander is not there")
			}
			continue
		}
		if !effectstest.ExileHas(g, commander) {
			t.Fatal("declined commander is not in exile")
		}
	}
}
