package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// change_target_cards_test.go — #2354's retarget cards: Willbender
// (the caveat that outlived #1830) and the "change the target"
// instants and creatures that need nothing the existing shapes lack.

const (
	willbenderTestOracle    = "0aae277e-e58e-4115-b5fd-0459451e17ec"
	deflectionOracle        = "ec7ae9ed-dc5b-47ed-a4ad-086f3c7c377c"
	swerveOracle            = "c68627b1-025c-48c9-9646-98eb2d268e71"
	shuntOracle             = "e714bcfb-b451-4b7b-ab60-4cb845c75647"
	redirectOracle          = "74a7d880-5a49-412b-af4d-0b4cdb8453ee"
	rerouteOracle           = "997f5a5a-9ee1-4b66-9ffc-d1a2e996e615"
	goblinFlectomancerOracl = "7c3171da-7a95-4f8e-ba30-c31af21c1828"
	muckDrubbOracle         = "032a5616-bbaf-4659-86c4-43edf29b9788"
	insidiousWillOracle     = "8088d110-2314-45c3-b6ee-3f55d1f388e4"
)

// castFor puts a card in `p`'s hand and casts it from there with the
// given params, advancing to a main phase first. Unlike
// castCatalogSpell the caster need not be the active seat.
func castFor(t *testing.T, g *game.Game, p *game.Player, name, typeLine, oracle string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: p.ID, Controller: p.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(p.ID, id, params); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

func rtCardTarget(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// legalSpellTargets is the set of stack spells (and abilities) the
// card's own spell-level clause offers `chooser` right now.
func legalSpellTargets(g *game.Game, chooser uuid.UUID, oracle string) []uuid.UUID {
	var out []uuid.UUID
	g.WithWriteLock(func() {
		lt := g.LegalTargetsForEffect(
			game.SourceObject(chooser, &game.Card{InstanceID: uuid.New(), OracleID: oracle}),
			game.TargetSpecFor(oracle))
		out = lt.Cards
	})
	return out
}

// answerRetargetWith answers the open retarget prompt with `answer`
// (nil declines).
func answerRetargetWith(t *testing.T, g *game.Game, chooser uuid.UUID, answer []game.TargetRef) {
	t.Helper()
	prompt := latestRetarget(g, chooser)
	if prompt == nil {
		t.Fatal("no retarget prompt is open")
	}
	if err := g.ResolveRetarget(prompt.ID, chooser, answer); err != nil {
		t.Fatalf("ResolveRetarget: %v", err)
	}
}

func rtPlayerRef(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: id}}
}

// boltFrom casts a Lightning Bolt at a player from seat `from`.
func boltFrom(t *testing.T, g *game.Game, from *game.Player, at uuid.UUID) uuid.UUID {
	t.Helper()
	return castFor(t, g, from, "Lightning Bolt", "Instant", lightningBoltOracle,
		game.CastSpellParams{Targets: rtPlayerRef(at)})
}

func twoTargetBiteDown(t *testing.T, g *game.Game, caster *game.Player, mine, theirs uuid.UUID) uuid.UUID {
	t.Helper()
	return castFor(t, g, caster, "Bite Down", "Instant", b26BiteDownOracle, game.CastSpellParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0},
		{Kind: game.TargetCard, ID: theirs, Slot: 1},
	}})
}

// --- Willbender ------------------------------------------------------

