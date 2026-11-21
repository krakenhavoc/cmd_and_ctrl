package testscm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Every printed static of the seam, each against the spells it names
// and the nearest ones it does not. The verb is the one every catalog
// counterspell calls.
func TestPrintedCounterShieldsCoverWhatTheyPrint(t *testing.T) {
	mine := func(c game.Card) effectstest.CsSpell { return effectstest.CsSpell{Card: c} }
	theirs := func(c game.Card) effectstest.CsSpell { return effectstest.CsSpell{Card: c, Caster: 1, Controller: 1} }
	for _, tc := range []struct {
		name, typeLine, oracle string
		covered, exposed       []effectstest.CsSpell
	}{
		{"Chimil, the Inner Sun", "Legendary Artifact", effectstest.ChimilOracle,
			[]effectstest.CsSpell{mine(effectstest.CsBlueInstant), mine(effectstest.CsBear), {Card: effectstest.CsSorcery, Copy: true}, {Card: effectstest.CsBlueInstant, Caster: 1}},
			[]effectstest.CsSpell{theirs(effectstest.CsBlueInstant), {Card: effectstest.CsBlueInstant, Controller: 1}}},
		{"Gaea's Herald", "Creature — Elf", effectstest.CsGaeasHeraldOracle,
			[]effectstest.CsSpell{mine(effectstest.CsBear), theirs(effectstest.CsBear)},
			[]effectstest.CsSpell{mine(effectstest.CsBlueInstant), theirs(effectstest.CsEnchantment)}},
		{"Leyline of Lifeforce", "Enchantment", effectstest.CsLeylineOfLifeforceOracle,
			[]effectstest.CsSpell{mine(effectstest.CsBigCreature), theirs(effectstest.CsBear)},
			[]effectstest.CsSpell{theirs(effectstest.CsSorcery)}},
		{"Prowling Serpopard", "Creature — Cat Snake", effectstest.CsProwlingSerpopardOracle,
			[]effectstest.CsSpell{mine(effectstest.CsBear)},
			[]effectstest.CsSpell{theirs(effectstest.CsBear), mine(effectstest.CsBlueInstant)}},
		{"Root Sliver", "Creature — Sliver", effectstest.CsRootSliverOracle,
			[]effectstest.CsSpell{mine(effectstest.CsSliver), theirs(effectstest.CsSliver), theirs(effectstest.CsChangeling)},
			[]effectstest.CsSpell{mine(effectstest.CsBear)}},
		{"Sphinx of the Final Word", "Creature — Sphinx", effectstest.CsSphinxFinalWordOracle,
			[]effectstest.CsSpell{mine(effectstest.CsBlueInstant), mine(effectstest.CsSorcery)},
			[]effectstest.CsSpell{mine(effectstest.CsBear), theirs(effectstest.CsBlueInstant)}},
		{"Spellbreaker Behemoth", "Creature — Beast", effectstest.CsSpellbreakerOracle,
			[]effectstest.CsSpell{mine(effectstest.CsBigCreature)},
			[]effectstest.CsSpell{mine(effectstest.CsBear), theirs(effectstest.CsBigCreature), mine(effectstest.CsFiveDrop)}},
		{"Surrak Dragonclaw", "Legendary Creature — Human Warrior", effectstest.CsSurrakOracle,
			[]effectstest.CsSpell{mine(effectstest.CsBear)},
			[]effectstest.CsSpell{mine(effectstest.CsBlueInstant), theirs(effectstest.CsBear)}},
		{"Allosaurus Shepherd", "Creature — Elf Shaman", effectstest.CsAllosaurusShepherdOracle,
			[]effectstest.CsSpell{mine(effectstest.CsGreenInstant), mine(effectstest.CsBear)},
			[]effectstest.CsSpell{mine(effectstest.CsBlueInstant), theirs(effectstest.CsGreenInstant)}},
		{"Hexing Squelcher", "Creature — Goblin Sorcerer", effectstest.CsHexingSquelcherOracle,
			[]effectstest.CsSpell{mine(effectstest.CsBlueInstant)},
			[]effectstest.CsSpell{theirs(effectstest.CsBlueInstant)}},
		{"Destiny Spinner", "Enchantment Creature — Human", effectstest.CsDestinySpinnerOracle,
			[]effectstest.CsSpell{mine(effectstest.CsBear), mine(effectstest.CsEnchantment)},
			[]effectstest.CsSpell{mine(effectstest.CsBlueInstant), theirs(effectstest.CsBear)}},
		// "Spells you cast": the caster, through a change of control, and
		// never a copy (CR 707.10).
		{"Thryx, the Sudden Storm", "Legendary Creature — Elemental Giant", effectstest.CsThryxOracle,
			[]effectstest.CsSpell{mine(effectstest.CsFiveDrop), mine(effectstest.CsBigCreature), {Card: effectstest.CsFiveDrop, Controller: 1}},
			[]effectstest.CsSpell{mine(effectstest.CsFourDrop), {Card: effectstest.CsFiveDrop, Copy: true}, theirs(effectstest.CsFiveDrop), {Card: effectstest.CsFiveDrop, Caster: 1}}},
		{"Cunning Nightbonder", "Creature — Human Rogue", effectstest.CsCunningNightbonderOracle,
			[]effectstest.CsSpell{mine(effectstest.CsFlashBeast)},
			[]effectstest.CsSpell{mine(effectstest.CsBear), {Card: effectstest.CsFlashBeast, Copy: true}, theirs(effectstest.CsFlashBeast)}},
		{"Lier, Disciple of the Drowned", "Legendary Creature — Human Wizard", effectstest.CsLierOracle,
			[]effectstest.CsSpell{mine(effectstest.CsBlueInstant), theirs(effectstest.CsBlueInstant), theirs(effectstest.CsBear)},
			nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for want, spells := range map[bool][]effectstest.CsSpell{true: tc.covered, false: tc.exposed} {
				for _, s := range spells {
					g := effectstest.NewCatalogGame(t)
					effectstest.PushCatalogPermanent(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, false)
					id := effectstest.CsPushSpell(g, s)
					if got := effectstest.CounterByEffect(t, g, id); got != want {
						t.Errorf("%s (cast by seat %d, controlled by seat %d, copy %v): survived the counter = %v, want %v",
							s.Card.Name, s.Caster, s.Controller, s.Copy, got, want)
					}
				}
			}
		})
	}
}

