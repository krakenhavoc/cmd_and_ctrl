package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- registration --------------------------------------------------

// Every card the batch registers, pinned by oracle ID → name. Alpine
// Meadow is a row in the Dominaria United tapland table, so a
// transposed row is invisible until someone plays that exact card.
//
// Price of Fame is NOT here even though the issue lists it as ready:
// #746 registered it before this batch ran, and its coverage lives in
// self_cost_modifier_test.go. Its row in the batch's "ready today"
// list was already stale.
func TestBatch42CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B42AbzanCharmOracle:          "Abzan Charm",
		effectstest.B42AlpineMeadowOracle:        "Alpine Meadow",
		effectstest.B42AncestorsAidOracle:        "Ancestors' Aid",
		effectstest.B42BoomerangBasicsOracle:     "Boomerang Basics",
		effectstest.B42BruteForceOracle:          "Brute Force",
		effectstest.B42ChildrenOfKorlisOracle:    "Children of Korlis",
		effectstest.B42CompulsiveResearchOracle:  "Compulsive Research",
		effectstest.B42CryptIncursionOracle:      "Crypt Incursion",
		effectstest.B42DeeprootWatersOracle:      "Deeproot Waters",
		effectstest.B42DesecratedTombOracle:      "Desecrated Tomb",
		effectstest.B42DreamFractureOracle:       "Dream Fracture",
		effectstest.B42DuskLegionZealotOracle:    "Dusk Legion Zealot",
		effectstest.B42FeveredVisionsOracle:      "Fevered Visions",
		effectstest.B42HighSocietyHunterOracle:   "High-Society Hunter",
		effectstest.B42ImperviousGreatwurmOracle: "Impervious Greatwurm",
		effectstest.B42LilianasTriumphOracle:     "Liliana's Triumph",
		effectstest.B42MerfolkLooterOracle:       "Merfolk Looter",
		effectstest.B42OrcishSiegemasterOracle:   "Orcish Siegemaster",
		effectstest.B42PawpatchFormationOracle:   "Pawpatch Formation",
		effectstest.B42PridemalkinOracle:         "Pridemalkin",
		effectstest.B42PrimordialSageOracle:      "Primordial Sage",
		effectstest.B42RingOfTheLuciiOracle:      "Ring of the Lucii",
		effectstest.B42SeersSundialOracle:        "Seer's Sundial",
		effectstest.B42TimberwatchElfOracle:      "Timberwatch Elf",
		effectstest.B42TitansStrengthOracle:      "Titan's Strength",
		effectstest.B42ViciousRumorsOracle:       "Vicious Rumors",
		effectstest.B42VindictiveVampireOracle:   "Vindictive Vampire",
		effectstest.B42WallOfReverenceOracle:     "Wall of Reverence",
		effectstest.B42WatcherOfTheSpheresOracle: "Watcher of the Spheres",
	}
	if len(want) != 29 {
		t.Fatalf("the batch registers 29 cards, the table lists %d", len(want))
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

// --- pump spells ---------------------------------------------------

func TestB42BruteForcePumpsUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Brute Force", "Instant", effectstest.B42BruteForceOracle, effectstest.B42CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("Brute Force makes a 2/2 into a 5/5: got %d/%d", p, tough)
	}
	effectstest.AdvancePastCleanupOf(t, g, 0)
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("the pump wears off at cleanup: power %d, want 2", p)
	}
}

func TestB42TitansStrengthPumpsAndScries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Titan's Strength", "Instant", effectstest.B42TitansStrengthOracle, effectstest.B42CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 5 || tough != 3 {
		t.Errorf("Titan's Strength is +3/+1: got %d/%d, want 5/3", p, tough)
	}
	effectstest.ResolveAnyScryFor(t, g, me.ID)
}

// The Treasure is a separate sentence, so it lands alongside the pump
// and the first-strike grant rather than instead of either.
func TestB42AncestorsAidPumpsGrantsFirstStrikeAndMakesATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	before := len(g.Battlefield.Cards)
	effectstest.CastCatalogSpell(t, g, "Ancestors' Aid", "Instant", effectstest.B42AncestorsAidOracle, effectstest.B42CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("Ancestors' Aid is +2/+0: got %d/%d, want 4/2", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("the target gains first strike until end of turn")
	}
	if got := len(g.Battlefield.Cards); got != before+1 {
		t.Errorf("a Treasure token joins the battlefield: %d → %d", before, got)
	}
}

