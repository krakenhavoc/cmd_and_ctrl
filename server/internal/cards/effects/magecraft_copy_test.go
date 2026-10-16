package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// magecraft_copy_test.go — #2090: magecraft is "whenever you cast OR
// COPY an instant or sorcery spell" (CR 207.2c). A copy put on the stack
// is created, not cast (CR 707.10), and triggers it once; a copy cast
// from another zone is cast (CR 707.12) and triggers it once, as a cast;
// a copy someone else makes is theirs (CR 707.10b) and triggers theirs.
//
// Archmage Emeritus draws a card per trigger, so its controller's hand
// size counts the triggers.

const extusTestOracle = "0b299983-9f0f-404a-acd1-8f142572b1f1"

// mcArchmage puts an Archmage Emeritus onto the battlefield under p and
// returns a counter of the cards p has drawn since.
func mcArchmage(g *game.Game, p *game.Player) func() int {
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Archmage Emeritus", TypeLine: "Creature — Human Wizard",
		OracleID: cavCArchmageOracle, Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID,
	})
	start := p.Hand.Size()
	return func() int { return p.Hand.Size() - start }
}

// mcSettle passes priority until the stack is empty, answering every
// prompt on the way: a CR 603.3b trigger order as offered, and a target
// prompt with `player` when it is offered (a copy's re-target prompt
// offers the original's target among the rest), else a card it has not
// picked before (two magecraft triggers on the stack at once each get
// their own card), else the first player.
func mcSettle(t *testing.T, g *game.Game, player uuid.UUID) {
	t.Helper()
	picked := map[uuid.UUID]bool{}
	for i := 0; i < 128; i++ {
		var open *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil {
				open = c
				break
			}
		}
		if open != nil {
			switch open.Kind {
			case game.PendingChoiceTriggerOrder:
				if err := g.ResolveTriggerOrder(open.ID, open.Chooser, append([]uuid.UUID(nil), open.TriggerOrderIDs...)); err != nil {
					t.Fatalf("ResolveTriggerOrder: %v", err)
				}
			case game.PendingChoicePickTarget:
				ref := game.TargetRef{Kind: game.TargetPlayer}
				switch {
				case player != uuid.Nil && hasID(open.PickTargetPlayers, player):
					ref.ID = player
				case len(open.PickTargetCards) > 0:
					ref = game.TargetRef{Kind: game.TargetCard, ID: open.PickTargetCards[0]}
					for _, id := range open.PickTargetCards {
						if !picked[id] {
							ref.ID = id
							break
						}
					}
					picked[ref.ID] = true
				case len(open.PickTargetPlayers) > 0:
					ref.ID = open.PickTargetPlayers[0]
				default:
					t.Fatalf("a target prompt with nothing to pick: %+v", open)
				}
				if err := g.ResolvePickTarget(open.ID, open.Chooser, ref); err != nil {
					t.Fatalf("ResolvePickTarget: %v", err)
				}
			default:
				t.Fatalf("unexpected %s prompt", open.Kind)
			}
			continue
		}
		if stackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never emptied")
}

