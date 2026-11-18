package effects

import (
	"testing"

	"github.com/google/uuid"

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

// --- Skyswimmer Koi --------------------------------------------------------

func TestNeoB04KoiLootsWhenYourArtifactEntersButNotAnOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	msc18Library(me, 8)
	cmdr2B05Card(g, me.ID, "Skyswimmer Koi", "Creature — Fish", neoB04Koi, 3, 3)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	answerMayChoice(t, g, me.ID, true)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d, want %d after the draw", me.Hand.Size(), hand+1)
	}
	discardFromHand(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d after the discard, want %d", me.Hand.Size(), hand)
	}
	passPriorityAroundTable(t, g)

	// An artifact creature counts; a non-artifact does not.
	castCatalogSpell(t, g, "Gearbot", "Artifact Creature — Construct", "", nil)
	passPriorityAroundTable(t, g)
	if latestConfirmFor(g, me.ID) == nil {
		t.Fatal("an artifact creature entering did not offer the loot")
	}
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if latestConfirmFor(g, me.ID) != nil {
		t.Error("a nonartifact creature entering offered the loot")
	}

	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID,
	})
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, CardID: theirs, Actor: opp.ID}) })
	if len(g.PendingTriggers) != 0 {
		t.Errorf("an opponent's artifact queued %d trigger(s)", len(g.PendingTriggers))
	}
}

// --- Runaway Trash-Bot -----------------------------------------------------

func TestNeoB04TrashBotCountsArtifactAndOrEnchantmentCardsInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bot := cmdr2B05Card(g, me.ID, "Runaway Trash-Bot", "Artifact Creature — Construct", neoB04TrashBot, 0, 4)
	if p := effectivePower(t, g, bot); p != 0 {
		t.Fatalf("power %d with an empty graveyard, want 0", p)
	}
	pushGraveyardPermanent(me, "Rock", "Artifact", "{1}")
	pushGraveyardPermanent(me, "Anthem", "Enchantment", "{1}{W}")
	pushGraveyardPermanent(me, "Gearbot", "Artifact Creature — Construct", "{2}")
	pushGraveyardPermanent(me, "Shrine Thing", "Enchantment Creature — Spirit", "{2}")
	pushGraveyardPermanent(me, "Bear", "Creature — Bear", "{1}{G}")
	pushGraveyardPermanent(me, "Bolt", "Instant", "{R}")
	pushGraveyardPermanent(opp, "Their Rock", "Artifact", "{1}")
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if p := effectivePower(t, g, bot); p != 4 {
		t.Errorf("power %d, want 4 (four artifact and/or enchantment cards of mine)", p)
	}
	if !hasEffectiveKeyword(t, g, bot, "trample") {
		t.Error("the Trash-Bot has no trample")
	}
	if tgh := effectiveToughness(t, g, bot); tgh != 4 {
		t.Errorf("toughness %d, want 4", tgh)
	}
}

// --- Replication Specialist ------------------------------------------------

func TestNeoB04ReplicationSpecialistCopiesANontokenArtifactForOneBlue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cmdr2B05Card(g, me.ID, "Replication Specialist", "Creature — Moonfolk Artificer", neoB04Replication, 3, 4)
	castCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	passPriorityUntilChoice(t, g)
	g.WithWriteLock(func() { _ = g.AddManaForEffect(me.ID, uuid.Nil, "{C}{U}") })
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	rocks := battlefieldIDsNamed(g, "Rock")
	if len(rocks) != 2 {
		t.Fatalf("%d Rocks, want the artifact and its copy", len(rocks))
	}
	tokens := 0
	for _, id := range rocks {
		if c, _ := battlefieldCard(g, id); c.IsToken() && c.IsArtifact() {
			tokens++
		}
	}
	if tokens != 1 {
		t.Errorf("%d token Rocks, want 1", tokens)
	}
	// The token entering is a token artifact: it must not offer another copy.
	if hasPayUnlessFor(g, me.ID) {
		t.Error("the token copy triggered a second copy")
	}
}

func TestNeoB04ReplicationSpecialistDeclinedOrIgnoresOtherPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cmdr2B05Card(g, me.ID, "Replication Specialist", "Creature — Moonfolk Artificer", neoB04Replication, 3, 4)
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Fatal("a nonartifact creature offered a copy")
	}
	castCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	passPriorityUntilChoice(t, g)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if n := len(battlefieldIDsNamed(g, "Rock")); n != 1 {
		t.Errorf("%d Rocks after declining, want 1", n)
	}
}