// --- targeted utility ----------------------------------------------

// The draw rider reads who CONTROLLED the permanent, before it moves:
// your own permanent cantrips, an opponent's does not.
func TestB42BoomerangBasicsDrawsOnlyForYourOwnPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 1, 1)
	before := len(me.Hand.Cards)
	effectstest.CastCatalogSpell(t, g, "Boomerang Basics", "Sorcery — Lesson", effectstest.B42BoomerangBasicsOracle, effectstest.B42CardTarget(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Fatal("the target is returned to its owner's hand")
	}
	// The bounced creature plus the drawn card. (The Boomerang itself
	// is net zero: the fixture puts it in hand to cast it.)
	if got := len(me.Hand.Cards); got != before+2 {
		t.Errorf("your own permanent cantrips: hand %d → %d, want +2 (bounced card in, card drawn)", before, got)
	}

	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 1, 1)
	handBefore := len(me.Hand.Cards)
	effectstest.CastCatalogSpell(t, g, "Boomerang Basics", "Sorcery — Lesson", effectstest.B42BoomerangBasicsOracle, effectstest.B42CardTarget(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("an opponent's permanent is returned too")
	}
	if got := len(me.Hand.Cards); got != handBefore {
		t.Errorf("no draw off an opponent's permanent: hand %d → %d, want unchanged", handBefore, got)
	}
	if !opp.Hand.Contains(theirs) {
		t.Error("their creature goes to THEIR hand")
	}
}

func TestB42DreamFractureCountersAndBothPlayersDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: victim, Name: "Doomed Spell", TypeLine: "Instant",
		Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.AdvanceToMain(t, g)
	// The opponent casts into our open mana; we answer at instant
	// speed from the same main phase.
	if err := g.CastSpell(opp.ID, victim, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the victim: %v", err)
	}
	myHand, theirHand := len(me.Hand.Cards), len(opp.Hand.Cards)

	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Dream Fracture", TypeLine: "Instant",
		OracleID: effectstest.B42DreamFractureOracle, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("cast Dream Fracture: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)

	if !opp.Graveyard.Contains(victim) {
		t.Error("the countered spell goes to its owner's graveyard")
	}
	if got := len(opp.Hand.Cards); got != theirHand+1 {
		t.Errorf("the countered spell's controller draws: %d → %d", theirHand, got)
	}
	if got := len(me.Hand.Cards); got != myHand+1 {
		t.Errorf("you draw too: hand %d → %d, want +1", myHand, got)
	}
}

// --- mass effects --------------------------------------------------

func TestB42ViciousRumorsPingsDiscardsMillsAndGains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lifeBefore := me.Life
	oppLife := map[uuid.UUID]int{}
	for _, p := range g.Seats[1:] {
		oppLife[p.ID] = p.Life
	}
	effectstest.CastCatalogSpell(t, g, "Vicious Rumors", "Sorcery", effectstest.B42ViciousRumorsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	for _, p := range g.Seats[1:] {
		if p.Life != oppLife[p.ID]-1 {
			t.Errorf("each opponent takes 1: %s %d → %d", p.Name, oppLife[p.ID], p.Life)
		}
	}
	// "You gain 1 life" is the LAST printed sentence, and since #1027
	// it waits for the run: the caster must not be a life up while the
	// table is still choosing what to pitch.
	if me.Life != lifeBefore {
		t.Errorf("you gained life before the table discarded: %d → %d", lifeBefore, me.Life)
	}
	for i, p := range g.Seats[1:] {
		graveBefore := len(p.Graveyard.Cards)
		effectstest.DiscardFromHand(t, g, p.ID)
		// The discard lands, then this seat's own "then" mills — two
		// cards, on this opponent's own leg.
		if got := len(p.Graveyard.Cards); got != graveBefore+2 {
			t.Errorf("%s discards then mills: graveyard %d → %d, want +2", p.Name, graveBefore, got)
		}
		last := i == len(g.Seats[1:])-1
		want := lifeBefore
		if last {
			want = lifeBefore + 1
		}
		if me.Life != want {
			t.Errorf("after %s answered, your life is %d, want %d (the gain is the whole "+
				"instruction's continuation, so it lands once, on the last leg)",
				p.Name, me.Life, want)
		}
	}
}

