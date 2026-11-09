package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// City of the Daleks — Plane — Skaro (Doctor Who, #4870):
//
//	"Whenever you attack, target opponent loses X life, where X is the
//	 number of artifacts you control.
//	 Whenever chaos ensues, for each opponent, you create a 3/3 black
//	 Dalek artifact creature token with menace that attacks that
//	 opponent this turn if able. Those tokens gain haste. Sacrifice them
//	 at the beginning of the next end step."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6).
// "Whenever you attack" is one trigger per attack declaration (the engine
// emits an attack event per attacker; OncePerBatch keeps the first), and X
// is counted as it resolves. The chaos ability (CR 311.7) makes one
// creation for every opponent, each group tagged with its opponent, so each
// Dalek is bound to attack the opponent it was made for this turn
// (game.AttackPlayerIfAbleMod). "Those tokens gain haste" states no
// duration, so it lasts while they stay on the battlefield (CR 611.2a);
// the sacrifice is a delayed trigger at the next end step (CR 603.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a2ef6b79-faeb-4183-a510-f10b14dff417",
		Name:         "City of the Daleks",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(OncePerBatch(On(game.EventAttack,
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclaredByYou(ev, source.Controller)
				}, "City of the Daleks — target opponent loses X life (X = artifacts you control)",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					x := len(MatchingBattlefield(ctx, And(Artifact(), YouControl())))
					if x == 0 {
						return nil
					}
					return TargetPlayerLosesLife{N: x}.Apply(ctx)
				})),
				TargetPlayer("target opponent", Opponent())),
			WheneverChaosEnsues("City of the Daleks — for each opponent, create a hasty 3/3 Dalek that attacks that opponent this turn; sacrifice them at the next end step",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					var groups []game.TokenGroup
					for _, opp := range ctx.Opponents() {
						groups = append(groups, game.TokenGroup{
							Template: ctx.TextToken(TokenCard("3/3 black Dalek artifact with menace")), Count: 1, Tag: opp,
						})
					}
					if len(groups) == 0 {
						return nil
					}
					return g.CreateTokensKeyedForEffect(game.TokenCreation{
						Controller: item.Controller,
						Source:     item.SourceCardID,
						Groups:     groups,
					}, ctx.ItemTail(cityOfTheDaleksBind))
				}),
		},
	})
}

// cityOfTheDaleksBind is "that attacks that opponent this turn if able.
// Those tokens gain haste. Sacrifice them at the beginning of the next end
// step", over the one creation: each token attacks the opponent its group
// was tagged with. A token a replacement added (an untagged group) is not
// one of the Daleks the text binds. An on-disk key: never renamed, never
// reused.
var cityOfTheDaleksBind = CardTail("city-of-the-daleks/bind-tagged-tokens", func(ctx *Context, o game.TailOutcome) error {
	tags := game.TokenTags(o, 0)
	var bound []uuid.UUID
	for _, tok := range o.IDs {
		opp, ok := tags[tok]
		if !ok {
			continue
		}
		if err := (ScopedEffectFor{
			Target:   tok,
			Mods:     []game.Mod{game.AttackPlayerIfAbleMod(opp)},
			Duration: DurationUntilEndOfTurn(ctx),
			Label:    "City of the Daleks — attacks that opponent this turn if able",
		}).Apply(ctx); err != nil {
			return err
		}
		if err := (ScopedEffectFor{
			Target:   tok,
			Mods:     []game.Mod{game.AddKeywordsMod("haste")},
			Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), tok),
			Label:    "City of the Daleks — haste",
		}).Apply(ctx); err != nil {
			return err
		}
		bound = append(bound, tok)
	}
	if len(bound) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label: "City of the Daleks — sacrifice the Daleks",
		Cards: bound,
		Body:  sacrificeListedCardsBody,
	}.Apply(ctx)
})
