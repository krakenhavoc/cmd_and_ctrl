package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_neo_b04_cards_test.go — Kamigawa: Neon Dynasty, S82 batch 4 (#5493).
// Each test plays the real catalog entries through the engine.

const (
	neoB04Koi         = "53b9a2cf-2e6e-4ce0-bb9f-f33fb27b7b5e"
	neoB04TrashBot    = "93b0bd0e-414b-45bd-9d3d-7323140ef0c0"
	neoB04Replication = "7ce7decb-1438-4c3d-88d2-26b0b3996b33"
	neoB04Operative   = "9f53aac8-f21f-4928-af8f-65b40cb9149c"
	neoB04Kami        = "7baffe00-78a4-4675-9c74-29eb3882bd01"
	neoB04Gloom       = "b58737a4-180e-4e2f-95ae-afd6d446d608"
	neoB04Searchlight = "426569c7-d868-4af1-bc83-5b5200532d17"
	neoB04Steward     = "993a91fa-84ed-4688-9884-e03cbad6c791"
	neoB04Blade       = "7818e900-2f4e-4710-87a4-5a9ac170a677"
	neoB04Kunai       = "fa5aefaf-432f-41d0-8e54-0e8b7bc7ba98"
	neoB04Cudgels     = "82ebd4fa-9db1-4723-8305-8b8a9ce958e2"
	neoB04Eater       = "b6f8cc3b-b8f1-4b1a-8768-eddcb0631d80"
)

func TestNeoB04EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		neoB04Koi: "Skyswimmer Koi", neoB04TrashBot: "Runaway Trash-Bot",
		neoB04Replication: "Replication Specialist", neoB04Operative: "Futurist Operative",
		neoB04Kami: "Kami of Transience", neoB04Gloom: "Gloomshrieker",
		neoB04Searchlight: "Searchlight Companion", neoB04Steward: "Shrine Steward",
		neoB04Blade: "Enormous Energy Blade", neoB04Kunai: "Ninja's Kunai",
		neoB04Cudgels: "Bronze Cudgels", neoB04Eater: "Eater of Virtue",
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

// --- Skyswimmer Koi --------------------------------------------------------

func TestNeoB04KoiLootsWhenYourArtifactEntersButNotAnOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.Msc18Library(me, 8)
	effectstest.Cmdr2B05Card(g, me.ID, "Skyswimmer Koi", "Creature — Fish", neoB04Koi, 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d, want %d after the draw", me.Hand.Size(), hand+1)
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d after the discard, want %d", me.Hand.Size(), hand)
	}
	effectstest.PassPriorityAroundTable(t, g)

	// An artifact creature counts; a non-artifact does not.
	effectstest.CastCatalogSpell(t, g, "Gearbot", "Artifact Creature — Construct", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) == nil {
		t.Fatal("an artifact creature entering did not offer the loot")
	}
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Error("a nonartifact creature entering offered the loot")
	}

	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID,
	})
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, CardID: theirs, Actor: opp.ID}) })
	if len(g.PendingTriggers) != 0 {
		t.Errorf("an opponent's artifact queued %d trigger(s)", len(g.PendingTriggers))
	}
}

// --- Runaway Trash-Bot -----------------------------------------------------

func TestNeoB04TrashBotCountsArtifactAndOrEnchantmentCardsInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bot := effectstest.Cmdr2B05Card(g, me.ID, "Runaway Trash-Bot", "Artifact Creature — Construct", neoB04TrashBot, 0, 4)
	if p := effectstest.EffectivePower(t, g, bot); p != 0 {
		t.Fatalf("power %d with an empty graveyard, want 0", p)
	}
	toGraveyard := func(p *game.Player, name, typeLine string) {
		id := uuid.New()
		p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
		if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: p.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: p.ID}, id); err != nil {
			t.Fatalf("MoveCardByID: %v", err)
		}
	}
	toGraveyard(me, "Rock", "Artifact")
	toGraveyard(me, "Anthem", "Enchantment")
	toGraveyard(me, "Gearbot", "Artifact Creature — Construct")
	toGraveyard(me, "Shrine Thing", "Enchantment Creature — Spirit")
	toGraveyard(me, "Bear", "Creature — Bear")
	toGraveyard(me, "Bolt", "Instant")
	toGraveyard(opp, "Their Rock", "Artifact")
	if p := effectstest.EffectivePower(t, g, bot); p != 4 {
		t.Errorf("power %d, want 4 (four artifact and/or enchantment cards of mine)", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bot, "trample") {
		t.Error("the Trash-Bot has no trample")
	}
	if tgh := effectstest.EffectiveToughness(t, g, bot); tgh != 4 {
		t.Errorf("toughness %d, want 4", tgh)
	}
}

