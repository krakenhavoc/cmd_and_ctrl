package testsi

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsa "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/a"
	cardsch "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/ch"
	cardsex "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/ex"
	cardsh "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/h"
	cardskr "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/kr"
	cardslu "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/lu"
	cardsri "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/ri"
	cardsso "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/so"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// keyed_asked_item_restore_test.go — #4264: a keyed pay-unless or
// any-player prompt carries the item that asked it as plain data
// (PendingChoice.Asked), so the remaining may-pay, any-player, upkeep and
// echo callers are restore points whose branches read what a closure over
// the item did: the asking trigger's event, its targets, the object (and
// epoch, CR 400.7) the ability came from, and the IDs the card put on the
// prompt. Each test takes a STRICT restore while the prompt is open and
// answers it on the restored table.

// askedPrompt is the asked item of the open prompt `chooser` owes, with the
// prompt, failing the test when there is none.
func askedPrompt(t *testing.T, g *game.Game, chooser uuid.UUID) (*game.PendingChoice, *game.AskedItem) {
	t.Helper()
	c := effectstest.PayUnlessFor(g, chooser)
	if c == nil {
		t.Fatalf("no pay-unless prompt for %s", chooser)
	}
	if c.Asked == nil {
		t.Fatalf("the prompt (key %q) carries no asking item", c.PickThen)
	}
	return c, c.Asked
}

// Mana Vault: "you may pay {4}. If you do, untap it." The Vault that asked
// is an object with an epoch, and the epoch rides the prompt through a
// restore: a Vault that left and came back while the payer decided is a new
// object and stays tapped (CR 400.7).
func TestManaVaultOfferKeepsTheVaultsObjectAcrossARestore(t *testing.T) {
	for _, flicker := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		owner := g.Seats[0]
		vault := effectstest.PushPermanentForTest(g, owner.ID, "Mana Vault", effectstest.ManaVaultOracle, "Artifact")
		if err := g.TapCard(vault, true); err != nil {
			t.Fatal(err)
		}
		effectstest.AdvanceToUpkeepOf(t, g, 0)
		effectstest.PassPriorityAroundTable(t, g)
		c, asked := askedPrompt(t, g, owner.ID)
		if c.PickThen != cardslu.ManaVaultPaid.Key() {
			t.Fatalf("the offer's key is %q, want the keyed may-pay", c.PickThen)
		}
		if asked.SourceObject == nil || asked.SourceObject.ID != vault {
			t.Fatalf("the asking item names object %+v, want the Vault %s", asked.SourceObject, vault)
		}
		epoch := asked.SourceObject.Epoch

		g = effectstest.RestoreRoundTrip(t, g, true)
		owner = g.Seats[0]
		_, asked = askedPrompt(t, g, owner.ID)
		if asked.SourceObject == nil || asked.SourceObject.ID != vault || asked.SourceObject.Epoch != epoch {
			t.Fatalf("restored: the asking item names object %+v, want the Vault at epoch %d", asked.SourceObject, epoch)
		}
		if flicker {
			g.BeginEventBatchForTest()
			effectstest.FlickerInResponse(t, g, vault)
			if err := g.TapCard(vault, true); err != nil {
				t.Fatal(err)
			}
		}
		effectstest.B06AddMana(owner, "C", "C", "C", "C")
		effectstest.AnswerPayUnless(t, g, owner.ID, true)
		if got := effectstest.UntapStaticTapped(t, g, vault); got != flicker {
			t.Errorf("flicker=%v: the Vault is tapped = %v, want %v (a new object is not untapped, the same one is)", flicker, got, flicker)
		}
	}
}

