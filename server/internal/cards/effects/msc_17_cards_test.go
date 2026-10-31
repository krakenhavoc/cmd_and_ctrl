package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_17_cards_test.go — Marvel Super Heroes Commander, seventeenth batch
// (#4455). Each test plays the real catalog entry and asserts what its
// text says.

const (
	msc17DaughterOracle   = "63e0f03f-16da-4c09-ad03-8945f660942d"
	msc17DoormanOracle    = "01ea2130-e212-4532-b68b-2f940d53ef6e"
	msc17AnnieOracle      = "4c3ad545-b375-44a7-a24e-58baacc4e4b6"
	msc17SymbiosisOracle  = "19504e19-2886-4d28-a813-b798a9a446a2"
	msc17DragonManOracle  = "c03ca2e6-fcce-4018-952f-1c451ab92cfd"
	msc17DoraOracle       = "406f1982-5d0c-4dba-9010-28b1c6a3602c"
	msc17ClosetOracle     = "072d05c4-e0d7-47af-b2a9-16b90810fd30"
	msc17FirstFamOracle   = "e119c709-bf9d-451e-827b-9787d455029a"
	msc17GenesisOracle    = "06f42f49-32f0-413b-8437-b8efa4d647c2"
	msc17DoomfallOracle   = "ba2c2a65-256a-46db-bf2b-b02b7d114a42"
	msc17HatutOracle      = "92ac307e-6531-4465-8870-07988458f5d4"
	msc17BatrocOracle     = "ebbea7a7-c990-43fa-8f67-a5da17fbbb10"
	msc17CaptainMarvelOra = "6d86d783-7116-4426-8c0f-718c2d904887"
	msc17PigeonOracle     = "ac8cae63-270c-4f73-b29b-50f8f2395fd9"
)

func TestMsc17EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		msc17DaughterOracle:   "Daughter of the Deep",
		msc17DoormanOracle:    "Doorman",
		msc17AnnieOracle:      "Annie Joins Up",
		msc17SymbiosisOracle:  "Alien Symbiosis",
		msc17DragonManOracle:  "Dragon Man, Reformed Robot",
		msc17DoraOracle:       "Dora Milaje Elite",
		msc17ClosetOracle:     "Costume Closet",
		msc17FirstFamOracle:   "First Family",
		msc17GenesisOracle:    "Genesis Ultimatum",
		msc17DoomfallOracle:   "Doomfall",
		msc17HatutOracle:      "Hatut Zeraze Strike Force",
		msc17BatrocOracle:     "Batroc the Leaper",
		msc17CaptainMarvelOra: "Captain Marvel, Apex Avenger",
	}
	for oracle, name := range want {
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

// msc17Cast puts a card in the player's hand (they must be the active
// seat) and casts it from a main phase, returning its id and the cast
// error. The spell is left on the stack.
func msc17Cast(t *testing.T, g *game.Game, p *game.Player, c game.Card, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	advanceToMain(t, g)
	p.Hand.PushTop(c)
	return c.InstanceID, g.CastSpell(p.ID, c.InstanceID, params)
}

// msc17Play casts a card and resolves everything it puts on the stack.
func msc17Play(t *testing.T, g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	t.Helper()
	id, err := msc17Cast(t, g, p, c, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast %s: %v", c.Name, err)
	}
	passPriorityAroundTable(t, g)
	return id
}

// msc17UntilATriggerIsOnTheStack passes priority until a triggered ability
// sits on the stack, leaving it unresolved.
func msc17UntilATriggerIsOnTheStack(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 16; i++ {
		for _, it := range g.StackMeta {
			if it != nil && it.Kind == game.StackItemTriggered {
				return
			}
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("no triggered ability reached the stack")
}

func msc17NamedOnBattlefield(g *game.Game, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			out = append(out, c)
		}
	}
	return out
}

// --- Daughter of the Deep -------------------------------------------------

func TestMsc17DaughterMakesAMerfolkOnlyOnTheSecondDrawEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, other := g.Seats[1], g.Seats[2]
	pushCatalogPermanent(g, me.ID, "Daughter of the Deep", "Creature — Merfolk Noble", msc17DaughterOracle, false)
	hobA5Draw(t, g, other, nil)
	hobA5Draw(t, g, other, nil)
	if n := countNamed(g, "Merfolk"); n != 0 {
		t.Fatalf("%d Merfolk after another player drew twice, want 0", n)
	}
	hobA5Draw(t, g, me, nil)
	if n := countNamed(g, "Merfolk"); n != 0 {
		t.Fatalf("%d Merfolk after the first draw, want 0", n)
	}
	hobA5Draw(t, g, me, nil)
	if n := countNamed(g, "Merfolk"); n != 1 {
		t.Fatalf("%d Merfolk after the second draw, want 1", n)
	}
	hobA5Draw(t, g, me, nil)
	if n := countNamed(g, "Merfolk"); n != 1 {
		t.Errorf("%d Merfolk after the third draw, want it to stay 1", n)
	}
}

