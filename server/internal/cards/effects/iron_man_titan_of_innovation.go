package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Iron Man, Titan of Innovation — Legendary Artifact Creature — Human Hero
// {3}{U}{R}, 4/4 (Marvel Universe, #4460):
//
//	"Flying, haste
//	 Genius Industrialist — Whenever Iron Man attacks, create a Treasure
//	 token, then you may sacrifice a noncreature artifact. If you do,
//	 search your library for an artifact card with mana value equal to 1
//	 plus the sacrificed artifact's mana value, put it onto the battlefield
//	 tapped, then shuffle."
//
// Shadow, Mysterious Assassin's optional sacrifice: the Treasure is made
// first (so it is a candidate: sacrificing it fetches a mana value 1
// artifact, a token's mana value being 0), then the controller is asked
// which noncreature artifact to sacrifice, with "none" allowed. The pick
// is keyed so a table waiting on it is a restore point. The sacrificed
// permanent's mana value is read before it leaves (it is the tail's N) and
// "if you do" is whether the sacrifice happened (oneWentThisWay). The
// search is for any artifact card of that mana value, a creature included,
// and it is shuffled into whatever library is left. The fetched card
// enters tapped through the effect's own clause.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "67ab82de-1167-42f5-90bd-21c63edcc712",
		Name:            "Iron Man, Titan of Innovation",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Genius Industrialist — create a Treasure token, then you may sacrifice a noncreature artifact",
				ironManGeniusIndustrialist),
		},
	})
}

func ironManGeniusIndustrialist(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (CreateToken{Template: TreasureToken(), N: 1}).Apply(ctx); err != nil {
		return err
	}
	var candidates []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsArtifact() && !c.IsCreature() && !g.CantBeSacrificedForEffect(c.InstanceID) {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: item.Controller,
		Source:     item.SourceCardID,
		Question:   "Iron Man, Titan of Innovation — sacrifice a noncreature artifact? (search for an artifact card with one more mana value)",
		Cards:      candidates,
		Min:        0,
		Max:        1,
		LiveZone:   game.ZoneBattlefield,
		ThenKey:    ironManSacrificePick,
		Asked:      game.NewAskedItem(item),
	})
	return nil
}

// ironManSacrificePick is the artifact chosen to sacrifice, keyed: the
// asking item rides the prompt as data, and the sacrifice's continuation is
// a keyed tail carrying the sacrificed artifact's mana value.
var ironManSacrificePick = askedPick("choose-cards/iron-man-titan-of-innovation-sacrifice", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	mv := 0
	if c, ok := ctx.Game.LookupCardForEffect(r.Picked[0]); ok {
		mv = c.ManaValue()
	}
	return ctx.Game.SacrificeAllKeyedForEffect(uuid.Nil, r.Picked[:1], withN(ctx.ItemTail(ironManSacrificed), mv))
})

// ironManSacrificed is "if you do": the search for an artifact card with
// mana value one more than the sacrificed artifact's (the tail's N), onto
// the battlefield tapped, then shuffle. An on-disk key.
var ironManSacrificed = CardTail("then/iron-man-titan-of-innovation-sacrificed", func(ctx *Context, o game.TailOutcome) error {
	if !oneWentThisWay(o) {
		return nil
	}
	want := o.Tail.N + 1
	return SearchLibrary{
		Player:        ctx.Controller(),
		Predicate:     func(c game.Card) bool { return c.IsArtifact() && c.ManaValue() == want },
		Dest:          game.ZoneBattlefield,
		Limit:         1,
		Shuffle:       true,
		TappedOnEntry: true,
		Source:        ctx.Source(),
		Reason:        "Iron Man, Titan of Innovation — an artifact card with mana value " + strconv.Itoa(want),
	}.Apply(ctx)
})