// --- Futurist Operative ----------------------------------------------------

func TestNeoB04OperativeIsAOneOneCitizenWhileTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	op := cmdr2B05Card(g, me.ID, "Futurist Operative", "Creature — Human Ninja", neoB04Operative, 3, 4)
	if p, tt := effectivePower(t, g, op), effectiveToughness(t, g, op); p != 3 || tt != 4 {
		t.Fatalf("untapped %d/%d, want 3/4", p, tt)
	}
	if c := effectiveCardForTest(t, g, op); !c.HasSubtype("Ninja") || c.HasSubtype("Citizen") {
		t.Errorf("untapped type line %q, want Human Ninja", c.TypeLine)
	}
	tapForTest(t, g, op)
	if p, tt := effectivePower(t, g, op), effectiveToughness(t, g, op); p != 1 || tt != 1 {
		t.Errorf("tapped %d/%d, want 1/1", p, tt)
	}
	c := effectiveCardForTest(t, g, op)
	if !c.HasSubtype("Human") || !c.HasSubtype("Citizen") || c.HasSubtype("Ninja") || !c.IsCreature() {
		t.Errorf("tapped type line %q creature=%v, want Human Citizen", c.TypeLine, c.IsCreature())
	}
	// A +1/+1 counter still applies on top of the base 1/1.
	addCountersForTest(g, op, game.CounterPlusOne, 1)
	if p := effectivePower(t, g, op); p != 2 {
		t.Errorf("tapped power %d with a counter, want 2", p)
	}
}

func TestNeoB04OperativeCantBeBlockedWhileTappedButCanBeWhenUntapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	op := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Futurist Operative", TypeLine: "Creature — Human Ninja", OracleID: neoB04Operative,
		Power: 3, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	vig := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Vigilant Operative", TypeLine: "Creature — Human Ninja", OracleID: neoB04Operative,
		Power: 3, Toughness: 4, Owner: me.ID, Controller: me.ID, Keywords: []string{"vigilance"},
	})
	b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	brAttack(t, g, op, vig)
	for _, set := range brOffered(t, g, opp.ID) {
		for _, b := range set {
			if b.Attacker == op {
				t.Fatalf("the tapped, attacking Operative was offered a blocker: %+v", set)
			}
		}
	}
	// The vigilant attacker never tapped, so it is a plain 3/4 and blockable.
	blockable := false
	for _, set := range brOffered(t, g, opp.ID) {
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	op := cmdr2B05Card(g, me.ID, "Futurist Operative", "Creature — Human Ninja", neoB04Operative, 3, 4)
	advanceToMain(t, g)
	tapForTest(t, g, op)
	if err := g.ActivateCatalogAbility(me.ID, op, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("untapped with no mana")
	}
	floatMana(t, g, me, "{C}{C}{U}")
	b16Activate(t, g, me.ID, op, 0, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	if tapped, _ := battlefieldCardTapped(g, op); tapped {
		t.Error("the Operative is still tapped")
	}
	if p := effectivePower(t, g, op); p != 3 {
		t.Errorf("power %d after untapping, want 3", p)
	}
}

// --- Kami of Transience ----------------------------------------------------

func TestNeoB04KamiGrowsOnEachEnchantmentSpellOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	kami := cmdr2B05Card(g, me.ID, "Kami of Transience", "Creature — Spirit", neoB04Kami, 2, 2)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Anthem", "Enchantment", "", nil)
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Aura Thing", "Enchantment — Aura", "", nil)
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, kami, game.CounterPlusOne); n != 2 {
		t.Fatalf("%d +1/+1 counters, want 2 (only the enchantment spells)", n)
	}
	if !hasEffectiveKeyword(t, g, kami, "trample") {
		t.Error("no trample")
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	kami := neoB04KamiInGraveyard(me)
	ench := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Anthem", TypeLine: "Enchantment", Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	killCreature(t, g, me.ID, ench) // moves the permanent to its owner's graveyard
	advanceTo(t, g, game.StepEnd)
	passPriorityUntilChoice(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if _, ok := handCardByID(g, me, kami); !ok {
		t.Fatal("the Kami did not return to hand")
	}
	if me.Graveyard.Contains(kami) {
		t.Error("the Kami is still in the graveyard")
	}
}

func TestNeoB04KamiMayDeclineAndNeedsAnEnchantmentFromTheBattlefield(t *testing.T) {
	// Declining leaves it.
	g := newCatalogGame(t)
	me := g.Seats[0]
	kami := neoB04KamiInGraveyard(me)
	ench := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Anthem", TypeLine: "Enchantment", Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	killCreature(t, g, me.ID, ench)
	advanceTo(t, g, game.StepEnd)
	passPriorityUntilChoice(t, g)
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(kami) {
		t.Error("declining still returned the Kami")
	}

	// An enchantment that did not come from the battlefield, a nonenchantment
	// permanent, and an opponent's graveyard do not count.
	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	kami2 := neoB04KamiInGraveyard(me2)
	pushGraveyardPermanent(me2, "Old Anthem", "Enchantment", "{1}{W}")
	rock := pushBattlefieldCardWithTimestamp(g2, game.Card{
		InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: me2.ID, Controller: me2.ID,
	})
	theirs := pushBattlefieldCardWithTimestamp(g2, game.Card{
		InstanceID: uuid.New(), Name: "Their Anthem", TypeLine: "Enchantment", Owner: opp2.ID, Controller: opp2.ID,
	})
	advanceToMain(t, g2)
	killCreature(t, g2, me2.ID, rock)
	killCreature(t, g2, opp2.ID, theirs)
	advanceTo(t, g2, game.StepEnd)
	passPriorityAroundTable(t, g2)
	if latestConfirmFor(g2, me2.ID) != nil {
		t.Error("the Kami offered to return without an enchantment put into my graveyard from the battlefield")
	}
	if !me2.Graveyard.Contains(kami2) {
		t.Error("the Kami left the graveyard")
	}
}

func TestNeoB04KamiReturnsOnAnOpponentsEndStepToo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	kami := neoB04KamiInGraveyard(me)
	ench := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Anthem", TypeLine: "Enchantment", Owner: me.ID, Controller: me.ID,
	})
	advanceToMainOf(t, g, 1) // the opponent's turn
	killCreature(t, g, me.ID, ench)
	advanceTo(t, g, game.StepEnd)
	passPriorityUntilChoice(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if _, ok := handCardByID(g, me, kami); !ok {
		t.Fatal("each end step: the Kami did not return on an opponent's turn")
	}
}

