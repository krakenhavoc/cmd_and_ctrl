package effects

import (
	"strconv"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// theme_deck_avacyn_test.go — S30's exit criterion (issue #678, the
// last open box on #95):
//
//	"Theme-deck smoke test (Avacyn + Reverberate + Clone + a fog
//	 plays 4 turns end-to-end)"
//
// A scripted engine test in the #506 / #645 shape, not a bot game.
// The bots cannot take these lines: the copy-target picker and the
// CR 707.10c re-target prompt are not in legal/cast.go's enumeration,
// so a bot-driven run would never choose what Clone copies or where
// the Reverberate copy points. The bot version waits on #89.
//
// The harness is the S27 smoke test's (theme_deck_smoke_test.go) and
// the S29 one's (theme_deck_graveyard_test.go), for the same three
// reasons:
//
//  1. THE CARDS COME OFF THE IMPORT ROAD. Every card is a cards.Card
//     row through deck.ToGameCard (themeDeckRow / importThemeCard).
//     Avacyn's flying, vigilance and indestructible therefore reach
//     the board the way a deck-imported card's do — through
//     Scryfall's `keywords` array and printedKeywords (#317 / #319 /
//     #320) — and not because the test wrote them onto a game.Card.
//
//  2. THE MANA IS REAL. Every cast goes through Strict + AutoTap, on
//     a four-colour mana base. Avacyn is eight mana with three white
//     pips; Reverberate is {R}{R} cast on somebody else's turn out of
//     what the previous turn's Clone left untapped. A cost this deck
//     could not actually pay fails here instead of resolving for
//     free.
//
//  3. THE TURNS ARE REAL. Four of the theme seat's turns at a
//     four-player table, with the other three seats' turns in
//     between, and the whole fog / Reverberate sequence happens on an
//     OPPONENT's turn at instant speed. Nothing is teleported into a
//     step: the fog expires because a cleanup happened, the blocked
//     creature survives because the combat-damage step ran, and the
//     Clone reverts because a zone change happened.
//
// The concessions, as #506 and #645 state theirs: sixteen lands start
// on the battlefield rather than being played one per turn (a land
// drop per turn caps turn 1 at one mana and puts an eight-drop out of
// reach for the length of the test, which would measure the mana
// curve instead of the cards), and the opening hand goes back into
// the library so the hand holds exactly the seven spells and no
// CR 514.1 discard gets in the way. Nothing else is placed by hand
// except the creatures the opponent is supposed to have brought.
//
// WHAT IT ASSERTS, BY CARD
//
//	Avacyn      her own printed indestructible and the grant to
//	            OTHER permanents you control (CR 702.12b) — a Wrath
//	            of God leaves your board and takes the opponent's,
//	            and a 2/2 that blocks a 5/5 keeps its damage marked
//	            and lives (CR 704.5g).
//	Reverberate the copy runs the original's effect against DIFFERENT
//	            targets (CR 707.10c), is not cast (no EventCast), and
//	            ceases to exist rather than reaching a graveyard
//	            (CR 707.10).
//	Clone       enters as a copy of a creature and has that
//	            creature's COPIABLE values (CR 707.2) — the copied
//	            creature's +1/+1 counter is not one of them — and on
//	            leaving the battlefield is a card named Clone again
//	            (CR 400.7).
//	Fog         combat damage that turn is prevented, non-combat
//	            damage on the same turn is not, and the prevention is
//	            gone the next turn (CR 615.6).
//
// Plus the two whole-run checks #678 asks for: no EventEffectError
// anywhere in the four turns, and a round-number check on every turn
// boundary, because advanceToMainOf returns at once when the cursor
// is already in that seat's main phase and would otherwise replay a
// turn silently — the bug an early draft of #645 had.
//
// WHAT IS DELIBERATELY NOT HERE
//
// Morph and protection, S30's two unshipped primitives (#656, #662);
// a copy of a PERMANENT spell, which is still an EventEffectError by
// design until #666; and the charged prevention shields, which are
// Mending Hands' business and are pinned in prevention_test.go. The
// four cards the criterion names are the four cards under test.

// The oracle keys are all declared by the focused S30 tests:
// avacynOracle in voltron_test.go, wrathOfGodOracle in
// s23_boardwipes_test.go, oracleClone in copy_effects_test.go,
// reverberateOracle in spell_copy_test.go, fogOracle in fog_test.go,
// lightningBoltOracle in cast_triggers_test.go and b829UnsummonOracle
// in once_per_batch_test.go. Reused rather than restated so a re-keyed
// card breaks in one place.

// avacynThemeLands is the mana base. Plains lead for the reason
// themeLandRow records: generic mana comes out of the front of the
// list, so the pips this deck needs LATE — Fog's {G}, Reverberate's
// {R}{R}, Unsummon's {U} — are still there after Avacyn has eaten
// eight mana and Clone another three out of the same untap cycle.
var avacynThemeLands = []themeLandRow{
	{"Plains", "Basic Land — Plains", "W", 8},
	{"Island", "Basic Land — Island", "U", 3},
	{"Mountain", "Basic Land — Mountain", "R", 3},
	{"Forest", "Basic Land — Forest", "G", 2},
}

// seedAvacynThemeDeck puts the lands on the battlefield and the seven
// spells in hand, all through deck.ToGameCard, after moving the
// opening hand back into the library. Returns the spells' instance
// IDs by name.
func seedAvacynThemeDeck(t *testing.T, g *game.Game, p *game.Player) map[string]uuid.UUID {
	t.Helper()

	avacyn := themeDeckRow("Avacyn, Angel of Hope", avacynOracle,
		"Legendary Creature — Angel", "{5}{W}{W}{W}")
	avacyn.Power, avacyn.Toughness = "8", "8"
	avacyn.Keywords = []string{"Flying", "Vigilance", "Indestructible"}

	clone := themeDeckRow("Clone", oracleClone, "Creature — Shapeshifter", "{2}{U}")
	clone.Power, clone.Toughness = "0", "0"

	spells := []cards.Card{
		avacyn,
		themeDeckRow("Wrath of God", wrathOfGodOracle, "Sorcery", "{2}{W}{W}"),
		clone,
		themeDeckRow("Fog", fogOracle, "Instant", "{G}"),
		themeDeckRow("Lightning Bolt", lightningBoltOracle, "Instant", "{R}"),
		themeDeckRow("Reverberate", reverberateOracle, "Instant", "{R}{R}"),
		themeDeckRow("Unsummon", b829UnsummonOracle, "Instant", "{U}"),
	}
	return seedThemeSeat(t, g, p, avacynThemeLands, spells)
}

// themeLandRow is one basic-land line of a theme deck's mana base.
// The ORDER of the rows matters and the auto-tapper is why:
// `solveColored` reserves sources for the coloured pips first and
// `recruitGeneric` then walks the restrictiveness-sorted list from the
// front, where every basic scores the same — so generic mana comes out
// of whatever was pushed first, and the colour a later turn still
// needs goes last.
type themeLandRow struct {
	name, typeLine, color string
	n                     int
}

// seedThemeSeat is the shared theme-deck setup for #678 and #726:
// return the opening hand to the library so the hand holds exactly the
// deck under test and no CR 514.1 discard gets in the way, put the
// pre-placed lands on the battlefield, then deal the spells. Every
// card goes through deck.ToGameCard. Returns the spells' instance IDs
// by name.
func seedThemeSeat(t *testing.T, g *game.Game, p *game.Player, lands []themeLandRow, spells []cards.Card) map[string]uuid.UUID {
	t.Helper()

	for len(p.Hand.Cards) > 0 {
		if _, err := game.MoveCard(p.Hand, p.Library, p.Hand.Cards[0].InstanceID); err != nil {
			t.Fatalf("return the opening hand: %v", err)
		}
	}
	for _, l := range lands {
		row := themeDeckRow(l.name, "", l.typeLine, "")
		row.ProducedMana = []string{l.color}
		for i := 0; i < l.n; i++ {
			pushBattlefieldCardWithTimestamp(g, importThemeCard(row, p.ID))
		}
	}
	ids := make(map[string]uuid.UUID, len(spells))
	for _, row := range spells {
		c := importThemeCard(row, p.ID)
		ids[c.Name] = c.InstanceID
		p.Hand.PushTop(c)
	}
	return ids
}

// themeCreatureRow is themeDeckRow for a card with a printed body:
// power, toughness and the Scryfall `keywords` array the import road
// turns into game.Card.Keywords.
func themeCreatureRow(name, typeLine, manaCost string, power, toughness int, keywords ...string) cards.Card {
	row := themeDeckRow(name, "", typeLine, manaCost)
	row.Power, row.Toughness = strconv.Itoa(power), strconv.Itoa(toughness)
	row.Keywords = append([]string(nil), keywords...)
	return row
}

// seedThemeCreature puts one imported creature onto the battlefield
// under `owner`'s control. The opponent's board is placed rather than
// cast because this deck has no way to give an opponent creatures and
// the criterion is about the four spells, not about the bodies they
// answer — the same concession theme_deck_smoke_test.go makes for its
// Raging Bear and Hulking Ogre.
func seedThemeCreature(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int, keywords ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g,
		importThemeCard(themeCreatureRow(name, typeLine, "", power, toughness, keywords...), owner))
}

