package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// b07BoltPlayer casts Lightning Bolt at a player and settles it.
func b07BoltPlayer(t *testing.T, g *game.Game, target uuid.UUID) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: target}})
	effectstest.PassPriorityAroundTable(t, g)
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. The Tron
// lands are a loop over a table, so a transposed row is invisible
// until someone plays that exact card.
func TestBatch07CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B07TwinflameTyrantOracle:       "Twinflame Tyrant",
		effectstest.B07SteelOverseerOracle:         "Steel Overseer",
		effectstest.B07MesaEnchantressOracle:       "Mesa Enchantress",
		effectstest.B07SaiMasterThopteristOracle:   "Sai, Master Thopterist",
		effectstest.B07GhostQuarterOracle:          "Ghost Quarter",
		effectstest.B07YoungPyromancerOracle:       "Young Pyromancer",
		effectstest.B07SoulsAttendantOracle:        "Soul's Attendant",
		effectstest.B07CurseOfTheSwineOracle:       "Curse of the Swine",
		effectstest.B07UrzasTowerOracle:            "Urza's Tower",
		effectstest.B07KodamaOfTheWestTreeOracle:   "Kodama of the West Tree",
		effectstest.B07MurderOracle:                "Murder",
		effectstest.B07NadiersNightbladeOracle:     "Nadier's Nightblade",
		effectstest.B07BloodthirstyConquerorOracle: "Bloodthirsty Conqueror",
		effectstest.B07AyaraOracle:                 "Ayara, First of Locthwain",
		effectstest.B07UrzasMineOracle:             "Urza's Mine",
		effectstest.B07FieryEmancipationOracle:     "Fiery Emancipation",
		effectstest.B07UrzasPowerPlantOracle:       "Urza's Power Plant",
		effectstest.B07FieryIsletOracle:            "Fiery Islet",
		effectstest.B07EleshNornOracle:             "Elesh Norn, Grand Cenobite",
		effectstest.B07TemptWithDiscoveryOracle:    "Tempt with Discovery",
		effectstest.B07CityOnFireOracle:            "City on Fire",
		effectstest.B07MysticGateOracle:            "Mystic Gate",
		effectstest.B07SplendidReclamationOracle:   "Splendid Reclamation",
		effectstest.B07MyrRetrieverOracle:          "Myr Retriever",
		effectstest.B07TemurAscendancyOracle:       "Temur Ascendancy",
		effectstest.B07SerumVisionsOracle:          "Serum Visions",
		effectstest.B07EidolonOfBlossomsOracle:     "Eidolon of Blossoms",
		effectstest.B07FieldOfRuinOracle:           "Field of Ruin",
	}
	if len(want) != 28 {
		t.Fatalf("the batch is 28 cards, the table lists %d", len(want))
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
		if spec.Completeness == effects.CompletenessUnreviewed {
			t.Errorf("%s ships without a completeness declaration", name)
		}
	}
}

// --- damage multipliers --------------------------------------------

func TestB07TwinflameTyrantDoublesDamageToOpponentsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Twinflame Tyrant", "Creature — Dragon", effectstest.B07TwinflameTyrantOracle, false)
	oppBefore, meBefore := opp.Life, me.Life

	b07BoltPlayer(t, g, opp.ID)
	if opp.Life != oppBefore-6 {
		t.Errorf("a Bolt at an opponent: %d → %d, want -6", oppBefore, opp.Life)
	}
	// Damage to yourself is not "to an opponent".
	b07BoltPlayer(t, g, me.ID)
	if me.Life != meBefore-3 {
		t.Errorf("a Bolt at yourself: %d → %d, want -3 (not doubled)", meBefore, me.Life)
	}
}

func TestB07FieryEmancipationTriplesEverything(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Fiery Emancipation", "Enchantment", effectstest.B07FieryEmancipationOracle, false)
	oppBefore, meBefore := opp.Life, me.Life

	b07BoltPlayer(t, g, opp.ID)
	if opp.Life != oppBefore-9 {
		t.Errorf("a Bolt at an opponent: %d → %d, want -9", oppBefore, opp.Life)
	}
	// Your own sources hurting you are tripled too, as printed.
	b07BoltPlayer(t, g, me.ID)
	if me.Life != meBefore-9 {
		t.Errorf("a Bolt at yourself: %d → %d, want -9", meBefore, me.Life)
	}
}

