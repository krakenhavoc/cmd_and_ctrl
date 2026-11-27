package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b581_cards_test.go — Ravnica: City of Guilds: removal, Auras, instants
// and sorceries (S82 older sets, batch 581, #5600). Each test plays the real
// catalog entry through the engine.

const (
	b581SeismicSpike = "25ec6309-0ccb-4f27-80f1-ec6dcc6461fa"
	b581Sisters      = "38e64b15-47c0-4453-a51e-a53c1aebc8d0"
	b581RollingSpoil = "9abe09fb-17ab-41e8-9d89-1467ed62f87b"
	b581Gaze         = "ec62aed5-b59b-4ba0-b265-a9c2a798ffb3"
	b581InstillFuror = "5726c4b1-0613-4724-b1fc-2e7f398bf1af"
	b581DreamLeash   = "b7fde102-7b89-48c7-818c-bdf0eb33c2ee"
	b581Flight       = "cd20a2a4-5e5e-420d-9420-651bce511f76"
	b581Mark         = "d9193641-a7f4-40d3-ac62-15015acece4e"
	b581Tunnel       = "0e134040-1fe3-410c-aac5-94ae1bd17637"
	b581Razia        = "60e95faf-1496-4298-9588-52b50b05fc6b"
	b581Nexus        = "6fe302f2-5594-478c-a66b-89b4b96e3611"
	b581Bear         = "Creature — Bear"
)

func TestOldB581EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b581SeismicSpike: "Seismic Spike", b581Sisters: "Sisters of Stone Death", b581RollingSpoil: "Rolling Spoil",
		b581Gaze: "Gaze of the Gorgon", b581InstillFuror: "Instill Furor", b581DreamLeash: "Dream Leash",
		b581Flight: "Flight of Fancy", b581Mark: "Mark of Eviction", b581Tunnel: "Tunnel Vision",
		b581Razia: "Razia's Purification", b581Nexus: "Mnemonic Nexus",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// b581Hand puts a card in the hand and returns its ID.
func b581Hand(p *game.Player, name, typeLine, oracle, cost string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: cost, Owner: p.ID, Controller: p.ID})
	return id
}

// b581Cast casts a card from the hand of `p` (strict, paying with `mana`).
func b581Cast(t *testing.T, g *game.Game, p *game.Player, name, typeLine, oracle, cost, mana string, targets []game.TargetRef) error {
	t.Helper()
	id := b581Hand(p, name, typeLine, oracle, cost)
	if mana != "" {
		effectstest.FloatForTest(g, p, mana)
	}
	return g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true, Targets: targets})
}

func b581SetTapped(g *game.Game, id uuid.UUID, tapped bool) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = tapped
			}
		}
	})
}

// --- Dream Leash ------------------------------------------------------------

func TestOldB581DreamLeashTargetsOnlyATappedPermanentAsCastAndKeepsControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tapped := effectstest.SeedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island")
	b581SetTapped(g, tapped, true)
	untapped := effectstest.SeedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island")
	b581SetTapped(g, untapped, false)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	leash := b581Hand(me, "Dream Leash", effectstest.AuraTypeLine, b581DreamLeash, "")
	card := me.Hand.Cards[len(me.Hand.Cards)-1]
	legal := g.LegalTargetsFor(game.SourceCast(me.ID, &card), b581DreamLeash)
	if legal == nil {
		t.Fatal("no legal-target set")
	}
	has := func(id uuid.UUID) bool {
		for _, c := range legal.Cards {
			if c == id {
				return true
			}
		}
		return false
	}
	if !has(tapped) || has(untapped) {
		t.Fatalf("legal set %v: want the tapped land only", legal.Cards)
	}
	err := g.CastSpell(me.ID, leash, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(untapped)})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("cast at an untapped permanent = %v, want ErrIllegalTarget", err)
	}
	if err := g.CastSpell(me.ID, leash, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(tapped)}); err != nil {
		t.Fatalf("cast at a tapped permanent: %v", err)
	}
	b581SetTapped(g, tapped, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(leash) {
		t.Fatal("Dream Leash did not resolve after its target untapped")
	}
	if host := effectstest.AttachmentHostOf(t, g, leash); host.ID != tapped {
		t.Fatalf("attached to %v, want the land", host.ID)
	}
	if got := effectstest.ControllerOf(t, g, tapped); got != me.ID {
		t.Errorf("the land is controlled by %s, want the Aura's controller", got)
	}
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(leash) {
		t.Error("the Aura fell off an untapped permanent")
	}
}

