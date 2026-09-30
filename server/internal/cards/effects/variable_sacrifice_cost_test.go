package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// variable_sacrifice_cost_test.go — ADR 0100 sub-PR 4: a variable
// sacrifice count as a cast's additional cost (CR 601.2b, 601.2f–h,
// 107.3a). The rules the board shows — what the announcement may name,
// what the count prices and what the resolution reads back — and one
// test per card.

const (
	viciousBetrayalOracle         = "be1015a7-2ace-4b97-8884-202abae8401b"
	devouringRageOracle           = "19a8c2d3-b06e-4a76-bb28-8928ce84186c"
	devouringGreedOracle          = "efbe96b5-1f58-4818-a346-ea554813adbc"
	eliminateTheCompetitionOracle = "26e549c0-a08b-475b-9138-6dde175cdf55"
	immoralBargainOracle          = "c82cc78b-ffd6-4f72-881d-85913830feb2"
	devastatingSummonsOracle      = "5eb6626a-1e0a-438e-9e95-a1e86be6489d"
	plumbTheForbiddenOracle       = "b099fc54-cdbc-46cb-b4e4-7c2ca77b115b"
	torgaarOracle                 = "4229140f-fa5b-4727-a16a-cbbd756d979e"
	dargoOracle                   = "a2414a2d-0fd8-4084-8762-ace48f70c853"
	rottenmouthViperOracle        = "2c75623c-59f4-4449-ab43-9d1225185ad9"
)

// vsPermanent puts a permanent onto the battlefield under `owner` and
// returns its ID.
func vsPermanent(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	return eitherPermanent(g, owner, name, typeLine)
}

func vsCreatures(g *game.Game, owner uuid.UUID, n int, typeLine string) []uuid.UUID {
	out := make([]uuid.UUID, n)
	for i := range out {
		out[i] = vsPermanent(g, owner, "Fodder", typeLine)
	}
	return out
}

func vsTarget(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// --- the announcement ---------------------------------------------------

// Vicious Betrayal: any number, zero included. Two creatures sacrificed
// with the spell on the stack, recorded, and read back as +4/+4.
func TestViciousBetrayalPumpsForEachSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fodder := vsCreatures(g, me.ID, 2, "Creature — Goblin")
	hero := vsPermanent(g, me.ID, "Hero", "Creature — Human")
	id, err := castWithTapParams(t, g, "Vicious Betrayal", "Sorcery", "{3}{B}{B}", viciousBetrayalOracle,
		game.CastSpellParams{Targets: vsTarget(hero), SacrificeIDs: fodder})
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	for _, f := range fodder {
		if onBattlefield(g, f) {
			t.Fatal("a sacrificed creature is still on the battlefield with the spell on the stack")
		}
	}
	if got := g.StackMeta[id].Paid.Sacrificed; got != 2 {
		t.Fatalf("Paid.Sacrificed = %d, want 2", got)
	}
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCardByID(g, hero)
	if c.CurrentPower() != 6 || c.CurrentToughness() != 6 {
		t.Fatalf("Hero is %d/%d, want 6/6 (2/2 with +4/+4)", c.CurrentPower(), c.CurrentToughness())
	}
}

// Zero is a legal payment of "any number", and pumps nothing.
func TestViciousBetrayalWithNothingSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hero := vsPermanent(g, me.ID, "Hero", "Creature — Human")
	id, err := castWithTapParams(t, g, "Vicious Betrayal", "Sorcery", "{3}{B}{B}", viciousBetrayalOracle,
		game.CastSpellParams{Targets: vsTarget(hero)})
	if err != nil {
		t.Fatalf("CastSpell with no sacrifice: %v", err)
	}
	if got := g.StackMeta[id].Paid.Sacrificed; got != 0 {
		t.Fatalf("Paid.Sacrificed = %d, want 0", got)
	}
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCardByID(g, hero)
	if c.CurrentPower() != 2 {
		t.Fatalf("Hero's power = %d, want 2", c.CurrentPower())
	}
}