func TestMsc17DaughterMakesAMerfolkUnblockableForAPrice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	daughter := pushCatalogPermanent(g, me.ID, "Daughter of the Deep", "Creature — Merfolk Noble", msc17DaughterOracle, false)
	merfolk := b12Creature(g, opp.ID, "Their Merfolk", "Creature — Merfolk", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, daughter, 0, game.ActivateAbilityParams{Targets: cardRefs(bear), Strict: true}); err == nil {
		t.Fatal("the ability targets a Merfolk, not any creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, daughter, 0, game.ActivateAbilityParams{Targets: cardRefs(merfolk), Strict: true}); err == nil {
		t.Fatal("the ability was activated without {U}")
	}
	floatMana(t, g, me, "{U}")
	if err := g.ActivateCatalogAbility(me.ID, daughter, 0, game.ActivateAbilityParams{Targets: cardRefs(merfolk), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	assertRestrictions(t, g, merfolk, game.CantBeBlocked)
	assertRestrictions(t, g, bear, 0)
	if c, _ := battlefieldCard(g, daughter); !c.Tapped {
		t.Error("{T} is part of the cost")
	}
}

// --- Doorman ----------------------------------------------------------------

func TestMsc17DoormanBarsSmallBlockersAndWalls(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	doorman := pushCatalogPermanent(g, me.ID, "Doorman", "Legendary Creature — Mutant Hero", msc17DoormanOracle, false)
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Ogre", 3, 3)
	bear := b12Creature(g, opp.ID, "Small Bear", "Creature — Bear", 2, 2)
	wall := b12Creature(g, opp.ID, "Big Wall", "Creature — Wall", 5, 5)
	ogre := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, doorman, 0, game.ActivateAbilityParams{Targets: cardRefs(attacker)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	declareAttack(t, g, opp.ID, attacker)
	advanceTo(t, g, game.StepDeclareBlockers)
	brRefusal(t, g.DeclareBlocker(bear, attacker), game.BlockReasonCantBeBlockedBy)
	brRefusal(t, g.DeclareBlocker(wall, attacker), game.BlockReasonCantBeBlockedBy)
	if err := g.DeclareBlocker(ogre, attacker); err != nil {
		t.Errorf("a non-Wall with power 3 blocks it: %v", err)
	}
}

func TestMsc17DoormanDoesNotBarBlockersFromACreatureItDidNotTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	doorman := pushCatalogPermanent(g, me.ID, "Doorman", "Legendary Creature — Mutant Hero", msc17DoormanOracle, false)
	chosen := b12Creature(g, me.ID, "Chosen", "Creature — Ogre", 3, 3)
	other := b12Creature(g, me.ID, "Other", "Creature — Ogre", 3, 3)
	bear := b12Creature(g, opp.ID, "Small Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, doorman, 0, game.ActivateAbilityParams{Targets: cardRefs(chosen)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	declareAttack(t, g, opp.ID, other)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(bear, other); err != nil {
		t.Errorf("the bar belongs to the target alone: %v", err)
	}
}

// --- Annie Joins Up -----------------------------------------------------------

func TestMsc17AnnieDealsFiveToACreatureOrPlaneswalkerAnOpponentControls(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
	mine := b12Creature(g, me.ID, "My Giant", "Creature — Giant", 5, 5)
	msc17Cast(t, g, me, game.Card{Name: "Annie Joins Up", TypeLine: "Legendary Enchantment", OracleID: msc17AnnieOracle}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	pickIfAsked(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("five damage did not kill the opponent's 5/5")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the caster's own creature is not a legal target")
	}
}

func TestMsc17AnnieAsksForNothingWhenOnlyYourCreaturesExist(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := b12Creature(g, me.ID, "My Giant", "Creature — Giant", 5, 5)
	msc17Cast(t, g, me, game.Card{Name: "Annie Joins Up", TypeLine: "Legendary Enchantment", OracleID: msc17AnnieOracle}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Fatalf("a target was asked for with none legal: %+v", p)
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the caster's own creature died")
	}
}

func TestMsc17AnnieTriggersALegendaryCreaturesAbilityAnAdditionalTime(t *testing.T) {
	foods := func(legendary, withAnnie bool, controller int) int {
		g := newCatalogGame(t)
		me := g.Seats[0]
		owner := g.Seats[controller]
		if withAnnie {
			pushCatalogPermanent(g, me.ID, "Annie Joins Up", "Legendary Enchantment", msc17AnnieOracle, false)
		}
		typeLine := "Creature — Bird"
		if legendary {
			typeLine = "Legendary Creature — Bird"
		}
		pigeon := pushCatalogPermanent(g, owner.ID, "City Pigeon", typeLine, msc17PigeonOracle, false)
		b27Kill(g, pigeon)
		passPriorityAroundTable(t, g)
		return countNamed(g, "Food")
	}
	if got := foods(true, false, 0); got != 1 {
		t.Fatalf("baseline: %d Food, want 1", got)
	}
	if got := foods(true, true, 0); got != 2 {
		t.Errorf("a legendary creature's trigger with Annie made %d Food, want 2", got)
	}
	if got := foods(false, true, 0); got != 1 {
		t.Errorf("a nonlegendary creature's trigger with Annie made %d Food, want 1", got)
	}
	if got := foods(true, true, 1); got != 1 {
		t.Errorf("an opponent's legendary creature's trigger made %d Food, want 1", got)
	}
}

// --- Alien Symbiosis -----------------------------------------------------------

func TestMsc17AlienSymbiosisPumpsGrantsMenaceAndAddsAType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	_, err := msc17Cast(t, g, me, game.Card{Name: "Alien Symbiosis", TypeLine: "Enchantment — Aura", OracleID: msc17SymbiosisOracle, ManaCost: "{1}{B}"},
		game.CastSpellParams{Targets: cardRefs(bear)})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, bear); got != 3 {
		t.Errorf("power %d, want 3", got)
	}
	assertKeywords(t, g, bear, "menace")
	if subs := ringBFCard(t, g, bear).Effective().Subtypes; !containsString(subs, "Symbiote") || !containsString(subs, "Bear") {
		t.Errorf("subtypes %v, want Bear and Symbiote", subs)
	}
}

func TestMsc17AlienSymbiosisIsCastFromTheGraveyardByDiscardingACard(t *testing.T) {
	g, me, _, spare := retraceTable(t)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !game.CardCastableFromZone(msc17SymbiosisOracle, game.ZoneGraveyard) {
		t.Fatal("the graveyard is not open to Alien Symbiosis")
	}
	offers := game.AlternativeCostsOfferedFromZone(msc17SymbiosisOracle, game.ZoneGraveyard)
	if len(offers) != 1 || offers[0].ManaCost != "{1}{B}" || offers[0].CardPaymentCount() != 1 {
		t.Fatalf("graveyard offers = %+v, want the printed {1}{B} plus one discarded card", offers)
	}
	aura := graveyardCardOf(g, me, "Alien Symbiosis", "Enchantment — Aura", "{1}{B}", msc17SymbiosisOracle)
	params := game.CastSpellParams{FromZone: "graveyard", AlternativeCost: offers[0].Key, Strict: true, Targets: cardRefs(bear)}

	// Without the discard, or without the mana, nothing is paid and nothing is cast.
	payMana(t, g, me, "{B}{B}")
	if err := g.CastSpell(me.ID, aura, params); err == nil {
		t.Fatal("cast from the graveyard with no card discarded")
	}
	if !inGraveyard(me, aura) {
		t.Fatal("the failed cast moved the card")
	}
	g.WithWriteLock(func() { me.ManaPool = nil })
	params.AltCostIDs = []uuid.UUID{spare}
	if err := g.CastSpell(me.ID, aura, params); err == nil {
		t.Fatal("cast from the graveyard without paying its mana cost")
	}
	if !me.Hand.Contains(spare) {
		t.Fatal("the failed cast discarded the card")
	}

	payMana(t, g, me, "{B}{B}")
	if err := g.CastSpell(me.ID, aura, params); err != nil {
		t.Fatalf("cast from the graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aura) {
		t.Error("the Aura did not enter")
	}
	if !inGraveyard(me, spare) {
		t.Error("the discarded card is not in the graveyard")
	}
	if got := effectivePower(t, g, bear); got != 3 {
		t.Errorf("the recast Aura gave power %d, want 3", got)
	}
}

// --- Dragon Man, Reformed Robot -------------------------------------------

// msc17DragonMan builds a table with Dragon Man, `ours` noncreature permanents
// of the given mana costs under his controller, `theirs` under an opponent,
// and the given cards in his controller's graveyard (name, type line, cost).
func msc17DragonMan(t *testing.T, ours, theirs []string, graveyard [][3]string) (*game.Game, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for _, cost := range ours {
		pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", ManaCost: cost, Owner: me.ID, Controller: me.ID})
	}
	for _, cost := range theirs {
		pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Relic", TypeLine: "Artifact", ManaCost: cost, Owner: opp.ID, Controller: opp.ID})
	}
	for _, c := range graveyard {
		graveyardCardOf(g, me, c[0], c[1], c[2], "")
	}
	dragon := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dragon Man, Reformed Robot", TypeLine: "Legendary Artifact Creature — Dragon Robot",
		OracleID: msc17DragonManOracle, ManaCost: "{2}{W}{U}", Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	return g, dragon
}

func TestMsc17DragonManPowerIsTheGreatestNoncreatureManaValue(t *testing.T) {
	g, dragon := msc17DragonMan(t, nil, nil, nil)
	if got := effectivePower(t, g, dragon); got != 0 {
		t.Errorf("with nothing else, power %d, want 0 (he is a creature and counts nothing)", got)
	}
	if c := ringBFCard(t, g, dragon); c.CurrentToughness() != 5 {
		t.Errorf("toughness %d, want 5", c.CurrentToughness())
	}

	// An artifact of mana value 3; an opponent's 9 and a creature's 8 do not count.
	g, dragon = msc17DragonMan(t, []string{"{3}"}, []string{"{9}"}, nil)
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Big Creature", TypeLine: "Creature — Ogre", ManaCost: "{8}", Power: 8, Toughness: 8, Owner: g.Seats[0].ID, Controller: g.Seats[0].ID})
	if got := effectivePower(t, g, dragon); got != 3 {
		t.Errorf("power %d, want 3", got)
	}

	// A noncreature card in the graveyard counts, a creature card does not.
	g, dragon = msc17DragonMan(t, []string{"{3}"}, nil, [][3]string{{"Big Spell", "Sorcery", "{5}"}, {"Dead Ogre", "Creature — Ogre", "{7}"}})
	if got := effectivePower(t, g, dragon); got != 5 {
		t.Errorf("power %d, want 5 (the sorcery in the graveyard)", got)
	}
}

func TestMsc17DragonManIsCastFromTheGraveyardByDiscardingACard(t *testing.T) {
	g, me, _, spare := retraceTable(t)
	dragon := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dragon, Name: "Dragon Man, Reformed Robot", TypeLine: "Legendary Artifact Creature — Dragon Robot",
		ManaCost: "{2}{W}{U}", OracleID: msc17DragonManOracle, Toughness: 5, Owner: me.ID, Controller: me.ID,
		KnownBy: map[uuid.UUID]bool{me.ID: true}})
	offers := game.AlternativeCostsOfferedFromZone(msc17DragonManOracle, game.ZoneGraveyard)
	if len(offers) != 1 || offers[0].ManaCost != "{2}{W}{U}" {
		t.Fatalf("graveyard offers = %+v", offers)
	}
	params := game.CastSpellParams{FromZone: "graveyard", AlternativeCost: offers[0].Key, Strict: true}
	payMana(t, g, me, "{W}{U}{R}{R}")
	if err := g.CastSpell(me.ID, dragon, params); err == nil {
		t.Fatal("cast from the graveyard with no card discarded")
	}
	params.AltCostIDs = []uuid.UUID{spare}
	if err := g.CastSpell(me.ID, dragon, params); err != nil {
		t.Fatalf("cast from the graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dragon) {
		t.Error("Dragon Man did not enter")
	}
	if !inGraveyard(me, spare) {
		t.Error("the discarded card is not in the graveyard")
	}
}

// --- Dora Milaje Elite -------------------------------------------------------

func msc17DoraCard() game.Card {
	return game.Card{Name: "Dora Milaje Elite", TypeLine: "Creature — Human Warrior", OracleID: msc17DoraOracle, ManaCost: "{1}{W}", Power: 2, Toughness: 2}
}

func TestMsc17DoraMilajeEliteMakesATappedVibraniumOnlyWhenAnOpponentHasMoreLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	pushCatalogPermanent(g, opp.ID, "Forest", "Basic Land — Forest", "", false)
	pushCatalogPermanent(g, opp.ID, "Forest", "Basic Land — Forest", "", false)
	msc17Play(t, g, me, msc17DoraCard())
	vib := msc17NamedOnBattlefield(g, "Vibranium")
	if len(vib) != 1 || !vib[0].Tapped || vib[0].Controller != me.ID {
		t.Fatalf("Vibranium tokens %+v, want one tapped token under the caster", vib)
	}
	if !vib[0].IsArtifact() {
		t.Error("Vibranium is an artifact")
	}

	even := newCatalogGame(t)
	me2, opp2 := even.Seats[0], even.Seats[1]
	pushCatalogPermanent(even, me2.ID, "Forest", "Basic Land — Forest", "", false)
	pushCatalogPermanent(even, opp2.ID, "Forest", "Basic Land — Forest", "", false)
	msc17Play(t, even, me2, msc17DoraCard())
	if n := len(msc17NamedOnBattlefield(even, "Vibranium")); n != 0 {
		t.Errorf("%d Vibranium with equal land counts, want 0", n)
	}
}

