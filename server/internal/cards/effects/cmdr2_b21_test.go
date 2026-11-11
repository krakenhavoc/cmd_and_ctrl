package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_b21_test.go — batch 21 of the wave-2 Commander deck cards (#5368):
// Equipment, artifact creatures, a Vehicle and base-P/T matters.

const (
	cmdr2B21WingedBootsOracle     = "a178a789-1f5f-46ca-a90b-574bd236a671"
	cmdr2B21AnchorOracle          = "70f3bc78-201f-4c56-b87b-9e8f2db9f7f7"
	cmdr2B21BloodchiefOracle      = "8985e545-07c3-45f9-9058-fe3c1f50de04"
	cmdr2B21SqueakOracle          = "6ede28a7-57ea-47d6-a5fb-b0f061da2689"
	cmdr2B21GriselbrandMaskOracle = "b46da6bd-578c-4040-97c3-0005b78e31be"
	cmdr2B21HamsphereOracle       = "a6c3aa17-8338-4daa-889e-d36dc715e546"
	cmdr2B21DuskanaOracle         = "68d97ebb-63f8-4795-bddb-03582e13b34d"
	cmdr2B21AugmenterOracle       = "098a43e3-341d-4b27-8c17-ef5990e21b0a"
	cmdr2B21ThopterEngOracle      = "4ac53b34-20d2-4a53-bade-4f998fb6ae67"
	cmdr2B21BladeSplicerOracle    = "194166ad-0179-42a3-86b9-ba7f322ec576"
)

// cmdr2B21Enter puts a creature card with the given base P/T onto the battlefield
// from hand through the entry pipeline, so its enters triggers fire and the entry
// is not a cast.
func cmdr2B21Enter(t *testing.T, g *game.Game, owner *game.Player, name, typeLine, oracle string, p, tough int) uuid.UUID {
	t.Helper()
	c := game.NewCard(name, owner.ID)
	c.TypeLine, c.OracleID, c.Power, c.Toughness = typeLine, oracle, p, tough
	owner.Hand.PushTop(c)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: owner.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, c.InstanceID); err != nil {
		t.Fatalf("enter %s: %v", name, err)
	}
	return c.InstanceID
}

// cmdr2B21CastCreature casts a vanilla creature spell from hand and lets it resolve.
func cmdr2B21CastCreature(t *testing.T, g *game.Game, owner *game.Player, name, typeLine string, p, tough int) uuid.UUID {
	t.Helper()
	c := game.NewCard(name, owner.ID)
	c.TypeLine, c.Power, c.Toughness = typeLine, p, tough
	owner.Hand.PushTop(c)
	if err := g.CastSpell(owner.ID, c.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	return c.InstanceID
}

func cmdr2B21Counters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return battlefieldCardCopy(t, g, id).Counters[game.CounterPlusOne]
}

func cmdr2B21Named(g *game.Game, controller uuid.UUID, name string) int {
	return countBattlefieldNamed(g, controller, name)
}

// --- Winged Boots ---------------------------------------------------------

func TestCmdr2B21WingedBootsGrantsFlyingAndWard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	boots := seedEquipment(g, me.ID, "Winged Boots", cmdr2B21WingedBootsOracle)
	if got := equipPrice(t, g, me.ID, boots, bear); got != "{1}" {
		t.Errorf("equip costs %s, want {1}", got)
	}
	equipTo(t, g, me.ID, boots, bear)
	if ab := effectiveAbilities(t, g, bear); !containsString(ab, "flying") {
		t.Errorf("abilities %v missing flying", ab)
	}

	castAtWardedCreature(t, g, opp, bear)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	c := payUnlessFor(g, opp.ID)
	if c == nil || c.PayCost != "{4}" {
		t.Fatalf("ward prompt = %+v, want a {4} payment from the opponent", c)
	}
	answerPayUnless(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("a declined ward {4} must counter the spell")
	}
}

// --- Arm-Mounted Anchor ---------------------------------------------------

