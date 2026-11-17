package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bro_b21_test.go — the eight Brothers' War prototype cards with rules-heavy
// text, batch 21 of S82 (#5482): each cast both ways through the engine,
// and each clause that reads "this creature's power" at both sizes.

var b21Rows = []protoRow{
	{"7ff8940f-d798-41eb-b82e-a1eba2163fe7", "Steel Seraph", "Artifact Creature — Angel", "{6}", 5, 4,
		"Prototype {1}{W}{W} — 3/3" + protoReminder + "\nFlying\nAt the beginning of combat on your turn, target creature you control gains your choice of flying, vigilance, or lifelink until end of turn.",
		"{1}{W}{W}", "W", 3, 3, []string{"flying"}},
	{"d3a5a830-cd14-49da-9412-c50049c74c92", "Phyrexian Fleshgorger", "Artifact Creature — Phyrexian Wurm", "{7}", 7, 5,
		"Prototype {1}{B}{B} — 3/3" + protoReminder + "\nMenace, lifelink\nWard—Pay life equal to this creature's power.",
		"{1}{B}{B}", "B", 3, 3, []string{"menace", "lifelink"}},
	{"7644148f-3b26-447d-9f52-28c94c111b69", "Rootwire Amalgam", "Artifact Creature — Golem", "{5}", 5, 5,
		"Prototype {1}{G} — 2/3" + protoReminder + "\n{3}{G}{G}, Sacrifice this creature: Create an X/X colorless Golem artifact creature token, where X is three times this creature's power. It gains haste until end of turn. Activate only as a sorcery.",
		"{1}{G}", "G", 2, 3, nil},
	{"124c3af4-8545-479a-b44c-446da1c8b535", "Woodcaller Automaton", "Artifact Creature — Construct", "{10}", 8, 8,
		"Prototype {2}{G}{G} — 3/3" + protoReminder + "\nWhen this creature enters, if you cast it, untap target land you control. It becomes a Treefolk creature with haste and base power and toughness equal to this creature's power and toughness. It's still a land.",
		"{2}{G}{G}", "G", 3, 3, nil},
	{"ff1dd9c1-d1e5-49dd-b8b2-df3e4a41d3e5", "Skitterbeam Battalion", "Artifact Creature — Construct", "{9}", 4, 4,
		"Prototype {3}{R}{R} — 2/2" + protoReminder + "\nTrample, haste\nWhen this creature enters, if you cast it, create two tokens that are copies of it.",
		"{3}{R}{R}", "R", 2, 2, []string{"trample", "haste"}},
	{"792f8235-f23b-4858-bdd4-c7a186b7a470", "Arcane Proxy", "Artifact Creature — Wizard", "{7}", 4, 3,
		"Prototype {1}{U}{U} — 2/1" + protoReminder + "\nWhen this creature enters, if you cast it, exile target instant or sorcery card with mana value less than or equal to this creature's power from your graveyard. Copy that card. You may cast the copy without paying its mana cost.",
		"{1}{U}{U}", "U", 2, 1, nil},
	{"b2b09dc3-8239-4f3b-91d2-e7cbbdc062f7", "Hulking Metamorph", "Artifact Creature — Shapeshifter", "{9}", 7, 7,
		"Prototype {2}{U}{U} — 3/3" + protoReminder + "\nYou may have this creature enter as a copy of an artifact or creature you control, except it's an artifact creature in addition to its other types, and its power and toughness are equal to this creature's power and toughness.",
		"{2}{U}{U}", "U", 3, 3, nil},
}

func b21Row(t *testing.T, name string) protoRow {
	t.Helper()
	for _, r := range b21Rows {
		if r.name == name {
			return r
		}
	}
	t.Fatalf("no row for %s", name)
	return protoRow{}
}

// b21Named lists the battlefield cards with this name, tokens included.
func b21Named(g *game.Game, name string) []game.Card {
	var out []game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name {
				out = append(out, c)
			}
		}
	})
	return out
}