// --- Replication Specialist ------------------------------------------------

func TestNeoB04ReplicationSpecialistCopiesANontokenArtifactForOneBlue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.Cmdr2B05Card(g, me.ID, "Replication Specialist", "Creature — Moonfolk Artificer", neoB04Replication, 3, 4)
	effectstest.CastCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	effectstest.PassPriorityUntilChoice(t, g)
	g.WithWriteLock(func() { _ = g.AddManaForEffect(me.ID, uuid.Nil, "{C}{U}") })
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	rocks := effectstest.BattlefieldIDsNamed(g, "Rock")
	if len(rocks) != 2 {
		t.Fatalf("%d Rocks, want the artifact and its copy", len(rocks))
	}
	tokens := 0
	for _, id := range rocks {
		if c, _ := effectstest.BattlefieldCard(g, id); c.IsToken() && c.IsArtifact() {
			tokens++
		}
	}
	if tokens != 1 {
		t.Errorf("%d token Rocks, want 1", tokens)
	}
	// The token entering is a token artifact: it must not offer another copy.
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("the token copy triggered a second copy")
	}
}

func TestNeoB04ReplicationSpecialistDeclinedOrIgnoresOtherPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.Cmdr2B05Card(g, me.ID, "Replication Specialist", "Creature — Moonfolk Artificer", neoB04Replication, 3, 4)
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("a nonartifact creature offered a copy")
	}
	effectstest.CastCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	effectstest.PassPriorityUntilChoice(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.BattlefieldIDsNamed(g, "Rock")); n != 1 {
		t.Errorf("%d Rocks after declining, want 1", n)
	}
}

// --- Futurist Operative ----------------------------------------------------

func TestNeoB04OperativeIsAOneOneCitizenWhileTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	op := effectstest.Cmdr2B05Card(g, me.ID, "Futurist Operative", "Creature — Human Ninja", neoB04Operative, 3, 4)
	if p, tt := effectstest.EffectivePower(t, g, op), effectstest.EffectiveToughness(t, g, op); p != 3 || tt != 4 {
		t.Fatalf("untapped %d/%d, want 3/4", p, tt)
	}
	if c := effectstest.EffectiveCardForTest(t, g, op); !c.HasSubtype("Ninja") || c.HasSubtype("Citizen") {
		t.Errorf("untapped type line %q, want Human Ninja", c.TypeLine)
	}
	effectstest.TapForTest(t, g, op)
	if p, tt := effectstest.EffectivePower(t, g, op), effectstest.EffectiveToughness(t, g, op); p != 1 || tt != 1 {
		t.Errorf("tapped %d/%d, want 1/1", p, tt)
	}
	c := effectstest.EffectiveCardForTest(t, g, op)
	if !c.HasSubtype("Human") || !c.HasSubtype("Citizen") || c.HasSubtype("Ninja") || !c.IsCreature() {
		t.Errorf("tapped type line %q creature=%v, want Human Citizen", c.TypeLine, c.IsCreature())
	}
	// A +1/+1 counter still applies on top of the base 1/1.
	effectstest.AddCountersForTest(g, op, game.CounterPlusOne, 1)
	if p := effectstest.EffectivePowerWithCounters(t, g, op); p != 2 {
		t.Errorf("tapped power %d with a counter, want 2", p)
	}
}

