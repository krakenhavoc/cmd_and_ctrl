package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_m6_cards_test.go — Marvel Super Heroes Commander, M–Z batch 6 (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	mscM6GauntletsOracle  = "4a9d36eb-db72-459a-8de9-6d04a1d7563c"
	mscM6WhipOracle       = "c077f939-73a9-455f-8de5-385162121b94"
	mscM6SuitOracle       = "2bc0821a-5a8f-465d-8532-3fb5c5d11d8f"
	mscM6NovaFlameOracle  = "f4310dfd-52eb-4c1f-b1ad-bfbc9b50fe7b"
	mscM6ScourglassOracle = "78541a40-c27a-49c0-aa3e-5a48b97065f0"
	mscM6SpiderUKOracle   = "ebf8d006-2997-431b-b4cb-9692741b45ca"
	mscM6WebSlingerOracle = "505231d8-714f-4c5b-a635-d1662bd18a78"
	mscM6TuskerOracle     = "9e8ffb63-1a7b-4ebe-839d-d5192522db65"
	mscM6PeggyOracle      = "2f3c6b8e-df89-4f0c-97bb-fbfcda5232b4"
	mscM6TaxOracle        = "ace8e88e-ca6a-4bdf-8c6d-1b914fc00259"
	mscM6MoonBoyOracle    = "bb7638ce-3802-4198-89d2-6f5b647bd8fb"
	mscM6LavamancerOracle = "5fc60833-6e03-4bfa-833f-7d51c696db93"
	mscM6OriginCapOracle  = "cf4ce0f2-101e-46ec-9368-7395872910b3"
)

func TestMscM6EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mscM6GauntletsOracle:  "Vibranium Strike Gauntlets",
		mscM6WhipOracle:       "Voltaic Whip",
		mscM6SuitOracle:       "Unstable Molecule Suit",
		mscM6NovaFlameOracle:  "Nova Flame",
		mscM6ScourglassOracle: "Scourglass",
		mscM6SpiderUKOracle:   "Spider-UK",
		mscM6WebSlingerOracle: "Spider-Man, Web-Slinger",
		mscM6TuskerOracle:     "Wakandan Tusker",
		mscM6PeggyOracle:      "Peggy Carter, Secret Agent",
		mscM6TaxOracle:        "Monologue Tax",
		mscM6MoonBoyOracle:    "Moon-Boy, Dino Rider",
		mscM6LavamancerOracle: "Molten Lavamancer",
		mscM6OriginCapOracle:  "Origin of Captain America",
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

// --- Vibranium Strike Gauntlets -----------------------------------------------

// mscM6CastGauntlets casts the Gauntlets from hand, answers the enters
// trigger's target with `host` and lets everything resolve.
func mscM6CastGauntlets(t *testing.T, g *game.Game, host uuid.UUID) uuid.UUID {
	t.Helper()
	id := castCatalogSpell(t, g, "Vibranium Strike Gauntlets", "Artifact — Equipment", mscM6GauntletsOracle, nil)
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if latestPickTarget(g, g.Seats[g.Turn.ActiveSeat].ID) != nil {
			break
		}
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	pickCard(t, g, g.Seats[g.Turn.ActiveSeat].ID, host)
	passPriorityAroundTable(t, g)
	return id
}

func TestMscM6GauntletsEnterAttachedAndGrantTheCombatDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)

	gauntlets := mscM6CastGauntlets(t, g, bear)
	if host := attachmentHostOf(t, g, gauntlets); host.ID != bear {
		t.Fatalf("the Gauntlets did not attach to the target on entry: %+v", host)
	}
	if got := effectivePower(t, g, bear); got != 5 {
		t.Errorf("equipped power %d, want 5 (2 + 3)", got)
	}
	if got := effectivePower(t, g, other); got != 2 {
		t.Errorf("unequipped power %d, want 2", got)
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the equipped creature lacks trample")
	}
	if hasEffectiveKeyword(t, g, other, "trample") {
		t.Error("the unequipped creature has trample")
	}

	hand := me.Hand.Size()
	attackWith(t, g, opp.ID, bear, other)
	passPriorityAroundTable(t, g)
	if opp.Life != 40-5-2 {
		t.Errorf("opponent life %d, want 33", opp.Life)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1 (only the equipped creature draws)", got)
	}
}

func TestMscM6GauntletsDrawNothingWhenBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	mscM6CastGauntlets(t, g, bear)

	hand := me.Hand.Size()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, bear); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("drew %d cards off a blocked attacker, want 0 (it dealt no damage to a player)", got)
	}
}

func TestMscM6GauntletsEquipMovesThePumpAndTheDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	other := seedBear(g, me.ID)
	gauntlets := seedEquipment(g, me.ID, "Vibranium Strike Gauntlets", mscM6GauntletsOracle)

	equipTo(t, g, me.ID, gauntlets, bear)
	equipTo(t, g, me.ID, gauntlets, other)
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("the old host has power %d, want 2", got)
	}
	if got := effectivePower(t, g, other); got != 5 {
		t.Errorf("the new host has power %d, want 5", got)
	}
}

func TestMscM6GauntletsEquipIsRefusedWithoutThreeMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	gauntlets := seedEquipment(g, me.ID, "Vibranium Strike Gauntlets", mscM6GauntletsOracle)

	floatMana(t, g, me, "{C}{C}")
	err := g.ActivateCatalogAbility(me.ID, gauntlets, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	if err == nil {
		t.Fatal("equip {3} was accepted with two mana")
	}
}

func TestMscM6GauntletsAreCastableAtInstantSpeedButTheWhipIsNot(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepBeginCombat)
	whip := handCardFull(me, "Voltaic Whip", "Artifact — Equipment", "{2}{B}", mscM6WhipOracle, []string{"B"})
	if err := g.CastSpell(me.ID, whip, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a non-flash Equipment was cast at the beginning of combat")
	}
	gauntlets := handCardFull(me, "Vibranium Strike Gauntlets", "Artifact — Equipment", "{4}", mscM6GauntletsOracle, nil)
	floatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.CastSpell(me.ID, gauntlets, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("flash Gauntlets at the beginning of combat: %v", err)
	}
}

// --- Voltaic Whip -----------------------------------------------------------------

func mscM6WhipBoard(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	armed := b12Creature(g, me.ID, "Armed Bear", "Creature — Bear", 2, 2)
	plain := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	whip := seedEquipment(g, me.ID, "Voltaic Whip", mscM6WhipOracle)
	equipTo(t, g, me.ID, whip, armed)
	if got := effectivePower(t, g, armed); got != 4 {
		t.Fatalf("equipped power %d, want 4 (2 + 2)", got)
	}
	return g, me, opp, armed, plain
}

func TestMscM6WhipDrawsAndLosesLifeWhenTheEquippedCreatureAttacksAlone(t *testing.T) {
	g, me, opp, armed, _ := mscM6WhipBoard(t)
	hand, life := me.Hand.Size(), me.Life

	declareAttack(t, g, opp.ID, armed)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
}

func TestMscM6WhipDoesNothingWhenTheEquippedCreatureAttacksWithAnother(t *testing.T) {
	g, me, opp, armed, plain := mscM6WhipBoard(t)
	hand, life := me.Hand.Size(), me.Life

	declareAttack(t, g, opp.ID, armed, plain)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("drew %d, want 0 (it did not attack alone)", got)
	}
	if me.Life != life {
		t.Errorf("life %d, want %d", me.Life, life)
	}
}

func TestMscM6WhipDoesNothingWhenAnUnequippedCreatureAttacksAlone(t *testing.T) {
	g, me, opp, _, plain := mscM6WhipBoard(t)
	hand := me.Hand.Size()

	declareAttack(t, g, opp.ID, plain)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("drew %d, want 0 (the Whip's creature did not attack)", got)
	}
}

// --- Unstable Molecule Suit --------------------------------------------------------

func TestMscM6SuitPumpsAndMakesTheCreatureIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	suit := seedEquipment(g, me.ID, "Unstable Molecule Suit", mscM6SuitOracle)
	equipTo(t, g, me.ID, suit, bear)

	if effectivePower(t, g, bear) != 4 || effectiveToughness(t, g, bear) != 4 {
		t.Errorf("equipped creature is %d/%d, want 4/4", effectivePower(t, g, bear), effectiveToughness(t, g, bear))
	}
	hobDestroy(t, g, bear)
	if !g.Battlefield.Contains(bear) {
		t.Error("the equipped creature was destroyed despite indestructible")
	}
}

