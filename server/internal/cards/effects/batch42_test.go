package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch42_test.go — card-level coverage for the card-coverage roadmap's
// batch 42 (#449, `edhrec_rank` 4354–4455). One test per observable
// behaviour, driven through a real cast, activation or attack rather
// than by calling primitives.

const (
	b42AbzanCharmOracle          = "4137a22c-f793-4e27-a9b2-72740a6e2121"
	b42AlpineMeadowOracle        = "8c281ebe-d9a1-48af-b58b-19c55aa4625b"
	b42AncestorsAidOracle        = "a95a0b13-d50f-41cf-8668-de60d445e7b0"
	b42BoomerangBasicsOracle     = "acd881eb-f3b4-4174-9fc2-e3c7c213886b"
	b42BruteForceOracle          = "9880ba09-d5b8-4675-bfb4-2161d86d2d41"
	b42ChildrenOfKorlisOracle    = "d44a8f89-b9c0-4e8a-9179-baa0c448a3c9"
	b42CompulsiveResearchOracle  = "f947808c-a1cc-41ea-9d06-e1279a0da527"
	b42CryptIncursionOracle      = "ce00e0d0-c6f4-4fe8-97c3-7279b06d8fdb"
	b42DeeprootWatersOracle      = "8e7b31eb-7a91-4992-b24a-d81173e1dbc8"
	b42DesecratedTombOracle      = "07ff619a-21ee-44b6-b666-ceab4a78096a"
	b42DreamFractureOracle       = "c5843d13-855f-41dd-ac13-3c7f7e18bb39"
	b42DuskLegionZealotOracle    = "ea2dce18-195e-4681-a687-f9819edaf9fc"
	b42FeveredVisionsOracle      = "70763549-4b4e-4cb8-8c02-0639ba18bb1a"
	b42HighSocietyHunterOracle   = "0dc158ba-7ec8-4558-91f5-0a87ef8380d4"
	b42ImperviousGreatwurmOracle = "65c0ab05-e740-4380-b39c-8bae9662f885"
	b42LilianasTriumphOracle     = "c2e68ac5-cfff-4b76-8062-25a82fdf9a5c"
	b42MerfolkLooterOracle       = "67362406-b1ca-49e2-800d-9050bfe8742a"
	b42OrcishSiegemasterOracle   = "d852099f-46c6-4676-860c-20a582f733d1"
	b42PawpatchFormationOracle   = "fa99cf82-ebf1-4526-ab9c-b24eb2970f2a"
	b42PridemalkinOracle         = "f9672b63-415a-448b-a3da-140df63a0f0c"
	b42PrimordialSageOracle      = "499f1e1e-d96e-481f-a5d9-eb38da927cd7"
	b42RingOfTheLuciiOracle      = "d37e595f-afe7-4ad9-87db-b9ceb3f435e8"
	b42SeersSundialOracle        = "9929d7ba-cdc5-4099-a0ee-3a15073336f3"
	b42TimberwatchElfOracle      = "50cee3ac-cba0-4abb-babf-de1928b1590e"
	b42TitansStrengthOracle      = "b0206b34-68b4-4b1f-ae79-6e2d0432ad4b"
	b42ViciousRumorsOracle       = "55b72b1f-6463-406a-8824-d99a3c028285"
	b42VindictiveVampireOracle   = "11ad664e-36d3-4d5b-8a87-59ea715877e0"
	b42WallOfReverenceOracle     = "0810983f-818a-43e6-a7b5-ebe0bc8b9f6a"
	b42WatcherOfTheSpheresOracle = "d43463d2-4395-4616-acc3-c4f80ebe3242"
)