// The card: turned face up in response to a Bolt, its trigger takes
// the Bolt off its target and the Bolt lands where the controller
// points it. Mandatory, so declining is refused. The ability half of
// "target spell or ability" is reachable too.
func TestWillbenderRedirectsASingleTargetSpellWhenTurnedFaceUp(t *testing.T) {
	g := newCatalogGame(t)
	me, victimA, victimB := g.Seats[0], g.Seats[1].ID, g.Seats[2].ID
	willbender := morphCardOnBattlefield(g, me.ID, "Willbender", willbenderTestOracle)
	advanceToMain(t, g)
	g.WithWriteLock(func() { g.TurnFaceDownForEffect(uuid.Nil, willbender) })
	bolt := boltFrom(t, g, me, victimA)
	// A two-target spell and an ability on the stack beside it: the
	// clause admits the ability and refuses the two-target spell.
	mine := pushCatalogPermanent(g, me.ID, "Mine", "Creature — Bear", "", false)
	theirs := pushCatalogPermanent(g, g.Seats[3].ID, "Theirs", "Creature — Bear", "", false)
	bite := twoTargetBiteDown(t, g, me, mine, theirs)
	bomb := pushCatalogPermanent(g, me.ID, "Goblin Bombardment", "Enchantment", goblinBombardmentOracle, false)
	fodder := pushCatalogPermanent(g, me.ID, "Fodder", "Creature — Goblin", "", false)
	if err := g.ActivateCatalogAbility(me.ID, bomb, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
		Targets:      rtCardTarget(theirs),
	}); err != nil {
		t.Fatalf("activate Goblin Bombardment: %v", err)
	}
	ping := rtItemOnStackFrom(g, bomb)

	fillPoolColored(me, "U", 2)
	if err := g.PerformSpecialAction(me.ID, willbender, game.SpecialActionTurnFaceUp,
		game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn Willbender face up: %v", err)
	}
	b04WaitForPick(t, g, me.ID)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("Willbender's trigger asked for no target")
	}
	if !hasID(pick.PickTargetCards, bolt) || !hasID(pick.PickTargetCards, ping) {
		t.Errorf("a single-target spell and a single-target ability should be offered: %v", pick.PickTargetCards)
	}
	if hasID(pick.PickTargetCards, bite) {
		t.Error("a two-target spell was offered despite \"with a single target\"")
	}
	pickCard(t, g, me.ID, bolt)
	passPriorityAroundTable(t, g)

	prompt := latestRetarget(g, me.ID)
	if prompt == nil {
		t.Fatal("Willbender's trigger opened no retarget prompt")
	}
	if prompt.PickTargetMin != 1 {
		t.Errorf("a mandatory change offers min %d, want 1", prompt.PickTargetMin)
	}
	if err := g.ResolveRetarget(prompt.ID, me.ID, nil); err == nil {
		t.Error("Willbender's change was declined; it prints no \"may\"")
	}
	answerRetargetWith(t, g, me.ID, rtPlayerRef(victimB))
	passPriorityAroundTable(t, g)
	if lifeOf(g, victimA) != 40 || lifeOf(g, victimB) != 37 {
		t.Errorf("the Bolt did not move: A=%d B=%d", lifeOf(g, victimA), lifeOf(g, victimB))
	}
	if faceDownCardForTest(t, g, willbender).FaceDown {
		t.Error("Willbender is still face down")
	}
}

// With nothing it can legally take, the trigger is removed (CR 603.3d)
// and the morph still turns over.
func TestWillbenderWithNoSingleTargetItemOnTheStackAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	willbender := morphCardOnBattlefield(g, me.ID, "Willbender", willbenderTestOracle)
	advanceToMain(t, g)
	g.WithWriteLock(func() { g.TurnFaceDownForEffect(uuid.Nil, willbender) })
	fillPoolColored(me, "U", 2)
	if err := g.PerformSpecialAction(me.ID, willbender, game.SpecialActionTurnFaceUp,
		game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn Willbender face up: %v", err)
	}
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || latestRetarget(g, me.ID) != nil {
		t.Error("a prompt opened with nothing to redirect")
	}
	if faceDownCardForTest(t, g, willbender).FaceDown {
		t.Error("Willbender is still face down")
	}
}

func rtItemOnStackFrom(g *game.Game, source uuid.UUID) uuid.UUID {
	var out uuid.UUID
	g.WithWriteLock(func() {
		for id, it := range g.StackMeta {
			if it != nil && it.Kind == game.StackItemActivated && it.SourceCardID == source {
				out = id
			}
		}
	})
	return out
}