func TestCmdr2B21AnchorPumpsMenaceAndEquipDiscount(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	anchor := seedEquipment(g, me.ID, "Arm-Mounted Anchor", cmdr2B21AnchorOracle)

	if me.Hand.Size() < 2 {
		t.Fatalf("setup: hand %d, want at least 2", me.Hand.Size())
	}
	if got := equipPrice(t, g, me.ID, anchor, bear); got != "{2}" {
		t.Errorf("equip with a full hand costs %s, want {2}", got)
	}
	g.WithWriteLock(func() { me.Hand.Cards = me.Hand.Cards[:1] })
	if got := equipPrice(t, g, me.ID, anchor, bear); got != "{0}" {
		t.Errorf("equip with one card in hand costs %s, want {0}", got)
	}
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	if got := equipPrice(t, g, me.ID, anchor, bear); got != "{0}" {
		t.Errorf("equip with an empty hand costs %s, want {0}", got)
	}

	equipTo(t, g, me.ID, anchor, bear)
	if got := effectivePower(t, g, bear); got != 4 {
		t.Errorf("power %d, want 4", got)
	}
	if got := effectiveToughness(t, g, bear); got != 4 {
		t.Errorf("toughness %d, want 4", got)
	}
	if ab := effectiveAbilities(t, g, bear); !containsString(ab, "menace") {
		t.Errorf("abilities %v missing menace", ab)
	}
}

func TestCmdr2B21AnchorDrawsTwoThenDiscardsTwoUnlessAPirate(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	anchor := seedEquipment(g, me.ID, "Arm-Mounted Anchor", cmdr2B21AnchorOracle)
	equipTo(t, g, me.ID, anchor, bear)
	pirate := handCardFull(me, "Pirate", "Creature — Human Pirate", "", "", nil)
	other := handCardFull(me, "Bear", "Creature — Bear", "", "", nil)
	hand := me.Hand.Size()

	dealCombatDamageToPlayer(g, bear, opp.ID, 4)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Fatalf("hand %d -> %d, want +2", hand, got)
	}
	c := discardChoiceFor(g, me.ID)
	if c == nil || c.ChooseMax != 2 {
		t.Fatalf("discard prompt = %+v, want two cards", c)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{other}); err == nil {
		t.Fatal("one non-Pirate card was accepted")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{pirate}); err != nil {
		t.Fatalf("one Pirate card is a complete discard: %v", err)
	}
	if me.Hand.Contains(pirate) || !me.Hand.Contains(other) {
		t.Error("only the Pirate should have been discarded")
	}
}

func TestCmdr2B21AnchorTwoOrdinaryCardsAreADiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	anchor := seedEquipment(g, me.ID, "Arm-Mounted Anchor", cmdr2B21AnchorOracle)
	equipTo(t, g, me.ID, anchor, bear)
	hand := me.Hand.Size()
	dealCombatDamageToPlayer(g, bear, opp.ID, 4)
	passPriorityAroundTable(t, g)
	discardFromHand(t, g, me.ID)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d, want %d (drew two, discarded two)", got, hand)
	}
}

// --- Blade of the Bloodchief ----------------------------------------------

func TestCmdr2B21BloodchiefCountersOnAnyCreatureDeath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	host := seedBear(g, me.ID)
	blade := seedEquipment(g, me.ID, "Blade of the Bloodchief", cmdr2B21BloodchiefOracle)
	victim := seedBear(g, opp.ID)

	killCreature(t, g, opp.ID, victim)
	passPriorityAroundTable(t, g)
	if got := cmdr2B21Counters(t, g, host); got != 0 {
		t.Errorf("unattached Blade put %d counters", got)
	}

	equipTo(t, g, me.ID, blade, host)
	victim2 := seedBear(g, opp.ID)
	killCreature(t, g, opp.ID, victim2)
	passPriorityAroundTable(t, g)
	if got := cmdr2B21Counters(t, g, host); got != 1 {
		t.Errorf("counters %d, want 1", got)
	}
}

func TestCmdr2B21BloodchiefTwoCountersOnAVampire(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	vamp := b21Push(g, me.ID, "Vampire", "Creature — Vampire", "", 2, 2)
	blade := seedEquipment(g, me.ID, "Blade of the Bloodchief", cmdr2B21BloodchiefOracle)
	equipTo(t, g, me.ID, blade, vamp)
	killCreature(t, g, opp.ID, seedBear(g, opp.ID))
	passPriorityAroundTable(t, g)
	if got := cmdr2B21Counters(t, g, vamp); got != 2 {
		t.Errorf("counters %d, want 2 on a Vampire", got)
	}
}

// --- Sword of the Squeak --------------------------------------------------

