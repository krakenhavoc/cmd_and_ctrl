package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exile_cast_strip_view_test.go — #1389, the castable-from-exile strip.
//
// Two answers ride the holder's own frame for a card in exile:
// `castable_here` ("you may cast it NOW", timing included) and
// `cast_prices` (what the engine will charge, after every CR 601.2f
// modifier). One test per mechanic in the issue's table, then the
// leak tests: an opponent and a spectator get neither, and a
// face-down foretold card stays a card back to anybody who does not
// know it.

// stripTable is a four-seat table in the ACTIVE seat's precombat main
// phase with the stack empty — the window every sorcery-speed exile
// cast opens in.
func stripTable(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := busyTable(t, 1)
	advanceTo(t, g, game.StepPrecombatMain)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	return g, me, opp
}

// exiledSpell is a single-faced spell every seat knows, ready for
// exileWithGrant.
func exiledSpell(owner uuid.UUID, name, typeLine, cost string) game.Card {
	c := game.NewCard(name, owner)
	c.TypeLine = typeLine
	c.ManaCost = cost
	c.OracleID = "test-strip-" + name
	return c
}

// stripCard is the viewer's copy of one exiled card.
func stripCard(t *testing.T, g *game.Game, viewer string, id uuid.UUID) *CardView {
	t.Helper()
	c := cardInZone(ViewOfGameFor(g, viewer).Exile, id)
	if c == nil {
		t.Fatalf("card %s is not in viewer %q's exile", id, viewer)
	}
	return c
}

// assertPrice checks the badge's input: the cheapest price, whether it
// is the printed cost, and the offer it claims.
func assertPrice(t *testing.T, label string, c *CardView, cost string, printed bool, key string) {
	t.Helper()
	if len(c.CastPrices) == 0 {
		t.Fatalf("%s: cast_prices is empty, want %s", label, cost)
	}
	p := c.CastPrices[0]
	if p.Cost != cost || p.Printed != printed || p.AlternativeCost != key {
		t.Errorf("%s: cast_prices[0] = %+v, want cost %s printed=%v alternative_cost=%q",
			label, p, cost, printed, key)
	}
}

// assertNothingFor checks that a seat that holds no permission — an
// opponent, a spectator — is told nothing about casting the card.
func assertNothingFor(t *testing.T, label string, c *CardView) {
	t.Helper()
	if c.CastableHere || len(c.CastPrices) > 0 {
		t.Errorf("%s: a seat with no permission got castable_here=%v cast_prices=%+v",
			label, c.CastableHere, c.CastPrices)
	}
	for i := range c.Faces {
		if c.Faces[i].CastableHere || len(c.Faces[i].CastPrices) > 0 {
			t.Errorf("%s: faces[%d] carries the holder's answer", label, i)
		}
	}
}

// Impulse exile: the printed cost, so no badge — and castable now.
func TestExileStripImpulseChargesThePrintedCost(t *testing.T) {
	g, me, opp := stripTable(t)
	id := exileWithGrant(t, g, exiledSpell(opp.ID, "Impulse Sorcery", "Sorcery", "{1}{R}"),
		game.CastPermission{Player: me.ID})

	c := stripCard(t, g, me.ID.String(), id)
	if !c.CastableHere {
		t.Error("an impulse-exiled sorcery in its holder's main phase is castable now")
	}
	assertPrice(t, "impulse", c, "{1}{R}", true, "")
	assertNothingFor(t, "impulse/owner", stripCard(t, g, opp.ID.String(), id))
	assertNothingFor(t, "impulse/spectator", stripCard(t, g, "", id))
}

// Impulse exile of a sorcery outside the holder's main phase: the card
// keeps its price (it is still theirs to cast later this turn) and
// loses the bit.
func TestExileStripSorceryOutsideTheMainPhaseIsNotCastableNow(t *testing.T) {
	g, me, opp := stripTable(t)
	id := exileWithGrant(t, g, exiledSpell(opp.ID, "Late Sorcery", "Sorcery", "{2}"),
		game.CastPermission{Player: me.ID})
	advanceTo(t, g, game.StepEnd)

	c := stripCard(t, g, me.ID.String(), id)
	if c.CastableHere {
		t.Error("a sorcery is not castable in the end step (CR 307.1)")
	}
	assertPrice(t, "impulse/end step", c, "{2}", true, "")
}