func TestOldB581DreamLeashEnchantsACreatureToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", b581Bear, 2, 2)
	b581SetTapped(g, bear, true)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := b581Cast(t, g, me, "Dream Leash", effectstest.AuraTypeLine, b581DreamLeash, "", "", effectstest.CardTarget(bear)); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, bear); got != me.ID {
		t.Errorf("controller %s, want me", got)
	}
}

// --- Flight of Fancy --------------------------------------------------------

func TestOldB581FlightOfFancyDrawsTwoAndGrantsFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", b581Bear, 2, 2)
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	hand := me.Hand.Size()
	if err := b581Cast(t, g, me, "Flight of Fancy", effectstest.AuraTypeLine, b581Flight, "", "", effectstest.CardTarget(bear)); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d → %d, want +2 (the Aura left it and two cards came in)", hand, got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the enchanted creature does not fly")
	}
}

// --- Mnemonic Nexus ---------------------------------------------------------

func TestOldB581MnemonicNexusShufflesEveryGraveyardIntoItsLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for _, p := range []*game.Player{me, opp} {
		for i := 0; i < 3; i++ {
			p.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
		}
	}
	effectstest.SeedLibrary(me, "A", "B")
	effectstest.SeedLibrary(opp, "C")
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := b581Cast(t, g, me, "Mnemonic Nexus", "Instant", b581Nexus, "", "", nil); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range []*game.Player{me, opp} {
		if n := p.Graveyard.Size(); n != 0 && p == opp {
			t.Errorf("the opponent's graveyard has %d cards, want 0", n)
		}
	}
	if n := me.Graveyard.Size(); n != 1 {
		t.Errorf("my graveyard has %d cards, want only Mnemonic Nexus", n)
	}
	if n := opp.Library.Size(); n < 4 {
		t.Errorf("the opponent's library has %d cards, want at least 4", n)
	}
	if n := me.Library.Size(); n < 5 {
		t.Errorf("my library has %d cards, want at least 5", n)
	}
}

// --- Seismic Spike ----------------------------------------------------------

func TestOldB581SeismicSpikeDestroysALandAndAddsRR(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.SeedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", b581Bear, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := b581Cast(t, g, me, "Seismic Spike", "Sorcery", b581SeismicSpike, "{2}{R}{R}", "RRCC", effectstest.CardTarget(bear)); err == nil {
		t.Fatal("a creature was accepted as the target")
	}
	me.ManaPool = nil
	if err := b581Cast(t, g, me, "Seismic Spike", "Sorcery", b581SeismicSpike, "{2}{R}{R}", "RRCC", effectstest.CardTarget(land)); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land survived")
	}
	if got := effectstest.PoolCount(me, "R"); got != 2 {
		t.Errorf("%d red mana in the pool, want 2", got)
	}
}

// --- Rolling Spoil ----------------------------------------------------------

func TestOldB581RollingSpoilShrinksCreaturesOnlyWhenBlackWasSpent(t *testing.T) {
	for _, tc := range []struct {
		name string
		mana string
		want int
	}{{"black spent", "GGBC", 1}, {"no black", "GGCC", 2}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			land := effectstest.SeedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island")
			mine := effectstest.B12Creature(g, me.ID, "Mine", b581Bear, 2, 2)
			theirs := effectstest.B12Creature(g, opp.ID, "Theirs", b581Bear, 2, 2)
			effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
			if err := b581Cast(t, g, me, "Rolling Spoil", "Sorcery", b581RollingSpoil, "{2}{G}{G}", tc.mana, effectstest.CardTarget(land)); err != nil {
				t.Fatal(err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(land) {
				t.Error("the land survived")
			}
			for _, id := range []uuid.UUID{mine, theirs} {
				if p, tough := effectstest.PtOf(t, g, id); p != tc.want || tough != tc.want {
					t.Errorf("a creature is %d/%d, want %d/%d", p, tough, tc.want, tc.want)
				}
			}
		})
	}
}

// --- Mark of Eviction -------------------------------------------------------

func b581ToUpkeepOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < 400; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat == seat {
			return
		}
	}
	t.Fatal("never reached the upkeep")
}

func TestOldB581MarkOfEvictionReturnsTheCreatureAndEveryAuraOnIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, opp.ID, "Host", b581Bear, 2, 2)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", b581Bear, 2, 2)
	mark := effectstest.Cmdr2B19Attach(g, me.ID, host, "Mark of Eviction", b581Mark)
	other := effectstest.Cmdr2B19Attach(g, opp.ID, host, "Other Aura", "")
	elsewhere := effectstest.Cmdr2B19Attach(g, opp.ID, bystander, "Elsewhere Aura", "")
	b581ToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{host, mark, other} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s is still on the battlefield", effectstest.CardNameOf(g, id))
		}
	}
	if !opp.Hand.Contains(host) || !opp.Hand.Contains(other) {
		t.Error("the creature and its other Aura are not in their owner's hand")
	}
	if !me.Hand.Contains(mark) {
		t.Error("Mark of Eviction is not in its owner's hand")
	}
	if !g.Battlefield.Contains(bystander) || !g.Battlefield.Contains(elsewhere) {
		t.Error("a permanent that was not enchanted by it was returned")
	}
}

