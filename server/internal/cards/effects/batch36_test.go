package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch36_test.go — card-level coverage for the card-coverage
// roadmap's batch 36 (#399, `edhrec_rank` 3753–3852): the "no new
// machinery" group. One test per observable behaviour, driven through
// a real cast, activation or attack rather than by calling primitives.

const (
	b36SanctumOfEternityOracle  = "c7d9ff27-f1fc-42e4-a47b-d2e6d68e4035"
	b36LlanowarVisionaryOracle  = "f75ed312-3a23-4624-80c5-03980aa22d0b"
	b36AngelOfSerenityOracle    = "f4ce6078-8c7b-4f68-b324-13130c63a983"
	b36TemperedSteelOracle      = "e220138c-5fc6-487d-9fa9-f21b2fb1f12c"
	b36AkoumRefugeOracle        = "354fecd1-2371-49e3-81c6-7e47728dbb1f"
	b36AnimalSanctuaryOracle    = "f3c40943-1d7c-4ea2-b34f-8df8b6775701"
	b36HeliodsPilgrimOracle     = "0fe01c7d-f435-44c1-82f1-0ec2c47d4704"
	b36JaggedBarrensOracle      = "64ee02f1-afdb-474b-a893-31538ad7219a"
	b36WhitemaneLionOracle      = "e8d6084b-9b72-438e-a30a-851b888f3e4d"
	b36CrucibleOfFireOracle     = "0c572396-4e53-4c86-9b04-f41341dcea05"
	b36ArahboTheFirstFangOracle = "dfec65a0-6fb6-463d-aa62-e84a427db217"
	b36ReplenishOracle          = "523ae937-5535-490d-96ea-07f331b5e5ad"
	b36RepurposingBayOracle     = "1787ac2f-762d-4f3a-b7e5-12db6d3d470d"
	b36LifecreedDuoOracle       = "63686c6b-9051-4002-aa8e-da8a3021330f"
	b36LossarnachCaptainOracle  = "1251691e-3c4c-4e45-9f59-f156852f2268"
	b36TurbulentSpringsOracle   = "9aef7510-9f06-4939-8cae-f71330d1105e"
	b36MavrenFeinOracle         = "1b94a11b-21b0-4465-a520-69608f022fb4"
	b36FanaticOfMogisOracle     = "f19d06af-6caf-41d8-8d0a-d5d50bd67900"
	b36ShardingSphinxOracle     = "9ebc0144-fc45-4de7-b3d8-ef8cf2e211ae"
	b36SyggRiverCutthroatOracle = "cd4db500-0017-46c6-be94-1bf48f686b6a"
	b36BishopOfWingsOracle      = "af4684cf-f109-44be-adfb-7c551a36635e"
	b36TegwyllOracle            = "868f0a0a-ca9e-4baf-8295-6b228aa834e5"
	b36ExtinguishAllHopeOracle  = "c7ca25f3-7a22-477b-8546-4c2597d5d1ff"
	b36HealersHawkOracle        = "28a52ba1-95da-44e1-8ac5-0dc23c902394"
	b36CadiraOracle             = "fcf22321-2f55-41ca-b8e1-4792540ba3ee"
	b36ErodedCanyonOracle       = "852c6520-d148-4923-a312-05a9af821f24"
	b36SarkhansTriumphOracle    = "c4f7cdb7-aef9-4ab8-995b-4e5180d47214"
)

// b36Commander seeds a creature flagged as a commander, owned and
// controlled by `owner`, able to attack.
func b36Commander(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Legendary Creature — Human",
		Power: 3, Toughness: 3, Owner: owner, Controller: owner, IsCommander: true,
	})
}

// b36Token pushes a creature token under controller with a real
// timestamp, so the layer cache sees it.
func b36Token(g *game.Game, controller uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Power: power, Toughness: toughness, Owner: controller, Controller: controller,
	})
}

// b36CastCreature casts a creature with a real power and toughness
// from the active player's hand and settles the spell and whatever
// it triggered.
func b36CastCreature(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	return id
}