func TestB42LilianasTriumphEdictsAndAddsDiscardWithALiliana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Their Only Creature", 2, 2)
	for _, p := range g.Seats[2:] {
		effectstest.PushVanillaCreature(g, p.ID, "Fodder", 1, 1)
	}
	effectstest.CastCatalogSpell(t, g, "Liliana's Triumph", "Instant", effectstest.B42LilianasTriumphOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, opp.ID, victim)
	if g.Battlefield.Contains(victim) {
		t.Error("the opponent sacrifices the creature they chose")
	}
	if effectstest.DiscardChoiceFor(g, opp.ID) != nil {
		t.Error("no Liliana on the battlefield means no discard")
	}
}

func TestB42LilianasTriumphDiscardsWhenYouControlALiliana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// A planeswalker needs loyalty, or the SBA kills it before the
	// Triumph ever resolves.
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Liliana, Dreadhorde General",
		TypeLine: "Legendary Planeswalker — Liliana",
		Counters: map[string]int{"loyalty": 5},
		Owner:    me.ID, Controller: me.ID,
	})
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Their Only Creature", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Liliana's Triumph", "Instant", effectstest.B42LilianasTriumphOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, opp.ID, victim)
	if effectstest.DiscardChoiceFor(g, opp.ID) == nil {
		t.Errorf("a Liliana planeswalker adds the discard for %s", opp.Name)
	}
}

func TestB42CryptIncursionExilesCreaturesAndGainsThreeEach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 3; i++ {
		opp.Graveyard.PushTop(game.Card{
			InstanceID: uuid.New(), Name: "Dead Thing", TypeLine: "Creature — Zombie",
			Owner: opp.ID, Controller: opp.ID,
		})
	}
	keep := uuid.New()
	opp.Graveyard.PushTop(game.Card{
		InstanceID: keep, Name: "Dead Sorcery", TypeLine: "Sorcery",
		Owner: opp.ID, Controller: opp.ID,
	})
	before := me.Life
	effectstest.CastCatalogSpell(t, g, "Crypt Incursion", "Instant", effectstest.B42CryptIncursionOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+9 {
		t.Errorf("three creature cards exiled is 9 life: %d → %d", before, me.Life)
	}
	if !opp.Graveyard.Contains(keep) {
		t.Error("only CREATURE cards are exiled — the sorcery stays")
	}
}

func TestB42CompulsiveResearchDrawsThreeThenAsksForTwoOrALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := len(me.Hand.Cards)
	effectstest.CastCatalogSpell(t, g, "Compulsive Research", "Sorcery", effectstest.B42CompulsiveResearchOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != before+3 {
		t.Fatalf("the target player draws three: hand %d → %d, want +3", before, got)
	}
	prompt := effectstest.DiscardChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("the discard prompt is queued after the draw")
	}
	// One non-land card is not a legal answer; two cards are.
	nonland := uuid.Nil
	for _, c := range me.Hand.Cards {
		if !c.IsLand() {
			nonland = c.InstanceID
			break
		}
	}
	if nonland == uuid.Nil {
		t.Skip("no nonland card in hand to test the Validate hook with")
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{nonland}); err == nil {
		t.Error("one non-land card must be refused — the printed card wants two, or one land")
	}
	// #626: and neither is discarding NOTHING. This shipped with
	// `UpTo: true`, whose floor is zero, and Validate is never asked
	// about an empty pick — so "discard two cards unless you discard
	// a land card" could be answered by pitching nothing at all. The
	// floor is DiscardPrompt.Min now.
	if prompt.ChooseMin != 1 {
		t.Errorf("discard floor = %d, want 1 — you discard something", prompt.ChooseMin)
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, nil); err == nil {
		t.Error("discarding nothing satisfied the clause")
	}
}

// --- modal ---------------------------------------------------------

