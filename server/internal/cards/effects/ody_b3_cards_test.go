package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ody_b3_cards_test.go — Odyssey, the third batch of #5613 (#5750 and
// #5798 were the first two). Each test plays the real catalog entry
// through the engine and asserts what its text says.

const (
	odyB3MysticCrusaderOracle   = "937d663d-c9a7-4312-a4cc-fcb54faa92bc"
	odyB3MysticEnforcerOracle   = "cf3b79cc-ce97-48e9-a578-625e323d1b62"
	odyB3FrightcrawlerOracle    = "96b611e9-38cc-4326-bb38-73084e99e6f5"
	odyB3ChildhoodHorrorOracle  = "29bebbaf-d47c-4369-8ff9-81816758da86"
	odyB3DedicatedMartyrOracle  = "ca39af65-985d-4ab1-874b-9e8d3455c349"
	odyB3DruidLyristOracle      = "b32a10ba-528c-4dc1-9828-b33f5d5a3091"
	odyB3DoggedHunterOracle     = "cf5b3249-619f-48f5-8885-00fd41d24825"
	odyB3BarbarianLunaticOracle = "12b5fbe5-a426-4b7e-bef0-abea71170e49"
	odyB3CephalidRetainerOracle = "9a0a926c-59ab-4bd3-a238-13988bdd96a5"
	odyB3WhisperingShadeOracle  = "0036d062-10dc-4267-98b4-1d2f6b190a61"
	odyB3LimestoneGolemOracle   = "0dddf5a9-fdf3-49e5-8103-a8cfd62f55c6"
	odyB3SkyshooterOracle       = "9c5a3407-e3e6-4f5f-9c24-6f24c19e1a2e"
	odyB3AvenArcherOracle       = "914ce8f6-5e92-436d-8541-1c72da8577dc"
	odyB3ThermalBlastOracle     = "52ad1a31-3ad8-4ead-ba77-675cad7e4a14"
	odyB3RefreshOracle          = "1a4c6643-c48b-4e83-b0c4-96873c13408c"
	odyB3KamahlOracle           = "153bb46f-ee8a-4da3-aab4-ae884f6ff403"
)

func TestOdyB3EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		odyB3MysticCrusaderOracle:   "Mystic Crusader",
		odyB3MysticEnforcerOracle:   "Mystic Enforcer",
		odyB3FrightcrawlerOracle:    "Frightcrawler",
		odyB3ChildhoodHorrorOracle:  "Childhood Horror",
		odyB3DedicatedMartyrOracle:  "Dedicated Martyr",
		odyB3DruidLyristOracle:      "Druid Lyrist",
		odyB3DoggedHunterOracle:     "Dogged Hunter",
		odyB3BarbarianLunaticOracle: "Barbarian Lunatic",
		odyB3CephalidRetainerOracle: "Cephalid Retainer",
		odyB3WhisperingShadeOracle:  "Whispering Shade",
		odyB3LimestoneGolemOracle:   "Limestone Golem",
		odyB3SkyshooterOracle:       "Skyshooter",
		odyB3AvenArcherOracle:       "Aven Archer",
		odyB3ThermalBlastOracle:     "Thermal Blast",
		odyB3RefreshOracle:          "Refresh",
		odyB3KamahlOracle:           "Kamahl, Pit Fighter",
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