func TestB07CityOnFireConvokesAndTriples(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.B07CityOnFireOracle)
	if !ok || spec.TapCost == nil {
		t.Fatal("City on Fire must declare convoke as its tap cost")
	}
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "City on Fire", "Enchantment", effectstest.B07CityOnFireOracle, false)
	before := opp.Life
	b07BoltPlayer(t, g, opp.ID)
	if opp.Life != before-9 {
		t.Errorf("a Bolt under City on Fire: %d → %d, want -9", before, opp.Life)
	}
}

// --- cast triggers -------------------------------------------------

func TestB07MesaEnchantressMayDrawOnAnEnchantmentSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Mesa Enchantress", "Creature — Human Druid", effectstest.B07MesaEnchantressOracle, false)
	before := me.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Glory", "Enchantment", "", nil)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	// castCatalogSpell adds the spell to the hand and casting removes
	// it, so the difference is exactly the cards drawn.
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}

	// A creature spell is not an enchantment spell: no prompt.
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	if len(g.PendingChoices) != 0 {
		t.Error("a creature spell prompted the Enchantress")
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestB07SaiMakesAThopterPerArtifactSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Sai, Master Thopterist", "Legendary Creature — Human Artificer", effectstest.B07SaiMasterThopteristOracle, false)

	effectstest.CastCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldNamed(g, me.ID, "Thopter") != 1 {
		t.Fatal("an artifact spell should make one Thopter")
	}
	thopter := effectstest.FindBattlefieldByName(g, "Thopter")
	if card, _ := effectstest.BattlefieldCard(g, thopter); !card.IsArtifact() || !card.IsCreature() {
		t.Error("the Thopter is an artifact creature")
	}
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, thopter), "flying") {
		t.Error("the Thopter flies")
	}

	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldNamed(g, me.ID, "Thopter") != 1 {
		t.Error("a creature spell is not an artifact spell")
	}
}

func TestB07YoungPyromancerMakesAnElementalPerInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Young Pyromancer", "Creature — Human Shaman", effectstest.B07YoungPyromancerOracle, false)

	b07BoltPlayer(t, g, opp.ID)
	effectstest.CastCatalogSpell(t, g, "Ponder", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Elemental"); n != 2 {
		t.Fatalf("an instant and a sorcery made %d Elementals, want 2", n)
	}
	elemental := effectstest.FindBattlefieldByName(g, "Elemental")
	if card, _ := effectstest.BattlefieldCard(g, elemental); !card.HasColor("R") {
		t.Error("the Elemental is red")
	}

	effectstest.CastCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldNamed(g, me.ID, "Elemental") != 2 {
		t.Error("an artifact spell triggered the Pyromancer")
	}
}

// --- ETB triggers --------------------------------------------------

func TestB07SoulsAttendantMayGainOnAnyOtherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := me.Life
	effectstest.CastCatalogSpell(t, g, "Soul's Attendant", "Creature — Human Cleric", effectstest.B07SoulsAttendantOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 {
		t.Fatal("the Attendant's own entry is not 'another creature'")
	}

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.RedGoblinToken(), 1) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d → %d, want +1 for an opponent's creature", before, me.Life)
	}

	// Declining is a real answer.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 1) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d → %d after declining, want still +1", before, me.Life)
	}
}

func TestB07EidolonOfBlossomsDrawsForItselfAndOtherEnchantments(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := me.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Eidolon of Blossoms", "Enchantment Creature — Spirit", effectstest.B07EidolonOfBlossomsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// castCatalogSpell adds the spell to the hand and casting removes
	// it, so the difference is exactly the cards drawn.
	if got := me.Hand.Size() - before; got != 1 {
		t.Fatalf("drew %d, want 1 — the Eidolon's own entry draws (constellation counts itself)", got)
	}

	effectstest.CastCatalogSpell(t, g, "Glory", "Enchantment", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 2 {
		t.Errorf("drew %d after another enchantment, want 2", got)
	}

	// A creature that is not an enchantment, and an opponent's
	// enchantment, are both silent.
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.SeedPermanentFor(g, opp.ID, "Their Glory", "Enchantment")
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 2 {
		t.Errorf("drew %d, want still 2", got)
	}
}

func TestB07TemurAscendancyGrantsHasteAndMayDrawForPowerFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Temur Ascendancy", "Enchantment", effectstest.B07TemurAscendancyOracle, false)
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, bear), "haste") {
		t.Error("creatures you control have haste")
	}
	before := me.Hand.Size()

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("3/3 green Beast"), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 {
		t.Fatal("a 3/3 entering must not prompt")
	}
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, game.Card{Name: "Wurm", TypeLine: "Token Creature — Wurm", Power: 5, Toughness: 5}, 1)
	})
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d, want 1 for the 5/5", got)
	}
}

