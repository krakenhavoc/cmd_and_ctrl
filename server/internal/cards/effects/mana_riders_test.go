package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// mana_riders_test.go — the S22 mana-ability-rider batch: the eight
// painlands, the six Talismans, Ancient Tomb, Mana Confluence and
// City of Brass.
//
// The distinction every test here is aimed at is COST vs RIDER,
// because getting it backwards is invisible in the happy path and
// wrong in exactly the direction this project refuses to ship:
//
//   - A rider (painlands, Ancient Tomb) is part of the ability's
//     effect. The source taps at any life total, the mana arrives,
//     and the damage happens whether the player likes it or not.
//   - A cost (Mana Confluence) is validated first. Too little life
//     and the activation is REJECTED with the land still untapped.
//
// A third shape, City of Brass, looks like a rider and isn't: its
// "whenever this land becomes tapped" clause is a real CR 603
// trigger, so it uses the stack and fires on taps that were never
// activations. Folding it into the mana ability would have made the
// card better than printed.

const (
	shivanReefOracle          = "0fe16212-66c3-4e45-a641-7391e9b2e304"
	adarkarWastesOracle       = "d5ad26cc-2bdb-46b7-b8bf-dd099d5fa09b"
	ancientTombOracle         = "23467047-6dba-4498-b783-1ebc4f74b8c2"
	manaConfluenceOracle      = "d0ee5bdc-2b69-4b73-9a20-ffcc18783b29"
	cityOfBrassOracle         = "f25351e3-539b-4bbc-b92d-6480acf4d722"
	talismanOfDominanceOracle = "4c0a0448-b9d6-43a0-8549-64066dac63f0"
)

// riderLatestManaPick returns the most recent PendingChoiceMana queued for
// a chooser, or nil.
func riderLatestManaPick(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	var out *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceMana && c.Chooser == chooser {
			out = c
		}
	}
	return out
}

// riderAnswerManaPicks answers every colour pick owed by chooser with
// `color` and returns how many it answered. #730 made an unanswered
// mana_pick gate the table, so a test that taps an any-colour source
// and then passes priority or walks the cursor on has to drain them
// first — which is what a player does anyway, since until the pick is
// answered the mana is not in the pool.
func riderAnswerManaPicks(t *testing.T, g *game.Game, chooser uuid.UUID, color string) int {
	t.Helper()
	n := 0
	for i := 0; i < 16; i++ {
		pick := riderLatestManaPick(g, chooser)
		if pick == nil {
			return n
		}
		if err := g.ResolveManaChoice(pick.ID, chooser, color); err != nil {
			t.Fatalf("ResolveManaChoice: %v", err)
		}
		n++
	}
	t.Fatal("colour picks did not drain")
	return n
}

// riderGiveCommander drops a commander with the supplied mana cost into a
// seat's command zone, which is what game.commanderIdentityFor reads
// to narrow a pipe ability. Used to prove the painland duals do NOT
// narrow.
func riderGiveCommander(g *game.Game, p *game.Player, manaCost string) {
	g.WithWriteLock(func() {
		p.Command.PushTop(game.Card{
			InstanceID:  uuid.New(),
			Name:        "Test Commander",
			TypeLine:    "Legendary Creature — Human",
			ManaCost:    manaCost,
			Owner:       p.ID,
			Controller:  p.ID,
			IsCommander: true,
		})
	})
}

// --- painlands ---------------------------------------------------

// The painless half is ability 0, and it is genuinely painless. If
// this ever regresses to "one ability with a rider", a painland
// becomes a strictly worse card and the auto-tapper stops being able
// to touch it.
func TestPainlandColorlessHalfCostsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Life
	land := seedPermanentWithOracle(g, me.ID, "Shivan Reef", "Land", shivanReefOracle)

	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(me.ManaPool) != 1 || me.ManaPool[0].Color != "C" {
		t.Errorf("pool = %v, want one {C}", me.ManaPool)
	}
	if me.Life != before {
		t.Errorf("life %d → %d; the colorless half of a painland is free", before, me.Life)
	}
}

// The colored half pays a point, and pays it as DAMAGE from the land
// — a source-reading effect has to see the land, not the player.
func TestPainlandColoredHalfDealsOneDamage(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Life
	land := seedPermanentWithOracle(g, me.ID, "Shivan Reef", "Land", shivanReefOracle)

	if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.Life != before-1 {
		t.Errorf("life %d → %d, want %d", before, me.Life, before-1)
	}
	var dmg *game.Event
	for i := range g.Events {
		if g.Events[i].Kind == game.EventDealDamage && g.Events[i].Target == me.ID {
			dmg = &g.Events[i]
		}
	}
	if dmg == nil {
		t.Fatal("no damage event; the rider must deal DAMAGE, not lose life — " +
			"prevention and doubling effects only see damage")
	}
	if dmg.Source != land {
		t.Errorf("damage source %v, want the land %v", dmg.Source, land)
	}
	// The colour still has to be picked; the damage did not wait for it.
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil {
		t.Fatal("the colored half is a two-colour pipe and must ask which colour")
	}
	if len(pick.ColorOptions) != 2 {
		t.Errorf("colour options %v, want exactly U and R", pick.ColorOptions)
	}
}