func TestNeoB04OperativeCantBeBlockedWhileTappedButCanBeWhenUntapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	op := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Futurist Operative", TypeLine: "Creature — Human Ninja", OracleID: neoB04Operative,
		Power: 3, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	vig := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Vigilant Operative", TypeLine: "Creature — Human Ninja", OracleID: neoB04Operative,
		Power: 3, Toughness: 4, Owner: me.ID, Controller: me.ID, Keywords: []string{"vigilance"},
	})
	effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.BrAttack(t, g, op, vig)
	for _, set := range effectstest.BrOffered(t, g, opp.ID) {
		for _, b := range set {
			if b.Attacker == op {
				t.Fatalf("the tapped, attacking Operative was offered a blocker: %+v", set)
			}
		}
	}
	// The vigilant attacker never tapped, so it is a plain 3/4 and blockable.
	blockable := false
	for _, set := range effectstest.BrOffered(t, g, opp.ID) {
		for _, b := range set {
			if b.Attacker == vig {
				blockable = true
			}
		}
	}
	if !blockable {
		t.Error("an untapped attacking Operative could not be blocked")
	}
}

func TestNeoB04OperativeUntapsForTwoBlue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	op := effectstest.Cmdr2B05Card(g, me.ID, "Futurist Operative", "Creature — Human Ninja", neoB04Operative, 3, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.TapForTest(t, g, op)
	if err := g.ActivateCatalogAbility(me.ID, op, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("untapped with no mana")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{U}")
	effectstest.B16Activate(t, g, me.ID, op, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, op); tapped {
		t.Error("the Operative is still tapped")
	}
	if p := effectstest.EffectivePower(t, g, op); p != 3 {
		t.Errorf("power %d after untapping, want 3", p)
	}
}

// --- Kami of Transience ----------------------------------------------------

func TestNeoB04KamiGrowsOnEachEnchantmentSpellOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	kami := effectstest.Cmdr2B05Card(g, me.ID, "Kami of Transience", "Creature — Spirit", neoB04Kami, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Anthem", "Enchantment", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Aura Thing", "Enchantment — Aura", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, kami, game.CounterPlusOne); n != 2 {
		t.Fatalf("%d +1/+1 counters, want 2 (only the enchantment spells)", n)
	}
	if !effectstest.HasEffectiveKeyword(t, g, kami, "trample") {
		t.Error("no trample")
	}
}

// neoB04PassToConfirm passes priority until the player has a yes/no prompt.
func neoB04PassToConfirm(t *testing.T, g *game.Game, who uuid.UUID) {
	t.Helper()
	for i := 0; i < 12 && neoB04Prompt(g, who) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
}

// neoB04Prompt is the open "you may" trigger prompt for who, if any.
func neoB04Prompt(g *game.Game, who uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == who {
			return c
		}
	}
	return nil
}

func neoB04AnswerPrompt(t *testing.T, g *game.Game, who uuid.UUID, yes bool) {
	t.Helper()
	c := neoB04Prompt(g, who)
	if c == nil {
		t.Fatal("no trigger prompt open")
	}
	if err := g.ResolveTriggerPrompt(c.ID, who, yes); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
}

// kamiInGraveyard puts a Kami of Transience into me's graveyard.
func neoB04KamiInGraveyard(me *game.Player) uuid.UUID {
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: "Kami of Transience", TypeLine: "Creature — Spirit", OracleID: neoB04Kami,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, Keywords: []string{"trample"},
	})
	return id
}

func TestNeoB04KamiReturnsAtEndStepAfterAnEnchantmentWentToYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	kami := neoB04KamiInGraveyard(me)
	ench := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Anthem", TypeLine: "Enchantment", Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.KillCreature(t, g, me.ID, ench) // moves the permanent to its owner's graveyard
	effectstest.AdvanceTo(t, g, game.StepEnd)
	neoB04PassToConfirm(t, g, me.ID)
	neoB04AnswerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.HandCardByID(g, me, kami); !ok {
		t.Fatal("the Kami did not return to hand")
	}
	if me.Graveyard.Contains(kami) {
		t.Error("the Kami is still in the graveyard")
	}
}