// Airbend: the permission's own {2} replaces the printed cost.
func TestExileStripAirbendChargesTheOverride(t *testing.T) {
	g, me, _ := stripTable(t)
	id := exileWithGrant(t, g, exiledSpell(me.ID, "Airbent Giant", "Creature — Giant", "{5}{G}"),
		game.CastPermission{Player: me.ID, Cost: "{2}", Duration: game.WhileInZoneDuration()})

	c := stripCard(t, g, me.ID.String(), id)
	if !c.CastableHere {
		t.Error("an airbent creature is castable in its owner's main phase")
	}
	assertPrice(t, "airbend", c, "{2}", false, "")
}

// Airbend on a card whose printed cost IS {2}: nothing to tell the
// player, so no badge.
func TestExileStripAirbendOnATwoDropIsThePrintedCost(t *testing.T) {
	g, me, _ := stripTable(t)
	id := exileWithGrant(t, g, exiledSpell(me.ID, "Two Drop", "Creature — Bear", "{2}"),
		game.CastPermission{Player: me.ID, Cost: "{2}", Duration: game.WhileInZoneDuration()})
	assertPrice(t, "airbend {2} on {2}", stripCard(t, g, me.ID.String(), id), "{2}", true, "")
}

// Warp: on the turn it was warped the permission is not live yet, so
// the holder gets neither the bit nor a price — only the public
// `exile_play.not_before_turn` the strip dims the card by. On the
// later turn both arrive.
func TestExileStripWarpWaitsForTheLaterTurn(t *testing.T) {
	g, me, _ := stripTable(t)
	id := exileWithGrant(t, g, exiledSpell(me.ID, "Warped Drake", "Creature — Drake", "{3}{U}"),
		game.CastPermission{
			Player: me.ID, NotBeforeSeq: g.Turn.Seq + 1,
			Duration: game.WhileInZoneDuration(), CastOnly: true,
		})

	c := stripCard(t, g, me.ID.String(), id)
	if c.CastableHere || len(c.CastPrices) > 0 {
		t.Errorf("warp before its turn: castable_here=%v cast_prices=%+v, want neither",
			c.CastableHere, c.CastPrices)
	}
	if c.ExilePlay == nil || c.ExilePlay.NotBeforeSeq != g.Turn.Seq+1 {
		t.Fatalf("exile_play = %+v, want the pending grant with its floor", c.ExilePlay)
	}

	g.Turn.Seq++
	c = stripCard(t, g, me.ID.String(), id)
	if !c.CastableHere {
		t.Error("warp on the later turn: castable now")
	}
	assertPrice(t, "warp", c, "{3}{U}", true, "")
}

// Plot, through the engine's own grant: free, from a later turn, at
// sorcery speed only.
func TestExileStripPlotIsFreeOnALaterTurnInTheMainPhase(t *testing.T) {
	g, me, _ := stripTable(t)
	card := exiledSpell(me.ID, "Plotted Bolt", "Instant", "{R}")
	card.KnownBy = map[uuid.UUID]bool{}
	for _, p := range g.Seats {
		card.KnownBy[p.ID] = true
	}
	g.Exile.PushTop(card)
	g.WithWriteLock(func() { g.PlotExiledCardForEffect(card.InstanceID, uuid.Nil) })
	id := card.InstanceID

	if c := stripCard(t, g, me.ID.String(), id); c.CastableHere || len(c.CastPrices) > 0 {
		t.Errorf("plotted this turn: castable_here=%v cast_prices=%+v, want neither", c.CastableHere, c.CastPrices)
	}

	g.Turn.Seq++
	c := stripCard(t, g, me.ID.String(), id)
	if !c.CastableHere {
		t.Error("a plotted card on a later turn, main phase, empty stack: castable now")
	}
	assertPrice(t, "plot", c, "{0}", false, "")

	// CR 702.170d: the plot window is the main phase, even for an
	// instant. The price stays; the bit goes.
	advanceTo(t, g, game.StepEnd)
	c = stripCard(t, g, me.ID.String(), id)
	if c.CastableHere {
		t.Error("a plotted instant is not castable in the end step (CR 702.170d)")
	}
	assertPrice(t, "plot/end step", c, "{0}", false, "")
}

