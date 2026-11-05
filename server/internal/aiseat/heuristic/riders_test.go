package heuristic_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// riders_test.go — #4978: scry, learn, discover, amass and the Ring
// tempting the bot, priced beside the purpose they ride on (riders.go),
// and nothing with Config.PriceRiders off.

// riderPrice is what the policy prices casting c from hand, with bf on
// the battlefield.
func riderPrice(t *testing.T, pol *heuristic.Policy, c protocol.CardView, bf ...protocol.CardView) float64 {
	t.Helper()
	bf = append(manaLands(6, 0, 100), bf...)
	v := newView([]protocol.PlayerView{newSeat(0, withHand(c)), newSeat(1)},
		withBattlefield(bf...), withTurn(3, 0, "precombat_main"))
	return rankValue(t, pol, input(0, v, passMove(0), castMove(t, 0, c.InstanceID, "Cast "+c.Name)), "Cast "+c.Name)
}

func noRiders() *heuristic.Policy {
	cfg := heuristic.DefaultConfig()
	cfg.PriceRiders = false
	return heuristic.NewWithConfig(cfg)
}

// Each rider adds its weight to the draw it rides on, and nothing with
// the price off.
func TestEachRiderAddsToTheSpellItRidesOn(t *testing.T) {
	pol := heuristic.New()
	cfg := pol.Config()
	hand := heuristic.DefaultWeights().Hand
	id := cardID(1)
	draw := func(p protocol.PurposeView) protocol.CardView {
		p.Draws = 2
		return sorcery(id, 0, "Rider Spell", "{2}{U}", &p)
	}
	plain := riderPrice(t, pol, draw(protocol.PurposeView{}))
	sickArmy := func(n int) float64 {
		a := creature(cardID(2), 0, "Army", n, n)
		a.IsToken, a.SummoningSick, a.TypeLine = true, true, "Token Creature — Army"
		return heuristic.DefaultWeights().CreatureValue(&a)
	}
	for _, c := range []struct {
		name string
		p    protocol.PurposeView
		want float64
	}{
		{"scry 1", protocol.PurposeView{Scry: 1}, hand * cfg.ScryWeight},
		{"scry 2", protocol.PurposeView{Scry: 2}, 2 * hand * cfg.ScryWeight},
		{"scry 5 is capped", protocol.PurposeView{Scry: 5}, float64(cfg.ScryCap) * hand * cfg.ScryWeight},
		{"learn", protocol.PurposeView{Learn: 1}, hand * cfg.LearnWeight},
		{"discover 4", protocol.PurposeView{Discover: 4}, hand + 4*cfg.DiscoverPerMana},
		{"amass 2 with no Army", protocol.PurposeView{Amass: 2}, sickArmy(2)},
		{"the Ring with no creature", protocol.PurposeView{RingTempts: 1}, cfg.RingTempt / 2},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := riderPrice(t, pol, draw(c.p)) - plain; !nearly(got, c.want) {
				t.Errorf("the rider added %.3f, want %.3f", got, c.want)
			}
			if c.want <= 0 {
				t.Errorf("the rider is worth %.3f, want more than nothing", c.want)
			}
			if with, without := riderPrice(t, noRiders(), draw(c.p)), riderPrice(t, noRiders(), draw(protocol.PurposeView{})); !nearly(with, without) {
				t.Errorf("with PriceRiders off the rider still moved the price: %.3f against %.3f", with, without)
			}
		})
	}
}

