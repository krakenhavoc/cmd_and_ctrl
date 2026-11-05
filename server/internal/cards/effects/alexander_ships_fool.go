package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Alexander, Ship's Fool — Legendary Creature — Klingon Human {3}{R}{W},
// 3/5 (Star Trek Commander, #4974):
//
//	"Vigilance, haste
//	 {T}: Each player reveals the top card of their library, then draws a
//	 card. Alexander deals X damage to a random opponent, where X is the
//	 greatest mana value among cards revealed this way."
//
// The players reveal in APNAP order (CR 101.4) and each one's revealed
// mana value is read as it is shown, so X is fixed before anyone draws;
// a player with an empty library reveals nothing and still draws (and
// loses to the empty draw as usual). Then every player draws one card,
// in the same order. The opponent is drawn from the game's keyed random
// stream (Knight Rampager's), among the activator's opponents still in
// the game, and logged. The damage comes from Alexander as the object
// that activated the ability, last known if it has left (CR 608.2h);
// X of 0 deals nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d629311f-a7fc-4af5-91ce-5bf506a3cdb3",
		Name:            "Alexander, Ship's Fool",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "haste"},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Each player reveals the top card of their library, then draws a card. Alexander deals X damage to a random opponent, where X is the greatest mana value among cards revealed this way.",
			Cost:    TapCost(),
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Effect:  alexanderShipsFoolReveal,
		}},
	})
}

// alexanderShipsFoolReveal is the {T} ability's body.
func alexanderShipsFoolReveal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	players := apnapLivingSeats(g)
	x := 0
	for _, p := range players {
		for _, id := range g.RevealTopOfLibraryForEffect(p, item.SourceCardID, 1, "Alexander, Ship's Fool — reveals the top card of their library") {
			if c, ok := g.LookupCardForEffect(id); ok && c.ManaValue() > x {
				x = c.ManaValue()
			}
		}
	}
	for _, p := range players {
		if err := (DrawCards{Player: p, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	pick := g.ChooseAtRandomForEffect(randomDraw(ctx), ctx.Opponents(), 1)
	if len(pick) != 1 {
		return nil
	}
	name := "an opponent"
	if p := g.PlayerByIDForEffect(pick[0]); p != nil && p.Name != "" {
		name = p.Name
	}
	g.EmitEvent(game.Event{
		Kind:   game.EventOptionChosen,
		Actor:  item.Controller,
		CardID: item.SourceCardID,
		Label:  "Alexander, Ship's Fool — " + name + " (at random)",
	})
	d := DealDamage{Source: item.SourceCardID, Target: pick[0], Amount: x}
	if ref, ok := ctx.SourceRef(); ok {
		d.SourceObject = &ref
	}
	return d.Apply(ctx)
}
