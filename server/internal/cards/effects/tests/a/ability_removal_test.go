package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Darksteel Mutation -------------------------------------------

func TestDarksteelMutationMakesAnIndestructible0_1Insect(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	commander := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Big Commander",
		TypeLine: "Legendary Creature — Human Warrior", Power: 6, Toughness: 6,
		Owner: me.ID, Controller: me.ID,
	})

	effectstest.Enchant(t, g, "Darksteel Mutation", effectstest.DarksteelMutationOracle, commander)

	if got := effectstest.EffectivePower(t, g, commander); got != 0 {
		t.Errorf("power %d, want 0", got)
	}
	if got := effectstest.EffectiveToughness(t, g, commander); got != 1 {
		t.Errorf("toughness %d, want 1", got)
	}
	types := effectstest.EffectiveTypes(t, g, commander)
	if !effectstest.ContainsString(types, "Artifact") || !effectstest.ContainsString(types, "Creature") {
		t.Errorf("types %v, want an artifact creature", types)
	}
	subtypes := effectstest.EffectiveSubtypes(t, g, commander)
	if len(subtypes) != 1 || subtypes[0] != "Insect" {
		t.Errorf("subtypes %v, want exactly [Insect] — it loses all other creature types", subtypes)
	}
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, commander), "indestructible") {
		t.Error("the Mutation grants indestructible in the same breath it removes everything else")
	}
}

// "Loses all other abilities, CARD TYPES, and CREATURE TYPES" — and
// nothing else. Supertypes are neither, so the commander is still
// legendary, which is the whole reason this is a Commander staple:
// it answers the general without answering the legend rule.
func TestDarksteelMutationLeavesSupertypesAndCountersAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	commander := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Big Commander",
		TypeLine: "Legendary Creature — Human Warrior", Power: 6, Toughness: 6,
		Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{"+1/+1": 2},
	})

	effectstest.Enchant(t, g, "Darksteel Mutation", effectstest.DarksteelMutationOracle, commander)

	card := effectstest.LayeredCard(t, g, commander)
	if !effectstest.ContainsString(card.Effective().Supertypes, "Legendary") {
		t.Errorf("supertypes %v, want Legendary kept", card.Effective().Supertypes)
	}
	if card.Counters["+1/+1"] != 2 {
		t.Errorf("counters %v — CR 613.1f removes abilities, not counters", card.Counters)
	}
	// Base 0/1 is a layer 7b SET, so the two +1/+1 counters (7d)
	// still apply on top of it.
	if got := card.CurrentPower(); got != 2 {
		t.Errorf("current power %d, want 2 — 7b sets the base, 7d adds the counters", got)
	}
}

// The gap this whole change closes, seen through a card: a mana
// ability read from the catalog at use time used to keep working
// under a layer-6 removal because nothing connected the two.
func TestDarksteelMutationSilencesACatalogManaAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	birds := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Birds of Paradise",
		TypeLine: "Creature — Bird", Power: 0, Toughness: 1,
		OracleID: "d3a0b660-358c-41bd-9cd2-41fbf3491b1a",
		Owner:    me.ID, Controller: me.ID,
	})
	if got := game.ManaAbilitiesForCard(effectstest.LayeredCard(t, g, birds)); len(got) != 1 {
		t.Fatalf("fixture is wrong: Birds should have one mana ability, got %v", got)
	}

	effectstest.Enchant(t, g, "Darksteel Mutation", effectstest.DarksteelMutationOracle, birds)

	if got := game.ManaAbilitiesForCard(effectstest.LayeredCard(t, g, birds)); len(got) != 0 {
		t.Errorf("mana abilities = %v, want none", got)
	}
	if err := g.ActivateManaAbility(me.ID, birds, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("activating a removed mana ability should be refused")
	}
}

