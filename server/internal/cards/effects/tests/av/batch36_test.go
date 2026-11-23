package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// b36Token pushes a creature token under controller with a real
// timestamp, so the layer cache sees it.
func b36Token(g *game.Game, controller uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Power: power, Toughness: toughness, Owner: controller, Controller: controller,
	})
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Akoum Refuge
// is a row in the gain-land cycle table, so a transposed row is
// invisible until someone plays that exact card.
func TestBatch36CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B36SanctumOfEternityOracle:  "Sanctum of Eternity",
		effectstest.B36LlanowarVisionaryOracle:  "Llanowar Visionary",
		effectstest.B36AngelOfSerenityOracle:    "Angel of Serenity",
		effectstest.B36TemperedSteelOracle:      "Tempered Steel",
		effectstest.B36AkoumRefugeOracle:        "Akoum Refuge",
		effectstest.B36AnimalSanctuaryOracle:    "Animal Sanctuary",
		effectstest.B36HeliodsPilgrimOracle:     "Heliod's Pilgrim",
		effectstest.B36JaggedBarrensOracle:      "Jagged Barrens",
		effectstest.B36WhitemaneLionOracle:      "Whitemane Lion",
		effectstest.B36CrucibleOfFireOracle:     "Crucible of Fire",
		effectstest.B36ArahboTheFirstFangOracle: "Arahbo, the First Fang",
		effectstest.B36ReplenishOracle:          "Replenish",
		effectstest.B36RepurposingBayOracle:     "Repurposing Bay",
		effectstest.B36LifecreedDuoOracle:       "Lifecreed Duo",
		effectstest.B36LossarnachCaptainOracle:  "Lossarnach Captain",
		effectstest.B36TurbulentSpringsOracle:   "Turbulent Springs",
		effectstest.B36MavrenFeinOracle:         "Mavren Fein, Dusk Apostle",
		effectstest.B36FanaticOfMogisOracle:     "Fanatic of Mogis",
		effectstest.B36ShardingSphinxOracle:     "Sharding Sphinx",
		effectstest.B36SyggRiverCutthroatOracle: "Sygg, River Cutthroat",
		effectstest.B36BishopOfWingsOracle:      "Bishop of Wings",
		effectstest.B36TegwyllOracle:            "Tegwyll, Duke of Splendor",
		effectstest.B36ExtinguishAllHopeOracle:  "Extinguish All Hope",
		effectstest.B36HealersHawkOracle:        "Healer's Hawk",
		effectstest.B36CadiraOracle:             "Cadira, Caller of the Small",
		effectstest.B36ErodedCanyonOracle:       "Eroded Canyon",
		effectstest.B36SarkhansTriumphOracle:    "Sarkhan's Triumph",
	}
	if len(want) != 27 {
		t.Fatalf("the batch registers 27 cards, the table lists %d", len(want))
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

// --- lands ---------------------------------------------------------

func TestB36AkoumRefugeEntersTappedGainsOneAndTapsForEitherColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := me.Life
	land := effectstest.B12PlayFromHand(t, g, "Akoum Refuge", "Land", effectstest.B36AkoumRefugeOracle, game.CastSpellParams{})
	if !effectstest.B16Tapped(t, g, land) {
		t.Fatal("the refuge enters tapped")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("you gain 1 life: %d → %d", before, me.Life)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	effectstest.B28TapForMana(t, g, me.ID, land, "R")
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("tapped for {R}: pool %v", got)
	}
}

func TestB36DesertDualsEnterTappedPingAnOpponentAndTapForTheirColours(t *testing.T) {
	for _, row := range []struct{ name, oracle, colour string }{
		{"Jagged Barrens", effectstest.B36JaggedBarrensOracle, "B"},
		{"Eroded Canyon", effectstest.B36ErodedCanyonOracle, "U"},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[2]
		before := opp.Life
		land := effectstest.B12PlayFromHand(t, g, row.name, "Land — Desert", row.oracle, game.CastSpellParams{})
		if !effectstest.B16Tapped(t, g, land) {
			t.Fatalf("%s: the Desert enters tapped", row.name)
		}
		if p := effectstest.LatestPickTarget(g, me.ID); p == nil || effectstest.HasID(p.PickTargetPlayers, me.ID) || !effectstest.HasID(p.PickTargetPlayers, opp.ID) {
			t.Fatalf("%s: the trigger asks for an opponent, never you: %+v", row.name, p)
		}
		effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
		effectstest.PassPriorityAroundTable(t, g)
		if opp.Life != before-1 {
			t.Errorf("%s: target opponent takes 1: %d → %d", row.name, before, opp.Life)
		}
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
		effectstest.B28TapForMana(t, g, me.ID, land, row.colour)
		if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != row.colour {
			t.Errorf("%s: tapped for {%s}: pool %v", row.name, row.colour, got)
		}
	}
}

