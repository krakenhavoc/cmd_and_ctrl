package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b038_cards_test.go — S82 older sets, Modern Horizons 2 batch 38
// (#5501): delirium, storm and Aura cards. Garth One-Eye is skipped.

const (
	oldB38Marauder  = "25ca4b6c-2c4a-40f1-aee8-3bd373572eb7"
	oldB38Scuttle   = "2be9c1e6-0226-497a-b548-21bf1cfbc286"
	oldB38Zealot    = "35b30274-bf8c-4d3c-b738-f9e62cdd5156"
	oldB38Dockhand  = "3d199876-09f3-4e14-8598-79f0ec2afafa"
	oldB38Drifter   = "c8618fec-e124-4f42-b440-0f41a7de3d1f"
	oldB38Aeve      = "44600a74-9584-488b-a342-2fccd8fa7599"
	oldB38Relay     = "973c684c-a396-4e40-8aaa-365bc64a6f9a"
	oldB38Insurrect = "bc59977f-b3a7-4bc1-9f20-0a1e36d418ae"
	oldB38SoShiny   = "4bc34109-2050-4736-8247-ba06f0a9114a"
	oldB38Ego       = "829436df-a48b-4535-84ec-64682a16949b"
	oldB38Boon      = "a03a9ddb-595a-4dff-b7cb-7ad60655f1ee"
)

func oldB38Delirium(g *game.Game, p *game.Player) {
	effectstest.PushGraveyardCardTyped(p, "G Creature", "Creature — Bear")
	effectstest.PushGraveyardCardTyped(p, "G Instant", "Instant")
	effectstest.PushGraveyardCardTyped(p, "G Land", "Land")
	effectstest.PushGraveyardCardTyped(p, "G Artifact", "Artifact")
}

// oldB38Discard moves a card of the type line from the hand to the
// graveyard through the engine, so the layers see the graveyard change.
func oldB38Discard(t *testing.T, g *game.Game, p *game.Player, typeLine string) {
	t.Helper()
	id := effectstest.B13HandCard(p, "Discarded "+typeLine, typeLine, "")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: p.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: p.ID}, id); err != nil {
		t.Fatal(err)
	}
}

func TestOldB38CardsAreFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB38Marauder: "Bloodbraid Marauder", oldB38Scuttle: "Scuttletide", oldB38Zealot: "Gouged Zealot",
		oldB38Dockhand: "Rishadan Dockhand", oldB38Drifter: "Ghost-Lit Drifter", oldB38Aeve: "Aeve, Progenitor Ooze",
		oldB38Relay: "Galvanic Relay", oldB38Insurrect: "Spreading Insurrection", oldB38SoShiny: "So Shiny",
		oldB38Ego: "Shattered Ego", oldB38Boon: "Skyblade's Boon",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB38MarauderCantBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	marauder := effectstest.B12Push(g, opp.ID, "Bloodbraid Marauder", "Creature — Human Berserker", oldB38Marauder, 3, 1)
	effectstest.BrAttack(t, g, attacker)
	if err := g.DeclareBlocker(marauder, attacker); err == nil {
		t.Error("Bloodbraid Marauder blocked")
	}
}