// What the list may name: only creatures (the clause's predicate), only
// the caster's own (CR 701.21a), each once. Nothing moves on a refusal.
func TestVariableSacrificeRefusals(t *testing.T) {
	g := newCatalogGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	mine := vsPermanent(g, me.ID, "Mine", "Creature — Goblin")
	theirs := vsPermanent(g, foe.ID, "Theirs", "Creature — Goblin")
	relic := vsPermanent(g, me.ID, "Relic", "Artifact")
	hero := vsPermanent(g, me.ID, "Hero", "Creature — Human")
	for _, tc := range []struct {
		why string
		ids []uuid.UUID
	}{
		{"an opponent's creature", []uuid.UUID{theirs}},
		{"a noncreature", []uuid.UUID{relic}},
		{"the same creature twice", []uuid.UUID{mine, mine}},
	} {
		if _, err := castWithTapParams(t, g, "Vicious Betrayal", "Sorcery", "{3}{B}{B}", viciousBetrayalOracle,
			game.CastSpellParams{Targets: vsTarget(hero), SacrificeIDs: tc.ids}); err == nil {
			t.Errorf("%s: cast accepted", tc.why)
		}
	}
	for _, id := range []uuid.UUID{mine, theirs, relic} {
		if !onBattlefield(g, id) {
			t.Fatal("a refused cast sacrificed something")
		}
	}
}

// Devouring Rage: +3/+0, and +3/+0 more per Spirit. Only Spirits pay.
func TestDevouringRageCountsSpirits(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spirit := vsPermanent(g, me.ID, "Kami", "Creature — Spirit")
	goblin := vsPermanent(g, me.ID, "Goblin", "Creature — Goblin")
	if _, err := castWithTapParams(t, g, "Devouring Rage", "Instant — Arcane", "{4}{R}", devouringRageOracle,
		game.CastSpellParams{Targets: vsTarget(goblin), SacrificeIDs: []uuid.UUID{goblin}}); err == nil {
		t.Fatal("a non-Spirit paid Devouring Rage's sacrifice")
	}
	if _, err := castWithTapParams(t, g, "Devouring Rage", "Instant — Arcane", "{4}{R}", devouringRageOracle,
		game.CastSpellParams{Targets: vsTarget(goblin), SacrificeIDs: []uuid.UUID{spirit}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCardByID(g, goblin)
	if c.CurrentPower() != 8 || c.CurrentToughness() != 2 {
		t.Fatalf("Goblin is %d/%d, want 8/2 (+6/+0)", c.CurrentPower(), c.CurrentToughness())
	}
}

// Devouring Greed: the target loses 2 plus 2 per Spirit, and you gain
// what was lost.
func TestDevouringGreedDrainsPerSpirit(t *testing.T) {
	g := newCatalogGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	spirits := vsCreatures(g, me.ID, 2, "Creature — Spirit")
	myLife, foeLife := me.Life, foe.Life
	if _, err := castWithTapParams(t, g, "Devouring Greed", "Sorcery — Arcane", "{2}{B}{B}", devouringGreedOracle,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: foe.ID}}, SacrificeIDs: spirits}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := foeLife - foe.Life; got != 6 {
		t.Errorf("target lost %d, want 6", got)
	}
	if got := me.Life - myLife; got != 6 {
		t.Errorf("caster gained %d, want 6", got)
	}
}

// --- "sacrifice X" --------------------------------------------------------

// Eliminate the Competition: X creatures sacrificed and X targets, one
// announced X (CR 107.3i). A count that disagrees with X is refused.
func TestEliminateTheCompetitionSacrificesX(t *testing.T) {
	g := newCatalogGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	fodder := vsCreatures(g, me.ID, 2, "Creature — Goblin")
	victims := vsCreatures(g, foe.ID, 2, "Creature — Ogre")
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: victims[0]}, {Kind: game.TargetCard, ID: victims[1]}}
	for _, tc := range []struct {
		why    string
		params game.CastSpellParams
	}{
		{"one sacrifice at X = 2", game.CastSpellParams{XValue: 2, Targets: targets, SacrificeIDs: fodder[:1]}},
		{"three sacrifices at X = 2", game.CastSpellParams{XValue: 2, Targets: targets, SacrificeIDs: append(append([]uuid.UUID(nil), fodder...), victims[0])}},
		{"two sacrifices at X = 1", game.CastSpellParams{XValue: 1, Targets: targets[:1], SacrificeIDs: fodder}},
	} {
		if _, err := castWithTapParams(t, g, "Eliminate the Competition", "Sorcery", "{4}{B}", eliminateTheCompetitionOracle, tc.params); err == nil {
			t.Errorf("%s: cast accepted", tc.why)
		}
	}
	if _, err := castWithTapParams(t, g, "Eliminate the Competition", "Sorcery", "{4}{B}", eliminateTheCompetitionOracle,
		game.CastSpellParams{XValue: 2, Targets: targets, SacrificeIDs: fodder}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range append(fodder, victims...) {
		if onBattlefield(g, id) {
			t.Fatal("a sacrificed or targeted creature survived")
		}
	}
}