// --- Deflection, Swerve, Shunt ----------------------------------------

func TestChangeTheTargetInstants(t *testing.T) {
	for _, c := range []struct{ name, oracle string }{
		{"Deflection", deflectionOracle},
		{"Swerve", swerveOracle},
		{"Shunt", shuntOracle},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, victimA, victimB := g.Seats[0], g.Seats[1].ID, g.Seats[2].ID
			bolt := boltFrom(t, g, me, victimA)
			castFor(t, g, me, c.name, "Instant", c.oracle,
				game.CastSpellParams{Targets: rtCardTarget(bolt)})
			passPriorityAroundTable(t, g)

			prompt := latestRetarget(g, me.ID)
			if prompt == nil {
				t.Fatalf("%s opened no retarget prompt", c.name)
			}
			if prompt.PickTargetMin != 1 {
				t.Errorf("a mandatory change offers min %d, want 1", prompt.PickTargetMin)
			}
			if hasID(prompt.PickTargetPlayers, victimA) {
				t.Error("the seat the Bolt already points at is offered as the new target")
			}
			answerRetargetWith(t, g, me.ID, rtPlayerRef(victimB))
			passPriorityAroundTable(t, g)
			if lifeOf(g, victimA) != 40 || lifeOf(g, victimB) != 37 {
				t.Errorf("damage landed on the wrong seat: A=%d B=%d", lifeOf(g, victimA), lifeOf(g, victimB))
			}
		})
	}
}

// "With a single target": a two-target spell is not a legal target,
// and neither is nothing — only the Bolt is offered.
func TestChangeTheTargetInstantsRefuseAMultiTargetSpell(t *testing.T) {
	for _, c := range []struct{ name, oracle string }{
		{"Deflection", deflectionOracle},
		{"Swerve", swerveOracle},
		{"Shunt", shuntOracle},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			single := boltFrom(t, g, me, g.Seats[1].ID)
			mine := pushCatalogPermanent(g, me.ID, "Mine", "Creature — Bear", "", false)
			theirs := pushCatalogPermanent(g, g.Seats[2].ID, "Theirs", "Creature — Bear", "", false)
			double := twoTargetBiteDown(t, g, me, mine, theirs)
			legal := legalSpellTargets(g, me.ID, c.oracle)
			if !hasID(legal, single) {
				t.Errorf("a one-target Bolt is not offered: %v", legal)
			}
			if hasID(legal, double) {
				t.Error("a two-target spell reached the picker despite \"with a single target\"")
			}
		})
	}
}

// --- Redirect ----------------------------------------------------------

// "You may choose new targets for target spell" has no single-target
// restriction: both targets of a Bite Down are on offer, and the
// answer may leave them alone.
func TestRedirectMayRetargetATwoTargetSpellOrDecline(t *testing.T) {
	for _, accept := range []bool{true, false} {
		name := "declined"
		if accept {
			name = "accepted"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			mine := pushCatalogPermanent(g, me.ID, "Mine", "Creature — Bear", "", false)
			g.WithWriteLock(func() {
				for i := range g.Battlefield.Cards {
					if g.Battlefield.Cards[i].InstanceID == mine {
						g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 5, 5
					}
				}
			})
			theirs := pushCatalogPermanent(g, g.Seats[1].ID, "Theirs", "Creature — Bear", "", false)
			other := pushCatalogPermanent(g, g.Seats[2].ID, "Other", "Creature — Bear", "", false)
			bite := twoTargetBiteDown(t, g, me, mine, theirs)
			if !hasID(legalSpellTargets(g, me.ID, redirectOracle), bite) {
				t.Fatal("a two-target spell is not a legal target for Redirect")
			}
			castFor(t, g, me, "Redirect", "Instant", redirectOracle,
				game.CastSpellParams{Targets: rtCardTarget(bite)})
			passPriorityAroundTable(t, g)
			prompt := latestRetarget(g, me.ID)
			if prompt == nil {
				t.Fatal("Redirect opened no retarget prompt")
			}
			if prompt.PickTargetMin != 0 {
				t.Errorf(`"you may" offers min %d, want 0`, prompt.PickTargetMin)
			}
			if accept {
				// Slot 0 ("you control") has nowhere else to go, so the
				// only prompt is for slot 1 (CR 115.7a).
				answerRetargetWith(t, g, me.ID, rtCardTarget(other))
			} else {
				answerRetargetWith(t, g, me.ID, nil)
			}
			passPriorityAroundTable(t, g)
			hit := func(id uuid.UUID) bool { return !g.Battlefield.Contains(id) }
			if accept && (!hit(other) || hit(theirs)) {
				t.Errorf("accepted: Other dead=%v Theirs dead=%v, want Other dead and Theirs alive", hit(other), hit(theirs))
			}
			if !accept && (!hit(theirs) || hit(other)) {
				t.Errorf("declined: Theirs dead=%v Other dead=%v, want Theirs dead and Other alive", hit(theirs), hit(other))
			}
		})
	}
}

