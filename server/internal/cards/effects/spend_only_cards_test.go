package effects

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// spend_only_cards_test.go — #1600's proof cards for mana only some
// spells may use and costs only some mana may pay (ADR 0040's
// 2026-10-03 amendment): Throne of Eldraine, Crypt Rats, Crimson
// Hellkite, Pillar of the Paruns and Obsidian Obelisk. The engine half
// is game/spend_only_test.go. The agreement tests at the bottom ask one
// board three questions — the view's legal_actions, the bot's
// enumerator and the real payment — and require one answer.

const (
	throneOfEldraineOracle  = "5b9a2b81-a645-43be-8001-03817ac210ca"
	cryptRatsOracle         = "104095ed-55e3-408e-bf70-4fe06bb16d2f"
	crimsonHellkiteOracle   = "37a1f340-10d5-407e-bf2b-cbbbb6cab985"
	pillarOfTheParunsOracle = "677b8ce7-f922-4ee3-b311-f199da9b352b"
	obsidianObeliskOracle   = "92a5e50d-8477-4654-9c1b-219e4fe3c5ab"
)

// basics puts n untapped basic lands of `land` (Plains, Island, …)
// under p's control and returns them.
func basics(g *game.Game, p *game.Player, land string, n int) []uuid.UUID {
	out := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, pushCatalogPermanent(g, p.ID, land, "Basic Land — "+land, "", false))
	}
	return out
}

// throneFor seeds a Throne of Eldraine under p with `color` chosen.
func throneFor(t *testing.T, g *game.Game, p *game.Player, color string) uuid.UUID {
	t.Helper()
	return pushChosenColorPermanent(t, g, p.ID, "Throne of Eldraine", "Legendary Artifact", throneOfEldraineOracle, color)
}

// --- Throne of Eldraine ---------------------------------------------

// The colour is chosen as it enters (#742's prompt, ColorForMana), and
// the mana ability then makes four of it, restricted.
func TestThroneOfEldraineChoosesAColorAndAddsFourOfIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	throne := castCatalogSpell(t, g, "Throne of Eldraine", "Legendary Artifact", throneOfEldraineOracle, nil)
	passPriorityAroundTable(t, g)
	prompt := answerColor(t, g, me.ID, "G")
	if prompt.ColorPurpose != game.ColorForMana {
		t.Errorf("the prompt declared purpose %q, want %q", prompt.ColorPurpose, game.ColorForMana)
	}
	if c, _ := battlefieldCard(g, throne); c.ChosenColor != "G" {
		t.Fatalf("chosen colour %q, want G", c.ChosenColor)
	}
	if err := g.ActivateManaAbility(me.ID, throne, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Throne: %v", err)
	}
	if got := poolColors(me); !slices.Equal(got, []string{"G", "G", "G", "G"}) {
		t.Fatalf("pool = %v, want four green", got)
	}
	for _, tok := range me.ManaPool {
		if !slices.Contains(tok.Restrictions, game.ManaRestrictMonocolored) || !slices.Contains(tok.Restrictions, game.ManaRestrictColor("G")) {
			t.Errorf("token %+v is not restricted to monocolored green spells", tok)
		}
	}
}

// "Spend this mana only to cast monocolored spells of that color": a
// mono-white spell yes; a hybrid white-blue spell (CR 202.2d — both
// colours), a colourless one and a blue one no.
func TestThroneOfEldraineManaCastsOnlyMonocoloredSpellsOfItsColor(t *testing.T) {
	g, me, _ := spendTable(t)
	throne := throneFor(t, g, me, "W")
	if err := g.ActivateManaAbility(me.ID, throne, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Throne: %v", err)
	}
	hybrid := handSpell(me, "Hybrid Spell", "Sorcery", "{W/U}{W/U}")
	colorless := handSpell(me, "Colorless Spell", "Artifact", "{2}")
	blue := handSpell(me, "Blue Spell", "Sorcery", "{U}")
	white := handSpell(me, "White Spell", "Sorcery", "{1}{W}{W}")
	refusedForMana(t, g.CastSpell(me.ID, hybrid, game.CastSpellParams{Strict: true}), "Throne mana for a white-blue hybrid spell")
	refusedForMana(t, g.CastSpell(me.ID, colorless, game.CastSpellParams{Strict: true}), "Throne mana for a colorless spell")
	refusedForMana(t, g.CastSpell(me.ID, blue, game.CastSpellParams{Strict: true}), "white Throne mana for a blue spell")
	if err := g.CastSpell(me.ID, white, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("Throne mana for a mono-white spell: %v", err)
	}
	if got := poolColors(me); len(got) != 1 {
		t.Errorf("pool = %v, want one white left", got)
	}
}