func TestB42PawpatchFormationDestroysAFlierAndDrawsWithFood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flier := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 2, Toughness: 2, Keywords: []string{"flying"},
		Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.CastModal(t, g, "Pawpatch Formation", "Instant", effectstest.B42PawpatchFormationOracle,
		[]int{0}, effectstest.B42CardTarget(flier))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(flier) {
		t.Error("mode 0 destroys the flier")
	}

	hand, board := len(me.Hand.Cards), len(g.Battlefield.Cards)
	effectstest.CastModal(t, g, "Pawpatch Formation", "Instant", effectstest.B42PawpatchFormationOracle, []int{2}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("mode 2 draws a card: hand %d → %d, want +1", hand, got)
	}
	if got := len(g.Battlefield.Cards); got != board+1 {
		t.Errorf("mode 2 makes a Food: battlefield %d → %d", board, got)
	}
}

func TestB42AbzanCharmExilesDrawsAndCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	big := effectstest.PushVanillaCreature(g, opp.ID, "Big", 4, 4)
	effectstest.CastModal(t, g, "Abzan Charm", "Instant", effectstest.B42AbzanCharmOracle, []int{0}, effectstest.B42CardTarget(big))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("mode 0 exiles a creature with power 3 or greater")
	}
	if opp.Graveyard.Contains(big) {
		t.Error("EXILED, not destroyed — it must not reach the graveyard")
	}

	hand, life := len(me.Hand.Cards), me.Life
	effectstest.CastModal(t, g, "Abzan Charm", "Instant", effectstest.B42AbzanCharmOracle, []int{1}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != hand+2 {
		t.Errorf("mode 1 draws two: hand %d → %d, want +2", hand, got)
	}
	if me.Life != life-2 {
		t.Errorf("mode 1 loses 2 life: %d → %d", life, me.Life)
	}

	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.CastModal(t, g, "Abzan Charm", "Instant", effectstest.B42AbzanCharmOracle, []int{2}, effectstest.B42CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.CurrentPower(t, g, bear); p != 4 {
		t.Errorf("mode 2 puts both counters on the one target: power %d, want 4", p)
	}
}

// --- creatures -----------------------------------------------------

func TestB42DuskLegionZealotDrawsAndLosesOneOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand, life := len(me.Hand.Cards), me.Life
	effectstest.CastCatalogSpell(t, g, "Dusk Legion Zealot", "Creature — Vampire Soldier", effectstest.B42DuskLegionZealotOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("the Zealot replaces itself: hand %d → %d, want +1", hand, got)
	}
	if me.Life != life-1 {
		t.Errorf("and costs 1 life: %d → %d", life, me.Life)
	}
}

func TestB42MerfolkLooterDrawsThenDiscards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	looter := effectstest.B42Permanent(g, me.ID, "Merfolk Looter", "Creature — Merfolk Rogue", effectstest.B42MerfolkLooterOracle, 1, 1)
	hand := len(me.Hand.Cards)
	if err := g.ActivateCatalogAbility(me.ID, looter, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the Looter: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Fatalf("the draw happens first: hand %d → %d", hand, got)
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	if got := len(me.Hand.Cards); got != hand {
		t.Errorf("then the discard: hand %d → %d, want back to %d", hand+1, got, hand)
	}
}

func TestB42VindictiveVampireDrainsOnYourCreatureDeaths(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B42Permanent(g, me.ID, "Vindictive Vampire", "Creature — Vampire", effectstest.B42VindictiveVampireOracle, 2, 3)
	fodder := effectstest.PushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	life := me.Life
	oppLife := map[uuid.UUID]int{}
	for _, p := range g.Seats[1:] {
		oppLife[p.ID] = p.Life
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(fodder) })
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats[1:] {
		if p.Life != oppLife[p.ID]-1 {
			t.Errorf("each opponent takes 1: %s %d → %d", p.Name, oppLife[p.ID], p.Life)
		}
	}
	if me.Life != life+1 {
		t.Errorf("you gain 1: %d → %d", life, me.Life)
	}
}

func TestB42PrimordialSageOffersADrawOnEachCreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B42Permanent(g, me.ID, "Primordial Sage", "Creature — Spirit", effectstest.B42PrimordialSageOracle, 4, 5)
	hand := len(me.Hand.Cards)
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("the Sage's optional draw: hand %d → %d, want +1", hand, got)
	}
}