func TestMscM6SuitCheapEquipIsCommanderOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	plain := seedBear(g, me.ID)
	general := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "General", TypeLine: "Legendary Creature — Human Soldier",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, IsCommander: true,
	})
	suit := seedEquipment(g, me.ID, "Unstable Molecule Suit", mscM6SuitOracle)

	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, suit, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: plain}},
	}); err == nil {
		t.Fatal("equip commander {2} accepted a creature that is not a commander")
	}
	if err := g.ActivateCatalogAbility(me.ID, suit, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: general}},
	}); err != nil {
		t.Fatalf("equip commander {2} on the commander: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left in the pool, want the {2} spent", me.ManaPool.Total())
	}
	if host := attachmentHostOf(t, g, suit); host.ID != general {
		t.Fatalf("the Suit is on %v, want the commander", host.ID)
	}

	// The plain Equip {4} takes any creature you control.
	floatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, suit, 1, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: plain}},
	}); err != nil {
		t.Fatalf("equip {4}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if host := attachmentHostOf(t, g, suit); host.ID != plain {
		t.Errorf("the Suit is on %v, want the plain creature", host.ID)
	}
}

// --- Nova Flame -----------------------------------------------------------------------

func mscM6NovaFlameCast(t *testing.T, g *game.Game, x int, target uuid.UUID) {
	t.Helper()
	b19CastXModal(t, g, "Nova Flame", "Sorcery", mscM6NovaFlameOracle, "{X}{2}{R}{R}", x, nil,
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
	passPriorityAroundTable(t, g)
}

func TestMscM6NovaFlameGrowsTheCreatureAndItDamagesEveryOtherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	hero := b27Push(g, me.ID, "Nova Hero", "Creature — Hero", "", "", 2, 2)
	ally := b12Creature(g, me.ID, "Ally", "Creature — Bear", 3, 3)
	big := b12Creature(g, opp.ID, "Big", "Creature — Bear", 4, 4)
	small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 1, 1)
	giant := b12Creature(g, opp.ID, "Giant", "Creature — Bear", 5, 5)

	mscM6NovaFlameCast(t, g, 2, hero)

	if got := plusOnes(g, hero); got != 2 {
		t.Errorf("Nova Hero has %d +1/+1 counters, want 2", got)
	}
	if !g.Battlefield.Contains(hero) {
		t.Fatal("the creature that dealt the damage died of it")
	}
	if damageMarkedOn(g, hero) != 0 {
		t.Errorf("the source took %d damage, want none (each OTHER creature)", damageMarkedOn(g, hero))
	}
	for name, id := range map[string]uuid.UUID{"ally": ally, "big": big, "small": small} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s survived 4 damage", name)
		}
	}
	if !g.Battlefield.Contains(giant) || damageMarkedOn(g, giant) != 4 {
		t.Errorf("the 5/5 should survive with 4 damage marked (on battlefield %v, damage %d)",
			g.Battlefield.Contains(giant), damageMarkedOn(g, giant))
	}
}

func TestMscM6NovaFlameWithXZeroDealsJustItsOwnPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	hero := b27Push(g, me.ID, "Nova Hero", "Creature — Hero", "", "", 2, 2)
	big := b12Creature(g, opp.ID, "Big", "Creature — Bear", 4, 4)
	small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 1, 1)

	mscM6NovaFlameCast(t, g, 0, hero)

	if plusOnes(g, hero) != 0 {
		t.Errorf("counters with X=0: %d", plusOnes(g, hero))
	}
	if g.Battlefield.Contains(small) {
		t.Error("the 1/1 survived 2 damage")
	}
	if !g.Battlefield.Contains(big) || damageMarkedOn(g, big) != 2 {
		t.Errorf("the 4/4 should have 2 damage marked (on battlefield %v, damage %d)",
			g.Battlefield.Contains(big), damageMarkedOn(g, big))
	}
}

func TestMscM6NovaFlameDamageComesFromTheCreatureSoLifelinkApplies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	hero := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Lifelinker", TypeLine: "Creature — Hero", Keywords: []string{"lifelink"},
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	b12Creature(g, opp.ID, "Victim", "Creature — Bear", 1, 9)
	life := me.Life

	mscM6NovaFlameCast(t, g, 1, hero)

	if me.Life != life+3 {
		t.Errorf("life %d, want %d (the 3-power lifelinker dealt 3 damage)", me.Life, life+3)
	}
}