// --- Gloomshrieker ---------------------------------------------------------

func TestNeoB04GloomshriekerReturnsAPermanentCardAndExilesInsteadOfDying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rock := pushGraveyardPermanent(me, "Rock", "Artifact", "{1}")
	bolt := pushGraveyardPermanent(me, "Bolt", "Instant", "{R}")
	castCatalogSpell(t, g, "Gloomshrieker", "Enchantment Creature — Cat Beast", neoB04Gloom, nil)
	passPriorityAroundTable(t, g)
	if _, ok := handCardByID(g, me, rock); !ok {
		t.Fatalf("the permanent card did not return to hand; hand %v", me.Hand.Cards)
	}
	if !me.Graveyard.Contains(bolt) {
		t.Error("the instant card was returned")
	}
	id := findBattlefieldByName(g, "Gloomshrieker")
	if id == uuid.Nil {
		t.Fatal("Gloomshrieker is not on the battlefield")
	}
	killCreature(t, g, me.ID, id)
	if g.Exile == nil || !g.Exile.Contains(id) {
		t.Errorf("Gloomshrieker did not go to exile: graveyard=%v", me.Graveyard.Contains(id))
	}
	if me.Graveyard.Contains(id) {
		t.Error("Gloomshrieker reached the graveyard")
	}
}

// --- Searchlight Companion -------------------------------------------------

func TestNeoB04SearchlightCompanionMakesASpirit(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Searchlight Companion", "Artifact Creature — Drone", neoB04Searchlight, nil)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Spirit"); n != 1 {
		t.Fatalf("%d Spirits, want 1", n)
	}
	spirit, _ := battlefieldCard(g, findBattlefieldByName(g, "Spirit"))
	if !spirit.IsToken() || !spirit.IsCreature() || len(spirit.Colors) != 0 || spirit.Power != 1 || spirit.Toughness != 1 {
		t.Errorf("the Spirit is %+v, want a 1/1 colorless Spirit creature token", spirit)
	}
}

