package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// grantedManaAbilitiesFor is grantedManaAbilityFor's plural, for cards
// that grant more than one bundle to the same recipient (Resonating
// Lute).
func grantedManaAbilitiesFor(g *game.Game, id uuid.UUID) ([]game.ManaAbilityShape, game.AbilityOrigins) {
	var abs []game.ManaAbilityShape
	var origins game.AbilityOrigins
	g.ReadSnapshot(func() {
		c, found := g.LookupCardForEffect(id)
		if !found {
			return
		}
		abs, origins = game.ManaAbilitiesWithOrigins(c)
	})
	var gAbs []game.ManaAbilityShape
	var gOrigins game.AbilityOrigins
	for i := range abs {
		if origins.At(i).Granted() {
			gAbs = append(gAbs, abs[i])
			gOrigins = append(gOrigins, origins[i])
		}
	}
	return gAbs, gOrigins
}

func TestSlice294bAbundantGrowthEnchantsALandAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat].ID
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me, Controller: me})
	before := len(g.Seats[seat].Hand.Cards)

	effectstest.CastCatalogSpell(t, g, "Abundant Growth", "Enchantment — Aura", effectstest.AbundantGrowthOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land}})
	effectstest.PassPriorityAroundTable(t, g)

	if after := len(g.Seats[seat].Hand.Cards); after != before+1 {
		// The Aura leaves the hand to the stack (-1), then the draw
		// trigger resolves (+1): net +1 over the pre-cast baseline.
		t.Errorf("hand size = %d, want %d (aura cast, one card drawn)", after, before+1)
	}
	idx, ref, ok := effectstest.GrantedManaAbilityFor(g, land)
	if !ok {
		t.Fatalf("the enchanted land should have the granted mana ability")
	}
	if err := g.ActivateManaAbility(me, land, idx, game.ManaAbilityParams{Ref: ref, Colors: []string{"B"}}); err != nil {
		t.Fatalf("activate granted ability: %v", err)
	}
}

func TestSlice294bEnduringVitalityHasVigilanceAndGrantsMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Enduring Vitality", TypeLine: "Enchantment Creature — Elk Glimmer", OracleID: effectstest.EnduringVitalityOracle, Owner: me, Controller: me})
	if !effectstest.HasEffectiveKeyword(t, g, id, "vigilance") {
		t.Errorf("Enduring Vitality should have vigilance")
	}
	if _, _, ok := effectstest.GrantedManaAbilityFor(g, id); !ok {
		t.Errorf("Enduring Vitality should grant itself the mana ability (\"creatures you control\")")
	}
}

func TestSlice294bElvenChorusGrantsAndTopOfLibraryPermissions(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Owner: me, Controller: me})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Elven Chorus", TypeLine: "Enchantment", OracleID: effectstest.ElvenChorusOracle, Owner: me, Controller: me})

	if _, _, ok := effectstest.GrantedManaAbilityFor(g, bear); !ok {
		t.Errorf("Elven Chorus should grant creatures you control the mana ability")
	}
	if v := game.CatalogLibraryTopVisible(effectstest.ElvenChorusOracle); v != game.LibraryTopOwner {
		t.Errorf("LibraryTopVisible = %v, want LibraryTopOwner (a private look, not a public reveal)", v)
	}
}

func TestSlice294bNightOfTheSweetsRevengeCreatesFoodAndPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})

	id := effectstest.CastCatalogSpell(t, g, "Night of the Sweets' Revenge", "Enchantment", effectstest.NightOfTheSweetsRevengeCard, nil)
	effectstest.PassPriorityAroundTable(t, g)

	var foodID uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.HasSubtype("Food") {
			foodID = c.InstanceID
		}
	}
	if foodID == uuid.Nil {
		t.Fatalf("no Food token on the battlefield")
	}
	idx, ref, ok := effectstest.GrantedManaAbilityFor(g, foodID)
	if !ok {
		t.Fatalf("the Food should have the granted mana ability")
	}
	if err := g.ActivateManaAbility(me.ID, foodID, idx, game.ManaAbilityParams{Ref: ref}); err != nil {
		t.Fatalf("activate granted ability: %v", err)
	}
	if n := me.ManaPool.Total(); n != 1 || me.ManaPool.Units()[0].Color != "G" {
		t.Errorf("pool = %+v, want one {G} token", me.ManaPool)
	}

	// Fund the sac ability's {5}{G}{G} cost and pump: one Food on the
	// board -> +1/+1.
	for i := 0; i < 7; i++ {
		me.ManaPool = append(me.ManaPool, game.ManaToken{Color: "G", Source: uuid.New()})
	}
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	if p := effectstest.EffectivePower(t, g, bear); p != 3 {
		t.Errorf("Bear power = %d, want 3 (base 2 + X=1 Food)", p)
	}
}