// A painland's colours are printed on the card, not derived from the
// command zone. Arcane Signet's narrowing must not leak onto it —
// a mono-blue deck's Shivan Reef still offers {R}.
func TestPainlandDualIgnoresCommanderIdentity(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	riderGiveCommander(g, me, "{W}")
	land := seedPermanentWithOracle(g, me.ID, "Adarkar Wastes", "Land", adarkarWastesOracle)

	if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil {
		t.Fatal("no mana pick")
	}
	if len(pick.ColorOptions) != 2 {
		t.Fatalf("colour options %v under a mono-white commander, want both W and U", pick.ColorOptions)
	}
}

// Contrast case: Command Tower prints "in your commander's color
// identity" and still narrows. This is what keeps the painland test
// honest — it proves NarrowToCommanderIdentity is doing the work
// rather than the narrowing having quietly died.
func TestCommandTowerStillNarrowsToCommanderIdentity(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	riderGiveCommander(g, me, "{W}")
	tower := seedPermanentWithOracle(g, me.ID, "Command Tower", "Land", "0895c9b7-ae7d-4bb3-af17-3b75deb50a25")

	if err := g.ActivateManaAbility(me.ID, tower, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil {
		t.Fatal("no mana pick")
	}
	if len(pick.ColorOptions) != 1 || pick.ColorOptions[0] != "W" {
		t.Errorf("Command Tower options %v under a mono-white commander, want just W", pick.ColorOptions)
	}
}

// A rider is not a cost: a painland at 1 life taps, produces, and
// kills its controller. The state-based-action pass has to run on the
// way out of the activation rather than at some later boundary.
func TestPainlandKillsYouAtOneLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Life = 1 })
	land := seedPermanentWithOracle(g, me.ID, "Shivan Reef", "Land", shivanReefOracle)

	if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility at 1 life must succeed — the damage is a rider, not a cost: %v", err)
	}
	if me.Life > 0 {
		t.Errorf("life = %d, want 0 or less", me.Life)
	}
	if !me.Eliminated {
		t.Error("the player survived a lethal painland activation; SBAs did not run")
	}
}

// --- Talismans ---------------------------------------------------

func TestTalismanHasBothHalves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Life
	rock := seedPermanentWithOracle(g, me.ID, "Talisman of Dominance", "Artifact", talismanOfDominanceOracle)

	if err := g.ActivateManaAbility(me.ID, rock, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("colorless half: %v", err)
	}
	if me.Life != before {
		t.Errorf("the colorless half cost %d life", before-me.Life)
	}
	// Untap it the way an untap step would, then take the painful line.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == rock {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	if err := g.ActivateManaAbility(me.ID, rock, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("colored half: %v", err)
	}
	if me.Life != before-1 {
		t.Errorf("life %d → %d, want %d", before, me.Life, before-1)
	}
}

// --- Ancient Tomb ------------------------------------------------

func TestAncientTombAddsTwoAndDealsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Life
	land := seedPermanentWithOracle(g, me.ID, "Ancient Tomb", "Land", ancientTombOracle)

	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(me.ManaPool) != 2 {
		t.Errorf("pool = %v, want two colorless", me.ManaPool)
	}
	if me.Life != before-2 {
		t.Errorf("life %d → %d, want %d", before, me.Life, before-2)
	}
}

// --- Mana Confluence: a COST, not a rider -------------------------

func TestManaConfluencePaysOneLifeAsACost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Life
	land := seedPermanentWithOracle(g, me.ID, "Mana Confluence", "Land", manaConfluenceOracle)

	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.Life != before-1 {
		t.Errorf("life %d → %d, want %d", before, me.Life, before-1)
	}
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil {
		t.Fatal("no mana pick for 'one mana of any color'")
	}
	if len(pick.ColorOptions) != 5 {
		t.Errorf("colour options %v, want all five — the card says 'any color'", pick.ColorOptions)
	}
}