func TestOldB38MarauderCascadesOnlyWithDelirium(t *testing.T) {
	for _, delirium := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.PushLibraryCardForTest(me, game.Card{Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid", ManaCost: "{G}"})
		if delirium {
			oldB38Delirium(g, me)
		} else {
			effectstest.PushGraveyardCardTyped(me, "G Creature", "Creature — Bear")
		}
		before := me.Library.Size()
		effectstest.AdvanceToMain(t, g)
		effectstest.P1g2Cast(t, g, "Bloodbraid Marauder", "Creature — Human Berserker", oldB38Marauder, "{1}{R}", game.CastSpellParams{})
		effectstest.PassPriorityAroundTable(t, g)
		exiled := before - me.Library.Size()
		if delirium && exiled == 0 {
			t.Error("delirium: cascade exiled nothing")
		}
		if !delirium && exiled != 0 {
			t.Errorf("no delirium: %d cards left the library", exiled)
		}
	}
}

func TestOldB38ScuttletideMakesCrabsAndPumpsThemWithDelirium(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	st := effectstest.SeedPermanentWithOracle(g, me.ID, "Scuttletide", "Enchantment", oldB38Scuttle)
	hand := effectstest.B13HandCard(me, "Discard Me", "Instant", "")
	effectstest.ApaMana(me, "C")
	effectstest.ApaActivate(t, g, me, st, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{hand}})
	if effectstest.B12ZoneOf(g, hand) != game.ZoneGraveyard {
		t.Error("the card was not discarded")
	}
	var crab uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Crab" && c.Controller == me.ID {
			crab = c.InstanceID
		}
	}
	if crab == uuid.Nil {
		t.Fatal("no Crab token")
	}
	if p, tough := effectstest.EffectivePower(t, g, crab), effectstest.EffectiveToughness(t, g, crab); p != 0 || tough != 3 {
		t.Errorf("without delirium the Crab is %d/%d, want 0/3", p, tough)
	}
	for _, tl := range []string{"Creature — Bear", "Land", "Artifact"} {
		oldB38Discard(t, g, me, tl)
	}
	if p, tough := effectstest.EffectivePower(t, g, crab), effectstest.EffectiveToughness(t, g, crab); p != 1 || tough != 4 {
		t.Errorf("with delirium the Crab is %d/%d, want 1/4", p, tough)
	}
	bear := effectstest.SeedBear(g, me.ID)
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("a non-Crab got +1/+1 (power %d)", p)
	}
}

func TestOldB38GougedZealotHitsDefendersCreaturesOnlyWithDelirium(t *testing.T) {
	for _, delirium := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		if delirium {
			oldB38Delirium(g, me)
		}
		zealot := effectstest.B12Push(g, me.ID, "Gouged Zealot", "Creature — Cyclops Berserker", oldB38Zealot, 4, 3)
		mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
		theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
		effectstest.DeclareAttack(t, g, opp.ID, zealot)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if delirium {
			want = 1
		}
		if got := effectstest.DamageOn(t, g, theirs); got != want {
			t.Errorf("delirium=%v: their creature has %d damage, want %d", delirium, got, want)
		}
		if got := effectstest.DamageOn(t, g, mine); got != 0 {
			t.Errorf("delirium=%v: my own creature took %d", delirium, got)
		}
	}
}

func TestOldB38DockhandTapsALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	hand := effectstest.B12Push(g, me.ID, "Rishadan Dockhand", "Creature — Merfolk", oldB38Dockhand, 1, 2)
	land := effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	if !effectstest.HasEffectiveKeyword(t, g, hand, "islandwalk") {
		t.Error("no islandwalk")
	}
	effectstest.ApaMana(me, "C")
	effectstest.ApaActivate(t, g, me, hand, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(land)})
	if !effectstest.IsTapped(g, land) {
		t.Error("the land stayed untapped")
	}
	if !effectstest.IsTapped(g, hand) {
		t.Error("the Dockhand did not tap as a cost")
	}
}

func TestOldB38DrifterGrantsFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	drifter := effectstest.B12Push(g, me.ID, "Ghost-Lit Drifter", "Creature — Spirit", oldB38Drifter, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.ApaMana(me, "U", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, drifter, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(drifter)}); err == nil {
		t.Error("the Drifter targeted itself (\"another\")")
	}
	effectstest.ApaActivate(t, g, me, drifter, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(bear)})
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the Bear did not gain flying")
	}
	effectstest.AdvancePastCleanupOf(t, g, 0)
	if effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("flying outlived the turn")
	}
}

func TestOldB38DrifterChannelGivesXCreaturesFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "C", "Creature — Bear", 2, 2)
	card := effectstest.B13HandCard(me, "Ghost-Lit Drifter", "Creature — Spirit", oldB38Drifter)
	effectstest.ApaMana(me, "U", "C", "C")
	effectstest.ApaActivate(t, g, me, card, 1, game.ActivateAbilityParams{Strict: true, XValue: 2, Targets: effectstest.CardTargets(a, b)})
	for id, want := range map[uuid.UUID]bool{a: true, b: true, c: false} {
		if got := effectstest.HasEffectiveKeyword(t, g, id, "flying"); got != want {
			t.Errorf("flying on %v = %v, want %v", id, got, want)
		}
	}
	if effectstest.B12ZoneOf(g, card) != game.ZoneGraveyard {
		t.Error("the Drifter was not discarded")
	}
}

