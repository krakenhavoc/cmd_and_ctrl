package effects

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rix_5531_b1_cards_test.go — Rivals of Ixalan, batch 1 (#5531): four
// vanilla bodies, five creatures whose whole text is keywords, and
// thirty-nine cards whose text a catalogued card already prints (enrage,
// raid, entry triggers, statics, Auras and spells). Each test plays the
// real catalog entry through the engine and asserts what its text says.

const (
	rixB1StormFleetSprinter = "8d244f4a-1d62-4697-81c7-fbc4de874add"
	rixB1GiltgroveStalker   = "2be3969b-a427-47be-a49c-47b98e3d6b5f"
	rixB1RiverDarter        = "939a6815-bed1-43b0-bf4d-f0581ba0fed5"
	rixB1HardyVeteran       = "fc112282-bc40-4b91-ac5f-ae49db1dabb8"
	rixB1GraspingScoundrel  = "d8c3298f-da96-46ee-8b46-ce1d0df95f62"
	rixB1Neckbreaker        = "bcf34a1c-162a-400c-b1db-9d1bb62bf7e9"
	rixB1KnightOfStampede   = "801d60c1-40b1-48b6-8800-afe1a21386f3"
	rixB1RelentlessRaptor   = "4727db28-98e9-453d-90c3-2d51365d32fc"
	rixB1EverdawnChampion   = "964e065f-380d-491b-9225-244c5f9edda7"
	rixB1FamishedPaladin    = "f80df9c4-f682-4a70-81fe-03c7c5c256e1"
	rixB1Horncrest          = "fa6413f9-aa32-4fdc-a539-327a4ea906fe"
	rixB1Pterodon           = "747dbf0c-dcea-49d1-ab6c-d74016d580a4"
	rixB1LuminousBonds      = "2c006a63-0550-4a87-a5fe-f4247857271c"
	rixB1Waterknot          = "002100cd-1ab1-44da-93a0-1269d47a712a"
	rixB1ImperialCeratops   = "86cee7f0-0aac-46f1-8e4e-062fc398477e"
	rixB1OvergrownArmasaur  = "9e104718-1f74-4414-977a-82def93b5a8e"
	rixB1SiegehornCeratops  = "f7730e2e-aa62-40e7-836d-e6ddb6893816"
	rixB1FrilledSpitter     = "a65c1fcc-61d1-43d9-a62c-65f156f98bce"
	rixB1NeedletoothRaptor  = "2c431df9-095a-44f1-993e-74eb215d75fe"
	rixB1Cacophodon         = "8bd6dea2-2fdd-482f-9fe8-6f3eeed13872"
	rixB1SwaggeringCorsair  = "48fa47c0-544e-4374-a0a6-ef8f65cdf0dd"
	rixB1BrazenFreebooter   = "13e8d063-eece-4f38-83f6-02403130defa"
	rixB1CrestedHerdcaller  = "86c0ddaa-6926-450d-949a-6c89f43c71c2"
	rixB1JunglebornPioneer  = "60662463-6050-46c6-84ab-64e203b67b4d"
	rixB1JadecraftArtisan   = "9290f549-3e5d-4e39-a20e-8be6b674abc7"
	rixB1JadeBearer         = "72ddbc9b-f584-433f-9d2d-6e4c374129c4"
	rixB1SanguineGlorifier  = "391f1331-7577-4bf4-a9ed-2f55ded4cbd8"
	rixB1VoraciousVampire   = "c584bdf2-b328-4e50-81e3-45ca6c2aff7a"
	rixB1SwiftWarden        = "0928ddda-20ea-462e-9f00-6c10ef1cfb4b"
	rixB1TombRobber         = "cc5ffb09-3d08-4a52-a880-50c94df30c15"
	rixB1SunCollaredRaptor  = "9debaed0-b605-4778-9977-216d194ecdeb"
	rixB1Impale             = "dcdf40a9-5a7e-45b1-aee4-36dda598baaf"
	rixB1ReaverAmbush       = "0385be55-8e8e-4e7d-b6b7-51adc6eb98fa"
	rixB1Hornswoggle        = "71b178bd-3b37-4de9-97a3-96e23d5444a4"
	rixB1StrengthOfThePack  = "be78a76c-e126-4f09-9e8e-b773a437cfeb"
	rixB1ShakeFoundations   = "99eb5cda-684d-4f6c-bc34-1f03c8797150"
	rixB1EnterTheUnknown    = "4b031bb3-4bbc-4158-8485-b3fc42c6bd50"
	rixB1CleansingRay       = "50f1948c-fa27-458d-be14-cb0e09e39f63"
	rixB1CrashingTide       = "4c3f87d7-a632-4b8c-9052-b4055ea88771"

	rixB1Bear     = "Creature — Bear"
	rixB1Dinosaur = "Creature — Dinosaur"
	rixB1Merfolk  = "Creature — Merfolk Warrior"
	rixB1Vampire  = "Creature — Vampire Knight"
	rixB1Pirate   = "Creature — Human Pirate"
)

// rixB1Vanilla are the batch's bodies with no rules text.
var rixB1Vanilla = map[string]string{
	"9dadeecc-1d4b-4799-9ac9-4ca492006e46": "Canal Monitor",
	"2da32ec3-f42f-4edc-b5b3-e5f39fcf370c": "Orazca Frillback",
	"507ffc26-6c45-4167-9e0d-3c88b065720a": "Orazca Raptor",
	"9f63ab29-b86e-4337-b9dd-2b35c8415f90": "Sworn Guardian",
}

// rixB1Keyworded are the batch's creatures whose whole text is printed
// keywords.
var rixB1Keyworded = map[string]struct {
	name, typeLine string
	keywords       []string
}{
	"d7ebd4a7-c542-469f-8a1c-d48b92fa540a": {"Exultant Skymarcher", "Creature — Vampire Soldier", []string{"flying"}},
	"4db00a5e-7dbe-4387-90e9-c21ec1daa448": {"Sun Sentinel", "Creature — Human Soldier", []string{"vigilance"}},
	"a48d2be2-395a-4377-a2ad-51dbbc5b4e17": {"Vampire Champion", "Creature — Vampire Soldier", []string{"deathtouch"}},
	"2f7305c1-51d2-4698-89d5-c64a57ecff80": {"Vampire Revenant", "Creature — Vampire Spirit", []string{"flying"}},
	"011b2243-b8be-471f-9d8f-824a40e599df": {"Soul of the Rapids", "Creature — Elemental", []string{"flying", "hexproof"}},
}

// rixB1PickAndSettle answers chooser's open pick-target prompt, if there
// is one, with ref and settles the stack. A clause with a single legal
// target may be filled without a prompt, so a missing prompt is not a
// failure; the assertions after it are.
func rixB1PickAndSettle(t *testing.T, g *game.Game, chooser uuid.UUID, ref game.TargetRef) {
	t.Helper()
	for i := 0; i < 4; i++ {
		if p := latestPickTarget(g, chooser); p != nil {
			if err := g.ResolvePickTarget(p.ID, chooser, ref); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
		}
		passPriorityAroundTable(t, g)
		if latestPickTarget(g, chooser) == nil && stackFullyEmpty(g) {
			return
		}
	}
}

// rixB1CardRef is the pick-target answer for a card.
func rixB1CardRef(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetCard, ID: id}
}