func TestB36TurbulentSpringsEntersUntappedOnlyAgainstEightOpposingLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.B36Lands(g, opp.ID, 4)
	effectstest.B36Lands(g, other.ID, 3)
	effectstest.B36Lands(g, me.ID, 5)
	tapped := effectstest.B12PlayFromHand(t, g, "Turbulent Springs", "Land — Island Mountain", effectstest.B36TurbulentSpringsOracle, game.CastSpellParams{})
	if !effectstest.B16Tapped(t, g, tapped) {
		t.Fatal("seven opposing lands — your own do not count — is not eight: it enters tapped")
	}
	effectstest.B36Lands(g, other.ID, 1)
	effectstest.AdvanceToMainOf(t, g, 0)
	springs := effectstest.B12PlayFromHand(t, g, "Turbulent Springs", "Land — Island Mountain", effectstest.B36TurbulentSpringsOracle, game.CastSpellParams{})
	if effectstest.B16Tapped(t, g, springs) {
		t.Fatal("eight opposing lands between two opponents: it enters untapped")
	}
	effectstest.B28TapForMana(t, g, me.ID, springs, "U")
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "U" {
		t.Errorf("tapped for {U}: pool %v", got)
	}
}

func TestB36SanctumOfEternityReturnsOnlyACommanderYouOwn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sanctum := effectstest.PushCatalogPermanent(g, me.ID, "Sanctum of Eternity", "Land", effectstest.B36SanctumOfEternityOracle, false)
	mine := effectstest.B36Commander(g, me.ID, "My Commander")
	theirs := effectstest.B36Commander(g, opp.ID, "Their Commander")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	spec, _ := effects.Lookup(effectstest.B36SanctumOfEternityOracle)
	// #743: "Activate only during your turn" is a condition, not sorcery
	// speed — activation_conditions_test.go pins the windows it opens.
	if len(spec.Activated) != 1 || spec.Activated[0].SorcerySpeed || spec.Activated[0].Condition == nil {
		t.Fatal("one ability, gated by its during-your-turn condition rather than sorcery speed")
	}
	effectstest.B28TapForMana(t, g, me.ID, sanctum, "C")
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Fatalf("tapped for {C}: pool %v", got)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(sanctum) })
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C")
	for _, bad := range []uuid.UUID{theirs, bear} {
		if err := g.ActivateCatalogAbility(me.ID, sanctum, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bad)}); err == nil {
			t.Fatal("an opponent's commander and a creature that is not a commander are not legal targets")
		}
	}
	effectstest.B16Activate(t, g, me.ID, sanctum, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(mine)})
	if !effectstest.B16Tapped(t, g, sanctum) {
		t.Error("the Sanctum taps to activate")
	}
	// CR 903.9b (#539): a commander headed to hand is offered the
	// command zone. Declined, it is in hand, as the card says.
	effectstest.B21DeclineCommandZone(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || !me.Hand.Contains(mine) {
		t.Error("your commander returns to your hand")
	}
	// Accepted, it goes to the command zone instead.
	again := effectstest.B36Commander(g, me.ID, "My Commander")
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(sanctum) })
	effectstest.B06AddMana(me, "C", "C")
	effectstest.B16Activate(t, g, me.ID, sanctum, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(again)})
	effectstest.B36AcceptCommandZone(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Command.Contains(again) || me.Hand.Contains(again) {
		t.Error("the owner may send it to the command zone instead")
	}
}

func TestB36AnimalSanctuaryGrowsAPetOfAnyoneAndNothingElse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sanctuary := effectstest.PushCatalogPermanent(g, me.ID, "Animal Sanctuary", "Land", effectstest.B36AnimalSanctuaryOracle, false)
	dog := effectstest.B12Creature(g, opp.ID, "Their Dog", "Creature — Dog", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, sanctuary, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Fatal("a Bear is none of Bird, Cat, Dog, Goat, Ox or Snake")
	}
	effectstest.B16Activate(t, g, me.ID, sanctuary, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(dog)})
	if effectstest.CounterOn(g, dog, game.CounterPlusOne) != 1 {
		t.Errorf("an opponent's Dog gets the counter: %d", effectstest.CounterOn(g, dog, game.CounterPlusOne))
	}
	if !effectstest.B16Tapped(t, g, sanctuary) {
		t.Error("the Sanctuary taps to activate")
	}
}

