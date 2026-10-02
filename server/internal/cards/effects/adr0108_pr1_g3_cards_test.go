package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adr0108_pr1_g3_cards_test.go — ADR 0108 Delivery PR 1 (#1886, #1887),
// group 3: the modal "that creature" bullets, the "dealt damage this
// way" sweeps, a split card with fuse, and the activated and triggered
// "can't be regenerated" cards.

const (
	p1g3AgateOracle        = "7b89b7d2-c724-4d5d-9f0b-7d3302ad1168"
	p1g3SuplexOracle       = "2d1bbeda-2e81-4aaa-9494-094ec3dd6c3b"
	p1g3PineconeOracle     = "961e3023-39ea-4141-99b6-738280a2815d"
	p1g3ExpulsionOracle    = "71d178b3-e5ca-4576-83f7-8dce74758acd"
	p1g3GnashingOracle     = "188a3f61-3f7c-450d-8ee7-e12ad13b86c6"
	p1g3CrushOracle        = "4aa119f7-d411-4188-956e-547f7d14e789"
	p1g3UnderworldOracle   = "ed3d113d-f2c3-4ef3-8bdd-ed4824a22ab0"
	p1g3GallifreyOracle    = "d4424585-9564-4ec2-8267-3f5438e1f29e"
	p1g3FlamebreakOracle   = "9e608ace-2844-419b-9204-9054127557b2"
	p1g3SpikeOracle        = "c983644d-6741-4aa1-aa68-a6e680c26bb6"
	p1g3JayaOracle         = "cbbad3da-695e-4527-9678-0942f3751287"
	p1g3OrcishHealerOracle = "9b7a0c9d-1045-4fc0-b5e3-18616106da08"
	p1g3CohortOracle       = "1e3b97d2-8fda-4510-9697-f36ca9ca2ab0"
	p1g3BoOracle           = "f5dc3dbd-7eab-40f3-afe2-1e88a0c00538"
)

// p1g3Permanent puts a permanent on the battlefield for `owner`.
func p1g3Permanent(g *game.Game, owner uuid.UUID, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = owner, owner
	if c.Name == "" {
		c.Name = "Test Permanent"
	}
	return pushBattlefieldCardWithTimestamp(g, c)
}

// p1g3ModeCard is one target announced for mode occurrence `occ`.
func p1g3ModeCard(id uuid.UUID, occ int) game.TargetRef {
	return game.TargetRef{Kind: game.TargetCard, ID: id, Mode: occ}
}

// p1g3CastModal casts a modal spell with its modes and targets
// announced together (CR 601.2b).
func p1g3CastModal(t *testing.T, g *game.Game, name, typeLine, oracle string, modes []int, targets ...game.TargetRef) uuid.UUID {
	t.Helper()
	return b02cCastModalSpell(t, g, name, typeLine, oracle, modes, targets)
}

// p1g3Activate has the active seat activate row `index` of `source`
// and resolves it.
func p1g3Activate(t *testing.T, g *game.Game, source uuid.UUID, index int, params game.ActivateAbilityParams) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	if err := g.ActivateCatalogAbility(me.ID, source, index, params); err != nil {
		t.Fatalf("activate row %d: %v", index, err)
	}
	passPriorityAroundTable(t, g)
}

// p1g3Untap untaps a permanent between activations.
func p1g3Untap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, id).Tapped = false })
}

// The four "deals N damage to target creature. If that creature would
// die this turn, exile it instead." bullets: the creature it kills is
// exiled, and so is a shielded one that was dealt nothing and is
// destroyed later — the replacement is the spell's.
func TestP1G3ModalDamageBulletsMarkTheirTarget(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		mode, toughness        int
	}{
		{"Agate Assault", "Sorcery", p1g3AgateOracle, 0, 4},
		{"Suplex", "Sorcery", p1g3SuplexOracle, 0, 3},
		{"Pinecone Strike", "Instant", p1g3PineconeOracle, 0, 3},
		{"Brutal Expulsion", "Instant", p1g3ExpulsionOracle, 1, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			opp := g.Seats[1].ID
			victim := p1Creature(g, opp, 1, tc.toughness)
			p1g3CastModal(t, g, tc.name, tc.typeLine, tc.oracle, []int{tc.mode}, p1g3ModeCard(victim, 0))
			passPriorityAroundTable(t, g)
			p1WantZone(t, g, victim, game.ZoneExile, "the creature the bullet killed")

			shielded := p1Creature(g, opp, 1, tc.toughness)
			p1ShieldCreature(g, shielded)
			p1g3CastModal(t, g, tc.name, tc.typeLine, tc.oracle, []int{tc.mode}, p1g3ModeCard(shielded, 0))
			passPriorityAroundTable(t, g)
			if got := damageMarkedOn(g, shielded); got != 0 {
				t.Fatalf("the shielded creature has %d damage", got)
			}
			p1Destroy(t, g, shielded)
			p1WantZone(t, g, shielded, game.ZoneExile, "the shielded creature destroyed later")
		})
	}
}