// b42Permanent pushes a catalog permanent with a real battlefield
// timestamp, so the layer cache sees it and its statics apply.
func b42Permanent(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// b42CardTarget is the one-target slice every targeted cast in this
// file passes.
func b42CardTarget(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

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
		b42AbzanCharmOracle:          "Abzan Charm",
		b42AlpineMeadowOracle:        "Alpine Meadow",
		b42AncestorsAidOracle:        "Ancestors' Aid",
		b42BoomerangBasicsOracle:     "Boomerang Basics",
		b42BruteForceOracle:          "Brute Force",
		b42ChildrenOfKorlisOracle:    "Children of Korlis",
		b42CompulsiveResearchOracle:  "Compulsive Research",
		b42CryptIncursionOracle:      "Crypt Incursion",
		b42DeeprootWatersOracle:      "Deeproot Waters",
		b42DesecratedTombOracle:      "Desecrated Tomb",
		b42DreamFractureOracle:       "Dream Fracture",
		b42DuskLegionZealotOracle:    "Dusk Legion Zealot",
		b42FeveredVisionsOracle:      "Fevered Visions",
		b42HighSocietyHunterOracle:   "High-Society Hunter",
		b42ImperviousGreatwurmOracle: "Impervious Greatwurm",
		b42LilianasTriumphOracle:     "Liliana's Triumph",
		b42MerfolkLooterOracle:       "Merfolk Looter",
		b42OrcishSiegemasterOracle:   "Orcish Siegemaster",
		b42PawpatchFormationOracle:   "Pawpatch Formation",
		b42PridemalkinOracle:         "Pridemalkin",
		b42PrimordialSageOracle:      "Primordial Sage",
		b42RingOfTheLuciiOracle:      "Ring of the Lucii",
		b42SeersSundialOracle:        "Seer's Sundial",
		b42TimberwatchElfOracle:      "Timberwatch Elf",
		b42TitansStrengthOracle:      "Titan's Strength",
		b42ViciousRumorsOracle:       "Vicious Rumors",
		b42VindictiveVampireOracle:   "Vindictive Vampire",
		b42WallOfReverenceOracle:     "Wall of Reverence",
		b42WatcherOfTheSpheresOracle: "Watcher of the Spheres",
	}
	if len(want) != 29 {
		t.Fatalf("the batch registers 29 cards, the table lists %d", len(want))
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

// --- pump spells ---------------------------------------------------

func TestB42BruteForcePumpsUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	castCatalogSpell(t, g, "Brute Force", "Instant", b42BruteForceOracle, b42CardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("Brute Force makes a 2/2 into a 5/5: got %d/%d", p, tough)
	}
	advancePastCleanupOf(t, g, 0)
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("the pump wears off at cleanup: power %d, want 2", p)
	}
}

func TestB42TitansStrengthPumpsAndScries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	castCatalogSpell(t, g, "Titan's Strength", "Instant", b42TitansStrengthOracle, b42CardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 5 || tough != 3 {
		t.Errorf("Titan's Strength is +3/+1: got %d/%d, want 5/3", p, tough)
	}
	resolveAnyScryFor(t, g, me.ID)
}

// The Treasure is a separate sentence, so it lands alongside the pump
// and the first-strike grant rather than instead of either.
func TestB42AncestorsAidPumpsGrantsFirstStrikeAndMakesATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	before := len(g.Battlefield.Cards)
	castCatalogSpell(t, g, "Ancestors' Aid", "Instant", b42AncestorsAidOracle, b42CardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("Ancestors' Aid is +2/+0: got %d/%d, want 4/2", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "first strike") {
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	mine := pushVanillaCreature(g, me.ID, "Mine", 1, 1)
	before := len(me.Hand.Cards)
	castCatalogSpell(t, g, "Boomerang Basics", "Sorcery — Lesson", b42BoomerangBasicsOracle, b42CardTarget(mine))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Fatal("the target is returned to its owner's hand")
	}
	// The bounced creature plus the drawn card. (The Boomerang itself
	// is net zero: the fixture puts it in hand to cast it.)
	if got := len(me.Hand.Cards); got != before+2 {
		t.Errorf("your own permanent cantrips: hand %d → %d, want +2 (bounced card in, card drawn)", before, got)
	}

	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 1, 1)
	handBefore := len(me.Hand.Cards)
	castCatalogSpell(t, g, "Boomerang Basics", "Sorcery — Lesson", b42BoomerangBasicsOracle, b42CardTarget(theirs))
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: victim, Name: "Doomed Spell", TypeLine: "Instant",
		Owner: opp.ID, Controller: opp.ID,
	})
	advanceToMain(t, g)
	// The opponent casts into our open mana; we answer at instant
	// speed from the same main phase.
	if err := g.CastSpell(opp.ID, victim, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the victim: %v", err)
	}
	myHand, theirHand := len(me.Hand.Cards), len(opp.Hand.Cards)

	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Dream Fracture", TypeLine: "Instant",
		OracleID: b42DreamFractureOracle, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("cast Dream Fracture: %v", err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
	me := g.Seats[0]
	lifeBefore := me.Life
	oppLife := map[uuid.UUID]int{}
	for _, p := range g.Seats[1:] {
		oppLife[p.ID] = p.Life
	}
	castCatalogSpell(t, g, "Vicious Rumors", "Sorcery", b42ViciousRumorsOracle, nil)
	passPriorityAroundTable(t, g)

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
		discardFromHand(t, g, p.ID)
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
	g := newCatalogGame(t)
	opp := g.Seats[1]
	victim := pushVanillaCreature(g, opp.ID, "Their Only Creature", 2, 2)
	for _, p := range g.Seats[2:] {
		pushVanillaCreature(g, p.ID, "Fodder", 1, 1)
	}
	castCatalogSpell(t, g, "Liliana's Triumph", "Instant", b42LilianasTriumphOracle, nil)
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, opp.ID, victim)
	if g.Battlefield.Contains(victim) {
		t.Error("the opponent sacrifices the creature they chose")
	}
	if discardChoiceFor(g, opp.ID) != nil {
		t.Error("no Liliana on the battlefield means no discard")
	}
}