func TestOdyB3PrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		keywords     []string
	}{
		{"Mystic Crusader", odyB3MysticCrusaderOracle, []string{"protection from black", "protection from red"}},
		{"Mystic Enforcer", odyB3MysticEnforcerOracle, []string{"protection from black"}},
		{"Frightcrawler", odyB3FrightcrawlerOracle, []string{"fear"}},
		{"Childhood Horror", odyB3ChildhoodHorrorOracle, []string{"flying"}},
		{"Whispering Shade", odyB3WhisperingShadeOracle, []string{"swampwalk"}},
		{"Skyshooter", odyB3SkyshooterOracle, []string{"reach"}},
		{"Aven Archer", odyB3AvenArcherOracle, []string{"flying"}},
		{"Kamahl, Pit Fighter", odyB3KamahlOracle, []string{"haste"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := blbB1Push(g, g.Seats[0].ID, tc.name, "Creature — Test", tc.oracle, 2, 2)
			for _, kw := range tc.keywords {
				if !hasKeywordOnBattlefield(t, g, id, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// Mystic Crusader's protection from red refuses a red spell's target, and
// Mystic Enforcer's protection from black alone does not (CR 702.16b).
func TestOdyB3MysticProtectionRefusesARedBoltOnlyForTheCrusader(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		refused      bool
	}{
		{"Mystic Crusader", odyB3MysticCrusaderOracle, true},
		{"Mystic Enforcer", odyB3MysticEnforcerOracle, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := blbB1Push(g, g.Seats[g.Turn.ActiveSeat].ID, tc.name, "Creature — Human Nomad Mystic", tc.oracle, 2, 2)
			if got := odyB2BoltRefused(t, g, id); got != tc.refused {
				t.Errorf("Lightning Bolt refused = %v, want %v", got, tc.refused)
			}
		})
	}
}

// The four threshold creatures: below seven cards in your own graveyard
// they are their printed selves; at seven they grow, and the Mystics fly
// while the Horrors can no longer block.
func TestOdyB3ThresholdCreaturesChangeAtSevenCards(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		bigP, bigT             int
		flies, cantBlock       bool
	}{
		{"Mystic Crusader", "Creature — Human Nomad Mystic", odyB3MysticCrusaderOracle, 2, 1, 3, 2, true, false},
		{"Mystic Enforcer", "Creature — Human Nomad Mystic", odyB3MysticEnforcerOracle, 3, 3, 6, 6, true, false},
		{"Frightcrawler", "Creature — Horror", odyB3FrightcrawlerOracle, 1, 1, 3, 3, false, true},
		{"Childhood Horror", "Creature — Horror", odyB3ChildhoodHorrorOracle, 2, 2, 4, 4, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp, _ := blbB1Seats(g)
			id := blbB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			// Seven cards in an OPPONENT's graveyard are not threshold.
			blbMPFillGraveyard(g, opp, 7)
			blbMPFillGraveyard(g, me, 6)
			if p, tough := ptOf(t, g, id); p != tc.power || tough != tc.toughness {
				t.Fatalf("six cards in your graveyard: %d/%d, want %d/%d", p, tough, tc.power, tc.toughness)
			}
			if restrictionsOf(t, g, id)&game.CantBlock != 0 {
				t.Fatal("can't block before threshold")
			}
			// Childhood Horror flies from the start; the Mystics only at threshold.
			if tc.name != "Childhood Horror" && hasKeywordOnBattlefield(t, g, id, "flying") {
				t.Fatal("flying before threshold")
			}
			blbMPFillGraveyard(g, me, 1)
			if p, tough := ptOf(t, g, id); p != tc.bigP || tough != tc.bigT {
				t.Errorf("seven cards in your graveyard: %d/%d, want %d/%d", p, tough, tc.bigP, tc.bigT)
			}
			if got := hasKeywordOnBattlefield(t, g, id, "flying"); got != tc.flies {
				t.Errorf("flying at threshold = %v, want %v", got, tc.flies)
			}
			if got := restrictionsOf(t, g, id)&game.CantBlock != 0; got != tc.cantBlock {
				t.Errorf("can't block at threshold = %v, want %v", got, tc.cantBlock)
			}
			if restrictionsOf(t, g, id)&game.CantAttack != 0 {
				t.Error("the threshold stopped the creature attacking")
			}
		})
	}
}

// odyB3TargetRefused activates the row at a target its clause does not
// allow and wants the announcement refused (CR 601.2c, 602.2b) with no
// cost paid: the source still untapped and on the battlefield, and
// nothing on the stack.
func odyB3TargetRefused(t *testing.T, g *game.Game, player, card, target uuid.UUID) {
	t.Helper()
	stack := g.Stack.Size()
	err := g.ActivateCatalogAbility(player, card, 0, game.ActivateAbilityParams{Targets: cardTarget(target)})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("activation at an illegal target: err = %v, want ErrIllegalTarget", err)
	}
	if !g.Battlefield.Contains(card) || b16Tapped(t, g, card) {
		t.Error("a refused activation paid its cost")
	}
	if g.Stack.Size() != stack {
		t.Error("a refused activation put something on the stack")
	}
}

// odyB3Source puts a catalog creature under the active seat in its main
// phase, not summoning sick, with `mana` in its pool.
func odyB3Source(t *testing.T, g *game.Game, name, typeLine, oracle string, mana ...string) (*game.Player, uuid.UUID) {
	t.Helper()
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := pushCatalogPermanent(g, me.ID, name, typeLine, oracle, false)
	b06AddMana(me, mana...)
	return me, id
}