func TestNeoB04KamiMayDeclineAndNeedsAnEnchantmentFromTheBattlefield(t *testing.T) {
	// Declining leaves it.
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	kami := neoB04KamiInGraveyard(me)
	ench := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Anthem", TypeLine: "Enchantment", Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.KillCreature(t, g, me.ID, ench)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	neoB04PassToConfirm(t, g, me.ID)
	neoB04AnswerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(kami) {
		t.Error("declining still returned the Kami")
	}

	// An enchantment that did not come from the battlefield, a nonenchantment
	// permanent, and an opponent's graveyard do not count.
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	kami2 := neoB04KamiInGraveyard(me2)
	effectstest.PushGraveyardPermanent(me2, "Old Anthem", "Enchantment", "{1}{W}")
	rock := effectstest.PushBattlefieldCardWithTimestamp(g2, game.Card{
		InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: me2.ID, Controller: me2.ID,
	})
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g2, game.Card{
		InstanceID: uuid.New(), Name: "Their Anthem", TypeLine: "Enchantment", Owner: opp2.ID, Controller: opp2.ID,
	})
	effectstest.AdvanceToMain(t, g2)
	effectstest.KillCreature(t, g2, me2.ID, rock)
	effectstest.KillCreature(t, g2, opp2.ID, theirs)
	effectstest.AdvanceTo(t, g2, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g2)
	if neoB04Prompt(g2, me2.ID) != nil {
		t.Error("the Kami offered to return without an enchantment put into my graveyard from the battlefield")
	}
	if !me2.Graveyard.Contains(kami2) {
		t.Error("the Kami left the graveyard")
	}
}

func TestNeoB04KamiReturnsOnAnOpponentsEndStepToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	kami := neoB04KamiInGraveyard(me)
	ench := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Anthem", TypeLine: "Enchantment", Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMainOf(t, g, 1) // the opponent's turn
	effectstest.KillCreature(t, g, me.ID, ench)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	neoB04PassToConfirm(t, g, me.ID)
	neoB04AnswerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.HandCardByID(g, me, kami); !ok {
		t.Fatal("each end step: the Kami did not return on an opponent's turn")
	}
}

// --- Gloomshrieker ---------------------------------------------------------

func TestNeoB04GloomshriekerReturnsAPermanentCardAndExilesInsteadOfDying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := effectstest.PushGraveyardPermanent(me, "Rock", "Artifact", "{1}")
	bolt := effectstest.PushGraveyardPermanent(me, "Bolt", "Instant", "{R}")
	effectstest.CastCatalogSpell(t, g, "Gloomshrieker", "Enchantment Creature — Cat Beast", neoB04Gloom, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickIfAsked(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.HandCardByID(g, me, rock); !ok {
		t.Fatalf("the permanent card did not return to hand; hand %v", me.Hand.Cards)
	}
	if !me.Graveyard.Contains(bolt) {
		t.Error("the instant card was returned")
	}
	id := effectstest.FindBattlefieldByName(g, "Gloomshrieker")
	if id == uuid.Nil {
		t.Fatal("Gloomshrieker is not on the battlefield")
	}
	effectstest.KillCreature(t, g, me.ID, id)
	if g.Exile == nil || !g.Exile.Contains(id) {
		t.Errorf("Gloomshrieker did not go to exile: graveyard=%v", me.Graveyard.Contains(id))
	}
	if me.Graveyard.Contains(id) {
		t.Error("Gloomshrieker reached the graveyard")
	}
}

// --- Searchlight Companion -------------------------------------------------

func TestNeoB04SearchlightCompanionMakesASpirit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Searchlight Companion", "Artifact Creature — Drone", neoB04Searchlight, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Spirit"); n != 1 {
		t.Fatalf("%d Spirits, want 1", n)
	}
	spirit, _ := effectstest.BattlefieldCard(g, effectstest.FindBattlefieldByName(g, "Spirit"))
	if !spirit.IsToken() || !spirit.IsCreature() || len(spirit.Colors) != 0 || spirit.Power != 1 || spirit.Toughness != 1 {
		t.Errorf("the Spirit is %+v, want a 1/1 colorless Spirit creature token", spirit)
	}
}

