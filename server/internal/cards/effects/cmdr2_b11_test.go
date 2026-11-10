package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_b11_test.go — wave-2 Commander batch 11, the Faeries (#5368).

const (
	tegwyllsScouringOracle = "58ebaaba-7e35-4588-9679-23b22e39789b"
)

func b11Flyer(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return apaPush(g, owner, owner, game.Card{Name: name, TypeLine: "Creature — Faerie", Power: 1, Toughness: 4, Colors: []string{"U"}, Keywords: []string{"flying"}})
}

func b11Tapped(g *game.Game, id uuid.UUID) (tapped bool) {
	g.WithWriteLock(func() { tapped = findBattlefieldCardForTest(g, id).Tapped })
	return
}

func b11Count(g *game.Game, name string) (n int) {
	g.WithWriteLock(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name {
				n++
			}
		}
	})
	return
}

// Tegwyll's Scouring at instant speed: tapping three untapped flyers opens
// the window (CR 601.3c), the taps are the cost, and the wrath leaves only
// the three new Faerie Rogues.
func TestTegwyllsScouringFlashByTappingThreeFlyers(t *testing.T) {
	g, me, opp := tapTable(t)
	scour := handCardOf(g, me, "Tegwyll's Scouring", "Sorcery", "{4}{B}{B}", tegwyllsScouringOracle)
	a, b, c := b11Flyer(g, me.ID, "Flyer A"), b11Flyer(g, me.ID, "Flyer B"), b11Flyer(g, me.ID, "Flyer C")
	ground := pr7Creature(g, me.ID, "Bear", 2, "G")
	theirs := pr7Creature(g, opp.ID, "Their Bear", 2, "G")
	for g.Turn.Step != game.StepBeginCombat {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	g.WithWriteLock(func() { fillPoolColored(me, "B", 6) })

	// Without announcing the cost a sorcery is closed outside a main phase.
	if err := g.CastSpell(me.ID, scour, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast at instant speed without paying for flash")
	}
	// Two flyers, or a ground creature in the third slot, cannot pay.
	if err := g.CastSpell(me.ID, scour, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}, TeamworkIDs: []uuid.UUID{a, b}}); err == nil {
		t.Fatal("two flyers paid for flash")
	}
	if err := g.CastSpell(me.ID, scour, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}, TeamworkIDs: []uuid.UUID{a, b, ground}}); err == nil {
		t.Fatal("a creature without flying paid for flash")
	}
	if b11Tapped(g, a) || b11Tapped(g, b) {
		t.Fatal("a refused cast tapped a creature")
	}

	if err := g.CastSpell(me.ID, scour, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}, TeamworkIDs: []uuid.UUID{a, b, c}}); err != nil {
		t.Fatalf("flash cast by tapping three flyers: %v", err)
	}
	if !b11Tapped(g, a) || !b11Tapped(g, b) || !b11Tapped(g, c) {
		t.Fatal("the three flyers were not tapped as a cost")
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c, ground, theirs} {
		if g.Battlefield.Contains(id) {
			t.Fatal("a creature survived Tegwyll's Scouring")
		}
	}
	if n := b11Count(g, "Faerie Rogue"); n != 3 {
		t.Fatalf("%d Faerie Rogue tokens, want 3", n)
	}
}