// --- creatures with an enters trigger ------------------------------

func TestB36LlanowarVisionaryDrawsOnEntryAndTapsForGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Llanowar Visionary", "Creature — Elf Druid", effectstest.B36LlanowarVisionaryOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("the card was seeded into hand and cast, then one was drawn: hand %d → %d", hand, me.Hand.Size())
	}
	elf := effectstest.PushCatalogPermanent(g, me.ID, "Llanowar Visionary", "Creature — Elf Druid", effectstest.B36LlanowarVisionaryOracle, false)
	effectstest.B28TapForMana(t, g, me.ID, elf, "G")
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("tapped for {G}: pool %v", got)
	}
}

func TestB36HeliodsPilgrimMayTutorAnAura(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Rancor", TypeLine: "Enchantment — Aura", ManaCost: "{G}"},
		game.Card{Name: "Ethereal Armor", TypeLine: "Enchantment — Aura", ManaCost: "{W}"},
		game.Card{Name: "Glorious Anthem", TypeLine: "Enchantment", ManaCost: "{1}{W}{W}"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
	)
	effectstest.CastCatalogSpell(t, g, "Heliod's Pilgrim", "Creature — Human Cleric", effectstest.B36HeliodsPilgrimOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("you may search: the chooser asks")
	}
	if effectstest.SearchOptionNamed(g, c, "Rancor") == uuid.Nil || effectstest.SearchOptionNamed(g, c, "Ethereal Armor") == uuid.Nil {
		t.Error("both Auras are offered")
	}
	if effectstest.SearchOptionNamed(g, c, "Glorious Anthem") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("a non-Aura enchantment and a creature are not")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Rancor")
	if !effectstest.B02bHandHasNamed(me, "Rancor") {
		t.Error("the pick goes to hand")
	}
}

