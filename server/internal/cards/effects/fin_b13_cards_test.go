package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fin_b13_cards_test.go — Final Fantasy batches 13 and 14 (#4462).

const (
	finB13CarpetOracle      = "2ffc6372-f63b-4f32-8dd0-2d7938aeb412"
	finB13AtraxaOracle      = "abbcb153-0763-44c6-964f-b4ff0eb64257"
	finB13BruseOracle       = "e39a1f6b-6b13-4a8b-ac0c-63ede4636812"
	finB13IshaiOracle       = "4aae7102-301a-429d-b169-4f462384d55d"
	finB13KinnanOracle      = "8d11aa49-d4cd-48b1-aa0f-8548fa733416"
	finB13KraumOracle       = "5f694f2d-9349-4988-b2b4-788136d4293e"
	finB13VialSmasherOracle = "1b355139-4891-436c-9079-0ce17ecc2d65"
	finB13StadiumOracle     = "2ed0c6fc-b8b4-47da-a04b-1d995bdeecb0"
	finB13EdgarOracle       = "264efed9-65ab-40ad-877f-c6c25ca0281e"
	finB13WhaleOracle       = "13b4aeac-4530-4e09-a547-548b4884bb2f"
	finB14CactuarOracle     = "e0bdfbb4-3060-4492-bd9b-0c47e719400a"
	finB14ExcaliburOracle   = "bfb6651c-e0b3-4dbd-b26a-3eeb690985fc"
)

func TestFinB13Registered(t *testing.T) {
	for oracle, name := range map[string]string{
		finB13BruseOracle:     "Bruse Tarl, Boorish Herder",
		finB13IshaiOracle:     "Ishai, Ojutai Dragonspeaker",
		finB13KraumOracle:     "Kraum, Ludevic's Opus",
		finB14CactuarOracle:   "Jumbo Cactuar",
		finB14ExcaliburOracle: "Excalibur II",
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

// --- Jumbo Cactuar -----------------------------------------------------------

func TestFinB14JumboCactuarGetsNineThousandNineHundredNinetyNineWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cactuar := b12Push(g, me.ID, "Jumbo Cactuar", "Creature — Plant", finB14CactuarOracle, 1, 7)
	advanceToMain(t, g)
	if p := effectivePower(t, g, cactuar); p != 1 {
		t.Fatalf("power %d before attacking, want 1", p)
	}
	declareAttack(t, g, opp.ID, cactuar)
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, cactuar); p != 10000 {
		t.Errorf("power %d after attacking, want 10000", p)
	}
}

// --- Excalibur II ------------------------------------------------------------

func TestFinB14ExcaliburIICountsChargeCountersOncePerLifeGain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	sword := seedEquipment(g, me.ID, "Excalibur II", finB14ExcaliburOracle)
	equipTo(t, g, me.ID, sword, bear)
	for i := 0; i < 3; i++ {
		g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 5) })
		passPriorityAroundTable(t, g)
	}
	if got := b12Counter(t, g, sword, "charge"); got != 3 {
		t.Fatalf("%d charge counters after three life gains, want 3", got)
	}
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("equipped Bear is %d/%d, want 5/5", p, tough)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, g.Seats[1].ID, 4) })
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, sword, "charge"); got != 3 {
		t.Errorf("an opponent's life gain made a counter: %d", got)
	}
}

// --- Ishai -------------------------------------------------------------------

func TestFinB13IshaiGrowsOnlyWhenAnOpponentCasts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ishai := b12Push(g, me.ID, "Ishai, Ojutai Dragonspeaker", "Legendary Creature — Bird Monk", finB13IshaiOracle, 1, 1)
	assertKeywords(t, g, ishai, "flying")
	b13OpponentCasts(t, g, opp, "Shock", "Instant", "", "", nil)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, ishai, game.CounterPlusOne); got != 1 {
		t.Fatalf("%d counters after an opponent's spell, want 1", got)
	}
	b12PlayFromHand(t, g, "Divination", "Sorcery", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if got := countersOn(g, ishai, game.CounterPlusOne); got != 1 {
		t.Errorf("your own spell grew Ishai: %d counters", got)
	}
}

// --- Kraum -------------------------------------------------------------------

func TestFinB13KraumDrawsOnlyOnAnOpponentsSecondSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Kraum, Ludevic's Opus", "Legendary Creature — Zombie Horror", finB13KraumOracle, 4, 4)
	hand := me.Hand.Size()
	b13OpponentCasts(t, g, opp, "Shock", "Instant", "", "", nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatalf("hand %d → %d after their first spell, want no draw", hand, me.Hand.Size())
	}
	b13OpponentCasts(t, g, opp, "Shock", "Instant", "", "", nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d → %d after their second spell, want +1", hand, me.Hand.Size())
	}
	b13OpponentCasts(t, g, opp, "Shock", "Instant", "", "", nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("their third spell drew again: %d", me.Hand.Size())
	}
}

func TestFinB13KraumIgnoresYourOwnSecondSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Kraum, Ludevic's Opus", "Legendary Creature — Zombie Horror", finB13KraumOracle, 4, 4)
	hand := me.Hand.Size()
	b12PlayFromHand(t, g, "Divination", "Sorcery", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	b12PlayFromHand(t, g, "Divination", "Sorcery", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d, your own spells must not draw", hand, me.Hand.Size())
	}
}

// --- Bruse Tarl --------------------------------------------------------------

func TestFinB13BruseTarlGrantsDoubleStrikeAndLifelinkOnEnterAndOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	b12PlayFromHand(t, g, "Bruse Tarl, Boorish Herder", "Legendary Creature — Human Ally", finB13BruseOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the enter trigger asked for no target")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
		t.Fatal("targeted a creature an opponent controls")
	}
	pickTriggerTarget(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	for _, kw := range []string{"double strike", "lifelink"} {
		if !hasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the Bear did not gain %s on enter", kw)
		}
	}
}