func TestB07AyaraDrainsOnBlackCreaturesAndEatsThemForCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := effectstest.LifeOfOpponents(g)
	meBefore := me.Life

	ayara := effectstest.CastCatalogSpell(t, g, "Ayara, First of Locthwain", "Legendary Creature — Elf Noble", effectstest.B07AyaraOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("opponent %d after Ayara's own entry: %d → %d, want -1", i+1, b, got)
		}
	}
	if me.Life != meBefore+1 {
		t.Errorf("Ayara's controller %d → %d, want +1", meBefore, me.Life)
	}

	// A black creature drains again; a green one does not.
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, effects.BlackZombieToken(), 1)
		_ = g.CreateTokenForEffect(me.ID, effects.TokenCard("3/3 green Beast"), 1)
	})
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-2 {
			t.Errorf("opponent %d after a Zombie and a Beast: %d → %d, want -2", i+1, b, got)
		}
	}

	// The sacrifice cost takes another black creature, not a green
	// one and not Ayara herself. She was cast this turn, so shake
	// off the summoning sickness for the tap cost.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == ayara {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	zombie := effectstest.FindBattlefieldByName(g, "Zombie")
	beast := effectstest.FindBattlefieldByName(g, "Beast")
	for _, bad := range []uuid.UUID{beast, ayara} {
		if err := g.ActivateCatalogAbility(me.ID, ayara, 0, game.ActivateAbilityParams{
			SacrificeIDs: []uuid.UUID{bad},
		}); err == nil {
			t.Errorf("%s was accepted for 'sacrifice another black creature'", bad)
		}
	}
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, ayara, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{zombie},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if card, _ := effectstest.BattlefieldCard(g, ayara); !card.Tapped {
		t.Error("the ability has a tap cost")
	}
}

// --- life-loss and leaves-the-battlefield triggers -----------------

func TestB07BloodthirstyConquerorGainsWhatOpponentsLose(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	conq := effectstest.PushCatalogPermanent(g, me.ID, "Bloodthirsty Conqueror", "Creature — Vampire Knight", effectstest.B07BloodthirstyConquerorOracle, false)
	if abs := effectstest.EffectiveAbilities(t, g, conq); !effects.EotHasAbility(abs, "flying") || !effects.EotHasAbility(abs, "deathtouch") {
		t.Error("printed flying and deathtouch did not reach the effective abilities")
	}
	before := me.Life

	b07BoltPlayer(t, g, opp.ID)
	if me.Life != before+3 {
		t.Fatalf("after a Bolt to an opponent: %d → %d, want +3", before, me.Life)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -2) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+5 {
		t.Errorf("after a 2-life drain: %d → %d, want +5", before, me.Life)
	}
	// Your own loss is not an opponent's.
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -1) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+4 {
		t.Errorf("after my own loss: %d → %d, want +4", before, me.Life)
	}
}

func TestB07NadiersNightbladeDrainsWhenATokenLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Nadier's Nightblade", "Creature — Elf Warrior", effectstest.B07NadiersNightbladeOracle, false)
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 1)
		_ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 1)
		_ = g.CreateTokenForEffect(opp.ID, effects.RedGoblinToken(), 1)
	})
	effectstest.PassPriorityAroundTable(t, g)
	before := effectstest.LifeOfOpponents(g)
	meBefore := me.Life

	// Sacrificed and bounced both count: "leaves the battlefield".
	treasure := effectstest.FindBattlefieldByName(g, "Treasure")
	goblin := uuid.Nil
	theirGoblin := uuid.Nil
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Goblin" {
			if c.Controller == me.ID {
				goblin = c.InstanceID
			} else {
				theirGoblin = c.InstanceID
			}
		}
	}
	g.WithWriteLock(func() {
		_ = g.SacrificePermanentForEffect(treasure)
		_ = g.BounceToHandForEffect(goblin)
	})
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-2 {
			t.Errorf("opponent %d: %d → %d, want -2 (two tokens left)", i+1, b, got)
		}
	}
	if me.Life != meBefore+2 {
		t.Errorf("controller %d → %d, want +2", meBefore, me.Life)
	}

	// An opponent's token and a nontoken creature are silent.
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	g.WithWriteLock(func() {
		_ = g.SacrificePermanentForEffect(theirGoblin)
		_ = g.SacrificePermanentForEffect(bear)
	})
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-2 {
			t.Errorf("opponent %d: %d → %d, want still -2", i+1, b, got)
		}
	}
}