// rixB1CastWithModes casts a catalog spell from the active player's hand
// in a main phase and returns the announce error.
func rixB1CastWithModes(t *testing.T, g *game.Game, name, typeLine, oracle string, modes []int, targets []game.TargetRef) error {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := b13HandCard(active, name, typeLine, oracle)
	advanceToMain(t, g)
	return g.CastSpell(active.ID, id, game.CastSpellParams{Modes: modes, Targets: targets})
}

func TestRix5531B1EveryCardIsRegisteredFull(t *testing.T) {
	all := map[string]string{
		rixB1StormFleetSprinter: "Storm Fleet Sprinter",
		rixB1GiltgroveStalker:   "Giltgrove Stalker",
		rixB1RiverDarter:        "River Darter",
		rixB1HardyVeteran:       "Hardy Veteran",
		rixB1GraspingScoundrel:  "Grasping Scoundrel",
		rixB1Neckbreaker:        "Dire Fleet Neckbreaker",
		rixB1KnightOfStampede:   "Knight of the Stampede",
		rixB1RelentlessRaptor:   "Relentless Raptor",
		rixB1EverdawnChampion:   "Everdawn Champion",
		rixB1FamishedPaladin:    "Famished Paladin",
		rixB1Horncrest:          "Stampeding Horncrest",
		rixB1Pterodon:           "Sun-Crested Pterodon",
		rixB1LuminousBonds:      "Luminous Bonds",
		rixB1Waterknot:          "Waterknot",
		rixB1ImperialCeratops:   "Imperial Ceratops",
		rixB1OvergrownArmasaur:  "Overgrown Armasaur",
		rixB1SiegehornCeratops:  "Siegehorn Ceratops",
		rixB1FrilledSpitter:     "Frilled Deathspitter",
		rixB1NeedletoothRaptor:  "Needletooth Raptor",
		rixB1Cacophodon:         "Cacophodon",
		rixB1SwaggeringCorsair:  "Swaggering Corsair",
		rixB1BrazenFreebooter:   "Brazen Freebooter",
		rixB1CrestedHerdcaller:  "Crested Herdcaller",
		rixB1JunglebornPioneer:  "Jungleborn Pioneer",
		rixB1JadecraftArtisan:   "Jadecraft Artisan",
		rixB1JadeBearer:         "Jade Bearer",
		rixB1SanguineGlorifier:  "Sanguine Glorifier",
		rixB1VoraciousVampire:   "Voracious Vampire",
		rixB1SwiftWarden:        "Swift Warden",
		rixB1TombRobber:         "Tomb Robber",
		rixB1SunCollaredRaptor:  "Sun-Collared Raptor",
		rixB1Impale:             "Impale",
		rixB1ReaverAmbush:       "Reaver Ambush",
		rixB1Hornswoggle:        "Hornswoggle",
		rixB1StrengthOfThePack:  "Strength of the Pack",
		rixB1ShakeFoundations:   "Shake the Foundations",
		rixB1EnterTheUnknown:    "Enter the Unknown",
		rixB1CleansingRay:       "Cleansing Ray",
		rixB1CrashingTide:       "Crashing Tide",
	}
	for oracle, name := range rixB1Vanilla {
		all[oracle] = name
	}
	for oracle, k := range rixB1Keyworded {
		all[oracle] = k.name
	}
	if len(all) != 48 {
		t.Fatalf("the batch lists %d cards, want 48", len(all))
	}
	for oracle, name := range all {
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

// --- bodies and keywords -----------------------------------------------------

func TestRix5531B1VanillaBodiesDeclareNoAbilities(t *testing.T) {
	for oracle, name := range rixB1Vanilla {
		spec, _ := Lookup(oracle)
		if spec.OnResolve != nil || len(spec.Triggered) != 0 || len(spec.Static) != 0 ||
			len(spec.Activated) != 0 || len(spec.PrintedKeywords) != 0 {
			t.Errorf("%s is a vanilla body and declares no abilities", name)
		}
	}
}

func TestRix5531B1CanalMonitorIsAPlainCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	monitor := b12Push(g, me.ID, "Canal Monitor", "Creature — Lizard", "9dadeecc-1d4b-4799-9ac9-4ca492006e46", 5, 3)
	bear := b12Creature(g, opp.ID, "Bear", rixB1Bear, 2, 2)
	if err := porBlock(t, g, monitor, opp.ID, bear); err != nil {
		t.Errorf("a vanilla Canal Monitor couldn't be blocked: %v", err)
	}
	if p, tough := ptOf(t, g, monitor); p != 5 || tough != 3 {
		t.Errorf("Canal Monitor is %d/%d, want 5/3", p, tough)
	}
}

func TestRix5531B1KeywordedCreaturesHaveTheirKeywords(t *testing.T) {
	for oracle, k := range rixB1Keyworded {
		t.Run(k.name, func(t *testing.T) {
			spec, _ := Lookup(oracle)
			if !slices.Equal(spec.PrintedKeywords, k.keywords) {
				t.Errorf("printed keywords %v, want %v", spec.PrintedKeywords, k.keywords)
			}
			if spec.OnResolve != nil || len(spec.Triggered) != 0 || len(spec.Static) != 0 || len(spec.Activated) != 0 {
				t.Errorf("%s declares more than its keywords", k.name)
			}
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, k.name, k.typeLine, oracle, 2, 2)
			for _, kw := range k.keywords {
				if !hasEffectiveKeyword(t, g, id, kw) {
					t.Errorf("%s has no %s on the battlefield", k.name, kw)
				}
			}
		})
	}
}

