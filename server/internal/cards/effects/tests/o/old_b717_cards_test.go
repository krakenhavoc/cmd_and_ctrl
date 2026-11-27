package testso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	b717OpalChampion     = "0d3617c3-1526-40e6-bc90-6139b5d3e1f1"
	b717AuraFlux         = "128304af-4555-4276-bf26-638017beb3ff"
	b717LurkingSkirge    = "21e90420-aedb-4a84-ba57-d5eae497b09b"
	b717HiddenGibbons    = "2315b044-bfd2-40dd-a73e-4cf064b36999"
	b717BrinkOfMadness   = "3209ad08-5482-4dd4-a02e-1dbe99ce6b54"
	b717ViashinoHeretic  = "36896550-151d-472c-bfc1-1d9bbf7cc04d"
	b717TreefolkMystic   = "68a6fd00-9cb0-4518-97aa-b7f475da58c1"
	b717PlanarCollapse   = "82bccc9f-f700-4f1f-b89e-bea63d973161"
	b717ImpendingDisast  = "cf0799f3-ef64-4fcc-924a-5265e52c9727"
	b717DevoutHarpist    = "f57901c2-80c6-49d6-8de3-c96aaeb72bbf"
	b717UnrelatedOracleA = "b717-unrelated-a"
)

func b717Enchantment(g *game.Game, who *game.Player, name, oracle string) uuid.UUID {
	c := effectstest.StCard(name, oracle, "Enchantment", 0, 0)
	c.Colors = []string{b717Colour[oracle]}
	return effectstest.ApaPush(g, who.ID, who.ID, c)
}

var b717Colour = map[string]string{
	b717OpalChampion: "W", b717AuraFlux: "U", b717LurkingSkirge: "B", b717HiddenGibbons: "G",
	b717BrinkOfMadness: "B", b717PlanarCollapse: "W", b717ImpendingDisast: "R",
}