func TestB07MyrRetrieverReturnsAnotherArtifactCardWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	retriever := effectstest.PushCatalogPermanent(g, me.ID, "Myr Retriever", "Artifact Creature — Myr", effectstest.B07MyrRetrieverOracle, false)
	relic := effectstest.Batch01GraveyardCard(me, "Sol Ring", "Artifact")
	bear := effectstest.PushGraveyardCardForTest(me, "Dead Bear")

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(retriever) })
	effectstest.B04WaitForPick(t, g, me.ID)
	pick := effectstest.LatestPickTarget(g, me.ID)
	for _, id := range pick.PickTargetCards {
		if id == bear {
			t.Error("a creature card was offered for 'target artifact card'")
		}
	}
	effectstest.PickCard(t, g, me.ID, relic)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(relic) {
		t.Error("the artifact card did not come back to hand")
	}
}

// --- statics -------------------------------------------------------

func TestB07EleshNornPumpsYoursAndShrinksTheirs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.SeedCreature(g, "Mine", me.ID)
	small := effectstest.SeedCreature(g, "Their Small", opp.ID)
	big := effectstest.B04Creature(g, opp.ID, "Their Big", 5, 5)

	norn := effectstest.CastCatalogSpell(t, g, "Elesh Norn, Grand Cenobite", "Legendary Creature — Phyrexian Praetor", effectstest.B07EleshNornOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if p, tough := effectstest.EffectivePower(t, g, mine), effectstest.EffectiveToughness(t, g, mine); p != 4 || tough != 4 {
		t.Errorf("my 2/2 = %d/%d, want 4/4", p, tough)
	}
	if p := effectstest.EffectivePower(t, g, norn); p != 0 {
		t.Errorf("Norn pumps OTHER creatures, her own power = %d, want the printed 0 of the fixture", p)
	}
	if g.Battlefield.Contains(small) {
		t.Error("an opponent's 2/2 dies to -2/-2")
	}
	if p, tough := effectstest.EffectivePower(t, g, big), effectstest.EffectiveToughness(t, g, big); p != 3 || tough != 3 {
		t.Errorf("their 5/5 = %d/%d, want 3/3", p, tough)
	}
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, norn), "vigilance") {
		t.Error("printed vigilance did not reach the effective abilities")
	}
}

func TestB07KodamaGivesModifiedCreaturesTrampleAndRampsOnDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Kodama of the West Tree", "Legendary Creature — Spirit", effectstest.B07KodamaOfTheWestTreeOracle, false)
	plain := effectstest.PushVanillaCreature(g, me.ID, "Plain", 2, 2)
	countered := effectstest.PushVanillaCreature(g, me.ID, "Countered", 2, 2)
	equipped := effectstest.PushVanillaCreature(g, me.ID, "Equipped", 2, 2)
	theirs := effectstest.SeedCreature(g, "Theirs", victim.ID)
	blade := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Blade", TypeLine: "Artifact — Equipment",
		Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(countered, "+1/+1", 1)
		_ = g.AddCounterForEffect(theirs, "+1/+1", 1)
		if err := g.AttachForEffect(blade, game.TargetRef{Kind: game.TargetCard, ID: equipped}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
	})

	for _, tc := range []struct {
		id   uuid.UUID
		want bool
		why  string
	}{
		{countered, true, "a creature with a counter is modified"},
		{equipped, true, "an equipped creature is modified"},
		{plain, false, "an unmodified creature gets nothing"},
		{theirs, false, "an opponent's modified creature is not 'you control'"},
	} {
		if got := effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, tc.id), "trample"); got != tc.want {
			t.Errorf("%s: trample = %v, want %v", tc.why, got, tc.want)
		}
	}

	forest := effectstest.B07LibraryLand(me, "Forest", "Basic Land — Forest")
	effectstest.AttackWith(t, g, victim.ID, countered, plain)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	card, ok := effectstest.BattlefieldCard(g, forest)
	if !ok {
		t.Fatal("one modified creature connecting should fetch one basic")
	}
	if !card.Tapped {
		t.Error("the basic enters tapped")
	}
}