// Foretell, through the engine's own special action: the card is face
// down, its price is the foretell cost, and only its owner may see
// either.
func TestExileStripForetellChargesTheForetellCostAndStaysHidden(t *testing.T) {
	g, me, opp := stripTable(t)
	const oracle = "test-strip-foretell"
	installSpecialAction(t, oracle, game.SpecialAction{
		Kind: game.SpecialActionForetell, Cost: game.ForetellExileCost, CastCost: "{1}{U}", Label: "Foretell {2}",
	})
	card := game.NewCard("Saw It Coming", me.ID)
	card.TypeLine = "Instant"
	card.ManaCost = "{1}{U}{U}"
	card.OracleID = oracle
	card.KnownBy = map[uuid.UUID]bool{me.ID: true}
	me.Hand.PushTop(card)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.PerformSpecialAction(me.ID, card.InstanceID, game.SpecialActionForetell,
		game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("foretell: %v", err)
	}
	id := card.InstanceID

	if c := stripCard(t, g, me.ID.String(), id); c.CastableHere || len(c.CastPrices) > 0 {
		t.Errorf("foretold this turn: castable_here=%v cast_prices=%+v, want neither", c.CastableHere, c.CastPrices)
	}

	g.Turn.Seq++
	mine := stripCard(t, g, me.ID.String(), id)
	if !mine.FaceVisible || mine.Name != "Saw It Coming" {
		t.Errorf("the owner knows their foretold card: face_visible=%v name=%q", mine.FaceVisible, mine.Name)
	}
	if !mine.CastableHere {
		t.Error("a foretold instant on a later turn is castable now")
	}
	assertPrice(t, "foretell", mine, "{1}{U}", false, game.AltCostKeyForetell)

	theirs := stripCard(t, g, opp.ID.String(), id)
	if theirs.Name != "" || theirs.FaceVisible || theirs.ManaCost != "" {
		t.Errorf("an opponent reads the foretold card: name=%q mana_cost=%q", theirs.Name, theirs.ManaCost)
	}
	assertNothingFor(t, "foretell/opponent", theirs)
	// The admin / replay frame ("") sees every card by design, but it
	// is nobody's seat, so nobody's price.
	assertNothingFor(t, "foretell/unseated", stripCard(t, g, "", id))
}

// Adventure (CR 715.4): the creature half, at its printed cost — the
// price is read off the FACE the grant opens, not the card's front.
func TestExileStripAdventureChargesTheCreatureHalf(t *testing.T) {
	g, me, opp := stripTable(t)
	c := adventureCard(me.ID, me.ID, opp.ID)
	c.Faces[1].ManaCost = "{R}"
	g.Exile.PushTop(c)
	g.GrantCastPermissionOverCardForEffect(c.InstanceID, game.CastPermission{
		Player: me.ID, Duration: game.WhileInZoneDuration(), CastOnly: true, Faces: []int{0},
	})

	v := stripCard(t, g, me.ID.String(), c.InstanceID)
	if !v.CastableHere {
		t.Error("an exiled adventurer is castable in its owner's main phase")
	}
	assertPrice(t, "adventure", v, "{2}{R}", true, "")
	assertNothingFor(t, "adventure/opponent", stripCard(t, g, opp.ID.String(), c.InstanceID))
}

// Prepare (ADR 0090): the copy of the prepare spell sits in exile and
// its controller casts it at the spell's printed cost.
func TestExileStripPrepareCopyChargesThePrepareSpell(t *testing.T) {
	g, me, opp := stripTable(t)
	c := game.NewCard("Wire Conductor", me.ID)
	c.OracleID = "strip-prepare-oracle"
	c.Layout = game.LayoutPrepare
	c.Faces = []game.Face{
		{Name: "Wire Conductor", TypeLine: "Creature — Bird Pilot", ManaCost: "{2}{U}", Power: 2, Toughness: 3},
		{Name: "Wire Aboard", TypeLine: "Instant", ManaCost: "{U}"},
	}
	c.SetFace(0)
	c.EnteredBattlefieldAt = 1
	c.KnownBy = map[uuid.UUID]bool{me.ID: true, opp.ID: true}
	g.Battlefield.PushTop(c)
	g.WithWriteLock(func() {
		if ok, err := g.BecomePreparedForEffect(c.InstanceID); !ok || err != nil {
			t.Fatalf("BecomePreparedForEffect = %v, %v", ok, err)
		}
	})
	var copyID uuid.UUID
	for _, e := range g.Exile.Cards {
		if e.PrepareCopy {
			copyID = e.InstanceID
		}
	}

	v := stripCard(t, g, me.ID.String(), copyID)
	if !v.CastableHere {
		t.Error("the prepare copy is castable by the prepared permanent's controller")
	}
	assertPrice(t, "prepare", v, "{U}", true, "")
	assertNothingFor(t, "prepare/opponent", stripCard(t, g, opp.ID.String(), copyID))
}

