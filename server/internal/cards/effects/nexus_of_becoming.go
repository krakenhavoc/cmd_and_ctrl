package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nexus of Becoming — Artifact {6}:
//
//	"At the beginning of combat on your turn, draw a card. Then you may
//	 exile an artifact or creature card from your hand. If you do,
//	 create a token that's a copy of the exiled card, except it's a 3/3
//	 Golem artifact creature in addition to its other types."
//
// The draw comes first, then the optional pick over the hand as it is
// after the draw (so the drawn card is a candidate). The token copies
// the exiled card, which is read from exile, and the exception adds
// Artifact, Creature and Golem to its types and sets a printed 3/3 (CR
// 707.9a/b). Because the copy keeps the card's oracle ID, its triggers
// and statics come along.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "07142ee9-dc2c-4b33-ad20-2f5285225e86",
		Name:         "Nexus of Becoming",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Nexus of Becoming — draw a card, then you may exile an artifact or creature card from your hand", nexusOfBecomingBody),
		},
	})
}

func nexusOfBecomingBody(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
		return err
	}
	player := item.Controller
	candidates := handCardsMatching(g, player, Or(Artifact(), Creature()))
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   item.SourceCardID,
		Question: "Nexus of Becoming — you may exile an artifact or creature card from your hand",
		Cards:    candidates,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneHand,
		ThenKey:  nexusOfBecomingPick,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// nexusOfBecomingPick is the card chosen to exile, keyed (#4253): the trigger
// item rides the prompt as data. The token copy is made once the card is gone.
var nexusOfBecomingPick = askedPick("choose-cards/nexus-of-becoming-exile", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	g := ctx.Game
	if err := g.ExileCardForEffect(r.Picked[0]); err != nil {
		return err
	}
	return CreateTokenCopy{
		Controller: r.Chooser,
		Copy:       r.Picked[0],
		N:          1,
		Except:     nexusGolemExceptionFor(g.TextReaderForEffect(ctx.Item)),
	}.Apply(ctx)
})

// nexusGolemExceptionFor is "except it's a 3/3 Golem artifact creature
// in addition to its other types", with "Golem" read through
// `reader`, the object whose text it is (CR 612.2, #3984); nil reads the
// printed word.
func nexusGolemExceptionFor(reader *game.Card) func(*game.Card) {
	golem := reader.TextSubtype("Golem")
	return func(t *game.Card) { nexusGolemExceptionAs(t, golem) }
}

func nexusGolemExceptionAs(t *game.Card, golem string) {
	super, types, subs := game.ParseTypeLine(t.TypeLine)
	for _, want := range []string{"Artifact", "Creature"} {
		found := false
		for _, have := range types {
			if strings.EqualFold(have, want) {
				found = true
			}
		}
		if !found {
			types = append(types, want)
		}
	}
	hasGolem := false
	for _, s := range subs {
		if strings.EqualFold(s, golem) {
			hasGolem = true
		}
	}
	if !hasGolem {
		subs = append([]string{golem}, subs...)
	}
	line := strings.Join(append(append([]string(nil), super...), types...), " ")
	t.TypeLine = line + " — " + strings.Join(subs, " ")
	t.Power, t.Toughness = 3, 3
	t.VariableToughness = false
	t.PrintedPTKnown = true
}