// --- activated abilities -------------------------------------------

func TestB07SteelOverseerGrowsEveryArtifactCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	overseer := effectstest.PushCatalogPermanent(g, me.ID, "Steel Overseer", "Artifact Creature — Construct", effectstest.B07SteelOverseerOracle, false)
	myr := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Myr", TypeLine: "Artifact Creature — Myr",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Myr", TypeLine: "Artifact Creature — Myr",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})

	if err := g.ActivateCatalogAbility(me.ID, overseer, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, tc := range []struct {
		id   uuid.UUID
		want int
	}{{overseer, 1}, {myr, 1}, {bear, 0}, {theirs, 0}} {
		if card, _ := effectstest.BattlefieldCard(g, tc.id); card.Counters["+1/+1"] != tc.want {
			t.Errorf("%s: +1/+1 counters = %d, want %d", card.Name, card.Counters["+1/+1"], tc.want)
		}
	}
}

func TestB07SteelOverseerRespectsSummoningSickness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	overseer := effectstest.PushCatalogPermanent(g, me.ID, "Steel Overseer", "Artifact Creature — Construct", effectstest.B07SteelOverseerOracle, true)
	if err := g.ActivateCatalogAbility(me.ID, overseer, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("a summoning-sick Overseer must not tap")
	}
}

func TestB07FieryIsletPaysLifeForColourAndCashesInForACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	islet := effectstest.SeedPermanentWithOracle(g, me.ID, "Fiery Islet", "Land", effectstest.B07FieryIsletOracle)
	before := me.Life
	if err := g.ActivateManaAbility(me.ID, islet, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.Life != before-1 {
		t.Errorf("life %d → %d, want -1", before, me.Life)
	}
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 2 {
		t.Fatalf("the Islet asks U or R, got %+v", pick)
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "U"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}

	fresh := effectstest.SeedPermanentWithOracle(g, me.ID, "Fiery Islet", "Land", effectstest.B07FieryIsletOracle)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, fresh, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(fresh) {
		t.Error("the Islet is sacrificed as a cost, at announce")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func TestB07GhostQuarterDestroysALandAndOffersItsControllerABasic(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	quarter := effectstest.PushCatalogPermanent(g, me.ID, "Ghost Quarter", "Land", effectstest.B07GhostQuarterOracle, false)
	coffers := effectstest.SeedLandOnBattlefield(g, opp.ID, "Cabal Coffers", "Land")
	swamp := effectstest.B07LibraryLand(opp, "Swamp", "Basic Land — Swamp")
	effectstest.B07LibraryLand(opp, "Swamp", "Basic Land — Swamp")

	if err := g.ActivateCatalogAbility(me.ID, quarter, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: coffers}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(quarter) {
		t.Error("the Quarter is sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(coffers) || !opp.Graveyard.Contains(coffers) {
		t.Error("the targeted land should be destroyed")
	}
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Error("the activator does not search — the land's controller does")
	}
	effectstest.AnswerSearchByID(t, g, opp.ID, swamp)
	card, ok := effectstest.BattlefieldCard(g, swamp)
	if !ok {
		t.Fatal("the victim's basic did not reach the battlefield")
	}
	if card.Tapped {
		t.Error("the basic enters untapped, as printed")
	}
}

func TestB07GhostQuarterOnYourOwnLandIsAFetch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	quarter := effectstest.PushCatalogPermanent(g, me.ID, "Ghost Quarter", "Land", effectstest.B07GhostQuarterOracle, false)
	mine := effectstest.SeedLandOnBattlefield(g, me.ID, "Wastes", "Basic Land")
	forest := effectstest.B07LibraryLand(me, "Forest", "Basic Land — Forest")
	effectstest.B07LibraryLand(me, "Forest", "Basic Land — Forest")

	if err := g.ActivateCatalogAbility(me.ID, quarter, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, forest)
	if !g.Battlefield.Contains(forest) {
		t.Error("targeting your own land should let you fetch")
	}
}

func TestB07FieldOfRuinDestroysANonbasicAndEveryoneSearches(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	field := effectstest.PushCatalogPermanent(g, me.ID, "Field of Ruin", "Land", effectstest.B07FieldOfRuinOracle, false)
	coffers := effectstest.SeedLandOnBattlefield(g, opp.ID, "Cabal Coffers", "Land")
	basic := effectstest.SeedLandOnBattlefield(g, opp.ID, "Swamp", "Basic Land — Swamp")
	// Seat 0 and seat 1 hold two basics each, seat 2 holds one, seat 3
	// holds none. A search for a basic land has a stated quality, so the
	// player with exactly one basic may still fail to find (CR 701.23b,
	// #3053): three prompts with a card to take. Seat 3 holds none and is
	// asked all the same (#4140): a library with nothing to find looks
	// like any other to the table, so a fourth prompt, with no cards.
	wanted := map[uuid.UUID]uuid.UUID{}
	for _, p := range []*game.Player{me, opp} {
		wanted[p.ID] = effectstest.B07LibraryLand(p, "Plains", "Basic Land — Plains")
		effectstest.B07LibraryLand(p, "Plains", "Basic Land — Plains")
	}
	only := effectstest.B07LibraryLand(g.Seats[2], "Island", "Basic Land — Island")
	wanted[g.Seats[2].ID] = only
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})

	if err := g.ActivateCatalogAbility(me.ID, field, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: basic}},
	}); err == nil {
		t.Fatal("a basic land was accepted for 'target nonbasic land'")
	}
	if err := g.ActivateCatalogAbility(me.ID, field, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: coffers}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(coffers) {
		t.Error("the targeted nonbasic should be destroyed")
	}
	if g.Battlefield.Contains(only) {
		t.Error("a player with exactly one basic is still asked: they may find nothing")
	}
	if effectstest.B07OpenSearchPrompts(g) != 4 {
		t.Fatalf("want a prompt for each of the four players, three with a basic to find, got %d", effectstest.B07OpenSearchPrompts(g))
	}
	effectstest.AnswerSearchFailToFind(t, g, g.Seats[3].ID)
	for _, p := range []*game.Player{me, opp, g.Seats[2]} {
		effectstest.AnswerSearchByID(t, g, p.ID, wanted[p.ID])
		if card, ok := effectstest.BattlefieldCard(g, wanted[p.ID]); !ok || card.Tapped {
			t.Errorf("%s's basic should be on the battlefield untapped", p.Name)
		}
	}
}