func TestMscM6NovaFlameRefusesACreatureYouDoNotControl(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	advanceToMain(t, g)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	err := castCatalogSpellErr(t, g, "Nova Flame", "Sorcery", mscM6NovaFlameOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	if err == nil {
		t.Error("Nova Flame was cast on a creature its caster does not control")
	}
}

// --- Scourglass ----------------------------------------------------------------------------

func TestMscM6ScourglassDestroysEverythingButArtifactsAndLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToUpkeepOf(t, g, 0)
	glass := pushCatalogPermanent(g, me.ID, "Scourglass", "Artifact", mscM6ScourglassOracle, false)
	myCreature := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	myRune := pushPermanent(g, me.ID, game.Card{Name: "My Rune", TypeLine: "Enchantment"})
	myRock := pushPermanent(g, me.ID, game.Card{Name: "My Rock", TypeLine: "Artifact"})
	myLand := pushPermanent(g, me.ID, game.Card{Name: "My Forest", TypeLine: "Basic Land — Forest"})
	theirCreature := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	theirWalker := pushPermanent(g, opp.ID, game.Card{Name: "Their Walker", TypeLine: "Legendary Planeswalker — Test"})
	theirGolem := pushPermanent(g, opp.ID, game.Card{Name: "Their Golem", TypeLine: "Artifact Creature — Golem", Power: 3, Toughness: 3})
	theirLand := pushPermanent(g, opp.ID, game.Card{Name: "Their Island", TypeLine: "Basic Land — Island"})
	sturdy := pushPermanent(g, opp.ID, game.Card{Name: "Sturdy Bear", TypeLine: "Creature — Bear", Keywords: []string{"indestructible"}, Power: 2, Toughness: 2})

	if err := g.ActivateCatalogAbility(me.ID, glass, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate in my upkeep: %v", err)
	}
	if g.Battlefield.Contains(glass) {
		t.Error("the sacrifice cost was not paid at announce")
	}
	passPriorityAroundTable(t, g)

	for name, id := range map[string]uuid.UUID{"my Bear": myCreature, "my Rune": myRune, "their Bear": theirCreature, "their Walker": theirWalker} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s survived", name)
		}
	}
	for name, id := range map[string]uuid.UUID{"my Rock": myRock, "my Forest": myLand, "their Golem": theirGolem, "their Island": theirLand, "indestructible Bear": sturdy} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s was destroyed", name)
		}
	}
}

func TestMscM6ScourglassOnlyActivatesDuringYourOwnUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	glass := pushCatalogPermanent(g, me.ID, "Scourglass", "Artifact", mscM6ScourglassOracle, false)
	// The harness parks the cursor on seat 0's draw step: too late.
	acRefused(t, g, me.ID, glass, 0, game.ActivateAbilityParams{})
	// An opponent's upkeep is not yours either.
	advanceToUpkeepOf(t, g, 1)
	acRefused(t, g, me.ID, glass, 0, game.ActivateAbilityParams{})
	advanceToMainOf(t, g, 0)
	acRefused(t, g, me.ID, glass, 0, game.ActivateAbilityParams{})
}

// --- Spider-Man, Web-Slinger / Spider-UK --------------------------------------------

func TestMscM6WebSlingerCostsOnlyWhiteWithAReturnedTappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	spider := handCardFull(me, "Spider-Man, Web-Slinger", "Legendary Creature — Spider Human Hero", "{2}{W}", mscM6WebSlingerOracle, []string{"W"})
	tapped := pushTappedVanillaCreature(g, me.ID, "Tapped Bear", 2, 2)
	floatMana(t, g, me, "{W}")

	if err := g.CastSpell(me.ID, spider, game.CastSpellParams{
		Strict:          true,
		AlternativeCost: "web_slinging",
		AltCostIDs:      []uuid.UUID{tapped},
	}); err != nil {
		t.Fatalf("web-slinging cast: %v", err)
	}
	if !me.Hand.Contains(tapped) {
		t.Error("the tapped creature did not return to hand")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(spider) {
		t.Error("Spider-Man did not enter the battlefield")
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left, want the single {W} spent", me.ManaPool.Total())
	}
}

