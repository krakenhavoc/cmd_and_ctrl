package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// optional_mode_count_test.go — #1655, ADR 0065's 2026-09-28
// amendment: a mode count that reads the optional costs announced
// WITH the modes (CR 601.2b — Inscription of Ruin's kicker, Depth
// Defiler's), and one that forces the higher count (Prophetic Titan's
// delirium, no "may"). Every refusal below is paired with an
// acceptance on the same board, so a refusal can only be the mode
// count talking.

const (
	inscriptionOfRuinOracle      = "760e8561-4ec6-4594-ba5d-f79cb9f25fd0"
	inscriptionOfAbundanceOracle = "a5e28749-18ea-4a2b-b7d9-905cf2913d4e"
	depthDefilerOracle           = "8ca4ca66-30b1-4074-a2e3-545b7682381b"
	propheticTitanOracle         = "224f5f1a-4f31-4935-bf5a-910fd0a666a1"
	letsPlayAGameOracle          = "dd893746-e8bd-49fa-a1ce-755d5bd4f513"
	wailOfTheForgottenOracle     = "030b5408-f216-43e4-8593-f78d22821876"
)

// ruinBoard seeds a target for each of Inscription of Ruin's three
// bullets: the opponent, a two-drop creature card in your graveyard
// and a three-drop creature on the opponent's side.
type ruinBoard struct {
	me, opp *game.Player
	dead    uuid.UUID
	theirs  uuid.UUID
}

func newRuinBoard(t *testing.T) (*game.Game, ruinBoard) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	b := ruinBoard{me: me, opp: opp}
	b.dead = b17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	b.theirs = uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: b.theirs, Name: "Their Ogre", TypeLine: "Creature — Ogre",
		ManaCost: "{2}{R}", Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID})
	for i := 0; i < 3; i++ {
		opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	}
	return g, b
}

// ruinTargets is the target list for `modes`, each bullet's pick
// stamped with its occurrence.
func (b ruinBoard) targets(modes []int) []game.TargetRef {
	var out []game.TargetRef
	for occ, m := range modes {
		switch m {
		case 0:
			out = append(out, game.TargetRef{Kind: game.TargetPlayer, ID: b.opp.ID, Mode: occ})
		case 1:
			out = append(out, game.TargetRef{Kind: game.TargetCard, ID: b.dead, Mode: occ})
		case 2:
			out = append(out, game.TargetRef{Kind: game.TargetCard, ID: b.theirs, Mode: occ})
		}
	}
	return out
}

// castRuin casts Inscription of Ruin paying exactly what the
// announcement costs: {2}{B}, plus {2}{B}{B} when kicked.
func castRuin(g *game.Game, b ruinBoard, modes []int, kicked bool) error {
	id := handCardFull(b.me, "Inscription of Ruin", "Sorcery", "{2}{B}", inscriptionOfRuinOracle, []string{"B"})
	mana, optional := "{B}{B}{B}", []int(nil)
	if kicked {
		mana, optional = "{B}{B}{B}{B}{B}{B}{B}", []int{0}
	}
	b.me.ManaPool.EmptyPool()
	if err := g.AddManaForEffect(b.me.ID, uuid.Nil, mana); err != nil {
		return err
	}
	err := g.CastSpell(b.me.ID, id, game.CastSpellParams{
		Strict: true, Modes: modes, OptionalCosts: optional, Targets: b.targets(modes),
	})
	if err != nil {
		_, _ = b.me.Hand.Remove(id)
		b.me.ManaPool.EmptyPool()
	}
	return err
}