func TestB42PridemalkinCountersATargetAndGrantsTrampleToCounteredCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	other := effectstest.PushVanillaCreature(g, me.ID, "No Counters", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Pridemalkin", "Creature — Cat", effectstest.B42PridemalkinOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)

	if p := effectstest.CurrentPower(t, g, bear); p != 3 {
		t.Errorf("the ETB puts a +1/+1 counter on the target: power %d, want 3", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("a creature you control with a +1/+1 counter has trample")
	}
	if effectstest.HasEffectiveKeyword(t, g, other, "trample") {
		t.Error("a creature with no counter gets nothing")
	}
}

func TestB42WallOfReverenceGainsLifeEqualToATargetsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B42Permanent(g, me.ID, "Wall of Reverence", "Creature — Spirit Wall", effectstest.B42WallOfReverenceOracle, 1, 6)
	fatty := effectstest.PushVanillaCreature(g, me.ID, "Fatty", 7, 7)
	life := me.Life
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PickCard(t, g, me.ID, fatty)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+7 {
		t.Errorf("gain life equal to the target's power: %d → %d, want +7", life, me.Life)
	}
}

func TestB42ChildrenOfKorlisGivesBackTheTurnsLifeLoss(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	kid := effectstest.B42Permanent(g, me.ID, "Children of Korlis", "Creature — Human Rebel Cleric", effectstest.B42ChildrenOfKorlisOracle, 1, 1)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -13) })
	low := me.Life
	if err := g.ActivateCatalogAbility(me.ID, kid, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(kid) {
		t.Error("the sacrifice is a cost, paid at announce")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != low+13 {
		t.Errorf("gain the life lost this turn: %d → %d, want +13", low, me.Life)
	}
}

func TestB42ImperviousGreatwurmConvokesAndSurvivesDestruction(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.B42ImperviousGreatwurmOracle)
	if !ok {
		t.Fatal("Impervious Greatwurm is not registered")
	}
	if spec.TapCost == nil {
		t.Error("convoke is a TapCost, not an alternative cost")
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	wurm := effectstest.B42Permanent(g, me.ID, "Impervious Greatwurm", "Creature — Wurm", effectstest.B42ImperviousGreatwurmOracle, 16, 16)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(wurm) })
	if !g.Battlefield.Contains(wurm) {
		t.Error("indestructible: destruction does nothing")
	}
}

func TestB42TimberwatchElfPumpsByTheElfCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elf := effectstest.B42Permanent(g, me.ID, "Timberwatch Elf", "Creature — Elf", effectstest.B42TimberwatchElfOracle, 1, 2)
	effectstest.B42Permanent(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", "", 1, 1)
	// The whole table's Elves count, not only yours.
	effectstest.B42Permanent(g, opp.ID, "Their Elf", "Creature — Elf", "", 1, 1)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, elf, 0, game.ActivateAbilityParams{
		Targets: effectstest.B42CardTarget(bear),
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 5 {
		t.Errorf("three Elves on the battlefield is +3/+3: power %d, want 5", p)
	}
}

func TestB42OrcishSiegemasterTramplesTheSwarmAndSwingsForTheBiggest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	boss := effectstest.B42Permanent(g, me.ID, "Orcish Siegemaster", "Creature — Orc Soldier", effectstest.B42OrcishSiegemasterOracle, 0, 5)
	goblin := effectstest.B42Permanent(g, me.ID, "Goblin", "Creature — Goblin", "", 1, 1)
	elf := effectstest.B42Permanent(g, me.ID, "Elf", "Creature — Elf", "", 1, 1)
	effectstest.B42Permanent(g, me.ID, "Fatty", "Creature — Beast", "", 6, 6)
	if !effectstest.HasEffectiveKeyword(t, g, goblin, "trample") {
		t.Error("other Goblins you control have trample")
	}
	if effectstest.HasEffectiveKeyword(t, g, elf, "trample") {
		t.Error("only Orcs and Goblins — an Elf gets nothing")
	}
	effectstest.DeclareAttack(t, g, opp.ID, boss)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, boss); p != 6 {
		t.Errorf("+X/+0 where X is your greatest power (6): power %d, want 6", p)
	}
}