func TestMscM6WebSlingerRefusesAnUntappedCreatureAndTheFullPriceNeedsFullMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	spider := handCardFull(me, "Spider-Man, Web-Slinger", "Legendary Creature — Spider Human Hero", "{2}{W}", mscM6WebSlingerOracle, []string{"W"})
	untapped := pushVanillaCreature(g, me.ID, "Untapped Bear", 2, 2)
	floatMana(t, g, me, "{W}")

	if err := g.CastSpell(me.ID, spider, game.CastSpellParams{
		Strict:          true,
		AlternativeCost: "web_slinging",
		AltCostIDs:      []uuid.UUID{untapped},
	}); err == nil {
		t.Error("web-slinging accepted an untapped creature")
	}
	if !g.Battlefield.Contains(untapped) {
		t.Error("a refused payment moved the creature")
	}
	if err := g.CastSpell(me.ID, spider, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("the {2}{W} price was accepted with a single {W}")
	}
}

func TestMscM6SpiderUKWebSlingsForTwoWhite(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	spider := handCardFull(me, "Spider-UK", "Legendary Creature — Spider Human Hero", "{3}{W}", mscM6SpiderUKOracle, []string{"W"})
	tapped := pushTappedVanillaCreature(g, me.ID, "Tapped Bear", 2, 2)
	floatMana(t, g, me, "{C}{C}{W}")

	if err := g.CastSpell(me.ID, spider, game.CastSpellParams{
		Strict:          true,
		AlternativeCost: "web_slinging",
		AltCostIDs:      []uuid.UUID{tapped},
	}); err != nil {
		t.Fatalf("web-slinging cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(spider) || !me.Hand.Contains(tapped) {
		t.Error("Spider-UK did not enter with the tapped creature returned")
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left, want {2}{W} spent", me.ManaPool.Total())
	}
}

func mscM6SpiderUKEndStep(t *testing.T, g *game.Game, extra func()) (hand, life int, me *game.Player) {
	t.Helper()
	me = g.Seats[0]
	advanceToMain(t, g)
	b12PlayFromHand(t, g, "Spider-UK", "Legendary Creature — Spider Human Hero", mscM6SpiderUKOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	extra()
	hand, life = me.Hand.Size(), me.Life
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	return hand, life, me
}

func TestMscM6SpiderUKDrawsAndGainsWhenASecondCreatureEntered(t *testing.T) {
	g := newCatalogGame(t)
	hand, life, me := mscM6SpiderUKEndStep(t, g, func() {
		b12PlayFromHand(t, g, "Second Bear", "Creature — Bear", "", game.CastSpellParams{})
		passPriorityAroundTable(t, g)
	})
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestMscM6SpiderUKCountsATokenAsTheSecondCreature(t *testing.T) {
	g := newCatalogGame(t)
	hand, _, me := mscM6SpiderUKEndStep(t, g, func() {
		g.WithWriteLock(func() {
			ctx := NewContext(g, &game.StackItem{ID: uuid.New(), Controller: mscM6Seat0(g), SourceCardID: uuid.Nil})
			if err := (CreateToken{Template: TokenCard("1/1 white Soldier"), N: 1, Controller: mscM6Seat0(g)}).Apply(ctx); err != nil {
				t.Fatalf("token: %v", err)
			}
		})
	})
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1 (a token entering counts)", got)
	}
}

func mscM6Seat0(g *game.Game) uuid.UUID { return g.Seats[0].ID }

func TestMscM6SpiderUKDoesNothingWithOnlyItselfOrAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	hand, life, me := mscM6SpiderUKEndStep(t, g, func() {
		g.WithWriteLock(func() {
			ctx := NewContext(g, &game.StackItem{ID: uuid.New(), Controller: opp.ID, SourceCardID: uuid.Nil})
			if err := (CreateToken{Template: TokenCard("1/1 white Soldier"), N: 1, Controller: opp.ID}).Apply(ctx); err != nil {
				t.Fatalf("token: %v", err)
			}
		})
	})
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("drew %d, want 0 (only one creature entered under my control)", got)
	}
	if me.Life != life {
		t.Errorf("life %d, want %d", me.Life, life)
	}
}

// --- Wakandan Tusker -------------------------------------------------------------------

func TestMscM6TuskerGainsLifeAndScriesWhenTappedByAnEffect(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	seedLibrary(me, "A", "B", "C")
	tusker := pushCatalogPermanent(g, me.ID, "Wakandan Tusker", "Creature — Elephant", mscM6TuskerOracle, false)
	life := me.Life

	g.WithWriteLock(func() { _ = g.TapTargetForEffect(tusker) })
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
	answerScryKeepAll(t, g, me.ID)
}