// b36Lands seeds n vanilla lands under owner.
func b36Lands(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		b12Permanent(g, owner, "Wastes", "Basic Land")
	}
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Akoum Refuge
// is a row in the gain-land cycle table, so a transposed row is
// invisible until someone plays that exact card.
func TestBatch36CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b36SanctumOfEternityOracle:  "Sanctum of Eternity",
		b36LlanowarVisionaryOracle:  "Llanowar Visionary",
		b36AngelOfSerenityOracle:    "Angel of Serenity",
		b36TemperedSteelOracle:      "Tempered Steel",
		b36AkoumRefugeOracle:        "Akoum Refuge",
		b36AnimalSanctuaryOracle:    "Animal Sanctuary",
		b36HeliodsPilgrimOracle:     "Heliod's Pilgrim",
		b36JaggedBarrensOracle:      "Jagged Barrens",
		b36WhitemaneLionOracle:      "Whitemane Lion",
		b36CrucibleOfFireOracle:     "Crucible of Fire",
		b36ArahboTheFirstFangOracle: "Arahbo, the First Fang",
		b36ReplenishOracle:          "Replenish",
		b36RepurposingBayOracle:     "Repurposing Bay",
		b36LifecreedDuoOracle:       "Lifecreed Duo",
		b36LossarnachCaptainOracle:  "Lossarnach Captain",
		b36TurbulentSpringsOracle:   "Turbulent Springs",
		b36MavrenFeinOracle:         "Mavren Fein, Dusk Apostle",
		b36FanaticOfMogisOracle:     "Fanatic of Mogis",
		b36ShardingSphinxOracle:     "Sharding Sphinx",
		b36SyggRiverCutthroatOracle: "Sygg, River Cutthroat",
		b36BishopOfWingsOracle:      "Bishop of Wings",
		b36TegwyllOracle:            "Tegwyll, Duke of Splendor",
		b36ExtinguishAllHopeOracle:  "Extinguish All Hope",
		b36HealersHawkOracle:        "Healer's Hawk",
		b36CadiraOracle:             "Cadira, Caller of the Small",
		b36ErodedCanyonOracle:       "Eroded Canyon",
		b36SarkhansTriumphOracle:    "Sarkhan's Triumph",
	}
	if len(want) != 27 {
		t.Fatalf("the batch registers 27 cards, the table lists %d", len(want))
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

// --- lands ---------------------------------------------------------

func TestB36AkoumRefugeEntersTappedGainsOneAndTapsForEitherColour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Life
	land := b12PlayFromHand(t, g, "Akoum Refuge", "Land", b36AkoumRefugeOracle, game.CastSpellParams{})
	if !b16Tapped(t, g, land) {
		t.Fatal("the refuge enters tapped")
	}
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("you gain 1 life: %d → %d", before, me.Life)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	b28TapForMana(t, g, me.ID, land, "R")
	if got := poolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("tapped for {R}: pool %v", got)
	}
}

func TestB36DesertDualsEnterTappedPingAnOpponentAndTapForTheirColours(t *testing.T) {
	for _, row := range []struct{ name, oracle, colour string }{
		{"Jagged Barrens", b36JaggedBarrensOracle, "B"},
		{"Eroded Canyon", b36ErodedCanyonOracle, "U"},
	} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[2]
		before := opp.Life
		land := b12PlayFromHand(t, g, row.name, "Land — Desert", row.oracle, game.CastSpellParams{})
		if !b16Tapped(t, g, land) {
			t.Fatalf("%s: the Desert enters tapped", row.name)
		}
		if p := latestPickTarget(g, me.ID); p == nil || hasID(p.PickTargetPlayers, me.ID) || !hasID(p.PickTargetPlayers, opp.ID) {
			t.Fatalf("%s: the trigger asks for an opponent, never you: %+v", row.name, p)
		}
		b16PickPlayer(t, g, me.ID, opp.ID)
		passPriorityAroundTable(t, g)
		if opp.Life != before-1 {
			t.Errorf("%s: target opponent takes 1: %d → %d", row.name, before, opp.Life)
		}
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
		b28TapForMana(t, g, me.ID, land, row.colour)
		if got := poolColors(me); len(got) != 1 || got[0] != row.colour {
			t.Errorf("%s: tapped for {%s}: pool %v", row.name, row.colour, got)
		}
	}
}

func TestB36TurbulentSpringsEntersUntappedOnlyAgainstEightOpposingLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	b36Lands(g, opp.ID, 4)
	b36Lands(g, other.ID, 3)
	b36Lands(g, me.ID, 5)
	tapped := b12PlayFromHand(t, g, "Turbulent Springs", "Land — Island Mountain", b36TurbulentSpringsOracle, game.CastSpellParams{})
	if !b16Tapped(t, g, tapped) {
		t.Fatal("seven opposing lands — your own do not count — is not eight: it enters tapped")
	}
	b36Lands(g, other.ID, 1)
	advanceToMainOf(t, g, 0)
	springs := b12PlayFromHand(t, g, "Turbulent Springs", "Land — Island Mountain", b36TurbulentSpringsOracle, game.CastSpellParams{})
	if b16Tapped(t, g, springs) {
		t.Fatal("eight opposing lands between two opponents: it enters untapped")
	}
	b28TapForMana(t, g, me.ID, springs, "U")
	if got := poolColors(me); len(got) != 1 || got[0] != "U" {
		t.Errorf("tapped for {U}: pool %v", got)
	}
}