func TestCmdr2B21SqueakCountsBasePowerOrToughnessOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	host := b21Push(g, me.ID, "Host", "Creature — Bear", "", 3, 3)
	b21Push(g, me.ID, "Squeaker", "Creature — Mouse", "", 1, 4)
	b21Push(g, me.ID, "Thin", "Creature — Rat", "", 4, 1)
	b21Push(g, me.ID, "Big", "Creature — Bear", "", 2, 2)
	b21Push(g, opp.ID, "Theirs", "Creature — Mouse", "", 1, 1)
	sword := seedEquipment(g, me.ID, "Sword of the Squeak", cmdr2B21SqueakOracle)
	equipTo(t, g, me.ID, sword, host)
	// two of my creatures have base power or toughness 1.
	if got := effectivePower(t, g, host); got != 5 {
		t.Errorf("power %d, want 5", got)
	}
	// A +1/+1 counter on the 2/2 does not change its base: it still doesn't count,
	// and one on the 1/4 does not stop it counting.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.Name == "Squeaker" || c.Name == "Big" {
				if c.Counters == nil {
					c.Counters = map[string]int{}
				}
				c.Counters[game.CounterPlusOne]++
			}
		}
	})
	if got := effectivePower(t, g, host); got != 5 {
		t.Errorf("power %d after counters, want 5", got)
	}
}

func TestCmdr2B21SqueakMayAttachToAnEnteringSqueaker(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sword := seedEquipment(g, me.ID, "Sword of the Squeak", cmdr2B21SqueakOracle)
	hamster := cmdr2B21Enter(t, g, me, "Hamster", "Creature — Hamster", "", 1, 1)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if host := attachmentHostOf(t, g, sword); host.ID != hamster {
		t.Errorf("Sword attached to %v, want the Hamster", host)
	}
	// A Bear does not trigger it.
	cmdr2B21Enter(t, g, me, "Bear", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	if host := attachmentHostOf(t, g, sword); host.ID != hamster {
		t.Error("a Bear must not move the Sword")
	}
}

func TestCmdr2B21SqueakIgnoresAnOpponentsSqueaker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sword := seedEquipment(g, me.ID, "Sword of the Squeak", cmdr2B21SqueakOracle)
	cmdr2B21Enter(t, g, opp, "Squirrel", "Creature — Squirrel", "", 1, 1)
	if len(g.PendingChoices) != 0 {
		t.Errorf("an opponent's Squirrel opened a prompt: %+v", g.PendingChoices)
	}
	if host := attachmentHostOf(t, g, sword); host.ID != uuid.Nil {
		t.Error("an opponent's Squirrel must not trigger the Sword")
	}
}

// --- Mask of Griselbrand --------------------------------------------------

func TestCmdr2B21MaskGrantsFlyingAndLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	mask := seedEquipment(g, me.ID, "Mask of Griselbrand", cmdr2B21GriselbrandMaskOracle)
	equipTo(t, g, me.ID, mask, bear)
	ab := effectiveAbilities(t, g, bear)
	if !containsString(ab, "flying") || !containsString(ab, "lifelink") {
		t.Errorf("abilities %v, want flying and lifelink", ab)
	}
}

func TestCmdr2B21MaskPaysItsPowerInLifeToDrawThatMany(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	big := b21Push(g, me.ID, "Big", "Creature — Demon", "", 3, 3)
	mask := seedEquipment(g, me.ID, "Mask of Griselbrand", cmdr2B21GriselbrandMaskOracle)
	equipTo(t, g, me.ID, mask, big)
	life, hand := me.Life, me.Hand.Size()

	killCreature(t, g, me.ID, big)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Life != life-3 {
		t.Errorf("life %d -> %d, want -3", life, me.Life)
	}
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand %d -> %d, want +3", hand, got)
	}
}

func TestCmdr2B21MaskDeclineLeavesLifeAndHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	big := b21Push(g, me.ID, "Big", "Creature — Demon", "", 3, 3)
	mask := seedEquipment(g, me.ID, "Mask of Griselbrand", cmdr2B21GriselbrandMaskOracle)
	equipTo(t, g, me.ID, mask, big)
	life, hand := me.Life, me.Hand.Size()
	killCreature(t, g, me.ID, big)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if me.Life != life || me.Hand.Size() != hand {
		t.Errorf("declining changed life %d -> %d or hand %d -> %d", life, me.Life, hand, me.Hand.Size())
	}
}

// --- Rolling Hamsphere ----------------------------------------------------