func TestFinB13BruseTarlTriggersOnAttackToo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bruse := b12Push(g, me.ID, "Bruse Tarl, Boorish Herder", "Legendary Creature — Human Ally", finB13BruseOracle, 3, 3)
	advanceToMain(t, g)
	declareAttack(t, g, opp.ID, bruse)
	pickTriggerTarget(t, g, me.ID, bruse)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bruse, "double strike") || !hasEffectiveKeyword(t, g, bruse, "lifelink") {
		t.Error("Bruse Tarl did not grant itself the keywords when attacking")
	}
}

// --- Genji Glove -------------------------------------------------------------

const finB14GenjiOracle = "8242cd78-6675-43b5-8924-0a88c3aa6f76"

func TestFinB14GenjiGloveGivesDoubleStrikeUntapsAndAddsOneCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	glove := seedEquipment(g, me.ID, "Genji Glove", finB14GenjiOracle)
	equipTo(t, g, me.ID, glove, bear)
	if !hasEffectiveKeyword(t, g, bear, "double strike") {
		t.Fatal("the equipped Bear has no double strike")
	}
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if b13Tapped(t, g, bear) {
		t.Error("the equipped creature stayed tapped after the trigger")
	}
	advanceTo(t, g, game.StepEndCombat)
	if next := stepOnce(t, g); next.Step != game.StepBeginCombat {
		t.Fatalf("after the first combat: %+v, want an additional combat", next)
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker in the second combat: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if !b13Tapped(t, g, bear) {
		t.Error("the trigger untapped the creature in the additional combat")
	}
	advanceTo(t, g, game.StepEndCombat)
	if next := stepOnce(t, g); next.Step == game.StepBeginCombat {
		t.Error("the additional combat added a third")
	}
}

// --- Raubahn ------------------------------------------------------------------

const finB14RaubahnOracle = "d81d5749-ab4d-42c5-8903-0b400f10fe33"

func TestFinB14RaubahnWardCostsLifeEqualToItsPowerNow(t *testing.T) {
	g, _, opp, raubahn := wardTable(t, "Raubahn, Bull of Ala Mhigo", "Legendary Creature — Human Warrior", finB14RaubahnOracle)
	g.WithWriteLock(func() { _ = g.AddCounterByForEffect(uuid.Nil, raubahn, game.CounterPlusOne, 3) })
	castAtWardedCreature(t, g, opp, raubahn)
	life := opp.Life
	prompt := passUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the spell's controller")
	}
	if prompt.LifeCost != 4 {
		t.Fatalf("prompt LifeCost = %d, want Raubahn's power 4", prompt.LifeCost)
	}
	if err := g.ResolveConfirm(prompt.ID, opp.ID, true); err != nil {
		t.Fatalf("pay: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("life %d → %d, want -4", life, opp.Life)
	}
}

func TestFinB14RaubahnWardDeclinedCountersTheSpell(t *testing.T) {
	g, _, opp, raubahn := wardTable(t, "Raubahn, Bull of Ala Mhigo", "Legendary Creature — Human Warrior", finB14RaubahnOracle)
	blade := castAtWardedCreature(t, g, opp, raubahn)
	prompt := passUntilConfirmFor(t, g, opp.ID)
	if prompt == nil || prompt.LifeCost != 1 {
		t.Fatalf("prompt %v, want a 1-life ward (power 1)", prompt)
	}
	if err := g.ResolveConfirm(prompt.ID, opp.ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(raubahn) || !opp.Graveyard.Contains(blade) {
		t.Error("declining the ward must counter the spell")
	}
}

func TestFinB14RaubahnAttachesAnEquipmentToAnAttackingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	raubahn := b12Push(g, me.ID, "Raubahn, Bull of Ala Mhigo", "Legendary Creature — Human Warrior", finB14RaubahnOracle, 2, 2)
	friend := seedBear(g, me.ID)
	idle := seedBear(g, me.ID)
	sword := finABEquipment(g, me.ID, "Sword")
	advanceToMain(t, g)
	declareAttack(t, g, opp.ID, raubahn, friend)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the attack trigger asked for no target")
	}
	pickTriggerTarget(t, g, me.ID, sword)
	if err := g.ResolvePickTarget(latestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: idle}); err == nil {
		t.Fatal("targeted a creature that is not attacking")
	}
	pickTriggerTarget(t, g, me.ID, friend)
	passPriorityAroundTable(t, g)
	c := b12Card(t, g, sword)
	if c.AttachedTo.ID != friend {
		t.Errorf("the Sword is attached to %v, want the attacking Bear", c.AttachedTo)
	}
}

// --- Clive's Hideaway -----------------------------------------------------------

const finB14HideawayOracle = "283f743f-6e79-49de-b7ed-08e6ffb64cc6"