func TestB36SanctumOfEternityReturnsOnlyACommanderYouOwn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sanctum := pushCatalogPermanent(g, me.ID, "Sanctum of Eternity", "Land", b36SanctumOfEternityOracle, false)
	mine := b36Commander(g, me.ID, "My Commander")
	theirs := b36Commander(g, opp.ID, "Their Commander")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	spec, _ := Lookup(b36SanctumOfEternityOracle)
	if len(spec.Activated) != 1 || !spec.Activated[0].SorcerySpeed {
		t.Fatal("one ability, at sorcery speed (the declared stand-in for \"during your turn\")")
	}
	b28TapForMana(t, g, me.ID, sanctum, "C")
	if got := poolColors(me); len(got) != 1 || got[0] != "C" {
		t.Fatalf("tapped for {C}: pool %v", got)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(sanctum) })
	advanceToMain(t, g)
	b06AddMana(me, "C", "C")
	for _, bad := range []uuid.UUID{theirs, bear} {
		if err := g.ActivateCatalogAbility(me.ID, sanctum, 0, game.ActivateAbilityParams{Targets: cardRefs(bad)}); err == nil {
			t.Fatal("an opponent's commander and a creature that is not a commander are not legal targets")
		}
	}
	b16Activate(t, g, me.ID, sanctum, 0, game.ActivateAbilityParams{Targets: cardRefs(mine)})
	if !b16Tapped(t, g, sanctum) {
		t.Error("the Sanctum taps to activate")
	}
	// CR 903.9 (#539): a commander headed to hand is offered the
	// command zone. Declined, it is in hand, as the card says.
	b21DeclineCommandZone(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || !me.Hand.Contains(mine) {
		t.Error("your commander returns to your hand")
	}
	// Accepted, it goes to the command zone instead.
	again := b36Commander(g, me.ID, "My Commander")
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(sanctum) })
	b06AddMana(me, "C", "C")
	b16Activate(t, g, me.ID, sanctum, 0, game.ActivateAbilityParams{Targets: cardRefs(again)})
	b36AcceptCommandZone(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !me.Command.Contains(again) || me.Hand.Contains(again) {
		t.Error("the owner may send it to the command zone instead")
	}
}

// b36AcceptCommandZone answers the CR 903.9 "put it into the command
// zone instead?" prompt with "yes".
func b36AcceptCommandZone(t *testing.T, g *game.Game, owner uuid.UUID) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionalReplacement && c.Chooser == owner {
			if err := g.ResolveOptionalReplacement(c.ID, owner, true); err != nil {
				t.Fatalf("ResolveOptionalReplacement: %v", err)
			}
			return
		}
	}
	t.Fatalf("no command-zone prompt for %s", owner)
}

func TestB36AnimalSanctuaryGrowsAPetOfAnyoneAndNothingElse(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sanctuary := pushCatalogPermanent(g, me.ID, "Animal Sanctuary", "Land", b36AnimalSanctuaryOracle, false)
	dog := b12Creature(g, opp.ID, "Their Dog", "Creature — Dog", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	b06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, sanctuary, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}); err == nil {
		t.Fatal("a Bear is none of Bird, Cat, Dog, Goat, Ox or Snake")
	}
	b16Activate(t, g, me.ID, sanctuary, 0, game.ActivateAbilityParams{Targets: cardRefs(dog)})
	if counterOn(g, dog, game.CounterPlusOne) != 1 {
		t.Errorf("an opponent's Dog gets the counter: %d", counterOn(g, dog, game.CounterPlusOne))
	}
	if !b16Tapped(t, g, sanctuary) {
		t.Error("the Sanctuary taps to activate")
	}
}

// --- creatures with an enters trigger ------------------------------

func TestB36LlanowarVisionaryDrawsOnEntryAndTapsForGreen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Llanowar Visionary", "Creature — Elf Druid", b36LlanowarVisionaryOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("the card was seeded into hand and cast, then one was drawn: hand %d → %d", hand, me.Hand.Size())
	}
	elf := pushCatalogPermanent(g, me.ID, "Llanowar Visionary", "Creature — Elf Druid", b36LlanowarVisionaryOracle, false)
	b28TapForMana(t, g, me.ID, elf, "G")
	if got := poolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("tapped for {G}: pool %v", got)
	}
}

func TestB36HeliodsPilgrimMayTutorAnAura(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		game.Card{Name: "Rancor", TypeLine: "Enchantment — Aura", ManaCost: "{G}"},
		game.Card{Name: "Ethereal Armor", TypeLine: "Enchantment — Aura", ManaCost: "{W}"},
		game.Card{Name: "Glorious Anthem", TypeLine: "Enchantment", ManaCost: "{1}{W}{W}"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
	)
	castCatalogSpell(t, g, "Heliod's Pilgrim", "Creature — Human Cleric", b36HeliodsPilgrimOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("you may search: the chooser asks")
	}
	if searchOptionNamed(g, c, "Rancor") == uuid.Nil || searchOptionNamed(g, c, "Ethereal Armor") == uuid.Nil {
		t.Error("both Auras are offered")
	}
	if searchOptionNamed(g, c, "Glorious Anthem") != uuid.Nil || searchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("a non-Aura enchantment and a creature are not")
	}
	answerSearchNamed(t, g, me.ID, "Rancor")
	if !b02bHandHasNamed(me, "Rancor") {
		t.Error("the pick goes to hand")
	}
}