func TestOdyB3DedicatedMartyrGainsThreeAsItIsSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me, martyr := odyB3Source(t, g, "Dedicated Martyr", "Creature — Human Cleric", odyB3DedicatedMartyrOracle, "W")
	life := me.Life
	b16Activate(t, g, me.ID, martyr, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(martyr) {
		t.Error("Dedicated Martyr was not sacrificed")
	}
	if me.Life != life+3 {
		t.Errorf("life %d -> %d, want +3", life, me.Life)
	}
}

func TestOdyB3DruidLyristDestroysAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, lyrist := odyB3Source(t, g, "Druid Lyrist", "Creature — Human Druid", odyB3DruidLyristOracle, "G")
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	odyB3TargetRefused(t, g, me.ID, lyrist, bear)
	ench := b12Creature(g, opp.ID, "Glorious Anthem", "Enchantment", 0, 0)
	b16Activate(t, g, me.ID, lyrist, 0, game.ActivateAbilityParams{Targets: cardTarget(ench)})
	if g.Battlefield.Contains(ench) {
		t.Error("the enchantment survived")
	}
	if g.Battlefield.Contains(lyrist) {
		t.Error("Druid Lyrist was not sacrificed")
	}
}

// A summoning-sick Lyrist can't pay its {T} (CR 302.6).
func TestOdyB3DruidLyristNeedsToBeUnsick(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp, _ := blbB1Seats(g)
	lyrist := pushCatalogPermanent(g, me.ID, "Druid Lyrist", "Creature — Human Druid", odyB3DruidLyristOracle, true)
	ench := b12Creature(g, opp.ID, "Glorious Anthem", "Enchantment", 0, 0)
	b06AddMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, lyrist, 0, game.ActivateAbilityParams{Targets: cardTarget(ench)}); err == nil {
		t.Error("a summoning-sick Lyrist activated its {T} ability")
	}
}

func TestOdyB3DoggedHunterDestroysOnlyACreatureToken(t *testing.T) {
	g := newCatalogGame(t)
	me, hunter := odyB3Source(t, g, "Dogged Hunter", "Creature — Human Nomad", odyB3DoggedHunterOracle)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	odyB3TargetRefused(t, g, me.ID, hunter, bear)
	token := b12Creature(g, opp.ID, "Squirrel", "Token Creature — Squirrel", 1, 1)
	b16Activate(t, g, me.ID, hunter, 0, game.ActivateAbilityParams{Targets: cardTarget(token)})
	if g.Battlefield.Contains(token) {
		t.Error("the creature token survived")
	}
	if !b16Tapped(t, g, hunter) {
		t.Error("Dogged Hunter did not tap")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the nontoken creature was destroyed")
	}
}

func TestOdyB3BarbarianLunaticDealsTwoAsItIsSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me, lunatic := odyB3Source(t, g, "Barbarian Lunatic", "Creature — Human Barbarian", odyB3BarbarianLunaticOracle, "C", "C", "R")
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	b16Activate(t, g, me.ID, lunatic, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)})
	if g.Battlefield.Contains(lunatic) {
		t.Error("Barbarian Lunatic was not sacrificed")
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the 2/2 survived 2 damage")
	}
}

func TestOdyB3CephalidRetainerTapsOnlyACreatureWithoutFlying(t *testing.T) {
	g := newCatalogGame(t)
	me, retainer := odyB3Source(t, g, "Cephalid Retainer", "Creature — Octopus", odyB3CephalidRetainerOracle, "U", "U")
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	flier := blbB1Push(g, opp.ID, "Aven Archer", "Creature — Bird Soldier Archer", odyB3AvenArcherOracle, 2, 2)
	odyB3TargetRefused(t, g, me.ID, retainer, flier)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	b16Activate(t, g, me.ID, retainer, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)})
	if !b16Tapped(t, g, bear) {
		t.Error("the creature without flying was not tapped")
	}
	if b16Tapped(t, g, retainer) {
		t.Error("Cephalid Retainer tapped itself; its cost has no {T}")
	}
}

func TestOdyB3WhisperingShadePumpsForB(t *testing.T) {
	g := newCatalogGame(t)
	me, shade := odyB3Source(t, g, "Whispering Shade", "Creature — Shade", odyB3WhisperingShadeOracle, "B", "B")
	b16Activate(t, g, me.ID, shade, 0, game.ActivateAbilityParams{})
	b16Activate(t, g, me.ID, shade, 0, game.ActivateAbilityParams{})
	if p, tough := ptOf(t, g, shade); p != 3 || tough != 3 {
		t.Errorf("Whispering Shade is %d/%d after two activations, want 3/3", p, tough)
	}
}