func TestCmdr2B21HamsphereGrowsWithHamsters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ship := pushDiesCreatureForTest(g, me.ID, "Rolling Hamsphere", cmdr2B21HamsphereOracle, "Artifact — Vehicle", 4, 4)
	if got := effectivePower(t, g, ship); got != 4 {
		t.Errorf("power %d with no Hamsters, want 4", got)
	}
	pushVanillaHamster := func(owner uuid.UUID) {
		pushDiesCreatureForTest(g, owner, "Hamster", "", "Token Creature — Hamster", 1, 1)
	}
	pushVanillaHamster(me.ID)
	pushVanillaHamster(me.ID)
	pushVanillaHamster(opp.ID)
	if got := effectivePower(t, g, ship); got != 6 {
		t.Errorf("power %d with two of my Hamsters, want 6", got)
	}
	if got := effectiveToughness(t, g, ship); got != 6 {
		t.Errorf("toughness %d, want 6", got)
	}
}

func TestCmdr2B21HamsphereAttackMakesThreeHamstersThenDealsXDamage(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ship := pushDiesCreatureForTest(g, me.ID, "Rolling Hamsphere", cmdr2B21HamsphereOracle, "Artifact Creature — Vehicle", 4, 4)
	pushDiesCreatureForTest(g, me.ID, "Hamster", "", "Token Creature — Hamster", 1, 1)
	life := opp.Life
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(ship, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	answerTriggerTargets2(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if got := cmdr2B21Named(g, me.ID, "Hamster"); got != 4 {
		t.Errorf("Hamsters %d, want 4 (one + three)", got)
	}
	if opp.Life != life-4 {
		t.Errorf("opponent life %d -> %d, want -4 (X = Hamsters after the tokens)", life, opp.Life)
	}
}

// answerTriggerTargets2 answers a pick-target prompt with a player.
func answerTriggerTargets2(t *testing.T, g *game.Game, chooser, player uuid.UUID) {
	t.Helper()
	pick := latestPickTarget(g, chooser)
	if pick == nil {
		t.Fatalf("no pick_target prompt for %s", chooser)
	}
	if err := g.ResolvePickTargets(pick.ID, chooser, []game.TargetRef{{Kind: game.TargetPlayer, ID: player}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
}

// --- Duskana, the Rage Mother ---------------------------------------------

func TestCmdr2B21DuskanaDrawsForEachBaseTwoByTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b21Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	b21Push(g, me.ID, "Bear 2", "Creature — Bear", "", 2, 2)
	b21Push(g, me.ID, "Not two", "Creature — Bear", "", 2, 3)
	counters := b21Push(g, me.ID, "Pumped one", "Creature — Bear", "", 1, 1)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == counters {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterPlusOne: 1}
			}
		}
	})
	hand := me.Hand.Size()
	cmdr2B21Enter(t, g, me, "Duskana, the Rage Mother", "Legendary Creature — Bear", cmdr2B21DuskanaOracle, 5, 5)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d -> %d, want +2", hand, got)
	}
}

func TestCmdr2B21DuskanaPumpsAttackingBaseTwoByTwos(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushDiesCreatureForTest(g, me.ID, "Duskana, the Rage Mother", cmdr2B21DuskanaOracle, "Legendary Creature — Bear", 5, 5)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	three := pushVanillaCreature(g, me.ID, "Three", 3, 3)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{bear, three} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, bear); got != 5 {
		t.Errorf("attacking 2/2 power %d, want 5", got)
	}
	if got := effectiveToughness(t, g, bear); got != 5 {
		t.Errorf("attacking 2/2 toughness %d, want 5", got)
	}
	if got := effectivePower(t, g, three); got != 3 {
		t.Errorf("attacking 3/3 power %d, want it unchanged", got)
	}
}

// --- Rapid Augmenter ------------------------------------------------------

