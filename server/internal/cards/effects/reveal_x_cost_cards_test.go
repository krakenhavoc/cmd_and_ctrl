package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reveal_x_cost_cards_test.go — #2598: "Reveal X <colour> cards from
// your hand" as an activated ability's cost (game.RevealCardsCost, ADR
// 0020's 2026-10-08 amendment) and the five Martyrs it pays for.

const (
	martyrOfBonesOracle  = "a099d8fa-d51e-4dbc-a03f-6f912097d41e"
	martyrOfSandsOracle  = "f2048140-b691-4753-8239-1aa75059e389"
	martyrOfSporesOracle = "46c128d5-92b7-4132-a865-4db299879579"
	martyrOfFrostOracle  = "196cacb0-a26a-4540-9143-57867a638085"
	martyrOfAshesOracle  = "7f00bc45-7c65-4455-9db0-58bd79bcdb4b"
)

// martyrTable seats a Martyr on seat 0's battlefield (not summoning
// sick: the ability has no {T}) with `mana` generic mana in the pool,
// and a hand of two black cards, one white card and one red card.
func martyrTable(t *testing.T, name, oracle string, mana int) (g *game.Game, me *game.Player, martyr uuid.UUID, hand map[string]uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[0]
	martyr = pushCatalogPermanent(g, me.ID, name, "Creature — Human Wizard", oracle, false)
	for i := 0; i < mana; i++ {
		me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
	hand = map[string]uuid.UUID{
		"black1": ccHandCard(me, "Black One", "Sorcery", "{1}{B}"),
		"black2": ccHandCard(me, "Black Two", "Instant", "{B}"),
		"white":  ccHandCard(me, "White One", "Sorcery", "{W}"),
		"red":    ccHandCard(me, "Red One", "Sorcery", "{R}"),
	}
	return g, me, martyr, hand
}

func ids(xs ...uuid.UUID) []uuid.UUID { return xs }

func onBattlefieldByID(g *game.Game, id uuid.UUID) bool { return g.Battlefield.Contains(id) }

// graveyardCards puts n cards in p's graveyard and returns their IDs.
func graveyardCards(p *game.Player, prefix string, n int) []uuid.UUID {
	var out []uuid.UUID
	for i := 0; i < n; i++ {
		out = append(out, pushGraveyardCard(nil, p, prefix))
	}
	return out
}

func gyTargets(cards ...uuid.UUID) []game.TargetRef {
	var out []game.TargetRef
	for _, c := range cards {
		out = append(out, game.TargetRef{Kind: game.TargetCard, ID: c})
	}
	return out
}

// The cost reveals X black cards: they are shown to the whole table,
// stay in the hand, the mana is the printed {1} whatever X is, the
// Martyr is sacrificed, and the targeted graveyard cards are exiled.
func TestMartyrOfBonesRevealsXBlackCardsAndExilesUpToX(t *testing.T) {
	g, me, martyr, hand := martyrTable(t, "Martyr of Bones", martyrOfBonesOracle, 1)
	opp := g.Seats[1]
	gy := graveyardCards(opp, "Their Card", 3)
	handBefore := me.Hand.Size()

	if err := g.ActivateCatalogAbility(me.ID, martyr, 0, game.ActivateAbilityParams{
		XValue:    2,
		RevealIDs: ids(hand["black1"], hand["black2"]),
		Targets:   gyTargets(gy[0], gy[1]),
	}); err != nil {
		t.Fatalf("activate at X=2: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("mana pool = %d, want the printed {1} spent and nothing more", len(me.ManaPool))
	}
	if onBattlefieldByID(g, martyr) {
		t.Error("the Martyr was not sacrificed")
	}
	if me.Hand.Size() != handBefore {
		t.Errorf("hand = %d, want %d — revealing moves nothing", me.Hand.Size(), handBefore)
	}
	for _, name := range []string{"black1", "black2"} {
		if got := revealedEvents(g, hand[name]); got != 1 {
			t.Errorf("%s: %d reveal events, want 1", name, got)
		}
		for _, seat := range g.Seats {
			var known bool
			for i := range me.Hand.Cards {
				if me.Hand.Cards[i].InstanceID == hand[name] {
					known = me.Hand.Cards[i].IsKnownTo(seat.ID)
				}
			}
			if !known {
				t.Errorf("%s: seat %s does not know the revealed card", name, seat.ID)
			}
		}
	}
	for _, name := range []string{"white", "red"} {
		if got := revealedEvents(g, hand[name]); got != 0 {
			t.Errorf("%s was revealed", name)
		}
	}
	if g.Exile.Contains(gy[0]) {
		t.Fatal("the graveyard was exiled before the ability resolved")
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(gy[0]) || !g.Exile.Contains(gy[1]) {
		t.Error("the two targeted cards are not in exile")
	}
	if g.Exile.Contains(gy[2]) || !opp.Graveyard.Contains(gy[2]) {
		t.Error("the untargeted card left the graveyard")
	}
}

// "Up to X": fewer targets than X is legal, and so is X = 0 with none.
func TestMartyrOfBonesUpToXAndXZero(t *testing.T) {
	g, me, martyr, hand := martyrTable(t, "Martyr of Bones", martyrOfBonesOracle, 1)
	opp := g.Seats[1]
	gy := graveyardCards(opp, "Their Card", 2)
	if err := g.ActivateCatalogAbility(me.ID, martyr, 0, game.ActivateAbilityParams{
		XValue: 2, RevealIDs: ids(hand["black1"], hand["black2"]), Targets: gyTargets(gy[0]),
	}); err != nil {
		t.Fatalf("one target at X=2: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(gy[0]) || g.Exile.Contains(gy[1]) {
		t.Error("exactly the one named card should be exiled")
	}

	g2, me2, martyr2, _ := martyrTable(t, "Martyr of Bones", martyrOfBonesOracle, 1)
	gy2 := graveyardCards(g2.Seats[1], "Their Card", 1)
	if err := g2.ActivateCatalogAbility(me2.ID, martyr2, 0, game.ActivateAbilityParams{XValue: 0}); err != nil {
		t.Fatalf("X=0: %v", err)
	}
	passPriorityAroundTable(t, g2)
	if g2.Exile.Contains(gy2[0]) {
		t.Error("X=0 exiled a card")
	}
	if onBattlefieldByID(g2, martyr2) {
		t.Error("the Martyr is the sacrificed cost at X=0 too")
	}
}

// The announcement and the payment agree in both directions, the
// colour is enforced, and a refusal spends nothing.
func TestMartyrOfBonesRefusals(t *testing.T) {
	g, me, martyr, hand := martyrTable(t, "Martyr of Bones", martyrOfBonesOracle, 1)
	opp := g.Seats[1]
	gy := graveyardCards(opp, "Their Card", 2)
	mine := graveyardCards(me, "My Card", 1)
	theirBlack := ccHandCard(opp, "Their Black", "Sorcery", "{B}")
	handBefore := me.Hand.Size()

	for _, tc := range []struct {
		why string
		p   game.ActivateAbilityParams
	}{
		{"X=2 revealing one", game.ActivateAbilityParams{XValue: 2, RevealIDs: ids(hand["black1"])}},
		{"X=1 revealing two", game.ActivateAbilityParams{XValue: 1, RevealIDs: ids(hand["black1"], hand["black2"])}},
		{"X=2 naming one card twice", game.ActivateAbilityParams{XValue: 2, RevealIDs: ids(hand["black1"], hand["black1"])}},
		{"a white card for a black reveal", game.ActivateAbilityParams{XValue: 1, RevealIDs: ids(hand["white"])}},
		{"a red card among the black", game.ActivateAbilityParams{XValue: 2, RevealIDs: ids(hand["black1"], hand["red"])}},
		{"another player's black card", game.ActivateAbilityParams{XValue: 1, RevealIDs: ids(theirBlack)}},
		{"a card that does not exist", game.ActivateAbilityParams{XValue: 1, RevealIDs: ids(uuid.New())}},
		{"X=1 revealing nothing", game.ActivateAbilityParams{XValue: 1}},
		{"X=0 revealing one", game.ActivateAbilityParams{XValue: 0, RevealIDs: ids(hand["black1"])}},
		{"more targets than X", game.ActivateAbilityParams{XValue: 1, RevealIDs: ids(hand["black1"]), Targets: gyTargets(gy[0], gy[1])}},
		{"two graveyards", game.ActivateAbilityParams{XValue: 2, RevealIDs: ids(hand["black1"], hand["black2"]), Targets: gyTargets(gy[0], mine[0])}},
		{"X=-1", game.ActivateAbilityParams{XValue: -1}},
	} {
		err := g.ActivateCatalogAbility(me.ID, martyr, 0, tc.p)
		if err == nil {
			t.Errorf("%s: activation accepted", tc.why)
			continue
		}
		if len(me.ManaPool) != 1 || me.Hand.Size() != handBefore || !onBattlefieldByID(g, martyr) {
			t.Fatalf("%s: a refused activation paid something (pool %d, hand %d, martyr on board %v)",
				tc.why, len(me.ManaPool), me.Hand.Size(), onBattlefieldByID(g, martyr))
		}
		if revealedEvents(g, hand["black1"])+revealedEvents(g, hand["black2"]) != 0 {
			t.Fatalf("%s: a refused activation revealed a card", tc.why)
		}
	}
}

// Reveal ids sent to an ability that has no reveal cost are refused
// rather than ignored.
func TestRevealIDsOnAnAbilityWithoutARevealCostAreRefused(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bomb := pushCatalogPermanent(g, me.ID, "Goblin Bombardment", "Enchantment", goblinBombardmentOracle, false)
	fodder := pushCatalogPermanent(g, me.ID, "Fodder", "Creature — Goblin", "", false)
	card := ccHandCard(me, "Black One", "Sorcery", "{1}{B}")
	err := g.ActivateCatalogAbility(me.ID, bomb, 0, game.ActivateAbilityParams{
		SacrificeIDs: ids(fodder), RevealIDs: ids(card),
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}},
	})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("err = %v, want ErrInvalidParam", err)
	}
}

// Martyr of Sands: three times X life.
func TestMartyrOfSandsGainsThreeTimesX(t *testing.T) {
	g, me, martyr, hand := martyrTable(t, "Martyr of Sands", martyrOfSandsOracle, 1)
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, martyr, 0, game.ActivateAbilityParams{
		XValue: 1, RevealIDs: ids(hand["white"]),
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life = %d, want %d", me.Life, life+3)
	}
	// A black card is not white.
	g2, me2, martyr2, hand2 := martyrTable(t, "Martyr of Sands", martyrOfSandsOracle, 1)
	if err := g2.ActivateCatalogAbility(me2.ID, martyr2, 0, game.ActivateAbilityParams{
		XValue: 1, RevealIDs: ids(hand2["black1"]),
	}); err == nil {
		t.Error("a black card paid a white reveal")
	}
}

// Martyr of Spores: +X/+X on the target creature until end of turn.
func TestMartyrOfSporesPumpsByX(t *testing.T) {
	g, me, martyr, hand := martyrTable(t, "Martyr of Spores", martyrOfSporesOracle, 1)
	// A green card in hand: ccHandCard colours by mana cost.
	green1 := ccHandCard(me, "Green One", "Sorcery", "{G}")
	green2 := ccHandCard(me, "Green Two", "Sorcery", "{2}{G}")
	_ = hand
	bear := pushCreatureToBattlefieldForTest(g, me.ID, "Bear")
	if err := g.ActivateCatalogAbility(me.ID, martyr, 0, game.ActivateAbilityParams{
		XValue: 2, RevealIDs: ids(green1, green2), Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, bear); got != 4 {
		t.Errorf("power = %d, want 2+2", got)
	}
}

// Martyr of Ashes: X damage to each creature without flying, the
// flier untouched.
func TestMartyrOfAshesDamagesTheGroundOnly(t *testing.T) {
	g, me, martyr, _ := martyrTable(t, "Martyr of Ashes", martyrOfAshesOracle, 2)
	red1 := ccHandCard(me, "Red A", "Sorcery", "{R}")
	red2 := ccHandCard(me, "Red B", "Sorcery", "{R}")
	bear := pushCreatureToBattlefieldForTest(g, g.Seats[1].ID, "Bear")
	drake := pushCatalogPermanent(g, g.Seats[1].ID, "Drake", "Creature — Drake", "", false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == drake {
				g.Battlefield.Cards[i].Toughness = 2
				g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, "flying")
			}
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, martyr, 0, game.ActivateAbilityParams{
		XValue: 2, RevealIDs: ids(red1, red2),
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if onBattlefieldByID(g, bear) {
		t.Error("the 2/2 without flying survived 2 damage")
	}
	if !onBattlefieldByID(g, drake) {
		t.Error("the flier took damage")
	}
}

// Martyr of Frost: the target spell's controller is asked to pay {X}.
func TestMartyrOfFrostTaxesByX(t *testing.T) {
	g, me, martyr, _ := martyrTable(t, "Martyr of Frost", martyrOfFrostOracle, 2)
	blue1 := ccHandCard(me, "Blue A", "Sorcery", "{U}")
	blue2 := ccHandCard(me, "Blue B", "Sorcery", "{U}")
	victim := castCatalogSpell(t, g, "Divination", "Sorcery", divinationDelveOracle, nil)
	if err := g.ActivateCatalogAbility(me.ID, martyr, 0, game.ActivateAbilityParams{
		XValue: 2, RevealIDs: ids(blue1, blue2), Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := pendingOfKind(g, game.PendingChoicePayUnless)
	if c == nil {
		t.Fatalf("no pay-unless prompt: %+v", g.PendingChoices)
	}
	if c.PayCost != "{2}" {
		t.Errorf("tax = %q, want {2}", c.PayCost)
	}
}

// The five are catalogued in full and declare the X.
func TestMartyrsAreCatalogedFull(t *testing.T) {
	for _, oracle := range []string{martyrOfBonesOracle, martyrOfSandsOracle, martyrOfSporesOracle, martyrOfFrostOracle, martyrOfAshesOracle} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Fatalf("%s is not in the catalog", oracle)
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness = %v, caveats = %v, want Full with none", spec.Name, spec.Completeness, spec.Caveats)
		}
		if !spec.XMatters {
			t.Errorf("%s: its whole effect is X, so XMatters keeps the bot off the X=0 no-op", spec.Name)
		}
		if len(spec.Activated) != 1 || !spec.Activated[0].Cost.DemandsX() || !game.RevealCardsCountFromX(spec.Activated[0].Cost.RevealCards) {
			t.Errorf("%s: the ability should announce X through its reveal cost", spec.Name)
		}
	}
}

// --- the registry refuses the shapes that cannot work ------------------

func TestRegisterRejectsRevealCostMisuse(t *testing.T) {
	noop := func(*game.Game, *game.StackItem) error { return nil }
	cases := []struct {
		name string
		spec Spec
		want string
	}{
		{"{X} in the mana cost beside it", Spec{OracleID: "rx-guard-two-x", Name: "rx-guard-two-x", Activated: []ActivatedAbility{{
			Label: "x", Cost: Plus(ManaCost("{X}{B}"), RevealX("X black cards", "B")), Effect: noop,
		}}}, "one announced X cannot pay both"},
		{"discard X beside it", Spec{OracleID: "rx-guard-discard-x", Name: "rx-guard-discard-x", Activated: []ActivatedAbility{{
			Label: "x", Cost: Plus(DiscardX("X cards"), RevealX("X black cards", "B")), Effect: noop,
		}}}, "one announced X cannot pay both"},
		{"sacrifice X beside it", Spec{OracleID: "rx-guard-sac-x", Name: "rx-guard-sac-x", Activated: []ActivatedAbility{{
			Label: "x", Cost: Plus(SacrificeX("X creatures", Creature()), RevealX("X black cards", "B")), Effect: noop,
		}}}, "one announced X cannot pay both"},
		{"a count as well", Spec{OracleID: "rx-guard-count", Name: "rx-guard-count", Activated: []ActivatedAbility{{
			Label: "x", Cost: game.AbilityCost{RevealCards: &game.RevealCardsCost{N: 2, CountFromX: true, Label: "x"}}, Effect: noop,
		}}}, "also declares a fixed count"},
		{"a fixed count of zero", Spec{OracleID: "rx-guard-zero", Name: "rx-guard-zero", Activated: []ActivatedAbility{{
			Label: "x", Cost: RevealN(0, "no cards", "B"), Effect: noop,
		}}}, "at least one"},
		{"an unknown colour", Spec{OracleID: "rx-guard-colour", Name: "rx-guard-colour", Activated: []ActivatedAbility{{
			Label: "x", Cost: RevealX("X pink cards", "P"), Effect: noop,
		}}}, "one of W, U, B, R, G"},
		{"behold", Spec{OracleID: "rx-guard-behold", Name: "rx-guard-behold", Activated: []ActivatedAbility{{
			Label: "x", Cost: game.AbilityCost{RevealCards: &game.RevealCardsCost{RevealCost: game.RevealCost{Behold: true}, N: 1, Label: "x"}}, Effect: noop,
		}}}, "behold"},
		{"no label", Spec{OracleID: "rx-guard-label", Name: "rx-guard-label", Activated: []ActivatedAbility{{
			Label: "x", Cost: game.AbilityCost{RevealCards: &game.RevealCardsCost{N: 1}}, Effect: noop,
		}}}, "no Label"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expectRegisterPanic(t, tc.want, func() { Register(tc.spec) })
		})
	}
}

// A spell's reveal branch names a creature type only: a colour there is
// a card file that has mixed the two owners up.
func TestRegisterRejectsAColourOnASpellRevealBranch(t *testing.T) {
	branch := &game.AdditionalCost{Reveal: &game.RevealCost{Subtype: "Elf", Color: "G"}, Label: "x", Key: "a"}
	expectRegisterPanic(t, "narrowed by colour", func() {
		Register(Spec{OracleID: "rx-guard-branch", Name: "rx-guard-branch", AdditionalCost: &game.AdditionalCost{
			Either: []game.AdditionalCost{*branch, {ManaCost: "{3}", Label: "pay", Key: "b"}},
		}})
	})
}
