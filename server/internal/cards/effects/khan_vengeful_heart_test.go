package effects

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

const khanVengefulHeartOracle = "d2a86d0c-a4fe-434c-a45c-812cdec9169f"

// khanCard is a printed Khan, Vengeful Heart owned by owner.
func khanCard(owner uuid.UUID) game.Card {
	return game.Card{
		InstanceID: uuid.New(), Name: "Khan, Vengeful Heart", TypeLine: "Legendary Creature — Human Warrior",
		OracleID: khanVengefulHeartOracle, ManaCost: "{2}{B}{B}", Colors: []string{"B"},
		Power: 5, Toughness: 3, Owner: owner, Controller: owner,
	}
}

// khanExiledBySwords puts Khan onto the battlefield and exiles it with a
// real Swords to Plowshares: an effect that is not Khan's, and grants no
// permission of its own. The table is left in its main phase with an empty
// stack.
func khanExiledBySwords(t *testing.T, g *game.Game, owner uuid.UUID) uuid.UUID {
	t.Helper()
	khan := pushBattlefieldCardWithTimestamp(g, khanCard(owner))
	castCatalogSpell(t, g, "Swords to Plowshares", "Instant", "b1544f21-7e98-461b-aed5-e748b0168c52",
		[]game.TargetRef{{Kind: game.TargetCard, ID: khan}})
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(khan) {
		t.Fatal("Swords to Plowshares did not exile Khan")
	}
	return khan
}

func addBlack(g *game.Game, p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.ManaPool.AddMana(game.ManaToken{Color: "B"})
	}
}

func TestKhanSacrificesAnArtifactToDrawTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rock := pushTypedCard(g, me.ID, "Rock", "Artifact", "{1}")
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	land := pushTypedCard(g, me.ID, "Swamp", "Basic Land — Swamp", "")
	khan := castAndResolveCreature(t, g, "Khan, Vengeful Heart", "Legendary Creature — Human Warrior", khanVengefulHeartOracle)
	passPriorityAroundTable(t, g)
	prompt := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if prompt == nil {
		t.Fatal("the sacrifice is asked as the trigger resolves")
	}
	if hasID(prompt.ChooseCards, khan) {
		t.Error("ANOTHER creature or artifact: Khan itself must not be offered")
	}
	if !hasID(prompt.ChooseCards, rock) || !hasID(prompt.ChooseCards, bear) {
		t.Errorf("offered %v, want the artifact and the creature", prompt.ChooseCards)
	}
	if hasID(prompt.ChooseCards, land) {
		t.Error("a land is neither a creature nor an artifact")
	}
	handBefore := me.Hand.Size()
	answerOwnPermanents(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !g.Battlefield.Contains(bear) {
		t.Fatal("exactly the chosen artifact is sacrificed")
	}
	if got := me.Hand.Size(); got != handBefore+2 {
		t.Errorf("hand %d, want %d (drew two)", got, handBefore+2)
	}
}

func TestKhanDecliningDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	castAndResolveCreature(t, g, "Khan, Vengeful Heart", "Legendary Creature — Human Warrior", khanVengefulHeartOracle)
	passPriorityAroundTable(t, g)
	handBefore := me.Hand.Size()
	answerOwnPermanents(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) || me.Hand.Size() != handBefore {
		t.Error("declining sacrifices nothing and draws nothing")
	}
}

// The clause the seam is for: Khan exiled by somebody else's effect is
// castable by its owner from exile, at its own cost, with no permission
// granted by anything — and the cast is a real one, so Khan's enters
// trigger asks again.
func TestKhanCastsItselfFromExileAfterAnotherEffectExiledIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rock := pushTypedCard(g, me.ID, "Rock", "Artifact", "{1}")
	khan := khanExiledBySwords(t, g, me.ID)
	if len(me.CastPermissions) != 0 {
		t.Fatalf("nothing should have stored a permission: %v", me.CastPermissions)
	}

	// Short a mana: refused by the cost, not let in free.
	addBlack(g, me, 3)
	if err := g.CastSpell(me.ID, khan, game.CastSpellParams{Strict: true, FromZone: "exile"}); err == nil {
		t.Fatal("a Khan cast from exile with three mana was accepted: the cast must pay {2}{B}{B}")
	}
	addBlack(g, me, 1)
	if err := g.CastSpell(me.ID, khan, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("cast Khan from exile: %v", err)
	}
	if !g.Stack.Contains(khan) || g.Exile.Contains(khan) {
		t.Fatal("Khan is not on the stack")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(khan) {
		t.Fatal("Khan cast from exile did not resolve onto the battlefield")
	}
	passPriorityAroundTable(t, g)
	if prompt := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID); prompt == nil || !hasID(prompt.ChooseCards, rock) {
		t.Fatal("Khan cast from exile is a new object that entered: its enters trigger asks the sacrifice")
	}
}

// "You" on a card with no controller is its owner (CR 109.5): exile is
// shared, and an opponent may not cast my Khan out of it.
func TestKhanInExileIsNotAnOpponentsToCast(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	khan := khanExiledBySwords(t, g, me.ID)
	addBlack(g, opp, 4)
	if err := g.CastSpell(opp.ID, khan, game.CastSpellParams{Strict: true, FromZone: "exile"}); !errors.Is(err, game.ErrNoPlayPermission) {
		t.Fatalf("opponent's cast of my exiled Khan: %v, want ErrNoPlayPermission", err)
	}
	if !g.Exile.Contains(khan) {
		t.Fatal("Khan left exile")
	}
}