func TestOldB38AeveEntersWithCountersPerOtherOozeAndStormCopiesAreNotLegendary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Creature(g, me.ID, "Ooze One", "Creature — Ooze", 1, 1)
	effectstest.B12Creature(g, me.ID, "Ooze Two", "Creature — Ooze", 1, 1)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	// Two spells first: storm makes two token copies.
	effectstest.CastCatalogSpell(t, g, "Filler", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Filler", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Aeve, Progenitor Ooze", "Legendary Creature — Ooze", oldB38Aeve, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var cards, tokens int
	for _, c := range g.Battlefield.Cards {
		if c.OracleID != oldB38Aeve && c.Name != "Aeve, Progenitor Ooze" {
			continue
		}
		view := effectstest.S58p6View(t, g, c.InstanceID)
		if c.IsToken() {
			tokens++
			if view.HasSupertype("Legendary") {
				t.Error("a token Aeve is legendary")
			}
		} else {
			cards++
			if !view.HasSupertype("Legendary") {
				t.Error("the card Aeve is not legendary")
			}
		}
	}
	if cards != 1 || tokens != 2 {
		t.Fatalf("%d card(s) and %d token(s) of Aeve on the battlefield, want 1 and 2 (the legend rule must not have fired)", cards, tokens)
	}
}

func TestOldB38AeveCountersCountOtherOozes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Creature(g, me.ID, "Ooze One", "Creature — Ooze", 1, 1)
	effectstest.B12Creature(g, me.ID, "Ooze Two", "Creature — Ooze", 1, 1)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	id := effectstest.CastCatalogSpell(t, g, "Aeve, Progenitor Ooze", "Legendary Creature — Ooze", oldB38Aeve, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != 2 {
		t.Errorf("Aeve has %d counters, want 2 (one per other Ooze)", got)
	}
}

func TestOldB38GalvanicRelayOpensOnlyOnYourNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	top := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Land", TypeLine: "Basic Land — Forest"})
	for i := 0; i < 8; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{Name: "Filler Land", TypeLine: "Basic Land — Forest"})
	}
	effectstest.CastCatalogSpell(t, g, "Galvanic Relay", "Sorcery", oldB38Relay, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, top) != game.ZoneExile {
		t.Fatalf("the top card is in %v, want exile", effectstest.B12ZoneOf(g, top))
	}
	me.LandDropsPerTurn += 2
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{Strict: true, FromZone: "exile"}); err == nil {
		t.Error("the exiled land was playable on the turn it was exiled")
	}
	// Opponents' turns in between: still not playable by me. Then my turn.
	effectstest.AdvanceToMainOf(t, g, 1)
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{Strict: true, FromZone: "exile"}); err == nil {
		t.Error("playable during an opponent's turn")
	}
	effectstest.AdvanceToMainOf(t, g, 0)
	if effectstest.B12ZoneOf(g, top) != game.ZoneExile {
		t.Fatalf("the card left exile by itself: %v", effectstest.B12ZoneOf(g, top))
	}
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Errorf("the exiled land was not playable during my next turn: %v", err)
	}
	if effectstest.B12ZoneOf(g, top) != game.ZoneBattlefield {
		t.Errorf("the land did not enter: it is in %v", effectstest.B12ZoneOf(g, top))
	}
}

func TestOldB38GalvanicRelayStormExilesACardPerCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 8; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{Name: "Filler Land", TypeLine: "Basic Land — Forest"})
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Filler", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	before := g.Exile.Size()
	effectstest.CastCatalogSpell(t, g, "Galvanic Relay", "Sorcery", oldB38Relay, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := g.Exile.Size() - before; got < 2 {
		t.Errorf("%d cards exiled with one prior spell, want 2 (the spell and its copy)", got)
	}
}

