package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestSiegeDefeatedGrantsItsBackFace is the seam's headline: a real
// Siege, defeated, is exiled carrying a free cast of face 1 and
// nothing else.
func TestSiegeDefeatedGrantsItsBackFace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	protector := g.Seats[(seat+1)%len(g.Seats)]

	row := effectstest.SiegeRow(effects.InvasionOfKarsusOracleID, "Invasion of Karsus", "Refraction Elemental",
		"{2}{R}{R}", "4", "Creature — Elemental", "4", "4")
	id := effectstest.ImportToHand(row, owner)

	// The import road is the thing under test as much as the seam is:
	// if card_faces[1] does not reach game.Card.Faces, the grant has
	// nothing to point at.
	inHand, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatal("the imported card is nowhere")
	}
	if inHand.FaceCount() != 2 {
		t.Fatalf("imported %d faces, want 2 — deck.ToGameCard dropped card_faces", inHand.FaceCount())
	}
	if inHand.Faces[1].Name != "Refraction Elemental" {
		t.Fatalf("back face imported as %q", inHand.Faces[1].Name)
	}
	if inHand.StartingDefense != 4 {
		t.Fatalf("front face defense imported as %d, want 4", inHand.StartingDefense)
	}

	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneHand, Owner: owner.ID},
		game.ZoneRef{Kind: game.ZoneBattlefield}, id,
	); err != nil {
		t.Fatalf("move battle to battlefield: %v", err)
	}
	effectstest.AnswerProtectorPrompt(t, g, owner.ID, protector.ID)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.DefeatBattle(t, g, id, 4)

	if !effectstest.InExile(g, id) {
		t.Fatal("the defeated Siege was not exiled")
	}
	grant := effectstest.ExileGrantFor(g, id)
	if grant.Player != owner.ID {
		t.Errorf("grant holder = %v, want the battle's controller %v", grant.Player, owner.ID)
	}
	if face, ok := grant.NamedFace(); !ok || face != 1 {
		t.Errorf("grant faces = %v, want just the back face 1", grant.Faces)
	}
	if grant.Cost != "{0}" {
		t.Errorf("grant cost = %q, want %q — the transformed cast is free", grant.Cost, "{0}")
	}
	if !grant.CastOnly {
		t.Error("the grant should be cast-only")
	}
	if grant.Duration.Kind != game.UntilEndOfTurn {
		t.Errorf("grant duration = %v, want until end of turn", grant.Duration.Kind)
	}
}

// TestInvasionOfKarsusSweepsThenBecomesTheElemental walks the whole
// card: the ETB sweep, the defeat, the free transformed cast, and the
// BACK FACE's own trigger firing off a later spell — which is the
// part that proves the "#1" catalog key is reached, not just that a
// creature landed.
func TestInvasionOfKarsusSweepsThenBecomesTheElemental(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	protector := g.Seats[(seat+1)%len(g.Seats)]

	// Two bystanders: one dies to the 3-damage sweep, one survives.
	small := effectstest.PushCreatureToBattlefieldForTest(g, protector.ID, "Doomed Bear")
	big := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: big,
		Name:       "Big Bear",
		TypeLine:   "Creature — Test",
		Power:      5,
		Toughness:  5,
		Owner:      protector.ID,
		Controller: protector.ID,
	})

	row := effectstest.SiegeRow(effects.InvasionOfKarsusOracleID, "Invasion of Karsus", "Refraction Elemental",
		"{2}{R}{R}", "4", "Creature — Elemental", "4", "4")
	id := effectstest.ImportToHand(row, owner)
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneHand, Owner: owner.ID},
		game.ZoneRef{Kind: game.ZoneBattlefield}, id,
	); err != nil {
		t.Fatalf("move battle to battlefield: %v", err)
	}
	effectstest.AnswerProtectorPrompt(t, g, owner.ID, protector.ID)
	effectstest.PassPriorityAroundTable(t, g)

	if effects.OnBattlefield(g, small) {
		t.Error("the 2/2 survived a 3-damage sweep")
	}
	if !effects.OnBattlefield(g, big) {
		t.Error("the 5/5 died to a 3-damage sweep")
	}
	// The Siege is not a creature or a planeswalker, so it must not
	// have swept itself: four defense counters, minus nothing.
	if c := effectstest.BattlefieldCardFor(g, id); c == nil {
		t.Fatal("the Siege swept itself off the battlefield")
	} else if got := c.Counters[game.CounterDefense]; got != 4 {
		t.Errorf("defense after its own sweep = %d, want 4", got)
	}

	effectstest.DefeatBattle(t, g, id, 4)

	// Cast the back face out of exile. No mana in the pool and the
	// front face costs {2}{R}{R}: only the {0} override can pay for
	// this, and only the grant's face can name it.
	effectstest.ToMainPhaseSameTurn(t, g)
	if err := g.CastSpell(owner.ID, id, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast the transformed back face: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	landed := effectstest.BattlefieldCardFor(g, id)
	if landed == nil {
		t.Fatal("Refraction Elemental never reached the battlefield")
	}
	if landed.Name != "Refraction Elemental" || !landed.IsCreature() {
		t.Fatalf("permanent is %q (%q), want the 4/4 Elemental", landed.Name, landed.TypeLine)
	}
	if landed.IsBattle() {
		t.Fatal("the defeated battle re-entered the battlefield as a battle")
	}

	// The back face's own trigger: "whenever you cast a spell, this
	// creature deals 2 damage to each opponent". Reached only through
	// the "#1" catalog key.
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}
	// Any spell at all — "whenever you cast a spell" has no type
	// filter, so an uncatalogued instant is the cheapest probe and
	// proves the trigger is not keyed on the spell's own spec.
	effectstest.CastCatalogSpell(t, g, "Filler Instant", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)

	for _, p := range g.Seats {
		if p.ID == owner.ID {
			continue
		}
		if got := before[p.ID] - p.Life; got != 2 {
			t.Errorf("opponent %s lost %d life, want the Elemental's 2", p.Name, got)
		}
	}
	if g.Seats[seat].Life != before[owner.ID] {
		t.Error("the Elemental's controller lost life; the trigger hits opponents only")
	}
}