// CR 613.6, through two real cards: a Rancor attached AFTER the
// Mutation has a later timestamp, lands after the removal in the
// layer-6 sort, and its trample survives. The reverse order does
// not. Nothing in either card file says this.
func TestRancorAfterTheMutationGrantsTrampleAndBeforeItDoesNot(t *testing.T) {
	for _, tc := range []struct {
		name         string
		rancorFirst  bool
		wantTrample  bool
		wantIndestr  bool
		wantPowerInc int
	}{
		{name: "Rancor first", rancorFirst: true, wantTrample: false, wantIndestr: true, wantPowerInc: 2},
		{name: "Mutation first", rancorFirst: false, wantTrample: true, wantIndestr: true, wantPowerInc: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			bear := effectstest.SeedBear(g, me.ID)

			if tc.rancorFirst {
				effectstest.Enchant(t, g, "Rancor", effectstest.RancorOracle, bear)
				effectstest.Enchant(t, g, "Darksteel Mutation", effectstest.DarksteelMutationOracle, bear)
			} else {
				effectstest.Enchant(t, g, "Darksteel Mutation", effectstest.DarksteelMutationOracle, bear)
				effectstest.Enchant(t, g, "Rancor", effectstest.RancorOracle, bear)
			}

			abilities := effectstest.EffectiveAbilities(t, g, bear)
			if effectstest.ContainsString(abilities, "trample") != tc.wantTrample {
				t.Errorf("trample in %v, want %v", abilities, tc.wantTrample)
			}
			if effectstest.ContainsString(abilities, "indestructible") != tc.wantIndestr {
				t.Errorf("indestructible in %v, want %v", abilities, tc.wantIndestr)
			}
			// Rancor's +2/+0 is layer 7c and is not an ability
			// removal's business either way.
			if got := effectstest.EffectivePower(t, g, bear); got != tc.wantPowerInc {
				t.Errorf("power %d, want %d — 7b sets 0, 7c adds Rancor's +2", got, tc.wantPowerInc)
			}
		})
	}
}

// --- Kenrith's Transformation -------------------------------------

func TestKenrithsTransformationDrawsAndMakesAGreen3_3Elk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Black Bear",
		TypeLine: "Creature — Bear", Power: 7, Toughness: 7,
		ManaCost: "{4}{B}{B}", Owner: me.ID, Controller: me.ID,
	})
	before := len(me.Hand.Cards)

	effectstest.Enchant(t, g, "Kenrith's Transformation", effectstest.KenrithTransformOracle, bear)

	if got := len(me.Hand.Cards); got != before+1 {
		t.Errorf("hand size %d, want %d — the Aura cantrips on entry", got, before+1)
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("power %d, want 3", got)
	}
	if got := effectstest.EffectiveToughness(t, g, bear); got != 3 {
		t.Errorf("toughness %d, want 3", got)
	}
	subtypes := effectstest.EffectiveSubtypes(t, g, bear)
	if len(subtypes) != 1 || subtypes[0] != "Elk" {
		t.Errorf("subtypes %v, want exactly [Elk]", subtypes)
	}
	colors := effectstest.EffectiveColors(t, g, bear)
	if len(colors) != 1 || colors[0] != "G" {
		t.Errorf("colors %v, want [G] — the catalog's first layer-5 effect", colors)
	}
}

// Still a creature, so the Equipment stays on. This is the control
// for the Song of the Dryads test below: the attachments fall off
// there because of the layer-4 TYPE change, not because of the
// ability removal the two cards share.
func TestKenrithsTransformationKeepsTheEquipmentAttached(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	splitter := effectstest.SeedEquipment(g, me.ID, "Bonesplitter", effectstest.BonesplitterOracle)
	effectstest.EquipTo(t, g, me.ID, splitter, bear)

	effectstest.Enchant(t, g, "Kenrith's Transformation", effectstest.KenrithTransformOracle, bear)
	effectstest.Settle(t, g)

	if host := effectstest.AttachmentHostOf(t, g, splitter); host.ID != bear {
		t.Errorf("the Bonesplitter fell off a permanent that is still a creature: %+v", host)
	}
	// Base 3/3 (7b) plus the Bonesplitter's +2/+0 (7c).
	if got := effectstest.EffectivePower(t, g, bear); got != 5 {
		t.Errorf("power %d, want 5", got)
	}
}

// --- Song of the Dryads -------------------------------------------

func TestSongOfTheDryadsTurnsSolRingIntoAColorlessForest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	solRing := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sol Ring", TypeLine: "Artifact",
		OracleID: effectstest.SolRingOracle, Owner: me.ID, Controller: me.ID,
	})
	before := game.ManaAbilitiesForCard(effectstest.LayeredCard(t, g, solRing))
	if len(before) != 1 || before[0].Produced != "{C}{C}" {
		t.Fatalf("fixture is wrong: Sol Ring should add {C}{C}, got %v", before)
	}

	effectstest.Enchant(t, g, "Song of the Dryads", effectstest.SongOfTheDryadsOracle, solRing)

	card := effectstest.LayeredCard(t, g, solRing)
	if !card.IsLand() || card.IsArtifact() {
		t.Errorf("types %v, want a land and not an artifact", card.Effective().Types)
	}
	if got := card.Effective().Colors; len(got) != 0 {
		t.Errorf("colors %v, want colourless", got)
	}
	// CR 305.7: the declared ability goes, the land type's intrinsic
	// one arrives, and both come from this one effect.
	after := game.ManaAbilitiesForCard(card)
	if len(after) != 1 || after[0].Produced != "{G}" {
		t.Errorf("mana abilities %v, want exactly the Forest's {G}", after)
	}
}