// Unkicked: the printed "choose one" — exactly one bullet.
func TestInscriptionOfRuinUnkickedChoosesExactlyOne(t *testing.T) {
	g, b := newRuinBoard(t)
	if err := castRuin(g, b, []int{0, 2}, false); err == nil {
		t.Fatal("unkicked: two bullets must be refused")
	}
	if err := castRuin(g, b, nil, false); err == nil {
		t.Fatal("unkicked: no bullet must be refused")
	}
	if err := castRuin(g, b, []int{2}, false); err != nil {
		t.Fatalf("unkicked: one bullet is legal: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(b.theirs) {
		t.Error("the destroy bullet resolved")
	}
	if g.Battlefield.Contains(b.dead) {
		t.Error("only the chosen bullet resolves: the dead bear stays dead")
	}
}

// Kicked: "choose any number instead" — all three are legal, and all
// three resolve.
func TestInscriptionOfRuinKickedChoosesAnyNumber(t *testing.T) {
	g, b := newRuinBoard(t)
	if err := castRuin(g, b, []int{0, 1, 2}, true); err != nil {
		t.Fatalf("kicked: all three bullets are legal: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !choiceFor(g, b.opp.ID) {
		t.Fatal("the discard bullet asks the opponent to discard")
	}
	// CR 608.2c (#2789): the later bullets wait for the discard.
	if g.Battlefield.Contains(b.dead) {
		t.Error("the reanimation bullet ran before the discard was answered")
	}
	discardFromHand(t, g, b.opp.ID)
	g.SettleResolution()
	if !g.Battlefield.Contains(b.dead) {
		t.Error("the reanimation bullet returns the two-drop")
	}
	if g.Battlefield.Contains(b.theirs) {
		t.Error("the destroy bullet destroys the three-drop")
	}
}

// A kicked three-bullet Inscription on the stack survives a snapshot
// round trip: the restored table resolves all three.
func TestInscriptionOfRuinKickedSurvivesARestore(t *testing.T) {
	g, b := newRuinBoard(t)
	if err := castRuin(g, b, []int{2, 1}, true); err != nil {
		t.Fatalf("kicked, two bullets: %v", err)
	}
	restored := restoreRoundTrip(t, g, false)
	passPriorityAroundTable(t, restored)
	if !restored.Battlefield.Contains(b.dead) || restored.Battlefield.Contains(b.theirs) {
		t.Error("the restored table resolves both chosen bullets")
	}
}

// Kicked, the minimum is still one: "any number" is a ceiling, not a
// floor of zero.
func TestInscriptionOfRuinKickedStillNeedsOneBullet(t *testing.T) {
	g, b := newRuinBoard(t)
	if err := castRuin(g, b, nil, true); err == nil {
		t.Fatal("kicked with no bullet must be refused")
	}
	if err := castRuin(g, b, []int{1}, true); err != nil {
		t.Fatalf("kicked with one bullet is legal: %v", err)
	}
}

// The hand card's view carries the count per announcement: 1..1
// with nothing announced, 1..3 on if_optional_paid, which the picker
// switches to when the caster ticks the kicker.
func TestInscriptionOfRuinViewShowsBothRanges(t *testing.T) {
	g, b := newRuinBoard(t)
	id := handCardFull(b.me, "Inscription of Ruin", "Sorcery", "{2}{B}", inscriptionOfRuinOracle, []string{"B"})
	b.me.Hand.Cards[len(b.me.Hand.Cards)-1].KnownBy = map[uuid.UUID]bool{b.me.ID: true}
	m := handModesFor(t, g, b.me, id)
	if m == nil || m.Min != 1 || m.Max != 1 {
		t.Fatalf("unkicked bounds: %+v, want 1..1", m)
	}
	if m.IfOptionalPaid == nil || m.IfOptionalPaid.Min != 1 || m.IfOptionalPaid.Max != 3 {
		t.Errorf("kicked bounds: %+v, want 1..3", m.IfOptionalPaid)
	}
	// A bystander sees the printed count and no per-seat raise.
	pub := handModesFor(t, g, b.opp, id)
	if pub != nil && pub.IfOptionalPaid != nil {
		t.Errorf("the public copy carries no if_optional_paid: %+v", pub.IfOptionalPaid)
	}
}

// choiceFor reports whether any prompt is waiting on `player`.
func choiceFor(g *game.Game, player uuid.UUID) bool {
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.PendingChoices {
			if c != nil && c.Chooser == player {
				found = true
			}
		}
	})
	return found
}

// handModesFor finds a card in the owner's hand in `viewer`'s frame
// and returns its mode view (nil when the card is hidden or not modal).
func handModesFor(t *testing.T, g *game.Game, viewer *game.Player, id uuid.UUID) *protocol.ModeSpecView {
	t.Helper()
	v := protocol.ViewOfGameFor(g, viewer.ID.String())
	for _, s := range v.Seats {
		for _, c := range s.Hand.Cards {
			if c.InstanceID == id.String() {
				return c.Modes
			}
		}
	}
	return nil
}

// Inscription of Abundance kicked, all three bullets on one pair of
// creatures: printed order (CR 608.2c) puts the counters on first, so
// the life bullet reads the bigger power and the fight deals it.
func TestInscriptionOfAbundanceKickedRunsInPrintedOrder(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Wall", 0, 4)
	id := handCardFull(me, "Inscription of Abundance", "Instant", "{1}{G}", inscriptionOfAbundanceOracle, []string{"G"})
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}{G}{G}{G}{G}"); err != nil {
		t.Fatal(err)
	}
	life := me.Life
	err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Strict: true, Modes: []int{2, 1, 0}, OptionalCosts: []int{0},
		Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: mine, Mode: 0, Slot: 0},
			{Kind: game.TargetCard, ID: theirs, Mode: 0, Slot: 1},
			{Kind: game.TargetPlayer, ID: me.ID, Mode: 1},
			{Kind: game.TargetCard, ID: mine, Mode: 2},
		},
	})
	if err != nil {
		t.Fatalf("kicked, all three: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("life %d, want %d — X is the bear's power AFTER its two counters", me.Life, life+4)
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("the 4-power bear kills the 0/4 wall in the fight")
	}
}