// Echo: both answers ride one key. The payment announces the echo cost
// paid, the decline sacrifices the permanent, and a permanent that left and
// came back is a new object and is not sacrificed (CR 400.7).
func TestEchoPromptIsARestorePoint(t *testing.T) {
	for _, tc := range []struct {
		name    string
		pay     bool
		flicker bool
	}{{"paid", true, false}, {"declined", false, false}, {"declined after it left and came back", false, true}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		patrol := effectstest.CastEchoCreature(t, g, "Goblin Patrol", "Creature — Goblin", effectstest.OracleGoblinPatrol)
		effectstest.LeaveUpkeepThenAdvanceTo(t, g, 0)
		effectstest.PassPriorityAroundTable(t, g)
		c, asked := askedPrompt(t, g, me.ID)
		if c.PickThen != effects.EchoPaidOrSacrificed.Key() || asked.SourceObject == nil || asked.SourceObject.ID != patrol {
			t.Fatalf("%s: the echo prompt is %+v asking %+v", tc.name, c, asked)
		}

		g = effectstest.RestoreRoundTrip(t, g, true)
		me = g.Seats[0]
		prompt := effectstest.EchoPrompt(g, me.ID)
		if prompt == nil || prompt.PickThen != effects.EchoPaidOrSacrificed.Key() || prompt.PayCost != "{R}" {
			t.Fatalf("%s: restored prompt %+v", tc.name, prompt)
		}
		if !g.ChoicePromptBlocksTable(prompt) {
			t.Errorf("%s: the restored echo prompt does not hold the upkeep", tc.name)
		}
		if tc.flicker {
			g.BeginEventBatchForTest()
			effectstest.FlickerInResponse(t, g, patrol)
		}
		if tc.pay {
			me.ManaPool.AddMana(game.ManaToken{Color: "R"})
		}
		before := len(g.Events)
		if err := g.ResolvePayUnless(prompt.ID, me.ID, tc.pay); err != nil {
			t.Fatalf("%s: answer: %v", tc.name, err)
		}
		_, there := effectstest.BattlefieldCard(g, patrol)
		if want := tc.pay || tc.flicker; there != want {
			t.Errorf("%s: Goblin Patrol on the battlefield = %v, want %v", tc.name, there, want)
		}
		announced := false
		for _, ev := range g.Events[before:] {
			if ev.Kind == game.EventEchoPaid && ev.CardID == patrol && ev.Actor == me.ID {
				announced = true
			}
		}
		if announced != tc.pay {
			t.Errorf("%s: EventEchoPaid announced = %v, want %v", tc.name, announced, tc.pay)
		}
	}
}

// Karplusan Minotaur's "Cumulative upkeep—Flip a coin": the coins flipped
// are one per age counter, worked out from the counters when the restored
// prompt is answered, and a decline sacrifices the Minotaur.
func TestKarplusanMinotaurFlipUpkeepIsARestorePoint(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		owner := g.Seats[1]
		minotaur := minotaurWithAge(g, owner.ID, 1)
		effectstest.AdvanceToUpkeepOf(t, g, 1)
		effectstest.PassPriorityAroundTable(t, g)
		if c := effectstest.RemoraUpkeepPrompt(t, g, owner.ID); c.PickThen != effects.CumulativeUpkeepFlipCoins.Key() {
			t.Fatalf("the flip upkeep's key is %q, want the keyed branches", c.PickThen)
		}

		g = effectstest.RestoreRoundTrip(t, g, true)
		owner = g.Seats[1]
		choice := effectstest.RemoraUpkeepPrompt(t, g, owner.ID)
		if choice.PickThen != effects.CumulativeUpkeepFlipCoins.Key() || choice.PayCost != "Flip two coins" {
			t.Fatalf("restored prompt: key %q cost %q", choice.PickThen, choice.PayCost)
		}
		if err := g.ResolvePayUnless(choice.ID, owner.ID, pay); err != nil {
			t.Fatal(err)
		}
		_, there := effectstest.BattlefieldCard(g, minotaur)
		flip := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceCoinCall, owner.ID)
		if pay {
			if !there || flip == nil || flip.CoinCount != 2 {
				t.Fatalf("paid: Minotaur there = %v, coin call %+v, want two coins", there, flip)
			}
			if flip.Reason != "Karplusan Minotaur — cumulative upkeep: flip a coin: flip two coins" {
				t.Errorf("the coin call asks %q", flip.Reason)
			}
		} else if there || flip != nil {
			t.Errorf("declined: Minotaur there = %v, coin call %+v, want it sacrificed and no flip", there, flip)
		}
	}
}

