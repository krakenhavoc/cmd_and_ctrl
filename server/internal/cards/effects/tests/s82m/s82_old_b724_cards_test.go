package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b724_cards_test.go — Urza's Saga verse counters, S82 OLD batch 724
// (#5624). Every test plays the real catalog entry through the engine.

const (
	b724Recantation       = "02269353-2949-4599-b5a3-f20fdb501285"
	b724TitaniasBoon      = "0bf4c01d-3cb0-47be-b84a-6a399bc2d3fc"
	b724SerrasLiturgy     = "17b7b7e4-4ae8-495d-bf0c-d18697f250ea"
	b724RumblingCrescendo = "2e7e19c2-5399-433d-9653-d956f11838f5"
	b724SerrasHymn        = "66a3ffde-4c81-4798-b99b-a6cc3a93c3e3"
	b724TorchSong         = "69826a0d-3ec9-47e7-b129-2fc13b41a07e"
	b724Vebulid           = "a887def5-c9e1-4cbb-8b08-79b7bd808deb"
	b724MidsummerRevel    = "d3579282-5045-4b37-9998-9a282e9932b9"
	b724WarDance          = "ecfc81df-1271-47ca-9184-49bf5798e195"
	b724DiscordantDirge   = "edd4741b-3ee6-421a-8bac-d60736755356"
	b724LiltingRefrain    = "f91a71b7-9bc0-4609-a312-ffede52dc0ad"
)

func TestB724EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b724Recantation:       "Recantation",
		b724TitaniasBoon:      "Titania's Boon",
		b724SerrasLiturgy:     "Serra's Liturgy",
		b724RumblingCrescendo: "Rumbling Crescendo",
		b724SerrasHymn:        "Serra's Hymn",
		b724TorchSong:         "Torch Song",
		b724Vebulid:           "Vebulid",
		b724MidsummerRevel:    "Midsummer Revel",
		b724WarDance:          "War Dance",
		b724DiscordantDirge:   "Discordant Dirge",
		b724LiltingRefrain:    "Lilting Refrain",
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

// b724Verse pushes the enchantment for `p` with `n` verse counters.
func b724Verse(g *game.Game, p *game.Player, name, oracle string, n int) uuid.UUID {
	id := effectstest.PushCatalogPermanent(g, p.ID, name, "Enchantment", oracle, false)
	if n > 0 {
		effectstest.AddCountersForTest(g, id, "verse", n)
	}
	return id
}

func b724Activate(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID, pool string, params game.ActivateAbilityParams) error {
	t.Helper()
	if pool != "" {
		effectstest.AddPool(t, g, p.ID, pool)
	}
	params.Strict = true
	return g.ActivateCatalogAbility(p.ID, id, 0, params)
}

func b724Settle(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.PassPriorityAroundTable(t, g)
}

// The upkeep counter is a "you may": yes adds one, no adds none.
func TestB724VerseCardsPutACounterOnYourUpkeepOnlyIfYouSayYes(t *testing.T) {
	cards := []struct{ name, oracle string }{
		{"Recantation", b724Recantation}, {"Serra's Liturgy", b724SerrasLiturgy}, {"Rumbling Crescendo", b724RumblingCrescendo},
		{"Serra's Hymn", b724SerrasHymn}, {"Torch Song", b724TorchSong}, {"Midsummer Revel", b724MidsummerRevel},
		{"War Dance", b724WarDance}, {"Discordant Dirge", b724DiscordantDirge}, {"Lilting Refrain", b724LiltingRefrain},
	}
	for _, c := range cards {
		for _, yes := range []bool{true, false} {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			id := b724Verse(g, me, c.name, c.oracle, 0)
			b724AdvanceToUpkeepOf(t, g, 0)
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
			b724Settle(t, g)
			want := 0
			if yes {
				want = 1
			}
			if got := effectstest.ApaLive(g, id).Counters["verse"]; got != want {
				t.Errorf("%s yes=%v: %d verse counters, want %d", c.name, yes, got, want)
			}
		}
	}
}

func b724AdvanceToUpkeepOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	start := g.Turn.Seq
	for g.Turn.ActiveSeat != seat || g.Turn.Step != game.StepUpkeep || g.Turn.Seq <= start {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}

func TestB724RecantationReturnsUpToXPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rec := b724Verse(g, me, "Recantation", b724Recantation, 2)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Bear", 2, 2)
	three := []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}, {Kind: game.TargetCard, ID: c}}
	if err := b724Activate(t, g, me, rec, "{U}", game.ActivateAbilityParams{Targets: three}); err == nil {
		t.Fatal("three targets were accepted with two verse counters")
	}
	if err := b724Activate(t, g, me, rec, "", game.ActivateAbilityParams{Targets: three[:2]}); err != nil {
		t.Fatalf("activate with two targets: %v", err)
	}
	b724Settle(t, g)
	if g.Battlefield.Contains(rec) {
		t.Error("the enchantment was not sacrificed")
	}
	if !opp.Hand.Contains(a) || !opp.Hand.Contains(b) || g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("the two targets did not return to their owner's hand")
	}
	if !g.Battlefield.Contains(c) {
		t.Error("the untargeted permanent returned")
	}
}

func TestB724RecantationWithNoCountersTargetsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rec := b724Verse(g, me, "Recantation", b724Recantation, 0)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	if err := b724Activate(t, g, me, rec, "{U}", game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}}}); err == nil {
		t.Fatal("a target was accepted with no verse counters")
	}
	if err := b724Activate(t, g, me, rec, "", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate with no targets: %v", err)
	}
	b724Settle(t, g)
	if g.Battlefield.Contains(rec) || !g.Battlefield.Contains(a) {
		t.Error("with X = 0 the enchantment should be sacrificed and nothing returned")
	}
}

func TestB724SerrasLiturgyDestroysUpToXArtifactsAndEnchantments(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lit := b724Verse(g, me, "Serra's Liturgy", b724SerrasLiturgy, 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	aura := effectstest.B12Permanent(g, opp.ID, "Ward", "Enchantment")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := b724Activate(t, g, me, lit, "{W}", game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Fatal("a plain creature was a legal target")
	}
	if err := b724Activate(t, g, me, lit, "", game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}, {Kind: game.TargetCard, ID: aura}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	b724Settle(t, g)
	if g.Battlefield.Contains(rock) || g.Battlefield.Contains(aura) || g.Battlefield.Contains(lit) {
		t.Error("the artifact, the enchantment and the sacrificed Liturgy should all be gone")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature was destroyed")
	}
}

func TestB724RumblingCrescendoDestroysUpToXLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cre := b724Verse(g, me, "Rumbling Crescendo", b724RumblingCrescendo, 1)
	land := effectstest.SeedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	other := effectstest.SeedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island")
	two := []game.TargetRef{{Kind: game.TargetCard, ID: land}, {Kind: game.TargetCard, ID: other}}
	if err := b724Activate(t, g, me, cre, "{R}", game.ActivateAbilityParams{Targets: two}); err == nil {
		t.Fatal("two lands were accepted with one verse counter")
	}
	if err := b724Activate(t, g, me, cre, "", game.ActivateAbilityParams{Targets: two[:1]}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	b724Settle(t, g)
	if g.Battlefield.Contains(land) || !g.Battlefield.Contains(other) {
		t.Error("exactly the one targeted land should be destroyed")
	}
}

func TestB724SerrasHymnPreventsXDamageDividedAsYouChoose(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hymn := b724Verse(g, me, "Serra's Hymn", b724SerrasHymn, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 5)
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetPlayer, ID: me.ID}}
	if err := b724Activate(t, g, me, hymn, "", game.ActivateAbilityParams{Targets: targets, Distribution: map[uuid.UUID]int{bear: 3, me.ID: 1}}); err == nil {
		t.Fatal("a division of four was accepted with three verse counters")
	}
	if err := b724Activate(t, g, me, hymn, "", game.ActivateAbilityParams{Targets: targets, Distribution: map[uuid.UUID]int{bear: 2, me.ID: 1}}); err != nil {
		t.Fatalf("a legal division: %v", err)
	}
	b724Settle(t, g)
	if g.Battlefield.Contains(hymn) {
		t.Error("the Hymn was not sacrificed")
	}
	life := me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(opp.ID, bear, 4) })
	if d := effectstest.DamageMarkedOn(g, bear); d != 2 {
		t.Errorf("4 damage to the Bear with a 2 shield left %d marked, want 2", d)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(opp.ID, me.ID, 3) })
	if me.Life != life-2 {
		t.Errorf("3 damage to me with a 1 shield: life %d → %d, want -2", life, me.Life)
	}
}

func TestB724TorchSongDealsXDamageToAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	song := b724Verse(g, me, "Torch Song", b724TorchSong, 3)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Bear", 5, 5)
	if err := b724Activate(t, g, me, song, "{C}{C}", game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: big}}}); err == nil {
		t.Fatal("activated without the {R}")
	}
	if err := b724Activate(t, g, me, song, "{R}", game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: big}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	b724Settle(t, g)
	if g.Battlefield.Contains(song) {
		t.Error("the enchantment was not sacrificed")
	}
	if got := effectstest.DamageMarkedOn(g, big); got != 3 {
		t.Errorf("the target took %d, want 3 (the verse counters it had)", got)
	}
	// A player is a legal target too.
	song2 := b724Verse(g, me, "Torch Song", b724TorchSong, 2)
	life := opp.Life
	if err := b724Activate(t, g, me, song2, "{C}{C}{R}", game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate at a player: %v", err)
	}
	b724Settle(t, g)
	if opp.Life != life-2 {
		t.Errorf("the opponent's life %d → %d, want -2", life, opp.Life)
	}
}

func TestB724MidsummerRevelMakesXBeasts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rev := b724Verse(g, me, "Midsummer Revel", b724MidsummerRevel, 3)
	if err := b724Activate(t, g, me, rev, "", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated without the {G}")
	}
	if err := b724Activate(t, g, me, rev, "{G}", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	b724Settle(t, g)
	n, _ := effectstest.CountTokensNamed(g, me.ID, "Beast")
	if n != 3 {
		t.Fatalf("%d Beast tokens, want 3", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Beast" && (c.Power != 3 || c.Toughness != 3 || !c.HasColor("G")) {
			t.Errorf("Beast token is %d/%d colors %v, want a green 3/3", c.Power, c.Toughness, c.Colors)
		}
	}
}

func TestB724WarDanceGivesPlusXPlusXUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dance := b724Verse(g, me, "War Dance", b724WarDance, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := b724Activate(t, g, me, dance, "", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated without a target")
	}
	if err := b724Activate(t, g, me, dance, "", game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	b724Settle(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 6 || tough != 6 {
		t.Errorf("the Bear is %d/%d, want 6/6", p, tough)
	}
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	b724Settle(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("after the turn the Bear is %d/%d, want 2/2", p, tough)
	}
}

func TestB724DiscordantDirgeChoosesUpToXCardsFromTheirHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dirge := b724Verse(g, me, "Discordant Dirge", b724DiscordantDirge, 2)
	opp.Hand.Cards = nil
	var cards []uuid.UUID
	for i := 0; i < 4; i++ {
		id := uuid.New()
		cards = append(cards, id)
		opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Card", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	}
	if err := b724Activate(t, g, me, dirge, "{B}", game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err == nil {
		t.Fatal("the controller was a legal target")
	}
	if err := b724Activate(t, g, me, dirge, "{B}", game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	b724Settle(t, g)
	pick := effectstest.OpenVariant(t, g)
	if pick.Count != 2 || !pick.PickUpTo {
		t.Fatalf("the pick is count %d up-to %v, want 2 and up to", pick.Count, pick.PickUpTo)
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, cards[:3]); err == nil {
		t.Error("three cards were accepted with two verse counters")
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, cards[:2]); err != nil {
		t.Fatalf("resolve the pick: %v", err)
	}
	b724Settle(t, g)
	if opp.Hand.Size() != 2 || !opp.Graveyard.Contains(cards[0]) || !opp.Graveyard.Contains(cards[1]) {
		t.Errorf("the opponent has %d cards in hand, want the two chosen discarded", opp.Hand.Size())
	}
}

func TestB724LiltingRefrainCountersUnlessTheCasterPaysX(t *testing.T) {
	for _, pays := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		ref := b724Verse(g, me, "Lilting Refrain", b724LiltingRefrain, 2)
		spell := uuid.New()
		g.Turn.ActiveSeat = 1
		opp.Hand.PushTop(game.Card{InstanceID: spell, Name: "Rock", TypeLine: "Artifact", ManaCost: "", Owner: opp.ID, Controller: opp.ID})
		effectstest.AdvanceToMainOf(t, g, 1)
		if pays {
			effectstest.AddPool(t, g, opp.ID, "{C}{C}")
		}
		if err := g.CastSpell(opp.ID, spell, game.CastSpellParams{HoldPriority: true}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		if err := b724Activate(t, g, me, ref, "", game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell}}}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if q := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePayUnless, opp.ID); q != nil {
			if err := g.ResolvePayUnless(q.ID, opp.ID, pays); err != nil {
				t.Fatalf("answer: %v", err)
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
		if pays && !g.Battlefield.Contains(spell) {
			t.Error("the spell was countered although its controller paid")
		}
		if !pays && g.Battlefield.Contains(spell) {
			t.Error("the spell resolved although its controller did not pay")
		}
	}
}

func TestB724TitaniasBoonPutsACounterOnEachCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Titania's Boon", "Sorcery", b724TitaniasBoon, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, a); p != 3 || tough != 3 {
		t.Errorf("A is %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, b); p != 2 || tough != 2 {
		t.Errorf("B is %d/%d, want 2/2", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("the opponent's creature is %d/%d, want 2/2", p, tough)
	}
}

func TestB724VebulidEntersWithACounterAndMayGrowEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	v := effectstest.CastWithCost(t, g, "Vebulid", "Creature — Phyrexian Horror", "{B}", b724Vebulid)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, v); p != 1 || tough != 1 {
		t.Fatalf("Vebulid enters as %d/%d, want 1/1 (one +1/+1 counter on a 0/0)", p, tough)
	}
	for _, yes := range []bool{true, false} {
		// Round the table to its controller's next upkeep.
		b724AdvanceToUpkeepOf(t, g, seat)
		before, _ := effectstest.PtOf(t, g, v)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		b724Settle(t, g)
		after, _ := effectstest.PtOf(t, g, v)
		want := before
		if yes {
			want = before + 1
		}
		if after != want {
			t.Errorf("yes=%v: power %d → %d, want %d", yes, before, after, want)
		}
	}
}

func TestB724VebulidIsDestroyedAtEndOfCombatAfterAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	v := effectstest.PushCatalogPermanent(g, me.ID, "Vebulid", "Creature — Phyrexian Horror", b724Vebulid, false)
	effectstest.AddCountersForTest(g, v, game.CounterPlusOne, 2)
	effectstest.DeclareAttack(t, g, opp.ID, v)
	b724Settle(t, g)
	if !g.Battlefield.Contains(v) {
		t.Fatal("Vebulid died before end of combat")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	b724Settle(t, g)
	if g.Battlefield.Contains(v) || !me.Graveyard.Contains(v) {
		t.Error("Vebulid survived the end of combat after attacking")
	}
}

func TestB724VebulidIsDestroyedAtEndOfCombatAfterBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 1, 1)
	v := effectstest.PushCatalogPermanent(g, opp.ID, "Vebulid", "Creature — Phyrexian Horror", b724Vebulid, false)
	effectstest.AddCountersForTest(g, v, game.CounterPlusOne, 5)
	effectstest.Pr7aBlockedCombat(t, g, opp.ID, attacker, v)
	b724Settle(t, g)
	if !g.Battlefield.Contains(v) {
		t.Fatal("Vebulid died before end of combat")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	b724Settle(t, g)
	if g.Battlefield.Contains(v) || !opp.Graveyard.Contains(v) {
		t.Error("Vebulid survived the end of combat after blocking")
	}
}

func TestB724VebulidStaysWhenItNeitherAttacksNorBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	v := effectstest.PushCatalogPermanent(g, me.ID, "Vebulid", "Creature — Phyrexian Horror", b724Vebulid, false)
	effectstest.AddCountersForTest(g, v, game.CounterPlusOne, 1)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	b724Settle(t, g)
	if !g.Battlefield.Contains(v) {
		t.Error("Vebulid was destroyed without attacking or blocking")
	}
}