// --- spells --------------------------------------------------------

func TestB07MurderDestroysACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.SeedCreature(g, "Bear", opp.ID)
	effectstest.CastCatalogSpell(t, g, "Murder", "Instant", effectstest.B07MurderOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Graveyard.Contains(bear) {
		t.Error("the creature should be in its owner's graveyard")
	}
}

func TestB07CurseOfTheSwineExilesXAndHandsOutBoars(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.SeedCreature(g, "Mine", me.ID)
	theirs := effectstest.SeedCreature(g, "Theirs", opp.ID)
	spared := effectstest.SeedCreature(g, "Spared", opp.ID)

	effectstest.CastXSpell(t, g, "Curse of the Swine", "Sorcery", effectstest.B07CurseOfTheSwineOracle, "{X}{U}{U}", 2,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{mine, theirs} {
		if !g.Exile.Contains(id) {
			t.Errorf("%s should be in exile", id)
		}
	}
	if !g.Battlefield.Contains(spared) {
		t.Error("an untargeted creature survives")
	}
	if effectstest.CountBattlefieldNamed(g, me.ID, "Boar") != 1 || effectstest.CountBattlefieldNamed(g, opp.ID, "Boar") != 1 {
		t.Error("each exiled creature's controller gets one Boar")
	}
}

func TestB07CurseOfTheSwineDemandsExactlyXTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.SeedCreature(g, "Theirs", opp.ID)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Curse of the Swine", TypeLine: "Sorcery",
		OracleID: effectstest.B07CurseOfTheSwineOracle, ManaCost: "{X}{U}{U}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		XValue:  2,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	}); err == nil {
		t.Error("one target was accepted for X=2")
	}
}

func TestB07SerumVisionsDrawsThenScriesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Serum Visions", "Sorcery", effectstest.B07SerumVisionsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	scry := effectstest.ScryChoiceFor(g, me.ID)
	if scry == nil || len(scry.ScryCards) != 2 {
		t.Fatalf("want a scry prompt over two cards, got %+v", scry)
	}
}