// Agate Assault and Suplex exile an artifact; Pinecone Strike destroys
// an artifact TOKEN and can't target a nontoken artifact.
func TestP1G3ArtifactBullets(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Agate Assault", p1g3AgateOracle},
		{"Suplex", p1g3SuplexOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			rock := p1g3Permanent(g, g.Seats[1].ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
			p1g3CastModal(t, g, tc.name, "Sorcery", tc.oracle, []int{1}, p1g3ModeCard(rock, 0))
			passPriorityAroundTable(t, g)
			p1WantZone(t, g, rock, game.ZoneExile, "the artifact")
		})
	}
	t.Run("Pinecone Strike", func(t *testing.T) {
		g := newCatalogGame(t)
		opp := g.Seats[1].ID
		rock := p1g3Permanent(g, opp, game.Card{Name: "Rock", TypeLine: "Artifact"})
		food := p1g3Permanent(g, opp, game.Card{Name: "Food", TypeLine: "Token Artifact — Food"})
		active := g.Seats[g.Turn.ActiveSeat]
		id := uuid.New()
		active.Hand.PushTop(game.Card{InstanceID: id, Name: "Pinecone Strike", TypeLine: "Instant",
			OracleID: p1g3PineconeOracle, Owner: active.ID, Controller: active.ID})
		if err := g.CastSpell(active.ID, id, game.CastSpellParams{Modes: []int{1},
			Targets: []game.TargetRef{p1g3ModeCard(rock, 0)}}); err == nil {
			t.Fatal("Pinecone Strike targeted a nontoken artifact")
		}
		p1g3CastModal(t, g, "Pinecone Strike", "Instant", p1g3PineconeOracle, []int{1}, p1g3ModeCard(food, 0))
		passPriorityAroundTable(t, g)
		if zoneOf(g, food) == game.ZoneBattlefield {
			t.Error("the artifact token survived Pinecone Strike")
		}
	})
}

// "Choose one or both": each bullet reads its own target.
func TestP1G3PineconeStrikeBothBullets(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1].ID
	bear := p1Creature(g, opp, p1BearPower, p1BearTo)
	food := p1g3Permanent(g, opp, game.Card{Name: "Food", TypeLine: "Token Artifact — Food"})
	p1g3CastModal(t, g, "Pinecone Strike", "Instant", p1g3PineconeOracle, []int{0, 1},
		p1g3ModeCard(bear, 0), p1g3ModeCard(food, 1))
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, bear, game.ZoneExile, "the bear")
	if zoneOf(g, food) == game.ZoneBattlefield {
		t.Error("the artifact token survived")
	}
}

// Brutal Expulsion: both bullets at once — a creature bounced and
// another exiled — and the first bullet returns a SPELL to its owner's
// hand without countering it.
func TestP1G3BrutalExpulsion(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1].ID
	big := p1Creature(g, opp, 5, 5)
	bear := p1Creature(g, opp, p1BearPower, p1BearTo)
	p1g3CastModal(t, g, "Brutal Expulsion", "Instant", p1g3ExpulsionOracle, []int{0, 1},
		p1g3ModeCard(big, 0), p1g3ModeCard(bear, 1))
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, big, game.ZoneHand, "the bounced 5/5")
	p1WantZone(t, g, bear, game.ZoneExile, "the bear dealt 2")

	target := p1Creature(g, opp, p1BearPower, p1BearTo)
	coil := castCatalogSpell(t, g, "Lava Coil", "Sorcery", p1LavaCoilOracle, pr6Card(target))
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Brutal Expulsion", TypeLine: "Instant",
		OracleID: p1g3ExpulsionOracle, Owner: active.ID, Controller: active.ID})
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Modes: []int{0},
		Targets: []game.TargetRef{p1g3ModeCard(coil, 0)}}); err != nil {
		t.Fatalf("Brutal Expulsion at Lava Coil: %v", err)
	}
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, coil, game.ZoneHand, "Lava Coil")
	p1WantZone(t, g, target, game.ZoneBattlefield, "Lava Coil's target")
}