// Ripples of Undeath: the cards that were milled ride the prompt, so a
// restored offer still offers exactly those three.
func TestRipplesOfUndeathOfferCarriesTheMilledCardsAcrossARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushPermanentForTest(g, me.ID, "Ripples of Undeath", effectstest.RipplesOfUndeathOracle, "Enchantment")
	older := effectstest.PushGraveyardCardForTest(me, "Older Card")
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.PayUnlessFor(g, me.ID)
	// The asking effect's controller rides first, the milled cards after it.
	if c == nil || c.PickThen != cardsri.RipplesOfUndeathPaid.Key() || len(c.OptionCarry) != 4 {
		t.Fatalf("the offer %v is not the keyed may-pay carrying the controller and the three milled cards", c != nil)
	}
	milled := append([]uuid.UUID(nil), c.OptionCarry[1:]...)

	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c = effectstest.PayUnlessFor(g, me.ID)
	if c == nil || len(c.OptionCarry) != 4 {
		t.Fatalf("restored offer %v does not carry the controller and the three milled cards", c != nil)
	}
	effectstest.B06AddMana(me, "B")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 3 {
		t.Fatalf("paying on the restored table offers %+v, want the 3 milled cards", pick)
	}
	for _, id := range pick.ChooseCards {
		if id == older {
			t.Error("a card already in the graveyard is offered")
		}
		found := false
		for _, m := range milled {
			found = found || m == id
		}
		if !found {
			t.Errorf("%s was offered but not milled", id)
		}
	}
}

// Hashaton: the discarded creature is the asking trigger's event, which
// rides the prompt, so the restored payment copies it.
func TestHashatonOfferCarriesTheTriggeringEventAcrossARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.HashatonDiscardOgre(t, g, me)
	c, asked := askedPrompt(t, g, me.ID)
	if asked.Trigger == nil || !asked.Trigger.Fired() || asked.Trigger.Event.CardID == uuid.Nil {
		t.Fatalf("the asking item carries trigger %+v, want the discard event", asked.Trigger)
	}
	discarded := asked.Trigger.Event.CardID
	if c.PickThen != cardsh.HashatonPaid.Key() {
		t.Fatalf("the offer's key is %q", c.PickThen)
	}

	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	if _, asked = askedPrompt(t, g, me.ID); asked.Trigger == nil || asked.Trigger.Event.CardID != discarded {
		t.Fatalf("restored: the asking item carries %+v, want the discard of %s", asked.Trigger, discarded)
	}
	effectstest.B06AddMana(me, "U", "C", "C")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	tokenID := effectstest.FindBattlefieldByName(g, "Test Ogre")
	if tokenID == uuid.Nil {
		t.Fatal("paying {2}{U} on the restored table created no token copy")
	}
	if tok, _ := effectstest.BattlefieldCard(g, tokenID); tok.Power != 4 || !tok.Tapped || tok.Controller != me.ID {
		t.Errorf("token %d/%d tapped=%v controller=%s, want a tapped 4/4 under %s", tok.Power, tok.Toughness, tok.Tapped, tok.Controller, me.ID)
	}
}

// Lich's Relic: the opponents counted as the offer was made ride the
// prompt, so the restored payment still destroys one creature per opponent.
func TestLichsRelicOfferCarriesItsOpponentsAcrossARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	var victims []uuid.UUID
	for _, s := range g.Seats {
		if s.ID != me.ID {
			victims = append(victims, effectstest.B16Creature(g, s.ID, "Bear", "Creature — Bear", 2, 2))
		}
	}
	effectstest.RfpbCastRelic(t, g, me, 2)
	if c := effectstest.PayUnlessFor(g, me.ID); c == nil || c.PickThen != cardskr.LichsRelicPaid.Key() || len(c.OptionCarry) != len(victims)+1 {
		t.Fatalf("the offer %v is not the keyed may-pay carrying the controller and %d opponents", c != nil, len(victims))
	}

	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	for i := 0; i < 8; i++ {
		effectstest.PassPriorityAroundTable(t, g)
		pick := effectstest.LatestPickTarget(g, me.ID)
		if pick == nil {
			break
		}
		var ref []game.TargetRef
		for _, v := range victims {
			if effectstest.HasID(pick.PickTargetCards, v) {
				ref = []game.TargetRef{effectstest.RfpbCardRef(v)}
			}
		}
		if err := g.ResolvePickTargets(pick.ID, me.ID, ref); err != nil {
			t.Fatalf("resolve the target prompt: %v", err)
		}
	}
	for _, v := range victims {
		if _, ok := effectstest.BattlefieldCard(g, v); ok {
			t.Error("a victim survived the Relic on the restored table")
		}
	}
}

