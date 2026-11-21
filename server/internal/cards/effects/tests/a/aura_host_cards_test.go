package testsa

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// The headline case: Sun Titan returns Animate Dead, which chooses a
// creature card in any graveyard as it enters (not a battlefield
// creature, not itself), enters attached to it, and its enters trigger
// then returns that card under its controller's control (ADR 0154).
func TestSunTitanReturnsAnimateDeadWhichChoosesAGraveyardCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aura := effectstest.AhAnimateDead(me)
	mine := effectstest.AdGraveyardCreature(me, "My Ghoul")
	theirs := effectstest.AdGraveyardCreature(opp, "Their Zombie")
	bear := effectstest.PushTribalCreature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)

	titan := effectstest.SunTitanReturns(t, g, aura)
	c := effectstest.AuraHostChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("Animate Dead did not ask what it enchants: %+v", g.PendingChoices)
	}
	if !effectstest.HasID(c.PickTargetCards, mine) || !effectstest.HasID(c.PickTargetCards, theirs) {
		t.Errorf("both creature cards in graveyards are offered: %v", c.PickTargetCards)
	}
	for _, id := range []uuid.UUID{bear, titan, aura} {
		if effectstest.HasID(c.PickTargetCards, id) {
			t.Errorf("%s is not a creature card in a graveyard", effectstest.CardByID(g, id).Name)
		}
	}
	if effectstest.AdZone(g, aura) != game.ZoneGraveyard {
		t.Fatal("nothing moves until the question is answered")
	}

	effectstest.AnswerAuraHost(t, g, me.ID, effectstest.CardRef(theirs))
	a := effectstest.AdCard(t, g, aura)
	if effectstest.AdZone(g, aura) != game.ZoneBattlefield || !a.IsAttachedTo(theirs) {
		t.Fatalf("Animate Dead enters attached to the chosen card: zone=%v host=%+v", effectstest.AdZone(g, aura), a.AttachedTo)
	}
	if effectstest.TriggerOnStack(g, aura) == nil {
		t.Fatal("its enters trigger waits on the stack")
	}
	effectstest.PassPriorityAroundTable(t, g)
	z := effectstest.AdCard(t, g, theirs)
	if effectstest.AdZone(g, theirs) != game.ZoneBattlefield || z.Controller != me.ID {
		t.Fatalf("the opponent's creature returns under my control: zone=%v controller=%v", effectstest.AdZone(g, theirs), z.Controller)
	}
	if a := effectstest.AdCard(t, g, aura); !a.IsAttachedTo(theirs) || effectstest.AdZone(g, aura) != game.ZoneBattlefield {
		t.Fatal("Animate Dead stays attached to the returned creature")
	}
	if got := effectstest.EffectivePower(t, g, theirs); got != 2 {
		t.Errorf("enchanted creature gets -1/-0: power %d, want 2", got)
	}
}

// CR 303.4g: with no creature card in any graveyard there is nothing it
// can enchant, so it stays in the graveyard — no prompt, no trip through
// the battlefield, no enters trigger.
func TestAnimateDeadWithNothingToEnchantStaysInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	aura := effectstest.AhAnimateDead(me)
	effectstest.PushTribalCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	effectstest.SunTitanReturns(t, g, aura)
	if effectstest.AuraHostChoiceFor(g, me.ID) != nil {
		t.Fatal("nothing to choose, nothing to ask")
	}
	if effectstest.AdZone(g, aura) != game.ZoneGraveyard {
		t.Fatalf("Animate Dead stays in the graveyard, is in %v", effectstest.AdZone(g, aura))
	}
	for _, ev := range g.Events {
		if ev.Kind == game.EventETB && ev.CardID == aura {
			t.Fatal("it never entered")
		}
	}
}