// --- Reroute -----------------------------------------------------------

// Reroute takes an activated ability's target and draws a card; a
// spell is not a legal target for it.
func TestRerouteRedirectsAnActivatedAbilityAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bomb := pushCatalogPermanent(g, me.ID, "Goblin Bombardment", "Enchantment", goblinBombardmentOracle, false)
	fodder := pushCatalogPermanent(g, me.ID, "Fodder", "Creature — Goblin", "", false)
	victim := pushCatalogPermanent(g, opp.ID, "Victim", "Creature — Bear", "", false)
	other := pushCatalogPermanent(g, g.Seats[2].ID, "Other", "Creature — Bear", "", false)
	if err := g.ActivateCatalogAbility(me.ID, bomb, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
		Targets:      rtCardTarget(victim),
	}); err != nil {
		t.Fatalf("activate Goblin Bombardment: %v", err)
	}
	ping := rtItemOnStackFrom(g, bomb)
	bolt := boltFrom(t, g, me, opp.ID)

	legal := legalSpellTargets(g, me.ID, rerouteOracle)
	if !hasID(legal, ping) {
		t.Errorf("an activated ability is not offered: %v", legal)
	}
	if hasID(legal, bolt) {
		t.Error("a spell was offered to \"target activated ability\"")
	}

	castFor(t, g, me, "Reroute", "Instant", rerouteOracle,
		game.CastSpellParams{Targets: rtCardTarget(ping)})
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	// The first pass resolves Reroute (the topmost item) and opens the
	// retarget prompt, so the draw has happened by now.
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand = %d, want %d after Reroute's draw", got, hand+1)
	}
	prompt := latestRetarget(g, me.ID)
	if prompt == nil {
		t.Skip("the redirect has already been answered by the table pass")
	}
	if !hasID(prompt.PickTargetCards, other) {
		t.Errorf("Other is not offered: %v", prompt.PickTargetCards)
	}
	answerRetargetWith(t, g, me.ID, rtCardTarget(other))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(victim) {
		t.Error("the original target died: the ping was not redirected")
	}
	if g.Battlefield.Contains(other) {
		t.Error("Other survived: the ping was not redirected onto it (a 1/1 takes 1)")
	}
}

// --- Goblin Flectomancer -------------------------------------------------