func TestMscM6TuskerTriggersWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	seedLibrary(me, "A", "B", "C")
	tusker := b12Push(g, me.ID, "Wakandan Tusker", "Creature — Elephant", mscM6TuskerOracle, 2, 3)
	life := me.Life

	declareAttack(t, g, opp.ID, tusker)
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d, want %d (attacking taps it)", me.Life, life+1)
	}
	answerScryKeepAll(t, g, me.ID)
}

func TestMscM6TuskerIgnoresAnAlreadyTappedCreatureStayingTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	tusker := pushCatalogPermanent(g, me.ID, "Wakandan Tusker", "Creature — Elephant", mscM6TuskerOracle, false)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(tusker) })
	passPriorityAroundTable(t, g)
	answerScryKeepAll(t, g, me.ID)
	passPriorityAroundTable(t, g)
	life := me.Life

	// Another creature becoming tapped is not this one.
	other := pushCatalogPermanent(g, me.ID, "Other Bear", "Creature — Bear", "", false)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(other) })
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d, want %d (a different creature became tapped)", me.Life, life)
	}
}

// --- Peggy Carter, Secret Agent -----------------------------------------------------

func TestMscM6PeggyGivesTheLoneAttackerIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Peggy Carter, Secret Agent", "Legendary Creature — Human Spy Hero", mscM6PeggyOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("the lone attacker did not gain indestructible")
	}
}

func TestMscM6PeggyGivesHerselfIndestructibleToo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	peggy := pushCatalogPermanent(g, me.ID, "Peggy Carter, Secret Agent", "Legendary Creature — Human Spy Hero", mscM6PeggyOracle, false)

	declareAttack(t, g, opp.ID, peggy)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, peggy, "indestructible") {
		t.Error("Peggy attacking alone did not gain indestructible")
	}
}

func TestMscM6PeggyDoesNothingWhenTwoCreaturesAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	peggy := pushCatalogPermanent(g, me.ID, "Peggy Carter, Secret Agent", "Legendary Creature — Human Spy Hero", mscM6PeggyOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	declareAttack(t, g, opp.ID, peggy, bear)
	passPriorityAroundTable(t, g)
	if hasEffectiveKeyword(t, g, bear, "indestructible") || hasEffectiveKeyword(t, g, peggy, "indestructible") {
		t.Error("indestructible was granted though two creatures attacked")
	}
}

// --- Monologue Tax ---------------------------------------------------------------------------

func TestMscM6MonologueTaxMakesATreasureOnAnOpponentsSecondSpellOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Monologue Tax", "Enchantment", mscM6TaxOracle, false)

	batch01OpponentCasts(t, g, opp, "First Spell", "", "", nil)
	passPriorityAroundTable(t, g)
	if got := mshB1Tokens(g, me.ID, "Treasure"); got != 0 {
		t.Fatalf("%d Treasures after one opposing spell, want 0", got)
	}
	batch01OpponentCasts(t, g, opp, "Second Spell", "", "", nil)
	passPriorityAroundTable(t, g)
	if got := mshB1Tokens(g, me.ID, "Treasure"); got != 1 {
		t.Fatalf("%d Treasures after the second spell, want 1", got)
	}
	batch01OpponentCasts(t, g, opp, "Third Spell", "", "", nil)
	passPriorityAroundTable(t, g)
	if got := mshB1Tokens(g, me.ID, "Treasure"); got != 1 {
		t.Errorf("%d Treasures after the third spell, want still 1", got)
	}
}

func TestMscM6MonologueTaxIgnoresYourOwnSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Monologue Tax", "Enchantment", mscM6TaxOracle, false)

	castFree(t, g, me, "Spell One", "Instant")
	castFree(t, g, me, "Spell Two", "Instant")
	if got := mshB1Tokens(g, me.ID, "Treasure"); got != 0 {
		t.Errorf("%d Treasures from my own spells, want 0", got)
	}
}

// --- Moon-Boy, Dino Rider -----------------------------------------------------------------

