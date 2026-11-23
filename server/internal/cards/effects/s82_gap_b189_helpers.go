package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b189_helpers.go — shared bodies of GAP batch 189 (#5948), the
// Warhammer 40,000 Commander Necron and Astartes creatures.

// CondGapB189SevenOrMoreCreaturesDied is "if seven or more creatures died
// this turn", read as the item resolves: the table-wide tally of creatures
// that died (tokens included), Tallyman of Nurgle's second branch. It declares
// no footprint: the threshold is not one a footprint step can name.
type CondGapB189SevenOrMoreCreaturesDied struct{}

func (CondGapB189SevenOrMoreCreaturesDied) Holds(ctx *Context) bool {
	return ctx.Game.TurnTally.CreaturesDied >= 7
}

func (CondGapB189SevenOrMoreCreaturesDied) Condition() (game.FootprintStep, bool) {
	return game.FootprintStep{}, false
}

// CondGapB189FewerThanSevenCreaturesDied is the other branch: at least one
// creature died and fewer than seven did ("instead" of the seven-card branch).
type CondGapB189FewerThanSevenCreaturesDied struct{}

func (CondGapB189FewerThanSevenCreaturesDied) Holds(ctx *Context) bool {
	n := ctx.Game.TurnTally.CreaturesDied
	return n >= 1 && n < 7
}

func (CondGapB189FewerThanSevenCreaturesDied) Condition() (game.FootprintStep, bool) {
	return game.FootprintStep{}, false
}

// GapB189ArtifactLeftBattlefieldToGraveyardOrExile is Shard of the Void
// Dragon's condition: an artifact, anyone's, card or token, was put into a
// graveyard from the battlefield or into exile from the battlefield. Whether
// it was an artifact is read as it last existed on the battlefield (CR
// 603.10a), from the type snapshot the exit stamps, so an animated Vehicle
// still counts and a creature that stopped being an artifact does not.
func GapB189ArtifactLeftBattlefieldToGraveyardOrExile(ev game.Event, g *game.Game) bool {
	if ev.Kind != game.EventLTB || (ev.NewZone != game.ZoneGraveyard && ev.NewZone != game.ZoneExile) {
		return false
	}
	if was, known := ev.WasType("artifact"); known {
		return was
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.HasCardType("artifact")
}

// GapB189LoseHalfLifeAndGainIt is Shard of the Nightbringer's body: the
// targeted opponent loses half their life, rounded up, then you gain life
// equal to the life lost this way. A target that is no longer legal loses
// nothing and gains you nothing (CR 608.2b).
func GapB189LoseHalfLifeAndGainIt(ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		p := ctx.Game.PlayerByIDForEffect(t.ID)
		if p == nil || p.Life <= 0 {
			return nil
		}
		// The gain rides the loss's continuation (#793), which reports the life
		// really lost once any replacement window has settled.
		return ctx.Game.LoseLifeEachKeyedForEffect(ctx.Source(), []uuid.UUID{t.ID}, game.HalfLifeRoundedUp(p.Life),
			game.KeyedTail{Ref: drainGainTheTotal, Player: ctx.Controller(), Source: ctx.Source()})
	}
	return nil
}

// GapB189TombBladeLoseOrSacrifice is Tomb Blade's punisher for the player it
// hit: that player loses life equal to the number of creatures they control
// unless they sacrifice a creature of their choice. Losing the life is always
// offered; the sacrifice only with a creature to give (CR 608.2d).
func GapB189TombBladeLoseOrSacrifice(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := NewContext(g, item)
	victim := item.Trigger.Event.Target
	if pl := g.PlayerByIDForEffect(victim); pl == nil || pl.Eliminated {
		return nil
	}
	creatures := CreaturesControlledByPlayer(g, victim)
	n := len(creatures)
	if n == 0 {
		return nil
	}
	options := []game.ChoiceOption{{Label: gapB189TombBladeLoseText, LifeCost: n}}
	if len(sacrificeableCreatures(g, victim)) > 0 {
		options = append(options, game.ChoiceOption{Label: gapB189TombBladeSacText})
	}
	return PickOption{
		Player:   victim,
		Question: "Tomb Blade — lose " + strconv.Itoa(n) + " life, or sacrifice a creature?",
		Options:  options,
		ThenKey:  gapB189TombBladeAnswered,
		Carry:    []uuid.UUID{},
	}.Apply(ctx)
}

const (
	gapB189TombBladeLoseText = "Lose life equal to the number of creatures you control"
	gapB189TombBladeSacText  = "Sacrifice a creature"
)

func sacrificeableCreatures(g *game.Game, player uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, id := range CreaturesControlledByPlayer(g, player) {
		if !g.CantBeSacrificedForEffect(id) {
			out = append(out, id)
		}
	}
	return out
}

// gapB189TombBladeAnswered is the victim's answer. An on-disk key: never
// renamed, never reused.
var gapB189TombBladeAnswered = OptionPickThen("option-pick/tomb-blade-lose-or-sacrifice", func(ctx *Context, r game.OptionPicked) error {
	victim := r.Chooser
	if r.Option != nil && r.Option.Label == gapB189TombBladeSacText {
		return SacrificeChoice{
			Player:     victim,
			Candidates: sacrificeableCreatures(ctx.Game, victim),
			Question:   "Tomb Blade — sacrifice a creature",
		}.Apply(ctx)
	}
	n := len(CreaturesControlledByPlayer(ctx.Game, victim))
	if n <= 0 {
		return nil
	}
	return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), victim, -n)
})

// GapB189TechnomancerMillThenReturn is Technomancer's body: mill three cards,
// then return any number of artifact creature cards with total mana value 6 or
// less from your graveyard to the battlefield. The set is chosen after the mill
// (CR 608.2c), so a card the mill put there is a legal pick, and the cards enter
// together.
func GapB189TechnomancerMillThenReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (MillCards{Player: item.Controller, N: 3}).Apply(ctx); err != nil {
		return err
	}
	candidates := GraveyardCardIDs(ctx, item.Controller, func(c game.Card) bool {
		return c.IsCreature() && c.HasCardType("artifact") && !c.IsToken()
	})
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: item.Controller,
		Source:     item.SourceCardID,
		Question:   "Technomancer — return any number of artifact creature cards with total mana value 6 or less to the battlefield",
		Cards:      candidates,
		Min:        0,
		Max:        len(candidates),
		ThenKey:    GapB189TechnomancerReturn,
		LiveZone:   game.ZoneGraveyard,
		Asked:      game.NewAskedItem(item),
	})
	return nil
}

// GapB189TechnomancerReturn puts the picked cards onto the battlefield as one
// entry. The set rule (total mana value 6 or less) is checked on the pick.
var GapB189TechnomancerReturn = AskedPickRule("choose-cards/technomancer-return", B23TotalManaValueAtMost(6),
	func(ctx *Context, r game.CardsChosen) error {
		if len(r.Picked) == 0 {
			return nil
		}
		return PutCardsFromZoneTogether(ctx, game.ZoneGraveyard, r.Picked, ctx.Controller(), false)
	})
