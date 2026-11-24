package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kaldheim batch 071 (S82 OLD, #5505).

const (
	oldB071Sculptor  = "73109353-51fa-43a1-93e0-b74536f372a5"
	oldB071Svella    = "7fd8a508-335c-4d5a-9916-23bf26dfff38"
	oldB071Augur     = "85aedc9f-df35-49bc-88c8-23efda9a4b1b"
	oldB071Priest    = "fb711b7d-bd7f-48e3-a6de-a12f94ef55d2"
	oldB071Harbinger = "0507a9d2-cbad-4a7b-b888-fbd5ac1ed02f"
	oldB071OldGrowth = "11e3f1d3-9d78-4c00-b965-e14407fd5210"
	oldB071Kolvori   = "490ac985-79b5-438d-a2e7-f47683b29683"
	oldB071Kaya      = "13a4f88b-6832-43a1-b4fa-c1ee60560bad"
	oldB071Niko      = "781162c4-9c0f-436c-b9a5-214e60e009e4"
	oldB071Valki     = "907ae517-22d5-4ac7-bc3a-3f4d5eaeeb57"
	oldB071Cosima    = "648100cd-d44a-48f8-87e4-a1d5467f348c"
	oldB071Cart      = "73da9297-4c7c-4388-820f-1189b5aa0b98"
)

func TestOldB071CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB071Sculptor: "Sculptor of Winter", oldB071Svella: "Svella, Ice Shaper", oldB071Augur: "Frost Augur",
		oldB071Priest: "Priest of the Haunted Edge", oldB071Harbinger: "Karfell Harbinger", oldB071OldGrowth: "Old-Growth Troll",
		oldB071Kolvori: "Kolvori, God of Kinship", oldB071Kolvori + "#1": "The Ringhart Crest", oldB071Kaya: "Kaya the Inexorable",
		oldB071Niko: "Niko Aris", oldB071Valki: "Valki, God of Lies", oldB071Valki + "#1": "Tibalt, Cosmic Impostor",
		oldB071Cosima: "Cosima, God of the Voyage", oldB071Cosima + "#1": "The Omenkeel", oldB071Cart: "Gilded Assault Cart",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB071SnowLand(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.BlbB1Push(g, owner, "Snow-Covered Forest", "Basic Snow Land — Forest", "", 0, 0)
}

func oldB071PlainLand(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.BlbB1Push(g, owner, "Forest", "Basic Land — Forest", "", 0, 0)
}

// oldB071Hand puts a fully specified card into a hand.
func oldB071Hand(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Hand.PushTop(c)
	return c.InstanceID
}

// oldB071Grave puts a fully specified card into a graveyard.
func oldB071Grave(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Graveyard.PushTop(c)
	return c.InstanceID
}

func oldB071Tapped(g *game.Game, id uuid.UUID) bool {
	c := effectstest.FindBattlefieldCardByID(g, id)
	return c != nil && c.Tapped
}

// --- Sculptor of Winter ------------------------------------------------------

func TestOldB071SculptorOfWinterUntapsOnlySnowLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sculptor := effectstest.BlbB1Push(g, me.ID, "Sculptor of Winter", "Snow Creature — Elf Rogue", oldB071Sculptor, 2, 2)
	snow, plain := oldB071SnowLand(g, me.ID), oldB071PlainLand(g, me.ID)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(snow); _ = g.TapTargetForEffect(plain) })
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, sculptor, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(plain)}); err == nil {
		t.Fatal("targeted a nonsnow land")
	}
	if err := g.ActivateCatalogAbility(me.ID, sculptor, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(snow)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if oldB071Tapped(g, snow) {
		t.Error("the snow land was not untapped")
	}
	if !oldB071Tapped(g, plain) {
		t.Error("the plain land was untapped")
	}
	if !oldB071Tapped(g, sculptor) {
		t.Error("the Sculptor did not tap")
	}
}

// --- Frost Augur -----------------------------------------------------------

