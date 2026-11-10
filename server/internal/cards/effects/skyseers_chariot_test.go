package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// skyseers_chariot_test.go — #5382: an activation cost increase that
// reaches costs with no mana component, and Skyseer's Chariot on it.
//
// The probes are catalogued cards whose abilities print NO mana: Krenko,
// Mob Boss ({T}-only), Sol Ring ({T}: Add {C}{C}, a mana ability), Sorin,
// Lord of Innistrad (loyalty) and the Chariot's own crew.

const skyseersChariotOracle = "92d28545-4403-4c24-8547-efab4dde6a53"

// chariotNaming puts a Skyseer's Chariot under owner's control and has
// owner answer its as-enters prompt with `name`.
func chariotNaming(t *testing.T, g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	t.Helper()
	id := b21Push(g, owner, "Skyseer's Chariot", "Artifact — Vehicle", skyseersChariotOracle, 3, 3, "W")
	g.WithWriteLock(func() {
		g.QueueCardNameChoiceForEffect(owner, id, "Skyseer's Chariot — choose a nonland card name")
	})
	actNameFor(t, g, owner, name)
	return id
}

func chariotTable(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	return g, g.Seats[0], g.Seats[1]
}

func giveWhite(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.ManaPool.AddMana(game.ManaToken{Color: "W"})
	}
}

func chariotTapped(g *game.Game, id uuid.UUID) bool {
	tapped := false
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				tapped = g.Battlefield.Cards[i].Tapped
			}
		}
	})
	return tapped
}

func activationPrice(t *testing.T, g *game.Game, who, card uuid.UUID) string {
	t.Helper()
	price, err := g.PriceActivation(who, card, 0, nil)
	if err != nil {
		t.Fatalf("PriceActivation: %v", err)
	}
	return price.Total.String()
}

func chariotOf(g *game.Game, owner uuid.UUID) uuid.UUID {
	var id uuid.UUID
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			c := g.Battlefield.Cards[i]
			if c.OracleID == skyseersChariotOracle && c.Controller == owner {
				id = c.InstanceID
			}
		}
	})
	return id
}

func TestSkyseersChariotIsFull(t *testing.T) {
	spec, ok := Lookup(skyseersChariotOracle)
	if !ok {
		t.Fatal("Skyseer's Chariot is not registered")
	}
	if spec.Name != "Skyseer's Chariot" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %v, caveats %v — want Full with none", spec.Name, spec.Completeness, spec.Caveats)
	}
}

// The name is asked as the Vehicle enters, of its controller.
func TestSkyseersChariotAsksForANameAsItEnters(t *testing.T) {
	g, me, _ := chariotTable(t)
	castCatalogSpell(t, g, "Skyseer's Chariot", "Artifact — Vehicle", skyseersChariotOracle, nil)
	passPriorityAroundTable(t, g)
	c := pendingOfKind(g, game.PendingChoiceCardName)
	if c == nil || c.Chooser != me.ID {
		t.Fatalf("no card-name prompt for the caster: %+v", g.PendingChoices)
	}
}

// A {T}-only ability of the named card costs {2}: refused with nothing
// floating (the source stays untapped and nothing is paid), and paid out
// of the pool once there is {2} in it.
func TestSkyseersChariotTaxesATapOnlyAbility(t *testing.T) {
	g, me, _ := chariotTable(t)
	chariotNaming(t, g, me.ID, "Krenko, Mob Boss")
	krenko := actPushKrenko(g, me.ID)

	if got := activationPrice(t, g, me.ID, krenko); got != "{2}" {
		t.Fatalf("Krenko's {T} ability prices at %q, want {2}", got)
	}
	err := g.ActivateCatalogAbility(me.ID, krenko, 0, game.ActivateAbilityParams{Strict: true})
	var short *game.InsufficientManaError
	if !errors.As(err, &short) {
		t.Fatalf("activating the taxed {T} ability with an empty pool: err = %v, want InsufficientManaError", err)
	}
	if chariotTapped(g, krenko) || len(g.Stack.Cards) != 0 {
		t.Fatalf("a refused activation tapped the source or used the stack (tapped %v, stack %d)", chariotTapped(g, krenko), len(g.Stack.Cards))
	}

	giveWhite(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, krenko, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activating with {2} available: %v", err)
	}
	if got := me.ManaPool.Total(); got != 1 {
		t.Errorf("pool = %d after paying the tax out of 3, want 1", got)
	}
	if !chariotTapped(g, krenko) {
		t.Error("Krenko was not tapped for its {T}")
	}
}