// Flying: a ground creature can't block Vampire Revenant (CR 702.9b).
func TestRix5531B1VampireRevenantIsBlockedOnlyByAFlyer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	revenant := b12Push(g, me.ID, "Vampire Revenant", "Creature — Vampire Spirit", "2f7305c1-51d2-4698-89d5-c64a57ecff80", 3, 1)
	bear := b12Creature(g, opp.ID, "Bear", rixB1Bear, 2, 2)
	if err := porBlock(t, g, revenant, opp.ID, bear); err == nil {
		t.Error("a ground creature blocked Vampire Revenant")
	}
}

// --- block rules and statics --------------------------------------------------

func TestRix5531B1StormFleetSprinterHasHasteAndCantBeBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sprinter := b12Push(g, me.ID, "Storm Fleet Sprinter", rixB1Pirate, rixB1StormFleetSprinter, 2, 2)
	if !hasEffectiveKeyword(t, g, sprinter, "haste") {
		t.Error("Storm Fleet Sprinter has no haste")
	}
	assertRestrictions(t, g, sprinter, game.CantBeBlocked)
	bear := b12Creature(g, opp.ID, "Bear", rixB1Bear, 2, 2)
	if err := porBlock(t, g, sprinter, opp.ID, bear); err == nil {
		t.Error("Storm Fleet Sprinter was blocked")
	}
}

func TestRix5531B1GiltgroveStalkerEvadesSmallBlockers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stalker := b12Push(g, me.ID, "Giltgrove Stalker", rixB1Merfolk, rixB1GiltgroveStalker, 2, 1)
	small := b12Creature(g, opp.ID, "Small", rixB1Bear, 2, 2)
	if err := porBlock(t, g, stalker, opp.ID, small); err == nil {
		t.Error("a power-2 creature blocked Giltgrove Stalker")
	}
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	stalker = b12Push(g, me.ID, "Giltgrove Stalker", rixB1Merfolk, rixB1GiltgroveStalker, 2, 1)
	big := b12Creature(g, opp.ID, "Big", "Creature — Ogre", 3, 3)
	if err := porBlock(t, g, stalker, opp.ID, big); err != nil {
		t.Errorf("a power-3 creature couldn't block Giltgrove Stalker: %v", err)
	}
}