// Brutal Expulsion's damage bullet: a planeswalker it kills is exiled.
func TestP1G3BrutalExpulsionExilesAPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	walker := p1g3Permanent(g, g.Seats[1].ID, game.Card{Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{"loyalty": 2}})
	p1g3CastModal(t, g, "Brutal Expulsion", "Instant", p1g3ExpulsionOracle, []int{1}, p1g3ModeCard(walker, 0))
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	p1WantZone(t, g, walker, game.ZoneExile, "the planeswalker")
}

// Gnashing of Teeth: the -5/-5 kills and the replacement exiles; the
// other bullet shrinks only the targeted player's creatures, and a
// creature it kills goes to the graveyard.
func TestP1G3GnashingOfTeeth(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	big := p1Creature(g, opp, 4, 4)
	p1g3CastModal(t, g, "Gnashing of Teeth", "Sorcery", p1g3GnashingOracle, []int{0}, p1g3ModeCard(big, 0))
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	p1WantZone(t, g, big, game.ZoneExile, "the 4/4 given -5/-5")

	small := p1Creature(g, opp, 1, 1)
	theirBear := p1Creature(g, opp, p1BearPower, p1BearTo)
	mine := p1Creature(g, me, 1, 1)
	p1g3CastModal(t, g, "Gnashing of Teeth", "Sorcery", p1g3GnashingOracle, []int{1},
		game.TargetRef{Kind: game.TargetPlayer, ID: opp})
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	p1WantZone(t, g, small, game.ZoneGraveyard, "the opponent's 1/1")
	p1WantZone(t, g, mine, game.ZoneBattlefield, "my 1/1")
	if c := apaLive(g, theirBear); c == nil || c.CurrentToughness() != 1 {
		t.Errorf("the opponent's bear is not 1/1 after -1/-1")
	}
}

// The "dealt damage this way" sweeps: a creature dealt damage and
// killed is exiled; one whose damage was all prevented is not marked.
func TestP1G3SweepsMarkOnlyWhatTheyDamaged(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		n            int
	}{
		{"Crush the Weak", p1g3CrushOracle, 2},
		{"Underworld Fires", p1g3UnderworldOracle, 1},
		{"Gallifrey Falls", p1g3GallifreyOracle, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			opp := g.Seats[1].ID
			dies := p1Creature(g, opp, 1, tc.n)
			shielded := p1Creature(g, opp, 1, tc.n)
			p1ShieldCreature(g, shielded)
			if tc.oracle == p1g3GallifreyOracle {
				p1g3CastSplit(t, g, game.CastSpellParams{})
			} else {
				castCatalogSpell(t, g, tc.name, "Sorcery", tc.oracle, nil)
			}
			passPriorityAroundTable(t, g)
			p1WantZone(t, g, dies, game.ZoneExile, "the creature the sweep killed")
			p1Destroy(t, g, shielded)
			p1WantZone(t, g, shielded, game.ZoneGraveyard, "the shielded creature")
		})
	}
}

// Crush the Weak declares foretell at its printed {R}.
func TestP1G3CrushTheWeakHasForetell(t *testing.T) {
	sa := foretellOf(p1g3CrushOracle)
	if sa == nil || sa.CastCost != "{R}" {
		t.Fatalf("Crush the Weak foretell = %+v, want cast cost {R}", sa)
	}
}

// Underworld Fires: "a PERMANENT dealt damage this way" — a planeswalker
// it kills is exiled.
func TestP1G3UnderworldFiresExilesAPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	walker := p1g3Permanent(g, g.Seats[1].ID, game.Card{Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{"loyalty": 1}})
	castCatalogSpell(t, g, "Underworld Fires", "Sorcery", p1g3UnderworldOracle, nil)
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	p1WantZone(t, g, walker, game.ZoneExile, "the planeswalker")
}