// ADR 0066: a granted alternative cost over an exiled card — a
// Snapcaster-shaped flashback pointed at exile. The price claims the
// synthesised offer and is still the printed mana.
func TestExileStripGrantedAlternativeCostClaimsTheOffer(t *testing.T) {
	g, me, _ := stripTable(t)
	id := exileWithGrant(t, g, exiledSpell(me.ID, "Granted Think", "Instant", "{2}{U}"),
		game.CastPermission{Player: me.ID, AltCostKey: "flashback", Label: "Flashback", ExileOnResolution: true})

	c := stripCard(t, g, me.ID.String(), id)
	if !c.CastableHere {
		t.Error("a granted-flashback instant is castable now")
	}
	assertPrice(t, "granted flashback", c, "{2}{U}", true, "flashback")
	if c.CastPrices[0].Label != "Flashback" {
		t.Errorf("label = %q, want the offer's own", c.CastPrices[0].Label)
	}
}

// CR 601.2f: the price is AFTER the board's modifiers. A Thalia-shaped
// tax turns a free plotted cast into {1}, and the engine then charges
// exactly that — the badge and the payment are one number.
func TestExileStripPriceIsAfterCostModifiersAndIsWhatTheCastCharges(t *testing.T) {
	g, me, _ := stripTable(t)
	const taxOracle = "test-strip-tax"
	installChargedCostModifier(t, taxOracle, game.CostModifier{
		Kind:   game.CostIncrease,
		Label:  "Spells cost {1} more to cast.",
		Amount: func(game.CostQuery) int { return 1 },
	})
	tax := game.NewCard("Tax Collector", me.ID)
	tax.TypeLine = "Artifact"
	tax.OracleID = taxOracle
	g.Battlefield.PushTop(tax)

	card := exiledSpell(me.ID, "Plotted Sorcery", "Sorcery", "{3}{B}")
	id := exileWithGrant(t, g, card, game.CastPermission{
		Player: me.ID, Cost: "{0}", Timing: game.TimingPlot,
		Duration: game.WhileInZoneDuration(), CastOnly: true,
	})

	c := stripCard(t, g, me.ID.String(), id)
	assertPrice(t, "plot under a tax", c, "{1}", false, "")

	me.ManaPool = nil
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("casting at the badge's price: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("the cast charged less than the badge said: %d mana left", len(me.ManaPool))
	}
}

// A discount is a price too: the airbend {2} under a {1} reduction
// reads {1}, and the printed flag stays off.
func TestExileStripDiscountReachesTheOverride(t *testing.T) {
	g, me, _ := stripTable(t)
	const oracle = "test-strip-discount"
	installChargedCostModifier(t, oracle, game.CostModifier{
		Kind:  game.CostReduction,
		Label: "Spells you cast from exile cost {1} less to cast.",
		AppliesTo: func(q game.CostQuery) bool {
			return q.FromZone == game.ZoneExile
		},
		Amount: func(game.CostQuery) int { return 1 },
	})
	src := game.NewCard("Exile Discounter", me.ID)
	src.TypeLine = "Enchantment"
	src.OracleID = oracle
	g.Battlefield.PushTop(src)

	id := exileWithGrant(t, g, exiledSpell(me.ID, "Airbent Beast", "Creature — Beast", "{4}{G}"),
		game.CastPermission{Player: me.ID, Cost: "{2}", Duration: game.WhileInZoneDuration()})
	assertPrice(t, "airbend under a discount", stripCard(t, g, me.ID.String(), id), "{1}", false, "")
}

// A land under a "you may PLAY it" grant is castable-now by the land
// rule and carries no price; under a cast-only grant it is stranded.
func TestExileStripLandPlays(t *testing.T) {
	g, me, opp := stripTable(t)
	land := exiledSpell(opp.ID, "Stolen Forest", "Basic Land — Forest", "")
	played := exileWithGrant(t, g, land, game.CastPermission{Player: me.ID})
	stranded := exileWithGrant(t, g, exiledSpell(opp.ID, "Stranded Island", "Basic Land — Island", ""),
		game.CastPermission{Player: me.ID, CastOnly: true})

	c := stripCard(t, g, me.ID.String(), played)
	if g.LandDropsRemainingFor(me.ID) > 0 && !c.CastableHere {
		t.Error("a play-granted land with a land drop left is playable now")
	}
	if len(c.CastPrices) > 0 {
		t.Errorf("a land is played, not cast, and has no price: %+v", c.CastPrices)
	}
	if s := stripCard(t, g, me.ID.String(), stranded); s.CastableHere {
		t.Error("a cast-only grant strands a land (CR 305.1)")
	}
}
