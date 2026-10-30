package effects

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// either_or_costs_4566_test.go — #4566: "sacrifice an artifact or discard
// a (nonland) card" as an activated ability's cost, as a resolution-time
// "you may … If you do / When you do", and "Ward—Discard a card or pay
// {2}". Every test drives the real cards: Bullseye, Death Dealer; K'un-Lun
// Warrior; Contract Hero; Vision of Love; Titania, Rugged Rumbler.

const (
	eoBullseyeOracle     = "10db57b4-f895-4514-b610-2252ba676353"
	eoKunLunOracle       = "0618a217-2a6e-4fa1-bfe6-89c6612a594d"
	eoContractHeroOracle = "a5c9b167-5824-4fc8-ba7a-ffa64363c029"
	eoVisionOracle       = "70c96b50-ac2b-493d-bd15-fc66eb767111"
	eoTitaniaOracle      = "e380e37d-926b-4a4b-a275-7844bf4956d5"
)

// eoTable is a catalog game in the active player's main phase, with the
// active player's and the next seat's hands emptied so each test says
// exactly what can pay.
func eoTable(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	for _, p := range []*game.Player{me, opp} {
		for p.Hand.Size() > 0 {
			if _, err := p.Hand.PopTop(); err != nil {
				t.Fatal(err)
			}
		}
	}
	return g, me, opp
}

// eoHand puts a card of the type in p's hand.
func eoHand(p *game.Player, name, typeLine string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.TypeLine = typeLine
	p.Hand.PushTop(c)
	return c.InstanceID
}

// eoArtifact puts a plain artifact under owner.
func eoArtifact(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return pushCatalogPermanent(g, owner, name, "Artifact", "", false)
}

// eoPrompt is the open either payment addressed to chooser, or nil.
func eoPrompt(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == chooser {
			if a := c.PayAction(); a != nil && a.Kind == game.PayActionEither {
				return c
			}
		}
	}
	return nil
}

// eoPay answers the open either payment with branch `option`.
func eoPay(t *testing.T, g *game.Game, chooser uuid.UUID, option int, cards ...uuid.UUID) {
	t.Helper()
	c := eoPrompt(g, chooser)
	if c == nil {
		t.Fatalf("no either payment for %s: %+v", chooser, g.PendingChoices)
	}
	if err := g.ResolvePayUnlessWithOption(c.ID, chooser, true, &option, cards); err != nil {
		t.Fatalf("pay option %d: %v", option, err)
	}
}

// eoOfferedOptions is the pay_option of every "pay" move the enumerator
// offers chooser on the open either payment.
func eoOfferedOptions(t *testing.T, g *game.Game, chooser uuid.UUID) map[int]bool {
	t.Helper()
	out := map[int]bool{}
	for _, m := range legal.EnumerateFor(g, chooser) {
		if m.Type != "resolve_choice" {
			continue
		}
		var p struct {
			Apply     *bool `json:"apply"`
			PayOption *int  `json:"pay_option"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		if p.PayOption != nil {
			out[*p.PayOption] = true
		}
	}
	return out
}

// eoActivationBranches is the cost_branch of every activation of
// `source` the enumerator offers seat; -1 marks a move with none.
func eoActivationBranches(t *testing.T, g *game.Game, seat, source uuid.UUID) map[int]bool {
	t.Helper()
	out := map[int]bool{}
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type != "activate_ability" || m.Source != source {
			continue
		}
		var p struct {
			CostBranch *int `json:"cost_branch"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		if p.CostBranch == nil {
			out[-1] = true
			continue
		}
		out[*p.CostBranch] = true
	}
	return out
}

// eoBullseye is Bullseye under me, its enters trigger settled with
// nothing to pay (so no question), untapped and not summoning sick.
func eoBullseye(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	b := pushCatalogPermanent(g, me.ID, "Bullseye, Death Dealer", "Legendary Creature — Human Assassin Villain", eoBullseyeOracle, false)
	passPriorityAroundTable(t, g)
	if eoPrompt(g, me.ID) != nil {
		t.Fatal("Bullseye's enters trigger asked with nothing to pay")
	}
	return b
}

func eoActivate(g *game.Game, me *game.Player, bull uuid.UUID, branch *int, sac, disc []uuid.UUID, target uuid.UUID) error {
	return g.ActivateCatalogAbility(me.ID, bull, 0, game.ActivateAbilityParams{
		CostBranch:   branch,
		SacrificeIDs: sac,
		DiscardIDs:   disc,
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: target}},
	})
}