func TestCmdr2B21AugmenterGivesHasteToBasePowerOneEntrants(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pushDiesCreatureForTest(g, me.ID, "Rapid Augmenter", cmdr2B21AugmenterOracle, "Creature — Otter Artificer", 1, 3)
	small := cmdr2B21Enter(t, g, me, "Small", "Creature — Mouse", "", 1, 4)
	passPriorityAroundTable(t, g)
	if ab := effectiveAbilities(t, g, small); !containsString(ab, "haste") {
		t.Errorf("abilities %v, want haste for a base power 1 creature", ab)
	}
	big := cmdr2B21Enter(t, g, me, "Big", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	if ab := effectiveAbilities(t, g, big); containsString(ab, "haste") {
		t.Errorf("abilities %v, a base power 2 creature must not gain haste", ab)
	}
}

func TestCmdr2B21AugmenterGrowsOnlyWhenTheEntrantWasntCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	aug := pushDiesCreatureForTest(g, me.ID, "Rapid Augmenter", cmdr2B21AugmenterOracle, "Creature — Otter Artificer", 1, 3)

	cmdr2B21CastCreature(t, g, me, "Cast Bear", "Creature — Bear", 2, 2)
	if got := cmdr2B21Counters(t, g, aug); got != 0 {
		t.Fatalf("a cast creature put %d counters, want 0", got)
	}
	assertRestrictions(t, g, aug, 0)

	cmdr2B21Enter(t, g, me, "Put Bear", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	if got := cmdr2B21Counters(t, g, aug); got != 1 {
		t.Errorf("counters %d, want 1 for a creature that wasn't cast", got)
	}
	assertRestrictions(t, g, aug, game.CantBeBlocked)
}

func TestCmdr2B21AugmenterIgnoresOpponentsCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aug := pushDiesCreatureForTest(g, me.ID, "Rapid Augmenter", cmdr2B21AugmenterOracle, "Creature — Otter Artificer", 1, 3)
	cmdr2B21Enter(t, g, opp, "Their Bear", "Creature — Bear", "", 1, 1)
	passPriorityAroundTable(t, g)
	if got := cmdr2B21Counters(t, g, aug); got != 0 {
		t.Errorf("an opponent's creature put %d counters", got)
	}
}

// --- Thopter Engineer -----------------------------------------------------

func TestCmdr2B21ThopterEngineerMakesAThopterAndGivesArtifactCreaturesHaste(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	golem := b21Push(g, me.ID, "Golem", "Artifact Creature — Golem", "", 3, 3)
	plain := seedBear(g, me.ID)
	theirs := b21Push(g, opp.ID, "Their Golem", "Artifact Creature — Golem", "", 3, 3)
	cmdr2B21Enter(t, g, me, "Thopter Engineer", "Creature — Human Artificer", cmdr2B21ThopterEngOracle, 1, 3)
	passPriorityAroundTable(t, g)
	if got := cmdr2B21Named(g, me.ID, "Thopter"); got != 1 {
		t.Errorf("Thopters %d, want 1", got)
	}
	if ab := effectiveAbilities(t, g, golem); !containsString(ab, "haste") {
		t.Errorf("my artifact creature has %v, want haste", ab)
	}
	if ab := effectiveAbilities(t, g, plain); containsString(ab, "haste") {
		t.Errorf("a nonartifact creature has %v, want no haste", ab)
	}
	if ab := effectiveAbilities(t, g, theirs); containsString(ab, "haste") {
		t.Errorf("an opponent's artifact creature has %v, want no haste", ab)
	}
}

// --- Blade Splicer --------------------------------------------------------

func TestCmdr2B21BladeSplicerMakesAGolemWithFirstStrike(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	theirs := b21Push(g, opp.ID, "Their Golem", "Artifact Creature — Golem", "", 3, 3)
	splicer := cmdr2B21Enter(t, g, me, "Blade Splicer", "Creature — Phyrexian Human Artificer", cmdr2B21BladeSplicerOracle, 1, 1)
	passPriorityAroundTable(t, g)
	if got := cmdr2B21Named(g, me.ID, "Phyrexian Golem"); got != 1 {
		t.Fatalf("Golems %d, want 1", got)
	}
	var golem uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Phyrexian Golem" && c.Controller == me.ID {
				golem = c.InstanceID
			}
		}
	})
	if got := effectivePower(t, g, golem); got != 3 {
		t.Errorf("Golem power %d, want 3", got)
	}
	if ab := effectiveAbilities(t, g, golem); !containsString(ab, "first strike") {
		t.Errorf("Golem has %v, want first strike", ab)
	}
	if ab := effectiveAbilities(t, g, splicer); containsString(ab, "first strike") {
		t.Errorf("the Splicer has %v, it is no Golem", ab)
	}
	if ab := effectiveAbilities(t, g, theirs); containsString(ab, "first strike") {
		t.Errorf("an opponent's Golem has %v, want none", ab)
	}
}