// The Throne's mana is for SPELLS: it cannot pay an activated ability,
// even one whose cost is its colour.
func TestThroneOfEldraineManaCannotPayAnAbility(t *testing.T) {
	g, me, _ := spendTable(t)
	throne := throneFor(t, g, me, "B")
	pest := pushCatalogPermanent(g, me.ID, "Pestilence", "Enchantment", pestilenceOracle, false)
	if err := g.ActivateManaAbility(me.ID, throne, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Throne: %v", err)
	}
	refusedForMana(t, g.ActivateCatalogAbility(me.ID, pest, 0, game.ActivateAbilityParams{Strict: true}), "Throne mana for Pestilence's {B}")
}

// "Spend only mana of the chosen color to activate this ability",
// through the auto-tapper: two Plains and three Islands cannot pay the
// {3} and nothing is tapped; a third Plains pays it, the Islands stay
// untapped, and the ability draws two.
func TestThroneOfEldraineDrawAutoTapsOnlyTheChosenColor(t *testing.T) {
	g, me, _ := spendTable(t)
	throne := throneFor(t, g, me, "W")
	plains := basics(g, me, "Plains", 2)
	islands := basics(g, me, "Island", 3)
	refusedForMana(t, g.ActivateCatalogAbility(me.ID, throne, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}),
		"two Plains and three Islands for a white-only {3}")
	for _, id := range append(append([]uuid.UUID{throne}, plains...), islands...) {
		if b31Tapped(t, g, id) {
			t.Fatalf("a refused activation tapped %s", id)
		}
	}
	plains = append(plains, basics(g, me, "Plains", 1)...)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, throne, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("three Plains for a white-only {3}: %v", err)
	}
	for _, id := range plains {
		if !b31Tapped(t, g, id) {
			t.Errorf("Plains %s was not tapped", id)
		}
	}
	for _, id := range islands {
		if b31Tapped(t, g, id) {
			t.Errorf("Island %s was tapped for a white-only cost", id)
		}
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d → %d, want +2", hand, got)
	}
}

// From the pool: two white and a blue do not pay it, three white do.
func TestThroneOfEldraineDrawFromThePool(t *testing.T) {
	g, me, _ := spendTable(t)
	throne := throneFor(t, g, me, "W")
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}{W}{U}"); err != nil {
		t.Fatal(err)
	}
	refusedForMana(t, g.ActivateCatalogAbility(me.ID, throne, 0, game.ActivateAbilityParams{Strict: true}), "{W}{W}{U} for a white-only {3}")
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, throne, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("three white for a white-only {3}: %v", err)
	}
	if got := poolColors(me); !slices.Equal(got, []string{"U"}) {
		t.Errorf("pool = %v, want the blue left over", got)
	}
}

// Under Chromatic Orrery the two halves part ways (CR 609.4b): the draw
// ability's cost may be paid with any mana, but the Throne's own mana
// still casts only monocolored spells of its colour.
func TestThroneOfEldraineUnderChromaticOrrery(t *testing.T) {
	g, me, _ := spendTable(t)
	throne := throneFor(t, g, me, "W")
	pushCatalogPermanent(g, me.ID, "Chromatic Orrery", "Legendary Artifact", chromaticOrreryOracle, false)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, throne, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("blue and colorless for a white-only {3} under the Orrery: %v", err)
	}
	passPriorityAroundTable(t, g)
	untap(g, throne)
	if err := g.ActivateManaAbility(me.ID, throne, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Throne: %v", err)
	}
	blue := handSpell(me, "Blue Spell", "Sorcery", "{U}")
	refusedForMana(t, g.CastSpell(me.ID, blue, game.CastSpellParams{Strict: true}), "white Throne mana for a blue spell under the Orrery")
}