// p1g3GallifreyCard is Gallifrey Falls // No More in hand.
func p1g3GallifreyCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(), OracleID: p1g3GallifreyOracle, Layout: game.LayoutSplit,
		Owner: owner, Controller: owner,
		Faces: []game.Face{
			{Name: "Gallifrey Falls", TypeLine: "Instant", ManaCost: "{4}{R}{R}",
				OracleText: "Gallifrey Falls deals 4 damage to each creature. If a creature dealt damage this way would die this turn, exile it instead.\nFuse (You may cast one or both halves of this card from your hand.)"},
			{Name: "No More", TypeLine: "Instant", ManaCost: "{2}{W}",
				OracleText: "Any number of target creatures you control phase out.\nFuse (You may cast one or both halves of this card from your hand.)"},
		},
	}
	c.SettleImported()
	return c
}

// p1g3CastSplit casts Gallifrey Falls // No More from the active seat's
// hand with `params`.
func p1g3CastSplit(t *testing.T, g *game.Game, params game.CastSpellParams) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	c := p1g3GallifreyCard(me.ID)
	me.Hand.PushTop(c)
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, c.InstanceID, params); err != nil {
		t.Fatalf("cast Gallifrey Falls // No More: %v", err)
	}
}

// Fused: Gallifrey Falls deals its damage, then No More phases out the
// creatures I chose before any state-based check, so they survive; the
// rest are exiled.
func TestP1G3GallifreyFallsFused(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	saved := p1Creature(g, me, p1BearPower, p1BearTo)
	left := p1Creature(g, me, p1BearPower, p1BearTo)
	theirs := p1Creature(g, opp, p1BearPower, p1BearTo)
	p1g3CastSplit(t, g, game.CastSpellParams{Fuse: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: saved, Slot: 0}}})
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if !g.PhasedOut.Contains(saved) {
		t.Errorf("the creature No More named is not phased out (zone %q)", zoneOf(g, saved))
	}
	p1WantZone(t, g, left, game.ZoneExile, "my creature No More did not name")
	p1WantZone(t, g, theirs, game.ZoneExile, "the opponent's creature")
}

// No More on its own phases out only creatures I control.
func TestP1G3NoMoreAlone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	a := p1Creature(g, me, p1BearPower, p1BearTo)
	b := p1Creature(g, me, p1BearPower, p1BearTo)
	theirs := p1Creature(g, opp, p1BearPower, p1BearTo)
	mePlayer := g.Seats[g.Turn.ActiveSeat]
	bad := p1g3GallifreyCard(mePlayer.ID)
	mePlayer.Hand.PushTop(bad)
	advanceToMain(t, g)
	if err := g.CastSpell(mePlayer.ID, bad.InstanceID, game.CastSpellParams{Face: 1,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}}); err == nil {
		t.Fatal("No More targeted a creature I don't control")
	}
	p1g3CastSplit(t, g, game.CastSpellParams{Face: 1, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}})
	passPriorityAroundTable(t, g)
	if !g.PhasedOut.Contains(a) || !g.PhasedOut.Contains(b) {
		t.Error("No More did not phase out both of my creatures")
	}
}