func TestRix5531B1RiverDarterCantBeBlockedByDinosaurs(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	darter := b12Push(g, me.ID, "River Darter", rixB1Merfolk, rixB1RiverDarter, 2, 3)
	dino := b12Creature(g, opp.ID, "Their Dinosaur", rixB1Dinosaur, 3, 3)
	if err := porBlock(t, g, darter, opp.ID, dino); err == nil {
		t.Error("a Dinosaur blocked River Darter")
	}
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	darter = b12Push(g, me.ID, "River Darter", rixB1Merfolk, rixB1RiverDarter, 2, 3)
	bear := b12Creature(g, opp.ID, "Bear", rixB1Bear, 2, 2)
	if err := porBlock(t, g, darter, opp.ID, bear); err != nil {
		t.Errorf("a non-Dinosaur couldn't block River Darter: %v", err)
	}
}

func TestRix5531B1HardyVeteranIsTougherOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	veteran := b12Push(g, me.ID, "Hardy Veteran", "Creature — Human Warrior", rixB1HardyVeteran, 2, 2)
	if p, tough := ptOf(t, g, veteran); p != 2 || tough != 4 {
		t.Errorf("on my turn %d/%d, want 2/4", p, tough)
	}
	advanceToNextSeatsTurn(t, g)
	if p, tough := ptOf(t, g, veteran); p != 2 || tough != 2 {
		t.Errorf("on their turn %d/%d, want 2/2", p, tough)
	}
}

func TestRix5531B1GraspingScoundrelPumpsOnlyWhileAttacking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	scoundrel := b12Push(g, me.ID, "Grasping Scoundrel", rixB1Pirate, rixB1GraspingScoundrel, 1, 1)
	if p, _ := ptOf(t, g, scoundrel); p != 1 {
		t.Fatalf("power %d at home, want 1", p)
	}
	declareAttack(t, g, opp.ID, scoundrel)
	if p, tough := ptOf(t, g, scoundrel); p != 2 || tough != 1 {
		t.Errorf("attacking it is %d/%d, want 2/1", p, tough)
	}
}

func TestRix5531B1NeckbreakerPumpsOnlyAttackingPirates(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	neck := b12Push(g, me.ID, "Dire Fleet Neckbreaker", "Creature — Orc Pirate", rixB1Neckbreaker, 3, 2)
	pirate := b12Creature(g, me.ID, "My Pirate", rixB1Pirate, 1, 1)
	home := b12Creature(g, me.ID, "Home Pirate", rixB1Pirate, 1, 1)
	bear := b12Creature(g, me.ID, "My Bear", rixB1Bear, 2, 2)
	if p, _ := ptOf(t, g, pirate); p != 1 {
		t.Fatalf("a Pirate at home has power %d, want 1", p)
	}
	declareAttack(t, g, opp.ID, neck, pirate, bear)
	if p, _ := ptOf(t, g, pirate); p != 3 {
		t.Errorf("the attacking Pirate has power %d, want 3", p)
	}
	if p, _ := ptOf(t, g, neck); p != 5 {
		t.Errorf("the attacking Neckbreaker has power %d, want 5", p)
	}
	if p, _ := ptOf(t, g, bear); p != 2 {
		t.Errorf("an attacking non-Pirate has power %d, want 2", p)
	}
	if p, _ := ptOf(t, g, home); p != 1 {
		t.Errorf("a Pirate that didn't attack has power %d, want 1", p)
	}
}

func TestRix5531B1KnightOfTheStampedeDiscountsYourDinosaurs(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Knight of the Stampede", "Creature — Human Knight", rixB1KnightOfStampede, false)
	if got := priceInHand(t, g, me, "Big Dinosaur", rixB1Dinosaur, "{4}{G}"); got != 3 {
		t.Errorf("{4}{G} Dinosaur: %d, want 3", got)
	}
	if got := priceInHand(t, g, me, "Green Dinosaur", rixB1Dinosaur, "{G}{G}"); got != 2 {
		t.Errorf("{G}{G} Dinosaur: %d, want 2 (coloured pips untouched)", got)
	}
	if got := priceInHand(t, g, me, "Grizzly Bears", rixB1Bear, "{2}{G}"); got != 3 {
		t.Errorf("a non-Dinosaur: %d, want 3 (undiscounted)", got)
	}
	if got := priceInHand(t, g, opp, "Their Dinosaur", rixB1Dinosaur, "{4}{G}"); got != 5 {
		t.Errorf("an opponent's Dinosaur: %d, want 5 (undiscounted)", got)
	}
}