// --- Shrine Steward --------------------------------------------------------

func TestNeoB04ShrineStewardFindsAnAuraOrAShrine(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Pacifism Test", TypeLine: "Enchantment — Aura"},
		game.Card{Name: "Shrine Test", TypeLine: "Legendary Enchantment — Shrine"},
		game.Card{Name: "Plain Enchantment", TypeLine: "Enchantment"},
		game.Card{Name: "Rock", TypeLine: "Artifact"},
	)
	effectstest.CastAndResolveCreature(t, g, "Shrine Steward", "Artifact Creature — Construct", neoB04Steward)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt — 'you may search' must still ask")
	}
	if effectstest.SearchOptionNamed(g, c, "Plain Enchantment") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Rock") != uuid.Nil {
		t.Error("a card that is neither an Aura nor a Shrine was offered")
	}
	if effectstest.SearchOptionNamed(g, c, "Pacifism Test") == uuid.Nil || effectstest.SearchOptionNamed(g, c, "Shrine Test") == uuid.Nil {
		t.Error("an Aura or a Shrine was not offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Shrine Test")
	if !effectstest.B02bHandHasNamed(me, "Shrine Test") {
		t.Error("the Shrine did not reach the hand")
	}
}

// --- Enormous Energy Blade -------------------------------------------------

func TestNeoB04BladeGivesPlusFourAndTapsTheCreatureItBecomesAttachedTo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	other := effectstest.B16Creature(g, me.ID, "Other", "Creature — Bear", 2, 2, "G")
	blade := effectstest.B21Push(g, me.ID, "Enormous Energy Blade", "Artifact — Equipment", neoB04Blade, 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.EquipTo(t, g, me.ID, blade, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 6 {
		t.Errorf("equipped power %d, want 6", p)
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, bear); !tapped {
		t.Error("the creature the Blade became attached to is untapped")
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, other); tapped {
		t.Error("an unrelated creature is tapped")
	}
	// Moving it taps the new host, not the old one's state.
	effectstest.TapForTest(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.EquipTo(t, g, me.ID, blade, other)
	effectstest.PassPriorityAroundTable(t, g)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, other); !tapped {
		t.Error("the new host was not tapped when the Blade moved onto it")
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("the old host keeps %d power, want 2", p)
	}
}

// --- Ninja's Kunai ---------------------------------------------------------

func TestNeoB04KunaiSacrificesItselfToDealThreeDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	kunai := effectstest.B21Push(g, me.ID, "Ninja's Kunai", "Artifact — Equipment", neoB04Kunai, 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.EquipTo(t, g, me.ID, kunai, bear)
	idx, ref := effectstest.GrantedActivatedIndex(t, g, bear)
	if idx < 0 {
		t.Fatal("the equipped creature has no granted ability")
	}
	params := game.ActivateAbilityParams{Ref: ref, Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}
	if err := g.ActivateCatalogAbility(me.ID, bear, idx, params); err == nil {
		t.Fatal("activated with no mana")
	}
	life := opp.Life
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.B16Activate(t, g, me.ID, bear, idx, params)
	if g.Battlefield.Contains(kunai) {
		t.Error("the Kunai was not sacrificed as a cost")
	}
	if !effectstest.B16Tapped(t, g, bear) {
		t.Error("the creature did not tap to pay")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d → %d, want -3", life, opp.Life)
	}
}

func TestNeoB04KunaiKillsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	victim := effectstest.B16Creature(g, opp.ID, "Victim", "Creature — Bear", 3, 3, "G")
	kunai := effectstest.B21Push(g, me.ID, "Ninja's Kunai", "Artifact — Equipment", neoB04Kunai, 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.EquipTo(t, g, me.ID, kunai, bear)
	idx, ref := effectstest.GrantedActivatedIndex(t, g, bear)
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.B16Activate(t, g, me.ID, bear, idx, game.ActivateAbilityParams{Ref: ref, Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the 3/3 survived 3 damage")
	}
}

// --- Bronze Cudgels --------------------------------------------------------

func TestNeoB04CudgelsGrowsByTheTimesItHasResolvedThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	cudgels := effectstest.B21Push(g, me.ID, "Bronze Cudgels", "Artifact — Equipment", neoB04Cudgels, 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.EquipTo(t, g, me.ID, cudgels, bear)
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Fatalf("power %d after equipping, want 2", p)
	}
	for i, want := range []int{3, 5, 8} { // +1, +2, +3
		effectstest.FloatMana(t, g, me, "{C}{C}")
		effectstest.B16Activate(t, g, me.ID, cudgels, 0, game.ActivateAbilityParams{Strict: true})
		effectstest.PassPriorityAroundTable(t, g)
		if p := effectstest.EffectivePower(t, g, bear); p != want {
			t.Fatalf("after resolution %d power is %d, want %d", i+1, p, want)
		}
	}
	// The boost wears off at end of turn.
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToMainOf(t, g, 1)
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("power %d next turn, want 2", p)
	}
}

