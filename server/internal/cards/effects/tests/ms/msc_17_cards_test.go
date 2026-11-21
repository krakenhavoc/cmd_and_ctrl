package testsms

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestMsc17EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		effectstest.Msc17DaughterOracle:   "Daughter of the Deep",
		effectstest.Msc17DoormanOracle:    "Doorman",
		effectstest.Msc17AnnieOracle:      "Annie Joins Up",
		effectstest.Msc17SymbiosisOracle:  "Alien Symbiosis",
		effectstest.Msc17DragonManOracle:  "Dragon Man, Reformed Robot",
		effectstest.Msc17DoraOracle:       "Dora Milaje Elite",
		effectstest.Msc17ClosetOracle:     "Costume Closet",
		effectstest.Msc17FirstFamOracle:   "First Family",
		effectstest.Msc17GenesisOracle:    "Genesis Ultimatum",
		effectstest.Msc17DoomfallOracle:   "Doomfall",
		effectstest.Msc17HatutOracle:      "Hatut Zeraze Strike Force",
		effectstest.Msc17BatrocOracle:     "Batroc the Leaper",
		effectstest.Msc17CaptainMarvelOra: "Captain Marvel, Apex Avenger",
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
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
	g := effectstest.NewCatalogGame(t)
	me, other := g.Seats[1], g.Seats[2]
	effectstest.PushCatalogPermanent(g, me.ID, "Daughter of the Deep", "Creature — Merfolk Noble", effectstest.Msc17DaughterOracle, false)
	effectstest.HobA5Draw(t, g, other, nil)
	effectstest.HobA5Draw(t, g, other, nil)
	if n := effectstest.CountNamed(g, "Merfolk"); n != 0 {
		t.Fatalf("%d Merfolk after another player drew twice, want 0", n)
	}
	effectstest.HobA5Draw(t, g, me, nil)
	if n := effectstest.CountNamed(g, "Merfolk"); n != 0 {
		t.Fatalf("%d Merfolk after the first draw, want 0", n)
	}
	effectstest.HobA5Draw(t, g, me, nil)
	if n := effectstest.CountNamed(g, "Merfolk"); n != 1 {
		t.Fatalf("%d Merfolk after the second draw, want 1", n)
	}
	effectstest.HobA5Draw(t, g, me, nil)
	if n := effectstest.CountNamed(g, "Merfolk"); n != 1 {
		t.Errorf("%d Merfolk after the third draw, want it to stay 1", n)
	}
}