func TestFinB14ClivesHideawayNeedsFourLegendaryCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	top := stackTopLibrary(me, "A", "B", "C", "D")
	land := game.NewCard("Clive's Hideaway", me.ID)
	land.OracleID = finB14HideawayOracle
	land.TypeLine = "Land — Town"
	enterWithHideaway(t, g, land, top[1])
	if c, ok := cardOnBattlefield(g, land.InstanceID); !ok || c.Tapped {
		t.Fatal("Clive's Hideaway should enter untapped")
	}
	activate := func() {
		t.Helper()
		untapForTest(g, land.InstanceID)
		floatForTest(g, me, "CC")
		if err := g.ActivateCatalogAbility(me.ID, land.InstanceID, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	for _, n := range []string{"Legend A", "Legend B", "Legend C"} {
		b12Push(g, me.ID, n, "Legendary Creature — Human", "", 1, 1)
	}
	pushVanillaCreature(g, me.ID, "Not Legendary", 1, 1)
	activate()
	if playHiddenGrant(g, top[1]) != nil {
		t.Fatal("granted with only three legendary creatures")
	}
	b12Push(g, me.ID, "Legend D", "Legendary Creature — Human", "", 1, 1)
	activate()
	if playHiddenGrant(g, top[1]) == nil {
		t.Fatal("no grant with four legendary creatures")
	}
}

// --- Firion -------------------------------------------------------------------

const finB14FirionOracle = "df42c768-34ca-4ebf-8f66-eaccf7261f70"

func TestFinB14FirionCopiesAnEnteringEquipmentWithCheaperEquipAndSacrificesItAtUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Firion, Wild Rose Warrior", "Legendary Creature — Human Rebel Warrior", finB14FirionOracle, 3, 3)
	bear := seedBear(g, me.ID)
	if hasEffectiveKeyword(t, g, bear, "haste") {
		t.Fatal("an unequipped Bear has haste")
	}
	b12PlayFromHand(t, g, "Genji Glove", "Artifact — Equipment", finB14GenjiOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	var originals, tokens []uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.OracleID == finB14GenjiOracle {
				if c.IsToken() {
					tokens = append(tokens, c.InstanceID)
				} else {
					originals = append(originals, c.InstanceID)
				}
			}
		}
	})
	if len(originals) != 1 || len(tokens) != 1 {
		t.Fatalf("%d Genji Gloves and %d tokens, want 1 and 1", len(originals), len(tokens))
	}
	floatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, originals[0], 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Fatal("the original Equipment equipped for {1}")
	}
	if err := g.ActivateCatalogAbility(me.ID, tokens[0], 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("the token copy's equip should cost {1}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("the equipped Bear has no haste")
	}
	advanceToEndStepOf(t, g, 0)
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	g.ReadSnapshot(func() {
		if g.Battlefield.Contains(tokens[0]) {
			t.Error("the token copy was not sacrificed at the next upkeep")
		}
	})
}

func TestFinB14FirionIgnoresATokenEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Firion, Wild Rose Warrior", "Legendary Creature — Human Rebel Warrior", finB14FirionOracle, 3, 3)
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, game.Card{
			Name: "Sword Token", TypeLine: "Token Artifact — Equipment", Owner: me.ID, Controller: me.ID,
		}, 1)
	})
	passPriorityAroundTable(t, g)
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.HasSubtype("Equipment") {
				n++
			}
		}
	})
	if n != 1 {
		t.Errorf("%d Equipment on the battlefield, want only the token (no copy of a token)", n)
	}
}

// --- Strixhaven Stadium -----------------------------------------------------

func TestFinB13StrixhavenStadiumCountsPointsAndEndsTheGame(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stadium := b12Push(g, me.ID, "Strixhaven Stadium", "Artifact", finB13StadiumOracle, 0, 0)
	advanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, stadium, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if got := b12Counter(t, g, stadium, "point"); got != 1 {
		t.Fatalf("%d point counters after tapping for mana, want 1", got)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(stadium, "point", 8) })
	bear := seedBear(g, me.ID)
	declareAttack(t, g, opp.ID, bear)
	advanceTo(t, g, game.StepEndCombat)
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, stadium, "point"); got != 10 && got != 0 {
		t.Fatalf("%d point counters, want 10 (then removed) after combat damage", got)
	}
	if !opp.Eliminated {
		t.Errorf("the opponent did not lose with ten point counters")
	}
}

func TestFinB13StrixhavenStadiumRemovesACounterWhenACreatureHitsYou(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stadium := b12Push(g, me.ID, "Strixhaven Stadium", "Artifact", finB13StadiumOracle, 0, 0)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(stadium, "point", 3) })
	attacker := pushVanillaCreature(g, opp.ID, "Raider", 2, 2)
	advanceToMainOf(t, g, 1)
	declareAttack(t, g, me.ID, attacker)
	advanceTo(t, g, game.StepEndCombat)
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, stadium, "point"); got != 2 {
		t.Errorf("%d point counters after being hit, want 2", got)
	}
}

// --- Vial Smasher ---------------------------------------------------------------

func finB13CastDivination(t *testing.T, g *game.Game) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Divination", TypeLine: "Sorcery", ManaCost: "{2}{U}",
		Owner: me.ID, Controller: me.ID,
	})
	floatForTest(g, me, "UUU")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
}

func TestFinB13VialSmasherHitsOneRandomOpponentForTheFirstSpellsManaValueOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Vial Smasher the Fierce", "Legendary Creature — Goblin Berserker", finB13VialSmasherOracle, 2, 3)
	advanceToMain(t, g)
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}
	finB13CastDivination(t, g)
	passPriorityAroundTable(t, g)
	hit := 0
	for _, p := range g.Seats[1:] {
		switch before[p.ID] - p.Life {
		case 0:
		case 3:
			hit++
		default:
			t.Errorf("%s lost %d life, want 0 or 3", p.Name, before[p.ID]-p.Life)
		}
	}
	if hit != 1 || me.Life != before[me.ID] {
		t.Fatalf("%d opponents were hit (own life %d → %d), want exactly one opponent", hit, before[me.ID], me.Life)
	}
	// The second spell this turn does nothing.
	after := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		after[p.ID] = p.Life
	}
	finB13CastDivination(t, g)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if p.Life != after[p.ID] {
			t.Errorf("%s lost life to the second spell", p.Name)
		}
	}
}