func TestS82BroB21CardsAreRegisteredFull(t *testing.T) {
	for _, r := range b21Rows {
		spec, ok := Lookup(r.oracle)
		if !ok {
			t.Errorf("%s is not registered", r.name)
			continue
		}
		if spec.Name != r.name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", r.name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
	if spec, ok := Lookup("b1870fff-b87e-434e-8d4c-d9e60f7f4284"); !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Mask of the Jadecrafter: %+v", spec)
	}
}

// Every prototype card, cast both ways: the inset's cost, colour and size
// when prototyped, the printed ones otherwise, keywords and types in both.
// Cards that turn an enters trigger or a copy prompt on are cast with
// nothing for them to act on, and the prompt is left unanswered.
func TestS82BroB21PrototypeCardsCastBothWays(t *testing.T) {
	for _, r := range b21Rows {
		for _, prototyped := range []bool{true, false} {
			g := newCatalogGame(t)
			id := castProtoRow(t, g, r, prototyped)
			perm, ok := g.LookupCardForEffect(id)
			if !ok {
				continue
			}
			eff := perm.Effective()
			wantCost, wantP, wantT, wantColors := r.cost, r.power, r.toughness, ""
			if prototyped {
				wantCost, wantP, wantT, wantColors = r.insetCost, r.insetPower, r.insetToughness, r.insetColor
			}
			if r.name != "Hulking Metamorph" && (perm.Prototyped() != prototyped || perm.ManaCost != wantCost || eff.Power != wantP ||
				eff.Toughness != wantT || strings.Join(eff.Colors, "") != wantColors) {
				t.Errorf("%s (prototyped=%v) = %s %d/%d %v, want %s %d/%d %q",
					r.name, prototyped, perm.ManaCost, eff.Power, eff.Toughness, eff.Colors, wantCost, wantP, wantT, wantColors)
			}
			for _, kw := range r.keywords {
				if !game.HasKeyword(&perm, kw) {
					t.Errorf("%s (prototyped=%v) has no %s", r.name, prototyped, kw)
				}
			}
		}
	}
}

// Steel Seraph: at the beginning of combat on its controller's turn, the
// chosen creature gains the keyword the controller picks.
func TestS82BroB21SteelSeraphGrantsTheChosenKeyword(t *testing.T) {
	for i, kw := range []string{"flying", "vigilance", "lifelink"} {
		g := newCatalogGame(t)
		me, _, _ := blbB1Seats(g)
		pushProtoRow(t, g, me.ID, b21Row(t, "Steel Seraph"), true)
		bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
		theirs := blbB1Push(g, g.Seats[(g.Turn.ActiveSeat+1)%4].ID, "Wolf", "Creature — Wolf", "", 2, 2)
		advanceTo(t, g, game.StepBeginCombat)
		ch := pickTargetPrompt(t, g)
		if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
			t.Fatal("an opponent's creature was accepted as the target")
		}
		answerPickTarget(t, g, bear)
		passPriorityAroundTable(t, g)
		answerOptionPick(t, g, me.ID, i)
		passPriorityAroundTable(t, g)
		for _, other := range []string{"flying", "vigilance", "lifelink"} {
			if got, want := hasEffectiveKeyword(t, g, bear, other), other == kw; got != want {
				t.Errorf("choice %s: bear has %s = %v, want %v", kw, other, got, want)
			}
		}
	}
}

func TestS82BroB21SteelSeraphDoesNotTriggerOnAnOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	pushProtoRow(t, g, me.ID, b21Row(t, "Steel Seraph"), false)
	blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	endTurn(t, g)
	advanceTo(t, g, game.StepBeginCombat)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			t.Fatal("Steel Seraph asked for a target at an opponent's beginning of combat")
		}
	}
}