// #1337: the return is a resolution-time choice (ReturnOneYouControl
// / ChoosePermanents), not a target picked when the trigger goes on
// the stack, so the prompt is an own_permanents choose-cards pick
// rather than a pick_target one, and it only appears once the
// trigger resolves.
func TestB36WhitemaneLionFlashesInAndReturnsAnyCreatureYouControlItselfIncluded(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	lion := effectstest.CastCatalogSpell(t, g, "Whitemane Lion", "Creature — Cat", effectstest.B36WhitemaneLionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil {
		t.Fatal("Whitemane Lion queued no return pick")
	}
	if !effectstest.HasID(p.ChooseCards, bear) || !effectstest.HasID(p.ChooseCards, lion) {
		t.Error("every creature you control is offered, the Lion itself included")
	}
	if effectstest.HasID(p.ChooseCards, theirs) {
		t.Error("an opponent's creature is not")
	}
	if p.ChooseMin != 1 {
		t.Error("the return is mandatory")
	}
	if err := g.ResolveOwnPermanents(p.ID, me.ID, []uuid.UUID{lion}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if g.Battlefield.Contains(lion) || !me.Hand.Contains(lion) {
		t.Error("the Lion returns itself to hand")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the Bear stays")
	}
	// Again, bouncing the Bear this time.
	second := effectstest.CastCatalogSpell(t, g, "Whitemane Lion", "Creature — Cat", effectstest.B36WhitemaneLionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	p2 := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p2 == nil {
		t.Fatal("Whitemane Lion queued no return pick")
	}
	if err := g.ResolveOwnPermanents(p2.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if g.Battlefield.Contains(bear) || !me.Hand.Contains(bear) {
		t.Error("the Bear returns to hand")
	}
	if !effectstest.HasEffectiveKeyword(t, g, second, "flash") {
		t.Error("printed flash did not reach the effective abilities")
	}
}

func TestB36FanaticOfMogisBurnsEachOpponentForDevotionToRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B27Push(g, me.ID, "Mountain Dweller", "Creature — Ogre", "", "{1}{R}{R}", 3, 3, "R")
	effectstest.B27Push(g, me.ID, "Hybrid", "Enchantment", "", "{R/W}", 0, 0, "R", "W")
	effectstest.B27Push(g, me.ID, "Signet", "Artifact", "", "{2}", 0, 0)
	before := effectstest.LifeOfOpponents(g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Fanatic of Mogis", TypeLine: "Creature — Minotaur Shaman",
		OracleID: effectstest.B36FanaticOfMogisOracle, ManaCost: "{3}{R}", Colors: []string{"R"},
		Power: 4, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "C", "R")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Fanatic: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// {R}{R} + {R/W} + the Fanatic's own {R} = 4; the Signet's {2} is not red.
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-4 {
			t.Errorf("opponent %d: %d → %d, want -4 (devotion to red counted as the trigger resolves)", i+1, b, got)
		}
	}
}

// --- anthems and lords ---------------------------------------------

func TestB36TemperedSteelLiftsOnlyYourArtifactCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Tempered Steel", "Enchantment", effectstest.B36TemperedSteelOracle, false)
	thopter := b36Token(g, me.ID, "Thopter", "Token Artifact Creature — Thopter", 1, 1)
	construct := effectstest.B12Creature(g, me.ID, "Construct", "Artifact Creature — Construct", 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, me.ID, "Signet", "Artifact")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Myr", "Artifact Creature — Myr", 1, 1)
	if effectstest.EffectivePower(t, g, thopter) != 3 || effectstest.EffectiveToughness(t, g, thopter) != 3 {
		t.Errorf("a Thopter token is 3/3: %d/%d", effectstest.EffectivePower(t, g, thopter), effectstest.EffectiveToughness(t, g, thopter))
	}
	if effectstest.EffectivePower(t, g, construct) != 6 {
		t.Errorf("the Construct is 6/6: power %d", effectstest.EffectivePower(t, g, construct))
	}
	if effectstest.EffectivePower(t, g, bear) != 2 {
		t.Error("a non-artifact creature is untouched")
	}
	if effectstest.EffectivePower(t, g, rock) != 0 {
		t.Error("a noncreature artifact is untouched")
	}
	if effectstest.EffectivePower(t, g, theirs) != 1 {
		t.Error("an opponent's artifact creature is untouched")
	}
}

func TestB36CrucibleOfFireLiftsEveryDragonYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Crucible of Fire", "Enchantment", effectstest.B36CrucibleOfFireOracle, false)
	dragon := effectstest.B12Creature(g, me.ID, "Dragon", "Creature — Dragon", 4, 4)
	changeling := effectstest.B34Changeling(g, me.ID, "Shapeshifter")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Dragon", "Creature — Dragon", 4, 4)
	if effectstest.EffectivePower(t, g, dragon) != 7 || effectstest.EffectiveToughness(t, g, dragon) != 7 {
		t.Errorf("your Dragon is 7/7: %d/%d", effectstest.EffectivePower(t, g, dragon), effectstest.EffectiveToughness(t, g, dragon))
	}
	if effectstest.EffectivePower(t, g, changeling) != 4 {
		t.Error("a changeling is a Dragon")
	}
	if effectstest.EffectivePower(t, g, bear) != 2 || effectstest.EffectivePower(t, g, theirs) != 4 {
		t.Error("a Bear and an opponent's Dragon are untouched")
	}
}

func TestB36ArahboPumpsOtherCatsAndMakesACatForEachNontokenCatEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arahbo := effectstest.B36CastCreature(t, g, "Arahbo, the First Fang", "Legendary Creature — Cat Avatar", effectstest.B36ArahboTheFirstFangOracle, 2, 2)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Cat"); n != 1 {
		t.Fatalf("Arahbo's own entry makes a Cat: %d", n)
	}
	token := effectstest.FindBattlefieldByName(g, "Cat")
	if effectstest.EffectivePower(t, g, token) != 2 || effectstest.EffectivePower(t, g, arahbo) != 2 {
		t.Errorf("the token Cat is lifted to 2/2 and Arahbo is not: %d / %d", effectstest.EffectivePower(t, g, token), effectstest.EffectivePower(t, g, arahbo))
	}
	// A nontoken Cat entering under your control: another Cat.
	effectstest.B36CastCreature(t, g, "Savannah Lions", "Creature — Cat", "", 2, 1)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Cat"); n != 2 {
		t.Errorf("a nontoken Cat entering makes a second Cat token: %d", n)
	}
	// The token entering, a non-Cat, and an opponent's Cat: nothing.
	effectstest.B36CastCreature(t, g, "Bear", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.B36WhiteCatToken(), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Cat"); n != 2 {
		t.Errorf("a token Cat, a Bear and an opponent's Cat make nothing: %d", n)
	}
}