// eoTopAbility is the activated ability on the stack, or nil.
func eoTopAbility(g *game.Game) *game.StackItem {
	for _, it := range g.StackMeta {
		if it != nil && it.Kind == game.StackItemActivated {
			return it
		}
	}
	return nil
}

// --- Bullseye's activated ability: an either/or activation cost --------

func TestEitherOrBullseyeActivationPaysTheBranchItNames(t *testing.T) {
	for _, tc := range []struct {
		name   string
		branch int
	}{{"sacrifice an artifact", 0}, {"discard a nonland card", 1}} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp := eoTable(t)
			bull := eoBullseye(t, g, me)
			art := eoArtifact(g, me.ID, "Bauble")
			spell := eoHand(me, "Shock", "Instant")
			fillPool(me, 3)
			before := opp.Life

			var sac, disc []uuid.UUID
			if tc.branch == 0 {
				sac = []uuid.UUID{art}
			} else {
				disc = []uuid.UUID{spell}
			}
			branch := tc.branch
			if err := eoActivate(g, me, bull, &branch, sac, disc, opp.ID); err != nil {
				t.Fatalf("activate: %v", err)
			}
			if it := eoTopAbility(g); it == nil || it.Paid.CostBranch != tc.branch+1 {
				t.Fatalf("the stack item records branch %v, want %d", it, tc.branch+1)
			}
			passPriorityAroundTable(t, g)
			if got := before - opp.Life; got != 2 {
				t.Errorf("damage %d, want 2", got)
			}
			if onBattlefield(g, art) == (tc.branch == 0) {
				t.Errorf("the artifact on the battlefield = %v after branch %d", onBattlefield(g, art), tc.branch)
			}
			if me.Hand.Contains(spell) == (tc.branch == 1) {
				t.Errorf("the nonland card in hand = %v after branch %d", me.Hand.Contains(spell), tc.branch)
			}
		})
	}
}

// No branch, a branch out of range, or a land for "a nonland card" is
// refused with nothing paid.
func TestEitherOrBullseyeActivationRefusesABadBranch(t *testing.T) {
	g, me, opp := eoTable(t)
	bull := eoBullseye(t, g, me)
	art := eoArtifact(g, me.ID, "Bauble")
	land := eoHand(me, "Forest", "Basic Land — Forest")
	fillPool(me, 3)
	two, one := 2, 1
	for name, err := range map[string]error{
		"no branch":    eoActivate(g, me, bull, nil, []uuid.UUID{art}, nil, opp.ID),
		"out of range": eoActivate(g, me, bull, &two, []uuid.UUID{art}, nil, opp.ID),
		"a land":       eoActivate(g, me, bull, &one, nil, []uuid.UUID{land}, opp.ID),
	} {
		if !errors.Is(err, game.ErrInvalidParam) {
			t.Errorf("%s: err %v, want a refusal", name, err)
		}
	}
	if c, _ := g.LookupCardForEffect(bull); c.Tapped || !onBattlefield(g, art) || !me.Hand.Contains(land) || eoTopAbility(g) != nil {
		t.Error("a refused activation paid something")
	}
}

// The enumerator offers each branch the seat can pay, and with neither
// it offers no activation (CR 118.3).
func TestEitherOrBullseyeActivationOffersOnlyPayableBranches(t *testing.T) {
	for _, tc := range []struct {
		name          string
		artifact      bool
		hand          string
		want, notWant []int
	}{
		{"only an artifact", true, "Basic Land — Forest", []int{0}, []int{1, -1}},
		{"only a nonland card", false, "Instant", []int{1}, []int{0, -1}},
		{"both", true, "Instant", []int{0, 1}, []int{-1}},
		{"neither", false, "Basic Land — Forest", nil, []int{0, 1, -1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := eoTable(t)
			bull := eoBullseye(t, g, me)
			if tc.artifact {
				eoArtifact(g, me.ID, "Bauble")
			}
			eoHand(me, "Card", tc.hand)
			fillPool(me, 3)
			got := eoActivationBranches(t, g, me.ID, bull)
			for _, b := range tc.want {
				if !got[b] {
					t.Errorf("branch %d not offered: %v", b, got)
				}
			}
			for _, b := range tc.notWant {
				if got[b] {
					t.Errorf("branch %d offered: %v", b, got)
				}
			}
		})
	}
}