func TestNeoB04CudgelsUnequippedDoesNothingAndNeedsMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cudgels := effectstest.B21Push(g, me.ID, "Bronze Cudgels", "Artifact — Equipment", neoB04Cudgels, 0, 0)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, cudgels, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no mana")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.B16Activate(t, g, me.ID, cudgels, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	// It resolved with nothing equipped; equipping now and resolving once more is the second.
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.EquipTo(t, g, me.ID, cudgels, bear)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.B16Activate(t, g, me.ID, cudgels, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 4 {
		t.Errorf("power %d, want 4 (+2 on the second resolution)", p)
	}
}

// --- Eater of Virtue -------------------------------------------------------

func TestNeoB04EaterExilesTheDeadCreatureAndGrantsItsKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	flyer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flyer", TypeLine: "Creature — Bird", Power: 1, Toughness: 1,
		Owner: me.ID, Controller: me.ID, Keywords: []string{"flying", "lifelink"},
	})
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	eater := effectstest.B21Push(g, me.ID, "Eater of Virtue", "Legendary Artifact — Equipment", neoB04Eater, 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.EquipTo(t, g, me.ID, eater, flyer)
	if p := effectstest.EffectivePower(t, g, flyer); p != 3 {
		t.Fatalf("equipped power %d, want 3", p)
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Fatal("the unequipped bear flies")
	}
	effectstest.KillCreature(t, g, me.ID, flyer)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile == nil || !g.Exile.Contains(flyer) {
		t.Fatalf("the dead creature was not exiled (graveyard=%v)", me.Graveyard.Contains(flyer))
	}
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.EquipTo(t, g, me.ID, eater, bear)
	for _, kw := range []string{"flying", "lifelink"} {
		if !effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the equipped bear lacks %s from the exiled card", kw)
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the bear has a keyword no exiled card has")
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 4 {
		t.Errorf("power %d, want 4", p)
	}
	// A card that leaves exile stops counting.
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneExile}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, flyer); err != nil {
		t.Fatalf("move: %v", err)
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the bear still flies after the exiled card left exile")
	}
}

func TestNeoB04EaterIgnoresACreatureThatDiesWhileNotEquipped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	other := effectstest.B16Creature(g, me.ID, "Other", "Creature — Bear", 2, 2, "G")
	eater := effectstest.B21Push(g, me.ID, "Eater of Virtue", "Legendary Artifact — Equipment", neoB04Eater, 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.EquipTo(t, g, me.ID, eater, bear)
	effectstest.KillCreature(t, g, me.ID, other)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(other) {
		t.Error("an unequipped creature was exiled")
	}
}