// Flamebreak: every creature without flying and every player is dealt
// 3; a creature dealt damage can't be regenerated, one whose damage was
// prevented can, and a flyer is untouched.
func TestP1G3Flamebreak(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := p1Creature(g, opp.ID, p1BearPower, p1BearTo)
	p1Regenerate(t, g, bear)
	shielded := p1Creature(g, opp.ID, p1BearPower, p1BearTo)
	p1ShieldCreature(g, shielded)
	p1Regenerate(t, g, shielded)
	flyer := p1g3Permanent(g, opp.ID, game.Card{TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	life := opp.Life
	castCatalogSpell(t, g, "Flamebreak", "Sorcery", p1g3FlamebreakOracle, nil)
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	p1WantZone(t, g, bear, game.ZoneGraveyard, "the regenerating bear Flamebreak dealt 3")
	p1WantZone(t, g, flyer, game.ZoneBattlefield, "the flyer")
	if got := damageMarkedOn(g, flyer); got != 0 {
		t.Errorf("the flyer has %d damage", got)
	}
	if opp.Life != life-3 {
		t.Errorf("opponent's life %d → %d, want -3", life, opp.Life)
	}
	p1Destroy(t, g, shielded)
	p1WantZone(t, g, shielded, game.ZoneBattlefield, "a regenerating creature Flamebreak dealt nothing")
}

// Serpentine Spike: 2, 3 and 4 damage to three different creatures,
// each exiled; a target dealt nothing is not marked; one creature can't
// fill two of the clauses.
func TestP1G3SerpentineSpike(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1].ID
	a := p1Creature(g, opp, 1, 2)
	b := p1Creature(g, opp, 1, 3)
	c := p1Creature(g, opp, 1, 4)
	slots := func(x, y, z uuid.UUID) []game.TargetRef {
		return []game.TargetRef{
			{Kind: game.TargetCard, ID: x, Slot: 0},
			{Kind: game.TargetCard, ID: y, Slot: 1},
			{Kind: game.TargetCard, ID: z, Slot: 2},
		}
	}
	if err := castCatalogSpellErr(t, g, "Serpentine Spike", "Sorcery", p1g3SpikeOracle, slots(a, a, b)); err == nil {
		t.Fatal("one creature filled two of Serpentine Spike's clauses")
	}
	castCatalogSpell(t, g, "Serpentine Spike", "Sorcery", p1g3SpikeOracle, slots(a, b, c))
	passPriorityAroundTable(t, g)
	for id, what := range map[uuid.UUID]string{a: "the 2-damage target", b: "the 3-damage target", c: "the 4-damage target"} {
		p1WantZone(t, g, id, game.ZoneExile, what)
	}

	x := p1Creature(g, opp, 1, 5)
	y := p1Creature(g, opp, 1, 5)
	z := p1Creature(g, opp, 1, 5)
	p1ShieldCreature(g, y)
	castCatalogSpell(t, g, "Serpentine Spike", "Sorcery", p1g3SpikeOracle, slots(x, y, z))
	passPriorityAroundTable(t, g)
	if dx, dy, dz := damageMarkedOn(g, x), damageMarkedOn(g, y), damageMarkedOn(g, z); dx != 2 || dy != 0 || dz != 4 {
		t.Fatalf("damage %d/%d/%d, want 2/0/4", dx, dy, dz)
	}
	p1Destroy(t, g, x)
	p1Destroy(t, g, y)
	p1WantZone(t, g, x, game.ZoneExile, "a survivor Serpentine Spike damaged")
	p1WantZone(t, g, y, game.ZoneGraveyard, "the shielded survivor")
}

// Jaya Ballard: each ability discards a card; the first destroys a blue
// permanent, the second's target can't be regenerated, the third hits
// each creature and each player.
func TestP1G3JayaBallard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	jaya := pushCatalogPermanent(g, me.ID, "Jaya Ballard, Task Mage", "Legendary Creature — Human Spellshaper", p1g3JayaOracle, false)
	discard := func() []uuid.UUID {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Spare", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
		return []uuid.UUID{id}
	}

	red := p1g3Permanent(g, opp.ID, game.Card{TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"R"}})
	if err := g.ActivateCatalogAbility(me.ID, jaya, 0, game.ActivateAbilityParams{Targets: pr6Card(red), DiscardIDs: discard()}); err == nil {
		t.Fatal("Jaya's first ability targeted a red creature")
	}
	blue := p1g3Permanent(g, opp.ID, game.Card{TypeLine: "Enchantment", Colors: []string{"U"}})
	spare := discard()
	p1g3Activate(t, g, jaya, 0, game.ActivateAbilityParams{Targets: pr6Card(blue), DiscardIDs: spare})
	p1WantZone(t, g, blue, game.ZoneGraveyard, "the blue permanent")
	p1WantZone(t, g, spare[0], game.ZoneGraveyard, "the discarded card")

	p1g3Untap(g, jaya)
	bear := p1Creature(g, opp.ID, p1BearPower, p1BearTo)
	p1Regenerate(t, g, bear)
	p1g3Activate(t, g, jaya, 1, game.ActivateAbilityParams{Targets: pr6Card(bear), DiscardIDs: discard()})
	g.RunStateChecksForTest()
	p1WantZone(t, g, bear, game.ZoneGraveyard, "the regenerating bear Jaya dealt 3")

	p1g3Untap(g, jaya)
	big := p1Creature(g, opp.ID, 5, 7)
	life := opp.Life
	p1g3Activate(t, g, jaya, 2, game.ActivateAbilityParams{DiscardIDs: discard()})
	if got := damageMarkedOn(g, big); got != 6 {
		t.Errorf("the 5/7 has %d damage, want 6", got)
	}
	if opp.Life != life-6 {
		t.Errorf("opponent's life %d → %d, want -6", life, opp.Life)
	}
}