// Springheart Nantuko: one key names both answers. Paying copies the
// enchanted creature, declining makes the Insect, on the restored table.
func TestSpringheartNantukoOfferIsARestorePoint(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		effectstest.SweepFBestowed(g, me.ID, bear)
		effectstest.PlayLandFromHand(t, g, "Forest", "")
		effectstest.PassPriorityAroundTable(t, g)
		if c := effectstest.PayUnlessFor(g, me.ID); c == nil || c.PickThen != cardsso.SpringheartCopyOrInsect.Key() || len(c.OptionCarry) != 2 || c.OptionCarry[1] != bear {
			t.Fatalf("the offer %v is not the keyed pay-unless carrying the controller and the Bear", c != nil)
		}

		g = effectstest.RestoreRoundTrip(t, g, true)
		me = g.Seats[g.Turn.ActiveSeat]
		if pay {
			effectstest.SweepFAddMana(g, me.ID, "{C}{G}")
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		bears, insects := effectstest.CountBattlefieldNamed(g, me.ID, "Bear"), effectstest.CountBattlefieldNamed(g, me.ID, "Insect")
		if pay && (bears != 2 || insects != 0) {
			t.Errorf("paid: %d Bears and %d Insects, want a copy and no Insect", bears, insects)
		}
		if !pay && (bears != 1 || insects != 1) {
			t.Errorf("declined: %d Bears and %d Insects, want the Insect only", bears, insects)
		}
	}
}

// Argothian Wurm: "any player may sacrifice a land. If a player does, put
// this creature on top of its owner's library." The Wurm is the asking
// item's object, so one that left and came back while the offer was open
// is a new object and stays where it is (CR 400.7), on a restored table.
func TestAnyPlayerOfferKeepsTheWurmsObjectAcrossARestore(t *testing.T) {
	for _, flicker := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		land := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
		wurm := effectstest.CastCatalogSpell(t, g, "Argothian Wurm", "Creature — Wurm", effectstest.AppArgothianWurm, nil)
		effectstest.PassPriorityAroundTable(t, g)
		link := effectstest.AppOpenLink(g)
		if link == nil || link.PickThen != effects.AnyPlayerTucksThis.Key() || link.Asked == nil || link.Asked.SourceObject == nil {
			t.Fatalf("the first link is %+v, want the keyed offer carrying the Wurm's object", link)
		}

		g = effectstest.RestoreRoundTrip(t, g, true)
		me, opp = g.Seats[0], g.Seats[1]
		if link = effectstest.AppOpenLink(g); link == nil || link.Asked == nil || link.Asked.SourceObject == nil || link.Asked.SourceObject.ID != wurm {
			t.Fatalf("restored link %+v, want the offer still naming the Wurm", link)
		}
		if flicker {
			g.BeginEventBatchForTest()
			effectstest.FlickerInResponse(t, g, wurm)
		}
		effectstest.AppAnswerWith(t, g, func(c *game.PendingChoice) (bool, []uuid.UUID) {
			if c.Chooser == opp.ID {
				return true, []uuid.UUID{land}
			}
			return c.Chooser == me.ID, nil
		})
		if tucked := effectstest.AppInZone(g, wurm, game.ZoneLibrary); tucked == flicker {
			t.Errorf("flicker=%v: the Wurm is in a library = %v, want %v", flicker, tucked, !flicker)
		}
	}
}

// Aether Rift: the discarded card rides the offer as an ID, so "if no one
// does" on a restored table returns that card.
func TestAetherRiftOfferCarriesTheDiscardedCardAcrossARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[1]
	effectstest.PushCatalogPermanent(g, owner.ID, "Aether Rift", "Enchantment", effectstest.AppAetherRift, false)
	effectstest.AppToUpkeepOf(t, g, 1)
	owner.Hand.Cards = owner.Hand.Cards[:0]
	card := uuid.New()
	owner.Hand.PushTop(game.Card{InstanceID: card, Name: "Discarded", TypeLine: "Creature — Beast", Power: 3, Toughness: 3, Owner: owner.ID, Controller: owner.ID})
	effectstest.PassPriorityAroundTable(t, g)
	link := effectstest.AppOpenLink(g)
	if link == nil || link.PickThen != cardsa.AetherRiftReturns.Key() || len(link.OptionCarry) != 1 || link.OptionCarry[0] != card {
		t.Fatalf("the first link is %+v, want the keyed offer carrying the discarded card", link)
	}

	g = effectstest.RestoreRoundTrip(t, g, true)
	effectstest.AppAnswerWith(t, g, effectstest.AppOnly(uuid.Nil))
	if !effectstest.AppInZone(g, card, game.ZoneBattlefield) {
		t.Error("nobody paid on the restored table and the creature did not return")
	}
}