func TestB42LilianasTriumphDiscardsWhenYouControlALiliana(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// A planeswalker needs loyalty, or the SBA kills it before the
	// Triumph ever resolves.
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Liliana, Dreadhorde General",
		TypeLine: "Legendary Planeswalker — Liliana",
		Counters: map[string]int{"loyalty": 5},
		Owner:    me.ID, Controller: me.ID,
	})
	victim := pushVanillaCreature(g, opp.ID, "Their Only Creature", 2, 2)
	castCatalogSpell(t, g, "Liliana's Triumph", "Instant", b42LilianasTriumphOracle, nil)
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, opp.ID, victim)
	if discardChoiceFor(g, opp.ID) == nil {
		t.Errorf("a Liliana planeswalker adds the discard for %s", opp.Name)
	}
}

func TestB42CryptIncursionExilesCreaturesAndGainsThreeEach(t *testing.T) {
	g := newCatalogGame(t)
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
	castCatalogSpell(t, g, "Crypt Incursion", "Instant", b42CryptIncursionOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if me.Life != before+9 {
		t.Errorf("three creature cards exiled is 9 life: %d → %d", before, me.Life)
	}
	if !opp.Graveyard.Contains(keep) {
		t.Error("only CREATURE cards are exiled — the sorcery stays")
	}
}

func TestB42CompulsiveResearchDrawsThreeThenAsksForTwoOrALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := len(me.Hand.Cards)
	castCatalogSpell(t, g, "Compulsive Research", "Sorcery", b42CompulsiveResearchOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != before+3 {
		t.Fatalf("the target player draws three: hand %d → %d, want +3", before, got)
	}
	prompt := discardChoiceFor(g, me.ID)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flier := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 2, Toughness: 2, Keywords: []string{"flying"},
		Owner: opp.ID, Controller: opp.ID,
	})
	castModal(t, g, "Pawpatch Formation", "Instant", b42PawpatchFormationOracle,
		[]int{0}, b42CardTarget(flier))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(flier) {
		t.Error("mode 0 destroys the flier")
	}

	hand, board := len(me.Hand.Cards), len(g.Battlefield.Cards)
	castModal(t, g, "Pawpatch Formation", "Instant", b42PawpatchFormationOracle, []int{2}, nil)
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("mode 2 draws a card: hand %d → %d, want +1", hand, got)
	}
	if got := len(g.Battlefield.Cards); got != board+1 {
		t.Errorf("mode 2 makes a Food: battlefield %d → %d", board, got)
	}
}