// The restriction is game state on both halves — the chosen colour on
// the permanent and the tags on the tokens — so it comes back from a
// snapshot whole.
func TestThroneOfEldraineSurvivesASnapshotRoundTrip(t *testing.T) {
	g, me, _ := spendTable(t)
	throne := throneFor(t, g, me, "W")
	if err := g.ActivateManaAbility(me.ID, throne, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Throne: %v", err)
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}{W}{U}"); err != nil {
		t.Fatal(err)
	}
	blob, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var snap game.GameSnapshot
	if err := json.Unmarshal(blob, &snap); err != nil {
		t.Fatal(err)
	}
	r, err := snap.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	rme := r.Seats[0]
	if c, _ := battlefieldCard(r, throne); c.ChosenColor != "W" {
		t.Fatalf("restored chosen colour %q, want W", c.ChosenColor)
	}
	untap(r, throne)
	// Four restricted white, two plain white and a blue: the restricted
	// four cannot pay an activation, so the draw is one white short.
	refusedForMana(t, r.ActivateCatalogAbility(rme.ID, throne, 0, game.ActivateAbilityParams{Strict: true}), "restored: the Throne's own mana for its draw")
	// The unrestricted mana pays a hybrid spell; the restricted four
	// must not be what pays it, although the solver spends restricted
	// mana first wherever it may.
	hybrid := handSpell(rme, "Hybrid Spell", "Sorcery", "{W/U}{W/U}")
	if err := r.CastSpell(rme.ID, hybrid, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("restored: unrestricted mana for a hybrid spell: %v", err)
	}
	restricted := 0
	for _, tok := range rme.ManaPool {
		if len(tok.Restrictions) > 0 {
			restricted++
		}
	}
	if restricted != 4 {
		t.Errorf("restored: %d restricted tokens left, want all four — a hybrid spell is multicolored", restricted)
	}
}

// --- Crypt Rats and Crimson Hellkite: "Spend only <colour> mana on X" -