func TestOldB071FrostAugurTakesATopSnowCardAndLeavesOthers(t *testing.T) {
	for _, snow := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		augur := effectstest.BlbB1Push(g, me.ID, "Frost Augur", "Snow Creature — Human Wizard", oldB071Augur, 1, 2)
		oldB071SnowLand(g, me.ID)
		tl := "Creature — Bear"
		if snow {
			tl = "Snow Creature — Bear"
		}
		top := effectstest.SeedLibraryTop(me, "Top Card", tl, "{G}")
		effectstest.AdvanceToMain(t, g)
		hand := me.Hand.Size()
		if err := g.ActivateCatalogAbility(me.ID, augur, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		p := effectstest.LatestChooseCardsFor(g, me.ID)
		if !snow {
			if p != nil && len(p.ChooseCards) > 0 {
				t.Errorf("a nonsnow top card was offered: %v", p.ChooseCards)
			}
			if me.Hand.Size() != hand || !me.Library.Contains(top) {
				t.Error("the nonsnow card left the library")
			}
			continue
		}
		if p == nil || len(p.ChooseCards) != 1 || p.ChooseCards[0] != top {
			for _, c := range g.PendingChoices {
				t.Logf("pending %s %s", c.Kind, c.Reason)
			}
			t.Logf("stack %d top card %v tapped %v", g.Stack.Size(), me.Library.Contains(top), oldB071Tapped(g, augur))
			t.Fatalf("no offer of the snow card")
		}
		if err := g.ResolveChooseCards(p.ID, me.ID, []uuid.UUID{top}); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
		if !me.Hand.Contains(top) {
			t.Error("the snow card did not reach the hand")
		}
	}
}

func TestOldB071FrostAugurNeedsSnowMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	augur := effectstest.BlbB1Push(g, me.ID, "Frost Augur", "Snow Creature — Human Wizard", oldB071Augur, 1, 2)
	oldB071PlainLand(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, augur, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err == nil {
		t.Fatal("{S} was paid with a nonsnow land")
	}
}

// --- Priest of the Haunted Edge ----------------------------------------------

func TestOldB071PriestShrinksByTheSnowLandsAndIsSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	priest := effectstest.BlbB1Push(g, me.ID, "Priest of the Haunted Edge", "Snow Creature — Zombie Cleric", oldB071Priest, 0, 4)
	oldB071SnowLand(g, me.ID)
	oldB071SnowLand(g, me.ID)
	oldB071PlainLand(g, me.ID)
	oldB071SnowLand(g, opp.ID)
	victim := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 5, 5)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, priest, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(victim)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if pw, tough := effectstest.EffectivePTOf(t, g, victim); pw != 3 || tough != 3 {
		t.Errorf("Ogre is %d/%d, want 3/3 (two snow lands of mine)", pw, tough)
	}
	if g.Battlefield.Contains(priest) {
		t.Error("the Priest was not sacrificed")
	}
}

// --- Karfell Harbinger -------------------------------------------------------

func TestOldB071KarfellHarbingerManaPaysForeteleAndInstantsAndSorceriesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	one := effectstest.B12Push(g, me.ID, "Karfell Harbinger", "Creature — Zombie Wizard", oldB071Harbinger, 1, 3)
	two := effectstest.B12Push(g, me.ID, "Karfell Harbinger", "Creature — Zombie Wizard", oldB071Harbinger, 1, 3)
	effectstest.AdvanceToMain(t, g)
	for _, id := range []uuid.UUID{one, two} {
		if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("tap for mana: %v", err)
		}
	}
	bear := oldB071Hand(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{U}", Power: 2, Toughness: 2})
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("the Harbinger's mana paid for a creature spell")
	}
	// {2} to foretell a card from the hand.
	foretold := oldB071Hand(me, game.Card{Name: "Saw It Coming", TypeLine: "Instant", OracleID: effectstest.SawItComingOracle, ManaCost: "{1}{U}{U}"})
	if err := g.PerformSpecialAction(me.ID, foretold, game.SpecialActionForetell, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("foretell with the Harbinger's mana: %v", err)
	}
	if !g.Exile.Contains(foretold) {
		t.Error("the card was not foretold")
	}
}