// Immoral Bargain: the X targets are any nonland permanents.
func TestImmoralBargainDestroysXNonlandPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	goblin := vsPermanent(g, me.ID, "Goblin", "Creature — Goblin")
	relic := vsPermanent(g, foe.ID, "Relic", "Artifact")
	if _, err := castWithTapParams(t, g, "Immoral Bargain", "Sorcery", "{1}{B}{G}", immoralBargainOracle,
		game.CastSpellParams{XValue: 1, Targets: vsTarget(relic), SacrificeIDs: []uuid.UUID{goblin}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if onBattlefield(g, relic) || onBattlefield(g, goblin) {
		t.Fatal("the artifact or the sacrificed Goblin is still on the battlefield")
	}
}

// Devastating Summons: X lands, two X/X Elementals.
func TestDevastatingSummonsMakesTwoXByX(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lands := vsCreatures(g, me.ID, 3, "Basic Land — Mountain")
	if _, err := castWithTapParams(t, g, "Devastating Summons", "Sorcery", "{R}", devastatingSummonsOracle,
		game.CastSpellParams{XValue: 3, SacrificeIDs: lands[:2]}); err == nil {
		t.Fatal("two lands paid for X = 3")
	}
	if _, err := castWithTapParams(t, g, "Devastating Summons", "Sorcery", "{R}", devastatingSummonsOracle,
		game.CastSpellParams{XValue: 3, SacrificeIDs: lands}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Elemental" && c.Controller == me.ID {
			n++
			if c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
				t.Errorf("Elemental is %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
			}
		}
	}
	if n != 2 {
		t.Fatalf("%d Elementals, want 2", n)
	}
}

// --- Plumb the Forbidden's "when you do" ---------------------------------

// Two creatures sacrificed: the reflexive trigger copies the spell
// twice, so the caster draws three and loses three. None sacrificed:
// no trigger, one draw.
func TestPlumbTheForbiddenCopiesPerSacrifice(t *testing.T) {
	for _, n := range []int{0, 2} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		fodder := vsCreatures(g, me.ID, n, "Creature — Goblin")
		life := me.Life
		if _, err := castWithTapParams(t, g, "Plumb the Forbidden", "Instant", "{1}{B}", plumbTheForbiddenOracle,
			game.CastSpellParams{SacrificeIDs: fodder}); err != nil {
			t.Fatalf("n=%d: CastSpell: %v", n, err)
		}
		hand := me.Hand.Size()
		passPriorityAroundTable(t, g)
		if got := me.Hand.Size() - hand; got != n+1 {
			t.Errorf("n=%d: drew %d, want %d", n, got, n+1)
		}
		if got := life - me.Life; got != n+1 {
			t.Errorf("n=%d: lost %d life, want %d", n, got, n+1)
		}
	}
}

// --- the price ------------------------------------------------------------

