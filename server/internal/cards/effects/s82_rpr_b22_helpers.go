package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_rpr_b22_helpers.go — shared pieces of the reprint pass, batch 22
// (#5908).

// mitoticSlimeOozeText is the printed token card's ability text.
const mitoticSlimeOozeText = "When this creature dies, create two 1/1 green Ooze creature tokens."

// mitoticSlimeOozeToken is the 2/2 green Ooze Mitotic Slime makes. It
// carries its own dies trigger, which makes two plain 1/1 green Oozes.
func MitoticSlimeOozeToken() game.Card { return TokenFromCatalog(printedMitoticSlimeOozeToken) }

func printedMitoticSlimeOozeToken() TokenTemplate {
	return TokenTemplate{
		Slug: "ooze-dies-two-1-1-oozes",
		Card: game.Card{
			Name:      "Ooze",
			TypeLine:  "Token Creature — Ooze",
			Power:     2,
			Toughness: 2,
			Colors:    []string{"G"},
		},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Ooze — create two 1/1 green Ooze creature tokens",
				Do(CreateToken{Template: TokenCard("1/1 green Ooze"), N: 2})),
		},
		Text: mitoticSlimeOozeText,
	}
}

// rprB22ControlsGreatestPower is "you control a creature with the greatest
// power among creatures on the battlefield" (Primal Empathy): some creature
// of yours has power at least equal to every creature's, a tie for greatest
// counting, read through the layers.
func RprB22ControlsGreatestPower(g *game.Game, you uuid.UUID) bool {
	g.RecomputeLayersIfStaleLocked()
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == you && c.IsCreature() && NeoB12HostHasGreatestPower(g, c.InstanceID) {
			return true
		}
	}
	return false
}

// rprB22PlusOneOnAChosenCreature is "put a +1/+1 counter on a creature you
// control": the choice is made as the effect resolves and does not target.
// One candidate is taken without a prompt; none does nothing.
func RprB22PlusOneOnAChosenCreature(ctx *Context) error {
	g, me := ctx.Game, ctx.Controller()
	var mine []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == me && c.IsCreature() {
			mine = append(mine, c.InstanceID)
		}
	}
	switch len(mine) {
	case 0:
		return nil
	case 1:
		return NothingIfGone(g.AddCounterForEffect(mine[0], game.CounterPlusOne, 1))
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  me,
		Source:   ctx.Source(),
		Question: "Primal Empathy — choose a creature you control to put a +1/+1 counter on",
		Cards:    mine,
		Min:      1,
		Max:      1,
		ThenKey:  rprB22PlusOnePick,
		LiveZone: game.ZoneBattlefield,
	})
	return nil
}

// rprB22PlusOnePick puts the counter on the creature picked. An on-disk
// key: never renamed, never reused.
var rprB22PlusOnePick = game.RegisterChooseCardsThen("choose-cards/rpr-primal-empathy-counter",
	func(g *game.Game, r game.CardsChosen) error {
		if len(r.Picked) != 1 {
			return nil
		}
		return NothingIfGone(g.AddCounterForEffect(r.Picked[0], game.CounterPlusOne, 1))
	})