func TestB42AbzanCharmExilesDrawsAndCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	big := pushVanillaCreature(g, opp.ID, "Big", 4, 4)
	castModal(t, g, "Abzan Charm", "Instant", b42AbzanCharmOracle, []int{0}, b42CardTarget(big))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("mode 0 exiles a creature with power 3 or greater")
	}
	if opp.Graveyard.Contains(big) {
		t.Error("EXILED, not destroyed — it must not reach the graveyard")
	}

	hand, life := len(me.Hand.Cards), me.Life
	castModal(t, g, "Abzan Charm", "Instant", b42AbzanCharmOracle, []int{1}, nil)
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != hand+2 {
		t.Errorf("mode 1 draws two: hand %d → %d, want +2", hand, got)
	}
	if me.Life != life-2 {
		t.Errorf("mode 1 loses 2 life: %d → %d", life, me.Life)
	}

	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	castModal(t, g, "Abzan Charm", "Instant", b42AbzanCharmOracle, []int{2}, b42CardTarget(bear))
	passPriorityAroundTable(t, g)
	if p := currentPower(t, g, bear); p != 4 {
		t.Errorf("mode 2 puts both counters on the one target: power %d, want 4", p)
	}
}

// --- creatures -----------------------------------------------------

func TestB42DuskLegionZealotDrawsAndLosesOneOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand, life := len(me.Hand.Cards), me.Life
	castCatalogSpell(t, g, "Dusk Legion Zealot", "Creature — Vampire Soldier", b42DuskLegionZealotOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("the Zealot replaces itself: hand %d → %d, want +1", hand, got)
	}
	if me.Life != life-1 {
		t.Errorf("and costs 1 life: %d → %d", life, me.Life)
	}
}

func TestB42MerfolkLooterDrawsThenDiscards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	looter := b42Permanent(g, me.ID, "Merfolk Looter", "Creature — Merfolk Rogue", b42MerfolkLooterOracle, 1, 1)
	hand := len(me.Hand.Cards)
	if err := g.ActivateCatalogAbility(me.ID, looter, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the Looter: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Fatalf("the draw happens first: hand %d → %d", hand, got)
	}
	discardFromHand(t, g, me.ID)
	if got := len(me.Hand.Cards); got != hand {
		t.Errorf("then the discard: hand %d → %d, want back to %d", hand+1, got, hand)
	}
}

func TestB42VindictiveVampireDrainsOnYourCreatureDeaths(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b42Permanent(g, me.ID, "Vindictive Vampire", "Creature — Vampire", b42VindictiveVampireOracle, 2, 3)
	fodder := pushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	life := me.Life
	oppLife := map[uuid.UUID]int{}
	for _, p := range g.Seats[1:] {
		oppLife[p.ID] = p.Life
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(fodder) })
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	b42Permanent(g, me.ID, "Primordial Sage", "Creature — Spirit", b42PrimordialSageOracle, 4, 5)
	hand := len(me.Hand.Cards)
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("the Sage's optional draw: hand %d → %d, want +1", hand, got)
	}
}

func TestB42PridemalkinCountersATargetAndGrantsTrampleToCounteredCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	other := pushVanillaCreature(g, me.ID, "No Counters", 2, 2)
	castCatalogSpell(t, g, "Pridemalkin", "Creature — Cat", b42PridemalkinOracle, nil)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)

	if p := currentPower(t, g, bear); p != 3 {
		t.Errorf("the ETB puts a +1/+1 counter on the target: power %d, want 3", p)
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("a creature you control with a +1/+1 counter has trample")
	}
	if hasEffectiveKeyword(t, g, other, "trample") {
		t.Error("a creature with no counter gets nothing")
	}
}

func TestB42WallOfReverenceGainsLifeEqualToATargetsPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b42Permanent(g, me.ID, "Wall of Reverence", "Creature — Spirit Wall", b42WallOfReverenceOracle, 1, 6)
	fatty := pushVanillaCreature(g, me.ID, "Fatty", 7, 7)
	life := me.Life
	advanceToEndStepOf(t, g, 0)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	pickCard(t, g, me.ID, fatty)
	passPriorityAroundTable(t, g)
	if me.Life != life+7 {
		t.Errorf("gain life equal to the target's power: %d → %d, want +7", life, me.Life)
	}
}

func TestB42ChildrenOfKorlisGivesBackTheTurnsLifeLoss(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	kid := b42Permanent(g, me.ID, "Children of Korlis", "Creature — Human Rebel Cleric", b42ChildrenOfKorlisOracle, 1, 1)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -13) })
	low := me.Life
	if err := g.ActivateCatalogAbility(me.ID, kid, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(kid) {
		t.Error("the sacrifice is a cost, paid at announce")
	}
	passPriorityAroundTable(t, g)
	if me.Life != low+13 {
		t.Errorf("gain the life lost this turn: %d → %d, want +13", low, me.Life)
	}
}