func TestB36TegwyllPumpsOtherFaeriesAndDrawsWhenOneDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tegwyll := effectstest.B12Push(g, me.ID, "Tegwyll, Duke of Splendor", "Legendary Creature — Faerie Noble", effectstest.B36TegwyllOracle, 2, 3)
	faerie := effectstest.B12Creature(g, me.ID, "Faerie", "Creature — Faerie Rogue", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Faerie", "Creature — Faerie", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, tegwyll, "flying") || !effectstest.HasEffectiveKeyword(t, g, tegwyll, "deathtouch") {
		t.Error("printed flying and deathtouch did not reach the effective abilities")
	}
	if effectstest.EffectivePower(t, g, faerie) != 2 || effectstest.EffectivePower(t, g, tegwyll) != 2 || effectstest.EffectivePower(t, g, theirs) != 1 {
		t.Error("other Faeries you control get +1/+1; Tegwyll and an opponent's Faerie do not")
	}
	life, hand := me.Life, me.Hand.Size()
	effectstest.B27Kill(g, faerie)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Errorf("another Faerie you control died: draw one, lose one — hand %d → %d, life %d → %d", hand, me.Hand.Size(), life, me.Life)
	}
	effectstest.B27Kill(g, bear)
	effectstest.B27Kill(g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Error("a Bear and an opponent's Faerie dying do nothing")
	}
	effectstest.B27Kill(g, tegwyll)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Error("Tegwyll's own death is not \"another Faerie\"")
	}
}

// --- go-wide and lifegain payoffs ----------------------------------

func TestB36LifecreedDuoGainsOnePerOtherCreatureYouControlEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	duo := effectstest.CastCatalogSpell(t, g, "Lifecreed Duo", "Creature — Bat Bird", effectstest.B36LifecreedDuoOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, duo, "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}
	before := me.Life
	if me.Life != before {
		t.Fatal("its own entry gains nothing")
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 2) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+2 {
		t.Errorf("two creatures entering gain two: %d → %d", before, me.Life)
	}
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 1)
		_ = g.CreateTokenForEffect(opp.ID, effects.RedGoblinToken(), 1)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+2 {
		t.Error("a noncreature entering and an opponent's creature gain nothing")
	}
}

func TestB36BishopOfWingsGainsFourPerAngelAndMakesASpiritWhenOneDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Bishop of Wings", "Creature — Human Cleric", effectstest.B36BishopOfWingsOracle, false)
	before := me.Life
	angel := effectstest.B36CastCreature(t, g, "Serra Angel", "Creature — Angel", "", 4, 4)
	if me.Life != before+4 {
		t.Errorf("an Angel entering gains 4: %d → %d", before, me.Life)
	}
	effectstest.B36CastCreature(t, g, "Bear", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.TokenCard("4/4 white Angel with flying"), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+4 {
		t.Error("a Bear and an opponent's Angel gain nothing")
	}
	effectstest.B27Kill(g, angel)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Spirit"); n != 1 {
		t.Fatalf("an Angel you control dying makes a Spirit: %d", n)
	}
	spirit := effectstest.FindBattlefieldByName(g, "Spirit")
	effectstest.AssertKeywords(t, g, spirit, "flying")
	theirs := effectstest.FindBattlefieldByName(g, "Angel")
	effectstest.B27Kill(g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Spirit"); n != 1 {
		t.Error("an opponent's Angel dying makes nothing")
	}
}

func TestB36LossarnachCaptainTapsOnHumansAndRaisesASoldierEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	captain := effectstest.CastCatalogSpell(t, g, "Lossarnach Captain", "Creature — Human Soldier", effectstest.B36LossarnachCaptainOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, captain, "first strike") {
		t.Error("printed first strike did not reach the effective abilities")
	}
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if !effectstest.HasID(p.PickTargetCards, theirs) || effectstest.HasID(p.PickTargetCards, mine) {
		t.Error("the Captain's own entry asks for a creature an opponent controls")
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, theirs) {
		t.Error("the chosen creature is tapped")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(theirs) })
	// A non-Human entering: silent.
	effectstest.B36CastCreature(t, g, "Elf", "Creature — Elf", "", 1, 1)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("an Elf is not a Human")
	}
	// Your upkeep: a Human Soldier, which is a Human entering.
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Human Soldier"); n != 1 {
		t.Fatalf("at the beginning of your upkeep, a Human Soldier: %d", n)
	}
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, theirs) {
		t.Error("the token entering fires the tap trigger too")
	}
	// An opponent's upkeep: nothing.
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Human Soldier"); n != 1 {
		t.Error("only your upkeep")
	}
}