// A prompt naming a key this binary lacks is refused, whichever of the new
// keys it names (the rollback case): an older binary has none of them.
func TestKeyedAskedItemPromptsWithAnUnknownKeyAreRefused(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[0]
	vault := effectstest.PushPermanentForTest(g, owner.ID, "Mana Vault", effectstest.ManaVaultOracle, "Artifact")
	if err := g.TapCard(vault, true); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.RefuseUnknownKey(t, g, cardslu.ManaVaultPaid.Key())
}

// The asking item is plain data: it survives a clone and a snapshot round
// trip field for field, and building the stack item twice gives two items.
func TestAskedItemIsPlainData(t *testing.T) {
	trigger := &game.TriggerContext{Event: game.Event{Kind: game.EventDiscardCard, CardID: uuid.New(), Actor: uuid.New()}}
	item := &game.StackItem{
		ID: uuid.New(), Kind: game.StackItemTriggered, Controller: uuid.New(), Owner: uuid.New(), SourceCardID: uuid.New(),
		SourceObject:    game.ObjectRef{ID: uuid.New(), Epoch: 3},
		SourceTurnOvers: 2, Label: "label", XValue: 5, Modes: []int{1},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: uuid.New()}},
		Trigger: trigger,
	}
	asked := game.NewAskedItem(item)
	back := asked.StackItem()
	if back.ID != item.ID || back.Kind != item.Kind || back.Controller != item.Controller || back.Owner != item.Owner ||
		back.SourceCardID != item.SourceCardID || back.SourceObject != item.SourceObject ||
		back.SourceTurnOvers != item.SourceTurnOvers || back.Label != item.Label || back.XValue != item.XValue ||
		len(back.Modes) != 1 || back.Modes[0] != 1 || len(back.Targets) != 1 || back.Targets[0] != item.Targets[0] ||
		back.Trigger == nil || back.Trigger.Event.CardID != trigger.Event.CardID {
		t.Fatalf("the rebuilt item %+v differs from %+v", back, item)
	}
	back.Targets[0].ID = uuid.New()
	back.Trigger.Event.CardID = uuid.New()
	if again := asked.StackItem(); again.Targets[0] != item.Targets[0] || again.Trigger.Event.CardID != trigger.Event.CardID {
		t.Error("editing a rebuilt item changed what the prompt carries")
	}
	if game.NewAskedItem(nil) != nil || (*game.AskedItem)(nil).StackItem() != nil {
		t.Error("a nil item is not nil data")
	}
}