func TestMsc17DaughterMakesAMerfolkUnblockableForAPrice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	daughter := effectstest.PushCatalogPermanent(g, me.ID, "Daughter of the Deep", "Creature — Merfolk Noble", effectstest.Msc17DaughterOracle, false)
	merfolk := effectstest.B12Creature(g, opp.ID, "Their Merfolk", "Creature — Merfolk", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, daughter, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err == nil {
		t.Fatal("the ability targets a Merfolk, not any creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, daughter, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(merfolk), Strict: true}); err == nil {
		t.Fatal("the ability was activated without {U}")
	}
	effectstest.FloatMana(t, g, me, "{U}")
	if err := g.ActivateCatalogAbility(me.ID, daughter, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(merfolk), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertRestrictions(t, g, merfolk, game.CantBeBlocked)
	effectstest.AssertRestrictions(t, g, bear, 0)
	if c, _ := effectstest.BattlefieldCard(g, daughter); !c.Tapped {
		t.Error("{T} is part of the cost")
	}
}

// --- Doorman ----------------------------------------------------------------

func TestMsc17DoormanBarsSmallBlockersAndWalls(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	doorman := effectstest.PushCatalogPermanent(g, me.ID, "Doorman", "Legendary Creature — Mutant Hero", effectstest.Msc17DoormanOracle, false)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Ogre", 3, 3)
	bear := effectstest.B12Creature(g, opp.ID, "Small Bear", "Creature — Bear", 2, 2)
	wall := effectstest.B12Creature(g, opp.ID, "Big Wall", "Creature — Wall", 5, 5)
	ogre := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, doorman, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(attacker)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.BrRefusal(t, g.DeclareBlocker(bear, attacker), game.BlockReasonCantBeBlockedBy)
	effectstest.BrRefusal(t, g.DeclareBlocker(wall, attacker), game.BlockReasonCantBeBlockedBy)
	if err := g.DeclareBlocker(ogre, attacker); err != nil {
		t.Errorf("a non-Wall with power 3 blocks it: %v", err)
	}
}

func TestMsc17DoormanDoesNotBarBlockersFromACreatureItDidNotTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	doorman := effectstest.PushCatalogPermanent(g, me.ID, "Doorman", "Legendary Creature — Mutant Hero", effectstest.Msc17DoormanOracle, false)
	chosen := effectstest.B12Creature(g, me.ID, "Chosen", "Creature — Ogre", 3, 3)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Ogre", 3, 3)
	bear := effectstest.B12Creature(g, opp.ID, "Small Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, doorman, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(chosen)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, other)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(bear, other); err != nil {
		t.Errorf("the bar belongs to the target alone: %v", err)
	}
}

// --- Annie Joins Up -----------------------------------------------------------

func TestMsc17AnnieDealsFiveToACreatureOrPlaneswalkerAnOpponentControls(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
	mine := effectstest.B12Creature(g, me.ID, "My Giant", "Creature — Giant", 5, 5)
	effectstest.Msc17Cast(t, g, me, game.Card{Name: "Annie Joins Up", TypeLine: "Legendary Enchantment", OracleID: effectstest.Msc17AnnieOracle}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickIfAsked(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("five damage did not kill the opponent's 5/5")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the caster's own creature is not a legal target")
	}
}

func TestMsc17AnnieAsksForNothingWhenOnlyYourCreaturesExist(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.B12Creature(g, me.ID, "My Giant", "Creature — Giant", 5, 5)
	effectstest.Msc17Cast(t, g, me, game.Card{Name: "Annie Joins Up", TypeLine: "Legendary Enchantment", OracleID: effectstest.Msc17AnnieOracle}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Fatalf("a target was asked for with none legal: %+v", p)
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the caster's own creature died")
	}
}

func TestMsc17AnnieTriggersALegendaryCreaturesAbilityAnAdditionalTime(t *testing.T) {
	foods := func(legendary, withAnnie bool, controller int) int {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		owner := g.Seats[controller]
		if withAnnie {
			effectstest.PushCatalogPermanent(g, me.ID, "Annie Joins Up", "Legendary Enchantment", effectstest.Msc17AnnieOracle, false)
		}
		typeLine := "Creature — Bird"
		if legendary {
			typeLine = "Legendary Creature — Bird"
		}
		pigeon := effectstest.PushCatalogPermanent(g, owner.ID, "City Pigeon", typeLine, effectstest.Msc17PigeonOracle, false)
		effectstest.B27Kill(g, pigeon)
		effectstest.PassPriorityAroundTable(t, g)
		return effectstest.CountNamed(g, "Food")
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	_, err := effectstest.Msc17Cast(t, g, me, game.Card{Name: "Alien Symbiosis", TypeLine: "Enchantment — Aura", OracleID: effectstest.Msc17SymbiosisOracle, ManaCost: "{1}{B}"},
		game.CastSpellParams{Targets: effectstest.CardRefs(bear)})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("power %d, want 3", got)
	}
	effectstest.AssertKeywords(t, g, bear, "menace")
	if subs := effectstest.RingBFCard(t, g, bear).Effective().Subtypes; !effectstest.ContainsString(subs, "Symbiote") || !effectstest.ContainsString(subs, "Bear") {
		t.Errorf("subtypes %v, want Bear and Symbiote", subs)
	}
}

func TestMsc17AlienSymbiosisIsCastFromTheGraveyardByDiscardingACard(t *testing.T) {
	g, me, _, spare := effectstest.RetraceTable(t)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !game.CardCastableFromZone(effectstest.Msc17SymbiosisOracle, game.ZoneGraveyard) {
		t.Fatal("the graveyard is not open to Alien Symbiosis")
	}
	offers := game.AlternativeCostsOfferedFromZone(effectstest.Msc17SymbiosisOracle, game.ZoneGraveyard)
	if len(offers) != 1 || offers[0].ManaCost != "{1}{B}" || offers[0].CardPaymentCount() != 1 {
		t.Fatalf("graveyard offers = %+v, want the printed {1}{B} plus one discarded card", offers)
	}
	aura := effectstest.GraveyardCardOf(g, me, "Alien Symbiosis", "Enchantment — Aura", "{1}{B}", effectstest.Msc17SymbiosisOracle)
	params := game.CastSpellParams{FromZone: "graveyard", AlternativeCost: offers[0].Key, Strict: true, Targets: effectstest.CardRefs(bear)}

	// Without the discard, or without the mana, nothing is paid and nothing is cast.
	effectstest.PayMana(t, g, me, "{B}{B}")
	if err := g.CastSpell(me.ID, aura, params); err == nil {
		t.Fatal("cast from the graveyard with no card discarded")
	}
	if !effectstest.InGraveyard(me, aura) {
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

	effectstest.PayMana(t, g, me, "{B}{B}")
	if err := g.CastSpell(me.ID, aura, params); err != nil {
		t.Fatalf("cast from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aura) {
		t.Error("the Aura did not enter")
	}
	if !effectstest.InGraveyard(me, spare) {
		t.Error("the discarded card is not in the graveyard")
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("the recast Aura gave power %d, want 3", got)
	}
}

// --- Dragon Man, Reformed Robot -------------------------------------------

// msc17DragonMan builds a table with Dragon Man, `ours` noncreature permanents
// of the given mana costs under his controller, `theirs` under an opponent,
// and the given cards in his controller's graveyard (name, type line, cost).
func msc17DragonMan(t *testing.T, ours, theirs []string, graveyard [][3]string) (*game.Game, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for _, cost := range ours {
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", ManaCost: cost, Owner: me.ID, Controller: me.ID})
	}
	for _, cost := range theirs {
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Relic", TypeLine: "Artifact", ManaCost: cost, Owner: opp.ID, Controller: opp.ID})
	}
	for _, c := range graveyard {
		effectstest.GraveyardCardOf(g, me, c[0], c[1], c[2], "")
	}
	dragon := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dragon Man, Reformed Robot", TypeLine: "Legendary Artifact Creature — Dragon Robot",
		OracleID: effectstest.Msc17DragonManOracle, ManaCost: "{2}{W}{U}", Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	return g, dragon
}

func TestMsc17DragonManPowerIsTheGreatestNoncreatureManaValue(t *testing.T) {
	g, dragon := msc17DragonMan(t, nil, nil, nil)
	if got := effectstest.EffectivePower(t, g, dragon); got != 0 {
		t.Errorf("with nothing else, power %d, want 0 (he is a creature and counts nothing)", got)
	}
	if c := effectstest.RingBFCard(t, g, dragon); c.CurrentToughness() != 5 {
		t.Errorf("toughness %d, want 5", c.CurrentToughness())
	}

	// An artifact of mana value 3; an opponent's 9 and a creature's 8 do not count.
	g, dragon = msc17DragonMan(t, []string{"{3}"}, []string{"{9}"}, nil)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Big Creature", TypeLine: "Creature — Ogre", ManaCost: "{8}", Power: 8, Toughness: 8, Owner: g.Seats[0].ID, Controller: g.Seats[0].ID})
	if got := effectstest.EffectivePower(t, g, dragon); got != 3 {
		t.Errorf("power %d, want 3", got)
	}

	// A noncreature card in the graveyard counts, a creature card does not.
	g, dragon = msc17DragonMan(t, []string{"{3}"}, nil, [][3]string{{"Big Spell", "Sorcery", "{5}"}, {"Dead Ogre", "Creature — Ogre", "{7}"}})
	if got := effectstest.EffectivePower(t, g, dragon); got != 5 {
		t.Errorf("power %d, want 5 (the sorcery in the graveyard)", got)
	}
}

func TestMsc17DragonManIsCastFromTheGraveyardByDiscardingACard(t *testing.T) {
	g, me, _, spare := effectstest.RetraceTable(t)
	dragon := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dragon, Name: "Dragon Man, Reformed Robot", TypeLine: "Legendary Artifact Creature — Dragon Robot",
		ManaCost: "{2}{W}{U}", OracleID: effectstest.Msc17DragonManOracle, Toughness: 5, Owner: me.ID, Controller: me.ID,
		KnownBy: map[uuid.UUID]bool{me.ID: true}})
	offers := game.AlternativeCostsOfferedFromZone(effectstest.Msc17DragonManOracle, game.ZoneGraveyard)
	if len(offers) != 1 || offers[0].ManaCost != "{2}{W}{U}" {
		t.Fatalf("graveyard offers = %+v", offers)
	}
	params := game.CastSpellParams{FromZone: "graveyard", AlternativeCost: offers[0].Key, Strict: true}
	effectstest.PayMana(t, g, me, "{W}{U}{R}{R}")
	if err := g.CastSpell(me.ID, dragon, params); err == nil {
		t.Fatal("cast from the graveyard with no card discarded")
	}
	params.AltCostIDs = []uuid.UUID{spare}
	if err := g.CastSpell(me.ID, dragon, params); err != nil {
		t.Fatalf("cast from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dragon) {
		t.Error("Dragon Man did not enter")
	}
	if !effectstest.InGraveyard(me, spare) {
		t.Error("the discarded card is not in the graveyard")
	}
}

// --- Dora Milaje Elite -------------------------------------------------------

func msc17DoraCard() game.Card {
	return game.Card{Name: "Dora Milaje Elite", TypeLine: "Creature — Human Warrior", OracleID: effectstest.Msc17DoraOracle, ManaCost: "{1}{W}", Power: 2, Toughness: 2}
}

func TestMsc17DoraMilajeEliteMakesATappedVibraniumOnlyWhenAnOpponentHasMoreLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	effectstest.PushCatalogPermanent(g, opp.ID, "Forest", "Basic Land — Forest", "", false)
	effectstest.PushCatalogPermanent(g, opp.ID, "Forest", "Basic Land — Forest", "", false)
	effectstest.Msc17Play(t, g, me, msc17DoraCard())
	vib := msc17NamedOnBattlefield(g, "Vibranium")
	if len(vib) != 1 || !vib[0].Tapped || vib[0].Controller != me.ID {
		t.Fatalf("Vibranium tokens %+v, want one tapped token under the caster", vib)
	}
	if !vib[0].IsArtifact() {
		t.Error("Vibranium is an artifact")
	}

	even := effectstest.NewCatalogGame(t)
	me2, opp2 := even.Seats[0], even.Seats[1]
	effectstest.PushCatalogPermanent(even, me2.ID, "Forest", "Basic Land — Forest", "", false)
	effectstest.PushCatalogPermanent(even, opp2.ID, "Forest", "Basic Land — Forest", "", false)
	effectstest.Msc17Play(t, even, me2, msc17DoraCard())
	if n := len(msc17NamedOnBattlefield(even, "Vibranium")); n != 0 {
		t.Errorf("%d Vibranium with equal land counts, want 0", n)
	}
}

func TestMsc17DoraMilajeEliteRechecksTheLandCountOnResolution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, opp.ID, "Forest", "Basic Land — Forest", "", false)
	if _, err := effectstest.Msc17Cast(t, g, me, msc17DoraCard(), game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.Msc17UntilATriggerIsOnTheStack(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(msc17NamedOnBattlefield(g, "Vibranium")); n != 0 {
		t.Errorf("%d Vibranium after the lands became equal, want 0 (CR 603.4)", n)
	}
}

func TestMsc17VibraniumIsIndestructibleAndItsManaCannotCastNonartifactSpells(t *testing.T) {
	tok := effects.VibraniumToken()
	if !effectstest.ContainsString(tok.Keywords, "indestructible") || tok.TypeLine != "Token Artifact — Vibranium" {
		t.Errorf("token %q keywords %v", tok.TypeLine, tok.Keywords)
	}
	abilities := game.ManaAbilitiesForCard(tok)
	if len(abilities) != 1 || abilities[0].Produced != "{C}" || !abilities[0].TapCost ||
		len(abilities[0].Restrictions) != 1 || abilities[0].Restrictions[0] != game.ManaRestrictNotNonartifactSpell {
		t.Errorf("mana abilities %+v, want {T}: restricted {C}", abilities)
	}
}

func TestMsc17DoraMilajeEliteSacrificeGivesLegendaryPermanentsIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elite := effectstest.B12Push(g, me.ID, "Dora Milaje Elite", "Creature — Human Warrior", effectstest.Msc17DoraOracle, 2, 2)
	legend := effectstest.B12Creature(g, me.ID, "Legend", "Legendary Creature — Human", 2, 2)
	relic := effectstest.B12Permanent(g, me.ID, "Legend Relic", "Legendary Artifact")
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Human", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Legend", "Legendary Creature — Human", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, elite, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(elite) {
		t.Error("the Elite was not sacrificed")
	}
	for _, id := range []uuid.UUID{legend, relic} {
		if !effectstest.HasKeywordOnBattlefield(t, g, id, "indestructible") {
			t.Errorf("legendary permanent %s is not indestructible", id)
		}
	}
	for _, id := range []uuid.UUID{plain, theirs} {
		if effectstest.HasKeywordOnBattlefield(t, g, id, "indestructible") {
			t.Errorf("permanent %s gained indestructible but is not a legendary permanent you control", id)
		}
	}
}

// --- Costume Closet ----------------------------------------------------------

func msc17ClosetCard() game.Card {
	return game.Card{Name: "Costume Closet", TypeLine: "Artifact", OracleID: effectstest.Msc17ClosetOracle, ManaCost: "{1}{W}"}
}

func TestMsc17CostumeClosetEntersWithTwoCountersAndMovesThemAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	closet := effectstest.Msc17Play(t, g, me, msc17ClosetCard())
	if n := effectstest.B12Counter(t, g, closet, game.CounterPlusOne); n != 2 {
		t.Fatalf("enters with %d counters, want 2", n)
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, closet, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(theirs)}); err == nil {
		t.Fatal("moved a counter onto an opponent's creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, closet, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Counter(t, g, closet, game.CounterPlusOne); n != 1 {
		t.Errorf("the Closet has %d counters, want 1", n)
	}
	if n := effectstest.B12Counter(t, g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("the creature has %d counters, want 1", n)
	}
}