func TestRix5531B1RelentlessRaptorMustAttack(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	raptor := b12Push(g, me.ID, "Relentless Raptor", rixB1Dinosaur, rixB1RelentlessRaptor, 3, 3)
	if !hasEffectiveKeyword(t, g, raptor, "vigilance") {
		t.Error("Relentless Raptor has no vigilance")
	}
	advanceToDeclareAttackersOf(t, g, seat)
	if err := g.PassPriority(); !errors.Is(err, game.ErrAttackRequirement) {
		t.Fatalf("pass with the Raptor home = %v, want ErrAttackRequirement", err)
	}
	if err := g.DeclareAttacker(raptor, opp.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the Raptor attacking: %v", err)
	}
}

func TestRix5531B1EverdawnChampionPreventsCombatDamageToItselfOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	attacker := pushVanillaCreature(g, me.ID, "Big Attacker", 5, 5)
	champion := b12Push(g, opp.ID, "Everdawn Champion", "Creature — Human Soldier", rixB1EverdawnChampion, 2, 2)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(champion, attacker); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, champion)
	if !ok {
		t.Fatal("the Champion died to combat damage")
	}
	if c.DamageMarked != 0 {
		t.Errorf("the Champion has %d damage marked", c.DamageMarked)
	}
	b27Damage(g, attacker, champion, 2)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, champion); ok {
		t.Error("2 noncombat damage did not kill the 2/2 Champion")
	}
}

func TestRix5531B1FamishedPaladinUntapsOnYourLifeGain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	paladin := b12Push(g, me.ID, "Famished Paladin", rixB1Vampire, rixB1FamishedPaladin, 3, 3)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(paladin) })
	afrB07Gain(g, opp, 2)
	passPriorityAroundTable(t, g)
	if !isTapped(g, paladin) {
		t.Fatal("an opponent's life gain untapped the Paladin")
	}
	afrB07Gain(g, me, 2)
	passPriorityAroundTable(t, g)
	if isTapped(g, paladin) {
		t.Error("gaining life did not untap the Paladin")
	}
}

func TestRix5531B1DinosaursGainAKeywordWithAnotherDinosaur(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, keyword string
	}{
		{"Stampeding Horncrest", rixB1Horncrest, "haste"},
		{"Sun-Crested Pterodon", rixB1Pterodon, "vigilance"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			id := b12Push(g, me.ID, tc.name, rixB1Dinosaur, tc.oracle, 2, 2)
			if hasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Errorf("%s with no other Dinosaur", tc.keyword)
			}
			b12Creature(g, g.Seats[1].ID, "Their Dinosaur", rixB1Dinosaur, 1, 1)
			if hasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Errorf("%s off an opponent's Dinosaur", tc.keyword)
			}
			b12Creature(g, me.ID, "My Dinosaur", rixB1Dinosaur, 1, 1)
			if !hasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Errorf("no %s with another Dinosaur", tc.keyword)
			}
		})
	}
}

// --- Auras -------------------------------------------------------------------

func TestRix5531B1LuminousBondsStopsAttacksAndBlocks(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	host := pushPermanentForTest(g, me.ID, "Host", "", rixB1Bear)
	aura := castCatalogSpell(t, g, "Luminous Bonds", auraTypeLine, rixB1LuminousBonds, cardTarget(host))
	passPriorityAroundTable(t, g)
	if got := attachmentHostOf(t, g, aura); got.Kind != game.TargetCard || got.ID != host {
		t.Fatalf("Luminous Bonds attached to %+v, want host %s", got, host)
	}
	assertRestrictions(t, g, host, game.CantAttackOrBlock)
}

func TestRix5531B1WaterknotTapsAndHoldsItsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	host := pushPermanentForTest(g, me.ID, "Host", "", rixB1Bear)
	aura := castCatalogSpell(t, g, "Waterknot", auraTypeLine, rixB1Waterknot, cardTarget(host))
	passPriorityAroundTable(t, g)
	if got := attachmentHostOf(t, g, aura); got.Kind != game.TargetCard || got.ID != host {
		t.Fatalf("Waterknot attached to %+v, want host %s", got, host)
	}
	if c, _ := battlefieldCard(g, host); !c.Tapped {
		t.Fatal("Waterknot's entry trigger did not tap its enchanted creature")
	}
	advanceToUpkeepOf(t, g, 0)
	if c, _ := battlefieldCard(g, host); !c.Tapped {
		t.Error("the enchanted creature untapped during its controller's untap step")
	}
}

// --- enrage --------------------------------------------------------------------

func TestRix5531B1ImperialCeratopsGainsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ceratops := b12Push(g, me.ID, "Imperial Ceratops", rixB1Dinosaur, rixB1ImperialCeratops, 3, 5)
	src := b12Creature(g, opp.ID, "Source", "Creature — Ogre", 3, 3)
	life := me.Life
	cmdr2B08Damage(g, src, ceratops, 2)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestRix5531B1OvergrownArmasaurMakesASaproling(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	armasaur := b12Push(g, me.ID, "Overgrown Armasaur", rixB1Dinosaur, rixB1OvergrownArmasaur, 4, 4)
	src := b12Creature(g, opp.ID, "Source", "Creature — Ogre", 3, 3)
	cmdr2B08Damage(g, src, armasaur, 1)
	passPriorityAroundTable(t, g)
	if got := len(battlefieldIDsNamed(g, "Saproling")); got != 1 {
		t.Errorf("%d Saprolings, want 1", got)
	}
}