// TestInvasionOfInnistradBecomesDelugeOfTheDead is the non-creature
// back face, and the card S27 shipped with its transformed cast
// declared missing. An enchantment proves the face model is not
// creature-shaped; Deluge's own ETB trigger firing proves the "#1"
// key is what the engine looked the landed permanent up under.
func TestInvasionOfInnistradBecomesDelugeOfTheDead(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	protector := g.Seats[(seat+1)%len(g.Seats)]
	victim := effectstest.PushCreatureToBattlefieldForTest(g, protector.ID, "Doomed Bear")

	row := effectstest.SiegeRow(effects.InvasionOfInnistradOracleID, "Invasion of Innistrad", "Deluge of the Dead",
		"{2}{B}{B}", "5", "Enchantment", "", "")
	id := effectstest.ImportToHand(row, owner)
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneHand, Owner: owner.ID},
		game.ZoneRef{Kind: game.ZoneBattlefield}, id,
	); err != nil {
		t.Fatalf("move battle to battlefield: %v", err)
	}
	effectstest.AnswerProtectorPrompt(t, g, owner.ID, protector.ID)
	effectstest.AnswerFirstPickTarget(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, victim) {
		t.Error("the -13/-13 creature survived")
	}

	effectstest.DefeatBattle(t, g, id, 5)
	effectstest.ToMainPhaseSameTurn(t, g)
	if err := g.CastSpell(owner.ID, id, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast Deluge of the Dead: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	landed := effectstest.BattlefieldCardFor(g, id)
	if landed == nil {
		t.Fatal("Deluge of the Dead never reached the battlefield")
	}
	if landed.Name != "Deluge of the Dead" || !landed.IsEnchantment() {
		t.Fatalf("permanent is %q (%q), want the enchantment", landed.Name, landed.TypeLine)
	}
	if n := effectstest.CountBattlefieldByName(g, "Zombie"); n != 2 {
		t.Errorf("Zombie tokens = %d, want the 2 from Deluge's ETB trigger", n)
	}
}