func TestB42HighSocietyHunterEatsACreatureAndDrawsOffNontokenDeaths(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hunter := effectstest.B42Permanent(g, me.ID, "High-Society Hunter", "Creature — Vampire Noble", effectstest.B42HighSocietyHunterOracle, 5, 3)
	snack := effectstest.PushVanillaCreature(g, me.ID, "Snack", 1, 1)

	hand := len(me.Hand.Cards)
	effectstest.DeclareAttack(t, g, opp.ID, hunter)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.AnswerChooseCards(t, g, me.ID, snack)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(snack) {
		t.Error("the chosen creature is sacrificed")
	}
	// The sacrifice is a nontoken creature death, so the second
	// ability draws as well.
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("a nontoken creature death draws a card: hand %d → %d", hand, got)
	}
	if p := effectstest.CurrentPower(t, g, hunter); p != 6 {
		t.Errorf("and the Hunter gets a +1/+1 counter: power %d, want 6", p)
	}
}

func TestB42HighSocietyHunterIgnoresTokenDeaths(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B42Permanent(g, me.ID, "High-Society Hunter", "Creature — Vampire Noble", effectstest.B42HighSocietyHunterOracle, 5, 3)
	tokens := effectstest.PushTokens(g, me.ID, effects.TokenCard("1/1 red Goblin"), 1)
	hand := len(me.Hand.Cards)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(tokens[0]) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != hand {
		t.Errorf("a TOKEN death draws nothing: hand %d → %d", hand, got)
	}
}

// --- artifacts and enchantments ------------------------------------

// "One or more creature cards leave your graveyard" is a batch
// trigger: two cards leaving at once is still one Bat.
func TestB42DesecratedTombMakesOneBatPerBatch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B42Permanent(g, me.ID, "Desecrated Tomb", "Artifact", effectstest.B42DesecratedTombOracle, 0, 0)
	var dead []uuid.UUID
	for i := 0; i < 2; i++ {
		id := uuid.New()
		me.Graveyard.PushTop(game.Card{
			InstanceID: id, Name: "Dead Thing", TypeLine: "Creature — Zombie",
			Owner: me.ID, Controller: me.ID,
		})
		dead = append(dead, id)
	}
	before := len(g.Battlefield.Cards)
	g.WithWriteLock(func() { _ = g.ExileCardsForEffect(dead) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(g.Battlefield.Cards); got != before+1 {
		t.Errorf("two creature cards leaving at once make ONE Bat: battlefield %d → %d", before, got)
	}
}

func TestB42SeersSundialOffersTwoManaForACardOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B42Permanent(g, me.ID, "Seer's Sundial", "Artifact", effectstest.B42SeersSundialOracle, 0, 0)
	hand := len(me.Hand.Cards)
	g.WithWriteLock(func() { _ = g.AddManaForEffect(me.ID, uuid.Nil, "{G}{G}") })
	effectstest.B12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	effectstest.PassPriorityUntilChoice(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != hand {
		t.Errorf("the land left the hand and the paid draw replaced it: hand %d → %d", hand, got)
	}
}

func TestB42DeeprootWatersMakesAHexproofMerfolkPerMerfolkSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B42Permanent(g, me.ID, "Deeproot Waters", "Enchantment", effectstest.B42DeeprootWatersOracle, 0, 0)
	before := len(g.Battlefield.Cards)
	effectstest.CastCatalogSpell(t, g, "Merfolk Thing", "Creature — Merfolk", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	var token *game.Card
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].Name == "Merfolk" {
				token = &g.Battlefield.Cards[i]
			}
		}
	})
	if token == nil {
		t.Fatalf("a Merfolk token joins the battlefield (was %d cards)", before)
	}
	if !effectstest.HasEffectiveKeyword(t, g, token.InstanceID, "hexproof") {
		t.Error("the token has hexproof")
	}
}