func TestRix5531B1SiegehornCeratopsGrowsWhenItSurvives(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ceratops := b12Push(g, me.ID, "Siegehorn Ceratops", rixB1Dinosaur, rixB1SiegehornCeratops, 2, 2)
	src := b12Creature(g, opp.ID, "Source", "Creature — Ogre", 1, 1)
	cmdr2B08Damage(g, src, ceratops, 1)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, ceratops, game.CounterPlusOne); n != 2 {
		t.Errorf("%d +1/+1 counters, want 2", n)
	}
}

func TestRix5531B1FrilledDeathspitterHitsTheTargetOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spitter := b12Push(g, me.ID, "Frilled Deathspitter", rixB1Dinosaur, rixB1FrilledSpitter, 3, 2)
	src := b12Creature(g, opp.ID, "Source", "Creature — Ogre", 1, 1)
	life := opp.Life
	cmdr2B08Damage(g, src, spitter, 1)
	rixB1PickAndSettle(t, g, me.ID, playerTarget(opp.ID))
	if opp.Life != life-2 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-2)
	}
}

func TestRix5531B1NeedletoothRaptorDealsFiveToAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	raptor := b12Push(g, me.ID, "Needletooth Raptor", rixB1Dinosaur, rixB1NeedletoothRaptor, 2, 2)
	victim := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 1, 7)
	cmdr2B08Damage(g, victim, raptor, 1)
	rixB1PickAndSettle(t, g, me.ID, rixB1CardRef(victim))
	if got := findCardAnywhere(t, g, victim).DamageMarked; got != 5 {
		t.Errorf("the target has %d damage, want 5", got)
	}
}

func TestRix5531B1CacophodonUntapsThePermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	caco := b12Push(g, me.ID, "Cacophodon", rixB1Dinosaur, rixB1Cacophodon, 2, 5)
	land := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(land) })
	src := b12Creature(g, opp.ID, "Source", "Creature — Ogre", 1, 1)
	cmdr2B08Damage(g, src, caco, 1)
	rixB1PickAndSettle(t, g, me.ID, rixB1CardRef(land))
	if isTapped(g, land) {
		t.Error("the land is still tapped")
	}
}

// --- raid --------------------------------------------------------------------

func TestRix5531B1SwaggeringCorsairEntersWithACounterOnlyAfterAnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	first := avB16CastCreature(t, g, "Swaggering Corsair", rixB1Pirate, rixB1SwaggeringCorsair, 2, 2)
	if n := countersOn(g, first, game.CounterPlusOne); n != 0 {
		t.Fatalf("%d counters without an attack this turn, want 0", n)
	}
	attacker := b12Creature(g, me.ID, "Attacker", rixB1Bear, 2, 2)
	declareAttack(t, g, opp.ID, attacker)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepPostcombatMain)
	second := avB16CastCreature(t, g, "Swaggering Corsair", rixB1Pirate, rixB1SwaggeringCorsair, 2, 2)
	if n := countersOn(g, second, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters after attacking this turn, want 1", n)
	}
}

// --- entry triggers ----------------------------------------------------------

func TestRix5531B1EntryTokens(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		token, keyword         string
	}{
		{"Brazen Freebooter", rixB1Pirate, rixB1BrazenFreebooter, "Treasure", ""},
		{"Crested Herdcaller", rixB1Dinosaur, rixB1CrestedHerdcaller, "Dinosaur", "trample"},
		{"Jungleborn Pioneer", "Creature — Merfolk Scout", rixB1JunglebornPioneer, "Merfolk", "hexproof"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := castHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, 3, 3)
			passPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(id) {
				t.Fatalf("%s did not resolve", tc.name)
			}
			tokens := battlefieldIDsNamed(g, tc.token)
			if len(tokens) != 1 {
				t.Fatalf("%d %s tokens, want 1", len(tokens), tc.token)
			}
			if tc.keyword != "" && !hasEffectiveKeyword(t, g, tokens[0], tc.keyword) {
				t.Errorf("the %s token has no %s", tc.token, tc.keyword)
			}
		})
	}
}

func TestRix5531B1CrestedHerdcallerHasTrample(t *testing.T) {
	g := newCatalogGame(t)
	id := b12Push(g, g.Seats[0].ID, "Crested Herdcaller", rixB1Dinosaur, rixB1CrestedHerdcaller, 3, 3)
	if !hasEffectiveKeyword(t, g, id, "trample") {
		t.Error("Crested Herdcaller has no trample")
	}
}

func TestRix5531B1JadecraftArtisanPumpsTheTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "My Bear", rixB1Bear, 2, 2)
	castHoldCreature(t, g, "Jadecraft Artisan", "Creature — Merfolk Shaman", rixB1JadecraftArtisan, 3, 3)
	rixB1PickAndSettle(t, g, me.ID, rixB1CardRef(bear))
	if p, tough := ptOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4", p, tough)
	}
}

