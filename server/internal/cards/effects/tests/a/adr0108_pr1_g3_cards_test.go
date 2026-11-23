package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// p1g3Permanent puts a permanent on the battlefield for `owner`.
func p1g3Permanent(g *game.Game, owner uuid.UUID, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = owner, owner
	if c.Name == "" {
		c.Name = "Test Permanent"
	}
	return effectstest.PushBattlefieldCardWithTimestamp(g, c)
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
		{"Agate Assault", "Sorcery", effectstest.P1g3AgateOracle, 0, 4},
		{"Suplex", "Sorcery", effectstest.P1g3SuplexOracle, 0, 3},
		{"Pinecone Strike", "Instant", effectstest.P1g3PineconeOracle, 0, 3},
		{"Brutal Expulsion", "Instant", effectstest.P1g3ExpulsionOracle, 1, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			opp := g.Seats[1].ID
			victim := effectstest.P1Creature(g, opp, 1, tc.toughness)
			effectstest.P1g3CastModal(t, g, tc.name, tc.typeLine, tc.oracle, []int{tc.mode}, effectstest.P1g3ModeCard(victim, 0))
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.P1WantZone(t, g, victim, game.ZoneExile, "the creature the bullet killed")

			shielded := effectstest.P1Creature(g, opp, 1, tc.toughness)
			effectstest.P1ShieldCreature(g, shielded)
			effectstest.P1g3CastModal(t, g, tc.name, tc.typeLine, tc.oracle, []int{tc.mode}, effectstest.P1g3ModeCard(shielded, 0))
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.DamageMarkedOn(g, shielded); got != 0 {
				t.Fatalf("the shielded creature has %d damage", got)
			}
			effectstest.P1Destroy(t, g, shielded)
			effectstest.P1WantZone(t, g, shielded, game.ZoneExile, "the shielded creature destroyed later")
		})
	}
}

// Agate Assault and Suplex exile an artifact; Pinecone Strike destroys
// an artifact TOKEN and can't target a nontoken artifact.
func TestP1G3ArtifactBullets(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Agate Assault", effectstest.P1g3AgateOracle},
		{"Suplex", effectstest.P1g3SuplexOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			rock := p1g3Permanent(g, g.Seats[1].ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
			effectstest.P1g3CastModal(t, g, tc.name, "Sorcery", tc.oracle, []int{1}, effectstest.P1g3ModeCard(rock, 0))
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.P1WantZone(t, g, rock, game.ZoneExile, "the artifact")
		})
	}
	t.Run("Pinecone Strike", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[1].ID
		rock := p1g3Permanent(g, opp, game.Card{Name: "Rock", TypeLine: "Artifact"})
		food := p1g3Permanent(g, opp, game.Card{Name: "Food", TypeLine: "Token Artifact — Food"})
		active := g.Seats[g.Turn.ActiveSeat]
		id := uuid.New()
		active.Hand.PushTop(game.Card{InstanceID: id, Name: "Pinecone Strike", TypeLine: "Instant",
			OracleID: effectstest.P1g3PineconeOracle, Owner: active.ID, Controller: active.ID})
		if err := g.CastSpell(active.ID, id, game.CastSpellParams{Modes: []int{1},
			Targets: []game.TargetRef{effectstest.P1g3ModeCard(rock, 0)}}); err == nil {
			t.Fatal("Pinecone Strike targeted a nontoken artifact")
		}
		effectstest.P1g3CastModal(t, g, "Pinecone Strike", "Instant", effectstest.P1g3PineconeOracle, []int{1}, effectstest.P1g3ModeCard(food, 0))
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.ZoneOf(g, food) == game.ZoneBattlefield {
			t.Error("the artifact token survived Pinecone Strike")
		}
	})
}

