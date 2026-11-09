package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Carpet of Flowers — Enchantment {G}:
//
//	"At the beginning of each of your main phases, if you haven't added
//	 mana with this ability this turn, you may add X mana of any one
//	 color, where X is the number of Islands target opponent controls."
//
// One triggered ability watching both of your main phases' beginnings (it
// is not a mana ability, so it uses the stack and targets: the printed
// ruling). The intervening if (CR 603.4) reads a per-turn, per-object mark
// this ability writes only when it really added mana, so a declined or
// empty (X = 0) first main phase leaves the second one live. "You may" is
// asked as the ability resolves, after the opponent is chosen; X is the
// Islands that opponent controls then, read through the layered view (a
// land that is an Island by any means counts), and the mana is one pick of
// a colour for all X.
func init() {
	const label = "Carpet of Flowers — you may add X mana of any one color, where X is the number of Islands target opponent controls"
	mainBegan := func(step game.Step) When { return StepBegan(step, true) }
	Register(Spec{
		OracleID:     "2ffc6372-f63b-4f32-8dd0-2d7938aeb412",
		Name:         "Carpet of Flowers",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(OnAny([]game.EventKind{game.EventStepBegan},
				func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					if !mainBegan(game.StepPrecombatMain)(ev, source, lki, g) && !mainBegan(game.StepPostcombatMain)(ev, source, lki, g) {
						return false
					}
					return !carpetAddedManaThisTurn(g, source.InstanceID)
				}, label, carpetOffer),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}

const carpetDoneLabel = "Carpet of Flowers — added mana"

func carpetAddedManaThisTurn(g *game.Game, source uuid.UUID) bool {
	return g.ResolvedThisTurn(source, carpetDoneLabel) > 0
}

// carpetOffer re-checks the intervening if, finds the opponent, and asks.
func carpetOffer(g *game.Game, item *game.StackItem) error {
	if carpetAddedManaThisTurn(g, item.SourceCardID) {
		return nil
	}
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		return MayChoice{
			Question: "Carpet of Flowers — add mana of any one color for each Island that player controls?",
			Then:     carpetAdd,
			Carry:    []uuid.UUID{t.ID},
		}.Apply(ctx)
	}
	return nil
}

// carpetAdd is the "yes": X mana of one chosen colour. An on-disk key.
var carpetAdd = MayYes("may/carpet-of-flowers-add-mana", func(ctx *Context, r game.ConfirmAnswer) error {
	if len(r.Carry) == 0 {
		return nil
	}
	x := 0
	for _, id := range permanentsControlledByMatching(ctx.Game, r.Carry[0], HasSubtype(ctx.TextSubtype("Island"))) {
		_ = id
		x++
	}
	if x == 0 {
		return nil
	}
	ctx.Game.NoteDoneThisTurn(ctx.Source(), carpetDoneLabel)
	return AddMana{Amount: game.ManaAmount{Options: []string{"W", "U", "B", "R", "G"}, N: x, OneColor: true}}.Apply(ctx)
})