// Phyrexian Fleshgorger: ward costs life equal to its power as it is paid.
func TestS82BroB21FleshgorgerWardCostsLifeEqualToItsPower(t *testing.T) {
	for _, tc := range []struct {
		prototyped bool
		want       int
	}{{true, 3}, {false, 7}} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		id := pushProtoRow(t, g, me.ID, b21Row(t, "Phyrexian Fleshgorger"), tc.prototyped)
		advanceTo(t, g, game.StepPrecombatMain)
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
		bolt := uuid.New()
		opp.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}",
			OracleID: lightningBoltOracle, Owner: opp.ID, Controller: opp.ID})
		floatMana(t, g, opp, "{R}")
		if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: id}},
		}); err != nil {
			t.Fatalf("a warded permanent is a legal target: %v", err)
		}
		life := opp.Life
		prompt := passUntilConfirmFor(t, g, opp.ID)
		if prompt == nil || prompt.LifeCost != tc.want {
			t.Fatalf("prototyped=%v: ward prompt %+v, want a %d-life cost", tc.prototyped, prompt, tc.want)
		}
		if err := g.ResolveConfirm(prompt.ID, opp.ID, true); err != nil {
			t.Fatalf("pay: %v", err)
		}
		passPriorityAroundTable(t, g)
		if opp.Life != life-tc.want {
			t.Errorf("prototyped=%v: life %d -> %d, want -%d", tc.prototyped, life, opp.Life, tc.want)
		}
	}
}

// Rootwire Amalgam: a Golem three times its power (6 or 15), with haste,
// at sorcery speed only, and the Amalgam is gone.
func TestS82BroB21RootwireAmalgamMakesAGolemThreeTimesItsPower(t *testing.T) {
	for _, tc := range []struct {
		prototyped bool
		want       int
	}{{true, 6}, {false, 15}} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		id := pushProtoRow(t, g, me.ID, b21Row(t, "Rootwire Amalgam"), tc.prototyped)
		advanceTo(t, g, game.StepPrecombatMain)
		me.ManaPool = nil
		if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
			t.Fatal("activated with no mana")
		}
		floatMana(t, g, me, "{C}{C}{C}{G}{G}")
		b16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{Strict: true})
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(id) {
			t.Error("the Amalgam was not sacrificed")
		}
		golems := b21Named(g, "Golem")
		if len(golems) != 1 {
			t.Fatalf("%d Golems, want 1", len(golems))
		}
		if p, tough := ptOf(t, g, golems[0].InstanceID); p != tc.want || tough != tc.want {
			t.Errorf("prototyped=%v: Golem is %d/%d, want %d/%d", tc.prototyped, p, tough, tc.want, tc.want)
		}
		if !golems[0].IsArtifact() || !golems[0].IsCreature() || len(golems[0].Colors) != 0 {
			t.Errorf("Golem is %q %v, want a colorless artifact creature", golems[0].TypeLine, golems[0].Colors)
		}
		if !hasEffectiveKeyword(t, g, golems[0].InstanceID, "haste") {
			t.Error("the Golem has no haste")
		}
	}
}

func TestS82BroB21RootwireAmalgamIsSorceryOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := pushProtoRow(t, g, me.ID, b21Row(t, "Rootwire Amalgam"), true)
	advanceTo(t, g, game.StepBeginCombat)
	floatMana(t, g, me, "{C}{C}{C}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated outside the main phase")
	}
}