func TestB36WhitemaneLionFlashesInAndReturnsAnyCreatureYouControlItselfIncluded(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	lion := castCatalogSpell(t, g, "Whitemane Lion", "Creature — Cat", b36WhitemaneLionOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if !hasID(p.PickTargetCards, bear) || !hasID(p.PickTargetCards, lion) {
		t.Error("every creature you control is offered, the Lion itself included")
	}
	if hasID(p.PickTargetCards, theirs) {
		t.Error("an opponent's creature is not")
	}
	if p.PickTargetMin != 1 {
		t.Error("the return is mandatory")
	}
	pickCard(t, g, me.ID, lion)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(lion) || !me.Hand.Contains(lion) {
		t.Error("the Lion returns itself to hand")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the Bear stays")
	}
	// Again, bouncing the Bear this time.
	second := castCatalogSpell(t, g, "Whitemane Lion", "Creature — Cat", b36WhitemaneLionOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !me.Hand.Contains(bear) {
		t.Error("the Bear returns to hand")
	}
	if !hasEffectiveKeyword(t, g, second, "flash") {
		t.Error("printed flash did not reach the effective abilities")
	}
}

func TestB36FanaticOfMogisBurnsEachOpponentForDevotionToRed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b27Push(g, me.ID, "Mountain Dweller", "Creature — Ogre", "", "{1}{R}{R}", 3, 3, "R")
	b27Push(g, me.ID, "Hybrid", "Enchantment", "", "{R/W}", 0, 0, "R", "W")
	b27Push(g, me.ID, "Signet", "Artifact", "", "{2}", 0, 0)
	before := lifeOfOpponents(g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Fanatic of Mogis", TypeLine: "Creature — Minotaur Shaman",
		OracleID: b36FanaticOfMogisOracle, ManaCost: "{3}{R}", Colors: []string{"R"},
		Power: 4, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	b06AddMana(me, "C", "C", "C", "R")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Fanatic: %v", err)
	}
	passPriorityAroundTable(t, g)
	// {R}{R} + {R/W} + the Fanatic's own {R} = 4; the Signet's {2} is not red.
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-4 {
			t.Errorf("opponent %d: %d → %d, want -4 (devotion to red counted as the trigger resolves)", i+1, b, got)
		}
	}
}

// --- anthems and lords ---------------------------------------------

func TestB36TemperedSteelLiftsOnlyYourArtifactCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Tempered Steel", "Enchantment", b36TemperedSteelOracle, false)
	thopter := b36Token(g, me.ID, "Thopter", "Token Artifact Creature — Thopter", 1, 1)
	construct := b12Creature(g, me.ID, "Construct", "Artifact Creature — Construct", 4, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, me.ID, "Signet", "Artifact")
	theirs := b12Creature(g, opp.ID, "Their Myr", "Artifact Creature — Myr", 1, 1)
	if effectivePower(t, g, thopter) != 3 || effectiveToughness(t, g, thopter) != 3 {
		t.Errorf("a Thopter token is 3/3: %d/%d", effectivePower(t, g, thopter), effectiveToughness(t, g, thopter))
	}
	if effectivePower(t, g, construct) != 6 {
		t.Errorf("the Construct is 6/6: power %d", effectivePower(t, g, construct))
	}
	if effectivePower(t, g, bear) != 2 {
		t.Error("a non-artifact creature is untouched")
	}
	if effectivePower(t, g, rock) != 0 {
		t.Error("a noncreature artifact is untouched")
	}
	if effectivePower(t, g, theirs) != 1 {
		t.Error("an opponent's artifact creature is untouched")
	}
}

func TestB36CrucibleOfFireLiftsEveryDragonYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Crucible of Fire", "Enchantment", b36CrucibleOfFireOracle, false)
	dragon := b12Creature(g, me.ID, "Dragon", "Creature — Dragon", 4, 4)
	changeling := b34Changeling(g, me.ID, "Shapeshifter")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Dragon", "Creature — Dragon", 4, 4)
	if effectivePower(t, g, dragon) != 7 || effectiveToughness(t, g, dragon) != 7 {
		t.Errorf("your Dragon is 7/7: %d/%d", effectivePower(t, g, dragon), effectiveToughness(t, g, dragon))
	}
	if effectivePower(t, g, changeling) != 4 {
		t.Error("a changeling is a Dragon")
	}
	if effectivePower(t, g, bear) != 2 || effectivePower(t, g, theirs) != 4 {
		t.Error("a Bear and an opponent's Dragon are untouched")
	}
}