// CR 704.5m and CR 704.5n, fired by a card in real play rather than
// by a hand-built board: the enchanted creature stops being a
// creature, so the Equipment on it unattaches and the "enchant
// creature" Aura on it goes to its owner's graveyard. #511 made the
// second of those reachable; this is the first card that reaches it
// without an effect putting an Aura onto the battlefield by hand.
func TestSongOfTheDryadsDropsTheHostsEquipmentAndAuras(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	splitter := effectstest.SeedEquipment(g, me.ID, "Bonesplitter", effectstest.BonesplitterOracle)
	effectstest.EquipTo(t, g, me.ID, splitter, bear)
	mastery := effectstest.Enchant(t, g, "Battle Mastery", effectstest.BattleMasteryOracle, bear)

	effectstest.Enchant(t, g, "Song of the Dryads", effectstest.SongOfTheDryadsOracle, bear)
	effectstest.Settle(t, g)

	if host := effectstest.AttachmentHostOf(t, g, splitter); host.Kind != "" {
		t.Errorf("CR 704.5n: the Equipment should have unattached, host = %+v", host)
	}
	if !g.Battlefield.Contains(splitter) {
		t.Error("CR 704.5n unattaches an Equipment; it does not destroy it")
	}
	if g.Battlefield.Contains(mastery) {
		t.Error("CR 704.5m: an \"enchant creature\" Aura on a permanent that stopped being a creature must leave")
	}
	// The Song itself says "enchant permanent", so it stays — and
	// that asymmetry is the card.
	if host := effectstest.AttachmentHostOf(t, g, bear); host.Kind != "" {
		t.Errorf("the host is not attached to anything: %+v", host)
	}
}

// The Song does not fall off the permanent it just turned into a
// land, because "enchant permanent" is still satisfied. Worth its
// own assertion: the CR 704.5m re-check runs the Aura's OWN target
// spec, so a wider clause is what keeps it there.
func TestSongOfTheDryadsStaysAttachedToWhatItMade(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.SeedBear(g, me.ID)

	song := effectstest.Enchant(t, g, "Song of the Dryads", effectstest.SongOfTheDryadsOracle, bear)
	effectstest.Settle(t, g)
	effectstest.Settle(t, g)

	if !g.Battlefield.Contains(song) {
		t.Fatal("the Song enchants a permanent, and a land is a permanent")
	}
	if host := effectstest.AttachmentHostOf(t, g, song); host.ID != bear {
		t.Errorf("host = %+v, want the land it made", host)
	}
}

// The claim the card file makes about counters, checked rather than
// asserted in prose: a planeswalker turned into a Forest keeps its
// loyalty counters, is NOT swept by CR 704.5i (it is not a
// planeswalker while the Song is on it), and gets everything back
// when the Song leaves.
func TestSongOfTheDryadsOnAPlaneswalkerKeepsItsLoyalty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Walker",
		TypeLine: "Legendary Planeswalker — Test", Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterLoyalty: 4},
	})

	song := effectstest.Enchant(t, g, "Song of the Dryads", effectstest.SongOfTheDryadsOracle, walker)
	effectstest.Settle(t, g)

	if !g.Battlefield.Contains(walker) {
		t.Fatal("CR 704.5i sweeps planeswalkers, and this is not one right now")
	}
	card := effectstest.LayeredCard(t, g, walker)
	if card.IsPlaneswalker() {
		t.Error("it should be a land")
	}
	if card.Counters[game.CounterLoyalty] != 4 {
		t.Errorf("loyalty counters %v — counters are not abilities", card.Counters)
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(song) })
	back := effectstest.LayeredCard(t, g, walker)
	if !back.IsPlaneswalker() || back.Counters[game.CounterLoyalty] != 4 {
		t.Errorf("the walker should come back whole: planeswalker=%v counters=%v",
			back.IsPlaneswalker(), back.Counters)
	}
}

// --- against the S24 restriction vocabulary -----------------------