// The headline, through a real Counterspell: with Chimil out, the
// opponent's Counterspell resolves and does nothing, and the creature
// resolves after it.
func TestChimilStopsARealCounterspell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Chimil, the Inner Sun", "Legendary Artifact", effectstest.ChimilOracle, false)

	bear := effectstest.CastFromHandForTest(t, g, me, "Grizzly Bears", "Creature — Bear", "{1}{G}", "", game.CastSpellParams{})
	if !g.SpellCantBeCounteredForEffect(bear) {
		t.Error("the stack chip should say the Bears can't be countered")
	}
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	counterspell := effectstest.Batch01OpponentCasts(t, g, opp, "Counterspell", effectstest.B14CounterspellOracle, "{U}{U}",
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(bear) {
		t.Fatal("Chimil's controller's creature spell was countered")
	}
	if !opp.Graveyard.Contains(counterspell) {
		t.Error("the Counterspell resolves (and does nothing) — it should be in the graveyard, not fizzled away elsewhere")
	}
}

// CR 613.1f: a Chimil that has lost all its abilities shields nothing.
func TestChimilWithoutItsAbilitiesShieldsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	chimil := effectstest.PushCatalogPermanent(g, me.ID, "Chimil, the Inner Sun", "Legendary Artifact", effectstest.ChimilOracle, false)
	g.WithWriteLock(func() {
		if !g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(chimil),
			[]game.Mod{game.LoseAllAbilitiesMod()}, game.IndefiniteDuration(), "test — loses all abilities") {
			t.Fatal("setup: the removal registered nothing")
		}
	})
	// Bring the layers up to date, as any priority pass would.
	g.ReadSnapshot(func() {})
	spell := effectstest.CsPushSpell(g, effectstest.CsSpell{Card: effectstest.CsBlueInstant})
	if effectstest.CounterByEffect(t, g, spell) {
		t.Error("a silenced Chimil still shielded the spell")
	}
}