// "Choose one or both": each bullet reads its own target.
func TestP1G3PineconeStrikeBothBullets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1].ID
	bear := effectstest.P1Creature(g, opp, effectstest.P1BearPower, effectstest.P1BearTo)
	food := p1g3Permanent(g, opp, game.Card{Name: "Food", TypeLine: "Token Artifact — Food"})
	effectstest.P1g3CastModal(t, g, "Pinecone Strike", "Instant", effectstest.P1g3PineconeOracle, []int{0, 1},
		effectstest.P1g3ModeCard(bear, 0), effectstest.P1g3ModeCard(food, 1))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1WantZone(t, g, bear, game.ZoneExile, "the bear")
	if effectstest.ZoneOf(g, food) == game.ZoneBattlefield {
		t.Error("the artifact token survived")
	}
}

// Brutal Expulsion: both bullets at once — a creature bounced and
// another exiled — and the first bullet returns a SPELL to its owner's
// hand without countering it.
func TestP1G3BrutalExpulsion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1].ID
	big := effectstest.P1Creature(g, opp, 5, 5)
	bear := effectstest.P1Creature(g, opp, effectstest.P1BearPower, effectstest.P1BearTo)
	effectstest.P1g3CastModal(t, g, "Brutal Expulsion", "Instant", effectstest.P1g3ExpulsionOracle, []int{0, 1},
		effectstest.P1g3ModeCard(big, 0), effectstest.P1g3ModeCard(bear, 1))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1WantZone(t, g, big, game.ZoneHand, "the bounced 5/5")
	effectstest.P1WantZone(t, g, bear, game.ZoneExile, "the bear dealt 2")

	target := effectstest.P1Creature(g, opp, effectstest.P1BearPower, effectstest.P1BearTo)
	coil := effectstest.CastCatalogSpell(t, g, "Lava Coil", "Sorcery", effectstest.P1LavaCoilOracle, effectstest.Pr6Card(target))
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Brutal Expulsion", TypeLine: "Instant",
		OracleID: effectstest.P1g3ExpulsionOracle, Owner: active.ID, Controller: active.ID})
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Modes: []int{0},
		Targets: []game.TargetRef{effectstest.P1g3ModeCard(coil, 0)}}); err != nil {
		t.Fatalf("Brutal Expulsion at Lava Coil: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1WantZone(t, g, coil, game.ZoneHand, "Lava Coil")
	effectstest.P1WantZone(t, g, target, game.ZoneBattlefield, "Lava Coil's target")
}

// Brutal Expulsion's damage bullet: a planeswalker it kills is exiled.
func TestP1G3BrutalExpulsionExilesAPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	walker := p1g3Permanent(g, g.Seats[1].ID, game.Card{Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{"loyalty": 2}})
	effectstest.P1g3CastModal(t, g, "Brutal Expulsion", "Instant", effectstest.P1g3ExpulsionOracle, []int{1}, effectstest.P1g3ModeCard(walker, 0))
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	effectstest.P1WantZone(t, g, walker, game.ZoneExile, "the planeswalker")
}

// Gnashing of Teeth: the -5/-5 kills and the replacement exiles; the
// other bullet shrinks only the targeted player's creatures, and a
// creature it kills goes to the graveyard.
func TestP1G3GnashingOfTeeth(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	big := effectstest.P1Creature(g, opp, 4, 4)
	effectstest.P1g3CastModal(t, g, "Gnashing of Teeth", "Sorcery", effectstest.P1g3GnashingOracle, []int{0}, effectstest.P1g3ModeCard(big, 0))
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	effectstest.P1WantZone(t, g, big, game.ZoneExile, "the 4/4 given -5/-5")

	small := effectstest.P1Creature(g, opp, 1, 1)
	theirBear := effectstest.P1Creature(g, opp, effectstest.P1BearPower, effectstest.P1BearTo)
	mine := effectstest.P1Creature(g, me, 1, 1)
	effectstest.P1g3CastModal(t, g, "Gnashing of Teeth", "Sorcery", effectstest.P1g3GnashingOracle, []int{1},
		game.TargetRef{Kind: game.TargetPlayer, ID: opp})
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	effectstest.P1WantZone(t, g, small, game.ZoneGraveyard, "the opponent's 1/1")
	effectstest.P1WantZone(t, g, mine, game.ZoneBattlefield, "my 1/1")
	if c := effectstest.ApaLive(g, theirBear); c == nil || c.CurrentToughness() != 1 {
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
		{"Crush the Weak", effectstest.P1g3CrushOracle, 2},
		{"Underworld Fires", effectstest.P1g3UnderworldOracle, 1},
		{"Gallifrey Falls", effectstest.P1g3GallifreyOracle, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			opp := g.Seats[1].ID
			dies := effectstest.P1Creature(g, opp, 1, tc.n)
			shielded := effectstest.P1Creature(g, opp, 1, tc.n)
			effectstest.P1ShieldCreature(g, shielded)
			if tc.oracle == effectstest.P1g3GallifreyOracle {
				p1g3CastSplit(t, g, game.CastSpellParams{})
			} else {
				effectstest.CastCatalogSpell(t, g, tc.name, "Sorcery", tc.oracle, nil)
			}
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.P1WantZone(t, g, dies, game.ZoneExile, "the creature the sweep killed")
			effectstest.P1Destroy(t, g, shielded)
			effectstest.P1WantZone(t, g, shielded, game.ZoneGraveyard, "the shielded creature")
		})
	}
}