func TestB36MavrenFeinMakesOneVampirePerCombatWhenNontokenVampiresAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mavren := effectstest.B12Push(g, me.ID, "Mavren Fein, Dusk Apostle", "Legendary Creature — Vampire Cleric", effectstest.B36MavrenFeinOracle, 2, 2)
	vamp := effectstest.B12Creature(g, me.ID, "Vampire Knight", "Creature — Vampire Knight", 2, 2)
	token := b36Token(g, me.ID, "Vampire", "Token Creature — Vampire", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, mavren, vamp, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Vampire"); n != 2 {
		t.Fatalf("two nontoken Vampires attacking is ONE trigger: %d Vampires (one seeded)", n)
	}
	made := uuid.Nil
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == me.ID && c.Name == "Vampire" && c.InstanceID != token {
			made = c.InstanceID
		}
	}
	effectstest.AssertKeywords(t, g, made, "lifelink")
	if c, _ := effectstest.BattlefieldCard(g, made); c.Power != 1 || c.Toughness != 1 || !c.HasSubtype("Vampire") {
		t.Error("a 1/1 Vampire")
	}
	// Next combat: only the token Vampire and the Bear attack — nothing.
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToMainOf(t, g, 0)
	effectstest.DeclareAttack(t, g, opp.ID, token, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Vampire"); n != 2 {
		t.Error("a token Vampire and a Bear attacking make nothing")
	}
}

func TestB36ShardingSphinxMayMakeAThopterPerArtifactCreatureThatConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sphinx := effectstest.B12Push(g, me.ID, "Sharding Sphinx", "Artifact Creature — Sphinx", effectstest.B36ShardingSphinxOracle, 4, 4)
	myr := effectstest.B12Creature(g, me.ID, "Myr", "Artifact Creature — Myr", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, sphinx, "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}
	effectstest.AttackWith(t, g, opp.ID, sphinx, myr, bear)
	prompts := effectstest.OwedMays(g, me.ID)
	if prompts != 2 {
		t.Fatalf("two artifact creatures connected, the Bear is not one: %d prompts", prompts)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Thopter"); n != 1 {
		t.Fatalf("one yes, one no: %d Thopters", n)
	}
	thopter := effectstest.FindBattlefieldByName(g, "Thopter")
	effectstest.AssertKeywords(t, g, thopter, "flying")
	if c, _ := effectstest.BattlefieldCard(g, thopter); !c.IsArtifact() || !c.IsCreature() {
		t.Error("the Thopter is an artifact creature")
	}
}

func TestB36CadiraMakesARabbitPerTokenWhenSheConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cadira := effectstest.B12Push(g, me.ID, "Cadira, Caller of the Small", "Legendary Creature — Orc Ranger", effectstest.B36CadiraOracle, 3, 3)
	effectstest.PushToken(g, me.ID, effects.TreasureToken())
	effectstest.PushToken(g, me.ID, effects.RedGoblinToken())
	effectstest.PushToken(g, opp.ID, effects.RedGoblinToken())
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, cadira, "trample") {
		t.Error("printed trample did not reach the effective abilities")
	}
	effectstest.AttackWith(t, g, opp.ID, cadira)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Rabbit"); n != 2 {
		t.Fatalf("a Treasure and a Goblin token: two Rabbits, the opponent's token and a nontoken Bear not counted — %d", n)
	}
	// Next hit: the two Rabbits count too.
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToMainOf(t, g, 0)
	effectstest.AttackWith(t, g, opp.ID, cadira)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Rabbit"); n != 6 {
		t.Errorf("four tokens now: four more Rabbits — %d", n)
	}
}

