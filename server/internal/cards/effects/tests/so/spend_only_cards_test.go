package testsso

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// throneFor seeds a Throne of Eldraine under p with `color` chosen.
func throneFor(t *testing.T, g *game.Game, p *game.Player, color string) uuid.UUID {
	t.Helper()
	return effectstest.PushChosenColorPermanent(t, g, p.ID, "Throne of Eldraine", "Legendary Artifact", effectstest.ThroneOfEldraineOracle, color)
}

// --- Throne of Eldraine ---------------------------------------------

// The colour is chosen as it enters (#742's prompt, ColorForMana), and
// the mana ability then makes four of it, restricted.
func TestThroneOfEldraineChoosesAColorAndAddsFourOfIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	throne := effectstest.CastCatalogSpell(t, g, "Throne of Eldraine", "Legendary Artifact", effectstest.ThroneOfEldraineOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.AnswerColor(t, g, me.ID, "G")
	if prompt.ColorPurpose != game.ColorForMana {
		t.Errorf("the prompt declared purpose %q, want %q", prompt.ColorPurpose, game.ColorForMana)
	}
	if c, _ := effectstest.BattlefieldCard(g, throne); c.ChosenColor != "G" {
		t.Fatalf("chosen colour %q, want G", c.ChosenColor)
	}
	if err := g.ActivateManaAbility(me.ID, throne, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Throne: %v", err)
	}
	if got := effectstest.PoolColors(me); !slices.Equal(got, []string{"G", "G", "G", "G"}) {
		t.Fatalf("pool = %v, want four green", got)
	}
	for _, tok := range me.ManaPool.Units() {
		if !slices.Contains(tok.Restrictions, game.ManaRestrictMonocolored) || !slices.Contains(tok.Restrictions, game.ManaRestrictColor("G")) {
			t.Errorf("token %+v is not restricted to monocolored green spells", tok)
		}
	}
}

// "Spend this mana only to cast monocolored spells of that color": a
// mono-white spell yes; a hybrid white-blue spell (CR 202.2d — both
// colours), a colourless one and a blue one no.
func TestThroneOfEldraineManaCastsOnlyMonocoloredSpellsOfItsColor(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	throne := throneFor(t, g, me, "W")
	if err := g.ActivateManaAbility(me.ID, throne, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Throne: %v", err)
	}
	hybrid := effectstest.HandSpell(me, "Hybrid Spell", "Sorcery", "{W/U}{W/U}")
	colorless := effectstest.HandSpell(me, "Colorless Spell", "Artifact", "{2}")
	blue := effectstest.HandSpell(me, "Blue Spell", "Sorcery", "{U}")
	white := effectstest.HandSpell(me, "White Spell", "Sorcery", "{1}{W}{W}")
	effectstest.RefusedForMana(t, g.CastSpell(me.ID, hybrid, game.CastSpellParams{Strict: true}), "Throne mana for a white-blue hybrid spell")
	effectstest.RefusedForMana(t, g.CastSpell(me.ID, colorless, game.CastSpellParams{Strict: true}), "Throne mana for a colorless spell")
	effectstest.RefusedForMana(t, g.CastSpell(me.ID, blue, game.CastSpellParams{Strict: true}), "white Throne mana for a blue spell")
	if err := g.CastSpell(me.ID, white, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("Throne mana for a mono-white spell: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 {
		t.Errorf("pool = %v, want one white left", got)
	}
}

// The Throne's mana is for SPELLS: it cannot pay an activated ability,
// even one whose cost is its colour.
func TestThroneOfEldraineManaCannotPayAnAbility(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	throne := throneFor(t, g, me, "B")
	pest := effectstest.PushCatalogPermanent(g, me.ID, "Pestilence", "Enchantment", effectstest.PestilenceOracle, false)
	if err := g.ActivateManaAbility(me.ID, throne, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Throne: %v", err)
	}
	effectstest.RefusedForMana(t, g.ActivateCatalogAbility(me.ID, pest, 0, game.ActivateAbilityParams{Strict: true}), "Throne mana for Pestilence's {B}")
}

// "Spend only mana of the chosen color to activate this ability",
// through the auto-tapper: two Plains and three Islands cannot pay the
// {3} and nothing is tapped; a third Plains pays it, the Islands stay
// untapped, and the ability draws two.
func TestThroneOfEldraineDrawAutoTapsOnlyTheChosenColor(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	throne := throneFor(t, g, me, "W")
	plains := effectstest.Basics(g, me, "Plains", 2)
	islands := effectstest.Basics(g, me, "Island", 3)
	effectstest.RefusedForMana(t, g.ActivateCatalogAbility(me.ID, throne, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}),
		"two Plains and three Islands for a white-only {3}")
	for _, id := range append(append([]uuid.UUID{throne}, plains...), islands...) {
		if effectstest.B31Tapped(t, g, id) {
			t.Fatalf("a refused activation tapped %s", id)
		}
	}
	plains = append(plains, effectstest.Basics(g, me, "Plains", 1)...)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, throne, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("three Plains for a white-only {3}: %v", err)
	}
	for _, id := range plains {
		if !effectstest.B31Tapped(t, g, id) {
			t.Errorf("Plains %s was not tapped", id)
		}
	}
	for _, id := range islands {
		if effectstest.B31Tapped(t, g, id) {
			t.Errorf("Island %s was tapped for a white-only cost", id)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d → %d, want +2", hand, got)
	}
}