func TestSlice294bAlchemistsTalentLevelTwoGrantsTreasureMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]

	id := effectstest.CastCatalogSpell(t, g, "Alchemist's Talent", "Enchantment — Class", effectstest.AlchemistsTalentOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	var treasureID uuid.UUID
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.HasSubtype("Treasure") {
			n++
			treasureID = c.InstanceID
		}
	}
	if n != 2 {
		t.Fatalf("Treasures on battlefield = %d, want 2", n)
	}
	if _, _, ok := effectstest.GrantedManaAbilityFor(g, treasureID); ok {
		t.Fatalf("Treasures should not have the grant before level 2")
	}

	me.ManaPool = append(me.ManaPool, game.ManaToken{Color: "R", Source: uuid.New()}, game.ManaToken{Color: "R", Source: uuid.New()})
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})

	idx, ref, ok := effectstest.GrantedManaAbilityFor(g, treasureID)
	if !ok {
		t.Fatalf("Treasures should have the grant at level 2")
	}
	// The Treasure was created TAPPED (ETB, "create two tapped
	// Treasure tokens") — untap it so the granted {T} ability can be
	// tested; this is the same permanent's own next untap step in a
	// real game.
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(treasureID) })
	if err := g.ActivateManaAbility(me.ID, treasureID, idx, game.ManaAbilityParams{Ref: ref, Colors: []string{"R"}}); err != nil {
		t.Fatalf("activate granted ability: %v", err)
	}
	if n := me.ManaPool.Total(); n != 2 {
		t.Errorf("pool has %d tokens, want two (from the sacrificed Treasure)", n)
	}
}

func TestSlice294bWrennGrantsAnimatesAndMills(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	wrenn := effectstest.PushCatalogPermanent(g, me.ID, "Wrenn and Realmbreaker", "Legendary Planeswalker — Wrenn", effectstest.WrennAndRealmbreakerOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(wrenn, game.CounterLoyalty, 4) })

	if _, _, ok := effectstest.GrantedManaAbilityFor(g, land); !ok {
		t.Fatalf("Wrenn should grant lands you control the any-color mana ability")
	}

	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, wrenn, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}}})
	if !effectstest.HasEffectiveKeyword(t, g, land, "haste") || !effectstest.HasEffectiveKeyword(t, g, land, "vigilance") || !effectstest.HasEffectiveKeyword(t, g, land, "hexproof") {
		t.Errorf("the animated land should have vigilance, hexproof and haste")
	}
	if p := effectstest.EffectivePower(t, g, land); p != 3 {
		t.Errorf("animated land power = %d, want 3", p)
	}

	effectstest.B39NextTurnOf(t, g, 0)
	libBefore := len(me.Library.Cards)
	effectstest.B16Activate(t, g, me.ID, wrenn, 1, game.ActivateAbilityParams{})
	if libAfter := len(me.Library.Cards); libBefore-libAfter != 3 {
		t.Errorf("library shrank by %d, want 3 (mill three)", libBefore-libAfter)
	}
}

func TestSlice294bMachineGodsEffigyTapsForBlue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Machine God's Effigy", TypeLine: "Artifact", OracleID: effectstest.MachineGodsEffigyOracle, Owner: me, Controller: me})
	if err := g.ActivateManaAbility(me, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if n := g.Seats[0].ManaPool.Total(); n != 1 || g.Seats[0].ManaPool.Units()[0].Color != "U" {
		t.Errorf("pool = %+v, want one {U} token", g.Seats[0].ManaPool)
	}
}

func TestSlice294bResonatingLuteRestrictedLandManaAndHandSizeDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	lute := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Resonating Lute", TypeLine: "Artifact", OracleID: effectstest.ResonatingLuteOracle, Owner: me.ID, Controller: me.ID})

	abs, _ := grantedManaAbilitiesFor(g, land)
	if len(abs) != 2 {
		t.Fatalf("land has %d granted mana abilities, want 2 (instant-only and sorcery-only)", len(abs))
	}
	for i, ab := range abs {
		if len(ab.Restrictions) == 0 {
			t.Errorf("granted ability %d has no spend restriction", i)
		}
	}

	// Draw ability is condition-gated on hand size >= 7.
	me.Hand.Cards = me.Hand.Cards[:0]
	if err := g.ActivateCatalogAbility(me.ID, lute, 0, game.ActivateAbilityParams{}); err == nil {
		t.Errorf("draw ability should be refused with fewer than seven cards in hand")
	}
	for len(me.Hand.Cards) < 7 {
		me.Hand.PushTop(game.NewCard("basic-filler", uuid.Nil))
	}
	if err := g.ActivateCatalogAbility(me.ID, lute, 0, game.ActivateAbilityParams{}); err != nil {
		t.Errorf("draw ability should be legal with seven cards in hand: %v", err)
	}
}