func TestMsc17CostumeClosetCannotActivateOutsideAMainPhase(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	closet := effectstest.Msc17Play(t, g, me, msc17ClosetCard())
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, closet, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Error("activated in the beginning of combat step")
	}
}

func TestMsc17CostumeClosetGrowsWhenAModifiedCreatureYouControlLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	closet := effectstest.Msc17Play(t, g, me, msc17ClosetCard())
	countersOnCloset := func() int { return effectstest.B12Counter(t, g, closet, game.CounterPlusOne) }
	equipped := effectstest.B12Creature(g, me.ID, "Equipped", "Creature — Bear", 2, 2)
	daggers := effectstest.SeedEquipment(g, me.ID, "Vibranium Energy Daggers", mshB10DaggersOracle)
	effectstest.EquipTo(t, g, me.ID, daggers, equipped)

	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	effectstest.B27Kill(g, plain)
	effectstest.PassPriorityAroundTable(t, g)
	if n := countersOnCloset(); n != 2 {
		t.Fatalf("an unmodified creature leaving changed the Closet to %d counters", n)
	}

	countered := effectstest.B12Creature(g, me.ID, "Countered", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterByForEffect(me.ID, countered, game.CounterPlusOne, 1) })
	effectstest.B27Kill(g, countered)
	effectstest.PassPriorityAroundTable(t, g)
	if n := countersOnCloset(); n != 3 {
		t.Fatalf("a creature with a counter leaving: Closet has %d counters, want 3", n)
	}

	theirs := effectstest.B12Creature(g, opp.ID, "Their Countered", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterByForEffect(opp.ID, theirs, game.CounterPlusOne, 1) })
	effectstest.B27Kill(g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if n := countersOnCloset(); n != 3 {
		t.Fatalf("an opponent's modified creature leaving changed the Closet to %d counters", n)
	}

	effectstest.B27Kill(g, equipped)
	effectstest.PassPriorityAroundTable(t, g)
	if n := countersOnCloset(); n != 4 {
		t.Errorf("an equipped creature leaving: Closet has %d counters, want 4", n)
	}
}