// Cleansing: "For each land, destroy that land unless any player pays 1
// life." Every leg of the per-object offer is plain data (the objects with
// their epochs, the picks so far) and the consequence a key, so the offer
// is a restore point at the first leg and again at the second, where the
// first seat's pick is on the offer, and the unsaved lands are destroyed
// after the last leg on the restored table.
func TestCleansingOfferIsARestorePointAtEveryLeg(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	a, b := g.Seats[0], g.Seats[1]
	shared := effectstest.CleansingLand(g, a.ID, "Forest")
	mine := effectstest.CleansingLand(g, b.ID, "Island")
	other := effectstest.CleansingLand(g, g.Seats[2].ID, "Swamp")
	effectstest.CleansingCast(t, g)
	leg := effectstest.EachOpenLeg(g)
	if leg == nil || leg.PickThen != cardsch.CleansingDestroysTheUnsaved.Key() || leg.Chooser != a.ID {
		t.Fatalf("the first leg is keyed %v for %v, want the keyed offer asking the active player", leg != nil && leg.PickThen == cardsch.CleansingDestroysTheUnsaved.Key(), leg != nil && leg.Chooser == a.ID)
	}
	if len(leg.AnyPlayerOffer.Objects) != 3 || leg.AnyPlayerOffer.Noun != "land" {
		t.Fatalf("the offer carries %d objects of %q, want the three lands", len(leg.AnyPlayerOffer.Objects), leg.AnyPlayerOffer.Noun)
	}

	// First leg, restored: seat 0 pays for its own land.
	g = effectstest.RestoreRoundTrip(t, g, true)
	a, b = g.Seats[0], g.Seats[1]
	leg = effectstest.EachOpenLeg(g)
	if leg == nil || leg.Chooser != a.ID || leg.PickThen != cardsch.CleansingDestroysTheUnsaved.Key() {
		t.Fatalf("restored first leg: %v", leg != nil)
	}
	if err := g.PassPriority(); !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("a pass while the restored leg was open = %v, want ErrChoicePending", err)
	}
	if err := g.ResolveChooseCards(leg.ID, a.ID, []uuid.UUID{shared}); err != nil {
		t.Fatalf("seat 0's pick on the restored table: %v", err)
	}

	// Second leg, restored: seat 0's pick is on the offer and shown.
	leg = effectstest.EachOpenLeg(g)
	if leg == nil || leg.Chooser != b.ID {
		t.Fatalf("second leg: open = %v, want seat 1's", leg != nil)
	}
	if earlier := leg.EarlierPermanentChoicesForEffect(); len(earlier) != 1 || earlier[0].Seat != a.ID || len(earlier[0].Cards) != 1 || earlier[0].Cards[0] != shared {
		t.Fatalf("second leg shows earlier picks %+v, want seat 0's land", earlier)
	}
	g = effectstest.RestoreRoundTrip(t, g, true)
	a, b = g.Seats[0], g.Seats[1]
	leg = effectstest.EachOpenLeg(g)
	if leg == nil || leg.Chooser != b.ID {
		t.Fatalf("restored second leg: open = %v, want seat 1's", leg != nil)
	}
	if earlier := leg.EarlierPermanentChoicesForEffect(); len(earlier) != 1 || earlier[0].Cards[0] != shared {
		t.Fatalf("restored second leg shows earlier picks %+v, want seat 0's land", earlier)
	}
	// Seat 1 pays for the shared land too and for its own; the rest pay for
	// nothing.
	effectstest.CleansingAnswer(t, g, map[uuid.UUID][]uuid.UUID{b.ID: {shared, mine}})
	if !g.Battlefield.Contains(shared) || !g.Battlefield.Contains(mine) {
		t.Error("a paid-for land was destroyed on the restored table")
	}
	if g.Battlefield.Contains(other) {
		t.Error("the unpaid land survived on the restored table")
	}
	if a.Life != game.StartingLife-1 || b.Life != game.StartingLife-2 {
		t.Errorf("lives %d, %d; want seat 0 down 1 (one land) and seat 1 down 2 (the shared land and its own)", a.Life, b.Life)
	}
}