// Mask of the Jadecrafter: X is announced and sets the Golem's size.
func TestS82BroB21MaskOfTheJadecrafterMakesAnXGolem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mask := pushCatalogPermanent(g, me.ID, "Mask of the Jadecrafter", "Artifact", "b1870fff-b87e-434e-8d4c-d9e60f7f4284", false)
	advanceTo(t, g, game.StepPrecombatMain)
	me.ManaPool = nil
	floatMana(t, g, me, "{C}{C}{C}")
	b16Activate(t, g, me.ID, mask, 0, game.ActivateAbilityParams{Strict: true, XValue: 3})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mask) {
		t.Error("the Mask was not sacrificed")
	}
	golems := b21Named(g, "Golem")
	if len(golems) != 1 {
		t.Fatalf("%d Golems, want 1", len(golems))
	}
	if p, tough := ptOf(t, g, golems[0].InstanceID); p != 3 || tough != 3 {
		t.Errorf("Golem is %d/%d, want 3/3", p, tough)
	}
	if !golems[0].IsArtifact() || !golems[0].IsCreature() {
		t.Errorf("Golem is %q", golems[0].TypeLine)
	}
}

func TestS82BroB21MaskOfTheJadecrafterIsSorceryOnlyAndNeedsItsMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mask := pushCatalogPermanent(g, me.ID, "Mask of the Jadecrafter", "Artifact", "b1870fff-b87e-434e-8d4c-d9e60f7f4284", false)
	advanceTo(t, g, game.StepPrecombatMain)
	me.ManaPool = nil
	if err := g.ActivateCatalogAbility(me.ID, mask, 0, game.ActivateAbilityParams{Strict: true, XValue: 2}); err == nil {
		t.Error("activated for X=2 with no mana")
	}
	advanceTo(t, g, game.StepBeginCombat)
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, mask, 0, game.ActivateAbilityParams{Strict: true, XValue: 2}); err == nil {
		t.Error("activated outside the main phase")
	}
}

// Unearth {2}{G}: from the graveyard, back to the battlefield, and exiled
// at the end step.
func TestS82BroB21MaskOfTheJadecrafterUnearths(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Mask of the Jadecrafter", TypeLine: "Artifact", ManaCost: "{2}",
		OracleID: "b1870fff-b87e-434e-8d4c-d9e60f7f4284", Owner: me.ID, Controller: me.ID})
	advanceTo(t, g, game.StepPrecombatMain)
	me.ManaPool = nil
	floatMana(t, g, me, "{C}{C}{G}")
	b16Activate(t, g, me.ID, id, 1, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Mask did not return from the graveyard")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) {
		t.Error("the unearthed Mask is still on the battlefield after the end step")
	}
}

// Woodcaller Automaton: the land untaps and becomes a Treefolk creature
// with haste, as big as the Automaton (3/3 or 8/8), and is still a land.
func TestS82BroB21WoodcallerAutomatonAnimatesALand(t *testing.T) {
	for _, tc := range []struct {
		prototyped bool
		want       int
	}{{true, 3}, {false, 8}} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		land := pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", true)
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == land {
					g.Battlefield.Cards[i].Tapped = true
				}
			}
		})
		castProtoRow(t, g, b21Row(t, "Woodcaller Automaton"), tc.prototyped)
		answerPickTargetIfOpen(t, g, land)
		passPriorityAroundTable(t, g)
		if b16Tapped(t, g, land) {
			t.Errorf("prototyped=%v: the land did not untap", tc.prototyped)
		}
		perm, _ := g.LookupCardForEffect(land)
		eff := perm.Effective()
		if !perm.IsLand() || !perm.IsCreature() || !perm.HasSubtype("Treefolk") || !perm.HasSubtype("Forest") {
			t.Errorf("prototyped=%v: the land is %v %v, want a Treefolk creature that is still a Forest land", tc.prototyped, eff.Types, eff.Subtypes)
		}
		if p, tough := ptOf(t, g, land); p != tc.want || tough != tc.want {
			t.Errorf("prototyped=%v: the land is %d/%d, want %d/%d", tc.prototyped, p, tough, tc.want, tc.want)
		}
		if !hasEffectiveKeyword(t, g, land, "haste") {
			t.Errorf("prototyped=%v: no haste", tc.prototyped)
		}
		// It lasts: the next turn it is still a creature.
		advanceTo(t, g, game.StepEnd)
		if perm, _ = g.LookupCardForEffect(land); !perm.IsCreature() {
			t.Errorf("prototyped=%v: the land stopped being a creature at end of turn", tc.prototyped)
		}
	}
}