func TestMsc17DoraMilajeEliteRechecksTheLandCountOnResolution(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, opp.ID, "Forest", "Basic Land — Forest", "", false)
	if _, err := msc17Cast(t, g, me, msc17DoraCard(), game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	msc17UntilATriggerIsOnTheStack(t, g)
	pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	passPriorityAroundTable(t, g)
	if n := len(msc17NamedOnBattlefield(g, "Vibranium")); n != 0 {
		t.Errorf("%d Vibranium after the lands became equal, want 0 (CR 603.4)", n)
	}
}

func TestMsc17VibraniumIsIndestructibleAndItsManaCannotCastNonartifactSpells(t *testing.T) {
	tok := VibraniumToken()
	if !containsString(tok.Keywords, "indestructible") || tok.TypeLine != "Token Artifact — Vibranium" {
		t.Errorf("token %q keywords %v", tok.TypeLine, tok.Keywords)
	}
	abilities := game.ManaAbilitiesForCard(tok)
	if len(abilities) != 1 || abilities[0].Produced != "{C}" || !abilities[0].TapCost ||
		len(abilities[0].Restrictions) != 1 || abilities[0].Restrictions[0] != game.ManaRestrictNotNonartifactSpell {
		t.Errorf("mana abilities %+v, want {T}: restricted {C}", abilities)
	}
}