// The predicate: the Aura's own enchant clause (Pacifism's "enchant
// creature": no land, no artifact), protection refuses a host (CR
// 702.16c), and hexproof does not — choosing a host is not targeting.
// One legal host is a forced choice and asks nothing.
func TestAuraHostFollowsTheEnchantClauseAndProtectionButNotHexproof(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aura := effectstest.AhPacifism(me)
	hexproof := effectstest.PushTribalCreature(g, opp.ID, "Hexproof Bear", "Creature — Bear", 2, 2, "hexproof")
	proWhite := effectstest.PushTribalCreature(g, opp.ID, "White Ward Bear", "Creature — Bear", 2, 2, "protection from white")
	rock := effectstest.PushCatalogPermanent(g, opp.ID, "Rock", "Artifact", "", false)
	land := effectstest.PushCatalogPermanent(g, opp.ID, "Forest", "Basic Land — Forest", "", false)

	titan := effectstest.SunTitanReturns(t, g, aura)
	c := effectstest.AuraHostChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("two legal hosts: the player chooses")
	}
	if !effectstest.HasID(c.PickTargetCards, hexproof) || !effectstest.HasID(c.PickTargetCards, titan) {
		t.Errorf("a hexproof creature and my own creature are legal hosts: %v", c.PickTargetCards)
	}
	for _, id := range []uuid.UUID{proWhite, rock, land} {
		if effectstest.HasID(c.PickTargetCards, id) {
			t.Errorf("%s is not a legal host for a white enchant-creature Aura", effectstest.CardByID(g, id).Name)
		}
	}
	if len(c.PickTargetPlayers) != 0 {
		t.Errorf("enchant creature offers no player: %v", c.PickTargetPlayers)
	}
	if err := g.ResolveAuraHost(c.ID, me.ID, effectstest.CardRef(proWhite)); err == nil {
		t.Fatal("an answer outside the legal set is refused")
	}
	effectstest.AnswerAuraHost(t, g, me.ID, effectstest.CardRef(hexproof))
	if host := effectstest.AttachmentHostOf(t, g, aura); host != effectstest.CardRef(hexproof) {
		t.Fatalf("Pacifism enchants the hexproof creature, got %+v", host)
	}
}

// Replenish's two Auras each ask their own question, among the
// permanents already there — not the enchantment creature Replenish is
// returning at the same time — and each enters attached to its answer.
func TestReplenishReturnsTwoAurasAndEachChoosesItsHost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	first := effectstest.AhPacifism(me)
	second := effectstest.AhPacifism(me)
	god := effectstest.B17GraveyardCard(me, "God", "Legendary Enchantment Creature — God", "{3}{W}")
	a := effectstest.PushTribalCreature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.PushTribalCreature(g, opp.ID, "Bear B", "Creature — Bear", 3, 3)

	effectstest.CastCatalogSpell(t, g, "Replenish", "Sorcery", effectstest.B36ReplenishOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.AuraHostChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the first Aura asks")
	}
	if effectstest.HasID(c.PickTargetCards, god) {
		t.Error("an Aura can't enchant another permanent Replenish returns")
	}
	if g.Battlefield.Contains(god) {
		t.Fatal("nothing lands until every Aura has chosen")
	}
	effectstest.AnswerAuraHost(t, g, me.ID, effectstest.CardRef(a))
	if effectstest.AuraHostChoiceFor(g, me.ID) == nil {
		t.Fatal("the second Aura asks its own question")
	}
	effectstest.AnswerAuraHost(t, g, me.ID, effectstest.CardRef(b))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(god) {
		t.Error("the enchantment creature returns")
	}
	hosts := map[game.TargetRef]bool{effectstest.AttachmentHostOf(t, g, first): true, effectstest.AttachmentHostOf(t, g, second): true}
	if !hosts[effectstest.CardRef(a)] || !hosts[effectstest.CardRef(b)] {
		t.Fatalf("each Pacifism enchants the creature chosen for it: %v", hosts)
	}
}