func TestB42FeveredVisionsDrawsForEveryoneAndBurnsAFullOpponentHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B42Permanent(g, me.ID, "Fevered Visions", "Enchantment", effectstest.B42FeveredVisionsOracle, 0, 0)
	effectstest.FillHandTo(t, g, opp, 4)
	// Measured at the end step, after their draw step has already
	// happened — the trigger is on the stack but has not resolved.
	effectstest.AdvanceToEndStepOf(t, g, 1)
	life, hand := opp.Life, len(opp.Hand.Cards)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(opp.Hand.Cards); got != hand+1 {
		t.Errorf("that player draws: hand %d → %d", hand, got)
	}
	if opp.Life != life-2 {
		t.Errorf("four or more cards in an opponent's hand is 2 damage: %d → %d", life, opp.Life)
	}
}

func TestB42WatcherOfTheSpheresDiscountsFliersOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B42Permanent(g, me.ID, "Watcher of the Spheres", "Creature — Bird Wizard", effectstest.B42WatcherOfTheSpheresOracle, 2, 2)
	// A flier is priced one generic cheaper; a ground creature and a
	// non-creature spell with flying-shaped text are not. The spells
	// are in no zone the layer engine reaches, so the keyword is read
	// off the card, which is the path HasKeyword takes off the
	// battlefield.
	if got := b42Price(t, g, me, "Serra Angel", "Creature — Angel", "{3}{W}{W}", "flying"); got != 2 {
		t.Errorf("a creature spell with flying costs {1} less: generic %d, want 2", got)
	}
	if got := b42Price(t, g, me, "Grizzly Bears", "Creature — Bear", "{1}{G}"); got != 1 {
		t.Errorf("a creature spell without flying is untouched: generic %d, want 1", got)
	}
	if got := b42Price(t, g, me, "Gust of Wind", "Instant", "{1}{U}", "flying"); got != 1 {
		t.Errorf("only CREATURE spells are discounted: generic %d, want 1", got)
	}
}

// b42Price prices a spell in the active seat's hand through whatever
// cost modifiers are on the battlefield, and returns the generic
// component — the half a reduction spends against (CR 601.2f).
func b42Price(t *testing.T, g *game.Game, seat *game.Player, name, typeLine, manaCost string, keywords ...string) int {
	t.Helper()
	base, err := game.ParseCost(manaCost)
	if err != nil {
		t.Fatalf("parse %q: %v", manaCost, err)
	}
	priced, err := g.ApplyCostModifiers(base, game.CostQuery{
		Card: game.Card{
			InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: manaCost,
			Keywords: keywords, Owner: seat.ID, Controller: seat.ID,
		},
		Controller: seat.ID,
		FromZone:   game.ZoneHand,
	})
	if err != nil {
		t.Fatalf("price %s: %v", name, err)
	}
	return priced.Generic
}

func TestB42RingOfTheLuciiTapsForTwoAndTapsDownForALife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ring := effectstest.B42Permanent(g, me.ID, "Ring of the Lucii", "Legendary Artifact", effectstest.B42RingOfTheLuciiOracle, 0, 0)
	if err := g.ActivateManaAbility(me.ID, ring, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if me.ManaPool.Total() != 2 {
		t.Errorf("{T}: Add {C}{C} adds two: pool %v", me.ManaPool)
	}

	g.WithWriteLock(func() {
		_ = g.UntapTargetForEffect(ring)
		_ = g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}")
	})
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Untapped", 2, 2)
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, ring, 0, game.ActivateAbilityParams{
		Targets: effectstest.B42CardTarget(victim),
	}); err != nil {
		t.Fatalf("activate the tap-down: %v", err)
	}
	if me.Life != life-1 {
		t.Errorf("the life is a cost, paid at announce: %d → %d", life, me.Life)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedOnBattlefield(t, g, victim) {
		t.Error("the target is tapped")
	}
}

// --- land ----------------------------------------------------------

func TestB42AlpineMeadowEntersTappedAndTapsForRedOrWhite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.B12PlayFromHand(t, g, "Alpine Meadow", "Snow Land — Mountain Plains", effectstest.B42AlpineMeadowOracle, game.CastSpellParams{})
	if !effectstest.B16Tapped(t, g, land) {
		t.Fatal("Alpine Meadow enters tapped")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	effectstest.B28TapForMana(t, g, me.ID, land, "W")
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "W" {
		t.Errorf("tapped for {W}: pool %v", got)
	}
}