func TestB36SyggDrawsAtAnyEndStepAfterAnOpponentLostThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.PushCatalogPermanent(g, me.ID, "Sygg, River Cutthroat", "Legendary Creature — Merfolk Rogue", effectstest.B36SyggRiverCutthroatOracle, false)
	hand := me.Hand.Size()
	// Two on one opponent and two on another is not three on one.
	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(uuid.Nil, opp.ID, 2)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, other.ID, -2)
	})
	effectstest.AdvanceToEndStepOf(t, g, 0)
	if effectstest.OwedMays(g, me.ID) > 0 {
		t.Fatal("no single opponent lost three")
	}
	effectstest.PassPriorityAroundTable(t, g)
	// An opponent's turn: one more on the first opponent — still only
	// two THIS turn, the earlier loss was last turn.
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -2) })
	effectstest.AdvanceToEndStepOf(t, g, 1)
	if effectstest.OwedMays(g, me.ID) > 0 {
		t.Fatal("\"this turn\" does not carry over")
	}
	effectstest.PassPriorityAroundTable(t, g)
	// Another opponent's turn: three on one opponent, as damage and a
	// drain — any player's end step.
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, 2)
	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(uuid.Nil, opp.ID, 2)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -1)
	})
	effectstest.AdvanceToEndStepOf(t, g, 2)
	if effectstest.OwedMays(g, me.ID) == 0 {
		t.Fatal("an opponent lost 3 this turn: the end step asks")
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("you draw a card: hand %d → %d", hand, me.Hand.Size())
	}
}

// --- removal and reanimation ---------------------------------------

func TestB36AngelOfSerenityExilesUpToThreeCreaturesAndHandsThemBackWhenSheLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	a := effectstest.B12Creature(g, opp.ID, "Their A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Their B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, other.ID, "Other C", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	dead := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	angel := effectstest.CastCatalogSpell(t, g, "Angel of Serenity", "Creature — Angel", effectstest.B36AngelOfSerenityOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, angel, "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if !effectstest.HasID(p.PickTargetCards, a) || !effectstest.HasID(p.PickTargetCards, c) || !effectstest.HasID(p.PickTargetCards, mine) {
		t.Error("every other creature on the battlefield is offered, your own included")
	}
	if effectstest.HasID(p.PickTargetCards, angel) {
		t.Error("the Angel herself is not offered")
	}
	if !effectstest.HasID(p.PickTargetCards, dead) {
		t.Error("a creature card in a graveyard is offered (#3172)")
	}
	if p.PickTargetMin != 0 || p.PickTargetMax != 3 {
		t.Errorf("\"up to three\": min %d max %d", p.PickTargetMin, p.PickTargetMax)
	}
	effectstest.B17PickCards(t, g, me.ID, a, dead, c)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, dead, c} {
		if !effectstest.InExile(g, id) {
			t.Errorf("%s is exiled", id)
		}
	}
	if !g.Battlefield.Contains(b) || !g.Battlefield.Contains(mine) {
		t.Error("the unchosen creature stays")
	}
	// One exiled card moves on before she leaves: it is not returned.
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(c) })
	handC := other.Hand.Size()
	// She dies: the rest come back to their owners' HANDS.
	effectstest.B27Kill(g, angel)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(a) || !opp.Hand.Contains(dead) {
		t.Error("the exiled cards, a graveyard card among them, return to their owner's hand")
	}
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(dead) {
		t.Error("to hand, not the battlefield")
	}
	if other.Hand.Size() != handC {
		t.Error("a card that already left exile is left alone")
	}
}

func TestB36AngelOfSerenityDeclinedExilesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "Their A", "Creature — Bear", 2, 2)
	angel := effectstest.CastCatalogSpell(t, g, "Angel of Serenity", "Creature — Angel", effectstest.B36AngelOfSerenityOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(a) {
		t.Error("declining exiles nothing")
	}
	effectstest.B27Kill(g, angel)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Contains(a) || !g.Battlefield.Contains(a) {
		t.Error("nothing was exiled with her, so nothing returns")
	}
}