func TestOldB581MarkOfEvictionDoesNothingOnAnotherPlayersUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, opp.ID, "Host", b581Bear, 2, 2)
	effectstest.Cmdr2B19Attach(g, me.ID, host, "Mark of Eviction", b581Mark)
	b581ToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(host) {
		t.Error("the creature was returned on its own controller's upkeep")
	}
}

// --- Instill Furor ----------------------------------------------------------

func TestOldB581InstillFurorSacrificesTheCreatureAtItsEndStepUnlessItAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", b581Bear, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := b581Cast(t, g, me, "Instill Furor", effectstest.AuraTypeLine, b581InstillFuror, "", "", effectstest.CardTarget(bear)); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature that did not attack was not sacrificed")
	}
}

func TestOldB581InstillFurorSparesACreatureThatAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, me.ID, "Bear", b581Bear, 2, 2)
	effectstest.Cmdr2B19Attach(g, me.ID, bear, "Instill Furor", b581InstillFuror)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature that attacked was sacrificed")
	}
}

// --- Razia's Purification ---------------------------------------------------

func TestOldB581RaziasPurificationKeepsThreePermanentsPerPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	var mine, theirs, thirds []uuid.UUID
	for i := 0; i < 5; i++ {
		mine = append(mine, effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest"))
	}
	mine = append(mine, effectstest.B12Creature(g, me.ID, "Mine", b581Bear, 2, 2))
	for i := 0; i < 4; i++ {
		theirs = append(theirs, effectstest.SeedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island"))
	}
	for i := 0; i < 3; i++ {
		thirds = append(thirds, effectstest.SeedLandOnBattlefield(g, third.ID, "Swamp", "Basic Land — Swamp"))
	}
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := b581Cast(t, g, me, "Razia's Purification", "Sorcery", b581Razia, "", "", nil); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil || pick.ChooseMin != 3 || pick.ChooseMax != 3 {
		t.Fatalf("want a pick of exactly three permanents, got %+v", pick)
	}
	if effectstest.ChooseCardsChoiceFor(g, third.ID) != nil {
		t.Error("a player with exactly three permanents was asked")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, mine[:2]); err == nil {
		t.Fatal("two permanents were accepted")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, mine[3], mine[4], mine[5])
	effectstest.AnswerChooseCards(t, g, opp.ID, theirs[1], theirs[2], theirs[3])
	effectstest.PassPriorityAroundTable(t, g)
	for i, id := range mine {
		if want := i >= 3; g.Battlefield.Contains(id) != want {
			t.Errorf("my permanent %d on the battlefield %v, want %v", i, !want, want)
		}
	}
	for i, id := range theirs {
		if want := i >= 1; g.Battlefield.Contains(id) != want {
			t.Errorf("their permanent %d on the battlefield %v, want %v", i, !want, want)
		}
	}
	for _, id := range thirds {
		if !g.Battlefield.Contains(id) {
			t.Error("a permanent of the player with three was sacrificed")
		}
	}
}

// --- Tunnel Vision ----------------------------------------------------------

func TestOldB581TunnelVisionPutsTheNamedCardOnTopAndTheRestIntoTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := opp.Library.Size()
	ids := effectstest.SeedLibrary(opp, "A", "B", "Wanted", "C", "D")
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := b581Cast(t, g, me, "Tunnel Vision", "Sorcery", b581Tunnel, "", "", []game.TargetRef{effectstest.PlayerTarget(opp.ID)}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ActNameFor(t, g, me.ID, "wanted")
	effectstest.PassPriorityAroundTable(t, g)
	top, err := opp.Library.Top()
	if err != nil || top.InstanceID != ids[2] {
		t.Fatalf("the top of the library is %q, want Wanted", top.Name)
	}
	if opp.Library.Size() != before+3 {
		t.Errorf("library has %d cards, want %d (Wanted, C, D and the rest)", opp.Library.Size(), before+3)
	}
	if !opp.Graveyard.Contains(ids[0]) || !opp.Graveyard.Contains(ids[1]) || opp.Graveyard.Size() != 2 {
		t.Errorf("graveyard has %d cards, want exactly A and B", opp.Graveyard.Size())
	}
}

func TestOldB581TunnelVisionShufflesWhenNoCardHasThatName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(opp, "A", "B", "C")
	before := opp.Library.Size()
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := b581Cast(t, g, me, "Tunnel Vision", "Sorcery", b581Tunnel, "", "", []game.TargetRef{effectstest.PlayerTarget(opp.ID)}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ActNameFor(t, g, me.ID, "Nonexistent")
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != before || opp.Graveyard.Size() != 0 {
		t.Errorf("library %d graveyard %d, want %d and 0 (nothing is put into the graveyard)", opp.Library.Size(), opp.Graveyard.Size(), before)
	}
	shuffles := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventLibraryShuffled {
			shuffles++
		}
	}
	if shuffles == 0 {
		t.Error("the library was not shuffled")
	}
}