// The branch an activation paid is on the stack item and survives a
// restore (and so an undo, which restores a snapshot).
func TestEitherOrBullseyeActivationBranchSurvivesARestore(t *testing.T) {
	g, me, opp := eoTable(t)
	bull := eoBullseye(t, g, me)
	spell := eoHand(me, "Shock", "Instant")
	fillPool(me, 3)
	one := 1
	if err := eoActivate(g, me, bull, &one, nil, []uuid.UUID{spell}, opp.ID); err != nil {
		t.Fatal(err)
	}
	g = restoreRoundTrip(t, g, true)
	opp = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	if it := eoTopAbility(g); it == nil || it.Paid.CostBranch != 2 {
		t.Fatalf("restored stack item %+v, want branch 2 recorded", it)
	}
	before := opp.Life
	passPriorityAroundTable(t, g)
	if before-opp.Life != 2 {
		t.Errorf("the restored activation dealt %d, want 2", before-opp.Life)
	}
}

// --- Bullseye's enters trigger: an either/or may-pay, "When you do" -----

func TestEitherOrBullseyeEntersSacrificeBranchDealsTwo(t *testing.T) {
	g, me, opp := eoTable(t)
	art := eoArtifact(g, me.ID, "Bauble")
	eoHand(me, "Forest", "Basic Land — Forest")
	enterFromHand(t, g, me.ID, "Bullseye, Death Dealer", "Legendary Creature — Human Assassin Villain", eoBullseyeOracle)
	passPriorityAroundTable(t, g)
	if got := eoOfferedOptions(t, g, me.ID); !got[0] || got[1] {
		t.Fatalf("offered %v, want the sacrifice only (the hand is a land)", got)
	}
	before := opp.Life
	eoPay(t, g, me.ID, 0, art)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if before-opp.Life != 2 || onBattlefield(g, art) {
		t.Errorf("damage %d, artifact still there %v; want 2 and gone", before-opp.Life, onBattlefield(g, art))
	}
}

func TestEitherOrBullseyeEntersDiscardBranchTakesNoLand(t *testing.T) {
	g, me, opp := eoTable(t)
	land := eoHand(me, "Forest", "Basic Land — Forest")
	spell := eoHand(me, "Shock", "Instant")
	enterFromHand(t, g, me.ID, "Bullseye, Death Dealer", "Legendary Creature — Human Assassin Villain", eoBullseyeOracle)
	passPriorityAroundTable(t, g)
	c := eoPrompt(g, me.ID)
	if c == nil {
		t.Fatal("no question")
	}
	branch := game.EitherBranch(c.PayAction(), 1)
	if opts := g.PayActionOptionsForEffect(me.ID, branch); len(opts) != 1 || opts[0] != spell {
		t.Fatalf("discard options %v, want the nonland card only", opts)
	}
	one := 1
	if err := g.ResolvePayUnlessWithOption(c.ID, me.ID, true, &one, []uuid.UUID{land}); err == nil {
		t.Fatal("a land paid \"discard a nonland card\"")
	}
	before := opp.Life
	eoPay(t, g, me.ID, 1, spell)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if before-opp.Life != 2 || me.Hand.Contains(spell) || !me.Hand.Contains(land) {
		t.Errorf("damage %d; want 2, the spell discarded and the land kept", before-opp.Life)
	}
}

func TestEitherOrBullseyeEntersDeclinedOrUnpayableDealsNothing(t *testing.T) {
	for _, payable := range []bool{true, false} {
		g, me, opp := eoTable(t)
		eoHand(me, "Forest", "Basic Land — Forest")
		if payable {
			eoArtifact(g, me.ID, "Bauble")
		}
		enterFromHand(t, g, me.ID, "Bullseye, Death Dealer", "Legendary Creature — Human Assassin Villain", eoBullseyeOracle)
		passPriorityAroundTable(t, g)
		if (eoPrompt(g, me.ID) != nil) != payable {
			t.Fatalf("payable %v: asked = %v", payable, eoPrompt(g, me.ID) != nil)
		}
		if payable {
			answerPayUnless(t, g, me.ID, false)
		}
		passPriorityAroundTable(t, g)
		if opp.Life != 40 || len(g.StackMeta) != 0 {
			t.Errorf("payable %v: life %d, stack %d; want no damage and nothing created", payable, opp.Life, len(g.StackMeta))
		}
	}
}

