package testst

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// seedThemeDeck puts the eight Plains onto the battlefield and the
// four spells into the theme seat's hand, all through
// deck.ToGameCard. Returns the hand instance IDs by card name.
func seedThemeDeck(t *testing.T, g *game.Game, p *game.Player) map[string]uuid.UUID {
	t.Helper()

	plains := effectstest.ThemeDeckRow("Plains", "", "Basic Land — Plains", "")
	plains.ProducedMana = []string{"W"}
	for i := 0; i < 8; i++ {
		land := effectstest.ImportThemeCard(plains, p.ID)
		land.InstanceID = uuid.New()
		effectstest.PushBattlefieldCardWithTimestamp(g, land)
	}

	benalia := effectstest.ThemeDeckRow("History of Benalia", effectstest.HistoryOfBenaliaOracle,
		"Enchantment — Saga", "{1}{W}{W}")

	copter := effectstest.ThemeDeckRow("Smuggler's Copter", effectstest.SmugglersCopterOracle,
		"Artifact — Vehicle", "{2}")
	copter.Power, copter.Toughness = "3", "3"
	copter.Keywords = []string{"Flying"}

	elspeth := effectstest.ThemeDeckRow("Elspeth, Sun's Champion", effectstest.ElspethSunsChampionOracle,
		"Legendary Planeswalker — Elspeth", "{4}{W}{W}")
	elspeth.Loyalty = "4"

	emperor := effectstest.ThemeDeckRow("The Wandering Emperor", effectstest.WanderingEmperorOracle,
		"Legendary Planeswalker — Wanderer", "{1}{W}{W}")
	emperor.Loyalty = "3"
	emperor.Keywords = []string{"Flash"}

	ids := make(map[string]uuid.UUID, 4)
	for _, row := range []cards.Card{benalia, copter, elspeth, emperor} {
		c := effectstest.ImportThemeCard(row, p.ID)
		ids[c.Name] = c.InstanceID
		p.Hand.PushTop(c)
	}
	return ids
}