// Torgaar: {2} less per creature sacrificed, read at CR 601.2f from the
// announced count; the reduction stops at the generic, so a fourth
// creature buys nothing.
func TestTorgaarPricesTheSacrifices(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fodder := vsCreatures(g, me.ID, 4, "Creature — Goblin")
	torgaar := game.Card{InstanceID: uuid.New(), Name: "Torgaar, Famine Incarnate", TypeLine: "Legendary Creature — Avatar",
		ManaCost: "{6}{B}{B}", OracleID: torgaarOracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(torgaar)
	for n, want := range []int{8, 6, 4, 2, 2} {
		price, err := g.PriceCast(me.ID, torgaar, game.CastSpellParams{SacrificeIDs: fodder[:n]})
		if err != nil {
			t.Fatalf("PriceCast: %v", err)
		}
		if got := price.Total.ManaValue(); got != want {
			t.Errorf("%d sacrificed: total %s, want mana value %d", n, price.Total.String(), want)
		}
	}
}

// The discount is charged: a strict cast with {B}{B} floating and three
// creatures sacrificed pays exactly that. Torgaar's entry then sets a
// target player to half the starting life total.
func TestTorgaarStrictCastAndEntry(t *testing.T) {
	g := newCatalogGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	fodder := vsCreatures(g, me.ID, 3, "Creature — Goblin")
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	id, err := castWithTapParams(t, g, "Torgaar, Famine Incarnate", "Legendary Creature — Avatar", "{6}{B}{B}", torgaarOracle,
		game.CastSpellParams{Strict: true, SacrificeIDs: fodder})
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if n := len(me.ManaPool); n != 0 {
		t.Fatalf("%d mana left floating", n)
	}
	foe.Life = 37
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, id) {
		t.Fatal("Torgaar did not enter")
	}
	var prompt *game.PendingChoice
	for _, choice := range g.PendingChoices {
		if choice != nil && choice.Kind == game.PendingChoicePickTarget {
			prompt = choice
			break
		}
	}
	if prompt == nil {
		t.Fatal("Torgaar did not ask for its entry trigger's target")
	}
	if err := g.ResolvePickTarget(prompt.ID, prompt.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: foe.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	if want := startingLifeOf(g) / 2; foe.Life != want {
		t.Fatalf("target's life = %d, want %d", foe.Life, want)
	}
}

// Dargo: {2} less per permanent sacrificed for it, and {2} less per
// OTHER artifact or creature sacrificed earlier this turn. A land
// sacrificed earlier does not count.
func TestDargoCountsBothDiscounts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fodder := vsCreatures(g, me.ID, 2, "Artifact")
	earlier := vsPermanent(g, me.ID, "Earlier", "Creature — Goblin")
	land := vsPermanent(g, me.ID, "Waste", "Land")
	dargo := game.Card{InstanceID: uuid.New(), Name: "Dargo, the Shipwrecker", TypeLine: "Legendary Creature — Giant Pirate",
		ManaCost: "{6}{R}", OracleID: dargoOracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(dargo)
	price := func() int {
		t.Helper()
		p, err := g.PriceCast(me.ID, dargo, game.CastSpellParams{SacrificeIDs: fodder})
		if err != nil {
			t.Fatalf("PriceCast: %v", err)
		}
		return p.Total.ManaValue()
	}
	if got := price(); got != 3 {
		t.Fatalf("two sacrificed, none earlier: mana value %d, want 3", got)
	}
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{land, earlier} {
			if err := g.SacrificePermanentForEffect(id); err != nil {
				t.Fatalf("sacrifice: %v", err)
			}
		}
	})
	if got := price(); got != 1 {
		t.Fatalf("two sacrificed, a creature and a land earlier: mana value %d, want 1", got)
	}
	// And the charge agrees: {R} floating pays for it. The cost's own
	// sacrifices land in the tally only after the total is settled, so
	// they are not counted twice.
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	if err := g.CastSpell(me.ID, dargo.InstanceID, game.CastSpellParams{Strict: true, SacrificeIDs: fodder}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if n := len(me.ManaPool); n != 0 {
		t.Fatalf("%d mana left floating", n)
	}
	if got := g.TurnTallyFor(me.ID).ArtifactsOrCreaturesSacrificed; got != 3 {
		t.Fatalf("tally = %d, want 3 (the earlier creature and the two artifacts, not the land)", got)
	}
}