func TestFinB13VialSmasherLetsYouPickAPlaneswalkerOfTheRandomOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Vial Smasher the Fierce", "Legendary Creature — Goblin Berserker", finB13VialSmasherOracle, 2, 3)
	walkers := map[uuid.UUID]uuid.UUID{}
	for _, p := range g.Seats[1:] {
		walkers[p.ID] = p1g2Planeswalker(g, p.ID, 10)
	}
	advanceToMain(t, g)
	finB13CastDivination(t, g)
	passPriorityAroundTable(t, g)
	choice := latestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID)
	if choice == nil {
		t.Fatal("no recipient prompt although every opponent controls a planeswalker")
	}
	idx := -1
	for i, o := range choice.PickOptions {
		if len(o.Cards) == 1 {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatalf("no planeswalker option in %+v", choice.PickOptions)
	}
	if err := g.ResolveOptionPick(choice.ID, me.ID, idx); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats[1:] {
		if p.Life != 40 {
			t.Errorf("%s lost life although the walker was chosen", p.Name)
		}
	}
	damaged := 0
	for _, w := range walkers {
		switch got := b12Card(t, g, w).Counters[game.CounterLoyalty]; got {
		case 10:
		case 7:
			damaged++
		default:
			t.Errorf("a planeswalker has loyalty %d, want 10 or 7", got)
		}
	}
	if damaged != 1 {
		t.Errorf("%d planeswalkers took damage, want exactly one", damaged)
	}
}

// --- Carpet of Flowers -------------------------------------------------------------

func finB13Islands(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		b12Permanent(g, owner, "Island", "Basic Land — Island")
	}
	b12Permanent(g, owner, "Swamp", "Basic Land — Swamp")
}

func TestFinB13CarpetOfFlowersAddsOneColorPerIslandTheOpponentControlsOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Permanent(g, me.ID, "Carpet of Flowers", "Enchantment")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].Name == "Carpet of Flowers" {
				g.Battlefield.Cards[i].OracleID = finB13CarpetOracle
			}
		}
	})
	finB13Islands(g, opp.ID, 3)
	advanceToMain(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the main-phase trigger asked for no opponent")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	pickMana := latestChoiceOfKindFor(g, game.PendingChoiceMana, me.ID)
	if pickMana == nil {
		t.Fatal("no colour prompt for the added mana")
	}
	if err := g.ResolveManaChoice(pickMana.ID, me.ID, "U"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.ManaPool.CountColor("U"); got != 3 || me.ManaPool.Total() != 3 {
		t.Fatalf("pool has %d blue of %d mana, want 3 blue (three Islands)", got, me.ManaPool.Total())
	}
	// The second main phase does not trigger once mana was added.
	advanceTo(t, g, game.StepPostcombatMain)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("the ability triggered again after adding mana this turn")
	}
}

func TestFinB13CarpetOfFlowersTriggersInTheSecondMainWhenTheFirstWasDeclined(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Permanent(g, me.ID, "Carpet of Flowers", "Enchantment")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].Name == "Carpet of Flowers" {
				g.Battlefield.Cards[i].OracleID = finB13CarpetOracle
			}
		}
	})
	finB13Islands(g, opp.ID, 2)
	advanceToMain(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt in the first main phase")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if me.ManaPool.Total() != 0 {
		t.Fatal("mana was added after declining")
	}
	advanceTo(t, g, game.StepPostcombatMain)
	if latestPickTarget(g, me.ID) == nil {
		t.Error("declining the first main phase must leave the second one live")
	}
}

// --- Atraxa -------------------------------------------------------------------

func finB13AtraxaLibrary(me *game.Player, typeLines ...string) []uuid.UUID {
	ids := make([]uuid.UUID, len(typeLines))
	for i := len(typeLines) - 1; i >= 0; i-- {
		c := game.NewCard("Revealed "+typeLines[i], me.ID)
		c.TypeLine = typeLines[i]
		ids[i] = c.InstanceID
		me.Library.PushTop(c)
	}
	return ids
}

func TestFinB13AtraxaPutsOneCardPerTypeAndBottomsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	top := finB13AtraxaLibrary(me,
		"Artifact Creature — Golem", // 0: artifact or creature
		"Creature — Bear",           // 1
		"Creature — Wolf",           // 2
		"Basic Land — Forest",       // 3
		"Instant",                   // 4
		"Sorcery",                   // 5
		"Enchantment",               // 6
		"Basic Land — Island",       // 7
		"Creature — Elk",            // 8
		"Legendary Planeswalker — Test",
	)
	b12PlayFromHand(t, g, "Atraxa, Grand Unifier", "Legendary Creature — Phyrexian Angel", finB13AtraxaOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	answer := func(pick ...uuid.UUID) {
		t.Helper()
		c := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
		if c == nil {
			t.Fatalf("no card prompt open")
		}
		if err := g.ResolveChooseCards(c.ID, me.ID, pick); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
	}
	// artifact: the Golem is the only candidate; take it for the artifact slot.
	answer(top[0])
	// battle: nothing to ask. creature: the Golem is spent, so the Bear.
	answer(top[1])
	// enchantment
	answer(top[6])
	// instant
	answer()
	// land: take the Forest
	answer(top[3])
	// planeswalker
	answer(top[9])
	// sorcery
	answer(top[5])
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 6 {
		t.Errorf("hand grew by %d, want 6", got)
	}
	for _, id := range []uuid.UUID{top[0], top[1], top[6], top[3], top[9], top[5]} {
		if !me.Hand.Contains(id) {
			t.Errorf("card %s is not in the hand", id)
		}
	}
	// The four left over are together at the bottom of the library: the
	// zone keeps its bottom at index 0, and nothing else was moved there.
	want := map[uuid.UUID]bool{top[2]: true, top[4]: true, top[7]: true, top[8]: true}
	for i := 0; i < 4; i++ {
		if !want[me.Library.Cards[i].InstanceID] {
			t.Errorf("library index %d is %q, want one of the four unchosen cards at the bottom", i, me.Library.Cards[i].Name)
		}
	}
}