func TestB42ImperviousGreatwurmConvokesAndSurvivesDestruction(t *testing.T) {
	spec, ok := Lookup(b42ImperviousGreatwurmOracle)
	if !ok {
		t.Fatal("Impervious Greatwurm is not registered")
	}
	if spec.TapCost == nil {
		t.Error("convoke is a TapCost, not an alternative cost")
	}
	g := newCatalogGame(t)
	me := g.Seats[0]
	wurm := b42Permanent(g, me.ID, "Impervious Greatwurm", "Creature — Wurm", b42ImperviousGreatwurmOracle, 16, 16)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(wurm) })
	if !g.Battlefield.Contains(wurm) {
		t.Error("indestructible: destruction does nothing")
	}
}

func TestB42TimberwatchElfPumpsByTheElfCount(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elf := b42Permanent(g, me.ID, "Timberwatch Elf", "Creature — Elf", b42TimberwatchElfOracle, 1, 2)
	b42Permanent(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", "", 1, 1)
	// The whole table's Elves count, not only yours.
	b42Permanent(g, opp.ID, "Their Elf", "Creature — Elf", "", 1, 1)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, elf, 0, game.ActivateAbilityParams{
		Targets: b42CardTarget(bear),
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, bear); p != 5 {
		t.Errorf("three Elves on the battlefield is +3/+3: power %d, want 5", p)
	}
}

func TestB42OrcishSiegemasterTramplesTheSwarmAndSwingsForTheBiggest(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	boss := b42Permanent(g, me.ID, "Orcish Siegemaster", "Creature — Orc Soldier", b42OrcishSiegemasterOracle, 0, 5)
	goblin := b42Permanent(g, me.ID, "Goblin", "Creature — Goblin", "", 1, 1)
	elf := b42Permanent(g, me.ID, "Elf", "Creature — Elf", "", 1, 1)
	b42Permanent(g, me.ID, "Fatty", "Creature — Beast", "", 6, 6)
	if !hasEffectiveKeyword(t, g, goblin, "trample") {
		t.Error("other Goblins you control have trample")
	}
	if hasEffectiveKeyword(t, g, elf, "trample") {
		t.Error("only Orcs and Goblins — an Elf gets nothing")
	}
	declareAttack(t, g, opp.ID, boss)
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, boss); p != 6 {
		t.Errorf("+X/+0 where X is your greatest power (6): power %d, want 6", p)
	}
}

func TestB42HighSocietyHunterEatsACreatureAndDrawsOffNontokenDeaths(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hunter := b42Permanent(g, me.ID, "High-Society Hunter", "Creature — Vampire Noble", b42HighSocietyHunterOracle, 5, 3)
	snack := pushVanillaCreature(g, me.ID, "Snack", 1, 1)

	hand := len(me.Hand.Cards)
	declareAttack(t, g, opp.ID, hunter)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	pickCard(t, g, me.ID, snack)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(snack) {
		t.Error("the chosen creature is sacrificed")
	}
	// The sacrifice is a nontoken creature death, so the second
	// ability draws as well.
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("a nontoken creature death draws a card: hand %d → %d", hand, got)
	}
	if p := currentPower(t, g, hunter); p != 6 {
		t.Errorf("and the Hunter gets a +1/+1 counter: power %d, want 6", p)
	}
}

func TestB42HighSocietyHunterIgnoresTokenDeaths(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b42Permanent(g, me.ID, "High-Society Hunter", "Creature — Vampire Noble", b42HighSocietyHunterOracle, 5, 3)
	tokens := pushTokens(g, me.ID, TokenCard("1/1 red Goblin"), 1)
	hand := len(me.Hand.Cards)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(tokens[0]) })
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != hand {
		t.Errorf("a TOKEN death draws nothing: hand %d → %d", hand, got)
	}
}

// --- artifacts and enchantments ------------------------------------