// The whole reason ManaAbilityCost.Life exists. At 0 life the cost
// cannot be paid, so the activation is rejected AND — the part that
// would be a real bug — the land does not tap. A half-paid cost that
// stranded the permanent would be strictly worse than no card at all.
func TestManaConfluenceRefusesAtZeroLifeWithoutTapping(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := seedPermanentWithOracle(g, me.ID, "Mana Confluence", "Land", manaConfluenceOracle)
	g.WithWriteLock(func() { me.Life = 0 })

	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("activated Mana Confluence at 0 life; CR 118.8 forbids paying life you don't have")
	}
	card, ok := battlefieldCard(g, land)
	if !ok {
		t.Fatal("the land left the battlefield")
	}
	if card.Tapped {
		t.Error("the rejected activation tapped the land anyway — costs must be validated before any are paid")
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool = %v after a rejected activation, want empty", me.ManaPool)
	}
}

// Paying down to exactly 0 is legal (CR 118.8) and then you lose to
// SBAs. The activation itself must succeed.
func TestManaConfluenceMayPayDownToZero(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := seedPermanentWithOracle(g, me.ID, "Mana Confluence", "Land", manaConfluenceOracle)
	g.WithWriteLock(func() { me.Life = 1 })

	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("paying life down to exactly 0 is legal: %v", err)
	}
	if me.Life != 0 {
		t.Errorf("life = %d, want 0", me.Life)
	}
	if !me.Eliminated {
		t.Error("a player at 0 life survived the activation; SBAs did not run")
	}
}

// The client needs the cost to render it. A life cost that never
// reaches the wire is a cost the player is surprised by.
func TestManaConfluenceLifeCostReachesTheWire(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := seedPermanentWithOracle(g, me.ID, "Mana Confluence", "Land", manaConfluenceOracle)
	knownToTable(g, land)

	view := protocol.ViewOfGameFor(g, me.ID.String())
	var found *protocol.ManaAbilityView
	for i := range view.Battlefield.Cards {
		c := view.Battlefield.Cards[i]
		if c.InstanceID == land.String() && len(c.ManaAbilities) > 0 {
			found = &c.ManaAbilities[0]
		}
	}
	if found == nil {
		t.Fatal("Mana Confluence has no mana ability on the wire")
	}
	if found.LifeCost != 1 {
		t.Errorf("life_cost = %d, want 1", found.LifeCost)
	}
}

// --- City of Brass: a TRIGGER, not a rider ------------------------

// Tapping for mana fires the trigger, and it goes on the STACK rather
// than resolving inline — players get a response window, exactly as
// printed.
func TestCityOfBrassTriggersOnTappingForMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Life
	land := seedPermanentWithOracle(g, me.ID, "City of Brass", "Land", cityOfBrassOracle)

	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	// The mana ability itself is painless — the damage is still on
	// its way to the stack.
	if me.Life != before {
		t.Errorf("life changed to %d before the trigger resolved; the damage must use the stack", me.Life)
	}
	if len(g.PendingTriggers) == 0 && triggerOnStack(g, land) == nil {
		t.Fatal("tapping City of Brass queued no trigger")
	}
	riderAnswerManaPicks(t, g, me.ID, "W")
	passPriorityAroundTable(t, g)
	if me.Life != before-1 {
		t.Errorf("life %d → %d after the trigger resolved, want %d", before, me.Life, before-1)
	}
}

// The clause says "becomes tapped", not "is tapped for mana". A tap
// from anywhere fires it. Modelling this as a mana-ability rider
// would have made the card strictly better than printed, which is the
// one direction this project does not ship.
func TestCityOfBrassTriggersOnAnyTapNotJustManaTaps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Life
	land := seedPermanentWithOracle(g, me.ID, "City of Brass", "Land", cityOfBrassOracle)

	// Tap it by a route that is not an activation at all.
	if err := g.TapCard(land, true); err != nil {
		t.Fatalf("TapCard: %v", err)
	}
	if len(g.PendingTriggers) == 0 && triggerOnStack(g, land) == nil {
		t.Fatal("a non-mana tap did not fire City of Brass")
	}
	passPriorityAroundTable(t, g)
	if me.Life != before-1 {
		t.Errorf("life %d → %d, want %d", before, me.Life, before-1)
	}
}

// --- auto-tapper interaction --------------------------------------