func TestMsc17DoraMilajeEliteSacrificeGivesLegendaryPermanentsIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elite := b12Push(g, me.ID, "Dora Milaje Elite", "Creature — Human Warrior", msc17DoraOracle, 2, 2)
	legend := b12Creature(g, me.ID, "Legend", "Legendary Creature — Human", 2, 2)
	relic := b12Permanent(g, me.ID, "Legend Relic", "Legendary Artifact")
	plain := b12Creature(g, me.ID, "Plain", "Creature — Human", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Legend", "Legendary Creature — Human", 2, 2)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, elite, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(elite) {
		t.Error("the Elite was not sacrificed")
	}
	for _, id := range []uuid.UUID{legend, relic} {
		if !hasKeywordOnBattlefield(t, g, id, "indestructible") {
			t.Errorf("legendary permanent %s is not indestructible", id)
		}
	}
	for _, id := range []uuid.UUID{plain, theirs} {
		if hasKeywordOnBattlefield(t, g, id, "indestructible") {
			t.Errorf("permanent %s gained indestructible but is not a legendary permanent you control", id)
		}
	}
}

// --- Costume Closet ----------------------------------------------------------

func msc17ClosetCard() game.Card {
	return game.Card{Name: "Costume Closet", TypeLine: "Artifact", OracleID: msc17ClosetOracle, ManaCost: "{1}{W}"}
}