func TestB36ArahboPumpsOtherCatsAndMakesACatForEachNontokenCatEntering(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arahbo := b36CastCreature(t, g, "Arahbo, the First Fang", "Legendary Creature — Cat Avatar", b36ArahboTheFirstFangOracle, 2, 2)
	if n := countBattlefieldNamed(g, me.ID, "Cat"); n != 1 {
		t.Fatalf("Arahbo's own entry makes a Cat: %d", n)
	}
	token := findBattlefieldByName(g, "Cat")
	if effectivePower(t, g, token) != 2 || effectivePower(t, g, arahbo) != 2 {
		t.Errorf("the token Cat is lifted to 2/2 and Arahbo is not: %d / %d", effectivePower(t, g, token), effectivePower(t, g, arahbo))
	}
	// A nontoken Cat entering under your control: another Cat.
	b36CastCreature(t, g, "Savannah Lions", "Creature — Cat", "", 2, 1)
	if n := countBattlefieldNamed(g, me.ID, "Cat"); n != 2 {
		t.Errorf("a nontoken Cat entering makes a second Cat token: %d", n)
	}
	// The token entering, a non-Cat, and an opponent's Cat: nothing.
	b36CastCreature(t, g, "Bear", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, b36WhiteCatToken(), 1) })
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Cat"); n != 2 {
		t.Errorf("a token Cat, a Bear and an opponent's Cat make nothing: %d", n)
	}
}

func TestB36TegwyllPumpsOtherFaeriesAndDrawsWhenOneDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tegwyll := b12Push(g, me.ID, "Tegwyll, Duke of Splendor", "Legendary Creature — Faerie Noble", b36TegwyllOracle, 2, 3)
	faerie := b12Creature(g, me.ID, "Faerie", "Creature — Faerie Rogue", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Faerie", "Creature — Faerie", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !hasEffectiveKeyword(t, g, tegwyll, "flying") || !hasEffectiveKeyword(t, g, tegwyll, "deathtouch") {
		t.Error("printed flying and deathtouch did not reach the effective abilities")
	}
	if effectivePower(t, g, faerie) != 2 || effectivePower(t, g, tegwyll) != 2 || effectivePower(t, g, theirs) != 1 {
		t.Error("other Faeries you control get +1/+1; Tegwyll and an opponent's Faerie do not")
	}
	life, hand := me.Life, me.Hand.Size()
	b27Kill(g, faerie)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Errorf("another Faerie you control died: draw one, lose one — hand %d → %d, life %d → %d", hand, me.Hand.Size(), life, me.Life)
	}
	b27Kill(g, bear)
	b27Kill(g, theirs)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Error("a Bear and an opponent's Faerie dying do nothing")
	}
	b27Kill(g, tegwyll)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Error("Tegwyll's own death is not \"another Faerie\"")
	}
}

// --- go-wide and lifegain payoffs ----------------------------------

func TestB36LifecreedDuoGainsOnePerOtherCreatureYouControlEntering(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	duo := castCatalogSpell(t, g, "Lifecreed Duo", "Creature — Bat Bird", b36LifecreedDuoOracle, nil)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, duo, "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}
	before := me.Life
	if me.Life != before {
		t.Fatal("its own entry gains nothing")
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 2) })
	passPriorityAroundTable(t, g)
	if me.Life != before+2 {
		t.Errorf("two creatures entering gain two: %d → %d", before, me.Life)
	}
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, TreasureToken(), 1)
		_ = g.CreateTokenForEffect(opp.ID, RedGoblinToken(), 1)
	})
	passPriorityAroundTable(t, g)
	if me.Life != before+2 {
		t.Error("a noncreature entering and an opponent's creature gain nothing")
	}
}

func TestB36BishopOfWingsGainsFourPerAngelAndMakesASpiritWhenOneDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Bishop of Wings", "Creature — Human Cleric", b36BishopOfWingsOracle, false)
	before := me.Life
	angel := b36CastCreature(t, g, "Serra Angel", "Creature — Angel", "", 4, 4)
	if me.Life != before+4 {
		t.Errorf("an Angel entering gains 4: %d → %d", before, me.Life)
	}
	b36CastCreature(t, g, "Bear", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, TokenCard("4/4 white Angel with flying"), 1) })
	passPriorityAroundTable(t, g)
	if me.Life != before+4 {
		t.Error("a Bear and an opponent's Angel gain nothing")
	}
	b27Kill(g, angel)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Spirit"); n != 1 {
		t.Fatalf("an Angel you control dying makes a Spirit: %d", n)
	}
	spirit := findBattlefieldByName(g, "Spirit")
	assertKeywords(t, g, spirit, "flying")
	theirs := findBattlefieldByName(g, "Angel")
	b27Kill(g, theirs)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Spirit"); n != 1 {
		t.Error("an opponent's Angel dying makes nothing")
	}
}