func TestOldB071KarfellHarbingerManaCastsAnInstant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	h := effectstest.B12Push(g, me.ID, "Karfell Harbinger", "Creature — Zombie Wizard", oldB071Harbinger, 1, 3)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, h, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	spell := oldB071Hand(me, game.Card{Name: "Zap", TypeLine: "Instant", ManaCost: "{U}"})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the Harbinger's mana could not cast an instant: %v", err)
	}
}

// --- Svella, Ice Shaper -------------------------------------------------------

func TestOldB071SvellaMakesAnIcyManalithThatTapsForSnowMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	svella := effectstest.BlbB1Push(g, me.ID, "Svella, Ice Shaper", "Legendary Snow Creature — Troll Warrior", oldB071Svella, 2, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, svella, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Icy Manalith"); n != 1 {
		t.Fatalf("%d Icy Manaliths, want 1", n)
	}
	var id uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Icy Manalith" {
			id = c.InstanceID
			if !c.HasSupertype("snow") || !c.IsArtifact() || len(c.Colors) != 0 {
				t.Errorf("Icy Manalith is %q colours %v, want a colorless snow artifact", c.TypeLine, c.Colors)
			}
		}
	}
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Icy Manalith: %v", err)
	}
	effectstest.ResolveManaPickAnyColor(t, g, me.ID)
	if n := len(me.ManaPool.Units()); n != 1 {
		c := effectstest.FindBattlefieldCardByID(g, id)
		t.Errorf("the Icy Manalith made %d mana, want 1 (card %q tapped %v pending %d)", n, c.TypeLine, c.Tapped, len(g.PendingChoices))
		for _, pc := range g.PendingChoices {
			t.Logf("pending %s", pc.Kind)
		}
	}
}

// --- Old-Growth Troll --------------------------------------------------------

func TestOldB071OldGrowthTrollReturnsAsAnAuraOnAForestWithTwoGreenAndATrollMaker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := oldB071PlainLand(g, me.ID)
	troll := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Old-Growth Troll", OracleID: oldB071OldGrowth,
		TypeLine: "Creature — Troll Warrior", ManaCost: "{G}{G}{G}", Colors: []string{"G"}, Power: 4, Toughness: 4})
	effectstest.RadDestroy(t, g, troll)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.GaLayered(t, g, troll)
	if effectstest.AdZone(g, troll) != game.ZoneBattlefield || !c.IsAttachedTo(forest) || !c.IsAura() || c.IsCreature() {
		t.Fatalf("the Troll did not return as an Aura on the Forest: zone %v attached %+v", effectstest.AdZone(g, troll), c.AttachedTo)
	}
	// The Forest has "{T}: Add {G}{G}".
	if err := effectstest.TapGrantedMana(t, g, me.ID, forest, "G"); err != nil {
		t.Fatalf("tap the enchanted Forest: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 2 || got[0] != "G" || got[1] != "G" {
		t.Errorf("pool %v, want {G}{G}", got)
	}
}

func TestOldB071OldGrowthTrollSacrificedForestMakesATappedTrollWarrior(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := oldB071PlainLand(g, me.ID)
	troll := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Old-Growth Troll", OracleID: oldB071OldGrowth,
		TypeLine: "Creature — Troll Warrior", ManaCost: "{G}{G}{G}", Colors: []string{"G"}, Power: 4, Toughness: 4})
	effectstest.RadDestroy(t, g, troll)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}")
	// The granted activated ability is the Forest's second row after its mana ability.
	var err error
	for idx := 0; idx < 3; idx++ {
		err = g.ActivateCatalogAbility(me.ID, forest, idx, game.ActivateAbilityParams{Strict: true})
		if err == nil {
			break
		}
	}
	if err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Troll Warrior"); n != 1 {
		t.Fatalf("%d Troll Warrior tokens, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Troll Warrior" {
			if !c.Tapped || c.Power != 4 || c.Toughness != 4 || !game.HasKeyword(&c, "trample") {
				t.Errorf("token is %d/%d tapped=%v, want a tapped 4/4 with trample", c.Power, c.Toughness, c.Tapped)
			}
		}
	}
	if g.Battlefield.Contains(forest) {
		t.Error("the Forest was not sacrificed")
	}
}