func TestMsc17CostumeClosetEntersWithTwoCountersAndMovesThemAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	closet := msc17Play(t, g, me, msc17ClosetCard())
	if n := b12Counter(t, g, closet, game.CounterPlusOne); n != 2 {
		t.Fatalf("enters with %d counters, want 2", n)
	}
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, closet, 0, game.ActivateAbilityParams{Targets: cardRefs(theirs)}); err == nil {
		t.Fatal("moved a counter onto an opponent's creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, closet, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := b12Counter(t, g, closet, game.CounterPlusOne); n != 1 {
		t.Errorf("the Closet has %d counters, want 1", n)
	}
	if n := b12Counter(t, g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("the creature has %d counters, want 1", n)
	}
}

func TestMsc17CostumeClosetCannotActivateOutsideAMainPhase(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	closet := msc17Play(t, g, me, msc17ClosetCard())
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, closet, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}); err == nil {
		t.Error("activated in the beginning of combat step")
	}
}

func TestMsc17CostumeClosetGrowsWhenAModifiedCreatureYouControlLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	closet := msc17Play(t, g, me, msc17ClosetCard())
	countersOnCloset := func() int { return b12Counter(t, g, closet, game.CounterPlusOne) }
	equipped := b12Creature(g, me.ID, "Equipped", "Creature — Bear", 2, 2)
	daggers := seedEquipment(g, me.ID, "Vibranium Energy Daggers", mshB10DaggersOracle)
	equipTo(t, g, me.ID, daggers, equipped)

	plain := b12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	b27Kill(g, plain)
	passPriorityAroundTable(t, g)
	if n := countersOnCloset(); n != 2 {
		t.Fatalf("an unmodified creature leaving changed the Closet to %d counters", n)
	}

	countered := b12Creature(g, me.ID, "Countered", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterByForEffect(me.ID, countered, game.CounterPlusOne, 1) })
	b27Kill(g, countered)
	passPriorityAroundTable(t, g)
	if n := countersOnCloset(); n != 3 {
		t.Fatalf("a creature with a counter leaving: Closet has %d counters, want 3", n)
	}

	theirs := b12Creature(g, opp.ID, "Their Countered", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterByForEffect(opp.ID, theirs, game.CounterPlusOne, 1) })
	b27Kill(g, theirs)
	passPriorityAroundTable(t, g)
	if n := countersOnCloset(); n != 3 {
		t.Fatalf("an opponent's modified creature leaving changed the Closet to %d counters", n)
	}

	b27Kill(g, equipped)
	passPriorityAroundTable(t, g)
	if n := countersOnCloset(); n != 4 {
		t.Errorf("an equipped creature leaving: Closet has %d counters, want 4", n)
	}
}

// --- First Family ----------------------------------------------------------------

func msc17FirstFamilyCard() game.Card {
	return game.Card{Name: "First Family", TypeLine: "Instant", OracleID: msc17FirstFamOracle, ManaCost: "{2}{G}{U}", Colors: []string{"U", "G"}}
}