// Crush the Weak declares foretell at its printed {R}.
func TestP1G3CrushTheWeakHasForetell(t *testing.T) {
	sa := effectstest.ForetellOf(effectstest.P1g3CrushOracle)
	if sa == nil || sa.CastCost != "{R}" {
		t.Fatalf("Crush the Weak foretell = %+v, want cast cost {R}", sa)
	}
}

// Underworld Fires: "a PERMANENT dealt damage this way" — a planeswalker
// it kills is exiled.
func TestP1G3UnderworldFiresExilesAPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	walker := p1g3Permanent(g, g.Seats[1].ID, game.Card{Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{"loyalty": 1}})
	effectstest.CastCatalogSpell(t, g, "Underworld Fires", "Sorcery", effectstest.P1g3UnderworldOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	effectstest.P1WantZone(t, g, walker, game.ZoneExile, "the planeswalker")
}

// p1g3GallifreyCard is Gallifrey Falls // No More in hand.
func p1g3GallifreyCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(), OracleID: effectstest.P1g3GallifreyOracle, Layout: game.LayoutSplit,
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
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, c.InstanceID, params); err != nil {
		t.Fatalf("cast Gallifrey Falls // No More: %v", err)
	}
}

// Fused: Gallifrey Falls deals its damage, then No More phases out the
// creatures I chose before any state-based check, so they survive; the
// rest are exiled.
func TestP1G3GallifreyFallsFused(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	saved := effectstest.P1Creature(g, me, effectstest.P1BearPower, effectstest.P1BearTo)
	left := effectstest.P1Creature(g, me, effectstest.P1BearPower, effectstest.P1BearTo)
	theirs := effectstest.P1Creature(g, opp, effectstest.P1BearPower, effectstest.P1BearTo)
	p1g3CastSplit(t, g, game.CastSpellParams{Fuse: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: saved, Slot: 0}}})
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if !g.PhasedOut.Contains(saved) {
		t.Errorf("the creature No More named is not phased out (zone %q)", effectstest.ZoneOf(g, saved))
	}
	effectstest.P1WantZone(t, g, left, game.ZoneExile, "my creature No More did not name")
	effectstest.P1WantZone(t, g, theirs, game.ZoneExile, "the opponent's creature")
}