// "If you cast it": an Automaton that was put onto the battlefield does nothing.
func TestS82BroB21WoodcallerAutomatonNeedsToHaveBeenCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", true)
	g.WithWriteLock(func() {
		c := b21Row(t, "Woodcaller Automaton").card(t, me.ID)
		c.Provenance = game.CastProvenance{}
		g.Battlefield.PushTop(c)
	})
	_ = land
	g.RunStateChecksForTest()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			t.Fatal("an Automaton that was not cast asked for a target")
		}
	}
}

// Skitterbeam Battalion: two copies, the size of the one that was cast.
func TestS82BroB21SkitterbeamBattalionMakesTwoCopies(t *testing.T) {
	for _, tc := range []struct {
		prototyped bool
		want       int
	}{{true, 2}, {false, 4}} {
		g := newCatalogGame(t)
		castProtoRow(t, g, b21Row(t, "Skitterbeam Battalion"), tc.prototyped)
		passPriorityAroundTable(t, g)
		all := b21Named(g, "Skitterbeam Battalion")
		if len(all) != 3 {
			t.Fatalf("prototyped=%v: %d Battalions, want the original and two tokens", tc.prototyped, len(all))
		}
		tokens := 0
		for _, c := range all {
			if p, tough := ptOf(t, g, c.InstanceID); p != tc.want || tough != tc.want {
				t.Errorf("prototyped=%v: a Battalion is %d/%d, want %d/%d", tc.prototyped, p, tough, tc.want, tc.want)
			}
			for _, kw := range []string{"trample", "haste"} {
				if !hasEffectiveKeyword(t, g, c.InstanceID, kw) {
					t.Errorf("prototyped=%v: a Battalion has no %s", tc.prototyped, kw)
				}
			}
			if c.IsToken() {
				tokens++
			}
		}
		if tokens != 2 {
			t.Errorf("prototyped=%v: %d tokens, want 2", tc.prototyped, tokens)
		}
	}
}

// The tokens were not cast, so they do not copy themselves again.
func TestS82BroB21SkitterbeamBattalionNeedsToHaveBeenCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() {
		c := b21Row(t, "Skitterbeam Battalion").card(t, me.ID)
		c.Provenance = game.CastProvenance{}
		g.Battlefield.PushTop(c)
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if n := len(b21Named(g, "Skitterbeam Battalion")); n != 1 {
		t.Errorf("%d Battalions, want only the one that was put onto the battlefield", n)
	}
}

// Arcane Proxy: the target's mana value is capped by the Proxy's power, and
// the copy is cast for free.
func TestS82BroB21ArcaneProxyCopiesAnInstantForFree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bolt := gyCard(me, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
	big := gyCard(me, "Divination", "Sorcery", "{2}{U}", divinationOracle)
	castProtoRow(t, g, b21Row(t, "Arcane Proxy"), true)
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the Proxy's trigger")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: big}}); err == nil {
		t.Fatal("a mana value 3 card was accepted for a 2-power Proxy")
	}
	answerTriggerTargets(t, g, me.ID, bolt)
	passPriorityAroundTable(t, g)
	if !inExile(g, bolt) {
		t.Fatal("the target was not exiled")
	}
	cp := copyOf(t, g, "Lightning Bolt")
	hand := me.Hand.Size()
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	life := opp.Life
	me.ManaPool = nil
	if err := castCopy(g, me.ID, cp.InstanceID, true, playerTarget(opp.ID)); err != nil {
		t.Fatalf("cast the copy for free: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d -> %d, want the copied Lightning Bolt to deal 3", life, opp.Life)
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d", hand, me.Hand.Size())
	}
	if !inExile(g, bolt) {
		t.Error("the original stays in exile")
	}
}