// xValuesOffered is every X the enumerator offers for source's ability.
func xValuesOffered(t *testing.T, g *game.Game, seat, source uuid.UUID) []int {
	t.Helper()
	var xs []int
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type != "activate_ability" || m.Source != source {
			continue
		}
		var p struct {
			XValue int `json:"x_value"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("move params: %v", err)
		}
		if !slices.Contains(xs, p.XValue) {
			xs = append(xs, p.XValue)
		}
	}
	slices.Sort(xs)
	return xs
}

// Two Swamps and three Mountains make five mana but only two black: the
// bot is offered X up to 2, the engine refuses X=3, and X=2 taps the
// Swamps, leaves the Mountains, and hits every creature and player.
func TestCryptRatsSpendsOnlyBlackManaOnX(t *testing.T) {
	g, me, opp := spendTable(t)
	rats := pushCatalogPermanent(g, me.ID, "Crypt Rats", "Creature — Rat", cryptRatsOracle, false)
	bear := b31Push(g, opp.ID, "Grizzly Bears", "Creature — Bear", "", "{1}{G}", 2, 3, "G")
	swamps := basics(g, me, "Swamp", 2)
	mountains := basics(g, me, "Mountain", 3)

	if got := xValuesOffered(t, g, me.ID, rats); !slices.Equal(got, []int{2}) {
		t.Errorf("enumerator offered X = %v, want [2] — the largest X two black sources pay, not the five any colour would", got)
	}
	refusedForMana(t, g.ActivateCatalogAbility(me.ID, rats, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true, XValue: 3}),
		"X=3 with two black sources")
	myLife, oppLife := me.Life, opp.Life
	if err := g.ActivateCatalogAbility(me.ID, rats, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true, XValue: 2}); err != nil {
		t.Fatalf("X=2 off two Swamps: %v", err)
	}
	for _, id := range mountains {
		if b31Tapped(t, g, id) {
			t.Error("a Mountain paid for black-only X")
		}
	}
	for _, id := range swamps {
		if !b31Tapped(t, g, id) {
			t.Error("a Swamp was not tapped")
		}
	}
	passPriorityAroundTable(t, g)
	if me.Life != myLife-2 || opp.Life != oppLife-2 {
		t.Errorf("life %d / %d, want %d / %d — X damage to each player", me.Life, opp.Life, myLife-2, oppLife-2)
	}
	if c, ok := battlefieldCard(g, bear); !ok || c.DamageMarked != 2 {
		t.Errorf("the 2/3 has %d damage, want 2", c.DamageMarked)
	}
}

// The Hellkite's X is red-only and its own {T} is part of the cost:
// two Mountains and two Forests pay X=2 and not X=3.
func TestCrimsonHellkiteSpendsOnlyRedManaOnX(t *testing.T) {
	g, me, opp := spendTable(t)
	kite := pushCatalogPermanent(g, me.ID, "Crimson Hellkite", "Creature — Dragon", crimsonHellkiteOracle, false)
	target := b31Push(g, opp.ID, "Grizzly Bears", "Creature — Bear", "", "{1}{G}", 2, 3, "G")
	basics(g, me, "Mountain", 2)
	forests := basics(g, me, "Forest", 2)
	at := []game.TargetRef{{Kind: game.TargetCard, ID: target}}

	if got := xValuesOffered(t, g, me.ID, kite); len(got) == 0 || got[len(got)-1] != 2 {
		t.Errorf("enumerator offered X = %v, want a largest X of 2", got)
	}
	refusedForMana(t, g.ActivateCatalogAbility(me.ID, kite, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true, XValue: 3, Targets: at}),
		"X=3 with two red sources")
	if err := g.ActivateCatalogAbility(me.ID, kite, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true, XValue: 2, Targets: at}); err != nil {
		t.Fatalf("X=2 off two Mountains: %v", err)
	}
	for _, id := range forests {
		if b31Tapped(t, g, id) {
			t.Error("a Forest paid for red-only X")
		}
	}
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, target); !ok || c.DamageMarked != 2 {
		t.Errorf("the target has %d damage, want 2", c.DamageMarked)
	}
}

// --- Pillar of the Paruns and Obsidian Obelisk: multicolored only --

func TestPillarOfTheParunsCastsOnlyMulticoloredSpells(t *testing.T) {
	g, me, _ := spendTable(t)
	pillar := pushCatalogPermanent(g, me.ID, "Pillar of the Paruns", "Land", pillarOfTheParunsOracle, false)
	if err := g.ActivateManaAbility(me.ID, pillar, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Pillar: %v", err)
	}
	color := resolveManaPickAnyColor(t, g, me.ID)
	mono := handSpell(me, "Mono Spell", "Sorcery", "{"+color+"}")
	refusedForMana(t, g.CastSpell(me.ID, mono, game.CastSpellParams{Strict: true}), "Pillar mana for a monocolored spell")
	other := "U"
	if color == "U" {
		other = "W"
	}
	gold := handSpell(me, "Gold Spell", "Sorcery", "{"+color+"/"+other+"}")
	if err := g.CastSpell(me.ID, gold, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("Pillar mana for a two-colour hybrid spell: %v", err)
	}
}

func TestObsidianObeliskEntersTappedAndRestrictsOnlyItsColoredMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	obelisk := castCatalogSpell(t, g, "Obsidian Obelisk", "Artifact", obsidianObeliskOracle, nil)
	passPriorityAroundTable(t, g)
	if !b31Tapped(t, g, obelisk) {
		t.Fatal("the Obelisk entered untapped")
	}
	untap(g, obelisk)
	if err := g.ActivateManaAbility(me.ID, obelisk, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Obelisk for {C}: %v", err)
	}
	if len(me.ManaPool) != 1 || me.ManaPool[0].Color != "C" || len(me.ManaPool[0].Restrictions) != 0 {
		t.Fatalf("pool %+v, want one unrestricted colorless", me.ManaPool)
	}
	me.ManaPool = nil
	untap(g, obelisk)
	if err := g.ActivateManaAbility(me.ID, obelisk, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Obelisk for a colour: %v", err)
	}
	resolveManaPickAnyColor(t, g, me.ID)
	if len(me.ManaPool) != 1 || !slices.Equal(me.ManaPool[0].Restrictions, []string{game.ManaRestrictCast, game.ManaRestrictMulticolored}) {
		t.Errorf("pool %+v, want one token restricted to multicolored spells", me.ManaPool)
	}
}

// --- agreement: the view, the bot and the payment --------------------

// throneActivationAgreement asks one board three questions about the
// Throne's draw ability and fails unless they give one answer: the
// view's legal_actions digest, the bot's enumerator, and the payment
// itself (strict, auto-tapped).
func throneActivationAgreement(t *testing.T, name string, g *game.Game, me *game.Player, throne uuid.UUID, want bool) {
	t.Helper()
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.KnownBy == nil {
				c.KnownBy = map[uuid.UUID]bool{}
			}
			for _, p := range g.Seats {
				c.KnownBy[p.ID] = true
			}
		}
	})
	view := false
	if v := protocol.ViewOfGameFor(g, me.ID.String()); v.LegalActions != nil {
		if src := v.LegalActions.Sources[throne.String()]; src != nil {
			view = slices.Contains(src.Kinds, legal.KindActivate)
		}
	}
	bot := false
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type == "activate_ability" && m.Source == throne {
			bot = true
		}
	}
	err := g.ActivateCatalogAbility(me.ID, throne, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true})
	if view != want || bot != want || (err == nil) != want {
		t.Errorf("%s: view %v, bot %v, payment %v — want all %v", name, view, bot, err, want)
	}
}

func TestThroneOfEldraineViewBotAndPaymentAgree(t *testing.T) {
	t.Run("three Plains", func(t *testing.T) {
		g, me, _ := spendTable(t)
		throne := throneFor(t, g, me, "W")
		basics(g, me, "Plains", 3)
		throneActivationAgreement(t, "three Plains", g, me, throne, true)
	})
	t.Run("two Plains and three Islands", func(t *testing.T) {
		g, me, _ := spendTable(t)
		throne := throneFor(t, g, me, "W")
		basics(g, me, "Plains", 2)
		basics(g, me, "Island", 3)
		throneActivationAgreement(t, "two Plains and three Islands", g, me, throne, false)
	})
	t.Run("three Islands under Chromatic Orrery", func(t *testing.T) {
		g, me, _ := spendTable(t)
		throne := throneFor(t, g, me, "W")
		// The Orrery is tapped so it cannot pay; the grant still applies.
		orrery := pushCatalogPermanent(g, me.ID, "Chromatic Orrery", "Legendary Artifact", chromaticOrreryOracle, false)
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == orrery {
					g.Battlefield.Cards[i].Tapped = true
				}
			}
		})
		basics(g, me, "Island", 3)
		throneActivationAgreement(t, "three Islands under the Orrery", g, me, throne, true)
	})
	t.Run("no colour chosen", func(t *testing.T) {
		g, me, _ := spendTable(t)
		throne := pushCatalogPermanent(g, me.ID, "Throne of Eldraine", "Legendary Artifact", throneOfEldraineOracle, false)
		basics(g, me, "Plains", 3)
		throneActivationAgreement(t, "no colour chosen", g, me, throne, false)
	})
}

// --- registration --------------------------------------------------

// Plus keeps the clause: a composed "{3}, {T}, spend only …" that
// dropped it would draw off any three mana.
func TestPlusKeepsTheSpendOnlyClause(t *testing.T) {
	c := Plus(ManaCost("{3}"), TapCost(), SpendOnlyManaOfTheChosenColor())
	if c.SpendOnly == nil || !c.SpendOnly.ChosenColor || c.Mana != "{3}" || !c.Tap {
		t.Errorf("Plus composed %+v", c)
	}
	x := Plus(ManaCost("{X}"), SpendOnlyOnX("B"))
	if x.SpendOnly == nil || !x.SpendOnly.XOnly || !slices.Equal(x.SpendOnly.Colors, []string{"B"}) {
		t.Errorf("Plus composed %+v", x)
	}
}

// Every shape the fold cannot pay is refused at boot.
func TestRegisterRefusesASpendOnlyClauseItCannotPay(t *testing.T) {
	for _, tc := range []struct {
		name, want string
		cost       game.AbilityCost
	}{
		{"no mana", "no mana component", Plus(TapCost(), SpendOnlyManaOfTheChosenColor())},
		{"on X with no X", "has no {X}", Plus(ManaCost("{2}"), SpendOnlyOnX("B"))},
		{"no colour", "names no colour", Plus(ManaCost("{X}"), SpendOnlyOnX())},
		{"colorless", "not one of W U B R G", Plus(ManaCost("{X}"), SpendOnlyOnX("C"))},
		{"waterbend", "waterbend", Plus(WaterbendCost("{3}"), SpendOnlyManaOfTheChosenColor())},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mustPanic(t, tc.want, func() { checkSpendOnlyClause("Test Card", 0, tc.cost) })
		})
	}
	checkSpendOnlyClause("Test Card", 0, Plus(ManaCost("{3}"), TapCost(), SpendOnlyManaOfTheChosenColor()))
}