// A restriction is not an ability of the restricted permanent, so
// removing that permanent's abilities does not lift it. `game/
// restrictions.go` states this as the reason restrictions ride in
// their own field rather than as a string in Abilities; here it is
// through two real cards, on a host that stays a creature so the
// Pacifism stays legally attached.
func TestDarksteelMutationDoesNotLiftAPacifism(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.SeedBear(g, me.ID)

	effectstest.Enchant(t, g, "Pacifism", effectstest.PacifismOracle, bear)
	effectstest.Enchant(t, g, "Darksteel Mutation", effectstest.DarksteelMutationOracle, bear)
	effectstest.Settle(t, g)

	card := effectstest.LayeredCard(t, g, bear)
	if !card.HasLostAllAbilities() {
		t.Fatal("fixture is wrong: the Insect should have lost its abilities")
	}
	if !game.Restricted(&card, game.CantAttack) {
		t.Error("the \"can't attack\" belongs to the Pacifism, not to the creature")
	}
}

// The other direction, and it is the one the recompute's second pass
// buys: silence the AURA and its restriction goes with it, because a
// permanent with no abilities generates no continuous effect at all.
func TestSongOfTheDryadsOnAPacifismLetsTheCreatureAttackAgain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.SeedBear(g, me.ID)

	pacifism := effectstest.Enchant(t, g, "Pacifism", effectstest.PacifismOracle, bear)
	if card := effectstest.LayeredCard(t, g, bear); !game.Restricted(&card, game.CantAttack) {
		t.Fatal("fixture is wrong: the bear should be pacified")
	}

	effectstest.Enchant(t, g, "Song of the Dryads", effectstest.SongOfTheDryadsOracle, pacifism)
	effectstest.Settle(t, g)

	card := effectstest.LayeredCard(t, g, bear)
	if game.Restricted(&card, game.CantAttack) {
		t.Error("a Pacifism that is now a Forest land restricts nothing")
	}
	if card.HasLostAllAbilities() {
		t.Error("the Song is on the Aura, not on the creature")
	}
}

// An Aura silencing another Aura. Control Magic's steal is a layer-2
// continuous effect contributed by its static ability, so removing
// its abilities hands the creature back — and layer 2 runs long
// before layer 6, which is the case the recompute's second pass
// exists for.
func TestSongOfTheDryadsOnAControlMagicGivesTheCreatureBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, owner := g.Seats[0], g.Seats[1]
	theirs := effectstest.SeedBear(g, owner.ID)

	steal := effectstest.Enchant(t, g, "Control Magic", effectstest.ControlMagicOracle, theirs)
	if got := effectstest.ControllerOf(t, g, theirs); got != me.ID {
		t.Fatalf("fixture is wrong: controller %s, want %s", got, me.ID)
	}

	effectstest.Enchant(t, g, "Song of the Dryads", effectstest.SongOfTheDryadsOracle, steal)
	effectstest.Settle(t, g)

	if got := effectstest.ControllerOf(t, g, theirs); got != owner.ID {
		t.Errorf("controller %s, want the creature back with %s", got, owner.ID)
	}
	// The Control Magic is a Forest now, not an Aura, so CR 704.5m
	// has nothing to say about it and it stays on the battlefield.
	if !g.Battlefield.Contains(steal) {
		t.Error("a permanent that stopped being an Aura is not swept by CR 704.5m")
	}
	// CR 704.5p is why the creature came back, and it is the whole
	// reason: a nonbattle, noncreature permanent that is neither an
	// Aura, an Equipment nor a Fortification becomes UNATTACHED. Song's
	// 2014-11-07 ruling says so in as many words. The ability removal
	// has nothing to do with it — layer 2 ran long before layer 4
	// could silence anything (CR 613.6), so without this rule the
	// Forest would keep the creature stolen.
	if c := effectstest.LayeredCard(t, g, steal); c.IsAttached() {
		t.Error("the Song'd Control Magic is still attached (CR 704.5p)")
	}
}

// The other half of CR 704.5p, and the one that shows the stale link
// after the Song leaves: a Bonesplitter that became a Forest is
// unattached, so the creature it was equipping is back to its printed
// power.
func TestSongOfTheDryadsUnattachesAnEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	splitter := effectstest.SeedEquipment(g, me.ID, "Bonesplitter", effectstest.BonesplitterOracle)
	effectstest.EquipTo(t, g, me.ID, splitter, bear)
	if got := effectstest.EffectivePower(t, g, bear); got != 4 {
		t.Fatalf("fixture is wrong: the equipped Bear is %d power, want 4", got)
	}

	effectstest.Enchant(t, g, "Song of the Dryads", effectstest.SongOfTheDryadsOracle, splitter)
	effectstest.Settle(t, g)

	if c := effectstest.LayeredCard(t, g, splitter); c.IsAttached() {
		t.Error("the Song'd Bonesplitter is still attached (CR 704.5p)")
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("the Bear is %d power, want its printed 2", got)
	}
}