// Orcish Healer: the first ability stops a regeneration shield; the
// others regenerate a black or green creature and nothing else.
func TestP1G3OrcishHealer(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	healer := pushCatalogPermanent(g, me.ID, "Orcish Healer", "Creature — Orc Cleric", p1g3OrcishHealerOracle, false)

	bear := p1Creature(g, opp.ID, p1BearPower, p1BearTo)
	p1Regenerate(t, g, bear)
	p1g3Activate(t, g, healer, 0, game.ActivateAbilityParams{Targets: pr6Card(bear)})
	p1Destroy(t, g, bear)
	p1WantZone(t, g, bear, game.ZoneGraveyard, "the bear Orcish Healer marked")

	for _, row := range []int{1, 2} {
		p1g3Untap(g, healer)
		red := p1g3Permanent(g, me.ID, game.Card{TypeLine: "Creature — Goblin", Power: 1, Toughness: 1, Colors: []string{"R"}})
		if err := g.ActivateCatalogAbility(me.ID, healer, row, game.ActivateAbilityParams{Targets: pr6Card(red)}); err == nil {
			t.Fatalf("row %d regenerated a red creature", row)
		}
		black := p1g3Permanent(g, me.ID, game.Card{TypeLine: "Creature — Zombie", Power: 1, Toughness: 1, Colors: []string{"B"}})
		p1g3Activate(t, g, healer, row, game.ActivateAbilityParams{Targets: pr6Card(black)})
		p1Destroy(t, g, black)
		p1WantZone(t, g, black, game.ZoneBattlefield, "the regenerated black creature")
	}
}

// Lim-Dûl's Cohort: blocking a creature, or being blocked by one, marks
// THAT creature, once per creature.
func TestP1G3LimDulsCohort(t *testing.T) {
	g := newCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	cohort := pushCatalogPermanent(g, me.ID, "Lim-Dûl's Cohort", "Creature — Zombie", p1g3CohortOracle, false)
	block := func(blocker, attacker uuid.UUID, actor uuid.UUID) {
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventBlock, Actor: actor, CardID: blocker, Target: attacker, Amount: 1})
		})
	}
	attacker := p1Creature(g, bob.ID, p1BearPower, p1BearTo)
	p1Regenerate(t, g, attacker)
	block(cohort, attacker, me.ID)
	passPriorityAroundTable(t, g)
	p1Destroy(t, g, attacker)
	p1WantZone(t, g, attacker, game.ZoneGraveyard, "the creature the Cohort blocked")

	b1 := p1Creature(g, bob.ID, p1BearPower, p1BearTo)
	b2 := p1Creature(g, bob.ID, p1BearPower, p1BearTo)
	p1Regenerate(t, g, b1)
	p1Regenerate(t, g, b2)
	block(b1, cohort, bob.ID)
	block(b2, cohort, bob.ID)
	passPriorityAroundTable(t, g)
	p1Destroy(t, g, b1)
	p1Destroy(t, g, b2)
	p1WantZone(t, g, b1, game.ZoneGraveyard, "the first blocker")
	p1WantZone(t, g, b2, game.ZoneGraveyard, "the second blocker")
}

// Nine-Ringed Bo: only a Spirit; it is marked whether or not the 1
// damage kills it.
func TestP1G3NineRingedBo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	bo := pushCatalogPermanent(g, me.ID, "Nine-Ringed Bo", "Artifact", p1g3BoOracle, false)
	bear := p1Creature(g, opp, 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, bo, 0, game.ActivateAbilityParams{Targets: pr6Card(bear)}); err == nil {
		t.Fatal("Nine-Ringed Bo targeted a non-Spirit")
	}
	spirit := p1g3Permanent(g, opp, game.Card{TypeLine: "Creature — Spirit", Power: 1, Toughness: 1})
	p1g3Activate(t, g, bo, 0, game.ActivateAbilityParams{Targets: pr6Card(spirit)})
	g.RunStateChecksForTest()
	p1WantZone(t, g, spirit, game.ZoneExile, "the Spirit the Bo killed")

	p1g3Untap(g, bo)
	big := p1g3Permanent(g, opp, game.Card{TypeLine: "Creature — Spirit", Power: 3, Toughness: 3})
	p1g3Activate(t, g, bo, 0, game.ActivateAbilityParams{Targets: pr6Card(big)})
	p1WantZone(t, g, big, game.ZoneBattlefield, "the 3/3 Spirit after 1 damage")
	p1Destroy(t, g, big)
	p1WantZone(t, g, big, game.ZoneExile, "the 3/3 Spirit destroyed later")
}