func TestFinB13AtraxaAsksNothingForTypesWithNoCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	finB13AtraxaLibrary(me, "Basic Land — Forest", "Basic Land — Island", "Basic Land — Swamp", "Basic Land — Plains", "Basic Land — Mountain",
		"Basic Land — Forest", "Basic Land — Island", "Basic Land — Swamp", "Basic Land — Plains", "Basic Land — Mountain")
	b12PlayFromHand(t, g, "Atraxa, Grand Unifier", "Legendary Creature — Phyrexian Angel", finB13AtraxaOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	c := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if c == nil || len(c.ChooseCards) != 10 {
		t.Fatalf("want one land prompt over ten lands, got %+v", c)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID) != nil {
		t.Error("asked about a type with no card")
	}
}

// --- Edgar, King of Figaro ---------------------------------------------------------

func finB13FlipEvents(g *game.Game, player uuid.UUID) (heads, won, total int) {
	g.ReadSnapshot(func() {
		for _, ev := range g.EventsThisTurn() {
			if ev.Kind != game.EventFlipCoin || ev.Actor != player {
				continue
			}
			total++
			if ev.Label == "heads" {
				heads++
			}
			if ev.Won {
				won++
			}
		}
	})
	return
}

func TestFinB13EdgarDrawsForEachArtifactYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Permanent(g, me.ID, "Rock", "Artifact")
	b12Permanent(g, me.ID, "Bauble", "Artifact Creature — Golem")
	b12Permanent(g, g.Seats[1].ID, "Their Rock", "Artifact")
	hand := me.Hand.Size()
	b12PlayFromHand(t, g, "Edgar, King of Figaro", "Legendary Creature — Human Artificer Noble", finB13EdgarOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d cards, want 2 (two artifacts you control)", got)
	}
}

func TestFinB13EdgarTwoHeadedCoinMakesTheFirstFlipEachTurnHeadsAndWon(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Edgar, King of Figaro", "Legendary Creature — Human Artificer Noble", finB13EdgarOracle, 4, 5)
	advanceToMain(t, g)
	g.WithWriteLock(func() { _, _ = g.FlipCoinsForEffect(game.RandomDraw{Player: me.ID}, 6) })
	if heads, won, total := finB13FlipEvents(g, me.ID); total != 6 || heads != 6 || won != 6 {
		t.Fatalf("first flip: %d of %d heads, %d won, want all six heads and won", heads, total, won)
	}
	// The second flip action this turn is not changed: not every flip is won.
	g.WithWriteLock(func() { _, _ = g.FlipCoinsForEffect(game.RandomDraw{Player: me.ID}, 40) })
	if _, won, total := finB13FlipEvents(g, me.ID); total != 46 || won != 6 {
		t.Errorf("after a second flip: %d flips, %d won, want 46 and 6 (the second flip is unmodified)", total, won)
	}
}

func TestFinB13EdgarTwoHeadedCoinWinsACalledFlipWhateverIsCalled(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Edgar, King of Figaro", "Legendary Creature — Human Artificer Noble", finB13EdgarOracle, 4, 5)
	advanceToMain(t, g)
	for _, call := range []string{"tails", "heads"} {
		var id uuid.UUID
		g.WithWriteLock(func() { id = g.FlipCoinForEffect(game.CoinFlipSpec{Flipper: me.ID, Coins: 3, Question: "call"}) })
		if err := g.ResolveCoinCall(id, me.ID, call); err != nil {
			t.Fatalf("ResolveCoinCall(%s): %v", call, err)
		}
		if call == "tails" {
			if heads, won, total := finB13FlipEvents(g, me.ID); total != 3 || heads != 3 || won != 3 {
				t.Fatalf("called tails: %d heads, %d won of %d, want every coin heads and won", heads, won, total)
			}
		}
	}
	// The second call this turn is the ordinary 50/50: not all six flips are won.
	if _, _, total := finB13FlipEvents(g, me.ID); total != 6 {
		t.Errorf("%d flips, want six", total)
	}
}

func TestFinB13EdgarsRuleIsYoursAloneAndNeedsEdgarOnTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	edgar := b12Push(g, me.ID, "Edgar, King of Figaro", "Legendary Creature — Human Artificer Noble", finB13EdgarOracle, 4, 5)
	advanceToMain(t, g)
	g.WithWriteLock(func() { _, _ = g.FlipCoinsForEffect(game.RandomDraw{Player: opp.ID}, 40) })
	if _, won, _ := finB13FlipEvents(g, opp.ID); won != 0 {
		t.Errorf("an opponent's flips were won: %d", won)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(edgar) })
	g.WithWriteLock(func() { _, _ = g.FlipCoinsForEffect(game.RandomDraw{Player: me.ID}, 40) })
	if _, won, _ := finB13FlipEvents(g, me.ID); won != 0 {
		t.Errorf("flips were won with Edgar gone: %d", won)
	}
}

// --- Kinnan, Bonder Prodigy -----------------------------------------------------------

func TestFinB13KinnanDoublesNonlandManaButNotLandMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Kinnan, Bonder Prodigy", "Legendary Creature — Human Druid", finB13KinnanOracle, 2, 2)
	stadium := b12Push(g, me.ID, "Strixhaven Stadium", "Artifact", finB13StadiumOracle, 0, 0)
	advanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, stadium, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Stadium: %v", err)
	}
	if got := me.ManaPool.CountColor("C"); got != 2 {
		t.Fatalf("%d {C} after tapping an artifact under Kinnan, want 2", got)
	}
}

func TestFinB13KinnanDigsForANonHumanCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	kinnan := b12Push(g, me.ID, "Kinnan, Bonder Prodigy", "Legendary Creature — Human Druid", finB13KinnanOracle, 2, 2)
	advanceToMain(t, g)
	ids := finB13AtraxaLibrary(me, "Creature — Human Soldier", "Creature — Elk", "Sorcery", "Basic Land — Forest", "Creature — Human Wizard")
	floatForTest(g, me, "CCCCCGU")
	if err := g.ActivateCatalogAbility(me.ID, kinnan, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if c == nil {
		t.Fatal("no pick prompt")
	}
	if len(c.ChooseCards) != 1 || c.ChooseCards[0] != ids[1] {
		t.Fatalf("candidates %v, want only the non-Human creature %v", c.ChooseCards, ids[1])
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ids[1]) {
		t.Error("the Elk did not enter the battlefield")
	}
	for i := 0; i < 4; i++ {
		if me.Library.Cards[i].InstanceID == ids[1] {
			t.Error("the Elk is still in the library")
		}
	}
}

// --- The Lunar Whale --------------------------------------------------------------

func TestFinB13LunarWhaleOpensTheTopOfTheLibraryOnlyOnceItHasAttacked(t *testing.T) {
	s := newCrewScene(t)
	whale := s.pushVehicle("The Lunar Whale", finB13WhaleOracle, 3, 5)
	pilot := s.pushCreature("Pilot", "Creature — Human", 1, 1)
	var top uuid.UUID
	s.g.WithWriteLock(func() {
		c := game.NewCard("On Top", s.me.ID)
		c.TypeLine = "Instant"
		s.me.Library.PushTop(c)
		top = c.InstanceID
	})
	granted := func() bool {
		out := false
		s.g.ReadSnapshot(func() {
			if card, ok := s.g.LookupCardForEffect(top); ok {
				out = s.g.CastPermissionForLocked(s.me.ID, card, game.ZoneLibrary).Granted()
			}
		})
		return out
	}
	if got := s.g.LibraryTopVisibilityForEffect(s.me.ID); got != game.LibraryTopOwner {
		t.Errorf("look at the top any time: visibility %v", got)
	}
	if granted() {
		t.Error("the top is playable before the Whale attacked")
	}
	s.crewAndResolve(t, whale, 0, pilot)
	if !hasEffectiveKeyword(t, s.g, whale, "flying") {
		t.Error("the Whale has no flying")
	}
	if granted() {
		t.Error("the top is playable once crewed but before attacking")
	}
	finABAttack(t, s.g, s.opp.ID, whale)
	passPriorityAroundTable(t, s.g)
	if !granted() {
		t.Error("the top is not playable after the Whale attacked")
	}
}

// --- transforming cards ------------------------------------------------------------

const (
	finB14JechtOracle     = "b6cebf8d-37e1-42de-b3f1-1c174ab17ac8"
	finB14CliveOracle     = "ceba3d03-be5a-4dd7-b829-7db216a0f44e"
	finB14SerahOracle     = "a8a7a3b7-015c-4fbd-8b90-e0ddf71663e4"
	finB14LeviathanOracle = "b38d1811-64d7-4a35-a86c-3291eecd8391"
)

func TestFinB14EveryOtherCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		finB13CarpetOracle:       "Carpet of Flowers",
		finB13AtraxaOracle:       "Atraxa, Grand Unifier",
		finB13KinnanOracle:       "Kinnan, Bonder Prodigy",
		finB13VialSmasherOracle:  "Vial Smasher the Fierce",
		finB13StadiumOracle:      "Strixhaven Stadium",
		finB13EdgarOracle:        "Edgar, King of Figaro",
		finB13WhaleOracle:        "The Lunar Whale",
		finB14JechtOracle:        "Jecht, Reluctant Guardian",
		finB14JechtOracle + "#1": "Braska's Final Aeon",
		finB14CliveOracle:        "Clive, Ifrit's Dominant",
		finB14CliveOracle + "#1": "Ifrit, Warden of Inferno",
		finB14SerahOracle:        "Serah Farron",
		finB14SerahOracle + "#1": "Crystallized Serah",
		finB14LeviathanOracle:    "Summon: Leviathan",
		finB14RaubahnOracle:      "Raubahn, Bull of Ala Mhigo",
		finB14GenjiOracle:        "Genji Glove",
		finB14HideawayOracle:     "Clive's Hideaway",
		finB14FirionOracle:       "Firion, Wild Rose Warrior",
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

func finB14Row(oracle, front, frontType, frontCost, back, backType, power, toughness string, colors []string, fp, ft string) cards.Card {
	row := transformRow(oracle, front, frontType, frontCost, back, backType, power, toughness, colors)
	row.CardFaces[0].Power, row.CardFaces[0].Toughness = fp, ft
	return row
}

// --- Jecht, Reluctant Guardian // Braska's Final Aeon ---------------------------

func finB14JechtRow() cards.Card {
	return finB14Row(finB14JechtOracle, "Jecht, Reluctant Guardian", "Legendary Creature — Human Warrior", "{3}{B}",
		"Braska's Final Aeon", "Legendary Enchantment Creature — Saga Nightmare", "7", "7", []string{"B"}, "4", "3")
}