// Rottenmouth Viper: {1} less per nonland permanent sacrificed; the
// entry puts a blight counter on and asks each opponent once per
// counter. Every opponent answering "lose 4 life" loses 4 per counter.
func TestRottenmouthViperPricesAndPunishes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fodder := []uuid.UUID{
		vsPermanent(g, me.ID, "Relic", "Artifact"),
		vsPermanent(g, me.ID, "Goblin", "Creature — Goblin"),
		vsPermanent(g, me.ID, "Charm", "Enchantment"),
	}
	land := vsPermanent(g, me.ID, "Waste", "Land")
	viper := game.Card{InstanceID: uuid.New(), Name: "Rottenmouth Viper", TypeLine: "Creature — Elemental Snake",
		ManaCost: "{5}{B}", OracleID: rottenmouthViperOracle, Owner: me.ID, Controller: me.ID, Power: 6, Toughness: 6}
	me.Hand.PushTop(viper)
	price, err := g.PriceCast(me.ID, viper, game.CastSpellParams{SacrificeIDs: fodder})
	if err != nil {
		t.Fatalf("PriceCast: %v", err)
	}
	if got := price.Total.ManaValue(); got != 3 {
		t.Fatalf("three sacrificed: total %s, want mana value 3", price.Total.String())
	}
	if _, err := castWithTapParams(t, g, "Rottenmouth Viper", "Creature — Elemental Snake", "{5}{B}", rottenmouthViperOracle,
		game.CastSpellParams{SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Fatal("a land paid \"any number of nonland permanents\"")
	}
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats[1:] {
		p.Hand.Cards = nil
		lives[p.ID] = p.Life
	}
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	me.Hand.PushTop(viper)
	if err := g.CastSpell(me.ID, viper.InstanceID, game.CastSpellParams{SacrificeIDs: fodder}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	drainRemainingTormentPrompts(t, g)
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCardByID(g, viper.InstanceID)
	if !ok || c.Counters["blight"] != 1 {
		t.Fatalf("Viper on the battlefield = %v with %d blight counters, want one", ok, c.Counters["blight"])
	}
	for _, p := range g.Seats[1:] {
		if got := lives[p.ID] - p.Life; got != 4 {
			t.Errorf("opponent %s lost %d, want 4 (one blight counter)", p.Name, got)
		}
	}
}

// --- the boot guard -------------------------------------------------------

// ADR 0100 §3's cross-slot rule and the shapes that stay refused. Each
// compiles and would be paid as something no card prints.
func TestRegisterRefusesVariableSacrificePlans(t *testing.T) {
	anyNumber := func() *game.TargetSpec { return sacrificeSpec("any number of creatures", Creature()).WithCount(0, 0) }
	noop := func(*game.Game, *game.StackItem) error { return nil }
	cases := []struct {
		name string
		spec Spec
		want string
	}{
		{"any number beside a sacrificing kicker", Spec{OracleID: "vs-guard-kicker", Name: "vs-guard-kicker",
			AdditionalCost: SacrificeAnyNumberCost("any number of creatures", Creature()),
			OptionalCosts:  []game.AdditionalCost{KickerSacrifice("a land", Land())},
		}, "one variable sacrifice clause"},
		{"sacrifice X beside pay X life", Spec{OracleID: "vs-guard-life", Name: "vs-guard-life",
			AdditionalCost: func() *game.AdditionalCost {
				c := SacrificeXCost("X creatures", Creature())
				c.PayLifeX = true
				return c
			}(),
		}, "pays X life"},
		{"any number in an optional cost", Spec{OracleID: "vs-guard-optional", Name: "vs-guard-optional",
			OptionalCosts: []game.AdditionalCost{{Optional: true, Key: game.KickerKey, Label: "Kicker—Sacrifice any number of creatures", Sacrifice: anyNumber()}},
		}, "at least one"},
		{"any number in an either/or branch", Spec{OracleID: "vs-guard-branch", Name: "vs-guard-branch",
			AdditionalCost: EitherCost(SacrificeAnyNumberCost("any number of creatures", Creature()).Keyed("sac"), DiscardCost(1).Keyed("discard")),
		}, "at least one"},
		{"any number on an activated ability", Spec{OracleID: "vs-guard-ability", Name: "vs-guard-ability", Activated: []ActivatedAbility{{
			Label: "x", Cost: game.AbilityCost{SacrificeOther: anyNumber()}, Effect: noop,
		}}}, "at least one"},
		{"sacrifice X in an optional cost", Spec{OracleID: "vs-guard-optional-x", Name: "vs-guard-optional-x",
			OptionalCosts: []game.AdditionalCost{{Optional: true, Key: game.KickerKey, Label: "Kicker—Sacrifice X creatures", Sacrifice: SacrificeXCost("X creatures", Creature()).Sacrifice}},
		}, "no X to announce"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustPanic(t, tc.want, func() { Register(tc.spec) })
			if Has(tc.spec.OracleID) {
				t.Error("a refused spec was registered anyway")
			}
		})
	}
}

// The shapes the cast's mandatory slot accepts: "sacrifice X" and "any
// number".
func TestRegisterAcceptsCastVariableSacrifices(t *testing.T) {
	for _, spec := range []Spec{
		{OracleID: "vs-ok-x", Name: "vs-ok-x", AdditionalCost: SacrificeXCost("X creatures", Creature())},
		{OracleID: "vs-ok-any", Name: "vs-ok-any", AdditionalCost: SacrificeAnyNumberCost("any number of creatures", Creature()),
			OptionalCosts: []game.AdditionalCost{Kicker("{2}")}},
	} {
		registerForTest(t, spec)
		if !Has(spec.OracleID) {
			t.Fatalf("%s was not registered", spec.Name)
		}
	}
}

// The any-number bounds, beside the #1213 shapes.
func TestSacrificeAnyNumberBounds(t *testing.T) {
	spec := SacrificeAnyNumberCost("any number of creatures", Creature()).Sacrifice
	if !game.SacrificeAnyNumber(spec) || !game.SacrificeCostVariable(spec) {
		t.Fatal("an any-number clause does not read as variable")
	}
	if lo, hi := game.SacrificeCostBounds(spec, 0); lo != 0 || hi != 0 {
		t.Fatalf("bounds = %d/%d, want 0/0 (open)", lo, hi)
	}
	for _, n := range []int{0, 1, 7} {
		if !game.SacrificeCountLegal(spec, 0, n) {
			t.Errorf("naming %d is refused", n)
		}
	}
	if got := game.SacrificeCostCount(spec); got != 0 {
		t.Errorf("SacrificeCostCount = %d, want the floor, 0", got)
	}
}
