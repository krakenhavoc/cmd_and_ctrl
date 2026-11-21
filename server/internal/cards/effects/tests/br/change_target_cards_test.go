package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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

// --- Willbender ------------------------------------------------------

// The card: turned face up in response to a Bolt, its trigger takes
// the Bolt off its target and the Bolt lands where the controller
// points it. Mandatory, so declining is refused. The ability half of
// "target spell or ability" is reachable too.
func TestWillbenderRedirectsASingleTargetSpellWhenTurnedFaceUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victimA, victimB := g.Seats[0], g.Seats[1].ID, g.Seats[2].ID
	willbender := effectstest.MorphCardOnBattlefield(g, me.ID, "Willbender", willbenderTestOracle)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { g.TurnFaceDownForEffect(uuid.Nil, willbender) })
	bolt := effectstest.BoltFrom(t, g, me, victimA)
	// A two-target spell and an ability on the stack beside it: the
	// clause admits the ability and refuses the two-target spell.
	mine := effectstest.PushCatalogPermanent(g, me.ID, "Mine", "Creature — Bear", "", false)
	theirs := effectstest.PushCatalogPermanent(g, g.Seats[3].ID, "Theirs", "Creature — Bear", "", false)
	bite := effectstest.TwoTargetBiteDown(t, g, me, mine, theirs)
	bomb := effectstest.PushCatalogPermanent(g, me.ID, "Goblin Bombardment", "Enchantment", effectstest.GoblinBombardmentOracle, false)
	fodder := effectstest.PushCatalogPermanent(g, me.ID, "Fodder", "Creature — Goblin", "", false)
	if err := g.ActivateCatalogAbility(me.ID, bomb, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
		Targets:      effectstest.RtCardTarget(theirs),
	}); err != nil {
		t.Fatalf("activate Goblin Bombardment: %v", err)
	}
	ping := rtItemOnStackFrom(g, bomb)

	effectstest.FillPoolColored(me, "U", 2)
	if err := g.PerformSpecialAction(me.ID, willbender, game.SpecialActionTurnFaceUp,
		game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn Willbender face up: %v", err)
	}
	effectstest.B04WaitForPick(t, g, me.ID)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("Willbender's trigger asked for no target")
	}
	if !effectstest.HasID(pick.PickTargetCards, bolt) || !effectstest.HasID(pick.PickTargetCards, ping) {
		t.Errorf("a single-target spell and a single-target ability should be offered: %v", pick.PickTargetCards)
	}
	if effectstest.HasID(pick.PickTargetCards, bite) {
		t.Error("a two-target spell was offered despite \"with a single target\"")
	}
	effectstest.PickCard(t, g, me.ID, bolt)
	effectstest.PassPriorityAroundTable(t, g)

	prompt := effectstest.LatestRetarget(g, me.ID)
	if prompt == nil {
		t.Fatal("Willbender's trigger opened no retarget prompt")
	}
	if prompt.PickTargetMin != 1 {
		t.Errorf("a mandatory change offers min %d, want 1", prompt.PickTargetMin)
	}
	if err := g.ResolveRetarget(prompt.ID, me.ID, nil); err == nil {
		t.Error("Willbender's change was declined; it prints no \"may\"")
	}
	effectstest.AnswerRetargetWith(t, g, me.ID, effectstest.RtPlayerRef(victimB))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LifeOf(g, victimA) != 40 || effectstest.LifeOf(g, victimB) != 37 {
		t.Errorf("the Bolt did not move: A=%d B=%d", effectstest.LifeOf(g, victimA), effectstest.LifeOf(g, victimB))
	}
	if effectstest.FaceDownCardForTest(t, g, willbender).FaceDown {
		t.Error("Willbender is still face down")
	}
}

