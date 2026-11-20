package testsu

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- the cost component itself -----------------------------------

func TestAbilityCostReadsXOutOfTheManaString(t *testing.T) {
	cases := []struct {
		mana    string
		demands bool
		slots   int
	}{
		{"", false, 0},
		{"{1}{B}", false, 0},
		{"{X}", true, 1},
		{"{X}{U}", true, 1},
		{"{X}{X}", true, 2},
	}
	for _, c := range cases {
		cost := game.AbilityCost{Mana: c.mana}
		if cost.DemandsX() != c.demands {
			t.Errorf("%q: DemandsX %v, want %v", c.mana, cost.DemandsX(), c.demands)
		}
		if cost.XSlots() != c.slots {
			t.Errorf("%q: XSlots %d, want %d", c.mana, cost.XSlots(), c.slots)
		}
	}
	// A floor only means anything on a cost that actually varies.
	if got := (game.AbilityCost{Mana: "{2}", MinX: 3}).FloorX(); got != 0 {
		t.Errorf("FloorX on a fixed cost: %d, want 0", got)
	}
	if got := (game.AbilityCost{Mana: "{X}", MinX: 1}).FloorX(); got != 1 {
		t.Errorf("FloorX: %d, want 1", got)
	}
}

func TestPlusCarriesTheXFloorAndTheCrewNumber(t *testing.T) {
	cost := effects.Plus(effects.ManaCost("{X}"), effects.TapCost(), effects.MinX(1))
	if !cost.Tap || cost.Mana != "{X}" || cost.MinX != 1 {
		t.Fatalf("Plus dropped a component: %+v", cost)
	}
	if crewed := effects.Plus(effects.ManaCost("{1}"), game.AbilityCost{Crew: 3}); crewed.Crew != 3 {
		t.Errorf("Plus dropped the crew number: %+v", crewed)
	}
}

func TestAbilityWithNoXRejectsAnAnnouncedX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	krenko := effectstest.PushCatalogPermanent(g, me.ID, "Krenko, Mob Boss", "Creature — Goblin", effectstest.KrenkoOracle, false)
	// Rejected rather than ignored: a client sending an X for an
	// ability that has none is confused about which ability it is
	// firing, and swallowing the field would hide that.
	if err := g.ActivateCatalogAbility(me.ID, krenko, 0, game.ActivateAbilityParams{XValue: 2}); err != game.ErrInvalidParam {
		t.Fatalf("activate with a bogus X: %v, want ErrInvalidParam", err)
	}
	if g.Battlefield.Contains(krenko) {
		if c := effectstest.FindTestCard(g, krenko); c != nil && c.Tapped {
			t.Error("a refused announcement must not have paid the tap cost")
		}
	}
}

func TestNegativeXOnAnAbilityIsRejected(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.FillPool(me, 8)
	vault := effectstest.PushCatalogPermanent(g, me.ID, "Treasure Vault", "Artifact Land", effectstest.TreasureVaultOracle, false)
	if err := g.ActivateCatalogAbility(me.ID, vault, 0, game.ActivateAbilityParams{XValue: -1}); err != game.ErrInvalidParam {
		t.Fatalf("negative X: %v, want ErrInvalidParam", err)
	}
}

// --- Treasure Vault ----------------------------------------------