// --- First Family ----------------------------------------------------------------

func msc17FirstFamilyCard() game.Card {
	return game.Card{Name: "First Family", TypeLine: "Instant", OracleID: effectstest.Msc17FirstFamOracle, ManaCost: "{2}{G}{U}", Colors: []string{"U", "G"}}
}

func TestMsc17FirstFamilyCountsTheColorsOfPermanentsAndSpellsCast(t *testing.T) {
	// Alone: the spell is green and blue.
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand, life := me.Hand.Size(), me.Life
	_, err := effectstest.Msc17Cast(t, g, me, msc17FirstFamilyCard(), game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("hand grew by %d, want 2 cards drawn", got)
	}
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d life, want 2", got)
	}

	// With a red and a white permanent and a black spell cast earlier: five colors.
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	effectstest.PushBattlefieldCardWithTimestamp(g2, game.Card{InstanceID: uuid.New(), Name: "Red Ogre", TypeLine: "Creature — Ogre", Power: 2, Toughness: 2, Colors: []string{"R"}, Owner: me2.ID, Controller: me2.ID})
	effectstest.PushBattlefieldCardWithTimestamp(g2, game.Card{InstanceID: uuid.New(), Name: "White Knight", TypeLine: "Creature — Knight", Power: 2, Toughness: 2, Colors: []string{"W"}, Owner: me2.ID, Controller: me2.ID})
	effectstest.PushBattlefieldCardWithTimestamp(g2, game.Card{InstanceID: uuid.New(), Name: "Their Black", TypeLine: "Creature — Zombie", Power: 2, Toughness: 2, Colors: []string{"B"}, Owner: opp2.ID, Controller: opp2.ID})
	if _, err := effectstest.Msc17Cast(t, g2, me2, game.Card{Name: "Earlier Black Spell", TypeLine: "Sorcery", Colors: []string{"B"}}, game.CastSpellParams{}); err != nil {
		t.Fatalf("earlier spell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	hand2, life2 := me2.Hand.Size(), me2.Life
	if _, err := effectstest.Msc17Cast(t, g2, me2, msc17FirstFamilyCard(), game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	// R and W from permanents; B, G and U from spells (an opponent's black permanent adds nothing new to the set).
	if got := me2.Life - life2; got != 5 {
		t.Errorf("gained %d life, want 5 (R, W, B, G, U)", got)
	}
	if got := me2.Hand.Size() - hand2; got != 5 {
		t.Errorf("hand grew by %d, want 5 cards drawn", got)
	}
}

func TestMsc17FirstFamilyIgnoresAnOpponentsPermanentsAndSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Red", TypeLine: "Creature — Ogre", Power: 2, Toughness: 2, Colors: []string{"R"}, Owner: opp.ID, Controller: opp.ID})
	life := me.Life
	if _, err := effectstest.Msc17Cast(t, g, me, msc17FirstFamilyCard(), game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d life, want 2: only your permanents count", got)
	}
}

// --- Genesis Ultimatum ------------------------------------------------------------

func TestMsc17GenesisUltimatumPutsPermanentsOntoTheBattlefieldTheRestInHandAndExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	below := effectstest.PlTop(me, "Below", "Creature — Bear", "{1}{G}")
	land := effectstest.PlTop(me, "Forest", "Basic Land — Forest", "")
	sorcery := effectstest.PlTop(me, "Rite", "Sorcery", "{1}")
	artifact := effectstest.PlTop(me, "Rock", "Artifact", "{2}")
	instant := effectstest.PlTop(me, "Zap", "Instant", "{1}")
	bear := effectstest.PlTop(me, "Bear", "Creature — Bear", "{1}{G}")
	spell := effectstest.B12PlayFromHand(t, g, "Genesis Ultimatum", "Sorcery", effectstest.Msc17GenesisOracle, game.CastSpellParams{})
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick offered")
	}
	if effectstest.HasID(pick.ChooseCards, sorcery) || effectstest.HasID(pick.ChooseCards, instant) || effectstest.HasID(pick.ChooseCards, below) {
		t.Errorf("offered %v: only permanent cards among the top five may be put onto the battlefield", pick.ChooseCards)
	}
	if !effectstest.HasID(pick.ChooseCards, land) || !effectstest.HasID(pick.ChooseCards, artifact) || !effectstest.HasID(pick.ChooseCards, bear) {
		t.Errorf("offered %v, want the land, the artifact and the creature", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, bear, artifact)
	effectstest.PassPriorityAroundTable(t, g)
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
	if !g.Exile.Contains(spell) || effectstest.InGraveyard(me, spell) {
		t.Error("Genesis Ultimatum is not exiled")
	}
}