// delirium seeds `types` distinct card types in the player's graveyard.
func delirium(p *game.Player, types ...string) {
	for _, tl := range types {
		p.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: tl + " Card", TypeLine: tl, Owner: p.ID, Controller: p.ID})
	}
}

// titanEnters casts Prophetic Titan (free) and resolves it, leaving
// its ETB trigger's mode_pick prompt open.
func titanEnters(t *testing.T, g *game.Game) *game.PendingChoice {
	t.Helper()
	castCatalogSpell(t, g, "Prophetic Titan", "Creature — Giant Wizard", propheticTitanOracle, nil)
	passPriorityAroundTable(t, g)
	p := pendingOfKind(g, game.PendingChoiceModePick)
	if p == nil {
		t.Fatal("Prophetic Titan's ETB asks for its modes")
	}
	return p
}

// With delirium the Titan must take BOTH bullets: one is refused.
func TestPropheticTitanWithDeliriumForcesBoth(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	delirium(me, "Creature — Bear", "Instant", "Sorcery", "Land")
	for i := 0; i < 5; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	}
	p := titanEnters(t, g)
	if p.ModeMin != 2 || p.ModeMax != 2 {
		t.Fatalf("delirium: prompt bounds %d..%d, want 2..2", p.ModeMin, p.ModeMax)
	}
	if err := g.ResolveModePick(p.ID, me.ID, []int{0}); err == nil {
		t.Fatal("delirium: one bullet must be refused")
	}
	if err := g.ResolveModePick(p.ID, me.ID, []int{0, 1}); err != nil {
		t.Fatalf("delirium: both is the answer: %v", err)
	}
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the damage bullet asks for its target")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
		t.Fatalf("target the opponent: %v", err)
	}
	life := opp.Life
	passPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("opponent at %d, want %d", opp.Life, life-4)
	}
	if p := pendingOfKind(g, game.PendingChoiceModePick); p != nil {
		t.Errorf("no second mode prompt: %+v", p)
	}
}

// Three card types is not delirium: exactly one bullet.
func TestPropheticTitanWithoutDeliriumChoosesOne(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	delirium(me, "Creature — Bear", "Instant", "Sorcery")
	p := titanEnters(t, g)
	if p.ModeMin != 1 || p.ModeMax != 1 {
		t.Fatalf("no delirium: prompt bounds %d..%d, want 1..1", p.ModeMin, p.ModeMax)
	}
	if err := g.ResolveModePick(p.ID, me.ID, []int{0, 1}); err == nil {
		t.Fatal("no delirium: both must be refused")
	}
	if err := g.ResolveModePick(p.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("no delirium: one bullet is the answer: %v", err)
	}
}

// The forced minimum is the prompt's, so it survives a snapshot round
// trip and an undo: a restored table still refuses one bullet.
func TestPropheticTitanForcedCountSurvivesRestoreAndUndo(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	delirium(me, "Creature — Bear", "Instant", "Sorcery", "Land")
	titanEnters(t, g)

	restored := restoreRoundTrip(t, g, false)
	rme := restored.Seats[g.Turn.ActiveSeat]
	p := pendingOfKind(restored, game.PendingChoiceModePick)
	if p == nil || p.ModeMin != 2 || p.ModeMax != 2 {
		t.Fatalf("restored prompt bounds: %+v, want 2..2", p)
	}
	if err := restored.ResolveModePick(p.ID, rme.ID, []int{1}); err == nil {
		t.Error("restored: one bullet must still be refused")
	}

	undo := g.Clone()
	p = pendingOfKind(g, game.PendingChoiceModePick)
	if err := g.ResolveModePick(p.ID, me.ID, []int{0, 1}); err != nil {
		t.Fatalf("both: %v", err)
	}
	g.RestoreFrom(undo)
	p = pendingOfKind(g, game.PendingChoiceModePick)
	if p == nil || p.ModeMin != 2 {
		t.Fatalf("after undo the prompt is back, still forcing two: %+v", p)
	}
	if err := g.ResolveModePick(p.ID, me.ID, []int{0}); err == nil {
		t.Error("after undo: one bullet must still be refused")
	}
}