// No More on its own phases out only creatures I control.
func TestP1G3NoMoreAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	a := effectstest.P1Creature(g, me, effectstest.P1BearPower, effectstest.P1BearTo)
	b := effectstest.P1Creature(g, me, effectstest.P1BearPower, effectstest.P1BearTo)
	theirs := effectstest.P1Creature(g, opp, effectstest.P1BearPower, effectstest.P1BearTo)
	mePlayer := g.Seats[g.Turn.ActiveSeat]
	bad := p1g3GallifreyCard(mePlayer.ID)
	mePlayer.Hand.PushTop(bad)
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(mePlayer.ID, bad.InstanceID, game.CastSpellParams{Face: 1,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}}); err == nil {
		t.Fatal("No More targeted a creature I don't control")
	}
	p1g3CastSplit(t, g, game.CastSpellParams{Face: 1, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.PhasedOut.Contains(a) || !g.PhasedOut.Contains(b) {
		t.Error("No More did not phase out both of my creatures")
	}
}

// Flamebreak: every creature without flying and every player is dealt
// 3; a creature dealt damage can't be regenerated, one whose damage was
// prevented can, and a flyer is untouched.
func TestP1G3Flamebreak(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.P1Creature(g, opp.ID, effectstest.P1BearPower, effectstest.P1BearTo)
	effectstest.P1Regenerate(t, g, bear)
	shielded := effectstest.P1Creature(g, opp.ID, effectstest.P1BearPower, effectstest.P1BearTo)
	effectstest.P1ShieldCreature(g, shielded)
	effectstest.P1Regenerate(t, g, shielded)
	flyer := p1g3Permanent(g, opp.ID, game.Card{TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Flamebreak", "Sorcery", effectstest.P1g3FlamebreakOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	effectstest.P1WantZone(t, g, bear, game.ZoneGraveyard, "the regenerating bear Flamebreak dealt 3")
	effectstest.P1WantZone(t, g, flyer, game.ZoneBattlefield, "the flyer")
	if got := effectstest.DamageMarkedOn(g, flyer); got != 0 {
		t.Errorf("the flyer has %d damage", got)
	}
	if opp.Life != life-3 {
		t.Errorf("opponent's life %d → %d, want -3", life, opp.Life)
	}
	effectstest.P1Destroy(t, g, shielded)
	effectstest.P1WantZone(t, g, shielded, game.ZoneBattlefield, "a regenerating creature Flamebreak dealt nothing")
}

// Serpentine Spike: 2, 3 and 4 damage to three different creatures,
// each exiled; a target dealt nothing is not marked; one creature can't
// fill two of the clauses.
func TestP1G3SerpentineSpike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1].ID
	a := effectstest.P1Creature(g, opp, 1, 2)
	b := effectstest.P1Creature(g, opp, 1, 3)
	c := effectstest.P1Creature(g, opp, 1, 4)
	slots := func(x, y, z uuid.UUID) []game.TargetRef {
		return []game.TargetRef{
			{Kind: game.TargetCard, ID: x, Slot: 0},
			{Kind: game.TargetCard, ID: y, Slot: 1},
			{Kind: game.TargetCard, ID: z, Slot: 2},
		}
	}
	if err := effectstest.CastCatalogSpellErr(t, g, "Serpentine Spike", "Sorcery", effectstest.P1g3SpikeOracle, slots(a, a, b)); err == nil {
		t.Fatal("one creature filled two of Serpentine Spike's clauses")
	}
	effectstest.CastCatalogSpell(t, g, "Serpentine Spike", "Sorcery", effectstest.P1g3SpikeOracle, slots(a, b, c))
	effectstest.PassPriorityAroundTable(t, g)
	for id, what := range map[uuid.UUID]string{a: "the 2-damage target", b: "the 3-damage target", c: "the 4-damage target"} {
		effectstest.P1WantZone(t, g, id, game.ZoneExile, what)
	}

	x := effectstest.P1Creature(g, opp, 1, 5)
	y := effectstest.P1Creature(g, opp, 1, 5)
	z := effectstest.P1Creature(g, opp, 1, 5)
	effectstest.P1ShieldCreature(g, y)
	effectstest.CastCatalogSpell(t, g, "Serpentine Spike", "Sorcery", effectstest.P1g3SpikeOracle, slots(x, y, z))
	effectstest.PassPriorityAroundTable(t, g)
	if dx, dy, dz := effectstest.DamageMarkedOn(g, x), effectstest.DamageMarkedOn(g, y), effectstest.DamageMarkedOn(g, z); dx != 2 || dy != 0 || dz != 4 {
		t.Fatalf("damage %d/%d/%d, want 2/0/4", dx, dy, dz)
	}
	effectstest.P1Destroy(t, g, x)
	effectstest.P1Destroy(t, g, y)
	effectstest.P1WantZone(t, g, x, game.ZoneExile, "a survivor Serpentine Spike damaged")
	effectstest.P1WantZone(t, g, y, game.ZoneGraveyard, "the shielded survivor")
}

// Jaya Ballard: each ability discards a card; the first destroys a blue
// permanent, the second's target can't be regenerated, the third hits
// each creature and each player.
func TestP1G3JayaBallard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	jaya := effectstest.PushCatalogPermanent(g, me.ID, "Jaya Ballard, Task Mage", "Legendary Creature — Human Spellshaper", effectstest.P1g3JayaOracle, false)
	discard := func() []uuid.UUID {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Spare", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
		return []uuid.UUID{id}
	}

	red := p1g3Permanent(g, opp.ID, game.Card{TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"R"}})
	if err := g.ActivateCatalogAbility(me.ID, jaya, 0, game.ActivateAbilityParams{Targets: effectstest.Pr6Card(red), DiscardIDs: discard()}); err == nil {
		t.Fatal("Jaya's first ability targeted a red creature")
	}
	blue := p1g3Permanent(g, opp.ID, game.Card{TypeLine: "Enchantment", Colors: []string{"U"}})
	spare := discard()
	effectstest.P1g3Activate(t, g, jaya, 0, game.ActivateAbilityParams{Targets: effectstest.Pr6Card(blue), DiscardIDs: spare})
	effectstest.P1WantZone(t, g, blue, game.ZoneGraveyard, "the blue permanent")
	effectstest.P1WantZone(t, g, spare[0], game.ZoneGraveyard, "the discarded card")

	effectstest.P1g3Untap(g, jaya)
	bear := effectstest.P1Creature(g, opp.ID, effectstest.P1BearPower, effectstest.P1BearTo)
	effectstest.P1Regenerate(t, g, bear)
	effectstest.P1g3Activate(t, g, jaya, 1, game.ActivateAbilityParams{Targets: effectstest.Pr6Card(bear), DiscardIDs: discard()})
	g.RunStateChecksForTest()
	effectstest.P1WantZone(t, g, bear, game.ZoneGraveyard, "the regenerating bear Jaya dealt 3")

	effectstest.P1g3Untap(g, jaya)
	big := effectstest.P1Creature(g, opp.ID, 5, 7)
	life := opp.Life
	effectstest.P1g3Activate(t, g, jaya, 2, game.ActivateAbilityParams{DiscardIDs: discard()})
	if got := effectstest.DamageMarkedOn(g, big); got != 6 {
		t.Errorf("the 5/7 has %d damage, want 6", got)
	}
	if opp.Life != life-6 {
		t.Errorf("opponent's life %d → %d, want -6", life, opp.Life)
	}
}