// Extrapolate the Impossible, the second and third prompts (#4264): at
// three seats the caster names the opponent who chooses (a keyed option
// pick carrying the caster and the revealed pair), and that opponent picks
// one of the pair (a keyed reveal pick). Each is a strict restore point,
// and the chosen card comes into the caster's hand on the restored table.
func TestExtrapolateTheImpossibleOpponentAndRevealPicksAreRestorePoints(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 3, effectstest.OutsideCard("Alpha", "Sorcery"), effectstest.OutsideCard("Beta", "Instant"))
	effectstest.CastCatalogSpell(t, g, "Extrapolate the Impossible", "Sorcery", effectstest.ExtrapolateOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	me := g.Seats[0]
	cards := me.SideboardCards()
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil || c.PickThen != cardsex.ExtrapolatePick.Key() {
		t.Fatalf("the first prompt is %v, want the keyed pick", c != nil)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{cards[0].InstanceID, cards[1].InstanceID}); err != nil {
		t.Fatalf("reveal the pair: %v", err)
	}

	// Second prompt, restored: the caster names the opponent.
	named := effectstest.PendingOfKind(g, game.PendingChoiceOptionPick)
	if named == nil || named.PickThen != cardsex.ExtrapolateOpponent.Key() || len(named.OptionCarry) != 3 || named.OptionCarry[0] != me.ID {
		t.Fatalf("the second prompt is not the keyed option pick carrying the caster and the pair")
	}
	g = effectstest.RestoreRoundTrip(t, g, true)
	me, chosen := g.Seats[0], g.Seats[2]
	if restored := effectstest.PendingOfKind(g, game.PendingChoiceOptionPick); restored == nil || restored.PickThen != cardsex.ExtrapolateOpponent.Key() || len(restored.OptionCarry) != 3 {
		t.Fatalf("restored second prompt is not the keyed option pick carrying the caster and the pair")
	}
	effectstest.AnswerChosenPlayer(t, g, me.ID, chosen.ID)

	// Third prompt, restored: the chosen opponent picks one of the pair.
	rp := effectstest.LatestRevealPickFor(g, chosen.ID)
	if rp == nil || rp.PickThen != cardsex.ExtrapolateTake.Key() || len(rp.ChooseCards) != 2 || rp.FromPlayer != me.ID {
		t.Fatalf("the third prompt is not the keyed reveal pick of the pair, owned by the caster")
	}
	g = effectstest.RestoreRoundTrip(t, g, true)
	me, chosen = g.Seats[0], g.Seats[2]
	rp = effectstest.LatestRevealPickFor(g, chosen.ID)
	if rp == nil || rp.PickThen != cardsex.ExtrapolateTake.Key() || len(rp.ChooseCards) != 2 || rp.FromPlayer != me.ID {
		t.Fatalf("restored third prompt is not the keyed reveal pick of the pair, owned by the caster")
	}
	if err := g.PassPriority(); !errors.Is(err, game.ErrChoicePending) {
		t.Errorf("a pass while the restored reveal pick was open = %v, want ErrChoicePending", err)
	}
	if err := g.ResolveRevealPick(rp.ID, chosen.ID, []uuid.UUID{cards[1].InstanceID}); err != nil {
		t.Fatalf("the opponent's pick on the restored table: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OutsideHandHas(me, "Beta") || effectstest.OutsideHandHas(me, "Alpha") {
		t.Error("the card the opponent chose did not come into the caster's hand alone")
	}
}

// A reveal pick or an opponent pick naming a key this binary lacks is
// refused on restore, as every keyed prompt is.
func TestExtrapolateTheImpossibleSecondAndThirdPromptsWithAnUnknownKeyAreRefused(t *testing.T) {
	for _, key := range []string{cardsex.ExtrapolateOpponent.Key(), cardsex.ExtrapolateTake.Key()} {
		g := effectstest.NewSideboardCatalogGame(t, 3, effectstest.OutsideCard("Alpha", "Sorcery"), effectstest.OutsideCard("Beta", "Instant"))
		effectstest.CastCatalogSpell(t, g, "Extrapolate the Impossible", "Sorcery", effectstest.ExtrapolateOracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		me := g.Seats[0]
		cards := me.SideboardCards()
		c := effectstest.LatestChooseCardsFor(g, me.ID)
		if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{cards[0].InstanceID, cards[1].InstanceID}); err != nil {
			t.Fatal(err)
		}
		if key == cardsex.ExtrapolateTake.Key() {
			effectstest.AnswerChosenPlayer(t, g, me.ID, g.Seats[2].ID)
		}
		effectstest.RefuseUnknownKey(t, g, key)
	}
}

// The Unagi of Kyoshi Island's "Ward—Waterbend {4}": the tap clause is
// rebuilt from its kind and extra cost (RegisterTapCost, PayTap), so the
// ward prompt is a restore point that still offers the taps. Paying with
// four taps on the restored table lets the removal resolve, and declining
// counters it.
func TestWaterbendWardIsARestorePoint(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		unagi := effectstest.PushCatalogPermanent(g, me.ID, "The Unagi of Kyoshi Island",
			"Legendary Creature — Serpent", effectstest.TheUnagiOfKyoshiIslandOracl, false)
		var helpers []uuid.UUID
		for i := 0; i < 4; i++ {
			helpers = append(helpers, effectstest.PushVanillaCreature(g, opp.ID, "Helper", 1, 1))
		}
		effectstest.AdvanceToMain(t, g)
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
		blade := effectstest.CastAtWardedCreature(t, g, opp, unagi)
		for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
		}
		prompt := effectstest.PayUnlessFor(g, opp.ID)
		if prompt == nil || prompt.PayTap == nil || prompt.PayTap.Kind != "waterbend" || prompt.PayTap.Extra != "{4}" || prompt.PickThen == "" {
			t.Fatalf("pay=%v: the ward prompt carries no keyed waterbend {4} clause", pay)
		}

		g = effectstest.RestoreRoundTrip(t, g, true)
		opp = g.Seats[1]
		prompt = effectstest.PayUnlessFor(g, opp.ID)
		if prompt == nil {
			t.Fatalf("pay=%v: no ward prompt after the restore", pay)
		}
		if tc := prompt.PayTapCost(); tc == nil || tc.Extra != "{4}" || tc.Key != "waterbend" || tc.Spec == nil {
			t.Fatalf("pay=%v: the restored prompt carries tap clause %+v, want a rebuilt waterbend {4}", pay, tc)
		}
		if !pay {
			effectstest.AnswerPayUnless(t, g, opp.ID, false)
			effectstest.PassPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(unagi) || !opp.Graveyard.Contains(blade) {
				t.Errorf("declined: the Unagi on the battlefield = %v, the spell in the graveyard = %v; want a countered spell",
					g.Battlefield.Contains(unagi), opp.Graveyard.Contains(blade))
			}
			continue
		}
		if err := g.ResolvePayUnlessWithTaps(prompt.ID, opp.ID, true, helpers); err != nil {
			t.Fatalf("pay with four taps on the restored table: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(unagi) {
			t.Error("the ward was paid on the restored table, so the removal should have resolved")
		}
		for _, id := range helpers {
			if !effectstest.IsTapped(g, id) {
				t.Errorf("%v was not tapped to pay the ward", id)
			}
		}
	}
}