// b717AtMyNextUpkeep walks to seat 0's next upkeep: through seat 1's, where
// `setup` runs (hands and the like), because the game opens already inside
// seat 0's first upkeep, whose beginning has passed.
func b717AtMyNextUpkeep(t *testing.T, g *game.Game, setup func()) {
	t.Helper()
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if setup != nil {
		setup()
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
}

func b717WantCreature(t *testing.T, g *game.Game, id uuid.UUID, sub string, p, tough int) *game.Card {
	t.Helper()
	c := effectstest.ApaLive(g, id)
	if c == nil {
		t.Fatal("the permanent is gone")
	}
	if !c.IsCreature() || c.HasCardType("enchantment") || !c.HasSubtype(sub) || c.CurrentPower() != p || c.CurrentToughness() != tough {
		t.Fatalf("creature %v enchantment %v subtype %v %d/%d; want a %d/%d %s creature that is not an enchantment",
			c.IsCreature(), c.HasCardType("enchantment"), c.HasSubtype(sub), c.CurrentPower(), c.CurrentToughness(), p, tough, sub)
	}
	return c
}

func b717WantEnchantment(t *testing.T, g *game.Game, id uuid.UUID, why string) {
	t.Helper()
	c := effectstest.ApaLive(g, id)
	if c == nil || !c.HasCardType("enchantment") || c.IsCreature() {
		t.Fatalf("%s: it is not still an enchantment", why)
	}
}

func TestB717OpalChampionBecomesAFirstStrikeKnightWhenAnOpponentCastsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, owner := g.Seats[0], g.Seats[1]
	champ := b717Enchantment(g, owner, "Opal Champion", b717OpalChampion)
	effectstest.B18CastFromHand(t, g, caster, "Opt", "Instant", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	b717WantEnchantment(t, g, champ, "an instant")
	effectstest.B18CastFromHand(t, g, caster, "Bear", "Creature — Bear", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	c := b717WantCreature(t, g, champ, "Knight", 3, 3)
	if !game.HasKeyword(c, "first strike") {
		t.Error("it has no first strike")
	}
	if !c.HasColor("W") {
		t.Error("it is no longer white")
	}
}

func TestB717OpalChampionIgnoresItsControllersCreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	champ := b717Enchantment(g, me, "Opal Champion", b717OpalChampion)
	effectstest.B18CastFromHand(t, g, me, "Bear", "Creature — Bear", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	b717WantEnchantment(t, g, champ, "its controller's creature spell")
}

func TestB717HiddenGibbonsBecomesAnApeWhenAnOpponentCastsAnInstant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, owner := g.Seats[0], g.Seats[1]
	gibbons := b717Enchantment(g, owner, "Hidden Gibbons", b717HiddenGibbons)
	effectstest.B18CastFromHand(t, g, caster, "Bear", "Creature — Bear", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	b717WantEnchantment(t, g, gibbons, "a creature spell")
	effectstest.B18CastFromHand(t, g, caster, "Opt", "Instant", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	b717WantCreature(t, g, gibbons, "Ape", 4, 4)
}

func TestB717LurkingSkirgeBecomesAFlyingImpWhenACreatureGoesToAnOpponentsGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	skirge := b717Enchantment(g, me, "Lurking Skirge", b717LurkingSkirge)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.B18Kill(t, g, mine)
	b717WantEnchantment(t, g, skirge, "its controller's creature died")
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.B18Kill(t, g, theirs)
	c := b717WantCreature(t, g, skirge, "Imp", 3, 2)
	if !c.HasSubtype("Phyrexian") || !game.HasKeyword(c, "flying") || !c.HasColor("B") {
		t.Errorf("it is %v, want a black Phyrexian Imp with flying", c.Effective().Subtypes)
	}
}

func TestB717LurkingSkirgeWatchesGraveyardsNotControl(t *testing.T) {
	// A creature the Skirge's controller controls but an opponent owns goes to
	// that opponent's graveyard.
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	skirge := b717Enchantment(g, me, "Lurking Skirge", b717LurkingSkirge)
	stolen := effectstest.ApaPush(g, opp.ID, me.ID, effectstest.ApaCreature("Stolen", b717UnrelatedOracleA, 2, 2))
	effectstest.B18Kill(t, g, stolen)
	b717WantCreature(t, g, skirge, "Imp", 3, 2)
}

func TestB717PlanarCollapseNeedsFourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	collapse := b717Enchantment(g, me, "Planar Collapse", b717PlanarCollapse)
	bears := []uuid.UUID{
		effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2),
		effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2),
		effectstest.B12Creature(g, me.ID, "Ox", "Creature — Ox", 2, 2),
	}
	b717AtMyNextUpkeep(t, g, nil)
	if !effects.OnBattlefield(g, collapse) || !effects.OnBattlefield(g, bears[0]) {
		t.Fatal("it triggered with only three creatures")
	}
}

func TestB717PlanarCollapseDestroysEveryCreatureWithFourOnTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	collapse := b717Enchantment(g, me, "Planar Collapse", b717PlanarCollapse)
	var ids []uuid.UUID
	for i := 0; i < 3; i++ {
		ids = append(ids, effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2))
	}
	ids = append(ids, effectstest.B12Creature(g, me.ID, "Ox", "Creature — Ox", 2, 2))
	b717AtMyNextUpkeep(t, g, nil)
	if effects.OnBattlefield(g, collapse) {
		t.Error("Planar Collapse was not sacrificed")
	}
	for _, id := range ids {
		if effects.OnBattlefield(g, id) {
			t.Error("a creature survived")
		}
	}
}

func TestB717PlanarCollapseCreaturesCantRegenerate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b717Enchantment(g, me, "Planar Collapse", b717PlanarCollapse)
	var shielded uuid.UUID
	for i := 0; i < 4; i++ {
		shielded = effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	}
	effectstest.P1Regenerate(t, g, shielded)
	b717AtMyNextUpkeep(t, g, nil)
	if effects.OnBattlefield(g, shielded) {
		t.Error("a regeneration shield saved a creature from Planar Collapse")
	}
}

func TestB717ImpendingDisasterNeedsSevenLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	disaster := b717Enchantment(g, me, "Impending Disaster", b717ImpendingDisast)
	var land uuid.UUID
	for i := 0; i < 6; i++ {
		land = effectstest.B12Permanent(g, opp.ID, "Mountain", "Basic Land — Mountain")
	}
	b717AtMyNextUpkeep(t, g, nil)
	if !effects.OnBattlefield(g, disaster) || !effects.OnBattlefield(g, land) {
		t.Fatal("it triggered with only six lands")
	}
}