// castDefiler casts Depth Defiler paying {3}{U}{U}, plus {C} kicked,
// and drains its cast trigger to the mode prompt.
func castDefiler(t *testing.T, g *game.Game, me *game.Player, kicked bool) *game.PendingChoice {
	t.Helper()
	id := handCardFull(me, "Depth Defiler", "Creature — Eldrazi", "{3}{U}{U}", depthDefilerOracle, nil)
	mana, optional := "{U}{U}{U}{U}{U}", []int(nil)
	if kicked {
		mana, optional = "{U}{U}{U}{U}{U}{C}", []int{0}
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: optional}); err != nil {
		t.Fatalf("cast Depth Defiler (kicked=%v): %v", kicked, err)
	}
	passPriorityAroundTable(t, g)
	p := pendingOfKind(g, game.PendingChoiceModePick)
	if p == nil {
		t.Fatal("the cast trigger asks for its modes")
	}
	return p
}

// Depth Defiler's cast trigger reads the SPELL's kicker: kicked, both
// bullets are forced; unkicked, exactly one.
func TestDepthDefilerKickedForcesBoth(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	pushVanillaCreature(g, me.ID, "Bounce Target", 1, 1)
	p := castDefiler(t, g, me, true)
	if p.ModeMin != 2 || p.ModeMax != 2 {
		t.Fatalf("kicked: prompt bounds %d..%d, want 2..2", p.ModeMin, p.ModeMax)
	}
	if err := g.ResolveModePick(p.ID, me.ID, []int{1}); err == nil {
		t.Fatal("kicked: one bullet must be refused")
	}
	if err := g.ResolveModePick(p.ID, me.ID, []int{0, 1}); err != nil {
		t.Fatalf("kicked: both: %v", err)
	}
}

func TestDepthDefilerUnkickedChoosesOne(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	pushVanillaCreature(g, me.ID, "Bounce Target", 1, 1)
	p := castDefiler(t, g, me, false)
	if p.ModeMin != 1 || p.ModeMax != 1 {
		t.Fatalf("unkicked: prompt bounds %d..%d, want 1..1", p.ModeMin, p.ModeMax)
	}
	if err := g.ResolveModePick(p.ID, me.ID, []int{0, 1}); err == nil {
		t.Fatal("unkicked: both must be refused")
	}
}

// Let's Play a Game: "choose one or more instead" with delirium.
func TestLetsPlayAGameDeliriumUnlocksEveryBullet(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	if err := b22TryModal(g, letsPlayAGameOracle, []int{0, 2}, nil); err == nil {
		t.Fatal("no delirium: two bullets must be refused")
	}
	delirium(me, "Creature — Bear", "Instant", "Sorcery", "Land")
	oppLife, myLife := opp.Life, me.Life
	bear := pushVanillaCreature(g, opp.ID, "Their Bear", 1, 1)
	if err := b22TryModal(g, letsPlayAGameOracle, []int{0, 1, 2}, nil); err != nil {
		t.Fatalf("delirium: every bullet: %v", err)
	}
	passPriorityAroundTable(t, g)
	// CR 608.2c (#2789): the life bullet waits for every opponent's
	// discard.
	if opp.Life != oppLife {
		t.Errorf("the life bullet ran before the discards were answered: %d", opp.Life)
	}
	for _, p := range g.Seats {
		if p.ID != me.ID {
			discardFromHand(t, g, p.ID)
		}
	}
	g.SettleResolution()
	if opp.Life != oppLife-3 || me.Life != myLife+3 {
		t.Errorf("life %d/%d, want %d/%d", opp.Life, me.Life, oppLife-3, myLife+3)
	}
	// The -1/-1 bullet made the 1/1 a 0/0, and the state-based check
	// after the resolution finished put it in the graveyard.
	if g.Battlefield.Contains(bear) {
		t.Errorf("the -1/-1 bullet shrinks the opponent's 1/1 to death (power %d)", currentPower(t, g, bear))
	}
}

// Wail of the Forgotten: descend 8 unlocks "one or more".
func TestWailOfTheForgottenDescendUnlocksEveryBullet(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	for i := 0; i < 4; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	}
	opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	rock := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: rock, Name: "Their Rock", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID})
	targets := []game.TargetRef{
		{Kind: game.TargetCard, ID: rock, Mode: 0},
		{Kind: game.TargetPlayer, ID: opp.ID, Mode: 1},
	}
	// Seven permanent cards and a sorcery: not descended 8.
	delirium(me, "Creature — A", "Creature — B", "Land", "Land", "Artifact", "Enchantment", "Creature — C", "Sorcery")
	if err := b22TryModal(g, wailOfTheForgottenOracle, []int{0, 1}, targets); err == nil {
		t.Fatal("seven permanent cards: two bullets must be refused")
	}
	delirium(me, "Land")
	if err := b22TryModal(g, wailOfTheForgottenOracle, []int{0, 1}, targets); err != nil {
		t.Fatalf("eight permanent cards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the bounce bullet returns the rock")
	}
}