// The open question survives a restore with its branches, the nonland
// rule included.
func TestEitherOrBullseyeEntersQuestionSurvivesARestore(t *testing.T) {
	g, me, _ := eoTable(t)
	eoHand(me, "Forest", "Basic Land — Forest")
	spell := eoHand(me, "Shock", "Instant")
	enterFromHand(t, g, me.ID, "Bullseye, Death Dealer", "Legendary Creature — Human Assassin Villain", eoBullseyeOracle)
	passPriorityAroundTable(t, g)
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	c := eoPrompt(g, me.ID)
	if c == nil || !c.PayAction().Either[1].Nonland {
		t.Fatalf("restored question %+v lost its branches", c)
	}
	eoPay(t, g, me.ID, 1, spell)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != 38 {
		t.Errorf("life %d after the restored payment, want 38", opp.Life)
	}
}

// --- K'un-Lun Warrior, Vision of Love: "If you do, draw" ---------------

func TestEitherOrKunLunWarriorDrawsOnlyWhenABranchIsPaid(t *testing.T) {
	for _, tc := range []struct {
		name   string
		option int // -1 declines
	}{{"sacrifice", 0}, {"discard", 1}, {"decline", -1}} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := eoTable(t)
			art := eoArtifact(g, me.ID, "Bauble")
			land := eoHand(me, "Forest", "Basic Land — Forest")
			enterFromHand(t, g, me.ID, "K'un-Lun Warrior", "Creature — Human Warrior Hero", eoKunLunOracle)
			passPriorityAroundTable(t, g)
			hand := me.Hand.Size()
			switch tc.option {
			case 0:
				eoPay(t, g, me.ID, 0, art)
			case 1:
				// "Discard a card": a land pays.
				eoPay(t, g, me.ID, 1, land)
			default:
				answerPayUnless(t, g, me.ID, false)
			}
			passPriorityAroundTable(t, g)
			want := hand
			if tc.option == 0 {
				want = hand + 1
			}
			if got := me.Hand.Size(); got != want {
				t.Errorf("hand %d, want %d", got, want)
			}
		})
	}
}

func TestEitherOrVisionOfLoveDrawsTwoForEitherBranch(t *testing.T) {
	g, me, _ := eoTable(t)
	art := eoArtifact(g, me.ID, "Bauble")
	castCatalogSpell(t, g, "Vision of Love", "Instant", eoVisionOracle, nil)
	passPriorityAroundTable(t, g)
	if got := eoOfferedOptions(t, g, me.ID); !got[0] || got[1] {
		t.Fatalf("offered %v, want the sacrifice only (the hand is empty)", got)
	}
	hand := me.Hand.Size()
	eoPay(t, g, me.ID, 0, art)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 || onBattlefield(g, art) {
		t.Errorf("drew %d, artifact there %v; want 2 and gone", got, onBattlefield(g, art))
	}

	g, me, _ = eoTable(t)
	castCatalogSpell(t, g, "Vision of Love", "Instant", eoVisionOracle, nil)
	passPriorityAroundTable(t, g)
	if eoPrompt(g, me.ID) != nil || me.Hand.Size() != 0 {
		t.Error("Vision of Love asked, or drew, with nothing to pay")
	}
}

// --- Contract Hero: a Treasure, then "If you do, +2/+0" on attack -------

func TestEitherOrContractHeroSacrificesItsTreasureForTwoPower(t *testing.T) {
	g, me, opp := eoTable(t)
	hero := enterFromHand(t, g, me.ID, "Contract Hero", "Creature — Human Mercenary Hero", eoContractHeroOracle)
	passPriorityAroundTable(t, g)
	treasure := uuid.Nil
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Treasure" && c.Controller == me.ID {
			treasure = c.InstanceID
		}
	}
	if treasure == uuid.Nil {
		t.Fatal("no Treasure on entry")
	}
	// It has been under my control since the turn began, so it may attack.
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == hero {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	}
	base := effectivePower(t, g, hero)
	declareAttack(t, g, opp.ID, hero)
	passPriorityAroundTable(t, g)
	if got := eoOfferedOptions(t, g, me.ID); !got[0] || got[1] {
		t.Fatalf("offered %v, want the sacrifice only (the hand is empty)", got)
	}
	eoPay(t, g, me.ID, 0, treasure)
	if got := effectivePower(t, g, hero); got != base+2 {
		t.Errorf("power %d, want %d", got, base+2)
	}
	if onBattlefield(g, treasure) {
		t.Error("the Treasure was not sacrificed")
	}
}