// A graveyard clause and a simultaneous entry: Replenish returns Animate
// Dead and an enchantment creature card together. That card is leaving
// the graveyard in the same entry, so it is not a host; the creature card
// in the opponent's graveyard is the only one, a forced choice, and it
// comes back under Animate Dead's controller.
func TestReplenishedAnimateDeadCantEnchantACardReplenishReturns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aura := effectstest.AhAnimateDead(me)
	god := effectstest.B17GraveyardCard(me, "God", "Legendary Enchantment Creature — God", "{3}{W}")
	ghoul := effectstest.AdGraveyardCreature(opp, "Their Ghoul")
	effectstest.CastCatalogSpell(t, g, "Replenish", "Sorcery", effectstest.B36ReplenishOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.AuraHostChoiceFor(g, me.ID); c != nil {
		t.Fatalf("one legal host, nothing to ask: %v", c.PickTargetCards)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(god) {
		t.Fatal("the enchantment creature returns")
	}
	if effectstest.AdZone(g, ghoul) != game.ZoneBattlefield || effectstest.AdCard(t, g, ghoul).Controller != me.ID {
		t.Fatalf("Animate Dead returns the opponent's Ghoul under my control: zone=%v", effectstest.AdZone(g, ghoul))
	}
	if !effectstest.AdCard(t, g, aura).IsAttachedTo(ghoul) {
		t.Fatal("Animate Dead stays on the Ghoul")
	}
}

// Undo mid-choice: a clone taken while the question is open restores
// the question, and answering it differently is honoured. Since #4253
// the paused entry's frame is data, so the table is a restore point as
// well (TestAnAuraHostQuestionIsARestorePoint answers it after one).
func TestAuraHostSurvivesAnUndoMidChoice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aura := effectstest.AhAnimateDead(me)
	mine := effectstest.AdGraveyardCreature(me, "My Ghoul")
	theirs := effectstest.AdGraveyardCreature(opp, "Their Zombie")
	effectstest.SunTitanReturns(t, g, aura)
	if effectstest.AuraHostChoiceFor(g, me.ID) == nil {
		t.Fatal("setup: Animate Dead asks")
	}
	if snap := g.CaptureSnapshot(); !snap.Restorable() {
		t.Errorf("the open aura_host question is not a restore point: %+v", snap.Continuations)
	}

	snap := g.Clone()
	effectstest.AnswerAuraHost(t, g, me.ID, effectstest.CardRef(theirs))
	if !effectstest.AdCard(t, g, aura).IsAttachedTo(theirs) {
		t.Fatal("the first answer attaches")
	}

	g.RestoreFrom(snap)
	me = g.Seats[0]
	if effectstest.AdZone(g, aura) != game.ZoneGraveyard || effectstest.AuraHostChoiceFor(g, me.ID) == nil {
		t.Fatalf("undo puts the question back with the Aura in the graveyard: zone=%v", effectstest.AdZone(g, aura))
	}
	effectstest.AnswerAuraHost(t, g, me.ID, effectstest.CardRef(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.AdZone(g, mine) != game.ZoneBattlefield || effectstest.AdZone(g, theirs) != game.ZoneGraveyard {
		t.Fatalf("the restored answer returns my Ghoul: mine=%v theirs=%v", effectstest.AdZone(g, mine), effectstest.AdZone(g, theirs))
	}
}

// The bot's move list: one answer per legal host, each accepted.
func TestAuraHostEnumeratorOffersEveryHost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aura := effectstest.AhAnimateDead(me)
	mine := effectstest.AdGraveyardCreature(me, "My Ghoul")
	theirs := effectstest.AdGraveyardCreature(opp, "Their Zombie")
	effectstest.SunTitanReturns(t, g, aura)
	c := effectstest.AuraHostChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("setup: Animate Dead asks")
	}
	offered := map[uuid.UUID]bool{}
	var mineMove *legal.Move
	moves := legal.EnumerateFor(g, me.ID)
	for i, m := range moves {
		if m.Type != legal.TypeResolveChoice {
			continue
		}
		var p struct {
			ChoiceID string `json:"choice_id"`
			Target   *struct {
				Kind string `json:"kind"`
				ID   string `json:"id"`
			} `json:"target"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil || p.ChoiceID != c.ID.String() || p.Target == nil {
			continue
		}
		id := uuid.MustParse(p.Target.ID)
		offered[id] = true
		if id == mine {
			mineMove = &moves[i]
		}
	}
	if len(offered) != 2 || !offered[mine] || !offered[theirs] || mineMove == nil {
		t.Fatalf("one answer per creature card in a graveyard, got %v", offered)
	}
	// The plural form the client's banner sends is routed too; a
	// two-ref answer is refused and leaves the prompt open.
	two, _ := json.Marshal(map[string]any{"choice_id": c.ID.String(), "targets": []map[string]string{
		{"kind": "card", "id": mine.String()}, {"kind": "card", "id": theirs.String()},
	}})
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(legal.TypeResolveChoice), Player: me.ID, Caller: me.ID, Params: two}); err == nil {
		t.Fatal("an Aura enchants one host")
	}
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(mineMove.Type), Player: me.ID, Caller: me.ID, Params: mineMove.Params}); err != nil {
		t.Fatalf("the enumerated answer is refused: %v", err)
	}
	if !effectstest.AdCard(t, g, aura).IsAttachedTo(mine) {
		t.Fatal("an enumerated answer is accepted")
	}
}

// Brilliant Restoration: the Aura returns attached. One legal host is a
// forced choice, so it asks nothing.
func TestBrilliantRestorationReturnsAnAuraAttached(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aura := effectstest.AhPacifism(me)
	bear := effectstest.PushTribalCreature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Brilliant Restoration", "Sorcery", effectstest.B21BrilliantRestorationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.AuraHostChoiceFor(g, me.ID) != nil {
		t.Fatal("one host: nothing to ask")
	}
	if host := effectstest.AttachmentHostOf(t, g, aura); host != effectstest.CardRef(bear) {
		t.Fatalf("Pacifism returns attached to the only creature, got %+v", host)
	}
}

// Open the Vaults: an opponent's Aura comes back under its owner, and
// its owner chooses what it enchants.
func TestOpenTheVaultsLetsEachAurasOwnerChoose(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aura := effectstest.AhPacifism(opp)
	mineBear := effectstest.PushTribalCreature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.PushTribalCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Open the Vaults", "Sorcery", effectstest.B38OpenTheVaultsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.AuraHostChoiceFor(g, me.ID) != nil || effectstest.AuraHostChoiceFor(g, opp.ID) == nil {
		t.Fatal("the Aura's owner, not the caster, chooses")
	}
	effectstest.AnswerAuraHost(t, g, opp.ID, effectstest.CardRef(mineBear))
	effectstest.PassPriorityAroundTable(t, g)
	a := effectstest.AdCard(t, g, aura)
	if a.Controller != opp.ID || !a.IsAttachedTo(mineBear) {
		t.Fatalf("their Pacifism returns under them on my Bear: controller=%v host=%+v", a.Controller, a.AttachedTo)
	}
}

// Carmen's attack returns an Aura card, which chooses its host.
func TestCarmenReturnsAnAuraThatChoosesItsHost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	carmen := effectstest.B12Push(g, me.ID, "Carmen, Cruel Skymarcher", "Legendary Creature — Vampire Soldier", effectstest.B26CarmenTestOracle, 2, 2)
	aura := effectstest.AhPacifism(me)
	bear := effectstest.PushTribalCreature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, carmen)
	effectstest.PickCard(t, g, me.ID, aura)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerAuraHost(t, g, me.ID, effectstest.CardRef(bear))
	if host := effectstest.AttachmentHostOf(t, g, aura); host != effectstest.CardRef(bear) {
		t.Fatalf("Pacifism enters on the chosen Bear, got %+v", host)
	}
}

// It That Betrays takes an opponent's sacrificed Aura, and ITS controller
// — the player it enters under — chooses the host.
func TestItThatBetraysTakesASacrificedAuraAndChoosesItsHost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	betrays := effectstest.PushAnnCatalogCreature(g, me.ID, "It That Betrays", effectstest.ItThatBetraysOracle, 11, 11)
	theirBear := effectstest.PushTribalCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	aura := effectstest.PushCatalogPermanent(g, opp.ID, "Pacifism", effectstest.AuraTypeLine, effectstest.PacifismOracle, false)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(aura, effectstest.CardRef(betrays)); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})
	if err := g.SacrificePermanent(opp.ID, aura); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.AuraHostChoiceFor(g, opp.ID) != nil {
		t.Fatal("the Aura enters under my control, so I choose")
	}
	effectstest.AnswerAuraHost(t, g, me.ID, effectstest.CardRef(theirBear))
	a := effectstest.AdCard(t, g, aura)
	if a.Controller != me.ID || !a.IsAttachedTo(theirBear) {
		t.Fatalf("their Pacifism is mine, on their Bear: controller=%v host=%+v", a.Controller, a.AttachedTo)
	}
}

// Nazgûl Battle-Mace: the same, through the equipped creature's granted
// trigger and the "unless that player pays 3 life" prompt.
func TestNazgulBattleMaceTakesASacrificedAuraThatChoosesItsHost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.PushAnnCatalogCreature(g, me.ID, "Knight", "", 3, 3)
	mace := effectstest.PushCatalogPermanent(g, me.ID, "Nazgûl Battle-Mace", "Artifact — Equipment", effectstest.NazgulBattleMaceOracle, false)
	theirBear := effectstest.PushTribalCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	aura := effectstest.PushCatalogPermanent(g, opp.ID, "Pacifism", effectstest.AuraTypeLine, effectstest.PacifismOracle, false)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(mace, effectstest.CardRef(host)); err != nil {
			t.Fatalf("attach: %v", err)
		}
		if err := g.AttachForEffect(aura, effectstest.CardRef(host)); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})
	// The grant is read off the host's characteristics, as the Mace's
	// own test reads them before the sacrifice.
	if c := effectstest.AnnCard(t, g, host); !game.HasKeyword(&c, "menace") {
		t.Fatal("setup: the Mace's grant")
	}
	if err := g.SacrificePermanent(opp.ID, aura); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, opp.ID, 1) // don't pay
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerAuraHost(t, g, me.ID, effectstest.CardRef(theirBear))
	a := effectstest.AdCard(t, g, aura)
	if a.Controller != me.ID || !a.IsAttachedTo(theirBear) {
		t.Fatalf("the Aura is mine, on the chosen Bear: controller=%v host=%+v", a.Controller, a.AttachedTo)
	}
}

// Ondu Spiritdancer: a token copy of an Aura chooses its host as it is
// created (CR 303.4f), and stays on the battlefield attached.
func TestOnduSpiritdancerCopiesAnAuraOntoAChosenHost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ondu := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ondu Spiritdancer", TypeLine: "Creature — Kor Cleric",
		OracleID: effectstest.OnduSpiritdancerOracle, Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	first := effectstest.PushTribalCreature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	second := effectstest.PushTribalCreature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	effectstest.CastFromHand(t, g, game.Card{Name: "Pacifism", TypeLine: effectstest.AuraTypeLine, OracleID: effectstest.PacifismOracle},
		game.CastSpellParams{Targets: []game.TargetRef{effectstest.CardRef(first)}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.AuraHostChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the token copy asks what it enchants")
	}
	if !effectstest.HasID(c.PickTargetCards, ondu) || !effectstest.HasID(c.PickTargetCards, second) {
		t.Errorf("every creature is a legal host: %v", c.PickTargetCards)
	}
	effectstest.AnswerAuraHost(t, g, me.ID, effectstest.CardRef(second))
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CardsNamed(g, "Pacifism", me.ID); n != 2 {
		t.Fatalf("%d Pacifisms, want the card and its token copy", n)
	}
	onSecond := false
	for _, card := range g.Battlefield.Cards {
		if card.Name == "Pacifism" && card.IsAttachedTo(second) {
			onSecond = true
		}
	}
	if !onSecond {
		t.Error("the token copy enchants the chosen Bear")
	}
}

// Zur the Enchanter: a fetched Aura chooses its host.
func TestZurFetchesAnAuraThatChoosesItsHost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zur := effectstest.B12Push(g, me.ID, "Zur the Enchanter", "Legendary Creature — Human Wizard", effectstest.B26ZurTheEnchanterOracle, 1, 4)
	bear := effectstest.PushTribalCreature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	lib := effectstest.SeedSearchLibrary(me, game.Card{Name: "Pacifism", TypeLine: effectstest.AuraTypeLine, OracleID: effectstest.PacifismOracle, ManaCost: "{1}{W}", Colors: []string{"W"}})
	effectstest.DeclareAttack(t, g, opp.ID, zur)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, lib[0])
	c := effectstest.AuraHostChoiceFor(g, me.ID)
	if c == nil || !effectstest.HasID(c.PickTargetCards, zur) || !effectstest.HasID(c.PickTargetCards, bear) {
		t.Fatalf("the fetched Aura asks between Zur and the Bear: %+v", c)
	}
	effectstest.AnswerAuraHost(t, g, me.ID, effectstest.CardRef(bear))
	if host := effectstest.AttachmentHostOf(t, g, lib[0]); host != effectstest.CardRef(bear) {
		t.Fatalf("Pacifism enters on the Bear, got %+v", host)
	}
}

// Guardian Sunmare: the saddled attack's fetch, the same question.
func TestGuardianSunmareFetchesAnAuraThatChoosesItsHost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mount := effectstest.PushMountForTest(g, me.ID, "Guardian Sunmare", effectstest.GuardianSunmareOracle, 4)
	saddler := effectstest.PushCrewerForTest(g, me.ID, "Big Saddler", 4)
	bear := effectstest.PushTribalCreature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	aura := effectstest.PushLibraryCard(me, game.Card{Name: "Pacifism", TypeLine: effectstest.AuraTypeLine, OracleID: effectstest.PacifismOracle, ManaCost: "{1}{W}", Colors: []string{"W"}})
	effectstest.SaddleThenAttack(t, g, me.ID, opp.ID, mount, saddler)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, aura)
	effectstest.AnswerAuraHost(t, g, me.ID, effectstest.CardRef(bear))
	if host := effectstest.AttachmentHostOf(t, g, aura); host != effectstest.CardRef(bear) {
		t.Fatalf("Pacifism enters on the Bear, got %+v", host)
	}
}

// Armored Skyhunter offers an Aura card among the six now, and the Aura
// chooses its host; it is not offered the Equipment attach.
func TestArmoredSkyhunterPutsAnAuraThatChoosesItsHost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	sky := effectstest.B12Push(g, me.ID, "Armored Skyhunter", "Creature — Cat Knight", effectstest.PlSkyhunterOracle, 3, 3)
	bear := effectstest.PushTribalCreature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	aura := effectstest.PushLibraryCard(me, game.Card{Name: "Pacifism", TypeLine: effectstest.AuraTypeLine, OracleID: effectstest.PacifismOracle, ManaCost: "{1}{W}", Colors: []string{"W"}})
	effectstest.DeclareAttack(t, g, opp.ID, sky)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, aura)
	effectstest.AnswerAuraHost(t, g, me.ID, effectstest.CardRef(bear))
	if host := effectstest.AttachmentHostOf(t, g, aura); host != effectstest.CardRef(bear) {
		t.Fatalf("Pacifism enters on the Bear, got %+v", host)
	}
	if c := effectstest.ChooseCardsChoiceFor(g, me.ID); c != nil {
		t.Error("an Aura is not offered the Equipment's attach")
	}
}

// Light-Paws names the host, so nothing is asked, and an Aura that can't
// enchant Light-Paws stays in the library (CR 303.4h).
func TestLightPawsLeavesAnAuraThatCantEnchantItInTheLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	paws := effectstest.B29Push(g, me.ID, "Light-Paws, Emperor's Voice", "Legendary Creature — Fox Advisor", effectstest.B29LightPawsOracle, "{1}{W}", 2, 2, "W")
	bear := effectstest.PushTribalCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	// A land it COULD enchant, were it choosing: it is not.
	effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	ground := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Fertile Ground", TypeLine: effectstest.AuraTypeLine, OracleID: effectstest.FertileGroundOracle, ManaCost: "{1}{G}", Colors: []string{"G"}})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	effectstest.AdvanceToMain(t, g)
	effectstest.B29CastAs(t, g, me, "Ethereal Armor", effectstest.AuraTypeLine, "", "{W}{W}", game.CastSpellParams{
		Targets: []game.TargetRef{effectstest.CardRef(bear)},
	})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, ground)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.AuraHostChoiceFor(g, me.ID) != nil {
		t.Fatal("the effect names the host: nothing to ask")
	}
	if !me.Library.Contains(ground) || g.Battlefield.Contains(ground) || me.Graveyard.Contains(ground) {
		t.Fatalf("an enchant-land Aura can't enchant Light-Paws (%v) and stays in the library", paws)
	}
}

// A copy of an Aura spell becomes a token attached to the copy's target
// (CR 707.10, 608.3f). The copy NAMES its host, so it is not asked CR
// 303.4f's question even with several legal hosts on the board.
func TestACopyOfAnAuraSpellEntersAttachedToItsTargetWithoutAsking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushTribalCreature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.PushTribalCreature(g, opp.ID, "Other Bear", "Creature — Bear", 2, 2)
	spell := effectstest.CastFromHand(t, g, game.Card{Name: "Pacifism", TypeLine: effectstest.AuraTypeLine, OracleID: effectstest.PacifismOracle},
		game.CastSpellParams{Targets: []game.TargetRef{effectstest.CardRef(bear)}})
	g.WithWriteLock(func() {
		if err := g.CopySpellForEffect(spell, me.ID, false, nil); err != nil {
			t.Fatalf("copy: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.AuraHostChoiceFor(g, me.ID) != nil {
		t.Fatal("the copy names its host: nothing to ask")
	}
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Pacifism" && c.IsAttachedTo(bear) {
			n++
		}
	}
	if n != 2 {
		t.Fatalf("the Aura and its token copy both enchant the target, got %d", n)
	}
}

// Gift of Immortality names its host too, and is not asked.
func TestGiftOfImmortalityReturnsAttachedWithoutAsking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.PushTribalCreature(g, me.ID, "Other Creature", "Creature — Elf", 1, 1)
	gift := effectstest.AuraCast(t, g, "Gift of Immortality", giftOfImmortalityOracle, bear)
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.AuraHostChoiceFor(g, me.ID) != nil {
		t.Fatal("Gift names its host: nothing to ask")
	}
	ids := effectstest.BattlefieldIDsNamed(g, "Bear")
	if len(ids) != 1 || effectstest.AttachmentHostOf(t, g, gift) != effectstest.CardRef(ids[0]) {
		t.Fatalf("Gift returns attached to the creature it brought back: %+v", effectstest.AttachmentHostOf(t, g, gift))
	}
}