func TestB717ImpendingDisasterDestroysEveryLandWithSevenOnTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	disaster := b717Enchantment(g, me, "Impending Disaster", b717ImpendingDisast)
	var lands []uuid.UUID
	for i := 0; i < 6; i++ {
		lands = append(lands, effectstest.B12Permanent(g, opp.ID, "Mountain", "Basic Land — Mountain"))
	}
	lands = append(lands, effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest"))
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	b717AtMyNextUpkeep(t, g, nil)
	if effects.OnBattlefield(g, disaster) {
		t.Error("Impending Disaster was not sacrificed")
	}
	for _, id := range lands {
		if effects.OnBattlefield(g, id) {
			t.Error("a land survived")
		}
	}
	if !effects.OnBattlefield(g, bear) {
		t.Error("a creature was destroyed")
	}
}

func TestB717BrinkOfMadnessNeedsAnEmptyHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	brink := b717Enchantment(g, me, "Brink of Madness", b717BrinkOfMadness)
	b717AtMyNextUpkeep(t, g, func() { effectstest.D2dSetHand(me, 1) })
	if !effects.OnBattlefield(g, brink) {
		t.Fatal("it triggered with a card in hand")
	}
	for _, s := range g.Seats[1:] {
		if effectstest.HandCount(g, s) == 0 {
			t.Errorf("%s discarded their hand", s.Name)
		}
	}
}

func TestB717BrinkOfMadnessSacrificesItselfAndEmptiesAnOpponentsHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	brink := b717Enchantment(g, me, "Brink of Madness", b717BrinkOfMadness)
	b717AtMyNextUpkeep(t, g, func() {
		effectstest.D2dSetHand(me, 0)
		for _, s := range g.Seats[1:] {
			effectstest.D2dSetHand(s, 3)
		}
	})
	target := effectstest.PickTargetPrompt(t, g)
	if len(target.PickTargetPlayers) != 3 {
		t.Fatalf("offered %d players, want the 3 opponents", len(target.PickTargetPlayers))
	}
	picked := target.PickTargetPlayers[len(target.PickTargetPlayers)-1]
	if err := g.ResolvePickTarget(target.ID, target.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: picked}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, brink) {
		t.Error("Brink of Madness was not sacrificed")
	}
	if p := g.PlayerByIDForEffect(picked); p == nil || effectstest.HandCount(g, p) != 0 {
		t.Error("the chosen opponent kept cards in hand")
	}
	emptied := 0
	for _, s := range g.Seats[1:] {
		if effectstest.HandCount(g, s) == 0 {
			emptied++
		}
	}
	if emptied != 1 {
		t.Errorf("%d opponents have an empty hand, want exactly the targeted one", emptied)
	}
}

func TestB717AuraFluxTaxesEveryOtherEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flux := b717Enchantment(g, me, "Aura Flux", b717AuraFlux)
	mine := b717Enchantment(g, me, "Mine", b717UnrelatedOracleA)
	theirs := b717Enchantment(g, opp, "Theirs", b717UnrelatedOracleA)
	// Seat 1's upkeep: only its own enchantment is taxed.
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePayUnless, me.ID) != nil {
		t.Error("the controller's enchantment was taxed on an opponent's upkeep")
	}
	choice := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePayUnless, opp.ID)
	if choice == nil {
		t.Fatal("the opponent's enchantment was not taxed on their upkeep")
	}
	if err := g.ResolvePayUnless(choice.ID, opp.ID, false); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, theirs) {
		t.Error("the declined enchantment was not sacrificed")
	}
	// Seat 0's upkeep: its other enchantment, never Aura Flux itself.
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	choice = effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePayUnless, me.ID)
	if choice == nil {
		t.Fatal("no pay-or-sacrifice prompt for the controller's other enchantment")
	}
	if err := g.ResolvePayUnless(choice.ID, me.ID, false); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, mine) {
		t.Error("the declined enchantment was not sacrificed")
	}
	if !effects.OnBattlefield(g, flux) {
		t.Error("Aura Flux itself was taxed")
	}
}