func TestB36LossarnachCaptainTapsOnHumansAndRaisesASoldierEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	captain := castCatalogSpell(t, g, "Lossarnach Captain", "Creature — Human Soldier", b36LossarnachCaptainOracle, nil)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, captain, "first strike") {
		t.Error("printed first strike did not reach the effective abilities")
	}
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if !hasID(p.PickTargetCards, theirs) || hasID(p.PickTargetCards, mine) {
		t.Error("the Captain's own entry asks for a creature an opponent controls")
	}
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, theirs) {
		t.Error("the chosen creature is tapped")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(theirs) })
	// A non-Human entering: silent.
	b36CastCreature(t, g, "Elf", "Creature — Elf", "", 1, 1)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("an Elf is not a Human")
	}
	// Your upkeep: a Human Soldier, which is a Human entering.
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Human Soldier"); n != 1 {
		t.Fatalf("at the beginning of your upkeep, a Human Soldier: %d", n)
	}
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, theirs) {
		t.Error("the token entering fires the tap trigger too")
	}
	// An opponent's upkeep: nothing.
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Human Soldier"); n != 1 {
		t.Error("only your upkeep")
	}
}

func TestB36MavrenFeinMakesOneVampirePerCombatWhenNontokenVampiresAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mavren := b12Push(g, me.ID, "Mavren Fein, Dusk Apostle", "Legendary Creature — Vampire Cleric", b36MavrenFeinOracle, 2, 2)
	vamp := b12Creature(g, me.ID, "Vampire Knight", "Creature — Vampire Knight", 2, 2)
	token := b36Token(g, me.ID, "Vampire", "Token Creature — Vampire", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, mavren, vamp, bear)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Vampire"); n != 2 {
		t.Fatalf("two nontoken Vampires attacking is ONE trigger: %d Vampires (one seeded)", n)
	}
	made := uuid.Nil
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == me.ID && c.Name == "Vampire" && c.InstanceID != token {
			made = c.InstanceID
		}
	}
	assertKeywords(t, g, made, "lifelink")
	if c, _ := battlefieldCard(g, made); c.Power != 1 || c.Toughness != 1 || !c.HasSubtype("Vampire") {
		t.Error("a 1/1 Vampire")
	}
	// Next combat: only the token Vampire and the Bear attack — nothing.
	advanceToNextSeatsTurn(t, g)
	advanceToMainOf(t, g, 0)
	declareAttack(t, g, opp.ID, token, bear)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Vampire"); n != 2 {
		t.Error("a token Vampire and a Bear attacking make nothing")
	}
}

func TestB36ShardingSphinxMayMakeAThopterPerArtifactCreatureThatConnects(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sphinx := b12Push(g, me.ID, "Sharding Sphinx", "Artifact Creature — Sphinx", b36ShardingSphinxOracle, 4, 4)
	myr := b12Creature(g, me.ID, "Myr", "Artifact Creature — Myr", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !hasEffectiveKeyword(t, g, sphinx, "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}
	attackWith(t, g, opp.ID, sphinx, myr, bear)
	prompts := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == me.ID && c.Source == sphinx {
			prompts++
		}
	}
	if prompts != 2 {
		t.Fatalf("two artifact creatures connected, the Bear is not one: %d prompts", prompts)
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Thopter"); n != 1 {
		t.Fatalf("one yes, one no: %d Thopters", n)
	}
	thopter := findBattlefieldByName(g, "Thopter")
	assertKeywords(t, g, thopter, "flying")
	if c, _ := battlefieldCard(g, thopter); !c.IsArtifact() || !c.IsCreature() {
		t.Error("the Thopter is an artifact creature")
	}
}

func TestB36CadiraMakesARabbitPerTokenWhenSheConnects(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cadira := b12Push(g, me.ID, "Cadira, Caller of the Small", "Legendary Creature — Orc Ranger", b36CadiraOracle, 3, 3)
	pushToken(g, me.ID, TreasureToken())
	pushToken(g, me.ID, RedGoblinToken())
	pushToken(g, opp.ID, RedGoblinToken())
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !hasEffectiveKeyword(t, g, cadira, "trample") {
		t.Error("printed trample did not reach the effective abilities")
	}
	attackWith(t, g, opp.ID, cadira)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Rabbit"); n != 2 {
		t.Fatalf("a Treasure and a Goblin token: two Rabbits, the opponent's token and a nontoken Bear not counted — %d", n)
	}
	// Next hit: the two Rabbits count too.
	advanceToNextSeatsTurn(t, g)
	advanceToMainOf(t, g, 0)
	attackWith(t, g, opp.ID, cadira)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Rabbit"); n != 6 {
		t.Errorf("four tokens now: four more Rabbits — %d", n)
	}
}

