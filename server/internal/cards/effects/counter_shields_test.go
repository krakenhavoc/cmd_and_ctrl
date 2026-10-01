package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counter_shields_test.go — ADR 0106 PR 2 (#1806): the printed
// "<these> spells can't be countered" statics, end to end through the
// real catalog and the engine's counter gate. The gate's own contract
// (every verb, every form, the source leaving) is pinned with stubs in
// game/cant_be_countered_test.go.

const (
	csGaeasHeraldOracle         = "f21c3d22-6149-475f-ae6e-37a25ef020e5"
	csLeylineOfLifeforceOracle  = "997478aa-b790-4269-a626-abf0cb30fea0"
	csProwlingSerpopardOracle   = "4bdfd718-f474-4223-88d8-7fa9fb0c86b4"
	csRootSliverOracle          = "cb752313-dc7f-47fc-9077-9e3298e8f5fb"
	csSphinxFinalWordOracle     = "d4246e4d-390d-4925-a5a8-89cd096a237c"
	csSpellbreakerOracle        = "cba07472-7212-4411-a9f9-38a48870ad69"
	csSurrakOracle              = "7c766365-82c2-46d6-8521-42e73129f4ef"
	csAllosaurusShepherdOracle  = "98295bdc-db58-4e1f-ad4c-63092aea207e"
	csHexingSquelcherOracle     = "96c2d71e-0fc9-42aa-bc6b-6d0ae3f66f8b"
	csDestinySpinnerOracle      = "7dd189f8-a23f-4b43-b0ef-fb174a9664ba"
	csThryxOracle               = "20004398-bd04-4583-8378-9df743246a32"
	csCunningNightbonderOracle  = "7351317d-62a1-4894-9811-aa950c49eece"
	csLierOracle                = "f224db3d-cdd2-43f2-b625-4b861efa6449"
	csTestGraveyardSorceryOracl = "test-1806-graveyard-sorcery"
)

// csSpell is a spell to put on the stack: its card, who cast it, who
// controls it now, and whether it is a copy.
type csSpell struct {
	card               game.Card
	caster, controller int // seat index; seat 0 controls the static
	copy               bool
}

func csCard(name, typeLine, cost string, power int, keywords ...string) game.Card {
	return game.Card{Name: name, TypeLine: typeLine, ManaCost: cost, Power: power, Keywords: keywords}
}

var (
	csBlueInstant  = csCard("Opt", "Instant", "{U}", 0)
	csGreenInstant = csCard("Giant Growth", "Instant", "{G}", 0)
	csSorcery      = csCard("Divination", "Sorcery", "{2}{U}", 0)
	csBear         = csCard("Grizzly Bears", "Creature — Bear", "{1}{G}", 2)
	csBigCreature  = csCard("Craw Wurm", "Creature — Wurm", "{4}{G}{G}", 6)
	csEnchantment  = csCard("Pacifism", "Enchantment — Aura", "{1}{W}", 0)
	csSliver       = csCard("Muscle Sliver", "Creature — Sliver", "{1}{G}", 1)
	csChangeling   = csCard("Changeling Outcast", "Creature — Shapeshifter", "{B}", 1, "changeling")
	csFiveDrop     = csCard("Mind Spring", "Sorcery", "{4}{U}", 0)
	csFourDrop     = csCard("Concentrate", "Sorcery", "{2}{U}{U}", 0)
	csFlashBeast   = csCard("Ambush Viper", "Creature — Snake", "{1}{G}", 2, "flash")
)

// csPushSpell puts a spell on the stack the way a cast leaves it: the
// card on g.Stack and its item in StackMeta.
func csPushSpell(g *game.Game, s csSpell) uuid.UUID {
	id := uuid.New()
	caster, controller := g.Seats[s.caster], g.Seats[s.controller]
	c := s.card
	c.InstanceID, c.Owner, c.Controller = id, caster.ID, controller.ID
	g.WithWriteLock(func() {
		g.Stack.PushTop(c)
		item := &game.StackItem{ID: id, Kind: game.StackItemSpell, Controller: controller.ID,
			Owner: caster.ID, SourceCardID: id, IsCopy: s.copy}
		if caster.ID != controller.ID {
			item.BaseController = caster.ID
		}
		if g.StackMeta == nil {
			g.StackMeta = map[uuid.UUID]*game.StackItem{}
		}
		g.StackMeta[id] = item
	})
	return id
}

// Every printed static of the seam, each against the spells it names
// and the nearest ones it does not. The verb is the one every catalog
// counterspell calls.
func TestPrintedCounterShieldsCoverWhatTheyPrint(t *testing.T) {
	mine := func(c game.Card) csSpell { return csSpell{card: c} }
	theirs := func(c game.Card) csSpell { return csSpell{card: c, caster: 1, controller: 1} }
	for _, tc := range []struct {
		name, typeLine, oracle string
		covered, exposed       []csSpell
	}{
		{"Chimil, the Inner Sun", "Legendary Artifact", chimilOracle,
			[]csSpell{mine(csBlueInstant), mine(csBear), {card: csSorcery, copy: true}, {card: csBlueInstant, caster: 1}},
			[]csSpell{theirs(csBlueInstant), {card: csBlueInstant, controller: 1}}},
		{"Gaea's Herald", "Creature — Elf", csGaeasHeraldOracle,
			[]csSpell{mine(csBear), theirs(csBear)},
			[]csSpell{mine(csBlueInstant), theirs(csEnchantment)}},
		{"Leyline of Lifeforce", "Enchantment", csLeylineOfLifeforceOracle,
			[]csSpell{mine(csBigCreature), theirs(csBear)},
			[]csSpell{theirs(csSorcery)}},
		{"Prowling Serpopard", "Creature — Cat Snake", csProwlingSerpopardOracle,
			[]csSpell{mine(csBear)},
			[]csSpell{theirs(csBear), mine(csBlueInstant)}},
		{"Root Sliver", "Creature — Sliver", csRootSliverOracle,
			[]csSpell{mine(csSliver), theirs(csSliver), theirs(csChangeling)},
			[]csSpell{mine(csBear)}},
		{"Sphinx of the Final Word", "Creature — Sphinx", csSphinxFinalWordOracle,
			[]csSpell{mine(csBlueInstant), mine(csSorcery)},
			[]csSpell{mine(csBear), theirs(csBlueInstant)}},
		{"Spellbreaker Behemoth", "Creature — Beast", csSpellbreakerOracle,
			[]csSpell{mine(csBigCreature)},
			[]csSpell{mine(csBear), theirs(csBigCreature), mine(csFiveDrop)}},
		{"Surrak Dragonclaw", "Legendary Creature — Human Warrior", csSurrakOracle,
			[]csSpell{mine(csBear)},
			[]csSpell{mine(csBlueInstant), theirs(csBear)}},
		{"Allosaurus Shepherd", "Creature — Elf Shaman", csAllosaurusShepherdOracle,
			[]csSpell{mine(csGreenInstant), mine(csBear)},
			[]csSpell{mine(csBlueInstant), theirs(csGreenInstant)}},
		{"Hexing Squelcher", "Creature — Goblin Sorcerer", csHexingSquelcherOracle,
			[]csSpell{mine(csBlueInstant)},
			[]csSpell{theirs(csBlueInstant)}},
		{"Destiny Spinner", "Enchantment Creature — Human", csDestinySpinnerOracle,
			[]csSpell{mine(csBear), mine(csEnchantment)},
			[]csSpell{mine(csBlueInstant), theirs(csBear)}},
		// "Spells you cast": the caster, through a change of control, and
		// never a copy (CR 707.10).
		{"Thryx, the Sudden Storm", "Legendary Creature — Elemental Giant", csThryxOracle,
			[]csSpell{mine(csFiveDrop), mine(csBigCreature), {card: csFiveDrop, controller: 1}},
			[]csSpell{mine(csFourDrop), {card: csFiveDrop, copy: true}, theirs(csFiveDrop), {card: csFiveDrop, caster: 1}}},
		{"Cunning Nightbonder", "Creature — Human Rogue", csCunningNightbonderOracle,
			[]csSpell{mine(csFlashBeast)},
			[]csSpell{mine(csBear), {card: csFlashBeast, copy: true}, theirs(csFlashBeast)}},
		{"Lier, Disciple of the Drowned", "Legendary Creature — Human Wizard", csLierOracle,
			[]csSpell{mine(csBlueInstant), theirs(csBlueInstant), theirs(csBear)},
			nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for want, spells := range map[bool][]csSpell{true: tc.covered, false: tc.exposed} {
				for _, s := range spells {
					g := newCatalogGame(t)
					pushCatalogPermanent(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, false)
					id := csPushSpell(g, s)
					if got := counterByEffect(t, g, id); got != want {
						t.Errorf("%s (cast by seat %d, controlled by seat %d, copy %v): survived the counter = %v, want %v",
							s.card.Name, s.caster, s.controller, s.copy, got, want)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Chimil, the Inner Sun", "Legendary Artifact", chimilOracle, false)

	bear := castFromHandForTest(t, g, me, "Grizzly Bears", "Creature — Bear", "{1}{G}", "", game.CastSpellParams{})
	if !g.SpellCantBeCounteredForEffect(bear) {
		t.Error("the stack chip should say the Bears can't be countered")
	}
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	counterspell := batch01OpponentCasts(t, g, opp, "Counterspell", b14CounterspellOracle, "{U}{U}",
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(bear) {
		t.Fatal("Chimil's controller's creature spell was countered")
	}
	if !opp.Graveyard.Contains(counterspell) {
		t.Error("the Counterspell resolves (and does nothing) — it should be in the graveyard, not fizzled away elsewhere")
	}
}

// CR 613.1f: a Chimil that has lost all its abilities shields nothing.
func TestChimilWithoutItsAbilitiesShieldsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	chimil := pushCatalogPermanent(g, me.ID, "Chimil, the Inner Sun", "Legendary Artifact", chimilOracle, false)
	g.WithWriteLock(func() {
		if !g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(chimil),
			[]game.Mod{game.LoseAllAbilitiesMod()}, game.IndefiniteDuration(), "test — loses all abilities") {
			t.Fatal("setup: the removal registered nothing")
		}
	})
	// Bring the layers up to date, as any priority pass would.
	g.ReadSnapshot(func() {})
	spell := csPushSpell(g, csSpell{card: csBlueInstant})
	if counterByEffect(t, g, spell) {
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
		{"Thryx, the Sudden Storm", "Legendary Creature — Elemental Giant", csThryxOracle, csFiveDrop, "UUUU"},
		{"Cunning Nightbonder", "Creature — Human Rogue", csCunningNightbonderOracle, csFlashBeast, "G"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			pushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
			floatForTest(g, me, tc.pool)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	shepherd := pushCatalogPermanent(g, me.ID, "Allosaurus Shepherd", "Creature — Elf Shaman", csAllosaurusShepherdOracle, false)
	myElf := pushCatalogPermanent(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", "", false)
	bear := pushCatalogPermanent(g, me.ID, "Grizzly Bears", "Creature — Bear", "", false)
	theirElf := pushCatalogPermanent(g, opp.ID, "Elvish Mystic", "Creature — Elf Druid", "", false)

	floatForTest(g, me, "GGGGGG")
	if err := g.ActivateCatalogAbility(me.ID, shepherd, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	late := pushCatalogPermanent(g, me.ID, "Fyndhorn Elves", "Creature — Elf Druid", "", false)

	for _, id := range []uuid.UUID{shepherd, myElf} {
		if p, tough := effectivePower(t, g, id), effectiveToughness(t, g, id); p != 5 || tough != 5 {
			t.Errorf("an Elf you control is %d/%d, want 5/5", p, tough)
		}
		if !hasAbility(effectiveSubtypes(t, g, id), "Dinosaur") {
			t.Errorf("an Elf you control is not a Dinosaur: %v", effectiveSubtypes(t, g, id))
		}
	}
	for _, id := range []uuid.UUID{bear, theirElf, late} {
		if p := effectivePower(t, g, id); p != 1 {
			t.Errorf("a creature outside the set was changed: power %d", p)
		}
	}
}

// Destiny Spinner's {3}{G}: the land is an X/X Elemental creature with
// trample and haste, X the enchantments its activator controls as it
// resolves, and still a land.
func TestDestinySpinnerAnimatesALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	spinner := pushCatalogPermanent(g, me.ID, "Destiny Spinner", "Enchantment Creature — Human", csDestinySpinnerOracle, false)
	pushCatalogPermanent(g, me.ID, "Honden of Seeing Winds", "Legendary Enchantment — Shrine", "", false)
	pushCatalogPermanent(g, me.ID, "Glorious Anthem", "Enchantment", "", false)
	forest := pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)

	floatForTest(g, me, "GGGG")
	if err := g.ActivateCatalogAbility(me.ID, spinner, 0, game.ActivateAbilityParams{Targets: cardRefs(forest)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	if p, tough := effectivePower(t, g, forest), effectiveToughness(t, g, forest); p != 3 || tough != 3 {
		t.Errorf("the Forest is %d/%d, want 3/3 for three enchantments", p, tough)
	}
	types := effectiveTypes(t, g, forest)
	if !hasAbility(types, "Creature") || !hasAbility(types, "Land") {
		t.Errorf("the Forest's types are %v, want a land creature", types)
	}
	if !hasAbility(effectiveSubtypes(t, g, forest), "Elemental") {
		t.Errorf("the Forest is not an Elemental: %v", effectiveSubtypes(t, g, forest))
	}
	abilities := effectiveAbilities(t, g, forest)
	if !hasAbility(abilities, "trample") || !hasAbility(abilities, "haste") {
		t.Errorf("the Forest's abilities are %v, want trample and haste", abilities)
	}
}

// Hexing Squelcher gives its controller's OTHER creatures ward — pay 2
// life, charged to the opponent who targets one.
func TestHexingSquelcherGrantsLifeWardToYourOtherCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Hexing Squelcher", "Creature — Goblin Sorcerer", csHexingSquelcherOracle, false)
	bear := pushCatalogPermanent(g, me.ID, "Grizzly Bears", "Creature — Bear", "", false)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	castAtWardedCreature(t, g, opp, bear)
	prompt := passUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the opponent who targeted the Bears")
	}
	if prompt.LifeCost != 2 {
		t.Errorf("the granted ward charges %d life, want 2", prompt.LifeCost)
	}
}

// Surrak's other creatures have trample; Surrak itself does not.
func TestSurrakGivesOtherCreaturesTrample(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	surrak := pushCatalogPermanent(g, me.ID, "Surrak Dragonclaw", "Legendary Creature — Human Warrior", csSurrakOracle, false)
	bear := pushCatalogPermanent(g, me.ID, "Grizzly Bears", "Creature — Bear", "", false)
	theirs := pushCatalogPermanent(g, opp.ID, "Hill Giant", "Creature — Giant", "", false)
	if !hasAbility(effectiveAbilities(t, g, bear), "trample") {
		t.Error("Surrak's controller's other creature has no trample")
	}
	if hasAbility(effectiveAbilities(t, g, surrak), "trample") || hasAbility(effectiveAbilities(t, g, theirs), "trample") {
		t.Error("trample reached Surrak itself or an opponent's creature")
	}
}

// Lier gives each instant and sorcery card in its controller's
// graveyard flashback for its mana cost, while Lier remains — and the
// card is exiled as it resolves (CR 702.34a).
func TestLierGivesYourGraveyardFlashback(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	lier := pushCatalogPermanent(g, me.ID, "Lier, Disciple of the Drowned", "Legendary Creature — Human Wizard", csLierOracle, false)
	sorcery := seedGraveyardCard(t, g, "Dead Sorcery", "Sorcery", csTestGraveyardSorceryOracl)
	bear := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Dead Bear", TypeLine: "Creature — Bear",
			Owner: me.ID, Controller: me.ID})
	})

	perm := grantedPermissionOn(g, me.ID, sorcery, game.ZoneGraveyard)
	if !perm.Granted() || perm.AltCostKey != "flashback" || perm.Cost != "" || !perm.ExileOnResolution {
		t.Fatalf("the sorcery's granted flashback = %+v, want flashback for its mana cost, exiled after", perm)
	}
	if p := grantedPermissionOn(g, me.ID, bear, game.ZoneGraveyard); p.Granted() {
		t.Error("a creature card gained flashback")
	}

	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback the sorcery: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(sorcery) {
		t.Error("a sorcery cast with flashback is exiled as it leaves the stack")
	}

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(lier) })
	again := seedGraveyardCard(t, g, "Another Sorcery", "Sorcery", csTestGraveyardSorceryOracl)
	if p := grantedPermissionOn(g, me.ID, again, game.ZoneGraveyard); p.Granted() {
		t.Error("flashback outlived Lier")
	}
}
