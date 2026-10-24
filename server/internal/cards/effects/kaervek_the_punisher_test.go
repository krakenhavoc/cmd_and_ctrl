package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kaervek_the_punisher_test.go — Kaervek, the Punisher (#2947, #2910).

const kaervekOracle = "fa9a6083-75ab-41ae-b607-39d21d9a3d3a"

func kaervekSetup(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	kaervek := pushDiesCreatureForTest(g, me.ID, "Kaervek, the Punisher", kaervekOracle, "Legendary Creature — Human Warlock", 3, 3)
	toMainForCost(t, g)
	return g, me, opp, kaervek
}

// kaervekResolveOn commits a crime for me, picks target as Kaervek's
// target and resolves the trigger, leaving the copy offer open.
func kaervekResolveOn(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) {
	t.Helper()
	commitCrimeForMinerTest(g, me.ID)
	passPriorityAroundTable(t, g)
	answerTriggerTargets(t, g, me.ID, target)
	passPriorityAroundTable(t, g)
}

// Casting the copy costs its printed mana cost, makes the copy do what
// the card does, and costs 2 life; the original stays in exile.
func TestKaervekCastsACopyAtItsCostAndLosesTwoLife(t *testing.T) {
	g, me, _, _ := kaervekSetup(t)
	ritual := gyCard(me, "Dark Ritual", "Instant", "{B}", darkRitualOracle)
	life := me.Life

	kaervekResolveOn(t, g, me, ritual)
	if !inExile(g, ritual) {
		t.Fatal("the targeted black card is exiled")
	}
	cp := copyOf(t, g, "Dark Ritual")
	g.ReadSnapshot(func() {
		perm := g.CastPermissionOnCardForEffect(cp, game.ZoneExile)
		if perm == nil || perm.Player != me.ID {
			t.Fatalf("no permission for Kaervek's controller over the copy: %+v", perm)
		}
		if perm.Cost != "" {
			t.Errorf("permission cost = %q, want the printed cost (not free)", perm.Cost)
		}
	})
	if me.Life != life {
		t.Fatalf("life moved to %d before the copy was cast", me.Life)
	}
	if err := castCopy(g, me.ID, cp.InstanceID, true); err == nil {
		t.Fatal("the copy was cast with no {B}: the cast is not free")
	}
	floatForTest(g, me, "B")
	if err := castCopy(g, me.ID, cp.InstanceID, true); err != nil {
		t.Fatalf("cast the copy for {B}: %v", err)
	}
	if me.Life != life-2 {
		t.Errorf("life %d -> %d, want -2 for casting the copy", life, me.Life)
	}
	passPriorityAroundTable(t, g)
	if got := me.ManaPool.Total(); got != 3 {
		t.Errorf("mana pool holds %d, want {B}{B}{B} from the copied Dark Ritual", got)
	}
	if !inExile(g, ritual) {
		t.Error("the original card stays in exile")
	}
	if n := len(cardCopies(g)); n != 0 {
		t.Errorf("%d copies left after the copy resolved", n)
	}
}

// Declining the cast loses no life and the copy ceases to exist. Passing
// priority is not a way to decline: nobody has priority while the
// trigger resolves (CR 608.2g). The crime here is real: Lightning Bolt
// at an opponent, and Kaervek's trigger goes above it.
func TestKaervekDecliningTheCopyLosesNoLife(t *testing.T) {
	g, me, opp, kaervek := kaervekSetup(t)
	ritual := gyCard(me, "Dark Ritual", "Instant", "{B}", darkRitualOracle)
	life := me.Life

	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{playerTarget(opp.ID)})
	if latestPickTarget(g, me.ID) == nil {
		passPriorityAroundTable(t, g)
	}
	if latestPickTarget(g, me.ID) == nil {
		t.Fatal("targeting an opponent is a crime: Kaervek should ask for a target")
	}
	answerTriggerTargets(t, g, me.ID, ritual)
	if triggerOnStack(g, kaervek) == nil {
		t.Fatal("Kaervek's trigger is not on the stack")
	}
	passPriorityAroundTable(t, g)
	if err := g.PassPriority(); err == nil {
		t.Fatal("a pass went through while the copy's cast was being decided")
	}
	if n := len(cardCopies(g)); n != 1 {
		t.Fatalf("%d copies after the refused pass, want the one on offer", n)
	}
	declineResolutionCasts(t, g, me.ID)

	if !inExile(g, ritual) {
		t.Error("the card is exiled whether or not the copy is cast")
	}
	if n := len(cardCopies(g)); n != 0 {
		t.Errorf("%d copies left after declining: an uncast copy ceases to exist", n)
	}
	if me.Life != life {
		t.Errorf("life %d -> %d: no life is lost when the copy is not cast", life, me.Life)
	}
}