// The auto-tapper may plan a painland, and when it does it reaches
// for the painless half. A planner that silently spent life to save a
// click would be a planner nobody should trust.
func TestAutoTapUsesThePainlessHalfOfAPainland(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Life
	seedPermanentWithOracle(g, me.ID, "Shivan Reef", "Land", shivanReefOracle)

	cost, err := game.ParseCost("{1}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	plan, ok := g.AutoTapForCost(me.ID, cost, 0)
	if !ok || len(plan) != 1 {
		t.Fatalf("auto-tap plan = %v ok=%v, want the painland planned as a colorless source", plan, ok)
	}
	if me.Life != before {
		t.Errorf("planning alone cost %d life", before-me.Life)
	}
}

// Ancient Tomb has no painless half, so the auto-tapper declines to
// plan it at all rather than taking 2 off the player's life
// uninvited. It stays fully activatable by hand.
func TestAutoTapDeclinesAncientTomb(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedPermanentWithOracle(g, me.ID, "Ancient Tomb", "Land", ancientTombOracle)

	cost, err := game.ParseCost("{1}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	if plan, ok := g.AutoTapForCost(me.ID, cost, 0); ok {
		t.Errorf("auto-tap planned Ancient Tomb (%v); its only ability deals 2 damage", plan)
	}
}

// Mana Confluence is declined for the same reason, one step stronger:
// its life payment is a cost, and spending a cost without being asked
// is worse than failing to plan.
func TestAutoTapDeclinesManaConfluence(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedPermanentWithOracle(g, me.ID, "Mana Confluence", "Land", manaConfluenceOracle)

	cost, err := game.ParseCost("{1}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	if plan, ok := g.AutoTapForCost(me.ID, cost, 0); ok {
		t.Errorf("auto-tap planned Mana Confluence (%v); its cost includes 1 life", plan)
	}
}

// --- registration --------------------------------------------------

// Both cycles are loops over tables, so a transposed oracle ID in one
// row is invisible until somebody plays that exact card. Pin every
// row, and pin that each one really produces something — the whole
// point of this batch is that a painland used to be a land with no
// mana ability at all.
func TestEveryRiderCardIsRegisteredAndProduces(t *testing.T) {
	want := map[string]string{
		// Painlands.
		"0fe16212-66c3-4e45-a641-7391e9b2e304": "Shivan Reef",
		"6b75b94e-83b7-457e-ac41-7ca90b5a59aa": "Battlefield Forge",
		"33de01e9-ce5a-42d4-afcb-343cd54a6d80": "Caves of Koilos",
		"40b36bc6-c185-4bda-99e7-0118953c2c97": "Yavimaya Coast",
		"32116127-cf96-4a1b-8896-a1ebc087b597": "Llanowar Wastes",
		"857febd9-cdd7-4f8e-a852-d88084b0cfbc": "Underground River",
		"d5ad26cc-2bdb-46b7-b8bf-dd099d5fa09b": "Adarkar Wastes",
		"f5c38c01-4a40-469f-91a0-7479daf4e8e7": "Sulfurous Springs",
		// Talismans.
		"4c0a0448-b9d6-43a0-8549-64066dac63f0": "Talisman of Dominance",
		"14d2979d-5728-42d7-a027-0eb1f754655d": "Talisman of Creativity",
		"1d9aeaaa-66f6-41cb-9bac-162d6fd8662c": "Talisman of Indulgence",
		"b693c3de-2eaf-4850-b405-e79d00adefda": "Talisman of Hierarchy",
		"00e35322-1a9a-41e3-9ce1-359c8eaa3bc7": "Talisman of Progress",
		"6c326439-5620-4ec6-a56a-fe9c3d5d2a46": "Talisman of Conviction",
		// Singles.
		"23467047-6dba-4498-b783-1ebc4f74b8c2": "Ancient Tomb",
		"d0ee5bdc-2b69-4b73-9a20-ffcc18783b29": "Mana Confluence",
		"f25351e3-539b-4bbc-b92d-6480acf4d722": "City of Brass",
	}
	if len(want) != 17 {
		t.Fatalf("the batch is 17 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
		if len(spec.ManaAbilities) == 0 {
			t.Errorf("%s has no mana ability — which is exactly the bug this batch exists to fix", name)
		}
	}
}

// The two cycles' shapes, pinned: two abilities, the first painless
// and colorless, the second a two-colour pipe with a damage rider.
// The ORDER matters — game.autoTapAbilityFor takes the first
// qualifying ability, so a swapped pair would hand the auto-tapper
// the painful line.
func TestPainCyclesPutThePainlessAbilityFirst(t *testing.T) {
	for _, oracle := range []string{shivanReefOracle, adarkarWastesOracle, talismanOfDominanceOracle} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Fatalf("%s not registered", oracle)
		}
		if len(spec.ManaAbilities) != 2 {
			t.Fatalf("%s has %d mana abilities, want 2", spec.Name, len(spec.ManaAbilities))
		}
		if spec.ManaAbilities[0].Rider != nil || spec.ManaAbilities[0].Cost.Life != 0 {
			t.Errorf("%s: ability 0 must be the painless one", spec.Name)
		}
		if spec.ManaAbilities[1].Rider == nil {
			t.Errorf("%s: ability 1 must carry the damage rider", spec.Name)
		}
		if spec.ManaAbilities[1].NarrowToCommanderIdentity {
			t.Errorf("%s: the dual names two printed colours and must not narrow to the commander's identity", spec.Name)
		}
	}
}