// --- Gilded Assault Cart -----------------------------------------------------

func TestOldB071GildedAssaultCartCrews(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cart := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Gilded Assault Cart", OracleID: oldB071Cart,
		TypeLine: "Artifact — Vehicle", Power: 5, Toughness: 1})
	crew := effectstest.B12Creature(g, me.ID, "Crew", "Creature — Dwarf", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, cart, 0, game.ActivateAbilityParams{Strict: true, CrewIDs: []uuid.UUID{crew}}); err != nil {
		t.Fatalf("crew: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.GaLayered(t, g, cart).IsCreature() {
		t.Error("the Cart did not become a creature")
	}
}

func TestOldB071GildedAssaultCartReturnsFromTheGraveyardForTwoTreasures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cart := oldB071Grave(me, game.Card{Name: "Gilded Assault Cart", TypeLine: "Artifact — Vehicle", OracleID: oldB071Cart, ManaCost: "{1}{R}{R}", Power: 5, Toughness: 1})
	t1 := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Treasure", TypeLine: "Token Artifact — Treasure", Owner: me.ID, Controller: me.ID})
	t2 := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Treasure", TypeLine: "Token Artifact — Treasure", Owner: me.ID, Controller: me.ID})
	t3 := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Treasure", TypeLine: "Token Artifact — Treasure", Owner: me.ID, Controller: me.ID})
	_ = t3
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, cart, 1, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{t1}}); err == nil {
		t.Fatal("one Treasure paid for two")
	}
	if err := g.ActivateCatalogAbility(me.ID, cart, 1, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{t1, t2}}); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(cart) {
		t.Error("the Cart did not return to the hand")
	}
	if g.Battlefield.Contains(t1) || g.Battlefield.Contains(t2) {
		t.Error("the Treasures were not sacrificed")
	}
}

// --- Kolvori, God of Kinship // The Ringhart Crest ------------------------------

func TestOldB071KolvoriGrowsWithThreeLegendaryCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	kol := effectstest.B12Push(g, me.ID, "Kolvori, God of Kinship", "Legendary Creature — God", oldB071Kolvori, 2, 4)
	if pw, tough := effectstest.EffectivePTOf(t, g, kol); pw != 2 || tough != 4 {
		t.Fatalf("Kolvori alone is %d/%d, want 2/4", pw, tough)
	}
	effectstest.B12Push(g, me.ID, "Legend A", "Legendary Creature — Elf", "", 1, 1)
	effectstest.B12Push(g, g.Seats[1].ID, "Their Legend", "Legendary Creature — Elf", "", 1, 1)
	effectstest.B12Push(g, me.ID, "Plain", "Creature — Elf", "", 1, 1)
	if pw, _ := effectstest.EffectivePTOf(t, g, kol); pw != 2 {
		t.Errorf("Kolvori is %d power with two legends of mine, want 2", pw)
	}
	effectstest.B12Push(g, me.ID, "Legend B", "Legendary Creature — Elf", "", 1, 1)
	if pw, tough := effectstest.EffectivePTOf(t, g, kol); pw != 6 || tough != 6 {
		t.Errorf("Kolvori is %d/%d with three legends, want 6/6", pw, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, kol, "vigilance") {
		t.Error("Kolvori has no vigilance with three legendary creatures")
	}
}

func TestOldB071KolvoriLooksForALegendaryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	kol := effectstest.B12Push(g, me.ID, "Kolvori, God of Kinship", "Legendary Creature — God", oldB071Kolvori, 2, 4)
	deep := effectstest.SeedLibraryTop(me, "Deep Legend", "Legendary Creature — Elf", "{G}")
	effectstest.SeedLibraryTop(me, "Filler", "Creature — Bear", "{G}")
	var rest []uuid.UUID
	for i := 0; i < 4; i++ {
		rest = append(rest, effectstest.SeedLibraryTop(me, "Bear", "Creature — Bear", "{G}"))
	}
	hit := effectstest.SeedLibraryTop(me, "Hit Legend", "Legendary Creature — Elf", "{G}")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{G}")
	if err := g.ActivateCatalogAbility(me.ID, kol, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChooseCardsFor(g, me.ID)
	if p == nil || len(p.ChooseCards) != 1 || p.ChooseCards[0] != hit {
		t.Fatalf("offered %+v, want only the legendary creature among the top six", p)
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, []uuid.UUID{hit}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(hit) {
		t.Error("the legendary creature did not reach the hand")
	}
	if n := len(me.Library.Cards); me.Library.Cards[n-1].InstanceID != deep {
		t.Errorf("the seventh card is not on top: %s", me.Library.Cards[n-1].Name)
	}
	_ = rest
}

func TestOldB071RinghartCrestManaCastsChosenTypeOrLegendaryCreatureSpellsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	crest := effectstest.PushNamedTribePermanent(t, g, me.ID, "The Ringhart Crest", "Legendary Artifact", oldB071Kolvori+"#1", "Elf")
	effectstest.AdvanceToMain(t, g)
	cast := func(name, typeLine string) error {
		if err := g.ActivateManaAbility(me.ID, crest, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("tap the Crest: %v", err)
		}
		id := oldB071Hand(me, game.Card{Name: name, TypeLine: typeLine, ManaCost: "{G}", Power: 1, Toughness: 1})
		err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
		if err == nil {
			effectstest.PassPriorityAroundTable(t, g)
		}
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(crest); me.ManaPool.EmptyPool() })
		return err
	}
	if err := cast("Elf Scout", "Creature — Elf Scout"); err != nil {
		t.Errorf("an Elf creature spell was refused: %v", err)
	}
	if err := cast("Legend", "Legendary Creature — Human"); err != nil {
		t.Errorf("a legendary creature spell was refused: %v", err)
	}
	if err := cast("Bear", "Creature — Bear"); err == nil {
		t.Error("a nonlegendary creature of another type was cast")
	}
	if err := cast("Elf Sorcery", "Sorcery"); err == nil {
		t.Error("a noncreature spell was cast")
	}
}

// --- Niko Aris --------------------------------------------------------------

func TestOldB071NikoArisMakesXShardsAndAShard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := oldB071Hand(me, game.Card{Name: "Niko Aris", TypeLine: "Legendary Planeswalker — Niko", OracleID: oldB071Niko, ManaCost: "{X}{W}{U}{U}"})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 3}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Shard"); n != 3 {
		t.Fatalf("%d Shard tokens, want 3", n)
	}
	if got := effectstest.LoyaltyOf(g, id); got != 3 {
		t.Errorf("Niko's loyalty is %d, want 3", got)
	}
	// −1: Create a Shard token.
	if err := g.ActivateCatalogAbility(me.ID, id, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the second −1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Shard"); n != 4 {
		t.Errorf("%d Shard tokens, want 4", n)
	}
}