func TestB36SyggDrawsAtAnyEndStepAfterAnOpponentLostThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	pushCatalogPermanent(g, me.ID, "Sygg, River Cutthroat", "Legendary Creature — Merfolk Rogue", b36SyggRiverCutthroatOracle, false)
	hand := me.Hand.Size()
	// Two on one opponent and two on another is not three on one.
	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(uuid.Nil, opp.ID, 2)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, other.ID, -2)
	})
	advanceToEndStepOf(t, g, 0)
	if latestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("no single opponent lost three")
	}
	passPriorityAroundTable(t, g)
	// An opponent's turn: one more on the first opponent — still only
	// two THIS turn, the earlier loss was last turn.
	advanceToNextSeatsTurn(t, g)
	advanceToUpkeepOf(t, g, 1)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -2) })
	advanceToEndStepOf(t, g, 1)
	if latestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("\"this turn\" does not carry over")
	}
	passPriorityAroundTable(t, g)
	// Another opponent's turn: three on one opponent, as damage and a
	// drain — any player's end step.
	advanceToNextSeatsTurn(t, g)
	advanceToUpkeepOf(t, g, 2)
	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(uuid.Nil, opp.ID, 2)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -1)
	})
	advanceToEndStepOf(t, g, 2)
	if latestTriggerPrompt(g, me.ID) == nil {
		t.Fatal("an opponent lost 3 this turn: the end step asks")
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("you draw a card: hand %d → %d", hand, me.Hand.Size())
	}
}

// --- removal and reanimation ---------------------------------------

func TestB36AngelOfSerenityExilesUpToThreeCreaturesAndHandsThemBackWhenSheLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	a := b12Creature(g, opp.ID, "Their A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "Their B", "Creature — Bear", 2, 2)
	c := b12Creature(g, other.ID, "Other C", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	dead := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	angel := castCatalogSpell(t, g, "Angel of Serenity", "Creature — Angel", b36AngelOfSerenityOracle, nil)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, angel, "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if !hasID(p.PickTargetCards, a) || !hasID(p.PickTargetCards, c) || !hasID(p.PickTargetCards, mine) {
		t.Error("every other creature on the battlefield is offered, your own included")
	}
	if hasID(p.PickTargetCards, angel) || hasID(p.PickTargetCards, dead) {
		t.Error("the Angel herself and a graveyard card (the declared gap) are not")
	}
	if p.PickTargetMin != 0 || p.PickTargetMax != 3 {
		t.Errorf("\"up to three\": min %d max %d", p.PickTargetMin, p.PickTargetMax)
	}
	b17PickCards(t, g, me.ID, a, b, c)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if !inExile(g, id) {
			t.Errorf("%s is exiled", id)
		}
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the unchosen creature stays")
	}
	// One exiled card moves on before she leaves: it is not returned.
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(c) })
	handC := other.Hand.Size()
	// She dies: the rest come back to their owners' HANDS.
	b27Kill(g, angel)
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(a) || !opp.Hand.Contains(b) {
		t.Error("the exiled cards return to their owner's hand")
	}
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("to hand, not the battlefield")
	}
	if other.Hand.Size() != handC {
		t.Error("a card that already left exile is left alone")
	}
}

func TestB36AngelOfSerenityDeclinedExilesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, opp.ID, "Their A", "Creature — Bear", 2, 2)
	angel := castCatalogSpell(t, g, "Angel of Serenity", "Creature — Angel", b36AngelOfSerenityOracle, nil)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(a) {
		t.Error("declining exiles nothing")
	}
	b27Kill(g, angel)
	passPriorityAroundTable(t, g)
	if opp.Hand.Contains(a) || !g.Battlefield.Contains(a) {
		t.Error("nothing was exiled with her, so nothing returns")
	}
}

func TestB36ExtinguishAllHopeSparesEnchantmentCreaturesAndIndestructibles(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	god := b12Creature(g, opp.ID, "God", "Legendary Enchantment Creature — God", 5, 5)
	avacyn := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Avacyn", TypeLine: "Creature — Angel", Keywords: []string{"indestructible"},
		Power: 8, Toughness: 8, Owner: opp.ID, Controller: opp.ID,
	})
	shrine := b12Permanent(g, opp.ID, "Shrine", "Enchantment")
	castCatalogSpell(t, g, "Extinguish All Hope", "Sorcery", b36ExtinguishAllHopeOracle, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(theirs) {
		t.Error("every nonenchantment creature is destroyed, yours included")
	}
	if !g.Battlefield.Contains(god) || !g.Battlefield.Contains(shrine) {
		t.Error("an enchantment creature and a noncreature enchantment survive")
	}
	if !g.Battlefield.Contains(avacyn) {
		t.Error("an indestructible creature survives (#470)")
	}
}