func TestMscM6MoonBoyMakesYourDinosaurSpellsCostOneLess(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Moon-Boy, Dino Rider", "Legendary Creature — Ape Human", mscM6MoonBoyOracle, false)

	card := func(owner uuid.UUID, typeLine string) game.Card {
		return game.Card{InstanceID: uuid.New(), Name: "Test", TypeLine: typeLine, ManaCost: "{2}{G}", Owner: owner, Controller: owner}
	}
	price := func(p *game.Player, c game.Card) string {
		pr, err := g.PriceCast(p.ID, c, game.CastSpellParams{})
		if err != nil {
			t.Fatalf("PriceCast: %v", err)
		}
		return pr.Total.String()
	}
	if got := price(me, card(me.ID, "Creature — Dinosaur")); got != "{1}{G}" {
		t.Errorf("my Dinosaur spell costs %s, want {1}{G}", got)
	}
	if got := price(me, card(me.ID, "Creature — Elephant")); got != "{2}{G}" {
		t.Errorf("my Elephant spell costs %s, want the full {2}{G}", got)
	}
	if got := price(opp, card(opp.ID, "Creature — Dinosaur")); got != "{2}{G}" {
		t.Errorf("an opponent's Dinosaur spell costs %s, want the full {2}{G}", got)
	}
}

func TestMscM6MoonBoyGrowsWhenHeAttacksWhileYouControlADinosaur(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	moon := b12Push(g, me.ID, "Moon-Boy, Dino Rider", "Legendary Creature — Ape Human", mscM6MoonBoyOracle, 2, 2)
	b12Creature(g, me.ID, "Devil Dinosaur", "Creature — Dinosaur", 6, 5)

	declareAttack(t, g, opp.ID, moon)
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, moon); got != 3 {
		t.Errorf("Moon-Boy has power %d, want 3", got)
	}
	if got := effectiveToughness(t, g, moon); got != 3 {
		t.Errorf("Moon-Boy has toughness %d, want 3", got)
	}
}

func TestMscM6MoonBoyStaysSmallWithoutADinosaurOrWithAnOpposingOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	moon := b12Push(g, me.ID, "Moon-Boy, Dino Rider", "Legendary Creature — Ape Human", mscM6MoonBoyOracle, 2, 2)
	b12Creature(g, opp.ID, "Their Dinosaur", "Creature — Dinosaur", 6, 5)

	declareAttack(t, g, opp.ID, moon)
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, moon); got != 2 {
		t.Errorf("Moon-Boy has power %d, want 2 (I control no Dinosaur)", got)
	}
}

// --- Molten Lavamancer ------------------------------------------------------------------------

func mscM6Lavamancer(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	lava := b12Push(g, me.ID, "Molten Lavamancer", "Creature — Elemental Warlock Villain", mscM6LavamancerOracle, 2, 3)
	return g, me, opp, lava
}

func TestMscM6LavamancerMakesAnElementalOnceEachTurnForNoncombatDamageToAnOpponent(t *testing.T) {
	g, me, opp, lava := mscM6Lavamancer(t)

	p2Ping(t, g, lava, opp.ID, 1)
	passPriorityAroundTable(t, g)
	if got := tokensNamedControlledBy(g, "Elemental", me.ID); got != 1 {
		t.Fatalf("%d Elementals after noncombat damage to an opponent, want 1", got)
	}
	p2Ping(t, g, lava, opp.ID, 1)
	passPriorityAroundTable(t, g)
	if got := tokensNamedControlledBy(g, "Elemental", me.ID); got != 1 {
		t.Errorf("%d Elementals after a second hit the same turn, want still 1", got)
	}
}

func TestMscM6LavamancerIgnoresDamageToYouToCreaturesAndCombatDamage(t *testing.T) {
	g, me, opp, lava := mscM6Lavamancer(t)
	theirs := p2Tough(g, opp.ID, "Theirs")

	p2Ping(t, g, lava, me.ID, 1)
	p2Ping(t, g, lava, theirs, 1)
	passPriorityAroundTable(t, g)
	if got := tokensNamedControlledBy(g, "Elemental", me.ID); got != 0 {
		t.Fatalf("%d Elementals from damage to me or a creature, want 0", got)
	}

	attackWith(t, g, opp.ID, lava)
	passPriorityAroundTable(t, g)
	if opp.Life >= 40 {
		t.Fatalf("combat damage did not connect: life %d", opp.Life)
	}
	if got := tokensNamedControlledBy(g, "Elemental", me.ID); got != 0 {
		t.Errorf("%d Elementals from combat damage, want 0", got)
	}
}