func TestMsc17GenesisUltimatumChoosingNothingPutsAllFiveInHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	var top []uuid.UUID
	for _, n := range []string{"A", "B", "C", "D", "E"} {
		top = append(top, effectstest.PlTop(me, n, "Creature — Bear", "{1}"))
	}
	spell := effectstest.B12PlayFromHand(t, g, "Genesis Ultimatum", "Sorcery", effectstest.Msc17GenesisOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
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
	return game.Card{Name: "Doomfall", TypeLine: "Sorcery", OracleID: effectstest.Msc17DoomfallOracle, ManaCost: "{2}{B}"}
}

func TestMsc17DoomfallLetsTheOpponentChooseWhichCreatureToExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	strong := effectstest.B12Creature(g, opp.ID, "Strong", "Creature — Ogre", 5, 5)
	weak := effectstest.B12Creature(g, opp.ID, "Weak", "Creature — Bear", 1, 1)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	if _, err := effectstest.Msc17Cast(t, g, me, msc17DoomfallCard(), game.CastSpellParams{Modes: []int{0}, Targets: []game.TargetRef{effectstest.Msc17TargetPlayer(me.ID)}}); err == nil {
		t.Fatal("Doomfall targeted its own caster")
	}
	if _, err := effectstest.Msc17Cast(t, g, me, msc17DoomfallCard(), game.CastSpellParams{Modes: []int{0}, Targets: []game.TargetRef{effectstest.Msc17TargetPlayer(opp.ID)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("the caster was asked to choose: the opponent chooses")
	}
	pick := effectstest.ChooseCardsChoiceFor(g, opp.ID)
	if pick == nil || !effectstest.HasID(pick.ChooseCards, strong) || !effectstest.HasID(pick.ChooseCards, weak) || effectstest.HasID(pick.ChooseCards, mine) {
		t.Fatalf("the opponent's pick %+v should offer exactly their creatures", pick)
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, weak)
	if !g.Exile.Contains(weak) {
		t.Error("the chosen creature is not exiled")
	}
	if !g.Battlefield.Contains(strong) || !g.Battlefield.Contains(mine) {
		t.Error("an unchosen creature left the battlefield")
	}
}

func TestMsc17DoomfallExileEdictAsksNothingOfAnOpponentWithNoCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	if _, err := effectstest.Msc17Cast(t, g, me, msc17DoomfallCard(), game.CastSpellParams{Modes: []int{0}, Targets: []game.TargetRef{effectstest.Msc17TargetPlayer(opp.ID)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, opp.ID) != nil || effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a choice was asked with no creature to exile")
	}
}

func TestMsc17DoomfallExilesANonlandCardChosenFromTheRevealedHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ids := effectstest.RevealHand(opp, effectstest.RhBolt(), game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	if _, err := effectstest.Msc17Cast(t, g, me, msc17DoomfallCard(), game.CastSpellParams{Modes: []int{1}, Targets: []game.TargetRef{effectstest.Msc17TargetPlayer(opp.ID)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.OpenVariant(t, g)
	if effectstest.HasID(pick.DiscardOptions, ids[1]) || !effectstest.HasID(pick.DiscardOptions, ids[0]) {
		t.Fatalf("options %v: only the nonland card may be chosen", pick.DiscardOptions)
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(ids[0]) || opp.Graveyard.Contains(ids[0]) {
		t.Error("the chosen card is not exiled")
	}
	if !opp.Hand.Contains(ids[1]) {
		t.Error("the land left their hand")
	}
}

// --- Hatut Zeraze Strike Force ------------------------------------------------------

func msc17HatutCard() game.Card {
	return game.Card{Name: "Hatut Zeraze Strike Force", TypeLine: "Creature — Human Spy Warrior", OracleID: effectstest.Msc17HatutOracle, ManaCost: "{3}{W}", Power: 2, Toughness: 2}
}

func TestMsc17HatutCopiesItselfForEachCommanderCastFromTheCommandZone(t *testing.T) {
	for _, tc := range []struct {
		casts, want int
	}{{0, 1}, {1, 2}, {2, 3}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		g.WithWriteLock(func() { me.CommanderCasts = map[uuid.UUID]int{uuid.New(): tc.casts} })
		effectstest.Msc17Play(t, g, me, msc17HatutCard())
		if got := len(msc17NamedOnBattlefield(g, "Hatut Zeraze Strike Force")); got != tc.want {
			t.Errorf("after %d commander casts: %d Hatut on the battlefield, want %d", tc.casts, got, tc.want)
		}
	}
}

func TestMsc17HatutIgnoresAnOpponentsCommanderCasts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { opp.CommanderCasts = map[uuid.UUID]int{uuid.New(): 3} })
	effectstest.Msc17Play(t, g, me, msc17HatutCard())
	if got := len(msc17NamedOnBattlefield(g, "Hatut Zeraze Strike Force")); got != 1 {
		t.Errorf("%d Hatut, want 1", got)
	}
}

func TestMsc17HatutDestroysAnArtifactOrEnchantmentWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.Msc17Cast(t, g, me, msc17HatutCard(), game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickIfAsked(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("a creature was destroyed")
	}
}

// --- Batroc the Leaper ----------------------------------------------------------------

func msc17BatrocCard() game.Card {
	return game.Card{Name: "Batroc the Leaper", TypeLine: "Legendary Creature — Human Villain", OracleID: effectstest.Msc17BatrocOracle, ManaCost: "{1}{R}", Power: 2, Toughness: 2}
}

func TestMsc17BatrocGrowsAndDamagesUpToXTargetsForEachKick(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Ogre", 4, 4)
	survivor := effectstest.B12Creature(g, opp.ID, "Survivor", "Creature — Ogre", 5, 5)
	life := opp.Life
	batroc, err := effectstest.Msc17Cast(t, g, me, msc17BatrocCard(), game.CastSpellParams{OptionalCosts: []int{0, 0}})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Counter(t, g, batroc, game.CounterPlusOne); n != 2 {
		t.Fatalf("%d counters, want one per kick (2)", n)
	}
	p := effectstest.LatestPickTarget(g, me.ID)
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
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Bear", 1, 1)
	batroc, err := effectstest.Msc17Cast(t, g, me, msc17BatrocCard(), game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Fatalf("an unkicked Batroc asked for targets: %+v", p)
	}
	if n := effectstest.B12Counter(t, g, batroc, game.CounterPlusOne); n != 0 {
		t.Errorf("%d counters, want 0", n)
	}
	if !g.Battlefield.Contains(bystander) {
		t.Error("an unkicked Batroc dealt damage")
	}
}

func TestMsc17BatrocDealsHisLastKnownPowerIfHeLeavesBeforeTheTriggerResolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Ogre", 3, 3)
	batroc, err := effectstest.Msc17Cast(t, g, me, msc17BatrocCard(), game.CastSpellParams{OptionalCosts: []int{0}})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.ResolveTopOfStack(t, g) // the creature spell; the trigger now needs its target
	effectstest.PickIfAsked(t, g, me.ID, victim)
	effectstest.B27Kill(g, batroc)
	effectstest.PassPriorityAroundTable(t, g)
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
	return effectstest.B12Push(g, owner, "Captain Marvel, Apex Avenger", "Legendary Creature — Human Kree Hero", effectstest.Msc17CaptainMarvelOra, 4, 4)
}

func msc17PutCounters(g *game.Game, actor, target uuid.UUID, kind string, n int) {
	g.WithWriteLock(func() { _ = g.AddCounterByForEffect(actor, target, kind, n) })
}

func TestMsc17CaptainMarvelCopiesCountersYouPutOnAnotherNonKree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	marvel := msc17Marvel(t, g, me.ID)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	msc17PutCounters(g, me.ID, bear, game.CounterPlusOne, 2)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Counter(t, g, marvel, game.CounterPlusOne); n != 2 {
		t.Fatalf("Captain Marvel has %d +1/+1 counters, want 2", n)
	}

	// Any kind of counter, on any other creature (an opponent's too).
	msc17PutCounters(g, me.ID, theirs, "charge", 1)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Counter(t, g, marvel, "charge"); n != 1 {
		t.Errorf("Captain Marvel has %d charge counters, want 1", n)
	}
	if n := effectstest.B12Counter(t, g, marvel, game.CounterPlusOne); n != 2 {
		t.Errorf("the +1/+1 count moved to %d", n)
	}
}

func TestMsc17CaptainMarvelMayDeclineAndIgnoresKreeAndOtherPlayersCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	marvel := msc17Marvel(t, g, me.ID)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	kree := effectstest.B12Creature(g, me.ID, "Kree Soldier", "Creature — Kree Soldier", 2, 2)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")

	msc17PutCounters(g, me.ID, bear, game.CounterPlusOne, 1)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Counter(t, g, marvel, game.CounterPlusOne); n != 0 {
		t.Fatalf("declined, but Captain Marvel has %d counters", n)
	}

	msc17PutCounters(g, me.ID, kree, game.CounterPlusOne, 1)
	effectstest.PassPriorityAroundTable(t, g)
	msc17PutCounters(g, opp.ID, bear, game.CounterPlusOne, 1)
	effectstest.PassPriorityAroundTable(t, g)
	msc17PutCounters(g, me.ID, rock, game.CounterPlusOne, 1)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt {
			t.Fatalf("an optional trigger asked for a counter placement it should ignore: %+v", c)
		}
	}
	if n := effectstest.B12Counter(t, g, marvel, game.CounterPlusOne); n != 0 {
		t.Errorf("Captain Marvel has %d counters, want 0 (a Kree, an opponent's placement and a non-creature put none)", n)
	}
}

func TestMsc17CaptainMarvelHasItsKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	marvel := msc17Marvel(t, g, g.Seats[0].ID)
	effectstest.AssertKeywords(t, g, marvel, "flying", "double strike", "indestructible")
}
