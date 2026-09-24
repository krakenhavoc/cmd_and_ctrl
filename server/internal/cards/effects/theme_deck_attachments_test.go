package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// theme_deck_attachments_test.go — S24's exit criterion (issue #726,
// the last open box on #76):
//
//	"Theme-deck smoke test (equipment / voltron-adjacent deck plays
//	 3 turns)"
//
// A scripted engine test in the #506 / #645 / #678 shape, not a bot
// game. The bots cannot take these lines either: equip is a targeted
// sorcery-speed activated ability whose whole point is WHICH creature
// it moves to, and nothing in the bot's ability enumeration ranks
// that choice. The bot version waits with #89.
//
// Every other S24 test (attachments_test.go, mind_control_test.go,
// attachments_departed_source_test.go) pins one rule against a
// hand-built fixture. This one is the opposite shape on purpose:
//
//  1. THE CARDS COME OFF THE IMPORT ROAD. Every card is a cards.Card
//     row through deck.ToGameCard (themeDeckRow / importThemeCard),
//     so the Equipment arrives as an "Artifact — Equipment" whose
//     equip ability the catalog attaches by oracle ID, and the Auras
//     as "Enchantment — Aura" whose Enchant clause the announce-time
//     target gate reads. The fixture tests all stamp those fields
//     onto a game.Card directly.
//
//  2. THE MANA IS REAL. Every cast AND every equip goes through
//     Strict + AutoTap out of a pre-placed base. An equip cost the
//     board cannot pay fails here instead of quietly attaching.
//
//  3. THE TURNS ARE REAL. Three of the theme seat's turns at a
//     four-player table, with the opponent taking their own turn in
//     between and casting their own Aura and Equipment. Nothing is
//     teleported into a step: the sorcery-speed gate closes because
//     the cursor is in an upkeep, the stolen creature is summoning
//     sick because a turn has not come round, and the two
//     state-based actions fire because a creature really died.
//
// The concessions, as #506 / #645 / #678 state theirs: thirteen lands
// for the theme seat and five for the opponent start on the
// battlefield rather than being played one per turn; both opening
// hands go back into their libraries so each hand holds exactly the
// deck under test; and the opponent's one creature is placed rather
// than cast, because this deck has no way to give an opponent a
// body and the criterion is about what gets attached to it.
//
// WHAT IT ASSERTS
//
//	Equip        sorcery speed (CR 702.6a) — refused in an upkeep,
//	             taken in a main phase; "target creature you control";
//	             the attachment stamped, the static reaching the host,
//	             and a second activation MOVING it (CR 701.3a).
//	Aura         Enchant creature as an announce-time target
//	             (CR 303.4a), the attachment stamped as the Aura
//	             ENTERS and not at announce, and the
//	             enchanted-creature statics applied.
//	CR 704.5m    the enchanted creature dies; the Aura is put into
//	             its OWNER's graveyard, which is not the graveyard of
//	             the player whose creature it was on.
//	CR 704.5n    the same death; the Equipment STAYS on the
//	             battlefield and becomes unattached, ready to move.
//	Mind Control the layer-2 control change (CR 613.1b): control
//	             moves, ownership does not (CR 108.3), the creature is
//	             summoning sick for its new controller (CR 302.6) —
//	             and the Equipment its old controller had on it stays
//	             attached, stays theirs, and keeps buffing the
//	             creature for its new controller. Control reverts on
//	             its own when the Aura goes.
//	Undo/replay  an attach survives the room's undo road (Clone /
//	             RestoreFrom): equip, restore, the Equipment is
//	             unattached and its static is gone from the host, and
//	             the same activation attaches again.
//
// WHAT IS DELIBERATELY NOT HERE
//
// Protection (#662) and CR 704.5p (#675), S24's two open tails; the
// Curse branch of the attachment relation, which attaches to a PLAYER
// and is pinned in attachments_test.go; and Bruna's three-zone
// multi-select, which has no prompt to drive.

// The oracle keys are declared by the focused S24 tests —
// bootsOracle, greavesOracle, skullclampOracle, bonesplitterOracle
// and rancorOracle in attachments_test.go, mindControlOracle in
// mind_control_test.go. Only the two this package has never needed as
// a constant are new here.
const (
	attachThemePacifismOracle      = "5f5e0b10-c8cf-450c-bfd3-bcb0528ec330"
	attachThemeDarksteelPlateOracl = "b5b4cf54-ed5e-42d0-9d98-5fec76b0b0b8"
)