func TestMscM6LavamancerOnlyWorksOnYourTurn(t *testing.T) {
	g, me, opp, lava := mscM6Lavamancer(t)
	advanceToMainOf(t, g, 1)

	p2Ping(t, g, lava, opp.ID, 1)
	passPriorityAroundTable(t, g)
	if got := tokensNamedControlledBy(g, "Elemental", me.ID); got != 0 {
		t.Errorf("%d Elementals from damage on an opponent's turn, want 0", got)
	}
}

func TestMscM6LavamancerHasProwess(t *testing.T) {
	spec, _ := Lookup(mscM6LavamancerOracle)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "prowess" {
		t.Errorf("printed keywords %v, want prowess", spec.PrintedKeywords)
	}
}

// --- Origin of Captain America -------------------------------------------------------

func TestMscM6OriginOfCaptainAmericaRunsAllThreeChapters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	hero := b12Creature(g, me.ID, "Hero", "Creature — Human Hero", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 3, 3)
	pickHero := &game.TargetRef{Kind: game.TargetCard, ID: hero}
	pickTheirs := &game.TargetRef{Kind: game.TargetCard, ID: theirs}

	saga := castCatalogSpell(t, g, "Origin of Captain America", "Enchantment — Saga", mscM6OriginCapOracle, nil)
	assistCardSettle(t, g, me.ID, pickHero, false)
	if got := plusOnes(g, hero); got != 1 {
		t.Errorf("chapter I: %d +1/+1 counters, want 1", got)
	}
	if !hasEffectiveKeyword(t, g, hero, "first strike") || !hasEffectiveKeyword(t, g, hero, "vigilance") {
		t.Error("chapter I: the creature lacks first strike or vigilance")
	}

	advanceToPrecombatMainOf(t, g, 0)
	assistCardSettle(t, g, me.ID, nil, false)
	if got := mshB1Tokens(g, me.ID, "Equipment"); got != 1 {
		t.Fatalf("chapter II: %d Equipment tokens, want 1", got)
	}
	var shield uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Sturdy Shield" && c.Controller == me.ID {
				shield = c.InstanceID
			}
		}
	})
	if shield == uuid.Nil {
		t.Fatal("chapter II: no Sturdy Shield")
	}
	if host := attachmentHostOf(t, g, shield); host.ID != uuid.Nil {
		t.Errorf("the Sturdy Shield came attached to %v, want unattached", host.ID)
	}
	if err := g.ActivateCatalogAbility(me.ID, shield, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: hero}},
	}); err != nil {
		t.Fatalf("equip {2} on the Sturdy Shield: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := hobA4Power(t, g, hero); got != 4 {
		t.Errorf("equipped power %d, want 4 (2 + counter + 1)", got)
	}
	if got := mscM6Toughness(g, hero); got != 5 {
		t.Errorf("equipped toughness %d, want 5 (2 + counter + 2)", got)
	}

	advanceToPrecombatMainOf(t, g, 0)
	assistCardSettle(t, g, me.ID, pickTheirs, false)
	if !b16Tapped(t, g, theirs) {
		t.Error("chapter III: the creature was not tapped")
	}
	if got := counterCount(g, theirs, game.CounterStun); got != 1 {
		t.Errorf("chapter III: %d stun counters, want 1", got)
	}
	if loreCountersOn(g, saga) >= 0 {
		t.Error("the Saga was not sacrificed after chapter III")
	}
}

func TestMscM6OriginOfCaptainAmericaChapterIIIMayTargetNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	hero := b12Creature(g, me.ID, "Hero", "Creature — Human Hero", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 3, 3)

	castCatalogSpell(t, g, "Origin of Captain America", "Enchantment — Saga", mscM6OriginCapOracle, nil)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: hero}, false)
	advanceToPrecombatMainOf(t, g, 0)
	assistCardSettle(t, g, me.ID, nil, false)
	advanceToPrecombatMainOf(t, g, 0)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("chapter III opened no target prompt")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
		t.Fatalf("choosing no target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if b16Tapped(t, g, theirs) || b16Tapped(t, g, hero) {
		t.Error("a creature was tapped though no target was chosen")
	}
}

// mscM6Toughness is the counter-aware toughness after the layers have run.
func mscM6Toughness(g *game.Game, id uuid.UUID) int {
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if c := findBattlefieldCardByID(g, id); c != nil {
		return c.CurrentToughness()
	}
	return -1
}