// With nothing it can legally take, the trigger is removed (CR 603.3d)
// and the morph still turns over.
func TestWillbenderWithNoSingleTargetItemOnTheStackAsksNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	willbender := effectstest.MorphCardOnBattlefield(g, me.ID, "Willbender", willbenderTestOracle)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { g.TurnFaceDownForEffect(uuid.Nil, willbender) })
	effectstest.FillPoolColored(me, "U", 2)
	if err := g.PerformSpecialAction(me.ID, willbender, game.SpecialActionTurnFaceUp,
		game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn Willbender face up: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.LatestRetarget(g, me.ID) != nil {
		t.Error("a prompt opened with nothing to redirect")
	}
	if effectstest.FaceDownCardForTest(t, g, willbender).FaceDown {
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
			g := effectstest.NewCatalogGame(t)
			me, victimA, victimB := g.Seats[0], g.Seats[1].ID, g.Seats[2].ID
			bolt := effectstest.BoltFrom(t, g, me, victimA)
			effectstest.CastFor(t, g, me, c.name, "Instant", c.oracle,
				game.CastSpellParams{Targets: effectstest.RtCardTarget(bolt)})
			effectstest.PassPriorityAroundTable(t, g)

			prompt := effectstest.LatestRetarget(g, me.ID)
			if prompt == nil {
				t.Fatalf("%s opened no retarget prompt", c.name)
			}
			if prompt.PickTargetMin != 1 {
				t.Errorf("a mandatory change offers min %d, want 1", prompt.PickTargetMin)
			}
			if effectstest.HasID(prompt.PickTargetPlayers, victimA) {
				t.Error("the seat the Bolt already points at is offered as the new target")
			}
			effectstest.AnswerRetargetWith(t, g, me.ID, effectstest.RtPlayerRef(victimB))
			effectstest.PassPriorityAroundTable(t, g)
			if effectstest.LifeOf(g, victimA) != 40 || effectstest.LifeOf(g, victimB) != 37 {
				t.Errorf("damage landed on the wrong seat: A=%d B=%d", effectstest.LifeOf(g, victimA), effectstest.LifeOf(g, victimB))
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
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			single := effectstest.BoltFrom(t, g, me, g.Seats[1].ID)
			mine := effectstest.PushCatalogPermanent(g, me.ID, "Mine", "Creature — Bear", "", false)
			theirs := effectstest.PushCatalogPermanent(g, g.Seats[2].ID, "Theirs", "Creature — Bear", "", false)
			double := effectstest.TwoTargetBiteDown(t, g, me, mine, theirs)
			legal := legalSpellTargets(g, me.ID, c.oracle)
			if !effectstest.HasID(legal, single) {
				t.Errorf("a one-target Bolt is not offered: %v", legal)
			}
			if effectstest.HasID(legal, double) {
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
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			mine := effectstest.PushCatalogPermanent(g, me.ID, "Mine", "Creature — Bear", "", false)
			g.WithWriteLock(func() {
				for i := range g.Battlefield.Cards {
					if g.Battlefield.Cards[i].InstanceID == mine {
						g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 5, 5
					}
				}
			})
			theirs := effectstest.PushCatalogPermanent(g, g.Seats[1].ID, "Theirs", "Creature — Bear", "", false)
			other := effectstest.PushCatalogPermanent(g, g.Seats[2].ID, "Other", "Creature — Bear", "", false)
			bite := effectstest.TwoTargetBiteDown(t, g, me, mine, theirs)
			if !effectstest.HasID(legalSpellTargets(g, me.ID, redirectOracle), bite) {
				t.Fatal("a two-target spell is not a legal target for Redirect")
			}
			effectstest.CastFor(t, g, me, "Redirect", "Instant", redirectOracle,
				game.CastSpellParams{Targets: effectstest.RtCardTarget(bite)})
			effectstest.PassPriorityAroundTable(t, g)
			prompt := effectstest.LatestRetarget(g, me.ID)
			if prompt == nil {
				t.Fatal("Redirect opened no retarget prompt")
			}
			if prompt.PickTargetMin != 0 {
				t.Errorf(`"you may" offers min %d, want 0`, prompt.PickTargetMin)
			}
			if accept {
				// Slot 0 ("you control") has nowhere else to go, so the
				// only prompt is for slot 1 (CR 115.7a).
				effectstest.AnswerRetargetWith(t, g, me.ID, effectstest.RtCardTarget(other))
			} else {
				effectstest.AnswerRetargetWith(t, g, me.ID, nil)
			}
			effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bomb := effectstest.PushCatalogPermanent(g, me.ID, "Goblin Bombardment", "Enchantment", effectstest.GoblinBombardmentOracle, false)
	fodder := effectstest.PushCatalogPermanent(g, me.ID, "Fodder", "Creature — Goblin", "", false)
	victim := effectstest.PushCatalogPermanent(g, opp.ID, "Victim", "Creature — Bear", "", false)
	other := effectstest.PushCatalogPermanent(g, g.Seats[2].ID, "Other", "Creature — Bear", "", false)
	if err := g.ActivateCatalogAbility(me.ID, bomb, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
		Targets:      effectstest.RtCardTarget(victim),
	}); err != nil {
		t.Fatalf("activate Goblin Bombardment: %v", err)
	}
	ping := rtItemOnStackFrom(g, bomb)
	bolt := effectstest.BoltFrom(t, g, me, opp.ID)

	legal := legalSpellTargets(g, me.ID, rerouteOracle)
	if !effectstest.HasID(legal, ping) {
		t.Errorf("an activated ability is not offered: %v", legal)
	}
	if effectstest.HasID(legal, bolt) {
		t.Error("a spell was offered to \"target activated ability\"")
	}

	effectstest.CastFor(t, g, me, "Reroute", "Instant", rerouteOracle,
		game.CastSpellParams{Targets: effectstest.RtCardTarget(ping)})
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	// The first pass resolves Reroute (the topmost item) and opens the
	// retarget prompt, so the draw has happened by now.
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand = %d, want %d after Reroute's draw", got, hand+1)
	}
	prompt := effectstest.LatestRetarget(g, me.ID)
	if prompt == nil {
		t.Skip("the redirect has already been answered by the table pass")
	}
	if !effectstest.HasID(prompt.PickTargetCards, other) {
		t.Errorf("Other is not offered: %v", prompt.PickTargetCards)
	}
	effectstest.AnswerRetargetWith(t, g, me.ID, effectstest.RtCardTarget(other))
	effectstest.PassPriorityAroundTable(t, g)
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
			g := effectstest.NewCatalogGame(t)
			me, victimA, victimB := g.Seats[0], g.Seats[1].ID, g.Seats[2].ID
			flect := effectstest.PushCatalogPermanent(g, me.ID, "Goblin Flectomancer", "Creature — Goblin Wizard", goblinFlectomancerOracl, false)
			bolt := effectstest.BoltFrom(t, g, me, victimA)
			creature := effectstest.CastFor(t, g, me, "Mizzium Meddler", "Creature — Vedalken Wizard", effectstest.MizziumMeddlerOracle, game.CastSpellParams{})
			if len(acLegalAbilityTargetIDs(g, me.ID, goblinFlectomancerOracl)) == 0 {
				t.Fatal("the ability offers no targets at all")
			}
			legal := effectstest.AcLegalAbilityTargets(g, me.ID, goblinFlectomancerOracl, 0)
			if !legal[bolt] {
				t.Error("an instant is not offered")
			}
			if legal[creature] {
				t.Error("a creature spell was offered to \"target instant or sorcery spell\"")
			}
			g.WithWriteLock(func() { _ = g.CounterTargetForEffect(creature) })
			if err := g.ActivateCatalogAbility(me.ID, flect, 0, game.ActivateAbilityParams{
				Targets: effectstest.RtCardTarget(bolt),
			}); err != nil {
				t.Fatalf("activate Goblin Flectomancer: %v", err)
			}
			if g.Battlefield.Contains(flect) {
				t.Error("the sacrifice cost was not paid at announce")
			}
			effectstest.PassPriorityAroundTable(t, g)
			prompt := effectstest.LatestRetarget(g, me.ID)
			if prompt == nil {
				t.Fatal("no retarget prompt")
			}
			if prompt.PickTargetMin != 0 {
				t.Errorf(`"you may" offers min %d, want 0`, prompt.PickTargetMin)
			}
			if accept {
				effectstest.AnswerRetargetWith(t, g, me.ID, effectstest.RtPlayerRef(victimB))
			} else {
				effectstest.AnswerRetargetWith(t, g, me.ID, nil)
			}
			effectstest.PassPriorityAroundTable(t, g)
			wantA, wantB := 40, 37
			if !accept {
				wantA, wantB = 37, 40
			}
			if effectstest.LifeOf(g, victimA) != wantA || effectstest.LifeOf(g, victimB) != wantB {
				t.Errorf("life A=%d B=%d, want A=%d B=%d", effectstest.LifeOf(g, victimA), effectstest.LifeOf(g, victimB), wantA, wantB)
			}
		})
	}
}

