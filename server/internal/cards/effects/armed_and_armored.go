package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Armed and Armored — Instant {1}{W} (S82 reprint pass):
//
//	"Vehicles you control become artifact creatures until end of turn.
//	 Choose a Dwarf you control. Attach any number of Equipment you control
//	 to it."
//
// Nothing targets. The Vehicles are animated as the spell resolves (Push the
// Limit's BecomeCreatureUntilEOT, which keeps the printed power and
// toughness), so a Vehicle that is an Equipment's host is not special. The
// Dwarf is a creature you control with the Dwarf type, picked as it
// resolves; with none, the rest does nothing. Then the controller picks any
// number (a floor of zero) of the Equipment they control, and each one still
// on the battlefield is attached to the Dwarf if that is still a Dwarf they
// control (CR 701.3). Equipment already on it stays put.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3d5b6151-d985-423b-8ade-f00958bbd3f6",
		Name:         "Armed and Armored",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			you := ctx.Controller()
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller != you || !c.HasSubtype("Vehicle") {
					continue
				}
				if err := (BecomeCreatureUntilEOT{Target: c.InstanceID, Label: "Armed and Armored — becomes an artifact creature"}).Apply(ctx.asGroupMember()); err != nil {
					return err
				}
			}
			dwarf := ctx.TextSubtype("Dwarf")
			dwarves := armedAndArmoredPermanents(ctx.Game, you, func(c game.Card) bool { return c.IsCreature() && c.HasSubtype(dwarf) })
			if len(dwarves) == 0 {
				return nil
			}
			ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:  you,
				Source:   item.SourceCardID,
				Question: "Armed and Armored — choose a Dwarf you control",
				Cards:    dwarves,
				Min:      1,
				Max:      1,
				LiveZone: game.ZoneBattlefield,
				ThenKey:  armedAndArmoredDwarfChosen,
				// "Dwarf" is a word of the card's text (CR 612.2).
				CarryWord: dwarf,
			})
			return nil
		},
	})
}

// armedAndArmoredDwarfChosen is the Dwarf picked, and the question of the
// Equipment that goes on it. An on-disk key.
var armedAndArmoredDwarfChosen = game.RegisterChooseCardsThen("choose-cards/armed-and-armored-dwarf", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	equipment := armedAndArmoredPermanents(g, r.Chooser, func(c game.Card) bool { return c.HasSubtype("Equipment") })
	if len(equipment) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  r.Chooser,
		Source:   r.Source,
		Question: "Armed and Armored — attach any number of Equipment you control to the Dwarf",
		Cards:    equipment,
		Min:      0,
		Max:      len(equipment),
		LiveZone: game.ZoneBattlefield,
		ThenKey:  armedAndArmoredEquipmentChosen,
		Carry:    []uuid.UUID{r.Picked[0]},
		// The Dwarf type, a word of the card's text, rides along.
		CarryWord: r.Word,
	})
	return nil
})

// armedAndArmoredEquipmentChosen attaches each Equipment picked to the Dwarf
// (the carry), if both are still what they were. An on-disk key.
var armedAndArmoredEquipmentChosen = game.RegisterChooseCardsThen("choose-cards/armed-and-armored-equipment", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Carry) == 0 {
		return nil
	}
	host, ok := g.LookupCardForEffect(r.Carry[0])
	if !ok || !g.Battlefield.Contains(host.InstanceID) || host.Controller != r.Chooser || !host.IsCreature() || !host.HasSubtype(r.Word) {
		return nil
	}
	for _, id := range r.Picked {
		if !g.Battlefield.Contains(id) {
			continue
		}
		if err := g.AttachForEffect(id, game.TargetRef{Kind: game.TargetCard, ID: host.InstanceID}); err != nil {
			return err
		}
	}
	return nil
})

// armedAndArmoredPermanents lists the battlefield permanents a player
// controls that match.
func armedAndArmoredPermanents(g *game.Game, controller uuid.UUID, match func(game.Card) bool) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && match(c) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}