// From the pool: two white and a blue do not pay it, three white do.
func TestThroneOfEldraineDrawFromThePool(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	throne := throneFor(t, g, me, "W")
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}{W}{U}"); err != nil {
		t.Fatal(err)
	}
	effectstest.RefusedForMana(t, g.ActivateCatalogAbility(me.ID, throne, 0, game.ActivateAbilityParams{Strict: true}), "{W}{W}{U} for a white-only {3}")
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, throne, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("three white for a white-only {3}: %v", err)
	}
	if got := effectstest.PoolColors(me); !slices.Equal(got, []string{"U"}) {
		t.Errorf("pool = %v, want the blue left over", got)
	}
}

// Under Chromatic Orrery the two halves part ways (CR 609.4b): the draw
// ability's cost may be paid with any mana, but the Throne's own mana
// still casts only monocolored spells of its colour.
func TestThroneOfEldraineUnderChromaticOrrery(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	throne := throneFor(t, g, me, "W")
	effectstest.PushCatalogPermanent(g, me.ID, "Chromatic Orrery", "Legendary Artifact", effectstest.ChromaticOrreryOracle, false)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, throne, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("blue and colorless for a white-only {3} under the Orrery: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Untap(g, throne)
	if err := g.ActivateManaAbility(me.ID, throne, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Throne: %v", err)
	}
	blue := effectstest.HandSpell(me, "Blue Spell", "Sorcery", "{U}")
	effectstest.RefusedForMana(t, g.CastSpell(me.ID, blue, game.CastSpellParams{Strict: true}), "white Throne mana for a blue spell under the Orrery")
}

// The restriction is game state on both halves — the chosen colour on
// the permanent and the tags on the tokens — so it comes back from a
// snapshot whole.
func TestThroneOfEldraineSurvivesASnapshotRoundTrip(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
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
	if c, _ := effectstest.BattlefieldCard(r, throne); c.ChosenColor != "W" {
		t.Fatalf("restored chosen colour %q, want W", c.ChosenColor)
	}
	effectstest.Untap(r, throne)
	// Four restricted white, two plain white and a blue: the restricted
	// four cannot pay an activation, so the draw is one white short.
	effectstest.RefusedForMana(t, r.ActivateCatalogAbility(rme.ID, throne, 0, game.ActivateAbilityParams{Strict: true}), "restored: the Throne's own mana for its draw")
	// The unrestricted mana pays a hybrid spell; the restricted four
	// must not be what pays it, although the solver spends restricted
	// mana first wherever it may.
	hybrid := effectstest.HandSpell(rme, "Hybrid Spell", "Sorcery", "{W/U}{W/U}")
	if err := r.CastSpell(rme.ID, hybrid, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("restored: unrestricted mana for a hybrid spell: %v", err)
	}
	restricted := 0
	for _, tok := range rme.ManaPool.Units() {
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
	g, me, opp := effectstest.SpendTable(t)
	rats := effectstest.PushCatalogPermanent(g, me.ID, "Crypt Rats", "Creature — Rat", effectstest.CryptRatsOracle, false)
	bear := effectstest.B31Push(g, opp.ID, "Grizzly Bears", "Creature — Bear", "", "{1}{G}", 2, 3, "G")
	swamps := effectstest.Basics(g, me, "Swamp", 2)
	mountains := effectstest.Basics(g, me, "Mountain", 3)

	if got := xValuesOffered(t, g, me.ID, rats); !slices.Equal(got, []int{2}) {
		t.Errorf("enumerator offered X = %v, want [2] — the largest X two black sources pay, not the five any colour would", got)
	}
	effectstest.RefusedForMana(t, g.ActivateCatalogAbility(me.ID, rats, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true, XValue: 3}),
		"X=3 with two black sources")
	myLife, oppLife := me.Life, opp.Life
	if err := g.ActivateCatalogAbility(me.ID, rats, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true, XValue: 2}); err != nil {
		t.Fatalf("X=2 off two Swamps: %v", err)
	}
	for _, id := range mountains {
		if effectstest.B31Tapped(t, g, id) {
			t.Error("a Mountain paid for black-only X")
		}
	}
	for _, id := range swamps {
		if !effectstest.B31Tapped(t, g, id) {
			t.Error("a Swamp was not tapped")
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife-2 || opp.Life != oppLife-2 {
		t.Errorf("life %d / %d, want %d / %d — X damage to each player", me.Life, opp.Life, myLife-2, oppLife-2)
	}
	if c, ok := effectstest.BattlefieldCard(g, bear); !ok || c.DamageMarked != 2 {
		t.Errorf("the 2/3 has %d damage, want 2", c.DamageMarked)
	}
}

// The Hellkite's X is red-only and its own {T} is part of the cost:
// two Mountains and two Forests pay X=2 and not X=3.
func TestCrimsonHellkiteSpendsOnlyRedManaOnX(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	kite := effectstest.PushCatalogPermanent(g, me.ID, "Crimson Hellkite", "Creature — Dragon", effectstest.CrimsonHellkiteOracle, false)
	target := effectstest.B31Push(g, opp.ID, "Grizzly Bears", "Creature — Bear", "", "{1}{G}", 2, 3, "G")
	effectstest.Basics(g, me, "Mountain", 2)
	forests := effectstest.Basics(g, me, "Forest", 2)
	at := []game.TargetRef{{Kind: game.TargetCard, ID: target}}

	if got := xValuesOffered(t, g, me.ID, kite); len(got) == 0 || got[len(got)-1] != 2 {
		t.Errorf("enumerator offered X = %v, want a largest X of 2", got)
	}
	effectstest.RefusedForMana(t, g.ActivateCatalogAbility(me.ID, kite, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true, XValue: 3, Targets: at}),
		"X=3 with two red sources")
	if err := g.ActivateCatalogAbility(me.ID, kite, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true, XValue: 2, Targets: at}); err != nil {
		t.Fatalf("X=2 off two Mountains: %v", err)
	}
	for _, id := range forests {
		if effectstest.B31Tapped(t, g, id) {
			t.Error("a Forest paid for red-only X")
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, target); !ok || c.DamageMarked != 2 {
		t.Errorf("the target has %d damage, want 2", c.DamageMarked)
	}
}

// --- Pillar of the Paruns and Obsidian Obelisk: multicolored only --

func TestPillarOfTheParunsCastsOnlyMulticoloredSpells(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	pillar := effectstest.PushCatalogPermanent(g, me.ID, "Pillar of the Paruns", "Land", effectstest.PillarOfTheParunsOracle, false)
	if err := g.ActivateManaAbility(me.ID, pillar, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Pillar: %v", err)
	}
	color := effectstest.ResolveManaPickAnyColor(t, g, me.ID)
	mono := effectstest.HandSpell(me, "Mono Spell", "Sorcery", "{"+color+"}")
	effectstest.RefusedForMana(t, g.CastSpell(me.ID, mono, game.CastSpellParams{Strict: true}), "Pillar mana for a monocolored spell")
	other := "U"
	if color == "U" {
		other = "W"
	}
	gold := effectstest.HandSpell(me, "Gold Spell", "Sorcery", "{"+color+"/"+other+"}")
	if err := g.CastSpell(me.ID, gold, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("Pillar mana for a two-colour hybrid spell: %v", err)
	}
}

func TestObsidianObeliskEntersTappedAndRestrictsOnlyItsColoredMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	obelisk := effectstest.CastCatalogSpell(t, g, "Obsidian Obelisk", "Artifact", effectstest.ObsidianObeliskOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B31Tapped(t, g, obelisk) {
		t.Fatal("the Obelisk entered untapped")
	}
	effectstest.Untap(g, obelisk)
	if err := g.ActivateManaAbility(me.ID, obelisk, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Obelisk for {C}: %v", err)
	}
	if me.ManaPool.Total() != 1 || me.ManaPool.Units()[0].Color != "C" || len(me.ManaPool.Units()[0].Restrictions) != 0 {
		t.Fatalf("pool %+v, want one unrestricted colorless", me.ManaPool)
	}
	me.ManaPool = nil
	effectstest.Untap(g, obelisk)
	if err := g.ActivateManaAbility(me.ID, obelisk, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Obelisk for a colour: %v", err)
	}
	effectstest.ResolveManaPickAnyColor(t, g, me.ID)
	if me.ManaPool.Total() != 1 || !slices.Equal(me.ManaPool.Units()[0].Restrictions, []string{game.ManaRestrictCast, game.ManaRestrictMulticolored}) {
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
		g, me, _ := effectstest.SpendTable(t)
		throne := throneFor(t, g, me, "W")
		effectstest.Basics(g, me, "Plains", 3)
		throneActivationAgreement(t, "three Plains", g, me, throne, true)
	})
	t.Run("two Plains and three Islands", func(t *testing.T) {
		g, me, _ := effectstest.SpendTable(t)
		throne := throneFor(t, g, me, "W")
		effectstest.Basics(g, me, "Plains", 2)
		effectstest.Basics(g, me, "Island", 3)
		throneActivationAgreement(t, "two Plains and three Islands", g, me, throne, false)
	})
	t.Run("three Islands under Chromatic Orrery", func(t *testing.T) {
		g, me, _ := effectstest.SpendTable(t)
		throne := throneFor(t, g, me, "W")
		// The Orrery is tapped so it cannot pay; the grant still applies.
		orrery := effectstest.PushCatalogPermanent(g, me.ID, "Chromatic Orrery", "Legendary Artifact", effectstest.ChromaticOrreryOracle, false)
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == orrery {
					g.Battlefield.Cards[i].Tapped = true
				}
			}
		})
		effectstest.Basics(g, me, "Island", 3)
		throneActivationAgreement(t, "three Islands under the Orrery", g, me, throne, true)
	})
	t.Run("no colour chosen", func(t *testing.T) {
		g, me, _ := effectstest.SpendTable(t)
		throne := effectstest.PushCatalogPermanent(g, me.ID, "Throne of Eldraine", "Legendary Artifact", effectstest.ThroneOfEldraineOracle, false)
		effectstest.Basics(g, me, "Plains", 3)
		throneActivationAgreement(t, "no colour chosen", g, me, throne, false)
	})
}