func TestRix5531B1CounterOnAnotherOfYourType(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, kin string
	}{
		{"Jade Bearer", "Creature — Merfolk Shaman", rixB1JadeBearer, rixB1Merfolk},
		{"Sanguine Glorifier", "Creature — Vampire Cleric", rixB1SanguineGlorifier, rixB1Vampire},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			kin := b12Creature(g, me.ID, "My Kin", tc.kin, 1, 1)
			id := castHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, 1, 1)
			rixB1PickAndSettle(t, g, me.ID, rixB1CardRef(kin))
			if n := countersOn(g, kin, game.CounterPlusOne); n != 1 {
				t.Errorf("the other creature has %d counters, want 1", n)
			}
			if n := countersOn(g, id, game.CounterPlusOne); n != 0 {
				t.Errorf("%s put %d counters on itself", tc.name, n)
			}
		})
	}
}

func TestRix5531B1VoraciousVampirePumpsAndGrantsMenace(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	kin := b12Creature(g, me.ID, "My Vampire", rixB1Vampire, 2, 2)
	id := castHoldCreature(t, g, "Voracious Vampire", rixB1Vampire, rixB1VoraciousVampire, 2, 2)
	rixB1PickAndSettle(t, g, me.ID, rixB1CardRef(kin))
	if !hasEffectiveKeyword(t, g, id, "menace") {
		t.Error("Voracious Vampire has no menace")
	}
	if p, tough := ptOf(t, g, kin); p != 3 || tough != 3 {
		t.Errorf("the Vampire is %d/%d, want 3/3", p, tough)
	}
	if !hasEffectiveKeyword(t, g, kin, "menace") {
		t.Error("the target Vampire did not gain menace")
	}
}

func TestRix5531B1SwiftWardenGrantsHexproof(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	kin := b12Creature(g, me.ID, "My Merfolk", rixB1Merfolk, 2, 2)
	id := castHoldCreature(t, g, "Swift Warden", rixB1Merfolk, rixB1SwiftWarden, 3, 3)
	rixB1PickAndSettle(t, g, me.ID, rixB1CardRef(kin))
	if !hasEffectiveKeyword(t, g, id, "flash") {
		t.Error("Swift Warden has no flash")
	}
	if !hasEffectiveKeyword(t, g, kin, "hexproof") {
		t.Error("the target Merfolk did not gain hexproof")
	}
}

// --- activated abilities -------------------------------------------------------

func TestRix5531B1TombRobberDiscardsToExplore(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	robber := b12Push(g, me.ID, "Tomb Robber", rixB1Pirate, rixB1TombRobber, 1, 1)
	if !hasEffectiveKeyword(t, g, robber, "menace") {
		t.Error("Tomb Robber has no menace")
	}
	d2dSetHand(me, 1)
	discard := me.Hand.Cards[0].InstanceID
	top := putOnTopForTest(me, "Forest", "Basic Land — Forest")
	apaMana(me, "C")
	apaActivate(t, g, me, robber, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{discard}})
	if !me.Graveyard.Contains(discard) {
		t.Error("the card was not discarded")
	}
	if !me.Hand.Contains(top) {
		t.Error("the revealed land did not go to the hand")
	}
}

func TestRix5531B1SunCollaredRaptorPumpsThreePower(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	raptor := b12Push(g, me.ID, "Sun-Collared Raptor", rixB1Dinosaur, rixB1SunCollaredRaptor, 1, 2)
	if !hasEffectiveKeyword(t, g, raptor, "trample") {
		t.Error("Sun-Collared Raptor has no trample")
	}
	apaMana(me, "R", "C", "C")
	apaActivate(t, g, me, raptor, 0, game.ActivateAbilityParams{})
	if p, tough := ptOf(t, g, raptor); p != 4 || tough != 2 {
		t.Errorf("the Raptor is %d/%d, want 4/2", p, tough)
	}
}

// --- spells --------------------------------------------------------------------

func TestRix5531B1ImpaleDestroysACreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	victim := b12Creature(g, opp.ID, "Victim", rixB1Bear, 2, 2)
	castCatalogSpell(t, g, "Impale", "Sorcery", rixB1Impale, cardTarget(victim))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the creature survived Impale")
	}
}

func TestRix5531B1ReaverAmbushExilesOnlyPowerThreeOrLess(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	big := b12Creature(g, opp.ID, "Big", "Creature — Ogre", 4, 4)
	small := b12Creature(g, opp.ID, "Small", "Creature — Ogre", 3, 3)
	if err := castCatalogSpellErr(t, g, "Reaver Ambush", "Instant", rixB1ReaverAmbush, cardTarget(big)); err == nil {
		t.Fatal("Reaver Ambush targeted a power-4 creature")
	}
	castCatalogSpell(t, g, "Reaver Ambush", "Instant", rixB1ReaverAmbush, cardTarget(small))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) || !g.Exile.Contains(small) {
		t.Error("the power-3 creature was not exiled")
	}
}