func TestB717AuraFluxPaidEnchantmentStays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	b717Enchantment(g, me, "Aura Flux", b717AuraFlux)
	mine := b717Enchantment(g, me, "Mine", b717UnrelatedOracleA)
	b717AtMyNextUpkeep(t, g, nil)
	choice := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePayUnless, me.ID)
	if choice == nil {
		t.Fatal("no prompt")
	}
	effectstest.ApaMana(me, "C", "C")
	if err := g.ResolvePayUnless(choice.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, mine) {
		t.Error("the enchantment was sacrificed after its tax was paid")
	}
}

func TestB717ViashinoHereticDestroysAnArtifactAndHurtsItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	heretic := effectstest.B12Push(g, me.ID, "Viashino Heretic", "Creature — Lizard", b717ViashinoHeretic, 1, 3)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, heretic).SummonedThisTurn = false })
	rock := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{3}"})
	effectstest.ApaMana(me, "R", "R")
	before := opp.Life
	effectstest.P7Activate(t, g, me, heretic, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}}})
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, rock) {
		t.Error("the artifact survived")
	}
	if opp.Life != before-3 {
		t.Errorf("life %d, want %d", opp.Life, before-3)
	}
}

func TestB717ViashinoHereticRefusesANonartifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	heretic := effectstest.B12Push(g, me.ID, "Viashino Heretic", "Creature — Lizard", b717ViashinoHeretic, 1, 3)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, heretic).SummonedThisTurn = false })
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.ApaMana(me, "R", "R")
	err := effectstest.ActivateOn(t, g, me.ID, heretic, 0, bear)
	if err == nil {
		t.Fatal("a creature that is no artifact was accepted as a target")
	}
}

func b717Pacifism(t *testing.T, g *game.Game, p *game.Player, host uuid.UUID) uuid.UUID {
	t.Helper()
	id := effectstest.ApaPush(g, p.ID, p.ID, game.Card{Name: "Pacifism", TypeLine: "Enchantment — Aura", Colors: []string{"W"}})
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(id, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatal(err)
		}
	})
	return id
}

func TestB717TreefolkMysticDestroysTheAurasOnACreatureItIsBlockedBy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mystic := effectstest.B12Push(g, me.ID, "Treefolk Mystic", "Creature — Treefolk", b717TreefolkMystic, 2, 4)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 1, 9)
	other := effectstest.B12Creature(g, opp.ID, "Other", "Creature — Bear", 1, 9)
	a1, a2 := b717Pacifism(t, g, opp, blocker), b717Pacifism(t, g, me, blocker)
	a3 := b717Pacifism(t, g, opp, other)
	effectstest.BrAttack(t, g, mystic)
	if err := g.DeclareBlocker(blocker, mystic); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, a1) || effects.OnBattlefield(g, a2) {
		t.Error("an Aura on the creature that blocked it survived")
	}
	if !effects.OnBattlefield(g, a3) {
		t.Error("an Aura on a creature that did not block was destroyed")
	}
}

func TestB717TreefolkMysticDestroysTheAurasOnACreatureItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 1, 9)
	mystic := effectstest.B12Push(g, opp.ID, "Treefolk Mystic", "Creature — Treefolk", b717TreefolkMystic, 2, 4)
	aura := b717Pacifism(t, g, opp, attacker)
	effectstest.BrAttack(t, g, attacker)
	if err := g.DeclareBlocker(mystic, attacker); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, aura) {
		t.Error("an Aura on the creature it blocked survived")
	}
}

func TestB717DevoutHarpistDestroysAnAuraAttachedToACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	harpist := effectstest.B12Push(g, me.ID, "Devout Harpist", "Creature — Human", b717DevoutHarpist, 1, 1)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, harpist).SummonedThisTurn = false })
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	aura := b717Pacifism(t, g, opp, bear)
	effectstest.P7Activate(t, g, me, harpist, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: aura}}})
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, aura) {
		t.Error("the Aura survived")
	}
}

func TestB717DevoutHarpistRefusesAnAuraOnANoncreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	harpist := effectstest.B12Push(g, me.ID, "Devout Harpist", "Creature — Human", b717DevoutHarpist, 1, 1)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, harpist).SummonedThisTurn = false })
	land := effectstest.B12Permanent(g, opp.ID, "Mountain", "Basic Land — Mountain")
	aura := b717Pacifism(t, g, opp, land)
	err := effectstest.ActivateOn(t, g, me.ID, harpist, 0, aura)
	if err == nil {
		t.Fatal("an Aura on a land was accepted as a target")
	}
}