// Orcish Healer: the first ability stops a regeneration shield; the
// others regenerate a black or green creature and nothing else.
func TestP1G3OrcishHealer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	healer := effectstest.PushCatalogPermanent(g, me.ID, "Orcish Healer", "Creature — Orc Cleric", effectstest.P1g3OrcishHealerOracle, false)

	bear := effectstest.P1Creature(g, opp.ID, effectstest.P1BearPower, effectstest.P1BearTo)
	effectstest.P1Regenerate(t, g, bear)
	effectstest.P1g3Activate(t, g, healer, 0, game.ActivateAbilityParams{Targets: effectstest.Pr6Card(bear)})
	effectstest.P1Destroy(t, g, bear)
	effectstest.P1WantZone(t, g, bear, game.ZoneGraveyard, "the bear Orcish Healer marked")

	for _, row := range []int{1, 2} {
		effectstest.P1g3Untap(g, healer)
		red := p1g3Permanent(g, me.ID, game.Card{TypeLine: "Creature — Goblin", Power: 1, Toughness: 1, Colors: []string{"R"}})
		if err := g.ActivateCatalogAbility(me.ID, healer, row, game.ActivateAbilityParams{Targets: effectstest.Pr6Card(red)}); err == nil {
			t.Fatalf("row %d regenerated a red creature", row)
		}
		black := p1g3Permanent(g, me.ID, game.Card{TypeLine: "Creature — Zombie", Power: 1, Toughness: 1, Colors: []string{"B"}})
		effectstest.P1g3Activate(t, g, healer, row, game.ActivateAbilityParams{Targets: effectstest.Pr6Card(black)})
		effectstest.P1Destroy(t, g, black)
		effectstest.P1WantZone(t, g, black, game.ZoneBattlefield, "the regenerated black creature")
	}
}