func TestRix5531B1HornswoggleCountersACreatureSpellAndMakesATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b13OpponentCasts(t, g, opp, "Their Bear", rixB1Bear, "", "{1}{G}", nil)
	id := b13HandCard(me, "Hornswoggle", "Instant", rixB1Hornswoggle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: b16TargetCard(bear)}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bear) {
		t.Error("the creature spell was not countered")
	}
	treasures := battlefieldIDsNamed(g, "Treasure")
	if len(treasures) != 1 {
		t.Fatalf("%d Treasures, want 1", len(treasures))
	}
	if c, _ := battlefieldCard(g, treasures[0]); c.Controller != me.ID {
		t.Error("the Treasure is not the caster's")
	}
}

func TestRix5531B1HornswoggleTargetsOnlyACreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := b13OpponentCasts(t, g, opp, "Their Bolt", "Instant", "", "{R}", nil)
	id := b13HandCard(me, "Hornswoggle", "Instant", rixB1Hornswoggle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: b16TargetCard(bolt)}); err == nil {
		t.Error("Hornswoggle targeted an instant spell")
	}
}

func TestRix5531B1StrengthOfThePackCountersOnlyYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	a := b12Creature(g, me.ID, "A", rixB1Bear, 2, 2)
	b := b12Creature(g, me.ID, "B", rixB1Bear, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", rixB1Bear, 2, 2)
	castCatalogSpell(t, g, "Strength of the Pack", "Sorcery", rixB1StrengthOfThePack, nil)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if n := countersOn(g, id, game.CounterPlusOne); n != 2 {
			t.Errorf("my creature has %d counters, want 2", n)
		}
	}
	if n := countersOn(g, theirs, game.CounterPlusOne); n != 0 {
		t.Errorf("an opponent's creature has %d counters, want 0", n)
	}
}

func TestRix5531B1ShakeTheFoundationsSparesFlyersAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := b12Creature(g, opp.ID, "Walker", "Creature — Human", 1, 1)
	flyer := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Shake the Foundations", "Instant", rixB1ShakeFoundations, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(walker) {
		t.Error("the creature without flying survived")
	}
	if !g.Battlefield.Contains(flyer) {
		t.Error("the flyer was dealt damage")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
}

func TestRix5531B1EnterTheUnknownExploresAndGrantsALandDrop(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	scout := b12Creature(g, me.ID, "Scout", rixB1Merfolk, 1, 1)
	top := putOnTopForTest(me, "Forest", "Basic Land — Forest")
	castCatalogSpell(t, g, "Enter the Unknown", "Sorcery", rixB1EnterTheUnknown, cardTarget(scout))
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(top) {
		t.Error("the revealed land did not go to the hand")
	}
	if got := g.ExtraLandDropsThisTurn[me.ID]; got != 1 {
		t.Errorf("extra land plays = %d, want 1", got)
	}
}

func TestRix5531B1CleansingRayEachMode(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	vampire := b12Creature(g, opp.ID, "Their Vampire", rixB1Vampire, 2, 2)
	bear := b12Creature(g, opp.ID, "Their Bear", rixB1Bear, 2, 2)
	ench := b12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	if err := rixB1CastWithModes(t, g, "Cleansing Ray", "Sorcery", rixB1CleansingRay, []int{0}, cardTarget(bear)); err == nil {
		t.Fatal("the Vampire bullet targeted a Bear")
	}
	castModal(t, g, "Cleansing Ray", "Sorcery", rixB1CleansingRay, []int{0}, cardTarget(vampire))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(vampire) {
		t.Error("the Vampire survived")
	}
	castModal(t, g, "Cleansing Ray", "Sorcery", rixB1CleansingRay, []int{1}, cardTarget(ench))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) {
		t.Error("the enchantment survived")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the Bear was destroyed")
	}
}

func TestRix5531B1CrashingTideHasFlashOnlyWithAMerfolk(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := b12Creature(g, opp.ID, "Victim", rixB1Bear, 2, 2)
	tide := b13HandCard(me, "Crashing Tide", "Sorcery", rixB1CrashingTide)
	advanceTo(t, g, game.StepBeginCombat)
	if err := g.CastSpell(me.ID, tide, game.CastSpellParams{Strict: true, Targets: cardTarget(victim)}); err == nil {
		t.Fatal("cast at instant speed with no Merfolk")
	}
	b12Creature(g, me.ID, "My Merfolk", rixB1Merfolk, 1, 1)
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, tide, game.CastSpellParams{Strict: true, Targets: cardTarget(victim)}); err != nil {
		t.Fatalf("flash with a Merfolk: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) || !opp.Hand.Contains(victim) {
		t.Error("the creature did not return to its owner's hand")
	}
	// The Tide left the hand and a card was drawn.
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand is %d, want %d", got, hand)
	}
}