func TestOldB071NikoArisDamagesATappedCreatureForEachCardDrawn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	niko := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Niko Aris", "Legendary Planeswalker — Niko", oldB071Niko, 3)
	for i := 0; i < 4; i++ {
		effectstest.SeedLibraryTop(me, "Filler", "Creature — Bear", "{G}")
	}
	tapped := effectstest.B12Creature(g, opp.ID, "Tapped Ogre", "Creature — Ogre", 2, 9)
	untapped := effectstest.B12Creature(g, opp.ID, "Untapped Ogre", "Creature — Ogre", 2, 9)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(tapped) })
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 2) })
	if err := g.ActivateCatalogAbility(me.ID, niko, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(untapped)}); err == nil {
		t.Fatal("targeted an untapped creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, niko, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(tapped)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.FindBattlefieldCardByID(g, tapped)
	if c == nil || c.DamageMarked < 4 {
		t.Errorf("the Ogre has %v damage, want at least 4 (2 per card drawn)", c)
	}
}

func TestOldB071NikoArisUnblockableCreatureReturnsWhenItDealsDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	niko := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Niko Aris", "Legendary Planeswalker — Niko", oldB071Niko, 3)
	runner := effectstest.B12Creature(g, me.ID, "Runner", "Creature — Rogue", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, niko, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(runner)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(runner, opp.ID, 2) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(runner) {
		t.Error("the creature was not returned to its owner's hand after dealing damage")
	}
}

// --- Kaya the Inexorable ------------------------------------------------------

func TestOldB071KayaGhostformReturnsADyingCreatureAndMakesASpirit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	kaya := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Kaya the Inexorable", "Legendary Planeswalker — Kaya", oldB071Kaya, 5)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, kaya, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("activate +1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, "ghostform"); n != 1 {
		t.Fatalf("%d ghostform counters, want 1", n)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("the creature did not return to its owner's hand")
	}
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Spirit"); n != 1 {
		t.Errorf("%d Spirit tokens, want 1", n)
	}
}

func TestOldB071KayaGhostformAlsoTriggersOnExileAndOnlyForNontokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kaya := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Kaya the Inexorable", "Legendary Planeswalker — Kaya", oldB071Kaya, 5)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Token Creature — Goblin",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, kaya, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(token)}); err == nil {
		t.Fatal("a token was targeted")
	}
	if err := g.ActivateCatalogAbility(me.ID, kaya, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(theirs)}); err != nil {
		t.Fatalf("activate +1 on an opposing creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(theirs) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(theirs) {
		t.Error("the exiled creature did not return to its owner's hand")
	}
	if n := effectstest.CountBattlefieldNamed(g, opp.ID, "Spirit"); n != 1 {
		t.Errorf("%d Spirit tokens for the creature's controller, want 1", n)
	}
}

func TestOldB071KayaExilesANonlandPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kaya := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Kaya the Inexorable", "Legendary Planeswalker — Kaya", oldB071Kaya, 5)
	rock := effectstest.PushCatalogPermanent(g, opp.ID, "Rock", "Artifact", "", false)
	land := effectstest.PushCatalogPermanent(g, opp.ID, "Forest", "Basic Land — Forest", "", false)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, kaya, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(land)}); err == nil {
		t.Fatal("exiled a land")
	}
	if err := g.ActivateCatalogAbility(me.ID, kaya, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(rock)}); err != nil {
		t.Fatalf("activate −3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(rock) {
		t.Error("the artifact was not exiled")
	}
}

func TestOldB071KayaEmblemOffersALegendarySpellFromHandGraveyardOrExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	kaya := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Kaya the Inexorable", "Legendary Planeswalker — Kaya", oldB071Kaya, 7)
	handLegend := effectstest.HandCardForTest(me, "Hand Legend", "Legendary Creature — Elf", "")
	plain := effectstest.HandCardForTest(me, "Plain Bear", "Creature — Bear", "")
	graveLegend := effectstest.B38Graveyard(g, me.ID, "Grave Legend", "Legendary Creature — Elf", "{3}{G}", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, kaya, 2, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate −7: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Emblems.Cards) != 1 {
		t.Fatalf("%d emblems, want 1", len(me.Emblems.Cards))
	}
	// The next upkeep: the emblem asks which legendary spell, if any.
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChooseCardsFor(g, me.ID)
	if p == nil {
		t.Fatal("the emblem asked nothing at the beginning of the upkeep")
	}
	want := map[uuid.UUID]bool{handLegend: true, graveLegend: true, kaya: true}
	if len(p.ChooseCards) != 3 {
		t.Fatalf("offered %v, want the three legendary cards (Kaya went to the graveyard at 0 loyalty)", p.ChooseCards)
	}
	for _, id := range p.ChooseCards {
		if !want[id] || id == plain {
			t.Errorf("offered a card that is not a legendary spell: %v", id)
		}
	}
}