// Lim-Dûl's Cohort: blocking a creature, or being blocked by one, marks
// THAT creature, once per creature.
func TestP1G3LimDulsCohort(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	cohort := effectstest.PushCatalogPermanent(g, me.ID, "Lim-Dûl's Cohort", "Creature — Zombie", effectstest.P1g3CohortOracle, false)
	block := func(blocker, attacker uuid.UUID, actor uuid.UUID) {
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventBlock, Actor: actor, CardID: blocker, Target: attacker, Amount: 1})
		})
	}
	attacker := effectstest.P1Creature(g, bob.ID, effectstest.P1BearPower, effectstest.P1BearTo)
	effectstest.P1Regenerate(t, g, attacker)
	block(cohort, attacker, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1Destroy(t, g, attacker)
	effectstest.P1WantZone(t, g, attacker, game.ZoneGraveyard, "the creature the Cohort blocked")

	b1 := effectstest.P1Creature(g, bob.ID, effectstest.P1BearPower, effectstest.P1BearTo)
	b2 := effectstest.P1Creature(g, bob.ID, effectstest.P1BearPower, effectstest.P1BearTo)
	effectstest.P1Regenerate(t, g, b1)
	effectstest.P1Regenerate(t, g, b2)
	block(b1, cohort, bob.ID)
	block(b2, cohort, bob.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1Destroy(t, g, b1)
	effectstest.P1Destroy(t, g, b2)
	effectstest.P1WantZone(t, g, b1, game.ZoneGraveyard, "the first blocker")
	effectstest.P1WantZone(t, g, b2, game.ZoneGraveyard, "the second blocker")
}

// Nine-Ringed Bo: only a Spirit; it is marked whether or not the 1
// damage kills it.
func TestP1G3NineRingedBo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	bo := effectstest.PushCatalogPermanent(g, me.ID, "Nine-Ringed Bo", "Artifact", effectstest.P1g3BoOracle, false)
	bear := effectstest.P1Creature(g, opp, 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, bo, 0, game.ActivateAbilityParams{Targets: effectstest.Pr6Card(bear)}); err == nil {
		t.Fatal("Nine-Ringed Bo targeted a non-Spirit")
	}
	spirit := p1g3Permanent(g, opp, game.Card{TypeLine: "Creature — Spirit", Power: 1, Toughness: 1})
	effectstest.P1g3Activate(t, g, bo, 0, game.ActivateAbilityParams{Targets: effectstest.Pr6Card(spirit)})
	g.RunStateChecksForTest()
	effectstest.P1WantZone(t, g, spirit, game.ZoneExile, "the Spirit the Bo killed")

	effectstest.P1g3Untap(g, bo)
	big := p1g3Permanent(g, opp, game.Card{TypeLine: "Creature — Spirit", Power: 3, Toughness: 3})
	effectstest.P1g3Activate(t, g, bo, 0, game.ActivateAbilityParams{Targets: effectstest.Pr6Card(big)})
	effectstest.P1WantZone(t, g, big, game.ZoneBattlefield, "the 3/3 Spirit after 1 damage")
	effectstest.P1Destroy(t, g, big)
	effectstest.P1WantZone(t, g, big, game.ZoneExile, "the 3/3 Spirit destroyed later")
}