// Thryx's and the Nightbonder's sentence is two statics; the discount
// half is real under strict mana.
func TestYouCastShieldCardsAlsoDiscount(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		spell                  game.Card
		pool                   string
	}{
		{"Thryx, the Sudden Storm", "Legendary Creature — Elemental Giant", effectstest.CsThryxOracle, effectstest.CsFiveDrop, "UUUU"},
		{"Cunning Nightbonder", "Creature — Human Rogue", effectstest.CsCunningNightbonderOracle, effectstest.CsFlashBeast, "G"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.AdvanceToMain(t, g)
			effectstest.PushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
			effectstest.FloatForTest(g, me, tc.pool)
			id := uuid.New()
			c := tc.spell
			c.InstanceID, c.Owner, c.Controller = id, me.ID, me.ID
			g.WithWriteLock(func() { me.Hand.PushTop(c) })
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
				t.Fatalf("cast %s for {1} less: %v", c.Name, err)
			}
			if !g.SpellCantBeCounteredForEffect(id) {
				t.Error("the spell the static discounted is also uncounterable")
			}
		})
	}
}

// Allosaurus Shepherd's {4}{G}{G}: each Elf creature its activator
// controls is a 5/5 Dinosaur until end of turn — the set fixed as it
// resolves (CR 611.2c), and nobody else's Elves.
func TestAllosaurusShepherdMakesYourElvesFiveFiveDinosaurs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	shepherd := effectstest.PushCatalogPermanent(g, me.ID, "Allosaurus Shepherd", "Creature — Elf Shaman", effectstest.CsAllosaurusShepherdOracle, false)
	myElf := effectstest.PushCatalogPermanent(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", "", false)
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Grizzly Bears", "Creature — Bear", "", false)
	theirElf := effectstest.PushCatalogPermanent(g, opp.ID, "Elvish Mystic", "Creature — Elf Druid", "", false)

	effectstest.FloatForTest(g, me, "GGGGGG")
	if err := g.ActivateCatalogAbility(me.ID, shepherd, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	late := effectstest.PushCatalogPermanent(g, me.ID, "Fyndhorn Elves", "Creature — Elf Druid", "", false)

	for _, id := range []uuid.UUID{shepherd, myElf} {
		if p, tough := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != 5 || tough != 5 {
			t.Errorf("an Elf you control is %d/%d, want 5/5", p, tough)
		}
		if !effectstest.HasAbility(effectstest.EffectiveSubtypes(t, g, id), "Dinosaur") {
			t.Errorf("an Elf you control is not a Dinosaur: %v", effectstest.EffectiveSubtypes(t, g, id))
		}
	}
	for _, id := range []uuid.UUID{bear, theirElf, late} {
		if p := effectstest.EffectivePower(t, g, id); p != 1 {
			t.Errorf("a creature outside the set was changed: power %d", p)
		}
	}
}

// Destiny Spinner's {3}{G}: the land is an X/X Elemental creature with
// trample and haste, X the enchantments its activator controls as it
// resolves, and still a land.
func TestDestinySpinnerAnimatesALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	spinner := effectstest.PushCatalogPermanent(g, me.ID, "Destiny Spinner", "Enchantment Creature — Human", effectstest.CsDestinySpinnerOracle, false)
	effectstest.PushCatalogPermanent(g, me.ID, "Honden of Seeing Winds", "Legendary Enchantment — Shrine", "", false)
	effectstest.PushCatalogPermanent(g, me.ID, "Glorious Anthem", "Enchantment", "", false)
	forest := effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)

	effectstest.FloatForTest(g, me, "GGGG")
	if err := g.ActivateCatalogAbility(me.ID, spinner, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(forest)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if p, tough := effectstest.EffectivePower(t, g, forest), effectstest.EffectiveToughness(t, g, forest); p != 3 || tough != 3 {
		t.Errorf("the Forest is %d/%d, want 3/3 for three enchantments", p, tough)
	}
	types := effectstest.EffectiveTypes(t, g, forest)
	if !effectstest.HasAbility(types, "Creature") || !effectstest.HasAbility(types, "Land") {
		t.Errorf("the Forest's types are %v, want a land creature", types)
	}
	if !effectstest.HasAbility(effectstest.EffectiveSubtypes(t, g, forest), "Elemental") {
		t.Errorf("the Forest is not an Elemental: %v", effectstest.EffectiveSubtypes(t, g, forest))
	}
	abilities := effectstest.EffectiveAbilities(t, g, forest)
	if !effectstest.HasAbility(abilities, "trample") || !effectstest.HasAbility(abilities, "haste") {
		t.Errorf("the Forest's abilities are %v, want trample and haste", abilities)
	}
}