// TestB07SplendidReclamationReturnsEveryLandCardTapped is #1654's
// closing half of #1284: the lands return via
// ReturnFromGraveyard.Tapped, not an untapped return followed by an
// OnETB tap. The assertion that earns its keep is the tap EVENT
// count, not the Tapped flag — the same discipline temples_test.go
// and reassembling_skeleton_test.go use for their own self-entry
// replacements: an enters-then-gets-tapped workaround would leave the
// permanent Tapped too, so only counting EventTapCard tells the two
// apart.
func TestB07SplendidReclamationReturnsEveryLandCardTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forest := effectstest.Batch01GraveyardCard(me, "Forest", "Basic Land — Forest")
	coffers := effectstest.Batch01GraveyardCard(me, "Cabal Coffers", "Land")
	bear := effectstest.PushGraveyardCardForTest(me, "Dead Bear")
	theirs := effectstest.Batch01GraveyardCard(opp, "Their Forest", "Basic Land — Forest")

	effectstest.CastCatalogSpell(t, g, "Splendid Reclamation", "Sorcery", effectstest.B07SplendidReclamationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{forest, coffers} {
		card, ok := effectstest.BattlefieldCard(g, id)
		if !ok {
			t.Errorf("%s did not return", id)
			continue
		}
		if !card.Tapped {
			t.Errorf("%s returned untapped", card.Name)
		}
		if card.Controller != me.ID {
			t.Errorf("%s returned under %s, want the caster", card.Name, card.Controller)
		}
		if n := effectstest.TapEventsFor(g, id); n != 0 {
			t.Errorf("%s entered via a replacement, not an OnETB tap — got %d EventTapCard, want 0", card.Name, n)
		}
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("a creature card stays in the graveyard")
	}
	if !opp.Graveyard.Contains(theirs) {
		t.Error("an opponent's graveyard is not 'your graveyard'")
	}
}

func TestB07TemptWithDiscoveryOffersEachOpponentInTurnAndPaysTheCasterBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp1, opp2, opp3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	// Every library holds two lands, so every search is a real
	// prompt; the caster's has four, one for each search it can earn.
	for _, p := range []*game.Player{me, opp1, opp2, opp3} {
		for i := 0; i < 2; i++ {
			effectstest.B07LibraryLand(p, "Forest", "Basic Land — Forest")
		}
	}
	effectstest.B07LibraryLand(me, "Cabal Coffers", "Land")
	effectstest.B07LibraryLand(me, "Gaea's Cradle", "Legendary Land")
	effectstest.B07LibraryLand(me, "Bear", "Creature — Bear")

	effectstest.CastCatalogSpell(t, g, "Tempt with Discovery", "Sorcery", effectstest.B07TemptWithDiscoveryOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	// Step 1: the caster's own search, and nothing else is open yet.
	step := func(who *game.Player, take bool, why string) {
		t.Helper()
		if effectstest.B07OpenSearchPrompts(g) != 1 {
			t.Fatalf("%s: %d search prompts open, want exactly 1 (the offer is sequential)", why, effectstest.B07OpenSearchPrompts(g))
		}
		c := effectstest.SearchChoiceFor(g, who.ID)
		if c == nil {
			t.Fatalf("%s: no search prompt for %s", why, who.Name)
		}
		for _, id := range c.SearchCards {
			card, _ := g.LookupCardForEffect(id)
			if !card.IsLand() {
				t.Errorf("%s: a non-land %q was offered for 'a land card'", why, card.Name)
			}
		}
		if take {
			effectstest.AnswerSearchByID(t, g, who.ID, c.SearchCards[0])
		} else {
			effectstest.AnswerSearchFailToFind(t, g, who.ID)
		}
	}
	// An opponent is first asked whether to search at all.
	ask := func(who *game.Player, yes bool) {
		t.Helper()
		effectstest.AnswerMayChoice(t, g, who.ID, yes)
	}
	step(me, true, "the caster's opening search")
	ask(opp1, true)
	step(opp1, true, "opponent 1 is offered a land")
	step(me, true, "opponent 1 accepted, so the caster searches again")
	ask(opp2, false)
	// Opponent 3 searches and takes nothing: still "searched a library
	// this way", so the caster still searches (the old caveat).
	ask(opp3, true)
	step(opp3, false, "opponent 3 searches")
	step(me, true, "opponent 3 searched, so the caster searches again")
	if effectstest.B07OpenSearchPrompts(g) != 0 {
		t.Fatalf("after the last opponent nothing should be open, got %d", effectstest.B07OpenSearchPrompts(g))
	}
	if spec, _ := effects.Lookup(effectstest.B07TemptWithDiscoveryOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("searching and finding nothing counts: nothing is declared")
	}

	for _, tc := range []struct {
		p    *game.Player
		want int
	}{{me, 3}, {opp1, 1}, {opp2, 0}, {opp3, 0}} {
		if got := effectstest.B07LandsOnBattlefield(g, tc.p.ID); got != tc.want {
			t.Errorf("%s controls %d lands, want %d", tc.p.Name, got, tc.want)
		}
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsLand() && c.Tapped {
			t.Errorf("%s entered tapped; the printed text says untapped", c.Name)
		}
	}
}

