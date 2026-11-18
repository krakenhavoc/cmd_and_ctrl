package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ludevicOracle is shared by both faces: the front registers under it, the
// back under ludevicOracle + "#1".
const ludevicOracle = "f6f684f6-0c8a-4676-ba4e-d3e064293128"

const ludevicOlagName = "Olag, Ludevic's Hubris"

// Ludevic, Necrogenius // Olag, Ludevic's Hubris — a transforming creature
// (Midnight Hunt):
//
//	Ludevic, Necrogenius {U}{B}, Legendary Creature — Human Wizard, 2/3
//	  "Whenever Ludevic enters or attacks, mill a card.
//	   {X}{U}{U}{B}{B}, Exile X creature cards from your graveyard:
//	   Transform Ludevic. X can't be 0. Activate only as a sorcery."
//
//	Olag, Ludevic's Hubris, Legendary Creature — Zombie, 4/4
//	  "As this creature transforms into Olag, Ludevic's Hubris, it becomes
//	   a copy of a creature card exiled with it, except its name is Olag,
//	   Ludevic's Hubris, it's 4/4, and it's a legendary blue and black
//	   Zombie in addition to its other colors and types. Put a number of
//	   +1/+1 counters on Olag equal to the number of creature cards exiled
//	   with it."
//
// The activation's one X is both the mana and the number of cards exiled
// (Necropolis Fiend's {X} with X cards), with a floor of one. The exiled
// cards are linked to Ludevic as the cost pays (ExileCost.LinkToSource,
// CR 607.2a), so they stay "exiled with it" through the transform, which
// is the same object (CR 712.18).
//
// Olag's clause is the face's AsTransformsInto hook: it runs inside the
// turn over, off the stack, whatever asked for the transform (CR 712.20).
// With one creature card exiled with it, that card is the copy; with
// several, the controller picks one. The copy has no duration (CR 611.2a),
// is of the card's copiable values as it is in exile (CR 707.2) with the
// exceptions applied (name, 4/4, legendary, Zombie, blue and black, all in
// addition where printed), and the counters are put on afterwards, one for
// each creature card exiled with it.
//
// No simplification.
func init() {
	exile := ExileXFromGraveyardAsCost("X creature cards", func(c game.Card) bool { return c.IsCreature() })
	exile.ExileCards.LinkToSource = true
	Register(Spec{
		OracleID:     ludevicOracle,
		Name:         "Ludevic, Necrogenius",
		Completeness: CompletenessFull,
		XMatters:     true,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Ludevic — mill a card", Do(MillCards{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:        "{X}{U}{U}{B}{B}, Exile X creature cards from your graveyard: Transform Ludevic. X can't be 0. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{X}{U}{U}{B}{B}"), MinX(1), exile),
			SorcerySpeed: true,
			Purpose:      game.Purpose{Answers: game.AnswerValue},
			Effect:       Do(TransformThis{}),
		}},
	})
	Register(Spec{
		OracleID:         ludevicOracle + "#1",
		Name:             ludevicOlagName,
		Completeness:     CompletenessFull,
		AsTransformsInto: ludevicBecomesOlag,
	})
}

// ludevicExiledCreatureCards are the creature cards exiled with the
// permanent `id`, in exile order.
func ludevicExiledCreatureCards(g *game.Game, id uuid.UUID) []uuid.UUID {
	ref, ok := g.PermanentRefForEffect(id)
	if !ok {
		return nil
	}
	var out []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		if c.IsCreature() {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// ludevicBecomesOlag is the clause: copy the one creature card, or ask for
// one of several, then the counters.
func ludevicBecomesOlag(card *game.Card, ctx *Context) error {
	g := ctx.Game
	cands := ludevicExiledCreatureCards(g, card.InstanceID)
	switch len(cands) {
	case 0:
		return nil
	case 1:
		return ludevicCopy(g, card.InstanceID, cands[0])
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  card.Controller,
		Source:   card.InstanceID,
		Question: "Ludevic — choose a creature card exiled with it to become a copy of",
		Cards:    cands,
		Min:      1,
		Max:      1,
		ThenKey:  ludevicPick,
		LiveZone: game.ZoneExile,
	})
	return nil
}

// ludevicPick is the pick answered (an on-disk key). A card that has left
// exile meanwhile is not a card exiled with it.
var ludevicPick = game.RegisterChooseCardsThen("choose-cards/ludevic-olag-copy", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	if !slices.Contains(ludevicExiledCreatureCards(g, r.Source), r.Picked[0]) {
		return nil
	}
	return ludevicCopy(g, r.Source, r.Picked[0])
})

// ludevicCopy makes the permanent a copy of `of` with Olag's exceptions and
// puts the counters on it.
func ludevicCopy(g *game.Game, self, of uuid.UUID) error {
	v, ok := g.CopiableValuesForEffect(of)
	if !ok {
		return nil
	}
	v.SetName(ludevicOlagName)
	v.SetPowerToughness(4, 4)
	v.AddSupertype("Legendary")
	v.AddSubtype("Zombie")
	for _, colour := range []string{"U", "B"} {
		if !slices.Contains(v.Colors, colour) {
			v.Colors = append(v.Colors, colour)
		}
	}
	targets := []uuid.UUID{self}
	if !g.BecomeCopyForEffect(self, of, targets, v, game.IndefiniteDuration(), "Olag, Ludevic's Hubris — becomes a copy") {
		return nil
	}
	return g.AddCounterForEffect(self, game.CounterPlusOne, len(ludevicExiledCreatureCards(g, self)))
}