// TestInvasionOfTarkirRevealsDragonsThenDamages walks the printed
// order (#636): only the Dragon cards in hand are offered, the
// reflexive "when you do" then goes on the stack and picks its target
// THERE (CR 603.3d) rather than when the entry trigger was announced,
// and it deals the number revealed plus two.
func TestInvasionOfTarkirRevealsDragonsThenDamages(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	protector := g.Seats[(seat+1)%len(g.Seats)]

	for _, n := range []string{"Shivan Dragon", "Bogardan Hellkite"} {
		owner.Hand.PushTop(game.Card{
			InstanceID: uuid.New(), Name: n, TypeLine: "Creature — Dragon",
			Owner: owner.ID, Controller: owner.ID,
		})
	}
	// A non-Dragon, to prove the candidate scan filters.
	owner.Hand.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		Owner: owner.ID, Controller: owner.ID,
	})

	row := effectstest.SiegeRow(effects.InvasionOfTarkirOracleID, "Invasion of Tarkir", "Defiant Thundermaw",
		"{1}{R}", "5", "Creature — Dragon", "4", "4")
	id := effectstest.ImportToHand(row, owner)
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneHand, Owner: owner.ID},
		game.ZoneRef{Kind: game.ZoneBattlefield}, id,
	); err != nil {
		t.Fatalf("move battle to battlefield: %v", err)
	}
	effectstest.AnswerProtectorPrompt(t, g, owner.ID, protector.ID)
	before := protector.Life
	// The entry trigger has no target clause of its own, so it
	// resolves on its own and asks what to reveal.
	effectstest.PassPriorityAroundTable(t, g)

	choice := effectstest.PendingChooseCards(t, g, owner.ID)
	if len(choice.ChooseCards) != 2 {
		t.Fatalf("offered %d cards to reveal, want the 2 Dragons", len(choice.ChooseCards))
	}
	if choice.ChooseMin != 0 {
		t.Errorf(`reveal floor = %d, want 0 — "any number" includes none`, choice.ChooseMin)
	}
	if err := g.ResolveChooseCards(choice.ID, owner.ID, choice.ChooseCards); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	// CR 603.3d: the reflexive trigger's target is chosen NOW, after
	// the reveal, knowing X.
	effectstest.AnswerPickTargetPlayer(t, g, protector.ID)
	// And it is its own stack item: the damage has not happened yet,
	// so the table has a window to respond to it.
	if effectstest.TriggerOnStack(g, id) == nil {
		t.Error(`"when you do" should be a second trigger on the stack`)
	}
	if protector.Life != before {
		t.Errorf("damage landed inside the entry trigger's resolution: life %d → %d", before, protector.Life)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := before - protector.Life; got != 4 {
		t.Errorf("damage = %d, want 2 revealed plus 2", got)
	}
}

// TestInvasionOfTarkirRevealingNothingStillDamages is the "(X can be
// 0.)" parenthesis: an empty reveal still satisfies "when you do", so
// the reflexive trigger happens with X = 0 and the damage floor is 2.
func TestInvasionOfTarkirRevealingNothingStillDamages(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	protector := g.Seats[(seat+1)%len(g.Seats)]

	row := effectstest.SiegeRow(effects.InvasionOfTarkirOracleID, "Invasion of Tarkir", "Defiant Thundermaw",
		"{1}{R}", "5", "Creature — Dragon", "4", "4")
	id := effectstest.ImportToHand(row, owner)
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneHand, Owner: owner.ID},
		game.ZoneRef{Kind: game.ZoneBattlefield}, id,
	); err != nil {
		t.Fatalf("move battle to battlefield: %v", err)
	}
	effectstest.AnswerProtectorPrompt(t, g, owner.ID, protector.ID)
	before := protector.Life
	// No Dragons in hand: no reveal prompt at all, straight to the
	// reflexive trigger's target.
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetPlayer(t, g, protector.ID)
	effectstest.PassPriorityAroundTable(t, g)

	if got := before - protector.Life; got != 2 {
		t.Errorf("damage = %d, want the X=0 floor of 2", got)
	}
}

// TestDefiantThundermawDragonAttackTrigger is the back face on its
// own: an attack trigger keyed on a subtype, with the ATTACKING
// Dragon as the damage source rather than the Thundermaw.
func TestDefiantThundermawDragonAttackTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, victimSeat := g.Seats[0], g.Seats[1]

	maw := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: maw,
		Name:       "Defiant Thundermaw",
		OracleID:   effects.InvasionOfTarkirOracleID,
		ActiveFace: 1,
		Layout:     game.LayoutTransform,
		TypeLine:   "Creature — Dragon",
		Power:      4,
		Toughness:  4,
		Owner:      owner.ID,
		Controller: owner.ID,
		Faces: []game.Face{
			{Name: "Invasion of Tarkir", TypeLine: "Battle — Siege", ManaCost: "{1}{R}", StartingDefense: 5},
			{Name: "Defiant Thundermaw", TypeLine: "Creature — Dragon", Power: 4, Toughness: 4},
		},
	})
	// If ActiveFace and the catalog key ever disagree the spec is
	// simply never found and the rest of this test goes quiet rather
	// than red, so assert the key first.
	c := effectstest.BattlefieldCardFor(g, maw)
	if c == nil || game.CatalogKey(*c) != effects.InvasionOfTarkirOracleID+"#1" {
		t.Fatal("the battlefield card does not key on the back face")
	}

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	before := victimSeat.Life
	if err := g.DeclareAttacker(maw, victimSeat.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	// #859: the declaration is announced at its lock-in — the
	// priority wrap inside declare_attackers — so the attack triggers
	// exist only after this.
	effectstest.LockInAttacks(t, g)
	effectstest.AnswerPickTargetPlayer(t, g, victimSeat.ID)
	effectstest.PassPriorityAroundTable(t, g)

	if got := before - victimSeat.Life; got != 2 {
		t.Errorf("attack-trigger damage = %d, want 2", got)
	}
}