// --- Shrine Steward --------------------------------------------------------

func TestNeoB04ShrineStewardFindsAnAuraOrAShrine(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedSearchLibrary(me,
		game.Card{Name: "Pacifism Test", TypeLine: "Enchantment — Aura"},
		game.Card{Name: "Shrine Test", TypeLine: "Legendary Enchantment — Shrine"},
		game.Card{Name: "Plain Enchantment", TypeLine: "Enchantment"},
		game.Card{Name: "Rock", TypeLine: "Artifact"},
	)
	castAndResolveCreature(t, g, "Shrine Steward", "Artifact Creature — Construct", neoB04Steward)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt — 'you may search' must still ask")
	}
	if searchOptionNamed(g, c, "Plain Enchantment") != uuid.Nil || searchOptionNamed(g, c, "Rock") != uuid.Nil {
		t.Error("a card that is neither an Aura nor a Shrine was offered")
	}
	if searchOptionNamed(g, c, "Pacifism Test") == uuid.Nil || searchOptionNamed(g, c, "Shrine Test") == uuid.Nil {
		t.Error("an Aura or a Shrine was not offered")
	}
	answerSearchNamed(t, g, me.ID, "Shrine Test")
	if !b02bHandHasNamed(me, "Shrine Test") {
		t.Error("the Shrine did not reach the hand")
	}
}

// --- Enormous Energy Blade -------------------------------------------------

func TestNeoB04BladeGivesPlusFourAndTapsTheCreatureItBecomesAttachedTo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	other := b16Creature(g, me.ID, "Other", "Creature — Bear", 2, 2, "G")
	blade := b21Push(g, me.ID, "Enormous Energy Blade", "Artifact — Equipment", neoB04Blade, 0, 0)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}")
	equipTo(t, g, me.ID, blade, bear)
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, bear); p != 6 {
		t.Errorf("equipped power %d, want 6", p)
	}
	if tapped, _ := battlefieldCardTapped(g, bear); !tapped {
		t.Error("the creature the Blade became attached to is untapped")
	}
	if tapped, _ := battlefieldCardTapped(g, other); tapped {
		t.Error("an unrelated creature is tapped")
	}
	// Moving it taps the new host, not the old one's state.
	tapForTest(t, g)
	floatMana(t, g, me, "{C}{C}")
	equipTo(t, g, me.ID, blade, other)
	passPriorityAroundTable(t, g)
	if tapped, _ := battlefieldCardTapped(g, other); !tapped {
		t.Error("the new host was not tapped when the Blade moved onto it")
	}
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("the old host keeps %d power, want 2", p)
	}
}

// --- Ninja's Kunai ---------------------------------------------------------

func TestNeoB04KunaiSacrificesItselfToDealThreeDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	kunai := b21Push(g, me.ID, "Ninja's Kunai", "Artifact — Equipment", neoB04Kunai, 0, 0)
	advanceToMain(t, g)
	equipTo(t, g, me.ID, kunai, bear)
	idx, ref := grantedActivatedIndex(t, g, bear)
	if idx < 0 {
		t.Fatal("the equipped creature has no granted ability")
	}
	params := game.ActivateAbilityParams{Ref: ref, Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}
	if err := g.ActivateCatalogAbility(me.ID, bear, idx, params); err == nil {
		t.Fatal("activated with no mana")
	}
	life := opp.Life
	floatMana(t, g, me, "{C}")
	b16Activate(t, g, me.ID, bear, idx, params)
	if g.Battlefield.Contains(kunai) {
		t.Error("the Kunai was not sacrificed as a cost")
	}
	if !b16Tapped(t, g, bear) {
		t.Error("the creature did not tap to pay")
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d → %d, want -3", life, opp.Life)
	}
}

func TestNeoB04KunaiKillsACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	victim := b16Creature(g, opp.ID, "Victim", "Creature — Bear", 3, 3, "G")
	kunai := b21Push(g, me.ID, "Ninja's Kunai", "Artifact — Equipment", neoB04Kunai, 0, 0)
	advanceToMain(t, g)
	equipTo(t, g, me.ID, kunai, bear)
	idx, ref := grantedActivatedIndex(t, g, bear)
	floatMana(t, g, me, "{C}")
	b16Activate(t, g, me.ID, bear, idx, game.ActivateAbilityParams{Ref: ref, Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the 3/3 survived 3 damage")
	}
}