func TestS82BroB21ArcaneProxyUnprototypedReachesManaValueFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	big := gyCard(me, "Divination", "Sorcery", "{2}{U}", divinationOracle)
	castProtoRow(t, g, b21Row(t, "Arcane Proxy"), false)
	passPriorityAroundTable(t, g)
	answerTriggerTargets(t, g, me.ID, big)
	passPriorityAroundTable(t, g)
	if !inExile(g, big) {
		t.Error("a mana value 3 card is a legal target for a 4-power Proxy")
	}
}

// Hulking Metamorph: it copies an artifact or creature you control, is an
// artifact creature too, and is the size of the Metamorph that was cast.
func TestS82BroB21HulkingMetamorphCopiesWithItsOwnSize(t *testing.T) {
	for _, tc := range []struct {
		prototyped bool
		want       int
	}{{true, 3}, {false, 7}} {
		g := newCatalogGame(t)
		me, opp, _ := blbB1Seats(g)
		bears := seedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
		theirs := seedCopyableCreature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
		id := castProtoRowUntilPrompt(t, g, b21Row(t, "Hulking Metamorph"), tc.prototyped)
		c := copyPrompt(g)
		if c == nil {
			t.Fatalf("prototyped=%v: no copy prompt", tc.prototyped)
		}
		offered := map[uuid.UUID]bool{}
		for _, o := range c.CopyOptions {
			offered[o] = true
		}
		if !offered[bears] || offered[theirs] {
			t.Fatalf("prototyped=%v: copy options %v, want only the creature I control", tc.prototyped, c.CopyOptions)
		}
		if err := g.ResolveCopyTarget(c.ID, c.Chooser, bears); err != nil {
			t.Fatalf("ResolveCopyTarget: %v", err)
		}
		passPriorityAroundTable(t, g)
		got := copyBattlefieldCard(t, g, id)
		if got.Name != "Grizzly Bears" || !got.IsArtifact() || !got.IsCreature() || !got.HasSubtype("Bear") {
			t.Errorf("prototyped=%v: entered as %q %q, want an artifact creature Grizzly Bears", tc.prototyped, got.Name, got.TypeLine)
		}
		if p, tough := ptOf(t, g, id); p != tc.want || tough != tc.want {
			t.Errorf("prototyped=%v: the copy is %d/%d, want %d/%d", tc.prototyped, p, tough, tc.want, tc.want)
		}
	}
}

func TestS82BroB21HulkingMetamorphMayDecline(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	id := castProtoRowUntilPrompt(t, g, b21Row(t, "Hulking Metamorph"), true)
	c := copyPrompt(g)
	if c == nil {
		t.Fatal("no copy prompt")
	}
	if err := g.ResolveCopyTarget(c.ID, c.Chooser, uuid.Nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := copyBattlefieldCard(t, g, id); got.Name != "Hulking Metamorph" {
		t.Errorf("declined, but it entered as %q", got.Name)
	}
}

// castProtoRowUntilPrompt casts the row and passes priority until the
// copy prompt of its entry opens, without answering it.
func castProtoRowUntilPrompt(t *testing.T, g *game.Game, r protoRow, prototyped bool) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	c := r.card(t, me.ID)
	me.Hand.PushTop(c)
	advanceTo(t, g, game.StepPrecombatMain)
	params := game.CastSpellParams{Strict: true}
	cost := r.cost
	if prototyped {
		params.Face = game.PrototypeFace
		cost = r.insetCost
	}
	me.ManaPool = nil
	floatMana(t, g, me, protoManaFor(t, cost))
	if err := g.CastSpell(me.ID, c.InstanceID, params); err != nil {
		t.Fatalf("cast %s: %v", r.name, err)
	}
	for i := 0; i < 8 && copyPrompt(g) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	return c.InstanceID
}