// A card exiled face down has no characteristics (CR 406.3a, 708.2), so
// Khan's own text is not there to open exile.
func TestKhanExiledFaceDownIsNotCastable(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	k := khanCard(me.ID)
	me.Library.PushTop(k)
	g.WithWriteLock(func() {
		if _, err := g.ExileTopFaceDownForEffect(me.ID, 1); err != nil {
			t.Fatalf("ExileTopFaceDownForEffect: %v", err)
		}
	})
	var exiled *game.Card
	for i := range g.Exile.Cards {
		if g.Exile.Cards[i].InstanceID == k.InstanceID {
			exiled = &g.Exile.Cards[i]
		}
	}
	if exiled == nil || !exiled.FaceDown {
		t.Fatal("Khan is not face down in exile")
	}
	addBlack(g, me, 4)
	if err := g.CastSpell(me.ID, k.InstanceID, game.CastSpellParams{Strict: true, FromZone: "exile"}); !errors.Is(err, game.ErrNoPlayPermission) {
		t.Fatalf("cast of a face-down exiled Khan: %v, want ErrNoPlayPermission", err)
	}
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type == "cast_spell" && m.Source == k.InstanceID {
			t.Fatalf("the enumerator offers a face-down Khan: %s", m.Label)
		}
	}
	if g.AnyCastPermissionsForEffect() {
		t.Error("a face-down Khan is not a reason to walk exile")
	}
}

// The bot and the client read the same answer the cast path does (ADR
// 0066 Decision 3): the owner is offered the exile cast and sees a cast
// surface on the card; an opponent is offered neither.
func TestKhanExileCastIsOfferedToItsOwnerAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	khan := khanExiledBySwords(t, g, me.ID)
	addBlack(g, me, 4)

	offered := func(seat uuid.UUID) bool {
		for _, m := range legal.EnumerateFor(g, seat) {
			if m.Type != "cast_spell" || m.Source != khan {
				continue
			}
			var p struct {
				FromZone string `json:"from_zone"`
			}
			if err := json.Unmarshal(m.Params, &p); err != nil {
				t.Fatalf("move params: %v", err)
			}
			if p.FromZone == "exile" {
				return true
			}
		}
		return false
	}
	if !offered(me.ID) {
		t.Error("the enumerator does not offer the owner Khan's cast from exile")
	}

	view := func(seat uuid.UUID) *protocol.CardView {
		v := protocol.ViewOfGameFor(g, seat.String())
		for i := range v.Exile.Cards {
			if v.Exile.Cards[i].InstanceID == khan.String() {
				return &v.Exile.Cards[i]
			}
		}
		t.Fatal("Khan is not in the exile view")
		return nil
	}
	mine := view(me.ID)
	if !mine.CastableHere {
		t.Error("the owner's view has no castable_here on the exiled Khan")
	}
	if mine.ExilePlay == nil || mine.ExilePlay.Player != me.ID.String() || !mine.ExilePlay.CastOnly {
		t.Errorf("exile_play = %+v, want the owner's own-text permission", mine.ExilePlay)
	}
	if theirs := view(opp.ID); theirs.CastableHere {
		t.Error("an opponent's view says they may cast my Khan")
	}
}

// Each copy answers for itself, by its own face: a Khan returned from exile
// to hand is cast from hand as usual, and a fresh Khan exiled afterwards is
// open again. Nothing about the first exile is remembered (CR 400.7).
func TestKhanExileOpeningFollowsTheCardNotAGrant(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	khan := khanExiledBySwords(t, g, me.ID)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneExile}, game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, khan); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneExile}, khan); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	addBlack(g, me, 4)
	if err := g.CastSpell(me.ID, khan, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("a Khan that re-entered exile is a new object with the same text: %v", err)
	}
}

// The #659 refusal holds for every card that does not print the clause:
// ZoneExile in CastableZones without the marker still panics at boot, and
// the marker without the zone is refused too.
func TestRegisterStillRefusesExileCastableWithoutOwnText(t *testing.T) {
	expectRegisterPanic(t, "declares ZoneExile in CastableZones", func() {
		Register(Spec{
			OracleID:      "test-khan-exile-without-marker",
			Name:          "Misplaced Rift Bolt",
			CastableZones: []game.ZoneKind{game.ZoneExile},
		})
	})
	expectRegisterPanic(t, "CastsItselfFromExile without listing ZoneExile", func() {
		Register(Spec{
			OracleID:             "test-khan-marker-without-zone",
			Name:                 "Half a Declaration",
			CastsItselfFromExile: true,
		})
	})
	if Has("test-khan-exile-without-marker") || Has("test-khan-marker-without-zone") {
		t.Error("a refused spec was registered anyway")
	}
	registerForTest(t, Spec{
		OracleID:             "test-khan-exile-with-marker",
		Name:                 "Proper Self-Exile Caster",
		CastableZones:        []game.ZoneKind{game.ZoneExile},
		CastsItselfFromExile: true,
	})
	if !Has("test-khan-exile-with-marker") {
		t.Error("the declaration with its marker did not register")
	}
}