// Hexing Squelcher gives its controller's OTHER creatures ward — pay 2
// life, charged to the opponent who targets one.
func TestHexingSquelcherGrantsLifeWardToYourOtherCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Hexing Squelcher", "Creature — Goblin Sorcerer", effectstest.CsHexingSquelcherOracle, false)
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Grizzly Bears", "Creature — Bear", "", false)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	effectstest.CastAtWardedCreature(t, g, opp, bear)
	prompt := effectstest.PassUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the opponent who targeted the Bears")
	}
	if prompt.LifeCost != 2 {
		t.Errorf("the granted ward charges %d life, want 2", prompt.LifeCost)
	}
}

// Surrak's other creatures have trample; Surrak itself does not.
func TestSurrakGivesOtherCreaturesTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	surrak := effectstest.PushCatalogPermanent(g, me.ID, "Surrak Dragonclaw", "Legendary Creature — Human Warrior", effectstest.CsSurrakOracle, false)
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Grizzly Bears", "Creature — Bear", "", false)
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Hill Giant", "Creature — Giant", "", false)
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, bear), "trample") {
		t.Error("Surrak's controller's other creature has no trample")
	}
	if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, surrak), "trample") || effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, theirs), "trample") {
		t.Error("trample reached Surrak itself or an opponent's creature")
	}
}

// Lier gives each instant and sorcery card in its controller's
// graveyard flashback for its mana cost, while Lier remains — and the
// card is exiled as it resolves (CR 702.34a).
func TestLierGivesYourGraveyardFlashback(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMain(t, g)
	lier := effectstest.PushCatalogPermanent(g, me.ID, "Lier, Disciple of the Drowned", "Legendary Creature — Human Wizard", effectstest.CsLierOracle, false)
	sorcery := effectstest.SeedGraveyardCard(t, g, "Dead Sorcery", "Sorcery", effectstest.CsTestGraveyardSorceryOracl)
	bear := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Dead Bear", TypeLine: "Creature — Bear",
			Owner: me.ID, Controller: me.ID})
	})

	perm := effectstest.GrantedPermissionOn(g, me.ID, sorcery, game.ZoneGraveyard)
	if !perm.Granted() || perm.AltCostKey != "flashback" || perm.Cost != "" || !perm.ExileOnResolution {
		t.Fatalf("the sorcery's granted flashback = %+v, want flashback for its mana cost, exiled after", perm)
	}
	if p := effectstest.GrantedPermissionOn(g, me.ID, bear, game.ZoneGraveyard); p.Granted() {
		t.Error("a creature card gained flashback")
	}

	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback the sorcery: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(sorcery) {
		t.Error("a sorcery cast with flashback is exiled as it leaves the stack")
	}

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(lier) })
	again := effectstest.SeedGraveyardCard(t, g, "Another Sorcery", "Sorcery", effectstest.CsTestGraveyardSorceryOracl)
	if p := effectstest.GrantedPermissionOn(g, me.ID, again, game.ZoneGraveyard); p.Granted() {
		t.Error("flashback outlived Lier")
	}
}