func acLegalAbilityTargetIDs(g *game.Game, chooser uuid.UUID, oracle string) []uuid.UUID {
	var out []uuid.UUID
	for id := range effectstest.AcLegalAbilityTargets(g, chooser, oracle, 0) {
		out = append(out, id)
	}
	return out
}

// --- Muck Drubb ----------------------------------------------------------

// A Bolt at a creature is pulled onto the Drubb; a Bolt at a player is
// not a legal target for the trigger's clause at all.
func TestMuckDrubbTakesASpellThatTargetsOnlyACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.PushCatalogPermanent(g, opp.ID, "Victim", "Creature — Bear", "", false)
	atCreature := effectstest.CastFor(t, g, me, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		game.CastSpellParams{Targets: effectstest.RtCardTarget(victim)})
	atPlayer := effectstest.BoltFrom(t, g, me, opp.ID)

	drubb := effectstest.CastFor(t, g, me, "Muck Drubb", "Creature — Beast", muckDrubbOracle, game.CastSpellParams{})
	g.WithWriteLock(func() {
		for i := range g.Stack.Cards {
			if g.Stack.Cards[i].InstanceID == drubb {
				g.Stack.Cards[i].Power, g.Stack.Cards[i].Toughness = 4, 4
			}
		}
	})
	for i := 0; i < 4 && effectstest.LatestPickTarget(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("Muck Drubb's trigger asked for no target")
	}
	if !effectstest.HasID(pick.PickTargetCards, atCreature) {
		t.Errorf("a spell aimed only at a creature is not offered: %v", pick.PickTargetCards)
	}
	if effectstest.HasID(pick.PickTargetCards, atPlayer) {
		t.Error("a spell aimed at a player was offered to \"targets only a single creature\"")
	}
	effectstest.PickCard(t, g, me.ID, atCreature)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(victim) {
		t.Error("the original creature died: the Bolt was not redirected")
	}
	if effectstest.DamageMarkedOn(g, drubb) != 3 {
		t.Errorf("Muck Drubb has %d damage, want 3", effectstest.DamageMarkedOn(g, drubb))
	}
}