// equipIndexOf is the ability index equipTo uses — the equip ability
// is always the last one a piece of Equipment declares. Split out so
// the refusal paths, which must not go through equipTo's t.Fatalf,
// can name the same index.
func equipIndexOf(t *testing.T, g *game.Game, equipment uuid.UUID) int {
	t.Helper()
	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: oracleOfBattlefieldCard(t, g, equipment)})
	if len(abilities) == 0 {
		t.Fatal("equipment has no activated ability")
	}
	return len(abilities) - 1
}

// tryEquip activates an equip without settling the stack and hands
// back the announce error. The test asserts on the refusals; equipTo
// covers the ones that are supposed to work.
func tryEquip(t *testing.T, g *game.Game, controller, equipment, creature uuid.UUID) error {
	t.Helper()
	return g.ActivateCatalogAbility(controller, equipment, equipIndexOf(t, g, equipment),
		game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: creature}},
		})
}

// TestAttachmentThemeDeckPlaysThreeTurns is S24's exit criterion. See
// the file comment for what each turn proves.
func TestAttachmentThemeDeckPlaysThreeTurns(t *testing.T) {
	g := newCatalogGame(t)
	themeSeat := g.Turn.ActiveSeat
	oppSeat := (themeSeat + 1) % len(g.Seats)
	me, opponent := g.Seats[themeSeat], g.Seats[oppSeat]

	runStart := len(g.Events)

	hand := seedThemeSeat(t, g, me,
		[]themeLandRow{
			{"Forest", "Basic Land — Forest", "G", 7},
			{"Island", "Basic Land — Island", "U", 6},
		},
		[]cards.Card{
			themeCreatureRow("Raging Bear", "Creature — Bear", "{1}{G}", 2, 2),
			themeCreatureRow("Scout Squire", "Creature — Human Soldier", "{G}", 1, 1),
			themeDeckRow("Swiftfoot Boots", bootsOracle, equipTypeLine, "{2}"),
			themeDeckRow("Lightning Greaves", greavesOracle, equipTypeLine, "{2}"),
			themeDeckRow("Darksteel Plate", attachThemeDarksteelPlateOracl, equipTypeLine, "{3}"),
			themeDeckRow("Skullclamp", skullclampOracle, equipTypeLine, "{1}"),
			themeDeckRow("Rancor", rancorOracle, auraTypeLine, "{G}"),
		})
	// Mind Control is the eighth card and would put the hand over the
	// CR 514.1 maximum on turn 3's draw, so it is dealt to the theme
	// seat separately and stays in hand for two turns.
	mindControlCard := importThemeCard(
		themeDeckRow("Mind Control", mindControlOracle, auraTypeLine, "{3}{U}{U}"), me.ID)
	mindControl := mindControlCard.InstanceID
	me.Hand.PushTop(mindControlCard)

	oppHand := seedThemeSeat(t, g, opponent,
		[]themeLandRow{{"Plains", "Basic Land — Plains", "W", 5}},
		[]cards.Card{
			themeDeckRow("Pacifism", attachThemePacifismOracle, auraTypeLine, "{1}{W}"),
			themeDeckRow("Bonesplitter", bonesplitterOracle, equipTypeLine, "{1}"),
		})

	// --- turn 1: the pile goes together ------------------------------
	advanceToMainOf(t, g, themeSeat)
	firstRound := g.Turn.Round

	// The opponent's body. Placed, not cast: see the file comment.
	theirs := seedThemeCreature(g, opponent.ID, "Pikeman", "Creature — Human Soldier", 3, 3)

	castThemeSpell(t, g, me, hand["Raging Bear"])
	hero := hand["Raging Bear"]
	if !onBattlefield(g, hero) {
		t.Fatal("the hero never reached the battlefield")
	}
	// Cast this turn, so CR 302.6 is in force until an Equipment says
	// otherwise.
	if !summoningSickOf(t, g, hero) {
		t.Fatal("setup: a creature cast this turn should be summoning sick")
	}

	castThemeSpell(t, g, me, hand["Swiftfoot Boots"])
	boots := hand["Swiftfoot Boots"]
	equipTo(t, g, me.ID, boots, hero)
	if host := attachmentHostOf(t, g, boots); host.Kind != game.TargetCard || host.ID != hero {
		t.Fatalf("the Boots' AttachedTo = %+v, want card %s", host, hero)
	}
	for _, want := range []string{"hexproof", "haste"} {
		if ab := effectiveAbilities(t, g, hero); !containsString(ab, want) {
			t.Errorf("the equipped hero's abilities %v missing %q", ab, want)
		}
	}

	// Rancor: an Aura cast with Enchant creature as an announce-time
	// target (CR 303.4a), which enters the battlefield already
	// attached to what it targeted — as it RESOLVES, not at announce.
	rancor := hand["Rancor"]
	if err := g.CastSpell(me.ID, rancor, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: hero}},
		Strict:  true, AutoTap: true,
	}); err != nil {
		t.Fatalf("cast Rancor: %v", err)
	}
	if g.Battlefield.Contains(rancor) {
		t.Fatal("the Aura reached the battlefield at announce; it is a spell on the stack until it resolves")
	}
	if got := effectivePower(t, g, hero); got != 2 {
		t.Errorf("the hero is %d power with Rancor still on the stack, want 2 — "+
			"the attachment is stamped on resolution", got)
	}
	passPriorityAroundTable(t, g)
	if host := attachmentHostOf(t, g, rancor); host.Kind != game.TargetCard || host.ID != hero {
		t.Fatalf("Rancor's AttachedTo = %+v, want card %s", host, hero)
	}
	if p, tough := effectivePower(t, g, hero), effectiveToughness(t, g, hero); p != 4 || tough != 2 {
		t.Errorf("the enchanted hero is %d/%d, want 4/2 (Rancor is +2/+0)", p, tough)
	}
	if ab := effectiveAbilities(t, g, hero); !containsString(ab, "trample") {
		t.Errorf("the enchanted hero's abilities %v missing trample", ab)
	}

	castThemeSpell(t, g, me, hand["Scout Squire"])
	squire := hand["Scout Squire"]

	// The Boots' haste is the reason this attack is legal at all: the
	// hero was cast this turn and CR 302.6 said no a moment ago.
	advanceToStepOf(t, g, themeSeat, game.StepDeclareAttackers)
	oppLife := opponent.Life
	if err := g.DeclareAttacker(hero, opponent.ID); err != nil {
		t.Fatalf("a hero wearing haste boots could not attack the turn it was cast: %v", err)
	}
	lockInAttacks(t, g)
	advanceToStepOf(t, g, themeSeat, game.StepCombatDamage)
	if got := oppLife - opponent.Life; got != 4 {
		t.Errorf("combat damage from the equipped, enchanted hero = %d, want 4", got)
	}

	// --- the opponent's turn: their Aura and their Equipment ---------
	advanceToMainOf(t, g, oppSeat)

	// Hexproof is the half of the Boots that is asymmetric, and this
	// is where it earns the extra mana over the Greaves: an opponent
	// cannot target the hero at all (CR 702.11b).
	pacifism := oppHand["Pacifism"]
	if err := g.CastSpell(opponent.ID, pacifism, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: hero}},
		Strict:  true, AutoTap: true,
	}); err == nil {
		t.Error("an opponent's Aura targeted a hexproof creature (CR 702.11b)")
	}
	if !opponent.Hand.Contains(pacifism) {
		t.Fatal("the refused Aura left the opponent's hand")
	}

	// The Squire is not wearing anything, so it can be enchanted.
	castThemeSpellTargeting(t, g, opponent, pacifism,
		game.TargetRef{Kind: game.TargetCard, ID: squire})
	if host := attachmentHostOf(t, g, pacifism); host.Kind != game.TargetCard || host.ID != squire {
		t.Fatalf("Pacifism's AttachedTo = %+v, want card %s", host, squire)
	}

	// Their own Equipment, on their own creature, so that turn 3 can
	// ask what a control change does to it.
	castThemeSpell(t, g, opponent, oppHand["Bonesplitter"])
	bonesplitter := oppHand["Bonesplitter"]
	equipTo(t, g, opponent.ID, bonesplitter, theirs)
	if got := effectivePower(t, g, theirs); got != 5 {
		t.Fatalf("the opponent's equipped creature is %d power, want 5", got)
	}

	// --- turn 2: the sorcery-speed gate, and undo/replay -------------
	advanceToStepOf(t, g, themeSeat, game.StepUpkeep)
	// CR 702.6a. The Boots are already attached; re-equipping them to
	// the Squire would be a legal MOVE in a main phase, and is not one
	// here.
	if err := tryEquip(t, g, me.ID, boots, squire); err != game.ErrSorcerySpeedRequired {
		t.Errorf("equip in an upkeep returned %v, want ErrSorcerySpeedRequired (CR 702.6a)", err)
	}
	if host := attachmentHostOf(t, g, boots); host.ID != hero {
		t.Errorf("the refused equip moved the Boots to %+v", host)
	}

	advanceToMainOf(t, g, themeSeat)
	if g.Turn.Round != firstRound+1 {
		t.Fatalf("round at the second main phase = %d, want %d", g.Turn.Round, firstRound+1)
	}

	castThemeSpell(t, g, me, hand["Darksteel Plate"])
	plate := hand["Darksteel Plate"]

	// Undo and replay, on the road the room's undo stack actually
	// uses: Game.Clone before the action, Game.RestoreFrom after. An
	// attach is the interesting case for it, because the state it
	// writes is a relation between two cards plus a CR 613.7e
	// timestamp — a shallow copy would leave the restored game with
	// the attachment still stamped on one side of it.
	beforeEquip := g.Clone()
	equipTo(t, g, me.ID, plate, hero)
	if host := attachmentHostOf(t, g, plate); host.Kind != game.TargetCard || host.ID != hero {
		t.Fatalf("the Plate's AttachedTo = %+v, want card %s", host, hero)
	}
	if ab := effectiveAbilities(t, g, hero); !containsString(ab, "indestructible") {
		t.Errorf("the plated hero's abilities %v missing indestructible", ab)
	}
	g.WithWriteLock(func() { g.RestoreFrom(beforeEquip) })
	// RestoreFrom swaps Game.Seats wholesale, so the *game.Player
	// handles taken before the undo now point at objects no longer in
	// the game. The room layer re-reads seats off the *Game after every
	// restore; a test that holds one has to do the same.
	me, opponent = g.Seats[themeSeat], g.Seats[oppSeat]
	if !g.Battlefield.Contains(plate) {
		t.Fatal("the undo took the Equipment off the battlefield as well")
	}
	if host := attachmentHostOf(t, g, plate); host.Kind != "" {
		t.Errorf("the Plate is still attached to %+v after the undo", host)
	}
	if ab := effectiveAbilities(t, g, hero); containsString(ab, "indestructible") {
		t.Errorf("the hero kept the undone Equipment's grant: %v", ab)
	}
	// Replay: the same activation, out of the restored state, and it
	// lands the same way — including paying the cost again, since the
	// restore untapped the lands it had tapped.
	equipTo(t, g, me.ID, plate, hero)
	if host := attachmentHostOf(t, g, plate); host.Kind != game.TargetCard || host.ID != hero {
		t.Fatalf("the replayed equip left the Plate at %+v", host)
	}
	if ab := effectiveAbilities(t, g, hero); !containsString(ab, "indestructible") {
		t.Errorf("the replayed equip did not restore the grant: %v", ab)
	}

	// Lightning Greaves on the Squire. Shroud is the half that is
	// symmetric, and it locks the Squire away from its OWN controller
	// (CR 702.18a) — so the Skullclamp that wants to go there next
	// cannot target it.
	castThemeSpell(t, g, me, hand["Lightning Greaves"])
	greaves := hand["Lightning Greaves"]
	equipTo(t, g, me.ID, greaves, squire)
	if ab := effectiveAbilities(t, g, squire); !containsString(ab, "shroud") {
		t.Errorf("the Squire's abilities %v missing shroud", ab)
	}

	castThemeSpell(t, g, me, hand["Skullclamp"])
	clamp := hand["Skullclamp"]
	if err := tryEquip(t, g, me.ID, clamp, squire); err == nil {
		t.Error("an equip targeted a creature with shroud (CR 702.18a)")
	}
	if host := attachmentHostOf(t, g, clamp); host.Kind != "" {
		t.Errorf("the refused equip attached the Clamp to %+v", host)
	}

	// CR 701.3a: a second activation MOVES the Equipment, and the
	// grant moves with it in the same beat.
	equipTo(t, g, me.ID, greaves, hero)
	if ab := effectiveAbilities(t, g, squire); containsString(ab, "shroud") {
		t.Errorf("the Squire kept shroud after the Greaves moved: %v", ab)
	}
	if ab := effectiveAbilities(t, g, hero); !containsString(ab, "shroud") {
		t.Errorf("the hero did not pick up shroud when the Greaves arrived: %v", ab)
	}

	// --- turn 3: Mind Control, then the two SBAs ---------------------
	advanceToStepOf(t, g, themeSeat, game.StepUpkeep)
	advanceToMainOf(t, g, themeSeat)
	if g.Turn.Round != firstRound+2 {
		t.Fatalf("round at the third main phase = %d, want %d", g.Turn.Round, firstRound+2)
	}

	castThemeSpellTargeting(t, g, me, mindControl,
		game.TargetRef{Kind: game.TargetCard, ID: theirs})
	if host := attachmentHostOf(t, g, mindControl); host.Kind != game.TargetCard || host.ID != theirs {
		t.Fatalf("Mind Control's AttachedTo = %+v, want card %s", host, theirs)
	}
	if got := controllerOf(t, g, theirs); got != me.ID {
		t.Errorf("the stolen creature's controller = %s, want %s (CR 613.1b)", got, me.ID)
	}
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == theirs && c.Owner != opponent.ID {
				t.Errorf("owner changed to %s; a control effect never touches ownership (CR 108.3)", c.Owner)
			}
		}
	})
	if !summoningSickOf(t, g, theirs) {
		t.Error("a stolen creature is summoning sick for its new controller (CR 302.6)")
	}
	// The interesting half for S24: the Equipment the OPPONENT put on
	// that creature did not move, did not change controller, and did
	// not stop working. An Equipment's controller is its own; the
	// creature it is attached to has nothing to do with it.
	if host := attachmentHostOf(t, g, bonesplitter); host.Kind != game.TargetCard || host.ID != theirs {
		t.Errorf("the Bonesplitter came off when its host changed control: %+v", host)
	}
	if got := controllerOf(t, g, bonesplitter); got != opponent.ID {
		t.Errorf("the Bonesplitter's controller = %s, want %s — stealing the creature "+
			"does not steal what is attached to it", got, opponent.ID)
	}
	if got := effectivePower(t, g, theirs); got != 5 {
		t.Errorf("the stolen creature is %d power, want 5 — an 'equipped creature gets' "+
			"static reaches its host whoever controls the host", got)
	}

	// Skullclamp the Squire. 1/1, +1/-1, dead on the toughness SBA
	// (CR 704.5f) — and its death is what fires the two attachment
	// state-based actions this sprint exists for.
	handBefore := me.Hand.Size()
	equipTo(t, g, me.ID, clamp, squire)
	if g.Battlefield.Contains(squire) {
		t.Fatal("a clamped 1/1 is a 2/0 and dies to the toughness SBA (CR 704.5f)")
	}
	// CR 704.5m: the Aura is put into its OWNER's graveyard — the
	// opponent's, not the graveyard of the player whose creature it
	// was enchanting.
	if g.Battlefield.Contains(pacifism) {
		t.Error("CR 704.5m: the Aura stayed on the battlefield with nothing to enchant")
	}
	if !opponent.Graveyard.Contains(pacifism) {
		t.Error("CR 704.5m / CR 400.3: the fallen Aura is not in its OWNER's graveyard")
	}
	if me.Graveyard.Contains(pacifism) {
		t.Error("the fallen Aura went to the graveyard of the enchanted creature's controller")
	}
	// CR 704.5n: the Equipment is NOT put anywhere. It becomes
	// unattached and waits for the next creature.
	if !g.Battlefield.Contains(clamp) {
		t.Fatal("CR 704.5n: the Equipment left the battlefield with its creature")
	}
	if host := attachmentHostOf(t, g, clamp); host.Kind != "" {
		t.Errorf("CR 704.5n: the Equipment is still attached to the dead creature: %+v", host)
	}
	if got := me.Hand.Size(); got != handBefore+2 {
		t.Errorf("hand %d -> %d, want +2 from Skullclamp's dies trigger", handBefore, got)
	}

	// And the control change unwinds by itself, because nothing
	// remembered the old controller: the layer-2 effect simply stops
	// being in the active set. The Equipment is still where it was.
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(mindControl); err != nil {
			t.Fatalf("destroy the Aura: %v", err)
		}
	})
	if got := controllerOf(t, g, theirs); got != opponent.ID {
		t.Errorf("control did not revert when the Aura left: %s, want %s", got, opponent.ID)
	}
	if host := attachmentHostOf(t, g, bonesplitter); host.Kind != game.TargetCard || host.ID != theirs {
		t.Errorf("the Bonesplitter moved when control reverted: %+v", host)
	}

	if n := countCatalogEvents(g, game.EventEffectError, runStart); n != 0 {
		t.Errorf("EventEffectError count across the three turns = %d, want 0", n)
	}
}