// castThemeSpellTargeting is castThemeSpell for the spells that
// choose their targets at announce (CR 601.2c). Same strict gate,
// same auto-tapper, same settle: split out rather than made variadic
// on castThemeSpell so the untargeted call sites keep reading as one
// line. Shared by the S30 theme deck (#678) and the S24 one (#726).
func castThemeSpellTargeting(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID, targets ...game.TargetRef) {
	t.Helper()
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{
		Targets: targets,
		Strict:  true,
		AutoTap: true,
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// knightIDs returns the Knight tokens History of Benalia has made so
// far, in battlefield order. The test only needs "a Knight" and "a
// second, different Knight", so the order is not load-bearing.
func knightIDs(g *game.Game) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Knight" {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// TestThemeDeckPlaysFourTurns is S27's exit criterion. See the file
// comment for what each turn proves and why the harness is shaped
// the way it is.
func TestThemeDeckPlaysFourTurns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	themeSeat := g.Turn.ActiveSeat
	me := g.Seats[themeSeat]
	opponent := g.Seats[(themeSeat+1)%len(g.Seats)]

	hand := seedThemeDeck(t, g, me)

	// --- turn 1 ------------------------------------------------------
	//
	// The Vehicle lands as a bare artifact and the Saga lands with its
	// first lore counter already on it (CR 714.3), which fires
	// chapter I before anybody gets priority back.
	effectstest.AdvanceToMainOf(t, g, themeSeat)

	effectstest.CastThemeSpell(t, g, me, hand["Smuggler's Copter"])
	copter := hand["Smuggler's Copter"]
	if !effects.OnBattlefield(g, copter) {
		t.Fatal("Smuggler's Copter never reached the battlefield")
	}
	if effectstest.IsCreatureNow(g, copter) {
		t.Error("an uncrewed Vehicle is a creature (CR 301.7 — it is not, until a crew ability resolves)")
	}
	// Flying rode in on the Scryfall record's `keywords` array rather
	// than the catalog entry: #317 / #319 / #320 put printed keywords
	// on the import road precisely so a card the catalog has never
	// heard of still flies.
	if c, _ := effectstest.BattlefieldCardByID(g, copter); !game.HasKeyword(&c, "flying") {
		t.Error("the Copter's printed flying did not survive deck.ToGameCard")
	}

	effectstest.CastThemeSpell(t, g, me, hand["History of Benalia"])
	saga := hand["History of Benalia"]
	if got := effectstest.CounterOn(g, saga, game.CounterLore); got != 1 {
		t.Fatalf("lore counters on the turn it entered = %d, want 1 (CR 714.3)", got)
	}
	if got := len(knightIDs(g)); got != 1 {
		t.Fatalf("Knights after chapter I = %d, want 1", got)
	}

	// A creature that arrived this turn may still crew: tapping to
	// crew is not paying a {T} cost (CR 702.122b). This is the rule
	// implementations most often get wrong in the strict direction.
	firstKnight := knightIDs(g)[0]
	if err := g.ActivateCatalogAbility(me.ID, copter, 0, game.ActivateAbilityParams{
		CrewIDs: []uuid.UUID{firstKnight},
	}); err != nil {
		t.Fatalf("crew with a summoning-sick Knight (CR 702.122b): %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.IsCreatureNow(g, copter) {
		t.Error("the Copter did not become a creature when its crew ability resolved")
	}

	// --- turn 2 ------------------------------------------------------
	//
	// Cleanup took the animation with it; the saga ticks to II on the
	// way through precombat main; and the Copter — on the battlefield
	// since last turn — can attack the moment it is crewed again.
	effectstest.AdvanceToStepOf(t, g, themeSeat, game.StepUpkeep)
	if effectstest.IsCreatureNow(g, copter) {
		t.Error("the crew animation survived cleanup (ADR 0035: until-EOT effects expire at CR 514.2)")
	}

	effectstest.AdvanceToMainOf(t, g, themeSeat)
	// The chapter ability goes on the stack as the lore counter
	// lands (CR 714.2b); it needs a trip round the table to resolve.
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, saga, game.CounterLore); got != 2 {
		t.Fatalf("lore counters after the second precombat main = %d, want 2 (CR 714.3)", got)
	}
	if got := len(knightIDs(g)); got != 2 {
		t.Fatalf("Knights after chapter II = %d, want 2", got)
	}

	secondKnight := knightIDs(g)[1]
	if err := g.ActivateCatalogAbility(me.ID, copter, 0, game.ActivateAbilityParams{
		CrewIDs: []uuid.UUID{secondKnight},
	}); err != nil {
		t.Fatalf("crew: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AdvanceToStepOf(t, g, themeSeat, game.StepDeclareAttackers)
	beforeLife := opponent.Life
	beforeHand := len(me.Hand.Cards)
	beforeYard := len(me.Graveyard.Cards)
	if err := g.DeclareAttacker(copter, opponent.ID); err != nil {
		t.Fatalf("a Vehicle crewed the turn after it entered could not attack (CR 302.6): %v", err)
	}
	// #859: the declaration is announced at its lock-in — the
	// priority wrap inside declare_attackers — so the attack triggers
	// exist only after this.
	effectstest.LockInAttacks(t, g)

	// "Whenever this Vehicle attacks ... you may draw a card. If you
	// do, discard a card." The draw is immediate; the discard is a
	// prompt the controller answers, and since #651 the table waits
	// for it — so what the trigger leaves behind is one extra card and
	// one card owed, and combat cannot walk on until it is paid.
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != beforeHand+1 {
		t.Errorf("hand size after the Copter's loot = %d, want %d (the draw half)", got, beforeHand+1)
	}
	if got := effectstest.DiscardOwed(g, me.ID); got != 1 {
		t.Errorf("discards owed after the Copter's loot = %d, want 1", got)
	}
	if got := len(me.Graveyard.Cards); got != beforeYard {
		t.Errorf("graveyard grew to %d before the owed discard was made", got)
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	if got := len(me.Graveyard.Cards); got != beforeYard+1 {
		t.Errorf("graveyard = %d after the discard, want %d", got, beforeYard+1)
	}

	effectstest.AdvanceToStepOf(t, g, themeSeat, game.StepCombatDamage)
	if loss := beforeLife - opponent.Life; loss != 3 {
		t.Errorf("Copter combat damage = %d, want 3", loss)
	}

	// --- turn 3 ------------------------------------------------------
	//
	// Chapter III pumps the Knights and, once it has RESOLVED, the
	// CR 704.5s state-based action sacrifices the Saga. Elspeth then
	// enters with the loyalty her Scryfall record printed.
	effectstest.AdvanceToMainOf(t, g, themeSeat)
	effectstest.PassPriorityAroundTable(t, g)
	// Chapter III both fired and finished: the pump is the evidence
	// it resolved, and the empty battlefield slot is the evidence
	// CR 704.5s waited for that resolution before sacrificing.
	if got := effectstest.EffectivePower(t, g, firstKnight); got != 4 {
		t.Errorf("Knight power after chapter III = %d, want 4 (2 printed + 2)", got)
	}
	if effects.OnBattlefield(g, saga) {
		t.Error("a Saga at its final chapter with nothing on the stack survived (CR 704.5s)")
	}
	if !effectstest.InGraveyardOf(g, me.ID, saga) {
		t.Error("the sacrificed Saga is not in its owner's graveyard")
	}

	effectstest.CastThemeSpell(t, g, me, hand["Elspeth, Sun's Champion"])
	elspeth := hand["Elspeth, Sun's Champion"]
	if got := effectstest.CounterOn(g, elspeth, game.CounterLoyalty); got != 4 {
		t.Fatalf("Elspeth's loyalty counters on entry = %d, want 4 — printed loyalty did not "+
			"survive deck.ToGameCard (#274)", got)
	}

	if err := g.ActivateCatalogAbility(me.ID, elspeth, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Elspeth +1: %v", err)
	}
	if got := effectstest.CounterOn(g, elspeth, game.CounterLoyalty); got != 5 {
		t.Errorf("loyalty after +1 = %d, want 5 — the cost is paid at announce (CR 606.4)", got)
	}
	// CR 307.1 / 606.3: with the +1 still on the stack the stack is
	// not empty, so a second loyalty activation is outside its window
	// (#1352) before the once-per-turn clause is even asked.
	err := g.ActivateCatalogAbility(me.ID, elspeth, 0, game.ActivateAbilityParams{})
	if !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Errorf("loyalty activation over the +1 on the stack returned %v, want ErrSorcerySpeedRequired", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// CR 606.3: one loyalty ability per planeswalker per turn — asked
	// once the stack is empty again.
	err = g.ActivateCatalogAbility(me.ID, elspeth, 0, game.ActivateAbilityParams{})
	if !errors.Is(err, game.ErrLoyaltyAlreadyActivated) {
		t.Errorf("second loyalty activation in one turn returned %v, want ErrLoyaltyAlreadyActivated", err)
	}
	if got := effectstest.CountNamed(g, "Soldier"); got != 3 {
		t.Errorf("Soldiers after Elspeth's +1 = %d, want 3", got)
	}

	// --- between turns: an opponent attacks the planeswalker ---------
	//
	// CR 508.1d — a creature may be declared attacking a planeswalker
	// its controller's opponent controls, and the damage comes off as
	// loyalty counters (CR 120.3c). This is the half of #406 that made
	// planeswalkers unkillable by damage.
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Raging Bear",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      opponent.ID,
		Controller: opponent.ID,
	})
	oppSeat := (themeSeat + 1) % len(g.Seats)
	effectstest.AdvanceToStepOf(t, g, oppSeat, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, elspeth); err != nil {
		t.Fatalf("declare an attacker against a planeswalker (CR 508.1d): %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToStepOf(t, g, oppSeat, game.StepCombatDamage)
	if got := effectstest.CounterOn(g, elspeth, game.CounterLoyalty); got != 3 {
		t.Fatalf("Elspeth's loyalty after taking 2 combat damage = %d, want 3 (#406)", got)
	}

	// --- turn 4 ------------------------------------------------------
	//
	// The chapter-III pump is long gone, the loyalty gate has
	// reopened, and Elspeth's −3 takes her to zero: the sweep resolves
	// and then CR 704.5i puts her in the graveyard.
	effectstest.AdvanceToMainOf(t, g, themeSeat)
	if got := effectstest.EffectivePower(t, g, firstKnight); got != 2 {
		t.Errorf("Knight power a turn after chapter III = %d, want 2 (the pump was until end of turn)", got)
	}

	effectstest.CastThemeSpell(t, g, me, hand["The Wandering Emperor"])
	emperor := hand["The Wandering Emperor"]
	if got := effectstest.CounterOn(g, emperor, game.CounterLoyalty); got != 3 {
		t.Errorf("The Wandering Emperor's loyalty on entry = %d, want 3", got)
	}

	ogre := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Hulking Ogre",
		TypeLine:   "Creature — Ogre",
		Power:      5,
		Toughness:  5,
		Owner:      opponent.ID,
		Controller: opponent.ID,
	})

	if err := g.ActivateCatalogAbility(me.ID, elspeth, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Elspeth −3 on a new turn (the CR 606.3 gate should have reopened): %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if effects.OnBattlefield(g, ogre) {
		t.Error("a 5/5 survived 'destroy all creatures with power 4 or greater'")
	}
	if !effects.OnBattlefield(g, firstKnight) {
		t.Error("a 2/2 Knight was swept by a 'power 4 or greater' sweeper")
	}
	if effectstest.CountNamed(g, "Soldier") != 3 {
		t.Error("the 1/1 Soldiers were swept by a 'power 4 or greater' sweeper")
	}
	if effects.OnBattlefield(g, elspeth) {
		t.Error("a planeswalker at zero loyalty survived (CR 704.5i)")
	}
	if !effectstest.InGraveyardOf(g, me.ID, elspeth) {
		t.Error("the dead planeswalker is not in its owner's graveyard")
	}
}