// A copy of a permanent card resolves as a token (CR 707.10b).
func TestKaervekCopyOfACreatureCardBecomesAToken(t *testing.T) {
	g, me, _, _ := kaervekSetup(t)
	nighthawk := gyCard(me, "Vampire Nighthawk", "Creature — Vampire Shaman", "{1}{B}{B}", vampireNighthawkOracle)

	kaervekResolveOn(t, g, me, nighthawk)
	cp := copyOf(t, g, "Vampire Nighthawk")
	floatForTest(g, me, "BBB")
	if err := castCopy(g, me.ID, cp.InstanceID, true); err != nil {
		t.Fatalf("cast the creature copy: %v", err)
	}
	passPriorityAroundTable(t, g)
	var token *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Vampire Nighthawk" {
			token = &g.Battlefield.Cards[i]
		}
	}
	if token == nil {
		t.Fatal("the creature copy did not resolve onto the battlefield")
	}
	if !token.IsToken() {
		t.Error("a copy of a permanent spell becomes a token")
	}
	if !inExile(g, nighthawk) {
		t.Error("the original card stays in exile")
	}
}

// Only black cards in your own graveyard can be chosen, and with none
// there the trigger is not put on the stack. Another player's crime
// does not trigger it.
func TestKaervekTargetsOnlyYourBlackCards(t *testing.T) {
	g, me, opp, kaervek := kaervekSetup(t)
	gyCard(me, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
	theirs := gyCard(opp, "Dark Ritual", "Instant", "{B}", darkRitualOracle)

	commitCrimeForMinerTest(g, opp.ID)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || triggerOnStack(g, kaervek) != nil {
		t.Fatal("another player's crime triggered Kaervek")
	}

	commitCrimeForMinerTest(g, me.ID)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("with no black card in your graveyard there is nothing to target")
	}
	if !opp.Graveyard.Contains(theirs) {
		t.Error("an opponent's graveyard card was touched")
	}

	ritual := gyCard(me, "Dark Ritual", "Instant", "{B}", darkRitualOracle)
	commitCrimeForMinerTest(g, me.ID)
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("a black card in your graveyard: Kaervek should ask for a target")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}); err == nil {
		t.Error("a card in an opponent's graveyard is not a legal target")
	}
	answerTriggerTargets(t, g, me.ID, ritual)
	passPriorityAroundTable(t, g)
	if !inExile(g, ritual) {
		t.Error("the chosen black card is exiled")
	}
}

// A target that left the graveyard in response is illegal: nothing is
// exiled or copied (CR 608.2b).
func TestKaervekDoesNothingWhenItsTargetLeaves(t *testing.T) {
	g, me, _, kaervek := kaervekSetup(t)
	ritual := gyCard(me, "Dark Ritual", "Instant", "{B}", darkRitualOracle)

	commitCrimeForMinerTest(g, me.ID)
	passPriorityAroundTable(t, g)
	answerTriggerTargets(t, g, me.ID, ritual)
	if triggerOnStack(g, kaervek) == nil {
		t.Fatal("Kaervek's trigger is not on the stack")
	}
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID},
		game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, ritual,
	); err != nil {
		t.Fatalf("move the target away: %v", err)
	}
	passPriorityAroundTable(t, g)
	if inExile(g, ritual) || !me.Hand.Contains(ritual) {
		t.Error("a target that left the graveyard is not exiled")
	}
	if n := len(cardCopies(g)); n != 0 {
		t.Errorf("%d copies made of a card that was never exiled", n)
	}
}