// A card the Chariot did not name pays nothing extra.
func TestSkyseersChariotLeavesAnUnnamedSourceAlone(t *testing.T) {
	g, me, _ := chariotTable(t)
	chariotNaming(t, g, me.ID, "Sol Ring")
	krenko := actPushKrenko(g, me.ID)
	if got := activationPrice(t, g, me.ID, krenko); got != "" {
		t.Fatalf("an unnamed {T} ability prices at %q, want no mana", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, krenko, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("an unnamed source with an empty pool: %v", err)
	}
	if row := activatedRowOf(t, g, me.ID, krenko, 0); row.ChargedManaCost != nil {
		t.Errorf("an untaxed {T}-only row stamps charged_mana_cost %q, want none", *row.ChargedManaCost)
	}
}

// "Sources with the chosen name": the opponent's copy is taxed exactly
// as the Chariot controller's own is.
func TestSkyseersChariotTaxesEveryonesSources(t *testing.T) {
	g, me, them := chariotTable(t)
	chariotNaming(t, g, me.ID, "Krenko, Mob Boss")
	mine := actPushKrenko(g, me.ID)
	theirs := actPushKrenko(g, them.ID)
	for who, id := range map[*game.Player]uuid.UUID{me: mine, them: theirs} {
		if got := activationPrice(t, g, who.ID, id); got != "{2}" {
			t.Errorf("%s's Krenko prices at %q, want {2}", who.Name, got)
		}
	}
	var short *game.InsufficientManaError
	if err := g.ActivateCatalogAbility(them.ID, theirs, 0, game.ActivateAbilityParams{Strict: true}); !errors.As(err, &short) {
		t.Errorf("the opponent's named Krenko with an empty pool: err = %v, want InsufficientManaError", err)
	}
}

// Two Chariots naming the same card add {2} each.
func TestTwoSkyseersChariotsStack(t *testing.T) {
	g, me, them := chariotTable(t)
	chariotNaming(t, g, me.ID, "Krenko, Mob Boss")
	chariotNaming(t, g, them.ID, "Krenko, Mob Boss")
	krenko := actPushKrenko(g, me.ID)
	if got := activationPrice(t, g, me.ID, krenko); got != "{4}" {
		t.Fatalf("two Chariots price Krenko at %q, want {4}", got)
	}
	if row := activatedRowOf(t, g, me.ID, krenko, 0); row.ChargedManaCost == nil || *row.ChargedManaCost != "{4}" {
		t.Errorf("the row's charged_mana_cost = %v, want {4}", row.ChargedManaCost)
	}
	giveWhite(me, 3)
	var short *game.InsufficientManaError
	if err := g.ActivateCatalogAbility(me.ID, krenko, 0, game.ActivateAbilityParams{Strict: true}); !errors.As(err, &short) {
		t.Fatalf("{3} for a {4} tax: err = %v, want InsufficientManaError", err)
	}
	if got := me.ManaPool.Total(); got != 3 {
		t.Fatalf("a refused activation spent mana: pool %d, want 3", got)
	}
	giveWhite(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, krenko, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{4} for a {4} tax: %v", err)
	}
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("pool = %d after paying {4}, want 0", got)
	}
}