// currentStatsOf returns a permanent's combat-relevant power and
// toughness through a snapshot, so the layer recompute has run.
// effectivePower / effectiveToughness read Effective() alone, which
// is the post-layer characteristic WITHOUT counters — Card.CurrentPower
// and CurrentToughness are where the +1/+1 counters are folded in — so
// an assertion about counters has to come through here.
func currentStatsOf(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	var power, toughness int
	var found bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID != id {
				continue
			}
			power, toughness, found = c.CurrentPower(), c.CurrentToughness(), true
			return
		}
	})
	if !found {
		t.Fatalf("card %s not on battlefield", id)
	}
	return power, toughness
}

// handCardByID returns a card from a seat's hand by instance ID.
func handCardByID(g *game.Game, p *game.Player, id uuid.UUID) (game.Card, bool) {
	var out game.Card
	var found bool
	g.ReadSnapshot(func() {
		for _, c := range p.Hand.Cards {
			if c.InstanceID == id {
				out, found = c, true
				return
			}
		}
	})
	return out, found
}

// countNamedIn counts cards with `name` in one zone.
func countNamedIn(z *game.Zone, name string) int {
	n := 0
	for _, c := range z.Cards {
		if c.Name == name {
			n++
		}
	}
	return n
}

// TestAvacynThemeDeckPlaysFourTurns is S30's exit criterion. See the
// file comment for what each turn proves and why the harness is
// shaped the way it is.
func TestAvacynThemeDeckPlaysFourTurns(t *testing.T) {
	g := newCatalogGame(t)
	themeSeat := g.Turn.ActiveSeat
	oppSeat := (themeSeat + 1) % len(g.Seats)
	thirdSeat := (themeSeat + 2) % len(g.Seats)
	me, opponent, third := g.Seats[themeSeat], g.Seats[oppSeat], g.Seats[thirdSeat]

	// Every event from here on is inside the four turns under test, so
	// the whole-run EventEffectError sweep starts at this mark.
	runStart := len(g.Events)

	hand := seedAvacynThemeDeck(t, g, me)

	// --- turn 1: Avacyn --------------------------------------------
	//
	// Eight mana out of the pre-placed base, and the two halves of
	// CR 702.12b land at once: the printed keyword on her, and the
	// static grant on everything else you control.
	advanceToMainOf(t, g, themeSeat)
	firstRound := g.Turn.Round

	bear := seedThemeCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	ogre := seedThemeCreature(g, opponent.ID, "Hulking Ogre", "Creature — Ogre", 5, 5)

	castThemeSpell(t, g, me, hand["Avacyn, Angel of Hope"])
	avacyn := hand["Avacyn, Angel of Hope"]
	if !onBattlefield(g, avacyn) {
		t.Fatal("Avacyn never reached the battlefield")
	}
	if p, tough := effectivePower(t, g, avacyn), effectiveToughness(t, g, avacyn); p != 8 || tough != 8 {
		t.Errorf("Avacyn is %d/%d, want 8/8 — printed P/T did not survive deck.ToGameCard", p, tough)
	}
	// Her own keywords rode in on the Scryfall record's `keywords`
	// array, not on the catalog entry: #317 / #319 / #320 put printed
	// keywords on the import road so this is the production path.
	for _, want := range []string{"flying", "vigilance", "indestructible"} {
		if ab := effectiveAbilities(t, g, avacyn); !containsString(ab, want) {
			t.Errorf("Avacyn's abilities %v missing %q", ab, want)
		}
	}
	// "Other permanents you control have indestructible" — the half
	// that is a real battlefield static, and it is controller-scoped.
	if ab := effectiveAbilities(t, g, bear); !containsString(ab, "indestructible") {
		t.Errorf("my creature's abilities %v missing indestructible (CR 702.12b grant)", ab)
	}
	if ab := effectiveAbilities(t, g, ogre); containsString(ab, "indestructible") {
		t.Errorf("an OPPONENT's creature picked up Avacyn's grant: %v", ab)
	}

	// --- turn 2: Wrath of God --------------------------------------
	//
	// The one-sided board wipe that only exists because #470 / #502
	// taught the MASS destroy path to read indestructible; before
	// that every Wrath in the catalog destroyed an Avacyn.
	advanceToStepOf(t, g, themeSeat, game.StepUpkeep)
	advanceToMainOf(t, g, themeSeat)
	if g.Turn.Round != firstRound+1 {
		t.Fatalf("round at the second main phase = %d, want %d", g.Turn.Round, firstRound+1)
	}

	castThemeSpell(t, g, me, hand["Wrath of God"])
	if !onBattlefield(g, avacyn) {
		t.Error("Avacyn was destroyed by a board wipe (CR 702.12b)")
	}
	if !onBattlefield(g, bear) {
		t.Error("a creature holding Avacyn's grant was destroyed by a board wipe (CR 702.12b)")
	}
	if onBattlefield(g, ogre) {
		t.Fatal("the opponent's creature survived 'destroy all creatures'")
	}
	if !inGraveyardOf(g, opponent.ID, ogre) {
		t.Error("the destroyed creature is not in its owner's graveyard (CR 400.3)")
	}

	// The board the opponent rebuilds with. Seeded now, after the
	// wipe, so both survive to be attacked with and copied.
	bigOgre := seedThemeCreature(g, opponent.ID, "Hulking Ogre", "Creature — Ogre", 5, 5)
	angel := seedThemeCreature(g, opponent.ID, "Serra Angel", "Creature — Angel", 4, 4, "Flying", "Vigilance")
	// A +1/+1 counter on the Angel, so that when Clone copies her
	// there is something on the board that is NOT a copiable value.
	if err := g.AddCounter(angel, game.CounterPlusOne, 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if p, tough := currentStatsOf(t, g, angel); p != 5 || tough != 5 {
		t.Fatalf("setup: the counter-bearing Angel is %d/%d, want 5/5", p, tough)
	}

	// --- the opponent's turn: a 5/5 into a 2/2 ----------------------
	//
	// CR 704.5g is the SBA indestructible SKIPS, not damage it stops:
	// the blocker keeps five marked damage on a two toughness and
	// stays on the battlefield anyway.
	advanceToStepOf(t, g, oppSeat, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bigOgre, me.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	advanceToStepOf(t, g, oppSeat, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(bear, bigOgre); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	advanceToStepOf(t, g, oppSeat, game.StepCombatDamage)
	if !onBattlefield(g, bear) {
		t.Fatal("a 2/2 with indestructible died to five combat damage (CR 704.5g)")
	}
	if got := damageMarkedOn(g, bear); got != 5 {
		t.Errorf("marked damage on the indestructible blocker = %d, want 5 — "+
			"indestructible skips the SBA, it does not prevent the damage (CR 702.12b)", got)
	}
	if me.Life != 40 {
		t.Errorf("life = %d, want 40 — the attacker was blocked", me.Life)
	}

	// --- turn 3: Clone ----------------------------------------------
	advanceToStepOf(t, g, themeSeat, game.StepUpkeep)
	advanceToMainOf(t, g, themeSeat)
	if g.Turn.Round != firstRound+2 {
		t.Fatalf("round at the third main phase = %d, want %d", g.Turn.Round, firstRound+2)
	}

	clone := hand["Clone"]
	if err := g.CastSpell(me.ID, clone, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("cast Clone: %v", err)
	}
	// "You may have this creature enter as a copy of any creature on
	// the battlefield" is a choice made as the permanent enters, not
	// targeting — the Angel's controller never gets a say.
	resolveWithCopyChoice(t, g, angel)

	cloned, ok := battlefieldCardByID(g, clone)
	if !ok {
		t.Fatal("Clone never reached the battlefield")
	}
	if cloned.Name != "Serra Angel" {
		t.Errorf("the Clone's name = %q, want %q", cloned.Name, "Serra Angel")
	}
	if !cloned.IsCopy() {
		t.Error("the Clone does not report as a copy")
	}
	if cloned.Controller != me.ID {
		t.Errorf("the Clone's controller = %s, want the caster %s — copying an "+
			"opponent's creature does not hand it to them", cloned.Controller, me.ID)
	}
	// CR 707.2: the copiable values are the PRINTED characteristics.
	// The +1/+1 counter on the original is not one of them, so the
	// copy is the printed 4/4 and not the 5/5 on the board.
	if p, tough := currentStatsOf(t, g, clone); p != 4 || tough != 4 {
		t.Errorf("the Clone is %d/%d, want 4/4 — a counter on the copied creature is not "+
			"a copiable value (CR 707.2)", p, tough)
	}
	for _, want := range []string{"flying", "vigilance"} {
		if ab := effectiveAbilities(t, g, clone); !containsString(ab, want) {
			t.Errorf("the Clone's abilities %v missing %q (CR 707.2)", ab, want)
		}
	}
	// And it is a permanent its controller controls, so Avacyn's
	// grant reaches it the instant it arrives.
	if ab := effectiveAbilities(t, g, clone); !containsString(ab, "indestructible") {
		t.Errorf("the Clone's abilities %v missing Avacyn's grant", ab)
	}

	// --- the opponent's turn: the fog --------------------------------
	//
	// The fog goes on an OPPONENT's turn, at instant speed, out of
	// whatever Clone left untapped — which is the only way a fog is
	// ever actually cast.
	advanceToStepOf(t, g, oppSeat, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bigOgre, me.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	advanceToStepOf(t, g, oppSeat, game.StepDeclareBlockers)

	castThemeSpell(t, g, me, hand["Fog"])
	if got := len(g.TurnScopedReplacements); got != 1 {
		t.Fatalf("turn-scoped replacements after the fog resolved = %d, want 1", got)
	}

	// Non-combat damage on the same turn is untouched: the fog's
	// predicate reads the combat flag the combat-damage resolver sets
	// and nothing else sets.
	oppLifeBeforeBolt := opponent.Life
	castThemeSpellTargeting(t, g, me, hand["Lightning Bolt"],
		game.TargetRef{Kind: game.TargetPlayer, ID: opponent.ID})
	if got := oppLifeBeforeBolt - opponent.Life; got != 3 {
		t.Errorf("non-combat damage under a fog = %d, want 3 (the fog is combat-only)", got)
	}

	advanceToStepOf(t, g, oppSeat, game.StepCombatDamage)
	if me.Life != 40 {
		t.Errorf("life after an unblocked 5/5 attacked into a fog = %d, want 40", me.Life)
	}
	if got := damageMarkedOn(g, bigOgre); got != 0 {
		t.Errorf("the attacker took %d damage; nothing blocked it", got)
	}

	// --- turn 4: Reverberate, and the Clone goes home ----------------
	advanceToStepOf(t, g, themeSeat, game.StepUpkeep)
	advanceToMainOf(t, g, themeSeat)
	if g.Turn.Round != firstRound+3 {
		t.Fatalf("round at the fourth main phase = %d, want %d", g.Turn.Round, firstRound+3)
	}
	// CR 615.6 / ADR 0013: the fog was turn-scoped and the cleanup
	// swept it.
	if got := len(g.TurnScopedReplacements); got != 0 {
		t.Errorf("turn-scoped replacements a turn after the fog = %d, want 0", got)
	}

	// Unsummon on my own Clone, Reverberate on the Unsummon, and the
	// copy re-aimed at the creature the Clone is copying. Both halves
	// of CR 707.10c in one cast: the copy runs the same effect, and
	// it runs it somewhere else.
	unsummon, reverberate := hand["Unsummon"], hand["Reverberate"]
	castsBefore := countCatalogEvents(g, game.EventCast, runStart)
	graveBefore := me.Graveyard.Size()

	if err := g.CastSpell(me.ID, unsummon, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: clone}},
		Strict:  true, AutoTap: true,
	}); err != nil {
		t.Fatalf("cast Unsummon: %v", err)
	}
	if err := g.CastSpell(me.ID, reverberate, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: unsummon}},
		Strict:  true, AutoTap: true,
	}); err != nil {
		t.Fatalf("cast Reverberate: %v", err)
	}

	// Reverberate is on top and resolves first, which opens the
	// CR 707.10c prompt before the copy reaches the stack.
	for i := 0; i < 8 && latestPickTarget(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no re-target prompt for the Reverberate copy (CR 707.10c)")
	}
	if !hasID(prompt.PickTargetCards, angel) {
		t.Fatalf("the copy's legal set should offer the opponent's Angel: %v", prompt.PickTargetCards)
	}
	if err := g.ResolvePickTarget(prompt.ID, me.ID,
		game.TargetRef{Kind: game.TargetCard, ID: angel}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)

	// The copy bounced the Angel; the original bounced the Clone.
	if onBattlefield(g, angel) {
		t.Error("the re-targeted copy did not resolve against the Angel")
	}
	if !opponent.Hand.Contains(angel) {
		t.Error("the bounced Angel is not in its owner's hand")
	}
	if onBattlefield(g, clone) {
		t.Error("the original Unsummon did not resolve against the Clone")
	}

	// CR 400.7 / CR 707.2: the copy effect applied to the PERMANENT.
	// The card in hand is a Clone again.
	back, ok := handCardByID(g, me, clone)
	if !ok {
		t.Fatal("the bounced Clone is not in its owner's hand")
	}
	if back.Name != "Clone" {
		t.Errorf("the bounced permanent is named %q in hand, want %q (CR 400.7)", back.Name, "Clone")
	}
	if back.IsCopy() {
		t.Error("the card in hand still reports as a copy")
	}

	// The copy was never cast (CR 707.10: it is put onto the stack,
	// not cast), so exactly two casts happened on this turn.
	if got := countCatalogEvents(g, game.EventCast, runStart) - castsBefore; got != 2 {
		t.Errorf("EventCast count for the turn = %d, want 2 — a copy is put onto the "+
			"stack, never cast (CR 707.10)", got)
	}
	// And it ceased to exist rather than going anywhere: two cards
	// joined the graveyard, not three.
	if got := me.Graveyard.Size() - graveBefore; got != 2 {
		t.Errorf("graveyard grew by %d, want 2 — a copy is not a card (CR 707.10)", got)
	}
	if got := countNamedIn(me.Graveyard, "Unsummon"); got != 1 {
		t.Errorf("Unsummons in the graveyard = %d, want 1", got)
	}

	// --- turn 4 combat: the fog really was turn-scoped ---------------
	//
	// Avacyn attacks into a board the copy just emptied of fliers.
	advanceToStepOf(t, g, themeSeat, game.StepDeclareAttackers)
	oppLifeBeforeSwing := opponent.Life
	if err := g.DeclareAttacker(avacyn, opponent.ID); err != nil {
		t.Fatalf("DeclareAttacker with Avacyn: %v", err)
	}
	lockInAttacks(t, g)
	advanceToStepOf(t, g, themeSeat, game.StepCombatDamage)
	if got := oppLifeBeforeSwing - opponent.Life; got != 8 {
		t.Errorf("combat damage the turn after a fog = %d, want 8 (CR 615.6 — the "+
			"prevention was 'this turn')", got)
	}
	// Vigilance came off the Scryfall record too, so she is still
	// untapped after attacking (CR 702.20).
	if c, _ := battlefieldCardByID(g, avacyn); c.Tapped {
		t.Error("Avacyn tapped to attack; her printed vigilance did not survive the import road")
	}

	// The third seat never took part and never lost life, which is
	// the control on every "all creatures" and "each opponent" read
	// above.
	if third.Life != 40 {
		t.Errorf("an uninvolved seat's life = %d, want 40", third.Life)
	}

	// #678's whole-run check: nothing in four turns logged an effect
	// error. The catalog soak fails the nightly run on one of these,
	// and a theme deck is exactly where a card combination first
	// produces one.
	if n := countCatalogEvents(g, game.EventEffectError, runStart); n != 0 {
		t.Errorf("EventEffectError count across the four turns = %d, want 0", n)
	}
}