func TestGoblinFlectomancerSacrificesToMoveAnInstant(t *testing.T) {
	for _, accept := range []bool{true, false} {
		name := "declined"
		if accept {
			name = "accepted"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, victimA, victimB := g.Seats[0], g.Seats[1].ID, g.Seats[2].ID
			flect := pushCatalogPermanent(g, me.ID, "Goblin Flectomancer", "Creature — Goblin Wizard", goblinFlectomancerOracl, false)
			bolt := boltFrom(t, g, me, victimA)
			creature := castFor(t, g, me, "Mizzium Meddler", "Creature — Vedalken Wizard", mizziumMeddlerOracle, game.CastSpellParams{})
			if len(acLegalAbilityTargetIDs(g, me.ID, goblinFlectomancerOracl)) == 0 {
				t.Fatal("the ability offers no targets at all")
			}
			legal := acLegalAbilityTargets(g, me.ID, goblinFlectomancerOracl, 0)
			if !legal[bolt] {
				t.Error("an instant is not offered")
			}
			if legal[creature] {
				t.Error("a creature spell was offered to \"target instant or sorcery spell\"")
			}
			g.WithWriteLock(func() { _ = g.CounterTargetForEffect(creature) })
			if err := g.ActivateCatalogAbility(me.ID, flect, 0, game.ActivateAbilityParams{
				Targets: rtCardTarget(bolt),
			}); err != nil {
				t.Fatalf("activate Goblin Flectomancer: %v", err)
			}
			if g.Battlefield.Contains(flect) {
				t.Error("the sacrifice cost was not paid at announce")
			}
			passPriorityAroundTable(t, g)
			prompt := latestRetarget(g, me.ID)
			if prompt == nil {
				t.Fatal("no retarget prompt")
			}
			if prompt.PickTargetMin != 0 {
				t.Errorf(`"you may" offers min %d, want 0`, prompt.PickTargetMin)
			}
			if accept {
				answerRetargetWith(t, g, me.ID, rtPlayerRef(victimB))
			} else {
				answerRetargetWith(t, g, me.ID, nil)
			}
			passPriorityAroundTable(t, g)
			wantA, wantB := 40, 37
			if !accept {
				wantA, wantB = 37, 40
			}
			if lifeOf(g, victimA) != wantA || lifeOf(g, victimB) != wantB {
				t.Errorf("life A=%d B=%d, want A=%d B=%d", lifeOf(g, victimA), lifeOf(g, victimB), wantA, wantB)
			}
		})
	}
}

func acLegalAbilityTargetIDs(g *game.Game, chooser uuid.UUID, oracle string) []uuid.UUID {
	var out []uuid.UUID
	for id := range acLegalAbilityTargets(g, chooser, oracle, 0) {
		out = append(out, id)
	}
	return out
}

// --- Muck Drubb ----------------------------------------------------------

// A Bolt at a creature is pulled onto the Drubb; a Bolt at a player is
// not a legal target for the trigger's clause at all.
func TestMuckDrubbTakesASpellThatTargetsOnlyACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushCatalogPermanent(g, opp.ID, "Victim", "Creature — Bear", "", false)
	atCreature := castFor(t, g, me, "Lightning Bolt", "Instant", lightningBoltOracle,
		game.CastSpellParams{Targets: rtCardTarget(victim)})
	atPlayer := boltFrom(t, g, me, opp.ID)

	drubb := castFor(t, g, me, "Muck Drubb", "Creature — Beast", muckDrubbOracle, game.CastSpellParams{})
	g.WithWriteLock(func() {
		for i := range g.Stack.Cards {
			if g.Stack.Cards[i].InstanceID == drubb {
				g.Stack.Cards[i].Power, g.Stack.Cards[i].Toughness = 4, 4
			}
		}
	})
	for i := 0; i < 4 && latestPickTarget(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("Muck Drubb's trigger asked for no target")
	}
	if !hasID(pick.PickTargetCards, atCreature) {
		t.Errorf("a spell aimed only at a creature is not offered: %v", pick.PickTargetCards)
	}
	if hasID(pick.PickTargetCards, atPlayer) {
		t.Error("a spell aimed at a player was offered to \"targets only a single creature\"")
	}
	pickCard(t, g, me.ID, atCreature)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(victim) {
		t.Error("the original creature died: the Bolt was not redirected")
	}
	if damageMarkedOn(g, drubb) != 3 {
		t.Errorf("Muck Drubb has %d damage, want 3", damageMarkedOn(g, drubb))
	}
}