// --- Bronze Cudgels --------------------------------------------------------

func TestNeoB04CudgelsGrowsByTheTimesItHasResolvedThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	cudgels := b21Push(g, me.ID, "Bronze Cudgels", "Artifact — Equipment", neoB04Cudgels, 0, 0)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}")
	equipTo(t, g, me.ID, cudgels, bear)
	if p := effectivePower(t, g, bear); p != 2 {
		t.Fatalf("power %d after equipping, want 2", p)
	}
	for i, want := range []int{3, 5, 8} { // +1, +2, +3
		floatMana(t, g, me, "{C}{C}")
		b16Activate(t, g, me.ID, cudgels, 0, game.ActivateAbilityParams{Strict: true})
		passPriorityAroundTable(t, g)
		if p := effectivePower(t, g, bear); p != want {
			t.Fatalf("after resolution %d power is %d, want %d", i+1, p, want)
		}
	}
	// The boost wears off at end of turn.
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	advanceToMainOf(t, g, 1)
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("power %d next turn, want 2", p)
	}
}

func TestNeoB04CudgelsUnequippedDoesNothingAndNeedsMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cudgels := b21Push(g, me.ID, "Bronze Cudgels", "Artifact — Equipment", neoB04Cudgels, 0, 0)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, cudgels, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no mana")
	}
	floatMana(t, g, me, "{C}{C}")
	b16Activate(t, g, me.ID, cudgels, 0, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	// It resolved with nothing equipped; equipping now and resolving once more is the second.
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	floatMana(t, g, me, "{C}")
	equipTo(t, g, me.ID, cudgels, bear)
	floatMana(t, g, me, "{C}{C}")
	b16Activate(t, g, me.ID, cudgels, 0, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, bear); p != 4 {
		t.Errorf("power %d, want 4 (+2 on the second resolution)", p)
	}
}

// --- Eater of Virtue -------------------------------------------------------

func TestNeoB04EaterExilesTheDeadCreatureAndGrantsItsKeywords(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	flyer := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flyer", TypeLine: "Creature — Bird", Power: 1, Toughness: 1,
		Owner: me.ID, Controller: me.ID, Keywords: []string{"flying", "lifelink"},
	})
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	eater := b21Push(g, me.ID, "Eater of Virtue", "Legendary Artifact — Equipment", neoB04Eater, 0, 0)
	advanceToMain(t, g)
	equipTo(t, g, me.ID, eater, flyer)
	if p := effectivePower(t, g, flyer); p != 3 {
		t.Fatalf("equipped power %d, want 3", p)
	}
	if hasEffectiveKeyword(t, g, bear, "flying") {
		t.Fatal("the unequipped bear flies")
	}
	killCreature(t, g, me.ID, flyer)
	passPriorityAroundTable(t, g)
	if g.Exile == nil || !g.Exile.Contains(flyer) {
		t.Fatalf("the dead creature was not exiled (graveyard=%v)", me.Graveyard.Contains(flyer))
	}
	floatMana(t, g, me, "{C}")
	equipTo(t, g, me.ID, eater, bear)
	for _, kw := range []string{"flying", "lifelink"} {
		if !hasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the equipped bear lacks %s from the exiled card", kw)
		}
	}
	if hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the bear has a keyword no exiled card has")
	}
	if p := effectivePower(t, g, bear); p != 4 {
		t.Errorf("power %d, want 4", p)
	}
	// A card that leaves exile stops counting.
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneExile}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, flyer); err != nil {
		t.Fatalf("move: %v", err)
	}
	if hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the bear still flies after the exiled card left exile")
	}
}

func TestNeoB04EaterIgnoresACreatureThatDiesWhileNotEquipped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	other := b16Creature(g, me.ID, "Other", "Creature — Bear", 2, 2, "G")
	eater := b21Push(g, me.ID, "Eater of Virtue", "Legendary Artifact — Equipment", neoB04Eater, 0, 0)
	advanceToMain(t, g)
	equipTo(t, g, me.ID, eater, bear)
	killCreature(t, g, me.ID, other)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(other) {
		t.Error("an unequipped creature was exiled")
	}
}
