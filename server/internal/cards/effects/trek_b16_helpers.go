package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b16_helpers.go — shared pieces for Star Trek batch 16 (#4973).

// trekB16NewPlanetToken is Genesis Effect's "tapped Planet land token named
// New Planet with "{T}: Add one mana of any color."" (ttrk 27).
func trekB16NewPlanetToken() game.Card { return tokenFromCatalog(printedTrekB16NewPlanetToken) }

// printedTrekB16NewPlanetToken is the New Planet as PRINTED, its mana
// ability included: a colourless land, Planet its only subtype.
func printedTrekB16NewPlanetToken() tokenTemplate {
	return tokenTemplate{
		Slug: "new-planet",
		Card: game.Card{
			Name:     "New Planet",
			TypeLine: "Token Land — Planet",
		},
		Mana: []game.ManaAbilityShape{{
			TapCost:  true,
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color",
		}},
		Text: "{T}: Add one mana of any color.",
	}
}

// damageEachOpponentAndEachCreatureTheyControl is "deals n damage to each
// opponent and each creature your opponents control" as one damage event:
// every opponent, then every creature an opponent controls, the
// set fixed before the first point lands (damageEachMatching). Tectonic
// Hazard's spell and Khan, Despot by Design's enters trigger.
func damageEachOpponentAndEachCreatureTheyControl(ctx *Context, n int) error {
	return ctx.Game.DamageInstanceForEffect(func() error {
		for _, opp := range ctx.Opponents() {
			if err := (DealDamage{Source: ctx.Source(), Target: opp, Amount: n}).Apply(ctx); err != nil {
				return err
			}
		}
		return damageEachMatching(ctx, And(Creature(), OpponentControls()), n)
	})
}

// twiceXAmount is a damage amount of twice X (Khaaaaaaaaaaaannn!). A call,
// not a value, so the XMatters guard sees the card read X.
func twiceXAmount() func(*game.StackItem, *Context) int {
	return func(_ *game.StackItem, ctx *Context) int { return 2 * ctx.X() }
}

// trekB16CreatureTypesAmongNonBorg is Infinite Diversity's X, "the number
// of creature types among non-Borg creatures you control": each creature
// type counted once however many creatures share it (CR 205.3m), read
// through the layers. A changeling is every creature type (CR 702.73a), so
// it is a Borg and counts nothing. "Borg" is read through the spell's text
// (CR 612.2).
func trekB16CreatureTypesAmongNonBorg(ctx *Context) int {
	return trekCreatureTypesAmongNonBorg(ctx.Game, ctx.Controller(), ctx.TextSubtype("Borg"))
}

// trekB16InfiniteDiversityTail runs once the targeted card's exile has
// settled: the card that REALLY reached exile is copied X times, each copy
// offered for a free cast (CopyAndCast, CR 707.12), then the spell exiles
// itself. A card a replacement sent elsewhere is not copied.
var trekB16InfiniteDiversityTail = CardTail("trek-b16/infinite-diversity/copy-x-times-then-exile-self", func(ctx *Context, o game.TailOutcome) error {
	item := ctx.Item
	ctx = NewContext(ctx.Game, item)
	if len(o.IDs) > 0 {
		x := trekB16CreatureTypesAmongNonBorg(ctx)
		cards := make([]uuid.UUID, 0, x*len(o.IDs))
		for _, id := range o.IDs {
			for i := 0; i < x; i++ {
				cards = append(cards, id)
			}
		}
		if len(cards) > 0 {
			if err := (CopyAndCast{Cards: cards, Free: true}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return ExileTarget{Target: item.SourceCardID}.Apply(ctx)
})

// lwaxanaTroiDoublerGrant is the bundle Lwaxana Troi's ability gives the
// chosen creature for as long as she remains tapped: its own triggered
// abilities trigger an additional time (CR 603.2d).
const lwaxanaTroiDoublerGrant = "lwaxana-troi/that-creature-triggers-again"

// lwaxanaTroiDoubler doubles a triggered ability whose source is the
// creature carrying the grant, whatever caused it. A spell's triggers and
// any other permanent's are not doubled.
func lwaxanaTroiDoubler() game.TriggerDoubler {
	return game.TriggerDoubler{
		Label:    "Lwaxana Troi — that creature's triggered ability triggers an additional time",
		RowLabel: "If a triggered ability of this creature triggers, that ability triggers an additional time.",
		Applies: func(_ *game.Game, q game.TriggerDoublingQuery) bool {
			return !q.FromSpell && q.Source.InstanceID == q.Doubler.InstanceID
		},
	}
}