// --- registration --------------------------------------------------

// Plus keeps the clause: a composed "{3}, {T}, spend only …" that
// dropped it would draw off any three mana.
func TestPlusKeepsTheSpendOnlyClause(t *testing.T) {
	c := effects.Plus(effects.ManaCost("{3}"), effects.TapCost(), effects.SpendOnlyManaOfTheChosenColor())
	if c.SpendOnly == nil || !c.SpendOnly.ChosenColor || c.Mana != "{3}" || !c.Tap {
		t.Errorf("Plus composed %+v", c)
	}
	x := effects.Plus(effects.ManaCost("{X}"), effects.SpendOnlyOnX("B"))
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
		{"no mana", "no mana component", effects.Plus(effects.TapCost(), effects.SpendOnlyManaOfTheChosenColor())},
		{"on X with no X", "has no {X}", effects.Plus(effects.ManaCost("{2}"), effects.SpendOnlyOnX("B"))},
		{"no colour", "names no colour", effects.Plus(effects.ManaCost("{X}"), effects.SpendOnlyOnX())},
		{"colorless", "not one of W U B R G", effects.Plus(effects.ManaCost("{X}"), effects.SpendOnlyOnX("C"))},
		{"waterbend", "waterbend", effects.Plus(effects.WaterbendCost("{3}"), effects.SpendOnlyManaOfTheChosenColor())},
	} {
		t.Run(tc.name, func(t *testing.T) {
			effectstest.MustPanic(t, tc.want, func() { effects.CheckSpendOnlyClause("Test Card", 0, tc.cost) })
		})
	}
	effects.CheckSpendOnlyClause("Test Card", 0, effects.Plus(effects.ManaCost("{3}"), effects.TapCost(), effects.SpendOnlyManaOfTheChosenColor()))
}