func TestFinB14JechtTransformsOnCombatDamageAndTheAeonMakesThemDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	jecht := importToBattlefield(t, g, finB14JechtRow(), me)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == jecht {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	pushCatalogHandCard(opp, "Discard Fodder", "Sorcery", "")
	hand := me.Hand.Size()
	declareAttack(t, g, opp.ID, jecht)
	for i := 0; i < 12 && latestConfirmFor(g, me.ID) == nil; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			passPriorityAroundTable(t, g)
		}
	}
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	var aeon uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.OracleID == finB14JechtOracle && TransformedPermanent(c) {
				aeon = c.InstanceID
			}
		}
	})
	if aeon == uuid.Nil {
		t.Fatal("Jecht did not come back transformed")
	}
	c := e2Card(t, g, aeon)
	if c.Name != "Braska's Final Aeon" || !game.IsSaga(c) || c.Counters[game.CounterLore] != 1 {
		t.Fatalf("back face %q saga=%v lore=%d, want the Aeon as a Saga on chapter I", c.Name, game.IsSaga(c), c.Counters[game.CounterLore])
	}
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, opp.ID, opp.Hand.Cards[0].InstanceID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("chapter I drew %d cards for you, want 1", got)
	}
}

func TestFinB14AeonChapterThreeMakesEachOpponentSacrificeTwoCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	mine := ctrlPushCreature(g, me.ID, "My Bear")
	a := ctrlPushCreature(g, opp.ID, "Their A")
	b := ctrlPushCreature(g, opp.ID, "Their B")
	c := ctrlPushCreature(g, opp.ID, "Their C")
	aeon := importToBattlefield(t, g, finB14JechtRow(), me)
	g.WithWriteLock(func() { _ = g.TransformPermanentForEffect(aeon) })
	// Chapters I and II (Jecht Beam) first, one at a time, with nothing in the
	// opponents' hands to discard.
	for _, p := range g.Seats[1:] {
		p.Hand.Cards = nil
	}
	for i := 0; i < 3; i++ {
		g.WithWriteLock(func() { _ = g.AddCounterForEffect(aeon, game.CounterLore, 1) })
		passPriorityAroundTable(t, g)
	}
	pc := sacrificeChoiceFor(g, opp.ID)
	if pc == nil {
		t.Fatal("chapter III asked the opponent for nothing")
	}
	if pc.ChooseMin != 2 || pc.ChooseMax != 2 {
		t.Fatalf("pick bounds %d..%d, want exactly two", pc.ChooseMin, pc.ChooseMax)
	}
	answerSacrifice(t, g, opp.ID, a, b)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if g.Battlefield.Contains(id) {
			t.Errorf("a chosen creature was not sacrificed")
		}
	}
	if !g.Battlefield.Contains(c) || !g.Battlefield.Contains(mine) {
		t.Error("an unchosen creature, or your own, was sacrificed")
	}
}

// --- Clive, Ifrit's Dominant // Ifrit, Warden of Inferno -----------------------------

func finB14CliveRow() cards.Card {
	row := finB14Row(finB14CliveOracle, "Clive, Ifrit's Dominant", "Legendary Creature — Human Noble Warrior", "{4}{R}{R}",
		"Ifrit, Warden of Inferno", "Legendary Enchantment Creature — Saga Demon", "9", "9", []string{"R"}, "5", "5")
	return row
}

func TestFinB14CliveDiscardsYourHandAndDrawsYourDevotionToRed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	// Two red permanents: {R}{R} and {1}{R}, so devotion 5 with Clive's own {R}{R}.
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Red A", TypeLine: "Creature — Goblin",
		ManaCost: "{R}{R}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Red B", TypeLine: "Creature — Goblin",
		ManaCost: "{1}{R}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	for i := 0; i < 3; i++ {
		pushCatalogHandCard(me, "Old Card", "Sorcery", "")
	}
	importAndCast(t, g, finB14CliveRow(), me)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 5 {
		t.Errorf("hand is %d, want 5 (discarded the hand, drew devotion 5)", got)
	}
	if me.Graveyard.Size() < 3 {
		t.Errorf("graveyard has %d cards, want the three discarded", me.Graveyard.Size())
	}
}

func TestFinB14CliveMayDeclineTheDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pushCatalogHandCard(me, "Old Card", "Sorcery", "")
	hand := me.Hand.Size()
	importAndCast(t, g, finB14CliveRow(), me)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d after declining, want unchanged", hand, me.Hand.Size())
	}
}

func TestFinB14CliveTransformsAtSorcerySpeedAndIfritFightsThenReturnsOnChapterThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	clive := importAndCast(t, g, finB14CliveRow(), me)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	target := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Bear",
		TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == clive {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	floatForTest(g, me, "CCCCRR")
	if err := g.ActivateCatalogAbility(me.ID, clive, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("transform ability: %v (step %v, stack %d, pending %d)", err, g.Turn.Step, len(g.StackMeta), len(g.PendingChoices))
	}
	passPriorityAroundTable(t, g)
	var ifrit uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.OracleID == finB14CliveOracle && TransformedPermanent(c) {
				ifrit = c.InstanceID
			}
		}
	})
	if ifrit == uuid.Nil {
		t.Fatal("Clive did not return transformed")
	}
	if c := e2Card(t, g, ifrit); c.Name != "Ifrit, Warden of Inferno" || c.Counters[game.CounterLore] != 1 {
		t.Fatalf("back face %q with %d lore counters, want Ifrit on chapter I", c.Name, c.Counters[game.CounterLore])
	}
	// Chapter I: Ifrit fights up to one other target creature.
	pickTriggerTarget(t, g, me.ID, target)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(target) {
		t.Error("Ifrit's fight left the 2/2 alive")
	}
	// Chapter II: {R}{R}{R}{R}, and Ifrit stays.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ifrit, game.CounterLore, 1) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := me.ManaPool.CountColor("R"); got != 4 {
		t.Fatalf("%d {R} after chapter II, want 4", got)
	}
	if c := e2Card(t, g, ifrit); !TransformedPermanent(c) {
		t.Fatal("Ifrit left after chapter II")
	}
	// Chapter III: {R}{R}{R}{R} again and Ifrit returns front face up.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ifrit, game.CounterLore, 1) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := me.ManaPool.CountColor("R"); got != 8 {
		t.Errorf("%d {R} after chapter III, want 8", got)
	}
	var back uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.OracleID == finB14CliveOracle {
				back = c.InstanceID
				if TransformedPermanent(c) || game.IsSaga(c) {
					t.Errorf("after chapter III the card is %q (saga=%v), want Clive front face up", c.Name, game.IsSaga(c))
				}
			}
		}
	})
	if back == uuid.Nil {
		t.Error("the card is not on the battlefield after chapter III")
	}
}