// A cast triggers magecraft, and so does a copy made by a spell you cast:
// the Bolt (cast), Reverberate (cast) and Reverberate's copy (copied).
func TestMagecraftTriggersOnACastAndOnACopy(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	drawn := mcArchmage(g, me)

	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
	castCatalogSpell(t, g, "Reverberate", "Instant", reverberateOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	mcSettle(t, g, victim.ID)
	if got := drawn(); got != 3 {
		t.Errorf("drew %d, want 3 (the Bolt cast, Reverberate cast, the copy)", got)
	}
}

// Storm copies are copies, one trigger each: two fillers, Grapeshot and
// its two copies are five triggers. The storm trigger itself is an
// ability and triggers nothing.
func TestMagecraftTriggersOncePerStormCopy(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drawn := mcArchmage(g, me)

	stormFiller(t, g)
	stormFiller(t, g)
	mcSettle(t, g, opp.ID)
	if got := drawn(); got != 2 {
		t.Fatalf("drew %d after two cast instants, want 2", got)
	}
	before := lifeOf(g, opp.ID)
	castCatalogSpell(t, g, "Grapeshot", "Sorcery", grapeshotOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	mcSettle(t, g, opp.ID)
	if got := lifeOf(g, opp.ID); got != before-3 {
		t.Fatalf("life %d → %d, want -3: Grapeshot and two copies", before, got)
	}
	if got := drawn(); got != 5 {
		t.Errorf("drew %d, want 5 (two fillers, Grapeshot, two storm copies)", got)
	}
}

// Casting a copy of a card is a cast (CR 707.12) and triggers magecraft
// once, as a cast. Making the card copy in exile triggers nothing: the
// rulings say a card copied in another zone triggers magecraft only if
// that copy is cast.
func TestMagecraftCastingACopyTriggersOnce(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	drawn := mcArchmage(g, me)
	bolt := gyCard(me, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)

	castMizzix(t, g, false, []game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	passPriorityAroundTable(t, g)
	if got := drawn(); got != 1 {
		t.Fatalf("drew %d after Mizzix's Mastery resolved, want 1: the copy in exile was not cast", got)
	}
	cp := copyOf(t, g, "Lightning Bolt")
	before := opp.Life
	if err := castCopy(g, me.ID, cp.InstanceID, true, playerTarget(opp.ID)); err != nil {
		t.Fatalf("casting the copy: %v", err)
	}
	mcSettle(t, g, opp.ID)
	if opp.Life != before-3 {
		t.Fatalf("life %d → %d: the cast copy resolves as a Bolt", before, opp.Life)
	}
	if got := drawn(); got != 2 {
		t.Errorf("drew %d, want 2 (Mizzix's Mastery, the cast copy once — not again as a copy)", got)
	}
}

// A copy belongs to the player who made it (CR 707.10b). An opponent's
// Reverberate on your Bolt is their cast and their copy: your magecraft
// sees only your Bolt, and theirs sees Reverberate and the copy.
func TestMagecraftIgnoresAnOpponentsCopyOfYourSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, victim := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := mcArchmage(g, me)
	theirs := mcArchmage(g, opp)

	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
	castFromSeat(t, g, opp, "Reverberate", "Instant", reverberateOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	mcSettle(t, g, victim.ID)
	if got := mine(); got != 1 {
		t.Errorf("I drew %d, want 1 (my Bolt only: the copy is the opponent's)", got)
	}
	if got := theirs(); got != 2 {
		t.Errorf("the opponent drew %d, want 2 (their Reverberate and their copy)", got)
	}
}

// --- Extus, Oriq Overlord // Awaken the Blood Avatar -------------------

// extusOnBattlefield puts Extus, front face up, onto p's battlefield.
func extusOnBattlefield(g *game.Game, p *game.Player) uuid.UUID {
	c := extusCard(p.ID)
	pushBattlefieldCardWithTimestamp(g, c)
	return c.InstanceID
}

func extusCard(owner uuid.UUID) game.Card {
	return mdfcCard(owner, extusTestOracle, game.LayoutModalDFC,
		game.Face{Name: "Extus, Oriq Overlord", TypeLine: "Legendary Creature — Human Warlock", ManaCost: "{1}{W}{B}{B}",
			Power: 2, Toughness: 4, Colors: []string{"W", "B"}},
		game.Face{Name: "Awaken the Blood Avatar", TypeLine: "Sorcery", ManaCost: "{6}{B}{R}", Colors: []string{"B", "R"}},
	)
}

// Magecraft on a cast and on a copy: Reverberate on a Bolt is two casts
// and a copy, and each returns a nonlegendary creature card. The
// legendary one in the graveyard is never offered.
func TestExtusReturnsACreatureForEachCastAndCopy(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	extusOnBattlefield(g, me)
	var bears []uuid.UUID
	for i := 0; i < 3; i++ {
		bears = append(bears, gyCard(me, "Grizzly Bears", "Creature — Bear", "{1}{G}", ""))
	}
	legend := gyCard(me, "Old Legend", "Legendary Creature — Human", "{2}{W}", "")
	gyCard(me, "Some Sorcery", "Sorcery", "{1}", "")

	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
	castCatalogSpell(t, g, "Reverberate", "Instant", reverberateOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	mcSettle(t, g, victim.ID)

	for _, id := range bears {
		if !me.Hand.Contains(id) {
			t.Errorf("a Grizzly Bears is still in the graveyard: three magecraft triggers return three")
		}
	}
	if me.Hand.Contains(legend) || !me.Graveyard.Contains(legend) {
		t.Error("a legendary creature card is not a legal target")
	}
}

// The target is a nonlegendary creature card in YOUR graveyard: with
// only an opponent's creature and your own legend, nothing is offered.
func TestExtusTargetsOnlyYourNonlegendaryCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	extusOnBattlefield(g, me)
	theirs := gyCard(opp, "Grizzly Bears", "Creature — Bear", "{1}{G}", "")
	legend := gyCard(me, "Old Legend", "Legendary Creature — Human", "{2}{W}", "")

	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	mcSettle(t, g, opp.ID)
	if !opp.Graveyard.Contains(theirs) || !me.Graveyard.Contains(legend) {
		t.Error("magecraft returned a card it may not target")
	}
}

// Awaken the Blood Avatar: {2} less per creature sacrificed, generic only;
// each opponent sacrifices a creature; a 3/6 black and red Avatar with
// haste whose attack deals 3 damage to each opponent.
func TestAwakenTheBloodAvatar(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opps := g.Seats[1:]
	fodder := vsCreatures(g, me.ID, 4, "Creature — Goblin")
	card := extusCard(me.ID)
	card.SetFace(1)
	me.Hand.PushTop(card)
	for n, want := range []int{8, 6, 4, 2, 2} {
		price, err := g.PriceCast(me.ID, card, game.CastSpellParams{Face: 1, SacrificeIDs: fodder[:n]})
		if err != nil {
			t.Fatalf("PriceCast: %v", err)
		}
		if got := price.Total.ManaValue(); got != want {
			t.Errorf("%d sacrificed: total %s, want mana value %d", n, price.Total.String(), want)
		}
	}

	victim := vsPermanent(g, opps[0].ID, "Victim", "Creature — Human")
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Strict: true, SacrificeIDs: fodder[:3]}); err != nil {
		t.Fatalf("cast Awaken the Blood Avatar with three sacrificed: %v", err)
	}
	if n := len(me.ManaPool); n != 0 {
		t.Fatalf("%d mana left floating: three sacrifices make it {B}{R}", n)
	}
	mcSettle(t, g, uuid.Nil)

	if g.Battlefield.Contains(victim) {
		t.Error("each opponent sacrifices a creature")
	}
	avatars := battlefieldIDsNamed(g, "Avatar")
	if len(avatars) != 1 {
		t.Fatalf("%d Avatars, want 1", len(avatars))
	}
	avatar := cardByID(g, avatars[0])
	if avatar.Controller != me.ID || avatar.Power != 3 || avatar.Toughness != 6 || !avatar.IsToken() {
		t.Errorf("Avatar = %+v, want my 3/6 token", avatar)
	}
	if !hasEffectiveKeyword(t, g, avatars[0], "haste") {
		t.Error("the Avatar has haste")
	}

	before := make([]int, len(opps))
	for i, p := range opps {
		before[i] = p.Life
	}
	declareAttack(t, g, opps[1].ID, avatars[0])
	mcSettle(t, g, uuid.Nil)
	for i, p := range opps {
		if p.Life != before[i]-3 {
			t.Errorf("opponent %d at %d, want %d: the attack trigger deals 3 to each opponent", i, p.Life, before[i]-3)
		}
	}
}