func TestB36ExtinguishAllHopeSparesEnchantmentCreaturesAndIndestructibles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	god := effectstest.B12Creature(g, opp.ID, "God", "Legendary Enchantment Creature — God", 5, 5)
	avacyn := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Avacyn", TypeLine: "Creature — Angel", Keywords: []string{"indestructible"},
		Power: 8, Toughness: 8, Owner: opp.ID, Controller: opp.ID,
	})
	shrine := effectstest.B12Permanent(g, opp.ID, "Shrine", "Enchantment")
	effectstest.CastCatalogSpell(t, g, "Extinguish All Hope", "Sorcery", effectstest.B36ExtinguishAllHopeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
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

// The Auras Replenish returns choose their hosts as they enter (#3158);
// that half is aura_host_cards_test.go's.
func TestB36ReplenishReturnsEnchantmentsFromYourGraveyardOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	anthem := effectstest.B17GraveyardCard(me, "Glorious Anthem", "Enchantment", "{1}{W}{W}")
	god := effectstest.B17GraveyardCard(me, "God", "Legendary Enchantment Creature — God", "{3}{W}")
	rock := effectstest.B17GraveyardCard(me, "Signet", "Artifact", "{2}")
	theirs := effectstest.B17GraveyardCard(opp, "Their Anthem", "Enchantment", "{1}{W}{W}")
	effectstest.CastCatalogSpell(t, g, "Replenish", "Sorcery", effectstest.B36ReplenishOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(anthem) || !g.Battlefield.Contains(god) {
		t.Error("every enchantment card returns, an enchantment creature included")
	}
	if !me.Graveyard.Contains(rock) || !opp.Graveyard.Contains(theirs) {
		t.Error("an artifact and an opponent's enchantment are left alone")
	}
}

// --- tutors --------------------------------------------------------

func TestB36SarkhansTriumphTutorsADragonCreatureToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Shivan Dragon", TypeLine: "Creature — Dragon", ManaCost: "{4}{R}{R}"},
		game.Card{Name: "Dragon Broodmother", TypeLine: "Creature — Dragon", ManaCost: "{2}{R}{R}{R}{G}"},
		game.Card{Name: "Dragon Tempest", TypeLine: "Enchantment", ManaCost: "{1}{R}"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
	)
	effectstest.CastCatalogSpell(t, g, "Sarkhan's Triumph", "Instant", effectstest.B36SarkhansTriumphOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("two Dragons: the searcher chooses")
	}
	if effectstest.SearchOptionNamed(g, c, "Dragon Tempest") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("a noncreature with Dragon in its name and a Bear are not offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Shivan Dragon")
	if !effectstest.B02bHandHasNamed(me, "Shivan Dragon") {
		t.Error("the pick goes to hand")
	}
	if me.Library.Size() != 3 {
		t.Errorf("library %d, want 3", me.Library.Size())
	}
}

func TestB36RepurposingBayPodsAnotherArtifactUpOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bay := effectstest.PushCatalogPermanent(g, me.ID, "Repurposing Bay", "Artifact", effectstest.B36RepurposingBayOracle, false)
	signet := effectstest.B27Push(g, me.ID, "Signet", "Artifact", "", "{2}", 0, 0)
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Three Rock", TypeLine: "Artifact", ManaCost: "{3}"},
		game.Card{Name: "Other Three", TypeLine: "Artifact Creature — Construct", ManaCost: "{3}"},
		game.Card{Name: "Four Rock", TypeLine: "Artifact", ManaCost: "{4}"},
		game.Card{Name: "Three Elf", TypeLine: "Creature — Elf", ManaCost: "{2}{G}"},
	)
	spec, _ := effects.Lookup(effectstest.B36RepurposingBayOracle)
	if len(spec.Activated) != 1 || !spec.Activated[0].SorcerySpeed {
		t.Fatal("one ability, sorcery speed")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, bay, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bay}}); err == nil {
		t.Fatal("\"another artifact\": the Bay cannot feed itself")
	}
	effectstest.B16Activate(t, g, me.ID, bay, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{signet}})
	if g.Battlefield.Contains(signet) || !effectstest.B16Tapped(t, g, bay) {
		t.Fatal("the artifact is sacrificed and the Bay taps")
	}
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("two artifacts at mana value 3: the searcher chooses")
	}
	if effectstest.SearchOptionNamed(g, c, "Four Rock") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Three Elf") != uuid.Nil {
		t.Error("only artifact cards with mana value exactly 3 are offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Other Three")
	if effectstest.FindBattlefieldByName(g, "Other Three") == uuid.Nil {
		t.Error("the pick enters the battlefield")
	}
}

// --- keywords only -------------------------------------------------

func TestB36HealersHawkFliesWithLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hawk := effectstest.PushCatalogPermanent(g, me.ID, "Healer's Hawk", "Creature — Bird", effectstest.B36HealersHawkOracle, false)
	effectstest.AssertKeywords(t, g, hawk, "flying", "lifelink")
}