// --- Serah Farron // Crystallized Serah -------------------------------------------------

func finB14SerahRow() cards.Card {
	return finB14Row(finB14SerahOracle, "Serah Farron", "Legendary Creature — Human Citizen", "{1}{G}{W}",
		"Crystallized Serah", "Legendary Artifact", "", "", []string{"G", "W"}, "2", "2")
}

func TestFinB14SerahDiscountsOnlyTheFirstLegendaryCreatureSpellByTwoGenericMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	importToBattlefield(t, g, finB14SerahRow(), me)
	cast := func(name, typeLine, cost string) error {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Power: 1, Toughness: 1,
			Owner: me.ID, Controller: me.ID})
		return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
	}
	floatForTest(g, me, "CC")
	if err := cast("Legend One", "Legendary Creature — Human", "{3}{R}"); err == nil {
		t.Fatal("a {3}{R} spell was cast with only two mana")
	}
	floatForTest(g, me, "CR")
	if err := cast("Legend One", "Legendary Creature — Human", "{3}{R}"); err != nil {
		t.Fatalf("the first legendary creature spell should cost {2} less: %v", err)
	}
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	floatForTest(g, me, "CC")
	if err := cast("Legend Two", "Legendary Creature — Human", "{3}"); err == nil {
		t.Error("the second legendary creature spell this turn was discounted")
	}
}

func TestFinB14SerahDoesNotDiscountANonlegendaryCreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	importToBattlefield(t, g, finB14SerahRow(), me)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Plain Bear", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Power: 1, Toughness: 1,
		Owner: me.ID, Controller: me.ID})
	floatForTest(g, me, "G")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a nonlegendary creature spell was discounted")
	}
}

func TestFinB14SerahTransformsAtCombatWithTwoOtherLegendsAndTheBackFaceIsAnAnthem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	serah := importToBattlefield(t, g, finB14SerahRow(), me)
	a := b12Push(g, me.ID, "Legend A", "Legendary Creature — Human", "", 2, 2)
	plain := pushVanillaCreature(g, me.ID, "Plain", 2, 2)
	b12Push(g, me.ID, "Legend B", "Legendary Creature — Human", "", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	c := e2Card(t, g, serah)
	if !TransformedPermanent(c) || c.Name != "Crystallized Serah" {
		t.Fatalf("Serah is %q (transformed=%v), want Crystallized Serah", c.Name, TransformedPermanent(c))
	}
	if p := effectivePower(t, g, a); p != 4 {
		t.Errorf("a legendary creature has power %d, want 4 (+2/+2)", p)
	}
	if p := effectivePower(t, g, plain); p != 2 {
		t.Errorf("a nonlegendary creature has power %d, want 2", p)
	}
}

func TestFinB14SerahWithOnlyOneOtherLegendDoesNotTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	serah := importToBattlefield(t, g, finB14SerahRow(), me)
	b12Push(g, me.ID, "Legend A", "Legendary Creature — Human", "", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if TransformedPermanent(e2Card(t, g, serah)) {
		t.Error("Serah transformed with only one other legendary creature")
	}
}

// --- Summon: Leviathan ------------------------------------------------------------------

func TestFinB14SummonLeviathanChapterOneBouncesEverythingButTheSeaCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	merfolk := b12Creature(g, opp.ID, "Their Merfolk", "Creature — Merfolk Wizard", 1, 1)
	serpent := b12Creature(g, me.ID, "My Serpent", "Creature — Serpent", 3, 3)
	castCatalogSpell(t, g, "Summon: Leviathan", "Enchantment Creature — Saga Leviathan", finB14LeviathanOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	for _, c := range []struct {
		id   uuid.UUID
		p    *game.Player
		name string
	}{{bear, me, "my Bear"}, {theirs, opp, "their Bear"}} {
		if g.Battlefield.Contains(c.id) || !c.p.Hand.Contains(c.id) {
			t.Errorf("%s was not returned to its owner's hand", c.name)
		}
	}
	if !g.Battlefield.Contains(merfolk) || !g.Battlefield.Contains(serpent) {
		t.Error("a Merfolk or a Serpent was bounced")
	}
}

func TestFinB14SummonLeviathanChapterTwoDrawsWheneverASeaCreatureAttacksThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	sea := b12Creature(g, me.ID, "My Octopus", "Creature — Octopus", 2, 2)
	land := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	whale := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Summon: Leviathan", TypeLine: "Enchantment Creature — Saga Leviathan",
		OracleID: finB14LeviathanOracle, Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterLore: 1},
	})
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(whale, game.CounterLore, 1) })
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	declareAttack(t, g, opp.ID, sea, land)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards for two attackers of which one was an Octopus, want 1", got)
	}
}