// --- Valki, God of Lies // Tibalt, Cosmic Impostor ----------------------------

func TestOldB071ValkiExilesACreatureCardFromEachOpponentUntilHeLeavesAndCopiesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirBear := oldB071Hand(opp, game.Card{Name: "Their Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 3, Toughness: 3})
	theirSpell := oldB071Hand(opp, game.Card{Name: "Their Spell", TypeLine: "Sorcery", ManaCost: "{R}"})
	valki := effectstest.CastCatalogSpell(t, g, "Valki, God of Lies", "Legendary Creature — God", oldB071Valki, nil)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChooseCardsFor(g, me.ID)
	if p == nil || len(p.ChooseCards) != 1 || p.ChooseCards[0] != theirBear {
		t.Fatalf("offered %+v, want only their creature card (not %v)", p, theirSpell)
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, []uuid.UUID{theirBear}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(theirBear) {
		t.Fatal("the creature card was not exiled")
	}
	// {X}: X = 2 copies it.
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, valki, 0, game.ActivateAbilityParams{Strict: true, XValue: 2}); err != nil {
		t.Fatalf("activate {X}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardByID(g, valki); c == nil || c.Name != "Their Bear" {
		t.Errorf("Valki is %v, want a copy of Their Bear", c)
	}
	// He leaves: the card comes back.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(valki) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(theirBear) && !g.Battlefield.Contains(theirBear) && !g.Exile.Contains(theirBear) {
		t.Log("the exiled card's whereabouts are not asserted further")
	}
	if g.Exile.Contains(theirBear) {
		t.Error("the exiled card did not return when Valki left")
	}
}

func TestOldB071TibaltExilesAndLetsYouPlayWhatItExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibraryTop(me, "My Top", "Creature — Bear", "{G}")
	theirTop := effectstest.SeedLibraryTop(opp, "Their Top", "Basic Land — Forest", "")
	tibalt := effectstest.CastCatalogSpell(t, g, "Tibalt, Cosmic Impostor", "Legendary Planeswalker — Tibalt", oldB071Valki+"#1", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyOf(g, tibalt); got != 5 {
		t.Fatalf("Tibalt has %d loyalty, want 5", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, tibalt, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate +2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(theirTop) {
		t.Fatal("the opponent's top card was not exiled")
	}
	if got := effectstest.LoyaltyOf(g, tibalt); got != 7 {
		t.Errorf("Tibalt has %d loyalty after +2, want 7", got)
	}
	// The exiled land may be played by me, and only me.
	if err := g.CastSpell(opp.ID, theirTop, game.CastSpellParams{Strict: true, FromZone: "exile"}); err == nil {
		t.Error("the opponent played a card exiled with Tibalt")
	}
	if err := g.CastSpell(me.ID, theirTop, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Errorf("I could not play the land exiled with Tibalt: %v", err)
	}
}

// --- Cosima, God of the Voyage // The Omenkeel ------------------------------

func TestOldB071CosimaExilesAtUpkeepCountsVoyagesAndReturnsWithCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cosima := effectstest.B12Push(g, me.ID, "Cosima, God of the Voyage", "Legendary Creature — God", oldB071Cosima, 2, 4)
	for i := 0; i < 4; i++ {
		effectstest.SeedLibraryTop(me, "Filler", "Creature — Bear", "{G}")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(cosima) {
		t.Fatal("Cosima was not exiled")
	}

	land := func() {
		g.WithWriteLock(func() {
			_ = g.CreateTokenForEffect(me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}, 1)
		})
		effectstest.PassPriorityAroundTable(t, g)
	}
	choose := func(label string) {
		c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID)
		if c == nil {
			t.Fatal("no voyage question")
		}
		for i, o := range c.PickOptions {
			if o.Label == label {
				if err := g.ResolveOptionPick(c.ID, me.ID, i); err != nil {
					t.Fatalf("ResolveOptionPick: %v", err)
				}
				effectstest.PassPriorityAroundTable(t, g)
				return
			}
		}
		t.Fatalf("no option %q in %+v", label, c.PickOptions)
	}
	land()
	choose("Put a voyage counter on Cosima")
	land()
	choose("Put a voyage counter on Cosima")
	if c, _ := g.LookupCardForEffect(cosima); c.Counters["voyage"] != 2 {
		t.Fatalf("%d voyage counters on the exiled card, want 2", c.Counters["voyage"])
	}
	hand := me.Hand.Size()
	land()
	choose("Don't: return Cosima to the battlefield with X +1/+1 counters and draw X cards")
	back := uuid.Nil
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Cosima, God of the Voyage" && c.Controller == me.ID {
			back = c.InstanceID
		}
	}
	if back == uuid.Nil {
		t.Fatal("Cosima did not return")
	}
	if n := effectstest.CountersOn(g, back, game.CounterPlusOne); n != 2 {
		t.Errorf("%d +1/+1 counters on Cosima, want 2", n)
	}
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d, want %d (drew two)", me.Hand.Size(), hand+2)
	}
}