func TestOdyB3LimestoneGolemMakesATargetPlayerDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, golem := odyB3Source(t, g, "Limestone Golem", "Artifact Creature — Golem", odyB3LimestoneGolemOracle, "C", "C")
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	seedLibrary(opp, "A", "B")
	hand := opp.Hand.Size()
	b16Activate(t, g, me.ID, golem, 0, game.ActivateAbilityParams{Targets: acPlayerRef(opp.ID)})
	if g.Battlefield.Contains(golem) {
		t.Error("Limestone Golem was not sacrificed")
	}
	if opp.Hand.Size() != hand+1 {
		t.Errorf("target's hand %d -> %d, want +1", hand, opp.Hand.Size())
	}
}

// odyB3Attacking marks a battlefield creature as attacking `defender`.
func odyB3Attacking(g *game.Game, id, defender uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].AttackingTarget = defender
			}
		}
	})
}

func TestOdyB3SkyshooterDestroysAnAttackingFlier(t *testing.T) {
	g := newCatalogGame(t)
	me, shooter := odyB3Source(t, g, "Skyshooter", "Creature — Centaur Archer", odyB3SkyshooterOracle)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	flier := blbB1Push(g, opp.ID, "Aven Archer", "Creature — Bird Soldier Archer", odyB3AvenArcherOracle, 2, 2)
	walker := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	odyB3Attacking(g, walker, me.ID)
	// Neither a flier out of combat nor an attacker without flying.
	odyB3TargetRefused(t, g, me.ID, shooter, flier)
	odyB3TargetRefused(t, g, me.ID, shooter, walker)
	odyB3Attacking(g, flier, me.ID)
	b16Activate(t, g, me.ID, shooter, 0, game.ActivateAbilityParams{Targets: cardTarget(flier)})
	if g.Battlefield.Contains(flier) {
		t.Error("the attacking flier survived")
	}
	if g.Battlefield.Contains(shooter) {
		t.Error("Skyshooter was not sacrificed")
	}
}

func TestOdyB3AvenArcherShootsAnAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, archer := odyB3Source(t, g, "Aven Archer", "Creature — Bird Soldier Archer", odyB3AvenArcherOracle, "C", "C", "W")
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	odyB3TargetRefused(t, g, me.ID, archer, bear)
	odyB3Attacking(g, bear, me.ID)
	b16Activate(t, g, me.ID, archer, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)})
	if g.Battlefield.Contains(bear) {
		t.Error("the attacking 2/2 survived 2 damage")
	}
	if !b16Tapped(t, g, archer) {
		t.Error("Aven Archer did not tap")
	}
}

func TestOdyB3ThermalBlastDealsFiveAtThreshold(t *testing.T) {
	for _, tc := range []struct {
		name  string
		cards int
		dies  bool
	}{
		{"six cards: 3 damage", 6, false},
		{"seven cards: 5 damage", 7, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp, _ := blbB1Seats(g)
			blbMPFillGraveyard(g, me, tc.cards)
			ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
			castCatalogSpell(t, g, "Thermal Blast", "Instant", odyB3ThermalBlastOracle, cardTarget(ogre))
			passPriorityAroundTable(t, g)
			if got := !g.Battlefield.Contains(ogre); got != tc.dies {
				t.Errorf("the 4/4 died = %v, want %v", got, tc.dies)
			}
		})
	}
}

func TestOdyB3RefreshShieldsTheTargetAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	seedLibrary(me, "A", "B")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Refresh", "Instant", odyB3RefreshOracle, cardTarget(bear))
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	var shields int
	g.ReadSnapshot(func() { shields = g.RegenerationShieldsOn(bear) })
	if shields != 1 {
		t.Errorf("the target carries %d regeneration shields, want 1", shields)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, me.Hand.Size())
	}
}

// Kamahl's haste lets it use {T} the turn it arrives (CR 702.10c).
func TestOdyB3KamahlPingsForThreeTheTurnItArrives(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp, _ := blbB1Seats(g)
	kamahl := pushCatalogPermanent(g, me.ID, "Kamahl, Pit Fighter", "Legendary Creature — Human Barbarian", odyB3KamahlOracle, true)
	life := opp.Life
	b16Activate(t, g, me.ID, kamahl, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{playerTarget(opp.ID)}})
	if opp.Life != life-3 {
		t.Errorf("opponent life %d -> %d, want -3", life, opp.Life)
	}
	if !b16Tapped(t, g, kamahl) {
		t.Error("Kamahl did not tap")
	}
}