// A mana ability of the named card must pay the tax too (CR 605.3),
// before its mana is made: Sol Ring cannot fund its own {2}, and with
// {2} floating it nets {C}{C} for the two spent.
func TestSkyseersChariotTaxesAManaAbility(t *testing.T) {
	g, me, them := chariotTable(t)
	chariotNaming(t, g, them.ID, "Sol Ring")
	ring := b21Push(g, me.ID, "Sol Ring", "Artifact", actSolRingOracle, 0, 0)

	var short *game.InsufficientManaError
	if err := g.ActivateManaAbility(me.ID, ring, 0, game.ManaAbilityParams{}); !errors.As(err, &short) {
		t.Fatalf("a taxed Sol Ring with an empty pool: err = %v, want InsufficientManaError", err)
	}
	if chariotTapped(g, ring) || me.ManaPool.Total() != 0 {
		t.Fatalf("a refused mana activation tapped the Ring or made mana (tapped %v, pool %d)", chariotTapped(g, ring), me.ManaPool.Total())
	}

	giveWhite(me, 2)
	if err := g.ActivateManaAbility(me.ID, ring, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("a taxed Sol Ring with {2} floating: %v", err)
	}
	colorless := 0
	for _, tok := range me.ManaPool {
		if tok.Color == "C" {
			colorless += max(tok.Count, 1)
		}
	}
	if me.ManaPool.Total() != 2 || colorless != 2 {
		t.Errorf("pool = %+v, want exactly {C}{C}: the two white paid the tax", me.ManaPool)
	}

	// The view's mana row shows the charge.
	activatedRowOf(t, g, me.ID, chariotOf(g, them.ID), 0) // stamps KnownBy on the board
	v := protocol.ViewOfGameFor(g, me.ID.String())
	found := false
	for _, cv := range v.Battlefield.Cards {
		if cv.InstanceID != ring.String() {
			continue
		}
		for _, row := range cv.ManaAbilities {
			found = true
			if row.ChargedManaCost == nil || *row.ChargedManaCost != "{2}" {
				t.Errorf("Sol Ring's mana row charged_mana_cost = %v, want {2}", row.ChargedManaCost)
			}
		}
	}
	if !found {
		t.Error("no mana row on Sol Ring in the view")
	}
}

// The auto-tapper pays the tax from the activator's other sources, and
// never with a taxed mana source it would have to pay to use.
func TestSkyseersChariotTaxIsAutoTapped(t *testing.T) {
	g, me, _ := chariotTable(t)
	chariotNaming(t, g, me.ID, "Krenko, Mob Boss")
	krenko := actPushKrenko(g, me.ID)
	forests := []uuid.UUID{pushLand(g, me.ID, "Forest"), pushLand(g, me.ID, "Forest")}
	if err := g.ActivateCatalogAbility(me.ID, krenko, 0, game.ActivateAbilityParams{AutoTap: true}); err != nil {
		t.Fatalf("auto-tapping the {2} tax off two Forests: %v", err)
	}
	for _, f := range forests {
		if !chariotTapped(g, f) {
			t.Errorf("Forest %s untapped: the auto-tapper did not pay the tax", f)
		}
	}
}

// The legal-move enumerator offers the taxed activation only when it can
// be paid, which is the gate the bot seat plays from.
func TestSkyseersChariotTaxGatesTheLegalMoves(t *testing.T) {
	g, me, _ := chariotTable(t)
	chariotNaming(t, g, me.ID, "Krenko, Mob Boss")
	krenko := actPushKrenko(g, me.ID)
	offered := func() bool {
		for _, m := range legal.EnumerateFor(g, me.ID) {
			if m.Type == legal.TypeActivateAbility && m.Source == krenko {
				return true
			}
		}
		return false
	}
	if offered() {
		t.Fatal("the taxed Krenko activation is offered with no way to pay {2}")
	}
	pushLand(g, me.ID, "Forest")
	pushLand(g, me.ID, "Forest")
	if !offered() {
		t.Fatal("the taxed Krenko activation is not offered with two untapped Forests")
	}
}

