package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Book of Exalted Deeds — Legendary Artifact — Book {W}{W}{W}
// (Adventures in the Forgotten Realms, S82):
//
//	"At the beginning of your end step, if you gained 3 or more life this
//	 turn, create a 3/3 white Angel creature token with flying.
//	 {W}{W}{W}, {T}, Exile The Book of Exalted Deeds: Put an enlightened
//	 counter on target Angel. It gains 'You can't lose the game and your
//	 opponents can't win the game.' Activate only as a sorcery."
//
// The end-step trigger is Bulette's intervening "if" over the controller's
// life-gain tally (CR 603.4). The activated ability exiles the Book as a cost.
// The gain has no stated duration, so it is an indefinite grant pinned to the
// Angel (CR 611.2a, as Balduvian Shaman's): the Angel's controller can't lose
// and their opponents can't win for as long as it stays on the battlefield
// (Cloudsteel Kirin's granted game-end gates).
//
// No simplification.
const bookOfExaltedDeedsGrant = "the-book-of-exalted-deeds/cant-lose"

func init() {
	Register(Spec{
		OracleID:     "046add85-952f-46ac-974a-463d2a7d3fd4",
		Name:         "The Book of Exalted Deeds",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:          bookOfExaltedDeedsGrant,
			GameEndGates: YouCantLoseOpponentsCantWin(),
			Text:         "You can't lose the game and your opponents can't win the game.",
		}},
		Triggered: []game.TriggeredAbility{
			IfVentureCond(ventureCond(func(g *game.Game, you uuid.UUID) bool { return b15LifeGainedThisTurn(g, you) >= 3 }),
				AtYourEndStep("The Book of Exalted Deeds — create a 3/3 white Angel creature token with flying",
					Do(CreateTokenAdvanced{Spec: Token(TokenCard("3/3 white Angel with flying")), N: 1}))),
		},
		Activated: []ActivatedAbility{{
			Label:        "{W}{W}{W}, {T}, Exile The Book of Exalted Deeds: Put an enlightened counter on target Angel. It gains \"You can't lose the game and your opponents can't win the game.\" Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{W}{W}{W}"), TapCost(), ExileThis()),
			SorcerySpeed: true,
			Targets:      TargetCreature("target Angel", Subtype("Angel")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard || !onBattlefield(g, t.ID) {
						continue
					}
					if err := (AddCounter{Target: t.ID, Kind: "enlightened", N: 1}).Apply(ctx); err != nil {
						return err
					}
					return GrantAbilitiesFor{
						Target:   t.ID,
						Keys:     []string{bookOfExaltedDeedsGrant},
						Duration: g.PinnedTo(game.IndefiniteDuration(), t.ID),
						Label:    "The Book of Exalted Deeds — can't lose the game",
					}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