// --- Insidious Will -------------------------------------------------------

func TestInsidiousWillCountersTargetSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1].ID
	bolt := effectstest.BoltFrom(t, g, me, victim)
	effectstest.CastFor(t, g, me, "Insidious Will", "Instant", insidiousWillOracle,
		game.CastSpellParams{Modes: []int{0}, Targets: effectstest.RtCardTarget(bolt)})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LifeOf(g, victim) != 40 {
		t.Errorf("the Bolt was not countered: life %d", effectstest.LifeOf(g, victim))
	}
}

func TestInsidiousWillMayChooseNewTargets(t *testing.T) {
	for _, accept := range []bool{true, false} {
		name := "declined"
		if accept {
			name = "accepted"
		}
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, victimA, victimB := g.Seats[0], g.Seats[1].ID, g.Seats[2].ID
			bolt := effectstest.BoltFrom(t, g, me, victimA)
			effectstest.CastFor(t, g, me, "Insidious Will", "Instant", insidiousWillOracle,
				game.CastSpellParams{Modes: []int{1}, Targets: effectstest.RtCardTarget(bolt)})
			effectstest.PassPriorityAroundTable(t, g)
			prompt := effectstest.LatestRetarget(g, me.ID)
			if prompt == nil {
				t.Fatal("no retarget prompt")
			}
			if prompt.PickTargetMin != 0 {
				t.Errorf(`"you may" offers min %d, want 0`, prompt.PickTargetMin)
			}
			if accept {
				effectstest.AnswerRetargetWith(t, g, me.ID, effectstest.RtPlayerRef(victimB))
			} else {
				effectstest.AnswerRetargetWith(t, g, me.ID, nil)
			}
			effectstest.PassPriorityAroundTable(t, g)
			wantA, wantB := 40, 37
			if !accept {
				wantA, wantB = 37, 40
			}
			if effectstest.LifeOf(g, victimA) != wantA || effectstest.LifeOf(g, victimB) != wantB {
				t.Errorf("life A=%d B=%d, want A=%d B=%d", effectstest.LifeOf(g, victimA), effectstest.LifeOf(g, victimB), wantA, wantB)
			}
		})
	}
}

// The copy bullet takes instants and sorceries only, and the copy may
// be pointed elsewhere.
func TestInsidiousWillCopiesAnInstantWithNewTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victimA, victimB := g.Seats[0], g.Seats[1].ID, g.Seats[2].ID
	bolt := effectstest.BoltFrom(t, g, me, victimA)
	creature := effectstest.CastFor(t, g, me, "Mizzium Meddler", "Creature — Vedalken Wizard", effectstest.MizziumMeddlerOracle, game.CastSpellParams{})
	legal := legalSpellTargetsForMode(g, me.ID)
	if !effectstest.HasID(legal, bolt) || effectstest.HasID(legal, creature) {
		t.Errorf("copy bullet legal targets = %v, want the Bolt and not the creature spell", legal)
	}
	g.WithWriteLock(func() { _ = g.CounterTargetForEffect(creature) })
	effectstest.CastFor(t, g, me, "Insidious Will", "Instant", insidiousWillOracle,
		game.CastSpellParams{Modes: []int{2}, Targets: effectstest.RtCardTarget(bolt)})
	for i := 0; i < 8 && effectstest.LatestPickTarget(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no re-target prompt for the copy")
	}
	if err := g.ResolvePickTarget(prompt.ID, me.ID,
		game.TargetRef{Kind: game.TargetPlayer, ID: victimB}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LifeOf(g, victimA) != 37 || effectstest.LifeOf(g, victimB) != 37 {
		t.Errorf("life A=%d B=%d, want both 37 (the original and the copy)", effectstest.LifeOf(g, victimA), effectstest.LifeOf(g, victimB))
	}
}

func legalSpellTargetsForMode(g *game.Game, chooser uuid.UUID) []uuid.UUID {
	spec, ok := effects.Lookup(insidiousWillOracle)
	if !ok || spec.Modes == nil || len(spec.Modes.Options) < 3 {
		return nil
	}
	var out []uuid.UUID
	g.WithWriteLock(func() {
		out = g.LegalTargetsForEffect(game.SourceChooser(chooser), spec.Modes.Options[2].Targets).Cards
	})
	return out
}