// resolutionCastFor is who's open resolution_cast prompt, or nil.
func resolutionCastFor(g *game.Game, who uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceResolutionCast && c.Chooser == who {
			return c
		}
	}
	return nil
}

// declineResolutionCasts answers who's resolution_cast prompt with "cast
// no more of them".
func declineResolutionCasts(t *testing.T, g *game.Game, who uuid.UUID) {
	t.Helper()
	c := resolutionCastFor(g, who)
	if c == nil {
		t.Fatal("no resolution_cast prompt to decline")
	}
	if err := g.ResolvePendingChoice(c.ID, who, nil); err != nil {
		t.Fatalf("decline the copies: %v", err)
	}
}

// #2947, CR 707.12 and CR 608.2g: the copy is cast while Kaervek's
// trigger is still resolving. The resolution waits on the caster's
// decision, nobody can pass, and the 2 life go only on the cast.
func TestKaervekResolutionQueuesOptionalCastDecision(t *testing.T) {
	g, me, _, _ := kaervekSetup(t)
	ritual := gyCard(me, "Dark Ritual", "Instant", "{B}", darkRitualOracle)
	life := me.Life

	kaervekResolveOn(t, g, me, ritual)
	c := resolutionCastFor(g, me.ID)
	if c == nil {
		t.Fatalf("pending choices: %v — the cast is decided inside the resolution", g.PendingChoices)
	}
	cp := copyOf(t, g, "Dark Ritual")
	if len(c.DiscardOptions) != 1 || c.DiscardOptions[0] != cp.InstanceID {
		t.Errorf("the prompt offers %v, want the one copy %v", c.DiscardOptions, cp.InstanceID)
	}
	if !g.ResolutionPaused() {
		t.Error("Kaervek's resolution is not held open by the cast decision")
	}
	if err := g.PassPriority(); err == nil {
		t.Error("a player passed priority in the middle of the resolution (CR 608.2g)")
	}
	if me.Life != life {
		t.Errorf("life moved to %d before anything was cast", me.Life)
	}

	floatForTest(g, me, "B")
	if err := g.CastForResolutionChoice(me.ID, c.ID, cp.InstanceID, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast the copy during the resolution: %v", err)
	}
	if me.Life != life-2 {
		t.Errorf("life %d -> %d, want -2 as the copy is cast", life, me.Life)
	}
	if resolutionCastFor(g, me.ID) != nil || g.ResolutionPaused() {
		t.Error("the decision is still open after the only copy was cast")
	}
	if g.Stack.Size() != 1 || g.Stack.Cards[0].InstanceID != cp.InstanceID {
		t.Fatalf("stack = %v, want the copy alone on top", g.Stack.Cards)
	}
	passPriorityAroundTable(t, g)
	if got := me.ManaPool.Total(); got != 3 {
		t.Errorf("mana pool holds %d, want {B}{B}{B} from the copied Dark Ritual", got)
	}
}

// The copy costs its printed mana cost even at a table that does not
// check mana: no unintended free cast. A refused cast leaves the
// decision open, and declining afterwards costs no life and leaves no
// copy behind.
func TestKaervekResolutionCastMustBePaidAndCanBeDeclined(t *testing.T) {
	g, me, _, _ := kaervekSetup(t)
	ritual := gyCard(me, "Dark Ritual", "Instant", "{B}", darkRitualOracle)
	life := me.Life

	kaervekResolveOn(t, g, me, ritual)
	cp := copyOf(t, g, "Dark Ritual")
	if err := castCopy(g, me.ID, cp.InstanceID, false); err == nil {
		t.Fatal("the copy was cast with no {B} and strict mana off: the cast is not free")
	}
	if resolutionCastFor(g, me.ID) == nil {
		t.Fatal("a refused cast closed the decision")
	}
	declineResolutionCasts(t, g, me.ID)
	if me.Life != life {
		t.Errorf("life %d -> %d: no life is lost when the copy is not cast", life, me.Life)
	}
	if n := len(cardCopies(g)); n != 0 {
		t.Errorf("%d copies left after declining: an uncast copy ceases to exist", n)
	}
	if !inExile(g, ritual) {
		t.Error("the original card stays in exile")
	}
	if g.ResolutionPaused() {
		t.Error("the resolution is still held after the decline")
	}
}