func TestOldB38SpreadingInsurrectionThreatensAndStorms(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Ogre", 3, 3)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, theirs).Tapped = true })
	effectstest.CastCatalogSpell(t, g, "Spreading Insurrection", "Sorcery", oldB38Insurrect, effectstest.CardTarget(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, theirs)
	if c.Controller != me.ID {
		t.Fatal("did not gain control")
	}
	if c.Tapped {
		t.Error("the stolen creature stayed tapped")
	}
	if !effectstest.HasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("no haste")
	}
	effectstest.AdvancePastCleanupOf(t, g, 0)
	c, _ = effectstest.BattlefieldCard(g, theirs)
	if c.Controller != opp.ID {
		t.Error("control did not return at end of turn")
	}
}

func TestOldB38SpreadingInsurrectionRefusesYourOwnCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	id := effectstest.B13HandCard(me, "Spreading Insurrection", "Sorcery", oldB38Insurrect)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(mine)}); err == nil {
		t.Error("targeted a creature I control")
	}
}

func TestOldB38SoShinyTapsAndScriesOnlyWithAToken(t *testing.T) {
	for _, token := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Ogre", 3, 3)
		if token {
			effectstest.B12Push(g, me.ID, "Token", "Token Creature — Soldier", "", 1, 1)
		}
		effectstest.AuraCast(t, g, "So Shiny", oldB38SoShiny, victim)
		effectstest.PassPriorityAroundTable(t, g)
		if token {
			effectstest.AnswerScryKeepAll(t, g, me.ID)
		}
		if got := effectstest.IsTapped(g, victim); got != token {
			t.Errorf("token=%v: tapped=%v", token, got)
		}
	}
}

func TestOldB38SoShinyKeepsItTappedThroughUntap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Ogre", 3, 3)
	effectstest.B12Push(g, me.ID, "Token", "Token Creature — Soldier", "", 1, 1)
	effectstest.AuraCast(t, g, "So Shiny", oldB38SoShiny, victim)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.AdvanceToMainOf(t, g, 1)
	if !effectstest.IsTapped(g, victim) {
		t.Error("the enchanted creature untapped")
	}
}

func TestOldB38ShatteredEgoShrinksAndTucks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Ogre", 4, 4)
	for i := 0; i < 5; i++ {
		effectstest.PushLibraryCardForTest(opp, game.Card{Name: "Filler", TypeLine: "Basic Land — Island"})
	}
	ego := effectstest.AuraCast(t, g, "Shattered Ego", oldB38Ego, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, victim); p != 1 {
		t.Errorf("power %d, want 1", p)
	}
	effectstest.ApaMana(me, "U", "U", "C", "C", "C")
	effectstest.ApaActivate(t, g, me, ego, 0, game.ActivateAbilityParams{Strict: true})
	if g.Battlefield.Contains(victim) {
		t.Fatal("the creature is still on the battlefield")
	}
	if names := effectstest.LibraryTopNames(opp, 3); len(names) != 3 || names[2] != "Victim" || names[0] == "Victim" {
		t.Errorf("third from the top, got %v", names)
	}
}

func TestOldB38SkybladesBoonPumpsAndReturnsFromEitherZone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	boon := effectstest.AuraCast(t, g, "Skyblade's Boon", oldB38Boon, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 3 {
		t.Errorf("power %d, want 3", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("no flying")
	}
	effectstest.ApaMana(me, "W", "C", "C")
	effectstest.ApaActivate(t, g, me, boon, 0, game.ActivateAbilityParams{Strict: true})
	if effectstest.B12ZoneOf(g, boon) != game.ZoneHand {
		t.Fatalf("from the battlefield the Boon is in %v, want hand", effectstest.B12ZoneOf(g, boon))
	}
	// And from the graveyard.
	card := game.NewCard("Skyblade's Boon", me.ID)
	card.TypeLine, card.OracleID = "Enchantment — Aura", oldB38Boon
	me.Graveyard.PushTop(card)
	effectstest.ApaMana(me, "W", "C", "C")
	effectstest.ApaActivate(t, g, me, card.InstanceID, 0, game.ActivateAbilityParams{Strict: true})
	if effectstest.B12ZoneOf(g, card.InstanceID) != game.ZoneHand {
		t.Errorf("from the graveyard the Boon is in %v, want hand", effectstest.B12ZoneOf(g, card.InstanceID))
	}
}