// --- lands ---------------------------------------------------------

func TestB07TronLandsScaleWithTheFullSet(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tower := effectstest.SeedPermanentWithOracle(g, me.ID, "Urza's Tower", "Land — Urza's Tower", effectstest.B07UrzasTowerOracle)
	if err := g.ActivateManaAbility(me.ID, tower, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("a lone Tower: pool %v, want [C]", got)
	}

	me.ManaPool.EmptyPool()
	mine := effectstest.SeedPermanentWithOracle(g, me.ID, "Urza's Mine", "Land — Urza's Mine", effectstest.B07UrzasMineOracle)
	effectstest.SeedPermanentWithOracle(g, me.ID, "Urza's Power Plant", "Land — Urza's Power-Plant", effectstest.B07UrzasPowerPlantOracle)
	tower2 := effectstest.SeedPermanentWithOracle(g, me.ID, "Urza's Tower", "Land — Urza's Tower", effectstest.B07UrzasTowerOracle)
	if err := g.ActivateManaAbility(me.ID, tower2, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 3 {
		t.Errorf("Tower with Tron: pool %v, want [C C C]", got)
	}
	me.ManaPool.EmptyPool()
	if err := g.ActivateManaAbility(me.ID, mine, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 2 {
		t.Errorf("Mine with Tron: pool %v, want [C C]", got)
	}
}

func TestB07MysticGateFiltersOneHybridIntoTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gate := effectstest.SeedPermanentWithOracle(g, me.ID, "Mystic Gate", "Land", effectstest.B07MysticGateOracle)

	if err := g.ActivateManaAbility(me.ID, gate, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the {C} half: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}

	// The wrong colour floating: refused, Gate untouched.
	fresh := effectstest.SeedPermanentWithOracle(g, me.ID, "Mystic Gate", "Land", effectstest.B07MysticGateOracle)
	me.ManaPool.EmptyPool()
	me.ManaPool.AddMana(game.ManaToken{Color: "G"})
	if err := g.ActivateManaAbility(me.ID, fresh, 1, game.ManaAbilityParams{}); err == nil {
		t.Fatal("{W/U} was paid with {G}")
	}
	if card, _ := effectstest.BattlefieldCard(g, fresh); card.Tapped {
		t.Fatal("a refused activation must not tap the Gate")
	}

	// {U} pays the hybrid; the pool then asks twice from {W, U}.
	me.ManaPool.EmptyPool()
	me.ManaPool.AddMana(game.ManaToken{Color: "U"})
	if err := g.ActivateManaAbility(me.ID, fresh, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	// The two picks are one production (#3939, CR 106.12b): one choice,
	// asked once per slot, and no mana until the last answer.
	picks := 0
	for c := effectstest.RiderLatestManaPick(g, me.ID); c != nil && picks < 3; c = effectstest.RiderLatestManaPick(g, me.ID) {
		picks++
		if len(c.ColorOptions) != 2 {
			t.Errorf("each slot must offer exactly W and U, got %v", c.ColorOptions)
		}
		if picks == 2 && me.ManaPool.Total() != 0 {
			t.Errorf("mana arrived before the last pick: %v", effectstest.Batch01PoolColors(me))
		}
		if err := g.ResolveManaChoice(c.ID, me.ID, "W"); err != nil {
			t.Fatalf("ResolveManaChoice: %v", err)
		}
	}
	if picks != 2 {
		t.Fatalf("want two colour picks, got %d", picks)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 2 || got[0] != "W" || got[1] != "W" {
		t.Errorf("pool %v, want [W W]", got)
	}
}