func TestMsc17FirstFamilyCountsTheColorsOfPermanentsAndSpellsCast(t *testing.T) {
	// Alone: the spell is green and blue.
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand, life := me.Hand.Size(), me.Life
	_, err := msc17Cast(t, g, me, msc17FirstFamilyCard(), game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("hand grew by %d, want 2 cards drawn", got)
	}
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d life, want 2", got)
	}

	// With a red and a white permanent and a black spell cast earlier: five colors.
	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	pushBattlefieldCardWithTimestamp(g2, game.Card{InstanceID: uuid.New(), Name: "Red Ogre", TypeLine: "Creature — Ogre", Power: 2, Toughness: 2, Colors: []string{"R"}, Owner: me2.ID, Controller: me2.ID})
	pushBattlefieldCardWithTimestamp(g2, game.Card{InstanceID: uuid.New(), Name: "White Knight", TypeLine: "Creature — Knight", Power: 2, Toughness: 2, Colors: []string{"W"}, Owner: me2.ID, Controller: me2.ID})
	pushBattlefieldCardWithTimestamp(g2, game.Card{InstanceID: uuid.New(), Name: "Their Black", TypeLine: "Creature — Zombie", Power: 2, Toughness: 2, Colors: []string{"B"}, Owner: opp2.ID, Controller: opp2.ID})
	if _, err := msc17Cast(t, g2, me2, game.Card{Name: "Earlier Black Spell", TypeLine: "Sorcery", Colors: []string{"B"}}, game.CastSpellParams{}); err != nil {
		t.Fatalf("earlier spell: %v", err)
	}
	passPriorityAroundTable(t, g2)
	hand2, life2 := me2.Hand.Size(), me2.Life
	if _, err := msc17Cast(t, g2, me2, msc17FirstFamilyCard(), game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g2)
	// R and W from permanents; B, G and U from spells (an opponent's black permanent adds nothing new to the set).
	if got := me2.Life - life2; got != 5 {
		t.Errorf("gained %d life, want 5 (R, W, B, G, U)", got)
	}
	if got := me2.Hand.Size() - hand2; got != 5 {
		t.Errorf("hand grew by %d, want 5 cards drawn", got)
	}
}

func TestMsc17FirstFamilyIgnoresAnOpponentsPermanentsAndSpells(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Red", TypeLine: "Creature — Ogre", Power: 2, Toughness: 2, Colors: []string{"R"}, Owner: opp.ID, Controller: opp.ID})
	life := me.Life
	if _, err := msc17Cast(t, g, me, msc17FirstFamilyCard(), game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d life, want 2: only your permanents count", got)
	}
}

// --- Genesis Ultimatum ------------------------------------------------------------

func TestMsc17GenesisUltimatumPutsPermanentsOntoTheBattlefieldTheRestInHandAndExilesItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	below := plTop(me, "Below", "Creature — Bear", "{1}{G}")
	land := plTop(me, "Forest", "Basic Land — Forest", "")
	sorcery := plTop(me, "Rite", "Sorcery", "{1}")
	artifact := plTop(me, "Rock", "Artifact", "{2}")
	instant := plTop(me, "Zap", "Instant", "{1}")
	bear := plTop(me, "Bear", "Creature — Bear", "{1}{G}")
	spell := b12PlayFromHand(t, g, "Genesis Ultimatum", "Sorcery", msc17GenesisOracle, game.CastSpellParams{})
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick offered")
	}
	if hasID(pick.ChooseCards, sorcery) || hasID(pick.ChooseCards, instant) || hasID(pick.ChooseCards, below) {
		t.Errorf("offered %v: only permanent cards among the top five may be put onto the battlefield", pick.ChooseCards)
	}
	if !hasID(pick.ChooseCards, land) || !hasID(pick.ChooseCards, artifact) || !hasID(pick.ChooseCards, bear) {
		t.Errorf("offered %v, want the land, the artifact and the creature", pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, bear, artifact)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) || !g.Battlefield.Contains(artifact) {
		t.Error("the chosen permanents did not enter")
	}
	for _, id := range []uuid.UUID{land, sorcery, instant} {
		if !me.Hand.Contains(id) {
			t.Errorf("card %s did not go to the hand", id)
		}
	}
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("hand grew by %d, want 3", got)
	}
	if !me.Library.Contains(below) {
		t.Error("the sixth card left the library")
	}
	if !g.Exile.Contains(spell) || inGraveyard(me, spell) {
		t.Error("Genesis Ultimatum is not exiled")
	}
}

func TestMsc17GenesisUltimatumChoosingNothingPutsAllFiveInHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	var top []uuid.UUID
	for _, n := range []string{"A", "B", "C", "D", "E"} {
		top = append(top, plTop(me, n, "Creature — Bear", "{1}"))
	}
	spell := b12PlayFromHand(t, g, "Genesis Ultimatum", "Sorcery", msc17GenesisOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	for _, id := range top {
		if !me.Hand.Contains(id) {
			t.Errorf("card %s is not in the hand", id)
		}
	}
	if !g.Exile.Contains(spell) {
		t.Error("Genesis Ultimatum is not exiled")
	}
}

// --- Doomfall -----------------------------------------------------------------------

func msc17DoomfallCard() game.Card {
	return game.Card{Name: "Doomfall", TypeLine: "Sorcery", OracleID: msc17DoomfallOracle, ManaCost: "{2}{B}"}
}

func msc17TargetPlayer(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetPlayer, ID: id, Mode: 0}
}