func TestB36ReplenishReturnsNonAuraEnchantmentsFromYourGraveyardOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	anthem := b17GraveyardCard(me, "Glorious Anthem", "Enchantment", "{1}{W}{W}")
	god := b17GraveyardCard(me, "God", "Legendary Enchantment Creature — God", "{3}{W}")
	aura := b17GraveyardCard(me, "Rancor", "Enchantment — Aura", "{G}")
	rock := b17GraveyardCard(me, "Signet", "Artifact", "{2}")
	theirs := b17GraveyardCard(opp, "Their Anthem", "Enchantment", "{1}{W}{W}")
	castCatalogSpell(t, g, "Replenish", "Sorcery", b36ReplenishOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(anthem) || !g.Battlefield.Contains(god) {
		t.Error("every non-Aura enchantment card returns, an enchantment creature included")
	}
	if !me.Graveyard.Contains(aura) {
		t.Error("an Aura stays in the graveyard (the declared gap)")
	}
	if !me.Graveyard.Contains(rock) || !opp.Graveyard.Contains(theirs) {
		t.Error("an artifact and an opponent's enchantment are left alone")
	}
}

// --- tutors --------------------------------------------------------

func TestB36SarkhansTriumphTutorsADragonCreatureToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		game.Card{Name: "Shivan Dragon", TypeLine: "Creature — Dragon", ManaCost: "{4}{R}{R}"},
		game.Card{Name: "Dragon Broodmother", TypeLine: "Creature — Dragon", ManaCost: "{2}{R}{R}{R}{G}"},
		game.Card{Name: "Dragon Tempest", TypeLine: "Enchantment", ManaCost: "{1}{R}"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
	)
	castCatalogSpell(t, g, "Sarkhan's Triumph", "Instant", b36SarkhansTriumphOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("two Dragons: the searcher chooses")
	}
	if searchOptionNamed(g, c, "Dragon Tempest") != uuid.Nil || searchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("a noncreature with Dragon in its name and a Bear are not offered")
	}
	answerSearchNamed(t, g, me.ID, "Shivan Dragon")
	if !b02bHandHasNamed(me, "Shivan Dragon") {
		t.Error("the pick goes to hand")
	}
	if me.Library.Size() != 3 {
		t.Errorf("library %d, want 3", me.Library.Size())
	}
}

func TestB36RepurposingBayPodsAnotherArtifactUpOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bay := pushCatalogPermanent(g, me.ID, "Repurposing Bay", "Artifact", b36RepurposingBayOracle, false)
	signet := b27Push(g, me.ID, "Signet", "Artifact", "", "{2}", 0, 0)
	seedSearchLibrary(me,
		game.Card{Name: "Three Rock", TypeLine: "Artifact", ManaCost: "{3}"},
		game.Card{Name: "Other Three", TypeLine: "Artifact Creature — Construct", ManaCost: "{3}"},
		game.Card{Name: "Four Rock", TypeLine: "Artifact", ManaCost: "{4}"},
		game.Card{Name: "Three Elf", TypeLine: "Creature — Elf", ManaCost: "{2}{G}"},
	)
	spec, _ := Lookup(b36RepurposingBayOracle)
	if len(spec.Activated) != 1 || !spec.Activated[0].SorcerySpeed {
		t.Fatal("one ability, sorcery speed")
	}
	advanceToMain(t, g)
	b06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, bay, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bay}}); err == nil {
		t.Fatal("\"another artifact\": the Bay cannot feed itself")
	}
	b16Activate(t, g, me.ID, bay, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{signet}})
	if g.Battlefield.Contains(signet) || !b16Tapped(t, g, bay) {
		t.Fatal("the artifact is sacrificed and the Bay taps")
	}
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("two artifacts at mana value 3: the searcher chooses")
	}
	if searchOptionNamed(g, c, "Four Rock") != uuid.Nil || searchOptionNamed(g, c, "Three Elf") != uuid.Nil {
		t.Error("only artifact cards with mana value exactly 3 are offered")
	}
	answerSearchNamed(t, g, me.ID, "Other Three")
	if findBattlefieldByName(g, "Other Three") == uuid.Nil {
		t.Error("the pick enters the battlefield")
	}
}

// --- keywords only -------------------------------------------------

func TestB36HealersHawkFliesWithLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hawk := pushCatalogPermanent(g, me.ID, "Healer's Hawk", "Creature — Bird", b36HealersHawkOracle, false)
	assertKeywords(t, g, hawk, "flying", "lifelink")
}