// The issue's acceptance test: with Preordain and a plain cantrip of the
// same cost in hand, the bot casts Preordain; with the price off the two
// tie and the bot no longer prefers it.
func TestTheBotPrefersTheSpellWithTheRider(t *testing.T) {
	preordain := sorcery(cardID(1), 0, "Preordain", "{U}", &protocol.PurposeView{Draws: 1, Scry: 2})
	plain := sorcery(cardID(2), 0, "Plain Cantrip", "{U}", &protocol.PurposeView{Draws: 1})
	v := newView([]protocol.PlayerView{newSeat(0, withHand(plain, preordain)), newSeat(1)},
		withBattlefield(manaLands(1, 0, 100)...), withTurn(3, 0, "precombat_main"))
	in := input(0, v, passMove(0),
		castMove(t, 0, plain.InstanceID, "Cast Plain Cantrip"),
		castMove(t, 0, preordain.InstanceID, "Cast Preordain"))
	pol := heuristic.New()
	if with, without := rankValue(t, pol, in, "Cast Preordain"), rankValue(t, pol, in, "Cast Plain Cantrip"); with <= without {
		t.Errorf("Preordain priced %.3f, the plain cantrip %.3f: want the scry to make it worth more", with, without)
	}
	if d, _ := pol.Decide(context.Background(), in); in.Moves[d.Index].Label != "Cast Preordain" {
		t.Errorf("the bot chose %q, want Preordain", in.Moves[d.Index].Label)
	}
	off := noRiders()
	if with, without := rankValue(t, off, in, "Cast Preordain"), rankValue(t, off, in, "Cast Plain Cantrip"); !nearly(with, without) {
		t.Errorf("with PriceRiders off Preordain priced %.3f against %.3f, want a tie", with, without)
	}
}

// A removal spell that also tempts the bot (Claim the Precious) is worth
// more than the same removal, and a creature that tempts it on entering
// bears the Ring itself, so its tempting is not halved.
func TestTheRingIsWorthMoreWithACreatureToBearIt(t *testing.T) {
	pol := heuristic.New()
	cfg := pol.Config()
	ring := sorcery(cardID(1), 0, "Ring Spell", "{1}{B}", &protocol.PurposeView{Draws: 2, RingTempts: 1})
	plain := sorcery(cardID(1), 0, "Ring Spell", "{1}{B}", &protocol.PurposeView{Draws: 2})
	bear := creature(cardID(5), 0, "Bear", 2, 2)
	if got := riderPrice(t, pol, ring, bear) - riderPrice(t, pol, plain, bear); !nearly(got, cfg.RingTempt) {
		t.Errorf("with a creature out the Ring added %.3f, want %.3f", got, cfg.RingTempt)
	}

	// Claim the Precious: "Destroy target creature. The Ring tempts you."
	// Its only declaration is the rider, and the rider is added to the
	// removal's proxy rather than replacing it.
	theirs := creature(cardID(6), 1, "Ogre", 4, 4)
	claim := sorcery(cardID(1), 0, "Claim the Precious", "{1}{B}{B}", &protocol.PurposeView{RingTempts: 1})
	murder := claim
	murder.Purpose = nil
	removal := func(pol *heuristic.Policy, c protocol.CardView) float64 {
		v := newView([]protocol.PlayerView{newSeat(0, withHand(c)), newSeat(1)},
			withBattlefield(append(manaLands(6, 0, 100), bear, theirs)...), withTurn(3, 0, "precombat_main"))
		mv := castMove(t, 0, c.InstanceID, "Cast "+c.Name, cardTarget(theirs.InstanceID))
		return rankValue(t, pol, input(0, v, passMove(0), mv), "Cast "+c.Name)
	}
	if got := removal(pol, claim) - removal(pol, murder); !nearly(got, cfg.RingTempt) {
		t.Errorf("the removal's tempting added %.3f, want %.3f", got, cfg.RingTempt)
	}
	if got := removal(noRiders(), claim) - removal(noRiders(), murder); !nearly(got, 0) {
		t.Errorf("with PriceRiders off the tempting added %.3f, want nothing", got)
	}

	// Uruk-hai Berserker: "When this creature enters, the Ring tempts you."
	uruk := creature(cardID(1), 0, "Uruk-hai Berserker", 3, 2)
	uruk.ManaCost = "{2}{B}"
	uruk.Purpose = &protocol.PurposeView{RingTempts: 1}
	body := uruk
	body.Purpose = nil
	if got := riderPrice(t, pol, uruk) - riderPrice(t, pol, body); !nearly(got, cfg.RingTempt) {
		t.Errorf("the creature's own tempting added %.3f, want %.3f", got, cfg.RingTempt)
	}
}