func TestMsc17DoomfallLetsTheOpponentChooseWhichCreatureToExile(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	strong := b12Creature(g, opp.ID, "Strong", "Creature — Ogre", 5, 5)
	weak := b12Creature(g, opp.ID, "Weak", "Creature — Bear", 1, 1)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	if _, err := msc17Cast(t, g, me, msc17DoomfallCard(), game.CastSpellParams{Modes: []int{0}, Targets: []game.TargetRef{msc17TargetPlayer(me.ID)}}); err == nil {
		t.Fatal("Doomfall targeted its own caster")
	}
	if _, err := msc17Cast(t, g, me, msc17DoomfallCard(), game.CastSpellParams{Modes: []int{0}, Targets: []game.TargetRef{msc17TargetPlayer(opp.ID)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("the caster was asked to choose: the opponent chooses")
	}
	pick := chooseCardsChoiceFor(g, opp.ID)
	if pick == nil || !hasID(pick.ChooseCards, strong) || !hasID(pick.ChooseCards, weak) || hasID(pick.ChooseCards, mine) {
		t.Fatalf("the opponent's pick %+v should offer exactly their creatures", pick)
	}
	answerChooseCards(t, g, opp.ID, weak)
	if !g.Exile.Contains(weak) {
		t.Error("the chosen creature is not exiled")
	}
	if !g.Battlefield.Contains(strong) || !g.Battlefield.Contains(mine) {
		t.Error("an unchosen creature left the battlefield")
	}
}

func TestMsc17DoomfallExileEdictAsksNothingOfAnOpponentWithNoCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	if _, err := msc17Cast(t, g, me, msc17DoomfallCard(), game.CastSpellParams{Modes: []int{0}, Targets: []game.TargetRef{msc17TargetPlayer(opp.ID)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, opp.ID) != nil || chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a choice was asked with no creature to exile")
	}
}

func TestMsc17DoomfallExilesANonlandCardChosenFromTheRevealedHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ids := revealHand(opp, rhBolt(), game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	if _, err := msc17Cast(t, g, me, msc17DoomfallCard(), game.CastSpellParams{Modes: []int{1}, Targets: []game.TargetRef{msc17TargetPlayer(opp.ID)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := openVariant(t, g)
	if hasID(pick.DiscardOptions, ids[1]) || !hasID(pick.DiscardOptions, ids[0]) {
		t.Fatalf("options %v: only the nonland card may be chosen", pick.DiscardOptions)
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(ids[0]) || opp.Graveyard.Contains(ids[0]) {
		t.Error("the chosen card is not exiled")
	}
	if !opp.Hand.Contains(ids[1]) {
		t.Error("the land left their hand")
	}
}

// --- Hatut Zeraze Strike Force ------------------------------------------------------

func msc17HatutCard() game.Card {
	return game.Card{Name: "Hatut Zeraze Strike Force", TypeLine: "Creature — Human Spy Warrior", OracleID: msc17HatutOracle, ManaCost: "{3}{W}", Power: 2, Toughness: 2}
}

func TestMsc17HatutCopiesItselfForEachCommanderCastFromTheCommandZone(t *testing.T) {
	for _, tc := range []struct {
		casts, want int
	}{{0, 1}, {1, 2}, {2, 3}} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		g.WithWriteLock(func() { me.CommanderCasts = map[uuid.UUID]int{uuid.New(): tc.casts} })
		msc17Play(t, g, me, msc17HatutCard())
		if got := len(msc17NamedOnBattlefield(g, "Hatut Zeraze Strike Force")); got != tc.want {
			t.Errorf("after %d commander casts: %d Hatut on the battlefield, want %d", tc.casts, got, tc.want)
		}
	}
}

func TestMsc17HatutIgnoresAnOpponentsCommanderCasts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { opp.CommanderCasts = map[uuid.UUID]int{uuid.New(): 3} })
	msc17Play(t, g, me, msc17HatutCard())
	if got := len(msc17NamedOnBattlefield(g, "Hatut Zeraze Strike Force")); got != 1 {
		t.Errorf("%d Hatut, want 1", got)
	}
}

func TestMsc17HatutDestroysAnArtifactOrEnchantmentWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	msc17Cast(t, g, me, msc17HatutCard(), game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	pickIfAsked(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("a creature was destroyed")
	}
}

// --- Batroc the Leaper ----------------------------------------------------------------

func msc17BatrocCard() game.Card {
	return game.Card{Name: "Batroc the Leaper", TypeLine: "Legendary Creature — Human Villain", OracleID: msc17BatrocOracle, ManaCost: "{1}{R}", Power: 2, Toughness: 2}
}

func TestMsc17BatrocGrowsAndDamagesUpToXTargetsForEachKick(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Ogre", 4, 4)
	survivor := b12Creature(g, opp.ID, "Survivor", "Creature — Ogre", 5, 5)
	life := opp.Life
	batroc, err := msc17Cast(t, g, me, msc17BatrocCard(), game.CastSpellParams{OptionalCosts: []int{0, 0}})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := b12Counter(t, g, batroc, game.CounterPlusOne); n != 2 {
		t.Fatalf("%d counters, want one per kick (2)", n)
	}
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatalf("no target prompt: %+v", g.PendingChoices)
	}
	three := []game.TargetRef{
		{Kind: game.TargetCard, ID: victim}, {Kind: game.TargetCard, ID: survivor}, {Kind: game.TargetPlayer, ID: opp.ID},
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, three); err == nil {
		t.Fatal("'up to X' with X=2 accepted three targets")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{three[0], three[2]}); err != nil {
		t.Fatalf("pick two: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("4 damage did not kill the 4/4")
	}
	if got := life - opp.Life; got != 4 {
		t.Errorf("the opponent lost %d life, want 4 (his power with two counters)", got)
	}
	if !g.Battlefield.Contains(survivor) {
		t.Error("an unchosen creature died")
	}
}

func TestMsc17BatrocUnkickedAsksForNoTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bystander := b12Creature(g, opp.ID, "Bystander", "Creature — Bear", 1, 1)
	batroc, err := msc17Cast(t, g, me, msc17BatrocCard(), game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Fatalf("an unkicked Batroc asked for targets: %+v", p)
	}
	if n := b12Counter(t, g, batroc, game.CounterPlusOne); n != 0 {
		t.Errorf("%d counters, want 0", n)
	}
	if !g.Battlefield.Contains(bystander) {
		t.Error("an unkicked Batroc dealt damage")
	}
}

func TestMsc17BatrocDealsHisLastKnownPowerIfHeLeavesBeforeTheTriggerResolves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Ogre", 3, 3)
	batroc, err := msc17Cast(t, g, me, msc17BatrocCard(), game.CastSpellParams{OptionalCosts: []int{0}})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	resolveTopOfStack(t, g) // the creature spell; the trigger now needs its target
	pickIfAsked(t, g, me.ID, victim)
	b27Kill(g, batroc)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(batroc) {
		t.Fatal("Batroc is still on the battlefield")
	}
	if g.Battlefield.Contains(victim) {
		t.Error("the trigger did nothing once Batroc was gone: his last-known power 3 should kill the 3/3 (CR 608.2h)")
	}
}

// --- Captain Marvel, Apex Avenger ----------------------------------------------------

func msc17Marvel(t *testing.T, g *game.Game, owner uuid.UUID) uuid.UUID {
	t.Helper()
	return b12Push(g, owner, "Captain Marvel, Apex Avenger", "Legendary Creature — Human Kree Hero", msc17CaptainMarvelOra, 4, 4)
}

func msc17PutCounters(g *game.Game, actor, target uuid.UUID, kind string, n int) {
	g.WithWriteLock(func() { _ = g.AddCounterByForEffect(actor, target, kind, n) })
}

func TestMsc17CaptainMarvelCopiesCountersYouPutOnAnotherNonKree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	marvel := msc17Marvel(t, g, me.ID)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	msc17PutCounters(g, me.ID, bear, game.CounterPlusOne, 2)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if n := b12Counter(t, g, marvel, game.CounterPlusOne); n != 2 {
		t.Fatalf("Captain Marvel has %d +1/+1 counters, want 2", n)
	}

	// Any kind of counter, on any other creature (an opponent's too).
	msc17PutCounters(g, me.ID, theirs, "charge", 1)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if n := b12Counter(t, g, marvel, "charge"); n != 1 {
		t.Errorf("Captain Marvel has %d charge counters, want 1", n)
	}
	if n := b12Counter(t, g, marvel, game.CounterPlusOne); n != 2 {
		t.Errorf("the +1/+1 count moved to %d", n)
	}
}