// --- Titania, Rugged Rumbler: "Ward—Discard a card or pay {2}" ----------

func TestEitherOrTitaniaWardEitherBranchKeepsTheSpell(t *testing.T) {
	for _, tc := range []struct {
		name   string
		option int // -1 declines
	}{{"discard", 0}, {"pay {2}", 1}, {"decline", -1}} {
		t.Run(tc.name, func(t *testing.T) {
			g, _, opp, titania := wardTable(t, "Titania, Rugged Rumbler", "Legendary Creature — Human Villain", eoTitaniaOracle)
			filler := fillHand(opp, "Filler")
			bolt := boltWardedCreature(t, g, opp, titania)
			opp.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
			for i := 0; i < 8 && eoPrompt(g, opp.ID) == nil; i++ {
				if err := g.PassPriority(); err != nil {
					t.Fatal(err)
				}
			}
			if got := eoOfferedOptions(t, g, opp.ID); !got[0] || !got[1] {
				t.Fatalf("offered %v, want both ways to pay", got)
			}
			switch tc.option {
			case 0:
				eoPay(t, g, opp.ID, 0, filler[0])
			case 1:
				eoPay(t, g, opp.ID, 1)
			default:
				answerPayUnless(t, g, opp.ID, false)
			}
			passPriorityAroundTable(t, g)
			paid := tc.option >= 0
			if boltResolved(g, titania) != paid {
				t.Errorf("bolt resolved %v, want %v", boltResolved(g, titania), paid)
			}
			if !paid && !opp.Graveyard.Contains(bolt) {
				t.Error("the declined ward did not counter the bolt")
			}
			if opp.Hand.Contains(filler[0]) == (tc.option == 0) {
				t.Error("the discard branch's card is where it should not be")
			}
			if left := opp.ManaPool.Total(); (left == 0) != (tc.option == 1) {
				t.Errorf("pool left %d after %s", left, tc.name)
			}
		})
	}
}

// Only the branch the caster can pay is offered; with neither, the spell
// is countered and nobody is asked (CR 118.3).
func TestEitherOrTitaniaWardOffersOnlyPayableBranches(t *testing.T) {
	g, _, opp, titania := wardTable(t, "Titania, Rugged Rumbler", "Legendary Creature — Human Villain", eoTitaniaOracle)
	for opp.Hand.Size() > 0 {
		if _, err := opp.Hand.PopTop(); err != nil {
			t.Fatal(err)
		}
	}
	bolt := boltWardedCreature(t, g, opp, titania)
	for i := 0; i < 8 && eoPrompt(g, opp.ID) == nil && !stackFullyEmpty(g); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if eoPrompt(g, opp.ID) != nil {
		t.Fatal("asked a caster with no card and no mana")
	}
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) || boltResolved(g, titania) {
		t.Error("a ward the caster can't pay must counter the spell")
	}

	g, _, opp, titania = wardTable(t, "Titania, Rugged Rumbler", "Legendary Creature — Human Villain", eoTitaniaOracle)
	fillHand(opp, "Filler")
	boltWardedCreature(t, g, opp, titania)
	for i := 0; i < 8 && eoPrompt(g, opp.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if got := eoOfferedOptions(t, g, opp.ID); !got[0] || got[1] {
		t.Errorf("offered %v, want the discard only (no mana)", got)
	}
}

// Titania's cast cost is the ADR 0100 either/or additional cost.
func TestEitherOrTitaniaCastsForEitherAdditionalCost(t *testing.T) {
	g, me, _ := eoTable(t)
	card := eoHand(me, "Titania, Rugged Rumbler", "Legendary Creature — Human Villain")
	for i := range me.Hand.Cards {
		if me.Hand.Cards[i].InstanceID == card {
			me.Hand.Cards[i].OracleID = eoTitaniaOracle
		}
	}
	filler := eoHand(me, "Filler", "Sorcery")
	zero := 0
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{CostBranch: &zero, DiscardIDs: []uuid.UUID{filler}}); err != nil {
		t.Fatalf("cast discarding a card: %v", err)
	}
	if me.Hand.Contains(filler) {
		t.Error("the discard branch did not discard")
	}
}