// Amass grows the Army the bot already has rather than making another.
func TestAmassGrowsTheArmyTheBotControls(t *testing.T) {
	pol := heuristic.New()
	w := heuristic.DefaultWeights()
	army := creature(cardID(5), 0, "Orc Army", 1, 1)
	army.IsToken, army.TypeLine = true, "Token Creature — Orc Army"
	grown := army
	grown.Power, grown.Toughness = 3, 3
	want := w.CreatureValue(&grown) - w.CreatureValue(&army)

	amass := sorcery(cardID(1), 0, "Amass Spell", "{1}{B}", &protocol.PurposeView{Draws: 2, Amass: 2})
	plain := sorcery(cardID(1), 0, "Amass Spell", "{1}{B}", &protocol.PurposeView{Draws: 2})
	if got := riderPrice(t, pol, amass, army) - riderPrice(t, pol, plain, army); !nearly(got, want) {
		t.Errorf("amass 2 onto a 1/1 Army added %.3f, want the two counters' %.3f", got, want)
	}
}

// An activated row whose whole effect is a rider (Crystal Ball's "{1},
// {T}: Scry 2.") is priced at the flat ActivateBase plus the rider.
func TestAScryRowIsPricedByItsRider(t *testing.T) {
	ball := rock(cardID(1), 0, "Crystal Ball", "{3}", protocol.ManaAbilityView{})
	ball.ManaAbilities = nil
	ball.ActivatedAbilities = []protocol.ActivatedAbilityView{{
		Index: 0, TapCost: true, ManaCost: "{1}",
		Purpose: &protocol.PurposeView{Scry: 2},
	}}
	act := legal.Move{
		Type: legal.TypeActivateAbility, Player: seatID(0), Kind: legal.KindActivate,
		Label: "Activate Crystal Ball", Source: uuid.MustParse(ball.InstanceID),
		Params: mustJSON(t, map[string]any{"source_card_id": ball.InstanceID, "ability_index": 0}),
	}
	price := func(pol *heuristic.Policy) float64 {
		v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
			withBattlefield(append(manaLands(3, 0, 100), ball)...), withTurn(3, 0, "precombat_main"))
		return rankValue(t, pol, input(0, v, passMove(0), act), act.Label)
	}
	cfg := heuristic.DefaultConfig()
	if got, want := price(heuristic.New()), cfg.ActivateBase+2*heuristic.DefaultWeights().Hand*cfg.ScryWeight; !nearly(got, want) {
		t.Errorf("Crystal Ball's scry 2 priced %.3f, want %.3f", got, want)
	}
	if got := price(noRiders()); !nearly(got, cfg.ActivateBase) {
		t.Errorf("with PriceRiders off the row priced %.3f, want the flat ActivateBase %.3f", got, cfg.ActivateBase)
	}
}

// A modal trigger's scry bullet (Charming Prince's "Scry 2.") is worth
// something, so on an "up to one" prompt the bot takes it rather than
// none, and with the price off takes none, the enumerator's first.
func TestAScryBulletIsChosen(t *testing.T) {
	bullets := []bullet{
		{0, untargeted("Scry 2.", &protocol.PurposeView{Scry: 2})},
		{1, untargeted("You gain 3 life.", nil)},
	}
	in := modePickInput(t, 0, 1, bullets)
	if got := pickWith(t, heuristic.New(), in); got != "Scry 2." {
		t.Errorf("the bot chose %q, want the scry bullet", got)
	}
	if got := pickWith(t, noRiders(), in); got != "none" {
		t.Errorf("with PriceRiders off the bot chose %q, want none", got)
	}
}