func TestMsc17CaptainMarvelMayDeclineAndIgnoresKreeAndOtherPlayersCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	marvel := msc17Marvel(t, g, me.ID)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	kree := b12Creature(g, me.ID, "Kree Soldier", "Creature — Kree Soldier", 2, 2)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")

	msc17PutCounters(g, me.ID, bear, game.CounterPlusOne, 1)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if n := b12Counter(t, g, marvel, game.CounterPlusOne); n != 0 {
		t.Fatalf("declined, but Captain Marvel has %d counters", n)
	}

	msc17PutCounters(g, me.ID, kree, game.CounterPlusOne, 1)
	passPriorityAroundTable(t, g)
	msc17PutCounters(g, opp.ID, bear, game.CounterPlusOne, 1)
	passPriorityAroundTable(t, g)
	msc17PutCounters(g, me.ID, rock, game.CounterPlusOne, 1)
	passPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt {
			t.Fatalf("an optional trigger asked for a counter placement it should ignore: %+v", c)
		}
	}
	if n := b12Counter(t, g, marvel, game.CounterPlusOne); n != 0 {
		t.Errorf("Captain Marvel has %d counters, want 0 (a Kree, an opponent's placement and a non-creature put none)", n)
	}
}

func TestMsc17CaptainMarvelHasItsKeywords(t *testing.T) {
	g := newCatalogGame(t)
	marvel := msc17Marvel(t, g, g.Seats[0].ID)
	assertKeywords(t, g, marvel, "flying", "double strike", "indestructible")
}