func TestOldB071CosimaWithoutHerUpkeepExileHasNoVoyageAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cosima := effectstest.B12Push(g, me.ID, "Cosima, God of the Voyage", "Legendary Creature — God", oldB071Cosima, 2, 4)
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(cosima) })
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}, 1)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID); c != nil {
		t.Error("a Cosima exiled another way asked about a voyage counter")
	}
}

func TestOldB071OmenkeelExilesCardsAndLetsYouPlayTheLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.SeedLibraryTop(opp, "Their Forest", "Basic Land — Forest", "")
	spell := effectstest.SeedLibraryTop(opp, "Their Spell", "Sorcery", "{R}")
	keel := effectstest.B12Push(g, me.ID, "The Omenkeel", "Legendary Artifact Creature — Vehicle", oldB071Cosima+"#1", 3, 3)
	effectstest.AttackWith(t, g, opp.ID, keel)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	if !g.Exile.Contains(spell) || !g.Exile.Contains(land) {
		t.Fatal("the opponent's top two cards were not exiled")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, FromZone: "exile"}); err == nil {
		t.Error("played a nonland card exiled this way")
	}
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Errorf("could not play the land exiled this way: %v", err)
	}
}

func TestOldB071TibaltMinusThreeAndMinusEightExileAndGivePlayPermission(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tibalt := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Tibalt, Cosmic Impostor", "Legendary Planeswalker — Tibalt", oldB071Valki+"#1", 8)
	g.WithWriteLock(func() { _ = g.SetChosenPlayerForEffect(tibalt, me.ID) })
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	oldB071Grave(me, game.Card{Name: "My Dead", TypeLine: "Creature — Elf", ManaCost: "{G}"})
	theirDead := oldB071Grave(opp, game.Card{Name: "Their Dead", TypeLine: "Basic Land — Forest"})
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, tibalt, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("activate −3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(bear) {
		t.Fatal("the creature was not exiled")
	}
	// One loyalty ability a turn for each: a second Tibalt activates the −8.
	other := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Tibalt, Cosmic Impostor", "Legendary Planeswalker — Tibalt", oldB071Valki+"#1", 8)
	g.WithWriteLock(func() { _ = g.SetChosenPlayerForEffect(other, me.ID) })
	if err := g.ActivateCatalogAbility(me.ID, other, 2, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate −8: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(theirDead) || len(me.Graveyard.Cards) != 0 || len(opp.Graveyard.Cards) != 0 {
		t.Error("the graveyards were not exiled")
	}
	r := 0
	for _, m := range me.ManaPool.Units() {
		if m.Color == "R" {
			r++
		}
	}
	if r != 3 {
		t.Errorf("%d red mana from the −8, want 3", r)
	}
	if err := g.CastSpell(me.ID, theirDead, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Errorf("could not play the land exiled by the −8: %v", err)
	}
}