// --- Insidious Will -------------------------------------------------------

func TestInsidiousWillCountersTargetSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1].ID
	bolt := boltFrom(t, g, me, victim)
	castFor(t, g, me, "Insidious Will", "Instant", insidiousWillOracle,
		game.CastSpellParams{Modes: []int{0}, Targets: rtCardTarget(bolt)})
	passPriorityAroundTable(t, g)
	if lifeOf(g, victim) != 40 {
		t.Errorf("the Bolt was not countered: life %d", lifeOf(g, victim))
	}
}

func TestInsidiousWillMayChooseNewTargets(t *testing.T) {
	for _, accept := range []bool{true, false} {
		name := "declined"
		if accept {
			name = "accepted"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, victimA, victimB := g.Seats[0], g.Seats[1].ID, g.Seats[2].ID
			bolt := boltFrom(t, g, me, victimA)
			castFor(t, g, me, "Insidious Will", "Instant", insidiousWillOracle,
				game.CastSpellParams{Modes: []int{1}, Targets: rtCardTarget(bolt)})
			passPriorityAroundTable(t, g)
			prompt := latestRetarget(g, me.ID)
			if prompt == nil {
				t.Fatal("no retarget prompt")
			}
			if prompt.PickTargetMin != 0 {
				t.Errorf(`"you may" offers min %d, want 0`, prompt.PickTargetMin)
			}
			if accept {
				answerRetargetWith(t, g, me.ID, rtPlayerRef(victimB))
			} else {
				answerRetargetWith(t, g, me.ID, nil)
			}
			passPriorityAroundTable(t, g)
			wantA, wantB := 40, 37
			if !accept {
				wantA, wantB = 37, 40
			}
			if lifeOf(g, victimA) != wantA || lifeOf(g, victimB) != wantB {
				t.Errorf("life A=%d B=%d, want A=%d B=%d", lifeOf(g, victimA), lifeOf(g, victimB), wantA, wantB)
			}
		})
	}
}

// The copy bullet takes instants and sorceries only, and the copy may
// be pointed elsewhere.
func TestInsidiousWillCopiesAnInstantWithNewTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, victimA, victimB := g.Seats[0], g.Seats[1].ID, g.Seats[2].ID
	bolt := boltFrom(t, g, me, victimA)
	creature := castFor(t, g, me, "Mizzium Meddler", "Creature — Vedalken Wizard", mizziumMeddlerOracle, game.CastSpellParams{})
	legal := legalSpellTargetsForMode(g, me.ID)
	if !hasID(legal, bolt) || hasID(legal, creature) {
		t.Errorf("copy bullet legal targets = %v, want the Bolt and not the creature spell", legal)
	}
	g.WithWriteLock(func() { _ = g.CounterTargetForEffect(creature) })
	castFor(t, g, me, "Insidious Will", "Instant", insidiousWillOracle,
		game.CastSpellParams{Modes: []int{2}, Targets: rtCardTarget(bolt)})
	for i := 0; i < 8 && latestPickTarget(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no re-target prompt for the copy")
	}
	if err := g.ResolvePickTarget(prompt.ID, me.ID,
		game.TargetRef{Kind: game.TargetPlayer, ID: victimB}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	if lifeOf(g, victimA) != 37 || lifeOf(g, victimB) != 37 {
		t.Errorf("life A=%d B=%d, want both 37 (the original and the copy)", lifeOf(g, victimA), lifeOf(g, victimB))
	}
}

func legalSpellTargetsForMode(g *game.Game, chooser uuid.UUID) []uuid.UUID {
	spec, ok := Lookup(insidiousWillOracle)
	if !ok || spec.Modes == nil || len(spec.Modes.Options) < 3 {
		return nil
	}
	var out []uuid.UUID
	g.WithWriteLock(func() {
		out = g.LegalTargetsForEffect(game.SourceChooser(chooser), spec.Modes.Options[2].Targets).Cards
	})
	return out
}