// Crew is an activated ability too: the Chariot may name itself.
func TestSkyseersChariotTaxesCrew(t *testing.T) {
	g, me, _ := chariotTable(t)
	chariot := chariotNaming(t, g, me.ID, "Skyseer's Chariot")
	bear := b21Push(g, me.ID, "Grizzly Bears", "Creature — Bear", "", 2, 2)
	params := game.ActivateAbilityParams{CrewIDs: []uuid.UUID{bear}, Strict: true}
	var short *game.InsufficientManaError
	if err := g.ActivateCatalogAbility(me.ID, chariot, 0, params); !errors.As(err, &short) {
		t.Fatalf("crewing a self-named Chariot with an empty pool: err = %v, want InsufficientManaError", err)
	}
	if chariotTapped(g, bear) {
		t.Fatal("a refused crew tapped the crewing creature")
	}
	giveWhite(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, chariot, 0, params); err != nil {
		t.Fatalf("crewing with {2} floating: %v", err)
	}
	if !chariotTapped(g, bear) || me.ManaPool.Total() != 0 {
		t.Errorf("after crewing: bear tapped %v, pool %d — want tapped and {2} spent", chariotTapped(g, bear), me.ManaPool.Total())
	}
}

// A loyalty ability of a named planeswalker owes the {2} beside its
// loyalty cost (CR 606.4).
func TestSkyseersChariotTaxesALoyaltyAbility(t *testing.T) {
	g, me, sorin := nghdNamedWalkerTable(t, func(t *testing.T, g *game.Game, owner uuid.UUID, name string) uuid.UUID {
		return chariotNaming(t, g, owner, name)
	}, "Sorin, Lord of Innistrad")
	var short *game.InsufficientManaError
	if err := g.ActivateCatalogAbility(me.ID, sorin, 0, game.ActivateAbilityParams{Strict: true}); !errors.As(err, &short) {
		t.Fatalf("a named walker's +1 with an empty pool: err = %v, want InsufficientManaError", err)
	}
	if got := loyaltyOf(g, sorin); got != 3 {
		t.Fatalf("loyalty = %d after a refused activation, want 3", got)
	}
	giveWhite(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, sorin, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("a named walker's +1 with {2} floating: %v", err)
	}
	if got := loyaltyOf(g, sorin); got != 4 || me.ManaPool.Total() != 0 {
		t.Errorf("loyalty %d, pool %d — want 4 and the {2} spent", got, me.ManaPool.Total())
	}
}

// A name that could only belong to a land taxes nothing: "a nonland
// card name" is kept where the name is read.
func TestSkyseersChariotDoesNotTaxALandName(t *testing.T) {
	g, me, _ := chariotTable(t)
	chariotNaming(t, g, me.ID, "Forest")
	forest := pushLand(g, me.ID, "Forest")
	if err := g.ActivateManaAbility(me.ID, forest, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("a Forest under a Chariot naming Forest: %v", err)
	}
}

// The chosen name is per instance state on the permanent, and it
// survives a restore: the restored table taxes the named card the same.
func TestSkyseersChariotNameSurvivesASnapshot(t *testing.T) {
	g, me, _ := chariotTable(t)
	chariot := chariotNaming(t, g, me.ID, "Krenko, Mob Boss")
	krenko := actPushKrenko(g, me.ID)
	restored := restoreRoundTrip(t, g, true)
	name := ""
	restored.ReadSnapshot(func() {
		for i := range restored.Battlefield.Cards {
			if restored.Battlefield.Cards[i].InstanceID == chariot {
				name = restored.Battlefield.Cards[i].ChosenName
			}
		}
	})
	if name != "Krenko, Mob Boss" {
		t.Fatalf("restored Chariot's chosen name = %q, want Krenko, Mob Boss", name)
	}
	if got := activationPrice(t, restored, me.ID, krenko); got != "{2}" {
		t.Errorf("after a restore Krenko prices at %q, want {2}", got)
	}
}