// "One or more creature cards leave your graveyard" is a batch
// trigger: two cards leaving at once is still one Bat.
func TestB42DesecratedTombMakesOneBatPerBatch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b42Permanent(g, me.ID, "Desecrated Tomb", "Artifact", b42DesecratedTombOracle, 0, 0)
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
	passPriorityAroundTable(t, g)
	if got := len(g.Battlefield.Cards); got != before+1 {
		t.Errorf("two creature cards leaving at once make ONE Bat: battlefield %d → %d", before, got)
	}
}

func TestB42SeersSundialOffersTwoManaForACardOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b42Permanent(g, me.ID, "Seer's Sundial", "Artifact", b42SeersSundialOracle, 0, 0)
	hand := len(me.Hand.Cards)
	g.WithWriteLock(func() { _ = g.AddManaForEffect(me.ID, uuid.Nil, "{G}{G}") })
	b12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	passPriorityUntilChoice(t, g)
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != hand {
		t.Errorf("the land left the hand and the paid draw replaced it: hand %d → %d", hand, got)
	}
}

func TestB42DeeprootWatersMakesAHexproofMerfolkPerMerfolkSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b42Permanent(g, me.ID, "Deeproot Waters", "Enchantment", b42DeeprootWatersOracle, 0, 0)
	before := len(g.Battlefield.Cards)
	castCatalogSpell(t, g, "Merfolk Thing", "Creature — Merfolk", "", nil)
	passPriorityAroundTable(t, g)
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
	if !hasEffectiveKeyword(t, g, token.InstanceID, "hexproof") {
		t.Error("the token has hexproof")
	}
}

func TestB42FeveredVisionsDrawsForEveryoneAndBurnsAFullOpponentHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b42Permanent(g, me.ID, "Fevered Visions", "Enchantment", b42FeveredVisionsOracle, 0, 0)
	fillHandTo(t, g, opp, 4)
	// Measured at the end step, after their draw step has already
	// happened — the trigger is on the stack but has not resolved.
	advanceToEndStepOf(t, g, 1)
	life, hand := opp.Life, len(opp.Hand.Cards)
	passPriorityAroundTable(t, g)
	if got := len(opp.Hand.Cards); got != hand+1 {
		t.Errorf("that player draws: hand %d → %d", hand, got)
	}
	if opp.Life != life-2 {
		t.Errorf("four or more cards in an opponent's hand is 2 damage: %d → %d", life, opp.Life)
	}
}

func TestB42WatcherOfTheSpheresDiscountsFliersOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b42Permanent(g, me.ID, "Watcher of the Spheres", "Creature — Bird Wizard", b42WatcherOfTheSpheresOracle, 2, 2)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ring := b42Permanent(g, me.ID, "Ring of the Lucii", "Legendary Artifact", b42RingOfTheLuciiOracle, 0, 0)
	if err := g.ActivateManaAbility(me.ID, ring, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if len(me.ManaPool) != 2 {
		t.Errorf("{T}: Add {C}{C} adds two: pool %v", me.ManaPool)
	}

	g.WithWriteLock(func() {
		_ = g.UntapTargetForEffect(ring)
		_ = g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}")
	})
	victim := pushVanillaCreature(g, opp.ID, "Untapped", 2, 2)
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, ring, 0, game.ActivateAbilityParams{
		Targets: b42CardTarget(victim),
	}); err != nil {
		t.Fatalf("activate the tap-down: %v", err)
	}
	if me.Life != life-1 {
		t.Errorf("the life is a cost, paid at announce: %d → %d", life, me.Life)
	}
	passPriorityAroundTable(t, g)
	if !tappedOnBattlefield(t, g, victim) {
		t.Error("the target is tapped")
	}
}

// --- land ----------------------------------------------------------

func TestB42AlpineMeadowEntersTappedAndTapsForRedOrWhite(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := b12PlayFromHand(t, g, "Alpine Meadow", "Snow Land — Mountain Plains", b42AlpineMeadowOracle, game.CastSpellParams{})
	if !b16Tapped(t, g, land) {
		t.Fatal("Alpine Meadow enters tapped")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	b28TapForMana(t, g, me.ID, land, "W")
	if got := poolColors(me); len(got) != 1 || got[0] != "W" {
		t.Errorf("tapped for {W}: pool %v", got)
	}
}