// An illegal target refuses the cast and keeps the decision open for a
// retry. The copy aimed at an opponent's creature is a crime of its own,
// so Kaervek triggers again; that trigger is put on the stack only once
// the first resolution has finished, above the copy (CR 608.2g, 603.3).
func TestKaervekResolutionCastTargetsAndTriggersAfterTheResolution(t *testing.T) {
	g, me, opp, kaervek := kaervekSetup(t)
	blade := gyCard(me, "Doom Blade", "Instant", "{1}{B}", doomBladeOracle)
	ritual := gyCard(me, "Dark Ritual", "Instant", "{B}", darkRitualOracle)
	bear := pushPermanentForTest(g, opp.ID, "Their Bear", "", "Creature — Bear")

	kaervekResolveOn(t, g, me, blade)
	c := resolutionCastFor(g, me.ID)
	if c == nil {
		t.Fatal("no cast decision")
	}
	cp := copyOf(t, g, "Doom Blade")
	floatForTest(g, me, "BB")
	if err := g.CastForResolutionChoice(me.ID, c.ID, cp.InstanceID, game.CastSpellParams{
		FromZone: "exile", Targets: []game.TargetRef{playerTarget(opp.ID)},
	}); err == nil {
		t.Fatal("Doom Blade's copy was cast at a player")
	}
	if resolutionCastFor(g, me.ID) == nil {
		t.Fatal("an illegal target closed the decision")
	}
	if err := g.CastForResolutionChoice(me.ID, c.ID, cp.InstanceID, game.CastSpellParams{
		FromZone: "exile", Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("cast the copy at the opponent's creature: %v", err)
	}
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the copy's crime did not trigger Kaervek once the resolution finished")
	}
	answerTriggerTargets(t, g, me.ID, ritual)
	trigger := triggerOnStack(g, kaervek)
	spell := g.StackMeta[cp.InstanceID]
	if trigger == nil || spell == nil {
		t.Fatalf("stack: trigger %v, copy %v — want both", trigger != nil, spell != nil)
	}
	if trigger.Seq <= spell.Seq {
		t.Errorf("Kaervek's new trigger (seq %d) is not above the copy (seq %d)", trigger.Seq, spell.Seq)
	}
}

// A table paused on the cast decision is a restore point: the prompt,
// the copy and its priced offer come back, and the restored decision is
// answered like the live one.
func TestKaervekResolutionCastSurvivesARestore(t *testing.T) {
	g, me, _, _ := kaervekSetup(t)
	ritual := gyCard(me, "Dark Ritual", "Instant", "{B}", darkRitualOracle)
	kaervekResolveOn(t, g, me, ritual)
	cp := copyOf(t, g, "Dark Ritual")

	r := restoreRoundTrip(t, g, true)
	rme := r.PlayerByIDForEffect(me.ID)
	c := resolutionCastFor(r, me.ID)
	if c == nil || len(c.DiscardOptions) != 1 || c.DiscardOptions[0] != cp.InstanceID {
		t.Fatalf("restored prompt = %+v, want the decision over the copy", c)
	}
	if !r.ResolutionPaused() {
		t.Error("the restored resolution is not held open")
	}
	life := rme.Life
	if err := castCopy(r, rme.ID, cp.InstanceID, true); err == nil {
		t.Fatal("the restored copy was cast with no {B}")
	}
	floatForTest(r, rme, "B")
	if err := castCopy(r, rme.ID, cp.InstanceID, true); err != nil {
		t.Fatalf("cast the restored copy: %v", err)
	}
	if rme.Life != life-2 || resolutionCastFor(r, me.ID) != nil {
		t.Errorf("restored cast: life %d (want %d), prompt still open %v", rme.Life, life-2, resolutionCastFor(r, me.ID) != nil)
	}
}