func TestTreasureVaultChargesBothXSlotsAndMakesXTreasures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	// "{X}{X}" is two slots, so X=3 is six mana, not three. Seven in
	// the pool leaves exactly one behind — which is the assertion
	// that catches a one-slot cost model.
	effectstest.FillPool(me, 7)
	vault := effectstest.PushCatalogPermanent(g, me.ID, "Treasure Vault", "Artifact Land", effectstest.TreasureVaultOracle, false)

	if err := g.ActivateCatalogAbility(me.ID, vault, 0, game.ActivateAbilityParams{
		XValue: 3,
		Strict: true,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := me.ManaPool.Total(); got != 1 {
		t.Errorf("pool after paying {X}{X} at X=3: %d, want 1", got)
	}
	// Tap and sacrifice are cost components, paid at announce.
	if g.Battlefield.Contains(vault) {
		t.Error("the sacrifice cost is paid at announce")
	}
	if effectstest.CountTreasures(g, me.ID) != 0 {
		t.Error("the Treasures wait for resolution")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountTreasures(g, me.ID); got != 3 {
		t.Errorf("Treasures: %d, want 3", got)
	}
}

func TestTreasureVaultRefusesAnXItCannotPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.FillPool(me, 5)
	vault := effectstest.PushCatalogPermanent(g, me.ID, "Treasure Vault", "Artifact Land", effectstest.TreasureVaultOracle, false)

	err := g.ActivateCatalogAbility(me.ID, vault, 0, game.ActivateAbilityParams{XValue: 3, Strict: true})
	var insufficient *game.InsufficientManaError
	if !errors.As(err, &insufficient) {
		t.Fatalf("X=3 on five mana: %v, want InsufficientManaError", err)
	}
	// Validate-all-then-pay: a refused activation leaves the board
	// exactly as it was.
	if !g.Battlefield.Contains(vault) {
		t.Error("a refused activation must not sacrifice the land")
	}
	if c := effectstest.FindTestCard(g, vault); c != nil && c.Tapped {
		t.Error("a refused activation must not tap the land")
	}
	if got := me.ManaPool.Total(); got != 5 {
		t.Errorf("pool after a refused activation: %d, want 5", got)
	}
}

func TestTreasureVaultAtXZeroIsLegalAndMakesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vault := effectstest.PushCatalogPermanent(g, me.ID, "Treasure Vault", "Artifact Land", effectstest.TreasureVaultOracle, false)
	// The card prints no floor, so zero is a legal (and pointless)
	// announcement, exactly as it is for an {X} spell.
	if err := g.ActivateCatalogAbility(me.ID, vault, 0, game.ActivateAbilityParams{XValue: 0, Strict: true}); err != nil {
		t.Fatalf("activate at X=0: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountTreasures(g, me.ID); got != 0 {
		t.Errorf("Treasures at X=0: %d, want 0", got)
	}
}

// --- Soothsaying -------------------------------------------------

func TestSoothsayingLooksAtTheTopXCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.FillPool(me, 3)
	sooth := effectstest.PushCatalogPermanent(g, me.ID, "Soothsaying", "Enchantment", effectstest.SoothsayingOracle, false)

	if err := g.ActivateCatalogAbility(me.ID, sooth, 1, game.ActivateAbilityParams{XValue: 3, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("pool after {X} at X=3: %d, want 0", got)
	}
	effectstest.PassPriorityAroundTable(t, g)

	choice := effectstest.LatestChoiceOfKind(g, game.PendingChoiceLookAtTop)
	if choice == nil {
		t.Fatal("expected a look-at-top prompt")
	}
	if len(choice.ScryCards) != 3 {
		t.Errorf("looked at %d cards, want 3", len(choice.ScryCards))
	}
	if choice.Chooser != me.ID {
		t.Errorf("chooser %v, want the activator", choice.Chooser)
	}
}

func TestSoothsayingAtXZeroLooksAtNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sooth := effectstest.PushCatalogPermanent(g, me.ID, "Soothsaying", "Enchantment", effectstest.SoothsayingOracle, false)
	if err := g.ActivateCatalogAbility(me.ID, sooth, 1, game.ActivateAbilityParams{XValue: 0, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKind(g, game.PendingChoiceLookAtTop) != nil {
		t.Error("X=0 looks at nothing, so there is nothing to reorder")
	}
}

func TestSoothsayingShufflesAndForgetsTheLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.FillPool(me, 3)
	me.ManaPool.AddMana(game.ManaToken{Color: "U"}, game.ManaToken{Color: "U"})
	sooth := effectstest.PushCatalogPermanent(g, me.ID, "Soothsaying", "Enchantment", effectstest.SoothsayingOracle, false)

	// Someone has peeked at the top of the library; the shuffle is
	// what takes that knowledge away again.
	top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID
	for i := range me.Library.Cards {
		if me.Library.Cards[i].InstanceID == top {
			me.Library.Cards[i].AddKnower(me.ID)
		}
	}
	size := me.Library.Size()

	if err := g.ActivateCatalogAbility(me.ID, sooth, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if me.Library.Size() != size {
		t.Errorf("library size %d, want %d — a shuffle moves nothing out", me.Library.Size(), size)
	}
	for _, c := range me.Library.Cards {
		if len(c.KnownBy) != 0 {
			t.Fatalf("shuffling must clear every KnownBy in the zone; %q still known", c.Name)
		}
	}
}

// --- Helm of Obedience -------------------------------------------

func TestHelmOfObedienceRefusesXZero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.FillPool(me, 5)
	helm := effectstest.PushCatalogPermanent(g, me.ID, "Helm of Obedience", "Artifact", effectstest.HelmOfObedienceOracle, false)

	// "X can't be 0" is a printed floor, so announcing under it is an
	// illegal announcement rather than a cheap one.
	err := g.ActivateCatalogAbility(me.ID, helm, 0, game.ActivateAbilityParams{
		XValue:  0,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		Strict:  true,
	})
	if err != game.ErrInvalidParam {
		t.Fatalf("X=0: %v, want ErrInvalidParam", err)
	}
	if c := effectstest.FindTestCard(g, helm); c != nil && c.Tapped {
		t.Error("a refused announcement must not have paid the tap cost")
	}
}

func TestHelmOfObedienceMillsUntilACreatureAndReanimatesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.FillPool(me, 5)
	helm := effectstest.PushCatalogPermanent(g, me.ID, "Helm of Obedience", "Artifact", effectstest.HelmOfObedienceOracle, false)

	// Top of the library is the LAST element, so push the creature
	// first and the two filler cards on top of it.
	beast := uuid.New()
	opp.Library.PushTop(game.Card{
		InstanceID: beast, Name: "Their Beast", TypeLine: "Creature — Beast",
		Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID,
	})
	opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler A", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler B", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	graveBefore := opp.Graveyard.Size()

	if err := g.ActivateCatalogAbility(me.ID, helm, 0, game.ActivateAbilityParams{
		XValue:  5,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		Strict:  true,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("pool after {X} at X=5: %d, want 0", got)
	}
	effectstest.PassPriorityAroundTable(t, g)

	// Two filler cards were milled; the run stopped AT the creature,
	// which is then taken off the graveyard onto our battlefield.
	if got := opp.Graveyard.Size() - graveBefore; got != 2 {
		t.Errorf("cards left in the graveyard: %d, want 2 (the creature is taken)", got)
	}
	if !g.Battlefield.Contains(beast) {
		t.Fatal("the milled creature should be on the battlefield")
	}
	c := effectstest.FindTestCard(g, beast)
	if c.Controller != me.ID {
		t.Errorf("controller %v, want the activator — 'under your control'", c.Controller)
	}
	if c.Owner != opp.ID {
		t.Errorf("owner %v, want the milled player", c.Owner)
	}
	if g.Battlefield.Contains(helm) {
		t.Error("finding a creature sacrifices the Helm")
	}
}

func TestHelmOfObedienceStopsAtXCardsAndKeepsItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.FillPool(me, 5)
	helm := effectstest.PushCatalogPermanent(g, me.ID, "Helm of Obedience", "Artifact", effectstest.HelmOfObedienceOracle, false)

	beast := uuid.New()
	opp.Library.PushTop(game.Card{
		InstanceID: beast, Name: "Their Beast", TypeLine: "Creature — Beast",
		Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID,
	})
	for i := 0; i < 4; i++ {
		opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	}
	graveBefore := opp.Graveyard.Size()

	// X=2 runs out before it reaches the creature: "whichever comes
	// first".
	if err := g.ActivateCatalogAbility(me.ID, helm, 0, game.ActivateAbilityParams{
		XValue:  2,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		Strict:  true,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := opp.Graveyard.Size() - graveBefore; got != 2 {
		t.Errorf("milled %d, want exactly X=2", got)
	}
	if g.Battlefield.Contains(beast) {
		t.Error("no creature was milled, so nothing is reanimated")
	}
	if !g.Battlefield.Contains(helm) {
		t.Error("the sacrifice is conditional on finding a creature — the Helm stays")
	}
}

func TestHelmOfObedienceLocksXOntoTheStackItem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.FillPool(me, 5)
	helm := effectstest.PushCatalogPermanent(g, me.ID, "Helm of Obedience", "Artifact", effectstest.HelmOfObedienceOracle, false)

	if err := g.ActivateCatalogAbility(me.ID, helm, 0, game.ActivateAbilityParams{
		XValue:  4,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		Strict:  true,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	// CR 602.2b: X is announced with the activation and cannot
	// change afterwards — the stack item is where it is locked, and
	// it is what the effect reads back.
	var found *game.StackItem
	for _, item := range g.StackMeta {
		if item.SourceCardID == helm {
			found = item
		}
	}
	if found == nil {
		t.Fatal("expected the ability on the stack")
	}
	if found.XValue != 4 {
		t.Errorf("StackItem.XValue %d, want 4", found.XValue)
	}
}