// In a main phase the spell is an ordinary sorcery with no tap cost.
func TestTegwyllsScouringAsASorcery(t *testing.T) {
	g, me, _ := tapTable(t)
	scour := handCardOf(g, me, "Tegwyll's Scouring", "Sorcery", "{4}{B}{B}", tegwyllsScouringOracle)
	bear := pr7Creature(g, me.ID, "Bear", 2, "G")
	g.WithWriteLock(func() { fillPoolColored(me, "B", 6) })
	if err := g.CastSpell(me.ID, scour, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("sorcery cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || b11Count(g, "Faerie Rogue") != 3 {
		t.Fatal("the sweep or the tokens are wrong")
	}
}

const (
	shadowPuppeteersOracle = "0a570fdb-2104-4592-b158-5b5d88f1d087"
	quicklingOracle        = "109899e1-abd7-4582-8b67-b0852755bf75"
	nymrisOracle           = "9ee50bae-47c4-4ae7-a82b-f2a0b366df35"
	stolenByTheFaeOracle   = "5bafd0e0-8bf8-4e06-a54a-81cc6cc4bdb7"
	rankleOracle           = "b8619990-9dc2-4fcc-bc7e-457b77cd2a8e"
	scionOfOonaOracle      = "b34601d2-2aeb-4649-9d73-d0895611bd69"
	cloudOfFaeriesOracle   = "adc1f3b4-cce7-42c9-a208-dbab1d1924f5"
	faerieFormationOracle  = "0366c2f6-6e78-4526-808c-fa7bace6006e"
	faerieBladecrafterOrcl = "164b064e-de50-468a-b874-b054897a5be9"
)

func TestCmdr2B11CardsAreFullWithNoCaveats(t *testing.T) {
	for _, id := range []string{tegwyllsScouringOracle, shadowPuppeteersOracle, quicklingOracle, nymrisOracle,
		stolenByTheFaeOracle, rankleOracle, scionOfOonaOracle, cloudOfFaeriesOracle, faerieFormationOracle, faerieBladecrafterOrcl} {
		spec, ok := Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %v %q, want Full with no caveats", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func b11PutOnBattlefield(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int, kws ...string) uuid.UUID {
	return apaPush(g, owner, owner, game.Card{Name: name, TypeLine: typeLine, OracleID: oracle, Power: power, Toughness: toughness, Keywords: kws})
}

// Shadow Puppeteers: two Faerie Rogues on entering; a flyer that attacks
// may become a red 4/4 Dragon until end of turn, and a ground attacker
// does not trigger it.
func TestShadowPuppeteersTokensAndDragon(t *testing.T) {
	g, me, opp := tapTable(t)
	sp := handCardOf(g, me, "Shadow Puppeteers", "Creature — Faerie Wizard", "{6}{U}", shadowPuppeteersOracle)
	g.WithWriteLock(func() {
		c := &me.Hand.Cards[0]
		c.Power, c.Toughness, c.Keywords, c.Colors = 4, 4, []string{"flying", "ward"}, []string{"U"}
		fillPoolColored(me, "U", 7)
	})
	if err := g.CastSpell(me.ID, sp, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := b11Count(g, "Faerie Rogue"); n != 2 {
		t.Fatalf("%d Faerie Rogue tokens, want 2", n)
	}

	var rogue, ground uuid.UUID
	g.WithWriteLock(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Faerie Rogue" {
				rogue = c.InstanceID
				findBattlefieldCardForTest(g, rogue).SummonedThisTurn = false
			}
		}
	})
	ground = pr7Creature(g, me.ID, "Ground Bear", 2, "G")
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(ground, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker ground: %v", err)
	}
	if err := g.DeclareAttacker(rogue, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker rogue: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	// Exactly one trigger (the flyer's), asked as a "you may".
	ask := b11TriggerPrompt(g, me.ID)
	for i := 0; ask == nil && i < 8; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
		ask = b11TriggerPrompt(g, me.ID)
	}
	if ask == nil {
		t.Fatal("no optional prompt for the flying attacker")
	}
	if err := g.ResolveTriggerPrompt(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	passPriorityAroundTable(t, g)
	ch := effectiveOf(t, g, rogue)
	if ch.Power != 4 || ch.Toughness != 4 {
		t.Errorf("the flyer is %d/%d, want 4/4", ch.Power, ch.Toughness)
	}
	hasDragon, hasRed, hasBlack := false, false, false
	for _, st := range ch.Subtypes {
		hasDragon = hasDragon || st == "Dragon"
	}
	for _, c := range ch.Colors {
		hasRed, hasBlack = hasRed || c == "R", hasBlack || c == "B"
	}
	if !hasDragon || !hasRed || !hasBlack {
		t.Errorf("the flyer is %v %v, want a red Dragon in addition to black Faerie Rogue", ch.Colors, ch.Subtypes)
	}
	if p := effectivePower(t, g, ground); p != 2 {
		t.Errorf("the ground attacker's power is %d, want 2", p)
	}
	if b11TriggerPrompt(g, me.ID) != nil {
		t.Error("a second prompt was asked for the ground attacker")
	}
}

// Declining leaves the attacker alone.
func TestShadowPuppeteersDecline(t *testing.T) {
	g, me, opp := tapTable(t)
	b11PutOnBattlefield(g, me.ID, "Shadow Puppeteers", "Creature — Faerie Wizard", shadowPuppeteersOracle, 4, 4, "flying")
	fl := b11Flyer(g, me.ID, "Flyer")
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(fl, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	ask := b11TriggerPrompt(g, me.ID)
	for i := 0; ask == nil && i < 8; i++ {
		if g.PassPriority() != nil {
			break
		}
		ask = b11TriggerPrompt(g, me.ID)
	}
	if ask == nil {
		t.Fatal("no optional prompt")
	}
	if err := g.ResolveTriggerPrompt(ask.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	passPriorityAroundTable(t, g)
	if ch := effectiveOf(t, g, fl); ch.Power != 1 {
		t.Errorf("declined, but the flyer is %d/%d", ch.Power, ch.Toughness)
	}
}

// Cloud of Faeries untaps up to two lands on entering.
func TestCloudOfFaeriesUntapsTwoLands(t *testing.T) {
	g, me, _ := tapTable(t)
	cloud := handCardOf(g, me, "Cloud of Faeries", "Creature — Faerie", "{1}{U}", cloudOfFaeriesOracle)
	var lands []uuid.UUID
	for i := 0; i < 3; i++ {
		id := b12Permanent(g, me.ID, "Island", "Basic Land — Island")
		g.WithWriteLock(func() { findBattlefieldCardForTest(g, id).Tapped = true })
		lands = append(lands, id)
	}
	g.WithWriteLock(func() { fillPoolColored(me, "U", 2) })
	if err := g.CastSpell(me.ID, cloud, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no untap prompt")
	}
	if pick.ChooseMax != 2 {
		t.Errorf("prompt allows %d lands, want up to 2", pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{lands[0], lands[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if b11Tapped(g, lands[0]) || b11Tapped(g, lands[1]) || !b11Tapped(g, lands[2]) {
		t.Error("exactly the two chosen lands should have untapped")
	}
}

// Faerie Formation makes a Faerie and draws.
func TestFaerieFormationMakesAFaerieAndDraws(t *testing.T) {
	g, me, _ := tapTable(t)
	ff := b11PutOnBattlefield(g, me.ID, "Faerie Formation", "Creature — Faerie", faerieFormationOracle, 5, 4, "flying")
	g.WithWriteLock(func() {
		fillPoolColored(me, "U", 4)
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	})
	handBefore := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, ff, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := b11Count(g, "Faerie"); n != 1 {
		t.Errorf("%d Faerie tokens, want 1", n)
	}
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d → %d, want one card drawn", handBefore, me.Hand.Size())
	}
}

// Scion of Oona: other Faeries you control get +1/+1 and shroud; the Scion
// itself, a non-Faerie and an opponent's Faerie get neither.
func TestScionOfOonaLord(t *testing.T) {
	g, me, opp := tapTable(t)
	scion := b11PutOnBattlefield(g, me.ID, "Scion of Oona", "Creature — Faerie Soldier", scionOfOonaOracle, 1, 1, "flying")
	mine := b11PutOnBattlefield(g, me.ID, "My Faerie", "Creature — Faerie", "", 1, 1)
	bear := b11PutOnBattlefield(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := b11PutOnBattlefield(g, opp.ID, "Their Faerie", "Creature — Faerie", "", 1, 1)
	if p := effectivePower(t, g, mine); p != 2 {
		t.Errorf("my Faerie has power %d, want 2", p)
	}
	for name, id := range map[string]uuid.UUID{"the Scion": scion, "a Bear": bear, "their Faerie": theirs} {
		if p := effectivePower(t, g, id); p != map[string]int{"the Scion": 1, "a Bear": 2, "their Faerie": 1}[name] {
			t.Errorf("%s has power %d", name, p)
		}
		if slices.Contains(effectiveOf(t, g, id).Abilities, "shroud") {
			t.Errorf("%s has shroud", name)
		}
	}
	if !slices.Contains(effectiveOf(t, g, mine).Abilities, "shroud") {
		t.Error("my other Faerie has no shroud")
	}
}

// Faerie Bladecrafter: one counter for any number of Faeries that connect,
// and a drain for its power on dying.
func TestFaerieBladecrafterCounterAndDrain(t *testing.T) {
	g, me, opp := tapTable(t)
	bc := b11PutOnBattlefield(g, me.ID, "Faerie Bladecrafter", "Creature — Faerie Rogue", faerieBladecrafterOrcl, 2, 2, "flying")
	other := b11PutOnBattlefield(g, me.ID, "Faerie Two", "Creature — Faerie", "", 1, 1, "flying")
	attackWith(t, g, opp.ID, bc, other)
	passPriorityAroundTable(t, g)
	if p := effectivePowerWithCounters(t, g, bc); p != 3 {
		t.Errorf("Bladecrafter has power %d after two Faeries connected, want 3 (one counter)", p)
	}
	meLife, oppLife := me.Life, opp.Life
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(bc) })
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-3 || me.Life != meLife+3 {
		t.Errorf("after dying life is me %d→%d, opp %d→%d, want +3/-3", meLife, me.Life, oppLife, opp.Life)
	}
}

// Nymris looks at two on the first spell during an opponent's turn only.
func TestNymrisFirstSpellOnAnOpponentsTurn(t *testing.T) {
	g, me, opp := tapTable(t)
	b11PutOnBattlefield(g, me.ID, "Nymris", "Legendary Creature — Faerie Knight", nymrisOracle, 1, 6, "flying", "flash")
	g.WithWriteLock(func() {
		for i := 0; i < 3; i++ {
			me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Lib", TypeLine: "Land", Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
		}
	})
	// On my own turn, no trigger.
	bolt := handCardOf(g, me, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
	g.WithWriteLock(func() { fillPoolColored(me, "R", 1) })
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if triggerOnStack(g, uuid.Nil) != nil || len(g.PendingChoices) != 0 {
		t.Fatal("Nymris triggered on my own turn")
	}
}

func b11TriggerPrompt(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		if c := g.PendingChoices[i]; c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// Rankle: choose the life-and-draw bullet and the sacrifice bullet; every
// player loses 1 life, draws a card and (with a creature) sacrifices it.
func TestRankleChoosesAnyNumberOfBullets(t *testing.T) {
	g, me, opp := tapTable(t)
	rankle := b11PutOnBattlefield(g, me.ID, "Rankle", "Legendary Creature — Faerie Rogue", rankleOracle, 3, 3, "flying", "haste")
	mine := pr7Creature(g, me.ID, "My Bear", 2, "G")
	theirs := pr7Creature(g, opp.ID, "Their Bear", 2, "G")
	g.WithWriteLock(func() {
		for _, p := range []*game.Player{me, opp} {
			p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Land", Owner: p.ID, Controller: p.ID})
		}
	})
	myHand, oppHand, myLife, oppLife := me.Hand.Size(), opp.Hand.Size(), me.Life, opp.Life
	dealCombatDamageToPlayer(g, rankle, opp.ID, 3)
	mode := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	for i := 0; mode == nil && i < 6; i++ {
		passPriorityAroundTable(t, g)
		mode = latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	}
	if mode == nil {
		t.Fatal("no mode pick")
	}
	if err := g.ResolveModePick(mode.ID, me.ID, []int{1, 2}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	for i := 0; i < 12; i++ {
		passPriorityAroundTable(t, g)
		if p := latestChooseCardsFor(g, me.ID); p != nil {
			_ = g.ResolveChooseCards(p.ID, me.ID, p.ChooseCards[:1])
		}
		if p := latestChooseCardsFor(g, opp.ID); p != nil {
			_ = g.ResolveChooseCards(p.ID, opp.ID, p.ChooseCards[:1])
		}
	}
	if me.Life != myLife-1 || opp.Life != oppLife-1 {
		t.Errorf("life me %d→%d opp %d→%d, want -1 each", myLife, me.Life, oppLife, opp.Life)
	}
	if me.Hand.Size() != myHand+1 || opp.Hand.Size() != oppHand+1 {
		t.Errorf("hands me %d→%d opp %d→%d, want +1 each", myHand, me.Hand.Size(), oppHand, opp.Hand.Size())
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("the opponent kept their only creature")
	}
	_ = mine
}