// --- Gaze of the Gorgon -----------------------------------------------------

func TestOldB581GazeOfTheGorgonRegeneratesAndDestroysItsCombatPartnersAtEndOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", b581Bear, 1, 5)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", b581Bear, 1, 5)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", b581Bear, 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	if err := b581Cast(t, g, me, "Gaze of the Gorgon", "Instant", b581Gaze, "", "", effectstest.CardTarget(attacker)); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(blocker) {
		t.Error("the blocker survived the end of combat")
	}
	if !g.Battlefield.Contains(attacker) {
		t.Error("the regenerated attacker died")
	}
	if !g.Battlefield.Contains(bystander) {
		t.Error("a creature that was not in combat with it was destroyed")
	}
}

func TestOldB581GazeOfTheGorgonDoesNothingWhenNothingBlockedOrWasBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", b581Bear, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", b581Bear, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := b581Cast(t, g, me, "Gaze of the Gorgon", "Instant", b581Gaze, "", "", effectstest.CardTarget(mine)); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(theirs) {
		t.Error("a creature was destroyed with no block")
	}
}

// --- Sisters of Stone Death -------------------------------------------------

func b581PushSisters(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B12Push(g, owner, "Sisters of Stone Death", "Legendary Creature — Gorgon", b581Sisters, 7, 7)
}

func b581Activate(t *testing.T, g *game.Game, who, source uuid.UUID, idx int, targets []game.TargetRef, mana ...string) error {
	t.Helper()
	err := effectstest.CrtActivateRaw(t, g, who, source, idx, targets, mana...)
	if err == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	return err
}

func TestOldB581SistersOfStoneDeathMakesATargetBlockIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	sisters := b581PushSisters(g, me.ID)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", b581Bear, 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := b581Activate(t, g, me.ID, sisters, 0, effectstest.CardTarget(theirs), "G"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.DeclareAttack(t, g, opp.ID, sisters)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlockers(nil); err == nil {
		t.Error("the opponent declared no blockers although the target must block if able")
	}
}

func TestOldB581SistersOfStoneDeathExilesABlockerThenReturnsItUnderYourControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sisters := b581PushSisters(g, me.ID)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", b581Bear, 1, 6)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", b581Bear, 1, 6)
	effectstest.CrtBlock(g, blocker, sisters, opp.ID)
	if err := b581Activate(t, g, me.ID, sisters, 1, effectstest.CardTarget(bystander), "B", "G"); err == nil {
		t.Fatal("a creature not in combat with the Sisters was accepted")
	}
	me.ManaPool = nil
	if err := b581Activate(t, g, me.ID, sisters, 1, effectstest.CardTarget(blocker), "B", "G"); err != nil {
		t.Fatalf("exile: %v", err)
	}
	if g.Battlefield.Contains(blocker) || !g.Exile.Contains(blocker) {
		t.Fatal("the blocker was not exiled")
	}
	if err := b581Activate(t, g, me.ID, sisters, 2, nil, "C", "C", "B"); err != nil {
		t.Fatalf("return: %v", err)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, blocker)
	effectstest.PassPriorityAroundTable(t, g)
	var back *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Blocker" {
			back = &g.Battlefield.Cards[i]
		}
	}
	if back == nil {
		t.Fatalf("the exiled creature did not return (choices %d, exile has it %v)", len(g.PendingChoices), g.Exile.Contains(blocker))
	}
	if back.Controller != me.ID {
		t.Errorf("the returned creature is controlled by %s, want me", back.Controller)
	}
}

func TestOldB581SistersOfStoneDeathExilesAnAttackerItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sisters := b581PushSisters(g, opp.ID)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", b581Bear, 2, 2)
	effectstest.CrtBlock(g, sisters, attacker, opp.ID)
	if err := b581Activate(t, g, opp.ID, sisters, 1, effectstest.CardTarget(attacker), "B", "G"); err != nil {
		t.Fatalf("exile: %v", err)
	}
	if g.Battlefield.Contains(attacker) || !g.Exile.Contains(attacker) {
		t.Error("the attacker it blocked was not exiled")
	}
}