// A prompt carrying a tap clause of a kind this binary lacks is refused on
// restore: restored without it the payment would be plain mana, harder to
// pay than printed.
func TestWaterbendWardWithAnUnknownTapKindIsRefused(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	unagi := effectstest.PushCatalogPermanent(g, me.ID, "The Unagi of Kyoshi Island",
		"Legendary Creature — Serpent", effectstest.TheUnagiOfKyoshiIslandOracl, false)
	effectstest.AdvanceToMain(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	effectstest.CastAtWardedCreature(t, g, opp, unagi)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	prompt := effectstest.PayUnlessFor(g, opp.ID)
	if prompt == nil || prompt.PayTap == nil {
		t.Fatal("the ward prompt carries no tap clause")
	}
	prompt.PayTap.Kind = "no-such-tap-kind"
	snap := g.CaptureSnapshot()
	if _, err := snap.RestoreStrict(); !errors.Is(err, game.ErrUnknownEffectKey) {
		t.Fatalf("restoring a prompt with an unknown tap kind = %v, want ErrUnknownEffectKey", err)
	}
}

// Dance of the Dead: "that player may pay {1}{B}. If the player does,
// untap that creature." The enchanted creature rides the prompt with the
// epoch it had when the trigger resolved (PayUnlessKeyedPrompt.Objects), so
// the offer is a restore point that still untaps that creature, and one
// that left and came back while the payer decided is a new object that
// stays tapped (CR 400.7).
func TestDanceOfTheDeadOfferCarriesTheCreaturesEpochAcrossARestore(t *testing.T) {
	for _, flicker := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		zombie := effectstest.AdGraveyardCreature(opp, "Gravecrawler")
		aura := effectstest.AdAnimate(t, g, "Dance of the Dead", effectstest.DanceOfTheDeadOracle, zombie)
		effectstest.AdvanceToUpkeepOf(t, g, 0)
		if effectstest.TriggerOnStack(g, aura) == nil {
			t.Fatal("Dance of the Dead's upkeep trigger is not on the stack")
		}
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.UntapStaticPendingPay(t, g, me.ID)
		if c.PickThen != effects.MayPayUntapEnchantedCreaturePaid.Key() || len(c.OptionObjects) != 1 || c.OptionObjects[0].ID != zombie {
			t.Fatalf("flicker=%v: the offer carries %d objects under key %q, want the enchanted creature", flicker, len(c.OptionObjects), c.PickThen)
		}
		epoch := c.OptionObjects[0].Epoch

		g = effectstest.RestoreRoundTrip(t, g, true)
		me = g.Seats[0]
		c = effectstest.UntapStaticPendingPay(t, g, me.ID)
		if len(c.OptionObjects) != 1 || c.OptionObjects[0].ID != zombie || c.OptionObjects[0].Epoch != epoch {
			t.Fatalf("flicker=%v: restored offer carries %+v, want the creature at epoch %d", flicker, c.OptionObjects, epoch)
		}
		if flicker {
			g.BeginEventBatchForTest()
			effectstest.FlickerInResponse(t, g, zombie)
			if err := g.TapCard(zombie, true); err != nil {
				t.Fatal(err)
			}
		}
		effectstest.FillPoolColored(me, "B", 2)
		if err := g.ResolvePayUnless(c.ID, me.ID, true); err != nil {
			t.Fatalf("flicker=%v: answer: %v", flicker, err)
		}
		if tapped := effectstest.AdCard(t, g, zombie).Tapped; tapped != flicker {
			t.Errorf("flicker=%v: the creature is tapped = %v, want %v (a new object stays tapped, the same one untaps)", flicker, tapped, flicker)
		}
	}
}
